<template>
  <div class="max-w-5xl mx-auto space-y-6" v-if="doc">
    <!-- Header -->
    <div class="flex items-start justify-between">
      <div>
        <h1 class="text-2xl font-bold">{{ doc.title }}</h1>
        <div class="flex items-center gap-3 mt-2">
          <UBadge :color="typeBadgeColor(doc.document_type)" variant="subtle">
            {{ typeLabel(doc.document_type) }}
          </UBadge>
          <UBadge :color="classBadgeColor(doc.classification)" variant="subtle">
            {{ classLabel(doc.classification) }}
          </UBadge>
          <UBadge :color="statusBadgeColor(doc.status)" variant="subtle">
            {{ statusLabel(doc.status) }}
          </UBadge>
          <span class="text-sm text-gray-500" v-if="doc.document_number">
            رقم: {{ doc.document_number }}
          </span>
        </div>
      </div>
      <div class="flex gap-2">
        <UButton v-if="authStore.canCreate" icon="i-heroicons-pencil" variant="outline" size="sm"
                 :to="`/documents/${doc.id}/edit`">
          تعديل
        </UButton>
        <UButton icon="i-heroicons-share" variant="outline" size="sm" @click="showShareModal = true">
          مشاركة
        </UButton>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Content -->
      <div class="lg:col-span-2 space-y-6">
        <!-- Description -->
        <UCard v-if="doc.description">
          <template #header><h3 class="font-semibold">الوصف</h3></template>
          <p class="text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{{ doc.description }}</p>
        </UCard>

        <!-- AI Extracted Data -->
        <UCard v-if="doc.ai_extracted && Object.keys(doc.ai_extracted).length">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-sparkles" class="w-5 h-5 text-purple-500" />
              <h3 class="font-semibold">البيانات المستخرجة بالذكاء الاصطناعي</h3>
            </div>
          </template>
          <div class="grid grid-cols-2 gap-4">
            <div v-for="(value, key) in doc.ai_extracted" :key="key">
              <p class="text-xs text-gray-500">{{ key }}</p>
              <p class="font-medium">{{ value || '-' }}</p>
            </div>
          </div>
        </UCard>

        <!-- Files -->
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="font-semibold">الملفات المرفقة ({{ doc.files?.length || 0 }})</h3>
              <UButton v-if="authStore.canCreate" icon="i-heroicons-plus" size="xs" variant="outline"
                       @click="fileInput?.click()">
                رفع ملف
              </UButton>
              <input ref="fileInput" type="file" multiple class="hidden" @change="uploadNewFile" />
            </div>
          </template>

          <div class="space-y-3">
            <div v-for="file in doc.files" :key="file.id"
                 class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition">
              <UIcon :name="fileIcon(file.mime_type)" class="w-8 h-8 text-gray-400" />
              <div class="flex-1 min-w-0">
                <p class="font-medium truncate">{{ file.original_name }}</p>
                <div class="flex items-center gap-3 text-xs text-gray-500">
                  <span>{{ formatFileSize(file.file_size) }}</span>
                  <UBadge
                    :color="file.ocr_status === 'completed' ? 'green' : file.ocr_status === 'processing' ? 'blue' : 'gray'"
                    variant="subtle" size="xs"
                  >
                    OCR: {{ ocrStatusLabel(file.ocr_status) }}
                  </UBadge>
                </div>
              </div>
              <UButton v-if="file.drive_url" icon="i-heroicons-arrow-down-tray" variant="ghost" size="xs"
                       :href="file.drive_url" target="_blank" />
            </div>

            <p v-if="!doc.files?.length" class="text-center text-gray-500 py-6">
              لا توجد ملفات مرفقة
            </p>
          </div>
        </UCard>

        <!-- Routing History -->
        <UCard v-if="doc.routings?.length">
          <template #header>
            <h3 class="font-semibold">مسار التوجيه</h3>
          </template>
          <div class="relative">
            <div class="absolute top-0 bottom-0 right-4 w-0.5 bg-gray-200 dark:bg-gray-700"></div>
            <div v-for="routing in doc.routings" :key="routing.id" class="relative flex gap-4 pb-6">
              <div class="w-8 h-8 rounded-full bg-primary-100 dark:bg-primary-900 flex items-center justify-center z-10">
                <UIcon name="i-heroicons-arrow-left" class="w-4 h-4 text-primary-600" />
              </div>
              <div class="flex-1 bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <div class="flex items-center gap-2 text-sm">
                  <span class="font-medium">{{ routing.from_entity }}</span>
                  <UIcon name="i-heroicons-arrow-left" class="w-4 h-4 text-gray-400" />
                  <span class="font-medium">{{ routing.to_entity }}</span>
                </div>
                <p class="text-xs text-gray-500 mt-1">
                  {{ routingAction(routing.action) }} -
                  {{ new Date(routing.action_date).toLocaleDateString('ar-IQ') }}
                </p>
                <p v-if="routing.notes" class="text-sm text-gray-600 mt-2">{{ routing.notes }}</p>
              </div>
            </div>
          </div>
        </UCard>
      </div>

      <!-- Sidebar -->
      <div class="space-y-6">
        <!-- Metadata -->
        <UCard>
          <template #header><h3 class="font-semibold">معلومات الوثيقة</h3></template>
          <dl class="space-y-3">
            <div v-if="doc.document_date">
              <dt class="text-xs text-gray-500">تاريخ الكتاب</dt>
              <dd>{{ new Date(doc.document_date).toLocaleDateString('ar-IQ') }}</dd>
            </div>
            <div v-if="doc.source_entity">
              <dt class="text-xs text-gray-500">الجهة المصدرة</dt>
              <dd>{{ doc.source_entity }}</dd>
            </div>
            <div v-if="doc.dest_entity">
              <dt class="text-xs text-gray-500">الجهة المستلمة</dt>
              <dd>{{ doc.dest_entity }}</dd>
            </div>
            <div>
              <dt class="text-xs text-gray-500">تاريخ الإنشاء</dt>
              <dd>{{ new Date(doc.created_at).toLocaleDateString('ar-IQ') }}</dd>
            </div>
          </dl>
        </UCard>

        <!-- Tags -->
        <UCard v-if="doc.tags?.length">
          <template #header><h3 class="font-semibold">الوسوم</h3></template>
          <div class="flex flex-wrap gap-2">
            <UBadge v-for="tag in doc.tags" :key="tag.id" variant="subtle" :style="{ backgroundColor: tag.color + '20', color: tag.color }">
              {{ tag.name }}
            </UBadge>
          </div>
        </UCard>

        <!-- Related Persons -->
        <UCard v-if="doc.persons?.length">
          <template #header><h3 class="font-semibold">الأشخاص المرتبطون</h3></template>
          <div class="space-y-2">
            <NuxtLink v-for="person in doc.persons" :key="person.id" :to="`/persons/${person.id}`"
                      class="flex items-center gap-2 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800">
              <UAvatar :alt="person.full_name" size="xs" />
              <div>
                <p class="text-sm font-medium">{{ person.full_name }}</p>
                <p class="text-xs text-gray-500">{{ person.department }}</p>
              </div>
            </NuxtLink>
          </div>
        </UCard>
      </div>
    </div>

    <!-- Share Modal -->
    <UModal v-model="showShareModal">
      <UCard>
        <template #header><h3 class="font-semibold">إنشاء رابط مشاركة</h3></template>
        <div class="space-y-4">
          <UFormGroup label="كلمة مرور (اختياري)">
            <UInput v-model="sharePassword" type="password" placeholder="اتركه فارغاً للوصول بدون كلمة مرور" />
          </UFormGroup>
          <UFormGroup label="الحد الأقصى للمشاهدات">
            <UInput v-model.number="shareMaxViews" type="number" placeholder="0 = بلا حدود" />
          </UFormGroup>
        </div>
        <template #footer>
          <div class="flex gap-2 justify-end">
            <UButton variant="ghost" @click="showShareModal = false">إلغاء</UButton>
            <UButton @click="createShareLink" :loading="sharing">إنشاء الرابط</UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>

  <div v-else class="flex items-center justify-center py-20">
    <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 animate-spin text-gray-400" />
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const store = useDocumentStore()
const authStore = useAuthStore()
const api = useApi()
const toast = useToast()

