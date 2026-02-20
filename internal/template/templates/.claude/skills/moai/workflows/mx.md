# Workflow: MX Tag Scan and Annotation

Purpose: Scan codebase and add @MX code-level annotations for AI agent context. Supports all MoAI-ADK languages (16 total) with automatic language detection.

## When to Use

- Legacy codebase without @MX tags
- Before major refactoring to mark danger zones
- After significant code changes to update annotations
- During `/moai sync` for MX validation

## Command

```
/moai mx [options]
```

## Flags

| Flag | Description |
|------|-------------|
| `--all` | Scan entire codebase (all languages, all P1+P2 files) |
| `--dry` | Preview only - show tags to add without modifying files |
| `--priority P1-P4` | Filter by priority level (default: all) |
| `--force` | Overwrite existing @MX tags |
| `--exclude pattern` | Additional exclude patterns (comma-separated) |
| `--lang go,py,ts` | Scan only specified languages (default: auto-detect) |
| `--threshold N` | Override fan_in threshold (default: 3) |
| `--no-discovery` | Skip Phase 0 codebase discovery |

## Priority Levels

| Priority | Condition | Tag Type |
|----------|-----------|----------|
| P1 | fan_in >= 3 callers | `@MX:ANCHOR` |
| P2 | goroutine/async, complexity >= 15 | `@MX:WARN` |
| P3 | magic constant, missing docstring | `@MX:NOTE` |
| P4 | missing test | `@MX:TODO` |

**Note**: Threshold lowered from 5 to 3 to capture more important functions.

## Workflow Phases

### Phase 0: Codebase Discovery (NEW)

**Purpose**: Detect project languages and load context before scanning.

**Steps**:
1. **Language Detection** (16 languages supported)

   Check indicator files in priority order:

   | Language | Indicator Files | Comment Prefix |
   |----------|----------------|----------------|
   | Go | go.mod, go.sum | `//` |
   | Python | pyproject.toml, setup.py, requirements.txt | `#` |
   | TypeScript | tsconfig.json, package.json (with typescript) | `//` |
   | JavaScript | package.json (without tsconfig) | `//` |
   | Rust | Cargo.toml, Cargo.lock | `//` |
   | Java | pom.xml, build.gradle, build.gradle.kts | `//` |
   | Kotlin | build.gradle.kts (with kotlin plugin) | `//` |
   | C# | .csproj, .sln, .fsproj | `//` |
   | Ruby | Gemfile, .ruby-version, Rakefile | `#` |
   | PHP | composer.json, composer.lock | `//` |
   | Elixir | mix.exs | `#` |
   | C++ | CMakeLists.txt, Makefile (with C++) | `//` |
   | Scala | build.sbt, build.sc | `//` |
   | R | DESCRIPTION, .Rproj, renv.lock | `#` |
   | Flutter/Dart | pubspec.yaml | `//` |
   | Swift | Package.swift, .xcodeproj | `//` |

2. **Project Context Loading**
   - Read `.moai/project/tech.md` for tech stack context
   - Read `.moai/project/structure.md` for architecture context
   - Read `.moai/project/product.md` for feature context
   - Read `README.md` for project overview

3. **Scan Scope Calculation**
   - Count files per language
   - Estimate token budget
   - Apply exclude patterns

**Output**:
```yaml
discovery:
  languages:
    - name: go
      files: 45
      enabled: true
    - name: python
      files: 23
      enabled: true
    - name: typescript
      files: 67
      enabled: true
  project_context:
    loaded: true
    tech_stack: "..."
    architecture: "..."
  scan_scope:
    total_files: 135
    estimated_tokens: 35000
```

### Pass 1: Full File Scan (IMPROVED)

**Purpose**: Scan all source files and generate priority queue.

**Complete Multi-Language Pattern Detection**:

| Language | ANCHOR (P1) | WARN (P2) | NOTE (P3) | TODO (P4) |
|----------|-------------|-----------|-----------|-----------|
| **Backend** |
| Go | fan_in >= 3 | `go func`, `go `, complexity >= 15 | magic constant, no godoc | no `*_test.go` |
| Python | fan_in >= 3 | `async def`, `threading`, complexity >= 15 | magic constant, no docstring | no `test_*.py` |
| Rust | fan_in >= 3 | `async fn`, `unsafe `, complexity >= 15 | magic constant, no doc | no `*test.rs` |
| Java | fan_in >= 3 | `new Thread`, `Executor`, complexity >= 15 | magic constant, no Javadoc | no `*Test.java` |
| Kotlin | fan_in >= 3 | `GlobalScope`, `runBlocking`, complexity >= 15 | magic constant, no KDoc | no `*Test.kt` |
| C# | fan_in >= 3 | `Task.Run`, `Thread.`, complexity >= 15 | magic constant, no XML doc | no `*Test.cs` |
| Ruby | fan_in >= 3 | `Thread.new`, complexity >= 15 | magic constant, no comment | no `*_test.rb` |
| PHP | fan_in >= 3 | `async `, complexity >= 15 | magic constant, no PHPDoc | no `*Test.php` |
| Elixir | fan_in >= 3 | `Task.async`, `spawn`, complexity >= 15 | magic constant, no doc | no `*_test.exs` |
| C++ | fan_in >= 3 | `std::thread`, `new `, complexity >= 15 | magic constant, no comment | no `*test.cpp` |
| Scala | fan_in >= 3 | `Future.`, `new Thread`, complexity >= 15 | magic constant, no doc | no `*Test.scala` |
| **Frontend** |
| TypeScript | fan_in >= 3 | `Promise.all`, `async `, complexity >= 15 | magic constant, no JSDoc | no `*.test.ts` |
| JavaScript | fan_in >= 3 | `Promise.all`, `async `, complexity >= 15 | magic constant, no JSDoc | no `*.test.js` |
| **Data Science** |
| R | fan_in >= 3 | `parallel::`, complexity >= 15 | magic constant, no comment | no `*test.R` |
| Flutter/Dart | fan_in >= 3 | `Isolate.`, `Future.`, complexity >= 15 | magic constant, no doc | no `*_test.dart` |
| **Mobile** |
| Swift | fan_in >= 3 | `Task.`, `DispatchQueue`, complexity >= 15 | magic constant, no doc | no `*Test.swift` |

