package types

import (
	"time"
)

type MethodData struct {
	Id          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Image       string    `json:"image,omitempty" db:"image"`
	Type        string    `json:"type" db:"type"`
	MinAmount   int       `json:"minAmount" db:"min_amount"`
	MaxAmount   int       `json:"maxAmount" db:"max_amount"`
	Fee         *int      `json:"fee,omitempty" db:"fee"`
	FeeType     *string   `json:"feeType,omitempty" validate:"required"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type CreateMethodData struct {
	Code        string  `json:"code" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Type        string  `json:"type" validate:"required"`
	Image       string  `json:"image" db:"image"`
	MinAmount   int     `json:"minAmount" validate:"min=0"`
	MaxAmount   int     `json:"maxAmount" validate:"min=0"`
	Fee         *int    `json:"fee,omitempty" db:"fee"`
	FeeType     *string `json:"feeType,omitempty" validate:"required"`
	Status      string  `json:"status"`
}

type UpdateMethodData struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        *string `json:"type,omitempty"`
	MinAmount   *int    `json:"minAmount,omitempty"`
	Image       *string `json:"image,omitempty" db:"image"`
	MaxAmount   *int    `json:"maxAmount,omitempty"`
	Fee         *int    `json:"fee,omitempty" db:"fee"`
	FeeType     *string `json:"feeType,omitempty"`
	Status      *string `json:"status,omitempty"`
}

const (
	TypeEWallet        = "EWALLET"
	TypeQRIS           = "QRIS"
	TypeVirtualAccount = "VIRTUAL_ACCOUNT"
	TypeRetail         = "CS_STORE"
)

const (
	FeeTypeFixed      = "FIXED"
	FeeTypePercentage = "PERCENTAGE"
)
