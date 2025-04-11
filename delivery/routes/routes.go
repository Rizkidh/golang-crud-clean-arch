package routes

import (
	"net/http"

	httpHandler "golang-crud-clean-arch/delivery/http"

	"github.com/go-chi/chi/v5"
)

// SetupRoutes initializes all application routes
func SetupRoutes(
	r *chi.Mux,
	repoHandler *httpHandler.RepositoryHandler,
	userHandler *httpHandler.UserHandler,
	healthHandler *httpHandler.HealthHandler,
) {
	SetupRepositoryRoutes(r, repoHandler)
	SetupUserRoutes(r, userHandler)
	SetupHealthRoutes(r, healthHandler)
}

// SetupRepositoryRoutes configures repository-related routes
func SetupRepositoryRoutes(r chi.Router, h *httpHandler.RepositoryHandler) {
	r.Route("/repositories", func(r chi.Router) {
		r.Post("/", http.HandlerFunc(h.Create))
		r.Get("/", http.HandlerFunc(h.GetAll))
		r.Get("/{id}", http.HandlerFunc(h.Get))
		r.Put("/{id}", http.HandlerFunc(h.Update))
		r.Delete("/{id}", http.HandlerFunc(h.Delete))
	})
}

// SetupUserRoutes configures user-related routes
func SetupUserRoutes(r chi.Router, h *httpHandler.UserHandler) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/", http.HandlerFunc(h.CreateUser))
		r.Get("/", http.HandlerFunc(h.GetAllUsers))
		r.Get("/{id}", http.HandlerFunc(h.GetUser))
		r.Put("/{id}", http.HandlerFunc(h.UpdateUser))
		r.Delete("/{id}", http.HandlerFunc(h.DeleteUser))
	})
}

// SetupHealthRoutes configures health check routes
func SetupHealthRoutes(r chi.Router, h *httpHandler.HealthHandler) {
	r.Route("/health", func(r chi.Router) {
		r.Get("/liveness", http.HandlerFunc(h.Liveness))
		r.Get("/readiness", http.HandlerFunc(h.Readiness))
	})
}
