# 용어집

**@DOC:GLOSSARY-001** | **최종 업데이트**: 2025-11-10 | **대상**: 초보자

MoAI-ADK에서 사용하는 핵심 용어를 알파벳 순서로 정렬했습니다. 각 용어는 정의, 한국어 설명, 예시로 구성되어 있습니다.

______________________________________________________________________

## A

<span class="material-icons">smart_toy</span> **AI 전문가 시스템**

### Agent (에이전트)

**정의**: 특정 작업을 수행하는 AI 전문가 시스템입니다.

**설명**: MoAI-ADK는 19명의 전문가 에이전트로 구성되어 있습니다. 각 에이전트는 고유한 역할을 가지며, Alfred가 이들을 조율합니다. 에이전트는 SPEC 작성, 코드 구현, 문서 동기화 등의 작업을 수행합니다.

**예시**: 
- `spec-builder`: SPEC 문서를 생성하는 전문가
- `tdd-implementer`: TDD 방식으로 코드를 작성하는 전문가
- `doc-syncer`: 문서를 동기화하는 전문가

______________________________________________________________________

### Alfred (알프레드)

**정의**: MoAI-ADK의 슈퍼에이전트로, 19명의 전문가 팀을 조율하는 AI 시스템입니다.

**설명**: Alfred는 사용자의 요청을 이해하고, 적절한 전문가 에이전트를 선택하여 작업을 위임합니다. SPEC → TDD → Sync 개발 워크플로우를 자동화하며, 모든 작업의 품질을 보장합니다.

**예시**: 
```bash
/alfred:1-plan "사용자 인증 기능"
```
이 명령을 실행하면 Alfred가 spec-builder 에이전트를 호출하여 사용자 인증 기능에 대한 SPEC 문서를 생성합니다.

______________________________________________________________________

## B

<span class="material-icons">account_tree</span> **Git 브랜치 시스템**

### Branch (브랜치)

**정의**: Git에서 독립적인 개발 작업을 수행하기 위한 분기입니다.

**설명**: MoAI-ADK는 GitFlow 전략을 사용합니다. 각 기능은 별도의 feature 브랜치에서 개발되며, develop 브랜치로 병합된 후 최종적으로 main 브랜치에 릴리스됩니다.

**예시**:
```bash
# 현재 브랜치 구조
feature/SPEC-001-user-auth  # 개발 중
  ↓
develop                      # 통합 테스트
  ↓
main                         # 프로덕션 릴리스
```

______________________________________________________________________

## C

<span class="material-icons">settings</span> **프로젝트 설정**

### CLAUDE.md

**정의**: 프로젝트의 AI 지침 문서입니다.

**설명**: Alfred가 이 문서를 읽고 프로젝트 규칙, 코드 스타일, 언어 정책, 품질 기준을 따릅니다. 프로젝트별로 맞춤 설정이 가능하며, 프로젝트 루트에 위치합니다.

**예시**:
```markdown
# CLAUDE.md
conversation_language: ko
codebase_language: python
git_strategy: team
```
이 설정을 통해 Alfred는 한국어로 대화하고, Python 코드를 작성하며, 팀 협업 Git 전략을 사용합니다.

______________________________________________________________________

### Command (명령어)

**정의**: 사용자가 Alfred에게 작업을 요청하는 진입점입니다.

**설명**: 명령어는 `/alfred:` 접두사로 시작하며, 여러 에이전트를 조율하여 복잡한 워크플로우를 실행합니다. 각 명령어는 특정 단계(계획, 실행, 동기화)를 담당합니다.

**예시**:
```bash
/alfred:0-project    # 프로젝트 초기화
/alfred:1-plan       # SPEC 계획 생성
/alfred:2-run        # TDD 구현 실행
/alfred:3-sync       # 문서/코드 동기화
```

______________________________________________________________________

### Config (설정 파일)

**정의**: `.moai/config.json` 파일로, 프로젝트의 모든 설정을 담고 있습니다.

**설명**: 언어 설정, Git 전략, 품질 기준, 리포트 생성 규칙 등 프로젝트의 모든 동작을 제어합니다. Alfred는 항상 이 파일의 설정을 최우선으로 따릅니다.

