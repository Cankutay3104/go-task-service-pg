// Role: Request sanitization and JSON stream parsing helper.
// Connects with: internal/handlers (used inside POST and PUT handlers).
// Responsibilities:
// - Hardens incoming JSON parsing using http.MaxBytesReader to protect against large payloads.
// - Rejects unknown keys via DisallowUnknownFields and prevents trailing garbage data.

package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// DecodePayload unmarshals JSON while enforcing body limits, rejecting unknown fields, and blocking trailing streams.
func DecodePayload[T any](w http.ResponseWriter, r *http.Request, maxBytes int64) (T, error) {
	// Guard against Out-Of-Memory (OOM) attacks by bounding the stream size
	reader := http.MaxBytesReader(w, r.Body, maxBytes)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	var payload T
	if err := decoder.Decode(&payload); err != nil {
		return payload, fmt.Errorf("failed to decode json body: %w", err)
	}

	// If reading past the target struct does not return io.EOF, extraneous trailing tokens exist; therefore, we reject the request.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return payload, errors.New("body must contain only a single JSON object")
	}

	return payload, nil
}
