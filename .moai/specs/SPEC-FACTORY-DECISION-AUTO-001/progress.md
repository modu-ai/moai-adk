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
| M2 shared verdict predicate + doctrine (O9 disposed: spec-workflow hash-subject sentence, local + template) | `2bde7f65e` | done in code/rules; plan-auditor.md verdict-block fields BLOCKED (agent file, outside manager-develop scope) |
| M7 bind cache + degraded inbox | `b387d9ff8` | done |
| M3 audit decider | — | BLOCKED: SPEC defect D-RUN-1 below |
| M0, M4, M5, M6, M8 | — | not started (M5 and the agent-file parts of M4/M6 are agent edits outside manager-develop scope) |

### SPEC defect D-RUN-1 (stopped on)

REQ-FDA-014 requires "the card carries no open blocker and no operator hold", but the factory card
record (`internal/homestate/card_record.go` `type Card`) carries no blocker or hold field, and
design.md §5 names no source for either. Implementing T8a without a mechanical definition would
either refuse always or drop a keep-set check. Needs manager-spec: name the source (queue `hold`
state / `[보류` marker, progress wait records) before M3 starts.

### AC evidence (run so far)

| AC | Command | Verbatim key output | Status |
|---|---|---|---|
| AC-FDA-001/002/003 | `go test -count=1 -race -cover ./internal/decision/...` | `ok  github.com/modu-ai/moai-adk/internal/decision 2.019s coverage: 85.6% of statements` | PASS |
| AC-FDA-001 (CLI, lane refusal) | `go test -count=1 -run TestDecisionCmd ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli 1.603s` | PASS |
| AC-FDA-009 | `go test -count=1 -cover ./internal/auditverdict/...` ; `go test -run TestFDA_T7 ./internal/homestate/` | `ok ... auditverdict 0.272s coverage: 100.0%` ; `ok ... homestate 3.396s` (RED before wiring: `label-only PASS at T7: err=<nil>`) | PASS (contract site applies label rule only — contract records only the label) |
| AC-FDA-007 | `go test ./internal/template/` (batch-gate doctrine A30 re-anchored) | `ok  github.com/modu-ai/moai-adk/internal/template 215.578s` | PASS |
| AC-FDA-006 | — | plan-auditor.md not edited | BLOCKED |
| AC-FDA-020/021/022 | `go test -count=1 -race -run TestFDA_ ./internal/hook/` | `ok  github.com/modu-ai/moai-adk/internal/hook 16.998s` (RED: `cache hit opened the broker 1 time(s)`; `first degraded inbox claim was not surfaced: ""`) | PASS (M0(a) load measurement not run) |
| others | — | — | not started / blocked |

Lint: `golangci-lint run` (v2.1.6) on hook, auditverdict, decision, contract, homestate, runtime → `0 issues.`

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: partial — stopped on SPEC defect D-RUN-1
run_commit_sha: b387d9ff8
ac_pass_count: 8
ac_blocked_or_open: 17
new_warnings_or_lints_introduced: 0
```

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
