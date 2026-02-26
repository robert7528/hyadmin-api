package feature

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

// ListByModule handles both:
//   GET /api/v1/features?module_id=...
//   GET /api/v1/admin/modules/:moduleId/features
func (h *Handler) ListByModule(c *gin.Context) {
	// Check URL param first (from /modules/:moduleId/features route)
	rawID := c.Param("moduleId")
	if rawID == "" {
		rawID = c.Query("module_id")
	}
	mid, err := strconv.ParseUint(rawID, 10, 64)
	if err != nil || mid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "module_id required"})
		return
	}
	features, err := h.svc.ListByModule(uint(mid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"features": features})
}

// Create POST /api/v1/admin/modules/:moduleId/features
func (h *Handler) Create(c *gin.Context) {
	var req CreateFeatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Allow module_id from URL param to override body
	if mid, err := strconv.ParseUint(c.Param("moduleId"), 10, 64); err == nil && mid > 0 {
		req.ModuleID = uint(mid)
	}
	f, err := h.svc.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, f)
}

// Get GET /api/v1/admin/features/:id
func (h *Handler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	f, err := h.svc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, f)
}

// Update PUT /api/v1/admin/features/:id
func (h *Handler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req UpdateFeatureRequest
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

// Delete DELETE /api/v1/admin/features/:id
func (h *Handler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
