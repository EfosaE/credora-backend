package response

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
)

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
	StatusCode int               `json:"-"`
	Success    bool              `json:"success"`
	Headers    map[string]string `json:"-"` // Custom headers to set
	Error      any               `json:"error"`
	Message    string            `json:"message,omitempty"`
}

// Render sets the proper status code and headers before rendering
func (e *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.StatusCode)

	// Set custom headers if any
	for key, value := range e.Headers {
		w.Header().Set(key, value)
	}

	return nil
}

// New creates a new ErrorResponse
func New(statusCode int, data any, message string) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: statusCode,
		Success:    false,
		Error:      data,
		Message:    message,
		Headers:    make(map[string]string),
	}
}

// WithHeader adds a custom header to the error response
func (e *ErrorResponse) WithHeader(key, value string) *ErrorResponse {
	if e.Headers == nil {
		e.Headers = make(map[string]string)
	}
	e.Headers[key] = value
	return e
}

// WithHeaders adds multiple custom headers to the error response
func (e *ErrorResponse) WithHeaders(headers map[string]string) *ErrorResponse {
	if e.Headers == nil {
		e.Headers = make(map[string]string)
	}
	for key, value := range headers {
		e.Headers[key] = value
	}
	return e
}

// BadRequest returns a 400 Bad Request error
func BadRequest(data any, message string) *ErrorResponse {
	var errData any
	if e, ok := data.(error); ok {
		errData = e.Error()
	} else {
		errData = data
	}

	return New(http.StatusBadRequest, errData, message)
}

// Conflict returns a 409 Conflict error
func Conflict(data any, message string) *ErrorResponse {
	var errData any
	if e, ok := data.(error); ok {
		errData = e.Error()
	} else {
		errData = data
	}
	return New(http.StatusConflict, errData, message)
}

// NotFound returns a 404 Not Found error
func NotFound(message string) *ErrorResponse {
	return New(http.StatusNotFound, "Resource not found", message)
}

// InternalServerError returns a 500 Internal Server Error
func InternalServerError(err error, message string) *ErrorResponse {
	return New(http.StatusInternalServerError, err.Error(), message)
}

func BuildServerError(statusCode int, err error, message string) *ErrorResponse {
	return New(statusCode, err.Error(), message)
}

// ServiceUnavailable returns a 503 Service Unavailable error with Retry-After header
func ServiceUnavailable(err error, message string, retryAfterSeconds int) *ErrorResponse {
	return New(http.StatusServiceUnavailable, err.Error(), message).
		WithHeader("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))
}

// Unauthorized returns a 401 Unauthorized error
func Unauthorized(message string) *ErrorResponse {
	return New(http.StatusUnauthorized, "Unauthorized", message)
}

// Forbidden returns a 403 Forbidden error
func Forbidden(message string) *ErrorResponse {
	return New(http.StatusForbidden, "Forbidden", message)
}

// ValidationError returns a 422 Unprocessable Entity error
func ValidationError(message string) *ErrorResponse {
	return New(http.StatusUnprocessableEntity, "This request cannot be processed", message)
}

// SendError is a convenience function to send an error response
func SendError(w http.ResponseWriter, r *http.Request, err *ErrorResponse) {
	render.Render(w, r, err)
}

// NotFoundHandler returns a custom 404 handler that responds with JSON
func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errorResponse := &ErrorResponse{
			StatusCode: http.StatusNotFound,
			Success:    false,
			Error:      "Not Found",
			Message:    fmt.Sprintf("A %s request doesn't exist on URL: '%s' on this server", r.Method, r.URL.Path),
			Headers:    make(map[string]string),
		}
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, errorResponse)
	}
}

// NotAllowedHandler returns a custom 405 handler that responds with JSON
func MethodNotAllowedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errorResponse := &ErrorResponse{
			StatusCode: http.StatusMethodNotAllowed,
			Success:    false,
			Error:      "Method Not Allowed",
			Message:    fmt.Sprintf("A %s request is not allowed on URL: '%s' on this server", r.Method, r.URL.Path),
			Headers:    make(map[string]string),
		}
		render.Status(r, http.StatusMethodNotAllowed)
		render.JSON(w, r, errorResponse)
	}
}
