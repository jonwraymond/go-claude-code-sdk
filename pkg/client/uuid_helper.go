package client

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// GenerateSessionID generates a new random UUID v4 for use as a session ID.
// This ensures compatibility with Claude Code CLI which requires UUID-formatted session IDs.
func GenerateSessionID() string {
	return uuid.New().String()
}

// IsValidUUID checks if a string is a valid UUID format.
func IsValidUUID(s string) bool {
	// First check with regex to ensure proper format with dashes
	if !uuidRegex.MatchString(s) {
		return false
	}
	// Then validate it's a parseable UUID
	_, err := uuid.Parse(s)
	return err == nil
}

// NormalizeSessionID ensures a session ID is in valid UUID format.
// If the input is empty, a new UUID is generated.
// If the input is already a valid UUID, it is returned unchanged.
// Otherwise, a deterministic UUID is generated from the input string.
func NormalizeSessionID(input string) (string, error) {
	// Empty input gets a new random UUID
	if input == "" {
		return GenerateSessionID(), nil
	}

	// Already a valid UUID
	if IsValidUUID(input) {
		return input, nil
	}

	// Generate deterministic UUID from input string
	// This allows non-UUID session names to be consistently mapped
	normalized, err := GenerateUUIDFromString(input)
	if err != nil {
		return "", fmt.Errorf("failed to normalize session ID: %w", err)
	}

	return normalized, nil
}

// GenerateUUIDFromString creates a deterministic UUID v4 from any string input.
// This is useful for converting human-readable session names to valid UUIDs.
func GenerateUUIDFromString(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("input string cannot be empty")
	}

	// Create SHA-256 hash of the input
	hash := sha256.Sum256([]byte(input))
	hashStr := hex.EncodeToString(hash[:])

	// Format as UUID v4
	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	// where y is one of 8, 9, A, or B
	uuidStr := fmt.Sprintf("%s-%s-4%s-%s-%s",
		hashStr[0:8],
		hashStr[8:12],
		hashStr[12:15],
		"8"+hashStr[16:19], // Set version bits for UUID v4
		hashStr[20:32],
	)

	// Validate the generated UUID
	if !IsValidUUID(uuidStr) {
		return "", fmt.Errorf("failed to generate valid UUID from string")
	}

	return uuidStr, nil
}

// ValidateSessionID checks if a session ID is valid for use with Claude Code CLI.
// Returns an error with a helpful message if the ID is invalid.
func ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	if !IsValidUUID(sessionID) {
		return fmt.Errorf("session ID must be a valid UUID (format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx). Got: %s", sessionID)
	}

	return nil
}

// SuggestSessionID provides a suggested session ID based on user input.
// If the input is already valid, it returns it unchanged.
// Otherwise, it returns both a normalized version and a new random UUID as options.
func SuggestSessionID(input string) (string, []string, error) {
	suggestions := []string{}

	// If empty, just suggest a new UUID
	if input == "" {
		newID := GenerateSessionID()
		return newID, []string{newID}, nil
	}

	// If already valid, return as-is
	if IsValidUUID(input) {
		return input, []string{input}, nil
	}

	// Generate deterministic UUID from input
	normalized, err := GenerateUUIDFromString(input)
	if err == nil {
		suggestions = append(suggestions, normalized)
	}

	// Also suggest a completely new UUID
	suggestions = append(suggestions, GenerateSessionID())

	// Return the normalized version as primary suggestion
	if len(suggestions) > 0 {
		return suggestions[0], suggestions, nil
	}

	return "", nil, fmt.Errorf("could not generate valid session ID from input: %s", input)
}

// FormatSessionIDError creates a helpful error message for invalid session IDs.
func FormatSessionIDError(providedID string) error {
	if providedID == "" {
		return fmt.Errorf("session ID is required. Use GenerateSessionID() to create a new one")
	}

	if !IsValidUUID(providedID) {
		normalized, _ := GenerateUUIDFromString(providedID)
		return fmt.Errorf(
			"invalid session ID '%s'. Session IDs must be UUIDs. "+
				"Suggested ID: %s (based on your input) or use GenerateSessionID() for a new random ID",
			providedID, normalized,
		)
	}

	return nil
}

// TrimSessionID cleans up a session ID by removing whitespace and converting to lowercase.
func TrimSessionID(sessionID string) string {
	return strings.ToLower(strings.TrimSpace(sessionID))
}