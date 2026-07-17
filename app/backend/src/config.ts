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

    oauth: (() => {
        const issuer = 'https://dex.noodles.quest';
        const publicUrl = 'https://noodles.quest';
        return {
            clientId: 'noodles-dashboard',
            clientSecret: process.env.DASHBOARD_CLIENT_SECRET || '',
            authorizeUrl: `${issuer}/auth`,
            tokenUrl: `${issuer}/token`,
            userinfoUrl: `${issuer}/userinfo`,
            callbackUrl: `${publicUrl}/api/auth/callback`,
            scopes: optional('OAUTH_SCOPES', 'openid profile email groups'),
        };
    })(),

    jwt: {
        secret: jwtSecret,
        expiresIn: optional('JWT_EXPIRES_IN', '8h'),
        cookieName: 'dashboard_token',
    },

    auth: (() => {
        const adminGroup = 'noodles-org:admin';
        const viewerGroup = 'noodles-org:developer';
        return {
            adminGroups: [adminGroup],
            allowedGroups: [adminGroup, viewerGroup],
        };
    })(),

    argocd: {
        url: 'https://argocd-server.argocd.svc.cluster.local',
        token: process.env.ARGOCD_TOKEN || '',
        insecure: optional('ARGOCD_INSECURE', 'true') === 'true',
    },

    namespaceLabel: optional('NAMESPACE_LABEL', 'noodles.dashboard/managed'),

    docsPath: resolve(optional('DOCS_PATH', '../../docs')),
    frontendPath: resolve(optional('FRONTEND_PATH', '../../frontend/dist')),
    corsOrigin: optional('CORS_ORIGIN', 'http://localhost:5173'),
};