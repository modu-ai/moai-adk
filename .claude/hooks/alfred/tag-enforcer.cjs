'use strict';

var fs = require('fs/promises');
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

// src/claude/hooks/tag-enforcer/tag-patterns.ts
var CODE_FIRST_PATTERNS = {
  // 핵심 TAG 라인들
  MAIN_TAG: /^\s*\*\s*@DOC:([A-Z]+):([A-Z0-9_-]+)\s*$/m,
  CHAIN_LINE: /^\s*\*\s*CHAIN:\s*(.+)\s*$/m,
  DEPENDS_LINE: /^\s*\*\s*DEPENDS:\s*(.+)\s*$/m,
  STATUS_LINE: /^\s*\*\s*STATUS:\s*(\w+)\s*$/m,
  CREATED_LINE: /^\s*\*\s*CREATED:\s*(\d{4}-\d{2}-\d{2})\s*$/m,
  IMMUTABLE_MARKER: /^\s*\*\s*@IMMUTABLE\s*$/m,
  // TAG 참조
  TAG_REFERENCE: /@([A-Z]+):([A-Z0-9-]+)/g
};
var VALID_CATEGORIES = {
  // Lifecycle (필수 체인)
  lifecycle: ["SPEC", "REQ", "DESIGN", "TASK", "TEST"],
  // Implementation (선택적)
  implementation: ["FEATURE", "API", "FIX"]
};

// src/claude/hooks/tag-enforcer/tag-validator.ts
var TagValidator = class {
  /**
   * TAG 블록 추출 (파일 최상단에서만)
   */
  extractTagBlock(content) {
    const lines = content.split("\n");
    let inBlock = false;
    let blockLines = [];
    let startLineNumber = 0;
    for (let i = 0; i < Math.min(lines.length, 30); i++) {
      const line = lines[i]?.trim();
      if (!line || line.startsWith("#!")) {
        continue;
      }
      if (line.startsWith("/**") && !inBlock) {
        inBlock = true;
        blockLines = [line];
        startLineNumber = i + 1;
        continue;
      }
      if (inBlock) {
        blockLines.push(line);
        if (line.endsWith("*/")) {
          const blockContent = blockLines.join("\n");
          if (CODE_FIRST_PATTERNS.MAIN_TAG.test(blockContent)) {
            return {
              content: blockContent,
              lineNumber: startLineNumber
            };
          }
          inBlock = false;
          blockLines = [];
          continue;
        }
      }
      if (!inBlock && line && !line.startsWith("//") && !line.startsWith("/*")) {
        break;
      }
    }
    return null;
  }
  /**
   * TAG 블록에서 메인 TAG 추출
   */
  extractMainTag(blockContent) {
    const match = CODE_FIRST_PATTERNS.MAIN_TAG.exec(blockContent);
    return match ? `@${match[1]}:${match[2]}` : "UNKNOWN";
  }
  /**
   * TAG 블록 정규화 (비교용)
   */
  normalizeTagBlock(blockContent) {
    return blockContent.split("\n").map((line) => line.trim()).filter((line) => line.length > 0).join("\n");
  }
  /**
   * Code-First TAG 유효성 검증
   */
  validateCodeFirstTag(content) {
    const violations = [];
    const warnings = [];
    let hasTag = false;
    const tagBlock = this.extractTagBlock(content);
    if (!tagBlock) {
      return {
        isValid: true,
        // TAG 블록이 없어도 차단하지 않음 (권장사항)
        violations: [],
        warnings: ["\uD30C\uC77C \uCD5C\uC0C1\uB2E8\uC5D0 TAG \uBE14\uB85D\uC774 \uC5C6\uC2B5\uB2C8\uB2E4 (\uAD8C\uC7A5\uC0AC\uD56D)"],
        hasTag: false
      };
    }
    hasTag = true;
    const blockContent = tagBlock.content;
    const tagMatch = CODE_FIRST_PATTERNS.MAIN_TAG.exec(blockContent);
    if (!tagMatch) {
      violations.push("@TAG \uB77C\uC778\uC774 \uBC1C\uACAC\uB418\uC9C0 \uC54A\uC558\uC2B5\uB2C8\uB2E4");
    } else {
      const [, category, domainId] = tagMatch;
      const allValidCategories = [
        ...VALID_CATEGORIES.lifecycle,
        ...VALID_CATEGORIES.implementation
      ];
      const validCategorySet = /* @__PURE__ */ new Set([...allValidCategories]);
      if (category && !validCategorySet.has(category)) {
        violations.push(`\uC720\uD6A8\uD558\uC9C0 \uC54A\uC740 TAG \uCE74\uD14C\uACE0\uB9AC: ${category}`);
      }
      if (domainId && !/^[A-Z0-9]+-\d{3,}$/.test(domainId)) {
        warnings.push(`\uB3C4\uBA54\uC778 ID \uD615\uC2DD \uAD8C\uC7A5: ${domainId} -> DOMAIN-001`);
      }
    }
    const chainMatch = CODE_FIRST_PATTERNS.CHAIN_LINE.exec(blockContent);
    if (chainMatch) {
      const chainStr = chainMatch[1];
      if (chainStr) {
        const chainTags = chainStr.split(/\s*->\s*/);
        for (const chainTag of chainTags) {
          if (!CODE_FIRST_PATTERNS.TAG_REFERENCE.test(chainTag.trim())) {
            warnings.push(`\uCCB4\uC778\uC758 TAG \uD615\uC2DD\uC744 \uD655\uC778\uD558\uC138\uC694: ${chainTag.trim()}`);
          }
        }
      }
    }
    const dependsMatch = CODE_FIRST_PATTERNS.DEPENDS_LINE.exec(blockContent);
    if (dependsMatch) {
      const dependsStr = dependsMatch[1];
      if (dependsStr && dependsStr.trim().toLowerCase() !== "none") {
        const dependsTags = dependsStr.split(/,\s*/);
        for (const dependTag of dependsTags) {
          if (!CODE_FIRST_PATTERNS.TAG_REFERENCE.test(dependTag.trim())) {
            warnings.push(`\uC758\uC874\uC131 TAG \uD615\uC2DD\uC744 \uD655\uC778\uD558\uC138\uC694: ${dependTag.trim()}`);
          }
        }
      }
    }
    const statusMatch = CODE_FIRST_PATTERNS.STATUS_LINE.exec(blockContent);
    if (statusMatch) {
      const status = statusMatch[1]?.toLowerCase();
      if (status && !["active", "deprecated", "completed"].includes(status)) {
        warnings.push(`\uC54C \uC218 \uC5C6\uB294 STATUS: ${status}`);
      }
    }
    const createdMatch = CODE_FIRST_PATTERNS.CREATED_LINE.exec(blockContent);
    if (createdMatch) {
      const created = createdMatch[1];
      if (created && !/^\d{4}-\d{2}-\d{2}$/.test(created)) {
        warnings.push(`\uC0DD\uC131 \uB0A0\uC9DC \uD615\uC2DD\uC744 \uD655\uC778\uD558\uC138\uC694: ${created} (YYYY-MM-DD)`);
      }
    }
    if (!CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.test(blockContent)) {
      warnings.push(
        "@IMMUTABLE \uB9C8\uCEE4\uB97C \uCD94\uAC00\uD558\uC5EC TAG \uBD88\uBCC0\uC131\uC744 \uBCF4\uC7A5\uD558\uB294 \uAC83\uC744 \uAD8C\uC7A5\uD569\uB2C8\uB2E4"
      );
    }
    return {
      isValid: violations.length === 0,
      violations,
      warnings,
      hasTag
    };
  }
};

