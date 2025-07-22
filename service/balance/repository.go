package balance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/wafi04/backendvazzz/pkg/model"
)

type BalanceRepository struct {
	DB *sql.DB
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{
		DB: db,
	}
}

func (repo *BalanceRepository) CreateBalanceHistory(c context.Context, history model.BalanceHistory) (*model.BalanceHistory, error) {
	query := `
	INSERT INTO balance_histories (
		username, 
		balance_use, 
		debit, 
		credit, 
		balance_now, 
		type, 
		description
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, created_at
	`

	var debitValue *int
	if history.Debit != nil && *history.Debit != 0 {
		debitValue = history.Debit
	}

	var creditValue *int
	if history.Credit != nil && *history.Credit != 0 {
		creditValue = history.Credit
	}

	var result model.BalanceHistory = history

	err := repo.DB.QueryRowContext(c, query,
		result.Username,
		result.BalanceUse,
		debitValue,
		creditValue,
		result.BalanceNow,
		result.Type,
		result.Description,
	).Scan(&result.ID, &result.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert balance history: %w", err)
	}

	return &result, nil
}

func (repo *BalanceRepository) GetCurrentBalance(c context.Context, username string) (int, error) {
	query := `
	SELECT COALESCE(balance_now, 0) 
	FROM balance_histories 
	WHERE username = $1 
	ORDER BY created_at DESC 
	LIMIT 1
	`

	var balance int
	err := repo.DB.QueryRowContext(c, query, username).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current balance for %s: %w", username, err)
	}

	return balance, nil
}

func (repo *BalanceRepository) ProcessBalanceTransaction(c context.Context, req model.CreateBalanceHistory) (*model.BalanceHistory, error) {
	// 1. Get current balance
	currentBalance, err := repo.GetCurrentBalance(c, req.Username)
	if err != nil {
		return nil, err
	}

	req.SetTransactionType()

	// Lakukan validasi saldo untuk transaksi debit
	if req.IsDebit && currentBalance < req.Amount {
		return nil, errors.New("insufficient balance")
	}

	balanceHistory := model.BalanceHistory{
		Username:    req.Username,
		BalanceUse:  currentBalance,
		Type:        req.Type,
		Description: req.Description,
	}

	if req.IsDebit {
		balanceHistory.Debit = &req.Amount
		balanceHistory.BalanceNow = currentBalance - req.Amount
	} else {
		balanceHistory.Credit = &req.Amount
		balanceHistory.BalanceNow = currentBalance + req.Amount
	}

	return repo.CreateBalanceHistory(c, balanceHistory)
}

func (repo *BalanceRepository) UpdateByID(c context.Context, id int, req model.UpdateBalanceHistory) error {
	query := `
	UPDATE balance_histories
	SET
		debit = $1,
		credit = $2,
		type = $3,
		description = $4,
		updated_at = NOW()
	WHERE id = $5
	`

	result, err := repo.DB.ExecContext(c, query,
		req.Debit,
		req.Credit,
		req.Type,
		req.Description,
		id)

	if err != nil {
		return fmt.Errorf("failed to update balance history with ID %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after update: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
