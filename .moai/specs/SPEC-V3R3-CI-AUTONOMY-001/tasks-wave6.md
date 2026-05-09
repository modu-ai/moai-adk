---
spec: SPEC-V3R3-CI-AUTONOMY-001
wave: 6
version: 0.2.0
status: draft
created_at: 2026-05-09
updated_at: 2026-05-09
author: manager-strategy
---

# Wave 6 Atomic Tasks (Phase 1 Output)

> Companion to strategy-wave6.md. SPEC-V3R3-CI-AUTONOMY-001 Wave 6 — T7 i18n Validator.
> Generated: 2026-05-09. Methodology: TDD (Go AST static analyzer with table-driven testdata fixtures + 30s budget perf harness). Wave Base: `origin/main 8760b89cd`.

---

## Atomic Task Table

| Task ID | Description | Files (provisional) | REQ | AC | Dependencies | File Ownership Scope | Status |
|---------|-------------|---------------------|-----|----|--------------|---------------------|--------|
| W6-T01 | AST parser (Layer 1). `go/parser.ParseFile` per Go file; `ast.Inspect` for `*ast.CallExpr` matching `<caller>.<Method>` where caller in `{assert, require, s, suite}` (configurable via `--callers`) and Method in `allowedMethods` const set (`Equal/Equalf/EqualValues/Contains/NotContains/ContainsString/Eq`). Extract `*ast.BasicLit` of `token.STRING` kind as direct LockedLiteral; `*ast.Ident` as forward-reference for Layer 2. Handles testify suite receiver heuristic (single-letter / `s` / `suite`). Emits LockedLiteral struct `{File, Line, Text, AssertionRef, Translatable}`. | `scripts/i18n-validator/main.go` (new), `scripts/i18n-validator/main_test.go` (new) | REQ-CIAUT-037 | AC-CIAUT-016 (transitive) | none (foundational) | implementer | pending |
| W6-T02 | Cross-file resolver (Layer 2). Build per-package symbol table by walking `*ast.GenDecl` with `token.CONST`/`token.VAR` and `*ast.CompositeLit` for map literals (flat `map[string]string` only — sufficient for `mockReleaseData` pattern; nested maps deferred). Lookup identifier from W6-T01 forwarded refs: (a) current file package symbols (b) repo-level package map. Returns declaration site `(file, line)`. Graceful skip on not-found + external package (`errors.ErrFoo` etc.) with reason logged. Shared `*ast.File` cache with W6-T01 to avoid double-parsing (refactor). | `scripts/i18n-validator/lockset.go` (new), `scripts/i18n-validator/lockset_test.go` (new) | REQ-CIAUT-037 | AC-CIAUT-016 (transitive) | W6-T01 | implementer | pending |
| W6-T03 | Lockset builder (Layer 3 part 1). Iterate corpus with exclusions (`vendor/`, `node_modules/`, `.git/`, `testdata/` — extracted to `corpusExclusions` const per CLAUDE.local.md §14). Use Layer 1 to extract assertions, Layer 2 to resolve identifiers. Build `Lockset{Literals: map[string]LockedLiteral, BuiltAt: time.Time, Corpus: []string}` keyed by `<file>:<line>` canonical. Collision: keep first seen. Provide `Freeze()` method returning immutable view. | `scripts/i18n-validator/lockset.go` (extend), `scripts/i18n-validator/lockset_test.go` (extend) | REQ-CIAUT-037 | AC-CIAUT-016 (transitive) | W6-T01, W6-T02 | implementer | pending |
| W6-T04 | Magic comment escape (cross-cutting Layer 1+3). During W6-T01 parse, capture `*ast.File.Comments`; for each `*ast.CommentGroup`, find ones touching the literal's `token.Pos` line OR line-1 (immediate preceding). Match comment text against `magicCommentToken = "i18n:translatable"` const (exact substring after `//` trim; partial typo `translateable` rejected). Set `LockedLiteral.Translatable = true`. W6-T05 diff comparator skips entries with `Translatable: true`. | `scripts/i18n-validator/main.go` (extend), `scripts/i18n-validator/lockset.go` (extend), `scripts/i18n-validator/main_test.go` (extend), `scripts/i18n-validator/testdata/translatable_comment/` (new fixture) | REQ-CIAUT-041 | AC-CIAUT-017 | W6-T03 | implementer | pending |
| W6-T05 | Diff comparator (Layer 4) + violation reporter — DUAL-MODE (revised v0.2.0 per plan-auditor rework path A). **`--all-files` mode** (default; lives in `main.go`): intra-state oracle. Scans corpus, builds lockset, compares test-asserted expected value against const-referenced literal text within the current tree. Catches partial-translation regressions (one side updated, other not). Cannot satisfy AC-CIAUT-016 alone because PR #783 translates BOTH data and assertion together (no intra-state mismatch). **`--diff <git-rev>` mode** (lives in `diff.go`; mandatory per plan.md:252): temporal/baseline oracle. Shells out to `git diff --unified=0 <rev>` to find changed lines; for each changed string literal at HEAD, retrieves the baseline literal via `git show <rev>:<file>` + `go/parser` re-parse; cross-references against baseline-built lockset; emits Violation when baseline literal was locked AND not marked translatable AND HEAD text differs. Drives AC-CIAUT-016 canonical verification via dual-tree fixture (`pr783_diff/baseline/` + `pr783_diff/head/`). Both modes share lockset (W6-T03) + magic-comment escape (W6-T04); both emit same Violation struct + exit code 1. Non-git context falls back to `--all-files` with stderr warning (W6-R6 mitigation tier d). On violation: stderr report `"<file>:<line> changed from 'X' to 'Y' while locked by <TestName>"` (--diff mode) OR `"<file>:<line>: translation-locked literal modified, locked by <test-file>:<test-line> in <TestName>"` (--all-files mode). | `scripts/i18n-validator/main.go` (extend, --all-files mode), `scripts/i18n-validator/diff.go` (new, --diff mode), `scripts/i18n-validator/main_test.go` (extend), `scripts/i18n-validator/diff_test.go` (new), `scripts/i18n-validator/testdata/pr783_diff/baseline/` (new fixture: data + test assert Korean), `scripts/i18n-validator/testdata/pr783_diff/head/` (new fixture: data + test BOTH translated to English), `scripts/i18n-validator/testdata/normal/` (new negative-case fixture) | REQ-CIAUT-038 | AC-CIAUT-016 | W6-T03, W6-T04 | implementer | pending |
| W6-T06 | ci-mirror integration. Extend `scripts/ci-mirror/lib/go.sh` (existing 834 bytes from Wave 1) by adding step 5/5: `go run ./scripts/i18n-validator/...` invocation. Update step counter `step N/4` → `step N/5` on lines 11, 14, 21, 24. Graceful degradation: skip with log message when `scripts/i18n-validator/main.go` absent (out-of-tree project). Propagate non-zero exit via existing `\|\| exit 2` pattern. NO `internal/template/templates/scripts/` mirror (dev-project tooling per strategy-wave6 §7). | `scripts/ci-mirror/lib/go.sh` (extend) | REQ-CIAUT-039 | AC-CIAUT-016 (consumption) | W6-T05 | implementer | pending |
| W6-T07 | Budget validation + progress streaming (QA). Add `--budget <duration>` flag default `defaultBudget = 30 * time.Second` const. Wrap scan loop in `time.AfterFunc` cancellation pattern; emit `[i18n-validator] scanning <file>` to stderr per file (no >5s silent gap per AC-CIAUT-023). On deadline: exit code 4 + stderr `"validator exceeded 30s budget, consider scoping to changed files only"` (verbatim AC-CIAUT-023 message). Perf harness `TestBudget_FullRepoScanWithin30Sec` measures `validator.RunOnCorpus(repoRoot)` wall-clock; assert <30s on dev project. Synthetic 5000-file fixture for timeout test. | `scripts/i18n-validator/main.go` (extend), `scripts/i18n-validator/main_test.go` (extend), `scripts/i18n-validator/testdata/budget_corpus/` (new synthetic fixture, may be generated at test runtime) | REQ-CIAUT-040 | AC-CIAUT-023 | W6-T06 | implementer | pending |

