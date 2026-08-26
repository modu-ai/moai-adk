# progress.md — SPEC-LEAD-DEBOTTLENECK-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_id: SPEC-LEAD-DEBOTTLENECK-001
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
plan_complete_at: 2026-08-26
plan_commit_ref: pending-backfill
red_now_baselines: measured-2026-08-26-t283-worktree-175d63f3f
open_clarifications: []
resolved_clarifications:
  - "매트릭스 확정 — RESOLVED 2026-08-26: spec.md §4 매트릭스를 리드 판정(a) 확정 후 Implementation Kickoff Approval에서 운영자 비준(매트릭스 비준 + run 승인 + 자율 진행 모드). 근거: plan.md §B.5 해소 기록"
notes: >-
  Tier M (surface: agent 1 + rules 2 + template mirrors; matrix/verification design fits
  spec.md+plan.md — no design.md needed). RED-now baselines recorded in acceptance.md §D.0
  (tools grep 0, deputy grep 0 x3, depth-seal ok 3.305s, mirror identical).
```

## §F Phase 4 Mode Selection

**M1-M3 (deputy charter + doctrine extension + mechanical verification) — Phase 4 decision (logged before the first run-phase manager-develop spawn):**

Input parameters:
- tier: M (13 AC, 12 REQ — within the 16/16 Tier M ceilings)
- scope (file count): ~6 tracked (EDIT `manager-lead.md` + mirror, EDIT `kanban-dispatch.md` + mirror, EDIT `kanban-dispatch-detail.md` + mirror) + SPEC artifacts
- domain count: 1 (lead-coordination doctrine — agent definition + rules + template mirrors, one cohesive mechanism)
- file language mix: markdown only; NO Go source changes (PRESERVE `internal/cli/**`, REQ-LDB-011)
- concurrency benefit: LOW (M1→M2→M3 hard sequence: M2's doctrine text codifies M1's charter, M3 verifies both)

Mode evaluation:
- Mode `direct`: NO — multi-file doctrine authoring with template-mirror + neutrality obligations, not a typo fix
- Mode `fanout`: NO — single domain, sequential milestone dependency (Anthropic coding-task parallelism caveat)
- Mode `sweep`: NO — 6 files, semantic authoring (not a uniform mechanical transform)
- Mode `agent-team`: NO — not operator-requested (kickoff selected autonomous GOAL progression, not the teams layer)
- Mode `serial`: YES

Decision: serial
Justification: M1-M3 form a strict dependency chain (charter → doctrine → verification) over markdown-heavy single-domain surfaces with template-mirror discipline; sequential manager-develop delegations (one per milestone) are the correct default per the coding-task parallelism caveat. Implementation Kickoff Approval passed 2026-08-26 (승인 / 매트릭스 비준 — the §4 matrix ratified by lead 판정 (a) + operator kickoff, marker resolved at `b20399b30` / 자율 goal arm). Progression mode: autonomous — `/moai goal` armed with the mechanical convergence condition (spec completed ∧ run PASS ∧ sync_commit_sha backfilled), turn ceiling 40, duration 4h.

Plan-audit gate note: iter-2 verdict PASS 0.923 ≥ Tier M threshold 0.80 on artifacts @ `ea27a72b9`; the marker-resolution commit `b20399b30` touched plan.md/progress.md only — acceptance.md (a ComputeHash subject) is unchanged since the verdict, so skip-eligibility holds.

## §E.2 Run-phase Evidence

### M1 — deputy charter (agent definition layer)

RED→GREEN flips vs acceptance.md §D.0 (measured on this tree, t283 worktree, base `b20399b30`):

- AC-001: `grep -c 'SendMessage' .claude/agents/moai/manager-lead.md` → `4` (RED-now was 0); `grep -c 'ListAgents'` → `2`; tools line remains a single CSV string (`tools: Read, Write, Edit, Bash, Grep, Glob, Agent, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill, mcp__moai__session_list, mcp__moai__goal_status, SendMessage, ListAgents`)
- AC-002: `grep -c 'Deputy dispatch surface' .claude/agents/moai/manager-lead.md` → `1`; `grep -ci 'deputy'` → `12` (RED-now was 0). Section codifies the 5 delegable duties (dispatch send + delivery-shape verification / bounded CI-watch polls / CodeRabbit two-condition read-report / first-pass evidence + `RECOMMEND:` / summary reporting) and the 6 retained powers under the marker.
- AC-003: `grep -c 'DEPUTY-RETAINED-BY-LEAD' .claude/agents/moai/manager-lead.md` → `2` (≥1); all 6 retained items enumerated under the marker (merge approval `LEAD-MERGE-APPROVED` / `FINAL VERDICT:` forbidden / operator gates / `moai todo` mutations / CodeRabbit adjudication / dispute coordination).
- AC-005: `grep -c 'routing'` → `2`; `name [ref]` re-send protocol present at line 221.
- AC-006 (invariant): `go test ./internal/template/ -run 'TestManagerLeadIsSoleAgentCarrier|TestManagerLeadCarriesAgent|TestNoNestedLeafWorkerCarrier' -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 0.550s` (baseline `3.305s`, same green). Sole-carrier grep: only manager-lead.md's tools line matches the `Agent` token.
- AC-007 (partial): mirror diff = exactly 1 line — the SPEC-TEAMMATE-REVIVAL provenance line (local-only, permitted modulo per AC-007(a)); `make build` rc=0; catalog.yaml regen = manager-lead hash only (`ccb8a1b2…` → `eceaee8c…`, 1 insertion 1 deletion); neutrality `grep -rc 'SPEC-LEAD-DEBOTTLENECK\|REQ-LDB\|SPEC-TEAMMATE-REVIVAL' internal/template/templates/` → 0 hits.
- Full affected package: `go test ./internal/template/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 199.174s` (includes TestTemplateNoInternalContentLeak neutrality guard + tool-catalog audit).
- AC-011 (M1 slice): working-tree diff touches no `internal/`, `pkg/`, `cmd/` Go source — only `internal/template/catalog.yaml` (generated hash regen, non-Go) and the two markdown mirrors.

M1: AC-001=PASS, AC-002=PASS, AC-003=PASS, AC-005=PASS, AC-006=PASS | evidence: this section (grep outputs verbatim above) + test runs on this tree | commit: pending-backfill (M1 commit)

Gaps: AC-004/009/010/012/013 are M3 runtime-probe ACs (not executable at M1 by §D.1 gate mapping); AC-007 full closure and AC-008 are M2 (kanban-dispatch doctrine surface untouched at M1 by scope).

### M2 — deputy doctrine surface (kanban-dispatch extension)

Measured on this tree, t283 worktree, base `9ff2e1ac2`:

- AC-008 [HARD] preservation: `grep -c '\[HARD\]' kanban-dispatch.md` → before `30` (auditor baseline, re-measured pre-edit), after `32`. Delta = 2 NEW clauses ("The deputy never holds a power of consequence", "Nothing structural moves with the delegation") — pure insertion, `git diff` shows `10 insertions(+), 0 deletions(-)`; every existing [HARD] clause survives verbatim (no deleted lines). `grep -ci deputy` → `5` (RED-now was 0); detail → `14` (RED-now 0).
- Detail companion: `git diff --numstat` → `23 1`; the single "deleted" line is the Factory "Evidence, verdict, integration unchanged" bullet extended in place (original sentence preserved verbatim; plan.md M2 item 3's one-line Factory connection). No [HARD] line touched.
- AC-007 full closure: (a) `diff -q` local vs mirror → STUB-IDENTICAL and DETAIL-IDENTICAL (zero divergence — additions carry no forbidden tokens in either copy); (b) `make build` rc=0, catalog.yaml no-diff (rules files are not catalog entries — verified by grep before build); (c) neutrality `grep -rc 'SPEC-LEAD-DEBOTTLENECK\|REQ-LDB' internal/template/templates/` → 0 hits; `grep -rc 't283' internal/template/templates/.claude/rules/moai/workflow/` → 0 hits. §25.3 5-item self-check passed on both mirrors (no SPEC IDs / REQ tokens / dates / audit citations / internal paths).
- Stub growth bound: +10 lines (≤ ~25 bound), +1,444 bytes — exceeds the 1,000B single-edit threshold of `rule-authoring.md` (b); growth statement + non-invoking cost carried in the M2 commit body per the duty.
- Full affected package: `go test ./internal/template/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/template 31.935s` (includes TestTemplateNoInternalContentLeak on the mirrors).
- spec lint: `bin/moai spec lint .moai/specs/SPEC-LEAD-DEBOTTLENECK-001/spec.md` → `✓ No findings`.
- PRESERVE: working tree touches only the 4 rule/mirror files + this progress.md; agent files (M1 charter) untouched; no `internal/`, `pkg/`, `cmd/` Go paths.

M2: AC-007=PASS (full closure), AC-008=PASS | evidence: this section (counts verbatim above) + test run on this tree | commit: pending-backfill (M2 commit)

Gaps: none at M2 close beyond the M3-set (AC-004/009/010/012/013 remain M3 runtime probes; AC-011 final closure spans the whole branch — Go paths remain 0).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
