package transactions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/wafi04/backendvazzz/pkg/lib"
)

type CallbackDuitku struct {
	MerchantCode    string `json:"merchantCode"`
	Amount          string `json:"amount"`
	RefId           string `json:"refId"`
	MerchantOrderId string `json:"merchantOrderId"`
	ResultCode      string `json:"resultCode"`
	Signature       string `json:"signature"`
}

func (repo *TransactionsRepository) CallbackTransactionFromDuitkuRaw(c context.Context, duitkuRawResponseBytes []byte) error {
	// Parse form data
	formData, err := url.ParseQuery(string(duitkuRawResponseBytes))
	if err != nil {
		return fmt.Errorf("failed to parse form data: %w", err)
	}

	// Convert ke struct
	data := CallbackDuitku{
		MerchantCode:    formData.Get("merchantCode"),
		Amount:          formData.Get("amount"),
		RefId:           formData.Get("refId"),
		MerchantOrderId: formData.Get("merchantOrderId"),
		ResultCode:      formData.Get("resultCode"),
		Signature:       formData.Get("signature"),
	}

	// Log the callback data
	logJSONBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal Duitku log data for order %s: %w", data.MerchantOrderId, err)
	}
	log.Printf("Duitku callback data: %s", string(logJSONBytes))

	if data.ResultCode != "00" {
		log.Printf("Payment not successful. Result code: %s for order: %s", data.ResultCode, data.MerchantOrderId)
		return fmt.Errorf("payment not successful for order %s, result code: %s", data.MerchantOrderId, data.ResultCode)
	}

	return repo.processPayment(c, data.MerchantOrderId)
}

func (repo *TransactionsRepository) processPayment(c context.Context, merchantOrderId string) error {
	digiflazz := lib.NewDigiflazzService(lib.DigiConfig{
		DigiKey:      "f99884cd-b12d-5f6e-abf2-90d60f297bda",
		DigiUsername: "casoyeDa3zJg",
	})

	var (
		TrxId             string
		UserId            string
		Zone              *string
		TransactionStatus string
		ProductCode       string
		Price             int
		TransactionType   string
		PaymentOrderId    string
		PaymentMethod     string
		Fee               int
		Username          *string
	)

	tx, err := repo.DB.BeginTx(c, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rErr := tx.Rollback(); rErr != nil && rErr != sql.ErrTxDone {
			fmt.Printf("Error during transaction rollback: %v\n", rErr)
		}
	}()

	querySelect := `
		SELECT 
			t.order_id,
			t.status,
			t.user_id,
			t.zone,
			t.price,
			t.provider_order_id,
			t.transaction_type,
			p.order_id,      
			p.method,
			p.fee_amount,
			t.username       
		FROM transactions t
		LEFT JOIN payments p ON t.order_id = p.order_id
		WHERE t.order_id = $1
	`
	err = tx.QueryRowContext(c, querySelect, merchantOrderId).Scan(
		&TrxId,
		&TransactionStatus,
		&UserId,
		&Zone,
		&Price,
		&ProductCode,
		&TransactionType,
		&PaymentOrderId,
		&PaymentMethod,
		&Fee,
		&Username,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("transaction with order ID %s not found", merchantOrderId)
		}
		return fmt.Errorf("failed to query transaction details for order %s: %w", merchantOrderId, err)
	}

	if TransactionType != "TOPUP" {
		return fmt.Errorf("transaction type for order %s is %s, expected TOPUP", merchantOrderId, TransactionType)
	}

	if TransactionStatus == "PAID" {
		return fmt.Errorf("transaction %s already paid", TrxId)
	}

	queryUpdate := `
		UPDATE transactions 
		SET 
			status = 'PAID',
			message = 'Pesanan Sudah Berhasil Dibayar',
			updated_at = NOW(),
			log = $1
		WHERE order_id = $2
	`
	result, err := tx.ExecContext(c, queryUpdate, "Pesanan dari duitku", TrxId)
	if err != nil {
		return fmt.Errorf("failed to update transaction status for order %s: %w", TrxId, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected for order %s: %w", TrxId, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected when updating transaction %s", TrxId)
	}
	var customerNo string
	if Zone != nil || *Zone != "" {
		customerNo = fmt.Sprintf("%s%s", UserId, *Zone)
	} else {
		customerNo = UserId
	}
	digi, _ := digiflazz.TopUp(c, lib.CreateTransactionToDigiflazz{
		BuyerSKUCode: ProductCode,
		CustomerNo:   customerNo,
		RefID:        merchantOrderId,
	})

	switch strings.ToUpper(digi.Data.Status) {
	case "GAGAL":

		refund := Price - Fee

		var messages string

		if Username != nil {
			messages = "Transaksi Gagal, Payment Otomatis jadi Saldo"

			// Refund ke user balance
			queryRefund := `
			UPDATE users 
			SET balance = balance + $1
			WHERE username = $2
		`
			_, err := tx.ExecContext(c, queryRefund, refund, Username)
			if err != nil {
				return fmt.Errorf("failed to process refund: %w", err)
			}
		} else {
			messages = "Transaksi Gagal, Silahkan Hubungi Admin"
		}

		// Update transaction status
		queryUpdateTransaction := `
		UPDATE transactions
		SET 
			message = $1,
			status = 'FAILED',
			log = $2,
			updated_at = NOW()
		WHERE order_id = $3
	`
		_, err := tx.ExecContext(c, queryUpdateTransaction, messages, "", merchantOrderId)
		if err != nil {
			return fmt.Errorf("failed to update transaction: %w", err)
		}

		return nil
	case "SUKSES", "PENDING":
		queryUpdate := `
			UPDATE transactions
			SET 
				puchase_price = $1,
				status = 'PAID'
				updated_at = NOW()
			WHERE order_id = $2
			`
		_, err := tx.ExecContext(c, queryUpdate, digi.Data.Price, merchantOrderId)
		if err != nil {
			return nil
		}
		return nil
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for order %s: %w", TrxId, err)
	}

	return nil
}
