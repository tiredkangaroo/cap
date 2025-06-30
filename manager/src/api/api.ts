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

    async getFilterCounts(): Promise<Record<string, Array<string>>> {
        const response = await fetch(`${this.url}/filterCounts`);
        if (!response.ok) {
            throw new Error(
                `failed to fetch filter counts: ${response.statusText}`,
            );
        }
        const respJSON = await response.json();
        return respJSON;
    }

    async getRequestBody(id: string): Promise<string> {
        const response = await fetch(`${this.url}/reqbody/${id}`);
        if (!response.ok) {
            throw new Error(
                `failed to fetch request body: ${response.statusText}`,
            );
        }
        const body = await response.text();
        return body;
    }

    async getResponseBody(id: string): Promise<string> {
        const response = await fetch(`${this.url}/respbody/${id}`);
        if (!response.ok) {
            throw new Error(
                `failed to fetch response body: ${response.statusText}`,
            );
        }
        const body = await response.text();
        return body;
    }

    async getRequestsWithFilter(
        filter: Record<string, string | undefined>,
        offset: number,
        limit: number,
    ): Promise<[Array<Request>, number]> {
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
        const respJSON = await response.json();
        return [respJSON.requests, respJSON.total];
    }

    async setRequestStarred(id: string, starred: boolean): Promise<void> {
        const response = await fetch(
            `${this.url}/setRequestStarred/${id}?starred=${starred}`,
            {
                method: "POST",
            },
        );
        if (!response.ok) {
            throw new Error(
                `failed to set request starred: ${response.statusText}`,
            );
        }
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
        this.getRequestsWithFilter({}, 0, 100)
            .then(() => {
                uCB();
            })
            .catch((err) => {
                console.error(
                    "get requests with filter initial on manage requests",
                    err,
                );
            });
        // setRequests: React.Dispatch<React.SetStateAction<Array<Request>>>,
        this.clientWS.ws = new WebSocket(`${this.url}/requestsWS`);
        if (!this.clientWS.ws) {
            console.error("failed to create websocket");
            return;
        }
        this.clientWS.ws.onopen = () => {};
        this.clientWS.ws.onmessage = (event: MessageEvent) => {
            const [reqs, ok] = this.clientWS.onmessage(event, this.requests);
            if (ok) {
                this.requests = reqs;
                this.updateCB!();
            }
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
        req.response = newrequest.response;
        req.tempBody = newrequest.tempBody;

        this.clientWS.updateRequest(newrequest);
        this.requests[reqIndex] = req;
        this.updateCB!();
    }
}
