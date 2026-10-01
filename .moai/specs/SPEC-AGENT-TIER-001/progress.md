# SPEC-AGENT-TIER-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).

- 14 GEARS requirements (`REQ-TIER-001..014`); 15 acceptance criteria (`AC-TIER-001..015`),
  13 release-blocking + 2 regression-guard.
- Methodology: tdd (tier constants are new code with a RED-first test,
  `TestClaudeTierConstants`).
- Milestones: M1 tier constants + audit pin defaults (High) → M2 profile matrix wiring (High)
  → M3 moai web tier widget (Medium). Ordered by decision-reversibility; M2 carries the one
  open in-constraint design decision (spawn-time injection channel; candidates + recommended
  default recorded in plan.md §F M2, to be decided and logged at M2 start).
- Operator decisions carried verbatim: audit pins are operator-fixed defaults (claude
  {claude-opus-5-5, high} · codex {gpt-6.1-sol, high} · glm {glm-5.3, max}); audit surfaces
  excluded from the tier matrix; tier pairs max {sonnet-5-5, max} / medium {sonnet-5-5, high} /
  low {sonnet-5-5, medium}; fallback {claude-opus-5-5, xhigh} recorded, not automated.
- Q2 terminology: adopted as recommended (tier tokens are configuration-key names, not effort
  values). One naming deviation recorded at spec.md REQ-TIER-002: constants keyed by TIER token
  (DefaultClaudeTierMax/Medium/Low), not by effort-slot suffix — the suggested
  Max/High/Medium suffixes would alias tier-medium to a `...TierHigh` constant and recreate the
  exact ambiguity Q2 prevents. Values are byte-exact to the operator table; no value deviation.
- Tree findings vs the delegation brief: pin block measured at `defaults.go:1218-1238` (brief
  said `:1202`); claude pin literals at `:1221-1222`; GLM pin currently absent from the Go
  default (ships empty in the template); the per-agent profile matrix is retired by
  SPEC-AGENT-MODEL-INHERIT-001, so the tier axis is a bounded partial supersession (spec.md
  §A.1).
- RED-now observations: 5 cells observed on the pre-implementation tree at
  `f130aa041bb90c81b235029afffa806a65e88480` (acceptance.md Evidence Ledger, EVID-TIER-A..E);
  4 planned run-start cells (R1..R4) to be executed and pinned before the first M1 edit.
- Grounding: `.moai/reports/agent-tier-design-20261001.md` (untracked, primary checkout);
  SPEC-critical figures carried in spec.md §E.

plan_status: audit-ready
plan_complete_at: 2026-10-01

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

### Mode Selection

Input parameters:

- tier: M · scope: ~8 files (internal/config, internal/cli, internal/web, template
  workflow.yaml + tests)
- domain count: 4 (Go config / Go CLI / Go web+templ / YAML template) — coding-heavy
- file language mix: Go + YAML + templ
- concurrency benefit: LOW (milestones are dependency-ordered M1→M2→M3; Anthropic
  coding-task parallelism caveat)
- Agent Teams prereqs: not requested

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-file semantic change |
| serial | **YES** | coding-heavy, dependency-ordered milestones; one implementation spawn carries M1→M2→M3 |
| fanout | no | coding-heavy work — parallelism caveat; no independent research lanes |
| sweep | no | semantic new-code work, ~8 files, not mechanical-uniform |
| agent-team | no | not operator-requested |

Decision: serial

Justification: coding-heavy single-tree work with strict milestone dependencies (M1's
constants feed M2's wiring and M3's figures), so the Anthropic coding-task parallelism
caveat puts this on the sequential path. One implementation delegate keeps exactly one
writer on the tree. The delegate is dispatched as a general-purpose worker performing the
manager-develop role — the typed manager-develop spawn auto-isolates to its own L1 tree and
refuses card-tree writes (card t1318 lesson), and the card worktree is the required landing
site. Boundary Case: none hit — scope and domain counts sit below the fanout thresholds.

### Implementation Kickoff decision record

- Gate form: **operator direct answer** (keep-set class: operator-held decisions). The
  operator answered the Kickoff panel directly, relayed by the leader (2026-10-01).
- Audit cross: plan-auditor PASS 0.91 ≥ 0.80 (Tier M) — report
  `.moai/reports/t1391/plan-audit.md` (auditor model glm-5.3-flash; GLM lanes cannot emit
  Opus audits — lane lesson, checked against the report's first line);
  `plan_status: audit-ready` recorded in §E.1; no open blocker.
- Operator decisions recorded verbatim:
  1. Q2 approved: tier tokens are configuration-key names; the effort field keeps the
     chart-axis vocabulary (low/medium/high/xhigh/max).
  2. Constant-naming deviation approved: `DefaultClaudeTierMax` / `DefaultClaudeTierMedium`
     / `DefaultClaudeTierLow`.
  3. Audit pins fixed: claude {claude-opus-5-5, high} · codex {gpt-6.1-sol, high} · glm
     {glm-5.3, max}; the 3-tier matrix applies to the general profile surfaces only — audit
     surfaces excluded.
  4. Minor amendments D1 + D2 approved and ordered applied BEFORE run entry ("D1·D2 반영 후
     run 진입", Kickoff record first).
- D1 (selector anchoring): the unanchored `-run TestClaudeTierConstants` citations in
  acceptance.md (AC-TIER-001 body, Evidence Ledger R1) and plan.md §F M1 become
  `go test ./internal/config/ -run '^TestClaudeTierConstants$'` — unanchored selectors are
  substring matches and can sweep a superset while reporting ok (t1371 root cause);
  anchoring pins the swept set to the exact test.
- D2 (GLM audit surface single-sourcing): new AC-TIER-015 (REQ-TIER-006) —
  `internal/cli/mcp_glm.go`'s audit defaults must resolve from the shared config constants
  ({DefaultGLM53, max}) with no inline model/effort literals; RED-now cell EVID-TIER-F
  observed on the run-start tree; flipped by M1.
- Post-amendment artifact hashes (run-entry basis; sha256, measured after D1/D2 land,
  before the first run-phase spawn):

  - spec.md: 9d94ce733a4c10f6e51fa165a1f598f54b49605ee94bbfcf7825ade031636d88
  - plan.md: ff5920332a6a9a1168922b6aebebd4d598d610381c4d0011a2e81b7b19d22c4e
  - acceptance.md: f3c0728bdd576bc6b3d1b1852741b432b345adddbcc0cf4efbe145fccb57ddce

- Pre-amendment verdict state: the plan-audit verdict judged the pre-D1/D2 artifact set;
  the deltas are exactly D1+D2 (operator-ordered), nothing else moved.
- Progression mode: operator-driven lane cadence — no `/moai goal` armed (the lane's Stop
  evaluator is not the card's judge; the leader advances the card on evidence per
  kanban-dispatch).
