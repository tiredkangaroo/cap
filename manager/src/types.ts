import { Timing } from "@/timing";
import { Proxy } from "./api/api";
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

    // require_approval is a boolean that determines whether the proxy should require approval for each request.
    require_approval: boolean;

    // get_client_process_info is a boolean that determines whether the proxy should provide information
    // about the client process. Getting this information can take a significant amount of time.
    get_client_process_info: boolean;

    // timeline_based_state_updates is a boolean that determines whether the proxy should use timeline-based
    // state updates.
    timeline_based_state_updates: boolean;
}

export interface Request {
    id: string;
    starred: boolean;
    datetime: number;
    secure: boolean;
    clientIP: string;
    clientApplication: string;
    clientAuthorization: string;
    clientAuthorizationUser?: string;
    clientAuthorizationPassword?: string;
    host: string;

    method?: string;
    path?: string;
    query?: Record<string, Array<string>>;
    headers?: Record<string, Array<string>>;
    bodyLength?: number;
    tempBody?: string;

    response?: {
        statusCode?: number;
        headers?: Record<string, Array<string>>;
        bodyLength?: number;
        tempBody?: string;
    };

    bytesTransferred?: number;

    state: string;

    timing?: Timing;
    timing_total?: number;

    error?: string;
}

export interface RequestsViewConfig {
    hideDate: boolean;
    hideHostCollapsed: boolean;
    hideClientApplication: boolean;
    hideState: boolean;

    hideError: boolean;
    hideID: boolean;
    hideHost: boolean; // not implemented yet
    hideClientIP: boolean; // not implemented yet
    hideClientUser: boolean;
    hideClientPassword: boolean;
    hideMethod: boolean;
    hidePath: boolean;
    hideQuery: boolean;
    hideRequestHeaders: boolean;
    hideRequestBody: boolean;
    hideResponseStatus: boolean;
    hideResponseHeaders: boolean;
    hideResponseBody: boolean;
    hideBytesTransferred: boolean;
}

export interface AppearanceConfig {
    darkModeSystem: boolean;
}

export interface RequestContentProps {
    request: Request;
    proxy: Proxy;
    editMode: boolean;
    setEditMode: React.Dispatch<React.SetStateAction<boolean>>;
    requestsViewConfig: RequestsViewConfig;
    setRequest: (req: Request) => void;
}

export interface FilterField {
    name: string;
    verboseName: string;
    type: "string" | "number" | "bool";
    uniqueValues: (string | number | boolean)[]; // uniqueValues should NOT be accessed for "boolean" type filters
    selectedValue?: string | number | boolean; // selectedValue is optional for "boolean" type filters
}

export type FilterType = FilterField[];
