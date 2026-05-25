---
id: SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001
title: "Local Agent Namespace Consolidation — Dev-Only Skill Migration + Template Generic Refactor + CLAUDE.local.md Pattern Externalization"
version: "0.1.1"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.7.0"
module: ".claude/agents/local + .claude/skills/moai/workflows + internal/template/templates + .moai/docs"
lifecycle: spec-anchored
tags: "local-namespace, dev-only, agent-migration, template-refactor, claude-local-externalization, sprint-10-lane-b, thin-command-pattern"
tier: M
depends_on: []
related_specs: []
---

# SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001

## A. Goal

Consolidate three structurally-related cleanup scopes into a single Tier M SPEC that establishes a canonical local-agent namespace (`.claude/agents/local/`), migrates two dev-only skills (`97-release-update`, `98-github`) into local agents under that namespace, refactors approximately 13 template files to remove leaked `CLAUDE.local.md` numbered-section cross-references, and externalizes generic patterns currently embedded in `CLAUDE.local.md` into a reusable `.moai/docs/generic-patterns-guide.md`.

The three scopes share a single failure mode — local-only doctrine bleeding into the template-distributed surface — and a single architectural principle — explicit user-vs-maintainer artifact separation enforced by the `moai update` PRESERVE list (CLAUDE.local.md §24.4 contract). Consolidating them into one SPEC produces one coherent commit cohort, one CHANGELOG entry, and one verification batch instead of three drifted micro-SPECs that would each re-derive the same namespace boundaries.

## B. Scope

### B.1 W3-arch — Local Agent Migration

Migrate two dev-only workflow skills currently located at `.claude/skills/moai/workflows/release-update.md` and `.claude/skills/moai/workflows/github.md` into the new `.claude/agents/local/` namespace. The migration emits two agent body files:

- `.claude/agents/local/release-update-specialist.md` — receives the 9-phase (Phase 0 through Phase 8) CC upstream tracker workflow body from `release-update.md`
- `.claude/agents/local/github-specialist.md` — receives the GitHub issue/PR Agent Teams workflow body from `github.md`

The two thin command wrappers `.claude/commands/97-release-update.md` and `.claude/commands/98-github.md` are NOT removed; they remain as 9-line thin wrappers (HARD: Thin Command Pattern per `coding-standards.md` §Thin Command Pattern). Their routing target changes from `Skill("moai") with arguments: release-update $ARGUMENTS` (skill invocation) to `Use the release-update-specialist subagent to ...` (agent delegation). The orchestrator-side delegation pattern replaces the skill-routing pattern while preserving the user-facing slash command surface.

Approximate scope: 2 new agent body files (~1184 LOC migrated from the two skill bodies) + 2 thin command wrappers updated (~9 LOC each) + 2 dev-only skill files removed.

### B.2 W4 — Template Generic Refactor

Approximately 13 files under `internal/template/templates/` contain literal `CLAUDE.local.md` or `CLAUDE.local` references with numbered-section cross-references (e.g., `CLAUDE.local.md §24`, `CLAUDE.local.md Section 22`). These references are leaks: `CLAUDE.local.md` is the local maintainer file (CLAUDE.local.md §2 Local-Only Files list) and does NOT exist in user projects after `moai init`. Each leak represents either (a) a generic rule that escaped to template without being rewritten, or (b) a rule that belongs in a canonical `.claude/rules/moai/` file but was inlined to template as a cross-reference.

Identified leak distribution (verified via grep at plan time, exact line numbers captured in plan.md M4):

