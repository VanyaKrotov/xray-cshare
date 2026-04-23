package crypto_helpers

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"

	"github.com/xtls/xray-core/transport/internet/tls"
)

// ExecuteCertChainHash calculates a SHA-256 hash for the full PEM certificate chain.
// Deprecated: xray-core v26.1.23 replaced the upstream chain-hash helper with a leaf-certificate hash API.
// This wrapper keeps the previous chain-hash behavior for backward compatibility.
// Input: certPem - file or cert content
func ExecuteCertChainHash(certPem string) (string, error) {
	certContent := []byte(certPem)
	if pathExists(certPem) {
		content, err := os.ReadFile(certPem)
		if err != nil {
			return "", err
		}

		certContent = content
	}

	return calculatePEMCertChainSHA256Hash(certContent), nil
}

// ExecuteLeafCertHash calculates a SHA-256 hash for the leaf PEM certificate using the current xray-core API.
// Input: certPem - file or cert content
func ExecuteLeafCertHash(certPem string) (string, error) {
	certContent := []byte(certPem)
	if pathExists(certPem) {
		content, err := os.ReadFile(certPem)
		if err != nil {
			return "", err
		}

		certContent = content
	}

	var certs []*x509.Certificate
	if bytes.Contains(certContent, []byte("BEGIN")) {
		for {
			block, remain := pem.Decode(certContent)
			if block == nil {
				break
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return "", err
			}

			certs = append(certs, cert)
			certContent = remain
		}
	} else {
		parsed, err := x509.ParseCertificates(certContent)
		if err != nil {
			return "", err
		}

		certs = parsed
	}

	if len(certs) == 0 {
		return "", errors.New("no certificates found")
	}

	return tls.GenerateCertHashHex(certs[0]), nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

func calculatePEMCertChainSHA256Hash(certContent []byte) string {
	var certChain [][]byte
	for {
		block, remain := pem.Decode(certContent)
		if block == nil {
			break
		}

		certChain = append(certChain, block.Bytes)
		certContent = remain
	}

	var hashValue []byte
	for _, certValue := range certChain {
		out := sha256.Sum256(certValue)
		if hashValue == nil {
			hashValue = out[:]
		} else {
			newHashValue := sha256.Sum256(append(hashValue, out[:]...))
			hashValue = newHashValue[:]
		}
	}

	return base64.StdEncoding.EncodeToString(hashValue)
}
