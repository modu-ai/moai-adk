---
id: SPEC-V3R5-CORE-SLIM-001
title: "V3R5 W2 — LR-08 Rule Refinement + Expert Agent Foundation Skill Symmetry"
version: "0.2.1"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v2.20.0"
module: "internal/cli/agent_lint.go + .claude/agents/moai/"
lifecycle: spec-anchored
tags: "v3r5, w2, core-slim, lr-08, lint-refinement, skill-preload, symmetry"
---

# SPEC-V3R5-CORE-SLIM-001 — LR-08 Rule Refinement + Expert Agent Foundation Skill Symmetry

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.2.1 | 2026-05-20 | GOOS Kim | plan-auditor iter1 REVISE (0.84) findings resolved — C-D1 expert-frontend insertion position pinned to preserve-existing-order; D-S1 W2-deferred manifest discrepancy documented with deletion DoD; D-S2 AC-CSLM-007 probe switched to complement-style regex; D-S3 AC-CSLM-008 added for EC-6 sentinel; M-D1 REQ-CSLM-004 re-classified Ubiquitous; M-D3 design↔research cross-link added. |
| 0.2.0 | 2026-05-20 | GOOS Kim | **Scope pivot** from mechanical agent preload addition to (a) LR-08 rule refinement exempting domain-scoped skill prefixes + (b) foundation-quality uniformity addition to 4 expert agents. Original v0.1.0 scope was determined empirically incorrect via EC-4 verification at plan-phase: all matrix domain skills were ALREADY present on the named source agents (research.md §3.2), so the 4-file additive scope was a no-op for AC-CSLM-001. Re-reading `internal/cli/agent_lint.go:892-977` `checkSkillPreloadDrift` revealed the rule enforces strict uniform symmetry with no domain-specificity exemption (research.md §3 Discovery Log). Adding domain skills like `moai-domain-backend` to every expert agent is semantically wrong (bloats frontend/devops contexts). Correct resolution: refine LR-08 to exempt domain-prefix skill names + add `moai-foundation-quality` to the 4 expert agents missing it (expert-backend/frontend/refactoring/devops). v0.2.0 supersedes v0.1.0. |
| 0.1.0 | 2026-05-20 | GOOS Kim | Initial draft — v3.5.0 Mega-Sprint Wave 2 (CORE-SLIM) follow-up to SPEC-V3R5-CONSTITUTION-DUAL-001 W1. Targeted 12 LR-08 warnings via mechanical preload addition (8 doc files). **Superseded by v0.2.0** — plan-phase verification revealed the 8-file mechanical addition would not resolve LR-08 (domain skills already present; LR-08 strict uniform symmetry conflicts with domain specificity). |

## 1. Goal

Eliminate the **12 LR-08 warnings** currently emitted by `./bin/moai agent lint --strict` for the `expert` agent category through two complementary tracks:

- **Track A — Rule refinement (Go code)**: Modify `internal/cli/agent_lint.go` `checkSkillPreloadDrift` to **exempt domain-specific skill prefixes** (`moai-domain-*`, `moai-design-*`, `moai-library-*`, `moai-framework-*`, `moai-platform-*`, `moai-ref-*`) from intra-category symmetry checking. Foundation skills (`moai-foundation-*`) and workflow skills (`moai-workflow-*`) remain subject to uniform symmetry enforcement — they are universally relevant.
- **Track B — Foundation symmetry (agent metadata)**: Add `moai-foundation-quality` to 4 expert agents currently missing it (`expert-backend`, `expert-frontend`, `expert-refactoring`, `expert-devops`) + their 4 template mirrors, completing the foundation-skill uniformity for the expert category. After Track A, `moai-foundation-quality` would otherwise remain the sole legitimate detection of the rule (only 2 of 6 experts preload it).

Combined effect: 10 domain-skill LR-08 warnings dissolve via Track A; 2 foundation-quality LR-08 warnings dissolve via Track B. Total 12 → 0.

This SPEC is the direct follow-up to SPEC-V3R5-CONSTITUTION-DUAL-001 (W1 of the v3.5.0 Mega-Sprint), which intentionally deferred these 12 findings as the documented `W2-deferred` baseline to break the chicken-and-egg admin-override deadlock (see project memory `project_v3r5_w1_constitution_complete`). With W1 now merged at main HEAD `175bad283`, this SPEC clears the residual baseline so that subsequent Mega-Sprint Waves (W3 HARNESS-AUTONOMY-001, W4 PROJECT-MEGA-001) can assert clean `NEW_COUNT=0` deltas without inheriting noise.

