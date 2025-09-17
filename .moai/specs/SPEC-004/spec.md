# SPEC-004: Claude Code 컨텍스트 메모리 최적화 @REQ:CONTEXT-MEMORY-001

> **@REQ:CONTEXT-MEMORY-001** "Claude Code 시작 시 컨텍스트 토큰 사용량 28% → 10% 최적화"

## 📋 요구사항 개요

### 비즈니스 요구사항

**WHEN** Claude Code가 MoAI-ADK 프로젝트에서 시작될 때
**THE SYSTEM SHALL** 컨텍스트 토큰 사용량을 10% 이하로 유지해야 한다
**IN ORDER TO** 개발자가 실제 작업에 충분한 컨텍스트 공간을 확보할 수 있게 한다

### 현재 문제점 분석

**심각도: 🔴 CRITICAL**

현재 상황 (55k/200k tokens, 28% 사용):
- **System tools**: 26.9k tokens (13.5%) - 🚨 최대 문제 영역
- **Memory files**: 15.3k tokens (7.7%) - ⚠️ 중요 문제
- **Custom agents**: 6.0k tokens (3.0%) - ⚠️ 개선 필요
- **MCP tools**: 3.5k tokens (1.8%) - 📝 관찰 대상
- **System prompt**: 3.3k tokens (1.7%) - ✅ 양호
- **Messages**: 345 tokens (0.2%) - ✅ 양호

### 비즈니스 임팩트

**비효율성 지표:**
- 개발자 작업 시작 전 이미 28% 컨텍스트 소모
- 복잡한 작업 시 컨텍스트 부족으로 품질 저하
- 빈번한 컨텍스트 초기화로 개발 흐름 중단
- 에이전트 체인 작업 시 메모리 부족 현상

### 목표 사용자

- **모든 MoAI-ADK 사용자**: 충분한 작업 공간 필요
- **대규모 프로젝트 개발자**: 복잡한 컨텍스트 관리 필요
- **에이전트 체인 사용자**: 긴 작업 시퀀스 실행 필요

## 🎯 핵심 기능 요구사항

### FR-1: 시스템 도구 설명 압축 @REQ:TOOL-COMPRESS-001

**WHEN** Claude Code가 초기화될 때
**THE SYSTEM SHALL** 도구 설명을 26.9k tokens에서 10k tokens 이하로 압축해야 한다

**압축 전략:**

```yaml
압축 대상 (우선순위):
  1. 중복 파라미터 설명: 5k tokens 절약
  2. 예시 코드 블록 축소: 8k tokens 절약
  3. 상세 설명 → 링크화: 4k tokens 절약

유지 대상:
  - 파라미터 타입과 필수 여부
  - 핵심 용도 설명 (1줄)
  - 에러 케이스 간략 설명
```

**목표 감소량:** 26.9k → 10k tokens (62% 감소)

### FR-2: 에이전트 지연 로딩 시스템 @REQ:LAZY-AGENT-001

**WHEN** 특정 에이전트가 실제로 호출될 때
**THE SYSTEM SHALL** 해당 에이전트만 메모리에 로드해야 한다

**로딩 전략:**

```yaml
기본 로드 (1k tokens):
  - claude-code-manager: 기본 관리
  - general-purpose: 범용 작업

온디맨드 로드 (5k tokens 절약):
  언어별:
    - python-pro, typescript-pro: 언어 전문 작업시
    - rust-pro, golang-pro: 시스템 언어 작업시

  도메인별:
    - database-architect: DB 설계시
    - ui-ux-designer: 디자인 작업시
    - deployment-specialist: 배포 작업시

  특수 목적:
    - fact-checker: 검증 작업시
    - security-scanner: 보안 분석시
```

**로딩 성능:**
- 에이전트 로딩 지연: < 100ms
- 메모리 해제: 30분 미사용시 자동

**목표 감소량:** 6k → 1k tokens (83% 감소)

### FR-3: 스마트 메모리 파일 선택 @REQ:MEMORY-SELECT-001

**WHEN** 프로젝트 컨텍스트가 로드될 때
**THE SYSTEM SHALL** 현재 작업과 관련된 메모리 파일만 선택적으로 로드해야 한다

**현재 문제:**

