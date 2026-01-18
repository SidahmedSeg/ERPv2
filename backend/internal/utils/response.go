package utils

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Meta contains pagination and additional metadata
type Meta struct {
	Page       int `json:"page,omitempty"`
	PageSize   int `json:"page_size,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
	TotalCount int `json:"total_count,omitempty"`
}

// JSON writes a JSON response with the given status code
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Success writes a successful JSON response
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta writes a successful JSON response with metadata
func SuccessWithMeta(w http.ResponseWriter, data interface{}, meta *Meta) {
	JSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Created writes a 201 Created response
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent writes a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes an error JSON response
func Error(w http.ResponseWriter, statusCode int, code, message string) {
	JSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorWithDetails writes an error JSON response with detailed field errors
func ErrorWithDetails(w http.ResponseWriter, statusCode int, code, message string, details map[string]string) {
	JSON(w, statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// BadRequest writes a 400 Bad Request response
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized writes a 401 Unauthorized response
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden writes a 403 Forbidden response
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound writes a 404 Not Found response
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict writes a 409 Conflict response
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message)
}

// UnprocessableEntity writes a 422 Unprocessable Entity response
func UnprocessableEntity(w http.ResponseWriter, message string, details map[string]string) {
	ErrorWithDetails(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message, details)
}

// InternalServerError writes a 500 Internal Server Error response
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// TooManyRequests writes a 429 Too Many Requests response
func TooManyRequests(w http.ResponseWriter, message string) {
	Error(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", message)
}

// ServiceUnavailable writes a 503 Service Unavailable response
func ServiceUnavailable(w http.ResponseWriter, message string) {
	Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}

// ParseJSONBody parses JSON request body into the provided struct
func ParseJSONBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// NewMeta creates a new Meta struct for pagination
func NewMeta(page, pageSize, totalCount int) *Meta {
	totalPages := (totalCount + pageSize - 1) / pageSize
	return &Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalCount: totalCount,
	}
}
