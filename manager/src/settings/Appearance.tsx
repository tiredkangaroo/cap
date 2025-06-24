import { AppearanceConfigContext, DarkModeContext } from "@/context/context";
import { CheckField } from "./SettingsFields";
import { useContext } from "react";

export function Appearance() {
    const [darkMode, setDarkMode] = useContext(DarkModeContext);
    const [appearanceConfig, setAppearanceConfig] = useContext(
        AppearanceConfigContext,
    );

    return (
        <div className="flex flex-col w-full gap-4">
            <div className="flex flex-col">
                <p className="text-gray-600">
                    Configure the appearance of the application.
                </p>
            </div>
            <div className="w-full flex flex-row">
                <div className="flex-1">Dark Mode</div>
                <div className="flex ml-auto">
                    <button
                        className="text-white w-24 border-r-1 border-white"
                        style={{
                            backgroundColor:
                                !appearanceConfig.darkModeSystem && !darkMode
                                    ? "#5383e6"
                                    : "#22355c",
                        }}
                        onClick={() => {
                            setAppearanceConfig({
                                ...appearanceConfig,
                                darkModeSystem: false,
                            });
                            setDarkMode(false);
                        }}
                    >
                        light
                    </button>
                    <button
                        className="text-white w-24 border-r-1 border-white"
                        style={{
                            backgroundColor:
                                !appearanceConfig.darkModeSystem && darkMode
                                    ? "#5383e6"
                                    : "#22355c",
                        }}
                        onClick={() => {
                            setAppearanceConfig({
                                ...appearanceConfig,
                                darkModeSystem: false,
                            });
                            setDarkMode(true);
                        }}
                    >
                        dark
                    </button>
                    <button
                        className="bg-gray-900 text-white w-24 border-r-1 border-white"
                        style={{
                            backgroundColor: appearanceConfig.darkModeSystem
                                ? "#5383e6"
                                : "#22355c",
                        }}
                        onClick={() => {
                            setAppearanceConfig({
                                ...appearanceConfig,
                                darkModeSystem: true,
                            });
                        }}
                    >
                        system
                    </button>
                </div>
            </div>
        </div>
    );
}
