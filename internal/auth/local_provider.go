package auth

import (
	"context"
	"fmt"

	"github.com/hysp/hyadmin-api/internal/adminuser"
)

// LocalProvider authenticates users against the local admin_users table.
type LocalProvider struct {
	userSvc *adminuser.Service
	expiry  int
}

// NewLocalProvider creates a LocalProvider.
func NewLocalProvider(userSvc *adminuser.Service, expiryHours int) *LocalProvider {
	return &LocalProvider{userSvc: userSvc, expiry: expiryHours}
}

func (p *LocalProvider) Name() string { return "local" }

func (p *LocalProvider) Authenticate(_ context.Context, creds map[string]string) (*Claims, error) {
	tenantCode := creds["tenant_code"]
	username := creds["username"]
	password := creds["password"]

	if tenantCode == "" || username == "" {
		return nil, fmt.Errorf("auth: tenant_code and username are required")
	}

	u, err := p.userSvc.VerifyPassword(tenantCode, username, password)
	if err != nil {
		return nil, err
	}

	return newClaims(u.ID, u.TenantCode, u.Username, "local", p.expiry), nil
}
