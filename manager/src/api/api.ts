import { objectToQueryString } from "@/utils.ts";
import { Config, Request } from "../types.ts";
import { ClientWS } from "./ws.ts";

export class Proxy {
    loaded: boolean = false;
    name: string;
    url: string;
    config: Config;
    // ws: WebSocket | null = null; // this will be used later
    clientWS: ClientWS; // this will be used later
    requests: Array<Request> = [];
    updateCB: (() => void) | null = null; // callback to update the requests view
    // token for accessing the proxy will be set here as well

    constructor(name: string, url: string) {
        this.name = name;
        this.url = url;
        this.config = {
            real_ip_header: false,
            certificate_lifetime: 0,
            mitm: false,
            provide_request_body: false,
            provide_response_body: false,
            perform_delay: 0,
            require_approval: false,
            get_client_process_info: false,
        };
        this.clientWS = new ClientWS();
    }

    async Init(): Promise<void> {
        this.config = await this.getConfig();
        this.loaded = true;
    }

    async load(): Promise<void> {
        console.log(this.name, this.url);
        const offset = this.requests.length; // storing num in a variable prevents race condition where
        // the requests is changed later on, and .then() for json adds it to another plae
        fetch(`${this.url}/requests?offset=${offset.toString()}&limit=100`)
            .then((data) => {
                if (!data.ok) {
                    console.error("failed to fetch requests:", data.statusText);
                    return;
                }
                data.json()
                    .then((requests: Array<Request>) => {
                        this.requests = [
                            ...this.requests.slice(0, offset),
                            ...requests, // why do i need to reverse here? why isn't it reversed? the server is sending them correct so is console logging the requests
                            ...this.requests.slice(offset),
                        ];
                        this.updateCB!();
                    })
                    .catch((err) => {
                        console.error("failed to parse requests:", err);
                    });
                this.updateCB!();
            })
            .catch((err) => {
                console.error(err);
            });
    }

    async getRequestsWithFilter(
        filter: Record<string, string | undefined>,
        offset: number,
        limit: number,
    ): Promise<Array<Request>> {
        const response = await fetch(
            `${this.url}/requestsMatchingFilter?${objectToQueryString({
                ...filter,
                offset: offset.toString(),
                limit: limit.toString(),
            })}`,
        );
        if (!response.ok) {
            throw new Error(`failed to fetch requests: ${response.statusText}`);
        }
        const requests = await response.json();
        return requests;
    }

    async getConfig(): Promise<Config> {
        const response = await fetch(`${this.url}/config`);
        if (!response.ok) {
            throw new Error(`failed to fetch config: ${response.statusText}`);
        }
        const config = await response.json();
        this.config = config;
        this.loaded = true;
        return config;
    }

    async setConfig(config: Config): Promise<void> {
        const response = await fetch(`${this.url}/config`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(config),
        });
        if (!response.ok) {
            throw new Error(`failed to set config: ${response.statusText}`);
        }
    }
    manageRequests(uCB: () => void) {
        // get requests from the server first
        this.updateCB = uCB;
        fetch(`${this.url}/requests?offset=0&limit=100`)
            .then((data) => {
                if (!data.ok) {
                    console.error("failed to fetch requests:", data.statusText);
                    return;
                }
                data.json()
                    .then((requests: Array<Request>) => {
                        this.requests = requests;
                        this.updateCB!();
                    })
                    .catch((err) => {
                        console.error("failed to parse requests:", err);
                    });
                uCB();
            })
            .catch((err) => {
                console.error(err);
            });
        // setRequests: React.Dispatch<React.SetStateAction<Array<Request>>>,
        this.clientWS.ws = new WebSocket(`${this.url}/requestsWS`);
        if (!this.clientWS.ws) {
            console.error("failed to create websocket");
            return;
        }
        this.clientWS.ws.onopen = () => {
            console.log("WebSocket connection established");
        };
        this.clientWS.ws.onmessage = (event: MessageEvent) => {
            this.requests = this.clientWS.onmessage(event, this.requests);
            this.updateCB!();
        };
        this.clientWS.ws.onclose = (event: CloseEvent) => {
            console.log("webSocket connection closed:", event);
        };
    }

    approveRequest(id: string): void {
        this.clientWS.approveRequest(id);
    }
    cancelRequest(id: string): void {
        this.clientWS.cancelRequest(id);
    }

    updateRequest(newrequest: Request): void {
        const reqIndex = this.requests.findIndex((r) => r.id == newrequest.id);
        if (reqIndex === -1) {
            console.warn(`Request with ID ${newrequest.id} not found.`);
            return;
        }
        const req = this.requests[reqIndex];

        req.host = newrequest.host;
        req.method = newrequest.method;
        req.headers = newrequest.headers;
        req.path = newrequest.path;
        req.query = newrequest.query;
        req.body = newrequest.body;
        req.response = newrequest.response;

        this.clientWS.updateRequest(newrequest);
        this.requests[reqIndex] = req;
        this.updateCB!();
    }
}

// async function decodedBody(
//     headers: Record<string, Array<string>>,
//     body: string,
// ): Promise<string> {
//     if (
//         headers["Content-Encoding"] == undefined ||
//         !headers["Content-Encoding"].includes("gzip")
//     ) {
//         return body;
//     }

//     // Convert string -> Uint8Array
//     const binaryData = new Uint8Array(body.length);
//     for (let i = 0; i < body.length; i++) {
//         binaryData[i] = body.charCodeAt(i);
//     }

//     // Create ReadableStream from binary
//     const stream = new ReadableStream({
//         start(controller) {
//             controller.enqueue(binaryData);
//             controller.close();
//         },
//     });

//     // Decompress using DecompressionStream
//     const decompressedStream = stream.pipeThrough(
//         new DecompressionStream("gzip"),
//     );

//     const text = await new Response(decompressedStream).text();
//     return text;
// }
