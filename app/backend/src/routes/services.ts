import {Router, Response} from 'express';
import {requireAuth} from '../middleware/auth';
import {AuthenticatedRequest} from '../types';
import {discoverServices} from '../services/kubernetes';
import {logger} from '../services/logger';

const router = Router();

router.get('/', requireAuth, async (_req: AuthenticatedRequest, res: Response) => {
    try {
        const services = await discoverServices();
        res.json(services);
    } catch (err) {
        logger.error('Service discovery failed', {error: err});
        res.status(500).json({error: 'Failed to discover services'});
    }
});

export default router;
