package client

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestGenerateSessionID(t *testing.T) {
	// Test that GenerateSessionID returns valid UUIDs
	for i := 0; i < 10; i++ {
		id := GenerateSessionID()
		if !IsValidUUID(id) {
			t.Errorf("GenerateSessionID() returned invalid UUID: %s", id)
		}

		// Ensure each ID is unique
		id2 := GenerateSessionID()
		if id == id2 {
			t.Errorf("GenerateSessionID() returned duplicate IDs: %s", id)
		}
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid UUID v4",
			input:    "550e8400-e29b-41d4-a716-446655440000",
			expected: true,
		},
		{
			name:     "Valid UUID with uppercase",
			input:    "550E8400-E29B-41D4-A716-446655440000",
			expected: true,
		},
		{
			name:     "Invalid UUID - wrong format",
			input:    "550e8400-e29b-41d4-a716",
			expected: false,
		},
		{
			name:     "Invalid UUID - no dashes",
			input:    "550e8400e29b41d4a716446655440000",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Random string",
			input:    "not-a-uuid",
			expected: false,
		},
		{
			name:     "Generated UUID",
			input:    uuid.New().String(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUUID(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidUUID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeSessionID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:        "Empty string generates new UUID",
			input:       "",
			shouldError: false,
		},
		{
			name:        "Valid UUID passes through",
			input:       "550e8400-e29b-41d4-a716-446655440000",
			shouldError: false,
		},
		{
			name:        "Non-UUID string gets normalized",
			input:       "my-test-session",
			shouldError: false,
		},
		{
			name:        "Same input produces same UUID",
			input:       "test-session-123",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeSessionID(tt.input)

			if tt.shouldError && err == nil {
				t.Errorf("NormalizeSessionID(%q) expected error but got none", tt.input)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("NormalizeSessionID(%q) unexpected error: %v", tt.input, err)
			}

			if err == nil {
				// Result should be a valid UUID
				if !IsValidUUID(result) {
					t.Errorf("NormalizeSessionID(%q) returned invalid UUID: %s", tt.input, result)
				}

				// If input was already a valid UUID, it should pass through unchanged
				if IsValidUUID(tt.input) && result != tt.input {
					t.Errorf("NormalizeSessionID(%q) changed valid UUID to %s", tt.input, result)
				}

				// Same input should produce same output (deterministic)
				if tt.input != "" {
					result2, _ := NormalizeSessionID(tt.input)
					if result != result2 {
						t.Errorf("NormalizeSessionID(%q) not deterministic: %s != %s", tt.input, result, result2)
					}
				}
			}
		})
	}
}

func TestGenerateUUIDFromString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:        "Simple string",
			input:       "test",
			shouldError: false,
		},
		{
			name:        "Complex string",
			input:       "user-123-session-abc",
			shouldError: false,
		},
		{
			name:        "Empty string",
			input:       "",
			shouldError: true,
		},
		{
			name:        "Unicode string",
			input:       "测试会话",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateUUIDFromString(tt.input)

			if tt.shouldError && err == nil {
				t.Errorf("GenerateUUIDFromString(%q) expected error but got none", tt.input)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("GenerateUUIDFromString(%q) unexpected error: %v", tt.input, err)
			}

			if err == nil {
				// Result should be a valid UUID
				if !IsValidUUID(result) {
					t.Errorf("GenerateUUIDFromString(%q) returned invalid UUID: %s", tt.input, result)
				}

				// Same input should always produce same output
				result2, _ := GenerateUUIDFromString(tt.input)
				if result != result2 {
					t.Errorf("GenerateUUIDFromString(%q) not deterministic: %s != %s", tt.input, result, result2)
				}

				// Different inputs should produce different outputs
				if tt.input != "test" {
					otherResult, _ := GenerateUUIDFromString("test")
					if result == otherResult {
						t.Errorf("GenerateUUIDFromString(%q) produced same UUID as 'test'", tt.input)
					}
				}
			}
		})
	}
}

