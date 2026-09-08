# SPEC-CODEX-DOCTOR-PATH-GUARD-001 — Progress

Card t570 · worktree `.claude/worktrees/t570` · branch `WT-codex-doctor-guard` · base `a4855f0b2`.

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier S; `acceptance.md` written
  at the lead's explicit request — Tier S normally inlines AC into `spec.md` §3).
- SPEC ID regex check executed as Bash at `a4855f0b2`:
  `ID="SPEC-CODEX-DOCTOR-PATH-GUARD-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`
- ID uniqueness: `ls .moai/specs | /usr/bin/grep -i "DOCTOR-PATH-GUARD"` → no match (rc=1).
- `development_mode: tdd` read this run from `.moai/config/sections/quality.yaml`.
- Requirements in GEARS notation; exclusions section carries four `### Out of Scope —` sub-headings.
- Status: `draft`. No production file touched by plan-phase.

### Plan-audit repair (2026-09-08, post-audit, still `draft`)

plan-audit returned PASS 0.87 with two blocking defects plus three coordinate corrections. All
repaired in place; every coordinate was independently re-derived in this tree before editing.

| Defect | Repair |
|---|---|
| D4 (major) — REQ-005/006/007 had no criterion carrier | Added AC-CDPG-005 (helper-name collision count) and AC-CDPG-007 (test-only scope, two-half git predicate); REQ-006 given an explicit no-AC-by-design section + named DoD carrier; added a REQ-to-carrier map. `/usr/bin/grep -nE '^### AC-' acceptance.md` now returns 7 rows covering REQ-001..007. |
| D5 (minor) — AC-001 RED-now cell claimed an executed revert | Rewritten in obligation tense; sibling cell in AC-002 swept likewise; AC-004 title changed to "to be executed"; a **Tense contract** paragraph added at the head of `acceptance.md`. |
| D1 — helper file coordinates wrong | `stubCodexHome` / `writeCodexHomeConfig` corrected to `doctor_codex_test.go:103/:122`; table now names four distinct files. Re-measured this run. |
| D2 — "three mutants" | Corrected to four logs (`bypass`, `blanket-wrap`, `reorder`, `seam`), with the seam mutant's uncommitted-working-tree status noted; t562's own §J sentence quoted verbatim in place of a count. |
| D3 — F6 credit misattributed | Corrected: t562's sync-audit F6 named the extended-length family and issued it forward; this card discharges it rather than discovering it. §F now cites F2 and F6 by `progress.md` line. |
| D6 (optional) — REQ-007 `Should` outside GEARS | Applied: rewritten as a `When … shall` unwanted-behaviour form with a mechanical predicate. |
| D7 (optional) — unpinned fixture | Applied: fixture pinned to `/Users/u/skills/probe/SKILL.md`, with a one-line non-vacuity note. |

None declined. Baselines measured at `a4855f0b2`: `type statRecorder` → exactly 1 declaration
(`doctor_codex_stale_skill_test.go:331`); working tree still carries no production change
(`git status --short` → only `.moai/` paths untracked).

## §E.2 Run-phase Evidence

Measured in worktree `.claude/worktrees/t570`, branch `WT-codex-doctor-guard`, at plan HEAD
`076847abab17064a65a89b3e7424b9ec1bf5e4a8`. Card base `a4855f0b2834f5179d07ebcd4c72ff0b452d6aa3`.
The full 5-section evidence record (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk),
including verbatim RED output and the mutant table, is `.moai/reports/t570/run-evidence.md`.

Deliverable: `internal/cli/doctor_codex_path_guard_test.go` — three tests, all non-parallel, all
overrides restored via `t.Cleanup`, no new type declared.

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-CDPG-001 | PASS | `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_(Absolute\|ExtendedLength)StatTargetIsConvertedForm' -timeout 1200s -v` | `--- PASS: TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm (0.00s)` |
| AC-CDPG-002 | PASS | same invocation | `--- PASS: TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm (0.00s)` |
| AC-CDPG-003 | PASS | `go test ./internal/cli/ -run 'TestCodexStaleSkillFinding_ClassificationPrecedesConversion' -timeout 1200s -v` | `--- PASS: TestCodexStaleSkillFinding_ClassificationPrecedesConversion (0.00s)` |
| AC-CDPG-004 | PASS | mutants M-1..M-4 executed; every production mutant reverted and the revert verified empty | M-1 CAUGHT (`m1-red.log` → `m1-green.log`), M-2 CAUGHT (`m2-red.log` → `m2-green.log`), M-3 CAUGHT (`m3-red.log` → `m3-green.log`), **M-4 MISSED** (`m4-missed.log`) |
| AC-CDPG-005 | PASS | `/usr/bin/grep -rn 'type statRecorder' internal/cli/ \| wc -l` | `1` (unchanged from the `a4855f0b2` baseline); `type pruneReadbackStatRecorder` likewise `1` |
| AC-CDPG-006 | discharged by constraint (no AC by design) | read the three test functions | no `t.Parallel()`; every override (`overrideSeparator`, `stubStatRecording`, `stubCodexHome`) restores through `t.Cleanup` |
| AC-CDPG-007 | PASS | `git diff a4855f0b2834f5179d07ebcd4c72ff0b452d6aa3 -- internal/cli/doctor_codex.go` | empty (0 bytes) |

