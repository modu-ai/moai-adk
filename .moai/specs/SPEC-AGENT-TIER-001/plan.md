---
id: SPEC-AGENT-TIER-001
title: "Implementation plan — three-tier subagent model-effort configuration"
version: "0.1.0"
created: 2026-10-01
---

# SPEC-AGENT-TIER-001 — Implementation Plan

Methodology: **tdd** (tier constants are new code with a RED-first test). Tier M — artifacts:
spec.md + plan.md + acceptance.md (+ progress.md). plan-auditor PASS threshold 0.80.

## §A Context

The operator's 2026-10-01 directive introduces a 3-tier subagent model·effort configuration
(max/medium/low) grounded in the Terminal-Bench 4.0 chart, and replaces the t1368 audit pins
with operator-fixed values. spec.md §A carries the full context, the partial supersession of
SPEC-AGENT-MODEL-INHERIT-001, and the Q2 terminology decision. The four implementation scopes:
(1) tier constants, (2) audit pin defaults across Go/template/resolvers/web, (3) the class→tier
profile wiring, (4) the moai web tier widget.

## §B Known Issues

1. **Pin-location drift.** The delegation brief placed the t1368 pins at
   `defaults.go:1202`; the measured location in this tree is the `Audit:` block at
   `defaults.go:1218-1238` (claude literals at `:1221-1222`). Re-grep at run entry before the
   first edit — line numbers move with every absorb.
2. **Constant-naming collision hazard.** The lane's suggested suffixes (Max/High/Medium)
   are effort-keyed and would alias tier-medium to a `...TierHigh` constant; spec.md
   REQ-TIER-002 records the tier-keyed deviation. Do not "normalize" back to effort-keyed
   suffixes during run.
3. **glm-5.3 grant surface.** `glm-5.3` is reachable on the Anthropic-compatible endpoint but
   "not granted on the native paas surface for every account" (defaults.go comment). Pinning
   glm to {glm-5.3, max} is safe only because the GLM side fails open to inconclusive
   (workflow.yaml comment); the SPEC relies on that fail-open — do not add hard-error handling.
4. **Template comment residue.** The template `workflow.yaml` audit comment already carries an
   internal date. Rewrite the comment's PIN VALUES accurately; do not add new SPEC IDs, card
   IDs, or dates (template neutrality, REQ-TIER-014).
5. **`max` effort routing.** The settings field `effortLevel` rejects `max`; a resolved max
   must ride the launcher launch-argument path (coding-standards.md compatibility table,
   `resolveLaunchEffort` precedent). Relevant to M2's injection channel decision.

## §C Pre-flight

- Re-measure the pin block (`grep -n -A 3 "Claude: ModelEffort{" internal/config/defaults.go`)
  and the template audit block immediately before the first edit.
- Baseline the affected packages before changes:
  `go test ./internal/config/... ./internal/cli/... ./internal/web/...` (record pass counts in
  progress.md §E.2 — needed to prove AC-TIER-011/012 regressions are genuinely preserved).
- Confirm no parallel session on this SPEC (`moai session list --json --filter-spec=SPEC-AGENT-TIER-001`).
- Read the full template audit comment block (`workflow.yaml:98-121`) before rewriting it.

## §D Constraints

- Template-First: any template edit starts in `internal/template/templates/`, then `make build`;
  tracked mirror byte-identical; `make embed-check` clean.
- Hardcoding prevention: env names in `envkeys.go`; model ids / effort values as named
  constants in `internal/config` (the `DefaultCodexAuditModel` / `DefaultGLM*` precedent);
  thresholds and pin values single-source.
- English code comments; Conventional Commits; no time estimates anywhere.
- Local verification is affected-packages only (AGENTS.local.md §4) — no full-suite local runs.
- Lane env scrubbing: run env-reading guard tests as one compound `unset … && <command>` call.
- 16-programming-language neutrality in any template content.

## §E Self-Verification

Run-phase evidence lands in progress.md §E.2 per the verbatim-output contract: every AC's
command + observed output + exit code, the RED-now cells from acceptance.md's Evidence Ledger
executed and pinned to the run-start SHA, and the affected-package batch
(`go vet`, `golangci-lint run`, `go test` on the three packages, `make build`,
`make agents-emit-check`). No AC is reported PASS without its observed output.

## §F Milestones

Ordered by decision-reversibility: M1 carries the data-model decisions (constant names, pin
values) that everything else consumes; M2 carries the interface wiring with the most design
freedom; M3 is the user-facing surface whose figures are already operator-pinned.

### M1 — Tier constants + audit pin defaults (Priority: High)

