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

### Resources

| Resource | Request | Limit  |
|----------|---------|--------|
| CPU      | `1`     | `2`    |
| Memory   | `4Gi`   | `16Gi` |

Satisfactory's game simulation is largely single-threaded, so per-core clock speed matters more than core count. The CPU limit is kept modest (`2`: one core for the simulation tick, one for background work such as nav-mesh build, autosave compression, and networking) to avoid CFS throttling.

## Mod manager sidecar

The deployment runs a second `mod-manager` container alongside the game server:

- **Image:** `mephalrith/mod-manager:v0.1`
- **Purpose:** manage Satisfactory mods via the `ficsit` CLI (SMM) against the shared game files.
- **Environment:** `HOME` is set to `/config/ficsit-home`.
- **Volume:** mounts the same `satisfactory-config` PVC at `/config`, so it shares the game files and configuration with the game server.
- **Security context:** runs as non-root (UID/GID 1000), matching the game container.
- **Command:** currently idles (`while true; do sleep 30; done;`) so that the `ficsit` CLI can be accessed at any time.

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
| `messaging` | 8888 | TCP      | 32172    |

Connect at `<EXTERNAL_IP>:7777`.
