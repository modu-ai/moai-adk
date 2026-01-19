# SPEC-RALPH-001: MoAI Ralph Engine

## TAG BLOCK

```yaml
SPEC-ID: SPEC-RALPH-001
Title: MoAI Ralph Engine - LSP + AST-grep + Loop 통합 시스템
Created: 2026-01-09
Status: Completed
Priority: HIGH
Assigned: manager-ddd
Related-SPECs: []
Epic: MoAI-ADK Core Enhancement
Labels: [lsp, ast-grep, automation, feedback-loop, code-quality]
Lifecycle: spec-anchored
Completed: 2026-01-09
```

---

## 1. Environment (환경)

### 1.1 시스템 개요

MoAI Ralph Engine은 다음 세 가지 핵심 기술을 통합하는 지능형 코드 품질 보증 시스템입니다:

1. **LSP (Language Server Protocol)**: 16개 언어에 대한 실시간 진단 및 코드 인텔리전스
2. **AST-grep**: 구조적 패턴 매칭 및 보안 스캐닝
3. **Ralph Loop**: Ralph Wiggum 플러그인에서 영감을 받은 자율 피드백 루프

### 1.2 기존 자산

| 자산 | 위치 | 설명 |
|------|------|------|
| `.lsp.json` | 프로젝트 루트 | 16개 언어 서버 설정 (Python, TypeScript, Go, Rust, Java 등) |
| `post_tool__ast_grep_scan.py` | `.claude/hooks/moai/` | Write/Edit 후 AST-grep 보안 스캔 훅 |
| `moai-tool-ast-grep` | `.claude/skills/` | AST-grep 패턴 검색 스킬 |

### 1.3 지원 언어

```
Python (pyright-langserver)     | TypeScript/JavaScript (typescript-language-server)
Go (gopls)                       | Rust (rust-analyzer)
Java (jdtls)                     | Kotlin (kotlin-language-server)
Swift (sourcekit-lsp)            | C# (OmniSharp)
C/C++ (clangd)                   | Ruby (solargraph)
PHP (intelephense)               | Elixir (elixir-ls)
Scala (metals)                   | R (languageserver)
Dart (dart language-server)      |
```

### 1.4 운영 환경

- **Python**: 3.13+
- **의존성 관리**: uv, poetry
- **테스트 프레임워크**: pytest, pytest-asyncio
- **타입 체킹**: mypy, pyright
- **코드 품질**: ruff, black

---

## 2. Assumptions (가정)

### 2.1 기술적 가정

| ID | 가정 | 신뢰도 | 근거 | 실패 시 영향 | 검증 방법 |
|----|------|--------|------|-------------|----------|
| A-01 | LSP 서버들이 JSON-RPC 2.0 프로토콜을 준수함 | HIGH | LSP 공식 스펙 | 통신 실패 | 프로토콜 핸드셰이크 테스트 |
| A-02 | AST-grep CLI (`sg`)가 시스템에 설치됨 | MEDIUM | 선택적 기능 | 스캔 기능 비활성화 | `shutil.which("sg")` 체크 |
| A-03 | `.lsp.json` 설정이 유효함 | HIGH | 기존 파일 존재 | LSP 초기화 실패 | JSON 스키마 검증 |
| A-04 | 비동기 처리가 필요함 (LSP 응답 대기) | HIGH | LSP 프로토콜 특성 | 성능 저하 | asyncio 패턴 적용 |

### 2.2 비즈니스 가정

| ID | 가정 | 신뢰도 | 근거 | 실패 시 영향 |
|----|------|--------|------|-------------|
| B-01 | 사용자가 브랜치/PR 생성을 선택적으로 원함 | HIGH | 사용자 요구사항 | 워크플로우 강제 |
| B-02 | 자동화 루프는 사용자 취소 가능해야 함 | HIGH | 안전성 요구 | 무한 루프 위험 |
| B-03 | 현재 브랜치에서 작업 가능해야 함 | HIGH | 사용자 요구사항 | 불필요한 브랜치 생성 |

### 2.3 통합 가정

