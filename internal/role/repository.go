package role

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"gorm.io/gorm"
)

type Repository struct {
	db       *gorm.DB
	enforcer *casbin.Enforcer
}

func NewRepository(db *gorm.DB, enforcer *casbin.Enforcer) *Repository {
	return &Repository{db: db, enforcer: enforcer}
}

func (r *Repository) Create(role *Role) error {
	return r.db.Create(role).Error
}

func (r *Repository) FindByID(id uint) (*Role, error) {
	var role Role
	err := r.db.First(&role, id).Error
	return &role, err
}

func (r *Repository) List(tenantCode string) ([]Role, error) {
	var roles []Role
	err := r.db.Where("tenant_code = ?", tenantCode).Order("id").Find(&roles).Error
	return roles, err
}

func (r *Repository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Role{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&Role{}, id).Error
}

// AssignPermissionsToRole replaces all p policies for this role with the given permission codes.
func (r *Repository) AssignPermissionsToRole(roleID uint, codes []string) error {
	sub := fmt.Sprintf("role:%d", roleID)
	if _, err := r.enforcer.DeletePermissionsForUser(sub); err != nil {
		return err
	}
	for _, code := range codes {
		if _, err := r.enforcer.AddPolicy(sub, code, "access"); err != nil {
			return err
		}
	}
	return nil
}

// GetPermissionCodesForRole returns all permission codes assigned to a role.
func (r *Repository) GetPermissionCodesForRole(roleID uint) ([]string, error) {
	sub := fmt.Sprintf("role:%d", roleID)
	policies, err := r.enforcer.GetPermissionsForUser(sub)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(policies))
	for _, p := range policies {
		if len(p) > 1 {
			codes = append(codes, p[1]) // p = [sub, obj, act]
		}
	}
	return codes, nil
}

// AssignRolesToUser replaces all g policies for this user with the given roles.
func (r *Repository) AssignRolesToUser(userID uint, roleIDs []uint) error {
	sub := fmt.Sprintf("user:%d", userID)
	if _, err := r.enforcer.DeleteRolesForUser(sub); err != nil {
		return err
	}
	for _, rid := range roleIDs {
		if _, err := r.enforcer.AddRoleForUser(sub, fmt.Sprintf("role:%d", rid)); err != nil {
			return err
		}
	}
	return nil
}

// GetPermissionCodesForUser collects all permission codes reachable by a user via their roles.
func (r *Repository) GetPermissionCodesForUser(userID uint) ([]string, error) {
	sub := fmt.Sprintf("user:%d", userID)
	roles, err := r.enforcer.GetRolesForUser(sub)
	if err != nil {
		return nil, err
	}
	codeSet := make(map[string]struct{})
	for _, roleSub := range roles {
		policies, err := r.enforcer.GetPermissionsForUser(roleSub)
		if err != nil {
			return nil, err
		}
		for _, p := range policies {
			if len(p) > 1 {
				codeSet[p[1]] = struct{}{}
			}
		}
	}
	codes := make([]string, 0, len(codeSet))
	for c := range codeSet {
		codes = append(codes, c)
	}
	return codes, nil
}

// GetRolesForUser returns role IDs assigned to a user.
func (r *Repository) GetRolesForUser(userID uint) ([]uint, error) {
	sub := fmt.Sprintf("user:%d", userID)
	roles, err := r.enforcer.GetRolesForUser(sub)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(roles))
	for _, rs := range roles {
		var id uint
		fmt.Sscanf(rs, "role:%d", &id)
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