- `.claude/rules/moai/core/agent-common-protocol.md` — 1 reference (line 339, race mitigation cross-ref)
- `.claude/rules/moai/development/agent-authoring.md` — 1 reference (line 34, namespace + moai update contract)
- `.claude/rules/moai/development/branch-origin-protocol.md` — 1 reference (line 73, dev-project specific notes)
- `.claude/rules/moai/development/skill-authoring.md` — 3 references (lines 264, 282, 305)
- `.claude/rules/moai/workflow/moai-memory.md` — 1 reference (line 17, file inventory)
- `.claude/output-styles/moai/moai.md` — 3 references (lines 426, 458, 707)
- `.claude/skills/moai/workflows/loop.md` — 1 reference (line 215, language neutrality §22 cross-ref)
- `.claude/skills/moai/workflows/project/doc-generation.md` — 1 reference (line 139, language neutrality §22 cross-ref)
- `.claude/skills/moai-workflow-loop/SKILL.md` — 1 reference (line 188, language neutrality §22 cross-ref)
- `.claude/skills/moai-workflow-loop/references/reference.md` — 1 reference (line 789, language neutrality §22 cross-ref)
- `.claude/skills/moai-workflow-loop/references/examples.md` — 1 reference (line 510, language neutrality §22 cross-ref)
- `.claude/skills/moai-meta-harness/SKILL.md` — 1 reference (line 180, harness namespace §24 cross-ref)
- `.moai/config/sections/lsp.yaml.tmpl` — 1 reference (line 4, language neutrality §22 cross-ref)

Total leak count: 17 references across 13 template files. The refactor MUST result in zero `CLAUDE.local.md` matches under `internal/template/templates/` after M4 completion. Replacement strategy is per-leak: (a) rewrite to point at the canonical rule file (e.g., `.claude/rules/moai/development/coding-standards.md` §16-language neutrality) when one exists, (b) inline the generic content directly when no canonical exists, or (c) point at the new `.moai/docs/generic-patterns-guide.md` (W5 deliverable) when the content is doctrine-shaped rather than rule-shaped.

### B.3 W5 — CLAUDE.local.md Generic Pattern Externalization

Author a new document at `.moai/docs/generic-patterns-guide.md` that externalizes generic operational patterns currently embedded in maintainer-only `CLAUDE.local.md`. The new guide is template-distributed (lives under `internal/template/templates/.moai/docs/generic-patterns-guide.md`), allowing user projects to benefit from the patterns without inheriting maintainer-specific local doctrine.

Externalized pattern surface (sourced from CLAUDE.local.md, generalized for user audience):

- Multi-session race mitigation procedure (extracted from §23.8 Multi-Session Race Mitigation, generalized — removes references to `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/` user-specific paths)
- Hook setup procedure for new machines (extracted from §23.1 pre-push hook manual setup, generalized — removes 1-person OSS Hybrid Trunk policy specifics; presents the pattern as "if you adopt Hybrid Trunk")
- Settings intent doctrine (extracted from §22 Dev Settings Intent, generalized — explains the four settings keys [defaultMode, enableAllProjectMcpServers, teammateMode, env.PATH] as patterns a user MAY customize, not a maintainer-specific policy)
- Late-branch Phase D 5-step recovery procedure (extracted from §23.6, generalized — git workflow recovery sequence)

W5 does NOT modify CLAUDE.local.md itself (out of scope per task brief). Future maintenance: when CLAUDE.local.md §22/§23 evolves with generic patterns, the maintainer manually mirrors the relevant change into the externalized guide (no automated sync — same as the README and CHANGELOG manual-mirroring patterns).

## C. Requirements (GEARS Notation)

### C.1 Ubiquitous Requirements

**REQ-LNC-001**: The `.claude/agents/local/` namespace shall house all maintainer-only domain specialist agents that are protected from `moai update` overwrite per the namespace separation contract.

**REQ-LNC-002**: The thin command wrappers `.claude/commands/97-release-update.md` and `.claude/commands/98-github.md` shall preserve their Thin Command Pattern compliance per REQ-LNC-013 (body content ≤ 20 LOC), and their YAML frontmatter (delimited by `---` markers, containing 8 fields: `description`, `argument-hint`, `allowed-tools`, `model`, `effort`, `runtime`, `color`, `category`) shall remain structurally unchanged across the migration.

**REQ-LNC-003**: The template surface under `internal/template/templates/` shall contain zero references to `CLAUDE.local.md` after migration completion.

