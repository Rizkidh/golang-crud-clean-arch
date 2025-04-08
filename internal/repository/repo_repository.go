package repository

import (
	"context"
	"fmt"

	"golang-crud-clean-arch/internal/entity"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RepoRepository struct {
	db     *mongo.Client
	redis  *redis.Client
	dbName string
}

func NewRepoRepository(db *mongo.Client, redis *redis.Client, dbName string) *RepoRepository {
	return &RepoRepository{
		db:     db,
		redis:  redis,
		dbName: dbName,
	}
}

func (r *RepoRepository) Create(ctx context.Context, repo *entity.Repository) error {
	collection := r.db.Database(r.dbName).Collection("repo")
	repo.ID = primitive.NewObjectID()
	_, err := collection.InsertOne(ctx, repo)
	if err == nil {
		r.redis.Del(ctx, "repositories:all")
	}
	return err
}

func (r *RepoRepository) GetAllRepositories(ctx context.Context) ([]entity.Repository, error) {
	collection := r.db.Database(r.dbName).Collection("repo")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var repos []entity.Repository
	if err = cursor.All(ctx, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func (r *RepoRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entity.Repository, error) {
	collection := r.db.Database(r.dbName).Collection("repo")
	var repo entity.Repository
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (r *RepoRepository) Update(ctx context.Context, repo *entity.Repository) error {
	collection := r.db.Database(r.dbName).Collection("repo")
	update := bson.M{
		"$set": bson.M{
			"name":       repo.Name,
			"url":        repo.URL,
			"ai_enabled": repo.AIEnabled,
			"user_id":    repo.UserID,
		},
	}
	result, err := collection.UpdateOne(ctx, bson.M{"_id": repo.ID}, update)
	if err == nil && result.MatchedCount > 0 {
		r.redis.Del(ctx, "repositories:all", fmt.Sprintf("repositories:%s", repo.ID.Hex()))
	}
	return err
}

func (r *RepoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	collection := r.db.Database(r.dbName).Collection("repo")
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err == nil && result.DeletedCount > 0 {
		r.redis.Del(ctx, "repositories:all", fmt.Sprintf("repositories:%s", id.Hex()))
	}
	return err
}
