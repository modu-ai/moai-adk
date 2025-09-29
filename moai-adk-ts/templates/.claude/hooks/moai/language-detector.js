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

// src/claude/hooks/workflow/language-detector.ts
var language_detector_exports = {};
__export(language_detector_exports, {
  LanguageDetector: () => LanguageDetector,
  main: () => main
});
module.exports = __toCommonJS(language_detector_exports);
var fs = __toESM(require("fs"));
var path = __toESM(require("path"));
var DEFAULT_MAPPINGS = {
  test_runners: {
    python: "pytest",
    javascript: "npm test",
    typescript: "npm test",
    go: "go test ./...",
    rust: "cargo test",
    java: "gradle test | mvn test",
    csharp: "dotnet test",
    cpp: "ctest | make test"
  },
  linters: {
    python: "ruff",
    javascript: "eslint",
    typescript: "eslint",
    go: "golangci-lint",
    rust: "cargo clippy",
    java: "checkstyle",
    csharp: "dotnet format",
    cpp: "clang-tidy"
  },
  formatters: {
    python: "black",
    javascript: "prettier",
    typescript: "prettier",
    go: "gofmt",
    rust: "rustfmt",
    java: "google-java-format",
    csharp: "dotnet format",
    cpp: "clang-format"
  }
};
var LanguageDetector = class {
  name = "language-detector";
  projectRoot;
  constructor(projectRoot) {
    this.projectRoot = projectRoot || process.cwd();
  }
  async execute(input) {
    try {
      const languages = this.detectProjectLanguages();
      if (languages.length === 0) {
        return { success: true };
      }
      const mappings = this.loadMappings();
      const output = this.generateOutput(languages, mappings);
      return {
        success: true,
        message: output,
        data: {
          languages: languages.map((lang) => ({
            language: lang,
            confidence: this.calculateConfidence(lang),
            testRunner: mappings.test_runners[lang] || "-",
            linter: mappings.linters[lang] || "-",
            formatter: mappings.formatters[lang] || "-"
          }))
        }
      };
    } catch (error) {
      return { success: true };
    }
  }
  /**
   * Detect programming languages in the project
   */
  detectProjectLanguages() {
    const rootPath = path.resolve(this.projectRoot);
    const languages = [];
    if (this.hasFile("pyproject.toml") || this.hasFiles("**/*.py")) {
      languages.push("python");
    }
    if (this.hasFile("package.json") || this.hasFiles("**/*.{js,jsx}")) {
      languages.push("javascript");
    }
    if (this.hasFiles("**/*.{ts,tsx}") || this.hasFile("tsconfig.json")) {
      if (!languages.includes("typescript")) {
        languages.push("typescript");
      }
    }
    if (this.hasFile("go.mod") || this.hasFiles("**/*.go")) {
      languages.push("go");
    }
    if (this.hasFile("Cargo.toml") || this.hasFiles("**/*.rs")) {
      languages.push("rust");
    }
    if (this.hasFile("pom.xml") || this.hasFile("build.gradle") || this.hasFile("build.gradle.kts") || this.hasFiles("**/*.java")) {
      languages.push("java");
    }
    if (this.hasFiles("**/*.sln") || this.hasFiles("**/*.csproj") || this.hasFiles("**/*.cs")) {
      languages.push("csharp");
    }
    if (this.hasFiles("**/*.{c,cpp,cxx}") || this.hasFile("CMakeLists.txt")) {
      languages.push("cpp");
    }
    return Array.from(new Set(languages));
  }
  /**
   * Check if a specific file exists
   */
  hasFile(filename) {
    return fs.existsSync(path.join(this.projectRoot, filename));
  }
  /**
   * Check if files matching pattern exist (simplified implementation)
   */
  hasFiles(pattern) {
    try {
      const extensionMatch = pattern.match(/\*\*\/?\*\.(\w+)/);
      if (!extensionMatch) {
        return false;
      }
      const extension = extensionMatch[1];
      return this.findFilesWithExtension(this.projectRoot, extension);
    } catch {
      return false;
    }
  }
  /**
   * Recursively find files with specific extension
   */
  findFilesWithExtension(dir, extension) {
    try {
      const entries = fs.readdirSync(dir, { withFileTypes: true });
      for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);
        if (entry.isDirectory()) {
          if (["node_modules", ".git", "__pycache__", ".pytest_cache", "dist", "build"].includes(entry.name)) {
            continue;
          }
          if (this.findFilesWithExtension(fullPath, extension)) {
            return true;
          }
        } else if (entry.isFile()) {
          if (entry.name.endsWith(`.${extension}`)) {
            return true;
          }
        }
      }
      return false;
    } catch {
      return false;
    }
  }
  /**
   * Load language mappings from configuration
   */
  loadMappings() {
    const mappingPath = path.join(this.projectRoot, ".moai", "config", "language_mappings.json");
    try {
      if (fs.existsSync(mappingPath)) {
        const data = fs.readFileSync(mappingPath, "utf-8");
        return JSON.parse(data);
      }
    } catch (error) {
    }
    return DEFAULT_MAPPINGS;
  }
  /**
   * Calculate confidence score for language detection
   */
  calculateConfidence(language) {
    return 0.85;
  }
  /**
   * Generate human-readable output
   */
  generateOutput(languages, mappings) {
    const lines = [];
    lines.push(`\u{1F310} \uAC10\uC9C0\uB41C \uC5B8\uC5B4: ${languages.join(", ")}`);
    if (languages.length > 0) {
      lines.push("\u{1F527} \uAD8C\uC7A5 \uB3C4\uAD6C:");
      for (const lang of languages) {
        const testRunner = mappings.test_runners[lang] || "-";
        const linter = mappings.linters[lang] || "-";
        const formatter = mappings.formatters[lang] || "-";
        lines.push(`- ${lang}: test=${testRunner}, lint=${linter}, format=${formatter}`);
      }
      lines.push("\u{1F4A1} \uD544\uC694 \uC2DC /moai:2-build \uB2E8\uACC4\uC5D0\uC11C \uD574\uB2F9 \uB3C4\uAD6C\uB97C \uC0AC\uC6A9\uD574 TDD\uB97C \uC2E4\uD589\uD558\uC138\uC694.");
    }
    return lines.join("\n");
  }
  /**
   * Get detected languages as structured data
   */
  getLanguages() {
    const languages = this.detectProjectLanguages();
    const mappings = this.loadMappings();
    return languages.map((lang) => ({
      language: lang,
      confidence: this.calculateConfidence(lang),
      testRunner: mappings.test_runners[lang],
      linter: mappings.linters[lang],
      formatter: mappings.formatters[lang]
    }));
  }
};
async function main() {
  try {
    const detector = new LanguageDetector();
    const result = await detector.execute({});
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
  LanguageDetector,
  main
});
