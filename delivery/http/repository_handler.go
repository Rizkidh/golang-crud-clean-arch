package http

import (
	"encoding/json"
	"net/http"

	"golang-crud-clean-arch/internal/entity"
	"golang-crud-clean-arch/internal/usecase"

	"github.com/go-chi/chi/v5"
)

type RepositoryHandler struct {
	usecase *usecase.RepositoryUsecase
}

func NewRepositoryHandler(usecase *usecase.RepositoryUsecase) *RepositoryHandler {
	return &RepositoryHandler{usecase}
}

func (h *RepositoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var repo entity.Repository
	if err := json.NewDecoder(r.Body).Decode(&repo); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if err := h.usecase.CreateRepository(r.Context(), &repo); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(repo)
}

func (h *RepositoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	repo, err := h.usecase.GetRepository(r.Context(), id)
	if err != nil {
		http.Error(w, "Repository not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(repo)
}

func (h *RepositoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	repos, err := h.usecase.GetAllRepositories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(repos)
}

func (h *RepositoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var repo entity.Repository
	if err := json.NewDecoder(r.Body).Decode(&repo); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	repo.ID = id

	if err := h.usecase.UpdateRepository(r.Context(), &repo); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	json.NewEncoder(w).Encode(repo)
}

func (h *RepositoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.usecase.DeleteRepository(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
