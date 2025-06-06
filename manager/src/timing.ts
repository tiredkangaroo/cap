export enum Timing {
    RequestInit,
    PrepRequest,
    WaitApproval,
    DelayPerform,
    RequestPerform,
    DumpResponse,
    WriteResponse,
    ProxyResponse,
    DialHost,
    ReadWriteTunnel,
    CertGenTLSHandshake,
    ReadParseRequest,
}

export function getTimingName(timing: Timing): string {
    switch (timing) {
        case Timing.RequestInit:
            return "Request Init";
        case Timing.PrepRequest:
            return "Prep Request";
        case Timing.WaitApproval:
            return "Wait Approval";
        case Timing.DelayPerform:
            return "Delay Perform";
        case Timing.RequestPerform:
            return "Request Perform";
        case Timing.DumpResponse:
            return "Dump Response";
        case Timing.WriteResponse:
            return "Write Response";
        case Timing.ProxyResponse:
            return "Proxy Response";
        case Timing.DialHost:
            return "Dial Host";
        case Timing.ReadWriteTunnel:
            return "R/W Tunnel";
        case Timing.CertGenTLSHandshake:
            return "Cert Gen + TLS Handshake";
        case Timing.ReadParseRequest:
            return "Read/Parse Request";
    }
}

export const TimesOrders = [
    [
        // HTTP
        Timing.RequestInit,
        Timing.PrepRequest,
        Timing.WaitApproval,
        Timing.DelayPerform,
        Timing.RequestPerform,
        Timing.DumpResponse,
        Timing.WriteResponse,
    ],
    [
        // HTTPS
        Timing.RequestInit,
        Timing.ProxyResponse,
        Timing.WaitApproval,
        Timing.DelayPerform,
        Timing.DialHost,
        Timing.ReadWriteTunnel,
    ],
    [
        // HTTPS with MITM
        Timing.RequestInit,
        Timing.CertGenTLSHandshake,
        Timing.ReadParseRequest,
        Timing.PrepRequest,
        Timing.WaitApproval,
        Timing.DelayPerform,
        Timing.RequestPerform,
        Timing.DumpResponse,
        Timing.WriteResponse,
    ],
];
