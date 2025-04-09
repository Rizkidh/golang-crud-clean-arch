// repository/user_repository_postgres.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang-crud-clean-arch/internal/entity"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type UserRepositoryPostgres struct {
	db       *sql.DB
	redis    *redis.Client
	validate *validator.Validate
}

func NewUserRepositoryPostgres(db *sql.DB, redis *redis.Client) *UserRepositoryPostgres {
	return &UserRepositoryPostgres{
		db:       db,
		redis:    redis,
		validate: validator.New(),
	}
}

func (r *UserRepositoryPostgres) Create(ctx context.Context, user *entity.User) error {
	if err := r.validate.Struct(user); err != nil {
		return errors.New("validation failed: " + err.Error())
	}

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
	if err == nil {
		r.redis.Del(ctx, "users:all")
	}
	return err
}

func (r *UserRepositoryPostgres) GetByID(ctx context.Context, id interface{}) (*entity.User, error) {
	var user entity.User
	var uuidID uuid.UUID

	switch v := id.(type) {
	case string:
		var err error
		uuidID, err = uuid.Parse(v)
		if err != nil {
			return nil, err
		}
	case uuid.UUID:
		uuidID = v
	default:
		return nil, fmt.Errorf("invalid id type")
	}

	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, uuidID)

	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) Update(ctx context.Context, user *entity.User) error {
	if err := r.validate.Struct(user); err != nil {
		return errors.New("validation failed: " + err.Error())
	}

	query := `UPDATE users SET name = $1, email = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.Email, time.Now(), user.ID)
	if err == nil {
		r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%v", user.ID))
	}
	return err
}

func (r *UserRepositoryPostgres) Delete(ctx context.Context, id interface{}) error {
	var uuidID uuid.UUID

	switch v := id.(type) {
	case string:
		var err error
		uuidID, err = uuid.Parse(v)
		if err != nil {
			return err
		}
	case uuid.UUID:
		uuidID = v
	default:
		return fmt.Errorf("invalid id type")
	}

	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, uuidID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%v", id))
	}
	return nil
}

func (r *UserRepositoryPostgres) GetAll(ctx context.Context) ([]entity.User, error) {
	query := `SELECT id, name, email, created_at, updated_at FROM users`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
