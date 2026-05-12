# Research — SPEC-V3R2-HRN-002 Evaluator Memory Scope Amendment

> Phase 0.5 deep codebase research preceding plan.md.
> Captures the as-is state of evaluator-active context inheritance, Sprint Contract durability, and the §11.4 surface of the design constitution that HRN-002 amends. Establishes the gap delta against spec.md §5 (19 REQs).

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec (HRN-002 plan author) | Initial research for HRN-002 plan phase — maps as-is evaluator memory leakage surface across constitution, agent body, GAN loop skill, harness Go module, and design.yaml/harness.yaml configs. |

---

## 1. Research Goal

Establish ground truth for SPEC-V3R2-HRN-002 plan:

1. Audit current evaluator-memory architecture across constitution clauses (§11), agent body, gan-loop skill, harness Go config loader, and yaml configs.
2. Locate the precise insertion point for the new `§11.4.1 Evaluator Memory Scope (Principle 4)` text declared in spec.md §1.2.
3. Quantify gap delta against spec.md §5 (19 REQs across 5 EARS modalities).
4. Surface adjacent SPECs (CON-001, CON-002, HRN-001, HRN-003) and the precise hook points each provides.
5. Derive milestone breakdown for plan.md (M1–M5 mirror of SPC-001 shape).

This research precedes plan.md per `.claude/skills/moai-workflow-spec` Phase 0.5 protocol and CON-002 amendment requirement (this SPEC itself is a CON-002 use case — the proposal document IS this SPEC).

---

## 2. As-Is — Design Constitution §11 Surface (FROZEN amendment target)

### 2.1 §11 GAN Loop Contract (existing, FROZEN)

- `.claude/rules/moai/design/constitution.md:274` — Section header `## 11. GAN Loop Contract` opens the FROZEN GAN protocol.
- `.claude/rules/moai/design/constitution.md:278-285` — `### Loop Mechanics` enumerates 6-step GAN cycle. No mention of evaluator context lifecycle.
- `.claude/rules/moai/design/constitution.md:287-292` — `### Escalation` invokes `evaluator-active generates a detailed failure report` (line 290) — silent on whether the report carries forward.
- `.claude/rules/moai/design/constitution.md:294-299` — `### Improvement Gate` declares `evaluator-active must identify a different dimension for improvement` (line 298) — implicitly requires comparison to prior iteration but does not declare memory scope.
- `.claude/rules/moai/design/constitution.md:301-306` — `### Strict Mode` requires minimum 2 iterations — same silence on memory.
- `.claude/rules/moai/design/constitution.md:308-335` — `### Sprint Contract Protocol` is the canonical amendment target.

### 2.2 §11.4 Sprint Contract Protocol (current text, before HRN-002 amendment)

- `.claude/rules/moai/design/constitution.md:308` — Section header `### Sprint Contract Protocol`.
- `.claude/rules/moai/design/constitution.md:312-316` — Contract Generation step declares evaluator produces `acceptance_checklist`, `priority_dimension`, `test_scenarios`, `pass_conditions`.
- `.claude/rules/moai/design/constitution.md:325-328` — Contract Evolution declares passed criteria carry forward, failed criteria refined, new criteria may be added. This is the durable Sprint Contract state surface — already in place.
- `.claude/rules/moai/design/constitution.md:330-335` — Sprint Contract rules: required at `thorough`, optional at `standard`, scoring constrained to contracted criteria, storage at `.moai/sprints/` from `design.yaml sprint_contract.artifact_dir`.
- **Gap**: §11.4 nowhere declares whether the evaluator's *judgment memory* carries forward between iterations. This silence is the bug Master §5.7 identifies. HRN-002 inserts new §11.4.1 directly below line 335 to close the gap.

### 2.3 §11 Loop Decision (existing handoff to evaluator context)

- `.claude/rules/moai/design/constitution.md:280-285` — Loop Mechanics declare evaluator scores artifacts but say nothing about whether the scoring agent receives a fresh context per call or inherits one.
- `.claude/rules/moai/design/constitution.md:295-299` — Improvement Gate requires evaluator to "identify a different dimension" — current implementation relies on context inheritance to remember prior dimension, which is precisely the cascade HRN-002 forbids. Post-amendment, the dimension state moves to the Sprint Contract criterion records (file-backed, durable) rather than evaluator memory (ephemeral).

### 2.4 §2 FROZEN Zone declaration (CON-001 baseline)

