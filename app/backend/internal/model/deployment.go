package model

type DeploymentInfo struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	Replicas          int32  `json:"replicas"`
	ReadyReplicas     int32  `json:"readyReplicas"`
	AvailableReplicas int32  `json:"availableReplicas"`
	Image             string `json:"image"`
	Paused            bool   `json:"paused"`
	OriginalReplicas  *int32 `json:"originalReplicas,omitempty"`
	ArgoApp           string `json:"argoApp,omitempty"`
	HealthStatus      string `json:"healthStatus"`
	SyncStatus        string `json:"syncStatus"`
	LastRestartedAt   string `json:"lastRestartedAt,omitempty"`
	CreatedAt         string `json:"createdAt,omitempty"`
}
