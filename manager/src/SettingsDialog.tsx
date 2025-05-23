import { useEffect, useState } from "react";

import { Proxy } from "./api";
import { Config } from "./types";
import { IoClose } from "react-icons/io5";

export function SettingsDialog(props: {
    proxy: Proxy | null;
    open: boolean;
    setOpen: React.Dispatch<React.SetStateAction<boolean>>;
}) {
    const [proxyConfig, setProxyConfig] = useState<Config>(props.proxy!.config);

    //NOTE: async/await in set state functions
    useEffect(() => {
        const load = async () => {
            if (props.proxy === null) {
                return;
            }
            await props.proxy.getConfig();
            setProxyConfig(props.proxy!.config);
        };
        load();
    }, []);

    if (props.proxy === null) {
        return <></>;
    }
    return (
        <dialog
            className="fixed inset-0 flex z-50 w-[100%] h-[100%] items-center justify-center"
            open={props.open}
            hidden={!props.open}
        >
            <div className="w-[90%] h-[90%] flex bg-gray-100">
                {/* config cmps  */}
                <div className="w-full h-full ml-auto mr-auto mt-auto mb-auto flex flex-col items-center justify-center">
                    <div className="w-full flex flex-col">
                        <button
                            className="ml-auto mt-2 mr-2 bg-black aspect-square w-8"
                            onClick={() => props.setOpen(false)}
                        >
                            <IoClose className="text-white ml-auto mr-auto" />
                        </button>
                        <h1 className="text-3xl text-center mt-7">Settings</h1>
                        <p className="text-center">
                            configuation stuff goes here :/
                        </p>
                    </div>

                    <div className="flex flex-col min-w-fit w-[65%] h-full justify-center">
                        <div className="w-full flex flex-row text-xl mt-6">
                            <div className="flex-col w-[50%]">
                                <p className="text-xl mt-auto mb-auto">MITM</p>
                                <p className="text-sm">
                                    MITM set on indicates that the proxy is
                                    responsible for the TLS connection, and will
                                    use its own certificates to encrypt the
                                    traffic.
                                </p>
                            </div>
                            <input
                                type="checkbox"
                                className="mt-auto mb-auto ml-auto mr-2"
                                defaultChecked={proxyConfig.mitm}
                                onChange={(e) => {
                                    proxyConfig.mitm = e.target.checked;
                                    props.proxy!.setConfig(proxyConfig);
                                    const newObj = Object.assign(
                                        {},
                                        proxyConfig,
                                    );
                                    setProxyConfig(newObj);
                                }}
                            ></input>
                        </div>

                        <div className="w-full flex flex-row text-xl mt-6">
                            <div className="flex-col w-[50%]">
                                <p className="text-xl mt-auto mb-auto">
                                    Real IP Header
                                </p>
                                <p className="text-sm">
                                    Real IP Header set on means that the proxy
                                    will attach the IP address of the client in
                                    the X-Forwarded-For header set for the host.
                                </p>
                            </div>
                            <input
                                type="checkbox"
                                className="mt-auto mb-auto ml-auto mr-2"
                                defaultChecked={proxyConfig.real_ip_header}
                                onChange={(e) => {
                                    proxyConfig.real_ip_header =
                                        e.target.checked;
                                    props.proxy!.setConfig(proxyConfig);
                                    const newObj = Object.assign(
                                        {},
                                        proxyConfig,
                                    );
                                    setProxyConfig(newObj);
                                }}
                            ></input>
                        </div>
                        <div className="w-full flex flex-row text-xl mt-6">
                            <div className="flex-col w-[50%]">
                                <p className="text-xl mt-auto mb-auto">
                                    Certificate Lifetime
                                </p>
                                <p className="text-sm">
                                    The lifetime of the certificate in hours. It
                                    is possible to set it to less than 0 in
                                    which any new certificates generated will
                                    not be valid.
                                </p>
                            </div>
                            <input
                                defaultValue={proxyConfig.certificate_lifetime}
                                type="number"
                                className="ml-auto mr-2 min-w-fit text-md border-2 border-black text-center"
                                onBlur={(e) => {
                                    proxyConfig.certificate_lifetime = parseInt(
                                        e.target.value,
                                    );
                                    props.proxy!.setConfig(proxyConfig);
                                    const newObj = Object.assign(
                                        {},
                                        proxyConfig,
                                    );
                                    setProxyConfig(newObj);
                                }}
                            ></input>
                        </div>
                        <div className="w-full flex flex-row text-xl mt-6">
                            <div className="flex-col w-[50%]">
                                <p className="text-xl mt-auto mb-auto">
                                    Request Body Dumping
                                </p>
                                <p className="text-sm">
                                    Request body dumping allows inspection of
                                    the request body by the client.
                                </p>
                            </div>
                            <input
                                type="checkbox"
                                className="mt-auto mb-auto ml-auto mr-2"
                                defaultChecked={
                                    proxyConfig.provide_request_body
                                }
                                onChange={(e) => {
                                    proxyConfig.provide_request_body =
                                        e.target.checked;
                                    props.proxy!.setConfig(proxyConfig);
                                    const newObj = Object.assign(
                                        {},
                                        proxyConfig,
                                    );
                                    setProxyConfig(newObj);
                                }}
                            ></input>
                        </div>
                    </div>
                </div>
            </div>
        </dialog>
    );
}
