import { downloadBody, downloadRequest } from "./downloadRequest";
import { Request, RequestsViewConfig } from "./types";
import { Proxy } from "./api/api";
import { CiLock, CiUnlock } from "react-icons/ci";
import { JsonEditor } from "json-edit-react";

import {
    Collapsible,
    CollapsibleTrigger,
    CollapsibleContent,
} from "./components/ui/collapsible";

import { useContext, useRef, useState } from "react";
import { FaRegTrashCan } from "react-icons/fa6";
import { Timeline } from "./Timeline";
import { nsToReadable, pascalCaseToCapitalSpace } from "./utils";
import { StatusCodes } from "./statuscodes";
import { DarkModeContext } from "./context/context";
import { encodeURLEncoded, parseURLEncoded } from "./encoding/urlencoded";

const stateColors: Record<string, string> = {
    Processing: "#000",
    Canceled: "#806262",
    Done: "#62806b",
    Error: "oklch(50.5% 0.213 27.518)",
    "Approval Timeout": "#806262",
    "Waiting Approval": "#806262",
};

const darkStateColors: Record<string, string> = {
    Processing: "#e0e0e0", // Light gray for contrast
    Canceled: "#f2a6a6", // Light desaturated red
    Done: "#a6f2c3", // Soft mint green
    Error: "oklch(85% 0.2 27.5)", // Lighter and more saturated version
    "Approval Timeout": "#f2a6a6", // Same as Canceled
    "Waiting Approval": "#f2a6a6", // Same as Canceled
};

