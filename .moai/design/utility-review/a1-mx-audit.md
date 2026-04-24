# A1 — @MX TAG System Audit

**Auditor:** Expert-Backend A1
**Date:** 2026-04-23
**Scope:** `internal/hook/mx/` (types.go, config.go, validator.go) + integration tests

---

## Executive Summary

- **LOC audited:** 2,264 (primary: 806; tests: 1,458)
- **Overall health:** 5.5/10
- **Test coverage:** 92.6% — well above the 85% target
- **Race detector:** PASS (`go test -race`)

### Top 5 Issues

1. **Method receiver blindspot (Critical):** `exportedFuncRe` only matches `^func [A-Z]`, missing all exported method declarations (`func (r *T) Method()`). P2 goroutine violations inside receiver methods are never detected.
2. **fan_in substring false positives (High):** `countFanIn` uses `strings.Count(data, funcName)` without word boundaries. `funcName="New"` matches `NewContext`, `Renew`, `RenewToken` — inflating counts and generating spurious P1 violations.
3. **Brace-in-string literal bug (High):** The depth counter in `extractFunctions` processes every character including those inside string literals (`s := "{ bad"`). A single unmatched `{` in a string causes `lineCount` to be wrong, producing false P3 violations.
4. **Unbounded goroutine fan-out (High):** `ValidateFiles` spawns one goroutine per file path with no semaphore. On large projects this can create thousands of goroutines, each capable of spawning a grep subprocess.
5. **@MX:REASON enforcement missing (Medium):** The protocol mandates `@MX:REASON` for every WARN and ANCHOR tag, but the validator never checks for its presence — a core feature gap against the spec.

### Verdict: **REFACTOR**

The architecture is sound (interface, types, config separation), but the core analysis engine (`extractFunctions` + `countFanIn`) needs targeted rewrites. A full rewrite is not warranted.

---

## D1 — Correctness Bugs

| # | Severity | Location | Issue | Proposed Fix |
|---|----------|----------|-------|--------------|
| B1 | **Critical** | `validator.go:19` `exportedFuncRe` | Regex `^func\s+([A-Z]\w+)` never matches method receivers: `func (r *FileReport) P1Count()` → zero P2/P3/P4 violations for any goroutine inside a receiver method | Add separate `exportedMethodRe = regexp.MustCompile(`^func\s+\(\w+\s+\*?\w+\)\s+([A-Z]\w+)`)` and merge results in `extractFunctions` |
| B2 | **High** | `validator.go:284` `countFanIn` | `strings.Count(data, funcName)` has no word boundary — `funcName="Get"` matches `GetUser`, `Getter`, `forget`. Over-counts fan_in, generating spurious P1 violations | Replace with `regexp.MustCompile(`\b` + regexp.QuoteMeta(funcName) + `\b`).FindAllIndex(data, -1)` |
| B3 | **High** | `validator.go:224-230` `extractFunctions` | Brace depth counter iterates over every rune including those inside string literals and comments. `s := "{"` adds depth=1, offsetting the entire function end detection | Either use Go's `go/scanner` package for tokenization, or skip lines that are inside string literals using a simple state machine |
| B4 | **Medium** | `validator.go:196` `extractFunctions` | MX tag scan stops at the first non-`//` line when scanning backwards. A single blank line between the `@MX:WARN` comment and the function declaration causes the tag to be missed — false violation generated | Treat blank lines as transparent in the backwards scan (allow up to 1 blank line gap) |
| B5 | **Medium** | `validator.go:233` | Goroutine detection: `strings.Contains(bodyLine, "\tgo ")` matches `//\tgo func` in comments and `"\tgo "` inside string literals. False P2 violations possible | Check for goroutine patterns only after stripping comment prefix from the line |
| B6 | **Medium** | `validator.go:143-155` P4 logic | P4 fires once per exported function when `service_test.go` is absent, but if `service_test.go` exists and tests nothing, zero P4 violations are generated. The check is coarse (file existence only, not coverage) — understood as intentional but under-documented | Add inline comment clarifying the intentional coarseness; consider alternative: check if test file references the function name |
| B7 | **Low** | `config.go:104,106,111,115` | All four error paths in `ParseValidationConfig` silently return defaults including for non-`ErrNotExist` OS errors (permission denied) and malformed YAML. Operators cannot distinguish "file missing" from "file corrupt" | Return `fmt.Errorf("parse mx config: %w", err)` for YAML and unexpected OS errors; keep silent fallback only for `ErrNotExist` |

