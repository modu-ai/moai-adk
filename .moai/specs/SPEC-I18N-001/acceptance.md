# SPEC-I18N-001 인수 기준 (Acceptance Criteria)

> **SPEC**: 다국어 템플릿 시스템 (한/영)
> **버전**: v0.0.1
> **상태**: draft

---

## 1. 기능 요구사항 (Functional Requirements)

### FR-1: 템플릿 디렉토리 분리

**Given**: MoAI-ADK 템플릿 디렉토리가 존재한다
**When**: 템플릿 디렉토리를 언어별로 분리한다
**Then**:
- ✅ `.claude-ko/` 디렉토리가 존재한다
- ✅ `.claude-en/` 디렉토리가 존재한다
- ✅ 두 디렉토리의 파일 구조가 동일하다
- ✅ 기존 `.claude/` 디렉토리는 제거된다

**검증 방법**:
```bash
# 디렉토리 존재 확인
ls src/moai_adk/templates/.claude-ko/
ls src/moai_adk/templates/.claude-en/

# 구조 비교
diff <(cd src/moai_adk/templates/.claude-ko && find . -type f | sort) \
     <(cd src/moai_adk/templates/.claude-en && find . -type f | sort)

# 결과: 차이 없음 (0개)
```

---

### FR-2: init 언어 선택 프롬프트

**Given**: 사용자가 `moai-adk init` 명령을 실행한다
**When**: CLI가 언어 선택 프롬프트를 표시한다
**Then**:
- ✅ 선택지 2개가 표시된다: "Korean (한국어)" / "English"
- ✅ 기본값은 "Korean (한국어)"이다
- ✅ 사용자가 선택을 완료하면 다음 단계로 진행한다

**검증 방법**:
```bash
# 테스트 실행
pytest tests/integration/test_init_i18n.py::test_init_language_prompt

# 예상 출력
? Select your preferred language for project templates:
  > Korean (한국어)
    English
```

---

### FR-3: 선택한 언어 템플릿 복사

**Given**: 사용자가 언어를 선택했다
**When**: init 프로세스가 템플릿을 복사한다
**Then**:
- ✅ 한국어 선택 시 `.claude-ko/` → `.claude/` 복사
- ✅ 영어 선택 시 `.claude-en/` → `.claude/` 복사
- ✅ 복사 완료 메시지가 영어로 표시된다

**검증 방법**:
```python
# 단위 테스트
def test_copy_korean_template(tmp_path):
    processor = TemplateProcessor(tmp_path)
    processor.copy_claude_template("ko")

    # 검증
    assert (tmp_path / ".claude" / "commands" / "1-spec.md").exists()

def test_copy_english_template(tmp_path):
    processor = TemplateProcessor(tmp_path)
    processor.copy_claude_template("en")

    # 검증
    assert (tmp_path / ".claude" / "commands" / "1-spec.md").exists()
```

---

### FR-4: config.json에 locale 저장

**Given**: 템플릿 복사가 완료되었다
**When**: init 프로세스가 config.json을 생성한다
**Then**:
- ✅ `.moai/config.json` 파일이 생성된다
- ✅ `project.locale` 필드에 선택한 언어 코드가 저장된다
- ✅ 값은 "ko" 또는 "en"이다

**검증 방법**:
```bash
# config.json 확인
cat .moai/config.json | jq '.project.locale'

# 예상 결과 (한국어 선택 시)
"ko"

# 예상 결과 (영어 선택 시)
"en"
```

---

### FR-5: CLI 메시지 영어 유지

**Given**: 모든 CLI 명령어를 실행한다
**When**: CLI가 메시지를 출력한다
**Then**:
- ✅ 모든 명령어 이름은 영어이다
- ✅ 모든 로그 메시지는 영어이다
- ✅ 모든 에러 메시지는 영어이다

**검증 방법**:
```bash
# 명령어 확인
moai-adk --help  # 모두 영어

# 로그 메시지 확인
moai-adk init  # "Initializing project..."

# 에러 메시지 확인
moai-adk init --mode invalid  # "Invalid mode: invalid"
```

