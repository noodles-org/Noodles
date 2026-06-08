# Foundry Cluster

The foundry cluster runs on a K3s single-node cluster hosted on a TrueNAS VM. It serves the FoundryVTT application and supporting infrastructure for the `noodles.quest` domain.

## Cluster Overview

| Component       | Tool / Version                  | Purpose                              |
|-----------------|---------------------------------|--------------------------------------|
| Kubernetes      | K3s                             | Lightweight Kubernetes distribution  |
| Ingress         | Traefik (bundled with K3s)      | Reverse proxy and ingress controller |
| TLS             | cert-manager + Let's Encrypt    | Automated wildcard certificates      |
| GitOps          | ArgoCD                          | Continuous deployment from Git       |
| Monitoring      | Prometheus + Grafana + Loki     | Metrics, dashboards, and logs        |
| Auth            | Dex                             | OIDC provider for cluster auth       |
| DNS             | Cloudflare                      | DNS management and record updates    |
| Secrets         | SOPS                            | Encrypted secrets in version control |
| Provisioning    | Ansible                         | Cluster setup and configuration      |

### Namespaces

- **foundry** — FoundryVTT application, backups, and CronJobs.
- **argocd** — ArgoCD server and application definitions.
- **monitoring** — Prometheus stack, Grafana, Loki, Alloy, and Pushgateway.
- **auth** — Dex OIDC provider, RBAC resources, and cert-manager TLS certificates.
- **traefik** — Traefik ingress route overrides.
- **pihole** — Pi-hole DNS sinkhole and ad blocker.
- **stalwart** — Stalwart self-hosted mail server.

### Cluster Setup

The cluster is provisioned via Ansible using `make setup-cluster`. The playbook (`setup_cluster.yaml`) performs the following:

1. Configures K3s API server authentication with Dex.
2. Installs cert-manager for TLS certificate management.
3. Adds Helm repos (ArgoCD, Prometheus, Dex) and installs their charts.
4. Applies Kubernetes manifests for auth, foundry, traefik, jellyfin, monitoring, Pi-hole, Stalwart, and ArgoCD resources.