- `.claude/rules/moai/design/constitution.md:39-45` — `[FROZEN] GAN Loop contract (Section 11)` listed verbatim — confirms §11.4 is FROZEN, so this SPEC's amendment must pass CON-002 5-layer gate (REQ-HRN-002-008).
- `.claude/rules/moai/design/constitution.md:42-44` — `[FROZEN] Pipeline phase ordering constraints (manager-spec always first, evaluator-active always last in loop)` — evaluator's loop position is FROZEN; HRN-002 does not change position, only memory scope. No conflict.
- `.claude/rules/moai/design/constitution.md:44` — `[FROZEN] Pass threshold floor (minimum 0.60, cannot be lowered by evolution)` — spec.md §7 affirms this floor is preserved (cross-reference SPEC-V3R2-HRN-001 REQ-012).
- `.claude/rules/moai/core/zone-registry.md` — registry table covers `design/constitution.md §11` as a Frozen entry; HRN-002 amendment will require a new CONST-V3R2-NNN entry (allocated by registry policy, deferred to M5).

### 2.5 Version metadata (target for version bump)

- `.claude/rules/moai/design/constitution.md:405` — `Version: 3.4.0` line. HRN-002 bumps to `3.5.0` (minor increment per amendment policy, REQ-HRN-002-002).
- `.claude/rules/moai/design/constitution.md:1-9` — `## HISTORY` block. HRN-002 appends a new row recording the amendment per REQ-HRN-002-002.

---

## 3. As-Is — evaluator-active agent body

### 3.1 Frontmatter declaration

- `.claude/agents/moai/evaluator-active.md:1-26` — Frontmatter declares `memory: project` on line 16 — this is the PROJECT-scoped persistent memory directory at `.claude/agent-memory/evaluator-active/`, NOT the per-iteration LLM context. The two are distinct concepts and the amendment text in §11.4.1 must clarify this so future readers do not conflate them.
- `.claude/agents/moai/evaluator-active.md:13` — `model: sonnet` — HRN-002 does not change model (spec.md §2.2 confirms out-of-scope).
- `.claude/agents/moai/evaluator-active.md:14` — `effort: high` — preserved by HRN-002.
- `.claude/agents/moai/evaluator-active.md:15` — `permissionMode: plan` — preserved.
- `.claude/agents/moai/evaluator-active.md:20-25` — `SubagentStop` hook fires `evaluator-completion`. Not affected by HRN-002 amendment.

### 3.2 Agent body cross-reference target

- `.claude/agents/moai/evaluator-active.md:91-103` — `## Sprint Contract Negotiation (Phase 2.0, thorough only)` is the contextually-correct insertion point for a single cross-reference line per REQ-HRN-002-006.
- `.claude/agents/moai/evaluator-active.md:99-103` — `## Intervention Modes` describes `per-sprint` (thorough) vs `final-pass` (standard). The HRN-002 amendment text in §11.4.1 declares fresh context per iteration; the agent body needs only a single line reference (e.g., immediately above or below line 91) telling readers "Per design-constitution §11.4.1, evaluator context is ephemeral per iteration; do NOT carry forward prior judgment rationale between iterations." No structural rewrite.

### 3.3 Current memory leakage observation

- `.claude/agents/moai/evaluator-active.md:104-108` — Mode-specific deployment declares: Sub-agent invoked via `Agent(subagent_type="evaluator-active")`, Team teammate via SendMessage, CG Leader performs directly. **All three modes inherit parent session context.** In particular:
  - Sub-agent mode: `Agent()` does NOT reset Claude context across calls; subsequent `Agent(evaluator-active)` invocations inherit the orchestrator's accumulated thread history including prior iteration's evaluator output.
  - Team mode: teammate persists across messages, accumulating mailbox history.
  - CG mode: Leader (Claude) runs evaluation in same context as orchestration.
- Result: with no fresh-respawn protocol, every iteration's `evaluator-active` invocation sees the entire prior evaluation transcript inside its context window. This is the cascade Zhuge et al. flagged.

### 3.4 What HRN-002 changes about the agent body

Per REQ-HRN-002-006: a single cross-reference line. Not structural. The enforcement is at the runner level (REQ-HRN-002-009), not the agent definition.

---

## 4. As-Is — moai-workflow-gan-loop skill

### 4.1 Skill frontmatter

- `.claude/skills/moai-workflow-gan-loop/SKILL.md:1-31` — Frontmatter. No `triggers.fresh_context` field — pure runtime concern.
- `.claude/skills/moai-workflow-gan-loop/SKILL.md:10` — `allowed-tools: Read, Write, Edit, Grep, Glob, Bash` — no Agent tool. The skill itself does not spawn the evaluator; that responsibility belongs to the runner (Phase 3 Evaluator Scoring step).