---

### FR-6: README 구조 변경

**Given**: 프로젝트 루트에 README 파일이 존재한다
**When**: README 파일을 확인한다
**Then**:
- ✅ `README.md`는 한국어로 작성되어 있다
- ✅ `README.en.md`는 영어로 작성되어 있다
- ✅ `README.md`에 영어 버전 링크가 있다
- ✅ `README.en.md`에 한국어 버전 링크가 있다

**검증 방법**:
```bash
# README.md 확인
grep "🇺🇸 English" README.md  # 링크 존재 확인

# README.en.md 확인
grep "🇰🇷 한국어" README.en.md  # 링크 존재 확인
```

---

## 2. 비기능 요구사항 (Non-Functional Requirements)

### NFR-1: 성능

**Given**: 사용자가 `moai-adk init`을 실행한다
**When**: 템플릿 복사 프로세스가 실행된다
**Then**:
- ✅ 전체 init 프로세스 ≤5초
- ✅ 템플릿 복사 단계 ≤1초
- ✅ 메모리 증가 ≤10MB

**검증 방법**:
```bash
# 시간 측정
time moai-adk init --mode personal --locale ko

# 예상 결과
real    0m3.5s
user    0m0.5s
sys     0m0.2s
```

---

### NFR-2: 호환성

**Given**: Python 3.10 이상 환경이다
**When**: MoAI-ADK를 설치하고 실행한다
**Then**:
- ✅ Python 3.10에서 정상 동작
- ✅ Python 3.11에서 정상 동작
- ✅ Python 3.12에서 정상 동작
- ✅ macOS, Linux, Windows에서 정상 동작

**검증 방법**:
```bash
# Python 버전별 테스트
pytest tests/ --python=3.10
pytest tests/ --python=3.11
pytest tests/ --python=3.12
```

---

### NFR-3: 유지보수성

**Given**: 새로운 언어를 추가해야 한다
**When**: 개발자가 새 템플릿 디렉토리를 추가한다
**Then**:
- ✅ `.claude-{locale}/` 디렉토리만 추가하면 됨
- ✅ processor.py에 locale 추가 (1줄)
- ✅ 기존 코드 수정 최소화

**검증 방법**:
```python
# processor.py 수정 예시
SUPPORTED_LOCALES = ["ko", "en", "ja"]  # ← ja 추가 (1줄)
```

---

## 3. 품질 게이트 (Quality Gates)

### QG-1: 테스트 커버리지

**조건**:
- ✅ 단위 테스트 커버리지 ≥85%
- ✅ 통합 테스트 커버리지 ≥80%
- ✅ 전체 테스트 커버리지 ≥85%

**검증 방법**:
```bash
pytest --cov=moai_adk --cov-report=term-missing

# 예상 결과
Name                                      Stmts   Miss  Cover
-------------------------------------------------------------
moai_adk/core/template/processor.py        120      8    93%
moai_adk/core/init/commands.py              80      5    94%
-------------------------------------------------------------
TOTAL                                       200     13    93%
```

---

### QG-2: 코드 품질

**조건**:
- ✅ mypy 타입 검사 통과 (0 에러)
- ✅ ruff 린트 검사 통과 (0 경고)
- ✅ 함수 ≤50 LOC
- ✅ 복잡도 ≤10

**검증 방법**:
```bash
# mypy 검사
mypy src/moai_adk/

# ruff 검사
ruff check src/moai_adk/

# 복잡도 검사
radon cc src/moai_adk/ -a

# 예상 결과: 모두 통과
```

---

### QG-3: 문서화

**조건**:
- ✅ README.md 완성 (한국어)
- ✅ README.en.md 완성 (영어)
- ✅ docs/i18n-guide.md 작성
- ✅ API Reference 업데이트

**검증 방법**:
```bash
# 문서 존재 확인
ls README.md README.en.md docs/i18n-guide.md

# 링크 유효성 검증
markdown-link-check README.md
markdown-link-check README.en.md
```

