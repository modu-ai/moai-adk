
작성일: 2025-11-13
버전: 1.0.0

---

## 1. 수용 기준 (Acceptance Criteria)

### AC-001: 로거 객체 생성
**요구사항**: setup_logger() 함수가 Logger 객체를 반환하고 올바른 이름을 설정

**Given**
```
애플리케이션이 초기화되고 로거가 미생성된 상태
```

**When**
```python
logger = setup_logger("myapp")
```

**Then**
```
1. Logger 객체가 반환됨
2. logger.name == "myapp"
3. logger.handlers 리스트에 최소 2개의 핸들러 포함 (console + file)
```

**테스트 매핑**: `TestLoggerSetup.test_setup_logger_creates_logger()`

**정의된 완료**: ✅

---

### AC-002: 민감정보 마스킹 - API Key
**요구사항**: API Key 패턴 자동 감지 및 마스킹

**Given**
```
로거가 초기화되고 DEBUG/INFO/WARNING 레벨 설정된 상태
로그 파일 .moai/logs/moai.log가 생성됨
```

**When**
```python
logger.info("API Key: sk-1234567890abcdef")
```

**Then**
```
1. 콘솔에는 "***REDACTED***"로 출력
2. 로그 파일에도 "sk-1234567890abcdef" 원본 없이 "***REDACTED***"로 기록
3. 로그 포맷은 정상: [날짜] [INFO] [로거명] ... ***REDACTED*** ...
```

**테스트 매핑**: `TestSensitiveDataMasking.test_api_key_masking()`

**정의된 완료**: ✅

---

### AC-003: 민감정보 마스킹 - 이메일
**요구사항**: 이메일 주소 자동 감지 및 마스킹

**Given**
```
로거가 초기화된 상태
```

**When**
```python
logger.info("User email: user@example.com")
```

**Then**
```
1. 로그 파일에서 "user@example.com" 원본 없음
2. 대신 "***REDACTED***"로 기록
3. 나머지 메시지 "User email: "는 보존
```

**테스트 매핑**: `TestSensitiveDataMasking.test_email_masking()`

**정의된 완료**: ✅

---

### AC-004: 민감정보 마스킹 - 비밀번호
**요구사항**: password/passwd/pwd 키워드 뒤의 값 자동 감지 및 마스킹

**Given**
```
로거가 초기화된 상태
```

**When**
```python
logger.info("Password: secret123")
logger.info("passwd: abc123xyz")
logger.info("pwd=mypassword")
```

**Then**
```
1. 첫 번째: "Password: ***REDACTED***"로 기록 (키워드 보존)
2. 두 번째: "passwd: ***REDACTED***"로 기록
3. 세 번째: "pwd=***REDACTED***"로 기록
4. 원본 비밀번호 값은 어떤 로그에도 노출되지 않음
```

**테스트 매핑**: `TestSensitiveDataMasking.test_password_masking()`

**정의된 완료**: ✅

---

### AC-005: 환경별 로그 레벨 - Development
**요구사항**: MOAI_ENV=development 설정 시 DEBUG 레벨로 설정

**Given**
```
환경변수 MOAI_ENV=development
```

**When**
```python
logger = setup_logger("app")
```

**Then**
```
1. logger.level == logging.DEBUG (10)
2. DEBUG 레벨 메시지가 출력됨
3. INFO, WARNING, ERROR, CRITICAL 모두 출력됨
```

**테스트 매핑**: `TestLogLevelByEnvironment.test_development_mode_debug_level()`

**정의된 완료**: ✅

---

### AC-006: 환경별 로그 레벨 - Production
**요구사항**: MOAI_ENV=production 설정 시 WARNING 레벨로 설정

**Given**
```
환경변수 MOAI_ENV=production
```

**When**
```python
logger = setup_logger("app")
```

**Then**
```
1. logger.level == logging.WARNING (30)
2. INFO, DEBUG 메시지는 출력되지 않음
3. WARNING, ERROR, CRITICAL만 출력됨
```

**테스트 매핑**: `TestLogLevelByEnvironment.test_production_mode_warning_level()`

**정의된 완료**: ✅

---

### AC-007: 환경별 로그 레벨 - Test
**요구사항**: MOAI_ENV=test 설정 시 INFO 레벨로 설정

**Given**
```
환경변수 MOAI_ENV=test
```

**When**
```python
logger = setup_logger("app")
```

**Then**
```
1. logger.level == logging.INFO (20)
2. DEBUG는 출력 안 됨
3. INFO, WARNING, ERROR, CRITICAL은 출력됨
```

