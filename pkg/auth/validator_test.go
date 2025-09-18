package auth

import (
	"strings"
	"testing"
)

func TestValidator_ValidateAPIKey(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		apiKey  string
		wantErr bool
		errType error
	}{
		{
			name:    "valid API key",
			apiKey:  "sk-ant-api03j9h8f7d6s5a4l3k2m1n0",
			wantErr: false,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			wantErr: true,
			errType: ErrMissingCredentials,
		},
		{
			name:    "too short",
			apiKey:  "sk-ant-short",
			wantErr: true,
			errType: ErrInvalidAPIKey,
		},
		{
			name:    "too long",
			apiKey:  "sk-ant-" + strings.Repeat("a", 200),
			wantErr: true,
			errType: ErrInvalidAPIKey,
		},
		{
			name:    "wrong prefix",
			apiKey:  "pk-ant-abcdefghijklmnopqrstuvwxyz123456789",
			wantErr: true,
			errType: ErrInvalidAPIKey,
		},
		{
			name:    "test key",
			apiKey:  "sk-ant-test-abcdefghijklmnopqrstuvwxyz",
			wantErr: true,
			errType: ErrWeakCredentials,
		},
		{
			name:    "example key",
			apiKey:  "sk-ant-example-abcdefghijklmnopqrstuvwxyz",
			wantErr: true,
			errType: ErrWeakCredentials,
		},
		{
			name:    "low entropy - repeated chars",
			apiKey:  "sk-ant-aaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			wantErr: true,
			errType: ErrWeakCredentials,
		},
		{
			name:    "sequential pattern",
			apiKey:  "sk-ant-abcdefghijk123456789012345",
			wantErr: true,
			errType: ErrWeakCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateAPIKey(tt.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errType != nil {
				if !strings.Contains(err.Error(), tt.errType.Error()) {
					t.Errorf("ValidateAPIKey() error = %v, want error containing %v", err, tt.errType)
				}
			}
		})
	}
}

func TestValidator_ValidateSessionKey(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name       string
		sessionKey string
		wantErr    bool
		errType    error
	}{
		{
			name:       "valid session key",
			sessionKey: "sess_abcdefghijklmnopqrstuvwxyz123456789",
			wantErr:    false,
		},
		{
			name:       "valid with dots and dashes",
			sessionKey: "sess_abc-def.ghi_jkl-mno.pqr_stu123456",
			wantErr:    false,
		},
		{
			name:       "empty session key",
			sessionKey: "",
			wantErr:    true,
			errType:    ErrMissingCredentials,
		},
		{
			name:       "too short",
			sessionKey: "short",
			wantErr:    true,
			errType:    ErrInvalidSessionKey,
		},
		{
			name:       "too long",
			sessionKey: strings.Repeat("a", 300),
			wantErr:    true,
			errType:    ErrInvalidSessionKey,
		},
		{
			name:       "invalid characters",
			sessionKey: "sess_abc!@#$%^&*()_+{}[]|\\:\";<>?,./~`123456",
			wantErr:    true,
			errType:    ErrInvalidSessionKey,
		},
		{
			name:       "missing numbers",
			sessionKey: "sess_abcdefghijklmnopqrstuvwxyzabcdef",
			wantErr:    true,
			errType:    ErrInvalidSessionKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateSessionKey(tt.sessionKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSessionKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errType != nil {
				if !strings.Contains(err.Error(), tt.errType.Error()) {
					t.Errorf("ValidateSessionKey() error = %v, want error containing %v", err, tt.errType)
				}
			}
		})
	}
}

