import { useContext } from "react";
import { RequestDialogContentPropsContext } from "./context/context";
import { RequestViewContent } from "./RequestView";
import { IoMdArrowRoundBack } from "react-icons/io";

export function RequestDialog() {
    const [contentProps, setContentProps] = useContext(
        RequestDialogContentPropsContext,
    );
    console.log(contentProps);
    if (!contentProps) {
        // if contentProps is undefined, we don't render the dialog
        return null;
    }
    return (
        // NOTE: those might not be necessary, but they are here bc sanity
        <dialog
            open={contentProps !== undefined}
            hidden={contentProps === undefined}
            className="w-full h-full z-50 pt-2 bg-gray-300 dark:bg-gray-900 overflow-y-auto p-4 space-y-4 font-[monospace] text-black dark:text-white"
        >
            <button
                className="flex flew-row gap-3 items-center rounded-sm px-3 py-1 bg-black text-white"
                onClick={() => setContentProps(undefined)}
            >
                <IoMdArrowRoundBack />
                back
            </button>
            <RequestViewContent {...contentProps!} />
        </dialog>
    );
}