**테스트 매핑**: `TestLogLevelByEnvironment.test_test_mode_info_level()`

**정의된 완료**: ✅

---

### AC-008: 기본 로그 레벨 (환경변수 미설정)
**요구사항**: MOAI_ENV 환경변수 미설정 시 기본 INFO 레벨로 설정

**Given**
```
환경변수 MOAI_ENV가 설정되지 않은 상태
```

**When**
```python
logger = setup_logger("app")
```

**Then**
```
1. logger.level == logging.INFO (20) (기본값)
2. INFO, WARNING, ERROR, CRITICAL 출력
3. DEBUG는 출력 안 됨
```

**테스트 매핑**: `TestLogLevelByEnvironment.test_default_mode_info_level()`

**정의된 완료**: ✅

---

### AC-009: 로그 파일 자동 생성
**요구사항**: 로그 디렉토리가 없으면 자동 생성, 로그 파일 생성

**Given**
```
.moai/logs 디렉토리가 존재하지 않는 상태
```

**When**
```python
logger = setup_logger("app", log_dir=".moai/logs")
logger.info("Test message")
```

**Then**
```
1. .moai/logs 디렉토리가 자동 생성됨
2. .moai/logs/moai.log 파일이 생성됨
3. 로그 메시지가 파일에 기록됨
```

**테스트 매핑**: `TestLoggerSetup.test_setup_logger_creates_log_directory()`, `test_setup_logger_creates_log_file()`

**정의된 완료**: ✅

---

### AC-010: 콘솔 출력 확인
**요구사항**: 로그 메시지가 콘솔(stdout)에도 출력됨

**Given**
```
로거가 초기화되고 콘솔 캡처 가능한 상태
```

**When**
```python
logger.info("Console test")
```

**Then**
```
1. 콘솔(stdout)에 메시지 출력
2. 포맷: [YYYY-MM-DD HH:MM:SS] [INFO] [로거명] Console test
3. 민감정보는 마스킹됨
```

**테스트 매핑**: `TestConsoleHandler.test_console_handler_format()`

**정의된 완료**: ✅

---

### AC-011: 파일 출력 확인
**요구사항**: 로그 메시지가 파일에도 정확히 기록됨

**Given**
```
로거가 초기화되고 로그 파일이 생성된 상태
```

**When**
```python
logger.info("File test message")
```

**Then**
```
1. .moai/logs/moai.log 파일에 메시지 기록
2. 콘솔과 동일한 포맷으로 기록
3. 파일 인코딩: UTF-8
4. 모드: append (덮어쓰기 아님)
```

**테스트 매핑**: `TestFileHandler.test_file_handler_writes_to_file()`

**정의된 완료**: ✅

---

### AC-012: 기존 핸들러 중복 제거
**요구사항**: 로거 재초기화 시 기존 핸들러 제거

**Given**
```
로거가 이미 생성되고 핸들러가 있는 상태
logger = setup_logger("app")  # 첫 번째 호출
```

**When**
```python
logger = setup_logger("app")  # 재호출
```

**Then**
```
1. 기존 핸들러가 제거됨 (handlers.clear())
2. 새로운 핸들러만 추가됨
3. 총 핸들러 수: 2개 (console + file)
4. 중복 출력 없음
```

**테스트 매핑**: 기존 구현에서 `handlers.clear()` 호출로 확보

**정의된 완료**: ✅

---

### AC-013: 여러 민감정보 동시 마스킹
**요구사항**: 한 메시지에 여러 민감정보가 있으면 모두 마스킹

**Given**
```
로거가 초기화된 상태
```

**When**
```python
logger.info("API: sk-abc123, Email: test@test.com, Password: pass123")
```

**Then**
```
1. API Key "sk-abc123" → "***REDACTED***"
2. 이메일 "test@test.com" → "***REDACTED***"
3. 비밀번호 "pass123" → "***REDACTED***" (password: 키워드 포함)
4. 최종 로그: "API: ***REDACTED***, Email: ***REDACTED***, Password: ***REDACTED***"
5. 원본 민감정보는 어디에도 노출되지 않음
```

**테스트 매핑**: `TestSensitiveDataMasking.test_multiple_sensitive_data_masking()`

**정의된 완료**: ✅

---

### AC-014: 일반 로그 보존
**요구사항**: 민감정보 없는 메시지는 그대로 보존

**Given**
```
로거가 초기화된 상태
```

**When**
```python
logger.info("This is a normal log message")
logger.info("User login successful from 192.168.1.1")
```

**Then**
```
1. 메시지가 정확히 그대로 기록됨
2. 불필요한 마스킹 없음
3. 일반 텍스트 "192.168.1.1"는 보존 (IP주소는 민감정보 아님)
```

