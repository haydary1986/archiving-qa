<template>
  <div class="max-w-4xl mx-auto space-y-6">
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-semibold">إنشاء وثيقة جديدة</h2>
          <UButton variant="ghost" icon="i-heroicons-x-mark" to="/documents" />
        </div>
      </template>

      <form @submit.prevent="handleSubmit" class="space-y-6">
        <!-- Basic Info -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <UFormGroup label="عنوان الوثيقة" required class="md:col-span-2">
            <UInput v-model="form.title" placeholder="أدخل عنوان الوثيقة" size="lg" />
          </UFormGroup>

          <UFormGroup label="رقم الكتاب">
            <UInput v-model="form.document_number" placeholder="مثال: 1234/ق" />
          </UFormGroup>

          <UFormGroup label="تاريخ الكتاب">
            <UInput v-model="form.document_date" type="date" />
          </UFormGroup>

          <UFormGroup label="نوع الوثيقة" required>
            <USelect v-model="form.document_type" :options="typeOptions" />
          </UFormGroup>

          <UFormGroup label="التصنيف الأمني" required>
            <USelect v-model="form.classification" :options="classOptions" />
          </UFormGroup>

          <UFormGroup label="الجهة المصدرة" v-if="form.document_type === 'incoming'">
            <UInput v-model="form.source_entity" placeholder="مثال: وزارة التعليم العالي" />
          </UFormGroup>

          <UFormGroup label="الجهة المستلمة" v-if="form.document_type === 'outgoing'">
            <UInput v-model="form.dest_entity" placeholder="مثال: رئاسة الجامعة" />
          </UFormGroup>

          <UFormGroup label="التصنيف">
            <USelect
              v-model="form.category_id"
              :options="categoryOptions"
              placeholder="اختر التصنيف"
            />
          </UFormGroup>
        </div>

        <!-- Description -->
        <UFormGroup label="الوصف">
          <UTextarea v-model="form.description" placeholder="وصف الوثيقة..." :rows="3" />
        </UFormGroup>

        <!-- Tags -->
        <UFormGroup label="الوسوم">
          <div class="flex flex-wrap gap-2">
            <UBadge
              v-for="tag in store.tags"
              :key="tag.id"
              :color="selectedTags.includes(tag.id) ? 'primary' : 'gray'"
              variant="subtle"
              class="cursor-pointer"
              @click="toggleTag(tag.id)"
            >
              {{ tag.name }}
            </UBadge>
          </div>
        </UFormGroup>

        <!-- Persons -->
        <UFormGroup label="الأشخاص ذوو العلاقة">
          <div class="space-y-2">
            <div v-for="(person, idx) in selectedPersons" :key="idx"
                 class="flex items-center gap-2 bg-gray-50 dark:bg-gray-800 rounded-lg p-2">
              <USelect
                v-model="person.person_id"
                :options="personOptions"
                placeholder="اختر شخص"
                class="flex-1"
              />
              <USelect
                v-model="person.relation"
                :options="relationOptions"
                class="w-32"
              />
              <UButton icon="i-heroicons-x-mark" variant="ghost" color="red" size="xs"
                       @click="selectedPersons.splice(idx, 1)" />
            </div>
            <UButton icon="i-heroicons-plus" variant="outline" size="sm" @click="addPerson">
              إضافة شخص
            </UButton>
          </div>
        </UFormGroup>

        <!-- File Upload -->
        <UFormGroup label="المرفقات">
          <div
            class="border-2 border-dashed border-gray-300 dark:border-gray-700 rounded-lg p-8 text-center
                   hover:border-primary-400 transition-colors cursor-pointer"
            @click="fileInput?.click()"
            @dragover.prevent
            @drop.prevent="handleDrop"
          >
            <UIcon name="i-heroicons-cloud-arrow-up" class="w-12 h-12 mx-auto text-gray-400 mb-4" />
            <p class="text-gray-600 dark:text-gray-400">اسحب الملفات هنا أو انقر للاختيار</p>
            <p class="text-xs text-gray-500 mt-1">PDF, Word, Images - حد أقصى 50MB</p>
            <input ref="fileInput" type="file" multiple accept=".pdf,.doc,.docx,.jpg,.jpeg,.png,.tiff" class="hidden"
                   @change="handleFileSelect" />
          </div>

          <!-- Selected Files -->
          <div v-if="selectedFiles.length" class="mt-4 space-y-2">
            <div v-for="(file, idx) in selectedFiles" :key="idx"
                 class="flex items-center gap-3 bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
              <UIcon :name="fileIcon(file.type)" class="w-5 h-5 text-gray-400" />
              <div class="flex-1 min-w-0">
                <p class="text-sm truncate">{{ file.name }}</p>
                <p class="text-xs text-gray-500">{{ formatFileSize(file.size) }}</p>
              </div>
              <UButton icon="i-heroicons-x-mark" variant="ghost" color="red" size="xs"
                       @click="selectedFiles.splice(idx, 1)" />
            </div>
          </div>
        </UFormGroup>

        <UAlert v-if="error" color="red" :title="error" />

        <!-- Actions -->
        <div class="flex gap-3 justify-end pt-4 border-t border-gray-200 dark:border-gray-800">
          <UButton variant="ghost" to="/documents">إلغاء</UButton>
          <UButton type="submit" :loading="submitting" icon="i-heroicons-check">
            حفظ الوثيقة
          </UButton>
        </div>
      </form>
    </UCard>
  </div>
