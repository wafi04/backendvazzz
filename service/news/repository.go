package news

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/wafi04/backendvazzz/pkg/model"
)

type NewsRepository struct {
	DB *sql.DB
}

func NewNewsRepository(db *sql.DB) *NewsRepository {
	return &NewsRepository{DB: db}
}

// ✅ Create News
func (repo *NewsRepository) Create(input *model.CreateNews) (*model.News, error) {
	now := time.Now()
	query := `
		INSERT INTO news (path, status, type, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, path, status, type, description, created_at, updated_at
	`
	var news model.News
	err := repo.DB.QueryRow(
		query,
		input.Path,
		input.Status,
		input.Type,
		input.Description,
		now,
		now,
	).Scan(
		&news.ID,
		&news.Path,
		&news.Status,
		&news.Type,
		&news.Description,
		&news.CreatedAt,
		&news.UpdatedAt,
	)
	return &news, err
}

// ✅ Get All News with optional filters
func (repo *NewsRepository) GetAll(status, newsType *string) ([]model.News, error) {
	baseQuery := `
		SELECT id, path, status, type, description, created_at, updated_at
		FROM news
		WHERE 1=1
	`
	var args []interface{}
	var filters []string

	if status != nil {
		args = append(args, *status)
		filters = append(filters, fmt.Sprintf("AND status = $%d", len(args)))
	}
	if newsType != nil {
		args = append(args, *newsType)
		filters = append(filters, fmt.Sprintf("AND type = $%d", len(args)))
	}

	finalQuery := baseQuery + " " + strings.Join(filters, " ") + " ORDER BY created_at DESC"
	rows, err := repo.DB.Query(finalQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allNews []model.News
	for rows.Next() {
		var n model.News
		err := rows.Scan(
			&n.ID,
			&n.Path,
			&n.Status,
			&n.Type,
			&n.Description,
			&n.CreatedAt,
			&n.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		allNews = append(allNews, n)
	}

	return allNews, nil
}

// ✅ Get News By ID
func (repo *NewsRepository) GetByID(id int) (*model.News, error) {
	query := `
		SELECT id, path, status, type, description, created_at, updated_at
		FROM news
		WHERE id = $1
	`
	var n model.News
	err := repo.DB.QueryRow(query, id).Scan(
		&n.ID,
		&n.Path,
		&n.Status,
		&n.Type,
		&n.Description,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &n, err
}

// ✅ Update News
func (repo *NewsRepository) Update(id int, input *model.CreateNews) (*model.News, error) {
	query := `
		UPDATE news
		SET path = $1, status = $2, type = $3, description = $4, updated_at = $5
		WHERE id = $6
		RETURNING id, path, status, type, description, created_at, updated_at
	`
	var n model.News
	err := repo.DB.QueryRow(
		query,
		input.Path,
		input.Status,
		input.Type,
		input.Description,
		time.Now(),
		id,
	).Scan(
		&n.ID,
		&n.Path,
		&n.Status,
		&n.Type,
		&n.Description,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	return &n, err
}

// ✅ Delete News
func (repo *NewsRepository) Delete(id int) error {
	query := `DELETE FROM news WHERE id = $1`
	_, err := repo.DB.Exec(query, id)
	return err
}
