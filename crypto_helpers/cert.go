package crypto_helpers

import (
	"crypto/x509"
	"strings"
	"time"

	"github.com/xtls/xray-core/common/protocol/tls/cert"
	"github.com/xtls/xray-core/main/commands/base"
)

// Generate TLS certificates.
func GenerateCert(options GenCertOptions) (*Cert, error) {
	options.initDefaults()

	var opts []cert.Option
	if options.IsCA {
		opts = append(opts, cert.Authority(true))
		opts = append(opts, cert.KeyUsage(x509.KeyUsageCertSign|x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature))
	}

	opts = append(opts, cert.NotAfter(time.Now().Add(*options.Expire)))
	opts = append(opts, cert.CommonName(options.CommonName))
	if len(options.DomainNames) > 0 {
		opts = append(opts, cert.DNSNames(options.DomainNames...))
	}

	opts = append(opts, cert.Organization(options.Organization))

	cert, err := cert.Generate(nil, opts...)
	if err != nil {
		return nil, err
	}

	certPEM, keyPEM := cert.ToPEM()

	return &Cert{
		Certificate: strings.Split(strings.TrimSpace(string(certPEM)), "\n"),
		Key:         strings.Split(strings.TrimSpace(string(keyPEM)), "\n"),
	}, nil
}

type GenCertOptions struct {
	// The domain name for the certificate.
	DomainNames stringList
	// The common name for the certificate.
	CommonName string
	// The organization name for the certificate.
	Organization string
	// Whether this certificate is a CA
	IsCA bool
	// Expire time of the certificate. Default value 3 months.
	Expire *time.Duration
}

func (opt *GenCertOptions) initDefaults() {
	if opt.Expire == nil {
		*opt.Expire = time.Hour * 24 * 90
	}

	if len(opt.CommonName) == 0 {
		opt.CommonName = "Xray Inc"
	}

	if len(opt.Organization) == 0 {
		opt.Organization = "Xray Inc"
	}
}

type Cert struct {
	Certificate []string `json:"certificate"`
	Key         []string `json:"key"`
}

type stringList []string

func (l *stringList) String() string {
	return "String list"
}

func (l *stringList) Set(v string) error {
	if v == "" {
		base.Fatalf("empty value")
	}
	*l = append(*l, v)

	return nil
}
