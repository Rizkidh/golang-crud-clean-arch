package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang-crud-clean-arch/internal/entity"
	"golang-crud-clean-arch/internal/event"
	"golang-crud-clean-arch/internal/notification"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type UserRepositoryMongo struct {
	db        *mongo.Client
	redis     *redis.Client
	dbName    string
	validate  *validator.Validate
	tracer    trace.Tracer
	publisher event.EventPublisher
}

func NewUserRepositoryMongo(db *mongo.Client, redis *redis.Client, dbName string, publisher event.EventPublisher) *UserRepositoryMongo {
	return &UserRepositoryMongo{
		db:        db,
		redis:     redis,
		dbName:    dbName,
		validate:  validator.New(),
		tracer:    otel.Tracer("user-repository-mongo"),
		publisher: publisher,
	}
}

func (r *UserRepositoryMongo) Create(ctx context.Context, user *entity.User) error {
	ctx, span := r.tracer.Start(ctx, "UserRepositoryMongo.Create")
	defer span.End()

	if err := r.validate.Struct(user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return errors.New("validation failed: " + err.Error())
	}

	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	span.SetAttributes(
		attribute.String("user.id", user.ID.(primitive.ObjectID).Hex()),
		attribute.String("user.email", user.Email),
	)

	collection := r.db.Database(r.dbName).Collection("users")
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "insert failed")
		return err
	}

	r.redis.Del(ctx, "users:all")

	// Publish Kafka event
	eventData := entity.Event{
		Type: "user.created",
		Data: user,
	}
	if err := r.publisher.Publish(ctx, "user-events", eventData.Type, eventData.Data); err != nil {
		return err
	}
	notification.SendTelegramMessage("✅ User created: " + user.Email)
	return nil
}

func (r *UserRepositoryMongo) GetByID(ctx context.Context, id interface{}) (*entity.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepositoryMongo.GetByID")
	defer span.End()

	var objectID primitive.ObjectID
	switch v := id.(type) {
	case string:
		oid, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid ID format")
			return nil, err
		}
		objectID = oid
	case primitive.ObjectID:
		objectID = v
	default:
		err := fmt.Errorf("invalid id type for MongoDB")
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid ID type")
		return nil, err
	}

	span.SetAttributes(attribute.String("user.id", objectID.Hex()))

	var user entity.User
	collection := r.db.Database(r.dbName).Collection("users")
	err := collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "user not found")
		return nil, err
	}

	span.SetAttributes(attribute.String("user.email", user.Email))
	span.SetStatus(codes.Ok, "user fetched")
	return &user, nil
}

func (r *UserRepositoryMongo) Update(ctx context.Context, user *entity.User) error {
	ctx, span := r.tracer.Start(ctx, "UserRepositoryMongo.Update")
	defer span.End()

	if err := r.validate.Struct(user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return errors.New("validation failed: " + err.Error())
	}

	span.SetAttributes(
		attribute.String("user.id", user.ID.(primitive.ObjectID).Hex()),
		attribute.String("user.email", user.Email),
	)

	collection := r.db.Database(r.dbName).Collection("users")
	update := bson.M{
		"$set": bson.M{
			"name":       user.Name,
			"email":      user.Email,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "update failed")
		return err
	}

	if result.MatchedCount > 0 {
		if objectID, ok := user.ID.(primitive.ObjectID); ok {
			r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%s", objectID.Hex()))
		}
	}

	// Publish Kafka event
	eventData := entity.Event{
		Type: "user.updated",
		Data: user,
	}
	if err := r.publisher.Publish(ctx, "user-events", eventData.Type, eventData.Data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "event publish failed")
		return err
	}

	span.SetAttributes(attribute.Int64("matched_count", result.MatchedCount))
	span.SetStatus(codes.Ok, "user updated")
	return nil
}

func (r *UserRepositoryMongo) Delete(ctx context.Context, id interface{}) error {
	ctx, span := r.tracer.Start(ctx, "UserRepositoryMongo.Delete")
	defer span.End()

	var objectID primitive.ObjectID
	switch v := id.(type) {
	case string:
		oid, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid ID format")
			return err
		}
		objectID = oid
	case primitive.ObjectID:
		objectID = v
	default:
		err := fmt.Errorf("invalid id type for MongoDB")
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid ID type")
		return err
	}

	span.SetAttributes(attribute.String("user.id", objectID.Hex()))

	collection := r.db.Database(r.dbName).Collection("users")
	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "delete failed")
		return err
	}

	if result.DeletedCount > 0 {
		r.redis.Del(ctx, "users:all", fmt.Sprintf("users:%s", objectID.Hex()))
	}

	// Publish Kafka event
	eventData := entity.Event{
		Type: "user.deleted",
		Data: map[string]interface{}{"id": objectID.Hex()},
	}
	if err := r.publisher.Publish(ctx, "user-events", eventData.Type, eventData.Data); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "event publish failed")
		return err
	}

	span.SetAttributes(attribute.Int64("deleted_count", result.DeletedCount))
	span.SetStatus(codes.Ok, "user deleted")
	return nil
}

func (r *UserRepositoryMongo) GetAll(ctx context.Context) ([]entity.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepositoryMongo.GetAll")
	defer span.End()

	collection := r.db.Database(r.dbName).Collection("users")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "find failed")
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []entity.User
	if err = cursor.All(ctx, &users); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode failed")
		return nil, err
	}

	span.SetAttributes(attribute.Int("user.count", len(users)))
	span.SetStatus(codes.Ok, "all users fetched")
	return users, nil
}

func (r *UserRepositoryMongo) PublishEvent(ctx context.Context, eventType string, eventData interface{}) error {
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
