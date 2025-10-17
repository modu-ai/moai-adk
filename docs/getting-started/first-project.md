# @DOC:START-FIRST-001 | Chain: @SPEC:DOCS-003 -> @DOC:START-001

# First Project: TODO 앱

실전 예제로 TODO 앱을 처음부터 끝까지 구현해보겠습니다.

## 프로젝트 목표

간단한 TODO 앱을 MoAI-ADK 워크플로우로 구현:

- ✅ TODO 항목 CRUD (Create, Read, Update, Delete)
- ✅ 우선순위 설정 (High, Medium, Low)
- ✅ 완료 상태 토글
- ✅ 전체 테스트 커버리지 ≥ 85%

---

## Step 1: 프로젝트 초기화

```bash
# 프로젝트 생성
moai-adk init todo-app
cd todo-app

# Claude Code 세션 시작
claude-code .
```

---

## Step 2: SPEC 작성

Alfred에게 SPEC 작성 요청:

```bash
/alfred:1-spec "TODO 앱 - CRUD 기능 구현"
```

생성된 SPEC (`.moai/specs/SPEC-TODO-001/spec.md`):

```markdown
## @SPEC:TODO-001 Overview

### EARS Requirements

**Ubiquitous Requirements**:
- REQ-TODO-001-001: 시스템은 TODO 항목 생성/조회/수정/삭제를 지원해야 한다
- REQ-TODO-001-002: 각 TODO는 제목, 설명, 우선순위, 완료 상태를 가져야 한다

**Event-driven Requirements**:
- REQ-TODO-001-003: WHEN 사용자가 TODO 생성하면, 시스템은 고유 ID를 할당해야 한다
- REQ-TODO-001-004: WHEN 사용자가 완료 토글하면, 시스템은 상태를 반전시켜야 한다

**State-driven Requirements**:
- REQ-TODO-001-005: WHILE TODO가 완료 상태일 때, UI에서 취소선을 표시해야 한다
```

---

## Step 3: TDD 구현

Alfred에게 TDD 구현 요청:

```bash
/alfred:2-build "SPEC-TODO-001"
```

### 생성된 코드 구조

```
src/todo_app/
  ├── models.py      # @CODE:TODO-MODEL-001: TODO 모델
  ├── service.py     # @CODE:TODO-SERVICE-001: 비즈니스 로직
  └── api.py         # @CODE:TODO-API-001: REST API

tests/
  ├── test_models.py    # @TEST:TODO-MODEL-001
  ├── test_service.py   # @TEST:TODO-SERVICE-001
  └── test_api.py       # @TEST:TODO-API-001
```

### 테스트 커버리지 확인

```bash
pytest --cov=src/todo_app --cov-report=html
# 커버리지: 92% ✅ (목표 85% 달성)
```

---

## Step 4: 문서 동기화

Alfred에게 문서 동기화 요청:

```bash
/alfred:3-sync
```

생성된 문서:

```
docs/
  ├── api/todo.md        # @DOC:TODO-API-001: API 참조
  └── guides/todo.md     # @DOC:TODO-GUIDE-001: 사용 가이드
```

---

## Step 5: 결과 확인

### TAG 체인 검증

```
@SPEC:TODO-001 (TODO 앱 요구사항)
  ├─ @CODE:TODO-MODEL-001 (모델)
  │   └─ @TEST:TODO-MODEL-001
  ├─ @CODE:TODO-SERVICE-001 (서비스)
  │   └─ @TEST:TODO-SERVICE-001
  ├─ @CODE:TODO-API-001 (API)
  │   └─ @TEST:TODO-API-001
  └─ @DOC:TODO-API-001 (문서)
```

### 품질 메트릭

- ✅ 테스트 커버리지: 92%
- ✅ TRUST 원칙 준수: 100%
- ✅ TAG 추적성: 완전 연결
- ✅ 문서화: 자동 생성

---

## 다음 단계

첫 프로젝트를 완성했습니다! 이제 고급 기능을 배워보세요:

1. [Configuration](../configuration/config-json.md) - Personal/Team 모드 설정
2. [Agents](../agents/spec-builder.md) - 각 에이전트 상세 가이드
3. [Hooks](../hooks/overview.md) - 커스텀 Hook 작성

---

**다음**: [Configuration →](../configuration/config-json.md)
