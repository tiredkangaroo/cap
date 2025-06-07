export function camelCaseToCapitalSpace(str: string): string {
    return str
        .replace(/([a-z])([A-Z])/g, "$1 $2") // insert space before capital letters
        .replace(/^./, (char) => char.toUpperCase()); // capitalize the first letter
}
