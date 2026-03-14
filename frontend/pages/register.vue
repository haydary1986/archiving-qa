<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950 px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-primary-600">نظام الأرشفة الإلكتروني</h1>
      </div>

      <UCard>
        <template #header>
          <h2 class="text-xl font-semibold text-center">إنشاء حساب جديد</h2>
        </template>

        <form @submit.prevent="handleRegister" class="space-y-4">
          <UFormGroup label="الاسم الكامل" required>
            <UInput v-model="form.fullName" placeholder="الاسم الثلاثي" icon="i-heroicons-user" size="lg" />
          </UFormGroup>

          <UFormGroup label="البريد الإلكتروني" required>
            <UInput v-model="form.email" type="email" placeholder="example@university.edu.iq"
                    icon="i-heroicons-envelope" size="lg" />
          </UFormGroup>

          <UFormGroup label="كلمة المرور" required>
            <UInput v-model="form.password" type="password" placeholder="8 أحرف على الأقل"
                    icon="i-heroicons-lock-closed" size="lg" />
          </UFormGroup>

          <UFormGroup label="تأكيد كلمة المرور" required>
            <UInput v-model="form.confirmPassword" type="password" placeholder="أعد كتابة كلمة المرور"
                    icon="i-heroicons-lock-closed" size="lg" />
          </UFormGroup>

          <UAlert v-if="error" color="red" variant="soft" :title="error" />

          <UButton type="submit" block size="lg" :loading="loading">إنشاء الحساب</UButton>

          <p class="text-center text-sm text-gray-500">
            لديك حساب؟
            <NuxtLink to="/login" class="text-primary-600 hover:underline">تسجيل الدخول</NuxtLink>
          </p>
        </form>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

const authStore = useAuthStore()
const form = reactive({ fullName: '', email: '', password: '', confirmPassword: '' })
const loading = ref(false)
const error = ref('')

const handleRegister = async () => {
  if (form.password !== form.confirmPassword) {
    error.value = 'كلمتا المرور غير متطابقتين'
    return
  }
  if (form.password.length < 8) {
    error.value = 'كلمة المرور يجب أن تكون 8 أحرف على الأقل'
    return
  }
  error.value = ''
  loading.value = true
  try {
    await authStore.register(form.email, form.password, form.fullName)
    navigateTo('/')
  } catch (e: any) {
    error.value = e?.data?.error || 'خطأ في إنشاء الحساب'
  } finally {
    loading.value = false
  }
}
</script>
