# Alfred 워크플로우 명령어 가이드

MoAI-ADK의 핵심 4단계 워크플로우를 제어하는 Alfred 커맨드들입니다.

> **중요**: Alfred 명령어는 **Claude Code 환경 내에서만** 사용 가능합니다.

## 워크플로우 개요

```
/alfred:0-project (초기화)
    ↓
/alfred:1-plan (계획: SPEC 작성)
    ↓
/alfred:2-run (실행: TDD 개발)
    ↓
/alfred:3-sync (동기화: 문서/검증)
    ↓
완료 및 PR 생성
```

______________________________________________________________________

## 1. /alfred:0-project

**프로젝트 설정 및 초기화**

### 문법

```
/alfred:0-project [option]
```

### 옵션

```
setting     현재 설정 조회
update      프로젝트 설정 수정
```

### 주요 기능

- 🎯 프로젝트 메타데이터 설정 (이름, 설명, 언어)
- 🌍 대화 언어 선택 (한국어, 영어, 일본어, 중국어)
- 🔧 개발 모드 선택 (solo/team/org)
- 📋 SPEC-First TDD 체크리스트 초기화
- 🏷️ TAG 시스템 활성화
- 📊 테스트 커버리지 목표 설정 (기본 85%)

### 상호작용 예시

```
/alfred:0-project

> Alfred: 프로젝트 이름을 입력해주세요
사용자: Hello World API

> Alfred: 프로젝트 설명?
사용자: 간단한 REST API 튜토리얼

> Alfred: 주로 사용할 언어는?
사용자: [1] Python  [2] TypeScript  [3] Go
선택: 1

> Alfred: 대화 언어?
사용자: [1] 한국어  [2] English
선택: 1

> Alfred: 개발 모드?
사용자: [1] 솔로  [2] 팀  [3] 조직
선택: 1

✅ 프로젝트 초기화 완료!
```

### 생성되는 설정

`.moai/config.json`:

```json
{
  "project": {
    "name": "Hello World API",
    "description": "간단한 REST API 튜토리얼",
    "language": "python"
  },
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "한국어"
  },
  "constitution": {
    "test_coverage_minimum": 85
  }
}
```

______________________________________________________________________

## 2. /alfred:1-plan

**SPEC 작성 및 계획 수립**

### 문법

```
/alfred:1-plan "제목1" "제목2" ... | SPEC-ID 수정사항
```

### 사용 사례

#### 새 SPEC 작성

```
/alfred:1-plan "사용자 인증 시스템"
```

또는 여러 SPEC:

```
/alfred:1-plan "로그인 기능" "회원가입" "비밀번호 재설정"
```

#### 기존 SPEC 수정

```
/alfred:1-plan SPEC-001 "로그인 기능 (수정: OAuth 2.0 추가)"
```

### Alfred의 계획 수립 과정

1. **의도 파악**: 요청 분석 및 명확화

   - 모호하면 AskUserQuestion으로 추가 정보 수집

2. **계획 수립**: 구조화된 실행 전략

   - 작업 분해 (Decomposition)
   - 의존성 분석 (Dependency Analysis)
   - 병렬화 기회 식별 (Parallelization)
   - 영향받는 파일 명시 (File List)
   - 예상 시간 추정 (Time Estimation)

3. **사용자 승인**: 계획 제시 및 승인 요청

   ```
   Alfred: 다음과 같이 계획했습니다. 진행하시겠습니까?

   📋 계획 요약:
   - SPEC-001: 로그인 기능
   - SPEC-002: 회원가입
   - 영향 파일: 5개
   - 예상 시간: 30분

   [진행] [수정] [취소]
   ```

4. **TodoWrite 초기화**: 모든 작업 항목 추적 시작

### 생성되는 파일

```
.moai/specs/SPEC-001/
├── spec.md              # SPEC 문서 (EARS 형식)
├── requirements.md      # 요구사항 상세
└── tests.md            # 테스트 계획
```

### SPEC 문서 구조