Outcome target:
- Pre-merge baseline (`./bin/moai agent lint --strict` on `feat/SPEC-V3R5-CORE-SLIM-001`): 12 LR-08 warnings in the `expert` category
- Post-merge state (`./bin/moai agent lint --strict` on `main`): 0 LR-08 warnings in the `expert` category
- Side-effect: total `moai agent lint --strict` finding count drops by 12 with no NEW findings introduced (delta-only D6 semantics per SPEC-V3R5-LINT-CLEAN-001 precedent)

## 2. Scope

### 2.1 In Scope

| Track | Path | File count | Edit type |
|-------|------|-----------|-----------|
| A — Rule fix | `internal/cli/agent_lint.go` | 1 | Add `domainExemptPrefixes` constant + `isDomainExemptSkill` helper + skip logic in `checkSkillPreloadDrift` |
| A — Tests | `internal/cli/agent_lint_test.go` | 1 | Add 4 new test cases verifying exemption + enforcement bidirectionally |
| B — Source agents | `.claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md` | 4 | Add `moai-foundation-quality` to YAML frontmatter `skills:` array |
| B — Template mirrors | `internal/template/templates/.claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md` | 4 | Identical mirror edits |
| B — Catalog refresh | `internal/template/catalog.yaml` | 1 | Generated by `gen-catalog-hashes.go --all` |
| **Total** | | **11** | 1 Go src + 1 Go test + 4 source agents + 4 template mirrors + 1 catalog |

#### Track A scope detail

Modify the `checkSkillPreloadDrift` function (currently at `internal/cli/agent_lint.go:892-977`) by adding:

1. A file-scoped constant `domainExemptPrefixes` enumerating the 6 domain-scoped skill prefixes
2. A helper `isDomainExemptSkill(skill string) bool` using `strings.HasPrefix` matching
3. A skip clause inside the drift inner loop: `if isDomainExemptSkill(skill) { continue }` placed before the drift count check at line 961

The skip logic exempts skills matching any of the 6 domain prefixes from the symmetry check. Foundation skills (`moai-foundation-*`) and workflow skills (`moai-workflow-*`) continue to be enforced.

#### Track B scope detail

Add `moai-foundation-quality` to the YAML frontmatter `skills:` array of these 4 expert agents (in alphabetical order within the existing arrays). All 4 expert agents currently have only `moai-foundation-core` + `moai-workflow-testing` (devops/refactoring) or only `moai-foundation-core` + `moai-domain-*` + `moai-workflow-testing` (backend/frontend) and lack `moai-foundation-quality`.

| Agent | Current `skills:` | After Track B |
|-------|------------------|---------------|
| `expert-backend` | core, domain-backend, domain-database, workflow-testing | core, domain-backend, domain-database, **foundation-quality**, workflow-testing |
| `expert-frontend` | core, domain-frontend, design-system, workflow-testing | core, domain-frontend, design-system, **foundation-quality**, workflow-testing |
| `expert-refactoring` | core, workflow-testing | core, **foundation-quality**, workflow-testing |
| `expert-devops` | core, workflow-testing | core, **foundation-quality**, workflow-testing |

The remaining 2 expert agents (`expert-performance`, `expert-security`) already preload `moai-foundation-quality` and are NOT edited.

> See `research.md` §2 for the verbatim `moai agent lint --strict` output, §3 for the LR-08 implementation analysis, and §4 for the EC-4 Discovery Log that motivated the v0.1.0 → v0.2.0 scope pivot.

### 2.2 Out of Scope (Exclusions)

This SPEC is deliberately narrow. The following are explicitly NOT in scope:

- **LR-08 semantics for other categories** (manager, builder, evaluator) — the same `checkSkillPreloadDrift` function applies to all categories. The exemption logic naturally extends across categories (no category-specific code). No matrix changes for non-expert categories.
- **Skill catalog full audit** — orphan skill detection, deprecated entries, unused-skill removal. Deferred to a separate SPEC.
- **Removal of `moai-foundation-quality` from `expert-performance`, `expert-security`** (alternative path to symmetry) — explicitly rejected: foundation-quality is universally needed by all expert agents, including those that already have it.
- **LR-01 through LR-07, LR-09, LR-10 (and other LR-* rules)** — this SPEC modifies LR-08 only. Other rules' implementations remain untouched.
- **Skill body content changes** — no edits to any `.claude/skills/<name>/SKILL.md` body.
- **New skill creation** — `moai-foundation-quality` already exists (verified in `research.md` §1).
- **SPEC frontmatter under `.moai/specs/`** — `./bin/moai spec lint --strict` must remain at 0.
- **Domain prefix list extensions** — the 6 prefixes in `domainExemptPrefixes` are derived from the current skill catalog taxonomy (research.md §3.4). Future prefix additions (e.g., a hypothetical `moai-mobile-*`) require a one-line constant extension; that is operational maintenance, not part of this SPEC's scope.
- **FROZEN zones** — no edits to zone-registry-enumerated paths.
- **Behavior change in agent runtime** — adding `moai-foundation-quality` to 4 expert agents is metadata only; their behavioral semantics during run-phase remain unchanged beyond the LR-08 detection signal.

