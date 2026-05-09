---
spec: SPEC-V3R3-CI-AUTONOMY-001
wave: 6
version: 0.2.0
status: draft
created_at: 2026-05-09
updated_at: 2026-05-09
author: manager-strategy
---

# Wave 6 Execution Strategy (Phase 1 Output)

> Audit trail. manager-strategy output for Wave 6 of SPEC-V3R3-CI-AUTONOMY-001 — T7 i18n Validator.
> Generated: 2026-05-09. Methodology: TDD (Go AST static analyzer with table-driven test fixtures + 30s budget perf harness).
> Wave Base: `origin/main 8760b89cd` (Wave 5 PR #745+#746/#792 merge baseline; Wave 6 depends on Wave 1 ci-mirror framework per plan.md §1).

---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-09 | manager-strategy | Initial draft. Resolved Wave6-Q1..Q4 inline (§7 Decisions). Architecture: 3-layer (AST parser → cross-file resolver → lockset+diff comparator), magic-comment escape cuts across all layers, ci-mirror integration (W6-T06) is the consumption surface. AC mapping: AC-CIAUT-016 (mockReleaseData block), AC-CIAUT-017 (magic comment exempt), AC-CIAUT-023 (30s budget). |
| 0.2.0 | 2026-05-09 | manager-strategy | Plan-auditor FAIL rework (path A). W6-T05 redesigned to dual-mode: `--all-files` (intra-state oracle) + `--diff <git-rev>` (temporal/baseline oracle, mandatory per plan.md:252). AC-CIAUT-016 verification path now uses `--diff` mode replay against a constructed baseline tree. Architecture extended with Layer 4 "Diff Comparator" (§4.1). New `scripts/i18n-validator/diff.go` source file. New testdata fixtures `pr783_diff/baseline/` + `pr783_diff/head/`. Risk W6-R6 added (git CLI integration overhead; shell-out to `git diff` MVP, `go-git` deferred). Internal "OQ4" renamed to "Wave6-Q4" to disambiguate from spec.md §7 OQ4 namespace. Resolves auditor findings 1-3; minor findings 4-6 addressed in tasks-wave6.md companion. |

---

## 1. Wave 6 Goal & Scope

### 1.1 Goal

Wave 6 ships a Go-only static analyzer at `scripts/i18n-validator/` that prevents the PR #783 regression class — translation PRs that mutate string literals which test files assert on, silently breaking the test suite. The validator scans Go source for string literals participating in testify assertions (`require.Equal`, `assert.Contains`, `assert.Equal`, `s.Equal`, etc.), follows cross-file references back to declaration sites (e.g., `mockReleaseData` map in `internal/cli/release_test.go`), builds a "translation-locked" set, and exits non-zero when a PR diff modifies any locked literal without an accompanying `// i18n:translatable` magic comment. The validator integrates into Wave 1's `make ci-local` pipeline via `scripts/ci-mirror/lib/go.sh` and runs as a required CI check. AC-CIAUT-016, 017, and 023 are not satisfiable without this Wave.

### 1.2 What This Wave Does NOT Do (P2 priority rationale)

- Does NOT scan the other 15 supported languages (Python/TypeScript/Rust/etc.); spec.md §2 Out of Scope explicitly defers that to a follow-up SPEC.
- Does NOT track string literals declared in non-test Go files unless they are referenced from a test (consumer-driven scope).
- Does NOT auto-fix or auto-update test assertions when a translation is detected — fail-closed only, manual reconciliation required.
- Does NOT integrate with IDE/LSP for in-editor warnings — pre-push + CI only.
- Does NOT propagate to `.moai/specs/` markdown files (Korean i18n there is purely human-facing, no Go test depends on it).

P2 priority is appropriate because: (a) Wave 1 pre-push hook already catches `golangci-lint` ST1005 on the typical case, (b) the regression class is narrow (test-asserted literal + translation PR), (c) deferring to a follow-up wave is non-blocking for the 5-PR sweep replay scenarios except #783 specifically.

### 1.3 Why a Standalone Tool (not a `go vet` analyzer)

Plan.md §8 specifies `scripts/i18n-validator/main.go` as a freestanding Go binary. Rationale:

- `go vet` analyzers require `golang.org/x/tools/go/analysis` framework + per-package invocation patterns; overhead exceeds value for a single-purpose check.
- A standalone CLI is composable into both `make ci-local` (Wave 1) and CI workflows (Wave 1 `.github/required-checks.yml`).
- Standalone enables the `--changed-only` mode (W6-T07 future flag) for pre-push integration where full-repo scan exceeds budget.

---

## 2. Requirements Mapping

| REQ ID | Description (verbatim from spec.md §3.7) | Implementing Task |
|--------|------------------------------------------|-------------------|
| REQ-CIAUT-037 (Ubiquitous) | Static analyzer at `scripts/i18n-validator/` shall scan Go source for string literals in test assertions and flag them as translation-locked. | W6-T01 (AST parser), W6-T02 (cross-file resolver), W6-T03 (lockset builder) |
| REQ-CIAUT-038 (Event-Driven) | When a translation PR modifies a translation-locked string, the validator shall exit non-zero and report file:line + dependent test. | W6-T05 (diff comparator + non-zero exit) |
| REQ-CIAUT-039 (Ubiquitous) | Validator shall be integrated into `make ci-local` and run in CI as a required check. | W6-T06 (`scripts/ci-mirror/lib/go.sh` extension) |
| REQ-CIAUT-040 (State-Driven) | While the validator processes a file, it shall report progress and not exceed 30s wall-clock for full repo. | W6-T07 (perf harness + budget enforcement) |
| REQ-CIAUT-041 (Optional) | A `// i18n:translatable` magic comment shall exempt a literal from the lock check. | W6-T04 (escape mechanism — cross-cutting across W6-T01..T05) |

All 5 spec.md §3.7 requirements are mapped. No requirement is unbound. No task is unmapped.

---

## 3. Acceptance Criteria Mapping

| AC ID | Description (verbatim) | Validating Task | Verification |
|-------|------------------------|-----------------|--------------|
| AC-CIAUT-016 | i18n validator blocks PR #783 regression — `mockReleaseData["..."] = "유효한 YAML 문서가 아닙니다"` → `"Not a valid YAML document"` change must non-zero exit with file:line + dependent test name. | W6-T05 (`--diff` mode) + W6-T06 | `--diff <baseline-rev>` mode replay test: testdata fixture `pr783_diff/baseline/` (data declared `"유효한 YAML 문서가 아닙니다"`, test asserts the same) + `pr783_diff/head/` (BOTH data and test translated to `"Not a valid YAML document"`). Test creates a temp git repo, commits baseline, then commits head, then runs `i18n-validator --diff <baseline-rev>` against HEAD. Expected: exit 1 + stderr contains `"file:line changed from 'X' to 'Y' while locked by TestReleaseFlow"`. This is the canonical AC-CIAUT-016 verification — `--all-files` mode alone cannot satisfy AC-CIAUT-016 because both data and assertion are translated together (no intra-state mismatch exists). |
| AC-CIAUT-017 | `// i18n:translatable` magic comment exempts a literal from lock check. | W6-T04 | testdata fixture: `const errMsg = "Failed to load config" // i18n:translatable`; integration test asserts exit code 0 |
| AC-CIAUT-023 | 30s wall-clock budget for full-repo scan (cold cache, first run); progress streaming (no >5s silent gap); 30s exceed → timeout exit + error message. | W6-T07 | perf harness in `scripts/i18n-validator/main_test.go` with `time.Now()`-based deadline; CI runner benchmark against repo `./...` |

All 3 W6-relevant AC entries from acceptance.md are mapped to validating tasks. AC verification uses both unit-level testdata fixtures and full-repo integration replay.

---

## 4. Architecture

### 4.1 Four-Layer Pipeline (revised v0.2.0)

```
Layer 1 — AST Parser (W6-T01)
    │
    │  Input: Go file → AST (go/parser, go/ast)
    │  Output: list of (file, line, literal_text, assertion_kind, source_ref)
    │
    │  Detects: testify call patterns, extracts string literal arguments
    │  Detects: identifier references → forwards to Layer 2
    │
    ▼
Layer 2 — Cross-File Resolver (W6-T02)
    │
    │  Input: identifier references from Layer 1
    │  Output: declaration site (file, line) for each identifier
    │
    │  Resolves: package-level const/var declarations, map literals (e.g., mockReleaseData)
    │  Skips: function-local variables (out of scope; testify assertions on locals are
    │         declared inline and Layer 1 already captured them)
    │
    ▼
Layer 3 — Lockset Builder (W6-T03)
    │
    │  Build {file:line → LockedLiteral} lockset over the corpus (current working tree)
    │  Output: immutable Lockset after Freeze()
    │
    ▼
Layer 4 — Diff Comparator (W6-T05) — DUAL ORACLE
    │
    │  Two oracles, selected by CLI flag (default `--all-files`):
    │
    │  Oracle A: `--all-files` (intra-state)
    │     Compares the test-asserted expected value to the const-referenced literal
    │     text WITHIN the current tree. Catches NEW translation regressions where
    │     data declared "X" but test still asserts "Y" (or vice versa). Cannot
    │     catch the case where BOTH are translated together (PR #783 pattern).
    │
    │  Oracle B: `--diff <git-rev>` (temporal/baseline) — MANDATORY per plan.md:252
    │     1. Run `git diff --unified=0 <rev>` to find changed lines
    │     2. Re-parse the changed files at HEAD via go/parser
    │     3. For each changed string literal at HEAD, look up its (file, line)
    │        in the baseline-built lockset (constructed by checking out <rev>
    │        in a `git worktree` temporary directory, OR by `git show <rev>:<file>`
    │        per changed file — Wave 6 ships `git show` MVP per W6-R6 mitigation)
    │     4. If the baseline literal was locked AND not marked translatable AND
    │        the HEAD literal text differs → emit Violation
    │     Catches PR #783 because the baseline lockset has the Korean strings
    │     locked; the HEAD diff shows them changed; the validator flags it.
    │
    │  Both oracles emit the same Violation struct and exit code (1 = violation).
    │  Output: exit 0 (no violation) or 1 (violation) + violation report on stderr
    │
    ▼
Cross-Cutting — Magic Comment Escape (W6-T04)
    │
    │  Layer 1 attaches `i18n:translatable` annotation to literals when comment is present
    │  Layer 4 skips annotated literals from BOTH oracles
    │
    ▼
Consumption — ci-mirror Integration (W6-T06)
    │
    │  scripts/ci-mirror/lib/go.sh adds a 5th step (after step 4 cross-compile):
    │     step 5/5: i18n-validator
    │  CI invocation uses `--diff origin/main` mode (catches PR-time regressions)
    │  Local `make ci-local` invocation uses `--all-files` mode (no baseline ref)
    │  Exit codes propagate (0 OK, 2 lint/test failure now extends to validator failure)
    │
    ▼
QA — Budget Validation (W6-T07)
    │
    │  Perf harness asserts <30s on the dev project's Go corpus (BOTH modes)
    │  Progress streaming via stderr (`[i18n-validator] scanning <file>`)
```

### 4.2 Package Structure

```
scripts/
└── i18n-validator/
    ├── main.go                    # NEW (W6-T01: entry point + AST parse + CLI flags)
    ├── lockset.go                 # NEW (W6-T02 + W6-T03: cross-file resolver + lockset builder)
    ├── diff.go                    # NEW (W6-T05 part 2: --diff mode oracle, git CLI shell-out, baseline lockset construction)
    ├── main_test.go               # NEW (table-driven integration tests; testdata-driven for both modes)
    ├── lockset_test.go            # NEW (cross-file resolver unit tests)
    ├── diff_test.go               # NEW (diff comparator tests; constructs temp git repos via t.TempDir() + git init)
    └── testdata/                  # NEW (Wave 6 fixtures)
        ├── pr783_diff/            # AC-CIAUT-016 fixture (DUAL-TREE; replaces single-tree pr783_mockreleasedata/)
        │   ├── baseline/          # tree state where data + test assert Korean string
        │   └── head/              # tree state where data + test assert English string
        ├── translatable_comment/  # AC-CIAUT-017 fixture: magic comment exempt
        ├── budget_corpus/         # AC-CIAUT-023 fixture: synthetic large input
        └── normal/                # negative case: no test references → no lock
```

### 4.3 Data Model (Go-internal)

```go
// LockedLiteral records a string literal that a test file depends on.
type LockedLiteral struct {
    File         string // declaration site file (e.g., internal/cli/release_test.go)
    Line         int    // declaration site line
    Text         string // literal content (Go quoted form)
    AssertionRef AssertionRef // where the test that locks it resides
    Translatable bool   // true if `// i18n:translatable` magic comment present
}

