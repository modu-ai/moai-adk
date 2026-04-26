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

# Implementation Tasks — SPEC-V3R3-CMD-CLEANUP-001

Tasks are organized into three sequential waves matching plan.md §2. Each task lists target files, expected output, and validation step.

---

## Wave 1 — Gate Command Addition (Priority High)

### W1-T1 — Create gate.md.tmpl in template tree

- **File**: `internal/template/templates/.claude/commands/moai/gate.md.tmpl` (NEW)
- **Action**: Write tool. Schema-mirror of `plan.md.tmpl` adapted for gate.
- **Content** (verbatim):

  ```
  ---
  description: {{if eq .ConversationLanguage "ko"}}lint+format+type-check+test 병렬 실행 (pre-commit 품질 게이트){{else if eq .ConversationLanguage "ja"}}lint+format+type-check+testを並列実行（プリコミット品質ゲート）{{else if eq .ConversationLanguage "zh"}}并行执行lint+format+type-check+test（提交前质量门禁）{{else}}Run lint+format+type-check+test in parallel (pre-commit quality gate){{end}}
  argument-hint: "[--fix]"
  allowed-tools: Skill
  ---

  Use Skill("moai") with arguments: gate $ARGUMENTS
  ```

- **Validation**: `wc -l internal/template/templates/.claude/commands/moai/gate.md.tmpl` returns ≤8 lines.
- **Maps to**: REQ-CMD-001, REQ-CMD-005

### W1-T2 — Regenerate embedded.go

- **File**: `internal/template/embedded.go` (REGENERATED)
- **Action**: Run `make build` from repo root.
- **Validation**: `git diff internal/template/embedded.go | grep -q "gate.md.tmpl"` confirms inclusion. `go build ./...` succeeds.
- **Maps to**: REQ-CMD-005

### W1-T3 — Create local gate.md (rendered)

- **File**: `.claude/commands/moai/gate.md` (NEW)
- **Action**: Write tool. Render `gate.md.tmpl` with `ConversationLanguage="ko"` (per local language.yaml).
- **Content** (verbatim):

  ```
  ---
  description: lint+format+type-check+test 병렬 실행 (pre-commit 품질 게이트)
  argument-hint: "[--fix]"
  allowed-tools: Skill
  ---

  Use Skill("moai") with arguments: gate $ARGUMENTS
  ```

- **Validation**: `wc -l .claude/commands/moai/gate.md` returns ≤8.
- **Maps to**: REQ-CMD-001

### W1-T4 — Run TestCommandsThinPattern

- **Command**: `cd /Users/goos/MoAI/moai-adk-go && go test ./internal/template/... -run TestCommandsThinPattern -v`
- **Expected**: PASS. New `gate.md.tmpl` validated against the Thin Command Pattern (frontmatter + ≤20 LOC body + Skill() invocation).
- **Maps to**: REQ-CMD-001 (verification)

### W1-T5 — Verify SKILL.md routing already references gate

- **File**: `.claude/skills/moai/SKILL.md` (READ-ONLY in Wave 1)
- **Action**: Grep verify line 74 contains `**gate** (aliases: check, pre-commit)` and §Priority 3 line 87 contains gate keyword routing.
- **Validation**: `grep -E '\*\*gate\*\*.*check.*pre-commit' .claude/skills/moai/SKILL.md` returns one match.
- **Outcome**: No edit needed. Routing already wired.
- **Maps to**: REQ-CMD-001

### W1 Commit

- **Message**: `feat(commands): SPEC-V3R3-CMD-CLEANUP-001 — add /moai gate command (Thin Command Pattern)`
- **Files**: `internal/template/templates/.claude/commands/moai/gate.md.tmpl`, `internal/template/embedded.go`, `.claude/commands/moai/gate.md`

---

## Wave 2 — Review/Sync Security Inline (Priority High)

### W2-T1 — Enhance review.md Phase 4 (template + local)

- **Files**:
  - `internal/template/templates/.claude/skills/moai/workflows/review.md`
  - `.claude/skills/moai/workflows/review.md`
