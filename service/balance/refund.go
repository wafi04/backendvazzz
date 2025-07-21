package balance

import (
	"context"
	"database/sql"
	"fmt"
)

func RefundSaldoWithSum(ctx context.Context, amount int, username string, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO balance_history 
		(username, type, description, balance_before, balance_after, amount)
		SELECT 
			$1,
			'REFUND',
			'Balance refund',
			COALESCE((SELECT balance_after FROM balance_history WHERE username = $1 ORDER BY created_at DESC LIMIT 1), 0),
			COALESCE((SELECT balance_after FROM balance_history WHERE username = $1 ORDER BY created_at DESC LIMIT 1), 0) + $2,
			$2
	`, username, amount)

	if err != nil {
		return fmt.Errorf("failed to insert refund history: %w", err)
	}

	query := `
		UPDATE users 
		SET balance = (
			SELECT COALESCE(SUM(amount), 0) 
			FROM balance_history 
			WHERE username = $1
		)
		WHERE username = $1
	`

	_, err = db.ExecContext(ctx, query, username)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}
