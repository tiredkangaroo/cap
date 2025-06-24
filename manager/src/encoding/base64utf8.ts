// vibe-coded this function.
export function isBase64Utf8(str: string): boolean {
    if (str.length % 4 !== 0) return false;
    const base64Map = new Uint8Array(128).fill(255); // invalid marker
    const chars =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    for (let i = 0; i < chars.length; i++) base64Map[chars.charCodeAt(i)] = i;
    base64Map["=".charCodeAt(0)] = 0;

    let i = 0;
    const utf8Buf: number[] = [];

    function validateUtf8Chunk(): boolean {
        let j = 0;
        const bytes = utf8Buf;
        while (j < bytes.length) {
            const byte1 = bytes[j];
            if (byte1 <= 0x7f) {
                j++;
            } else if ((byte1 & 0xe0) === 0xc0) {
                if (j + 1 >= bytes.length) return false;
                const b2 = bytes[j + 1];
                const cp = ((byte1 & 0x1f) << 6) | (b2 & 0x3f);
                if ((b2 & 0xc0) !== 0x80 || cp < 0x80) return false;
                j += 2;
            } else if ((byte1 & 0xf0) === 0xe0) {
                if (j + 2 >= bytes.length) return false;
                const b2 = bytes[j + 1],
                    b3 = bytes[j + 2];
                const cp =
                    ((byte1 & 0x0f) << 12) | ((b2 & 0x3f) << 6) | (b3 & 0x3f);
                if ((b2 & 0xc0) !== 0x80 || (b3 & 0xc0) !== 0x80 || cp < 0x800)
                    return false;
                j += 3;
            } else if ((byte1 & 0xf8) === 0xf0) {
                if (j + 3 >= bytes.length) return false;
                const b2 = bytes[j + 1],
                    b3 = bytes[j + 2],
                    b4 = bytes[j + 3];
                const cp =
                    ((byte1 & 0x07) << 18) |
                    ((b2 & 0x3f) << 12) |
                    ((b3 & 0x3f) << 6) |
                    (b4 & 0x3f);
                if (
                    (b2 & 0xc0) !== 0x80 ||
                    (b3 & 0xc0) !== 0x80 ||
                    (b4 & 0xc0) !== 0x80 ||
                    cp < 0x10000 ||
                    cp > 0x10ffff
                )
                    return false;
                j += 4;
            } else {
                return false;
            }
        }
        bytes.length = 0; // clear buffer
        return true;
    }

    while (i < str.length) {
        let sextet = 0;
        for (let j = 0; j < 4; j++) {
            const charCode = str.charCodeAt(i++);
            if (charCode > 127) return false;
            const val = base64Map[charCode];
            if (val === 255) return false;
            sextet = (sextet << 6) | val;
        }

        // Extract up to 3 bytes
        utf8Buf.push((sextet >> 16) & 0xff);
        utf8Buf.push((sextet >> 8) & 0xff);
        utf8Buf.push(sextet & 0xff);

        // If padding was present, remove corresponding bytes
        if (str[i - 1] === "=") {
            utf8Buf.pop();
            if (str[i - 2] === "=") utf8Buf.pop();
        }

        if (!validateUtf8Chunk()) return false;
    }

    return true;
}
