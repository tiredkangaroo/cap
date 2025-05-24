import {
    Collapsible,
    CollapsibleTrigger,
    CollapsibleContent,
} from "./components/ui/collapsible";
import { Request } from "./types";

export function RequestView(props: { request: Request }) {
    console.log(props.request);
    return (
        <Collapsible className="border-b-1 border-b-black wrap-anywhere">
            <CollapsibleTrigger className="w-full bg-gray-200">
                <div className="relative flex flex-row w-full pt-4 pb-4">
                    <p
                        className="flex-1"
                        style={{
                            color: props.request.secure ? "#0b5c00" : "#5c0300",
                        }}
                    >
                        {props.request.secure ? "HTTPS" : "HTTP"}
                    </p>
                    <p className="flex-1 ml-1 text-sm">{props.request.id}</p>
                    <p className="flex-1 ml-1 text-center">
                        {props.request.host}
                    </p>
                    <p className="flex-1 ml-1 mr-1">{props.request.clientIP}</p>
                </div>
            </CollapsibleTrigger>
            <CollapsibleContent className="bg-gray-300">
                <div className="ml-2 pt-2 pb-1">
                    <p>
                        <b>Request: </b>
                    </p>
                    <div className="mt-2 ml-5">
                        <FieldView
                            name="Client Authorization"
                            value={props.request.clientAuthorization}
                        />
                        <FieldView name="Method" value={props.request.method} />
                        <FieldView
                            name="Headers"
                            value={props.request.headers}
                        />
                        <FieldView name="Body" value={props.request.body} />
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
                        <FieldView
                            name="Body"
                            value={props.request.response?.body}
                        />
                    </div>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}

function FieldView(props: {
    name: string;
    value: number | string | Record<string, Array<string>> | undefined;
}) {
    function ValueView() {
        if (props.value === undefined) return <i>none or unavailable</i>;
        if (
            typeof props.value === "string" ||
            typeof props.value === "number"
        ) {
            if (props.name == "Body") {
                return <p>{props.value}</p>;
            } else {
                return <>{props.value}</>;
            }
        }
        return (
            <>
                {Object.entries(props.value!).map((v) => (
                    <p key={v[0]}>
                        {v[0]}: {v[1].join(", ")}
                    </p>
                ))}
            </>
        );
    }
    return (
        <div className="mb-2">
            <b>{props.name}: </b>
            <ValueView />
        </div>
    );
}
