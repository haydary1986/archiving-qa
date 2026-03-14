<template>
  <div class="max-w-3xl mx-auto space-y-6">
    <UCard>
      <template #header>
        <h3 class="font-semibold">إعدادات النظام</h3>
      </template>

      <div class="space-y-6" v-if="settings">
        <UFormGroup label="اسم النظام">
          <UInput v-model="settings.system_name" />
        </UFormGroup>

        <UFormGroup label="اسم النظام (إنجليزي)">
          <UInput v-model="settings.system_name_en" />
        </UFormGroup>

        <UDivider label="المصادقة" />

        <div class="flex items-center justify-between">
          <div>
            <p class="font-medium">تسجيل الدخول المحلي</p>
            <p class="text-sm text-gray-500">السماح بإنشاء حسابات باستخدام بريد إلكتروني وكلمة مرور</p>
          </div>
          <UToggle v-model="localAuthEnabled" />
        </div>

        <div class="flex items-center justify-between">
          <div>
            <p class="font-medium">تسجيل الدخول بـ Google</p>
            <p class="text-sm text-gray-500">السماح بتسجيل الدخول عبر Google Workspace</p>
          </div>
          <UToggle v-model="googleAuthEnabled" />
        </div>

        <UFormGroup label="النطاقات المسموحة (Google)">
          <UInput v-model="settings.allowed_domains" placeholder="@university.edu.iq" />
          <p class="text-xs text-gray-500 mt-1">النطاقات المسموح لها بتسجيل الدخول عبر Google</p>
        </UFormGroup>

        <UDivider label="المعالجة الذكية" />

        <div class="flex items-center justify-between">
          <div>
            <p class="font-medium">استخراج النص (OCR)</p>
            <p class="text-sm text-gray-500">تفعيل الاستخراج التلقائي للنص من الصور والملفات</p>
          </div>
          <UToggle v-model="ocrEnabled" />
        </div>

        <div class="flex items-center justify-between">
          <div>
            <p class="font-medium">تحليل الذكاء الاصطناعي</p>
            <p class="text-sm text-gray-500">تفعيل التحليل التلقائي للوثائق واستخراج البيانات</p>
          </div>
          <UToggle v-model="aiEnabled" />
        </div>

        <UDivider label="التخزين" />

        <UFormGroup label="الحد الأقصى لحجم الملف (MB)">
          <UInput v-model="settings.max_file_size_mb" type="number" />
        </UFormGroup>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <UButton @click="saveSettings" :loading="saving" icon="i-heroicons-check">
            حفظ الإعدادات
          </UButton>
        </div>
      </template>
    </UCard>

    <!-- Roles Management -->
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="font-semibold">إدارة الأدوار والصلاحيات</h3>
          <UButton icon="i-heroicons-plus" size="sm" @click="showRoleModal = true">دور جديد</UButton>
        </div>
      </template>

      <div class="space-y-4">
        <div v-for="role in roles" :key="role.id"
             class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
          <div class="flex items-center justify-between mb-3">
            <div>
              <h4 class="font-semibold">{{ roleLabel(role.name) }}</h4>
              <p class="text-sm text-gray-500">{{ role.description }}</p>
            </div>
            <UButton icon="i-heroicons-pencil" variant="ghost" size="xs" @click="editRole(role)" />
          </div>
          <div class="flex flex-wrap gap-1">
            <UBadge v-for="perm in role.permissions" :key="perm.id" variant="subtle" size="xs">
              {{ perm.description }}
            </UBadge>
          </div>
        </div>
      </div>
    </UCard>

    <!-- Role Modal -->
    <UModal v-model="showRoleModal">
      <UCard>
        <template #header><h3 class="font-semibold">{{ editingRole ? 'تعديل دور' : 'إنشاء دور جديد' }}</h3></template>
        <div class="space-y-4">
          <UFormGroup label="اسم الدور" required>
            <UInput v-model="roleForm.name" />
          </UFormGroup>
          <UFormGroup label="الوصف">
            <UInput v-model="roleForm.description" />
          </UFormGroup>
          <UFormGroup label="الصلاحيات">
            <div class="space-y-2 max-h-60 overflow-y-auto">
              <UCheckbox v-for="perm in permissions" :key="perm.id"
                         :model-value="roleForm.permission_ids.includes(perm.id)"
                         @change="togglePermission(perm.id)"
                         :label="perm.description" />
            </div>
          </UFormGroup>
        </div>
        <template #footer>
          <div class="flex gap-2 justify-end">
            <UButton variant="ghost" @click="showRoleModal = false">إلغاء</UButton>
            <UButton @click="saveRole" :loading="savingRole">حفظ</UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import type { Role, Permission } from '~/types'

