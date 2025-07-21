package model

import "time"

type Transaction struct {
	ID                int       `json:"id" db:"id"`
	OrderID           string    `json:"orderId" db:"order_id"`
	Username          *string   `json:"username,omitempty" db:"username"`
	PurchasePrice     int       `json:"purchasePrice" db:"purchase_price"`
	Discount          int       `json:"discount" db:"discount"`
	UserID            string    `json:"userId" db:"user_id"`
	Zone              string    `json:"zone" db:"zone"`
	Nickname          *string   `json:"nickname,omitempty" db:"nickname"`
	ServiceName       string    `json:"serviceName" db:"service_name"`
	Price             int       `json:"price" db:"price"`
	Profit            int       `json:"profit" db:"profit"`
	Message           *string   `json:"message,omitempty" db:"message"`
	ProfitAmount      int       `json:"profitAmount" db:"profit_amount"`
	ProviderOrderID   *string   `json:"providerOrderId,omitempty" db:"provider_order_id"`
	Status            string    `json:"status" db:"status"`
	Log               *string   `json:"log,omitempty" db:"log"`
	SerialNumber      *string   `json:"serialNumber,omitempty" db:"serial_number"`
	IsReOrder         *string   `json:"isReOrder" db:"is_re_order"`
	TransactionType   string    `json:"transactionType" db:"transaction_type"`
	IsDigi            *string   `json:"isDigi" db:"is_digi"`
	RefID             *string   `json:"refId" db:"ref_id"`
	SuccessReportSent string    `json:"successReportSent" db:"success_report_sent"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time `json:"updatedAt" db:"updated_at"`
}

type TransactionWithPayment struct {
	// Transaction fields
	ID                int       `json:"id" db:"id"`
	OrderID           string    `json:"orderId" db:"order_id"`
	Username          *string   `json:"username,omitempty" db:"username"`
	PurchasePrice     int       `json:"purchasePrice" db:"purchase_price"`
	Discount          int       `json:"discount" db:"discount"`
	UserID            string    `json:"userId" db:"user_id"`
	Zone              string    `json:"zone" db:"zone"`
	Nickname          *string   `json:"nickname,omitempty" db:"nickname"`
	ServiceName       string    `json:"serviceName" db:"service_name"`
	Price             int       `json:"price" db:"price"`
	Profit            int       `json:"profit" db:"profit"`
	Message           *string   `json:"message,omitempty" db:"message"`
	ProfitAmount      int       `json:"profitAmount" db:"profit_amount"`
	ProviderOrderID   *string   `json:"providerOrderId,omitempty" db:"provider_order_id"`
	Status            string    `json:"status" db:"status"`
	Log               *string   `json:"log,omitempty" db:"log"`
	SerialNumber      *string   `json:"serialNumber,omitempty" db:"serial_number"`
	IsReOrder         *string   `json:"isReOrder" db:"is_re_order"`
	TransactionType   string    `json:"transactionType" db:"transaction_type"`
	IsDigi            *string   `json:"isDigi" db:"is_digi"`
	RefID             *string   `json:"refId" db:"ref_id"`
	SuccessReportSent string    `json:"successReportSent" db:"success_report_sent"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time `json:"updatedAt" db:"updated_at"`

	// Payment details (nullable jika LEFT JOIN)
	PaymentDetail *PaymentDetail `json:"paymentDetail,omitempty"`
}

type FilterTransaction struct {
	Limit     int     `json:"limit"`
	Page      int     `json:"page"`
	Type      string  `json:"type"`
	Search    *string `json:"search,omitempty"`
	StartDate *string `json:"startDate,omitempty"`
	EndDate   *string `json:"endDate,omitempty"`
	Status    *string `json:"status,omitempty"`
}