const fileInput = ref<HTMLInputElement>()
const showShareModal = ref(false)
const sharePassword = ref('')
const shareMaxViews = ref(0)
const sharing = ref(false)

const doc = computed(() => store.currentDocument)

const typeLabel = (t: string) => ({ incoming: 'وارد', outgoing: 'صادر', internal: 'داخلي' }[t] || t)
const classLabel = (c: string) => ({ normal: 'عادي', confidential: 'سري', secret: 'سري للغاية' }[c] || c)
const statusLabel = (s: string) => ({ draft: 'مسودة', processing: 'قيد المعالجة', completed: 'مكتمل', archived: 'مؤرشف' }[s] || s)
const typeBadgeColor = (t: string) => ({ incoming: 'blue', outgoing: 'green', internal: 'purple' }[t] || 'gray') as any
const classBadgeColor = (c: string) => ({ normal: 'gray', confidential: 'yellow', secret: 'red' }[c] || 'gray') as any
const statusBadgeColor = (s: string) => ({ draft: 'gray', processing: 'blue', completed: 'green', archived: 'purple' }[s] || 'gray') as any

const routingAction = (a: string) => ({ referred: 'تمت الإحالة', forwarded: 'تم التحويل', returned: 'تمت الإعادة', archived: 'تمت الأرشفة' }[a] || a)
const ocrStatusLabel = (s: string) => ({ pending: 'قيد الانتظار', processing: 'جاري المعالجة', completed: 'مكتمل', failed: 'فشل' }[s] || s)

const fileIcon = (mime: string) => {
  if (mime?.includes('pdf')) return 'i-heroicons-document'
  if (mime?.includes('image')) return 'i-heroicons-photo'
  if (mime?.includes('word') || mime?.includes('document')) return 'i-heroicons-document-text'
  return 'i-heroicons-paper-clip'
}

const formatFileSize = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

const uploadNewFile = async (e: Event) => {
  const files = (e.target as HTMLInputElement).files
  if (!files || !doc.value) return
  for (const file of files) {
    try {
      await store.uploadFile(doc.value.id, file)
      toast.add({ title: `تم رفع ${file.name}`, color: 'green' })
      await store.fetchDocument(doc.value.id)
    } catch {
      toast.add({ title: `خطأ في رفع ${file.name}`, color: 'red' })
    }
  }
}

const createShareLink = async () => {
  if (!doc.value) return
  sharing.value = true
  try {
    const result = await api.post<any>('/share', {
      document_id: doc.value.id,
      password: sharePassword.value,
      max_views: shareMaxViews.value,
    })
    const shareUrl = `${window.location.origin}/share/${result.token}`
    await navigator.clipboard.writeText(shareUrl)
    toast.add({ title: 'تم نسخ رابط المشاركة', color: 'green' })
    showShareModal.value = false
  } catch {
    toast.add({ title: 'خطأ في إنشاء رابط المشاركة', color: 'red' })
  } finally {
    sharing.value = false
  }
}

onMounted(async () => {
  await store.fetchDocument(route.params.id as string)
})
</script>
