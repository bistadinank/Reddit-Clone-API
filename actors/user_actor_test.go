package actors

import (
	"testing"
)

func TestValidateToken(t *testing.T) {
	// Create a new UserActor
	userActor := NewUserActor()

	// Setup test cases
	tests := []struct {
		name          string
		token         string
		wantUsername  string
		wantValid     bool
		setupUser     bool  // whether to create user before test
	}{
		{
			name:         "Valid token with existing user",
			token:        "reddit-token-testuser1",
			wantUsername: "testuser1",
			wantValid:    true,
			setupUser:    true,
		},
		{
			name:         "Invalid token format",
			token:        "invalid-token",
			wantUsername: "",
			wantValid:    false,
			setupUser:    false,
		},
		{
			name:         "Valid format but user doesn't exist",
			token:        "reddit-token-nonexistent",
			wantUsername: "nonexistent",
			wantValid:    false,
			setupUser:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create user if needed
			if tt.setupUser {
				userMutex.Lock()
				globalUsers[tt.wantUsername] = "password123"
				userMutex.Unlock()
			}

			// Test token validation
			gotUsername, gotValid := userActor.ValidateToken(tt.token)

			// Check results
			if gotUsername != tt.wantUsername {
				t.Errorf("ValidateToken() username = %v, want %v", gotUsername, tt.wantUsername)
			}
			if gotValid != tt.wantValid {
				t.Errorf("ValidateToken() valid = %v, want %v", gotValid, tt.wantValid)
			}

			// Cleanup
			if tt.setupUser {
				userMutex.Lock()
				delete(globalUsers, tt.wantUsername)
				userMutex.Unlock()
			}
		})
	}
} 