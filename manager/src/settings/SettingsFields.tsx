export function CheckField(props: {
    name: string;
    defaultChecked: boolean;
    onChange(v: boolean): void;
    children: string;
}) {
    return (
        <div className="flex justify-between items-start mt-4 gap-4">
            <div className="flex flex-col">
                <label className="font-semibold">{props.name}</label>
                <p className="text-sm text-gray-600">{props.children}</p>
            </div>
            <input
                type="checkbox"
                className="mt-2 accent-black w-5 h-5"
                checked={props.defaultChecked}
                onChange={(e) => props.onChange(e.target.checked)}
            />
        </div>
    );
}

export function InputField(props: {
    name: string;
    defaultValue: string | number;
    type: string;
    onChange(v: string): void;
    children: string;
}) {
    return (
        <div className="flex justify-between items-start mt-4 gap-4">
            <div className="flex flex-col">
                <label className="font-semibold">{props.name}</label>
                <p className="text-sm text-gray-600">{props.children}</p>
            </div>
            <input
                type={props.type}
                defaultValue={props.defaultValue}
                className="border border-gray-400 rounded py-1 w-24 ml-auto mr-2 min-w-fit text-md text-center"
                onBlur={(e) => props.onChange(e.target.value)}
            />
        </div>
    );
}
