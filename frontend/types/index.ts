export interface User {
  id: string
  email: string
  full_name: string
  avatar_url?: string
  provider: 'local' | 'google'
  is_active: boolean
  role_id: string
  role?: Role
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface Role {
  id: string
  name: string
  description: string
  permissions?: Permission[]
  created_at: string
}

export interface Permission {
  id: string
  name: string
  resource: string
  action: string
  description: string
}

export interface Document {
  id: string
  title: string
  description?: string
  document_number?: string
  document_date?: string
  document_type: 'incoming' | 'outgoing' | 'internal'
  category_id?: string
  category?: Category
  classification: 'normal' | 'confidential' | 'secret'
  source_entity?: string
  dest_entity?: string
  custom_fields?: Record<string, any>
  ocr_text?: string
  ai_extracted?: AIExtracted
  status: 'draft' | 'processing' | 'completed' | 'archived'
  created_by_id: string
  created_by?: User
  files?: FileRecord[]
  persons?: Person[]
  routings?: Routing[]
  tags?: Tag[]
  created_at: string
  updated_at: string
  deleted_at?: string
}

export interface FileRecord {
  id: string
  document_id: string
  file_name: string
  original_name: string
  mime_type: string
  file_size: number
  drive_file_id: string
  drive_url?: string
  thumbnail_url?: string
  ocr_text?: string
  ocr_status: 'pending' | 'processing' | 'completed' | 'failed'
  uploaded_by_id: string
  created_at: string
}

export interface Category {
  id: string
  name: string
  parent_id?: string
  parent?: Category
  children?: Category[]
  sort_order: number
  created_at: string
}

export interface Tag {
  id: string
  name: string
  color: string
  created_at: string
}

export interface Person {
  id: string
  full_name: string
  title?: string
  department?: string
  email?: string
  phone?: string
  person_type: 'academic' | 'employee' | 'external'
  created_at: string
}

export interface Routing {
  id: string
  document_id: string
  from_entity: string
  to_entity: string
  action: 'referred' | 'forwarded' | 'returned' | 'archived'
  notes?: string
  action_date: string
  action_by_id: string
  action_by?: User
  created_at: string
}

export interface AuditLog {
  id: string
  user_id: string
  user_name?: string
  user_email?: string
  action: string
  resource: string
  resource_id?: string
  details?: Record<string, any>
  ip_address: string
  user_agent?: string
  created_at: string
}

export interface AIExtracted {
  عنوان_الكتاب?: string
  الجهة_المصدرة?: string
  رقم_العدد?: string
  تاريخ_الكتاب?: string
  ملخص?: string
}

export interface CustomFieldDef {
  id: string
  name: string
  label: string
  field_type: 'text' | 'number' | 'date' | 'select' | 'multiselect' | 'boolean'
  options?: string[]
  required: boolean
  category_id?: string
  sort_order: number
}

export interface ShareLink {
  id: string
  document_id: string
  token: string
  expires_at?: string
  max_views: number
  view_count: number
  is_active: boolean
  created_at: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface DashboardStats {
  total_documents: number
  total_files: number
  total_users: number
  total_persons: number
  documents_by_type: Record<string, number>
  documents_by_classification: Record<string, number>
  recent_activity: ActivityEntry[]
  total_storage_bytes: number
  total_storage_mb: number
}

export interface ActivityEntry {
  action: string
  resource: string
  user_name: string
  created_at: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: User
}

export interface DocumentFilter {
  search?: string
  document_type?: string
  category_id?: string
  classification?: string
  status?: string
  date_from?: string
  date_to?: string
  person_id?: string
  tag_id?: string
  page?: number
  page_size?: number
  sort_by?: string
  sort_order?: string
}
