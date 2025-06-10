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
    const [currentlyShownRequests, _] = getCurrentlyShownRequests(
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
        <div className="w-full h-full flex flex-col">
            {/* Header */}
            <div className="h-10 flex flex-row w-full">
                <h1 className="ml-2 text-2xl font-bold mb-2">Requests</h1>
            </div>

            {/* Filters */}
            <div className="h-8 flex flex-row gap-6 ml-2">
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

            {/* Table Headers */}
            <div className="h-8 mt-2 flex flex-row w-full text-center bg-gray-700 pt-1 text-white">
                {!props.requestsViewConfig.hideDate && (
                    <p className="flex-1">Date</p>
                )}
                {!props.requestsViewConfig.hideHostCollapsed && (
                    <p className="flex-1">Host</p>
                )}
                {!props.requestsViewConfig.hideClientApplication && (
                    <p className="flex-1">Client App</p>
                )}
                {!props.requestsViewConfig.hideState && (
                    <p className="flex-1">State</p>
                )}
            </div>

            <div className="h-[calc(100vh-13%-123px)] flex flex-col">
                {/* Scrollable Request List */}
                <div className="overflow-y-auto h-max">
                    {currentlyShownRequests.map((request, index) => (
                        <RequestView
                            key={index}
                            proxy={props.proxy}
                            request={request}
                            requestsViewConfig={props.requestsViewConfig}
                            setRequest={(req: Request) => {
                                const newRequests = [...requests];
                                const idx = requests.findIndex(
                                    (v) => v.id === req.id,
                                );
                                requests[idx] = req;
                                setRequests(newRequests);
                            }}
                            imOpen={(s) => {
                                if (currentCollapsibleSetOpen !== null) {
                                    currentCollapsibleSetOpen(false);
                                }
                                currentCollapsibleSetOpen = s;
                            }}
                        />
                    ))}
                </div>

                {/* Bottom Section - Always visible */}
                <div className="fixed mt-auto bottom-0 w-full h-[calc(5%)] bg-gray-900 flex items-center justify-center shrink-0">
                    good morning
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
            {Object.entries(props.filter).map(([key, _]) => {
                const verboseKey = camelCaseToCapitalSpace(key);

                const countFilter = { ...props.filter, [key]: undefined };
                const [countSource, __] = getCurrentlyShownRequests(
                    props.requests,
                    countFilter,
                );

                const uniqueValues = [
                    ...new Set(
                        props.requests.map(
                            (item) => item[key as keyof Request],
                        ),
                    ),
                ].filter((v) => v !== undefined && v !== null);

                // Count occurrences in countSource
                const counts: Record<string, number> = countSource.reduce(
                    (acc, item) => {
                        const value = item[key as keyof Request];
                        if (value !== undefined && value !== null) {
                            const stringValue = String(value);
                            acc[stringValue] = (acc[stringValue] || 0) + 1;
                        }
                        return acc;
                    },
                    {} as Record<string, number>,
                );

                // Build a sorted array of [key, count] tuples based on uniqueValues
                const sorted = uniqueValues
                    .map((value) => {
                        const stringValue = String(value);
                        return [stringValue, counts[stringValue] || 0] as [
                            string,
                            number,
                        ];
                    })
                    .sort((a, b) => b[1] - a[1]);

                // Convert back to a Record<string, number>
                const result: Record<string, number> =
                    Object.fromEntries(sorted);

                return (
                    <div className="flex flex-row gap-1 items-center" key={key}>
                        <Select
                            key={`${props.filter}-${props.filter[key]}-${key}`}
                            value={props.filter[key] || undefined}
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
                                    {Object.keys(result).map((key, index) => (
                                        <SelectItem key={index} value={key}>
                                            {key} ({result[key]})
                                        </SelectItem>
                                    ))}
                                </SelectGroup>
                            </SelectContent>
                        </Select>
                        {props.filter[key] !== undefined &&
                        props.filter[key] !== "" ? (
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
