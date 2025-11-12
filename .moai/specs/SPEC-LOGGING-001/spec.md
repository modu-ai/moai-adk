---
id: LOGGING-001
domain: LOGGING
title: "로깅 유틸리티 - 기본 시스템"
version: "1.0.0"
status: "in-progress"
created: "2025-11-13"
author: "GoosLab"
---

# @SPEC:LOGGING-001 | @EXPERT:BACKEND | @EXPERT:DEVOPS

## SPEC 개요

이 SPEC은 MoAI-ADK의 로깅 유틸리티 기본 시스템을 정의합니다. 애플리케이션 초기화 시점에 민감정보 마스킹, 환경별 로그 레벨 관리, 콘솔 및 파일 이중 출력을 제공하는 로깅 인프라를 구축합니다.

## 환경 (Environment)

- **실행 시점**: 애플리케이션 초기화 단계
- **환경변수**: `MOAI_ENV` (development/test/production)
- **로그 저장 경로**: `.moai/logs/`
- **타겟 런타임**: Python 3.10+

## 가정 (Assumptions)

1. Python 3.10 이상의 표준 logging 모듈 사용
2. 파일시스템 접근 및 디렉토리 생성 권한 확보
3. UTF-8 인코딩 지원 환경
4. 싱글 프로세스 애플리케이션 (멀티프로세싱은 별도 고려)

## 요구사항 (Requirements)

### 함수형 요구사항 (Functional Requirements)

#### REQ-LOG-001: 로거 초기화
**WHEN** `setup_logger(name)` 함수 호출
**THEN** 지정된 이름의 Logger 객체를 반환하고, 콘솔 및 파일 핸들러가 자동 구성됨

#### REQ-LOG-002: 환경별 동적 로그 레벨
**WHEN** `setup_logger()` 호출 시 `MOAI_ENV` 환경변수 확인
**THEN** 환경별로 로그 레벨 설정:
- `MOAI_ENV=development` → DEBUG
- `MOAI_ENV=test` → INFO
- `MOAI_ENV=production` → WARNING
- 미설정 시 → INFO (기본값)

#### REQ-LOG-003: 민감정보 자동 마스킹
**WHEN** 로그 메시지에 민감정보 패턴 포함
**THEN** 다음 패턴들을 자동 감지하여 "***REDACTED***"로 변환:
- API Key: `sk-`로 시작하는 32자 이상의 영숫자 문자열
- 이메일: 표준 이메일 주소 형식 (`user@domain.com`)
- 비밀번호: `password:`, `passwd:`, `pwd:` 뒤의 값

#### REQ-LOG-004: 로그 파일 자동 생성
**WHEN** 로그 디렉토리가 존재하지 않는 경우
**THEN** 로그 디렉토리 자동 생성 및 `moai.log` 파일 생성

#### REQ-LOG-005: 구조화된 로그 포맷
**WHEN** 로그 메시지 기록
**THEN** 다음 형식으로 출력:
```
[YYYY-MM-DD HH:MM:SS] [LEVEL] [NAME] MESSAGE
```

### 비함수형 요구사항 (Non-Functional Requirements)

#### REQ-PERF-001: 로깅 성능
- 단일 로그 메시지 처리 시간: < 10ms
- 민감정보 마스킹 오버헤드: < 2ms

#### REQ-SECURITY-001: 민감정보 보호
- 모든 민감정보 패턴이 감지 및 마스킹됨
- 로그 파일 접근 권한: 현재 사용자만 읽기 가능

#### REQ-RELIABILITY-001: 에러 처리
- 로그 디렉토리 생성 실패 시: 예외 발생하지 않고 로깅 계속
- 파일 쓰기 실패 시: 콘솔에만 출력 (graceful degradation)

## 명세 (Specifications)

### 1. 로거 초기화 명세

