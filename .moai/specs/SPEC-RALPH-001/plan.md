# SPEC-RALPH-001: 구현 계획

## TAG BLOCK

```yaml
SPEC-ID: SPEC-RALPH-001
Title: MoAI Ralph Engine - 구현 계획
Created: 2026-01-09
Status: Planned
Related-Files:
  - spec.md
  - acceptance.md
```

---

## 1. 구현 개요

### 1.1 목표

MoAI Ralph Engine의 단계적 구현을 통해:
1. LSP 통합으로 실시간 코드 진단 제공
2. AST-grep 강화로 구조적 코드 분석 향상
3. Ralph Loop로 자율 피드백 메커니즘 구현
4. 명령어 및 훅으로 사용자 인터페이스 제공

### 1.2 구현 원칙

- **TDD (Test-Driven Development)**: 모든 기능은 테스트 우선 개발
- **점진적 통합**: 각 레이어를 독립적으로 개발 후 통합
- **기존 자산 활용**: `.lsp.json`, `post_tool__ast_grep_scan.py` 재사용
- **하위 호환성**: 기존 MoAI-ADK 워크플로우 유지

---

## 2. 마일스톤

### Phase 1: LSP 클라이언트 기본 구현

**목표**: LSP 서버와의 기본 통신 및 진단 기능 구현

**산출물**:
- `src/moai_adk/lsp/__init__.py`
- `src/moai_adk/lsp/protocol.py`
- `src/moai_adk/lsp/client.py`
- `src/moai_adk/lsp/server_manager.py`
- `src/moai_adk/lsp/models.py`
- `tests/lsp/test_protocol.py`
- `tests/lsp/test_client.py`
- `tests/lsp/test_server_manager.py`

**작업 항목**:

1. **LSP 프로토콜 구현** (Priority: HIGH)
   - JSON-RPC 2.0 메시지 직렬화/역직렬화
   - 요청/응답/알림 처리
   - 에러 핸들링

2. **LSP 서버 매니저 구현** (Priority: HIGH)
   - `.lsp.json` 설정 파싱
   - 언어별 서버 프로세스 관리
   - 서버 시작/종료 생명주기

3. **LSP 클라이언트 구현** (Priority: HIGH)
   - `get_diagnostics()` - 진단 정보 조회
   - `initialize()` / `shutdown()` 핸드셰이크
   - 파일 동기화 (didOpen, didChange, didClose)

4. **테스트 작성** (Priority: HIGH)
   - 유닛 테스트: 프로토콜 파싱
   - 통합 테스트: 실제 LSP 서버 (pyright)
   - Mock 테스트: 서버 응답 시뮬레이션

**의존성**:
- Python 3.13+
- asyncio
- `.lsp.json` 설정 파일

**검증 기준**:
- pyright-langserver와 성공적 핸드셰이크
- Python 파일에서 진단 정보 조회 성공
- 테스트 커버리지 ≥ 85%

---

### Phase 2: AST-grep 분석기 강화

**목표**: 기존 AST-grep 훅을 확장하여 분석기 클래스 구현

**산출물**:
- `src/moai_adk/astgrep/__init__.py`
- `src/moai_adk/astgrep/analyzer.py`
- `src/moai_adk/astgrep/models.py`
- `src/moai_adk/astgrep/rules.py`
- `tests/astgrep/test_analyzer.py`

**작업 항목**:

1. **분석기 클래스 설계** (Priority: HIGH)
   - 기존 `post_tool__ast_grep_scan.py` 로직 리팩토링
   - 클래스 기반 재사용 가능한 API
   - 설정 기반 규칙 로딩

2. **스캔 기능 구현** (Priority: HIGH)
   - `scan_file()` - 단일 파일 스캔
   - `scan_project()` - 프로젝트 전체 스캔
   - 결과 집계 및 리포팅

3. **패턴 검색/변환 구현** (Priority: MEDIUM)
   - `pattern_search()` - 커스텀 패턴 검색
   - `pattern_replace()` - 패턴 기반 변환
   - dry-run 모드 지원

4. **규칙 관리** (Priority: MEDIUM)
   - 보안 규칙 로딩
   - 품질 규칙 로딩
   - 커스텀 규칙 지원

