import {resolve} from 'path';
import {randomBytes} from 'crypto';

function required(key: string): string {
    const val = process.env[key];
    if (!val) throw new Error(`Missing required env: ${key}`);
    return val;
}

function optional(key: string, fallback: string): string {
    return process.env[key] || fallback;
}

const jwtSecret = process.env.JWT_SECRET || randomBytes(32).toString('hex');
if (!process.env.JWT_SECRET) {
    console.warn('JWT_SECRET not set — generated ephemeral secret (sessions won\'t survive restarts)');
}

export const config = {
    port: parseInt(optional('PORT', '3000')),
    metricsPort: parseInt(optional('METRICS_PORT', '9090')),
    nodeEnv: optional('NODE_ENV', 'development'),
    isProduction: optional('NODE_ENV', 'development') === 'production',

    // TODO: Update all OAuth URLs to your Dex instance
    oauth: {
        clientId: required('OAUTH_CLIENT_ID'),
        clientSecret: required('OAUTH_CLIENT_SECRET'),
        authorizeUrl: required('OAUTH_AUTHORIZE_URL'),
        tokenUrl: required('OAUTH_TOKEN_URL'),
        userinfoUrl: required('OAUTH_USERINFO_URL'),
        callbackUrl: required('OAUTH_CALLBACK_URL'),
        scopes: optional('OAUTH_SCOPES', 'openid profile email groups'),
    },

    jwt: {
        secret: jwtSecret,
        expiresIn: optional('JWT_EXPIRES_IN', '8h'),
        cookieName: 'dashboard_token',
    },

    // TODO: Update GitHub org group names to match your org
    auth: {
        adminGroups: optional('AUTH_ADMIN_GROUPS', 'example-org:admin')
            .split(',').map((s) => s.trim()),
        allowedGroups: optional('AUTH_ALLOWED_GROUPS', 'example-org:admin,example-org:developer')
            .split(',').map((s) => s.trim()),
    },

    // TODO: Update if your ArgoCD lives at a different in-cluster address
    argocd: {
        url: optional('ARGOCD_URL', 'https://argocd-server.argocd.svc.cluster.local'),
        token: process.env.ARGOCD_TOKEN || '',
        insecure: optional('ARGOCD_INSECURE', 'true') === 'true',
    },

    namespaceLabel: optional('NAMESPACE_LABEL', 'dashboard.cluster/managed'),

    docsPath: resolve(optional('DOCS_PATH', '../docs')),
    configPath: resolve(optional('CONFIG_PATH', '../config')),
    frontendPath: resolve(optional('FRONTEND_PATH', '../frontend/dist')),
    corsOrigin: optional('CORS_ORIGIN', 'http://localhost:5173'),
};