# Stalwart

Stalwart is a self-hosted mail server deployed on the cluster. It provides SMTP, IMAP, POP3, and ManageSieve services for the `noodles.quest` domain, with a web-based admin panel for configuration.

## Deployment

- **Image:** `docker.io/stalwartlabs/stalwart:latest`
- **Replicas:** 1
- **Namespace:** `stalwart`
- **Container ports:**
  - 8080 (HTTP, admin panel)
  - 443 (HTTPS)
  - 25, 587, 465 (SMTP)
  - 143, 993 (IMAP)
  - 110, 995 (POP3)
  - 4190 (ManageSieve)

## Storage

Two paths are persisted on a single 10Gi PVC (`stalwart-pvc`, `local-path` storageClass):

| Mount Path         | subPath  | Purpose                          |
|--------------------|----------|----------------------------------|
| `/etc/stalwart`    | config   | Server configuration files       |
| `/var/lib/stalwart`| db       | RocksDB application data         |

The PVC has the `argocd.argoproj.io/sync-options: Delete=false` annotation to prevent accidental deletion by ArgoCD.

## Networking

### Web Admin

- Routed via Traefik `IngressRoute` on the `web` and `websecure` entrypoints.
- **Admin URL:** `https://mail.noodles.quest`
- HTTPS redirect middleware is applied.
- TLS uses the shared `noodles-quest-prod-tls` wildcard certificate.

### Mail Protocols

Mail protocol ports (SMTP, IMAP, POP3, ManageSieve) are exposed via a `ClusterIP` service. These ports require additional external exposure (e.g., LoadBalancer or hostPort) since Traefik only handles HTTP/HTTPS traffic.

## DNS

Stalwart is configured to manage DNS records automatically via the Cloudflare API. This handles SPF, DKIM, DMARC, MX, and autoconfig records for the mail domain. A Cloudflare API token with DNS edit permissions is stored as a Kubernetes Secret in the `stalwart` namespace.
