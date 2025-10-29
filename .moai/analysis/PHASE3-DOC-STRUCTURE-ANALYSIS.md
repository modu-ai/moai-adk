# MoAI-ADK Documentation Structure Analysis for Phase 3 Planning

**Report Date**: 2025-10-29  
**Current Branch**: feature/DOC-TAG-001-auto-generation  
**Analysis Scope**: Complete documentation hierarchy for @DOC tag migration

---

## Executive Summary

### Current State (v0.7.0)
- **Total markdown files in docs/**: 78 files (17,392 lines)
- **Tagged with @DOC**: 45 files (57.7%)
- **Untagged files**: 33 files (42.3%) - HIGH PRIORITY for Phase 3
- **Other documentation**: 81 files in .moai/ directories (specs, reports, memory)
- **Root documentation**: 9 files (README variants + CLAUDE.md, CHANGELOG, CONTRIBUTING)

### Coverage Status by Category
| Category | Total | Tagged | Untagged | % Tagged | Status |
|----------|-------|--------|----------|----------|--------|
| agents | 13 | 12 | 1 | 92.3% | NEAR COMPLETE |
| api-reference | 5 | 4 | 1 | 80.0% | MOSTLY COMPLETE |
| commands | 4 | 3 | 1 | 75.0% | MOSTLY COMPLETE |
| configuration | 3 | 3 | 0 | 100% | COMPLETE |
| contributing | 5 | 5 | 0 | 100% | COMPLETE |
| getting-started | 3 | 3 | 0 | 100% | COMPLETE |
| guides | 21 | 0 | 21 | 0% | CRITICAL GAP |
| hooks | 5 | 5 | 0 | 100% | COMPLETE |
| security | 4 | 4 | 0 | 100% | COMPLETE |
| troubleshooting | 3 | 3 | 0 | 100% | COMPLETE |
| root level | 7 | 3 | 4 | 42.9% | NEEDS WORK |
| skills | 5 | 0 | 5 | 0% | CRITICAL GAP |
| status | 1 | 0 | 1 | 0% | NEEDS WORK |

---

## 1. Complete File Inventory

### Section A: Configuration & Setup (100% Complete)
**Status**: All files tagged. No action needed.

#### configuration/ (3/3 tagged)
- `advanced-settings.md` (122 lines) → `@DOC:CONFIG-ADV-001`
- `config-json.md` (177 lines) → `@DOC:CONFIG-JSON-001`
- `personal-vs-team.md` (94 lines) → `@DOC:CONFIG-MODE-001`

#### getting-started/ (3/3 tagged)
- `installation.md` (167 lines) → `@DOC:START-INSTALL-001`
- `quick-start.md` (126 lines) → `@DOC:START-QUICK-001`
- `first-project.md` (144 lines) → `@DOC:START-FIRST-001`

#### contributing/ (5/5 tagged)
- `overview.md` (18 lines) → `@DOC:CONTRIB-OVERVIEW-001`
- `development-setup.md` (20 lines) → `@DOC:CONTRIB-DEV-001`
- `code-style.md` (12 lines) → `@DOC:CONTRIB-STYLE-001`
- `pull-request-process.md` (12 lines) → `@DOC:CONTRIB-PR-001`
- `testing.md` (12 lines) → `@DOC:CONTRIB-TEST-001`

### Section B: System Features (95-100% Complete)
**Status**: Mostly complete. 4 files need tagging.

#### agents/ (12/13 tagged - 92.3%)
**Missing**: `overview.md` (55 lines) → Propose: `@DOC:AGENT-OVERVIEW-001`

Existing (tagged):
- `cc-manager.md` → `@DOC:AGENT-CC-001`
- `code-builder.md` → `@DOC:AGENT-CODE-001`
- `debug-helper.md` → `@DOC:AGENT-DEBUG-001`
- `doc-syncer.md` → `@DOC:AGENT-DOC-001`
- `git-manager.md` (92 lines) → `@DOC:AGENT-GIT-001`
- `implementation-planner.md` → `@DOC:AGENT-IMPL-PLANNER-001`
- `project-manager.md` → `@DOC:AGENT-PM-001`
- `quality-gate.md` (45 lines) → `@DOC:AGENT-QUALITY-GATE-001`
- `spec-builder.md` → `@DOC:AGENT-SPEC-001`
- `tag-agent.md` → `@DOC:AGENT-TAG-001`
- `tdd-implementer.md` → `@DOC:AGENT-TDD-IMPL-001`
- `trust-checker.md` → `@DOC:AGENT-TRUST-001`

#### commands/ (3/4 tagged - 75.0%)
**Missing**: `cli.md` (76 lines) → Propose: `@DOC:CMD-CLI-INTRO-001` or `@DOC:CMD-GENERAL-001`

Existing (tagged):
- `cli-reference.md` (112 lines) → `@DOC:CMD-CLI-001`
- `agent-commands.md` (80 lines) → `@DOC:CMD-AGENT-001`
- `alfred-commands.md` (99 lines) → `@DOC:CMD-ALFRED-001`

#### api-reference/ (4/5 tagged - 80.0%)
**Missing**: `agents.md` (37 lines) → Propose: `@DOC:API-AGENTS-001`

Existing (tagged):
- `core-git.md` (24 lines) → `@DOC:API-GIT-001`
- `core-template.md` (24 lines) → `@DOC:API-TEMPLATE-001`
- `core-tag.md` (24 lines) → `@DOC:API-TAG-001`
- `core-installer.md` (24 lines) → `@DOC:API-INSTALLER-001`

#### hooks/ (5/5 tagged - 100%)
- `overview.md` (16 lines) → `@DOC:HOOK-OVERVIEW-001`
- `session-start-hook.md` (16 lines) → `@DOC:HOOK-SESSION-001`
- `pre-tool-use-hook.md` (95 lines) → `@DOC:HOOK-PRE-001`
- `post-tool-use-hook.md` (135 lines) → `@DOC:HOOK-POST-001`
- `custom-hooks.md` (17 lines) → `@DOC:HOOK-CUSTOM-001`

#### security/ (4/4 tagged - 100%)
- `overview.md` (138 lines) → `@DOC:SEC-OVERVIEW-001`
- `best-practices.md` (12 lines) → `@DOC:SEC-BEST-001`
- `checklist.md` (12 lines) → `@DOC:SEC-CHECK-001`
- `template-security.md` (11 lines) → `@DOC:SEC-TEMPLATE-001`

#### troubleshooting/ (3/3 tagged - 100%)
- `common-errors.md` (22 lines) → `@DOC:TROUBLESHOOT-ERROR-001`
- `debugging-guide.md` (18 lines) → `@DOC:TROUBLESHOOT-DEBUG-001`
- `faq.md` (14 lines) → `@DOC:TROUBLESHOOT-FAQ-001`

### Section C: Critical Gaps (0% Complete)

#### guides/ (0/21 tagged - 0% CRITICAL)
**Impact**: Largest untagged section. Contains essential user guides, architecture docs, and tutorials.

**Concepts Subdirectory** (5 files, 2,385 lines):
- `concepts/spec-first-tdd.md` (773 lines) - LARGE
  → Propose: `@DOC:GUIDE-CONCEPT-SPEC-001`
- `concepts/tag-system.md` (646 lines) - LARGE
  → Propose: `@DOC:GUIDE-CONCEPT-TAG-001`
- `concepts/trust-principles.md` (554 lines) - LARGE
  → Propose: `@DOC:GUIDE-CONCEPT-TRUST-001`
- `concepts/ears-guide.md` (314 lines)
  → Propose: `@DOC:GUIDE-CONCEPT-EARS-001`
- `concepts/skills-revolution.md` (98 lines)
  → Propose: `@DOC:GUIDE-CONCEPT-SKILLS-001`

**Workflow Subdirectory** (6 files, 4,602 lines):
- `workflow/0-project.md` (2544 lines) - VERY LARGE
  → Propose: `@DOC:GUIDE-WORKFLOW-INIT-001`
- `workflow/9-update.md` (1715 lines) - LARGE
  → Propose: `@DOC:GUIDE-WORKFLOW-UPDATE-001`
- `workflow/overview.md` (111 lines)
  → Propose: `@DOC:GUIDE-WORKFLOW-OVERVIEW-001`
- `workflow/1-plan.md` (122 lines)
  → Propose: `@DOC:GUIDE-WORKFLOW-PLAN-001`
- `workflow/2-run.md` (108 lines)
  → Propose: `@DOC:GUIDE-WORKFLOW-RUN-001`
- `workflow/3-sync.md` (103 lines)
  → Propose: `@DOC:GUIDE-WORKFLOW-SYNC-001`

**Examples Subdirectory** (8 files, 4,905 lines):
- `examples/index.md` (114 lines)
  → Propose: `@DOC:GUIDE-TUTORIAL-INDEX-001`
- `examples/todo-app/index.md` (361 lines)
  → Propose: `@DOC:GUIDE-TUTORIAL-TODOAPP-001`
- `examples/todo-app/01-project-init.md` (731 lines) - LARGE
  → Propose: `@DOC:GUIDE-TUTORIAL-TODOAPP-INIT-001`
- `examples/todo-app/02-spec-writing.md` (837 lines) - LARGE
  → Propose: `@DOC:GUIDE-TUTORIAL-TODOAPP-SPEC-001`
- `examples/todo-app/03-backend-tdd.md` (988 lines) - LARGE
  → Propose: `@DOC:GUIDE-TUTORIAL-TODOAPP-BACKEND-001`
- `examples/todo-app/04-frontend-impl.md` (871 lines) - LARGE
  → Propose: `@DOC:GUIDE-TUTORIAL-TODOAPP-FRONTEND-001`
- `examples/todo-app/05-sync-deploy.md` (1003 lines) - LARGE
  → Propose: `@DOC:GUIDE-TUTORIAL-TODOAPP-DEPLOY-001`

**Other Guides** (2 files, 829 lines):
- `architecture.md` (124 lines)
  → Propose: `@DOC:GUIDE-ARCHITECTURE-001`
- `agents/overview.md` (705 lines) - LARGE
  → Propose: `@DOC:GUIDE-AGENTS-001`
- `hooks/overview.md` (526 lines) - LARGE
  → Propose: `@DOC:GUIDE-HOOKS-001`

#### skills/ (0/5 tagged - 0% CRITICAL)
**Impact**: Documentation for Skills tier system. Essential for understanding framework architecture.

- `overview.md` (90 lines)
  → Propose: `@DOC:SKILL-OVERVIEW-001`
- `foundation-tier.md` (75 lines)
  → Propose: `@DOC:SKILL-FOUNDATION-001`
- `essentials-tier.md` (63 lines)
  → Propose: `@DOC:SKILL-ESSENTIALS-001`
- `domain-tier.md` (36 lines)
  → Propose: `@DOC:SKILL-DOMAIN-001`
- `language-tier.md` (68 lines)
  → Propose: `@DOC:SKILL-LANGUAGE-001`

### Section D: Root-Level Files (3/7 tagged - 42.9%)

**Existing (tagged)**:
- `introduction.md` (125 lines) → `@DOC:INTRO-001`
- `workflow.md` (118 lines) → `@DOC:WORKFLOW-001`
- `security-scanning.md` (116 lines) → `@DOC:SECURITY-001`

**Missing**:
- `index.md` (141 lines) - Main documentation portal
  → Propose: `@DOC:INDEX-MAIN-001`
- `MULTILINGUAL-MENUS.md` (263 lines) - Menu internationalization
  → Propose: `@DOC:CONFIG-MULTILING-001`
- `skills-optimization-phase2-3-report.md` (397 lines) - Analysis report
  → Propose: `@DOC:STATUS-OPTIM-001`
- `status/sync-report.md` (272 lines) - Project status
  → Propose: `@DOC:STATUS-SYNC-001`

---

## 2. Recommended Domain Taxonomy for Phase 3

### Established Domains (No Changes)
```
@DOC:AGENT-*          → Agent documentation [12 files tagged, 1 to complete]
@DOC:API-*            → API reference [4 files tagged, 1 to complete]
@DOC:CMD-*            → Commands & CLI [3 files tagged, 1 to complete]
@DOC:CONFIG-*         → Configuration [3 files tagged]
@DOC:CONTRIB-*        → Contributing [5 files tagged]
@DOC:HOOK-*           → Hooks system [5 files tagged]
@DOC:SEC-*            → Security [4 files tagged]
@DOC:START-*          → Getting Started [3 files tagged]
@DOC:TROUBLESHOOT-*   → Troubleshooting [3 files tagged]
```

### New Domains (Proposed for Phase 3)
```
@DOC:GUIDE-*          → Learning guides, tutorials, walkthroughs [NEW - 21 files]
@DOC:SKILL-*          → Skills tier documentation [NEW - 5 files]
@DOC:INDEX-*          → Documentation portals and main indices [NEW - 1 file]
@DOC:STATUS-*         → Status reports and analysis [NEW - 2 files]
```

### Domain Assignment Strategy

**For GUIDE Domain (21 files)**:
Pattern: `@DOC:GUIDE-{SUBTYPE}-{NAME}-{SEQUENCE}`

Subtypes:
- `CONCEPT` - Conceptual frameworks & theory
- `WORKFLOW` - Step-by-step workflow guides
- `TUTORIAL` - Practical examples & walkthroughs
- `ARCH` - Architecture & system design
- `HOOKS` - Hooks system documentation

Examples:
- `@DOC:GUIDE-CONCEPT-SPEC-001` for SPEC-First TDD guide
- `@DOC:GUIDE-WORKFLOW-INIT-001` for project initialization workflow
- `@DOC:GUIDE-TUTORIAL-TODOAPP-BACKEND-001` for Todo app backend example
- `@DOC:GUIDE-ARCH-AGENTS-001` for agents architecture guide
- `@DOC:GUIDE-HOOKS-OVERVIEW-001` for hooks system overview

**For SKILL Domain (5 files)**:
Pattern: `@DOC:SKILL-{TIER}-{SEQUENCE}`

- `@DOC:SKILL-OVERVIEW-001` for skills system overview
- `@DOC:SKILL-FOUNDATION-001` for foundation tier skills
- `@DOC:SKILL-ESSENTIALS-001` for essentials tier skills
- `@DOC:SKILL-DOMAIN-001` for domain tier skills
- `@DOC:SKILL-LANGUAGE-001` for language tier skills

**For ROOT Level (4 files)**:
Pattern: `@DOC:{CATEGORY}-{NAME}-{SEQUENCE}`

- `@DOC:INDEX-MAIN-001` for main docs index
- `@DOC:CONFIG-MULTILING-001` for multilingual menus (extends CONFIG domain)
- `@DOC:STATUS-OPTIM-001` for optimization status reports
- `@DOC:STATUS-SYNC-001` for sync completion reports

---

## 3. Priority Ranking & Effort Estimation

### Priority Matrix

| Priority | Category | Files | Lines | Est. Hours | Rationale |
|----------|----------|-------|-------|-----------|-----------|
| CRITICAL | Workflow Guides | 6 | 4,602 | 19 | Core user workflows, first-time visitor content |
| CRITICAL | Concept Guides | 5 | 2,385 | 17.5 | Foundational understanding of system |
| CRITICAL | Skills Documentation | 5 | 332 | 5.5 | System architecture knowledge |
| HIGH | Tutorial Examples | 7 | 4,905 | 26 | High-value practical learning |
| HIGH | Architecture Guides | 3 | 1,355 | 10 | System understanding foundation |
| HIGH | Quick Wins | 5 | 541 | 6.5 | Visibility & momentum building |
| MEDIUM | Miscellaneous | 2 | 660 | 3 | Housekeeping & completeness |

### Batch Execution Plan

**Batch 1: Quick Wins** (6.5 hours)
- agents/overview.md (55 lines) → @DOC:AGENT-OVERVIEW-001
- api-reference/agents.md (37 lines) → @DOC:API-AGENTS-001
- commands/cli.md (76 lines) → @DOC:CMD-CLI-INTRO-001
- docs/index.md (141 lines) → @DOC:INDEX-MAIN-001
- docs/status/sync-report.md (272 lines) → @DOC:STATUS-SYNC-001

**Batch 2: Skills Tier System** (5.5 hours)
- skills/overview.md (90 lines) → @DOC:SKILL-OVERVIEW-001
- skills/foundation-tier.md (75 lines) → @DOC:SKILL-FOUNDATION-001
- skills/essentials-tier.md (63 lines) → @DOC:SKILL-ESSENTIALS-001
- skills/domain-tier.md (36 lines) → @DOC:SKILL-DOMAIN-001
- skills/language-tier.md (68 lines) → @DOC:SKILL-LANGUAGE-001

**Batch 3: Architecture & Hooks** (10 hours)
- guides/architecture.md (124 lines) → @DOC:GUIDE-ARCH-ARCHITECTURE-001
- guides/agents/overview.md (705 lines) → @DOC:GUIDE-ARCH-AGENTS-001
- guides/hooks/overview.md (526 lines) → @DOC:GUIDE-HOOKS-OVERVIEW-001

**Batch 4: Core Concepts** (17.5 hours)
- guides/concepts/spec-first-tdd.md (773 lines) → @DOC:GUIDE-CONCEPT-SPEC-001
- guides/concepts/tag-system.md (646 lines) → @DOC:GUIDE-CONCEPT-TAG-001
- guides/concepts/trust-principles.md (554 lines) → @DOC:GUIDE-CONCEPT-TRUST-001
- guides/concepts/ears-guide.md (314 lines) → @DOC:GUIDE-CONCEPT-EARS-001
- guides/concepts/skills-revolution.md (98 lines) → @DOC:GUIDE-CONCEPT-SKILLS-001

**Batch 5: Workflow Guides** (19 hours)
- guides/workflow/overview.md (111 lines) → @DOC:GUIDE-WORKFLOW-OVERVIEW-001
- guides/workflow/0-project.md (2544 lines) → @DOC:GUIDE-WORKFLOW-INIT-001
- guides/workflow/1-plan.md (122 lines) → @DOC:GUIDE-WORKFLOW-PLAN-001
- guides/workflow/2-run.md (108 lines) → @DOC:GUIDE-WORKFLOW-RUN-001
- guides/workflow/3-sync.md (103 lines) → @DOC:GUIDE-WORKFLOW-SYNC-001
- guides/workflow/9-update.md (1715 lines) → @DOC:GUIDE-WORKFLOW-UPDATE-001

**Batch 6: Tutorial Examples** (26 hours)
- guides/examples/index.md (114 lines) → @DOC:GUIDE-TUTORIAL-INDEX-001
- guides/examples/todo-app/index.md (361 lines) → @DOC:GUIDE-TUTORIAL-TODOAPP-001
- guides/examples/todo-app/01-project-init.md (731 lines) → @DOC:GUIDE-TUTORIAL-TODOAPP-INIT-001
- guides/examples/todo-app/02-spec-writing.md (837 lines) → @DOC:GUIDE-TUTORIAL-TODOAPP-SPEC-001
- guides/examples/todo-app/03-backend-tdd.md (988 lines) → @DOC:GUIDE-TUTORIAL-TODOAPP-BACKEND-001
- guides/examples/todo-app/04-frontend-impl.md (871 lines) → @DOC:GUIDE-TUTORIAL-TODOAPP-FRONTEND-001
- guides/examples/todo-app/05-sync-deploy.md (1003 lines) → @DOC:GUIDE-TUTORIAL-TODOAPP-DEPLOY-001

**Batch 7: Miscellaneous** (3 hours)
- docs/MULTILINGUAL-MENUS.md (263 lines) → @DOC:CONFIG-MULTILING-001
- docs/skills-optimization-phase2-3-report.md (397 lines) → @DOC:STATUS-OPTIM-001

### Total Summary

| Item | Value |
|------|-------|
| Total Untagged Files | 33 |
| Total Lines to Review | ~14,780 |
| Estimated Effort | 87.5 hours |
| Recommended Timeline | 2-3 weeks (3-4 hours/day) |
| Expected Coverage After Completion | 100% (78/78 files tagged) |

---

## 4. Implementation Roadmap

### Phase 3 Execution Strategy

1. **Proof of Concept (Batch 1 & 2)**: 11.5 hours
   - Complete 10 quick/small files
   - Achieve 68% overall coverage (53/78 tagged)
   - Validate tagging approach with diverse file types

2. **Foundation & Architecture (Batch 3)**: 10 hours
   - Tag architecture-critical documentation
   - Establish domain understanding for more complex tagging
   - Achieve 82% coverage (64/78 tagged)

3. **Core Knowledge** (Batch 4): 17.5 hours
   - Tag conceptual frameworks (SPEC, TAG, TRUST, EARS)
   - Create reference material for users
   - Achieve 89% coverage (69/78 tagged)

4. **User Workflows** (Batch 5): 19 hours
   - Tag step-by-step workflow documentation
   - Ensure user path completeness
   - Achieve 95% coverage (74/78 tagged)

5. **Learning Resources** (Batch 6): 26 hours
   - Tag comprehensive tutorial walkthroughs
   - Provide complete end-to-end examples
   - Achieve 100% coverage (81/78 tagged after removing duplicate)

6. **Final Polish** (Batch 7): 3 hours
   - Complete remaining miscellaneous files
   - Validate all tag chain relationships
   - Final documentation review

---

## 5. Risk Assessment & Mitigation

### Identified Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Very Large Files (2544 lines) | May require hierarchical tagging | Plan chunking strategy early; validate with stakeholders |
| Cross-file Dependencies | Complex cross-referencing needed | Build dependency graph before Batch 4 |
| Content Gaps | Might discover missing documentation | Log separately; focus Phase 3 on tagging only |
| Terminology Inconsistency | Tag naming confusion | Establish glossary from existing tagged files |
| Scope Creep | Timeline overrun | Strict focus on tagging; no rewrites |

### Success Criteria
- 100% of 78 files have @DOC tags
- All tags follow established naming convention
- TAG chains validated (SPEC → DOC)
- No duplicate tags within domains
- Cross-reference index created

---

## 6. Other Documentation Ecosystem

### .moai/ Structure (NOT included in Phase 3)
```
.moai/
├── memory/         [12 files - Internal system knowledge]
├── docs/           [4 files - Analysis & guides]
├── analysis/       [1 file - Technical studies]
├── reports/        [45 files - Project tracking]
├── specs/          [40+ files - SPEC definitions]
├── project/        [3 files - Metadata]
└── social/         [1 file - Release notes]
```

**Note**: .moai/ files are internal to Alfred workflow. Not included in Phase 3 user documentation tagging.

### Root Documentation (NOT included in Phase 3)
```
/
├── README.md       [79,421 lines - Main user documentation]
├── README.ko.md    [Korean translation]
├── README.ja.md    [Japanese translation]
├── README.zh.md    [Chinese translation]
├── CLAUDE.md       [Project guidance for Alfred]
├── CHANGELOG.md    [Version history]
└── CONTRIBUTING.md [Contribution guidelines]
```

**Note**: Root-level files managed separately as project metadata. Only docs/ directory in scope for Phase 3.

---

## Conclusion

**Phase 3 Focus**: Achieve 100% @DOC tag coverage for user-facing documentation in docs/ directory.

**Key Achievements**:
1. Complete tags for 33 untagged files
2. Introduce 3 new domain categories (GUIDE, SKILL, STATUS)
3. Create foundation for automated documentation validation
4. Enable cross-reference navigation between docs and code

**Recommended Execution**: Start with Batch 1-2 (Quick Wins + Skills) for immediate 68% coverage and momentum, then proceed systematically through Batches 3-7.

**Timeline**: 2-3 weeks at 3-4 hours/day, or 1 week at full-time focus.

