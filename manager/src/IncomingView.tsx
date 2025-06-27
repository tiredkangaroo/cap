import { Fragment, useEffect, useRef, useState } from "react";
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
import { IoSettingsSharp } from "react-icons/io5";
import { FaSnowflake } from "react-icons/fa6";

export function IncomingView(props: {
    proxy: Proxy;
    requestsViewConfig: RequestsViewConfig;
    setSettingsDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
}) {
    const [requests, setRequests] = useState<Array<Request>>([]);
    const [currentlyShownRequests, setCurrentlyShownRequests] = useState<
        Array<Request>
    >([]);

    const [freeze, setFreeze] = useState<boolean>(false);

    // filter is used to filter the requests shown in the view.
    const [filter, setFilter] = useState<Record<string, string | undefined>>(
        {},
    );

    const [pageNumber, setPageNumber] = useState<number>(0);
    const [resultsPerPage, setResultsPerPage] = useState<number>(
        parseInt(localStorage.getItem("resultsPerPage") || "15"),
    );
    // NOTE: possible remove totalPages for totalResults only and using ceil division
    const totalPages = useRef<number>(0);
    const totalResults = useRef<number>(0);
    useEffect(() => {
        localStorage.setItem("resultsPerPage", resultsPerPage.toString());
    }, [resultsPerPage]);

    useEffect(() => {
        props.proxy.manageRequests(() => {
            const newObj = Object.assign([], props.proxy.requests);
            setRequests(newObj);
        });
    }, [props.proxy]);

    useEffect(() => {
        const h = async () => {
            await reloadCurrentlyShownRequests(
                setCurrentlyShownRequests,
                freeze,
                pageNumber,
                resultsPerPage,
                requests,
                props.proxy,
                filter,
                totalPages,
                totalResults,
            );
        };
        h();
    }, [
        requests,
        pageNumber,
        freeze,
        resultsPerPage,
        props.proxy,
        props.proxy.loaded,
        filter,
    ]);

    useEffect(() => {
        props.proxy.requests = requests;
    }, [props.proxy, requests]);

    const [currentRequestCollapsibleOpen, setCurrentRequestCollapsibleOpen] =
        useState<string | undefined>(undefined);

    useEffect(() => {
        // this can happen on filter adj or results per page change
        if (pageNumber > totalPages.current - 1) {
            setPageNumber(0);
        }
    }, [pageNumber, totalPages]);

    return (
        <div className="w-full h-full flex flex-col">
            <div className="w-full h-full flex flex-col space-y-4 bg-gray-100 dark:bg-gray-950 text-white p-4">
                {/* Filter Section */}
                <div className="bg-gray-200 dark:bg-gray-700 rounded-xl text-black border-1 border-black p-3 flex items-center justify-between">
                    <FilterSelects
                        proxy={props.proxy}
                        requests={requests}
                        currentlyShownRequests={currentlyShownRequests}
                        filter={filter}
                        setFilter={setFilter}
                    />
                    <button
                        className="text-sm text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-600 font-medium"
                        onClick={() => {
                            Object.keys(filter).forEach((key) => {
                                filter[key] = undefined;
                            });
                            setFilter({});
                        }}
                    >
                        Clear Filters
                    </button>
                </div>

                <div className="text-gray-600 dark:text-gray-200 text-sm ml-1">
                    {totalResults.current} results
                </div>

                <div className="bg-gray-700 text-white text-sm font-medium rounded-md py-2 px-4 grid grid-cols-4 text-center sticky top-0 z-10">
                    {!props.requestsViewConfig.hideDate && <span>Date</span>}
                    {!props.requestsViewConfig.hideHostCollapsed && (
                        <span>Host</span>
                    )}
                    {!props.requestsViewConfig.hideClientApplication && (
                        <span>Client App</span>
                    )}
                    {!props.requestsViewConfig.hideState && <span>State</span>}
                </div>

                <div className="flex-1 overflow-y-auto pr-2 max-h-[calc(100vh-300px)]">
                    {currentlyShownRequests
                        .slice(
                            pageNumber * resultsPerPage,
                            (pageNumber + 1) * resultsPerPage,
                        )
                        .map((request, idx) => (
                            <div key={idx} className="bg-white rounded-lg">
                                <RequestView
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
                                        if (idx === -1) {
                                            newRequests.push(req);
                                        } else {
                                            newRequests[idx] = req;
                                        }
                                        setRequests(newRequests);
                                    }}
                                    open={
                                        request.id ===
                                        currentRequestCollapsibleOpen
                                    }
                                    setOpen={(o: boolean) => {
                                        setCurrentRequestCollapsibleOpen(
                                            o ? request.id : undefined,
                                        );
                                    }}
                                />
                            </div>
                        ))}
                </div>
            </div>

            <div className="bg-gray-700 dark:bg-gray-900 sticky bottom-0 shadow-inner border-t mt-4 px-4 py-2 flex items-center justify-between">
                <Pagination
                    requests={requests}
                    currentlyShownRequests={currentlyShownRequests}
                    pageNumber={pageNumber}
                    setPageNumber={setPageNumber}
                    resultsPerPage={resultsPerPage}
                    totalPages={totalPages.current}
                />

                <div className="flex items-center gap-4">
                    <span className="text-sm text-white">
                        Results per page:
                    </span>
                    <Select
                        value={resultsPerPage.toString() ?? "15"}
                        onValueChange={(v) => {
                            setResultsPerPage(parseInt(v) ?? 15);
                        }}
                    >
                        <SelectTrigger className="min-w-[120px] text-white border-gray-300 shadow-sm">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent className="text-black">
                            {["10", "15", "25", "50", "75", "100"].map((v) => (
                                <SelectItem key={v} value={v}>
                                    {v}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <IncomingFreezeButton
                        freeze={freeze}
                        setFreeze={setFreeze}
                    />
                    <button
                        className="p-2 bg-gray-100 hover:bg-gray-200 rounded-md text-gray-700"
                        onClick={() => props.setSettingsDialogOpen(true)}
                    >
                        <IoSettingsSharp size={18} />
                    </button>
                </div>
            </div>
        </div>
    );
}

async function reloadCurrentlyShownRequests(
    setCurrentlyShownRequests: React.Dispatch<
        React.SetStateAction<Array<Request>>
    >,
    freeze: boolean,
    pageNumber: number,
    resultsPerPage: number,
    requests: Array<Request>,
    proxy: Proxy,
    filter: Record<string, string | undefined>,
    totalPages: React.RefObject<number>,
    totalResults: React.RefObject<number>,
) {
    if (freeze) {
        // donut reload if frozen 🍩
        return;
    }
    const [cR, tP, tC] = await getCurrentlyShownRequests(
        pageNumber,
        resultsPerPage,
        requests,
        proxy,
        filter,
    );
    setCurrentlyShownRequests(cR);
    totalPages.current = tP;
    totalResults.current = tC;
}

function IncomingFreezeButton(props: {
    freeze: boolean;
    setFreeze: React.Dispatch<React.SetStateAction<boolean>>;
}) {
    return (
        <button
            className="p-2 rounded-md border-1 border-blue-300 hover:border-blue-400"
            style={{
                backgroundColor: props.freeze
                    ? "var(--color-blue-300)"
                    : "transparent",
                color: props.freeze ? "black" : "white",
            }}
            onClick={() => {
                props.setFreeze(!props.freeze);
            }}
        >
            <FaSnowflake />
        </button>
    );
}

function Pagination(props: {
    requests: Array<Request>;
    currentlyShownRequests: Array<Request>;
    pageNumber: number;
    setPageNumber: React.Dispatch<React.SetStateAction<number>>;
    resultsPerPage: number;
    totalPages: number;
}) {
    const totalPages = props.totalPages;
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
                    <Fragment key={v}>
                        {ellipses ? <span>..</span> : <></>}
                        <PaginationItem key={v} v={v} />
                    </Fragment>
                );
            })}
        </div>
    );
}

