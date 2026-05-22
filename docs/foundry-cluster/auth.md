# Auth

The `auth` namespace handles authentication, authorization, and TLS certificate management for the cluster. It contains Dex (an OIDC provider), Kubernetes RBAC bindings, cert-manager resources, and networking routes.

## Dex

Dex is an OpenID Connect (OIDC) identity provider that enables GitHub-based authentication for cluster access via `kubelogin`.

- **Helm chart:** `dex/dex` deployed in the `auth` namespace
- **Issuer URL:** `https://dex.noodles.quest`
- **Connector:** GitHub OAuth via the `noodles-org` organization
- **Static client:** `kubelogin` — used for `kubectl` OIDC login, redirects to `http://localhost:8000`
- **TLS:** Dex serves its own TLS on port 5554 using a `dex-tls` Secret mounted at `/etc/dex/tls`. External traffic is terminated by Traefik using the cluster's Let's Encrypt wildcard cert (`noodles-quest-prod-tls`).
- **Storage:** In-memory
- **Credentials:** GitHub OAuth client ID/secret loaded from the `dex-github-oauth` Secret via `envFrom`

The K3s API server is configured to accept Dex-issued JWTs via an `AuthenticationConfiguration` resource (`config/dex.yaml`). It maps the `email` claim to the username and the `groups` claim to Kubernetes groups.

## Kubernetes RBAC

RBAC bindings in `k8s/auth/rbac.yaml` map GitHub organization teams to Kubernetes roles:

| Binding                  | Kind               | Role            | Subjects                                          | Scope              |
|--------------------------|--------------------|-----------------|---------------------------------------------------|--------------------|
| `dex-cluster-admin`      | ClusterRoleBinding | `cluster-admin` | `noodles-org:admin`                               | Cluster-wide       |
| `dex-foundry-developer`  | RoleBinding        | `admin`         | `noodles-org:developer`, `noodles-org:admin`      | `foundry` namespace |

- Members of `noodles-org:admin` receive full cluster-admin privileges.
- Members of `noodles-org:developer` (and admins) receive admin-level access within the `foundry` namespace.

## TLS Certificates

TLS is managed by cert-manager with Let's Encrypt. The cert-manager configuration lives in `k8s/auth/cert-manager.yaml` since the `ClusterIssuer` is a cluster-scoped resource (not namespaced).

- **ClusterIssuer:** `letsencrypt-prod` (ACME with Cloudflare DNS-01 solver)
- **Certificate:** Wildcard cert for `*.noodles.quest` and `noodles.quest` (namespaced to `auth`)
- **Secret:** `noodles-quest-prod-tls` — referenced by all IngressRoutes across the cluster.
- **Credentials:** Cloudflare API key stored in the `cloudflare` Secret (`global-api-key`).

## Networking

Dex is exposed via a Traefik IngressRoute:

- **External URL:** `https://dex.noodles.quest`
- An `IngressRoute` matches `Host(dex.noodles.quest)` on the `websecure` entrypoint.
- Routes to the `dex` Service in the `auth` namespace on port 5554.
- TLS termination uses the `noodles-quest-prod-tls` Secret.
