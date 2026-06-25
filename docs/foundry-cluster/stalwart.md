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

### Inbound Webhook Sidecar

A Go-based sidecar container (`mephalrith/stalwart-inbound-webhook`) handles inbound email delivery from Resend. It listens on port 3000 and receives webhook POSTs at `/webhook/inbound`. When an email arrives, it fetches the raw RFC822 message from the Resend API and delivers it to Stalwart via local SMTP (localhost:25).

- **Image:** `docker.io/mephalrith/stalwart-inbound-webhook:latest`
- **Source:** `infra/k8s/stalwart/inbound-webhook/`
- **Port:** 3000

## Storage

Two PVCs persist Stalwart's data (`local-path` storageClass):

| PVC                    | Size | Mount Path          | Purpose                    |
|------------------------|------|---------------------|----------------------------|
| `stalwart-config-pvc`  | 1Gi  | `/etc/stalwart`     | Server configuration files |
| `stalwart-data-pvc`    | 10Gi | `/var/lib/stalwart` | RocksDB application data   |

Both PVCs have the `argocd.argoproj.io/sync-options: Delete=false` annotation to prevent accidental deletion by ArgoCD.

## Networking

### Web Admin

- Routed via Traefik `IngressRoute` on the `web` and `websecure` entrypoints.
- **Admin URL:** `https://mail.noodles.quest/admin`
- HTTPS redirect middleware is applied.
- TLS uses the shared `noodles-quest-prod-tls` wildcard certificate.

### Mail Protocols

Mail protocol ports (SMTP, IMAP, POP3, ManageSieve) are exposed directly on the node using `hostPort` mappings, bypassing Traefik. OPNsense port forwarding rules are required for external access on ports 25, 465, 587, and 993.

### Inbound Webhook

The `/webhook/inbound` path on `mail.noodles.quest` is routed to the webhook sidecar (port 3000) via a dedicated Traefik route rule. All other paths are routed to Stalwart's web UI.

## DNS

Stalwart is configured to manage DNS records automatically via the Cloudflare API. This handles SPF, DKIM, DMARC, and autoconfig records for the mail domain. A Cloudflare API token with DNS edit permissions is stored as a Kubernetes Secret in the `stalwart` namespace.

An explicit A record for `mail.noodles.quest` is required in Cloudflare (the wildcard `*.noodles.quest` record is insufficient due to CNAME interactions). The `mail.noodles.quest` DNS record ID is included in the Foundry fetch-ip CronJob for automatic IP updates.

## Email Routing

Due to Comcast blocking port 25 in both directions for residential connections:

| Direction    | Method                                                              |
|--------------|---------------------------------------------------------------------|
| **Outbound** | Stalwart relays through Resend SMTP (`smtp.resend.com:465`)        |
| **Inbound**  | Resend receives mail (MX), webhooks to the inbound sidecar, which delivers to Stalwart via local SMTP |