**예시**:
```json
{
  "language": {
    "conversation_language": "ko",
    "codebase_language": "python"
  },
  "git_strategy": {
    "mode": "team",
    "use_gitflow": true
  },
  "quality": {
    "test_coverage_threshold": 80
  }
}
```

______________________________________________________________________

## D

<span class="material-icons">developer_board</span> **개발 환경**

### Develop Branch (개발 브랜치)

**정의**: GitFlow에서 기능 통합 및 테스트를 수행하는 브랜치입니다.

**설명**: 모든 feature 브랜치는 develop 브랜치로 병합됩니다. develop 브랜치에서 통합 테스트를 완료한 후 main 브랜치로 릴리스합니다.

**예시**:
```bash
# Feature 브랜치를 develop에 병합
git checkout develop
git merge feature/SPEC-001-user-auth
git push origin develop
```

______________________________________________________________________

## E

<span class="material-icons">assignment</span> **요구사항 정의**

### EARS (요구사항 작성 패턴)

**정의**: **E**vent-**A**ction-**R**esponse-**S**tate 형식의 요구사항 작성 방법입니다.

**설명**: 명확하고 테스트 가능한 요구사항을 작성하기 위한 템플릿입니다. 각 요소는 특정 상황, 동작, 결과, 상태를 정의합니다.

**예시**:
```
WHEN 사용자가 로그인 버튼을 클릭하면 (Event)
  AND 이메일과 비밀번호를 입력했을 때 (Condition)
THEN 시스템은 인증을 수행하고 (Action)
  AND 대시보드 페이지로 이동한다 (Response)
WHERE 사용자 세션이 생성된 상태 (State)
```

______________________________________________________________________

## F

<span class="material-icons">new_releases</span> **기능 개발**

### Feature Branch (기능 브랜치)

**정의**: 특정 기능을 개발하기 위한 독립적인 브랜치입니다.

**설명**: `feature/SPEC-XXX-기능명` 형식으로 생성되며, 해당 기능의 개발이 완료되면 develop 브랜치로 병합됩니다.

**예시**:
```bash
# Feature 브랜치 생성
git checkout -b feature/SPEC-001-user-authentication
```

______________________________________________________________________

## G

<span class="material-icons">source</span> **Git 워크플로우**

### GitFlow

**정의**: Git 브랜치 전략으로, main, develop, feature 브랜치를 계층적으로 관리합니다.

**설명**: MoAI-ADK의 기본 Git 전략입니다. 각 브랜치는 명확한 역할을 가지며, 안전한 릴리스 프로세스를 보장합니다.

**예시**:
```
feature/SPEC-XXX → develop → main
   (개발)         (통합)    (릴리스)
```

______________________________________________________________________

### GREEN (TDD 단계)

**정의**: TDD 사이클의 두 번째 단계로, 테스트를 통과시키는 최소한의 코드를 작성합니다.

**설명**: RED 단계에서 작성한 실패하는 테스트를 통과시키기 위한 코드를 작성합니다. 과도한 기능 추가 없이 테스트만 통과시키는 것이 원칙입니다.

**예시**:
```python
# RED 단계에서 작성한 테스트
def test_login():
    result = authenticate("user@example.com", "password123")
    assert result == True

# GREEN 단계: 테스트를 통과시키는 최소 코드
def authenticate(email, password):
    return True  # 일단 테스트만 통과
```

______________________________________________________________________

## H

<span class="material-icons">webhook</span> **이벤트 훅 시스템**

### Hook (훅)

**정의**: 특정 이벤트 발생 시 자동으로 실행되는 가벼운 검사 스크립트입니다.

**설명**: 세션 시작, 도구 사용 전 등의 이벤트에 반응하여 안전 검사, 컨텍스트 제공, 상태 카드 표시 등을 수행합니다. 100ms 이하의 빠른 실행 시간을 가집니다.

**예시**:
- `SessionStart`: 세션 시작 시 프로젝트 정보 제공
- `PreToolUse`: 파괴적인 명령 실행 전 경고

______________________________________________________________________

## M

<span class="material-icons">rocket_launch</span> **프로덕션 릴리스**

### Main Branch (메인 브랜치)

