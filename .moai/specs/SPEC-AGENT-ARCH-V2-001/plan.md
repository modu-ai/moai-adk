---
id: SPEC-AGENT-ARCH-V2-001
title: "MoAI Agent Architecture v2 — Implementation Plan"
version: "0.2.0"
status: completed
created: 2026-07-09
updated: 2026-07-10
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents/moai + internal/config"
lifecycle: spec-anchored
era: V3R6
tier: L
tags: "agent-arch, super-advisor, manager-design, no-haiku, 3-tier, claude-design, token-policy, plan"
---

# SPEC-AGENT-ARCH-V2-001 — Plan

> plan.md is the derived execution plan. WHAT/WHY SSOT is spec.md. The design authority is `.moai/reports/agent-architecture-redesign-v2-20260709.html` (the SSOT HTML); this document carries the HOW skeleton per milestone.

## §A Context

### §A.1 Problem summary

The v2 architecture (4 changes per §01 of the design SSOT) has zero footprint in the live tree, and two Workflow-Reflex draft SPECs whose concerns v2 subsumes remain active. The v2 architecture moves the catalog from 8 to 10 agents (super-advisor + manager-design new), eliminates Haiku across the catalog (Haiku → Sonnet effort low/medium), and adds a 3-tier token policy (`max`/`medium`/`low`) wired through an extended `RouteModelFor(specTier, phase, perfTier)` accessor. The two supersede targets (ADVISOR-RUNG-001, MODEL-ROUTING-WIRE-001) carry `status: draft` and must exit via `* → superseded`.

### §A.2 Evidence baselines (measured 2026-07-09 by this agent via Read/Bash, vci §2 attribution)

```
ls .claude/agents/moai/                                        → no super-advisor.md, no manager-design.md
grep -rn "super-advisor\|manager-design" .claude/ CLAUDE.md    → 0 matches
CLAUDE.md §4 (lines 95-106)                                    → "8 retained agents" (ceiling = 8)
internal/config/model_routing.go:89                            → func (c *Config) RouteModelFor(tier, phase string) (ModelRoutingEntry, error)  (2-arg signature)
grep -rn "RouteModelFor" --include='*.go' internal/ cmd/ pkg/  → internal/config/{model_routing.go:69,83,87,89, types.go:361,365}  (zero external call sites)
.moai/config/sections/workflow.yaml:171-183                    → single model_routing: block (12 entries, S/M/L × plan/run/sync/mx)
.moai/config/sections/workflow.yaml:154                         → workflow_agents.read-only-extract: { model: haiku, effort: low }
.moai/config/sections/llm.yaml:5-9                             → performance_tier: "" + claude_models: {high: opus, medium: sonnet, low: haiku}
internal/config/model_routing.go:31                            → "haiku": true in validRoutingModels
grep -n "Inherit-by-Default" model-policy.md                   → line 30 (haiku-exception prose: "except manager-docs and manager-git")
.mcp.json                                                      → no DesignSync server registered (GAP — see research.md §H)
.moai/specs/SPEC-ADVISOR-RUNG-001/spec.md:5                    → status: draft
.moai/specs/SPEC-MODEL-ROUTING-WIRE-001/spec.md:5              → status: draft
```

Template mirrors verified present (2026-07-09): CLAUDE.md, agent files, spec-workflow.md, model-policy.md, agent-authoring.md, workflow.yaml, llm.yaml, team-protocol.md, team-pattern-cookbook.md — all under `internal/template/templates/`. Line numbers are indicative; re-verify content anchors at run-phase (line-drift asymmetry lesson).

### §A.3 Approach — 4 milestones (per §06 of the design SSOT)