**의존성**:
- ast-grep CLI (`sg`)
- 기존 `sgconfig.yml`

**검증 기준**:
- Python/TypeScript 파일 스캔 성공
- 보안 취약점 탐지 (SQL injection, XSS 등)
- 패턴 변환 dry-run 정확성

---

### Phase 3: 루프 컨트롤러 구현

**목표**: Ralph-style 자율 피드백 루프 메커니즘 구현

**산출물**:
- `src/moai_adk/loop/__init__.py`
- `src/moai_adk/loop/controller.py`
- `src/moai_adk/loop/state.py`
- `src/moai_adk/loop/feedback.py`
- `src/moai_adk/loop/storage.py`
- `tests/loop/test_controller.py`
- `tests/loop/test_state.py`

**작업 항목**:

1. **루프 상태 관리** (Priority: HIGH)
   - `LoopState` 데이터클래스
   - 상태 직렬화/역직렬화
   - 히스토리 추적

2. **루프 컨트롤러 구현** (Priority: HIGH)
   - `start_loop()` - 루프 초기화
   - `check_completion()` - 완료 조건 검사
   - `cancel_loop()` - 루프 취소

3. **피드백 루프 실행** (Priority: HIGH)
   - LSP 진단 수집
   - AST-grep 스캔 실행
   - 결과 통합 및 피드백 생성

4. **피드백 생성기** (Priority: MEDIUM)
   - Claude 친화적 형식 변환
   - 우선순위 기반 이슈 정렬
   - 수정 제안 포함

5. **상태 저장소** (Priority: MEDIUM)
   - 파일 기반 상태 저장
   - 세션 간 상태 복원
   - 히스토리 조회

**의존성**:
- Phase 1 (LSP 클라이언트)
- Phase 2 (AST-grep 분석기)

**검증 기준**:
- 루프 시작/중지/취소 정상 동작
- 완료 조건 정확한 평가
- 최대 반복 제한 준수

---

### Phase 4: 명령어 및 훅 통합

**목표**: 사용자 인터페이스 및 Claude Code 통합

**산출물**:
- `.claude/commands/moai/alfred.md`
- `.claude/commands/moai/moai-loop.md`
- `.claude/commands/moai/moai-fix.md`
- `.claude/commands/moai/cancel-loop.md`
- `.claude/hooks/moai/post_tool__lsp_diagnostic.py`
- `.claude/hooks/moai/stop__loop_controller.py`
- `.claude/skills/moai-ralph/SKILL.md`

**작업 항목**:

1. **명령어 구현** (Priority: HIGH)
   - `/alfred`: 원클릭 자동화 플로우
   - `/moai-loop`: Ralph 루프 시작
   - `/moai-fix`: 자동 수정 트리거
   - `/cancel-loop`: 루프 취소

2. **PostToolUse 훅 구현** (Priority: HIGH)
   - `post_tool__lsp_diagnostic.py`: Write/Edit 후 LSP 진단
   - 기존 `post_tool__ast_grep_scan.py`와 조화

3. **Stop 훅 구현** (Priority: HIGH)
   - `stop__loop_controller.py`: Ralph 루프 제어
   - 완료 조건 검사
   - 추가 작업 요청

4. **스킬 정의** (Priority: MEDIUM)
   - `moai-ralph/SKILL.md`: Ralph Engine 스킬
   - 컨텍스트 및 트리거 정의
   - 관련 에이전트 연결

5. **설정 통합** (Priority: MEDIUM)
   - `.moai/config/sections/ralph.yaml`
   - 브랜치/PR 선택적 설정
   - 루프 설정 커스터마이징

**의존성**:
- Phase 1, 2, 3 완료
- Claude Code hooks 시스템

**검증 기준**:
- 모든 명령어 정상 실행
- 훅이 적절한 시점에 트리거
- 설정 변경이 즉시 반영

---

### Phase 5: 테스팅 및 문서화

**목표**: 종합 테스트 및 문서 완성

**산출물**:
- `tests/integration/test_ralph_engine.py`
- `tests/e2e/test_full_workflow.py`
- `docs/ralph-engine.md`
- `CHANGELOG.md` 업데이트

**작업 항목**:

1. **통합 테스트** (Priority: HIGH)
   - LSP + AST-grep 통합 테스트
   - 루프 컨트롤러 통합 테스트
   - 명령어 통합 테스트

