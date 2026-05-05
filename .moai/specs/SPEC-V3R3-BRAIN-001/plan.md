---
id: SPEC-V3R3-BRAIN-001
version: "0.1.0"
status: draft
created_at: 2026-05-04
updated_at: 2026-05-04
author: MoAI Plan Workflow
priority: P1
labels: [brain, ideation, workflow, handoff, claude-design, v3r3, plan]
issue_number: null
phase: "v3.0.0 — Phase 8 — Brain Workflow Introduction"
module: "internal/cli/, .claude/skills/moai/workflows/, .claude/skills/moai-domain-ideation/, .claude/skills/moai-domain-research/, .claude/skills/moai-domain-design-handoff/, .claude/agents/manager-brain.md, .claude/commands/moai-brain.md, .moai/brain/"
breaking: false
harness: standard
---

# Plan: SPEC-V3R3-BRAIN-001 — `/moai brain` Implementation Plan

## 1. Strategy Overview

This SPEC introduces a brand-new pre-spec ideation workflow. The implementation strategy follows the Template-First discipline mandated by `CLAUDE.local.md` §2: every artifact under `.claude/`, `.moai/` is added to `internal/template/templates/` first, then regenerated via `make build`, then exposed locally.

**Approach**: Skill-first build order, then agent, then CLI binding, then integration patches, then tests. Skills can be developed in isolation against the foundation primitives without depending on the agent/CLI being ready.

**Reuse strategy**: The brain workflow is a composition over `moai-foundation-thinking`. New `moai-domain-*` skills are thin orchestrators that delegate phase-specific logic to the foundation. This minimizes new-code risk and aligns with the constitutional principle "Reuse over re-invention" (CLAUDE.md §1).

---

## 2. Implementation Phases

### Phase A1: Foundation Skills (Domain Skills)

**Files**: deliverables #2, #3, #4 (3 new skills under `.claude/skills/moai-domain-*/`)
**Dependencies**: None (only `moai-foundation-thinking` which already exists)
**Estimated LOC**: ~1,200 markdown

Tasks:
1. Create `.claude/skills/moai-domain-ideation/SKILL.md` — thin orchestrator for Diverge → Cluster → Converge pipeline. Delegates 90% to `moai-foundation-thinking/modules/diverge-converge.md`. Adds Lean Canvas section assembly logic (Problem / Customer / UVP / Solution / Channels / Revenue / Cost / Metrics / Unfair Advantage). Adds "SPEC Decomposition Candidates" assembler with grammar `- SPEC-{DOMAIN}-{NUM}: {scope}`.
2. Create `.claude/skills/moai-domain-research/SKILL.md` — patterns for parallel WebSearch + Context7 + market scan. Documents Anthropic's parallel-tool-call convention. Includes failure-graceful pattern (REQ-BRAIN-003 partial-failure tolerance).
3. Create `.claude/skills/moai-domain-design-handoff/SKILL.md` — Claude Design prompt template (5-section: Goal / References / Brand Voice / Acceptance / Out-of-scope). Documents brand-absent fallback per REQ-BRAIN-006.
4. Mirror all 3 skills to `internal/template/templates/.claude/skills/moai-domain-*/SKILL.md`.

**Verification**: Skills are valid markdown with proper frontmatter. No runtime testing required at this phase — skills are loaded lazily.

### Phase A2: Workflow Orchestration Skill

**Files**: deliverable #1 (`.claude/skills/moai/workflows/brain.md`)
**Dependencies**: Phase A1 complete (skills must exist before workflow references them)
**Estimated LOC**: ~600 markdown

Tasks:
1. Create `.claude/skills/moai/workflows/brain.md` with the 7-phase contract:
   - Phase 1 Discovery (uses `moai-foundation-thinking/modules/deep-questioning.md` + AskUserQuestion)
   - Phase 2 Diverge (uses `moai-domain-ideation`)
   - Phase 3 Research (uses `moai-domain-research`)
   - Phase 4 Converge (uses `moai-domain-ideation`)
   - Phase 5 Critical Evaluation (uses `moai-foundation-thinking/modules/critical-evaluation.md` + `first-principles.md`)
   - Phase 6 Proposal (uses `moai-domain-ideation` for SPEC decomposition)
   - Phase 7 Handoff (uses `moai-domain-design-handoff`)
