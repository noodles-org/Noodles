import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import LoginView from '../views/LoginView.vue';

const router = createRouter({
    history: createWebHistory(),
    routes: [
        { path: '/login', name: 'Login', component: LoginView, meta: { public: true } },
        { path: '/', redirect: '/services' },
        {
            path: '/services',
            name: 'Services',
            component: () => import('../views/ServicesView.vue'),
        },
        {
            path: '/deployments',
            name: 'Deployments',
            component: () => import('../views/DeploymentsView.vue'),
        },
        {
            path: '/docs',
            name: 'Docs',
            component: () => import('../views/DocsView.vue'),
        },
        // Add new pages here — just create a view and add a route entry
    ],
});

router.beforeEach(async (to) => {
    const auth = useAuthStore();
    if (auth.loading) await auth.checkAuth();
    if (!to.meta.public && !auth.isAuthenticated) return '/login';
    if (to.path === '/login' && auth.isAuthenticated) return '/services';
});

export default router;