- **Action**: Edit tool. Locate Phase 4 (security perspective). Add or restructure to include three explicit subsections:

  1. **Dependency vulnerability scan**:
     - Enumerate manifest files: `go.mod`, `package.json`, `requirements.txt`, `Cargo.toml`, `pyproject.toml`, plus `Gemfile`, `composer.json`, `mix.exs`, `Package.swift`, `pubspec.yaml`.
     - Auto-detect language from project markers; invoke `expert-security` with detected manifest.
     - Reference: `${CLAUDE_SKILL_DIR}/workflows/security.md` for full procedure.

  2. **Secrets git-history scan**:
     - Scan FULL git history (not just working tree) — pattern: `git log -p --all -G '(-----BEGIN [A-Z]+ PRIVATE KEY-----|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36})'` or equivalent.
     - Cross-reference with `.gitignore` to confirm no historical leaks.
     - Distinguish from working-tree-only scans (which other tools cover).

  3. **Data isolation check**:
     - Verify multi-tenant boundaries (no cross-tenant data flow).
     - Verify PII separation (PII never logged or sent to telemetry endpoints).
     - Verify shared-state leakage (no mutable globals carrying request-scoped data).

  - Each subsection MUST reference `Skill("moai")` `security` workflow as the canonical procedure.

- **Validation**:
  ```bash
  grep -i "dependency.*vulnerability" .claude/skills/moai/workflows/review.md
  grep -iE "secrets.*(git history|git log|--all)" .claude/skills/moai/workflows/review.md
  grep -iE "data isolation|multi-tenant|PII separation" .claude/skills/moai/workflows/review.md
  diff .claude/skills/moai/workflows/review.md internal/template/templates/.claude/skills/moai/workflows/review.md
  ```
- **Maps to**: REQ-CMD-002

### W2-T2 — Enhance sync.md Phase 0.55 (template + local)

- **Files**:
  - `internal/template/templates/.claude/skills/moai/workflows/sync.md`
  - `.claude/skills/moai/workflows/sync.md`
- **Action**: Edit tool. Locate Phase 0.55. Add explicit instruction:

  - Phase 0.55 audits BOTH:
    - (a) changed source files of the current SPEC (existing behavior), AND
    - (b) all dependency-manifest files at project root (NEW), regardless of whether they changed in this SPEC.
  - Manifest list (canonical): `go.mod`, `package.json`, `requirements.txt`, `Cargo.toml`, `pyproject.toml`, `Gemfile`, `composer.json`, `mix.exs`, `Package.swift`, `pubspec.yaml`.
  - When any manifest is detected, invoke `expert-security` for dependency vulnerability scan.
  - Rationale: detect drift between SPEC implementation and dependency surface (e.g., transitive vulnerability introduced by an unrelated dependency change since last sync).

- **Validation**:
  ```bash
  grep -i "manifest" .claude/skills/moai/workflows/sync.md
  grep -ciE "(go\.mod|package\.json|requirements\.txt|Cargo\.toml|pyproject\.toml)" .claude/skills/moai/workflows/sync.md  # >= 3
  diff .claude/skills/moai/workflows/sync.md internal/template/templates/.claude/skills/moai/workflows/sync.md
  ```
- **Maps to**: REQ-CMD-003

### W2-T3 — Regenerate embedded.go

- **Command**: `make build`
- **Validation**: `go build ./...` succeeds.

### W2 Commit

- **Message**: `feat(skills): SPEC-V3R3-CMD-CLEANUP-001 — inline security depth into review/sync workflows`
- **Files**: `internal/template/templates/.claude/skills/moai/workflows/review.md`, `internal/template/templates/.claude/skills/moai/workflows/sync.md`, `.claude/skills/moai/workflows/review.md`, `.claude/skills/moai/workflows/sync.md`, `internal/template/embedded.go`

---

## Wave 3 — Context Skill Removal (Priority Medium)

### W3-T1 — Delete context.md from template

- **File**: `internal/template/templates/.claude/skills/moai/workflows/context.md` (DELETE)
- **Action**: `rm internal/template/templates/.claude/skills/moai/workflows/context.md` via Bash.
- **Pre-check**: Confirm no `.tmpl` variant: `ls internal/template/templates/.claude/skills/moai/workflows/context.md*`.
- **Validation**: `! test -f internal/template/templates/.claude/skills/moai/workflows/context.md`.
- **Maps to**: REQ-CMD-004, REQ-CMD-005

