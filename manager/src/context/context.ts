import { AppearanceConfig, RequestContentProps } from "@/types";
import { createContext } from "react";

export const DarkModeContext = createContext<[boolean, (v: boolean) => void]>([
    false,
    () => {},
]);

export const AppearanceConfigContext = createContext<
    [AppearanceConfig, (v: AppearanceConfig) => void]
>([
    {
        darkModeSystem: true,
    },
    () => {},
]);

export const KeyboardContext = createContext<Set<string>>(new Set());

export const RequestDialogContentPropsContext = createContext<
    [
        RequestContentProps | undefined,
        React.Dispatch<React.SetStateAction<RequestContentProps | undefined>>,
    ]
>([undefined, () => {}]);
