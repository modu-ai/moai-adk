---
id: SPEC-UTIL-001
title: "MX Validator Correctness + Tree-sitter 16-Language Complexity + Windows Go-Native Scan + @MX:REASON Enforcement"
version: "0.1.0"
status: implemented
created: 2026-04-24
updated: 2026-04-24
author: Wave v2.14 SPEC Writer
priority: P0 Critical
phase: "v2.14.0 — Phase 1 — Utility Hardening"
module: "internal/hook/mx/"
dependencies: []
related_problem: [IMP-V3U-001, IMP-V3U-005, IMP-V3U-007]
related_pattern: []
related_principle: []
related_decision: [D4, D5, D6]
related_theme: "v2.14.0 Utility Hardening"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "mx, validator, tree-sitter, multi-language, windows, v2.14"
---

# SPEC-UTIL-001: MX Validator Correctness + Tree-sitter 16-Language Complexity + Windows Go-Native Scan + @MX:REASON Enforcement

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.1 | 2026-04-24 | MoAI (release/v2.14.0) | Implementation complete, merged as commit `b69cd1b2d`. AC coverage 18/18. Characterization tests confirm B1/B2/P1 bugs reproduce pre-fix and pass post-fix. Coverage: mx 87.5%, complexity 86.2%. TRUST 5: 5/5 (post review fix 52a9bf81f). |
| 0.1.0 | 2026-04-24 | Wave v2.14 SPEC Writer | Initial draft from `docs/design/v2.14.0-release-plan.md` §2.2 + D4/D5/D6 decisions. Owns IMP-V3U-001 (method receiver blindspot), IMP-V3U-005 (substring fan-in false positives), IMP-V3U-007 (unbounded goroutine fan-out). Extends scope to Go-native cross-platform file walk (D4), per-language cyclomatic complexity via tree-sitter (D5), and `@MX:REASON` pairing enforcement in the MX layer (D6). All changes `breaking: false` per release plan §1.2 semver gate. |

---

## 1. Goal (목적)

Close four Tier-1 correctness defects in the MX validator (`internal/hook/mx/validator.go`) and extend the validator to cover three scope gaps that together make it non-functional on Windows and on 15 of the 16 moai-supported languages. The correctness defects — method receiver blindspot (IMP-V3U-001), substring fan-in false positives (IMP-V3U-005), unbounded goroutine fan-out (IMP-V3U-007), and the grep-subprocess platform dependency — silently drop signal on the majority of idiomatic Go code and prevent MX validation from running at all on Windows CI. The scope extensions — Go-native file walker (D4), tree-sitter per-language cyclomatic complexity (D5), and `@MX:REASON` pairing enforcement (D6) — align the validator with the moai `@MX` protocol (`.claude/rules/moai/workflow/mx-tag-protocol.md`) and with the 16-language neutrality principle in CLAUDE.local.md §15.

The SPEC is `breaking: false` per release plan §1.2: no user-visible API change, no config-key change, no wire-format change. Detection improvements that surface previously-invisible violations (method receivers, paired `@MX:REASON`) are documented in CHANGELOG §Detection Improvements per release plan §6.2, mitigated by an additive `transition_mode` grace flag in `mx.yaml`.

## 2. Scope (범위)

### 2.1 In Scope

- **Method receiver detection** (IMP-V3U-001). Add `exportedMethodRe = regexp.MustCompile(`^func\s+\(\w+\s+\*?\w+\)\s+([A-Z]\w+)`)` at package init and extend `extractFunctions` to try both `exportedFuncRe` and the new method regex. Methods flow through the same P1/P2/P3/P4 emission pipeline as top-level functions.
- **Word-boundary fan-in counter** (IMP-V3U-005). Replace `strings.Count(data, funcName)` and `grep ... funcName` substring matching with a single shared helper that uses `\b` + `regexp.QuoteMeta(funcName)` + `\b` compiled once per identifier and invoked via `FindAllIndex`.
- **Bounded worker pool in `ValidateFiles`** (IMP-V3U-007). Wrap the per-file worker goroutine with a `make(chan struct{}, runtime.NumCPU()*2)` semaphore; workers acquire before body work and release in `defer`. WaitGroup and channel-close semantics preserved.
- **Go-native project walker** (D4). Replace `exec.Command("grep", ...)` invocations at `validator.go:264` and any transitive subprocess dependents with `filepath.WalkDir` + `bufio.Scanner` + `regexp` scans that run identically on macOS/Linux/Windows.
- **Windows CI coverage**. Add a Windows-runner job to the test matrix that exercises the MX validator against a representative fixture project; validates zero `exec.Command("grep", ...)` calls reachable from any code path in `internal/hook/mx/`.
- **Tree-sitter complexity package** (D5). New package `internal/hook/mx/complexity/` depending on `github.com/smacker/go-tree-sitter`, exposing `Measure(lang, content, funcName, startLine) (Result, error)` with `Result{Cyclomatic, IfBranches, Supported}`.
- **5-language McCabe query seeding** (D5). Seed query files under `internal/hook/mx/complexity/queries/{go,python,typescript,javascript,rust}.scm` with tree-sitter s-expression patterns that match decision nodes (`if_statement`, `for_statement`, `switch_case`, binary `&&` / `||`, catch clauses, ternary expressions).
- **11-language scaffolding** (D5). Register stub entries for `java`, `kotlin`, `csharp`, `ruby`, `php`, `elixir`, `cpp`, `scala`, `r`, `flutter`, `swift`. Stub returns `Result{Supported: false, Cyclomatic: 0}` without importing the respective tree-sitter sub-packages (binary size containment per v2.14 plan §0.5 D5 caveat).
- **`@MX:REASON` pairing enforcement** (D6). `extractFunctions` gains `hasAnchorReason` / `hasWarnReason` flags set when an `@MX:REASON` sub-line is present within ±1 line of an `@MX:ANCHOR` or `@MX:WARN` tag. `analyzeFile` emits a violation using the existing P1 priority (for ANCHOR without REASON) or P2 priority (for WARN without REASON); `MissingTag` field is set to `"@MX:REASON"`.
- **Characterization tests before fixes** (Reproduction-First / Rule 4). New test files that pin the pre-fix behavior for B1/B2/P1 before the regex and semaphore changes land, then flip to validate post-fix behavior. New test files for complexity per-language fixtures and Windows-native walker.
- **Transition mode grace flag** (additive, non-breaking). `mx.yaml` gains `transition_mode: bool` that downgrades newly-surfaced violations (method-receiver P2/P3/P4, paired-REASON P1/P2) from blocking to advisory for one minor cycle. Default `false`.

