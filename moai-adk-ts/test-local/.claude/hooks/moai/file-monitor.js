"use strict";
var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(
  // If the importer is in node compatibility mode or this is not an ESM
  // file that has been converted to a CommonJS file using a Babel-
  // compatible transform (i.e. "__esModule" has not been set), then set
  // "default" to the CommonJS "module.exports" for node compatibility.
  isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target,
  mod
));
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/claude/hooks/workflow/file-monitor.ts
var file_monitor_exports = {};
__export(file_monitor_exports, {
  FileMonitor: () => FileMonitor,
  main: () => main
});
module.exports = __toCommonJS(file_monitor_exports);
var fs = __toESM(require("fs"));
var path = __toESM(require("path"));
var import_events = require("events");
var FileMonitor = class extends import_events.EventEmitter {
  name = "file-monitor";
  projectRoot;
  isRunning = false;
  changedFiles = /* @__PURE__ */ new Set();
  lastCheckpointTime = 0;
  checkpointInterval = 3e5;
  // 5 minutes in milliseconds
  watcher;
  // Essential file patterns to watch
  watchPatterns = /* @__PURE__ */ new Set([".py", ".js", ".ts", ".md", ".json", ".yml", ".yaml"]);
  // Directories to ignore
  ignorePatterns = /* @__PURE__ */ new Set([".git", "__pycache__", "node_modules", ".pytest_cache", "dist", "build"]);
  constructor(projectRoot) {
    super();
    this.projectRoot = projectRoot || process.cwd();
  }
  async execute(input) {
    try {
      if (this.isMoAIProject()) {
        if (this.watchFiles()) {
          return {
            success: true,
            message: "\u{1F4C1} File monitoring started"
          };
        } else {
          return {
            success: true,
            message: "\u26A0\uFE0F  Could not start file monitoring"
          };
        }
      }
      return { success: true };
    } catch (error) {
      return { success: true };
    }
  }
  /**
   * Start file watching
   */
  watchFiles() {
    try {
      if (this.isRunning) {
        return true;
      }
      this.watcher = fs.watch(
        this.projectRoot,
        { recursive: true },
        (eventType, filename) => {
          if (filename) {
            const fullPath = path.join(this.projectRoot, filename);
            this.onFileChanged(fullPath);
          }
        }
      );
      this.isRunning = true;
      return true;
    } catch (error) {
      return false;
    }
  }
  /**
   * Stop file watching
   */
  stopWatching() {
    if (this.watcher && this.isRunning) {
      this.watcher.close();
      this.isRunning = false;
    }
  }
  /**
   * Handle file change event
   */
  onFileChanged(filePath) {
    if (!this.shouldMonitorFile(filePath)) {
      return;
    }
    this.changedFiles.add(filePath);
    const event = {
      path: filePath,
      type: "modified",
      // Simplified for now
      timestamp: /* @__PURE__ */ new Date()
    };
    this.emit("fileChanged", event);
    if (this.shouldCreateCheckpoint()) {
      this.createCheckpoint();
    }
  }
  /**
   * Determine if checkpoint should be created
   */
  shouldCreateCheckpoint() {
    const currentTime = Date.now();
    return currentTime - this.lastCheckpointTime > this.checkpointInterval && this.changedFiles.size > 0;
  }
  /**
   * Create checkpoint snapshot
   */
  createCheckpoint() {
    try {
      const currentTime = Date.now();
      this.changedFiles.clear();
      this.lastCheckpointTime = currentTime;
      this.emit("checkpoint", {
        timestamp: new Date(currentTime),
        changedFiles: Array.from(this.changedFiles)
      });
      return true;
    } catch (error) {
      return false;
    }
  }
  /**
   * Check if file should be monitored
   */
  shouldMonitorFile(filePath) {
    const parsedPath = path.parse(filePath);
    const pathParts = filePath.split(path.sep);
    for (const part of pathParts) {
      if (this.ignorePatterns.has(part)) {
        return false;
      }
    }
    return this.watchPatterns.has(parsedPath.ext);
  }
  /**
   * Check if this is a MoAI project
   */
  isMoAIProject() {
    const moaiPath = path.join(this.projectRoot, ".moai");
    return fs.existsSync(moaiPath);
  }
  /**
   * Get list of changed files since last checkpoint
   */
  getChangedFiles() {
    return Array.from(this.changedFiles);
  }
  /**
   * Get monitoring statistics
   */
  getStats() {
    return {
      isRunning: this.isRunning,
      changedFiles: this.changedFiles.size,
      lastCheckpoint: this.lastCheckpointTime > 0 ? new Date(this.lastCheckpointTime) : null
    };
  }
};
async function main() {
  try {
    const projectRoot = process.cwd();
    const monitor = new FileMonitor(projectRoot);
    const result = await monitor.execute({});
    if (result.message) {
      console.log(result.message);
    }
    if (monitor.getStats().isRunning) {
      process.on("SIGINT", () => {
        monitor.stopWatching();
        process.exit(0);
      });
      process.on("SIGTERM", () => {
        monitor.stopWatching();
        process.exit(0);
      });
    }
  } catch (error) {
  }
}
if (require.main === module) {
  main().catch(() => {
  });
}
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  FileMonitor,
  main
});
