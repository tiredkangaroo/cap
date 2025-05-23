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
                    <p className="flex-1 ml-1 text-md">{props.request.id}</p>
                    <p className="flex-1 ml-1 text-center">
                        {props.request.host}
                    </p>
                    <p className="flex-1 ml-1 mr-1">{props.request.clientIP}</p>
                </div>
            </CollapsibleTrigger>
            <CollapsibleContent className="bg-gray-300">
                <div className="ml-2 pt-2 pb-1">
                    <p className="mb-2">
                        <b>Client Authorization: </b>
                        {props.request.clientAuthorization != "" ? (
                            props.request.clientAuthorization
                        ) : (
                            <i>No Client Authorization</i>
                        )}
                    </p>
                    {props.request.method ? (
                        <p className="mb-2">
                            <b>Method: </b> {props.request.method}
                        </p>
                    ) : (
                        <p>
                            <b>Method: </b> unavailable
                        </p>
                    )}
                    {props.request.headers ? (
                        <div className="mb-2">
                            <b>Headers:</b>
                            {Object.entries(props.request.headers!).map((v) => (
                                <p key={v[0]}>
                                    {v[0]}: {v[1].join(", ")}
                                </p>
                            ))}
                        </div>
                    ) : (
                        <p>
                            Headers: <i>none or unavailable</i>
                        </p>
                    )}
                    {props.request.body ? (
                        <div className="mb-2">
                            <p>
                                <b>Body: </b>
                            </p>
                            {props.request.body}
                        </div>
                    ) : (
                        <p className="mb-2">
                            Body: <i>none or unavailable</i>
                        </p>
                    )}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
