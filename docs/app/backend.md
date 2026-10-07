# Backend

Go API server (Chi router) running on port 3000.

## Architecture

```
backend/
├── cmd/server/
│   └── main.go                # Chi app setup, middleware, route mounting, SPA fallback
├── internal/
│   ├── config/
│   │   └── config.go          # Centralized configuration from env vars
│   ├── errs/                  # Typed error definitions (auth, oauth, resource, etc.)
│   ├── handlers/
│   │   ├── auth.go            # OAuth login/callback/logout via Dex
│   │   ├── clients.go         # Client registry API (list, approve, reject, revoke)
│   │   ├── deployments.go     # CRUD operations on k8s deployments
│   │   ├── docs.go            # Serves docs TOC and markdown content
│   │   ├── files.go           # Foundry file management (proxied to the file sidecar)
│   │   └── services.go        # Service directory (k8s discovery)
│   ├── middleware/
│   │   ├── auth.go            # JWT verification, role-based access (bypassed in dev)
│   │   ├── cors.go            # CORS middleware for development
│   │   └── logging.go         # Request logging middleware
│   ├── model/                 # Shared Go structs (client, deployment, doc, service, user)
│   ├── respond/
│   │   └── respond.go         # JSON response helpers
│   └── services/
│       ├── argocd.go          # ArgoCD API client for sync/health status
│       ├── clients.go         # ConfigMap-backed client registry with cache
│       ├── files.go           # File service: sidecar HTTP client (prod) / jailed local FS (dev)
│       ├── kubernetes.go      # K8s client: namespace/deployment/service discovery
│       ├── logger.go          # Structured slog logger
│       └── metrics.go         # Prometheus metrics via client_golang
└── mocks/                     # Test mocks
```

## API Routes

All routes except `/healthz` and `/api/auth/*` require authentication.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health check |
| GET | `/api/auth/login` | Initiate Dex OAuth (or dev bypass) |
| GET | `/api/auth/callback` | OAuth callback |
| POST | `/api/auth/logout` | Clear session cookie |
| GET | `/api/auth/me` | Current user info |
| GET | `/api/deployments` | List deployments in managed namespaces |
| POST | `/api/deployments/{namespace}/{name}/restart` | Restart a deployment |
| POST | `/api/deployments/{namespace}/{name}/pause` | Scale to 0, saving original replicas |
| POST | `/api/deployments/{namespace}/{name}/resume` | Restore original replica count |
| GET | `/api/services` | List discovered services |
| GET | `/api/files` | List entries at `?path=` under the Foundry `Data` root |
| GET | `/api/files/download` | Stream a file's bytes at `?path=` |
| POST | `/api/files/upload` | Multipart upload into `?path=` (admin/client_admin) |
| DELETE | `/api/files` | Delete a file or empty dir at `?path=` (admin/client_admin) |
| POST | `/api/files/mkdir` | Create a folder at `?path=` (admin/client_admin) |
| POST | `/api/files/rename` | `{from, to}` — rename or move (admin/client_admin) |
| GET | `/api/docs/toc` | Table of contents (parsed from `docs/toc.md`) |
| GET | `/api/docs/content?path=...` | Markdown content for a doc page |
| GET | `/api/clients` | List approved clients (admin only) |
| GET | `/api/clients/pending` | List pending access requests (admin only) |
| POST | `/api/clients/approve` | `{email, role}` — move pending → approved (admin only) |
| POST | `/api/clients/reject` | `{email}` — drop from pending (admin only) |
| POST | `/api/clients/revoke` | `{email}` — drop from approved (admin only) |

## Static File Serving & SPA Fallback

In production, the backend serves the frontend's built assets and handles SPA routing via a `NotFound` handler:

1. Requests to `/api/*` that don't match a defined route return a JSON 404
2. Requests matching a static file in the frontend dist directory are served directly
3. All other requests serve `index.html`, allowing Vue Router to handle client-side routes (e.g. `/login`, `/services`, `/deployments`, `/docs`, `/pending`, `/admin/clients`)

## Authentication

In production, authentication uses Dex OIDC:
1. `/login` redirects to Dex
2. Dex redirects back to `/callback` with an auth code
3. The backend exchanges the code for tokens, resolves the user's role, and sets an `httpOnly` JWT cookie

`resolveIdentity` in `handlers/auth.go` resolves the role in order: GitHub group match → staff (`admin` / `viewer`), then the client registry → `client` / `client_admin`, then the pending list → `pending`, otherwise a new pending entry is recorded. See [Client Access](../foundry-cluster/client-access.md) for the identity model.

### Roles and Route Gating

`model.Role` has five values — `admin`, `viewer`, `client_admin`, `client`, `pending` — with the predicates `IsStaff()`, `CanRead()`, and `CanMutate()`.

| Middleware | Applied to | Effect |
|------------|-----------|--------|
| `RequireAuth` | all routes except `/healthz` and `/api/auth/login`, `/api/auth/callback` | Parses the session JWT into `model.User` |
| `RequireApproved` | `/api/deployments`, `/api/services`, `/api/docs` | 403 for any role failing `CanRead()` (i.e. `pending`) |
| `RequireApproved` | `/api/files` | 403 for any role failing `CanRead()` (i.e. `pending`); covers list/download |
| `RequireRole(admin, client_admin)` | restart / pause / resume, file upload / delete / mkdir / rename | Mutating deployment and file actions |
| `RequireRole(admin)` | `/api/clients/*` | Staff-admin-only approval API |

