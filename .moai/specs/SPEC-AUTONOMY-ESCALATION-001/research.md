# research.md — SPEC-AUTONOMY-ESCALATION-001

## §A — Measurement basis

All premises below were measured in this run against tree `develop` at `ca1d5dc43`
(worktree `.claude/worktrees/t1235`, branch `WT-escalation-detector`), with `grep`,
`sed -n`, and `ls` over the working tree. No premise is carried over from the design
document; where the design table asserted an asset, this section records what the tree
actually holds.

## §B — Existing-asset premise table

| # | Asset named in the design table | Exists? | Evidence (file:line) | Consequence for scope |
|---|---|---|---|---|
| P1 | AC snapshot commit guard (`scripts/ac-baseline`) | **Yes, but local-only and a different comparison** | `scripts/ac-baseline/check-staged.sh:1-4` ("Local-only dev tool: no template mirror, never distributed"); `:25` baseline = `.moai/reports/t338/ac-count-baseline.txt`; counter body extracted from `manager-docs.md` sentinels (`:117-143`) | Mechanism not reusable in a distributed detector. Class 1 consumes A1 verify's recorded-vs-measured `acceptance.sha256` / `acceptance.ac_count` (REQ-AE-005); A1 ships its own Go port of the counter (P15), so the guard is not reused. |
| P2 | `frozenInstructionFiles` hook | **Yes, but identity-scoped** | `internal/hook/pre_tool.go:1230-1231` (`CLAUDE.md`, `CLAUDE.local.md`); `:1236-1258` `checkHarnessFrozenZone` returns early unless `agentID == harnessLearnerIdentity` (`:1237-1239`); prefix zones `:1225-1228` | The *path set* is reusable; the *trigger* is not (it fires only for harness-learner). REQ-AE-007 trips for any caller and does not change the existing deny. |
| P3 | Constitution zone registry | **Yes (doc + loader)** | `.claude/rules/moai/core/zone-registry.md:33-36` (`zone_class` enum, Frozen → `canary_gate`); runtime loader `internal/constitution/loader.go:78` `LoadRegistry`, zone filter `:47` `FilterByZone` | A1 derives `frozen_files` from the registry Frozen list supplied through this loader (A1 `design.md` § Frozen Files item 1, line 179 at `67a2f55cb`). REQ-AE-007 consumes that derived list, obtained once at arming (design.md §C.6); the detector does not read the registry itself. |
| P4 | Status-transition ownership matrix → PreToolUse path check | **Matrix + advisory PostToolUse hook exist; no per-contract path check** | `.claude/hooks/moai/status-transition-ownership.sh:1-5` (PostToolUse, invoker vs matrix, advisory) | Class 3 (REQ-AE-008) is new detection logic; the existing hook is not extended or rewired. |
| P5 | Code graph `graph_file_api` before/after diff | **Tool exists; working-tree only** | `internal/cli/mcp_server.go:598-604` registration; `internal/graph/codequery.go:65-81` `FileAPI(projectRoot, relPath)` opens the working-tree file, no commit argument | Before-side must be obtained another way (design.md §C.4). Declaration extraction depends on CGO/grammar availability (`codequery.go:93-95` returns "extraction unsupported") → REQ-AE-022 not-observed labelling. |
| P6 | `audit_multi` disagreement flag | **Yes, tri-state** | `internal/cli/mcp_convergence.go:119-130` `DisagreementFlag *bool`; `:17-21` "Disagreement is INFORMATION, not a GATE" | Class 5 treats `nil` as not-observed, `true` as trip, `false` as no trip (REQ-AE-010, REQ-AE-022). The convergence doctrine is unchanged. |
| P7 | Local-pass / CI-fail reader | **Not located** | `grep -rn 'check-runs\|checkRuns\|statusCheckRollup' --include='*.go' internal` → 0 lines (non-test). Not established absent — only not located by this pattern set. | Class 5's CI limb consumes a recorded verdict only; a CI poller is out of scope. |
| P8 | Deny rules / destructive-command guard | **Yes, partial for class 6** | `internal/hook/pre_tool.go:505-521` denylist runs unconditionally; `:219-220` force-push to main/master and `branch -D main`; `.claude/settings.json:543-552` force-push denies (deny block starts `:527`); `.claude/settings.json:453` `Bash(git tag:*)` sits in the **allow** block (`:411`) | A plain push to `main`, a tag, and a release are not denied today. Class 6 detects and reports them; adding denies is A3. |
| P9 | Super-advisor E1 (same diagnostic ×3) | **Doctrine only** | `.claude/agents/moai/super-advisor.md:60` E1 row | No counter exists. Class 8 needs its own fingerprint count (REQ-AE-015). Related mechanical signals: `internal/hook/evidence_writer.go:578-595` emits `test_fail:<pkg>:` keys; goal stagnation guard `internal/goal/evaluate.go:44,71,95` counts no-progress iterations, a different predicate. |
| P10 | Card needs-decision state in the queue store | **No, and forbidden** | States `internal/kanban/backlog_store.go:60-66` (`queued`/`picked`/`dropped`); `internal/kanban/backlog_schema_freeze_test.go:3-5` ("admit no fourth state"), `:70` pinned CHECK. Positive control: `grep -rn 'needs-decision\|needs_decision\|NeedsDecision'` over `internal`, `.claude`, `.moai/config` → 0 lines, while the same grep shape for `BacklogStatePicked` → 17 non-test lines | Needs-decision is a record beside the queue (REQ-AE-018). Its location is fixed by lead ruling 09-26 #4 (spec.md §I). |
| P11 | Audit retry ceiling (class 9) | **Yes** | `.moai/config/sections/harness.yaml:69-78` `plan_audit_tier_ceilings` (Tier S = 1, M = 2, L = 3; absent tier → L) | Superseded for this SPEC by lead ruling 09-26 #5: REQ-AE-016 takes the cap from `budget.audit_retries` for plan-audit and sync-audit alike; no sync-audit ceiling key exists or is added. |
| P12 | Budget counters (class 7) | **Partial** | `internal/goal/schema.go:35-44` goal `Ceiling` / `MaxTurns` (default 30) exists only when a goal is armed | No operation counter exists; the detector counts operations itself (design.md §C.6). |
| P13 | Existing `autonomy` config key | **No `workflow.autonomy` key** | `grep -n autonomy .moai/config/sections/workflow.yaml` and the template copy → 0 lines; `MOAI_AUTONOMY_TIER` at `internal/config/envkeys.go:139-148`, reader `internal/config/autonomy.go:8-9` | New key has no collision. `harness.escalation` at `internal/config/types.go:1199-1243` is a homonym with a different meaning (C2). |
| P14 | Mission contract sealing / decision validation (named by the A1 draft) | **Yes** | `internal/mission/contract.go:32` `MissionContract`, `:80` `SealMissionContract`; `internal/mission/policy.go:200` `ValidateMissionDecision` | The mission-validator projection moved to card t1245 (spec.md §K); this SPEC does not use it. |
| P16 | Card → SPEC link in the queue store | **Field exists, almost never filled** | `internal/kanban/backlog_store.go:81` `SpecID *string` with JSON tag `spec_id`; absent spec id is JSON null (`:69`). Fill rate: 0 of 119 live cards (plan-audit iteration 3 P1, `.moai/reports/t1235/plan-audit-iter3.md`; not re-measured here) | Not an input. The resolver reads the contract's `card` field instead (REQ-AE-002, lead ruling 09-26 (3) #1); keeping `spec_id` filled is F1's concern. |
| P15 | A1 contract schema | **On branch `WT-contract-schema`, not on `develop`** | SPEC-AUTONOMY-CONTRACT-001 v0.5.0 `design.md` at commit `67a2f55cb`, read with `git show 67a2f55cb:.moai/specs/SPEC-AUTONOMY-CONTRACT-001/design.md`; derived fields lines 162-166; no `card` field (only `card` mentions are the header and the F1 record note, lines 4-5, 516-527); verify-from-PreToolUse in A1 `spec.md:448`; mission projection still labelled "(for A2)" at `design.md:446` | Every contract-reading requirement keeps the tag 「A1 plan-audit 통과본으로 재확인」; requests R8-R10 in spec.md §F.2. |

## §C — Gaps in this measurement

- The A1 pin `67a2f55cb` was read with `git show` from this worktree; `6d98ca466` was read the
  same way and is byte-identical to the earlier scratch copy (`cmp` exit 0). Differences between
  the two were read with `diff`: only the kickoff decider section changed.
- The `spec_id` fill rate is carried from the iteration-3 audit, not re-measured in this run.

- The deny-list classification of `git tag` was read from the local `.claude/settings.json`
  only; the distributed template `settings.json.tmpl` was not read.
- No test was run; every row is a source read, not a behavior observation.
- `graph_file_api` behaviour on a non-Go language was not exercised.
- P7 is a pattern search; a CI reader under another name would not be found by it.

## §D — Prior art in this repository

- `SPEC-AUTONOMY-TIERS-001` defines `MOAI_AUTONOMY_TIER` (hook block strength); this SPEC is
  orthogonal and keeps it untouched.
- `SPEC-ACSNAPSHOT-COMMIT-GUARD-001` owns the AC snapshot guard; this SPEC does not modify it.
