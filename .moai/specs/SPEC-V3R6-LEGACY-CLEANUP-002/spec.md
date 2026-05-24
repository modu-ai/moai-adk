---
id: SPEC-V3R6-LEGACY-CLEANUP-002
title: "Template mirror cascade for v2.x agency keyword cleanup (LEGACY-CLEANUP-001 follow-up)"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template/templates"
lifecycle: spec-anchored
tags: "cleanup, legacy, template-mirror, sprint-2, docs"
tier: S
---

# SPEC-V3R6-LEGACY-CLEANUP-002 — Template Mirror Cascade for LEGACY-CLEANUP-001

## §1 Goal

Mirror the 5 surgical agency-keyword cleanup edits applied by SPEC-V3R6-LEGACY-CLEANUP-001 (commits ffa65ab15..19bc873ff) from the user-facing `.claude/` paths into the corresponding `internal/template/templates/.claude/` paths, so that future `moai init` runs deploy the corrected content to new user projects.

Without this SPEC, CLAUDE.local.md §2 [HARD] Template-First Rule is violated and permanent desync persists: every fresh `moai init` continues to deploy outdated v2.x agency-keyword content despite LEGACY-CLEANUP-001 having cleaned the orchestrator-side paths.

**Scope**: 5 source/target file pairs (one source file per agency-keyword-touched location, mirrored to the embedded template tree). All edits in LEGACY-CLEANUP-001 totaled +11/-11 lines across these 5 files (surgical keyword-replacement only).

## §2 EARS Requirements

### REQ-LCL2-001 — design constitution mirror

WHEN this SPEC's M1 runs, the system SHALL ensure `internal/template/templates/.claude/rules/moai/design/constitution.md` is byte-identical (SHA-256 match) to `.claude/rules/moai/design/constitution.md` at HEAD.

### REQ-LCL2-002 — brand-design skill mirror

WHEN this SPEC's M1 runs, the system SHALL ensure `internal/template/templates/.claude/skills/moai-domain-brand-design/SKILL.md` is byte-identical (SHA-256 match) to `.claude/skills/moai-domain-brand-design/SKILL.md` at HEAD.

### REQ-LCL2-003 — copywriting skill mirror

WHEN this SPEC's M1 runs, the system SHALL ensure `internal/template/templates/.claude/skills/moai-domain-copywriting/SKILL.md` is byte-identical (SHA-256 match) to `.claude/skills/moai-domain-copywriting/SKILL.md` at HEAD.

### REQ-LCL2-004 — gan-loop workflow skill mirror

WHEN this SPEC's M1 runs, the system SHALL ensure `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md` is byte-identical (SHA-256 match) to `.claude/skills/moai-workflow-gan-loop/SKILL.md` at HEAD.

### REQ-LCL2-005 — design workflow skill mirror

WHEN this SPEC's M1 runs, the system SHALL ensure `internal/template/templates/.claude/skills/moai/workflows/design.md` is byte-identical (SHA-256 match) to `.claude/skills/moai/workflows/design.md` at HEAD.

## §3 Acceptance Criteria

All ACs are binary verifiable via SHA-256 comparison. Each AC corresponds 1:1 to a REQ above.

### AC-LCL2-001 — constitution byte-identical

`shasum -a 256 .claude/rules/moai/design/constitution.md internal/template/templates/.claude/rules/moai/design/constitution.md` → both SHA-256 digests identical.

### AC-LCL2-002 — brand-design SKILL byte-identical

`shasum -a 256 .claude/skills/moai-domain-brand-design/SKILL.md internal/template/templates/.claude/skills/moai-domain-brand-design/SKILL.md` → both SHA-256 digests identical.

### AC-LCL2-003 — copywriting SKILL byte-identical

`shasum -a 256 .claude/skills/moai-domain-copywriting/SKILL.md internal/template/templates/.claude/skills/moai-domain-copywriting/SKILL.md` → both SHA-256 digests identical.

### AC-LCL2-004 — gan-loop SKILL byte-identical

`shasum -a 256 .claude/skills/moai-workflow-gan-loop/SKILL.md internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md` → both SHA-256 digests identical.

### AC-LCL2-005 — design workflow byte-identical

`shasum -a 256 .claude/skills/moai/workflows/design.md internal/template/templates/.claude/skills/moai/workflows/design.md` → both SHA-256 digests identical.

## §A Sources of Change Traceability

This SPEC is a pure follow-up cascade; no new content originates here. Every byte mirrored derives from LEGACY-CLEANUP-001 commits in the range `ffa65ab15..19bc873ff` (inclusive).

### §A.1 LEGACY-CLEANUP-001 Commit Range

