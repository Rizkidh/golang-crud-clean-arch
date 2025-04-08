package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
)

type HealthHandler struct {
	MongoClient *mongo.Client
	RedisClient *redis.Client
}

func NewHealthHandler(mongoClient *mongo.Client, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		MongoClient: mongoClient,
		RedisClient: redisClient,
	}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("liveness OK"))
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Check MongoDB
	if err := h.MongoClient.Ping(ctx, nil); err != nil {
		http.Error(w, "MongoDB not ready", http.StatusServiceUnavailable)
		return
	}

	// Check Redis
	if err := h.RedisClient.Ping(ctx).Err(); err != nil {
		http.Error(w, "Redis not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("readiness OK"))
}
