---
id: SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001
title: "Acceptance — Harness Namespace 누출 검증 및 정리"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: Low
phase: "v3.0.0 cleanup"
module: ".claude/agents/, .claude/skills/, internal/template/"
lifecycle: spec-anchored
tags: "harness, namespace, cleanup, acceptance, tier-s"
issue_number: null
tier: S
---

# Acceptance — SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001

## §D. AC Matrix

| AC ID | REQ 매핑 | 검증 방법 | Severity |
|-------|---------|----------|----------|
| AC-HNC-001 | REQ-HNC-001 | Template structural invariant grep (3 commands all 0) | MUST |
| AC-HNC-002 | REQ-HNC-002 | 6 leaked files + 2 empty dirs removed | MUST |
| AC-HNC-003 | REQ-HNC-003 | Cross-reference doc evidence (6 cross-refs verified) | SHOULD |
| AC-HNC-004 | REQ-HNC-004 | 3-grep post-cleanup verification all return 0 | MUST |
| AC-HNC-005 | REQ-HNC-005 | Go integration test 2 cases PASS | MUST |
| AC-HNC-006 | REQ-HNC-006 | `git diff --stat -- internal/template/templates/` empty | MUST |
| AC-HNC-007 | REQ-HNC-007 | Backup directory with `.complete` marker exists | MUST |

## §D.1 Given-When-Then Scenarios

### AC-HNC-001 (Template Invariant)

**Given** template directory `internal/template/templates/.claude/`
**When** the auditor runs these 3 commands:
```bash
find internal/template/templates/.claude/agents -type d -name harness | wc -l
ls -d internal/template/templates/.claude/skills/moai-harness-* 2>/dev/null | wc -l
find internal/template/templates/.claude/agents -mindepth 1 -maxdepth 1 -type d | sort
```
**Then**:
- Command 1 returns `0`
- Command 2 returns `1` (only `moai-harness-learner`)
- Command 3 returns 3 lines matching `{core, expert, meta}` in alphabetical order

### AC-HNC-002 (Local Cleanup)

**Given** local moai-adk-go dev project at HEAD before run-phase
**When** run-phase M1 (Backup + Cleanup) completes
**Then** the following filesystem state holds:
- `.claude/agents/harness/` directory does NOT exist (verified by `test ! -d .claude/agents/harness && echo OK`)
- `.claude/skills/moai-harness-cli-template/` directory does NOT exist
- `.claude/skills/moai-harness-patterns/` directory does NOT exist
- `.claude/skills/moai-harness-learner/` directory STILL exists (preserved per template-managed sync)

### AC-HNC-003 (Cross-Reference Self-Consistency)

**Given** §24 SSOT contract in CLAUDE.local.md
**When** the auditor reads the 6 cross-reference targets enumerated in REQ-HNC-003
**Then**:
- All 6 cross-refs cite §24 contract without contradicting language
- `internal/cli/update.go:1240-1244` explicitly notes `agents/harness/` is EXCLUDED from `isMoaiManaged`
- `internal/cli/update.go:1186` includes `agents/harness/` in `isUserOwnedNamespace`
- Run-phase evidence is recorded in `progress.md` §E.2 Run-phase Evidence

### AC-HNC-004 (Post-Cleanup 3-Grep)

**Given** run-phase M1 has executed and reported completion
**When** the orchestrator runs the 3 post-cleanup verifications in parallel:
```bash
find internal/template/templates/.claude/agents -type d -name harness 2>/dev/null | wc -l
find .claude/agents/harness -type f 2>/dev/null | wc -l
ls -d .claude/skills/moai-harness-cli-template .claude/skills/moai-harness-patterns 2>/dev/null | wc -l
```
**Then** all 3 commands return `0`.

### AC-HNC-005 (Go Integration Test)

