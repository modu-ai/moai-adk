# Electron Framework Reference Guide

## Platform Version Matrix

### Electron 33 (October 2024) - Current Stable

- Chromium: 130
- Node.js: 20.18.0
- V8: 13.0
- Key Features:
  - Enhanced security defaults with sandbox enabled by default
  - Improved context isolation patterns
  - Native ESM support in main process
  - Service Worker support in renderer
  - WebGPU API support for GPU-accelerated graphics
  - Improved auto-updater with differential updates
  - Utility process enhancements for background tasks
  - Better crash reporting integration

### Electron 32 (August 2024)

- Chromium: 128
- Node.js: 20.16.0
- Key Features:
  - Utility process improvements
  - Enhanced file system access
  - Better macOS notarization support
  - Improved window management APIs

### Electron 31 (June 2024)

- Chromium: 126
- Node.js: 20.14.0
- Key Features:
  - Performance improvements for large apps
  - Enhanced IPC serialization
  - Better TypeScript support

## Context7 Library Mappings

### Core Framework

```
/electron/electron              - Electron framework
/electron/forge                 - Electron Forge tooling
/electron-userland/electron-builder - App packaging
```

### Build Tools

```
/nickmeinhold/electron-vite    - Vite integration
/nickmeinhold/electron-esbuild - esbuild integration
```

### Native Modules

```
/nickmeinhold/better-sqlite3   - SQLite database
/nickmeinhold/keytar           - Secure credential storage
/nickmeinhold/node-pty         - Terminal emulation
```

### Auto-Update

```
/electron-userland/electron-updater - Auto-update support
```

### Testing

```
/nickmeinhold/spectron         - E2E testing (deprecated)
/nickmeinhold/playwright       - Modern E2E testing
```

---

## Architecture Patterns

### Process Model

Electron Process Architecture:

```
                    ┌─────────────────────────────────────┐
                    │          Main Process               │
                    │  - Single instance per app          │
                    │  - Full Node.js access              │
                    │  - Creates BrowserWindows           │
                    │  - Manages app lifecycle            │
                    │  - Native OS integration            │
                    └─────────────┬───────────────────────┘
                                  │
            ┌─────────────────────┼─────────────────────┐
            │                     │                     │
            ▼                     ▼                     ▼
┌───────────────────┐ ┌───────────────────┐ ┌───────────────────┐
│  Renderer Process │ │  Renderer Process │ │  Utility Process  │
│  - Web content    │ │  - Web content    │ │  - Background     │
│  - Sandboxed      │ │  - Sandboxed      │ │    tasks          │
│  - No Node.js     │ │  - No Node.js     │ │  - Node.js access │
│    (default)      │ │    (default)      │ │  - No GUI         │
└───────────────────┘ └───────────────────┘ └───────────────────┘
```

### Recommended Project Structure

Directory Layout:

```
electron-app/
├── src/
│   ├── main/                   # Main process code
│   │   ├── index.ts           # Entry point
│   │   ├── app.ts             # App lifecycle
│   │   ├── ipc/               # IPC handlers
│   │   │   ├── index.ts
│   │   │   ├── file-handlers.ts
│   │   │   └── window-handlers.ts
│   │   ├── services/          # Business logic
│   │   │   ├── storage.ts
│   │   │   └── updater.ts
│   │   └── windows/           # Window management
│   │       ├── main-window.ts
│   │       └── settings-window.ts
│   ├── preload/               # Preload scripts
│   │   ├── index.ts           # Main preload
│   │   └── api.ts             # Exposed APIs
│   ├── renderer/              # React/Vue/Svelte app
│   │   ├── src/
│   │   ├── index.html
│   │   └── vite.config.ts
│   └── shared/                # Shared types/constants
│       ├── types.ts
│       └── constants.ts
├── resources/                 # App resources
│   ├── icons/
│   └── locales/
├── electron.vite.config.ts
├── electron-builder.yml
└── package.json
```

---

## Main Process APIs

### App Lifecycle