## 3. Requirements (EARS)

### Ubiquitous Requirements

- **REQ-CSLM-001**: The LR-08 rule SHALL exempt domain-specific skill prefixes (`moai-domain-`, `moai-design-`, `moai-library-`, `moai-framework-`, `moai-platform-`, `moai-ref-`) from intra-category symmetry checking. Skills whose names begin with any of these 6 prefixes SHALL NOT trigger drift warnings, regardless of how many agents in the category preload them.

- **REQ-CSLM-002**: Foundation skills (whose names begin with `moai-foundation-`) and workflow skills (whose names begin with `moai-workflow-`) SHALL remain subject to uniform symmetry enforcement under LR-08. These skills are universally relevant across agents in a category and continue to trigger drift warnings when not preloaded by all agents in the category.

- **REQ-CSLM-005**: The lint rule fix SHALL include unit tests verifying (a) the exemption path (domain-prefix skills do NOT trigger LR-08 even when not uniformly preloaded) and (b) the enforcement path (foundation/workflow-prefix skills DO trigger LR-08 when not uniformly preloaded). Tests SHALL cover edge cases including empty skill arrays, single-agent categories, and foundation-domain-lookalike skill names.

- **REQ-CSLM-004** (Ubiquitous, re-classified per plan-auditor v0.2.1 M-D1): The 4 source agent files (`.claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md`) and their corresponding 4 template mirror files (`internal/template/templates/.claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md`) SHALL contain byte-identical YAML frontmatter (including the `skills:` array). The `gen-catalog-hashes.go` tool SHALL be invoked on Track B Step B.4 to refresh `internal/template/catalog.yaml` so the catalog-hash audit test (`go test ./internal/template/...`) passes.

### Event-Driven Requirements

- **REQ-CSLM-003**: WHEN `./bin/moai agent lint --strict` is invoked on `main` HEAD after this SPEC's run-phase PR merges, THEN the LR-08 rule SHALL emit zero warnings for the `expert` category, reducing the previously-deferred `W2-deferred` baseline of 12 findings to 0.

## 4. Acceptance Criteria (EARS Hierarchical)

Concrete shell-verifiable AC tree. Full Given-When-Then scenarios live in `acceptance.md`.

- **AC-CSLM-001** (domain-prefix exemption — composite leaf set):
  - Given: Track A LR-08 rule fix is applied to `internal/cli/agent_lint.go`
  - When: `./bin/moai agent lint --strict 2>&1 | grep "LR-08"` runs against `.claude/agents/moai/`
  - Then: every leaf below returns zero matching lines
    - **AC-CSLM-001.a**: no `LR-08` warning mentions `moai-domain-backend`
    - **AC-CSLM-001.b**: no `LR-08` warning mentions `moai-domain-database`
    - **AC-CSLM-001.c**: no `LR-08` warning mentions `moai-domain-frontend`
    - **AC-CSLM-001.d**: no `LR-08` warning mentions `moai-design-system`

- **AC-CSLM-002** (foundation-quality symmetry — binary metric):
  - Given: Track B preload addition is applied to 4 expert agents
  - When: `./bin/moai agent lint --strict 2>&1 | grep "LR-08" | grep "moai-foundation-quality"` runs
  - Then: the command emits zero matching lines (LR-08 warning count for `moai-foundation-quality` across the expert category = 0)

- **AC-CSLM-003** (template mirror parity — composite leaf set):
  - Given: Track B source edits committed to `feat/SPEC-V3R5-CORE-SLIM-001`
  - When: `diff .claude/agents/moai/expert-<name>.md internal/template/templates/.claude/agents/moai/expert-<name>.md` is run for each of `backend, frontend, refactoring, devops`
  - Then: every diff returns empty output and exit code 0 for all 4 pairs
    - **AC-CSLM-003.a**: `expert-backend` source/mirror diff empty
    - **AC-CSLM-003.b**: `expert-frontend` source/mirror diff empty
    - **AC-CSLM-003.c**: `expert-refactoring` source/mirror diff empty
    - **AC-CSLM-003.d**: `expert-devops` source/mirror diff empty

