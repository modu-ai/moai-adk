# Sync-Audit Verdict — SPEC-PRECOMMIT-VET-MONOREPO-001 (card t237)

**Verdict: PASS** — harmonic-mean score **0.90** (4-dimension: Functionality 0.95 / Security 0.90 / Craft 0.90 / Consistency 0.85)

**Scope**: independent post-implementation audit of the sync phase, worktree `.claude/worktrees/t237`, branch `WT-precommit-vet-module`, audited HEAD `8f813ca86` (4 unpushed commits on develop baseline `b9298de32`: `bfdf833e8` plan → `45c60e1a3` run → `f964c73ed` sync close → `8f813ca86` D3 backfill). The known red `TestBinaryLag_DoctorCheckNameSetIsUnchanged` (t466 axis) is context, evidence-checked below, and NOT attributed to this card.

**Findings: 1 total — Medium ×1.** (F-1 below; one Low observation carried in Residual-risk.)

---

## Claim

11 of 12 acceptance criteria (AC-PVM-001..012) PASS as recorded; AC-PVM-010 records an honestly-attributed FAIL (inherited red, not authored by this card); the sync surface (CHANGELOG entry, `in-progress → completed` frontmatter transition on the sync commit, D3 two-commit `sync_commit_sha` backfill, SPEC-body untouched by sync) is correct; the twin pair is byte-identical; the named gaps (main `internal/cli` coverage unmeasured; space-in-path word-splitting) stay named.

## Evidence

Every command below was re-executed by this auditor on this run, against this tree (audited HEAD `8f813ca86`), with `MOAI_KANBAN*` env scrubbed in the same invocation for test commands. Verbatim outputs quoted.

### 1. AC matrix re-execution (mechanical)

**Contrast tests** — `go test -count=1 -run 'TestPreCommitHook_Submodule' ./internal/cli/...`:
```text
ok  	github.com/modu-ai/moai-adk/internal/cli	1.862s
ok  	github.com/modu-ai/moai-adk/internal/cli/agentlint	0.448s [no tests to run]
... (16 subpackages, all `[no tests to run]` — expected, selector scoped)
EXIT=0
```
GREEN as claimed (AC-PVM-002, AC-PVM-003).

**Byte-identity** — `go test -count=1 -run 'TestPreCommitTemplateMatchesConstant' -v ./internal/cli/`:
```text
=== RUN   TestPreCommitTemplateMatchesConstant
--- PASS: TestPreCommitTemplateMatchesConstant (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.687s
EXIT=0
```
(AC-PVM-001). The test asserts whole-file bytes `string(templateBytes) != preCommitHookContent` — not a normalized subset.

**Preserved-behavior guards batch** — `go test -count=1 -run 'TestPreCommitHook_GoVetBlocks|TestPreCommitHook_GofmtBlocks|TestPreCommitHook_SkipBypass|TestPreCommitHook_ToolchainAbsent|TestPreCommitHook_NoStagedGo' -v ./internal/cli/`:
```text
=== RUN   TestPreCommitHook_GofmtBlocks
--- PASS: TestPreCommitHook_GofmtBlocks (0.27s)
=== RUN   TestPreCommitHook_SkipBypass
--- PASS: TestPreCommitHook_SkipBypass (0.20s)
=== RUN   TestPreCommitHook_NoStagedGo
--- PASS: TestPreCommitHook_NoStagedGo (0.23s)
=== RUN   TestPreCommitHook_GoVetBlocks
--- PASS: TestPreCommitHook_GoVetBlocks (0.30s)
=== RUN   TestPreCommitHook_ToolchainAbsent
--- PASS: TestPreCommitHook_ToolchainAbsent (0.21s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.943s
EXIT=0
```
(AC-PVM-004, 007, 008, 009.) Additionally, `git show 45c60e1a3 -- internal/cli/hook_install_precommit_test.go | grep '^@@'` returns a SINGLE hunk `@@ -563,3 +563,93 @@` — a pure append after line 563; the guards' existing bodies are byte-untouched, as claimed.

**AC-PVM-012 exact formula** — `sed -n '/^## §E.3/,/^## §E.4/p' .moai/specs/SPEC-PRECOMMIT-VET-MONOREPO-001/progress.md | grep -c 'release card t204'`:
```text
1
EXIT=0
```
Expected 1, observed 1.