// AssertionRef identifies the test call site that established the lock.
type AssertionRef struct {
    File     string // test file path
    Line     int    // assertion call line
    TestName string // enclosing func or method receiver.Method
    Method   string // testify method (Equal, Contains, NotContains, etc.)
}

// Lockset is the union of locked literals across the scan corpus.
type Lockset struct {
    Literals map[string]LockedLiteral // key: "<file>:<line>" canonical
    BuiltAt  time.Time
    Corpus   []string // file paths included in the build
}

// Violation is what the diff comparator emits.
type Violation struct {
    File         string
    Line         int
    OldText      string
    NewText      string
    LockedBy     AssertionRef
    Reason       string // e.g., "translation-locked literal modified"
}
```

### 4.4 Dependencies (External Go Stdlib + zero third-party)

- `go/parser`, `go/ast`, `go/token` — AST parsing
- `os/exec` — `git diff` and `git show` invocation for `--diff <git-rev>` mode (REQUIRED for that mode per plan.md:252; `--all-files` mode does not depend on `os/exec`)
- `flag` — CLI argument handling
- `time` — budget enforcement
- `path/filepath`, `io/fs` — corpus walking

[HARD] No third-party dependencies. CLAUDE.local.md §14 no-hardcoding compliance: scope const, exclusion list const.

---

## 5. TDD Strategy per Sub-task

### W6-T01 — AST Parser (Layer 1)

**RED — Test cases (test file only, no implementation):**

1. `TestParseTestFile_ExtractsTestifyEqualLiteral` — input: minimal `*_test.go` with `assert.Equal(t, "expected", got)`; expected: 1 LockedLiteral captured at `(file, line, "expected", AssertionRef{Method: "Equal"})`.
2. `TestParseTestFile_ExtractsRequireContains` — input: `require.Contains(t, slice, "needle")`; expected: 1 LockedLiteral with method `Contains`.
3. `TestParseTestFile_HandlesSuiteReceiverPattern` — input: `s.Equal("expected", got)` inside testify suite (method receiver `*MySuite`); expected: 1 LockedLiteral with `TestName: "(*MySuite).TestX"`.
4. `TestParseTestFile_IgnoresNonAssertionCalls` — input: `fmt.Println("hello")`; expected: 0 LockedLiterals (negative case).
5. `TestParseTestFile_HandlesIdentifierReference` — input: `assert.Equal(t, expectedConst, got)` where `expectedConst` is package-level; expected: 1 reference forwarded to Layer 2 with identifier name `expectedConst`.

**GREEN — Minimal implementation:**

- `go/parser.ParseFile` per file
- Walk AST via `ast.Inspect`, match `*ast.CallExpr` with selector pattern `<pkg-or-receiver>.<MethodName>`
- Allowed methods set (const): `{"Equal", "Equalf", "EqualValues", "Contains", "NotContains", "ContainsString", "Eq"}` extensible via flag
- Allowed callers: `assert`, `require`, `s` (suite receiver heuristic — any single-letter or `suite` ident)
- Extract `*ast.BasicLit` of `token.STRING` kind; for `*ast.Ident`, forward to Layer 2

**REFACTOR concerns:**

- Shared parser cache: W6-T02 also parses files; provide a `*ast.File` cache keyed by absolute path so both layers share a single parse invocation per file. Implement as a private field on the lockset builder, populated lazily.
- Move method/caller allow-list to a `var allowedMethods = map[string]struct{}{...}` to avoid allocations per call.

### W6-T02 — Cross-File Resolver (Layer 2)

**RED — Test cases:**

1. `TestResolveIdentifier_PackageLevelConst` — input: file declares `const ExpectedMsg = "hello"` and a test references `ExpectedMsg`; expected: declaration site `(file, line)` for `ExpectedMsg`.
2. `TestResolveIdentifier_MapLiteralValue` — input: `var mockReleaseData = map[string]string{"key": "value"}` and test references `mockReleaseData["key"]`; expected: declaration site of the value at the key, line of `"value"`.
3. `TestResolveIdentifier_NotFound` — input: identifier referenced but never declared in scanned corpus; expected: `(ErrIdentifierNotFound, nil)` with warning logged (graceful skip, no fatal).
4. `TestResolveIdentifier_ExternalPackage` — input: identifier from another module (e.g., `errors.ErrFoo` from stdlib); expected: skipped with reason `"external-package"` (out of scope per §1.2).

**GREEN — Minimal implementation:**

- Build symbol table per package: walk all `*ast.GenDecl` with `token.CONST` or `token.VAR`, capture identifier → value position
- Walk `*ast.CompositeLit` for map literals; record `(key, value)` pairs with positions
- Lookup identifier: (a) check current file's package symbols (b) check repo-level package map (one entry per dir)

**REFACTOR concerns:**

- Map literal traversal needs care: nested maps/slices, struct literals. Wave 6 supports flat `map[string]string` only (sufficient for `mockReleaseData` pattern). Document scope limit in `lockset.go` doc comment.
- Defer cross-package import resolution; in-package resolution covers AC-CIAUT-016.

### W6-T03 — Lockset Builder (Layer 3 part 1)

**RED — Test cases:**

1. `TestBuildLockset_EmptyCorpus` — no files; expected: `Lockset{Literals: map{}, Corpus: []}`.
2. `TestBuildLockset_TestRefsConst` — corpus has 1 const + 1 test; expected: 1 LockedLiteral keyed by const declaration site.
3. `TestBuildLockset_DuplicateLiteralsAtDifferentSites` — 2 tests both reference the same const; expected: 1 LockedLiteral entry, but `AssertionRef` records the first encountered (or both if we extend; Wave 6 picks first to keep schema simple).
4. `TestBuildLockset_TestSelfReferenceInline` — assertion uses inline literal `"foo"` (not a const ref); expected: LockedLiteral keyed by the inline literal's own (file, line) — the test file itself is the declaration site.

**GREEN — Minimal implementation:**

- Iterate corpus, parse each file once (using shared cache from W6-T01 refactor)
- Use Layer 1 to extract assertions, use Layer 2 to resolve identifiers
- Insert into `Lockset.Literals` map; collision resolution: keep first seen

**REFACTOR concerns:**

- Lockset is mutable during build, frozen after. Provide `Freeze()` method that returns immutable view.

### W6-T04 — Magic Comment Escape (cross-cutting)

**RED — Test cases:**

1. `TestMagicComment_ExemptsLiteralOnSameLine` — `const Msg = "hello" // i18n:translatable`; lockset entry `Translatable: true`; diff comparator skips it.
2. `TestMagicComment_ExemptsLiteralOnPrecedingLineComment` — comment `// i18n:translatable` immediately above `const Msg = "hello"`; expected: same exemption.
3. `TestMagicComment_DoesNotExemptUnrelatedLine` — comment 5 lines away; expected: NOT exempted.
4. `TestMagicComment_PartialMatchIgnored` — comment text `// i18n:translateable` (typo); expected: NOT exempted (exact match required).

