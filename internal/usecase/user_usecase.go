// internal/usecase/user_usecase.go
package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang-crud-clean-arch/internal/entity"
	"golang-crud-clean-arch/internal/event"

	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Interface untuk UserRepository, mendukung berbagai jenis database (PostgreSQL, MongoDB, dll)
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id interface{}) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id interface{}) error
	GetAll(ctx context.Context) ([]entity.User, error)
}

// UserUsecase adalah struct utama untuk semua logika bisnis terkait user
type UserUsecase struct {
	repo      UserRepository
	redis     *redis.Client
	cb        *gobreaker.CircuitBreaker
	tracer    trace.Tracer
	publisher event.EventPublisher
}

// NewUserUsecase membuat instance baru UserUsecase lengkap dengan circuit breaker & tracer
func NewUserUsecase(repo UserRepository, redis *redis.Client, publisher event.EventPublisher) *UserUsecase {
	cbSettings := gobreaker.Settings{
		Name:        "UserGetAllBreaker",
		MaxRequests: 1,
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	return &UserUsecase{
		repo:      repo,
		redis:     redis,
		cb:        gobreaker.NewCircuitBreaker(cbSettings),
		tracer:    otel.Tracer("user-usecase"),
		publisher: publisher,
	}
}

// CreateUser menyimpan user baru ke database dan mem-publish event ke Kafka
func (u *UserUsecase) CreateUser(ctx context.Context, user *entity.User) error {
	ctx, span := u.tracer.Start(ctx, "CreateUser")
	defer span.End()

	if err := user.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Validation failed")
		return err
	}

	err := u.repo.Create(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Create failed")
		return err
	}

	// Publish event ke Kafka
	eventData := entity.Event{
		Type: "user.created",
		Data: user,
	}
	if err := u.publisher.Publish(ctx, "user-events", eventData.Type, eventData.Data); err != nil {
		log.Printf("❌ Failed to publish Kafka event: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Kafka publish failed")
	} else {
		log.Println("✅ Kafka event published: user.created")
		span.SetStatus(codes.Ok, "User created & event published")
	}

	return nil
}

// GetUser mengambil user berdasarkan ID dari repository
func (u *UserUsecase) GetUser(ctx context.Context, id interface{}) (*entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "GetUser")
	defer span.End()

	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Get failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "User fetched")
	return user, nil
}

// UpdateUser mengupdate data user di database dan publish event ke Kafka
func (u *UserUsecase) UpdateUser(ctx context.Context, user *entity.User) error {
	ctx, span := u.tracer.Start(ctx, "UpdateUser")
	defer span.End()

	if err := user.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Validation failed")
		return err
	}

	err := u.repo.Update(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Update failed")
		return err
	}

	// Publish Kafka event
	eventData := entity.Event{
		Type: "user.updated",
		Data: user,
	}
	if err := u.publisher.Publish(ctx, "user-events", eventData.Type, eventData.Data); err != nil {
		log.Printf("❌ Failed to publish Kafka event: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Kafka publish failed")
	} else {
		log.Println("✅ Kafka event published: user.updated")
		span.SetStatus(codes.Ok, "User updated & event published")
	}

	return nil
}

// DeleteUser menghapus user dari database dan publish event ke Kafka
func (u *UserUsecase) DeleteUser(ctx context.Context, id interface{}) error {
	ctx, span := u.tracer.Start(ctx, "DeleteUser")
	defer span.End()

	err := u.repo.Delete(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete failed")
		return err
	}

	// Publish Kafka event
	eventData := entity.Event{
		Type: "user.deleted",
		Data: map[string]interface{}{"id": id},
	}
	if err := u.publisher.Publish(ctx, "user-events", eventData.Type, eventData.Data); err != nil {
		log.Printf("❌ Failed to publish Kafka event: %v", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Kafka publish failed")
	} else {
		log.Println("✅ Kafka event published: user.deleted")
		span.SetStatus(codes.Ok, "User deleted & event published")
	}

	return nil
}

// GetAllUsers mengambil semua user dengan circuit breaker untuk menangani kegagalan
func (u *UserUsecase) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "GetAllUsers")
	defer span.End()

	result, err := u.cb.Execute(func() (interface{}, error) {
		return u.repo.GetAll(ctx)
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Circuit Breaker triggered")
		return nil, fmt.Errorf("user service unavailable: %w", err)
	}

	users, ok := result.([]entity.User)
	if !ok {
		span.SetStatus(codes.Error, "Type assertion failed")
		return nil, fmt.Errorf("user service: type assertion failed")
	}

	span.SetStatus(codes.Ok, "Users fetched")
	return users, nil
}
