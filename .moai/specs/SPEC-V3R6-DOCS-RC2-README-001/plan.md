---
id: SPEC-V3R6-DOCS-RC2-README-001
title: "Implementation Plan — v3.0.0-rc2 README + CHANGELOG factual-alignment"
version: "0.3.0"
status: in-progress
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "repo-root docs"
lifecycle: spec-anchored
tags: "docs, readme, changelog, plan, v3r6, rc2"
era: V3R6
tier: M
---

# Implementation Plan — SPEC-V3R6-DOCS-RC2-README-001

## §A. Context

This plan implements the GEARS requirements in `spec.md` §D as a sequence of documentation edits across 4 repo-root files. The work is **prose/diagram-only** — zero Go code changes. The drift inventory (`spec.md` §C) provides exact `file:line` evidence for every edit.

The plan is structured as 6 milestones (M1..M6). Milestones are ordered by surface (EN README → Mermaid/db → KO README → CHANGELOG → CLAUDE.md → self-verification) so that each milestone produces a coherent, independently-committable doc slice.

### A.1 Sequencing Rationale

- **M1 before M2**: M1 fixes the easy quantitative tokens (grep-replaceable); M2 rewrites the Mermaid diagram and removes the `/moai db` section (structural edits). Splitting them keeps each commit reviewable.
- **M3 after M1/M2**: The KO README mirrors the EN README; doing EN first establishes the canonical wording KO mirrors.
- **M4 (CHANGELOG) independent**: CHANGELOG edits do not depend on README state.
- **M5 (CLAUDE.md) independent**: Single path-fix.
- **M6 last**: Self-verification runs the grep-based AC suite from `acceptance.md` across all 4 files.

---

## §B. Known Issues (Pre-Existing)

- **B.1** — The local `.moai/config/sections/llm.yaml` carries the stale `glm-5.1` value; the TEMPLATE SSOT (`internal/template/templates/.moai/config/sections/llm.yaml`) carries `glm-5.2[1m]`. Docs follow the TEMPLATE SSOT value. This SPEC does NOT fix the local llm.yaml drift — that is a config-alignment concern outside this SPEC's scope (see §H Exclusions).
- **B.2** — The `CLAUDE.md` template counterpart at `internal/template/templates/CLAUDE.md` may carry the same stale builder-harness path. This plan does NOT edit the template; see §G for the run-phase decision.
- **B.3** — The KO "What's New" section (DRIFT-KO-02) is the largest staleness surface (frozen at v2.17.0, two major versions behind). The rewrite is the highest-risk edit because it is prose-heavy rather than grep-replaceable.
- **B.4 (iter-1 D6)** — **Vacuously-satisfied AC hazard**: the authoritative values this SPEC aligns to (8 retained agents, 16 languages, manager-develop, 3-phase lifecycle) ALREADY appear correctly in OTHER parts of README.md/README.ko.md today (verified LIVE 2026-06-19). A whole-file `grep -c <new-value>` would therefore pass WITHOUT touching the drift line. Every affected AC in acceptance.md has been line-anchored (`sed -n '<line>p' | grep`); the run-phase agent MUST honor the line-anchor contract and re-resolve line numbers if earlier edits shift content (see acceptance.md edge case E.6).

---

## §C. Pre-flight Checks (Before M1)

Run these read-only verifications to confirm the baseline before any edit:

