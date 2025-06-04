import { useEffect, useState } from "react";

import { Proxy } from "./api/api";
import { IncomingView } from "./IncomingView";

import { IoSettingsSharp } from "react-icons/io5";
import { Config } from "./types";
import { SettingsDialog } from "./settings/SettingsDialog";

export function ProxyView() {
    //eslint-disable-next-line @typescript-eslint/no-unused-vars
    const [proxy, _] = useState<Proxy | null>(
        new Proxy("Proxy 1", "http://localhost:8001"),
    );
    const [proxyConfig, setProxyConfig] = useState<Config | null>(null);
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [requestsViewConfig, setRequestsViewConfig] = useState({
        hideDate: false,
        hideHostCollapsed: false,
        hideClientApplication: false,
        hideState: false,

        hideHost: false,
        hideClientIP: false,
        hideError: false,
        hideID: false,
        hideClientUser: false,
        hideClientPassword: false,
        hideMethod: false,
        hidePath: false,
        hideQuery: false,
        hideRequestHeaders: false,
        hideRequestBody: false,
        hideResponseStatus: false,
        hideResponseHeaders: false,
        hideResponseBody: false,
        hideBytesTransferred: false,
    });

    //NOTE: async/await in set state functions
    useEffect(() => {
        if (proxy === null) return;
        const load = async () => {
            await proxy.Init();
            setProxyConfig(proxy.config);
        };
        load();
    }, [proxy]);

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
                requestsViewConfig={requestsViewConfig}
                setRequestsViewConfig={setRequestsViewConfig}
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
            <IncomingView
                proxy={proxy}
                requestsViewConfig={requestsViewConfig}
            />
        </div>
    );
}
