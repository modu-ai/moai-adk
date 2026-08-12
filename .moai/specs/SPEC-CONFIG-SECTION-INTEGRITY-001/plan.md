# SPEC-CONFIG-SECTION-INTEGRITY-001 — Implementation Plan

## §A Context

- **Baseline**: HEAD `865cd8aa2` = `origin/main` (ff-synced). Plan-phase executes in the main checkout (spec-workflow.md Step 1 HARD — no worktree). The working tree carries one unrelated modification (`.moai/config/sections/llm.yaml`) which this SPEC MUST NOT touch.
- **Parent lineage**: This SPEC is slice (c) of the 3-way split of `SPEC-CONFIG-TIER-PERSIST-001`, recorded in the parent's §L Split Branches. The split carving preserves the parent's exact slice-(c) semantics; see `spec.md` § Split Lineage for the REQ-CTP → REQ-CSI mapping. The §K reconciliation with `SPEC-V3R6-UPDATE-NOISE-001` is migrated into this child's `spec.md` §K verbatim — REQ-CSI-012's declared supersession of `REQ-UN-007` is unlandable without it.
- **Code under change**:
  - `internal/config/loader.go` — section loader error-handling path (`:121-131` and thirteen siblings, all `slog.Warn(... "using defaults")` today). The loader needs a failed-section record distinct from the absent-section record.
  - `internal/config/manager.go` — `ConfigManager.Save` refusal path (needs to consult the failed-section record before serialising any section).
  - `internal/cli/update/merge/merge.go` — `MergeGitignoreFile` (`:58`), the `UserPatternsMarker` constant (`:40`), and the not-in-template heuristic branch.
  - `internal/cli/update/backup/restore.go` — the 3-way failure path (`:131-134`, currently `recordFallback(..., false, os.Stderr)` then silent fall-through), the 2-way failure warning (`:139-141`), and the absent-base path (`:118-120`, currently neither `recordFallback` nor a warning).
  - `internal/cli/update_noise.go` — `fallbackAdvisoryThreshold = 3` (`:37`), the advisory emission, and the `--verbose` bypass that REQ-CSI-012 makes redundant for the *first* failure.
  - Hook entry points that read config via `hook.ConfigProvider.Get()` — for REQ-CSI-004's operator-visible advisory.
  - CLI command entry points that surface loader errors — for REQ-CSI-003's non-zero exit.
- **Out of scope** (full list in `spec.md` §C): tier resolution (slice a); atomic, mode-preserving writes (slice b, closed); the YAML merge engine internals (`SPEC-UPDATE-YAML-PRESERVE-001`); 3-way base provenance (`SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`); template content.

## §B Known Issues