- **AC-CSLM-004** (catalog hash refresh):
  - Given: source + mirror edits committed
  - When: `go run internal/template/scripts/gen-catalog-hashes.go --all` runs and writes `internal/template/catalog.yaml`
  - Then: `go test ./internal/template/...` passes (catalog hash audit test green), and `internal/template/catalog.yaml` diff shows updates for exactly 4 entries (`expert-backend`, `expert-frontend`, `expert-refactoring`, `expert-devops`)

- **AC-CSLM-005** (rule fix unit tests — composite leaf set):
  - Given: Track A code change applied to `internal/cli/agent_lint.go` + 4 new test cases added to `internal/cli/agent_lint_test.go`
  - When: `go test ./internal/cli/... -run "TestSkillPreloadDriftExemption"` runs
  - Then: every leaf below returns PASS
    - **AC-CSLM-005.a**: `TestSkillPreloadDriftExemption_DomainSkills` PASS (negative case — domain skills NOT flagged even when not uniform across category)
    - **AC-CSLM-005.b**: `TestSkillPreloadDriftExemption_FoundationSkills` PASS (positive case — foundation skills STILL flagged when not uniform)
    - **AC-CSLM-005.c**: `TestSkillPreloadDriftExemption_WorkflowSkills` PASS (positive case — workflow skills STILL flagged when not uniform)
    - **AC-CSLM-005.d**: `TestSkillPreloadDriftExemption_EdgeCases` PASS (empty skill array, single-agent category, foundation-domain-lookalike names)

- **AC-CSLM-006** (full LR-08 dissolution — binary success metric):
  - Given: Tracks A and B both committed on `main` HEAD post-merge
  - When: `./bin/moai agent lint --strict 2>&1 | grep "LR-08" | wc -l` runs across all categories (manager, expert, builder, evaluator)
  - Then: total LR-08 warning count = 0. **This is the primary binary success metric for the SPEC.**

- **AC-CSLM-007** (non-regression — orthogonal lint surfaces, complement-style):
  - Given: Tracks A and B both committed
  - When: `./bin/moai agent lint --strict 2>&1 | grep -E "^! \[LR-" | grep -v "^! \[LR-08\]" | wc -l` runs
  - Then: output is equal to the pre-merge complement baseline (auto-tracks ALL LR rules except LR-08 without manual enumeration; baseline pre-recorded in `acceptance.md` § Verification Matrix as `N`; post-merge MUST equal N to prove no orthogonal regression was introduced as a side effect of Track A logic refactor)

