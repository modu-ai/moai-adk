---
id: SPEC-HOOK-001
version: "1.0.0"
status: "draft"
created: "2026-01-13"
updated: "2026-01-13"
author: "Alfred"
priority: "HIGH"
tags: [hook-system, quality-improvement, type-safety, file-operations, logging]
---

# SPEC-HOOK-001: Hook System 통합 품질 개선

## HISTORY

| 버전 | 날짜 | 변경 사항 | 작성자 |
|-------|------|-----------|--------|
| 1.0.0 | 2026-01-13 | 초기 SPEC 작성 | Alfred |

---

## 개요

### 목적

.claude/hooks/moai/ 디렉토리 내 모든 Python 파일의 품질을 개선하여 5가지 P1 이슈를 해결하고 안정적이고 유지보수 가능한 Hook System을 구축합니다.

### 범위

- 대상 디렉토리: `.claude/hooks/moai/`
- 대상 파일: 디렉토리 내 모든 Python 파일 (.py)
- 해결 이슈:
  - H1: 타입 어노테이션 통합 (typing 모듈 → builtin)
  - H2: 파일 크기 검증 (10MB 제한)
  - H3: 원자적 파일 연산 (경쟁 조건 방지)
  - H4: 구조화된 로깅 (일관된 포맷)
  - H5: 코드 중복 제거 (공통 패턴 추출)

### 배경

Hook System은 MoAI-ADK의 핵심 구성 요소로서 세션 시작/종료, 도구 사용 전후 등 다양한 이벤트에서 자동화된 작업을 수행합니다. 현재 5가지 품질 문제가 식별되었으며, 이를 해결하여 시스템 안정성과 유지보수성을 개선해야 합니다.

---

## 환경 및 가정

### 환경

- Python 3.13+ 실행 환경
- MoAI-ADK 템플릿 구조
- Unix-like 파일 시스템 (macOS, Linux)
- Claude Code Hook 시스템 통합

### 가정

1. 모든 Hook 파일은 Python 3.13+ 문법을 준수
2. 파일 시스템은 원자적 연산을 지원 (fsync, rename)
3. 로그는 stdout/stderr로 출력되며 Claude Code가 수집
4. Hook 실행 시간은 5초 이내여야 함 (timeout)
5. 모든 Hook은 독립적으로 실행 가능해야 함

---

## 요구사항 (EARS 형식)

### H1: 타입 어노테이션 통합

#### Ubiquitous Requirements (항상 활성)

- 시스템은 **항상** `str | None` 형식의 타입 어노테이션을 사용해야 한다
- 시스템은 **항상`typing.Optional[str]` 대신 builtin 타입을 사용해야 한다
- 시스템은 **항상** 함수 시그니처에 명확한 타입 어노테이션을 포함해야 한다

#### Event-Driven Requirements (이벤트 기반)

- **WHEN** 새로운 Hook 함수가 작성될 때 **THEN** 시스템은 모든 파라미터와 반환값에 타입 어노테이션을 포함해야 한다
- **WHEN** 기존 코드가 수정될 때 **THEN** 시스템은 타입 어노테이션을 `str | None` 형식으로 통합해야 한다

#### State-Driven Requirements (상태 기반)

- **IF** 함수가 문자열을 반환할 수 있는 경우 **THEN** 반환 타입은 `str | None`으로 명시해야 한다
- **IF** 파라미터가 nullable일 경우 **THEN** 기본값 `None`과 타입 `str | None`을 명시해야 한다

#### Unwanted Requirements (금지 사항)

- 시스템은 `typing` 모듈의 `Optional`, `List`, `Dict`를 사용하지 않아야 한다
- 시스템은 타입 어노테이션 없이 함수를 정의하지 않아야 한다

#### Optional Requirements (선택 사항)

- **가능하면** mypy strict mode를 통과하는 타입 어노테이션을 제공한다

---

### H2: 파일 크기 검증

#### Ubiquitous Requirements (항상 활성)

- 시스템은 **항상** 파일을 읽기 전에 크기를 확인해야 한다
- 시스템은 **항상** 10MB 제한을 초과하는 파일 읽기를 거부해야 한다

#### Event-Driven Requirements (이벤트 기반)

- **WHEN** 파일 읽기 작업이 시도될 때 **THEN** 시스템은 먼저 파일 크기를 확인해야 한다
- **WHEN** 파일 크기가 10MB를 초과할 때 **THEN** 시스템은 `ValueError`를 발생시키고 로깅해야 한다

#### State-Driven Requirements (상태 기반)

- **IF** 파일 크기가 10MB 이하일 경우 **THEN** 시스템은 정상적으로 파일을 읽어야 한다
- **IF** 파일 크기가 확인 불가능할 경우 **THEN** 시스템은 안전을 위해 읽기를 거부해야 한다

#### Unwanted Requirements (금지 사항)

- 시스템은 크기 확인 없이 대용량 파일을 읽지 않아야 한다
- 시스템은 크기 초과 시 무한 루프나 크래시를 일으키지 않아야 한다

#### Optional Requirements (선택 사항)

- **가능하면** 설정 가능한 크기 제한을 지원한다 (기본값 10MB)

---

### H3: 원자적 파일 연산

#### Ubiquitous Requirements (항상 활성)

