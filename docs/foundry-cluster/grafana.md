# Grafana

Grafana is the primary observability dashboard, deployed as part of the `kube-prometheus-stack` Helm chart.

## Access

- **URL:** `https://noodles.quest/grafana/`
- **Authentication:** Dex OIDC SSO via generic OAuth (Grafana is registered as the `grafana` Dex static client). Dex in turn authenticates users against the `noodles-org` GitHub organization.
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
- Dashboard ConfigMaps are stored in `k8s/monitoring/dashboards/`:
  - `noodles-dashboard-cm.yaml` — Application metrics dashboard (`noodles-dashboard.json`) for per-route traffic, latency, and auth event drill-downs.
  - `foundry-dashboard-cm.yaml` — Foundry VTT dashboard (`foundry-dashboard.json`, folder `Foundry`) covering performance (CPU/memory vs limits, network, PV usage), ingress/HTTP (request rate, latency percentiles, 5xx errors — Traefik queries use the `service=~"foundry-foundry.*"` label), and an improved Loki logs view with `container` (Foundry server vs `file-sidecar`) and `search` (regex) template variables. The `foundry` container option surfaces Foundry's own stdout/stderr (world loads, player logins, module errors). File-service action metrics intentionally live on the Noodles dashboard (`dashboard_file_actions_total`), not here.
  - `slo-core-apps-cm.yaml` — SLO / SLI dashboard for core applications and APIs (`slo-core-apps.json`), tracking availability, latency percentiles, error budget burn rates, and Dex/Foundry ingress health.
  - `slo-scheduled-jobs-cm.yaml` — SLO / SLI dashboard for scheduled jobs (`slo-scheduled-jobs.json`), tracking execution freshness, failure rates, processed operations, and backup completion.

## Alerts

- Alert rules and notification policies are loaded via a sidecar that watches for ConfigMaps with the `grafana_alert: "1"` label.
- Alert ConfigMaps are stored in `k8s/monitoring/alerts/`:
  - `slo-alerts-cm.yaml` — Unified alerts for core application SLOs (fast error burn rates, high latency, Dex/Foundry ingress 5xx errors, overdue backups, and Fetch IP failures/overdue execution).
  - `slo-notification-policies-cm.yaml` — Contact points (Discord `#alerts` webhook) and default notification routing policy (all alerts and recoveries routed to Discord).

## Prometheus

Prometheus is deployed alongside Grafana in the `monitoring` namespace:

- Scrapes standard Kubernetes metrics and custom jobs.
- **Pushgateway** is enabled at port 9091 for push-based metrics (used by the fetch-ip CronJob).
- Pushgateway has a `ServiceMonitor` with `honorLabels: true`.
- The Foundry file-service sidecar is scraped via the `service-monitor` `ServiceMonitor` in the `foundry` namespace (selects the `file-sidecar` Service's `file-service` port at `/metrics`), exposing `filesvc_actions_total` for ad-hoc queries (the Foundry dashboard no longer charts these; the Noodles dashboard tracks file activity via `dashboard_file_actions_total`).
- Prometheus metrics are exposed internally at `noodles.local/metrics`.

## Loki

Grafana Loki provides log aggregation:

- Exposed internally at `loki.noodles.quest/loki/api/v1`.
- Grafana Alloy is deployed as the log collection agent, shipping logs to Loki.
- Alloy pushes to the in-cluster Loki gateway Service (`http://loki-gateway.monitoring.svc.cluster.local/loki/api/v1/push`), **not** the public `loki.noodles.quest` ingress. The public ingress is protected by the `loki-basic-auth` Traefik middleware, and Alloy's `loki.write` block has no credentials — pushing there returns `401` and silently drops all logs. Writing to the internal gateway (the same endpoint the Grafana Loki datasource reads from) bypasses that auth.
- Alloy's config (`helm/grafana-alloy/values.yaml`) uses a `discovery.relabel` step to promote Kubernetes metadata into real log labels (`namespace`, `pod`, `container`, `node`, `app`). Without this, `loki.source.kubernetes` drops all `__meta_*` labels and logs cannot be filtered by namespace/container — which is required for the Foundry dashboard's log panels to return data.

## Networking

Monitoring services are exposed via Traefik IngressRoutes in the `monitoring` namespace:

| Service    | URL                                        | Entrypoints       | TLS |
|------------|--------------------------------------------|--------------------|-----|
| Grafana    | `https://noodles.quest/grafana/`           | web, websecure     | Yes |
| Loki       | `http://loki.noodles.quest/loki/api/v1`   | web                | No  |
| Prometheus | `http://noodles.local/metrics`             | web                | No  |