**테스트 매핑**: `TestSensitiveDataFilterClass.test_filter_preserves_non_sensitive_data()`

**정의된 완료**: ✅

---

### AC-015: 커스텀 로그 레벨 지정
**요구사항**: level 파라미터로 명시적 레벨 설정 시 환경변수 무시

**Given**
```
환경변수 MOAI_ENV=production (기본: WARNING)
```

**When**
```python
logger = setup_logger("app", level=logging.DEBUG)
```

**Then**
```
1. logger.level == logging.DEBUG
2. 환경변수 MOAI_ENV 설정 무시
3. DEBUG 메시지부터 출력됨
```

**테스트 매핑**: 기존 구현에서 level 파라미터 지원으로 확보

**정의된 완료**: ❓ (명시적 테스트 필요 - M1-2에서 확인)

---

### AC-016: 커스텀 로그 디렉토리
**요구사항**: log_dir 파라미터로 커스텀 로그 디렉토리 지정

**Given**
```
커스텀 디렉토리 경로 "/custom/logs"
```

**When**
```python
logger = setup_logger("app", log_dir="/custom/logs")
logger.info("Test")
```

**Then**
```
1. 로그 파일이 "/custom/logs/moai.log"에 생성됨
2. ".moai/logs/moai.log"가 아니라 지정된 경로 사용
3. 디렉토리가 없으면 자동 생성됨
```

**테스트 매핑**: `TestLoggerSetup.test_setup_logger_creates_log_file()` (tmp_path 파라미터 사용)

**정의된 완료**: ✅

---

## 2. 테스트 시나리오

### 테스트 그룹 1: 기본 로거 설정

#### TS-001: 로거 객체 생성
```gherkin
Feature: 로거 객체 생성
  Scenario: setup_logger가 Logger 객체 반환
    Given 애플리케이션이 초기화된 상태
    When setup_logger("test")를 호출
    Then logging.Logger 객체가 반환됨
    And logger.name == "test"
```

#### TS-002: 디렉토리 자동 생성
```gherkin
Feature: 로그 디렉토리 자동 생성
  Scenario: 미존재 디렉토리 자동 생성
    Given .moai/logs 디렉토리가 없음
    When setup_logger()를 호출하고 logger.info()를 실행
    Then .moai/logs 디렉토리가 생성됨
    And .moai/logs/moai.log 파일이 생성됨
```

---

### 테스트 그룹 2: 민감정보 마스킹

#### TS-003: API Key 마스킹
```gherkin
Feature: API Key 마스킹
  Scenario: sk-로 시작하는 API Key 마스킹
    Given 로거가 초기화됨
    When "API Key: sk-1234567890abcdef" 로그
    Then 파일에 "sk-1234567890abcdef" 없음
    And "***REDACTED***" 포함
```

#### TS-004: 이메일 마스킹
```gherkin
Feature: 이메일 마스킹
  Scenario: 표준 이메일 주소 마스킹
    Given 로거가 초기화됨
    When "Email: user@example.com" 로그
    Then 파일에 "user@example.com" 없음
    And "***REDACTED***" 포함
```

#### TS-005: 비밀번호 마스킹
```gherkin
Feature: 비밀번호 마스킹
  Scenario: password 키워드 뒤의 값 마스킹
    Given 로거가 초기화됨
    When "password: secret123" 로그
    Then 파일에 "secret123" 없음
    And "password: ***REDACTED***" 포함
```

#### TS-006: 다중 민감정보 마스킹
```gherkin
Feature: 다중 민감정보 마스킹
  Scenario: 한 메시지에 여러 민감정보 동시 마스킹
    Given 로거가 초기화됨
    When "API: sk-abc, Email: test@test.com, pwd: pass" 로그
    Then 모든 민감정보가 마스킹됨
    And "***REDACTED***" 3번 이상 포함
```

---

### 테스트 그룹 3: 환경별 로그 레벨

#### TS-007: Development 환경
```gherkin
Feature: Development 모드 로그 레벨
  Scenario: MOAI_ENV=development이면 DEBUG 레벨
    Given MOAI_ENV=development
    When setup_logger()를 호출
    Then logger.level == DEBUG
    And DEBUG 메시지 출력됨
```

#### TS-008: Production 환경
```gherkin
Feature: Production 모드 로그 레벨
  Scenario: MOAI_ENV=production이면 WARNING 레벨
    Given MOAI_ENV=production
    When setup_logger()를 호출
    Then logger.level == WARNING
    And INFO 메시지 미출력
```

