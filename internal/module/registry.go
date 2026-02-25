package module

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type Registry struct {
	mu      sync.RWMutex
	modules []Module
}

func NewRegistry() *Registry {
	return &Registry{modules: []Module{}}
}

// Register adds a module to the registry at runtime.
func (r *Registry) Register(m Module) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.modules = append(r.modules, m)
}

// ListModules handles GET /api/v1/modules
func (r *Registry) ListModules(c *gin.Context) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c.JSON(http.StatusOK, gin.H{"modules": r.modules})
}
