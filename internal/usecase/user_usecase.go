package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"golang-crud-clean-arch/internal/entity"
	"golang-crud-clean-arch/internal/repository"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserUsecase struct {
	userRepo *repository.UserRepository
	redis    *redis.Client
}

func NewUserUsecase(userRepo *repository.UserRepository, redis *redis.Client) *UserUsecase {
	return &UserUsecase{userRepo: userRepo, redis: redis}
}

func (u *UserUsecase) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	cacheKey := "users:all"
	cachedData, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var users []entity.User
		if err := json.Unmarshal([]byte(cachedData), &users); err == nil {
			fmt.Println("Mengambil data dari cache Redis")
			return users, nil
		}
	}

	users, err := u.userRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(users)
	u.redis.Set(ctx, cacheKey, jsonData, 10*time.Minute)
	fmt.Println("Mengambil data dari database dan menyimpannya ke Redis")

	return users, nil
}

func (u *UserUsecase) GetUser(ctx context.Context, id primitive.ObjectID) (*entity.User, error) {
	cacheKey := fmt.Sprintf("users:%s", id.Hex())
	cachedData, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var user entity.User
		if err := json.Unmarshal([]byte(cachedData), &user); err == nil {
			fmt.Println("Mengambil user dari cache Redis")
			return &user, nil
		}
	}

	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(user)
	u.redis.Set(ctx, cacheKey, jsonData, 10*time.Minute)
	fmt.Println("Mengambil user dari database dan menyimpannya ke Redis")

	return user, nil
}

func (u *UserUsecase) CreateUser(ctx context.Context, user *entity.User) error {
	return u.userRepo.Create(ctx, user)
}

func (u *UserUsecase) UpdateUser(ctx context.Context, user *entity.User) error {
	return u.userRepo.Update(ctx, user)
}

func (u *UserUsecase) DeleteUser(ctx context.Context, id primitive.ObjectID) error {
	return u.userRepo.Delete(ctx, id)
}
