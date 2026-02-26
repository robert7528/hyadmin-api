package permission

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

// ListByFeature GET /api/v1/admin/features/:id/permissions
func (h *Handler) ListByFeature(c *gin.Context) {
	fid, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	perms, err := h.svc.ListByFeature(uint(fid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"permissions": perms})
}

// Create POST /api/v1/admin/features/:id/permissions
func (h *Handler) Create(c *gin.Context) {
	fid, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if fid > 0 {
		req.FeatureID = uint(fid)
	}
	p, err := h.svc.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// BatchCreate POST /api/v1/admin/features/:id/permissions/batch
func (h *Handler) BatchCreate(c *gin.Context) {
	fid, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req BatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if fid > 0 {
		req.FeatureID = uint(fid)
	}
	perms, err := h.svc.BatchCreate(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"permissions": perms})
}

// Update PUT /api/v1/admin/permissions/:id
func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdatePermissionRequest
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

// Delete DELETE /api/v1/admin/permissions/:id
func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
