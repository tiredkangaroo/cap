package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

type Certificates struct {
	caCert *x509.Certificate
	caKey  any
}

func (c *Certificates) Init() error {
	rawproxyCACert, err := os.ReadFile(os.Getenv("PROXY_CACERT"))
	if err != nil {
		return fmt.Errorf("proxy cacert readfile: %w", err)
	}
	rawproxyCAKey, err := os.ReadFile(os.Getenv("PROXY_CAKEY"))
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

func (c *Certificates) getTLSCert(host string) (tls.Certificate, error) {
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
		NotAfter:              time.Now().Add(time.Hour * 7200),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// create certificate
	cert, err := x509.CreateCertificate(rand.Reader, config, c.caCert, &pk.PublicKey, c.caKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("creating the x509 certificate: %w", err)
	}

	// encode certificate and private key
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	if pemCert == nil {
		return tls.Certificate{}, fmt.Errorf("encode the cert with pe (unknown error)")
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(pk)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("marshal private key: %w", err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemKey == nil {
		return tls.Certificate{}, fmt.Errorf("encode the private key with pem (unknown error)")
	}

	return tls.X509KeyPair(pemCert, pemKey)
}

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
