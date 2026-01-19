package crypto_helpers

import (
	"os"

	"github.com/xtls/xray-core/transport/internet/tls"
)

// Calculate TLS certificates hash.
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

	return tls.CalculatePEMCertChainSHA256Hash(certContent), nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}