```
전체 로딩 방식 (15.3k tokens):
├── CLAUDE.md (3.8k) - ✅ 항상 필요
├── project_guidelines.md (2.0k) - ✅ 항상 필요
├── product.md (2.0k) - 🔶 비전 관련 작업시만
├── structure.md (2.8k) - 🔶 아키텍처 작업시만
├── tech.md (2.8k) - 🔶 기술 결정시만
└── shared_checklists.md (0.5k) - 🔶 PR/테스트시만
└── 기타 파일들 (1.4k) - 🔴 특수 상황에만 필요
```

**스마트 로딩 알고리즘:**

```python
class SmartContextLoader:
    BASE_FILES = ["CLAUDE.md", "project_guidelines.md"]  # 5.8k

    def select_by_command(self, command: str) -> List[str]:
        if command.startswith("/moai:2-spec"):
            return self.BASE_FILES + ["product.md"]  # +2k = 7.8k
        elif command.startswith("/moai:3-plan"):
            return self.BASE_FILES + ["structure.md"]  # +2.8k = 8.6k
        elif command.startswith("/moai:5-dev"):
            return self.BASE_FILES + ["tech.md"]  # +2.8k = 8.6k
        else:
            return self.BASE_FILES  # 5.8k only

    def select_by_keywords(self, text: str) -> List[str]:
        """키워드 기반 스마트 선택"""
        additional = []
        if any(word in text.lower() for word in ["architecture", "design", "structure"]):
            additional.append("structure.md")
        if any(word in text.lower() for word in ["performance", "optimization", "tech"]):
            additional.append("tech.md")
        if any(word in text.lower() for word in ["vision", "product", "business"]):
            additional.append("product.md")
        return self.BASE_FILES + additional
```

**목표 감소량:** 15.3k → 5-8k tokens (50% 감소)

### FR-4: 동적 컨텍스트 관리 @REQ:DYNAMIC-CONTEXT-001

**WHILE** 사용자가 작업을 진행하는 동안
**THE SYSTEM SHALL** 컨텍스트 사용량을 실시간으로 모니터링하고 최적화해야 한다

**실시간 최적화:**

```python
class ContextManager:
    def __init__(self, max_tokens: int = 200000):
        self.max_tokens = max_tokens
        self.warning_threshold = 0.8  # 80%
        self.critical_threshold = 0.9  # 90%

    def monitor_usage(self) -> ContextAlert:
        current_usage = self.calculate_current_usage()
        usage_ratio = current_usage / self.max_tokens

        if usage_ratio > self.critical_threshold:
            return ContextAlert.CRITICAL_CLEANUP_NEEDED
        elif usage_ratio > self.warning_threshold:
            return ContextAlert.OPTIMIZATION_SUGGESTED
        else:
            return ContextAlert.NORMAL

    def suggest_optimizations(self) -> List[str]:
        """최적화 제안 생성"""
        suggestions = []
        if self.has_unused_agents():
            suggestions.append("Unload unused agents")
        if self.has_old_context():
            suggestions.append("Clear old conversation context")
        if self.has_large_files():
            suggestions.append("Remove large temporary files")
        return suggestions
```

## 📊 비기능 요구사항

### NFR-1: 성능 목표

- **시작 시 컨텍스트**: < 20k tokens (10%)
- **에이전트 로딩 지연**: < 100ms
- **메모리 파일 선택 정확도**: > 90%
- **기능 손실**: 0% (모든 기능 유지)

### NFR-2: 사용자 경험

- **투명성**: 최적화 과정 사용자에게 표시
- **제어권**: 사용자가 수동으로 선택 가능
- **복구성**: 필요시 전체 컨텍스트 복원 가능

### NFR-3: 호환성

- **기존 명령어**: 모든 /moai:* 명령어 정상 작동
- **에이전트 API**: 기존 에이전트 인터페이스 호환
- **메모리 시스템**: 기존 @import 구조 유지

## 🔄 사용자 여정 시나리오

### 시나리오 1: 최적화된 프로젝트 시작

```gherkin
GIVEN 사용자가 MoAI-ADK 프로젝트를 처음 열 때
WHEN Claude Code가 초기화될 때
THEN 컨텍스트 사용량이 15k tokens (7.5%) 이하로 표시되고
  AND 필수 도구와 기본 에이전트만 로드되고
  AND "Context optimized for MoAI-ADK" 알림이 표시되고
  AND 실제 작업 공간이 180k tokens 이상 확보된다
```

