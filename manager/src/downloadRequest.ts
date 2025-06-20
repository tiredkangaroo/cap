import { mimeToExtension } from "./mime";
import { Request } from "./types";

export function downloadRequest(request: Request): void {
    const blob = new Blob([JSON.stringify(request)], { type: "text/plain" });
    dlBlob(blob, `${request.id}.json`);
}

export function downloadBody(
    id: string,
    body: string | null | undefined,
    contentTypeHeader: string | undefined,
): void {
    if (body == null || body == undefined) {
        console.error("request body is null/undefined, cannot download");
        return;
    }
    let type = "text/plain";
    if (contentTypeHeader != undefined && contentTypeHeader != "") {
        type = contentTypeHeader;
    }

    const blob = new Blob([body], {
        type: type,
    });
    const filename = `${id}_body.${mimeToExtension[type]}`;
    dlBlob(blob, filename);
}

function dlBlob(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();

    // cleanup
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}
