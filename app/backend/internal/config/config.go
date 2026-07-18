package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	AuthorizeURL string
	TokenURL     string
	UserinfoURL  string
	CallbackURL  string
	Scopes       string
}

type JWTConfig struct {
	Secret     string
	ExpiresIn  string
	CookieName string
}

type AuthConfig struct {
	AdminGroups   []string
	AllowedGroups []string
}

type ArgoCDConfig struct {
	URL      string
	Token    string
	Insecure bool
}

type Config struct {
	Port         int
	MetricsPort  int
	NodeEnv      string
	IsProduction bool

	OAuth  OAuthConfig
	JWT    JWTConfig
	Auth   AuthConfig
	ArgoCD ArgoCDConfig

	NamespaceLabel string
	DocsPath       string
	FrontendPath   string
	CORSOrigin     string
}

func Load() *Config {
	env := optional("NODE_ENV", "development")
	if env != "production" {
		_ = godotenv.Load(filepath.Join("..", ".env"))
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatal("failed to generate JWT secret:", err)
		}
		jwtSecret = hex.EncodeToString(b)
		log.Println("JWT_SECRET not set — generated ephemeral secret (sessions won't survive restarts)")
	}

	issuer := "https://dex.noodles.quest"
	publicURL := "https://noodles.quest"

	return &Config{
		Port:         optionalInt("PORT", 3000),
		MetricsPort:  optionalInt("METRICS_PORT", 9090),
		NodeEnv:      env,
		IsProduction: env == "production",

		OAuth: OAuthConfig{
			ClientID:     "noodles-dashboard",
			ClientSecret: os.Getenv("DASHBOARD_CLIENT_SECRET"),
			AuthorizeURL: fmt.Sprintf("%s/auth", issuer),
			TokenURL:     fmt.Sprintf("%s/token", issuer),
			UserinfoURL:  fmt.Sprintf("%s/userinfo", issuer),
			CallbackURL:  fmt.Sprintf("%s/api/auth/callback", publicURL),
			Scopes:       optional("OAUTH_SCOPES", "openid profile email groups"),
		},

		JWT: JWTConfig{
			Secret:     jwtSecret,
			ExpiresIn:  optional("JWT_EXPIRES_IN", "8h"),
			CookieName: "dashboard_token",
		},

		Auth: AuthConfig{
			AdminGroups:   []string{"noodles-org:admin"},
			AllowedGroups: []string{"noodles-org:admin", "noodles-org:developer"},
		},

		ArgoCD: ArgoCDConfig{
			URL:      "http://argocd-server.argocd.svc.cluster.local",
			Token:    os.Getenv("ARGOCD_TOKEN"),
			Insecure: optional("ARGOCD_INSECURE", "true") == "true",
		},

		NamespaceLabel: optional("NAMESPACE_LABEL", "noodles.dashboard/managed"),
		DocsPath:       filepath.Clean(optional("DOCS_PATH", "../../docs")),
		FrontendPath:   filepath.Clean(optional("FRONTEND_PATH", "../../frontend/dist")),
		CORSOrigin:     optional("CORS_ORIGIN", "http://localhost:5173"),
	}
}

func optional(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func optionalInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Fatalf("Invalid integer for %s: %s", key, val)
	}
	return n
}