### Client Registry

`services.ClientRegistry` (`services/clients.go`) stores clients in the `clients.yaml` key of the `noodles-clients` and `noodles-clients-pending` ConfigMaps, in the namespace given by `CLIENTS_NAMESPACE` (default `dashboard`). Both lists are cached in memory, refreshed on a 30s ticker and invalidated synchronously on every write. Lookups are case-insensitive, `AddPending` is idempotent and capped at 20 entries, and malformed YAML degrades to an empty list with an error log rather than panicking. `Approve` moves an entry from the pending to the approved ConfigMap, `Reject` drops it from pending, and `Revoke` removes an already approved client from the approved ConfigMap. Sentinel errors live in `errs/clients.go` (`PendingFull` 503, `NotPending` 404, `NotApproved` 404, `InvalidClientRole` 400) and login rejections in `errs/identity.go` (`NotAuthorized`, `EmailUnverified`, `RequestsClosed`, `RegistryFailed`).

In development (`NODE_ENV=development`), auth is fully bypassed:
- `requireAuth` middleware injects a mock admin user
- `/login` issues a JWT cookie directly without contacting Dex
- The client registry runs purely in memory, since there is no cluster to read ConfigMaps from. It is seeded once at startup from `mocks/clients.json`: every entry lands in the pending list with role `pending` and the approved list starts empty, so approve / reject / revoke can be exercised locally. Nothing is written back to disk, and the backend must be run from `app/backend` for the relative `mocks` path to resolve

## File Management

`services.FileService` (`services/files.go`) is the backend's only path to the Foundry files; clients never talk to the sidecar directly.

- **Production:** an HTTP client to the file sidecar at `FILESVC_URL` (default `http://file-sidecar.foundry.svc.cluster.local`), sending `Authorization: Bearer <FILESVC_TOKEN>`. Non-2xx responses are translated into typed errors (e.g. `409` → `DirNotEmpty`, `404` → `FileNotFound`, in `errs/resource.go`). Downloads are streamed straight through.
- **Development (`NODE_ENV=development`):** operations run against a local jailed directory rooted at `FILESVC_ROOT` (default `mocks/files`), so `make dev` works with no cluster. On startup the service copies the committed `mocks/files` seed into an ephemeral `files-work-*` temp dir and points dev operations at the copy, so mutations never touch the seed; the temp dir is removed on shutdown (SIGINT/SIGTERM or normal exit), so the tree resets on each backend restart.

`handlers/files.go` re-validates every `path` (rejecting `..` and absolute paths) as defense in depth even though the sidecar enforces its own jail, audit-logs the acting `user.Email` on mutations, and increments the `dashboard_file_actions_total` metric. The route group's config lives in `FileSvcConfig` (`FILESVC_URL`, `FILESVC_TOKEN`, `FILESVC_ROOT`) in `config/config.go`. See [Foundry](../foundry-cluster/foundry.md#file-sidecar) for the sidecar side.

## ArgoCD Integration

The backend connects to ArgoCD at `http://argocd-server.argocd.svc.cluster.local` (plain HTTP, since ArgoCD runs with `server.insecure: true`) to fetch application sync and health status. An `ARGOCD_TOKEN` is required for API authentication.

## Kubernetes Integration

The backend connects to the local k8s API using in-cluster credentials (production). Remote clusters are configured via the `REMOTE_CLUSTERS` environment variable — a JSON array where each entry specifies a cluster name, API URL, token env var name, and inline CA PEM. Tokens are resolved from the referenced env vars at startup. In development, mock data is used instead. If neither in-cluster credentials nor remote cluster config is available, a warning is logged and k8s-dependent features degrade gracefully.

Deployment actions (restart, pause, resume) are routed to the correct cluster based on namespace ownership.

### Service Discovery

Services are discovered by listing IngressRoutes with the label `noodles.dashboard/service=true`. URLs are derived from route rules unless overridden by a `noodles.dashboard/url` annotation.

## Metrics

Prometheus metrics are exposed on a separate server on port 9090 at `/metrics`. A `ServiceMonitor` is configured for Prometheus scraping.

Custom metrics:

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `dashboard_auth_events_total` | Counter | `status`, `reason` | Auth events (login, logout, failures) |
| `dashboard_unique_authenticated_users` | Gauge | — | Unique users since last restart |
| `dashboard_deployment_actions_total` | Counter | `action`, `namespace`, `deployment` | Deployment actions (pause, resume, restart) |
| `dashboard_file_actions_total` | Counter | `action`, `status` | Foundry file management actions (list, download, upload, delete, mkdir, rename) |
| `dashboard_unauthorized_access_attempts_total` | Counter | `path` | Admin action attempts without admin role |
| `dashboard_http_requests_total` | Counter | `method`, `route`, `status_code` | HTTP requests |
| `dashboard_http_request_duration_seconds` | Histogram | `method`, `route` | HTTP request duration |

Notes:
- `/healthz` requests are excluded from HTTP metrics to avoid noise from k8s probes.
- Requests to unregistered routes are grouped under the `"unmatched"` route label.
- Missing auth tokens (unauthenticated requests to protected endpoints) are not recorded as auth failures — only invalid/expired tokens are.

## Development

```bash
cd app
make dev-backend    # go run ./cmd/server
```

The dev server runs on `http://localhost:3000`. The frontend's Vite dev server proxies `/api` requests here.
