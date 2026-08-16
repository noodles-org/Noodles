package errs

// IdentityError is a login rejection carrying the metric reason and the query
// parameter shown on the login page.
type IdentityError struct {
	Reason   string
	Redirect string
	Message  string
}

func (e *IdentityError) Error() string { return e.Message }

var (
	NotAuthorized   = &IdentityError{Reason: "unauthorized_group", Redirect: "not_authorized", Message: "Auth: user not authorized"}
	EmailUnverified = &IdentityError{Reason: "email_unverified", Redirect: "email_unverified", Message: "Auth: email missing or unverified"}
	RequestsClosed  = &IdentityError{Reason: "requests_closed", Redirect: "requests_closed", Message: "Auth: pending list full, request refused"}
	RegistryFailed  = &IdentityError{Reason: "registry_error", Redirect: "auth_failed", Message: "Auth: failed to record access request"}
)