#### TS-009: Test 환경
```gherkin
Feature: Test 모드 로그 레벨
  Scenario: MOAI_ENV=test이면 INFO 레벨
    Given MOAI_ENV=test
    When setup_logger()를 호출
    Then logger.level == INFO
    And DEBUG 미출력, INFO는 출력
```

#### TS-010: 기본값
```gherkin
Feature: 기본 로그 레벨
  Scenario: 환경변수 미설정이면 INFO 레벨
    Given MOAI_ENV 미설정
    When setup_logger()를 호출
    Then logger.level == INFO (기본값)
```

---

### 테스트 그룹 4: 핸들러 검증

#### TS-011: 콘솔 핸들러
```gherkin
Feature: 콘솔 핸들러 존재
  Scenario: StreamHandler가 추가됨
    Given 로거가 초기화됨
    When logger.handlers 확인
    Then StreamHandler 최소 1개 존재
    And handler.level이 설정됨
```

#### TS-012: 파일 핸들러
```gherkin
Feature: 파일 핸들러 존재
  Scenario: FileHandler가 추가됨
    Given 로거가 초기화됨
    When logger.handlers 확인
    Then FileHandler 정확히 1개 존재
    And 파일 경로 == .moai/logs/moai.log
```

#### TS-013: 로그 포맷
```gherkin
Feature: 로그 메시지 포맷
  Scenario: 정의된 포맷으로 출력
    Given 로거가 초기화됨
    When logger.info("Test")
    Then 로그 포맷 == "[YYYY-MM-DD HH:MM:SS] [LEVEL] [NAME] MESSAGE"
    And 모든 필드 포함됨
```

---

### 테스트 그룹 5: 엣지 케이스

#### TS-014: 로거 재초기화
```gherkin
Feature: 로거 재초기화
  Scenario: 동일 이름으로 재초기화 시 핸들러 제거
    Given 로거 logger1이 생성됨 (2개 핸들러)
    When setup_logger("logger1") 재호출
    Then 핸들러 수 == 2 (중복 없음)
```

#### TS-015: 매우 긴 메시지
```gherkin
Feature: 큰 메시지 처리
  Scenario: 매우 긴 로그 메시지 처리
    Given 로거가 초기화됨
    When 10,000자 메시지 로그
    Then 파일에 전체 메시지 기록됨
    And 성능 >= 시간 제한 (정의 필요)
```

#### TS-016: 유니코드 처리
```gherkin
Feature: 유니코드 문자 처리
  Scenario: 한글, 이모지 등 유니코드 로깅
    Given 로거가 초기화됨
    When "안녕하세요 👋 user@test.com" 로그
    Then 파일에 한글과 이모지 정상 저장
    And 이메일만 마스킹됨
```

#### TS-017: 특수문자 처리
```gherkin
Feature: 특수문자 처리
  Scenario: 따옴표, 백슬래시 등 특수문자
    Given 로거가 초기화됨
    When "Message: \"test\\path\"" 로그
    Then 파일에 정상 저장
    And 포맷 손상 없음
```

---

## 3. 품질 게이트 기준

### 필수 기준 (Must Have)

#### QG-001: 테스트 통과율
- **목표**: 100% (18/18 테스트 통과)
- **측정**: `pytest tests/unit/test_logger.py -v`
- **실패 기준**: 1개 이상 실패

#### QG-002: 코드 커버리지
- **목표**: ≥ 85%
- **측정**: `pytest --cov=src/moai_adk/utils/logger --cov-report=term-missing`
- **현재 예상**: 95% 이상
- **실패 기준**: 85% 미만

#### QG-003: 민감정보 마스킹 검증
- **목표**: 정의된 모든 패턴 검증
  - API Key: ✅
  - 이메일: ✅
  - 비밀번호: ✅
- **실패 기준**: 1개 패턴이라도 미검증

#### QG-004: 코드 포매팅
- **도구**: black, flake8, isort
- **목표**: 0 violations
- **실패 기준**: 1개 이상 위반

### 권장 기준 (Should Have)

#### QG-005: 타입 힌팅
- **목표**: 함수 서명 100% 타입 힌팅
- **도구**: mypy
- **실패 기준**: mypy 에러 3개 이상

#### QG-006: docstring 커버리지
- **목표**: 모든 public 함수/클래스에 docstring
- **도구**: pydocstyle
- **실패 기준**: 1개 이상 누락

---

## 4. 검증 방법 및 도구

### 자동 검증

#### 유닛 테스트
```bash
# 모든 테스트 실행
pytest tests/unit/test_logger.py -v

# 특정 테스트 클래스
pytest tests/unit/test_logger.py::TestSensitiveDataMasking -v

# 커버리지 리포트
pytest tests/unit/test_logger.py --cov=src/moai_adk/utils/logger --cov-report=html
```

