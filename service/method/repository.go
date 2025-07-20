package method

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/wafi04/backendvazzz/pkg/types"
)

// MethodRepository struct
type MethodRepository struct {
	DB *sql.DB
}

// NewMethodRepository constructor
func NewMethodRepository(DB *sql.DB) *MethodRepository {
	return &MethodRepository{
		DB: DB,
	}
}

// Create method untuk membuat method baru
func (repo *MethodRepository) Create(ctx context.Context, req *types.CreateMethodData) (*types.MethodData, error) {
	fmt.Println(req)
	query := `
		INSERT INTO payment_methods (
			code, name, description, type, min_amount, max_amount, 
			fee, fee_type, status, image, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
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
		req.Status,
		req.Image, // Added image field
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
	method.Status = req.Status
	method.Image = req.Image // Added image field

	return &method, nil
}

// GetByID method untuk mengambil method berdasarkan ID
func (repo *MethodRepository) GetByID(ctx context.Context, id int) (*types.MethodData, error) {
	query := `
		SELECT id, code, name, description, type, min_amount, max_amount,
			   fee, fee_type, status, image, created_at, updated_at
		FROM payment_methods WHERE id = $1`

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
		&method.Status,
		&method.Image, // Added image field
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
			   fee, fee_type, status, image, created_at, updated_at
		FROM payment_methods WHERE code = $1`

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
		&method.Status,
		&method.Image, // Added image field
		&method.CreatedAt,
		&method.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &method, nil
}

func (repo *MethodRepository) GetAll(ctx context.Context, skip, limit int, search, filterType, status string) ([]types.MethodData, int, error) {
	// Debug: Log semua parameter yang masuk

	// Count query - simplified untuk debug
	countQuery := `SELECT COUNT(*) FROM payment_methods`

	var totalCount int
	err := repo.DB.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		log.Printf("Count query error: %v", err)
		return nil, 0, err
	}

	fmt.Printf("Total records in table: %d\n", totalCount)

	// Jika tidak ada filter, ambil semua data dengan pagination
	if search == "" && filterType == "" && status == "" {
		query := `
			SELECT id, code, name, description, type, min_amount, max_amount,
				   fee, fee_type, status, image, created_at, updated_at
			FROM payment_methods 
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`

		rows, err := repo.DB.QueryContext(ctx, query, limit, skip)
		if err != nil {
			log.Printf("Data query error: %v", err)
			return nil, 0, err
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
				&method.Status,
				&method.Image,
				&method.CreatedAt,
				&method.UpdatedAt,
			)
			if err != nil {
				log.Printf("Scan error: %v", err)
				continue
			}
			methods = append(methods, method)
		}

		if err := rows.Err(); err != nil {
			log.Printf("Rows iteration error: %v", err)
			return nil, 0, err
		}

		fmt.Printf("Records found: %d\n", len(methods))
		return methods, totalCount, nil
	}

	// Jika ada filter, gunakan query dengan kondisi
	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if filterType != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, filterType)
		argIndex++
	}

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count query dengan filter
	filteredCountQuery := fmt.Sprintf("SELECT COUNT(*) FROM payment_methods %s", whereClause)

	var filteredCount int
	err = repo.DB.QueryRowContext(ctx, filteredCountQuery, args...).Scan(&filteredCount)
	if err != nil {
		log.Printf("Filtered count query error: %v", err)
		return nil, 0, err
	}

	// Data query dengan filter
	dataQuery := fmt.Sprintf(`
		SELECT id, code, name, description, type, min_amount, max_amount,
			   fee, fee_type, status, image, created_at, updated_at
		FROM payment_methods %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	limitArgs := append(args, limit, skip)

	rows, err := repo.DB.QueryContext(ctx, dataQuery, limitArgs...)
	if err != nil {
		log.Printf("Filtered data query error: %v", err)
		return nil, 0, err
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
			&method.Status,
			&method.Image,
			&method.CreatedAt,
			&method.UpdatedAt,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		methods = append(methods, method)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows iteration error: %v", err)
		return nil, 0, err
	}

	return methods, filteredCount, nil
}

// GetActiveOnly method untuk mengambil methods yang aktif saja
func (repo *MethodRepository) GetActiveOnly(ctx context.Context, limit, offset int) ([]types.MethodData, error) {
	query := `
		SELECT id, code, name, description, type, min_amount, max_amount,
			   fee, fee_type, status, image, created_at, updated_at
		FROM payment_methods 
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
			&method.Status,
			&method.Image, // Added image field
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
			   fee, fee_type, status, image, created_at, updated_at
		FROM payment_methods 
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
			&method.Status,
			&method.Image, // Added image field
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
func (repo *MethodRepository) Update(ctx context.Context, id int, req *types.UpdateMethodData) (*types.MethodData, error) {
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
	if req.Status != nil {
		setParts = append(setParts, "status = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}
	if req.Image != nil {
		setParts = append(setParts, "image = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *req.Image)
		argIndex++
	}

	setParts = append(setParts, "updated_at = $"+fmt.Sprintf("%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE payment_methods  
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
func (repo *MethodRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM payment_methods WHERE id = $1`
	_, err := repo.DB.ExecContext(ctx, query, id)
	return err
}

// UpdateStatus method untuk mengupdate status active
func (repo *MethodRepository) UpdateStatus(ctx context.Context, id int64, status bool) error {
	query := `
		UPDATE payment_methods 
		SET status = $1, updated_at = $2
		WHERE id = $3`

	_, err := repo.DB.ExecContext(ctx, query, status, time.Now(), id)
	return err
}
