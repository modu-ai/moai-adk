---
id: SPEC-MODEL-ROUTING-WIRE-001
title: "Wire the Tier×Phase Model Routing Matrix into Spawn Paths — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "model-routing, tier-phase-matrix, moai-route, spawn-wiring, model-policy, haiku-inherit, workflow-reflex, plan"
---

# SPEC-MODEL-ROUTING-WIRE-001 — Plan

> plan.md is the derived execution plan. WHAT/WHY SSOT is spec.md. This document carries the HOW skeleton; function names/signatures are run-phase discretion.

## §A Context

### §A.1 Problem summary

The `model_routing` Tier×Phase matrix and its typed accessor `RouteModelFor` are dead wiring (zero call sites outside `internal/config`; no skill/rule references), while the workflow.yaml comment claims spawn-time consultation that exists nowhere (SPEC-TOKEN-ROUTING-001 D1 debt). Default orchestration paths (Mode 4 fan-out, Mode 5 run path) give zero model/effort guidance, so spawns inherit the orchestrator's expensive model. Two contradictions compound it: manager-docs/manager-git frontmatter flipped `haiku`→`inherit` (uncommitted) against model-policy.md's prose, and `team.default_model: opus[1m]` sits outside the documented enum with the hazardous `[1m]` suffix.

### §A.2 Evidence baselines (measured 2026-07-09 by this agent via Bash/Read, vci §2 attribution)

```
grep -rn "RouteModelFor" --include='*.go' internal/ cmd/ pkg/ | grep -v _test
  → internal/config/model_routing.go:{69,83,87,89} + internal/config/types.go:{361,365}  (zero external call sites)
grep -c model .claude/skills/moai/workflows/run.md                → 0
grep -c model .claude/skills/moai/workflows/run/phase-execution.md → 0
grep -n '^model:' .claude/agents/moai/manager-docs.md             → 14: model: inherit   (uncommitted flip; git status M)
grep -n '^model:' .claude/agents/moai/manager-git.md              → 13: model: inherit   (uncommitted flip; template mirrors same)
model-policy.md § Inherit-by-Default                              → "except manager-docs and manager-git which use model: haiku"
                                                                    + "Exceptions (do NOT migrate to inherit): model: haiku agents"
builder-harness.md model table (near :165)                        → haiku speed-critical exception row present
workflow.yaml:24                                                  → default_model: opus[1m]  (out-of-enum + [1m] suffix)
workflow.yaml:154-163                                             → aspirational comment "the orchestrator consults RouteModelFor..."
workflow.yaml:164-176                                             → 12-entry S/M/L × plan/run/sync/mx matrix
run/mode-orchestration.md (near :43)                              → "implementer (inherit) + tester (inherit) + reviewer (inherit)"
                                                                    vs workflow.yaml role_profiles (implementer sonnet/xhigh, researcher haiku/low, ...)
internal/cli/harness_route.go:161                                 → Use: "route" under `moai harness` parent (name-adjacent, no cobra collision)
```

Template mirrors verified present (2026-07-09): workflow.yaml, run.md, run/mode-orchestration.md, orchestration-mode-selection.md, model-policy.md, manager-docs.md, manager-git.md — all under `internal/template/templates/`. Template-First applies to every doc/config edit; `moai route` Go code is NOT templated.

Line numbers are indicative; re-verify content anchors at run-phase (line-drift asymmetry lesson).

### §A.3 Approach — three milestones

- **M1 — `moai route` CLI (Go, TDD)**: new top-level cobra subcommand calling `RouteModelFor`; plain + `--json` output; invalid-enum rejection; `TestRoute_NoAskUserQuestion` static guard; table-driven tests over the 12 matrix entries + fallback path.
- **M2 — doc wiring (prompt layer)**: run.md Phase 0.95 pre-spawn instruction (consult `model_routing[<TIER>-<phase>]` via `moai route`, pass as per-spawn runtime args, yield to explicit override); orchestration-mode-selection.md §B.1 consultation input + §A row 4/§C.2 Mode 4 worker-model guidance (sonnet default, haiku for pure extraction, effort per workflow_agents).
- **M3 — contradiction reconciliation + config fixes + template sync**: apply the D1-decided direction across manager-docs/manager-git frontmatter + model-policy.md prose + builder-harness.md table; `team.default_model: inherit`; truthful model_routing comment; mode-orchestration.md role_profiles reconciliation; template-first for all of the above + `make build`.

