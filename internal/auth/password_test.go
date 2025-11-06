package auth

import (
	"testing"
)

// TestHashPassword tests password hashing functionality
func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Short password",
			password: "pwd",
			wantErr:  false,
		},
		{
			name:     "Long password",
			password: "thisIsAVeryLongPasswordThatShouldStillWorkFine123!@#",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash the password
			hashed, err := HashPassword(tt.password)

			// Check for errors
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify hash is not empty
			if !tt.wantErr && hashed == "" {
				t.Error("HashPassword() returned empty hash")
			}

			// Verify hash is different from original password
			if !tt.wantErr && hashed == tt.password {
				t.Error("HashPassword() returned unhashed password")
			}
		})
	}
}

// TestComparePasswords tests password comparison functionality
func TestComparePasswords(t *testing.T) {
	password := "testPassword123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		want           bool
	}{
		{
			name:           "Correct password",
			hashedPassword: hashedPassword,
			password:       password,
			want:           true,
		},
		{
			name:           "Incorrect password",
			hashedPassword: hashedPassword,
			password:       "wrongPassword",
			want:           false,
		},
		{
			name:           "Empty password",
			hashedPassword: hashedPassword,
			password:       "",
			want:           false,
		},
		{
			name:           "Case sensitive",
			hashedPassword: hashedPassword,
			password:       "TestPassword123",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare passwords
			result := ComparePasswords(tt.hashedPassword, tt.password)

			// Verify result
			if result != tt.want {
				t.Errorf("ComparePasswords() = %v, want %v", result, tt.want)
			}
		})
	}
}

// TestHashPasswordConsistency verifies that the same password produces different hashes
func TestHashPasswordConsistency(t *testing.T) {
	password := "consistencyTest123"

	// Hash the same password twice
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	// Check for errors
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to hash password: err1=%v, err2=%v", err1, err2)
	}

	// Hashes should be different (bcrypt uses random salt)
	if hash1 == hash2 {
		t.Error("HashPassword() produced identical hashes for the same password")
	}

	// But both should validate against the original password
	if !ComparePasswords(hash1, password) {
		t.Error("First hash does not validate against original password")
	}
	if !ComparePasswords(hash2, password) {
		t.Error("Second hash does not validate against original password")
	}
}