---

## Dependency Graph

```
W6-T01 (AST parser — Layer 1)
   │
   └─→ W6-T02 (cross-file resolver — Layer 2)
          │
          └─→ W6-T03 (lockset builder — Layer 3 part 1)
                 │
                 ├─→ W6-T04 (magic comment escape — cross-cutting)
                 │      │
                 │      └─→ W6-T05 (diff comparator — Layer 4, DUAL ORACLE)
                 │             │   ├─ --all-files oracle (intra-state, main.go)
                 │             │   └─ --diff <git-rev> oracle (temporal, diff.go)
                 │             │
                 │             └─→ W6-T06 (ci-mirror integration)
                 │                    │
                 │                    └─→ W6-T07 (budget validation — QA)
                 │
                 └─→ W6-T05 (T05 also depends directly on T03; no new task dependency)
```

**Sequential**: T01 → T02 → T03 → {T04, T05} → T06 → T07.
**Parallel opportunity**: T04 and T05 share T03 as parent (can be developed in parallel within Phase 2). In solo mode (lessons #13), sequential execution per commit cadence below is simpler and reduces merge complexity.

---

## File Ownership Assignment (Solo 모드 — sub-agent + --branch 패턴, lessons #13)

### Implementer Scope (write access)

```
scripts/i18n-validator/main.go                                            # W6-T01, W6-T04, W6-T05 (--all-files mode), W6-T07
scripts/i18n-validator/lockset.go                                         # W6-T02, W6-T03, W6-T04
scripts/i18n-validator/diff.go                                            # W6-T05 (--diff <git-rev> mode; git CLI shell-out)
scripts/i18n-validator/testdata/pr783_diff/baseline/                      # W6-T05 fixture (AC-CIAUT-016 baseline tree: data + test assert Korean)
scripts/i18n-validator/testdata/pr783_diff/head/                          # W6-T05 fixture (AC-CIAUT-016 head tree: BOTH data and test translated to English)
scripts/i18n-validator/testdata/translatable_comment/                     # W6-T04 fixture (AC-CIAUT-017)
scripts/i18n-validator/testdata/budget_corpus/                            # W6-T07 fixture (AC-CIAUT-023, runtime-synthesizable)
scripts/i18n-validator/testdata/normal/                                   # W6-T05 negative case fixture
scripts/ci-mirror/lib/go.sh                                               # W6-T06 (extend existing 834 bytes)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave6.md                   # this strategy (read-only after write)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave6.md                      # this tasks file (read-only after write)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/progress.md                         # extend (Phase 1 + 1.5 entries; copy-paste block in §5 below)
```

### Tester Scope (test files only, production code read-only)

```
scripts/i18n-validator/main_test.go                                       # table-driven; W6-T01/T04/T05 (--all-files)/T07 unit + integration
scripts/i18n-validator/lockset_test.go                                    # W6-T02/T03/T04 unit
scripts/i18n-validator/diff_test.go                                       # W6-T05 (--diff mode) tests; constructs temp git repo via t.TempDir() + git init + commit baseline + commit head
```

### Read-Only Scope (cross-Wave consumer / SSoT source)

```
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/spec.md                             # frozen for Wave 6 (§3.7 REQ-037..041 source)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/plan.md                             # frozen for Wave 6 (§8 Wave 6 outline source)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/acceptance.md                       # frozen for Wave 6 (AC-CIAUT-016/017/023 source)
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave[1-5].md               # prior waves audit trail
.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave[1-5].md                  # prior waves audit trail
scripts/ci-mirror/run.sh                                                  # Wave 1 entry point (verify dispatch contract for W6-T06)
scripts/ci-mirror/lib/_common.sh                                          # Wave 1 common helpers (read-only reference)
scripts/ci-mirror/lib/go.sh                                               # 834-byte baseline before W6-T06 extension
internal/cli/release_test.go                                              # PR #783 regression source (oracle for W6-T05 fixture design)
.claude/rules/moai/development/coding-standards.md                        # 16-language + hardcoding rules
.claude/rules/moai/workflow/spec-workflow.md                              # Plan Audit Gate criteria
CLAUDE.md, CLAUDE.local.md                                                # project rules (§14 no-hardcoding, §15 16-language neutrality)
```

### Implicit Read Access (모든 task)

- `.claude/rules/moai/**` (auto-loaded rules — coding-standards, agent-common-protocol, askuser-protocol)
- 모든 Wave 1-5 산출물 (read-only consumer; 특히 Wave 1 ci-mirror framework)

---

## AC Mapping

| Wave 6 Task | Drives AC | Validation |
|-------------|-----------|------------|
| W6-T01 + W6-T02 + W6-T03 | AC-CIAUT-016 (transitive — foundation for diff comparator) | Unit test cases (testify Equal/Contains, suite receiver, identifier ref, package-level const, map literal); `t.TempDir()` isolation |
| W6-T04 | AC-CIAUT-017 (i18n:translatable magic comment exempt) | testdata fixture `translatable_comment/`: `const errMsg = "Failed to load config" // i18n:translatable`; integration test asserts exit 0 |
| W6-T05 | AC-CIAUT-016 (mockReleaseData regression block) | DUAL-MODE verification. (a) `--all-files` mode tests: `TestAllFiles_NoMismatch`, `TestAllFiles_LockedLiteralMismatch`, `TestAllFiles_TranslatableLiteralMismatch`. (b) `--diff <baseline-rev>` mode tests (canonical AC-CIAUT-016 verification): `TestDiff_ExitsNonZeroOnPR783Mockreleasedata` (uses `pr783_diff/baseline/` + `pr783_diff/head/` fixtures, constructs temp git repo via `t.TempDir()` + `git init`, commits baseline, commits head, runs `i18n-validator --diff <baseline-rev>`, asserts exit 1 + stderr contains canonical message), `TestDiff_PassesNormalI18nChange`, `TestDiff_RespectsTranslatableMarker`, `TestDiff_NoChangeBetweenRevs`. |
| W6-T06 | AC-CIAUT-016 (CI consumption surface) | shell-level test (or Go test invoking `bash scripts/ci-mirror/lib/go.sh` against fixture); asserts validator invocation + non-zero exit propagation |
| W6-T07 | AC-CIAUT-023 (30s budget + progress streaming + timeout exit) | `TestBudget_FullRepoScanWithin30Sec` against dev project (real `./...` corpus); `TestBudget_ProgressStreamingNoSilentGap` captures stderr; `TestBudget_TimeoutExitOnExcess` synthetic 5000-file fixture with `-deadline 1s` |

---

## TRUST 5 Targets (Wave 6 SPEC-Level DoD)

| Pillar | Target | Verification |
|--------|--------|--------------|
| **Tested** | `scripts/i18n-validator/` package coverage ≥ 85%; all unit + integration tests PASS; testdata fixtures cover AC-CIAUT-016/017/023; `go test -race` clean; `t.TempDir()` isolation 100% (no host-repo state mutation) | `go test -cover ./scripts/i18n-validator/...`; `go test -race ./scripts/i18n-validator/...` |
| **Readable** | godoc on all exported types (LockedLiteral, AssertionRef, Lockset, Violation) + main.go package comment explaining "translation-locked" semantic; `golangci-lint run ./scripts/i18n-validator/...` 0 issue; CLI `--help` output documents all flags | `golangci-lint run`; `go doc ./scripts/i18n-validator` |
| **Unified** | gofmt + goimports clean; CLI flag naming matches convention (`--budget`, `--callers`, `--scope`); error messages match AC-CIAUT-023 verbatim phrasing where applicable | `gofmt -l ./scripts/i18n-validator/...` empty; `goimports -l` empty |
| **Secured** | Validator is read-only (no file writes outside stderr); no shell injection in error messages; corpus walking respects `corpusExclusions` (no traversal into `.git/` etc.) | code review: no `os.Create`, `os.WriteFile`, `exec.Command` outside test fixtures; corpus walking unit test verifies exclusion |
| **Trackable** | All commits reference SPEC-V3R3-CI-AUTONOMY-001 W6; Conventional Commits + 🗿 MoAI co-author trailer; testdata fixture filenames cite the AC they validate | `git log --grep='SPEC-V3R3-CI-AUTONOMY-001 W6'`; testdata directory naming convention |

---

## Per-Wave DoD Checklist

- [ ] All 7 W6 tasks complete (W6-T01 ~ W6-T07; see table above)
- [ ] `scripts/i18n-validator/main.go` + `lockset.go` + `diff.go` + test files + 5 testdata fixture directories (`pr783_diff/baseline/`, `pr783_diff/head/`, `translatable_comment/`, `budget_corpus/`, `normal/`) all created
- [ ] `scripts/ci-mirror/lib/go.sh` extended with step 5/5 i18n-validator invocation (graceful skip when validator absent)
- [ ] **NO `internal/template/templates/scripts/` mirror** (dev-project tooling per strategy-wave6 §7; verify-via-grep `ls internal/template/templates/scripts/` empty)
- [ ] Test cases per W6-T01..T07 RED phases all written and initially fail
- [ ] All tests PASS post-GREEN with `go test -race ./scripts/i18n-validator/...`
- [ ] `go test -cover ./scripts/i18n-validator/...` ≥ 85%
- [ ] `golangci-lint run ./scripts/i18n-validator/...` 0 issue
- [ ] `make ci-local` passes (no Wave 1 framework regression; W6-T06 step 5/5 succeeds on green main)
- [ ] AC-CIAUT-016 fixture replay: `--diff <baseline-rev>` mode against `pr783_diff/baseline/` + `pr783_diff/head/` produces exit non-zero + stderr report (file:line + "changed from 'X' to 'Y'" + locked-by test name)
- [ ] AC-CIAUT-017 fixture replay: `translatable_comment/` produces exit 0 (magic comment exempt)
- [ ] AC-CIAUT-023 perf harness: full-repo scan completes in <30s on dev project; synthetic timeout fixture exits with code 4 + canonical message
- [ ] All const/path/budget extracted (CLAUDE.local.md §14): `defaultBudget`, `magicCommentToken`, `allowedMethods`, `defaultCallers`, `corpusExclusions`
- [ ] Wave 6 is Go-only by design (CLAUDE.local.md §15 — extension to other languages is follow-up SPEC; verify no `lib/python.sh`/`lib/typescript.sh` etc. modifications)
- [ ] No release/tag automation introduced
- [ ] PR labeled with `type:feature`, `priority:P2`, `area:ci`
- [ ] Conventional Commits + 🗿 MoAI co-author trailer all commits
- [ ] CHANGELOG.md updated with Wave 6 entry on merge

---

## Suggested Commit Cadence

Conventional Commits format. Every commit body MUST end with the verbatim trailer (separated by a blank line):

```
🗿 MoAI <email@mo.ai.kr>
```

15 commits including RED/GREEN pairs (revised v0.2.0: W6-T05 split into `--all-files` + `--diff` mode commits):

1. `test(i18n-validator): W6-T01 RED — AST parser cases (testify Equal/Contains, suite receiver, non-assertion ignore, identifier ref)`
2. `feat(i18n-validator): W6-T01 implement AST parser + testify call detection`
3. `test(i18n-validator): W6-T02 RED — Cross-file resolver cases (package const, map literal, not found, external pkg)`
4. `feat(i18n-validator): W6-T02 implement cross-file resolver + symbol table`
5. `test(i18n-validator): W6-T03 RED — Lockset builder cases (empty/refs/duplicate/inline)`
6. `feat(i18n-validator): W6-T03 implement Lockset.Build + Freeze`
7. `test(i18n-validator): W6-T04 RED — Magic comment escape cases (same-line/preceding/unrelated/typo)`
8. `feat(i18n-validator): W6-T04 implement i18n:translatable comment exempt`
9. `test(i18n-validator): W6-T05 RED — --all-files oracle cases (no-mismatch, locked-mismatch, translatable-mismatch)`
10. `feat(i18n-validator): W6-T05 implement --all-files oracle (intra-state)`
11. `test(i18n-validator): W6-T05 RED — --diff oracle cases + PR #783 dual-tree fixture (pr783_diff/baseline + head)`
12. `feat(i18n-validator): W6-T05 implement --diff <git-rev> oracle (temporal/baseline) + git show shell-out`
13. `chore(ci-mirror): W6-T06 add step 5/5 i18n-validator to lib/go.sh (--diff origin/main in CI, --all-files locally)`
14. `test(i18n-validator): W6-T07 budget harness + 30s deadline + progress streaming (both modes)`
15. `feat(i18n-validator): W6-T07 wire --budget flag + AC-CIAUT-023 timeout message`

Each commit body ends with the verbatim trailer:

```
🗿 MoAI <email@mo.ai.kr>
```

Optional REFACTOR commits between feat steps if shared `*ast.File` parser cache (W6-T01/T02) requires extraction.

---

## §5 Phase 1 + 1.5 Entry for progress.md (Copy-Paste Block)

The orchestrator should append the following block to `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/progress.md` after Phase 1 (manager-strategy) completion + audit:

```markdown
## Wave 6 — Phase 1 (Strategy + Tasks) — v0.2.0 (Rework)

- date: 2026-05-09
- author: manager-strategy
- artifacts:
  - .moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave6.md (status: draft, version 0.2.0)
  - .moai/specs/SPEC-V3R3-CI-AUTONOMY-001/tasks-wave6.md (status: draft, version 0.2.0)
- summary:
  - Wave 6 scope: T7 i18n validator (P2). Standalone Go static analyzer at scripts/i18n-validator/.
  - 7 atomic tasks (W6-T01..W6-T07) covering 4-layer architecture (AST parser → cross-file resolver → lockset → diff comparator with DUAL ORACLE) + magic comment escape (cross-cutting) + ci-mirror integration + budget validation.
  - W6-T05 ships BOTH `--all-files` (intra-state oracle, default) AND `--diff <git-rev>` (temporal/baseline oracle, mandatory per plan.md:252) per plan-auditor rework path A.
  - REQ mapping: REQ-CIAUT-037..041 all bound to tasks.
  - AC mapping: AC-CIAUT-016 (mockReleaseData block via `--diff` mode dual-tree fixture), AC-CIAUT-017 (magic comment exempt), AC-CIAUT-023 (30s budget) all bound to tasks.
  - 4 Open Questions resolved inline (strategy §6 Wave6-Q1..Q4): vendor exclusion, testify suite receiver heuristic, const reference tracking, dual-mode oracle design.
  - Solo mode (--branch pattern, lessons #13). Wave Base: origin/main 8760b89cd.
  - No template-first mirror (dev-project tooling per strategy-wave6 §7).
- next:
  - Phase 1.5 RE-AUDIT: plan-auditor re-review (must-pass: REQ coverage, EARS compliance, AC coverage with `--diff` oracle verification path, file ownership clarity)
  - On PASS: Phase 2 (manager-tdd delegation; sub-agent context error fallback to main-session per Wave 5 §C-6 pattern)

## Wave 6 — Phase 1.5 (Plan Audit) — Iteration 2

- date: <pending>
- author: plan-auditor
- previous_verdict: FAIL (iteration 1; W6-T05 lacked temporal/baseline oracle, contradicted plan.md:252)
- previous_findings_addressed:
  - F1 (W6-T05 cannot satisfy AC-CIAUT-016): RESOLVED via `--diff` oracle ship in v0.2.0
  - F2 (plan.md:252 prescribes diff input): RESOLVED — `--diff <git-rev>` mode now in scope
  - F3 (strategy:289-296 internal contradiction): RESOLVED — §5 W6-T05 rewritten with dual-mode design statement
  - F4 (byte-count miscitation): VERIFIED — actual `wc -c scripts/ci-mirror/lib/go.sh` = 834 bytes (citation correct)
  - F5 (OQ4 namespace collision): RESOLVED — internal section renamed Wave6-Q1..Q4
  - F6 (commit cadence missing trailer text): RESOLVED — verbatim "🗿 MoAI <email@mo.ai.kr>" trailer now present in both commit cadence sections
- verdict: <pending re-audit>
- report: .moai/reports/plan-audit/SPEC-V3R3-CI-AUTONOMY-001-2026-05-09.md
```

---

## Honest Scope Concerns

1. **Dual-oracle design + `--diff` mode git-repo dependency (revised v0.2.0 per plan-auditor finding 1+3)**: Wave 6 now ships BOTH `--all-files` (intra-state oracle, default) AND `--diff <git-rev>` (temporal/baseline oracle, canonical AC-CIAUT-016 verification path). The `--all-files` mode catches partial-translation regressions (one side updated, other not). The `--diff` mode catches full-translation regressions (PR #783 pattern — both sides updated together, no intra-state mismatch detectable without baseline comparison). Honest concern: `--diff` mode requires the working tree to be a git repo. Behavior in non-git context: validator emits stderr warning + falls back to `--all-files` mode (W6-R6 mitigation tier d). Documented in `diff.go` package comment + `--help` output. The `--all-files` mode remains usable in non-git contexts. The wording "translation-locked" is preserved from spec.md §3.7; both oracles operate on the same Lockset abstraction.

2. **testify suite receiver heuristic (resolved §7 OQ2)**: `s.Equal(...)` is detected via single-letter / `s` / `suite` / `suite` ident in the call selector. False positives possible for unrelated types named `s`. Mitigation: `--callers` flag allows project-specific override. moai-adk-go convention uses `s` for testify suites consistently — heuristic suffices for AC-CIAUT-016 fixture replay.

3. **Cross-module identifier resolution out of scope (R-W6-R4)**: Wave 6 resolves identifiers within a single Go module. Multi-module workspaces (e.g., consumers using `replace` directives) get warning + skip. AC-CIAUT-016 is single-module so this does not affect Wave 6 acceptance.

4. **Synthetic budget fixture generation (W6-T07)**: Creating 5000-file fixture statically would balloon the repo. Decision: generate the fixture at test runtime in `t.TempDir()` and clean up automatically. Document the synthesis approach in `TestBudget_TimeoutExitOnExcess` Go-doc comment.

5. **AC-CIAUT-023 hot-cache target (≤5s)**: The acceptance criterion phrases this as "목표" (aspirational target), not a required threshold. Wave 6 ships cold-cache only; hot-cache is deferred. Document the deferral in `main.go` package comment so the next maintainer understands the gap.

6. **W6-T06 graceful degradation when validator absent**: User projects consuming MoAI-ADK templates may not have `scripts/i18n-validator/`. The `if [ -f "$REPO_ROOT/scripts/i18n-validator/main.go" ]` check prevents `go run` failure. This matches the existing `golangci-lint not installed — skipping` pattern in lib/go.sh.

7. **Wave 5 sub-agent 1M context inheritance error precedent (Wave 4/5 lesson)**: Phase 2 may experience the same delegation error. Mitigation: Wave 6 Go code volume estimate ~800-1200 LOC across 3 source files (`main.go`, `lockset.go`, `diff.go`) + 3 test files (`main_test.go`, `lockset_test.go`, `diff_test.go`) + 5 fixture directories (`pr783_diff/baseline/`, `pr783_diff/head/`, `translatable_comment/`, `budget_corpus/`, `normal/`) — manageable for main-session direct implementation as fallback. Phase 2 plan should explicitly state the fallback path.

8. **Performance — `go run` cold-compile overhead in W6-T06**: First invocation per CI run pays 2-5s `go run` compile cost. Defer optimization (`go build -o /tmp/i18n-validator` cache) to a follow-up wave; Wave 6 priority is correctness over CI runtime. Phase 2 perf measurement (W6-T07) will quantify the actual overhead in dev project context.

9. **No Template-First mirror (intentional)**: `scripts/i18n-validator/` is dev-project tooling, not user-facing template. Verify-via-grep: `ls internal/template/templates/scripts/` should remain empty after Wave 6 completion. CLAUDE.local.md §2 explicitly excludes dev-project files from Template-First requirement; Wave 4 set the precedent for `.github/workflows/optional/`.

No hard blockers identified. Wave 6 ready for Phase 1.5 (plan-auditor) upon strategy + tasks approval, then Phase 2 (manager-tdd 위임 또는 main-session 직접 구현 fallback).

---

Version: 0.2.0
Status: pending Phase 1.5 RE-AUDIT (plan-auditor iteration 2; v0.1.0 verdict FAIL, rework path A applied)
Last Updated: 2026-05-09
