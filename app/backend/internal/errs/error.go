package errs

import (
	"fmt"
	"net/http"
)

// Error represents a structured error with an HTTP status code.
type Error struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
	Cause   error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%d: %s: %v", e.Status, e.Message, e.Cause)
	}
	return fmt.Sprintf("%d: %s", e.Status, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// Wrap returns a copy of the error with the given cause attached.
func (e *Error) Wrap(cause error) *Error {
	return &Error{Status: e.Status, Message: e.Message, Cause: cause}
}

// Internal creates a new 500 error with a custom message.
func Internal(msg string) *Error {
	return &Error{Status: http.StatusInternalServerError, Message: msg}
}
