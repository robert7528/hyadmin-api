package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for authentication.
type Handler struct {
	svc *Service
}

// NewHandler creates an auth Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type loginRequest struct {
	Provider   string `json:"provider"`    // default "local"
	TenantCode string `json:"tenant_code"` // required for local
	Username   string `json:"username"`
	Password   string `json:"password"`
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creds := map[string]string{
		"tenant_code": req.TenantCode,
		"username":    req.Username,
		"password":    req.Password,
	}
	provider := req.Provider
	if provider == "" {
		provider = "local"
	}

	token, err := h.svc.Login(provider, creds)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "provider": provider})
}

// Logout handles POST /api/v1/auth/logout (client-side token invalidation)
func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