func TestValidator_checkComplexity(t *testing.T) {
	tests := []struct {
		name       string
		validator  *Validator
		credential string
		wantErr    bool
	}{
		{
			name: "meets all requirements",
			validator: &Validator{
				requireUppercase: true,
				requireLowercase: true,
				requireNumbers:   true,
				requireSpecial:   true,
			},
			credential: "Test123!@#",
			wantErr:    false,
		},
		{
			name: "missing uppercase",
			validator: &Validator{
				requireUppercase: true,
			},
			credential: "test123",
			wantErr:    true,
		},
		{
			name: "missing lowercase",
			validator: &Validator{
				requireLowercase: true,
			},
			credential: "TEST123",
			wantErr:    true,
		},
		{
			name: "missing numbers",
			validator: &Validator{
				requireNumbers: true,
			},
			credential: "TestOnly",
			wantErr:    true,
		},
		{
			name: "missing special",
			validator: &Validator{
				requireSpecial: true,
			},
			credential: "Test123",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.checkComplexity(tt.credential)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkComplexity() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_isTestKey(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name   string
		key    string
		isTest bool
	}{
		{"normal key", "sk-ant-api03j9h8f7d6s5a4l3k2m1n0", false},
		{"test key", "sk-ant-test-123456789", true},
		{"example key", "sk-ant-example-abcdef", true},
		{"demo key", "sk-ant-demo-xyz", true},
		{"sample key", "sk-ant-sample-key", true},
		{"dummy key", "sk-ant-dummy-credentials", true},
		{"fake key", "sk-ant-fake-apikey", true},
		{"mock key", "sk-ant-mock-testing", true},
		{"xxx pattern", "sk-ant-xxx-xxx-xxx", true},
		{"123 pattern", "sk-ant-123-456-789", true},
		{"abc pattern", "sk-ant-abc-def-ghi", true},
		{"repeated pattern", "sk-ant-aaaaaaaaaa", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.isTestKey(tt.key); got != tt.isTest {
				t.Errorf("isTestKey() = %v, want %v", got, tt.isTest)
			}
		})
	}
}

func TestValidator_hasLowEntropy(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name       string
		key        string
		lowEntropy bool
	}{
		{"high entropy", "sk-ant-a1b2c3d4e5f6g7h8i9j0k1l2m3n4", false},
		{"low entropy - repeated", "sk-ant-aaaaaaaaaaaaaaaaaaaaaaaaa", true},
		{"low entropy - pattern", "sk-ant-abababababababababababab", true},
		{"medium entropy", "sk-ant-a1b2c3d4e5f6g7h8i9j0k1", false},
		{"good mix", "sk-ant-Xy9Kp2Lm5Qr8Nt4Ws7Bv3Zf6", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.hasLowEntropy(tt.key); got != tt.lowEntropy {
				t.Errorf("hasLowEntropy() = %v, want %v", got, tt.lowEntropy)
			}
		})
	}
}

func TestValidator_hasRepeatingPattern(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name       string
		s          string
		hasPattern bool
	}{
		{"no pattern", "a1b2c3d4e5", false},
		{"sequential numbers", "abc123def", true},
		{"sequential letters", "123abcxyz", true},
		{"reverse sequential", "xyz321cba", true},
		{"repeated substring", "abcabcabc", true},
		{"partial repeat", "testtest123", true},
		{"complex no pattern", "Xy9Kp2Lm5Q", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.hasRepeatingPattern(tt.s); got != tt.hasPattern {
				t.Errorf("hasRepeatingPattern() = %v, want %v", got, tt.hasPattern)
			}
		})
	}
}

func TestNewCustomValidator(t *testing.T) {
	// Test with custom options
	v := NewCustomValidator(
		WithAPIKeyPattern(`^custom-[a-z0-9]+$`),
		WithAPIKeyLength(10, 50),
		WithSessionKeyPattern(`^sess:[a-zA-Z0-9]+$`),
		WithComplexityRequirements(true, true, true, true),
	)

	// Test custom API key pattern
	err := v.ValidateAPIKey("custom-randomkey")
	if err != nil {
		t.Errorf("Custom API key validation failed: %v", err)
	}

	err = v.ValidateAPIKey("sk-ant-notmatching")
	if err == nil {
		t.Error("Expected error for non-matching API key pattern")
	}

	// Test custom length
	err = v.ValidateAPIKey("custom-ab")
	if err == nil {
		t.Error("Expected error for too short API key")
	}
}
