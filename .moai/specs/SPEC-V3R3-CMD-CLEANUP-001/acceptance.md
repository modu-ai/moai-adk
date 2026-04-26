---
id: SPEC-V3R3-CMD-CLEANUP-001
version: "0.1.0"
status: draft
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
priority: P0
labels: [commands, skills, cleanup, security, template-first]
issue_number: null
---

# Acceptance Criteria — SPEC-V3R3-CMD-CLEANUP-001

## 1. AC-REQ Traceability Matrix

| AC | REQ-CMD-001 | REQ-CMD-002 | REQ-CMD-003 | REQ-CMD-004 | REQ-CMD-005 | REQ-CMD-006 |
|----|:-----------:|:-----------:|:-----------:|:-----------:|:-----------:|:-----------:|
| AC-001 |     X       |             |             |             |             |             |
| AC-002 |             |     X       |             |             |             |             |
| AC-003 |             |             |     X       |             |             |             |
| AC-004 |             |             |             |     X       |     X       |             |
| AC-005 |             |             |             |             |             |     X       |

Coverage: 6/6 REQs covered, 5/5 ACs map to at least one REQ. 100% traceability.

---

## 2. Detailed Given-When-Then Scenarios

### AC-001 — Gate command callable end-to-end

**Covers**: REQ-CMD-001

**Given**:
- Local repository has `.claude/commands/moai/gate.md` (Thin Command Pattern, <20 LOC body)
- Local repository has `.claude/skills/moai/workflows/gate.md` (existing skill, unchanged)
- `internal/template/embedded.go` has been regenerated via `make build` and includes `.claude/commands/moai/gate.md.tmpl`
- A Go project with at least one `.go` file is present at the repo root

**When**:
1. User starts a Claude Code session at the repo root
2. User types `/moai gate` and submits

**Then**:
- Claude Code resolves `/moai gate` to `.claude/commands/moai/gate.md`
- The command frontmatter `allowed-tools: Skill` is honored
- The body invokes `Skill("moai")` with argument `gate`
- The `moai` skill router (SKILL.md) matches the `gate` subcommand at §Priority 1
- The router reads `${CLAUDE_SKILL_DIR}/workflows/gate.md` per the SKILL.md routing entry (line 134)
- The gate skill executes lint, format, type-check, and test in parallel for Go (`go vet`, `gofmt -l`, `go vet`, `go test -race ./...`) per the language-detection logic
- A single aggregate exit code is reported (0 on all-pass, non-zero on any-fail)

**Verification commands**:
```bash
# File exists and follows Thin Command Pattern
test -f .claude/commands/moai/gate.md
wc -l < .claude/commands/moai/gate.md  # should be < 20
grep -q '^allowed-tools: Skill$' .claude/commands/moai/gate.md
grep -q 'Use Skill("moai") with arguments: gate' .claude/commands/moai/gate.md

# Template parallel exists
test -f internal/template/templates/.claude/commands/moai/gate.md.tmpl

# Audit test passes
cd /Users/goos/MoAI/moai-adk-go && go test ./internal/template/... -run TestCommandsThinPattern
```

**Edge cases**:
- Project without recognized language markers: gate skill passes silently per `coding-standards.md` §Language-Specific Guidelines.
- `--fix` argument: passed through to gate skill which auto-applies fixers where supported.
- Mixed-language project: gate skill iterates over each detected language toolchain.

---

### AC-002 — Review Phase 4 includes security depth

**Covers**: REQ-CMD-002

**Given**:
- `.claude/skills/moai/workflows/review.md` and `internal/template/templates/.claude/skills/moai/workflows/review.md` have been edited per Wave 2

**When**:
- The Phase 4 section (security perspective) of either review.md is read

**Then**:
- The text contains an explicit reference to "dependency vulnerability scan" with at least one manifest filename enumerated (go.mod, package.json, requirements.txt, Cargo.toml, or pyproject.toml)
- The text contains an explicit reference to "secrets" qualified by "full git history" or "git log" (distinguishing from working-tree-only scans)
- The text contains an explicit reference to "data isolation" covering at least one of: multi-tenant boundary, PII separation, shared-state leakage
- The text references the `security` skill (`security.md`) as the canonical procedure (preserving single source of truth)

