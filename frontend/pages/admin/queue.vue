<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-semibold">الطابور والعمليات</h2>
      <div class="flex gap-2">
        <UButton icon="i-heroicons-arrow-path" variant="outline" @click="refresh" :loading="loading">
          تحديث
        </UButton>
        <UButton icon="i-heroicons-trash" variant="outline" color="gray" @click="clearCompleted" v-if="stats.by_status?.completed">
          حذف المكتملة
        </UButton>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
      <UCard>
        <div class="text-center">
          <p class="text-2xl font-bold text-yellow-500">{{ stats.by_status?.pending || 0 }}</p>
          <p class="text-sm text-gray-500">في الانتظار</p>
        </div>
      </UCard>
      <UCard>
        <div class="text-center">
          <p class="text-2xl font-bold text-blue-500">{{ stats.active_count || 0 }}</p>
          <p class="text-sm text-gray-500">قيد المعالجة</p>
        </div>
      </UCard>
      <UCard>
        <div class="text-center">
          <p class="text-2xl font-bold text-green-500">{{ stats.by_status?.completed || 0 }}</p>
          <p class="text-sm text-gray-500">مكتملة</p>
        </div>
      </UCard>
      <UCard>
        <div class="text-center">
          <p class="text-2xl font-bold text-red-500">{{ stats.failed_count || 0 }}</p>
          <p class="text-sm text-gray-500">فاشلة</p>
        </div>
      </UCard>
      <UCard>
        <div class="text-center">
          <p class="text-2xl font-bold text-gray-500">
            {{ Object.values(stats.by_status || {}).reduce((a: number, b: any) => a + (b as number), 0) }}
          </p>
          <p class="text-sm text-gray-500">الإجمالي</p>
        </div>
      </UCard>
    </div>

    <!-- OCR Stats -->
    <UCard v-if="stats.ocr_status">
      <template #header>
        <h3 class="font-semibold">حالة OCR للملفات</h3>
      </template>
      <div class="flex flex-wrap gap-4">
        <div v-for="(count, status) in (stats.ocr_status as Record<string, number>)" :key="status"
             class="flex items-center gap-2">
          <UBadge :color="ocrStatusColor(status as string)" variant="subtle">
            {{ ocrStatusLabel(status as string) }}
          </UBadge>
          <span class="font-medium">{{ count }}</span>
        </div>
      </div>
    </UCard>

    <!-- Filters -->
    <div class="flex gap-3">
      <USelect v-model="filterStatus" :options="statusOptions" placeholder="كل الحالات" class="w-40"
               @change="fetchJobs" />
      <USelect v-model="filterType" :options="typeOptions" placeholder="كل الأنواع" class="w-40"
               @change="fetchJobs" />
    </div>

    <!-- Jobs Table -->
    <UCard>
      <UTable :columns="columns" :rows="jobs" :loading="loading">
        <template #task_type-data="{ row }">
          <UBadge :color="taskTypeColor(row.task_type)" variant="subtle">
            {{ taskTypeLabel(row.task_type) }}
          </UBadge>
        </template>

        <template #status-data="{ row }">
          <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full" :class="statusDotClass(row.status)" />
            <span>{{ statusLabel(row.status) }}</span>
          </div>
        </template>

        <template #doc_title-data="{ row }">
          <div class="max-w-48 truncate">
            <span v-if="row.doc_title">{{ row.doc_title }}</span>
            <span v-else class="text-gray-400">-</span>
          </div>
        </template>

        <template #file_name-data="{ row }">
          <div class="max-w-32 truncate">
            <span v-if="row.file_name">{{ row.file_name }}</span>
            <span v-else class="text-gray-400">-</span>
          </div>
        </template>

        <template #attempts-data="{ row }">
          {{ row.attempts }} / {{ row.max_retries }}
        </template>

        <template #created_at-data="{ row }">
          <span class="text-sm text-gray-500">{{ formatDate(row.created_at) }}</span>
        </template>

        <template #actions-data="{ row }">
          <div class="flex gap-1">
            <UButton v-if="row.status === 'failed'" icon="i-heroicons-arrow-path" variant="ghost" size="xs"
                     color="yellow" @click="retryJob(row.id)" title="إعادة المحاولة" />
            <UButton v-if="row.error_message" icon="i-heroicons-exclamation-triangle" variant="ghost" size="xs"
                     color="red" @click="showError(row)" title="عرض الخطأ" />
          </div>
        </template>
      </UTable>

      <div v-if="!jobs.length && !loading" class="text-center py-12 text-gray-500">
        <UIcon name="i-heroicons-queue-list" class="w-12 h-12 mx-auto mb-4 text-gray-300" />
        <p>لا توجد مهام في الطابور</p>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex justify-center mt-4">
        <UPagination v-model="page" :page-count="pageSize" :total="total" @update:model-value="fetchJobs" />
      </div>
    </UCard>

    <!-- Error Modal -->
    <UModal v-model="showErrorModal">
      <UCard>
        <template #header>
          <h3 class="font-semibold text-red-500">تفاصيل الخطأ</h3>
        </template>
        <div class="space-y-2">
          <p class="text-sm"><strong>المهمة:</strong> {{ selectedJob?.task_type }}</p>
          <p class="text-sm"><strong>الملف:</strong> {{ selectedJob?.file_name || '-' }}</p>
          <div class="bg-red-50 dark:bg-red-900/20 p-3 rounded-lg">
            <pre class="text-sm text-red-600 dark:text-red-400 whitespace-pre-wrap">{{ selectedJob?.error_message }}</pre>
          </div>
        </div>
        <template #footer>
          <div class="flex justify-end">
            <UButton @click="showErrorModal = false" variant="ghost">إغلاق</UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
