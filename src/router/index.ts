import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import FlagQuizView from '@/views/FlagQuizView.vue'
import MapQuizView from '@/views/MapQuizView.vue'
import ScoreBoardView from '@/views/ScoreBoardView.vue'
import LeafletPage from '@/pages/LeafletPage.vue'
import SignupView from '@/views/SignupView.vue'
import LoginView from '@/views/LoginView.vue'
import AccountView from '@/views/AccountView.vue'
import ProfileView from '@/views/ProfileView.vue'
import AdminView from '@/views/AdminView.vue'
import { getToken, getStoredUser } from '@/lib/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/signup', name: 'signup', component: SignupView, meta: { guest: true } },
    { path: '/login', name: 'login', component: LoginView, meta: { guest: true } },
    { path: '/account', name: 'account', component: AccountView, meta: { requiresAuth: true } },
    { path: '/admin', name: 'admin', component: AdminView, meta: { requiresAuth: true, requiresAdmin: true } },
    { path: '/profile/:username', name: 'profile', component: ProfileView },
    {
      path: '/flag-quiz',
      name: 'flag-quiz',
      component: FlagQuizView,
      meta: { requiresAuth: true },
    },
    {
      path: '/flagquiz',
      redirect: '/flag-quiz',
    },
    {
      path: '/mapquiz',
      name: 'mapquiz',
      component: MapQuizView,
      meta: { requiresAuth: true },
    },
    {
      path: '/score-board',
      name: 'score-board',
      component: ScoreBoardView,
    },
    {
      path: '/map-quiz',
      redirect: '/mapquiz',
    },
    {
      path: '/map',
      name: 'map',
      component: LeafletPage,
    },
    {
      path: '/leafletjs',
      redirect: '/map',
    },
  ],
})

router.beforeEach((to) => {
  const token = getToken()
  const user = getStoredUser()

  if (to.meta.requiresAuth && !token) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.meta.requiresAdmin && !user?.is_admin) {
    return { name: 'home' }
  }
  if (to.meta.guest && token) {
    return { name: 'home' }
  }
  return true
})

export default router
