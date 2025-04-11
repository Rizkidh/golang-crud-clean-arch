package usecase

import (
	"context"
	"fmt"
	"time"

	"golang-crud-clean-arch/internal/entity"
	"golang-crud-clean-arch/internal/event"

	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type RepoRepository interface {
	Create(ctx context.Context, repo *entity.Repository) error
	GetByID(ctx context.Context, id interface{}) (*entity.Repository, error)
	Update(ctx context.Context, repo *entity.Repository) error
	Delete(ctx context.Context, id interface{}) error
	GetAllRepositories(ctx context.Context) ([]entity.Repository, error)
}

type RepositoryUsecase struct {
	repo      RepoRepository
	redis     *redis.Client
	cb        *gobreaker.CircuitBreaker
	tracer    trace.Tracer
	publisher event.EventPublisher
}

func NewRepositoryUsecase(repo RepoRepository, redis *redis.Client, publisher event.EventPublisher) *RepositoryUsecase {
	cbSettings := gobreaker.Settings{
		Name:        "RepoGetAllBreaker",
		MaxRequests: 1,
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}
	return &RepositoryUsecase{
		repo:      repo,
		redis:     redis,
		cb:        gobreaker.NewCircuitBreaker(cbSettings),
		tracer:    otel.Tracer("repository-usecase"),
		publisher: publisher,
	}
}

func (u *RepositoryUsecase) CreateRepository(ctx context.Context, repo *entity.Repository) error {
	ctx, span := u.tracer.Start(ctx, "CreateRepository")
	defer span.End()

	if err := repo.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Validation failed")
		return err
	}

	err := u.repo.Create(ctx, repo)
	if err == nil {
		_ = u.publisher.Publish(ctx, "repository-events", "repo.created", repo)
	}
	return err
}

func (u *RepositoryUsecase) GetRepository(ctx context.Context, id interface{}) (*entity.Repository, error) {
	ctx, span := u.tracer.Start(ctx, "GetRepository")
	defer span.End()

	repo, err := u.repo.GetByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Not found")
		return nil, err
	}

	span.SetStatus(codes.Ok, "Repository fetched")
	return repo, nil
}

func (u *RepositoryUsecase) UpdateRepository(ctx context.Context, repo *entity.Repository) error {
	ctx, span := u.tracer.Start(ctx, "UpdateRepository")
	defer span.End()

	if err := repo.Validate(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Validation failed")
		return err
	}

	err := u.repo.Update(ctx, repo)
	if err == nil {
		_ = u.publisher.Publish(ctx, "repository-events", "repo.updated", repo)
	}
	return err
}

func (u *RepositoryUsecase) DeleteRepository(ctx context.Context, id interface{}) error {
	ctx, span := u.tracer.Start(ctx, "DeleteRepository")
	defer span.End()

	err := u.repo.Delete(ctx, id)
	if err == nil {
		_ = u.publisher.Publish(ctx, "repository-events", "repo.deleted", map[string]interface{}{"id": id})
	}
	return err
}

func (u *RepositoryUsecase) GetAllRepositories(ctx context.Context) ([]entity.Repository, error) {
	ctx, span := u.tracer.Start(ctx, "GetAllRepositories")
	defer span.End()

	result, err := u.cb.Execute(func() (interface{}, error) {
		return u.repo.GetAllRepositories(ctx)
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Circuit breaker triggered")
		return nil, fmt.Errorf("repository service unavailable: %w", err)
	}

	repos, ok := result.([]entity.Repository)
	if !ok {
		span.SetStatus(codes.Error, "Type assertion failed")
		return nil, fmt.Errorf("repository service: type assertion failed")
	}

	span.SetStatus(codes.Ok, "Repositories retrieved")
	return repos, nil
}
