import {createRouter, createWebHistory} from 'vue-router';
import {useAuthStore} from '../stores/auth';
import LoginView from '../views/LoginView.vue';

const router = createRouter({
    history: createWebHistory(),
    routes: [
        {path: '/login', name: 'Login', component: LoginView, meta: {public: true}},
        {path: '/pending', name: 'Pending', component: () => import('../views/PendingView.vue')},
        {path: '/', redirect: '/services'},
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
        {
            path: '/admin/clients',
            name: 'Clients',
            component: () => import('../views/ClientsView.vue'),
            meta: {admin: true},
        },
    ],
});

router.beforeEach(async (to) => {
    const auth = useAuthStore();
    if (auth.loading) await auth.checkAuth();
    if (!to.meta.public && !auth.isAuthenticated) return '/login';
    if (auth.isPending) return to.path === '/pending' ? true : '/pending';
    if (to.path === '/pending') return '/services';
    if (to.meta.admin && !auth.isAdmin) return '/services';
    if (to.path === '/login' && auth.isAuthenticated) return '/services';
});

export default router;