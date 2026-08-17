package errs

import "net/http"

// Client registry errors.
var (
	PendingFull       = &Error{Status: http.StatusServiceUnavailable, Message: "Access requests are temporarily closed"}
	NotPending        = &Error{Status: http.StatusNotFound, Message: "Email not found in pending list"}
	NotApproved       = &Error{Status: http.StatusNotFound, Message: "Email not found in approved list"}
	InvalidClientRole = &Error{Status: http.StatusBadRequest, Message: "Invalid client role"}
)