const api = useApi()
const toast = useToast()

const settings = ref<Record<string, string>>({})
const roles = ref<Role[]>([])
const permissions = ref<Permission[]>([])
const saving = ref(false)
const savingRole = ref(false)
const showRoleModal = ref(false)
const editingRole = ref<Role | null>(null)

const localAuthEnabled = computed({
  get: () => settings.value?.local_auth_enabled === 'true',
  set: (v: boolean) => { settings.value.local_auth_enabled = String(v) },
})
const googleAuthEnabled = computed({
  get: () => settings.value?.google_auth_enabled === 'true',
  set: (v: boolean) => { settings.value.google_auth_enabled = String(v) },
})
const ocrEnabled = computed({
  get: () => settings.value?.ocr_enabled === 'true',
  set: (v: boolean) => { settings.value.ocr_enabled = String(v) },
})
const aiEnabled = computed({
  get: () => settings.value?.ai_analysis_enabled === 'true',
  set: (v: boolean) => { settings.value.ai_analysis_enabled = String(v) },
})

const roleForm = reactive({
  name: '',
  description: '',
  permission_ids: [] as string[],
})

const roleLabel = (name: string) => ({
  super_admin: 'مدير النظام',
  qa_manager: 'مدير الجودة',
  data_entry: 'مدخل بيانات',
  viewer: 'مشاهد (قراءة فقط)',
}[name] || name)

const togglePermission = (id: string) => {
  const idx = roleForm.permission_ids.indexOf(id)
  if (idx >= 0) roleForm.permission_ids.splice(idx, 1)
  else roleForm.permission_ids.push(id)
}

const editRole = (role: Role) => {
  editingRole.value = role
  roleForm.name = role.name
  roleForm.description = role.description
  roleForm.permission_ids = role.permissions?.map(p => p.id) || []
  showRoleModal.value = true
}

const saveSettings = async () => {
  saving.value = true
  try {
    for (const [key, value] of Object.entries(settings.value)) {
      await api.put('/admin/settings', { key, value })
    }
    toast.add({ title: 'تم حفظ الإعدادات', color: 'green' })
  } catch {
    toast.add({ title: 'خطأ في حفظ الإعدادات', color: 'red' })
  } finally {
    saving.value = false
  }
}

const saveRole = async () => {
  savingRole.value = true
  try {
    if (editingRole.value) {
      await api.put(`/admin/roles/${editingRole.value.id}`, roleForm)
    } else {
      await api.post('/admin/roles', roleForm)
    }
    toast.add({ title: 'تم حفظ الدور', color: 'green' })
    showRoleModal.value = false
    fetchData()
  } catch {
    toast.add({ title: 'خطأ في حفظ الدور', color: 'red' })
  } finally {
    savingRole.value = false
  }
}

const fetchData = async () => {
  const [s, r, p] = await Promise.all([
    api.get<Record<string, string>>('/admin/settings'),
    api.get<Role[]>('/admin/roles'),
    api.get<Permission[]>('/admin/permissions'),
  ])
  settings.value = s
  roles.value = r
  permissions.value = p
}

onMounted(fetchData)
</script>