```typescript
// src/main/app.ts
import { app, BrowserWindow, session } from "electron";
import { join } from "path";

class Application {
  private mainWindow: BrowserWindow | null = null;

  async initialize(): Promise<void> {
    // Set app user model ID for Windows
    if (process.platform === "win32") {
      app.setAppUserModelId(app.getName());
    }

    // Prevent multiple instances
    const gotSingleLock = app.requestSingleInstanceLock();
    if (!gotSingleLock) {
      app.quit();
      return;
    }

    app.on("second-instance", () => {
      if (this.mainWindow) {
        if (this.mainWindow.isMinimized()) {
          this.mainWindow.restore();
        }
        this.mainWindow.focus();
      }
    });

    // macOS: Re-create window when dock icon clicked
    app.on("activate", () => {
      if (BrowserWindow.getAllWindows().length === 0) {
        this.createMainWindow();
      }
    });

    // Quit when all windows closed (except macOS)
    app.on("window-all-closed", () => {
      if (process.platform !== "darwin") {
        app.quit();
      }
    });

    // Wait for ready
    await app.whenReady();

    // Configure session
    this.configureSession();

    // Create main window
    this.createMainWindow();
  }

  private configureSession(): void {
    // Configure Content Security Policy
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

    // Configure permissions
    session.defaultSession.setPermissionRequestHandler(
      (webContents, permission, callback) => {
        const allowedPermissions = ["notifications", "clipboard-read"];
        callback(allowedPermissions.includes(permission));
      },
    );
  }

  private createMainWindow(): void {
    this.mainWindow = new BrowserWindow({
      width: 1200,
      height: 800,
      minWidth: 800,
      minHeight: 600,
      show: false,
      webPreferences: {
        preload: join(__dirname, "../preload/index.js"),
        sandbox: true,
        contextIsolation: true,
        nodeIntegration: false,
        webSecurity: true,
      },
    });

    // Show window when ready
    this.mainWindow.on("ready-to-show", () => {
      this.mainWindow?.show();
    });

    // Load app content
    if (process.env.NODE_ENV === "development") {
      this.mainWindow.loadURL("http://localhost:5173");
      this.mainWindow.webContents.openDevTools();
    } else {
      this.mainWindow.loadFile(join(__dirname, "../renderer/index.html"));
    }
  }
}

export const application = new Application();
```

### Window Management

```typescript
// src/main/windows/window-manager.ts
import {
  BrowserWindow,
  BrowserWindowConstructorOptions,
  screen,
} from "electron";
import { join } from "path";

interface WindowState {
  width: number;
  height: number;
  x?: number;
  y?: number;
  isMaximized: boolean;
}

export class WindowManager {
  private windows = new Map<string, BrowserWindow>();
  private stateStore: Map<string, WindowState> = new Map();

  createWindow(
    id: string,
    options: BrowserWindowConstructorOptions = {},
  ): BrowserWindow {
    // Restore previous state
    const savedState = this.stateStore.get(id);
    const { width, height } = screen.getPrimaryDisplay().workAreaSize;

    const defaultOptions: BrowserWindowConstructorOptions = {
      width: savedState?.width ?? Math.floor(width * 0.8),
      height: savedState?.height ?? Math.floor(height * 0.8),
      x: savedState?.x,
      y: savedState?.y,
      show: false,
      webPreferences: {
        preload: join(__dirname, "../preload/index.js"),
        sandbox: true,
        contextIsolation: true,
        nodeIntegration: false,
      },
    };

    const window = new BrowserWindow({
      ...defaultOptions,
      ...options,
      webPreferences: {
        ...defaultOptions.webPreferences,
        ...options.webPreferences,
      },
    });

    // Restore maximized state
    if (savedState?.isMaximized) {
      window.maximize();
    }

    // Save state on close
    window.on("close", () => {
      this.saveWindowState(id, window);
    });

    window.on("closed", () => {
      this.windows.delete(id);
    });

    this.windows.set(id, window);
    return window;
  }

  getWindow(id: string): BrowserWindow | undefined {
    return this.windows.get(id);
  }

  closeWindow(id: string): void {
    const window = this.windows.get(id);
    if (window && !window.isDestroyed()) {
      window.close();
    }
  }

  private saveWindowState(id: string, window: BrowserWindow): void {
    const bounds = window.getBounds();
    this.stateStore.set(id, {
      width: bounds.width,
      height: bounds.height,
      x: bounds.x,
      y: bounds.y,
      isMaximized: window.isMaximized(),
    });
  }
}

export const windowManager = new WindowManager();
```

---

## IPC Communication

### Type-Safe IPC Pattern

```typescript
// src/shared/ipc-types.ts
export interface IpcChannels {
  // Main -> Renderer
  "app:update-available": { version: string };
  "app:update-downloaded": void;

  // Renderer -> Main (invoke)
  "file:open": { path: string };
  "file:save": { path: string; content: string };
  "file:read": string; // Returns file content
  "window:minimize": void;
  "window:maximize": void;
  "window:close": void;
  "storage:get": { key: string };
  "storage:set": { key: string; value: unknown };
}

export type IpcChannel = keyof IpcChannels;
export type IpcPayload<C extends IpcChannel> = IpcChannels[C];
```

### Main Process Handlers

