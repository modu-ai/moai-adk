'use strict';

// Auto-generated from TypeScript source - DO NOT EDIT DIRECTLY
var __require = /* @__PURE__ */ ((x) => typeof require !== "undefined" ? require : typeof Proxy !== "undefined" ? new Proxy(x, {
  get: (a, b) => (typeof require !== "undefined" ? require : a)[b]
}) : x)(function(x) {
  if (typeof require !== "undefined") return require.apply(this, arguments);
  throw Error('Dynamic require of "' + x + '" is not supported');
});

// src/hooks/pre-write-guard/index.ts
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
    const { parseClaudeInput, outputResult } = await import('../index');
    const input = await parseClaudeInput();
    const preWriteGuard = new PreWriteGuard();
    const result = await preWriteGuard.execute(input);
    outputResult(result);
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
