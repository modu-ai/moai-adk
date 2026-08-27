# SPEC-INTEGRATION-LOCK-LIVENESS-001 — progress (card t298)

Plan-phase artifacts authored 2026-08-27 on tree `d29b8942e` @ `WT-integration-lock`
(worktree t298). Tier M. Status: draft.

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set complete for Tier M: spec.md, plan.md, acceptance.md, progress.md (this file).
- SPEC ID regex check — command run in the iter-1 repair pass, verbatim:
  `ID="SPEC-INTEGRATION-LOCK-LIVENESS-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL`
  → observed output `PASS`. The pattern is the multi-segment form (`^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`),
  which is what admits this ID's multi-hyphen domain. The single-segment form quoted in
  `spec-frontmatter-schema.md` § Field Reference (`^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`) rejects it;
  multi-segment IDs are an established repo-wide pattern and `moai spec lint` emits no ID finding
  for this SPEC, so the ID is not the defect — the earlier unattributed `PASS` claim was.
- Frontmatter: 12 canonical fields present; `status: draft` set at creation; `phase` carries a
  release target ("v3.1.4 target"), not a lifecycle stage.
- ID uniqueness: no SPEC under this worktree's `.moai/specs/` shares the ID.
- Requirements: REQ-INL-001..010 in GEARS notation (no IF/THEN modality).
- Out of Scope section present with four `### Out of Scope — <topic>` H3 sub-headings, each
  with `-` bullets.
- Acceptance: AC-INL-001..011, two-cell discipline (RED-now pinned to `d29b8942e` + green path
  naming its flipping milestone); M1 is the failing-test-first milestone.
- Premises line-pinned to `d29b8942e` (spec.md §B, 13 items) — re-verify line numbers before
  citing from any other tree.

### iter-1 audit repair (v0.1.1)

plan-auditor iter-1 returned FAIL 0.795 (Tier M threshold 0.80; report
`.moai/reports/t298/plan-audit-iter1.md`). Five blocking defects (D1-D5) plus four optional
(D6-D9) were repaired in this pass. Every claim below is measured in THIS run.

- **Baseline still valid despite HEAD advancing.** The artifacts pin `d29b8942e`; HEAD is now
  `c67a6ea64` (`git rev-parse --short HEAD`). `git merge-base --is-ancestor d29b8942e HEAD`
  → exit 0 (`d29b8942e` is an ancestor), and
  `git diff --stat d29b8942e HEAD -- internal/kanban/integration_lock.go internal/cli/integration.go internal/hook/integration_lock_guard.go internal/session/session_pid.go internal/session/registry.go CLAUDE.local.md .claude/rules/local/gitflow-lane-protocol.md`
  → **empty output**. Every §B premise file and both doc targets are byte-unchanged, so the
  RED-now cells pinned to `d29b8942e` still hold on the current tree.
- **D1** — `grep -n 'PIDSource\|pid_source' internal/kanban/integration_lock.go` → no match
  (rc=1). AC-INL-001's **Then** now asserts the written marker as clause (b).
- **D2** — `awk '/^#### §4\.1\.2/,/^#### §4\.1\.3/' CLAUDE.local.md` → a bash procedure block
  only; `grep -n '직렬화' CLAUDE.local.md` → no match;
  `grep -n 'integration acquire\|integration release' CLAUDE.local.md` → exactly 2 hits
  (lines 312, 317), both command lines inside that block. The earlier RED premise
  ("§4.1.2 instructs the lead-notice-only workaround") is FALSE; AC-INL-011 and spec.md §B item 13
  are restated against the measurement. Baseline for AC-INL-011's new greps:
  `awk '/^### §4\.1 /,/^## 5\./' CLAUDE.local.md | grep -c '세션 프로세스'` → **0**, same pipeline
  with `'재획득'` → **0**.
- **D3** — `grep -rn 'internal/kanban' internal/session/` → **1 hit**
  (`internal/session/session_pid.go:49`, a comment, not an import);
  `grep -rn '"github.com/modu-ai/moai-adk/internal/kanban"' internal/session/` → **0 hits**
  (rc=1). plan.md §C now uses the import-scoped form.