```typescript
// src/main/ipc/index.ts
import { ipcMain, dialog, BrowserWindow } from "electron";
import { readFile, writeFile } from "fs/promises";
import Store from "electron-store";

const store = new Store();

export function registerIpcHandlers(): void {
  // File operations
  ipcMain.handle("file:open", async () => {
    const result = await dialog.showOpenDialog({
      properties: ["openFile"],
      filters: [
        { name: "All Files", extensions: ["*"] },
        { name: "Text", extensions: ["txt", "md"] },
      ],
    });

    if (result.canceled || result.filePaths.length === 0) {
      return null;
    }

    const filePath = result.filePaths[0];
    const content = await readFile(filePath, "utf-8");
    return { path: filePath, content };
  });

  ipcMain.handle(
    "file:save",
    async (_event, { path, content }: { path: string; content: string }) => {
      await writeFile(path, content, "utf-8");
      return { success: true };
    },
  );

  ipcMain.handle("file:read", async (_event, path: string) => {
    return readFile(path, "utf-8");
  });

  // Window operations
  ipcMain.handle("window:minimize", (event) => {
    const window = BrowserWindow.fromWebContents(event.sender);
    window?.minimize();
  });

  ipcMain.handle("window:maximize", (event) => {
    const window = BrowserWindow.fromWebContents(event.sender);
    if (window?.isMaximized()) {
      window.unmaximize();
    } else {
      window?.maximize();
    }
  });

  ipcMain.handle("window:close", (event) => {
    const window = BrowserWindow.fromWebContents(event.sender);
    window?.close();
  });

  // Storage operations
  ipcMain.handle("storage:get", (_event, { key }: { key: string }) => {
    return store.get(key);
  });

  ipcMain.handle(
    "storage:set",
    (_event, { key, value }: { key: string; value: unknown }) => {
      store.set(key, value);
      return { success: true };
    },
  );
}
```

### Preload Script

```typescript
// src/preload/index.ts
import { contextBridge, ipcRenderer } from "electron";

// Expose protected methods for renderer
const api = {
  // Window controls
  window: {
    minimize: () => ipcRenderer.invoke("window:minimize"),
    maximize: () => ipcRenderer.invoke("window:maximize"),
    close: () => ipcRenderer.invoke("window:close"),
  },

  // File operations
  file: {
    open: () => ipcRenderer.invoke("file:open"),
    save: (path: string, content: string) =>
      ipcRenderer.invoke("file:save", { path, content }),
    read: (path: string) => ipcRenderer.invoke("file:read", path),
  },

  // Storage
  storage: {
    get: <T>(key: string): Promise<T | undefined> =>
      ipcRenderer.invoke("storage:get", { key }),
    set: (key: string, value: unknown) =>
      ipcRenderer.invoke("storage:set", { key, value }),
  },

  // App events
  onUpdateAvailable: (callback: (version: string) => void) => {
    const handler = (
      _event: Electron.IpcRendererEvent,
      { version }: { version: string },
    ) => {
      callback(version);
    };
    ipcRenderer.on("app:update-available", handler);
    return () => ipcRenderer.removeListener("app:update-available", handler);
  },

  onUpdateDownloaded: (callback: () => void) => {
    const handler = () => callback();
    ipcRenderer.on("app:update-downloaded", handler);
    return () => ipcRenderer.removeListener("app:update-downloaded", handler);
  },
};

contextBridge.exposeInMainWorld("electronAPI", api);

// Type declaration for renderer
declare global {
  interface Window {
    electronAPI: typeof api;
  }
}
```

---

## Auto-Update

### Update Service

```typescript
// src/main/services/updater.ts
import { autoUpdater, UpdateInfo } from "electron-updater";
import { BrowserWindow, dialog } from "electron";
import log from "electron-log";

export class UpdateService {
  private mainWindow: BrowserWindow | null = null;

  initialize(window: BrowserWindow): void {
    this.mainWindow = window;

    // Configure logging
    autoUpdater.logger = log;
    autoUpdater.autoDownload = false;
    autoUpdater.autoInstallOnAppQuit = true;

    // Event handlers
    autoUpdater.on("checking-for-update", () => {
      log.info("Checking for updates...");
    });

    autoUpdater.on("update-available", (info: UpdateInfo) => {
      log.info("Update available:", info.version);
      this.mainWindow?.webContents.send("app:update-available", {
        version: info.version,
      });
      this.promptForUpdate(info);
    });

    autoUpdater.on("update-not-available", () => {
      log.info("No updates available");
    });

    autoUpdater.on("error", (error) => {
      log.error("Update error:", error);
    });

    autoUpdater.on("download-progress", (progress) => {
      log.info(`Download progress: ${progress.percent.toFixed(1)}%`);
    });

    autoUpdater.on("update-downloaded", () => {
      log.info("Update downloaded");
      this.mainWindow?.webContents.send("app:update-downloaded");
      this.promptForRestart();
    });
  }

  async checkForUpdates(): Promise<void> {
    try {
      await autoUpdater.checkForUpdates();
    } catch (error) {
      log.error("Failed to check for updates:", error);
    }
  }

  private async promptForUpdate(info: UpdateInfo): Promise<void> {
    const result = await dialog.showMessageBox(this.mainWindow!, {
      type: "info",
      title: "Update Available",
      message: `Version ${info.version} is available. Would you like to download it?`,
      buttons: ["Download", "Later"],
    });

    if (result.response === 0) {
      autoUpdater.downloadUpdate();
    }
  }

  private async promptForRestart(): Promise<void> {
    const result = await dialog.showMessageBox(this.mainWindow!, {
      type: "info",
      title: "Update Ready",
      message:
        "A new version has been downloaded. Restart to apply the update?",
      buttons: ["Restart Now", "Later"],
    });

    if (result.response === 0) {
      autoUpdater.quitAndInstall(false, true);
    }
  }
}

export const updateService = new UpdateService();
```

