import { useEffect, useState } from "react";

import { Request, RequestsViewConfig } from "./types";
import { Proxy } from "./api/api";

import { RequestView } from "./RequestView";

export function IncomingView(props: {
    proxy: Proxy;
    requestsViewConfig: RequestsViewConfig;
}) {
    const [requests, setRequests] = useState<Array<Request>>([]);

    useEffect(() => {
        props.proxy.manageRequests(() => {
            console.log("update cb", props.proxy.requests);
            const newObj = Object.assign([], props.proxy.requests);
            setRequests(newObj);
        });
    }, [props.proxy]);

    useEffect(() => {
        props.proxy.requests = requests;
    }, [requests]);

    let currentCollapsibleSetOpen: React.Dispatch<
        React.SetStateAction<boolean>
    > | null = null;
    return (
        <div className="flex flex-col w-full h-full">
            <div className="flex flex-row w-full">
                <h1 className="ml-2 text-2xl font-bold mb-2">
                    Incoming Requests
                </h1>
                <button
                    className="bg-gray-700 pl-3 pr-3 ml-auto mr-2 w-32 text-white"
                    onClick={() => {
                        setRequests([]);
                    }}
                >
                    Clear
                </button>
            </div>
            <div className="mt-2 flex flex-row w-full space-y-2 text-center bg-gray-700 pt-1 text-white">
                {!props.requestsViewConfig.hideDate ? (
                    <p className="flex-1">Date</p>
                ) : (
                    <></>
                )}
                {!props.requestsViewConfig.hideHostCollapsed ? (
                    <p className="flex-1">Host</p>
                ) : (
                    <></>
                )}
                {!props.requestsViewConfig.hideClientApplication ? (
                    <p className="flex-1">Client Application</p>
                ) : (
                    <></>
                )}
                {!props.requestsViewConfig.hideState ? (
                    <p className="flex-1">State</p>
                ) : (
                    <></>
                )}
            </div>
            <div className="flex w-full h-[calc(100vh-25vh)] overflow-y-auto mb-2">
                <div className="w-full overflow-y-auto h-full">
                    {/* NOTE: ui should auto scroll to the bottom most point of where a request has opened IF it overflows its larger container */}
                    {requests.map((request, index) => (
                        <RequestView
                            key={index}
                            proxy={props.proxy}
                            request={request}
                            requestsViewConfig={props.requestsViewConfig}
                            setRequest={(req: Request) => {
                                const newRequests = [...requests];
                                newRequests[index] = req;
                                setRequests(newRequests);
                            }}
                            imOpen={(s) => {
                                // im open im open no zawg ur not
                                if (currentCollapsibleSetOpen !== null) {
                                    currentCollapsibleSetOpen(false);
                                }
                                currentCollapsibleSetOpen = s;
                            }}
                        />
                    ))}
                </div>
            </div>
        </div>
    );
}
