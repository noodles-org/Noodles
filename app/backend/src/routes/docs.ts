import { Router, Response } from 'express';
import { readFileSync, existsSync } from 'fs';
import path from 'path';
import { config } from '../config';
import { requireAuth } from '../middleware/auth';
import { AuthenticatedRequest } from '../types';

const router = Router();

router.get('/toc', requireAuth, (_req: AuthenticatedRequest, res: Response) => {
    try {
        const raw = readFileSync(path.join(config.configPath, 'toc.json'), 'utf-8');
        res.json(JSON.parse(raw));
    } catch {
        res.status(500).json({ error: 'Failed to load table of contents' });
    }
});

router.get('/content', requireAuth, (req: AuthenticatedRequest, res: Response) => {
    const docPath = req.query.path as string;
    if (!docPath) return res.status(400).json({ error: 'Missing path parameter' });

    // Block directory traversal
    const normalized = path.normalize(docPath);
    if (normalized.includes('..') || path.isAbsolute(normalized)) {
        return res.status(400).json({ error: 'Invalid path' });
    }

    const full = path.join(config.docsPath, normalized);
    if (!existsSync(full)) return res.status(404).json({ error: 'Not found' });

    try {
        res.json({ content: readFileSync(full, 'utf-8') });
    } catch {
        res.status(500).json({ error: 'Failed to read document' });
    }
});

export default router;