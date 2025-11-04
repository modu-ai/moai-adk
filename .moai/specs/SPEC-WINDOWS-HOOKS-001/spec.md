---
id: WINDOWS-HOOKS-001
version: 0.1.0
status: implementation-complete
created: 2025-10-18
updated: 2025-10-18
author: @Goos
priority: high
category: bugfix
labels:
  - windows
  - cross-platform
  - hooks
  - stdin
related_issue: "https://github.com/modu-ai/moai-adk/issues/25"
scope:
  packages:
    - .claude/hooks/alfred
  files:
    - alfred_hooks.py
---

# @SPEC:WINDOWS-HOOKS-001: Windows 환경에서 Claude Code 훅 stdin 처리 개선

## HISTORY

### v0.1.0 (2025-10-18)
- **COMPLETED**: TDD 구현 완료 (RED → GREEN → REFACTOR)
- **TESTED**: 4/4 테스트 통과 (Windows/macOS/Linux 크로스 플랫폼)
- **CODE**: Iterator 패턴으로 stdin 읽기 개선
  - alfred_hooks.py:125 - `sys.stdin.read()` → `for line in sys.stdin`
  - 빈 stdin 처리: `{}` 기본값 반환
- **VERIFIED**: 모든 SPEC 요구사항 충족
  - Windows/macOS/Linux stdin 안정적 읽기 ✓
  - 빈 stdin 처리 ✓
  - JSON 파싱 에러 처리 ✓
  - 크로스 플랫폼 호환성 ✓
- **COMMITS**:
  - 31097b3 - 🔴 RED: Windows stdin 처리 테스트 작성
  - 711bf44 - 🟢 GREEN: Iterator 패턴으로 stdin 읽기 구현
- **FILES**:
  - tests/hooks/test_alfred_hooks_stdin.py (155줄 추가)
  - .claude/hooks/alfred/alfred_hooks.py (17줄 변경)
- **AUTHOR**: @Goos
- **FIXES**: #25, #31 (GitHub Issues)

### v0.0.1 (2025-10-18)
- **INITIAL**: Windows 환경에서 stdin 읽기 개선 명세 작성
- **AUTHOR**: @Goos
- **REASON**: GitHub Issue #25, #31에서 보고된 Windows 환경 스크립트 오류 해결
- **RELATED**: https://github.com/modu-ai/moai-adk/issues/25

---

## 개요

Windows 환경에서 Claude Code Hook 스크립트가 stdin으로부터 JSON 페이로드를 읽을 때 발생하는 EOF 처리 문제를 해결합니다. 현재 `sys.stdin.read()`를 사용하는 방식은 Windows에서 EOF를 올바르게 처리하지 못할 수 있어, 크로스 플랫폼 호환성이 보장되는 Iterator 패턴으로 개선합니다.

## 배경 (Background)

### 현재 문제점

**위치**: `.claude/hooks/alfred/alfred_hooks.py:123`

```python
# 현재 구현
input_data = sys.stdin.read()
data = json.loads(input_data)
```

**문제**:
- Windows 환경에서 stdin EOF 처리가 불확실
- GitHub Issue #25, #31에서 스크립트 오류 보고
- 크로스 플랫폼 동작 보장이 명확하지 않음

### 기대 동작

- Windows/macOS/Linux 모든 환경에서 안정적으로 stdin 읽기
- JSON 페이로드를 정확하게 파싱
- 빈 stdin 처리 시 에러 없이 기본값 반환
- 모든 에러는 stderr로 출력, stdout은 JSON만 출력

---

## EARS 요구사항 (Easy Approach to Requirements Syntax)

### Ubiquitous Requirements (기본 요구사항)

- 시스템은 Windows/macOS/Linux 모든 환경에서 stdin을 안정적으로 읽어야 한다
- 시스템은 JSON 형식의 페이로드를 stdin으로부터 파싱해야 한다
- 시스템은 stdin 읽기 로직이 플랫폼에 무관하게 동작해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN Windows 환경에서 stdin이 제공되면, 시스템은 EOF를 올바르게 처리해야 한다
- WHEN stdin 읽기가 실패하면, 시스템은 명확한 에러 메시지를 stderr에 출력해야 한다
- WHEN JSON 파싱이 실패하면, 시스템은 JSONDecodeError를 발생시키고 exit code 1을 반환해야 한다
- WHEN stdin이 비어있으면, 시스템은 빈 JSON 객체 `{}`를 파싱하고 기본 HookResult()를 반환해야 한다

### State-driven Requirements (상태 기반)

- WHILE stdin에서 데이터를 읽는 중, 시스템은 플랫폼별 차이를 추상화해야 한다
- WHILE 데이터를 수신하는 동안, 시스템은 블로킹 모드로 동작해야 한다
- WHILE 핸들러가 실행 중일 때, 시스템은 모든 에러를 stderr로 출력해야 한다

