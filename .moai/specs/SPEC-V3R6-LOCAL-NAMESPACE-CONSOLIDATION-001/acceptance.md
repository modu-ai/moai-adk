---
id: SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001
title: "Local Agent Namespace Consolidation — Acceptance Criteria"
version: "0.1.2"
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

# Acceptance Criteria — SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001

## A. AC Index

This SPEC has 12 acceptance criteria (AC-LNC-001 through AC-LNC-012). All MUST pass for the SPEC to transition `in-progress → implemented` at sync-phase. Each AC has an independently verifiable command sequence specified in §B.

| AC ID | Title | Coverage | Severity |
|-------|-------|----------|----------|
| AC-LNC-001 | `.claude/agents/local/` directory exists with both specialist agents | REQ-LNC-001, REQ-LNC-008 | MUST |
| AC-LNC-002 | Thin Command Pattern preserved on both 97 and 98 wrappers | REQ-LNC-002, REQ-LNC-013 | MUST |
| AC-LNC-003 | `release-update-specialist.md` agent body created with full migrated content | REQ-LNC-005, REQ-LNC-008 | MUST |
| AC-LNC-004 | `github-specialist.md` agent body created with full migrated content | REQ-LNC-006 | MUST |
| AC-LNC-005 | Predecessor dev-only skill files removed | REQ-LNC-001 (cleanup obligation) | MUST |
| AC-LNC-006 | `.claude/rules/moai/development/agent-authoring.md` namespace contract updated + `skill-authoring.md` deprecation entries | REQ-LNC-011 | MUST |
| AC-LNC-007 | Template surface contains zero `CLAUDE.local.md` references | REQ-LNC-003, REQ-LNC-007 | MUST |
| AC-LNC-008 | `.moai/docs/dev-only-commands-isolation.md` verification checklist updated | REQ-LNC-001 (operational discipline) | MUST |
| AC-LNC-009 | `internal/template/templates/.claude/agents/local/` does NOT exist (REQ-LNC-012 negative test) | REQ-LNC-012 | MUST |
| AC-LNC-010 | `.moai/docs/generic-patterns-guide.md` exists in both local and template with 4 sections | REQ-LNC-004, REQ-LNC-010 | MUST |
| AC-LNC-011 | Full project `go test ./...` passes (commands_audit_test.go non-regression) | REQ-LNC-002, REQ-LNC-013 | MUST |
| AC-LNC-012 | REQ-LNC-009 deferred-verification traceability marker (binds orphan REQ to acceptance for MP-3 compliance; substantive verification deferred to follow-up SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001) | REQ-LNC-009 | SOFT (deferred) |

## B. Per-AC Verification Commands

Each verification command MUST be run from the project root `/Users/goos/MoAI/moai-adk-go/`. Expected output is documented per AC.

### AC-LNC-001 — Local agent namespace directory + 2 agent files exist

**Given** the migration is complete,
**When** the orchestrator executes the verification command,
**Then** the listing shows both specialist agent files present with non-zero size.

```bash
ls -la .claude/agents/local/release-update-specialist.md .claude/agents/local/github-specialist.md
```

Expected: Two file entries, each with size > 1000 bytes (file body is non-trivial migration of phase workflow). Exit code: 0.

### AC-LNC-002 — Thin Command Pattern preserved (97/98 wrappers ≤ 20 LOC body)

**Given** the thin command rewiring is complete,
**When** the orchestrator counts lines in both wrappers,
**Then** each file is ≤ 20 lines total (frontmatter ~7 lines + body 1-3 lines + trailing newline).

```bash
wc -l .claude/commands/97-release-update.md .claude/commands/98-github.md
```

Expected: Both line counts ≤ 20. The body line (last non-empty line) of 97-release-update.md MUST contain the literal string `release-update-specialist`. The body line of 98-github.md MUST contain the literal string `github-specialist`. Verify via:

```bash
grep -E "release-update-specialist|github-specialist" .claude/commands/97-release-update.md .claude/commands/98-github.md
```

Expected: 2 matches (one per file). Exit code: 0.

### AC-LNC-003 — release-update-specialist body contains all 9 phases (Phase 0 through Phase 8)

**Given** M2 migration is complete,
**When** the orchestrator greps for canonical phase headers,
**Then** all 9 phase headers (Phase 0 through Phase 8 — total 9 entries because the sequence is inclusive on both endpoints) are present.

```bash
grep -cE "^### Phase [0-8]" .claude/agents/local/release-update-specialist.md
```

