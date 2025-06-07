export interface MinorTime {
    duration: number; // duration in ns (time.duration)
}

export interface MajorTime {
    duration: number; // duration in ns (time.duration)
    minorTimeKeys: Array<string> | null; // keys for minor times (time.minorTimeKeys)
    minorTimeValues: Array<MinorTime> | null;
}

export interface Timing {
    majorTimeKeys: Array<string>;
    majorTimeValues: Array<MajorTime>;
}
