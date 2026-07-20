---
description: "SPEC 파일 frontmatter canonical 12-field 스키마 — Single Source of Truth (SSOT)"
paths: "**/.moai/specs/**,internal/spec/**"
---

# SPEC Frontmatter Schema — SSOT

> **Single Source of Truth** for the canonical SPEC frontmatter schema.
> Enforcement: `internal/spec/lint.go` `FrontmatterSchemaRule`.
> Cross-referenced by: `.claude/skills/moai/workflows/plan.md` § Pre-Write Frontmatter Checklist.

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
draft → in-progress → implemented → completed
   ↑                            ↓
(planned, legacy-optional)  superseded | archived | rejected
```

Valid values: `draft`, `planned`, `in-progress`, `implemented`, `completed`, `superseded`, `archived`, `rejected`

> **`planned` is legacy-optional — NOT part of the active V3R6 flow.** The modern flow transitions `draft → in-progress` directly (manager-develop, on the first run-phase commit). `planned` is retained in the enum for backward compatibility with pre-V3R6 SPECs that recorded it, but it has **no active-flow owner** — no agent authors a `draft → planned` transition in the current lifecycle, and the Status Transition Ownership Matrix below deliberately omits a `draft → planned` row. Do NOT invent a new owner for it, and do NOT remove it from the enum (removal would break parsing of grandfathered SPECs that carry `status: planned`).

> **`completed → in-progress (amendment)` transition** — a SPEC with `status: completed` MAY transition back to `status: in-progress` for an in-place amendment. This reuses the existing `in-progress` status (NO new `amended` enum value is added). The amendment is declared via the `amendment_of:` optional frontmatter field (self-referential for in-place, or parent SPEC ID for successor) plus a HISTORY `## Amendments` sub-section recording the prior completed version, prior_completed_sha, rationale, and scope. See the `completed → in-progress (amendment)` row in the matrix below. During amendment, the `internal/spec/audit.go` completed-no-drift predicate does NOT fire (frontmatter is `in-progress`, not `completed`), so normal drift detection resumes — no Go change required.

## Status Transition Ownership Matrix

