package feature

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(req *CreateFeatureRequest) (*Feature, error) {
	f := &Feature{
		ModuleID:    req.ModuleID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Icon:        req.Icon,
		Path:        req.Path,
		SortOrder:   req.SortOrder,
		Enabled:     true,
	}
	if err := s.repo.Create(f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) GetByID(id uint) (*Feature, error) {
	return s.repo.FindByID(id)
}

func (s *Service) ListByModule(moduleID uint) ([]Feature, error) {
	return s.repo.ListByModule(moduleID)
}

func (s *Service) Update(id uint, req *UpdateFeatureRequest) error {
	updates := make(map[string]interface{})
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	return s.repo.Update(id, updates)
}

func (s *Service) Delete(id uint) error {
	return s.repo.Delete(id)
}
