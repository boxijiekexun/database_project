import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import UserHome from '../views/UserHome.vue'
import MerchantHome from '../views/MerchantHome.vue'
import AdminHome from '../views/AdminHome.vue'

const routes = [
  { path: '/', redirect: '/login' },
  { path: '/login', component: Login },
  { path: '/user', component: UserHome },
  { path: '/merchant', component: MerchantHome },
  { path: '/admin', component: AdminHome }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

function defaultPathByRole(role) {
  if (role === 'ADMIN') return '/admin'
  if (role === 'MERCHANT') return '/merchant'
  if (role === 'STUDENT' || role === 'TEACHER') return '/user'
  return '/login'
}

function canVisit(path, role) {
  if (path === '/login') return true
  if (path === '/user') return role === 'STUDENT' || role === 'TEACHER'
  if (path === '/merchant') return role === 'MERCHANT'
  if (path === '/admin') return role === 'ADMIN'
  return false
}

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const role = localStorage.getItem('role')

  if (!token && to.path !== '/login') {
    next('/login')
    return
  }

  if (token && to.path === '/login') {
    next(defaultPathByRole(role))
    return
  }

  if (token && !canVisit(to.path, role)) {
    next(defaultPathByRole(role))
    return
  }

  next()
})

export default router
