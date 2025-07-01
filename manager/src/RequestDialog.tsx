export function RequestDialog(props: {
    F: React.ElementType | undefined;
    setF: React.Dispatch<React.SetStateAction<React.ElementType | undefined>>;
}) {
    if (props.F === undefined) {
        return <></>;
    }
    return (
        // NOTE: those might not be necessary, but they are here bc sanity
        <dialog open={props.F !== undefined} hidden={props.F === undefined}>
            <div>
                <props.F />
            </div>
        </dialog>
    );
}
