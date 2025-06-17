export function camelCaseToCapitalSpace(str: string): string {
    return str
        .replace(/([a-z])([A-Z])/g, "$1 $2") // insert space before capital letters
        .replace(/^./, (char) => char.toUpperCase()); // capitalize the first letter
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
