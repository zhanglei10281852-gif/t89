import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../store/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { guest: true }
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    meta: { requiresAuth: true },
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '工作台' }
      },
      {
        path: 'customers',
        name: 'Customers',
        component: () => import('../views/Customers.vue'),
        meta: { title: '客户管理' }
      },
      {
        path: 'customers/funnel',
        name: 'CustomerFunnel',
        component: () => import('../views/CustomerFunnel.vue'),
        meta: { title: '转化漏斗' }
      },
      {
        path: 'packages',
        name: 'Packages',
        component: () => import('../views/Packages.vue'),
        meta: { title: '套餐管理' }
      },
      {
        path: 'quotes',
        name: 'Quotes',
        component: () => import('../views/Quotes.vue'),
        meta: { title: '报价方案' }
      },
      {
        path: 'quotes/editor/:customerId',
        name: 'QuoteEditor',
        component: () => import('../views/QuoteEditor.vue'),
        meta: { title: '报价编辑器' }
      },
      {
        path: 'schedule',
        name: 'Schedule',
        component: () => import('../views/Schedule.vue'),
        meta: { title: '档期日历' }
      },
      {
        path: 'contracts',
        name: 'Contracts',
        component: () => import('../views/Contracts.vue'),
        meta: { title: '签约管理' }
      },
      {
        path: 'stats',
        name: 'Stats',
        component: () => import('../views/Stats.vue'),
        meta: { title: '统计报表' }
      },
      {
        path: 'planner',
        name: 'PlannerWorkbench',
        component: () => import('../views/PlannerWorkbench.vue'),
        meta: { title: '策划师工作台' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/login')
  } else if (to.meta.guest && userStore.isLoggedIn) {
    next('/')
  } else {
    next()
  }
})

export default router
