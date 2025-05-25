import { useState } from "react";
import {
    Collapsible,
    CollapsibleTrigger,
    CollapsibleContent,
} from "./components/ui/collapsible";
import { Request } from "./types";
import { downloadBody, downloadRequest } from "./downloadRequest";

const stateColors = {
    Processing: "#000",
    Canceled: "#806262",
    Done: "#62806b",
    Error: "oklch(50.5% 0.213 27.518)",
};
export function RequestView(props: { request: Request }) {
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
                    <p className="flex-1 ml-1 text-sm">{props.request.id}</p>
                    <p className="flex-1 ml-1 text-center">
                        {props.request.host}
                    </p>
                    <p
                        className="flex-1"
                        style={{
                            color: stateColors[props.request.state],
                        }}
                    >
                        {props.request.state}
                    </p>
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
                    {props.request.error != undefined ? (
                        <p>
                            <span className="text-red-700">Error</span>:{" "}
                            {props.request.error}
                        </p>
                    ) : (
                        <></>
                    )}
                    <p>
                        <b>Request: </b>
                    </p>
                    <div className="hashi w-full pt-2 pl-5 pr-5">
                        <FieldView
                            name="Client Username"
                            value={props.request.clientAuthorizationUser}
                        />
                        <ShowHideFieldView
                            name="Client Password"
                            value={props.request.clientAuthorizationPassword}
                            hiddenValue="********"
                            defaultShow={false}
                        />
                        <FieldView name="Method" value={props.request.method} />
                        <FieldView name="Path" value={props.request.path} />
                        <FieldView
                            name="Headers"
                            value={props.request.headers}
                        />
                        <button
                            className="bg-gray-600 text-white border-black border-1 mt-2 pl-2 pr-2"
                            onClick={() =>
                                downloadBody(
                                    props.request.id,
                                    props.request.body,
                                    props.request.headers!["Content-Type"][0],
                                )
                            }
                        >
                            Download Body
                        </button>
                        <ShowHideFieldView
                            name={
                                "Body (" +
                                props.request.body?.length +
                                " bytes)"
                            }
                            value={props.request.body}
                            hiddenValue=""
                            defaultShow={false}
                        />
                    </div>
                    <p>
                        <b>Response: </b>
                    </p>
                    <div className="mt-2 ml-5">
                        <FieldView
                            name="Status"
                            value={props.request.response?.statusCode}
                        />
                        <FieldView
                            name="Headers"
                            value={props.request.response?.headers}
                        />
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
                        <ShowHideFieldView
                            name={
                                "Body (" +
                                props.request.response?.body?.length +
                                " bytes)"
                            }
                            value={props.request.response?.body}
                            defaultShow={false}
                            hiddenValue=""
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
                <p key={v[0]} className="text-sm">
                    <b>{v[0]}</b>: {v[1].join(", ")}
                </p>
            ))}
        </>
    );
}
function FieldView(props: {
    name: string;
    value: number | string | Record<string, Array<string>> | undefined;
    italic?: boolean;
}) {
    return (
        <div className="mb-2 text-lg flex flex-row w-full">
            <b className="flex-1">{props.name} </b>
            <div className="flex-1 text-start">
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
    value: string | undefined;
    hiddenValue: string;
    defaultShow: boolean;
}) {
    const [show, setShow] = useState(props.defaultShow);
    return (
        <div className="mb-2 text-lg flex flex-row w-full">
            <b className="flex-1">{props.name}</b>
            <div className="flex-1 text-start">
                <button
                    className="text-sm pl-3 pr-3 bg-gray-600 text-white"
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
