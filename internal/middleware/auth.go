package middleware

import "github.com/gin-gonic/gin"

// AuthMiddleware is a stub; replace with real JWT/session validation.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: validate Authorization header / session token
		c.Next()
	}
}