func TestValidateSessionID(t *testing.T) {
	tests := []struct {
		name        string
		sessionID   string
		shouldError bool
		errorPart   string
	}{
		{
			name:        "Valid UUID",
			sessionID:   "550e8400-e29b-41d4-a716-446655440000",
			shouldError: false,
		},
		{
			name:        "Empty string",
			sessionID:   "",
			shouldError: true,
			errorPart:   "cannot be empty",
		},
		{
			name:        "Invalid format",
			sessionID:   "not-a-uuid",
			shouldError: true,
			errorPart:   "must be a valid UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSessionID(tt.sessionID)

			if tt.shouldError && err == nil {
				t.Errorf("ValidateSessionID(%q) expected error but got none", tt.sessionID)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("ValidateSessionID(%q) unexpected error: %v", tt.sessionID, err)
			}

			if tt.shouldError && err != nil && tt.errorPart != "" {
				if !strings.Contains(err.Error(), tt.errorPart) {
					t.Errorf("ValidateSessionID(%q) error = %v, want error containing %q", tt.sessionID, err, tt.errorPart)
				}
			}
		})
	}
}

func TestFormatSessionIDError(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantError bool
	}{
		{
			name:      "Empty ID",
			sessionID: "",
			wantError: true,
		},
		{
			name:      "Invalid ID",
			sessionID: "my-session",
			wantError: true,
		},
		{
			name:      "Valid UUID",
			sessionID: "550e8400-e29b-41d4-a716-446655440000",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FormatSessionIDError(tt.sessionID)

			if tt.wantError && err == nil {
				t.Errorf("FormatSessionIDError(%q) expected error but got none", tt.sessionID)
			}

			if !tt.wantError && err != nil {
				t.Errorf("FormatSessionIDError(%q) unexpected error: %v", tt.sessionID, err)
			}

			// Check that error messages are helpful
			if err != nil {
				errStr := err.Error()
				if tt.sessionID == "" && !strings.Contains(errStr, "GenerateSessionID()") {
					t.Errorf("Error for empty ID should mention GenerateSessionID()")
				}
				if tt.sessionID != "" && !IsValidUUID(tt.sessionID) && !strings.Contains(errStr, "Suggested ID:") {
					t.Errorf("Error for invalid ID should include a suggestion")
				}
			}
		})
	}
}

func TestSuggestSessionID(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Empty input",
			input: "",
		},
		{
			name:  "Valid UUID input",
			input: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:  "Invalid input",
			input: "my-custom-session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary, suggestions, err := SuggestSessionID(tt.input)

			if err != nil {
				t.Errorf("SuggestSessionID(%q) unexpected error: %v", tt.input, err)
			}

			// Primary should be a valid UUID
			if !IsValidUUID(primary) {
				t.Errorf("SuggestSessionID(%q) primary suggestion is not valid UUID: %s", tt.input, primary)
			}

			// All suggestions should be valid UUIDs
			for _, s := range suggestions {
				if !IsValidUUID(s) {
					t.Errorf("SuggestSessionID(%q) suggestion is not valid UUID: %s", tt.input, s)
				}
			}

			// Should have at least one suggestion
			if len(suggestions) == 0 {
				t.Errorf("SuggestSessionID(%q) returned no suggestions", tt.input)
			}

			// If input was valid UUID, primary should be the same
			if IsValidUUID(tt.input) && primary != tt.input {
				t.Errorf("SuggestSessionID(%q) changed valid UUID to %s", tt.input, primary)
			}
		})
	}
}

func TestTrimSessionID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Uppercase UUID",
			input:    "550E8400-E29B-41D4-A716-446655440000",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "UUID with spaces",
			input:    "  550e8400-e29b-41d4-a716-446655440000  ",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "Mixed case with spaces",
			input:    "  550E8400-E29B-41D4-A716-446655440000  ",
			expected: "550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimSessionID(tt.input)
			if result != tt.expected {
				t.Errorf("TrimSessionID(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
