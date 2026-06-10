---
description: "SPEC 파일 frontmatter canonical 12-field 스키마 — Single Source of Truth (SSOT)"
paths: "**/*.md,.moai/specs/**/*.md"
---

# SPEC Frontmatter Schema — SSOT

> **Single Source of Truth** for the canonical SPEC frontmatter schema.
> Enforcement: `internal/spec/lint.go` `FrontmatterSchemaRule` (REQ-SPC-003-006).
> Cross-referenced by: `.claude/skills/moai/workflows/plan.md` § Pre-Write Frontmatter Checklist,
> `.claude/skills/moai/team/plan.md` § Pre-Write Frontmatter Checklist.

## Canonical 12 Required Fields

All SPEC documents (`spec.md`) MUST contain exactly these 12 fields in YAML frontmatter.
Missing any field or using a snake_case alias causes `FrontmatterInvalid` lint findings.

```yaml
---
id: SPEC-{DOMAIN}-{NUM}
title: "Human-readable title"
version: "X.Y.Z"
status: draft
created: YYYY-MM-DD
updated: YYYY-MM-DD
author: Author Name
priority: P1
phase: "vX.Y.Z target"
module: "path/to/module"
lifecycle: spec-anchored
tags: "tag1, tag2, tag3"
---
```

## Field Reference

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `id` | string | `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` | Domain-namespaced identifier |
| `title` | string | non-empty, quoted | Human-readable description |
| `version` | string | semver `X.Y.Z`, quoted | Start at `"0.1.0"` |
| `status` | enum | See Status Enum below | Lifecycle state |
| `created` | date | `YYYY-MM-DD` ISO format | Creation date |
| `updated` | date | `YYYY-MM-DD` ISO format | Last update date |
| `author` | string | non-empty | Author name |
| `priority` | enum | `P0`\|`P1`\|`P2`\|`P3` or `High`\|`Medium`\|`Low`\|`Critical` | Default `P1` |
| `phase` | string | non-empty, typically release target | e.g. `"v3.0.0"` |
| `module` | string | non-empty, path-like | Affected Go module or directory |
| `lifecycle` | enum | `spec-anchored`\|`spec-lite`\|`exploratory` | Default `spec-anchored` |
| `tags` | string | comma-separated, non-empty | Searchable labels |

## Status Enum (8 values)

```
draft → planned → in-progress → implemented → completed
                                         ↓
                               superseded | archived | rejected
```

Valid values: `draft`, `planned`, `in-progress`, `implemented`, `completed`, `superseded`, `archived`, `rejected`

## Status Transition Ownership Matrix

