package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secretKey  string
	expiration time.Duration
}

// NewJWTManager creates a new JWT manager instance
// secretKey: The secret key used to sign tokens
// expiration: How long tokens remain valid
func NewJWTManager(secretKey string, expiration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:  secretKey,
		expiration: expiration,
	}
}

// GenerateToken creates a new JWT token for a user
// userID: The unique identifier for the user
// email: The user's email address
// Returns the signed token string or an error
func (m *JWTManager) GenerateToken(userID, email string) (string, error) {
	// Create claims with user data and expiration
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken verifies and parses a JWT token
// tokenString: The token string to validate
// Returns the claims if valid, or an error if invalid/expired
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract and validate claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