---

## Security Best Practices

### Security Checklist

Mandatory Security Settings:

- contextIsolation: true (always enable)
- nodeIntegration: false (never enable in renderer)
- sandbox: true (always enable)
- webSecurity: true (never disable)

IPC Security Rules:

- Validate all inputs from renderer
- Never expose Node.js APIs directly
- Use invoke/handle pattern (not send/on for sensitive operations)
- Whitelist allowed operations

Content Security Policy:

- Restrict script sources to 'self'
- Disable unsafe-inline for scripts
- Use nonce or hash for inline scripts if needed

### Input Validation

```typescript
// src/main/ipc/validators.ts
import { z } from "zod";

const FilePathSchema = z.string().refine(
  (path) => {
    // Prevent path traversal
    const normalized = path.replace(/\\/g, "/");
    return !normalized.includes("..") && !normalized.startsWith("/");
  },
  { message: "Invalid file path" },
);

const StorageKeySchema = z
  .string()
  .min(1)
  .max(100)
  .regex(/^[a-zA-Z0-9_.-]+$/);

export const validators = {
  filePath: (path: unknown) => FilePathSchema.parse(path),
  storageKey: (key: unknown) => StorageKeySchema.parse(key),
};
```

---

## Native Integration

### System Tray

```typescript
// src/main/services/tray.ts
import { Tray, Menu, app, nativeImage } from "electron";
import { join } from "path";

export class TrayService {
  private tray: Tray | null = null;

  initialize(): void {
    const iconPath = join(__dirname, "../../resources/icons/tray.png");
    const icon = nativeImage.createFromPath(iconPath);

    this.tray = new Tray(icon);
    this.tray.setToolTip(app.getName());

    const contextMenu = Menu.buildFromTemplate([
      {
        label: "Show App",
        click: () => {
          const { windowManager } = require("./window-manager");
          const mainWindow = windowManager.getWindow("main");
          mainWindow?.show();
          mainWindow?.focus();
        },
      },
      { type: "separator" },
      {
        label: "Preferences",
        accelerator: "CmdOrCtrl+,",
        click: () => {
          // Open preferences
        },
      },
      { type: "separator" },
      {
        label: "Quit",
        accelerator: "CmdOrCtrl+Q",
        click: () => app.quit(),
      },
    ]);

    this.tray.setContextMenu(contextMenu);

    // macOS: Click to show app
    this.tray.on("click", () => {
      const { windowManager } = require("./window-manager");
      const mainWindow = windowManager.getWindow("main");
      mainWindow?.show();
      mainWindow?.focus();
    });
  }

  destroy(): void {
    this.tray?.destroy();
    this.tray = null;
  }
}

export const trayService = new TrayService();
```

### Native Menu

```typescript
// src/main/services/menu.ts
import { Menu, app, shell, MenuItemConstructorOptions } from "electron";

export function createApplicationMenu(): void {
  const isMac = process.platform === "darwin";

  const template: MenuItemConstructorOptions[] = [
    // App menu (macOS only)
    ...(isMac
      ? [
          {
            label: app.name,
            submenu: [
              { role: "about" as const },
              { type: "separator" as const },
              { role: "services" as const },
              { type: "separator" as const },
              { role: "hide" as const },
              { role: "hideOthers" as const },
              { role: "unhide" as const },
              { type: "separator" as const },
              { role: "quit" as const },
            ],
          },
        ]
      : []),

    // File menu
    {
      label: "File",
      submenu: [
        {
          label: "New",
          accelerator: "CmdOrCtrl+N",
          click: () => {
            // Handle new file
          },
        },
        {
          label: "Open...",
          accelerator: "CmdOrCtrl+O",
          click: () => {
            // Handle open
          },
        },
        { type: "separator" },
        {
          label: "Save",
          accelerator: "CmdOrCtrl+S",
          click: () => {
            // Handle save
          },
        },
        { type: "separator" },
        isMac ? { role: "close" } : { role: "quit" },
      ],
    },

    // Edit menu
    {
      label: "Edit",
      submenu: [
        { role: "undo" },
        { role: "redo" },
        { type: "separator" },
        { role: "cut" },
        { role: "copy" },
        { role: "paste" },
        { role: "selectAll" },
      ],
    },

    // View menu
    {
      label: "View",
      submenu: [
        { role: "reload" },
        { role: "forceReload" },
        { role: "toggleDevTools" },
        { type: "separator" },
        { role: "resetZoom" },
        { role: "zoomIn" },
        { role: "zoomOut" },
        { type: "separator" },
        { role: "togglefullscreen" },
      ],
    },

    // Help menu
    {
      label: "Help",
      submenu: [
        {
          label: "Documentation",
          click: () => shell.openExternal("https://docs.example.com"),
        },
        {
          label: "Report Issue",
          click: () =>
            shell.openExternal("https://github.com/example/repo/issues"),
        },
      ],
    },
  ];

  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);
}
```

