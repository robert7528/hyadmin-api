// Package auth implements JWT authentication with a pluggable provider pattern.
// Current providers: local (username/password).
// Future: hysso, oauth2, saml etc. — implement the Provider interface.
package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the standard JWT payload used across all providers.
type Claims struct {
	UserID     uint   `json:"uid"`
	TenantCode string `json:"tc"`
	Username   string `json:"username"`
	Provider   string `json:"provider"` // "local" | "hysso" | ...
	jwt.RegisteredClaims
}

// Provider is the pluggable authentication interface.
// Implement this to add OAuth2, SAML, LDAP, hysso etc.
type Provider interface {
	Name() string
	// Authenticate verifies credentials and returns Claims on success.
	Authenticate(ctx context.Context, creds map[string]string) (*Claims, error)
}

// newClaims creates a Claims struct with the given fields and expiry.
func newClaims(userID uint, tenantCode, username, provider string, expiryHours int) *Claims {
	now := time.Now()
	return &Claims{
		UserID:     userID,
		TenantCode: tenantCode,
		Username:   username,
		Provider:   provider,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiryHours) * time.Hour)),
		},
	}
}