---

## 4. 수락 시나리오 (Acceptance Scenarios)

### 시나리오 1: 한국어 템플릿 선택

```gherkin
Feature: 한국어 템플릿 프로젝트 초기화

  Scenario: 사용자가 한국어 템플릿을 선택한다
    Given 사용자가 새 디렉토리에 있다
    When 사용자가 "moai-adk init" 명령을 실행한다
    And 언어 선택 프롬프트에서 "Korean (한국어)"를 선택한다
    Then ".claude/" 디렉토리가 생성된다
    And ".claude/commands/1-spec.md" 파일이 한국어로 작성되어 있다
    And ".moai/config.json"에 "locale": "ko"가 저장된다
```

---

### 시나리오 2: 영어 템플릿 선택

```gherkin
Feature: 영어 템플릿 프로젝트 초기화

  Scenario: 사용자가 영어 템플릿을 선택한다
    Given 사용자가 새 디렉토리에 있다
    When 사용자가 "moai-adk init" 명령을 실행한다
    And 언어 선택 프롬프트에서 "English"를 선택한다
    Then ".claude/" 디렉토리가 생성된다
    And ".claude/commands/1-spec.md" 파일이 영어로 작성되어 있다
    And ".moai/config.json"에 "locale": "en"이 저장된다
```

---

### 시나리오 3: 지원되지 않는 언어 Fallback

```gherkin
Feature: 지원되지 않는 언어 처리

  Scenario: 사용자가 지원되지 않는 언어를 입력한다
    Given 사용자가 새 디렉토리에 있다
    When 사용자가 "moai-adk init --locale ja" 명령을 실행한다
    Then 경고 메시지가 표시된다: "Unsupported locale 'ja', falling back to English"
    And ".claude-en/" 템플릿이 복사된다
    And ".moai/config.json"에 "locale": "en"이 저장된다
```

---

## 5. 회귀 테스트 (Regression Tests)

### RT-1: 기존 기능 호환성

**조건**:
- ✅ `moai-adk doctor` 정상 동작
- ✅ `moai-adk status` 정상 동작
- ✅ `/alfred:1-spec` 정상 동작
- ✅ `/alfred:2-build` 정상 동작
- ✅ `/alfred:3-sync` 정상 동작

**검증 방법**:
```bash
# 전체 워크플로우 테스트
moai-adk init --locale ko
/alfred:1-spec "테스트 기능"
/alfred:2-build TEST-001
/alfred:3-sync

# 모두 정상 동작 확인
```

---

## 6. Definition of Done (DoD)

### 완료 조건 체크리스트

- [ ] **코드**:
  - [ ] 모든 단위 테스트 통과 (≥85% 커버리지)
  - [ ] 모든 통합 테스트 통과
  - [ ] mypy, ruff 검사 통과
  - [ ] 코드 리뷰 완료 (최소 1명)

- [ ] **문서**:
  - [ ] README.md 한국어 완성
  - [ ] README.en.md 영어 완성
  - [ ] docs/i18n-guide.md 작성
  - [ ] API Reference 업데이트

- [ ] **테스트**:
  - [ ] 한국어 템플릿 E2E 테스트 통과
  - [ ] 영어 템플릿 E2E 테스트 통과
  - [ ] 회귀 테스트 통과

- [ ] **배포**:
  - [ ] CHANGELOG.md 업데이트
  - [ ] Git 태그 생성 (v0.0.1)
  - [ ] PyPI 배포 (alpha)

---

## 7. 사용자 피드백 기준

### 만족도 지표

**목표**:
- ✅ GitHub Issues: 긍정적 피드백 ≥80%
- ✅ 템플릿 품질: 번역 오류 0건
- ✅ 사용성: 언어 전환 성공률 100%

**측정 방법**:
- GitHub Issues 분석
- 커뮤니티 설문조사
- 사용 로그 분석

---

_이 인수 기준은 SPEC-I18N-001 구현 완료를 판단하는 기준입니다._
