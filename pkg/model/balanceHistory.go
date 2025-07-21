package model

import "time"

type BalanceHistory struct {
	ID            int       `json:"id"`
	Username      string    `json:"username"`
	Type          string    `json:"type"`
	OrderID       *string   `json:"orderID"`
	Desc          *string   `json:"description,omitempty"`
	BalanceBefore int       `json:"balanceBefore"`
	BalanceAfter  int       `json:"balanceAfter"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type CreateBalanceHistory struct {
	Username string  `json:"username"`
	Type     string  `json:"type"`
	OrderID  *string `json:"orderID"`
	Desc     *string `json:"description,omitempty"`
	Amount   int     `json:"amount"`
}
