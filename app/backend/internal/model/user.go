package model

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleViewer Role = "viewer"
)

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