**REQ-LNC-004**: The externalized generic-patterns guide shall reside at `.moai/docs/generic-patterns-guide.md` and shall be mirrored at `internal/template/templates/.moai/docs/generic-patterns-guide.md` per the Template-First Rule (CLAUDE.local.md §2).

### C.2 Event-Driven Requirements

**REQ-LNC-005**: When a maintainer invokes `/97-release-update` from the Claude Code chat input, the thin command wrapper shall delegate to the `release-update-specialist` subagent instead of invoking `Skill("moai") with arguments: release-update`.

**REQ-LNC-006**: When a maintainer invokes `/98-github` from the Claude Code chat input, the thin command wrapper shall delegate to the `github-specialist` subagent instead of invoking `Skill("moai/workflows/github")`.

**REQ-LNC-007**: When the lint engine executes `grep -rln "CLAUDE.local.md\|CLAUDE\.local" internal/template/templates/`, the command shall return exit code 1 (no matches) on a clean working tree after M4 completion.

### C.3 State-Driven Requirements

**REQ-LNC-008**: While the `release-update-specialist` agent is active, the agent shall execute the canonical 9-phase (Phase 0 through Phase 8) CC upstream tracker workflow with structural fidelity preserved (Phase 0-8 headers + Anti-Patterns section) as migrated from the predecessor skill body.

**REQ-LNC-009**: While `moai update` is executing against a user project that contains a `.claude/agents/local/` directory, the update mechanism shall preserve the directory and all files within it without modification, deletion, or sync overwrite.

### C.4 Where-Capability Requirements

**REQ-LNC-010**: Where the `.moai/docs/generic-patterns-guide.md` document is present in a user project, the document shall describe the four externalized pattern families (multi-session race mitigation, hook setup, settings intent doctrine, late-branch Phase D recovery) in user-audience-neutral prose.

**REQ-LNC-011**: The namespace separation contract shall be documented at the two in-scope SSOT locations: `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention table (adding a `.claude/agents/local/` row alongside the existing `.claude/agents/harness/` row) and `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy table (marking the removed `97-release-update` and `98-github` skill slots as deprecated with migration target `.claude/agents/local/`).

### C.5 Unwanted Behavior

**REQ-LNC-012**: The migration shall not introduce any `.claude/agents/local/*` file into `internal/template/templates/`.

**REQ-LNC-013**: The thin command wrappers `.claude/commands/97-release-update.md` and `.claude/commands/98-github.md` shall not exceed 20 lines of body content (excluding YAML frontmatter), preserving the Thin Command Pattern body-LOC bound.

### C.6 Where-Capability Requirements (continued)