- **M1 — super-advisor agent file (P-High)**: author `.claude/agents/moai/super-advisor.md` (+ template mirror) with `model: inherit` / `effort: xhigh` FIXED / read-only tools / `permissionMode: plan` / `NOT for:` mutual-exclusion vs auditors; flip CLAUDE.md §4 ceiling 8→10 (live + template); embed E1-E4 escalation doctrine in `agent-common-protocol.md` § Error Recovery Pattern.
- **M2 — manager-design agent file + design phase (P-High)**: author `.claude/agents/moai/manager-design.md` (+ template mirror) with the §04 frontmatter verbatim + H1-H9 contract embedded VERBATIM in body; author `.claude/skills/moai/workflows/design.md` with D1-D5 pipeline; add `plan → design → run` conditional route to `spec-workflow.md` (UI-surfaced SPECs only); annotate `role_profiles.designer` + pencil MCP as absorbed.
- **M3 — No-Haiku 3-tier token policy (Go code, P-High)**: extend `RouteModelFor` 2-arg → 3-arg in `internal/config/model_routing.go`; add `model_routing_profiles.{max,medium,low}` (3 × 12-cell matrices per §2-D) to workflow.yaml + template; add `--model-policy max|medium|low` flag to `moai init`; flip `llm.yaml claude_models.low` haiku→sonnet; add `HaikuResidualRule` lint (HARD gate); substitute haiku→sonnet across `workflow_agents`, `role_profiles`, doctrine files.
- **M4 — doctrine refresh (P-Med)**: refresh `model-policy.md` (§2-B supersede § Model Policy Tiers / § Effort Calibration Matrix per SSOT §06 M4 verbatim; fable enum · v2.1.196 · v2.1.198 reflected; § Inherit-by-Default haiku-exception removed), `agent-authoring.md` (10-agent catalog + new patterns), `agent-patterns.md` (4-loop mapping + 4 rejected alternatives per §06 M4).

