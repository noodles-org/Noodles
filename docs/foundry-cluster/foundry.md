# Foundry

FoundryVTT is the primary application running on the cluster. It uses a custom Node.js-based Docker image.

## Deployment

- **Image:** `docker.io/mephalrith/foundry-server:latest`
- **Replicas:** 1
- **Container port:** 30000
- **Namespace:** `foundry`

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

Manual backup and restore jobs are also available under `k8s/foundry/jobs/manual/`.

## Fetch IP CronJob

A custom Go application updates Cloudflare DNS records with the cluster's current public IP:

- **CronJob:** `fetch-vm-public-ip`
- **Schedule:** Every 15 minutes
- **Image:** `docker.io/mephalrith/fetch-foundry-ip:v0.1`
- Pushes metrics to the Prometheus Pushgateway for monitoring.
- Uses Cloudflare API credentials from the `cloudflare` Kubernetes Secret.
