import { useEffect, useState } from "react";

import { Proxy } from "./api";
import { IncomingView } from "./IncomingView";

import { IoSettingsSharp } from "react-icons/io5";
import { Config } from "./types";
import { SettingsDialog } from "./SettingsDialog";

export function ProxyView() {
    const [proxy, _] = useState<Proxy | null>(
        new Proxy("Proxy 1", "http://localhost:8001"),
    );
    const [proxyConfig, setProxyConfig] = useState<Config | null>(null);
    const [settingsOpen, setSettingsOpen] = useState(false);

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
            <SettingsDialog
                proxy={proxy}
                open={settingsOpen}
                setOpen={setSettingsOpen}
            />
            <div className="flex flex-row mt-4 w-full">
                <div className="ml-auto mr-auto text-center">
                    <h1 className="text-4xl font-bold mb-1">{proxy.name}</h1>
                    <p className="text-md">URL: {proxy.url}</p>
                </div>
                <button
                    className="mr-2 bg-gray-800 min-w-fit h-11 w-11 rounded-xl"
                    onClick={() => setSettingsOpen(true)}
                >
                    <IoSettingsSharp className="ml-auto mr-auto text-white" />
                </button>
            </div>
            <IncomingView proxy={proxy} />
        </div>
    );
}
