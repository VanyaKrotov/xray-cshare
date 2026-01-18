package crypto_helpers

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/VanyaKrotov/xray_cshare/transfer"
	"lukechampine.com/blake3"
)

const (
	InvalidPrivateKeyLength int = 1001
	GenerateCurve25519Error int = 1002
)

func Curve25519Genkey(input_base64 string, enc *base64.Encoding) *transfer.Response {
	var privateKey []byte
	if len(input_base64) > 0 {
		privateKey, _ = enc.DecodeString(input_base64)
		if len(privateKey) != 32 {
			return transfer.New(InvalidPrivateKeyLength, "Invalid length of X25519 private key.")
		}
	}

	privateKey, password, hash32, err := genCurve25519(privateKey)
	if err != nil {
		return transfer.New(GenerateCurve25519Error, err.Error())
	}

	return transfer.Success(fmt.Sprintf("%v|%v|%v", enc.EncodeToString(privateKey), enc.EncodeToString(password), enc.EncodeToString(hash32[:])))

}

func genCurve25519(inputPrivateKey []byte) (privateKey []byte, password []byte, hash32 [32]byte, returnErr error) {
	if len(inputPrivateKey) > 0 {
		privateKey = inputPrivateKey
	}

	if privateKey == nil {
		privateKey = make([]byte, 32)
		rand.Read(privateKey)
	}

	// Modify random bytes using algorithm described at:
	// https://cr.yp.to/ecdh.html
	// (Just to make sure printing the real private key)
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	key, err := ecdh.X25519().NewPrivateKey(privateKey)
	if err != nil {
		returnErr = err
		return
	}

	password = key.PublicKey().Bytes()
	hash32 = blake3.Sum256(password)

	return
}