**정의**: GitFlow에서 프로덕션 릴리스를 담당하는 브랜치입니다.

**설명**: main 브랜치는 항상 배포 가능한 상태를 유지합니다. develop 브랜치에서 충분한 테스트를 거친 후에만 병합됩니다.

**예시**:
```bash
# develop에서 main으로 릴리스
git checkout main
git merge develop
git tag v0.21.0
git push origin main --tags
```

______________________________________________________________________

## P

<span class="material-icons">merge</span> **코드 병합**

### PR (Pull Request)

**정의**: 코드 병합을 요청하는 GitHub 기능입니다.

**설명**: feature 브랜치에서 develop 또는 develop에서 main으로 병합하기 위한 공식 절차입니다. 코드 리뷰, 자동 테스트, 품질 검증을 거칩니다.

**예시**:
```bash
# PR 생성 (feature → develop)
gh pr create --base develop --head feature/SPEC-001-user-auth \
  --title "SPEC-001: 사용자 인증 기능 추가" \
  --body "사용자 이메일/비밀번호 인증 구현"
```

______________________________________________________________________

## R

<span class="material-icons">fact_check</span> **테스트 주도 개발**

### RED (TDD 단계)

**정의**: TDD 사이클의 첫 번째 단계로, 실패하는 테스트를 먼저 작성합니다.

**설명**: 구현하고자 하는 기능의 기대 동작을 테스트로 먼저 정의합니다. 이 테스트는 아직 구현이 없기 때문에 실패해야 합니다.

**예시**:
```python
# RED 단계: 실패하는 테스트 작성
def test_user_authentication():
    user = User("test@example.com", "password123")
    assert user.authenticate() == True
    # 실패: User 클래스가 아직 구현되지 않음
```

______________________________________________________________________

### REFACTOR (TDD 단계)

**정의**: TDD 사이클의 세 번째 단계로, 테스트를 유지하면서 코드 품질을 개선합니다.

**설명**: GREEN 단계에서 작성한 코드를 리팩토링합니다. 테스트는 여전히 통과해야 하며, 코드의 가독성, 유지보수성, 성능을 향상시킵니다.

**예시**:
```python
# GREEN 단계 코드
def authenticate(email, password):
    return True

# REFACTOR 단계: 실제 인증 로직 추가
def authenticate(email, password):
    if not email or not password:
        return False
    user = database.find_user(email)
    return user and user.verify_password(password)
```

______________________________________________________________________

## S

<span class="material-icons">psychology</span> **지식 시스템**

### Skill (스킬)

**정의**: 500단어 이하의 재사용 가능한 지식 캡슐입니다.

**설명**: 특정 작업에 대한 가이드라인, 템플릿, 체크리스트를 담고 있습니다. `.claude/skills/` 디렉토리에 저장되며, 필요할 때만 로드됩니다.

**예시**:
```bash
# Skill 호출
Skill("moai-alfred-agent-guide")
Skill("moai-foundation-tags")
```

______________________________________________________________________

### SPEC (사양 문서)

**정의**: 요구사항을 명확하게 정의한 문서입니다.

**설명**: 모든 개발은 SPEC에서 시작합니다. SPEC-First 원칙에 따라 요구사항을 먼저 정의하고, 이를 기반으로 테스트와 코드를 작성합니다.

**예시**:
```markdown
# SPEC-001: 사용자 인증

@SPEC:AUTH-001

## 요구사항

WHEN 사용자가 로그인 폼을 제출하면
  AND 올바른 이메일과 비밀번호를 입력했을 때
THEN 시스템은 사용자를 인증하고
  AND 대시보드로 리디렉션한다
```

______________________________________________________________________

### SPEC-First

**정의**: 요구사항 정의를 가장 먼저 수행하는 개발 원칙입니다.

**설명**: 코드나 테스트를 작성하기 전에 SPEC 문서를 먼저 작성합니다. 이를 통해 명확한 목표 설정과 추적 가능한 개발이 가능합니다.

**예시**:
```bash
# 1. SPEC 작성
/alfred:1-plan "사용자 인증 기능"

# 2. TDD 구현
/alfred:2-run SPEC-001

# 3. 문서 동기화
/alfred:3-sync auto SPEC-001
```

