---
id: SPEC-V3R3-CMD-CLEANUP-001
version: "1.0.0"
status: implemented
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
priority: P0
labels: [commands, skills, cleanup, security, template-first]
issue_number: null
breaking: false
bc_id: []
lifecycle: spec-anchored
related_specs: [SPEC-THIN-CMDS-001, SPEC-V3R3-DEF-001, SPEC-V3R3-DEF-007]
released_in: v2.15.0
---

# SPEC-V3R3-CMD-CLEANUP-001 — Commands Cleanup: gate 추가, security 흡수, context 제거

## HISTORY

- 2026-04-26: Initial draft. v3r3 extreme aggressive handoff §3.1 추출. Three-axis cleanup of `/moai` command surface — add missing `/moai gate` command file, strengthen security workflows in review/sync skills, remove unused `context` skill.

---

## 1. Background and Motivation

The current `/moai` command surface has three drift points after several SPEC waves:

1. `gate` skill exists at `.claude/skills/moai/workflows/gate.md` (5,668 bytes) and is referenced by SKILL.md routing, but the `/moai gate` slash command file is **missing** from `.claude/commands/moai/`. Users cannot invoke it directly — they must rely on automatic invocation from `run` Phase 2.75 / `sync` Phase 0.0.1.
2. `security` skill exists at `.claude/skills/moai/workflows/security.md` (8,656 bytes) and is referenced by SKILL.md routing entry `security`, but per the v3r3 plan the security capability should remain a **callable skill only** — no standalone `/moai security` slash command should be added. Instead, `review.md` Phase 4 and `sync.md` Phase 0.55 should explicitly invoke the security capability with the depth currently in `security.md`.
3. `context` skill exists at `.claude/skills/moai/workflows/context.md` (5,841 bytes) and is referenced by SKILL.md routing entry `context`, but it has not been used in the Plan→Run→Sync workflow since git-based memory was superseded by `@MX` annotations and `auto-memory`. The skill is dead code.

These drift points cause three concrete problems:

- Users cannot run a standalone pre-commit quality check via `/moai gate` (must invoke via `/moai run` or `/moai sync`).
- `review.md` Phase 4 lacks the dependency vulnerability scan, secrets git-history scan, and data-isolation checks that `security.md` performs.
- `sync.md` Phase 0.55 audits only changed files, missing dependency-manifest drift (go.mod, package.json, requirements.txt, Cargo.toml).
- `context.md` skill loads ~5,800 bytes of unused content into the routing surface and confuses users searching for git-context capabilities.

## 2. Goals

1. Provide a callable `/moai gate` slash command backed by the existing `gate.md` skill (no skill changes).
2. Strengthen `review.md` Phase 4 and `sync.md` Phase 0.55 to inline the security depth currently centralized in `security.md`. The `security.md` skill remains as a callable workflow for `review`/`sync` but is **not** exposed as a top-level command.
3. Remove the `context.md` skill and all routing references to it.
4. Update SKILL.md routing table to reflect the new state.
5. Keep template (`internal/template/templates/`) and local `.claude/` synchronized via Template-First discipline.

## 3. Non-Goals (Exclusions — What NOT to Build)

- **No new `/moai security` slash command.** Security audit remains accessible only as a skill called from `review` and `sync`. Users requesting standalone audit invoke `/moai review` (which now includes full security depth) or directly mention "security audit" (routed by SKILL.md keyword detection to the `security` skill).
- **No changes to `gate.md` skill body.** Only the missing command file is added; the skill content is preserved verbatim.
- **No changes to `security.md` skill body.** Its content is referenced (not copied) by review.md Phase 4 and sync.md Phase 0.55.
- **No deprecation warning for users invoking `/moai context`.** The command does not exist as a slash file currently; only the skill body and SKILL.md routing entry exist. Removing both is sufficient.
- **No new agent definitions.** Existing `expert-security` is reused.
- **No changes to language-specific behavior.** All 16 supported languages must continue to receive equal treatment in `gate.md` (project-marker auto-detection is already in place).
- **No GitHub workflow / CI changes.** Quality gate enforcement at PR-level is out of scope.

## 4. EARS Requirements

### REQ-CMD-001 (Gate Command Availability — Event-Driven)

WHEN user invokes `/moai gate` from Claude Code, THE system SHALL execute lint, format, type-check, and test in parallel via the `gate` skill and return aggregate exit status (zero on all-pass, non-zero on any failure).

### REQ-CMD-002 (Review Security Enhancement — Event-Driven)

WHEN user invokes `/moai review`, THE `review` skill SHALL include in Phase 4 (security perspective) the following audit steps with the depth currently defined in `security.md`:

- (a) dependency vulnerability scan against the project's manifest file (go.mod / package.json / requirements.txt / Cargo.toml / pyproject.toml as applicable),
- (b) secrets scan across full git history (not just current working tree),
- (c) data-isolation checks (multi-tenant boundary, PII separation, shared-state leakage).

### REQ-CMD-003 (Sync Manifest Audit — Event-Driven)

WHEN user invokes `/moai sync`, THE `sync` skill SHALL audit in Phase 0.55 both (a) changed source files in the current SPEC and (b) all dependency-manifest files at project root, regardless of whether they changed in this SPEC.

### REQ-CMD-004 (Context Skill Removal — Ubiquitous)

THE system SHALL NOT contain `.claude/skills/moai/workflows/context.md` nor a routing entry for `context` (alias `ctx`, `memory`) in `.claude/skills/moai/SKILL.md`.

### REQ-CMD-005 (Template Synchronization — Event-Driven)