**Invariants.** The existing suite is unbroken: `go test ./internal/cli/... -timeout 1200s` → `rc=0`,
17 packages `ok`, `internal/cli` at `611.738s` (`full-package.log`) — which re-measures in this run
why `-timeout 1200s` is mandatory. `go vet ./internal/cli/` → exit 0, no output.

**RED evidence (E8), verbatim, captured under mutant M-1 before the guard could pass:**

```
=== RUN   TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm
    doctor_codex_path_guard_test.go:52: stat target = "/Users/u/skills/probe/SKILL.md", want the converted form "\\Users\\u\\skills\\probe\\SKILL.md"
    doctor_codex_path_guard_test.go:55: stat target = "/Users/u/skills/probe/SKILL.md" — the DECLARED form: the conversion is not applied
--- FAIL: TestCodexStaleSkillFinding_AbsoluteStatTargetIsConvertedForm (0.00s)
=== RUN   TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm
    doctor_codex_path_guard_test.go:93: stat target = "//?/C:/Users/u/skills/probe/SKILL.md", want the converted extended-length form "\\\\?\\C:\\Users\\u\\skills\\probe\\SKILL.md"
--- FAIL: TestCodexStaleSkillFinding_ExtendedLengthStatTargetIsConvertedForm (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.186s
```

**AC-CDPG-003 non-discriminating statement (required verbatim by `acceptance.md`):** a zero-stat-call
assertion does NOT discriminate the two classification orders. Both the `codexPathRelative` and the
`codexPathOddlyFormed` branches `continue` before reaching `osStatFn`. Confirmed empirically by the
M-2 run: under the forbidden order the two Detail assertions fired while the `len(rec.paths) != 0`
assertion stayed silent.

**Missed mutant, recorded rather than omitted.** M-4 replaced the seam call with a hard-coded
separator (`strings.ReplaceAll` on `"/"` → `"\\"`). All three tests PASSED. The guard pins the
converted VALUE under a pinned separator, not the USE of the `fromConfigPath` /
`configPathSeparator` seam. Closing that would need a second `'/'`-pinned identity run, outside this
card's scope.

**Gaps** (not observed, therefore not claimed): Windows resolution of the converted form; the
seam-usage axis (M-4); the home-relative branch; `golangci-lint`; per-test PASS lines inside the
non-`-v` full run; a `GOOS=windows` build. Detail in `run-evidence.md` § Gaps.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-08
run_commit_sha: pending-backfill
run_status: complete
ac_pass_count: 6          # AC-001,002,003,004,005,007 — AC-006 is no-AC-by-design (constraint-discharged)
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run   # the lane does not push; the develop push is the lead's batched act
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: none observed (go vet ./internal/cli/ exit 0; golangci-lint not run — see Gaps)
cross_platform_build:
  darwin: not-run           # no production code added, so no build claim is made either way
  windows: not-run
total_run_phase_files: 1 test file, plus SPEC and report artifacts
m1_to_mN_commit_strategy: single milestone M1 — one implementation commit plus one evidence commit
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-08
sync_commit_sha: pending-backfill-sync   # a commit cannot cite its own hash; backfilled in the following commit
sync_status: complete
b12_self_test_a: pass    # /usr/bin/grep -c 'SPEC-CODEX-DOCTOR-PATH-GUARD-001' CHANGELOG.md -> 0 before append (no duplicate entry)
b12_self_test_b: pass    # AC count cited (7) matches acceptance.md; see the discriminator note below
b12_self_test_c: pass    # every path claimed in the CHANGELOG entry verified to exist before commit
changelog_entry_position: "[Unreleased] > Added, first entry (newest-first, line 12)"
frontmatter_status_transitions:
  spec_md: in-progress -> completed    # merged 3-phase close, carried on this single sync commit
  plan_md: n/a                          # stateless on the status axis (no status: field)
  acceptance_md: n/a                    # stateless on the status axis (no status: field)
  progress_md: n/a                      # phase state lives in body sections, not frontmatter
canary_compliance_check: n/a            # this SPEC defines no forward-looking policy its own sync tests
```

**AC-count discriminator (recorded, because the raw counter over-reports).** The generic sweep
`/usr/bin/grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l` prints **9** on this
file. Two of those nine — `AC-CSRB-003` and `AC-CSRB-006` — are cross-references to the sibling
SPEC-CODEX-SKILL-PATH-READBACK-001 (card t562) cited in the "Contrast with t562" and edge-case
prose; they are not criteria of this SPEC. This SPEC's own criteria are the seven `AC-CDPG-*`
identifiers, and 7 is the figure the CHANGELOG entry cites. No occurrence in this file carries a
`[RETIRED]` or `[REF]` reserved token, so no identifier is excluded on adjacency grounds and none
is ambiguous.

Of the seven, six are verifiable criteria and all six are PASS (§E.3 `ac_pass_count: 6`);
`AC-CDPG-006` is deliberately no-AC-by-design — REQ-CDPG-006 is discharged as a `plan.md` §D-2
constraint plus a Definition-of-Done checkbox verified by reading the test source, because its
failure mode is a nondeterministic cross-test race that a passing run cannot establish.

**Sync-phase scope.** This commit touches `CHANGELOG.md`, `spec.md` frontmatter (`status:` only —
`updated:` was already today's date), and this section. No SPEC body content, no implementation
file, and no other SPEC directory was modified.
