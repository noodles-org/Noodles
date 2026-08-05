# Auth

The games cluster uses the same Dex OIDC provider running on the foundry cluster for Kubernetes authentication. No separate Dex instance is deployed — the K3s API server is configured to validate tokens issued by `dex.noodles.quest`.

## Dex

- **Issuer:** `https://dex.noodles.quest` (hosted on the foundry cluster)
- **Client:** `kubelogin` (shared static client — same as foundry)
- **Connector:** GitHub OAuth via the `noodles-org` organization

The Dex instance and its configuration live in the foundry cluster. See the [foundry Auth docs](../foundry-cluster/auth.md#dex) for details on the Dex deployment.

## Kubernetes RBAC

RBAC resources are defined in `gameserver_deployment/infra/k8s/auth/rbac.yaml`:

| Resource             | Name                    | Scope                                              | Permissions    |
|----------------------|-------------------------|-----------------------------------------------------|----------------|
| `ClusterRoleBinding` | `dex-cluster-admin`     | Cluster-wide                                        | `cluster-admin` |
| `RoleBinding`        | `dex-games-developer`   | `satisfactory`, `enshrouded`, `valheim`, `soba`     | `admin`         |

Role mapping:

| GitHub Team               | Kubernetes Role | Scope                                          |
|---------------------------|-----------------|-------------------------------------------------|
| `noodles-org:admin`       | `cluster-admin` | Full cluster access                             |
| `noodles-org:developer`   | `admin`         | Namespaced access to game server namespaces     |

## K3s Configuration

The K3s API server is configured to validate Dex-issued JWTs via an `AuthenticationConfiguration` resource (`config/dex.yaml`). This file is a copy of the foundry cluster's version — keep both in sync.