______________________________________________________________________

## T

<span class="material-icons">link</span> **추적성 시스템**

### @TAG (태그)

**정의**: 요구사항(SPEC), 테스트(TEST), 코드(CODE), 문서(DOC)를 연결하는 추적성 시스템입니다.

**설명**: 각 기능마다 고유한 TAG를 부여하여 어느 문서에서 어디로 구현되었는지 추적할 수 있습니다. 4가지 유형의 TAG가 체인을 형성합니다.

**예시**:
```
@SPEC:AUTH-001 (요구사항 정의)
    ↓
@TEST:AUTH-001 (테스트 케이스)
    ↓
@CODE:AUTH-001:LOGIN (구현 코드)
    ↓
@DOC:AUTH-001 (문서화)
```

______________________________________________________________________

### TDD (Test-Driven Development)

**정의**: 테스트를 먼저 작성하고 구현하는 개발 방법론입니다.

**설명**: RED(실패 테스트) → GREEN(테스트 통과) → REFACTOR(코드 개선) 사이클을 반복합니다. MoAI-ADK는 TDD를 필수로 적용합니다.

**예시**:
```bash
# TDD 사이클 실행
/alfred:2-run SPEC-001

# 1. RED: 실패하는 테스트 작성
# 2. GREEN: 테스트를 통과시키는 코드 작성
# 3. REFACTOR: 코드 품질 개선
```

______________________________________________________________________

## TR

<span class="material-icons">verified</span> **품질 보증**

### TRUST 5 (품질 원칙)

**정의**: MoAI-ADK의 5가지 핵심 품질 원칙입니다.

**설명**:
- **T**est First: 테스트를 먼저 작성
- **R**eadable: 읽기 쉬운 코드
- **U**nified: 일관된 스타일
- **S**ecured: 보안 우선
- **T**rackable: 추적 가능한 변경

**예시**:
```python
# Test First: 테스트부터 작성
def test_calculate_total():
    assert calculate_total([10, 20, 30]) == 60

# Readable: 명확한 변수명
def calculate_total(prices):
    return sum(prices)

# Unified: 일관된 코드 스타일 (PEP 8)
# Secured: 입력 검증
# Trackable: @TAG 추가
```

______________________________________________________________________

## U

<span class="material-icons">package</span> **패키지 관리**

### UV (패키지 매니저)

**정의**: Python 프로젝트의 의존성을 관리하는 빠른 패키지 매니저입니다.

**설명**: pip보다 빠르고 안정적이며, MoAI-ADK의 기본 패키지 관리 도구입니다. 가상환경 생성, 의존성 설치, 스크립트 실행을 담당합니다.

**예시**:
```bash
# UV 설치
curl -LsSf https://astral.sh/uv/install.sh | sh

# 의존성 설치
uv sync

# 스크립트 실행
uv run dev
```

______________________________________________________________________

## W

<span class="material-icons">workflow</span> **개발 프로세스**

### Workflow (워크플로우)

**정의**: MoAI-ADK의 3단계 개발 프로세스입니다.

**설명**: SPEC → BUILD → SYNC 순서로 진행됩니다. 각 단계는 명령어로 실행되며, 자동화된 검증을 거칩니다.

**예시**:
```bash
# Phase 1: SPEC 정의
/alfred:1-plan "사용자 인증 기능"

# Phase 2: TDD 구현
/alfred:2-run SPEC-001

# Phase 3: 동기화 및 PR 생성
/alfred:3-sync auto SPEC-001
```

______________________________________________________________________

## 추가 참고 자료

<span class="material-icons">library_books</span> **문서 라이브러리**

- **전체 가이드**: [MoAI-ADK 문서](https://adk.mo.ai.kr)
- **SPEC 작성법**: [SPEC 기초 가이드](../guides/specs/basics.md)
- **TDD 실습**: [TDD 워크플로우](../guides/tdd/index.md)
- **Alfred 명령어**: [Alfred 가이드](../guides/alfred/index.md)

______________________________________________________________________

*최종 업데이트: 2025-11-10 | 버전: v0.21.2 | 상태: Phase 1 완료*