| Commit | Milestone | Files touched |
|--------|-----------|---------------|
| `ffa65ab15` | M1 | backup + skills/rule (includes 5 source files in scope) |
| `e517d59e9` | M2 | docs-site ko + en (out of scope for this SPEC) |
| `42bc8024d` | M3 | docs-site ja + zh (out of scope) |
| `ccd1fa9cf` | M4 | root markdown + final verification (out of scope) |
| `aea0cf7b9` | sync | frontmatter status + version 0.1.0→0.2.0 (out of scope) |
| `19bc873ff` | sync | CHANGELOG + AC correction (out of scope) |

### §A.2 5 Source/Target File Pairs

| # | Source path (already cleaned in LCL-001) | Target path (this SPEC mirrors to) |
|---|-------------------------------------------|------------------------------------|
| 1 | `.claude/rules/moai/design/constitution.md` | `internal/template/templates/.claude/rules/moai/design/constitution.md` |
| 2 | `.claude/skills/moai-domain-brand-design/SKILL.md` | `internal/template/templates/.claude/skills/moai-domain-brand-design/SKILL.md` |
| 3 | `.claude/skills/moai-domain-copywriting/SKILL.md` | `internal/template/templates/.claude/skills/moai-domain-copywriting/SKILL.md` |
| 4 | `.claude/skills/moai-workflow-gan-loop/SKILL.md` | `internal/template/templates/.claude/skills/moai-workflow-gan-loop/SKILL.md` |
| 5 | `.claude/skills/moai/workflows/design.md` | `internal/template/templates/.claude/skills/moai/workflows/design.md` |

### §A.3 Pre-Plan SHA-256 Divergence Snapshot (HEAD = 2509de913)

All 5 pairs CURRENTLY DIFFER (confirmed at plan time):

| # | Source SHA-256 (HEAD) | Target SHA-256 (HEAD) | Status |
|---|------------------------|------------------------|--------|
| 1 | `aa45b0255eed2f5278087078272b57fab991d000db7859631372866ca45d0670` | `0519f5e4e1cedd83538e4f3eeab5e4c3d5b4c2a01c1e36d3fd1366a4005eefde` | DIFFER |
| 2 | `5185d309df1fd0ef9265b87f9952207ed297428601446c1397360c1386577421` | `02fd775882a1c7975435a3f7345207a9905e6bc4077dcf09c4c7e0c8c3f47c2b` | DIFFER |
| 3 | `e00607b138afe71fab38dfb43026ea1d3794c51d946cc266b0711eb2d28ec1af` | `dc69e8164fb514ad6a80be90d20e8026c87aba498ca53fa7095e8906f60fbbc8` | DIFFER |
| 4 | `e57bff5ccc8cf403fb96d9e0967e6f93c7d9f56146765a245909f3d1d3a2fde5` | `c10f0565a46e770633a2adcf7de61f12a2cb97d8cf59d354bcb19558571da288` | DIFFER |
| 5 | `fadac136e27bebf0f591044a7dcb88e288a01ea470b003a2c03027b18779c32c` | `6ed27ea61739841f8caef7fa1907cf44cb3f3a5216e089b1c648892f0287e940` | DIFFER |

After M1 succeeds, all 5 target SHA-256 digests MUST equal the corresponding source SHA-256 digests (which are stable at HEAD since LEGACY-CLEANUP-001 closed and no further source edits planned).

### §A.4 Cumulative Diff Volume (from LEGACY-CLEANUP-001 commit range)

`git diff ffa65ab15~1..19bc873ff --stat` for these 5 paths:

```
.claude/rules/moai/design/constitution.md        | 4 ++--
.claude/skills/moai-domain-brand-design/SKILL.md | 4 ++--
.claude/skills/moai-domain-copywriting/SKILL.md  | 4 ++--
.claude/skills/moai-workflow-gan-loop/SKILL.md   | 4 ++--
.claude/skills/moai/workflows/design.md          | 6 +++---
5 files changed, 11 insertions(+), 11 deletions(-)
```

Mirror operation is a verbatim file-copy (`cp -p` semantics or equivalent); no manual edit needed since the source already contains the corrected content.

### §A.5 Exclusions (out of scope for this SPEC)

Mirror-target files NOT touched by LEGACY-CLEANUP-001 source edits — they remain unchanged in this SPEC. Specifically: docs-site, root markdown, CHANGELOG, README, and other 26+ template files unaffected by the 5-source-file scope.

- Docs-site locale content (`docs-site/content/{en,ko,ja,zh}/...`) — handled by LEGACY-CLEANUP-001 M2/M3 directly on user-facing paths; no template mirror exists for docs-site content.
- Root markdown files (`README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, `CHANGELOG.md`) — not present under `internal/template/templates/` at the same relative path, so no mirror cascade applies.
- The 26+ other template files under `internal/template/templates/` whose source counterparts contained no `agency` keyword in LEGACY-CLEANUP-001 — preserved byte-identical.