### 시나리오 2: 에이전트 동적 로딩

```gherkin
GIVEN 기본 컨텍스트가 로드된 상태에서
WHEN 사용자가 "Python 코드 최적화 도움" 요청을 할 때
THEN python-pro 에이전트가 100ms 이내에 로드되고
  AND 컨텍스트 사용량이 2k tokens 증가하고
  AND "Loading python-pro specialist..." 진행 표시가 나타나고
  AND 로딩 완료 후 전문적인 Python 지원이 제공된다
```

### 시나리오 3: 컨텍스트 자동 최적화

```gherkin
GIVEN 장시간 작업으로 컨텍스트 사용량이 80%에 도달했을 때
WHEN 시스템이 최적화 필요성을 감지할 때
THEN "Context optimization recommended" 알림이 표시되고
  AND 사용하지 않는 에이전트와 파일 목록이 제공되고
  AND "One-click optimize" 버튼이 표시되고
  AND 클릭 시 안전하게 컨텍스트가 정리된다
```

### 시나리오 4: 메모리 파일 스마트 선택

```gherkin
GIVEN 사용자가 "/moai:3-plan SPEC-001" 명령을 실행할 때
WHEN 시스템이 계획 수립 작업임을 감지할 때
THEN 기본 파일(5.8k) + structure.md(2.8k)만 로드되고
  AND "Loaded architecture context for planning" 메시지가 표시되고
  AND 불필요한 tech.md, product.md는 로드되지 않고
  AND 총 8.6k tokens만 사용하여 효율적인 작업 환경을 제공한다
```

## ✅ 수락 기준

### AC-1: 도구 설명 압축

```
✅ System tools 토큰 사용량 26.9k → 10k 이하
✅ 핵심 기능 설명 유지 (파라미터 타입, 필수 여부)
✅ 상세 설명은 /help 명령어로 별도 조회 가능
✅ 기존 도구 호출 방식 100% 호환
✅ 압축률 60% 이상 달성
```

### AC-2: 에이전트 지연 로딩

```
✅ 시작 시 필수 에이전트 2개만 로드 (1k tokens)
✅ 특정 에이전트 호출 시 100ms 이내 로드
✅ 30분 미사용 에이전트 자동 언로드
✅ 로딩/언로딩 상태 실시간 표시
✅ 에이전트 기능 손실 없음
```

### AC-3: 메모리 파일 스마트 선택

```
✅ 명령어별 관련 파일만 선택적 로드
✅ 키워드 기반 추가 파일 선택 정확도 90% 이상
✅ 기본 파일(CLAUDE.md, guidelines.md) 항상 로드
✅ 파일 선택 근거 사용자에게 설명
✅ 수동 오버라이드 옵션 제공
```

### AC-4: 동적 컨텍스트 관리

```
✅ 실시간 컨텍스트 사용량 모니터링
✅ 80% 도달 시 최적화 제안 표시
✅ 90% 도달 시 자동 정리 옵션 제공
✅ 원클릭 최적화 기능 제공
✅ 최적화 전후 사용량 비교 표시
```

### AC-5: 전체 성능 목표

```
✅ 시작 시 총 컨텍스트 < 20k tokens (10%)
✅ 모든 기존 기능 정상 작동
✅ 사용자 워크플로우 변경 최소화
✅ 50% 이상 토큰 절약 달성
✅ 기능별 성능 저하 없음
```

## 🔧 기술 구현 요구사항

### 1. 토큰 사용량 추적 시스템

```python
from dataclasses import dataclass
from typing import Dict, List

@dataclass
class TokenUsage:
    component: str
    tokens: int
    percentage: float
    is_essential: bool

class TokenTracker:
    def __init__(self, max_tokens: int = 200000):
        self.max_tokens = max_tokens
        self.components: Dict[str, TokenUsage] = {}

    def track_component(self, name: str, tokens: int, essential: bool = False):
        percentage = (tokens / self.max_tokens) * 100
        self.components[name] = TokenUsage(name, tokens, percentage, essential)

    def get_optimization_candidates(self) -> List[TokenUsage]:
        """최적화 가능한 컴포넌트 반환"""
        return [usage for usage in self.components.values()
                if not usage.is_essential and usage.tokens > 1000]

    def calculate_total_usage(self) -> int:
        return sum(usage.tokens for usage in self.components.values())
```

