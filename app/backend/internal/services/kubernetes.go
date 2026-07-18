package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mephalrith/noodles/backend/internal/config"
	"github.com/mephalrith/noodles/backend/internal/model"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	pauseAnnotation   = "noodles.dashboard/original-replicas"
	restartAnnotation = "kubectl.kubernetes.io/restartedAt"
	argoLabel         = "argocd.argoproj.io/instance"
	serviceLabel      = "noodles.dashboard/service"
	annPrefix         = "noodles.dashboard/"
	nsCacheTTL        = 60 * time.Second
)

type K8sService struct {
	cfg      *config.Config
	client   kubernetes.Interface
	isDev    bool
	mocksDir string

	nsCache     []string
	nsCacheTime time.Time
	nsMu        sync.Mutex
}

func NewK8sService(cfg *config.Config) *K8sService {
	svc := &K8sService{
		cfg:      cfg,
		isDev:    !cfg.IsProduction,
		mocksDir: filepath.Join("mocks"),
	}

	if !svc.isDev {
		restCfg, err := rest.InClusterConfig()
		if err != nil {
			Logger.Warn("K8s: failed to load in-cluster config", "error", err)
			return svc
		}
		clientset, err := kubernetes.NewForConfig(restCfg)
		if err != nil {
			Logger.Warn("K8s: failed to create client", "error", err)
			return svc
		}
		svc.client = clientset
	}

	return svc
}

func (s *K8sService) loadMock(file string, v any) error {
	data, err := os.ReadFile(filepath.Join(s.mocksDir, file))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (s *K8sService) DiscoverNamespaces(ctx context.Context) ([]string, error) {
	if s.isDev {
		return []string{"foundry", "jellyfin", "stalwart"}, nil
	}

	s.nsMu.Lock()
	defer s.nsMu.Unlock()

	if s.nsCache != nil && time.Since(s.nsCacheTime) < nsCacheTTL {
		return s.nsCache, nil
	}

	nsList, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=true", s.cfg.NamespaceLabel),
	})
	if err != nil {
		Logger.Error("Failed to discover namespaces", "error", err)
		if s.nsCache == nil {
			s.nsCache = []string{}
		}
		return s.nsCache, err
	}

	s.nsCache = make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		s.nsCache = append(s.nsCache, ns.Name)
	}
	s.nsCacheTime = time.Now()

	if len(s.nsCache) == 0 {
		Logger.Warn("No namespaces found with label", "label", s.cfg.NamespaceLabel)
	}

	return s.nsCache, nil
}

func (s *K8sService) IsManagedNamespace(ctx context.Context, namespace string) (bool, error) {
	namespaces, err := s.DiscoverNamespaces(ctx)
	if err != nil {
		return false, err
	}
	for _, ns := range namespaces {
		if ns == namespace {
			return true, nil
		}
	}
	return false, nil
}

func deriveHealth(dep *appsv1.Deployment) string {
	desired := int32(0)
	if dep.Spec.Replicas != nil {
		desired = *dep.Spec.Replicas
	}
	available := dep.Status.AvailableReplicas
	ready := dep.Status.ReadyReplicas
	if desired == 0 {
		return "Suspended"
	}
	if ready == desired && available == desired {
		return "Healthy"
	}
	if ready > 0 {
		return "Progressing"
	}
	return "Degraded"
}

func (s *K8sService) ListDeployments(ctx context.Context) ([]model.DeploymentInfo, error) {
	if s.isDev {
		var deps []model.DeploymentInfo
		if err := s.loadMock("deployments.json", &deps); err != nil {
			return nil, err
		}
		return deps, nil
	}

	namespaces, err := s.DiscoverNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	var all []model.DeploymentInfo
	for _, ns := range namespaces {
		depList, err := s.client.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			Logger.Error("Failed listing deployments", "namespace", ns, "error", err)
			continue
		}

		for i := range depList.Items {
			dep := &depList.Items[i]
			ann := dep.Annotations
			if ann == nil {
				ann = map[string]string{}
			}

			var savedReplicas *int32
			if v, ok := ann[pauseAnnotation]; ok {
				n := int32(0)
				fmt.Sscanf(v, "%d", &n)
				savedReplicas = &n
			}

			replicas := int32(0)
			if dep.Spec.Replicas != nil {
				replicas = *dep.Spec.Replicas
			}

			image := "unknown"
			if len(dep.Spec.Template.Spec.Containers) > 0 {
				image = dep.Spec.Template.Spec.Containers[0].Image
			}

			argoApp := ""
			if dep.Labels != nil {
				argoApp = dep.Labels[argoLabel]
			}

			lastRestarted := ""
			if dep.Spec.Template.Annotations != nil {
				lastRestarted = dep.Spec.Template.Annotations[restartAnnotation]
			}

			createdAt := ""
			if !dep.CreationTimestamp.IsZero() {
				createdAt = dep.CreationTimestamp.Format(time.RFC3339)
			}

			all = append(all, model.DeploymentInfo{
				Name:              dep.Name,
				Namespace:         dep.Namespace,
				Replicas:          replicas,
				ReadyReplicas:     dep.Status.ReadyReplicas,
				AvailableReplicas: dep.Status.AvailableReplicas,
				Image:             image,
				Paused:            replicas == 0 && savedReplicas != nil,
				OriginalReplicas:  savedReplicas,
				ArgoApp:           argoApp,
				HealthStatus:      deriveHealth(dep),
				SyncStatus:        "Unknown",
				LastRestartedAt:   lastRestarted,
				CreatedAt:         createdAt,
			})
		}
	}

	return all, nil
}

