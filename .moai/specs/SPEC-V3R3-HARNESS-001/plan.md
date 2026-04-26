---
id: SPEC-V3R3-HARNESS-001
artifact: plan
version: "0.1.0"
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
---

# SPEC-V3R3-HARNESS-001 — Implementation Plan

## 1. Overview

This plan describes how to deliver the meta-harness skill (`moai-meta-harness`), remove 16 static skills (BC-V3R3-007), establish the `moai-*` vs `my-harness-*` namespace separation, and extend the `moai update` migrator. Work is decomposed into 5 milestones executable in priority order. No time estimates are provided per `.claude/rules/moai/core/agent-common-protocol.md` Time Estimation rule.

## 2. Milestones (Priority Order)

### Milestone M1 (Priority Critical) — Meta-Harness Skill Authoring

**Goal**: Create `moai-meta-harness` skill with revfactory/harness 7-Phase workflow.

Deliverables:
- `internal/template/templates/.claude/skills/moai-meta-harness/SKILL.md` (Template-First source of truth)
- `.claude/skills/moai-meta-harness/SKILL.md` (mirror, post `make build`)
- Apache 2.0 attribution block at top of SKILL.md
- 7-Phase workflow adapted to MoAI agent vocabulary (manager-spec, manager-tdd, manager-ddd, expert-*, evaluator-active)
- Frontmatter declaring `triggers.phases: ["plan", "run", "sync"]` and `triggers.keywords: ["harness", "project-init", "meta-skill"]`

Exit criteria:
- SKILL.md passes frontmatter validation (skill-authoring.md schema)
- 7-Phase narrative cross-references existing MoAI agents (no new agent introduction)
- Attribution block visible above the H1 title

### Milestone M2 (Priority High) — 16 Static Skills Removal

**Goal**: Remove the 16 enumerated skills from both source-of-truth and local mirror.

Deliverables:
- 16 skill directories removed from `internal/template/templates/.claude/skills/` (verbatim §3 list)
- 16 skill directories removed from `.claude/skills/` (post `make build`)
- Build-time enforcement test: `internal/template/skills_removal_test.go` fails if any removed skill reappears
- Git history preserves the deleted skill content (commit history, not working tree)

Exit criteria:
- `ls .claude/skills/moai-* | wc -l` returns exactly 23
- `internal/template/skills_removal_test.go` passes
- No accidental removal of `moai-foundation-*`, `moai-workflow-{plan,run,sync,project,fix,loop,review,clean,coverage,e2e,thinking,spec,docs,quality}`, `moai-ref-*`, `moai-design-*`, or `moai-domain-{copywriting,brand-design}`

### Milestone M3 (Priority High) — Namespace Separation Enforcement

**Goal**: Enforce `moai-*` vs `my-harness-*` namespace boundary at build time and at `moai doctor` runtime.

Deliverables:
- `moai doctor` extension: detects unauthorized `moai-*` skills outside the static core (i.e., not in the 23-skill allowlist)
- `moai doctor` extension: detects skills under `.claude/skills/my-harness-*/` (informational, confirms user customization is detected)
- `moai update` migrator: skips any directory matching `.claude/skills/my-harness-*/` entirely
- `moai update` migrator: skips `.claude/agents/my-harness/`

Exit criteria:
- `moai doctor` warns when an unknown `moai-*` skill exists
- `moai update` end-to-end test confirms `.claude/skills/my-harness-test/` is preserved untouched
- `internal/template/templates/.claude/skills/` contains no `my-harness-*` entries

### Milestone M4 (Priority High) — moai update Migrator + Archive

**Goal**: Implement automatic migration of removed skills to `.moai/archive/skills/v2.16/`.

Deliverables:
- `internal/cli/update/archive.go` — archive function copying full skill directory tree
- `internal/cli/update/archive_test.go` — table-driven tests with 16 skills × {present, absent} matrix
- `moai migrate restore-skill <skill-id>` subcommand — restores from `.moai/archive/skills/v2.16/<id>/`
- One-line status print per archived skill + final summary count
- Idempotency: running `moai update` twice does not double-archive

