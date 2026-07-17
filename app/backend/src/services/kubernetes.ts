import * as k8s from '@kubernetes/client-node';
import {readFileSync} from 'fs';
import {resolve} from 'path';
import {DeploymentInfo, ServiceLink} from '../types';
import {config} from '../config';
import {logger} from './logger';

const isDev = !config.isProduction;

const kc = new k8s.KubeConfig();
if (!isDev) {
    try {
        kc.loadFromCluster();
    } catch {
        logger.warn('K8s: failed to load in-cluster config');
    }
}

const MOCKS_DIR = resolve(__dirname, '..', '..', 'mocks');

function loadMock<T>(file: string): T {
    return JSON.parse(readFileSync(resolve(MOCKS_DIR, file), 'utf-8'));
}

const appsApi = isDev ? null : kc.makeApiClient(k8s.AppsV1Api);
const coreApi = isDev ? null : kc.makeApiClient(k8s.CoreV1Api);
const customApi = isDev ? null : kc.makeApiClient(k8s.CustomObjectsApi);

const PAUSE_ANNOTATION = 'noodles.dashboard/original-replicas';
const RESTART_ANNOTATION = 'kubectl.kubernetes.io/restartedAt';
const ARGO_LABEL = 'argocd.argoproj.io/instance';
const PATCH_HEADERS = {headers: {'Content-Type': 'application/strategic-merge-patch+json'}};

let nsCache: string[] | null = null;
let nsCacheTime = 0;
const NS_CACHE_TTL = 60_000;

export async function discoverNamespaces(): Promise<string[]> {
    if (isDev) return ['foundry', 'jellyfin', 'stalwart'];
    if (nsCache && Date.now() - nsCacheTime < NS_CACHE_TTL) return nsCache;

    try {
        const {body} = await coreApi!.listNamespace(
            undefined, undefined, undefined, undefined,
            `${config.namespaceLabel}=true`,
        );
        nsCache = body.items
            .map((ns) => ns.metadata?.name)
            .filter(Boolean) as string[];
        nsCacheTime = Date.now();

        if (!nsCache.length) {
            logger.warn(`No namespaces found with label ${config.namespaceLabel}=true`);
        }
    } catch (err) {
        logger.error('Failed to discover namespaces', {error: err});
        nsCache = nsCache ?? [];
    }

    return nsCache;
}

export async function isManagedNamespace(namespace: string): Promise<boolean> {
    const managed = await discoverNamespaces();
    return managed.includes(namespace);
}

// ── Deployment helpers ────────────────────────────────────────────

function deriveHealth(dep: k8s.V1Deployment): string {
    const desired = dep.spec?.replicas ?? 0;
    const available = dep.status?.availableReplicas ?? 0;
    const ready = dep.status?.readyReplicas ?? 0;
    if (desired === 0) return 'Suspended';
    if (ready === desired && available === desired) return 'Healthy';
    if (ready > 0) return 'Progressing';
    return 'Degraded';
}

export async function listDeployments(): Promise<DeploymentInfo[]> {
    if (isDev) return loadMock<DeploymentInfo[]>('deployments.json');
    const namespaces = await discoverNamespaces();
    const all: DeploymentInfo[] = [];

    for (const ns of namespaces) {
        try {
            const {body} = await appsApi!.listNamespacedDeployment(ns);
            for (const dep of body.items) {
                const ann = dep.metadata?.annotations ?? {};
                const savedReplicas = ann[PAUSE_ANNOTATION]
                    ? parseInt(ann[PAUSE_ANNOTATION])
                    : undefined;

                all.push({
                    name: dep.metadata?.name ?? '',
                    namespace: dep.metadata?.namespace ?? ns,
                    replicas: dep.spec?.replicas ?? 0,
                    readyReplicas: dep.status?.readyReplicas ?? 0,
                    availableReplicas: dep.status?.availableReplicas ?? 0,
                    image: dep.spec?.template?.spec?.containers?.[0]?.image ?? 'unknown',
                    paused: dep.spec?.replicas === 0 && !!savedReplicas,
                    originalReplicas: savedReplicas,
                    argoApp: dep.metadata?.labels?.[ARGO_LABEL],
                    healthStatus: deriveHealth(dep),
                    syncStatus: 'Unknown',
                    lastRestartedAt:
                        dep.spec?.template?.metadata?.annotations?.[RESTART_ANNOTATION],
                    createdAt: dep.metadata?.creationTimestamp?.toISOString(),
                });
            }
        } catch (err) {
            logger.error(`Failed listing deployments in ${ns}`, {error: err});
        }
    }

    return all;
}