2. **E2E 테스트** (Priority: HIGH)
   - `/alfred` 전체 플로우
   - `/moai-loop` 완료까지 실행
   - 오류 복구 시나리오

3. **성능 테스트** (Priority: MEDIUM)
   - LSP 응답 시간 측정
   - AST-grep 스캔 시간 측정
   - 메모리 사용량 프로파일링

4. **문서화** (Priority: MEDIUM)
   - API 문서
   - 사용자 가이드
   - 트러블슈팅 가이드

5. **CHANGELOG 업데이트** (Priority: LOW)
   - 새 기능 설명
   - Breaking changes 명시
   - 마이그레이션 가이드

**의존성**:
- Phase 1-4 완료

**검증 기준**:
- 전체 테스트 커버리지 ≥ 85%
- E2E 테스트 100% 통과
- 문서 완성도 검토

---

## 3. 기술적 접근 방식

### 3.1 LSP 통합 전략

```
┌─────────────────────────────────────────────────────────────┐
│                    LSP Integration Flow                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Configuration Loading                                   │
│     .lsp.json → LSPConfig → ServerManager                   │
│                                                             │
│  2. Server Lifecycle                                        │
│     File Open → Detect Language → Start Server (if needed)  │
│     File Close → Check References → Stop Server (if idle)   │
│                                                             │
│  3. Diagnostic Flow                                         │
│     Edit Event → didChange → Server Diagnostics → Client    │
│                                                             │
│  4. Request/Response                                        │
│     Client Request → JSON-RPC → Server → JSON-RPC → Client  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**핵심 설계 결정**:

1. **비동기 우선**: 모든 LSP 통신은 `asyncio` 기반
2. **Lazy Loading**: 필요 시에만 LSP 서버 시작
3. **Connection Pooling**: 언어별 단일 서버 인스턴스 유지
4. **Timeout 처리**: 모든 요청에 5초 타임아웃

### 3.2 AST-grep 통합 전략

```
┌─────────────────────────────────────────────────────────────┐
│                  AST-grep Integration Flow                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Rule Configuration                                      │
│     sgconfig.yml → RuleLoader → RuleSet                     │
│                                                             │
│  2. Scan Execution                                          │
│     File/Project → sg CLI → JSON Output → ScanResult        │
│                                                             │
│  3. Result Processing                                       │
│     ScanResult → Filter → Prioritize → Format               │
│                                                             │
│  4. Pattern Operations                                      │
│     Pattern → Validate → Search/Replace → Preview/Apply     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**핵심 설계 결정**:

1. **CLI 래퍼**: `sg` CLI를 subprocess로 호출
2. **JSON 출력**: `--json` 플래그로 구조화된 결과
3. **증분 스캔**: 변경된 파일만 스캔 (캐싱)
4. **규칙 모듈화**: 보안/품질/커스텀 규칙 분리

### 3.3 루프 컨트롤러 전략

