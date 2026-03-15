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

    <!-- Google Drive Connection -->
    <UCard>
      <template #header>
        <div class="flex items-center gap-2">
          <svg class="w-6 h-6" viewBox="0 0 87.3 78" xmlns="http://www.w3.org/2000/svg">
            <path d="m6.6 66.85 3.85 6.65c.8 1.4 1.95 2.5 3.3 3.3l13.75-23.8h-27.5c0 1.55.4 3.1 1.2 4.5z" fill="#0066da"/>
            <path d="m43.65 25-13.75-23.8c-1.35.8-2.5 1.9-3.3 3.3l-20.4 35.3c-.8 1.4-1.2 2.95-1.2 4.5h27.5z" fill="#00ac47"/>
            <path d="m73.55 76.8c1.35-.8 2.5-1.9 3.3-3.3l1.6-2.75 7.65-13.25c.8-1.4 1.2-2.95 1.2-4.5h-27.502l5.852 11.5z" fill="#ea4335"/>
            <path d="m43.65 25 13.75-23.8c-1.35-.8-2.9-1.2-4.5-1.2h-18.5c-1.6 0-3.15.45-4.5 1.2z" fill="#00832d"/>
            <path d="m59.8 53h-32.3l-13.75 23.8c1.35.8 2.9 1.2 4.5 1.2h50.8c1.6 0 3.15-.45 4.5-1.2z" fill="#2684fc"/>
            <path d="m73.4 26.5-10.2-17.65c-.8-1.4-1.95-2.5-3.3-3.3l-13.75 23.8 16.15 23.8h27.45c0-1.55-.4-3.1-1.2-4.5z" fill="#ffba00"/>
          </svg>
          <h3 class="font-semibold">ربط Google Drive</h3>
        </div>
      </template>

      <div class="space-y-4">
        <!-- Connection Status -->
        <div v-if="driveStatus.connected" class="flex items-center gap-3 p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
          <UIcon name="i-heroicons-check-circle" class="text-green-500 text-xl flex-shrink-0" />
          <div class="flex-1">
            <p class="font-medium text-green-700 dark:text-green-400">متصل بـ Google Drive</p>
            <p class="text-sm text-green-600 dark:text-green-500" v-if="driveStatus.email">{{ driveStatus.email }}</p>
          </div>
          <UButton color="red" variant="soft" size="sm" @click="disconnectDrive" :loading="driveDisconnecting">
            قطع الاتصال
          </UButton>
        </div>

        <div v-else class="flex items-center gap-3 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
          <UIcon name="i-heroicons-cloud" class="text-gray-400 text-xl flex-shrink-0" />
          <div class="flex-1">
            <p class="font-medium">غير متصل</p>
            <p class="text-sm text-gray-500">قم بربط حسابك في Google Drive لرفع الملفات تلقائياً</p>
          </div>
        </div>

        <!-- Step 1: Enter Client Credentials -->
        <div v-if="!driveStatus.connected" class="space-y-4">
          <div class="p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
            <p class="text-sm text-blue-700 dark:text-blue-400 mb-2 font-medium">خطوات الربط:</p>
            <ol class="text-sm text-blue-600 dark:text-blue-400 space-y-1 list-decimal list-inside">
              <li>اذهب إلى <a href="https://console.cloud.google.com/apis/credentials" target="_blank" class="underline font-medium">Google Cloud Console</a></li>
              <li>أنشئ OAuth 2.0 Client ID (نوع Web Application)</li>
              <li>أضف <code class="bg-blue-100 dark:bg-blue-800 px-1 rounded text-xs">{{ redirectURI }}</code> ضمن Authorized redirect URIs</li>
              <li>انسخ Client ID و Client Secret وألصقهما أدناه</li>
            </ol>
          </div>

          <UFormGroup label="Client ID">
            <UInput v-model="driveForm.client_id" placeholder="xxxxx.apps.googleusercontent.com" dir="ltr" />
          </UFormGroup>

          <UFormGroup label="Client Secret">
            <UInput v-model="driveForm.client_secret" type="password" placeholder="GOCSPX-xxxxx" dir="ltr" />
          </UFormGroup>

          <UFormGroup label="معرف مجلد Drive (اختياري)">
            <UInput v-model="driveForm.folder_id" placeholder="مثال: 1PHRruHdL6gelbM4NCNz7fgKGqMxw50_R" dir="ltr" />
            <p class="text-xs text-gray-500 mt-1">معرف المجلد من رابط Google Drive. اتركه فارغاً للرفع في الجذر</p>
          </UFormGroup>

          <div class="flex gap-2">
            <UButton @click="saveAndConnect" :loading="driveSaving" icon="i-heroicons-link" color="blue">
              حفظ وربط Google Drive
            </UButton>
          </div>
        </div>

        <!-- Folder config when connected -->
        <div v-if="driveStatus.connected" class="space-y-3">
          <UFormGroup label="معرف مجلد Drive">
            <div class="flex gap-2">
              <UInput v-model="driveFolderID" placeholder="معرف المجلد" dir="ltr" class="flex-1" />
              <UButton @click="updateFolder" :loading="folderSaving" size="sm">تحديث</UButton>
            </div>
            <p class="text-xs text-gray-500 mt-1">معرف المجلد الذي سيتم رفع الملفات إليه</p>
          </UFormGroup>
        </div>
      </div>
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
const route = useRoute()
const runtimeConfig = useRuntimeConfig()

