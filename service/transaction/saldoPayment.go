package transaction

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/wafi04/backendvazzz/pkg/lib"
)

type CreatePaymentUsingSaldo struct {
	Username string
	OrderID  string
	Total    int
	WhatsApp string
	NoTujuan string
	Price    int
	Tx       *sql.Tx
}

type ResponsePaymentSaldo struct {
	Success bool
	OrderID string
}

func (repo *TransactionRepository) PaymentUsingSaldo(c context.Context, req CreatePaymentUsingSaldo) (*ResponsePaymentSaldo, error) {
	digiflazz := lib.NewDigiflazzService(lib.DigiConfig{
		DigiKey:      "f99884cd-b12d-5f6e-abf2-90d60f297bda",
		DigiUsername: "casoyeDa3zJg",
	})
	digi, err := digiflazz.TopUp(c, lib.CreateTransactionToDigiflazz{
		Username:     "casoyeDa3zJg",
		BuyerSKUCode: "CHECKIDS",
		CustomerNo:   req.NoTujuan,
		RefID:        req.OrderID,
	})

	fmt.Printf("Digiflazz Response: %+v\n", digi)

	if err != nil {
		return &ResponsePaymentSaldo{
			Success: false,
			OrderID: "",
		}, fmt.Errorf("failed to insert payment: %w", err)
	}

	insertPaymentQuery := `
		INSERT INTO payments (
			order_id, price, total_amount, buyer_number, fee, 
			fee_amount, status, method
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`

	_, err = req.Tx.ExecContext(c, insertPaymentQuery,
		req.OrderID,
		fmt.Sprintf("%d", req.Price),
		req.Total,
		req.WhatsApp,
		0,
		0,
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
