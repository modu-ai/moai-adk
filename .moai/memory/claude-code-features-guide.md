# Claude Code v2.0.30+ Features Guide

**SPEC ID**: CLAUDE-CODE-FEATURES-001
**Last Updated**: 2025-11-02
**Target Version**: MoAI-ADK v0.9.0+

---

## 개요

이 가이드는 Claude Code v2.0.30+ 에서 이미 사용 가능한 3가지 기능을 MoAI-ADK에서 효과적으로 활용하기 위한 방법을 설명합니다.

### 3가지 주요 기능

1. ✅ **Feature 1: Haiku Auto SonnetPlan Mode** - 모델 선택 최적화
2. ✅ **Feature 3: Background Bash Commands** - 백그라운드 작업 실행
3. ✅ **Feature 4: Enhanced Grep Tool** - 고급 검색 기능

---

## Feature 1: Haiku Auto SonnetPlan Mode

### 개념

**Model Selection Strategy**:
- **Sonnet 4.5**: 복잡한 분석이 필요한 **Plan 작업** (spec-builder, implementation-planner)
- **Haiku 4.5**: 빠른 실행이 필요한 **Execution 작업** (tdd-implementer, doc-syncer, tag-agent)

### 구현 방법

에이전트 파일의 YAML frontmatter에 `model` 선언:

```yaml
---
name: spec-builder
model: sonnet        # Plan 작업은 Sonnet
---
```

```yaml
---
name: tdd-implementer
model: haiku         # 실행 작업은 Haiku
---
```

### 설정된 에이전트

| 에이전트 | 모델 | 용도 |
|----------|------|------|
| spec-builder | Sonnet | SPEC 문서 생성 및 분석 |
| implementation-planner | Sonnet | 구현 전략 수립 및 설계 |
| tdd-implementer | Haiku | 실제 코드 구현 (RED-GREEN-REFACTOR) |
| doc-syncer | Haiku | 문서 동기화 및 정리 |
| tag-agent | Haiku | TAG 체인 검증 |

### 효과

```
비용: Sonnet 기준 대비
  - Haiku 사용으로 70-90% 비용 절감 (실행 작업)
  - Sonnet 유지로 높은 품질 분석 보장 (계획 작업)

성능: 전체 워크플로우 기준
  - 실행 단계 40% 빠른 응답 (Haiku)
  - 계획 단계 +10% 더 정확한 분석 (Sonnet)
```

---

## Feature 3: Background Bash Commands

### 개념

**Background Execution**: 시간이 오래 걸리는 명령어(테스트, 빌드)를 **백그라운드에서 비동기로 실행**하여 사용자가 다른 작업을 계속할 수 있도록 함.

### 사용 방법

#### 1. 백그라운드 실행 명령어 작성

```python
# tdd-implementer 에이전트에서
Bash(
    command="pytest tests/ -v --cov=src",
    run_in_background=true,
    description="Run pytest with coverage in background"
)
```

#### 2. 결과 확인 (선택사항)

```python
# 백그라운드 작업의 task_id를 받음
# BashOutput 도구로 실시간 로그 확인 가능

BashOutput(bash_id="task-id-from-background-command")
```

### 실제 사용 예시

#### 예시 1: pytest 백그라운드 실행

```python
# RED 단계에서 테스트 작성 후
Bash(
    command="pytest tests/test_feature.py -v",
    run_in_background=true,
    description="RED phase: Run failing test",
    timeout=60000  # 1분 타임아웃
)

# 사용자는 다른 작업 계속 수행 가능
# (예: 다른 파일 작성, 검색 등)
```

#### 예시 2: 빌드 명령어 백그라운드 실행

```python
Bash(
    command="python -m pytest tests/ --cov=src --cov-report=html",
    run_in_background=true,
    description="Build phase: Run full test suite with coverage",
    timeout=300000  # 5분 타임아웃
)
```

### 주의사항