**Status at sync commit** — `git show f964c73ed:.moai/specs/SPEC-PRECOMMIT-VET-MONOREPO-001/spec.md | grep '^status:'`:
```text
status: completed
```
Current tree reads `status: completed` identically.

**AC-PVM-005 ordering (both surfaces, this tree)** — `grep -n 'BT_TAGS=""'` / `grep -n 'cd "\$_mr"'`:
```text
internal/cli/hook_install_precommit.go:  bt=95  first_cd=145   (95 < 145)
internal/template/templates/.git_hooks/pre-commit:  bt=52  first_cd=102   (52 < 102)
```
No ORDER-FAIL; matches the recorded values exactly.

**AC-PVM-006 subshell form** — `grep -c '( cd "\$_mr" && go vet' <both files>`:
```text
internal/cli/hook_install_precommit.go:1
internal/template/templates/.git_hooks/pre-commit:1
```

### 2. RED-cell authenticity — first-hand behavioral contrast on a real fixture

Beyond trusting the recorded RED, this auditor extracted the PRE-EDIT hook (`git show bfdf833e8:internal/template/templates/.git_hooks/pre-commit`) and the NEW hook (current template) and ran both as scripts against a purpose-built fixture: a `git init` repo inside `/tmp` whose ONLY `go.mod` (`module sample`, go 1.21) lives in `submod/`, with a staged gofmt-clean, vet-clean `submod/clean.go`.

OLD hook (expect exit 1 — the reported defect):
```text
[pre-commit] FAILED: go vet reported issues in the staged packages.
[pre-commit] Hint: run go vet on the affected packages, fix, then re-commit.
[pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit
OLD_HOOK_EXIT=1
```
Byte-identical to the recorded RED cell in progress.md §E.2. The old constant's shape was also verified at source: line 40 `printf './%s\n' "$(dirname "$f")"` (repo-root-relative paths), line 53 `if ! go vet $BT_TAGS $PKGS` (single invocation, no cd), line 54 the recorded failure message.

