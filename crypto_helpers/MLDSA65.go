package crypto_helpers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
)

// Generate key pair for ML-DSA-65 post-quantum signature (REALITY).
// Input: seed in base64.RawURLEncoding; Returns: {Seed|Verify}
func ExecuteMLDSA65(input string) (*MLDSA65, error) {
	var seed [32]byte
	if len(input) > 0 {
		s, _ := base64.RawURLEncoding.DecodeString(input)
		if len(s) != 32 {
			return nil, errors.New("Invalid length of ML-DSA-65 seed.")
		}

		seed = [32]byte(s)
	} else {
		rand.Read(seed[:])
	}

	pub, _ := mldsa65.NewKeyFromSeed(&seed)

	return &MLDSA65{
		Seed:   base64.RawURLEncoding.EncodeToString(seed[:]),
		Verify: base64.RawURLEncoding.EncodeToString(pub.Bytes()),
	}, nil
}

type MLDSA65 struct {
	Seed   string `json:"seed"`
	Verify string `json:"verify"`
}
