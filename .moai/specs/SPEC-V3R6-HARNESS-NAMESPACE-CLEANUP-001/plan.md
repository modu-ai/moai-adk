---
id: SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001
title: "Plan — Harness Namespace 누출 검증 및 정리"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: Low
phase: "v3.0.0 cleanup"
module: ".claude/agents/, .claude/skills/, internal/template/"
lifecycle: spec-anchored
tags: "harness, namespace, cleanup, tier-s, minimal"
issue_number: null
tier: S
---

# Plan — SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001

## §A. 구현 계획 (Tier S Minimal — 1-Pass)

본 SPEC은 **Tier S (minimal)** 분류이다. 사유:

- 변경 범위: 6 파일 삭제 + 1 디렉토리 제거 + Go test 1-2개 추가 + cross-ref 검증 (read-only)
- 코드 복잡도: Go test scaffold 1-2 파일 (`internal/template/` 패키지 하위)
- Risk: 매우 낮음 — runtime 동작에 무관한 dev-only 잔여물 + 회귀 방지 안전망 추가
- 의존성: SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 머지 완료 (`767bc04a4`)
- Section B-E 생략 (Tier S 정책): Trade-off / Known Issues / Pre-flight / Constraints / Anti-Patterns 별도 섹션 생략, 핵심 정보는 spec.md §3-§7에서 이미 충분히 다룸

### §A.1 Audit 검증 결과 (Plan-phase)

본 plan-phase에서 다음 4-shot audit를 수행했다 (read-only):

| Audit 항목 | 명령 | 결과 |
|------------|------|------|
| Template `agents/` 구조 | `find internal/template/templates/.claude/agents -type d` | `{., core, expert, meta}` — `harness/` 없음 (PASS) |
| Template `moai-harness-*` skill | `ls -d internal/template/templates/.claude/skills/moai-harness-*` | `moai-harness-learner/` 단일 (PASS) |
| 로컬 `.claude/agents/harness/` | `ls .claude/agents/harness/` | 4개 specialist `.md` 파일 (VIOLATION) |
| 로컬 `moai-harness-*` skills | `ls -d .claude/skills/moai-harness-*` | `cli-template`, `learner`, `patterns` 3개 (VIOLATION 2개: `cli-template`, `patterns`) |

**결론**: Template은 정상 (REQ-HNC-001 PASS), 로컬 dev 프로젝트만 정리 필요 (REQ-HNC-002 actionable).

### §A.2 §24 Cross-Reference 일관성 audit (Plan-phase)

REQ-HNC-003 검증 대상 6개 cross-ref를 grep으로 확인했다:

| 위치 | 라인 | 정합성 |
|------|------|--------|
| `internal/cli/update.go` | 1133-1136 (isUserAreaPath) | PASS — `my-harness-*` + `.claude/agents/my-harness/` 보호 |
| `internal/cli/update.go` | 1166 (isUserOwnedNamespace REQ-UNP-002 comment) | PASS — `.claude/agents/harness/` 명시 |
| `internal/cli/update.go` | 1240-1244 (isMoaiManaged exclusion) | PASS — `agents/harness/` 의도적 제외 명시 |
| `internal/cli/update_namespace_protect.go` | 7-10 | PASS — REQ-UNP-006 sentinel 동작 명시 |
| `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy | (read 보류, 본 SPEC 범위 외 검증) | DEFER to run-phase |
| `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention | (read 보류) | DEFER to run-phase |

Run-phase에서 마지막 2개 (skill-authoring.md, agent-authoring.md)는 5-10분 read-only 확인으로 검증한다.

### §A.3 Run-phase Milestones (3개, 우선순위 순)

#### M1: Backup + Cleanup (Priority High, Atomic)

**Scope**: REQ-HNC-002 + REQ-HNC-007 + REQ-HNC-004(1, 2, 3)

1. `mkdir -p .moai/backups/harness-namespace-cleanup-{ISO}/` (ISO format: `2026-05-25T10-30-45Z`, hyphenated)
2. 다음 6개 파일을 backup 경로에 복사:
   - `.claude/agents/harness/cli-template-specialist.md`
   - `.claude/agents/harness/hook-ci-specialist.md`
   - `.claude/agents/harness/quality-specialist.md`
   - `.claude/agents/harness/workflow-specialist.md`
   - `.claude/skills/moai-harness-cli-template/SKILL.md`
   - `.claude/skills/moai-harness-patterns/SKILL.md`
3. `.complete` marker 파일 작성 (UPDATE-NAMESPACE-PROTECT-001 패턴 답습)
4. Verify backup byte-identical 후 원본 삭제 (`rm` 6개 + `rmdir` 비어진 `.claude/agents/harness/` + `rmdir` `.claude/skills/moai-harness-cli-template/` + `rmdir` `.claude/skills/moai-harness-patterns/`)
5. REQ-HNC-004 3-grep verification 수행 (모두 0 results 확인)

**Files modified**: 0 source files. 6 dev-only files deleted. 1 backup directory created.

**Risk**: Low — backup-first pattern으로 복구 가능. `git status` 확인 후 commit 시 의도적 deletion.

#### M2: Go Integration Test 추가 (Priority Medium)

**Scope**: REQ-HNC-005

`internal/template/` 패키지 하위 (구체적 파일명 run-phase manager-develop 결정, 예: `embedded_namespace_test.go` 또는 기존 `embedded_test.go`에 추가) 다음 2개 invariant test 추가:

```go
// TestTemplateAgentsStructure — REQ-HNC-005 invariant 1
func TestTemplateAgentsStructure(t *testing.T) {
    expected := map[string]bool{"core": true, "expert": true, "meta": true}
    entries, err := embedded.ReadDir(".claude/agents")
    require.NoError(t, err)
    actual := make(map[string]bool)
    for _, e := range entries {
        if e.IsDir() {
            actual[e.Name()] = true
        }
    }
    require.Equal(t, expected, actual,
        "internal/template/templates/.claude/agents/ MUST only contain {core,expert,meta} subdirs (CLAUDE.local.md §24)")
}

// TestTemplateMoaiHarnessSkillsAllowlist — REQ-HNC-005 invariant 2
func TestTemplateMoaiHarnessSkillsAllowlist(t *testing.T) {
    allowed := map[string]bool{"moai-harness-learner": true}
    entries, err := embedded.ReadDir(".claude/skills")
    require.NoError(t, err)
    for _, e := range entries {
        name := e.Name()
        if e.IsDir() && strings.HasPrefix(name, "moai-harness-") {
            require.True(t, allowed[name],
                "Unexpected moai-harness-* skill in template: %s (CLAUDE.local.md §24.1 only allows moai-harness-learner)", name)
        }
    }
}
```

**Files modified**: 1 Go test file (`internal/template/embedded_namespace_test.go` 권장 신규 파일, run-phase 확정).

**Risk**: Low — read-only assertions against embedded.go.

#### M3: Cross-Reference Doc Verification (Priority Low, Read-only)

**Scope**: REQ-HNC-003 잔여 검증

run-phase manager-develop이 다음 2개 파일을 Read tool로 확인 + §24 contract 일관성 evidence를 progress.md에 기록:

- `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention

**Files modified**: 0 (verification only). Evidence를 progress.md에 추가.

**Risk**: 매우 낮음.

### §A.4 PRESERVE Files (5 항목)

다음 파일들은 run-phase에서 **읽지도 수정하지도 않는다**:

1. `internal/template/templates/.claude/agents/{core,expert,meta}/**` — template system agents (M2 test만 metadata 읽음)
2. `internal/template/templates/.claude/skills/moai-harness-learner/**` — valid `moai-harness-*` skill
3. `internal/template/templates/.claude/skills/moai-meta-harness/**` — valid `moai-meta-*` skill (builder)
4. `.moai/harness/**` — user-owned per REQ-UNP-003
5. `internal/governance/**`, `internal/session/registry*` — COORD-001 활성 작업 영역 (disjoint scope)

### §A.5 EXTEND Files (run-phase에서 신규/변경)

| 파일 | 동작 | LOC 추정 |
|------|------|----------|
| `.moai/backups/harness-namespace-cleanup-{ISO}/` | NEW (backup dir + 6 files + .complete marker) | ~22 KB copy |
| `internal/template/embedded_namespace_test.go` (또는 추가) | NEW Go test 파일 | ~50 LOC |
| `.claude/agents/harness/*.md` | DELETE 4 files | -9,267 bytes |
| `.claude/skills/moai-harness-cli-template/SKILL.md` | DELETE + DELETE empty dir | -4,683 bytes |
| `.claude/skills/moai-harness-patterns/SKILL.md` | DELETE + DELETE empty dir | -10,052 bytes |
| `.moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001/progress.md` | UPDATE run-evidence | +60 lines |

**Total impact**: ~24 KB의 dev-only 누출 파일 제거, ~50 LOC Go test 추가, 1 backup 디렉토리 생성.

### §A.6 Verification Batch (Run-phase Completion)

Run-phase 완료 시 orchestrator가 7-item parallel verification batch 수행 (per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution):

```bash
# 1. Go test 추가분 PASS
go test -run TestTemplateAgentsStructure -v ./internal/template/...
go test -run TestTemplateMoaiHarnessSkillsAllowlist -v ./internal/template/...

# 2. Full Go test suite 회귀 없음
go test ./...

# 3. REQ-HNC-004 3-grep verification (post-cleanup)
find internal/template/templates/.claude/agents -type d -name harness | wc -l  # → 0
find .claude/agents/harness -type f 2>/dev/null | wc -l                          # → 0
ls -d .claude/skills/moai-harness-cli-template .claude/skills/moai-harness-patterns 2>/dev/null | wc -l  # → 0

# 4. Backup integrity
ls .moai/backups/harness-namespace-cleanup-*/.complete  # → 1 file

# 5. Lint clean
golangci-lint run --timeout=2m

# 6. Build verifies embedded.go invariant
make build

# 7. moai update Contract 회귀 (Optional, M2 test가 statically cover하므로)
go test -run TestPreserveMyHarness -v ./internal/cli/...
```

### §A.7 Cross-References

- spec.md §1-§7 (배경, EARS, 비요구사항, 제약, 의존성, 영향, exclusions)
- acceptance.md §D (AC matrix)
- CLAUDE.local.md §24 (Harness Namespace 분리 정책 SSOT)
- SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 spec.md + plan.md (구현 contract 원천)
- chore commit `4f1135684` (2026-05-23 template cleanup 선례)
- `internal/cli/update_namespace_protect.go` (`isUserOwnedNamespace` + backup pattern 참조)
- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution (verification batch 패턴)
