# 구조 설계 @STRUCT:MOAI-ADK

> **@STRUCT:MOAI-ADK** "3계층 플러그인 아키텍처 기반 모듈형 설계"

## 🏗️ 전체 아키텍처

### 핵심 설계 원칙

1. **모듈형 설계**: 각 기능은 독립적인 모듈로 구현
2. **플러그인 아키텍처**: 확장 가능한 에이전트 시스템
3. **계층 분리**: CLI → Core → Plugins 3계층 구조
4. **의존성 역전**: 인터페이스 기반 확장성
5. **불변성 보장**: Git 기반 체크포인트 시스템

## 📁 디렉토리 구조

```
MoAI-ADK/
├── .claude/                    # Claude Code 표준 자산
│   ├── commands/moai/          # 6개 슬래시 명령어 (연번 순)
│   │   ├── 1_project.py        # 프로젝트 초기화
│   │   ├── 2_spec.py           # 명세 작성
│   │   ├── 3_plan.py           # 계획 수립
│   │   ├── 4_tasks.py          # 작업 분해
│   │   ├── 5_dev.py            # 코드 구현
│   │   └── 6_sync.py           # 문서 동기화
│   ├── agents/moai/            # 11개 전문 에이전트
│   │   ├── steering_architect.py      # 비전/아키텍처
│   │   ├── spec_manager.py           # EARS 명세 작성
│   │   ├── plan_architect.py         # Constitution 검증
│   │   ├── task_decomposer.py        # TDD 작업 분해
│   │   ├── code_generator.py         # Red-Green-Refactor
│   │   ├── test_automator.py         # 테스트 자동화
│   │   ├── doc_syncer.py             # 문서 동기화
│   │   ├── deployment_specialist.py  # 배포 관리
│   │   ├── integration_manager.py    # API 연동
│   │   ├── tag_indexer.py            # TAG 인덱싱
│   │   └── claude_code_manager.py    # Claude Code 최적화
│   ├── hooks/moai/             # 5개 Python Hook Scripts
│   │   ├── pre_tool_use.py     # Constitution 검증
│   │   ├── post_tool_use.py    # TAG 동기화
│   │   ├── session_start.py    # 프로젝트 상태 알림
│   │   ├── context_selector.py # Top-K 메모리 추천
│   │   └── quality_gate.py     # 품질 게이트 검증
│   ├── memory/                 # 공유 메모리 (링크만 유지)
│   │   ├── README.md           # 메모리 인덱스
│   │   ├── project_guidelines.md  # 프로젝트 가이드라인
│   │   └── shared_checklists.md   # 공통 체크리스트
│   └── settings.json           # 권한 및 Hook 설정
├── .moai/                      # MoAI 문서 시스템
│   ├── steering/               # 프로젝트 방향성 문서
│   │   ├── product.md          # 제품 비전 (@VISION)
│   │   ├── structure.md        # 구조 설계 (@STRUCT)
│   │   └── tech.md             # 기술 스택 (@TECH)
│   ├── specs/                  # SPEC 문서 (동적 생성)
│   │   ├── SPEC-001/
│   │   │   ├── spec.md         # EARS 형식 명세
│   │   │   ├── scenarios.md    # GWT 시나리오
│   │   │   └── acceptance.md   # 수락 기준
│   │   ├── SPEC-002/
│   │   └── SPEC-003/
│   ├── memory/                 # 프로젝트 메모리 (automatic)
│   │   ├── constitution.md     # 헌법 및 거버넌스
│   │   ├── operations.md       # 운영·협업 전문
│   │   ├── engineering-standards.md # 엔지니어링 표준
│   │   ├── common.md           # 공통 운영 체크
│   │   ├── backend-python.md   # Python 백엔드 스택
│   │   └── frontend-*.md       # 프론트엔드 스택별 메모
│   ├── scripts/                # 검증 스크립트
│   │   ├── check-traceability.py  # TAG 추적성 검증
│   │   ├── run-tests.sh           # 전체 테스트 실행
│   │   └── repair_tags.py         # TAG 복구 스크립트
│   ├── indexes/                # 인덱스 시스템
│   │   ├── tags.json           # 16-Core TAG 인덱스
│   │   ├── state.json          # 프로젝트 상태
│   │   └── version.json        # 버전 정보
│   └── config.json             # MoAI 설정
├── src/                        # 소스 코드 (동적 생성)
│   ├── core/                   # 핵심 모듈
│   ├── agents/                 # 에이전트 구현
│   ├── plugins/                # 플러그인 시스템
│   └── utils/                  # 유틸리티
├── tests/                      # 테스트 코드
│   ├── unit/                   # 단위 테스트
│   ├── integration/            # 통합 테스트
│   └── e2e/                    # E2E 테스트
├── docs/                       # 문서 (자동 생성)
│   ├── api/                    # API 문서
│   ├── guides/                 # 사용 가이드
│   └── examples/               # 예제 코드
├── CLAUDE.md                   # 프로젝트 메모리 허브
├── pyproject.toml              # Python 프로젝트 설정
├── README.md                   # 프로젝트 README
└── .gitignore                  # Git 무시 파일
```

## 🔧 3계층 아키텍처

### Layer 1: CLI Interface
```python
# .claude/commands/moai/
class MoAICommand:
    """슬래시 명령어 인터페이스"""
    def execute(self, args: List[str]) -> Result
    def validate_preconditions(self) -> bool
    def post_execution_hooks(self) -> None
```

**책임:**
- 사용자 입력 파싱 및 검증
- 에이전트 선택 및 호출
- 결과 포맷팅 및 출력

