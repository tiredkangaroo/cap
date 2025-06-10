package certificate

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"sync"
	"time"

	"github.com/tiredkangaroo/bigproxy/proxy/config"
)

// Certificates is a struct that provides services for making TLS connections.
type Certificates struct {
	// caCert is the CA certificate used to sign the certificates for the hosts.
	caCert *x509.Certificate
	// caKey is the CA private key used to sign the certificates for the hosts.
	caKey any

	// cache is a map of hostnames to TLS certificates. A sync.Map is used over
	// a regular map plus a lock because it is more performant for the specific
	// use case of this proxy and reduces lock contention.
	//
	// From the docs of sync.Map: "The Map type is optimized for two common use cases:
	// (1) when the entry for a given key is only ever written once but read many times,
	// as in caches that only grow [...]."
	cache sync.Map
}

// Init initializes the Certificates struct by reading the CA certificate from the
// environment variable PROXY_CACERT and the CA private key from the environment
// variable PROXY_CAKEY. It parses the certificate and key and stores them in
// the caCert and caKey fields respectively.
func (c *Certificates) Init() error {
	proxyCACert := os.Getenv("PROXY_CACERT")
	proxyCAKey := os.Getenv("PROXY_CAKEY")

	// default values for the CA certificate and key will be used when the user can manually drop in proxy cacert and cakey
	// files
	if proxyCACert == "" {
		proxyCACert = "certs/ca.crt"
		slog.Warn("PROXY_CACERT environment variable is not set, using default: certs/ca.crt")
	}
	if proxyCAKey == "" {
		proxyCAKey = "certs/ca.key"
		slog.Warn("PROXY_CAKEY environment variable is not set, using default: certs/ca.key")
	}
	rawproxyCACert, err := os.ReadFile(proxyCACert)
	if err != nil {
		return fmt.Errorf("proxy cacert readfile: %w", err)
	}
	rawproxyCAKey, err := os.ReadFile(proxyCAKey)
	if err != nil {
		return fmt.Errorf("proxy cakey readfile: %w", err)
	}

	certBlock, _ := pem.Decode(rawproxyCACert)
	keyBlock, _ := pem.Decode(rawproxyCAKey)

	c.caCert, err = x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("parse cert: %w", err)
	}

	c.caKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("parse key: %w", err)
	}

	return nil
}

// getTLSCert generates a new TLS certificate for the given host. It first checks if
// the certificate is already in the cache. If it is, it returns the cached
// certificate. If it is not, it generates a new certificate and stores it in the
// cache.
//
// The certificate is valid for the lifetime specified in the config (CertificateLifetime).
func (c *Certificates) getTLSCert(host string) (tls.Certificate, error) {
	// check if the certificate is already in the cache
	cachedCertificate, ok := c.cache.Load(host)
	if ok {
		if time.Until(cachedCertificate.(tls.Certificate).Leaf.NotAfter) < time.Minute {
			c.cache.Delete(host) // delete the certificate if it is expired or about to expire
		} else {
			return cachedCertificate.(tls.Certificate), nil
		}
	}

	// generate a new private key for the new certificate
	pk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("an error occured while attempting to generate an ecdsa key: %s", err.Error())
	}

	// create a serial number for certificate
	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("an error occured while attempting to generate a certificate serial number: %s", err.Error())
	}

	// create cert config
	config := &x509.Certificate{
		SerialNumber: sn,
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"N/A"},
		},
		DNSNames:              []string{host},
		NotBefore:             time.Now().Add(-(time.Hour * 7200)),
		NotAfter:              time.Now().Add(time.Hour * time.Duration(config.DefaultConfig.CertificateLifetime)),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// create certificate
	cert, err := x509.CreateCertificate(rand.Reader, config, c.caCert, &pk.PublicKey, c.caKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("creating the x509 certificate: %w", err)
	}

	// encode certificate
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	if pemCert == nil {
		return tls.Certificate{}, fmt.Errorf("encode the cert with pe (unknown error)")
	}

	// encode the private key
	privBytes, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("marshal private key: %w", err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemKey == nil {
		return tls.Certificate{}, fmt.Errorf("encode the private key with pem (unknown error)")
	}

	// create the certificate
	tlscert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create x509 key pair: %w", err)
	}

	// store the certificate in the cache
	c.cache.Store(host, tlscert)
	return tlscert, nil
}

// TLSConn creates a new TLS connection using the given config and connection signed for the specified host.
func (c *Certificates) TLSConn(conn net.Conn, host string) (*tls.Conn, error) {
	cert, err := c.getTLSCert(host)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:               tls.VersionTLS10,
		MaxVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{cert},
	}
	return tls.Server(conn, tlsConfig), nil
}
