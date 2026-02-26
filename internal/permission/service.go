package permission

import "fmt"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(req *CreatePermissionRequest) (*Permission, error) {
	pType := req.Type
	if pType == "" {
		pType = "button"
	}
	p := &Permission{
		FeatureID:   req.FeatureID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Type:        pType,
		SortOrder:   req.SortOrder,
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

// BatchCreate generates permissions from standard suffixes.
func (s *Service) BatchCreate(req *BatchCreateRequest) ([]Permission, error) {
	suffixNames := map[string]string{
		"view":   "頁面存取",
		"create": "新增",
		"update": "編輯",
		"delete": "刪除",
		"export": "匯出",
	}
	result := make([]Permission, 0, len(req.Suffixes))
	for _, suffix := range req.Suffixes {
		name, ok := suffixNames[suffix]
		if !ok {
			name = suffix
		}
		pType := "button"
		if suffix == "view" {
			pType = "menu"
		}
		p := &Permission{
			FeatureID: req.FeatureID,
			Code:      fmt.Sprintf("%s.%s", req.CodePrefix, suffix),
			Name:      name,
			Type:      pType,
		}
		if err := s.repo.Create(p); err != nil {
			return nil, err
		}
		result = append(result, *p)
	}
	return result, nil
}

func (s *Service) GetByID(id uint) (*Permission, error) {
	return s.repo.FindByID(id)
}

func (s *Service) ListByFeature(featureID uint) ([]Permission, error) {
	return s.repo.ListByFeature(featureID)
}

func (s *Service) Update(id uint, req *UpdatePermissionRequest) error {
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	return s.repo.Update(id, updates)
}

func (s *Service) Delete(id uint) error {
	return s.repo.Delete(id)
}