func (s *K8sService) RestartDeployment(ctx context.Context, namespace, name string) error {
	if s.isDev {
		return fmt.Errorf("K8s not available in dev mode")
	}

	patch := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"%s":"%s"}}}}}`,
		restartAnnotation, time.Now().Format(time.RFC3339))

	_, err := s.client.AppsV1().Deployments(namespace).Patch(
		ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return err
	}

	Logger.Info("Deployment restarted", "namespace", namespace, "name", name)
	return nil
}

func (s *K8sService) PauseDeployment(ctx context.Context, namespace, name string) error {
	if s.isDev {
		return fmt.Errorf("K8s not available in dev mode")
	}

	dep, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	current := int32(1)
	if dep.Spec.Replicas != nil {
		current = *dep.Spec.Replicas
	}
	if current == 0 {
		return fmt.Errorf("Already paused")
	}

	patch := fmt.Sprintf(`{"metadata":{"annotations":{"%s":"%d"}},"spec":{"replicas":0}}`,
		pauseAnnotation, current)

	_, err = s.client.AppsV1().Deployments(namespace).Patch(
		ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return err
	}

	Logger.Info("Deployment paused", "namespace", namespace, "name", name, "previousReplicas", current)
	return nil
}

func (s *K8sService) ResumeDeployment(ctx context.Context, namespace, name string) error {
	if s.isDev {
		return fmt.Errorf("K8s not available in dev mode")
	}

	dep, err := s.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	replicas := int32(0)
	if dep.Spec.Replicas != nil {
		replicas = *dep.Spec.Replicas
	}
	if replicas > 0 {
		return fmt.Errorf("Not currently paused")
	}

	target := int32(1)
	if ann := dep.Annotations; ann != nil {
		if v, ok := ann[pauseAnnotation]; ok {
			fmt.Sscanf(v, "%d", &target)
		}
	}

	patch := fmt.Sprintf(`{"spec":{"replicas":%d}}`, target)

	_, err = s.client.AppsV1().Deployments(namespace).Patch(
		ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return err
	}

	Logger.Info("Deployment resumed", "namespace", namespace, "name", name, "replicas", target)
	return nil
}

func (s *K8sService) DiscoverServices(ctx context.Context) ([]model.ServiceLink, error) {
	if s.isDev {
		var links []model.ServiceLink
		if err := s.loadMock("services.json", &links); err != nil {
			return nil, err
		}
		return links, nil
	}

	type ingressRouteList struct {
		Items []struct {
			Metadata struct {
				Annotations map[string]string `json:"annotations"`
			} `json:"metadata"`
			Spec struct {
				TLS    *struct{} `json:"tls"`
				Routes []struct {
					Match string `json:"match"`
				} `json:"routes"`
			} `json:"spec"`
		} `json:"items"`
	}

	data, err := s.client.Discovery().RESTClient().Get().
		AbsPath("/apis/traefik.io/v1alpha1/ingressroutes").
		Param("labelSelector", fmt.Sprintf("%s=true", serviceLabel)).
		DoRaw(ctx)
	if err != nil {
		Logger.Error("Failed to discover services from IngressRoutes", "error", err)
		return nil, err
	}

	var list ingressRouteList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	var services []model.ServiceLink
	for _, item := range list.Items {
		ann := item.Metadata.Annotations
		name := ann[annPrefix+"name"]
		if name == "" {
			continue
		}

		hasTLS := item.Spec.TLS != nil
		url := ann[annPrefix+"url"]
		if url == "" {
			url = deriveURLFromRoutes(item.Spec.Routes, hasTLS)
		}

		services = append(services, model.ServiceLink{
			Name:        name,
			URL:         url,
			Description: ann[annPrefix+"description"],
			Category:    orDefault(ann[annPrefix+"category"], "Other"),
		})
	}

	return services, nil
}

func deriveURLFromRoutes(routes []struct {
	Match string `json:"match"`
}, hasTLS bool) string {
	for _, route := range routes {
		match := route.Match
		if match == "" {
			continue
		}
		host := extractRegex(match, `Host\(`+"`"+`([^`+"`"+`]+)`+"`"+`\)`)
		if host == "" {
			continue
		}
		scheme := "http"
		if hasTLS {
			scheme = "https"
		}
		pathPrefix := extractRegex(match, `PathPrefix\(`+"`"+`([^`+"`"+`]+)`+"`"+`\)`)
		if pathPrefix != "" {
			return fmt.Sprintf("%s://%s%s", scheme, host, pathPrefix)
		}
		return fmt.Sprintf("%s://%s", scheme, host)
	}
	return ""
}

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
