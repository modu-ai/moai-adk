# SPEC-CODEX-SKILL-PATH-READBACK-001 — progress

Card t562 · worktree `.claude/worktrees/t562` · branch `WT-codex-read-inverse` · Tier M ·
cycle_type tdd · semi-autonomous (lane reports at each milestone boundary; M2/M3 start only after
the lead's go).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-08
plan_audit_verdict: "PASS (plan-auditor, iter-2, overall 1.00 — iter-1 0.94, D1 fixed; see .moai/reports/t562/plan-audit.md)"
```

Plan artifacts authored by manager-spec (plan-phase); the audit record lives at
`.moai/reports/t562/plan-audit.md`.

## §E.2 Run-phase Evidence

### M1 — absorb gate, the RED, and the doctor baseline (this milestone)

Scope: plan.md §E M1 exactly. The production diff is ZERO lines — M1 is tests + evidence only;
M2 owns the two one-line conversions.

**AC-CSRB-001 — absorb gate: PASS.**
- Command: `go build ./...` → exit 0, no output.
- Seam symbols measured (`/usr/bin/grep`): `codex_config_path.go:39` `var configPathSeparator`,
  `:51` `toConfigPath`, `:61` `fromConfigPath`; publisher conversions at
  `codex_skills_disable.go:247` and `:277`.
- Evidence: `.moai/reports/t562/ac-csrb-001.txt` (tree `69cfdce6f`).
- The `depends_on` pre-flight reading (SPEC-CODEX-SKILL-PATH-SLASH-001 absent before the merge)
  is satisfied by the absorb merge itself per spec.md §B.4 — recorded satisfied at M1 completion.

**AC-CSRB-002 — prune stat-target conversion (behavioral RED): RED-now OBSERVED.**
- Command: `go test ./internal/cli/ -run 'TestJudgeCodexSkillEntry_SeparatorConversion' -count=1 -timeout 1800s -v` → exit 1.
- Verbatim failing line: `stat target = "/var/folders/.../gone/SKILL.md", want the converted form "\\var\\folders\\...\\gone\\SKILL.md"` — the recorder observed the DECLARED form (`statPath = e.Path` verbatim), exactly the conversion absence M2 will close.
- Final test name pinned: `TestJudgeCodexSkillEntry_SeparatorConversion` (in `internal/cli/codex_skills_prune_readback_test.go`; the acceptance.md placeholder name shipped unchanged).
- Evidence: `.moai/reports/t562/red-csrb-002.log` + `red-csrb-002.md` (tree `69cfdce6f`).
- The fix is deliberately NOT written here — M2 owns it (RED-now observed before any fix is this milestone's point).

**AC-CSRB-003 / 004 / 005 — ordering, no-double-conversion, eligibility pins: GREEN-before (guards).**
- Command: `go test ./internal/cli/ -run 'TestJudgeCodexSkillEntry_(ClassifiesDeclaredFormBeforeConversion|HomeRelativeStatTargetStaysNative|EligibilityGatingPins)' -count=1 -timeout 1800s -v` → PASS (3 top-level + 3 subtests).
- Guards claim no RED (plan.md §E M1); their discriminating power is recorded as the M3 mutants (reorder / blanket-wrap), not executed this milestone.

**AC-CSRB-006 — doctor regression guard: baseline CAPTURED in this window.**
- Command: `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard' -count=1 -timeout 1800s -v` → exit 0.
- Guard test authored in this M1 window (`internal/cli/codex_stale_skill_readback_test.go`), BEFORE M2 lands, per the plan-audit D1 amendment.
- Exercised-count recorded next to the raw output: **4 fixture entries traversed** (1 resolved, 1 missing/enabled-bucket, 1 relative, 1 oddly-formed; the `declares 4` phrase is the parsed-count witness).
- Evidence: `.moai/reports/t562/ac-csrb-006-baseline.log` + `ac-csrb-006-baseline.md` (tree `69cfdce6f`).
- Corroboration only: `.moai/reports/t540/ac-006-base.log` is t540's AC-CSPS-006 executed-test control (per t540 `run-m1-m2.md` §2.7), not a doctor capture; the binding baseline for this AC is the M1 capture above.

**AC-CSRB-007 … 010 — not yet attempted** (M3 scope: seam-out pin, scope pin, cross-platform build, package suite with executed-test control).

**Milestone-boundary verification (scoped, per the card's constraints):**
- `go vet ./internal/cli/` → exit 0.
- `golangci-lint run --timeout=2m ./internal/cli/` → `0 issues.`
- Full package suite NOT run locally (the `internal/cli` 1800s floor belongs to M3's AC-CSRB-010 run; the full-suite verdict is CI's).

### M2 — the two one-line conversions (STEP 0 collision repair + production diff)

**STEP 0 — test symbol collision repair (commit `b78d2e425`, BEFORE any M2 edit).**
- The absorb merge `2d1dad058` (origin/develop `c72dc1baf`, carrying t563's doctor stat seam + its test file) broke the package test binary's compilation: both this card's test and the absorbed `doctor_codex_stale_skill_test.go:331` declared a package-scope type `statRecorder` (theirs carries a `paths` field — `rec.paths` cascade).
- Enumeration of ALL 9 package-scope identifiers of this card's two test files against every other `*_test.go` in the package: exactly ONE collision (`statRecorder`); the other 8 CLEAN.
- Repair: rename to `pruneReadbackStatRecorder` in THIS card's file only; the absorbed file and production files untouched.
- Evidence: `.moai/reports/t562/step0-symbol-enumeration.md`. Post-rename M1 state verified unchanged (SeparatorConversion FAIL, guards PASS, vet clean).

**AC-CSRB-002 — GREEN observed (the M2 commit).**
- Command: `go test ./internal/cli/ -run 'TestJudgeCodexSkillEntry_(SeparatorConversion|ClassifiesDeclaredFormBeforeConversion|HomeRelativeStatTargetStaysNative|EligibilityGatingPins)|TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard' -count=1 -timeout 1800s -v` → exit 0.
- `TestJudgeCodexSkillEntry_SeparatorConversion`: FAIL at M1 → **PASS at M2**. All guards stayed PASS.
- Production diff: exactly the two one-line conversions — `judgeCodexSkillEntry` and `codexStaleSkillFinding`, `codexPathAbsolute` branch only, `statPath = fromConfigPath(e.Path, configPathSeparator)`. Home-relative branches untouched.
- **AC-CSRB-007 re-anchored attribution**: doctor_codex.go now CARRIES t563's `osStatFn` seam via the absorb (provenance `c72dc1baf`, two call sites `:459`/`:857`). This card's diff to the file adds ZERO `osStatFn` tokens — measured `git diff b78d2e425 -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'` → `0` pre-commit; re-measured post-commit via `git show` (see session report / `green-csrb-002-m2.md`).
- **Doctor guard counters IDENTICAL to the M1 baseline**: the only diff between `ac-csrb-006-baseline.log` and `green-csrb-002-m2.log` is the per-run TempDir path; every counter phrase byte-identical. Expected — darwin `sep='/'` makes the conversion the identity; a moved counter would have meant a darwin-visible behaviour change.
- **Coexistence with absorbed t563 tests**: `TestCodexStaleSkillFinding_|TestInspectSkillMirror_` scoped run → 27 PASS (26 absorbed + this card's guard), 0 FAIL.
- Evidence: `.moai/reports/t562/green-csrb-002-m2.log` + `green-csrb-002-m2.md`. vet rc=0; golangci-lint `0 issues.`

### §E.2 AC matrix (run-phase, as of M2)

| AC | Status | Evidence |
|---|---|---|
| AC-CSRB-001 | PASS | `ac-csrb-001.txt` |
| AC-CSRB-002 | RED-now observed at M1 → **GREEN at M2** | `red-csrb-002.log`, `red-csrb-002.md`, `green-csrb-002-m2.log`, `green-csrb-002-m2.md` |
| AC-CSRB-003 | GREEN-before guard (mutant = reorder, M3) | scoped run, this milestone |
| AC-CSRB-004 | GREEN-before guard (mutant = blanket-wrap, M3) | scoped run, this milestone |
| AC-CSRB-005 | GREEN-before guard (pins; verdicts only, no deletion) | scoped run, this milestone |
| AC-CSRB-006 | Baseline captured (M3 re-runs + diffs) | `ac-csrb-006-baseline.log`, `ac-csrb-006-baseline.md` |
| AC-CSRB-007..010 | Not yet attempted (M3) | — |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: partial   # M1 + M2 (incl. STEP-0 repair) complete; M3 pending the lead's go
run_complete_at: ""
run_commit_sha: pending-backfill-m2   # the M2 production-diff commit; backfilled in the M3 commit per the D3 exemption
ac_pass_count: 2        # AC-CSRB-001 PASS; AC-CSRB-002 flipped RED→GREEN at M2
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: performed at commit time (see M1 report)
l44_post_push_fetch: n/a — no push (lane reports merge SHA; push is the lead's)
new_warnings_or_lints_introduced: 0   # vet exit 0; golangci-lint 0 issues on internal/cli
cross_platform_build:
  darwin_native: "go build ./... rc=0 (AC-CSRB-001)"
  windows: "not yet run — M3 (AC-CSRB-009) scope"
total_run_phase_files: 4   # two new test files, two new evidence md/log pairs + ac-csrb-001.txt + this file (see commit)
m1_to_mN_commit_strategy: one commit per milestone plus a standalone STEP-0 repair commit (M1 tests+evidence f1654c924; STEP-0 rename b78d2e425; M2 the two-line production diff + GREEN evidence; M3 mutants+pins+verification)
```

## PRESERVE carried forward

- `internal/cli/codex_config_path.go`, `codex_config_path_test.go`, `codex_skills_disable.go`,
  `codex_skills_disable_path_test.go` — t540's absorbed surface, read-only this card (REQ-CSRB-006).
- `internal/codexwiring/skills.go` — probe-empty at M3 (REQ-CSRB-006).
- `internal/cli/doctor_codex.go` — NO osStatFn seam may be introduced (t563's scope; AC-CSRB-007).
