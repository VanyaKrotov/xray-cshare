package crypto_helpers

import (
	"crypto/mlkem"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/VanyaKrotov/xray_cshare/transfer"
	"lukechampine.com/blake3"
)

func ExecuteMLKEM768(input_mlkem768 string) *transfer.Response {
	var seed [64]byte
	if len(input_mlkem768) > 0 {
		s, _ := base64.RawURLEncoding.DecodeString(input_mlkem768)
		if len(s) != 64 {
			return transfer.New(1, "Invalid length of ML-KEM-768 seed.")
		}

		seed = [64]byte(s)
	} else {
		rand.Read(seed[:])
	}
	seed, client, hash32 := genMLKEM768(&seed)

	// Seed: %v
	// Client: %v
	// Hash32: %v
	return transfer.Success(fmt.Sprintf("%v|%v|%v", base64.RawURLEncoding.EncodeToString(seed[:]), base64.RawURLEncoding.EncodeToString(client), base64.RawURLEncoding.EncodeToString(hash32[:])))
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