</template>

<script setup lang="ts">
const store = useDocumentStore()
const toast = useToast()

const fileInput = ref<HTMLInputElement>()
const submitting = ref(false)
const error = ref('')

const form = reactive({
  title: '',
  description: '',
  document_number: '',
  document_date: '',
  document_type: 'incoming',
  classification: 'normal',
  source_entity: '',
  dest_entity: '',
  category_id: '',
})

const selectedTags = ref<string[]>([])
const selectedPersons = ref<{ person_id: string; relation: string }[]>([])
const selectedFiles = ref<File[]>([])

const typeOptions = [
  { value: 'incoming', label: 'وارد' },
  { value: 'outgoing', label: 'صادر' },
  { value: 'internal', label: 'داخلي' },
]

const classOptions = [
  { value: 'normal', label: 'عادي' },
  { value: 'confidential', label: 'سري' },
  { value: 'secret', label: 'سري للغاية' },
]

const relationOptions = [
  { value: 'sender', label: 'مرسل' },
  { value: 'receiver', label: 'مستلم' },
  { value: 'related', label: 'ذو علاقة' },
  { value: 'cc', label: 'نسخة' },
]

const categoryOptions = computed(() =>
  store.categories.map(c => ({ value: c.id, label: c.name }))
)

const personOptions = computed(() =>
  store.persons.map(p => ({ value: p.id, label: `${p.full_name} - ${p.department || ''}` }))
)

const toggleTag = (id: string) => {
  const idx = selectedTags.value.indexOf(id)
  if (idx >= 0) selectedTags.value.splice(idx, 1)
  else selectedTags.value.push(id)
}

const addPerson = () => {
  selectedPersons.value.push({ person_id: '', relation: 'related' })
}

const handleFileSelect = (e: Event) => {
  const files = (e.target as HTMLInputElement).files
  if (files) selectedFiles.value.push(...Array.from(files))
}

const handleDrop = (e: DragEvent) => {
  const files = e.dataTransfer?.files
  if (files) selectedFiles.value.push(...Array.from(files))
}

const fileIcon = (type: string) => {
  if (type.includes('pdf')) return 'i-heroicons-document'
  if (type.includes('image')) return 'i-heroicons-photo'
  return 'i-heroicons-paper-clip'
}

const formatFileSize = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

const handleSubmit = async () => {
  if (!form.title) {
    error.value = 'عنوان الوثيقة مطلوب'
    return
  }

  submitting.value = true
  error.value = ''

  try {
    const docData: any = {
      ...form,
      tag_ids: selectedTags.value,
      persons: selectedPersons.value.filter(p => p.person_id),
    }

    if (form.document_date) {
      docData.document_date = new Date(form.document_date).toISOString()
    } else {
      delete docData.document_date
    }

    if (!form.category_id) delete docData.category_id

    const doc = await store.createDocument(docData)

    // Upload files
    for (const file of selectedFiles.value) {
      try {
        await store.uploadFile(doc.id, file)
      } catch (e) {
        console.error('Failed to upload file:', file.name, e)
      }
    }

    toast.add({ title: 'تم إنشاء الوثيقة بنجاح', color: 'green' })
    navigateTo(`/documents/${doc.id}`)
  } catch (e: any) {
    error.value = e?.data?.error || 'خطأ في إنشاء الوثيقة'
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await Promise.all([
    store.fetchCategories(),
    store.fetchTags(),
    store.fetchPersons(),
  ])
})
</script>
