package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/sdk/trace"
)

type HealthHandler struct {
	MongoClient          *mongo.Client
	RedisClient          *redis.Client
	PostgresDB           *sql.DB
	KafkaAddr            string
	JaegerTracerProvider *trace.TracerProvider
}

func NewHealthHandler(mongoClient *mongo.Client, redisClient *redis.Client, postgresDB *sql.DB, kafkaAddr string, tracer *trace.TracerProvider) *HealthHandler {
	return &HealthHandler{
		MongoClient:          mongoClient,
		RedisClient:          redisClient,
		PostgresDB:           postgresDB,
		KafkaAddr:            kafkaAddr,
		JaegerTracerProvider: tracer,
	}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := map[string]string{
		"mongo":    "ok",
		"redis":    "ok",
		"postgres": "ok",
		"kafka":    "ok",
		"jaeger":   "ok",
	}
	overallStatus := http.StatusOK

	// Check MongoDB
	if err := h.MongoClient.Ping(ctx, nil); err != nil {
		status["mongo"] = "unreachable"
		overallStatus = http.StatusServiceUnavailable
	}

	// Check Redis
	if err := h.RedisClient.Ping(ctx).Err(); err != nil {
		status["redis"] = "unreachable"
		overallStatus = http.StatusServiceUnavailable
	}

	// Check PostgreSQL
	if err := h.PostgresDB.PingContext(ctx); err != nil {
		status["postgres"] = "unreachable"
		overallStatus = http.StatusServiceUnavailable
	}

	// Check Kafka
	kafkaConn, err := kafka.DialContext(ctx, "tcp", h.KafkaAddr)
	if err != nil {
		status["kafka"] = "unreachable"
		overallStatus = http.StatusServiceUnavailable
	} else {
		kafkaConn.Close()
	}

	// Check Jaeger (based on tracer presence)
	jaegerURL := "http://jaeger:14268/api/traces"
	req, _ := http.NewRequestWithContext(ctx, "POST", jaegerURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		status["jaeger"] = "unreachable"
		overallStatus = http.StatusServiceUnavailable
	} else {
		status["jaeger"] = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(overallStatus)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": status,
	})
}
