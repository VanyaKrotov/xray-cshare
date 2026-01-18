package crypto_helpers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/VanyaKrotov/xray_cshare/transfer"
	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
)

const (
	MLDSA65OverflowInputLength int = 1021
)

func ExecuteMLDSA65(input string) *transfer.Response {
	var seed [32]byte
	if len(input) > 0 {
		s, _ := base64.RawURLEncoding.DecodeString(input)
		if len(s) != 32 {
			return transfer.New(MLDSA65OverflowInputLength, "Invalid length of ML-DSA-65 seed.")
		}

		seed = [32]byte(s)
	} else {
		rand.Read(seed[:])
	}

	pub, _ := mldsa65.NewKeyFromSeed(&seed)

	return transfer.Success(fmt.Sprintf("%v|%v", base64.RawURLEncoding.EncodeToString(seed[:]), base64.RawURLEncoding.EncodeToString(pub.Bytes())))
}