| ID | 가정 | 신뢰도 | 근거 |
|----|------|--------|------|
| I-01 | Claude Code hooks 시스템과 호환됨 | HIGH | 기존 훅 작동 중 |
| I-02 | MoAI-ADK 명령 체계와 통합됨 | HIGH | 기존 `/moai:*` 명령 |
| I-03 | 기존 AST-grep 훅과 충돌하지 않음 | MEDIUM | 훅 우선순위 관리 필요 |

---

## 3. Requirements (요구사항)

### 3.1 LSP 통합 레이어 요구사항

#### REQ-LSP-001: LSP 클라이언트 기본 기능
- **EARS Pattern**: Event-Driven
- **WHEN** 파일이 편집되면 **THEN** 시스템은 해당 언어의 LSP 서버에서 진단 정보를 가져와야 한다
- **Priority**: HIGH
- **Rationale**: 실시간 코드 품질 피드백 제공

#### REQ-LSP-002: LSP 서버 생명주기 관리
- **EARS Pattern**: State-Driven
- **IF** 특정 언어 파일이 열리면 **THEN** 해당 LSP 서버가 자동으로 시작되어야 한다
- **IF** 해당 언어 파일이 모두 닫히면 **THEN** LSP 서버가 종료되어야 한다
- **Priority**: HIGH
- **Rationale**: 리소스 효율적 관리

#### REQ-LSP-003: 심볼 참조 검색
- **EARS Pattern**: Event-Driven
- **WHEN** 심볼 참조 검색이 요청되면 **THEN** 시스템은 프로젝트 전체에서 모든 참조를 찾아 반환해야 한다
- **Priority**: MEDIUM
- **Rationale**: 리팩토링 지원

#### REQ-LSP-004: 안전한 심볼 이름 변경
- **EARS Pattern**: Event-Driven
- **WHEN** 심볼 이름 변경이 요청되면 **THEN** 시스템은 모든 참조를 안전하게 업데이트해야 한다
- **Priority**: MEDIUM
- **Rationale**: 자동 리팩토링 지원

#### REQ-LSP-005: 타입 및 문서 정보 조회
- **EARS Pattern**: Event-Driven
- **WHEN** 심볼에 대한 호버 정보가 요청되면 **THEN** 시스템은 타입과 문서 정보를 반환해야 한다
- **Priority**: LOW
- **Rationale**: 코드 이해도 향상

### 3.2 AST-grep 강화 레이어 요구사항

#### REQ-AST-001: 단일 파일 스캔
- **EARS Pattern**: Event-Driven
- **WHEN** 파일 스캔이 요청되면 **THEN** 시스템은 보안 및 품질 규칙을 적용하여 스캔해야 한다
- **Priority**: HIGH
- **Rationale**: 실시간 보안 검사

#### REQ-AST-002: 프로젝트 전체 스캔
- **EARS Pattern**: Event-Driven
- **WHEN** 프로젝트 스캔이 요청되면 **THEN** 시스템은 모든 지원 파일을 스캔하고 통합 리포트를 생성해야 한다
- **Priority**: HIGH
- **Rationale**: 전체 코드베이스 품질 평가

#### REQ-AST-003: 커스텀 패턴 검색
- **EARS Pattern**: Event-Driven
- **WHEN** 사용자 정의 패턴이 제공되면 **THEN** 시스템은 해당 패턴과 일치하는 모든 코드를 찾아야 한다
- **Priority**: MEDIUM
- **Rationale**: 유연한 코드 검색

#### REQ-AST-004: 배치 패턴 변환
- **EARS Pattern**: Event-Driven
- **WHEN** 패턴 변환이 요청되면 **THEN** 시스템은 일치하는 모든 코드를 안전하게 변환해야 한다
- **IF** 변환 중 오류가 발생하면 **THEN** 시스템은 모든 변경을 롤백해야 한다
- **Priority**: MEDIUM
- **Rationale**: 대규모 리팩토링 지원

### 3.3 루프 컨트롤러 요구사항

#### REQ-LOOP-001: 루프 초기화
- **EARS Pattern**: Event-Driven
- **WHEN** 루프가 시작되면 **THEN** 시스템은 현재 상태를 기록하고 완료 조건을 설정해야 한다
- **Priority**: HIGH
- **Rationale**: 루프 추적 및 관리

#### REQ-LOOP-002: 완료 조건 검사
- **EARS Pattern**: Ubiquitous
- 시스템은 **항상** 각 루프 반복 후 완료 조건을 검사해야 한다
- **Priority**: HIGH
- **Rationale**: 루프 종료 보장