```
┌─────────────────────────────────────────────────────────────┐
│                   Ralph Loop Architecture                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐                                               │
│  │  Start   │ ─────────────────────────────────────────┐    │
│  └────┬─────┘                                          │    │
│       │                                                │    │
│       ▼                                                │    │
│  ┌──────────────┐                                      │    │
│  │ Run Feedback │ ◄───────────────────────────────┐    │    │
│  │ (LSP + AST)  │                                 │    │    │
│  └──────┬───────┘                                 │    │    │
│         │                                         │    │    │
│         ▼                                         │    │    │
│  ┌──────────────┐     ┌─────────┐                 │    │    │
│  │ Check        │ ──► │Complete │ ───► End        │    │    │
│  │ Completion   │     └─────────┘                 │    │    │
│  └──────┬───────┘                                 │    │    │
│         │ No                                      │    │    │
│         ▼                                         │    │    │
│  ┌──────────────┐                                 │    │    │
│  │ Check Max    │ ──► Exceeded ───► End           │    │    │
│  │ Iterations   │                                 │    │    │
│  └──────┬───────┘                                 │    │    │
│         │ OK                                      │    │    │
│         ▼                                         │    │    │
│  ┌──────────────┐                                 │    │    │
│  │ Generate     │                                 │    │    │
│  │ Feedback     │ ────────────────────────────────┘    │    │
│  └──────────────┘                                      │    │
│                                                        │    │
│  ┌──────────┐                                          │    │
│  │ Cancel   │ ─────────────────────────────────────────┘    │
│  └──────────┘                                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**핵심 설계 결정**:

1. **Promise 기반**: 명시적 완료 조건 정의
2. **히스토리 추적**: 각 반복의 진단 결과 저장
3. **안전 장치**: 최대 반복 제한, 사용자 취소
4. **Stop 훅 통합**: Claude 응답 완료 시 루프 체크

### 3.4 훅 통합 전략

```
┌─────────────────────────────────────────────────────────────┐
│                    Hook Integration Points                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  PostToolUse (Write/Edit)                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ 1. post_tool__ast_grep_scan.py (기존)               │    │
│  │    - AST-grep 보안 스캔                             │    │
│  │    - 보안 이슈 리포팅                               │    │
│  │                                                     │    │
│  │ 2. post_tool__lsp_diagnostic.py (신규)              │    │
│  │    - LSP 진단 조회                                  │    │
│  │    - 타입 오류/경고 리포팅                          │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  Stop                                                       │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ stop__loop_controller.py (신규)                     │    │
│  │    - 활성 루프 검사                                 │    │
│  │    - 완료 조건 평가                                 │    │
│  │    - 추가 작업 요청 (BLOCK/ALLOW)                   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**핵심 설계 결정**:

1. **훅 순서**: AST-grep → LSP (보안 우선)
2. **비차단 기본**: 정보는 제공하되 워크플로우 차단 최소화
3. **루프 훅**: Stop 이벤트에서만 루프 제어
4. **성능 최적화**: 캐싱, 타임아웃, 증분 처리

---

## 4. 아키텍처 설계

### 4.1 디렉토리 구조

```
src/moai_adk/
├── lsp/
│   ├── __init__.py
│   ├── client.py           # MoAILSPClient
│   ├── server_manager.py   # LSPServerManager
│   ├── protocol.py         # JSON-RPC 프로토콜
│   └── models.py           # LSP 데이터 모델
│
├── astgrep/
│   ├── __init__.py
│   ├── analyzer.py         # MoAIASTGrepAnalyzer
│   ├── models.py           # AST-grep 데이터 모델
│   └── rules.py            # 규칙 로딩/관리
│
├── loop/
│   ├── __init__.py
│   ├── controller.py       # MoAILoopController
│   ├── state.py            # LoopState 데이터클래스
│   ├── feedback.py         # FeedbackGenerator
│   └── storage.py          # 상태 저장소
│
└── ralph/
    ├── __init__.py
    └── engine.py           # RalphEngine (통합 레이어)

.claude/
├── commands/moai/
│   ├── alfred.md
│   ├── moai-loop.md
│   ├── moai-fix.md
│   └── cancel-loop.md
│
├── hooks/moai/
│   ├── post_tool__ast_grep_scan.py  (기존)
│   ├── post_tool__lsp_diagnostic.py (신규)
│   └── stop__loop_controller.py     (신규)
│
└── skills/
    └── moai-ralph/
        ├── SKILL.md
        └── modules/
            └── troubleshooting.md

.moai/config/sections/
└── ralph.yaml              # Ralph Engine 설정

tests/
├── lsp/
│   ├── test_client.py
│   ├── test_server_manager.py
│   └── test_protocol.py
├── astgrep/
│   └── test_analyzer.py
├── loop/
│   ├── test_controller.py
│   └── test_state.py
└── integration/
    └── test_ralph_engine.py
```

### 4.2 의존성 그래프

