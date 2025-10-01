"use strict";
var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __esm = (fn, res) => function __init() {
  return fn && (res = (0, fn[__getOwnPropNames(fn)[0]])(fn = 0)), res;
};
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

// node_modules/tsup/assets/cjs_shims.js
var init_cjs_shims = __esm({
  "node_modules/tsup/assets/cjs_shims.js"() {
    "use strict";
  }
});

// src/claude/hooks/types.ts
var init_types = __esm({
  "src/claude/hooks/types.ts"() {
    "use strict";
    init_cjs_shims();
  }
});

// src/claude/hooks/security/steering-guard.ts
async function main() {
  try {
    const { parseClaudeInput: parseClaudeInput2, outputResult: outputResult2 } = await Promise.resolve().then(() => (init_hooks(), hooks_exports));
    const input = await parseClaudeInput2();
    const guard = new SteeringGuard();
    const result = await guard.execute(input);
    outputResult2(result);
  } catch (error) {
    console.error(`ERROR steering_guard: ${error instanceof Error ? error.message : "Unknown error"}`);
    process.exit(1);
  }
}
var fs, path, os, BANNED_PATTERNS, SESSION_NOTIFIED_FILE, SteeringGuard;
var init_steering_guard = __esm({
  "src/claude/hooks/security/steering-guard.ts"() {
    "use strict";
    init_cjs_shims();
    fs = __toESM(require("fs"));
    path = __toESM(require("path"));
    os = __toESM(require("os"));
    BANNED_PATTERNS = [
      {
        pattern: /ignore (the )?(claude|constitution|steering|instructions)/i,
        message: "\uD5CC\uBC95/\uC9C0\uCE68 \uBB34\uC2DC\uB294 \uD5C8\uC6A9\uB418\uC9C0 \uC54A\uC2B5\uB2C8\uB2E4.",
        severity: "critical"
      },
      {
        pattern: /disable (all )?(hooks?|guards?|polic(y|ies))/i,
        message: "Hook/Guard \uD574\uC81C \uC694\uCCAD\uC740 \uCC28\uB2E8\uB418\uC5C8\uC2B5\uB2C8\uB2E4.",
        severity: "critical"
      },
      {
        pattern: /rm -rf/i,
        message: "\uC704\uD5D8\uD55C \uC178 \uBA85\uB839\uC744 \uD504\uB86C\uD504\uD2B8\uB85C \uC81C\uCD9C\uD560 \uC218 \uC5C6\uC2B5\uB2C8\uB2E4.",
        severity: "high"
      },
      {
        pattern: /drop (all )?safeguards/i,
        message: "\uC548\uC804\uC7A5\uCE58 \uC81C\uAC70 \uC694\uCCAD\uC740 \uAC70\uBD80\uB429\uB2C8\uB2E4.",
        severity: "critical"
      },
      {
        pattern: /clear (all )?(memory|steering)/i,
        message: "Steering \uBA54\uBAA8\uB9AC\uB97C \uAC15\uC81C \uC0AD\uC81C\uD558\uB294 \uC694\uCCAD\uC740 \uC9C0\uC6D0\uD558\uC9C0 \uC54A\uC2B5\uB2C8\uB2E4.",
        severity: "high"
      }
    ];
    SESSION_NOTIFIED_FILE = path.join(os.tmpdir(), "moai_session_notified");
    SteeringGuard = class {
      name = "steering-guard";
      async execute(input) {
        this.showSessionNotice();
        const prompt = input.prompt;
        if (!prompt || typeof prompt !== "string") {
          return { success: true };
        }
        for (const pattern of BANNED_PATTERNS) {
          if (pattern.pattern.test(prompt)) {
            return {
              success: false,
              blocked: true,
              message: pattern.message,
              exitCode: 2
            };
          }
        }
        return {
          success: true,
          message: "Steering Guard: \uAC1C\uBC1C \uAC00\uC774\uB4DC\uACFC TAG \uADDC\uCE59\uC744 \uC900\uC218\uD558\uBA70 \uC791\uC5C5\uC744 \uC9C4\uD589\uD569\uB2C8\uB2E4."
        };
      }
      /**
       * Check if this is a MoAI project
       */
      checkMoAIProject() {
        const currentDir = process.cwd();
        const moaiPath = path.join(currentDir, ".moai");
        const claudePath = path.join(currentDir, "CLAUDE.md");
        return fs.existsSync(moaiPath) && fs.existsSync(claudePath);
      }
      /**
       * Check hybrid system status
       */
      checkHybridSystemStatus() {
        const currentDir = process.cwd();
        const tsProject = path.join(currentDir, "moai-adk-ts");
        const hasTypeScript = fs.existsSync(tsProject) && fs.existsSync(path.join(tsProject, "package.json"));
        const pythonBridge = path.join(currentDir, "src", "moai_adk", "core", "bridge");
        const hasPythonBridge = fs.existsSync(pythonBridge) && fs.existsSync(path.join(pythonBridge, "typescript_bridge.py"));
        if (hasTypeScript && hasPythonBridge) {
          return {
            status: "full_hybrid",
            description: "Python + TypeScript \uC644\uC804 \uD1B5\uD569 \u{1F517}"
          };
        } else if (hasPythonBridge) {
          return {
            status: "python_only",
            description: "Python \uBE0C\uB9BF\uC9C0 (TypeScript \uC5C6\uC74C) \u{1F40D}"
          };
        } else if (hasTypeScript) {
          return {
            status: "typescript_only",
            description: "TypeScript (\uBE0C\uB9BF\uC9C0 \uC5C6\uC74C) \u26A0\uFE0F"
          };
        } else {
          return {
            status: "legacy",
            description: "\uAE30\uC874 Python \uC2DC\uC2A4\uD15C \u{1F4E6}"
          };
        }
      }
      /**
       * Show session notice (first time only)
       */
      showSessionNotice() {
        if (fs.existsSync(SESSION_NOTIFIED_FILE)) {
          return;
        }
        if (!this.checkMoAIProject()) {
          return;
        }
        const hybridStatus = this.checkHybridSystemStatus();
        console.error("\u{1F680} MoAI-ADK \uD558\uC774\uBE0C\uB9AC\uB4DC \uD504\uB85C\uC81D\uD2B8\uAC00 \uAC10\uC9C0\uB418\uC5C8\uC2B5\uB2C8\uB2E4!");
        console.error("\u{1F4D6} \uAC1C\uBC1C \uAC00\uC774\uB4DC: CLAUDE.md | TRUST \uC6D0\uCE59: .moai/memory/development-guide.md");
        console.error("\u26A1 \uD558\uC774\uBE0C\uB9AC\uB4DC \uC6CC\uD06C\uD50C\uB85C\uC6B0: /moai:1-spec \u2192 /moai:2-build \u2192 /moai:3-sync");
        console.error(`\u{1F517} \uC2DC\uC2A4\uD15C \uC0C1\uD0DC: ${hybridStatus.description}`);
        console.error("\u{1F527} \uB514\uBC84\uAE45: /moai:4-debug | \uC124\uC815 \uAD00\uB9AC: @agent-cc-manager");
        console.error("");
        try {
          fs.writeFileSync(SESSION_NOTIFIED_FILE, "notified");
        } catch {
        }
      }
    };
    if (require.main === module) {
      main().catch((error) => {
        console.error(`ERROR steering_guard: ${error instanceof Error ? error.message : "Unknown error"}`);
        process.exit(1);
      });
    }
  }
});