**REQ-LNC-014**: Where the maintainer references `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy for namespace classification, the table shall list the removed `97-release-update` and `98-github` slots as deprecated entries with `migration target: .claude/agents/local/<specialist-name>.md` annotation.

## D. Constraints

### D.1 HARD Constraints

- [HARD] **Thin Command Pattern preservation** (`coding-standards.md` § Thin Command Pattern, line 56-77): Both `97-release-update.md` and `98-github.md` MUST remain thin routing wrappers with YAML frontmatter (description, argument-hint, allowed-tools: Skill or single Agent invocation) + single-body-line. The migration changes the routing target (skill → agent) but preserves the wrapper shape. Violation reverts to the workflow-body-inline anti-pattern that the rule was authored to prevent.
- [HARD] **Template-First Rule preservation** (CLAUDE.local.md §2): The new `.moai/docs/generic-patterns-guide.md` MUST be authored under `internal/template/templates/.moai/docs/generic-patterns-guide.md` FIRST, then `make build` regenerates embedded files, then the document appears in local `.moai/docs/` via deployment. Authoring local-first is prohibited.
- [HARD] **Namespace contract update** (two in-scope SSOT locations): The namespace documentation MUST be expanded to include `.claude/agents/local/` as a user-owned namespace alongside `.claude/agents/harness/` at exactly two in-scope SSOT locations: `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention table (add `.claude/agents/local/` row), and `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy table (mark `97-release-update` and `98-github` skill slot removal with migration target annotation). `CLAUDE.local.md` §24.2 (Storage Roots) and §24.4 (`moai update` Contract) update is **explicitly out of scope** for this SPEC and is scheduled as a follow-up maintenance touch in §E Out of Scope.
- [HARD] **Dev-only isolation contract preservation** (`.moai/docs/dev-only-commands-isolation.md`): The verification checklist MUST be updated to reflect the new agent locations. New checklist entries: `find internal/template/templates -path "*/agents/local/*"` returns empty, `find internal/template/templates -name "release-update-specialist.md"` returns empty, `find internal/template/templates -name "github-specialist.md"` returns empty. Existing skill-targeting checklist entries (find `release-update.md` -path `*/workflows/*`, find `github.md` -path `*/workflows/*`) are retained as verification that the predecessor skill bodies do not regress into the template.
- [HARD] **GEARS notation discipline** (`skill-authoring.md`, `moai-workflow-spec`): All 13 REQ-LNC-XXX statements MUST use GEARS patterns. Zero IF/THEN statements. Zero passive-voice "MUST be" without an explicit subject. Subjects MAY be non-"system" nouns (namespace, document, command wrapper, lint engine, etc.) per generalized GEARS subject substitution.

### D.2 SHOULD Constraints

- [SHOULD] Each migrated agent body preserves the predecessor skill body's section structure (Purpose & Scope / Activation / Phase Sequence / Agent Delegation Map / Output Artifacts / Verification Gate / Anti-Patterns / References) to ease historical traceability via git blame.
- [SHOULD] The W4 template refactor preserves the surrounding doctrinal intent of each leak. Example: when `agent-authoring.md` line 34 reads "see `CLAUDE.local.md` §24.2", the replacement reads "see `agent-authoring.md` § Agent Directory Convention" (which already contains the namespace separation policy verbatim per the source file inspection) — not a deletion that would orphan the cross-reference.

## E. Out of Scope

- **CLAUDE.local.md §24.2 (Storage Roots) and §24.4 (`moai update` Contract) PRESERVE-list entry for `.claude/agents/local/`** — explicitly deferred to a follow-up maintenance commit or `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001` if a code-level enforcement track is opened. Rationale: `CLAUDE.local.md` is local-only documentation, not user-distributed, and the in-scope SSOT updates (`agent-authoring.md` + `skill-authoring.md`) provide the user-facing contract.
- **Template mirror of `.moai/docs/dev-only-commands-isolation.md`** — intentionally absent. The file is local-maintainer-only per its own §21 isolation policy (`.moai/docs/dev-only-commands-isolation.md`). M1 updates the local copy only; no template mirror is created or expected. AC-LNC-008 verification command grep targets the local path only.
- **CLAUDE.local.md body modification**: W5 EXTRACTS generic patterns into a new template-distributed guide. It does NOT modify CLAUDE.local.md itself. The maintainer file remains the source of detailed local doctrine; the externalized guide is a generalized derivative.
- **Predecessor skill body content change**: The migration copies the 8-phase workflow body verbatim from the dev-only skills into the new agent bodies. Editorial changes to phase sequencing, agent delegation patterns, or anti-pattern catalogs are out of scope. M2 is mechanical translation; behavioral revisions belong to a follow-up SPEC if needed.
- **`moai update` Go implementation change** (`internal/cli/update.go`, `internal/cli/update_archive.go`): The PRESERVE-list contract update is documentation-only (in agent-authoring.md, skill-authoring.md). Verifying that the Go code actually honors the `.claude/agents/local/` PRESERVE entry is deferred to SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 (per CLAUDE.local.md §24.4 forward-reference). If the Go code already covers `.claude/agents/harness/` via a generic pattern, `.claude/agents/local/` may inherit protection at runtime; verification is the follow-up SPEC's scope.
- **Other dev-only artifact migration**: The 99-release.md command + `workflows/release.md` skill (production release workflow per `.moai/docs/dev-only-commands-isolation.md`) is NOT migrated in this SPEC. Migration is deferred to a follow-up if the namespace pattern proves stable. The dev-only-commands-isolation.md update in M3 documents the rationale for the 97/98 vs 99 differential treatment.
- **CHANGELOG entry authoring**: manager-docs writes the CHANGELOG entry during sync-phase. Plan-phase does not pre-author the entry.
- **PR creation**: This is a Tier M trunk-direct workflow per Hybrid Trunk policy (CLAUDE.local.md §23.7); no feat branch + PR opt-in is requested.

## F. Risks

### F.1 Risk 1 — Thin Command Pattern regression during migration

**Probability**: Low. **Impact**: High (regression triggers `commands_audit_test.go` failure on every `go test ./...`).

**Mitigation**: M2 explicitly preserves the YAML frontmatter shape. The body changes from `Use Skill("moai") with arguments: release-update $ARGUMENTS` to `Use the release-update-specialist subagent with arguments: $ARGUMENTS` — same single-line shape, different routing target. Verification: AC-LNC-002 runs the commands_audit_test directly.

### F.2 Risk 2 — Template leak count discrepancy

**Probability**: Medium. **Impact**: Medium (under-count means M4 declares clean prematurely; over-count means unnecessary refactor work).

**Mitigation**: The plan-time grep yielded 17 references in 13 files; verification re-runs the same grep at run-phase start and at sync-phase completion. The grep is path-bounded (`internal/template/templates/`) and pattern-bounded (`CLAUDE.local.md\|CLAUDE\.local`) — false positives are unlikely. If a 14th file or 18th reference emerges, M4 scope expands without requiring SPEC re-plan.

### F.3 Risk 3 — Externalized guide content quality drift

**Probability**: Medium. **Impact**: Low (user-facing guide quality affects perception but not correctness).

**Mitigation**: W5 sources content from CLAUDE.local.md verbatim then performs targeted generalizations (user-specific paths → placeholders, 1-person OSS policy → "if you adopt this policy", maintainer-machine assumptions → user-neutral framing). The guide is reviewed by plan-auditor as part of the run-phase manager-quality validation hand-off implicit in the manager-develop completion handoff.

### F.4 Risk 4 — Sprint 10 lane-A race

**Probability**: Medium. **Impact**: Low (race-absorbed pattern per L52 — disjoint scope avoids conflict).

**Mitigation**: Lane B SPEC scope is disjoint from any concurrent lane-A SPEC (lane A would touch different paths). Pre-spawn fetch (agent-common-protocol.md §Pre-Spawn Sync Check) catches origin-ahead state at every commit boundary.

## G. HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial plan-phase authoring — 3-scope consolidation (W3-arch + W4 + W5) per Sprint 10 lane B entry directive. 13 REQ-LNC + 4 HARD constraints + 4 risks. SPEC ID pre-write self-check PASSED (decomposition: SPEC ✓ \| V3R6 ✓ \| LOCAL ✓ \| NAMESPACE ✓ \| CONSOLIDATION ✓ \| 001 ✓ → PASS). Frontmatter 12-canonical-field validation PASSED. |
| 0.1.1 | 2026-05-25 | manager-spec | iter-2 focused defect resolution per plan-auditor iter-1 0.73 FAIL verdict — D3 HARD #3 `§24.4` drop + §E out-of-scope clarification (CLAUDE.local.md update deferred), D6 REQ-LNC-011 broaden to 2 in-scope SSOTs + REQ-LNC-014 NEW Where-capability, D7 `dev-only-commands-isolation.md` local-only acknowledgement, D9 `8-phase` → `9-phase (Phase 0 through Phase 8)`, D11 REQ-LNC-008 verbatim → structural fidelity, D14 REQ-LNC-002 vs REQ-LNC-013 reconciliation (REQ-LNC-002 cross-refs REQ-LNC-013 for LOC budget, retains YAML structural invariant). REQ count 13 → 14 (REQ-LNC-014 added). AC count 11 → 11 (AC-LNC-006 binding broadens to REQ-LNC-011 + REQ-LNC-014). 5 HARD constraints preserved. Frontmatter tier:M added per D13. |
