package adminuser

import (
	"fmt"

	"github.com/hysp/hyadmin-api/internal/crypto"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      *Repository
	encryptor crypto.Encryptor
}

func NewService(repo *Repository, enc crypto.Encryptor) *Service {
	return &Service{repo: repo, encryptor: enc}
}

func (s *Service) Create(req *CreateUserRequest) (*AdminUserDTO, error) {
	u := &AdminUser{
		TenantCode: req.TenantCode,
		Username:   req.Username,
		Provider:   req.Provider,
		ProviderID: req.ProviderID,
		Enabled:    true,
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("adminuser: hash password: %w", err)
		}
		u.PasswordHash = string(hash)
	}
	if req.Provider == "" {
		u.Provider = "local"
	}
	var err error
	u.DisplayNameEnc, err = s.encryptor.Encrypt(req.DisplayName)
	if err != nil {
		return nil, err
	}
	u.EmailEnc, err = s.encryptor.Encrypt(req.Email)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return s.toDTO(u)
}

func (s *Service) GetByID(id uint) (*AdminUserDTO, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.toDTO(u)
}

func (s *Service) GetByUsername(tenantCode, username string) (*AdminUser, error) {
	return s.repo.FindByUsername(tenantCode, username)
}

func (s *Service) List(tenantCode string, page, pageSize int) ([]AdminUserDTO, int64, error) {
	users, total, err := s.repo.List(tenantCode, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]AdminUserDTO, 0, len(users))
	for i := range users {
		dto, err := s.toDTO(&users[i])
		if err != nil {
			return nil, 0, err
		}
		dtos = append(dtos, *dto)
	}
	return dtos, total, nil
}

func (s *Service) Update(id uint, req *UpdateUserRequest) error {
	updates := make(map[string]interface{})
	if req.DisplayName != "" {
		enc, err := s.encryptor.Encrypt(req.DisplayName)
		if err != nil {
			return err
		}
		updates["display_name"] = enc
	}
	if req.Email != "" {
		enc, err := s.encryptor.Encrypt(req.Email)
		if err != nil {
			return err
		}
		updates["email"] = enc
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	return s.repo.Update(id, updates)
}

func (s *Service) ChangePassword(id uint, req *ChangePasswordRequest) error {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("adminuser: old password mismatch")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.Update(id, map[string]interface{}{"password_hash": string(hash)})
}

func (s *Service) Delete(id uint) error {
	return s.repo.Delete(id)
}

// VerifyPassword is used by LocalProvider.
func (s *Service) VerifyPassword(tenantCode, username, password string) (*AdminUser, error) {
	u, err := s.repo.FindByUsername(tenantCode, username)
	if err != nil {
		return nil, fmt.Errorf("adminuser: user not found")
	}
	if !u.Enabled {
		return nil, fmt.Errorf("adminuser: user disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("adminuser: invalid password")
	}
	return u, nil
}

func (s *Service) toDTO(u *AdminUser) (*AdminUserDTO, error) {
	dn, err := s.encryptor.Decrypt(u.DisplayNameEnc)
	if err != nil {
		return nil, err
	}
	em, err := s.encryptor.Decrypt(u.EmailEnc)
	if err != nil {
		return nil, err
	}
	return &AdminUserDTO{
		ID:          u.ID,
		TenantCode:  u.TenantCode,
		Username:    u.Username,
		DisplayName: dn,
		Email:       em,
		Provider:    u.Provider,
		Enabled:     u.Enabled,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}, nil
}
