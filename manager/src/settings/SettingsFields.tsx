export function CheckField(props: {
    name: string;
    defaultChecked: boolean;
    onChange(v: boolean): void;
    children: string;
}) {
    return (
        <div className="w-full flex flex-row text-xl mt-6">
            <div className="flex-col w-[50%]">
                <p className="text-xl mt-auto mb-auto">{props.name}</p>
                <p className="text-sm">{props.children}</p>
            </div>
            <input
                type="checkbox"
                className="mt-auto mb-auto ml-auto mr-2"
                defaultChecked={props.defaultChecked}
                onChange={(e) => {
                    props.onChange(e.target.checked);
                }}
            ></input>
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
        <div className="w-full flex flex-row text-xl mt-6">
            <div className="flex-col w-[50%]">
                <p className="text-xl mt-auto mb-auto">{props.name}</p>
                <p className="text-sm">{props.children}</p>
            </div>
            <input
                defaultValue={props.defaultValue}
                type="number"
                className="ml-auto mr-2 min-w-fit text-md border-2 border-black text-center"
                onBlur={(e) => props.onChange(e.target.value)}
            ></input>
        </div>
    );
}
