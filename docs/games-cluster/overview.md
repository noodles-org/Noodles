# Games Cluster

The games cluster runs on a K3s single-node cluster hosted on a TrueNAS VM. It serves dedicated game servers for multiplayer sessions.

## Cluster Overview

| Component       | Tool / Version                  | Purpose                              |
|-----------------|---------------------------------|--------------------------------------|
| Kubernetes      | K3s                             | Lightweight Kubernetes distribution  |
| Auth            | Dex (on foundry cluster)        | OIDC provider for cluster auth       |
| GitOps          | ArgoCD (managed from foundry)   | Continuous deployment from Git       |
| Monitoring      | Grafana Alloy                   | Log collection and forwarding        |
| DNS             | Cloudflare                      | DNS management and record updates    |
| Secrets         | SOPS                            | Encrypted secrets in version control |
| Provisioning    | Ansible                         | Cluster setup and configuration      |

### Namespaces

- **satisfactory** — Satisfactory dedicated server.
- **enshrouded** — Enshrouded dedicated server.
- **valheim** — Valheim dedicated server.
- **soba** — Soba Discord bot for server IP lookups.
- **argocd** — ArgoCD service account and RBAC for remote management.
- **monitoring** — Grafana Alloy log collector.

### Cluster Setup

The cluster is provisioned via Ansible using `make setup-cluster`. The playbook (`setup_cluster.yaml`) performs the following:

1. Configures K3s API server authentication with Dex (shared with the foundry cluster).
2. Configures DNS to bypass systemd-resolved and sets inotify limits.
3. Grows the root partition, physical volume and logical volume to fill the disk, and enables `fstrim.timer` so freed blocks return to the ZFS pool.
4. Verifies the kubeconfig can authenticate before touching the cluster, failing early with an explanation if the API server rejects the credentials.
5. Installs cert-manager for TLS certificate management.
6. Adds the Grafana Helm repo and installs the Alloy chart.
7. Applies Kubernetes manifests for auth, ArgoCD, monitoring, Satisfactory, Enshrouded, Valheim, and Soba resources.

The playbooks depend on the `community.general` and `ansible.posix` collections, pinned in `infra/requirements.yml` and installed automatically by `make setup-cluster`.
