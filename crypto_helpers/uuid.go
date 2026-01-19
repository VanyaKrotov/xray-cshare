package crypto_helpers

import (
	"errors"

	"github.com/xtls/xray-core/common/uuid"
)

// Generate UUIDv4 or UUIDv5 (VLESS).
// Input: initial word (30 bytes); Returns: UUIDv4 (random)
func ExecuteUUID(input string) (string, error) {
	l := len(input)
	if l == 0 {
		u := uuid.New()

		return u.String(), nil
	}

	if l <= 30 {
		u, _ := uuid.ParseString(input)

		return u.String(), nil
	}

	return "", errors.New("Input must be within 30 bytes.")
}