Cross-cutting: REQ-AA2-015 (template-first) applies to M1-M4; REQ-AA2-016 (haiku-residual-0 HARD success metric) is the M3 closure gate; REQ-AA2-017 (supersede 2 targets) is plan-phase (this SPEC's own authoring scope).

### §A.4 Tier evidence (L)

- Files affected: ~25-35 (2 new agent files × 2 trees; 1 new workflow skill; 2 Go files modified — model_routing.go + init.go — + tests; 1 new lint rule file; 6-8 doctrine/config files × template mirrors; CLAUDE.md; 2 supersede-target frontmatter flips). Comfortably in Tier L's >15-files band.
- LOC estimate: 800-1500 (Go signature extension + tests ~300-500; agent file bodies ~200-300 each × 2; doctrine refreshes ~200-400; lint rule ~100-150) — Tier L band.
- Constitutional dimension: CLAUDE.md §4 catalog ceiling change + new design workflow phase = constitutional adjacency; combined with the multi-surface scope, Tier L is justified per §06 ("규모 재판정: v1의 Tier S~M에서 Tier L로 상향").
- plan-auditor PASS threshold for Tier L: 0.85.

### §A.5 PRESERVE / EXTEND map

| Surface | Disposition |
|---------|-------------|
| `internal/config/model_routing.go` RouteModelFor accessor + ValidateModelRouting + LoadModelRouting | EXTEND (3-arg signature; preserve fallback semantics + closed-set validation) |
| `workflow.yaml model_routing` (single 12-entry block) | REPLACE with `model_routing_profiles.{max,medium,low}` (3 × 12 entries); old block retired |
| `workflow.yaml workflow_agents` (haiku references) | EXTEND (haiku→sonnet substitution per REQ-AA2-013) |
| `workflow.yaml role_profiles` (researcher=haiku) | EXTEND (haiku→sonnet) |
| `llm.yaml claude_models` (low: haiku) | EXTEND (low: sonnet; haiku key removed) |
| `llm.yaml performance_tier` (unused empty field) | EXTEND (now populated by `moai init --model-policy`) |
| `CLAUDE.md §4` Retained Agents table (8 agents) | EXTEND (ceiling 8→10; add super-advisor + manager-design rows + Selection Decision Tree entries 10/11) |
| `.claude/rules/moai/core/agent-common-protocol.md` § Error Recovery | EXTEND (embed E1-E4 escalation doctrine + cross-reference super-advisor) |
| `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline | EXTEND (add plan→design→run conditional route for UI-surfaced SPECs) |
| `.claude/rules/moai/development/model-policy.md` | EXTEND (§2-B supersede per SSOT §06 M4 verbatim — Effort Calibration Matrix를 §2-B 표로 대체; fable enum · v2.1.196 · v2.1.198 doctrine reflected; § Inherit-by-Default haiku-exception prose removed) |
| `.claude/rules/moai/development/agent-authoring.md` | EXTEND (10-agent catalog + super-advisor/manager-design patterns) |
| `.claude/rules/moai/workflow/team-protocol.md` + `team-pattern-cookbook.md` | EXTEND (researcher haiku→sonnet; designer role_profile absorbed-by annotation) |
| `internal/spec/lint.go` + new lint rule file | EXTEND (add HaikuResidualRule — HARD gate) |
| `internal/cli/init.go` | EXTEND (`--model-policy max|medium|low` flag rename + enum validation) |
| `model-policy.md § Inherit-by-Default` | PRESERVE (`[1m]` workaround rationale); EXTEND (remove haiku-exception prose — obsolete under No-Haiku) |
| `archived-agent-rejection.md §C` per-spawn Agent(general-purpose) pattern | PRESERVE (basis super-advisor promotes; the per-spawn pattern remains for non-catalog specialists) |
| CLAUDE.md §15 CG Avoid list | PRESERVE (mirror source for GLM carve-out; unmodified) |
| SPEC-ADVISOR-RUNG-001 frontmatter | TRANSITION (draft → superseded; add superseded_by) |
| SPEC-MODEL-ROUTING-WIRE-001 frontmatter | TRANSITION (draft → superseded; add superseded_by) |
| Pre-existing uncommitted edits (llm.yaml team_mode, manager-docs.md/manager-git.md frontmatter, internal/statusline/*, pkg/version/version.go, system.yaml) | PRESERVE VERBATIM (NOT in scope; run-phase MUST NOT git-add these) |

## §B Known Issues (filtered, Tier L)

- **B1 Cross-platform Build Tags**: pure Go config + CLI work; no syscall layer. Verify `GOOS=windows GOARCH=amd64 go build ./...` after M3 signature extension.
- **B2 Cross-SPEC conflicts**: SPEC-TOKEN-ROUTING-001 (closed) authored the original 2-arg `RouteModelFor`; the 3-arg extension preserves its semantics (fallback, closed-set validation). SPEC-CC2178-MODEL-POLICY-REPAIR-001 touched model-policy.md — re-read § Baseline-Refill Breaker before M4 to avoid disturbing Sonnet 5 resolution prose. SPEC-SONNET5-1M-TEAM-DISABLE closed the team-mode default flip — do NOT re-flip `team.enabled`.
- **B3 Subagent Boundary Discipline (C-HRA-008)**: super-advisor and manager-design are subagents — they MUST NOT invoke AskUserQuestion. manager-design's `/design-login` / `/design-sync` invocations are user-only TUI commands; the agent guides, never invokes. `grep -rn 'AskUserQuestion' .claude/agents/moai/{super-advisor,manager-design}.md` MUST return 0 matches.
- **B4 Frontmatter Canonical Schema**: both new agent files MUST follow the 12-canonical-field schema. `effort: xhigh` is quoted-or-unquoted per the existing agent file convention (verify against manager-spec.md).
- **B5 CI 3-tier awareness**: spec-lint, golangci-lint, Test (per OS) each fail independently. M3's HaikuResidualRule is a spec-lint gate; the Go test suite must cover the 3×12 routing matrix; golangci-lint must not flag the new accessor.
- **B6 spec-lint Heading convention**: spec.md § Out of Scope uses `### Out of Scope — <topic>` H3 sub-headings (verified in spec.md).
- **B8 Working-tree hygiene**: PRE-EXISTING uncommitted edits (llm.yaml team_mode, manager-docs.md/manager-git.md, internal/statusline/*, pkg/version/version.go, system.yaml) are NOT in scope. Run-phase commits MUST `git add` only v2-owned paths (specific-path commits; NEVER blanket `git add -A`).
- **B10 Scope discipline**: Workflow-Reflex siblings (HARNESS-RATCHET-REWIRE, LOOP-VERDICT-CONTRACT, CADENCE-BRIDGE, OBSERVE-HYGIENE) remain active — do not touch their surfaces.
- **B11 DesignSync MCP tool availability (M2-specific)**: the DesignSync server is NOT registered in `.mcp.json` at plan-phase. M2 run-phase MUST verify operational availability before D2; absence triggers H1 blocker-report path (graceful degradation — the agent file is still authored, the workflow skill is still authored, but D2 execution waits on the tool).
- **B12 CHANGELOG emission (sync-phase)**: manager-docs reads all implementation files before drafting CHANGELOG; the v2 architecture is substantial enough that the CHANGELOG entry will reference all 4 milestones.

## §C Pre-flight checklist

```bash
git branch --show-current && git rev-parse HEAD
git status --porcelain | head -20                                       # confirm pre-existing edits present, do NOT touch
grep -rn "RouteModelFor" --include='*.go' internal/ | grep -v _test     # re-verify zero external call sites
grep -rn "haiku" .claude/agents/moai/ .moai/config/sections/            # baseline haiku count (pre-M3)
grep -n "Retained Agents" CLAUDE.md                                     # confirm "8 retained agents" (pre-M1)
ls .claude/agents/moai/                                                 # confirm no super-advisor.md / manager-design.md
go build ./... && GOOS=windows GOARCH=amd64 go build ./...              # cross-platform baseline
golangci-lint run --timeout=2m 2>&1 | tail -5                           # lint baseline
go test ./internal/config/... ./internal/cli/... 2>&1 | tail -5         # test baseline
cat .mcp.json | grep -i design || echo "DesignSync not registered"      # M2 precondition check
```

## §D Constraints + open decisions

Constraints: see spec.md §Constraints (Design SSOT authority; `[1m]` workaround preserved; haiku-residual-0 HARD; DesignSync availability; subagent boundary; pre-existing edits preserved; template-first; GEARS + era V3R6; Implementation Kickoff Approval mandatory).

**D1 — super-advisor Opus injection mechanism (M1).** The SSOT (§05 + §2-B) specifies super-advisor gets Opus at max/medium tiers, Sonnet xhigh at low tier. The injection mechanism is per-spawn runtime arg (`model: opus` or `model: sonnet` resolved by `RouteModelFor(specTier="L", phase="<current>", perfTier=<user-selected>)`), NOT a frontmatter pin (frontmatter `model: inherit` preserves `[1m]`-safety per Constraint #2). RECOMMENDED: document the injection as an orchestrator-side responsibility (the orchestrator reads `moai route` output and passes the model arg to `Agent(general-purpose, model: <routed>)`). Alternative: a dedicated orchestrator-side helper — rejected for Tier L (orchestrator-layer change out of scope).

**D2 — manager-design `effort: xhigh` FIXED across all tiers (M2).** Per §04 codeblock + §2-B manager-design row: `effort: xhigh` is FIXED in frontmatter (the only frontmatter-fixed effort in the catalog). Rationale (§2-B note): "전 티어 xhigh 고정(frontmatter) — 핸드오프 충실도·드리프트 판정·주석→요구 변환은 심층 추론". RECOMMENDED: frontmatter `effort: xhigh` literal. This is the ONE exception to the "tier-routing injects effort" rule; document it explicitly in the agent body.

**D3 — HaikuResidualRule scope (M3).** The lint rule (REQ-AA2-012) covers: agent frontmatter, `claude_models`, `model_routing_profiles` cells, `workflow_agents`, `role_profiles`, `validRoutingModels` Go map. RECOMMENDED: the rule is a HARD gate (NOT skip-able via `lint.skip`). Alternative: advisory warning — REJECTED (§08 row 1 makes haiku-residual-0 a HARD success metric; advisory contradicts the metric).

**D4 — `moai init` flag migration (M3).** The legacy `--high/medium/low` flag names are reused with new semantics (per §2-E: "개명 + 의미 재정의"). RECOMMENDED: accept `--model-policy max|medium|low` (new canonical name) AND keep `--high/medium/low` as deprecated aliases (one-cycle backward compat) with a stderr warning. Alternative: hard rename without alias — REJECTED (breaks existing users' init scripts).

**D5 — DesignSync MCP server registration (M2 run-phase precondition).** The `.mcp.json` does NOT register DesignSync at plan-phase. RECOMMENDED: M2 run-phase begins by verifying DesignSync availability; if absent, the agent file + workflow skill are still authored (they describe the contract per §04), but D2-D5 live execution is gated on the tool. The user registers DesignSync separately (it requires Claude Code v2.1.181+ and a Pro+ Claude Design account per §07 Risk HIGH).

**D6 — legacy `model_routing` key disposition (M3, added plan-audit iter-2 D4).** The legacy single `model_routing:` block (workflow.yaml, S/M/L × plan/run/sync/mx = 12 cells) is REPLACED by `model_routing_profiles.{max,medium,low}` (3 × 12 cells) per REQ-AA2-009. Open decision: is the flat `model_routing:` key RETAINED as a backward-compat alias or REMOVED? RECOMMENDED: RETAIN the legacy `model_routing:` block as a `medium` profile alias for one backward-compat cycle (with a deprecation note in workflow.yaml), because `model_routing_profiles.medium` is the default performance tier and existing user configs reference the flat `model_routing:` key. This resolves the progress.md §E.1 open-decision that previously collided with this section's D-numbering (iter-2 D4 numbering-unification fix).

## §E Self-Verification (run-phase deliverables)

Per manager-develop-prompt-template.md §E (Tier L full form), vci 5-section format each:

- **E1**: AC matrix (acceptance.md §D) with verbatim outputs for all 17 REQs / their ACs.
- **E2**: cross-platform builds exit 0 (`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`); `make build` exits 0 after template edits.
- **E3**: `go test -cover ./internal/config/... ./internal/cli/... ./internal/spec/...` — coverage ≥85% per package; golden table-driven tests for the 3×12 routing matrix (36 entries per REQ-AA2-009).
- **E4**: subagent-boundary grep — `grep -rn 'AskUserQuestion\|mcp__askuser' .claude/agents/moai/{super-advisor,manager-design}.md` returns 0 matches (C-HRA-008 family).
- **E5**: lint — `golangci-lint run --timeout=2m` clean (NEW vs baseline distinguished); `moai spec lint` clean on all SPECs; HaikuResidualRule green (haiku grep returns 0).
- **E6**: commit SHAs + push state; commit subjects follow `feat(SPEC-AGENT-ARCH-V2-001): M{N} ...` per the Status Transition Ownership Matrix; pre-existing uncommitted edits NOT in commits (specific-path `git add`).
- **E7**: blocker report if any D1-D5 decision was not pre-resolved, OR if DesignSync MCP is unavailable at M2 run-phase (B11).

## §F Milestones (priority-ordered; no time estimates)

| Milestone | Scope | REQs | Exit criterion |
|-----------|-------|------|----------------|
| M1 — super-advisor agent file + ceiling + escalation doctrine | super-advisor.md ×2 trees; CLAUDE.md §4 ceiling 8→10 ×2; agent-common-protocol.md E1-E4 doctrine | REQ-AA2-001, 002, 003 | AC-AA2-001..003 PASS; ceiling grep returns "10 retained agents"; super-advisor agent file exists with correct frontmatter |
| M2 — manager-design agent file + design phase + role_profile absorption | manager-design.md ×2 trees with H1-H9 verbatim; design.md workflow skill; spec-workflow.md plan→design→run; role_profiles.designer + pencil MCP absorption annotation | REQ-AA2-004, 005, 006, 007 | AC-AA2-004..007 PASS; manager-design agent file exists; H1-H9 verbatim in body; design.md skill exists |
| M3 — No-Haiku 3-tier token policy (Go code) | RouteModelFor 3-arg; model_routing_profiles.{max,medium,low} 3 matrices; moai init --model-policy; llm.yaml claude_models.low haiku→sonnet; HaikuResidualRule lint; workflow_agents + role_profiles haiku→sonnet | REQ-AA2-008..013, 016 | AC-AA2-008..013 PASS; haiku grep returns 0 (HARD SUCCESS METRIC); 3×12 golden tests PASS; moai init --model-policy enum validation PASS |
| M4 — doctrine refresh | model-policy.md (§2-B supersede per SSOT §06 M4 verbatim; fable enum · v2.1.196 · v2.1.198 reflected; Inherit-by-Default haiku-exception removed); agent-authoring.md (10-agent catalog); agent-patterns.md (4-loop mapping + 4 rejected alternatives) | REQ-AA2-014 | AC-AA2-014 PASS; doctrine files refreshed ×2 trees |
| Cross-cutting | template-first application (M1-M4); supersede flips (plan-phase, already done in this SPEC's scope) | REQ-AA2-015, 017 | AC-AA2-015, 017 PASS; both supersede targets carry status: superseded + superseded_by |

Dependency note: M1 and M2 are independent of each other and may execute in parallel (different surfaces). M3 blocks on neither but is the largest single milestone (Go signature extension + 3 matrices + lint rule). M4 blocks on M1 + M3 (doctrine reflects the new catalog + the No-Haiku policy). The whole SPEC sequences post Implementation Kickoff Approval (Constraint #9).

## §G Anti-Patterns (do NOT)

- Re-deriving or inventing architecture not in the design SSOT (`.moai/reports/agent-architecture-redesign-v2-20260709.html`). The SSOT is the architecture authority.
- Pinning `model:` in super-advisor or manager-design FRONTMATTER — both use `model: inherit` (tier-routing injects at spawn); frontmatter pins trigger the `[1m]` bug the policy exists to avoid.
- Making `effort:` configurable for manager-design — its `effort: xhigh` is FIXED across all tiers (the ONE frontmatter-fixed effort in the catalog per §04 codeblock + §2-B).
- Adding `--model-policy` as a new flag without aliasing the legacy `--high/medium/low` (breaks existing init scripts per D4).
- Making HaikuResidualRule advisory or skip-able — §08 row 1 + REQ-AA2-012/016 make it a HARD gate.
- Invoking `/design-login` or `/design-sync` from inside manager-design — they are user-only TUI commands; the agent guides, never invokes.
- Allowing super-advisor to issue binding verdicts — its prescription is non-binding; binding PASS/FAIL remains auditor-owned (REQ-AA2-001 `NOT for:` clause).
- Blanket `git add` — pre-existing uncommitted edits (llm.yaml team_mode, manager-docs.md/manager-git.md frontmatter, internal/statusline/*, pkg/version/version.go, system.yaml) MUST NOT enter v2 commits (B8).
- Touching the 4 remaining Workflow-Reflex siblings — they are active SPECs, not supersede targets.
- Skipping Implementation Kickoff Approval — the plan→run HUMAN GATE is mandatory regardless of Tier L complexity or plan-audit score (Constraint #9).

## §H Cross-References

- spec.md (SSOT), acceptance.md (AC matrix), design.md (architecture detail + H1-H9 verbatim), research.md (file:line baselines), progress.md (§E skeleton).
- Design authority: `.moai/reports/agent-architecture-redesign-v2-20260709.html` (§01-§08).
- Extension points: `internal/config/model_routing.go:89`; `internal/cli/init.go`; `internal/spec/lint.go`.
- `.claude/rules/moai/development/model-policy.md` § Inherit-by-Default (M3/M4 baseline — read before editing).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix (supersede flip pattern).
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline (M2 plan→design→run extension point).
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` header (Implementation Kickoff Approval mandatory-restoration policy).
- Superseded SPECs: SPEC-ADVISOR-RUNG-001, SPEC-MODEL-ROUTING-WIRE-001 (frontmatter flips in this SPEC's scope).
- Workflow-Reflex Epic remaining: SPEC-HARNESS-RATCHET-REWIRE-001, SPEC-LOOP-VERDICT-CONTRACT-001, SPEC-CADENCE-BRIDGE-001, SPEC-OBSERVE-HYGIENE-001.
