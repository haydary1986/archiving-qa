<template>
  <div class="relative" ref="wrapper">
    <UInput
      v-model="searchQuery"
      :placeholder="placeholder"
      @focus="showDropdown = true"
      @input="onInput"
      autocomplete="off"
    >
      <template #trailing>
        <UIcon
          v-if="modelValue"
          name="i-heroicons-x-mark"
          class="w-4 h-4 cursor-pointer text-gray-400 hover:text-gray-600"
          @click.stop="clearSelection"
        />
        <UIcon
          v-else
          name="i-heroicons-chevron-down"
          class="w-4 h-4 text-gray-400"
        />
      </template>
    </UInput>

    <!-- Dropdown -->
    <div
      v-if="showDropdown"
      class="absolute z-50 w-full mt-1 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg max-h-60 overflow-y-auto"
    >
      <!-- Filtered options -->
      <button
        v-for="option in filteredOptions"
        :key="option.value"
        type="button"
        class="w-full text-right px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-800 text-sm transition-colors"
        :class="{ 'bg-primary-50 dark:bg-primary-900/20 text-primary-600': option.value === modelValue }"
        @click="selectOption(option)"
      >
        {{ option.label }}
      </button>

      <!-- No results + Add new -->
      <div v-if="!filteredOptions.length && searchQuery.trim()" class="p-2">
        <button
          type="button"
          class="w-full text-right px-3 py-2 hover:bg-green-50 dark:hover:bg-green-900/20 text-sm rounded-md flex items-center gap-2 text-green-600"
          @click="handleAddNew"
        >
          <UIcon name="i-heroicons-plus-circle" class="w-4 h-4" />
          <span>إضافة "{{ searchQuery.trim() }}"</span>
        </button>
      </div>

      <!-- Add new option even when results exist -->
      <div v-else-if="searchQuery.trim() && !exactMatch" class="border-t border-gray-200 dark:border-gray-700 p-2">
        <button
          type="button"
          class="w-full text-right px-3 py-2 hover:bg-green-50 dark:hover:bg-green-900/20 text-sm rounded-md flex items-center gap-2 text-green-600"
          @click="handleAddNew"
        >
          <UIcon name="i-heroicons-plus-circle" class="w-4 h-4" />
          <span>إضافة "{{ searchQuery.trim() }}"</span>
        </button>
      </div>

      <!-- Empty state -->
      <p v-if="!filteredOptions.length && !searchQuery.trim()" class="px-3 py-4 text-center text-sm text-gray-500">
        اكتب للبحث أو إضافة جديد
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Option {
  value: string
  label: string
}

const props = defineProps<{
  modelValue: string
  options: Option[]
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'add-new': [name: string]
}>()

const wrapper = ref<HTMLElement>()
const showDropdown = ref(false)
const searchQuery = ref('')

// Sync display text with selected value
watch(() => props.modelValue, (val) => {
  if (val) {
    const found = props.options.find(o => o.value === val)
    searchQuery.value = found ? found.label : val
  } else {
    searchQuery.value = ''
  }
}, { immediate: true })

// Also watch options changes to update display
watch(() => props.options, () => {
  if (props.modelValue) {
    const found = props.options.find(o => o.value === props.modelValue)
    if (found) searchQuery.value = found.label
  }
})

const filteredOptions = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return props.options
  return props.options.filter(o => o.label.toLowerCase().includes(q))
})

const exactMatch = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  return props.options.some(o => o.label.toLowerCase() === q)
})

const onInput = () => {
  showDropdown.value = true
  // If the user is typing freely (entity fields), emit the raw text as value
  const found = props.options.find(o => o.label === searchQuery.value)
  if (!found) {
    emit('update:modelValue', searchQuery.value)
  }
}

const selectOption = (option: Option) => {
  emit('update:modelValue', option.value)
  searchQuery.value = option.label
  showDropdown.value = false
}

const clearSelection = () => {
  emit('update:modelValue', '')
  searchQuery.value = ''
}

const handleAddNew = () => {
  const name = searchQuery.value.trim()
  if (name) {
    emit('add-new', name)
    emit('update:modelValue', name)
    showDropdown.value = false
  }
}

// Close dropdown when clicking outside
onMounted(() => {
  const handler = (e: MouseEvent) => {
    if (wrapper.value && !wrapper.value.contains(e.target as Node)) {
      showDropdown.value = false
    }
  }
  document.addEventListener('click', handler)
  onUnmounted(() => document.removeEventListener('click', handler))
})
</script>
