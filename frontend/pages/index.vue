<template>
  <div class="space-y-6">
    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="stat-card">
        <div class="stat-icon bg-blue-100 dark:bg-blue-900/30">
          <UIcon name="i-heroicons-document-text" class="w-6 h-6 text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <p class="text-2xl font-bold">{{ stats?.total_documents || 0 }}</p>
          <p class="text-sm text-gray-500">إجمالي الوثائق</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon bg-green-100 dark:bg-green-900/30">
          <UIcon name="i-heroicons-paper-clip" class="w-6 h-6 text-green-600 dark:text-green-400" />
        </div>
        <div>
          <p class="text-2xl font-bold">{{ stats?.total_files || 0 }}</p>
          <p class="text-sm text-gray-500">إجمالي الملفات</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon bg-purple-100 dark:bg-purple-900/30">
          <UIcon name="i-heroicons-users" class="w-6 h-6 text-purple-600 dark:text-purple-400" />
        </div>
        <div>
          <p class="text-2xl font-bold">{{ stats?.total_users || 0 }}</p>
          <p class="text-sm text-gray-500">المستخدمون</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon bg-orange-100 dark:bg-orange-900/30">
          <UIcon name="i-heroicons-server-stack" class="w-6 h-6 text-orange-600 dark:text-orange-400" />
        </div>
        <div>
          <p class="text-2xl font-bold">{{ stats?.total_storage_mb || 0 }} MB</p>
          <p class="text-sm text-gray-500">حجم التخزين</p>
        </div>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Documents by Type -->
      <UCard>
        <template #header>
          <h3 class="font-semibold">الوثائق حسب النوع</h3>
        </template>
        <div class="space-y-3">
          <div v-for="(count, type) in stats?.documents_by_type" :key="type"
               class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full" :class="typeColor(type as string)"></span>
              <span>{{ typeLabel(type as string) }}</span>
            </div>
            <UBadge :color="typeBadgeColor(type as string)" variant="subtle">{{ count }}</UBadge>
          </div>
          <div v-if="!stats?.documents_by_type" class="text-center text-gray-500 py-8">
            لا توجد بيانات بعد
          </div>
        </div>
      </UCard>

      <!-- Documents by Classification -->
      <UCard>
        <template #header>
          <h3 class="font-semibold">الوثائق حسب التصنيف</h3>
        </template>
        <div class="space-y-3">
          <div v-for="(count, cls) in stats?.documents_by_classification" :key="cls"
               class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full" :class="classColor(cls as string)"></span>
              <span>{{ classLabel(cls as string) }}</span>
            </div>
            <UBadge :color="classBadgeColor(cls as string)" variant="subtle">{{ count }}</UBadge>
          </div>
          <div v-if="!stats?.documents_by_classification" class="text-center text-gray-500 py-8">
            لا توجد بيانات بعد
          </div>
        </div>
      </UCard>
    </div>

    <!-- Recent Activity -->
    <UCard>
      <template #header>
        <h3 class="font-semibold">آخر النشاطات</h3>
      </template>
      <div class="divide-y divide-gray-200 dark:divide-gray-800">
        <div v-for="activity in stats?.recent_activity" :key="activity.created_at"
             class="flex items-center gap-4 py-3">
          <UIcon :name="actionIcon(activity.action)" class="w-5 h-5 text-gray-400" />
          <div class="flex-1">
            <p class="text-sm">
              <span class="font-medium">{{ activity.user_name }}</span>
              {{ actionLabel(activity.action) }}
              <span class="text-gray-500">{{ activity.resource }}</span>
            </p>
          </div>
          <time class="text-xs text-gray-500">{{ formatDate(activity.created_at) }}</time>
        </div>
        <div v-if="!stats?.recent_activity?.length" class="text-center text-gray-500 py-8">
          لا توجد نشاطات بعد
        </div>
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { DashboardStats } from '~/types'

const api = useApi()
const stats = ref<DashboardStats | null>(null)

onMounted(async () => {
  try {
    stats.value = await api.get<DashboardStats>('/dashboard')
  } catch (e) {
    console.error('Failed to fetch dashboard stats:', e)
  }
})

const typeLabel = (type: string) => ({
  incoming: 'وارد',
  outgoing: 'صادر',
  internal: 'داخلي',
}[type] || type)

const typeColor = (type: string) => ({
  incoming: 'bg-blue-500',
  outgoing: 'bg-green-500',
  internal: 'bg-purple-500',
}[type] || 'bg-gray-500')

const typeBadgeColor = (type: string) => ({
  incoming: 'blue',
  outgoing: 'green',
  internal: 'purple',
}[type] || 'gray') as any

const classLabel = (cls: string) => ({
  normal: 'عادي',
  confidential: 'سري',
  secret: 'سري للغاية',
}[cls] || cls)

const classColor = (cls: string) => ({
  normal: 'bg-gray-500',
  confidential: 'bg-yellow-500',
  secret: 'bg-red-500',
}[cls] || 'bg-gray-500')

const classBadgeColor = (cls: string) => ({
  normal: 'gray',
  confidential: 'yellow',
  secret: 'red',
}[cls] || 'gray') as any

const actionIcon = (action: string) => ({
  create: 'i-heroicons-plus-circle',
  update: 'i-heroicons-pencil-square',
  delete: 'i-heroicons-trash',
  upload: 'i-heroicons-arrow-up-tray',
  download: 'i-heroicons-arrow-down-tray',
  login: 'i-heroicons-arrow-right-on-rectangle',
  view: 'i-heroicons-eye',
}[action] || 'i-heroicons-document')

const actionLabel = (action: string) => ({
  create: 'أنشأ',
  update: 'عدّل',
  delete: 'حذف',
  upload: 'رفع ملفاً في',
  download: 'حمّل',
  login: 'سجل دخول',
  view: 'شاهد',
  restore: 'استعاد',
}[action] || action)

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return 'الآن'
  if (minutes < 60) return `منذ ${minutes} دقيقة`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `منذ ${hours} ساعة`
  return date.toLocaleDateString('ar-IQ')
}
</script>
