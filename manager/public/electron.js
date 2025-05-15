import { app, BrowserWindow } from "electron";
import { exec } from "child_process";

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
        height: 750,
        webPreferences: {
            nodeIntegration: true,
            contextIsolation: false,
        },
    });

    if (process.env.DEBUG == "true") {
        await waitForPort(PORT, CHECK_INTERVAL);
        // this is the url for the vite server
        win.loadURL("http://localhost:5173");
    } else {
        // this is the url for the production build
        win.loadFile("index.html");
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
