package role

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(req *CreateRoleRequest) (*Role, error) {
	r := &Role{
		TenantCode:  req.TenantCode,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Service) GetByID(id uint) (*Role, error) {
	return s.repo.FindByID(id)
}

func (s *Service) List(tenantCode string) ([]Role, error) {
	return s.repo.List(tenantCode)
}

func (s *Service) Update(id uint, req *UpdateRoleRequest) error {
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	return s.repo.Update(id, updates)
}

func (s *Service) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *Service) AssignPermissions(roleID uint, codes []string) error {
	return s.repo.AssignPermissionsToRole(roleID, codes)
}

func (s *Service) GetPermissionCodes(roleID uint) ([]string, error) {
	return s.repo.GetPermissionCodesForRole(roleID)
}

func (s *Service) AssignRolesToUser(userID uint, roleIDs []uint) error {
	return s.repo.AssignRolesToUser(userID, roleIDs)
}

func (s *Service) GetPermissionCodesForUser(userID uint) ([]string, error) {
	return s.repo.GetPermissionCodesForUser(userID)
}

func (s *Service) GetRolesForUser(userID uint) ([]uint, error) {
	return s.repo.GetRolesForUser(userID)
}
