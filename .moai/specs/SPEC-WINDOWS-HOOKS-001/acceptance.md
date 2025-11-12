# Acceptance Criteria: Windows 환경 stdin 처리 개선

> **상태**: draft (v0.0.1)
> **작성일**: 2025-10-18

---

## 수락 기준 개요


---

## Given-When-Then 시나리오

### 시나리오 1: Windows 환경에서 정상 JSON 입력 처리

**Given**:
- Windows 10 환경에서 Claude Code 실행
- SessionStart 이벤트 발생
- 유효한 JSON 페이로드 stdin으로 전송
  ```json
  {"cwd": "/path/to/project", "event": "SessionStart"}
  ```

**When**:
- `python alfred_hooks.py SessionStart` 실행
- stdin으로 JSON 페이로드 읽기

**Then**:
- stdin에서 JSON을 성공적으로 읽고 파싱
- 프로젝트 상태 메시지를 stdout에 JSON 형식으로 출력
- exit code 0 반환
- stderr에 에러 메시지 없음

**검증 방법**:
```bash
# Windows PowerShell
echo '{"cwd": ".", "event": "SessionStart"}' | python .claude/hooks/alfred/alfred_hooks.py SessionStart
# 예상: JSON 출력, exit code 0
```

**자동화 테스트**:
```python
def test_stdin_normal_json_windows(monkeypatch):
    """Windows 환경에서 정상 JSON 입력 처리"""
    test_input = '{"cwd": ".", "event": "SessionStart"}\n'
    monkeypatch.setattr('sys.stdin', io.StringIO(test_input))
    monkeypatch.setattr('sys.argv', ['alfred_hooks.py', 'SessionStart'])

    with pytest.raises(SystemExit) as exc_info:
        main()

    assert exc_info.value.code == 0
```

---

### 시나리오 2: 빈 stdin 처리

**Given**:
- Claude Code가 빈 stdin 전송 (엣지 케이스)
- stdin 내용: `""` (빈 문자열)

**When**:
- `python alfred_hooks.py SessionStart` 실행
- stdin이 비어있음

**Then**:
- 시스템은 빈 JSON 객체 `{}`를 파싱
- 기본 HookResult()를 반환 (`{"blocked": false}`)
- exit code 0 반환
- stderr에 에러 메시지 없음

**검증 방법**:
```bash
# Unix/macOS
echo -n '' | python .claude/hooks/alfred/alfred_hooks.py SessionStart
# 예상: {"blocked": false}, exit code 0

# Windows PowerShell
[Console]::In.Read() | python .claude/hooks/alfred/alfred_hooks.py SessionStart
# 예상: {"blocked": false}, exit code 0
```

**자동화 테스트**:
```python
def test_stdin_empty_input(monkeypatch):
    """빈 stdin 처리"""
    test_input = ''  # 빈 문자열
    monkeypatch.setattr('sys.stdin', io.StringIO(test_input))
    monkeypatch.setattr('sys.argv', ['alfred_hooks.py', 'SessionStart'])

    with pytest.raises(SystemExit) as exc_info:
        main()

    assert exc_info.value.code == 0
```

---

### 시나리오 3: JSON 파싱 오류 처리

**Given**:
- stdin에 잘못된 JSON 형식 제공
- stdin 내용: `{invalid json}`

**When**:
- `python alfred_hooks.py SessionStart` 실행
- JSON 파싱 시도

**Then**:
- JSONDecodeError를 stderr에 출력
  ```
  JSON parse error: Expecting property name enclosed in double quotes: line 1 column 2 (char 1)
  ```
- exit code 1 반환
- stdout에 출력 없음

**검증 방법**:
```bash
echo '{invalid json}' | python .claude/hooks/alfred/alfred_hooks.py SessionStart
# 예상: stderr에 에러 메시지, exit code 1
```

**자동화 테스트**:
```python
def test_stdin_invalid_json(monkeypatch, capsys):
    """JSON 파싱 오류 처리"""
    test_input = '{invalid json}'
    monkeypatch.setattr('sys.stdin', io.StringIO(test_input))
    monkeypatch.setattr('sys.argv', ['alfred_hooks.py', 'SessionStart'])

    with pytest.raises(SystemExit) as exc_info:
        main()

    assert exc_info.value.code == 1
    captured = capsys.readouterr()
    assert 'JSON parse error' in captured.err
```

