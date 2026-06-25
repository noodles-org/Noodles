import { ref } from 'vue';
import { defineStore } from 'pinia';
import api from '../api/client';
import type { DeploymentInfo } from '../types';

export const useDeploymentsStore = defineStore('deployments', () => {
    const deployments = ref<DeploymentInfo[]>([]);
    const loading = ref(false);
    const error = ref<string | null>(null);

    async function fetchDeployments() {
        loading.value = true;
        error.value = null;
        try {
            const { data } = await api.get('/deployments');
            deployments.value = data;
        } catch {
            error.value = 'Failed to load deployments';
        } finally {
            loading.value = false;
        }
    }

    async function restart(ns: string, name: string) {
        await api.post(`/deployments/${ns}/${name}/restart`);
        await fetchDeployments();
    }

    async function pause(ns: string, name: string) {
        await api.post(`/deployments/${ns}/${name}/pause`);
        await fetchDeployments();
    }

    async function resume(ns: string, name: string) {
        await api.post(`/deployments/${ns}/${name}/resume`);
        await fetchDeployments();
    }

    return { deployments, loading, error, fetchDeployments, restart, pause, resume };
});