```markdown
# SPEC-001: 로그인 기능

## 요구사항 (EARS 형식)

### 기본 요구사항
- GIVEN: 사용자가 로그인 페이지 방문
  WHEN: 유효한 이메일과 비밀번호 입력
  THEN: 세션 생성 및 대시보드 리다이렉트

### 오류 처리
- GIVEN: 로그인 페이지
  WHEN: 잘못된 비밀번호 입력
  THEN: "비밀번호가 맞지 않습니다" 메시지

## 테스트 계획
- [ ] 유효한 인증정보로 로그인 성공
- [ ] 잘못된 인증정보로 로그인 실패
- [ ] 신규 사용자 가입 후 로그인
```

______________________________________________________________________

## 3. /alfred:2-run

**TDD 구현 실행**

### 문법

```
/alfred:2-run [SPEC-ID | "all"]
```

### 사용 사례

#### 특정 SPEC 개발

```
/alfred:2-run SPEC-001
```

#### 모든 SPEC 개발

```
/alfred:2-run all
```

### 실행 워크플로우

Alfred는 TDD의 3단계를 엄격히 따릅니다:

#### Phase 1: RED (빨강) - 실패하는 테스트 작성

```
Alfred: RED 단계 시작
- 테스트 파일 생성: tests/test_login.py
- 테스트 작성 (SPEC 기반)
- 실행 → 모두 실패 ❌

✅ RED 단계 완료
모든 테스트가 실패하는 상태입니다.

[GREEN 단계로 진행]
```

**샘플 테스트**:

```python
# tests/test_login.py @TEST:SPEC-001:*
import pytest
from app import login

def test_valid_login():
    """GIVEN: 로그인 페이지
       WHEN: 유효한 인증정보
       THEN: 세션 생성"""
    result = login("user@example.com", "password123")
    assert result["status"] == "success"
    assert result["session"] is not None

def test_invalid_password():
    """GIVEN: 로그인 페이지
       WHEN: 잘못된 비밀번호
       THEN: 오류 메시지"""
    result = login("user@example.com", "wrong")
    assert result["status"] == "error"
    assert "비밀번호" in result["message"]
```

#### Phase 2: GREEN (초록) - 최소 구현으로 테스트 통과

```
Alfred: GREEN 단계 시작
- 최소한의 구현 추가: app.py
- 테스트 실행 → 모두 통과 ✅

✅ GREEN 단계 완료
모든 테스트가 통과합니다.

[REFACTOR 단계로 진행]
```

**샘플 구현**:

```python
# app.py @CODE:SPEC-001:*
def login(email, password):
    """로그인 처리"""
    if password == "password123":
        return {
            "status": "success",
            "session": "session_123"
        }
    else:
        return {
            "status": "error",
            "message": "비밀번호가 맞지 않습니다"
        }
```

#### Phase 3: REFACTOR (리팩토) - 코드 품질 개선

```
Alfred: REFACTOR 단계 시작
- 에러 처리 개선
- 데이터 검증 추가
- 코드 정리

✅ REFACTOR 단계 완료
모든 테스트 통과, 코드 품질 개선.

[모든 작업 완료]
```

**개선된 구현**:

```python
# app.py (개선 후)
from flask import session
from werkzeug.security import check_password_hash
from models import User

def login(email, password):
    """로그인 처리 (개선된 버전)"""
    # 입력 검증
    if not email or not password:
        raise ValueError("이메일과 비밀번호는 필수입니다")

    # 사용자 조회
    user = User.query.filter_by(email=email).first()
    if not user:
        return {
            "status": "error",
            "message": "가입되지 않은 사용자입니다"
        }

    # 비밀번호 검증
    if not check_password_hash(user.password_hash, password):
        return {
            "status": "error",
            "message": "비밀번호가 맞지 않습니다"
        }

    # 세션 생성
    session['user_id'] = user.id
    return {
        "status": "success",
        "session": session.sid,
        "user": {
            "id": user.id,
            "email": user.email,
            "name": user.name
        }
    }
```

### TodoWrite 추적

Alfred는 각 단계를 자동으로 추적합니다:

```
[in_progress] RED: SPEC-001 테스트 작성
[completed]   RED: SPEC-001 테스트 작성
[in_progress] GREEN: SPEC-001 최소 구현
[completed]   GREEN: SPEC-001 최소 구현
[in_progress] REFACTOR: SPEC-001 코드 개선
[completed]   REFACTOR: SPEC-001 코드 개선
```

______________________________________________________________________

## 4. /alfred:3-sync

**문서 동기화 및 검증**

### 문법

```
/alfred:3-sync [Mode] [Target] [Path]
```

