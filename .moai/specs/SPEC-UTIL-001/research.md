# Research — SPEC-UTIL-001

> Scope: MX Validator Correctness + Tree-sitter 16-Language Complexity + Windows Go-Native Scan + @MX:REASON Enforcement
> Branch: `release/v2.14.0`
> Date: 2026-04-24
> Author: Wave v2.14 SPEC Writer
> Upstream audits: A1 (MX), A4 (external), SYNTHESIS Tier 1

---

## 1. Problem Statement

The MX validator at `internal/hook/mx/validator.go` is the enforcement surface for the @MX tag protocol defined in `.claude/rules/moai/workflow/mx-tag-protocol.md`. Four independent Tier-1 defects and three scope gaps (Windows, complexity, @MX:REASON) together make the validator silently unreliable on idiomatic Go code and entirely non-functional on Windows and on 15 of the 16 moai-supported languages.

### Evidence-backed defects

- **Method receiver blindspot (IMP-V3U-001 / A1 B1 / A1 F1).** `exportedFuncRe = ^func\s+([A-Z]\w+)` at `validator.go:19` cannot match `func (r *T) Method()`. In idiomatic Go the majority of exported API is method receivers; P2 goroutine, P3 length, and P4 no-test violations inside any method are therefore invisible to the validator. A1 §D1 classifies this as Critical and A1 §D6 lists it as the highest-severity feature gap (F1).
- **Substring false positives in fan-in (IMP-V3U-005 / A1 B2 / A1 P2).** `countFanIn` at `validator.go:258-293` uses two substring-matching mechanisms — `exec.CommandContext("grep", ..., funcName, ...)` at line 264 (no `-w`) and `strings.Count(data, funcName)` at line 284 — neither with a word boundary. For short names (`New`, `Get`, `Set`, `Parse`), matches against `NewContext`, `Getter`, `RenewToken`, `SetupLogger`, `ParseError` inflate fan-in past the ANCHOR threshold of 3, generating spurious P1 violations that users cannot suppress without adding inappropriate ANCHOR tags.
- **Unbounded goroutine fan-out (IMP-V3U-007 / A1 P1).** `ValidateFiles` at `validator.go:302-380` launches one goroutine per file path via `for _, path := range filePaths { wg.Add(1); go func(fp string) { ... } }` with no semaphore. For a 500-file project where every file may trigger a `countFanIn` subprocess per missing-ANCHOR exported function, the worst case is ~5,000 concurrent grep processes.
- **Platform dependency (A1 Open Question #4).** The grep invocation at `validator.go:264` makes the MX validator non-functional on Windows (no `grep` binary on `%PATH%` by default). The moai-adk-go project targets cross-platform CLI deployment (per `CLAUDE.local.md` §15 language neutrality); the current implementation violates that contract for the Windows audience.
- **Cyclomatic complexity gap (A1 F4).** The @MX protocol (`.claude/rules/moai/workflow/mx-tag-protocol.md`) mandates `@MX:WARN` for cyclomatic complexity >= 15 and if-branch count >= 8. The validator only detects goroutine patterns for P2; complexity analysis is entirely absent. A1 §D6 classifies F4 as High.
- **Missing @MX:REASON enforcement (A1 F2 / D6 Decision).** The protocol mandates `@MX:REASON` as a mandatory sub-line for every `@MX:WARN` and `@MX:ANCHOR` tag. The validator detects the absence of a WARN/ANCHOR tag but never verifies that an existing WARN/ANCHOR has a paired REASON. v2.14 release plan §0.5 D6 extends this enforcement to the MX layer.

### Release plan binding context

Per `docs/design/v2.14.0-release-plan.md` §2.2 (SPEC-UTIL-001 scope), this SPEC owns IMP-V3U-001, IMP-V3U-005, IMP-V3U-007 and additionally carries the D4 (Windows Go-native), D5 (tree-sitter 16-language complexity), and D6 extension (MX layer @MX:REASON enforcement) decisions. All changes are `breaking: false` by the semver gate in §1.2 — no user-visible API, config key, or wire format changes. Detection improvements that surface previously-invisible violations are documented as Detection Improvements in CHANGELOG per §6.2.

---

## 2. Current Implementation Analysis

### 2.1 validator.go structure (480 LOC)

| Segment | Lines | Responsibility | Defect citations |
|---------|-------|----------------|------------------|
| Regex package-init | 17-22 | `exportedFuncRe`, `goroutineRe` compiled once | B1 (missing method regex) |
| `mxValidator` struct + `NewValidator` | 24-46 | Factory with hardcoded `fanInThreshold: 3` | A1 A4 (threshold not wired to config) |
| `ValidateFile` entry | 48-80 | Single file read + `analyzeFile` dispatch | — |
| `analyzeFile` | 82-161 | P1/P2/P3/P4 violation emission | Depends on `extractFunctions` + `countFanIn` |
| `funcInfo` + `extractFunctions` | 163-252 | Regex-based function discovery, backward tag scan, brace counting | B1, B3, B4, B5, M1 |
| `countFanIn` | 254-293 | `grep` subprocess + per-file `strings.Count` re-read | B2, P2, P3, Platform dependency |
| `ValidateFiles` | 295-380 | Unbounded goroutine per file; closer goroutine | P1 (no semaphore) |
| `testFileFor`, `fileExists`, `formatReport`, `collectByPriority` | 382-480 | Helpers + report rendering | M3 (formatReport length) |

### 2.2 extractFunctions regex gap (B1)

Current line 19:
```
exportedFuncRe = regexp.MustCompile(`^func\s+([A-Z]\w+)`)
```

Matches `func NewFoo(...)` but fails on:
```
func (r *FileReport) P1Count() int   // method receiver, pointer
func (v mxValidator) Close() error   // method receiver, value
```

A1 R1 fix requires a companion regex:
```
exportedMethodRe = regexp.MustCompile(`^func\s+\(\w+\s+\*?\w+\)\s+([A-Z]\w+)`)
```

`extractFunctions` must try both regexes at line 186 and merge results. Group 1 of the successful match is the function/method name.

### 2.3 countFanIn substring defect (B2)

Current lines 264 and 284:
```
cmd := exec.CommandContext(ctx, "grep", "-r", "--include=*.go", "-l", funcName, v.projectRoot)
// ...
count := strings.Count(string(data), funcName)
```

Neither mechanism respects identifier boundaries. For `funcName = "New"` in a codebase containing `NewContext`, `NewValidator`, `RenewToken`, `RenewSession`, the `strings.Count` on one 2000-line file may tally 15 "matches" that are actually zero callers of `New`.

A1 R2 fix:
- Subprocess: add `-w` flag (grep word boundary).
- In-process: `re := regexp.MustCompile(`\b` + regexp.QuoteMeta(funcName) + `\b`)` + `re.FindAllIndex(data, -1)`.

However, the D4 decision in the release plan §0.5 removes the subprocess entirely in favor of a Go-native walker (see §2.5 below). The word-boundary fix collapses to the single regex path post-D4.

### 2.4 ValidateFiles unbounded concurrency (P1)

Current lines 317-346:
```go
for _, path := range filePaths {
    wg.Add(1)
    go func(fp string) {
        defer wg.Done()
        // ... ValidateFile -> analyzeFile -> countFanIn (may spawn grep)
    }(path)
}
```

No semaphore. The existing `@MX:WARN` at line 302 documents the issue but the structural fix is missing. A1 R3 proposes a `runtime.NumCPU()*2` cap via semaphore channel:
```go
sem := make(chan struct{}, runtime.NumCPU()*2)
// acquire before launching goroutine, release via defer
```

### 2.5 Platform dependency (grep subprocess)

`validator.go:264` and transitively `validator.go:317` via the goroutine call chain depend on `exec.Command("grep", ...)`. Windows does not ship grep on `%PATH%` by default. Per A1 Open Question #4 and the D4 release decision, the fix is a Go-native rewrite using `filepath.WalkDir` (Go 1.16+) plus `bufio.Scanner` and `regexp.Regexp`.

Affected invocation sites (to be removed):
- `validator.go:264` — `grep -r --include=*.go -l funcName projectRoot` → replaced by `filepath.WalkDir` + in-memory inverted index.
- `validator.go:284` — `strings.Count(string(data), funcName)` → replaced by word-boundary regex `FindAllIndex`.

The Go-native path additionally enables the bounded worker pool (P1) as a natural byproduct because the walker and counter share the same semaphore.

### 2.6 Complexity analysis gap (F4)

`analyzeFile` at `validator.go:117-128` detects goroutines via `goroutineRe` (`\bgo\s+func\s*\(`) and `strings.Contains(bodyLine, "\tgo ")`. There is no cyclomatic complexity measurement, no if-branch count, no switch-case count. The protocol §"When to Add Tags → @MX:WARN" requires these checks.

The D5 release decision mandates tree-sitter Go bindings for per-language AST queries. A regex-based complexity counter was considered and rejected (see §4.1 below) because of string-literal false positives identical to B3 and because it cannot scale to the 16 moai-supported languages.

### 2.7 @MX:REASON enforcement gap (F2 / D6)

`extractFunctions` at `validator.go:193-211` scans preceding comment lines for `@MX:ANCHOR`, `@MX:WARN`, `@MX:NOTE`, `@MX:TODO` presence (sets `fn.hasAnchor`, `fn.hasWarn`, etc.) but never records whether a paired `@MX:REASON` sub-line exists. The protocol `mx-tag-protocol.md` §"Mandatory Fields" explicitly says `@MX:REASON` is MANDATORY for WARN and ANCHOR.

D6 in the release plan extends the @MX:REASON pairing rule (originally about ast-grep suppressions) to the MX layer: every `@MX:WARN` and `@MX:ANCHOR` tag must have an adjacent `@MX:REASON` sub-line within 1 line above or below. Absent pairing produces a `MISSING_REASON` violation (new diagnostic reason string, no new violation priority tier — reuses P1 for ANCHOR + P2 for WARN per A1 R4 proposal).

---

## 3. Target State

### 3.1 Correctness fixes (IMP-V3U-001, -005, -007)

- Add `exportedMethodRe = regexp.MustCompile(`^func\s+\(\w+\s+\*?\w+\)\s+([A-Z]\w+)`)` at package init. `extractFunctions` tries `exportedFuncRe` first, then `exportedMethodRe`. Group 1 of whichever matches becomes `fn.name`. Methods and top-level functions flow through the same P1/P2/P3/P4 emission pipeline.
- Replace `strings.Count` and `grep -l` with a shared `wordBoundaryCount(data []byte, funcName string) int` helper that compiles `\b` + `regexp.QuoteMeta(funcName)` + `\b` once per call and returns `len(re.FindAllIndex(data, -1))`. Counter is reused by both the `countFanIn` replacement and by unit tests.
- `ValidateFiles` wraps each file-worker goroutine in a semaphore: `sem := make(chan struct{}, runtime.NumCPU()*2)` at function entry; workers acquire with `sem <- struct{}{}` before body work and release with `<-sem` in defer. wg.Add / Wait / close(resultsCh) semantics are preserved.

### 3.2 Go-native file walker (D4)

A new unexported helper `scanProjectForIdentifier(ctx, projectRoot, funcName string) (map[string]int, error)` uses `filepath.WalkDir(projectRoot, walkFn)` to enumerate `.go` files, skipping `vendor/`, hidden dirs, and files matching `mx.yaml` exclude patterns. For each file:
1. `os.ReadFile` returns the content bytes.
2. `wordBoundaryCount(content, funcName)` returns the identifier match count.
3. Results accumulate in a `map[string]int` (file path → match count).

The subprocess invocations at `validator.go:264` and any transitive dependents are removed. The helper is pure Go, works on macOS/Linux/Windows identically, and has no `exec.CommandContext` call in its code path.

### 3.3 Tree-sitter per-language complexity (D5)

A new package `internal/hook/mx/complexity/` exposes:

```
package complexity

type Result struct {
    Cyclomatic  int  // McCabe complexity (decision count + 1)
    IfBranches  int  // direct if / else-if nodes inside the function
    Supported   bool // false when language has no seeded query
}

// Measure returns per-function complexity for the given file content and language.
func Measure(lang string, content []byte, funcName string, startLine int) (Result, error)
```

Implementation depends on `github.com/smacker/go-tree-sitter` (confirmed in A4 T3.4 and A4 rec-3 as the tree-sitter Go binding). Per-language query files under `internal/hook/mx/complexity/queries/{go,python,typescript,javascript,rust}.scm` define s-expression patterns that match decision nodes (`if_statement`, `for_statement`, `case_clause`, `switch_case`, binary `&&`/`||`). McCabe = number of decision-node matches intersecting the function byte range + 1.

Initial v2.14 coverage per D5:
- Seeded languages (5): `go`, `python`, `typescript`, `javascript`, `rust`.
- Scaffolded languages (11): `java`, `kotlin`, `csharp`, `ruby`, `php`, `elixir`, `cpp`, `scala`, `r`, `flutter`, `swift`. For these, `Measure` returns `Result{Supported: false, Cyclomatic: 0}` and the caller skips the P2-complexity check silently (no false positive, no false negative — documented limitation per v2.14 CHANGELOG).

The validator reads the language from `code_comments` (per `mx-tag-protocol.md` §"Language Settings") or from the file extension via an internal mapping. If the resolved language is scaffolded, the legacy goroutine-pattern P2 detection still runs — complexity simply does not add new signal for that language in v2.14.

### 3.4 @MX:REASON pairing enforcement (D6 extension)

`extractFunctions` gains two new fields on `funcInfo`:

```
type funcInfo struct {
    // existing fields ...
    hasAnchorReason bool
    hasWarnReason   bool
}
```

During the backward comment scan (currently `validator.go:194-211`), when a line contains `@MX:ANCHOR` or `@MX:WARN`, the scanner inspects the adjacent lines (±1) for `@MX:REASON`. `hasAnchorReason` / `hasWarnReason` flags are set accordingly.

`analyzeFile` then emits a new diagnostic class when pairing is missing:
- If `fn.hasAnchor && !fn.hasAnchorReason`: emit Violation with `Priority: P1`, `MissingTag: "@MX:REASON"`, `Reason: "@MX:ANCHOR present but @MX:REASON sub-line missing within 1 line"`.
- If `fn.hasWarn && !fn.hasWarnReason`: emit Violation with `Priority: P2`, `MissingTag: "@MX:REASON"`, `Reason: "@MX:WARN present but @MX:REASON sub-line missing within 1 line"`.

Reuse of existing priority tiers (P1/P2) avoids schema changes in `Violation.Priority` — `breaking: false` semver gate passes.

---

## 4. External Research (A4 findings)

### 4.1 Tree-sitter Go bindings landscape (A4 T3.4, A4 rec-3)

Two production-grade Go bindings exist as of 2026-04:

| Library | Maintainer | License | Stars | v2.14 verdict |
|---------|-----------|---------|-------|---------------|
| `github.com/smacker/go-tree-sitter` | smacker | MIT | ~1.3k | SELECTED — cited by A4 T3.4 and used by production tools; bundles grammars as sub-packages; stable API for `Query` / `QueryCursor` |
| `github.com/tree-sitter/go-tree-sitter` | upstream | MIT | ~450 | Considered, declined — newer (2025-11), API still iterating, fewer bundled grammars as of 2026-04 |

The smacker binding was selected based on A4 validation and the existing Go community convergence (A4 §93 cites tree-sitter at "10-100× throughput of equivalent regex scans on 50k+-file repos"). Release plan §2.2 names `github.com/smacker/go-tree-sitter` explicitly.

### 4.2 Grammar availability for 16 moai languages

Per A4 T2 §61 ("26 built-in tree-sitter grammars with the default feature flag `builtin-parser`"), all 16 moai-supported languages have production-ready tree-sitter grammars via the smacker sub-packages or via community-maintained modules:

| Language | Smacker sub-package | Status |
|----------|---------------------|--------|
| go | `smacker/go-tree-sitter/golang` | seeded in v2.14 |
| python | `smacker/go-tree-sitter/python` | seeded in v2.14 |
| typescript | `smacker/go-tree-sitter/typescript` | seeded in v2.14 |
| javascript | `smacker/go-tree-sitter/javascript` | seeded in v2.14 |
| rust | `smacker/go-tree-sitter/rust` | seeded in v2.14 |
| java | `smacker/go-tree-sitter/java` | scaffolded (v2.15+) |
| kotlin | `smacker/go-tree-sitter/kotlin` | scaffolded (v2.15+) |
| csharp | `smacker/go-tree-sitter/csharp` | scaffolded (v2.15+) |
| ruby | `smacker/go-tree-sitter/ruby` | scaffolded (v2.15+) |
| php | `smacker/go-tree-sitter/php` | scaffolded (v2.15+) |
| elixir | `smacker/go-tree-sitter/elixir` | scaffolded (v2.15+) |
| cpp | `smacker/go-tree-sitter/cpp` | scaffolded (v2.15+) |
| scala | `smacker/go-tree-sitter/scala` | scaffolded (v2.15+) |
| r | community (`r-lib/tree-sitter-r`) | scaffolded (v2.15+) |
| flutter (Dart) | `smacker/go-tree-sitter/dart` | scaffolded (v2.15+) |
| swift | `smacker/go-tree-sitter/swift` | scaffolded (v2.15+) |

### 4.3 Memory and binary-size footprint

Per v2.14 release plan §0.5 D5 caveat: "~10 MB per grammar × 16 = ~50-100 MB embedded." Initial v2.14 ships 5 grammars only (~50 MB). Remaining 11 scaffolded queries ship the query file but do not import the tree-sitter sub-package until v2.15+ rotation, keeping v2.14 binary growth bounded at ~50 MB.

### 4.4 Go subprocess idioms 2026 (A4 T5)

Per A4 T5 rec-1 through rec-10, the 2026 baseline for `exec.Command*` includes `cmd.WaitDelay` (graceful timeout), `Setpgid: true` (process-group cleanup), and a stderr-drain goroutine. UTIL-001's D4 decision removes the subprocess path entirely, which sidesteps the T5 hygiene discussion for this SPEC. The stderr-drain pattern is adopted by sibling SPEC-UTIL-003 for the LSP subprocess; UTIL-001 inherits no residual subprocess discipline burden.

### 4.5 Alternatives considered for complexity analysis

- **lizard (Python CLI, 14-lang support).** Mature (since 2012), wide language support, but a Python subprocess spawn per scan violates D4's subprocess-free goal and adds a Python runtime dependency (moai-adk-go ships as a single Go binary).
- **gocyclo (Go-only library).** Zero-subprocess, direct `go/ast` parse, but Go-only. Fails the 16-language neutrality requirement (CLAUDE.local.md §15) and cannot scale to Python, TS, JS, Rust in v2.14.
- **Hand-rolled regex complexity counter.** Rejected: regex detection of `if` / `for` / `switch` tokens has the same string-literal false-positive class as defect B3. Tree-sitter AST queries are semantic, not lexical, and ignore tokens inside strings and comments by construction.

### 4.6 Word-boundary regex performance

Go's `regexp` package does not use backtracking and compiles `\bfunc\b` to a single-pass NFA. Per A4 §93 empirical baselines, the overhead of `re.FindAllIndex` versus `strings.Count` is ~1.3× for short identifiers (measured on 100 MB corpus); correctness gain dominates the cost. Pre-compilation of the regex (lazy `sync.Once` cache keyed on funcName) amortizes to near-`strings.Count` speed when the same identifier is counted across many files.

---

## 5. Risk Assessment

### 5.1 Binary size increase

Adding 5 tree-sitter grammar sub-packages grows the moai binary by ~50 MB. The scaffolded 11 grammars add query files (~2 KB each, static embed) but do not import their sub-packages until v2.15+, bounding v2.14 growth at ~50 MB. Mitigation: release notes prominently state the binary size impact; users opting out of complexity analysis can set `mx.yaml` `complexity.enabled: false` (config key addition, additive — non-breaking).

### 5.2 Test surface expansion

A1 D4 Gap T1-T6 lists 6 missing test categories; v2.14 adds characterization tests for each pre-fix state (Rule 4 Reproduction-First) plus post-fix validation:
- `validator_method_receiver_test.go` (B1 pre/post)
- `validator_fanin_word_boundary_test.go` (B2 pre/post)
- `validator_semaphore_test.go` (P1 bounded concurrency)
- `validator_windows_native_test.go` (D4 Go-native scan)
- `complexity/complexity_test.go` (D5 per-language) — new package tests
- `validator_reason_pairing_test.go` (D6 @MX:REASON pairing)

Existing test suite at `validator_test.go` (894 LOC) is preserved unchanged; characterization tests pin the pre-fix behavior for B1/B2/P1 before the regex / semaphore changes land, then post-fix tests flip to green.

### 5.3 False-positive elimination — fan_in count reduction

B2 fix changes substring-match to word-boundary match. For real-world projects, fan_in counts of `Get`, `Set`, `New`, `Parse` will decrease. A function previously ANCHOR-flagged (fan_in = 7 when counted as substring) may drop to fan_in = 2 post-fix and no longer require `@MX:ANCHOR`. This is the intended correction, but users who preemptively added ANCHOR tags to suppress the spurious warnings will see their tags flagged as `fan_in < threshold` in subsequent validation cycles. v2.14 CHANGELOG §Detection Improvements must call out this transition. Mitigation: `.moai/config/sections/mx.yaml` grace-period flag `transition_mode: true` emits fan_in-decrease violations as warnings for one minor cycle (v2.14 → v2.15). Implementation of the grace flag itself is additive, non-breaking.

### 5.4 Method receiver detection surfaces hidden violations

B1 fix activates P2/P3/P4 detection on method receivers that were previously invisible. Projects using moai-adk-go MX validation will see new violations on their first v2.14 scan. Release plan §8 risk R1 addresses this with `transition_mode: true` grace flag. Characterization tests in UTIL-001 acceptance.md include a "before/after violation count delta" case pinning the expected set of newly-surfaced violations in moai-adk-go itself, providing a concrete upper bound on user-observable surprise.

### 5.5 Tree-sitter grammar version drift

smacker/go-tree-sitter sub-packages embed grammar sources at vendored snapshot points. Upstream grammar bugfixes (e.g., tree-sitter-go adding new syntax node types) require a dependency bump. UTIL-001 pins the smacker version at `go.mod` entry (exact version selected at implementation time, not in SPEC scope). Upgrade cadence documentation is deferred to v2.15's utility hardening follow-up.

### 5.6 Tree-sitter query correctness per language

Per D5 caveat: "tree-sitter node-type taxonomy differs per language (e.g., `if_statement` in Go vs `if_expression` in Rust)." Each of the 5 seeded query files must be validated against representative fixtures. Acceptance criteria AC-UTIL-001-11 through AC-UTIL-001-14 pin per-language complexity expectations against fixture files. Scaffolded languages (11) simply return `Supported: false` and escape the validation requirement in v2.14.

### 5.7 Semaphore contention under high parallelism

`runtime.NumCPU()*2` bound may be too tight for I/O-heavy workloads on SSD storage. Per A1 R3 the cap is tunable via `mx.yaml` in a future minor (out of v2.14 scope). v2.14 ships the fixed bound; regression is detected by `validator_semaphore_test.go` measuring observed-concurrency ceiling under a 1000-file synthetic load.

### 5.8 Goroutine lifecycle — closer goroutine pre-existing risk

`validator.go:348-354` spawns a closer goroutine to close `resultsCh` after `wg.Wait()`. This is pre-existing behavior with an existing `@MX:WARN` at line 348. UTIL-001 does not modify closer goroutine semantics; only the per-file worker goroutines gain the semaphore. The closer goroutine retains its documented risk; sibling SPEC-UTIL-003 addresses related subprocess-hygiene concerns in the LSP subsystem but not here.

---

## 6. References

### 6.1 Audit citations

- `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a1-mx-audit.md`
  - §D1 B1 (method receiver blindspot) — line reference `validator.go:19`
  - §D1 B2 (substring fan-in false positive) — line references `validator.go:264, 284`
  - §D3 P1 (unbounded goroutines) — line reference `validator.go:302-303, 317-345`
  - §D3 P2 (per-function grep) — line reference `validator.go:264`
  - §D6 F1 (method detection missing) — line reference `validator.go:19`
  - §D6 F2 (`@MX:REASON` not enforced) — line reference `validator.go:196-210`
  - §D6 F4 (complexity check missing) — line reference `validator.go:117-128`
  - Open Question #4 (Windows grep dependency)
  - Open Question #5 (P2 complexity regex fragility)
  - Recommendations R1 (method regex), R2 (word boundary), R3 (worker pool), R4 (@MX:REASON)

- `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/SYNTHESIS.md`
  - §1.1 B1/B2 (MX Critical bugs) — IMP-V3U-001, IMP-V3U-005
  - §1.2 P1 (MX unbounded fan-out) — IMP-V3U-007
  - §4.1 (cross-cutting unbounded goroutine spawn)
  - §1.3 F1-F9 (MX protocol gaps)
  - §1.4 (A4 external research validation — tree-sitter, @MX:REASON promotion)
  - Top-5 cross-cutting findings item 1 (unbounded goroutines appear in all three audits)

- `/Users/goos/MoAI/moai-adk-go/.moai/design/utility-review/a4-external-best-practices.md`
  - §T3.4 (LSP 3.17 callHierarchy; tree-sitter fallback confirmation)
  - §T2 §61 (26 built-in tree-sitter grammars, Dart re-enabled 0.42.x)
  - §93 (tree-sitter throughput baseline: 10-100× regex)
  - T5 rec-1 through rec-10 (Go subprocess hygiene 2026 baseline)

### 6.2 Release plan binding

- `/Users/goos/MoAI/moai-adk-go/docs/design/v2.14.0-release-plan.md`
  - §2.2 SPEC-UTIL-001 scope paragraph
  - §0.5 D4 (Windows Go-native approach)
  - §0.5 D5 (tree-sitter 16-language complexity)
  - §0.5 D6 (ast-grep suppression + @MX:REASON extension to MX layer)
  - §1.2 (semver discipline — `breaking: false` gate)
  - §6.3 (code changes scope)
  - §8 R1 (detection improvements risk)

### 6.3 Protocol binding

- `/Users/goos/MoAI/moai-adk-go/.claude/rules/moai/workflow/mx-tag-protocol.md`
  - §"Tag Types" (@MX:NOTE, @MX:WARN, @MX:ANCHOR, @MX:TODO)
  - §"When to Add Tags → @MX:WARN" (complexity >= 15, if-branches >= 8)
  - §"Mandatory Fields" (@MX:REASON mandatory for WARN and ANCHOR)
  - §"Tag Lifecycle Rules" (ANCHOR demotion, WARN persistence)

### 6.4 Current source

- `/Users/goos/MoAI/moai-adk-go/internal/hook/mx/validator.go` (480 LOC, 2026-04-23 state)
- `/Users/goos/MoAI/moai-adk-go/internal/hook/mx/types.go` (170 LOC — unchanged by v2.14)
- `/Users/goos/MoAI/moai-adk-go/internal/hook/mx/config.go` (156 LOC — unchanged by v2.14)
- `/Users/goos/MoAI/moai-adk-go/internal/hook/mx/validator_test.go` (894 LOC — extended, not replaced)

### 6.5 Sibling SPECs (v2.14.0 cohort)

- SPEC-UTIL-002 (ast-grep integration hardening + 5-lang rule seeding) — Related; distinct file scope under `internal/astgrep/` and `internal/hook/security/`
- SPEC-UTIL-003 (LSP subprocess hygiene + Diagnostic type alias) — Related; distinct file scope under `internal/lsp/core/`

No dependency or blocker relationship between SPEC-UTIL-001 and its siblings; all three may land in parallel per release plan §5 Phase 3 PARALLEL gate.
