package model

type ClientEntry struct {
	Email     string `yaml:"email" json:"email"`
	Name      string `yaml:"name" json:"name"`
	Sub       string `yaml:"sub" json:"sub"`
	Role      Role   `yaml:"role" json:"role"`
	CreatedAt string `yaml:"createdAt" json:"createdAt"`
}