**GREEN — Minimal implementation:**

- During Layer 1 parse, capture `*ast.File.Comments`; for each `*ast.CommentGroup`, find ones touching the literal's `token.Pos` line or line-1
- Match comment text against const `magicCommentToken = "i18n:translatable"` (exact substring after `//` trim)
- Set `LockedLiteral.Translatable = true`

**REFACTOR concerns:**

- Comment-position arithmetic: trailing comment vs leading comment vs floating doc comment. Simplify by accepting "same line OR immediately preceding line" only; rely on `token.FileSet.Position()` for line numbers.

### W6-T05 — Diff Comparator (Layer 4) — DUAL ORACLE (revised v0.2.0)

**Design statement**: Wave 6 ships BOTH `--all-files` (intra-state oracle, default) AND `--diff <git-rev>` (temporal/baseline oracle, mandatory per plan.md:252). The two oracles cover complementary failure classes and together satisfy AC-CIAUT-016. See §4.1 Layer 4 diagram for oracle definitions.

**RED — Test cases (covering both modes):**

`--all-files` mode (intra-state oracle):

1. `TestAllFiles_NoMismatch` — corpus where all locked literals match their testify-asserted expected values; expected: 0 violations, exit 0.
2. `TestAllFiles_LockedLiteralMismatch` — corpus where const says `"Y"` but test asserts `"X"`; expected: 1 violation, exit 1, stderr contains `"file:line"` + `"locked by <test name>"`.
3. `TestAllFiles_TranslatableLiteralMismatch` — same as case 2 but with `// i18n:translatable` on the const; expected: 0 violations (exempt), exit 0.