#### REQ-LOOP-003: LSP+AST 피드백 실행
- **EARS Pattern**: State-Driven
- **IF** 루프가 활성 상태이면 **THEN** 시스템은 LSP 진단과 AST-grep 스캔을 실행하고 결과를 Claude에 피드백해야 한다
- **Priority**: HIGH
- **Rationale**: 자동 품질 개선

#### REQ-LOOP-004: 루프 취소
- **EARS Pattern**: Event-Driven
- **WHEN** 사용자가 루프 취소를 요청하면 **THEN** 시스템은 즉시 현재 루프를 중단하고 상태를 정리해야 한다
- **Priority**: HIGH
- **Rationale**: 사용자 제어권 보장

#### REQ-LOOP-005: 최대 반복 제한
- **EARS Pattern**: Unwanted
- 시스템은 설정된 최대 반복 횟수(기본값: 10)를 **초과하지 않아야 한다**
- **Priority**: HIGH
- **Rationale**: 무한 루프 방지

#### REQ-LOOP-006: 진행 상태 보고
- **EARS Pattern**: Ubiquitous
- 시스템은 **항상** 현재 루프 상태(반복 횟수, 남은 이슈, 예상 완료)를 보고해야 한다
- **Priority**: MEDIUM
- **Rationale**: 사용자 가시성

### 3.4 명령어 요구사항

#### REQ-CMD-001: /alfred 명령
- **EARS Pattern**: Event-Driven
- **WHEN** `/alfred` 명령이 실행되면 **THEN** 시스템은 1-plan, 2-run, 3-sync를 순차적으로 실행해야 한다
- **IF** 브랜치/PR 설정이 비활성화되어 있으면 **THEN** 현재 브랜치에서 작업해야 한다
- **Priority**: HIGH
- **Rationale**: 원클릭 자동화

#### REQ-CMD-002: /moai-loop 명령
- **EARS Pattern**: Event-Driven
- **WHEN** `/moai-loop` 명령이 실행되면 **THEN** 시스템은 Ralph-style 피드백 루프를 시작해야 한다
- **Priority**: HIGH
- **Rationale**: 자율 개선 루프

#### REQ-CMD-003: /moai-fix 명령
- **EARS Pattern**: Event-Driven
- **WHEN** `/moai-fix` 명령이 실행되면 **THEN** 시스템은 현재 LSP 오류와 AST 경고를 자동으로 수정해야 한다
- **Priority**: MEDIUM
- **Rationale**: 빠른 문제 해결

#### REQ-CMD-004: /cancel-loop 명령
- **EARS Pattern**: Event-Driven
- **WHEN** `/cancel-loop` 명령이 실행되면 **THEN** 시스템은 활성 루프를 취소해야 한다
- **Priority**: HIGH
- **Rationale**: 사용자 제어

### 3.5 훅 요구사항

#### REQ-HOOK-001: PostToolUse LSP 진단
- **EARS Pattern**: Event-Driven
- **WHEN** Write/Edit 작업이 완료되면 **THEN** LSP 진단을 실행하고 결과를 Claude에 제공해야 한다
- **Priority**: HIGH
- **Rationale**: 실시간 피드백

#### REQ-HOOK-002: Stop 훅 루프 컨트롤
- **EARS Pattern**: Event-Driven
- **WHEN** Claude 응답이 완료되면 **THEN** 루프 컨트롤러가 완료 조건을 검사하고 필요 시 추가 작업을 요청해야 한다
- **Priority**: HIGH
- **Rationale**: Ralph-style 자율 루프

---

## 4. Specifications (명세)

### 4.1 아키텍처 개요

