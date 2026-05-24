# Jellyfin

Jellyfin is a free, open-source media server deployed on the cluster for streaming media content.

## Deployment

- **Image:** `docker.io/jellyfin/jellyfin:latest`
- **Replicas:** 1
- **Container ports:** 8096 (TCP, HTTP web UI), 7359 (UDP, client discovery)
- **Namespace:** `jellyfin`

## Storage

Jellyfin configuration and data is persisted using a `PersistentVolumeClaim`:

- **PVC name:** `jellyfin-pvc`
- **Storage class:** `local-path`
- **Size:** 5Gi
- **Access mode:** ReadWriteOnce
- **Deletion protection:** The PVC is annotated with `argocd.argoproj.io/sync-options: Delete=false` to prevent accidental deletion during ArgoCD syncs.

Data is mounted at `/config` inside the container, which stores server settings, the user database, and installed plugins.

## Networking

Jellyfin is exposed via Traefik IngressRoutes:

- **External URL:** `https://jellyfin.noodles.quest`
- A `ClusterIP` Service maps port 80 → container port 8096 (TCP) and port 7359 → container port 7359 (UDP).
- A `TraefikService` (weighted) routes traffic to the ClusterIP Service.
- An `IngressRoute` matches `Host(jellyfin.noodles.quest)` on both `web` and `websecure` entrypoints.
- A `secure-redirect` middleware enforces HTTPS.

## Plugins

Plugins can be installed through the Jellyfin web UI under **Dashboard → Plugins → Catalog**. Since the `/config` directory is persisted via the PVC, installed plugins survive pod restarts.
