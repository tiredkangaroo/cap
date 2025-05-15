import { useEffect, useState } from "react";
import { Proxy } from "./api";
import { Switch } from "./components/ui/switch";
import { Config } from "./types";
import {
    HoverCard,
    HoverCardContent,
    HoverCardTrigger,
} from "./components/ui/hover-card";

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
        <div className="flex flex-col items-center w-full h-full">
            <div className="text-center mt-4">
                <h1 className="text-4xl font-bold mb-1">{proxy.name}</h1>
                <p className="text-md">URL: {proxy.url}</p>
            </div>
            <div className="flex flex-col w-[50%] min-w-fit h-full justify-center">
                <div className="flex flex-row text-3xl">
                    <HoverCard>
                        <HoverCardTrigger className="self-start">
                            <p className="text-3xl mt-auto mb-auto">
                                MITM {proxyConfig.mitm}
                            </p>
                        </HoverCardTrigger>
                        <HoverCardContent className="min-w-fit w-[85%]">
                            MITM set on indicates that the proxy is responsible
                            for the TLS connection, and will use its own
                            certificates to encrypt the traffic.
                        </HoverCardContent>
                    </HoverCard>
                    <Switch
                        className="mt-auto mb-auto ml-auto mr-2"
                        checked={proxyConfig?.mitm}
                        onCheckedChange={(checked: boolean) => {
                            proxyConfig.mitm = checked;
                            proxy.setConfig(proxyConfig);
                            const newObj = Object.assign({}, proxyConfig);
                            setProxyConfig(newObj);
                        }}
                    ></Switch>
                </div>
                <div className="flex flex-row text-3xl mt-6">
                    <HoverCard>
                        <HoverCardTrigger>
                            <p className="text-3xl mt-auto mb-auto">
                                Real IP Header {proxyConfig.real_ip_header}
                            </p>
                        </HoverCardTrigger>
                        <HoverCardContent className="min-w-fit w-[85%]">
                            Real IP Header set on means that the proxy will add
                            the IP address of the client in the X-Forwarded-For
                            header.
                        </HoverCardContent>
                    </HoverCard>
                    <Switch
                        className="mt-auto mb-auto ml-auto mr-2"
                        checked={proxyConfig?.real_ip_header}
                        onCheckedChange={(checked: boolean) => {
                            proxyConfig.real_ip_header = checked;
                            proxy.setConfig(proxyConfig);
                            const newObj = Object.assign({}, proxyConfig);
                            setProxyConfig(newObj);
                        }}
                    ></Switch>
                </div>
                <div className="flex flex-row text-3xl mt-6">
                    <HoverCard>
                        <HoverCardTrigger>
                            <p className="text-3xl mt-auto mb-auto">
                                Certificate Lifetime:{" "}
                            </p>
                        </HoverCardTrigger>
                        <HoverCardContent className="min-w-fit w-[85%]">
                            The lifetime of the certificate in hours. It is
                            possible to set it to less than 0 in which any new
                            certificates generated will not be valid.
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
                            proxy.setConfig(proxyConfig);
                            const newObj = Object.assign({}, proxyConfig);
                            setProxyConfig(newObj);
                        }}
                    ></input>
                </div>
            </div>
        </div>
    );
}