---

## Configuration

### Electron Forge Configuration

```javascript
// forge.config.js
module.exports = {
  packagerConfig: {
    asar: true,
    darwinDarkModeSupport: true,
    executableName: "my-app",
    appBundleId: "com.example.myapp",
    appCategoryType: "public.app-category.developer-tools",
    icon: "./resources/icons/icon",
    osxSign: {
      identity: "Developer ID Application: Your Name (TEAM_ID)",
      "hardened-runtime": true,
      entitlements: "./entitlements.plist",
      "entitlements-inherit": "./entitlements.plist",
      "signature-flags": "library",
    },
    osxNotarize: {
      appleId: process.env.APPLE_ID,
      appleIdPassword: process.env.APPLE_PASSWORD,
      teamId: process.env.APPLE_TEAM_ID,
    },
  },
  rebuildConfig: {},
  makers: [
    {
      name: "@electron-forge/maker-squirrel",
      config: {
        name: "my_app",
        setupIcon: "./resources/icons/icon.ico",
      },
    },
    {
      name: "@electron-forge/maker-zip",
      platforms: ["darwin"],
    },
    {
      name: "@electron-forge/maker-dmg",
      config: {
        icon: "./resources/icons/icon.icns",
        format: "ULFO",
      },
    },
    {
      name: "@electron-forge/maker-deb",
      config: {
        options: {
          maintainer: "Your Name",
          homepage: "https://example.com",
        },
      },
    },
    {
      name: "@electron-forge/maker-rpm",
      config: {},
    },
  ],
  plugins: [
    {
      name: "@electron-forge/plugin-vite",
      config: {
        build: [
          {
            entry: "src/main/index.ts",
            config: "vite.main.config.ts",
          },
          {
            entry: "src/preload/index.ts",
            config: "vite.preload.config.ts",
          },
        ],
        renderer: [
          {
            name: "main_window",
            config: "vite.renderer.config.ts",
          },
        ],
      },
    },
  ],
  publishers: [
    {
      name: "@electron-forge/publisher-github",
      config: {
        repository: {
          owner: "your-username",
          name: "your-repo",
        },
        prerelease: false,
      },
    },
  ],
};
```

### Electron Builder Configuration

```yaml
# electron-builder.yml
appId: com.example.myapp
productName: My App
copyright: Copyright (c) 2025 Your Name

directories:
  output: dist
  buildResources: resources

files:
  - "!**/.vscode/*"
  - "!src/*"
  - "!docs/*"
  - "!*.md"

extraResources:
  - from: resources/
    to: resources/
    filter:
      - "**/*"

asar: true
compression: maximum

mac:
  category: public.app-category.developer-tools
  icon: resources/icons/icon.icns
  hardenedRuntime: true
  gatekeeperAssess: false
  entitlements: entitlements.mac.plist
  entitlementsInherit: entitlements.mac.plist
  notarize:
    teamId: ${APPLE_TEAM_ID}
  target:
    - target: dmg
      arch: [x64, arm64]
    - target: zip
      arch: [x64, arm64]

dmg:
  sign: false
  contents:
    - x: 130
      y: 220
    - x: 410
      y: 220
      type: link
      path: /Applications

win:
  icon: resources/icons/icon.ico
  signingHashAlgorithms: [sha256]
  signAndEditExecutable: true
  target:
    - target: nsis
      arch: [x64]
    - target: portable
      arch: [x64]

nsis:
  oneClick: false
  allowToChangeInstallationDirectory: true
  installerIcon: resources/icons/icon.ico
  uninstallerIcon: resources/icons/icon.ico
  installerHeaderIcon: resources/icons/icon.ico
  createDesktopShortcut: true
  createStartMenuShortcut: true

linux:
  icon: resources/icons
  category: Development
  target:
    - target: AppImage
      arch: [x64]
    - target: deb
      arch: [x64]
    - target: rpm
      arch: [x64]

publish:
  provider: github
  owner: your-username
  repo: your-repo
```

---

## Utility Process (Electron 33+)

### Background Task Worker

Utility processes run in a separate Node.js environment for CPU-intensive tasks without blocking the main process:

