import { ref, computed } from 'vue';
import { defineStore } from 'pinia';
import api from '../api/client';
import type { User } from '../types';

export const useAuthStore = defineStore('auth', () => {
    const user = ref<User | null>(null);
    const loading = ref(true);

    const isAuthenticated = computed(() => !!user.value);

    async function checkAuth() {
        try {
            const { data } = await api.get('/auth/me');
            user.value = data;
        } catch {
            user.value = null;
        } finally {
            loading.value = false;
        }
    }

    function login() {
        window.location.href = '/api/auth/login';
    }

    async function logout() {
        try { await api.post('/auth/logout'); } catch { /* proceed */ }
        user.value = null;
        window.location.href = '/login';
    }

    return { user, loading, isAuthenticated, checkAuth, login, logout };
});