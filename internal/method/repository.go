package method

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/wafi04/backendvazzz/pkg/types"
)

// MethodRepository struct
type MethodRepository struct {
	DB *sql.DB
}

// NewMethodRepository constructor
func NewMethodRepository(DB *sql.DB) types.MethodRepositoryInterface {
	return &MethodRepository{
		DB: DB,
	}
}

// Create method untuk membuat method baru
func (repo *MethodRepository) Create(ctx context.Context, req *types.CreateMethodData) (*types.MethodData, error) {
	query := `
		INSERT INTO methods (
			code, name, description, type, min_amount, max_amount, 
			fee, fee_type, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id, created_at, updated_at`

	var method types.MethodData
	now := time.Now()

	err := repo.DB.QueryRowContext(ctx, query,
		req.Code,
		req.Name,
		req.Description,
		req.Type,
		req.MinAmount,
		req.MaxAmount,
		req.Fee,
		req.FeeType,
		req.Active,
		now,
		now,
	).Scan(&method.Id, &method.CreatedAt, &method.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Fill the method struct with request data
	method.Code = req.Code
	method.Name = req.Name
	method.Description = req.Description
	method.Type = req.Type
	method.MinAmount = req.MinAmount
	method.MaxAmount = req.MaxAmount
	method.Fee = req.Fee
	method.FeeType = req.FeeType
	method.Active = req.Active

	return &method, nil
}

// GetByID method untuk mengambil method berdasarkan ID
func (repo *MethodRepository) GetByID(ctx context.Context, id int64) (*types.MethodData, error) {
	query := `
		SELECT id, code, name, description, type, min_amount, max_amount,
			   fee, fee_type, active, created_at, updated_at
		FROM methods WHERE id = $1`

	var method types.MethodData
	err := repo.DB.QueryRowContext(ctx, query, id).Scan(
		&method.Id,
		&method.Code,
		&method.Name,
		&method.Description,
		&method.Type,
		&method.MinAmount,
		&method.MaxAmount,
		&method.Fee,
		&method.FeeType,
		&method.Active,
		&method.CreatedAt,
		&method.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &method, nil
}

// GetByCode method untuk mengambil method berdasarkan Code
func (repo *MethodRepository) GetByCode(ctx context.Context, code string) (*types.MethodData, error) {
	query := `
		SELECT id, code, name, description, type, min_amount, max_amount,
			   fee, fee_type, active, created_at, updated_at
		FROM methods WHERE code = $1`

	var method types.MethodData
	err := repo.DB.QueryRowContext(ctx, query, code).Scan(
		&method.Id,
		&method.Code,
		&method.Name,
		&method.Description,
		&method.Type,
		&method.MinAmount,
		&method.MaxAmount,
		&method.Fee,
		&method.FeeType,
		&method.Active,
		&method.CreatedAt,
		&method.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &method, nil
}

// GetAll method untuk mengambil semua methods dengan pagination
func (repo *MethodRepository) GetAll(ctx context.Context, limit, offset int) ([]types.MethodData, error) {
	query := `
		SELECT id, code, name, description, type, min_amount, max_amount,
			   fee, fee_type, active, created_at, updated_at
		FROM methods
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := repo.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []types.MethodData
	for rows.Next() {
		var method types.MethodData
		err := rows.Scan(
			&method.Id,
			&method.Code,
			&method.Name,
			&method.Description,
			&method.Type,
			&method.MinAmount,
			&method.MaxAmount,
			&method.Fee,
			&method.FeeType,
			&method.Active,
			&method.CreatedAt,
			&method.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}

	return methods, nil
}

// GetActiveOnly method untuk mengambil methods yang aktif saja
func (repo *MethodRepository) GetActiveOnly(ctx context.Context, limit, offset int) ([]types.MethodData, error) {
	query := `
		SELECT id, code, name, description, type, min_amount, max_amount,
			   fee, fee_type, active, created_at, updated_at
		FROM methods
		WHERE active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := repo.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []types.MethodData
	for rows.Next() {
		var method types.MethodData
		err := rows.Scan(
			&method.Id,
			&method.Code,
			&method.Name,
			&method.Description,
			&method.Type,
			&method.MinAmount,
			&method.MaxAmount,
			&method.Fee,
			&method.FeeType,
			&method.Active,
			&method.CreatedAt,
			&method.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}

	return methods, nil
}

// GetByType method untuk mengambil methods berdasarkan type
func (repo *MethodRepository) GetByType(ctx context.Context, methodType string) ([]types.MethodData, error) {
	query := `
		SELECT id, code, name, description, type, min_amount, max_amount,
			   fee, fee_type, active, created_at, updated_at
		FROM methods
		WHERE type = $1 AND active = true
		ORDER BY name ASC`

	rows, err := repo.DB.QueryContext(ctx, query, methodType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []types.MethodData
	for rows.Next() {
		var method types.MethodData
		err := rows.Scan(
			&method.Id,
			&method.Code,
			&method.Name,
			&method.Description,
			&method.Type,
			&method.MinAmount,
			&method.MaxAmount,
			&method.Fee,
			&method.FeeType,
			&method.Active,
			&method.CreatedAt,
			&method.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}

	return methods, nil
}

// Update method untuk mengupdate method
func (repo *MethodRepository) Update(ctx context.Context, id int64, req *types.UpdateMethodData) (*types.MethodData, error) {
	// Build dynamic query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		setParts = append(setParts, "name = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Description != nil {
		setParts = append(setParts, "description = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.Type != nil {
		setParts = append(setParts, "type = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.Type)
		argIndex++
	}
	if req.MinAmount != nil {
		setParts = append(setParts, "min_amount = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.MinAmount)
		argIndex++
	}
	if req.MaxAmount != nil {
		setParts = append(setParts, "max_amount = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.MaxAmount)
		argIndex++
	}
	if req.Fee != nil {
		setParts = append(setParts, "fee = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.Fee)
		argIndex++
	}
	if req.FeeType != nil {
		setParts = append(setParts, "fee_type = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.FeeType)
		argIndex++
	}
	if req.Active != nil {
		setParts = append(setParts, "active = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.Active)
		argIndex++
	}

	setParts = append(setParts, "updated_at = $"+fmt.Sprintf("%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE methods 
		SET %s
		WHERE id = $%d`,
		strings.Join(setParts, ", "),
		argIndex)

	_, err := repo.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Get updated record
	return repo.GetByID(ctx, id)
}

// Delete method untuk menghapus method
func (repo *MethodRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM methods WHERE id = $1`
	_, err := repo.DB.ExecContext(ctx, query, id)
	return err
}

// UpdateStatus method untuk mengupdate status active
func (repo *MethodRepository) UpdateStatus(ctx context.Context, id int64, active bool) error {
	query := `
		UPDATE methods 
		SET active = $1, updated_at = $2
		WHERE id = $3`

	_, err := repo.DB.ExecContext(ctx, query, active, time.Now(), id)
	return err
}
