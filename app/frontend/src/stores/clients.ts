import {ref} from 'vue';
import {defineStore} from 'pinia';
import api from '../api/client';
import type {ClientEntry, Role} from '../types';

export const useClientsStore = defineStore('clients', () => {
    const pending = ref<ClientEntry[]>([]);
    const approved = ref<ClientEntry[]>([]);
    const loading = ref(false);
    const error = ref<string | null>(null);

    async function fetchClients() {
        loading.value = true;
        error.value = null;
        try {
            const [pendingRes, approvedRes] = await Promise.all([
                api.get('/clients/pending'),
                api.get('/clients'),
            ]);
            pending.value = pendingRes.data ?? [];
            approved.value = approvedRes.data ?? [];
        } catch {
            error.value = 'Failed to load clients';
        } finally {
            loading.value = false;
        }
    }

    async function approve(email: string, role: Role) {
        await api.post('/clients/approve', {email, role});
        await fetchClients();
    }

    async function reject(email: string) {
        await api.post('/clients/reject', {email});
        await fetchClients();
    }

    return {pending, approved, loading, error, fetchClients, approve, reject};
});
