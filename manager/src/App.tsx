import { useState } from "react";
import "./App.css";
import { ProxyView } from "./ProxyView.tsx";

function App() {
    const [darkMode, setDarkMode] = useState<boolean>(
        window.matchMedia &&
            window.matchMedia("(prefers-color-scheme: dark)").matches,
    );
    const toggleDarkMode = () => {
        setDarkMode((prev) => !prev);
        document.documentElement.classList.toggle("dark", !darkMode);
    };
    window.matchMedia("(prefers-color-scheme: dark)").onchange = (e) => {
        if (e.matches) {
            setDarkMode(true);
            document.documentElement.classList.add("dark");
        } else {
            setDarkMode(false);
            document.documentElement.classList.remove("dark");
        }
    };

    return (
        <div className="flex flex-row w-full h-full">
            <ProxyView />
        </div>
    );
}

export default App;
