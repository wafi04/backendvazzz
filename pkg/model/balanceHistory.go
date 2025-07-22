package model

import "time"

type BalanceHistory struct {
	ID          int        `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	BalanceUse  int        `json:"balanceBefore" db:"balance_use"`
	Debit       *int       `json:"debit,omitempty" db:"debit"`
	Credit      *int       `json:"credit,omitempty" db:"credit"`
	BalanceNow  int        `json:"balanceAfter" db:"balance_now"`
	Type        string     `json:"type" db:"type"`
	Description *string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" db:"updated_at"`
}

type CreateBalanceHistory struct {
	Username    string  `json:"username" validate:"required"`
	Type        string  `json:"type" validate:"required,oneof=topup withdraw transfer purchase refund"`
	OrderID     *string `json:"orderID,omitempty"`
	Description *string `json:"description,omitempty"`
	Amount      int     `json:"amount" validate:"required,gt=0"`
	IsDebit     bool    `json:"-"`
}

func (c *CreateBalanceHistory) SetTransactionType() {
	switch c.Type {
	case "TOPUP", "REFUND":
		c.IsDebit = false
	case "WITHDRAW", "TRANSFER", "PURCHASE":
		c.IsDebit = true
	}
}

func (c *CreateBalanceHistory) GetDebit() *int {
	if c.IsDebit {
		return &c.Amount
	}
	return nil
}

type UpdateBalanceHistory struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Debit       *int      `json:"debit,omitempty"`
	Credit      *int      `json:"credit,omitempty"`
	Type        string    `json:"type"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"` // Untuk kondisi WHERE
}
