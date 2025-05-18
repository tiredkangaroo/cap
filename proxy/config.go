package main

// Config is the configuration for the proxy.
type Config struct {
	// Debug is a boolean that determines whether the proxy should run in debug mode.
	Debug bool `json:"debug"`
	// RealIPHeader is a boolean that determines whether the proxy should add the IP
	// address of the client in the X-Forwarded-For header.
	RealIPHeader bool `json:"real_ip_header"`
	// CertificateLifetime is the lifetime of the certificate in hours. It is possible
	// to set it to less than 0 in which any new certificates generated will not be valid.
	CertificateLifetime int `json:"certificate_lifetime"`
	// MITM determines who is responsible for the TLS connection. If true, the responsibility
	// is on the proxy. If false, the responsibility is on the client.
	//
	// MITM is useful for inspecting the traffic between the client and the server. However,
	// it is more resource intensive to generate and store the certificates for each host, perform a
	// TLS handshake as well as to decrypt the traffic, reencrypt it, move requests and responses.
	MITM bool `json:"mitm"`
	// ProvideRequestBody is a boolean that determines whether the proxy should provide the request body.
	// It can be useful for debugging purposes, however is very resource intensive, especially with larger
	// bodies.
	ProvideRequestBody bool `json:"provide_request_body"`
}