// src/claude/hooks/security/pre-write-guard.ts
async function main2() {
  try {
    const { parseClaudeInput: parseClaudeInput2, outputResult: outputResult2 } = await Promise.resolve().then(() => (init_hooks(), hooks_exports));
    const input = await parseClaudeInput2();
    const preWriteGuard = new PreWriteGuard();
    const result = await preWriteGuard.execute(input);
    outputResult2(result);
  } catch (error) {
    process.exit(0);
  }
}
var SENSITIVE_KEYWORDS, PROTECTED_PATHS, PreWriteGuard;
var init_pre_write_guard = __esm({
  "src/claude/hooks/security/pre-write-guard.ts"() {
    "use strict";
    init_cjs_shims();
    SENSITIVE_KEYWORDS = [
      ".env",
      "/secrets",
      "/.git/",
      "/.ssh"
    ];
    PROTECTED_PATHS = [
      ".moai/memory/"
    ];
    PreWriteGuard = class {
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
        return toolInput["file_path"] || toolInput["filePath"] || toolInput["path"] || null;
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
        for (const protectedPath of PROTECTED_PATHS) {
          if (filePath.includes(protectedPath)) {
            return false;
          }
        }
        return true;
      }
    };
    if (require.main === module) {
      main2().catch(() => {
        process.exit(0);
      });
    }
  }
});

