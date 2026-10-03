# progress.md — SPEC-FACTORY-DECISION-AUTO-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-FACTORY-DECISION-AUTO-001
card: t1481
tier: L
branch: WT-decision-automation
base: d7112d005
probe_tree: 5d094991fb586b9c4aadee33b663fe92b686ff18
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md, decision-index.md]
spec_version: 0.4.0
req_count: 25
ac_count: 25 (24 release-blocking, 1 RG)
plan_audit: iter1 FAIL 0.74 (.moai/reports/t1481/plan-audit-iter1.md); iter2 FAIL 0.80 (.moai/reports/t1481/plan-audit-iter2.md); revision 0.3.0 addresses N1-N7 + O1-O4; iter3 FAIL 0.83 (.moai/reports/t1481/plan-audit-iter3.md, Tier L ceiling, no regression); leader-ruled one delta round; revision 0.4.0 addresses N8-N11 + O6 within the iter3 fix_scope; iter4 delta PASS 0.885, no blockers (.moai/reports/t1481/plan-audit-iter4.md, audited_sha a13b83868)
plan_artifacts_frozen_at: a13b83868 (no spec/plan/acceptance/design/research/decision-index edit after the PASS, so the audited state and hash stay bound)
recorded_debts:
  - O9 (dispose_in: run M2): plan.md M2 row does not name the spec-workflow.md hash-subject sentence edit (local + template); run M2 follows design.md §5, which assigns it
  - O10 (dispose_in: run M3): the REQ-FR-019 Amendments obligation of REQ-FDA-016 is asserted only through AC-FDA-015; run M3 evidence names REQ-FDA-016 beside AC-FDA-015