TDD RED first: author `TestClaudeTierConstants` in `internal/config` (and the pin tests)
BEFORE any constant exists; execute it via `go test ./internal/config/ -run
'^TestClaudeTierConstants$'` and observe the compile-failure RED and record it verbatim in
progress.md §E.2 (AC-TIER-001's RED-now cell). Then:

- `internal/config/defaults.go`: add `DefaultClaudeTierMax` / `DefaultClaudeTierMedium` /
  `DefaultClaudeTierLow` (pair-carrying; decomposed `...Model`/`...Effort` const components are
  acceptable where Go's const rules require them) beside the `DefaultGLM*` family, plus
  `DefaultClaudeTierMaxFallback` = {claude-opus-5-5, xhigh} with the chart grounding in its doc
  comment. Flip the Audit block: claude effort `medium`→`high`; add the GLM pin {glm-5.3, max}
  (use `DefaultGLM53`, NOT `DefaultGLMHigh` — the flash variant is the slot default, the pin
  targets full glm-5.3).
- `internal/config/closed_sets.go`: add named constants for the claude pin model/effort where
  the `DefaultCodexAuditModel` precedent applies; keep codex {gpt-6.1-sol, high} unchanged.
- `internal/cli/mcp_claude.go` (`resolveClaudeAuditModelEffort` terminal fallback) and
  `internal/cli/mcp_glm.go` (GLM audit pin path): reference the new constants — no inline
  literals.
- `internal/template/templates/.moai/config/sections/workflow.yaml`: audit block claude
  `medium`→`high`; glm `""` → {glm-5.3, max}; comment pin sentence updated (values only).
  Then `make build`.
- Tests (RED-first where new): tier constants pairs (AC-TIER-001/008), pin byte-exactness
  (AC-TIER-002), template mirror parity (AC-TIER-003), resolver fallbacks (AC-TIER-004).

### M2 — Profile matrix wiring (Priority: High)

- Config surface for the tier table + class→tier assignment (default candidate: a
  `workflow.agent_tiers` block in workflow.yaml; alternative: an llm.yaml key — decide at run
  start and record the choice + reason in progress.md §E.1; the loader lives in
  `internal/config` with `Validate()` against the closed set {max, medium, low}).
- Default assignment per REQ-TIER-008; audit surfaces excluded and asserted by test
  (AC-TIER-007); unknown-token rejection test with a fixture (AC-TIER-006).
- **Open design decision (within REQ-TIER-010's constraints): the spawn-time injection
  channel.** Candidates: (a) launcher/session-env injection of the resolved pair via the
  `CLAUDE_CODE_SUBAGENT_MODEL` chain — recommended default: rides Claude Code's native
  resolution chain (spawn model → frontmatter → env → main conversation), requires zero
  doctrine change, and preserves the observer's declared/inherit semantics with tier-derived
  spawns classifying as declared; (b) delegation-layer spawn-argument injection; (c) an agent
  frontmatter `tier:` key (rejected for this SPEC — it is the per-agent-file channel
  SPEC-AGENT-MODEL-INHERIT-001 removed and touches the C1/C2/C3 emit pipeline). Record the
  chosen channel + reason in progress.md at M2 start; any choice outside (a)-(c) returns a
  blocker to the orchestrator.
- Terminology separation checks (AC-TIER-009): effort-vocabulary surfaces byte-unchanged.

### M3 — moai web tier widget (Priority: Medium)

- `internal/web` (templ surfaces — settings/screens/fieldsets where the model·effort controls
  and the audit-pin fields live): render the three tiers with the chart figures (70.6% · ~$11 /
  45% · ~$2.3 / 29% · ~$0.8), selection closed to {max, medium, low}; audit-pin fields render
  the new defaults consistently (REQ-TIER-005).
- Render tests asserting the figures and the closed set (AC-TIER-010); run `make build` after
  any `.templ` edit (templ codegen).

### M2→M3 dependency note

M3 consumes M2's config surface for selection values; M1's constants alone suffice for M3's
figures. If M2's channel decision stalls, M3 may proceed on M1 + the closed set and rejoin.

## §G Anti-Patterns

- No time estimates; priority labels and phase ordering only.
- No inline pin/tier literals outside the declared single-source locations (REQ-TIER-013).
- No automatic failover machinery (§C of spec.md).
- No edits to `glm_task` defaults, the GLM effort overlay, or main-session model policy.
- No `model:`/`effort:` (or `tier:`) frontmatter added to agent files; no doctrine rewrites
  beyond what REQ-TIER-010 states.
- Do not trust remembered line numbers — re-grep at run entry (§B.1).
- Do not run the full test suite locally (affected packages only).

## §H Cross-References

- spec.md (requirements + evidence base) · acceptance.md (AC matrix + RED-now Evidence Ledger)
- SPEC-AGENT-MODEL-INHERIT-001 (bounded partial supersession) · SPEC-MODEL-MATRIX-UPDATE-001
  (t1368 pins replaced here) · SPEC-V3R6-AUDIT-MODEL-PIN-001 (REQ-AMP-006/008 semantics)
- Grounding: `.moai/reports/agent-tier-design-20261001.md` (untracked, primary checkout)
