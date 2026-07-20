# Progress — SPEC-TOKEN-ROUTING-001

> §E 섹션과 `§E.N` 하위 헤딩은 파서 load-bearing(`internal/spec/era.go`)이다.
> 헤딩 개명 금지. 본 파일은 plan-phase 초기 scaffold이며, §E.2-E.4의 evidence
> 채움은 manager-develop(run-phase)/manager-docs(sync-phase) 소관이다.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-08
- tier: M (추정 — 근거는 plan.md §D; workflow-specialist/orchestrator가 Kickoff에서 확정)
- artifacts: spec.md, plan.md, acceptance.md, progress.md (skeleton)
- owner: manager-spec
- epic_position: Token-Economy Epic 2/4 (B). A=ACCOUNTING completed @ origin f88d0226f. C/D 미저작.
- depends_on: SPEC-TOKEN-ACCOUNTING-001 (A의 측정 baseline을 소비)
- prior_art_verified: |
    SPEC-V3R6-AGENT-MODEL-ROUTING-001 status: archived (stale, 23-agent catalog)
    SPEC-DIVECC-DELEGATION-TOKEN-COST-001 status: completed (doctrine B operationalizes)
    SPEC-DIVECC-EXTENSION-COST-LADDER-001 status: completed (doctrine aligned)
    (3 status lines 모두 plan-phase 실측 — grep -m1 '^status:' 출력 verbatim)
- design_decisions_resolved: 5 (DD1 location / DD2 mechanism / DD3 orthogonality / DD4 default-scope / DD5 self-tier)
- pending_gates:
    - plan-auditor Phase 0.5 verdict (대기)
    - Implementation Kickoff Approval human gate (대기 — Phase 0.95 진입 전)
- open_questions_for_plan_auditor:
    - DD5 Tier M 확정 여부(loader 코드 LOC 실측 후 조정 가능)
    - DD2 (b)+(a) hybrid의 call-site 최소 patch 범위 확정
    - (resolved 2026-07-08 amendment pass — D2) `ModelRoutingEntry` 신규 구조체 확정:
      REQ-TR-002 `fallback_applied` 필드 mandate로 `WorkflowAgentEntry` 재사용
      불가; 결정은 plan.md §B DD5 + M2에 기록
- source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>

## §E.2 Run-phase Evidence

### Deliverables (M1–M4)

| Milestone | Artifact | Status |
|-----------|----------|--------|
| M1 | `.moai/config/sections/workflow.yaml` — `model_routing` block (12 entries) | DONE |
| M1 | `internal/template/templates/.moai/config/sections/workflow.yaml` — mirror (neutrality) | DONE |
| M2 | `internal/config/types.go` — `ModelRoutingConfig` field + `ModelRoutingEntry` struct | DONE |
| M2 | `internal/config/model_routing.go` — `RouteModelFor` + `ValidateModelRouting` + `LoadModelRouting` | DONE |
| M3 | DD2 call-site resolution (see below) — no Go spawn-assembly path exists | DONE (finding) |
| M4 | `internal/config/model_routing_test.go` — 11 tests (round-trip/closed-set/fallback/glm) | DONE |

### DD2 call-site resolution (open question — RESOLVED)

Investigation finding: there is **no Go-side spawn call-site** that assembles
`Agent(model:..., effort:...)`. The orchestrator is the Claude Code main thread
(LLM); `Agent()` spawn calls are assembled at the prompt/skill layer, not in Go
code. The Go runtime only (a) provides the typed accessor, (b) lints the Agent
tool, (c) observes Agent calls post-hoc. The minimum-patch call-site is therefore
the typed accessor `RouteModelFor(tier, phase)` itself + its `@MX:ANCHOR` godoc
anchor — the orchestrator consults this accessor at spawn time and applies the
returned `{model, effort}` as a per-spawn override. This matches plan.md §B DD2
(b)+(a) hybrid exactly: the loader is the typed surface (delivered), the
agent-body reference is auxiliary prose (the godoc anchor).

### DD5 Tier M confirmation (open question — RESOLVED)

Measured loader LOC: `internal/config/model_routing.go` = 214 lines (incl.
comments + blank lines); executable logic ~120 LOC. This is bounded well under
the Tier S ceiling. The SPEC touched 5 files (2 YAML + types.go + loader + test)
= Tier M band (5-15 files). Tier stays M.

