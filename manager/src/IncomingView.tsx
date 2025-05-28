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

    console.log("22", requests);
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
            <div className="mt-2 w-full overflow-y-auto h-[80%]">
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
                    />
                ))}
            </div>
        </div>
    );
}
