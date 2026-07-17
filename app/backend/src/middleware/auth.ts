import {NextFunction, Response} from 'express';
import jwt from 'jsonwebtoken';
import {config} from '../config';
import {AuthenticatedRequest, Role, User} from '../types';
import {DEV_USER} from '../constants';
import {logger} from '../services/logger';
import {authEvents} from '../services/metrics';

export function requireAuth(
    req: AuthenticatedRequest,
    res: Response,
    next: NextFunction,
): void {
    if (!config.isProduction) {
        req.user = DEV_USER;
        return next();
    }

    const token = req.cookies?.[config.jwt.cookieName];

    if (!token) {
        authEvents.inc({status: 'failure', reason: 'no_token'});
        logger.warn('Auth: missing token', {ip: req.ip, path: req.path});
        res.status(401).json({error: 'Authentication required'});
        return;
    }

    try {
        req.user = jwt.verify(token, config.jwt.secret) as User;
        next();
    } catch (err) {
        authEvents.inc({status: 'failure', reason: 'invalid_token'});
        logger.warn('Auth: invalid token', {
            ip: req.ip,
            error: (err as Error).message,
        });
        res.status(401).json({error: 'Invalid or expired token'});
    }
}

export function requireRole(...roles: Role[]) {
    return (
        req: AuthenticatedRequest,
        res: Response,
        next: NextFunction,
    ): void => {
        if (!req.user || !roles.includes(req.user.role)) {
            logger.warn('Auth: insufficient role', {
                user: req.user?.email,
                role: req.user?.role,
                required: roles,
            });
            res.status(403).json({error: 'Insufficient permissions'});
            return;
        }
        next();
    };
}