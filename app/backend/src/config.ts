import { resolve } from 'path';
import { randomBytes } from 'crypto';

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

    oauth: {
        clientId: required('OAUTH_CLIENT_ID'),
        clientSecret: required('OAUTH_CLIENT_SECRET'),
        authorizeUrl: required('OAUTH_AUTHORIZE_URL'),
        tokenUrl: required('OAUTH_TOKEN_URL'),
        userinfoUrl: required('OAUTH_USERINFO_URL'),
        callbackUrl: required('OAUTH_CALLBACK_URL'),
        scopes: optional('OAUTH_SCOPES', 'openid profile email'),
    },

    jwt: {
        secret: jwtSecret,
        expiresIn: optional('JWT_EXPIRES_IN', '8h'),
        cookieName: 'dashboard_token',
    },

    argocd: {
        url: optional('ARGOCD_URL', 'https://argocd-server.argocd.svc.cluster.local'),
        token: process.env.ARGOCD_TOKEN || '',
        insecure: optional('ARGOCD_INSECURE', 'true') === 'true',
    },

    managedNamespaces: optional('MANAGED_NAMESPACES', 'default')
        .split(',')
        .map((s) => s.trim()),

    docsPath: resolve(optional('DOCS_PATH', '../docs')),
    configPath: resolve(optional('CONFIG_PATH', '../config')),
    frontendPath: resolve(optional('FRONTEND_PATH', '../frontend/dist')),
    corsOrigin: optional('CORS_ORIGIN', 'http://localhost:5173'),
} as const;