2. Document IDEA-NNN auto-increment logic (scan `.moai/brain/IDEA-*` directories, take max + 1).
3. Document brand-context detection (`.moai/project/brand/brand-voice.md` existence check).
4. Document REQ-BRAIN-009 next-action AskUserQuestion (with options a/b/c).
5. Mirror to `internal/template/templates/.claude/skills/moai/workflows/brain.md`.

**Verification**: Manual review of phase contracts; cross-reference REQ-BRAIN-001..012 against workflow steps.

### Phase A3: Workflow Router Patch

**Files**: deliverable #5 (`.claude/skills/moai/SKILL.md` patch)
**Dependencies**: Phase A2 complete (workflow file must exist before router references it)
**Estimated LOC**: ~30 markdown

Tasks:
1. Add `brain` to subcommand router section in `.claude/skills/moai/SKILL.md`. Match the existing pattern used for `plan`, `run`, `sync`, etc.
2. Add "Priority 1 routing" entry: `/moai brain "<idea>"` → load `workflows/brain.md`.
3. Mirror to template.

**Verification**: Manual diff review against existing subcommand entries.

### Phase A4: Agent Definition

**Files**: deliverable #6 (`.claude/agents/manager-brain.md`)
**Dependencies**: Phase A2 complete
**Estimated LOC**: ~250 markdown

Tasks:
1. Create `.claude/agents/manager-brain.md` following the `manager-spec.md` pattern.
2. Frontmatter: `name`, `description`, `model: opus`, `tools: Read, Write, Edit, Glob, Grep, Bash, WebSearch, WebFetch, ToolSearch, AskUserQuestion, mcp__context7__*`.
3. Body: workflow phases 1-7 prose, delegation patterns, blocker-report contract for missing inputs.
4. Mirror to `internal/template/templates/.claude/agents/manager-brain.md`.

**Verification**: `internal/template/agents_audit_test.go` (existing test) verifies agent frontmatter and body structure on next `make build`.

### Phase A5: Command Wrappers

**Files**: deliverables #7, #8 (`.claude/commands/moai-brain.md` NEW, `.claude/commands/moai.md` PATCH)
**Dependencies**: Phase A2, A4 complete
**Estimated LOC**: #7 ≤80 lines (≤20 LOC body); #8 patch ~10 lines

Tasks:
1. Create `.claude/commands/moai-brain.md`:
   - Frontmatter per Thin Command Pattern (`description`, `argument-hint`, `allowed-tools` CSV, `model`).
   - Body: ≤20 LOC, single instruction "Use the moai workflow brain skill to execute /moai brain with arguments: {{args}}".
2. Patch `.claude/commands/moai.md` to add `brain` to recognized subcommands list.
3. Mirror both to template.

