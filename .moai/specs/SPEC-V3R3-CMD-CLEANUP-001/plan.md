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

# Implementation Plan — SPEC-V3R3-CMD-CLEANUP-001

## 1. Overview

Three-axis command/skill surface cleanup, executed in three sequential waves. Each wave is independently committable and testable.

- **Wave 1**: Add `/moai gate` slash command file (template + local + audit test verification).
- **Wave 2**: Strengthen `review.md` Phase 4 and `sync.md` Phase 0.55 (single commit, two file edits).
- **Wave 3**: Remove `context.md` skill (template + local + SKILL.md routing entries + commands_audit_test rerun).

Total file changes: ~7 files. No new agents, no LOC-heavy work.

## 2. Approach

### Wave 1 — Gate Command Addition (Priority High)

**Goal**: REQ-CMD-001, REQ-CMD-005 (gate part).

**Tasks**:

- **W1-T1**: Create template file `internal/template/templates/.claude/commands/moai/gate.md.tmpl` following Thin Command Pattern. Use existing `plan.md.tmpl` as schema reference. Include 4-language description (ko/ja/zh/en) per project convention. argument-hint should advertise `[--fix]` flag (existing gate skill supports auto-fix).
- **W1-T2**: Run `make build` to regenerate `internal/template/embedded.go`. Verify embedded fileset includes `gate.md.tmpl`.
- **W1-T3**: Mirror locally: create `.claude/commands/moai/gate.md` (rendered, not template). Use ko description matching local language.yaml (`conversation_language: ko`).
- **W1-T4**: Run `go test ./internal/template/... -run TestCommandsThinPattern` to verify Thin Command Pattern compliance for the new file.
- **W1-T5**: Verify SKILL.md routing already references `gate` (it does — line 74). No SKILL.md edit needed in Wave 1.

**Files modified**:
- `internal/template/templates/.claude/commands/moai/gate.md.tmpl` (new)
- `internal/template/embedded.go` (regenerated)
- `.claude/commands/moai/gate.md` (new)

### Wave 2 — Review/Sync Security Inline (Priority High)

**Goal**: REQ-CMD-002, REQ-CMD-003.

**Tasks**:

- **W2-T1**: Edit `internal/template/templates/.claude/skills/moai/workflows/review.md` Phase 4. Add explicit references to: (1) dependency vulnerability scan (enumerate go.mod/package.json/requirements.txt/Cargo.toml/pyproject.toml), (2) secrets scan against full git history (specify `git log -p --all -G '...'` or equivalent), (3) data-isolation checks (multi-tenant boundary, PII separation, shared-state leakage). Reference `security.md` skill for full procedural detail. Mirror locally.
- **W2-T2**: Edit `internal/template/templates/.claude/skills/moai/workflows/sync.md` Phase 0.55. Add explicit instruction to audit dependency-manifest files (go.mod, package.json, requirements.txt, Cargo.toml, pyproject.toml) IN ADDITION to changed source files. Mirror locally.
- **W2-T3**: Run `make build`. Verify regenerated embedded matches local edits.

**Files modified**:
- `internal/template/templates/.claude/skills/moai/workflows/review.md`
- `internal/template/templates/.claude/skills/moai/workflows/sync.md`
- `.claude/skills/moai/workflows/review.md` (mirror)
- `.claude/skills/moai/workflows/sync.md` (mirror)
- `internal/template/embedded.go` (regenerated)

### Wave 3 — Context Skill Removal (Priority Medium)

**Goal**: REQ-CMD-004, REQ-CMD-005 (context part).

**Tasks**:

- **W3-T1**: Delete `internal/template/templates/.claude/skills/moai/workflows/context.md`. Verify no `.tmpl` variant exists.
- **W3-T2**: Delete local `.claude/skills/moai/workflows/context.md`.
- **W3-T3**: Edit `internal/template/templates/.claude/skills/moai/SKILL.md`:
  - Remove line 73 `context (aliases: ctx, memory)` from §Priority 1 subcommand list.
  - Remove `context` from §Priority 3 keyword routing if present.
  - Mirror to local `.claude/skills/moai/SKILL.md`.
- **W3-T4**: Run `make build`. Run `go test ./internal/template/... -run TestCommandsThinPattern` (no new failures expected — context was a skill, not a command).
- **W3-T5**: Run full test suite `go test ./...` to catch any test that referenced `context.md`.

**Files modified**:
- `internal/template/templates/.claude/skills/moai/workflows/context.md` (delete)
- `.claude/skills/moai/workflows/context.md` (delete)
- `internal/template/templates/.claude/skills/moai/SKILL.md` (edit — 1-2 line removal)
- `.claude/skills/moai/SKILL.md` (mirror)
- `internal/template/embedded.go` (regenerated)

## 3. Milestones

| Milestone | Wave | Deliverable | Priority |
|-----------|------|-------------|----------|
| M1 | W1 | `/moai gate` command file deployed in template + local + audit test green | High |
| M2 | W2 | review.md Phase 4 + sync.md Phase 0.55 contain explicit security depth | High |
| M3 | W3 | context.md skill + routing entries removed; `go test ./...` green | Medium |
| M4 | All | One commit per wave, all 5 ACs verified | High |

Sequencing: M1 → M2 → M3 → M4. M1 and M2 are independent of each other (could parallelize), but Wave 1 is shorter and validates the test harness first; running serially is simpler and safer.