```
┌─────────────────────────────────────────────────────────────────────┐
│                        MoAI Ralph Engine                            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐    │
│  │   LSP Layer     │  │  AST-grep Layer │  │   Loop Layer    │    │
│  │                 │  │                 │  │                 │    │
│  │ - client.py     │  │ - analyzer.py   │  │ - controller.py │    │
│  │ - server_mgr.py │  │                 │  │ - state.py      │    │
│  │ - protocol.py   │  │                 │  │ - feedback.py   │    │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘    │
│           │                    │                    │              │
│           └────────────────────┼────────────────────┘              │
│                                │                                   │
│                    ┌───────────▼───────────┐                       │
│                    │   Integration Layer   │                       │
│                    │                       │                       │
│                    │   - engine.py         │                       │
│                    └───────────┬───────────┘                       │
│                                │                                   │
├────────────────────────────────┼───────────────────────────────────┤
│                                │                                   │
│  ┌─────────────────┐  ┌───────▼───────┐  ┌─────────────────┐      │
│  │    Commands     │  │     Hooks     │  │     Skills      │      │
│  │                 │  │               │  │                 │      │
│  │ - /alfred  │  │ - post_lsp    │  │ - moai-ralph    │      │
│  │ - /moai-loop    │  │ - stop_loop   │  │                 │      │
│  │ - /moai-fix     │  │               │  │                 │      │
│  │ - /cancel-loop  │  │               │  │                 │      │
│  └─────────────────┘  └───────────────┘  └─────────────────┘      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.2 모듈 명세

#### 4.2.1 LSP 통합 레이어 (`src/moai_adk/lsp/`)

**client.py - MoAILSPClient**

```python
class MoAILSPClient:
    """LSP 클라이언트 인터페이스"""

    async def get_diagnostics(self, file_path: str) -> list[Diagnostic]:
        """파일의 진단 정보(오류/경고) 조회"""

    async def find_references(self, file_path: str, position: Position) -> list[Location]:
        """심볼의 모든 참조 위치 검색"""

    async def rename_symbol(self, file_path: str, position: Position, new_name: str) -> WorkspaceEdit:
        """심볼 이름을 안전하게 변경"""

    async def get_hover_info(self, file_path: str, position: Position) -> HoverInfo:
        """심볼의 타입 및 문서 정보 조회"""
```

**server_manager.py - LSPServerManager**

```python
class LSPServerManager:
    """LSP 서버 생명주기 관리"""

    async def start_server(self, language: str) -> LSPServer:
        """특정 언어의 LSP 서버 시작"""

    async def stop_server(self, language: str) -> None:
        """LSP 서버 종료"""

    async def get_server(self, language: str) -> LSPServer | None:
        """활성 LSP 서버 조회"""

    def get_language_for_file(self, file_path: str) -> str | None:
        """파일 확장자로 언어 결정"""
```

**protocol.py**

```python
class LSPProtocol:
    """LSP JSON-RPC 2.0 프로토콜 구현"""

    async def send_request(self, method: str, params: dict) -> Any:
        """요청 전송 및 응답 대기"""

    async def send_notification(self, method: str, params: dict) -> None:
        """알림 전송 (응답 없음)"""

    def handle_response(self, response: dict) -> None:
        """응답 처리"""
```

#### 4.2.2 AST-grep 강화 레이어 (`src/moai_adk/astgrep/`)

**analyzer.py - MoAIASTGrepAnalyzer**

```python
class MoAIASTGrepAnalyzer:
    """AST-grep 기반 코드 분석기"""

    def scan_file(self, file_path: str, config: ScanConfig | None = None) -> ScanResult:
        """단일 파일 스캔"""

    def scan_project(self, project_path: str, config: ScanConfig | None = None) -> ProjectScanResult:
        """프로젝트 전체 스캔"""

    def pattern_search(self, pattern: str, language: str, path: str) -> list[Match]:
        """커스텀 패턴 검색"""

    def pattern_replace(self, pattern: str, replacement: str, language: str, path: str, dry_run: bool = True) -> ReplaceResult:
        """패턴 기반 코드 변환"""
```

#### 4.2.3 루프 컨트롤러 (`src/moai_adk/loop/`)

**controller.py - MoAILoopController**

```python
class MoAILoopController:
    """Ralph-style 피드백 루프 컨트롤러"""

    def start_loop(self, promise: str, max_iterations: int = 10) -> LoopState:
        """새 루프 시작"""

    def check_completion(self, state: LoopState) -> CompletionResult:
        """완료 조건 검사"""

    async def run_feedback_loop(self, state: LoopState) -> FeedbackResult:
        """LSP+AST 피드백 실행"""

    def cancel_loop(self, loop_id: str) -> bool:
        """활성 루프 취소"""
