# Noodles Dashboard — Application Docs

Developer documentation for the Noodles Dashboard frontend and backend.

- [Backend](backend.md)
- [Frontend](frontend.md)

## Quick Start

```bash
cd app

# Install dependencies
make install-all

# Start both frontend and backend in dev mode
make dev
```

The frontend runs at `http://localhost:5173` and proxies API requests to the backend at `http://localhost:3000`.

In development, OAuth is bypassed — hitting `/api/auth/login` issues a JWT for a mock admin user without requiring Dex.

## Project Structure

```
app/
├── backend/          # Express API server (TypeScript)
│   └── mocks/        # Mock data for dev (no k8s cluster needed)
├── frontend/         # Vue 3 SPA (TypeScript + Vite)
├── docs/             # This documentation
├── .env              # Local environment (gitignored)
├── .env.example      # Template for .env
├── Makefile          # Dev, build, and deploy commands
Dockerfile            # (project root) Multi-stage Docker build
```

## Makefile Targets

| Target                  | Description |
|-------------------------|-------------|
| `make dev`              | Run frontend + backend concurrently |
| `make install-all`      | `npm ci` for both frontend and backend |
| `make update-dashboard` | Build, push, and rolling restart |

## Environment Variables

Copy `.env.example` to `.env`. In development, only `NODE_ENV=development` is needed — all other variables have sensible defaults, and auth/k8s features are bypassed. When `NODE_ENV=development`, the backend serves mock deployment and service data from `app/backend/mocks/` and reads docs from the repo's `docs/` directory directly.

In production, the following must be set (via k8s Secrets): `DASHBOARD_CLIENT_SECRET`, `JWT_SECRET`, `ARGOCD_TOKEN`.
