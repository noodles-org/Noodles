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

### Node Capacity

The single node (`foundry`) runs on a TrueNAS VM with **16 vCPUs** (2 CPUs x 2 cores x 4 threads) and **8 GiB of memory**. Its root disk is a **300 GiB** zvol (`foundryvm-szhcrl`) on the `foundry` pool, and it backs every `local-path` PersistentVolume.

Memory is the binding constraint, so workload limits are sized to fit within 8 GiB alongside the K3s control plane and monitoring stack. CPU limits are set generously because threads are plentiful.

| Workload            | CPU request / limit | Memory request / limit | Ephemeral request / limit |
|---------------------|---------------------|------------------------|---------------------------|
| foundry             | 500m / 4            | 1Gi / 3Gi              | 1Gi / 4Gi                 |
| jellyfin            | 250m / 6            | 512Mi / 1536Mi         | 512Mi / 2Gi               |
| stalwart            | 100m / 1            | 256Mi / 1Gi            | 256Mi / 2Gi               |
| inbound-webhook     | 10m / 200m          | 32Mi / 128Mi           | 64Mi / 256Mi              |
| pihole              | 50m / 500m          | 128Mi / 512Mi          | 128Mi / 1Gi               |
| noodles-dashboard   | 50m / 200m          | 128Mi / 256Mi          | 64Mi / 512Mi              |
| restic backup jobs  | 200m / 1            | 512Mi / 1Gi            | 1Gi / 2Gi                 |

### Node Storage

`local-path` is the default storage class and provisions directories under `/var/lib/rancher/k3s/storage` on the node's root filesystem. It does **not** enforce the size in a PVC's request, so a claim can grow until the disk is full, taint the node with `node.kubernetes.io/disk-pressure`, and evict unrelated pods.

Jellyfin's media is the exception: it is backed by an NFS `PersistentVolume` exported from TrueNAS, so it does not consume root disk.

The Ubuntu installer allocates only part of the volume group to `ubuntu-lv`, leaving a new VM with a root filesystem about half its disk size. The setup playbook corrects this automatically, and also enables `fstrim.timer` so blocks freed in the guest are returned to the ZFS pool on a schedule rather than staying charged against it. See [Node Disk Sizing](../../foundry_deployment/README.md#node-disk-sizing) for growing the disk and recovering from disk pressure.

### Cluster Setup

The cluster is provisioned via Ansible using `make setup-cluster`. The playbook (`setup_cluster.yaml`) performs the following:

1. Grows the root partition, physical volume and logical volume to fill the disk, and enables `fstrim.timer`.
2. Configures K3s API server authentication with Dex.
3. Installs cert-manager for TLS certificate management.
4. Adds Helm repos (ArgoCD, Prometheus, Dex) and installs their charts.
5. Applies Kubernetes manifests for auth, foundry, traefik, jellyfin, monitoring, Pi-hole, Stalwart, and ArgoCD resources.