### Optional Features (선택적 기능)

- WHERE stdin 읽기가 타임아웃될 경우, 시스템은 select 모듈을 사용하여 대기할 수 있다 (대안 방안)

### Constraints (제약사항)

- stdin 읽기는 타임아웃 없이 데이터를 완전히 읽어야 한다 (Claude Code가 데이터 전송 완료 보장)
- 모든 에러는 stderr로 출력하고 stdout은 JSON만 출력해야 한다
- stdin 읽기 로직은 기존 핸들러 로직과 분리되어야 한다
- IF stdin이 비어있으면, 시스템은 JSONDecodeError를 발생시키지 않고 빈 객체 `{}`로 처리해야 한다
- 코드 변경은 기존 테스트를 통과해야 하며, 회귀가 없어야 한다

---

## 기술 사양 (Technical Specifications)

### 권장 해결 방안: Iterator 패턴

**구현 위치**: `.claude/hooks/alfred/alfred_hooks.py:123`

**변경 전**:
```python
input_data = sys.stdin.read()
data = json.loads(input_data)
```

**변경 후**:
```python
# Iterator 패턴으로 stdin 읽기 (크로스 플랫폼 호환)
input_data = ""
for line in sys.stdin:
    input_data += line

# 빈 stdin 처리
if not input_data.strip():
    input_data = "{}"

data = json.loads(input_data)
```

**근거**:
- Python 표준 Iterator 패턴은 모든 플랫폼에서 EOF를 일관되게 처리
- Windows/Unix 계열 모두 `for line in sys.stdin` 동작 보장
- 추가 라이브러리 불필요
- 기존 코드 구조 유지 가능

### 대안 1: Binary 모드

```python
input_data = sys.stdin.buffer.read().decode('utf-8')
```

**장점**:
- 명시적 인코딩 처리
- 바이너리 레벨 제어

**단점**:
- 인코딩 오류 처리 필요
- 복잡도 증가

### 대안 2: select 모듈 (타임아웃 추가)

```python
import select

if select.select([sys.stdin], [], [], 1.0)[0]:
    input_data = sys.stdin.read()
else:
    input_data = "{}"
```

**장점**:
- 타임아웃 제어 가능

**단점**:
- Windows에서 select는 소켓만 지원 (파일/파이프 미지원)
- 크로스 플랫폼 호환성 낮음
- Claude Code가 타임아웃 없이 데이터 전송 보장하므로 불필요

---

## 성능 및 보안 고려사항

### 성능
- stdin 읽기는 일회성 작업으로 성능 영향 무시 가능
- Iterator 패턴은 메모리 효율적 (라인 단위 처리)

### 보안
- stdin 입력은 Claude Code가 제어하므로 신뢰 가능
- JSON 파싱 실패 시 명확한 에러 처리로 악의적 입력 방지

### 호환성
- Python 3.8+ 모든 버전 지원
- Windows 10+, macOS 10.15+, Ubuntu 20.04+ 검증 필요

---

## 구현 범위 (Scope)

### In Scope
- `.claude/hooks/alfred/alfred_hooks.py` stdin 읽기 로직 개선
- 빈 stdin 처리 로직 추가
- 단위 테스트 작성 (Windows/Mac/Linux 시뮬레이션)
- 통합 테스트 (실제 Claude Code Hook 호출)

### Out of Scope
- 다른 핸들러 로직 변경 (SessionStart, PreToolUse 등)
- JSON 스키마 변경
- 에러 메시지 포맷 변경

---

## 테스트 계획

### 단위 테스트
- Windows 환경 stdin 읽기 시뮬레이션
- macOS/Linux 환경 stdin 읽기 시뮬레이션
- 빈 stdin 처리
- JSON 파싱 오류 처리

### 통합 테스트
- 실제 Claude Code SessionStart 이벤트
- 실제 Claude Code PreToolUse 이벤트
- 크로스 플랫폼 CI/CD 검증 (GitHub Actions)

---

## 참고 자료

- GitHub Issue #25: Windows stdin 문제 보고
- GitHub Issue #31: 스크립트 오류 관련
- Python 공식 문서: [sys.stdin](https://docs.python.org/3/library/sys.html#sys.stdin)
- Python Iterator 패턴: [PEP 234](https://peps.python.org/pep-0234/)

---

## Traceability (추적성)

- **SPEC**: `@SPEC:WINDOWS-HOOKS-001`
- **TEST**: `tests/hooks/test_alfred_hooks_stdin.py` (신규 생성 예정)
- **CODE**: `.claude/hooks/alfred/alfred_hooks.py:123`
- **DOC**: `.moai/specs/SPEC-WINDOWS-HOOKS-001/spec.md`

---

**작성일**: 2025-10-18
**작성자**: @Goos
**버전**: v0.0.1 (INITIAL)
