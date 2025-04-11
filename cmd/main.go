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
	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize Jaeger tracing
	cleanup := config.InitTracer("golang-clean-arch")
	defer cleanup()
	fmt.Println("✅ Jaeger tracing initialized")

	// Connect to PostgreSQL
	postgresDB, err := config.PostgresConnect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	// Connect to MongoDB
	mongoClient := config.MongoConnect()
	mongoDBName := os.Getenv("MONGO_DB_NAME")

	// Connect to Redis
	redisClient := config.ConnectRedis()

	// Initialize Kafka publishers
	kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	publisherusers := event.NewKafkaPublisher(kafkaBrokers, "user-events")
	publisherrepos := event.NewKafkaPublisher(kafkaBrokers, "repo-events")
	defer publisherusers.Close()
	fmt.Println("✅ Kafka publisher initialized")

	// --- Repositories ---
	userRepoPostgres := repository.NewUserRepositoryPostgres(postgresDB, redisClient)
	userRepoMongo := repository.NewUserRepositoryMongo(mongoClient, redisClient, mongoDBName)

	repoRepoPostgres := repository.NewRepoRepositoryPostgres(postgresDB, redisClient)
	repoRepoMongo := repository.NewRepoRepository(mongoClient, redisClient, mongoDBName)

	// --- Usecases ---
	userUsecasePostgres := usecase.NewUserUsecase(userRepoPostgres, redisClient, publisherusers)
	userUsecaseMongo := usecase.NewUserUsecase(userRepoMongo, redisClient, publisherusers)

	repoUsecasePostgres := usecase.NewRepositoryUsecase(repoRepoPostgres, redisClient, publisherrepos)
	repoUsecaseMongo := usecase.NewRepositoryUsecase(repoRepoMongo, redisClient, publisherrepos)

	// --- Handlers ---
	userHandlerPostgres := httpHandler.NewUserHandler(userUsecasePostgres)
	userHandlerMongo := httpHandler.NewUserHandler(userUsecaseMongo)

	repoHandlerPostgres := httpHandler.NewRepositoryHandler(repoUsecasePostgres)
	repoHandlerMongo := httpHandler.NewRepositoryHandler(repoUsecaseMongo)

	// --- Router ---
	r := chi.NewRouter()

	r.Route("/pg", func(r chi.Router) {
		routes.SetupUserRoutes(r, userHandlerPostgres)
		routes.SetupRepositoryRoutes(r, repoHandlerPostgres)
	})

	r.Route("/mongo", func(r chi.Router) {
		routes.SetupUserRoutes(r, userHandlerMongo)
		routes.SetupRepositoryRoutes(r, repoHandlerMongo)
	})

	// Root health check
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("🚀 API is running on /pg/* and /mongo/*"))
	})

	fmt.Println("🌍 Server berjalan di port :9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}
