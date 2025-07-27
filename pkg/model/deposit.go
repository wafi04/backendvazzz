package model

type DepositData struct {
	ID               int     `json:"id"`
	Username         string  `json:"username"`
	Method           string  `json:"method"`
	DepositID        string  `json:"deposit_id"`
	PaymentReference string  `json:"payment_reference"`
	Amount           int     `json:"amount"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	Log              *string `json:"log,omitempty"`
}

type CreateDeposit struct {
	PaymentReference string `json:"payment_reference"`
	Method           string `json:"method" validate:"required"`
	Amount           int    `json:"amount" validate:"required"`
}
