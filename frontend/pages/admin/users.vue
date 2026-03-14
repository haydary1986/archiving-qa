<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-semibold">إدارة المستخدمين</h2>
      <UButton icon="i-heroicons-plus" @click="showCreateModal = true">مستخدم جديد</UButton>
    </div>

    <UCard>
      <UTable :columns="columns" :rows="users" :loading="loading">
        <template #role-data="{ row }">
          <UBadge :color="roleBadgeColor(row.role?.name)" variant="subtle">
            {{ roleLabel(row.role?.name) }}
          </UBadge>
        </template>

        <template #is_active-data="{ row }">
          <UToggle v-model="row.is_active" @change="toggleActive(row)" />
        </template>

        <template #last_login_at-data="{ row }">
          {{ row.last_login_at ? new Date(row.last_login_at).toLocaleDateString('ar-IQ') : 'لم يسجل دخول' }}
        </template>

        <template #provider-data="{ row }">
          <UBadge :color="row.provider === 'google' ? 'blue' : 'gray'" variant="subtle" size="xs">
            {{ row.provider === 'google' ? 'Google' : 'محلي' }}
          </UBadge>
        </template>

        <template #actions-data="{ row }">
          <div class="flex gap-1">
            <UButton icon="i-heroicons-pencil" variant="ghost" size="xs" @click="editUser(row)" />
            <UButton icon="i-heroicons-trash" variant="ghost" color="red" size="xs"
                     @click="confirmDelete(row)" v-if="row.id !== authStore.user?.id" />
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Create User Modal -->
    <UModal v-model="showCreateModal">
      <UCard>
        <template #header>
          <h3 class="font-semibold">{{ editingUser ? 'تعديل مستخدم' : 'إنشاء مستخدم جديد' }}</h3>
        </template>
        <form @submit.prevent="handleSaveUser" class="space-y-4">
          <UFormGroup label="الاسم الكامل" required>
            <UInput v-model="userForm.full_name" />
          </UFormGroup>
          <UFormGroup label="البريد الإلكتروني" required v-if="!editingUser">
            <UInput v-model="userForm.email" type="email" />
          </UFormGroup>
          <UFormGroup label="كلمة المرور" :required="!editingUser" v-if="!editingUser">
            <UInput v-model="userForm.password" type="password" />
          </UFormGroup>
          <UFormGroup label="الدور" required>
            <USelect v-model="userForm.role_id" :options="roleOptions" />
          </UFormGroup>
          <UFormGroup label="الأقسام المسموح بالوصول إليها" v-if="isViewerRole(userForm.role_id)">
            <div class="space-y-2">
              <UCheckbox v-for="cat in categories" :key="cat.id"
                         v-model="userForm.allowed_categories"
                         :value="cat.id"
                         :label="cat.name" />
            </div>
            <p class="text-xs text-gray-500 mt-1">اترك فارغاً للوصول إلى جميع الأقسام</p>
          </UFormGroup>
        </form>
        <template #footer>
          <div class="flex gap-2 justify-end">
            <UButton variant="ghost" @click="showCreateModal = false">إلغاء</UButton>
            <UButton @click="handleSaveUser" :loading="saving">
              {{ editingUser ? 'تحديث' : 'إنشاء' }}
            </UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { User, Role, Category } from '~/types'

const api = useApi()
const authStore = useAuthStore()
const toast = useToast()

const users = ref<User[]>([])
const roles = ref<Role[]>([])
const categories = ref<Category[]>([])
const loading = ref(true)
const saving = ref(false)
const showCreateModal = ref(false)
const editingUser = ref<User | null>(null)

const userForm = reactive({
  full_name: '',
  email: '',
  password: '',
  role_id: '',
  allowed_categories: [] as string[],
})

const columns = [
  { key: 'full_name', label: 'الاسم' },
  { key: 'email', label: 'البريد الإلكتروني' },
  { key: 'role', label: 'الدور' },
  { key: 'provider', label: 'نوع الحساب' },
  { key: 'is_active', label: 'نشط' },
  { key: 'last_login_at', label: 'آخر دخول' },
  { key: 'actions', label: 'إجراءات' },
]

const roleOptions = computed(() =>
  roles.value.map(r => ({ value: r.id, label: roleLabel(r.name) }))
)

const roleLabel = (name: string) => ({
  super_admin: 'مدير النظام',
  qa_manager: 'مدير الجودة',
  data_entry: 'مدخل بيانات',
  viewer: 'مشاهد (قراءة فقط)',
}[name] || name)

const roleBadgeColor = (name: string) => ({
  super_admin: 'red',
  qa_manager: 'blue',
  data_entry: 'green',
  viewer: 'gray',
}[name] || 'gray') as any

const isViewerRole = (roleId: string) => {
  const role = roles.value.find(r => r.id === roleId)
  return role?.name === 'viewer'
}

const editUser = (user: User) => {
  editingUser.value = user
  userForm.full_name = user.full_name
  userForm.role_id = user.role_id
  showCreateModal.value = true
}

const toggleActive = async (user: User) => {
  try {
    await api.put(`/admin/users/${user.id}`, { is_active: user.is_active })
    toast.add({ title: `تم ${user.is_active ? 'تفعيل' : 'تعطيل'} الحساب`, color: 'green' })
  } catch {
    toast.add({ title: 'خطأ في تحديث الحالة', color: 'red' })
  }
}

const handleSaveUser = async () => {
  saving.value = true
  try {
    if (editingUser.value) {
      await api.put(`/admin/users/${editingUser.value.id}`, {
        full_name: userForm.full_name,
        role_id: userForm.role_id,
      })
      toast.add({ title: 'تم تحديث المستخدم', color: 'green' })
    } else {
      await api.post('/admin/users', userForm)
      toast.add({ title: 'تم إنشاء المستخدم', color: 'green' })
    }
    showCreateModal.value = false
    editingUser.value = null
    await fetchData()
  } catch (e: any) {
    toast.add({ title: e?.data?.error || 'خطأ', color: 'red' })
  } finally {
    saving.value = false
  }
}

const confirmDelete = async (user: User) => {
  if (!confirm(`هل أنت متأكد من حذف ${user.full_name}؟`)) return
  try {
    await api.delete(`/admin/users/${user.id}`)
    toast.add({ title: 'تم حذف المستخدم', color: 'green' })
    await fetchData()
  } catch {
    toast.add({ title: 'خطأ في حذف المستخدم', color: 'red' })
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const [u, r, c] = await Promise.all([
      api.get<User[]>('/admin/users'),
      api.get<Role[]>('/admin/roles'),
      api.get<Category[]>('/categories'),
    ])
    users.value = u
    roles.value = r
    categories.value = c
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>