WHEN `moai update` runs against a user project, THE template deployment SHALL (a) deploy `gate.md` command file under `.claude/commands/moai/`, (b) deploy unchanged `gate.md` and `security.md` skills, and (c) NOT deploy `context.md` skill, and SHALL NOT delete user-customized files outside the protected-paths list.

### REQ-CMD-006 (Security as Skill-Only — Ubiquitous)

THE security workflow SHALL remain accessible only as a skill (loaded by `review` Phase 4 and `sync` Phase 0.55, plus keyword-routed direct invocation per SKILL.md §Priority 3). THE system SHALL NOT provide a `/moai security` slash command file under `.claude/commands/moai/`.

## 5. Acceptance Criteria (Given-When-Then)

### AC-001 (REQ-CMD-001) — Gate command callable

- **Given** the project has Go source files and a Makefile,
- **When** the user enters `/moai gate` in Claude Code,
- **Then** Claude Code resolves the command via `.claude/commands/moai/gate.md` (Thin Command Pattern), routes to `Skill("moai")` with argument `gate`, the `gate` skill executes lint/format/type-check/test in parallel, and a single aggregate exit code is reported.

### AC-002 (REQ-CMD-002) — Review Phase 4 includes security depth

- **Given** the `/moai review` skill body at `.claude/skills/moai/workflows/review.md`,
- **When** the file is read,
- **Then** Phase 4 contains explicit textual references to: (1) "dependency vulnerability scan" with manifest-file enumeration, (2) "secrets scan" qualified as "full git history" (not just working tree), (3) "data isolation" check covering multi-tenant boundary or PII separation.

### AC-003 (REQ-CMD-003) — Sync Phase 0.55 audits manifests

- **Given** the `/moai sync` skill body at `.claude/skills/moai/workflows/sync.md`,
- **When** the file is read,
- **Then** Phase 0.55 contains explicit instruction to audit the project's dependency-manifest files (go.mod, package.json, requirements.txt, Cargo.toml, pyproject.toml) IN ADDITION to changed source files of the current SPEC.

### AC-004 (REQ-CMD-004, REQ-CMD-005) — Context removed and tests pass

- **Given** the SPEC has been applied,
- **When** the test suite runs,
- **Then** (1) `.claude/skills/moai/workflows/context.md` does not exist in either local repo or `internal/template/templates/`, (2) `.claude/skills/moai/SKILL.md` contains no `context` routing entry under §Priority 1 nor `context` keyword routing under §Priority 3, (3) `internal/template/commands_audit_test.go` (TestCommandsThinPattern) passes against the full embedded template fileset.

### AC-005 (REQ-CMD-006) — Security as skill only

- **Given** the SPEC has been applied,
- **When** the file system is inspected,
- **Then** `.claude/commands/moai/security.md` does NOT exist in local repo or template, AND `.claude/skills/moai/workflows/security.md` exists in both local repo and template (unchanged byte-for-byte from current state), AND SKILL.md `security` routing entry remains intact under §Priority 1 (subcommand list) and §Priority 3 (keyword detection).

## 6. Constraints

- **[HARD] Template-First**: Every change is applied to `internal/template/templates/` first, then mirrored to local `.claude/`. After template edits, run `make build` to regenerate `internal/template/embedded.go`.
- **[HARD] Thin Command Pattern (SPEC-THIN-CMDS-001)**: `gate.md` command body MUST be under 20 LOC, contain `Skill("moai")` invocation, and have YAML frontmatter with `description`, `argument-hint`, `allowed-tools` (CSV string).
- **[HARD] 16-language Neutrality (CONST-V3R2-004)**: `gate.md` skill (already existing) auto-detects via project markers; no language-specific bias may be introduced.
- **[HARD] No breaking change for existing users**: `context` skill removal is non-breaking because no slash command file currently invokes it; only SKILL.md routing referenced it. Users typing `/moai context` would have hit the SKILL.md router which would now route via "Other" or default.
- **[HARD] Frontmatter canonical schema**: `gate.md` command file MUST follow Thin Command Pattern frontmatter (description, argument-hint, allowed-tools as CSV string).
- All edits in user's `code_comments` language (ko); skill/command instruction text in English (per coding-standards.md §Language Policy).

## 7. Dependencies

- **Upstream**: SPEC-THIN-CMDS-001 (Thin Command Pattern enforcement), SPEC-V3R3-DEF-007 (skills convention sweep — gate.md and security.md already pass schema).
- **Downstream**: None blocked.
- **Parallel**: None.

## 8. Open Questions

- **OQ-1 (resolved)**: Should `/moai security` exist as a separate top-level command? **Resolved**: NO. Security depth is inlined into `review` Phase 4 and `sync` Phase 0.55. Skill remains for direct keyword routing per SKILL.md §Priority 3. (See REQ-CMD-006.)
- **OQ-2 (resolved)**: Should `context` removal include a deprecation notice? **Resolved**: NO. No slash command file existed; only skill + routing entry. Silent removal is appropriate. (See Non-Goals.)

## 9. Traceability

REQ-CMD-001 → AC-001 → tasks W1-T1, W1-T2, W1-T3 (plan.md)
REQ-CMD-002 → AC-002 → tasks W2-T1 (plan.md)
REQ-CMD-003 → AC-003 → tasks W2-T2 (plan.md)
REQ-CMD-004 → AC-004 → tasks W3-T1, W3-T2, W3-T3 (plan.md)
REQ-CMD-005 → AC-004 → tasks W1-T2, W3-T2 (plan.md, deployment + test)
REQ-CMD-006 → AC-005 → tasks (verification only — no creation work) (plan.md)