### §A.4 Tier evidence (M)

- Files affected: ~12-14 (1-2 new Go + test, 4 doc surfaces, 2 agent files, 1 config — each × template mirror where applicable) — at the upper edge of Tier M's 5-15 band.
- LOC estimate: 300-600 (small CLI + tests; doc edits are surgical) — Tier M band.
- No constitutional change; contradiction resolution is textual alignment to an existing doctrine → not Tier L.

### §A.5 PRESERVE / EXTEND map

| Surface | Disposition |
|---------|-------------|
| `internal/config/model_routing.go` (accessor + validation) | PRESERVE (call, don't change) |
| workflow.yaml matrix VALUES (164-176) | PRESERVE |
| workflow.yaml comment (154-163) + `team.default_model` (24) | EXTEND/FIX |
| run.md Phase 0.95 + orchestration-mode-selection.md §A/§B.1/§C.2 | EXTEND |
| run/mode-orchestration.md team-composition line | FIX (reconcile to role_profiles SSOT) |
| manager-docs.md / manager-git.md frontmatter (+ mirrors) | FIX per D1 direction |
| model-policy.md § Inherit-by-Default prose; builder-harness.md table | FIX per D1 direction |
| `moai harness route` | PRESERVE (untouched; help-text distinctness only on the NEW command) |
| workflow.yaml role_profiles / workflow_agents values | PRESERVE |

## §B Known Issues (filtered, Tier M)

- **B2 Cross-SPEC conflicts**: SPEC-TOKEN-ROUTING-001 (closed) authored the matrix; SPEC-CC2178-MODEL-POLICY-REPAIR-001 and the GLM allowlist SPECs touched model-policy.md. Re-read model-policy.md § Baseline-Refill Breaker before editing — the Sonnet 5 resolution prose must not be disturbed by the haiku-exception edit.
- **B4 Frontmatter schema**: agent frontmatter (`model:`) edits must not disturb other fields; agent files are lint-checked by agentlint.
- **B8 Working-tree hygiene**: manager-docs.md/manager-git.md are ALREADY modified in the working tree (the uncommitted flip). The D1 decision determines whether the run-phase reverts or keeps those hunks — coordinate with the orchestrator before committing; do NOT blanket `git add`.
- **B10 Scope discipline**: parallel Workflow-Reflex siblings touch loop.md / harness code — stay off those surfaces.
- **B1 Cross-platform**: pure cobra + config read; verify `GOOS=windows` build anyway.

## §C Pre-flight checklist

```bash
git branch --show-current && git rev-parse HEAD
git status --porcelain | grep -E 'manager-docs|manager-git'        # confirm the uncommitted flip is still present
grep -rn "RouteModelFor" --include='*.go' internal/ | grep -v _test # re-verify zero external call sites
grep -n "route" internal/cli/root.go internal/cli/harness_route.go  # name-collision surface re-check
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5
go test ./internal/cli/... ./internal/config/... 2>&1 | tail -5     # baseline
```

## §D Constraints + open decisions

Constraints: see spec.md §Constraints ([1m] workaround preserved; matrix yields to explicit override; no CLI prompting; naming distinctness).

**D1 — haiku↔inherit direction (plan-phase decision point; REQ-MRW-005).** Both directions, per the brief:

| Option | Change set | Rationale | Risk |
|--------|-----------|-----------|------|
| **A — revert flips to `haiku` (RECOMMENDED)** | Revert manager-docs/manager-git (+ mirrors) to `model: haiku`; model-policy.md + builder-harness.md prose stand as-is | haiku has no `[1m]` variant → immune to the entitlement-inheritance bug; matrix S-sync/S-mx/M-mx rows expect haiku-class cheap routing; policy prose already teaches this | If the flip was made for an unrecorded quality regression in haiku sync output, reverting reintroduces it — check git stash/session provenance first |
| B — ratify `inherit`, update policy | Keep frontmatter `inherit`; rewrite model-policy.md exception prose + builder-harness.md table to remove the haiku exception | Uniform inherit simplifies the catalog | Loses the cheap-model routing for mechanical sync/git work that the matrix assumes; contradicts workflow_agents taxonomy economics |

The orchestrator MUST surface D1 to the user at Implementation Kickoff Approval (AskUserQuestion, option A first with "(권장)") before M3 executes. M1/M2 are independent of D1 and may proceed first.

**D2 — new command naming.** `moai route` (brief-specified; different cobra parent than `moai harness route`, no registration collision) — RECOMMENDED. Alternative if the user prefers disambiguation: `moai config route`. Help text must state "Tier×Phase model/effort routing (see also: moai harness route — harness-LEVEL routing, a different concern)".

**D3 — output contract.** Plain output single line `model=<m> effort=<e> fallback=<bool>`; `--json` mirrors `ModelRoutingEntry`. Exact field names run-phase discretion; must be stable enough for the run.md instruction to cite.

## §E Self-Verification (run-phase deliverables)

Per manager-develop-prompt-template.md §E (E1-E7), vci 5-section format each:
- E1: AC matrix (acceptance.md §D) with verbatim outputs.
- E2: cross-platform builds exit 0 (incl. `make build` after template edits).
- E3: `go test -cover ./internal/cli/...` (new command file ≥85%).
- E4: `TestRoute_NoAskUserQuestion` grep guard green.
- E5: lint NEW vs baseline.
- E6: commit SHAs + push state; D1 direction recorded in commit body.
- E7: blocker report if D1 was not pre-resolved by the orchestrator.

## §F Milestones (priority-ordered; no time estimates)

| Milestone | Scope | REQs | Exit criterion |
|-----------|-------|------|----------------|
| M1 — `moai route` CLI | new subcommand + tests + static guard | REQ-MRW-002, REQ-MRW-003 | AC-MRW-001..004, AC-MRW-013 PASS |
| M2 — doc wiring | run.md pre-spawn instruction; orchestration-mode-selection.md §B.1 + §A row 4/§C.2 guidance | REQ-MRW-001, REQ-MRW-004 | AC-MRW-005..007 PASS |
| M3 — reconciliation + config + template sync | D1 direction applied across 5 surfaces; team.default_model=inherit; truthful comment; mode-orchestration reconciliation; template-first + make build | REQ-MRW-005..009 | AC-MRW-008..012 PASS |

Dependency note: M3 blocks on the D1 user decision; M1/M2 do not.

## §G Anti-Patterns (do NOT)

- Pinning `model:` in any agent FRONTMATTER as the wiring mechanism — the wiring channel is per-spawn runtime args (frontmatter pins trigger the `[1m]` bug the policy exists to avoid).
- Changing matrix values or RouteModelFor logic "while wiring".
- Registering the new command under the `harness` parent or reusing its flag surface (concern confusion).
- Editing live-tree docs without the template mirror (Template-First violation; CI neutrality guard).
- Blanket `git add` — the working tree carries unrelated modified files plus the D1-affected agent files.
- Resolving D1 silently inside the run-phase without the recorded user decision.

## §H Cross-References

- spec.md (SSOT), acceptance.md (AC matrix), progress.md (§E skeleton).
- SPEC-TOKEN-ROUTING-001 progress.md (D1 wiring debt origin — AC-TR-005 DD2).
- `.claude/rules/moai/development/model-policy.md` (per-spawn-arg [1m]-safety rationale; § Inherit-by-Default; § Baseline-Refill Breaker — read before editing).
- `internal/cli/CLAUDE.md` (subcommand registration + no-AskUserQuestion static guard conventions).
- Downstream: SPEC-ADVISOR-RUNG-001 (depends on this SPEC's M1+M2 outputs).
