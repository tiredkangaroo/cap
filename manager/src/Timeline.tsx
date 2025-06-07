import { Timing, MajorTime } from "./timing";
import {
    HoverCard,
    HoverCardContent,
    HoverCardTrigger,
} from "./components/ui/hover-card";

const colors = [
    "#fa5b50",
    "#acf797",
    "#9ceeff",
    "#b39ff5",
    "#ffb8f1",
    "#7a485f",
    "#48507a",
    "#487a4f",
    "#7a7848",
    "#7d7d7a",
];

export function Timeline(props: {
    totalTime?: number;
    timing?: Timing;
    className?: string;
}) {
    if (!props.timing || !props.totalTime) {
        return (
            <div
                className={`w-full bg-black text-white p-2 ${props.className}`}
            >
                No data
            </div>
        );
    }

    return (
        <div
            className={`flex flex-row w-[calc(100%-1rem)] bg-black overflow-hidden rounded ${props.className}`}
        >
            {props.timing.majorTimeKeys.map((majorKey, i) => {
                const majorValue = props.timing!.majorTimeValues[i];
                const widthPercent =
                    (majorValue.duration / props.totalTime!) * 100;
                return (
                    <TimeView
                        idx={i}
                        majorKey={majorKey}
                        majorValue={majorValue}
                        widthPercent={widthPercent}
                    />
                );
            })}
        </div>
    );
}

function TimeView(props: {
    idx: number;
    majorKey: string;
    majorValue: MajorTime;
    widthPercent: number;
}) {
    return (
        <HoverCard>
            <HoverCardTrigger
                className="w-full flex flex-row wrap-break-word h-8 items-center justify-center overflow-hidden text-xs text-black font-semibold px-1"
                style={{
                    width: `${props.widthPercent}%`,
                    backgroundColor: colors[props.idx % colors.length],
                    minWidth: "25px", // Ensure visibility
                }}
            >
                <span className="whitespace-nowrap overflow-hidden text-ellipsis">
                    {props.majorKey} ({nsToReadable(props.majorValue.duration)},{" "}
                    {props.widthPercent.toFixed(1)}%)
                </span>
                {/* </div> */}
            </HoverCardTrigger>
            <HoverCardContent>
                <h1 className="text-md font-semibold">
                    {props.majorKey} ({props.widthPercent.toFixed(2)}%){" "}
                </h1>
                {props.majorValue.minorTimeKeys ? (
                    <div className="flex flex-col mt-2">
                        {props.majorValue.minorTimeKeys.map((minorKey, i) => {
                            const minorValue =
                                props.majorValue.minorTimeValues![i];
                            return (
                                <div
                                    key={i}
                                    className="flex flex-row items-center justify-between"
                                >
                                    <span className="text-sm">{minorKey}:</span>
                                    <span className="text-sm">
                                        {nsToReadable(minorValue.duration)}
                                    </span>
                                </div>
                            );
                        })}
                    </div>
                ) : (
                    <></>
                )}
            </HoverCardContent>
        </HoverCard>
    );
}

function nsToReadable(ns: number): string {
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
