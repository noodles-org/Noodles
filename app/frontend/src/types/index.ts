export type Role = 'admin' | 'viewer' | 'client_admin' | 'client' | 'pending';

export interface User {
    sub: string;
    email: string;
    name: string;
    role: Role;
    groups: string[];
}

export interface ClientEntry {
    email: string;
    name: string;
    sub: string;
    role: Role;
    createdAt: string;
}

export interface DeploymentInfo {
    name: string;
    namespace: string;
    cluster: string;
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

export interface FileEntry {
    name: string;
    path: string;
    isDir: boolean;
    size: number;
    modTime: string;
}

export interface DocTocSection {
    title: string;
    items: { title: string; path: string }[];
}