import {Router, Response} from 'express';
import {readFileSync, existsSync} from 'fs';
import path from 'path';
import {config} from '../config';
import {requireAuth} from '../middleware/auth';
import {AuthenticatedRequest, DocTocSection} from '../types';

const router = Router();

function parseTocMarkdown(md: string): { sections: DocTocSection[] } {
    const sections: DocTocSection[] = [];
    let current: DocTocSection | null = null;

    for (const line of md.split('\n')) {
        const sectionMatch = line.match(/^## (.+)/);
        if (sectionMatch) {
            current = {title: sectionMatch[1], items: []};
            sections.push(current);
            continue;
        }
        const itemMatch = line.match(/^- \[(.+?)]\((.+?)(?:#.+?)?\)/);
        if (itemMatch && current) {
            if (!current.items.some(i => i.path === itemMatch[2])) {
                current.items.push({title: itemMatch[1], path: itemMatch[2]});
            }
        }
    }

    return {sections};
}

let tocCache: { sections: DocTocSection[] } | null = null;

router.get('/toc', requireAuth, (_req: AuthenticatedRequest, res: Response) => {
    try {
        if (!tocCache) {
            const md = readFileSync(path.join(config.docsPath, 'toc.md'), 'utf-8');
            tocCache = parseTocMarkdown(md);
        }
        res.json(tocCache);
    } catch {
        res.status(500).json({error: 'Failed to load table of contents'});
    }
});

process.on('SIGHUP', () => {
    tocCache = null;
});

router.get('/content', requireAuth, (req: AuthenticatedRequest, res: Response) => {
    const docPath = req.query.path as string;
    if (!docPath) return res.status(400).json({error: 'Missing path parameter'});

    // Block directory traversal
    const normalized = path.normalize(docPath);
    if (normalized.includes('..') || path.isAbsolute(normalized)) {
        return res.status(400).json({error: 'Invalid path'});
    }

    const full = path.join(config.docsPath, normalized);
    if (!existsSync(full)) return res.status(404).json({error: 'Not found'});

    try {
        res.json({content: readFileSync(full, 'utf-8')});
    } catch {
        res.status(500).json({error: 'Failed to read document'});
    }
});

export default router;
