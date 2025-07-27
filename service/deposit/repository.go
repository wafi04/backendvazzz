package deposit

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/wafi04/backendvazzz/pkg/model"
)

type DepositRepository struct {
	Repo *sql.DB
}

func NewDepositRepository(db *sql.DB) *DepositRepository {
	return &DepositRepository{
		Repo: db,
	}
}

// Create - Membuat deposit baru
func (r *DepositRepository) Create(c context.Context, deposit model.CreateDeposit, username string, depositID, logs string) (string, error) {
	tx, err := r.Repo.BeginTx(c, nil)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}()

	var methodName string

	queryMethod := `
		SELECT name
		FROM payment_methods
		WHERE code = $1
	`
	err = tx.QueryRowContext(c, queryMethod,
		deposit.Method).Scan(&methodName)
	if err != nil {
		return "", fmt.Errorf("failed to insert deposit: %w", err)
	}
	query := `INSERT INTO deposits (
		method, 
		amount,
		username,
		deposit_id,
		payment_reference,
		status,
		created_at,
		updated_at,
		log
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	var id int
	err = tx.QueryRowContext(c, query,
		methodName,
		deposit.Amount,
		username,
		depositID,
		deposit.PaymentReference, // payment_reference same as depositID?
		"PENDING",
		time.Now(),
		time.Now(),
		logs).Scan(&id)

	if err != nil {

		return "", fmt.Errorf("failed to insert deposit: %w", err)
	}

	return depositID, nil
}

func (r *DepositRepository) GetByID(c context.Context, id int) (*model.DepositData, error) {
	query := `SELECT id, username, method, deposit_id, payment_reference, amount, status, created_at, updated_at, log 
			  FROM deposits WHERE id = $1`

	var deposit model.DepositData
	var createdAt, updatedAt time.Time

	err := r.Repo.QueryRowContext(c, query, id).Scan(
		&deposit.ID,
		&deposit.Username,
		&deposit.Method,
		&deposit.DepositID,
		&deposit.PaymentReference,
		&deposit.Amount,
		&deposit.Status,
		&createdAt,
		&updatedAt,
		&deposit.Log,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deposit with id %d not found", id)
		}
		return nil, err
	}

	deposit.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
	deposit.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

	return &deposit, nil
}

func (r *DepositRepository) GetByDepositID(c context.Context, depositID string) (*model.DepositData, error) {
	query := `SELECT id, username, method, deposit_id, payment_reference, amount, status, created_at, updated_at, log 
			  FROM deposits WHERE deposit_id = $1`

	var deposit model.DepositData
	var createdAt, updatedAt time.Time

	err := r.Repo.QueryRowContext(c, query, depositID).Scan(
		&deposit.ID,
		&deposit.Username,
		&deposit.Method,
		&deposit.DepositID,
		&deposit.PaymentReference,
		&deposit.Amount,
		&deposit.Status,
		&createdAt,
		&updatedAt,
		&deposit.Log,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deposit with deposit_id %s not found", depositID)
		}
		return nil, err
	}

	deposit.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
	deposit.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

	return &deposit, nil
}

func (r *DepositRepository) GetByDepositByUsername(c context.Context, username string, limit, offset int) ([]model.DepositData, int, error) {
	countQuery := `
		SELECT COUNT(*) 
		FROM deposits
		WHERE ($1 = '' OR username ILIKE '%' || $1 || '%')
	`
	var totalCount int
	err := r.Repo.QueryRowContext(c, countQuery, username).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}
	query := `SELECT id, username, method, deposit_id, payment_reference, amount, status, created_at, updated_at, log 
			  FROM deposits 
			  WHERE ($1 = '' OR username ILIKE '%' || $1 || '%')
				ORDER BY created_at DESC 
				LIMIT $2 OFFSET $3
			  `

	rows, err := r.Repo.QueryContext(c, query, username, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deposits []model.DepositData

	for rows.Next() {
		var deposit model.DepositData
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&deposit.ID,
			&deposit.Username,
			&deposit.Method,
			&deposit.DepositID,
			&deposit.PaymentReference,
			&deposit.Amount,
			&deposit.Status,
			&createdAt,
			&updatedAt,
			&deposit.Log,
		)

		if err != nil {
			return nil, 0, err
		}

		deposit.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		deposit.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

		deposits = append(deposits, deposit)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return deposits, totalCount, nil
}

func (r *DepositRepository) GetAll(c context.Context, limit, offset int, search, status string) ([]model.DepositData, int, error) {
	countQuery := `
		SELECT COUNT(*) 
		FROM deposits
		WHERE ($1 = '' OR username ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR type = $2)
	`

	var totalCount int
	err := r.Repo.QueryRowContext(c, countQuery, search, status).Scan(&totalCount)
	if err != nil {
		log.Printf("GetAll Categories count error: %v", err)
		return nil, 0, err
	}

	query := `SELECT 
				id, 
				username, 
				method, 
				deposit_id, 
				payment_reference, 
				amount, 
				status, 
				created_at, 
				updated_at, 
				log 
			FROM deposits 
			WHERE ($1 = '' OR username ILIKE '%' || $1 || '%')
		  		AND ($2 = '' OR type = $2)
			ORDER BY created_at DESC 
			LIMIT $3 OFFSET $4`

	rows, err := r.Repo.QueryContext(c, query, search, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var deposits []model.DepositData

	for rows.Next() {
		var deposit model.DepositData
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&deposit.ID,
			&deposit.Username,
			&deposit.Method,
			&deposit.DepositID,
			&deposit.PaymentReference,
			&deposit.Amount,
			&deposit.Status,
			&createdAt,
			&updatedAt,
			&deposit.Log,
		)

		if err != nil {
			return nil, 0, err
		}

		deposit.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		deposit.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

		deposits = append(deposits, deposit)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return deposits, totalCount, nil
}

func (r *DepositRepository) Delete(c context.Context, id int) error {
	query := `UPDATE deposits SET status = 'DELETED', updated_at = $1 WHERE id = $2`

	result, err := r.Repo.ExecContext(c, query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deposit with id %d not found", id)
	}

	return nil
}

// GetTotalAmount - Menghitung total amount berdasarkan username dan status
func (r *DepositRepository) GetTotalAmount(c context.Context, username string, status string) (int, error) {
	query := `SELECT COALESCE(SUM(amount), 0) FROM deposits WHERE username = $1 AND status = $2`

	var totalAmount int
	err := r.Repo.QueryRowContext(c, query, username, status).Scan(&totalAmount)
	if err != nil {
		return 0, err
	}

	return totalAmount, nil
}
