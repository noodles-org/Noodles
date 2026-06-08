# Pi-hole

Pi-hole is a network-wide DNS sinkhole and ad blocker deployed on the cluster. It provides DNS filtering for the local network and a web-based admin dashboard.

## Deployment

- **Image:** `docker.io/pihole/pihole:latest`
- **Replicas:** 1
- **Container ports:** 80 (TCP, web UI), 53 (TCP/UDP, DNS)
- **Namespace:** `pihole`

## Networking

### DNS (port 53)

DNS ports are exposed directly on the node using `hostPort: 53` on both TCP and UDP. This bypasses Traefik and kube-proxy, allowing Pi-hole to see the real client IP addresses in its query logs. Point your router's DNS setting to the node IP (`10.0.128.157`).

The `FTLCONF_dns_listeningMode` environment variable is set to `"all"` to accept DNS queries from all network origins. This is required because the `hostPort` setup causes queries to arrive from IPs on the `10.0.128.x` subnet, which Pi-hole's default `LOCAL` mode does not recognize as local. This is safe as long as port 53 is not forwarded on the router.

### Web UI

- Routed via Traefik `IngressRoute` on the `web` entrypoint.
- **Admin URL:** `http://pihole.noodles.local/admin`
- Requires a `/etc/hosts` entry: `10.0.128.157  pihole.noodles.local`

## Node DNS Configuration

Because the node hosts Pi-hole, it cannot use Pi-hole as its own DNS server (circular dependency). The `setup_cluster.yaml` playbook configures the node to:

1. Use `8.8.8.8` as a static DNS server via a netplan override (`/etc/netplan/99-dns-override.yaml`), with `dhcp4-overrides.use-dns: false` to ignore DHCP-provided DNS.
2. Bypass the `systemd-resolved` stub listener by symlinking `/etc/resolv.conf` to `/run/systemd/resolve/resolv.conf`.

## OPNsense Router Configuration

To use Pi-hole as the DNS server for your network:

1. **System DNS:** Go to System → Settings → General and add `10.0.128.157` as a DNS server. Add `8.8.8.8` as a second DNS server for fallback in case Pi-hole is down. Uncheck "Allow DNS server list to be overridden by DHCP/PPP on WAN".
2. **DHCP DNS:** Go to Services → ISC DHCPv4 → [LAN] and add `10.0.128.157` under DNS servers.
3. **Verify:** Run `nslookup google.com 10.0.128.157` from a client to confirm DNS resolution through Pi-hole.
