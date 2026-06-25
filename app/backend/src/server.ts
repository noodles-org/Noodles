import './env';

import express from 'express';
import helmet from 'helmet';
import cors from 'cors';
import cookieParser from 'cookie-parser';
import rateLimit from 'express-rate-limit';
import path from 'path';

import { config } from './config';
import { logger } from './services/logger';
import { register } from './services/metrics';
import { requestLogger } from './middleware/logging';

import authRoutes from './routes/auth';
import deploymentRoutes from './routes/deployments';
import docsRoutes from './routes/docs';
import serviceRoutes from './routes/services';

const app = express();

// ── Security ──────────────────────────────────────────────────────
app.set('trust proxy', 1);
app.use(helmet({
    contentSecurityPolicy: {
        directives: {
            defaultSrc: ["'self'"],
            scriptSrc: ["'self'"],
            styleSrc: ["'self'", "'unsafe-inline'"],
            imgSrc: ["'self'", 'data:'],
        },
    },
}));
app.use(cors({
    origin: config.isProduction ? false : config.corsOrigin,
    credentials: true,
}));
app.use(cookieParser());
app.use(express.json());
app.use(requestLogger);

// ── Health check (unauthenticated) ────────────────────────────────
app.get('/healthz', (_req, res) => res.json({ status: 'ok' }));

// ── Auth rate limiter ─────────────────────────────────────────────
const authLimiter = rateLimit({
    windowMs: 15 * 60 * 1000,
    max: 50,
    standardHeaders: true,
    legacyHeaders: false,
    message: { error: 'Too many auth attempts' },
});

// ── API routes ────────────────────────────────────────────────────
app.use('/api/auth', authLimiter, authRoutes);
app.use('/api/deployments', deploymentRoutes);
app.use('/api/docs', docsRoutes);
app.use('/api/services', serviceRoutes);

// ── API 404 ───────────────────────────────────────────────────────
app.all('/api/*', (_req, res) => res.status(404).json({ error: 'Not found' }));

// ── Serve frontend in production ──────────────────────────────────
if (config.isProduction) {
    app.use(express.static(config.frontendPath));
    app.get('*', (_req, res) => {
        res.sendFile(path.join(config.frontendPath, 'index.html'));
    });
}

// ── Start ─────────────────────────────────────────────────────────
app.listen(config.port, () => {
    logger.info(`Server listening on :${config.port}`);
});

// Metrics on a separate port — not exposed through ingress
const metricsApp = express();
metricsApp.get('/metrics', async (_req, res) => {
    res.set('Content-Type', register.contentType);
    res.end(await register.metrics());
});
metricsApp.listen(config.metricsPort, () => {
    logger.info(`Metrics server on :${config.metricsPort}`);
});