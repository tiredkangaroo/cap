// Config is the configuration for the proxy.
export interface Config {
    // RealIPHeader is a boolean that determines whether the proxy should add the IP
    // address of the client in the X-Forwarded-For header.
    real_ip_header: boolean;
    // CertificateLifetime is the lifetime of the certificate in hours. It is possible
    // to set it to less than 0 in which any new certificates generated will not be valid.
    certificate_lifetime: number;
    // MITM determines who is responsible for the TLS connection. If true, the responsibility
    // is on the proxy. If false, the responsibility is on the client.
    //
    // MITM is useful for inspecting the traffic between the client and the server. However,
    // it is more resource intensive to generate and store the certificates for each host, perform a
    // TLS handshake as well as to decrypt the traffic, reencrypt it, move requests and responses.
    mitm: boolean;
}
