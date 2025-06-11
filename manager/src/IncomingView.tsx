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
    const [pageNumber, setPageNumber] = useState<number>(0);
    const [resultsPerPage, setResultsPerPage] = useState<number>(15);

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

    // FIXME: the currently open one will change if a new request comes in
    let currentCollapsibleSetOpen: React.Dispatch<
        React.SetStateAction<boolean>
    > | null = null;

    // this can happen on filter adj or results per page change
    if (
        pageNumber > Math.ceil(currentlyShownRequests.length / resultsPerPage)
    ) {
        setPageNumber(0);
    }

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
            <div className="ml-2 min-h-fit py-1 text-gray-500">
                {currentlyShownRequests.length} results
            </div>

            {/* Table Headers */}
            <div className="h-8 flex flex-row w-full text-center bg-gray-700 pt-1 text-white">
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
                    {currentlyShownRequests
                        .slice(
                            pageNumber * resultsPerPage,
                            pageNumber * resultsPerPage + resultsPerPage,
                        )
                        .map((request) => {
                            const index = requests.findIndex(
                                (v) => v.id === request.id,
                            );
                            return (
                                <RequestView
                                    key={index}
                                    proxy={props.proxy}
                                    request={request}
                                    requestsViewConfig={
                                        props.requestsViewConfig
                                    }
                                    setRequest={(req: Request) => {
                                        const newRequests = [...requests];
                                        const idx = requests.findIndex(
                                            (v) => v.id === req.id,
                                        );
                                        requests[idx] = req;
                                        setRequests(newRequests);
                                    }}
                                    imOpen={(s) => {
                                        if (
                                            currentCollapsibleSetOpen !== null
                                        ) {
                                            currentCollapsibleSetOpen(false);
                                        }
                                        currentCollapsibleSetOpen = s;
                                    }}
                                />
                            );
                        })}
                </div>

                <div className="mt-10"></div>
                {/* Bottom Section - Always visible */}
                <div className="fixed mt-auto bottom-0 w-full h-[calc(5%)] bg-gray-900 flex flex-row items-center justify-between shrink-0">
                    <Pagination
                        requests={requests}
                        currentlyShownRequests={currentlyShownRequests}
                        pageNumber={pageNumber}
                        setPageNumber={setPageNumber}
                        resultsPerPage={resultsPerPage}
                    />
                    <div className="flex flex-row text-white text-sm items-center mr-2">
                        Results per page:
                        <Select
                            value={resultsPerPage.toString() ?? "15"}
                            onValueChange={(v) => {
                                setResultsPerPage(parseInt(v) ?? 15);
                            }}
                        >
                            <SelectTrigger className="ml-4 border-white text-white min-w-[150px]">
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent className="text-black">
                                {["10", "15", "25", "50", "75", "100"].map(
                                    (v) => (
                                        <SelectItem
                                            className="text-black"
                                            value={v}
                                        >
                                            {v}
                                        </SelectItem>
                                    ),
                                )}
                            </SelectContent>
                        </Select>
                    </div>
                </div>
            </div>
        </div>
    );
}

function Pagination(props: {
    requests: Array<Request>;
    currentlyShownRequests: Array<Request>;
    pageNumber: number;
    setPageNumber: React.Dispatch<React.SetStateAction<number>>;
    resultsPerPage: number;
}) {
    const totalPages = Math.ceil(
        props.currentlyShownRequests.length / props.resultsPerPage,
    );

    const current = props.pageNumber;
    const pagesToShow: number[] = [];

    if (totalPages <= 5) {
        // Show all pages if <= 5
        for (let i = 0; i < totalPages; i++) {
            pagesToShow.push(i);
        }
    } else {
        if (current <= 2) {
            // Near the start
            pagesToShow.push(0, 1, 2, 3, totalPages - 1);
        } else if (current >= totalPages - 3) {
            // Near the end
            pagesToShow.push(
                0,
                totalPages - 4,
                totalPages - 3,
                totalPages - 2,
                totalPages - 1,
            );
        } else {
            // Middle
            pagesToShow.push(
                0,
                current - 1,
                current,
                current + 1,
                totalPages - 1,
            );
        }
    }

    // ensure uniqueness and sort
    const uniquePages = Array.from(new Set(pagesToShow)).sort((a, b) => a - b);

    function PaginationItem({ v }: { v: number }) {
        return (
            <a
                key={v}
                className="cursor-pointer px-2"
                style={{
                    textDecoration: v === current ? "underline" : "none",
                }}
                onClick={() => props.setPageNumber(v)}
            >
                {v + 1}
            </a>
        );
    }

    return (
        <div
            className="flex flex-row gap-2 ml-4 text-white"
            key={`${props.requests.length}-${props.pageNumber}`}
        >
            {uniquePages.map((v, i) => {
                let ellipses = false;
                if (i > 0) {
                    const prev = uniquePages[i - 1];
                    ellipses = v - prev > 1;
                }

                return (
                    <>
                        {ellipses ? <span>..</span> : <></>}
                        <PaginationItem key={v} v={v} />
                    </>
                );
            })}
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
    let currentlyShownRequests = requests.reverse();

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
