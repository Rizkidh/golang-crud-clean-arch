package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"golang-crud-clean-arch/internal/entity"
	"golang-crud-clean-arch/internal/usecase"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RepositoryHandler struct {
	repoUsecase *usecase.RepositoryUsecase
	userUsecase *usecase.UserUsecase
}

func NewRepositoryHandler(repoU *usecase.RepositoryUsecase, userU *usecase.UserUsecase) *RepositoryHandler {
	return &RepositoryHandler{
		repoUsecase: repoU,
		userUsecase: userU,
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *RepositoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var repo entity.Repository
	if err := json.NewDecoder(r.Body).Decode(&repo); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	ctx := context.Background()
	if err := h.repoUsecase.CreateRepository(ctx, &repo); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(repo)
}

func (h *RepositoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	repos, err := h.repoUsecase.GetAllRepositories(ctx)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch repositories")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repos)
}

func (h *RepositoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid repository ID format")
		return
	}

	ctx := context.Background()
	repo, err := h.repoUsecase.GetRepository(ctx, id)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "Repository not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repo)
}

func (h *RepositoryHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = h.userUsecase.GetUser(ctx, id)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *RepositoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid repository ID format")
		return
	}

	var updatedRepo entity.Repository
	if err := json.NewDecoder(r.Body).Decode(&updatedRepo); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	updatedRepo.ID = id
	ctx := context.Background()
	if err := h.repoUsecase.UpdateRepository(ctx, &updatedRepo); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Repository updated successfully"})
}

func (h *RepositoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid repository ID format")
		return
	}

	ctx := context.Background()
	if err := h.repoUsecase.DeleteRepository(ctx, id); err != nil {
		writeErrorResponse(w, http.StatusNotFound, "Repository not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Repository deleted successfully"})
}
