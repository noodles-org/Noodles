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
│   │   ├── deployments.go     # CRUD operations on k8s deployments
│   │   ├── docs.go            # Serves docs TOC and markdown content
│   │   └── services.go        # Service directory (k8s discovery)
│   ├── middleware/
│   │   ├── auth.go            # JWT verification, role-based access (bypassed in dev)
│   │   ├── cors.go            # CORS middleware for development
│   │   └── logging.go         # Request logging middleware
│   ├── model/                 # Shared Go structs (deployment, doc, service, user)
│   ├── respond/
│   │   └── respond.go         # JSON response helpers
│   └── services/
│       ├── argocd.go          # ArgoCD API client for sync/health status
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
| GET | `/api/docs/toc` | Table of contents (parsed from `docs/toc.md`) |
| GET | `/api/docs/content?path=...` | Markdown content for a doc page |

## Static File Serving & SPA Fallback

In production, the backend serves the frontend's built assets and handles SPA routing via a `NotFound` handler:

1. Requests to `/api/*` that don't match a defined route return a JSON 404
2. Requests matching a static file in the frontend dist directory are served directly
3. All other requests serve `index.html`, allowing Vue Router to handle client-side routes (e.g. `/login`, `/services`, `/deployments`, `/docs`)

## Authentication

In production, authentication uses Dex OIDC:
1. `/login` redirects to Dex
2. Dex redirects back to `/callback` with an auth code
3. The backend exchanges the code for tokens, resolves the user's role from group membership, and sets an `httpOnly` JWT cookie

In development (`NODE_ENV=development`), auth is fully bypassed:
- `requireAuth` middleware injects a mock admin user
- `/login` issues a JWT cookie directly without contacting Dex

## ArgoCD Integration

The backend connects to ArgoCD at `http://argocd-server.argocd.svc.cluster.local` (plain HTTP, since ArgoCD runs with `server.insecure: true`) to fetch application sync and health status. An `ARGOCD_TOKEN` is required for API authentication.

## Kubernetes Integration

The backend connects to the k8s API using in-cluster credentials (production). In development, mock data is used instead. If neither is available, a warning is logged and k8s-dependent features degrade gracefully.

### Service Discovery

Services are discovered by listing IngressRoutes with the label `noodles.dashboard/service=true`. URLs are derived from route rules unless overridden by a `noodles.dashboard/url` annotation.

## Metrics

Prometheus metrics are exposed on a separate server on port 9090 at `/metrics`. A `ServiceMonitor` is configured for Prometheus scraping.

## Development

```bash
cd app
make dev-backend    # go run ./cmd/server
```

The dev server runs on `http://localhost:3000`. The frontend's Vite dev server proxies `/api` requests here.
