package database

const migrationCreateExtensions = `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
`

const migrationCreateRolesTable = `
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
`

const migrationCreatePermissionsTable = `
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(resource, action)
);
`

const migrationCreateRolePermissionsTable = `
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
`

const migrationCreateUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) DEFAULT '',
    full_name VARCHAR(255) NOT NULL,
    avatar_url TEXT DEFAULT '',
    provider VARCHAR(50) NOT NULL DEFAULT 'local',
    google_id VARCHAR(255) DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    role_id UUID NOT NULL REFERENCES roles(id),
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id) WHERE google_id != '';
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
`

const migrationCreateCategoriesTable = `
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories(deleted_at);
`

const migrationCreateDocumentsTable = `
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    description TEXT DEFAULT '',
    document_number VARCHAR(100) DEFAULT '',
    document_date TIMESTAMPTZ,
    document_type VARCHAR(50) NOT NULL DEFAULT 'internal',
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    classification VARCHAR(50) NOT NULL DEFAULT 'normal',
    source_entity VARCHAR(255) DEFAULT '',
    dest_entity VARCHAR(255) DEFAULT '',
    custom_fields JSONB DEFAULT '{}',
    ocr_text TEXT DEFAULT '',
    ai_extracted JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_by_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_documents_type ON documents(document_type);