### 2.2 Out of Scope

- **No changes to `Validator` interface** (`internal/hook/mx/types.go`). The 2-method interface (`ValidateFile`, `ValidateFiles`) is preserved byte-for-byte. A1 A1 (analyzer any dead weight) and A2 (no FanInCounter interface) are deferred to v3.0 (breaking).
- **No changes to `Violation` struct field set**. New detection reasons reuse existing `Priority` enum values and populate existing `MissingTag` / `Reason` fields. The JSON schema for hook output remains unchanged.
- **No changes to `ValidationConfig` struct field set**. `transition_mode` flag is added, but existing fields and their YAML binding are untouched.
- **No changes to the `mx.yaml` threshold defaults**. `fan_in_anchor: 3` remains hardcoded in `NewValidator`. A1 M4 (config threshold not wired) is deferred to v2.15.
- **No `@MX:LEGACY`, `@MX:SPEC`, `@MX:TEST`, `auto_tag` coverage** (A1 F5-F9). Deferred to v3.0 strategic MX protocol work.
- **No per-file-limit enforcement** (A1 F3). Deferred to v3.0 — requires cross-file ANCHOR demotion logic and touches the `formatReport` contract.
- **No brace-in-string-literal fix** (A1 B3, D4 gap T3). Deferred to v2.16 types domain because the clean fix requires go/scanner tokenization which changes `lineCount` semantics.
- **No blank-line-gap fix** (A1 B4, D4 gap T4). Deferred to v2.16.
- **No changes to sibling packages.** `internal/astgrep/`, `internal/hook/security/`, `internal/lsp/*` are out of scope; belongs to SPEC-UTIL-002 and SPEC-UTIL-003.
- **No subprocess-hygiene work.** UTIL-001's D4 decision removes the subprocess path entirely; `cmd.WaitDelay` / `Setpgid` / stderr-drain patterns are adopted by SPEC-UTIL-003 for the LSP subsystem only.
- **No documentation site changes** (docs-site/). Release-time CHANGELOG and `docs-site/content/{ko,en,ja,zh}/reference/mx-validator.md` updates are owned by Phase 5 of the release plan, not by this SPEC's implementation phase.
- **No grammar version upgrade cadence policy**. Pinning of `github.com/smacker/go-tree-sitter` is an implementation decision; ongoing upgrade policy documentation deferred to v2.15.

## 3. Environment (환경)

Current moai-adk-go state (branch `release/v2.14.0`, 2026-04-24):

- `internal/hook/mx/validator.go` — 480 LOC. `exportedFuncRe` at line 19; `countFanIn` at lines 254-293 with `exec.Command("grep", ...)` at line 264 and `strings.Count(string(data), funcName)` at line 284; `ValidateFiles` at lines 295-380 launches unbounded goroutines via `for path := range filePaths { go func(fp) { ... } }` at line 317.
- `internal/hook/mx/types.go` — 170 LOC. `Validator` interface unchanged; `FileReport`, `ValidationReport`, `Violation`, `Priority` (P1-P4) all stable.
- `internal/hook/mx/config.go` — 156 LOC. `ValidationConfig` with `Thresholds`, `Limits`, `Exclude` sections; parsed from `.moai/config/sections/mx.yaml`.
- `internal/hook/mx/validator_test.go` — 894 LOC, 92.6% coverage per A1 audit §D4, race-clean per `go test -race`.

Reference documents:

- `.claude/rules/moai/workflow/mx-tag-protocol.md` — authoritative @MX protocol; mandates `@MX:REASON` for WARN and ANCHOR (§"Mandatory Fields").
- `docs/design/v2.14.0-release-plan.md` §2.2 (UTIL-001 scope) and §0.5 (D4/D5/D6 binding decisions).
- `.moai/design/utility-review/a1-mx-audit.md` (B1/B2/P1/F1/F2/F4 defects with line citations).
- `.moai/design/utility-review/a4-external-best-practices.md` §T3.4 §T2 §T5 (tree-sitter validation; subprocess hygiene baseline).

Affected files (implementation scope):

