---
id: SPEC-FACTORY-LANE-JOIN-SOCKET-001
title: "Acceptance — factory lane join tolerates run-record absence"
version: "0.1.0"
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
---

# acceptance.md — SPEC-FACTORY-LANE-JOIN-SOCKET-001

## §A Conventions

Every criterion is Given-When-Then and binary-testable. "Verified leader" means the discovery
probe accepted the candidate under REQ-002/003 (live by pid + process-start fingerprint; leader
label; project membership; readable run identity). Tests run under the REQ-012 isolation
discipline (isolated `HOME`/`MOAI_HOME`, project directory outside the repository worktree set,
stubbed launched binary). RED-now cells were measured on this worktree at `68e37864a` (document
pin; criteria carrying no pin of their own bind to it).

## §B AC Matrix

| AC | Subject | Requirement | Verification verb | Release-blocking |
|---|---|---|---|---|
| AC-001 | Defect contrast: live leader + zero active runs joins | REQ-001/003 | executed test | yes |
| AC-002 | Zero leader refusal stands | REQ-001/006 | executed test | yes |
| AC-003 | Explicit run id skips discovery | REQ-006 | executed test | yes |
| AC-004 | Multi-leader ambiguity fails closed | REQ-005 | executed test | yes |
| AC-005 | Socket existence is not liveness | REQ-002 | executed test (mutant) | yes |
| AC-006 | Absent record created on resume | REQ-004 | executed test | yes |
| AC-007 | Retired record resumed with leader owner stamp | REQ-004/007 | executed test (mutant) | yes |
| AC-008 | Resume event auditable | REQ-004 | executed test | yes |
| AC-009 | --lead default and targeting | REQ-008 | executed test | yes |
| AC-010 | Legacy spelling refused | REQ-008 | executed test | yes |
| AC-011 | --lead with no match refuses | REQ-008/001 | executed test | yes |
| AC-012 | Lane env + hook bind into resumed run | REQ-009 | executed test | yes |
| AC-013 | Mirror parity across the join point | REQ-010 | executed test + source assert | yes |
| AC-014 | Parse invariants preserved | REQ-011 | executed test (existing) | yes |
| AC-015 | Lane auto-assignment unchanged | REQ-011 | executed test (existing) | yes |
| AC-016 | Docs and help updated | REQ-008/001 | file check | yes |
| AC-017 | Non-member and unreadable-identity candidates declined | REQ-003 | executed test (two legs) | yes |

## §C RED-now evidence ledger (tree pin 68e37864a)

Measured during plan phase on this worktree. The new-behavior criteria (AC-001..AC-013,
AC-016, AC-017) are red today for one stated reason: **the discovery surface does not exist** — the
launcher's lane join consults no live-leader source whatsoever, so the refusal is unconditional.
Three observations establish that jointly:

| Cell | Command | Verbatim output | Exit code |
|---|---|---|---|
| R-1 (refusal observed) | `MOAI_HOME=/tmp/t1330-red-home /tmp/t1330-moai cc -f lane` (fresh fixture project, zero active rows) | stderr: `No_active_factory.` (banner-emitted) | 1 |
| R-2 (law + positive control green) | `go test ./internal/cli -run '^(TestFactoryRunSelectionAtomicSlotsAndArgv\|TestGLM_FactoryLeadRunIsJoinableByLane)$' -count=1` | `ok  github.com/modu-ai/moai-adk/internal/cli 3.531s` | 0 |
| R-3 (absence with positive control) | `grep -rniE 'discover.*leader\|leader.*discover\|findLiveLeader' internal/cli internal/homestate internal/kanban --include='*.go' \| grep -v _test.go \| wc -l` (control: same shape for `enterSelectedFactoryRun` → 6) | `0` | 0 |

R-1 shows the behavior AC-001 requires is absent (refusal despite live leaders on the machine —
the live `tm3yoq` leader, pid verified via `lsof` in plan research). R-3 shows it is absent
because nothing leader-aware exists to run, not because a fixture failed — the red is for the
right reason. AC-014/AC-015 pin **existing** green tests; their RED-now is not applicable (they
are regression guards over current behavior, adopted as-is).

## §D Acceptance criteria

### AC-001 — Defect contrast (the t1330 defect shape)

- **Given** a fixture project whose factory state holds zero `active` runs, and a live process
  the probe can verify as this project's leader (stub binary; leader label in argv; this
  project's canonical cwd; its own `MOAI_KANBAN_ID` env),
