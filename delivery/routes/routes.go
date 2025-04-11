package routes

import (
	"net/http"

	httpHandler "golang-crud-clean-arch/delivery/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
		r.Method("POST", "/", otelhttp.NewHandler(http.HandlerFunc(h.Create), "CreateRepository"))
		r.Method("GET", "/", otelhttp.NewHandler(http.HandlerFunc(h.GetAll), "GetAllRepositories"))
		r.Method("GET", "/{id}", otelhttp.NewHandler(http.HandlerFunc(h.Get), "GetRepository"))
		r.Method("PUT", "/{id}", otelhttp.NewHandler(http.HandlerFunc(h.Update), "UpdateRepository"))
		r.Method("DELETE", "/{id}", otelhttp.NewHandler(http.HandlerFunc(h.Delete), "DeleteRepository"))
	})
}

// SetupUserRoutes configures user-related routes
func SetupUserRoutes(r chi.Router, h *httpHandler.UserHandler) {
	r.Route("/users", func(r chi.Router) {
		r.Method("POST", "/", otelhttp.NewHandler(http.HandlerFunc(h.CreateUser), "CreateUser"))
		r.Method("GET", "/", otelhttp.NewHandler(http.HandlerFunc(h.GetAllUsers), "GetAllUsers"))
		r.Method("GET", "/{id}", otelhttp.NewHandler(http.HandlerFunc(h.GetUser), "GetUser"))
		r.Method("PUT", "/{id}", otelhttp.NewHandler(http.HandlerFunc(h.UpdateUser), "UpdateUser"))
		r.Method("DELETE", "/{id}", otelhttp.NewHandler(http.HandlerFunc(h.DeleteUser), "DeleteUser"))
	})
}

// SetupHealthRoutes configures health check routes
func SetupHealthRoutes(r chi.Router, h *httpHandler.HealthHandler) {
	r.Route("/health", func(r chi.Router) {
		r.Method("GET", "/liveness", otelhttp.NewHandler(http.HandlerFunc(h.Liveness), "Liveness"))
		r.Method("GET", "/readiness", otelhttp.NewHandler(http.HandlerFunc(h.Readiness), "Readiness"))
	})
}