- `internal/hook/mx/validator.go` — method regex added, countFanIn rewritten, ValidateFiles gains semaphore, extractFunctions gains REASON pairing fields.
- `internal/hook/mx/complexity/` (new package) — `complexity.go`, per-language query `.scm` files, unit tests.
- `internal/hook/mx/validator_test.go` — extended with characterization + post-fix tests.
- `internal/hook/mx/validator_method_receiver_test.go` (new).
- `internal/hook/mx/validator_fanin_word_boundary_test.go` (new).
- `internal/hook/mx/validator_semaphore_test.go` (new).
- `internal/hook/mx/validator_windows_native_test.go` (new).
- `internal/hook/mx/validator_reason_pairing_test.go` (new).
- `internal/hook/mx/complexity/complexity_test.go` (new).
- `internal/hook/mx/config.go` — `transition_mode bool` field added (additive).
- `go.mod` / `go.sum` — `github.com/smacker/go-tree-sitter` plus 5 grammar sub-packages added.
- `.github/workflows/` — Windows runner job enabling MX validator integration test.

## 4. Assumptions (가정)

- Go 1.22+ is the build target (matches `internal/template/templates/go.mod` minimum; `filepath.WalkDir` available since Go 1.16).
- `github.com/smacker/go-tree-sitter` remains API-stable at its pinned v2.14 version; grammar sub-packages (`golang`, `python`, `typescript`, `javascript`, `rust`) all expose the same `GetLanguage() *sitter.Language` factory pattern.
- The 5 seeded languages cover the most-used moai-supported languages per project marker telemetry; the 11 scaffolded languages can return `Supported: false` without breaking any existing caller because the previous validator had no complexity analysis at all (regression-free for scaffolded languages).
- Tree-sitter's NFA-based query engine does not suffer regex-style catastrophic backtracking; McCabe counting via query-cursor iteration is O(number of decision nodes in the function byte range).
- Word-boundary regex performance is within 1.5× of `strings.Count` for the fan-in hot path (A4 §93 baseline), acceptable for v2.14 without further caching. Future optimization (pre-built inverted index per A1 R3 addendum) deferred to v2.15.
- `runtime.NumCPU()*2` is a reasonable concurrency cap for the file-walk + regex-match workload on both macOS (typical 8-12 cores) and Windows CI runners (typical 2-4 cores).
- The existing `closer goroutine` at `validator.go:348-354` (closes `resultsCh` after `wg.Wait()`) retains its pre-existing `@MX:WARN` documentation and is not modified by UTIL-001.
- Files matching `mx.yaml` `exclude` patterns (`**/vendor/**`, `**/*_generated.go`, `**/mock_*.go`) are skipped by the Go-native walker identically to grep's `--include=*.go` + directory-prune behavior.
- Tests created under `t.TempDir()` per CLAUDE.local.md §6 (test isolation HARD rule) run on all supported CI platforms without file-permission surprises.
- The `transition_mode` grace flag is consumed only at violation-emission time; no schema change propagates to SARIF, hook JSON, or `.moai/state/`.
- The 5-language query files must be representative — per-language fixtures exercise at least one `if`, one `for`/`while`/`case`, one logical operator, and one ternary (where applicable) to validate the query.
- No coexistence conflict with sibling SPEC-UTIL-002 (touches `internal/astgrep/`) or SPEC-UTIL-003 (touches `internal/lsp/core/`); all three SPECs write to disjoint file sets per release plan §5 PARALLEL gate.

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

- REQ-UTIL-001-001: The `extractFunctions` helper SHALL use both `exportedFuncRe` (top-level functions) and a new package-level `exportedMethodRe = regexp.MustCompile(`^func\s+\(\w+\s+\*?\w+\)\s+([A-Z]\w+)`)` to identify exported Go symbols, emitting method receivers into the same `funcInfo` emission pipeline as top-level functions.
- REQ-UTIL-001-002: The fan-in counting routine SHALL apply a word-boundary match (`\b` + `regexp.QuoteMeta(funcName)` + `\b`) against file contents via `regexp.Regexp.FindAllIndex` rather than `strings.Count` substring matching or grep subprocess substring matching.
- REQ-UTIL-001-003: The `ValidateFiles` method SHALL bound concurrent per-file worker goroutines by a semaphore of capacity `runtime.NumCPU() * 2`, acquired before body execution and released via `defer`.
- REQ-UTIL-001-004: The fan-in identifier scan SHALL be implemented by a Go-native routine using `filepath.WalkDir` plus `os.ReadFile` plus compiled-once word-boundary regex, with zero `exec.Command` or `os/exec` invocations reachable from `internal/hook/mx/validator.go` code paths.
- REQ-UTIL-001-005: The MX validator code path SHALL execute identically on macOS, Linux, and Windows without conditional build tags, without platform-specific shell invocations, and without reliance on external binaries (`grep`, `findstr`, or equivalent) being present on `PATH`.
- REQ-UTIL-001-006: A new package `internal/hook/mx/complexity/` SHALL exist with exported symbol `Measure(lang string, content []byte, funcName string, startLine int) (Result, error)` and exported type `Result { Cyclomatic int; IfBranches int; Supported bool }`.
- REQ-UTIL-001-007: The complexity package SHALL depend on `github.com/smacker/go-tree-sitter` and its five grammar sub-packages (`golang`, `python`, `typescript`, `javascript`, `rust`) for seeded language support.
- REQ-UTIL-001-008: The complexity package SHALL ship query files at `internal/hook/mx/complexity/queries/{go,python,typescript,javascript,rust}.scm` defining tree-sitter s-expression patterns that match McCabe decision nodes (conditional branches, loops, switch cases, short-circuit operators, catch clauses, ternary expressions where applicable).
- REQ-UTIL-001-009: The complexity package SHALL register scaffolding entries for `java`, `kotlin`, `csharp`, `ruby`, `php`, `elixir`, `cpp`, `scala`, `r`, `flutter`, and `swift` such that `Measure` invocations return `Result{Supported: false, Cyclomatic: 0, IfBranches: 0}` without panic, without importing their tree-sitter sub-packages, and without emitting a P2 complexity violation for that language.
- REQ-UTIL-001-010: The `extractFunctions` helper SHALL populate two new fields on `funcInfo` — `hasAnchorReason bool` and `hasWarnReason bool` — set to `true` only when a comment line containing `@MX:REASON` is present within ±1 line of the line containing `@MX:ANCHOR` or `@MX:WARN` respectively.
- REQ-UTIL-001-011: Characterization tests SHALL be added that pin the pre-fix behavior for IMP-V3U-001 (method receiver miss), IMP-V3U-005 (substring over-count), and IMP-V3U-007 (unbounded goroutine spawn) before the corresponding regex, counter, and semaphore changes are merged, per Rule 4 Reproduction-First.
- REQ-UTIL-001-012: A Windows-specific CI runner job SHALL execute `go test ./internal/hook/mx/...` against a representative fixture project and assert zero `exec.Command` call depth reachable from `mxValidator.ValidateFile` or `mxValidator.ValidateFiles`.

