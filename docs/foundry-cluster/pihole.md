# Pi-hole

Pi-hole is a network-wide DNS sinkhole and ad blocker deployed on the cluster. It provides DNS filtering for the local network and a web-based admin dashboard.

## Deployment

- **Image:** `docker.io/pihole/pihole:latest`
- **Replicas:** 1
- **Container ports:** 80 (TCP, web UI), 53 (TCP/UDP, DNS)
- **Namespace:** `pihole`

## Networking

Pi-hole uses a `ClusterIP` service with Traefik routing for both the admin UI and DNS.

### Traefik DNS Entrypoints

Traefik must be configured with `dns-tcp` (TCP/53) and `dns-udp` (UDP/53) entrypoints for DNS routing. The `HelmChartConfig` in `foundry_deployment/infra/k8s/traefik/config.yaml` adds these entrypoints and sets `service.single: false` to create separate TCP and UDP Kubernetes Services (required because Kubernetes doesn't support mixed TCP/UDP on a single Service). This is applied automatically during cluster setup.

### Web UI

- Routed via Traefik `IngressRoute` on the `web` entrypoint.
- **Admin URL:** `http://pihole.noodles.local/admin`
- Requires a `/etc/hosts` entry: `10.0.128.157  pihole.noodles.local`

### DNS (port 53)

- Routed via Traefik `IngressRouteTCP` and `IngressRouteUDP` on the `dns-tcp` and `dns-udp` entrypoints.
- Point your router's DNS setting to the node IP (`10.0.128.157`).

## OPNsense Router Configuration

To use Pi-hole as the DNS server for your network:

1. **System DNS:** Go to System → Settings → General and add `10.0.128.157` as a DNS server. Uncheck "Allow DNS server list to be overridden by DHCP/PPP on WAN".
2. **DHCP DNS:** Go to Services → ISC DHCPv4 → [LAN] and add `10.0.128.157` under DNS servers.
3. **Verify:** Run `nslookup google.com 10.0.128.157` from a client to confirm DNS resolution through Pi-hole.
