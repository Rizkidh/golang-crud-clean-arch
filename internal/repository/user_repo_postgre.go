package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang-crud-clean-arch/internal/entity"
	"golang-crud-clean-arch/internal/event"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// UserRepositoryPostgres adalah struct untuk meng-handle operasi data user ke PostgreSQL dan Redis
type UserRepositoryPostgres struct {
	db        *sql.DB              // koneksi ke database PostgreSQL
	redis     *redis.Client        // koneksi ke Redis
	validate  *validator.Validate  // validasi data dengan go-playground/validator
	publisher event.EventPublisher // publisher untuk mempublikasikan event
}

func NewUserRepositoryPostgres(db *sql.DB, redis *redis.Client, publisher event.EventPublisher) *UserRepositoryPostgres {
	return &UserRepositoryPostgres{
		db:        db,
		redis:     redis,
		validate:  validator.New(),
		publisher: publisher,
	}
}

// Create menambahkan data user baru ke PostgreSQL
func (r *UserRepositoryPostgres) Create(ctx context.Context, user *entity.User) error {
	// Validasi data user sebelum disimpan
	if err := r.validate.Struct(user); err != nil {
		return errors.New("validation failed: " + err.Error())
	}

	// Set ID baru untuk user
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Query untuk menyimpan user ke database
	query := `INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
	if err == nil {
		// Hapus cache Redis jika insert berhasil
		r.redis.Del(ctx, "users:all")
		fmt.Println("✅ User created successfully.")
	}
	return err
}

// GetByID mengambil user berdasarkan ID dari PostgreSQL
func (r *UserRepositoryPostgres) GetByID(ctx context.Context, id interface{}) (*entity.User, error) {
	var user entity.User
	var uuidID uuid.UUID

	// Konversi ID ke uuid.UUID
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

	// Query untuk mencari user berdasarkan ID
	query := `SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, uuidID)

	// Scan hasil query ke dalam struct user
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	fmt.Println("✅ User retrieved successfully.")
	return &user, nil
}

// Update memperbarui data user di PostgreSQL
func (r *UserRepositoryPostgres) Update(ctx context.Context, user *entity.User) error {
	// Validasi data user sebelum update
	if err := r.validate.Struct(user); err != nil {
		return errors.New("validation failed: " + err.Error())
	}

	// Query untuk mengupdate data user
	query := `UPDATE users SET name = $1, email = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.Email, time.Now(), user.ID)
	if err == nil {
		// Hapus cache Redis jika update berhasil
		r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%v", user.ID))
		fmt.Println("✅ User updated successfully.")
	}
	return err
}

// Delete menghapus data user dari PostgreSQL berdasarkan ID
func (r *UserRepositoryPostgres) Delete(ctx context.Context, id interface{}) error {
	var uuidID uuid.UUID

	// Konversi ID ke uuid.UUID
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

	// Query untuk menghapus user berdasarkan ID
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, uuidID)
	if err != nil {
		return err
	}

	// Cek apakah ada baris yang dihapus dan hapus cache jika berhasil
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%v", id))
		fmt.Println("✅ User deleted successfully.")
	}
	return nil
}

// GetAll mengambil semua data user dari PostgreSQL
func (r *UserRepositoryPostgres) GetAll(ctx context.Context) ([]entity.User, error) {
	// Query untuk mengambil semua data user
	query := `SELECT id, name, email, created_at, updated_at FROM users`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Menyimpan hasil query ke slice user
	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	fmt.Println("✅ All users retrieved successfully.")
	return users, nil
}

func (r *UserRepositoryPostgres) PublishEvent(ctx context.Context, eventType string, eventData interface{}) error {
	// Buat event berdasarkan tipe dan data yang diterima
	event := entity.Event{
		Type: eventType,
		Data: eventData,
	}

	// Publikasikan event ke Kafka atau broker event lainnya
	if err := r.publisher.Publish(ctx, "user-events", event.Type, event.Data); err != nil {
		// Jika ada kesalahan saat mempublikasikan event, kembalikan error
		return fmt.Errorf("failed to publish event: %v", err)
	}

	return nil
}
