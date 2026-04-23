package main

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"strings"
	"testing"

	"github.com/xtls/xray-core/core"
)

func TestGetXrayCoreVersion(t *testing.T) {
	resp := unpackResponse(t, GetXrayCoreVersion())
	if resp.Code != 0 {
		t.Fatalf("expected success code, got %d", resp.Code)
	}
	if resp.ContentType != testContentMessage {
		t.Fatalf("expected message content type, got %d", resp.ContentType)
	}
	if resp.Body != core.Version() {
		t.Fatalf("expected version %q, got %q", core.Version(), resp.Body)
	}
}

func TestCurve25519Genkey(t *testing.T) {
	type payload struct {
		PrivateKey string `json:"private_key"`
		Password   string `json:"password"`
		Hash32     string `json:"hash32"`
	}

	resp := unpackResponse(t, curve25519GenkeyString(""))
	if resp.Code != 0 || resp.ContentType != testContentPayload {
		t.Fatalf("expected successful payload response, got %+v", resp)
	}

	value := decodeJSONBody[payload](t, resp.Body)
	if value.PrivateKey == "" || value.Password == "" || value.Hash32 == "" {
		t.Fatalf("expected non-empty X25519 payload, got %+v", value)
	}
}

func TestCurve25519GenkeyWG(t *testing.T) {
	type payload struct {
		PrivateKey string `json:"private_key"`
	}

	resp := unpackResponse(t, curve25519GenkeyWGString(""))
	if resp.Code != 0 || resp.ContentType != testContentPayload {
		t.Fatalf("expected successful payload response, got %+v", resp)
	}

	value := decodeJSONBody[payload](t, resp.Body)
	if value.PrivateKey == "" || !strings.HasSuffix(value.PrivateKey, "=") {
		t.Fatalf("expected padded base64 private key, got %+v", value)
	}
}

func TestExecuteUUID(t *testing.T) {
	resp := unpackResponse(t, executeUUIDString(""))
	if resp.Code != 0 || resp.ContentType != testContentMessage {
		t.Fatalf("expected successful UUID response, got %+v", resp)
	}
	if len(resp.Body) != 36 {
		t.Fatalf("expected UUID string, got %q", resp.Body)
	}
}

func TestExecuteUUIDInvalid(t *testing.T) {
	resp := unpackResponse(t, executeUUIDString(strings.Repeat("a", 31)))
	if resp.Code == 0 || resp.ContentType != testContentError {
		t.Fatalf("expected error response, got %+v", resp)
	}
}

func TestExecuteMLDSA65(t *testing.T) {
	type payload struct {
		Seed   string `json:"seed"`
		Verify string `json:"verify"`
	}

	resp := unpackResponse(t, executeMLDSA65String(""))
	if resp.Code != 0 || resp.ContentType != testContentPayload {
		t.Fatalf("expected successful payload response, got %+v", resp)
	}

	value := decodeJSONBody[payload](t, resp.Body)
	if value.Seed == "" || value.Verify == "" {
		t.Fatalf("expected non-empty MLDSA payload, got %+v", value)
	}
}

func TestExecuteMLKEM768(t *testing.T) {
	type payload struct {
		Seed   string `json:"seed"`
		Client string `json:"client"`
		Hash32 string `json:"hash32"`
	}

	resp := unpackResponse(t, executeMLKEM768String(""))
	if resp.Code != 0 || resp.ContentType != testContentPayload {
		t.Fatalf("expected successful payload response, got %+v", resp)
	}

	value := decodeJSONBody[payload](t, resp.Body)
	if value.Seed == "" || value.Client == "" || value.Hash32 == "" {
		t.Fatalf("expected non-empty MLKEM payload, got %+v", value)
	}
}

func TestExecuteVLESSEnc(t *testing.T) {
	type payload struct {
		Decryption   string `json:"decryption"`
		Encryption   string `json:"encryption"`
		DecryptionPQ string `json:"decryption_pq"`
		EncryptionPQ string `json:"encryption_pq"`
	}

	resp := unpackResponse(t, ExecuteVLESSEnc())
	if resp.Code != 0 || resp.ContentType != testContentPayload {
		t.Fatalf("expected successful payload response, got %+v", resp)
	}

	value := decodeJSONBody[payload](t, resp.Body)
	if value.Decryption == "" || value.Encryption == "" || value.DecryptionPQ == "" || value.EncryptionPQ == "" {
		t.Fatalf("expected non-empty VLESS encryption payload, got %+v", value)
	}
}

func TestGenerateCert(t *testing.T) {
	type payload struct {
		Certificate []string `json:"certificate"`
		Key         []string `json:"key"`
	}

	resp := unpackResponse(t, generateCertStrings("example.com,www.example.com", "example.com", "Example Org", 0, ""))
	if resp.Code != 0 || resp.ContentType != testContentPayload {
		t.Fatalf("expected successful payload response, got %+v", resp)
	}

	value := decodeJSONBody[payload](t, resp.Body)
	if len(value.Certificate) == 0 || len(value.Key) == 0 {
		t.Fatalf("expected certificate and key payload, got %+v", value)
	}
	if value.Certificate[0] != "-----BEGIN CERTIFICATE-----" {
		t.Fatalf("unexpected certificate header: %q", value.Certificate[0])
	}
	if !strings.HasPrefix(value.Key[0], "-----BEGIN ") || !strings.Contains(value.Key[0], "PRIVATE KEY-----") {
		t.Fatalf("unexpected key header: %q", value.Key[0])
	}
}