function FilterSelects(props: {
    proxy: Proxy;
    requests: Array<Request>;
    currentlyShownRequests: Array<Request>;
    filter: Record<string, string | undefined>;
    setFilter: React.Dispatch<
        React.SetStateAction<Record<string, string | undefined>>
    >;
}) {
    // [filterName: {uniqueValue: # time of appearance of unique value}]
    const [filterUniqueValuesCounts, setFilterUniqueValuesCounts] = useState<
        Record<string, Record<string, number>>
    >({});

    useEffect(() => {
        // NOTE: name this function something more descriptive
        const h = async () => {
            const result = await props.proxy.getFilterCounts();
            setFilterUniqueValuesCounts(
                resolveWithLocalFC(props.requests, result),
            );
            if (Object.keys(props.filter).length === 0) {
                props.setFilter(
                    Object.keys(result).reduce(
                        (acc, key) => {
                            acc[key] = undefined;
                            return acc;
                        },
                        {} as Record<string, string | undefined>,
                    ),
                );
            }
        };
        h();
    }, [props.proxy, props.requests, props.currentlyShownRequests]);

    return (
        <div className="flex flex-row gap-10 items-center text-black dark:text-white">
            <p>Query</p>
            {Object.entries(props.filter).map(([key, _]) => {
                const verboseKey = camelCaseToCapitalSpace(key);
                const uniqueValuesAndCounts = filterUniqueValuesCounts[key]; // get the unique values and counts for the current filter key
                if (!uniqueValuesAndCounts) {
                    return <Fragment key={key}></Fragment>;
                }
                const uniqueValues = Object.keys(uniqueValuesAndCounts); // unique values for the current filter key

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
                            <SelectTrigger className="border-1 border-black dark:border-gray-200 min-w-[150px] bg-gray-200 dark:bg-gray-500 hover:dark:bg-gray-600 text-black dark:text-white">
                                <SelectValue placeholder={verboseKey} />
                            </SelectTrigger>
                            <SelectContent className="border-1 border-black dark:border-white bg-white dark:bg-gray-800 text-black dark:text-white">
                                <SelectGroup>
                                    <SelectLabel>{verboseKey}</SelectLabel>
                                    {uniqueValues.map((key, index) => (
                                        <SelectItem
                                            className="hover:bg-gray-300 hover:dark:bg-gray-600"
                                            key={index}
                                            value={key}
                                        >
                                            <div className="justify-between flex flex-row w-full">
                                                {key} (
                                                {uniqueValuesAndCounts[key]})
                                            </div>
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

function resolveWithLocalFC(
    requests: Array<Request>,
    filterUniqueValuesCounts: Record<string, Record<string, number>>,
): Record<string, Record<string, number>> {
    Object.keys(filterUniqueValuesCounts).forEach((filterKey) => {
        const localUniqueValuesAndCounts: Record<string, number> = {};

        // Add local requests counts
        requests.forEach((request) => {
            const value = request[filterKey as keyof Request] as
                | string
                | number;
            if (value !== undefined && value !== null && value !== "") {
                if (localUniqueValuesAndCounts[value] === undefined) {
                    localUniqueValuesAndCounts[value] = 0;
                }
                localUniqueValuesAndCounts[value] += 1;
            }
        });

        Object.keys(localUniqueValuesAndCounts).forEach((localValue) => {
            const count = localUniqueValuesAndCounts[localValue];
            const dbCount = filterUniqueValuesCounts[filterKey][localValue];
            if (dbCount == undefined) {
                filterUniqueValuesCounts[filterKey][localValue] = count;
            } else {
                // NOTE: this can cause inaccuracy if requests isn't 100% deviated from the db
                filterUniqueValuesCounts[filterKey][localValue] = Math.max(
                    dbCount,
                    count,
                );
            }
        });
    });

    return filterUniqueValuesCounts;
}

async function getCurrentlyShownRequests(
    pageNumber: number,
    resultsPerPage: number,

    requests: Array<Request>,
    proxy: Proxy,
    filter: Record<string, string | undefined>,
    // currentlyShownRequests, totalPages, totalResults
): Promise<[Array<Request>, number, number]> {
    const loadedRequestBodies: Record<string, string> = {};
    const loadedResponseBodies: Record<string, string> = {};
    requests.forEach((r) => {
        if (r.tempBody) {
            loadedRequestBodies[r.id] = r.tempBody;
        }
        if (r.response!.tempBody) {
            loadedResponseBodies[r.id] = r.response!.tempBody;
        }
    });

    let dbCurrentlyShownRequests: Array<Request> = [];
    let dbTotalPages = 0;
    let dbTotalCount = 0;
    try {
        [dbCurrentlyShownRequests, dbTotalCount] =
            await proxy.getRequestsWithFilter(
                filter,
                0,
                (pageNumber + 1) * resultsPerPage,
            );
        dbTotalPages = Math.ceil(dbTotalCount / resultsPerPage);
    } catch (error) {
        console.error("error getting db currently shown requests:", error);
    }

    let localCurrentlyShownRequests: Array<Request> = requests;
    if (filter.clientApplication) {
        localCurrentlyShownRequests = localCurrentlyShownRequests.filter(
            (req) => req.clientApplication === filter.clientApplication,
        );
    }
    if (filter.host) {
        localCurrentlyShownRequests = localCurrentlyShownRequests.filter(
            (req) => req.host === filter.host,
        );
    }

    const currentlyShownRequests = [
        // using map elimnates id dups
        ...new Map(
            [...localCurrentlyShownRequests, ...dbCurrentlyShownRequests].map(
                (obj) => [obj.id, obj],
            ),
        ).values(),
    ];

    Object.keys(loadedRequestBodies).forEach((id) => {
        const request = currentlyShownRequests.find((r) => r.id === id);
        if (request) {
            request.tempBody = loadedRequestBodies[id];
        }
    });
    Object.keys(loadedResponseBodies).forEach((id) => {
        const request = currentlyShownRequests.find((r) => r.id === id);
        if (request) {
            request.response!.tempBody = loadedResponseBodies[id];
        }
    });

    currentlyShownRequests.sort((a, b) => {
        const dateA = new Date(a.datetime).getTime();
        const dateB = new Date(b.datetime).getTime();
        return dateB - dateA; // Sort by date descending
    });

    let totalPages = dbTotalPages;
    if (
        Math.ceil(currentlyShownRequests.length / resultsPerPage) > dbTotalPages
    ) {
        totalPages = Math.ceil(currentlyShownRequests.length / resultsPerPage);
    }

    let totalCount = dbTotalCount;
    if (currentlyShownRequests.length > dbTotalCount) {
        totalCount = currentlyShownRequests.length;
    }

    return [currentlyShownRequests, totalPages, totalCount];
}
