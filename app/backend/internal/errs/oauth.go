package errs

import "net/http"

// OAuth flow errors.
var (
	TokenRequestFailed    = &Error{Status: http.StatusBadGateway, Message: "Token request failed"}
	TokenDecodeFailed     = &Error{Status: http.StatusBadGateway, Message: "Failed to decode token response"}
	UserinfoRequestFailed = &Error{Status: http.StatusInternalServerError, Message: "Failed to create userinfo request"}
	UserinfoFetchFailed   = &Error{Status: http.StatusBadGateway, Message: "Userinfo fetch failed"}
	UserinfoReadFailed    = &Error{Status: http.StatusBadGateway, Message: "Failed to read userinfo response"}
	UserinfoDecodeFailed  = &Error{Status: http.StatusBadGateway, Message: "Failed to parse userinfo response"}
)
