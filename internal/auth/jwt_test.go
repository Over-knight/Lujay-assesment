package auth

import (
	"testing"
	"time"
)

// TestJWTManager_GenerateToken tests JWT token generation
func TestJWTManager_GenerateToken(t *testing.T) {
	secretKey := "test-secret-key"
	expiration := 1 * time.Hour
	manager := NewJWTManager(secretKey, expiration)

	tests := []struct {
		name    string
		userID  string
		email   string
		wantErr bool
	}{
		{
			name:    "Valid user data",
			userID:  "123456789",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Empty user ID",
			userID:  "",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Empty email",
			userID:  "123456789",
			email:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate token
			token, err := manager.GenerateToken(tt.userID, tt.email)

			// Check for errors
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify token is not empty
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

// TestJWTManager_ValidateToken tests JWT token validation
func TestJWTManager_ValidateToken(t *testing.T) {
	secretKey := "test-secret-key"
	expiration := 1 * time.Hour
	manager := NewJWTManager(secretKey, expiration)

	// Generate a valid token
	userID := "123456789"
	email := "test@example.com"
	validToken, err := manager.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "Invalid token format",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Malformed token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.malformed",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate token
			claims, err := manager.ValidateToken(tt.token)

			// Check for errors
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify claims for valid tokens
			if !tt.wantErr {
				if claims == nil {
					t.Error("ValidateToken() returned nil claims")
					return
				}
				if claims.UserID != userID {
					t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, userID)
				}
				if claims.Email != email {
					t.Errorf("ValidateToken() email = %v, want %v", claims.Email, email)
				}
			}
		})
	}
}

// TestJWTManager_TokenExpiration tests JWT token expiration
func TestJWTManager_TokenExpiration(t *testing.T) {
	secretKey := "test-secret-key"
	expiration := 1 * time.Second // Short expiration for testing
	manager := NewJWTManager(secretKey, expiration)

	// Generate token
	token, err := manager.GenerateToken("123456789", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid immediately
	_, err = manager.ValidateToken(token)
	if err != nil {
		t.Errorf("Token should be valid immediately after creation: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Token should now be expired
	_, err = manager.ValidateToken(token)
	if err == nil {
		t.Error("Token should be expired after expiration time")
	}
}

// TestJWTManager_DifferentSecretKey tests token validation with different secret keys
func TestJWTManager_DifferentSecretKey(t *testing.T) {
	manager1 := NewJWTManager("secret-key-1", 1*time.Hour)
	manager2 := NewJWTManager("secret-key-2", 1*time.Hour)

	// Generate token with first manager
	token, err := manager1.GenerateToken("123456789", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate with first manager (should succeed)
	_, err = manager1.ValidateToken(token)
	if err != nil {
		t.Errorf("Token should be valid with same secret key: %v", err)
	}

	// Validate with second manager (should fail)
	_, err = manager2.ValidateToken(token)
	if err == nil {
		t.Error("Token should not be valid with different secret key")
	}
}

// TestNewJWTManager tests JWT manager creation
func TestNewJWTManager(t *testing.T) {
	tests := []struct {
		name       string
		secretKey  string
		expiration time.Duration
	}{
		{
			name:       "Standard configuration",
			secretKey:  "my-secret-key",
			expiration: 24 * time.Hour,
		},
		{
			name:       "Short expiration",
			secretKey:  "another-key",
			expiration: 1 * time.Minute,
		},
		{
			name:       "Long expiration",
			secretKey:  "long-key",
			expiration: 168 * time.Hour, // 1 week
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create JWT manager
			manager := NewJWTManager(tt.secretKey, tt.expiration)

			// Verify manager is not nil
			if manager == nil {
				t.Error("NewJWTManager() returned nil")
				return
			}

			// Verify manager can generate and validate tokens
			token, err := manager.GenerateToken("test-user", "test@example.com")
			if err != nil {
				t.Errorf("Failed to generate token: %v", err)
				return
			}

			_, err = manager.ValidateToken(token)
			if err != nil {
				t.Errorf("Failed to validate token: %v", err)
			}
		})
	}
}
