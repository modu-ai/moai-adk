# research.md — SPEC-DOCSITE-ADVANCED-001

> Plan-phase evidence file. Documents the parity-debt findings (§A), the
> per-page content-source readiness verification (§B), and the memory
> carry-over risk assessment (§C). All claims below are verified against
> the live repository state at plan-phase (2026-07-13).

## §A Parity debt findings

### A.1 The verbatim recon output

The plan-phase recon listed all advanced content files across 4 locales. The `find` output (truncated to the relevant section):

```
docs-site/content/{ko,en,ja,zh}/advanced/ — 15 files per locale (14 pages + _index.md + _meta.yaml)
```

Concrete count per locale (excluding `_index.md` and `_meta.yaml`):

```
agent-guide.md, builder-agents.md, catalog-system.md, claude-md-guide.md,
decision-memory.md, harness-profiles.md, harness-v4-builder.md, hooks-guide.md,
hooks-reference.md, security-notes.md, settings-json.md, skill-guide.md,
statusline.md, ultracode-workflows.md
```

**14 content files per locale × 4 locales = 56 content files. All present.**

The verbatim per-locale `_meta.yaml` entry inventory (verified by Read at plan-phase):

| Locale | Entries in `_meta.yaml` | Content files present | Delta (content minus _meta) |
|--------|-------------------------|----------------------|------------------------------|
| ko     | 11 (skill-guide, agent-guide, builder-agents, hooks-guide, hooks-reference, settings-json, security-notes, statusline, claude-md-guide, harness-profiles, catalog-system) | 14 | 3 missing: **decision-memory**, **harness-v4-builder**, **ultracode-workflows** |
| en     | 8 (skill-guide, agent-guide, builder-agents, hooks-guide, settings-json, security-notes, statusline, claude-md-guide) | 14 | 6 missing: **catalog-system**, **decision-memory**, **harness-profiles**, **harness-v4-builder**, **hooks-reference**, **ultracode-workflows** |
| ja     | 8 (same set as en) | 14 | 6 missing: (same set as en) |
| zh     | 8 (same set as en) | 14 | 6 missing: (same set as en) |

### A.2 Comparison with the brief's claim

The brief said: "KO advanced/_meta.yaml currently has ~11 entries (includes hooks-reference, harness-profiles, catalog-system). EN/JA/ZH advanced/_meta.yaml currently have ~8 entries (missing those three)."

**Verified facts**:
- KO = 11 entries ✓ (brief correct)
- EN/JA/ZH = 8 entries each ✓ (brief correct)
- The 3 KO-only entries (hooks-reference, harness-profiles, catalog-system) ARE present in KO _meta and absent in EN/JA/ZH _meta ✓ (brief correct)
- **BUT KO is ITSELF missing 3 more entries** (decision-memory, harness-v4-builder, ultracode-workflows) — the brief did NOT mention these.

**Conclusion**: the parity debt is wider than the brief described. All 4 locales are broken; KO is missing 3 entries; EN/JA/ZH are missing 6. The sidebar menu (`main.yaml` Advanced section, lines 348-431) correctly registers all 14 with 4-locale names — so the visible navigation works; only the geekdoc page-title resolver (`_meta.yaml`) is broken.

### A.3 Root cause hypothesis

The most plausible root cause (NOT verified — hypothesis only): the per-locale `_meta.yaml` files were last updated when there were 8 (or 11) advanced pages, and the additional pages (decision-memory, harness-v4-builder, ultracode-workflows in KO; plus catalog-system, harness-profiles, hooks-reference in EN/JA/ZH) were added later without updating `_meta.yaml`. The `main.yaml` was updated correctly (it has all 14 entries), but `_meta.yaml` was missed.

This is consistent with the `hns-oss-docs-structure-map` skill's note that `gen_menu.py` DOES NOT exist — menu/page-title edits are manual, and manual edits can drift.

### A.4 Resolution approach

The parity debt fix (M1) brings all 4 locales' `_meta.yaml` to a clean 14-entry baseline BEFORE adding the 6 new entries (M5). The titles for the existing-but-unregistered slugs come from `main.yaml` (lines 367-383 for decision-memory, 373-377 for harness-v4-builder, 367-371 for ultracode-workflows, etc.) — these are the canonical 4-locale titles and they are reused verbatim in `_meta.yaml`.

