# 구현 계획: Windows 환경 stdin 처리 개선

> **상태**: draft (v0.0.1)
> **작성일**: 2025-10-18

---

## 현재 문제점 분석

### 1. 코드 위치 및 현재 구현

**파일**: `.claude/hooks/alfred/alfred_hooks.py`
**라인**: 123

```python
# 현재 구현 (main 함수 내부)
try:
    # Read JSON from stdin
    input_data = sys.stdin.read()  # ⚠️ 문제: Windows EOF 처리 불확실
    data = json.loads(input_data)

    cwd = data.get("cwd", ".")
    # ...
except json.JSONDecodeError as e:
    print(f"JSON parse error: {e}", file=sys.stderr)
    sys.exit(1)
```

### 2. 문제점 상세

| 문제                        | 설명                                                                         | 영향                                       |
| --------------------------- | ---------------------------------------------------------------------------- | ------------------------------------------ |
| **Windows EOF 처리**        | `sys.stdin.read()`는 Windows에서 EOF 시그널을 일관되게 처리하지 못할 수 있음 | GitHub Issue #25, #31 보고된 스크립트 오류 |
| **빈 stdin 처리**           | stdin이 비어있을 때 `json.loads("")`는 JSONDecodeError 발생                  | Claude Code가 빈 페이로드 전송 시 오류     |
| **크로스 플랫폼 검증 부족** | Windows/Mac/Linux 동작 차이 미검증                                           | 특정 환경에서만 동작 불안정                |

### 3. GitHub Issue 분석

**Issue #25**: Windows 환경에서 Hook 스크립트 오류
- 증상: stdin 읽기 실패, 스크립트 종료
- 재현: Windows 10 + Claude Code + SessionStart 이벤트
- 빈도: 간헐적 (stdin 타이밍 문제로 추정)

**Issue #31**: 스크립트 실행 오류 (관련성 검토 필요)
- 증상: JSON 파싱 오류
- 재현: 불명확 (stdin 관련 가능성)

---

## 해결 방안

### 권장 방안: Iterator 패턴

**장점**:
- ✅ Python 표준 Iterator 패턴 (추가 라이브러리 불필요)
- ✅ Windows/Unix 계열 모두 EOF 일관 처리
- ✅ 코드 가독성 우수
- ✅ 기존 구조 최소 변경

**구현 코드**:
```python
try:
    # Iterator 패턴으로 stdin 읽기 (크로스 플랫폼 호환)
    input_data = ""
    for line in sys.stdin:
        input_data += line

    # 빈 stdin 처리
    if not input_data.strip():
        input_data = "{}"

    data = json.loads(input_data)
    cwd = data.get("cwd", ".")
    # ...
except json.JSONDecodeError as e:
    print(f"JSON parse error: {e}", file=sys.stderr)
    sys.exit(1)
except Exception as e:
    print(f"Unexpected error: {e}", file=sys.stderr)
    sys.exit(1)
```

### 대안 비교

| 방안              | 장점                          | 단점                               | 선택       |
| ----------------- | ----------------------------- | ---------------------------------- | ---------- |
| **Iterator 패턴** | 표준 방식, 크로스 플랫폼 보장 | -                                  | ✅ **권장** |
| **Binary 모드**   | 명시적 인코딩 제어            | 복잡도 증가, 인코딩 오류 처리 필요 | ❌          |
| **select 모듈**   | 타임아웃 제어 가능            | Windows 미지원 (소켓만 지원)       | ❌          |

---

## TDD 계획 (RED-GREEN-REFACTOR)

### Phase 1: RED (테스트 작성 - 실패)

**테스트 파일**: `tests/hooks/test_alfred_hooks_stdin.py` (신규 생성)

**테스트 케이스**:

1. **test_stdin_normal_json_windows**
   - 시뮬레이션: Windows 환경, 정상 JSON 입력
   - 예상: 성공적으로 JSON 파싱

2. **test_stdin_empty_input**
   - 시뮬레이션: 빈 stdin (`""`)
   - 예상: 빈 JSON 객체 `{}` 파싱, exit code 0

3. **test_stdin_invalid_json**
   - 시뮬레이션: 잘못된 JSON (`{invalid}`)
   - 예상: JSONDecodeError, exit code 1, stderr 메시지

4. **test_stdin_cross_platform**
   - 시뮬레이션: macOS/Linux 환경, 동일 JSON
   - 예상: Windows와 동일 동작

**실행**:
```bash
pytest tests/hooks/test_alfred_hooks_stdin.py -v
# 예상: 모든 테스트 FAILED (구현 전)
```

### Phase 2: GREEN (최소 구현 - 통과)

**구현 위치**: `.claude/hooks/alfred/alfred_hooks.py:123`

**변경 내용**:
```python
# Before
input_data = sys.stdin.read()

# After
input_data = ""
for line in sys.stdin:
    input_data += line

if not input_data.strip():
    input_data = "{}"
```

**실행**:
```bash
pytest tests/hooks/test_alfred_hooks_stdin.py -v
# 예상: 모든 테스트 PASSED
```

### Phase 3: REFACTOR (리팩토링 - 품질 개선)

**개선 항목**:

1. **에러 처리 강화**:
   ```python
   try:
       # stdin 읽기
       input_data = ""
       for line in sys.stdin:
           input_data += line

       # 빈 stdin 처리
       if not input_data.strip():
           input_data = "{}"

       data = json.loads(input_data)
   except json.JSONDecodeError as e:
       print(f"JSON parse error: {e}", file=sys.stderr)
       sys.exit(1)
   except OSError as e:
       print(f"stdin read error: {e}", file=sys.stderr)
       sys.exit(1)
   except Exception as e:
       print(f"Unexpected error: {e}", file=sys.stderr)
       sys.exit(1)
   ```

