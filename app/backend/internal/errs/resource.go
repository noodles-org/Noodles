package errs

import "net/http"

// Resource and request validation errors.
var (
	NotFound            = &Error{Status: http.StatusNotFound, Message: "Not found"}
	BadRequest          = &Error{Status: http.StatusBadRequest, Message: "Bad request"}
	InvalidTarget       = &Error{Status: http.StatusBadRequest, Message: "Invalid namespace or deployment name"}
	NamespaceNotManaged = &Error{Status: http.StatusForbidden, Message: "Namespace not managed by this dashboard"}
	MissingPath         = &Error{Status: http.StatusBadRequest, Message: "Missing path parameter"}
	InvalidPath         = &Error{Status: http.StatusBadRequest, Message: "Invalid path"}
)
