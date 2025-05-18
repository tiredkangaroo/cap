import { useEffect, useState } from "react";

import { Proxy } from "./api";
import { Config } from "./types";

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
        </div>
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
            <DialogTrigger>
                <button className="bg-gray-800 p-4 rounded-xl">
                    <IoSettingsSharp className="text-white" />
                </button>
            </DialogTrigger>
            <DialogContent>
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