const api = useApi()
const toast = useToast()

const loading = ref(false)
const jobs = ref<any[]>([])
const stats = ref<any>({})
const total = ref(0)
const page = ref(1)
const pageSize = 50
const totalPages = computed(() => Math.ceil(total.value / pageSize))

const filterStatus = ref('')
const filterType = ref('')

const showErrorModal = ref(false)
const selectedJob = ref<any>(null)

const columns = [
  { key: 'task_type', label: 'النوع' },
  { key: 'status', label: 'الحالة' },
  { key: 'doc_title', label: 'الوثيقة' },
  { key: 'file_name', label: 'الملف' },
  { key: 'attempts', label: 'المحاولات' },
  { key: 'created_at', label: 'التاريخ' },
  { key: 'actions', label: '' },
]

const statusOptions = [
  { value: '', label: 'كل الحالات' },
  { value: 'pending', label: 'في الانتظار' },
  { value: 'processing', label: 'قيد المعالجة' },
  { value: 'completed', label: 'مكتملة' },
  { value: 'failed', label: 'فاشلة' },
]

const typeOptions = [
  { value: '', label: 'كل الأنواع' },
  { value: 'ocr:process', label: 'OCR' },
  { value: 'ai:analyze', label: 'تحليل AI' },
  { value: 'file:compress', label: 'ضغط' },
]

const taskTypeLabel = (type: string) => {
  const map: Record<string, string> = {
    'ocr:process': 'استخراج نص OCR',
    'ai:analyze': 'تحليل ذكي AI',
    'file:compress': 'ضغط ملف',
  }
  return map[type] || type
}

const taskTypeColor = (type: string) => {
  const map: Record<string, string> = {
    'ocr:process': 'blue',
    'ai:analyze': 'purple',
    'file:compress': 'orange',
  }
  return map[type] || 'gray'
}

const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    'pending': 'في الانتظار',
    'processing': 'قيد المعالجة',
    'completed': 'مكتملة',
    'failed': 'فاشلة',
  }
  return map[status] || status
}

const statusDotClass = (status: string) => {
  const map: Record<string, string> = {
    'pending': 'bg-yellow-400',
    'processing': 'bg-blue-400 animate-pulse',
    'completed': 'bg-green-400',
    'failed': 'bg-red-400',
  }
  return map[status] || 'bg-gray-400'
}

const ocrStatusLabel = (status: string) => {
  const map: Record<string, string> = {
    'pending': 'في الانتظار',
    'processing': 'قيد المعالجة',
    'completed': 'مكتمل',
    'failed': 'فشل',
  }
  return map[status] || status
}

const ocrStatusColor = (status: string) => {
  const map: Record<string, string> = {
    'pending': 'yellow',
    'processing': 'blue',
    'completed': 'green',
    'failed': 'red',
  }
  return map[status] || 'gray'
}

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('ar-IQ', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit'
  })
}

const fetchJobs = async () => {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize }
    if (filterStatus.value) params.status = filterStatus.value
    if (filterType.value) params.task_type = filterType.value

    const res = await api.get<any>('/admin/queue', params)
    jobs.value = res.data || []
    total.value = res.total || 0
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    stats.value = await api.get<any>('/admin/queue/stats')
  } catch {
    // ignore
  }
}

const retryJob = async (id: string) => {
  try {
    await api.post(`/admin/queue/${id}/retry`)
    toast.add({ title: 'تم إعادة المهمة', color: 'green' })
    refresh()
  } catch {
    toast.add({ title: 'خطأ', color: 'red' })
  }
}

const clearCompleted = async () => {
  if (!confirm('هل تريد حذف جميع المهام المكتملة؟')) return
  try {
    await api.delete('/admin/queue/completed')
    toast.add({ title: 'تم الحذف', color: 'green' })
    refresh()
  } catch {
    toast.add({ title: 'خطأ', color: 'red' })
  }
}

const showError = (job: any) => {
  selectedJob.value = job
  showErrorModal.value = true
}

const refresh = () => {
  fetchJobs()
  fetchStats()
}

// Auto-refresh every 10 seconds
let interval: ReturnType<typeof setInterval>
onMounted(() => {
  refresh()
  interval = setInterval(refresh, 10000)
})
onUnmounted(() => clearInterval(interval))
</script>
