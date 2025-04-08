package entity

import (
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id,omitempty" json:"user_id" validate:"required"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	URL       string             `bson:"url" json:"url" validate:"required,url"`
	AIEnabled bool               `bson:"ai_enabled" json:"ai_enabled"`
	CreatedAt time.Time          `bson:"created_at,omitempty" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty" json:"updated_at"`
}

// Fungsi untuk validasi struct Repository
func (r *Repository) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