export function RequestView(props: {
    proxy: Proxy;
    request: Request;
    setRequest: (req: Request) => void;
    requestsViewConfig: RequestsViewConfig;

    open: boolean;
    setOpen: (o: boolean) => void;
}) {
    const [editMode, setEditMode] = useState(false);
    return (
        <Collapsible
            className="border-b border-gray-300"
            open={props.open}
            onOpenChange={(o) => {
                props.setOpen(o);
            }}
        >
            <CollapsibleTrigger className="w-full bg-white dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-950 transition-colors">
                <div className="flex flex-wrap items-center justify-between gap-4 px-4 py-3">
                    <ParagraphView
                        hide={props.requestsViewConfig.hideDate}
                        className="text-sm text-gray-600 dark:text-gray-500"
                    >
                        {formatDate(props.request.datetime)}
                    </ParagraphView>

                    <ParagraphView
                        hide={props.requestsViewConfig.hideHostCollapsed}
                    >
                        <div className="flex w-full justify-center items-center gap-2 text-md">
                            {props.request.host}
                        </div>
                    </ParagraphView>

                    <ParagraphView
                        hide={props.requestsViewConfig.hideClientApplication}
                        className="text-sm text-gray-700"
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

            <CollapsibleContent className="bg-gray-300 dark:bg-gray-900 max-h-[60vh] overflow-y-auto p-4 space-y-4 font-[monospace] text-black dark:text-white">
                <div className="flex gap-3">
                    <button
                        className="bg-blue-600 hover:bg-blue-700 dark:bg-blue-300 dark:hover-blue-400 text-white font-semibold px-3 py-1 rounded shadow"
                        onClick={() => downloadRequest(props.request)}
                    >
                        Download Request
                    </button>
                    <EditButton
                        proxy={props.proxy}
                        request={props.request}
                        editMode={editMode}
                        setEditMode={setEditMode}
                    />
                </div>

                {props.request.error && !props.requestsViewConfig.hideError && (
                    <p className="text-blue-800 dark:text-blue-400">
                        <b>Error:</b> {props.request.error}
                    </p>
                )}

                {/* Request Info */}
                <div className="bg-white dark:bg-gray-700 rounded-lg shadow p-4 space-y-3 font-[Manrope]">
                    <div className="flex flex-row items-center gap-2">
                        {!props.requestsViewConfig.hideMethod &&
                        props.request.method !== undefined ? (
                            <p
                                className="w-16 min-w-fit px-1 h-8 items-center flex flex-row justify-center text-black dark:text-white"
                                style={{
                                    backgroundColor: getMethodColor(
                                        props.request.method!,
                                    ),
                                }}
                                onClick={() => {
                                    if (editMode) {
                                        const methods = [
                                            "GET",
                                            "POST",
                                            "PUT",
                                            "PATCH",
                                            "DELETE",
                                            "OPTIONS",
                                            "HEAD",
                                        ];
                                        const currentIndex = methods.indexOf(
                                            props.request.method!,
                                        );
                                        const nextIndex =
                                            (currentIndex + 1) % methods.length;
                                        const nextMethod = methods[nextIndex];
                                        props.setRequest({
                                            ...props.request,
                                            method: nextMethod,
                                        });
                                    }
                                }}
                            >
                                {props.request.method}
                            </p>
                        ) : null}
                        <div className="flex flex-row items-center gap-2 text-2xl">
                            {props.request.secure ? (
                                <CiLock
                                    className="text-green-800 dark:text-green-400"
                                    onClick={() => {
                                        props.setRequest({
                                            ...props.request,
                                            secure: false,
                                        });
                                    }}
                                />
                            ) : (
                                <CiUnlock
                                    className="text-red-600 dark:text-red-400"
                                    onClick={() => {
                                        props.setRequest({
                                            ...props.request,
                                            secure: true,
                                        });
                                    }}
                                />
                            )}
                            <b>Request</b>{" "}
                            {!props.requestsViewConfig.hideID
                                ? props.request.id
                                : null}{" "}
                            {!props.requestsViewConfig.hideClientIP ? (
                                <span>
                                    by <b>{props.request.clientIP}</b>
                                </span>
                            ) : null}
                        </div>
                    </div>

                    <FieldView
                        name="Host"
                        value={props.request.host}
                        hide={props.requestsViewConfig.hideHost}
                        editMode={editMode}
                        setValue={(v: string) => {
                            props.setRequest({ ...props.request, host: v });
                        }}
                    />
                    <div className="mt-4"></div>
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
                        name="Path"
                        value={props.request.path}
                        hide={props.requestsViewConfig.hidePath}
                        editMode={editMode}
                        setValue={(v) =>
                            props.setRequest({ ...props.request, path: v })
                        }
                    />
                    <FieldView
                        name="Query"
                        value={props.request.query}
                        hide={props.requestsViewConfig.hideQuery}
                        editMode={editMode}
                        setValue={(v) =>
                            props.setRequest({ ...props.request, query: v })
                        }
                    />
                    <FieldView
                        name="Headers"
                        value={props.request.headers}
                        hide={props.requestsViewConfig.hideRequestHeaders}
                        editMode={editMode}
                        setValue={(v) =>
                            props.setRequest({ ...props.request, headers: v })
                        }
                    />

                    {!props.requestsViewConfig.hideRequestBody && (
                        <>
                            <BodyView
                                request={props.request}
                                isRequestBody={true}
                                hide={false}
                                editMode={editMode}
                                setValue={(v: string | undefined) => {
                                    props.setRequest({
                                        ...props.request,
                                        tempBody: v,
                                    });
                                }}
                                loadBody={async () => {
                                    const body =
                                        await props.proxy.getRequestBody(
                                            props.request.id,
                                        );
                                    props.setRequest({
                                        ...props.request,
                                        tempBody: body,
                                    });
                                }}
                            />
                        </>
                    )}
                </div>

                {/* Response Info */}
                <div className="bg-white dark:bg-gray-700 rounded-lg shadow p-4 space-y-3">
                    <h2 className="text-lg font-semibold">Response</h2>
                    <div className="flex flex-row text-lg gap-2">
                        {props.request.response?.statusCode != undefined ? (
                            <>
                                <p
                                    className="mt-auto mb-auto w-3 h-3 rounded-4xl"
                                    style={{
                                        backgroundColor: getStatusCodeBGColor(
                                            props.request.response!.statusCode!,
                                        ),
                                    }}
                                ></p>
                                {props.request.response?.statusCode}{" "}
                                {statusCodeToName(
                                    props.request.response?.statusCode,
                                )}
                            </>
                        ) : null}
                    </div>
                    {/* <FieldView
                        name="Status"
                        value={props.request.response?.statusCode}
                        hide={props.requestsViewConfig.hideResponseStatus}
                        editMode={editMode}
                        disableEdits
                    /> */}
                    <FieldView
                        name="Headers"
                        value={props.request.response?.headers}
                        hide={props.requestsViewConfig.hideResponseHeaders}
                        editMode={editMode}
                    />
                    {!props.requestsViewConfig.hideResponseBody && (
                        <>
                            <BodyView
                                request={props.request}
                                isRequestBody={false}
                                hide={false}
                                editMode={editMode}
                                setValue={(v: string | undefined) =>
                                    props.setRequest({
                                        ...props.request,
                                        response: {
                                            tempBody: v,
                                        },
                                    })
                                }
                                loadBody={async () => {
                                    const body =
                                        await props.proxy.getResponseBody(
                                            props.request.id,
                                        );
                                    console.log(body);
                                    props.setRequest({
                                        ...props.request,
                                        response: {
                                            tempBody: body,
                                        },
                                    });
                                }}
                            />
                        </>
                    )}
                </div>

                <FieldView
                    name="Bytes Transferred"
                    hide={props.requestsViewConfig.hideBytesTransferred}
                    value={props.request.bytesTransferred}
                    editMode={editMode}
                    disableEdits
                />

                {props.request.timing && props.request.timing_total && (
                    <div>
                        <h2 className="text-lg font-semibold mt-4">
                            Timeline ({nsToReadable(props.request.timing_total)}
                            )
                        </h2>
                        <Timeline
                            timing={props.request.timing}
                            totalTime={props.request.timing_total}
                            className="mt-2"
                        />
                    </div>
                )}
            </CollapsibleContent>
        </Collapsible>
    );
}