### 5.2 Event-Driven Requirements

- REQ-UTIL-001-020: WHEN `extractFunctions` processes a line matching `exportedMethodRe` pattern, the function SHALL record the method name (regex group 1) into a `funcInfo` record identical in structure to the top-level function path, including backward comment scan for MX tags and forward brace scan for body length and goroutine patterns.
- REQ-UTIL-001-021: WHEN `ValidateFiles` is invoked with a non-empty `filePaths` slice, the function SHALL initialize a semaphore channel `make(chan struct{}, runtime.NumCPU()*2)` before the `for` loop, send to it before launching each worker goroutine, and receive from it in the worker goroutine's `defer`.
- REQ-UTIL-001-022: WHEN `analyzeFile` finds a `funcInfo` with `hasAnchor == true && hasAnchorReason == false`, the function SHALL emit a `Violation` with `Priority = P1`, `MissingTag = "@MX:REASON"`, `Reason = "@MX:ANCHOR present but @MX:REASON sub-line missing within 1 line"`, and `Blocking = true`.
- REQ-UTIL-001-023: WHEN `analyzeFile` finds a `funcInfo` with `hasWarn == true && hasWarnReason == false`, the function SHALL emit a `Violation` with `Priority = P2`, `MissingTag = "@MX:REASON"`, `Reason = "@MX:WARN present but @MX:REASON sub-line missing within 1 line"`, and `Blocking = true`.
- REQ-UTIL-001-024: WHEN `complexity.Measure` is called with a seeded `lang` value (`"go"`, `"python"`, `"typescript"`, `"javascript"`, `"rust"`) and valid file content containing the named function, the function SHALL return `Result{Supported: true, Cyclomatic: <McCabe count>, IfBranches: <direct if/else-if count>}` computed from the tree-sitter query for that language.
- REQ-UTIL-001-025: WHEN `complexity.Measure` is called with a scaffolded `lang` value (one of the 11 non-seeded languages), the function SHALL return `Result{Supported: false, Cyclomatic: 0, IfBranches: 0}` and a nil error without attempting any tree-sitter parse.
- REQ-UTIL-001-026: WHEN the Go-native file walker encounters a directory name matching `vendor` or a file matching the `mx.yaml` exclude glob patterns, the walker SHALL skip the directory (via `filepath.SkipDir`) or the file (via continue) without issuing a `Read` or regex match.

### 5.3 State-Driven Requirements

- REQ-UTIL-001-030: WHILE the semaphore in `ValidateFiles` is at capacity (`len(sem) == cap(sem)`), newly-scheduled file workers SHALL block on the send operation `sem <- struct{}{}` until a running worker releases a slot via `<-sem`, ensuring observed concurrency never exceeds `runtime.NumCPU()*2`.
- REQ-UTIL-001-031: WHILE a `funcInfo` has `hasAnchor == true && hasAnchorReason == false && transitionMode == true`, the emitted `Violation` SHALL set `Blocking = false` (advisory mode) rather than `true`, allowing the grace period for pre-existing ANCHOR tags that lack REASON sub-lines.
- REQ-UTIL-001-032: WHILE `complexity.Measure` is parsing a file whose content exceeds 1 MiB, the function SHALL return `Result{Supported: false, Cyclomatic: 0, IfBranches: 0}` and a nil error to prevent tree-sitter from consuming unbounded memory on pathological inputs.

### 5.4 Optional Requirements

