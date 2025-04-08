package routes

import (
	"golang-crud-clean-arch/delivery/http"

	"github.com/go-chi/chi/v5"
)

// Setup Repository Routes
func SetupRepositoryRoutes(r *chi.Mux, repoHandler *http.RepositoryHandler) {
	r.Route("/repositories", func(r chi.Router) {
		r.Post("/", repoHandler.Create)
		r.Get("/", repoHandler.GetAll)
		r.Get("/{id}", repoHandler.GetByID)
		r.Put("/{id}", repoHandler.Update)
		r.Delete("/{id}", repoHandler.Delete)
	})
}

// Setup User Routes
func SetupUserRoutes(r *chi.Mux, userHandler *http.UserHandler) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/", userHandler.Create)
		r.Get("/", userHandler.GetAll) // Get all users
		r.Get("/{id}", userHandler.Get)
		r.Put("/{id}", userHandler.Update)
		r.Delete("/{id}", userHandler.Delete)
	})
}

// SetupHealthRoutes sets up the health check routes
// for liveness and readiness probes.
func SetupHealthRoutes(r chi.Router, handler *http.HealthHandler) {
	r.Get("/health/liveness", handler.Liveness)
	r.Get("/health/readiness", handler.Readiness)
}
