import './env';

import express from 'express';
import helmet from 'helmet';
import cors from 'cors';
import cookieParser from 'cookie-parser';
import rateLimit from 'express-rate-limit';
import path from 'path';

import {config} from './config';
import {logger} from './services/logger';
import {register} from './services/metrics';
import {requestLogger} from './middleware/logging';

import authRoutes from './routes/auth';
import deploymentRoutes from './routes/deployments';
import docsRoutes from './routes/docs';
import serviceRoutes from './routes/services';

const app = express();

app.set('trust proxy', 1);
app.use(
    helmet({
        contentSecurityPolicy: {
            directives: {
                defaultSrc: ["'self'"],
                scriptSrc: ["'self'"],
                styleSrc: ["'self'", "'unsafe-inline'"],
                imgSrc: ["'self'", 'data:'],
            },
        },
    }),
);
app.use(
    cors({
        origin: config.isProduction ? false : config.corsOrigin,
        credentials: true,
    }),
);
app.use(cookieParser());
app.use(express.json());
app.use(requestLogger);

app.get('/healthz', (_req, res) => res.json({status: 'ok'}));

const authLimiter = rateLimit({
    windowMs: 15 * 60 * 1000,
    max: 50,
    standardHeaders: true,
    legacyHeaders: false,
    message: {error: 'Too many auth attempts'},
});

const apiLimiter = rateLimit({
    windowMs: 15 * 60 * 1000,
    max: 200,
    standardHeaders: true,
    legacyHeaders: false,
    message: {error: 'Too many requests'},
});

app.use('/api/auth', authLimiter, authRoutes);
app.use('/api/deployments', apiLimiter, deploymentRoutes);
app.use('/api/docs', docsRoutes);
app.use('/api/services', serviceRoutes);

app.all('/api/*', (_req, res) => res.status(404).json({error: 'Not found'}));

if (config.isProduction) {
    app.use(express.static(config.frontendPath));
    app.get('*', (_req, res) => {
        res.sendFile(path.join(config.frontendPath, 'index.html'));
    });
}

app.listen(config.port, () => logger.info(`Server on :${config.port}`));

const metricsApp = express();
metricsApp.get('/metrics', async (_req, res) => {
    res.set('Content-Type', register.contentType);
    res.end(await register.metrics());
});
metricsApp.listen(config.metricsPort, () =>
    logger.info(`Metrics on :${config.metricsPort}`),
);