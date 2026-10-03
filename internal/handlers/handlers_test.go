// Role: HTTP transport layer unit test suite for the Task API.
// Connects with: internal/handlers (TaskHandler), internal/models (Task), and repository.go.
// Responsibilities:
// - Verifies HTTP status codes, routing values, and JSON payload handling using httptest.
// - Implements an in-memory mock repository to test handlers in isolation without PostgreSQL.

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"task-service-pg/internal/models"
	"task-service-pg/internal/repository"
)

// mockTaskRepository satisfies repository.TaskRepository using an in-memory map.
type mockTaskRepository struct {
	tasks map[string]*models.Task
}

func newMockTaskRepository() *mockTaskRepository {
	return &mockTaskRepository{
		tasks: make(map[string]*models.Task),
	}
}

func (m *mockTaskRepository) Create(ctx context.Context, task *models.Task) error {
	if _, exists := m.tasks[task.ID]; exists {
		return repository.ErrConflict
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) GetByID(ctx context.Context, id string) (*models.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return task, nil
}

func (m *mockTaskRepository) List(ctx context.Context) ([]*models.Task, error) {
	result := make([]*models.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		result = append(result, task)
	}
	return result, nil
}

func (m *mockTaskRepository) Update(ctx context.Context, task *models.Task) error {
	if _, exists := m.tasks[task.ID]; !exists {
		return repository.ErrNotFound
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) Delete(ctx context.Context, id string) error {
	if _, exists := m.tasks[id]; !exists {
		return repository.ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}

func TestTaskHandler_CreateTest(t *testing.T) {
	mockRepo := newMockTaskRepository()
	handler := NewTaskHandler(mockRepo)

	inputTask := models.Task{
		ID:          "task-123",
		Title:       "Test handler Task",
		Description: "Testing CreateTask handler",
		Status:      models.StatusTodo,
		Metadata: models.TaskMetadata{
			Tags:     []string{"unit-test"},
			Priority: 2,
		},
	}

	bodyBytes, err := json.Marshal(inputTask)
	if err != nil {
		t.Fatalf("failed to marshal input task: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.CreateTask(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var createdTask models.Task
	if err := json.NewDecoder(rec.Body).Decode(&createdTask); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if createdTask.Title != inputTask.Title {
		t.Errorf("expected title %q, got %q", inputTask.Title, createdTask.Title)
	}
	if createdTask.Status != inputTask.Status {
		t.Errorf("expected status %q, got %q", inputTask.Status, createdTask.Status)
	}
}

func TestTaskHandler_GetTask_NotFound(t *testing.T) {
	mockRepo := newMockTaskRepository()
	handler := NewTaskHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/tasks/non-existent-id", nil)
	req.SetPathValue("id", "non-existent-id")

	rec := httptest.NewRecorder()
	handler.GetTask(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
}
