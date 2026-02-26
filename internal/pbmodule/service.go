package pbmodule

import (
	"github.com/hysp/hyadmin-api/internal/feature"
	"github.com/hysp/hyadmin-api/internal/permission"
)

type Service struct {
	repo       *Repository
	featureSvc *feature.Service
	permRepo   *permission.Repository
}

func NewService(repo *Repository, featureSvc *feature.Service, permRepo *permission.Repository) *Service {
	return &Service{repo: repo, featureSvc: featureSvc, permRepo: permRepo}
}

func (s *Service) Create(req *CreateModuleRequest) (*PlatformModule, error) {
	m := &PlatformModule{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Icon:        req.Icon,
		Route:       req.Route,
		URL:         req.URL,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Enabled:     true,
	}
	if err := s.repo.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) GetByID(id uint) (*PlatformModule, error) {
	return s.repo.FindByID(id)
}

func (s *Service) List() ([]PlatformModule, error) {
	return s.repo.List()
}

// ListForUser returns modules visible to the user based on their permission codes.
// A module is visible if the user has at least one menu-type permission for a feature in it.
func (s *Service) ListForUser(permCodes []string) ([]PlatformModule, error) {
	if len(permCodes) == 0 {
		return []PlatformModule{}, nil
	}

	// Find permissions matching user's codes
	perms, err := s.permRepo.FindByCodes(permCodes)
	if err != nil {
		return nil, err
	}

	// Collect feature IDs that have menu-type permissions
	featureIDs := make(map[uint]struct{})
	for _, p := range perms {
		if p.Type == "menu" {
			featureIDs[p.FeatureID] = struct{}{}
		}
	}
	if len(featureIDs) == 0 {
		return []PlatformModule{}, nil
	}

	// Get all enabled modules, then filter those with visible features
	allModules, err := s.repo.ListEnabled()
	if err != nil {
		return nil, err
	}

	visibleModuleIDs := make(map[uint]struct{})
	for fid := range featureIDs {
		f, err := s.featureSvc.GetByID(fid)
		if err != nil {
			continue
		}
		visibleModuleIDs[f.ModuleID] = struct{}{}
	}

	result := make([]PlatformModule, 0)
	for _, m := range allModules {
		if _, ok := visibleModuleIDs[m.ID]; ok {
			result = append(result, m)
		}
	}
	return result, nil
}

func (s *Service) Update(id uint, req *UpdateModuleRequest) error {
	updates := make(map[string]interface{})
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.URL != "" {
		updates["url"] = req.URL
	}
	if req.Description != "" {
		updates["description"] = req.Description
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
