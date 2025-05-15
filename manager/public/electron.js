import { app, BrowserWindow } from "electron";

function createWindow() {
    const win = new BrowserWindow({
        width: 1250,
        height: 750,
        webPreferences: {
            nodeIntegration: true,
            contextIsolation: false,
        },
    });

    // this is the url for the vite server
    win.loadURL("http://localhost:5173");
}

// NOTE: why does this work but not await app.whenReady()?
app.whenReady().then(() => {
    createWindow();
});

app.on("window-all-closed", () => {
    if (process.platform !== "darwin") {
        app.quit();
    }
});
