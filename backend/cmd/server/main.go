package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/haydary1986/archiving-qa/internal/config"
	"github.com/haydary1986/archiving-qa/internal/database"
	"github.com/haydary1986/archiving-qa/internal/middleware"
	"github.com/haydary1986/archiving-qa/internal/routes"
	"github.com/haydary1986/archiving-qa/internal/workers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create default super admin if not exists
	createDefaultAdmin(db, cfg)

	// Start background worker (if not in worker-only mode)
	if os.Getenv("RUN_WORKER") == "true" {
		workerServer := workers.NewWorkerServer(db.DB, cfg)
		log.Fatal(workerServer.Start())
		return
	}

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS
	r.Use(middleware.CORS())

	// Rate limiting
	r.Use(middleware.RateLimit())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "archiving-qa-api",
		})
	})

	// Setup routes
	routes.Setup(r, db.DB, cfg)

	// Start server
	addr := ":" + cfg.Server.Port
	log.Printf("Server starting on %s", addr)
	log.Printf("Mode: %s", cfg.Server.Mode)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createDefaultAdmin(db *database.DB, cfg *config.Config) {
	var exists bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = 'admin@university.edu.iq')").Scan(&exists)
	if exists {
		return
	}

	// Create default admin
	// Password: Admin@123456 (should be changed on first login)
	_, err := db.Exec(`
		INSERT INTO users (email, password_hash, full_name, provider, role_id, is_active)
		VALUES ('admin@university.edu.iq',
		        '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
		        'مدير النظام', 'local',
		        'a0000000-0000-0000-0000-000000000001', true)
	`)
	if err != nil {
		log.Printf("Note: Default admin may already exist: %v", err)
	} else {
		log.Println("Default admin created: admin@university.edu.iq / Admin@123456")
	}
}
