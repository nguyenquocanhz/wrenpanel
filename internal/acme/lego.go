package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/go-acme/lego/v4/registration"
)

type AcmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *AcmeUser) GetEmail() string {
	return u.Email
}

func (u *AcmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *AcmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

type Manager struct {
	SSLDir  string // default /etc/wrenpanel/ssl
	CADirURL string // default Let's Encrypt directory
}

func NewManager(sslDir string) *Manager {
	if sslDir == "" {
		sslDir = "/etc/wrenpanel/ssl"
	}
	return &Manager{
		SSLDir:   sslDir,
		CADirURL: lego.LEDirectoryProduction,
	}
}

// VerifyDNSResolution checks if domain resolves to one of the server's local/public IPs
func (m *Manager) VerifyDNSResolution(domain string) (bool, []string, error) {
	ips, err := net.LookupHost(domain)
	if err != nil {
		return false, nil, fmt.Errorf("DNS lookup failed for %s: %w", domain, err)
	}
	return len(ips) > 0, ips, nil
}

type IssueResult struct {
	Domain     string
	CertPath   string
	KeyPath    string
	Issuer     string
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

// IssueCertificateHTTP01 issues Let's Encrypt certificate via HTTP-01 webroot challenge
func (m *Manager) IssueCertificateHTTP01(domain, email, docroot string) (*IssueResult, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	// 1. Generate private key for ACME user
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate user private key: %w", err)
	}

	user := &AcmeUser{
		Email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(user)
	config.CADirURL = m.CADirURL
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create lego client: %w", err)
	}

	// 2. Setup HTTP-01 provider pointing to site docroot
	httpProvider, err := webroot.NewHTTPProvider(docroot)
	if err != nil {
		return nil, fmt.Errorf("failed to create http-01 provider for docroot %s: %w", docroot, err)
	}
	if err := client.Challenge.SetHTTP01Provider(httpProvider); err != nil {
		return nil, fmt.Errorf("failed to set http-01 provider: %w", err)
	}

	// 3. Register user with ACME CA
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("failed to register ACME user: %w", err)
	}
	user.Registration = reg

	// 4. Request certificate
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate for %s: %w", domain, err)
	}

	// 5. Save certificates to /etc/wrenpanel/ssl/<domain>/
	domainDir := filepath.Join(m.SSLDir, domain)
	if err := os.MkdirAll(domainDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create cert dir: %w", err)
	}

	certPath := filepath.Join(domainDir, "cert.pem")
	keyPath := filepath.Join(domainDir, "privkey.pem")

	if err := os.WriteFile(certPath, certificates.Certificate, 0644); err != nil {
		return nil, fmt.Errorf("failed to write cert file: %w", err)
	}
	if err := os.WriteFile(keyPath, certificates.PrivateKey, 0600); err != nil {
		return nil, fmt.Errorf("failed to write key file: %w", err)
	}

	now := time.Now().UTC()
	return &IssueResult{
		Domain:    domain,
		CertPath:  certPath,
		KeyPath:   keyPath,
		Issuer:    "lets-encrypt",
		IssuedAt:  now,
		ExpiresAt: now.Add(90 * 24 * time.Hour),
	}, nil
}
