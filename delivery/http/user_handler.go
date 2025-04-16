package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang-crud-clean-arch/internal/entity"
	"golang-crud-clean-arch/internal/usecase"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type UserHandler struct {
	usecase *usecase.UserUsecase
}

func NewUserHandler(usecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{usecase}
}

// CreateUser creates a new user in the database
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	tr := otel.Tracer("user-handler")
	ctx, span := tr.Start(r.Context(), "CreateUserHandler")
	defer span.End()

	var user entity.User
	// Decode the incoming user data
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid JSON")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("user.email", user.Email))

	// Use case to create the user in the database
	if err := h.usecase.CreateUser(ctx, &user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "CreateUser failed")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Set the ID as a string to match the expected response format
	span.SetAttributes(attribute.String("user.id", fmt.Sprintf("%v", user.ID)))
	span.SetStatus(codes.Ok, "User created")

	w.WriteHeader(http.StatusCreated)
	// Return the created user
	json.NewEncoder(w).Encode(user)
}

// GetUser fetches a user by ID from the database
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	tr := otel.Tracer("user-handler")
	ctx, span := tr.Start(r.Context(), "GetUserHandler")
	defer span.End()

	id := chi.URLParam(r, "id")
	span.SetAttributes(attribute.String("user.id", id))

	// Use case to get the user from the database
	user, err := h.usecase.GetUser(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetUser failed")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetAttributes(attribute.String("user.email", user.Email))
	span.SetStatus(codes.Ok, "User fetched")

	// Return the user details
	json.NewEncoder(w).Encode(user)
}

// UpdateUser updates an existing user in the database
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	tr := otel.Tracer("user-handler")
	ctx, span := tr.Start(r.Context(), "UpdateUserHandler")
	defer span.End()

	id := chi.URLParam(r, "id")
	span.SetAttributes(attribute.String("user.id", id))

	var user entity.User
	// Decode the incoming user data for updating
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Invalid JSON")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Set the user ID and pass to use case for updating
	user.ID = id
	span.SetAttributes(attribute.String("user.email", user.Email))

	// Use case to update the user in the database
	if err := h.usecase.UpdateUser(ctx, &user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "UpdateUser failed")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	span.SetStatus(codes.Ok, "User updated")
	// Return the updated user
	json.NewEncoder(w).Encode(user)
}

// DeleteUser deletes a user by ID from the database
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	tr := otel.Tracer("user-handler")
	ctx, span := tr.Start(r.Context(), "DeleteUserHandler")
	defer span.End()

	id := chi.URLParam(r, "id")
	span.SetAttributes(attribute.String("user.id", id))

	// Use case to delete the user from the database
	if err := h.usecase.DeleteUser(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "DeleteUser failed")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	span.SetStatus(codes.Ok, "User deleted")
	w.WriteHeader(http.StatusNoContent)
}

// GetAllUsers fetches all users from the database
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	tr := otel.Tracer("user-handler")
	ctx, span := tr.Start(r.Context(), "GetAllUsersHandler")
	defer span.End()

	// Use case to fetch all users from the database
	users, err := h.usecase.GetAllUsers(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetAllUsers failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("user.count", len(users)))
	span.SetStatus(codes.Ok, "Users fetched")

	// Return the list of users
	json.NewEncoder(w).Encode(users)
}

// TestCircuitBreaker tests the circuit breaker functionality
func (h *UserHandler) TestCircuitBreaker(w http.ResponseWriter, r *http.Request) {
	tr := otel.Tracer("user-handler")
	ctx, span := tr.Start(r.Context(), "TestCircuitBreaker")
	defer span.End()

	// Simulating multiple attempts to fetch all users to test the circuit breaker
	for i := 1; i <= 5; i++ {
		_, err := h.usecase.GetAllUsers(ctx)
		if err != nil {
			log.Printf("❌ Attempt %d failed: %v\n", i, err)
		} else {
			log.Printf("✅ Attempt %d succeeded\n", i)
		}
		time.Sleep(500 * time.Millisecond)
	}

	span.SetStatus(codes.Ok, "Test completed")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Test circuit breaker complete. Check logs."))
}
