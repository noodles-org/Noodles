package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mephalrith/noodles/backend/internal/config"
)

type ArgoStatus struct {
	Health string
	Sync   string
}

type ArgoCDService struct {
	cfg    *config.Config
	client *http.Client
}

func NewArgoCDService(cfg *config.Config) *ArgoCDService {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.ArgoCD.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &ArgoCDService{
		cfg: cfg,
		client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
	}
}

func (a *ArgoCDService) GetHealthMap() map[string]ArgoStatus {
	result := make(map[string]ArgoStatus)

	if a.cfg.ArgoCD.Token == "" {
		return result
	}

	url := fmt.Sprintf("%s/api/v1/applications", a.cfg.ArgoCD.URL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Logger.Error("ArgoCD: failed to create request", "error", err)
		return result
	}
	req.Header.Set("Authorization", "Bearer "+a.cfg.ArgoCD.Token)

	resp, err := a.client.Do(req)
	if err != nil {
		Logger.Error("ArgoCD fetch failed", "error", err)
		return result
	}
	defer resp.Body.Close()

	var data struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
			Status struct {
				Health struct {
					Status string `json:"status"`
				} `json:"health"`
				Sync struct {
					Status string `json:"status"`
				} `json:"sync"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		Logger.Error("ArgoCD: failed to decode response", "error", err)
		return result
	}

	for _, app := range data.Items {
		health := app.Status.Health.Status
		if health == "" {
			health = "Unknown"
		}
		sync := app.Status.Sync.Status
		if sync == "" {
			sync = "Unknown"
		}
		result[app.Metadata.Name] = ArgoStatus{Health: health, Sync: sync}
	}

	return result
}
