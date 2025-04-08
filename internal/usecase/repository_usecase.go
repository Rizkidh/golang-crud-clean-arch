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

type RepositoryUsecase struct {
	repoRepo *repository.RepoRepository
	redis    *redis.Client
}

func NewRepositoryUsecase(repoRepo *repository.RepoRepository, redis *redis.Client) *RepositoryUsecase {
	return &RepositoryUsecase{repoRepo: repoRepo, redis: redis}
}

func (u *RepositoryUsecase) CreateRepository(ctx context.Context, repo *entity.Repository) error {
	if err := repo.Validate(); err != nil {
		return err
	}
	repo.ID = primitive.NewObjectID()
	repo.CreatedAt = time.Now()
	repo.UpdatedAt = time.Now()
	if err := u.repoRepo.Create(ctx, repo); err != nil {
		return err
	}
	u.redis.Del(ctx, "repositories:all")
	return nil
}

func (u *RepositoryUsecase) GetAllRepositories(ctx context.Context) ([]entity.Repository, error) {
	cacheKey := "repositories:all"
	cachedData, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var repos []entity.Repository
		if err := json.Unmarshal([]byte(cachedData), &repos); err == nil {
			fmt.Println("Mengambil data dari cache Redis")
			return repos, nil
		}
	}

	repos, err := u.repoRepo.GetAllRepositories(ctx)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(repos)
	u.redis.Set(ctx, cacheKey, jsonData, 10*time.Minute)
	fmt.Println("Mengambil data dari database dan menyimpannya ke Redis")

	return repos, nil
}

func (u *RepositoryUsecase) GetRepository(ctx context.Context, id primitive.ObjectID) (*entity.Repository, error) {
	cacheKey := fmt.Sprintf("repositories:%s", id.Hex())
	cachedData, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var repo entity.Repository
		if err := json.Unmarshal([]byte(cachedData), &repo); err == nil {
			fmt.Println("Mengambil repository dari cache Redis")
			return &repo, nil
		}
	}

	repo, err := u.repoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(repo)
	u.redis.Set(ctx, cacheKey, jsonData, 10*time.Minute)
	fmt.Println("Mengambil repository dari database dan menyimpannya ke Redis")

	return repo, nil
}

func (u *RepositoryUsecase) UpdateRepository(ctx context.Context, repo *entity.Repository) error {
	if err := repo.Validate(); err != nil {
		return err
	}
	repo.UpdatedAt = time.Now()
	if err := u.repoRepo.Update(ctx, repo); err != nil {
		return err
	}
	u.redis.Del(ctx, "repositories:all", fmt.Sprintf("repositories:%s", repo.ID.Hex()))
	return nil
}

func (u *RepositoryUsecase) DeleteRepository(ctx context.Context, id primitive.ObjectID) error {
	if err := u.repoRepo.Delete(ctx, id); err != nil {
		return err
	}
	u.redis.Del(ctx, "repositories:all", fmt.Sprintf("repositories:%s", id.Hex()))
	return nil
}
