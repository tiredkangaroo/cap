import { Fragment, useEffect, useRef, useState } from "react";
import { IoMdClose } from "react-icons/io";

import { Request, RequestContentProps, RequestsViewConfig } from "./types";
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
import { camelCaseToCapitalSpace, equalArray } from "./utils";
import { IoSettingsSharp } from "react-icons/io5";
import { FaSnowflake } from "react-icons/fa6";

export function IncomingView(props: {
    proxy: Proxy;
    requestsViewConfig: RequestsViewConfig;
    setSettingsDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
    requestDialogContentProps: RequestContentProps | undefined;
    setRequestDialogContentProps: React.Dispatch<
        React.SetStateAction<RequestContentProps | undefined>
    >;
}) {
    const [localRequests, setLocalRequests] = useState<Array<Request>>([]);
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
            setLocalRequests(newObj);
        });
    }, [props.proxy]);

    const dbCurrentlyShownRequests = useRef<Array<Request>>([]);
    const dbTotalPages = useRef<number>(0);
    const dbTotalCount = useRef<number>(0);

    useEffect(() => {
        const h = async () => {
            try {
                const [dbReqs, dbCount] =
                    await props.proxy.getRequestsWithFilter(
                        filter,
                        0,
                        (pageNumber + 1) * resultsPerPage,
                    );
                dbCurrentlyShownRequests.current = dbReqs;
                dbTotalCount.current = dbCount;
                dbTotalPages.current = Math.ceil(dbCount / resultsPerPage);
                // fix ts
                reloadCurrentlyShownRequests(
                    dbCurrentlyShownRequests.current,
                    dbTotalPages.current,
                    dbTotalCount.current,
                    setCurrentlyShownRequests,
                    freeze,
                    resultsPerPage,
                    localRequests,
                    filter,
                    totalPages,
                    totalResults,
                );
            } catch (error) {
                console.error(
                    "error getting db currently shown requests:",
                    error,
                );
            }
        };
        h();
    }, [pageNumber, resultsPerPage, filter, props.proxy]);

    useEffect(() => {
        const h = async () => {
            await reloadCurrentlyShownRequests(
                dbCurrentlyShownRequests.current,
                dbTotalPages.current,
                dbTotalCount.current,

                setCurrentlyShownRequests,
                freeze,
                resultsPerPage,
                localRequests,
                filter,
                totalPages,
                totalResults,
            );
        };
        h();
    }, [
        localRequests,
        pageNumber,
        freeze,
        resultsPerPage,
        props.proxy,
        props.proxy.loaded,
        filter,
    ]);

    useEffect(() => {
        props.proxy.requests = localRequests;
    }, [props.proxy, localRequests]);

    const [currentRequestCollapsibleOpen, setCurrentRequestCollapsibleOpen] =
        useState<string | undefined>(undefined);

    useEffect(() => {
        // this can happen on filter adj or results per page change
        if (pageNumber > totalPages.current - 1) {
            setPageNumber(0);
        }
    }, [pageNumber, totalPages]);

    const requestDialogContentPropsRef = useRef(
        props.requestDialogContentProps,
    );
    useEffect(() => {
        requestDialogContentPropsRef.current = props.requestDialogContentProps;
    }, [props.requestDialogContentProps]);

    return (
        <div className="w-full h-full flex flex-col">
            <div className="w-full h-full flex flex-col space-y-4 bg-gray-100 dark:bg-gray-950 text-white p-4">
                {/* Filter Section */}
                <div className="bg-gray-200 dark:bg-gray-700 rounded-xl text-black border-1 border-black p-3 flex items-center justify-between">
                    <FilterSelects
                        proxy={props.proxy}
                        requests={localRequests}
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

                <div className="flex flex-wrap items-center justify-between gap-4 px-4 py-3 bg-gray-700 text-white text-sm font-medium rounded-md">
                    <span className="w-6 text-center"></span>

                    {!props.requestsViewConfig.hideDate && (
                        <span className="flex-1 text-center">Date</span>
                    )}

                    {!props.requestsViewConfig.hideHostCollapsed && (
                        <span className="flex-1 text-center">Host</span>
                    )}

                    {!props.requestsViewConfig.hideClientApplication && (
                        <span className="flex-1 text-center mr-3">
                            Client App
                        </span>
                    )}

                    {!props.requestsViewConfig.hideState && (
                        <span className="flex-1 text-center mr-2">State</span>
                    )}
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
                                        if (
                                            // not sure what else would trigger setRequest but not the request in the dialog but the second clause will be there for my sanity
                                            requestDialogContentPropsRef.current !==
                                                undefined &&
                                            requestDialogContentPropsRef.current
                                                .request.id === req.id
                                        ) {
                                            console.log("hel");
                                            props.setRequestDialogContentProps({
                                                ...requestDialogContentPropsRef.current,
                                                request: req,
                                            });
                                        }
                                        const localIDX =
                                            localRequests.findIndex(
                                                (r) => r.id === req.id,
                                            );
                                        if (localIDX !== -1) {
                                            const newRequests = [
                                                ...localRequests,
                                            ];
                                            newRequests[localIDX] = req;
                                            setLocalRequests(newRequests);
                                            return;
                                        }
                                        const dbIdx =
                                            dbCurrentlyShownRequests.current.findIndex(
                                                (r) => r.id === req.id,
                                            );
                                        if (dbIdx !== -1) {
                                            const newRequests = [
                                                ...dbCurrentlyShownRequests.current,
                                            ];
                                            newRequests[dbIdx] = req;
                                            dbCurrentlyShownRequests.current =
                                                newRequests;
                                        }
                                        reloadCurrentlyShownRequests(
                                            dbCurrentlyShownRequests.current,
                                            dbTotalPages.current,
                                            dbTotalCount.current,

                                            setCurrentlyShownRequests,
                                            freeze,
                                            resultsPerPage,
                                            localRequests,
                                            filter,
                                            totalPages,
                                            totalResults,
                                        );
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
                        <SelectContent className="text-black dark:text-white">
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
    dbCurrentlyShownRequests: Array<Request>,
    dbTotalPages: number,
    dbTotalCount: number,

    setCurrentlyShownRequests: React.Dispatch<
        React.SetStateAction<Array<Request>>
    >,
    freeze: boolean,
    resultsPerPage: number,
    requests: Array<Request>,
    filter: Record<string, string | undefined>,
    totalPages: React.RefObject<number>,
    totalResults: React.RefObject<number>,
) {
    if (freeze) {
        // donut reload if frozen 🍩
        return;
    }
    const [cR, tP, tC] = await getCurrentlyShownRequests(
        dbCurrentlyShownRequests,
        dbTotalPages,
        dbTotalCount,

        resultsPerPage,
        requests,
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
            key={`${props.currentlyShownRequests.length}-${props.pageNumber}`}
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
    const [filterUniqueValues, setFilterUniqueValues] = useState<
        Record<string, Array<string>>
    >({});

    const filterRef = useRef(props.filter);
    useEffect(() => {
        filterRef.current = props.filter;
    }, [props.filter]);
    async function loadUniqueValues() {
        const remoteFilterCounts = await props.proxy.getFilterCounts();
        const resolved = resolveWithLocalUniqueValues(
            props.requests,
            remoteFilterCounts,
        );
        setFilterUniqueValues(resolved);
        if (
            !equalArray(Object.keys(filterRef.current), Object.keys(resolved))
        ) {
            // if the filter keys are not the same as the resolved keys, reset the filter
            const newFilter: Record<string, string | undefined> = {};
            Object.keys(resolved).forEach((key) => {
                newFilter[key] = filterRef.current[key];
            });
            props.setFilter(newFilter);
        }
        setTimeout(() => {
            loadUniqueValues();
        }, 1000);
    }
    useEffect(() => {
        loadUniqueValues();
    }, [props.filter]);

    return (
        <div className="flex flex-row gap-10 items-center text-black dark:text-white">
            <p>Query</p>
            {Object.entries(props.filter).map(([key, _]) => {
                const verboseKey = camelCaseToCapitalSpace(key);
                const uniqueValues = filterUniqueValues[key]; // get the unique values and counts for the current filter key
                if (!uniqueValues) {
                    return <Fragment key={key}></Fragment>;
                }

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
                                                {key}
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

function resolveWithLocalUniqueValues(
    requests: Array<Request>,
    filterCounts: Record<string, Array<string>>,
): Record<string, Array<string>> {
    Object.keys(filterCounts).forEach((key) => {
        const uniqueValues = new Set<string>(filterCounts[key]);
        const vKey = key as "clientIP" | "host" | "clientApplication";
        requests.forEach((req) => {
            if (req[vKey]) {
                uniqueValues.add(req[vKey]!);
            }
        });
        filterCounts[key] = Array.from(uniqueValues);
    });
    return filterCounts;
}

async function getCurrentlyShownRequests(
    dbCurrentlyShownRequests: Array<Request>,
    dbTotalPages: number,
    dbTotalCount: number,

    resultsPerPage: number,

    requests: Array<Request>,
    filter: Record<string, string | undefined>,
    // currentlyShownRequests, totalPages, totalResults
): Promise<[Array<Request>, number, number]> {
    // don't unload request and response bodies that are already loaded
    const loadedRequestBodies: Record<string, string> = {};
    const loadedResponseBodies: Record<string, string> = {};
    requests.forEach((r) => {
        if (r.tempBody) {
            loadedRequestBodies[r.id] = r.tempBody;
        }
        if (r.response && r.response.tempBody) {
            loadedResponseBodies[r.id] = r.response!.tempBody;
        }
    });

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
    if (filter.clientIP) {
        localCurrentlyShownRequests = localCurrentlyShownRequests.filter(
            (req) => req.clientIP === filter.clientIP,
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
