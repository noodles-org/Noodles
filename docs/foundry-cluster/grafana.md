# Grafana

Grafana is the primary observability dashboard, deployed as part of the `kube-prometheus-stack` Helm chart.

## Access

- **URL:** `https://noodles.quest/grafana/`
- **Authentication:** GitHub OAuth via the `noodles-org` GitHub organization.
- Grafana is served from the `/grafana` subpath.

## Role Mapping

GitHub organization teams are mapped to Grafana roles:

| GitHub Team               | Grafana Role |
|---------------------------|--------------|
| `noodles-org:admin`       | Admin        |
| `noodles-org:developer`   | Editor       |
| All others                | None         |

## Dashboards

- Default Kubernetes dashboards are enabled (`defaultDashboardsEnabled: true`).
- Custom dashboards are loaded via a sidecar that watches for ConfigMaps with the `grafana_folder` annotation.
- Dashboard ConfigMaps are stored in `k8s/monitoring/dashboards/`.

## Alerts

- Alert rules are loaded via a sidecar that watches for ConfigMaps.
- Alert ConfigMaps are stored in `k8s/monitoring/alerts/` (e.g., `fetch-ip-alert-cm.yaml`).

## Prometheus

Prometheus is deployed alongside Grafana in the `monitoring` namespace:

- Scrapes standard Kubernetes metrics and custom jobs.
- **Pushgateway** is enabled at port 9091 for push-based metrics (used by the fetch-ip CronJob).
- Pushgateway has a `ServiceMonitor` with `honorLabels: true`.
- Prometheus metrics are exposed internally at `noodles.local/metrics`.

## Loki

Grafana Loki provides log aggregation:

- Exposed internally at `loki.noodles.quest/loki/api/v1`.
- Grafana Alloy is deployed as the log collection agent, shipping logs to Loki.

## Networking

Monitoring services are exposed via Traefik IngressRoutes in the `monitoring` namespace:

| Service    | URL                                        | Entrypoints       | TLS |
|------------|--------------------------------------------|--------------------|-----|
| Grafana    | `https://noodles.quest/grafana/`           | web, websecure     | Yes |
| Loki       | `http://loki.noodles.quest/loki/api/v1`   | web                | No  |
| Prometheus | `http://noodles.local/metrics`             | web                | No  |
