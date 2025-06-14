import { Timing, MajorTime } from "./timing";
import {
    HoverCard,
    HoverCardContent,
    HoverCardTrigger,
} from "./components/ui/hover-card";
import { nsToReadable } from "./utils";

const colors = [
    "#ffc1ab",
    "#adc7ff",
    "#ffadf1",
    "#fa3c5f",
    "#e6d7f5",
    "#d7d8de",
    "#defc97",
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

    let percents = 0;
    return (
        <div
            className={`flex flex-row w-[calc(100%-1rem)] bg-black overflow-hidden rounded ${props.className}`}
        >
            {props.timing.majorTimeKeys.map((majorKey, i) => {
                const majorValue = props.timing!.majorTimeValues[i];
                const widthPercent =
                    (majorValue.duration / props.totalTime!) * 100;
                percents += widthPercent;
                return (
                    <TimeView
                        idx={i}
                        majorKey={majorKey}
                        majorValue={majorValue}
                        widthPercent={widthPercent}
                    />
                );
            })}
            <TimeView
                idx={props.timing.majorTimeKeys.length}
                majorKey="Miscellaneous"
                majorValue={{
                    duration: ((100 - percents) / 100) * props.totalTime!,
                    minorTimeKeys: [],
                    minorTimeValues: [],
                }}
                widthPercent={100 - percents}
            />
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
                <p className="text-gray-500">
                    {nsToReadable(props.majorValue.duration)}
                </p>
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
