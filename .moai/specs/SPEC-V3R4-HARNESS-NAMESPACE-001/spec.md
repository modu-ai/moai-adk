---
id: SPEC-V3R4-HARNESS-NAMESPACE-001
title: "Harness Namespace + Lifecycle Governance"
version: "0.4.0"
status: completed
created: 2026-05-16
updated: 2026-05-16
author: manager-spec
priority: P1
phase: "v3.0.0-rc1 — Harness governance closeout"
module: ".moai/harness/, .moai/specs/SPEC-V3R4-HARNESS-*/, .claude/skills/moai/workflows/harness.md, .claude/commands/moai/harness.md, internal/cli/harness.go (deprecation marker only), .moai/SPEC-NAMING-CONVENTION.md (governance reference)"
lifecycle: spec-anchored
tags: "harness, namespace, governance, v3r4, lifecycle, lint, docs-only, breaking-none, target-rc1"
dependencies:
  - SPEC-V3R4-HARNESS-001
related_specs:
  - SPEC-V3R4-HARNESS-001
  - SPEC-V3R4-HARNESS-002
  - SPEC-V3R4-HARNESS-003
  - SPEC-V3R3-HARNESS-001
  - SPEC-V3R3-HARNESS-LEARNING-001
  - SPEC-V3R3-PROJECT-HARNESS-001
  - SPEC-V3R4-SPECLINT-DEBT-002
  - SPEC-V3R4-SDF-001
breaking: false
related_theme: "Harness governance closeout — namespace + lifecycle + slash-only contract"
target_release: v3.0.0-rc1
---

