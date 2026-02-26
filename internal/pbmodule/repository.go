package pbmodule

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(m *PlatformModule) error {
	return r.db.Create(m).Error
}

func (r *Repository) FindByID(id uint) (*PlatformModule, error) {
	var m PlatformModule
	err := r.db.First(&m, id).Error
	return &m, err
}

func (r *Repository) List() ([]PlatformModule, error) {
	var modules []PlatformModule
	err := r.db.Where("deleted_at IS NULL").Order("sort_order, id").Find(&modules).Error
	return modules, err
}

func (r *Repository) ListEnabled() ([]PlatformModule, error) {
	var modules []PlatformModule
	err := r.db.Where("enabled = true AND deleted_at IS NULL").Order("sort_order, id").Find(&modules).Error
	return modules, err
}

func (r *Repository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&PlatformModule{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&PlatformModule{}, id).Error
}
