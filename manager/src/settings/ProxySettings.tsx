import { useEffect, useState } from "react";
import { Proxy } from "@/api/api";
import { CheckField, InputField } from "./SettingsFields";
import { Config } from "@/types";

export function ProxySettingsView(props: { proxy: Proxy }) {
    const [proxyConfig, setProxyConfig] = useState<Config>(props.proxy!.config);

    useEffect(() => {
        const load = async () => {
            await props.proxy.getConfig();
            setProxyConfig(props.proxy!.config);
        };
        load();
    }, [props.proxy]);

    return (
        <div className="flex flex-col w-full space-y-4 p-2">
            <CheckField
                name="MITM"
                defaultChecked={proxyConfig.mitm}
                onChange={(v: boolean) => {
                    proxyConfig.mitm = v;
                    props.proxy!.setConfig(proxyConfig);
                    setProxyConfig({ ...proxyConfig });
                }}
            >
                MITM set on indicates that the proxy is responsible for the TLS
                connection, and will use its own certificates to encrypt the
                traffic.
            </CheckField>

            <CheckField
                name="Real IP Header"
                defaultChecked={proxyConfig.real_ip_header}
                onChange={(v: boolean) => {
                    proxyConfig.real_ip_header = v;
                    props.proxy!.setConfig(proxyConfig);
                    setProxyConfig({ ...proxyConfig });
                }}
            >
                Real IP Header set on means that the proxy will attach the IP
                address of the client in the X-Forwarded-For header set for the
                host.
            </CheckField>

            <InputField
                name="Certificate Lifetime"
                defaultValue={proxyConfig.certificate_lifetime}
                type="number"
                onChange={(v: string) => {
                    proxyConfig.certificate_lifetime = parseInt(v);
                    props.proxy!.setConfig(proxyConfig);
                    setProxyConfig({ ...proxyConfig });
                }}
            >
                The lifetime of the certificate in hours. It is possible to set
                it to less than 0 in which any new certificates generated will
                not be valid.
            </InputField>

            <InputField
                name="Perform Delay"
                defaultValue={proxyConfig.perform_delay}
                type="number"
                onChange={(v: string) => {
                    proxyConfig.perform_delay = parseInt(v);
                    props.proxy!.setConfig(proxyConfig);
                    setProxyConfig({ ...proxyConfig });
                }}
            >
                The delay in milliseconds that the proxy should wait before
                performing to a request.
            </InputField>

            <CheckField
                name="Request Body Dumping"
                defaultChecked={proxyConfig.provide_request_body}
                onChange={(v: boolean) => {
                    proxyConfig.provide_request_body = v;
                    props.proxy!.setConfig(proxyConfig);
                    setProxyConfig({ ...proxyConfig });
                }}
            >
                Request body dumping allows inspection of the request body by
                the client. It is resource intensive, especially with larger
                bodies.
            </CheckField>

            <CheckField
                name="Response Body Dumping"
                defaultChecked={proxyConfig.provide_response_body}
                onChange={(v: boolean) => {
                    proxyConfig.provide_response_body = v;
                    props.proxy!.setConfig(proxyConfig);
                    setProxyConfig({ ...proxyConfig });
                }}
            >
                Response body dumping allows inspection of the response body by
                the client. It is resource intensive, especially with larger
                bodies.
            </CheckField>

            <CheckField
                name="Require Approval"
                defaultChecked={proxyConfig.require_approval}
                onChange={(v: boolean) => {
                    proxyConfig.require_approval = v;
                    props.proxy!.setConfig(proxyConfig);
                    setProxyConfig({ ...proxyConfig });
                }}
            >
                Require approval makes the proxy wait for the client to approve
                the request before proceeding. This is useful for debugging and
                security purposes.
            </CheckField>

            <CheckField
                name="Client Process Info"
                defaultChecked={proxyConfig.get_client_process_info}
                onChange={(v: boolean) => {
                    proxyConfig.get_client_process_info = v;
                    props.proxy!.setConfig(proxyConfig);
                    setProxyConfig({ ...proxyConfig });
                }}
            >
                Client process info allows the proxy to provide information
                about the client process. This information can prove to be
                useful, however it takes a significant amount of time to gather.
            </CheckField>

            <CheckField
                name="Timeline-based State Updates"
                defaultChecked={proxyConfig.timeline_based_state_updates}
                onChange={(v: boolean) => {
                    proxyConfig.timeline_based_state_updates = v;
                    props.proxy!.setConfig(proxyConfig);
                    setProxyConfig({ ...proxyConfig });
                }}
            >
                Timeline-based state updates allows the proxy to set the state
                of the request based on the timeline of the request.
            </CheckField>
        </div>
    );
}
