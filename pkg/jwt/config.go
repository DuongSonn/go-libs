package _jwt

import (
	"errors"
	"time"
)

// Config holds configuration for JWT
type Config struct {
	// Secret key used to sign tokens
	SecretKey string `json:"secret_key" yaml:"secret_key"`

	// Issuer claim (iss)
	Issuer string `json:"issuer" yaml:"issuer"`

	// Audience claim (aud)
	Audience string `json:"audience" yaml:"audience"`

	// Access token expiration time
	AccessTokenExpiration time.Duration `json:"access_token_expiration" yaml:"access_token_expiration"`

	// Refresh token expiration time
	RefreshTokenExpiration time.Duration `json:"refresh_token_expiration" yaml:"refresh_token_expiration"`
}

// DefaultConfig returns a default configuration for JWT
func DefaultConfig() *Config {
	return &Config{
		SecretKey:              "your-secret-key", // Should be overridden in production
		Issuer:                 "go-libs",
		Audience:               "api",
		AccessTokenExpiration:  15 * time.Minute,
		RefreshTokenExpiration: 24 * time.Hour,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.SecretKey == "" {
		return errors.New("secret key is required")
	}

	if c.AccessTokenExpiration <= 0 {
		return errors.New("access token expiration must be greater than 0")
	}

	if c.RefreshTokenExpiration <= 0 {
		return errors.New("refresh token expiration must be greater than 0")
	}

	return nil
}

// TokenType represents the type of token
type TokenType string

const (
	// AccessToken is used for API access
	AccessToken TokenType = "access"

	// RefreshToken is used to obtain new access tokens
	RefreshToken TokenType = "refresh"
)