**Given** run-phase M2 has added the 2 integration tests in `internal/template/`
**When** the orchestrator runs:
```bash
go test -run TestTemplateAgentsStructure -v ./internal/template/...
go test -run TestTemplateMoaiHarnessSkillsAllowlist -v ./internal/template/...
```
**Then**:
- Both tests PASS with exit code 0
- Test output contains `--- PASS: TestTemplateAgentsStructure` and `--- PASS: TestTemplateMoaiHarnessSkillsAllowlist`

### AC-HNC-006 (No Template Modification)

**Given** run-phase staged changes
**When** the orchestrator runs `git diff --stat -- internal/template/templates/`
**Then** the output is empty (no template files modified by this SPEC's run-phase).

NOTE: A NEW Go test file under `internal/template/` (NOT under `internal/template/templates/`) is permitted by REQ-HNC-005. The `templates/` subdirectory specifically remains untouched.

### AC-HNC-007 (Backup Exists)

**Given** run-phase M1 has executed
**When** the orchestrator runs:
```bash
ls -d .moai/backups/harness-namespace-cleanup-*/ 2>/dev/null | wc -l
find .moai/backups/harness-namespace-cleanup-*/ -name '.complete' 2>/dev/null | wc -l
find .moai/backups/harness-namespace-cleanup-*/ -type f -not -name '.complete' 2>/dev/null | wc -l
```
**Then**:
- Command 1 returns `1` (single backup directory)
- Command 2 returns `1` (`.complete` marker exists)
- Command 3 returns `6` (six original files backed up)

## §D.2 Edge Cases

### EC-HNC-001: Backup Race Collision

**Scenario**: Two run-phase executions within the same second.
**Mitigation**: REQ-HNC-007 backup directory naming SHALL handle same-second collision via numeric suffix (`-1`, `-2`, ...) per UPDATE-NAMESPACE-PROTECT-001 NFR-UNP-004 precedent. Implementation may delegate to existing `resolveNamespaceBackupDir` helper or replicate the pattern.

### EC-HNC-002: Partial Cleanup Failure

**Scenario**: M1 succeeds in deleting 4 files but errors on the 5th (e.g., permission denied).
**Mitigation**: M1 SHALL be **atomic-ish** — if any delete step fails after backup is complete, the run-phase MUST report blocker WITHOUT attempting partial recovery. Backup is the recovery source. The orchestrator can `cp -r .moai/backups/harness-namespace-cleanup-{ISO}/.claude/* .claude/` to restore. No silent partial state allowed.

### EC-HNC-003: Multi-Session Race

**Scenario**: COORD-001 run-phase modifies `internal/governance/` or `internal/session/registry*` while this SPEC's run-phase modifies `.claude/agents/harness/`.
**Mitigation**: Scopes are disjoint (verified at plan-phase). Pre-spawn `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` MUST return `N 0` or `0 0` before run-phase starts (per `.claude/rules/moai/core/agent-common-protocol.md` §Pre-Spawn Sync Check). If `N M` (diverged) detected, run-phase HALTS and surfaces blocker.

### EC-HNC-004: `moai-harness-learner` Skill Modified Externally

**Scenario**: Between plan-phase audit and run-phase, `moai-harness-learner` skill is modified or deleted.
**Mitigation**: REQ-HNC-005 Go test (`TestTemplateMoaiHarnessSkillsAllowlist`) verifies the allowlist at every build. If the skill is missing from template at run-phase, test FAILS and blocks merge — no silent regression.

## §D.3 Quality Gates

| Gate | Threshold |
|------|-----------|
| `go test ./...` | 100% PASS (no regression) |
| `go test ./internal/template/...` coverage | ≥ 85% (current baseline) |
| `golangci-lint run` | 0 new issues introduced |
| `git diff --stat -- internal/template/templates/` | empty (REQ-HNC-006) |
| Backup `.complete` marker | exactly 1 file at `.moai/backups/harness-namespace-cleanup-*/.complete` |
| 6 file deletions | all in `git status` as deletions, no untracked leftovers |

## §D.4 Definition of Done

본 SPEC은 다음 조건이 모두 충족될 때 `status: implemented`로 전환된다 (manager-develop run-phase 책임):

- [ ] AC-HNC-001 PASS (template invariant grep 3건 모두 0/1/3 expected values)
- [ ] AC-HNC-002 PASS (`.claude/agents/harness/`, `.claude/skills/moai-harness-cli-template/`, `.claude/skills/moai-harness-patterns/` 부재)
- [ ] AC-HNC-003 PASS (6 cross-refs 일관, evidence in progress.md)
- [ ] AC-HNC-004 PASS (3-grep post-cleanup all 0)
- [ ] AC-HNC-005 PASS (2 Go tests PASS via `go test ./internal/template/...`)
- [ ] AC-HNC-006 PASS (`git diff --stat -- internal/template/templates/` empty)
- [ ] AC-HNC-007 PASS (backup directory + `.complete` marker + 6 backed-up files)
- [ ] Verification batch 7-item parallel run all 0 exit codes
- [ ] No `internal/governance/` or `internal/session/registry*` touched (disjoint scope confirmed)
- [ ] progress.md §E.2 Run-phase Evidence updated by manager-develop

이후 `status: completed` 전환은 manager-docs sync-phase 책임 (CHANGELOG.md entry 추가 + 4 frontmatter status field 일괄 갱신).

## §D.5 Traceability Matrix

| REQ ID | AC ID | Plan Section | Verification Mechanism |
|--------|-------|--------------|------------------------|
| REQ-HNC-001 | AC-HNC-001 | plan.md §A.1, §A.6.6 | grep + `make build` |
| REQ-HNC-002 | AC-HNC-002 | plan.md §A.3 M1 | filesystem state check |
| REQ-HNC-003 | AC-HNC-003 | plan.md §A.2, §A.3 M3 | manual cross-ref read, evidence in progress.md |
| REQ-HNC-004 | AC-HNC-004 | plan.md §A.3 M1.5 | 3-grep batch |
| REQ-HNC-005 | AC-HNC-005 | plan.md §A.3 M2 | `go test` |
| REQ-HNC-006 | AC-HNC-006 | plan.md §A.4 PRESERVE | `git diff --stat` |
| REQ-HNC-007 | AC-HNC-007 | plan.md §A.3 M1.1-1.3 | backup directory inspection |

## §D.6 Closure Gates (Sync-phase)

manager-docs sync-phase responsibilities:

1. `CHANGELOG.md` Unreleased section에 entry 추가:
   ```
   ### Removed
   - `.claude/agents/harness/{cli-template,hook-ci,quality,workflow}-specialist.md` (4 dev-only leaked specialist agents)
   - `.claude/skills/moai-harness-cli-template/` (dev-only leaked skill)
   - `.claude/skills/moai-harness-patterns/` (dev-only leaked skill)
   ### Added
   - `internal/template/embedded_namespace_test.go` (or equivalent) regression test for CLAUDE.local.md §24 namespace invariant
   ```
2. 4 frontmatter `status` 전환: `draft` → `implemented` (spec.md, plan.md, acceptance.md, progress.md)
3. spec-lint baseline 회귀 점검

## §D.7 Forward-looking Self-Check

다음 항목은 본 SPEC 종료 후에도 향후 SPEC들이 §24 정책을 일관 유지하는지 확인할 회귀 신호:

- 새 `moai-harness-*` skill이 template에 추가될 경우 `TestTemplateMoaiHarnessSkillsAllowlist`가 자동 실패 → CI 차단
- 새 `.claude/agents/harness/<name>.md` 파일이 template에 추가될 경우 `TestTemplateAgentsStructure`가 자동 실패 → CI 차단
- `moai update` 회귀 테스트 (`TestPreserveMyHarness`)는 UPDATE-NAMESPACE-PROTECT-001 머지 시점부터 이미 활성 — 본 SPEC은 그 위에 template-side invariant를 추가하는 dual gate