1. **Version baseline**: `grep -E "^    version:" .moai/config/sections/system.yaml` → must return `v3.0.0-rc2`.
2. **Git tag baseline**: `git tag --list "v3*"` → must include `v3.0.0-rc2`, must NOT include `v3.0.0` (bare stable).
3. **Agent count baseline**: `ls .claude/agents/moai/*.md | wc -l` → must return 7 (the 7 MoAI-custom agents).
4. **builder-harness path baseline**: `test -f .claude/agents/moai/builder-harness.md && echo OK` → must print OK; `test -d .claude/agents/builder && echo EXISTS || echo MISSING` → must print MISSING.
5. **Stale-token presence baseline**: capture the pre-edit grep counts for every stale token in `spec.md` §C (these become the "before" snapshot for the §E self-verification).
6. **Skills count baseline (iter-3 D1' fix)**: `find .claude/skills -name SKILL.md -path "*moai*" | wc -l` → returns 33; `find .claude/skills -name SKILL.md -path "*moai*" | grep -v "harness-moaiadk-" | wc -l` → returns **31**. The README surfaces the 31 count (moai-* template-managed, excluding 2 `harness-moaiadk-*` user-owned). Root cause for the iter-1→iter-3 delta (32→31): `moai-design-system` skill was removed by `SPEC-V3R6-DESIGN-SYSTEM-RETIRE-001` after iter-1 authoring. Re-derive LIVE; do NOT trust the numbers in this plan.
7. **Go LOC baseline (iter-3 D3 fix)**: the §B pipeline (`find . -name "*.go" -not -path "./vendor/*" -not -path "./internal/template/embedded.go" -not -name "*.pb.go" -not -name "zz_*.go" | xargs wc -l | tail -1`) over-counts non-deterministically (iter-3 LIVE 2026-06-19 = 191,248; earlier drafts cited ~193,616 / ~198,945). README MUST NOT hardcode any precise figure; use "100K+ lines" graceful-aging phrasing per spec.md §E.3.
8. **CHANGELOG v2.20.0-rc1 baseline (iter-1 D4 fix)**: `grep -cE '^##.*v2\.20\.0-rc1' CHANGELOG.md` → returns **4** heading lines (L378/L487/L507/L527); `grep -c "v2.20.0-rc1" CHANGELOG.md` → returns 28 total (4 headings + 24 body). M4 consolidates the 4 HEADING lines; body prose inside consolidated content is acceptable.

---

## §D. Constraints (Carried from spec.md §E)

- D.1 — Project-owned files; template-internal-content isolation does NOT bind.
- D.2 — `CLAUDE.md` template counterpart decision deferred to run-phase (§G).
- D.3 — Graceful-aging phrasing preferred for Go scale.
- D.4 — `v3.0.0-rc2` wording only; no `v3.0.0 stable`.
- D.5 — Every AC grep-verifiable.
- D.6 — EN/KO quantitative parity.

---

## §E. Milestones

### M1 — README.md (EN) quantitative fixes

**Scope**: DRIFT-EN-01, DRIFT-EN-02, DRIFT-EN-03, DRIFT-EN-04 + statusline FAQ refresh (DRIFT-EN-09).

**REQs covered**: REQ-EN-001, REQ-EN-002, REQ-EN-003, REQ-EN-004, REQ-EN-005, REQ-EN-010.

**Edits**:
- `README.md:40` hero line — replace "24 specialized AI agents and 52 skills" with "8 retained AI agents and 31 moai-* skills" (state the harness-moaiadk-* exclusion explicitly per REQ-EN-001; final wording at run-phase discretion but the 31 count MUST be qualified as moai-* template-managed and the harness-moaiadk-* exclusion noted).
- `README.md:62` — replace "38,700+ lines / 38 packages" with graceful-aging phrasing "100K+ lines of Go across 100+ packages". Do NOT hardcode the precise LOC figure (the §B pipeline over-counts non-deterministically; iter-3 LIVE = 191,248; drifts on every commit).
- `README.md:64` Key Numbers — replace "26 agents + 47 skills" → "8 agents + 31 moai-* skills"; replace "18 languages" → "16 languages".
- `README.md:262` — replace "delegates to 24 specialized agents" → "delegates to 8 retained agents". (NOTE per D6: L297 already has a different "8 retained agents" token — the edit target is L262 specifically; the line-anchor AC verifies L262.)
- `README.md:1235,1245,1281,1289` — refresh statusline FAQ illustrative versions to Opus 4.7+ / CC 2.1.17x+ OR mark explicitly illustrative. (Concrete AC predicate: `grep -cE "explicitly illustrative|example only" README.md` ≥1 OR `grep -cE "Opus 4\.[78]|CC 2\.1\.17" README.md` ≥1.)

**Verification**: AC-EN-001 through AC-EN-005, AC-EN-010 (grep-based, see acceptance.md).

**Commit subject (run-phase)**: `docs(SPEC-V3R6-DOCS-RC2-README-001): M1 README.md quantitative factual-alignment`

---

### M2 — README.md (EN) Mermaid rewrite + /moai db removal + expert-frontend replacement + lifecycle note

**Scope**: DRIFT-EN-05, DRIFT-EN-06, DRIFT-EN-07, DRIFT-EN-08.

**REQs covered**: REQ-EN-006, REQ-EN-007, REQ-EN-008, REQ-EN-009.

**Edits**:
- `README.md:264-286` — rewrite AI Agent Orchestration Mermaid (fence L264, close L286) to show only the 8 retained agents (manager-spec, manager-develop, manager-docs, manager-git, plan-auditor, sync-auditor, builder-harness, Explore). The current block uses BARE category labels `Manager (8)`, `Expert (8)`, `Builder (3)`, `Evaluator (2)`, `Design System (4+1)` and a node body `backend · frontend · security · devops<br/>performance · debug · testing · refactoring` — ALL of these stale bare labels + the archived-agent node body MUST be removed (the hyphenated `expert-(frontend|...)` form does NOT appear in the Mermaid block today; it appears only in the Design System table, so do NOT rely on that form for verification — see AC-EN-006a).
- `README.md:451-478` — add a short 3-phase lifecycle note (plan→run→sync, Mx retired per `SPEC-V3R6-LIFECYCLE-REDESIGN-001`). Optional: mention dynamic workflows + `/effort ultracode` per `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001`.
- `README.md:921,939,969` — replace `expert-frontend` in the Design System implementer table with `manager-develop` (cycle_type-based) or harness-generated specialist.
- `README.md:1111-1210` — remove the `/moai db` section. Replace with at most a one-line pointer: "Database schema tooling: see `moai hook db-schema-sync` in the CLI reference."

**Verification**: AC-EN-006, AC-EN-007, AC-EN-008, AC-EN-009.

**Commit subject (run-phase)**: `docs(SPEC-V3R6-DOCS-RC2-README-001): M2 README.md Mermaid + /moai db + expert-frontend + lifecycle note`

---

### M3 — README.ko.md (KO) mirror + What's New rewrite

**Scope**: DRIFT-KO-01 through DRIFT-KO-05.

**REQs covered**: REQ-KO-001, REQ-KO-002, REQ-KO-003.

**Edits**:
- Mirror all M1+M2 EN fixes at the KO equivalent line locations (agent count, Go scale, language count, Mermaid, `/moai db`, `expert-frontend`).
- `README.ko.md:46-90` — REWRITE the "## v2.17.0의 새로운 기능" section to the v3/V3R6 generation. New section title: "## v3.0.0-rc2 (V3R6)의 새로운 기능" (or equivalent). Content must cover: 8-agent retained catalog, `glm-5.2[1m]` model, 3-phase lifecycle (plan→run→sync, Mx retired), CG mode default, dynamic workflows, `/effort ultracode`.
- `README.ko.md:58` — replace "my-harness-*" with "moai-* (template-managed) vs harness-* (user-owned)" per `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`.

**Verification**: AC-KO-001, AC-KO-002, AC-KO-003.

**Commit subject (run-phase)**: `docs(SPEC-V3R6-DOCS-RC2-README-001): M3 README.ko.md mirror + What's New v3 rewrite`

---

### M4 — CHANGELOG.md rc2 promotion + mis-labeled header cleanup

**Scope**: DRIFT-CL-01, DRIFT-CL-02.

**REQs covered**: REQ-CL-001, REQ-CL-002, REQ-CL-003.

**Edits**:
- Promote the top `## [Unreleased]` block to `## [v3.0.0-rc2] — 2026-06-19` (or the actual rc2 date; confirm via `git log v3.0.0-rc2 -1 --format=%ci` at run-phase). Content must reflect the rc2 V3R6 cohort: 3-phase lifecycle, harness namespace V2, runtime recovery doctrine, orchestrator interrupt ledger, `glm-5.2[1m]`, dynamic workflows, 8-agent catalog.
- Consolidate / remove the **4** mis-labeled `## [Unreleased] — v2.20.0-rc1` subsection HEADING lines (LIVE 2026-06-19: L378/L487/L507/L527 — NOT 3 as earlier drafts claimed) so no `^##.*v2\.20\.0-rc1` heading line remains. Merge their body content under the correct `## [v3.0.0-rc2]` section or a fresh `## [Unreleased]`. Body prose that legitimately references the historical milestone inside consolidated content is acceptable; the AC greps the heading-line form `^##.*v2\.20\.0-rc1` only (NOT bare `v2.20.0-rc1`, which would over-match the 24 body mentions — see AC-CL-002).
- Do NOT create a `## [v3.0.0]` stable section (REQ-CL-003).

**Verification**: AC-CL-001, AC-CL-002, AC-CL-003.

**Commit subject (run-phase)**: `docs(SPEC-V3R6-DOCS-RC2-README-001): M4 CHANGELOG rc2 promotion + mis-label cleanup`

---

### M5 — CLAUDE.md §4 builder-harness path correction

**Scope**: DRIFT-CLAUDE-01.

**REQs covered**: REQ-CLAUDE-001.

**Edits**:
- `CLAUDE.md` §4 Agent Catalog — replace `.claude/agents/builder/builder-harness.md` with `.claude/agents/moai/builder-harness.md`.

**Verification**: AC-CLAUDE-001.

**Commit subject (run-phase)**: `docs(SPEC-V3R6-DOCS-RC2-README-001): M5 CLAUDE.md builder-harness path fix`

---

### M6 — Self-verification (grep-based AC suite)

**Scope**: Run the full acceptance.md AC suite across all 4 files.

**REQs covered**: Cross-cutting verification (REQ-X-001, REQ-X-002, REQ-X-003) + all per-surface REQs.

**Actions**:
1. Run every `grep -c <stale-token>` assertion from acceptance.md → all must return 0.
2. Run every `grep -c <new-value>` assertion from acceptance.md → all must return ≥1.
3. Run the cross-cutting version-wording check (no `v3.0.0 stable` anywhere in the 4 files).
4. Run the cross-cutting superseded-SPEC check (no `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` citation in the 4 files).
5. Run `moai spec lint` on this SPEC → must return 0 findings.
6. Capture the §E.2 run-phase evidence (grep command + output verbatim) for the progress.md §E.2 section.

**Commit subject (run-phase)**: `docs(SPEC-V3R6-DOCS-RC2-README-001): M6 self-verification evidence`

---

## §F. Technical Approach

### F.1 Edit strategy

- **Grep-replaceable tokens** (agent counts, LOC, language count, builder-harness path, `my-harness-*`): direct `Edit` tool calls with unique `old_string` anchors from the §C drift inventory.
- **Mermaid rewrite** (DRIFT-EN-05): single `Edit` replacing the whole diagram block (L264-286). The replacement diagram must be validated for Mermaid syntax (no orphan edges, valid node shapes).
- **Prose rewrites** (KO What's New DRIFT-KO-02, CHANGELOG rc2 block DRIFT-CL-01): `Edit` replacing the section body. These are the highest-judgment edits; the run-phase agent has discretion over final wording provided the §B baseline values are honored.
- **Section removal** (`/moai db` DRIFT-EN-08): `Edit` removing the L1111-1210 block and substituting the one-line pointer.

### F.2 Parallelization

- M1 and M4 and M5 are independent (different files, no shared state) — they MAY be executed in parallel by separate agent delegations IF the orchestrator chooses team mode.
- M2 depends on M1 conceptually (both edit README.md) — execute sequentially to avoid Edit conflicts on the same file.
- M3 depends on M1+M2 (KO mirrors EN canonical wording) — execute after M2.
- M6 depends on all prior milestones.

**Recommended serial path**: M1 → M2 → M3 → (M4 ∥ M5) → M6.

### F.3 Rollback

Each milestone is a single commit. If any AC fails at M6, the offending milestone commit can be reverted independently without unwinding the whole SPEC. The git history provides per-milestone rollback granularity.

---

## §G. Risks & Mitigations

### G.1 CLAUDE.md template counterpart (M5 risk)

**Risk**: `internal/template/templates/CLAUDE.md` may carry the same stale `.claude/agents/builder/builder-harness.md` path. If we fix only the project-owned `CLAUDE.md`, the template SSOT remains stale and `moai init` / `moai update` will regenerate the bug.

**Mitigation (run-phase decision, NOT plan-phase)**: The run-phase agent MUST grep `internal/template/templates/CLAUDE.md` for the stale path. If found, open a follow-up decision via the orchestrator's AskUserQuestion channel:
- Option A: Fix the template in the same SPEC (expands scope; requires template-neutrality CI guard pass).
- Option B: Defer the template fix to a separate template-only SPEC (keeps this SPEC scoped to project-owned docs).

This plan does NOT pre-decide; the run-phase agent surfaces the finding.

### G.2 Mermaid syntax breakage (M2 risk)

**Risk**: The rewritten diagram may use invalid Mermaid syntax (orphan edges, undeclared nodes), breaking the GitHub render.

**Mitigation**: M2 verification MUST include a visual/syntax sanity check. The AC for REQ-EN-006 includes a grep assertion that all 8 retained agent names appear as declared nodes AND that zero archived agent names appear.

### G.3 KO What's New prose drift (M3 risk)

**Risk**: The KO rewrite is prose-heavy and may inadvertently introduce a new unverified claim (e.g., asserting a feature that is not actually in rc2).

**Mitigation**: The KO rewrite MUST cite only features whose authority is in the §B baseline or in the cited completed SPECs (`SPEC-V3R6-LIFECYCLE-REDESIGN-001` code-landed, `SPEC-V3R6-HARNESS-NAMESPACE-V2-001`, `SPEC-V3R6-WORKFLOW-EFFORT-MAP-001`). REQ-X-003 binds: draft SPECs (RULES-*) MUST be phrased as in-flight if mentioned.

### G.4 Graceful-aging phrasing ambiguity (M1 risk)

**Risk**: "100K+ lines" is a range; a future reader may not know whether the range is still accurate.

**Mitigation**: The range is intentionally loose ("100K+") so it ages gracefully. The AC checks for the range token, not a precise figure. The next drift sweep re-evaluates.

### G.5 Multi-session race on shared main (process risk)

**Risk**: Per the `feedback_shared_main_orphan_race` memory, parallel Claude Code sessions on the shared main working tree can orphan unpushed commits. Doc edits on `README.md` / `CHANGELOG.md` are exactly the kind of shared-file edit vulnerable to this race.

**Mitigation**: The orchestrator MUST run the Pre-Spawn Sync Check (`agent-common-protocol.md` § Pre-Spawn Sync Check) before any M1..M5 delegation. `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` MUST return `0 N` or `0 0` before spawn.

---

## §H. Cross-References

- `spec.md` §B (Ground Truth baseline), §C (Drift Inventory with file:line evidence), §D (GEARS REQs), §E (Constraints).
- `acceptance.md` §D (AC Matrix — the grep-based verification suite M6 runs).
- `.claude/rules/moai/development/sprint-round-naming.md` — this is a single-SPEC 6-milestone plan (no Sprint/Round split).
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` — era classification (this SPEC sets `era: V3R6`).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1 + §5 — no-unobserved-claim invariant (binds REQ-X-001, REQ-CL-003, and the G.3 mitigation).

---

## §I. Self-Verification Checklist (Plan-phase)

- [x] SPEC ID decomposition self-check printed in spec.md §J (`→ PASS`).
- [x] All 12 canonical frontmatter fields present in spec.md.
- [x] `era: V3R6` set in spec.md frontmatter (avoids H-6 unclassified fallback).
- [x] Every REQ cites a DRIFT-ID from §C, and every DRIFT-ID cites a file:line.
- [x] §H Exclusions contains ≥1 `### Out of Scope — <topic>` H3 with `-` bullets (satisfies `OutOfScopeRule`).
- [x] No Go code edits in any milestone.
- [x] No `internal/template/templates/**` edits in any milestone (G.1 deferred to run-phase).
- [x] `v3.0.0-rc2` wording consistent across plan; no `v3.0.0 stable` claim.
- [x] No citation of superseded `SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001` (REQ-X-002).
- [x] Draft SPECs (RULES-*) not asserted as finalized (REQ-X-003).
- [x] **iter-1 audit (2026-06-19, v0.2.0)**: all §B baseline numbers re-derived LIVE (skills 32 moai-* excl. 2 harness-moaiadk-*; LOC ~193,616; v2.20.0-rc1 = 4 headings; Mermaid bare labels `Manager (8)`/`Expert (8)`/`Builder (3)`/`Evaluator (2)`/`Design System (4+1)` confirmed at L264-286). Every "new value present" AC where the token pre-exists elsewhere is line-anchored (D2/D3/D6). AC-EN-010 carries a concrete grep predicate (D7). v2.20.0-rc1 AC scoped to heading-line form (D4). Go LOC figure dropped from README in favor of graceful-aging phrasing (D5).
- [x] **iter-3 plan-audit fix (2026-06-19, v0.3.0)**: plan-auditor iter-2 re-audit FAIL 0.74 (2 BLOCKING + 2 SHOULD-FIX). D1' skills 32→31 LIVE (DSR rebase deleted `moai-design-system`). D2' AC-KO-001b line-anchored on `README.ko.md:111` (KO D6 discipline parity). D3 precise Go LOC dropped from §B (pipeline over-counts; 191,248 LIVE non-deterministic). D5 AC-CL-001 contradiction split into binary checks. plan.md §C.6/§C.7/M1 edits updated in lockstep.