---

## D2 — API Design

### Strengths

- `Validator` interface in `types.go` is clean and appropriately minimal (2 methods).
- `FileReport.Error` field rather than returning an error from `ValidateFile` is a deliberate design that matches the "observation-only, never panics" contract. This is correct for hook integrations.
- `ValidationConfig` is flat and YAML-decodable without custom marshaling logic — idiomatic Go.

### Issues

| # | Severity | Location | Issue | Proposed Fix |
|---|----------|----------|-------|--------------|
| A1 | **High** | `validator.go:40` `NewValidator` | `analyzer any` parameter is permanently `nil` everywhere in the codebase (Grep fallback is always used). The `any` type leaks an unimplemented future extension point into the public API surface with no type enforcement | Either remove the parameter (breaking change, deferred to v3) or define `type Analyzer interface{ AnalyzeFile(...) }` and document the placeholder contract explicitly |
| A2 | **Medium** | `validator.go:258` `countFanIn` | `countFanIn` is tightly coupled to the filesystem (spawns grep). There is no way to inject a mock or alternative fan_in source for unit testing without a real project directory. This limits test determinism | Extract a `FanInCounter` interface: `type FanInCounter interface { Count(ctx, funcName, currentFile string) int }` and inject via `mxValidator` struct |
| A3 | **Medium** | `types.go:152-163` | `Validator` interface has no `io.Closer` or lifecycle methods. The `mxValidator` spawns goroutines in `ValidateFiles` (the closer goroutine). There is no way for the caller to cancel in-flight goroutines if they are no longer needed | The current design relies on context cancellation alone, which is correct for the hook use case. Acceptable as-is but worth documenting explicitly |
| A4 | **Low** | `validator.go:26-33` | `mxValidator` struct has `analyzer any` and `fanInThreshold int` as unexported fields, but `fanInThreshold` is always set to 3 in `NewValidator` and never configurable from `ValidationConfig`. The config system and the validator are not connected | Wire `ValidationConfig` thresholds into the validator constructor or expose a `WithThreshold(n int)` functional option |

---

## D3 — Performance

| # | Severity | Location | Issue | Proposed Fix |
|---|----------|----------|-------|--------------|
| P1 | **High** | `validator.go:307-380` `ValidateFiles` | Unbounded goroutine fan-out: `len(filePaths)` goroutines spawned, each capable of spawning N grep subprocesses (one per exported function without ANCHOR). For a project with 500 files × 10 functions = 5,000 grep subprocesses, all running concurrently | Add a worker pool: `sem := make(chan struct{}, runtime.NumCPU()*2)` wrapping the `go func(fp string)` body |
| P2 | **High** | `validator.go:264` `countFanIn` | Per-function grep call: for every exported function missing ANCHOR, a new `exec.CommandContext("grep", ...)` is spawned. This is O(N_functions) process forks per file. For a file with 20 exported functions, 20 grep processes run sequentially inside the goroutine | Pre-build a word-frequency map for the whole project once (single grep or single `os.ReadFile` pass) and cache in the validator. Use `sync.Once` for lazy initialization |
| P3 | **Medium** | `validator.go:264` `countFanIn` | The grep call re-reads the entire project for every function name. For a project with 10,000 `.go` files this is extremely expensive on cold disk cache | Build a single inverted index at validator construction time or on first use: `map[funcName]int` keyed by all identifiers found in the project |
| P4 | **Medium** | `validator.go:274-292` `countFanIn` | After grep lists matching files, each file is read again with `os.ReadFile` for `strings.Count`. This is double I/O (grep scan + Go re-read). Files already read for `ValidateFile` are re-read here | Pass the already-read file content to countFanIn to avoid re-reads for the current file; cache other files in a sync.Map |
| P5 | **Low** | `validator.go:19,22` | Both regexes (`exportedFuncRe`, `goroutineRe`) are compiled at package init with `regexp.MustCompile` — this is correct. No performance issue here. | N/A (this is a strength) |