### 4.2 Iteration loop semantics (current)

- `.claude/skills/moai-workflow-gan-loop/SKILL.md:77-141` — `### GAN Loop Execution Flow` describes Phase 1-5. Phase 3 (Evaluator Scoring) at lines 109-118 says "Evaluator scores against the 4 dimensions using the Evaluator Leniency Prevention mechanisms" — silent on context lifecycle.
- `.claude/skills/moai-workflow-gan-loop/SKILL.md:120-131` — Phase 4 Loop Decision: `if overall_score >= pass_threshold` EXIT; `elif iteration >= max_iterations` ESCALATE; `else ITERATE`. The ITERATE path is the cascade surface: in current implementation, the next iteration's evaluator inherits the previous Phase 3 transcript.
- `.claude/skills/moai-workflow-gan-loop/SKILL.md:133-140` — Phase 5 Iteration Feedback: "Evaluator generates targeted feedback per failed criterion. Builder receives the feedback and previous Sprint Contract. Previously passed criteria carry forward (no regression allowed). New Sprint Contract is generated for failed criteria only." — Sprint Contract is durable (file-backed), feedback is in-memory.
- `.claude/skills/moai-workflow-gan-loop/SKILL.md:206-240` — Sprint Contract Structure JSON shows `passed | failed | refined | new` status carried per criterion — this is the durable substrate HRN-002 leverages.
- `.claude/skills/moai-workflow-gan-loop/SKILL.md:147-158` — Stagnation detection uses score deltas across iterations — this requires the SCORE record (durable, in the sprint artifact) but NOT the evaluator's judgment rationale.

### 4.3 What HRN-002 changes about the skill

Per REQ-HRN-002-007: the skill body cites §11.4.1 and declares fresh-respawn semantics on iteration entry. M3 task. Concrete addition: between Phase 4 (line 131) and Phase 5 (line 133), insert a step "**Iteration handoff**: at iteration boundary, respawn `evaluator-active` via fresh `Agent()` call. Prompt MUST contain only (BRIEF reference, Sprint Contract criterion states, artifact path). Prior iteration's evaluator transcript MUST NOT appear." Cross-reference design-constitution §11.4.1.

### 4.4 Mechanism 5 Regression Baseline interaction

- `.claude/skills/moai-workflow-gan-loop/SKILL.md:200-202` — Mechanism 5: "If a previous iteration passed a criterion, the current iteration must maintain that criterion. Regression from a previously passed criterion triggers an automatic score reduction in the relevant dimension." This rule is ENFORCED through the **Sprint Contract** (criterion state durable), not through the evaluator's memory — exactly what HRN-002 codifies.

---

## 5. As-Is — Harness Go module (HRN-001 hook point)

### 5.1 Loader and types

- `internal/harness/types.go:1-100` — Existing `HarnessConfig`-related types focus on observability/learning (Event, EventType, Tier, Pattern). No `Evaluator` struct yet — HRN-001 declares the harness-level loader for `harness.yaml` per its REQ-001 through REQ-005, but `evaluator.memory_scope` is HRN-002 territory.
- `internal/harness/loader.go:1` — package-level loader; HRN-002 extends this loader (or a sibling) to validate `evaluator.memory_scope == per_iteration` per REQ-HRN-002-011.
- `internal/harness/applier.go:167-176` — `Apply(proposal Proposal, evaluator SafetyEvaluator, ...)` — note this `evaluator` is the CON-002 safety evaluator (FrozenGuard pipeline), unrelated to GAN evaluator-active. Avoid name conflict in new code.

### 5.2 Where the new field lands

Per spec.md §10 traceability code-side paths:
- `internal/config/types.go` — `HarnessConfig` struct (from HRN-001) gains `Evaluator EvaluatorConfig` sub-struct with field `MemoryScope string`.
- `internal/config/loader.go` — `LoadHarnessConfig()` and `LoadDesignConfig()` both validate `evaluator.memory_scope == per_iteration`; non-matching values return `HRN_EVAL_MEMORY_FROZEN`.
- `internal/config/loader_test.go` — adds fixtures for valid `per_iteration` and invalid `cumulative` per AC-HRN-002-04.

### 5.3 Existing evaluator-profile config infrastructure

