import { DarkModeContext } from "@/context/contextDarkMode";
import { CheckField } from "./SettingsFields";
import { useContext } from "react";

export function Appearance() {
    const [darkMode, setDarkMode] = useContext(DarkModeContext);
    return (
        <div className="flex flex-col gap-4">
            <div className="flex flex-col">
                <p className="text-gray-600">
                    Configure the appearance of the application.
                </p>
            </div>
            <div className="flex flex-col gap-2">
                <CheckField
                    name="Dark Mode"
                    defaultChecked={darkMode}
                    onChange={(v) => setDarkMode(v)}
                >
                    Enable dark mode for a more comfortable viewing experience
                    in low-light environments.
                </CheckField>
            </div>
        </div>
    );
}