- REQ-UTIL-001-040: WHERE `complexity.Measure` computes `Cyclomatic >= 15` for a function, the `analyzeFile` caller SHALL emit a new P2 Violation with `MissingTag = "@MX:WARN"` and `Reason = "cyclomatic complexity <N> >= 15"` if `fn.hasWarn == false`.
- REQ-UTIL-001-041: WHERE `complexity.Measure` computes `IfBranches >= 8` for a function, the `analyzeFile` caller SHALL emit a new P2 Violation with `MissingTag = "@MX:WARN"` and `Reason = "if-branches <N> >= 8"` if `fn.hasWarn == false`.
- REQ-UTIL-001-042: WHERE `mx.yaml` sets `transition_mode: true`, the validator SHALL emit all newly-surfaced violations (method-receiver-derived P2/P3/P4, paired-REASON P1/P2) with `Blocking = false`, while keeping pre-existing violation categories at their historical `Blocking` values.

### 5.5 Complex Requirements

- REQ-UTIL-001-050: WHILE the host operating system is Windows AND the fixture project contains at least one exported method receiver with a goroutine pattern, WHEN `ValidateFiles` is invoked against the fixture, THEN the validator SHALL return a `ValidationReport` with at least one `Violation` of `Priority = P2`, at least one `FileReport` with `Fallback = true` (Grep fallback marker preserved for API compatibility), and zero `exec.Command` process spawns observable via a process-tracer shim.

## 6. Acceptance Criteria (수용 기준)

- AC-UTIL-001-01: Given a Go source file containing `func (r *FileReport) Handle() { go func(){}() }` with no `@MX:WARN` tag, When `ValidateFile` is called, Then the returned `FileReport.Violations` contains exactly one entry with `Priority == P2` and `FuncName == "Handle"`. (maps REQ-UTIL-001-001, REQ-UTIL-001-020, REQ-UTIL-001-011)
- AC-UTIL-001-02: Given a project directory containing `NewContext`, `NewValidator`, `RenewToken`, and a single call to `New` inside a test file, When `countFanIn` (or its Go-native successor) is invoked with `funcName = "New"`, Then the returned count equals 1 (only the word-boundary match for `New`), not 4 or greater (substring false positives). (maps REQ-UTIL-001-002, REQ-UTIL-001-004, REQ-UTIL-001-011)
- AC-UTIL-001-03: Given a synthetic fixture project with 1000 `.go` files each containing one exported function that would trigger `countFanIn`, When `ValidateFiles` is invoked, Then the observed maximum concurrent in-flight worker goroutines does not exceed `runtime.NumCPU()*2` as measured by an atomic counter wired into a test-only shim. (maps REQ-UTIL-001-003, REQ-UTIL-001-021, REQ-UTIL-001-030, REQ-UTIL-001-011)
- AC-UTIL-001-04: Given the `internal/hook/mx/` package as built on the `release/v2.14.0` branch, When static analysis (`grep -R "exec.Command" internal/hook/mx/`) is run, Then no match is returned (no `os/exec` usage reachable from the MX validator). (maps REQ-UTIL-001-004, REQ-UTIL-001-005)
- AC-UTIL-001-05: Given the CI matrix includes a Windows runner, When `go test ./internal/hook/mx/...` is executed against a fixture containing one exported method with a goroutine pattern, Then the test suite passes, the validator emits the expected P2 violation, and zero subprocess spawns occur. (maps REQ-UTIL-001-005, REQ-UTIL-001-012, REQ-UTIL-001-050)
- AC-UTIL-001-06: Given the package `internal/hook/mx/complexity/` as built, When its exported surface is enumerated, Then exactly the symbols `Measure` (function) and `Result` (struct with fields `Cyclomatic int`, `IfBranches int`, `Supported bool`) are exported, matching the REQ-UTIL-001-006 contract. (maps REQ-UTIL-001-006)
- AC-UTIL-001-07: Given the complexity package imports, When `go mod graph` is inspected, Then the direct dependency `github.com/smacker/go-tree-sitter` is present along with its five seeded grammar sub-packages (`golang`, `python`, `typescript`, `javascript`, `rust`). (maps REQ-UTIL-001-007)
- AC-UTIL-001-08: Given query files at `internal/hook/mx/complexity/queries/{go,python,typescript,javascript,rust}.scm`, When each query is loaded and compiled against its language grammar, Then all five queries compile without error and produce at least one match against a representative fixture function in the corresponding language. (maps REQ-UTIL-001-008, REQ-UTIL-001-024)
- AC-UTIL-001-09: Given a call `complexity.Measure("java", content, "foo", 1)` with any valid Java content, When executed, Then the returned `Result.Supported == false`, `Result.Cyclomatic == 0`, and the returned error is nil. (maps REQ-UTIL-001-009, REQ-UTIL-001-025)
- AC-UTIL-001-10: Given a Go function of the form `func ComplexDecision(x int) int { if x > 0 { if x > 10 { return 1 } }; for i := 0; i < x; i++ { if i%2 == 0 { continue } }; switch x { case 1, 2: return x; default: return 0 } }`, When `complexity.Measure("go", content, "ComplexDecision", 1)` is called, Then `Result.Cyclomatic >= 6` reflecting the actual McCabe decision count. (maps REQ-UTIL-001-008, REQ-UTIL-001-024)
- AC-UTIL-001-11: Given a Go file with an `@MX:ANCHOR: ...` comment but no adjacent `@MX:REASON:` sub-line, When `ValidateFile` analyzes the file, Then the emitted violations include one entry with `Priority == P1`, `MissingTag == "@MX:REASON"`, and `Reason` matching the prefix `"@MX:ANCHOR present but @MX:REASON sub-line missing"`. (maps REQ-UTIL-001-010, REQ-UTIL-001-022)
- AC-UTIL-001-12: Given a Go file with an `@MX:WARN: ...` comment and an adjacent `@MX:REASON: rationale here` on the line immediately below, When `ValidateFile` analyzes the file, Then no violation with `MissingTag == "@MX:REASON"` is emitted for that function. (maps REQ-UTIL-001-010, REQ-UTIL-001-023)
- AC-UTIL-001-13: Given a Go file containing a high-complexity function (McCabe 20) without an `@MX:WARN` tag, When `ValidateFile` runs with tree-sitter complexity enabled for Go, Then the emitted violations include one entry with `Priority == P2`, `MissingTag == "@MX:WARN"`, and `Reason` matching `"cyclomatic complexity 20 >= 15"`. (maps REQ-UTIL-001-040, REQ-UTIL-001-024)
- AC-UTIL-001-13b: Given a Go file containing a function with 10 top-level `if` / `else if` branches (IfBranches == 10) but low overall cyclomatic complexity and no `@MX:WARN` tag, When `ValidateFile` runs, Then the emitted violations include one entry with `Priority == P2`, `MissingTag == "@MX:WARN"`, and `Reason` matching `"if-branches 10 >= 8"`. (maps REQ-UTIL-001-041, REQ-UTIL-001-024)
- AC-UTIL-001-14: Given characterization tests named `TestValidateFile_MethodReceiverBlindspot_PreFix`, `TestCountFanIn_SubstringFalsePositive_PreFix`, and `TestValidateFiles_UnboundedConcurrency_PreFix` marked with a `// CHARACTERIZATION: pins pre-fix behavior` comment, When the pre-fix state of validator.go is checked out, Then all three tests FAIL (expected failures pinning the defects); when the post-fix state is checked out, all three tests PASS. (maps REQ-UTIL-001-011)
- AC-UTIL-001-15: Given `mx.yaml` sets `transition_mode: true` and the project contains a pre-existing `@MX:ANCHOR` tag without a paired `@MX:REASON`, When `ValidateFile` runs, Then the emitted REASON-pairing violation has `Blocking == false`. (maps REQ-UTIL-001-031, REQ-UTIL-001-042)
- AC-UTIL-001-16: Given `mx.yaml` sets `transition_mode: false` (default) and the same scenario, When `ValidateFile` runs, Then the emitted REASON-pairing violation has `Blocking == true`. (maps REQ-UTIL-001-022, REQ-UTIL-001-023)
- AC-UTIL-001-17: Given a project directory with a `vendor/` subdirectory and a generated `zz_generated.go` file, When the Go-native walker enumerates files for fan-in counting, Then neither the vendor tree nor the generated file is read (verified via a read-tracing shim installed in tests). (maps REQ-UTIL-001-026)
- AC-UTIL-001-18: Given a 2 MiB synthetic source file containing the target function, When `complexity.Measure("go", content, "foo", 1)` is invoked, Then the function returns `Result{Supported: false, Cyclomatic: 0, IfBranches: 0}` with a nil error within 100 milliseconds. (maps REQ-UTIL-001-032)

