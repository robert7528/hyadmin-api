package auditlog

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// List GET /api/v1/admin/audit-logs
func (h *Handler) List(c *gin.Context) {
	tc := c.Query("tenant_code")
	resource := c.Query("resource")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}

	q := h.db.Model(&AuditLog{})
	if tc != "" {
		q = q.Where("tenant_code = ?", tc)
	}
	if resource != "" {
		q = q.Where("resource LIKE ?", "%"+resource+"%")
	}

	var total int64
	q.Count(&total)

	var logs []AuditLog
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs, "total": total})
}