2. **코드 가독성 개선**:
   - stdin 읽기 로직 주석 추가
   - 변수명 명확화

3. **docstring 업데이트**:
   ```python
   def main() -> None:
       """메인 진입점 - Claude Code Hook 스크립트

       stdin으로부터 JSON 페이로드를 읽고, 이벤트 핸들러를 실행합니다.

       stdin 읽기:
           - Iterator 패턴 사용 (크로스 플랫폼 호환)
           - 빈 stdin은 빈 JSON 객체로 처리
           - EOF는 플랫폼 무관하게 일관 처리

       ...
       """
   ```

**실행**:
```bash
pytest tests/hooks/test_alfred_hooks_stdin.py -v
# 예상: 모든 테스트 여전히 PASSED (회귀 없음)
```

---

## 우선순위별 작업 단계

### 1차 목표: 핵심 기능 구현
- [ ] 테스트 파일 생성 (`tests/hooks/test_alfred_hooks_stdin.py`)
- [ ] RED: 4개 테스트 케이스 작성 (Windows, 빈 stdin, 잘못된 JSON, 크로스 플랫폼)
- [ ] GREEN: Iterator 패턴으로 stdin 읽기 구현
- [ ] 빈 stdin 처리 로직 추가

### 2차 목표: 품질 개선
- [ ] REFACTOR: 에러 처리 강화 (OSError 처리)
- [ ] docstring 업데이트
- [ ] 코드 주석 추가

### 3차 목표: 통합 검증
- [ ] 실제 Claude Code SessionStart 이벤트 테스트
- [ ] Windows 환경 실제 검증 (가능 시)
- [ ] CI/CD 크로스 플랫폼 검증 (GitHub Actions)

---

## 기술적 접근 방법

### stdin 읽기 패턴 비교

**현재 방식 (read)**:
```python
input_data = sys.stdin.read()  # EOF까지 블로킹
```
- 문제: Windows에서 EOF 처리 불확실

**Iterator 방식 (권장)**:
```python
input_data = ""
for line in sys.stdin:  # 라인 단위 읽기, EOF 자동 처리
    input_data += line
```
- 해결: Python Iterator 프로토콜이 EOF 일관 처리

**Binary 방식**:
```python
input_data = sys.stdin.buffer.read().decode('utf-8')
```
- 복잡: 인코딩 오류 처리 필요

### 빈 stdin 처리 전략

**문제**:
```python
json.loads("")  # JSONDecodeError: Expecting value
```

**해결**:
```python
if not input_data.strip():
    input_data = "{}"  # 빈 JSON 객체

data = json.loads(input_data)
# data == {} → HookResult() 반환 (blocked=False)
```

---

## 리스크 및 대응 방안

### 리스크 1: 기존 테스트 회귀

**리스크**:
- stdin 읽기 로직 변경으로 기존 테스트 실패 가능

**대응**:
- 기존 테스트 전체 실행 (`pytest tests/hooks/ -v`)
- 실패 시 Iterator 패턴 검증 및 수정

### 리스크 2: Windows 실제 환경 검증 부족

**리스크**:
- 개발 환경이 macOS인 경우 Windows 검증 어려움

**대응**:
- GitHub Actions에서 Windows CI 추가
- 단위 테스트로 Windows 동작 시뮬레이션

### 리스크 3: Claude Code 동작 변경

**리스크**:
- Claude Code가 stdin 전송 방식 변경 가능성

**대응**:
- Iterator 패턴은 표준 방식으로 향후 변경 대응 가능
- 테스트 케이스로 회귀 방지

---

## 아키텍처 설계 방향

### 변경 범위

```
.claude/hooks/alfred/
├── alfred_hooks.py          # ✅ 수정: stdin 읽기 로직 (Iterator 패턴)
├── core/
│   ├── __init__.py
│   ├── context.py
│   ├── project.py
│   └── ...
└── handlers/
    ├── session.py
    └── ...

tests/hooks/
└── test_alfred_hooks_stdin.py  # ✅ 신규: stdin 테스트 전용
```

### 모듈 의존성

```
alfred_hooks.py (main)
    ├── sys.stdin (Iterator 패턴)
    ├── json.loads (파싱)
    └── handlers/* (이벤트 라우팅)
```

**변경 없음**:
- `core/*` 모듈
- `handlers/*` 모듈
- 기존 테스트 (`tests/integration/test_cli_*.py`)

---

## 검증 체크리스트

### 코드 품질
- [ ] 함수 ≤ 50 LOC (main 함수 현재 30 LOC, 변경 후 35 LOC 이하)
- [ ] 파일 ≤ 300 LOC (alfred_hooks.py 현재 161 LOC)
- [ ] 복잡도 ≤ 10 (stdin 읽기 로직 단순)

### 테스트
- [ ] 단위 테스트 커버리지 ≥ 85% (stdin 읽기 로직)
- [ ] 통합 테스트 통과 (기존 테스트 회귀 없음)
- [ ] 크로스 플랫폼 CI 통과 (Windows/macOS/Linux)

### 문서
- [ ] docstring 업데이트
- [ ] HISTORY 섹션 업데이트 (v0.0.1 → v0.1.0 구현 완료 시)
- [ ] README/CLAUDE.md 업데이트 불필요 (내부 구현 변경)

---

**작성일**: 2025-10-18
**다음 단계**: `/alfred:2-run WINDOWS-HOOKS-001` 실행