Exit criteria:
- Archive directory layout matches REQ-HARNESS-007: full skill content preserved (frontmatter, body, modules, examples, reference)
- `moai migrate restore-skill <id>` reproduces the original `.claude/skills/<id>/` directory byte-for-byte (excluding runtime artifacts like `.moai/cache/`)
- `moai update` test against fixture project containing all 16 skills produces 16 archive entries + meta-harness installation

### Milestone M5 (Priority Medium) — Migration Guide + Release Documentation

**Goal**: Author user-facing migration documentation and release notes.

Deliverables:
- `.moai/release/MIGRATION-v2.17.0.md` — full migration guide with 16-skill list, automatic steps, manual fallback, restore syntax, deprecation timeline
- `CHANGELOG.md` — v2.17.0 entry announcing BC-V3R3-007
- `.moai/release/RELEASE-NOTES-v2.17.0.md` — release notes with prominent BC-V3R3-007 callout
- `README.md` / `README.ko.md` — version line bump (v2.17.0)
- `.moai/config/sections/system.yaml` and `internal/template/templates/.moai/config/config.yaml` — version bump

Exit criteria:
- Migration guide reviewed against REQ-HARNESS-006 checklist
- All 4 release artifacts cross-link consistently
- Apache 2.0 attribution cited in CHANGELOG and migration guide (not just SKILL.md)

## 3. Technical Approach

### 3.1 7-Phase Workflow Adaptation Table

| revfactory/harness Phase | MoAI Equivalent | Owning Agent |
|-------------------------|-----------------|--------------|
| 1. Discovery | Socratic interview (16 questions) | manager-spec (driven by `/moai project` Phase 5+) |
| 2. Analysis | Codebase + brand context analysis | manager-spec + manager-strategy |
| 3. Synthesis | SPEC document with EARS format | manager-spec |
| 4. Skeleton | `.moai/harness/main.md` + extension files | meta-harness skill (this SPEC) |
| 5. Customization | `.claude/agents/my-harness/*.md` + `.claude/skills/my-harness-*/SKILL.md` generation | meta-harness skill (this SPEC) |
| 6. Evaluation | Sprint Contract scoring | evaluator-active |
| 7. Iteration | Self-learning loop | LEARNING-001 (separate SPEC) |

The meta-harness skill body documents all 7 phases but only owns Phases 4-5. Phase 1-3 are owned by `/moai project` (PROJECT-HARNESS-001). Phase 6 is owned by `evaluator-active`. Phase 7 is owned by `moai-harness-learner` (LEARNING-001).

### 3.2 Apache 2.0 Attribution Pattern

Top of `moai-meta-harness/SKILL.md`:

```
<!-- ATTRIBUTION
Original work: revfactory/harness (https://github.com/revfactory/harness)
License: Apache License 2.0
Adaptations: 7-Phase workflow integrated with MoAI agent ecosystem (manager-*, expert-*, evaluator-active)
NOTICE: This file contains modifications. See SPEC-V3R3-HARNESS-001 for derivation history.
-->
```

This satisfies Apache 2.0 §4(c) "retain ... attribution notices" and §4(b) "carry prominent notices stating that You changed the files".

### 3.3 Archive Layout

```
.moai/archive/skills/v2.16/
├── moai-domain-backend/
│   ├── SKILL.md
│   ├── modules/
│   ├── examples.md
│   └── reference.md
├── moai-domain-frontend/
│   └── ...
└── (14 more)
```

`moai migrate restore-skill <id>` copies `.moai/archive/skills/v2.16/<id>/` → `.claude/skills/<id>/`. The archive is read-only after first write; second `moai update` is idempotent.

### 3.4 Build-Time Enforcement Test

`internal/template/skills_removal_test.go` (NEW):

