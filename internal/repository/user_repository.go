package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang-crud-clean-arch/internal/entity"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	db       *mongo.Client
	redis    *redis.Client
	validate *validator.Validate
	dbName   string
}

func NewUserRepository(db *mongo.Client, redis *redis.Client, dbName string) *UserRepository {
	return &UserRepository{
		db:       db,
		redis:    redis,
		validate: validator.New(),
		dbName:   dbName,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	if err := r.validate.Struct(user); err != nil {
		return errors.New("validasi gagal: " + err.Error())
	}

	collection := r.db.Database(r.dbName).Collection("users")
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := collection.InsertOne(ctx, user)
	if err == nil {
		r.redis.Del(ctx, "users:all")
	}
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.User, error) {
	collection := r.db.Database(r.dbName).Collection("users")
	var user entity.User
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	if err := r.validate.Struct(user); err != nil {
		return errors.New("validasi gagal: " + err.Error())
	}

	collection := r.db.Database(r.dbName).Collection("users")
	update := bson.M{
		"$set": bson.M{
			"name":       user.Name,
			"email":      user.Email,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err == nil && result.MatchedCount > 0 {
		r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%s", user.ID.Hex()))
	}
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	collection := r.db.Database(r.dbName).Collection("users")
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err == nil && result.DeletedCount > 0 {
		r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%s", id.Hex()))
	}
	return err
}

func (r *UserRepository) GetAll(ctx context.Context) ([]entity.User, error) {
	collection := r.db.Database(r.dbName).Collection("users")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []entity.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}