- **When** the operator runs the lane join (`moai cc -f lane`, and the same fixture through
  `moai glm -f lane`),
- **Then** the join succeeds, the lane's run identity is the leader's own run id, and the run
  row is active with the leader's verified identity as owner stamp.
- Green path: M3 (join-point integration lands). RED-now: R-1/R-2/R-3.

### AC-002 — Zero leader refusal stands

- **Given** zero `active` runs and no candidate that passes the REQ-002/003 proof,
- **When** the lane join runs,
- **Then** it fails with `NO_ACTIVE_FACTORY` and no run row is created or modified.

### AC-003 — Explicit run id skips discovery

- **Given** zero `active` runs and a live verifiable leader, and the operator passes
  `--factory-run <id-of-a-non-active-run>`,
- **When** the lane join runs,
- **Then** it fails with `NO_ACTIVE_FACTORY`, discovery is not attempted, and the live leader's
  run is not resumed.

### AC-004 — Multi-leader ambiguity fails closed

- **Given** zero `active` runs and **two** live verifiable leaders (distinct run ids),
- **When** the lane join runs,
- **Then** it fails closed naming both candidates with their run ids, no run row is mutated,
  and neither leader's run is resumed.

### AC-005 — Socket existence is not liveness (mutant)

- **Given** a candidate whose socket file (or any other record) exists but whose pid probes dead
  (or whose process-start fingerprint mismatches),
- **When** the lane join runs with zero `active` runs,
- **Then** the candidate is declined, the join fails with `NO_ACTIVE_FACTORY`, and no resume
  occurs. Mutation probe: deleting the liveness check (accepting on record presence alone) must
  turn this criterion red.

### AC-006 — Absent record created on resume

- **Given** a live verifiable leader whose run id has **no** row in `runs`,
- **When** the lane join runs,
- **Then** the row is created `active` with the verified leader's pid + process-start as owner
  stamp, the join succeeds, and the lane binds the leader's run broker.

### AC-007 — Retired record resumed with leader owner stamp (mutant)

- **Given** a run row `retired` whose stamped owner is dead, and a live verifiable leader for
  that same run id,
