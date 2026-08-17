# Client Access

Clients are non-staff users who sign in to the dashboard at `noodles.quest` with a Google account. They are never added to the GitHub organization, never granted ArgoCD or Kubernetes permissions, and only ever talk to the dashboard backend.

## Identity

Dex exposes two connectors and the dashboard continues to use the single `noodles-dashboard` static client:

- **GitHub** — staff, resolved from `noodles-org` team membership (unchanged).
- **Google** — clients, resolved from the client registry by verified email address.

Dex only proves identity. Google returns no useful `groups` claim, so authorization is owned entirely by the dashboard backend.

## Roles

| Role | Services | Docs | Deployments (read) | Restart/Pause/Resume | Clients page |
|------|----------|------|--------------------|----------------------|--------------|
| `admin` (staff) | ✅ | ✅ | ✅ | ✅ | ✅ |
| `viewer` (staff) | ✅ | ✅ | ✅ | ❌ | ❌ |
| `client_admin` | ✅ | ✅ | ✅ | ✅ | ❌ |
| `client` | ✅ | ✅ | ✅ | ❌ | ❌ |
| `pending` | ❌ | ❌ | ❌ | ❌ | ❌ |

Clients see all deployments — there is no per-client namespace scoping. How these roles are enforced is documented in [Backend](../app/backend.md#roles-and-route-gating) and [Frontend](../app/frontend.md#roles-and-guards).

## Sign-in Flow

The dashboard renders the provider choice itself: the login page offers **Sign in with GitHub** and **Sign in with Google**, and `/api/auth/login?connector=github|google` forwards the choice to Dex as `connector_id`. Dex therefore skips its own unthemed connector-selection screen and goes straight to the provider. An unknown or missing `connector` value is ignored and the Dex chooser is shown.

Staff always win: a GitHub group match resolves to `admin` or `viewer` before the client registry is consulted. Otherwise the verified Google email is looked up in the approved registry, then the pending list, and an unknown email is recorded as a new access request. The Google `email_verified` claim must be true — an unverified email is rejected and nothing is written.

### Waiting Room

An unknown user lands on a waiting-room page with no navigation and no data access until staff approve them. New pending entries are refused once the pending list holds **20** entries; the user sees a "requests temporarily closed" message. The list is also cleared automatically every week, so abandoned requests do not accumulate towards the cap.

## Registry Storage

Two ConfigMaps in the `dashboard` namespace, each holding a `clients.yaml` key with a YAML list of `{email, name, sub, role, createdAt}`:

| ConfigMap | Contents |
|-----------|----------|
| `noodles-clients` | Approved clients and their role |
| `noodles-clients-pending` | Outstanding access requests |

Both are **backend-owned at runtime**. Git holds only empty bootstrap shells (`clients.yaml: "[]"`) so the objects exist on a fresh cluster; the manifests must never be populated by hand, since entries are client PII and this repository is public. The dashboard's ArgoCD Application ignores the `data` field of both objects, and each carries `argocd.argoproj.io/compare-options: IgnoreExtraneous` with `Prune=false`, so a sync can never reset the allowlist.

They are not sops-encrypted: they contain no credentials (Google does the authenticating), and an encrypted ConfigMap could not be written by the backend, which would break both the waiting room and the approval page.

There is no git history of the allowlist. If the ConfigMap is lost, clients simply drop back to `pending` and are re-approved — nothing is compromised.

## Approval

Staff admins approve or reject requests from the **User Management** page in the dashboard, assigning `client` or `client_admin`. Approving moves the entry between the two ConfigMaps. An already approved client can be revoked from the same page, which removes them from `noodles-clients`; the action asks for confirmation first so it cannot be triggered by a stray click. Because the role is carried in the session JWT, a revoked client keeps their access until the 8h cookie expires.

The page and its admin-only API are documented in [Frontend](../app/frontend.md#clients) and [Backend](../app/backend.md#api-routes).

## Kubernetes Resources

| Resource | File | Purpose |
|----------|------|---------|
| ConfigMap | `clients.yaml` | Empty shell for the approved registry |
| ConfigMap | `clients-pending.yaml` | Empty shell for pending requests |
| Role / RoleBinding | `rbac.yaml` | `get/update/patch` for the dashboard ServiceAccount on those two ConfigMaps only |
| CronJob | `jobs/pending-flush.yaml` | Clears the pending list every Monday at 03:00, with its own minimal ServiceAccount |

The backend reads and writes these ConfigMaps in the namespace given by `CLIENTS_NAMESPACE` (default `dashboard`).

## Google OAuth Client

The Dex Google connector reads `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` from the `dex-github-oauth` secret in the `auth` namespace. The OAuth client is a *Web application* credential in Google Cloud with `https://dex.noodles.quest/callback` as its authorized redirect URI.
