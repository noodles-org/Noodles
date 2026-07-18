package errs

import "net/http"

// Auth-related errors.
var (
	Unauthorized  = &Error{Status: http.StatusUnauthorized, Message: "Authentication required"}
	InvalidToken  = &Error{Status: http.StatusUnauthorized, Message: "Invalid or expired token"}
	InvalidClaims = &Error{Status: http.StatusUnauthorized, Message: "Invalid token claims"}
	Forbidden     = &Error{Status: http.StatusForbidden, Message: "Insufficient permissions"}
)