---

### 시나리오 4: 크로스 플랫폼 호환성

**Given**:
- macOS/Linux 환경에서 동일한 JSON 페이로드 전송
- stdin 내용: `{"cwd": ".", "event": "SessionStart"}`

**When**:
- `python alfred_hooks.py SessionStart` 실행

**Then**:
- Windows와 동일한 동작 보장 (회귀 없음)
- JSON 정상 파싱
- exit code 0 반환
- stderr에 에러 메시지 없음

**검증 방법**:
```bash
# macOS/Linux
echo '{"cwd": ".", "event": "SessionStart"}' | python .claude/hooks/alfred/alfred_hooks.py SessionStart
# 예상: JSON 출력, exit code 0
```

**자동화 테스트**:
```python
@pytest.mark.parametrize("platform", ["darwin", "linux", "win32"])
def test_stdin_cross_platform(monkeypatch, platform):
    """크로스 플랫폼 호환성"""
    monkeypatch.setattr('sys.platform', platform)
    test_input = '{"cwd": ".", "event": "SessionStart"}\n'
    monkeypatch.setattr('sys.stdin', io.StringIO(test_input))
    monkeypatch.setattr('sys.argv', ['alfred_hooks.py', 'SessionStart'])

    with pytest.raises(SystemExit) as exc_info:
        main()

    assert exc_info.value.code == 0
```

---

### 시나리오 5: 실제 Claude Code SessionStart 이벤트

**Given**:
- 실제 Claude Code 환경에서 SessionStart 이벤트 발생
- Claude Code가 stdin으로 페이로드 전송

**When**:
- Hook 스크립트 자동 실행

**Then**:
- 프로젝트 상태 메시지가 Claude Code UI에 표시
- 에러 없이 정상 실행
- GitHub Issue #25, #31 재현 불가

**검증 방법**:
1. Claude Code 실행
2. 프로젝트 디렉토리 열기
3. SessionStart Hook 자동 실행 확인
4. UI에 프로젝트 상태 메시지 표시 확인

**수동 검증 체크리스트**:
- [ ] Windows 10: SessionStart 정상 동작
- [ ] macOS: SessionStart 정상 동작 (회귀 없음)
- [ ] Linux: SessionStart 정상 동작 (회귀 없음)

---

## 품질 게이트 (Quality Gates)

### 코드 품질

- [ ] **TRUST - Test First**: 모든 시나리오에 대한 단위 테스트 작성 및 통과
  - `tests/hooks/test_alfred_hooks_stdin.py` 생성
  - 최소 4개 테스트 케이스 통과
  - 테스트 커버리지 ≥ 85% (stdin 읽기 로직)

- [ ] **TRUST - Readable**: 코드 가독성 기준 준수
  - 함수 ≤ 50 LOC (main 함수 35 LOC 이하)
  - 의도 드러내는 변수명 사용 (`input_data`, `line`)
  - 주석 추가 (Iterator 패턴 설명)

- [ ] **TRUST - Unified**: 아키텍처 일관성
  - 기존 핸들러 로직 변경 없음
  - stdin 읽기 로직만 수정 (SRP 준수)

- [ ] **TRUST - Secured**: 보안 기준
  - stdin 입력은 Claude Code 제어 (신뢰 가능)
  - JSON 파싱 오류 시 명확한 에러 처리

- [ ] **TRUST - Trackable**: 추적성
  - HISTORY 섹션 업데이트 (v0.1.0 구현 완료)

### 테스트 통과 기준

- [ ] **단위 테스트**: `pytest tests/hooks/test_alfred_hooks_stdin.py -v`
  - 모든 테스트 PASSED
  - 커버리지 ≥ 85%

- [ ] **회귀 테스트**: `pytest tests/hooks/ -v`
  - 기존 테스트 전체 PASSED (회귀 없음)

- [ ] **통합 테스트**: `pytest tests/integration/ -v`
  - CLI 테스트 전체 PASSED

