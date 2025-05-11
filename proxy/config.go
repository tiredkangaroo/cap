package main

// Config is the configuration for the proxy.
type Config struct {
	// RealIPHeader is a boolean that determines whether the proxy should add the IP
	// address of the client in the X-Forwarded-For header.
	RealIPHeader bool `json:"real_ip_header"`
	// CertificateLifetime is the lifetime of the certificate in hours. It is possible
	// to set it to less than 0 in which any new certificates generated will not be valid.
	CertificateLifetime int `json:"certificate_lifetime"`
}