```go
func TestRemovedSkillsNotPresent(t *testing.T) {
    removed := []string{
        "moai-domain-backend", "moai-domain-frontend", "moai-domain-database",
        "moai-domain-db-docs", "moai-domain-mobile",
        "moai-framework-electron",
        "moai-library-shadcn", "moai-library-mermaid", "moai-library-nextra",
        "moai-tool-ast-grep",
        "moai-platform-auth", "moai-platform-deployment", "moai-platform-chrome-extension",
        "moai-workflow-research", "moai-workflow-pencil-integration",
        "moai-formats-data",
    }
    for _, id := range removed {
        path := filepath.Join("templates", ".claude", "skills", id)
        if _, err := os.Stat(path); !os.IsNotExist(err) {
            t.Errorf("removed skill still present: %s", id)
        }
    }
}
```

This locks the §3 removal list to a CI-enforced contract.

### 3.5 Doctor Allowlist

`internal/cli/doctor/skills.go` extension: hardcode the 23-skill allowlist. Skills outside the allowlist with `moai-` prefix → warning. Skills with `my-harness-` prefix → informational (no warning).

```go
var staticCoreAllowlist = []string{
    // foundation (4)
    "moai-foundation-cc", "moai-foundation-context", "moai-foundation-core", "moai-foundation-thinking",
    // workflow (8 — research and pencil-integration removed)
    "moai-workflow-spec", "moai-workflow-project", "moai-workflow-testing", "moai-workflow-templates",
    "moai-workflow-design-import", "moai-workflow-gan-loop", "moai-workflow-jit-docs", "moai-workflow-thinking",
    // ref (5)
    "moai-ref-anthropic", "moai-ref-go", "moai-ref-python", "moai-ref-typescript", "moai-ref-claude-code",
    // design (3 — pencil-integration removed)
    "moai-design-system", "moai-design-tokens", "moai-design-evaluator",
    // FROZEN domain (2)
    "moai-domain-copywriting", "moai-domain-brand-design",
    // meta (1 NEW)
    "moai-meta-harness",
}
```

(The exact 22 base skill names will be verified against the live tree before final SKILL.md authoring; the list above is illustrative.)

## 4. Risks (Plan-Level)

| Risk | Mitigation |
|------|------------|
| Migrator deletes user customization by accident | REQ-HARNESS-008 enforced by test: `internal/cli/update/preserve_my_harness_test.go` |
| Apache 2.0 attribution missing or incorrect | M1 exit criteria + manual review by manager-docs |
| Archive corrupts on edge cases (symlinks, permissions) | M4 test matrix covers `os.MkdirAll` + `os.CopyFS` semantics; deterministic byte-for-byte comparison |
| Doctor allowlist drifts from actual skill set | Build-time test compares allowlist to filesystem snapshot |
| Migration guide too dense for users | Worked example (one of the 16 skills) walked through end-to-end |

## 5. Quality Gates

- LSP gates per `.moai/config/sections/quality.yaml`: zero errors, zero type errors, zero lint errors
- TRUST 5: 85%+ coverage on new Go code (`internal/cli/update/archive*.go`)
- Conventional Commits with Korean body
- Template-First verification: every `.claude/skills/moai-meta-harness/` file mirrored under `internal/template/templates/`
- BC-V3R3-007 announcement present in CHANGELOG.md and RELEASE-NOTES-v2.17.0.md before tag

## 6. Plan Audit Checklist (for plan-auditor)

- [ ] All 11 REQs (REQ-HARNESS-001 through 010, plus implicit REQ-HARNESS-010) covered by an AC
- [ ] EARS keywords used in every REQ ("WHEN", "IF", "shall", "shall not")
- [ ] §3 16-skill list verbatim from handoff §4.2 / §6
- [ ] Frontmatter has all 9 required fields per `.claude/agents/moai/manager-spec.md` schema
- [ ] FROZEN zone constraints respected (no modification of design constitution §2)
- [ ] Apache 2.0 attribution required by REQ-HARNESS-001
- [ ] No file under `.claude/agents/moai/` modified
- [ ] depends_on: [SPEC-V3R3-HARNESS-LEARNING-001] declared
- [ ] breaking: true + bc_id: [BC-V3R3-007] declared
