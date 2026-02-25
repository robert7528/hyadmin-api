package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	TenantHeader = "X-Tenant-ID"
	TenantKey    = "tenant_id"
)

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader(TenantHeader)
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Tenant-ID header"})
			c.Abort()
			return
		}
		c.Set(TenantKey, tenantID)
		c.Next()
	}
}
