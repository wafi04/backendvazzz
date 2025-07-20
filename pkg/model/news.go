package model

import "time"

type News struct {
	ID          int       `json:"id"`
	Path        string    `json:"path"`
	Status      string    `json:"status"`
	Type        string    `json:"type"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateNews struct {
	Path        string  `json:"path"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
	Description *string `json:"description,omitempty"`
}
