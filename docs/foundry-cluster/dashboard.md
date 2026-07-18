# Dashboard

The Noodles Dashboard is a web application at `noodles.quest` that provides a central interface for monitoring and managing cluster services.

## Features

- **Service Directory** — Auto-discovered list of all labeled cluster services and their URLs
- **Deployment Management** — View, restart, pause, and resume deployments across managed namespaces
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

### Secrets

The deployment references a `noodles-dashboard` Secret with the following keys:

- `DASHBOARD_CLIENT_SECRET` — Dex OIDC client secret
- `JWT_SECRET` — Secret for signing session JWTs
- `ARGOCD_TOKEN` — ArgoCD API token for sync status

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

### Namespace Discovery

The dashboard tracks namespaces labeled with `noodles.dashboard/managed: "true"` for deployment listing.

## Application Documentation

For frontend and backend development documentation, see the [app docs](https://github.com/noodles-org/Noodles/tree/main/app/docs).
