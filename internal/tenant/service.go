package tenant

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListTenants() ([]Tenant, error) {
	return s.repo.FindAll()
}

func (s *Service) GetTenant(id uint) (*Tenant, error) {
	return s.repo.FindByID(id)
}

func (s *Service) CreateTenant(t *Tenant) error {
	return s.repo.Create(t)
}

func (s *Service) UpdateTenant(t *Tenant) error {
	return s.repo.Update(t)
}

func (s *Service) DeleteTenant(id uint) error {
	return s.repo.Delete(id)
}