- 시스템은 **항상** 상태 파일과 캐시 파일에 원자적 연산을 사용해야 한다
- 시스템은 **항상** 파일 쓰기에 임시 파일 + rename 패턴을 사용해야 한다

#### Event-Driven Requirements (이벤트 기반)

- **WHEN** 상태 파일이 업데이트될 때 **THEN** 시스템은 임시 파일에 쓰고 원자적으로 rename해야 한다
- **WHEN** 여러 프로세스가 동시에 파일에 접근할 때 **THEN** 시스템은 경쟁 조건을 방지해야 한다

#### State-Driven Requirements (상태 기반)

- **IF** 파일 쓰기가 중단될 경우 **THEN** 시스템은 원본 파일을 손상하지 않아야 한다
- **IF** rename이 실패할 경우 **THEN** 시스템은 임시 파일을 정리하고 오류를 보고해야 한다

#### Unwanted Requirements (금지 사항)

- 시스템은 직접 원본 파일을 덮어쓰지 않아야 한다
- 시스템은 fsync 없이 중요한 데이터를 쓰지 않아야 한다

#### Optional Requirements (선택 사항)

- **가능하면** file locking (fcntl/flock)을 사용하여 동시 접근을 방지한다

---

### H4: 구조화된 로깅

#### Ubiquitous Requirements (항상 활성)

- 시스템은 **항상** 구조화된 JSON 형식으로 로그를 출력해야 한다
- 시스템은 **항상** 로그 레벨 (INFO, WARNING, ERROR)을 명시해야 한다

#### Event-Driven Requirements (이벤트 기반)

- **WHEN** Hook이 실행될 때 **THEN** 시스템은 시작 로그를 출력해야 한다
- **WHEN** 오류가 발생할 때 **THEN** 시스템은 ERROR 레벨 로그와 스택 트레이스를 출력해야 한다

#### State-Driven Requirements (상태 기반)

- **IF** 실행이 성공할 경우 **THEN** 시스템은 INFO 레벨로 성공 메시지를 출력해야 한다
- **IF** 실행이 실패할 경우 **THEN** 시스템은 ERROR 레벨로 실패 원인을 출력해야 한다

#### Unwanted Requirements (금지 사항)

- 시스템은 print() 문으로 직접 로그를 출력하지 않아야 한다
- 시스템은 구조화되지 않은 자유 형식 텍스트를 로그로 사용하지 않아야 한다

#### Optional Requirements (선택 사항)

- **가능하면** 로그에 timestamp, hook_name, execution_id 필드를 포함한다

---

### H5: 코드 중복 제거

#### Ubiquitous Requirements (항상 활성)

- 시스템은 **항상** 반복되는 패턴을 헬퍼 함수로 추출해야 한다
- 시스템은 **항상** 공통 유틸리티를 중앙 모듈로 관리해야 한다

#### Event-Driven Requirements (이벤트 기반)

- **WHEN** 중복 코드가 발견될 때 **THEN** 시스템은 이를 재사용 가능한 함수로 리팩토링해야 한다
- **WHEN** 새로운 Hook이 작성될 때 **THEN** 시스템은 기존 헬퍼 함수를 재사용해야 한다

#### State-Driven Requirements (상태 기반)

- **IF** 3개 이상의 Hook에서 유사한 패턴이 사용될 경우 **THEN** 시스템은 이를 공통 모듈로 추출해야 한다
- **IF** 헬퍼 함수가 20줄을 초과할 경우 **THEN** 시스템은 이를 더 작은 함수로 분할해야 한다

#### Unwanted Requirements (금지 사항)

- 시스템은 copy-paste 코드 중복을 허용하지 않아야 한다
- 시스템은 명확한 목적 없는 범용 헬퍼 함수를 만들지 않아야 한다

#### Optional Requirements (선택 사항)

- **가능하면** Hook 기반 클래스를 만들어 상속을 통한 코드 재사용을 지원한다

---

## 인수 기준 요약

### 공통 성공 기준

1. 모든 Python 파일이 mypy strict mode를 통과
2. 모든 파일 읽기 작업이 10MB 제한을 준수
3. 모든 상태 파일 연산이 원자적임을 보장
4. 모든 로그가 구조화된 JSON 형식
5. 코드 중복이 30% 이상 감소

### 품질 게이트

- [ ] TRUST-5 프레임워크 준수 (Test-first, Readable, Unified, Secured, Trackable)
- [ ] 85% 이상 테스트 커버리지
- [ ] ruff linter 통과 (zero warnings)
- [ ] OWASP 보안 기준 준수

---

## 추적성

### 관련 문서

- [CLAUDE.md](../../../../CLAUDE.md) - 실행 지침
- [moai-foundation-core](../../../../.claude/skills/moai-foundation-core) - 품질 기준
- [moai-lang-python](../../../../.claude/skills/moai-lang-python) - Python 3.13 패턴

### 관련 이슈

- P1 이슈 5개 (H1-H5)

### 다음 단계

```bash
# TDD 실행
/moai:2-run SPEC-HOOK-001

# 문서 동기화
/moai:3-sync SPEC-HOOK-001
```

---

## 참고 자료

- Python 3.13 Type Hints: [PEP 484](https://peps.python.org/pep-0484/)
- Atomic File Operations: [atomicwrites](https://github.com/untitaker/python-atomicwrites)
- Structured Logging: [structlog](https://www.structlog.org/)
