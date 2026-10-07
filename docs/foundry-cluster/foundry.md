# Foundry

FoundryVTT is the primary application running on the cluster. It uses a custom Node.js-based Docker image.

## Deployment

- **Image:** `docker.io/mephalrith/foundry-server:latest`
- **Replicas:** 1
- **Container port:** 30000
- **Namespace:** `foundry`
- **Requests:** 500m CPU, 1Gi memory, 1Gi ephemeral-storage
- **Limits:** 4 CPU, 3Gi memory, 4Gi ephemeral-storage

The image is intentionally left on the `latest` tag with `imagePullPolicy: Always` so a rebuild rolls out without a manifest change.

The Foundry version is updated by rebuilding the Docker image:
```
make update-foundry-version URL=<timed_url>
```

## Storage

Foundry data is persisted using a `PersistentVolumeClaim`:

- **PVC name:** `foundry-pvc`
- **Storage class:** `local-path`
- **Size:** 80Gi
- **Access mode:** ReadWriteOnce
- **Deletion protection:** The PVC is annotated with `argocd.argoproj.io/sync-options: Delete=false` to prevent accidental deletion during ArgoCD syncs.

Data is mounted at `/foundrydata` inside the container.

## File Sidecar

A second container, `file-sidecar`, rides inside the `foundry` pod to power the dashboard's **Foundry Files** feature. It is a small custom Go HTTP service (`docker.io/mephalrith/foundry-file-sidecar:latest`) that shares the same `foundry-pv-storage` volume mounted at `/foundrydata`, exposing a narrow internal API (list, download, upload, delete, mkdir, rename) over the files it manages.

- **Jail root:** every operation is confined to `FILESVC_ROOT` (`/foundrydata/Data`); paths that escape the root via `..`, absolute paths, or symlinks are rejected. The root itself cannot be deleted or renamed.
- **Delete semantics:** non-recursive — a file or an *empty* directory can be removed; a non-empty directory returns `409 Conflict`.
- **Auth:** a shared bearer token (`FILESVC_TOKEN`, from the `file-sidecar` Secret) is required on every request; the dashboard backend is the only intended caller.
- **Port / metrics:** listens on `8080` (`FILESVC_PORT`), serving `/healthz` and Prometheus `/metrics` (`filesvc_actions_total`). The pod's `prometheus.io/*` annotations point the scrape at this port.
- **Service:** an internal `ClusterIP` Service `file-sidecar` (no IngressRoute) reachable in-cluster at `http://file-sidecar.foundry.svc.cluster.local`.

The sidecar source, Dockerfile, and `make update-file-sidecar` build target live under `k8s/foundry/file-sidecar/`. Because it rides inside the existing `foundry` deployment, it needs no separate ArgoCD Application. See [Dashboard](dashboard.md) for the user-facing feature and [Client Access](client-access.md#roles) for the permission mapping.

## Networking

Foundry is exposed via Traefik IngressRoutes:

- **External URL:** `https://foundry.noodles.quest`
- A `ClusterIP` Service maps port 80 → container port 30000.
- A `TraefikService` (weighted) routes traffic to the ClusterIP Service.
- An `IngressRoute` matches `Host(foundry.noodles.quest)` on both `web` and `websecure` entrypoints.
- A `secure-redirect` middleware enforces HTTPS.

## Backups

Foundry data is backed up using Restic to an S3 bucket:

- **CronJob:** `foundry-backup`
- **Schedule:** Mondays at 9:00 AM UTC
- **Destination:** `s3:https://s3.us-west-1.amazonaws.com/noodles-foundry-bucket/restic`
- **Excludes:** `/foundrydata/Backups/*`
- **Credentials:** Stored in the `restic` Kubernetes Secret (AWS keys + Restic password).

Manual backup and restore jobs are available under `k8s/_manual-jobs/foundry/`:

- **Jobs:** `foundry-manual-backup`, `foundry-restore`
- **Not managed by ArgoCD:** they live outside every ArgoCD Application source path (and are not applied by `setup_cluster.yaml`), so ArgoCD never creates, recreates, or prunes them.
- **Usage:** apply them by hand when needed, and delete the completed Job before re-running it:

```
kubectl apply -f foundry_deployment/infra/k8s/_manual-jobs/foundry/manual-backup.yaml
kubectl delete job -n foundry foundry-manual-backup
```

## Fetch IP CronJob

A custom Go application updates Cloudflare DNS records with the cluster's current public IP:

- **CronJob:** `fetch-vm-public-ip`
- **Schedule:** Every 15 minutes
- **Image:** `docker.io/mephalrith/fetch-foundry-ip:v0.1`
- Pushes metrics to the Prometheus Pushgateway for monitoring.
- Uses Cloudflare API credentials from the `cloudflare` Kubernetes Secret.
