package types

import "time"

type UserData struct {
	Id            int       `json:"id" db:"id"`
	Name          *string   `json:"name,omitempty" db:"name"`
	Username      string    `json:"username" db:"username"`
	Role          string    `json:"role"`
	PhoneNumber   *string   `json:"phone_number,omitempty"`
	Otp           string    `json:"otp"`
	ApiKey        string    `json:"apikey"`
	Password      string    `json:"password"`
	Token         string    `json:"token,omitempty"`
	Balance       int       `json:"balance"`
	LastPaymentAt time.Time `json:"last_payment_at"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}
