package types

import (
	"context"
	"time"
)

type MethodData struct {
	Id          int64     `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"`
	MinAmount   int64     `json:"minAmount" db:"min_amount"`
	MaxAmount   int64     `json:"maxAmount" db:"max_amount"`
	Fee         int64     `json:"fee" db:"fee"`
	FeeType     string    `json:"feeType" db:"fee_type"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type CreateMethodData struct {
	Code        string `json:"code" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Type        string `json:"type" validate:"required"`
	MinAmount   int64  `json:"minAmount" validate:"min=0"`
	MaxAmount   int64  `json:"maxAmount" validate:"min=0"`
	Fee         int64  `json:"fee" validate:"min=0"`
	FeeType     string `json:"feeType" validate:"required"`
	Active      bool   `json:"active"`
}

type UpdateMethodData struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        *string `json:"type,omitempty"`
	MinAmount   *int64  `json:"minAmount,omitempty"`
	MaxAmount   *int64  `json:"maxAmount,omitempty"`
	Fee         *int64  `json:"fee,omitempty"`
	FeeType     *string `json:"feeType,omitempty"`
	Active      *bool   `json:"active,omitempty"`
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

type MethodRepositoryInterface interface {
	Create(ctx context.Context, req *CreateMethodData) (*MethodData, error)
	GetByID(ctx context.Context, id int64) (*MethodData, error)
	GetByCode(ctx context.Context, code string) (*MethodData, error)
	GetAll(ctx context.Context, limit, offset int) ([]MethodData, error)
	GetActiveOnly(ctx context.Context, limit, offset int) ([]MethodData, error)
	GetByType(ctx context.Context, methodType string) ([]MethodData, error)
	Update(ctx context.Context, id int64, req *UpdateMethodData) (*MethodData, error)
	Delete(ctx context.Context, id int64) error
	UpdateStatus(ctx context.Context, id int64, active bool) error
}

func (m *MethodData) CalculateFee(amount int64) int64 {
	if m.FeeType == FeeTypeFixed {
		return m.Fee
	} else if m.FeeType == FeeTypePercentage {
		return (amount * m.Fee) / 100
	}
	return 0
}

func (m *MethodData) IsValidAmount(amount int64) bool {
	return amount >= m.MinAmount && amount <= m.MaxAmount
}