open_decisions: none — Q1-Q26 LEADER-DECIDED 2026-10-03 (mission contract 07d28c4b)
evidence_needed_in_run: M0(a) degraded-notice rate with/without bind cache (Q5); M0(b) recheck cache cost (Q4)
release_target: v3.2.0
audit_ready: true
```

## §E.2 Run-phase Evidence

### Kickoff decision record (autonomous form, auto-semantics §9.1)

decision record: decided_by=claude+leader evidence_refs=.moai/reports/t1481/plan-audit-iter4.md;verdict=PASS;score=0.885;audited_sha=a13b83868;plan_artifacts_frozen_at=a13b83868 ladder_path=gate-row plan→run Kickoff (AUTONOMOUS, auto-semantics §9.1; leader decision under mission contract 07d28c4b)

- Gaps: no codex audit receipt issued (`audit.gates.codex` not required in this tree).
- Run agent: manager-develop role taken over by the plan session in this tree (the separately spawned run agent could not shell into the tree — worktree guard), cycle_type=tdd.

### Sibling-SPEC dependency check

`grep -rn "MERGE-WINDOW-QUEUE\|FACTORY-QUEUE-RECORD\|t1479\|t1480" .moai/specs/SPEC-FACTORY-DECISION-AUTO-001/*.md` →
mentions only (research §1 evidence, decision-index Q16, spec §F Out of Scope for t1479); no
design or acceptance criterion consumes code from SPEC-MERGE-WINDOW-QUEUE-001 (t1479) or
SPEC-FACTORY-QUEUE-RECORD-001 (t1480). No milestone is blocked on them. Residual: t1480 also edits
the SPEC-FACTORY-RECORD-001 state machine (`internal/homestate/card_transition.go`), so M3's T8a
edge is a likely textual merge conflict at integration, not a code dependency.

### Pre-flight baselines (plan §C) — tree cb8b7e03a

| Command | Verbatim result |
|---|---|
| `go test -count=1 -timeout 30m ./internal/homestate/...` | `ok  	github.com/modu-ai/moai-adk/internal/homestate	95.547s` / `exit=0` |
| `go test -count=1 -timeout 30m ./internal/contract/... ./internal/hook/... ./internal/runtime/...` | contract/runtime all `ok`; `FAIL	github.com/modu-ai/moai-adk/internal/hook	730.094s` (13 failures, lane env leaking) |
| same `./internal/hook/` with lane env scrubbed (`unset MOAI_* CLAUDE_CODE_* && go test`) | `--- FAIL: TestStaleRunNoticeFactoryLegacyLabel (1.78s)` / `exit=1` — pre-existing at base |

### Milestone commits

| M | Commit | Status |
|---|---|---|
| M1 board + CLI | `fe9bcd7a6` | done |
| M2 shared verdict predicate + doctrine (debt O9 disposed) | `2bde7f65e` | done |
| M7 bind cache + degraded inbox | `b387d9ff8` | done |
| D-RUN-1 clarification (manager-spec role, leader ruling, mission contract 11c79e1a) | `818569869` | done — decision-index Q27 |
| M4/M5 agent + config edits (plan-auditor/manager-spec/sync-auditor C1+C2, C3 via `make agents-emit`; ceiling policy config) | `27c41ffb8` | done |
| M3 audit decider T8a + Amendments (debt O10: REQ-FDA-016's REQ-FR-019 amendment is evidenced in this commit beside AC-FDA-015) | `90d085183` | done |
| M6 doctrine (watchdog, waits, recheck, MCP fallback, 4dim binding conditions) | `19dd5d9ce` | done |
| M0 measurements | — | NOT RUN (gap) |
| M8 mirror/neutrality sweep | this commit | done (no new SPEC/card/date token in added template lines) |

### SPEC defect D-RUN-1 — resolved

REQ-FDA-014's "open blocker / operator hold" had no source on the factory card record. Leader
ruling (mission contract 11c79e1a): blocker = card record state `blocked`/`needs-decision`; hold =
queue item state `hold`; unreadable source fails closed. Recorded as Q27 (`818569869`).

### AC matrix

| AC | Command | Verbatim key output | Status |
|---|---|---|---|
| 001/002/003 | `go test -count=1 -race -cover ./internal/decision/...` | `ok ... internal/decision 2.019s coverage: 85.6% of statements` | PASS |
| 001 CLI lane refusal | `go test -run TestDecisionCmd ./internal/cli/` | `ok ... internal/cli 1.603s` | PASS |
| 004 | `grep -c 'moai decision read'` watchdog SKILL (local+template) | 1 / 1 (was 0, P4) | PASS (doctrine) |
| 005 | manager-spec C1/C2 carry `board:<record-id>` pin rule | `grep -c` = 3 per copy | PASS (doctrine; no fixture test) |
| 006 | plan-auditor C1/C2 verdict fields | `blocking_count`, `must_pass_failed`, `fix_scope`, `defect_class`, `reread_hunks` present (P6/P11/P13 were exit=1) | PASS (doctrine) |
| 007 | `go test ./internal/template/` (A30 re-anchored) | `ok ... internal/template 215.578s` | PASS |
| 008 | sync-auditor C1/C2 section + `sync-audit-4dim.js` undisposed→FAIL | doctrine + script edit; no executed fixture | PARTIAL (4dim script not executed — gap) |
| 009 | `go test -cover ./internal/auditverdict/...`; `-run TestFDA_T7 ./internal/homestate/` | `coverage: 100.0%`; `ok ... 3.396s` (RED `label-only PASS at T7: err=<nil>`) | PASS |
| 010 | harness.yaml `plan_audit_ceiling_policy` both trees; `go test ./internal/config/` | `ok ... internal/config 6.429s` | PASS |
| 011/012/013 | ceiling procedure text in plan-auditor + spec-workflow | doctrine only; no executable fixture (procedure is agent-driven) | PARTIAL |
| 014 | `go test -run TestFDA_ ./internal/homestate/` + CLI hold test | `ok ... internal/homestate 19.440s`; `ok ... internal/cli 8.332s` | PASS (blocked/needs-decision arms: edge starts at kickoff, so such cards are refused as illegal transitions — no dedicated fixture) |
| 015 | `go test ./internal/homestate/...` (AC-005 now 66/295) | `ok ... internal/homestate 125.189s` | PASS |
| 016 | `go test -run 'TestFDA_|TestSD_AC016' ./internal/cli/` + Amendments rows | `ok ... internal/cli 8.332s` | PASS |
| 017/018 | manager-spec C1/C2 Class/Default/product-level/DEFAULT-APPLIED at plan close | doctrine; decider fixtures in TestFDA_AuditDeciderFounderRows | PASS (doctrine + decider side) |
| 019 | auto-semantics §14 + watchdog | doctrine; `workflow.watchdog.wait_recheck_minutes` key NOT added to workflow.yaml (operator note: do not commit workflow.yaml) | PARTIAL |
| 020/021/022 | `go test -race -run TestFDA_ ./internal/hook/` | `ok ... internal/hook 16.998s` (RED `cache hit opened the broker 1 time(s)`) | PASS |
| 023 | kanban-dispatch-detail `fallback=CLI` | doctrine (P23 was exit=1) | PASS (doctrine) |
| 024 | push still human: CLI test case "push" with `--decider audit` refused | `ok ... internal/cli 8.332s` | PASS (RG) |
| 025 | `make agents-emit`; `go test ./internal/template/agentemit/...` | `ok ... agentemit 0.458s` | PASS |

Lint: `golangci-lint run` (v2.1.6) on cli, homestate, hook, auditverdict, decision, contract, runtime → `0 issues.`

Pre-existing, not caused here: `TestStaleRunNoticeFactoryLegacyLabel` (hook, fails at base with env scrubbed); `TestDetectDefaultBranch` (internal/workflow, untouched package).

### Sync-audit FAIL 72 repair (`.moai/reports/t1481/sync-audit.md`)

| Finding | Fix | RED (before) | GREEN (after) |
|---|---|---|---|
| F1 NaN/±Inf/out-of-range score passed the threshold | `auditverdict.Parse` accepts only finite scores in [0,1] | `score "+Inf" admitted`, `score "1.5" admitted`; decider: `NaN score: audit approval accepted, want refusal` | `ok …auditverdict 1.283s` (-race) |
| F2 DEFAULT-APPLIED restriction skipped for non-FOUNDER labels | restriction runs before the label branch | `DECIDED row holding DEFAULT-APPLIED on product-level: audit approval accepted, want refusal` | `ok …homestate 109.381s` (-race, full package) |
| F3 any non-empty §E.1 counted as audit-ready | requires the explicit `audit_ready: true` line | `audit_ready false: audit approval accepted, want refusal` | same run |
| F4 `decision read` hid fields | prints predicate/supersedes/resolves/release/cards | — (display only) | `ok …cli 15.091s` (-race, TestFDA_/TestDecisionCmd/TestFR_AC015/TestSD_AC016) |

golangci-lint v2.1.6 on auditverdict, homestate, cli: `0 issues.` This SPEC's §E.1 now carries
`audit_ready: true`.

### Sync re-audit FAIL 78 repair (`.moai/reports/t1481/sync-audit-2.md`)

| Finding | Fix | RED | GREEN |
|---|---|---|---|
| N1 duplicated decision key last-wins | `auditverdict.Parse` records duplicates of verdict/score/must_pass_failed/blocking_count/plan_artifact_hash; `Admit` refuses any | `second must_pass_failed: admitted ()`, `second blocking_count: admitted ()`, `second score: admitted ()`, `second plan_artifact_hash: admitted ()`, `sync phase: duplicated verdict admitted` | `ok …auditverdict 1.119s` (-race) |
| N2 queue hold read before the transition | `TransitionRequest.QueueHoldRead` is called inside the transaction right before commit; CLI passes the reader | build failed: `unknown field QueueHoldRead in struct literal of type TransitionRequest` | `ok …homestate 83.153s` (-race, full package) |
| N3 conflicting `audit_ready` values | any non-true `audit_ready` line makes it not ready | same RED run (test compiled only after the fix) | same run |
| stale comments | `auditReadyRecorded`, `founderRowRefusal`, CLI queue-read comment updated | — | — |

CLI selectors (`TestFDA_|TestDecisionCmd|TestFR_AC015|TestSD_AC016`, -race): `ok …cli 12.983s`.
golangci-lint v2.1.6: `0 issues.`

### Recorded debts (sync re-audit)

- D1: no writer emits `audit_ready: true` (0 of 40 recent SPECs; the manager-spec template does not
  emit it), so the audit decider refuses every other SPEC (fail-closed, inert) until the plan-close
  step writes the signal.
- N4: board rows missing required fields still take part in `supersedes` handling and can hide a
  standing record (`internal/decision/board.go`).
- C2: plan artifacts were edited after the audited SHA a13b83868 (Q27, commit 818569869) while
  §E.1 still claims `plan_artifacts_frozen_at`; needs a delta plan re-audit or a record correction.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete-with-gaps
run_commit_sha: 19dd5d9ce
ac_pass_count: 21
ac_partial: [AC-FDA-008, AC-FDA-011, AC-FDA-012, AC-FDA-013, AC-FDA-019]
gaps: [M0 measurements not run, 4dim script not executed, wait_recheck_minutes config key absent, ceiling procedure untested mechanically]
new_warnings_or_lints_introduced: 0
```

### Residual-risk (leader decision, mission contract 11c79e1a)

- **Landing order: this card lands LAST among the v3.2.0 cards.** After it lands, any plan-audit
  verdict lacking `must_pass_failed`, `blocking_count`, or `plan_artifact_hash` is refused at
  Kickoff and at T7. Cards in flight with an older verdict must be re-audited; landing earlier would
  stall them.
- Any SPEC carrying `decision-index.md` takes one skip-cache miss (hash input widened).
- Expected textual conflict with t1480 in `internal/homestate/card_transition.go` (T8a edge).

### Recorded debts (leader decision; no new work in this card)

- AC-FDA-008, 011, 012, 013, 019 PARTIAL (see matrix); M0 measurements not run.
- `workflow.watchdog.wait_recheck_minutes` key → follow-up card (workflow.yaml is not committed in
  card trees right now).

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-03
sync_commit_sha: 4293979c7
sync_status: re-close pending — spec.md reverted to in-progress until a sync audit passes
sync_audit: FAIL 72 (.moai/reports/t1481/sync-audit.md) → F1-F4 repaired; re-audit FAIL 78 (.moai/reports/t1481/sync-audit-2.md) → N1-N3 repaired; next re-audit pending
changelog_entry_position: "CHANGELOG.md [Unreleased] ### Added (first entry)"
frontmatter_status_transitions:
  spec.md: in-progress -> completed (merged implemented+completed, single sync commit)
amended_specs:
  SPEC-FACTORY-RECORD-001: stays completed — Amendments row only; the amended behavior (T8a, AC-005 66/295) is implemented and tested in this card (homestate package ok)
  SPEC-FACTORY-SELF-DISPATCH-001: stays completed — Amendments row only; the REQ-SD-016 exception is tested here (TestFDA_LaneAuditDecideAdmission, TestSD_AC016 ok)
b12_self_test:
  pre_emission_grep: "grep -c SPEC-FACTORY-DECISION-AUTO-001 CHANGELOG.md -> 0 before emission"
  ac_count_match: "entry states 25 ACs; acceptance.md matrix rows = 25"
  file_paths_verified: ".moai/specs/SPEC-FACTORY-DECISION-AUTO-001/spec.md exists"
```
