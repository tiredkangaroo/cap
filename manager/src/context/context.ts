import { AppearanceConfig } from "@/types";
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
