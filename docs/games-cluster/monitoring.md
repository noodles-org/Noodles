# Monitoring

The games cluster uses Grafana Alloy for log collection and forwarding to the monitoring stack on the foundry cluster.

## Grafana Alloy

- **Helm chart:** `grafana/alloy`
- **Namespace:** `monitoring`
- **Values:** `gameserver_deployment/infra/helm/grafana-alloy/values.yaml`

Alloy collects logs from all pods on the games cluster and forwards them to the Loki instance on the foundry cluster for centralized log aggregation and querying via Grafana.
