package utils

import (
	"github.com/google/uuid"
)

// GenerateSecurePublicLink generates a secure UUID string (37 characters) for public links
// This provides maximum security and uniqueness for public board access
// Format: "p" + full UUID (e.g., "p550e8400-e29b-41d4-a716-446655440000")
// Security: 2^128 possible combinations, making brute force attacks practically impossible
func GenerateShortUUID() string {
	return "p" + uuid.New().String() // "p" prefix + full UUID = 37 total
}

// GenerateBoardID generates a board ID with "b" prefix and 8-character UUID
func GenerateBoardID() string {
	return "b" + uuid.New().String()[:8]
}

// GenerateIdeaID generates an idea ID with "i" prefix and 8-character UUID
func GenerateIdeaID() string {
	return "i" + uuid.New().String()[:8]
}

// GenerateFullUUID generates a full UUID string for cases where maximum uniqueness is needed
func GenerateFullUUID() string {
	return uuid.New().String()
}