### Layer 2: Core Engine
```python
# src/core/
class AgentOrchestrator:
    """에이전트 오케스트레이션"""
    def dispatch_agent(self, agent_type: str, task: Task) -> Result
    def validate_constitution(self, task: Task) -> bool
    def track_tag_changes(self, changes: List[Change]) -> None

class TagSystem:
    """16-Core TAG 관리"""
    def index_tags(self, files: List[Path]) -> TagIndex
    def validate_traceability(self) -> TraceabilityReport
    def repair_broken_links(self) -> RepairResult
```

**책임:**
- 에이전트 라이프사이클 관리
- TAG 시스템 운영
- Constitution 검증
- Git 체크포인트 관리

### Layer 3: Agent Plugins
```python
# .claude/agents/moai/
class BaseAgent:
    """에이전트 베이스 클래스"""
    def execute(self, task: Task) -> Result
    def validate_input(self, input: Input) -> bool
    def generate_output(self, result: Result) -> Output

class SpecManager(BaseAgent):
    """EARS 명세 작성 전문"""
    def create_ears_spec(self, requirements: str) -> Spec
    def validate_completeness(self, spec: Spec) -> bool
```

**책임:**
- 특화된 작업 수행
- 도메인별 검증 로직
- 표준 출력 형식 생성

## 📊 모듈 설계 원칙

### 1. 단일 책임 원칙 (SRP)
- **각 모듈은 하나의 책임만** 가짐
- **최대 300 LOC** 제한 (Constitution 원칙)
- **함수당 최대 50 LOC** 제한

### 2. 개방-폐쇄 원칙 (OCP)
- **인터페이스를 통한 확장** 가능
- **기존 코드 수정 없이** 새 에이전트 추가
- **플러그인 시스템**으로 기능 확장

### 3. 의존성 역전 원칙 (DIP)
```python
# 올바른 의존성 구조
class AgentInterface(ABC):
    @abstractmethod
    def execute(self, task: Task) -> Result: ...

class SpecManager(AgentInterface):
    def execute(self, task: Task) -> Result:
        # EARS 명세 생성 로직
        pass

class Orchestrator:
    def __init__(self, agents: Dict[str, AgentInterface]):
        self._agents = agents  # 인터페이스에 의존
```

### 4. 인터페이스 분리 원칙 (ISP)
```python
# 각 역할별로 분리된 인터페이스
class Validator(ABC):
    @abstractmethod
    def validate(self, data: Any) -> bool: ...

class Generator(ABC):
    @abstractmethod
    def generate(self, input: Any) -> Any: ...

class Tracker(ABC):
    @abstractmethod
    def track(self, changes: List[Change]) -> None: ...
```

## 🔄 데이터 흐름 설계

### 1. 명령어 실행 흐름
```
사용자 입력 → CLI 파싱 → 전처리 Hook → 에이전트 선택 → 작업 실행 → 후처리 Hook → 결과 출력
```

### 2. TAG 추적 흐름
```
파일 변경 감지 → TAG 추출 → 인덱스 업데이트 → 추적성 검증 → 의존성 갱신
```

### 3. Constitution 검증 흐름
```
작업 요청 → 5원칙 검증 → 위반 사항 차단 → 승인된 작업 실행 → 사후 검증
```

## 🛡️ 품질 보장 메커니즘

### 1. 자동화된 검증
- **Pre-commit hooks**: 코드 품질 검사
- **Constitution gates**: 5원칙 자동 검증
- **TAG validation**: 추적성 체인 검증
- **Test coverage**: 최소 85% 커버리지

### 2. 실시간 모니터링
```python
class QualityMonitor:
    def check_complexity(self, module: Path) -> ComplexityReport
    def validate_dependencies(self) -> DependencyReport
    def monitor_performance(self) -> PerformanceMetrics
```

### 3. 자동 복구 시스템
```python
class AutoRepair:
    def repair_broken_tags(self) -> RepairResult
    def fix_dependency_cycles(self) -> FixResult
    def restore_checkpoint(self, commit_hash: str) -> RestoreResult
```

## 🔧 확장성 설계

### 1. 새로운 에이전트 추가
```python
# 1. 인터페이스 구현
class NewAgent(BaseAgent):
    def execute(self, task: Task) -> Result:
        # 새로운 기능 구현
        pass

# 2. 등록
agent_registry.register("new-agent", NewAgent)

# 3. 명령어에서 사용
/moai:custom new-agent "작업 내용"
```

### 2. 새로운 TAG 카테고리 추가
```json
{
  "tag_system": {
    "categories": {
      "CUSTOM": ["NEW1", "NEW2", "NEW3"]
    }
  }
}
```

### 3. 새로운 Hook 추가
```python
class CustomHook:
    def on_file_change(self, file_path: Path) -> None:
        # 커스텀 로직
        pass
```

## 📏 성능 및 확장성 목표

### 1. 성능 목표
- **명령어 응답 시간**: < 2초
- **파일 변경 감지**: < 100ms
- **TAG 인덱싱**: < 500ms
- **Constitution 검증**: < 200ms

### 2. 확장성 목표
- **최대 프로젝트 크기**: 10,000 파일
- **최대 TAG 수**: 50,000개
- **동시 에이전트 실행**: 5개
- **메모리 사용량**: < 500MB

### 3. 안정성 목표
- **에이전트 성공률**: > 95%
- **TAG 정확성**: > 98%
- **Git 체크포인트**: 100% 신뢰성
- **롤백 성공률**: 100%

---

> **@STRUCT:MOAI-ADK** 태그를 통해 이 구조 설계가 프로젝트 전체에 일관되게 적용됩니다.
>
> **모든 모듈은 Constitution 5원칙을 준수하며, 플러그인 아키텍처를 통해 무한 확장 가능합니다.**