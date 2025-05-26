import { CheckField } from "./SettingsFields";
import { RequestsViewConfig } from "../types";

export function ViewSettings(props: {
    requestsViewConfig: RequestsViewConfig;
    setRequestsViewConfig: React.Dispatch<
        React.SetStateAction<RequestsViewConfig>
    >;
}) {
    return (
        <div className="flex flex-col w-full space-y-4 p-2">
            <CheckField
                name="Hide Errors"
                defaultChecked={props.requestsViewConfig.hideError}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideError: c,
                    }));
                }}
            >
                Hide the error field.
            </CheckField>
            <CheckField
                name="Hide ID"
                defaultChecked={props.requestsViewConfig.hideID}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideID: c,
                    }));
                }}
            >
                Hide the ID field.
            </CheckField>
            <CheckField
                name="Hide Client User"
                defaultChecked={props.requestsViewConfig.hideClientUser}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideClientUser: c,
                    }));
                }}
            >
                Hide the client user field.
            </CheckField>
            <CheckField
                name="Hide Client Password"
                defaultChecked={props.requestsViewConfig.hideClientPassword}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideClientPassword: c,
                    }));
                }}
            >
                Hide the client password field.
            </CheckField>
            <CheckField
                name="Hide Method"
                defaultChecked={props.requestsViewConfig.hideMethod}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideMethod: c,
                    }));
                }}
            >
                Hide the method field.
            </CheckField>
            <CheckField
                name="Hide Path"
                defaultChecked={props.requestsViewConfig.hidePath}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hidePath: c,
                    }));
                }}
            >
                Hide the path field.
            </CheckField>
            <CheckField
                name="Hide Query"
                defaultChecked={props.requestsViewConfig.hideQuery}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideQuery: c,
                    }));
                }}
            >
                Hide the query field.
            </CheckField>
            <CheckField
                name="Hide Request Headers"
                defaultChecked={props.requestsViewConfig.hideRequestHeaders}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideRequestHeaders: c,
                    }));
                }}
            >
                Hide the request headers field.
            </CheckField>
            <CheckField
                name="Hide Request Body"
                defaultChecked={props.requestsViewConfig.hideRequestBody}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideRequestBody: c,
                    }));
                }}
            >
                Hide the request body field.
            </CheckField>
            <CheckField
                name="Hide Response Status"
                defaultChecked={props.requestsViewConfig.hideResponseStatus}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideResponseStatus: c,
                    }));
                }}
            >
                Hide the response status field.
            </CheckField>
            <CheckField
                name="Hide Response Headers"
                defaultChecked={props.requestsViewConfig.hideResponseHeaders}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideResponseHeaders: c,
                    }));
                }}
            >
                Hide the response headers field.
            </CheckField>
            <CheckField
                name="Hide Response Body"
                defaultChecked={props.requestsViewConfig.hideResponseBody}
                onChange={(c: boolean) => {
                    props.setRequestsViewConfig((prev) => ({
                        ...prev,
                        hideResponseBody: c,
                    }));
                }}
            >
                Hide the response body field.
            </CheckField>
        </div>
    );
}
