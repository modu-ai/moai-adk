---
spec_id: SPEC-V3R3-DEF-007
title: Acceptance Criteria — Convention Compliance Sweep
version: "1.0.0"
status: draft
created: 2026-04-25
related_spec: .moai/specs/SPEC-V3R3-DEF-007/spec.md
---

# Acceptance Criteria — SPEC-V3R3-DEF-007

## AC-DEF007-01: 11개 skill SKILL.md frontmatter에 progressive_disclosure 블록 존재

**Given** the 11 target skills exist under `.claude/skills/` and `internal/template/templates/.claude/skills/`
**When** the convention sweep is applied
**Then** each of the 22 SKILL.md files (11 skills × 2 paths) contains a `progressive_disclosure:` block with `enabled: true`, `level1_tokens: 100`, `level2_tokens: 5000`

### Verification

```bash
SKILLS="moai-domain-backend moai-domain-frontend moai-domain-db-docs moai-formats-data moai-framework-electron moai-library-mermaid moai-library-nextra moai-library-shadcn moai-tool-ast-grep moai-workflow-ddd moai-workflow-loop"
for skill in $SKILLS; do
  for base in .claude/skills internal/template/templates/.claude/skills; do
    grep -q "progressive_disclosure:" "$base/$skill/SKILL.md" || echo "MISSING: $base/$skill/SKILL.md"
    grep -q "level1_tokens: 100" "$base/$skill/SKILL.md" || echo "MISSING level1: $base/$skill/SKILL.md"
    grep -q "level2_tokens: 5000" "$base/$skill/SKILL.md" || echo "MISSING level2: $base/$skill/SKILL.md"
  done
done
# Expected: empty output (no MISSING)
```

Maps to: REQ-DEF007-001, REQ-DEF007-002

---

## AC-DEF007-02: manager-git agent body에 Scope Boundaries + Delegation Protocol 섹션 존재

**Given** the manager-git.md exists at both local and template paths
**When** the agent body update is applied
**Then** both `.claude/agents/moai/manager-git.md` and `internal/template/templates/.claude/agents/moai/manager-git.md` contain `## Scope Boundaries` and `## Delegation Protocol` sections

### Verification

```bash
for path in .claude/agents/moai/manager-git.md internal/template/templates/.claude/agents/moai/manager-git.md; do
  grep -q "^## Scope Boundaries" "$path" || echo "MISSING Scope Boundaries: $path"
  grep -q "^## Delegation Protocol" "$path" || echo "MISSING Delegation Protocol: $path"
done
# Expected: empty output
```

Maps to: REQ-DEF007-003, REQ-DEF007-004

---

## AC-DEF007-03: 변경이 additive only (본문 / 기존 frontmatter 무수정)

**Given** a baseline checkout of the affected files
**When** the convention sweep is applied
**Then** `git diff` reveals only insertion of the progressive_disclosure block and the two new agent sections; no deletions, no reordering of existing fields

### Verification

```bash
git diff --stat .claude/skills/ internal/template/templates/.claude/skills/ .claude/agents/moai/manager-git.md internal/template/templates/.claude/agents/moai/manager-git.md
# Expected: only added lines (no removed lines except whitespace normalization)
git diff -U0 .claude/skills/moai-domain-backend/SKILL.md | grep "^-" | grep -v "^---"
# Expected: empty (no removed content lines)
```

Maps to: REQ-DEF007-006, REQ-DEF007-008, REQ-DEF007-009

---

## AC-DEF007-04: make build 성공 + embedded.go 갱신

**Given** template files have been modified
**When** the developer runs `make build`
**Then** the build succeeds and `internal/template/embedded.go` is regenerated to reflect the template changes

### Verification

```bash
make build
# Expected: exit code 0
ls -la internal/template/embedded.go
# Expected: mtime newer than template modification time
```

Maps to: REQ-DEF007-007

---

## AC-DEF007-05: 회귀 테스트 통과

**Given** all changes applied
**When** the developer runs `go test ./internal/template/...`
**Then** all tests pass

### Verification

```bash
go test -count=1 ./internal/template/...
# Expected: PASS, no FAIL
```

Maps to: REQ-DEF007-007

---

## Edge Cases

### EC-1: Skill에 이미 progressive_disclosure 블록 일부만 존재

If a skill already has `progressive_disclosure:` but missing `level2_tokens`, the sweep MUST detect this and update only the missing field. The sweep MUST NOT duplicate the block.

### EC-2: manager-git.md에 이미 Scope Boundaries 섹션 존재

If grep detects `^## Scope Boundaries` already present, the sweep MUST skip insertion for that file and report "Already compliant: <path>".

### EC-3: Template과 local에 서로 다른 내용

If template and local diverge before the sweep, the sweep MUST first surface the diff via `git diff` and STOP — do not blindly apply both. Surface to operator.

---

## Definition of Done

- [ ] AC-DEF007-01: 22개 SKILL.md verification 통과
- [ ] AC-DEF007-02: 2개 manager-git.md verification 통과
- [ ] AC-DEF007-03: git diff additive only 확인
- [ ] AC-DEF007-04: make build 성공
- [ ] AC-DEF007-05: go test 통과
- [ ] Edge Case EC-1, EC-2, EC-3 처리 로직 검증
