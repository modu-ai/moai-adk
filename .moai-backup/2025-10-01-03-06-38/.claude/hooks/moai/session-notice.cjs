'use strict';

var child_process = require('child_process');
var fs = require('fs');
var path = require('path');

function _interopNamespace(e) {
  if (e && e.__esModule) return e;
  var n = Object.create(null);
  if (e) {
    Object.keys(e).forEach(function (k) {
      if (k !== 'default') {
        var d = Object.getOwnPropertyDescriptor(e, k);
        Object.defineProperty(n, k, d.get ? d : {
          enumerable: true,
          get: function () { return e[k]; }
        });
      }
    });
  }
  n.default = e;
  return Object.freeze(n);
}

var fs__namespace = /*#__PURE__*/_interopNamespace(fs);
var path__namespace = /*#__PURE__*/_interopNamespace(path);

var __require = /* @__PURE__ */ ((x) => typeof require !== "undefined" ? require : typeof Proxy !== "undefined" ? new Proxy(x, {
  get: (a, b) => (typeof require !== "undefined" ? require : a)[b]
}) : x)(function(x) {
  if (typeof require !== "undefined") return require.apply(this, arguments);
  throw Error('Dynamic require of "' + x + '" is not supported');
});
var SessionNotifier = class {
  name = "session-notice";
  projectRoot;
  moaiConfigPath;
  constructor(projectRoot) {
    this.projectRoot = projectRoot || process.cwd();
    this.moaiConfigPath = path__namespace.join(this.projectRoot, ".moai", "config.json");
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
          message: "\u{1F4A1} Run `/moai:8-project` to initialize MoAI-ADK"
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
      projectName: path__namespace.basename(this.projectRoot),
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
      path__namespace.join(this.projectRoot, ".moai"),
      path__namespace.join(this.projectRoot, ".claude", "commands", "moai")
    ];
    return requiredPaths.every((p) => fs__namespace.existsSync(p));
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
    const criticalFiles = [".moai/memory/development-guide.md", "CLAUDE.md"];
    const violations = [];
    for (const filePath of criticalFiles) {
      if (!fs__namespace.existsSync(path__namespace.join(this.projectRoot, filePath))) {
        violations.push(`Missing critical file: ${filePath}`);
      }
    }
    return {
      status: violations.length === 0 ? "ok" : "violations_found",
      violations
    };
  }
  /**
   * Get MoAI-ADK version from package.json
   * Falls back to config.json if package.json is unavailable
   */
  getMoAIVersion() {
    try {
      const packageJsonPath = path__namespace.join(
        this.projectRoot,
        "node_modules",
        "moai-adk",
        "package.json"
      );
      if (fs__namespace.existsSync(packageJsonPath)) {
        const packageData = fs__namespace.readFileSync(packageJsonPath, "utf-8");
        const packageJson = JSON.parse(packageData);
        if (packageJson.version) {
          return packageJson.version;
        }
      }
      if (fs__namespace.existsSync(this.moaiConfigPath)) {
        const configData = fs__namespace.readFileSync(this.moaiConfigPath, "utf-8");
        const config = JSON.parse(configData);
        const version = config.project?.version;
        if (version && !version.includes("{{") && !version.includes("}}")) {
          return version;
        }
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
      if (fs__namespace.existsSync(this.moaiConfigPath)) {
        const configData = fs__namespace.readFileSync(this.moaiConfigPath, "utf-8");
        const config = JSON.parse(configData);
        return config.pipeline?.current_stage || "unknown";
      }
    } catch (error) {
    }
    const specsDir = path__namespace.join(this.projectRoot, ".moai", "specs");
    if (fs__namespace.existsSync(specsDir)) {
      const hasSpecs = fs__namespace.readdirSync(specsDir).some((dir) => fs__namespace.existsSync(path__namespace.join(specsDir, dir, "spec.md")));
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
    const specsDir = path__namespace.join(this.projectRoot, ".moai", "specs");
    if (!fs__namespace.existsSync(specsDir)) {
      return { total: 0, completed: 0 };
    }
    try {
      const specDirs = fs__namespace.readdirSync(specsDir).filter((name) => fs__namespace.statSync(path__namespace.join(specsDir, name)).isDirectory()).filter((name) => name.startsWith("SPEC-"));
      const totalSpecs = specDirs.length;
      let completed = 0;
      for (const specDir of specDirs) {
        const specPath = path__namespace.join(specsDir, specDir, "spec.md");
        const planPath = path__namespace.join(specsDir, specDir, "plan.md");
        if (fs__namespace.existsSync(specPath) && fs__namespace.existsSync(planPath)) {
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
      const proc = child_process.spawn("git", args, {
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
   * Check for latest version from npm registry
   */
  async checkLatestVersion() {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 2e3);
      const response = await fetch(
        "https://registry.npmjs.org/moai-adk/latest",
        {
          signal: controller.signal,
          headers: {
            Accept: "application/json"
          }
        }
      );
      clearTimeout(timeoutId);
      if (!response.ok) {
        return null;
      }
      const data = await response.json();
      const latest = data.version;
      const current = this.getMoAIVersion();
      const hasUpdate = this.compareVersions(current, latest) < 0;
      return {
        current,
        latest,
        hasUpdate
      };
    } catch (error) {
      return null;
    }
  }
  /**
   * Compare two semantic version strings
   */
  compareVersions(v1, v2) {
    const parts1 = v1.split(".").map(Number);
    const parts2 = v2.split(".").map(Number);
    for (let i = 0; i < Math.max(parts1.length, parts2.length); i++) {
      const num1 = parts1[i] || 0;
      const num2 = parts2[i] || 0;
      if (num1 < num2) return -1;
      if (num1 > num2) return 1;
    }
    return 0;
  }
  /**
   * Generate session output message
   */
  async generateSessionOutput(status) {
    const lines = [];
    const versionCheck = await this.checkLatestVersion();
    if (status.constitutionStatus.violations.length > 0) {
      lines.push("\u26A0\uFE0F  Development guide violations detected:");
      for (const violation of status.constitutionStatus.violations) {
        lines.push(`   \u2022 ${violation}`);
      }
      lines.push("");
    }
    lines.push(`\u{1F5FF} MoAI-ADK \uD504\uB85C\uC81D\uD2B8: ${status.projectName}`);
    if (versionCheck && versionCheck.latest) {
      if (versionCheck.hasUpdate) {
        lines.push(
          `\u{1F4E6} \uBC84\uC804: v${versionCheck.current} \u2192 \u26A1 v${versionCheck.latest} \uC5C5\uB370\uC774\uD2B8 \uAC00\uB2A5`
        );
      } else {
        lines.push(`\u{1F4E6} \uBC84\uC804: v${versionCheck.current} (\uCD5C\uC2E0)`);
      }
    } else {
      lines.push(`\u{1F4E6} \uBC84\uC804: v${status.moaiVersion}`);
    }
    const gitInfo = await this.getGitInfo();
    const shortCommit = gitInfo.commit.substring(0, 7);
    const shortMessage = gitInfo.message.substring(0, 50);
    const ellipsis = gitInfo.message.length > 50 ? "..." : "";
    lines.push(
      `\u{1F33F} \uD604\uC7AC \uBE0C\uB79C\uCE58: ${gitInfo.branch} (${shortCommit} ${shortMessage}${ellipsis})`
    );
    if (gitInfo.changesCount > 0) {
      lines.push(`\u{1F4DD} \uBCC0\uACBD\uC0AC\uD56D: ${gitInfo.changesCount}\uAC1C \uD30C\uC77C`);
    }
    const remaining = status.specProgress.total - status.specProgress.completed;
    lines.push(
      `\u{1F4DD} SPEC \uC9C4\uD589\uB960: ${status.specProgress.completed}/${status.specProgress.total} (\uBBF8\uC644\uB8CC ${remaining}\uAC1C)`
    );
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
if (__require.main === module) {
  main().catch(() => {
  });
}

exports.SessionNotifier = SessionNotifier;
exports.main = main;
