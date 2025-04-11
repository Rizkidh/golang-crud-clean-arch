package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang-crud-clean-arch/config"
	httpHandler "golang-crud-clean-arch/delivery/http"
	"golang-crud-clean-arch/delivery/routes"
	"golang-crud-clean-arch/internal/event"
	"golang-crud-clean-arch/internal/repository"
	"golang-crud-clean-arch/internal/usecase"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/otel"
)

func main() {
	// ✅ Inisialisasi Jaeger Tracer
	cleanup, tracerProvider := config.InitTracerWithProvider("golang-clean-arch")
	defer cleanup()
	otel.SetTracerProvider(tracerProvider)

	// PostgreSQL
	postgresDB, err := config.PostgresConnect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	// MongoDB
	mongoClient := config.MongoConnect()
	mongoDBName := os.Getenv("MONGO_DB_NAME")

	// Redis
	redisClient := config.ConnectRedis()

	// Kafka
	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	kafkaAddr := kafkaBrokers[0]

	publisherUsers := event.NewKafkaPublisher(kafkaBrokers, "user-events")
	publisherRepos := event.NewKafkaPublisher(kafkaBrokers, "repo-events")
	defer publisherUsers.Close()
	defer publisherRepos.Close()
	fmt.Println("✅ Kafka publisher initialized")

	// Repositories
	userRepoPostgres := repository.NewUserRepositoryPostgres(postgresDB, redisClient)
	userRepoMongo := repository.NewUserRepositoryMongo(mongoClient, redisClient, mongoDBName)
	repoRepoPostgres := repository.NewRepoRepositoryPostgres(postgresDB, redisClient)
	repoRepoMongo := repository.NewRepoRepository(mongoClient, redisClient, mongoDBName)

	// Usecases
	userUsecasePostgres := usecase.NewUserUsecase(userRepoPostgres, redisClient, publisherUsers)
	userUsecaseMongo := usecase.NewUserUsecase(userRepoMongo, redisClient, publisherUsers)
	repoUsecasePostgres := usecase.NewRepositoryUsecase(repoRepoPostgres, redisClient, publisherRepos)
	repoUsecaseMongo := usecase.NewRepositoryUsecase(repoRepoMongo, redisClient, publisherRepos)

	// Handlers
	userHandlerPostgres := httpHandler.NewUserHandler(userUsecasePostgres)
	userHandlerMongo := httpHandler.NewUserHandler(userUsecaseMongo)
	repoHandlerPostgres := httpHandler.NewRepositoryHandler(repoUsecasePostgres)
	repoHandlerMongo := httpHandler.NewRepositoryHandler(repoUsecaseMongo)

	// ✅ Health Handler with Jaeger tracer info
	healthHandler := httpHandler.NewHealthHandler(mongoClient, redisClient, postgresDB, kafkaAddr, tracerProvider)

	// Router
	r := chi.NewRouter()

	r.Route("/pg", func(r chi.Router) {
		routes.SetupUserRoutes(r, userHandlerPostgres)
		routes.SetupRepositoryRoutes(r, repoHandlerPostgres)
	})

	r.Route("/mongo", func(r chi.Router) {
		routes.SetupUserRoutes(r, userHandlerMongo)
		routes.SetupRepositoryRoutes(r, repoHandlerMongo)
	})

	// Healthcheck
	routes.SetupHealthRoutes(r, healthHandler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("🚀 API is running on /pg/* and /mongo/*"))
	})

	fmt.Println("🌍 Server berjalan di port :9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}