function getStatusCodeBGColor(statusCode: number) {
    if (statusCode >= 200 && statusCode < 300) {
        return "#4CAF50"; // Vibrant Green
    } else if (statusCode >= 300 && statusCode < 400) {
        return "#FFEB3B"; // Amber
    } else if (statusCode >= 400 && statusCode < 500) {
        return "#FF5722"; // Deep Orange/Red
    } else if (statusCode >= 500) {
        return "#D32F2F"; // Darker Red
    }
    return "#424242"; // Dark Grey
}

function statusCodeToName(statusCode: number | undefined): string {
    return pascalCaseToCapitalSpace(StatusCodes[statusCode!] || "");
}

function getMethodColor(method: string): string {
    // colors are taken from Postman
    switch (method) {
        case "GET":
            return "#007F31";
        case "POST":
            return "#AD7A03";
        case "PUT":
            return "#0053B8";
        case "PATCH":
            return "#623497";
        case "DELETE":
            return "#8E1A10";
        case "OPTIONS":
            return "#A61468";
        case "HEAD":
        default:
            return "#9E9E9E"; // Grey for unknown methods
    }
}

function TableMapView(props: {
    editMode: boolean;
    disableEdits?: boolean;
    value: Record<string, Array<string>>;
    setValue: (v: Record<string, Array<string>>) => void;
}) {
    const objectKeyRef = useRef<HTMLInputElement | null>(null);
    const objectValueRef = useRef<HTMLInputElement | null>(null);

    return (
        <div className="flex flex-col gap-2">
            <table className="table-auto border-collapse border border-gray-400 text-sm">
                <thead>
                    <tr className="bg-gray-100 dark:bg-gray-800">
                        <th className="border border-gray-300 dark:border-gray-500 px-2 py-1 w-64">
                            Key
                        </th>
                        <th className="border border-gray-300 dark:border-gray-500 px-2 py-1 w-96">
                            Values
                        </th>
                        {props.editMode && !props.disableEdits && (
                            <th className="border border-gray-300 dark:border-gray-500 px-2 py-1">
                                Actions
                            </th>
                        )}
                    </tr>
                </thead>
                <tbody>
                    {Object.entries(props.value).map(([key, values]) => {
                        return values.map((val, i) => (
                            <tr key={`${key}-${values}-${val}-i`}>
                                <td className="border border-gray-300 dark:border-gray-500 px-2 py-1 w-64 break-all">
                                    {key}
                                </td>
                                <td className="border border-gray-300  dark:border-gray-500 px-2 py-1 w-96 break-all">
                                    <span key={i} className="block">
                                        {val}
                                    </span>
                                </td>
                                {props.editMode && !props.disableEdits && (
                                    <td className="border border-gray-300 dark:border-gray-500 px-2 py-1">
                                        <button
                                            onClick={() => {
                                                const newValue = {
                                                    ...props.value,
                                                };
                                                if (newValue[key].length == 1) {
                                                    delete newValue[key];
                                                } else {
                                                    newValue[key] = newValue[
                                                        key
                                                    ].filter(
                                                        (_, index) =>
                                                            index !== i,
                                                    );
                                                }
                                                props.setValue(newValue);
                                            }}
                                        >
                                            <FaRegTrashCan />
                                        </button>
                                    </td>
                                )}
                            </tr>
                        ));
                    })}
                </tbody>
            </table>

            {props.editMode && !props.disableEdits && (
                <div className="flex flex-row gap-3 text-sm items-center mt-2">
                    <input
                        type="text"
                        placeholder="Key"
                        ref={objectKeyRef}
                        className="border-2 border-black rounded-2xl px-2 py-1 w-[40%]"
                    />
                    <input
                        type="text"
                        placeholder="Value"
                        ref={objectValueRef}
                        className="border-2 border-black rounded-2xl px-2 py-1 w-[40%]"
                    />
                    <button
                        className="bg-black text-white rounded-full px-3 py-1"
                        onClick={() => {
                            const newValue = { ...props.value };
                            const key = objectKeyRef.current!.value;
                            const val = objectValueRef.current!.value;
                            if (!key || !val) return;
                            if (newValue[key]) {
                                newValue[key].push(val);
                            } else {
                                newValue[key] = [val];
                            }
                            props.setValue(newValue);
                            objectKeyRef.current!.value = "";
                            objectValueRef.current!.value = "";
                        }}
                    >
                        +
                    </button>
                </div>
            )}
        </div>
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
            <TableMapView
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
                className="bg-gray-200 text-black w-full pl-2 wrap-anywhere"
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
    const [darkMode, _] = useContext(DarkModeContext);
    if (props.hide) {
        return <></>;
    }
    if (props.state == "Waiting Approval") {
        return (
            <div className="flex-1 flex flex-row content-center items-center justify-center">
                <button
                    className="bg-gray-700 dark:bg-gray-300 text-white dark:text-black px-3 pt-1 pb-1 rounded-xl"
                    onClick={(e) => {
                        e.stopPropagation();
                        props.setEditMode(false);
                        props.proxy.approveRequest(props.id);
                    }}
                >
                    Approve
                </button>
                <button
                    className="ml-2 bg-gray-700 dark:bg-gray-300 dark:text-black text-white px-3 pt-1 pb-1 rounded-xl"
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
                color: darkMode
                    ? darkStateColors[props.state]
                    : stateColors[props.state],
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
            className="bg-gray-600 hover:bg-gray-700 border-1 border-black text-white font-semibold px-2 rounded shadow"
            // className="bg-gray-600 text-white border-black border-1 ml-2 mt-2 pl-2 pr-2"
            onClick={() => {
                if (props.editMode) {
                    // pressed save
                    props.setEditMode(false);
                    console.log(props.request);
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
        <div
            className="mb-2 gap-4 text-md flex flex-row w-full"
            style={{
                alignItems: typeof props.value != "object" ? "center" : "",
            }}
        >
            <div className="w-64">
                <b className="">{props.name}:</b>
            </div>
            <div className="min-w-64 w-fit">
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
        <div className="mb-2 text-md gap-4 flex flex-row w-full">
            <b className="w-64">{props.name}: </b>
            <div className="min-w-64 w-fit text-start">
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
                {!isValueEmpty && (
                    <button
                        className="text-sm pl-3 pr-3 bg-gray-600 dark:bg-gray-300 text-white ml-4"
                        onClick={() => setShow(!show)}
                    >
                        {show ? "Hide" : "Show"}
                    </button>
                )}
            </div>
        </div>
    );
}

function BodyView(props: {
    request: Request;
    isRequestBody?: boolean;
    hide: boolean;
    editMode: boolean;
    setValue?: (v: string | undefined) => void;
    loadBody: () => void;
}) {
    const headers = props.isRequestBody
        ? props.request.headers
        : props.request.response?.headers;
    let bodyBytes = 0;
    if (props.isRequestBody) {
        bodyBytes = props.request.tempBody
            ? props.request.tempBody.length
            : props.request.bodyLength || 0;
    } else {
        bodyBytes = props.request.response?.tempBody
            ? props.request.response.tempBody.length
            : props.request.response?.bodyLength || 0;
    }
    const [showBody, setShowBody] = useState<boolean>(false);

    const tempBody = props.isRequestBody
        ? props.request.tempBody
        : props.request.response?.tempBody;

    if (props.hide) {
        return <></>;
    }

    return (
        <div className="mb-2 text-lg w-full">
            <div className="flex flex-row gap-6">
                <button
                    className="text-sm bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded shadow mt-2 disabled:opacity-50 disabled:hover:bg-blue-600 disabled:cursor-not-allowed"
                    disabled={bodyBytes == 0}
                    onClick={() =>
                        downloadBody(
                            props.request.id,
                            tempBody,
                            headers?.["Content-Type"]?.[0],
                        )
                    }
                >
                    Download Body
                </button>
            </div>
            <div className="flex flex-row items-center mt-1 mb-2">
                <b className="">Body ({bodyBytes} bytes)</b>
                <div className="ml-4 flex flex-row">
                    {/* if no body, don't show; if not editing; show button; if editing BUT the body isn't showing; show button; */}
                    {bodyBytes != 0 && (!props.editMode || !showBody) ? (
                        <>
                            <button
                                className="text-md pl-3 pr-3 bg-gray-600 dark:bg-gray-300 text-white dark:text-black mr-4"
                                onClick={() => {
                                    if (!showBody) {
                                        // show it now
                                        props.loadBody(); // loading it now
                                    } else if (tempBody != undefined) {
                                        // hiding it now
                                        props.setValue!(undefined);
                                    }
                                    setShowBody(!showBody);
                                }}
                            >
                                {showBody ? "Hide" : "Show"}
                            </button>
                        </>
                    ) : (
                        <></>
                    )}
                </div>
            </div>
            <BodyContentView
                contentType={headers?.["Content-Type"]?.[0]}
                body={tempBody}
                editMode={props.editMode}
                showBody={showBody}
                setValue={props.setValue!}
            />
        </div>
    );
}

function BodyContentView(props: {
    contentType?: string;
    body?: string | null;

    editMode: boolean;
    showBody?: boolean;
    setValue: (v: string) => void;
}) {
    if (!props.showBody) {
        return <></>;
    }
    if (!props.body) {
        if (props.editMode) {
            return (
                <GenericContent
                    editMode={props.editMode}
                    body=""
                    setValue={props.setValue}
                />
            );
        } else {
            return <></>;
        }
    }
    switch (props.contentType) {
        case "application/json":
            try {
                return (
                    <JSONContent
                        body={props.body}
                        editMode={props.editMode}
                        setBody={props.setValue}
                    />
                );
            } catch {
                return (
                    <GenericContent
                        editMode={props.editMode}
                        body={props.body!}
                        setValue={props.setValue}
                    />
                );
            }
        case "application/x-www-form-urlencoded":
            try {
                return (
                    <URLFormEncodedContent
                        editMode={props.editMode}
                        body={props.body!}
                        setValue={props.setValue}
                    />
                );
            } catch {
                return (
                    <GenericContent
                        editMode={props.editMode}
                        body={props.body!}
                        setValue={props.setValue}
                    />
                );
            }
    }
    return (
        <GenericContent
            editMode={props.editMode}
            body={props.body!}
            setValue={props.setValue}
        />
    );
}

function JSONContent(props: {
    body: string;
    setBody: (v: string) => void;
    editMode: boolean;
}) {
    return (
        <JsonEditor
            data={JSON.parse(props.body)}
            setData={(v) => props.setBody(JSON.stringify(v))}
            restrictEdit={!props.editMode}
            restrictAdd={!props.editMode}
            restrictDelete={!props.editMode}
            rootName=""
            showCollectionCount={false}
            showArrayIndices={false}
        />
    );
}

function GenericContent(props: {
    editMode: boolean;
    body: string;
    setValue: (v: string) => void;
}) {
    if (props.editMode) {
        return (
            <textarea
                defaultValue={props.body ?? ""}
                className="w-full border-2 border-black p-1"
                onBlur={(e) => {
                    props.setValue!(e.target.value);
                }}
            ></textarea>
        );
    }
    return (
        <div className="ml-2 font-[monospace]">
            <pre className="ml-2 mt-2 whitespace-pre-wrap wrap-anywhere">
                {props.body}
            </pre>
        </div>
    );
}

function URLFormEncodedContent(props: {
    body: string;
    setValue: (v: string) => void;

    editMode: boolean;
}) {
    return (
        <div className="w-full">
            <TableMapView
                editMode={props.editMode}
                disableEdits={false}
                value={parseURLEncoded(props.body)}
                setValue={(v: Record<string, Array<string>>) =>
                    props.setValue(encodeURLEncoded(v))
                }
            />
        </div>
    );
}

function ParagraphView(props: {
    children?: React.ReactNode;
    hide: boolean;
    className?: string;
}) {
    if (props.hide) {
        return <></>;
    }
    return (
        <div
            className={
                "flex-1 text-center text-black dark:text-white " +
                props.className
            }
        >
            {props.children != "" ? props.children : <i>unavailable</i>}
        </div>
    );
}

function formatDate(dateMS: number): string {
    const date = new Date(dateMS);
    const months = [
        "Jan",
        "Feb",
        "Mar",
        "Apr",
        "May",
        "Jun",
        "Jul",
        "Aug",
        "Sep",
        "Oct",
        "Nov",
        "Dec",
    ];

    const month = months[date.getMonth()];
    const day = String(date.getDate()).padStart(2, "0");
    const year = date.getFullYear();

    const hours = String(date.getHours() % 12).padStart(2, "0");
    const minutes = String(date.getMinutes()).padStart(2, "0");
    const seconds = String(date.getSeconds()).padStart(2, "0");
    const millis = String(date.getMilliseconds()).padStart(3, "0");

    const period = date.getHours() >= 12 ? "PM" : "AM";

    return `${month} ${day}, ${year} ${hours}:${minutes}:${seconds}.${millis} ${period}`;
}