## 7. Constraints (제약)

- **Go version**. Build target Go 1.22+; `filepath.WalkDir` (1.16+) and `runtime.NumCPU()` (forever) are unconditionally available. No conditional build tags.
- **Dependencies**. Exactly one new top-level module dependency: `github.com/smacker/go-tree-sitter` with five grammar sub-packages (`golang`, `python`, `typescript`, `javascript`, `rust`). No other new dependencies.
- **Binary size**. Grammar embedding adds ~50 MiB to the compiled moai binary per v2.14 release plan §0.5 D5 caveat. The 11 scaffolded languages MUST NOT import their tree-sitter sub-packages (stubs only) to contain v2.14 growth at the ~50 MiB ceiling.
- **Non-breaking guarantee**. No public API change, no `mx.yaml` schema removal or rename, no `hook` JSON wire-format change, no `Violation.Priority` enum addition or reordering. Any fix that would require one is out of scope (deferred to v2.16 types domain or v3.0 breaking changes).
- **Memory bound**. `complexity.Measure` refuses inputs above 1 MiB (REQ-UTIL-001-032) to prevent tree-sitter memory exhaustion on pathological files.
- **Concurrency bound**. Worker-pool cap is `runtime.NumCPU()*2`; not user-configurable in v2.14 (deferred to v2.15 with a `mx.yaml` tuning knob).
- **Test isolation**. All new tests use `t.TempDir()` per CLAUDE.local.md §6 HARD rule; no writes to `~/.claude/` or project working tree permitted in test contexts.
- **Language-neutrality**. Per CLAUDE.local.md §15, `internal/template/templates/**` is out of scope; this SPEC touches only `internal/hook/mx/` and the new `internal/hook/mx/complexity/` package.
- **Platform independence**. Zero `exec.Command` reachable from `internal/hook/mx/**` code paths post-fix (REQ-UTIL-001-004). Zero platform-conditional build tags.
- **Reproduction-first**. Characterization tests for B1, B2, P1 MUST be committed BEFORE the corresponding fixes (per CLAUDE.md Rule 4 and REQ-UTIL-001-011).
- **Protocol fidelity**. `@MX:REASON` pairing detection honors the protocol definition (`.claude/rules/moai/workflow/mx-tag-protocol.md` §"Mandatory Fields"); pairing window is ±1 line matching D6 decision.
- **Grammar version pinning**. `go.mod` pins `github.com/smacker/go-tree-sitter` to an exact version selected at implementation time; upgrade cadence policy deferred to v2.15.

