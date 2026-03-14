package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ==================== Auth Requests ====================

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required,min=2"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	User         User   `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ==================== Document Requests ====================

type CreateDocumentRequest struct {
	Title          string          `json:"title" binding:"required"`
	Description    string          `json:"description"`
	DocumentNumber string          `json:"document_number"`
	DocumentDate   *time.Time      `json:"document_date"`
	DocumentType   string          `json:"document_type" binding:"required,oneof=incoming outgoing internal"`
	CategoryID     *uuid.UUID      `json:"category_id"`
	Classification string          `json:"classification" binding:"required,oneof=normal confidential secret"`
	SourceEntity   string          `json:"source_entity"`
	DestEntity     string          `json:"dest_entity"`
	CustomFields   json.RawMessage `json:"custom_fields"`
	PersonIDs      []PersonRelation `json:"persons"`
	TagIDs         []uuid.UUID     `json:"tag_ids"`
}

type UpdateDocumentRequest struct {
	Title          *string          `json:"title"`
	Description    *string          `json:"description"`
	DocumentNumber *string          `json:"document_number"`
	DocumentDate   *time.Time       `json:"document_date"`
	DocumentType   *string          `json:"document_type" binding:"omitempty,oneof=incoming outgoing internal"`
	CategoryID     *uuid.UUID       `json:"category_id"`
	Classification *string          `json:"classification" binding:"omitempty,oneof=normal confidential secret"`
	SourceEntity   *string          `json:"source_entity"`
	DestEntity     *string          `json:"dest_entity"`
	CustomFields   json.RawMessage  `json:"custom_fields"`
	PersonIDs      []PersonRelation `json:"persons"`
	TagIDs         []uuid.UUID      `json:"tag_ids"`
}

type PersonRelation struct {
	PersonID uuid.UUID `json:"person_id" binding:"required"`
	Relation string    `json:"relation" binding:"required,oneof=sender receiver related cc"`
}

// ==================== Search & Filter ====================

type DocumentFilter struct {
	Search         string     `form:"search"`
	DocumentType   string     `form:"document_type"`
	CategoryID     *uuid.UUID `form:"category_id"`
	Classification string     `form:"classification"`
	Status         string     `form:"status"`
	DateFrom       *time.Time `form:"date_from"`
	DateTo         *time.Time `form:"date_to"`
	PersonID       *uuid.UUID `form:"person_id"`
	TagID          *uuid.UUID `form:"tag_id"`
	CreatedByID    *uuid.UUID `form:"created_by_id"`
	Page           int        `form:"page,default=1"`
	PageSize       int        `form:"page_size,default=20"`
	SortBy         string     `form:"sort_by,default=created_at"`
	SortOrder      string     `form:"sort_order,default=desc"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// ==================== Routing Requests ====================

type CreateRoutingRequest struct {
	DocumentID uuid.UUID `json:"document_id" binding:"required"`
	FromEntity string    `json:"from_entity" binding:"required"`
	ToEntity   string    `json:"to_entity" binding:"required"`
	Action     string    `json:"action" binding:"required,oneof=referred forwarded returned archived"`
	Notes      string    `json:"notes"`
	ActionDate time.Time `json:"action_date" binding:"required"`
}

// ==================== Category Requests ====================

type CreateCategoryRequest struct {
	Name     string     `json:"name" binding:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
	SortOrder int       `json:"sort_order"`
}

// ==================== Person Requests ====================

type CreatePersonRequest struct {
	FullName   string `json:"full_name" binding:"required"`
	Title      string `json:"title"`
	Department string `json:"department"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	PersonType string `json:"person_type" binding:"required,oneof=academic employee external"`
}

// ==================== Custom Field Requests ====================

type CreateCustomFieldRequest struct {
	Name       string          `json:"name" binding:"required"`
	Label      string          `json:"label" binding:"required"`
	FieldType  string          `json:"field_type" binding:"required,oneof=text number date select multiselect boolean"`
	Options    json.RawMessage `json:"options"`
	Required   bool            `json:"required"`
	CategoryID *uuid.UUID      `json:"category_id"`
	SortOrder  int             `json:"sort_order"`
}

// ==================== Share Link Requests ====================

type CreateShareLinkRequest struct {
	DocumentID uuid.UUID  `json:"document_id" binding:"required"`
	Password   string     `json:"password"`
	ExpiresAt  *time.Time `json:"expires_at"`
	MaxViews   int        `json:"max_views"`
}

// ==================== User Management Requests ====================

type UpdateUserRequest struct {
	FullName string     `json:"full_name"`
	IsActive *bool      `json:"is_active"`
	RoleID   *uuid.UUID `json:"role_id"`
}

type CreateRoleRequest struct {
	Name          string      `json:"name" binding:"required"`
	Description   string      `json:"description"`
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

// ==================== AI Analysis ====================

type AIAnalysisResult struct {
	Title        string `json:"عنوان_الكتاب"`
	Source       string `json:"الجهة_المصدرة"`
	IssueNumber  string `json:"رقم_العدد"`
	DocumentDate string `json:"تاريخ_الكتاب"`
	Summary      string `json:"ملخص,omitempty"`
}

// ==================== Export Request ====================

type ExportRequest struct {
	DocumentIDs []uuid.UUID `json:"document_ids" binding:"required"`
	IncludeFiles bool       `json:"include_files"`
	Format       string     `json:"format" binding:"required,oneof=excel zip"`
}

// ==================== System Settings ====================

type UpdateSettingRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type ToggleLocalAuthRequest struct {
	Enabled bool `json:"enabled"`
}