## 4. Technical Approach

### gate.md.tmpl content (Wave 1)

Schema mirror of `plan.md.tmpl` (1 line of frontmatter per language + thin body).

```
---
description: {{if eq .ConversationLanguage "ko"}}lint+format+type-check+test 병렬 실행 (pre-commit 품질 게이트){{else if eq .ConversationLanguage "ja"}}lint+format+type-check+testを並列実行（プリコミット品質ゲート）{{else if eq .ConversationLanguage "zh"}}并行执行lint+format+type-check+test（提交前质量门禁）{{else}}Run lint+format+type-check+test in parallel (pre-commit quality gate){{end}}
argument-hint: "[--fix]"
allowed-tools: Skill
---

Use Skill("moai") with arguments: gate $ARGUMENTS
```

Body LOC: 1 (under the 20-LOC Thin Command threshold).

### Local gate.md (Wave 1)

Render the template with `ConversationLanguage="ko"` for local copy:

```
---
description: lint+format+type-check+test 병렬 실행 (pre-commit 품질 게이트)
argument-hint: "[--fix]"
allowed-tools: Skill
---

Use Skill("moai") with arguments: gate $ARGUMENTS
```

### review.md Phase 4 enhancement (Wave 2)

Read existing Phase 4 section. Append (or restructure) a sub-list under "Security Perspective":

- Dependency vulnerability scan: enumerate manifest files (go.mod, package.json, requirements.txt, Cargo.toml, pyproject.toml). Invoke `expert-security` with the relevant manifest as input. Reference `Skill("moai")` `security` workflow for the full procedure.
- Secrets git-history scan: scan full git history (not just working tree) using `git log -p --all` and pattern detection (AWS keys, GitHub tokens, private keys). Cross-reference with `.gitignore` to confirm no historical leaks.
- Data isolation check: verify multi-tenant boundaries, PII separation, shared-state leakage via cross-file reference graph.

### sync.md Phase 0.55 enhancement (Wave 2)

Read existing Phase 0.55 section. Add explicit instruction:

- Phase 0.55 audits BOTH (a) changed source files of the current SPEC AND (b) all dependency-manifest files at project root, regardless of whether they changed. Manifest list: go.mod, package.json, requirements.txt, Cargo.toml, pyproject.toml, Gemfile, composer.json, mix.exs, Package.swift, pubspec.yaml.
- Manifest audit triggers `expert-security` for dependency vulnerability check if any manifest is detected.

### SKILL.md edit (Wave 3)

Two locations:

- §Priority 1 (line 73): delete `**context** (aliases: ctx, memory): Extract and display git-based context memory`.
- §Priority 3 keyword routing: scan for "context" keyword routing line; remove if present.

## 5. Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Existing test references `context.md` content | Low | Medium | Run `go test ./...` after Wave 3 deletion; if any test fails referencing context, surface the test name and update or skip with rationale |
| `make build` fails due to embedded.go regeneration issue | Low | High | Verify with `go build ./...` after each wave; rollback wave if build breaks |
| Users in field have customized `context.md` locally | Very Low | Low | Skill is dead code; no documented user customization path. SKILL.md routing entry removal is non-breaking |
| review.md Phase 4 enhancement creates conflict with security.md skill content | Low | Low | review.md Phase 4 references security.md skill rather than copying its content — single source of truth preserved |
| `gate.md` command file conflicts with future `--team` flag work | Low | Low | argument-hint advertises only `[--fix]`; team-mode integration is delegated to the gate skill itself |

## 6. Validation

After all waves complete:

- [ ] `ls .claude/commands/moai/gate.md` returns the file (Wave 1)
- [ ] `ls internal/template/templates/.claude/commands/moai/gate.md.tmpl` returns the file (Wave 1)
- [ ] `go test ./internal/template/... -run TestCommandsThinPattern` passes (Wave 1)
- [ ] `grep -E "(dependency vulnerability|secrets.*git history|data isolation)" .claude/skills/moai/workflows/review.md` returns 3+ matches (Wave 2)
- [ ] `grep -E "(go.mod|package.json|requirements.txt|Cargo.toml).*manifest" .claude/skills/moai/workflows/sync.md` returns matches (Wave 2)
- [ ] `! ls .claude/skills/moai/workflows/context.md 2>/dev/null` (Wave 3 — file absent)
- [ ] `! grep -E "context.*aliases.*ctx" .claude/skills/moai/SKILL.md` (Wave 3 — routing absent)
- [ ] `go test ./...` green (all waves)
- [ ] AskUserQuestion: gate command discoverable via `/moai` autocomplete in a fresh Claude Code session

## 7. Out of Scope (deferred)

- E2E test harness for `/moai gate` user-flow (requires Claude Code instance — manual smoke test only).
- Migration of existing user `.claude/skills/moai/workflows/context.md` — no documented users; silent removal is acceptable per Non-Goals.
- Adding `effort` level configuration to gate command — defaults to medium per Opus 4.7 prompt philosophy (speed-critical).

## 8. Estimated Complexity

- Wave 1: Trivial (1 new file + render). 20 min.
- Wave 2: Low (2 files edit, ~30 lines of prose each). 45 min.
- Wave 3: Trivial (3 deletions + 1-2 line SKILL.md edit). 15 min.
- Test/validation: 15 min.
- Total: ~1.5 hours.
