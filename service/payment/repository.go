package payment

import (
	"database/sql"
	"time"
)

type PaymentData struct {
	Id            int       `json:"id"`
	OrderId       string    `json:"order_id"`
	Price         int       `json:"price"`
	Fee           int       `json:"fee"`
	Status        string    `json:"status" db:"status"`
	TotalAmount   int       `json:"total_amount"`
	BuyerNumber   int       `json:"buyer_number"`
	PaymentNumber string    `json:"payment_number"`
	Method        string    `json:"method"`
	Reference     string    `json:"reference"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}
type CreatePayment struct {
	OrderId       string `json:"order_id"`
	Price         int    `json:"price"`
	Fee           int    `json:"fee"`
	Status        string `json:"status" db:"status"`
	TotalAmount   int    `json:"total_amount"`
	BuyerNumber   int    `json:"buyer_number"`
	PaymentNumber string `json:"payment_number"`
	Method        string `json:"method"`
	Reference     string `json:"reference"`
}

type PaymentRepository struct {
	DB *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{
		DB: db,
	}
}

func (repo *PaymentRepository) Create(create CreatePayment) error {
	query := `
		INSERT INTO payments (
			order_id, price, fee, status, total_amount, 
			buyer_number, payment_number, method, reference, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()

	_, err := repo.DB.Exec(
		query,
		create.OrderId,
		create.Price,
		create.Fee,
		create.Status,
		create.TotalAmount,
		create.BuyerNumber,
		create.PaymentNumber,
		create.Method,
		create.Reference,
		now,
		now,
	)

	if err != nil {
		return err
	}

	return nil
}