```typescript
// src/main/workers/utility-worker.ts
import { utilityProcess, MessageChannelMain } from "electron";
import { join } from "path";

export class UtilityWorker {
  private worker: Electron.UtilityProcess | null = null;
  private port: Electron.MessagePortMain | null = null;

  async spawn(): Promise<void> {
    const { port1, port2 } = new MessageChannelMain();

    this.worker = utilityProcess.fork(
      join(__dirname, "workers/image-processor.js"),
      [],
      {
        serviceName: "image-processor",
        allowLoadingUnsignedLibraries: false,
      }
    );

    this.worker.postMessage({ type: "init" }, [port1]);
    this.port = port2;

    this.worker.on("exit", (code) => {
      console.log(`Utility process exited with code ${code}`);
      this.worker = null;
    });
  }

  async processImage(imagePath: string): Promise<Buffer> {
    return new Promise((resolve, reject) => {
      if (!this.port) {
        reject(new Error("Worker not initialized"));
        return;
      }

      const handler = (event: Electron.MessageEvent) => {
        if (event.data.type === "result") {
          this.port?.removeListener("message", handler);
          resolve(Buffer.from(event.data.buffer));
        } else if (event.data.type === "error") {
          this.port?.removeListener("message", handler);
          reject(new Error(event.data.message));
        }
      };

      this.port.on("message", handler);
      this.port.postMessage({ type: "process", path: imagePath });
    });
  }

  terminate(): void {
    this.worker?.kill();
    this.worker = null;
    this.port = null;
  }
}
```

### Utility Process Script

```typescript
// src/main/workers/image-processor.js
const sharp = require("sharp");

process.parentPort.on("message", async (event) => {
  const [port] = event.ports;

  port.on("message", async (msgEvent) => {
    const { type, path } = msgEvent.data;

    if (type === "process") {
      try {
        const buffer = await sharp(path)
          .resize(800, 600, { fit: "inside" })
          .webp({ quality: 80 })
          .toBuffer();

        port.postMessage({ type: "result", buffer });
      } catch (error) {
        port.postMessage({ type: "error", message: error.message });
      }
    }
  });

  port.start();
});
```

---

## Protocol Handlers and Deep Linking

### Custom Protocol Registration

```typescript
// src/main/protocol.ts
import { app, protocol, net } from "electron";
import { join } from "path";
import { pathToFileURL } from "url";

const PROTOCOL_NAME = "myapp";

export function registerProtocols(): void {
  // Register as default protocol client (for deep linking)
  if (process.defaultApp) {
    if (process.argv.length >= 2) {
      app.setAsDefaultProtocolClient(PROTOCOL_NAME, process.execPath, [
        join(process.argv[1]),
      ]);
    }
  } else {
    app.setAsDefaultProtocolClient(PROTOCOL_NAME);
  }

  // Register custom protocol handler for local resources
  protocol.handle("app", (request) => {
    const url = new URL(request.url);
    const filePath = join(__dirname, "../renderer", url.pathname);
    return net.fetch(pathToFileURL(filePath).toString());
  });
}

export function handleProtocolUrl(url: string): void {
  // Parse the URL: myapp://action/path?query=value
  const parsedUrl = new URL(url);

  switch (parsedUrl.hostname) {
    case "open":
      handleOpenAction(parsedUrl.pathname, parsedUrl.searchParams);
      break;
    case "auth":
      handleAuthCallback(parsedUrl.searchParams);
      break;
    default:
      console.warn("Unknown protocol action:", parsedUrl.hostname);
  }
}

function handleOpenAction(
  path: string,
  params: URLSearchParams
): void {
  // Handle open file/project action
  const filePath = decodeURIComponent(path.slice(1));
  // Send to renderer or process directly
}

function handleAuthCallback(params: URLSearchParams): void {
  // Handle OAuth callback
  const code = params.get("code");
  const state = params.get("state");
  // Process authentication
}
```

### macOS Deep Link Handling

```typescript
// src/main/index.ts
app.on("open-url", (event, url) => {
  event.preventDefault();
  handleProtocolUrl(url);
});

// Handle deep link on Windows/Linux (second instance)
app.on("second-instance", (_event, commandLine) => {
  const url = commandLine.find((arg) => arg.startsWith(`${PROTOCOL_NAME}://`));
  if (url) {
    handleProtocolUrl(url);
  }

  // Focus the main window
  const mainWindow = windowManager.getWindow("main");
  if (mainWindow) {
    if (mainWindow.isMinimized()) mainWindow.restore();
    mainWindow.focus();
  }
});
```

---

## Security Hardening (OWASP Aligned)

### Comprehensive Security Configuration

```typescript
// src/main/security.ts
import { app, session, shell, BrowserWindow } from "electron";

