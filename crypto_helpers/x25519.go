package crypto_helpers

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"lukechampine.com/blake3"
)

// Generate key pair for X25519 key exchange (REALITY, VLESS Encryption).
// Input: private key (base64.RawURLEncoding)
func Curve25519Genkey(input_base64 string, enc *base64.Encoding) (*C25519Key, error) {
	var privateKey []byte
	if len(input_base64) > 0 {
		privateKey, _ = enc.DecodeString(input_base64)
		if len(privateKey) != 32 {
			return nil, errors.New("Invalid length of X25519 private key.")
		}
	}

	privateKey, password, hash32, err := genCurve25519(privateKey)
	if err != nil {
		return nil, err
	}

	return &C25519Key{PrivateKey: enc.EncodeToString(privateKey), Password: enc.EncodeToString(password), Hash32: enc.EncodeToString(hash32[:])}, nil

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

type C25519Key struct {
	PrivateKey string `json:"private_key"`
	Password   string `json:"password"`
	Hash32     string `json:"hash32"`
}
