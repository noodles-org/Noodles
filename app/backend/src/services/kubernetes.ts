import * as k8s from '@kubernetes/client-node';
import { DeploymentInfo } from '../types';
import { logger } from './logger';

const kc = new k8s.KubeConfig();
try {
    kc.loadFromCluster();
} catch {
    kc.loadFromDefault();
    logger.info('K8s: using local kubeconfig');
}

const appsApi = kc.makeApiClient(k8s.AppsV1Api);

const PAUSE_ANNOTATION = 'dashboard.cluster/original-replicas';
const RESTART_ANNOTATION = 'kubectl.kubernetes.io/restartedAt';
const ARGO_LABEL = 'argocd.argoproj.io/instance';
const PATCH_HEADERS = { headers: { 'Content-Type': 'application/strategic-merge-patch+json' } };

function deriveHealth(dep: k8s.V1Deployment): string {
    const desired = dep.spec?.replicas ?? 0;
    const available = dep.status?.availableReplicas ?? 0;
    const ready = dep.status?.readyReplicas ?? 0;
    if (desired === 0) return 'Suspended';
    if (ready === desired && available === desired) return 'Healthy';
    if (ready > 0) return 'Progressing';
    return 'Degraded';
}

export async function listDeployments(namespaces: string[]): Promise<DeploymentInfo[]> {
    const all: DeploymentInfo[] = [];

    for (const ns of namespaces) {
        try {
            const { body } = await appsApi.listNamespacedDeployment(ns);
            for (const dep of body.items) {
                const ann = dep.metadata?.annotations ?? {};
                const savedReplicas = ann[PAUSE_ANNOTATION] ? parseInt(ann[PAUSE_ANNOTATION]) : undefined;

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
                    lastRestartedAt: dep.spec?.template?.metadata?.annotations?.[RESTART_ANNOTATION],
                    createdAt: dep.metadata?.creationTimestamp?.toISOString(),
                });
            }
        } catch (err) {
            logger.error(`Failed listing deployments in ${ns}`, { error: err });
        }
    }

    return all;
}

export async function restartDeployment(namespace: string, name: string): Promise<void> {
    await appsApi.patchNamespacedDeployment(
        name, namespace,
        { spec: { template: { metadata: { annotations: { [RESTART_ANNOTATION]: new Date().toISOString() } } } } },
        undefined, undefined, undefined, undefined, undefined,
        PATCH_HEADERS,
    );
    logger.info('Deployment restarted', { namespace, name });
}

export async function pauseDeployment(namespace: string, name: string): Promise<void> {
    const { body: dep } = await appsApi.readNamespacedDeployment(name, namespace);
    const current = dep.spec?.replicas ?? 1;
    if (current === 0) throw new Error('Already paused');

    await appsApi.patchNamespacedDeployment(
        name, namespace,
        {
            metadata: { annotations: { [PAUSE_ANNOTATION]: String(current) } },
            spec: { replicas: 0 },
        },
        undefined, undefined, undefined, undefined, undefined,
        PATCH_HEADERS,
    );
    logger.info('Deployment paused', { namespace, name, previousReplicas: current });
}

export async function resumeDeployment(namespace: string, name: string): Promise<void> {
    const { body: dep } = await appsApi.readNamespacedDeployment(name, namespace);
    const target = parseInt(dep.metadata?.annotations?.[PAUSE_ANNOTATION] ?? '1');
    if ((dep.spec?.replicas ?? 0) > 0) throw new Error('Not currently paused');

    await appsApi.patchNamespacedDeployment(
        name, namespace,
        { spec: { replicas: target } },
        undefined, undefined, undefined, undefined, undefined,
        PATCH_HEADERS,
    );
    logger.info('Deployment resumed', { namespace, name, replicas: target });
}