<template>
  <div class="space-y-6">
    <!-- Filters -->
    <UCard>
      <div class="flex flex-wrap items-center gap-4">
        <UInput
          v-model="search"
          icon="i-heroicons-magnifying-glass"
          placeholder="بحث في الوثائق..."
          class="w-64"
          @keyup.enter="applyFilters"
        />

        <USelect
          v-model="filterType"
          :options="typeOptions"
          placeholder="نوع الوثيقة"
          class="w-40"
        />

        <USelect
          v-model="filterClassification"
          :options="classOptions"
          placeholder="التصنيف"
          class="w-40"
        />

        <USelect
          v-model="filterStatus"
          :options="statusOptions"
          placeholder="الحالة"
          class="w-40"
        />

        <div class="flex gap-2 mr-auto">
          <UButton @click="applyFilters" icon="i-heroicons-funnel" size="sm">تصفية</UButton>
          <UButton @click="resetFilters" variant="ghost" size="sm">إعادة تعيين</UButton>
          <UButton v-if="authStore.canCreate" to="/documents/new" icon="i-heroicons-plus" color="primary" size="sm">
            وثيقة جديدة
          </UButton>
        </div>
      </div>
    </UCard>

    <!-- Documents Table -->
    <UCard>
      <UTable
        :columns="columns"
        :rows="store.documents"
        :loading="store.loading"
        @select="openDocument"
      >
        <template #document_type-data="{ row }">
          <UBadge :color="typeBadgeColor(row.document_type)" variant="subtle" size="xs">
            {{ typeLabel(row.document_type) }}
          </UBadge>
        </template>

        <template #classification-data="{ row }">
          <UBadge :color="classBadgeColor(row.classification)" variant="subtle" size="xs">
            {{ classLabel(row.classification) }}
          </UBadge>
        </template>

        <template #status-data="{ row }">
          <UBadge :color="statusBadgeColor(row.status)" variant="subtle" size="xs">
            {{ statusLabel(row.status) }}
          </UBadge>
        </template>

        <template #document_date-data="{ row }">
          {{ row.document_date ? new Date(row.document_date).toLocaleDateString('ar-IQ') : '-' }}
        </template>

        <template #created_at-data="{ row }">
          {{ new Date(row.created_at).toLocaleDateString('ar-IQ') }}
        </template>

        <template #actions-data="{ row }">
          <div class="flex gap-1">
            <UButton icon="i-heroicons-eye" variant="ghost" size="xs" @click.stop="openDocument(row)" />
            <UButton v-if="authStore.canCreate" icon="i-heroicons-pencil" variant="ghost" size="xs"
                     @click.stop="editDocument(row)" />
            <UButton v-if="authStore.isAdmin" icon="i-heroicons-trash" variant="ghost" color="red" size="xs"
                     @click.stop="confirmDelete(row)" />
          </div>
        </template>
      </UTable>

      <!-- Pagination -->
      <div class="flex items-center justify-between mt-4 pt-4 border-t border-gray-200 dark:border-gray-800">
        <p class="text-sm text-gray-500">
          عرض {{ (store.page - 1) * store.pageSize + 1 }} إلى {{ Math.min(store.page * store.pageSize, store.total) }}
          من {{ store.total }} نتيجة
        </p>
        <UPagination
          v-model="currentPage"
          :total="store.total"
          :page-count="store.pageSize"
          @update:model-value="changePage"
        />
      </div>
    </UCard>

    <!-- Delete Confirmation -->
    <UModal v-model="showDeleteModal">
      <UCard>
        <template #header>
          <h3 class="font-semibold text-red-600">تأكيد الحذف</h3>
        </template>
        <p>هل أنت متأكد من حذف الوثيقة "{{ documentToDelete?.title }}"؟</p>
        <p class="text-sm text-gray-500 mt-2">سيتم نقل الوثيقة إلى سلة المهملات ويمكن استعادتها لاحقاً.</p>
        <template #footer>
          <div class="flex gap-2 justify-end">
            <UButton variant="ghost" @click="showDeleteModal = false">إلغاء</UButton>
            <UButton color="red" @click="deleteDocument" :loading="deleting">حذف</UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { Document } from '~/types'