### 2. 동적 로더 시스템

```python
import asyncio
from typing import Optional
from datetime import datetime, timedelta

class DynamicAgentLoader:
    def __init__(self):
        self.loaded_agents = {}
        self.last_used = {}
        self.loading_cache = {}

    async def load_agent_on_demand(self, agent_name: str) -> Optional[object]:
        """에이전트 온디맨드 로딩"""
        if agent_name in self.loaded_agents:
            self.last_used[agent_name] = datetime.now()
            return self.loaded_agents[agent_name]

        # 로딩 중 중복 요청 방지
        if agent_name in self.loading_cache:
            return await self.loading_cache[agent_name]

        # 비동기 로딩 시작
        loading_task = asyncio.create_task(self._load_agent(agent_name))
        self.loading_cache[agent_name] = loading_task

        try:
            agent = await loading_task
            self.loaded_agents[agent_name] = agent
            self.last_used[agent_name] = datetime.now()
            return agent
        finally:
            del self.loading_cache[agent_name]

    def cleanup_unused_agents(self, threshold_minutes: int = 30) -> List[str]:
        """미사용 에이전트 정리"""
        threshold_time = datetime.now() - timedelta(minutes=threshold_minutes)
        unloaded = []

        for agent_name, last_used in list(self.last_used.items()):
            if last_used < threshold_time:
                del self.loaded_agents[agent_name]
                del self.last_used[agent_name]
                unloaded.append(agent_name)

        return unloaded
```

### 3. 스마트 컨텍스트 선택기

```python
import re
from typing import Set

class SmartContextSelector:
    def __init__(self):
        self.base_files = ["CLAUDE.md", "project_guidelines.md"]
        self.keyword_mapping = {
            "architecture": ["structure.md"],
            "design": ["structure.md"],
            "performance": ["tech.md"],
            "optimization": ["tech.md"],
            "vision": ["product.md"],
            "business": ["product.md"],
            "testing": ["shared_checklists.md"],
            "security": ["shared_checklists.md"]
        }

    def select_files_by_command(self, command: str) -> List[str]:
        """명령어 기반 파일 선택"""
        files = self.base_files.copy()

        if re.match(r'/moai:[2]', command):  # spec
            files.append("product.md")
        elif re.match(r'/moai:[3]', command):  # plan
            files.append("structure.md")
        elif re.match(r'/moai:[45]', command):  # tasks/dev
            files.append("tech.md")

        return files

    def select_files_by_content(self, text: str) -> List[str]:
        """내용 기반 파일 선택"""
        files = self.base_files.copy()
        text_lower = text.lower()

        for keyword, file_list in self.keyword_mapping.items():
            if keyword in text_lower:
                files.extend(file_list)

        return list(set(files))  # 중복 제거

    def explain_selection(self, selected_files: List[str], reason: str) -> str:
        """파일 선택 근거 설명"""
        base_msg = f"Selected {len(selected_files)} files for {reason}:"
        file_list = "\n".join(f"  • {file}" for file in selected_files)
        return f"{base_msg}\n{file_list}"
```

## 📈 성능 목표 및 측정

### 기준 성능 (Before) vs 목표 성능 (After)

| 구분 | Before | After | 개선율 |
|------|--------|--------|--------|
| **전체 토큰 사용량** | 55k (28%) | 20k (10%) | 64% ↓ |
| **System tools** | 26.9k | 10k | 62% ↓ |
| **Memory files** | 15.3k | 5-8k | 50% ↓ |
| **Custom agents** | 6.0k | 1k | 83% ↓ |
| **사용 가능 공간** | 145k (72%) | 180k (90%) | 24% ↑ |

### 성능 벤치마크 스크립트

