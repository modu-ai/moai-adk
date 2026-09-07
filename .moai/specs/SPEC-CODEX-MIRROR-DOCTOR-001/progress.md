# Progress — SPEC-CODEX-MIRROR-DOCTOR-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t498 · worktree `.claude/worktrees/t498` · branch `WT-codex-mirror-doctor` · base `ace1c5440`
- Tier: M (3-artifact set: spec.md + plan.md + acceptance.md) · plan-auditor PASS threshold 0.80
- Requirements: 11 (ceiling 16) · Acceptance criteria: 15 (ceiling 16)
- SPEC ID regex check executed as Bash: `PASS` (re-executed at iteration 2)
- Authoritative input: `.moai/reports/t498/root-cause.md` (cited, not re-derived)
- Status: `draft` — awaiting plan audit and Implementation Kickoff Approval

### Repair iteration 2 (spec.md 0.2.0 · acceptance.md 0.2.0 · plan.md unchanged)

Closes the six blocking findings of `.moai/reports/t498/plan-audit.md` (iteration 1, FAIL 0.775):

| Finding | Where closed |
|---|---|
| D1 — AC-CMD-008 fails as written (chmod 0o000 breaks `t.TempDir()` cleanup) | acceptance.md AC-CMD-008 rewritten onto the file's symlink-loop idiom (`doctor_codex_test.go:335-352`) with its three fixture obligations |
| D2 — nine ACs decidable by an empty sweep | acceptance.md gains a file-level **swept-count gate** section plus a per-AC clause; AC-CMD-001 gains RED-now + green-path cells pinned to tree `dcb3ba0c7` |
| D5 — REQ-CMD-009 tail-drop clause unjudged | acceptance.md AC-CMD-013 (new), including the lead-summary exception sub-case |
| D6 — REQ-CMD-001/011 uncovered, no AC states its REQ | acceptance.md AC-CMD-014 / AC-CMD-015 (new) + a `REQ:` line under all 15 AC headings; AC-CMD-001 states it **supplements** `TestCheckCodexWiring_ClaudeOnlyMachineStaysSilent` |
| D4 — REQ-CMD-006/007 wiring precondition | spec.md §2 — both requirements gain `Where the project declares Codex wiring`, matching plan.md §D M2; AC-CMD-007 gains sub-case (b) to judge it |
| D3 — spec.md §6 asserts an unmeasured condition | spec.md §6 cross-platform clause re-stated conditionally, matching §3 row 3 |

D7-D10 (optional) are deliberately not acted on, per the auditor's own recommendation. D8 (path-literal
drift between the doctor and the producer's unexported `mirrorSkillsRelDir`) is recorded as a
follow-up card candidate, not fixed here.

Mechanical re-verification at iteration 2 (tree `dcb3ba0c7`):

```
$ ~/go/bin/moai spec lint .moai/specs/SPEC-CODEX-MIRROR-DOCTOR-001/spec.md
✓ No findings — all SPEC documents are valid                                    EXIT=0
$ grep -c '^## AC-CMD-' acceptance.md → 15   $ grep -c '^REQ: ' acceptance.md → 15
$ REQ: lines cite REQ-CMD-001…011 → all 11 covered, none orphaned
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
