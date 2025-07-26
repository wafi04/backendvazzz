package deposit

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/wafi04/backendvazzz/pkg/model"
	"github.com/wafi04/backendvazzz/pkg/utils"
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
func (r *DepositRepository) Create(c context.Context, deposit model.CreateDeposit, username string, log string) (int, error) {
	depStr := "DEP"
	depositID := utils.GenerateUniqeID(&depStr)
	query := `INSERT INTO deposits (
		method, 
		amount,
		username,
		deposit_id,
		payment_reference,
		status,
		created_at,
		updated_at,
		log) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9) RETURNING id`
	var id int
	err := r.Repo.QueryRowContext(c, query, deposit.Method, deposit.Amount, username, depositID, depositID, "PENDING", time.Now(), time.Now(), log).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID - Mengambil deposit berdasarkan ID
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

// GetByDepositID - Mengambil deposit berdasarkan Deposit ID
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

// GetByUsername - Mengambil semua deposit berdasarkan username
func (r *DepositRepository) GetByUsername(c context.Context, username string, limit, offset int) ([]model.DepositData, error) {
	query := `SELECT id, username, method, deposit_id, payment_reference, amount, status, created_at, updated_at, log 
			  FROM deposits WHERE username = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.Repo.QueryContext(c, query, username, limit, offset)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		deposit.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		deposit.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

		deposits = append(deposits, deposit)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return deposits, nil
}

// GetAll - Mengambil semua deposit dengan pagination
func (r *DepositRepository) GetAll(c context.Context, limit, offset int) ([]model.DepositData, error) {
	query := `SELECT id, username, method, deposit_id, payment_reference, amount, status, created_at, updated_at, log 
			  FROM deposits ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.Repo.QueryContext(c, query, limit, offset)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		deposit.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		deposit.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

		deposits = append(deposits, deposit)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return deposits, nil
}

// GetByStatus - Mengambil deposit berdasarkan status
func (r *DepositRepository) GetByStatus(c context.Context, status string, limit, offset int) ([]model.DepositData, error) {
	query := `SELECT id, username, method, deposit_id, payment_reference, amount, status, created_at, updated_at, log 
			  FROM deposits WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.Repo.QueryContext(c, query, status, limit, offset)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		deposit.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
		deposit.UpdatedAt = updatedAt.Format("2006-01-02 15:04:05")

		deposits = append(deposits, deposit)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return deposits, nil
}

// UpdateStatus - Mengupdate status deposit
func (r *DepositRepository) UpdateStatus(c context.Context, id int, status string, log *string) error {
	query := `UPDATE deposits SET status = $1, updated_at = $2, log = $3 WHERE id = $4`

	_, err := r.Repo.ExecContext(c, query, status, time.Now(), log, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateStatusByDepositID - Mengupdate status berdasarkan deposit_id
func (r *DepositRepository) UpdateStatusByDepositID(c context.Context, depositID string, status string, log *string) error {
	query := `UPDATE deposits SET status = $1, updated_at = $2, log = $3 WHERE deposit_id = $4`

	_, err := r.Repo.ExecContext(c, query, status, time.Now(), log, depositID)
	if err != nil {
		return err
	}

	return nil
}

// UpdatePaymentReference - Mengupdate payment reference
func (r *DepositRepository) UpdatePaymentReference(c context.Context, id int, paymentReference string) error {
	query := `UPDATE deposits SET payment_reference = $1, updated_at = $2 WHERE id = $3`

	_, err := r.Repo.ExecContext(c, query, paymentReference, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

// Delete - Menghapus deposit (soft delete dengan mengubah status)
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

// HardDelete - Menghapus deposit secara permanen
func (r *DepositRepository) HardDelete(c context.Context, id int) error {
	query := `DELETE FROM deposits WHERE id = $1`

	result, err := r.Repo.ExecContext(c, query, id)
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

// CountByUsername - Menghitung total deposit berdasarkan username
func (r *DepositRepository) CountByUsername(c context.Context, username string) (int, error) {
	query := `SELECT COUNT(*) FROM deposits WHERE username = $1`

	var count int
	err := r.Repo.QueryRowContext(c, query, username).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CountByStatus - Menghitung total deposit berdasarkan status
func (r *DepositRepository) CountByStatus(c context.Context, status string) (int, error) {
	query := `SELECT COUNT(*) FROM deposits WHERE status = $1`

	var count int
	err := r.Repo.QueryRowContext(c, query, status).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
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
