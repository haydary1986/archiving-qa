import { defineStore } from 'pinia'
import type { User, AuthResponse } from '~/types'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    isAuthenticated: false,
  }),

  getters: {
    roleName: (state) => state.user?.role?.name || '',
    isAdmin: (state) => ['super_admin', 'qa_manager'].includes(state.user?.role?.name || ''),
    isSuperAdmin: (state) => state.user?.role?.name === 'super_admin',
    canCreate: (state) => ['super_admin', 'qa_manager', 'data_entry'].includes(state.user?.role?.name || ''),
  },

  actions: {
    async login(email: string, password: string) {
      const api = useApi()
      const response = await api.post<AuthResponse>('/auth/login', { email, password })

      const token = useCookie('auth_token', { maxAge: response.expires_in })
      const refreshToken = useCookie('refresh_token', { maxAge: 7 * 24 * 3600 })

      token.value = response.access_token
      refreshToken.value = response.refresh_token
      this.user = response.user
      this.isAuthenticated = true

      return response
    },

    async register(email: string, password: string, fullName: string) {
      const api = useApi()
      const response = await api.post<AuthResponse>('/auth/register', {
        email,
        password,
        full_name: fullName,
      })

      const token = useCookie('auth_token', { maxAge: response.expires_in })
      const refreshToken = useCookie('refresh_token', { maxAge: 7 * 24 * 3600 })

      token.value = response.access_token
      refreshToken.value = response.refresh_token
      this.user = response.user
      this.isAuthenticated = true

      return response
    },

    async fetchProfile() {
      try {
        const api = useApi()
        this.user = await api.get<User>('/auth/profile')
        this.isAuthenticated = true
      } catch {
        this.logout()
      }
    },

    logout() {
      const token = useCookie('auth_token')
      const refreshToken = useCookie('refresh_token')

      token.value = null
      refreshToken.value = null
      this.user = null
      this.isAuthenticated = false

      navigateTo('/login')
    },

    async initAuth() {
      const token = useCookie('auth_token')
      if (token.value) {
        await this.fetchProfile()
      }
    },
  },
})