- **B1 — Absent vs malformed is today indistinguishable in the loader.** All fourteen section loaders in `internal/config/loader.go` swallow a parse failure into `slog.Warn` and `return` early — identical to the absent-file path except for the log line. REQ-CSI-001/002 requires the loader to carry a new failed-section record distinct from the absent-section record. The record's shape (a `failedSections map[string]error` on `Loader` keyed by section name) is the load-bearing design decision; everything downstream (CLI error, hook advisory, `Save` refusal) reads that record.
- **B2 — The CLI/hook split for malformed sections is a caller-context decision, not a loader decision.** The loader records failure uniformly; the caller decides whether to escalate (CLI: REQ-CSI-003, error + non-zero exit) or to advise-and-continue (hook: REQ-CSI-004, advisory + fail-open). The `hook.ConfigProvider.Get()` path MUST NOT abort; the CLI command path MUST abort. Confusing the two would either break the fail-open norm that `.claude/rules/moai/workflow/main-checkout-branch-guard.md` establishes, or silently swallow the malformed file in a CLI context where the operator expects an error.
- **B3 — `Save` refusal must be scoped, not blanket.** A blanket refusal (any section failed → refuse all writes) would block five good sections for one bad one. REQ-CSI-006 requires the refusal to be per-section: the failed section's `saveSection` call returns an error, clean sections' calls proceed normally. The error returned by `Save()` MUST name the failed section so the operator can act.
- **B4 — `MergeGitignoreFile`'s header is already emitted but not parsed.** The `UserPatternsMarker` constant at `merge.go:40` (`# User Custom Patterns (preserved by moai update)`) is emitted on write but skipped as a comment on read (comments are filtered before classification). REQ-CSI-007 requires the header to be parsed as a real section boundary so that only lines *below* it are treated as user-authored. The risk is that a pre-header backup (predating the convention) is now parsed strictly and its patterns discarded — REQ-CSI-011 mitigates with the heuristic fallback when no header is present.
- **B5 — A pre-existing characterization test already pins idempotence.** `TestMergeGitignore_DoubleReMerge_Idempotent` (in `internal/cli/update/merge/gitignore_merge_characterization_test.go`) already asserts double-re-merge idempotence and passes on the baseline tree. REQ-CSI-010 is therefore NOT a fresh behavioural claim — its value is that M2's header-parse and dedupe must not break an invariant that already holds. The new idempotence test must exercise a case the existing one does not (a backup carrying a header AND duplicate user lines, so dedupe and re-merge interact), and the pre-existing test must still pass afterwards (AC-CSI-011).
- **B6 — `TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive` is the most likely to need a same-commit update.** It already asserts that pattern lines under a *previous* run's header are re-collected under the new header. M2's header parse changes the classification path those lines take, so this test is the most likely to need a same-commit update with a comment naming this SPEC — never a deletion.
- **B7 — R9 is High/Med and the mitigation is the declared §K supersession.** The single largest risk in this SPEC is that making the 3-way fallback loud produces noise on every update while `quality.yaml` still fails 3-way AND it reverses an implemented sibling SPEC's whole purpose. The mitigation is that the reversal is a DECLARED supersession (§K, intentional, not accidental); the counter and 3-strike advisory from `REQ-UN-007`/`REQ-UN-008` survive; and `SPEC-UPDATE-YAML-PRESERVE-001` removes the underlying cause. REQ-CSI-014 keeps the ledger governing repetition. Landing REQ-CSI-012 without the §K reconciliation content in `spec.md` is prohibited.
- **B8 — Subagent boundary: n/a.** This SPEC touches `internal/config/`, `internal/cli/update/`, and `internal/cli/update_noise.go`, none of which is subagent-domain code; no `AskUserQuestion` grep applies.
- **B9 — Cross-SPEC policy conflict scan.** `grep -r "Retired\|superseded\|SPEC-V3R6-UPDATE-NOISE" internal/config/ internal/cli/update/` — `SPEC-V3R6-UPDATE-NOISE-001` is `status: implemented` and is the SPEC whose `REQ-UN-007` clause REQ-CSI-012 declares a supersession of; the supersession is recorded in `spec.md` §K (not performed unilaterally on the sibling's artifacts). No retired/superseded SPEC conflicts with the slice-(c) scope.

## §C Pre-flight (run before any code change)

```bash
# 1. Branch + baseline
git branch --show-current                                   # expect: main (or a feature branch — repo-local PR policy applies)
git rev-parse HEAD                                          # expect: 865cd8aa2 (baseline)

# 2. Cross-platform build
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Lint baseline (distinguish NEW vs pre-existing)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Confirm the loader error-handling path at its current line numbers
grep -n 'slog.Warn.*using defaults' internal/config/loader.go | head -20

# 5. Confirm MergeGitignoreFile + the header marker
grep -n 'UserPatternsMarker\|MergeGitignoreFile\|func MergeGitignore' internal/cli/update/merge/merge.go

# 6. Confirm the restore.go 3-way / 2-way / absent-base paths
grep -n 'recordFallback\|3-way\|2-way\|has3Way\|REQ-UN-007' internal/cli/update/backup/restore.go

# 7. Confirm the noise-suppression threshold + the --verbose bypass
grep -n 'fallbackAdvisoryThreshold\|REQ-UN-007\|REQ-UN-008\|verbose' internal/cli/update_noise.go

# 8. Confirm the pre-existing idempotence characterization test
go test -count=1 -v ./internal/cli/update/merge/ 2>&1 | grep -i idempot
```

## §D Constraints

- **PRESERVE** — the fail-open norm for hooks (REQ-CSI-004). A hook that reads a malformed section MUST NOT abort; it continues with compiled defaults AND emits an advisory. This is the norm `.claude/rules/moai/workflow/main-checkout-branch-guard.md` establishes.
- **PRESERVE** — the absent-section silent-default behaviour (REQ-CSI-001). A greenfield project with no `user.yaml` MUST continue to yield compiled defaults silently; M1 must not over-correct greenfield into errors.
- **PRESERVE** — the clean-section persistence path (REQ-CSI-006). One bad file MUST NOT block five good ones; the `Save` refusal is per-section.
- **PRESERVE** — the noise-suppression ledger's governance of repetition (REQ-CSI-014). Making the first occurrence visible MUST NOT disable the 3-strike threshold advisory or the counter reset on success (`REQ-UN-008`/`REQ-UN-009`, which survive the §K supersession).
- **PRESERVE** — the pre-existing characterization tests. `TestMergeGitignore_DoubleReMerge_Idempotent` and `TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive` are the two M2 most directly risks; where a characterization test pins the pre-fix behaviour this SPEC deliberately changes, it is updated in the same commit with a comment naming this SPEC — never deleted.
- **Forbidden commands** — `--no-verify` (a warn-only pre-commit hook is normal); `--amend` after push; force-push to `main`.
- **Required commands** — Conventional Commits format (`feat(SPEC-CONFIG-SECTION-INTEGRITY-001): M{N} <subject>`); commit messages in English per `language.yaml`. The M3 commit body MUST state the `REQ-UN-007` supersession and reference `spec.md` §K (NFR-CSI-006).
- **Repo-local PR policy** — `main` is protected (`enforce_admins: true`); all tiers use Route B (PR), self-merge allowed, CI must pass. The orchestrator handles PR strategy after plan-audit; this plan-phase does NOT commit or push.
- **Scope discipline** — touch only `internal/config/{loader,loader_test,manager,manager_test}.go`, `internal/cli/update/merge/{merge,merge_test,gitignore_merge_characterization_test}.go`, `internal/cli/update/backup/{restore,restore_test}.go`, `internal/cli/update_noise{,_test}.go`, and the hook/CLI entry points that surface malformed-section errors. Do NOT touch `internal/template/templates/**`, the unrelated `.moai/config/sections/llm.yaml` modification in the working tree, or any sibling SPEC directory.

## §E Milestones

Order by decision-reversibility (highest-change-likelihood first), per the SPEC builder plan-ordering rule.

### M0 — probe: baseline current behaviour (Milestone Deliverable, NOT a footnote)

**Deliverable**: a probe record at `.moai/specs/SPEC-CONFIG-SECTION-INTEGRITY-001/baseline-probe.md` capturing, against the baseline tree:

1. Malformed-section handling today: write a truncated `user.yaml` into a `t.TempDir()` fixture, call `Loader.Load`, and record (a) whether the section appears in `LoadedSections()`, (b) whether any failure is visible to the caller, (c) what a follow-up `Save()` writes to the malformed file.
2. `MergeGitignoreFile` current behaviour: reproduce the §A `.gitignore` probe (template v1 contains `DS_Store`, user adds `mysecret.env` and `build/`, template v2 drops `DS_Store`) and record the post-second-merge bytes verbatim.
3. The 3-way/2-way/absent-base fallback paths: locate each in `internal/cli/update/backup/restore.go`, record the line numbers, and confirm which paths emit an advisory today and which do not.

**Why M0 leads** — R6 (Med/Med), R8 (Med/High), and R9 (High/Med) all depend on the probe record for their mitigation evidence. The malformed-section contract and the gitignore header parse both risk breaking workflows that currently work; the probe establishes the baseline those workflows currently produce so a reviewer can tell a real change from a no-op. M0 is a milestone deliverable, not a footnote.

**Work item**: produce the probe record. No code change. No AC beyond the probe's existence (referenced in acceptance.md §A discipline preamble).

### M1 — malformed-section writeback refusal (REQ-CSI-001..006) — ONE revertable commit

The malformed-section contract, shipped alone so it can be reverted without unwinding M2/M3 (NFR-CSI-005).

- **REQ-CSI-001** — the section loaders distinguish **absent** (silent compiled defaults, unchanged) from **malformed** (recorded as failed). The absent path is PRESERVED verbatim; only the malformed path gains a record.
- **REQ-CSI-002** — when a section file is present but fails to parse, the `Loader` records that section in a failed-section map (keyed by section name, value the parse error) and does NOT mark it loaded. Both the failed-section map and the loaded-section map are populated by the loader; the caller reads both.
- **REQ-CSI-003** — where the caller is a CLI command, a malformed section surfaces as a returned error naming the file and the parse failure, and the command exits non-zero. The caller-context split (CLI vs hook) is a caller decision, not a loader decision (B2).
- **REQ-CSI-004** — where the caller is a hook, the hook continues with compiled defaults AND emits an operator-visible advisory on stderr naming the file, in addition to the existing `slog.Warn`. The fail-open norm is PRESERVED (REQ-CSI-004 binds the hook path; `.claude/rules/moai/workflow/main-checkout-branch-guard.md`).
- **REQ-CSI-005** — when any section failed to parse during the most recent `Load`, `ConfigManager.Save` refuses to persist that section and returns an error naming it, rather than writing compiled defaults over the user's file. The on-disk bytes of the malformed file are preserved byte-identical (the load-bearing half of the AC).
- **REQ-CSI-006** — the refusal is scoped to the failed section alone; clean sections' `saveSection` calls proceed normally and write their updated bytes.

**Tests** (TDD RED-GREEN-REFACTOR — `quality.yaml` `development_mode: tdd`):
- `TestLoader_MalformedSectionIsRecordedFailed` (AC-CSI-001)
- `TestLoader_AbsentSectionIsSilent` (AC-CSI-002) — greenfield protection
- `TestCLI_MalformedSectionReturnsError` (AC-CSI-003)
- `TestHook_MalformedSectionAdvisesButContinues` (AC-CSI-004) — both halves required
- `TestConfigManager_SaveRefusesFailedSection` (AC-CSI-005) — byte-identity assertion is load-bearing
- `TestConfigManager_SaveRefusalIsScopedToFailedSection` (AC-CSI-006)

**Why this ordering** — M1 is the data-loss prevention layer; it is the highest-change-likelihood milestone (a new failed-section record in the loader, a new refusal path in `Save`) and ships first so it can be reverted independently of the gitignore (M2) and fallback-visibility (M3) changes (NFR-CSI-005).

### M2 — gitignore merge correctness (REQ-CSI-007..011) — ONE commit

The `.gitignore` merge correctness change, shipped as a single commit that touches `MergeGitignoreFile` and its tests.

- **REQ-CSI-007** — parse the `# User Custom Patterns` header (`UserPatternsMarker`) as a real section boundary. Only lines below that header in the backup are treated as user-authored.
- **REQ-CSI-008** — when the new template drops a pattern that the previous template contained and the user never added, do not migrate that pattern into the user section. The dropped-pattern case is the whole point of the header-parse: classification by header membership, not by not-in-template heuristic.
- **REQ-CSI-009** — emit each user pattern at most once, preserving the first occurrence's original text and discarding subsequent exact duplicates.
- **REQ-CSI-010** — remain idempotent: merging a file with its own previous output produces byte-identical content AFTER the header-parse and dedupe changes. The new idempotence test MUST exercise a case the pre-existing `TestMergeGitignore_DoubleReMerge_Idempotent` does not (a backup carrying a header AND duplicate user lines, so dedupe and re-merge interact), and the pre-existing test MUST still pass (AC-CSI-011).
- **REQ-CSI-011** — when a backup predates the header convention and contains no header, fall back to the current not-in-template heuristic rather than discarding the user's patterns. This is the R8 mitigation; without it, the header-parse becomes a data-loss regression.

**Tests**:
- `TestMergeGitignoreFile_TemplateDroppedPatternNotMigrated` (AC-CSI-007) — the §A probe promoted to a fixture
- `TestMergeGitignoreFile_DeduplicatesUserPatterns` (AC-CSI-008)
- `TestMergeGitignoreFile_IsIdempotent` (AC-CSI-009) — exercises header + duplicates together
- `TestMergeGitignoreFile_PreHeaderBackupFallsBackToHeuristic` (AC-CSI-010) — R8 mitigation
- Existing characterization test suite still passes (AC-CSI-011) — `TestMergeGitignore_DoubleReMerge_Idempotent` and `TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive` in particular; same-commit update with a comment naming this SPEC where a characterization test pins pre-fix behaviour, never deletion.

**Why a single commit** — REQ-CSI-007 (header boundary) without REQ-CSI-011 (heuristic fallback) is the R8 data-loss regression. The two MUST ship together.

### M3 — fallback visibility + §K reconciliation (REQ-CSI-012..014) — ONE commit, REQUIRES §K in place

The fallback-visibility change, shipped as a single commit. **M3 REQUIRES the §K reconciliation content to be in place in `spec.md`** (it is — see `spec.md` §K, migrated verbatim from the parent). Landing REQ-CSI-012 without the §K reconciliation is prohibited.

- **REQ-CSI-012** — when a 3-way merge fails and the restore path falls back to 2-way, the operator receives a stderr advisory naming the file, on the first occurrence rather than after three. This SUPERSEDES `REQ-UN-007`'s silent-first-occurrence clause. The counter (`fallback_count`) and the 3-strike threshold advisory survive (§K.2). The M3 commit body MUST state the supersession and reference `spec.md` §K (NFR-CSI-006).
- **REQ-CSI-013** — when the 3-way base file is absent and control falls through to 2-way without a merge attempt, the operator receives an advisory distinguishable from the merge-failure advisory of REQ-CSI-012. Today this path (`restore.go:118-120`, the `err == nil` guard failing on `os.ReadFile(basePath)`) emits neither a `recordFallback` call nor a warning; making it visible adds a signal rather than reversing a suppression, so REQ-CSI-013 conflicts with nothing and lands even if the §K reversal is rejected at Implementation Kickoff Approval.
- **REQ-CSI-014** — the REQ-CSI-012 advisory is at least as prominent as the existing 2-way failure warning at `restore.go:139-141`; the noise-suppression ledger continues to govern *repetition*, not *first occurrence*. The 3-strike threshold advisory (`REQ-UN-008`, which survives per §K.2) still marks sustained failure.

**Tests**:
- `TestRestore_ThreeWayFailureAdvisesOnFirstOccurrence` (AC-CSI-012) — first occurrence visible
- `TestRestore_AbsentBaseAdvisoryIsDistinct` (AC-CSI-013) — text differs from the merge-failure advisory
- `TestRestore_LedgerStillSuppressesRepetition` (AC-CSI-014) — repeated failures do not produce an unbounded advisory stream; the ledger still suppresses repetition

**Why this ordering** — M3 is last because (a) it depends on nothing in M1/M2 (the restore path is independent of the loader and the gitignore merger), and (b) its risk profile (R9, High/Med) is the highest in this SPEC and benefits from the §K reconciliation being settled before it lands. If the §K reversal is rejected at Implementation Kickoff Approval, M3 reduces to REQ-CSI-013 (the absent-base advisory) which conflicts with nothing.

## §F Anti-Patterns

- **AP-1 — Blanket `Save` refusal.** Refusing all writes when any section failed blocks five good sections for one bad one. REQ-CSI-006 requires the refusal to be per-section; the error returned by `Save()` names the failed section.
- **AP-2 — Escalating in the loader.** The loader records failure uniformly; the CLI-vs-hook split is a caller decision (B2). Escalating in the loader either breaks the fail-open norm (hook path) or silently swallows the malformed file (CLI path).
- **AP-3 — Header-parse without heuristic fallback.** REQ-CSI-007 without REQ-CSI-011 is the R8 data-loss regression. The two MUST ship together in one commit.
- **AP-4 — Deleting `TestMergeGitignore_ReMergeWithExistingHeader_PatternsSurvive`.** It pins behaviour M2 changes; the obligation is a same-commit update with a comment naming this SPEC, never deletion. Same for `TestMergeGitignore_DoubleReMerge_Idempotent`.
- **AP-5 — Landing REQ-CSI-012 without §K.** The reversal of an implemented sibling SPEC's whole purpose, made silently, is the exact defect §K exists to close. Without §K in place in `spec.md`, REQ-CSI-012 is unlandable.
- **AP-6 — Disabling the ledger's repetition governance.** REQ-CSI-014 keeps the ledger governing repetition. Making the first occurrence visible MUST NOT disable the 3-strike threshold advisory or the counter reset on success.
- **AP-7 — Bundling M1/M2/M3 into one commit.** NFR-CSI-005 requires the malformed-section refusal (M1) to ship in one revertable commit so a workflow that currently limps along on defaults can be restored to prior behaviour without unwinding M2/M3.

## §G Cross-References

- `spec.md` §D — REQ-CSI-001..014 (the requirements this plan implements).
- `spec.md` §K — Reconciliation with `SPEC-V3R6-UPDATE-NOISE-001` (the declared reversal, the rejected `--verbose` alternative, the surviving `REQ-UN-*` clauses). M3 REQUIRES this content be in place.
- `acceptance.md` §B — AC-CSI-001..014 (the acceptance criteria this plan's milestones satisfy).
- Parent `SPEC-CONFIG-TIER-PERSIST-001` §L — split-branch mapping (this child = slice (c)).
- Sibling `SPEC-CONFIG-ATOMIC-WRITE-001` — slice (b), closed. Owns the atomic-write helper; this child's `.moai/config/**` writers MAY adopt it. The helper's API is locked at slice-(b) `plan.md` M1.
- Sibling `SPEC-CONFIG-TIER-RESOLVE-001` — slice (a). Owns tier resolution; orthogonal.
- `.moai/specs/SPEC-V3R6-UPDATE-NOISE-001/` — `status: implemented`; its `REQ-UN-007` silent-first-occurrence clause is superseded by REQ-CSI-012 (§K).
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — the hook fail-open norm REQ-CSI-004 preserves.

## §H Self-Verification (manager-develop E1-E8, attribution per VCI §3)

When manager-develop reports completion, it MUST include:
- **E1** — AC Binary PASS/FAIL matrix (AC-CSI-001..014), each row naming the command + verbatim observed output.
- **E2** — Cross-platform build (`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`).
- **E3** — Coverage (`go test -cover ./internal/config/... ./internal/cli/update/...`).
- **E4** — Subagent-boundary grep (n/a for this SPEC's packages, but state it).
- **E5** — Lint status (NEW vs pre-existing baseline).
- **E6** — Branch HEAD + push state (per repo-local PR policy, Route B).
- **E7** — Blocker report (if any; NEVER call AskUserQuestion). In particular: if Implementation Kickoff Approval rejects the §K reversal, surface it as a blocker and reduce M3 to REQ-CSI-013.
- **E8** — RED failure output (TDD verbatim pre-GREEN evidence).

Each E-item names (a) the command, (b) the observed output, (c) the baseline-attribution (this run, this tree, HEAD SHA). Per VCI §3, summarized evidence like "all tests passed" is NOT acceptable.
