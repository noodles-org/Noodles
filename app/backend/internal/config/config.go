package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

type FileSvcConfig struct {
	URL   string
	Token string
	Root  string
}

type RemoteClusterConfig struct {
	Name     string `json:"name"`
	APIURL   string `json:"apiURL"`
	TokenEnv string `json:"tokenEnv"`
	Token    string `json:"-"`
	CAPem    string `json:"caPem"`
}

type Config struct {
	Port         int
	MetricsPort  int
	NodeEnv      string
	IsProduction bool

	OAuth          OAuthConfig
	JWT            JWTConfig
	Auth           AuthConfig
	ArgoCD         ArgoCDConfig
	FileSvc        FileSvcConfig
	RemoteClusters []RemoteClusterConfig

	NamespaceLabel   string
	ClientsNamespace string
	DocsPath         string
	FrontendPath     string
	CORSOrigin       string
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

		FileSvc: FileSvcConfig{
			URL:   optional("FILESVC_URL", "http://file-sidecar.foundry.svc.cluster.local"),
			Token: optional("FILESVC_TOKEN", "dev-token"),
			Root:  optional("FILESVC_ROOT", filepath.Join("mocks", "files")),
		},

		RemoteClusters: loadRemoteClusters(),

		NamespaceLabel:   optional("NAMESPACE_LABEL", "noodles.dashboard/managed"),
		ClientsNamespace: optional("CLIENTS_NAMESPACE", "dashboard"),
		DocsPath:         filepath.Clean(optional("DOCS_PATH", "../../docs")),
		FrontendPath:     filepath.Clean(optional("FRONTEND_PATH", "../../frontend/dist")),
		CORSOrigin:       corsOrigin(env, publicURL),
	}
}

func loadRemoteClusters() []RemoteClusterConfig {
	data := os.Getenv("REMOTE_CLUSTERS")
	if data == "" {
		return nil
	}

	var clusters []RemoteClusterConfig
	if err := json.Unmarshal([]byte(data), &clusters); err != nil {
		log.Printf("Failed to parse REMOTE_CLUSTERS: %v", err)
		return nil
	}

	for i := range clusters {
		if clusters[i].TokenEnv != "" {
			clusters[i].Token = os.Getenv(clusters[i].TokenEnv)
		}
	}

	return clusters
}

func optional(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func corsOrigin(env, publicURL string) string {
	if env != "production" {
		return "http://localhost:5173"
	}
	return publicURL
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
