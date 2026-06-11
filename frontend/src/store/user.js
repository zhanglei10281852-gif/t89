import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import request from '../utils/request'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.role === 'admin')
  const isPlanner = computed(() => userInfo.value?.role === 'planner')

  async function login(username, password) {
    const res = await request.post('/auth/login', { username, password })
    token.value = res.token
    userInfo.value = res.user
    localStorage.setItem('token', res.token)
    localStorage.setItem('user', JSON.stringify(res.user))
    return res
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  async function fetchCurrentUser() {
    const res = await request.get('/auth/me')
    userInfo.value = res
    localStorage.setItem('user', JSON.stringify(res))
    return res
  }

  return { token, userInfo, isLoggedIn, isAdmin, isPlanner, login, logout, fetchCurrentUser }
})
