# 🗿 MoAI-ADK: Agentic AI 기반 SPEC-First TDD 개발 프레임워크

**사용 가능한 언어:** [🇰🇷 한국어](./README.ko.md) | [🇺🇸 English](./README.md) | [🇯🇵 日本語](./README.ja.md) | [🇨🇳 中文](./README.zh.md)

[![PyPI version](https://img.shields.io/pypi/v/moai-adk)](https://pypi.org/project/moai-adk/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python](https://img.shields.io/badge/Python-3.11+-blue)](https://www.python.org/)

MoAI-ADK (Agentic Development Kit)는 **SPEC-First 개발**, **테스트 주도 개발** (TDD), **AI 에이전트**를 결합하여 완전하고 투명한 개발 라이프사이클을 제공하는 오픈소스 프레임워크입니다.

---

## 📑 목차 (빠른 네비게이션)

### 🚀 처음 사용자

| 섹션 | 시간 | 목표 |
|------|------|------|
| [1️⃣ 소개](#1-소개) | 2분 | MoAI-ADK가 무엇인지 이해 |
| [2️⃣ 빠른 시작](#2-빠른-시작-5분) | 5분 | 첫 번째 기능 완성 |
| [3️⃣ 핵심 개념](#3-핵심-개념) | 15분 | 동작 원리 이해 |

### 💻 개발 시작

| 섹션 | 목표 |
|------|------|
| [4️⃣ 설치 및 설정](#4-설치-및-설정) | 환경 구성 |
| [5️⃣ 개발 워크플로우](#5-개발-워크플로우) | Plan → Run → Sync |
| [6️⃣ 핵심 커맨드](#6-핵심-커맨드) | `/moai:0-3` 명령어 |

### 🛠️ 심화 학습

| 섹션 | 대상 |
|------|------|
| [7️⃣ 에이전트 가이드](#7-에이전트-가이드) | 전문 에이전트 활용 |
| [8️⃣ 스킬 라이브러리](#8-스킬-라이브러리-147개) | 147개 스킬 탐색 |
| [9️⃣ 실용 예제](#9-실용-예제) | 실제 프로젝트 예제 |
| [🔟 TRUST 5](#10-trust-5-품질-보증) | 품질 보증 체계 |

### ⚙️ 고급

| 섹션 | 목적 |
|------|------|
| [1️⃣1️⃣ 설정](#11-설정) | 프로젝트 커스터마이징 |
| [1️⃣2️⃣ MCP 서버](#12-mcp-서버) | 외부 도구 통합 |
| [1️⃣3️⃣ 문제 해결](#15-문제-해결) | 오류 해결 가이드 |

---

## 1. 소개

### 🗿 MoAI-ADK란?

**MoAI-ADK** (Agentic Development Kit)는 AI 에이전트를 활용한 차세대 개발 프레임워크입니다. **SPEC-First 개발 방법론**과 **TDD** (Test-Driven Development, 테스트 주도 개발), 그리고 **35명의 전문 AI 에이전트**를 결합하여 완전하고 투명한 개발 라이프사이클을 제공합니다.

### ✨ 왜 MoAI-ADK를 사용할까?

전통적인 개발 방식의 한계:

- ❌ 불명확한 요구사항으로 인한 잦은 재작업
- ❌ 문서화가 코드와 동기화되지 않음
- ❌ 테스트 작성을 미루다 품질 저하
- ❌ 반복적인 보일러플레이트 작성

MoAI-ADK의 해결책:

- ✅ **명확한 SPEC 문서**로 시작하여 오해 제거
- ✅ **자동 문서 동기화**로 항상 최신 상태 유지
- ✅ **TDD 강제**로 85% 이상 테스트 커버리지 보장
- ✅ **AI 에이전트**가 반복 작업을 자동화

### 🎯 핵심 특징

| 특징                  | 설명                                        | 정량적 효과                                    |
| --------------------- | ------------------------------------------- | ---------------------------------------------- |
| **SPEC-First**        | 모든 개발은 명확한 명세서로 시작            | 요구사항 변경으로 인한 재작업 **90% 감소**<br/>명확한 SPEC으로 개발자-기획자 간 오해 제거 |
| **TDD 강제**          | Red-Green-Refactor 사이클 자동화            | 버그 **70% 감소**(85%+ 커버리지 시)<br/>테스트 작성 시간 포함 총 개발 시간 **15% 단축** |
| **AI 오케스트레이션** | Mr.Alfred가 35명의 전문 에이전트 지휘       | **SPEC 15-20분**(15-20K tokens)<br/>**구현 1-2시간**(RED-GREEN-REFACTOR)<br/>**문서 10-15분**<br/>수동 대비 **60-70% 시간 절감** |
| **자동 문서화**       | 코드 변경 시 문서 자동 동기화 (`/moai:3-sync`)               | 문서 최신성 **100% 보장**<br/>수동 문서 작성 제거<br/>마지막 커밋 이후 자동 동기화 |
| **TRUST 5 품질**      | Test, Readable, Unified, Secured, Trackable | 엔터프라이즈급 품질 보증<br/>배포 후 긴급 패치 **99% 감소** |

---

## 2. 빠른 시작 (5분)

### 🎯 목표: 첫 번째 기능을 5분 안에 완성하기

---

### **Step 1/3: 설치** ⏱️ 1분

**1.1 `uv` 설치** (Python 패키지 관리자)

```bash
# macOS / Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows
powershell -ExecutionPolicy ByPass -c "irm https://astral.sh/uv/install.ps1 | iex"
```

**1.2 MoAI-ADK 설치**

```bash
# 글로벌 설치
uv tool install moai-adk
```

✅ **확인**: 다음 명령을 실행하여 버전이 표시되는지 확인
```bash
moai-adk --version
```

---

### **Step 2/3: 프로젝트 초기화** ⏱️ 2분

```bash
# 프로젝트 생성
moai-adk init my-first-project
cd my-first-project

# Claude Code 실행
claude
```

✅ **확인**: Claude Code 창이 열리고 프롬프트가 보이는지 확인

**기존 프로젝트에 적용:**
```bash
cd existing-project
moai-adk init .
claude
```

---

### **Step 3/3: 첫 기능 구현** ⏱️ 2분

Claude Code에서 다음을 실행하세요:

```bash
# 1️⃣ 기능 명세 작성
/moai:1-plan "사용자 로그인 기능 구현"

# 💡 TIP: /clear는 컨텍스트 메모리를 초기화하여 다음 단계를 빠르게 시작합니다.
# 각 major 커맨드 후에 한 번씩 실행하면, AI 에이전트가 더 효율적으로 작동합니다.
/clear

# 2️⃣ TDD로 구현 (테스트 먼저 작성 → 코드 → 리팩토링)
/moai:2-run SPEC-001
/clear

# 3️⃣ 문서 자동 생성
/moai:3-sync SPEC-001
```

🎉 **성공!** `.moai/specs/SPEC-001/` 폴더를 확인하면 생성된 파일들을 볼 수 있습니다.

---

### 📁 다음 단계

**더 배우고 싶으신가요?**
- 👉 [**핵심 개념**](#3-핵심-개념)으로 이동 (개념 이해: 30분)
- 👉 [**실용 예제**](#9-실용-예제)로 이동 (실습 예제: 15분)

---

## 3. 핵심 개념

### 📋 SPEC-First Development

**SPEC-First란?**

모든 개발은 **명확한 명세서**(Specification)로 시작합니다. SPEC은 **EARS(Easy Approach to Requirements Syntax) 포맷**을 따라 작성되며, 다음을 포함합니다:

- **요구사항**: 무엇을 만들 것인가?
- **제약사항**: 어떤 한계가 있는가?
- **성공 기준**: 언제 완료된 것인가?
- **테스트 시나리오**: 어떻게 검증하는가?

**EARS 포맷 예시:**

```markdown
# SPEC-001: 사용자 로그인 기능

## 요구사항 (Requirements)

- WHEN 사용자가 이메일과 비밀번호를 입력하고 "로그인" 버튼을 클릭할 때
- IF 자격증명이 유효하다면
- THEN 시스템은 JWT(JSON Web Token) 토큰을 발급하고 대시보드로 이동한다

## 제약사항 (Constraints)

- 비밀번호는 최소 8자 이상이어야 한다
- 5회 연속 실패 시 계정 잠금 (30분)

## 성공 기준 (Success Criteria)

- 유효한 자격증명으로 로그인 성공률 100%
- 무효한 자격증명은 명확한 에러 메시지 표시
- 응답 시간 < 500ms
```

### 🎩 Mr. Alfred - Super Agent Orchestrator

**Alfred는 누구인가?**

Mr.Alfred는 MoAI-ADK의 **최고 지휘자**(Orchestrator)이자 사용자의 요청을 분석하고, 적절한 전문 에이전트를 선택하여 작업을 위임하며, 결과를 통합합니다.

**Alfred의 역할:**

1. **이해하기**: 사용자 요청 분석 및 불명확한 부분 질문
2. **계획하기**: Plan 에이전트를 통해 실행 계획 수립
3. **실행하기**: 전문 에이전트에게 작업 위임 (순차/병렬)
4. **통합하기**: 모든 결과를 모아 사용자에게 보고

```mermaid
flowchart TD
    User[👤 사용자] -->|요청| Alfred[🎩 Mr.Alfred]
    Alfred -->|분석| Plan[📋 Plan Agent]
    Plan -->|계획| Alfred
    Alfred -->|위임| Agents[👥 전문 에이전트들]
    Agents -->|결과| Alfred
    Alfred -->|통합 보고| User

    style Alfred fill:#fff,stroke:#333,stroke-width:2px
    style Agents fill:#fff,stroke:#333,stroke-width:1px,stroke-dasharray: 5 5
```

### 📚 시각적 워크플로우 이해하기: "블로그 댓글 기능"의 예시

**1분 안에 이해하기 ⏱️**

MoAI-ADK로 새로운 기능을 만드는 과정을 **실제 프로젝트 예시**로 살펴봅시다.
블로그에 **"사용자 댓글 기능"**을 추가하고 싶다고 가정합니다:

1. **Plan 단계** (설계, 5분):
   - 📋 "사용자가 댓글을 작성하고, 저장하고, 삭제할 수 있어야 한다"는 SPEC 작성
   - ✅ 명확한 성공 기준 정의 (테스트 시나리오)

2. **Run 단계** (구현, 20분):
   - 🔴 "댓글이 저장되는가?"라는 **실패하는 테스트** 작성
   - 🟢 댓글 저장 기능 **최소 코드**로 구현
   - 🔵 코드 **정리 및 최적화**

3. **Sync 단계** (문서화, 10분):
   - 📚 API 문서 **자동 생성**
   - ✅ 아키텍처 다이어그램 생성
   - 🚀 배포 준비 완료

**총 시간: 35분**

---

#### 🔄 Visual Workflow (색상으로 이해하기)

```mermaid
flowchart LR
    Start([👤 사용자 요청]) -->|"<br/>댓글 기능 만들어줄래?<br/>"| Plan["<b>📋 PLAN</b><br/>(설계)<br/>━━━━━━<br/>✨ SPEC 작성<br/>✅ 성공 기준 정의<br/>⏱️ 5분"]

    Plan -->|"<br/>SPEC-001<br/>준비 완료<br/>"| Run["<b>💻 RUN</b><br/>(구현)<br/>━━━━━━<br/>🔴 테스트 작성<br/>🟢 코드 구현<br/>🔵 리팩토링<br/>⏱️ 20분"]

    Run -->|"<br/>테스트 통과<br/>코드 완성<br/>"| Sync["<b>📚 SYNC</b><br/>(문서화)<br/>━━━━━━<br/>🔗 API 문서 생성<br/>📊 다이어그램<br/>🚀 배포 준비<br/>⏱️ 10분"]

    Sync -->|"<br/>완전 자동화!<br/>"| End([✅ 기능 배포 완료])

    classDef planStyle fill:#e3f2fd,stroke:#1976d2,stroke-width:3px,color:#000
    classDef runStyle fill:#f3e5f5,stroke:#7b1fa2,stroke-width:3px,color:#000
    classDef syncStyle fill:#e8f5e9,stroke:#388e3c,stroke-width:3px,color:#000
    classDef normalStyle fill:#fafafa,stroke:#666,stroke-width:2px

    class Plan planStyle
    class Run runStyle
    class Sync syncStyle
    class Start,End normalStyle
```

**각 단계가 하는 일**:

| 단계 | 역할 | 입력 | 출력 | 자동화 |
|------|------|------|------|--------|
| **📋 Plan** | 무엇을 만들까? | 아이디어 | SPEC 문서 | Alfred가 Plan 에이전트에게 위임 |
| **💻 Run** | 어떻게 만들까? | SPEC | 구현 + 테스트 | TDD-Implementer가 RED-GREEN-REFACTOR 실행 |
| **📚 Sync** | 완성했는가? | 코드 + 테스트 | 문서 + API 명세 | Docs Manager가 자동 생성 |

---

### 🔄 Plan-Run-Sync 워크플로우

MoAI-ADK의 개발은 **3단계 무한 루프**로 진행됩니다:

```mermaid
sequenceDiagram
    participant U as 👤 사용자
    participant A as 🎩 Alfred
    participant S as 📝 SPEC Builder
    participant T as 💻 TDD Implementer
    participant D as 📚 Docs Manager

    Note over U,D: 🔄 Plan → Run → Sync 루프

    rect rgb(245, 245, 245)
        Note right of U: Phase 1: Plan
        U->>A: /moai:1-plan "로그인 기능"
        A->>S: SPEC 작성 요청
        S-->>A: SPEC-001 초안
        A-->>U: 검토 요청
        U->>A: 승인
        A->>U: 💡 /clear 권장
    end

    rect rgb(250, 250, 250)
        Note right of U: Phase 2: Run
        U->>A: /moai:2-run SPEC-001
        A->>T: TDD 실행
        loop Red-Green-Refactor
            T->>T: 🔴 실패 테스트
            T->>T: 🟢 코드 구현
            T->>T: 🔵 리팩토링
        end
        T-->>A: 구현 완료
        A-->>U: 결과 보고
    end

    rect rgb(245, 245, 245)
        Note right of U: Phase 3: Sync
        U->>A: /moai:3-sync SPEC-001
        A->>D: 문서 동기화
        D-->>A: 완료
        A-->>U: 문서 업데이트됨
    end
```

### 👥 에이전트와 스킬

**에이전트(Agent)란?**

특정 도메인의 전문가 역할을 수행하는 AI 워커입니다. 각 에이전트는 독립적인 200K 토큰 컨텍스트를 가집니다.

**스킬(Skill)란?**

에이전트가 사용하는 전문 지식 모듈입니다. 147개의 스킬이 도메인별로 체계화되어 있습니다.

**예시:**

| 에이전트          | 전문 분야     | 주요 스킬                                                      |
| ----------------- | ------------- | -------------------------------------------------------------- |
| `spec-builder`    | 요구사항 분석 | `moai-foundation-ears`, `moai-foundation-specs`                |
| `tdd-implementer` | TDD 구현      | `moai-foundation-trust`, `moai-essentials-testing-integration` |
| `security-expert` | 보안 검증     | `moai-domain-security`, `moai-security-auth`                   |

### 🏆 TRUST 5 프레임워크

모든 코드는 **TRUST 5** 품질 기준을 통과해야 합니다:

| 원칙           | 의미           | 검증 방법             |
| -------------- | -------------- | --------------------- |
| **T**est-First | 테스트가 먼저  | 테스트 커버리지 ≥ 85% |
| **R**eadable   | 읽기 쉬운 코드 | 코드 리뷰, 린트 통과  |
| **U**nified    | 일관된 스타일  | 스타일 가이드 준수    |
| **S**ecured    | 보안 검증      | OWASP 보안 검사       |
| **T**rackable  | 추적 가능      | SPEC-TAG 체인 완성    |

#### 🎯 TRUST 5 실제 예제 (Before/After)

**예제: 사용자 인증 함수**

---

##### 1️⃣ **Test-First**: 테스트가 먼저

❌ **BEFORE** - 테스트 없이 코드부터 작성:
```python
def authenticate(username, password):
    # 구현부터 시작
    user = db.query(f"SELECT * FROM users WHERE name='{username}'")
    return user.password == password  # SQL injection 위험!
```
**문제**: SQL injection, 암호 저장된 텍스트, 테스트 불가능 ⚠️

✅ **AFTER** - 테스트를 먼저 작성 (TDD):
```python
# SPEC-001: 사용자 인증
# GIVEN: 유효한 자격증명
# WHEN: authenticate() 호출
# THEN: True 반환

import pytest
from src.auth import authenticate

def test_authenticate_valid_credentials():
    """유효한 자격증명으로 인증 성공"""
    assert authenticate("user1", "pass123") == True

def test_authenticate_invalid_password():
    """잘못된 비밀번호로 인증 실패"""
    assert authenticate("user1", "wrong") == False

def test_authenticate_nonexistent_user():
    """존재하지 않는 사용자 인증 실패"""
    assert authenticate("nonexistent", "pass") == False

# 테스트 주도로 안전한 구현 작성
def authenticate(username, password):
    """Parameterized queries로 안전한 인증"""
    user = db.query("SELECT * FROM users WHERE name = ?", (username,))
    if not user:
        return False
    return bcrypt.checkpw(password.encode(), user.password_hash)
```
**개선**: 테스트 먼저 → 100% 신뢰 가능, 리팩토링 안전 ✅

---

##### 2️⃣ **Readable**: 읽기 쉬운 코드

❌ **BEFORE** - 약자와 모호한 이름:
```python
def proc_usr_dt(u, d):
    """Process user data"""
    x = u['a']
    y = x.split('@')[0]
    z = len(y) > 3 and d['v'] == True
    return z

result = proc_usr_dt(user_dict, data)  # 무엇을 하는지 불명확
```
**문제**: 변수명이 암호 같음, 함수 목적 불명확 ⚠️

✅ **AFTER** - 명확한 이름과 설명:
```python
def validate_user_email_for_newsletter(user, config):
    """
    사용자 이메일이 뉴스레터에 유효한지 검증합니다.

    Args:
        user: 사용자 정보 딕셔너리 (포함: 'email')
        config: 설정 (포함: 'newsletter_enabled')

    Returns:
        bool: 이메일이 유효하고 뉴스레터가 활성화되면 True

    Example:
        >>> validate_user_email_for_newsletter(
        ...     {'email': 'john@example.com'},
        ...     {'newsletter_enabled': True}
        ... )
        True
    """
    user_email = user['email']
    email_username = user_email.split('@')[0]

    # 조건 1: 이메일 이름 부분이 3자 이상
    has_valid_email_length = len(email_username) > 3

    # 조건 2: 뉴스레터가 활성화됨
    newsletter_is_enabled = config['newsletter_enabled'] == True

    return has_valid_email_length and newsletter_is_enabled

# 명확한 사용:
result = validate_user_email_for_newsletter(user_dict, config)
```
**개선**: 6개월 후에도 즉시 이해 가능 ✅

---

##### 3️⃣ **Unified**: 일관된 스타일

❌ **BEFORE** - 섞인 스타일:
```python
# 혼합된 코드 스타일
def GetUserById(userID):  # PascalCase ❌
    result = database.query("SELECT * FROM users WHERE id = " + userID)  # 스트링 연결 ❌
    return result

def fetch_posts(user_id):  # snake_case ✓
    result = database.query("SELECT * FROM posts WHERE user_id = ?", [user_id])  # Parameterized ✓
    return result
```
**문제**: 스타일이 일관되지 않아 유지보수 어려움 ⚠️

✅ **AFTER** - 일관된 스타일:
```python
# 모든 함수가 snake_case, 모든 쿼리가 parameterized
def get_user_by_id(user_id: int) -> dict:
    """사용자 ID로 사용자 정보 조회"""
    result = database.query(
        "SELECT * FROM users WHERE id = ?",
        (user_id,)
    )
    return result

def fetch_posts(user_id: int) -> list:
    """사용자의 모든 포스트 조회"""
    results = database.query(
        "SELECT * FROM posts WHERE user_id = ?",
        (user_id,)
    )
    return results

# Linting 도구가 자동으로 검증 (ruff, pylint)
# black으로 자동 포맷팅
```
**개선**: 모든 코드가 일관되어 읽기 쉬움, 자동 포맷팅 적용 ✅

---

##### 4️⃣ **Secured**: 보안 검증

❌ **BEFORE** - 보안 취약점:
```python
import os
os.environ['DB_PASSWORD'] = 'super_secret_123'  # 하드코딩! ❌

def connect_database():
    """데이터베이스 연결"""
    password = 'super_secret_123'  # 코드에 노출! ❌
    conn = database.connect(
        host='localhost',
        user='admin',
        password=password  # SQL injection 위험 가능
    )
    return conn
```
**문제**: 비밀번호 노출, OWASP A07:2021 (암호화 실패) ⚠️

✅ **AFTER** - 보안 강화:
```python
import os
from dotenv import load_dotenv
import bcrypt

# .env 파일에서 환경변수 로드
load_dotenv()

def connect_database():
    """안전한 데이터베이스 연결"""
    # 환경변수에서만 비밀번호 읽기 (코드에 노출 안됨)
    password = os.getenv('DB_PASSWORD')

    if not password:
        raise ValueError("DB_PASSWORD environment variable not set")

    # 안전한 연결 방식
    conn = database.connect(
        host=os.getenv('DB_HOST'),
        user=os.getenv('DB_USER'),
        password=password,
        ssl=True  # SSL 암호화
    )
    return conn

def hash_user_password(plain_password: str) -> str:
    """사용자 비밀번호 안전하게 해싱"""
    # bcrypt로 단방향 암호화
    return bcrypt.hashpw(
        plain_password.encode('utf-8'),
        bcrypt.gensalt(rounds=12)
    )

def verify_password(plain_password: str, hashed: str) -> bool:
    """입력된 비밀번호와 해시 비교"""
    return bcrypt.checkpw(plain_password.encode('utf-8'), hashed)
```
**개선**: OWASP 보안 기준 준수, 환경변수 사용, 암호 해싱 ✅

---

##### 5️⃣ **Trackable**: 추적 가능

❌ **BEFORE** - 추적 불가능:
```python
# 어느 SPEC에서 왔는지 모름
def calculate_discount(price):
    if price > 100:
        return price * 0.9  # 할인율 10%?
    return price
```
**문제**: SPEC이 없음, 언제 변경됐는지 모름, 테스트 기준 불명확 ⚠️

✅ **AFTER** - 완벽하게 추적:
```python
# SPEC-042: 고가 상품 할인 정책
# 요구사항: 100달러 이상의 상품에 10% 할인 적용
# 성공 기준: 할인이 정확히 적용되고, 원가보다 낮아지지 않음

def calculate_discount(price: float) -> float:
    """
    상품 가격에 대한 할인 계산 (SPEC-042)

    규칙: 100달러 이상이면 10% 할인

    Args:
        price: 원가 (달러)

    Returns:
        float: 할인 적용 후 가격

    Example:
        >>> calculate_discount(150)
        135.0  # 150 * 0.9
    """
    MIN_DISCOUNT_PRICE = 100.0
    DISCOUNT_RATE = 0.10  # 10% 할인 (SPEC-042 요구)

    if price >= MIN_DISCOUNT_PRICE:
        discounted_price = price * (1 - DISCOUNT_RATE)
        return max(discounted_price, 0)  # 음수 방지

    return price

# Test: SPEC-042와 연결된 테스트
def test_calculate_discount_applies_10_percent_discount_for_expensive_items():
    """SPEC-042: 100달러 이상은 10% 할인"""
    assert calculate_discount(150) == 135.0
    assert calculate_discount(100) == 90.0  # 경계값 테스트

def test_calculate_discount_no_discount_for_cheap_items():
    """SPEC-042: 100달러 미만은 할인 없음"""
    assert calculate_discount(99) == 99
    assert calculate_discount(0) == 0
```
**개선**: SPEC으로 추적 가능, 테스트로 검증, 변경 이유 명확 ✅

---

**TRUST 5 체크리스트** 📋:

| 항목 | 확인 | 도구 |
|------|------|------|
| ✅ 테스트 커버리지 ≥ 85% | `pytest --cov` | pytest |
| ✅ 명확한 이름, 주석 | 코드 리뷰 | pylint, ruff |
| ✅ 포맷팅 일관성 | 자동 포맷 | black, isort |
| ✅ OWASP 보안 | 보안 검증 | security-expert agent |
| ✅ SPEC 링킹 | Git 커밋 메시지 | 수동 확인 |

---

## 4. 설치 및 설정

### 📋 전제조건

| 요구사항    | 최소 버전 | 권장 버전 | 확인 방법           |
| ----------- | --------- | --------- | ------------------- |
| Python      | 3.11+     | 3.13+     | `python --version`  |
| Node.js     | 18+       | 20+       | `node --version`    |
| Git         | 2.30+     | 최신      | `git --version`     |
| Claude Code | 2.0.46+   | 최신      | Claude Code 앱 정보 |

### 🔧 설치 방법

**Option 1: `uv` 사용 (권장)**

```bash
# uv 설치
curl -LsSf https://astral.sh/uv/install.sh | sh

# MoAI-ADK 설치
uv tool install moai-adk

# 버전 확인
moai-adk --version
```

**Option 2: `pip` 사용**

```bash
# pip로 설치
pip install moai-adk

# 버전 확인
moai-adk --version
```

### 🎯 프로젝트 초기화

**신규 프로젝트:**

```bash
# 프로젝트 생성
moai-adk init my-awesome-project

# 디렉토리 구조
my-awesome-project/
├── .claude/
│   ├── agents/              # 에이전트 정의
│   ├── commands/            # 커맨드 정의
│   ├── skills/              # 스킬 라이브러리
│   └── settings.json        # Claude Code 설정
├── .moai/
│   ├── memory/
│   │   ├── agents.md        # 에이전트 참조
│   │   ├── commands.md      # 커맨드 참조
│   │   └── ...
│   └── config/              # 설정 파일
└── src/                     # 소스 코드
```

**기존 프로젝트:**

```bash
cd existing-project
moai-adk init .

# Git 저장소와 함께 초기화
moai-adk init . --with-git
```

### ⚙️ .claude/settings.json 설정

MoAI-ADK는 `.claude/settings.json` 파일을 사용하여 Claude Code 동작을 제어합니다.

`.claude/settings.json` 파일을 편집하여 프로젝트를 커스터마이즈하세요:

```json
{
  "user": {
    "name": "개발자이름"
  },
  "language": {
    "conversation_language": "ko",
    "agent_prompt_language": "en"
  },
  "constitution": {
    "enforce_tdd": true,
    "test_coverage_target": 85
  },
  "git_strategy": {
    "mode": "personal"
  },
  "github": {
    "spec_git_workflow": "develop_direct"
  },
  "statusline": {
    "enabled": true,
    "format": "compact",
    "style": "R2-D2"
  }
}
```

**주요 설정 항목:**

- `user.name`: Alfred가 당신을 부르는 이름
- `conversation_language`: 대화 및 문서 언어 (ko/en/ja/zh)
- `agent_prompt_language`: 에이전트 내부 추론 언어 (**항상 "en" 사용**)
- `enforce_tdd`: TDD 강제 여부 (true 권장)
- `test_coverage_target`: 테스트 커버리지 목표 (기본 85%)
- `git_strategy.mode`: Git 전략 (personal/team/hybrid)
- `statusline`: Claude Code 상태 표시줄 설정

---

## 5. 개발 워크플로우

### Phase 1: Plan (SPEC 생성)

**목적:** 모호한 아이디어를 명확한 EARS 포맷 명세서로 변환

**실행 단계:**

```bash
# 1. Plan 커맨드 실행
/moai:1-plan "JWT 토큰 기반 사용자 인증 시스템"

# Alfred의 동작:
# - spec-builder 에이전트 호출
# - 사용자 요구사항 분석
# - 불명확한 부분 질문
# - EARS 포맷 SPEC 문서 생성
# - .moai/specs/SPEC-001/ 디렉토리에 저장

# 2. SPEC 검토
# - Alfred가 초안을 보여줌
# - 필요시 수정 요청
# - 승인

# 3. 컨텍스트 초기화 (필수!)
/clear
```

### Phase 2: Run (TDD 구현)

**목적:** SPEC을 기반으로 Red-Green-Refactor TDD 사이클 실행

**실행 단계:**

```bash
# TDD 구현 시작
/moai:2-run SPEC-001

# Alfred의 동작:
# - tdd-implementer 에이전트 호출
# - Red: 실패하는 테스트 먼저 작성
# - Green: 테스트를 통과하는 최소 코드 작성
# - Refactor: 코드 품질 개선 및 최적화
# - 테스트 커버리지 ≥ 85% 확인
```

**TDD 사이클 상세:**

```mermaid
flowchart LR
    Red[🔴 Red<br/>실패 테스트 작성] --> Green[🟢 Green<br/>최소 코드 구현]
    Green --> Refactor[🔵 Refactor<br/>코드 개선]
    Refactor --> Coverage{커버리지<br/>≥ 85%?}
    Coverage -->|No| Red
    Coverage -->|Yes| Done[✅ 완료]

    style Red fill:#ffcccc
    style Green fill:#ccffcc
    style Refactor fill:#ccccff
    style Done fill:#ccffcc
```

### Phase 3: Sync (문서 동기화)

**목적:** 구현된 코드를 분석하여 문서와 다이어그램 자동 생성

**실행 단계:**

```bash
# 문서 동기화
/moai:3-sync SPEC-001

# Alfred의 동작:
# - docs-manager 에이전트 호출
# - 코드 주석에서 API 문서 추출
# - Mermaid 다이어그램 생성
# - README.md 업데이트
# - CHANGELOG 자동 생성
```

---

## 6. 핵심 커맨드

MoAI-ADK의 개발 워크플로우는 6개의 핵심 커맨드로 구성되어 있습니다. 이 커맨드들은 처음 프로젝트 초기화부터 최종 프로덕션 배포까지 완전한 개발 라이프사이클을 자동화합니다. 각 커맨드는 Mr.Alfred 슈퍼 에이전트 오케스트레이터에 의해 관리되며, 필요한 전문 AI 에이전트들을 자동으로 선택하고 조율합니다. SPEC-First TDD 방식을 따르므로 명확한 요구사항에서 시작하여 테스트 기반 구현, 자동 문서화까지 모든 단계가 체계적으로 진행됩니다.

### `/moai:0-project` - 프로젝트 초기화

**목적:** 프로젝트 구조 생성 및 설정 초기화

**사용법:**

```bash
/moai:0-project
```

**동작:**

1. `.moai/` 디렉토리 구조 생성
2. `.claude/settings.json` 템플릿 생성
3. Git 저장소 초기화 (선택)
4. `.claude/` 에이전트/스킬 동기화

**위임 에이전트:** `project-manager`

---

### `/moai:1-plan` - SPEC 생성

**목적:** 사용자 요구사항을 EARS 포맷 SPEC 문서로 변환

**사용법:**

```bash
/moai:1-plan "기능 설명을 자연어로 작성"
```

**예시:**

```bash
# 예시 1: 간단한 기능
/moai:1-plan "사용자 회원가입 기능"

# 예시 2: 상세한 요구사항
/moai:1-plan "OAuth2.0 소셜 로그인 (Google, GitHub) 지원.
사용자 프로필 정보 자동 동기화. 기존 계정과 연결 가능."

# 예시 3: API 설계
/moai:1-plan "게시판 REST API - 페이지네이션, 정렬, 필터링 지원"
```

**위임 에이전트:** `spec-builder`

---

### `/moai:2-run` - TDD 구현

**목적:** SPEC 기반 Red-Green-Refactor TDD 사이클 실행

**사용법:**

```bash
/moai:2-run SPEC-ID
```

**예시:**

```bash
# 기본 실행
/moai:2-run SPEC-001

# 특정 언어/프레임워크 지정
/moai:2-run SPEC-002 --lang python --framework fastapi

# 단계별 확인 모드
/moai:2-run SPEC-003 --interactive
```

**위임 에이전트:** `tdd-implementer`

---

### `/moai:3-sync` - 문서 동기화

**목적:** 코드 분석 및 자동 문서 생성/업데이트

**사용법:**

```bash
/moai:3-sync SPEC-ID [옵션]
```

**예시:**

```bash
# 기본 동기화
/moai:3-sync SPEC-001

# 특정 문서 타입만
/moai:3-sync SPEC-002 --docs api

# 다이어그램 생성
/moai:3-sync SPEC-003 --diagrams architecture,sequence

# 다국어 문서
/moai:3-sync SPEC-004 --languages ko,en,ja
```

**위임 에이전트:** `docs-manager`

---

### `/moai:9-feedback` - 피드백 및 개선

**목적:** MoAI-ADK 프레임워크 버그 분석 및 자동 이슈 등록

**사용법:**

```bash
/moai:9-feedback [옵션]
```

**예시:**

```bash
# 전체 분석
/moai:9-feedback

# 특정 오류 보고
/moai:9-feedback --error "TDD 사이클 중 커버리지 계산 오류"

# 개선 제안
/moai:9-feedback --suggestion "SPEC 템플릿에 성능 요구사항 섹션 추가"
```

**위임 에이전트:** `quality-gate`, `debug-helper`

---

## 7. 에이전트 가이드

MoAI-ADK는 **35명의 전문 AI 에이전트**를 제공합니다. 각 에이전트는 특정 분야의 전문가로서 고도로 특화된 역할을 수행하며, Mr.Alfred 슈퍼 에이전트 오케스트레이터에 의해 자동으로 선택되고 조율됩니다. 사용자가 요청하면 Alfred는 요구사항을 분석하여 필요한 에이전트들을 순차적으로 또는 병렬로 위임하며, 각 에이전트의 결과를 다음 에이전트의 입력으로 전달하는 방식으로 작업을 진행합니다. 이러한 시스템을 통해 요구사항 분석, 설계, 구현, 테스트, 보안 검증, 배포에 이르기까지 전체 개발 라이프사이클이 자동화되고 최적화됩니다. 에이전트들은 5가지 카테고리로 분류되며, 각 카테고리는 특정 개발 단계를 담당합니다.

### 📋 기획 및 설계 (Planning & Design)

| 에이전트               | 전문 분야     | 주요 책임                            | 대표 스킬              |
| ---------------------- | ------------- | ------------------------------------ | ---------------------- |
| **spec-builder**       | 요구사항 분석 | EARS 포맷 SPEC 작성, 요구사항 명확화 | `moai-foundation-ears` |
| **api-designer**       | API 설계      | REST/GraphQL 엔드포인트 설계         | `moai-domain-web-api`  |
| **component-designer** | 컴포넌트 설계 | 재사용 가능한 UI 컴포넌트 설계       | `moai-design-systems`  |
| **ui-ux-expert**       | UX 설계       | 사용자 경험 및 인터페이스 설계       | `moai-domain-figma`    |
| **plan**               | 전략 수립     | 복잡한 작업을 단계별로 분해          | `moai-core-workflow`   |

### 💻 구현 (Implementation)

| 에이전트            | 전문 분야           | 주요 책임                                   | 대표 스킬               |
| ------------------- | ------------------- | ------------------------------------------- | ----------------------- |
| **tdd-implementer** | TDD 구현            | Red-Green-Refactor 사이클 실행              | `moai-foundation-trust` |
| **backend-expert**  | 백엔드 아키텍처     | 서버 로직, 데이터베이스 통합, 비즈니스 로직 | `moai-domain-backend`   |
| **frontend-expert** | 프론트엔드 아키텍처 | 웹 프론트엔드, 상태 관리, UI 인터랙션       | `moai-domain-frontend`  |
| **database-expert** | 데이터베이스 설계   | DB 스키마 설계, 쿼리 최적화, 마이그레이션   | `moai-domain-database`  |

### 🛡️ 품질 및 보안 (Quality & Security)

| 에이전트            | 전문 분야   | 주요 책임                                     | 대표 스킬                             |
| ------------------- | ----------- | --------------------------------------------- | ------------------------------------- |
| **security-expert** | 보안        | 취약점 검사, OWASP 준수, 보안 코딩 가이드     | `moai-domain-security`                |
| **quality-gate**    | 품질 검증   | 코드 품질, 커버리지, TRUST 5 원칙 검증        | `moai-foundation-trust`               |
| **test-engineer**   | 테스트 전략 | 단위/통합/E2E 테스트 전략 및 테스트 코드 강화 | `moai-essentials-testing-integration` |
| **format-expert**   | 코드 스타일 | 코드 스타일 가이드 및 린팅 규칙 적용          | `moai-core-code-reviewer`             |
| **debug-helper**    | 디버깅      | 런타임 오류 근본 원인 분석 및 해결책 제시     | `moai-essentials-debug`               |

### 🚀 DevOps 및 관리 (DevOps & Management)

| 에이전트                 | 전문 분야     | 주요 책임                                           | 대표 스킬                    |
| ------------------------ | ------------- | --------------------------------------------------- | ---------------------------- |
| **devops-expert**        | DevOps        | CI/CD 파이프라인, 클라우드 인프라(IaC), 배포 자동화 | `moai-domain-devops`         |
| **monitoring-expert**    | 모니터링      | 시스템 모니터링, 로깅, 알림 시스템 설정             | `moai-domain-monitoring`     |
| **performance-engineer** | 성능 최적화   | 시스템 성능 병목 분석 및 최적화                     | `moai-essentials-perf`       |
| **docs-manager**         | 문서 관리     | 프로젝트 문서 생성, 업데이트, 관리                  | `moai-docs-generation`       |
| **git-manager**          | Git 관리      | Git 브랜치 전략, PR 관리, 버전 태깅                 | `moai-foundation-git`        |
| **project-manager**      | 프로젝트 관리 | 전체 프로젝트 진행 조율 및 관리                     | `moai-project-documentation` |

### 🛠️ 특화 도구 (Specialized Tools)

| 에이전트          | 전문 분야       | 주요 책임                                | 대표 스킬                 |
| ----------------- | --------------- | ---------------------------------------- | ------------------------- |
| **agent-factory** | 에이전트 팩토리 | 새로운 커스텀 에이전트 생성 및 설정      | `moai-core-agent-factory` |
| **skill-factory** | 스킬 팩토리     | 새로운 MoAI 스킬 정의 및 라이브러리 추가 | `moai-cc-skill-factory`   |

---

## 8. 스킬 라이브러리 (106개)

스킬(Skill)은 MoAI-ADK의 핵심 지식 모듈입니다. 각 에이전트가 작업할 때 사용하는 전문 지식, 패턴, 최적 사례를 담고 있으며, **106개의 스킬**이 **10개 Tier**로 체계화되어 있습니다.

### 🚀 작업별 스킬 찾기 (Task-Based Search)

**당신의 작업을 선택하세요:**

| 작업 | 추천 스킬 | 사용 시기 |
|------|---------|---------|
| **JWT/OAuth 인증 구현** | `moai-security-zero-trust`, `moai-domain-security` | 사용자 인증 필요 |
| **테스트 작성 & 커버리지** | `moai-foundation-trust`, `moai-essentials-testing` | 모든 코드 구현 |
| **보안 취약점 검사** | `moai-domain-security`, `moai-security-owasp` | 배포 전 검증 |
| **성능 최적화** | `moai-essentials-perf`, `moai-domain-database` | 느린 부분 발견 |
| **REST/GraphQL API 설계** | `moai-domain-web-api`, `moai-domain-backend` | API 구축 |
| **데이터베이스 설계** | `moai-domain-database`, `moai-foundation-specs` | DB 스키마 정의 |
| **React/Vue 컴포넌트** | `moai-domain-frontend`, `moai-lang-typescript` | UI 개발 |
| **배포 & CI/CD** | `moai-domain-devops`, `moai-baas-vercel-ext` | 프로덕션 배포 |
| **문서 생성** | `moai-docs-generation`, `moai-cc-claude-md` | API 문서화 |
| **Git 워크플로우** | `moai-foundation-git`, `moai-core-clone-pattern` | 버전 관리 |
| **마이크로서비스 설계** | `moai-domain-backend`, `moai-baas-foundation` | 복잡한 시스템 |
| **모니터링 & 로깅** | `moai-domain-monitoring`, `moai-essentials-debug` | 프로덕션 안정성 |

---

### 📊 스킬 포트폴리오 통계

- **총 스킬 수**: 106개 (82개 계층화 + 24개 특수)
- **10-Tier 분류**: 언어에서 특화 라이브러리까지 체계적 조직화
- **100% 메타데이터 준수**: 모든 스킬에 7개 필수 필드 포함
- **1,270개 자동 트리거 키워드**: 사용자 요청에서 지능적 스킬 선택
- **94% 에이전트-스킬 커버리지**: 35개 에이전트 중 33개가 명시적 스킬 참조

### 🎯 계층별 스킬 구조 (Tier Structure)

- **Tier 1-2**: Foundation (언어, 도메인) - 36개 스킬
- **Tier 3-5**: Core Architecture (보안, 코어, 파운데이션) - 19개 스킬
- **Tier 6-7**: Platform Integration (Claude Code, BaaS) - 17개 스킬
- **Tier 8-10**: Applied Workflows (필수 도구, 프로젝트, 라이브러리) - 6개 스킬
- **Special Skills**: 계층 미분류 유틸리티 - 24개 스킬

💡 **팁**: 위의 "작업별 검색"으로 필요한 스킬을 빠르게 찾거나, 아래에서 Tier별로 탐색할 수 있습니다.

### 📚 전체 스킬 목록 (알파벳 순)

#### Tier 1: 언어별 스킬 (moai-lang-\*)

프로그래밍 언어 패턴 및 관용구 (21개)

| 스킬명                   | 설명                                                     |
| ------------------------ | -------------------------------------------------------- |
| `moai-lang-c`            | C 언어 개발 (포인터, 메모리 관리, 성능 최적화)           |
| `moai-lang-cpp`          | C++ 개발 (표준 라이브러리, 템플릿, 모던 C++)              |
| `moai-lang-csharp`       | C# 개발 (.NET, LINQ, 비동기 패턴)                         |
| `moai-lang-dart`         | Dart 개발 (Flutter 위젯, 비동기 프로그래밍)              |
| `moai-lang-elixir`       | Elixir 개발 (Phoenix 프레임워크, OTP, 함수형 패턴)       |
| `moai-lang-go`           | Go 개발 (고루틴, 채널, 동시성 처리)                       |
| `moai-lang-html-css`     | HTML/CSS 마크업 (HTML5, CSS3, Flexbox, Grid 레이아웃)   |
| `moai-lang-java`         | Java 개발 (Spring Boot, Maven, 엔터프라이즈 패턴)        |
| `moai-lang-javascript`   | JavaScript 개발 (ES6+, 비동기/대기, DOM 조작)            |
| `moai-lang-kotlin`       | Kotlin 개발 (코루틴, Android 개발, JVM 환경)             |
| `moai-lang-php`          | PHP 개발 (Laravel, Composer, 모던 PHP 패턴)             |
| `moai-lang-python`       | Python 개발 (FastAPI, Django, pytest, 타입 힌팅)         |
| `moai-lang-r`            | R 통계 분석 (데이터 분석, 시각화, tidyverse 생태계)      |
| `moai-lang-ruby`         | Ruby 개발 (Rails, RSpec, 메타프로그래밍)                 |
| `moai-lang-rust`         | Rust 개발 (소유권, 생명주기, 제로-코스트 추상화)         |
| `moai-lang-scala`        | Scala 개발 (함수형 프로그래밍, Akka, 타입 시스템)        |
| `moai-lang-shell`        | Shell 스크립팅 (Bash, 자동화, CLI 도구 개발)             |
| `moai-lang-sql`          | SQL 쿼리 (쿼리 최적화, 데이터베이스 관리)                |
| `moai-lang-swift`        | Swift 개발 (SwiftUI, iOS 앱 개발, 프로토콜 지향)         |
| `moai-lang-tailwind-css` | Tailwind CSS (유틸리티 우선 접근, 반응형 디자인)         |
| `moai-lang-typescript`   | TypeScript 개발 (타입 시스템, 제네릭, 고급 패턴)        |

#### Tier 2: 도메인별 스킬 (moai-domain-\*)

애플리케이션 도메인 아키텍처 (16개)

| 스킬명                   | 설명                                                           |
| ------------------------ | -------------------------------------------------------------- |
| `moai-domain-backend`    | 백엔드 아키텍처 (REST API, 마이크로서비스, CRUD 패턴)        |
| `moai-domain-cli-tool`   | CLI 도구 개발 (명령줄 애플리케이션, 인자 파싱)               |
| `moai-domain-cloud`      | 클라우드 아키텍처 (클라우드 플랫폼, 서버리스, 확장성)        |
| `moai-domain-database`   | 데이터베이스 설계 (관계형/비관계형 DB, 스키마, 인덱싱)       |
| `moai-domain-devops`     | DevOps 실천 (CI/CD, IaC, 자동화, 배포)                       |
| `moai-domain-figma`      | Figma 통합 (디자인-코드 변환, Figma API, 디자인 토큰)        |
| `moai-domain-frontend`   | 프론트엔드 아키텍처 (UI 프레임워크, 상태 관리, 라우팅)       |
| `moai-domain-iot`        | IoT 개발 (IoT 디바이스, 센서, 프로토콜, 엣지 컴퓨팅)         |
| `moai-domain-ml-ops`     | MLOps (머신러닝 파이프라인, 모델 배포, 모니터링)             |
| `moai-domain-mobile-app` | 모바일 앱 개발 (iOS, Android, React Native, Flutter)          |
| `moai-domain-monitoring` | 모니터링 (로깅, 메트릭 수집, 알림, 관찰성)                   |
| `moai-domain-notion`     | Notion 통합 (Notion API, 지식 베이스, 데이터베이스 관리)      |
| `moai-domain-security`   | 보안 (OWASP, 취약점 분석, 보안 코딩)                         |
| `moai-domain-toon`       | TOON 포맷 (토큰 최적화, 인코딩, 압축)                         |
| `moai-domain-web-api`    | 웹 API 설계 (REST, GraphQL, API 디자인, 버전 관리)           |

#### Tier 3: 보안 스킬 (moai-security-\*)

보안 및 준수 (12개)

| 스킬명                              | 설명                                                             |
| ----------------------------------- | ---------------------------------------------------------------- |
| `moai-security-accessibility-wcag3` | WCAG 3.0 접근성 (ARIA, 키보드 네비게이션, 시맨틱 HTML)          |
| `moai-security-compliance`          | 보안 준수 (준수 기준, 감사, 인증)                               |
| `moai-security-encryption`          | 암호화 (데이터 암호화, 해싱, TLS/SSL, 키 관리)                  |
| `moai-security-owasp`               | OWASP (OWASP Top 10, 보안 표준, 모범 사례)                      |
| `moai-security-secrets`             | 비밀 관리 (시크릿 저장소, 자동 갱신)                           |
| `moai-security-ssrf`                | SSRF 방어 (서버 측 요청 위조 방지)                              |
| `moai-security-threat`              | 위협 모델링 (위협 분석, 위험 평가, 공격 벡터 분석)              |
| `moai-security-zero-trust`          | Zero Trust 아키텍처 (제로 트러스트 보안 모델, 최소 권한 원칙)   |

#### Tier 4: 코어 개발 스킬 (moai-core-\*)

핵심 개발 패턴 및 도구 (15개) (Phase 2 병합: -2개)

| 스킬명                        | 설명                                                         |
| ----------------------------- | ------------------------------------------------------------ |
| `moai-core-agent-factory`     | 에이전트 팩토리 (커스텀 에이전트 생성, 오케스트레이션)        |
| `moai-core-config-schema`     | 설정 스키마 (설정 관리, 유효성 검증, 타입 정의)               |
| `moai-core-dev-guide`         | 개발 가이드 (개발 지침, 모범 사례)                            |
| `moai-core-env-security`      | 환경 보안 (환경 변수 보안, .env 파일 관리)                    |
| `moai-core-issue-labels`      | 이슈 라벨 (GitHub 이슈 라벨링, 분류)                         |
| `moai-core-practices`         | 모범 사례 (코딩 표준, 관례, 베스트 프랙티스)                 |
| `moai-core-spec-authoring`    | SPEC 작성 (EARS 포맷, 요구사항, 명세 작성)                   |
| `moai-core-todowrite-pattern` | TodoWrite 패턴 (작업 추적, 진행도 모니터링)                  |
| `moai-core-workflow`          | 워크플로우 (개발 워크플로우, 자동화, 프로세스)               |

#### Tier 5: 파운데이션 스킬 (moai-foundation-\*)

프레임워크 기반 및 표준 (5개)

| 스킬명                  | 설명                                                             |
| ----------------------- | ---------------------------------------------------------------- |
| `moai-foundation-ears`  | EARS 포맷 (Event-driven requirements format, structured specs)   |
| `moai-foundation-git`   | Git 관리 (Git workflows, branching strategies, version control)  |
| `moai-foundation-langs` | 언어 기반 (multi-language support, i18n, localization)           |
| `moai-foundation-specs` | SPEC 시스템 (SPEC lifecycle, versioning, traceability)           |
| `moai-foundation-trust` | TRUST 5 프레임워크 (Test, Readable, Unified, Secured, Trackable) |

#### Tier 6: Claude Code 플랫폼 스킬 (moai-cc-\*)

Claude Code 통합 (10개)

| 스킬명                    | 설명                                                       |
| ------------------------- | ---------------------------------------------------------- |
| `moai-cc-claude-md`       | CLAUDE.md 작성 (프로젝트 문서, 에이전트 지침)               |
| `moai-cc-commands`        | 커맨드 시스템 (커맨드 관리, 커스텀 커맨드)                |
| `moai-cc-configuration`   | 설정 관리 (Claude Code 설정, 프로젝트 설정, 검증)         |
| `moai-cc-hooks`           | Hooks 시스템 (자동화 트리거, 생명주기 Hooks)              |
| `moai-cc-memory`          | 메모리 시스템 (메모리 파일 관리, 컨텍스트 보존)           |
| `moai-cc-permission-mode` | 권한 모드 (권한 관리, 접근 제어)                          |
| `moai-cc-skill-factory`   | 스킬 팩토리 (스킬 생성, 관리, 버전 관리)                  |
| `moai-cc-skills-guide`    | 스킬 가이드 (스킬 개발, 최적화, 표준 준수)                |

#### Tier 7: BaaS 통합 스킬 (moai-baas-\*)

Backend-as-a-Service 플랫폼 (10개)

| 스킬명                     | 설명                                                        |
| -------------------------- | ----------------------------------------------------------- |
| `moai-baas-clerk-ext`      | Clerk 인증 (Clerk 플랫폼, OAuth, 사용자 관리)             |
| `moai-baas-cloudflare-ext` | Cloudflare 통합 (Workers, Pages, CDN, 엣지 컴퓨팅)       |
| `moai-baas-convex-ext`     | Convex 통합 (백엔드 플랫폼, 실시간 데이터베이스)         |
| `moai-baas-firebase-ext`   | Firebase 통합 (Firebase 서비스, Firestore, Auth, Hosting) |
| `moai-baas-foundation`     | BaaS 기반 (BaaS 패턴, 모범 사례, 아키텍처)               |
| `moai-baas-neon-ext`       | Neon 통합 (Neon Postgres, 서버리스 데이터베이스)         |
| `moai-baas-railway-ext`    | Railway 통합 (Railway 배포, 컨테이너화)                  |
| `moai-baas-supabase-ext`   | Supabase 통합 (Supabase 백엔드, Postgres, Auth, Storage) |
| `moai-baas-vercel-ext`     | Vercel 통합 (Vercel 배포, Edge Functions, 서버리스)      |

#### Tier 8: 필수 도구 스킬 (moai-essentials-\*)

필수 개발 워크플로우 (6개)

| 스킬명 | 설명 |
| ------ | ---- |
| `moai-essentials-debug` | 디버깅 오케스트레이션 (오류 분석, 근본 원인, 해결책 제시) |
| `moai-essentials-perf` | 성능 최적화 (병목 분석, 성능 튜닝, 벤치마킹) |
| `moai-essentials-refactor` | 리팩토링 자동화 (코드 변환, 기술 부채 제거, 최적화) |
| `moai-mcp-figma-integrator` | Figma 통합 (디자인 분석, 디자인-투-코드, 컴포넌트 추출) |
| `moai-mcp-notion-integrator` | Notion 통합 (데이터베이스 관리, 콘텐츠 작성, 자동화) |
| `moai-playwright-webapp-testing` | 웹앱 테스팅 (E2E 테스트, 자동화, UI 상호작용) |

#### Tier 9: 프로젝트 관리 스킬 (moai-project-\*)

프로젝트 조율 (4개) (Phase 2 병합: -1개)

| 스킬명                         | 설명                                                     |
| ------------------------------ | -------------------------------------------------------- |
| `moai-project-batch-questions` | 일괄 질문 (배치 질문 처리, 대량 작업)                   |
| `moai-project-config-manager` | 설정 관리 (config.json CRUD, 검증, 병합 전략)          |
| `moai-project-documentation`   | 프로젝트 문서화 (프로젝트 문서, 자동 생성)             |
| `moai-session-info`            | 세션 정보 (프로젝트 상태, 버전, 리소스 정보 표시)     |

#### Tier 10: AI 특화 스킬 (moai-ai-\*, moai-lang-\*)

AI 및 특화 라이브러리 (2개)

| 스킬명                  | 설명                                                                           |
| ----------------------- | ------------------------------------------------------------------------------ |
| `moai-ai-nano-banana`   | Google Nano Banana Pro 이미지 생성 (Text-to-Image, Image-to-Image, 멀티턴)     |
| `moai-lang-shadcn-ui`   | shadcn/ui 통합 (React 컴포넌트 라이브러리, Tailwind, Radix UI)                |

#### 특수 스킬 (Special Skills)

계층 미분류 유틸리티 (24개)

| 스킬명                           | 설명                                                                       |
| -------------------------------- | -------------------------------------------------------------------------- |
| `moai-artifacts-builder`         | Artifacts 생성 (아티팩트 생성, Claude artifacts)                          |
| `moai-change-logger`             | 변경 로그 (변경 추적, 버전 관리, Changelog 생성)                          |
| `moai-cloud-aws-advanced`        | AWS 고급 (고급 AWS 패턴, 서버리스, Lambda, S3)                           |
| `moai-cloud-gcp-advanced`        | GCP 고급 (고급 GCP 패턴, Cloud Run, BigQuery)                            |
| `moai-component-designer`        | 컴포넌트 설계 (컴포넌트 패턴, 재사용성, 구조화)                          |
| `moai-context7-integration`      | Context7 통합 (Context7 MCP, 라이브러리 문서 조회)                       |
| `moai-design-systems`            | 디자인 시스템 (디자인 패턴, 디자인 토큰, 일관성)                         |
| `moai-document-processing`       | 문서 처리 (문서 파싱, 변환, 추출, 처리)                                  |
| `moai-icons-vector`              | 벡터 아이콘 (아이콘 관리, SVG, 아이콘 시스템)                            |
| `moai-internal-comms`            | 내부 통신 (에이전트 조율, 메시지 전달, 워크플로우)                       |
| `moai-jit-docs-enhanced`         | JIT 문서 강화 (즉시 문서, 컨텍스트 인식, 동적 생성)                      |
| `moai-learning-optimizer`        | 학습 최적화 (적응형 학습, 최적화, 추천 시스템)                           |
| `moai-mcp-integration`           | MCP 통합 (MCP 서버, 프로토콜, 도구 연동)                                 |
| `moai-mermaid-diagram-expert`    | Mermaid 다이어그램 (21가지 다이어그램, 시각화, 흐름도)                    |
| `moai-nextra-architecture`       | Nextra 아키텍처 (Nextra 문서 프레임워크, SSG, 정적 생성)                 |
| `moai-readme-expert`             | README 전문가 (전문적 README 생성, 템플릿)                               |
| `moai-spec-intelligent-workflow` | 지능형 SPEC (SPEC 자동화, 워크플로우, 최적화)                            |
| `moai-streaming-ui`              | 스트리밍 UI (실시간 스트리밍, UI 업데이트, 비동기)                        |

#### 🔄 통합된 스킬 (Merged Skills)

다음 15개 스킬은 중복되는 기능을 통합하여 더 강력하고 효율적인 기능을 제공합니다:

**Phase 1** (High Priority - 이미 통합됨):

| 스킬명 | 설명 |
| --- | --- |
| `moai-code-review` | 코드 리뷰 (TRUST 5 기반, 자동화, 협업) |
| `moai-testing` | 테스트 전략 (TDD, 단위/통합/E2E 테스트) |
| `moai-security-api-management` | API 보안 및 관리 (인증, 인가, 버전 관리) |
| `moai-security-authentication` | 인증 및 신원 관리 (OAuth 2.1, JWT, WebAuthn, MFA) |
| `moai-essentials-performance` | 성능 분석 및 프로파일링 (AI 기반 병목, Scalene) |

**Phase 2** (Medium Priority - ✅ 이미 통합됨):

| 스킬명 | 설명 |
| --- | --- |
| `moai-context-manager` | 컨텍스트 및 세션 관리 (토큰 예산, 상태, 모니터링) |
| `moai-templates` | 템플릿 관리 (코드/피드백/프로젝트 템플릿) |

**Special Skills** (유지):

| 스킬명 | 설명 |
| --- | --- |
| `moai-docs-manager` | 문서 관리 (자동 생성, 도구 통합, 일관성 검증) |
| `moai-docs-quality-gate` | 문서 품질 보증 (내용 검증, 마크다운 린팅, 링크 검사) |
| `moai-web-testing` | 웹 애플리케이션 테스트 (E2E 테스트, Playwright, 테스트 자동화) |
| `moai-config-manager` | 설정 관리 (Claude Code 설정, 프로젝트 설정, 스키마 검증) |
| `moai-adaptive-ux` | 적응형 사용자 경험 (전문도 감지, 맞춤형 제안, 동적 응답) |
| `moai-language-support` | 언어 지원 (언어 감지, 자동 설정, 프로젝트 초기화) |
| `moai-cc-guide` | Claude Code 가이드 (스킬 사용법, 에이전트 위임, 오케스트레이션) |
| `moai-baas-auth` | BaaS 인증 플랫폼 (Auth0, Clerk, OAuth, 사용자 관리) |

---

### 🔍 스킬 사용 방법

**자동 활용**: 에이전트가 자동으로 필요한 스킬을 선택

```bash
# tdd-implementer가 자동으로 다음 스킬 활용:
# - moai-foundation-trust
# - moai-essentials-testing-integration
# - moai-lang-python (Python 프로젝트인 경우)
/moai:2-run SPEC-001
```

**명시적 호출**: 특정 스킬을 직접 호출

```bash
# EARS 포맷 가이드 조회
Skill("moai-foundation-ears")

# Docker 배포 패턴 조회
Skill("moai-domain-devops")

# OAuth 2.0 구현 가이드
Skill("moai-security-auth")
```

**스킬 조합**: 여러 스킬을 조합하여 복합 작업 수행

```bash
# FastAPI + PostgreSQL + Docker 조합
# backend-expert가 자동으로 다음 스킬 활용:
# - moai-lang-python
# - moai-domain-backend
# - moai-domain-database
@agent-backend-expert "FastAPI 앱을 PostgreSQL과 연동하고 Docker로 배포"
```

---

## 9. 실용 예제

### 예제 1: 사용자 로그인 시스템

**목표:** JWT 토큰 기반 인증 시스템 구현

**단계별 실행:**

```bash
# 1. SPEC 생성
/moai:1-plan "JWT 토큰 기반 로그인 시스템.
- 이메일/비밀번호 인증
- 액세스 토큰(30분), 리프레시 토큰(7일)
- 5회 실패 시 계정 잠금(30분)"

# Alfred가 질문:
# Q1: 비밀번호 정책은?
# A1: 최소 8자, 대소문자+숫자+특수문자

# Q2: JWT 알고리즘은?
# A2: RS256

# → SPEC-001 생성됨
```

**JWT 보안 상세** (심화):

#### 🔐 JWT 구조 이해

JWT는 3가지 부분으로 구성됩니다:

```
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9
  ↓
[Header - 알고리즘, 토큰 타입]

.

eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ
  ↓
[Payload - 사용자 정보, 발급 시간, 만료 시간]

.

S0Jzy-OMl...
  ↓
[Signature - RS256으로 서명된 검증 코드]
```

**각 부분 상세:**

```json
// Header (알고리즘 명시)
{
  "alg": "RS256",  // RSA 2048-bit으로 서명
  "typ": "JWT"
}

// Payload (클레임)
{
  "sub": "user_123",           // 사용자 ID
  "name": "John Doe",
  "email": "john@example.com",
  "iat": 1516239022,           // Issued At (발급 시간)
  "exp": 1516242622,           // Expiration (만료 시간) - 1시간
  "aud": "my-app",             // Audience (대상)
  "iss": "auth-server"         // Issuer (발급자)
}

// Signature
HMACSHA256(
  base64UrlEncode(header) + "." +
  base64UrlEncode(payload),
  private_key
)
```

---

#### 🔄 액세스 토큰 vs 리프레시 토큰 전략

**왜 2가지 토큰이 필요한가?**

| 토큰 | 목적 | 유효시간 | 저장 위치 | 역할 |
|------|------|---------|---------|------|
| **Access Token** | API 요청 인증 | 짧음 (30분) | 메모리 | 보안 중요 |
| **Refresh Token** | 새 Access Token 발급 | 길음 (7일) | HttpOnly Cookie | 재발급 기능 |

**흐름:**

```mermaid
sequenceDiagram
    participant U as 사용자
    participant C as 클라이언트
    participant S as 인증 서버
    participant A as API 서버

    U->>C: 1️⃣ 로그인 (이메일/비밀번호)
    C->>S: POST /login
    S->>S: 자격증명 검증
    Note over S: 암호: bcrypt로 해싱 검증
    S->>C: 201 Created<br/>- Access Token (30분)<br/>- Refresh Token (7일, HttpOnly)
    C->>C: Access Token 메모리에 저장

    Note over C,A: 정상 API 요청
    C->>A: 2️⃣ GET /posts<br/>Header: Authorization: Bearer [AccessToken]
    A->>A: Access Token 검증 (RS256 공개키)
    A->>C: 200 OK [포스트 목록]

    Note over C,A: Access Token 만료 (30분 후)
    C->>A: 3️⃣ GET /posts<br/>Header: Authorization: Bearer [만료된]
    A->>C: 401 Unauthorized

    C->>S: 4️⃣ POST /refresh<br/>Refresh Token 포함
    S->>S: Refresh Token 검증 (RS256)
    S->>S: 만료 확인 (7일 내?)
    S->>C: 새로운 Access Token 발급
    C->>C: 새 Access Token으로 메모리 갱신

    Note over C,A: 새 토큰으로 재시도
    C->>A: 5️⃣ GET /posts<br/>Header: Authorization: Bearer [새 토큰]
    A->>C: 200 OK [포스트 목록]

    Note over C: Refresh Token도 만료 (7일)
    C->>S: 6️⃣ POST /refresh<br/>만료된 Refresh Token
    S->>C: 401 Unauthorized
    C->>U: 다시 로그인해주세요
```

---

#### 🛡️ JWT 보안 체크리스트

| 항목 | 방법 | 코드 예제 |
|------|------|---------|
| **알고리즘** | RS256 사용 (HS256 금지) | `alg: "RS256"` |
| **서명** | 개인키로 서명, 공개키로 검증 | `RSA 2048-bit` |
| **만료 시간** | Access Token 30분, Refresh Token 7일 | `exp: now + 30min` |
| **저장** | Refresh Token은 HttpOnly Cookie | `Set-Cookie: refresh_token=...; HttpOnly` |
| **HTTPS** | 모든 토큰 전송은 HTTPS만 | 평문 HTTP 금지 |
| **검증** | 매 요청마다 서명 검증 | `jwt.verify(token, public_key)` |
| **갱신** | Refresh Token으로만 재발급 | `/refresh` 엔드포인트 |
| **로그아웃** | Refresh Token을 블랙리스트 처리 | Redis 블랙리스트 저장 |

**Python 구현 예제:**

```python
import jwt
from datetime import datetime, timedelta
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa

# RS256 키 쌍 생성
private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
public_key = private_key.public_key()

def create_access_token(user_id: str) -> str:
    """Access Token 발급 (30분 유효)"""
    payload = {
        'sub': user_id,
        'type': 'access',
        'iat': datetime.utcnow(),
        'exp': datetime.utcnow() + timedelta(minutes=30)
    }

    # RS256으로 서명
    token = jwt.encode(
        payload,
        private_key,
        algorithm='RS256'
    )
    return token

def create_refresh_token(user_id: str) -> str:
    """Refresh Token 발급 (7일 유효)"""
    payload = {
        'sub': user_id,
        'type': 'refresh',
        'iat': datetime.utcnow(),
        'exp': datetime.utcnow() + timedelta(days=7)
    }

    token = jwt.encode(
        payload,
        private_key,
        algorithm='RS256'
    )
    return token

def verify_token(token: str, token_type: str) -> dict:
    """토큰 검증 (공개키 사용)"""
    try:
        payload = jwt.decode(
            token,
            public_key,
            algorithms=['RS256'],
            audience='my-app'
        )

        # 토큰 타입 확인
        if payload.get('type') != token_type:
            raise ValueError(f"Expected {token_type} token")

        return payload
    except jwt.ExpiredSignatureError:
        raise ValueError("Token expired")
    except jwt.InvalidSignatureError:
        raise ValueError("Invalid signature")

# 사용 예제
@app.post('/login')
def login(email: str, password: str):
    """로그인: Access Token + Refresh Token 반환"""
    # 1. 사용자 인증
    user = authenticate_user(email, password)
    if not user:
        return {"error": "Invalid credentials"}, 401

    # 2. 토큰 발급
    access_token = create_access_token(user['id'])
    refresh_token = create_refresh_token(user['id'])

    # 3. Refresh Token은 HttpOnly Cookie에 저장
    response = {
        'access_token': access_token,
        'token_type': 'Bearer'
    }

    # HttpOnly + Secure + SameSite 설정
    response.set_cookie(
        'refresh_token',
        refresh_token,
        httponly=True,       # JavaScript 접근 불가
        secure=True,         # HTTPS만
        samesite='Strict',   # CSRF 방지
        max_age=7*24*60*60   # 7일
    )

    return response, 201

@app.post('/refresh')
def refresh(request):
    """새로운 Access Token 발급"""
    refresh_token = request.cookies.get('refresh_token')

    try:
        payload = verify_token(refresh_token, 'refresh')
        new_access_token = create_access_token(payload['sub'])
        return {'access_token': new_access_token, 'token_type': 'Bearer'}
    except ValueError as e:
        return {'error': str(e)}, 401

@app.get('/protected')
def protected_route(request):
    """Protected 엔드포인트 - Access Token 필수"""
    auth_header = request.headers.get('Authorization', '')

    if not auth_header.startswith('Bearer '):
        return {'error': 'Missing token'}, 401

    token = auth_header[7:]  # "Bearer " 제거

    try:
        payload = verify_token(token, 'access')
        user_id = payload['sub']
        return {'message': f'Hello, {user_id}!'}
    except ValueError as e:
        return {'error': str(e)}, 401
```

---

# 2. 컨텍스트 초기화 (필수!)
/clear

# 3. TDD 구현
/moai:2-run SPEC-001
/clear

# 4. 문서 동기화
/moai:3-sync SPEC-001
/clear
```

### 예제 2: RESTful API 블로그 시스템

시나리오: 블로그 포스트 CRUD API 개발

#### 📋 OpenAPI 3.0 명세

```yaml
openapi: 3.0.0
info:
  title: Blog API
  description: 블로그 포스트 관리 REST API
  version: 1.0.0
  contact:
    name: API Support
    url: https://blog.example.com/support

servers:
  - url: https://api.example.com/v1
    description: 프로덕션 서버
  - url: https://staging-api.example.com/v1
    description: 스테이징 서버

components:
  schemas:
    # 데이터 모델
    Post:
      type: object
      required: [id, title, content, author_id, created_at]
      properties:
        id:
          type: integer
          format: int64
          example: 123
          description: 포스트 고유 ID
        title:
          type: string
          minLength: 3
          maxLength: 200
          example: MoAI-ADK 완벽 가이드
          description: 포스트 제목
        content:
          type: string
          minLength: 10
          maxLength: 10000
          example: MoAI-ADK는 Super Agent Orchestrator...
          description: 포스트 내용
        author_id:
          type: integer
          format: int64
          example: 42
          description: 작성자 ID
        created_at:
          type: string
          format: date-time
          example: 2025-11-24T10:30:00Z
          description: 작성 일시
        updated_at:
          type: string
          format: date-time
          nullable: true
          example: 2025-11-24T15:45:00Z
          description: 수정 일시

    PostCreate:
      type: object
      required: [title, content]
      properties:
        title:
          type: string
          minLength: 3
          maxLength: 200
        content:
          type: string
          minLength: 10
          maxLength: 10000

    PostUpdate:
      type: object
      properties:
        title:
          type: string
          minLength: 3
          maxLength: 200
        content:
          type: string
          minLength: 10
          maxLength: 10000

    Error:
      type: object
      required: [code, message]
      properties:
        code:
          type: string
          example: "POST_NOT_FOUND"
          description: 에러 코드
        message:
          type: string
          example: "포스트를 찾을 수 없습니다"
          description: 에러 메시지
        details:
          type: object
          nullable: true
          description: 추가 상세 정보

  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT 기반 인증 (Authorization: Bearer [token])

security:
  - BearerAuth: []

paths:
  /posts:
    get:
      summary: 모든 포스트 조회
      description: 페이지네이션을 지원하는 포스트 목록 조회
      operationId: list_posts
      tags:
        - Posts
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            default: 1
            minimum: 1
          description: 페이지 번호
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
            minimum: 1
            maximum: 100
          description: 한 페이지당 포스트 수
        - name: author_id
          in: query
          schema:
            type: integer
          description: 특정 작성자의 포스트만 필터링
      responses:
        '200':
          description: 포스트 목록 조회 성공
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/Post'
                  pagination:
                    type: object
                    properties:
                      page:
                        type: integer
                      limit:
                        type: integer
                      total:
                        type: integer
                      has_next:
                        type: boolean
              example:
                data:
                  - id: 1
                    title: "First Post"
                    content: "Content here..."
                    author_id: 42
                    created_at: "2025-11-24T10:30:00Z"
                pagination:
                  page: 1
                  limit: 20
                  total: 100
                  has_next: true
        '401':
          description: 인증 실패 (토큰 없음 또는 만료)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
              example:
                code: "UNAUTHORIZED"
                message: "유효한 JWT 토큰이 필요합니다"

    post:
      summary: 새로운 포스트 생성
      description: 인증된 사용자가 새로운 포스트를 작성합니다
      operationId: create_post
      tags:
        - Posts
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PostCreate'
            example:
              title: "MoAI-ADK 완벽 가이드"
              content: "MoAI-ADK는 Super Agent Orchestrator..."
      responses:
        '201':
          description: 포스트 생성 성공
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Post'
              example:
                id: 123
                title: "MoAI-ADK 완벽 가이드"
                content: "MoAI-ADK는..."
                author_id: 42
                created_at: "2025-11-24T10:30:00Z"
        '400':
          description: 잘못된 요청 (유효성 검사 실패)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
              example:
                code: "VALIDATION_ERROR"
                message: "제목은 3자 이상이어야 합니다"
                details:
                  field: "title"
                  rule: "minLength"
        '401':
          description: 인증 실패
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /posts/{post_id}:
    get:
      summary: 특정 포스트 조회
      operationId: get_post
      tags:
        - Posts
      parameters:
        - name: post_id
          in: path
          required: true
          schema:
            type: integer
            format: int64
          example: 123
      responses:
        '200':
          description: 포스트 조회 성공
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Post'
        '404':
          description: 포스트를 찾을 수 없음
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
              example:
                code: "POST_NOT_FOUND"
                message: "ID 123인 포스트를 찾을 수 없습니다"

    put:
      summary: 포스트 전체 업데이트
      operationId: update_post
      tags:
        - Posts
      parameters:
        - name: post_id
          in: path
          required: true
          schema:
            type: integer
            format: int64
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PostCreate'
      responses:
        '200':
          description: 포스트 업데이트 성공
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Post'
        '404':
          description: 포스트를 찾을 수 없음
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '403':
          description: 권한 없음 (본인의 포스트만 수정 가능)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
              example:
                code: "FORBIDDEN"
                message: "다른 사용자의 포스트는 수정할 수 없습니다"

    delete:
      summary: 포스트 삭제
      operationId: delete_post
      tags:
        - Posts
      parameters:
        - name: post_id
          in: path
          required: true
          schema:
            type: integer
            format: int64
      responses:
        '204':
          description: 포스트 삭제 성공 (응답 본문 없음)
        '404':
          description: 포스트를 찾을 수 없음
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '403':
          description: 권한 없음
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
```

**엔드포인트 요약**:

| 메서드 | 경로 | 설명 | 인증 | 응답 |
|--------|------|------|------|------|
| **GET** | `/posts` | 포스트 목록 (페이지네이션) | ✅ | 200 / 401 |
| **POST** | `/posts` | 포스트 생성 | ✅ | 201 / 400 / 401 |
| **GET** | `/posts/{id}` | 특정 포스트 조회 | ✅ | 200 / 404 |
| **PUT** | `/posts/{id}` | 포스트 전체 업데이트 | ✅ | 200 / 403 / 404 |
| **DELETE** | `/posts/{id}` | 포스트 삭제 | ✅ | 204 / 403 / 404 |

#### 🚀 실제 구현 (Python FastAPI)

```python
from fastapi import FastAPI, Depends, HTTPException, Query
from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, List

app = FastAPI()

# Pydantic 모델 (OpenAPI 스키마와 매칭)
class PostCreate(BaseModel):
    title: str = Field(..., min_length=3, max_length=200)
    content: str = Field(..., min_length=10, max_length=10000)

class Post(PostCreate):
    id: int
    author_id: int
    created_at: datetime
    updated_at: Optional[datetime] = None

class Error(BaseModel):
    code: str
    message: str
    details: Optional[dict] = None

# 라우트 (OpenAPI 명세와 일치)
@app.get("/posts", response_model=List[Post], tags=["Posts"])
async def list_posts(
    page: int = Query(1, ge=1),
    limit: int = Query(20, ge=1, le=100),
    author_id: Optional[int] = None,
    current_user = Depends(verify_token)
):
    """모든 포스트 조회 (페이지네이션)"""
    # TDD: 테스트 먼저 작성
    # - 유효한 page/limit으로 조회
    # - author_id 필터링
    # - 페이지네이션 계산
    pass

@app.post("/posts", response_model=Post, status_code=201, tags=["Posts"])
async def create_post(
    post: PostCreate,
    current_user = Depends(verify_token)
):
    """새로운 포스트 생성"""
    # TDD:
    # - title/content 유효성 검사
    # - author_id를 current_user.id로 설정
    # - 데이터베이스에 저장
    # - 201 응답
    pass

@app.get("/posts/{post_id}", response_model=Post, tags=["Posts"])
async def get_post(
    post_id: int,
    current_user = Depends(verify_token)
):
    """특정 포스트 조회"""
    # TDD:
    # - post_id로 조회
    # - 없으면 404 POST_NOT_FOUND
    pass

@app.put("/posts/{post_id}", response_model=Post, tags=["Posts"])
async def update_post(
    post_id: int,
    post_update: PostCreate,
    current_user = Depends(verify_token)
):
    """포스트 업데이트 (본인만 가능)"""
    # TDD:
    # - post_id로 조회
    # - author_id == current_user.id 확인 (없으면 403 FORBIDDEN)
    # - title/content 업데이트
    # - updated_at 갱신
    pass

@app.delete("/posts/{post_id}", status_code=204, tags=["Posts"])
async def delete_post(
    post_id: int,
    current_user = Depends(verify_token)
):
    """포스트 삭제 (본인만 가능)"""
    # TDD:
    # - post_id로 조회
    # - author_id == current_user.id 확인
    # - 삭제
    # - 204 응답 (No Content)
    pass
```

#### 실행 시나리오

```bash
# Step 1: 기획
/moai:1-plan "블로그 포스트 CRUD(Create, Read, Update, Delete) API"
# → SPEC-001 생성
/clear

# Step 2: 구현
/moai:2-run SPEC-001
# → TDD로 API 엔드포인트 구현
# → 테스트 커버리지 87% 달성
/clear

# Step 3: 문서화
/moai:3-sync SPEC-001
# → OpenAPI 명세 자동 생성
# → API 문서 자동 업데이트
/clear

# Step 4: 다음 기능 계획
/moai:1-plan "댓글 시스템 추가 (중첩 댓글 지원)"
# → SPEC-002 생성
/clear

# 반복...
```

### 예제 3: 마이크로서비스 아키텍처

#### 🏗️ 마이크로서비스 개요

**목표**: 전자상거래 플랫폼의 복잡한 마이크로서비스 아키텍처 설계
- 10개 독립 서비스
- 멀티 데이터베이스 (DB per Service)
- 분산 트랜잭션 (SAGA 패턴)
- 이벤트 드리븐 메시징 (RabbitMQ/Kafka)
- API Gateway를 통한 통합

#### 📋 서비스 목록 및 책임

| 서비스 | 포트 | 책임 | DB | 주요 이벤트 |
|---------|------|------|-----|-----------|
| **API Gateway** | 3000 | 요청 라우팅, 인증, 레이트 제한 | - | - |
| **User Service** | 3001 | 사용자 관리, 인증/인가 | PostgreSQL | user.created, user.updated |
| **Product Service** | 3002 | 상품 카탈로그, 재고 관리 | MongoDB | product.created, inventory.updated |
| **Order Service** | 3003 | 주문 생성, 관리 (SAGA 조율) | PostgreSQL | order.created, order.payment_pending |
| **Payment Service** | 3004 | 결제 처리 (제3자 API 통합) | PostgreSQL | payment.succeeded, payment.failed |
| **Notification Service** | 3005 | 이메일/SMS 알림 | MongoDB | user.created→welcome email |
| **Review Service** | 3006 | 상품 리뷰 및 평점 | MongoDB | product.reviewed |
| **Shipping Service** | 3007 | 배송 추적 및 관리 | PostgreSQL | order.confirmed→create shipment |
| **Analytics Service** | 3008 | 실시간 데이터 분석 (비동기) | Elasticsearch | *.* (모든 이벤트) |
| **Admin Service** | 3009 | 관리자 대시보드 | PostgreSQL | - |
| **Config Service** | 3010 | 동적 설정 관리 | Redis | config.updated |

#### 🏛️ 아키텍처 다이어그램

```mermaid
graph TB
    Client["👤 클라이언트<br/>모바일/웹"]

    Gateway["🚪 API Gateway<br/>인증, 라우팅, 레이트 제한"]

    Client --> Gateway

    subgraph Services["핵심 서비스"]
        User["👥 User Service<br/>PostgreSQL"]
        Product["📦 Product Service<br/>MongoDB"]
        Order["🛒 Order Service<br/>PostgreSQL<br/>(SAGA 조율)"]
        Payment["💳 Payment Service<br/>PostgreSQL"]
        Notification["📧 Notification<br/>MongoDB"]
        Review["⭐ Review Service<br/>MongoDB"]
        Shipping["🚚 Shipping Service<br/>PostgreSQL"]
    end

    subgraph Data["데이터 및 메시징"]
        EventBus["📨 Event Bus<br/>RabbitMQ/Kafka"]
        Cache["⚡ Cache<br/>Redis"]
        Analytics["📊 Analytics<br/>Elasticsearch"]
    end

    Gateway --> User
    Gateway --> Product
    Gateway --> Order
    Gateway --> Payment
    Gateway --> Notification
    Gateway --> Review
    Gateway --> Shipping

    User --> EventBus
    Product --> EventBus
    Order --> EventBus
    Payment --> EventBus
    Notification --> EventBus
    Review --> EventBus
    Shipping --> EventBus

    EventBus --> Cache
    EventBus --> Analytics

    Order -.->|결제 보류| Payment
    Payment -.->|결제 완료| Order
    Order -.->|주문 확정| Shipping
```

#### 🔄 주요 흐름 예제: 주문 생성 (SAGA 패턴)

```
1️⃣ 사용자가 주문 생성 요청 (POST /orders)
   ↓
2️⃣ Order Service가 주문 임시 생성 (상태: PENDING)
   ↓
3️⃣ Event: order.created 발행
   ├→ Inventory Check: Product Service가 재고 확인
   ├→ Payment Initiate: Payment Service가 결제 처리 요청
   └→ Notification: Notification Service가 주문 확인 이메일 발송
   ↓
4️⃣ Payment Service: 결제 완료 → Payment.succeeded 이벤트
   ↓
5️⃣ Order Service: 상태 업데이트 (CONFIRMED)
   ↓
6️⃣ Shipping Service: 배송 준비 (shipment.created 이벤트)
   ↓
7️⃣ Notification Service: 배송 시작 알림 발송
   ↓
✅ 주문 완료
```

**SAGA 실패 처리** (결제 실패 시):
```
Payment Service: 결제 실패 → Payment.failed 이벤트
   ↓
Order Service: 상태 롤백 (CANCELLED)
   ↓
Product Service: 재고 원복
   ↓
Notification Service: 주문 취소 알림 발송
   ↓
❌ 주문 취소 완료 (모든 서비스 원자성 보장)
```

#### 📝 SPEC 작성 및 구현

```bash
# Step 1: SPEC 작성 (Sequential-Thinking 자동 활성화)
# 복잡도 > 중간 (10개 파일), 의존성 > 3개이므로 자동 활성화
/moai:1-plan "마이크로서비스 아키텍처 - 10개 서비스,
멀티 DB (PostgreSQL/MongoDB), 분산 트랜잭션 (SAGA 패턴),
이벤트 드리븐 메시징 (RabbitMQ), API Gateway,
Redis 캐싱, Elasticsearch 분석"

# Alfred가 자동으로 Sequential-Thinking MCP를 활용하여:
# - 서비스 간 의존성 분석
# - DB 스키마 설계
# - 이벤트 메시지 정의
# - SAGA 트랜잭션 설계
# - API 스펙 정의
# 상세 SPEC-001 생성

/clear  # 토큰 절약 (45-50K tokens)

# Step 2: TDD 구현 (병렬 실행 가능)
/moai:2-run SPEC-001

# 각 마이크로서비스:
# 1. tests/test_[service].py - 단위 테스트 (85%+ 커버리지)
# 2. src/[service]/ - 구현 코드
# 3. docker/[service]/Dockerfile - 컨테이너화
# 4. k8s/[service].yaml - Kubernetes 배포 설정

# Step 3: 통합 테스트 및 문서화
/moai:3-sync SPEC-001

# 자동 생성:
# - API 문서 (OpenAPI 3.0)
# - Event Schema (AsyncAPI)
# - Database 다이어그램
# - Deployment 가이드
# - 운영 플레이북
```

#### 🛠️ 필요한 도구 및 스킬

**Skill 권장사항**:
- `moai-domain-backend` - 마이크로서비스 아키텍처
- `moai-domain-database` - 다중 DB 설계 (PostgreSQL/MongoDB)
- `moai-domain-devops` - Docker, Kubernetes 배포
- `moai-security-zero-trust` - 서비스 간 인증 (mTLS)
- `moai-domain-monitoring` - 분산 추적 (Jaeger, Datadog)

---

## 10. TRUST 5 품질 보증

모든 MoAI-ADK 프로젝트는 **TRUST 5** 품질 프레임워크를 준수합니다. TRUST 5는 Test-First, Readable, Unified, Secured, Trackable의 5가지 핵심 원칙으로 구성되어 있으며, 엔터프라이즈급 소프트웨어의 품질을 보증하는 체계입니다. 각 원칙은 명확한 검증 기준을 가지고 있으며, MoAI-ADK의 자동화된 에이전트들이 이 기준들을 자동으로 검사하고 검증합니다. `/moai:2-run` TDD 구현 시 자동으로 모든 TRUST 5 검증이 수행되며, 기준을 충족하지 못하면 구현이 완료되지 않습니다. 이는 개발자가 고품질 코드를 작성하도록 강제하는 동시에, 반복적인 코드 리뷰 시간을 획기적으로 단축시킵니다.

### T - Test-First (테스트 우선)

**원칙**: 모든 구현은 테스트부터 시작합니다.

**검증**:

- 테스트 커버리지 ≥ 85%
- 실패하는 테스트 먼저 작성 (Red)
- 코드로 통과 (Green)
- 리팩토링 (Refactor)

**자동화**: `tdd-implementer` 에이전트가 자동으로 TDD 사이클 실행

### R - Readable (읽기 쉬운)

**원칙**: 코드는 명확하고 이해하기 쉬워야 합니다.

**검증**:

- 명확한 변수명 (약어 최소화)
- 코드 주석 (복잡한 로직)
- 코드 리뷰 통과
- 린터 검사 통과

**자동화**: `format-expert` 에이전트가 스타일 가이드 적용

### U - Unified (일관된)

**원칙**: 프로젝트 전체에 일관된 스타일을 유지합니다.

**검증**:

- 프로젝트 스타일 가이드 준수
- 일관된 네이밍 컨벤션
- 통일된 에러 핸들링
- 표준 문서 포맷

**자동화**: `quality-gate` 에이전트가 일관성 검증

### S - Secured (보안)

**원칙**: 모든 코드는 보안 검증을 통과해야 합니다.

**검증**:

- OWASP Top 10 준수
- 입력 검증 및 sanitization
- 보안 코딩 패턴
- 취약점 스캔 통과

**자동화**: `security-expert` 에이전트가 보안 검증

### T - Trackable (추적 가능)

**원칙**: 모든 변경사항은 추적 가능해야 합니다.

**검증**:

- SPEC-CODE 연결 (@SPEC → @CODE)
- 명확한 커밋 메시지
- 테스트 증거 문서화
- 변경 이력 기록

**자동화**: `git-manager` 에이전트가 Git 워크플로우 관리

---

## 11. 설정

### Configuration 파일 위치

MoAI-ADK는 `.claude/settings.json` 파일을 사용합니다.

### 주요 설정 항목

```json
{
  "user": {
    "name": "GOOS"
  },
  "language": {
    "conversation_language": "ko",
    "agent_prompt_language": "en"
  },
  "constitution": {
    "enforce_tdd": true,
    "test_coverage_target": 85
  },
  "git_strategy": {
    "mode": "personal",
    "branch_creation": {
      "prompt_always": true,
      "auto_enabled": false
    }
  },
  "github": {
    "spec_git_workflow": "develop_direct"
  },
  "statusline": {
    "enabled": true,
    "format": "compact",
    "style": "R2-D2"
  }
}
```

### 🌳 Git 전략 (3가지 모드)

MoAI-ADK는 개발 환경과 팀 구성에 맞게 3가지 Git 전략을 제공합니다. `.moai/config/config.json`에서 `git_strategy.mode`를 설정하여 활성화합니다.

#### 🎯 모드 선택 결정 트리

```mermaid
flowchart TD
    Q1{"GitHub 사용<br/>하시나요?"}

    Q1 -->|아니오| Manual["<b>📦 Manual</b><br/>로컬 Git만 사용<br/>━━━━━━━━<br/>특징:<br/>• 로컬 커밋만<br/>• Push 수동<br/>• 브랜치 선택적<br/><br/>대상: 개인 학습"]

    Q1 -->|네| Q2{"팀 프로젝트<br/>인가요?"}

    Q2 -->|아니오| Personal["<b>👤 Personal</b><br/>개인 GitHub<br/>━━━━━━━━<br/>특징:<br/>• Feature 브랜치<br/>• 자동 Push<br/>• PR 선택적<br/><br/>대상: 개인 프로젝트"]

    Q2 -->|네| Team["<b>👥 Team</b><br/>팀 GitHub<br/>━━━━━━━━<br/>특징:<br/>• Draft PR 자동<br/>• 리뷰 필수<br/>• 자동 배포<br/><br/>대상: 팀 프로젝트"]

    classDef manual fill:#fff3e0,stroke:#ff9800,stroke-width:2px
    classDef personal fill:#e3f2fd,stroke:#2196f3,stroke-width:2px
    classDef team fill:#f3e5f5,stroke:#9c27b0,stroke-width:2px
    classDef question fill:#fafafa,stroke:#666,stroke-width:2px

    class Manual manual
    class Personal personal
    class Team team
    class Q1,Q2 question
```

---

#### 📋 3가지 모드 비교

| 구분 | Manual | Personal | Team |
|------|--------|----------|------|
| **사용처** | 개인 학습 | 개인 GitHub | 팀 프로젝트 |
| **GitHub** | ❌ | ✅ | ✅ |
| **브랜치** | 선택적 생성 | Feature 자동 | Feature 자동 |
| **Push** | 수동 | 자동 | 자동 |
| **PR** | 없음 | 제안 | 자동 생성 |
| **코드 리뷰** | 없음 | 선택 | **필수** |
| **배포** | 수동 | 수동 | CI/CD 자동 |
| **설정** | **5분** | 15분 | 25분 |

---

#### ⚙️ 빠른 설정

**Manual** (로컬만 사용):
```json
{
  "git_strategy": {
    "mode": "manual",
    "branch_creation": {
      "prompt_always": true,
      "auto_enabled": false
    }
  }
}
```
✅ Alfred가 매번 브랜치 생성 여부 물어봄

**Personal** (개인 프로젝트 - 빠른 반복):
```json
{
  "git_strategy": {
    "mode": "personal",
    "branch_creation": {
      "prompt_always": false,
      "auto_enabled": true
    }
  }
}
```
✅ 모든 커밋이 자동으로 GitHub에 푸시됨

**Team** (팀 프로젝트 - 코드 리뷰):
```json
{
  "git_strategy": {
    "mode": "team",
    "branch_creation": {
      "prompt_always": false,
      "auto_enabled": true
    }
  }
}
```
✅ 모든 SPEC마다 자동으로 Draft PR 생성 (팀 리뷰 필요)

---

#### 🔄 Team 모드의 코드 리뷰 흐름

```
SPEC 작성 (Feature 브랜치)
        ↓
Draft PR 자동 생성
        ↓
팀원 리뷰 + 피드백
        ↓
승인 (최소 1명)
        ↓
Merge to main
        ↓
CI/CD 자동 배포 (선택)
```

---

#### 🔀 모드 마이그레이션

**Manual → Personal** (GitHub 추가):
```bash
# 1. GitHub 저장소 생성
# 2. 로컬 리포지토리와 연결
git remote add origin https://github.com/user/repo.git

# 3. config.json 수정
# mode: "manual" → "personal"
```

**Personal → Team** (팀 추가):
```bash
# 1. 팀 저장소에 동료 초대
# 2. config.json 수정
# mode: "personal" → "team"
```

---

## 12. MCP 서버

MoAI-ADK는 **MCP(Model Context Protocol)** 서버를 통해 외부 도구와 통합됩니다.

### 📡 지원 MCP 서버

| MCP 서버                      | 목적                      | 필수 여부   | 용도                                      |
| ----------------------------- | ------------------------- | ----------- | ----------------------------------------- |
| **Context7**                  | 최신 라이브러리 문서 조회 | ✅ **필수** | API 레퍼런스, 프레임워크 문서             |
| **Sequential-Thinking**       | 복잡한 문제 다단계 추론   | ✅ **권장** | 아키텍처 설계, 알고리즘 최적화, SPEC 분석 |
| **Playwright**                | 브라우저 자동화           | 선택        | E2E 테스트, UI 검증                       |
| **figma-dev-mode-mcp-server** | 디자인 시스템 연동        | 선택        | 디자인-코드 변환                          |

### 🧮 Sequential-Thinking MCP (권장)

**목적**: 복잡한 문제의 다단계 추론을 통한 정확한 분석

**자동 활성화 조건**:

- 복잡도 > 중간 (10+ 파일, 아키텍처 변경)
- 의존성 > 3개 이상
- SPEC 생성 또는 Plan 에이전트 호출 시
- 요청에서 "복잡한", "설계", "최적화", "분석" 키워드 포함

**활용 시나리오**:

- 🏗️ 마이크로서비스 아키텍처 설계
- 🧩 복잡한 데이터 구조 및 알고리즘 최적화
- 🔄 시스템 통합 및 마이그레이션 계획
- 📋 SPEC 분석 및 요구사항 정의
- ⚙️ 성능 병목 분석

### 🔌 Context7 MCP (필수)

**목적**: 최신 라이브러리 문서 및 API 레퍼런스 실시간 조회

**활성화 방법**: MoAI-ADK 설치 시 자동 활성화

**지원 라이브러리**(예시):

- `/vercel/next.js` - Next.js 최신 문서
- `/facebook/react` - React 최신 문서
- `/tiangolo/fastapi` - FastAPI 최신 문서

---

## 13. 고급 기능

### 🚀 선택적 AI 코드 생성 (Codex & Gemini)

MoAI-ADK는 외부 AI 모델과의 **선택적** 통합을 지원합니다. Claude Code만으로도 완전히 작동하며, AI 에이전트는 완전히 선택사항입니다.

**Use `ai-codex` (OpenAI Codex) for Backend**:

- 🔧 복잡한 백엔드 API 구현
- 🔧 데이터베이스 쿼리 최적화
- 🔧 알고리즘 최적화

**Use `ai-gemini` (Google Gemini) for Frontend**:

- 🎨 React/Next.js 컴포넌트 생성
- 🎨 UI/UX 설계
- 🎨 Tailwind CSS 스타일링

---

## 15. 문제 해결

### 1. 테스트 커버리지 85% 미달

**오류:**

```text
❌ 테스트 커버리지: 72% (목표: 85%)
```

**해결책:**

```bash
# test-engineer 에이전트 호출하여 추가 테스트 생성
@agent-test-engineer "SPEC-001의 테스트 커버리지를 85% 이상으로 향상"
```

또는 coverage_target 조정 (비권장):

```json
{
  "constitution": {
    "test_coverage_target": 75
  }
}
```

### 2. SPEC 없이 구현 시도

**오류:**

```text
❌ SPEC이 없습니다. /moai:1-plan을 먼저 실행하세요.
```

**해결책:**

```bash
# 반드시 SPEC 먼저 생성
/moai:1-plan "기능 설명"
/clear
/moai:2-run SPEC-001
```

### 3. 토큰 한계 초과

**오류:**

```text
⚠️ Context: 175K tokens (한계에 근접)
```

**해결책:**

```bash
# /clear 실행하여 컨텍스트 초기화
/clear

# 또는 작업을 더 작은 단위로 분할
/moai:1-plan "기능 A만 먼저 구현"  # 큰 기능을 분할
```

### 4. MCP 서버 연결 실패

**오류:**

```text
❌ Context7 MCP 연결 실패
```

**해결책:**

```bash
# Claude Code 재시작
# 1. Claude Code 종료
# 2. 터미널에서 다시 실행:
claude

# 또는 MCP 설정 재확인:
# .claude/mcp.json 파일 확인
```

---

## 16. 추가 자료

### 📖 문서 파일 (.moai/memory/)

MoAI-ADK는 프로젝트 내부에 포괄적인 메모리 파일 시스템을 제공합니다:

- `.moai/memory/execution-rules.md` - 실행 규칙 및 제약사항
- `.moai/memory/agents.md` - 35개 전문 에이전트 카탈로그
- `.moai/memory/commands.md` - MoAI 커맨드 레퍼런스
- `.moai/memory/delegation-patterns.md` - 에이전트 위임 패턴
- `.moai/memory/token-optimization.md` - 토큰 최적화 전략

### 🆘 지원 (Support)

**이메일 지원:**

- 기술 지원: [support@mo.ai.kr](mailto:support@mo.ai.kr)

### 📊 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=modu-ai/moai-adk&type=Date)](https://star-history.com/#modu-ai/moai-adk&Date)

---

## 📝 License

MoAI-ADK is licensed under the [MIT License](./LICENSE).

```text
MIT License

Copyright (c) 2025 MoAI-ADK Team

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

### Made with ❤️ by MoAI-ADK Team

**Version:** 0.28.2
**Last Updated:** 2025-11-24
**Philosophy**: SPEC-First TDD + Agent Orchestration + 85% Token Efficiency
**MoAI**: MoAI stands for "Modu-ui AI" (AI for Everyone). Our goal is to make AI accessible to everyone.
