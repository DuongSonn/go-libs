package _jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Service implements the TokenService interface
type Service struct {
	config *Config
}

// NewService creates a new JWT service
func NewService(config *Config) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Service{
		config: config,
	}, nil
}

// Generate creates a new token with the given claims
func (s *Service) Generate(claims Claims) (string, error) {
	// Create a new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":    claims.ID,
		"sub":    claims.Subject,
		"iss":    claims.Issuer,
		"iat":    claims.IssuedAt.Unix(),
		"exp":    claims.ExpiresAt.Unix(),
		"nbf":    claims.NotBefore.Unix(),
		"aud":    claims.Audience,
		"type":   string(claims.TokenType),
		"roles":  claims.Roles,
		"scopes": claims.Scopes,
		"custom": claims.Custom,
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(s.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// Parse parses and validates a token string
func (s *Service) Parse(tokenString string) (*Token, error) {
	// Parse the token
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		token := &Token{
			Raw:     tokenString,
			Valid:   parsedToken.Valid,
			Headers: parsedToken.Header,
			Claims:  mapClaimsToClaims(claims),
		}
		return token, nil
	}

	return nil, errors.New("invalid token claims")
}

// Validate validates a token and returns its claims
func (s *Service) Validate(tokenString string) (*Claims, error) {
	token, err := s.Parse(tokenString)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &token.Claims, nil
}

// Refresh generates a new access token using a refresh token
func (s *Service) Refresh(refreshToken string) (string, error) {
	// Validate the refresh token
	claims, err := s.Validate(refreshToken)
	if err != nil {
		return "", err
	}

	// Check if the token is a refresh token
	if claims.TokenType != RefreshToken {
		return "", errors.New("not a refresh token")
	}

	// Generate a new access token
	return s.GenerateAccessToken(claims.Subject, claims.Roles, claims.Custom)
}

// GenerateAccessToken generates an access token for a user
func (s *Service) GenerateAccessToken(userID string, roles []string, customClaims map[string]interface{}) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.AccessTokenExpiration)

	claims := Claims{
		ID:        uuid.New().String(),
		Subject:   userID,
		Issuer:    s.config.Issuer,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
		NotBefore: now,
		Audience:  s.config.Audience,
		TokenType: AccessToken,
		Roles:     roles,
		Custom:    customClaims,
	}

	return s.Generate(claims)
}

// GenerateRefreshToken generates a refresh token for a user
func (s *Service) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.RefreshTokenExpiration)

	claims := Claims{
		ID:        uuid.New().String(),
		Subject:   userID,
		Issuer:    s.config.Issuer,
		IssuedAt:  now,
		ExpiresAt: expiresAt,
		NotBefore: now,
		Audience:  s.config.Audience,
		TokenType: RefreshToken,
	}

	return s.Generate(claims)
}

// GenerateTokenPair generates both access and refresh tokens
func (s *Service) GenerateTokenPair(userID string, roles []string, customClaims map[string]interface{}) (string, string, error) {
	accessToken, err := s.GenerateAccessToken(userID, roles, customClaims)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// mapClaimsToClaims converts jwt.MapClaims to Claims
func mapClaimsToClaims(mapClaims jwt.MapClaims) Claims {
	claims := Claims{}

	// Extract standard claims
	if id, ok := mapClaims["jti"].(string); ok {
		claims.ID = id
	}
	if sub, ok := mapClaims["sub"].(string); ok {
		claims.Subject = sub
	}
	if iss, ok := mapClaims["iss"].(string); ok {
		claims.Issuer = iss
	}
	if iat, ok := mapClaims["iat"].(float64); ok {
		claims.IssuedAt = time.Unix(int64(iat), 0)
	}
	if exp, ok := mapClaims["exp"].(float64); ok {
		claims.ExpiresAt = time.Unix(int64(exp), 0)
	}
	if nbf, ok := mapClaims["nbf"].(float64); ok {
		claims.NotBefore = time.Unix(int64(nbf), 0)
	}
	if aud, ok := mapClaims["aud"].(string); ok {
		claims.Audience = aud
	}

	// Extract custom claims
	if tokenType, ok := mapClaims["type"].(string); ok {
		claims.TokenType = TokenType(tokenType)
	}

	// Extract roles
	if roles, ok := mapClaims["roles"].([]interface{}); ok {
		for _, role := range roles {
			if roleStr, ok := role.(string); ok {
				claims.Roles = append(claims.Roles, roleStr)
			}
		}
	}

	// Extract scopes
	if scopes, ok := mapClaims["scopes"].([]interface{}); ok {
		for _, scope := range scopes {
			if scopeStr, ok := scope.(string); ok {
				claims.Scopes = append(claims.Scopes, scopeStr)
			}
		}
	}

	// Extract custom map
	if custom, ok := mapClaims["custom"].(map[string]interface{}); ok {
		claims.Custom = custom
	}

	return claims
}
