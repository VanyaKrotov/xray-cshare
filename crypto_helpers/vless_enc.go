package crypto_helpers

import (
	"encoding/base64"
	"strings"
)

// Generate decryption/encryption json pair (VLESS Encryption).
func ExecuteVLESSEnc() VLessEnc {
	privateKey, password, _, _ := genCurve25519(nil)
	serverKey := base64.RawURLEncoding.EncodeToString(privateKey)
	clientKey := base64.RawURLEncoding.EncodeToString(password)
	decryption := generateDotConfig("mlkem768x25519plus", "native", "600s", serverKey)
	encryption := generateDotConfig("mlkem768x25519plus", "native", "0rtt", clientKey)

	seed, client, _ := genMLKEM768(nil)
	serverKeyPQ := base64.RawURLEncoding.EncodeToString(seed[:])
	clientKeyPQ := base64.RawURLEncoding.EncodeToString(client)
	decryptionPQ := generateDotConfig("mlkem768x25519plus", "native", "600s", serverKeyPQ)
	encryptionPQ := generateDotConfig("mlkem768x25519plus", "native", "0rtt", clientKeyPQ)

	return VLessEnc{Decryption: decryption, Encryption: encryption, EncryptionPQ: encryptionPQ, DecryptionPQ: decryptionPQ}
}

func generateDotConfig(fields ...string) string {
	return strings.Join(fields, ".")
}

type VLessEnc struct {
	// Authentication: X25519, not Post-Quantum decryption
	Decryption string `json:"decryption"`
	// Authentication: X25519, not Post-Quantum encryption
	Encryption string `json:"encryption"`
	// Authentication: ML-KEM-768, Post-Quantum decryption
	DecryptionPQ string `json:"decryption_pq"`
	// Authentication: ML-KEM-768, Post-Quantum encryption
	EncryptionPQ string `json:"encryption_pq"`
}
