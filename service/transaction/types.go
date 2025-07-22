package transaction

import "errors"

// Domain errors
var (
	ErrServiceNotFound     = errors.New("service not found")
	ErrInvalidRole         = errors.New("invalid user role")
	ErrUsernameRequired    = errors.New("username is required for SALDO payment")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrVoucherInvalid      = errors.New("voucher is invalid or expired")
)

// DTOs
type CreateTransaction struct {
	ProductCode string  `json:"productCode" validate:"required"`
	MethodCode  string  `json:"methodCode" validate:"required"`
	WhatsApp    string  `json:"whatsapp" validate:"required"`
	Username    *string `json:"username,omitempty"`
	Role        *string `json:"role,omitempty"`
	VoucherCode *string `json:"voucherCode,omitempty"`
	GameId      string  `json:"gameId" validate:"required"`
	Zone        *string `json:"zone,omitempty"`
}

type CreateTransactionResponse struct {
	OrderID string `json:"orderId"`
	Total   int    `json:"total"`
	Fee     int    `json:"fee"`
}

// Domain models
type Service struct {
	Price          int    `db:"price"`
	PricePlatinum  int    `db:"price_platinum"`
	PriceReseller  int    `db:"price_reseller"`
	PricePurchase  int    `db:"price_purchase"`
	Profit         int    `db:"profit"`
	ProfitPlatinum int    `db:"profit_platinum"`
	ProfitReseller int    `db:"profit_reseller"`
	ProviderID     string `db:"provider_id"`
	IsProfitFixed  string `db:"is_profit_fixed"`
	ServiceName    string `db:"service_name"`
}

type PricingResult struct {
	UserPrice        int
	UserProfit       int
	UserProfitAmount int
}

type PaymentMethod struct {
	Fee        int    `db:"fee"`
	FeeType    string `db:"fee_type"`
	MethodName string `db:"method_name"`
}
