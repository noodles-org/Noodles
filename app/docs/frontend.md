# Frontend

Vue 3 SPA built with Vite and TypeScript.

## Architecture

```
frontend/src/
├── App.vue                # Root component with NavBar and router-view
├── main.ts                # App entry point, Pinia + Router setup
├── api/
│   └── client.ts          # Axios instance with 401 redirect handling
├── router/
│   └── index.ts           # Vue Router with auth guard
├── stores/
│   ├── auth.ts            # User state, login/logout, auth check
│   └── deployments.ts     # Deployment list state and actions
├── views/
│   ├── LoginView.vue      # Login page
│   ├── ServicesView.vue   # Service directory grid
│   ├── DeploymentsView.vue # Deployment management table
│   └── DocsView.vue       # Documentation reader with sidebar
└── components/
    ├── NavBar.vue          # Top navigation bar
    ├── ServiceCard.vue     # Service link card
    ├── DeploymentCard.vue  # Deployment status and action card
    └── DocsSidebar.vue     # Docs table of contents sidebar
```

## Views

### Services
Fetches the service list from `/api/services` and renders a card grid grouped by category. Each card links to the service URL.

### Deployments
Lists deployments from `/api/deployments` with health status indicators. Admins can restart, pause, and resume deployments.

### Docs
Loads the table of contents from `/api/docs/toc` into a sidebar. Selecting an item fetches the markdown from `/api/docs/content?path=...`, renders it with `marked`, and sanitizes it with `DOMPurify`.

## Auth Flow

The auth store checks `/api/auth/me` on app load. If unauthenticated, the router guard redirects to the login view. The login button navigates to `/api/auth/login`, which handles the OAuth flow (or dev bypass) server-side.

## API Client

The Axios client (`api/client.ts`) is configured with `withCredentials: true` for cookie-based auth. On 401 responses, it redirects to the login page.

## Development

```bash
cd app/frontend
npm run dev       # Vite dev server with HMR on http://localhost:5173
```

The Vite config proxies `/api` requests to `http://localhost:3000` (the backend dev server).
