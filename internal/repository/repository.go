// Role: Repository interface contract and domain sentinel error definitions.
// Connects with: internal/models (Task entity) and internal/handlers (via TaskRepository interface).
// Responsibilities:
// - Defines the abstract TaskRepository contract enabling test mocking.
// - Declares sentinel domain errors (ErrNotFound, ErrConflict) for consistent error handling.

package repository

import (
	"context"
	"errors"

	"task-service-pg/internal/models"
)

var (
	ErrNotFound = errors.New("task not found")
	ErrConflict = errors.New("task already exists")
)

// TaskRepository defines the abstract persistence boundary for Task entities.
type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	GetByID(ctx context.Context, id string) (*models.Task, error)
	List(ctx context.Context) ([]*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
}