- **When** the lane join runs,
- **Then** the row is `active`, its owner stamp is the **verified leader's** pid +
  process-start (not the joining launcher's), and the join succeeds. Mutation probe: resuming
  via `recordFactoryRunStart` (stamping the lane) must turn this criterion red — observable by
  classifying the run after the lane process exits: a lane-stamped run classifies dead; a
  leader-stamped one stays live.

### AC-008 — Resume event auditable

- **Given** a resume performed under AC-006 or AC-007,
- **When** the run's events are read,
- **Then** a resume event exists recording the run id and the verification basis (which probe
  facts produced the restore), distinguishable from `run.started`.

### AC-009 — --lead default and targeting

- **Given** a live verifiable leader,
- **When** the lane join runs with no `--lead` flag, and again with `--lead leader`,
- **Then** both target the canonical leader label (`kanban.LeaderLabel()`) and both succeed
  identically; and **when** a second live leader runs under a different label and the operator
  passes `--lead <that-label>`, discovery targets that leader's run.

### AC-010 — Legacy spelling refused

- **Given** any lane join,
- **When** the operator passes `--lead lead` (or any `lead-<suffix>` form),
- **Then** the launch is refused with the canonical-form error (the `leader` noun), nothing is
  parsed further, nothing launches, and nothing is written — the same refusal shape
  `refuseLegacyEntryNames` applies to `--name`.

### AC-011 — --lead with no matching live leader refuses

- **Given** zero `active` runs and no live leader carrying the requested label,
- **When** the lane join runs with `--lead <unmatched-label>`,
- **Then** it fails with `NO_ACTIVE_FACTORY`.

### AC-012 — Lane env + hook bind into the resumed run

- **Given** a join performed through discovery (AC-001 shape),
- **When** the child session's env and its SessionStart hook bind are examined,
- **Then** the env carries `MOAI_KANBAN_ID` = the resumed run id and `MOAI_KANBAN_LEAD_NAME` =
  the target leader's name, and the hook binds the lane peer into that run's broker at the bind
  an ordinary join produces (pending+1 — launch-pending 1 → bind 2, the measured chain; the
  literal `Generation: 1` is the want-shape at `factory_messages.go:96`, not the measured bind;
  evidence in progress.md §E.2) — the run identity and broker outcome of an ordinary join, with
  the leader name added by this SPEC on the discovery path (ordinary joins export no leader
  name and are unchanged).

### AC-013 — Mirror parity across the join point

- **Given** the cc, glm, and codex lane entry points,
- **When** the AC-001 / AC-002 matrix is executed through each,
- **Then** all three exhibit identical join behavior; and a source-level assertion confirms the
  discovery path lives in the shared join point with no launcher-private copy.

### AC-014 — Parse invariants preserved (existing guards stay green)

- **Given** the existing entry-parse regression set,
- **When** the full lane-entry truth table is exercised (`-k` + `-f` error; `-f lane` +
  operator `--name` conflict; `--name lane-2` alone does not join; codex-leader adoption
  refusal),
- **Then** every existing assertion holds unchanged after the change.

### AC-015 — Lane auto-assignment unchanged (existing guards stay green)

- **Given** the existing lane-numbering regression set,
- **When** auto-assignment is exercised (dead-claim pruning, highest-live + 1, bump rule,
  project-scoped registry shared across runs),
- **Then** every existing assertion holds unchanged; numbering never resets per run.

### AC-016 — Docs and help updated

- **Given** the operator-facing surface (`-f lane` tolerance + `-l/--lead`),
- **When** the docs inventory is checked,
- **Then** the cc and glm help text (the `-f lane` / `-f lane-<n>` flag block), the 4-locale
  docs-site pages (`factory-mode.md`, `kanban-mode.md`, `cli-reference/launchers.md`), and the
  4 README occurrences describe the record-absence tolerance and the new flag, all locales in
  parity.

### AC-017 — Decline branches of the identity proof fail closed (REQ-003)

- **Given** zero `active` runs and a candidate that is live and carries the targeted leader
  label but fails one of the remaining REQ-003 facts — leg (a): the candidate's cwd
  canonicalizes to a **different** project key, so it is not this project's leader; leg (b):
  the candidate's `MOAI_KANBAN_ID` env **cannot be read** (platform or privilege),
- **When** the lane join runs (one fixture per leg),
- **Then** in both legs the candidate is declined, the join fails with `NO_ACTIVE_FACTORY`,
  and no run row is created or modified. Mutation probe: removing the membership check (leg a)
  or the readable-identity requirement (leg b) must turn the respective leg red — a mutant
  that joins across projects or resumes an unverified run id fails here.

## §E Edge cases

- Leader alive but its run row is `active` already → ordinary gate path, no discovery (discovery
  fires only on the refusal branch).
- Candidate leader belongs to a **different** canonical project (cwd or run-id env evidence) →
  declined; zero-leader refusal stands (AC-017 leg a).
- Candidate matches the label but is a lane (`lane-<n>`), not a leader → declined (label
  evidence is leader-specific).
- Run row exists `active` but its leader is dead → REQ-007: no discovery involvement (the gate
  passed on the record); the existing reconciliation machinery owns that case, unchanged.
- Discovery finds a live leader whose run id row is `retired` **and** another live leader whose
  row is absent → two verified candidates → AC-004 fail-closed.
- Unreadable candidate env (platform/privilege) → candidate declined (REQ-003); refusal stands
  (AC-017 leg b).
- Operator passes `--lead` on a leader entry, or `--lead` with `--factory-run` → parse error
  (REQ-008).

## §F Quality gates

- TRUST 5 per `.moai/config/sections/quality.yaml`; cycle_type=tdd (RED-GREEN-REFACTOR).
- Targeted packages: `go test ./internal/cli/... ./internal/homestate/... ./internal/kanban/...`
  (full suite is CI's job — lane-local scope discipline).
- `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` green.
- `golangci-lint run` with no NEW findings vs the pre-change baseline.
- spec-lint on this SPEC directory: no error findings.

## §G Definition of Done

1. All release-blocking ACs PASS with evidence attributed per the VCI §3 five-section format
   (command + verbatim output + tree SHA).
2. The regression pair (`TestGLM_FactoryLeadRunIsJoinableByLane`,
   `TestFactoryRunSelectionAtomicSlotsAndArgv`) plus the new discovery/resume tests green.
3. No run-resolution, retirement, or numbering semantics changed (AC-002/003/004/014/015
   collectively pin this).
4. Documentation parity landed (AC-016).
