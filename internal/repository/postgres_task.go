// Role: Concrete PostgreSQL implementation of the TaskRepository interface.
// Connects with: internal/database (receives *sql.DB), internal/models (Task entity), and repository.go.
// Responsibilities:
// - Executes parameterized SQL queries (CRUD) against PostgreSQL via pgx.
// - Maps SQL result rows and errors to domain entities and sentinel errors.

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"task-service-pg/internal/models"
)

type PostgresTaskRepository struct {
	db *sql.DB
}

// NewPostgresTaskRepository injects the database connection pool into the repository.
func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (r *PostgresTaskRepository) Create(ctx context.Context, task *models.Task) error {
	if task.CreatedAt.IsZero() {
		now := time.Now().UTC()
		task.CreatedAt = now
		task.UpdatedAt = now
	}

	query := `
	INSERT INTO tasks (id, title, description, status, metadata, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err := r.db.ExecContext(
		ctx,
		query,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.Metadata,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}

	return nil
}

func (r *PostgresTaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	query := `
		SELECT id, title, description, status, metadata, created_at, updated_at
		FROM tasks
		WHERE id = $1;`

	var task models.Task

	// We pass memory pointers to Scan in the exact column order of the SELECT query.
	// Therefore, each database column is copied directly into the struct's fields.
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Metadata,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get task by id: %w", err)
	}

	return &task, nil
}

func (r *PostgresTaskRepository) List(ctx context.Context) ([]*models.Task, error) {
	query := `
	SELECT id, title, description, status, metadata, created_at, updated_at 
	FROM tasks 
	ORDER BY created_at DESC;`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]*models.Task, 0)
	for rows.Next() {
		var task models.Task

		err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Metadata,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", err)
		}

		tasks = append(tasks, &task)
	}

	// Moreover, checking rows.Err() ensures that any network drops or streaming issues during row iteration are not silently missed.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return tasks, nil
}

func (r *PostgresTaskRepository) Update(ctx context.Context, task *models.Task) error {
	query := `
	UPDATE tasks 
	SET title = $1, description = $2, status = $3, metadata = $4, updated_at = $5 
	WHERE id = $6;`

	task.UpdatedAt = time.Now().UTC()
	result, err := r.db.ExecContext(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Status,
		task.Metadata,
		task.UpdatedAt,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresTaskRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1;`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
