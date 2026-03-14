<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-semibold">إدارة التصنيفات</h2>
      <UButton icon="i-heroicons-plus" @click="openCreate()">تصنيف جديد</UButton>
    </div>

    <UCard>
      <div class="space-y-2">
        <template v-for="cat in categories" :key="cat.id">
          <div class="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-folder" class="w-5 h-5 text-yellow-500" />
              <span class="font-medium">{{ cat.name }}</span>
            </div>
            <div class="flex gap-1">
              <UButton icon="i-heroicons-plus" variant="ghost" size="xs" @click="openCreate(cat.id)" title="إضافة تصنيف فرعي" />
              <UButton icon="i-heroicons-pencil" variant="ghost" size="xs" @click="editCategory(cat)" />
              <UButton icon="i-heroicons-trash" variant="ghost" color="red" size="xs" @click="deleteCategory(cat.id)" />
            </div>
          </div>
          <!-- Children -->
          <div v-for="child in cat.children" :key="child.id"
               class="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 mr-8">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-folder-open" class="w-5 h-5 text-blue-500" />
              <span>{{ child.name }}</span>
            </div>
            <div class="flex gap-1">
              <UButton icon="i-heroicons-pencil" variant="ghost" size="xs" @click="editCategory(child)" />
              <UButton icon="i-heroicons-trash" variant="ghost" color="red" size="xs" @click="deleteCategory(child.id)" />
            </div>
          </div>
        </template>

        <p v-if="!categories.length" class="text-center text-gray-500 py-8">
          لا توجد تصنيفات بعد
        </p>
      </div>
    </UCard>

    <UModal v-model="showModal">
      <UCard>
        <template #header>
          <h3 class="font-semibold">{{ editing ? 'تعديل تصنيف' : 'تصنيف جديد' }}</h3>
        </template>
        <div class="space-y-4">
          <UFormGroup label="اسم التصنيف" required>
            <UInput v-model="form.name" />
          </UFormGroup>
          <UFormGroup label="الترتيب">
            <UInput v-model.number="form.sort_order" type="number" />
          </UFormGroup>
        </div>
        <template #footer>
          <div class="flex gap-2 justify-end">
            <UButton variant="ghost" @click="showModal = false">إلغاء</UButton>
            <UButton @click="saveCategory" :loading="saving">حفظ</UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { Category } from '~/types'

const api = useApi()
const toast = useToast()

const categories = ref<Category[]>([])
const showModal = ref(false)
const editing = ref<Category | null>(null)
const saving = ref(false)

const form = reactive({
  name: '',
  parent_id: null as string | null,
  sort_order: 0,
})

const openCreate = (parentId?: string) => {
  editing.value = null
  form.name = ''
  form.parent_id = parentId || null
  form.sort_order = 0
  showModal.value = true
}

const editCategory = (cat: Category) => {
  editing.value = cat
  form.name = cat.name
  form.parent_id = cat.parent_id || null
  form.sort_order = cat.sort_order
  showModal.value = true
}

const saveCategory = async () => {
  saving.value = true
  try {
    if (editing.value) {
      await api.put(`/categories/${editing.value.id}`, form)
    } else {
      await api.post('/categories', form)
    }
    toast.add({ title: 'تم الحفظ', color: 'green' })
    showModal.value = false
    fetchCategories()
  } catch {
    toast.add({ title: 'خطأ', color: 'red' })
  } finally {
    saving.value = false
  }
}

const deleteCategory = async (id: string) => {
  if (!confirm('هل أنت متأكد من حذف هذا التصنيف؟')) return
  try {
    await api.delete(`/categories/${id}`)
    toast.add({ title: 'تم الحذف', color: 'green' })
    fetchCategories()
  } catch {
    toast.add({ title: 'خطأ', color: 'red' })
  }
}

const fetchCategories = async () => {
  categories.value = await api.get<Category[]>('/categories')
}

onMounted(fetchCategories)
</script>
