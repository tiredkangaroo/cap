import { useEffect, useState } from "react";

import { Proxy } from "./api";
import { Config, Request } from "./types";

import { Switch } from "./components/ui/switch";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "./components/ui/dialog";
import {
    HoverCard,
    HoverCardContent,
    HoverCardTrigger,
} from "./components/ui/hover-card";

import { IoSettingsSharp } from "react-icons/io5";
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from "./components/ui/collapsible";

export function ProxyView() {
    const [proxy, _] = useState<Proxy | null>(
        new Proxy("Proxy 1", "http://localhost:8001"),
    );
    const [proxyConfig, setProxyConfig] = useState<Config | null>(null);

    //NOTE: async/await in set state functions
    useEffect(() => {
        if (proxy === null) return;
        const load = async () => {
            await proxy.Init();
            setProxyConfig(proxy.config);
        };
        load();
    }, []);

    if (proxy === null || proxyConfig == null) {
        return (
            <div className="flex flex-col justify-center items-center w-full h-full">
                <h1 className="text-4xl font-bold mb-1">No Proxy</h1>
            </div>
        );
    }

    return (
        <div className="flex flex-col w-full h-full">
            <div className="flex flex-row mt-4 w-full">
                <div className="ml-auto mr-auto text-center">
                    <h1 className="text-4xl font-bold mb-1">{proxy.name}</h1>
                    <p className="text-md">URL: {proxy.url}</p>
                </div>
                <div className="absolute right-2">
                    <SettingsView proxy={proxy} />
                </div>
            </div>
            <IncomingRequestsView proxy={proxy} />
        </div>
    );
}

function IncomingRequestsView(props: { proxy: Proxy }) {
    const [requests, setRequests] = useState<Array<Request>>([]);

    useEffect(() => {
        props.proxy.manageRequests(() => {
            console.log("update cb", props.proxy.requests);
            const newObj = Object.assign([], props.proxy.requests);
            setRequests(newObj);
        });
    }, []);

    return (
        <div className="flex flex-col w-full h-full">
            <h1 className="ml-2 text-2xl font-bold mb-2">Incoming Requests</h1>
            <div className="w-full overflow-y-auto h-[80%]">
                {requests.map((request, index) => (
                    <RequestView key={index} request={request} />
                ))}
            </div>
        </div>
    );
}

