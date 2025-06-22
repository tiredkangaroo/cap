import { app, BrowserWindow } from "electron";
import { exec, execFile, spawn } from "child_process";
import path from "path";
import { stdin } from "process";

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

    console.log(import.meta.url);

    if (process.env.DEBUG == "true") {
        await waitForPort(PORT, CHECK_INTERVAL);
        // this is the url for the vite server
        win.loadURL("http://localhost:5173");
    } else if (process.env.BUILT != "false") {
        // add better error handling; kill all procs on err
        const script = new URL("../../proxy-app", import.meta.url).pathname;
        const child = spawn("bash", ["-c", script], {
            stdio: ["inherit", "pipe", "pipe"],
        });
        child.stdout.pipe(process.stdout);
        child.stderr.pipe(process.stderr);

        // execFile(
        //     "bash",
        //     ["-c", scri],
        //     (err, stdout, stderr) => {
        //         console.log("err", err);
        //         console.log("stdout", stdout);
        //         console.error("stderr", stderr);
        //     },
        // );
        win.loadFile("../dist/index.html");
    } else {
        win.loadFile("dist/index.html");
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
