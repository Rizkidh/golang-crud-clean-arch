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
)

func main() {
	// Inisialisasi database
	mongoClient := config.MongoConnect()
	dbName := os.Getenv("MONGO_DB_NAME")

	// Inisialisasi Redis
	redisClient := config.ConnectRedis()

	// Inisialisasi repository
	repoRepo := repository.NewRepoRepository(mongoClient, redisClient, dbName)
	userRepo := repository.NewUserRepository(mongoClient, redisClient, dbName)

	// Inisialisasi usecase
	repoUsecase := usecase.NewRepositoryUsecase(repoRepo, redisClient)
	userUsecase := usecase.NewUserUsecase(userRepo, redisClient)

	// Inisialisasi handler
	repoHandler := httpHandler.NewRepositoryHandler(repoUsecase, userUsecase)
	userHandler := httpHandler.NewUserHandler(userUsecase)
	healthHandler := httpHandler.NewHealthHandler(mongoClient, redisClient) // <- Tambah ini

	// Setup router
	r := chi.NewRouter()
	routes.SetupRepositoryRoutes(r, repoHandler)
	routes.SetupUserRoutes(r, userHandler)
	routes.SetupHealthRoutes(r, healthHandler) // <- Tambah ini

	fmt.Println("Server berjalan di port :9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}
