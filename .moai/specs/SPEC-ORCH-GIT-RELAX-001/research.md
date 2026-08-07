# research.md — SPEC-ORCH-GIT-RELAX-001

> Research-first evidence incorporation. The canonical evidence base is `.moai/reports/agent-skill-hook-redesign-evidence-20260804.md` (Agent(Explore)×4 + Agent(general-purpose)×1 parallel fan-out, 2026-08-04, checkout `c7b61777b`). This file incorporates verbatim the §4 change-surface catalog and §5 web-research findings, plus 12 verified URLs.

---

## §1. The structural paradox (evidence §0)

The current doctrine forces the MOST context-sensitive operation (primary-branch changes, multi-worktree discrimination) onto the agent with the LEAST context (manager-git, an isolated subagent holding only a session-start git snapshot). PR #1338 handling is the incident that proved the paradox:

- manager-git, holding a stale snapshot, attempted primary→main restore.
- The stale snapshot did not reflect a concurrent session's worktree state.
- manager-git cross-swapped the concurrent session's worktree branch to main.
- The orchestrator's DIRECT recovery — with live HEAD, the active-worktree registry, the concurrent-session state, and the full transcript — was correct.

Root cause: structural, not operator error. The doctrine assigned the wrong class of work to the wrong actor. Forcing state-sensitive ops onto an isolated subagent is the anti-pattern; the relaxation inverts it.

---

## §2. Context-sensitivity classification (evidence §1.3)

| Agent | Classification | Rationale |
|-------|----------------|-----------|
| **manager-git** | **context-SENSITIVE (vulnerable)** | branch-state op + multi-worktree discrimination depend on live git state; isolated subagent context holds only a stale session-start snapshot → stale state action risk. The incident site. |
| manager-develop, manager-docs, manager-spec, manager-design | context-INSENSITIVE | self-contained prompts (file authoring); isolated context is an asset. Delegation appropriate. |
| plan-auditor, sync-auditor | context-INSENSITIVE | evaluate artifacts (documents/code); independent judgment benefits from isolation. Delegation appropriate. |
| super-advisor | context-INSENSITIVE | returns non-binding prescriptions only; no state mutation. Delegation appropriate. |
| e2e-tester, builder-harness | context-INSENSITIVE | self-contained tasks. Delegation appropriate. |

**Conclusion (evidence §1.3)**: manager-git is the ONLY context-SENSITIVE-vulnerable retained agent. The relaxation's justification is grounded in this asymmetry — the catalog has exactly one agent whose isolation is a structural liability rather than an asset.

---

## §3. The 12-location change-surface catalog (evidence §4, verbatim)

The relaxation must update exactly these doctrine locations. File:line are as of checkout `c7b61777b` (evidence report baseline); run-phase MUST re-verify each location before editing (a newer commit may shift lines).