**Complexity Analysis:**

Current: `O(F × G × P)` where F = files validated, G = goroutines, P = project files scanned by grep.
For moai-adk-go itself (~200 Go files, ~15 exported funcs/file): ~3,000 grep invocations per full scan.
Target: `O(F × 1)` with pre-built project index.

---

## D4 — Testing Quality

**Coverage: 92.6%** — exceeds the 85% target. Race detector passes.

### Strengths

- All tests use `t.TempDir()` for isolation — correct.
- Table-driven tests are used throughout (`TestValidateFile_ANCHORDetection`, `TestValidateFile_GoroutineDetection`).
- Integration coverage includes timeout, non-Go files, non-existent files, and read-only invariant.
- The read-only invariant test (`TestValidateFile_ReadOnly` with SHA-256 checksum) is excellent.

### Gaps

| # | Severity | Location | Gap | Recommended Test |
|---|----------|----------|-----|-----------------|
| T1 | **Critical** | `validator_test.go` | Zero tests for exported method receivers with goroutines. The B1 bug (method blindspot) has no test coverage | Add `TestValidateFile_MethodWithGoroutine`: `func (w *Worker) Start() { go func(){}() }` → expect P2 violation |
| T2 | **High** | `validator_test.go` | No test for substring false positive in `countFanIn`. `funcName="New"` in a project containing `NewUser`, `NewContext`, `RenewToken` — all counted as callers | Add `TestCountFanIn_SubstringFalsePositive` |
| T3 | **High** | `validator_test.go` | No test for brace-in-string-literal. `s := "{"` should not affect `lineCount` | Add `TestExtractFunctions_BraceInStringLiteral` |
| T4 | **Medium** | `validator_test.go` | No test for blank line between `@MX:WARN` tag and function declaration (B4 bug) | Add `TestExtractFunctions_BlankLineBetweenTagAndFunc` |
| T5 | **Medium** | `validator_test.go` | `TestValidateFiles_TimeoutPartialResults` uses 1ns timeout which may pass before any goroutine checks `ctx.Done()`. Test is non-deterministic | Use a slow validator mock or add artificial delay |
| T6 | **Medium** | `validator_test.go` | No benchmark tests. `countFanIn` is the hot path but has no `func BenchmarkCountFanIn(b *testing.B)` | Add `BenchmarkValidateFiles_LargeProject` with synthetic 50-file project |
| T7 | **Low** | `config_test.go` | No test for malformed YAML (B7 gap): `invalid: yaml: {` should ideally produce an error from `ParseValidationConfig` | Add `TestParseValidationConfig_MalformedYAML` |

---

## D5 — Hook Integration

| # | Severity | Location | Issue | Notes |
|---|----------|----------|-------|-------|
| H1 | **High** | `session_end.go:167` `getModifiedGoFiles` | `git diff --name-only HEAD` discovers files modified relative to the last commit. Untracked new files (never committed) are silently skipped. If a developer creates a new `.go` file in this session without committing, it will not be validated | Use `git diff --name-only HEAD` combined with `git ls-files --others --exclude-standard --cached` to include untracked+staged files |
| H2 | **Medium** | `session_end.go:101-104` | The 4s timeout is hardcoded in `session_end.go` and not read from `SessionEndConfig.TimeoutMs`. The `ValidationConfig` infrastructure exists but is not wired to `validateMxTags` | Pass `cfg.SessionEnd.TimeoutMs` to the function or read from `mx.yaml` |
| H3 | **Low** | `post_tool.go` (integration) | `NewPostToolHandlerWithMxValidator` creates a fresh `mx.NewValidator(nil, projectRoot)` with no shared project index. Each handler invocation starts from zero. For hot loop scenarios (multiple Write calls per second), there is no caching between calls | Share a validator instance at the process level; the validator is stateless and safe for concurrent use |
| H4 | **Low** | `post_tool_mx_test.go:30` | Test creates file, writes goroutine code, then calls `Handle()` with a *different* content string (`"package svc\n"`). The test exercises the code path where the file on disk has a goroutine but the tool input content does not. This simulates the race window between Write completion and MX validation | Add a comment explaining this is intentional — the validator reads the file off disk, not the tool input content |

