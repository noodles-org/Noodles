# ArgoCD

ArgoCD provides GitOps-based continuous deployment for the cluster. It watches the GitHub repository and syncs Kubernetes manifests automatically.

## Access

- **URL:** `https://argocd.noodles.quest`
- **Authentication:** GitHub OAuth via the `noodles-org` GitHub organization.

## Projects

ArgoCD is configured with two projects:

### foundry-project
Manages the core Foundry infrastructure:
- **foundry** app — Syncs `foundry_deployment/infra/k8s/foundry` and `foundry_deployment/infra/k8s/foundry/jobs`.
- **traefik** app — Syncs `foundry_deployment/infra/k8s/traefik`.
- **monitoring** app — Syncs `foundry_deployment/infra/k8s/monitoring`.

Allowed destination namespaces: `foundry`, `traefik`, `monitoring`.

### games-project
Manages game server deployments (defined separately in `gameserver_deployment/`).

## RBAC

ArgoCD roles are mapped to GitHub organization teams:

| GitHub Team               | ArgoCD Role | Permissions                                    |
|---------------------------|-------------|------------------------------------------------|
| `noodles-org:admin`       | `admin`     | Full access to applications, clusters, repos, logs, exec |
| `noodles-org:developer`   | `developer` | Read-only access to applications, projects, and logs     |

Project-level RBAC further restricts the `developer` role to read-only on `foundry-project` applications.

## Helm Chart

ArgoCD is deployed via a Helm wrapper chart:
- **Upstream chart:** `argo-cd` v8.2.5
- **Values:** `foundry_deployment/infra/helm/argocd/values.yaml`
- Prometheus annotations are enabled for scraping.
- The built-in ingress is disabled in favor of Traefik IngressRoutes.
