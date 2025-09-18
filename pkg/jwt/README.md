# JWT Package

A flexible JWT (JSON Web Token) implementation for authentication and authorization in Go applications.

## Features

-   Token generation and validation
-   Support for access and refresh tokens
-   Role-based authorization
-   Customizable claims
-   Thread-safe implementation

## Installation

```bash
go get github.com/golang-jwt/jwt/v5
```

## Usage

### Basic Setup

```go
package main

import (
    "fmt"
    "log"
    "time"

    jwt "github.com/DuongSonn/go-libs/pkg/jwt"
)

func main() {
    // Create configuration
    config := jwt.DefaultConfig()
    config.SecretKey = "your-secret-key" // Use a secure key in production
    config.Issuer = "my-app"
    config.AccessTokenExpiration = 15 * time.Minute

    // Create JWT service
    service, err := jwt.NewService(config)
    if err != nil {
        log.Fatalf("Failed to create JWT service: %v", err)
    }

    // Generate tokens
    userID := "user123"
    roles := []string{"user", "admin"}
    customClaims := map[string]interface{}{
        "email": "user@example.com",
    }

    accessToken, refreshToken, err := service.GenerateTokenPair(userID, roles, customClaims)
    if err != nil {
        log.Fatalf("Failed to generate tokens: %v", err)
    }

    fmt.Printf("Access Token: %s\n", accessToken)
    fmt.Printf("Refresh Token: %s\n", refreshToken)
}
```

### Token Validation

```go
// Validate a token
claims, err := service.Validate(accessToken)
if err != nil {
    log.Fatalf("Invalid token: %v", err)
}

fmt.Printf("User ID: %s\n", claims.Subject)
fmt.Printf("Roles: %v\n", claims.Roles)

// Check if user has a specific role
hasAdminRole := false
for _, role := range claims.Roles {
    if role == "admin" {
        hasAdminRole = true
        break
    }
}

if hasAdminRole {
    fmt.Println("User has admin role")
} else {
    fmt.Println("User does not have admin role")
}
```

### Refreshing Tokens

```go
// Refresh an access token using a refresh token
newAccessToken, err := service.Refresh(refreshToken)
if err != nil {
    log.Fatalf("Failed to refresh token: %v", err)
}

fmt.Printf("New Access Token: %s\n", newAccessToken)
```

### Manual Token Extraction

```go
// Extract token from Authorization header
func extractTokenFromHeader(authHeader string) (string, error) {
    if authHeader == "" {
        return "", fmt.Errorf("missing authorization header")
    }

    // Check if the header has the Bearer prefix
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return "", fmt.Errorf("invalid authorization header format")
    }

    return parts[1], nil
}

// Example usage in an HTTP handler
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Get token from header
    authHeader := r.Header.Get("Authorization")
    tokenString, err := extractTokenFromHeader(authHeader)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    // Validate token
    claims, err := service.Validate(tokenString)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    // Check if it's an access token
    if claims.TokenType != jwt.AccessToken {
        http.Error(w, "Invalid token type", http.StatusUnauthorized)
        return
    }

    // Use claims data
    userID := claims.Subject
    fmt.Fprintf(w, "Hello, %s", userID)
}
```

## Configuration Options

The `Config` struct provides the following options:

```go
type Config struct {
    // Secret key used to sign tokens
    SecretKey string

    // Issuer claim (iss)
    Issuer string

    // Audience claim (aud)
    Audience string

    // Access token expiration time
    AccessTokenExpiration time.Duration

    // Refresh token expiration time
    RefreshTokenExpiration time.Duration
}
```

## Claims Structure

The JWT claims structure includes standard JWT claims plus custom fields:

```go
type Claims struct {
    // Standard claims
    ID        string    // JWT ID
    Subject   string    // Subject (usually user ID)
    Issuer    string    // Issuer
    IssuedAt  time.Time // Issued at
    ExpiresAt time.Time // Expires at
    NotBefore time.Time // Not valid before
    Audience  string    // Audience

    // Custom claims
    TokenType TokenType            // Token type (access or refresh)
    Roles     []string             // User roles
    Scopes    []string             // Token scopes
    Custom    map[string]interface{} // Custom claims
}
```

## Error Handling

The package provides specific error types for common JWT issues:

-   `ErrMissingAuthHeader`: Missing Authorization header
-   `ErrInvalidAuthHeader`: Invalid Authorization header format
-   `ErrInvalidToken`: Invalid token
-   `ErrExpiredToken`: Token has expired
-   `ErrInvalidSigningMethod`: Invalid signing method
-   `ErrInvalidTokenType`: Invalid token type
-   `ErrNotRefreshToken`: Not a refresh token
-   `ErrInsufficientPermissions`: Insufficient permissions

## Security Considerations

1. **Secret Key**: Use a strong, unique secret key and keep it secure.
2. **Token Expiration**: Use short-lived access tokens and longer-lived refresh tokens.
3. **HTTPS**: Always use HTTPS when transmitting tokens.
4. **Sensitive Data**: Avoid storing sensitive data in tokens as they can be decoded (though not modified without the secret).

## License

This library is distributed under the same license as the go-libs project.