**State Persistence:** MX validation results are emitted to `slog` only. There is no `.moai/state/` persistence of violations. This is consistent with the "observation-only" contract. No state is leaked between sessions.

**Failure Recovery:** `ValidateFile` never returns a non-nil error (errors are embedded in `FileReport.Error`). `validateMxTags` never returns any value. The hook system cannot be broken by validator panics because there are none — the code is safe. `recover()` is not needed.

---

## D6 — Feature Gaps vs MX Protocol Spec

Source: `.claude/rules/moai/workflow/mx-tag-protocol.md`

| # | Severity | Feature | Status | Gap |
|---|----------|---------|--------|-----|
| F1 | **Critical** | Method receiver detection | **MISSING** | `exportedFuncRe` only matches top-level functions. All exported methods on types are invisible to the validator. See B1 |
| F2 | **High** | `@MX:REASON` enforcement | **MISSING** | Protocol states: `@MX:REASON` is MANDATORY for WARN and ANCHOR. The validator detects the absence of WARN/ANCHOR tags but never checks whether an existing WARN/ANCHOR has a paired `@MX:REASON` sub-line |
| F3 | **High** | Per-file limits enforcement | **MISSING** | `mx.yaml` defines `anchor_per_file: 3` and `warn_per_file: 5`. The validator does not read these limits and does not report when a file exceeds them. The reference doc (`mx-tag.md`) says "Demote excess ANCHOR tags" — this is undefined behavior |
| F4 | **High** | Cyclomatic complexity check (P2) | **MISSING** | Protocol: `@MX:WARN` for cyclomatic complexity >= 15 and `if_branches >= 8`. The validator only detects goroutine patterns for P2. Complexity analysis is entirely absent |
| F5 | **Medium** | `@MX:LEGACY` tag type | **MISSING** | Protocol lists `@MX:LEGACY` as a valid sub-key. Neither detection (existing LEGACY tags) nor suggestion (functions in files named `*_legacy.go`) is implemented |
| F6 | **Medium** | `@MX:SPEC` sub-line | **MISSING** | Protocol lists `@MX:SPEC` as an optional sub-line. The validator does not recognize or report on SPEC references in existing tags |
| F7 | **Medium** | `auto_tag` config flag | **MISSING** | `mx.yaml` supports `auto_tag: enable/disable`. The `ValidationConfig` struct has no `AutoTag` field. The flag is defined in the protocol but not modeled |
| F8 | **Low** | `@MX:TEST` sub-line | **MISSING** | Protocol lists `@MX:TEST` as a valid sub-key linking to test names. Not implemented |
| F9 | **Low** | `code_comments` language setting | **PARTIALLY MET** | Protocol says tag descriptions must respect `code_comments` from `language.yaml`. The validator detects tags in any language but has no config hook to enforce the language of generated violation messages |

**Implemented correctly:**
- `@MX:ANCHOR` detection (P1, fan_in >= 3) — implemented
- `@MX:WARN` detection (P2, goroutine patterns) — partially implemented (misses methods)
- `@MX:NOTE` detection (P3, >= 100 lines) — implemented
- `@MX:TODO` detection (P4, no test file) — implemented (coarse)
- `[AUTO]` prefix recognition — implicitly via `@MX:ANCHOR` string match

---

## D7 — Maintainability

