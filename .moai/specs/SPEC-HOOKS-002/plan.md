# SPEC-HOOKS-002 Implementation Plan

> **moai_hooks.py Self-contained Hook Script**
>
> 600줄 단일 파일, Zero Dependencies, PEP 723 준수

---

## Implementation Overview

### Goals
- 600 LOC 이내 단일 Python 스크립트 작성
- Zero external dependencies (표준 라이브러리만 사용)
- PEP 723 inline metadata 준수
- 9개 Claude Code hook events 지원
- 20개 프로그래밍 언어 자동 감지
- JIT Context Retrieval 구현
- stdin/stdout JSON 프로토콜

### Success Criteria
- [ ] 파일 크기 ≤ 600 LOC
- [ ] 외부 pip 패키지 zero
- [ ] 모든 9개 hook 동작
- [ ] 테스트 커버리지 ≥ 85%
- [ ] SessionStart < 500ms, 기타 < 100ms
- [ ] PEP 723 검증 통과

---

## Phase 1: PEP 723 Setup

### Task 1.1: Create moai_hooks.py skeleton
**Priority**: Critical

**작업 내용**:
1. `templates/.claude/hooks/moai_hooks.py` 파일 생성
2. Shebang 추가
3. PEP 723 header 작성
4. Docstring 작성
5. 실행 권한 부여

**검증 기준**:
- [ ] PEP 723 header 존재
- [ ] dependencies = []
- [ ] shebang 존재

**예상 LOC**: 20

---

### Task 1.2: Define core data structures
**Priority**: Critical

**작업 내용**:
1. Import statements
2. HookEvent Enum 정의 (9개)
3. TypedDict 정의 (HookPayload, HookResult)

**검증 기준**:
- [ ] 표준 라이브러리만 사용
- [ ] 9개 enum 값 존재

**예상 LOC**: 50

---

## Phase 2: 9 Hook Handlers Implementation

### Task 2.1: SessionStart handler
- Git 정보 수집
- SPEC 진행률 계산
- 언어 감지
- 배너 생성

**예상 LOC**: 30

### Task 2.2: UserPromptSubmit handler
- JIT Context Retrieval
- 명령어 패턴 매칭

**예상 LOC**: 20

### Task 2.3: PreCompact handler
- 세션 요약 생성
- 권장사항 제공

**예상 LOC**: 15

### Task 2.4: Other 6 handlers
- SessionEnd, PreToolUse, PostToolUse, Notification, Stop, SubagentStop

**예상 LOC**: 60

---

## Phase 3: Language Detection

### Task 3.1: Implement LanguageDetector class
- 20개 언어 시그니처 정의
- detect() 메서드
- _match_signature() 메서드 (glob 지원)

**검증 기준**:
- [ ] 20개 언어 모두 정의
- [ ] Multi-language 프로젝트 지원

**예상 LOC**: 80

---

## Phase 4: SPEC Progress Tracker

### Task 4.1: Implement SpecProgressTracker class
- calculate() 메서드
- format() 메서드
- .moai/specs/ 디렉토리 스캔

**검증 기준**:
- [ ] 진행률 계산 정확
- [ ] 형식: "X/Y (Z%)"

**예상 LOC**: 60

---

## Phase 5: Git Integration

### Task 5.1: Implement GitInfoProvider class
- get_info() 메서드
- _run_git_command() 메서드 (타임아웃)
- _count_changes() 메서드

**검증 기준**:
- [ ] 2초 타임아웃 적용
- [ ] Git 저장소 아닐 때 None 반환

**예상 LOC**: 80

---

## Phase 6: JIT Context Retrieval

### Task 6.1: Implement JITContextRetriever class
- COMMAND_CONTEXT_MAP 정의
- retrieve() 메서드
- read_content() 메서드 (10KB 제한)

**검증 기준**:
- [ ] /alfred:1-spec → spec-metadata.md 매칭
- [ ] 파일 존재하지 않으면 skip

