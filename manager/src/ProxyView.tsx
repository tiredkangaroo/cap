import { useEffect, useRef, useState } from "react";

import { Proxy } from "./api/api";
import { IncomingView } from "./IncomingView";

import { Config, RequestContentProps, RequestsViewConfig } from "./types";
import { SettingsDialog } from "./settings/SettingsDialog";
import { RequestDialogContentPropsContext } from "./context/context";
import { RequestDialog } from "./RequestDialog";

export function ProxyView() {
    const [proxy, setProxy] = useState<Proxy | null>(
        new Proxy("Proxy 1", "http://localhost:8001"),
    );
    const [proxyConfig, setProxyConfig] = useState<Config | null>(null);
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [requestDialogContentProps, setRequestDialogContentProps] = useState<
        RequestContentProps | undefined
    >(undefined);
    const newControlURLRef = useRef<HTMLInputElement | null>(null);
    const [requestsViewConfig, setRequestsViewConfig] =
        // JSON.parse(null) will return null, so its ok to use a non-null assertion here since behavior will be the same
        useState<RequestsViewConfig>(
            JSON.parse(localStorage.getItem("requestsViewConfig")!) || {
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
            },
        );

    useEffect(() => {
        // save requestsViewConfig to localStorage whenever it changes
        localStorage.setItem(
            "requestsViewConfig",
            JSON.stringify(requestsViewConfig),
        );
    }, [requestsViewConfig]);

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
                <div className="flex flex-row w-full justify-center items-center">
                    <input
                        ref={newControlURLRef}
                        className="mt-4 border-1 border-black p-3 min-w-12 w-[30%]"
                        autoFocus={true}
                        placeholder="Control URL"
                    ></input>
                    {/* capture erros with new proxy (connection issue etc) */}
                    <button
                        className="ml-2 mt-auto mb-2 border-1 border-white bg-black text-white px-3 py-1 rounded-sm"
                        onClick={() => {
                            setProxy(
                                new Proxy(
                                    "Proxy 1",
                                    newControlURLRef.current!.value!,
                                ),
                            );
                        }}
                    >
                        Connect
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="flex flex-col w-full h-full bg-gray-100 dark:bg-gray-950">
            <RequestDialogContentPropsContext
                value={[
                    requestDialogContentProps,
                    setRequestDialogContentProps,
                ]}
            >
                <SettingsDialog
                    proxy={proxy}
                    open={settingsOpen}
                    setOpen={setSettingsOpen}
                    requestsViewConfig={requestsViewConfig}
                    setRequestsViewConfig={setRequestsViewConfig}
                />
                <RequestDialog />
                <IncomingView
                    proxy={proxy}
                    setSettingsDialogOpen={setSettingsOpen}
                    requestsViewConfig={requestsViewConfig}
                    requestDialogContentProps={requestDialogContentProps}
                    setRequestDialogContentProps={setRequestDialogContentProps}
                />
            </RequestDialogContentPropsContext>
        </div>
    );
}
