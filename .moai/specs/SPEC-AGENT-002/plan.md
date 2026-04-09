# Implementation Plan: SPEC-AGENT-002

Created: 2026-04-09
SPEC Version: 1.0.0
Agent: manager-strategy

## 1. Overview

16개 MoAI 에이전트 정의의 본문 크기를 88% 축소(평균 700줄 → 80-120줄)하면서 기존 워크플로우를 100% 보존한다. 핵심 전략은 "이동, 삭제 아님" — 공통 콘텐츠는 rule로, 비기능적 콘텐츠만 삭제.

## 2. Architecture Decision

### 콘텐츠 이동 전략

| 콘텐츠 유형 | 현재 위치 | 이동 대상 | 근거 |
|-------------|----------|----------|------|
| Language Handling | 16개 에이전트 본문 | agent-common-protocol.md rule | 동일 내용 반복 |
| Output Format Rules | 16개 에이전트 본문 | agent-common-protocol.md rule | 동일 내용 반복 |
| MCP Fallback Strategy | 3개 에이전트 본문 | agent-common-protocol.md rule | 동일 내용 반복 |
| Essential Reference | 16개 에이전트 본문 | agent-common-protocol.md rule | CLAUDE.md 자동 로드 중복 |
| Research Integration | 4개 에이전트 본문 | 삭제 | 비기능적 가상 프로세스 |
| Orchestration Metadata | 16개 에이전트 본문 | 삭제 | Claude Code 미인식 |
| Team Collaboration 예시 | 8개 에이전트 본문 | 삭제 (요약은 유지) | 코드 블록 예시 불필요 |
| Works Well With | 8개 에이전트 본문 | 삭제 | Agent Catalog에서 관리 |
| Workflow Steps | 에이전트 본문 | 본문에 간결하게 유지 | 핵심 행동 지침 |

### Rule 파일 설계

```
internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
```

paths frontmatter 없음 (모든 에이전트에 적용). 내용:
- Language Handling Protocol (conversation_language 기반 응답)
- Output Format Protocol (Markdown for users, XML for agent coordination)
- MCP Fallback Strategy (Context7 unavailable 시 대응)
- CLAUDE.md 참조 원칙 (자동 로드되므로 본문에서 반복 불필요)

## 3. Implementation Phases

### Phase 1: Rule 파일 생성 (REQ-001)

1. `agent-common-protocol.md` rule 파일 생성
2. 기존 에이전트에서 공통 섹션 추출하여 통합
3. 검증: rule 내용이 기존 에이전트 본문의 공통 섹션을 완전히 포함하는지 확인

### Phase 2: Agent Body 축소 (REQ-002) — 16개 에이전트 순차 처리

처리 순서 (의존성 없음, 복잡도순):

**Group A - 가장 큰 에이전트 (5개)**
1. manager-git (~1,330줄)
2. manager-spec (~1,040줄)
3. manager-project (~980줄)
4. expert-backend (~970줄)
5. expert-frontend (~874줄)

**Group B - 중간 크기 에이전트 (6개)**
6. manager-strategy (~820줄)
7. expert-devops (~810줄)
8. manager-ddd (~800줄)
9. expert-testing (~770줄)
10. manager-tdd (~747줄)
11. manager-docs (~675줄)

**Group C - 작은 에이전트 (5개)**
12. manager-quality (~680줄)
13. expert-performance (~665줄)
14. expert-security (~653줄)
15. expert-debug (~431줄)
16. expert-refactoring (~226줄)

각 에이전트 처리 절차:
1. 현재 본문에서 Workflow Steps 추출 → 간결 버전 작성 (WHY/IMPACT 주석 축약, 단계 보존)
2. 공통 섹션 제거 (Language, Output, Essential Reference, MCP Fallback)
3. 비기능 섹션 제거 (Research Integration, Orchestration Metadata, Works Well With)
4. Team Collaboration 코드 블록 예시 제거 (요약은 Delegation Protocol에 유지)
5. 결과 검증: 워크플로우 단계 수 비교 (before/after 동일해야 함)

### Phase 3: Frontmatter 수정 (REQ-003 ~ REQ-006)

모든 에이전트에 동시 적용:
1. maxTurns 필드 제거 (22개 에이전트)
2. permissionMode 수정 (4개 에이전트)
3. tools 목록 최소화 (5개 에이전트)
4. Hook matcher에 MultiEdit 추가 (2개 에이전트)

### Phase 4: Stale Reference 정리 (REQ-007)

1. Grep으로 존재하지 않는 참조 일괄 탐색
2. 수정 또는 제거
3. 검증: 모든 참조가 실존 자원을 가리키는지 확인

### Phase 5: Template Rebuild + 검증

1. `make build` 실행하여 embedded.go 재생성
2. `go test ./...` 전체 테스트 통과 확인
3. 로컬 `.claude/agents/moai/`에도 동일 변경 적용

## 4. Workflow Preservation Verification

[HARD] 각 에이전트 축소 후 다음을 검증:

| 검증 항목 | 방법 |
|----------|------|
| Workflow Step 보존 | Before/After 단계 수 비교 — 동일해야 함 |
| Delegation Protocol 보존 | 위임 대상 에이전트 목록 비교 — 동일해야 함 |
| Scope Boundaries 보존 | IN/OUT 목록 비교 — 동일해야 함 |
| SEMAP Contract 보존 | Pre/Post/Invariant/Forbidden 비교 — 동일해야 함 |
| Skill References 보존 | frontmatter skills 목록 비교 — 동일해야 함 |
| Hook Configuration 보존 | hooks 설정 비교 — 동일 (MultiEdit 추가 제외) |

## 5. Risk Analysis

| Risk | Impact | Likelihood | Mitigation |
|------|--------|-----------|-----------|
| Workflow step 누락 | High | Medium | Before/After 단계 수 자동 비교 스크립트 |
| Rule 로드 실패 | High | Low | paths frontmatter 없이 전역 적용 검증 |
| make build 실패 | Medium | Low | 즉시 감지 가능, 빠른 수정 |
| 테스트 실패 | Medium | Low | go test ./... 필수 실행 |

## 6. Success Metrics

| Metric | Before | Target |
|--------|--------|--------|
| 평균 에이전트 본문 크기 | ~700줄 | 80-120줄 |
| 토큰 소비 (에이전트 호출 당) | ~10-20K | ~1-2K |
| permissionMode 불일치 | 4건 | 0건 |
| 존재하지 않는 참조 | ~50건+ | 0건 |
| 중복 콘텐츠 인스턴스 | 48+ (16 agents x 3 sections) | 1 (공통 rule) |
| Workflow Step 보존율 | N/A | 100% |