**예상 LOC**: 70

---

## Phase 7: HookRuntime Main Class

### Task 7.1: Implement HookRuntime class
- __init__() 메서드
- run() 메서드 (동적 디스패치)
- Helper methods

**검증 기준**:
- [ ] 모든 컴포넌트 초기화
- [ ] 예외 처리 적용

**예상 LOC**: 180

---

## Phase 8: Main Entry Point

### Task 8.1: Implement main() function
- stdin JSON 읽기
- Hook 실행
- stdout JSON 출력
- Error handling

**검증 기준**:
- [ ] JSON 파싱 오류 처리
- [ ] exit code 설정

**예상 LOC**: 50

---

## Phase 9: Error Handling & Edge Cases

### Task 9.1: JSON parsing errors
- stderr 출력
- exit code 1

### Task 9.2: File not found errors
- Gracefully skip
- 기본값 반환

### Task 9.3: Timeout enforcement
- 2초 타임아웃

---

## Phase 10: Testing

### Task 10.1: Unit tests (pytest)
- PEP 723 header validation
- LanguageDetector (20개 언어)
- SpecProgressTracker
- GitInfoProvider
- JITContextRetriever
- 9개 hook handlers

**검증 기준**:
- [ ] 커버리지 ≥ 85%
- [ ] 모든 테스트 통과

### Task 10.2: Integration tests
- E2E stdin/stdout flow
- uv run execution
- Timeout enforcement
- Error scenarios

### Task 10.3: Performance tests
- SessionStart < 500ms
- 기타 핸들러 < 100ms
- 메모리 < 50MB

---

## Phase 11: Documentation

### Task 11.1: Inline comments
- 모든 클래스/메서드 docstring
- Type hints 명확히

### Task 11.2: Usage examples
- README.md 업데이트
- plan.md 상세 실행 방법

---

## Phase 12: Migration & Deployment

### Task 12.1: Remove TypeScript hooks
```bash
rm -rf templates/.claude/hooks/index.ts
```

### Task 12.2: Deploy Python hooks
```bash
cp moai-adk-py/templates/.claude/hooks/moai_hooks.py \
   templates/.claude/hooks/moai_hooks.py

chmod +x templates/.claude/hooks/moai_hooks.py
```

### Task 12.3: Update settings.json
- 9개 hook 명령어 업데이트

### Task 12.4: Test deployment
- 모든 hook 정상 동작 확인

---

## Dependencies

### Internal Dependencies
- **SPEC-PY314-001**: Python 3.14 기반 프로젝트 구조
- **SPEC-HOOKS-001**: Hooks 런타임 시스템

### External Dependencies
**런타임**: 없음 (zero dependencies)

---

## Timeline Summary

| Phase | 작업 내용 | 예상 소요 시간 |
|-------|----------|-------------|
| Phase 1-2 | Setup & Handlers | 4일 |
| Phase 3-6 | Core Components | 4일 |
| Phase 7-9 | Runtime & Error Handling | 4일 |
| Phase 10-12 | Testing & Deployment | 4일 |
| **Total** | | **16일** |

---

## Validation Checklist

### Code Quality
- [ ] 파일 크기 ≤ 600 LOC
- [ ] 외부 pip 패키지 zero
- [ ] PEP 723 준수
- [ ] Type hints 정확
- [ ] Docstring 완비

### Functionality
- [ ] 9개 hook 모두 동작
- [ ] 20개 언어 감지
- [ ] SPEC 진행률 정확
- [ ] Git 정보 수집
- [ ] JIT Context Retrieval

### Performance
- [ ] SessionStart < 500ms
- [ ] 기타 핸들러 < 100ms
- [ ] 메모리 < 50MB

### Testing
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 모든 unit tests 통과
- [ ] 모든 integration tests 통과

### Security
- [ ] Zero dependencies
- [ ] Read-only operations
- [ ] No network calls
- [ ] Timeout protection

---

**작성자**: @Goos
**작성일**: 2025-10-13
