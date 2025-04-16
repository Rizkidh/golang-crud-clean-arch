package usecase

import (
	"context"
	"fmt"
	"time"

	"golang-crud-clean-arch/internal/entity"
	"golang-crud-clean-arch/internal/event"
	"golang-crud-clean-arch/internal/notification"

	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id interface{}) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id interface{}) error
	GetAll(ctx context.Context) ([]entity.User, error)
	PublishEvent(ctx context.Context, eventType string, data interface{}) error
}

type UserUsecase struct {
	repo      UserRepository
	redis     *redis.Client
	cb        *gobreaker.CircuitBreaker
	tracer    trace.Tracer
	publisher event.EventPublisher
}

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
		publisher: publisher, // tambahkan publisher di sini
	}
}

func (u *UserUsecase) CreateUser(ctx context.Context, user *entity.User) error {
	ctx, span := u.tracer.Start(ctx, "CreateUser")
	defer span.End()

	span.SetAttributes(attribute.String("operation", "create_user"))

	if err := user.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Validation failed")
		return err
	}

	// Call the repository to create the user and handle Kafka event publishing
	err := u.repo.Create(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Create failed")
		return err
	}

	// Publish event to Kafka
	eventData := entity.Event{
		Type: "user.created",
		Data: user,
	}
	if err := u.repo.PublishEvent(ctx, eventData.Type, eventData.Data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Kafka event publish failed")
		return err
	}

	span.SetAttributes(attribute.String("user.id", fmt.Sprintf("%v", user.ID)))
	span.SetStatus(codes.Ok, "User created")
	return nil
}

func (u *UserUsecase) GetUser(ctx context.Context, id interface{}) (*entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "GetUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "get_user"),
		attribute.String("user.id", fmt.Sprintf("%v", id)),
	)

	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Get failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "User fetched")
	return user, nil
}

func (u *UserUsecase) UpdateUser(ctx context.Context, user *entity.User) error {
	ctx, span := u.tracer.Start(ctx, "UpdateUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "update_user"),
		attribute.String("user.id", fmt.Sprintf("%v", user.ID)),
	)

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

	// Publish event to Kafka
	eventData := entity.Event{
		Type: "user.updated",
		Data: user,
	}
	if err := u.repo.PublishEvent(ctx, eventData.Type, eventData.Data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Kafka event publish failed")
		return err
	}

	span.SetStatus(codes.Ok, "User updated")
	return nil
}

func (u *UserUsecase) DeleteUser(ctx context.Context, id interface{}) error {
	ctx, span := u.tracer.Start(ctx, "DeleteUser")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation", "delete_user"),
		attribute.String("user.id", fmt.Sprintf("%v", id)),
	)

	err := u.repo.Delete(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete failed")
		return err
	}

	// Publish event to Kafka
	eventData := entity.Event{
		Type: "user.deleted",
		Data: map[string]interface{}{"id": id},
	}
	if err := u.repo.PublishEvent(ctx, eventData.Type, eventData.Data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Kafka event publish failed")
		return err
	}

	span.SetStatus(codes.Ok, "User deleted")
	return nil
}

func (u *UserUsecase) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	ctx, span := u.tracer.Start(ctx, "GetAllUsers")
	defer span.End()

	span.SetAttributes(attribute.String("operation", "get_all_users"))

	result, err := u.cb.Execute(func() (interface{}, error) {
		return u.repo.GetAll(ctx)
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Circuit Breaker triggered")

		// Notifikasi Telegram
		notification.SendTelegramMessage("⚠️ Circuit Breaker aktif di GetAllUsers: " + err.Error())

		return nil, fmt.Errorf("user service unavailable: %w", err)
	}

	users, ok := result.([]entity.User)
	if !ok {
		span.SetStatus(codes.Error, "Type assertion failed")
		return nil, fmt.Errorf("user service: type assertion failed")
	}

	span.SetAttributes(attribute.Int("user.count", len(users)))
	span.SetStatus(codes.Ok, "Users fetched")
	return users, nil
}