**Verification**: `internal/template/commands_audit_test.go` extended (deliverable #17) checks Thin Command Pattern compliance.

### Phase A6: Output Directory Templates

**Files**: deliverables #11, #12
**Dependencies**: None
**Estimated LOC**: ~250 markdown total (5 example files × ~50 lines)

Tasks:
1. Create `internal/template/templates/.moai/brain/.gitkeep` (zero bytes).
2. Create `internal/template/templates/.moai/brain/IDEA-EXAMPLE/`:
   - `research.md` (worked example: market research for hypothetical SaaS)
   - `ideation.md` (worked example: Lean Canvas filled in)
   - `proposal.md` (worked example: 3 SPEC decomposition candidates)
   - `claude-design-handoff/prompt.md` (worked example: paste-ready prompt)
   - `claude-design-handoff/context.md`, `references.md`, `acceptance.md`, `checklist.md` (worked stubs)
3. Examples MUST follow REQ-BRAIN-008 (16-language neutrality) — example proposal does not name a specific tech stack.

**Verification**: `moai init` in `/tmp/test-project` produces `.moai/brain/IDEA-EXAMPLE/` directory with all 8 files.

### Phase A7: Go CLI Entry

**Files**: deliverables #9, #10
**Dependencies**: Phase A2 (slash command must exist for delegation)
**Estimated LOC**: brain.go ~150 LOC; root.go patch ~10 LOC

Tasks:
1. Create `internal/cli/brain.go` with cobra command `brainCmd`:
   - Use: `brain "<idea>"`
   - Short: "Run the /moai brain ideation workflow"
   - Run: prints user-facing instruction "Run `/moai brain \"<idea>\"` in Claude Code chat — `moai brain` is a CLI hint, the actual workflow runs in Claude Code session."
   - Optional: `--instructions-only` flag to print 7-phase contract without spawning.
2. Patch `internal/cli/root.go` to register `brainCmd`.
3. Follow §1 CLI vs slash boundary in `CLAUDE.local.md`: terminal `moai brain` is informational; the workflow IS a slash command.

**Verification**: `go build ./...` succeeds; `moai brain --help` prints expected text.

### Phase A8: Workflow Integration Patches

**Files**: deliverables #13, #14, #15
**Dependencies**: Phase A2 complete (brain workflow output format defined)
**Estimated LOC**: project.md +80, plan.md +100, design.md +40

Tasks:
1. Patch `.claude/skills/moai/workflows/project.md`:
   - Add CLI flag `--from-brain IDEA-XXX`.
   - When flag present, load `.moai/brain/IDEA-XXX/proposal.md` and prepend its content to the product/structure/tech generation prompt.
   - Document precedence: `proposal.md` (primary) > codebase scan (secondary).
2. Patch `.claude/skills/moai/workflows/plan.md`:
   - Detect `.moai/brain/IDEA-*/proposal.md` existence on `/moai plan` invocation.
   - Parse "SPEC Decomposition Candidates" section using grammar `- SPEC-{DOMAIN}-{NUM}: {scope}`.
   - Surface candidates to user via AskUserQuestion (recommendation field) — never auto-create SPECs.
   - Defensive parser: warn (not error) on grammar drift.
3. Patch `.claude/skills/moai/workflows/design.md`:
   - When `--path A` and no `--bundle` arg, scan `.moai/brain/IDEA-*/claude-design-handoff/` for candidate bundles.
   - Suggest detected bundles via AskUserQuestion.
4. Mirror all 3 patches to template.

**Verification**: Manual workflow trace for each integration point.

### Phase A9: Tests

**Files**: deliverables #16, #17
**Dependencies**: Phases A1-A8 complete
**Estimated LOC**: brain_test.go ~200 LOC; commands_audit_test.go extension ~50 LOC

Tasks:
1. Create `internal/cli/brain_test.go`:
   - Table-driven tests for `brainCmd.Run` with various args.
   - `t.TempDir()` isolation per `CLAUDE.local.md` §6.
   - Test `--help` output structure.
   - Test `--instructions-only` flag.
   - Test missing args error message.
2. Extend `internal/template/commands_audit_test.go`:
   - Add `moai-brain.md` to the list of commands verified for Thin Command Pattern.
   - Verify body LOC ≤20.
   - Verify frontmatter completeness.
3. Run full test suite: `go test ./... -race -count=1`.

**Verification**: All tests pass; coverage for `brain.go` ≥85%.

---

## 3. Risk Analysis

### Risk 1: Phase 7 Prompt Quality (HIGH)

**Description**: A poorly-templated `prompt.md` produces Claude Design output missing the user's brand voice or feature scope, requiring manual editing — defeating REQ-BRAIN-005's "paste-ready" promise.

**Probability**: Medium (template quality is the only variable; user testing will surface gaps quickly)
**Impact**: High (user friction on the headline feature)

**Mitigation**:
- Template includes self-review checklist before final write.
- Phase 7 acceptance gate: AskUserQuestion confirms user reviewed handoff package before workflow exit.
- IDEA-EXAMPLE/ template ships with worked example so users see expected output shape.
- `--regenerate` flag (REQ-BRAIN-009 option c) allows iteration without re-running phases 1-6.
- Acceptance scenario #1 verifies all 5 handoff files are produced with required structure.

### Risk 2: Brand Context Absent (MEDIUM)

**Description**: First-time projects without `.moai/project/brand/` populated produce generic-voice prompts. REQ-BRAIN-006 specifies brand integration but does not define absent-fallback.

**Probability**: High (brand interview is opt-in)
**Impact**: Medium (workflow still completes; output quality reduced)

**Mitigation**:
- When brand context absent, Phase 7 emits "Brand Voice (default — please customize)" section with placeholder warning.
- AskUserQuestion offers user the option to run brand interview before Phase 7 (or accept default).
- Acceptance scenario #2 verifies graceful fallback.

### Risk 3: Claude.com Design External Dependency (MEDIUM)

**Description**: Claude Design product UX changes break prompt template assumptions; the workflow has no direct integration test with claude.com.

**Probability**: Low (Claude Design is stable per Anthropic roadmap)
**Impact**: Medium (workflow output stale, requires template update)

**Mitigation**:
- prompt.md template designed as human-readable instructions, not machine-formatted — robust to UX changes.
- references.md and acceptance.md are static design-quality artifacts independent of Claude Design's UI.
- This SPEC's acceptance scenario #1 stops at "handoff package generated" — does not require Claude Design execution.
- Long-term: `references/claude-design-prompt-best-practices.md` reference doc tracks Anthropic's published guidance.

### Risk 4: Self-Bootstrap Order Pressure (LOW)

**Description**: Pressure to start SPEC-V3R3-WEB-001 before BRAIN ships could lead to brain workflow being designed around web-specific assumptions, breaking generality.

**Probability**: Low (SPEC sequencing is enforced by dependency declaration)
**Impact**: High if it occurs (REQ-BRAIN-008 violation, generality lost)

**Mitigation**:
- This SPEC explicitly lists "Web UI for brain" in Out-of-Scope §7.
- REQ-BRAIN-008 (16-language neutrality) is testable at acceptance — see scenario #5.
- IDEA-EXAMPLE/ ships with non-web-specific worked example.

### Risk 5: SPEC Decomposition Format Drift (LOW)

**Description**: proposal.md's "SPEC Decomposition Candidates" section format diverges from what `/moai plan --from-brain` expects.

**Probability**: Low (template fixes the grammar)
**Impact**: Low (defensive parser surfaces warnings)

**Mitigation**:
- proposal.md template fixes section grammar: `### SPEC Decomposition Candidates` heading + bullet list `- SPEC-{DOMAIN}-{NUM}: {scope}`.
- plan.md patch implements defensive parser: surfaces non-conforming entries as warnings, never errors.
- Acceptance scenario #4 verifies parseability.

---

## 4. Task Decomposition (Granular)

| Task ID | Phase | Description | Deliverables | Priority | Status |
|---------|-------|-------------|--------------|----------|--------|
| T-A1.1 | A1 | Create moai-domain-ideation skill | #2 | P1 | pending |
| T-A1.2 | A1 | Create moai-domain-research skill | #3 | P1 | pending |
| T-A1.3 | A1 | Create moai-domain-design-handoff skill | #4 | P1 | pending |
| T-A1.4 | A1 | Mirror domain skills to template | (mirrors of #2/3/4) | P1 | pending |
| T-A2.1 | A2 | Create workflows/brain.md (7-phase contract) | #1 | P1 | pending |
| T-A2.2 | A2 | Mirror brain.md to template | (mirror of #1) | P1 | pending |
| T-A3.1 | A3 | Patch moai/SKILL.md router for `brain` subcommand | #5 | P1 | pending |
| T-A4.1 | A4 | Create manager-brain agent | #6 | P1 | pending |
| T-A5.1 | A5 | Create moai-brain.md command (Thin Pattern) | #7 | P1 | pending |
| T-A5.2 | A5 | Patch commands/moai.md subcommand list | #8 | P1 | pending |
| T-A6.1 | A6 | Create .moai/brain/.gitkeep template | #11 | P1 | pending |
| T-A6.2 | A6 | Create IDEA-EXAMPLE/ worked example (8 files) | #12 | P2 | pending |
| T-A7.1 | A7 | Create internal/cli/brain.go | #9 | P1 | pending |
| T-A7.2 | A7 | Patch internal/cli/root.go (register brainCmd) | #10 | P1 | pending |
| T-A8.1 | A8 | Patch project.md (--from-brain flag) | #13 | P1 | pending |
| T-A8.2 | A8 | Patch plan.md (decomposition parser) | #14 | P1 | pending |
| T-A8.3 | A8 | Patch design.md (bundle auto-detect) | #15 | P2 | pending |
| T-A9.1 | A9 | Create internal/cli/brain_test.go | #16 | P1 | pending |
| T-A9.2 | A9 | Extend commands_audit_test.go | #17 | P1 | pending |
| T-A9.3 | A9 | Run full test suite + lint | (verification) | P1 | pending |
| T-A9.4 | A9 | `make build` (regenerate embedded templates) | (verification) | P1 | pending |

**Total tasks**: 21
**Sequential ordering**: A1 → A2 → A3 → A4 → A5 → A6 → A7 → A8 → A9 (later phases may parallelize within their bounds)

---

## 5. Reference Implementations

| Pattern | Reference | Used For |
|---------|-----------|----------|
| Thin Command Pattern | `.claude/commands/moai/plan.md` (existing) | deliverable #7 |
| Workflow Orchestration | `.claude/skills/moai/workflows/plan.md` (existing) | deliverable #1 |
| Manager Agent Definition | `.claude/agents/moai/manager-spec.md` (existing) | deliverable #6 |
| Domain Skill Structure | `.claude/skills/moai-domain-backend/` (existing) | deliverables #2, #3, #4 |
| AskUserQuestion + ToolSearch Preload | `.claude/rules/moai/core/askuser-protocol.md` | Phase 1 Discovery, Phase 6, Phase 7 |
| Template-First Discipline | `internal/template/templates/.moai/specs/` (existing) | deliverables #11, #12 |
| Cobra Command Registration | `internal/cli/plan.go` + `root.go` (existing) | deliverables #9, #10 |
| Commands Audit Test | `internal/template/commands_audit_test.go` (existing) | deliverable #17 |

---

## 6. MX Tag Plan

Per `.claude/rules/moai/workflow/mx-tag-protocol.md`, the following MX tags will be added during the run phase:

### @MX:NOTE (Context delivery)

- `internal/cli/brain.go` — explain why this is a thin CLI wrapper (delegation to slash command).
- `.claude/skills/moai/workflows/brain.md` — explain the 7-phase contract and its derivation from `moai-foundation-thinking`.

### @MX:WARN (Danger zone)

- `.claude/skills/moai/workflows/brain.md` — Phase 7 prompt template generation. **Reason**: Output is paste-ready into external system (claude.com); changes affect user trust. @MX:REASON required for any template modification.
- `internal/cli/brain.go` — argument parsing. **Reason**: User-facing CLI; error messages must be helpful (16-language neutrality applies to error messages too).

### @MX:ANCHOR (Invariant contract)

- `.claude/skills/moai-domain-design-handoff/SKILL.md` — `prompt.md` template structure (5 sections: Goal / References / Brand Voice / Acceptance / Out-of-scope). **High fan-in**: consumed by every brain workflow execution.
- `.claude/skills/moai-domain-ideation/SKILL.md` — "SPEC Decomposition Candidates" section grammar. **High fan-in**: consumed by `/moai plan --from-brain` patch.

### @MX:TODO (Incomplete work — resolved in run phase)

- `internal/template/templates/.moai/brain/IDEA-EXAMPLE/` — initial example may be domain-specific; should be replaced with a more generic example based on user feedback in v0.2. (Resolution: keep current, document as "subject to v0.2 refinement" comment.)

---

## 7. Estimated LOC

| Component | LOC | Confidence |
|-----------|-----|------------|
| `.claude/skills/moai/workflows/brain.md` | ~600 | High |
| `moai-domain-ideation/SKILL.md` | ~400 | High |
| `moai-domain-research/SKILL.md` | ~350 | Medium |
| `moai-domain-design-handoff/SKILL.md` | ~450 | Medium |
| `manager-brain.md` | ~250 | High |
| `moai-brain.md` (Thin Command) | ~80 | High |
| `moai.md` patch | ~10 | High |
| `moai/SKILL.md` patch (router) | ~30 | High |
| `internal/cli/brain.go` | ~150 | High |
| `internal/cli/root.go` patch | ~10 | High |
| IDEA-EXAMPLE/ (8 files × ~50) | ~250 | Medium |
| `project.md` patch | ~80 | High |
| `plan.md` patch | ~100 | Medium |
| `design.md` patch | ~40 | High |
| `brain_test.go` | ~200 | High |
| `commands_audit_test.go` extension | ~50 | High |
| **Total** | **~3,050** | Medium-High |

---

## 8. Build & Quality Gates

### 8.1 Build Steps (executed in Phase A9)

1. `make build` — regenerate `internal/template/embedded.go` from `internal/template/templates/`
2. `go test ./... -race -count=1` — full test suite, no caching
3. `golangci-lint run ./...` — zero warnings target
4. `go vet ./...` — static analysis
5. Manual smoke test: `./moai brain --help` produces expected output

### 8.2 plan-auditor Gate

This SPEC has `harness: standard`. plan-auditor (independent skeptical assessor) is auto-engaged at Phase 2.3 of `/moai plan` workflow. Audit checklist:

- [ ] Frontmatter 9 required fields present in spec.md
- [ ] EARS requirements use proper keywords (WHEN/WHILE/WHERE/IF/SHALL)
- [ ] Acceptance scenarios in Given-When-Then format (5+ scenarios)
- [ ] Out-of-Scope section non-empty (10 entries in this SPEC)
- [ ] Risk analysis covers minimum 3 risks (this SPEC: 5 risks)
- [ ] Reuse inventory present (this SPEC: research §5)
- [ ] No implementation code in spec.md (only WHAT/WHY)
- [ ] No time estimates in plan.md (priority labels only — verified above)
- [ ] Cross-references to existing skills/agents validated (this SPEC: research §1)
- [ ] 16-language neutrality verified (REQ-BRAIN-008)

### 8.3 Run-Phase Quality Gates (LSP-based, per CLAUDE.md §6)

- Zero LSP errors
- Zero type errors
- Zero lint errors (`golangci-lint`)
- Test coverage ≥85% for `internal/cli/brain.go`

---

## 9. Sequencing Strategy (Run Phase Recommendation)

**Recommended execution mode**: `sequential` (single-developer, single-branch).

**Rationale**:
- 21 tasks but strong dependency chain (A1 → A2 → A3 → ...)
- File ownership concentrated in `.claude/skills/` (3 new skills) and `internal/cli/` (1 new file + 1 patch) — no parallel write conflict potential
- Tests in Phase A9 verify all earlier phases — running A1-A8 in parallel risks broken intermediate state
- Worktree isolation NOT required (single SPEC, no cross-SPEC dependency)

**Alternative**: If pressure to ship faster, Phase A1's 3 domain skills can be developed in parallel (independent files, no cross-references between them at draft time). All other phases must remain sequential.

---

## 10. Cross-Reference Validation

Validation that this plan aligns with the SPEC requirements:

| REQ | Verified by Phase |
|-----|-------------------|
| REQ-BRAIN-001 (7-phase execution) | Phase A2 (workflows/brain.md) |
| REQ-BRAIN-002 (Discovery rounds ≤5) | Phase A2 + Phase A4 (manager-brain) |
| REQ-BRAIN-003 (parallel research) | Phase A1.2 (moai-domain-research) |
| REQ-BRAIN-004 (SPEC decomposition list) | Phase A1.1 (moai-domain-ideation) + Phase A8.2 (plan.md patch) |
| REQ-BRAIN-005 (paste-ready prompt) | Phase A1.3 (moai-domain-design-handoff) |
| REQ-BRAIN-006 (brand integration) | Phase A1.3 + Phase A2 (brand-detect logic) |
| REQ-BRAIN-007 (--from-brain flag) | Phase A8.1 (project.md patch) |
| REQ-BRAIN-008 (16-language neutrality) | Phase A6.2 (IDEA-EXAMPLE) + acceptance scenario #5 |
| REQ-BRAIN-009 (next-action AskUserQuestion) | Phase A2 (workflow Phase 7 exit) |
| REQ-BRAIN-010 (no auto-project) | Phase A2 (negative invariant in workflow) |
| REQ-BRAIN-011 (no tech-stack in proposal) | Phase A1.1 (ideation skill rule) + acceptance scenario #5 |
| REQ-BRAIN-012 (AskUserQuestion enforcement) | Phase A2 + Phase A4 (agent ban on prose questions) |

All 12 requirements traced to implementation phases.

---

## 11. References

- `research.md` (sibling) — codebase context, decisions, risk surface
- `spec.md` (sibling) — EARS requirements, file deliverables, exclusions
- `acceptance.md` (sibling) — Given-When-Then scenarios
- `.claude/rules/moai/development/coding-standards.md` — Thin Command Pattern, frontmatter schema
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — MX tag types and reasoning
- `CLAUDE.local.md` §2 (Template-First), §6 (Test Isolation), §16 (Self-check)

---

**Status**: audit-ready (post plan-auditor PASS at Phase 2.3)