| # | Severity | Location | Issue | Fix |
|---|----------|----------|-------|-----|
| M1 | **High** | `validator.go:178-252` `extractFunctions` (74 LOC) | Single function responsible for: (a) scanning function declarations, (b) backwards MX tag scan, (c) brace counting, (d) goroutine detection. Cyclomatic complexity ~12. Multiple responsibilities | Extract: `scanMXTags(lines []string, funcLine int) tagSet`, `measureFuncBody(lines []string, startLine int) (lineCount int, hasGoroutine bool)` |
| M2 | **Medium** | `validator.go:258-293` `countFanIn` (35 LOC) | External process invocation (`exec.CommandContext`) is embedded in the analysis layer. Testing requires a real filesystem and grep binary. Breaking the abstraction makes the code fragile on Windows (grep not available) | Inject `FanInCounter` interface (see A2); keep grep implementation as `grepFanInCounter` |
| M3 | **Medium** | `validator.go:397-467` `formatReport` (70 LOC) | Long formatting function with repeated `collectByPriority` → `range` → `fmt.Fprintf` pattern. Each priority section is copy-pasted with minor variation | Refactor to: `for _, p := range []Priority{P1, P2, P3, P4} { renderPrioritySection(&buf, report, p) }` |
| M4 | **Medium** | `validator.go:26-33` | `fanInThreshold: 3` is hardcoded in `NewValidator`. This value exists in `mx.yaml` as `fan_in_anchor: 3` but the threshold from config is never read by the validator | Read threshold from `ValidationConfig.Thresholds.FanInAnchor` or accept via constructor option |
| M5 | **Low** | `validator.go:82-84` `analyzeFile` | The `@MX:WARN` tag on `analyzeFile` says "Called in parallel via goroutines from ValidateFiles. Can be interrupted via ctx.Done()". This is accurate but misleading: `analyzeFile` is a regular method called from within a goroutine started in `ValidateFiles`. The comment would be clearer if it said "Must be safe for concurrent calls" | Update comment to reflect the correct framing |
| M6 | **Low** | `validator.go:217-218` | Brace depth initialization comment: "Find opening brace on function declaration line or next line" — but the code only checks the current line. Multi-line function signatures (common in Go) where `{` is on the next line are not handled by this initialization, causing depth to start at 0, and the loop condition `(depth > 0 || j == bodyStart)` handles it via `j == bodyStart` — but this is implicit and fragile | Make the behavior explicit: scan forward for `{` if not found on declaration line |

---

## Detailed Recommendations (Prioritized)

### R1 [Critical] — Fix method receiver blindspot in `extractFunctions`

**File:** `validator.go:178-252`
**Root cause:** `exportedFuncRe` only matches `^func\s+([A-Z]\w+)`. All exported methods — which constitute the majority of the public API in idiomatic Go — are invisible to P1/P2/P3/P4 detection.

**Impact:** ValidateFiles is called on the entire project but silently misses goroutines in receiver methods. Since most Go production code uses methods, the P2 violation coverage is effectively broken for typical codebases.

**Fix:**
```go
var exportedMethodRe = regexp.MustCompile(`^func\s+\(\w+\s+\*?\w+\)\s+([A-Z]\w+)`)

// In extractFunctions, add after the existing exportedFuncRe check:
if m := exportedMethodRe.FindStringSubmatch(strings.TrimSpace(line)); len(m) >= 2 {
    // same logic as top-level function handling
}
```

**Test to add:** `TestExtractFunctions_ExportedMethodWithGoroutine`

---

### R2 [High] — Fix fan_in substring false positives with word-boundary matching

**File:** `validator.go:264,284`
**Root cause:** `grep -l funcName` and `strings.Count(data, funcName)` match substrings. `funcName="Handle"` counts `HandleError`, `HandleRequest`, `HandleTimeout` as callers.

**Impact:** Functions with common short names (e.g., `New`, `Get`, `Set`, `Parse`) will have grossly inflated fan_in counts, generating P1 violations that should not exist.

**Fix (grep side):**
```go
// Use -w for word boundary in grep
cmd := exec.CommandContext(ctx, "grep", "-r", "--include=*.go", "-l", "-w", funcName, v.projectRoot)
```

**Fix (strings.Count side):**
```go
re := regexp.MustCompile(`\b` + regexp.QuoteMeta(funcName) + `\b`)
count := len(re.FindAllIndex(data, -1))
```