```python
# scripts/context_benchmark.py
import time
from dataclasses import dataclass

@dataclass
class BenchmarkResult:
    startup_time: float
    initial_tokens: int
    agent_load_time: float
    memory_selection_accuracy: float

class ContextBenchmark:
    def measure_startup_performance(self) -> BenchmarkResult:
        """시작 성능 측정"""
        start_time = time.perf_counter()

        # Claude Code 초기화 시뮬레이션
        initial_tokens = self.measure_initial_context()
        startup_time = time.perf_counter() - start_time

        # 에이전트 로딩 시간 측정
        agent_start = time.perf_counter()
        self.load_test_agent()
        agent_load_time = time.perf_counter() - agent_start

        # 메모리 선택 정확도 측정
        accuracy = self.test_memory_selection_accuracy()

        return BenchmarkResult(
            startup_time=startup_time,
            initial_tokens=initial_tokens,
            agent_load_time=agent_load_time,
            memory_selection_accuracy=accuracy
        )

    def validate_performance_targets(self) -> Dict[str, bool]:
        """성능 목표 달성 여부 검증"""
        result = self.measure_startup_performance()

        return {
            "startup_tokens_under_20k": result.initial_tokens < 20000,
            "agent_load_under_100ms": result.agent_load_time < 0.1,
            "memory_accuracy_over_90": result.memory_selection_accuracy > 0.9,
            "total_startup_under_1s": result.startup_time < 1.0
        }
```

## 🔗 연관 태그

- **@DESIGN:TOKEN-MANAGER** → 토큰 관리 시스템 설계
- **@DESIGN:CONTEXT-OPTIMIZER** → 컨텍스트 최적화 엔진
- **@TASK:TOOL-COMPRESSION** → 도구 설명 압축 구현
- **@TASK:LAZY-LOADING** → 에이전트 지연 로딩 구현
- **@TASK:MEMORY-SELECTION** → 스마트 메모리 선택 구현
- **@TEST:PERFORMANCE-CONTEXT** → 컨텍스트 성능 테스트

---

## ✅ 명확화 완료 항목

### 1. 도구 설명 압축 우선순위 기준

**압축 우선순위 (높음 → 낮음):**

1. **중복 설명 제거 (5k tokens 절약)**
   - 동일한 파라미터 타입 설명 반복 → 공통 참조로 교체
   - 유사한 도구 간 중복 예시 → 대표 예시 1개만 유지
   - 반복되는 에러 케이스 설명 → 공통 에러 가이드로 통합

2. **예시 코드 블록 축소 (8k tokens 절약)**
   - 긴 예시 코드 → 핵심 라인만 유지 (5줄 이하)
   - 다중 예시 → 가장 일반적인 사용법 1개만 유지
   - 상세 코멘트 → 간단 설명으로 교체

3. **상세 설명 링크화 (4k tokens 절약)**
   - 장문 사용법 설명 → "자세한 내용: /help [도구명]"으로 링크
   - 주의사항과 Best Practice → 별도 문서화
   - 고급 사용법 → 온디맨드 로딩

**필수 유지 요소:**
- 파라미터 이름, 타입, 필수/선택 여부
- 도구의 핵심 목적 (1줄 설명)
- 필수 파라미터의 기본 예시값
- 치명적 에러를 방지하는 경고사항

**압축 적용 우선순위:**
1. Task, Agent 관련 도구 (사용 빈도 높음, 설명 길음)
2. 파일 조작 도구 (Read, Write, Edit 등)
3. 검색 도구 (Grep, Glob 등)
4. 시스템 도구 (Bash, WebFetch 등)

### 2. 에이전트 로딩 우선순위 기준

**기본 로드 에이전트 선택 기준:**

1. **claude-code-manager** (필수 기본 로드)
   - **사용 빈도**: 모든 MoAI 명령어에서 참조됨 (100%)
   - **기능 중요도**: Claude Code 설정 최적화 전담 (CRITICAL)
   - **의존성**: 다른 에이전트들의 권한 관리 담당
   - **토큰 효율성**: 400 tokens (경량)

2. **general-purpose** (필수 기본 로드)
   - **사용 빈도**: 범용 작업, 검색, 탐색 (80% 이상)
   - **기능 중요도**: 모든 도구에 접근 가능한 유일한 에이전트
   - **의존성**: 다른 전문 에이전트 호출 전 초기 분석 담당
   - **토큰 효율성**: 600 tokens (중간)

**지연 로드 에이전트 분류:**

**Tier 1: 고빈도 (10분 캐시)**
- `spec-manager`, `code-generator`, `test-automator`
- 일일 사용 빈도 > 50%, 개발 핵심 워크플로우