```

**state.py - LoopState**

```python
@dataclass
class LoopState:
    """루프 상태 추적"""
    loop_id: str
    promise: str                    # 완료 조건 (예: "모든 LSP 오류 해결")
    current_iteration: int
    max_iterations: int
    status: LoopStatus              # RUNNING, COMPLETED, CANCELLED, FAILED
    created_at: datetime
    updated_at: datetime
    diagnostics_history: list[DiagnosticSnapshot]
    ast_issues_history: list[ASTIssueSnapshot]
```

**feedback.py**

```python
class FeedbackGenerator:
    """Claude에 제공할 피드백 생성"""

    def generate_feedback(self, lsp_result: DiagnosticResult, ast_result: ScanResult) -> str:
        """LSP 및 AST 결과를 Claude 피드백으로 변환"""

    def format_for_hook(self, feedback: str) -> dict:
        """훅 출력 형식으로 변환"""
```

### 4.3 데이터 모델

```python
# 공통 타입
@dataclass
class Position:
    line: int
    character: int

@dataclass
class Range:
    start: Position
    end: Position

@dataclass
class Location:
    uri: str
    range: Range

# LSP 타입
@dataclass
class Diagnostic:
    range: Range
    severity: DiagnosticSeverity  # ERROR, WARNING, INFO, HINT
    code: str | int | None
    source: str
    message: str

# AST-grep 타입
@dataclass
class ASTMatch:
    rule_id: str
    severity: str
    message: str
    file_path: str
    range: Range
    suggested_fix: str | None

# 루프 타입
class LoopStatus(Enum):
    RUNNING = "running"
    COMPLETED = "completed"
    CANCELLED = "cancelled"
    FAILED = "failed"
    PAUSED = "paused"
```

### 4.4 명령어 명세

#### `/alfred`

```yaml
name: alfred
description: 원클릭 자동화 - SPEC 생성부터 동기화까지
parameters:
  - name: spec_description
    type: string
    required: true
    description: SPEC 설명
  - name: branch
    type: boolean
    default: false
    description: 새 브랜치 생성 여부
  - name: pr
    type: boolean
    default: false
    description: PR 생성 여부
workflow:
  1. /moai:1-plan "$spec_description"
  2. /moai:2-run SPEC-XXX
  3. /moai:3-sync SPEC-XXX
  4. (optional) 브랜치/PR 생성
```

#### `/moai-loop`

```yaml
name: moai-loop
description: Ralph-style 자율 피드백 루프 시작
parameters:
  - name: promise
    type: string
    required: true
    description: 완료 조건 (예: "모든 타입 오류 해결")
  - name: max_iterations
    type: integer
    default: 10
    description: 최대 반복 횟수
workflow:
  1. 루프 상태 초기화
  2. LSP 진단 + AST 스캔 실행
  3. 결과를 Claude에 피드백
  4. Claude가 수정 수행
  5. 완료 조건 검사
  6. 미완료 시 2단계로 복귀
```

### 4.5 훅 명세

#### `post_tool__lsp_diagnostic.py`

```yaml
event: PostToolUse
tools: [Write, Edit]
behavior:
  1. 수정된 파일의 언어 감지
  2. 해당 언어 LSP 서버에서 진단 조회
  3. 진단 결과를 additionalContext로 반환
output:
  hookSpecificOutput:
    hookEventName: PostToolUse
    additionalContext: "LSP: 2 errors, 1 warning in file.py\n  - Line 10: Type error..."
exit_codes:
  0: 성공 (진단 없음 또는 정보만)
  2: 주의 필요 (오류 발견)
```

#### `stop__loop_controller.py`

```yaml
event: Stop
behavior:
  1. 활성 루프 상태 확인
  2. LSP + AST 진단 실행
  3. 완료 조건 검사
  4. 미완료 시 추가 작업 요청
output:
  decision: BLOCK | ALLOW
  reason: "루프 미완료: 3개 오류 남음"
  follow_up_prompt: "다음 오류를 수정하세요: ..."