## 8. Risks & Mitigations

- **R1 — Detection improvements surface hidden violations.** Method receiver detection and REASON pairing enforcement will emit violations for code that previously passed validation. Users may perceive this as "v2.14 broke my CI." Mitigation: `transition_mode: true` grace flag (REQ-UTIL-001-031, REQ-UTIL-001-042) downgrades newly-surfaced violations to advisory for one minor cycle; v2.14 CHANGELOG §Detection Improvements calls out the expected transition; characterization tests quantify the delta on the moai-adk-go repository itself.
- **R2 — Binary size growth (~50 MiB).** Embedding 5 tree-sitter grammars grows the `moai` binary significantly. Mitigation: the 11 scaffolded languages are stubbed (no grammar import) to cap v2.14 growth; CHANGELOG §Binary Size prominently documents the change; opt-out path via `mx.yaml` `complexity.enabled: false` (additive config key, non-breaking) available for air-gapped or size-sensitive deployments.
- **R3 — Tree-sitter query correctness per language.** Per v2.14 plan §0.5 D5 caveat, decision-node names differ across language grammars (`if_statement` in Go vs `if_expression` in Rust). A defective query would under-count or over-count complexity silently. Mitigation: per-language fixture tests (AC-UTIL-001-08, -10, -13) validate queries against curated input files; each seeded language must have at least one positive and one negative fixture.
- **R4 — Word-boundary regex performance regression.** Per A4 §93 the regex path is ~1.3× slower than `strings.Count` substring matching. Mitigation: the Go-native walker amortizes cost by compiling the identifier regex once per `countFanIn` invocation and reusing across all scanned files; benchmark `BenchmarkValidateFiles_LargeProject` (per A1 D4 gap T6) captures the regression and pins an upper bound. Further optimization (pre-built inverted index) deferred to v2.15.
- **R5 — Semaphore contention on Windows CI runners.** Windows CI runners typically have 2 cores (runtime.NumCPU()*2 = 4); the bound may be too tight for I/O-heavy workloads. Mitigation: `validator_semaphore_test.go` measures worst-case wall time on the fixture project; if the Windows runner wall time exceeds 3× the macOS runner baseline, the SPEC is amended in v2.14 implementation phase to tune the multiplier (tunable bound is a v2.15 enhancement).
- **R6 — Tree-sitter grammar drift.** Upstream grammars (`tree-sitter-go`, etc.) may publish syntax-node renames between the pinned version and user-installed upstream tools. Mitigation: `go.mod` pins exact grammar versions; CI includes a smoke test that loads all 5 queries against their grammars at build time; grammar upgrades are gated by a dedicated PR that re-runs the full per-language fixture suite.
- **R7 — `exec.Command` reachability in transitive dependencies.** While `internal/hook/mx/validator.go` is rewritten to drop `os/exec`, a transitive import might reintroduce a subprocess call. Mitigation: AC-UTIL-001-04 pins `grep -R "exec.Command" internal/hook/mx/` to zero matches; CI enforces this invariant on every PR.
- **R8 — Closer goroutine interaction with the new semaphore.** The existing closer goroutine at `validator.go:348-354` closes `resultsCh` after `wg.Wait()`. Adding a semaphore changes worker lifetime distribution but preserves the `Wait() → close(ch)` ordering. Mitigation: `validator_semaphore_test.go` verifies no "send on closed channel" panic under 1000-file synthetic load.

## 9. Dependencies

### 9.1 Blocked by

- None. SPEC-UTIL-001 can begin implementation immediately after approval; no prerequisite SPECs, no external tool releases required.

### 9.2 Blocks

- v2.15 MX threshold tuning work (the planned `mx.yaml` `complexity.enabled` and `fan_in_anchor` config wiring) depends on the complexity package and bounded worker pool shipping first.
- v2.16 MX type unification work (consolidating `Diagnostic` types across `hook`, `lsp`, `gopls`) expects the MX validator to be at v2.14 correctness baseline before field-level consolidation begins.

### 9.3 Related

- **SPEC-UTIL-002 — ast-grep Integration Hardening + 5-Language Rule Seeding.** Related via the D6 decision: both SPECs implement `@MX:REASON` pairing enforcement (UTIL-001 in the MX layer, UTIL-002 in the ast-grep suppression layer). File scope is disjoint (`internal/astgrep/` vs `internal/hook/mx/`); SPECs may land in parallel per release plan §5 Phase 3 PARALLEL gate.
- **SPEC-UTIL-003 — LSP Subprocess Hygiene + Diagnostic Type Alias.** Related via release plan §2.1 shared cross-cutting themes (unbounded goroutine fan-out appears in MX validator, ast-grep ScanMultiple, and LSP getOrSpawn). UTIL-001 addresses the MX instance; UTIL-003 addresses the LSP instance via `singleflight`. File scope is disjoint (`internal/lsp/core/` vs `internal/hook/mx/`); SPECs may land in parallel.
- **Release plan `docs/design/v2.14.0-release-plan.md`** — binding parent document; §2.2 defines UTIL-001 scope; §0.5 D4/D5/D6 bind the three decision gates.