const store = useDocumentStore()
const authStore = useAuthStore()
const route = useRoute()
const toast = useToast()

const search = ref(route.query.search as string || '')
const filterType = ref('')
const filterClassification = ref('')
const filterStatus = ref('')
const currentPage = ref(1)
const showDeleteModal = ref(false)
const documentToDelete = ref<Document | null>(null)
const deleting = ref(false)

const columns = [
  { key: 'title', label: 'العنوان', sortable: true },
  { key: 'document_number', label: 'رقم الكتاب' },
  { key: 'document_type', label: 'النوع' },
  { key: 'classification', label: 'التصنيف' },
  { key: 'status', label: 'الحالة' },
  { key: 'document_date', label: 'التاريخ', sortable: true },
  { key: 'created_at', label: 'تاريخ الإنشاء', sortable: true },
  { key: 'actions', label: 'إجراءات' },
]

const typeOptions = [
  { value: '', label: 'الكل' },
  { value: 'incoming', label: 'وارد' },
  { value: 'outgoing', label: 'صادر' },
  { value: 'internal', label: 'داخلي' },
]

const classOptions = [
  { value: '', label: 'الكل' },
  { value: 'normal', label: 'عادي' },
  { value: 'confidential', label: 'سري' },
  { value: 'secret', label: 'سري للغاية' },
]

const statusOptions = [
  { value: '', label: 'الكل' },
  { value: 'draft', label: 'مسودة' },
  { value: 'processing', label: 'قيد المعالجة' },
  { value: 'completed', label: 'مكتمل' },
  { value: 'archived', label: 'مؤرشف' },
]

const typeLabel = (t: string) => ({ incoming: 'وارد', outgoing: 'صادر', internal: 'داخلي' }[t] || t)
const classLabel = (c: string) => ({ normal: 'عادي', confidential: 'سري', secret: 'سري للغاية' }[c] || c)
const statusLabel = (s: string) => ({ draft: 'مسودة', processing: 'قيد المعالجة', completed: 'مكتمل', archived: 'مؤرشف' }[s] || s)

const typeBadgeColor = (t: string) => ({ incoming: 'blue', outgoing: 'green', internal: 'purple' }[t] || 'gray') as any
const classBadgeColor = (c: string) => ({ normal: 'gray', confidential: 'yellow', secret: 'red' }[c] || 'gray') as any
const statusBadgeColor = (s: string) => ({ draft: 'gray', processing: 'blue', completed: 'green', archived: 'purple' }[s] || 'gray') as any

const applyFilters = () => {
  store.fetchDocuments({
    search: search.value,
    document_type: filterType.value,
    classification: filterClassification.value,
    status: filterStatus.value,
    page: 1,
  })
}

const resetFilters = () => {
  search.value = ''
  filterType.value = ''
  filterClassification.value = ''
  filterStatus.value = ''
  currentPage.value = 1
  store.resetFilters()
  store.fetchDocuments()
}

const changePage = (page: number) => {
  store.fetchDocuments({ page })
}

const openDocument = (doc: Document) => {
  navigateTo(`/documents/${doc.id}`)
}

const editDocument = (doc: Document) => {
  navigateTo(`/documents/${doc.id}/edit`)
}

const confirmDelete = (doc: Document) => {
  documentToDelete.value = doc
  showDeleteModal.value = true
}

const deleteDocument = async () => {
  if (!documentToDelete.value) return
  deleting.value = true
  try {
    await store.deleteDocument(documentToDelete.value.id)
    toast.add({ title: 'تم حذف الوثيقة بنجاح', color: 'green' })
    showDeleteModal.value = false
  } catch {
    toast.add({ title: 'خطأ في حذف الوثيقة', color: 'red' })
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  store.fetchDocuments({ search: search.value })
})
</script>
