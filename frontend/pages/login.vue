<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950 px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-primary-600 dark:text-primary-400">نظام الأرشفة الإلكتروني</h1>
        <p class="text-gray-500 mt-2">قسم ضمان الجودة وتقييم الأداء</p>
      </div>

      <UCard>
        <template #header>
          <h2 class="text-xl font-semibold text-center">تسجيل الدخول</h2>
        </template>

        <form @submit.prevent="handleLogin" class="space-y-4">
          <UFormGroup label="البريد الإلكتروني" required>
            <UInput
              v-model="form.email"
              type="email"
              placeholder="admin@university.edu.iq"
              icon="i-heroicons-envelope"
              size="lg"
            />
          </UFormGroup>

          <UFormGroup label="كلمة المرور" required>
            <UInput
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="أدخل كلمة المرور"
              icon="i-heroicons-lock-closed"
              size="lg"
              :ui="{ icon: { trailing: { pointer: '' } } }"
            >
              <template #trailing>
                <UButton
                  :icon="showPassword ? 'i-heroicons-eye-slash' : 'i-heroicons-eye'"
                  variant="link"
                  :padded="false"
                  @click="showPassword = !showPassword"
                />
              </template>
            </UInput>
          </UFormGroup>

          <UAlert v-if="error" color="red" variant="soft" :title="error" icon="i-heroicons-exclamation-triangle" />

          <UButton type="submit" block size="lg" :loading="loading">
            تسجيل الدخول
          </UButton>

          <UDivider label="أو" />

          <UButton
            block
            variant="outline"
            size="lg"
            icon="i-heroicons-globe-alt"
            @click="loginWithGoogle"
          >
            الدخول بحساب Google
          </UButton>

          <p class="text-center text-sm text-gray-500">
            ليس لديك حساب؟
            <NuxtLink to="/register" class="text-primary-600 hover:underline">إنشاء حساب</NuxtLink>
          </p>
        </form>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const authStore = useAuthStore()

const form = reactive({
  email: '',
  password: '',
})
const showPassword = ref(false)
const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  error.value = ''
  loading.value = true
  try {
    await authStore.login(form.email, form.password)
    navigateTo('/')
  } catch (e: any) {
    error.value = e?.data?.error || 'خطأ في تسجيل الدخول'
  } finally {
    loading.value = false
  }
}

const loginWithGoogle = () => {
  const config = useRuntimeConfig()
  window.location.href = `${config.public.apiBase}/auth/google/callback`
}
</script>