NEW hook, same fixture (expect exit 0):
```text
NEW_HOOK_CLEAN_EXIT=0
```
(No `[pre-commit] FAILED` line; the trailing WARN/graph-freshness lines come from the opt-in heavy-gate block firing because `moai` is on the auditor's PATH — a fixture-environment artifact, advisory, exit 0; the test harness shims moai absent for exactly this reason.)

NEW hook, after additionally staging vet-dirty `submod/bad.go` (`fmt.Printf("%d", "not-a-number")`):
```text
[pre-commit] FAILED: go vet reported issues in the staged packages (module root: submod).
[pre-commit] Hint: cd submod && go vet .
[pre-commit] Override: SKIP_MOAI_PRECOMMIT=1 git commit
NEW_HOOK_DIRTY_EXIT=1
```
The full contrast triangle holds first-hand: OLD blocks a clean submodule file (the defect), NEW passes it, NEW blocks a genuinely dirty file AND names the module root with a reproducible hint (REQ-PVM-003/007/008). The vacuous-green concern is discharged by observation, not assertion.

### 3. Twin-pair byte identity — auditor's own mechanical comparison

Method (stated per dispatch): extract the template verbatim from commit `8f813ca86`; extract the constant body from the same commit with `sed -n '/^const preCommitHookContent = `/,/^`$/p'` → strip the declaration prefix on line 1 (`sed '1s/^const preCommitHookContent = `//'`) → drop the closing-backtick line (`sed '$d'`) — this preserves the `#!/bin/sh` shebang, which lives on the declaration line inside the raw string (a naive range extraction drops it; the auditor's first attempt did, and was corrected). Result:
```text
/tmp/t237-a-template.txt: 134 lines
/tmp/t237-b-const2.txt:   134 lines
diff /tmp/t237-a-template.txt /tmp/t237-b-const2.txt → no output, DIFF_EXIT=0
```
Byte-identical at the audited HEAD, confirmed both by this diff (rc=0) and by `TestPreCommitTemplateMatchesConstant` (PASS above).

### 4. Inherited-red attribution re-verified (AC-PVM-010)

- `git show 45c60e1a3 --stat` lists exactly 6 files (`acceptance.md`, `progress.md`, `spec.md`, `internal/cli/hook_install_precommit.go`, `internal/cli/hook_install_precommit_test.go`, `internal/template/templates/.git_hooks/pre-commit`); greps for `doctor.go` and `binary_lag_test.go` in the touched set: **0 / 0** — the run commit touches neither of the red test's inputs.
- `git show b9298de32:internal/cli/doctor.go | grep -c 'Hook Delivery'` → **1**; `git show b9298de32:internal/cli/binary_lag_test.go | grep -c 'Hook Delivery'` → **0** — the red pre-exists the card at the develop baseline.
- Re-run on this tree: `go test -count=1 -run 'TestBinaryLag_DoctorCheckNameSetIsUnchanged' ./internal/cli/` →
  ```text
  --- FAIL: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.05s)
      binary_lag_test.go:205: this SPEC added doctor check name "Hook Delivery"; REQ-BLV-009 rewires the existing "Binary Freshness" item and registers no new name
  FAIL
  EXIT=1
  ```
  Byte-matches the recorded attribution. progress.md §E.2 records AC-PVM-010 as **"FAIL — inherited red, not authored by this card"** with the full chain (failing test, authoring commit `2b4582a43`, why this card's 4-file diff cannot reach the test's inputs, isolation repro) — a FAIL honestly recorded with evidence, not a pass.

### 5. Sync-phase surface

- `git diff --stat 45c60e1a3 f964c73ed` → exactly 3 files: `progress.md` (16 lines, §E.4), `spec.md` (1 line: `status: in-progress → completed` — the 3-phase close transition rides the sync commit, frontmatter only), `CHANGELOG.md` (+2: one entry + blank). plan.md and acceptance.md untouched by sync.
- `git diff f964c73ed 8f813ca86` → `progress.md` 1 line only: `sync_commit_sha: pending-backfill-sync` → `sync_commit_sha: f964c73ed` (D3 two-commit pattern, correct).
- plan.md untouched across ALL three post-plan commits (verified via the per-commit stats).
- CHANGELOG: entry count for the SPEC-ID = **1** (no duplicate); positioned at the top of `[Unreleased] → Added`, newest-first per the file's convention (above t466, itself above t462); headline and body are user-facing (behavior, before/after, how projects pick it up); the closing internal-flavored tail (AC counts, D3 mechanics) matches the established style of the sibling entries.
- CHANGELOG factual spot-checks: "byte-identical to the installer's hook constant, enforced by test" — re-proven above; "12 acceptance criteria, 11 PASS / 1 FAIL" — matches acceptance.md (AC-PVM-001..012) and the §E.2 matrix.

### 6. Run-phase surface sanity

`git show 45c60e1a3 -- internal/cli/hook_install_precommit.go | grep '^@@'` → a SINGLE hunk `@@ -75,31 +75,80 @@`, entirely inside `preCommitHookContent` (constant spans line 44 onward) — no installer machinery, imports, or functions touched, matching §E.3's "nothing outside the vet block changed". The re-authored block implements the SPEC's requirements as written: `_moai_module_root()` upward walk with `.` fallback (REQ-PVM-002), `MODROOTS` grouping, per-module `( cd "$_mr" && go vet $BT_TAGS $PKGS … )` subshell (REQ-PVM-001/005), `BT_TAGS` read before any cd (REQ-PVM-004), module-relative package paths (root `./dir`, module-root `.`, nested stripped prefix), failure message + hint naming the module root (REQ-PVM-003).

### 7. Named gaps stay named

- Main `internal/cli` package coverage unmeasured (inherited red suppresses `-cover`): named in §E.2 (line "Coverage (main internal/cli package) — **Unmeasured**"), §E.3 Gaps, and §E.4 gaps. Still unmeasured at sync — consistently a gap, never a claim.
- Space-in-path word-splitting: spec.md §E.1 Out of Scope + §E.3 Residual risk. The re-authored block indeed inherits unquoted `$MODROOTS`/`$PKGS` (source-verified at the `for _mr in $MODROOTS` loop) — documented, not fixed, correctly out of scope.

## Baseline-attribution

All evidence above was produced in this audit run, against this tree: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t237`, branch `WT-precommit-vet-module`, HEAD `8f813ca86` (full SHA `8f813ca86e008cdc6f23bc3de4fe74e58940b5e0`). Historical items were verified at their pinned commits: old hook extracted at `bfdf833e8` (plan commit, pre-edit constant), baseline greps at `b9298de32` (develop baseline), sync-transition diff at `f964c73ed`. Test binaries were compiled from the audited HEAD; the go toolchain resolved from PATH on this host (darwin/arm64). Fixture-contrast runs executed in `/tmp/t237-contrast2/repo` with the host go/gofmt on PATH.

## Gaps (explicitly NOT observed by this audit)

1. **Full `go test ./internal/cli/...` sweep NOT re-executed** by this auditor. The card's §E.2 records it as exit 1 with exactly the one inherited red; this audit re-ran the red test in isolation (FAIL reproduced, 0.05s) and the SPEC-scoped selectors, but did not re-run the ~463s full sweep. The inherited-red isolation repro plus the baseline greps make the attribution sound without it.
2. **AC-PVM-011 (cross-platform builds) NOT re-executed** — accepted from the §E.2 record (both exit 0 at run phase); the card's diff is shell-content-only inside a Go raw-string constant plus a test append, which cannot change windows buildability, but this audit observed no fresh build run.
3. **Installer end-to-end path not exercised** — no fresh `moai init` + hook-install roundtrip was run; installer correctness rests on the unchanged installer code (single-hunk proof) plus the existing install tests asserting `content == preCommitHookContent`.
4. **golangci-lint not re-run** — the §E.2 zero-new-findings claim vs the M1 baseline is accepted on record; the diff's shape (shell string + test append) makes new Go lint findings implausible but not observed here.
5. **`moai spec audit` / drift tooling not run** against the closed SPEC (era/drift classification on the completed frontmatter) — out of this audit's dispatched scope.

## Residual-risk

1. **F-1 (Medium — Consistency): the run commit modified `acceptance.md` BODY content.** `45c60e1a3` rewrote AC-PVM-012's verification formula (from `grep -c 't204' progress.md` ≥ 1 to the §E.3-scoped `sed … | grep -c 'release card t204'` form). The ownership matrix (spec-frontmatter-schema, Forbidden ownership crossings) reserves acceptance.md body for manager-spec; the mandated path was a blocker report → re-delegation. Mitigations, verified by this audit: the edit TIGHTENED the criterion — the original form was vacuously satisfiable by §E.1's plan-phase pre-mention, and the tightened form measured 0 on the plan-phase tree, flipping to 1 only because §E.3 genuinely carries the mandated sentence — so the change closed a vacuous-green hole rather than gaming a pass, and the tightened formula itself re-verified (value 1, above). Process violation with benign outcome; recorded for the lead, not verdict-blocking.
2. **Low observation:** the CHANGELOG tail says `sync_commit_sha` "is recorded as `pending-backfill-sync` in this commit" — true at authoring time and qualified by "in this commit", but on the post-backfill tree the recorded value is `f964c73ed`; the sentence now reads as a historical note. No action required.
3. The unquoted `$MODROOTS`/`$PKGS` word-splitting remains live in the shipped hook for module/package paths containing spaces — inherited, documented (spec.md §E.1), and correctly deferred to a future card; a monorepo with such a path would still see misbehavior.
4. The heavy-gate block fired during the auditor's fixture run because `moai` was on PATH — confirming the gate runs whenever moai is present; its advisory default kept the fixture exit clean, but environments where the heavy gate BLOCKS could mask/duplicate the fast-subset verdict. Pre-existing behavior, untouched by this card.
5. Every verdict here is local to the unpushed branch; the remote CI verdict on the integration branch remains the full-suite authority after the lead's batch push.

---

**Scores**: Functionality 0.95 (fix proven first-hand on a real fixture, both directions, plus the in-module block with naming; twin identity mechanically re-proven) · Security 0.90 (no new attack surface, no secrets, subshell-scoped cd, inherited word-splitting documented) · Craft 0.90 (single-hunk scoped change, append-only tests, verbatim two-cell RED/GREEN with tree SHAs) · Consistency 0.85 (3-phase close, D3 backfill, CHANGELOG conventions all correct; one acceptance.md ownership crossing) → harmonic mean **0.90**.

**Verdict: PASS** — 11/12 AC verified PASS by re-execution; AC-PVM-010's FAIL is honestly recorded and mechanically attributed to the t466/BinaryLag axis (red pre-exists at baseline `b9298de32`; run commit touches neither input; reproduced in isolation). One Medium process finding (F-1) reported to the lead; no verdict-blocking defect.
