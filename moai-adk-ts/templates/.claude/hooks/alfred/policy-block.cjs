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

// src/claude/hooks/policy-block.ts
var DANGEROUS_COMMANDS = [
  "rm -rf /",
  "rm -rf --no-preserve-root",
  "sudo rm",
  "dd if=/dev/zero",
  ":(){:|:&};:",
  "mkfs."
];
var ALLOWED_PREFIXES = [
  "git ",
  "python",
  "pytest",
  "npm ",
  "node ",
  "go ",
  "cargo ",
  "poetry ",
  "pnpm ",
  "rg ",
  "ls ",
  "cat ",
  "echo ",
  "which ",
  "make ",
  "moai "
];
var PolicyBlock = class {
  name = "policy-block";
  async execute(input) {
    if (input.tool_name !== "Bash") {
      return { success: true };
    }
    const command = this.extractCommand(input.tool_input || {});
    if (!command) {
      return { success: true };
    }
    const commandLower = command.toLowerCase();
    for (const dangerousCommand of DANGEROUS_COMMANDS) {
      if (commandLower.includes(dangerousCommand)) {
        return {
          success: false,
          blocked: true,
          message: `\uC704\uD5D8 \uBA85\uB839\uC774 \uAC10\uC9C0\uB418\uC5C8\uC2B5\uB2C8\uB2E4 (${dangerousCommand}).`,
          exitCode: 2
        };
      }
    }
    if (!this.isAllowedPrefix(command)) {
      console.error(
        "NOTICE: \uB4F1\uB85D\uB418\uC9C0 \uC54A\uC740 \uBA85\uB839\uC785\uB2C8\uB2E4. \uD544\uC694 \uC2DC settings.json \uC758 allow \uBAA9\uB85D\uC744 \uAC31\uC2E0\uD558\uC138\uC694."
      );
    }
    return { success: true };
  }
  /**
   * Extract command from tool input
   */
  extractCommand(toolInput) {
    const raw = toolInput.command || toolInput.cmd;
    if (Array.isArray(raw)) {
      return raw.map(String).join(" ");
    }
    if (typeof raw === "string") {
      return raw.trim();
    }
    return null;
  }
  /**
   * Check if command starts with allowed prefix
   */
  isAllowedPrefix(command) {
    return ALLOWED_PREFIXES.some((prefix) => command.startsWith(prefix));
  }
};
async function main() {
  try {
    const { parseClaudeInput: parseClaudeInput2, outputResult: outputResult2 } = await Promise.resolve().then(() => (init_claude(), claude_exports));
    const input = await parseClaudeInput2();
    const policyBlock = new PolicyBlock();
    const result = await policyBlock.execute(input);
    outputResult2(result);
  } catch (error) {
    console.error(
      `ERROR policy_block: ${error instanceof Error ? error.message : "Unknown error"}`
    );
    process.exit(1);
  }
}
if (__require.main === module) {
  main().catch((error) => {
    console.error(
      `ERROR policy_block: ${error instanceof Error ? error.message : "Unknown error"}`
    );
    process.exit(1);
  });
}

exports.PolicyBlock = PolicyBlock;
exports.main = main;
