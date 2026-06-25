import client from 'prom-client';

client.collectDefaultMetrics({ prefix: 'dashboard_' });

export const authEvents = new client.Counter({
    name: 'dashboard_auth_events_total',
    help: 'Total authentication events',
    labelNames: ['status', 'reason'] as const,
});

export const activeUsers = new client.Gauge({
    name: 'dashboard_active_users',
    help: 'Approximate active user count (resets on restart)',
});

export const deploymentActions = new client.Counter({
    name: 'dashboard_deployment_actions_total',
    help: 'Deployment management actions',
    labelNames: ['action', 'namespace', 'deployment'] as const,
});

export const httpRequests = new client.Counter({
    name: 'dashboard_http_requests_total',
    help: 'Total HTTP requests',
    labelNames: ['method', 'route', 'status_code'] as const,
});

export const httpDuration = new client.Histogram({
    name: 'dashboard_http_request_duration_seconds',
    help: 'HTTP request duration',
    labelNames: ['method', 'route'] as const,
    buckets: [0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5],
});

export const register = client.register;