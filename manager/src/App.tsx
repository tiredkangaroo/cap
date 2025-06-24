import { useEffect, useState } from "react";
import "./App.css";
import { ProxyView } from "./ProxyView.tsx";
import { AppearanceConfigContext, DarkModeContext } from "./context/context.ts";
import { AppearanceConfig } from "./types.ts";

function App() {
    // appearance config:
    const appearanceConfigLSItem = localStorage.getItem("appearanceConfig");
    const [appearanceConfig, setAppearanceConfig] = useState<AppearanceConfig>(
        JSON.parse(appearanceConfigLSItem!) || {
            darkModeSystem: true,
        }, // json.parse returns null if input is null, so this is safe
    );
    useEffect(() => {
        localStorage.setItem(
            "appearanceConfig",
            JSON.stringify(appearanceConfig),
        );
        if (appearanceConfig.darkModeSystem) {
            // if dark mode is derived from system, set dark mode to true if system is in dark mode
            const isDarkMode =
                window.matchMedia &&
                window.matchMedia("(prefers-color-scheme: dark)").matches;
            document.documentElement.classList.toggle("dark", isDarkMode);
        }
    }, [appearanceConfig]);

    // dark mode:
    let darkModeInitial = false;
    if (!appearanceConfig.darkModeSystem) {
        darkModeInitial = localStorage.getItem("dark_mode") === "true";
    } else {
        darkModeInitial =
            window.matchMedia &&
            window.matchMedia("(prefers-color-scheme: dark)").matches;
    }
    const [darkMode, setDarkMode] = useState<boolean>(darkModeInitial);

    useEffect(() => {
        document.documentElement.classList.toggle("dark", darkMode);
        localStorage.setItem("dark_mode", darkMode ? "true" : "false");
    }, [darkMode]);
    window.matchMedia("(prefers-color-scheme: dark)").onchange = (e) => {
        console.log(appearanceConfig.darkModeSystem);
        if (!appearanceConfig.darkModeSystem) return; // do nothing if dark mode derived from system is disabled
        if (e.matches) {
            setDarkMode(true);
        } else {
            setDarkMode(false);
        }
    };

    return (
        <div className="flex flex-row w-full h-full">
            <DarkModeContext value={[darkMode, (v) => setDarkMode(v)]}>
                <AppearanceConfigContext
                    value={[appearanceConfig, (v) => setAppearanceConfig(v)]}
                >
                    <ProxyView />
                </AppearanceConfigContext>
            </DarkModeContext>
        </div>
    );
}

export default App;
