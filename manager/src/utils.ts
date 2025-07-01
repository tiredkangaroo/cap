export function camelCaseToCapitalSpace(str: string): string {
    return str
        .replace(/([a-z])([A-Z])/g, "$1 $2") // insert space before capital letters
        .replace(/^./, (char) => char.toUpperCase()); // capitalize the first letter
}

export function pascalCaseToCapitalSpace(str: string): string {
    return str
        .replace(/([a-z])([A-Z])/g, "$1 $2")
        .replace(/([A-Z])([A-Z][a-z])/g, "$1 $2");
}

export function nsToReadable(ns: number): string {
    if (ns < 1000) {
        return `${ns.toFixed(2)} ns`;
    }

    const milliseconds = ns / 1_000_000;
    if (milliseconds < 1000) {
        return `${milliseconds.toFixed(2)} ms`;
    }

    const seconds = milliseconds / 1000;
    if (seconds <= 60) {
        return `${seconds.toFixed(0)} sec`;
    }

    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.round(seconds % 60);
    return `${minutes} min ${remainingSeconds} sec`;
}

export function objectToQueryString(
    obj: Record<string, string | number | boolean | undefined>,
): string {
    return Object.entries(obj)
        .map(
            ([key, value]) =>
                `${encodeURIComponent(key)}=${encodeURIComponent(value || "")}`,
        )
        .join("&");
}

export function equalArray<T extends string | number>(
    a: Array<T>,
    b: Array<T>,
) {
    if (a === b) return true; // references are the same
    if (a == null || b == null) return false; // one of them is null or undefined
    if (a.length !== b.length) return false; // different lengths

    a.every((value, idx) => value === b[idx]);
    return true;
}

export function cmdOrCtrlPressed(keyboard: Set<string>) {
    // sybau "'platform' is deprecated" looking ahh 🥀
    switch (navigator.platform) {
        // mac platforms will use the command key
        case "MacIntel":
        case "MacPPC":
        case "Mac68K":
            return keyboard.has("MetaLeft") || keyboard.has("MetaRight");
        default: // others platform will use the control key
            return keyboard.has("Control");
    }
}
