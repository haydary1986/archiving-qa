export default defineNuxtConfig({
  devtools: { enabled: true },

  modules: [
    '@nuxt/ui',
    '@pinia/nuxt',
    '@nuxtjs/color-mode',
    '@vueuse/nuxt',
  ],

  colorMode: {
    preference: 'light',
    classSuffix: '',
  },

  ui: {
    icons: ['heroicons', 'lucide'],
  },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'https://api-qa.uoturath.edu.iq/api/v1',
    },
  },

  app: {
    head: {
      title: 'نظام الأرشفة الإلكتروني - ضمان الجودة',
      htmlAttrs: { lang: 'ar', dir: 'rtl' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'نظام الأرشفة الإلكتروني لقسم ضمان الجودة وتقييم الأداء' },
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=IBM+Plex+Sans+Arabic:wght@300;400;500;600;700&display=swap' },
      ],
    },
  },

  css: ['~/assets/css/main.css'],

  compatibilityDate: '2024-04-01',
})
