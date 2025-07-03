import { Fragment, useEffect, useRef } from "react";
import { Proxy } from "./api/api";
import { equalArray } from "./utils";
import { FilterType, Request } from "./types";
import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectLabel,
    SelectTrigger,
    SelectValue,
} from "./components/ui/select";
import { IoMdClose } from "react-icons/io";

export function FilterSelects(props: {
    proxy: Proxy;
    requests: Array<Request>;
    currentlyShownRequests: Array<Request>;
    filter: FilterType;
    setFilter: React.Dispatch<React.SetStateAction<FilterType>>;
}) {
    const filterRef = useRef(props.filter);
    useEffect(() => {
        filterRef.current = props.filter;
    }, [props.filter]);
    async function loadUniqueValues() {
        const newFilter = await props.proxy.getFilter();
        const resolved = resolveFilterWithLocalUniqueValues(
            props.requests,
            newFilter,
        );
        if (
            !equalArray(
                namesFromFilter(filterRef.current),
                namesFromFilter(resolved),
            )
        ) {
            // if the filter keys are not the same as the resolved keys, set the keys of filterRef.current to the resolved keys
            const newFilter: FilterType = [];
            resolved.forEach((field) => {
                newFilter.push(field);
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
            {props.filter.map((key, index) => {
                if (key.type === "string") {
                    return (
                        <StrValueFilter
                            key={index}
                            filters={props.filter}
                            index={index}
                            setFilters={props.setFilter}
                        />
                    );
                } else if (key.type === "boolean") {
                    return (
                        // <BooleanValueFilter
                        //     key={index}
                        //     verboseKey={key.verboseName}
                        //     keyName={key.name}
                        //     filter={props.filter}
                        //     setFilter={props.setFilter}
                        // />
                        <p>bool</p>
                    );
                }
                return <Fragment key={index}></Fragment>;
            })}
        </div>
    );
}

function StrValueFilter(props: {
    filters: FilterType;
    index: number;
    setFilters: React.Dispatch<React.SetStateAction<FilterType>>;
}) {
    const filter = props.filters[props.index] as {
        name: string;
        verboseName: string;
        type: "string";
        uniqueValues: string[];
        selectedValue?: string;
    };
    const uniqueValues = filter.uniqueValues as string[];
    const verboseName = filter.verboseName;
    return (
        <div className="flex flex-row gap-1 items-center" key={props.index}>
            <Select
                key={`${props.filters}-${filter}-${props.index}`}
                value={filter.selectedValue}
                onValueChange={(v) => {
                    const newFilter = [...props.filters];
                    newFilter[props.index] = {
                        ...newFilter[props.index],
                        selectedValue: v,
                    };
                    props.setFilters(newFilter);
                }}
            >
                <SelectTrigger className="border-1 border-black dark:border-gray-200 min-w-[150px] bg-gray-200 dark:bg-gray-500 hover:dark:bg-gray-600 text-black dark:text-white">
                    <SelectValue placeholder={verboseName} />
                </SelectTrigger>
                <SelectContent className="border-1 border-black dark:border-white bg-white dark:bg-gray-800 text-black dark:text-white">
                    <SelectGroup>
                        <SelectLabel>{verboseName}</SelectLabel>
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
            {/* we could js replace ts w a non-triple equals */}
            {filter.selectedValue !== undefined &&
            filter.selectedValue !== "" ? (
                <button
                    onClick={() => {
                        const newFilter = [...props.filters];
                        newFilter[props.index] = {
                            ...newFilter[props.index],
                            selectedValue: undefined,
                        };
                        props.setFilters(newFilter);
                    }}
                >
                    <IoMdClose />
                </button>
            ) : (
                <></>
            )}
        </div>
    );
}

function resolveFilterWithLocalUniqueValues(
    requests: Array<Request>,
    filter: FilterType,
): FilterType {
    filter.forEach((key) => {
        if (key.type !== "string" && key.type !== "number") {
            return;
        }
        const uniqueValues = new Set(key.uniqueValues as string[] | number[]);
        const vKey = key.name as "clientIP" | "clientApplication" | "host";
        requests.forEach((req) => {
            if (req[vKey]) {
                uniqueValues.add(req[vKey]!);
            }
        });
        key.uniqueValues = Array.from(uniqueValues);
    });
    return filter;
}

function namesFromFilter(filter: FilterType): string[] {
    return filter.map((item) => item.name);
}