```

### 4.6 통합 포인트

#### Claude Code Hooks

| 훅 | 이벤트 | 목적 |
|----|--------|------|
| `post_tool__lsp_diagnostic.py` | PostToolUse | 편집 후 LSP 진단 |
| `post_tool__ast_grep_scan.py` | PostToolUse | 편집 후 AST 스캔 (기존) |
| `stop__loop_controller.py` | Stop | Ralph 루프 제어 |

#### 설정 통합

```yaml
# .moai/config/sections/ralph.yaml
ralph:
  enabled: true
  lsp:
    auto_start: true
    timeout_seconds: 30
  ast_grep:
    config_path: ".claude/skills/moai-tool-ast-grep/rules/sgconfig.yml"
    security_scan: true
  loop:
    max_iterations: 10
    auto_fix: false
    require_confirmation: true
  git:
    auto_branch: false      # 사용자 요구: 브랜치 생성 선택적
    auto_pr: false          # 사용자 요구: PR 생성 선택적
```

---

## 5. Constraints (제약 조건)

### 5.1 기술적 제약

| ID | 제약 | 영향 | 대응 방안 |
|----|------|------|----------|
| C-01 | LSP 서버 응답 시간 ≤ 5초 | 사용자 경험 | 타임아웃 및 캐싱 |
| C-02 | AST-grep 스캔 시간 ≤ 30초 | 워크플로우 지연 | 증분 스캔 |
| C-03 | 메모리 사용량 ≤ 500MB | 시스템 리소스 | 서버 lazy loading |
| C-04 | Python 3.13+ 필수 | 호환성 | 버전 체크 추가 |

### 5.2 보안 제약

| ID | 제약 | 근거 |
|----|------|------|
| S-01 | LSP 서버는 로컬에서만 실행 | 코드 유출 방지 |
| S-02 | 자동 수정 전 확인 필요 (기본값) | 의도하지 않은 변경 방지 |
| S-03 | 민감 파일 스캔 제외 | 보안 정보 보호 |

### 5.3 운영 제약

| ID | 제약 | 영향 |
|----|------|------|
| O-01 | 최대 동시 LSP 서버: 5개 | 리소스 제한 |
| O-02 | 루프 최대 반복: 10회 (기본) | 무한 루프 방지 |
| O-03 | Git 작업은 선택적 | 사용자 워크플로우 존중 |

---

## 6. Traceability (추적성)

### 6.1 요구사항-구현 매핑

| 요구사항 | 구현 파일 | 테스트 파일 |
|---------|----------|------------|
| REQ-LSP-001 | `src/moai_adk/lsp/client.py` | `tests/lsp/test_client.py` |
| REQ-LSP-002 | `src/moai_adk/lsp/server_manager.py` | `tests/lsp/test_server_manager.py` |
| REQ-AST-001 | `src/moai_adk/astgrep/analyzer.py` | `tests/astgrep/test_analyzer.py` |
| REQ-LOOP-001 | `src/moai_adk/loop/controller.py` | `tests/loop/test_controller.py` |
| REQ-CMD-001 | `.claude/commands/moai/alfred.md` | `tests/commands/test_alfred.py` |
| REQ-HOOK-001 | `.claude/hooks/moai/post_tool__lsp_diagnostic.py` | `tests/hooks/test_lsp_hook.py` |

### 6.2 관련 문서

| 문서 | 위치 | 설명 |
|------|------|------|
| 구현 계획 | `.moai/specs/SPEC-RALPH-001/plan.md` | 단계별 구현 계획 |
| 인수 조건 | `.moai/specs/SPEC-RALPH-001/acceptance.md` | Given-When-Then 시나리오 |
| LSP 설정 | `.lsp.json` | 언어 서버 설정 |
| AST-grep 스킬 | `.claude/skills/moai-tool-ast-grep/` | AST-grep 통합 |

---

## Appendix A: 기술 참조

### A.1 LSP 공식 스펙

- [Language Server Protocol Specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/)
- JSON-RPC 2.0 기반 통신
- 주요 메서드: `textDocument/diagnostic`, `textDocument/references`, `textDocument/rename`

### A.2 AST-grep 문서

- [AST-grep Official Documentation](https://ast-grep.github.io/)
- 패턴 문법: `$VAR` (단일 노드), `$$$ARGS` (가변 노드)
- 규칙 파일: YAML 형식

### A.3 Ralph Wiggum 플러그인 참조

- Stop 훅 기반 자율 루프
- Promise-based 완료 조건
- 사용자 취소 지원

---

**문서 버전**: 1.0.0
**최종 수정**: 2026-01-09
**작성자**: workflow-spec agent
