# Electron Framework Examples

Production-ready code examples for Electron 33+ desktop application development.

---

## Complete Electron App Setup with Vite

### Package Configuration

```json
// package.json
{
  "name": "electron-app",
  "version": "1.0.0",
  "main": "dist/main/index.js",
  "scripts": {
    "dev": "electron-vite dev",
    "build": "electron-vite build",
    "preview": "electron-vite preview",
    "package": "electron-builder",
    "package:mac": "electron-builder --mac",
    "package:win": "electron-builder --win",
    "package:linux": "electron-builder --linux",
    "postinstall": "electron-builder install-app-deps"
  },
  "dependencies": {
    "electron-store": "^8.1.0",
    "electron-updater": "^6.1.7"
  },
  "devDependencies": {
    "@electron-toolkit/preload": "^3.0.0",
    "@electron-toolkit/utils": "^3.0.0",
    "@types/node": "^20.10.0",
    "@vitejs/plugin-react": "^4.2.0",
    "electron": "^33.0.0",
    "electron-builder": "^24.9.1",
    "electron-vite": "^2.0.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

### Electron Vite Configuration

```typescript
// electron.vite.config.ts
import { defineConfig, externalizeDepsPlugin } from "electron-vite";
import react from "@vitejs/plugin-react";
import { resolve } from "path";

export default defineConfig({
  main: {
    plugins: [externalizeDepsPlugin()],
    build: {
      rollupOptions: {
        input: {
          index: resolve(__dirname, "src/main/index.ts"),
        },
      },
    },
  },
  preload: {
    plugins: [externalizeDepsPlugin()],
    build: {
      rollupOptions: {
        input: {
          index: resolve(__dirname, "src/preload/index.ts"),
        },
      },
    },
  },
  renderer: {
    root: resolve(__dirname, "src/renderer"),
    plugins: [react()],
    build: {
      rollupOptions: {
        input: {
          index: resolve(__dirname, "src/renderer/index.html"),
        },
      },
    },
  },
});
```

### Main Process Entry Point

```typescript
// src/main/index.ts
import { app, BrowserWindow, ipcMain, session, shell } from "electron";
import { join } from "path";
import { electronApp, optimizer, is } from "@electron-toolkit/utils";
import { registerIpcHandlers } from "./ipc";
import { UpdateService } from "./services/updater";
import { WindowManager } from "./windows/window-manager";

const windowManager = new WindowManager();
const updateService = new UpdateService();

async function createMainWindow(): Promise<BrowserWindow> {
  const mainWindow = windowManager.createWindow("main", {
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    titleBarStyle: "hiddenInset",
    trafficLightPosition: { x: 15, y: 15 },
    webPreferences: {
      preload: join(__dirname, "../preload/index.js"),
      sandbox: true,
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  // Open external links in default browser
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    shell.openExternal(url);
    return { action: "deny" };
  });

  // Load app
  if (is.dev && process.env["ELECTRON_RENDERER_URL"]) {
    mainWindow.loadURL(process.env["ELECTRON_RENDERER_URL"]);
    mainWindow.webContents.openDevTools({ mode: "detach" });
  } else {
    mainWindow.loadFile(join(__dirname, "../renderer/index.html"));
  }

  return mainWindow;
}

app.whenReady().then(async () => {
  // Set app user model ID for Windows
  electronApp.setAppUserModelId("com.example.myapp");

  // Watch for shortcuts to optimize new windows
  app.on("browser-window-created", (_, window) => {
    optimizer.watchWindowShortcuts(window);
  });

  // Configure session security
  configureSession();

  // Register IPC handlers
  registerIpcHandlers();

  // Create main window
  const mainWindow = await createMainWindow();

  // Initialize auto-updater
  if (!is.dev) {
    updateService.initialize(mainWindow);
    updateService.checkForUpdates();
  }

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createMainWindow();
    }
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});

function configureSession(): void {
  // Content Security Policy
  session.defaultSession.webRequest.onHeadersReceived((details, callback) => {
    callback({
      responseHeaders: {
        ...details.responseHeaders,
        "Content-Security-Policy": [
          "default-src 'self'; " +
            "script-src 'self'; " +
            "style-src 'self' 'unsafe-inline'; " +
            "img-src 'self' data: https:; " +
            "font-src 'self' data:; " +
            "connect-src 'self' https://api.github.com",
        ],
      },
    });
  });

  // Permission handler
  session.defaultSession.setPermissionRequestHandler(
    (webContents, permission, callback) => {
      const allowedPermissions = ["notifications", "clipboard-read"];
      callback(allowedPermissions.includes(permission));
    },
  );

  // Block navigation to external URLs
  session.defaultSession.setPermissionCheckHandler(() => false);
}

// Single instance lock
const gotSingleLock = app.requestSingleInstanceLock();
if (!gotSingleLock) {
  app.quit();
} else {
  app.on("second-instance", (_event, _commandLine, _workingDirectory) => {
    const mainWindow = windowManager.getWindow("main");
    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore();
      mainWindow.focus();
    }
  });
}
```

---

## Type-Safe IPC Implementation

### Shared Type Definitions

```typescript
// src/shared/types/ipc.ts
export interface FileInfo {
  path: string;
  content: string;
  encoding?: BufferEncoding;
}

export interface SaveResult {
  success: boolean;
  path: string;
  error?: string;
}

export interface DialogOptions {
  title?: string;
  filters?: { name: string; extensions: string[] }[];
  defaultPath?: string;
}

export interface StorageItem<T = unknown> {
  key: string;
  value: T;
  timestamp?: number;
}

// IPC Channel Definitions
export interface IpcMainToRenderer {
  "app:update-available": { version: string; releaseNotes: string };
  "app:update-progress": {
    percent: number;
    transferred: number;
    total: number;
  };
  "app:update-downloaded": { version: string };
  "window:maximize-change": boolean;
  "file:external-open": { path: string };
}

export interface IpcRendererToMain {
  // File operations
  "file:open-dialog": DialogOptions;
  "file:save-dialog": DialogOptions;
  "file:read": string;
  "file:write": { path: string; content: string };
  "file:exists": string;

