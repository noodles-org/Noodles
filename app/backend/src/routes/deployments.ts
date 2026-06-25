import { Router, Response } from 'express';
import { AuthenticatedRequest } from '../types';
import { requireAuth } from '../middleware/auth';
import { config } from '../config';
import * as k8s from '../services/kubernetes';
import { getArgoHealthMap } from '../services/argocd';
import { logger } from '../services/logger';
import { deploymentActions } from '../services/metrics';

const router = Router();
const NAME_RE = /^[a-z0-9]([a-z0-9\-.]*[a-z0-9])?$/;

function validate(res: Response, namespace: string, name: string): boolean {
    if (!NAME_RE.test(namespace) || !NAME_RE.test(name)) {
        res.status(400).json({ error: 'Invalid namespace or deployment name' });
        return false;
    }
    if (!config.managedNamespaces.includes(namespace)) {
        res.status(403).json({ error: 'Namespace not managed by this dashboard' });
        return false;
    }
    return true;
}

router.get('/', requireAuth, async (_req: AuthenticatedRequest, res: Response) => {
    try {
        const [deployments, argoMap] = await Promise.all([
            k8s.listDeployments(config.managedNamespaces),
            getArgoHealthMap(),
        ]);

        for (const dep of deployments) {
            if (dep.argoApp && argoMap.has(dep.argoApp)) {
                const a = argoMap.get(dep.argoApp)!;
                dep.healthStatus = a.health;
                dep.syncStatus = a.sync;
            }
        }

        res.json(deployments);
    } catch (err) {
        logger.error('Failed listing deployments', { error: err });
        res.status(500).json({ error: 'Failed to list deployments' });
    }
});

router.post('/:namespace/:name/restart', requireAuth, async (req: AuthenticatedRequest, res: Response) => {
    const { namespace, name } = req.params;
    if (!validate(res, namespace, name)) return;

    try {
        await k8s.restartDeployment(namespace, name);
        deploymentActions.inc({ action: 'restart', namespace, deployment: name });
        logger.info('Restart requested', { namespace, name, user: req.user?.email });
        res.json({ ok: true });
    } catch (err) {
        logger.error('Restart failed', { namespace, name, error: err });
        res.status(500).json({ error: 'Restart failed' });
    }
});

router.post('/:namespace/:name/pause', requireAuth, async (req: AuthenticatedRequest, res: Response) => {
    const { namespace, name } = req.params;
    if (!validate(res, namespace, name)) return;

    try {
        await k8s.pauseDeployment(namespace, name);
        deploymentActions.inc({ action: 'pause', namespace, deployment: name });
        logger.info('Pause requested', { namespace, name, user: req.user?.email });
        res.json({ ok: true });
    } catch (err) {
        logger.error('Pause failed', { namespace, name, error: err });
        res.status(500).json({ error: (err as Error).message });
    }
});

router.post('/:namespace/:name/resume', requireAuth, async (req: AuthenticatedRequest, res: Response) => {
    const { namespace, name } = req.params;
    if (!validate(res, namespace, name)) return;

    try {
        await k8s.resumeDeployment(namespace, name);
        deploymentActions.inc({ action: 'resume', namespace, deployment: name });
        logger.info('Resume requested', { namespace, name, user: req.user?.email });
        res.json({ ok: true });
    } catch (err) {
        logger.error('Resume failed', { namespace, name, error: err });
        res.status(500).json({ error: (err as Error).message });
    }
});

export default router;