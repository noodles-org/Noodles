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
│   ├── auth.ts            # User state, login/logout, auth check, role computeds
│   ├── clients.ts         # Client registry state, approve/reject/revoke actions
│   └── deployments.ts     # Deployment list state and actions
├── views/
│   ├── LoginView.vue      # Login page
│   ├── ServicesView.vue   # Service directory grid
│   ├── DeploymentsView.vue # Deployment management table
│   ├── DocsView.vue       # Documentation reader with sidebar
│   ├── PendingView.vue    # Waiting room for unapproved users
│   └── ClientsView.vue    # Admin-only client approval page
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
Lists deployments from `/api/deployments` with health status indicators. Users whose role passes `auth.canMutate` (`admin`, `client_admin`) can restart, pause, and resume deployments; the action block on `DeploymentCard.vue` is hidden for everyone else.

### Docs
Loads the table of contents from `/api/docs/toc` into a sidebar. Selecting an item fetches the markdown from `/api/docs/content?path=...`, renders it with `marked`, and sanitizes it with `DOMPurify`.

### Pending
The waiting room at `/pending` shown to users whose access request has not been approved yet. It has no data access, and `App.vue` hides the `NavBar` for pending users so no navigation is offered.

### Clients
The admin-only page at `/admin/clients` lists pending requests and approved clients from `/api/clients*`, with Approve (plus a role selector) and Reject controls on pending rows, and a Revoke control on approved rows. Revoke asks for confirmation first via a native `confirm()` prompt. Empty sections render a left-aligned `.clients-empty` placeholder rather than the centered global `.empty` utility.

## Auth Flow

The auth store checks `/api/auth/me` on app load. If unauthenticated, the router guard redirects to the login view. The login button navigates to `/api/auth/login`, which handles the OAuth flow (or dev bypass) server-side. Login errors are surfaced by `LoginView.vue` from the `?error=` query parameter (`not_authorized`, `email_unverified`, `requests_closed`).

### Roles and Guards

The `Role` union in `types/index.ts` covers `admin`, `viewer`, `client_admin`, `client`, and `pending`. The auth store exposes `isAdmin`, `isStaff`, `isPending`, and `canMutate`.

The router guard:
- redirects unauthenticated users to `/login`
- pins `pending` users to `/pending` and pushes everyone else away from it
- restricts `/admin/clients` to admins

The **Clients** nav entry in `NavBar.vue` is only rendered for admins. See [Client Access](../foundry-cluster/client-access.md) for the overall model.

## API Client

The Axios client (`api/client.ts`) is configured with `withCredentials: true` for cookie-based auth. On 401 responses, it redirects to the login page.

## Development

```bash
cd app/frontend
npm run dev       # Vite dev server with HMR on http://localhost:5173
```

The Vite config proxies `/api` requests to `http://localhost:3000` (the backend dev server).
