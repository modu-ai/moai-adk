---
id: SPEC-HOOK-001
version: "1.0.0"
status: "draft"
created: "2026-01-13"
updated: "2026-01-13"
author: "Alfred"
priority: "HIGH"
tags: [hook-system, quality-improvement, type-safety, file-operations, logging]
spec_id: SPEC-HOOK-001
---

# 구현 계획: SPEC-HOOK-001 Hook System 통합 품질 개선

## 구현 개요

이 계획은 `.claude/hooks/moai/` 디렉토리 내 모든 Python 파일의 품질을 개선하여 5가지 P1 이슈를 해결하는 접근 방식을 설명합니다.

### 핵심 목표

1. **타입 안전성 강화**: 모든 함수에 명확한 타입 어노테이션 추가
2. **안정성 보장**: 파일 크기 검증 및 원자적 연산으로 데이터 손실 방지
3. **가시성 개선**: 구조화된 로깅으로 디버깅 및 모니터링 향상
4. **유지보수성 향상**: 코드 중복 제거로 재사용성 증가

---

## 우선순위별 마일스톤

### 1차 목표 (Priority High)

**범위**: H1 타입 어노테이션 통합, H4 구조화된 로깅

**목적**:
- 타입 안전성 기반 구축
- 실행 가시성 확보
- 정적 분석 도구 통합

**작업**:
1. 모든 Hook 파일의 타입 어노테이션을 `str | None` 형식으로 통합
2. `typing` 모듈 의존성 제거 (Optional, List, Dict → builtin)
3. 구조화된 로깅 시스템 구현 (JSON 형식)
4. 로그 레벨별 출력 표준 정립

**성공 기준**:
- mypy strict mode 통과
- 모든 로그가 JSON 형식
- 0개의 타입 관련 lint 경고

---

### 2차 목표 (Priority High)

**범위**: H2 파일 크기 검증, H3 원자적 파일 연산

**목적**:
- 대용량 파일 처리로 인한 시스템 불안정 방지
- 경쟁 조건으로 인한 데이터 손상 방지

**작업**:
1. 파일 읽기 전 크기 확인 함수 구현 (10MB 제한)
2. 크기 초과 시 명확한 에러 메시지와 로깅
3. 상태 파일 및 캐시 파일에 원자적 연산 적용
4. 임시 파일 + fsync + rename 패턴 구현

**성공 기준**:
- 모든 파일 읽기가 크기 제한 준수
- 동시 실행 시 데이터 손상 0건
- 원자적 연산 테스트 커버리지 90%+

---

### 3차 목표 (Priority Medium)

**범위**: H5 코드 중복 제거

**목적**:
- 재사용 가능한 헬퍼 함수 라이브러리 구축
- 유지보수 비용 절감

**작업**:
1. 중복 패턴 식별 (file operations, logging, validation)
2. 공통 헬퍼 함수 모듈 생성 (`hook_utils.py`)
3. 기존 Hook 파일 리팩토링하여 헬퍼 함수 사용
4. 코드 중복 30% 이상 감소 달성

**성공 기준**:
- 중복 코드 라인 30% 이상 감소
- 헬퍼 함수 재사용률 80%+
- 리팩토링 후 기능 동일성 보장

---

### 4차 목표 (Priority Medium)

**범위**: 테스트 커버리지 및 문서화

**목적**:
- TRUST-5 프레임워크 Test-first pillar 준수
- 향후 유지보수를 위한 문서화

**작업**:
1. 각 기능별 단위 테스트 작성 (pytest)
2. 통합 테스트 작성 (파일 연산, 로깅)
3. Edge case 테스트 (대용량 파일, 동시 접근)
4. API 문서화 (docstring, type hints 활용)

**성공 기준**:
- 전체 테스트 커버리지 85%+
- 모든 공개 함수에 docstring 존재
- CI/CD 파이프라인에서 테스트 자동 실행

---

### 5차 목표 (Priority Low)

**범위**: 성능 최적화 및 고급 기능

**목적**:
- Hook 실행 시간 최적화
- 선택적 고급 기능 구현