**Note:** The `mx-tag.md` reference doc explicitly says approximate fan_in is acceptable ("False positives are acceptable"). However, the current level of false positives (substring matching with no boundary) goes beyond "approximate" to "unreliable". Word-boundary grep is a one-character fix (`-w` flag).

---

### R3 [High] — Add goroutine worker pool in `ValidateFiles` to bound concurrency

**File:** `validator.go:307-380`
**Root cause:** One goroutine per file, each capable of spawning N grep subprocesses. No semaphore.

**Impact:** On a 500-file project, `ValidateFiles` spawns 500 goroutines simultaneously. Each goroutine may call `countFanIn` which spawns a grep subprocess. This creates up to 500 concurrent grep invocations on the filesystem, potentially overwhelming I/O and exhausting OS resources.

**Fix:**
```go
const maxWorkers = 8 // or runtime.NumCPU() * 2
sem := make(chan struct{}, maxWorkers)

for _, path := range filePaths {
    wg.Add(1)
    sem <- struct{}{} // acquire
    go func(fp string) {
        defer wg.Done()
        defer func() { <-sem }() // release
        // ... existing goroutine body
    }(path)
}
```

The existing `@MX:WARN` already documents this unbounded fan-out. This recommendation resolves the flagged issue.

---

### R4 [High] — Implement `@MX:REASON` presence check for existing WARN/ANCHOR tags

**File:** `validator.go:196-210`
**Root cause:** The protocol mandates `@MX:REASON` for every WARN and ANCHOR. The validator detects missing WARN/ANCHOR tags, but never validates that existing ones have a REASON sub-line.

**Fix in `funcInfo` struct:**
```go
type funcInfo struct {
    // existing fields ...
    hasAnchorReason bool
    hasWarnReason   bool
}
```

**Detection logic:**
```go
if strings.Contains(commentLine, "@MX:REASON") {
    // Check which tag this REASON belongs to by looking at sibling lines
    fn.hasAnchorReason = fn.hasAnchor
    fn.hasWarnReason = fn.hasWarn
}
```

**New violation type (or reuse existing P2/P1):**
```go
if fn.hasAnchor && !fn.hasAnchorReason {
    violations = append(violations, Violation{
        Priority:   P2, // or a new P1 sub-type
        MissingTag: "@MX:REASON",
        Reason:     "@MX:ANCHOR present but @MX:REASON sub-line missing",
    })
}
```

---

### R5 [Medium] — Wire `ValidationConfig` thresholds into `mxValidator`

**File:** `validator.go:40-46`, `config.go:98-121`
**Root cause:** `NewValidator` hardcodes `fanInThreshold: 3` and ignores the `ValidationConfig`. The config system and the validator are decoupled in a way that prevents runtime configuration.

**Fix:**
```go
func NewValidatorWithConfig(cfg *ValidationConfig, analyzer any, projectRoot string) Validator {
    threshold := 3
    if cfg != nil {
        // cfg.Thresholds.FanInAnchor when that field is added
    }
    return &mxValidator{
        analyzer:       analyzer,
        projectRoot:    projectRoot,
        fanInThreshold: threshold,
    }
}
```

---

### R6 [Medium] — Implement per-file limit enforcement

**File:** `validator.go:analyzeFile`
**Root cause:** `anchor_per_file: 3` is defined in `mx.yaml` but never enforced. Files can accumulate unlimited violations without triggering the demotion logic described in `mx-tag.md`.

**Fix:** After collecting all violations, apply per-file caps:
```go
func capViolations(violations []Violation, anchorLimit, warnLimit int) []Violation {
    p1Count, p2Count := 0, 0
    result := violations[:0]
    for _, v := range violations {
        switch v.Priority {
        case P1:
            if p1Count < anchorLimit {
                result = append(result, v)
                p1Count++
            }
        case P2:
            if p2Count < warnLimit {
                result = append(result, v)
                p2Count++
            }
        default:
            result = append(result, v)
        }
    }
    return result
}
```

---

## Proposed Refactor Sketch

