# Dashboard

The Noodles Dashboard is a web application at `noodles.quest` that provides a central interface for monitoring and managing cluster services.

## Features

- **Service Directory** — Auto-discovered list of all labeled cluster services and their URLs
- **Deployment Management** — View, restart, pause, and resume deployments across managed namespaces on both the foundry and gameserver clusters
- **Documentation** — Integrated docs reader serving content from the `docs/` directory
- **Authentication** — Dex OIDC integration with role-based access control (admin/viewer)

## Deployment

The dashboard runs in the `dashboard` namespace with 1 replica. It is managed by ArgoCD via the `foundry-apps` ApplicationSet.

### Kubernetes Resources

| Resource | File | Purpose |
|----------|------|---------|
| Deployment | `deployment.yaml` | Go app with health/readiness probes and resource limits |
| Service | `service.yaml` | ClusterIP service on port 3000 with a metrics port on 9090 |
| IngressRoute | `routes.yaml` | Traefik route for `noodles.quest` with TLS and security headers |
| RBAC | `rbac.yaml` | ServiceAccount, ClusterRoles, and bindings for namespace/deployment/IngressRoute access |
| ConfigMap | `remote-clusters.yaml` | Remote cluster definitions injected as `REMOTE_CLUSTERS` env var (name, API URL, token env var, inline CA cert) |

### Secrets

The deployment references a `noodles-dashboard` Secret with the following keys:

- `DASHBOARD_CLIENT_SECRET` — Dex OIDC client secret
- `JWT_SECRET` — Secret for signing session JWTs
- `ARGOCD_TOKEN` — ArgoCD API token for sync status
- `GAMESERVER_TOKEN` — Long-lived service account token from the gameserver cluster's `argocd-manager` ServiceAccount, used to query gameserver deployments

### Service Discovery

Services are automatically discovered by querying IngressRoutes with the label `noodles.dashboard/service: "true"`. Metadata is read from annotations:

```yaml
labels:
  noodles.dashboard/service: "true"
annotations:
  noodles.dashboard/name: "Service Name"
  noodles.dashboard/description: "What it does"
  noodles.dashboard/category: "Applications"
  noodles.dashboard/url: "https://..."  # optional, derived from route rules if omitted
```

### Multi-Cluster Support

The dashboard monitors deployments across multiple clusters. The local (foundry) cluster is accessed via the in-cluster service account. Remote clusters are configured via the `REMOTE_CLUSTERS` environment variable (a JSON array), which is injected from the `remote-clusters` ConfigMap using `envFrom`.

Each entry specifies:

- `name` — Cluster display name (e.g., `gameserver`)
- `apiURL` — Kubernetes API server URL
- `tokenEnv` — Name of the env var containing the bearer token (sourced from the `noodles-dashboard` secret)
- `caPem` — The cluster's CA certificate PEM inline (public, not sensitive)

Additional clusters can be added by appending entries to the ConfigMap JSON array and adding the corresponding token to the secret.

### Namespace Discovery

The dashboard tracks namespaces labeled with `noodles.dashboard/managed: "true"` for deployment listing. This label must be applied on each cluster whose namespaces should appear in the dashboard.

## Application Documentation

For development documentation, see the [Backend](../app/backend.md) and [Frontend](../app/frontend.md) docs.
