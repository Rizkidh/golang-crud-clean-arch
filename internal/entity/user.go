package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        interface{} `json:"id" bson:"_id,omitempty"`
	Name      string      `json:"name" bson:"name"`
	Email     string      `json:"email" bson:"email"`
	CreatedAt time.Time   `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" bson:"updated_at"`
}

func (u *User) Validate() error {
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.TrimSpace(u.Email)

	if u.Name == "" {
		return errors.New("name is required")
	}
	if u.Email == "" {
		return errors.New("email is required")
	}
	return nil
}

// Helper methods to handle different ID types
func (u *User) GetIDString() string {
	switch id := u.ID.(type) {
	case primitive.ObjectID:
		return id.Hex()
	case uuid.UUID:
		return id.String()
	default:
		return ""
	}
}
