package transactions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/wafi04/backendvazzz/pkg/lib"
	"github.com/wafi04/backendvazzz/pkg/utils"
)

type CallbackDuitku struct {
	MerchantCode    string `json:"merchantCode"`
	Amount          string `json:"amount"`
	RefId           string `json:"refId"`
	MerchantOrderId string `json:"merchantOrderId"`
	ResultCode      string `json:"resultCode"`
	Signature       string `json:"signature"`
}

var (
	depositPattern = regexp.MustCompile(`^DEP\d+$`)
	paymentPattern = regexp.MustCompile(`^VAZZ\d+$`)
)

type TransactionType string

const (
	TransactionDeposit  TransactionType = "DEPOSIT"
	TransactionPayment  TransactionType = "PAYMENT"
	TransactionWithdraw TransactionType = "WITHDRAW"
	TransactionRefund   TransactionType = "REFUND"
	TransactionUnknown  TransactionType = "UNKNOWN"
)

func detectTransactionType(merchantOrderId string) TransactionType {
	id := strings.TrimSpace(strings.ToUpper(merchantOrderId))

	switch {
	case depositPattern.MatchString(id):
		return TransactionDeposit
	case paymentPattern.MatchString(id):
		return TransactionPayment

	default:
		return TransactionUnknown
	}
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

	switch detectTransactionType(data.MerchantOrderId) {
	case TransactionDeposit:
		return repo.processDeposit(c, data.MerchantOrderId)
	case TransactionPayment:
		return repo.processPayment(c, data.MerchantOrderId)
	default:
		return fmt.Errorf("unsupported transaction type for order %s", data.MerchantOrderId)
	}
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
func (repo *TransactionsRepository) processDeposit(ctx context.Context, merchantOrderId string) error {
	tx, err := repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rErr := tx.Rollback(); rErr != nil && rErr != sql.ErrTxDone {
			log.Printf("Error during transaction rollback: %v", rErr)
		}
	}()

	// 1. Ambil data deposit
	var deposit struct {
		ID       int
		Username string
		Amount   int
		Status   string
		Method   string
	}

	err = tx.QueryRowContext(ctx, `
        SELECT id, username, amount, status, method 
        FROM deposits 
        WHERE deposit_id = $1`, merchantOrderId).Scan(
		&deposit.ID, &deposit.Username, &deposit.Amount, &deposit.Status, &deposit.Method)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("deposit not found: %s", merchantOrderId)
		}
		return fmt.Errorf("failed to get deposit: %w", err)
	}

	// 2. Validasi status deposit
	if strings.ToUpper(deposit.Status) != "PENDING" {
		return fmt.Errorf("deposit already processed: status %s", deposit.Status)
	}

	// 3. Hitung amount dengan fee
	var amountWithFee int
	switch deposit.Method {
	case "Mandiri Virtual Account":
		fee := utils.CalculateFeeQris(deposit.Amount)
		amountWithFee = deposit.Amount - fee
		log.Printf("Applied fee %d for Mandiri Virtual Account, final amount: %d", fee, amountWithFee)
	default:
		amountWithFee = deposit.Amount
	}

	_, err = tx.ExecContext(ctx, `
        WITH updated_deposit AS (
            UPDATE deposits 
            SET status = 'SUCCESS', 
                log = 'Deposit berhasil diproses', 
                updated_at = NOW()
            WHERE deposit_id = $1
            RETURNING id
        ),
        updated_user AS (
            UPDATE users 
            SET balance = COALESCE(balance, 0) + $2,
                updated_at = NOW()
            WHERE username = $3
            RETURNING id, balance, balance - $2 AS old_balance
        )
        INSERT INTO balance_histories (
            username, 
            platform_id,
            batch_id,
            balance_before,
            balance_after,
            description,
            amount_changed,
            change_type,
            created_at
        )
        SELECT 
            $3, -- username
            'DUITKU',  
            $4, -- batch_id
            old_balance,      
            balance,           
            'Deposit via ' || $5, -- description
            $2, -- amount_changed                  
            'CREDIT', -- change_type            
            NOW()
        FROM updated_user`,
		merchantOrderId,     // $1
		amountWithFee,       // $2
		deposit.Username,    // $3
		uuid.New().String(), // $4
		deposit.Method)      // $5

	if err != nil {
		return fmt.Errorf("failed to process deposit transaction: %w", err)
	}

	// 5. Commit transaksi
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Deposit processed successfully - ID: %s, Amount: %d, User: %s, Method: %s",
		merchantOrderId, deposit.Amount, deposit.Username, deposit.Method)

	return nil
}
