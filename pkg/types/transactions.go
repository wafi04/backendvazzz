package types

import "time"

type CreateTransactions struct {
	ProductCode    string  `json:"productCode" validate:"required"`
	MethodCode     string  `json:"methodCode" validate:"required"`
	GameId         string  `json:"gameId" validate:"required"`
	Zone           *string `json:"zone,omitempty"`
	VoucherCode    *string `json:"voucherCode,omitempty"`
	WhatsAppNumber string  `json:"whatsAppNumber" validate:"required"`
	Nickname       string  `json:"nickname" validate:"required"`
	Username       *string `json:"username,omitempty"`
	Ip             *string `json:"ip,omitempty"`
	UserAgent      *string `json:"userAgent,omitempty"`
}

type Transaction struct {
	ID             int64      `json:"id" db:"id"`
	OrderId        string     `json:"orderId" db:"order_id"`
	ProductCode    string     `json:"productCode" db:"product_code"`
	MethodCode     string     `json:"methodCode" db:"method_code"`
	GameId         string     `json:"gameId" db:"game_id"`
	Zone           *string    `json:"zone,omitempty" db:"zone"`
	VoucherCode    *string    `json:"voucherCode,omitempty" db:"voucher_code"`
	WhatsAppNumber string     `json:"whatsAppNumber" db:"whatsapp_number"`
	Nickname       string     `json:"nickname" db:"nickname"`
	Username       *string    `json:"username,omitempty" db:"username"`
	Ip             *string    `json:"ip,omitempty" db:"ip"`
	UserAgent      *string    `json:"userAgent,omitempty" db:"user_agent"`
	Status         string     `json:"status" db:"status"`
	PurchasePrice  int64      `json:"purchasePrice" db:"purchase_price"`
	WebPrice       int64      `json:"webPrice" db:"web_price"`
	DuitkuPrice    int64      `json:"duitkuPrice" db:"duitku_price"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
	CompletedAt    *time.Time `json:"completedAt,omitempty" db:"completed_at"`
}

const (
	StatusPending   = "PENDING"
	StatusProcess   = "PROCESS"
	StatusSuccess   = "SUCCESS"
	StatusFailed    = "FAILED"
	StatusCancelled = "CANCELED"
)