**Steps**:
1. For each enabled language:
   - Glob all source files using language-specific patterns
   - Fan-in analysis: Count function/method references across files
   - Complexity detection: Lines, branches, nesting depth
   - Pattern detection: Language-specific danger patterns
2. Build priority queue (ALL files included, ranked by score)
3. Output: Priority list P1-P4

### Pass 2: Selective Deep Read (IMPROVED)

**Purpose**: Read files and generate accurate tag descriptions.

**Expanded Scope**:
- **Previous**: P1 files only (fan_in >= 5)
- **Current**: P1 + P2 files (fan_in >= 3 OR complexity >= 15)

**Steps**:
1. For each P1 and P2 file:
   - Full file Read with context
   - Analyze function signatures and call patterns
   - Generate tag descriptions using project context
   - Use language-specific comment syntax

**Project Context Integration**:
- Tech stack information from `tech.md`
- Architecture patterns from `structure.md`
- Business domain from `product.md`

**Language-Specific Strategies**:
- **Go**: Extract godoc comments, analyze interface contracts
- **Python**: Extract docstrings, analyze class hierarchies, decorators
- **TypeScript**: Extract JSDoc, analyze type definitions, interfaces
- **Rust**: Extract docs, analyze trait bounds, lifetimes
- **Java**: Extract Javadoc, analyze class hierarchies, annotations
- **Kotlin**: Extract KDoc, analyze extension functions, coroutines
- **C#**: Extract XML docs, analyze async/await patterns
- **Ruby**: Extract comments, analyze metaprogramming, DSL patterns
- **PHP**: Extract PHPDoc, analyze PSR standards
- **Elixir**: Extract docs, analyze GenServer/GenStage patterns
- **C++**: Extract comments, analyze template metaprogramming
- **Scala**: Extract Scaladoc, analyze type classes, implicits
- **R**: Extract comments, analyze pipe operations, tidyverse patterns
- **Flutter/Dart**: Extract docs, analyze widget trees, state management
- **Swift**: Extract docs, analyze SwiftUI, Combine patterns
- **JavaScript**: Extract JSDoc, analyze CommonJS/ES modules

### Pass 3: Batch Edit

**Purpose**: Insert tags into files.

**Language Comment Syntax**:

| Language | Prefix | Example |
|----------|--------|---------|
| Go, Java, TS, JS, Rust, Kotlin, Swift, Scala, C++, C#, Dart | `//` | `// @MX:NOTE:` |
| Python, Ruby, R, Elixir | `#` | `# @MX:WARN:` |
| Haskell | `--` | `-- @MX:ANCHOR:` |

**Steps**:
1. One Edit call per file
2. All tags for a given file inserted in single operation
3. Preserve existing @MX tags (unless --force)
4. Generate final report

## Output

After completion, generates report:

```markdown
## @MX Tag Report

### Discovery Summary
- Languages detected: Go (45), Python (23), TypeScript (67)
- Project context: Loaded from .moai/project/tech.md
- Scan scope: 135 files, 35,000 estimated tokens

### Summary
- Files scanned: 135
- Tags added: 87
- Tags updated: 23
- Tags skipped (existing): 12

### Tags by Type
- @MX:ANCHOR: 32 (P1) - High fan_in functions (>= 3 callers)
- @MX:WARN: 18 (P2) - Complex/dangerous patterns
- @MX:NOTE: 28 (P3) - Context annotations
- @MX:TODO: 9 (P4) - Missing tests

### Tags by Language
- Go: 32 tags (15 ANCHOR, 5 WARN, 8 NOTE, 4 TODO)
- Python: 18 tags (5 ANCHOR, 4 WARN, 7 NOTE, 2 TODO)
- TypeScript: 25 tags (10 ANCHOR, 6 WARN, 7 NOTE, 2 TODO)
- Rust: 5 tags (1 ANCHOR, 2 WARN, 1 NOTE, 1 TODO)
- Java: 4 tags (1 ANCHOR, 1 WARN, 1 NOTE, 1 TODO)
- Kotlin: 2 tags (0 ANCHOR, 0 WARN, 2 NOTE, 0 TODO)
- JavaScript: 1 tag (0 ANCHOR, 0 WARN, 1 NOTE, 0 TODO)

### Files Modified
- internal/core/handler.go: +5 tags
- src/api/server.ts: +4 tags
- lib/utils/helper.py: +3 tags

### Attention Required
- High fan_in functions (>= 10 callers): handler.go:ProcessRequest
- Complex functions (complexity >= 20): server.ts:HandleConnection
```

