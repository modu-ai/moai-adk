# SPEC-CLI-001 수락 기준

## Given-When-Then 테스트 시나리오

### 시나리오 1: moai 명령 실행
**Given**: moai-adk가 설치됨
**When**: `moai --version` 명령을 실행함
**Then**:
- [ ] "0.3.0" 버전이 출력됨
- [ ] 에러 없이 종료됨

### 시나리오 2: moai init 실행
**Given**: 빈 디렉토리가 존재함
**When**: `moai init .` 명령을 실행함
**Then**:
- [ ] ".moai/" 디렉토리가 생성됨
- [ ] "✓ Project initialized successfully" 메시지가 출력됨
- [ ] 종료 코드 0

### 시나리오 3: moai doctor 실행
**Given**: 프로젝트가 초기화됨
**When**: `moai doctor` 명령을 실행함
**Then**:
- [ ] 시스템 요구사항 체크 결과가 표시됨
- [ ] 각 항목에 ✓ 또는 ✗ 아이콘이 표시됨
- [ ] Python 버전 확인 결과 포함

### 시나리오 4: moai status 실행
**Given**: 프로젝트가 초기화됨
**When**: `moai status` 명령을 실행함
**Then**:
- [ ] 프로젝트 모드(personal/team) 표시됨
- [ ] Locale 설정 표시됨
- [ ] SPEC 개수 표시됨

### 시나리오 5: moai --help 실행
**Given**: moai-adk가 설치됨
**When**: `moai --help` 명령을 실행함
**Then**:
- [ ] ASCII 로고가 표시됨
- [ ] 4개 명령어 목록이 표시됨
- [ ] 각 명령어의 설명이 표시됨

---

## 품질 게이트 기준

### 1. 명령어 완성도
- [ ] 4개 명령어 모두 구현됨
- [ ] --help 옵션 모두 작동함
- [ ] 에러 메시지 명확함

### 2. Rich 출력 품질
- [ ] 색상이 올바르게 표시됨
- [ ] 아이콘(✓, ✗, ℹ, ⚠)이 일관됨
- [ ] ASCII 로고가 정상 표시됨

### 3. 에러 처리
- [ ] FileNotFoundError 처리
- [ ] PermissionError 처리
- [ ] 잘못된 인자 처리

---

## 검증 방법 및 도구

### 자동화 테스트
```python
# tests/cli/test_main.py
from click.testing import CliRunner
from moai_adk.cli.main import cli

def test_version():
    runner = CliRunner()
    result = runner.invoke(cli, ['--version'])
    assert result.exit_code == 0
    assert '0.3.0' in result.output

def test_init():
    runner = CliRunner()
    with runner.isolated_filesystem():
        result = runner.invoke(cli, ['init', '.'])
        assert result.exit_code == 0
        assert '✓' in result.output
```

### 수동 검증
1. **명령어 실행**: 각 명령어 터미널에서 실행
2. **출력 확인**: 색상, 아이콘, 메시지 정확성
3. **에러 테스트**: 잘못된 인자로 실행

---

## 완료 조건 (Definition of Done)

### 필수 조건
- [ ] 4개 명령어 구현 완료
- [ ] ASCII 로고 표시
- [ ] Rich 출력 적용
- [ ] Click 테스트 통과

### 선택 조건
- [ ] Progress Bar 추가 (긴 작업)
- [ ] Autocomplete 지원 (bash/zsh)

### 문서화
- [ ] 명령어 사용법 README에 추가
- [ ] 예제 스크린샷 첨부