// src/claude/hooks/workflow/file-monitor.ts
async function main3() {
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
var fs2, path2, import_events, FileMonitor;
var init_file_monitor = __esm({
  "src/claude/hooks/workflow/file-monitor.ts"() {
    "use strict";
    init_cjs_shims();
    fs2 = __toESM(require("fs"));
    path2 = __toESM(require("path"));
    import_events = require("events");
    FileMonitor = class extends import_events.EventEmitter {
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
          this.watcher = fs2.watch(
            this.projectRoot,
            { recursive: true },
            (eventType, filename) => {
              if (filename) {
                const fullPath = path2.join(this.projectRoot, filename);
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
        const parsedPath = path2.parse(filePath);
        const pathParts = filePath.split(path2.sep);
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
        const moaiPath = path2.join(this.projectRoot, ".moai");
        return fs2.existsSync(moaiPath);
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
    if (require.main === module) {
      main3().catch(() => {
      });
    }
  }
});

// src/claude/hooks/workflow/language-detector.ts
async function main4() {
  try {
    const detector = new LanguageDetector();
    const result = await detector.execute({});
    if (result.message) {
      console.log(result.message);
    }
  } catch (error) {
  }
}
var fs3, path3, DEFAULT_MAPPINGS, LanguageDetector;
var init_language_detector = __esm({
  "src/claude/hooks/workflow/language-detector.ts"() {
    "use strict";
    init_cjs_shims();
    fs3 = __toESM(require("fs"));
    path3 = __toESM(require("path"));
    DEFAULT_MAPPINGS = {
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
    LanguageDetector = class {
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
        const rootPath = path3.resolve(this.projectRoot);
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
        return fs3.existsSync(path3.join(this.projectRoot, filename));
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
          const entries = fs3.readdirSync(dir, { withFileTypes: true });
          for (const entry of entries) {
            const fullPath = path3.join(dir, entry.name);
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
        const mappingPath = path3.join(this.projectRoot, ".moai", "config", "language_mappings.json");
        try {
          if (fs3.existsSync(mappingPath)) {
            const data = fs3.readFileSync(mappingPath, "utf-8");
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
    if (require.main === module) {
      main4().catch(() => {
      });
    }
  }
});

// src/claude/hooks/workflow/test-runner.ts
async function main5() {
  const testRunner = new TestRunner();
  const result = await testRunner.execute({});
  if (result.message) {
    console.log(result.message);
  }
  process.exit(0);
}
var import_child_process, fs4, path4, TIMEOUT_SECONDS, TestRunner;
var init_test_runner = __esm({
  "src/claude/hooks/workflow/test-runner.ts"() {
    "use strict";
    init_cjs_shims();
    import_child_process = require("child_process");
    fs4 = __toESM(require("fs"));
    path4 = __toESM(require("path"));
    TIMEOUT_SECONDS = 3e5;
    TestRunner = class {
      name = "test-runner";
      projectRoot;
      constructor(projectRoot) {
        this.projectRoot = projectRoot || process.cwd();
      }
      async execute(input) {
        return {
          success: true,
          message: "Stop Hook: \uBE44\uD65C\uC131\uD654\uB428 - \uD14C\uC2A4\uD2B8\uB294 /moai:2-build \uB2E8\uACC4\uC5D0\uC11C\uB9CC \uC2E4\uD589\uB429\uB2C8\uB2E4."
        };
      }
      /**
       * Run all detected test commands
       */
      async runTests() {
        const commands = this.collectCommands();
        const results = [];
        for (const command of commands) {
          const result = await this.runCommand(command);
          results.push(result);
        }
        return results;
      }
      /**
       * Detect and collect available test commands
       */
      collectCommands() {
        const commands = [];
        const pytestCommand = this.detectPytest();
        if (pytestCommand) {
          commands.push(pytestCommand);
        }
        const npmCommand = this.detectNpm();
        if (npmCommand) {
          commands.push(npmCommand);
        }
        const goCommand = this.detectGo();
        if (goCommand) {
          commands.push(goCommand);
        }
        const cargoCommand = this.detectCargo();
        if (cargoCommand) {
          commands.push(cargoCommand);
        }
        return commands;
      }
      /**
       * Detect pytest command
       */
      detectPytest() {
        const testsDir = path4.join(this.projectRoot, "tests");
        if (fs4.existsSync(testsDir) && this.commandExists("pytest")) {
          return {
            name: "pytest",
            args: ["python", "-m", "pytest", "-q"]
          };
        }
        return null;
      }
      /**
       * Detect npm test command
       */
      detectNpm() {
        const packageJson = path4.join(this.projectRoot, "package.json");
        if (fs4.existsSync(packageJson) && this.commandExists("npm")) {
          return {
            name: "npm test",
            args: ["npm", "test", "--", "--watch=false"]
          };
        }
        return null;
      }
      /**
       * Detect go test command
       */
      detectGo() {
        const goMod = path4.join(this.projectRoot, "go.mod");
        if (fs4.existsSync(goMod) && this.commandExists("go")) {
          return {
            name: "go test",
            args: ["go", "test", "./..."]
          };
        }
        return null;
      }
      /**
       * Detect cargo test command
       */
      detectCargo() {
        const cargoToml = path4.join(this.projectRoot, "Cargo.toml");
        if (fs4.existsSync(cargoToml) && this.commandExists("cargo")) {
          return {
            name: "cargo test",
            args: ["cargo", "test", "--quiet"]
          };
        }
        return null;
      }
      /**
       * Check if command exists in PATH
       */
      commandExists(command) {
        try {
          const { execSync } = require("child_process");
          execSync(`which ${command}`, { stdio: "ignore" });
          return true;
        } catch {
          return false;
        }
      }
      /**
       * Run a single test command
       */
      async runCommand(command) {
        const startTime = Date.now();
        return new Promise((resolve2) => {
          const proc = (0, import_child_process.spawn)(command.args[0], command.args.slice(1), {
            cwd: this.projectRoot,
            stdio: "pipe"
          });
          let stdout = "";
          let stderr = "";
          proc.stdout?.on("data", (data) => {
            stdout += data.toString();
          });
          proc.stderr?.on("data", (data) => {
            stderr += data.toString();
          });
          const timeout = setTimeout(() => {
            proc.kill();
            resolve2({
              runner: command.name,
              exitCode: 124,
              stdout: stdout.trim(),
              stderr: `${command.name} timed out after ${TIMEOUT_SECONDS / 1e3}s`,
              duration: Date.now() - startTime
            });
          }, TIMEOUT_SECONDS);
          proc.on("close", (code) => {
            clearTimeout(timeout);
            resolve2({
              runner: command.name,
              exitCode: code || 0,
              stdout: stdout.trim(),
              stderr: stderr.trim(),
              duration: Date.now() - startTime
            });
          });
          proc.on("error", (error) => {
            clearTimeout(timeout);
            resolve2({
              runner: command.name,
              exitCode: 1,
              stdout: stdout.trim(),
              stderr: `${command.name} failed: ${error.message}`,
              duration: Date.now() - startTime
            });
          });
        });
      }
      /**
       * Generate test report from results
       */
      generateReport(results) {
        const lines = [];
        lines.push("\u{1F9EA} Test Results:");
        lines.push("");
        for (const result of results) {
          const status = result.exitCode === 0 ? "\u2705" : "\u274C";
          const duration = (result.duration / 1e3).toFixed(2);
          lines.push(`${status} ${result.runner} (${duration}s)`);
          if (result.exitCode !== 0) {
            lines.push(`   Error: ${result.stderr}`);
          }
          if (result.stdout) {
            lines.push(`   Output: ${result.stdout.substring(0, 200)}${result.stdout.length > 200 ? "..." : ""}`);
          }
          lines.push("");
        }
        const passed = results.filter((r) => r.exitCode === 0).length;
        const total = results.length;
        lines.push(`\u{1F4CA} Summary: ${passed}/${total} test suites passed`);
        return lines.join("\n");
      }
      /**
       * Get available test commands
       */
      getAvailableCommands() {
        return this.collectCommands();
      }
    };
    if (require.main === module) {
      main5().catch(() => {
        process.exit(0);
      });
    }
  }
});

// src/claude/hooks/session/session-notice.ts
async function main6() {
  try {
    const notifier = new SessionNotifier();
    const result = await notifier.execute({});
    if (result.message) {
      console.log(result.message);
    }
  } catch (error) {
  }
}
var import_child_process2, fs5, path5, SessionNotifier;
var init_session_notice = __esm({
  "src/claude/hooks/session/session-notice.ts"() {
    "use strict";
    init_cjs_shims();
    import_child_process2 = require("child_process");
    fs5 = __toESM(require("fs"));
    path5 = __toESM(require("path"));
    SessionNotifier = class {
      name = "session-notice";
      projectRoot;
      moaiConfigPath;
      constructor(projectRoot) {
        this.projectRoot = projectRoot || process.cwd();
        this.moaiConfigPath = path5.join(this.projectRoot, ".moai", "config.json");
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
          projectName: path5.basename(this.projectRoot),
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
          path5.join(this.projectRoot, ".moai"),
          path5.join(this.projectRoot, ".claude", "commands", "moai")
        ];
        return requiredPaths.every((p) => fs5.existsSync(p));
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
          if (!fs5.existsSync(path5.join(this.projectRoot, filePath))) {
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
          if (fs5.existsSync(this.moaiConfigPath)) {
            const configData = fs5.readFileSync(this.moaiConfigPath, "utf-8");
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
          if (fs5.existsSync(this.moaiConfigPath)) {
            const configData = fs5.readFileSync(this.moaiConfigPath, "utf-8");
            const config = JSON.parse(configData);
            return config.pipeline?.current_stage || "unknown";
          }
        } catch (error) {
        }
        const specsDir = path5.join(this.projectRoot, ".moai", "specs");
        if (fs5.existsSync(specsDir)) {
          const hasSpecs = fs5.readdirSync(specsDir).some((dir) => fs5.existsSync(path5.join(specsDir, dir, "spec.md")));
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
        const specsDir = path5.join(this.projectRoot, ".moai", "specs");
        if (!fs5.existsSync(specsDir)) {
          return { total: 0, completed: 0 };
        }
        try {
          const specDirs = fs5.readdirSync(specsDir).filter((name) => fs5.statSync(path5.join(specsDir, name)).isDirectory()).filter((name) => name.startsWith("SPEC-"));
          const totalSpecs = specDirs.length;
          let completed = 0;
          for (const specDir of specDirs) {
            const specPath = path5.join(specsDir, specDir, "spec.md");
            const planPath = path5.join(specsDir, specDir, "plan.md");
            if (fs5.existsSync(specPath) && fs5.existsSync(planPath)) {
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
        return new Promise((resolve2) => {
          const proc = (0, import_child_process2.spawn)("git", args, {
            cwd: this.projectRoot,
            stdio: "pipe"
          });
          let stdout = "";
          proc.stdout?.on("data", (data) => {
            stdout += data.toString();
          });
          const timeout = setTimeout(() => {
            proc.kill();
            resolve2(null);
          }, 2e3);
          proc.on("close", (code) => {
            clearTimeout(timeout);
            if (code === 0) {
              resolve2(stdout.trim());
            } else {
              resolve2(null);
            }
          });
          proc.on("error", () => {
            clearTimeout(timeout);
            resolve2(null);
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
    if (require.main === module) {
      main6().catch(() => {
      });
    }
  }
});

// src/claude/hooks/index.ts
var hooks_exports = {};
__export(hooks_exports, {
  FileMonitor: () => FileMonitor,
  HookSystem: () => HookSystem,
  LanguageDetector: () => LanguageDetector,
  PolicyBlock: () => PolicyBlock,
  PreWriteGuard: () => PreWriteGuard,
  SessionNotifier: () => SessionNotifier,
  SteeringGuard: () => SteeringGuard,
  TestRunner: () => TestRunner,
  outputResult: () => outputResult,
  parseClaudeInput: () => parseClaudeInput
});
function parseClaudeInput() {
  return new Promise((resolve2, reject) => {
    let data = "";
    process.stdin.on("data", (chunk) => {
      data += chunk.toString();
    });
    process.stdin.on("end", () => {
      try {
        if (data.trim()) {
          const parsed = JSON.parse(data);
          resolve2(parsed);
        } else {
          resolve2({});
        }
      } catch (error) {
        reject(new Error(`Invalid JSON input: ${error instanceof Error ? error.message : "Unknown error"}`));
      }
    });
    process.stdin.on("error", (error) => {
      reject(error);
    });
  });
}
function outputResult(result) {
  if (result.blocked) {
    console.error(`BLOCKED: ${result.message}`);
    process.exit(2);
  } else if (!result.success) {
    console.error(`ERROR: ${result.message}`);
    process.exit(result.exitCode || 1);
  } else if (result.message) {
    console.log(result.message);
  }
  process.exit(0);
}
var path6, fs6, HookSystem;
var init_hooks = __esm({
  "src/claude/hooks/index.ts"() {
    "use strict";
    init_cjs_shims();
    path6 = __toESM(require("path"));
    fs6 = __toESM(require("fs"));
    init_types();
    init_steering_guard();
    init_policy_block();
    init_pre_write_guard();
    init_file_monitor();
    init_language_detector();
    init_test_runner();
    init_session_notice();
    HookSystem = class {
      hooks = /* @__PURE__ */ new Map();
      config;
      constructor(config) {
        this.config = {
          enabled: true,
          timeout: 1e4,
          // 10 seconds
          disabledHooks: [],
          security: {
            allowedCommands: ["git", "npm", "node", "python", "pytest", "go", "cargo"],
            blockedPatterns: ["rm -rf", "sudo", "chmod 777"],
            requireApproval: ["--force", "--hard"]
          },
          ...config
        };
      }
      /**
       * Register a hook
       */
      registerHook(hook) {
        if (this.config.disabledHooks.includes(hook.name)) {
          return;
        }
        this.hooks.set(hook.name, hook);
      }
      /**
       * Execute a specific hook
       */
      async executeHook(hookName, input) {
        if (!this.config.enabled) {
          return { success: true, message: "Hooks disabled" };
        }
        const hook = this.hooks.get(hookName);
        if (!hook) {
          return { success: false, message: `Hook ${hookName} not found` };
        }
        try {
          const result = await Promise.race([
            hook.execute(input),
            this.createTimeoutPromise()
          ]);
          return result;
        } catch (error) {
          return {
            success: false,
            message: `Hook ${hookName} failed: ${error instanceof Error ? error.message : "Unknown error"}`,
            exitCode: 1
          };
        }
      }
      /**
       * Execute all registered hooks
       */
      async executeAllHooks(input) {
        const results = [];
        for (const [name, hook] of this.hooks) {
          try {
            const result = await this.executeHook(name, input);
            results.push(result);
            if (result.blocked) {
              break;
            }
          } catch (error) {
            results.push({
              success: false,
              message: `Hook ${name} failed: ${error instanceof Error ? error.message : "Unknown error"}`,
              exitCode: 1
            });
          }
        }
        return results;
      }
      /**
       * Load configuration from file
       */
      static async loadConfig(projectRoot) {
        const configPath = path6.join(projectRoot, ".moai", "config", "hooks.json");
        try {
          if (fs6.existsSync(configPath)) {
            const configData = fs6.readFileSync(configPath, "utf-8");
            const config = JSON.parse(configData);
            return config;
          }
        } catch (error) {
        }
        return {
          enabled: true,
          timeout: 1e4,
          disabledHooks: [],
          security: {
            allowedCommands: ["git", "npm", "node", "python", "pytest", "go", "cargo"],
            blockedPatterns: ["rm -rf", "sudo", "chmod 777"],
            requireApproval: ["--force", "--hard"]
          }
        };
      }
      /**
       * Get list of registered hooks
       */
      getRegisteredHooks() {
        return Array.from(this.hooks.keys());
      }
      /**
       * Check if hook is registered
       */
      hasHook(hookName) {
        return this.hooks.has(hookName);
      }
      createTimeoutPromise() {
        return new Promise((_, reject) => {
          setTimeout(() => {
            reject(new Error(`Hook execution timed out after ${this.config.timeout}ms`));
          }, this.config.timeout);
        });
      }
    };
  }
});

// src/claude/hooks/security/policy-block.ts
var policy_block_exports = {};
__export(policy_block_exports, {
  PolicyBlock: () => PolicyBlock,
  main: () => main7
});
module.exports = __toCommonJS(policy_block_exports);
async function main7() {
  try {
    const { parseClaudeInput: parseClaudeInput2, outputResult: outputResult2 } = await Promise.resolve().then(() => (init_hooks(), hooks_exports));
    const input = await parseClaudeInput2();
    const policyBlock = new PolicyBlock();
    const result = await policyBlock.execute(input);
    outputResult2(result);
  } catch (error) {
    console.error(`ERROR policy_block: ${error instanceof Error ? error.message : "Unknown error"}`);
    process.exit(1);
  }
}
var DANGEROUS_COMMANDS, ALLOWED_PREFIXES, PolicyBlock;
var init_policy_block = __esm({
  "src/claude/hooks/security/policy-block.ts"() {
    init_cjs_shims();
    DANGEROUS_COMMANDS = [
      "rm -rf /",
      "rm -rf --no-preserve-root",
      "sudo rm",
      "dd if=/dev/zero",
      ":(){:|:&};:",
      "mkfs."
    ];
    ALLOWED_PREFIXES = [
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
    PolicyBlock = class {
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
          console.error("NOTICE: \uB4F1\uB85D\uB418\uC9C0 \uC54A\uC740 \uBA85\uB839\uC785\uB2C8\uB2E4. \uD544\uC694 \uC2DC settings.json \uC758 allow \uBAA9\uB85D\uC744 \uAC31\uC2E0\uD558\uC138\uC694.");
        }
        return { success: true };
      }
      /**
       * Extract command from tool input
       */
      extractCommand(toolInput) {
        const raw = toolInput["command"] || toolInput["cmd"];
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
    if (require.main === module) {
      main7().catch((error) => {
        console.error(`ERROR policy_block: ${error instanceof Error ? error.message : "Unknown error"}`);
        process.exit(1);
      });
    }
  }
});
init_policy_block();
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  PolicyBlock,
  main
});
