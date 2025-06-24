import { useEffect, useState } from "react";
import "./App.css";
import { ProxyView } from "./ProxyView.tsx";
import { DarkModeContext } from "./context/contextDarkMode.ts";

function App() {
    const [darkMode, setDarkMode] = useState<boolean>(
        window.matchMedia &&
            window.matchMedia("(prefers-color-scheme: dark)").matches,
    );
    useEffect(() => {
        document.documentElement.classList.toggle("dark", darkMode);
    }, [darkMode]);
    window.matchMedia("(prefers-color-scheme: dark)").onchange = (e) => {
        if (e.matches) {
            setDarkMode(true);
        } else {
            setDarkMode(false);
        }
    };

    return (
        <div className="flex flex-row w-full h-full">
            <DarkModeContext value={[darkMode, (v) => setDarkMode(v)]}>
                <ProxyView />
            </DarkModeContext>
        </div>
    );
}

export default App;
