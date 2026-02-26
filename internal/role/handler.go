package role

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List GET /api/v1/admin/roles?tenant_code=...
func (h *Handler) List(c *gin.Context) {
	tc := c.Query("tenant_code")
	roles, err := h.svc.List(tc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// Create POST /api/v1/admin/roles
func (h *Handler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.svc.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, r)
}

// Get GET /api/v1/admin/roles/:id
func (h *Handler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	r, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	codes, _ := h.svc.GetPermissionCodes(uint(id))
	c.JSON(http.StatusOK, gin.H{"role": r, "permission_codes": codes})
}

// Update PUT /api/v1/admin/roles/:id
func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Update(uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// Delete DELETE /api/v1/admin/roles/:id
func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// GetPermissions GET /api/v1/admin/roles/:id/permissions
func (h *Handler) GetPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	codes, err := h.svc.GetPermissionCodes(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"permission_codes": codes})
}

// AssignPermissions PUT /api/v1/admin/roles/:id/permissions
func (h *Handler) AssignPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Codes []string `json:"codes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.AssignPermissions(uint(id), req.Codes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "permissions assigned"})
}

// AssignUsers PUT /api/v1/admin/roles/:id/users
func (h *Handler) AssignUsers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req AssignUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for _, uid := range req.UserIDs {
		// get existing roles for user then add this role
		roles, _ := h.svc.GetRolesForUser(uid)
		rolesSet := make(map[uint]struct{})
		for _, r := range roles {
			rolesSet[r] = struct{}{}
		}
		rolesSet[uint(id)] = struct{}{}
		newRoles := make([]uint, 0, len(rolesSet))
		for r := range rolesSet {
			newRoles = append(newRoles, r)
		}
		if err := h.svc.AssignRolesToUser(uid, newRoles); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "users assigned to role"})
}
