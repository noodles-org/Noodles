import { Request, Response, NextFunction } from 'express';
import { httpRequests, httpDuration } from '../services/metrics';

export function requestLogger(req: Request, res: Response, next: NextFunction): void {
    const start = Date.now();

    res.on('finish', () => {
        const duration = (Date.now() - start) / 1000;
        const route = req.route?.path || req.path;
        httpRequests.inc({ method: req.method, route, status_code: String(res.statusCode) });
        httpDuration.observe({ method: req.method, route }, duration);
    });

    next();
}