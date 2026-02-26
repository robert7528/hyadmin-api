package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hysp/hyadmin-api/internal/config"
)

// Service signs/verifies JWTs and delegates authentication to Providers.
type Service struct {
	providers map[string]Provider
	cfg       *config.Config
}

// NewService constructs an auth Service with one or more providers.
func NewService(cfg *config.Config, providers ...Provider) *Service {
	m := make(map[string]Provider, len(providers))
	for _, p := range providers {
		m[p.Name()] = p
	}
	return &Service{providers: m, cfg: cfg}
}

// Login authenticates via the named provider and returns a signed JWT.
func (s *Service) Login(providerName string, creds map[string]string) (string, error) {
	if providerName == "" {
		providerName = "local"
	}
	p, ok := s.providers[providerName]
	if !ok {
		return "", fmt.Errorf("auth: unknown provider %q", providerName)
	}
	claims, err := p.Authenticate(nil, creds)
	if err != nil {
		return "", err
	}
	// Stamp times from Service config (authoritative source)
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(time.Duration(s.cfg.JWT.ExpiryHours) * time.Hour))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// ParseToken validates a JWT string and returns its Claims.
func (s *Service) ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("auth: invalid token: %w", err)
	}
	c, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("auth: invalid claims")
	}
	return c, nil
}
