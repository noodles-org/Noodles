import { Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';
import { config } from '../config';
import { AuthenticatedRequest, User } from '../types';
import { logger } from '../services/logger';
import { authEvents } from '../services/metrics';

export function requireAuth(req: AuthenticatedRequest, res: Response, next: NextFunction): void {
    const token = req.cookies?.[config.jwt.cookieName];

    if (!token) {
        authEvents.inc({ status: 'failure', reason: 'no_token' });
        logger.warn('Auth: missing token', { ip: req.ip, path: req.path });
        res.status(401).json({ error: 'Authentication required' });
        return;
    }

    try {
        const decoded = jwt.verify(token, config.jwt.secret) as User;
        req.user = decoded;
        next();
    } catch (err) {
        authEvents.inc({ status: 'failure', reason: 'invalid_token' });
        logger.warn('Auth: invalid token', { ip: req.ip, error: (err as Error).message });
        res.status(401).json({ error: 'Invalid or expired token' });
    }
}