**Verification commands**:
```bash
# Each axis verified by independent grep
grep -i "dependency.*vulnerability" .claude/skills/moai/workflows/review.md
grep -iE "go\.mod|package\.json|requirements\.txt|Cargo\.toml|pyproject\.toml" .claude/skills/moai/workflows/review.md
grep -iE "secrets.*(git history|git log|--all)" .claude/skills/moai/workflows/review.md
grep -iE "data isolation|multi-tenant|PII separation" .claude/skills/moai/workflows/review.md

# Template mirror identical
diff .claude/skills/moai/workflows/review.md internal/template/templates/.claude/skills/moai/workflows/review.md
```

**Edge cases**:
- review.md Phase 4 may already mention "security" generically; the test specifically requires the three new axes (dependency, history scan, isolation).
- Reference (not copy) of security.md skill — full procedural detail stays in `security.md` to avoid duplication.

---

### AC-003 — Sync Phase 0.55 audits manifests

**Covers**: REQ-CMD-003

**Given**:
- `.claude/skills/moai/workflows/sync.md` and template mirror have been edited per Wave 2

**When**:
- The Phase 0.55 section of sync.md is read

**Then**:
- The text contains an explicit instruction to audit dependency-manifest files at project root in addition to changed source files
- At least 3 manifest filenames are enumerated (from: go.mod, package.json, requirements.txt, Cargo.toml, pyproject.toml, Gemfile, composer.json, mix.exs, Package.swift, pubspec.yaml)
- The audit triggers `expert-security` (or invokes the `security` skill) when any manifest is detected

**Verification commands**:
```bash
grep -iE "manifest" .claude/skills/moai/workflows/sync.md
grep -ciE "(go\.mod|package\.json|requirements\.txt|Cargo\.toml|pyproject\.toml)" .claude/skills/moai/workflows/sync.md  # >= 3
diff .claude/skills/moai/workflows/sync.md internal/template/templates/.claude/skills/moai/workflows/sync.md
```

---

### AC-004 — Context skill removed and audit test green

**Covers**: REQ-CMD-004, REQ-CMD-005

**Given**:
- Wave 3 has been applied (deletion + SKILL.md edits + `make build`)

**When**:
- The repository file system is inspected
- The audit test suite is run

**Then**:
- `.claude/skills/moai/workflows/context.md` does NOT exist
- `internal/template/templates/.claude/skills/moai/workflows/context.md` does NOT exist
- `.claude/skills/moai/SKILL.md` contains no `context` routing entry under §Priority 1 (subcommand list)
- `.claude/skills/moai/SKILL.md` contains no `context` keyword under §Priority 3 (keyword routing)
- Template mirror SKILL.md identical to local
- `go test ./internal/template/... -run TestCommandsThinPattern` returns PASS
- `go test ./...` returns PASS (no test references the removed file)

**Verification commands**:
```bash
# Files absent
! test -f .claude/skills/moai/workflows/context.md
! test -f internal/template/templates/.claude/skills/moai/workflows/context.md

# Routing entries absent
! grep -E '\*\*context\*\*.*ctx.*memory' .claude/skills/moai/SKILL.md
! grep -E 'context.*aliases.*ctx' .claude/skills/moai/SKILL.md

# Template mirror identical to local
diff .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md

# Tests green
cd /Users/goos/MoAI/moai-adk-go && go test ./internal/template/... -run TestCommandsThinPattern
cd /Users/goos/MoAI/moai-adk-go && go test ./...
```

**Edge cases**:
- A test file (`*_test.go`) under `internal/` may reference the deleted skill name as a string literal. If detected, update the test to no longer expect `context` in routing tables.
- `embedded.go` size will decrease by ~5,800 bytes after `make build`.

---

### AC-005 — Security as skill only

**Covers**: REQ-CMD-006

**Given**:
- All waves applied

**When**:
- The file system is inspected

**Then**:
- `.claude/commands/moai/security.md` does NOT exist
- `internal/template/templates/.claude/commands/moai/security.md.tmpl` does NOT exist
- `.claude/skills/moai/workflows/security.md` exists (unchanged byte-for-byte from current state)
- `internal/template/templates/.claude/skills/moai/workflows/security.md` exists (unchanged byte-for-byte from current state)
- `.claude/skills/moai/SKILL.md` retains the `security` routing entry under §Priority 1 (line 75: `**security** (aliases: audit, sec)`)
- `.claude/skills/moai/SKILL.md` retains the `security` keyword routing under §Priority 3 (line 88)

