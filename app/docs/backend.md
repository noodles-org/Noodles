# Backend

Express + TypeScript API server running on port 3000.

## Architecture

```
backend/src/
├── server.ts              # Express app setup, middleware, route mounting
├── config.ts              # Centralized configuration from env vars
├── env.ts                 # Loads .env in non-production
├── types.ts               # Shared TypeScript interfaces
├── middleware/
│   ├── auth.ts            # JWT verification, role-based access (bypassed in dev)
│   └── logging.ts         # Request logging middleware
├── routes/
│   ├── auth.ts            # OAuth login/callback/logout via Dex
│   ├── deployments.ts     # CRUD operations on k8s deployments
│   ├── docs.ts            # Serves docs TOC and markdown content
│   └── services.ts        # Service directory (k8s discovery with JSON fallback)
└── services/
    ├── kubernetes.ts       # K8s client: namespace/deployment/service discovery
    ├── logger.ts           # Winston logger (debug in dev, info in prod)
    └── metrics.ts          # Prometheus metrics via prom-client
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
| POST | `/api/deployments/:ns/:name/restart` | Restart a deployment |
| POST | `/api/deployments/:ns/:name/pause` | Scale to 0, saving original replicas |
| POST | `/api/deployments/:ns/:name/resume` | Restore original replica count |
| GET | `/api/services` | List discovered services |
| GET | `/api/docs/toc` | Table of contents (parsed from `docs/toc.md`) |
| GET | `/api/docs/content?path=...` | Markdown content for a doc page |

## Authentication

In production, authentication uses Dex OIDC:
1. `/login` redirects to Dex with PKCE
2. Dex redirects back to `/callback` with an auth code
3. The backend exchanges the code for tokens, resolves the user's role from group membership, and sets an `httpOnly` JWT cookie

In development (`NODE_ENV=development`), auth is fully bypassed:
- `requireAuth` middleware injects a mock admin user
- `/login` issues a JWT cookie directly without contacting Dex

## Kubernetes Integration

The backend connects to the k8s API using in-cluster credentials (production) or local kubeconfig (development). If neither is available, a debug-level warning is logged and k8s-dependent features degrade gracefully.

### Service Discovery

Services are discovered by listing IngressRoutes with the label `noodles.dashboard/service=true`. URLs are derived from route rules unless overridden by a `noodles.dashboard/url` annotation.

## Metrics

Prometheus metrics are exposed on port 9090 at `/metrics`. Custom metrics include auth events (login success/failure counts).

## Development

```bash
cd app/backend
npm run dev-server    # tsx watch with hot reload
```

The dev server runs on `http://localhost:3000`. The frontend's Vite dev server proxies `/api` requests here.
