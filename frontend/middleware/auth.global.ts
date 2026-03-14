export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie('auth_token')
  const publicPages = ['/login', '/register', '/share']

  const isPublic = publicPages.some(page => to.path.startsWith(page))

  if (!token.value && !isPublic) {
    return navigateTo('/login')
  }

  if (token.value && (to.path === '/login' || to.path === '/register')) {
    return navigateTo('/')
  }
})
