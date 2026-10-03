// Role: Core business entity definition for the application.
// Connects with: internal/repository (used as input/output) and internal/handlers (payload targets).
// Responsibilities:
// - Defines the Task struct, validation logic, and TaskStatus enum.
// - Implements sql.Scanner and driver.Valuer to serialize/deserialize TaskMetadata to PostgreSQL JSONB.

package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// TaskStatus defines the constrained enumeration for task lifecycles.
type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

// TaskMetadata is stored inside PostgreSQL as binary JSONB.
type TaskMetadata struct {
	Tags     []string `json:"tags"`
	Priority int      `json:"priority"`
}

// Value serializes TaskMetadata to JSON bytes for PostgreSQL JSONB persistence.
// Therefore, TaskMetadata satisfies the database/sql/driver.Valuer interface.
func (m TaskMetadata) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan unmarshals PostgreSQL JSONB data directly into the TaskMetadata struct.
// For that purpose, we must use a pointer receiver to mutate the fields in place.
func (m *TaskMetadata) Scan(src any) error {
	if src == nil {
		*m = TaskMetadata{}
		return nil
	}

	var sourceBytes []byte
	switch v := src.(type) {
	case []byte:
		sourceBytes = v
	case string:
		sourceBytes = []byte(v)
	default:
		return fmt.Errorf("unsupported type for TaskMetadata: %T", src)
	}

	return json.Unmarshal(sourceBytes, m)
}

// Task represents the primary domain model stored in PostgreSQL.
type Task struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Metadata    TaskMetadata `json:"metadata"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Validate checks domain invariants before persistence or updates.
func (t *Task) Validate() error {
	if t.Title == "" {
		return errors.New("title is required and cannot be empty")
	}
	switch t.Status {
	case StatusTodo, StatusInProgress, StatusDone:
		return nil
	default:
		return fmt.Errorf("invalid status: %s", t.Status)
	}
}