CREATE INDEX IF NOT EXISTS idx_documents_category ON documents(category_id);
CREATE INDEX IF NOT EXISTS idx_documents_classification ON documents(classification);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_created_by ON documents(created_by_id);
CREATE INDEX IF NOT EXISTS idx_documents_deleted_at ON documents(deleted_at);
CREATE INDEX IF NOT EXISTS idx_documents_date ON documents(document_date);
CREATE INDEX IF NOT EXISTS idx_documents_number ON documents(document_number);
CREATE INDEX IF NOT EXISTS idx_documents_custom_fields ON documents USING GIN(custom_fields);
CREATE INDEX IF NOT EXISTS idx_documents_title_trgm ON documents USING GIN(title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_documents_ocr_trgm ON documents USING GIN(ocr_text gin_trgm_ops);
`

const migrationCreateFilesTable = `
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    file_name VARCHAR(500) NOT NULL,
    original_name VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    drive_file_id VARCHAR(255) DEFAULT '',
    drive_url TEXT DEFAULT '',
    thumbnail_url TEXT DEFAULT '',
    ocr_text TEXT DEFAULT '',
    ocr_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    uploaded_by_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_files_document ON files(document_id);
CREATE INDEX IF NOT EXISTS idx_files_deleted_at ON files(deleted_at);
`

const migrationCreateTagsTable = `
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    color VARCHAR(7) DEFAULT '#3B82F6',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
`

const migrationCreateDocumentTagsTable = `
CREATE TABLE IF NOT EXISTS document_tags (
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (document_id, tag_id)
);
`

const migrationCreatePersonsTable = `
CREATE TABLE IF NOT EXISTS persons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name VARCHAR(255) NOT NULL,
    title VARCHAR(255) DEFAULT '',
    department VARCHAR(255) DEFAULT '',
    email VARCHAR(255) DEFAULT '',
    phone VARCHAR(50) DEFAULT '',
    person_type VARCHAR(50) NOT NULL DEFAULT 'employee',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_persons_name_trgm ON persons USING GIN(full_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_persons_deleted_at ON persons(deleted_at);
`

const migrationCreateDocumentPersonsTable = `
CREATE TABLE IF NOT EXISTS document_persons (
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    relation VARCHAR(50) NOT NULL DEFAULT 'related',
    PRIMARY KEY (document_id, person_id)
);
`

const migrationCreateRoutingsTable = `
CREATE TABLE IF NOT EXISTS routings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    from_entity VARCHAR(255) NOT NULL,
    to_entity VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    notes TEXT DEFAULT '',
    action_date TIMESTAMPTZ NOT NULL,
    action_by_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_routings_document ON routings(document_id);
`

const migrationCreateAuditLogTable = `
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(45) DEFAULT '',
    user_agent TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_logs(resource, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);
`

const migrationCreateCustomFieldDefsTable = `
CREATE TABLE IF NOT EXISTS custom_field_defs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    label VARCHAR(255) NOT NULL,
    field_type VARCHAR(50) NOT NULL,
    options JSONB DEFAULT '[]',
    required BOOLEAN NOT NULL DEFAULT false,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
`

const migrationCreateSystemSettingsTable = `
CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationCreateShareLinksTable = `
CREATE TABLE IF NOT EXISTS share_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) DEFAULT '',
    expires_at TIMESTAMPTZ,
    max_views INT NOT NULL DEFAULT 0,
    view_count INT NOT NULL DEFAULT 0,
    created_by_id UUID NOT NULL REFERENCES users(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_share_links_token ON share_links(token);
CREATE INDEX IF NOT EXISTS idx_share_links_document ON share_links(document_id);
`

const migrationCreateUserCategoryAccess = `
CREATE TABLE IF NOT EXISTS user_category_access (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_user_cat_access_user ON user_category_access(user_id);
`

const migrationCreateJobsTable = `
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    file_id UUID REFERENCES files(id) ON DELETE CASCADE,
    payload JSONB DEFAULT '{}',
    result JSONB DEFAULT '{}',
    error_message TEXT DEFAULT '',
    attempts INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs(task_type);
CREATE INDEX IF NOT EXISTS idx_jobs_document ON jobs(document_id);
CREATE INDEX IF NOT EXISTS idx_jobs_created ON jobs(created_at DESC);
`

const migrationSeedDefaultData = `
-- Seed default roles
INSERT INTO roles (id, name, description) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'super_admin', 'مدير النظام - صلاحيات كاملة على جميع أجزاء النظام'),
    ('a0000000-0000-0000-0000-000000000002', 'qa_manager', 'مدير الجودة - إدارة الوثائق والمستخدمين والتقارير'),
    ('a0000000-0000-0000-0000-000000000003', 'data_entry', 'مدخل بيانات - إضافة وتعديل الوثائق ورفع الملفات'),
    ('a0000000-0000-0000-0000-000000000004', 'viewer', 'مشاهد - قراءة فقط (يمكن تقييده بأقسام محددة)')
ON CONFLICT (name) DO NOTHING;

-- Seed default permissions
INSERT INTO permissions (name, resource, action, description) VALUES
    ('documents.create', 'documents', 'create', 'إنشاء وثائق جديدة'),
    ('documents.read', 'documents', 'read', 'عرض الوثائق'),
    ('documents.update', 'documents', 'update', 'تعديل الوثائق'),
    ('documents.delete', 'documents', 'delete', 'حذف الوثائق'),
    ('documents.export', 'documents', 'export', 'تصدير الوثائق'),
    ('files.create', 'files', 'create', 'رفع الملفات'),
    ('files.read', 'files', 'read', 'تنزيل الملفات'),
    ('files.delete', 'files', 'delete', 'حذف الملفات'),
    ('users.create', 'users', 'create', 'إنشاء مستخدمين'),
    ('users.read', 'users', 'read', 'عرض المستخدمين'),
    ('users.update', 'users', 'update', 'تعديل المستخدمين'),
    ('users.delete', 'users', 'delete', 'حذف المستخدمين'),
    ('categories.manage', 'categories', 'admin', 'إدارة التصنيفات'),
    ('audit.read', 'audit', 'read', 'عرض سجل التدقيق'),
    ('settings.manage', 'settings', 'admin', 'إدارة إعدادات النظام'),
    ('persons.manage', 'persons', 'admin', 'إدارة الأشخاص'),
    ('share.manage', 'share', 'admin', 'إدارة روابط المشاركة')
ON CONFLICT (name) DO NOTHING;

-- Assign all permissions to super_admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'a0000000-0000-0000-0000-000000000001', id FROM permissions
ON CONFLICT DO NOTHING;

-- Assign permissions to qa_manager (everything except system settings)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'a0000000-0000-0000-0000-000000000002', id FROM permissions
WHERE name NOT IN ('settings.manage')
ON CONFLICT DO NOTHING;

-- Assign data_entry permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'a0000000-0000-0000-0000-000000000003', id FROM permissions
WHERE name IN ('documents.create', 'documents.read', 'documents.update', 'documents.export', 'files.create', 'files.read', 'persons.manage')
ON CONFLICT DO NOTHING;

-- Assign viewer permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT 'a0000000-0000-0000-0000-000000000004', id FROM permissions
WHERE name IN ('documents.read', 'files.read')
ON CONFLICT DO NOTHING;

-- Seed default system settings
INSERT INTO system_settings (key, value) VALUES
    ('local_auth_enabled', 'true'),
    ('google_auth_enabled', 'true'),
    ('max_file_size_mb', '50'),
    ('allowed_domains', '@university.edu.iq'),
    ('ocr_enabled', 'true'),
    ('ai_analysis_enabled', 'true'),
    ('system_name', 'نظام الأرشفة الإلكتروني - قسم ضمان الجودة وتقييم الأداء'),
    ('system_name_en', 'QA Archiving System')
ON CONFLICT (key) DO NOTHING;
`
