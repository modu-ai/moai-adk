'use strict';

var path = require('path');
var child_process = require('child_process');
var fs = require('fs');

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

var path__namespace = /*#__PURE__*/_interopNamespace(path);
var fs__namespace = /*#__PURE__*/_interopNamespace(fs);

// Auto-generated from TypeScript source - DO NOT EDIT DIRECTLY
var __require = /* @__PURE__ */ ((x) => typeof require !== "undefined" ? require : typeof Proxy !== "undefined" ? new Proxy(x, {
  get: (a, b) => (typeof require !== "undefined" ? require : a)[b]
}) : x)(function(x) {
  if (typeof require !== "undefined") return require.apply(this, arguments);
  throw Error('Dynamic require of "' + x + '" is not supported');
});
function isMoAIProject(projectRoot) {
  const moaiDir = path__namespace.join(projectRoot, ".moai");
  const alfredCommands = path__namespace.join(
    projectRoot,
    ".claude",
    "commands",
    "alfred"
  );
  return fs__namespace.existsSync(moaiDir) && fs__namespace.existsSync(alfredCommands);
}
function checkConstitutionStatus(projectRoot) {
  if (!isMoAIProject(projectRoot)) {
    return {
      status: "not_initialized",
      violations: []
    };
  }
  const criticalFiles = [".moai/memory/development-guide.md", "CLAUDE.md"];
  const violations = [];
  for (const filePath of criticalFiles) {
    if (!fs__namespace.existsSync(path__namespace.join(projectRoot, filePath))) {
      violations.push(`Missing critical file: ${filePath}`);
    }
  }
  return {
    status: violations.length === 0 ? "ok" : "violations_found",
    violations
  };
}
function getMoAIVersion(projectRoot) {
  try {
    const moaiConfigPath = path__namespace.join(projectRoot, ".moai", "config.json");
    if (fs__namespace.existsSync(moaiConfigPath)) {
      const configData = fs__namespace.readFileSync(moaiConfigPath, "utf-8");
      const config = JSON.parse(configData);
      if (config.moai?.version && !config.moai.version.includes("{{")) {
        return config.moai.version;
      }
      if (config.project?.version && !config.project.version.includes("{{")) {
        return config.project.version;
      }
    }
    const packageJsonPath = path__namespace.join(
      projectRoot,
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
  } catch (_error) {
  }
  return "unknown";
}
function getCurrentPipelineStage(projectRoot) {
  try {
    const moaiConfigPath = path__namespace.join(projectRoot, ".moai", "config.json");
    if (fs__namespace.existsSync(moaiConfigPath)) {
      const configData = fs__namespace.readFileSync(moaiConfigPath, "utf-8");
      const config = JSON.parse(configData);
      return config.pipeline?.current_stage || "unknown";
    }
  } catch (_error) {
  }
  const specsDir = path__namespace.join(projectRoot, ".moai", "specs");
  if (fs__namespace.existsSync(specsDir)) {
    const hasSpecs = fs__namespace.readdirSync(specsDir).some((dir) => fs__namespace.existsSync(path__namespace.join(specsDir, dir, "spec.md")));
    if (hasSpecs) {
      return "implementation";
    }
  }
  if (isMoAIProject(projectRoot)) {
    return "specification";
  }
  return "initialization";
}
function getSpecProgress(projectRoot) {
  const specsDir = path__namespace.join(projectRoot, ".moai", "specs");
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
  } catch (_error) {
    return { total: 0, completed: 0 };
  }
}
async function runGitCommand(projectRoot, args) {
  return new Promise((resolve) => {
    const proc = child_process.spawn("git", args, {
      cwd: projectRoot,
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
async function getGitChangesCount(projectRoot) {
  try {
    const output = await runGitCommand(projectRoot, ["status", "--porcelain"]);
    if (output) {
      const lines = output.trim().split("\n").filter((line) => line.trim().length > 0);
      return lines.length;
    }
    return 0;
  } catch (_error) {
    return 0;
  }
}
async function getGitInfo(projectRoot) {
  const defaultInfo = {
    branch: "unknown",
    commit: "unknown",
    message: "No commit message",
    changesCount: 0
  };
  try {
    const [branch, commit, message, changesCount] = await Promise.all([
      runGitCommand(projectRoot, ["rev-parse", "--abbrev-ref", "HEAD"]),
      runGitCommand(projectRoot, ["rev-parse", "HEAD"]),
      runGitCommand(projectRoot, ["log", "-1", "--pretty=%s"]),
      getGitChangesCount(projectRoot)
    ]);
    return {
      branch: branch || defaultInfo.branch,
      commit: commit || defaultInfo.commit,
      message: message || defaultInfo.message,
      changesCount
    };
  } catch (_error) {
    return defaultInfo;
  }
}
async function checkLatestVersion(currentVersion) {
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 2e3);
    const response = await fetch("https://registry.npmjs.org/moai-adk/latest", {
      signal: controller.signal,
      headers: {
        Accept: "application/json"
      }
    });
    clearTimeout(timeoutId);
    if (!response.ok) {
      return null;
    }
    const data = await response.json();
    const latest = data.version;
    const hasUpdate = compareVersions(currentVersion, latest) < 0;
    return {
      current: currentVersion,
      latest,
      hasUpdate
    };
  } catch (_error) {
    return null;
  }
}
function compareVersions(v1, v2) {
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

// src/claude/hooks/session-notice/message-builder.ts
function buildViolationMessages(violations) {
  if (violations.length === 0) return [];
  const lines = ["\u26A0\uFE0F  Development guide violations detected:"];
  for (const violation of violations) {
    lines.push(`   \u2022 ${violation}`);
  }
  lines.push("");
  return lines;
}
function buildVersionMessage(status, versionCheck) {
  if (versionCheck?.latest) {
    if (versionCheck.hasUpdate) {
      return `\u{1F4E6} \uBC84\uC804: v${versionCheck.current} \u2192 \u26A1 v${versionCheck.latest} \uC5C5\uB370\uC774\uD2B8 \uAC00\uB2A5`;
    } else {
      return `\u{1F4E6} \uBC84\uC804: v${versionCheck.current} (\uCD5C\uC2E0)`;
    }
  }
  return `\u{1F4E6} \uBC84\uC804: v${status.moaiVersion}`;
}
function buildGitMessage(gitInfo) {
  const shortCommit = gitInfo.commit.substring(0, 7);
  const shortMessage = gitInfo.message.substring(0, 50);
  const ellipsis = gitInfo.message.length > 50 ? "..." : "";
  const lines = [
    `\u{1F33F} \uD604\uC7AC \uBE0C\uB79C\uCE58: ${gitInfo.branch} (${shortCommit} ${shortMessage}${ellipsis})`
  ];
  if (gitInfo.changesCount > 0) {
    lines.push(`\u{1F4DD} \uBCC0\uACBD\uC0AC\uD56D: ${gitInfo.changesCount}\uAC1C \uD30C\uC77C`);
  }
  return lines;
}
function buildSpecProgressMessage(status) {
  const remaining = status.specProgress.total - status.specProgress.completed;
  return `\u{1F4DD} SPEC \uC9C4\uD589\uB960: ${status.specProgress.completed}/${status.specProgress.total} (\uBBF8\uC644\uB8CC ${remaining}\uAC1C)`;
}
async function generateSessionOutput(status, projectRoot) {
  const lines = [];
  const currentVersion = status.moaiVersion;
  const versionCheck = await checkLatestVersion(currentVersion);
  lines.push(...buildViolationMessages(status.constitutionStatus.violations));
  lines.push(`\u{1F5FF} MoAI-ADK \uD504\uB85C\uC81D\uD2B8: ${status.projectName}`);
  lines.push(buildVersionMessage(status, versionCheck));
  const gitInfo = await getGitInfo(projectRoot);
  lines.push(...buildGitMessage(gitInfo));
  lines.push(buildSpecProgressMessage(status));
  lines.push("\u2705 \uD1B5\uD569 \uCCB4\uD06C\uD3EC\uC778\uD2B8 \uC2DC\uC2A4\uD15C \uC0AC\uC6A9 \uAC00\uB2A5");
  return lines.join("\n");
}

// src/claude/hooks/session-notice/index.ts
var SessionNotifier = class {
  name = "session-notice";
  projectRoot;
  constructor(projectRoot) {
    this.projectRoot = projectRoot || process.cwd();
  }
  async execute(_input) {
    try {
      if (isMoAIProject(this.projectRoot)) {
        const status = await this.getProjectStatus();
        const output = await generateSessionOutput(status, this.projectRoot);
        return {
          success: true,
          message: output,
          data: status
        };
      } else {
        return {
          success: true,
          message: "\u{1F4A1} Run `/alfred:8-project` to initialize MoAI-ADK"
        };
      }
    } catch (_error) {
      return { success: true };
    }
  }
  async getProjectStatus() {
    return {
      projectName: path__namespace.basename(this.projectRoot),
      moaiVersion: getMoAIVersion(this.projectRoot),
      initialized: isMoAIProject(this.projectRoot),
      constitutionStatus: checkConstitutionStatus(this.projectRoot),
      pipelineStage: getCurrentPipelineStage(this.projectRoot),
      specProgress: getSpecProgress(this.projectRoot)
    };
  }
  // Backward compatibility for tests
  isMoAIProject = () => isMoAIProject(this.projectRoot);
  checkConstitutionStatus = () => checkConstitutionStatus(this.projectRoot);
  getMoAIVersion = () => getMoAIVersion(this.projectRoot);
  getCurrentPipelineStage = () => getCurrentPipelineStage(this.projectRoot);
  getSpecProgress = () => getSpecProgress(this.projectRoot);
  getGitInfo = () => getGitInfo(this.projectRoot);
  getGitChangesCount = () => getGitChangesCount(this.projectRoot);
  checkLatestVersion = () => checkLatestVersion(this.getMoAIVersion());
};
async function main() {
  try {
    const notifier = new SessionNotifier();
    const result = await notifier.execute({});
    if (result.message) {
      console.log(result.message);
    }
  } catch (_error) {
  }
}
if (__require.main === module) {
  main().catch(() => {
  });
}

exports.SessionNotifier = SessionNotifier;
exports.main = main;