const settings = ref<Record<string, string>>({})
const roles = ref<Role[]>([])
const permissions = ref<Permission[]>([])
const saving = ref(false)
const savingRole = ref(false)
const showRoleModal = ref(false)
const editingRole = ref<Role | null>(null)

// Drive state
const driveStatus = ref<{ connected: boolean; email: string; folder_id: string }>({ connected: false, email: '', folder_id: '' })
const driveForm = reactive({ client_id: '', client_secret: '', folder_id: '' })
const driveFolderID = ref('')
const driveSaving = ref(false)
const driveDisconnecting = ref(false)
const folderSaving = ref(false)

const apiBase = runtimeConfig.public.apiBase as string || ''
const redirectURI = computed(() => {
  const base = apiBase.replace(/\/api\/v1$/, '')
  return base + '/api/v1/admin/drive/callback'
})

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

// Drive functions
const fetchDriveStatus = async () => {
  try {
    driveStatus.value = await api.get('/admin/drive/status')
    driveFolderID.value = driveStatus.value.folder_id || ''
  } catch {}
}

const saveAndConnect = async () => {
  if (!driveForm.client_id || !driveForm.client_secret) {
    toast.add({ title: 'يرجى إدخال Client ID و Client Secret', color: 'red' })
    return
  }
  driveSaving.value = true
  try {
    await api.post('/admin/drive/credentials', driveForm)
    const res = await api.get<{ url: string }>('/admin/drive/auth-url')
    window.location.href = res.url
  } catch (e: any) {
    toast.add({ title: e?.message || 'خطأ في الربط', color: 'red' })
  } finally {
    driveSaving.value = false
  }
}

const disconnectDrive = async () => {
  driveDisconnecting.value = true
  try {
    await api.post('/admin/drive/disconnect')
    driveStatus.value = { connected: false, email: '', folder_id: '' }
    toast.add({ title: 'تم قطع الاتصال', color: 'green' })
  } catch {
    toast.add({ title: 'خطأ', color: 'red' })
  } finally {
    driveDisconnecting.value = false
  }
}

const updateFolder = async () => {
  if (!driveFolderID.value) return
  folderSaving.value = true
  try {
    await api.put('/admin/drive/folder', { folder_id: driveFolderID.value })
    toast.add({ title: 'تم تحديث المجلد', color: 'green' })
  } catch {
    toast.add({ title: 'خطأ', color: 'red' })
  } finally {
    folderSaving.value = false
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

onMounted(async () => {
  await Promise.all([fetchData(), fetchDriveStatus()])
  // Show success if redirected from OAuth callback
  if (route.query.drive === 'connected') {
    toast.add({ title: 'تم ربط Google Drive بنجاح', color: 'green' })
    await fetchDriveStatus()
  }
})
</script>
