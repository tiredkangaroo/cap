async function loadIndexMD() {
    const markdownTextResp = await fetch(`index.md`, {
        cache: "no-cache",
    });
    const markdownText = await markdownTextResp.text();
    const htmlContent = marked.parse(markdownText);
    document.getElementById("content").innerHTML = htmlContent;
}

window.onload = async () => {
    await loadIndexMD();
    hljs.highlightAll();
    document.querySelectorAll("ul").forEach((ul) => {
        const lis = ul.querySelectorAll("li");
        const hasP = Array.from(lis).some((li) => li.querySelector("p"));
        if (!hasP) {
            ul.classList.add("no-li-p");
        }
    });
};
