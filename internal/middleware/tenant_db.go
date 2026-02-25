package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hysp/hyadmin-api/internal/database"
	"gorm.io/gorm"
)

const tenantDBKey = "tenantDB"

// TenantDBMiddleware resolves the *gorm.DB for the current tenant and stores
// it in the gin context. Must run after TenantMiddleware (requires TenantKey).
func TenantDBMiddleware(mgr *database.DBManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCode := c.GetString(TenantKey)
		if tenantCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing tenant"})
			c.Abort()
			return
		}

		db, err := mgr.GetDB(tenantCode)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "tenant database unavailable"})
			c.Abort()
			return
		}

		// WithContext ensures GORM respects request cancellation.
		c.Set(tenantDBKey, db.WithContext(c.Request.Context()))
		c.Next()
	}
}

// GetTenantDB retrieves the tenant *gorm.DB injected by TenantDBMiddleware.
// Returns nil if not present (route not protected by TenantDBMiddleware).
func GetTenantDB(c *gin.Context) *gorm.DB {
	val, exists := c.Get(tenantDBKey)
	if !exists {
		return nil
	}
	db, _ := val.(*gorm.DB)
	return db
}
