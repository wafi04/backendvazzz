package model

import "time"

type PaymentDetail struct {
	OrderID       string    `json:"orderId" db:"order_id"`
	Price         string    `json:"price" db:"price"`
	TotalAmount   int       `json:"totalAmount" db:"total_amount"`
	PaymentNumber string    `json:"paymentNumber" db:"payment_number"`
	BuyerNumber   string    `json:"buyerNumber" db:"buyer_number"`
	Fee           int       `json:"fee" db:"fee"`
	FeeAmount     int       `json:"feeAmount" db:"fee_amount"`
	Status        string    `json:"status" db:"status"`
	Method        string    `json:"method" db:"method"`
	Reference     string    `json:"reference" db:"reference"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}