- `.moai/config/sections/harness.yaml:6-8` — `default_profile: "default"` references `.moai/config/evaluator-profiles/{name}.md`. HRN-001 already wires this. HRN-002 does NOT touch profile files (out of scope per spec.md §2.2).
- `.moai/config/sections/harness.yaml:67-101` — Each level (minimal/standard/thorough) declares `evaluator: bool`, `evaluator_mode`, optionally `evaluator_profile`. HRN-002 adds a single sibling key `evaluator.memory_scope` at the **harness top level** (parallel to `default_profile`), NOT per-level — applies to ALL levels uniformly.
- `.moai/config/sections/design.yaml:13-29` — `gan_loop:` block holds `sprint_contract.artifact_dir: ".moai/sprints"`. HRN-002 adds `evaluator.memory_scope: per_iteration` as a sibling under the top-level `design:` namespace (mirroring harness placement).

### 5.4 Runner integration point (REQ-HRN-002-009)

The actual GAN loop runner (the orchestrator-side loop that calls `Agent(subagent_type: "evaluator-active")` once per iteration) is currently *implicit* — the SKILL.md documents the flow but the orchestrator interprets it at runtime. No Go-side `internal/harness/gan_loop.go` exists yet. spec.md §10 lists this file as "new or modified".

Decision (D1, see §11 below): HRN-002 plan does NOT mandate a Go-side gan_loop.go file. The runner is the orchestrator skill (moai-workflow-gan-loop SKILL.md) which prescribes the spawn protocol. REQ-HRN-002-009 enforcement is via SKILL.md text + the leak-detection regression test (REQ-HRN-002-017) which can be a Go integration test placed under `internal/harness/` even without a runner module.

---

## 6. As-Is — Sprint Contract storage

### 6.1 Storage location

- `.moai/config/sections/design.yaml:27` — `artifact_dir: ".moai/sprints"` — canonical location.
- `.moai/sprints/` — currently empty on disk (no in-flight GAN design runs). Confirmed by `ls .moai/sprints/`. Architecture supports both `{team-id}/contract.yaml` and `{spec-id}/contract.yaml` aliases per REQ-HRN-002-016.

### 6.2 Sprint Contract YAML shape (existing)

- `.claude/skills/moai-workflow-gan-loop/SKILL.md:210-240` — JSON format documented. Fields: `sprint_id`, `iteration`, `priority_dimension`, `acceptance_checklist[]`, `test_scenarios[]`, `pass_conditions{}`, `negotiation_history[]`, `created_at`. Status enum: `pending | passed | failed`.
- HRN-002 spec.md §2.2 explicitly puts schema changes OUT of scope. The existing shape already supports the durability mechanic — the criterion `status` field carries forward state.

---

## 7. As-Is — SPEC corpus and adjacent dependencies

### 7.1 Blocked-by dependencies (verified on main)

- `.moai/specs/SPEC-V3R2-CON-001/` — exists on main (PR merged earlier in v3 Sprint 1). Provides zone-registry CONST-V3R2-NNN entries and `.claude/rules/moai/core/zone-registry.md` infrastructure. Required for HRN-002 to register its amendment as a new Frozen entry post-merge.
- `.moai/specs/SPEC-V3R2-CON-002/` — exists on main. Provides 5-layer graduation protocol Go types (`FrozenGuard.Check`, `Canary.Evaluate`, `ContradictionDetector.Scan`, `RateLimiter.Admit`, `HumanOversight.Approve`) per its REQ-CON-002-002. Required for HRN-002 amendment paperwork. CON-002 spec.md:169 cites SPEC-V3R2-HRN-002 explicitly as a consumer.
- `.moai/specs/SPEC-V3R2-HRN-001/` — exists on main. Provides `HarnessConfig` struct + `LoadHarnessConfig()` loader. HRN-001 spec.md:286 cites HRN-002 explicitly: "Evaluator memory amendment uses the typed harness config to declare `memory_scope: per_iteration`." Required for HRN-002 to extend the struct.

### 7.2 Blocks (downstream SPECs)

- SPEC-V3R2-HRN-003 — hierarchical per-leaf scoring depends on fresh-judgment semantics. Each leaf in the hierarchical AC tree (from SPC-001) gets independent scoring under fresh evaluator memory.
- SPEC-V3R2-WF-003 — multi-mode router uses Sprint Contract per this SPEC at `thorough` harness level.

### 7.3 Related SPECs

