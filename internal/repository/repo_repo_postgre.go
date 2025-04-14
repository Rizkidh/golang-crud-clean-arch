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

// RepoRepositoryPostgres adalah struct untuk meng-handle operasi data repository ke PostgreSQL dan Redis
type RepoRepositoryPostgres struct {
	db    *sql.DB
	redis *redis.Client
}

// NewRepoRepositoryPostgres membuat instance baru dari RepoRepositoryPostgres
func NewRepoRepositoryPostgres(db *sql.DB, redis *redis.Client) *RepoRepositoryPostgres {
	return &RepoRepositoryPostgres{db: db, redis: redis}
}

// Helper: parse interface{} ke uuid.UUID
func parseUserIDAsUUID(raw interface{}) (uuid.UUID, error) {
	switch v := raw.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, errors.New("invalid user_id type (must be string or uuid.UUID)")
	}
}

// Create menambahkan data repository baru ke PostgreSQL
func (r *RepoRepositoryPostgres) Create(ctx context.Context, repo *entity.Repository) error {
	id := uuid.New()
	repo.ID = id
	repo.CreatedAt = time.Now()
	repo.UpdatedAt = time.Now()

	// Konversi UserID ke uuid.UUID
	userID, err := parseUserIDAsUUID(repo.UserID)
	if err != nil {
		return fmt.Errorf("UserID parse error: %w", err)
	}

	query := `INSERT INTO repositories (id, user_id, name, url, ai_enabled, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = r.db.ExecContext(ctx, query,
		id, userID, repo.Name, repo.URL, repo.AIEnabled, repo.CreatedAt, repo.UpdatedAt,
	)
	if err == nil {
		r.redis.Del(ctx, "repositories:all")
		fmt.Println("✅ Repository created successfully.")
	}
	return err
}

// GetAllRepositories mengambil semua repository dari PostgreSQL
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
	fmt.Println("✅ Repositories retrieved successfully.")
	return repos, nil
}

// GetByID mengambil repository berdasarkan ID dari PostgreSQL
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
	fmt.Println("✅ Repository retrieved successfully.")
	return &repo, nil
}

// Update memperbarui data repository di PostgreSQL
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
		fmt.Println("✅ Repository updated successfully.")
	}
	return err
}

// Delete menghapus data repository dari PostgreSQL berdasarkan ID
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
		fmt.Println("✅ Repository deleted successfully.")
	}
	return nil
}
