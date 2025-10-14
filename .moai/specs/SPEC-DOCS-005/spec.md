---
id: DOCS-005
version: 0.0.1
status: draft
created: 2025-10-14
updated: 2025-10-14
author: @Goos
priority: high
category: docs
labels:
  - documentation
  - python-3.13
  - consistency
  - v0.3.0
depends_on:
  - DOCS-004
scope:
  packages:
    - docs/
  files:
    - docs/**/*.md
---

# @SPEC:DOCS-005: 온라인 문서 v0.3.0 정합성 확보

## HISTORY
### v0.0.1 (2025-10-14)
- **INITIAL**: docs/ 온라인 문서를 Python v0.3.0에 맞게 업데이트
- **AUTHOR**: @Goos
- **SCOPE**: CLI 명령어 "🚧 Coming in v0.4.0" 배지, Alfred 커맨드 강조, Python 예제 코드 추가

---

## 개요

docs/ 디렉토리의 온라인 문서 17개 파일을 Python v0.3.0 기준으로 업데이트하여 README.md와 일관성을 확보합니다.

**핵심 문제**:
- 온라인 문서에 TypeScript v0.2.x 내용 포함
- CLI 명령어 설명 (실제로는 모두 미구현)
- `/alfred:0-project` vs `/alfred:8-project` 혼용
- Python 예제 코드 부족

**목표**:
- Python v0.3.0 실제 상태 정확 반영
- CLI 미구현 기능 "🚧 Coming in v0.4.0" 배지 추가
- Alfred 커맨드 명명 통일 (/alfred:0-project)
- Python 예제 코드 추가

---

## Ubiquitous Requirements (기본 요구사항)

시스템은 다음 기능을 제공해야 한다:
1. docs/ 디렉토리 17개 파일을 Python v0.3.0에 맞게 업데이트
2. CLI 미구현 기능 "🚧 Coming in v0.4.0" 배지 추가
3. `/alfred:0-project` 명명 통일
4. Python 예제 코드 추가
5. TypeScript 관련 내용 제거

---

## Event-driven Requirements (이벤트 기반)

- WHEN TypeScript 예제 코드를 발견하면, 시스템은 Python 예제 코드로 교체해야 한다
- WHEN CLI 명령어 설명을 발견하면, 시스템은 "🚧 Coming in v0.4.0" 배지를 추가해야 한다
- WHEN `/alfred:8-project`를 발견하면, 시스템은 `/alfred:0-project`로 변경해야 한다

---

## State-driven Requirements (상태 기반)

- WHILE 온라인 문서를 업데이트하는 동안, 시스템은 기존 섹션 구조를 최대한 유지해야 한다
- WHILE Python 예제 코드를 추가하는 동안, 시스템은 Python 3.13+ 문법을 따라야 한다

---

## Constraints (제약사항)

- 각 문서 파일은 1000 줄을 초과하지 않아야 한다
- IF 기존 섹션을 제거하면, 시스템은 제거 이유를 SPEC 문서에 기록해야 한다
- 모든 코드 예제는 Python 3.13+ 문법을 따라야 한다

---

## 수정 대상 파일 (17개)

### 1. docs/index.md
**수정 사항**:
- CLI 명령어 "🚧 Coming in v0.4.0" 배지 추가
- Python 3.13+ 설치 방법 강조
- Alfred 커맨드 중심 워크플로우 강조

### 2. docs/getting-started/installation.md
**수정 사항**:
- TypeScript 설치 방법 제거 (Bun, npm)
- Python 설치 방법 추가 (uv, pip)
- CLI 명령어 "🚧 Coming in v0.4.0" 배지 추가

### 3. docs/getting-started/quick-start.md
**수정 사항**:
- Python 기준 Quick Start 재작성
- Claude Code에서 Alfred 사용법 강조
- `/alfred:0-project` 명명 통일

### 4. docs/getting-started/first-project.md
**수정 사항**:
- Python 프로젝트 예제 추가
- Alfred 커맨드 사용법 강조
- TypeScript 예제 제거

### 5. docs/guides/alfred-superagent.md
**수정 사항**:
- `/alfred:0-project` 명명 통일
- Python 예제 코드 추가
- Alfred 커맨드 설명 강화

### 6. docs/guides/spec-first-tdd.md
**수정 사항**:
- Python TDD 예제 추가
- SPEC 메타데이터 예제 (Python 주석)
- pytest 사용법 추가

### 7. docs/guides/tag-system.md
**수정 사항**:
- Python TAG 사용 예시 추가
- `@CODE:ID` Python 주석 스타일
- pytest 테스트 예제

### 8. docs/guides/trust-principles.md
**수정 사항**:
- Python TRUST 원칙 구현 예제
- pytest, mypy, ruff 사용법
- Python 보안 베스트 프랙티스

