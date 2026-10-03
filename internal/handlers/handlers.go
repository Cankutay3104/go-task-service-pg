// Role: HTTP transport layer and controller logic for the Task API.
// Connects with: internal/repository (TaskRepository interface) and internal/validator (DecodePayload).
// Responsibilities:
// - Parses incoming HTTP requests, validates inputs, and delegates to the repository.
// - Maps domain errors (ErrNotFound, ErrConflict) to canonical HTTP response status codes.

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"task-service-pg/internal/models"
	"task-service-pg/internal/repository"
	"task-service-pg/internal/validator"
)

type TaskHandler struct {
	repo repository.TaskRepository
}

// NewTaskHandler constructs a TaskHandler utilizing Constructor Dependency Injection.
func NewTaskHandler(repo repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// writeJSON standardizes JSON content negotiation and status code delivery.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	task, err := validator.DecodePayload[models.Task](w, r, 1024*1024)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := task.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.repo.Create(r.Context(), &task)
	if errors.Is(err, repository.ErrConflict) {
		http.Error(w, repository.ErrConflict.Error(), http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	task, err := h.repo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, repository.ErrNotFound.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	task, err := validator.DecodePayload[models.Task](w, r, 1024*1024)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// We explicitly bind the route parameter to task.ID; that's why any mismatched ID provided in the JSON body gets overridden safely.
	task.ID = id

	if err = task.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.repo.Update(r.Context(), &task)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, repository.ErrNotFound.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.repo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, repository.ErrNotFound.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