func TestExecuteCertChainHash(t *testing.T) {
	type certPayload struct {
		Certificate []string `json:"certificate"`
	}

	generated := unpackResponse(t, generateCertStrings("example.com", "example.com", "Example Org", 0, "24h"))
	if generated.Code != 0 {
		t.Fatalf("failed to generate certificate: %+v", generated)
	}

	payload := decodeJSONBody[certPayload](t, generated.Body)
	certPEM := strings.Join(payload.Certificate, "\n") + "\n"

	resp := unpackResponse(t, executeCertChainHashString(certPEM))
	if resp.Code != 0 || resp.ContentType != testContentMessage {
		t.Fatalf("expected successful cert hash response, got %+v", resp)
	}
	if resp.Body == "" {
		t.Fatal("expected non-empty certificate hash")
	}
}

func TestSetEnv(t *testing.T) {
	key := "XRAY_CSHARE_TEST_ENV"
	t.Cleanup(func() {
		_ = os.Unsetenv(key)
	})

	resp := unpackResponse(t, setEnvStrings(key, "test-value"))
	if resp.Code != 0 || resp.ContentType != testContentMessage || resp.Body != "done" {
		t.Fatalf("expected successful SetEnv response, got %+v", resp)
	}
	if got := os.Getenv(key); got != "test-value" {
		t.Fatalf("expected environment variable to be set, got %q", got)
	}
}

func TestPingDirect(t *testing.T) {
	server := newHeadServer()
	defer server.Close()

	resp := unpackResponse(t, pingPort(0, server.URL))
	if resp.Code != 0 || resp.ContentType != testContentPayload {
		t.Fatalf("expected successful payload response, got %+v", resp)
	}

	var result struct {
		Port    int    `json:"port"`
		Timeout int    `json:"timeout"`
		Error   string `json:"error"`
	}
	result = decodeJSONBody[struct {
		Port    int    `json:"port"`
		Timeout int    `json:"timeout"`
		Error   string `json:"error"`
	}](t, resp.Body)
	if result.Port != 0 || result.Timeout < 0 || result.Error != "" {
		t.Fatalf("unexpected ping result: %+v", result)
	}
}

func TestFreePointerNil(t *testing.T) {
	FreePointer(nil)
}

func TestCurve25519GenkeyInvalid(t *testing.T) {
	resp := unpackResponse(t, curve25519GenkeyString("invalid"))
	if resp.Code == 0 || resp.ContentType != testContentError {
		t.Fatalf("expected error response, got %+v", resp)
	}
}

func TestExecuteMLDSA65Invalid(t *testing.T) {
	resp := unpackResponse(t, executeMLDSA65String("invalid"))
	if resp.Code == 0 || resp.ContentType != testContentError {
		t.Fatalf("expected error response, got %+v", resp)
	}
}

func TestExecuteMLKEM768Invalid(t *testing.T) {
	resp := unpackResponse(t, executeMLKEM768String("invalid"))
	if resp.Code == 0 || resp.ContentType != testContentError {
		t.Fatalf("expected error response, got %+v", resp)
	}
}

func TestGenerateCertInvalidExpire(t *testing.T) {
	resp := unpackResponse(t, generateCertStrings("example.com", "example.com", "Example Org", 0, "not-a-duration"))
	if resp.Code == 0 || resp.ContentType != testContentError {
		t.Fatalf("expected error response, got %+v", resp)
	}
}

func TestExecuteCertChainHashFromFile(t *testing.T) {
	type certPayload struct {
		Certificate []string `json:"certificate"`
	}

	generated := unpackResponse(t, generateCertStrings("example.org", "example.org", "Example Org", 0, "24h"))
	if generated.Code != 0 {
		t.Fatalf("failed to generate certificate: %+v", generated)
	}

	payload := decodeJSONBody[certPayload](t, generated.Body)
	certPEM := strings.Join(payload.Certificate, "\n") + "\n"
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		t.Fatal("failed to decode PEM certificate")
	}
	if _, err := x509.ParseCertificate(block.Bytes); err != nil {
		t.Fatalf("generated certificate should be parseable: %v", err)
	}

	file, err := os.CreateTemp(t.TempDir(), "cert-*.pem")
	if err != nil {
		t.Fatalf("failed to create temp cert file: %v", err)
	}
	if _, err := file.WriteString(certPEM); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("failed to close cert file: %v", err)
	}

	resp := unpackResponse(t, executeCertChainHashString(file.Name()))
	if resp.Code != 0 || resp.ContentType != testContentMessage || resp.Body == "" {
		t.Fatalf("expected successful cert hash response, got %+v", resp)
	}
}

func TestCurve25519GenkeyRoundTrip(t *testing.T) {
	type payload struct {
		PrivateKey string `json:"private_key"`
		Password   string `json:"password"`
		Hash32     string `json:"hash32"`
	}

	first := unpackResponse(t, curve25519GenkeyString(""))
	if first.Code != 0 {
		t.Fatalf("failed to generate initial X25519 key: %+v", first)
	}
	value := decodeJSONBody[payload](t, first.Body)

	second := unpackResponse(t, curve25519GenkeyString(value.PrivateKey))
	if second.Code != 0 {
		t.Fatalf("failed to regenerate X25519 key: %+v", second)
	}
	value2 := decodeJSONBody[payload](t, second.Body)

	if value.PrivateKey != value2.PrivateKey || value.Password != value2.Password || value.Hash32 != value2.Hash32 {
		t.Fatalf("expected deterministic round-trip, got first=%+v second=%+v", value, value2)
	}
	if _, err := base64.RawURLEncoding.DecodeString(value.PrivateKey); err != nil {
		t.Fatalf("expected valid raw URL base64 private key: %v", err)
	}
}