### 9. docs/guides/workflow.md
**수정 사항**:
- Python 워크플로우 예제
- Alfred 커맨드 중심 설명
- `/alfred:0-project` 명명 통일

### 10-12. docs/agents/*.md (3개)
**수정 사항**:
- Python 예제 코드 추가
- Alfred 커맨드 설명 강화
- CLI 명령어 "🚧" 배지 추가

### 13-15. docs/api/*.md (3개)
**수정 사항**:
- Python API 문서 (v0.4.0 Coming Soon)
- TypeScript API 제거
- Alfred 커맨드 중심 설명

### 16-17. docs/specs/*.md (2개)
**수정 사항**:
- SPEC 메타데이터 예제 (Python 주석)
- Python TAG 사용 예시
- `/alfred:0-project` 명명 통일

---

## 공통 수정 사항

### 1. CLI 명령어 표시
**기존**:
```bash
moai init .
moai doctor
moai status
```

**변경**:
```bash
# 🚧 Coming in v0.4.0
moai init .
moai doctor
moai status
```

### 2. Alfred 커맨드 명명 통일
**기존**:
```text
/alfred:8-project
```

**변경**:
```text
/alfred:0-project
```

### 3. Python 예제 코드 추가
**기존** (TypeScript):
```typescript
// @CODE:AUTH-001
export class AuthService {
  async authenticate(username: string, password: string): Promise<AuthResult> {
    // ...
  }
}
```

**변경** (Python):
```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py
"""
@CODE:AUTH-001: JWT 인증 서비스

TDD 이력:
- RED: pytest 테스트 작성
- GREEN: bcrypt + PyJWT 구현
- REFACTOR: 타입 힌트 추가
"""

class AuthService:
    async def authenticate(self, username: str, password: str) -> AuthResult:
        """사용자 인증을 수행합니다."""
        # ...
```

---

## 검증 기준

### 필수 검증 항목
1. ✅ CLI 명령어 "🚧 Coming in v0.4.0" 배지 추가 (17개 파일)
2. ✅ `/alfred:0-project` 명명 통일 (17개 파일)
3. ✅ Python 예제 코드 추가 (주요 가이드 문서)
4. ✅ TypeScript 관련 내용 제거 (모든 파일)

### 선택 검증 항목
1. ⚠️ 모든 코드 예제가 Python 3.13+ 문법을 따르는지 확인
2. ⚠️ 각 문서 파일이 1000 줄 이하인지 확인
3. ⚠️ 모든 링크가 유효한지 확인

---

## 구현 전략

### 1단계: 파일 목록 확인 (1분)
```bash
# docs/ 디렉토리 구조 확인
find docs/ -name "*.md" -type f
```

### 2단계: 공통 수정 작업 (10분)
- 모든 파일에 CLI 명령어 "🚧" 배지 추가
- `/alfred:0-project` 명명 통일
- TypeScript 관련 내용 제거

### 3단계: Python 예제 코드 추가 (5분)
- docs/guides/*.md 파일에 Python 예제 추가
- docs/api/*.md 파일에 "Coming Soon" 배지 추가

### 4단계: 검증 (4분)
```bash
# CLI 명령어 확인
rg "moai (init|doctor|status)" docs/ -n

# Alfred 커맨드 확인
rg "/alfred:8-project" docs/ -n

# TypeScript 흔적 확인
rg "(typescript|bun|npm)" docs/ -i -n
```

---

## 우선순위

### Critical (즉시 수정 필요)
1. docs/getting-started/installation.md (TypeScript 설치 방법)
2. docs/getting-started/quick-start.md (Quick Start 재작성)
3. docs/index.md (메인 페이지)

### High (높은 우선순위)
4. docs/guides/alfred-superagent.md (Alfred 사용법)
5. docs/guides/spec-first-tdd.md (TDD 가이드)
6. docs/guides/tag-system.md (TAG 시스템)
7. docs/guides/trust-principles.md (TRUST 원칙)
8. docs/guides/workflow.md (워크플로우)

### Medium (중간 우선순위)
9-11. docs/agents/*.md (에이전트 문서)
12. docs/getting-started/first-project.md (첫 프로젝트)

### Low (낮은 우선순위)
13-15. docs/api/*.md (API 문서, v0.4.0 Coming Soon)
16-17. docs/specs/*.md (SPEC 문서)

---

## 관련 SPEC

- **SPEC-DOCS-004**: README.md Python v0.3.0 업데이트 (blocks DOCS-005)
- **SPEC-ALFRED-CMD-001**: Alfred 커맨드 명명 통일 (blocks DOCS-005)

---

## 참고 문서

- `README.md`: Python v0.3.0 기준 메인 문서
- `CLAUDE.md`: Alfred 커맨드 및 에이전트 설명
- `pyproject.toml`: Python v0.3.0 실제 의존성 목록
- `.moai/memory/development-guide.md`: TDD 워크플로우 및 TRUST 원칙
