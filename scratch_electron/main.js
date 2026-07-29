const { app, BrowserWindow } = require('electron');

const CAFE_NAME = process.env.CAFE_NAME || 'KafePlat';
const CAFE_WEBSITE = process.env.CAFE_WEBSITE || 'localhost:5173';

function createWindow () {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    autoHideMenuBar: true,
    title: CAFE_NAME,
    webPreferences: {
      nodeIntegration: false
    }
  });

  win.loadURL(`https://${CAFE_WEBSITE}`);
}

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});
