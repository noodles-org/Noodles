# Valheim

Valheim is a dedicated game server for multiplayer Viking survival sessions.

## Deployment

- **Image:** `lloesche/valheim-server:latest`
- **Replicas:** 1
- **Namespace:** `valheim`
- **Update strategy:** Recreate

Key environment variables:

| Variable       | Value        | Source            |
|----------------|--------------|-------------------|
| `SERVER_NAME`  | `noodleheim` | Inline            |
| `WORLD_NAME`   | `EggNoodle`  | Inline            |
| `SERVER_PASS`  | —            | `valheim-secret`  |
| `SERVER_PORT`  | `2456`       | Inline            |

## Storage

Valheim uses two PersistentVolumeClaims:

| PVC Name                   | Size | Mount Path     | Purpose      |
|----------------------------|------|----------------|--------------|
| `valheim-world-data`       | 2Gi  | `/config`      | World saves  |
| `valheim-server-base-data` | 10Gi | `/opt/valheim` | Server files |

Both use `local-path` storage class with `ReadWriteOnce` access and are annotated with `argocd.argoproj.io/sync-options: Delete=false` for deletion protection.

## Networking

Valheim is exposed via a `LoadBalancer` Service:

| Port Name       | Port | Protocol  | NodePort |
|-----------------|------|-----------|----------|
| `gameport-udp`  | 2456 | UDP       | 30743    |
| `gameport-tcp`  | 2456 | TCP       | 30743    |
| `queryport-udp` | 2457 | UDP       | 31994    |
| `queryport-tcp` | 2457 | TCP       | 31994    |

Connect at `<EXTERNAL_IP>:2457`.
