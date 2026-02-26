package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/hysp/hyadmin-api/internal/auth"
)

// PermCodesLoader is any type that can return permission codes for a user ID.
// Implemented by role.Service.
type PermCodesLoader interface {
	GetPermissionCodesForUser(userID uint) ([]string, error)
}

// PermissionLoaderMiddleware resolves permission codes after JWT auth and stores them in context.
// Must run after AuthMiddleware.
func PermissionLoaderMiddleware(loader PermCodesLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims != nil {
			codes, err := loader.GetPermissionCodesForUser(claims.UserID)
			if err == nil {
				SetPermissionCodes(c, codes)
			}
		}
		c.Next()
	}
}

const (
	claimsKey         = "auth_claims"
	permCodesKey      = "perm_codes"
)

// AuthMiddleware validates JWT Bearer tokens.
// Routes that skip auth should be registered before this middleware.
func AuthMiddleware(authSvc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := authSvc.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// GetClaims retrieves JWT claims from the Gin context.
// Returns nil if not present (unauthenticated route).
func GetClaims(c *gin.Context) *auth.Claims {
	v, exists := c.Get(claimsKey)
	if !exists {
		return nil
	}
	claims, _ := v.(*auth.Claims)
	return claims
}

// SetPermissionCodes stores resolved permission codes in the context.
func SetPermissionCodes(c *gin.Context, codes []string) {
	c.Set(permCodesKey, codes)
}

// GetPermissionCodes retrieves permission codes from context.
func GetPermissionCodes(c *gin.Context) []string {
	v, exists := c.Get(permCodesKey)
	if !exists {
		return nil
	}
	codes, _ := v.([]string)
	return codes
}

// PermissionMiddleware checks X-Permission header against Casbin enforcer.
// Only enforces when the header is present; routes without the header pass through.
func PermissionMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		permCode := c.GetHeader("X-Permission")
		if permCode != "" && claims != nil {
			sub := fmt.Sprintf("user:%d", claims.UserID)
			if ok, _ := enforcer.Enforce(sub, permCode, "access"); !ok {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