  // Window operations
  "window:minimize": void;
  "window:maximize": void;
  "window:close": void;
  "window:is-maximized": void;

  // Storage operations
  "storage:get": string;
  "storage:set": StorageItem;
  "storage:delete": string;
  "storage:clear": void;

  // App operations
  "app:get-version": void;
  "app:get-path": "home" | "appData" | "userData" | "temp" | "downloads";
  "app:open-external": string;

  // Update operations
  "update:check": void;
  "update:download": void;
  "update:install": void;
}

// Return types for IPC handlers
export interface IpcReturnTypes {
  "file:open-dialog": FileInfo | null;
  "file:save-dialog": string | null;
  "file:read": string;
  "file:write": SaveResult;
  "file:exists": boolean;
  "window:minimize": void;
  "window:maximize": void;
  "window:close": void;
  "window:is-maximized": boolean;
  "storage:get": unknown;
  "storage:set": void;
  "storage:delete": void;
  "storage:clear": void;
  "app:get-version": string;
  "app:get-path": string;
  "app:open-external": void;
  "update:check": void;
  "update:download": void;
  "update:install": void;
}
```

### Main Process IPC Handlers

```typescript
// src/main/ipc/index.ts
import { ipcMain, dialog, app, shell, BrowserWindow } from "electron";
import { readFile, writeFile, access } from "fs/promises";
import { constants } from "fs";
import Store from "electron-store";
import { z } from "zod";
import type {
  DialogOptions,
  FileInfo,
  SaveResult,
  StorageItem,
} from "../../shared/types/ipc";

const store = new Store({
  encryptionKey: process.env.STORE_ENCRYPTION_KEY,
});

// Validation schemas
const FilePathSchema = z
  .string()
  .min(1)
  .refine(
    (path) => {
      const normalized = path.replace(/\\/g, "/");
      return !normalized.includes("..") && !normalized.includes("\0");
    },
    { message: "Invalid file path" },
  );

const StorageKeySchema = z
  .string()
  .min(1)
  .max(256)
  .regex(/^[a-zA-Z0-9_.-]+$/);

export function registerIpcHandlers(): void {
  // File operations
  ipcMain.handle(
    "file:open-dialog",
    async (_event, options: DialogOptions): Promise<FileInfo | null> => {
      const result = await dialog.showOpenDialog({
        title: options.title ?? "Open File",
        properties: ["openFile"],
        filters: options.filters ?? [{ name: "All Files", extensions: ["*"] }],
        defaultPath: options.defaultPath,
      });

      if (result.canceled || result.filePaths.length === 0) {
        return null;
      }

      const path = result.filePaths[0];
      const content = await readFile(path, "utf-8");
      return { path, content };
    },
  );

  ipcMain.handle(
    "file:save-dialog",
    async (_event, options: DialogOptions): Promise<string | null> => {
      const result = await dialog.showSaveDialog({
        title: options.title ?? "Save File",
        filters: options.filters ?? [{ name: "All Files", extensions: ["*"] }],
        defaultPath: options.defaultPath,
      });

      return result.canceled ? null : (result.filePath ?? null);
    },
  );

  ipcMain.handle("file:read", async (_event, path: string): Promise<string> => {
    const validPath = FilePathSchema.parse(path);
    return readFile(validPath, "utf-8");
  });

  ipcMain.handle(
    "file:write",
    async (
      _event,
      { path, content }: { path: string; content: string },
    ): Promise<SaveResult> => {
      try {
        const validPath = FilePathSchema.parse(path);
        await writeFile(validPath, content, "utf-8");
        return { success: true, path: validPath };
      } catch (error) {
        return {
          success: false,
          path,
          error: error instanceof Error ? error.message : "Unknown error",
        };
      }
    },
  );

  ipcMain.handle(
    "file:exists",
    async (_event, path: string): Promise<boolean> => {
      try {
        const validPath = FilePathSchema.parse(path);
        await access(validPath, constants.F_OK);
        return true;
      } catch {
        return false;
      }
    },
  );

  // Window operations
  ipcMain.handle("window:minimize", (event): void => {
    BrowserWindow.fromWebContents(event.sender)?.minimize();
  });

  ipcMain.handle("window:maximize", (event): void => {
    const window = BrowserWindow.fromWebContents(event.sender);
    if (window?.isMaximized()) {
      window.unmaximize();
    } else {
      window?.maximize();
    }
  });

  ipcMain.handle("window:close", (event): void => {
    BrowserWindow.fromWebContents(event.sender)?.close();
  });

  ipcMain.handle("window:is-maximized", (event): boolean => {
    return BrowserWindow.fromWebContents(event.sender)?.isMaximized() ?? false;
  });

  // Storage operations
  ipcMain.handle("storage:get", (_event, key: string): unknown => {
    const validKey = StorageKeySchema.parse(key);
    return store.get(validKey);
  });

  ipcMain.handle("storage:set", (_event, { key, value }: StorageItem): void => {
    const validKey = StorageKeySchema.parse(key);
    store.set(validKey, value);
  });

  ipcMain.handle("storage:delete", (_event, key: string): void => {
    const validKey = StorageKeySchema.parse(key);
    store.delete(validKey);
  });

  ipcMain.handle("storage:clear", (): void => {
    store.clear();
  });

  // App operations
  ipcMain.handle("app:get-version", (): string => {
    return app.getVersion();
  });

  ipcMain.handle(
    "app:get-path",
    (
      _event,
      name: "home" | "appData" | "userData" | "temp" | "downloads",
    ): string => {
      return app.getPath(name);
    },
  );

  ipcMain.handle(
    "app:open-external",
    async (_event, url: string): Promise<void> => {
      // Validate URL before opening
      const parsedUrl = new URL(url);
      if (parsedUrl.protocol === "https:" || parsedUrl.protocol === "http:") {
        await shell.openExternal(url);
      }
    },
  );
}
```

### Preload Script with Full API

```typescript
// src/preload/index.ts
import { contextBridge, ipcRenderer, IpcRendererEvent } from "electron";
import type {
  DialogOptions,
  FileInfo,
  SaveResult,
  StorageItem,
  IpcMainToRenderer,
} from "../shared/types/ipc";

