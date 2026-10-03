// Role: Integration test suite for the PostgreSQL repository implementation.
// Connects with: internal/database (opens test pool), internal/models (Task entity), and postgres_task.go.
// Responsibilities:
// - Verifies CRUD SQL queries, JSONB conversions, and error mappings against live PostgreSQL.
// - Uses a transaction rollback pattern or immediate cleanup to leave zero dirty test rows.

package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"task-service-pg/internal/database"
	"task-service-pg/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// setupTestDB connects to PostgreSQL, runs migrations, and handles teardown.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Load local .env if present
	_ = godotenv.Load("../../.env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is not set for integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.NewPostgresDB(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := database.Migrate(ctx, db); err != nil {
		_ = db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestPostgresTaskRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresTaskRepository(db)
	ctx := context.Background()

	task := &models.Task{
		ID:          "test-task-1",
		Title:       "Integration Test Task",
		Description: "Validating PostgreSQL integration",
		Status:      models.StatusTodo,
		Metadata: models.TaskMetadata{
			Tags:     []string{"backend", "postgres"},
			Priority: 1,
		},
	}

	// Always cleanup the inserted task even if an assertion fails midway
	t.Cleanup(func() {
		_ = repo.Delete(context.Background(), task.ID)
	})

	// 1. Test Create
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// 2. Test GetByID and verify JSONB deserialization
	fetched, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("failed to fetch task: %v", err)
	}
	if fetched.Title != task.Title {
		t.Errorf("got title %q, want %q", fetched.Title, task.Title)
	}
	if fetched.Metadata.Priority != 1 {
		t.Errorf("got priority %d, want 1", fetched.Metadata.Priority)
	}
	if len(fetched.Metadata.Tags) != 2 {
		t.Errorf("got %d tags, want 2", len(fetched.Metadata.Tags))
	}

	// 3. Test Update
	task.Status = models.StatusDone
	task.Title = "Updated Title"
	if err := repo.Update(ctx, task); err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	check, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated task: %v", err)
	}
	if check.Status != models.StatusDone || check.Title != "Updated Title" {
		t.Errorf("update mismatch: status = %q, title = %q", check.Status, check.Title)
	}

	// 4. Test Delete
	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	// 5. Verify the record is gone and returns ErrNotFound
	_, err = repo.GetByID(ctx, task.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after deletion, got: %v", err)
	}
}
