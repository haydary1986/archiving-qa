<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <UInput v-model="search" icon="i-heroicons-magnifying-glass" placeholder="بحث عن شخص..."
                class="w-64" @keyup.enter="fetchPersons" />
        <USelect v-model="filterType" :options="typeOptions" class="w-40" />
      </div>
      <UButton v-if="authStore.canCreate" icon="i-heroicons-plus" @click="showModal = true">
        إضافة شخص
      </UButton>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <UCard v-for="person in persons" :key="person.id" class="cursor-pointer hover:shadow-md transition"
             @click="navigateTo(`/persons/${person.id}`)">
        <div class="flex items-center gap-4">
          <UAvatar :alt="person.full_name" size="lg" />
          <div class="flex-1 min-w-0">
            <h3 class="font-semibold truncate">{{ person.full_name }}</h3>
            <p class="text-sm text-gray-500">{{ person.title }}</p>
            <p class="text-xs text-gray-400">{{ person.department }}</p>
            <UBadge :color="personTypeColor(person.person_type)" variant="subtle" size="xs" class="mt-1">
              {{ personTypeLabel(person.person_type) }}
            </UBadge>
          </div>
        </div>
      </UCard>
    </div>

    <p v-if="!persons.length" class="text-center text-gray-500 py-12">
      لا توجد نتائج
    </p>

    <!-- Create Modal -->
    <UModal v-model="showModal">
      <UCard>
        <template #header><h3 class="font-semibold">إضافة شخص جديد</h3></template>
        <form @submit.prevent="createPerson" class="space-y-4">
          <UFormGroup label="الاسم الكامل" required>
            <UInput v-model="form.full_name" />
          </UFormGroup>
          <UFormGroup label="اللقب/المنصب">
            <UInput v-model="form.title" placeholder="أ.د. / م.م. / موظف" />
          </UFormGroup>
          <UFormGroup label="القسم">
            <UInput v-model="form.department" />
          </UFormGroup>
          <UFormGroup label="البريد الإلكتروني">
            <UInput v-model="form.email" type="email" />
          </UFormGroup>
          <UFormGroup label="الهاتف">
            <UInput v-model="form.phone" />
          </UFormGroup>
          <UFormGroup label="النوع" required>
            <USelect v-model="form.person_type" :options="personTypeOptions" />
          </UFormGroup>
        </form>
        <template #footer>
          <div class="flex gap-2 justify-end">
            <UButton variant="ghost" @click="showModal = false">إلغاء</UButton>
            <UButton @click="createPerson" :loading="saving">حفظ</UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { Person } from '~/types'

const api = useApi()
const authStore = useAuthStore()
const toast = useToast()

const persons = ref<Person[]>([])
const search = ref('')
const filterType = ref('')
const showModal = ref(false)
const saving = ref(false)

const form = reactive({
  full_name: '',
  title: '',
  department: '',
  email: '',
  phone: '',
  person_type: 'employee',
})

const typeOptions = [
  { value: '', label: 'جميع الأنواع' },
  { value: 'academic', label: 'أكاديمي' },
  { value: 'employee', label: 'موظف' },
  { value: 'external', label: 'خارجي' },
]

const personTypeOptions = [
  { value: 'academic', label: 'أكاديمي' },
  { value: 'employee', label: 'موظف' },
  { value: 'external', label: 'خارجي' },
]

const personTypeLabel = (t: string) => ({ academic: 'أكاديمي', employee: 'موظف', external: 'خارجي' }[t] || t)
const personTypeColor = (t: string) => ({ academic: 'blue', employee: 'green', external: 'orange' }[t] || 'gray') as any

const fetchPersons = async () => {
  const params: any = {}
  if (search.value) params.search = search.value
  if (filterType.value) params.type = filterType.value
  persons.value = await api.get<Person[]>('/persons', params)
}

const createPerson = async () => {
  saving.value = true
  try {
    await api.post('/persons', form)
    toast.add({ title: 'تم إضافة الشخص', color: 'green' })
    showModal.value = false
    fetchPersons()
  } catch {
    toast.add({ title: 'خطأ', color: 'red' })
  } finally {
    saving.value = false
  }
}

onMounted(fetchPersons)
</script>
