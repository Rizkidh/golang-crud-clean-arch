package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang-crud-clean-arch/config"
	httpHandler "golang-crud-clean-arch/delivery/http"
	"golang-crud-clean-arch/delivery/routes"
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

	// Connect to PostgreSQL
	postgresDB, err := config.PostgresConnect()
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresDB.Close()

	// Connect to MongoDB
	mongoClient := config.MongoConnect()
	mongoDBName := os.Getenv("MONGO_DB_NAME")

	// Connect to Redis
	redisClient := config.ConnectRedis()

	// --- Repositories ---
	userRepoPostgres := repository.NewUserRepositoryPostgres(postgresDB, redisClient)
	userRepoMongo := repository.NewUserRepositoryMongo(mongoClient, redisClient, mongoDBName)

	// --- Usecases ---
	userUsecasePostgres := usecase.NewUserUsecase(userRepoPostgres, redisClient)
	userUsecaseMongo := usecase.NewUserUsecase(userRepoMongo, redisClient)

	// --- Handlers ---
	userHandlerPostgres := httpHandler.NewUserHandler(userUsecasePostgres)
	userHandlerMongo := httpHandler.NewUserHandler(userUsecaseMongo)

	// --- Router ---
	r := chi.NewRouter()

	// Setup routes dengan prefix
	r.Route("/pg", func(r chi.Router) {
		routes.SetupUserRoutes(r, userHandlerPostgres)
	})

	r.Route("/mongo", func(r chi.Router) {
		routes.SetupUserRoutes(r, userHandlerMongo)
	})

	// Optional: tambahkan health check ke root
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("🚀 User Service is running!"))
	})

	// Jalankan server
	fmt.Println("Server berjalan di port :9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}
