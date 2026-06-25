import { Router, Response } from 'express';
import { readFileSync } from 'fs';
import path from 'path';
import { config } from '../config';
import { requireAuth } from '../middleware/auth';
import { AuthenticatedRequest, ServiceLink } from '../types';

const router = Router();

let cache: ServiceLink[] | null = null;

function load(): ServiceLink[] {
    if (cache) return cache;
    try {
        cache = JSON.parse(readFileSync(path.join(config.configPath, 'services.json'), 'utf-8'));
        return cache!;
    } catch {
        return [];
    }
}

// Reload on SIGHUP so ConfigMap updates can be picked up without restart
process.on('SIGHUP', () => { cache = null; });

router.get('/', requireAuth, (_req: AuthenticatedRequest, res: Response) => {
    res.json(load());
});

export default router;