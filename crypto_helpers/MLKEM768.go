package crypto_helpers

import (
	"crypto/mlkem"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"lukechampine.com/blake3"
)

// Generate key pair for ML-KEM-768 post-quantum key exchange (VLESS Encryption).
// Input: seed in base64.RawURLEncoding;
func ExecuteMLKEM768(input_mlkem768 string) (*MLKEM768, error) {
	var seed [64]byte
	if len(input_mlkem768) > 0 {
		s, _ := base64.RawURLEncoding.DecodeString(input_mlkem768)
		if len(s) != 64 {
			return nil, errors.New("Invalid length of ML-KEM-768 seed.")
		}

		seed = [64]byte(s)
	} else {
		rand.Read(seed[:])
	}

	seed, client, hash32 := genMLKEM768(&seed)

	return &MLKEM768{
		Seed:   base64.RawURLEncoding.EncodeToString(seed[:]),
		Client: base64.RawURLEncoding.EncodeToString(client),
		Hash32: base64.RawURLEncoding.EncodeToString(hash32[:]),
	}, nil
}

func genMLKEM768(inputSeed *[64]byte) (seed [64]byte, client []byte, hash32 [32]byte) {
	if inputSeed == nil {
		rand.Read(seed[:])
	} else {
		seed = *inputSeed
	}

	key, _ := mlkem.NewDecapsulationKey768(seed[:])
	client = key.EncapsulationKey().Bytes()
	hash32 = blake3.Sum256(client)

	return seed, client, hash32
}

type MLKEM768 struct {
	Seed   string `json:"seed"`
	Client string `json:"client"`
	Hash32 string `json:"hash32"`
}
