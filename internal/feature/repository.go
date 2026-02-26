package feature

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(f *Feature) error {
	return r.db.Create(f).Error
}

func (r *Repository) FindByID(id uint) (*Feature, error) {
	var f Feature
	err := r.db.First(&f, id).Error
	return &f, err
}

func (r *Repository) ListByModule(moduleID uint) ([]Feature, error) {
	var features []Feature
	err := r.db.Where("module_id = ? AND deleted_at IS NULL", moduleID).Order("sort_order, id").Find(&features).Error
	return features, err
}

func (r *Repository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Feature{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Feature{}, id).Error
}
