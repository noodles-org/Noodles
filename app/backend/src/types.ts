import {Request} from 'express';

export type Role = 'admin' | 'viewer';

export interface User {
    sub: string;
    email: string;
    name: string;
    role: Role;
    groups: string[];
}

export interface AuthenticatedRequest extends Request {
    user?: User;
}

export interface DeploymentInfo {
    name: string;
    namespace: string;
    replicas: number;
    readyReplicas: number;
    availableReplicas: number;
    image: string;
    paused: boolean;
    originalReplicas?: number;
    argoApp?: string;
    healthStatus: string;
    syncStatus: string;
    lastRestartedAt?: string;
    createdAt?: string;
}

export interface ServiceLink {
    name: string;
    url: string;
    description: string;
    category: string;
}

export interface DocTocSection {
    title: string;
    items: { title: string; path: string }[];
}