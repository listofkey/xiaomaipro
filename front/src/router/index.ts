import { createRouter, createWebHistory } from 'vue-router'
import Layout from '../views/layout/index.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'root',
      component: Layout,
      redirect: '/home',
      children: [
        {
          path: 'home',
          name: 'Home',
          component: () => import('../views/home/index.vue'),
        },
        {
          path: 'concert',
          name: 'Concert',
          component: () => import('../views/concert/index.vue'),
        },
        {
          path: 'festival',
          name: 'Festival',
          component: () => import('../views/festival/index.vue'),
        },
        {
          path: 'drama',
          name: 'Drama',
          component: () => import('../views/drama/index.vue'),
        },
        {
          path: 'sports',
          name: 'Sports',
          component: () => import('../views/sports/index.vue'),
        },
        {
          path: 'exhibition',
          name: 'Exhibition',
          component: () => import('../views/exhibition/index.vue'),
        },
        {
          path: 'activity/:id',
          name: 'ActivityDetail',
          component: () => import('../views/activity/detail.vue'),
        },
        {
          path: 'order/create',
          name: 'OrderCreate',
          component: () => import('../views/order/create.vue'),
        },
        {
          path: 'order/list',
          name: 'OrderList',
          component: () => import('../views/order/list.vue'),
        },
        {
          path: 'user/profile',
          name: 'UserProfile',
          component: () => import('../views/user/profile.vue'),
        }
      ]
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/login/index.vue'),
    },
    {
      path: '/payment/processing',
      name: 'PaymentProcessing',
      component: () => import('../views/payment/processing.vue'),
    }
  ]
})

export default router