### W3-T2 — Delete local context.md

- **File**: `.claude/skills/moai/workflows/context.md` (DELETE)
- **Action**: `rm .claude/skills/moai/workflows/context.md`.
- **Validation**: `! test -f .claude/skills/moai/workflows/context.md`.
- **Maps to**: REQ-CMD-004

### W3-T3 — Edit SKILL.md routing entries (template + local)

- **Files**:
  - `internal/template/templates/.claude/skills/moai/SKILL.md`
  - `.claude/skills/moai/SKILL.md`
- **Action**: Edit tool. Two edits per file:

  1. Remove §Priority 1 line 73 (or wherever it lives):
     - Remove: `- **context** (aliases: ctx, memory): Extract and display git-based context memory`
  2. Remove from §Priority 3 keyword routing if any line references "context" as a routing keyword (unlikely; verify with grep first).

- **Validation**:
  ```bash
  ! grep -E '\*\*context\*\*.*aliases.*ctx.*memory' .claude/skills/moai/SKILL.md
  ! grep -E '\*\*context\*\*.*aliases.*ctx.*memory' internal/template/templates/.claude/skills/moai/SKILL.md
  diff .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md
  ```
- **Maps to**: REQ-CMD-004

### W3-T4 — Regenerate embedded.go and run audit test

- **Command**: `make build`
- **Validation**:
  ```bash
  cd /Users/goos/MoAI/moai-adk-go
  go build ./...
  go test ./internal/template/... -run TestCommandsThinPattern -v
  ```
- **Expected**: PASS.

### W3-T5 — Full test suite

- **Command**: `cd /Users/goos/MoAI/moai-adk-go && go test ./...`
- **Expected**: PASS (no test references the removed `context.md` skill).
- **Mitigation**: If any test fails referencing `context`, examine whether the test asserts the skill exists. Update the test to no longer expect `context` in routing tables.

### W3 Commit

- **Message**: `chore(skills): SPEC-V3R3-CMD-CLEANUP-001 — remove unused context skill + routing`
- **Files**: `internal/template/templates/.claude/skills/moai/workflows/context.md` (deleted), `.claude/skills/moai/workflows/context.md` (deleted), `internal/template/templates/.claude/skills/moai/SKILL.md`, `.claude/skills/moai/SKILL.md`, `internal/template/embedded.go`

---

## Final Validation (after all waves)

```bash
# AC-001 — gate command
test -f .claude/commands/moai/gate.md && \
test -f internal/template/templates/.claude/commands/moai/gate.md.tmpl && \
go test ./internal/template/... -run TestCommandsThinPattern

# AC-002 — review.md Phase 4
grep -i "dependency.*vulnerability" .claude/skills/moai/workflows/review.md && \
grep -iE "secrets.*(git history|git log|--all)" .claude/skills/moai/workflows/review.md && \
grep -iE "data isolation|multi-tenant|PII separation" .claude/skills/moai/workflows/review.md

# AC-003 — sync.md Phase 0.55
grep -i "manifest" .claude/skills/moai/workflows/sync.md && \
[ "$(grep -ciE 'go\.mod|package\.json|requirements\.txt|Cargo\.toml|pyproject\.toml' .claude/skills/moai/workflows/sync.md)" -ge 3 ]

# AC-004 — context removed
! test -f .claude/skills/moai/workflows/context.md && \
! test -f internal/template/templates/.claude/skills/moai/workflows/context.md && \
! grep -E '\*\*context\*\*.*aliases.*ctx' .claude/skills/moai/SKILL.md

# AC-005 — security as skill only
! test -f .claude/commands/moai/security.md && \
! test -f internal/template/templates/.claude/commands/moai/security.md && \
test -f .claude/skills/moai/workflows/security.md && \
test -f internal/template/templates/.claude/skills/moai/workflows/security.md && \
grep -q '\*\*security\*\*.*aliases.*audit.*sec' .claude/skills/moai/SKILL.md

# Build + tests
go build ./... && go test ./...
```

All commands must succeed for SPEC to be marked Done.
