package adminuser

import (
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(u *AdminUser) error {
	return r.db.Create(u).Error
}

func (r *Repository) FindByID(id uint) (*AdminUser, error) {
	var u AdminUser
	err := r.db.First(&u, id).Error
	return &u, err
}

func (r *Repository) FindByUsername(tenantCode, username string) (*AdminUser, error) {
	var u AdminUser
	err := r.db.Where("tenant_code = ? AND username = ? AND deleted_at IS NULL", tenantCode, username).First(&u).Error
	return &u, err
}

func (r *Repository) List(tenantCode string, page, pageSize int) ([]AdminUser, int64, error) {
	var users []AdminUser
	var total int64
	q := r.db.Model(&AdminUser{}).Where("tenant_code = ?", tenantCode)
	q.Count(&total)
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *Repository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&AdminUser{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&AdminUser{}, id).Error
}
