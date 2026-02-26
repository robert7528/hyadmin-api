package permission

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(p *Permission) error {
	return r.db.Create(p).Error
}

func (r *Repository) FindByID(id uint) (*Permission, error) {
	var p Permission
	err := r.db.First(&p, id).Error
	return &p, err
}

func (r *Repository) ListByFeature(featureID uint) ([]Permission, error) {
	var perms []Permission
	err := r.db.Where("feature_id = ? AND deleted_at IS NULL", featureID).Order("sort_order, id").Find(&perms).Error
	return perms, err
}

func (r *Repository) FindByCodes(codes []string) ([]Permission, error) {
	var perms []Permission
	err := r.db.Where("code IN ?", codes).Find(&perms).Error
	return perms, err
}

func (r *Repository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Permission{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Permission{}, id).Error
}
