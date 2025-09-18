package _jwt

import (
	"time"
)

// Claims represents the standard JWT claims plus custom claims
type Claims struct {
	// Standard claims
	ID        string    `json:"jti,omitempty"` // JWT ID
	Subject   string    `json:"sub,omitempty"` // Subject (usually user ID)
	Issuer    string    `json:"iss,omitempty"` // Issuer
	IssuedAt  time.Time `json:"iat,omitempty"` // Issued at
	ExpiresAt time.Time `json:"exp,omitempty"` // Expires at
	NotBefore time.Time `json:"nbf,omitempty"` // Not valid before
	Audience  string    `json:"aud,omitempty"` // Audience

	// Custom claims
	TokenType TokenType              `json:"type,omitempty"`   // Token type (access or refresh)
	Roles     []string               `json:"roles,omitempty"`  // User roles
	Scopes    []string               `json:"scopes,omitempty"` // Token scopes
	Custom    map[string]interface{} `json:"custom,omitempty"` // Custom claims
}

// Token represents a JWT token
type Token struct {
	Raw     string                 // The raw token string
	Claims  Claims                 // The token's claims
	Valid   bool                   // Whether the token is valid
	Headers map[string]interface{} // The token's headers
}

// TokenService defines the interface for JWT token operations
type TokenService interface {
	// Generate creates a new token with the given claims
	Generate(claims Claims) (string, error)

	// Parse parses and validates a token string
	Parse(tokenString string) (*Token, error)

	// Validate validates a token and returns its claims
	Validate(tokenString string) (*Claims, error)

	// Refresh generates a new access token using a refresh token
	Refresh(refreshToken string) (string, error)

	// GenerateAccessToken generates an access token for a user
	GenerateAccessToken(userID string, roles []string, customClaims map[string]interface{}) (string, error)

	// GenerateRefreshToken generates a refresh token for a user
	GenerateRefreshToken(userID string) (string, error)

	// GenerateTokenPair generates both access and refresh tokens
	GenerateTokenPair(userID string, roles []string, customClaims map[string]interface{}) (accessToken string, refreshToken string, err error)
}
