import { useEffect, useState } from "react";
import { IoMdClose } from "react-icons/io";

import { Request, RequestsViewConfig } from "./types";
import { Proxy } from "./api/api";

import { RequestView } from "./RequestView";
import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectLabel,
    SelectTrigger,
    SelectValue,
} from "./components/ui/select";
import { camelCaseToCapitalSpace } from "./utils";

export function IncomingView(props: {
    proxy: Proxy;
    requestsViewConfig: RequestsViewConfig;
}) {
    const [requests, setRequests] = useState<Array<Request>>([]);

    // filter is used to filter the requests shown in the view.
    const [filter, setFilter] = useState<Record<string, string | undefined>>({
        clientApplication: undefined,
        host: undefined,
        state: undefined,
    });

    // currentlyShownRequests is used to show requests that are currently shown in the view. It may or may not be the same as
    // requests, depending on whether the user has filtered or sorted the requests. It is possible to check deviation with
    // isShownDeviated.
    const [currentlyShownRequests, isShownDeviated] = getCurrentlyShownRequests(
        requests,
        filter,
    );

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
                <h1 className="ml-2 text-2xl font-bold mb-2">Requests</h1>
                {/* <button
                    className="bg-gray-900 pl-3 pr-3 ml-auto mr-2 w-48 min-h-8 h-max text-white"
                    onClick={() => {
                        setRequests([]);
                    }}
                >
                    Clear Requests
                </button> */}
            </div>
            <div className="flex flex-row gap-6 ml-2">
                <FilterSelects
                    requests={requests}
                    filter={filter}
                    setFilter={setFilter}
                />
                <button
                    className="bg-gray-900 pl-3 pr-3 ml-auto mr-2 w-32 min-h-8 h-max text-white"
                    onClick={() => {
                        setFilter({
                            clientApplication: undefined,
                            host: undefined,
                            state: undefined,
                        });
                    }}
                >
                    Clear Filters
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
            <div className="flex w-full h-[calc(100vh-29vh)] overflow-y-auto mb-2">
                <div className="w-full overflow-y-auto h-full">
                    {/* NOTE: ui should auto scroll to the bottom most point of where a request has opened IF it overflows its larger container */}
                    {currentlyShownRequests.map((request, index) => (
                        <RequestView
                            key={index}
                            proxy={props.proxy}
                            request={request}
                            requestsViewConfig={props.requestsViewConfig}
                            setRequest={(req: Request) => {
                                const newRequests = [...requests];
                                const idx = requests.findIndex(
                                    (v) => v.id == req.id,
                                );
                                requests[idx] = req;

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

function FilterSelects(props: {
    requests: Array<Request>;
    filter: Record<string, string | undefined>;
    setFilter: React.Dispatch<
        React.SetStateAction<Record<string, string | undefined>>
    >;
}) {
    return (
        <div className="flex flex-row gap-10">
            {Object.entries(props.filter).map(([key, currentValue]) => {
                const verboseKey = camelCaseToCapitalSpace(key);
                const uniqueValues = [
                    ...new Set(
                        props.requests.map(
                            (item) => item[key as keyof Request],
                        ),
                    ),
                ].filter((v) => v !== undefined && v !== null);

                return (
                    <div className="flex flex-row gap-1 items-center" key={key}>
                        <Select
                            key={currentValue ?? "unset"} // 🔁 key forces remount
                            value={currentValue}
                            onValueChange={(v) => {
                                props.setFilter((prev) => ({
                                    ...prev,
                                    [key]: v,
                                }));
                            }}
                        >
                            <SelectTrigger className="border-black min-w-[150px]">
                                <SelectValue placeholder={verboseKey} />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectGroup>
                                    <SelectLabel>{verboseKey}</SelectLabel>
                                    {uniqueValues.map((value, index) => (
                                        <SelectItem
                                            key={index}
                                            value={String(value)}
                                        >
                                            {String(value)}
                                        </SelectItem>
                                    ))}
                                </SelectGroup>
                            </SelectContent>
                        </Select>
                        {currentValue !== undefined && currentValue !== "" ? (
                            <button
                                onClick={() => {
                                    props.setFilter((prev) => ({
                                        ...prev,
                                        [key]: undefined,
                                    }));
                                }}
                            >
                                <IoMdClose />
                            </button>
                        ) : (
                            <></>
                        )}
                    </div>
                );
            })}
        </div>
    );
}
function getCurrentlyShownRequests(
    requests: Array<Request>,
    filter: Record<string, string | undefined>,
): [Array<Request>, boolean] {
    let currentlyShownRequests = requests;

    if (filter.clientApplication) {
        currentlyShownRequests = currentlyShownRequests.filter(
            (req) => req.clientApplication === filter.clientApplication,
        );
    }

    if (filter.host) {
        currentlyShownRequests = currentlyShownRequests.filter(
            (req) => req.host === filter.host,
        );
    }

    if (filter.state) {
        currentlyShownRequests = currentlyShownRequests.filter(
            (req) => req.state === filter.state,
        );
    }

    const isDeviated =
        currentlyShownRequests.length !== requests.length ||
        JSON.stringify(currentlyShownRequests) !== JSON.stringify(requests);

    return [currentlyShownRequests, isDeviated];
}