#### 정적 분석
```bash
# 코드 포매팅
black src/moai_adk/utils/logger.py
isort src/moai_adk/utils/logger.py

# 린팅
flake8 src/moai_adk/utils/logger.py

# 타입 체킹
mypy src/moai_adk/utils/logger.py
```

### 수동 검증

#### 로그 파일 검증
```bash
# 로그 파일 생성 확인
ls -la .moai/logs/moai.log

# 로그 내용 확인
cat .moai/logs/moai.log

# 민감정보 노출 확인 (grep)
grep "sk-" .moai/logs/moai.log  # 결과 없어야 함
grep "@" .moai/logs/moai.log    # 이메일 검증
```

#### 환경별 테스트
```bash
# Development
MOAI_ENV=development python -c "from moai_adk.utils.logger import setup_logger; logger = setup_logger('test'); print(logger.level)"

# Production
MOAI_ENV=production python -c "from moai_adk.utils.logger import setup_logger; logger = setup_logger('test'); print(logger.level)"

# Test
MOAI_ENV=test python -c "from moai_adk.utils.logger import setup_logger; logger = setup_logger('test'); print(logger.level)"
```

---

## 5. 정의된 완료 (Definition of Done)

### 코드 변경
- [ ] 타입 힌팅 완료
- [ ] docstring 업데이트
- [ ] 코드 포매팅 (black, isort, flake8) 통과

### 테스트
- [ ] 18개 기존 테스트 모두 통과 (PASS)
- [ ] 테스트 커버리지 ≥ 85%
- [ ] 모든 AC 매핑 확인
- [ ] 추가 테스트 케이스 (필요시) 작성 및 통과

### 문서화
- [ ] spec.md 작성 및 검증
- [ ] plan.md 작성 완료
- [ ] acceptance.md 완성

### 검증
- [ ] 보안 요구사항 검증
- [ ] 성능 베이스라인 확인
- [ ] 호환성 검증 (Python 3.10+)
- [ ] Windows/Linux/macOS 테스트 (CI/CD)

### 배포 준비
- [ ] 변경 사항 문서화
- [ ] 마이그레이션 가이드 (필요시)
- [ ] 릴리스 노트 준비
- [ ] git commit 메시지 작성 준비

---

## 6. 테스트 매핑 요약

| AC ID | 설명 | 기존 테스트 | 상태 |
|-------|------|----------|------|
| AC-001 | 로거 생성 | test_setup_logger_creates_logger | ✅ |
| AC-002 | API Key 마스킹 | test_api_key_masking | ✅ |
| AC-003 | 이메일 마스킹 | test_email_masking | ✅ |
| AC-004 | 비밀번호 마스킹 | test_password_masking | ✅ |
| AC-005 | Dev 레벨 | test_development_mode_debug_level | ✅ |
| AC-006 | Prod 레벨 | test_production_mode_warning_level | ✅ |
| AC-007 | Test 레벨 | test_test_mode_info_level | ✅ |
| AC-008 | 기본값 | test_default_mode_info_level | ✅ |
| AC-009 | 디렉토리 생성 | test_setup_logger_creates_log_directory | ✅ |
| AC-010 | 콘솔 핸들러 | test_console_handler_exists | ✅ |
| AC-011 | 파일 출력 | test_file_handler_writes_to_file | ✅ |
| AC-012 | 핸들러 중복 제거 | (구현에서 handlers.clear()) | ✅ |
| AC-013 | 다중 마스킹 | test_multiple_sensitive_data_masking | ✅ |
| AC-014 | 일반 로그 보존 | test_filter_preserves_non_sensitive_data | ✅ |
| AC-015 | 커스텀 레벨 | (파라미터 지원 확인 필요) | ❓ |
| AC-016 | 커스텀 디렉토리 | (파라미터 지원 확인 필요) | ❓ |

---

## 7. 위험 요소 및 대응

### 위험 1: 민감정보 패턴 미탐지
**가능성**: 낮음 | **심각도**: 높음
**대응**: 정규식 패턴 매칭률 100% 검증 필요

### 위험 2: 성능 저하 (매우 큰 메시지)
**가능성**: 낮음 | **심각도**: 중간
**대응**: 벤치마크 테스트 추가 (Phase 2)

### 위험 3: Windows 파일 잠금
**가능성**: 중간 | **심각도**: 중간
**대응**: 예외 처리 강화, CI/CD Windows 테스트

---

**버전 1.0 완료**