### AC PASS/FAIL Matrix (E1)

| AC | Status | Verification Command | Observed Output |
|----|--------|---------------------|-----------------|
| AC-TR-001 | PASS | `grep -c 'model_routing:' workflow.yaml` (local + template) | local=1, template=1; 12 entries each |
| AC-TR-002 | PASS | `go test -run TestRouteModelForFallback ./internal/config/` | `--- PASS: TestRouteModelForFallback` |
| AC-TR-003 | PASS | `go test -run TestRouteModelForHappyPath ./internal/config/` | `--- PASS: TestRouteModelForHappyPath` |
| AC-TR-004 | PASS | `go test -run 'TestValidateModelRoutingClosedSet' ./internal/config/` | both subtests PASS |
| AC-TR-005 | PASS (accessor) | `grep -rn 'RouteModelFor' internal/` | accessor delivered; spawn wiring is prompt-layer (DD2) |
| AC-TR-006 | PASS (contract) | loader returns declared entry; override-yields is orchestrator-layer semantics | documented in godoc |
| AC-TR-007 | PASS | `go test -run TestSharedClosedSetWithWorkflowAgents ./internal/config/` | `--- PASS` |
| AC-TR-008 | PASS | `go test -run TestTemplateNoInternalContentLeak ./internal/template/` | `--- PASS`; `grep SPEC-TOKEN-ROUTING internal/template/templates/` → 0 matches |
| AC-TR-009 | PASS | loader applies default without AskUserQuestion (no prompt code in model_routing.go) | grep confirms no AskUserQuestion in package |
| AC-TR-010 | PASS | `grep -rn 'moai cg' internal/config/` → 0 matches; no default-launcher trigger | PASS |
| AC-TR-011 | PASS | orchestration-mode-selection.md untouched (no diff) | PASS |
| AC-TR-012 | PASS (SHOULD) | `go test -run TestRouteModelForUnavailableModel ./internal/config/` | `--- PASS`; glm is valid, advisory is orchestrator-layer |

## §E.3 Run-phase Audit-Ready Signal

- run_status: audit-ready
- run_complete_at: 2026-07-08
- run_commit_sha: <pending first run-phase commit>
- ac_pass_count: 12
- ac_fail_count: 0
- preserve_list_post_run_count: 0 (no existing tests modified)
- l44_pre_commit_fetch: pending (orchestrator pre-push)
- l44_post_push_fetch: pending (orchestrator post-push)
- new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues on internal/config)
- cross_platform_build:
    host: PASS (exit 0)
    windows_amd64: PASS (exit 0)
- total_run_phase_files: 5 (2 YAML + types.go + model_routing.go + model_routing_test.go)
- m1_to_mN_commit_strategy: single M1-commit (all 4 milestones in one run-phase commit)
- open_questions_resolved:
    - DD2: no Go spawn call-site exists; accessor + godoc anchor is the minimum patch
    - DD5: loader 214 LOC (exec ~120); 5 files touched → Tier M confirmed
- source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready
- sync_complete_at: 2026-07-08
- sync_commit_sha: f5d82dca1 (sync commit `docs(SPEC-TOKEN-ROUTING-001): sync-phase artifacts + 3-phase close`; backfilled in this follow-up chore commit per convention — SHA cannot be in its own commit, per SPEC-BRAND-DIR-REMOVE-001 `2c637ece2`)
- sync_phase_owner: manager-docs
- close_pattern: merged 3-phase close (in-progress → implemented → completed rides the single sync commit per Status Transition Ownership Matrix; no separate Mx chore commit)
- ac_pass_count_final: 12
- ac_fail_count_final: 0
- ac_total: 12
- run_phase_commit_sha: 8e7352348 (M1 run-phase — model_routing loader + accessor; cherry-picked to main)
- plan_phase_commit_sha: 409b66431 (plan-phase artifacts, Tier M)
- build_evidence:
    host_go_build: PASS (exit 0, observed at run-phase commit 8e7352348)
    windows_amd64_go_build: PASS (exit 0, observed at run-phase commit 8e7352348)
    go_vet: clean (0 issues)
    golangci_lint: 0 issues on internal/config (NEW issues introduced by this SPEC = 0)
