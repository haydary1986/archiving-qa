package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ==================== Base Model ====================

type BaseModel struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ==================== User & Auth ====================

type User struct {
	BaseModel
	Email        string `json:"email" db:"email"`
	PasswordHash string `json:"-" db:"password_hash"`
	FullName     string `json:"full_name" db:"full_name"`
	AvatarURL    string `json:"avatar_url,omitempty" db:"avatar_url"`
	Provider     string `json:"provider" db:"provider"` // "local" or "google"
	GoogleID     string `json:"-" db:"google_id"`
	IsActive     bool   `json:"is_active" db:"is_active"`
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	Role         *Role  `json:"role,omitempty"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

type Role struct {
	BaseModel
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Permissions []Permission `json:"permissions,omitempty"`
}

type Permission struct {
	BaseModel
	Name        string `json:"name" db:"name"`
	Resource    string `json:"resource" db:"resource"`
	Action      string `json:"action" db:"action"` // "create", "read", "update", "delete", "export", "admin"
	Description string `json:"description" db:"description"`
}

type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID uuid.UUID `json:"permission_id" db:"permission_id"`
}

// ==================== Documents ====================

type Document struct {
	BaseModel
	Title          string          `json:"title" db:"title"`
	Description    string          `json:"description,omitempty" db:"description"`
	DocumentNumber string          `json:"document_number,omitempty" db:"document_number"`
	DocumentDate   *time.Time      `json:"document_date,omitempty" db:"document_date"`
	DocumentType   string          `json:"document_type" db:"document_type"` // "incoming", "outgoing", "internal"
	CategoryID     *uuid.UUID      `json:"category_id,omitempty" db:"category_id"`
	Category       *Category       `json:"category,omitempty"`
	Classification string          `json:"classification" db:"classification"` // "normal", "confidential", "secret"
	SourceEntity   string          `json:"source_entity,omitempty" db:"source_entity"`
	DestEntity     string          `json:"dest_entity,omitempty" db:"dest_entity"`
	CustomFields   json.RawMessage `json:"custom_fields,omitempty" db:"custom_fields"`
	OCRText        string          `json:"ocr_text,omitempty" db:"ocr_text"`
	AIExtracted    json.RawMessage `json:"ai_extracted,omitempty" db:"ai_extracted"`
	Status         string          `json:"status" db:"status"` // "draft", "processing", "completed", "archived"
	CreatedByID    uuid.UUID       `json:"created_by_id" db:"created_by_id"`
	CreatedBy      *User           `json:"created_by,omitempty"`
	Files          []File          `json:"files,omitempty"`
	Persons        []Person        `json:"persons,omitempty"`
	Routings       []Routing       `json:"routings,omitempty"`
	Tags           []Tag           `json:"tags,omitempty"`
}

type Category struct {
	BaseModel
	Name     string     `json:"name" db:"name"`
	ParentID *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Parent   *Category  `json:"parent,omitempty"`
	Children []Category `json:"children,omitempty"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
}

type Tag struct {
	BaseModel
	Name  string `json:"name" db:"name"`
	Color string `json:"color,omitempty" db:"color"`
}

type DocumentTag struct {
	DocumentID uuid.UUID `json:"document_id" db:"document_id"`
	TagID      uuid.UUID `json:"tag_id" db:"tag_id"`
}

// ==================== Files ====================

type File struct {
	BaseModel
	DocumentID   uuid.UUID `json:"document_id" db:"document_id"`
	FileName     string    `json:"file_name" db:"file_name"`
	OriginalName string    `json:"original_name" db:"original_name"`
	MimeType     string    `json:"mime_type" db:"mime_type"`
	FileSize     int64     `json:"file_size" db:"file_size"`
	DriveFileID  string    `json:"drive_file_id" db:"drive_file_id"`
	DriveURL     string    `json:"drive_url,omitempty" db:"drive_url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
	OCRText      string    `json:"ocr_text,omitempty" db:"ocr_text"`
	OCRStatus    string    `json:"ocr_status" db:"ocr_status"` // "pending", "processing", "completed", "failed"
	UploadedByID uuid.UUID `json:"uploaded_by_id" db:"uploaded_by_id"`
}

// ==================== Persons ====================

type Person struct {
	BaseModel
	FullName    string `json:"full_name" db:"full_name"`
	Title       string `json:"title,omitempty" db:"title"`
	Department  string `json:"department,omitempty" db:"department"`
	Email       string `json:"email,omitempty" db:"email"`
	Phone       string `json:"phone,omitempty" db:"phone"`
	PersonType  string `json:"person_type" db:"person_type"` // "academic", "employee", "external"
}

type DocumentPerson struct {
	DocumentID uuid.UUID `json:"document_id" db:"document_id"`
	PersonID   uuid.UUID `json:"person_id" db:"person_id"`
	Relation   string    `json:"relation" db:"relation"` // "sender", "receiver", "related", "cc"
}

// ==================== Routing ====================

type Routing struct {
	BaseModel
	DocumentID  uuid.UUID  `json:"document_id" db:"document_id"`
	FromEntity  string     `json:"from_entity" db:"from_entity"`
	ToEntity    string     `json:"to_entity" db:"to_entity"`
	Action      string     `json:"action" db:"action"` // "referred", "forwarded", "returned", "archived"
	Notes       string     `json:"notes,omitempty" db:"notes"`
	ActionDate  time.Time  `json:"action_date" db:"action_date"`
	ActionByID  uuid.UUID  `json:"action_by_id" db:"action_by_id"`
	ActionBy    *User      `json:"action_by,omitempty"`
}

// ==================== Audit Log ====================

type AuditLog struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	UserID     uuid.UUID       `json:"user_id" db:"user_id"`
	User       *User           `json:"user,omitempty"`
	Action     string          `json:"action" db:"action"` // "create", "update", "delete", "download", "view", "share", "login"
	Resource   string          `json:"resource" db:"resource"` // "document", "file", "user", "category", etc.
	ResourceID *uuid.UUID      `json:"resource_id,omitempty" db:"resource_id"`
	Details    json.RawMessage `json:"details,omitempty" db:"details"`
	IPAddress  string          `json:"ip_address" db:"ip_address"`
	UserAgent  string          `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// ==================== Custom Field Definition ====================

type CustomFieldDef struct {
	BaseModel
	Name         string          `json:"name" db:"name"`
	Label        string          `json:"label" db:"label"`
	FieldType    string          `json:"field_type" db:"field_type"` // "text", "number", "date", "select", "multiselect", "boolean"
	Options      json.RawMessage `json:"options,omitempty" db:"options"` // for select/multiselect
	Required     bool            `json:"required" db:"required"`
	CategoryID   *uuid.UUID      `json:"category_id,omitempty" db:"category_id"` // if null, applies to all
	SortOrder    int             `json:"sort_order" db:"sort_order"`
}

// ==================== System Settings ====================

type SystemSetting struct {
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ==================== Share Links ====================

type ShareLink struct {
	BaseModel
	DocumentID   uuid.UUID  `json:"document_id" db:"document_id"`
	Token        string     `json:"token" db:"token"`
	PasswordHash string     `json:"-" db:"password_hash"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	MaxViews     int        `json:"max_views" db:"max_views"`
	ViewCount    int        `json:"view_count" db:"view_count"`
	CreatedByID  uuid.UUID  `json:"created_by_id" db:"created_by_id"`
	IsActive     bool       `json:"is_active" db:"is_active"`
}
