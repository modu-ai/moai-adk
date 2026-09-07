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
- **AC-CSRB-007 re-anchored attribution**: doctor_codex.go now CARRIES t563's `osStatFn` seam via the absorb (provenance `c72dc1baf`, two call sites `:459`/`:857`). This card's diff to the file adds ZERO `osStatFn` tokens — measured `git diff b78d2e425 -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'` → `0` pre-commit; re-measured post-commit on the diff BODY of `c007e5409`: `git show --format='' c007e5409 -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'` → `0` (added-lines-only also `0`). Methodology note: the bare prescribed one-liner (without `--format=''`) prints `2` because `git show` includes the commit message, whose own attribution sentences mention "osStatFn" — the message's words, not diff lines (full record in `green-csrb-002-m2.md`).
- **Doctor guard counters IDENTICAL to the M1 baseline**: the only diff between `ac-csrb-006-baseline.log` and `green-csrb-002-m2.log` is the per-run TempDir path; every counter phrase byte-identical. Expected — darwin `sep='/'` makes the conversion the identity; a moved counter would have meant a darwin-visible behaviour change.
- **Coexistence with absorbed t563 tests**: `TestCodexStaleSkillFinding_|TestInspectSkillMirror_` scoped run → 27 PASS (26 absorbed + this card's guard), 0 FAIL.
- Evidence: `.moai/reports/t562/green-csrb-002-m2.log` + `green-csrb-002-m2.md`. vet rc=0; golangci-lint `0 issues.`

### M3 — doctor guard re-run, mutants, pins, verification

Scope: plan.md §E M3 exactly. **The production diff of this milestone is ZERO lines** — every code
edit was a deliberate temporary mutant, injected, measured, and reverted, with the tree proven
byte-clean after each revert (`git status --porcelain` on the production files empty; `git diff
--stat -- internal/` empty; HEAD never moved from `4aa8915ee` during the mutant window).
Full 5-section record: `.moai/reports/t562/run-m3.md`.

**AC-CSRB-006 — doctor regression guard re-run: PASS.**
- Command: `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_ShapeReadbackBaselineGuard' -count=1 -timeout 1800s -v` → exit 0 (`ac-csrb-006-post-m2.log`).
- `diff` against the M1 baseline shows exactly two differing lines, both per-run nonces: the
  `t.TempDir()` directory name and the package elapsed time (`0.843s` → `1.109s`). With both masked
  the diff is EMPTY. Every counter phrase — `declares 4 [[skills.config]] entries`, `1 with a path
  that no longer exists (1 enabled, 0 disabled, 0 unspecified, 0 non-boolean)`, `1 relative entry`,
  `1 oddly-formed entry` — is byte-identical to the pre-change capture.
- Expected on darwin: `configPathSeparator == '/'` makes the M2 conversion the identity; a moved
  counter would have meant a darwin-visible behaviour change.

**Mutant reachability probes — all three sites REACHED before any mutant was credited.**
An unreached mutant and a real survivor print the same `ok`, so each site was first proven executed
with an injected `panic`: P1 `codex_skills_prune.go:83` (conversion) under AC-CSRB-002, P2 `:101`
(stat call) under AC-CSRB-004, P3 `:76` (`classifyCodexSkillPath` switch) under AC-CSRB-003 — each
panicked with its probe string and a stack frame naming the site (`probe-p1/p2/p3.log`).

**Mutants — 4 injected, 4 CAUGHT, 0 missed, 0 committed, 0 surviving.**

| Mutant | Judge | Verdict | Observed failure |
|---|---|---|---|
| bypass (`statPath = e.Path`) | AC-CSRB-002 | CAUGHT | `stat target = "/var/…/gone/SKILL.md", want the converted form "\\var\\…"` |
| blanket-wrap (wrap at stat site, branch conversion removed) | AC-CSRB-004 | CAUGHT | `stat target = "\\var\\…\\x\\SKILL.md", want the native Join product "…/x/SKILL.md" unrewritten` |
| reorder (convert before classify) | AC-CSRB-003 | CAUGHT | `SkipReason = "oddly-formed path — not resolvable here", want "relative path — no observed resolution base"` |
| seam (`osStatFn` line added to `doctor_codex.go`) | AC-CSRB-007 delta predicate | CAUGHT | added-osStatFn-line count `0 → 1`; pin FAILS |

The seam mutant was exercised as an UNCOMMITTED working-tree change under the same delta predicate
(committing a mutant is prohibited); the committed axis is covered instead by the per-commit sweep
below. Logs: `mutant-bypass.log`, `mutant-blanket-wrap.log`, `mutant-reorder.log`, `mutant-seam.log`.

**AC-CSRB-007 — doctor seam-out pin: PASS.**
- `git show c007e5409 --format='' -- internal/cli/doctor_codex.go | /usr/bin/grep -c 'osStatFn'` → **0**
  (added-lines-only `^+` also 0), with the POSITIVE CONTROL
  `/usr/bin/grep -c 'osStatFn' internal/cli/codex_skills_prune.go` → **1**. Neither operand is empty.
- Provenance context: `/usr/bin/grep -c 'osStatFn' internal/cli/doctor_codex.go` → 2 (absorbed t563).
- Methodology witness: the same command WITHOUT `--format=''` prints 2 — the commit message's own
  attribution prose, not diff lines.
- Generalized across EVERY card-authored commit (`fe2c8f51f`, `f1654c924`, `b78d2e425`, `c007e5409`,
  `835215bab`, `4aa8915ee`): added-osStatFn-lines = **0** on all six; the absorb merge `2d1dad058`
  adds 2. Clean attribution. Evidence: `ac-csrb-007.log`.

**AC-CSRB-008 — scope pin: PASS.**
- `git fetch origin develop` (→ `91d25bc61`); `CARD_BASE=$(git merge-base origin/develop HEAD)` →
  `c72dc1baf`, re-derived at read time, never a pinned SHA.
- `git diff --name-only "$CARD_BASE"..HEAD` → **31 paths**, all inside the plan.md §G allowlist:
  8 under `internal/cli/` (this card's 4 + the 4 absorbed t540 files) and 23 under `.moai/`.
- Probe `-- internal/codexwiring/skills.go` → EMPTY, against the non-zero control of the 31-path
  listing.
- PRESERVE strengthened: `git diff --stat 2d1dad058..HEAD -- <the four t540 files>` → EMPTY, against
  the non-zero control `10 files changed, 227 insertions(+), 30 deletions(-)` over the same range.
  Zero card-authored change to the PRESERVE surface after the absorb. Evidence: `ac-csrb-008.log`.

**AC-CSRB-009 — cross-platform build: PASS (structural only).**
- `go build ./...` rc=0; `GOOS=windows GOARCH=amd64 go build ./...` rc=0.
- **Stated limit**: a GOOS cross-build does NOT compile `*_test.go` files. Added as a stronger
  witness: `GOOS=windows GOARCH=amd64 go vet ./internal/cli/` rc=0, which does type-check test files
  under the target GOOS. Still compile-level — no Windows runtime behaviour is claimed or observed.
  Evidence: `ac-csrb-009.log`.

**AC-CSRB-010 — scoped suite with executed-test control: PASS against a lower-bound BEFORE.**
- `go test ./internal/cli/... -timeout 1800s -v` → **rc=0**; `--- PASS: ` = **6938**; `--- FAIL: ` = 0;
  `^FAIL` = 0; `--- SKIP: ` = 30. Trailing space in the pattern is load-bearing (go appends
  ` (0.06s)`, so a `$` anchor would count 0 on every line).
- BEFORE = **6886**, carried from `.moai/reports/t540/ac-006-base.log` (same command, t540 pre-flight,
  tree `b4ce67468`). Predicate `6938 >= 6886 AND 6938 > 0` holds.
- This card contributes 8 `--- PASS: ` lines (5 top-level + 3 subtests), all present in the AFTER log;
  AC-CSRB-003/004/005 are confirmed GREEN in this same full-suite run, not carried from M2.
- **Gap (reported, not folded into the pass)**: the t562-own pre-flight base
  `.moai/reports/t562/ac-010-base.log` required by plan.md §C was never captured and cannot be
  captured now. 6886 predates t540's own tests and t563's absorbed tests, so it is a strict LOWER
  bound, and the +52 delta is not decomposed beyond this card's named 8. See `run-m3.md` §4 G1 / §5 R1.
- Evidence: `ac-csrb-010.log`, raw `ac-010.log`.

**Milestone-boundary verification:** `go vet ./internal/cli/...` rc=0; `golangci-lint run
--timeout=5m ./internal/cli/...` → `0 issues.` (`quality-gates-m3.log`). Full repository suite NOT run
locally — that verdict is CI's, on the pushed head.

**Tree attribution (lead's CWD-drift warning, applied post-hoc).** `/Users/goos/moai/moai-adk-go` and
`/Users/goos/MoAI/moai-adk-go` are ONE directory under two spellings (identical dev:inode
`16777231:253706617`); the real hazard axis is worktree → primary drift. Measured: of the files this
card reads, only `doctor_codex.go` exists in BOTH trees with different content (`osStatFn` count: 0 in
the primary at `main`, 2 in the worktree) — the one read where drift would have succeeded silently.
Every recorded M3 value was RE-MEASURED with the tree pinned absolutely (`git -C <abs>`, `go -C <abs>`,
absolute file paths) and **all reproduced identically**; the AC-CSRB-008 predicate "paths outside
`internal/cli/` + `.moai/` = 0" was additionally counted rather than eyeballed. Two measurements were
not re-executed and are attributed by positive witness instead: the AC-CSRB-010 suite log carries a
test name that exists in ZERO files of the primary checkout, and the probe logs carry go panic stack
frames printing the worktree path verbatim. Recorded as gap G8, not silently kept.
Evidence: `tree-attribution-remeasure.log`, `ac-csrb-006-remeasure.log`, `run-m3.md` §7.

All line citations in this card's M3 evidence (`codex_skills_prune.go:76/:83/:101`,
`doctor_codex.go:861`) are pinned to tree `.claude/worktrees/t562`, branch `WT-codex-read-inverse`,
SHA `4aa8915ee` — unchanged at `757ef601f`, since M3 added zero production lines.

### §E.2 AC matrix (run-phase, complete)

| AC | Status | Evidence |
|---|---|---|
| AC-CSRB-001 | PASS | `ac-csrb-001.txt` |
| AC-CSRB-002 | RED-now observed at M1 → **GREEN at M2**; bypass mutant CAUGHT at M3 | `red-csrb-002.log`, `red-csrb-002.md`, `green-csrb-002-m2.log`, `green-csrb-002-m2.md`, `mutant-bypass.log` |
| AC-CSRB-003 | **PASS** — guard green; reorder mutant CAUGHT (site reached, probe P3) | `mutant-reorder.log`, `probe-p3.log`, `ac-010.log` |
| AC-CSRB-004 | **PASS** — guard green; blanket-wrap mutant CAUGHT (site reached, probe P2) | `mutant-blanket-wrap.log`, `probe-p2.log`, `ac-010.log` |
| AC-CSRB-005 | **PASS** — all 3 arms green in the full-suite run; verdicts only, no deletion | `ac-010.log`, `ac-csrb-010.log` |
| AC-CSRB-006 | **PASS** — post-M2 re-run reproduces the M1 baseline; only per-run nonces differ | `ac-csrb-006-baseline.log`, `ac-csrb-006-post-m2.log` |
| AC-CSRB-007 | **PASS** — delta 0 with positive control 1; all 6 card commits 0; seam mutant CAUGHT | `ac-csrb-007.log`, `mutant-seam.log` |
| AC-CSRB-008 | **PASS** — 31 paths all allowlisted; probe empty vs non-zero control; PRESERVE 0 | `ac-csrb-008.log` |
| AC-CSRB-009 | **PASS** — both builds rc=0; test-file compile limit stated | `ac-csrb-009.log` |
| AC-CSRB-010 | **PASS** against a lower-bound BEFORE (gap G1 reported) | `ac-csrb-010.log`, `ac-010.log` |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete   # M1 + M2 (incl. STEP-0 repair) + M3 complete
run_complete_at: 2026-09-08
run_commit_sha: c007e5409   # the M2 production-diff commit (the card's only production change), backfilled in the M3 commit per the D3 exemption
ac_pass_count: 10       # AC-CSRB-001 … AC-CSRB-010 all PASS (AC-CSRB-010 against a lower-bound BEFORE; gap G1 named)
ac_fail_count: 0
preserve_list_post_run_count: 0   # measured: git diff --stat 2d1dad058..HEAD -- <the four t540 files> EMPTY, against a non-zero control over the same range
l44_pre_commit_fetch: performed at commit time (M1 report; re-performed at M3 for AC-CSRB-008 — origin/develop 91d25bc61, CARD_BASE c72dc1baf re-derived)
l44_post_push_fetch: n/a — no push (lane reports merge SHA; push is the lead's)
new_warnings_or_lints_introduced: 0   # go vet ./internal/cli/... rc=0; golangci-lint run ./internal/cli/... -> "0 issues."
cross_platform_build:
  darwin_native: "go build ./... rc=0 (AC-CSRB-001 at M1, re-measured AC-CSRB-009 at M3)"
  windows: "GOOS=windows GOARCH=amd64 go build ./... rc=0 (AC-CSRB-009). LIMIT: go build does not compile *_test.go; the added witness GOOS=windows go vet ./internal/cli/ rc=0 does type-check them. Compile-level only — no Windows runtime behaviour observed."
total_run_phase_files: 31   # 16 in the committed run-phase commits (4 internal/cli + 4 SPEC artifacts + 8 evidence) + 15 new M3 evidence files
m1_to_mN_commit_strategy: one commit per milestone plus a standalone STEP-0 repair commit (M1 tests+evidence f1654c924; STEP-0 rename b78d2e425; M2 the two-line production diff c007e5409 + GREEN evidence; M3 mutants+pins+verification — M3's production diff is ZERO lines, all mutants reverted)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-08
sync_commit_sha: 18bf8cc06   # backfilled in the following commit per spec-frontmatter-schema.md D3 (a commit cannot cite its own hash)
sync_status: complete
b12_self_test_a: "pre-emission grep — /usr/bin/grep -c 'SPEC-CODEX-SKILL-PATH-READBACK-001' CHANGELOG.md -> 0 (no duplicate; emission proceeds)"
b12_self_test_b: "AC count vs acceptance.md (SSOT) — /usr/bin/grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l -> 11 raw uniques, of which AC-CGM-011 (acceptance.md:203) is a CROSS-SPEC reference to t533 layer 2, not an AC of this SPEC. Live count for this SPEC = 10 (AC-CSRB-001..010), matching the §E.3 ac_pass_count and the CHANGELOG entry's stated 10. Reserved-token markers ([RETIRED]/[REF]): 0 occurrences, so no marked-identifier exclusion applies and the count is unambiguous."
b12_self_test_c: "file-path verification — ls of every path named in the CHANGELOG entry (internal/cli/codex_skills_prune.go, internal/cli/doctor_codex.go, internal/cli/codex_config_path.go, internal/cli/codex_skills_prune_readback_test.go, internal/cli/codex_stale_skill_readback_test.go, .moai/specs/SPEC-CODEX-SKILL-PATH-SLASH-001/spec.md, .moai/reports/t562/run-m3.md) -> all present, exit 0"
changelog_entry_position: "CHANGELOG.md [Unreleased] -> ### Fixed, first bullet (immediately above SPEC-DOCTOR-STAT-SEAM-001)"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed, merged into this single sync commit (status + updated only; §A-§H body untouched)"
  plan_md: "n/a — measured: plan.md carries NO YAML frontmatter block, so it has no status: field to transition (artifact statelessness, spec-frontmatter-schema.md)"
  acceptance_md: "n/a — measured: acceptance.md carries NO YAML frontmatter block (its line-14 '---' is a horizontal rule), so it has no status: field to transition"
  progress_md: "n/a — progress.md records phase state in body sections (§E.1-§E.4), not in frontmatter; §E.2/§E.3 run-phase content untouched by this sync"
canary_compliance_check:
  applicable: false
  reason: "this SPEC defines no forward-looking policy that its own sync phase would test; the change is two one-line production conversions in internal/cli"
docs_synchronized: "none — no user-facing document (README 4-locale set, docs-site ko/en/ja/zh) describes the path-separator readback behaviour; verified by grep over docs-site/content/**/{moai-clean,doctor}.md and README*.md for separator/backslash/Windows path statements. No doc change was manufactured to look complete."
gaps_carried_into_sync:
  - "G1 (load-bearing): no t562-own AC-CSRB-010 BEFORE baseline; .moai/reports/t562/ac-010-base.log does not exist and the pre-absorb tree is gone. BEFORE 6886 is t540's capture at tree b4ce67468 (same command) and is a STRICT LOWER BOUND. AC-CSRB-010 is PASS against a lower bound; acceptance.md's 'expected delta accounted' clause is only partially satisfied. Residual risk R1: up to 52 silently-lost tests would still satisfy AFTER >= BEFORE, and FAIL: 0 does not close it."
  - "G2: no Windows runtime observation — GOOS=windows build rc=0 and GOOS=windows vet rc=0 are compile/type-check level only; a GOOS cross-build does not compile *_test.go."
  - "G3: the seam mutant was enforced as an uncommitted working-tree change, so the demonstrated property is that the delta discriminant flips 0->1, not that a committed form flips the fixed-SHA command (insensitive to later commits by construction). The commit axis is covered by the per-commit sweep, 6/6 = 0."
  - "G4: full repository suite not run locally (CI's verdict, on the pushed head)."
  - "G5: no -race run."
  - "G6: the doctor guard exercises the darwin identity path only."
  - "G7: 30 SKIPs counted but not enumerated or attributed."
  - "G8: two measurements (the AC-CSRB-010 suite run and the probe/mutant runs) were tree-attributed by positive witness rather than re-executed under absolute tree pinning. Witness establishes WHERE the recorded run happened; it says nothing about whether a fresh run would reproduce it."
```

Sync-phase scope actually performed: CHANGELOG `[Unreleased]` → `### Fixed` entry (drafted after reading
`internal/cli/codex_skills_prune.go` and `internal/cli/doctor_codex.go`, not from plan.md prose), this
§E.4 block, and the frontmatter `in-progress → implemented → completed` transition on `spec.md` — the
only artifact of the four that carries a `status:` field (measured, not assumed). No SPEC body content
(`§A`–`§H` of spec.md / plan.md / acceptance.md) was modified.

## PRESERVE carried forward

- `internal/cli/codex_config_path.go`, `codex_config_path_test.go`, `codex_skills_disable.go`,
  `codex_skills_disable_path_test.go` — t540's absorbed surface, read-only this card (REQ-CSRB-006).
- `internal/codexwiring/skills.go` — probe-empty at M3 (REQ-CSRB-006).
- `internal/cli/doctor_codex.go` — NO osStatFn seam may be introduced (t563's scope; AC-CSRB-007).