---

## §B Per-page content-source readiness (verification-claim-integrity §1.1 surface 3 binding)

The brief cautioned: "if a content source is insufficient to author a page, flag that page as a blocker." Per `feedback_defect_claim_verification` (the spawn prompt's applied-lessons context), the agent verified each source by direct inspection rather than accepting the brief's cautionary language.

### B.1 tokenomics-overview — READY

| Source | Path | Verified | Content contribution |
|--------|------|----------|----------------------|
| Readme redesign report | `.moai/reports/readme-docs-redesign-20260713.md` | ✓ exists (19654 bytes) | 3-pillar narrative structure |
| Readme redesign HTML twin | `.moai/reports/readme-docs-redesign-20260713.html` | ✓ exists (26259 bytes) | (HTML twin — not used for agent context; the .md is canonical per `project_v3_tokenomics_docs_plan`) |
| v3-rc11 draft | `.moai/reports/readme-draft-v3-rc11.md` | ✓ exists (46514 bytes) | Substantive draft of the 3-pillar narrative |
| Memory | `project_v3_tokenomics_docs_plan.md` | ✓ read in full | 4-axis material inventory + linked reference topics |
| Memory | `reference_tokenomics_press_2026_07.md` | (in index; not re-read at plan-phase) | Market context hook |
| Token-Economy Epic | SPEC-TOKEN-ACCOUNTING / SPEC-TOKEN-ROUTING / SPEC-TOKEN-BUDGET-STOP | ✓ exist | Layer A/B/D implemented behavior |

**Verdict**: source substantially ready. NO blocker. NO honesty caveat required (the page describes product positioning, not unverified behavior).

### B.2 token-budget — READY

| Source | Path | Verified |
|--------|------|----------|
| SPEC | `.moai/specs/SPEC-TOKEN-BUDGET-STOP-001/` | ✓ exists (6 files) |
| Doctrine | `.claude/rules/moai/workflow/context-window-management.md` | ✓ (always-loaded; full content in agent context) |
| Doctrine | `.claude/rules/moai/workflow/session-handoff.md` | ✓ (always-loaded; full content in agent context — the 6-block paste-ready format) |
| Doctrine | `.claude/rules/moai/core/agent-common-protocol.md` § File-redirect contract | ✓ (always-loaded) |
| Memory | `project_token_budget_stop_plan_handoff.md` | (in MEMORY.md index, ✅ close marker) | plan-phase handoff |

**Verdict**: source ready. NO blocker. NO honesty caveat (behavior is fully implemented).

### B.3 no-haiku-3tier — READY (with caveat)

| Source | Path | Verified |
|--------|------|----------|
| Design report (v2) | `.moai/reports/agent-architecture-redesign-v2-20260709.html` | ✓ exists (53327 bytes) | Design SSOT |
| Design report (v1) | `.moai/reports/agent-architecture-redesign-20260709.html` | ✓ exists (39827 bytes) | (v1 — superseded by v2) |
| Memory | `project_model_tier_plantype_001_completed.md` | ✓ read in full | `ApplyTierProfile` is live (60-cell profile) |
| Memory | `project_agent_arch_v2_plan_handoff.md` | ✗ file not found at the path attempted | (likely named differently — the MEMORY.md index entry exists; content not needed since design report is SSOT) |

**Verdict**: source ready. **Caveat REQ-DA-061 required**: the page MUST distinguish the v2 design report (architectural intent) from the live `ApplyTierProfile` behavior (implemented per SPEC-MODEL-TIER-PLANTYPE-001). Readers must be able to tell what is design vs what ships today.

### B.4 plan-type-profiles — READY (authoritative; with GLM honesty caveat)

| Source | Path | Verified |
|--------|------|----------|
| SPEC (CLOSED) | `.moai/specs/SPEC-MODEL-TIER-PLANTYPE-001/` | ✓ exists (5 artifacts: spec/plan/acceptance/progress/research) | authoritative |
| Design report | `.moai/reports/model-tier-redesign-20260712.md` | ✓ exists (10437 bytes) | md twin (agent context canonical) |
| Design report HTML | `.moai/reports/model-tier-redesign-20260712.html` | ✓ exists (15425 bytes) | (HTML twin) |
| Memory | `project_model_tier_plantype_001_completed.md` | ✓ read in full | contains the HARD honesty constraint |

**Verdict**: source authoritative and ready. **Caveat REQ-DA-060 required**: the GLM backend effort overlay's wire-effectiveness is UNVERIFIED. The page MUST say "구현 + 배선 완료, wire 유효성 실증 예정" / "implemented + wired, wire validity pending live verification", NOT "works guaranteed" — per the memory's §HARD honesty constraint.

### B.5 self-evolving — READY (with roadmap caveat)

| Source | Path | Verified |
|--------|------|----------|
| Design report v5.1 FINAL SSOT | `.moai/reports/harness-self-evolving-redesign-final-20260712.html` | ✓ exists (36508 bytes) | canonical SSOT |
| Design report v1 | `.moai/reports/harness-self-evolving-redesign-20260712.html` | ✓ exists (40203 bytes) | (v1 — superseded by final) |
| SPEC (closed) | `.moai/specs/SPEC-HARNESS-EVOLVE-{001,002,003}/` | ✓ all 3 exist | Loop 0/1/2 implemented layers |
| Memory | `project_harness_evolve_epic.md` | ✓ read in full | 3/5 closed + EVOLVE-004/005 pending |
| Reference | `reference_lilian_weng_harness_selfimprove.md` | (in index) | Weng original |

**Verdict**: source substantially ready. **Caveat REQ-DA-063 required**: disclose that EVOLVE-004 (console verbs `/moai harness evolve/promote/demote/freeze`) and EVOLVE-005 (Recall wiring + typed parser + template distribution) are PENDING — not yet implemented.

### B.6 autonomous-loops — READY (with distinction caveat)

| Source | Path | Verified |
|--------|------|----------|
| Doctrine (canonical /goal) | `.claude/rules/moai/workflow/goal-directive.md` | ✓ (always-loaded; full content in agent context) | HUMAN-ONLY TUI semantics + MoAI integration notes |
| SPEC (CLOSED) | `.moai/specs/SPEC-GOAL-ENGINE-001/` | ✓ exists (8 files) | `moai goal arm/status/clear` CLI + session-start PruneOrphans |
| Memory | `project_agentic_core_epic_progress.md` | (in index) | AGENTIC-CORE epic (SPEC-1 closed, SPEC-2 pending) |
| Memory | `project_goal_engine_cli_gap_handoff.md` | (in index, ✅ close marker) | GOAL-ENGINE-001 close |

**Verdict**: source substantially ready. **Caveat REQ-DA-062 required**: distinguish native `/goal` (HUMAN-ONLY TUI command the model cannot invoke on the user's behalf) from `/moai goal` (programmatic, MoAI-owned, Axis B per `native-invocation-model.md`) from `/moai loop` (Ralph Engine diagnostic preset — NOT an alias for `/moai run --mode loop` despite the dispatch-axis crosswalk).

### B.7 Source-readiness summary

| Page | Source status | Honesty caveat |
|------|---------------|----------------|
| tokenomics-overview | READY | none |
| token-budget | READY | none |
| no-haiku-3tier | READY | REQ-DA-061 (design vs live) |
| plan-type-profiles | READY (authoritative) | REQ-DA-060 (GLM wire validity) |
| self-evolving | READY | REQ-DA-063 (EVOLVE-004/005 roadmap) |
| autonomous-loops | READY | REQ-DA-062 (native vs programmatic) |

**All 6 pages have substantially-ready sources — NO blockers, NO pages deferred to a later SPEC.**

---

## §C Memory carry-over risk assessment

The spawn prompt cited 3 applied-lessons contexts. Their relevance to this SPEC:

### C.1 `feedback_defect_claim_verification`

**Lesson**: a defect claim is a hypothesis until the domain's tool confirms it. The "db section follow-up" incident (memory carry-over claimed the section was already done; verification revealed otherwise).

**Application**: the agent verified the parity-debt state by direct `find` + Read of all 4 `_meta.yaml` files at plan-phase. The brief's claim ("KO ~11, EN/JA/ZH ~8") was correct on the surface but UNDERSTATED the gap — KO itself is missing 3 more entries. This finding is recorded in §A.2 above.

**Run-phase obligation**: manager-develop MUST re-verify the parity state at M0 (pre-flight) before assuming the plan-phase numbers still hold. If a parallel session has already partially closed the debt, re-measure and adjust M1 scope accordingly.

### C.2 `feedback_manager_docs_sync_phase_discipline`

**Lesson**: sync-phase CHANGELOG hallucination + `git add -A` kitchen-sink.

**Application**: this is a plan-phase SPEC, so sync-phase discipline binds the LATER manager-docs sync invocation, not this plan-phase authoring. Recorded in plan.md §B.12 (B12 CHANGELOG discipline) and §D.2 (forbidden `git add -A`).

### C.3 `feedback_shared_checkout_concurrent_commit_race`

**Lesson**: shared-checkout concurrent commit race; specific-path commit is one defense; only `git show <ref>:<path>` is fully trustworthy.

**Application**: the per-locale `_meta.yaml` is shared across all advanced-page SPECs. If a parallel session (e.g., the stalled SPEC-DESKTOP-NATIVE-E2E-001 referenced by the SubagentStart hook) edits the same `_meta.yaml`, race conditions apply. Recorded in plan.md §B.14 (pre-spawn sync check) and §B.15 (_meta.yaml merge hazard).

---

## §D Open questions / assumptions

### D.1 [ASSUMPTION] — Route A (Hybrid Trunk main-direct)

The plan assumes Route A (direct-to-main, no PR) per CLAUDE.local.md §23 (1-person OSS Hybrid Trunk). This is the default for Tier S/M; Tier L defaults to Route B (PR route) per spec-workflow.md. **This SPEC is Tier L but docs-only** — the recommendation is Route A unless the user explicitly requests `--pr`. This is recorded as plan.md §A.1 recommendation; the orchestrator's Implementation Kickoff Approval round (CLAUDE.local.md §19.1) confirms the route.

### D.2 [ASSUMPTION] — Title translations

The 4-locale title table in design.md §D is the manager-spec's best-effort translation. A native-speaker review (especially for ja/zh) is RECOMMENDED at run-phase — the locale-translator specialist may refine the strings during derivation. The titles are not AC-binding (CMD-D checks for the slug's presence + 4-locale name field existence, not for the exact string).

### D.3 [ASSUMPTION] — No new icon variants

The icon-shortcode palette (check/check-circle/x/x-circle/warning/info/bulb/rocket/star/flash/sparkles/target/package/book/search/wrench/database/rotate/clock/arrow-right) is treated as fixed. If a page genuinely needs an icon outside this set, the page author substitutes the closest semantic match. Adding new icons to `layouts/shortcodes/icon.html` is OUT OF SCOPE (per spec.md §E exclusions).

### D.4 [NEEDS CLARIFICATION: none]

At plan-phase close, there are NO `[NEEDS CLARIFICATION]` markers requiring orchestrator AskUserQuestion rounds before Implementation Kickoff Approval. All design decisions are resolved in design.md §F; all sources are verified ready; the parity-debt fix approach is mechanical.

The Implementation Kickoff Approval gate is the sole user-decision point — it confirms:
1. Route A (direct-to-main) vs Route B (PR)
2. Tier L classification acceptance
3. The 6-page scope and pillar grouping
4. The M1 parity-debt pre-fix approach

---

## §E Cross-references

- **spec.md** §A.2 — the parity-debt claim (this file §A is the evidence)
- **spec.md** §A.3 — the source-readiness claim (this file §B is the evidence)
- **plan.md** §B.13 — memory carry-over discipline (this file §C is the evidence)
- **plan.md** §C — pre-flight commands (re-verification of this file §A.1 + §B at run-phase)
- **design.md** §C — page-by-page outlines (sourced from this file §B per-page)
- **acceptance.md** CMD-A through CMD-J — the runnable exit-gate recipe (verifies the claims in this file)

---

Version: 0.1.0 | Tier: L | Status: draft (plan-phase) | Verified at: 2026-07-13