Per the canonical agent-responsibility realignment policy (DRI ownership at agent-artifact granularity per Anthropic Best Practice #7) and the agent catalog consolidation policy (8 retained agents). This matrix is the **schema-level SSOT** for which agent performs each canonical status transition. Owner columns reference only the 7 MoAI-custom retained agents (`manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`) plus orchestrator-direct entries; archived agent names (`manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher`, and the 6 `expert-*` agents) MUST NOT appear as owners — see `.claude/rules/moai/workflow/archived-agent-rejection.md` for migration guidance. Cross-referenced by the `## SPEC Artifact Ownership` body sections in `.claude/agents/moai/manager-{spec,develop,docs}.md`.

| Transition | Owning agent | Canonical commit subject pattern |
|------------|--------------|----------------------------------|
| `(none) → draft` | manager-spec | `feat(SPEC-{ID}): plan-phase artifacts ({tier}, {N} artifacts)` — `{N}` is the Tier artifact count (Tier S = 2, Tier M = 3, Tier L = 5 per § SPEC Complexity Tier); do NOT hardcode a fixed count |
| `draft → in-progress` | manager-develop (on M1 commit start) | `fix(SPEC-{ID}): M1 ...` or `feat(SPEC-{ID}): M1 ...` — first run-phase commit |
| `in-progress → implemented → completed` | manager-docs (on the single sync commit — the `completed` transition is merged into the sync commit, NOT a separate Mx chore commit) | `docs(SPEC-{ID}): sync-phase artifacts` or `chore(SPEC-{ID}): sync-phase artifacts` (this same sync commit carries the `completed` transition + the 3-phase close) |
| `* → superseded` | manager-spec (when authoring the new superseding SPEC) | `feat(SPEC-{NEW-ID}): supersedes SPEC-{OLD-ID}` |
| `* → archived` | manager-docs (administrative cleanup) | `chore(specs): archive SPEC-{ID}` |
| `* → rejected` | orchestrator decision, recorded by manager-docs | `chore(SPEC-{ID}): rejected per <rationale>` |
| `completed → in-progress (amendment)` | manager-spec (re-delegation per D-NEW-1 inline-fix pattern) | `feat(SPEC-{ID}): in-place amendment <rationale-summary>` — distinct from `(none) → draft` and `* → superseded`; the SPEC's HISTORY `## Amendments` sub-section MUST record the prior completed version + prior_completed_sha + rationale + scope |

> **3-phase close (plan→run→sync)** — the MoAI lifecycle is exactly three phases (`plan`, `run`, `sync`); MX Tag is a cross-cutting concern validated during sync, NOT a separate fourth phase. The `completed` status transition rides the sync commit (manager-docs owns it); there is no separate "Mx chore commit". The progress.md §E structure is 4 sections (§E.1 Plan / §E.2 Run Evidence / §E.3 Run Audit-Ready / §E.4 Sync Audit-Ready) — the former `§E.5 Mx-phase` section is retired (folded into §E.4).

## progress.md Section Map (canonical SSOT)

`progress.md` uses lettered top-level sections. **Section §E and its `§E.N` sub-sections are parser-load-bearing**: `internal/spec/era.go` `ClassifyEra()` classifies a SPEC's era by string-matching the literal heading tokens `§E.2`, `§E.3`, `§E.4`, `§E.5` and the literal field names `sync_commit_sha` / `mx_commit_sha` in `progress.md` content (see `hasProgressMarker` / `hasAnyProgressMarker` / `extractProgressField`). Renaming any `§E.N` heading or either field name would silently break era classification (a V3R6 SPEC would misclassify as V3R2-R4 or V3R5). The map below is the SSOT for section-letter allocation; it PRESERVES the parsed §E.* headings verbatim and assigns every other progress.md concern a distinct, non-colliding letter so no two concerns share `§E`.

| Section | Purpose | Parsed by era.go? | Owner / when written |
|---------|---------|-------------------|----------------------|
| `## §E.1 Plan-phase Audit-Ready Signal` | Plan-phase completion signal (`plan_status: audit-ready`, `plan_complete_at`) | No (only §E.2-§E.5 headings + the two SHA fields are matched) | manager-spec (plan-phase) |
| `## §E.2 Run-phase Evidence` | Run-phase evidence; the literal `§E.2` heading is the run-evidence START marker era.go detects | **YES** — literal `§E.2` heading | manager-develop (run-phase) |
| `## §E.3 Run-phase Audit-Ready Signal` | Run-phase completion signal | **YES** — literal `§E.3` heading | manager-develop (run-phase) |
| `## §E.4 Sync-phase Audit-Ready Signal` | Sync-phase close; carries the literal `sync_commit_sha:` field (populated by the single sync commit) | **YES** — literal `§E.4` heading + `sync_commit_sha` field | manager-docs (sync commit) |
| `## §E.5 Mx-phase` (RETIRED) | Legacy Mx-phase marker; folded into §E.4. era.go still matches it for the H-4-legacy migration-window dual predicate (§E.5 + `mx_commit_sha`) so pre-redesign SPECs classify as V3R6 | **YES** — literal `§E.5` heading + `mx_commit_sha` field (legacy only) | (retired — do NOT author new §E.5 sections; retained in the parser for backward-compat classification of legacy SPECs) |
| `## §F Phase 4 Mode Selection` | Orchestrator's Phase 4 mode-selection log (must preserve the `Mode Selection` token for the grep AC) | No | orchestrator (before first run-phase `Agent()` spawn) |
| `## §H Recursive Self-Diagnosis Log` | Phase 2 bounded recursive self-diagnosis loop record (DIAGNOSE-PATCH-VERIFY, mechanical failures) | No | manager-develop / orchestrator (run-phase) |
| `## §I Token Accounting` | Token-accounting baseline — sync-close per-SPEC token spend measurement (`tokens_spent` + attribution confidence qualifier) | No (only `§E.*` headings + the two SHA fields are matched) | token-accounting mechanism at sync-close (manager-docs invokes the writer) |

**Section-letter allocation rule.** New progress.md concerns MUST claim a fresh top-level letter (`§F`, `§G`, `§H`, `§I`, ...) — they MUST NOT overload `§E` or any `§E.N` sub-heading, because the `§E.*` namespace is reserved for the era.go-parsed lifecycle-phase structure. A concern that reuses `## §E — <something else>` collides with the parser's heading match and is prohibited.

> **§F disambiguation (progress.md vs SPEC-body §F).** The `§F` allocated here is a **progress.md** section (Phase 4 Mode Selection log). It is distinct from the SPEC-body "§F" that `plan.md` / `sync.md` skill workflows reference (a spec.md / plan.md body section, a different document). The two live in different files and do not collide; when citing "§F", name the file (`progress.md §F` vs `spec.md §F`).

### Close-subject full-ID mandate

Per the drift-detector close-subject convention, every close commit (the sync commit carrying the `implemented → completed` transition above) MUST name exactly one individual full SPEC-ID in its subject scope — e.g. `chore(SPEC-{DOMAIN}-{SUB}-001): … 3-phase close`. A **combined/abbreviated scope** that names only a shared prefix (e.g. `chore(SPEC-{DOMAIN}): … 3-phase close (SUB-A + SUB-B)`) is **prohibited**: the drift detector's exact-token SPEC-ID extraction cannot map an abbreviated prefix to its sibling SPECs, so combined-scope close subjects regenerate lifecycle drift false-positives. When closing N sibling SPECs together, emit N separate close commits, one per full SPEC-ID — combined/abbreviated scope is disallowed in close subjects.

> **D4 reconciliation note**: The close-subject convention above owns the close-infix matcher contract. The legacy `"4-phase close"` infix was amended to the canonical `"3-phase close"` in this prose, and the drift detector's close-infix matcher (`internal/spec/transitions.go` `closeInfixMatch`) has been extended to accept BOTH infixes — the legacy `"4-phase close"` is RETAINED in the matcher because historical close commits in git history carry it. A doc-only rename without the dual-infix matcher update was forbidden (it would silently break drift close-recognition for all future closes).

### Forbidden ownership crossings

- `manager-docs` MUST NOT modify `spec.md` / `plan.md` / `acceptance.md` body content (frontmatter `status:` + `updated:` updates across the full `in-progress → implemented → completed` sync-commit transition are allowed — the single sync commit carries the terminal `completed` transition, so manager-docs' allowed frontmatter scope reaches `completed`, NOT only `implemented`; ALL other body modifications are forbidden). When sync-phase reveals a need to modify SPEC body content, manager-docs MUST return a blocker report and the orchestrator re-delegates to manager-spec.
- `manager-develop` MUST NOT modify `spec.md` / `plan.md` / `acceptance.md` body content (frontmatter `status:` + `updated:` updates on the `draft → in-progress` transition are allowed; ALL other body modifications are forbidden). When run-phase reveals a need to modify SPEC body content, manager-develop MUST return a blocker report and the orchestrator re-delegates to manager-spec for the scope-doc update before re-delegating back.

> **SHA placeholder backfill exemption (D3)** — The forbidden crossings above bind `spec.md` / `plan.md` / `acceptance.md` body content. The `progress.md` §E.3 / §E.4 `sync_commit_sha` / `mx_commit_sha` SHA fields are a DISTINCT surface: a sync-phase (or legacy Mx-phase) commit cannot reference its own SHA — a commit does not know its own hash until after it lands — so the established pattern writes a `pending-backfill-*` placeholder in the phase's own commit and backfills the real SHA in a follow-up commit. This mechanical placeholder completion is NOT an ownership crossing; it is the self-referential-hazard workaround for a field that, by physics, cannot be populated within the same commit it describes. The ownership matrix therefore scopes its forbidden crossings to SPEC body content (spec/plan/acceptance) and does NOT bind `progress.md` SHA-field backfill performed by the phase-owning agent (manager-develop for §E.3, manager-docs for §E.4) in a subsequent commit.

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
| `amendment_of` | string | SPEC ID declaring this SPEC is an in-place amendment of a prior completed SPEC. Self-referential (value = own ID) for in-place amendment; parent SPEC ID for successor amendment. Paired with a HISTORY `## Amendments` sub-section recording prior completed version, prior_completed_sha, rationale, and scope. See Status Transition Ownership Matrix `completed → in-progress (amendment)` row. |
| `tier` | enum (S\|M\|L) | SPEC complexity Tier classification. Tier S = 2 files (spec.md + plan.md); Tier M = 3 files (spec.md + plan.md + acceptance.md); Tier L = 5 files (spec.md + plan.md + acceptance.md + design.md + research.md). When `tier:` is absent, the SPEC is treated as Tier L for backward compat. See `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier. |

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
- **REQ coverage**: covered by the schema-lint SPEC (owning-SPEC pointer retained in the footer)
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
created: YYYY-MM-DD
updated: YYYY-MM-DD
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

## Owning SPEC

- Schema enforcement (`FrontmatterSchemaRule`): the lifecycle-redesign SPEC
- Close-subject convention + 3-phase close infix matcher: the drift-detector convention SPEC
