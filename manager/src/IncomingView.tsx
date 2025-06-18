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

export function IncomingView(props: {
    proxy: Proxy;
    requestsViewConfig: RequestsViewConfig;
}) {
    const [requests, setRequests] = useState<Array<Request>>([]);
    const [currentlyShownRequests, setCurrentlyShownRequests] = useState<
        Array<Request>
    >([]);

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
            const [cR, tP, tC] = await getCurrentlyShownRequests(
                pageNumber,
                resultsPerPage,
                requests,
                props.proxy,
                filter,
            );
            setCurrentlyShownRequests(cR);
            totalPages.current = tP;
            totalResults.current = tC;
        };
        h();
    }, [
        pageNumber,
        resultsPerPage,
        props.proxy,
        props.proxy.loaded,
        requests,
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
            {/* Header */}
            <div className="h-10 flex flex-row w-full">
                <h1 className="ml-2 text-2xl font-bold mb-2">Requests</h1>
            </div>

            {/* Filters */}
            <div className="h-8 flex flex-row gap-6 ml-2">
                <FilterSelects
                    proxy={props.proxy}
                    requests={requests}
                    currentlyShownRequests={currentlyShownRequests}
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
                {totalResults.current} results
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
                            (pageNumber + 1) * resultsPerPage,
                        )
                        .map((request, idx) => {
                            return (
                                <RequestView
                                    key={idx}
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
                                    open={
                                        request.id ===
                                        currentRequestCollapsibleOpen
                                    }
                                    setOpen={(o: boolean) => {
                                        if (o) {
                                            setCurrentRequestCollapsibleOpen(
                                                request.id,
                                            );
                                        } else {
                                            setCurrentRequestCollapsibleOpen(
                                                undefined,
                                            );
                                        }
                                    }}
                                />
                            );
                        })}
                </div>

                <div className="mt-8"></div>
                {/* Bottom Section - Always visible */}
                <div className="fixed mt-auto bottom-0 w-full h-[calc(5%)] bg-gray-900 flex flex-row items-center justify-between shrink-0">
                    <Pagination
                        requests={requests}
                        currentlyShownRequests={currentlyShownRequests}
                        pageNumber={pageNumber}
                        setPageNumber={setPageNumber}
                        resultsPerPage={resultsPerPage}
                        totalPages={totalPages.current}
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
                                            key={v}
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
        <div className="flex flex-row gap-10">
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
                            <SelectTrigger className="border-black min-w-[150px]">
                                <SelectValue placeholder={verboseKey} />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectGroup>
                                    <SelectLabel>{verboseKey}</SelectLabel>
                                    {uniqueValues.map((key, index) => (
                                        <SelectItem key={index} value={key}>
                                            {key} ({uniqueValuesAndCounts[key]})
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
            const value = request[filterKey];
            if (value !== undefined) {
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
            [...dbCurrentlyShownRequests, ...localCurrentlyShownRequests].map(
                (obj) => [obj.id, obj],
            ),
        ).values(),
    ];

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