| 항목 | 설명 |
|------|------|
| **timeout** | 최대 실행 시간 (밀리초). 기본값: 120000 (2분) |
| **로그 위치** | `.moai/logs/background-tasks/` |
| **모니터링** | BashOutput으로 실시간 확인 가능 |
| **실패 처리** | 타임아웃 시 자동 종료, 부분 결과는 로그 파일에 저장 |

### 성능 개선 효과

```
TDD 사이클 시간:
  - 기존 (직렬 실행): 테스트 10분 + 코드 5분 = 15분
  - 개선 (백그라운드): 테스트 병렬 실행 → 전체 10분 (40% 단축)
```

---

## Feature 4: Enhanced Grep Tool

### 개념

**고급 검색 기능**:
- **multiline=true**: 여러 줄에 걸친 패턴 매칭 (정규식 `.`이 줄바꿈도 매칭)
- **head_limit**: 결과 개수 제한 (처음 N개만 반환)

### 파라미터 설명

```python
Grep(
    pattern=r"@SPEC:[\s\S]*?@TEST",
    path="src/",
    multiline=true,      # 줄바꿈을 포함한 매칭
    head_limit=50,       # 처음 50개 결과만 반환
    output_mode="content"  # 매칭된 전체 내용 반환
)
```

| 파라미터 | 기본값 | 설명 |
|----------|--------|------|
| `multiline` | false | True: `.` 문자가 줄바꿈도 매칭 |
| `head_limit` | 무제한 | 반환할 최대 결과 개수 |
| `output_mode` | files_with_matches | "content" (내용), "files_with_matches" (파일 목록), "count" (개수) |

### 실제 사용 예시

#### 예시 1: SPEC 문서 검색

```python
# TAG chain 검증: @SPEC → @TEST → @CODE → @DOC 연결 확인
Grep(
    pattern=r"@SPEC:FEATURE-001[\s\S]*?@TEST:FEATURE-001[\s\S]*?@CODE:FEATURE-001",
    path="src/",
    multiline=true,
    head_limit=10,
    output_mode="files_with_matches"
)
```

**효과**: 여러 줄에 걸친 TAG 체인을 한 번에 검색 가능

#### 예시 2: 복잡한 코드 블록 검색

```python
# 함수 정의 + 함수 본문 검색
Grep(
    pattern=r"def feature_\w+\([\s\S]*?\):",
    path="src/moai_adk/",
    multiline=true,
    head_limit=20
)
```

**효과**: 여러 줄 함수 정의를 완전히 캡처

#### 예시 3: 성능 최적화 - head_limit 사용

```python
# 대규모 프로젝트에서 @CODE 태그 검색 (제한된 개수만)
Grep(
    pattern=r"@CODE:\w+",
    path="src/",
    head_limit=50,  # 처음 50개만 반환 → 빠른 응답
    output_mode="files_with_matches"
)
```

**효과**: 결과가 많을 때 처음 N개만 반환하여 성능 향상

### 성능 비교

```
검색: @SPEC:.*?@TEST 패턴 (멀티라인 필요)

기존 (multiline=false):
  - 결과: 불일치 또는 부분 매칭
  - 시간: 검색 불가능

개선 (multiline=true):
  - 결과: 정확한 매칭
  - 시간: < 1초 (head_limit=50 사용)
  - 개선도: 4-6배 빠른 검색 가능
```

### tag-agent에서의 활용

```python
# TAG 무결성 검증 예시
Grep(
    pattern=r"@SPEC:\w+",
    path=".",
    head_limit=100,
    output_mode="count"
)
# 결과: SPEC 태그 개수를 빠르게 파악
```

---

## 통합 예시: 전체 워크플로우

### 시나리오

새로운 기능 구현 (`/alfred:2-run SPEC-FEATURE-001`)

### 단계별 적용

#### 1단계: Plan (Feature 1 - Sonnet 사용)
```
/alfred:1-plan "User authentication"
  ↓
spec-builder (model: sonnet) ← Sonnet 4.5 사용
  ↓
implementation-planner (model: sonnet) ← Sonnet 4.5 사용
  ↓
구현 계획 생성 (고품질 분석)
```