// Type-safe event listener helper
type EventCallback<T> = (data: T) => void;
type Unsubscribe = () => void;

function createEventListener<K extends keyof IpcMainToRenderer>(
  channel: K,
  callback: EventCallback<IpcMainToRenderer[K]>,
): Unsubscribe {
  const handler = (_event: IpcRendererEvent, data: IpcMainToRenderer[K]) => {
    callback(data);
  };
  ipcRenderer.on(channel, handler);
  return () => ipcRenderer.removeListener(channel, handler);
}

const electronAPI = {
  // File operations
  file: {
    openDialog: (options: DialogOptions = {}): Promise<FileInfo | null> =>
      ipcRenderer.invoke("file:open-dialog", options),
    saveDialog: (options: DialogOptions = {}): Promise<string | null> =>
      ipcRenderer.invoke("file:save-dialog", options),
    read: (path: string): Promise<string> =>
      ipcRenderer.invoke("file:read", path),
    write: (path: string, content: string): Promise<SaveResult> =>
      ipcRenderer.invoke("file:write", { path, content }),
    exists: (path: string): Promise<boolean> =>
      ipcRenderer.invoke("file:exists", path),
  },

  // Window operations
  window: {
    minimize: (): Promise<void> => ipcRenderer.invoke("window:minimize"),
    maximize: (): Promise<void> => ipcRenderer.invoke("window:maximize"),
    close: (): Promise<void> => ipcRenderer.invoke("window:close"),
    isMaximized: (): Promise<boolean> =>
      ipcRenderer.invoke("window:is-maximized"),
    onMaximizeChange: (callback: (isMaximized: boolean) => void): Unsubscribe =>
      createEventListener("window:maximize-change", callback),
  },

  // Storage operations
  storage: {
    get: <T = unknown>(key: string): Promise<T | undefined> =>
      ipcRenderer.invoke("storage:get", key) as Promise<T | undefined>,
    set: <T = unknown>(key: string, value: T): Promise<void> =>
      ipcRenderer.invoke("storage:set", { key, value } as StorageItem<T>),
    delete: (key: string): Promise<void> =>
      ipcRenderer.invoke("storage:delete", key),
    clear: (): Promise<void> => ipcRenderer.invoke("storage:clear"),
  },

  // App operations
  app: {
    getVersion: (): Promise<string> => ipcRenderer.invoke("app:get-version"),
    getPath: (
      name: "home" | "appData" | "userData" | "temp" | "downloads",
    ): Promise<string> => ipcRenderer.invoke("app:get-path", name),
    openExternal: (url: string): Promise<void> =>
      ipcRenderer.invoke("app:open-external", url),
  },

  // Update operations
  update: {
    check: (): Promise<void> => ipcRenderer.invoke("update:check"),
    download: (): Promise<void> => ipcRenderer.invoke("update:download"),
    install: (): Promise<void> => ipcRenderer.invoke("update:install"),
    onAvailable: (
      callback: (info: { version: string; releaseNotes: string }) => void,
    ): Unsubscribe => createEventListener("app:update-available", callback),
    onProgress: (
      callback: (progress: {
        percent: number;
        transferred: number;
        total: number;
      }) => void,
    ): Unsubscribe => createEventListener("app:update-progress", callback),
    onDownloaded: (
      callback: (info: { version: string }) => void,
    ): Unsubscribe => createEventListener("app:update-downloaded", callback),
  },

  // Platform info
  platform: {
    isMac: process.platform === "darwin",
    isWindows: process.platform === "win32",
    isLinux: process.platform === "linux",
  },
};

contextBridge.exposeInMainWorld("electronAPI", electronAPI);

// Type declaration for renderer
export type ElectronAPI = typeof electronAPI;

declare global {
  interface Window {
    electronAPI: ElectronAPI;
  }
}
```

---

## Auto-Update Integration

### Complete Update Service

```typescript
// src/main/services/updater.ts
import {
  autoUpdater,
  UpdateInfo,
  ProgressInfo,
  UpdateDownloadedEvent,
} from "electron-updater";
import { BrowserWindow, dialog, Notification } from "electron";
import log from "electron-log";

export interface UpdateServiceOptions {
  /** Check for updates on startup */
  checkOnStartup?: boolean;
  /** Auto-download updates */
  autoDownload?: boolean;
  /** Auto-install on quit */
  autoInstallOnAppQuit?: boolean;
  /** Check interval in milliseconds (default: 1 hour) */
  checkInterval?: number;
}

export class UpdateService {
  private mainWindow: BrowserWindow | null = null;
  private checkIntervalId: NodeJS.Timeout | null = null;

  constructor(
    private options: UpdateServiceOptions = {
      checkOnStartup: true,
      autoDownload: false,
      autoInstallOnAppQuit: true,
      checkInterval: 3600000,
    },
  ) {
    // Configure logging
    autoUpdater.logger = log;
    log.transports.file.level = "info";
  }

  initialize(window: BrowserWindow): void {
    this.mainWindow = window;

    // Configure auto-updater
    autoUpdater.autoDownload = this.options.autoDownload ?? false;
    autoUpdater.autoInstallOnAppQuit =
      this.options.autoInstallOnAppQuit ?? true;

    // Set up event handlers
    this.setupEventHandlers();

    // Check on startup if enabled
    if (this.options.checkOnStartup) {
      // Delay initial check to let app settle
      setTimeout(() => this.checkForUpdates(), 5000);
    }

    // Set up periodic checking
    if (this.options.checkInterval && this.options.checkInterval > 0) {
      this.checkIntervalId = setInterval(
        () => this.checkForUpdates(),
        this.options.checkInterval,
      );
    }
  }

