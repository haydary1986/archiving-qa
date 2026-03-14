package middleware

import (
	"database/sql"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RBACMiddleware struct {
	db *sql.DB
}

func NewRBACMiddleware(db *sql.DB) *RBACMiddleware {
	return &RBACMiddleware{db: db}
}

func (m *RBACMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID := c.GetString("role_id")
		if roleID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "التوثيق مطلوب"})
			return
		}

		var count int
		err := m.db.QueryRow(`
			SELECT COUNT(*) FROM role_permissions rp
			JOIN permissions p ON rp.permission_id = p.id
			WHERE rp.role_id = $1 AND p.resource = $2 AND p.action = $3
		`, roleID, resource, action).Scan(&count)

		if err != nil || count == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "ليس لديك صلاحية لتنفيذ هذا الإجراء"})
			return
		}

		c.Next()
	}
}

func (m *RBACMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		roleName := c.GetString("role_name")
		if _, ok := allowed[roleName]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "دورك لا يملك صلاحية الوصول"})
			return
		}
		c.Next()
	}
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "43200")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RateLimit middleware - per IP
func RateLimit() gin.HandlerFunc {
	var mu sync.Mutex
	type entry struct {
		count int
		start time.Time
	}
	clients := make(map[string]*entry)
	window := time.Minute
	maxReqs := 100

	go func() {
		for range time.Tick(window) {
			mu.Lock()
			now := time.Now()
			for ip, e := range clients {
				if now.Sub(e.start) > window {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Skip rate limiting for health checks
		if strings.HasPrefix(c.Request.URL.Path, "/health") {
			c.Next()
			return
		}

		mu.Lock()
		now := time.Now()
		e, exists := clients[ip]
		if !exists || now.Sub(e.start) > window {
			clients[ip] = &entry{count: 1, start: now}
			mu.Unlock()
			c.Next()
			return
		}
		e.count++
		if e.count > maxReqs {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "عدد الطلبات تجاوز الحد المسموح"})
			return
		}
		mu.Unlock()
		c.Next()
	}
}