## Integration with Other Workflows

### With /moai sync

During sync phase, MX validation runs automatically:
1. Scan files modified since last sync (all languages)
2. Check for missing @MX tags in modified functions
3. Add tags if `--skip-mx` flag not provided
4. Include tag changes in sync report

### With /moai run

During DDD ANALYZE phase:
1. If codebase has zero @MX tags, 3-Pass auto-triggers
2. Existing tags are validated and updated
3. New tags added for new code

## Configuration

Project settings in `.mx.yaml` (complete 16-language support):

```yaml
mx:
  version: "2.1"

  languages:
    go:
      enabled: auto
      patterns: ["*.go"]
      exclude: ["*_generated.go", "vendor/**"]
    python:
      enabled: auto
      patterns: ["*.py"]
      exclude: ["**/__pycache__/**", "**/venv/**"]
    typescript:
      enabled: auto
      patterns: ["*.ts", "*.tsx"]
      exclude: ["**/node_modules/**", "**/*.d.ts"]
    # ... (13 more languages)

  thresholds:
    fan_in_anchor: 3  # Lowered from 5
    complexity_warn: 15
    branch_warn: 8
```

## Examples

```bash
# Scan entire codebase (all 16 languages)
/moai mx --all

# Preview tags without modifying files
/moai mx --dry

# Only P1 priority (high fan_in functions)
/moai mx --priority P1

# Force overwrite existing tags
/moai mx --all --force

# Exclude test files
/moai mx --all --exclude "**/*_test.go,**/*_test.py"

# Scan only Go and Python
/moai mx --all --lang go,python

# Lower threshold for more coverage
/moai mx --all --threshold 2

# Skip Phase 0 discovery
/moai mx --all --no-discovery
```

## Language-Specific Examples

### Go Project
```bash
/moai mx --all
# Scans *.go files
# Detects: go func, goroutines, godoc
# Adds: @MX:ANCHOR (fan_in >= 3), @MX:WARN (goroutines)
```

### Python Project
```bash
/moai mx --all
# Scans *.py files
# Detects: async def, threading, docstrings
# Adds: @MX:ANCHOR (fan_in >= 3), @MX:WARN (async/threading)
```

### TypeScript Project
```bash
/moai mx --all
# Scans *.ts, *.tsx files
# Detects: Promise.all, async/await, JSDoc
# Adds: @MX:ANCHOR (fan_in >= 3), @MX:WARN (Promise chains)
```

### Multi-Language Project (Go + Python + TypeScript)
```bash
/moai mx --all
# Auto-detects all 3 languages
# Scans: *.go, *.py, *.ts, *.tsx
# Generates language-specific tags for each
```

## Complete Language Support Matrix

| # | Language | File Patterns | Comment | Test Pattern | WARN Patterns |
|---|----------|---------------|---------|--------------|---------------|
| 1 | Go | *.go | `//` | *_test.go | go func, go  |
| 2 | Python | *.py | `#` | test_*.py | async def, threading |
| 3 | TypeScript | *.ts, *.tsx | `//` | *.test.ts | Promise.all, async  |
| 4 | Rust | *.rs | `//` | *test.rs | async fn, unsafe  |
| 5 | Java | *.java | `//` | *Test.java | new Thread, Executor |
| 6 | Kotlin | *.kt, *.kts | `//` | *Test.kt | GlobalScope, runBlocking |
| 7 | C# | *.cs | `//` | *Test.cs | Task.Run, Thread. |
| 8 | Ruby | *.rb | `#` | *_test.rb | Thread.new |
| 9 | PHP | *.php | `//` | *Test.php | async  |
| 10 | Elixir | *.ex, *.exs | `#` | *_test.exs | Task.async, spawn |
| 11 | C++ | *.cpp, *.cc, *.h | `//` | *test.cpp | std::thread, new  |
| 12 | Scala | *.scala | `//` | *Test.scala | Future., new Thread |
| 13 | R | *.R, *.r | `#` | *test.R | parallel:: |
| 14 | Flutter | *.dart | `//` | *_test.dart | Isolate., Future. |
| 15 | Swift | *.swift | `//` | *Test.swift | Task., DispatchQueue |
| 16 | JavaScript | *.js, *.jsx | `//` | *.test.js | Promise.all, async  |

---

Version: 2.1.0 (Complete 16-Language Support)
Last Updated: 2026-02-20
Source: SPEC-MX-001
