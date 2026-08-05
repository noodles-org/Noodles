# Enshrouded

Enshrouded is a dedicated game server for multiplayer survival sessions.

## Deployment

- **Image:** `sknnr/enshrouded-dedicated-server:latest`
- **Replicas:** 1
- **Namespace:** `enshrouded`

Key environment variables:

| Variable         | Value   | Source                |
|------------------|---------|-----------------------|
| `SERVER_NAME`    | —       | `enshrouded-secret`   |
| `SERVER_PASSWORD` | —      | `enshrouded-secret`   |
| `PORT`           | `15637` | Inline                |
| `STEAM_PORT`     | `27015` | Inline                |
| `SERVER_SLOTS`   | `10`    | Inline                |

The container runs as UID/GID 10000 with resource limits of 4 CPU / 8Gi memory.

## Storage

- **PVC name:** `enshrouded-data`
- **Storage class:** `local-path`
- **Size:** 30Gi
- **Access mode:** ReadWriteOnce
- **Mount path:** `/home/steam/enshrouded/savegame`
- **Deletion protection:** The PVC is annotated with `argocd.argoproj.io/sync-options: Delete=false` to prevent accidental deletion during ArgoCD syncs.

## Networking

Enshrouded is exposed via a `LoadBalancer` Service:

| Port Name        | Port  | Protocol | NodePort |
|------------------|-------|----------|----------|
| `query-port-udp` | 15637 | UDP      | 32638    |
| `steam-port-udp` | 27015 | UDP      | 31201    |

Connect at `<EXTERNAL_IP>:15637`.