Per the canonical agent-responsibility realignment policy (DRI ownership at agent-artifact granularity per Anthropic Best Practice #7) and the agent catalog consolidation policy (8 retained agents). This matrix is the **schema-level SSOT** for which agent performs each canonical status transition. Owner columns reference only the 7 MoAI-custom retained agents (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`) plus orchestrator-direct entries; archived agent names (`manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, and the 6 `expert-*` agents) MUST NOT appear as owners — see `.claude/rules/moai/workflow/archived-agent-rejection.md` for migration guidance. Cross-referenced by the `## SPEC Artifact Ownership` body sections in `.claude/agents/moai/manager-{spec,develop,docs}.md`.

| Transition | Owning agent | Canonical commit subject pattern |
|------------|--------------|----------------------------------|
| `(none) → draft` | manager-spec | `feat(SPEC-{ID}): plan-phase artifacts ({tier} Section A-E, 4 artifacts)` |
| `draft → in-progress` | manager-develop (on M1 commit start) | `fix(SPEC-{ID}): M1 ...` or `feat(SPEC-{ID}): M1 ...` — first run-phase commit |
| `in-progress → implemented` | manager-docs (on sync commit) | `docs(SPEC-{ID}): sync-phase artifacts` or `chore(SPEC-{ID}): sync-phase artifacts` |
| `implemented → completed` | manager-docs OR orchestrator (on Mx chore commit) | `chore(SPEC-{ID}): Mx-phase audit-ready signal + 4-phase close` |
| `* → superseded` | manager-spec (when authoring the new superseding SPEC) | `feat(SPEC-{NEW-ID}): supersedes SPEC-{OLD-ID}` |
| `* → archived` | manager-docs (administrative cleanup) | `chore(specs): archive SPEC-{ID}` |
| `* → rejected` | orchestrator decision, recorded by manager-docs | `chore(SPEC-{ID}): rejected per <rationale>` |

### Close-subject full-ID mandate

Per the drift-detector close-subject convention, every close commit (the `implemented → completed` transition above) MUST name exactly one individual full SPEC-ID in its subject scope — e.g. `chore(SPEC-{DOMAIN}-{SUB}-001): … 4-phase close`. A **combined/abbreviated scope** that names only a shared prefix (e.g. `chore(SPEC-{DOMAIN}): … 4-phase close (SUB-A + SUB-B)`) is **prohibited**: the drift detector's exact-token SPEC-ID extraction cannot map an abbreviated prefix to its sibling SPECs, so combined-scope close subjects regenerate lifecycle drift false-positives. When closing N sibling SPECs together, emit N separate close commits, one per full SPEC-ID — combined/abbreviated scope is disallowed in close subjects.

### Forbidden ownership crossings

- `manager-docs` MUST NOT modify `spec.md` / `plan.md` / `acceptance.md` body content (frontmatter `status:` + `updated:` updates on the `in-progress → implemented` transition are allowed; ALL other body modifications are forbidden). When sync-phase reveals a need to modify SPEC body content, manager-docs MUST return a blocker report and the orchestrator re-delegates to manager-spec.
- `manager-develop` MUST NOT modify `spec.md` / `plan.md` / `acceptance.md` body content (frontmatter `status:` + `updated:` updates on the `draft → in-progress` transition are allowed; ALL other body modifications are forbidden). When run-phase reveals a need to modify SPEC body content, manager-develop MUST return a blocker report and the orchestrator re-delegates to manager-spec for the scope-doc update before re-delegating back.

### Forward-looking enforcement (optional defense-in-depth)

A future PostToolUse hook MAY validate at execution time that the agent performing a Write on a SPEC artifact body matches the expected owner per this matrix. This is OPTIONAL (deferred to a follow-up SPEC if desired per the agent-responsibility realignment policy). The primary intervention is the declarative ownership in the agent body sections + this schema matrix; hook-based enforcement is a complementary layer.

## Optional Fields

These fields may be included when needed but are NOT required by `FrontmatterSchemaRule`:

| Field | Type | Notes |
|-------|------|-------|
| `issue_number` | integer or null | GitHub issue number. Omit entirely when not tracking. |
| `depends_on` | list | SPEC IDs this SPEC depends on. Used by BODP signal A. |
| `lint.skip` | list | Lint rule codes to skip. Use only for documented debt. |
| `bc_id` | string | Backward-compatibility tracking ID. |

## Rejected Snake_Case Aliases

The YAML struct decoder in `internal/spec/lint.go` uses `yaml:"created"`, `yaml:"updated"`, `yaml:"tags"` tags.
Snake_case aliases are silently dropped by the decoder, causing empty-value `FrontmatterInvalid` findings.

| Do NOT use | Use instead |
|------------|-------------|
| `created_at:` | `created:` |
| `updated_at:` | `updated:` |
| `labels:` | `tags:` |
| `spec_id:` | `id:` |

## Lint Rule Implementation

`FrontmatterSchemaRule` in `internal/spec/lint.go`:

- **Rule code**: `FrontmatterInvalid`
- **Severity**: Warning
- **REQ coverage**: REQ-SPC-003-006
- **Check**: Iterates all 12 required fields; emits one finding per missing/empty field.
- **YAML binding**: `SPECFrontmatter` struct uses canonical field names (`created`, `updated`, `tags`).
  Snake_case aliases in the source YAML file are not recognized — they produce empty values.

See `internal/spec/lint.go` `FrontmatterSchemaRule.Check()` for the authoritative implementation.

## OwnershipTransitionRule Cross-Reference

The Status Transition Ownership Matrix above is enforced at lint-time by the `OwnershipTransitionRule` in `internal/spec/lint_ownership.go` (registered in `defaultRules()` of `internal/spec/lint.go`). The rule emits two finding codes:

- **`OwnershipTransitionInvalid`** (Warning severity): Emitted when a SPEC's git-log history shows a status transition performed by an agent whose commit subject prefix does NOT match the canonical owner for that transition. Example: `manager-docs` performing `draft → in-progress` (which the matrix above assigns to `manager-develop`) triggers a finding.
- **`OwnershipTransitionUnreachable`** (Info severity): Emitted when the rule cannot read git history for the SPEC file (non-git environment, fresh clone without history, or `git log --follow` error). Graceful observation — no panic, no error escalation.

Default subset (per the ownership-transition lint policy): the rule evaluates the two most common transitions by default (`draft → in-progress` and `in-progress → implemented`). Terminal states (`superseded`, `archived`, `rejected`) are exempted via the `terminalStatusEnum` shared with `StatusGitConsistencyRule`.

Configuration: severity can be promoted to Error under `--strict` mode (same as `StatusGitConsistencyRule`). Per-SPEC opt-out via `lint.skip: [OwnershipTransitionInvalid]` in optional frontmatter (see Optional Fields above).

Implementation files: `internal/spec/lint_ownership.go` (rule body) + `internal/spec/lint_ownership_test.go` (TDD coverage).

## Examples

### Correct (all 12 fields, canonical names)

```yaml
---
id: SPEC-AUTH-001
title: "OAuth2 Authentication"
version: "0.1.0"
status: draft
created: 2026-05-16
updated: 2026-05-16
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/auth"
lifecycle: spec-anchored
tags: "auth, oauth2, security"
---
```

### Wrong (snake_case aliases — produces 3 FrontmatterInvalid findings)

```yaml
---
id: SPEC-AUTH-001
title: "OAuth2 Authentication"
version: "0.1.0"
status: draft
created_at: 2026-05-16   # WRONG — use created:
updated_at: 2026-05-16   # WRONG — use updated:
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/auth"
lifecycle: spec-anchored
labels: [auth, oauth2]   # WRONG — use tags: "auth, oauth2"
---
```

## Version History

| Date | Author | Change |
|------|--------|--------|
| (initial) | maintainer | Initial creation — resolves dual-schema drift between plan.md (9-field) and lint.go (12-field) |