#### setup_logger(name, log_dir=None, level=None) 함수
```python
def setup_logger(
    name: str,
    log_dir: str | None = None,
    level: int | None = None,
) -> logging.Logger:
    """
    로거 인스턴스를 생성하고 설정합니다.

    Args:
        name: 로거 이름 (모듈명 또는 애플리케이션명)
        log_dir: 로그 저장 디렉토리 (기본값: .moai/logs)
        level: 로깅 레벨 (logging.DEBUG/INFO/WARNING/ERROR/CRITICAL)
               미지정 시 MOAI_ENV 환경변수로부터 결정

    Returns:
        설정된 Logger 객체
    """
```

**동작 흐름**:
1. 로그 레벨 결정: 명시 여부 확인 → 환경변수 확인 → 기본값 적용
2. Logger 객체 생성 또는 재사용
3. 기존 핸들러 제거 (중복 방지)
4. 로그 디렉토리 생성 (필요시)
5. Formatter 정의
6. StreamHandler (콘솔) 추가
7. FileHandler (파일) 추가
8. SensitiveDataFilter 적용

### 2. 민감정보 필터 명세

#### SensitiveDataFilter 클래스
```python
class SensitiveDataFilter(logging.Filter):
    """로그 메시지에서 민감정보를 감지하고 마스킹합니다."""

    PATTERNS = [
        (r"sk-[a-zA-Z0-9]+", "***REDACTED***"),  # API Key
        (r"\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b", "***REDACTED***"),  # Email
        (r"(?i)(password|passwd|pwd)[\s:=]+\S+", r"\1: ***REDACTED***"),  # Password
    ]

    def filter(self, record: logging.LogRecord) -> bool:
        """로그 레코드에서 민감정보를 마스킹합니다."""
```

**마스킹 규칙**:
- API Key: `sk-` 로 시작하는 패턴 → `***REDACTED***`
- 이메일: 표준 이메일 형식 → `***REDACTED***`
- 비밀번호: `password:`, `passwd:`, `pwd:` 키워드 → `password: ***REDACTED***`

### 3. 핸들러 명세

#### StreamHandler (콘솔 출력)
- **대상**: stdout (표준 출력)
- **포맷**: `[YYYY-MM-DD HH:MM:SS] [LEVEL] [NAME] MESSAGE`
- **인코딩**: UTF-8
- **필터**: SensitiveDataFilter 적용

#### FileHandler (파일 기록)
- **경로**: `{log_dir}/moai.log`
- **기본 log_dir**: `.moai/logs`
- **모드**: append (w가 아닌 a)
- **인코딩**: UTF-8
- **포맷**: 콘솔과 동일
- **필터**: SensitiveDataFilter 적용

## 수용 기준 (Acceptance Criteria)

### AC-001: 로거 생성
**Given** 로거가 미생성된 상태
**When** `setup_logger("myapp")` 호출
**Then** Logger 객체가 반환되고 name 속성이 "myapp"

### AC-002: 민감정보 마스킹 - API Key
**Given** 로거가 초기화된 상태
**When** API Key "sk-1234567890abcdef" 로그 기록
**Then** 로그 파일에 원본이 아닌 "***REDACTED***" 저장

### AC-003: 민감정보 마스킹 - 이메일
**Given** 로거가 초기화된 상태
**When** 이메일 "user@example.com" 로그 기록
**Then** 로그 파일에 "***REDACTED***" 저장

### AC-004: 민감정보 마스킹 - 비밀번호
**Given** 로거가 초기화된 상태
**When** "password: secret123" 로그 기록
**Then** 로그 파일에 "password: ***REDACTED***" 저장

### AC-005: 환경별 로그 레벨 - development
**Given** `MOAI_ENV=development`
**When** `setup_logger()` 호출
**Then** logger.level == logging.DEBUG

### AC-006: 환경별 로그 레벨 - production
**Given** `MOAI_ENV=production`
**When** `setup_logger()` 호출
**Then** logger.level == logging.WARNING

### AC-007: 로그 파일 자동 생성
**Given** `.moai/logs` 디렉토리가 없는 상태
**When** 로거 초기화 후 로그 메시지 기록
**Then** `.moai/logs/moai.log` 파일이 자동 생성됨

