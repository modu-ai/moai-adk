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

// src/claude/hooks/session/session-notice.ts
var session_notice_exports = {};
__export(session_notice_exports, {
  SessionNotifier: () => SessionNotifier,
  main: () => main
});
module.exports = __toCommonJS(session_notice_exports);
var import_child_process = require("child_process");
var fs = __toESM(require("fs"));
var path = __toESM(require("path"));
var SessionNotifier = class {
  name = "session-notice";
  projectRoot;
  moaiConfigPath;
  constructor(projectRoot) {
    this.projectRoot = projectRoot || process.cwd();
    this.moaiConfigPath = path.join(this.projectRoot, ".moai", "config.json");
  }
  async execute(input) {
    try {
      if (this.isMoAIProject()) {
        const status = await this.getProjectStatus();
        const output = await this.generateSessionOutput(status);
        return {
          success: true,
          message: output,
          data: status
        };
      } else {
        return {
          success: true,
          message: "\u{1F4A1} Run `/moai:0-project` to initialize MoAI-ADK"
        };
      }
    } catch (error) {
      return { success: true };
    }
  }
  /**
   * Get comprehensive project status
   */
  async getProjectStatus() {
    return {
      projectName: path.basename(this.projectRoot),
      moaiVersion: this.getMoAIVersion(),
      initialized: this.isMoAIProject(),
      constitutionStatus: this.checkConstitutionStatus(),
      pipelineStage: this.getCurrentPipelineStage(),
      specProgress: this.getSpecProgress()
    };
  }
  /**
   * Check if this is a MoAI project
   */
  isMoAIProject() {
    const requiredPaths = [
      path.join(this.projectRoot, ".moai"),
      path.join(this.projectRoot, ".claude", "commands", "moai")
    ];
    return requiredPaths.every((p) => fs.existsSync(p));
  }
  /**
   * Check development guide compliance status
   */
  checkConstitutionStatus() {
    if (!this.isMoAIProject()) {
      return {
        status: "not_initialized",
        violations: []
      };
    }
    const criticalFiles = [
      ".moai/memory/development-guide.md",
      "CLAUDE.md"
    ];
    const violations = [];
    for (const filePath of criticalFiles) {
      if (!fs.existsSync(path.join(this.projectRoot, filePath))) {
        violations.push(`Missing critical file: ${filePath}`);
      }
    }
    return {
      status: violations.length === 0 ? "ok" : "violations_found",
      violations
    };
  }
  /**
   * Get MoAI-ADK version
   */
  getMoAIVersion() {
    try {
      if (fs.existsSync(this.moaiConfigPath)) {
        const configData = fs.readFileSync(this.moaiConfigPath, "utf-8");
        const config = JSON.parse(configData);
        return config.project?.version || "unknown";
      }
    } catch (error) {
    }
    return "unknown";
  }
  /**
   * Get current pipeline stage
   */
  getCurrentPipelineStage() {
    try {
      if (fs.existsSync(this.moaiConfigPath)) {
        const configData = fs.readFileSync(this.moaiConfigPath, "utf-8");
        const config = JSON.parse(configData);
        return config.pipeline?.current_stage || "unknown";
      }
    } catch (error) {
    }
    const specsDir = path.join(this.projectRoot, ".moai", "specs");
    if (fs.existsSync(specsDir)) {
      const hasSpecs = fs.readdirSync(specsDir).some((dir) => fs.existsSync(path.join(specsDir, dir, "spec.md")));
      if (hasSpecs) {
        return "implementation";
      }
    }
    if (this.isMoAIProject()) {
      return "specification";
    }
    return "initialization";
  }
  /**
   * Get SPEC progress information
   */
  getSpecProgress() {
    const specsDir = path.join(this.projectRoot, ".moai", "specs");
    if (!fs.existsSync(specsDir)) {
      return { total: 0, completed: 0 };
    }
    try {
      const specDirs = fs.readdirSync(specsDir).filter((name) => fs.statSync(path.join(specsDir, name)).isDirectory()).filter((name) => name.startsWith("SPEC-"));
      const totalSpecs = specDirs.length;
      let completed = 0;
      for (const specDir of specDirs) {
        const specPath = path.join(specsDir, specDir, "spec.md");
        const planPath = path.join(specsDir, specDir, "plan.md");
        if (fs.existsSync(specPath) && fs.existsSync(planPath)) {
          completed++;
        }
      }
      return { total: totalSpecs, completed };
    } catch (error) {
      return { total: 0, completed: 0 };
    }
  }
  /**
   * Get Git information
   */
  async getGitInfo() {
    const defaultInfo = {
      branch: "unknown",
      commit: "unknown",
      message: "No commit message",
      changesCount: 0
    };
    try {
      const [branch, commit, message, changesCount] = await Promise.all([
        this.runGitCommand(["rev-parse", "--abbrev-ref", "HEAD"]),
        this.runGitCommand(["rev-parse", "HEAD"]),
        this.runGitCommand(["log", "-1", "--pretty=%s"]),
        this.getGitChangesCount()
      ]);
      return {
        branch: branch || defaultInfo.branch,
        commit: commit || defaultInfo.commit,
        message: message || defaultInfo.message,
        changesCount
      };
    } catch (error) {
      return defaultInfo;
    }
  }
  /**
   * Get count of Git changes
   */
  async getGitChangesCount() {
    try {
      const output = await this.runGitCommand(["status", "--porcelain"]);
      if (output) {
        const lines = output.trim().split("\n").filter((line) => line.trim().length > 0);
        return lines.length;
      }
      return 0;
    } catch (error) {
      return 0;
    }
  }
  /**
   * Run a Git command and return output
   */
  async runGitCommand(args) {
    return new Promise((resolve) => {
      const proc = (0, import_child_process.spawn)("git", args, {
        cwd: this.projectRoot,
        stdio: "pipe"
      });
      let stdout = "";
      proc.stdout?.on("data", (data) => {
        stdout += data.toString();
      });
      const timeout = setTimeout(() => {
        proc.kill();
        resolve(null);
      }, 2e3);
      proc.on("close", (code) => {
        clearTimeout(timeout);
        if (code === 0) {
          resolve(stdout.trim());
        } else {
          resolve(null);
        }
      });
      proc.on("error", () => {
        clearTimeout(timeout);
        resolve(null);
      });
    });
  }
  /**
   * Generate session output message
   */
  async generateSessionOutput(status) {
    const lines = [];
    if (status.constitutionStatus.violations.length > 0) {
      lines.push("\u26A0\uFE0F  Development guide violations detected:");
      for (const violation of status.constitutionStatus.violations) {
        lines.push(`   \u2022 ${violation}`);
      }
      lines.push("");
    }
    lines.push(`\u{1F5FF} MoAI-ADK \uD504\uB85C\uC81D\uD2B8: ${status.projectName}`);
    const gitInfo = await this.getGitInfo();
    const shortCommit = gitInfo.commit.substring(0, 7);
    const shortMessage = gitInfo.message.substring(0, 50);
    const ellipsis = gitInfo.message.length > 50 ? "..." : "";
    lines.push(`\u{1F33F} \uD604\uC7AC \uBE0C\uB79C\uCE58: ${gitInfo.branch} (${shortCommit} ${shortMessage}${ellipsis})`);
    if (gitInfo.changesCount > 0) {
      lines.push(`\u{1F4DD} \uBCC0\uACBD\uC0AC\uD56D: ${gitInfo.changesCount}\uAC1C \uD30C\uC77C`);
    }
    const remaining = status.specProgress.total - status.specProgress.completed;
    lines.push(`\u{1F4DD} SPEC \uC9C4\uD589\uB960: ${status.specProgress.completed}/${status.specProgress.total} (\uBBF8\uC644\uB8CC ${remaining}\uAC1C)`);
    lines.push("\u2705 \uD1B5\uD569 \uCCB4\uD06C\uD3EC\uC778\uD2B8 \uC2DC\uC2A4\uD15C \uC0AC\uC6A9 \uAC00\uB2A5");
    return lines.join("\n");
  }
};
async function main() {
  try {
    const notifier = new SessionNotifier();
    const result = await notifier.execute({});
    if (result.message) {
      console.log(result.message);
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
  SessionNotifier,
  main
});
