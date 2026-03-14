<template>
  <div class="space-y-6">
    <UCard>
      <template #header>
        <div class="flex flex-wrap items-center gap-4">
          <h3 class="font-semibold">سجل التدقيق</h3>
          <USelect v-model="filterAction" :options="actionOptions" placeholder="نوع العملية" class="w-40" />
          <USelect v-model="filterResource" :options="resourceOptions" placeholder="المورد" class="w-40" />
          <UButton @click="fetchLogs" icon="i-heroicons-funnel" size="sm">تصفية</UButton>
        </div>
      </template>

      <UTable :columns="columns" :rows="logs" :loading="loading">
        <template #action-data="{ row }">
          <UBadge :color="actionColor(row.action)" variant="subtle" size="xs">
            {{ actionLabel(row.action) }}
          </UBadge>
        </template>

        <template #resource-data="{ row }">
          {{ resourceLabel(row.resource) }}
        </template>

        <template #created_at-data="{ row }">
          {{ new Date(row.created_at).toLocaleString('ar-IQ') }}
        </template>
      </UTable>

      <div class="flex justify-center mt-4">
        <UPagination v-model="page" :total="total" :page-count="50" @update:model-value="fetchLogs" />
      </div>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { AuditLog, PaginatedResponse } from '~/types'

const api = useApi()
const logs = ref<AuditLog[]>([])
const loading = ref(true)
const page = ref(1)
const total = ref(0)
const filterAction = ref('')
const filterResource = ref('')

const columns = [
  { key: 'user_name', label: 'المستخدم' },
  { key: 'action', label: 'العملية' },
  { key: 'resource', label: 'المورد' },
  { key: 'ip_address', label: 'عنوان IP' },
  { key: 'created_at', label: 'التاريخ' },
]

const actionOptions = [
  { value: '', label: 'الكل' },
  { value: 'create', label: 'إنشاء' },
  { value: 'update', label: 'تعديل' },
  { value: 'delete', label: 'حذف' },
  { value: 'upload', label: 'رفع' },
  { value: 'download', label: 'تحميل' },
  { value: 'login', label: 'تسجيل دخول' },
]

const resourceOptions = [
  { value: '', label: 'الكل' },
  { value: 'document', label: 'وثيقة' },
  { value: 'file', label: 'ملف' },
  { value: 'user', label: 'مستخدم' },
]

const actionLabel = (a: string) => ({
  create: 'إنشاء', update: 'تعديل', delete: 'حذف',
  upload: 'رفع', download: 'تحميل', login: 'دخول', view: 'مشاهدة', restore: 'استعادة',
}[a] || a)

const actionColor = (a: string) => ({
  create: 'green', update: 'blue', delete: 'red',
  upload: 'purple', download: 'gray', login: 'gray',
}[a] || 'gray') as any

const resourceLabel = (r: string) => ({
  document: 'وثيقة', file: 'ملف', user: 'مستخدم', category: 'تصنيف',
}[r] || r)

const fetchLogs = async () => {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: 50 }
    if (filterAction.value) params.action = filterAction.value
    if (filterResource.value) params.resource = filterResource.value

    const res = await api.get<PaginatedResponse<AuditLog>>('/admin/audit-logs', params)
    logs.value = res.data
    total.value = res.total
  } finally {
    loading.value = false
  }
}

onMounted(fetchLogs)
</script>