```
┌─────────────────────────────────────────────────────────────┐
│                    Dependency Graph                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                    ┌───────────────┐                        │
│                    │ RalphEngine   │                        │
│                    └───────┬───────┘                        │
│                            │                                │
│              ┌─────────────┼─────────────┐                  │
│              │             │             │                  │
│              ▼             ▼             ▼                  │
│     ┌────────────┐ ┌────────────┐ ┌────────────┐           │
│     │ LSPClient  │ │ASTAnalyzer │ │LoopControl │           │
│     └─────┬──────┘ └─────┬──────┘ └─────┬──────┘           │
│           │              │              │                   │
│           ▼              │              │                   │
│     ┌────────────┐       │              │                   │
│     │ServerMgr   │       │              │                   │
│     └─────┬──────┘       │              │                   │
│           │              │              │                   │
│           ▼              ▼              ▼                   │
│     ┌────────────┐ ┌────────────┐ ┌────────────┐           │
│     │ Protocol   │ │  sg CLI    │ │  Storage   │           │
│     └────────────┘ └────────────┘ └────────────┘           │
│                                                             │
│     External:                                               │
│     ┌────────────┐ ┌────────────┐ ┌────────────┐           │
│     │ .lsp.json  │ │sgconfig.yml│ │ .moai/     │           │
│     └────────────┘ └────────────┘ └────────────┘           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 5. 리스크 및 대응 방안

### 5.1 기술적 리스크

| ID | 리스크 | 확률 | 영향 | 대응 방안 |
|----|--------|------|------|----------|
| R-01 | LSP 서버 설치 안됨 | MEDIUM | HIGH | Graceful degradation, 설치 가이드 제공 |
| R-02 | AST-grep CLI 없음 | MEDIUM | MEDIUM | 기존 훅과 동일하게 스킵 처리 |
| R-03 | 비동기 처리 복잡성 | LOW | MEDIUM | asyncio 패턴 표준화, 테스트 강화 |
| R-04 | 훅 충돌 | LOW | HIGH | 훅 우선순위 명시, 테스트 |
| R-05 | 메모리 누수 (LSP 서버) | LOW | HIGH | 서버 풀 관리, 주기적 정리 |

### 5.2 일정 리스크

| ID | 리스크 | 확률 | 대응 방안 |
|----|--------|------|----------|
| S-01 | Phase 1 지연 | LOW | LSP 프로토콜은 표준, 참조 구현 활용 |
| S-02 | 통합 테스트 복잡성 | MEDIUM | 단계별 통합, Mock 활용 |
| S-03 | 문서화 지연 | MEDIUM | 코드 작성 시 병행 문서화 |

### 5.3 대응 전략

1. **Graceful Degradation**: LSP/AST-grep 없어도 기본 기능 동작
2. **점진적 롤아웃**: Feature flag로 단계적 활성화
3. **롤백 계획**: 각 Phase 별 롤백 지점 정의
4. **모니터링**: 성능 메트릭 수집 및 알림

---

## 6. 품질 게이트

### 6.1 Phase 완료 조건

| Phase | 테스트 커버리지 | 통합 테스트 | 문서화 | 코드 리뷰 |
|-------|----------------|------------|--------|----------|
| Phase 1 | ≥ 85% | LSP 핸드셰이크 | API 문서 | 필수 |
| Phase 2 | ≥ 85% | 스캔 통합 | API 문서 | 필수 |
| Phase 3 | ≥ 85% | 루프 통합 | API 문서 | 필수 |
| Phase 4 | ≥ 80% | E2E 플로우 | 사용자 가이드 | 필수 |
| Phase 5 | ≥ 85% (전체) | 전체 E2E | 완성 | 필수 |

### 6.2 TRUST 5 체크리스트

- [ ] **Test-first**: 모든 기능 테스트 우선 작성
- [ ] **Readable**: 명확한 변수명, 함수명, 주석
- [ ] **Unified**: 일관된 코드 스타일 (ruff, black)
- [ ] **Secured**: 보안 취약점 스캔 통과
- [ ] **Trackable**: 명확한 커밋 메시지, 변경 추적

---

## 7. 다음 단계

1. Phase 1 시작 전 기술 스파이크:
   - pyright-langserver 연동 테스트
   - JSON-RPC 라이브러리 선택 (`python-lsp-jsonrpc` vs 직접 구현)

2. `/moai:2-run SPEC-RALPH-001` 실행 시:
   - Phase 1부터 TDD로 구현 시작
   - RED-GREEN-REFACTOR 사이클 준수

3. 진행 상황 추적:
   - `.moai/specs/SPEC-RALPH-001/` 내 상태 업데이트
   - GitHub Issue 생성 (Team 모드 시)

---

**문서 버전**: 1.0.0
**최종 수정**: 2026-01-09
**작성자**: workflow-spec agent
