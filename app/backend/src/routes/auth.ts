import {Router, Request, Response} from 'express';
import {randomBytes} from 'crypto';
import jwt from 'jsonwebtoken';
import axios from 'axios';
import {config} from '../config';
import {AuthenticatedRequest, User, Role} from '../types';
import {DEV_USER} from '../constants';
import {requireAuth} from '../middleware/auth';
import {logger} from '../services/logger';
import {authEvents, trackUniqueUser} from '../services/metrics';

const router = Router();

/** Map Dex group claims to dashboard role */
function resolveRole(groups: string[]): Role {
    if (config.auth.adminGroups.some((g) => groups.includes(g))) return 'admin';
    return 'viewer';
}

function isAllowedUser(groups: string[]): boolean {
    return config.auth.allowedGroups.some((g) => groups.includes(g));
}

router.get('/login', (req: Request, res: Response) => {
    // In development, bypass Dex and issue a token for a mock user
    if (!config.isProduction) {
        const token = jwt.sign(DEV_USER, config.jwt.secret);
        res.cookie(config.jwt.cookieName, token, {
            httpOnly: true,
            secure: false,
            sameSite: 'lax',
            maxAge: 8 * 60 * 60 * 1000,
            path: '/',
        });

        logger.info('Auth: dev login bypass', {ip: req.ip});
        return res.redirect('/');
    }

    const state = randomBytes(32).toString('hex');
    res.cookie('oauth_state', state, {
        httpOnly: true,
        secure: config.isProduction,
        sameSite: 'lax',
        maxAge: 600_000,
        path: '/api/auth',
    });

    logger.info('Auth: login initiated', {ip: req.ip});

    const params = new URLSearchParams({
        client_id: config.oauth.clientId,
        redirect_uri: config.oauth.callbackUrl,
        response_type: 'code',
        scope: config.oauth.scopes,
        state,
    });

    res.redirect(`${config.oauth.authorizeUrl}?${params}`);
});

router.get('/callback', async (req: Request, res: Response) => {
    const {code, state, error: oauthError} = req.query;
    const storedState = req.cookies?.oauth_state;

    res.clearCookie('oauth_state', {path: '/api/auth'});

    if (oauthError) {
        logger.warn('Auth: provider error', {error: oauthError, ip: req.ip});
        authEvents.inc({status: 'failure', reason: 'oauth_error'});
        return res.redirect('/login?error=oauth_denied');
    }

    if (!state || !storedState || state !== storedState) {
        logger.warn('Auth: state mismatch', {ip: req.ip});
        authEvents.inc({status: 'failure', reason: 'invalid_state'});
        return res.redirect('/login?error=invalid_state');
    }

    if (!code) {
        logger.warn('Auth: no code returned', {ip: req.ip});
        authEvents.inc({status: 'failure', reason: 'no_code'});
        return res.redirect('/login?error=no_code');
    }

    try {
        // Exchange code for token with Dex
        const tokenRes = await axios.post(
            config.oauth.tokenUrl,
            new URLSearchParams({
                grant_type: 'authorization_code',
                client_id: config.oauth.clientId,
                client_secret: config.oauth.clientSecret,
                code: code as string,
                redirect_uri: config.oauth.callbackUrl,
            }).toString(),
            {headers: {'Content-Type': 'application/x-www-form-urlencoded'}},
        );

        // Dex userinfo returns standard OIDC claims + groups from GitHub org
        const userRes = await axios.get(config.oauth.userinfoUrl, {
            headers: {Authorization: `Bearer ${tokenRes.data.access_token}`},
        });

        const ui = userRes.data;
        const groups: string[] = ui.groups || [];

        if (!isAllowedUser(groups)) {
            logger.warn('Auth: user not in allowed groups', {
                email: ui.email,
                groups,
                ip: req.ip,
            });
            authEvents.inc({status: 'failure', reason: 'unauthorized_group'});
            return res.redirect('/login?error=not_authorized');
        }

        const role = resolveRole(groups);

        const user: User = {
            sub: ui.sub || ui.id,
            email: ui.email,
            name: ui.name || ui.preferred_username || ui.email,
            groups,
            role,
        };

        const token = jwt.sign(user, config.jwt.secret);

        res.cookie(config.jwt.cookieName, token, {
            httpOnly: true,
            secure: config.isProduction,
            sameSite: 'lax',
            maxAge: 8 * 60 * 60 * 1000,
            path: '/',
        });

        trackUniqueUser(user.sub);

        logger.info('Auth: login success', {
            user: user.email,
            role,
            groups,
            ip: req.ip,
        });
        authEvents.inc({status: 'success', reason: 'login'});

        res.redirect('/');
    } catch (err) {
        logger.error('Auth: token exchange failed', {
            error: (err as Error).message,
            ip: req.ip,
        });
        authEvents.inc({status: 'failure', reason: 'token_exchange'});
        res.redirect('/login?error=auth_failed');
    }
});

router.get('/me', requireAuth, (req: AuthenticatedRequest, res: Response) => {
    res.json(req.user);
});

router.post('/logout', requireAuth, (req: AuthenticatedRequest, res: Response) => {
    logger.info('Auth: logout', {user: req.user?.email, ip: req.ip});
    authEvents.inc({status: 'success', reason: 'logout'});
    res.clearCookie(config.jwt.cookieName, {path: '/'});
    res.json({ok: true});
});

export default router;