## 10. Traceability

- **Method receiver blindspot (IMP-V3U-001, REQ-UTIL-001-001, REQ-UTIL-001-020, AC-UTIL-001-01, AC-UTIL-001-14)**
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:17-18` (executive summary top issue #1)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:34` (B1 fix proposal with `exportedMethodRe`)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:126-127` (F1 feature gap citation)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:161-178` (R1 detailed recommendation)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/SYNTHESIS.md:45` (§1.1 B1 synthesis)
  - `/Users/goos/MoAI/moai-adk-go/internal/hook/mx/validator.go:19` (current defective regex)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:114` (IMP-V3U-001 release scope)

- **Substring fan-in false positives (IMP-V3U-005, REQ-UTIL-001-002, REQ-UTIL-001-004, AC-UTIL-001-02, AC-UTIL-001-17)**
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:18-19` (top issue #2)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:35` (B2 fix proposal with `\b` boundary)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:182-201` (R2 detailed recommendation)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/SYNTHESIS.md:46` (§1.1 B2 synthesis)
  - `/Users/goos/MoAI/moai-adk-go/internal/hook/mx/validator.go:264` (grep subprocess line)
  - `/Users/goos/MoAI/moai-adk-go/internal/hook/mx/validator.go:284` (strings.Count line)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:122` (IMP-V3U-005 release scope)

- **Unbounded goroutine fan-out (IMP-V3U-007, REQ-UTIL-001-003, REQ-UTIL-001-021, REQ-UTIL-001-030, AC-UTIL-001-03)**
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:21` (top issue #4)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:66-67` (D3 P1 performance citation)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:205-228` (R3 detailed recommendation)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/SYNTHESIS.md:55` (§1.2 P1 synthesis + 5000-grep estimate)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/SYNTHESIS.md:27` (cross-cutting theme #1)
  - `/Users/goos/MoAI/moai-adk-go/internal/hook/mx/validator.go:302-303` (existing `@MX:WARN` on unbounded fan-out)
  - `/Users/goos/MoAI/moai-adk-go/internal/hook/mx/validator.go:317-345` (goroutine spawn loop)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:126` (IMP-V3U-007 release scope)

- **Windows Go-native scan (D4, REQ-UTIL-001-004, REQ-UTIL-001-005, REQ-UTIL-001-012, REQ-UTIL-001-026, AC-UTIL-001-04, AC-UTIL-001-05, AC-UTIL-001-17)**
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:354-356` (Open Question #4 — Windows support)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:42-45` (§0.5 D4 decision text)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:133-135` (UTIL-001 Go-native scan scope)
  - `/Users/goos/MoAI/moai-adk-go/CLAUDE.local.md` §15 (template language neutrality HARD rule)

- **Tree-sitter 16-language complexity (D5, REQ-UTIL-001-006 through REQ-UTIL-001-009, REQ-UTIL-001-024, REQ-UTIL-001-025, REQ-UTIL-001-032, REQ-UTIL-001-040, REQ-UTIL-001-041, AC-UTIL-001-06 through AC-UTIL-001-10, AC-UTIL-001-13, AC-UTIL-001-18)**
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:130` (F4 complexity feature gap)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:357-358` (Open Question #5 — P2 complexity check ownership)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a4-external-best-practices.md:61` (26 built-in tree-sitter grammars baseline)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a4-external-best-practices.md:93` (tree-sitter 10-100× regex throughput baseline)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:47-51` (§0.5 D5 decision text + caveat)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:135-136` (UTIL-001 tree-sitter scope)

- **`@MX:REASON` pairing enforcement (D6 extension, REQ-UTIL-001-010, REQ-UTIL-001-022, REQ-UTIL-001-023, REQ-UTIL-001-031, REQ-UTIL-001-042, AC-UTIL-001-11, AC-UTIL-001-12, AC-UTIL-001-15, AC-UTIL-001-16)**
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:22` (top issue #5 — `@MX:REASON` enforcement missing)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:128` (F2 feature gap citation)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:232-265` (R4 detailed recommendation)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:53-56` (§0.5 D6 decision text)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:136-137` (UTIL-001 `@MX:REASON` MX-layer extension)
  - `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/workflow/mx-tag-protocol.md` §"Mandatory Fields" (protocol mandate for `@MX:REASON`)

- **Characterization tests (Rule 4 Reproduction-First, REQ-UTIL-001-011, AC-UTIL-001-14)**
  - `/Users/goos/MoAI/moai-adk-go/CLAUDE.md` §7 Rule 4 (Reproduction-First Bug Fixing)
  - `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md:94-102` (D4 testing gaps T1-T6)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:238` (Release plan §6.1 characterization tests mandate)

- **Semver non-breaking gate (Constraints, lifecycle: spec-anchored)**
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:83-91` (§1.2 semver discipline three-way gate)
  - `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md:130` (§2.2 `breaking: false` + `lifecycle: spec-anchored` declaration)

---

Version: 0.1.0
Classification: release-scoped, non-breaking
Owner: MoAI orchestrator + manager-ddd (Run phase)
Last updated: 2026-04-24
Referenced by: `docs/design/v2.14.0-release-plan.md` §2.2, `CHANGELOG.md` §v2.14.0 (to be written in Phase 5)
