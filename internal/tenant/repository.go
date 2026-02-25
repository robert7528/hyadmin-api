package tenant

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAll() ([]Tenant, error) {
	var tenants []Tenant
	err := r.db.Find(&tenants).Error
	return tenants, err
}

func (r *Repository) FindByID(id uint) (*Tenant, error) {
	var t Tenant
	err := r.db.First(&t, id).Error
	return &t, err
}

func (r *Repository) Create(t *Tenant) error {
	return r.db.Create(t).Error
}

func (r *Repository) Update(t *Tenant) error {
	return r.db.Save(t).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Tenant{}, id).Error
}