# SPEC-V3R4-HARNESS-NAMESPACE-001 — Harness Namespace + Lifecycle Governance

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.4.0   | 2026-05-16 | MoAI orchestrator | **Sync-phase lifecycle COMPLETE.** Run PR #945 merged into main `571258a1e` (2026-05-16T05:19:54Z). Sync PR #N (this commit) closes lifecycle with `sync(spec):` prefix per lesson #16 walker filter requirement. Status `implemented → completed`. All 8 ACs (AC-HRN-NS-001~008) binary PASS verified across Wave 1 + Wave 2 audit reports (`.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave{1,2}-2026-05-16.md`). PR #908 closeout final (CLOSED + branch deleted via user-selected Option 1 rollback-tip-then-close). v3.0.0-rc1 governance authority for downstream SPEC-V3R4-HARNESS-{004..008} namespace compliance now permanently in force. Lifecycle precedents established: (1) governance-only SPECs ship zero LOC code change via Wave-decomposed audit reports + EC-001 escalation pattern for PR closeout, (2) `sync(spec):` prefix mandatory per lesson #16 to engage walker filter, (3) plan-in-main + worktree-ban policy per feedback_worktree_never_use. SPEC-V3R4-HARNESS-NAMESPACE-001 file family (5 artifacts: acceptance/design/plan/spec/tasks) now FROZEN as governance reference for harness family. |
| 0.3.0   | 2026-05-16 | MoAI orchestrator | Run-phase status transition `draft → implemented`. Wave 1 (7 governance verification tasks T-Wave1-001~007) all PASS — SPEC ID regex 0 violations across 7 SPECs, dependency graph acyclic (4 V3R4-HARNESS-* SPECs lifecycle-independent), `.moai/harness/` hierarchy canonical subset (main.md + README.md + usage-log.jsonl), `moai spec lint --strict` 0 errors (4 → 3 warnings after this transition; 3 remaining are pre-existing V3R4-HARNESS-001/002/003 drift unrelated to this SPEC), CLI deprecation grace re-asserted (MARKER_PRESENT + NO_REGISTRATION confirmed in `internal/cli/harness.go` 433 lines + `internal/cli/root.go` clean), CHANGELOG v3.0.0-rc1 Governance entry appended, Wave 1 audit report persisted at `.moai/reports/governance/`. Wave 2 (5 PR #908 closeout tasks) — EC-001 escalation path EXPECTED at HEAD `a41d6d139c8c` (2-commit divergence from absorbed tip `452aa638f`), user-selected Option 1 (rollback-tip-then-close): branch reset to `452aa638f` + force-with-lease push + attribution comment + close-with-delete-branch; harness.md verb-surface re-verified at exactly 4 verbs (status/apply/rollback/disable, no drift); auto-memory entry persisted; Wave 2 audit report appended. All 8 ACs satisfied (AC-HRN-NS-001~008 PASS binary). Zero LOC code change. Governance authority for downstream SPEC-V3R4-HARNESS-{004..008} namespace compliance now in force. |
| 0.2.0   | 2026-05-16 | manager-spec | plan-auditor REVISE 0.69 round-1 fix. 4 Critical + 4 High defects + Q3 EC reorder resolved: (D-001) verb-surface regex corrected to match numbered headers `### 2.1 status` pattern (4 verified matches); (D-002) SPEC ID regex broadened to `^SPEC-V3R[0-9]+(-[A-Z][A-Z0-9]*)*-HARNESS(-[A-Z][A-Z0-9]*)?-[0-9]{3}$` to cover `SPEC-V3R3-PROJECT-HARNESS-001` (SCOPE-before-HARNESS); (D-003) lint invocation switched to no-args full-repo scan (CLI rejects directory args with ParseFailure); (D-004) PR #908 HEAD divergence (`a41d6d139` ≠ `452aa638f`) reframed as EXPECTED scenario with 2-commit truth + prefix matching; (D-005) REQ-HRN-NS-002 reframed to permit sibling deps as registered blockers (HARNESS-003 → -002 case from main HEAD); (D-006) `main.md` + `README.md` added to canonical reserved names matching observed `.moai/harness/` state; (D-007) CHANGELOG sketch enriched with retention keywords for AC-HRN-NS-008 grep PASS; (D-008) Glossary "PR #908 closeout" updated with 2-commit state + 4-option escalation. |
| 0.1.0   | 2026-05-16 | manager-spec | Initial draft. Pure governance SPEC formalizing the V3R4 harness namespace (SPEC ID convention, `.moai/harness/*` directory hierarchy, retention policy) and lifecycle independence guarantees for SPEC-V3R4-HARNESS-{001..008}. Complements (does NOT amend) SPEC-V3R4-HARNESS-001 foundation. Closes out PR #908 (`feat/cmd-harness-slash-wrapper`, OPEN since 2026-05-13) as fully absorbed by PR #910 commit `bb80ea0f4`. Re-asserts `/moai harness` 4-verb contract and AskUserQuestion gating from REQ-HRN-FND-003/004. Plan-in-main mode (no worktree). Zero LOC code change in run-phase — docs + governance verification only. |

---

## 1. Goal

The MoAI-ADK harness system has accumulated seven prior SPECs across V3R3 and V3R4 release cycles (three V3R3 SPECs superseded by V3R4-HARNESS-001, plus V3R4-HARNESS-{001, 002, 003} with 004-008 anticipated). The foundation SPEC SPEC-V3R4-HARNESS-001 already encodes the CLI-retirement contract (BC-V3R4-HARNESS-001-CLI-RETIREMENT), the slash-only verb surface (`status`, `apply`, `rollback`, `disable`), the 5-Layer Safety preservation, the FROZEN-zone immutability, and the AskUserQuestion gating for Tier-4 application. What it does NOT do — and what this SPEC formalizes — is the namespace governance layer that sits *across* the eight-SPEC family: (a) the canonical SPEC ID format `SPEC-V3R{N}-HARNESS-{SCOPE}-{NNN}` that every harness-family SPEC MUST follow, (b) the lifecycle-independence guarantee that prevents one harness SPEC's merge from blocking another, (c) the verbatim re-assertion of the `/moai harness` 4-verb surface as a stable governance contract that downstream SPECs MUST NOT extend without a new governance SPEC, (d) the `.moai/harness/*` directory hierarchy and file-naming convention with retention policy, and (e) the explicit closeout of PR #908 (`feat/cmd-harness-slash-wrapper`, OPEN since 2026-05-13) that has been fully absorbed into V3R4-HARNESS-001 run-phase via PR #910 commit `bb80ea0f4`. This SPEC ships ZERO code changes; run-phase deliverables are markdown updates (CHANGELOG, governance verification artifacts) and a manager-git PR closeout for #908.

### 1.1 Background

- SPEC-V3R4-HARNESS-001 (merged via PR #910, main commit `bb80ea0f4`, status `completed`) is the foundation SPEC. It explicitly declares seven downstream SPECs (002 through 008) as blocked by its merge — but it does NOT specify the cross-SPEC governance contract those downstream SPECs operate under. The absence of an authoritative namespace SPEC means downstream SPEC authors must each re-derive the naming pattern, directory hierarchy, and verb-surface stability guarantees by reading the foundation SPEC's §1, §2, and §11 — a duplication-prone pattern that motivated the SPECLINT-DEBT-002 SSOT (single source of truth) doctrine.
- Three V3R4 SPECs currently exist on disk: `SPEC-V3R4-HARNESS-001` (completed), `SPEC-V3R4-HARNESS-002` (plan/spec/research/tasks present, status TBD), and `SPEC-V3R4-HARNESS-003` (plan only, OPEN PR #923 carrying 9334 LOC for the embedding-cluster classifier). The remaining four downstream SPECs (004-008) are anticipated but not yet authored. Without a governance SPEC, each anticipated SPEC risks drifting from the foundation contract in subtle ways (verb name, file path, AskUserQuestion gating semantics).
- PR #908 (`feat/cmd-harness-slash-wrapper`, OPEN since 2026-05-13) currently has a **2-commit divergence** state as of plan-phase 2026-05-16:
  - **Commit 1** (`452aa638f` — slash thin-wrapper introducing `.claude/commands/moai/harness.md`): ALREADY imported and applied within V3R4-HARNESS-001 run-phase PR #910 (commit `bb80ea0f4`). Absorbed.
  - **Commit 2** (`a41d6d139c8c769bf395a25a055d59c14e180191` — `docs(audit): harness 서브시스템 end-to-end 진단 보고서`): pushed to the branch after Commit 1, NOT YET absorbed into main.
  - Branch HEAD as observed via `gh pr view 908 --json headRefOid`: `a41d6d139c8c769bf395a25a055d59c14e180191`.
  - The currently-OPEN PR #908 is therefore stale-with-residue, not simply duplicate. Leaving it OPEN risks (a) accidental re-merge of Commit 1's already-absorbed content + (b) loss of Commit 2's audit doc content if closed without disposition. This SPEC's run-phase MUST close PR #908 via one of the four escalation paths defined in EC-001, all of which are routed through orchestrator-issued `AskUserQuestion` (no force-close). The 2-commit state means EC-001 is the **EXPECTED** scenario at SPEC merge time, not an edge case.
- The V3R3 supersede chain (V3R3-HARNESS-001, V3R3-HARNESS-LEARNING-001, V3R3-PROJECT-HARNESS-001 → V3R4-HARNESS-001) was already declared in V3R4-HARNESS-001's `supersedes:` frontmatter. The status transition of those three V3R3 SPECs to `superseded` was deferred to a follow-up `manager-git` commit per V3R4-HARNESS-001 REQ-HRN-FND-013. This SPEC re-verifies that supersede consistency holds under `moai spec lint --strict` and documents the verification as a binary acceptance criterion.
- Recent lint debt SPECs (SPEC-V3R4-SPECLINT-DEBT-002 closing dual-schema drift via PR #943, SPEC-V3R4-SDF-001 closing 77-SPEC status drift via PR #940) have established a precedent for pure-governance SPECs that ship zero code change and produce binary-verifiable lint outcomes. This SPEC follows the same pattern.

### 1.2 User-Locked Governance Decisions

The following decisions are locked-in and MUST NOT be re-litigated within this SPEC or any downstream harness SPEC:

1. **SPEC ID format**: `SPEC-V3R{N}-HARNESS-{SCOPE}-{NNN}` where `{N}` is the release cycle digit (3 or 4 currently), `{SCOPE}` is an uppercase domain tag (`FND`, `LEARNING`, `PROJECT`, `NAMESPACE`, `EMBEDDING`, etc., or omitted for numbered-only IDs like `HARNESS-001`), and `{NNN}` is the zero-padded sequence number. The lint rule that enforces this format is `internal/spec/lint.go` `FrontmatterSchemaRule` (existing). This SPEC does not introduce a new lint rule; it documents the convention as binding for the harness family.
2. **Verb surface stability**: `/moai harness` supports exactly four verbs — `status`, `apply`, `rollback <YYYY-MM-DD>`, `disable`. Any extension requires a new governance SPEC that explicitly amends this list. Downstream SPECs MUST NOT silently introduce a fifth verb.
3. **Lifecycle independence**: Each SPEC in the V3R4-HARNESS-{001..008} family ships on its own lifecycle. The merge of one SPEC's plan PR, run PR, or sync PR MUST NOT block any other family member's lifecycle. Dependencies are declared via the `dependencies:` frontmatter field and validated by plan-auditor — not by ad-hoc cross-PR conventions.
4. **Namespace hierarchy**: `.moai/harness/*` is the only project-local harness state root. Files under `.claude/agents/my-harness/` and `.claude/skills/my-harness-*/` are project-generated specialist artifacts (preserved from V3R3-HARNESS-001 REQ-HARNESS-008 contract). Skill bodies under `.claude/skills/moai-*/` and agent bodies under `.claude/agents/moai/` remain FROZEN-zone and protected by V3R4-HARNESS-001 REQ-HRN-FND-006.
5. **PR #908 closeout**: PR #908 currently has 2-commit divergence (Commit 1 absorbed by PR #910, Commit 2 audit doc unabsorbed). The closeout disposition is orchestrator-decided via `AskUserQuestion` with the 4 EC-001 options (rollback-tip-then-close (Recommended) / cherry-pick-to-new-PR-then-close / abandon-new-commits / leave-open). The execution after user selection is owned by `manager-git`. No `feat/cmd-harness-slash-wrapper` content survives in the branch post-closeout (options 1, 2, 3); option 4 defers the closeout to a later SPEC.

### 1.3 Non-Goals

This SPEC is pure governance. The following are explicitly OUT OF SCOPE:

- Any code change to `internal/cli/harness.go`, `.claude/skills/moai/workflows/harness.md`, `.claude/commands/moai/harness.md`, or any other harness runtime artifact. The deprecation marker contract from V3R4-HARNESS-001 REQ-HRN-FND-001/002 remains in force unchanged.
- Any modification to SPEC-V3R4-HARNESS-001 (status `completed`, merged). Governance SPEC complements; it does NOT amend the foundation.
- Any new lint rule. The existing `FrontmatterSchemaRule` and walker-filter machinery (from SPECLINT-DEBT-002 and LINT-STATUS-CHORE-SKIP-001 lifecycles) already enforce the SPEC ID format and status consistency.
- Adding `SPEC-V3R4-HARNESS-{004..008}` as dependencies. Only `-001` (foundation) is a hard dependency; `-002` and `-003` are non-blocking related_specs because they exist on disk but are not foundationally required for governance.
- Any modification to `.claude/rules/moai/design/constitution.md` (FROZEN file).
- Any modification to the three superseded V3R3 SPEC files (V3R4-HARNESS-001 REQ-HRN-FND-013 prohibition is inherited).
- Re-litigating the CLI-retirement decision (BC-V3R4-HARNESS-001-CLI-RETIREMENT remains binding).
- Adding worktree, tmux, or Agent Teams configuration. Plan-in-main mode only (CLAUDE.local.md §18.12 BODP signals all-negative → main @ origin/main).
- GUI, dashboard, or non-text governance inspection surface.
- Migration of existing project state under `.moai/harness/`. The hierarchy documentation reflects the current layout; users with pre-existing state continue unchanged.

---

## 2. Scope

### 2.1 In Scope

- **SPEC ID format documentation**: Canonicalize `SPEC-V3R{N}-HARNESS-{SCOPE}-{NNN}` as the binding format for every harness-family SPEC. Document the format in this SPEC's §5 (REQ-HRN-NS-001) and verify via `moai spec lint --strict` PASS on the harness-family directory.
- **Lifecycle independence guarantee**: Document the contract that no harness SPEC's plan PR, run PR, or sync PR may block another family member's lifecycle. Verify via cross-reference of `dependencies:` frontmatter fields across `.moai/specs/SPEC-V3R4-HARNESS-*/spec.md` — no cyclic dependency, no transitive blocker.
- **Verb surface re-assertion**: Re-state the `/moai harness` 4-verb contract verbatim from V3R4-HARNESS-001 REQ-HRN-FND-003 as a governance-stable surface. Document that extension requires a new governance SPEC. AskUserQuestion gating for `apply` is re-asserted from V3R4-HARNESS-001 REQ-HRN-FND-004 (orchestrator-only invocation).
- **`.moai/harness/*` directory hierarchy**: Document the canonical structure (usage-log.jsonl, proposals/, learning-history/snapshots/, learning-history/applied/, learning-history/frozen-guard-violations.jsonl, learning-history/tier-promotions.jsonl) as a governance reference. Document retention policy: usage-log.jsonl is append-only with no rotation in this SPEC scope (rotation deferred); snapshots retain for the duration of the rate-limiter window (7 days minimum per V3R4-HARNESS-001 REQ-HRN-FND-012).
- **Supersede consistency verification**: Re-verify that the V3R3 → V3R4 supersede chain is consistent under `moai spec lint --strict`. The three V3R3 SPECs are declared superseded by V3R4-HARNESS-001's `supersedes:` field; this SPEC's run-phase verifies no lint regression.
- **PR #908 closeout**: Identify PR #908 (`feat/cmd-harness-slash-wrapper`, OPEN 2026-05-13) as stale/duplicate (content fully absorbed by PR #910 commit `bb80ea0f4`) and prescribe closeout via `manager-git` in run-phase. Document the closeout decision in CHANGELOG + memory entry. Branch `feat/cmd-harness-slash-wrapper` to be deleted post-closeout.
- **CLI deprecation grace window documentation**: Re-state the V3R4-HARNESS-001 BC-V3R4-HARNESS-001-CLI-RETIREMENT grace as a governance reference. This SPEC does NOT modify the grace window — it documents the current state for downstream SPEC authors.
- **v3.0.0-rc1 release alignment**: Declare this SPEC's `target_release: v3.0.0-rc1`. The governance artifacts (SPEC + CHANGELOG entry) ship as part of the v3.0.0-rc1 closeout cleanup.

### 2.2 Out of Scope

The Non-Goals listed in §1.3. In addition:

- Modifying any runtime code path. Run-phase deliverables are markdown only (SPEC artifacts already authored in plan-phase + CHANGELOG entry + manager-git PR closeout for #908).
- Authoring SPEC-V3R4-HARNESS-{004..008}. Those SPECs will be authored individually when their scope crystallizes; this SPEC documents the governance contract they MUST follow but does NOT create them.
- Introducing telemetry or metrics for namespace compliance beyond what `moai spec lint --strict` already provides.
- Networking, telemetry upload, or cross-machine governance synchronization.

---

## 3. Stakeholders

| Role | Interest |
|------|----------|
| MoAI-ADK maintainer | Single authoritative namespace governance SPEC; reduces per-SPEC cognitive load; clear convention for SPECs 004-008 authors. |
| Downstream SPEC author (V3R4-HARNESS-{002, 003, 004..008}) | Predictable naming, directory, and verb conventions; no re-derivation from foundation SPEC §1-§11; explicit lifecycle independence guarantee removes cross-PR coordination overhead. |
| `plan-auditor` subagent | Single governance SPEC to audit against; explicit cross-reference table for verb-surface stability and `.moai/harness/*` hierarchy. |
| `manager-git` subagent | Owns the PR #908 closeout in run-phase; clear delegation contract via REQ-HRN-NS-007. |
| `manager-spec` subagent | Future harness-family SPEC authoring is gated by this SPEC's namespace convention. |
| User upgrading from V3R3 | Slash command continues to work; namespace stability across the eight-SPEC family is a quality-of-life win. |
| v3.0.0-rc1 release cleanup owner | This SPEC's merge is a documented closeout asset feeding into the rc1 release notes. |

---

## 4. Exclusions (What NOT to Build)

[HARD] This SPEC explicitly EXCLUDES the following — building any of these within this governance PR is a scope violation:

1. Any modification to `internal/cli/harness.go`, `internal/cli/root.go`, or any Go file under `internal/cli/`. The CLI deprecation contract is owned by V3R4-HARNESS-001 REQ-HRN-FND-001/002 and remains in force unchanged.
2. Any modification to `.claude/skills/moai/workflows/harness.md` or `.claude/commands/moai/harness.md`. The workflow body is the verb-implementation surface; this SPEC documents the verb contract but does NOT change implementation.
3. Any modification to SPEC-V3R4-HARNESS-001 spec.md, plan.md, acceptance.md, tasks.md, or follow-up.md (foundation is `completed`, do not retroactively amend).
4. Any modification to the three superseded V3R3 SPEC files (inherited from V3R4-HARNESS-001 REQ-HRN-FND-013).
5. Any modification to `.claude/rules/moai/design/constitution.md` (FROZEN file).
6. Any new lint rule, new walker filter, or new linter machinery. The existing rule set (FrontmatterSchemaRule, drift walker, status consistency rule) already covers governance enforcement.
7. CLI re-registration. The slash-only contract is binding.
8. Authoring SPEC-V3R4-HARNESS-{004..008}. Those SPECs are anticipated but out of scope for this governance SPEC.
9. Any change to `.moai/harness/*` runtime files (usage-log.jsonl, learning-history/, proposals/). This SPEC documents the hierarchy but does NOT modify any user state.
10. Any change to `.claude/agents/my-harness/` or `.claude/skills/my-harness-*/` (V3R3-HARNESS-001 REQ-HARNESS-008 inviolate boundary).
11. Migration tooling that converts pre-existing project state into a new schema.
12. GUI, dashboard, web UI, or any non-text interface.
13. Networking, telemetry upload, or external API integration.
14. Worktree, tmux, or Agent Teams configuration in plan.md (plan-in-main only).

---

## 5. Requirements (EARS format)

Each requirement is identified by `REQ-HRN-NS-NNN` and uses one of the five EARS patterns: Ubiquitous (system shall always), Event-Driven (when X then system shall Y), State-Driven (while X the system shall Y), Optional Feature (where X exists the system shall Y), Unwanted Behavior (if X then system shall not / shall reject Y).

### REQ-HRN-NS-001 (Ubiquitous — SPEC ID format)

The harness-family SPEC ID **shall** follow a canonical format that permits both **HARNESS-leading** (`SPEC-V3R{N}-HARNESS-{SCOPE}?-{NNN}`) and **SCOPE-leading** (`SPEC-V3R{N}-{SCOPE}-HARNESS-{NNN}`) variants, where `{N}` is the release-cycle digit, `{SCOPE}` is zero or more uppercase domain tokens, and `{NNN}` is the zero-padded three-digit sequence number. Both variants are observed in the current corpus (`SPEC-V3R4-HARNESS-001`, `SPEC-V3R4-HARNESS-NAMESPACE-001`, `SPEC-V3R3-PROJECT-HARNESS-001`). The unified governance regex is `^SPEC-V3R[0-9]+(-[A-Z][A-Z0-9]*)*-HARNESS(-[A-Z][A-Z0-9]*)?-[0-9]{3}$`. The narrower `internal/spec/lint.go` `FrontmatterSchemaRule` `id` regex `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` (which covers both variants by being domain-agnostic) **shall** continue to be the underlying lint guard; this governance regex is the harness-family-specific verifier and is enforced by AC-HRN-NS-001 shell check, NOT by a new lint rule.

### REQ-HRN-NS-002 (Ubiquitous — declared-dependency lifecycle discipline)

The lifecycle of each SPEC in the `SPEC-V3R4-HARNESS-{001..008}` family **shall** be governed by explicit `dependencies:` frontmatter declarations: **either** (a) the SPEC declares only the foundation dependency (`SPEC-V3R4-HARNESS-001`) or an empty `[]`, making it independently merge-able regardless of sibling state; **or** (b) the SPEC declares one or more sibling dependencies (e.g., `SPEC-V3R4-HARNESS-003` declares `[-001, -002]`), in which case those siblings act as **registered merge blockers** — the dependent SPEC's run PR and sync PR **shall not** merge until each declared sibling has reached `status: implemented` or `status: completed`. Foundation `SPEC-V3R4-HARNESS-001` dependency is always permitted and is the canonical baseline. Ad-hoc cross-PR coordination outside the `dependencies:` field is prohibited; all blocking relationships **shall** be declarative and auditable via `plan-auditor` during plan-phase.

### REQ-HRN-NS-003 (Ubiquitous — verb surface re-assertion)

The `/moai harness` slash command **shall** expose exactly four verbs — `status`, `apply`, `rollback <YYYY-MM-DD>`, `disable` — as defined verbatim in V3R4-HARNESS-001 REQ-HRN-FND-003. The `apply` verb **shall** be gated by an orchestrator-issued `AskUserQuestion` round per V3R4-HARNESS-001 REQ-HRN-FND-004; subagents engaged in the lifecycle **shall not** invoke `AskUserQuestion` directly. Extension of this verb list (a fifth verb, a renamed verb, or a removed verb) **shall** require a new governance SPEC that explicitly amends this requirement.

### REQ-HRN-NS-004 (Ubiquitous — namespace hierarchy)

The project-local harness state **shall** be rooted at `.moai/harness/` with the following canonical reserved names (verified against actual project state as of plan-phase 2026-05-16): `README.md` (directory orientation, optional), `main.md` (harness instance metadata, V3R4-HARNESS-001 run-phase artifact), `usage-log.jsonl` (PostToolUse observation log, JSONL append-only), `proposals/` (Tier-X candidate proposal artifacts), `learning-history/snapshots/<ISO-DATE>/` (pre-application byte-identical snapshots per V3R4-HARNESS-001 REQ-HRN-FND-007), `learning-history/applied/` (applied evolution records), `learning-history/frozen-guard-violations.jsonl` (L1 Frozen Guard rejection audit log per V3R4-HARNESS-001 REQ-HRN-FND-014), and `learning-history/tier-promotions.jsonl` (Tier 1→2→3→4 promotion event log feeding the success-metric exposure of V3R4-HARNESS-001 REQ-HRN-FND-016). Project-generated specialist artifacts **shall** live under `.claude/agents/my-harness/` and `.claude/skills/my-harness-*/` (preserved from V3R3-HARNESS-001 REQ-HARNESS-008 inviolate boundary).

### REQ-HRN-NS-005 (Event-Driven — supersede consistency verification)

**When** `moai spec lint --strict` is invoked against `.moai/specs/SPEC-V3R4-HARNESS-*/` and `.moai/specs/SPEC-V3R3-HARNESS-*/`, the lint **shall** PASS with zero findings related to supersede consistency (no orphan `supersedes:` reference, no missing `superseded_by:` on superseded V3R3 SPECs, no status-transition drift). The lint PASS criterion **shall** be enforced as a binary acceptance criterion in this SPEC's `acceptance.md`.

### REQ-HRN-NS-006 (Ubiquitous — CLI deprecation grace re-assertion)

The CLI deprecation contract (`BC-V3R4-HARNESS-001-CLI-RETIREMENT`) defined in SPEC-V3R4-HARNESS-001 **shall** remain in force unchanged. The `internal/cli/harness.go` file **shall** remain in the tree as a deprecation marker without registration into the public cobra subcommand tree. Physical deletion of `internal/cli/harness.go` and `internal/cli/harness_test.go` **shall not** be performed within this SPEC's run-phase; deletion remains the responsibility of a future cleanup SPEC after downstream SPECs 002-008 are merged (inheriting V3R4-HARNESS-001 §2.2 Out of Scope item 1).

### REQ-HRN-NS-007 (Event-Driven — PR #908 closeout via orchestrator escalation)

**When** this SPEC's run-phase executes, the orchestrator **shall** detect PR #908's 2-commit divergence (HEAD = `a41d6d139c8c769bf395a25a055d59c14e180191` ≠ absorbed-commit prefix `452aa638f*`) via `gh pr view 908 --json headRefOid` with **prefix matching** (`[[ "$head" == 452aa638f* ]]`, not string equality), **and shall** invoke `AskUserQuestion` with the 4 disposition options enumerated in EC-001 (rollback-tip-then-close (Recommended) / cherry-pick-to-new-PR-then-close / abandon-new-commits / leave-open). The `manager-git` subagent **shall** execute the user-selected option (subagents do not invoke `AskUserQuestion` per inherited V3R4-HARNESS-001 REQ-HRN-FND-015). The closeout decision (whichever option the user chose) **shall** be recorded in `CHANGELOG.md` under the v3.0.0-rc1 entry and in an auto-memory project entry per session-handoff conventions. For options 1-3, the orchestrator **shall** confirm closeout by invoking `gh pr view 908 --json state` and expecting `state: CLOSED`; for option 4 (leave-open), the orchestrator **shall** record the deferral rationale in the audit artifact without closing the PR.

### REQ-HRN-NS-008 (Unwanted — silent verb extension prevention)

**If** any future SPEC under `.moai/specs/SPEC-V3R4-HARNESS-*/` introduces an additional `/moai harness` verb (a fifth verb beyond `status`, `apply`, `rollback`, `disable`) or renames an existing verb without explicitly amending REQ-HRN-NS-003 via a new governance SPEC, **then** the change **shall** be considered a governance violation and **shall** be rejected by `plan-auditor` during the offending SPEC's plan-phase audit.

### REQ-HRN-NS-009 (Unwanted — namespace path drift prevention)

**If** any future SPEC under `.moai/specs/SPEC-V3R4-HARNESS-*/` introduces a project-local harness state file outside of `.moai/harness/*`, or relocates an existing canonical sub-path (REQ-HRN-NS-004), without explicitly amending REQ-HRN-NS-004 via a new governance SPEC, **then** the change **shall** be considered a governance violation and **shall** be rejected by `plan-auditor` during the offending SPEC's plan-phase audit.

### REQ-HRN-NS-010 (Optional Feature — usage-log retention policy)

**Where** a downstream SPEC introduces a rotation, compaction, or pruning policy for `.moai/harness/usage-log.jsonl`, the policy **shall** preserve at minimum the most recent 7-day rolling window of observations to support the Tier-4 rate-limit calculation per V3R4-HARNESS-001 REQ-HRN-FND-012. Rotation **shall not** delete or summarize `learning-history/frozen-guard-violations.jsonl` or `learning-history/snapshots/` — those remain inviolate.

---

## 6. Acceptance Coverage Map

The acceptance criteria are defined in `acceptance.md` (sibling file). Every REQ from §5 maps to at least one AC. Full Given-When-Then scenarios are in `acceptance.md`.

| AC ID | Covers REQ IDs |
|-------|----------------|
| AC-HRN-NS-001 | REQ-HRN-NS-001 |
| AC-HRN-NS-002 | REQ-HRN-NS-002 |
| AC-HRN-NS-003 | REQ-HRN-NS-003, REQ-HRN-NS-008 |
| AC-HRN-NS-004 | REQ-HRN-NS-004, REQ-HRN-NS-009 |
| AC-HRN-NS-005 | REQ-HRN-NS-005 |
| AC-HRN-NS-006 | REQ-HRN-NS-006 |
| AC-HRN-NS-007 | REQ-HRN-NS-007 |
| AC-HRN-NS-008 | REQ-HRN-NS-010 |

Coverage: 10 REQs ↔ 8 ACs, every REQ appears in at least one AC.

---

## 7. Constraints

[HARD] Language: All SPEC artifact content in English (per `.claude/rules/moai/development/coding-standards.md` § Language Policy). Internal reasoning during drafting may use Korean.

[HARD] FROZEN zone preservation: `.claude/rules/moai/design/constitution.md` §2 and §5 are NOT modified by this SPEC.

[HARD] No modification of SPEC-V3R4-HARNESS-001 or the three superseded V3R3 SPECs within this PR (inherited from V3R4-HARNESS-001 REQ-HRN-FND-013).

[HARD] No code change in run-phase. All deliverables are markdown (SPEC artifacts already produced in plan-phase + CHANGELOG entry + memory entry + PR #908 closeout comment).

[HARD] EARS format mandatory for all REQs. Every REQ uses one of the five EARS patterns listed in §5 preamble.

[HARD] Conventional Commits format for all commits originating from this SPEC. Plan commit prefix: `plan(spec):`. Run commit prefix: `feat(spec):` (even though no code, the SPEC artifact authoring qualifies as feature work for this family). Sync commit prefix: `sync(spec):` per V3R4-SDF-001 lesson #16.

[HARD] No tech-stack implementation assumptions in spec.md. Requirements describe contracts and capabilities. Implementation decisions belong to `plan.md`.

[HARD] No emojis in user-facing output (per `.claude/rules/moai/development/coding-standards.md` § Content Restrictions).

[HARD] No time estimates or duration predictions in any artifact (per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation). Use priority labels (P0/P1/P2/P3 — this SPEC is P1) and phase ordering instead.

[HARD] Plan-in-main mode (CLAUDE.local.md §18.12 BODP signals A=¬ B=¬ C=¬ → base `origin/main`, branch `feat/SPEC-V3R4-HARNESS-NAMESPACE-001-plan`). Worker worktree zero. No tmux session.

[HARD] 12-field canonical frontmatter (per `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT — established by SPEC-V3R4-SPECLINT-DEBT-002 PR #943): id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags. Optional extension fields used: dependencies, related_specs, breaking, related_theme, target_release.

[HARD] `moai spec lint --strict` (no-args, full-repo scan) MUST PASS with `✓ No findings — all SPEC documents are valid` output before plan-auditor delegation. The CLI does NOT accept directory arguments (`moai spec lint --strict <dir>` produces `ParseFailure: is a directory`); pass either no args (recommended for governance verification — corpus-wide PASS is a stronger guarantee for this SPEC) or individual file paths.

---

## 8. Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Downstream SPEC author overlooks this governance SPEC and introduces a fifth `/moai harness` verb without amendment | Medium | REQ-HRN-NS-008 makes silent extension a `plan-auditor` rejection criterion. The governance SPEC is referenced in V3R4-HARNESS-001 plan.md follow-up section and CHANGELOG v3.0.0-rc1 entry, raising discoverability. |
| PR #908 closeout has unmerged commits beyond `452aa638f` (CONFIRMED at plan-phase) | **High (EXPECTED)** | Current branch HEAD = `a41d6d139c8c769bf395a25a055d59c14e180191`. REQ-HRN-NS-007 + EC-001 treat this as the **expected path**, not exception: orchestrator invokes `AskUserQuestion` with 4 disposition options (rollback-tip-then-close (Recommended) / cherry-pick-to-new-PR-then-close / abandon-new-commits / leave-open). Force-close is prohibited regardless of option chosen. |
| `moai spec lint --strict` regression introduced between plan-phase authoring and run-phase verification | Low | Plan-phase verification (this SPEC's Step 3 in workflow) runs `moai spec lint --strict` immediately on the new SPEC directory. Run-phase re-verification is a binary AC (AC-HRN-NS-005). Any drift caught at run-phase blocks merge. |
| Inherited V3R4-HARNESS-001 REQ-HRN-FND-013 prohibition is misread as forbidding ALL V3R3 SPEC changes | Low | This SPEC's §1.3 (Non-Goals) and §4 item 4 explicitly restate the prohibition with attribution to V3R4-HARNESS-001. The follow-up V3R3 status transition remains a `manager-git` responsibility outside this SPEC's scope. |
| `target_release: v3.0.0-rc1` slips and this SPEC becomes a release-blocking dependency | Low | This SPEC is pure governance with zero LOC code change; it can ship independently of feature-bearing SPECs. If v3.0.0-rc1 slips, target_release rolls forward without functional impact. |
| Future cleanup SPEC physically deleting `internal/cli/harness.go` accidentally also removes the governance SPEC's REQ-HRN-NS-006 anchor | Low | The anchor is a documentation-only reference; the cleanup SPEC inherits the deprecation marker contract and replaces it with a `removed:` historical note. No runtime regression. |

---

## 9. Dependencies

| SPEC | Relationship | Notes |
|------|--------------|-------|
| `SPEC-V3R4-HARNESS-001` | Foundation dependency (blocking) | Status `completed`, merged PR #910 commit `bb80ea0f4`. This governance SPEC complements but does NOT amend the foundation. REQ references throughout this SPEC cite foundation REQs verbatim. |
| `SPEC-V3R4-HARNESS-002` | Related (non-blocking) | Exists on disk (plan/spec/research/tasks present). Future merge of -002 MUST adhere to REQ-HRN-NS-001 through REQ-HRN-NS-009. |
| `SPEC-V3R4-HARNESS-003` | Related (non-blocking) | Plan only, OPEN PR #923 (9334 LOC for embedding-cluster classifier). Future merge of -003 MUST adhere to REQ-HRN-NS-001 through REQ-HRN-NS-009; specifically, embedding-cluster artifacts MUST live under `.moai/harness/*` per REQ-HRN-NS-004. |
| `SPEC-V3R3-HARNESS-001` | Related (superseded by V3R4-HARNESS-001) | Status transition to `superseded` is the responsibility of the inherited V3R4-HARNESS-001 follow-up `manager-git` commit, not this SPEC. |
| `SPEC-V3R3-HARNESS-LEARNING-001` | Related (superseded by V3R4-HARNESS-001) | Same as above. |
| `SPEC-V3R3-PROJECT-HARNESS-001` | Related (superseded by V3R4-HARNESS-001) | Same as above. |
| `SPEC-V3R4-SPECLINT-DEBT-002` | Reference (non-blocking) | Established the 12-field canonical frontmatter SSOT this SPEC follows. PR #943 merged. |
| `SPEC-V3R4-SDF-001` | Reference (non-blocking) | Established the status-drift sweep + terminal-state exemption + `sync(spec):` commit prefix discipline this SPEC follows. PR #940 merged. |

---

## 10. Glossary

- **Harness family**: The set of SPECs under `.moai/specs/SPEC-V3R4-HARNESS-*/` consisting of `-001` (foundation, completed), `-002` (TBD), `-003` (plan only, OPEN PR #923), and anticipated `-004` through `-008`.
- **Governance SPEC**: A SPEC whose run-phase produces zero LOC code change and whose deliverables are markdown + lint verification + cross-PR coordination. Precedent: SPEC-V3R4-SPECLINT-DEBT-002, SPEC-V3R4-SDF-001, this SPEC.
- **Namespace convention**: The combination of SPEC ID format (REQ-HRN-NS-001), directory hierarchy (REQ-HRN-NS-004), and verb surface (REQ-HRN-NS-003) that every harness-family SPEC MUST follow.
- **Lifecycle independence**: The guarantee (REQ-HRN-NS-002) that no single harness SPEC's plan PR, run PR, or sync PR may block another family member's lifecycle. Enforced via explicit `dependencies:` frontmatter declarations only.
- **Verb surface**: The exact four-verb list (`status`, `apply`, `rollback`, `disable`) exposed by `/moai harness`. Extension requires a new governance SPEC (REQ-HRN-NS-008).
- **FROZEN zone (inherited)**: Path prefixes the harness MUST NOT auto-modify. Defined in V3R4-HARNESS-001 REQ-HRN-FND-006 and design constitution §2. This SPEC does NOT extend or weaken the FROZEN zone.
- **PR #908 closeout**: The orchestrator-coordinated resolution of PR #908 (`feat/cmd-harness-slash-wrapper`)'s **2-commit divergence** state. Commit 1 (`452aa638f`, slash-wrapper) was absorbed into PR #910 commit `bb80ea0f4` (V3R4-HARNESS-001 run-phase). Commit 2 (`a41d6d139c8c769bf395a25a055d59c14e180191`, audit doc) is unabsorbed. Orchestrator invokes `AskUserQuestion` with 4 options: (1) rollback-tip-then-close (Recommended — `git reset --hard 452aa638f && git push --force-with-lease` then close, preserves SPEC intent), (2) cherry-pick-to-new-PR-then-close (preserves audit doc in separate PR), (3) abandon-new-commits (close with audit doc loss), (4) leave-open (defer closeout to a later SPEC). User-selected option is executed by `manager-git` per REQ-HRN-NS-007.
- **Plan-in-main mode**: Per CLAUDE.local.md §18.12 BODP, plan-phase execution in the main checkout on a feature branch (`feat/SPEC-V3R4-HARNESS-NAMESPACE-001-plan`) without a worktree. Applies when BODP signals are all-negative (no `depends_on` path overlap, no co-located SPEC, no open PR on current branch).

---

## 11. References

- Foundation SPEC (binding contract):
  - `.moai/specs/SPEC-V3R4-HARNESS-001/spec.md` — REQ-HRN-FND-001 through REQ-HRN-FND-018.
  - `.moai/specs/SPEC-V3R4-HARNESS-001/plan.md` — Wave A through Wave E run-phase roadmap.
  - `.moai/specs/SPEC-V3R4-HARNESS-001/follow-up.md` — V3R3 status-transition manager-git instructions.
- Related downstream SPECs:
  - `.moai/specs/SPEC-V3R4-HARNESS-002/` — observer expansion (plan/spec/research/tasks present, status TBD).
  - `.moai/specs/SPEC-V3R4-HARNESS-003/` — embedding-cluster classifier (plan only, OPEN PR #923).
- Superseded V3R3 SPECs (historical):
  - `.moai/specs/SPEC-V3R3-HARNESS-001/spec.md`
  - `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/spec.md`
  - `.moai/specs/SPEC-V3R3-PROJECT-HARNESS-001/spec.md`
- Governance SPEC precedents:
  - `.moai/specs/SPEC-V3R4-SPECLINT-DEBT-002/spec.md` — 12-field canonical frontmatter SSOT.
  - `.moai/specs/SPEC-V3R4-SDF-001/spec.md` — status drift sweep + `sync(spec):` discipline.
- Design constitution (FROZEN — inherited only):
  - `.claude/rules/moai/design/constitution.md` §2 (Frozen vs Evolvable Zones), §5 (Safety Architecture).
- SSOT documents:
  - `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical schema.
- Orchestrator-subagent contract (FROZEN — inherited only):
  - `.claude/rules/moai/core/agent-common-protocol.md` § User Interaction Boundary.
  - `.claude/rules/moai/core/askuser-protocol.md` — Canonical AskUserQuestion protocol.
- Workflow rules:
  - `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline.
  - CLAUDE.local.md §18.12 Branch Origin Decision Protocol (BODP).
- Harness runtime artifacts (DO NOT MODIFY in this SPEC):
  - `.claude/skills/moai/workflows/harness.md` — verb implementation body (~19KB).
  - `.claude/commands/moai/harness.md` — slash command thin wrapper.
  - `internal/cli/harness.go` — deprecation marker (not registered in `internal/cli/root.go`).
- PR references:
  - PR #908 `feat/cmd-harness-slash-wrapper` — OPEN since 2026-05-13 (closeout target per REQ-HRN-NS-007).
  - PR #910 commit `bb80ea0f4` — V3R4-HARNESS-001 run-phase absorption of #908 content.
  - PR #923 — V3R4-HARNESS-003 plan, OPEN (9334 LOC, embedding-cluster classifier).
  - PR #943 — V3R4-SPECLINT-DEBT-002 sync, MERGED (`82880cd49` main HEAD).
