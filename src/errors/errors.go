package errors

import (
	"fmt"
	"net/http"
)

type Error struct {
	Code           string
	Message        string
	HTTPStatusCode int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code, message string, httpStatusCode int) *Error {
	return &Error{
		Code:           code,
		Message:        message,
		HTTPStatusCode: httpStatusCode,
	}
}

var (
	ErrInvalidData     = NewError("INVALID_DATA", "Datos inválidos", http.StatusBadRequest)
	ErrUserNotFound    = NewError("USER_NOT_FOUND", "Usuario no encontrado", http.StatusNotFound)
	ErrCourseNotFound  = NewError("COURSE_NOT_FOUND", "Curso no encontrado", http.StatusNotFound)
	ErrInternalServer  = NewError("INTERNAL_SERVER_ERROR", "Error interno del servidor", http.StatusInternalServerError)
	ErrDuplicateEnroll = NewError("DUPLICATE_ENROLL", "El estudiante ya está inscrito en este curso", http.StatusConflict)
	ErrMissingUserId   = NewError("MISSING_USER_ID", "El ID de usuario es requerido", http.StatusBadRequest)
	ErrMissingCourseId = NewError("MISSING_COURSE_ID", "El ID del curso es requerido", http.StatusBadRequest)
	ErrNoResults       = NewError("NO_RESULTS", "No se encontraron resultados", http.StatusNotFound)
)
