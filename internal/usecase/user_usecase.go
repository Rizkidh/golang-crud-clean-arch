package usecase

import (
	"context"
	"errors"
	"golang-crud-clean-arch/internal/entity"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id interface{}) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id interface{}) error
	GetAll(ctx context.Context) ([]entity.User, error)
}

type UserUsecase struct {
	repo  UserRepository
	redis *redis.Client
}

func NewUserUsecase(repo UserRepository, redis *redis.Client) *UserUsecase {
	return &UserUsecase{repo, redis}
}

func (u *UserUsecase) CreateUser(ctx context.Context, user *entity.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	return u.repo.Create(ctx, user)
}

func (u *UserUsecase) GetUser(ctx context.Context, idStr string) (*entity.User, error) {
	// Coba parse ke UUID (PostgreSQL)
	if uuidID, err := uuid.Parse(idStr); err == nil {
		return u.repo.GetByID(ctx, uuidID)
	}

	// Coba parse ke ObjectID (MongoDB)
	if objectID, err := primitive.ObjectIDFromHex(idStr); err == nil {
		return u.repo.GetByID(ctx, objectID)
	}

	return nil, errors.New("invalid ID format")
}

func (u *UserUsecase) UpdateUser(ctx context.Context, user *entity.User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	return u.repo.Update(ctx, user)
}

func (u *UserUsecase) DeleteUser(ctx context.Context, idStr string) error {
	// UUID
	if uuidID, err := uuid.Parse(idStr); err == nil {
		return u.repo.Delete(ctx, uuidID)
	}

	// ObjectID
	if objectID, err := primitive.ObjectIDFromHex(idStr); err == nil {
		return u.repo.Delete(ctx, objectID)
	}

	return errors.New("invalid ID format")
}

func (u *UserUsecase) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	return u.repo.GetAll(ctx)
}