`--diff` mode (temporal/baseline oracle, drives AC-CIAUT-016):

4. `TestDiff_ExitsNonZeroOnPR783Mockreleasedata` — uses `pr783_diff/baseline/` (Korean) and `pr783_diff/head/` (English) fixtures; constructs temp git repo via `t.TempDir()` + `git init`, commits baseline, then commits head, then runs `i18n-validator --diff <baseline-rev>`. Expected: exit 1, stderr contains `"file:line changed from '유효한 YAML 문서가 아닙니다' to 'Not a valid YAML document' while locked by TestReleaseFlow"`. (AC-CIAUT-016 canonical verification)
5. `TestDiff_PassesNormalI18nChange` — translation PR that modifies a string literal NOT referenced by any test (e.g., a CLI user-facing message in `cmd/moai/error.go`); expected: 0 violations, exit 0.
6. `TestDiff_RespectsTranslatableMarker` — translation PR modifies a literal marked `// i18n:translatable`; expected: 0 violations (exempt), exit 0.
7. `TestDiff_NoChangeBetweenRevs` — `git diff <rev>` is empty; expected: 0 violations, exit 0.

**GREEN — Minimal implementation:**

- `--all-files` (default mode, lives in `main.go`): builds lockset from current tree, iterates entries, compares testify-asserted expected vs const-referenced literal text; emits Violation on mismatch unless `Translatable: true`.
- `--diff <git-rev>` (mode lives in `diff.go`):
  1. Shell out to `git diff --unified=0 <rev>` to get changed hunks
  2. Parse hunk headers + added/removed lines; collect set of changed file paths and approximate line ranges
  3. For each changed file, run `git show <rev>:<file>` to get the baseline content; write to a temp file under `t.TempDir()` (in production, `os.TempDir()`); parse that file via `go/parser` to build a baseline lockset entry for the affected file only
  4. Re-parse the HEAD file via `go/parser`; for each string literal whose line falls inside a changed hunk, look up the corresponding baseline literal at the same logical position (key by `(file, identifier)` for const-referenced literals; key by `(file, line-number-in-baseline)` for inline literals)
  5. If the baseline literal was locked AND not marked translatable AND HEAD text differs → emit Violation with both old and new text
  6. Exit 1 on any violation; exit 0 on clean
