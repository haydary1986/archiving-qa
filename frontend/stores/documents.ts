import { defineStore } from 'pinia'
import type { Document, DocumentFilter, PaginatedResponse, Category, Tag, Person } from '~/types'

export const useDocumentStore = defineStore('documents', {
  state: () => ({
    documents: [] as Document[],
    currentDocument: null as Document | null,
    total: 0,
    page: 1,
    pageSize: 20,
    totalPages: 0,
    loading: false,
    categories: [] as Category[],
    tags: [] as Tag[],
    persons: [] as Person[],
    filters: {} as DocumentFilter,
  }),

  actions: {
    async fetchDocuments(filters?: DocumentFilter) {
      this.loading = true
      try {
        const api = useApi()
        const params = { ...this.filters, ...filters }
        const response = await api.get<PaginatedResponse<Document>>('/documents', params)
        this.documents = response.data
        this.total = response.total
        this.page = response.page
        this.pageSize = response.page_size
        this.totalPages = response.total_pages
      } finally {
        this.loading = false
      }
    },

    async fetchDocument(id: string) {
      this.loading = true
      try {
        const api = useApi()
        this.currentDocument = await api.get<Document>(`/documents/${id}`)
        return this.currentDocument
      } finally {
        this.loading = false
      }
    },

    async createDocument(data: any) {
      const api = useApi()
      const doc = await api.post<Document>('/documents', data)
      this.documents.unshift(doc)
      return doc
    },

    async updateDocument(id: string, data: any) {
      const api = useApi()
      await api.put(`/documents/${id}`, data)
      await this.fetchDocument(id)
    },

    async deleteDocument(id: string) {
      const api = useApi()
      await api.delete(`/documents/${id}`)
      this.documents = this.documents.filter(d => d.id !== id)
    },

    async restoreDocument(id: string) {
      const api = useApi()
      await api.post(`/documents/${id}/restore`)
    },

    async uploadFile(documentId: string, file: File) {
      const api = useApi()
      const formData = new FormData()
      formData.append('file', file)
      return await api.upload(`/documents/${documentId}/files`, formData)
    },

    async fetchCategories() {
      const api = useApi()
      this.categories = await api.get<Category[]>('/categories')
    },

    async fetchTags() {
      const api = useApi()
      this.tags = await api.get<Tag[]>('/tags')
    },

    async fetchPersons(search?: string) {
      const api = useApi()
      this.persons = await api.get<Person[]>('/persons', { search })
    },

    setFilters(filters: DocumentFilter) {
      this.filters = { ...this.filters, ...filters }
    },

    resetFilters() {
      this.filters = {}
    },
  },
})
