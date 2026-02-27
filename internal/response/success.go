package response

import (
	"net/http"

	"github.com/go-chi/render"
)

func Obj(key string, value any) map[string]any {
	return map[string]any{key: value}
}

type KV struct {
	Key   string
	Value any
}

func ObjKV(pairs ...KV) map[string]any {
	m := make(map[string]any, len(pairs))
	for _, p := range pairs {
		m[p.Key] = p.Value
	}
	return m
}

// SuccessResponse represents the standard API success contract.
type SuccessResponse struct {
	StatusCode int            `json:"-"`
	Success    bool           `json:"success"`
	Message    string         `json:"message,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
	Meta       map[string]any `json:"meta,omitempty"`
}

// Render sets the HTTP status code before rendering.
func (s *SuccessResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, s.StatusCode)
	return nil
}

// NewSuccess creates a success response following the API contract.
func NewSuccess(
	statusCode int,
	data map[string]any,
	meta map[string]any,
	message string,
) *SuccessResponse {

	// Ensure contract consistency (never nil objects)
	if data == nil {
		data = map[string]any{}
	}

	if meta == nil {
		meta = map[string]any{}
	}

	return &SuccessResponse{
		StatusCode: statusCode,
		Success:    true,
		Message:    message,
		Data:       data,
		Meta:       meta,
	}
}

// OK returns 200 OK.
func OK(data map[string]any, meta map[string]any, message string) *SuccessResponse {
	return NewSuccess(http.StatusOK, data, meta, message)
}

// Created returns 201 Created.
func Created(data map[string]any, meta map[string]any, message string) *SuccessResponse {
	return NewSuccess(http.StatusCreated, data, meta, message)
}

// Accepted returns 202 Accepted.
func Accepted(data map[string]any, meta map[string]any, message string) *SuccessResponse {
	return NewSuccess(http.StatusAccepted, data, meta, message)
}

// NoContent returns 204 with empty contract-compliant body.
func NoContent() *SuccessResponse {
	return &SuccessResponse{
		StatusCode: http.StatusNoContent,
		Data:       map[string]any{},
		Meta:       map[string]any{},
	}
}

// SendSuccess sends the response.
func SendSuccess(w http.ResponseWriter, r *http.Request, res *SuccessResponse) {
	render.Render(w, r, res)
}