- Both modes share the lockset builder (W6-T03) and the magic-comment escape (W6-T04) via the `Lockset` API.

**REFACTOR concerns:**

- Shared baseline lockset cache: `--diff` mode parses each changed baseline file individually via `git show`. If many files change, repeated `git show` invocations may slow the run. MVP: per-file `git show` (acceptable for typical PR sizes < 50 files). Refactor: bulk checkout into temp worktree via `git worktree add /tmp/baseline-XXX <rev>` (fewer subprocess invocations but more setup cost). Decision (resolves W6-R6): MVP shell-out per file in Wave 6; bulk worktree refactor deferred unless perf measurement shows >5s overhead on representative PRs.
- `go-git` library evaluated and rejected: adds binary size + dependency surface; `git` CLI is universally available; shell-out semantics are well understood. Same rationale as Wave 5's `os/exec` choice for git primitives.
- `--diff` mode requires the working tree to be a git repo. If `git rev-parse --git-dir` fails (non-git context), validator emits warning and falls back to `--all-files` mode behavior. Documented in `diff.go` package comment + `--help` output. (Honest concern preserved in tasks-wave6.md §Honest Scope Concerns #1.)

### W6-T06 — ci-mirror Integration (consumption)

**RED — Test cases:**

1. `TestCiMirrorGoSh_InvokesValidator` — shell-level test (or Go-level test that runs `bash scripts/ci-mirror/lib/go.sh` against a fixture project): asserts the validator was invoked.
2. `TestCiMirrorGoSh_PropagatesNonZeroExit` — validator returns 1; ci-mirror exits with non-zero (extending exit code 2 contract).

**GREEN — Minimal implementation:**

- Edit `scripts/ci-mirror/lib/go.sh` to add a 5th step:

```sh
log "step 5/5: i18n-validator"
if [ -f "$REPO_ROOT/scripts/i18n-validator/main.go" ]; then
    go run ./scripts/i18n-validator/... || exit 2
else
    log "i18n-validator not present — skipping (out-of-tree consumer)"
fi
```

- The existing `set -eu` at the top + step counter update from `step N/4` → `step N/5`

**REFACTOR concerns:**

- Skip when validator absent (out-of-tree projects don't have it) — graceful degradation matching the existing `golangci-lint` skip pattern.
- Future: pre-build the binary (`go build -o /tmp/i18n-validator ./scripts/i18n-validator/...`) and reuse across runs to amortize compile time. Defer to W6-T07 if budget exceeded.

### W6-T07 — Budget Validation (QA)

**RED — Test cases:**

1. `TestBudget_FullRepoScanWithin30Sec` — `go test -run TestBudget` measures wall-clock of `validator.RunOnCorpus(repoRoot)`; assert duration < 30s on dev project.
2. `TestBudget_ProgressStreamingNoSilentGap` — capture stderr stream during scan; assert no >5s gap between `[i18n-validator] scanning <file>` lines.
3. `TestBudget_TimeoutExitOnExcess` — synthetic large fixture (synthetic 5000-file corpus) with `-deadline 1s` flag; assert exit with timeout error message containing the canonical phrase from AC-CIAUT-023 ("validator exceeded ... budget").

**GREEN — Minimal implementation:**

- Add `--budget <duration>` flag (default 30s)
- Wrap scan loop in `time.AfterFunc(budget, cancel)` pattern; emit progress every file via stderr
- On deadline, exit with code 4 (budget exceeded) and message `"validator exceeded 30s budget, consider scoping to changed files only"` (matches AC-CIAUT-023 verbatim)

**REFACTOR concerns:**

- Caching: AC-CIAUT-023 mentions "caching enabled mode (재실행, hot cache): ≤ 5초 목표". Wave 6 ships cold-cache only; hot-cache layer is deferred (mentioned in spec as goal, not requirement; the wording is "목표" / aspirational target). Document the deferral in main.go doc comment.

---

## 6. Open Questions (Resolved Inline)

> **Namespace note**: Wave6-Q1..Q4 is local to this Wave 6 strategy and is intentionally renamed from "OQ1..OQ4" in v0.2.0 to avoid collision with `spec.md §7 OQ1..OQ4` (which were resolved earlier in the SPEC-level annotation cycle). The renaming addresses auditor finding 5.

### Wave6-Q1: Should validator scan vendor/?

**Decision**: NO. Vendor directories are external code; modifying their string literals is outside developer control and would generate noise. Implementation: scan walks corpus with `if strings.Contains(path, "/vendor/")` skip + `node_modules/` skip.

**Rationale**: Wave 6 scope is the dev project's own Go code. False positives from vendored testify mocks would dwarf signal.

### Wave6-Q2: testify nested call patterns — `assert.Equal(t, expected, actual)` vs `s.Equal(expected, actual)` for testify suites

**Decision**: BOTH supported in Wave 6.

**Rationale**: AC-CIAUT-016 fixture (`internal/cli/release_test.go`) uses `assert.*` calls. But moai-adk-go has multiple testify suite-based test files (e.g., `internal/lsp/...`). Layer 1 parser detects both forms via heuristic: caller is `assert`/`require` (free function) OR a single-letter receiver / `s` / `suite` (method call on suite struct). The allow-list is configurable via `--callers <list>` flag (default `assert,require,s,suite`).

**Rationale**: Excluding suite-pattern would create a false-negative class for ~40% of moai-adk-go's test corpus. Including all forms with a const allow-list is low cost.

### Wave6-Q3: Constants vs string literals — does validator track `const HelloMsg = "..."` references?

**Decision**: YES — this is exactly what W6-T02 (cross-file resolver) does.

**Rationale**: AC-CIAUT-016's `mockReleaseData` is a package-level map; the test references the map's value via key. Without identifier resolution, Layer 1 would only catch literals inline in the assertion, missing the entire regression class. The cross-file resolver is the value-add of Wave 6 over `golangci-lint` ST1005.

### Wave6-Q4: `--all-files` vs `--diff <rev>` mode for AC-CIAUT-016 (revised v0.2.0)

**Decision (revised)**: Ship BOTH `--all-files` and `--diff <git-rev>` in Wave 6 as a dual-oracle design. `--all-files` is the default (works without a baseline ref); `--diff` is the canonical AC-CIAUT-016 verification path.

**Rationale**:

- The PR #783 regression class (data + assertion translated together) leaves no intra-state mismatch — `--all-files` cannot detect it because the const says `"Y"` AND the test asserts `"Y"`, so the intra-state oracle returns clean. A temporal/baseline oracle is required.
- plan.md §10.3 line 252 explicitly prescribes diff input: "PR diff 입력 시 변경된 string literal 중 lock set과 교차 검사 → 충돌 시 non-zero exit". Manager-strategy lacks authority to drop this requirement.
- The v0.1.0 reasoning that "the test asserts the Korean string but the const is now English" assumed translators would update only one side; in practice, a competent translator updates both, defeating intra-state detection. This was the auditor's finding 1 + 3.
- Both oracles share the lockset builder (W6-T03) and magic-comment escape (W6-T04), so the marginal implementation cost is the diff comparator (`diff.go`) + git CLI shell-out.

**Implementation note**: "translation-locked" wording is preserved. With both oracles:
- `--all-files`: catches the partial-translation regression class (one side updated, other not — typically caught by `go test` already, but this validator catches it pre-test as a fast pre-push gate).
- `--diff`: catches the full-translation regression class (PR #783 — both sides updated, no intra-state mismatch, only detectable by comparing against baseline).

Document in `main.go` package comment + `diff.go` package comment.

---

## 7. Constraints Compliance

Reference: spec.md §4.1 Hard Constraints + CLAUDE.local.md §14 (no hardcoding) + §15 (16-language neutrality) + §2 (Template-First).

| Constraint | Wave 6 Compliance |
|------------|-------------------|
| **Template-First** (CLAUDE.local.md §2) | NOT applicable for `scripts/i18n-validator/` — these are dev-project tooling, not user-facing template files. No `internal/template/templates/` mirror needed. `scripts/ci-mirror/lib/go.sh` is also dev-project (Wave 1 framework). [HARD] verify-via-grep: `ls internal/template/templates/scripts/` should remain empty after Wave 6. |
| **16-language neutrality** (CLAUDE.local.md §15) | Wave 6 is **Go-only by design** (spec.md §2 Out of Scope explicitly defers other languages). The validator binary is invoked from `scripts/ci-mirror/lib/go.sh`, which is the Go-language module — symmetric placement. Other-language modules (`lib/python.sh`, etc.) are NOT modified in Wave 6. Future SPEC may add `scripts/i18n-validator-py/` etc. with the same architecture. |
| **No hardcoding** (CLAUDE.local.md §14) | All paths/budgets/exclusions extracted to const: `defaultBudget = 30 * time.Second`, `magicCommentToken = "i18n:translatable"`, `allowedMethods = map[string]struct{}{...}`, `defaultCallers = []string{"assert", "require", "s", "suite"}`, `corpusExclusions = []string{"vendor/", "node_modules/", ".git/", "testdata/"}`. CLI flags allow override but defaults are SSoT. |
| **AskUserQuestion HARD** (agent-common-protocol.md §User Interaction Boundary) | Wave 6 ships zero user-interaction surface. The validator is invoked non-interactively from ci-mirror; on violation it exits non-zero with stderr report. orchestrator-level escalation (e.g., during `/moai run` Phase 2) is out of Wave 6 scope. |
| **No release/tag automation** (feedback_release_no_autoexec.md) | Wave 6 introduces zero git tag, gh release, or goreleaser invocations. Validator is read-only against the source tree. |
| **Conventional Commits + 🗿 MoAI co-author** | All Wave 6 commits follow this format (see §11). |

---

## 8. File Ownership

### Implementer Scope (write access)

```
scripts/i18n-validator/main.go                                            # NEW (W6-T01, W6-T05 `--all-files` oracle, W6-T07)
scripts/i18n-validator/lockset.go                                         # NEW (W6-T02, W6-T03, W6-T04)
scripts/i18n-validator/diff.go                                            # NEW (W6-T05 `--diff <git-rev>` oracle; git CLI shell-out)
scripts/i18n-validator/testdata/pr783_diff/baseline/                      # NEW (AC-CIAUT-016 fixture: data + test assert Korean)
scripts/i18n-validator/testdata/pr783_diff/head/                          # NEW (AC-CIAUT-016 fixture: BOTH data and test translated to English)
scripts/i18n-validator/testdata/translatable_comment/                     # NEW (AC-CIAUT-017 fixture)
scripts/i18n-validator/testdata/budget_corpus/                            # NEW (AC-CIAUT-023 fixture, may be synthesized at test runtime)
scripts/i18n-validator/testdata/normal/                                   # NEW (negative case fixture)
scripts/ci-mirror/lib/go.sh                                               # EXTEND (W6-T06: add step 5/5)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave6.md                   # this file
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave6.md                      # companion
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/progress.md                         # extend (Phase 1 + 1.5 entries; copy-paste block in tasks-wave6.md §5)
```

### Tester Scope (test files only)

```
scripts/i18n-validator/main_test.go                                       # NEW (table-driven; W6-T01, W6-T05, W6-T07 unit + integration)
scripts/i18n-validator/lockset_test.go                                    # NEW (W6-T02, W6-T03, W6-T04 unit)
scripts/i18n-validator/diff_test.go                                       # NEW (W6-T05 `--diff` mode tests; constructs temp git repo via t.TempDir() + git init + commit baseline + commit head)
```

### Read-Only Scope

```
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/spec.md                             # frozen for Wave 6
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/plan.md                             # frozen for Wave 6
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/acceptance.md                       # frozen for Wave 6
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave[1-5].md               # prior waves audit trail
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave[1-5].md                  # prior waves audit trail
scripts/ci-mirror/run.sh                                                  # Wave 1 entry point (W6 reads to verify dispatch contract)
scripts/ci-mirror/lib/_common.sh                                          # Wave 1 common helpers (read-only reference)
internal/cli/release_test.go                                              # PR #783 regression source (read-only oracle for AC-CIAUT-016 fixture design)
.claude/rules/moai/development/coding-standards.md                        # 16-language + hardcoding rules
CLAUDE.md, CLAUDE.local.md                                                # project rules
```

### Mode

[HARD] Solo mode (single-implementer, `--branch` pattern; lessons #13). Wave 6 scope is too narrow for `--team` parallelism — only 7 source files (3 Go production + 3 Go test + 1 shell extension) plus 5 fixture directories (`pr783_diff/baseline/`, `pr783_diff/head/`, `translatable_comment/`, `budget_corpus/`, `normal/`), all in adjacent paths.

---

## 9. Dependency Graph

```
W6-T01 (AST parser — Layer 1)
    │
    └─→ W6-T02 (cross-file resolver — Layer 2)
            │
            └─→ W6-T03 (lockset builder — Layer 3 part 1)
                    │
                    ├─→ W6-T04 (magic comment escape — cross-cutting)
                    │       │
                    │       └─→ W6-T05 (diff comparator — Layer 4, DUAL ORACLE)
                    │               │   ├─ --all-files oracle (intra-state, main.go)
                    │               │   └─ --diff <git-rev> oracle (temporal, diff.go)
                    │               │
                    │               └─→ W6-T06 (ci-mirror integration)
                    │                       │
                    │                       └─→ W6-T07 (budget validation — QA)
                    │
                    └─→ W6-T05 (alternate path: T05 also depends directly on T03)
```

**Sequential**: T01 → T02 → T03 → {T04, T05} → T06 → T07.

**Parallel opportunity (within Phase 2 implementer)**: T04 and T05 share T03 as parent and can be developed in parallel. In solo mode, sequential implementation per the commit cadence (§11) is simpler.

---

## 10. Commit Cadence

7+ commits, Conventional Commits format. Every commit MUST end with the verbatim trailer (separated by a blank line from the body):

```
🗿 MoAI <email@mo.ai.kr>
```

Cadence (revised v0.2.0 to add W6-T05 dual-mode commits):

1. `test(i18n-validator): W6-T01 RED — AST parser cases (testify Equal/Contains, suite receiver, non-assertion ignore, identifier ref)`
2. `feat(i18n-validator): W6-T01 implement AST parser + testify call detection`
3. `test(i18n-validator): W6-T02 RED — Cross-file resolver cases (package const, map literal, not found, external pkg)`
4. `feat(i18n-validator): W6-T02 implement cross-file resolver + symbol table`
5. `test(i18n-validator): W6-T03 RED — Lockset builder cases (empty/refs/duplicate/inline)`
6. `feat(i18n-validator): W6-T03 implement Lockset.Build + Freeze`
7. `test(i18n-validator): W6-T04 RED — Magic comment escape cases`
8. `feat(i18n-validator): W6-T04 implement i18n:translatable comment exempt`
9. `test(i18n-validator): W6-T05 RED — --all-files oracle cases (no-mismatch, locked-mismatch, translatable-mismatch)`
10. `feat(i18n-validator): W6-T05 implement --all-files oracle (intra-state)`
11. `test(i18n-validator): W6-T05 RED — --diff oracle cases + PR #783 dual-tree fixture (pr783_diff/baseline + head)`
12. `feat(i18n-validator): W6-T05 implement --diff <git-rev> oracle (temporal/baseline) + git show shell-out`
13. `chore(ci-mirror): W6-T06 add step 5/5 i18n-validator to lib/go.sh (--diff origin/main in CI, --all-files locally)`
14. `test(i18n-validator): W6-T07 budget harness + 30s deadline + progress streaming (both modes)`
15. `feat(i18n-validator): W6-T07 wire --budget flag + AC-CIAUT-023 timeout message`

Optional REFACTOR commits between feat steps if shared parser cache (W6-T01/T02) or bulk-checkout optimization (W6-T05 W6-R6) requires extraction.

Each commit body ends with the verbatim trailer:

```
🗿 MoAI <email@mo.ai.kr>
```

---

## 11. Risk Register

| ID | Risk | Mitigation |
|----|------|-----------|
| W6-R1 | False positives — validator flags a literal that is "translation-locked" but actually a stable internal token (e.g., URL path, JSON key). | Magic comment escape (W6-T04) provides explicit opt-out. Document common patterns in `lockset.go` doc comment. |
| W6-R2 | 30s budget overrun on dev project's 1000+ Go files. | (a) shared parser cache (W6-T01 refactor) avoids double-parsing, (b) corpus walking skips testdata/vendor/node_modules, (c) `--budget` flag allows CI override (default 30s, CI runner can bump to 60s if needed without changing AC). Phase 2 perf measurement during W6-T07 confirms compliance; if breach, scope-limit to `_test.go` files + their referenced declaration files. |
| W6-R3 | AST parser fails on cgo files (Go files with `import "C"`). | go/parser tolerates cgo at parse level (treats it as a regular import). If walker errors, log warning + skip file (graceful degradation matching the `git rev-parse HEAD` failure pattern from Wave 5). |
| W6-R4 | Cross-file resolver misses identifier references that span multiple Go modules in a workspace. | Wave 6 scope (§1.2) limits to single-module resolution. `--scope <dir>` flag restricts walk root; multi-module is follow-up. Document the limit in `lockset.go` doc comment + W6-T02 test case 4 `TestResolveIdentifier_ExternalPackage`. |
| W6-R5 | testify suite receiver heuristic (`s` / `suite` / single-letter) over- or under-matches. | Allow-list configurable via `--callers` flag; default covers moai-adk-go's actual patterns. CI runs validator against full repo as part of W6-T06 — if the validator emits 0 violations on green main, the heuristic is sufficient. If false positives appear post-merge, tune via flag without code change. |
| W6-R6 | git CLI integration overhead in `--diff` mode (per-file `git show <rev>:<file>` shell-out incurs subprocess startup cost; large PRs with 100+ changed files may exceed the 30s budget). | (a) MVP ships per-file `git show` (acceptable for typical PR sizes <50 files; resolves diff comparator scope creep). (b) `go-git` library evaluated and rejected to avoid binary-size + dependency surface. (c) Bulk-checkout refactor (`git worktree add /tmp/baseline-XXX <rev>` once + read files from there) is documented as a deferred optimization (W6-T05 REFACTOR concerns) and triggers only if W6-T07 perf measurement shows >5s overhead on representative PRs. (d) `--diff` mode requires git repo context; non-git fallback to `--all-files` mode with stderr warning prevents hard failure in non-git CI runners. |

---

## 12. Out of Wave 6 Scope

- Non-Go i18n validators (Python/TypeScript/Rust/etc.) — separate Wave per language family in a follow-up SPEC.
- IDE/LSP integration — pre-push + CI only.
- Auto-fix suggestions — Wave 6 is fail-closed; manual reconciliation required.
- Bulk-checkout optimization for `--diff` mode (single `git worktree add /tmp/baseline-XXX <rev>` instead of per-file `git show`) — deferred unless W6-T07 perf measurement shows >5s overhead on representative PRs (W6-R6 mitigation tier 2).
- Hot-cache mode for ≤5s target on rerun (mentioned in AC-CIAUT-023 as aspirational `목표`, not required).
- Magic comment auto-injection on legitimate translations — manual responsibility of the translator.
- Cross-module / multi-workspace identifier resolution — single Go module assumed.
- `internal/template/templates/scripts/` mirror — `scripts/i18n-validator/` is dev-project tooling, not user-facing.
- Validator-driven test generation — fail-closed only.
- Integration with `make ci-disable` (Wave 4) — i18n-validator is required, cannot be disabled per W6-T06.

---

Version: 0.2.0
Status: pending plan-auditor re-audit (post-rework path A: dual-mode W6-T05 with --diff oracle)
Last Updated: 2026-05-09
