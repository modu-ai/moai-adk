# SPEC-AUDIT-GATE-INTEGRITY-001 — Acceptance Criteria

> 모든 AC는 기계 검증 가능 (명령 + 기대 출력). `verification-claim-integrity.md` §3.2 — Evidence는 verbatim 출력이며 요약 불가.

## §A Given-When-Then 시나리오

### 시나리오 1 — D7 BLOCKING이 Verdict를 강제한다 (R1)

- **Given**: superseded 상태의 SPEC을 reconciliation 절 없이 참조하는 SPEC 아티팩트
- **When**: plan-auditor가 plan-phase 감사를 수행하면
- **Then**: 보고서 `## Must-Pass Results`에 `MP-5 ... FAIL` 행이 기록되고, aggregate score가 threshold 이상이어도 `Verdict: FAIL`이며, 해당 finding이 `## Defects Found`에 severity=critical로 나타난다

### 시나리오 2 — hierarchical 모드 보고서가 worked example 형식을 따른다 (R2)

- **Given**: `harness.yaml`에 `evaluator_mode: hierarchical` 설정
- **When**: sync-auditor가 구현을 평가하면
- **Then**: 보고서가 `### Hierarchical-Mode Output Example`에 제시된 형식(sub-criterion 행, canonical anchor 0.25/0.50/0.75/1.00, per-dimension aggregation, must-pass verdict)을 따르고, 각 차원 Evidence 셀에 기계 검증 명령의 verbatim 출력이 인용된다

### 시나리오 3 — manager-spec이 실행된 Bash 출력으로 SPEC ID를 검증한다 (R4)

- **Given**: 신규 SPEC ID 후보 (예: `SPEC-EXAMPLE-001`)
- **When**: manager-spec이 spec.md `Write`를 준비하면
- **Then**: 응답 본문에 실행된 Bash one-liner와 그 verbatim 출력(`PASS` 또는 `FAIL`)이 인용되고, `FAIL`이면 Write가 중단되고 blocker report가 반환된다

### 시나리오 4 — 두 보고서 스트림이 상호 참조된다 (R3)

- **Given**: Phase 0.5 게이트가 skip-eligibility를 평가하는 상황
- **When**: 오케스트레이터가 spec-workflow.md § Report Persistence를 참조하면
- **Then**: "plan-phase review stream"의 final-iteration verdict가 run-gate 입력이고 "run-gate stream" date-file이 artifact-hash 대상임이 명시되어 있으며, plan-auditor.md에서도 동일한 두 스트림 구분과 상호 참조를 발견한다

## §D AC Matrix

모든 명령은 리포지토리 루트에서 실행. `L` = live, `T` = template mirror.