  private setupEventHandlers(): void {
    autoUpdater.on("checking-for-update", () => {
      log.info("Checking for updates...");
    });

    autoUpdater.on("update-available", (info: UpdateInfo) => {
      log.info("Update available:", info.version);
      this.notifyUpdateAvailable(info);
    });

    autoUpdater.on("update-not-available", () => {
      log.info("No updates available");
    });

    autoUpdater.on("error", (error: Error) => {
      log.error("Update error:", error.message);
      this.notifyError(error);
    });

    autoUpdater.on("download-progress", (progress: ProgressInfo) => {
      log.info(`Download progress: ${progress.percent.toFixed(1)}%`);
      this.mainWindow?.webContents.send("app:update-progress", {
        percent: progress.percent,
        transferred: progress.transferred,
        total: progress.total,
      });
    });

    autoUpdater.on("update-downloaded", (event: UpdateDownloadedEvent) => {
      log.info("Update downloaded:", event.version);
      this.notifyUpdateDownloaded(event);
    });
  }

  async checkForUpdates(): Promise<void> {
    try {
      await autoUpdater.checkForUpdates();
    } catch (error) {
      log.error("Failed to check for updates:", error);
    }
  }

  async downloadUpdate(): Promise<void> {
    try {
      await autoUpdater.downloadUpdate();
    } catch (error) {
      log.error("Failed to download update:", error);
    }
  }

  installUpdate(): void {
    autoUpdater.quitAndInstall(false, true);
  }

  private async notifyUpdateAvailable(info: UpdateInfo): Promise<void> {
    // Send to renderer
    this.mainWindow?.webContents.send("app:update-available", {
      version: info.version,
      releaseNotes:
        typeof info.releaseNotes === "string"
          ? info.releaseNotes
          : (info.releaseNotes
              ?.map((n) => `${n.version}: ${n.note}`)
              .join("\n") ?? ""),
    });

    // Show system notification if supported
    if (Notification.isSupported()) {
      new Notification({
        title: "Update Available",
        body: `Version ${info.version} is available for download.`,
      }).show();
    }

    // Show dialog
    const result = await dialog.showMessageBox(this.mainWindow!, {
      type: "info",
      title: "Update Available",
      message: `A new version (${info.version}) is available.`,
      detail: "Would you like to download and install it?",
      buttons: ["Download", "Later"],
      defaultId: 0,
      cancelId: 1,
    });

    if (result.response === 0) {
      this.downloadUpdate();
    }
  }

  private async notifyUpdateDownloaded(
    event: UpdateDownloadedEvent,
  ): Promise<void> {
    // Send to renderer
    this.mainWindow?.webContents.send("app:update-downloaded", {
      version: event.version,
    });

    // Show dialog
    const result = await dialog.showMessageBox(this.mainWindow!, {
      type: "info",
      title: "Update Ready",
      message: `Version ${event.version} has been downloaded.`,
      detail: "Would you like to restart and install it now?",
      buttons: ["Restart Now", "Later"],
      defaultId: 0,
      cancelId: 1,
    });

    if (result.response === 0) {
      this.installUpdate();
    }
  }

  private notifyError(error: Error): void {
    dialog.showErrorBox(
      "Update Error",
      `An error occurred while updating: ${error.message}`,
    );
  }

  dispose(): void {
    if (this.checkIntervalId) {
      clearInterval(this.checkIntervalId);
      this.checkIntervalId = null;
    }
  }
}
```

---

## System Tray and Native Menu

### System Tray Service

```typescript
// src/main/services/tray.ts
import {
  Tray,
  Menu,
  MenuItemConstructorOptions,
  app,
  nativeImage,
  BrowserWindow,
} from "electron";
import { join } from "path";
import { is } from "@electron-toolkit/utils";

export class TrayService {
  private tray: Tray | null = null;
  private mainWindow: BrowserWindow | null = null;

  initialize(mainWindow: BrowserWindow): void {
    this.mainWindow = mainWindow;

    // Create tray icon
    const iconPath = is.dev
      ? join(__dirname, "../../resources/icons/tray.png")
      : join(process.resourcesPath, "icons/tray.png");

    // Use template icon on macOS
    const icon = nativeImage.createFromPath(iconPath);
    if (process.platform === "darwin") {
      icon.setTemplateImage(true);
    }

    this.tray = new Tray(icon);
    this.tray.setToolTip(app.getName());

    // Set context menu
    this.updateContextMenu();

    // Click behavior
    this.tray.on("click", () => {
      this.toggleMainWindow();
    });

    // Double-click to show (Windows)
    this.tray.on("double-click", () => {
      this.showMainWindow();
    });
  }

  updateContextMenu(additionalItems: MenuItemConstructorOptions[] = []): void {
    if (!this.tray) return;

    const contextMenu = Menu.buildFromTemplate([
      {
        label: "Show App",
        click: () => this.showMainWindow(),
      },
      { type: "separator" },
      ...additionalItems,
      { type: "separator" },
      {
        label: "Preferences",
        accelerator: "CmdOrCtrl+,",
        click: () => this.openPreferences(),
      },
      { type: "separator" },
      {
        label: "Check for Updates",
        click: () => this.checkForUpdates(),
      },
      { type: "separator" },
      {
        label: "Quit",
        accelerator: "CmdOrCtrl+Q",
        click: () => app.quit(),
      },
    ]);

    this.tray.setContextMenu(contextMenu);
  }

  private toggleMainWindow(): void {
    if (!this.mainWindow) return;

    if (this.mainWindow.isVisible()) {
      if (this.mainWindow.isFocused()) {
        this.mainWindow.hide();
      } else {
        this.mainWindow.focus();
      }
    } else {
      this.showMainWindow();
    }
  }

  private showMainWindow(): void {
    if (!this.mainWindow) return;

    this.mainWindow.show();
    this.mainWindow.focus();

    // Restore if minimized
    if (this.mainWindow.isMinimized()) {
      this.mainWindow.restore();
    }
  }

  private openPreferences(): void {
    // Emit event or open preferences window
    this.mainWindow?.webContents.send("app:open-preferences");
  }

