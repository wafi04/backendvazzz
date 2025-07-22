package transactions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/wafi04/backendvazzz/pkg/config"
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

func (repo *TransactionsRepository) CallbackTransactionFromDuitku(c context.Context, duitkuRawResponseBytes []byte) error {
	var data CallbackDuitku
	digiflazz := lib.NewDigiflazzService(lib.DigiConfig{
		DigiKey:      config.GetEnv("DIGI_KEY", ""),
		DigiUsername: config.GetEnv("DIGI_USERNAME", ""),
	})
	err := json.Unmarshal(duitkuRawResponseBytes, &data)
	if err != nil {

		return fmt.Errorf("failed to unmarshal Duitku response: %w", err)
	}

	logJSONBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal Duitku log data to JSON for order %s: %w", data.MerchantOrderId, err)
	}
	logDataString := string(logJSONBytes)

	var (
		TrxId             string
		TransactionStatus string
		TransactionType   string
		PaymentOrderId    sql.NullString
		PaymentMethod     sql.NullString
		Fee               sql.NullInt64
		Username          sql.NullString
	)

	// Mulai transaksi database
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
			t.type,
			p.order_id,      
			p.method,
			p.fee_amount,
			t.username       
		FROM transactions t
		LEFT JOIN payments p ON t.order_id = p.order_id
		WHERE t.order_id = $1
	`
	err = tx.QueryRowContext(c, querySelect, data.MerchantOrderId).Scan(
		&TrxId,
		&TransactionStatus,
		&TransactionType,
		&PaymentOrderId,
		&PaymentMethod,
		&Fee,
		&Username,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("transaction with order ID %s not found", data.MerchantOrderId)
		}
		return fmt.Errorf("failed to query transaction details for order %s: %w", data.MerchantOrderId, err)
	}

	if TransactionType != "TOPUP" {

		return fmt.Errorf("transaction type for order %s is %s, expected TOPUP", data.MerchantOrderId, TransactionType)
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
	result, err := tx.ExecContext(c, queryUpdate, logDataString, TrxId)
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

	digiflazz.TopUp(c, lib.CreateTransactionToDigiflazz{
		Username:     "casoyeDa3zJg",
		BuyerSKUCode: "CHECKIDS",
		CustomerNo:   "139600730",
		RefID:        "casoyeDa3zJg",
	})

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for order %s: %w", TrxId, err)
	}

	return nil // Sukses
}