- [ ] **CI/CD**: GitHub Actions
  - Windows/macOS/Linux 모두 PASSED

### 문서화 기준

- [ ] **docstring 업데이트**: `main()` 함수 docstring에 stdin 읽기 로직 설명 추가
- [ ] **HISTORY 업데이트**: SPEC 문서의 HISTORY 섹션에 v0.1.0 항목 추가

---

## Definition of Done (완료 조건)

다음 모든 조건이 만족되면 SPEC이 완료됩니다:

### 필수 조건 (Must Have)

1. **코드 구현 완료**
   - [ ] Iterator 패턴으로 stdin 읽기 구현
   - [ ] 빈 stdin 처리 로직 추가
   - [ ] 에러 처리 강화 (JSONDecodeError, OSError)

2. **테스트 통과**
   - [ ] 단위 테스트 4개 이상 작성 및 통과
   - [ ] 기존 테스트 전체 통과 (회귀 없음)
   - [ ] 테스트 커버리지 ≥ 85%

3. **품질 검증**
   - [ ] TRUST 5원칙 모두 준수
   - [ ] 코드 복잡도 ≤ 10
   - [ ] 함수 ≤ 50 LOC

4. **문서화**
   - [ ] docstring 업데이트
   - [ ] HISTORY 섹션 업데이트 (v0.1.0)

### 선택 조건 (Nice to Have)

5. **실제 환경 검증**
   - [ ] Windows 10 실제 테스트 (가능 시)
   - [ ] macOS 실제 테스트 (회귀 확인)
   - [ ] Linux 실제 테스트 (회귀 확인)

6. **GitHub Issue 종료**
   - [ ] Issue #25 종료 (Windows stdin 문제 해결)
   - [ ] Issue #31 검토 (관련성 확인)

---

## 검증 도구 및 명령어

### 로컬 테스트

```bash
# 1. 단위 테스트 실행
pytest tests/hooks/test_alfred_hooks_stdin.py -v

# 2. 회귀 테스트 실행
pytest tests/hooks/ -v

# 3. 전체 통합 테스트
pytest tests/integration/ -v

# 4. 커버리지 확인
pytest tests/hooks/test_alfred_hooks_stdin.py --cov=.claude/hooks/alfred --cov-report=html
```

### 수동 테스트

```bash
# Windows (PowerShell)
echo '{"cwd": ".", "event": "SessionStart"}' | python .claude/hooks/alfred/alfred_hooks.py SessionStart

# macOS/Linux
echo '{"cwd": ".", "event": "SessionStart"}' | python .claude/hooks/alfred/alfred_hooks.py SessionStart

# 빈 stdin 테스트
echo -n '' | python .claude/hooks/alfred/alfred_hooks.py SessionStart
```

### CI/CD 검증

```bash
# GitHub Actions 워크플로우 트리거
git push origin feature/SPEC-WINDOWS-HOOKS-001

# Windows/macOS/Linux 모두 PASSED 확인
```

---

## 롤백 조건

다음 상황 발생 시 변경사항을 롤백합니다:

1. **기존 테스트 실패**
   - macOS/Linux 환경에서 회귀 발생
   - 통합 테스트 실패

2. **성능 저하**
   - stdin 읽기 시간 >100ms 증가 (현재 <10ms)

3. **새로운 버그 발생**
   - Iterator 패턴으로 인한 새로운 오류 발견
   - Claude Code 호환성 문제

---

## 다음 단계 (Next Steps)

SPEC 완료 후:

1. **문서 동기화**: `/alfred:3-sync` 실행
   - Living Document 자동 생성
   - TAG 체인 검증
   - HISTORY 업데이트

2. **GitHub Issue 종료**
   - Issue #25에 해결 방법 및 테스트 결과 코멘트
   - Issue 종료

3. **PR 생성** (Team 모드인 경우)
   - Draft PR → Ready 전환
   - CI/CD 통과 확인
   - 코드 리뷰 요청

---

**작성일**: 2025-10-18
**승인 기준**: 모든 필수 조건 만족 + TRUST 5원칙 준수
