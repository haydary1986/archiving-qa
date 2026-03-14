package routes

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"github.com/haydary1986/archiving-qa/internal/config"
	"github.com/haydary1986/archiving-qa/internal/handlers"
	"github.com/haydary1986/archiving-qa/internal/middleware"
	"github.com/haydary1986/archiving-qa/internal/services"
)

func Setup(r *gin.Engine, db *sql.DB, cfg *config.Config) {
	// Services
	driveService := services.NewDriveService(&cfg.Google)
	compressService := services.NewCompressService()

	// Handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	docHandler := handlers.NewDocumentHandler(db, cfg, driveService, compressService)
	catHandler := handlers.NewCategoryHandler(db)
	personHandler := handlers.NewPersonHandler(db)
	userHandler := handlers.NewUserHandler(db)
	roleHandler := handlers.NewRoleHandler(db)
	adminHandler := handlers.NewAdminHandler(db)
	tagHandler := handlers.NewTagHandler(db)
	shareHandler := handlers.NewShareHandler(db)
	exportHandler := handlers.NewExportHandler(db)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)
	rbacMiddleware := middleware.NewRBACMiddleware(db)

	// API v1
	api := r.Group("/api/v1")

	// Public routes
	{
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/refresh", authHandler.RefreshToken)
		api.GET("/auth/google/callback", authHandler.GoogleCallback)

		// Public share access
		api.GET("/share/:token", shareHandler.Access)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(authMiddleware.Authenticate())
	{
		// Profile
		protected.GET("/auth/profile", authHandler.GetProfile)

		// Documents
		docs := protected.Group("/documents")
		{
			docs.GET("", rbacMiddleware.RequirePermission("documents", "read"), docHandler.List)
			docs.GET("/:id", rbacMiddleware.RequirePermission("documents", "read"), docHandler.Get)
			docs.POST("", rbacMiddleware.RequirePermission("documents", "create"), docHandler.Create)
			docs.PUT("/:id", rbacMiddleware.RequirePermission("documents", "update"), docHandler.Update)
			docs.DELETE("/:id", rbacMiddleware.RequirePermission("documents", "delete"), docHandler.Delete)
			docs.POST("/:id/restore", rbacMiddleware.RequireRole("super_admin"), docHandler.Restore)
			docs.POST("/:id/files", rbacMiddleware.RequirePermission("files", "create"), docHandler.UploadFile)
		}

		// Routings
		protected.POST("/routings", rbacMiddleware.RequirePermission("documents", "update"), docHandler.AddRouting)

		// Categories
		cats := protected.Group("/categories")
		{
			cats.GET("", catHandler.List)
			cats.POST("", rbacMiddleware.RequirePermission("categories", "admin"), catHandler.Create)
			cats.PUT("/:id", rbacMiddleware.RequirePermission("categories", "admin"), catHandler.Update)
			cats.DELETE("/:id", rbacMiddleware.RequirePermission("categories", "admin"), catHandler.Delete)
		}

		// Persons
		persons := protected.Group("/persons")
		{
			persons.GET("", personHandler.List)
			persons.GET("/:id", personHandler.Get)
			persons.POST("", rbacMiddleware.RequirePermission("persons", "admin"), personHandler.Create)
			persons.PUT("/:id", rbacMiddleware.RequirePermission("persons", "admin"), personHandler.Update)
			persons.DELETE("/:id", rbacMiddleware.RequirePermission("persons", "admin"), personHandler.Delete)
		}

		// Tags
		tags := protected.Group("/tags")
		{
			tags.GET("", tagHandler.List)
			tags.POST("", rbacMiddleware.RequirePermission("documents", "create"), tagHandler.Create)
			tags.DELETE("/:id", rbacMiddleware.RequirePermission("documents", "delete"), tagHandler.Delete)
		}

		// Share links
		protected.POST("/share", rbacMiddleware.RequirePermission("share", "admin"), shareHandler.Create)

		// Export
		protected.POST("/export", rbacMiddleware.RequirePermission("documents", "export"), exportHandler.Export)

		// Dashboard
		protected.GET("/dashboard", adminHandler.Dashboard)
	}

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(authMiddleware.Authenticate())
	admin.Use(rbacMiddleware.RequireRole("super_admin", "qa_manager"))
	{
		// Users management
		admin.GET("/users", userHandler.List)
		admin.GET("/users/:id", userHandler.Get)
		admin.POST("/users", userHandler.Create)
		admin.PUT("/users/:id", userHandler.Update)
		admin.DELETE("/users/:id", userHandler.Delete)

		// Roles management
		admin.GET("/roles", roleHandler.List)
		admin.POST("/roles", roleHandler.Create)
		admin.PUT("/roles/:id", roleHandler.Update)
		admin.GET("/permissions", roleHandler.ListPermissions)

		// Audit logs
		admin.GET("/audit-logs", adminHandler.ListAuditLogs)

		// System settings
		admin.GET("/settings", adminHandler.GetSettings)
		admin.PUT("/settings", adminHandler.UpdateSetting)

		// Custom fields
		admin.GET("/custom-fields", adminHandler.ListCustomFields)
		admin.POST("/custom-fields", adminHandler.CreateCustomField)
		admin.DELETE("/custom-fields/:id", adminHandler.DeleteCustomField)

		// Trash - super_admin only
		admin.GET("/trash", rbacMiddleware.RequireRole("super_admin"), docHandler.ListTrash)
	}

	// Super admin only routes (restore, permanent actions)
	superAdmin := api.Group("/admin")
	superAdmin.Use(authMiddleware.Authenticate())
	superAdmin.Use(rbacMiddleware.RequireRole("super_admin"))
	{
		superAdmin.POST("/restore/:id", docHandler.Restore)
	}
}
