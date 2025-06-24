import { createContext } from "react";

export const DarkModeContext = createContext<[boolean, (v: boolean) => void]>([
    false,
    () => {},
]);
