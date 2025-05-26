import { useEffect, useState } from "react";

import { Request, RequestsViewConfig } from "./types";
import { Proxy } from "./api";

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

    return (
        <div className="flex flex-col w-full h-full">
            <h1 className="ml-2 text-2xl font-bold mb-2">Incoming Requests</h1>
            <div className="w-full overflow-y-auto h-[90%]">
                {requests.map((request, index) => (
                    <RequestView
                        key={index}
                        proxy={props.proxy}
                        request={request}
                        requestsViewConfig={props.requestsViewConfig}
                    />
                ))}
            </div>
        </div>
    );
}