**Verification commands**:
```bash
# Command file absent
! test -f .claude/commands/moai/security.md
! test -f internal/template/templates/.claude/commands/moai/security.md
! test -f internal/template/templates/.claude/commands/moai/security.md.tmpl

# Skill file present and unchanged from baseline
test -f .claude/skills/moai/workflows/security.md
test -f internal/template/templates/.claude/skills/moai/workflows/security.md
# byte-for-byte unchanged from pre-SPEC commit:
git diff HEAD~1 -- .claude/skills/moai/workflows/security.md  # should be empty
git diff HEAD~1 -- internal/template/templates/.claude/skills/moai/workflows/security.md  # should be empty

# Routing entries retained
grep -q '\*\*security\*\*.*aliases.*audit.*sec' .claude/skills/moai/SKILL.md
grep -qE 'security.*owasp.*vulnerability.*injection.*xss.*csrf' .claude/skills/moai/SKILL.md
```

---

## 3. Definition of Done (DoD)

The SPEC is considered Done when ALL of the following hold:

1. All 5 acceptance criteria (AC-001 .. AC-005) verified via the listed commands.
2. All 6 EARS requirements (REQ-CMD-001 .. REQ-CMD-006) traced to at least one passing AC.
3. `go build ./...` succeeds.
4. `go test ./...` succeeds with no new failures (compare against baseline coverage).
5. `make build` runs cleanly and `internal/template/embedded.go` reflects the changes.
6. Each wave is committed separately with a Conventional Commit message referencing this SPEC ID:
   - W1: `feat(commands): SPEC-V3R3-CMD-CLEANUP-001 — add /moai gate command (Thin Command Pattern)`
   - W2: `feat(skills): SPEC-V3R3-CMD-CLEANUP-001 — inline security depth into review/sync workflows`
   - W3: `chore(skills): SPEC-V3R3-CMD-CLEANUP-001 — remove unused context skill + routing`
7. All file edits use Edit tool (per `.claude/rules/moai/core/moai-constitution.md` §Tool Selection Priority); Write tool used only for the new gate.md files.
8. No flag (--no-verify, --no-gpg-sign) bypassed during commit.

## 4. Quality Gates

- **TRUST 5 — Tested**: gate command verified by existing TestCommandsThinPattern; review/sync edits verified by grep-based AC; context removal verified by file-absence + go test.
- **TRUST 5 — Readable**: New gate.md.tmpl follows the existing plan.md.tmpl pattern (4-language description); review/sync edits use prose consistent with surrounding sections.
- **TRUST 5 — Unified**: Template-First discipline preserved (all changes in template, then mirrored).
- **TRUST 5 — Secured**: Security capabilities are strengthened (review.md Phase 4 + sync.md Phase 0.55) rather than weakened.
- **TRUST 5 — Trackable**: Three Conventional Commits, each referencing SPEC ID; AC-REQ matrix in §1.

## 5. Rollback Plan

Each wave is independently revertable:

- **Wave 1 rollback**: `git revert <W1-commit>`. The new gate.md command files are deleted; existing gate skill is unaffected (already worked via run/sync auto-invocation).
- **Wave 2 rollback**: `git revert <W2-commit>`. review.md Phase 4 and sync.md Phase 0.55 revert to pre-SPEC state. Security capability remains accessible via direct `/moai security` keyword routing (per §Priority 3).
- **Wave 3 rollback**: `git revert <W3-commit>`. context.md skill restored, SKILL.md routing entries restored. No data lost (skill body is preserved in git history).

## 6. Open Issues / Follow-ups

- After SPEC merge: consider whether existing `/moai review` and `/moai sync` user prompts (in skill body templates) need translation updates to reflect the new Phase 4 / Phase 0.55 capabilities. Out of scope for this SPEC.
- Consider whether `expert-security` agent definition needs an update to document the new invocation context (called from review/sync). Out of scope; tracked separately if needed.
