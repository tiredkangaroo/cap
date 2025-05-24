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

    // provide_request_body is a boolean that determines whether the proxy should provide the request body.
    // It can be useful for debugging purposes, however is very resource intensive, especially with larger
    // bodies.
    provide_request_body: boolean;

    // provide_response_body is a boolean that determines whether the proxy should provide the response body.
    // It can be useful for debugging purposes, however is very resource intensive, especially with larger
    // bodies.
    provide_response_body: boolean;

    // perform_delay is a number that determines the delay in milliseconds that the proxy should perform
    // before responding to a request. This can be useful for testing purposes, such as simulating
    // network latency. This number must be greater than or equal to 0.
    perform_delay: number;
}

export interface Request {
    id: string;
    secure: boolean;
    clientIP: string;
    clientAuthorizationUser?: string;
    clientAuthorizationPassword?: string;
    host: string;

    method?: string;
    headers?: Record<string, Array<string>>;
    body?: string;

    response?: {
        statusCode?: number;
        headers?: Record<string, Array<string>>;
        body?: string;
    };

    state: "Processing" | "Canceled" | "Done" | "Error";
    error?: string;
}
