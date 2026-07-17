import axios, {AxiosInstance} from 'axios';
import https from 'https';
import {config} from '../config';
import {logger} from './logger';

let client: AxiosInstance | null = null;

function getClient(): AxiosInstance | null {
    if (!config.argocd.token) return null;
    if (!client) {
        client = axios.create({
            baseURL: `${config.argocd.url}/api/v1`,
            headers: {Authorization: `Bearer ${config.argocd.token}`},
            httpsAgent: config.argocd.insecure
                ? new https.Agent({rejectUnauthorized: false})
                : undefined,
            timeout: 10_000,
        });
    }
    return client;
}

export async function getArgoHealthMap(): Promise<Map<string, { health: string; sync: string }>> {
    const map = new Map<string, { health: string; sync: string }>();
    const api = getClient();
    if (!api) return map;

    try {
        const {data} = await api.get('/applications');
        for (const app of data.items ?? []) {
            map.set(app.metadata.name, {
                health: app.status?.health?.status ?? 'Unknown',
                sync: app.status?.sync?.status ?? 'Unknown',
            });
        }
    } catch (err) {
        logger.error('ArgoCD fetch failed', {error: (err as Error).message});
    }

    return map;
}