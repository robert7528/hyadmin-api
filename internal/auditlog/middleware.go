package auditlog

import (
	"github.com/gin-gonic/gin"
	"github.com/hysp/hyadmin-api/internal/middleware"
	"gorm.io/gorm"
)

// AuditMiddleware records POST/PUT/DELETE actions automatically.
func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" {
			c.Next()
			return
		}
		claims := middleware.GetClaims(c)
		if claims == nil {
			c.Next()
			return
		}

		actionMap := map[string]string{
			"POST":   "CREATE",
			"PUT":    "UPDATE",
			"DELETE": "DELETE",
		}

		c.Next() // execute handler first

		log := &AuditLog{
			TenantCode: claims.TenantCode,
			UserID:     claims.UserID,
			Username:   claims.Username,
			Action:     actionMap[method],
			Resource:   c.FullPath(),
			ResourceID: c.Param("id"),
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
		}
		db.Create(log) // best-effort; ignore error
	}
}