  private checkForUpdates(): void {
    this.mainWindow?.webContents.send("app:check-updates");
  }

  setBadge(text: string): void {
    if (process.platform === "darwin") {
      app.dock.setBadge(text);
    } else if (this.tray) {
      // Update tray title/tooltip for other platforms
      this.tray.setTitle(text);
    }
  }

  destroy(): void {
    this.tray?.destroy();
    this.tray = null;
  }
}
```

### Application Menu

```typescript
// src/main/services/menu.ts
import {
  Menu,
  app,
  shell,
  BrowserWindow,
  MenuItemConstructorOptions,
} from "electron";

export function createApplicationMenu(mainWindow: BrowserWindow): void {
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
              {
                label: "Preferences",
                accelerator: "CmdOrCtrl+,",
                click: () =>
                  mainWindow.webContents.send("app:open-preferences"),
              },
              { type: "separator" as const },
              { role: "services" as const },
              { type: "separator" as const },
              { role: "hide" as const },
              { role: "hideOthers" as const },
              { role: "unhide" as const },
              { type: "separator" as const },
              { role: "quit" as const },
            ],
          } as MenuItemConstructorOptions,
        ]
      : []),

    // File menu
    {
      label: "File",
      submenu: [
        {
          label: "New",
          accelerator: "CmdOrCtrl+N",
          click: () => mainWindow.webContents.send("file:new"),
        },
        {
          label: "Open...",
          accelerator: "CmdOrCtrl+O",
          click: () => mainWindow.webContents.send("file:open"),
        },
        { type: "separator" },
        {
          label: "Save",
          accelerator: "CmdOrCtrl+S",
          click: () => mainWindow.webContents.send("file:save"),
        },
        {
          label: "Save As...",
          accelerator: "CmdOrCtrl+Shift+S",
          click: () => mainWindow.webContents.send("file:save-as"),
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
        { role: "delete" },
        { type: "separator" },
        { role: "selectAll" },
        ...(isMac
          ? [
              { type: "separator" as const },
              {
                label: "Speech",
                submenu: [
                  { role: "startSpeaking" as const },
                  { role: "stopSpeaking" as const },
                ],
              },
            ]
          : []),
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

    // Window menu
    {
      label: "Window",
      submenu: [
        { role: "minimize" },
        { role: "zoom" },
        ...(isMac
          ? [
              { type: "separator" as const },
              { role: "front" as const },
              { type: "separator" as const },
              { role: "window" as const },
            ]
          : [{ role: "close" as const }]),
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
          label: "Release Notes",
          click: () =>
            shell.openExternal("https://github.com/example/repo/releases"),
        },
        { type: "separator" },
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

## Window State Persistence

### Window Manager with State

```typescript
// src/main/windows/window-manager.ts
import {
  BrowserWindow,
  BrowserWindowConstructorOptions,
  screen,
  Rectangle,
} from "electron";
import Store from "electron-store";
import { join } from "path";

interface WindowState {
  width: number;
  height: number;
  x: number | undefined;
  y: number | undefined;
  isMaximized: boolean;
  isFullScreen: boolean;
}

interface WindowConfig {
  id: string;
  defaultWidth: number;
  defaultHeight: number;
  minWidth?: number;
  minHeight?: number;
}

const windowStateStore = new Store<Record<string, WindowState>>({
  name: "window-state",
});

export class WindowManager {
  private windows = new Map<string, BrowserWindow>();
  private stateUpdateDebounce = new Map<string, NodeJS.Timeout>();

  createWindow(
    id: string,
    options: BrowserWindowConstructorOptions = {},
  ): BrowserWindow {
    // Get saved state or calculate default
    const savedState = this.getWindowState(id);
    const { width, height } = screen.getPrimaryDisplay().workAreaSize;

    const defaultConfig: WindowConfig = {
      id,
      defaultWidth: Math.floor(width * 0.8),
      defaultHeight: Math.floor(height * 0.8),
      minWidth: options.minWidth ?? 400,
      minHeight: options.minHeight ?? 300,
    };

    // Calculate initial bounds
    const bounds = this.calculateBounds(savedState, defaultConfig);

    const window = new BrowserWindow({
      ...bounds,
      show: false,
      ...options,
      webPreferences: {
        preload: join(__dirname, "../preload/index.js"),
        sandbox: true,
        contextIsolation: true,
        nodeIntegration: false,
        ...options.webPreferences,
      },
    });

    // Restore maximized/fullscreen state
    if (savedState?.isMaximized) {
      window.maximize();
    }
    if (savedState?.isFullScreen) {
      window.setFullScreen(true);
    }

    // Show when ready
    window.once("ready-to-show", () => {
      window.show();
    });

    // Track state changes
    this.trackWindowState(id, window);

    // Handle close
    window.on("closed", () => {
      this.windows.delete(id);
      const timeout = this.stateUpdateDebounce.get(id);
      if (timeout) {
        clearTimeout(timeout);
        this.stateUpdateDebounce.delete(id);
      }
    });

    this.windows.set(id, window);
    return window;
  }

  getWindow(id: string): BrowserWindow | undefined {
    const window = this.windows.get(id);
    if (window && !window.isDestroyed()) {
      return window;
    }
    return undefined;
  }

  getAllWindows(): BrowserWindow[] {
    return Array.from(this.windows.values()).filter((w) => !w.isDestroyed());
  }

  closeWindow(id: string): void {
    const window = this.getWindow(id);
    if (window) {
      window.close();
    }
  }

  closeAll(): void {
    for (const window of this.getAllWindows()) {
      window.close();
    }
  }

  private getWindowState(id: string): WindowState | undefined {
    return windowStateStore.get(id);
  }

  private saveWindowState(id: string, state: WindowState): void {
    windowStateStore.set(id, state);
  }

  private calculateBounds(
    savedState: WindowState | undefined,
    config: WindowConfig,
  ): Rectangle {
    const { width, height, x, y } = screen.getPrimaryDisplay().workAreaSize;

    // Use saved state if available and valid
    if (savedState && this.isValidBounds(savedState)) {
      return {
        width: savedState.width,
        height: savedState.height,
        x: savedState.x ?? Math.floor((width - savedState.width) / 2),
        y: savedState.y ?? Math.floor((height - savedState.height) / 2),
      };
    }

    // Center window with default size
    return {
      width: config.defaultWidth,
      height: config.defaultHeight,
      x: Math.floor((width - config.defaultWidth) / 2),
      y: Math.floor((height - config.defaultHeight) / 2),
    };
  }

  private isValidBounds(state: WindowState): boolean {
    const displays = screen.getAllDisplays();

    // Check if window is visible on any display
    return displays.some((display) => {
      const { x, y, width, height } = display.bounds;
      const windowX = state.x ?? 0;
      const windowY = state.y ?? 0;

      return (
        windowX >= x - state.width &&
        windowX <= x + width &&
        windowY >= y - state.height &&
        windowY <= y + height
      );
    });
  }

  private trackWindowState(id: string, window: BrowserWindow): void {
    const saveState = (): void => {
      // Debounce state updates
      const existing = this.stateUpdateDebounce.get(id);
      if (existing) {
        clearTimeout(existing);
      }

      this.stateUpdateDebounce.set(
        id,
        setTimeout(() => {
          if (window.isDestroyed()) return;

          const bounds = window.getBounds();
          this.saveWindowState(id, {
            width: bounds.width,
            height: bounds.height,
            x: bounds.x,
            y: bounds.y,
            isMaximized: window.isMaximized(),
            isFullScreen: window.isFullScreen(),
          });
        }, 500),
      );
    };

    window.on("resize", saveState);
    window.on("move", saveState);
    window.on("maximize", saveState);
    window.on("unmaximize", saveState);
    window.on("enter-full-screen", saveState);
    window.on("leave-full-screen", saveState);

    // Save on close
    window.on("close", () => {
      if (!window.isDestroyed()) {
        const bounds = window.getBounds();
        this.saveWindowState(id, {
          width: bounds.width,
          height: bounds.height,
          x: bounds.x,
          y: bounds.y,
          isMaximized: window.isMaximized(),
          isFullScreen: window.isFullScreen(),
        });
      }
    });
  }
}

export const windowManager = new WindowManager();
```

---

## Secure File Operations

### File Service with Validation

```typescript
// src/main/services/file-service.ts
import {
  readFile,
  writeFile,
  access,
  mkdir,
  stat,
  readdir,
  unlink,
  rename,
} from "fs/promises";
import { constants, createReadStream, createWriteStream } from "fs";
import { join, dirname, basename, extname, normalize, resolve } from "path";
import { app } from "electron";
import { z } from "zod";
import { pipeline } from "stream/promises";
import { createHash } from "crypto";

// Validation schemas
const SafePathSchema = z
  .string()
  .min(1)
  .max(4096)
  .refine(
    (path) => {
      const normalized = normalize(path);
      // Prevent path traversal
      return !normalized.includes("..") && !normalized.includes("\0");
    },
    { message: "Invalid path: potential path traversal detected" },
  );

const FileNameSchema = z
  .string()
  .min(1)
  .max(255)
  .regex(/^[^<>:"/\\|?*\x00-\x1f]+$/, {
    message: "Invalid filename: contains reserved characters",
  });

export interface FileResult<T = unknown> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface FileMetadata {
  name: string;
  path: string;
  size: number;
  isDirectory: boolean;
  created: Date;
  modified: Date;
  extension: string;
}

export class FileService {
  private allowedDirectories: string[];

  constructor() {
    // Define allowed directories for file operations
    this.allowedDirectories = [
      app.getPath("documents"),
      app.getPath("downloads"),
      app.getPath("userData"),
      app.getPath("temp"),
    ];
  }

  async read(filePath: string): Promise<FileResult<string>> {
    try {
      const validPath = this.validatePath(filePath);
      const content = await readFile(validPath, "utf-8");
      return { success: true, data: content };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async readBinary(filePath: string): Promise<FileResult<Buffer>> {
    try {
      const validPath = this.validatePath(filePath);
      const content = await readFile(validPath);
      return { success: true, data: content };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async write(
    filePath: string,
    content: string | Buffer,
  ): Promise<FileResult<void>> {
    try {
      const validPath = this.validatePath(filePath);

      // Ensure directory exists
      await mkdir(dirname(validPath), { recursive: true });

      await writeFile(validPath, content);
      return { success: true };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async exists(filePath: string): Promise<boolean> {
    try {
      const validPath = this.validatePath(filePath);
      await access(validPath, constants.F_OK);
      return true;
    } catch {
      return false;
    }
  }

  async getMetadata(filePath: string): Promise<FileResult<FileMetadata>> {
    try {
      const validPath = this.validatePath(filePath);
      const stats = await stat(validPath);

      return {
        success: true,
        data: {
          name: basename(validPath),
          path: validPath,
          size: stats.size,
          isDirectory: stats.isDirectory(),
          created: stats.birthtime,
          modified: stats.mtime,
          extension: extname(validPath),
        },
      };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async listDirectory(dirPath: string): Promise<FileResult<FileMetadata[]>> {
    try {
      const validPath = this.validatePath(dirPath);
      const entries = await readdir(validPath, { withFileTypes: true });

      const metadata: FileMetadata[] = await Promise.all(
        entries.map(async (entry) => {
          const entryPath = join(validPath, entry.name);
          const stats = await stat(entryPath);

          return {
            name: entry.name,
            path: entryPath,
            size: stats.size,
            isDirectory: entry.isDirectory(),
            created: stats.birthtime,
            modified: stats.mtime,
            extension: entry.isDirectory() ? "" : extname(entry.name),
          };
        }),
      );

      return { success: true, data: metadata };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async delete(filePath: string): Promise<FileResult<void>> {
    try {
      const validPath = this.validatePath(filePath);
      await unlink(validPath);
      return { success: true };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async move(sourcePath: string, destPath: string): Promise<FileResult<void>> {
    try {
      const validSource = this.validatePath(sourcePath);
      const validDest = this.validatePath(destPath);

      await mkdir(dirname(validDest), { recursive: true });
      await rename(validSource, validDest);

      return { success: true };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async copy(sourcePath: string, destPath: string): Promise<FileResult<void>> {
    try {
      const validSource = this.validatePath(sourcePath);
      const validDest = this.validatePath(destPath);

      await mkdir(dirname(validDest), { recursive: true });

      await pipeline(
        createReadStream(validSource),
        createWriteStream(validDest),
      );

      return { success: true };
    } catch (error) {
      return this.handleError(error);
    }
  }

  async getHash(
    filePath: string,
    algorithm: "md5" | "sha256" = "sha256",
  ): Promise<FileResult<string>> {
    try {
      const validPath = this.validatePath(filePath);
      const hash = createHash(algorithm);
      const stream = createReadStream(validPath);

      return new Promise((resolve) => {
        stream.on("data", (chunk) => hash.update(chunk));
        stream.on("end", () => {
          resolve({ success: true, data: hash.digest("hex") });
        });
        stream.on("error", (error) => {
          resolve(this.handleError(error));
        });
      });
    } catch (error) {
      return this.handleError(error);
    }
  }

  private validatePath(filePath: string): string {
    // Validate path format
    const safePath = SafePathSchema.parse(filePath);

    // Resolve to absolute path
    const absolutePath = resolve(safePath);

    // Validate filename if applicable
    const fileName = basename(absolutePath);
    if (fileName && !fileName.startsWith(".")) {
      FileNameSchema.parse(fileName);
    }

    // Optional: Check if path is within allowed directories
    // Uncomment if you want to restrict file access
    // this.validateAllowedDirectory(absolutePath);

    return absolutePath;
  }

  private validateAllowedDirectory(absolutePath: string): void {
    const isAllowed = this.allowedDirectories.some((dir) =>
      absolutePath.startsWith(dir),
    );

    if (!isAllowed) {
      throw new Error(`Access denied: path is outside allowed directories`);
    }
  }

  private handleError<T>(error: unknown): FileResult<T> {
    const message =
      error instanceof Error ? error.message : "Unknown error occurred";
    return { success: false, error: message };
  }
}

export const fileService = new FileService();
```

---

## React Renderer Integration

### Electron API Hook

```typescript
// src/renderer/src/hooks/useElectron.ts
import { useEffect, useState, useCallback } from "react";

// Type from preload
type ElectronAPI = Window["electronAPI"];

export function useElectron(): ElectronAPI {
  if (!window.electronAPI) {
    throw new Error("Electron API not available. Are you running in Electron?");
  }
  return window.electronAPI;
}

export function useWindowMaximized(): boolean {
  const [isMaximized, setIsMaximized] = useState(false);
  const electron = useElectron();

  useEffect(() => {
    // Get initial state
    electron.window.isMaximized().then(setIsMaximized);

    // Subscribe to changes
    const unsubscribe = electron.window.onMaximizeChange(setIsMaximized);
    return unsubscribe;
  }, [electron]);

  return isMaximized;
}

export function useAutoUpdate() {
  const electron = useElectron();
  const [updateAvailable, setUpdateAvailable] = useState<{
    version: string;
    releaseNotes: string;
  } | null>(null);
  const [downloadProgress, setDownloadProgress] = useState<{
    percent: number;
    transferred: number;
    total: number;
  } | null>(null);
  const [updateReady, setUpdateReady] = useState(false);

  useEffect(() => {
    const unsubAvailable = electron.update.onAvailable((info) => {
      setUpdateAvailable(info);
    });

    const unsubProgress = electron.update.onProgress((progress) => {
      setDownloadProgress(progress);
    });

    const unsubDownloaded = electron.update.onDownloaded(() => {
      setUpdateReady(true);
      setDownloadProgress(null);
    });

    return () => {
      unsubAvailable();
      unsubProgress();
      unsubDownloaded();
    };
  }, [electron]);

  const checkForUpdates = useCallback(() => {
    electron.update.check();
  }, [electron]);

  const downloadUpdate = useCallback(() => {
    electron.update.download();
  }, [electron]);

  const installUpdate = useCallback(() => {
    electron.update.install();
  }, [electron]);

  return {
    updateAvailable,
    downloadProgress,
    updateReady,
    checkForUpdates,
    downloadUpdate,
    installUpdate,
  };
}

export function useStorage<T>(
  key: string,
  defaultValue: T,
): [T, (value: T) => Promise<void>, boolean] {
  const electron = useElectron();
  const [value, setValue] = useState<T>(defaultValue);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    electron.storage.get<T>(key).then((stored) => {
      if (stored !== undefined) {
        setValue(stored);
      }
      setLoading(false);
    });
  }, [electron, key]);

  const updateValue = useCallback(
    async (newValue: T) => {
      setValue(newValue);
      await electron.storage.set(key, newValue);
    },
    [electron, key],
  );

  return [value, updateValue, loading];
}
```

### Custom Title Bar Component

```tsx
// src/renderer/src/components/TitleBar.tsx
import { useElectron, useWindowMaximized } from "../hooks/useElectron";
import styles from "./TitleBar.module.css";

interface TitleBarProps {
  title?: string;
}

export function TitleBar({ title = "My App" }: TitleBarProps) {
  const electron = useElectron();
  const isMaximized = useWindowMaximized();

  return (
    <div className={styles.titleBar}>
      {/* Drag region */}
      <div className={styles.dragRegion}>
        <span className={styles.title}>{title}</span>
      </div>

      {/* Window controls */}
      <div className={styles.windowControls}>
        {!electron.platform.isMac && (
          <>
            <button
              className={styles.controlButton}
              onClick={() => electron.window.minimize()}
              aria-label="Minimize"
            >
              <MinimizeIcon />
            </button>
            <button
              className={styles.controlButton}
              onClick={() => electron.window.maximize()}
              aria-label={isMaximized ? "Restore" : "Maximize"}
            >
              {isMaximized ? <RestoreIcon /> : <MaximizeIcon />}
            </button>
            <button
              className={`${styles.controlButton} ${styles.closeButton}`}
              onClick={() => electron.window.close()}
              aria-label="Close"
            >
              <CloseIcon />
            </button>
          </>
        )}
      </div>
    </div>
  );
}

// Icon components
function MinimizeIcon() {
  return (
    <svg width="10" height="1" viewBox="0 0 10 1">
      <rect fill="currentColor" width="10" height="1" />
    </svg>
  );
}

function MaximizeIcon() {
  return (
    <svg width="10" height="10" viewBox="0 0 10 10">
      <rect
        fill="none"
        stroke="currentColor"
        strokeWidth="1"
        x="0.5"
        y="0.5"
        width="9"
        height="9"
      />
    </svg>
  );
}

function RestoreIcon() {
  return (
    <svg width="10" height="10" viewBox="0 0 10 10">
      <rect
        fill="none"
        stroke="currentColor"
        strokeWidth="1"
        x="2.5"
        y="0.5"
        width="7"
        height="7"
      />
      <rect
        fill="none"
        stroke="currentColor"
        strokeWidth="1"
        x="0.5"
        y="2.5"
        width="7"
        height="7"
      />
    </svg>
  );
}

function CloseIcon() {
  return (
    <svg width="10" height="10" viewBox="0 0 10 10">
      <path
        fill="currentColor"
        d="M1 0L0 1l4 4-4 4 1 1 4-4 4 4 1-1-4-4 4-4-1-1-4 4-4-4z"
      />
    </svg>
  );
}
```

```css
/* src/renderer/src/components/TitleBar.module.css */
.titleBar {
  display: flex;
  height: 32px;
  background: var(--titlebar-bg, #2d2d2d);
  color: var(--titlebar-color, #ffffff);
  user-select: none;
}

.dragRegion {
  flex: 1;
  display: flex;
  align-items: center;
  padding-left: 12px;
  -webkit-app-region: drag;
}

.title {
  font-size: 12px;
  font-weight: 500;
}

.windowControls {
  display: flex;
  -webkit-app-region: no-drag;
}

.controlButton {
  width: 46px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  transition: background-color 0.1s;
}

.controlButton:hover {
  background: rgba(255, 255, 255, 0.1);
}

.controlButton:active {
  background: rgba(255, 255, 255, 0.2);
}

.closeButton:hover {
  background: #e81123;
}
```

---

## Testing with Playwright

### E2E Test Setup

```typescript
// e2e/electron.spec.ts
import { test, expect, _electron as electron } from "@playwright/test";
import type { ElectronApplication, Page } from "@playwright/test";

let app: ElectronApplication;
let page: Page;

test.beforeAll(async () => {
  // Launch Electron app
  app = await electron.launch({
    args: ["."],
    env: {
      ...process.env,
      NODE_ENV: "test",
    },
  });

  // Get the first window
  page = await app.firstWindow();

  // Wait for app to be ready
  await page.waitForLoadState("domcontentloaded");
});

test.afterAll(async () => {
  await app.close();
});

test.describe("Main Window", () => {
  test("should display title", async () => {
    const title = await page.title();
    expect(title).toBe("My App");
  });

  test("should have correct dimensions", async () => {
    const { width, height } = page.viewportSize()!;
    expect(width).toBeGreaterThanOrEqual(800);
    expect(height).toBeGreaterThanOrEqual(600);
  });

  test("should show main content", async () => {
    await expect(page.locator("#app")).toBeVisible();
  });
});

test.describe("Window Controls", () => {
  test("should minimize window", async () => {
    // Click minimize button
    await page.click('[aria-label="Minimize"]');

    // Verify window is minimized
    const isMinimized = await app.evaluate(({ BrowserWindow }) => {
      const window = BrowserWindow.getAllWindows()[0];
      return window.isMinimized();
    });

    expect(isMinimized).toBe(true);

    // Restore for next tests
    await app.evaluate(({ BrowserWindow }) => {
      const window = BrowserWindow.getAllWindows()[0];
      window.restore();
    });
  });

  test("should maximize/restore window", async () => {
    // Click maximize button
    await page.click('[aria-label="Maximize"]');

    // Verify window is maximized
    let isMaximized = await app.evaluate(({ BrowserWindow }) => {
      const window = BrowserWindow.getAllWindows()[0];
      return window.isMaximized();
    });

    expect(isMaximized).toBe(true);

    // Click restore button
    await page.click('[aria-label="Restore"]');

    // Verify window is not maximized
    isMaximized = await app.evaluate(({ BrowserWindow }) => {
      const window = BrowserWindow.getAllWindows()[0];
      return window.isMaximized();
    });

    expect(isMaximized).toBe(false);
  });
});

test.describe("IPC Communication", () => {
  test("should get app version", async () => {
    const version = await page.evaluate(async () => {
      return window.electronAPI.app.getVersion();
    });

    expect(version).toMatch(/^\d+\.\d+\.\d+$/);
  });

  test("should access storage", async () => {
    // Set value
    await page.evaluate(async () => {
      await window.electronAPI.storage.set("test-key", { foo: "bar" });
    });

    // Get value
    const value = await page.evaluate(async () => {
      return window.electronAPI.storage.get("test-key");
    });

    expect(value).toEqual({ foo: "bar" });

    // Clean up
    await page.evaluate(async () => {
      await window.electronAPI.storage.delete("test-key");
    });
  });
});

test.describe("File Operations", () => {
  test("should check if file exists", async () => {
    const exists = await page.evaluate(async () => {
      return window.electronAPI.file.exists("package.json");
    });

    expect(exists).toBe(true);
  });
});
```

### Playwright Configuration

```typescript
// playwright.config.ts
import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 30000,
  expect: {
    timeout: 5000,
  },
  fullyParallel: false, // Electron tests should run sequentially
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: "html",
  use: {
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
});
```

---

Version: 1.1.0
Last Updated: 2026-01-10
Changes: Aligned with SKILL.md v1.1.0 updates
