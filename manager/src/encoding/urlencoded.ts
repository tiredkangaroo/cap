export function parseURLEncoded(input: string): Record<string, string[]> {
    const result: Record<string, string[]> = {};

    // Split input by '&' to get key=value pairs
    const pairs = input.split("&");

    for (const pair of pairs) {
        if (!pair) continue;

        const [rawKey, rawValue = ""] = pair.split("=");
        const key = decodeURIComponent(rawKey.replace(/\+/g, " "));
        const value = decodeURIComponent(rawValue.replace(/\+/g, " "));

        // If key already exists, push to its array
        if (Object.prototype.hasOwnProperty.call(result, key)) {
            result[key].push(value);
        } else {
            result[key] = [value];
        }
    }

    return result;
}

export function encodeURLEncoded(data: Record<string, string[]>): string {
    const pairs: string[] = [];

    for (const key in data) {
        if (Object.prototype.hasOwnProperty.call(data, key)) {
            const values = data[key];
            for (const value of values) {
                pairs.push(
                    `${encodeURIComponent(key)}=${encodeURIComponent(value)}`,
                );
            }
        }
    }

    return pairs.join("&");
}