// src/claude/hooks/tag-enforcer.ts
var CodeFirstTAGEnforcer = class {
  name = "tag-enforcer";
  validator;
  constructor() {
    this.validator = new TagValidator();
  }
  /**
   * 새로운 Code-First TAG 불변성 검사 실행
   */
  async execute(input) {
    try {
      if (!this.isWriteOperation(input.tool_name)) {
        return { success: true };
      }
      const filePath = this.extractFilePath(input.tool_input || {});
      if (!filePath || !this.shouldEnforceTags(filePath)) {
        return { success: true };
      }
      const oldContent = await this.getOriginalFileContent(filePath);
      const newContent = this.extractFileContent(input.tool_input || {});
      const immutabilityCheck = this.checkImmutability(
        oldContent,
        newContent,
        filePath
      );
      if (immutabilityCheck.violated) {
        return {
          success: false,
          blocked: true,
          message: `\u{1F6AB} @IMMUTABLE TAG \uC218\uC815 \uAE08\uC9C0: ${immutabilityCheck.violationDetails}`,
          data: {
            suggestions: this.generateImmutabilityHelp(immutabilityCheck)
          },
          exitCode: 2
        };
      }
      const validation = this.validator.validateCodeFirstTag(newContent);
      if (!validation.isValid) {
        return {
          success: false,
          blocked: true,
          message: `\u{1F3F7}\uFE0F Code-First TAG \uAC80\uC99D \uC2E4\uD328: ${validation.violations.join(", ")}`,
          data: {
            suggestions: this.generateTagSuggestions(filePath, newContent)
          },
          exitCode: 2
        };
      }
      if (validation.warnings.length > 0) {
        console.error(`\u26A0\uFE0F TAG \uAC1C\uC120 \uAD8C\uC7A5: ${validation.warnings.join(", ")}`);
      }
      return {
        success: true,
        message: validation.hasTag ? `\u2705 Code-First TAG \uAC80\uC99D \uC644\uB8CC` : `\u{1F4DD} TAG \uBE14\uB85D\uC774 \uC5C6\uB294 \uD30C\uC77C (\uAD8C\uC7A5\uC0AC\uD56D)`
      };
    } catch (error) {
      console.error(
        `TAG Enforcer \uACBD\uACE0: ${error instanceof Error ? error.message : "Unknown error"}`
      );
      return { success: true };
    }
  }
  /**
   * 파일 쓰기 작업 확인
   */
  isWriteOperation(toolName) {
    return !!toolName && ["Write", "Edit", "MultiEdit", "NotebookEdit"].includes(toolName);
  }
  /**
   * 도구 입력에서 파일 경로 추출
   */
  extractFilePath(toolInput) {
    return toolInput.file_path || toolInput.filePath || toolInput.notebook_path || null;
  }
  /**
   * 도구 입력에서 파일 내용 추출
   */
  extractFileContent(toolInput) {
    if (toolInput.content) return toolInput.content;
    if (toolInput.new_string) return toolInput.new_string;
    if (toolInput.new_source) return toolInput.new_source;
    if (toolInput.edits && Array.isArray(toolInput.edits)) {
      return toolInput.edits.map((edit) => edit.new_string).join("\n");
    }
    return "";
  }
  /**
   * TAG 검증 대상 파일인지 확인
   */
  shouldEnforceTags(filePath) {
    const enforceExtensions = [
      ".ts",
      ".tsx",
      ".js",
      ".jsx",
      ".py",
      ".md",
      ".go",
      ".rs",
      ".java",
      ".cpp",
      ".hpp"
    ];
    const ext = path__namespace.extname(filePath);
    if (filePath.includes("test") || filePath.includes("spec") || filePath.includes("__test__")) {
      return false;
    }
    if (filePath.includes("node_modules") || filePath.includes(".git") || filePath.includes("dist") || filePath.includes("build")) {
      return false;
    }
    return enforceExtensions.includes(ext);
  }
  /**
   * 기존 파일 내용 읽기
   */
  async getOriginalFileContent(filePath) {
    try {
      return await fs__namespace.readFile(filePath, "utf-8");
    } catch (_error) {
      return "";
    }
  }
  /**
   * @IMMUTABLE TAG 블록 수정 검사 (핵심 불변성 보장)
   */
  checkImmutability(oldContent, newContent, _filePath) {
    if (!oldContent) {
      return { violated: false };
    }
    const oldTagBlock = this.validator.extractTagBlock(oldContent);
    const newTagBlock = this.validator.extractTagBlock(newContent);
    if (!oldTagBlock) {
      return { violated: false };
    }
    const wasImmutable = CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.test(
      oldTagBlock.content
    );
    if (!wasImmutable) {
      return { violated: false };
    }
    if (!newTagBlock) {
      return {
        violated: true,
        modifiedTag: this.validator.extractMainTag(oldTagBlock.content),
        violationDetails: "@IMMUTABLE TAG \uBE14\uB85D\uC774 \uC0AD\uC81C\uB418\uC5C8\uC2B5\uB2C8\uB2E4"
      };
    }
    const oldNormalized = this.validator.normalizeTagBlock(oldTagBlock.content);
    const newNormalized = this.validator.normalizeTagBlock(newTagBlock.content);
    if (oldNormalized !== newNormalized) {
      return {
        violated: true,
        modifiedTag: this.validator.extractMainTag(oldTagBlock.content),
        violationDetails: "@IMMUTABLE TAG \uBE14\uB85D\uC758 \uB0B4\uC6A9\uC774 \uBCC0\uACBD\uB418\uC5C8\uC2B5\uB2C8\uB2E4"
      };
    }
    return { violated: false };
  }
  /**
   * @IMMUTABLE 위반 시 도움말 생성
   */
  generateImmutabilityHelp(immutabilityCheck) {
    const help = [
      "\u{1F6AB} @IMMUTABLE TAG \uC218\uC815\uC774 \uAC10\uC9C0\uB418\uC5C8\uC2B5\uB2C8\uB2E4.",
      "",
      "\u{1F4CB} Code-First TAG \uADDC\uCE59:",
      "\u2022 @IMMUTABLE \uB9C8\uCEE4\uAC00 \uC788\uB294 TAG \uBE14\uB85D\uC740 \uC218\uC815\uD560 \uC218 \uC5C6\uC2B5\uB2C8\uB2E4",
      "\u2022 TAG\uB294 \uD55C\uBC88 \uC791\uC131\uB418\uBA74 \uBD88\uBCC0(immutable)\uC785\uB2C8\uB2E4",
      "\u2022 \uAE30\uB2A5 \uBCC0\uACBD \uC2DC\uC5D0\uB294 \uC0C8\uB85C\uC6B4 TAG\uB97C \uC0DD\uC131\uD558\uC138\uC694",
      "",
      "\u2705 \uAD8C\uC7A5 \uD574\uACB0 \uBC29\uBC95:",
      "1. \uC0C8\uB85C\uC6B4 TAG ID\uB85C \uC0C8 \uAE30\uB2A5\uC744 \uAD6C\uD604\uD558\uC138\uC694",
      "   \uC608: @DOC:FEATURE:AUTH-002",
      "2. \uAE30\uC874 TAG\uC5D0 @DOC \uB9C8\uCEE4\uB97C \uCD94\uAC00\uD558\uC138\uC694",
      "3. \uC0C8 TAG\uC5D0\uC11C \uC774\uC804 TAG\uB97C \uCC38\uC870\uD558\uC138\uC694",
      "   \uC608: REPLACES: FEATURE:AUTH-001",
      "",
      `\u{1F50D} \uC218\uC815 \uC2DC\uB3C4\uB41C TAG: ${immutabilityCheck.modifiedTag || "UNKNOWN"}`
    ];
    return help.join("\n");
  }
  /**
   * TAG 제안 생성
   */
  generateTagSuggestions(filePath, _content) {
    const fileName = path__namespace.basename(filePath, path__namespace.extname(filePath));
    const suggestions = [
      "\u{1F4DD} Code-First TAG \uBE14\uB85D \uC608\uC2DC:",
      "",
      "```",
      "/**",
      ` * @DOC:FEATURE:${fileName.toUpperCase()}-001`,
      ` * CHAIN: REQ:${fileName.toUpperCase()}-001 -> DESIGN:${fileName.toUpperCase()}-001 -> TASK:${fileName.toUpperCase()}-001 -> TEST:${fileName.toUpperCase()}-001`,
      " * DEPENDS: NONE",
      " * STATUS: active",
      ` * CREATED: ${(/* @__PURE__ */ new Date()).toISOString().split("T")[0]}`,
      " * @IMMUTABLE",
      " */",
      "```",
      "",
      "\u{1F3AF} TAG \uCE74\uD14C\uACE0\uB9AC \uAC00\uC774\uB4DC:",
      "\u2022 SPEC, REQ, DESIGN, TASK, TEST: \uD544\uC218 \uC0DD\uBA85\uC8FC\uAE30",
      "\u2022 FEATURE, API, FIX: \uAD6C\uD604 \uCE74\uD14C\uACE0\uB9AC",
      "",
      "\u{1F4A1} \uCD94\uAC00 \uD301:",
      "\u2022 TAG \uBE14\uB85D\uC740 \uD30C\uC77C \uCD5C\uC0C1\uB2E8\uC5D0 \uC704\uCE58",
      "\u2022 @IMMUTABLE \uB9C8\uCEE4\uB85C \uBD88\uBCC0\uC131 \uBCF4\uC7A5",
      "\u2022 \uCCB4\uC778\uC73C\uB85C \uAD00\uB828 TAG\uB4E4 \uC5F0\uACB0"
    ];
    return suggestions.join("\n");
  }
};
async function main() {
  try {
    const { parseClaudeInput: parseClaudeInput2 } = await Promise.resolve().then(() => (init_claude(), claude_exports));
    const input = await parseClaudeInput2();
    const enforcer = new CodeFirstTAGEnforcer();
    const result = await enforcer.execute(input);
    if (result.blocked) {
      console.error(`BLOCKED: ${result.message}`);
      if (result.data?.suggestions) {
        console.error(
          `
\u{1F4DD} Code-First TAG \uAC00\uC774\uB4DC:
${result.data.suggestions}`
        );
      }
      process.exit(2);
    } else if (!result.success) {
      console.error(`ERROR: ${result.message}`);
      process.exit(result.exitCode || 1);
    } else if (result.message) {
      console.log(result.message);
    }
    process.exit(0);
  } catch (error) {
    console.error(
      `Code-First TAG Enforcer \uC624\uB958: ${error instanceof Error ? error.message : "Unknown error"}`
    );
    process.exit(0);
  }
}
if (__require.main === module) {
  main().catch((error) => {
    console.error(
      `Code-First TAG Enforcer \uCE58\uBA85\uC801 \uC624\uB958: ${error instanceof Error ? error.message : "Unknown error"}`
    );
    process.exit(0);
  });
}

exports.CodeFirstTAGEnforcer = CodeFirstTAGEnforcer;
exports.main = main;
