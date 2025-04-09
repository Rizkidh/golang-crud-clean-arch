package routes

import (
	"golang-crud-clean-arch/delivery/http"

	"github.com/go-chi/chi/v5"
)

// SetupRoutes initializes all application routes
func SetupRoutes(
	r *chi.Mux,
	repoHandler *http.RepositoryHandler,
	userHandler *http.UserHandler,
	healthHandler *http.HealthHandler,
) {
}

// SetupRepositoryRoutes configures repository-related routes
func SetupRepositoryRoutes(r chi.Router, repoHandler *http.RepositoryHandler) {
	r.Route("/repositories", func(r chi.Router) {
		r.Post("/", repoHandler.Create)
		r.Get("/", repoHandler.GetAll)
		r.Get("/{id}", repoHandler.GetByID)
		r.Put("/{id}", repoHandler.Update)
		r.Delete("/{id}", repoHandler.Delete)

	})
}

// SetupUserRoutes configures user-related routes
func SetupUserRoutes(r chi.Router, h *http.UserHandler) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.CreateUser)
		r.Get("/", h.GetAllUsers)
		r.Get("/{id}", h.GetUser)
		r.Put("/{id}", h.UpdateUser)
		r.Delete("/{id}", h.DeleteUser)
	})
}

// SetupHealthRoutes configures health check routes
func SetupHealthRoutes(r chi.Router, healthHandler *http.HealthHandler) {
	r.Route("/health", func(r chi.Router) {
		r.Get("/liveness", healthHandler.Liveness)
		r.Get("/readiness", healthHandler.Readiness)
	})
}