- **D4** — `grep -n 'Flock\|flock\|LockFile' internal/kanban/integration_lock.go internal/cli/integration.go`
  → exactly 5 lines: 2 `flock` prose lines in the package header (15, 19) and 3
  `IntegrationLockFileName` constant/use lines (38, 39, 106) matching only the `LockFile`
  substring. **No call site**, and `internal/cli/integration.go` contributes no match. `AcquireIntegrationLock` (integration_lock.go:146-181) reads at line 155 and
  writes at line 177 with nothing between; `writeIntegrationLock` stages to a fixed
  `path + ".tmp"` (line 229). Hazard recorded in spec.md §G; plan.md M5 carries a wording bound.
- **D5** — the regex claim above is now attributed to its exact command and pattern.
- **D6/D7/D8/D9 applied**: full REQ IDs in the §D matrix (`grep -c 'REQ-INL-008' acceptance.md`
  0 → **1**); REQ-INL-003 relabeled `(Event-driven)`; pid-reuse + absent-host hazards added to §G
  (`IntegrationLock` integration_lock.go:71-79 has no `Host` field; `session.Entry`
  registry.go:86-95 does); REQ-INL-006 reworded as a preserved invariant.

**Provenance of this repair pass is UNRESOLVED.** The directing actor for the iter-1 repair was
not identified. Three things are known and none of them names one: the orchestrator did not
instruct the repair; the audit report `.moai/reports/t298/plan-audit-iter1.md` contains no
repair record; and the only positive signal is a timeline plus a capability — the report landed
at 19:25:57 and the four artifacts were modified 19:33:09-19:34:36, and `plan-auditor`'s tool
list carries `Write` and `Edit`. Capability plus adjacency is not attribution, so this is
recorded as an open provenance gap rather than as a conclusion: the repair content itself was
re-measured in that pass and is accepted on its own evidence, independently of who directed it.

### v0.1.2 surgical amendment (plan phase, still `draft`)

Three additive changes; no existing REQ or AC renumbered, reworded, or removed.

- **AC-INL-012 added** (acceptance.md §D.1 + matrix row + MUST-PASS gate) — the acquire-refusal
  path. Gap it closes: AC-INL-001 gates the STATUS read and AC-INL-003 the takeover direction,
  so a mutant that anchors the owner pid for `Stale()` while leaving `acquire` willing to
  displace a live holder passes the whole prior set. **No new REQ was added**: REQ-INL-004 already
  names "acquire refusal" as a consumer leg, and REQ-INL-010 already forbids silently overwriting
  another lane's recorded trace. Measurements taken in this run on `c67a6ea64`:
  `grep -rln "exec.Command" internal/cli/integration_lock_cli_test.go internal/kanban/integration_lock_test.go`
  → no output (rc=1), i.e. no cross-process refusal coverage exists; the refusal branch is
  `internal/kanban/integration_lock.go:161-168` and is bypassed via `case current.Stale():`
  (line 163) because line 175 records the exiting CLI's own pid; the displaced-holder line is
  `internal/cli/integration.go:196-199`. The runtime leg is deliberately **measured-by-test**
  (M1) rather than by hand — `moai integration acquire` mutates the shared primary-checkout
  record and live lanes hold real merge windows in this repository right now.
- **spec.md §A extended** with the two 2026-08-27 production field observations (bare acquire
  displacing a live holder 42s after it was recorded; both lanes told the window was theirs,
  empty file intersection so no corruption) and with the identity-of-record vs atomicity layer
  distinction, so the card's scope is not read as delivering mutual exclusion.
- **plan.md M1** gains the corresponding RED test bullet; §H cross-reference range widened to
  AC-INL-012. Frontmatter `version` → `0.1.2` (spec.md is the only artifact carrying
  frontmatter); `status` remains `draft` — the Implementation Kickoff Approval gate has NOT been
  passed.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