Expected: Count = 9 (Phase 0 + Phase 1 + Phase 2 + Phase 3 + Phase 4 + Phase 5 + Phase 6 + Phase 7 + Phase 8). Verify Anti-Patterns section also migrated:

```bash
grep -c "^## Anti-Patterns" .claude/agents/local/release-update-specialist.md
```

Expected: 1.

### AC-LNC-004 — github-specialist body contains migrated workflow content

**Given** M2 migration is complete,
**When** the orchestrator greps for the GitHub workflow scope markers,
**Then** the body shows the expected migrated structure.

```bash
grep -cE "^## (Purpose|Activation|Phase|Anti-Pattern|Reference|Output|Verification|Agent)" .claude/agents/local/github-specialist.md
```

Expected: Count ≥ 10 (matches the rich migrated section structure from `.claude/skills/moai/workflows/github.md` — accounts for Purpose & Scope + Activation + Phase Sequence + Agent Delegation Map + Output Artifacts + Verification Gate + Anti-Patterns + References + sub-sections).

### AC-LNC-005 — Predecessor dev-only skill files removed

**Given** M3 removal is complete,
**When** the orchestrator attempts to stat the predecessor skill files,
**Then** both files are absent.

```bash
ls -la .claude/skills/moai/workflows/release-update.md .claude/skills/moai/workflows/github.md 2>&1
```

Expected output contains `No such file or directory` for both paths. Exit code: non-zero (ls reports missing files).

### AC-LNC-006 — agent-authoring.md + skill-authoring.md namespace contract updated

**Given** M1 contract update is complete,
**When** the orchestrator greps for the new `local/` row in the namespace table AND for deprecation markers in the skills namespace policy table,
**Then** entries are present at both SSOT locations (agent-authoring.md table + skill-authoring.md deprecation entries) and in both local and template copies.

Verification 1 — agent-authoring.md namespace table (verifies REQ-LNC-011 first clause):

```bash
grep -c "\.claude/agents/local/" .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md
```

Expected: Each file shows at least 2 matches (table row + at least one body reference). Both files must report ≥ 2.

Verification 2 — skill-authoring.md deprecation entries (verifies REQ-LNC-011 second clause):

```bash
grep -cE "97-release-update|98-github" .claude/rules/moai/development/skill-authoring.md internal/template/templates/.claude/rules/moai/development/skill-authoring.md
```

Expected: Each file shows at least 2 matches (one per deprecated slot). Both files must report ≥ 2.

### AC-LNC-007 — Template surface zero `CLAUDE.local.md` references

**Given** M4 leak elimination is complete,
**When** the orchestrator greps the entire template surface,
**Then** no matches are returned.

```bash
grep -rln "CLAUDE.local.md\|CLAUDE\.local" internal/template/templates/
```

Expected: stdout empty; exit code in {0, 1}. Rationale: `grep -E` returns 1 on no matches (canonical "clean" result); a `0` exit MAY occur under robust shell variants where upstream pipefail or unusual locale settings short-circuit the no-match path. Exit code 2 (grep pattern/IO error) is FAIL. The truth source is stdout emptiness — if stdout is non-empty, AC-LNC-007 FAILS regardless of exit code.

### AC-LNC-008 — dev-only-commands-isolation.md checklist updated (LOCAL-ONLY per spec.md §E)

**Given** M1 documentation update is complete,
**When** the orchestrator greps for the new agent-local-namespace checklist entries in the LOCAL copy only (the template mirror is intentionally absent per `.moai/docs/dev-only-commands-isolation.md` §21 isolation policy + spec.md §E Out of Scope),
**Then** at least 3 new verification command markers are present.

```bash
grep -cE "agents/local|release-update-specialist|github-specialist" .moai/docs/dev-only-commands-isolation.md
```

Expected: ≥ 3 matches in the single local file. The template path `internal/template/templates/.moai/docs/dev-only-commands-isolation.md` is intentionally NOT in this grep — its absence is verified separately by AC-LNC-009's `find` command pattern (the directory itself does not exist under the template surface for `dev-only-commands-isolation.md` per §21).

### AC-LNC-009 — `internal/template/templates/.claude/agents/local/` does NOT exist (REQ-LNC-012 negative test)

**Given** the migration is complete and the namespace separation contract holds,
**When** the orchestrator asserts directory absence AND searches the template surface for any agents/local/ artifact,
**Then** the directory does not exist AND zero artifact files are returned.

Primary assertion — directory absence (most robust, independent of `find` exit-code quirks):

```bash
[ ! -d internal/template/templates/.claude/agents/local ] && echo PASS || echo FAIL
```

