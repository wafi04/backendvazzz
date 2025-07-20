package model

import "time"

type UserData struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Username      string    `json:"username"`
	Whatsapp      string    `json:"whatsapp"`
	Balance       int       `json:"balance"`
	Role          string    `json:"role"`
	Otp           *string   `json:"otp,omitempty"`
	Token         string    `json:"token,omitempty"`
	Password      string    `json:"-"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	LastPaymentAt time.Time `json:"lastPaymentAt"`
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	Whatsapp string `json:"whatsapp" validate:"required"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