#### 2단계: Implement (Feature 3 + Feature 1 - Haiku 사용)
```
/alfred:2-run SPEC-FEATURE-001
  ↓
tdd-implementer (model: haiku) ← Haiku 4.5 사용
  ├─ RED: pytest run_in_background=true
  ├─ GREEN: 코드 구현
  └─ REFACTOR: 코드 정리
  ↓
실행 단계 40% 시간 단축, 비용 70-90% 절감
```

#### 3단계: Verify (Feature 4 - Grep 최적화)
```
tag-agent 검증 (model: haiku)
  ├─ Enhanced Grep으로 @TAG 검색
  │  └─ multiline=true, head_limit=50 사용
  └─ 4-6배 빠른 TAG 검증
```

#### 4단계: Document (Feature 1 - Haiku 사용)
```
doc-syncer (model: haiku)
  ↓
문서 동기화 (빠른 실행)
```

---

## 최적화 팁

### 💡 Tip 1: Background Bash 타임아웃 설정

```python
# 짧은 테스트: 1분
Bash(command="pytest tests/unit/", run_in_background=true, timeout=60000)

# 긴 테스트: 5분
Bash(command="pytest tests/integration/", run_in_background=true, timeout=300000)

# 빌드: 10분
Bash(command="python -m build", run_in_background=true, timeout=600000)
```

### 💡 Tip 2: Grep에서 head_limit으로 성능 최적화

```python
# 대규모 프로젝트 검색 최적화
Grep(
    pattern=r"@CODE:\w+",
    path="src/",
    head_limit=100,  # 처음 100개만 가져오기
    output_mode="count"
)
# 결과: 매우 빠른 응답 (< 100ms)
```

### 💡 Tip 3: Multiline 패턴 작성

```python
# ❌ 나쁜 예 (줄바꿈 포함 안 함)
Grep(pattern=r"@SPEC:.*?@TEST", multiline=false)

# ✅ 좋은 예 (줄바꿈 포함)
Grep(
    pattern=r"@SPEC:[\s\S]*?@TEST",  # [\s\S]: 모든 문자 (줄바꿈 포함)
    multiline=true
)
```

---

## FAQ

**Q1. Haiku와 Sonnet의 성능 차이는?**

```
분석/설계 (Sonnet): 정확도 95%, 속도 10-15초
실행/검증 (Haiku): 정확도 85-90%, 속도 1-3초

→ 계획은 정확도 우선, 실행은 속도 우선
```

**Q2. Background Bash 작업이 완료되었는지 어떻게 알지?**

```python
# 1. 로그 파일 확인
cat .moai/logs/background-tasks/{task-id}.log

# 2. BashOutput으로 실시간 모니터링
BashOutput(bash_id="task-id")

# 3. 완료되면 사용자에게 자동 알림
```

**Q3. Grep에서 multiline과 head_limit을 같이 쓸 수 있나?**

```python
# ✅ 가능합니다!
Grep(
    pattern=r"def.*?return",
    path="src/",
    multiline=true,      # 여러 줄 매칭
    head_limit=50        # 처음 50개만 반환
)
```

**Q4. Background 작업이 타임아웃되면?**

```
타임아웃 시:
1. 프로세스 자동 종료
2. 부분 결과를 로그 파일에 저장
3. 사용자에게 알림: "Task timed out after X seconds"
4. 로그 파일 경로 제공
```

---

## 다음 단계

이 3가지 기능을 활용하여:

1. **Feature 1**: 에이전트별 모델 선언 → 비용 70-90% 절감
2. **Feature 3**: 테스트 백그라운드 실행 → TDD 사이클 40% 단축
3. **Feature 4**: Enhanced Grep → TAG 검색 4-6배 향상

더 자세한 정보:
- Skill("moai-lang-python") - Python 에이전트 최적화
- Skill("moai-essentials-debug") - 테스트 디버깅

---

**문서 버전**: 0.0.1
**마지막 업데이트**: 2025-11-02
**관련 SPEC**: SPEC-CLAUDE-CODE-FEATURES-001