export async function restartDeployment(namespace: string, name: string): Promise<void> {
    if (isDev) throw new Error('K8s not available in dev mode');
    await appsApi!.patchNamespacedDeployment(
        name, namespace,
        {
            spec: {
                template: {
                    metadata: {
                        annotations: {
                            [RESTART_ANNOTATION]: new Date().toISOString(),
                        }
                    }
                }
            }
        },
        undefined, undefined, undefined, undefined, undefined,
        PATCH_HEADERS,
    );
    logger.info('Deployment restarted', {namespace, name});
}

export async function pauseDeployment(namespace: string, name: string): Promise<void> {
    if (isDev) throw new Error('K8s not available in dev mode');
    const {body: dep} = await appsApi!.readNamespacedDeployment(name, namespace);
    const current = dep.spec?.replicas ?? 1;
    if (current === 0) throw new Error('Already paused');

    await appsApi!.patchNamespacedDeployment(
        name, namespace,
        {
            metadata: {annotations: {[PAUSE_ANNOTATION]: String(current)}},
            spec: {replicas: 0},
        },
        undefined, undefined, undefined, undefined, undefined,
        PATCH_HEADERS,
    );
    logger.info('Deployment paused', {namespace, name, previousReplicas: current});
}

export async function resumeDeployment(namespace: string, name: string): Promise<void> {
    if (isDev) throw new Error('K8s not available in dev mode');
    const {body: dep} = await appsApi!.readNamespacedDeployment(name, namespace);
    const target = parseInt(dep.metadata?.annotations?.[PAUSE_ANNOTATION] ?? '1');
    if ((dep.spec?.replicas ?? 0) > 0) throw new Error('Not currently paused');

    await appsApi!.patchNamespacedDeployment(
        name, namespace,
        {spec: {replicas: target}},
        undefined, undefined, undefined, undefined, undefined,
        PATCH_HEADERS,
    );
    logger.info('Deployment resumed', {namespace, name, replicas: target});
}

// ── Service discovery via IngressRoute labels ─────────────────────

const SERVICE_LABEL = 'noodles.dashboard/service';
const ANN_PREFIX = 'noodles.dashboard/';

function deriveUrlFromRoutes(spec: any, hasTls: boolean): string | undefined {
    const routes: any[] = spec?.routes ?? [];
    for (const route of routes) {
        const match = route.match as string | undefined;
        if (!match) continue;
        const hostMatch = match.match(/Host\(`([^`]+)`\)/);
        if (!hostMatch) continue;
        const host = hostMatch[1];
        const pathMatch = match.match(/PathPrefix\(`([^`]+)`\)/);
        const scheme = hasTls ? 'https' : 'http';
        return pathMatch ? `${scheme}://${host}${pathMatch[1]}` : `${scheme}://${host}`;
    }
    return undefined;
}

export async function discoverServices(): Promise<ServiceLink[]> {
    if (isDev) return loadMock<ServiceLink[]>('services.json');
    const services: ServiceLink[] = [];

    try {
        const {body} = await customApi!.listClusterCustomObject(
            'traefik.io', 'v1alpha1', 'ingressroutes',
            undefined, undefined, undefined, undefined,
            `${SERVICE_LABEL}=true`,
        ) as { body: { items: any[] } };

        for (const item of body.items ?? []) {
            const ann = item.metadata?.annotations ?? {};
            const name = ann[`${ANN_PREFIX}name`];
            if (!name) continue;

            const hasTls = !!item.spec?.tls;
            const url = ann[`${ANN_PREFIX}url`] || deriveUrlFromRoutes(item.spec, hasTls);

            services.push({
                name,
                url: url || '',
                description: ann[`${ANN_PREFIX}description`] || '',
                category: ann[`${ANN_PREFIX}category`] || 'Other',
            });
        }
    } catch (err) {
        logger.error('Failed to discover services from IngressRoutes', {error: err});
    }

    return services;
}