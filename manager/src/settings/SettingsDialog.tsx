import { Proxy } from "../api/api";
import { IoClose } from "react-icons/io5";
import { ProxySettingsView } from "./ProxySettings";
import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
} from "../components/ui/tabs";
import { RequestsViewConfig } from "../types";
import { ViewSettings } from "./ViewSettings";
import { Appearance } from "./Appearance";

export function SettingsDialog(props: {
    proxy: Proxy | null;
    open: boolean;
    setOpen: React.Dispatch<React.SetStateAction<boolean>>;
    requestsViewConfig: RequestsViewConfig;
    setRequestsViewConfig: React.Dispatch<
        React.SetStateAction<RequestsViewConfig>
    >;
}) {
    if (props.proxy === null) return <></>;

    return (
        <dialog
            open={props.open}
            hidden={!props.open}
            className="fixed inset-0 z-50 flex items-center justify-center w-full h-full bg-black/40 backdrop-blur-sm"
        >
            <div className="w-[90%] h-[90%] bg-gray-100 rounded-4xl shadow-lg flex flex-col overflow-hidden p-2">
                {/* Entire inner content in vertical flex layout */}
                <div className="flex flex-col flex-grow min-h-0 px-2 py-2">
                    <button
                        className="ml-auto bg-black px-1 py-1 aspect-square rounded-4xl"
                        onClick={() => props.setOpen(false)}
                    >
                        <IoClose className="text-white mx-auto" />
                    </button>

                    <h1 className="text-3xl text-center">Settings</h1>
                    <p className="text-center mb-4">
                        configure some cool stuff here :)
                    </p>

                    {/* Flex-grow Tabs with min-h-0 is critical */}
                    <Tabs
                        defaultValue="appearance"
                        className="flex flex-col flex-grow min-h-0"
                    >
                        <TabsList className="bg-gray-200 dark:bg-gray-800 text-black dark:text-white rounded-xl p-1 mb-4">
                            <TabsTrigger value="appearance">
                                Appearance
                            </TabsTrigger>
                            <TabsTrigger value="proxy">Proxy</TabsTrigger>
                            <TabsTrigger value="requests-view">
                                Request View
                            </TabsTrigger>
                        </TabsList>

                        <TabsContent
                            value="appearance"
                            className="flex-grow min-h-0 overflow-y-auto pr-1"
                        >
                            <Appearance />
                        </TabsContent>
                        <TabsContent
                            value="proxy"
                            className="flex-grow min-h-0 overflow-y-auto pr-1"
                        >
                            <ProxySettingsView proxy={props.proxy} />
                        </TabsContent>

                        <TabsContent
                            className="flex-grow min-h-0 overflow-y-auto pr-1"
                            value="requests-view"
                        >
                            <ViewSettings
                                requestsViewConfig={props.requestsViewConfig}
                                setRequestsViewConfig={
                                    props.setRequestsViewConfig
                                }
                            />
                        </TabsContent>
                    </Tabs>
                </div>
            </div>
        </dialog>
    );
}