export function configureSecurity(): void {
  // 1. Disable remote module (deprecated but ensure disabled)
  app.on("remote-get-builtin", (event) => event.preventDefault());
  app.on("remote-get-current-web-contents", (event) => event.preventDefault());
  app.on("remote-get-current-window", (event) => event.preventDefault());

  // 2. Block navigation to untrusted origins
  app.on("web-contents-created", (_event, contents) => {
    // Block navigation
    contents.on("will-navigate", (event, url) => {
      const allowedOrigins = ["http://localhost", "file://"];
      const isAllowed = allowedOrigins.some((origin) => url.startsWith(origin));
      if (!isAllowed) {
        event.preventDefault();
        console.warn("Blocked navigation to:", url);
      }
    });

    // Block new windows
    contents.setWindowOpenHandler(({ url }) => {
      // Open external URLs in default browser
      if (url.startsWith("https://") || url.startsWith("http://")) {
        shell.openExternal(url);
      }
      return { action: "deny" };
    });

    // Block webview creation
    contents.on("will-attach-webview", (event) => {
      event.preventDefault();
      console.warn("Blocked webview creation");
    });
  });
}

export function configureSessionSecurity(): void {
  const ses = session.defaultSession;

  // Content Security Policy
  ses.webRequest.onHeadersReceived((details, callback) => {
    callback({
      responseHeaders: {
        ...details.responseHeaders,
        "Content-Security-Policy": [
          [
            "default-src 'self'",
            "script-src 'self'",
            "style-src 'self' 'unsafe-inline'",
            "img-src 'self' data: https:",
            "font-src 'self' data:",
            "connect-src 'self' https://api.example.com",
            "frame-ancestors 'none'",
            "base-uri 'self'",
            "form-action 'self'",
          ].join("; "),
        ],
        "X-Content-Type-Options": ["nosniff"],
        "X-Frame-Options": ["DENY"],
        "X-XSS-Protection": ["1; mode=block"],
      },
    });
  });

  // Permission request handler
  ses.setPermissionRequestHandler((webContents, permission, callback) => {
    const allowedPermissions: Electron.PermissionType[] = [
      "notifications",
      "clipboard-read",
    ];

    const denied: Electron.PermissionType[] = [
      "geolocation",
      "media",
      "mediaKeySystem",
      "midi",
      "pointerLock",
      "fullscreen",
    ];

    if (denied.includes(permission)) {
      console.warn(`Denied permission: ${permission}`);
      callback(false);
      return;
    }

    callback(allowedPermissions.includes(permission));
  });

  // Certificate error handler (for development, not production)
  if (process.env.NODE_ENV !== "development") {
    ses.setCertificateVerifyProc((request, callback) => {
      // Reject invalid certificates in production
      if (request.errorCode !== 0) {
        console.error("Certificate error:", request.hostname);
        callback(-2); // Reject
        return;
      }
      callback(0); // Accept
    });
  }
}
```

### Secure BrowserWindow Factory

```typescript
// src/main/windows/secure-window.ts
import { BrowserWindow, BrowserWindowConstructorOptions } from "electron";
import { join } from "path";

export function createSecureWindow(
  options: BrowserWindowConstructorOptions = {}
): BrowserWindow {
  const secureDefaults: BrowserWindowConstructorOptions = {
    webPreferences: {
      // Security: Isolate renderer from Node.js
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: true,

      // Security: Disable dangerous features
      webSecurity: true,
      allowRunningInsecureContent: false,
      experimentalFeatures: false,
      enableWebSQL: false,

      // Security: Preload script for safe API exposure
      preload: join(__dirname, "../preload/index.js"),

      // Security: Disable devtools in production
      devTools: process.env.NODE_ENV === "development",

      // Performance: Disable unused features
      spellcheck: false,
      backgroundThrottling: true,
    },

    // Security: Prevent title from showing sensitive data
    title: "My App",
  };

  return new BrowserWindow({
    ...secureDefaults,
    ...options,
    webPreferences: {
      ...secureDefaults.webPreferences,
      ...options.webPreferences,
    },
  });
}
```

---

## Crash Reporting and Telemetry

### Crash Reporter Setup

```typescript
// src/main/crash-reporter.ts
import { crashReporter, app } from "electron";
import { join } from "path";

export function initializeCrashReporter(): void {
  crashReporter.start({
    productName: app.getName(),
    companyName: "Your Company",
    submitURL: "https://your-crash-server.com/submit",
    uploadToServer: true,
    ignoreSystemCrashHandler: false,
    rateLimit: true,
    compress: true,
    extra: {
      version: app.getVersion(),
      platform: process.platform,
      arch: process.arch,
    },
  });

  // Log crash reports location for debugging
  console.log("Crash reports path:", app.getPath("crashDumps"));
}

export function addCrashContext(key: string, value: string): void {
  crashReporter.addExtraParameter(key, value);
}
```

### Error Boundary in Main Process

```typescript
// src/main/error-handler.ts
import { dialog, app } from "electron";
import log from "electron-log";

