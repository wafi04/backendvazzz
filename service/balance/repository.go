package balance

import (
	"context"
	"database/sql"
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

func (repo *BalanceRepository) UpsertBalance(ctx context.Context, req model.CreateBalanceHistory) (*model.BalanceHistory, error) {
	currentBalance := 0
	query := `
		SELECT balance_after
		FROM balance_history 
		WHERE username = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := repo.DB.QueryRowContext(ctx, query, req.Username).Scan(&currentBalance)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get current balance: %w", err)
	}

	newBalance := currentBalance + req.Amount
	insertQuery := `
		INSERT INTO balance_history 
		(
			username,
			type,
			order_id,
			description,
			balance_before,
			balance_after,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`

	var result model.BalanceHistory
	err = repo.DB.QueryRowContext(
		ctx,
		insertQuery,
		req.Username,
		req.Type,
		req.OrderID,
		req.Desc,
		currentBalance,
		newBalance,
	).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to insert balance history: %w", err)
	}

	// Fill the result struct
	result.Username = req.Username
	result.Type = req.Type
	result.OrderID = req.OrderID
	result.Desc = req.Desc
	result.BalanceBefore = currentBalance
	result.BalanceAfter = newBalance

	return &result, nil
}

func (repo *BalanceRepository) GetCurrentBalance(ctx context.Context, username string) (int, error) {
	var balance int
	query := `
		SELECT balance_after
		FROM balance_history 
		WHERE username = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := repo.DB.QueryRowContext(ctx, query, username).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current balance: %w", err)
	}

	return balance, nil
}

func (repo *BalanceRepository) GetBalanceHistory(ctx context.Context, username string, limit, offset int) ([]model.BalanceHistory, error) {
	query := `
		SELECT 
			id,
			username,
			type,
			order_id,
			description,
			balance_before,
			balance_after,
			created_at,
			updated_at
		FROM balance_history 
		WHERE username = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := repo.DB.QueryContext(ctx, query, username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance history: %w", err)
	}
	defer rows.Close()

	var histories []model.BalanceHistory
	for rows.Next() {
		var h model.BalanceHistory
		err := rows.Scan(
			&h.ID,
			&h.Username,
			&h.Type,
			&h.OrderID,
			&h.Desc,
			&h.BalanceBefore,
			&h.BalanceAfter,
			&h.CreatedAt,
			&h.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan balance history: %w", err)
		}
		histories = append(histories, h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return histories, nil
}
