import { Request } from "@/types";
import { Timing } from "@/timing";

interface IDMessage {
    id: string;
}

export class ClientWS {
    ws: WebSocket | null = null;
    paused: boolean = false;
    messageQueue: Array<MessageEvent> = [];

    constructor() {
        this.ws = null;
    }

    getActionData(ev: MessageEvent): [string, object] {
        const sp = ev.data.split(" ");
        return [sp[0], JSON.parse(sp.slice(1).join(" "))];
    }

    isOpen(): boolean {
        return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
    }

    pauseLiveActions(): void {
        this.paused = true;
    }
    resumeLiveActions(requests: Array<Request>): Array<Request> {
        for (const event of this.messageQueue) {
            requests = this.handleMessage(event, requests);
        }
        this.paused = false;
        this.messageQueue = [];
        return requests;
    }

    onmessage = (
        event: MessageEvent,
        requests: Array<Request>,
    ): [Array<Request>, boolean] => {
        if (this.paused) {
            this.messageQueue.push(event);
            return [requests, false];
        }
        return [this.handleMessage(event, requests), true];
    };

    handleMessage = (
        event: MessageEvent,
        requests: Array<Request>,
    ): Array<Request> => {
        const [action, rawdata] = this.getActionData(event);
        switch (action) {
            case "NEW": {
                const data = rawdata as Request;
                data.state = "Created";
                // Check if the request already exists
                const existingIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                [
                    data.clientAuthorizationUser,
                    data.clientAuthorizationPassword,
                ] = clientAuthorizationToUserPass(data.clientAuthorization);
                if (existingIndex !== -1) {
                    // Update the existing request
                    requests[existingIndex] = data;
                    console.warn(
                        `Request with ID ${data.id} already exists, updating it.`,
                    );
                } else {
                    requests = [data, ...requests]; // add request at the start because newest requests are first
                }

                break;
            }
            case "STATE": {
                const data = rawdata as {
                    id: string;
                    state: string;
                };
                const requestIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.state = data.state;
                    requests[requestIndex] = request;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }
            case "TUNNEL": {
                const data = rawdata as IDMessage;
                const requestIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.state = "Processing";
                    requests[requestIndex] = request;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }
            case "REQUEST": {
                const data = rawdata as {
                    id: string;
                    method: string;
                    path: string;
                    query: Record<string, Array<string>>;
                    headers: Record<string, Array<string>>;
                    bodyID: string;
                    bodyLength: number;
                    bytesTransferred: number;
                };
                const requestIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.state = "Processing";
                    request.method = data.method;
                    request.path = data.path;
                    request.query = data.query;
                    request.headers = data.headers;
                    request.bodyLength = data.bodyLength;
                    request.bytesTransferred = data.bytesTransferred;
                    requests[requestIndex] = request;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }
            case "SET-BYTES-TRANSFERRED": {
                const data = rawdata as {
                    id: string;
                    bytesTransferred: number;
                };
                const requestIndex = requests.findIndex((r) => r.id == data.id);
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.bytesTransferred = data.bytesTransferred;
                    requests[requestIndex] = request;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }
            case "RESPONSE": {
                const data = rawdata as {
                    id: string;
                    statusCode: number;
                    headers: Record<string, Array<string>>;
                    bodyID: string;
                    bodyLength: number;
                };
                const requestIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.response = {
                        statusCode: data.statusCode,
                        headers: data.headers,
                        bodyLength: data.bodyLength,
                    };
                    requests[requestIndex] = request;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }
            case "ERROR": {
                const data = rawdata as {
                    id: string;
                    error: string;
                    bytesTransferred: number;
                    timing: Timing;
                    timing_total: number;
                };
                const requestIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.state = "Error";
                    request.bytesTransferred = data.bytesTransferred;
                    request.timing_total = data.timing_total;
                    request.timing = data.timing;
                    request.error = data.error;
                    requests[requestIndex] = request;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }
            case "APPROVAL-WAIT": {
                const data = rawdata as { id: string };
                const requestIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.state = "Waiting Approval";
                    requests[requestIndex] = request;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }
            case "APPROVAL-RECIEVED": {
                const data = rawdata as { id: string };
                const requestIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.state = "Processing";
                    requests[requestIndex] = request;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }
            case "APPROVAL-CANCELED": {
                const data = rawdata as { id: string };
                const requestIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.state = "Canceled";
                    requests[requestIndex] = request;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }
            case "DONE": {
                const data = rawdata as {
                    id: string;
                    bytesTransferred: number;
                    timing: Timing;
                    timing_total: number;
                };
                const requestIndex = requests.findIndex(
                    (r) => r.id === data.id,
                );
                if (requestIndex !== -1) {
                    const request = requests[requestIndex];
                    request.state = "Done";
                    request.timing_total = data.timing_total;
                    request.timing = data.timing;
                } else {
                    console.warn(`Request with ID ${data.id} not found.`);
                }
                break;
            }

            default:
                break;
        }
        return requests;
    };

    approveRequest(id: string): void {
        if (this.ws == null || !this.isOpen()) {
            console.error("WebSocket is not initialized or not open");
            return;
        }
        const data: IDMessage = {
            id: id,
        };
        this.ws.send(`APPROVAL-APPROVE ${JSON.stringify(data)}`);
    }

    cancelRequest(id: string): void {
        if (this.ws == null || !this.isOpen()) {
            console.error("WebSocket is not initialized or not open");
            return;
        }
        const data: IDMessage = {
            id: id,
        };
        this.ws.send(`APPROVAL-CANCEL ${JSON.stringify(data)}`);
    }

    updateRequest(request: Request): void {
        if (this.ws == null || !this.isOpen()) {
            console.error("WebSocket is not initialized or not open");
            return;
        }
        const data = {
            id: request.id,
            request: request,
        };
        this.ws.send(`UPDATE-REQUEST ${JSON.stringify(data)}`);
    }
}

function clientAuthorizationToUserPass(
    authorization: string,
): [string, string] {
    if (!authorization.startsWith("Basic ")) {
        return ["", ""];
    }
    const base64 = authorization.slice(6); // Remove "Basic " prefix
    const decoded = atob(base64);
    const [user, password] = decoded.split(":");
    return [user, password];
}
