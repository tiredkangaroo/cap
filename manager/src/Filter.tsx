import { Fragment, useEffect, useRef, useState } from "react";
import { Proxy } from "./api/api";
import { camelCaseToCapitalSpace, equalArray } from "./utils";
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
            !equalArray(Object.keys(filterRef.current), Object.keys(resolved))
        ) {
            // if the filter keys are not the same as the resolved keys, set the keys of filterRef.current to the resolved keys
            const newFilter: FilterType = {};
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
        <div className="flex flex-row gap-10 items-center text-black dark:text-white"></div>
    );
}

function StrValueFilter(props: {
    verboseKey: string;
    key: string;
    filter: Record<string, string | undefined>;
    setFilter: React.Dispatch<React.SetStateAction<FilterType>>;
}) {
    const uniqueValues = filterUniqueValues[props.key]; // get the unique values and counts for the current filter key
    if (!uniqueValues) {
        return <Fragment key={props.key}></Fragment>;
    }
    return (
        <div className="flex flex-row gap-1 items-center" key={props.key}>
            <Select
                key={`${props.filter}-${props.filter[props.key]}-${props.key}`}
                value={props.filter[props.key] || undefined}
                onValueChange={(v) => {
                    props.setFilter((prev) => ({
                        ...prev,
                        [props.key]: v,
                    }));
                }}
            >
                <SelectTrigger className="border-1 border-black dark:border-gray-200 min-w-[150px] bg-gray-200 dark:bg-gray-500 hover:dark:bg-gray-600 text-black dark:text-white">
                    <SelectValue placeholder={props.verboseKey} />
                </SelectTrigger>
                <SelectContent className="border-1 border-black dark:border-white bg-white dark:bg-gray-800 text-black dark:text-white">
                    <SelectGroup>
                        <SelectLabel>{props.verboseKey}</SelectLabel>
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
            {props.filter[key] !== undefined && props.filter[key] !== "" ? (
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