- **AC-CSLM-008** (EC-6 sentinel contract documented — binary metric):
  - Given: spec.md §5 EC-6 documents the `MIRROR_MISSING_BLOCKER` sentinel contract
  - When: `grep -c "MIRROR_MISSING_BLOCKER" .moai/specs/SPEC-V3R5-CORE-SLIM-001/spec.md` runs
  - Then: output ≥ 1 (verifies the sentinel contract is durably documented; actual run-phase enforcement is the implementer's responsibility per Scenario 6 in `acceptance.md`)

## 5. Edge Cases / Error Conditions

- **EC-1** (new domain skill prefix added in future): A future SPEC introduces a new domain-scoped skill prefix (e.g., `moai-mobile-`, `moai-cloud-`) that should also be exempted from LR-08 symmetry. Mitigation: `domainExemptPrefixes` is a file-scoped constant list; extending it requires a one-line addition. No schema migration. This edge case is operational maintenance, NOT a defect.

- **EC-2** (foundation skill named with domain-like pattern): A hypothetical skill named `moai-foundation-frontend-test` exists. The rule's `strings.HasPrefix(skill, "moai-foundation-")` check matches first (it is checked for enforcement, not exemption). Explicit prefix matching with literal strings prevents false exemption. The 6 exemption prefixes start with `moai-domain-`, `moai-design-`, `moai-library-`, `moai-framework-`, `moai-platform-`, `moai-ref-` — no overlap with `moai-foundation-` or `moai-workflow-`. Mitigation: the prefix list uses literal matching; no substring or fuzzy logic.

- **EC-3** (empty skills array in agent): An agent file with `skills: []` (or missing the `skills:` field entirely) is parsed by `parseSkillsList`. The existing `len(agents) < 2` early return at line 946 already handles single-agent categories. For empty arrays in multi-agent categories, the inner loop has no entries to iterate. Mitigation: no rule change required; existing behavior preserved.

- **EC-4** (single agent in category): If a category contains only 1 agent (none currently, but possible in future), the existing `len(agents) < 2` early return at line 946 skips the symmetry check. Mitigation: no change required; existing behavior preserved.

- **EC-5** (Track A regression on non-expert categories): Track A's logic refactor must not introduce false negatives or false positives in other categories (manager, builder, evaluator). Mitigation: AC-CSLM-007 verifies non-regression across all category lint counts. Unit tests AC-CSLM-005 use multi-category fixtures to confirm.

- **EC-6** (Track B mirror missing at run-time): The 4 template mirrors at `internal/template/templates/.claude/agents/moai/expert-{backend,frontend,refactoring,devops}.md` are all present (verified 2026-05-20 in `research.md` §1.2). If a mirror becomes absent at run-time (e.g., concurrent destructive operation), the run-phase MUST halt with sentinel `MIRROR_MISSING_BLOCKER` and emit a blocker report — no source edit is applied without mirror parity.

## 6. Out of Scope (Exclusions)

Explicit non-goals (repeated for visibility per `manager-spec` rule "every spec.md MUST include `## Exclusions (What NOT to Build)`"):

- LR-08 semantics for non-`expert` agent categories beyond the implicit exemption-list extension
- Skill catalog full audit (orphan / deprecated)
- Removal of `moai-foundation-quality` from `expert-performance`, `expert-security`
- All non-LR-08 lint rules (LR-01..LR-07, LR-09 and all subsequent LR rules — no implementation changes; this SPEC modifies LR-08 only)
- Skill body content edits
- New skill creation
- SPEC frontmatter under `.moai/specs/` edits
- Behavior change in agent runtime
- Domain prefix list extensions beyond the canonical 6
- FROZEN zone files

## 7. Dependencies

- **SPEC-V3R5-CONSTITUTION-DUAL-001** (W1, COMPLETED at main HEAD `175bad283`) — this SPEC is the documented follow-up that resolves W1's `W2-deferred` 12-finding baseline. Memory reference: `project_v3r5_w1_constitution_complete`.
- **Existing skill `moai-foundation-quality`** (verified present 2026-05-20):
  - `.claude/skills/moai-foundation-quality/SKILL.md`
- **Domain skill catalog taxonomy** — the 6 exemption prefixes are derived from the canonical skill prefix list in `.claude/rules/moai/development/skill-authoring.md` and verified against the actual `.claude/skills/` tree (research.md §3.4 and §3.5).
- **Template synchronization tooling** (`make build`, `gen-catalog-hashes.go`) — pre-existing; no new tooling required.

## 8. Risk Assessment

- **Track A (Go rule fix)**: **LOW**. Adds 1 file-scoped constant + 1 helper function + 1 skip clause. Bidirectional test coverage (exemption + enforcement) prevents regression. Existing `internal/cli/agent_lint_test.go` provides scaffolding patterns for new tests.
- **Track B (agent metadata)**: **LOW**. Pure YAML frontmatter array additions, no body content edits, no schema changes, no breaking API changes. Mirrors the SPEC-V3R5-LINT-CLEAN-001 metadata-edit pattern (proven safe).
- **Reversibility**: Trivial for both tracks — `git restore internal/cli/agent_lint.go internal/cli/agent_lint_test.go .claude/agents/moai/expert-*.md internal/template/templates/.claude/agents/moai/expert-*.md internal/template/catalog.yaml`.
- **Blast radius**: 11 files maximum (1 Go src + 1 Go test + 4 source agents + 4 template mirrors + 1 catalog).

## 9. References

- `internal/cli/agent_lint.go:892-977` — `checkSkillPreloadDrift` current implementation (the target of Track A modification)
- `internal/cli/agent_lint_test.go` — existing test scaffolding for Track A new tests
- `.claude/rules/moai/development/skill-authoring.md` — skill prefix taxonomy source (foundation/workflow/domain/design/library/framework/platform/ref)
- `.claude/rules/moai/development/agent-authoring.md` — agent frontmatter rules
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — SPEC 12-field canonical schema (applied to this `spec.md` frontmatter)
- `.moai/specs/SPEC-V3R5-CONSTITUTION-DUAL-001/spec.md` — W1 source documenting the deferral
- `.moai/specs/SPEC-V3R5-LINT-CLEAN-001/` — sibling lint-cleanup SPEC providing the delta-only D6 NEW=0 enforcement pattern and the W2-deferred manifest at `.moai/state/lint-w2-deferred.json`
- Memory `project_v3r5_w1_constitution_complete` — chicken-and-egg deferral rationale
- `CLAUDE.local.md §2` (Template-First Rule) — `internal/template/templates/` is the canonical source for agent definitions; live `.claude/` is the sync target
