import { downloadBody, downloadRequest } from "./downloadRequest";
import { Request, RequestsViewConfig } from "./types";
import { Proxy } from "./api/api";

import {
    Collapsible,
    CollapsibleTrigger,
    CollapsibleContent,
} from "./components/ui/collapsible";

import { useRef, useState } from "react";
import { FaRegTrashCan } from "react-icons/fa6";

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
    setRequest: (req: Request) => void;
    requestsViewConfig: RequestsViewConfig;

    imOpen: (setOpen: React.Dispatch<React.SetStateAction<boolean>>) => void;
}) {
    const [editMode, setEditMode] = useState(false);
    const [open, setOpen] = useState(false);
    return (
        <Collapsible
            className="border-b-1 border-b-black wrap-anywhere"
            open={open}
            onOpenChange={(op: boolean) => {
                if (op) {
                    props.imOpen(setOpen);
                }
                setOpen(op);
            }}
        >
            <CollapsibleTrigger className="w-full bg-gray-200">
                <div className="relative flex flex-row w-full pt-4 pb-4">
                    <ParagraphView
                        hide={props.requestsViewConfig.hideDate}
                        className="text-sm"
                    >
                        {props.request.datetime}
                    </ParagraphView>
                    <ParagraphView
                        hide={props.requestsViewConfig.hideHostCollapsed}
                    >
                        {props.request.host}
                    </ParagraphView>
                    <ParagraphView
                        hide={props.requestsViewConfig.hideClientApplication}
                    >
                        {props.request.clientApplication}
                    </ParagraphView>
                    <StateView
                        proxy={props.proxy}
                        id={props.request.id}
                        state={props.request.state}
                        setEditMode={setEditMode}
                        hide={props.requestsViewConfig.hideState}
                    />
                </div>
            </CollapsibleTrigger>
            <CollapsibleContent className="bg-gray-300 max-h-[50vh] overflow-y-auto">
                <div className="flex flex-row">
                    <button
                        className="bg-gray-600 text-white border-black border-1 ml-2 mt-2 pl-2 pr-2"
                        onClick={() => {
                            downloadRequest(props.request);
                        }}
                    >
                        Download
                    </button>
                    <EditButton
                        proxy={props.proxy}
                        request={props.request}
                        editMode={editMode}
                        setEditMode={setEditMode}
                    />
                </div>
                <div className="ml-2 pt-2 pb-1">
                    {props.request.error != undefined &&
                    !props.requestsViewConfig.hideError ? (
                        <p>
                            <span className="text-red-700">Error</span>:
                            {props.request.error}
                        </p>
                    ) : (
                        <></>
                    )}
                    <FieldView
                        name="ID"
                        value={props.request.id}
                        hide={props.requestsViewConfig.hideID}
                        editMode={editMode}
                        disableEdits={true}
                    />
                    <p>
                        <b>Request: </b>
                    </p>
                    <div className="hashi w-full pt-2 pl-5 pr-5">
                        <FieldView
                            name="Host"
                            value={props.request.host}
                            hide={props.requestsViewConfig.hideHost}
                            editMode={editMode}
                            setValue={(v: string) => {
                                props.setRequest({
                                    ...props.request,
                                    host: v,
                                });
                            }}
                        />
                        <FieldView
                            name="Client IP"
                            value={props.request.clientIP}
                            hide={props.requestsViewConfig.hideClientIP}
                            editMode={editMode}
                            disableEdits={true}
                        />
                        <FieldView
                            name="Client Username"
                            value={props.request.clientAuthorizationUser}
                            hide={props.requestsViewConfig.hideClientUser}
                            editMode={editMode}
                        />
                        <ShowHideFieldView
                            name="Client Password"
                            value={props.request.clientAuthorizationPassword}
                            hiddenValue="********"
                            defaultShow={false}
                            hide={props.requestsViewConfig.hideClientPassword}
                            editMode={editMode}
                        />
                        <FieldView
                            name="Method"
                            value={props.request.method}
                            hide={props.requestsViewConfig.hideMethod}
                            editMode={editMode}
                            setValue={(v) => {
                                props.setRequest({
                                    ...props.request,
                                    method: v,
                                });
                            }}
                        />
                        <FieldView
                            name="Path"
                            value={props.request.path}
                            hide={props.requestsViewConfig.hidePath}
                            editMode={editMode}
                            setValue={(v) => {
                                props.setRequest({
                                    ...props.request,
                                    path: v,
                                });
                            }}
                        />
                        <FieldView
                            name="Query"
                            value={props.request.query}
                            hide={props.requestsViewConfig.hideQuery}
                            editMode={editMode}
                            setValue={(v) =>
                                props.setRequest({
                                    ...props.request,
                                    query: v,
                                })
                            }
                        />
                        <FieldView
                            name="Headers"
                            value={props.request.headers}
                            hide={props.requestsViewConfig.hideRequestHeaders}
                            editMode={editMode}
                            setValue={(v) => {
                                props.setRequest({
                                    ...props.request,
                                    headers: v,
                                });
                            }}
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
                            editMode={editMode}
                            setValue={(v: string) => {
                                props.setRequest({
                                    ...props.request,
                                    body: v,
                                });
                            }}
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
                            editMode={editMode}
                            disableEdits={true}
                        />
                        <FieldView
                            name="Headers"
                            value={props.request.response?.headers}
                            hide={props.requestsViewConfig.hideResponseHeaders}
                            editMode={editMode}
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
                            editMode={editMode}
                            setValue={(v: string) => {
                                //NOTE: not implemented
                                props.setRequest({
                                    ...props.request,
                                    response: {
                                        ...props.request.response!,
                                        body: v,
                                    },
                                });
                            }}
                        />
                    </div>
                    <FieldView
                        name="Bytes Transferred"
                        hide={props.requestsViewConfig.hideBytesTransferred}
                        value={props.request.bytesTransferred}
                        editMode={editMode}
                        disableEdits={true}
                    ></FieldView>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

function MapView(props: {
    editMode: boolean;
    disableEdits?: boolean;
    value: Record<string, Array<string>>;
    setValue: (v: Record<string, Array<string>>) => void;
}) {
    const objectKeyRef = useRef<HTMLInputElement | null>(null);
    const objectValueRef = useRef<HTMLInputElement | null>(null);

    return (
        <>
            {Object.keys(props.value!).length !== 0 ? (
                Object.entries(props.value!).map((v) => (
                    <div key={v[0]} className="flex flex-row gap-3 text-sm">
                        {v[1].map((x, i) => (
                            <p key={i}>
                                <b>{v[0]}</b>: {x}
                            </p>
                        ))}

                        {props.editMode && !props.disableEdits ? (
                            <button
                                onClick={() => {
                                    const newValue = {
                                        ...(props.value as Record<
                                            string,
                                            Array<string>
                                        >),
                                    };
                                    if (newValue[v[0]]) {
                                        delete newValue[v[0]];
                                    }
                                    props.setValue!(newValue);
                                }}
                            >
                                <FaRegTrashCan />
                            </button>
                        ) : (
                            <></>
                        )}
                    </div>
                ))
            ) : (
                <i className="text-sm">none or unavailable</i>
            )}
            {props.editMode && !props.disableEdits ? (
                <div className="flex flex-row gap-3 text-sm max-h-fit items-center mb-4">
                    <input
                        type="text"
                        placeholder="Key"
                        ref={objectKeyRef}
                        className="pt-1 pb-1 pl-2 mt-2 border-2 border-black rounded-2xl text-sm w-[40%]"
                    ></input>
                    <input
                        type="text"
                        placeholder="Value"
                        ref={objectValueRef}
                        className="pt-1 pb-1 pl-2 mt-2 border-2 border-black rounded-2xl text-sm w-[40%]"
                    ></input>
                    <button
                        className="pl-2 pr-2 bg-black aspect-square text-white mt-auto"
                        onClick={() => {
                            const newValue = {
                                ...(props.value as Record<
                                    string,
                                    Array<string>
                                >),
                            };
                            const key = objectKeyRef.current!.value;
                            const val = objectValueRef.current!.value;
                            if (newValue[key] != undefined) {
                                const v = newValue[key];
                                v.push(val);
                                newValue[key] = v;
                            } else {
                                newValue[key] = [val];
                            }
                            props.setValue!(newValue);
                            objectKeyRef.current!.value = "";
                            objectValueRef.current!.value = "";
                        }}
                    >
                        +
                    </button>
                </div>
            ) : (
                <></>
            )}
        </>
    );
}

function ValueView<
    T extends number | string | Record<string, Array<string>> | undefined,
>(props: {
    name: string;
    value: T;
    setValue?: (v: T) => void;
    editMode: boolean;
    disableEdits?: boolean;
}) {
    if (props.value == undefined || props.value === "") {
        return <i>none or unavailable</i>;
    }
    if (typeof props.value === "object") {
        return (
            <MapView
                editMode={props.editMode}
                disableEdits={props.disableEdits}
                value={props.value as Record<string, Array<string>>}
                setValue={
                    props.setValue as (v: Record<string, Array<string>>) => void
                }
            />
        );
    }
    if (props.editMode && !props.disableEdits) {
        return (
            <input
                type={typeof props.value === "number" ? "number" : "text"}
                className="bg-gray-200 text-black w-full pl-2"
                defaultValue={props.value}
                onBlur={(e) => {
                    const v = e.target.value as T;
                    props.setValue!(v);
                }}
            />
        );
    }

    return <>{props.value}</>;
}

function StateView(props: {
    proxy: Proxy;
    id: string;
    state: string;
    setEditMode: React.Dispatch<React.SetStateAction<boolean>>;
    hide: boolean;
}) {
    if (props.hide) {
        return <></>;
    }
    if (props.state == "Waiting Approval") {
        return (
            <div className="flex-1 flex flex-row content-center items-center justify-center">
                <button
                    className="bg-gray-700 text-white pl-2 pr-2 pt-1 pb-1"
                    onClick={(e) => {
                        e.stopPropagation();
                        props.setEditMode(false);
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

function EditButton(props: {
    proxy: Proxy;
    request: Request;
    editMode: boolean;
    setEditMode: React.Dispatch<React.SetStateAction<boolean>>;
}) {
    if (props.request.state !== "Waiting Approval") {
        return <></>;
    }
    return (
        <button
            className="bg-gray-600 text-white border-black border-1 ml-2 mt-2 pl-2 pr-2"
            onClick={() => {
                if (props.editMode) {
                    // pressed save
                    console.log("pre", props.request);
                    props.setEditMode(false);
                    props.proxy.updateRequest(props.request);
                } else {
                    props.setEditMode(true);
                }
            }}
        >
            {props.editMode ? "Save" : "Edit"}
        </button>
    );
}

function FieldView<
    T extends string | number | Record<string, string[]> | undefined,
>(props: {
    disableEdits?: boolean;
    editMode: boolean;
    hide: boolean;
    name: string;
    value: T;
    setValue?: (v: T) => void;
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
                    editMode={props.editMode}
                    disableEdits={props.disableEdits}
                    setValue={props.setValue}
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
    editMode: boolean;
}) {
    const [show, setShow] = useState(props.defaultShow);

    const isValueEmpty = props.value === undefined || props.value.trim() === "";

    if (props.hide) {
        return <></>;
    }

    return (
        <div className="mb-2 text-lg flex flex-row w-full">
            <b className="flex-1">{props.name}</b>
            <div className="flex-1 text-start">
                {!isValueEmpty && (
                    <button
                        className="text-sm pl-3 pr-3 bg-gray-600 text-white mr-4"
                        onClick={() => setShow(!show)}
                    >
                        {show ? "Hide" : "Show"}
                    </button>
                )}
                {isValueEmpty ? (
                    <i>none or unavailable</i>
                ) : show ? (
                    <ValueView
                        name={props.name}
                        value={props.value}
                        editMode={props.editMode}
                    />
                ) : (
                    <i>{props.hiddenValue}</i>
                )}
            </div>
        </div>
    );
}

function BodyView(props: {
    value: string | null | undefined;
    hide: boolean;
    editMode: boolean;
    setValue?: (v: string) => void;
}) {
    const bodyBytes =
        props.value != null && props.value != undefined
            ? props.value.length
            : 0;
    const [showBody, setShowBody] = useState<boolean>(
        bodyBytes == 0 ? true : false,
    );

    if (props.hide) {
        return <></>;
    }
    return (
        <div className="mb-2 text-lg w-full">
            <div className="flex flex-row items-center">
                <b className="ml-2 flex-1">Body ({bodyBytes} bytes)</b>
                <div className="flex-1 flex flex-row">
                    {bodyBytes != 0 ? (
                        <>
                            <button
                                className="text-sm pl-3 pr-3 bg-gray-600 text-white mr-4"
                                onClick={() => setShowBody(!showBody)}
                            >
                                {showBody ? "Hide" : "Show"}
                            </button>
                            {!showBody && <i className="mr-4">hidden</i>}
                        </>
                    ) : (
                        <></>
                    )}
                </div>
            </div>
            <div className="ml-2 font-[monospace]">
                {props.editMode && showBody ? (
                    <textarea
                        defaultValue={props.value ?? ""}
                        className="w-full border-2 border-black p-1"
                        onBlur={(e) => {
                            props.setValue!(e.target.value);
                        }}
                    ></textarea>
                ) : showBody ? (
                    <pre className="ml-2 mt-2 whitespace-pre-wrap">
                        {props.value}
                    </pre>
                ) : null}{" "}
            </div>
        </div>
    );
}

function ParagraphView(props: {
    children?: React.ReactNode;
    hide: boolean;
    className?: string;
}) {
    console.log(props);
    if (props.hide) {
        return <></>;
    }
    return (
        <p className={"flex-1 text-center" + props.className}>
            {props.children != "" ? props.children : <>-</>}
        </p>
    );
}