function RequestView(props: { request: Request }) {
    console.log(props.request);
    return (
        <Collapsible className="border-b-1 border-b-black wrap-anywhere">
            <CollapsibleTrigger className="w-full bg-gray-200">
                <div className="relative flex flex-row w-full pt-4 pb-4">
                    <p className="flex-1">
                        {props.request.secure ? "HTTPS" : "HTTP"}
                    </p>
                    <p className="flex-1 ml-1 text-md">{props.request.id}</p>
                    <p className="flex-1 ml-1 text-center">
                        {props.request.host}
                    </p>
                    <p className="flex-1 ml-1 mr-1">{props.request.clientIP}</p>
                </div>
            </CollapsibleTrigger>
            <CollapsibleContent className="bg-gray-300">
                <div className="ml-2 pt-2 pb-1">
                    {props.request.method ? (
                        <p className="mb-2">
                            <b>Method: </b> {props.request.method}
                        </p>
                    ) : (
                        <></>
                    )}
                    <p className="mb-2">
                        <b>Client Authorization: </b>
                        {props.request.clientAuthorization != "" ? (
                            props.request.clientAuthorization
                        ) : (
                            <i>No Client Authorization</i>
                        )}
                    </p>
                    {props.request.headers ? (
                        <div className="mb-2">
                            <b>Headers:</b>
                            {Object.entries(props.request.headers!).map((v) => (
                                <p>
                                    {v[0]}: {v[1].join(", ")}
                                </p>
                            ))}
                        </div>
                    ) : (
                        <></>
                    )}
                    {props.request.body ? (
                        <div className="mb-2">
                            <p>
                                <b>Body: </b>
                            </p>
                            {props.request.body}
                        </div>
                    ) : (
                        <></>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

function SettingsView(props: { proxy: Proxy }) {
    const [proxyConfig, setProxyConfig] = useState<Config>(props.proxy.config);

    //NOTE: async/await in set state functions
    useEffect(() => {
        const load = async () => {
            await props.proxy.getConfig();
            setProxyConfig(props.proxy.config);
        };
        load();
    }, []);

    return (
        <Dialog>
            <DialogTrigger className="bg-gray-800 p-4 rounded-xl">
                <IoSettingsSharp className="text-white" />
            </DialogTrigger>
            <DialogContent className="w-full h-full">
                <DialogHeader>
                    <DialogTitle className="text-center">Settings</DialogTitle>
                    <DialogDescription className="text-center">
                        Configure the proxy settings.
                    </DialogDescription>
                </DialogHeader>
                <div className="flex flex-col min-w-fit h-full justify-center">
                    <div className="flex flex-row text-xl">
                        <HoverCard>
                            <HoverCardTrigger className="self-start">
                                <p className="text-xl mt-auto mb-auto">
                                    MITM {proxyConfig.mitm}
                                </p>
                            </HoverCardTrigger>
                            <HoverCardContent className="min-w-fit w-[85%]">
                                MITM set on indicates that the proxy is
                                responsible for the TLS connection, and will use
                                its own certificates to encrypt the traffic.
                            </HoverCardContent>
                        </HoverCard>
                        <Switch
                            className="mt-auto mb-auto ml-auto mr-2"
                            checked={proxyConfig?.mitm}
                            onCheckedChange={(checked: boolean) => {
                                proxyConfig.mitm = checked;
                                props.proxy.setConfig(proxyConfig);
                                const newObj = Object.assign({}, proxyConfig);
                                setProxyConfig(newObj);
                            }}
                        ></Switch>
                    </div>
                    <div className="flex flex-row text-xl mt-6">
                        <HoverCard>
                            <HoverCardTrigger>
                                <p className="text-xl mt-auto mb-auto">
                                    Real IP Header {proxyConfig.real_ip_header}
                                </p>
                            </HoverCardTrigger>
                            <HoverCardContent className="min-w-fit w-[85%]">
                                Real IP Header set on means that the proxy will
                                add the IP address of the client in the
                                X-Forwarded-For header.
                            </HoverCardContent>
                        </HoverCard>
                        <Switch
                            className="mt-auto mb-auto ml-auto mr-2"
                            checked={proxyConfig?.real_ip_header}
                            onCheckedChange={(checked: boolean) => {
                                proxyConfig.real_ip_header = checked;
                                props.proxy.setConfig(proxyConfig);
                                const newObj = Object.assign({}, proxyConfig);
                                setProxyConfig(newObj);
                            }}
                        ></Switch>
                    </div>
                    <div className="flex flex-row text-xl mt-6">
                        <HoverCard>
                            <HoverCardTrigger>
                                <p className="text-xl mt-auto mb-auto">
                                    Certificate Lifetime:{" "}
                                </p>
                            </HoverCardTrigger>
                            <HoverCardContent className="min-w-fit w-[85%]">
                                The lifetime of the certificate in hours. It is
                                possible to set it to less than 0 in which any
                                new certificates generated will not be valid.
                            </HoverCardContent>
                        </HoverCard>
                        <input
                            defaultValue={proxyConfig.certificate_lifetime}
                            type="number"
                            className="ml-auto mr-2 min-w-fit text-md border-2 border-black text-center"
                            onBlur={(e) => {
                                proxyConfig.certificate_lifetime = parseInt(
                                    e.target.value,
                                );
                                props.proxy.setConfig(proxyConfig);
                                const newObj = Object.assign({}, proxyConfig);
                                setProxyConfig(newObj);
                            }}
                        ></input>
                    </div>
                    <div className="flex flex-row text-xl mt-6">
                        <HoverCard>
                            <HoverCardTrigger>
                                <p className="text-xl mt-auto mb-auto">
                                    Request Body Dumping
                                </p>
                            </HoverCardTrigger>
                            <HoverCardContent className="min-w-fit w-[85%]">
                                Request body dumping allows inspection of the
                                request body by the client.
                            </HoverCardContent>
                        </HoverCard>
                        <Switch
                            className="mt-auto mb-auto ml-auto mr-2"
                            checked={proxyConfig?.provide_request_body}
                            onCheckedChange={(checked: boolean) => {
                                proxyConfig.provide_request_body = checked;
                                props.proxy.setConfig(proxyConfig);
                                const newObj = Object.assign({}, proxyConfig);
                                setProxyConfig(newObj);
                            }}
                        ></Switch>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
}