| AC | REQ | 검증 명령 | 기대 출력 |
|----|-----|-----------|-----------|
| AC-AGI-001 | AGI-001 | `grep -cE '\(MP-5\)\|\(MP-6\)' .claude/agents/moai/plan-auditor.md` | `≥ 2` (baseline 0) |
| AC-AGI-002 | AGI-002 | `grep -c 'must-pass-equivalent' .claude/agents/moai/plan-auditor.md` | `≥ 1` (baseline 0) |
| AC-AGI-003 | AGI-003 | `grep -c 'MP-5' .claude/agents/moai/plan-auditor.md && grep -c 'MP-6' .claude/agents/moai/plan-auditor.md` | 각각 `≥ 2` (M5 정의 + Must-Pass Results 표 행; baseline 각 0) |
| AC-AGI-004 | AGI-004 | `grep -c 'evaluator_mode: hierarchical' .claude/agents/moai/sync-auditor.md` | `≥ 2` (baseline 1 — 기존 HRN-003 절 + 신규 선택 규칙) |
| AC-AGI-005 | AGI-005 | `grep -c '### Hierarchical-Mode Output Example' .claude/agents/moai/sync-auditor.md` | `= 1` (baseline 0) |
| AC-AGI-006a | AGI-006 | `grep -c 'go test ./...' .claude/agents/moai/sync-auditor.md` | `≥ 1` (baseline 0) |
| AC-AGI-006b | AGI-006 | `grep -c 'golangci-lint run' .claude/agents/moai/sync-auditor.md` | `≥ 1` (baseline 0) |
| AC-AGI-006c | AGI-006 | `grep -cE 'go test -cover\|coverprofile' .claude/agents/moai/sync-auditor.md` | `≥ 1` (baseline 0) |
| AC-AGI-006d | AGI-006 | `grep -c 'verbatim' .claude/agents/moai/sync-auditor.md` | `≥ 1` (Evidence verbatim 인용 의무 문구; baseline 0) |
| AC-AGI-007 | AGI-007 | `grep -c 'moai-ref-owasp-checklist' .claude/agents/moai/sync-auditor.md && grep -c 'moai-ref-testing-pyramid' .claude/agents/moai/sync-auditor.md` | 각각 `= 1` (frontmatter skills; baseline 각 0) |
| AC-AGI-008a | AGI-008 | `grep -c 'plan-phase review stream' .claude/rules/moai/workflow/spec-workflow.md && grep -c 'run-gate stream' .claude/rules/moai/workflow/spec-workflow.md` | 각각 `≥ 1` (baseline 각 0) |
| AC-AGI-008b | AGI-008 | `grep -c 'plan-phase review stream' .claude/agents/moai/plan-auditor.md && grep -c 'run-gate stream' .claude/agents/moai/plan-auditor.md` | 각각 `≥ 1` (baseline 각 0) |
| AC-AGI-009 | AGI-009 | `grep -c 'final-iteration verdict' .claude/rules/moai/workflow/spec-workflow.md` | `≥ 1` (baseline 0) |
| AC-AGI-010a | AGI-010 | `grep -cE '=~ \^SPEC' .claude/agents/moai/manager-spec.md` | `≥ 1` (baseline 0 — 실행형 Bash one-liner 존재) |
| AC-AGI-010b | AGI-010 | `grep -c 'mentally' .claude/agents/moai/manager-spec.md` | `= 0` (baseline ≥1; self-check 섹션 외 잔존 발견 시 run-phase에서 baseline-delta로 재해석 후 보고) |
| AC-AGI-011 | AGI-011 | `grep -c 'L32 chain context' .claude/agents/moai/manager-spec.md` | `= 0` (baseline 1 — 역사 서사 압축) |
| AC-AGI-012a | AGI-012 | `grep -cE '\(MP-5\)' internal/template/templates/.claude/agents/moai/plan-auditor.md` | `≥ 1` (mirror 반영) |
| AC-AGI-012b | AGI-012 | `grep -c '### Hierarchical-Mode Output Example' internal/template/templates/.claude/agents/moai/sync-auditor.md` | `= 1` (mirror 반영) |
| AC-AGI-012c | AGI-012 | `grep -cE '=~ \^SPEC' internal/template/templates/.claude/agents/moai/manager-spec.md` | `≥ 1` (mirror 반영) |
| AC-AGI-012d | AGI-012 | `grep -c 'run-gate stream' internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md` | `≥ 1` (mirror 반영) |
| AC-AGI-012e | AGI-012 | `grep -rc 'SPEC-AUDIT-GATE-INTEGRITY\|REQ-AGI' internal/template/templates/.claude/agents/moai/ internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md \| awk -F: '{s+=$2} END {print s}'` | `0` (Neutrality — 내부 토큰 유입 없음) |
| AC-AGI-013 | AGI-012 | `make build; echo "exit=$?"` | `exit=0` |
| AC-AGI-014 | 전체 | `moai spec lint 2>&1 \| grep -c 'SPEC-AUDIT-GATE-INTEGRITY-001'` | `0` (본 SPEC lint finding 없음; 정확한 CLI 플래그는 run-phase pre-flight에서 확인) |

## §D.1 Edge Cases

- **E1 (병렬 세션 선점)**: pre-flight baseline grep이 이미 기대값이면 해당 REQ는 verify-then-SKIP — 중복 편집 금지, blocker report로 보고
- **E2 (mentally 토큰 다중 출현)**: manager-spec.md의 self-check 섹션 밖에서 `mentally`가 발견되면 그 출현은 scope 밖 — AC-AGI-010b를 "self-check 섹션 내 0건"으로 재해석하고 근거 grep -n 출력 인용
- **E3 (Must-Pass 표 서식 변형)**: plan-auditor 보고서 템플릿의 MP-5/6 행이 `[PASS/FAIL/N/A]` 3-상태를 갖는 경우 — D7/D8 검증 verb가 실행 불능(대상 파일 부재 등)일 때 N/A 허용 여부는 MP-4 선례(N/A auto-pass)를 따라 명문화
- **E4 (make build 실패)**: template 편집 후 build 실패 시 mirror 편집을 revert하지 말고 원인(YAML frontmatter 손상 등) 진단 후 수정 — 커밋 전 build green 필수

## §D.2 Quality Gate

- 전 AC 판정은 단일 턴 병렬 검증 배치로 실행하고 verbatim 출력을 file-redirect contract에 따라 보존
- plan-auditor 독립 감사 (Phase 2.3) PASS ≥ 0.85 (Tier M threshold)
- 커밋은 pathspec 한정 — 무관 파일 무포함을 `git show --stat`로 확인

## §D.3 Definition of Done

1. AC-AGI-001..014 전 행 PASS (verbatim 증거 인용)
2. live 4파일 + mirror 4파일 편집 완료, `make build` exit 0
3. Neutrality grep 0건 (AC-AGI-012e)
4. milestone별 pathspec 커밋 완료, 무관 변경 무포함
5. progress.md §E.2/§E.3에 run-phase 증거 기록 (manager-develop 소관)
