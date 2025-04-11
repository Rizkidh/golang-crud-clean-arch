package entity

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type Repository struct {
	ID        interface{} `json:"id"`
	UserID    interface{} `json:"user_id" validate:"required"`
	Name      string      `json:"name" validate:"required"`
	URL       string      `json:"url" validate:"required,url"`
	AIEnabled bool        `json:"ai_enabled"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (r *Repository) Validate() error {
	return validator.New().Struct(r)
}