export function setupErrorHandlers(): void {
  // Unhandled promise rejections
  process.on("unhandledRejection", (reason, promise) => {
    log.error("Unhandled Rejection:", reason);

    if (process.env.NODE_ENV === "development") {
      dialog.showErrorBox(
        "Unhandled Promise Rejection",
        String(reason)
      );
    }
  });

  // Uncaught exceptions
  process.on("uncaughtException", (error) => {
    log.error("Uncaught Exception:", error);

    dialog.showErrorBox(
      "Application Error",
      `An unexpected error occurred: ${error.message}\n\nThe application will now quit.`
    );

    app.quit();
  });

  // Renderer process crashes
  app.on("render-process-gone", (event, webContents, details) => {
    log.error("Renderer process gone:", details);

    if (details.reason === "crashed") {
      const options = {
        type: "error" as const,
        title: "Window Crashed",
        message: "This window has crashed.",
        buttons: ["Reload", "Close"],
      };

      dialog.showMessageBox(options).then((result) => {
        if (result.response === 0) {
          webContents.reload();
        }
      });
    }
  });

  // GPU process crashes
  app.on("child-process-gone", (event, details) => {
    log.error("Child process gone:", details);

    if (details.type === "GPU" && details.reason === "crashed") {
      log.warn("GPU process crashed, app may have rendering issues");
    }
  });
}
```

---

## Native Module Integration

### Rebuilding Native Modules

```json
// package.json scripts
{
  "scripts": {
    "postinstall": "electron-builder install-app-deps",
    "rebuild": "electron-rebuild -f -w better-sqlite3,keytar"
  }
}
```

### Native Module Usage Pattern

```typescript
// src/main/services/database.ts
import Database from "better-sqlite3";
import { app } from "electron";
import { join } from "path";

export class DatabaseService {
  private db: Database.Database;

  constructor() {
    const dbPath = join(app.getPath("userData"), "app.db");
    this.db = new Database(dbPath, {
      verbose: process.env.NODE_ENV === "development" ? console.log : undefined,
    });

    // Enable WAL mode for better concurrency
    this.db.pragma("journal_mode = WAL");

    // Initialize schema
    this.initializeSchema();
  }

  private initializeSchema(): void {
    this.db.exec(`
      CREATE TABLE IF NOT EXISTS settings (
        key TEXT PRIMARY KEY,
        value TEXT NOT NULL,
        updated_at INTEGER DEFAULT (strftime('%s', 'now'))
      );

      CREATE TABLE IF NOT EXISTS documents (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        content TEXT,
        created_at INTEGER DEFAULT (strftime('%s', 'now')),
        updated_at INTEGER DEFAULT (strftime('%s', 'now'))
      );
    `);
  }

  getSetting<T>(key: string): T | undefined {
    const row = this.db
      .prepare("SELECT value FROM settings WHERE key = ?")
      .get(key) as { value: string } | undefined;
    return row ? JSON.parse(row.value) : undefined;
  }

  setSetting<T>(key: string, value: T): void {
    this.db
      .prepare(
        "INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, strftime('%s', 'now'))"
      )
      .run(key, JSON.stringify(value));
  }

  close(): void {
    this.db.close();
  }
}
```

---

## Troubleshooting

### Common Issues

Issue: "Electron could not be found" error
Symptoms: App fails to start, module not found
Solution:

- Ensure electron is in devDependencies
- Run npm rebuild electron
- Check NODE_ENV is set correctly

Issue: White screen on launch
Symptoms: Window opens but content doesn't load
Solution:

- Check preload script path is correct
- Verify loadFile/loadURL path
- Enable devTools to see console errors
- Check for CSP blocking scripts

Issue: IPC not working
Symptoms: invoke returns undefined, no response
Solution:

- Verify channel names match exactly
- Check handler is registered before window loads
- Ensure contextBridge is used correctly
- Verify preload script is loaded

Issue: Native modules fail to load
Symptoms: "Module was compiled against different Node.js version"
Solution:

- Run electron-rebuild after npm install
- Match Electron Node.js version
- Use postinstall script for automatic rebuild

Issue: Auto-update not working
Symptoms: No update notification, silent failure
Solution:

- Check app is signed (required for updates)
- Verify publish configuration
- Check network/firewall settings
- Enable electron-updater logging

---

## External Resources

### Official Documentation

- Electron Documentation: https://www.electronjs.org/docs
- Electron Forge: https://www.electronforge.io/
- Electron Builder: https://www.electron.build/

### Security

- Security Checklist: https://www.electronjs.org/docs/tutorial/security
- Context Isolation: https://www.electronjs.org/docs/tutorial/context-isolation

### Build & Distribution

- Code Signing: https://www.electronjs.org/docs/tutorial/code-signing
- Auto Updates: https://www.electronjs.org/docs/tutorial/updates
- macOS Notarization: https://www.electronjs.org/docs/tutorial/mac-app-store-submission-guide

### Testing

- Playwright: https://playwright.dev/
- Testing Guide: https://www.electronjs.org/docs/tutorial/testing-on-headless-ci

---

Version: 1.1.0
Last Updated: 2026-01-10
Changes: Added Utility Process patterns, Protocol Handlers, Deep Linking, OWASP-aligned Security Hardening, Crash Reporting, Native Module Integration