Expected: stdout = `PASS`. If `FAIL` is emitted, AC-LNC-009 FAILS immediately — the protected directory exists on the template surface, breaking REQ-LNC-012.

Supplementary negative test — path-based:

```bash
find internal/template/templates -path "*/agents/local/*"
```

Expected: Empty stdout. (Note: `find` returns exit code 0 whether or not the directory exists, so the truth source is stdout emptiness — paired with the `[ ! -d ... ]` assertion above for completeness.)

Additional negative test for the specific agent file names:

```bash
find internal/template/templates -name "release-update-specialist.md" -o -name "github-specialist.md"
```

Expected: Empty stdout.

### AC-LNC-010 — generic-patterns-guide.md exists in both locations with 4 sections

**Given** M5 W5 authoring is complete,
**When** the orchestrator verifies file presence and section count,
**Then** both local and template copies exist with all 4 expected sections.

```bash
ls -la .moai/docs/generic-patterns-guide.md internal/template/templates/.moai/docs/generic-patterns-guide.md
```

Expected: Both files present with size > 5000 bytes. Then verify section structure:

```bash
grep -cE "^## (Multi-Session Race|Hook Setup|Settings Intent|Late-Branch)" internal/template/templates/.moai/docs/generic-patterns-guide.md
```

Expected: Count = 4 (one heading per externalized pattern family).

### AC-LNC-011 — Full Go test suite passes (commands_audit non-regression)

**Given** all M1-M5 changes are committed,
**When** the orchestrator runs the full Go test suite,
**Then** all tests pass including the Thin Command Pattern audit.

```bash
go test ./...
```

Expected: All packages report PASS, no FAIL. Specifically verify `internal/template/commands_audit_test.go` and any tests in `internal/template/` that validate the YAML frontmatter of `.claude/commands/97-*` and `.claude/commands/98-*`.

```bash
go test -v -run TestCommandsAudit ./internal/template/...
```

Expected: PASS for the audit test (Thin Command Pattern body LOC bound and YAML frontmatter shape verified).

### AC-LNC-012 — REQ-LNC-009 deferred verification marker

**Binds REQ**: REQ-LNC-009 (moai update PRESERVE behavior)

**Verification status**: Deferred to follow-up `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001` (code-level enforcement track) — REQ-LNC-009 documents a runtime invariant (`moai update` must preserve `.claude/agents/local/`) but the Go implementation change is explicitly out of scope per spec.md §E ("**`moai update` Go implementation change** ... is deferred to SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001"). This AC serves as a traceability anchor to maintain MP-3 bidirectional REQ↔AC binding within the present SPEC.

**Verification command** (deferred to follow-up SPEC):

```bash
grep -F "REQ-LNC-009" .moai/specs/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001/spec.md
# Expected (when follow-up SPEC exists): ≥ 1 cross-reference back to REQ-LNC-009 demonstrating handoff
# Until follow-up SPEC opens: returns 0 (acceptable for plan-phase; surfaced in M6 as a flag for sync-phase)
```

**Note**: This AC is intentionally "soft" within the present SPEC scope. The doctrinal anchor (REQ-LNC-009 + AC-LNC-012 pair) ensures the deferred work is not forgotten when SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 is opened. AC-LNC-012 does NOT block the 11 MUST-PASS acceptance criteria (AC-LNC-001 through AC-LNC-011); the Definition of Done in §D continues to require those 11 PASS. AC-LNC-012 transitions from SOFT to MUST only when the follow-up SPEC is authored.

## C. Edge Cases

| Edge Case | Handling | AC Coverage |
|-----------|----------|-------------|
| Local-only `CLAUDE.local.md` references in `.claude/` (NOT in `internal/template/templates/`) | Preserved — they are valid in maintainer-local context where CLAUDE.local.md exists | AC-LNC-007 only inspects template surface |
| 14th file or 18th leak found during M4 execution | M4 scope expands inline; AC-LNC-007 final pass is the truth source (zero matches) | AC-LNC-007 |
| `.moai/docs/dev-only-commands-isolation.md` already added 97-release-update.md row but not 98-github.md row (incomplete predecessor state) | M1 verifies BOTH rows present after update | AC-LNC-008 |
| `commands_audit_test.go` checks for `allowed-tools: Skill` literal (would fail after switch to `Agent`) | M3 must also update the test if the test hardcodes `Skill`; otherwise AC-LNC-011 fails and signals test must be updated | AC-LNC-011 |
| Template mirror has additional leaks (e.g., 18 occurrences vs grep-time estimate of 17) | M4 expands; AC-LNC-007 final pass is the truth source | AC-LNC-007 |

