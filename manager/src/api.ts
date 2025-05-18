import { Config, Request } from "./types.ts";

export class Proxy {
    loaded: boolean = false;
    name: string;
    url: string;
    config: Config;
    ws: WebSocket | null = null; // this will be used later
    requests: Array<Request> = [];
    // token for accessing the proxy will be set here as well

    constructor(name: string, url: string) {
        this.name = name;
        this.url = url;
        this.config = {
            real_ip_header: false,
            certificate_lifetime: 0,
            mitm: false,
            provide_request_body: false,
        };
    }

    async Init(): Promise<void> {
        this.config = await this.getConfig();
        this.loaded = true;
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
    manageRequests(updateCB: () => void) {
        // setRequests: React.Dispatch<React.SetStateAction<Array<Request>>>,
        const ws = new WebSocket(`${this.url}/requestsWS`);
        if (!ws) {
            console.error("failed to create websocket");
            return;
        }
        if (this.ws != ws) {
            this.ws?.close();
        }
        this.ws = ws;

        this.ws.onmessage = (event: MessageEvent) => {
            const sp = event.data.split(" ");
            // FIXME: there may be a bug where a NEW is recieved multiple times
            switch (sp[0]) {
                case "NEW": {
                    const rawdata = sp.slice(1).join(" ");
                    const data = JSON.parse(rawdata);
                    console.log("new", data.id);
                    const request: Request = {
                        id: data.id,
                        secure: data.secure,
                        clientIP: data.clientIP,
                        clientAuthorization: data.clientAuthorization,
                        host: data.host,
                    };
                    const r = this.requests.find(
                        (v: Request) => v.id == request.id,
                    );
                    if (r != undefined) {
                        return;
                    }

                    this.requests = [request, ...this.requests];
                    updateCB();
                    break;
                }
                case "HTTP": {
                    const rawdata = sp.slice(1).join(" ");
                    const data = JSON.parse(rawdata);

                    const reqs = this.requests;
                    const req = reqs.find((v: Request) => v.id == data.id);
                    if (req == undefined) {
                        console.error("got info for a request we don't have");
                        return;
                    }
                    const i = reqs.indexOf(req);
                    req.method = data.method;
                    req.headers = data.headers;
                    req.body = data.body;
                    reqs.splice(i, 1, req);

                    this.requests = reqs;
                    updateCB();
                    break;
                }
                case "HTTPS-MITM": {
                    const rawdata = sp.slice(1).join(" ");
                    const data = JSON.parse(rawdata);

                    const reqs = this.requests;
                    const req = reqs.find((v: Request) => v.id == data.id);
                    if (req == undefined) {
                        console.log(data.id);
                        console.error("got info for a request we don't have");
                        return;
                    }
                    const i = reqs.indexOf(req);
                    req.method = data.method;
                    req.headers = data.headers;
                    req.body = data.body;
                    reqs.splice(i, 1, req);

                    this.requests = reqs;
                    updateCB();
                    break;
                }
                default:
                    console.log("unknown message", event.data);
            }
        };
    }
}
