<template>
  <div class="min-h-screen flex">
    <!-- Sidebar -->
    <aside class="no-print w-64 bg-white dark:bg-gray-900 border-l border-gray-200 dark:border-gray-800 flex flex-col fixed h-full z-30"
           :class="{ '-translate-x-full': !sidebarOpen, 'translate-x-0': sidebarOpen }"
           style="transition: transform 0.3s ease;">
      <!-- Logo -->
      <div class="p-6 border-b border-gray-200 dark:border-gray-800">
        <h1 class="text-lg font-bold text-primary-600 dark:text-primary-400">
          نظام الأرشفة الإلكتروني
        </h1>
        <p class="text-xs text-gray-500 mt-1">قسم ضمان الجودة وتقييم الأداء</p>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 p-4 space-y-1 overflow-y-auto">
        <NuxtLink to="/" class="sidebar-link" active-class="active">
          <UIcon name="i-heroicons-home" class="w-5 h-5" />
          <span>لوحة التحكم</span>
        </NuxtLink>

        <NuxtLink to="/documents" class="sidebar-link" active-class="active">
          <UIcon name="i-heroicons-document-text" class="w-5 h-5" />
          <span>الوثائق</span>
        </NuxtLink>

        <NuxtLink to="/documents/new" class="sidebar-link" active-class="active" v-if="authStore.canCreate">
          <UIcon name="i-heroicons-plus-circle" class="w-5 h-5" />
          <span>وثيقة جديدة</span>
        </NuxtLink>

        <NuxtLink to="/persons" class="sidebar-link" active-class="active">
          <UIcon name="i-heroicons-users" class="w-5 h-5" />
          <span>الأشخاص</span>
        </NuxtLink>

        <div class="pt-4 mt-4 border-t border-gray-200 dark:border-gray-800" v-if="authStore.isAdmin">
          <p class="px-4 text-xs font-semibold text-gray-400 uppercase mb-2">الإدارة</p>

          <NuxtLink to="/admin/users" class="sidebar-link" active-class="active">
            <UIcon name="i-heroicons-user-group" class="w-5 h-5" />
            <span>المستخدمون</span>
          </NuxtLink>

          <NuxtLink to="/admin/categories" class="sidebar-link" active-class="active">
            <UIcon name="i-heroicons-folder" class="w-5 h-5" />
            <span>التصنيفات</span>
          </NuxtLink>

          <NuxtLink to="/admin/audit" class="sidebar-link" active-class="active">
            <UIcon name="i-heroicons-clipboard-document-list" class="w-5 h-5" />
            <span>سجل التدقيق</span>
          </NuxtLink>

          <NuxtLink to="/admin/settings" class="sidebar-link" active-class="active" v-if="authStore.isSuperAdmin">
            <UIcon name="i-heroicons-cog-6-tooth" class="w-5 h-5" />
            <span>الإعدادات</span>
          </NuxtLink>

          <NuxtLink to="/admin/queue" class="sidebar-link" active-class="active">
            <UIcon name="i-heroicons-queue-list" class="w-5 h-5" />
            <span>الطابور</span>
          </NuxtLink>

          <NuxtLink to="/admin/trash" class="sidebar-link" active-class="active" v-if="authStore.isSuperAdmin">
            <UIcon name="i-heroicons-trash" class="w-5 h-5" />
            <span>سلة المهملات</span>
          </NuxtLink>
        </div>
      </nav>

      <!-- User Info -->
      <div class="p-4 border-t border-gray-200 dark:border-gray-800">
        <div class="flex items-center gap-3">
          <UAvatar :alt="authStore.user?.full_name || ''" size="sm" />
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium truncate">{{ authStore.user?.full_name }}</p>
            <p class="text-xs text-gray-500 truncate">{{ authStore.user?.email }}</p>
          </div>
          <UButton icon="i-heroicons-arrow-right-on-rectangle" variant="ghost" size="xs" @click="authStore.logout()" />
        </div>
      </div>
    </aside>

    <!-- Main Content -->
    <div class="flex-1 mr-64">
      <!-- Top Bar -->
      <header class="no-print bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 sticky top-0 z-20">
        <div class="flex items-center justify-between px-6 py-3">
          <div class="flex items-center gap-4">
            <UButton
              icon="i-heroicons-bars-3"
              variant="ghost"
              class="lg:hidden"
              @click="sidebarOpen = !sidebarOpen"
            />
            <h2 class="text-lg font-semibold">{{ pageTitle }}</h2>
          </div>

          <div class="flex items-center gap-3">
            <UInput
              v-model="globalSearch"
              icon="i-heroicons-magnifying-glass"
              placeholder="بحث سريع..."
              class="w-64"
              @keyup.enter="performSearch"
            />
            <UButton
              :icon="colorMode.value === 'dark' ? 'i-heroicons-sun' : 'i-heroicons-moon'"
              variant="ghost"
              @click="colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'"
            />
          </div>
        </div>
      </header>

      <!-- Page Content -->
      <main class="page-container">
        <slot />
      </main>
    </div>

    <!-- Mobile Overlay -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 bg-black/50 z-20 lg:hidden"
      @click="sidebarOpen = false"
    />
  </div>
</template>

<script setup lang="ts">
const authStore = useAuthStore()
const colorMode = useColorMode()
const route = useRoute()

const sidebarOpen = ref(true)
const globalSearch = ref('')

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    '/': 'لوحة التحكم',
    '/documents': 'الوثائق',
    '/documents/new': 'وثيقة جديدة',
    '/persons': 'الأشخاص',
    '/admin/users': 'إدارة المستخدمين',
    '/admin/categories': 'التصنيفات',
    '/admin/audit': 'سجل التدقيق',
    '/admin/settings': 'إعدادات النظام',
    '/admin/queue': 'الطابور والعمليات',
    '/admin/trash': 'سلة المهملات',
  }
  return titles[route.path] || 'نظام الأرشفة'
})

const performSearch = () => {
  if (globalSearch.value) {
    navigateTo({ path: '/documents', query: { search: globalSearch.value } })
  }
}
</script>
