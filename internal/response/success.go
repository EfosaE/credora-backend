package response


import (
	"net/http"

	"github.com/go-chi/render"
)

// SuccessResponse represents a standard success response structure
type SuccessResponse struct {
	StatusCode int    `json:"-"`              // Not rendered in the JSON body
	Data       any    `json:"data,omitempty"` // Optional payload
	Message    string `json:"message,omitempty"`
}

// Render sets the HTTP status code before rendering the response
func (s *SuccessResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, s.StatusCode)
	return nil
}

// NewSuccess creates a new SuccessResponse with data and optional message
func NewSuccess(statusCode int, data any, message string) *SuccessResponse {
	return &SuccessResponse{
		StatusCode: statusCode,
		Data:       data,
		Message:    message,
	}
}

// OK returns a 200 OK success response
func OK(data any, message string) *SuccessResponse {
	return NewSuccess(http.StatusOK, data, message)
}

// Created returns a 201 Created success response
func Created(data any, message string) *SuccessResponse {
	return NewSuccess(http.StatusCreated, data, message)
}

// NoContent returns a 204 No Content success response (with no data or message)
func NoContent() *SuccessResponse {
	return &SuccessResponse{
		StatusCode: http.StatusNoContent,
	}
}

// SendSuccess is a convenience function to send a success response
func SendSuccess(w http.ResponseWriter, r *http.Request, res *SuccessResponse) {
	render.Render(w, r, res)
}