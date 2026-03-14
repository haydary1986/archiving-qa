<template>
  <div class="space-y-6">
    <UCard>
      <template #header>
        <h3 class="font-semibold">سلة المهملات</h3>
      </template>

      <UTable :columns="columns" :rows="documents" :loading="loading">
        <template #document_type-data="{ row }">
          <UBadge :color="row.document_type === 'incoming' ? 'blue' : row.document_type === 'outgoing' ? 'green' : 'purple'" variant="subtle" size="xs">
            {{ { incoming: 'وارد', outgoing: 'صادر', internal: 'داخلي' }[row.document_type] }}
          </UBadge>
        </template>

        <template #deleted_at-data="{ row }">
          {{ row.deleted_at ? new Date(row.deleted_at).toLocaleDateString('ar-IQ') : '-' }}
        </template>

        <template #actions-data="{ row }">
          <UButton icon="i-heroicons-arrow-uturn-right" variant="ghost" size="xs" @click="restore(row)"
                   title="استعادة">
            استعادة
          </UButton>
        </template>
      </UTable>

      <p v-if="!loading && !documents.length" class="text-center text-gray-500 py-8">
        سلة المهملات فارغة
      </p>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import type { Document } from '~/types'

const api = useApi()
const toast = useToast()

const documents = ref<Document[]>([])
const loading = ref(true)

const columns = [
  { key: 'title', label: 'العنوان' },
  { key: 'document_number', label: 'رقم الكتاب' },
  { key: 'document_type', label: 'النوع' },
  { key: 'deleted_at', label: 'تاريخ الحذف' },
  { key: 'actions', label: 'إجراءات' },
]

const restore = async (doc: Document) => {
  try {
    await api.post(`/documents/${doc.id}/restore`)
    documents.value = documents.value.filter(d => d.id !== doc.id)
    toast.add({ title: 'تم استعادة الوثيقة', color: 'green' })
  } catch {
    toast.add({ title: 'خطأ في الاستعادة', color: 'red' })
  }
}

onMounted(async () => {
  try {
    documents.value = await api.get<Document[]>('/admin/trash')
  } finally {
    loading.value = false
  }
})
</script>