**Tier 2: 중빈도 (30분 캐시)**
- `python-pro`, `typescript-pro`, `frontend-developer`
- 언어/도메인별 전문성, 프로젝트 타입에 따라 집중 사용

**Tier 3: 저빈도 (60분 캐시)**
- `deployment-specialist`, `database-architect`, `ui-ux-designer`
- 특정 단계나 역할에서만 집중 사용

**Tier 4: 온디맨드 (캐시 없음)**
- `fact-checker`, `c-sharp-pro`, `swift-pro`
- 특수 상황이나 특정 기술 스택에서만 사용

**선택 기준 가중치:**
- 사용 빈도: 40%
- 의존성 중요도: 30%
- 토큰 효율성: 20%
- 기능 대체 가능성: 10%

### 3. 메모리 파일 선택 알고리즘 세부사항

**정확도 측정 방법:**

**테스트 케이스 분류 (총 100개):**

1. **명령어 기반 테스트 (40개)**
   ```yaml
   /moai:2-spec: product.md 포함 여부 (10개)
   /moai:3-plan: structure.md 포함 여부 (10개)
   /moai:4-tasks: tech.md 포함 여부 (10개)
   /moai:5-dev: tech.md + shared_checklists.md 포함 여부 (10개)
   ```

2. **키워드 기반 테스트 (40개)**
   ```yaml
   "architecture, design": structure.md 선택 (10개)
   "performance, optimization": tech.md 선택 (10개)
   "vision, business, product": product.md 선택 (10개)
   "testing, security": shared_checklists.md 선택 (10개)
   ```

3. **복합 시나리오 테스트 (20개)**
   ```yaml
   다중 키워드: "UI design + performance" → structure.md + tech.md
   모호한 요청: "프로젝트 개선" → 기본 파일만
   전체 컨텍스트: "전체 검토" → 모든 파일
   특수 케이스: 새로운 키워드 패턴
   ```

**정확도 계산 공식:**
```python
def calculate_accuracy(test_results):
    correct_selections = 0
    total_tests = len(test_results)

    for test in test_results:
        expected_files = test.expected_files
        selected_files = test.actual_files

        # 필수 파일이 모두 포함되고, 불필요한 파일이 없으면 정확
        if (set(expected_files).issubset(set(selected_files)) and
            len(selected_files) <= len(expected_files) + 1):  # 1개 여유 허용
            correct_selections += 1

    return (correct_selections / total_tests) * 100
```

**평가 기준:**
- **Perfect Match (100점)**: 예상 파일과 정확히 일치
- **Good Match (80점)**: 필수 파일 포함 + 1개 추가 파일
- **Acceptable Match (60점)**: 필수 파일 포함 + 2개 추가 파일
- **Poor Match (0점)**: 필수 파일 누락 또는 3개 이상 불필요 파일

**목표 성능:**
- 명령어 기반: 95% 정확도 (명확한 패턴)
- 키워드 기반: 90% 정확도 (추론 필요)
- 복합 시나리오: 85% 정확도 (복잡도 높음)
- **전체 평균: 90% 이상**

**자동 테스트 스크립트:**
```python
# tests/context_selection_test.py
class ContextSelectionTest:
    def run_accuracy_test(self) -> Dict[str, float]:
        command_tests = self.test_command_based_selection()
        keyword_tests = self.test_keyword_based_selection()
        complex_tests = self.test_complex_scenarios()

        return {
            "command_accuracy": self.calculate_accuracy(command_tests),
            "keyword_accuracy": self.calculate_accuracy(keyword_tests),
            "complex_accuracy": self.calculate_accuracy(complex_tests),
            "overall_accuracy": self.calculate_overall_accuracy()
        }
```

### 4. 사용자 경험 최적화 방안

**지연 최소화 전략:**

1. **백그라운드 프리로딩**
   ```python
   # 예측적 로딩 - 사용자가 인지하기 전에 미리 준비
   class PredictiveLoader:
       def __init__(self):
           self.usage_patterns = self.load_user_patterns()

       async def preload_likely_agents(self, command_pattern: str):
           # 과거 패턴 기반으로 다음 필요한 에이전트 예측
           likely_agents = self.predict_next_agents(command_pattern)
           for agent in likely_agents:
               asyncio.create_task(self.load_agent_background(agent))
   ```