### 모드 (Mode)

```
auto         자동 동기화 (권장)
force        강제 동기화
status       현재 상태만 조회
project      전체 프로젝트 검증
```

### 대상 (Target)

```
SPEC-001         특정 SPEC 동기화
all              모든 SPEC 동기화
```

### 주요 기능

1. **문서 생성** (생성 설정에 따라)

   - API 문서 자동 생성
   - 아키텍처 다이어그램
   - 배포 가이드

2. **TAG 검증**

   - SPEC → TEST → CODE → DOC 연결 확인
   - 고아 TAG 탐지 및 제거
   - 추적성 무결성 확인

3. **품질 검증**

   - 테스트 커버리지 85% 이상 확인
   - 모든 테스트 통과 확인
   - 코드 스타일 검사

4. **PR 생성** (팀 모드)

   - develop 브랜치 대상 PR 생성
   - 변경사항 요약
   - 검증 결과 포함

### 동기화 과정

```
/alfred:3-sync auto SPEC-001

➡️ 1단계: SPEC 검증
✅ SPEC-001 구조 정상
✅ 요구사항 8개 확인

➡️ 2단계: TAG 검증
✅ @TEST:SPEC-001 태그 12개
✅ @CODE:SPEC-001 태그 15개
✅ @DOC:SPEC-001 태그 3개
⚠️ 고아 TAG 2개 제거됨

➡️ 3단계: 품질 검증
✅ 테스트 커버리지: 92%
✅ 모든 테스트 통과
✅ 코드 스타일 정상

➡️ 4단계: 문서 생성
✅ API 문서 생성: docs/api/SPEC-001.md
✅ 아키텍처 다이어그램 생성

➡️ 5단계: PR 생성
✅ PR #23 생성
📝 제목: "feat: SPEC-001 로그인 기능 구현"
```

______________________________________________________________________

## 5. /alfred:9-feedback

**GitHub Issue 생성 (피드백)**

### 문법

```
/alfred:9-feedback
```

### 기능

- 🐛 버그 보고
- 💡 기능 제안
- 📝 개선사항
- ❓ 질문

### 상호작용 예시

```
/alfred:9-feedback

> Alfred: 피드백 유형?
선택: [1] 버그  [2] 기능  [3] 개선  [4] 질문

> 선택: 1

> 제목?
입력: "로그인 후 세션 유지 안됨"

> 설명?
입력: "로그인 후 새로고침하면 로그아웃됨"

> 재현 단계?
입력: "1. 로그인 2. 새로고침 3. 대시보드 접근"

> 기대 동작?
입력: "세션 유지되어야 함"

✅ GitHub Issue #24 생성됨
📝 제목: "🐛 로그인 후 세션 유지 안됨"
```

______________________________________________________________________

## 명령어 빠른 참고

### 완전한 워크플로우

```bash
# 1. 프로젝트 초기화
/alfred:0-project

# 2. SPEC 작성
/alfred:1-plan "로그인 기능" "회원가입"

# 3. TDD 구현
/alfred:2-run all

# 4. 동기화 및 검증
/alfred:3-sync auto all

# 완료! 자동으로 PR 생성됨
```

### 부분 워크플로우

```bash
# 특정 SPEC만 수정
/alfred:1-plan SPEC-001 "로그인 기능 (OAuth 추가)"

# 그 SPEC만 개발
/alfred:2-run SPEC-001

# 그 SPEC만 동기화
/alfred:3-sync auto SPEC-001
```

______________________________________________________________________

## 오류 처리

### "Alfred 명령어를 인식하지 못함"

```bash
# Claude Code 재시작
exit

# 새 세션 시작
claude

# 프로젝트 재초기화
/alfred:0-project
```

### "SPEC 파일을 찾을 수 없음"

```bash
# 프로젝트 상태 확인
moai-adk status

# 재초기화
moai-adk init . --force
/alfred:0-project
```

### "테스트 커버리지 부족"

```bash
# 현재 커버리지 확인
moai-adk status

# 누락된 테스트 추가
# tests/ 디렉토리에 테스트 추가

# 다시 동기화
/alfred:3-sync auto SPEC-001
```

______________________________________________________________________

**다음**: [moai-adk 명령어 참고서](moai-adk.md) 또는 [Alfred 개념](../../guides/alfred/index.md)
