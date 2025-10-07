'use strict';

// Auto-generated from TypeScript source - DO NOT EDIT DIRECTLY
var __defProp = Object.defineProperty;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __require = /* @__PURE__ */ ((x) => typeof require !== "undefined" ? require : typeof Proxy !== "undefined" ? new Proxy(x, {
  get: (a, b) => (typeof require !== "undefined" ? require : a)[b]
}) : x)(function(x) {
  if (typeof require !== "undefined") return require.apply(this, arguments);
  throw Error('Dynamic require of "' + x + '" is not supported');
});
var __esm = (fn, res) => function __init() {
  return fn && (res = (0, fn[__getOwnPropNames(fn)[0]])(fn = 0)), res;
};
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};

// src/claude/index.ts
var claude_exports = {};
__export(claude_exports, {
  outputResult: () => outputResult,
  parseClaudeInput: () => parseClaudeInput
});
async function parseClaudeInput() {
  return new Promise((resolve, reject) => {
    let data = "";
    process.stdin.setEncoding("utf8");
    process.stdin.on("data", (chunk) => {
      data += chunk;
    });
    process.stdin.on("end", () => {
      try {
        if (!data.trim()) {
          resolve({
            tool_name: "Unknown",
            tool_input: {},
            context: {}
          });
          return;
        }
        const parsed = JSON.parse(data);
        resolve(parsed);
      } catch (error) {
        reject(
          new Error(
            `Failed to parse input: ${error instanceof Error ? error.message : "Unknown error"}`
          )
        );
      }
    });
    process.stdin.on("error", (error) => {
      reject(new Error(`Failed to read stdin: ${error.message}`));
    });
  });
}
function outputResult(result) {
  if (result.blocked) {
    console.error(`BLOCKED: ${result.message || "Operation blocked"}`);
    if (result.data?.suggestions) {
      console.error(`
${result.data.suggestions}`);
    }
    process.exit(result.exitCode || 2);
  } else if (!result.success) {
    console.error(`ERROR: ${result.message || "Operation failed"}`);
    if (result.warnings && result.warnings.length > 0) {
      console.error(`Warnings:
${result.warnings.join("\n")}`);
    }
    process.exit(result.exitCode || 1);
  } else {
    if (result.message) {
      console.log(result.message);
    }
    if (result.warnings && result.warnings.length > 0) {
      console.warn(`Warnings:
${result.warnings.join("\n")}`);
    }
    process.exit(0);
  }
}
var init_claude = __esm({
  "src/claude/index.ts"() {
  }
});

// src/claude/hooks/pre-write-guard.ts
var SENSITIVE_KEYWORDS = [".env", "/secrets", "/.git/", "/.ssh"];
var PROTECTED_PATHS = [".moai/memory/"];
var PreWriteGuard = class {
  name = "pre-write-guard";
  async execute(input) {
    const toolName = input.tool_name;
    if (!toolName || !["Write", "Edit", "MultiEdit"].includes(toolName)) {
      return { success: true };
    }
    const toolInput = input.tool_input || {};
    const filePath = this.extractFilePath(toolInput);
    if (!this.checkFileSafety(filePath || "")) {
      return {
        success: false,
        blocked: true,
        message: "\uBBFC\uAC10\uD55C \uD30C\uC77C\uC740 \uD3B8\uC9D1\uD560 \uC218 \uC5C6\uC2B5\uB2C8\uB2E4.",
        exitCode: 2
      };
    }
    return { success: true };
  }
  /**
   * Extract file path from tool input
   */
  extractFilePath(toolInput) {
    return toolInput.file_path || toolInput.filePath || toolInput.path || null;
  }
  /**
   * Check if file is safe to edit
   */
  checkFileSafety(filePath) {
    if (!filePath) {
      return true;
    }
    const pathLower = filePath.toLowerCase();
    for (const keyword of SENSITIVE_KEYWORDS) {
      if (pathLower.includes(keyword)) {
        return false;
      }
    }
    const isTemplate = filePath.includes("/templates/.moai/");
    if (!isTemplate) {
      for (const protectedPath of PROTECTED_PATHS) {
        if (filePath.includes(protectedPath)) {
          return false;
        }
      }
    }
    return true;
  }
};
async function main() {
  try {
    const { parseClaudeInput: parseClaudeInput2, outputResult: outputResult2 } = await Promise.resolve().then(() => (init_claude(), claude_exports));
    const input = await parseClaudeInput2();
    const preWriteGuard = new PreWriteGuard();
    const result = await preWriteGuard.execute(input);
    outputResult2(result);
  } catch (_error) {
    process.exit(0);
  }
}
if (__require.main === module) {
  main().catch(() => {
    process.exit(0);
  });
}

exports.PreWriteGuard = PreWriteGuard;
exports.main = main;