Since the verdict is REFACTOR (not rewrite), the package layout is preserved. Only the `validator.go` internals change:

```
internal/hook/mx/
├── types.go          (unchanged)
├── config.go         (minor: return error for malformed YAML)
├── validator.go      (split into sub-units below)
│   ├── extractor.go  (extractFunctions + scanMXTags + measureFuncBody)
│   ├── fanin.go      (FanInCounter interface + grepFanInCounter impl)
│   ├── analyze.go    (analyzeFile using extracted helpers)
│   ├── report.go     (formatReport refactored to loop over priorities)
│   └── pool.go       (worker pool for ValidateFiles)
├── config_test.go    (unchanged)
└── validator_test.go (add: T1-T6 from D4 gaps)
```

**Key invariants to preserve:**
- `Validator` interface unchanged (API stability)
- `ValidateFile` never returns non-nil error
- Read-only contract maintained
- Context cancellation propagation

---

## Open Questions for v3 Architect

1. **AST vs Grep trade-off:** The `analyzer any` field suggests future AST-grep integration. When this lands, which violations move from Grep to AST? Specifically: should `extractFunctions` be replaced by `go/ast.Inspect` which handles both methods and functions natively and correctly handles string literals?

2. **Sidecar JSON generation:** The audit found no evidence of sidecar JSON (`.moai/state/*.json` per-file tag persistence). Is this planned for v3? If so, the `ValidationReport` struct would need a `WriteSidecar(path string) error` method.

3. **Cross-file ANCHOR demotion:** The protocol says "Demote excess ANCHOR by lowest fan_in when per-file limit exceeded." This requires a sorted list of existing ANCHOR tags across files — currently the validator is stateless between files. Should `ValidateFiles` maintain cross-file state for demotion decisions?

4. **Windows support:** `countFanIn` spawns `grep` which is not available on Windows. Is Windows support required? If yes, replace the external grep with a pure Go implementation (`filepath.WalkDir` + `regexp.FindAll`).

5. **P2 complexity check:** The protocol says P2 triggers for `cyclomatic complexity >= 15` or `if_branches >= 8`. Implementing this purely with regex/string analysis is fragile. Is the v3 plan to defer complexity checks to the AST analyzer, or add a heuristic branch counter to the current Grep fallback?

---

## References: File:Line Citations

| Finding | Location |
|---------|----------|
| B1 method blindspot | `validator.go:19` `exportedFuncRe` definition |
| B2 substring fan_in | `validator.go:264,284` `countFanIn` |
| B3 brace in string | `validator.go:224-230` depth counter loop |
| B4 blank line gap | `validator.go:194-210` backwards tag scan |
| B5 goroutine in comment | `validator.go:233` goroutine detection |
| B7 silent YAML error | `config.go:104,106,111,115` |
| A1 `analyzer any` | `validator.go:40` `NewValidator` signature |
| A2 no mock interface | `validator.go:258` `countFanIn` |
| A4 threshold not wired | `validator.go:44`, `config.go:71-93` |
| P1 unbounded goroutines | `validator.go:302-303` `@MX:WARN` comment + `validator.go:317-345` |
| P2 per-function grep | `validator.go:264` |
| H1 untracked files | `session_end.go:167` `git diff --name-only HEAD` |
| H2 hardcoded timeout | `session_end.go:101` `4*time.Second` |
| F1 method detection | `validator.go:19` |
| F2 REASON not checked | `validator.go:196-210` |
| F3 per-file limits | `validator.go:86-160` `analyzeFile` (no cap logic) |
| F4 complexity check | `validator.go:117-128` (only goroutine for P2) |
| M1 extractFunctions | `validator.go:178-252` |
| M3 formatReport | `validator.go:397-467` |
| M4 hardcoded threshold | `validator.go:44` |
| T1 missing method test | `validator_test.go` (absent) |
| T2 missing substring test | `validator_test.go` (absent) |
| T3 missing brace test | `validator_test.go` (absent) |
| T6 no benchmarks | `validator_test.go` (absent) |

---

*Generated by expert-backend A1 — Read-only audit, no source files modified.*