## D. Definition of Done

The SPEC is considered DONE when all of the following hold simultaneously:

1. All 11 AC-LNC-XXX checks pass per §B verification commands (orchestrator-side verification batch executed and all 7 commands return expected output).
2. Frontmatter `status: draft` → `implemented` in all 4 artifacts (spec.md, plan.md, acceptance.md, progress.md), with `sync_commit_sha:` populated atomically (per L60 atomic backfill obligation).
3. CHANGELOG.md contains exactly one new entry attributing SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 with AC count = 11, file count summary, and merged_commit reference.
4. progress.md §E Run-phase Audit-Ready Signal section is authored with verification command outputs verbatim.
5. Mx Step C judgment recorded (EVALUATE-SKIP expected for markdown-only changes, but if any .go file is touched in M3 to update commands_audit_test.go, EVALUATE-EXECUTE applies and Mx must produce @MX delta annotation).
6. 4-phase close marker emitted: `<moai>DONE</moai>` from the final orchestrator turn.

## E. Quality Gate Criteria

| Gate | Threshold | Verification |
|------|-----------|--------------|
| EARS/GEARS compliance | 11/11 AC use Given-When-Then structure | Manual review at plan-auditor + sync-phase audit |
| Independent verifiability | 11/11 AC have grep/test/file-existence commands | This document § B (each AC has ≥ 1 verification command) |
| Thin Command Pattern | Both 97/98 wrappers ≤ 20 LOC body | AC-LNC-002 |
| Template-First Rule | All template changes precede local changes (or are atomic with `make build` regen) | M5 milestone authoring order; AC-LNC-010 |
| Namespace separation contract | `.claude/agents/local/` registered in 3 SSOT locations | AC-LNC-006 + AC-LNC-008 |
| Zero template leak | grep returns 0 matches under `internal/template/templates/` | AC-LNC-007 (truth source) |
| Go test suite non-regression | All packages PASS | AC-LNC-011 |

## F. HISTORY

| Version | Date | Author | Iteration | Description |
|---------|------|--------|-----------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | iter-1 | Initial acceptance criteria authoring — 11 AC-LNC-XXX entries (AC-LNC-001 through AC-LNC-011) with per-AC verification commands, edge cases table, Definition of Done, Quality Gate Criteria. |
| 0.1.1 | 2026-05-25 | manager-spec | iter-2 | Focused defect resolution per plan-auditor iter-1 0.73 FAIL — D4 AC-LNC-007 exit code `1` → `in {0, 1}` (allows robust shell variants), D5 AC-LNC-009 prepended `[ ! -d ... ]` directory-absence assertion as primary truth source (find exit-code quirks not load-bearing), D6 AC-LNC-006 verification broadened to dual SSOT (agent-authoring.md + skill-authoring.md) with two grep commands, D7 AC-LNC-008 grep target reduced to local path only (template mirror intentionally absent per spec.md §E), D9 AC-LNC-003 wording `8 phases` → `9 phases (Phase 0 through Phase 8)` clarified inclusive endpoints, D11 AC-LNC-004 threshold ≥4 → ≥10 (matches observed predecessor section structure depth), D8 HISTORY section NEW. tier:M frontmatter added per D13. AC count 11 → 11 (no addition; AC-LNC-006 binding broadens to REQ-LNC-011 + REQ-LNC-014). |
| 0.1.2 | 2026-05-25 | manager-spec | iter-3 | Narrow-scope surgical defect resolution per plan-auditor iter-2 0.74 PASS-WITH-DEBT (stagnation, LEAN STOP signal): D_new4 (MUST-FIX) NEW AC-LNC-012 deferred-verification marker added — binds orphan REQ-LNC-009 (moai update PRESERVE behavior) which previously had 0 AC binding (MP-3 bidirectional traceability FAIL). AC-LNC-012 is SOFT severity (substantive verification deferred to follow-up SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001), does NOT block the 11 MUST-PASS criteria. AC count 11 → 12. D_new3 propagation: AC-LNC-006 binding reverted to `REQ-LNC-011` only (REQ-LNC-014 deleted in spec.md iter-3 as redundant subset of REQ-LNC-011 second clause); Verification 1/Verification 2 dual grep commands preserved (both still verify REQ-LNC-011's two clauses — first clause `.claude/agents/local/` row in agent-authoring.md, second clause deprecation entries in skill-authoring.md). |
