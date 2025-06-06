import { Fragment } from "react/jsx-runtime";

export function Timeline(props: {
    times?: Record<string, number>;
    order?: Array<string>;
    className: string;
}) {
    const { times, order, className } = props;

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

    if (!times || !order || order.length === 0) {
        return (
            <div className={`w-full bg-black text-white p-2 ${className}`}>
                No data
            </div>
        );
    }

    const total = times["total"] ?? 1; // prevent divide by zero

    return (
        <div
            className={`flex flex-row h-8 w-full bg-black overflow-hidden rounded ${className}`}
        >
            {order.map((key, i) => {
                const time = times[key];
                const widthPercent = (time / total) * 100;
                console.log(times, key, time, nsToReadable(time));

                if (time === undefined) {
                    return <Fragment key={key}></Fragment>;
                }
                return (
                    <div
                        key={key}
                        style={{
                            width: `${widthPercent}%`,
                            backgroundColor: colors[i % colors.length],
                            minWidth: "50px", // Ensure visibility
                        }}
                        className="flex items-center justify-center overflow-hidden text-xs text-black font-semibold px-1"
                        title={`${key} - ${nsToReadable(time)}, ${widthPercent.toFixed(1)}%`}
                    >
                        <span className="whitespace-nowrap overflow-hidden text-ellipsis">
                            {key} ({nsToReadable(time)},{" "}
                            {widthPercent.toFixed(1)}%)
                        </span>
                    </div>
                );
            })}
        </div>
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
