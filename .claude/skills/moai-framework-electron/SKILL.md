---
name: "moai-framework-electron"
description: "Electron 33+ desktop app development specialist covering Main/Renderer process architecture, IPC communication, auto-update, packaging with Electron Forge and electron-builder, and security best practices. Use when building cross-platform desktop applications, implementing native OS integrations, or packaging Electron apps for distribution."
version: 1.0.0
category: "framework"
modularized: false
user-invocable: false
tags: ["electron", "desktop", "cross-platform", "nodejs", "chromium"]
context7-libraries:
  [
    "/electron/electron",
    "/electron/forge",
    "/electron-userland/electron-builder",
  ]
related-skills:
  ["moai-lang-typescript", "moai-domain-frontend", "moai-lang-javascript"]
updated: 2026-01-10
status: "active"
allowed-tools:
  - Read
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

## Quick Reference (30 seconds)

Electron 33+ Desktop App Development Specialist - Cross-platform desktop applications with web technologies.

Auto-Triggers: Electron projects (`electron.vite.config.ts`, `electron-builder.yml`), desktop app development, IPC communication patterns

Core Capabilities:

- Electron 33: Chromium 130, Node.js 20.18, native ESM support
- Process Architecture: Main process, Renderer process, Preload scripts
- IPC Communication: Type-safe invoke/handle patterns with contextBridge
- Auto-Update: electron-updater with GitHub/S3 publishing
- Packaging: Electron Forge, electron-builder for all platforms
- Security: contextIsolation, sandbox, CSP, input validation

Quick Commands:

```bash
# Create new Electron app with Vite
npm create @electron/create-electron-app@latest my-app -- --template=vite-typescript

# Install electron-builder
npm install -D electron-builder

# Install auto-updater
npm install electron-updater
```

---

## Implementation Guide (5 minutes)

### Project Structure

Recommended Directory Layout:

```
electron-app/
├── src/
│   ├── main/                   # Main process code
│   │   ├── index.ts           # Entry point
│   │   ├── ipc/               # IPC handlers
│   │   ├── services/          # Business logic
│   │   └── windows/           # Window management
│   ├── preload/               # Preload scripts
│   │   ├── index.ts           # Main preload
│   │   └── api.ts             # Exposed APIs
│   ├── renderer/              # React/Vue/Svelte app
│   │   ├── src/
│   │   ├── index.html
│   │   └── vite.config.ts
│   └── shared/                # Shared types/constants
├── resources/                 # App resources
├── electron.vite.config.ts
├── electron-builder.yml
└── package.json
```

### Main Process Setup

Application Lifecycle:

```typescript
// src/main/index.ts
import { app, BrowserWindow, session } from "electron";
import { join } from "path";

// Enable sandbox globally for all renderers (security best practice)
app.enableSandbox();

let mainWindow: BrowserWindow | null = null;

async function createMainWindow(): Promise<void> {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    show: false,
    webPreferences: {
      preload: join(__dirname, "../preload/index.js"),
      sandbox: true,
      contextIsolation: true,
      nodeIntegration: false,
      webSecurity: true,
    },
  });

  mainWindow.on("ready-to-show", () => mainWindow?.show());

  if (process.env.NODE_ENV === "development") {
    mainWindow.loadURL("http://localhost:5173");
    mainWindow.webContents.openDevTools();
  } else {
    mainWindow.loadFile(join(__dirname, "../renderer/index.html"));
  }
}

app.whenReady().then(createMainWindow);

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") app.quit();
});

app.on("activate", () => {
  if (BrowserWindow.getAllWindows().length === 0) createMainWindow();
});
```

Single Instance Lock:

```typescript
const gotSingleLock = app.requestSingleInstanceLock();
if (!gotSingleLock) {
  app.quit();
} else {
  app.on("second-instance", () => {
    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore();
      mainWindow.focus();
    }
  });
}
```

### Type-Safe IPC Communication

Define IPC Types:

```typescript
// src/shared/ipc-types.ts
export interface IpcChannels {
  "file:open": { path: string };
  "file:save": { path: string; content: string };
  "file:read": string;
  "window:minimize": void;
  "window:maximize": void;
  "storage:get": { key: string };
  "storage:set": { key: string; value: unknown };
}
```

Main Process Handlers:

```typescript
// src/main/ipc/index.ts
import { ipcMain, dialog, BrowserWindow } from "electron";
import { readFile, writeFile } from "fs/promises";
import Store from "electron-store";

const store = new Store();

export function registerIpcHandlers(): void {
  ipcMain.handle("file:open", async () => {
    const result = await dialog.showOpenDialog({
      properties: ["openFile"],
    });
    if (result.canceled) return null;
    const content = await readFile(result.filePaths[0], "utf-8");
    return { path: result.filePaths[0], content };
  });

  ipcMain.handle("file:save", async (_, { path, content }) => {
    await writeFile(path, content, "utf-8");
    return { success: true };
  });

  ipcMain.handle("storage:get", (_, { key }) => store.get(key));
  ipcMain.handle("storage:set", (_, { key, value }) => {
    store.set(key, value);
    return { success: true };
  });
}
```

Preload Script with contextBridge:

```typescript
// src/preload/index.ts
import { contextBridge, ipcRenderer } from "electron";

const api = {
  file: {
    open: () => ipcRenderer.invoke("file:open"),
    save: (path: string, content: string) =>
      ipcRenderer.invoke("file:save", { path, content }),
    read: (path: string) => ipcRenderer.invoke("file:read", path),
  },
  storage: {
    get: <T>(key: string): Promise<T | undefined> =>
      ipcRenderer.invoke("storage:get", { key }),
    set: (key: string, value: unknown) =>
      ipcRenderer.invoke("storage:set", { key, value }),
  },
  onUpdateAvailable: (callback: (version: string) => void) => {
    const handler = (_: unknown, { version }: { version: string }) =>
      callback(version);
    ipcRenderer.on("app:update-available", handler);
    return () => ipcRenderer.removeListener("app:update-available", handler);
  },
};

contextBridge.exposeInMainWorld("electronAPI", api);

declare global {
  interface Window {
    electronAPI: typeof api;
  }
}
```

### Security Best Practices

Mandatory Security Settings:

```typescript
const mainWindow = new BrowserWindow({
  webPreferences: {
    contextIsolation: true, // Always enable
    nodeIntegration: false, // Never enable in renderer
    sandbox: true, // Always enable
    webSecurity: true, // Never disable
    preload: join(__dirname, "../preload/index.js"),
  },
});
```

Content Security Policy:

```typescript
session.defaultSession.webRequest.onHeadersReceived((details, callback) => {
  callback({
    responseHeaders: {
      ...details.responseHeaders,
      "Content-Security-Policy": [
        "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'",
      ],
    },
  });
});
```

Input Validation with Zod:

```typescript
import { z } from "zod";

const FilePathSchema = z
  .string()
  .refine((path) => !path.includes("..") && !path.startsWith("/"), {
    message: "Invalid file path",
  });

ipcMain.handle("file:read", async (_, path: string) => {
  const validPath = FilePathSchema.parse(path);
  return readFile(validPath, "utf-8");
});
```

### Auto-Update Implementation

Simplified Auto-Update (for GitHub releases):

```typescript
// Simplified auto-update (recommended for GitHub releases)
// npm install update-electron-app
require("update-electron-app")();

// Or use electron-updater for more control (shown below)
```

Full Update Service:

```typescript
// src/main/services/updater.ts
import { autoUpdater, UpdateInfo } from "electron-updater";
import { BrowserWindow, dialog } from "electron";
import log from "electron-log";

export class UpdateService {
  private mainWindow: BrowserWindow | null = null;

  initialize(window: BrowserWindow): void {
    this.mainWindow = window;
    autoUpdater.logger = log;
    autoUpdater.autoDownload = false;

    autoUpdater.on("update-available", (info: UpdateInfo) => {
      this.mainWindow?.webContents.send("app:update-available", {
        version: info.version,
      });
      this.promptForUpdate(info);
    });

    autoUpdater.on("update-downloaded", () => {
      this.promptForRestart();
    });
  }

  async checkForUpdates(): Promise<void> {
    await autoUpdater.checkForUpdates();
  }

  private async promptForUpdate(info: UpdateInfo): Promise<void> {
    const result = await dialog.showMessageBox(this.mainWindow!, {
      type: "info",
      title: "Update Available",
      message: `Version ${info.version} is available. Download now?`,
      buttons: ["Download", "Later"],
    });
    if (result.response === 0) autoUpdater.downloadUpdate();
  }

  private async promptForRestart(): Promise<void> {
    const result = await dialog.showMessageBox(this.mainWindow!, {
      type: "info",
      title: "Update Ready",
      message: "Restart to apply the update?",
      buttons: ["Restart Now", "Later"],
    });
    if (result.response === 0) autoUpdater.quitAndInstall(false, true);
  }
}
```

### App Packaging

Electron Builder Configuration:

```yaml
# electron-builder.yml
appId: com.example.myapp
productName: My App
directories:
  output: dist
  buildResources: resources

mac:
  category: public.app-category.developer-tools
  hardenedRuntime: true
  gatekeeperAssess: false
  entitlements: entitlements.mac.plist
  entitlementsInherit: entitlements.mac.plist
  notarize:
    teamId: ${APPLE_TEAM_ID}
  target:
    - target: dmg
      arch: [x64, arm64]

win:
  icon: resources/icons/icon.ico
  signingHashAlgorithms: [sha256]
  signAndEditExecutable: true
  target:
    - target: nsis
      arch: [x64]

linux:
  category: Development
  target:
    - AppImage
    - deb

publish:
  provider: github
  owner: your-username
  repo: your-repo
```

Electron Forge Configuration:

```javascript
// forge.config.js
module.exports = {
  packagerConfig: {
    asar: true,
    darwinDarkModeSupport: true,
    icon: "./resources/icons/icon",
    osxSign: { identity: "Developer ID Application: Your Name (TEAM_ID)" },
  },
  makers: [
    { name: "@electron-forge/maker-squirrel" },
    { name: "@electron-forge/maker-dmg" },
    { name: "@electron-forge/maker-deb" },
  ],
  plugins: [
    {
      name: "@electron-forge/plugin-vite",
      config: {
        build: [
          { entry: "src/main/index.ts", config: "vite.main.config.ts" },
          { entry: "src/preload/index.ts", config: "vite.preload.config.ts" },
        ],
        renderer: [{ name: "main_window", config: "vite.renderer.config.ts" }],
      },
    },
  ],
};
```

---

## Advanced Patterns

For comprehensive documentation including:

- Window state persistence and multi-window management
- System tray and native menus
- Utility processes for background tasks
- Native module integration
- Deep linking and protocol handlers
- Performance optimization

See: [reference.md](reference.md) and [examples.md](examples.md)

---

## Works Well With

- `moai-lang-typescript` - TypeScript patterns for type-safe Electron development
- `moai-domain-frontend` - React/Vue/Svelte renderer development
- `moai-lang-javascript` - Node.js patterns for main process
- `moai-domain-backend` - Backend API integration
- `moai-workflow-testing` - Testing strategies for desktop apps

---

## Quick Troubleshooting

Common Issues:

White Screen on Launch:

- Check preload script path is correct
- Verify loadFile/loadURL path
- Check for CSP blocking scripts
- Enable devTools to see errors

IPC Not Working:

- Verify channel names match exactly
- Check handler registered before window loads
- Ensure contextBridge used correctly

Native Modules Fail:

- Run `electron-rebuild` after npm install
- Match Electron Node.js version
- Use postinstall script for automatic rebuild

Auto-Update Not Working:

- Ensure app is signed (required for updates)
- Check publish configuration
- Enable electron-updater logging

Debug Commands:

```bash
# Rebuild native modules
npx electron-rebuild

# Check Electron version
npx electron --version

# Run with verbose logging
DEBUG=electron-updater npm start
```