2. **점진적 로딩 (Progressive Loading)**
   ```
   Phase 1 (0ms): 기본 인터페이스 표시
   Phase 2 (50ms): 필수 도구 + 기본 에이전트
   Phase 3 (100ms): 컨텍스트 파일 선택 완료
   Phase 4 (200ms): 예측된 에이전트 백그라운드 로딩
   ```

**실시간 피드백 시스템:**

1. **로딩 상태 표시**
   ```yaml
   최적화 시작:
     "🔧 Context optimizing... (2.1s estimated)"

   에이전트 로딩:
     "⚡ Loading python-pro specialist... ▓▓▓░░░ 60%"

   메모리 선택:
     "📚 Selecting relevant docs... ✓ 3 files selected"

   완료:
     "✅ Optimized: 55k→18k tokens (67% reduction)"
   ```

2. **진행률 표시 방식**
   ```
   최소 지연 (< 100ms): 스피너만 표시
   중간 지연 (100-500ms): 진행률 바 + 남은 시간
   긴 지연 (> 500ms): 단계별 설명 + 취소 옵션
   ```

**사용자 제어 옵션:**

1. **최적화 레벨 선택**
   ```
   🚀 Fast Mode: 기본 파일만 (5초 이내)
   ⚖️ Balanced: 스마트 선택 (10초 이내)
   🔬 Complete: 전체 컨텍스트 (15초 이내)
   ```

2. **수동 오버라이드**
   ```python
   # 사용자가 직접 조정 가능한 설정
   class UserContextPreferences:
       def __init__(self):
           self.always_load_files = ["CLAUDE.md", "project_guidelines.md"]
           self.never_load_files = ["legacy_docs.md"]
           self.preferred_agents = ["python-pro", "typescript-pro"]
           self.max_context_usage = 0.15  # 15%까지 허용
   ```

**투명성 및 설명 기능:**

1. **선택 근거 표시**
   ```
   💡 Why these files were selected:
   ├── CLAUDE.md (always required)
   ├── tech.md (detected: "performance optimization")
   └── structure.md (command: /moai:3-plan)

   📊 Context usage: 18.2k/200k tokens (9.1%)
   💾 Memory saved: 37k tokens (67% reduction)
   ```

2. **원클릭 설정 변경**
   ```
   현재 설정이 부족하다면:
   "Need more context? [🔧 Load additional files] [⚙️ Customize]"

   너무 많은 컨텍스트라면:
   "Want to optimize further? [⚡ Aggressive mode] [📝 Manual selection]"
   ```

**에러 복구 및 폴백:**

1. **자동 폴백 메커니즘**
   ```python
   class GracefulDegradation:
       def handle_optimization_failure(self, error):
           if error.type == "agent_load_timeout":
               return self.fallback_to_general_purpose()
           elif error.type == "memory_selection_error":
               return self.load_all_essential_files()
           else:
               return self.disable_optimization_temporarily()
   ```

2. **사용자 선택 옵션**
   ```
   ❌ Optimization failed (network timeout)

   Options:
   [🔄 Retry optimization]
   [⏭️ Skip and continue with full context]
   [⚙️ Manual selection mode]
   ```

**성능 모니터링 및 개선:**

1. **실시간 성능 지표**
   ```
   📈 Performance Dashboard:
   ├── Avg. startup time: 0.8s (target: <1s) ✅
   ├── Context accuracy: 92% (target: >90%) ✅
   ├── User satisfaction: 4.6/5.0 ✅
   └── Memory usage: 18k tokens (9% of limit) ✅
   ```

2. **적응형 학습**
   ```python
   class AdaptiveLearning:
       def learn_from_user_feedback(self, selection, user_action):
           """사용자의 수동 조정을 학습하여 다음 선택 개선"""
           if user_action == "added_file":
               self.increase_weight(selection.context, user_action.file)
           elif user_action == "removed_file":
               self.decrease_weight(selection.context, user_action.file)
   ```

---

> **@REQ:CONTEXT-MEMORY-001** 을 통해 이 컨텍스트 최적화 요구사항이 설계와 구현 단계로 추적됩니다.
>
> **Claude Code의 컨텍스트 효율성을 극대화하여 개발자가 실제 작업에 집중할 수 있는 환경을 제공합니다.**