# Agent Delegation Patterns

Alfred가 Task()로 에이전트에게 위임할 때 사용하는 핵심 패턴들.

## 기본 위임 문법

```
result = await Task(
    subagent_type="에이전트_이름",
    prompt="구체적이고 명확한 작업 설명",
    context={"필요한": "컨텍스트"}
)
```

---

## 패턴 1: 직렬 위임 (Sequential)

**용도**: 에이전트 간에 의존성이 있을 때

**흐름**:
```
1. design-phase 완료
   ↓
2. 결과를 다음 에이전트의 context로 전달
   ↓
3. implementation-phase 시작
```

**예시**:
```
# Phase 1: 설계
design = Task(api-designer, "API 설계")

# Phase 2: 구현 (설계 결과 전달)
implementation = Task(
    backend-expert,
    "API 구현",
    context={"api_design": design}
)
```

---

## 패턴 2: 병렬 위임 (Parallel)

**용도**: 에이전트 간 독립적일 때

**흐름**:
```
Task1 (백엔드)   →  완료 기다림
Task2 (프론트엔드) → 완료 기다림
Task3 (문서)     → 완료 기다림
      모두 완료 → 통합
```

**예시**:
```
results = await Promise.all([
    Task(backend-expert, "백엔드 구현"),
    Task(frontend-expert, "프론트엔드 구현"),
    Task(docs-manager, "문서 생성")
])
```

---

## 패턴 3: 조건부 위임 (Conditional)

**용도**: 결과에 따라 다른 에이전트를 선택

**흐름**:
```
분석 완료
  ↓
문제 유형 판단
  ├→ 보안 문제 → security-expert
  ├→ 성능 문제 → performance-engineer
  └→ 품질 문제 → quality-gate
```

---

## 컨텍스트 전달 가이드

### 필수 컨텍스트 포함

```
context={
    "spec_id": "SPEC-001",           # 작업 ID
    "requirements": [리스트],        # 요구사항
    "constraints": [제약사항]         # 제약사항
}
```

### 불필요한 것 제외

❌ 전체 코드베이스
❌ 모든 파일 내용
❌ 과거 대화 이력
❌ 큰 바이너리 데이터

### 최적 컨텍스트 크기

- 최소: 에이전트가 작업하기 필요한 정보만
- 최대: 50K 토큰 이내

---

## 에러 처리

**실패 시**:
1. `debug-helper`에게 오류 분석 요청
2. 근본 원인 파악
3. 다른 에이전트로 재시도 또는 복구

---

## 위임 체크리스트

- [ ] 에이전트 이름 정확 (소문자, 하이픈)
- [ ] prompt가 구체적인가
- [ ] context에 필요한 정보만 포함
- [ ] 의존성을 고려했는가 (순차 vs 병렬)
- [ ] 에러 처리 계획이 있는가

---

자세한 에이전트 설명은 @.moai/memory/agents.md를 참고한다.