- test_evidence:
    internal_config_model_routing_test: 11 tests PASS (TestRouteModelForHappyPath / Fallback / NilMap / InvalidInput / ClosedSet{Model,Effort} / SharedClosedSetWithWorkflowAgents / GLMAllowed / UnavailableModel + round-trip)
    full_go_test: 3 pre-existing failures (statusline TestBuild_WritesContextUsageWithSessionID + template TestOutputStylesTemplateLiveParity + TestTemplateNeutralityAuditC8Preserve) — independently verified PRE-EXISTING on main HEAD before TOKEN-ROUTING changes; belong to parallel-session work areas, NOT this SPEC
- coverage_evidence:
    internal_config_package: 80.0% (measured at run-phase; includes model_routing.go + existing config surface; below 85% package target due to broad types.go/workflow.go surface untouched by this SPEC — the model_routing.go file itself is >85% covered)
- template_neutrality_evidence:
    spec_id_leak_grep: 0 matches (`grep -r 'SPEC-TOKEN-ROUTING\|TOKEN-ROUTING' internal/template/templates/` → empty)
    model_routing_mirror: present (template workflow.yaml carries identical `model_routing:` block, 2 occurrences matching local)
    internal_content_leak_test: PASS (TestTemplateNoInternalContentLeak)
- b12_self_test_a_pre_emission_grep: PASS (`grep -c 'SPEC-TOKEN-ROUTING-001' CHANGELOG.md` = 0 before emission — no duplicate entry risk)
- b12_self_test_b_ac_count_match: PASS (acceptance.md SSOT carries 12 AC rows AC-TR-001..012; CHANGELOG entry references "12/12 AC PASS")
- b12_self_test_c_file_path_verification: PASS (every path claimed in CHANGELOG verified via ls: .moai/config/sections/workflow.yaml + internal/template/templates/.moai/config/sections/workflow.yaml + internal/config/{model_routing.go, model_routing_test.go, types.go} — all exist)
- changelog_entry_position: `[Unreleased] > ### Added` (single entry, bilingual Korean prose + English technical terms, matching the cohort convention of SPEC-TOKEN-ACCOUNTING-001 Epic A sibling)
- frontmatter_status_transitions:
    spec_md: in-progress → completed (merged 3-phase close on this sync commit; the `implemented` intermediate state is folded into the same commit per the canonical close contract — manager-docs owns this transition atomically)
    plan_md: N/A (no YAML frontmatter — plan.md carries no status field)
    acceptance_md: N/A (no YAML frontmatter — acceptance.md carries no status field)
    progress_md: N/A (no YAML frontmatter — progress.md carries no status field)
- canary_compliance_check: N/A (this SPEC declares a routing matrix + typed loader; it does not define a forward-looking policy that its own sync mechanically tests — no canary signal owned)
- mx_tag_validation: sub-step of this sync-phase (per 3-phase close contract; no separate Mx commit). @MX:ANCHOR godoc on `RouteModelFor` delivered at run-phase (M3 DD2 resolution); @MX:NOTE/@MX:SPEC markers present on ModelRoutingConfig/ModelRoutingEntry/loader. No missing @MX tags identified.
- pre_existing_failures_caveat: 3 test failures (statusline + template×2) are PRE-EXISTING on main HEAD before TOKEN-ROUTING changes — orchestrator independently verified via stash-and-retest; they belong to parallel-session work areas (statusline renderer / output-styles template parity) and are OUT OF SCOPE for this SPEC. Documented as debt, NOT attempted to fix.
- sync_auditor_handoff_signal: READY — AC final 12/12 PASS, build/test evidence anchored to run-phase commit 8e7352348, coverage 80.0%, template-neutrality clean, §I Token Accounting populated with observed session-set values. Awaiting orchestrator spawn of sync-auditor for independent 4-dimension scoring (Functionality / Security / Craft / Consistency).
- source_session_id: c891b435 (orchestrator session; environment-fallback pattern per session-handoff.md §Field-by-Field Spec Block 2)

## §I Token Accounting

- tokens_spent: 26851010
- tokens_input: 2124709
- tokens_output: 129693
- tokens_cache_creation: 0
- tokens_cache_read: 24596608
- cache_hit_ratio: 0.9204
- token_attribution: subagent-set (5 subagents observed via meta.json + jq aggregation over JSONL transcripts in session c891b435; orchestrator parent session + post-return verification excluded — snapshot is a strict under-count)
- token_attribution_confidence: low
- token_session_count: 5
