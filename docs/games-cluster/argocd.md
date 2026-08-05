# ArgoCD

The games cluster is managed remotely by the ArgoCD instance running on the foundry cluster. A dedicated service account and RBAC resources are deployed on the games cluster to grant ArgoCD the necessary permissions.

## Remote Management

ArgoCD on the foundry cluster connects to the games cluster via the `games-context` cluster destination. The `games-project` in ArgoCD defines which namespaces and resources can be deployed. See the [foundry ArgoCD docs](../foundry-cluster/argocd.md#games-project) for project and ApplicationSet details.

## RBAC

The following resources are deployed on the games cluster:

| Resource             | Name                  | Namespace     | Purpose                                  |
|----------------------|-----------------------|---------------|------------------------------------------|
| `ServiceAccount`     | `argocd-manager`      | `kube-system` | Identity for ArgoCD to authenticate      |
| `ClusterRole`        | `argocd-manager-role` | —             | Full cluster-wide permissions            |
| `ClusterRoleBinding` | `argocd-manager-role-binding` | —     | Binds the role to the service account    |

These resources allow ArgoCD to manage all namespaces and resource types on the games cluster.
