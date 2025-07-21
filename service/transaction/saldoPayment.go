package transaction

import (
	"context"
	"database/sql"
	"fmt"
)

type CreatePaymentUsingSaldo struct {
	Username string
	OrderID  string
	Total    int
	WhatsApp string
	FeeType  string
	Fee      int
	Price    int
	Tx       *sql.Tx
}

type ResponsePaymentSaldo struct {
	Success bool
	OrderID string
}

func (repo *TransactionRepository) PaymentUsingSaldo(c context.Context, req CreatePaymentUsingSaldo) (*ResponsePaymentSaldo, error) {
	insertPaymentQuery := `
		INSERT INTO payments (
			order_id, price, total_amount, buyer_number, fee, 
			fee_amount, status, method
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err := req.Tx.ExecContext(c, insertPaymentQuery,
		req.OrderID,
		fmt.Sprintf("%d", req.Price),
		req.Total,
		req.WhatsApp,
		req.Fee,
		req.Fee,
		"PENDING",
		"SALDO",
	)

	if err != nil {
		return &ResponsePaymentSaldo{
			Success: false,
			OrderID: "",
		}, fmt.Errorf("failed to insert payment: %w", err)
	}

	return &ResponsePaymentSaldo{
		Success: true,
		OrderID: req.OrderID,
	}, nil
}
