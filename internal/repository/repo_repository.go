package repository

import (
	"context"
	"errors"
	"fmt"

	"golang-crud-clean-arch/internal/entity"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RepoRepository adalah struct untuk meng-handle operasi data repository (repo) ke MongoDB dan Redis
type RepoRepository struct {
	db     *mongo.Client // koneksi MongoDB
	redis  *redis.Client // koneksi Redis
	dbName string        // nama database
}

// NewRepoRepository membuat instance baru dari RepoRepository
func NewRepoRepository(db *mongo.Client, redis *redis.Client, dbName string) *RepoRepository {
	return &RepoRepository{db, redis, dbName}
}

// Create menambahkan data repository baru ke MongoDB
func (r *RepoRepository) Create(ctx context.Context, repo *entity.Repository) error {
	collection := r.db.Database(r.dbName).Collection("repo")

	// Buat ID baru untuk repository
	id := primitive.NewObjectID()
	repo.ID = id

	// Simpan ke database
	_, err := collection.InsertOne(ctx, repo)
	if err == nil {
		// Hapus cache jika insert berhasil
		r.redis.Del(ctx, "repositories:all")
		fmt.Println("✅ Repository created successfully.")
	}
	return err
}

// GetAllRepositories mengambil seluruh data repository dari MongoDB
func (r *RepoRepository) GetAllRepositories(ctx context.Context) ([]entity.Repository, error) {
	collection := r.db.Database(r.dbName).Collection("repo")

	// Ambil semua dokumen dari koleksi repo
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode hasil query ke slice entity.Repository
	var repos []entity.Repository
	if err = cursor.All(ctx, &repos); err != nil {
		return nil, err
	}

	fmt.Println("✅ All repositories retrieved successfully.")
	return repos, nil
}

// GetByID mengambil repository berdasarkan ID
func (r *RepoRepository) GetByID(ctx context.Context, id interface{}) (*entity.Repository, error) {
	// Konversi ID ke ObjectID
	objectID, ok := id.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("invalid ID format, expected ObjectID")
	}

	collection := r.db.Database(r.dbName).Collection("repo")

	// Cari data repository berdasarkan ID
	var repo entity.Repository
	err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&repo)
	if err != nil {
		return nil, err
	}

	fmt.Println("✅ Repository retrieved successfully.")
	return &repo, nil
}

// Update memperbarui data repository di MongoDB
func (r *RepoRepository) Update(ctx context.Context, repo *entity.Repository) error {
	// Validasi bahwa ID adalah ObjectID
	objectID, ok := repo.ID.(primitive.ObjectID)
	if !ok {
		return errors.New("invalid ID type (expected ObjectID)")
	}

	collection := r.db.Database(r.dbName).Collection("repo")

	// Siapkan field yang akan diupdate
	update := bson.M{
		"$set": bson.M{
			"name":       repo.Name,
			"url":        repo.URL,
			"ai_enabled": repo.AIEnabled,
			"user_id":    repo.UserID,
		},
	}

	// Jalankan update
	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err == nil && result.MatchedCount > 0 {
		// Hapus cache jika berhasil update
		r.redis.Del(ctx, "repositories:all", fmt.Sprintf("repositories:%s", objectID.Hex()))
		fmt.Println("✅ Repository updated successfully.")
	}
	return err
}

// Delete menghapus repository dari MongoDB berdasarkan ID
func (r *RepoRepository) Delete(ctx context.Context, id interface{}) error {
	// Validasi bahwa ID adalah ObjectID
	objectID, ok := id.(primitive.ObjectID)
	if !ok {
		return errors.New("invalid ID type for Delete (expected ObjectID)")
	}

	collection := r.db.Database(r.dbName).Collection("repo")

	// Hapus dokumen berdasarkan ID
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err == nil && result.DeletedCount > 0 {
		// Hapus cache jika delete berhasil
		r.redis.Del(ctx, "repositories:all", fmt.Sprintf("repositories:%s", objectID.Hex()))
		fmt.Println("✅ Repository deleted successfully.")
	}
	return err
}
