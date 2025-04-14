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

// UserRepositoryMongo adalah implementasi repository untuk MongoDB
type UserRepositoryMongo struct {
	db       *mongo.Client       // koneksi MongoDB
	redis    *redis.Client       // koneksi Redis untuk cache
	dbName   string              // nama database MongoDB
	validate *validator.Validate // validator untuk validasi struct
}

// NewUserRepositoryMongo membuat instance baru dari UserRepositoryMongo
func NewUserRepositoryMongo(db *mongo.Client, redis *redis.Client, dbName string) *UserRepositoryMongo {
	return &UserRepositoryMongo{
		db:       db,
		redis:    redis,
		dbName:   dbName,
		validate: validator.New(),
	}
}

// Create menyimpan user baru ke MongoDB
func (r *UserRepositoryMongo) Create(ctx context.Context, user *entity.User) error {
	// Validasi struct user
	if err := r.validate.Struct(user); err != nil {
		return errors.New("validation failed: " + err.Error())
	}

	// Set ID dan waktu
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Simpan ke koleksi MongoDB
	collection := r.db.Database(r.dbName).Collection("users")
	_, err := collection.InsertOne(ctx, user)
	if err == nil {
		// Hapus cache agar data tidak stale
		r.redis.Del(ctx, "users:all")
		fmt.Println("✅ User created successfully.")
	}
	return err
}

// GetByID mengambil user berdasarkan ID
func (r *UserRepositoryMongo) GetByID(ctx context.Context, id interface{}) (*entity.User, error) {
	var objectID primitive.ObjectID

	// Konversi ID ke ObjectID
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

	// Query ke MongoDB
	var user entity.User
	collection := r.db.Database(r.dbName).Collection("users")
	err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	fmt.Println("✅ User retrieved successfully.")
	return &user, nil
}

// Update memperbarui data user di MongoDB
func (r *UserRepositoryMongo) Update(ctx context.Context, user *entity.User) error {
	// Validasi data sebelum update
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

	// Jalankan update
	result, err := collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err == nil && result.MatchedCount > 0 {
		// Jika berhasil, hapus cache terkait
		if objectID, ok := user.ID.(primitive.ObjectID); ok {
			r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%s", objectID.Hex()))
		}
		fmt.Println("✅ User updated successfully.")
	}
	return err
}

// Delete menghapus user berdasarkan ID
func (r *UserRepositoryMongo) Delete(ctx context.Context, id interface{}) error {
	var objectID primitive.ObjectID

	// Konversi ID ke ObjectID
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

	// Jalankan delete di MongoDB
	collection := r.db.Database(r.dbName).Collection("users")
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err == nil && result.DeletedCount > 0 {
		// Jika berhasil, hapus cache
		r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%s", objectID.Hex()))
		fmt.Println("✅ User deleted successfully.")
	}
	return err
}

// GetAll mengambil semua data user dari MongoDB
func (r *UserRepositoryMongo) GetAll(ctx context.Context) ([]entity.User, error) {
	collection := r.db.Database(r.dbName).Collection("users")

	// Ambil semua data user
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []entity.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	fmt.Println("✅ All users retrieved successfully.")
	return users, nil
}

// Simulated failure for circuit breaker test
//func (r *UserRepositoryMongo) GetAll(ctx context.Context) ([]entity.User, error) {
//	return nil, fmt.Errorf("simulated db failure for circuit breaker test")
//}
