package transactions

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
)

type CallbackData struct {
	Data CallbackDetail `json:"data"`
}

type CallbackDetail struct {
	RefID        string `json:"ref_id"`
	BuyerSKUCode string `json:"buyer_sku_code"`
	CustomerNo   string `json:"customer_no"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	SN           string `json:"sn"`
}

func (cd *TransactionsRepository) Callback(c context.Context, data CallbackData) error {
	detail := data.Data

	if detail.RefID == "" {
		return fmt.Errorf("ref_id tidak boleh kosong")
	}

	// Mulai transaksi database
	tx, err := cd.DB.BeginTx(c, nil)
	if err != nil {
		return fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback()

	var updatedAt time.Time
	var username *string
	var currentStatus string
	var price int

	var message string
	statusUpper := strings.ToUpper(detail.Status)

	switch statusUpper {
	case "SUCCESS", "COMPLETED", "SUKSES":
		message = "Transaksi Berhasil"
	case "FAILED", "ERROR", "CANCELLED", "GAGAL":
		if detail.Message != "" {
			message = "Transaksi Berhasil"
		} else {
			message = "Transaksi Gagal"
		}
	default:
		message = detail.Message
	}

	updateQuery := `
		UPDATE transactions 
		SET status = $1, 
			message = $2, 
			serial_number = $3, 
			updated_at = $4
		WHERE order_id = $5
		RETURNING username, status, updated_at, price`

	err = tx.QueryRowContext(c, updateQuery,
		statusUpper,
		message,
		detail.SN,
		time.Now(),
		detail.RefID,
	).Scan(
		&username,
		&currentStatus,
		&updatedAt,
		&price,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("transaksi dengan order_id %s tidak ditemukan", detail.RefID)
		}
		return fmt.Errorf("gagal update transaksi: %w", err)
	}

	// Ambil payment method
	var methodName string
	paymentQuery := `
		SELECT method FROM payments
		WHERE order_id = $1`

	err = tx.QueryRowContext(c, paymentQuery, detail.RefID).Scan(&methodName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Payment method tidak ditemukan untuk order_id: %s", detail.RefID)
			methodName = "UNKNOWN"
		} else {
			return fmt.Errorf("gagal mengambil payment method: %w", err)
		}
	}

	// Process berdasarkan status
	switch statusUpper {
	case "SUCCESS", "COMPLETED", "SUKSES":
		log.Printf("Transaksi sukses - RefID: %s, CustomerNo: %s, SN: %s, Method: %s",
			detail.RefID, detail.CustomerNo, detail.SN, methodName)

		// Jika berhasil, langsung commit tanpa proses tambahan
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("gagal commit transaksi: %w", err)
		}
		return nil

	case "FAILED", "ERROR", "CANCELLED", "GAGAL":
		log.Printf("Transaksi gagal - RefID: %s, Status: %s, Message: %s",
			detail.RefID, detail.Status, detail.Message)

		err = cd.processFailedTransaction(c, tx, detail, username, methodName, price)
		if err != nil {
			return fmt.Errorf("gagal proses transaksi gagal: %w", err)
		}

	default:
		log.Printf("Status tidak dikenali - RefID: %s, Status: %s", detail.RefID, detail.Status)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("gagal commit transaksi: %w", err)
	}

	return nil
}

func (cd *TransactionsRepository) processFailedTransaction(c context.Context, tx *sql.Tx, detail CallbackDetail, username *string, methodName string, price int) error {
	if username == nil {
		log.Printf("Username kosong untuk order_id: %s, skip refund", detail.RefID)
		return nil
	}

	log.Printf("Processing failed transaction for user: %s", *username)

	// Logic untuk transaksi gagal
	if methodName == "QRIS (All Payment)" {
		// QRIS dikenakan fee 0.7%
		fee := math.Round(float64(price) * (0.7 / 100))
		refundAmount := float64(price) - fee

		log.Printf("QRIS Refund - Original: %d, Fee: %.2f, Refund: %.2f",
			price, fee, refundAmount)

		// Refund dengan potongan fee
		err := cd.refundUserBalance(c, tx, *username, refundAmount, detail.RefID, "QRIS_REFUND")
		if err != nil {
			return fmt.Errorf("gagal refund balance QRIS: %w", err)
		}
	} else {
		// Untuk method lain, refund full amount
		log.Printf("Full Refund - Amount: %d", price)

		err := cd.refundUserBalance(c, tx, *username, float64(price), detail.RefID, "FULL_REFUND")
		if err != nil {
			return fmt.Errorf("gagal refund balance: %w", err)
		}
	}

	return nil
}

func (cd *TransactionsRepository) refundUserBalance(c context.Context, tx *sql.Tx, username string, amount float64, orderID string, refundType string) error {
	// Ambil balance sebelumnya
	var balanceBefore float64
	getBalanceQuery := `SELECT balance FROM users WHERE username = $1`
	err := tx.QueryRowContext(c, getBalanceQuery, username).Scan(&balanceBefore)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user dengan username %s tidak ditemukan", username)
		}
		return fmt.Errorf("gagal mengambil balance user: %w", err)
	}

	// Update user balance
	balanceQuery := `
		UPDATE users 
		SET balance = balance + $1, updated_at = $2
		WHERE username = $3`

	result, err := tx.ExecContext(c, balanceQuery, amount, time.Now(), username)
	if err != nil {
		return fmt.Errorf("gagal update balance: %w", err)
	}

	// Cek rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal cek rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("tidak ada user yang diupdate untuk username: %s", username)
	}

	// Hitung balance setelah update
	balanceAfter := balanceBefore + amount

	// Buat description berdasarkan refund type
	var description string
	switch refundType {
	case "QRIS_REFUND":
		description = fmt.Sprintf("Refund QRIS (dengan fee) untuk transaksi gagal: %s", orderID)
	case "FULL_REFUND":
		description = fmt.Sprintf("Refund penuh untuk transaksi gagal: %s", orderID)
	default:
		description = fmt.Sprintf("Refund untuk transaksi gagal: %s", orderID)
	}

	// Insert balance history
	historyQuery := `
		INSERT INTO balance_histories (
			username,
			platform_id,
			batch_id,
			balance_before,
			balance_after,
			amount_changed,
			change_type,
			description,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.ExecContext(c, historyQuery,
		username,
		"DIGIFLAZZ",
		orderID,
		balanceBefore,
		balanceAfter,
		amount,
		"REFUND",
		description,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("gagal insert balance history: %w", err)
	}

	log.Printf("Refund berhasil untuk user %s sebesar %.2f (Balance: %.2f -> %.2f)",
		username, amount, balanceBefore, balanceAfter)
	return nil
}

func (cd *TransactionsRepository) GetTransactionByOrderID(c context.Context, orderID string) (*TransactionDetail, error) {
	query := `
		SELECT order_id, username, status, message, serial_number, price, created_at, updated_at
		FROM transactions 
		WHERE order_id = $1`

	var t TransactionDetail
	err := cd.DB.QueryRowContext(c, query, orderID).Scan(
		&t.OrderID, &t.Username, &t.Status,
		&t.Message, &t.SN, &t.Price,
		&t.CreatedAt, &t.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaksi dengan order_id %s tidak ditemukan", orderID)
		}
		return nil, fmt.Errorf("gagal mengambil data transaksi: %w", err)
	}

	return &t, nil
}

type TransactionDetail struct {
	OrderID   string    `json:"order_id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	SN        string    `json:"sn"`
	Price     int       `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
