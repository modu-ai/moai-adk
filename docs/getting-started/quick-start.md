# Quick Start Guide

> **3분 만에 MoAI-ADK 시작하기**
>
> SPEC → TDD → Sync 3단계 워크플로우 빠른 실습

---

## 📋 Table of Contents

- [준비물 체크리스트](#준비물-체크리스트)
- [Step 1: 프로젝트 초기화](#step-1-프로젝트-초기화-1분)
- [Step 2: Alfred와 인사하기](#step-2-alfred와-인사하기-30초)
- [Step 3: 첫 기능 개발](#step-3-첫-기능-개발-1분-30초)
- [전체 워크플로우 체크리스트](#전체-워크플로우-체크리스트)
- [출력 스타일 변경](#출력-스타일-변경)
- [다음 단계](#다음-단계)

---

## 준비물 체크리스트

시작하기 전에 다음 항목을 확인하세요:

### ✅ 필수 항목

- [ ] **Python 3.13+** 설치됨
  ```bash
  python --version
  # Python 3.13.0 이상
  ```

- [ ] **MoAI-ADK** 설치됨
  ```bash
  moai --version
  # moai-adk v0.3.0
  ```

- [ ] **Claude Code** 실행 중
  ```bash
  claude --version
  # Claude Code v1.2.0 이상
  ```

- [ ] **Git** 설치됨 (선택사항, 권장)
  ```bash
  git --version
  # git version 2.30.0 이상
  ```

### 📦 설치가 안 되었다면?

➡️ **[Installation Guide](./installation.md)** 참고

---

## Step 1: 프로젝트 초기화 (1분)

### 1-1. 새 프로젝트 생성

**터미널에서 실행**:

```bash
# 새 프로젝트 디렉토리 생성 + MoAI-ADK 설치
moai init my-first-project

# 프로젝트 디렉토리로 이동
cd my-first-project
```

**실행 결과 (예시)**:
```
🚀 MoAI-ADK 프로젝트 초기화 시작...

✅ 프로젝트 디렉토리 생성: /Users/goos/my-first-project
✅ .moai/ 디렉토리 생성
  ├── .moai/config.json         (프로젝트 설정)
  ├── .moai/memory/             (개발 가이드)
  ├── .moai/specs/              (SPEC 문서 저장소)
  └── .moai/reports/            (동기화 리포트)

✅ .claude/ 디렉토리 생성
  ├── .claude/custom-commands/  (Alfred 커맨드)
  ├── .claude/agents/           (10개 AI 에이전트)
  └── .claude/settings.json     (Claude Code 설정)

📝 생성된 파일: 24개
📁 생성된 디렉토리: 8개

🎯 다음 단계:
  1. cd my-first-project
  2. claude (Claude Code 실행)
  3. /alfred:0-project (프로젝트 컨텍스트 생성)
```

---

### 1-2. 기존 프로젝트에 설치

이미 프로젝트가 있다면, 현재 디렉토리에 MoAI-ADK를 설치할 수 있습니다:

```bash
# 프로젝트 디렉토리로 이동
cd existing-project

# MoAI-ADK 설치
moai init .
```

**실행 결과 (예시)**:
```
🚀 MoAI-ADK 설치 시작...

🔍 기존 프로젝트 감지:
  - 언어: Python 3.13
  - 프레임워크: FastAPI
  - Git 저장소: ✓

✅ .moai/ 디렉토리 생성
✅ .claude/ 디렉토리 생성
✅ 템플릿 파일 복사 완료

⚠️ 기존 파일 보호:
  - .moai/specs/ (사용자 SPEC 보존)
  - .moai/reports/ (리포트 보존)

🎯 다음 단계:
  1. claude (Claude Code 실행)
  2. /alfred:0-project (프로젝트 컨텍스트 생성)
```

---

### 1-3. 프로젝트 구조 확인

설치가 완료되면 다음과 같은 디렉토리 구조가 생성됩니다:

```
my-first-project/
├── .moai/                      # MoAI-ADK 시스템 디렉토리
│   ├── config.json             # 프로젝트 설정 (mode, locale 등)
│   ├── memory/                 # 개발 가이드 및 메모리
│   │   ├── development-guide.md   # 개발 워크플로우 가이드
│   │   └── spec-metadata.md       # SPEC 메타데이터 표준
│   ├── specs/                  # SPEC 문서 저장소 (사용자 작성)
│   │   └── .gitkeep
│   └── reports/                # 동기화 리포트
│       └── .gitkeep
│
├── .claude/                    # Claude Code 통합
│   ├── custom-commands/        # Alfred 커맨드
│   │   ├── 0-project.md
│   │   ├── 1-spec.md
│   │   ├── 2-build.md
│   │   └── 3-sync.md
│   ├── agents/                 # 10개 AI 에이전트
│   │   ├── alfred.yaml
│   │   ├── spec-builder.yaml
│   │   ├── code-builder.yaml
│   │   ├── doc-syncer.yaml
│   │   ├── tag-agent.yaml
│   │   ├── debug-helper.yaml
│   │   ├── trust-checker.yaml
│   │   ├── git-manager.yaml
│   │   ├── cc-manager.yaml
│   │   └── project-manager.yaml
│   └── settings.json           # Claude Code 설정
│
├── src/                        # 소스 코드 (사용자 작성)
├── tests/                      # 테스트 코드 (사용자 작성)
├── docs/                       # 문서 (자동 생성)
├── .gitignore                  # Git 제외 파일
└── README.md                   # 프로젝트 README
```

---

## Step 2: Alfred와 인사하기 (30초)

### 2-1. Claude Code 실행

```bash
# 터미널에서 Claude Code 실행
claude
```

Claude Code가 실행되면, 프로젝트 디렉토리를 엽니다.

---

### 2-2. Alfred 활성화 (/alfred:0-project)

**Claude Code 채팅 창에서** 다음 명령어를 입력:

```text
/alfred:0-project
```

**Alfred의 응답 (Phase 1: 분석)**:

```
📖 Alfred SuperAgent 시작합니다.

▶◀ 프로젝트 초기화 계획 수립 중...

🔍 Phase 1: 현재 프로젝트 분석

1️⃣ 프로젝트 구조 스캔:
  - 루트 디렉토리: /Users/goos/my-first-project
  - 파일 총 개수: 12개
  - 디렉토리 총 개수: 5개

2️⃣ 언어 및 프레임워크 감지:
  - Python 3.13.0
  - 패키지 관리자: pip (없음), poetry (없음)
  - 추천: uv 사용 권장

3️⃣ Git 저장소 확인:
  - Git 초기화: ✓
  - 현재 브랜치: main
  - 커밋 내역: 0개

4️⃣ 생성할 문서 계획:
  - .moai/project/product.md (200 lines) - 제품 개요, 목표, 핵심 기능
  - .moai/project/structure.md (150 lines) - 디렉토리 구조, 모듈 설계
  - .moai/project/tech.md (180 lines) - 기술 스택, 도구 체인

📋 계획 확인:
  - "진행" 또는 "시작" 입력 시 Phase 2 실행
  - "수정 [내용]" 입력 시 계획 재수립
  - "중단" 입력 시 작업 취소
```

**사용자 응답**:
```text
진행
```

**Alfred의 응답 (Phase 2: 실행)**:

```
✅ Phase 2: 프로젝트 초기화 실행

📝 1. product.md 작성 중...
✅ .moai/project/product.md (205 lines)

📝 2. structure.md 작성 중...
✅ .moai/project/structure.md (148 lines)

📝 3. tech.md 작성 중...
✅ .moai/project/tech.md (182 lines)

🎉 프로젝트 초기화 완료!

📊 생성된 파일 요약:
  - .moai/project/product.md   (제품 개요, 목표, 핵심 기능)
  - .moai/project/structure.md (디렉토리 구조, 모듈 설계)
  - .moai/project/tech.md      (Python 도구 체인, 테스트 전략)

🎯 다음 단계:
  /alfred:1-spec "첫 기능 설명"
```

---

### 2-3. 생성된 문서 확인

`/alfred:0-project`가 자동으로 생성한 문서를 확인해보세요:

```bash
# product.md 확인
cat .moai/project/product.md

# structure.md 확인
cat .moai/project/structure.md

# tech.md 확인
cat .moai/project/tech.md
```

---

## Step 3: 첫 기능 개발 (1분 30초)

이제 MoAI-ADK의 핵심인 **3단계 워크플로우**를 실습해봅시다:

### 3-1. SPEC 작성 (/alfred:1-spec)

**Claude Code에서 실행**:

```text
/alfred:1-spec "사용자 로그인 API"
```

**Alfred의 응답 (Phase 1: 분석)**:

```
🏗️ spec-builder 에이전트 시작합니다.

📋 Phase 1: SPEC 후보 분석

1️⃣ 요청 분석:
  - 기능: 사용자 로그인 API
  - 도메인: AUTH
  - 추정 복잡도: Medium

2️⃣ SPEC ID 생성:
  - 제안 ID: AUTH-001
  - 디렉토리: .moai/specs/SPEC-AUTH-001/
  - 파일명: spec.md

3️⃣ 중복 ID 확인:
  ✅ AUTH-001 사용 가능 (중복 없음)

4️⃣ EARS 명세 구조:
  - Ubiquitous: 기본 로그인 기능
  - Event-driven: 로그인 성공/실패 처리
  - State-driven: 인증 상태 관리
  - Optional: 소셜 로그인 지원
  - Constraints: 보안 제약사항

5️⃣ Git 브랜치 계획:
  - 브랜치명: feature/SPEC-AUTH-001
  - 베이스: main

📋 계획 확인:
  "진행" 또는 "시작" 입력 시 SPEC 문서 작성 시작
```

**사용자 응답**:
```text
진행
```

**Alfred의 응답 (Phase 2: 실행)**:

```
✅ Phase 2: SPEC 문서 작성 및 Git 작업

📝 1. SPEC 문서 작성 중...
✅ .moai/specs/SPEC-AUTH-001/spec.md (280 lines)

📂 생성된 SPEC 구조:
  ---
  id: AUTH-001
  version: 0.0.1
  status: draft
  created: 2025-10-14
  updated: 2025-10-14
  author: @Goos
  priority: high
  category: feature
  labels: [authentication, api]
  ---

  # @SPEC:AUTH-001: 사용자 로그인 API

  ## HISTORY
  ### v0.0.1 (2025-10-14)
  - **INITIAL**: 사용자 로그인 API 명세 작성

  ## Overview
  JWT 기반 사용자 로그인 API를 구현한다.

  ## Requirements (EARS)
  ### Ubiquitous Requirements
  - 시스템은 이메일/비밀번호 기반 로그인 기능을 제공해야 한다

  ### Event-driven Requirements
  - WHEN 사용자가 유효한 자격증명을 제공하면, 시스템은 JWT 토큰을 발급해야 한다
  - WHEN 로그인 시도가 3회 실패하면, 시스템은 계정을 일시 잠금해야 한다

  ...

🌿 2. Git 브랜치 생성...
✅ feature/SPEC-AUTH-001 브랜치 생성 완료

📤 3. Git 커밋...
✅ Commit: 🔴 RED: SPEC-AUTH-001 사용자 로그인 API 명세 작성

🎉 SPEC 작성 완료!

📊 생성된 파일:
  - .moai/specs/SPEC-AUTH-001/spec.md (280 lines)

🎯 다음 단계:
  /alfred:2-build AUTH-001
```

---

### 3-2. TDD 구현 (/alfred:2-build)

**Claude Code에서 실행**:

```text
/alfred:2-build AUTH-001
```

**Alfred의 응답 (Phase 1: 분석)**:

```
💎 code-builder 에이전트 시작합니다.

📋 Phase 1: TDD 구현 계획 수립

1️⃣ SPEC 문서 읽기:
  ✅ .moai/specs/SPEC-AUTH-001/spec.md
  - 요구사항: 12개
  - 제약사항: 5개
  - 보안 요구사항: 3개

2️⃣ TDD 사이클 계획:
  🔴 RED: 실패하는 테스트 작성
    - tests/auth/test_login.py
    - 테스트 케이스: 6개 (성공, 실패, 엣지케이스)

  🟢 GREEN: 최소 구현
    - src/auth/login_service.py
    - src/auth/jwt_manager.py

  🔵 REFACTOR: 코드 품질 개선
    - 타입 힌트 추가
    - 에러 처리 강화
    - TRUST 5원칙 적용

3️⃣ 의존성 확인:
  - bcrypt (비밀번호 해싱)
  - PyJWT (JWT 토큰)
  - pytest (테스트)

📋 계획 확인:
  "진행" 또는 "시작" 입력 시 TDD 구현 시작
```

**사용자 응답**:
```text
진행
```

**Alfred의 응답 (Phase 2: 실행 - RED)**:

```
✅ Phase 2: TDD 구현

🔴 RED: 실패하는 테스트 작성

📝 1. tests/auth/test_login.py 작성...
✅ 테스트 파일 생성 (120 lines)

# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001/spec.md

import pytest
from src.auth.login_service import LoginService

class TestLoginService:
    def test_login_with_valid_credentials(self):
        """유효한 자격증명으로 로그인 성공"""
        service = LoginService()
        result = service.login("user@example.com", "password123")
        assert result.success is True
        assert result.token is not None

    def test_login_with_invalid_password(self):
        """잘못된 비밀번호로 로그인 실패"""
        service = LoginService()
        result = service.login("user@example.com", "wrong")
        assert result.success is False
        assert result.error == "Invalid credentials"

    ...

🧪 테스트 실행...
❌ FAILED tests/auth/test_login.py::TestLoginService::test_login_with_valid_credentials
  ModuleNotFoundError: No module named 'src.auth.login_service'

✅ RED 단계 완료: 테스트가 실패합니다 (예상된 동작)

📤 Git 커밋...
✅ Commit: 🔴 RED: SPEC-AUTH-001 로그인 서비스 테스트 작성
```

**Alfred의 응답 (Phase 2: 실행 - GREEN)**:

```
🟢 GREEN: 최소 구현으로 테스트 통과

📝 1. src/auth/login_service.py 작성...
✅ 구현 파일 생성 (85 lines)

# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001/spec.md | TEST: tests/auth/test_login.py

"""
@CODE:AUTH-001: 사용자 로그인 서비스

TDD 이력:
- RED: tests/auth/test_login.py 작성
- GREEN: 최소 구현 (bcrypt, PyJWT)
"""

from dataclasses import dataclass
import bcrypt
import jwt

@dataclass
class LoginResult:
    success: bool
    token: str | None = None
    error: str | None = None

class LoginService:
    # @CODE:AUTH-001:API: 로그인 API
    def login(self, email: str, password: str) -> LoginResult:
        # @CODE:AUTH-001:DOMAIN: 입력 검증
        if not self._validate_input(email, password):
            return LoginResult(success=False, error="Invalid input")

        # @CODE:AUTH-001:DATA: 사용자 조회
        user = self._find_user(email)
        if not user:
            return LoginResult(success=False, error="Invalid credentials")

        # @CODE:AUTH-001:DOMAIN: 비밀번호 검증
        if not bcrypt.checkpw(password.encode(), user.password_hash):
            return LoginResult(success=False, error="Invalid credentials")

        # @CODE:AUTH-001:API: JWT 토큰 발급
        token = self._generate_token(user)
        return LoginResult(success=True, token=token)

    ...

🧪 테스트 실행...
✅ PASSED tests/auth/test_login.py::TestLoginService::test_login_with_valid_credentials
✅ PASSED tests/auth/test_login.py::TestLoginService::test_login_with_invalid_password
✅ PASSED (6/6 tests)

📊 커버리지 리포트:
  - src/auth/login_service.py: 92%

✅ GREEN 단계 완료: 모든 테스트 통과

📤 Git 커밋...
✅ Commit: 🟢 GREEN: SPEC-AUTH-001 로그인 서비스 구현
```

**Alfred의 응답 (Phase 2: 실행 - REFACTOR)**:

```
🔵 REFACTOR: 코드 품질 개선

📝 1. 타입 힌트 강화...
✅ mypy 타입 체크 통과

📝 2. 에러 처리 개선...
✅ 커스텀 예외 클래스 추가

📝 3. TRUST 5원칙 적용...
  ✅ Test First: 테스트 커버리지 92%
  ✅ Readable: 함수당 평균 15 LOC
  ✅ Unified: 타입 안전성 보장
  ✅ Secured: bcrypt 해싱, SQL Injection 방어
  ✅ Trackable: @TAG 체인 완전

📝 4. 코드 포맷팅...
✅ ruff format 적용

🧪 테스트 재실행...
✅ PASSED (6/6 tests)

✅ REFACTOR 단계 완료: 코드 품질 개선

📤 Git 커밋...
✅ Commit: ♻️ REFACTOR: SPEC-AUTH-001 코드 품질 개선

🎉 TDD 구현 완료!

📊 생성된 파일:
  - tests/auth/test_login.py (120 lines)
  - src/auth/login_service.py (95 lines)
  - src/auth/jwt_manager.py (60 lines)

📈 품질 지표:
  - 테스트 커버리지: 92%
  - 타입 안전성: ✓
  - TRUST 5원칙: ✓

🎯 다음 단계:
  /alfred:3-sync
```

---

### 3-3. 문서 동기화 (/alfred:3-sync)

**Claude Code에서 실행**:

```text
/alfred:3-sync
```

**Alfred의 응답 (Phase 1: 분석)**:

```
📖 doc-syncer 에이전트 시작합니다.

📋 Phase 1: 동기화 범위 분석

1️⃣ Git 상태 확인:
  - 브랜치: feature/SPEC-AUTH-001
  - 변경 파일: 3개
    - .moai/specs/SPEC-AUTH-001/spec.md
    - tests/auth/test_login.py
    - src/auth/login_service.py

2️⃣ TAG 체인 검증:
  ✅ @SPEC:AUTH-001 → .moai/specs/SPEC-AUTH-001/spec.md
  ✅ @TEST:AUTH-001 → tests/auth/test_login.py
  ✅ @CODE:AUTH-001 → src/auth/login_service.py
  ✅ TAG 체인 무결성: 완전

3️⃣ 고아 TAG 확인:
  ✅ 고아 TAG 없음

4️⃣ Living Document 생성 계획:
  - docs/api/auth.md (자동 생성)
  - README.md 업데이트 (API 섹션 추가)

5️⃣ SPEC 버전 업데이트 계획:
  - AUTH-001: v0.0.1 (draft) → v0.1.0 (completed)

📋 계획 확인:
  "진행" 또는 "시작" 입력 시 동기화 실행
```

**사용자 응답**:
```text
진행
```

**Alfred의 응답 (Phase 2: 실행)**:

```
✅ Phase 2: 문서 동기화 및 품질 검증

📝 1. Living Document 생성...
✅ docs/api/auth.md (150 lines)
  - API 엔드포인트 자동 추출
  - 요청/응답 예시 생성
  - 에러 코드 문서화

📝 2. README.md 업데이트...
✅ API 섹션 추가

📝 3. SPEC 버전 업데이트...
✅ SPEC-AUTH-001: v0.0.1 → v0.1.0
✅ status: draft → completed

📝 4. 동기화 리포트 생성...
✅ .moai/reports/sync-report-2025-10-14.md

📊 TAG 추적성 통계:
  - @SPEC 태그: 1개
  - @TEST 태그: 1개
  - @CODE 태그: 1개
  - @DOC 태그: 1개
  - TAG 체인 무결성: 100%

📤 5. Git 커밋...
✅ Commit: 📝 DOCS: SPEC-AUTH-001 문서 동기화

🎉 문서 동기화 완료!

📊 생성/업데이트된 파일:
  - docs/api/auth.md (150 lines)
  - README.md (updated)
  - .moai/specs/SPEC-AUTH-001/spec.md (version: v0.1.0)
  - .moai/reports/sync-report-2025-10-14.md

🎯 다음 단계:
  - 새 기능 개발: /alfred:1-spec "다음 기능 설명"
  - 품질 검증: @agent-trust-checker
  - Git 작업: @agent-git-manager "PR 생성"
```

---

## 전체 워크플로우 체크리스트

MoAI-ADK의 3단계 워크플로우를 완료했습니다! 🎉

### ✅ 완료된 작업

- [x] **0단계: 프로젝트 초기화** (`/alfred:0-project`)
  - [x] .moai/project/product.md 생성
  - [x] .moai/project/structure.md 생성
  - [x] .moai/project/tech.md 생성

- [x] **1단계: SPEC 작성** (`/alfred:1-spec`)
  - [x] .moai/specs/SPEC-AUTH-001/spec.md 생성
  - [x] EARS 형식 요구사항 작성
  - [x] Git 브랜치 생성 (feature/SPEC-AUTH-001)

- [x] **2단계: TDD 구현** (`/alfred:2-build`)
  - [x] 🔴 RED: tests/auth/test_login.py 작성
  - [x] 🟢 GREEN: src/auth/login_service.py 구현
  - [x] 🔵 REFACTOR: 코드 품질 개선

- [x] **3단계: 문서 동기화** (`/alfred:3-sync`)
  - [x] docs/api/auth.md 자동 생성
  - [x] TAG 체인 검증 완료
  - [x] SPEC 버전 업데이트 (v0.1.0)

### 📊 생성된 아티팩트

| 파일                                     | 라인 수 | 설명                   |
| ---------------------------------------- | ------- | ---------------------- |
| `.moai/specs/SPEC-AUTH-001/spec.md`     | 280     | EARS 형식 명세서       |
| `tests/auth/test_login.py`              | 120     | 테스트 코드            |
| `src/auth/login_service.py`             | 95      | 구현 코드              |
| `docs/api/auth.md`                       | 150     | API 문서 (자동 생성)   |
| `.moai/reports/sync-report-2025-10-14.md` | 80      | 동기화 리포트          |

**총 라인 수**: 725 lines

---

## 출력 스타일 변경

Alfred는 개발 상황에 따라 **4가지 대화 스타일**을 제공합니다.

### 🎨 사용 가능한 스타일

| 스타일                      | 대상           | 특징                         |
| --------------------------- | -------------- | ---------------------------- |
| **MoAI Professional**       | 실무 개발자    | 간결, 기술적, 결과 중심      |
| **MoAI Beginner Learning**  | 개발 입문자    | 친절, 상세 설명, 단계별 안내 |
| **MoAI Pair Collaboration** | 협업 개발자    | 질문 기반, 브레인스토밍      |
| **MoAI Study Deep**         | 신기술 학습자  | 개념 → 실습 → 전문가 팁      |

### 🔄 스타일 전환 방법

**Claude Code에서 실행**:

```bash
# 실무 개발 스타일 (기본값)
/output-style alfred-pro

# 학습 스타일
/output-style beginner-learning

# 협업 스타일
/output-style pair-collab

# 심화 학습 스타일
/output-style study-deep
```

**예시: 학습 스타일 전환**:
```text
/output-style beginner-learning
```

**Alfred의 응답**:
```
✅ 출력 스타일 변경 완료!

📚 **MoAI Beginner Learning** 스타일이 활성화되었습니다.

이제부터 Alfred는:
- 각 단계를 상세하게 설명합니다
- 코드 예시를 더 많이 제공합니다
- 개념 설명을 친절하게 추가합니다
- 실수하기 쉬운 부분을 미리 안내합니다

💡 다시 전문가 스타일로 돌아가려면:
  /output-style alfred-pro
```

---

## 다음 단계

### 1. 실전 튜토리얼: Todo 앱 만들기

실제 Todo 앱을 처음부터 끝까지 만들면서 MoAI-ADK를 깊이 이해하세요:

➡️ **[First Project Tutorial](./first-project.md)**

**다루는 내용**:
- 프로젝트 설계 및 SPEC 작성
- 다중 SPEC 관리 (USER-001, TODO-001, AUTH-001)
- RESTful API 설계 및 TDD 구현
- 데이터베이스 통합 (SQLAlchemy)
- API 문서 자동 생성
- Git 워크플로우 (브랜치, PR, 머지)

---

### 2. Alfred SuperAgent 심화 가이드

10개 AI 에이전트를 활용한 고급 개발 기법:

➡️ **[Alfred SuperAgent Guide](https://moai-adk.vercel.app/guides/alfred-superagent/)**

**다루는 내용**:
- 에이전트 조율 전략 (Sequential, Parallel)
- 온디맨드 에이전트 호출 (`@agent-*`)
- 디버깅 전략 (`@agent-debug-helper`)
- TAG 시스템 관리 (`@agent-tag-agent`)
- Git 워크플로우 자동화 (`@agent-git-manager`)

---

### 3. SPEC-First TDD 방법론

EARS 형식 명세 작성 및 TDD 사이클 마스터:

➡️ **[SPEC-First TDD Guide](https://moai-adk.vercel.app/guides/spec-first-tdd/)**

**다루는 내용**:
- EARS 5가지 구문 (Ubiquitous, Event-driven, State-driven, Optional, Constraints)
- SPEC 메타데이터 표준 (필수 7개 + 선택 9개 필드)
- TDD RED-GREEN-REFACTOR 사이클
- TRUST 5원칙 적용 (Test, Readable, Unified, Secured, Trackable)

---

### 4. TAG 시스템 완전 가이드

코드 추적성을 보장하는 @TAG 시스템:

➡️ **[TAG System Guide](https://moai-adk.vercel.app/guides/tag-system/)**

**다루는 내용**:
- TAG 체계 철학 (`@SPEC → @TEST → @CODE → @DOC`)
- CODE-FIRST 원칙 (코드 직접 스캔)
- TAG 무결성 검증 및 고아 TAG 탐지
- 언어별 TAG 사용 예시 (Python, TypeScript, Java, Go, Rust, Dart, Swift, Kotlin)

---

### 5. 온디맨드 에이전트 활용

필요 시 즉시 호출하는 전문 에이전트:

```text
# 디버깅 & 분석
@agent-debug-helper "TypeError: 'NoneType' object"

# TAG 관리
@agent-tag-agent "AUTH 도메인 TAG 목록 조회"

# TRUST 검증
@agent-trust-checker "테스트 커버리지 확인"

# Git 작업
@agent-git-manager "PR 생성"
```

---

## 커뮤니티 및 지원

### 📖 공식 문서

- **[전체 문서 사이트](https://moai-adk.vercel.app)**
- **[API Reference](https://moai-adk.vercel.app/api/)**

### 💬 커뮤니티

- **[GitHub Issues](https://github.com/modu-ai/moai-adk/issues)** - 버그 리포트, 기능 요청
- **[GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)** - 질문, 아이디어, 피드백

### 📦 패키지

- **[PyPI Package](https://pypi.org/project/moai-adk/)** - Python 패키지

---

**마지막 업데이트**: 2025-10-14
**버전**: v0.3.0