- SPEC-V3R2-SPC-001 — recently merged (PR #870, main `07dabe011`). Pioneers the CON-002 paperwork pattern in run phase and the hierarchical-AC schema. HRN-002 reuses both: CON-002 pattern verbatim for M5 paperwork; hierarchical AC for acceptance.md self-demonstration.
- SPEC-DESIGN-CONST-AMEND-001 — v2-legacy precedent for design-constitution amendments. Establishes amendment-writing pattern (HISTORY append, version bump, FROZEN tag).
- SPEC-V3R2-EVAL-001 — v3-legacy evaluator profile schema. Profile YAMLs consume the new `memory_scope` flag indirectly.
- SPEC-V3R2-EXT-004 — versioned migration auto-apply. BC-V3R2-010 migration (config flag addition) runs via session-start hook on first v3.0.0-beta.1 upgrade.

---

## 8. Gap Analysis — HRN-002 vs Already Landed

| Capability | spec.md REQ | Implementation status | Gap |
|------------|-------------|------------------------|-----|
| §11.4.1 text in design-constitution.md | REQ-HRN-002-001 | NOT LANDED | M2 task — insert verbatim per spec.md §1.2 |
| Constitution version bump 3.4.0 → 3.5.0 | REQ-HRN-002-002 | NOT LANDED | M2 task |
| `evaluator.memory_scope: per_iteration` in design.yaml | REQ-HRN-002-003 | NOT LANDED | M4 task |
| `evaluator.memory_scope: per_iteration` in harness.yaml | REQ-HRN-002-004 | NOT LANDED | M4 task |
| `HarnessConfig.Evaluator.MemoryScope` Go field | REQ-HRN-002-005 | NOT LANDED — HarnessConfig exists (HRN-001) but no Evaluator sub-struct | M2 task — extend struct |
| evaluator-active agent body cross-ref | REQ-HRN-002-006 | NOT LANDED | M3 task — single-line addition near agent body line 91 |
| moai-workflow-gan-loop SKILL.md cite | REQ-HRN-002-007 | NOT LANDED | M3 task — insert iteration-handoff step between Phase 4 and Phase 5 |
| CON-002 5-layer execution | REQ-HRN-002-008 | NOT LANDED — CON-002 infrastructure live, but no run yet for this amendment | M5 task — mirror SPC-001 evidence pattern |
| Fresh respawn at iteration start | REQ-HRN-002-009 | NOT LANDED — current runner inherits context | M3 task — encoded in SKILL.md text + leak-detection test |
| evolution-log.md record | REQ-HRN-002-010 | NOT LANDED — `.moai/research/evolution-log.md` does not exist yet (CON-002 ships infrastructure; HRN-002 may be the first writer) | M5 task |
| Loader validator `HRN_EVAL_MEMORY_FROZEN` | REQ-HRN-002-011 | NOT LANDED | M4 task |
| Sprint Contract state durability preserved | REQ-HRN-002-012 | ALREADY LANDED — `.claude/skills/moai-workflow-gan-loop/SKILL.md:206-240` Sprint Contract structure carries criterion state forward | — (read-path compatibility, no new work) |
| Evaluator memory volatility (fresh respawn) | REQ-HRN-002-013 | NOT LANDED — same gap as REQ-009 | M3 task (paired) |
| FROZEN value preservation across v3.0.0 | REQ-HRN-002-014 | NOT LANDED — needs validator + test | M4 task |
| Canary verdict attached to AskUser record | REQ-HRN-002-015 (Optional) | NOT LANDED — M5 task | M5 task — included in evidence block |
| Solo-mode `{spec-id}/contract.yaml` alias | REQ-HRN-002-016 (Optional) | ALREADY SUPPORTED by file-path conventions, but undocumented | M3 task — documented in SKILL.md |
| Prior-judgment leak detection test | REQ-HRN-002-017 | NOT LANDED | M2 or M3 task — Go integration test scanning spawn prompts |
| `cumulative` value rejection at loader | REQ-HRN-002-018 | Same as REQ-011 | M4 task (paired) |
| FrozenGuard rejects unauthorized §11.4.1 edits | REQ-HRN-002-019 | Infrastructure live via CON-002 — needs evidence run | M5 task |

Summary: HRN-002 is ~95% greenfield (only Sprint Contract durability is already in place). Plan therefore focuses on:

1. **M1** — plan-phase artifacts (this PR).
2. **M2** — Constitution amendment text insertion + HarnessConfig struct extension.
3. **M3** — Agent body cross-reference, SKILL.md iteration-handoff step, prior-judgment leak detection test.
4. **M4** — design.yaml + harness.yaml config keys, loader validator `HRN_EVAL_MEMORY_FROZEN`.
5. **M5** — CON-002 5-layer paperwork (mirror SPC-001 pattern), evolution-log.md record, AskUserQuestion approval.

---

## 9. Risks Identified During Research

| # | Risk | Severity | Mitigation milestone |
|---|------|----------|-----------------------|
| R1 | Canary step finds <3 design projects in `.moai/design/` corpus (insufficient subjects per spec.md §8) | MEDIUM | M5 — document corpus shortage; cross-validate using the last 3 v2-legacy design completions if v3 corpus is empty; if still insufficient, proceed with explicit "CanaryUnavailable" note per REQ-CON-002-020 |
| R2 | Fresh respawn overhead per iteration (Opus 4.7 spawn latency) | MEDIUM | M2 — measure baseline once; document in PR if >5% of typical iteration |
| R3 | Human reviewer approves M5 paperwork without reading R1 §9 evidence citation | HIGH | M5 — bundle R1 §9 citation + canary verdict + design-principles.md Principle 4 + this research.md §10 in the AskUserQuestion options[].description, with first option labeled "(권장) Approve with full evidence reviewed" |
| R4 | Existing v2 evaluator sessions persist memory through session resume | MEDIUM | M5 — release-notes entry for BC-V3R2-010 documenting automatic session retirement on first v3.0.0-beta.1 startup via session-start hook |
| R5 | Sprint Contract state grows unbounded across long-running design projects | LOW | Carry-forward criteria list is bounded by SPEC acceptance count (typically <20); document archival on SPEC completion in SKILL.md M3 task |
| R6 | evaluator-active prompt reconstruction subtly leaks prior context via Sprint Contract serialization artifacts | HIGH | REQ-HRN-002-017 integration test scans every spawn prompt substring for `Score:` / `Feedback:` / `Verdict:` substrings; fails CI on leak |
| R7 | CON-002 rate limiter (3/week cap) delays the amendment if other v3.x amendments already burned the quota | LOW | This SPEC is high-priority P0 Critical (Master §5.7); rate-limiter burst allowance applies per CON-002 §5 Layer 4 design |
| R8 | Evolution-log.md write fails or is corrupted; amendment unrecorded | MEDIUM | REQ-HRN-002-010 is blocking; amendment not considered complete until log write succeeds — verified by progress.md run-phase update |
| R9 | Third-party evaluator skills or plugins break on fresh-context (unknown territory) | LOW | No known third-party evaluator skills today; release notes document the behavior change; plugin authors get 30-day notice via CHANGELOG |
| R10 | Confusion between `memory: project` agent frontmatter (persistent disk memory) and per-iteration LLM context (Claude thread) | MEDIUM | M2 — amendment text explicitly distinguishes; M3 — agent body cross-reference clarifies |
| R11 | Solo-mode `{spec-id}/contract.yaml` alias undocumented, causes runtime confusion | LOW | M3 — documented in SKILL.md (REQ-HRN-002-016) |
| R12 | spec.md §2.1 declares evolution-log path as `.moai/design/v3-research/evolution-log.md` while CON-002 §5 §6 declares `.moai/research/evolution-log.md`; path disambiguation needed | MEDIUM | M1/M5 — `.moai/design/v3-research/evolution-log.md` is the design-subsystem log (per design-constitution §6); CON-002's core-level log at `.moai/research/evolution-log.md` is separate. HRN-002 writes to BOTH: the core log per CON-002 §6, and a cross-reference entry in the design log per design-constitution §6. M5 task records this dual-write |

---

## 10. Cross-Reference Summary (file:line anchors, ≥25)

1. `.claude/rules/moai/design/constitution.md:274` — `## 11. GAN Loop Contract` section header.
2. `.claude/rules/moai/design/constitution.md:278-285` — Loop Mechanics 6-step cycle (memory-scope silent).
3. `.claude/rules/moai/design/constitution.md:287-292` — Escalation step invoking evaluator failure report.
4. `.claude/rules/moai/design/constitution.md:294-299` — Improvement Gate requiring different dimension per iteration.
5. `.claude/rules/moai/design/constitution.md:301-306` — Strict Mode minimum 2 iterations.
6. `.claude/rules/moai/design/constitution.md:308` — `### Sprint Contract Protocol` section header (amendment insertion site, new §11.4.1 follows §11.4).
7. `.claude/rules/moai/design/constitution.md:312-316` — Contract Generation declaring criterion records.
8. `.claude/rules/moai/design/constitution.md:325-328` — Contract Evolution declaring passed/failed/refined/new states (durable substrate).
9. `.claude/rules/moai/design/constitution.md:330-335` — Sprint Contract rules (HARD list + artifact_dir).
10. `.claude/rules/moai/design/constitution.md:39-45` — `[FROZEN]` zone list confirming §11 frozen.
11. `.claude/rules/moai/design/constitution.md:42-44` — `[FROZEN]` pipeline phase ordering + pass threshold floor.
12. `.claude/rules/moai/design/constitution.md:405` — `Version: 3.4.0` — target for bump.
13. `.claude/rules/moai/design/constitution.md:1-9` — `## HISTORY` block — target for amendment row.
14. `.claude/agents/moai/evaluator-active.md:1-26` — Frontmatter (memory: project ≠ LLM context distinction).
15. `.claude/agents/moai/evaluator-active.md:13` — `model: sonnet` preservation.
16. `.claude/agents/moai/evaluator-active.md:15` — `permissionMode: plan` preservation.
17. `.claude/agents/moai/evaluator-active.md:91-103` — `## Sprint Contract Negotiation` section — cross-reference insertion site.
18. `.claude/agents/moai/evaluator-active.md:99-103` — `## Intervention Modes` per-sprint/final-pass.
19. `.claude/agents/moai/evaluator-active.md:104-108` — Mode-Specific Deployment (sub-agent/team/CG).
20. `.claude/skills/moai-workflow-gan-loop/SKILL.md:77-141` — GAN Loop Execution Flow.
21. `.claude/skills/moai-workflow-gan-loop/SKILL.md:109-118` — Phase 3 Evaluator Scoring (memory-scope silent).
22. `.claude/skills/moai-workflow-gan-loop/SKILL.md:120-131` — Phase 4 Loop Decision (cascade surface at ITERATE branch).
23. `.claude/skills/moai-workflow-gan-loop/SKILL.md:133-140` — Phase 5 Iteration Feedback.
24. `.claude/skills/moai-workflow-gan-loop/SKILL.md:206-240` — Sprint Contract JSON shape (durable criterion records).
25. `.claude/skills/moai-workflow-gan-loop/SKILL.md:200-202` — Mechanism 5 Regression Baseline.
26. `.claude/skills/moai-workflow-gan-loop/SKILL.md:147-158` — Stagnation detection (uses score deltas, not judgment rationale).
27. `.moai/config/sections/harness.yaml:6-8` — `default_profile: "default"` + evaluator-profile infrastructure.
28. `.moai/config/sections/harness.yaml:67-101` — Per-level evaluator config (minimal/standard/thorough).
29. `.moai/config/sections/design.yaml:13-29` — gan_loop block + sprint_contract.artifact_dir target.
30. `.moai/config/sections/design.yaml:6-9` — `design:` top-level namespace (HRN-002 sibling key insertion point).
31. `internal/harness/types.go:1-100` — Existing types (Event/Pattern/Tier); no Evaluator struct yet.
32. `internal/harness/loader.go:1` — Loader entry point.
33. `internal/harness/applier.go:167-176` — Apply() with safety evaluator (CON-002, unrelated name).
34. `.moai/specs/SPEC-V3R2-HRN-001/spec.md:286` — Explicit cite of HRN-002 consumer relationship.
35. `.moai/specs/SPEC-V3R2-HRN-001/spec.md:118-119` — Sprint Contract + fresh-memory evaluator declared OUT of HRN-001 scope.
36. `.moai/specs/SPEC-V3R2-CON-002/spec.md:169` — CON-002 cites HRN-002 as a consumer of the 5-layer protocol.
37. `.moai/specs/SPEC-V3R2-CON-002/spec.md:43-46` — Goal of 5-layer protocol generalization.
38. `.moai/specs/SPEC-V3R2-CON-002/spec.md:94-100` — REQs-CON-002-001 through 007 covering the layer interfaces.
39. `.moai/specs/SPEC-V3R2-CON-002/spec.md:111-113` — REQ-CON-002-020 declaring `CanaryUnavailable` when <3 SPECs.
40. `.moai/specs/SPEC-V3R2-SPC-001/spec.md` — SPC-001 amendment pattern (recently merged via PR #870, main `07dabe011`).
41. `.moai/specs/SPEC-V3R2-SPC-001/con-002-amendment-evidence.md` — SPC-001 5-layer evidence template, mirrored by HRN-002 M5.
42. `.moai/specs/SPEC-V3R2-SPC-001/canary-v2-reparse.txt` — SPC-001 Canary log structure.
43. `.moai/sprints/` — empty directory; storage target for Sprint Contracts.
44. `.moai/research/` — directory referenced by design-constitution §6 for design observations; cross-referenced by CON-002 for core evolution-log.
45. `.claude/rules/moai/core/zone-registry.md` — Frozen entry registry; HRN-002 amendment adds a new CONST-V3R2-NNN entry in M5.
46. `.moai/specs/SPEC-V3R2-HRN-002/spec.md:73-83` — Verbatim §11.4.1 amendment text declared in spec.md §1.2.
47. `.moai/specs/SPEC-V3R2-HRN-002/spec.md:163-230` — Full REQ list (19 REQs).
48. `.moai/specs/SPEC-V3R2-HRN-002/spec.md:239-249` — AC-HRN-002-01 through 11 declared in spec.md §6.
49. `.moai/specs/SPEC-V3R2-HRN-002/spec.md:307` — Traceability matrix from spec.md §10 (REQ→AC mapping).
50. `.moai/specs/SPEC-V3R2-HRN-002/spec.md:317-331` — Code-side path list (target files for M2-M5).

Total file:line anchors: 50 (>25 minimum required).

---

## 11. Decisions Driving Plan.md

D1. **Runner enforcement via SKILL.md + Go integration test, not via new `internal/harness/gan_loop.go` module.** The current orchestrator-level runner (Phase 4 ITERATE branch in moai-workflow-gan-loop SKILL.md) is the actual integration point. Creating a parallel Go-side runner would duplicate logic. REQ-HRN-002-009 enforcement is via SKILL.md text (declarative) + REQ-HRN-002-017 leak-detection regression test (validation). Test placement: `internal/harness/evaluator_leak_test.go` under the existing harness module — this avoids the need to create new package structure.

D2. **Constitution amendment text is the canonical surface.** The verbatim §11.4.1 text from spec.md §1.2 is the single source of truth. design.yaml/harness.yaml flag and Go struct field are the runtime enforcement substrate; SKILL.md and agent body changes are documentation cross-references. M2 carries the constitutional weight; M3-M4 carry the operational weight; M5 carries the paperwork weight.

D3. **CON-002 paperwork mirrors SPC-001 pattern verbatim.** PR #870 (main `07dabe011`) just exercised the 5-layer evidence cycle. HRN-002 M5 produces an identical-format `con-002-amendment-evidence.md` at `.moai/specs/SPEC-V3R2-HRN-002/con-002-amendment-evidence.md` + a Canary log at `.moai/specs/SPEC-V3R2-HRN-002/canary-fresh-memory-eval.txt`. No methodology innovation needed.

D4. **Hierarchical AC schema is consumed, not amended.** SPC-001 unblocked the hierarchical AC schema and merged PR #870. HRN-002's acceptance.md self-demonstrates the schema on ≥3 ACs to (a) recursively prove the schema works for HRN-002 paperwork itself and (b) satisfy the operator-confirmation criteria for plan-auditor.

D5. **Dual evolution-log write.** Per Risk R12, HRN-002 writes the amendment record to BOTH `.moai/research/evolution-log.md` (CON-002 core log) AND `.moai/design/v3-research/evolution-log.md` (design-subsystem cross-reference). The cross-reference entry includes a `core_log_ref: <core entry ID>` pointer to avoid duplication.

D6. **Backward compatibility via session-start hook.** Per REQ-HRN-002-014 + AC-HRN-002-11, v2 evaluator sessions are retired on first v3.0.0-beta.1 upgrade by the session-start hook. M5 task — documented in CHANGELOG entry, no code change required (mechanism is the BC-V3R2-010 migration already declared in spec.md §3).

D7. **FROZEN value enforcement is symmetric: loader rejects, FrozenGuard rejects.** REQ-HRN-002-011 (loader returns `HRN_EVAL_MEMORY_FROZEN` on invalid value) and REQ-HRN-002-019 (FrozenGuard blocks unauthorized §11.4.1 edits via CON-001 zone registry) form a two-layer defense. Tests for both land in M4 (loader test) and M5 (guard test, mirrors SPC-001's existing guard test for CONST-V3R2-001).

D8. **Performance benchmark NOT required at plan phase.** Unlike SPC-001 which had AC-SPC-001-14 perf budget (365-leaf parse <500ms), HRN-002 has no analogous quantitative budget. Spawn latency (Risk R2) is documented but not gated. If observed overhead exceeds 5% of iteration duration, a follow-up SPEC may add a budget; HRN-002 ships without one.

D9. **Plan-in-main (no SPEC worktree).** Per CLAUDE.local.md §18.12 BODP and PR #822 doctrine, plan-phase artifacts work on a feature branch from main. Plan PR is `plan/SPEC-V3R2-HRN-002` cut from `main` HEAD `07dabe011`. Run-phase work creates a fresh worktree at run start per spec-workflow.md Step 2.

D10. **No SubagentStop hook change.** evaluator-active's existing `SubagentStop → evaluator-completion` hook (agent body line 20-25) does not need to fire any new logic. The per-iteration respawn is orchestrator-level; the hook fires once per Agent() spawn invocation regardless. Hook handler in `internal/hook/agents/` remains unchanged.

---

End of research.
