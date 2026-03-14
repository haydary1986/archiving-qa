import type { AuthResponse } from '~/types'

export const useApi = () => {
  const config = useRuntimeConfig()
  const token = useCookie('auth_token')
  const refreshTokenCookie = useCookie('refresh_token')

  const baseURL = config.public.apiBase as string

  const apiFetch = async <T>(url: string, options: any = {}): Promise<T> => {
    const headers: Record<string, string> = {
      ...options.headers,
    }

    if (token.value) {
      headers['Authorization'] = `Bearer ${token.value}`
    }

    try {
      const response = await $fetch<T>(url, {
        baseURL,
        headers,
        ...options,
      })
      return response
    } catch (error: any) {
      if (error?.statusCode === 401 && refreshTokenCookie.value) {
        // Try to refresh token
        try {
          const refreshResponse = await $fetch<AuthResponse>('/auth/refresh', {
            baseURL,
            method: 'POST',
            body: { refresh_token: refreshTokenCookie.value },
          })

          token.value = refreshResponse.access_token
          refreshTokenCookie.value = refreshResponse.refresh_token

          // Retry original request
          headers['Authorization'] = `Bearer ${refreshResponse.access_token}`
          return await $fetch<T>(url, {
            baseURL,
            headers,
            ...options,
          })
        } catch {
          // Refresh failed, logout
          token.value = null
          refreshTokenCookie.value = null
          navigateTo('/login')
          throw error
        }
      }
      throw error
    }
  }

  return {
    get: <T>(url: string, params?: Record<string, any>) =>
      apiFetch<T>(url, { method: 'GET', params }),

    post: <T>(url: string, body?: any) =>
      apiFetch<T>(url, { method: 'POST', body }),

    put: <T>(url: string, body?: any) =>
      apiFetch<T>(url, { method: 'PUT', body }),

    delete: <T>(url: string) =>
      apiFetch<T>(url, { method: 'DELETE' }),

    upload: <T>(url: string, formData: FormData) =>
      apiFetch<T>(url, {
        method: 'POST',
        body: formData,
      }),
  }
}