**작업**:
1. Hook 실행 시간 프로파일링 및 병목 지점 최적화
2. 설정 가능한 크기 제한 기능 (기본값 10MB)
3. File locking (fcntl/flock)을 통한 동시 접근 제어
4. Hook 기반 클래스 구현을 통한 상속 패턴 지원

**성공 기준**:
- 평균 Hook 실행 시간 2초 이내
- 설정 파일을 통한 크기 제한 변경 가능
- 선택적 file locking 활성화 가능

---

## 기술 접근 방식

### 아키텍처 설계

#### 모듈 구조

```
.claude/hooks/moai/
├── __init__.py
├── hook_utils.py          # 공통 헬퍼 함수 (H5)
├── logging_config.py      # 구조화된 로깅 설정 (H4)
├── file_operations.py     # 안전한 파일 연산 (H2, H3)
│
├── session_start__*.py    # 세션 시작 Hook들
├── session_end__*.py      # 세션 종료 Hook들
├── pre_tool_use__*.py     # 도구 사용 전 Hook들
└── post_tool_use__*.py    # 도구 사용 후 Hook들
```

#### 핵심 모듈 설계

**1. hook_utils.py** (H5: 코드 중복 제거)

```python
"""공통 Hook 유틸리티 함수"""

from pathlib import Path
from typing import Any

def validate_file_size(file_path: Path, max_size_mb: int = 10) -> None:
    """파일 크기가 제한 내인지 확인 (H2)"""
    # Implementation

def atomic_write(file_path: Path, content: str) -> None:
    """원자적 파일 쓰기 (H3)"""
    # Implementation

def setup_structured_logging(hook_name: str) -> None:
    """구조화된 로거 초기화 (H4)"""
    # Implementation
```

**2. logging_config.py** (H4: 구조화된 로깅)

```python
"""구조화된 로깅 설정"""

import json
import sys
from typing import Any

def log_message(level: str, hook_name: str, message: str, **kwargs: Any) -> None:
    """구조화된 JSON 로그 출력"""
    log_entry = {
        "level": level,
        "hook_name": hook_name,
        "message": message,
        **kwargs
    }
    print(json.dumps(log_entry), file=sys.stderr)
```

**3. file_operations.py** (H2, H3: 파일 연산)

```python
"""안전한 파일 연산"""

import os
import tempfile
from pathlib import Path

def safe_read_file(file_path: Path, max_size_mb: int = 10) -> str:
    """크기 검증이 포함된 안전한 파일 읽기"""
    # Implementation with size validation

def atomic_write_state(file_path: Path, data: str) -> None:
    """상태 파일 원자적 쓰기 (temp file + fsync + rename)"""
    # Implementation with atomic operations
```

---

## 기술 스택 명세

### 핵심 라이브러리

| 라이브러리 | 버전 | 용도 |
|-----------|------|------|
| Python | 3.13+ | 실행 환경 |
| mypy | 1.11+ | 정적 타입 검사 |
| pytest | 8.0+ | 테스트 프레임워크 |
| pytest-cov | 4.1+ | 커버리지 측정 |
| ruff | 0.1+ | Linter 및 Formatter |

### 타입 검사 설정

```toml
[tool.mypy]
python_version = "3.13"
strict = true
warn_return_any = true
warn_unused_ignores = true
disallow_untyped_defs = true
```

### 테스트 설정

```toml
[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py"]
python_functions = ["test_*"]
addopts = "--cov=.claude/hooks --cov-report=term-missing --cov-report=html"
```

---

## 리스크 분석 및 완화 전략

### 리스크 1: 기존 Hook 호환성

**위험도**: Medium

**설명**: 타입 어노테이션과 로깅 시스템 변경이 기존 Hook 동작에 영향을 줄 수 있음

**완화 전략**:
- 점진적 마이그레이션 (Hook마다 개별 적용)
- 회귀 테스트로 기존 동작 보장
- Backward compatibility layer 제공

---

### 리스크 2: 성능 저하

**위험도**: Low

**설명**: 구조화된 로깅과 파일 크기 검증이 Hook 실행 시간을 증가시킬 수 있음

**완화 전략**:
- Lazy evaluation 로깅
- 파일 크기 확인은 OS call (빠름)
- 프로파일링으로 병목 지점 식별 및 최적화

---

### 리스크 3: 원자적 연산 실패

