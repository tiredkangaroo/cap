import {
    Collapsible,
    CollapsibleTrigger,
    CollapsibleContent,
} from "./components/ui/collapsible";
import { Request, RequestsViewConfig } from "./types";
import { downloadBody, downloadRequest } from "./downloadRequest";
import { useState } from "react";
import { Proxy } from "./api";

const stateColors: Record<string, string> = {
    Processing: "#000",
    Canceled: "#806262",
    Done: "#62806b",
    Error: "oklch(50.5% 0.213 27.518)",
    "Approval Timeout": "#806262",
    "Waiting Approval": "#806262",
};
export function RequestView(props: {
    proxy: Proxy;
    request: Request;
    requestsViewConfig: RequestsViewConfig;
}) {
    return (
        <Collapsible className="border-b-1 border-b-black wrap-anywhere">
            <CollapsibleTrigger className="w-full bg-gray-200">
                <div className="relative flex flex-row w-full pt-4 pb-4">
                    {/* <p
                        className="flex-1"
                        style={{
                            color: props.request.secure ? "#0b5c00" : "#5c0300",
                        }}
                    >
                        {props.request.secure ? "HTTPS" : "HTTP"}
                    </p> */}
                    <p className="flex-1 ml-1 text-sm">
                        {props.request.datetime}
                    </p>
                    <p className="flex-1 ml-1 text-center">
                        {props.request.host}
                    </p>
                    <StateView
                        proxy={props.proxy}
                        id={props.request.id}
                        state={props.request.state}
                    />
                    {/* <p className="flex-1 ml-1 mr-1">{props.request.clientIP}</p> */}
                </div>
            </CollapsibleTrigger>
            <CollapsibleContent className="bg-gray-300 max-h-[50vh] overflow-y-auto">
                <button
                    className="bg-gray-600 text-white border-black border-1 ml-2 mt-2 pl-2 pr-2"
                    onClick={() => {
                        downloadRequest(props.request);
                    }}
                >
                    Download
                </button>
                <div className="ml-2 pt-2 pb-1">
                    {props.request.error != undefined &&
                    !props.requestsViewConfig.hideError ? (
                        <p>
                            <span className="text-red-700">Error</span>:{" "}
                            {props.request.error}
                        </p>
                    ) : (
                        <></>
                    )}
                    <FieldView
                        name="ID"
                        value={props.request.id}
                        hide={props.requestsViewConfig.hideID}
                    />
                    <p>
                        <b>Request: </b>
                    </p>
                    <div className="hashi w-full pt-2 pl-5 pr-5">
                        <FieldView
                            name="Client Username"
                            value={props.request.clientAuthorizationUser}
                            hide={props.requestsViewConfig.hideClientUser}
                        />
                        <ShowHideFieldView
                            name="Client Password"
                            value={props.request.clientAuthorizationPassword}
                            hiddenValue="********"
                            defaultShow={false}
                            hide={props.requestsViewConfig.hideClientPassword}
                        />
                        <FieldView
                            name="Method"
                            value={props.request.method}
                            hide={props.requestsViewConfig.hideMethod}
                        />
                        <FieldView
                            name="Path"
                            value={props.request.path}
                            hide={props.requestsViewConfig.hidePath}
                        />
                        <FieldView
                            name="Query"
                            value={props.request.query}
                            hide={props.requestsViewConfig.hideQuery}
                        />
                        <FieldView
                            name="Headers"
                            value={props.request.headers}
                            hide={props.requestsViewConfig.hideRequestHeaders}
                        />
                        {props.requestsViewConfig.hideRequestBody ? (
                            <></>
                        ) : (
                            <button
                                className="bg-gray-600 text-white border-black border-1 mt-2 pl-2 pr-2"
                                onClick={() =>
                                    downloadBody(
                                        props.request.id,
                                        props.request.body,
                                        props.request.headers![
                                            "Content-Type"
                                        ][0],
                                    )
                                }
                            >
                                Download Body
                            </button>
                        )}
                        <BodyView
                            value={props.request.body}
                            hide={props.requestsViewConfig.hideRequestBody}
                        />
                    </div>
                    <p>
                        <b>Response: </b>
                    </p>
                    <div className="mt-2 ml-5">
                        <FieldView
                            name="Status"
                            value={props.request.response?.statusCode}
                            hide={props.requestsViewConfig.hideResponseStatus}
                        />
                        <FieldView
                            name="Headers"
                            value={props.request.response?.headers}
                            hide={props.requestsViewConfig.hideResponseHeaders}
                        />
                        {props.requestsViewConfig.hideResponseBody ? (
                            <></>
                        ) : (
                            <button
                                className="bg-gray-600 text-white border-black border-1 mt-2 pl-2 pr-2"
                                onClick={() =>
                                    downloadBody(
                                        props.request.id,
                                        props.request.response!.body,
                                        props.request.response!.headers![
                                            "Content-Type"
                                        ][0],
                                    )
                                }
                            >
                                Download Body
                            </button>
                        )}
                        <BodyView
                            value={props.request.response?.body}
                            hide={props.requestsViewConfig.hideResponseBody}
                        />
                    </div>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

function ValueView(props: {
    name: string;
    value: number | string | Record<string, Array<string>> | undefined;
    italic?: boolean;
}) {
    if (props.value == undefined || props.value == "") {
        return <i>none or unavailable</i>;
    }
    if (typeof props.value === "string" || typeof props.value === "number") {
        if (props.name.startsWith("Body")) {
            if (props.italic === true) {
                return <i>{props.value}</i>;
            } else {
                return <p className="text-sm">{props.value}</p>;
            }
        } else {
            if (props.italic === true) {
                return <i>{props.value}</i>;
            } else {
                return <>{props.value}</>;
            }
        }
    }
    return (
        <>
            {Object.entries(props.value!).map((v) => (
                <div key={v[0]} className="text-sm">
                    {v[1].map((x, i) => (
                        <p key={i}>
                            <b>{v[0]}</b>: {x}
                        </p>
                        // <span className="underline">{v[1].join(", ")}</span>
                    ))}
                </div>
            ))}
        </>
    );
}

function StateView(props: { proxy: Proxy; id: string; state: string }) {
    if (props.state == "Waiting Approval") {
        return (
            <div className="flex-1 flex flex-row">
                <button
                    className="bg-gray-700 text-white pl-2 pr-2 pt-1 pb-1"
                    onClick={(e) => {
                        e.stopPropagation();
                        props.proxy.approveRequest(props.id);
                    }}
                >
                    Approve
                </button>
                <button
                    className="ml-2 bg-gray-700 text-white pl-2 pr-2 pt-1 pb-1"
                    onClick={(e) => {
                        e.stopPropagation();
                        props.proxy.cancelRequest(props.id);
                    }}
                >
                    Cancel
                </button>
            </div>
        );
    }
    return (
        <p
            className="flex-1"
            style={{
                color: stateColors[props.state],
            }}
        >
            {props.state}
        </p>
    );
}

function FieldView(props: {
    hide: boolean;
    name: string;
    value: number | string | Record<string, Array<string>> | undefined;
    italic?: boolean;
}) {
    if (props.hide) {
        return <></>;
    }
    return (
        <div className="mb-2 text-lg flex flex-row w-full">
            <b className="flex-1">{props.name} </b>
            <div className="flex-1">
                <ValueView
                    name={props.name}
                    value={props.value}
                    italic={props.italic}
                />
            </div>
        </div>
    );
}

function ShowHideFieldView(props: {
    name: string;
    hide: boolean;
    value: string | undefined;
    hiddenValue: string;
    defaultShow: boolean;
}) {
    const [show, setShow] = useState(props.defaultShow);
    if (props.hide) {
        return <></>;
    }
    return (
        <div className="mb-2 text-lg flex flex-row w-full">
            <b className="flex-1">{props.name}</b>
            <div className="flex-1 text-start">
                <button
                    className="text-sm pl-3 pr-3 bg-gray-600 text-white mr-4"
                    onClick={() => setShow(!show)}
                >
                    {show ? "Hide" : "Show"}
                </button>
                {show ? (
                    <ValueView
                        name={props.name}
                        value={props.value}
                        italic={false}
                    />
                ) : (
                    <i>{props.hiddenValue}</i>
                )}
            </div>
        </div>
    );
}

function BodyView(props: { value: string | undefined; hide: boolean }) {
    const [show, setShow] = useState(false);
    if (props.hide) {
        return <></>;
    }
    return (
        <div className="mb-2 text-lg w-full">
            <div className="flex flex-row items-center">
                <b className="ml-2 flex-1">
                    Body (
                    {props.value !=
                    "body will not be provided under configuration rules"
                        ? props.value?.length
                        : 0}{" "}
                    bytes)
                </b>
                <div className="flex-1 flex flex-row">
                    <button
                        className="text-sm pl-3 pr-3 bg-gray-600 text-white mr-4"
                        onClick={() => setShow(!show)}
                    >
                        {show ? "Hide" : "Show"}
                    </button>
                    {!show && <i className="mr-4">hidden</i>}
                </div>
            </div>

            {show && (
                <pre className="ml-2 mt-2 whitespace-pre-wrap">
                    {props.value}
                </pre>
            )}
        </div>
    );
}