| # | Location | Current rule | After relaxation |
|---|----------|--------------|------------------|
| 1 | `CLAUDE.md:75` (selection tree row 6) | "Tier L OR `--pr` → manager-git" | Tier S/M default orchestrator-direct; Tier L / `--pr` only → manager-git |
| 2 | `CLAUDE.md:90` (catalog table row) | "PR creation per Tier + Late-Branch" | Refresh description: "Tier L / `--pr` PR + Late-Branch; Tier S/M orchestrator-direct" |
| 3 | `.moai/docs/git-local-workflow-doctrine.md:149` (§23.9, HARD) | "모든 tier (S/M/L) PR 경유, manager-git 담당" (HARD) | Tier S/M orchestrator-direct; Tier L / `--pr` → manager-git |
| 4 | `.moai/docs/git-local-workflow-doctrine.md:153-156` (Tier table) | S/M/L rows all "manager-git" | S/M rows → orchestrator-direct; L row stays manager-git |
| 5 | `.moai/docs/git-local-workflow-doctrine.md:160` (REQ-ATR-020) | "모든 tier PR 생성은 manager-git 책임" | Carve-out: Tier S/M exempted; Tier L / `--pr` retained |
| 6 | `.moai/docs/git-local-workflow-doctrine.md:162` (Late-Branch 4-Phase) | Tier L 4-Phase | UNCHANGED (Tier L only — stays) |
| 7 | `.claude/agents/moai/manager-git.md:5` (invocation gate) | "push + PR is always delegated to manager-git" | "Tier L / `--pr` delegated; Tier S/M orchestrator-direct" |
| 8 | `.claude/agents/moai/manager-git.md:88-129` (Late-Branch body) | Phases A–D | UNCHANGED (Tier L only — stays) |
| 9 | `.moai/config/sections/delegation.yaml:74` | "manager-git is Tier-L / --pr conditional" | Refresh note |
| 10 | `.claude/rules/moai/workflow/main-checkout-branch-guard.md:104,124` | exemption = `AgentType=="manager-git"` OR env | Document orchestrator-direct Tier S/M uses env path; Go unchanged expected |
| 11 | `.claude/rules/moai/core/agent-common-protocol.md` §Pre-Spawn/Pre-Edit | manager-git identity exemption context | Add orchestrator-direct Tier S/M context (env-path + pre-session sync check) |
| 12 | `internal/hook/branch_guard.go:140-154` (`isExemptAgent`) | env OR identity exemption | VERIFY env path admits orchestrator-direct Tier S/M with no Go change; minimal change only if verification fails |
| 13 | `.claude/rules/moai/workflow/spec-workflow.md:26` (Route B description — **iter-2 addition, template-mirrored**) | "Route B — PR route (Tier L OR `--pr`): `manager-git` creates a feature branch and opens a PR per phase" | Qualify Route B actor for Tier S/M (per `repo-local-pr-policy.md` ALL-tier Route B override + this SPEC's relaxation). Mirror to `internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md`. |
| 14 | `CLAUDE.md:131` (§5 Agent Chain — **iter-2 enumeration**) | `[optional Tier L OR --pr] PR (manager-git)` | **NO CHANGE** — 3rd CLAUDE.md site (catalog at #1/#2 listed only lines 75, 90), already carve-out-scoped (`[optional Tier L OR --pr]` prefix). Enumerated for census completeness. |

**Key (evidence §4 + iter-2 audit)**: change-surface is 12 edit locations from evidence §4's original census (11 doctrine + 1 Go verification site) + 1 iter-2 edit addition (`spec-workflow.md` Route B, location #13, missed by evidence §4) + 1 iter-2 no-change enumeration (`CLAUDE.md:131`, location #14). 13 edit locations + 1 verified no-change. The Go site is expected to require NO change because the env-var branch already admits the op before the identity branch.

---

## §4. Quantitative grounding (evidence §5.1)

Multi-agent orchestration carries measurable overhead, and the cost is NOT linear in capability:

- **Multi-agent invocation ≈ 15× tokens vs single chat** (Anthropic multi-agent guidance). Each subagent spawn re-pays the system prompt + rules + skill prefix.
- **Agent-team plan-mode ≈ 7× tokens** (Dive-into-Claude-Code, arXiv:2604.14228 — companion repo `github.com/VILA-Lab/Dive-into-Claude-Code`).
- **Coordination overhead — not model intelligence — is the dominant constraint on scaling** (arXiv:2512.08296). Below a single-agent baseline of ~45% accuracy, adding agents produces **negative returns** (β = −0.408). Sequential state-coupled tasks under multi-agent delegation degrade **−39% to −70%** (PlanCraft benchmark). Parallel decomposable tasks gain **+80.9%** (Finance Agent benchmark).
- **LangChain 4-pattern survey** (Subagents / Skills / Handoffs / Routers): "start with a single agent, add tools first, only add agents when a clear limit is reached."

**Implication for this SPEC**: Tier S/M push+PR is a **sequential state-coupled** op (the orchestrator's run-phase state feeds directly into the PR-op; delegation adds a stale-snapshot hop that breaks the coupling). The arXiv:2512.08296 −39% to −70% degradation range is the quantitative prediction of the PR-#1338 incident class. Tier L release / multi-step merge, by contrast, is **decomposable + independently verifiable** (CI gates, release notes, merge method config) — the +80.9% parallel-decomposable regime where delegation earns its overhead.

---

## §5. The 7 delegation-vs-direct principles (evidence §5.2)

These are the gating principles for the delegation split. For each op class, all 7 principles must be checked.

| # | Principle | Tier S/M push+PR | Tier L release / multi-merge |
|---|-----------|-----------------|------------------------------|
| P1 | **Context-sensitivity** — op depends on orchestrator-only state (HEAD, worktree count, transcript) → direct | FAIL (state-sensitive) → direct | PASS (config-driven, state-light) → delegate |
| P2 | **Decomposability** — sequential state-coupled ops must NOT be decomposed | FAIL (one-shot state op) → direct | PASS (multi-phase, CI-gated) → delegate |
| P3 | **45% baseline** — if single actor is already reliable, delegation is net-negative | FAIL (orchestrator has full state; delegating loses it) → direct | PASS (manager-git has the release playbook) → delegate |
| P4 | **20% coordination overhead** — if delegation overhead > 20% of the task, direct | FAIL (subagent spawn ≫ 20% of a 3-line `gh pr create`) → direct | PASS (release is multi-hour; spawn overhead is negligible) → delegate |
| P5 | **Verification asymmetry** — if verification requires orchestrator context, direct; if mechanical, delegate | FAIL (verifying correct branch requires live HEAD) → direct | PASS (CI gates verify mechanically) → delegate |
| P6 | **Reversibility-weighted risk** — shared/irreversible ops → full-context actor | FAIL (push is shared-tree, semi-irreversible) → direct | PASS (release has rollback playbook) → delegate |
| P7 | **Recurring stable pattern** — only stable recurring patterns justify a static agent | PARTIAL (Tier S/M recurrence is high, but state-coupling overrides) → direct | PASS (Late-Branch 4-Phase is stable + recurring) → delegate |

**Scorecard**:
- **Tier S/M push+PR**: P1, P2, P3, P4, P5, P6 all FAIL → **orchestrator-direct**. P7 is partial but overridden by P1–P6.
- **Tier L release / multi-step merge / Late-Branch closure**: P1–P7 all PASS → **manager-git retained**.

The relaxation is NOT ad-hoc — it is the output of a 7-principle gate applied per op class.

---

## §6. Over-decomposition symptom check (evidence §5.3)

The arXiv:2604.14228 / LangChain over-decomposition checklist (9 symptoms) was applied. Matching symptoms for the current doctrine:

- **Symptom 2** ("context-sensitive op routed to context-poor agent") — exact signature of the PR-#1338 incident. manager-git is the context-poor agent; branch-state op is the context-sensitive work.
- **Symptom 6** ("verification gap — the delegating orchestrator cannot mechanically verify the delegated state op") — the orchestrator cannot re-observe the worktree state manager-git acted on; verification degrades to "trust the subagent."
- **Symptom 7** ("doctrine bloat compensating for delegation ambiguity") — the Late-Branch 4-Phase body grew dense to compensate for the isolated-subagent state hazard. The relaxation reduces the doctrine pressure on Late-Branch (now Tier L only).

The relaxation resolves Symptoms 2, 6, 7 simultaneously without touching the Late-Branch body (which stays as the canonical Tier L closure).

---

## §7. MoAI 4-layer health (evidence §5.4)

| Layer | Health | Note |
|-------|--------|------|
| **Skills** (progressive disclosure) | healthy | unchanged by this SPEC (skills cleanup is a follow-up SPEC) |
| **Handoffs** (phase transitions + Implementation Kickoff Approval) | healthy | unchanged |
| **Routers** (delegation-map + Intent Router) | healthy | delegation.yaml note refreshed (location #9) |
| **Subagents** | ONE over-delegation site | manager-git Tier S/M — this SPEC relaxes it |

The 4-layer diagnosis localizes the problem to exactly one layer (Subagents) and one agent within it (manager-git). There is no catalog-wide over-delegation pattern.

---

## §8. Verified URLs (12, from evidence §5 web research)

These URLs were fetched and verified by the evidence-base `Agent(general-purpose)` lens. Cited for traceability; the doctrine update does NOT add new external claims beyond these.

1. https://arxiv.org/abs/2512.08296 — multi-agent scaling limits (β = −0.408 below 45% baseline; −39% to −70% sequential state-coupled degradation).
2. https://arxiv.org/abs/2604.14228 — "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (companion repo: `github.com/VILA-Lab/Dive-into-Claude-Code`). 4-tier context-cost ladder (Hooks < Skills < Plugins < MCP), 7× plan-mode overhead.
3. https://github.com/VILA-Lab/Dive-into-Claude-Code — companion repository for arXiv:2604.14228.
4. https://www.anthropic.com/engineering/multi-agent-research-systems — Anthropic's multi-agent research systems post (≈15× token overhead figure).
5. https://www.anthropic.com/engineering/effective-multi-agents — Anthropic's effective agents guidance ("start single, add tools first").
6. https://docs.anthropic.com/en/docs/build-with-claude/agentic-tools — agentic tooling patterns.
7. https://python.langchain.com/docs/concepts/multi_agent/ — LangChain 4-pattern survey (Subagents / Skills / Handoffs / Routers).
8. https://blog.langchain.com/... — LangChain blog multi-agent series (specific post URL in evidence report).
9. https://code.claude.com/docs/en/sub-agents — Claude Code sub-agents documentation.
10. https://code.claude.com/docs/en/agent-teams — Claude Code agent teams documentation.
11. https://code.claude.com/docs/en/hooks-guide — Claude Code hooks documentation.
12. https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices — Anthropic prompting best practices (literal-instruction-following + Opus 4.8 subagent spawn steering).

> **Note**: the evidence report (§5) carries the verbatim URL list and fetch results. The 12 URLs are re-listed here so `research.md` is self-contained per the Tier L research-artifact convention; they were not re-fetched during plan-phase (the evidence base was fetched 2026-08-04, this plan-phase same-day).

---

## §9. Run-phase verification TODO (thin — evidence §8 deferred)

These items are evidence-base §8 "thin" items NOT covered by this SPEC (out of scope per spec.md §F) — recorded here so the follow-up SPECs (skills/hooks) can pick them up:

- security-{scan,turn,commit} 3-hook deep duplication verification.
- 5 possible-dead hooks actual-event wiring confirmation.
- output-style 3-persona banner drift quantification.
- `branch_guard.go` env path admitting orchestrator-direct Tier S/M — this IS covered (M3, REQ-OGR-011).

---

## §10. Open questions carried into plan.md [NEEDS CLARIFICATION]

One item deferred to Implementation Kickoff Approval resolution (per `[NEEDS CLARIFICATION]` marker convention) — an operational detail of the relaxation, not a scope question:

1. **branch-guard opt-in default state for this project** — is `Workflow.BranchGuard.Enabled` on or off for the maintainer checkout? Determines whether the env path is load-bearing or dormant. See plan.md §B.

This item does not block plan-phase audit-readiness; it MUST be resolved at the Implementation Kickoff Approval human gate before run-phase entry.

> **iter-2 resolution note (D4)**: the prior item #2 (`MOAI_BRANCH_GUARD_EXEMPT=1` lifecycle) was resolved IN-SPEC in iter-2 — per-invocation inline is mandated per `design.md` §3.3 (the persistent-settings alternative was rejected in `plan.md` §G). See plan.md §B for the resolved text. Only item #1 remains open.
