package model

type Role string

const (
	RoleAdmin       Role = "admin"
	RoleViewer      Role = "viewer"
	RoleClientAdmin Role = "client_admin"
	RoleClient      Role = "client"
	RolePending     Role = "pending"
)

func (r Role) CanMutate() bool {
	return r == RoleAdmin || r == RoleClientAdmin
}

func (r Role) CanRead() bool {
	switch r {
	case RoleAdmin, RoleViewer, RoleClientAdmin, RoleClient:
		return true
	default:
		return false
	}
}

func (r Role) IsStaff() bool {
	return r == RoleAdmin || r == RoleViewer
}

type User struct {
	Sub    string   `json:"sub"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Role   Role     `json:"role"`
	Groups []string `json:"groups"`
}

var DevUser = User{
	Sub:    "dev",
	Email:  "dev@localhost",
	Name:   "Dev User",
	Groups: []string{"noodles-org:admin"},
	Role:   RoleAdmin,
}
