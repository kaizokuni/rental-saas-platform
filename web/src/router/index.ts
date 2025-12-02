import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../stores/auth';
import LoginView from '../views/LoginView.vue';
import FleetView from '../views/FleetView.vue';

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/login',
            name: 'login',
            component: LoginView,
        },
        {
            path: '/fleet',
            name: 'fleet',
            component: FleetView,
            meta: { requiresAuth: true },
        },
        {
            path: '/',
            redirect: '/fleet',
        },
        {
            path: '/dashboard',
            name: 'dashboard',
            component: () => import('../views/DashboardView.vue'),
            meta: { requiresAuth: true },
        },
        {
            path: '/settings',
            name: 'settings',
            component: () => import('../views/SettingsView.vue'),
            meta: { requiresAuth: true },
        },
    ],
});

router.beforeEach((to, from, next) => {
    const authStore = useAuthStore();

    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
        next('/login');
    } else if (to.path === '/login' && authStore.isAuthenticated) {
        next('/fleet');
    } else {
        next();
    }
});

export default router;