**위험도**: Medium

**설명**: 일부 파일 시스템에서 rename이 원자적이지 않을 수 있음

**완화 전략**:
- 테스트로 대상 파일 시스템 동작 확인
- 실패 시 롤백 메커니즘 구현
- 명확한 에러 메시지와 복구 가이드 제공

---

### 리스크 4: 과도한 추상화

**위험도**: Low

**설명**: 헬퍼 함수 추출이 코드 가독성을 해치거나 over-engineering을 유발할 수 있음

**완화 전략**:
- YAGNI 원칙 준수 (필요한 것만 구현)
- 간단한 use case에 대한 복잡한 추상화 피함
- Code review로 적절한 추상화 수준 확인

---

## 작업 분해 및 의존성

### 작업 순서

```
Phase 1: 기반 구축
├── Task 1.1: hook_utils.py 스캐폴딩
├── Task 1.2: logging_config.py 구현
└── Task 1.3: file_operations.py 구현

Phase 2: H1 타입 어노테이션
├── Task 2.1: 기존 Hook 파일 타입 분석
├── Task 2.2: 타입 어노테이션 추가 (각 Hook 파일)
└── Task 2.3: mypy strict mode 통과 확인

Phase 3: H4 구조화된 로깅
├── Task 3.1: 기존 print 문 식별
├── Task 3.2: 구조화된 로깅으로 교체
└── Task 3.3: 로그 출력 형식 검증

Phase 4: H2 파일 크기 검증
├── Task 4.1: 파일 읽기 지점 식별
├── Task 4.2: 크기 검증 로직 추가
└── Task 4.3: 크기 초과 에러 처리

Phase 5: H3 원자적 파일 연산
├── Task 5.1: 상태 파일 쓰기 지점 식별
├── Task 5.2: 원자적 연산으로 교체
└── Task 5.3: 동시 실행 테스트

Phase 6: H5 코드 중복 제거
├── Task 6.1: 중복 패턴 분석
├── Task 6.2: 헬퍼 함수 추출
└── Task 6.3: 기존 코드 리팩토링

Phase 7: 테스트 및 문서화
├── Task 7.1: 단위 테스트 작성
├── Task 7.2: 통합 테스트 작성
├── Task 7.3: 문서화 (docstring)
└── Task 7.4: acceptance 기준 검증
```

### 의존성 그래프

```
hook_utils.py (기반)
    ↓
logging_config.py, file_operations.py
    ↓
H1 (타입), H4 (로깅), H2 (크기), H3 (원자적 연산)
    ↓
H5 (코드 중복 제거)
    ↓
테스트 및 문서화
```

---

## 성공 지표 및 측정 방법

### 정량적 지표

| 지표 | 기준값 | 측정 방법 |
|-----|--------|----------|
| 타입 커버리지 | 100% | mypy report |
| 테스트 커버리지 | 85%+ | pytest-cov |
| 코드 중복 감소율 | 30%+ | pylint similarity |
| Linter 경고 수 | 0 | ruff check |
| 평균 Hook 실행 시간 | <2초 | pytest-benchmark |

### 정성적 지표

- 모든 Hook이 mypy strict mode 통과
- 모든 로그가 구조화된 JSON 형식
- 코드 가독성 및 유지보수성 향상
- 신규 Hook 작성 시 헬퍼 함수 재사용 용이

---

## 참고 자료

### 내부 참고

- [moai-foundation-core](../../../../.claude/skills/moai-foundation-core) - TRUST-5 프레임워크
- [moai-lang-python](../../../../.claude/skills/moai-lang-python) - Python 3.13 패턴
- [spec.md](./spec.md) - 요구사항 상세

### 외부 참고

- [Python Type Hints (PEP 484)](https://peps.python.org/pep-0484/)
- [Atomic File Operations](https://github.com/untitaker/python-atomicwrites)
- [Structured Logging](https://www.structlog.org/)
- [Test-Driven Development](https://pytest.org/)

---

## 다음 단계

```bash
# TDD 실행 (이 SPEC을 기반으로 구현 시작)
/moai:2-run SPEC-HOOK-001

# 구현 완료 후 문서 동기화
/moai:3-sync SPEC-HOOK-001
```
