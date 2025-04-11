package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang-crud-clean-arch/internal/entity"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RepoRepositoryPostgres struct {
	db    *sql.DB
	redis *redis.Client
}

func NewRepoRepositoryPostgres(db *sql.DB, redis *redis.Client) *RepoRepositoryPostgres {
	return &RepoRepositoryPostgres{db: db, redis: redis}
}

func (r *RepoRepositoryPostgres) Create(ctx context.Context, repo *entity.Repository) error {
	id := uuid.New()
	repo.ID = id
	repo.CreatedAt = time.Now()
	repo.UpdatedAt = time.Now()

	userID, ok := repo.UserID.(uuid.UUID)
	if !ok {
		return errors.New("invalid UserID type (expected uuid.UUID)")
	}

	query := `INSERT INTO repositories (id, user_id, name, url, ai_enabled, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		id, userID, repo.Name, repo.URL, repo.AIEnabled, repo.CreatedAt, repo.UpdatedAt,
	)
	if err == nil {
		r.redis.Del(ctx, "repositories:all")
	}
	return err
}

func (r *RepoRepositoryPostgres) GetAllRepositories(ctx context.Context) ([]entity.Repository, error) {
	query := `SELECT id, user_id, name, url, ai_enabled, created_at, updated_at FROM repositories`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []entity.Repository
	for rows.Next() {
		var (
			id, userID uuid.UUID
			repo       entity.Repository
		)
		err := rows.Scan(&id, &userID, &repo.Name, &repo.URL, &repo.AIEnabled, &repo.CreatedAt, &repo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		repo.ID = id
		repo.UserID = userID
		repos = append(repos, repo)
	}
	return repos, nil
}

func (r *RepoRepositoryPostgres) GetByID(ctx context.Context, id interface{}) (*entity.Repository, error) {
	var uuidID uuid.UUID
	switch v := id.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, err
		}
		uuidID = parsed
	case uuid.UUID:
		uuidID = v
	default:
		return nil, errors.New("invalid ID type for GetByID")
	}

	query := `SELECT id, user_id, name, url, ai_enabled, created_at, updated_at FROM repositories WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, uuidID)

	var (
		repo   entity.Repository
		userID uuid.UUID
	)
	err := row.Scan(&uuidID, &userID, &repo.Name, &repo.URL, &repo.AIEnabled, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, err
	}

	repo.ID = uuidID
	repo.UserID = userID
	return &repo, nil
}

func (r *RepoRepositoryPostgres) Update(ctx context.Context, repo *entity.Repository) error {
	uuidID, ok := repo.ID.(uuid.UUID)
	if !ok {
		return errors.New("invalid ID type (expected uuid.UUID)")
	}
	repo.UpdatedAt = time.Now()

	query := `UPDATE repositories SET name = $1, url = $2, ai_enabled = $3, updated_at = $4 WHERE id = $5`
	_, err := r.db.ExecContext(ctx, query,
		repo.Name, repo.URL, repo.AIEnabled, repo.UpdatedAt, uuidID,
	)

	if err == nil {
		r.redis.Del(ctx, "repositories:all", fmt.Sprintf("repositories:%v", uuidID))
	}
	return err
}

func (r *RepoRepositoryPostgres) Delete(ctx context.Context, id interface{}) error {
	var uuidID uuid.UUID
	switch v := id.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		uuidID = parsed
	case uuid.UUID:
		uuidID = v
	default:
		return errors.New("invalid ID type for Delete")
	}

	query := `DELETE FROM repositories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, uuidID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.redis.Del(ctx, "repositories:all", fmt.Sprintf("repositories:%v", uuidID))
	}
	return nil
}