### AC-008: 콘솔 출력 확인
**Given** 로거가 초기화된 상태
**When** `logger.info("Test")` 호출
**Then** 콘솔에 `[날짜시간] [INFO] [로거이름] Test` 형식으로 출력

### AC-009: 파일 출력 확인
**Given** 로거가 초기화된 상태
**When** `logger.info("File test")` 호출
**Then** `.moai/logs/moai.log` 파일에 메시지 저장

### AC-010: 기존 핸들러 중복 제거
**Given** 로거가 이미 생성된 상태 (핸들러 보유)
**When** 동일 이름으로 `setup_logger()` 재호출
**Then** 기존 핸들러가 제거되고 새 핸들러만 남음

### AC-011: 여러 민감정보 동시 마스킹
**Given** 로거가 초기화된 상태
**When** "API: sk-abc, Email: test@test.com, Password: pass123" 로그
**Then** 모든 민감정보가 마스킹됨

### AC-012: 일반 로그 보존
**Given** 로거가 초기화된 상태
**When** 민감정보 없는 메시지 "Normal log" 기록
**Then** 로그 파일에 "Normal log" 그대로 저장됨

### AC-013: 커스텀 로그 레벨 지정
**Given** `MOAI_ENV=production` (WARNING)
**When** `setup_logger("app", level=logging.DEBUG)` 호출
**Then** logger.level == logging.DEBUG (환경변수 무시)

### AC-014: 커스텀 로그 디렉토리
**Given** 커스텀 디렉토리 경로 지정
**When** `setup_logger("app", log_dir="/custom/path")` 호출
**Then** `/custom/path/moai.log`에 로그 저장

## 트레이서빌리티 체인

```
@SPEC:LOGGING-001
  ↓ (구현)
@CODE:LOGGING-001 (src/moai_adk/utils/logger.py)
  ↓ (검증)
@TEST:LOGGING-001 (tests/unit/test_logger.py)
```

## 서브컴포넌트

### 도메인 관리 (DOMAIN)
- **@CODE:LOGGING-001:DOMAIN** - 민감정보 패턴 정의 및 관리
- **@CODE:LOGGING-001:DOMAIN** - 환경별 로그 레벨 결정 로직

### 인프라 (INFRA)
- **@CODE:LOGGING-001:INFRA** - Logger 객체 생성 및 설정
- **@CODE:LOGGING-001:INFRA** - 로그 디렉토리 생성 및 검증
- **@CODE:LOGGING-001:INFRA** - 로그 포맷 정의
- **@CODE:LOGGING-001:INFRA** - StreamHandler 구성 (콘솔)
- **@CODE:LOGGING-001:INFRA** - FileHandler 구성 (파일)

## 관련 기술 및 라이브러리

| 항목 | 버전 | 용도 |
|------|------|------|
| Python | 3.10+ | 기본 runtime |
| logging | Built-in | 로깅 모듈 |
| pathlib | Built-in | 디렉토리 경로 관리 |
| re | Built-in | 민감정보 패턴 매칭 |

**외부 의존성**: 없음 (순수 표준 라이브러리 사용)

## 제약사항 및 제한

### 지원하지 않는 기능
- ❌ 로그 로테이션 (향후 SPEC-LOGGING-002에서 제공)
- ❌ JSON 구조화된 로깅 (향후 SPEC-LOGGING-002에서 제공)
- ❌ 분산 로깅 (syslog, 원격 로깅 등)
- ❌ 멀티프로세싱 동시성 관리 (single-process 전제)

### 알려진 제한
- 매우 큰 로그 파일은 메모리 영향 가능
- 파일 기반 로깅이므로 대량 concurrent logging 시 성능 저하 가능

## 참고사항

- 기존 구현: `src/moai_adk/utils/logger.py`
- 테스트 파일: `tests/unit/test_logger.py` (18개 테스트 케이스)
- 이 SPEC은 현재 구현을 정규화하고 `@CODE:LOGGING-001` 고아 TAG를 해결합니다.
