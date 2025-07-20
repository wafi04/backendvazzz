package transactions

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/wafi04/backendvazzz/pkg/types"
)

type TransactionsRepositoryInterface interface {
	Create(req *types.CreateTransactions) (*types.Transaction, error)
	GetByID(id int64) (*types.Transaction, error)
	GetByOrderID(orderId string) (*types.Transaction, error)
	GetByGameID(gameId string) ([]types.Transaction, error)
	GetAll(limit, offset int) ([]types.Transaction, error)
}

type TransactionsRepository struct {
	DB *sql.DB
}

func NewTransactionsRepository(DB *sql.DB) TransactionsRepositoryInterface {
	return &TransactionsRepository{
		DB: DB,
	}
}

func generateOrderID() string {
	timestamp := time.Now().Unix()
	bytes := make([]byte, 4)
	rand.Read(bytes)
	random := hex.EncodeToString(bytes)
	return fmt.Sprintf("VAZZ%d%s", timestamp, random)
}

func (repo *TransactionsRepository) Create(req *types.CreateTransactions) (*types.Transaction, error) {
	query := `
		INSERT INTO transactions (
			order_id, product_code, method_code, game_id, zone, voucher_code,
			whatsapp_number, nickname, username, ip, user_agent, 
			status, purchase_price, web_price, duitku_price, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING id, created_at, updated_at`

	var transaction types.Transaction
	now := time.Now()
	orderID := generateOrderID()

	err := repo.DB.QueryRow(
		query,
		orderID,
		req.ProductCode,
		req.MethodCode,
		req.GameId,
		req.Zone,
		req.VoucherCode,
		req.WhatsAppNumber,
		req.Nickname,
		req.Username,
		req.Ip,
		req.UserAgent,
		types.StatusPending,
		0,
		0,
		0,
		now,
		now,
	).Scan(&transaction.ID, &transaction.CreatedAt, &transaction.UpdatedAt)

	if err != nil {
		return nil, err
	}

	transaction.OrderId = orderID
	transaction.ProductCode = req.ProductCode
	transaction.MethodCode = req.MethodCode
	transaction.GameId = req.GameId
	transaction.Zone = req.Zone
	transaction.VoucherCode = req.VoucherCode
	transaction.WhatsAppNumber = req.WhatsAppNumber
	transaction.Nickname = req.Nickname
	transaction.Username = req.Username
	transaction.Ip = req.Ip
	transaction.UserAgent = req.UserAgent
	transaction.Status = types.StatusPending
	transaction.PurchasePrice = 0
	transaction.WebPrice = 0
	transaction.DuitkuPrice = 0

	return &transaction, nil
}

// GetByID method untuk mengambil transaction berdasarkan ID
func (repo *TransactionsRepository) GetByID(id int64) (*types.Transaction, error) {
	query := `
		SELECT id, order_id, product_code, method_code, game_id, zone, voucher_code,
			   whatsapp_number, nickname, username, ip, user_agent,
			   status, purchase_price, web_price, duitku_price, 
			   created_at, updated_at, completed_at
		FROM transactions WHERE id = $1`

	var transaction types.Transaction
	err := repo.DB.QueryRow(query, id).Scan(
		&transaction.ID,
		&transaction.OrderId,
		&transaction.ProductCode,
		&transaction.MethodCode,
		&transaction.GameId,
		&transaction.Zone,
		&transaction.VoucherCode,
		&transaction.WhatsAppNumber,
		&transaction.Nickname,
		&transaction.Username,
		&transaction.Ip,
		&transaction.UserAgent,
		&transaction.Status,
		&transaction.PurchasePrice,
		&transaction.WebPrice,
		&transaction.DuitkuPrice,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
		&transaction.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// GetByOrderID method untuk mengambil transaction berdasarkan Order ID
func (repo *TransactionsRepository) GetByOrderID(orderId string) (*types.Transaction, error) {
	query := `
		SELECT id, order_id, product_code, method_code, game_id, zone, voucher_code,
			   whatsapp_number, nickname, username, ip, user_agent,
			   status, purchase_price, web_price, duitku_price, 
			   created_at, updated_at, completed_at
		FROM transactions WHERE order_id = $1`

	var transaction types.Transaction
	err := repo.DB.QueryRow(query, orderId).Scan(
		&transaction.ID,
		&transaction.OrderId,
		&transaction.ProductCode,
		&transaction.MethodCode,
		&transaction.GameId,
		&transaction.Zone,
		&transaction.VoucherCode,
		&transaction.WhatsAppNumber,
		&transaction.Nickname,
		&transaction.Username,
		&transaction.Ip,
		&transaction.UserAgent,
		&transaction.Status,
		&transaction.PurchasePrice,
		&transaction.WebPrice,
		&transaction.DuitkuPrice,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
		&transaction.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// GetByGameID method untuk mengambil transactions berdasarkan game ID
func (repo *TransactionsRepository) GetByGameID(gameId string) ([]types.Transaction, error) {
	query := `
		SELECT id, order_id, product_code, method_code, game_id, zone, voucher_code,
			   whatsapp_number, nickname, username, ip, user_agent,
			   status, purchase_price, web_price, duitku_price, 
			   created_at, updated_at, completed_at
		FROM transactions WHERE game_id = $1
		ORDER BY created_at DESC`

	rows, err := repo.DB.Query(query, gameId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []types.Transaction
	for rows.Next() {
		var transaction types.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.OrderId,
			&transaction.ProductCode,
			&transaction.MethodCode,
			&transaction.GameId,
			&transaction.Zone,
			&transaction.VoucherCode,
			&transaction.WhatsAppNumber,
			&transaction.Nickname,
			&transaction.Username,
			&transaction.Ip,
			&transaction.UserAgent,
			&transaction.Status,
			&transaction.PurchasePrice,
			&transaction.WebPrice,
			&transaction.DuitkuPrice,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&transaction.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

// UpdateStatus method untuk update status transaction
func (repo *TransactionsRepository) UpdateStatus(id int64, status string) error {
	query := `
		UPDATE transactions 
		SET status = $1, updated_at = $2, completed_at = $3
		WHERE id = $4`

	var completedAt *time.Time
	if status == types.StatusSuccess || status == types.StatusFailed || status == types.StatusCancelled {
		now := time.Now()
		completedAt = &now
	}

	_, err := repo.DB.Exec(query, status, time.Now(), completedAt, id)
	return err
}

// GetAll method untuk mengambil semua transactions dengan pagination
func (repo *TransactionsRepository) GetAll(limit, offset int) ([]types.Transaction, error) {
	query := `
		SELECT id, order_id, product_code, method_code, game_id, zone, voucher_code,
			   whatsapp_number, nickname, username, ip, user_agent,
			   status, purchase_price, web_price, duitku_price, 
			   created_at, updated_at, completed_at
		FROM transactions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := repo.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []types.Transaction
	for rows.Next() {
		var transaction types.Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.OrderId,
			&transaction.ProductCode,
			&transaction.MethodCode,
			&transaction.GameId,
			&transaction.Zone,
			&transaction.VoucherCode,
			&transaction.WhatsAppNumber,
			&transaction.Nickname,
			&transaction.Username,
			&transaction.Ip,
			&transaction.UserAgent,
			&transaction.Status,
			&transaction.PurchasePrice,
			&transaction.WebPrice,
			&transaction.DuitkuPrice,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&transaction.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
