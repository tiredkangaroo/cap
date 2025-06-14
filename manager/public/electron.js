import { app, BrowserWindow } from "electron";
import { exec, execFile } from "child_process";
import path from "path";

const PORT = 5173;
const CHECK_INTERVAL = 500; // ms

function checkPortListening(port) {
    return new Promise((resolve) => {
        exec(`lsof -i :${port} -sTCP:LISTEN`, (err, stdout) => {
            if (stdout && stdout.includes("LISTEN")) {
                resolve(true);
            } else {
                resolve(false);
            }
        });
    });
}

async function waitForPort(port, interval = 500) {
    return new Promise((resolve) => {
        const intervalId = setInterval(async () => {
            const isListening = await checkPortListening(port);
            if (isListening) {
                clearInterval(intervalId);
                resolve();
            }
        }, interval);
    });
}

async function createWindow() {
    const win = new BrowserWindow({
        width: 1250,
        height: 865,
        webPreferences: {
            nodeIntegration: true,
            contextIsolation: false,
        },
    });

    win.setMinimumSize(755, 865);
    win.setMaximumSize(10000, 2700); // only really setting a maximum height
    // win.on("will-resize", (ev, newBounds) => {
    //     const oldBounds = win.getBounds();
    //     if (newBounds.width < 300 || newBounds.height < 680) {
    //         ev.preventDefault();
    //         win.setBounds(oldBounds);
    //         return;
    //     }
    // });

    console.log(import.meta.url);

    if (process.env.DEBUG == "true") {
        await waitForPort(PORT, CHECK_INTERVAL);
        // this is the url for the vite server
        win.loadURL("http://localhost:5173");
    } else {
        // this is the url for the production build
        // app/Content/Resources/app.asar/../proxy-app
        // app/Content/Resources/app.asar/../dist/index.html

        // add better error handling; kill all procs on err
        execFile(
            "bash",
            ["-c", new URL("../../proxy-app", import.meta.url).pathname],
            (err, stdout, stderr) => {
                console.log("err", err);
                console.log("stdout", stdout);
                console.error("stderr", stderr);
            },
        );
        win.loadFile("../dist/index.html");
    }
}

// NOTE: why does this work but not await app.whenReady()?
app.whenReady().then(async () => {
    await createWindow();
});

app.on("window-all-closed", () => {
    if (process.platform !== "darwin") {
        app.quit();
    }
});
