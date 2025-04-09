// repository/user_repository_mongo.go
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

type UserRepositoryMongo struct {
	db       *mongo.Client
	redis    *redis.Client
	dbName   string
	validate *validator.Validate
}

func NewUserRepositoryMongo(db *mongo.Client, redis *redis.Client, dbName string) *UserRepositoryMongo {
	return &UserRepositoryMongo{
		db:       db,
		redis:    redis,
		dbName:   dbName,
		validate: validator.New(),
	}
}

func (r *UserRepositoryMongo) Create(ctx context.Context, user *entity.User) error {
	if err := r.validate.Struct(user); err != nil {
		return errors.New("validation failed: " + err.Error())
	}

	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	collection := r.db.Database(r.dbName).Collection("users")
	_, err := collection.InsertOne(ctx, user)
	if err == nil {
		r.redis.Del(ctx, "users:all")
	}
	return err
}

func (r *UserRepositoryMongo) GetByID(ctx context.Context, id interface{}) (*entity.User, error) {
	var objectID primitive.ObjectID
	switch v := id.(type) {
	case string:
		oid, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return nil, err
		}
		objectID = oid
	case primitive.ObjectID:
		objectID = v
	default:
		return nil, fmt.Errorf("invalid id type for MongoDB")
	}

	var user entity.User
	collection := r.db.Database(r.dbName).Collection("users")
	err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryMongo) Update(ctx context.Context, user *entity.User) error {
	if err := r.validate.Struct(user); err != nil {
		return errors.New("validation failed: " + err.Error())
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
		if objectID, ok := user.ID.(primitive.ObjectID); ok {
			r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%s", objectID.Hex()))
		}
	}
	return err
}

func (r *UserRepositoryMongo) Delete(ctx context.Context, id interface{}) error {
	var objectID primitive.ObjectID
	switch v := id.(type) {
	case string:
		oid, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return err
		}
		objectID = oid
	case primitive.ObjectID:
		objectID = v
	default:
		return fmt.Errorf("invalid id type for MongoDB")
	}

	collection := r.db.Database(r.dbName).Collection("users")
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err == nil && result.DeletedCount > 0 {
		r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%s", objectID.Hex()))
	}
	return err
}

func (r *UserRepositoryMongo) GetAll(ctx context.Context) ([]entity.User, error) {
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
