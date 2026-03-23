import { createRouter, createWebHistory } from 'vue-router'
import AdminLayout from '../layout/AdminLayout.vue'
import Dashboard from '../views/Dashboard.vue'
import ComingSoon from '../views/ComingSoon.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/dashboard',
      component: AdminLayout,
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: Dashboard
        },
        {
          path: 'activities/list',
          name: 'Activity List',
          component: ComingSoon
        },
        {
          path: 'activities/create',
          name: 'Create Activity',
          component: ComingSoon
        },
        {
          path: 'sessions',
          name: 'Sessions',
          component: ComingSoon
        },
        {
          path: 'inventory',
          name: 'Inventory',
          component: ComingSoon
        },
        {
          path: 'orders',
          name: 'Orders',
          component: ComingSoon
        },
        {
          path: 'risk',
          name: 'Risk Control',
          component: ComingSoon
        },
        {
          path: 'finance',
          name: 'Finance',
          component: ComingSoon
        },
        {
          path: 'reports',
          name: 'Reports',
          component: ComingSoon
        },
        {
          path: 'settings',
          name: 'Settings',
          component: ComingSoon
        }
      ]
    }
  ]
})

export default router
