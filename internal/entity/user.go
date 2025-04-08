package entity

import (
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `json:"name" validate:"required"`
	Email     string             `json:"email" validate:"required,email"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// Fungsi untuk validasi struct User
func (u *User) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}
