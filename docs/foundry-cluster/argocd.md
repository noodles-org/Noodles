# ArgoCD

ArgoCD provides GitOps-based continuous deployment for the cluster. It watches the GitHub repository and syncs Kubernetes manifests automatically.

## Access

- **URL:** `https://argocd.noodles.quest`
- **Authentication:** Dex OIDC SSO (ArgoCD is registered as the `argocd` Dex static client). Dex in turn authenticates users against the `noodles-org` GitHub organization.

## Projects

ArgoCD is configured with two projects:

### foundry-project
Manages the core Foundry infrastructure. Allowed destination namespaces: `foundry`, `traefik`, `monitoring`, `jellyfin`, `pihole`, `stalwart` (all targeting `in-cluster`).

The `foundry-apps` ApplicationSet generates applications from a list of directory paths with recursive sync:

| Application    | Source Path                                    | Namespace    |
|----------------|------------------------------------------------|--------------|
| `foundry`      | `foundry_deployment/infra/k8s/foundry`         | `foundry`    |
| `traefik`      | `foundry_deployment/infra/k8s/traefik`         | `traefik`    |
| `monitoring`   | `foundry_deployment/infra/k8s/monitoring`       | `monitoring` |
| `jellyfin`     | `foundry_deployment/infra/k8s/jellyfin`         | `jellyfin`   |
| `pihole`       | `foundry_deployment/infra/k8s/pihole`           | `pihole`     |
| `stalwart`     | `foundry_deployment/infra/k8s/stalwart`         | `stalwart`   |

Cluster-scoped resources are denied by default; `Namespace` and `PersistentVolume` are whitelisted (the latter is required by the `jellyfin-media-pv` media volume). `ResourceQuota`, `LimitRange`, and `NetworkPolicy` are blacklisted at the namespace scope.

### games-project
Manages game server deployments on a separate cluster context (`games-context`). Allowed destination namespaces: `enshrouded`, `satisfactory`, `soba`, `valheim`.

The `games-apps` ApplicationSet generates applications from a list of directory paths with recursive sync:

| Application    | Source Path                                    | Namespace      |
|----------------|------------------------------------------------|----------------|
| `soba`         | `gameserver_deployment/infra/soba`             | `soba`         |
| `enshrouded`   | `gameserver_deployment/infra/enshrouded`       | `enshrouded`   |
| `satisfactory` | `gameserver_deployment/infra/satisfactory`     | `satisfactory` |
| `valheim`      | `gameserver_deployment/infra/valheim`          | `valheim`      |

Both ApplicationSets pull from `https://github.com/noodles-org/Noodles` at `HEAD`.

## Unmanaged manifests

Because each Application syncs its source path recursively, any manifest placed under those paths is managed by ArgoCD. Manifests that must only ever be triggered manually are therefore kept outside them:

| Path                                          | Contents                                        |
|-----------------------------------------------|-------------------------------------------------|
| `foundry_deployment/infra/k8s/_manual-jobs/`   | One-off Jobs (e.g. Foundry manual backup/restore) |

These are applied with `kubectl apply -f` by hand. This also avoids sync failures caused by the immutable `spec.template` of a `Job` that already exists in the cluster.

## RBAC

ArgoCD roles are mapped to GitHub organization teams:

| GitHub Team               | ArgoCD Role | Permissions                                    |
|---------------------------|-------------|------------------------------------------------|
| `noodles-org:admin`       | `admin`     | Full access to applications, clusters, repos, logs, exec |
| `noodles-org:developer`   | `developer` | Read-only access to applications, projects, and logs     |

Project-level RBAC further restricts the `developer` role to read-only on each project's applications.

## Helm Chart

ArgoCD is deployed via a Helm wrapper chart:
- **Upstream chart:** `argo-cd` v8.2.5
- **Values:** `foundry_deployment/infra/helm/argocd/values.yaml`
- Prometheus annotations are enabled for scraping.
- The built-in ingress is disabled in favor of Traefik IngressRoutes.
