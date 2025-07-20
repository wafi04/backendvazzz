package subcategory

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/wafi04/backendvazzz/pkg/model"
)

// Models

// Repository
type SubCategoryRepository struct {
	DB *sql.DB
}

func NewSubCategory(DB *sql.DB) *SubCategoryRepository {
	return &SubCategoryRepository{
		DB: DB,
	}
}

// Create - Membuat subcategory baru
func (repo *SubCategoryRepository) Create(ctx context.Context, data model.CreateSubcategory) (*model.SubCategory, error) {
	query := `
		INSERT INTO sub_categories 
		(
			name,
			category_id,
			code,
			status
		) VALUES 
		($1, $2, $3, $4)
		RETURNING id, category_id, code, name, status
	`

	var subCategory model.SubCategory
	err := repo.DB.QueryRowContext(ctx, query, data.Name, data.CategoryId, data.Code, data.Status).
		Scan(&subCategory.Id, &subCategory.CategoryId, &subCategory.Code, &subCategory.Name,
			&subCategory.Status)

	if err != nil {
		log.Printf("Create model.SubCategory error: %v", err)
		return nil, errors.New("failed to create sub category")
	}

	return &subCategory, nil
}

// GetAll - Mendapatkan semua subcategory dengan pagination dan filter
func (repo *SubCategoryRepository) GetAll(ctx context.Context, skip, limit int, search, status string) ([]model.SubCategory, int, error) {
	// Query untuk mendapatkan total count
	countQuery := `
		SELECT COUNT(*) 
		FROM sub_categories
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR code ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR status = $2)
	`

	var totalCount int
	err := repo.DB.QueryRowContext(ctx, countQuery, search, status).Scan(&totalCount)
	if err != nil {
		log.Printf("GetAll SubCategories count error: %v", err)
		return nil, 0, err
	}

	// Query untuk mendapatkan data dengan pagination
	query := `
		SELECT sc.id, sc.category_id, sc.code, sc.name, sc.status
		FROM sub_categories sc
		WHERE ($1 = '' OR sc.name ILIKE '%' || $1 || '%' OR sc.code ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR sc.status = $2)
		LIMIT $3 OFFSET $4
	`

	rows, err := repo.DB.QueryContext(ctx, query, search, status, limit, skip)
	if err != nil {
		log.Printf("GetAll SubCategories error: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var subCategories []model.SubCategory
	for rows.Next() {
		var subCat model.SubCategory
		err := rows.Scan(
			&subCat.Id, &subCat.CategoryId, &subCat.Code, &subCat.Name,
			&subCat.Status,
		)
		if err != nil {
			log.Printf("Scan model.SubCategory error: %v", err)
			continue
		}
		subCategories = append(subCategories, subCat)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return nil, 0, err
	}

	return subCategories, totalCount, nil
}

// GetByCategoryID - Mendapatkan subcategory berdasarkan category ID dengan pagination
func (repo *SubCategoryRepository) GetByCategoryID(ctx context.Context, categoryID, skip, limit int, search, status string) ([]model.SubCategory, int, error) {
	// Query untuk mendapatkan total count
	countQuery := `
		SELECT COUNT(*) 
		FROM sub_categories
		WHERE category_id = $1
		  AND ($2 = '' OR name ILIKE '%' || $2 || '%' OR code ILIKE '%' || $2 || '%')
		  AND ($3 = '' OR status = $3)
	`

	var totalCount int
	err := repo.DB.QueryRowContext(ctx, countQuery, categoryID, search, status).Scan(&totalCount)
	if err != nil {
		log.Printf("GetByCategoryID SubCategories count error: %v", err)
		return nil, 0, err
	}

	query := `
		SELECT sc.id, sc.category_id, sc.code, sc.name, sc.status
		FROM sub_categories sc
		WHERE sc.category_id = $1
		  AND ($2 = '' OR sc.name ILIKE '%' || $2 || '%' OR sc.code ILIKE '%' || $2 || '%')
		  AND ($3 = '' OR sc.status = $3)
		ORDER BY sc.created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := repo.DB.QueryContext(ctx, query, categoryID, search, status, limit, skip)
	if err != nil {
		log.Printf("GetByCategoryID SubCategories error: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var subCategories []model.SubCategory
	for rows.Next() {
		var subCat model.SubCategory
		err := rows.Scan(
			&subCat.Id, &subCat.CategoryId, &subCat.Code, &subCat.Name,
			&subCat.Status,
		)
		if err != nil {
			log.Printf("Scan model.SubCategory error: %v", err)
			continue
		}
		subCategories = append(subCategories, subCat)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return nil, 0, err
	}

	return subCategories, totalCount, nil
}

// GetByID - Mendapatkan subcategory berdasarkan ID
func (repo *SubCategoryRepository) GetByID(ctx context.Context, id int) (*model.SubCategory, error) {
	query := `
		SELECT id, category_id, code, name, status
		FROM sub_categories
		WHERE id = $1
	`

	var subCategory model.SubCategory
	err := repo.DB.QueryRowContext(ctx, query, id).Scan(
		&subCategory.Id, &subCategory.CategoryId, &subCategory.Code, &subCategory.Name,
		&subCategory.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("subcategory not found")
		}
		log.Printf("GetByID model.SubCategory error: %v", err)
		return nil, errors.New("failed to get subcategory")
	}

	return &subCategory, nil
}

// Update - Mengupdate subcategory
func (repo *SubCategoryRepository) Update(ctx context.Context, id int, data model.UpdateSubcategory) (*model.SubCategory, error) {
	// Build dynamic query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if data.CategoryId != nil {
		setParts = append(setParts, "category_id = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *data.CategoryId)
		argIndex++
	}
	if data.Code != nil {
		setParts = append(setParts, "code = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *data.Code)
		argIndex++
	}
	if data.Name != nil {
		setParts = append(setParts, "name = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *data.Name)
		argIndex++
	}
	if data.Status != nil {
		setParts = append(setParts, "status = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, *data.Status)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil, errors.New("no fields to update")
	}

	// Add updated_at and id
	args = append(args, id)

	query := "UPDATE sub_categories SET " + strings.Join(setParts, ", ") +
		" WHERE id = $" + fmt.Sprintf("%d", argIndex) +
		" RETURNING id, category_id, code, name, status"

	var subCategory model.SubCategory
	err := repo.DB.QueryRowContext(ctx, query, args...).Scan(
		&subCategory.Id, &subCategory.CategoryId, &subCategory.Code, &subCategory.Name,
		&subCategory.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("subcategory not found")
		}
		log.Printf("Update model.SubCategory error: %v", err)
		return nil, errors.New("failed to update subcategory")
	}

	return &subCategory, nil
}

func (repo *SubCategoryRepository) Delete(ctx context.Context, id int) error {
	query := `
		UPDATE sub_categories 
		SET status = 'inactive'
		WHERE id = $1
	`

	result, err := repo.DB.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("Delete model.SubCategory error: %v", err)
		return errors.New("failed to delete subcategory")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Delete model.SubCategory rows affected error: %v", err)
		return errors.New("failed to delete subcategory")
	}

	if rowsAffected == 0 {
		return errors.New("subcategory not found")
	}

	return nil
}
