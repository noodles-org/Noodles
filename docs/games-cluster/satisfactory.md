# Satisfactory

Satisfactory is a dedicated game server for multiplayer factory-building sessions.

## Deployment

- **Image:** `wolveix/satisfactory-server:latest`
- **Replicas:** 1
- **Namespace:** `satisfactory`

Key environment variables:

| Variable       | Value   |
|----------------|---------|
| `AUTOPAUSE`    | `true`  |
| `MAXPLAYERS`   | `8`     |
| `ROOTLESS`     | `true`  |
| `SKIPUPDATE`   | `false` |
| `STEAMBETA`    | `false` |

The container runs as non-root (UID/GID 1000) with a liveness probe using the built-in health check script.

## Storage

- **PVC name:** `satisfactory-config-pvc`
- **Storage class:** `local-path`
- **Size:** 30Gi
- **Access mode:** ReadWriteOnce
- **Mount path:** `/config`
- **Deletion protection:** The PVC is annotated with `argocd.argoproj.io/sync-options: Delete=false` to prevent accidental deletion during ArgoCD syncs.

## Networking

Satisfactory is exposed via a `LoadBalancer` Service:

| Port Name   | Port | Protocol | NodePort |
|-------------|------|----------|----------|
| `api`       | 7777 | TCP      | 32171    |
| `game`      | 7777 | UDP      | 32171    |
| `messaging` | 8888 | TCP      | —        |

Connect at `<EXTERNAL_IP>:7777`.
