# 핵심 Sub-agents 상세 가이드

Alfred의 10명 핵심 에이전트에 대한 완전한 참고서입니다.

## 개요

| #   | 에이전트               | 역할            | 스킬 수 | 최적 사이즈     |
| --- | ---------------------- | --------------- | ------- | --------------- |
| 1   | project-manager        | 프로젝트 초기화 | 5개     | 1-person 팀     |
| 2   | spec-builder           | SPEC 작성       | 8개     | 모든 팀         |
| 3   | implementation-planner | 계획 수립       | 6개     | 팀 프로젝트     |
| 4   | tdd-implementer        | TDD 실행        | 12개    | 모든 팀         |
| 5   | doc-syncer             | 문서 동기화     | 8개     | 모든 팀         |
| 6   | tag-agent              | TAG 검증        | 4개     | 중대형 프로젝트 |
| 7   | git-manager            | Git 자동화      | 10개    | 모든 팀         |
| 8   | trust-checker          | 품질 검증       | 7개     | 릴리즈 단계     |
| 9   | quality-gate           | 릴리즈 준비     | 6개     | 프로덕션        |
| 10  | debug-helper           | 오류 해결       | 9개     | 문제 발생 시    |

______________________________________________________________________

## 1. project-manager

**역할**: 프로젝트 초기화 및 메타데이터 관리

### 활성화 조건

```
/alfred:0-project [setting|update]
```

### 주요 책임

- 프로젝트 메타데이터 설정 (이름, 설명, 팀 크기)
- 대화 언어 선택 및 적용
- 개발 모드 결정 (solo/team/org)
- `.moai/config.json` 초기화
- TRUST 5 원칙 기본 설정

### 상호작용 형식

```
User: /alfred:0-project

Alfred: 프로젝트 이름?
→ project-manager: 입력값 검증 및 저장

Alfred: 개발 모드?
→ project-manager: 팀 크기에 따른 설정 결정

Alfred: 대화 언어?
→ project-manager: 모든 후속 커뮤니케이션 언어 설정

Result: .moai/config.json 생성 완료
```

### 생성 파일 구조

```
.moai/
├── config.json           # 프로젝트 설정
├── specs/               # SPEC 저장 디렉토리
├── docs/                # 생성 문서
├── reports/             # 분석 보고서
└── scripts/             # 유틸리티
```

### 사용 시나리오

- **새 프로젝트 시작**: 처음 Alfred 초기화
- **설정 변경**: 언어, 팀 모드, 테스트 커버리지 목표 수정
- **여러 프로젝트**: 각 프로젝트별 독립적 설정

______________________________________________________________________

## 2. spec-builder

**역할**: EARS 형식의 SPEC 문서 작성

### 활성화 조건

```
/alfred:1-plan "제목1" "제목2" ...
/alfred:1-plan SPEC-ID "수정사항"
```

### 주요 책임

- 사용자 요구사항을 EARS 형식으로 구조화
- SPEC ID 자동 생성 (SPEC-001, SPEC-002...)
- 요구사항 명확성 검증
- 테스트 계획 초안 작성
- 구현 범위 정의

### EARS 형식 구조

```
GIVEN:     초기 상황 설명
WHEN:      사용자 액션
THEN:      예상 결과
```

### 예시

**사용자 입력**:

```
/alfred:1-plan "사용자 인증 시스템"
```

**생성되는 SPEC**:

```markdown
# SPEC-001: 사용자 인증 시스템

## 요구사항

### 로그인 기능
- GIVEN: 사용자가 로그인 페이지 방문
  WHEN: 유효한 이메일과 비밀번호 입력
  THEN: 세션 생성 및 대시보드 리다이렉트

### 비밀번호 오류 처리
- GIVEN: 로그인 페이지
  WHEN: 잘못된 비밀번호 입력
  THEN: "비밀번호 오류" 메시지 표시

## 테스트 계획
- [ ] 정상 로그인
- [ ] 비밀번호 오류
- [ ] 계정 잠금 (5회 실패)
```

### 품질 기준

- 명확한 요구사항 (5개 이상)
- 모호하지 않은 표현
- 테스트 가능한 조건
- 구현 가능한 범위

______________________________________________________________________

## 3. implementation-planner

**역할**: 아키텍처 및 실행 계획 수립

### 활성화 조건

```
/alfred:2-run SPEC-ID (시작 시)
```

### 주요 책임

- SPEC을 구현 단계로 분해
- 파일 및 디렉토리 구조 설계
- 작업 의존성 분석
- 병렬 실행 기회 식별
- 예상 시간 및 난도 추정

### 계획 수립 프로세스

```
SPEC 분석
    ↓
작업 분해 (5-10 단계)
    ↓
의존성 맵핑
    ↓
병렬화 기회 식별
    ↓
영향 파일 목록화
    ↓
시간 추정
    ↓
사용자 승인 요청
```

### 계획 문서 예시

```
SPEC-001: 사용자 인증 시스템

📋 작업 분해:
1. 데이터 모델 설계 (User, Session)
2. 데이터베이스 스키마 생성
3. 비밀번호 해싱 함수 구현
4. 로그인 엔드포인트 구현
5. 세션 관리 미들웨어 작성
6. 로그아웃 엔드포인트
7. 비밀번호 재설정
8. 계정 잠금 메커니즘

🔄 의존성:
1 → 2 → 3 → 4
     ↓
     5 → 6, 7 → 8

⚡ 병렬화:
- 4와 5는 병렬 가능
- 6, 7, 8은 병렬 가능

📁 영향 파일:
- models/user.py (NEW)
- models/session.py (NEW)
- api/auth.py (NEW)
- middleware/session.py (NEW)
- tests/test_auth.py (NEW)
- docs/auth.md (NEW)

⏱️ 예상 시간: 2시간 (3단계: RED/GREEN/REFACTOR)
```

______________________________________________________________________

## 4. tdd-implementer

**역할**: RED-GREEN-REFACTOR 사이클 실행

### 활성화 조건

```
/alfred:2-run SPEC-ID (중 실행)
```

### 주요 책임

- RED 단계: 실패 테스트 작성
- GREEN 단계: 최소 구현
- REFACTOR 단계: 코드 품질 개선
- 각 단계 완료 후 TodoWrite 업데이트
- 테스트 상태 추적

### TDD 3단계 구현

#### Phase 1: RED (빨강)

```python
# 실패하는 테스트만 작성
def test_user_registration():
    user = register_user("user@example.com", "password123")
    assert user.email == "user@example.com"
    assert user.is_verified == False

# 실행 → FAIL ❌
```

#### Phase 2: GREEN (초록)

```python
# 최소 구현
def register_user(email, password):
    user = User(email=email)
    user.set_password(password)
    db.session.add(user)
    db.session.commit()
    return user

# 실행 → PASS ✅
```

#### Phase 3: REFACTOR (리팩토)

```python
# 코드 품질 개선 (테스트는 유지)
def register_user(email, password):
    """사용자 등록"""
    # 입력 검증
    if not is_valid_email(email):
        raise ValueError("Invalid email")
    if len(password) < 8:
        raise ValueError("Password too short")

    # 중복 확인
    if User.query.filter_by(email=email).first():
        raise ValueError("User already exists")

    # 사용자 생성
    user = User(email=email)
    user.set_password(password)
    db.session.add(user)
    db.session.commit()

    return user
```

### TodoWrite 추적

```
[in_progress] RED: SPEC-001 테스트 작성
[completed]   RED: SPEC-001 테스트 작성
[in_progress] GREEN: SPEC-001 최소 구현
[completed]   GREEN: SPEC-001 최소 구현
[in_progress] REFACTOR: SPEC-001 코드 개선
[completed]   REFACTOR: SPEC-001 코드 개선
```

______________________________________________________________________

## 5. doc-syncer

**역할**: 문서 자동 생성 및 동기화

### 활성화 조건

```
/alfred:3-sync auto [SPEC-ID]
```

### 주요 책임

- API 문서 자동 생성 (OpenAPI/Swagger)
- 아키텍처 다이어그램 생성
- 배포 가이드 작성
- 변경사항 요약 문서 생성
- 문서 링크 검증

### 생성 문서 종류

| 문서         | 내용                | 형식        |
| ------------ | ------------------- | ----------- |
| API Spec     | RESTful 엔드포인트  | OpenAPI 3.1 |
| Architecture | 시스템 다이어그램   | Mermaid     |
| Deployment   | 배포 절차           | Markdown    |
| Changelog    | 변경사항            | Markdown    |
| Migration    | 데이터 마이그레이션 | SQL + 설명  |

### 생성 위치

```
docs/
├── api/
│   └── SPEC-001.md          # API 문서
├── architecture/
│   └── SPEC-001.md          # 아키텍처
├── deployment/
│   └── SPEC-001.md          # 배포 가이드
├── migrations/
│   └── 001_create_users.sql # 마이그레이션
└── changelog/
    └── v1.0.0.md            # 변경사항
```

______________________________________________________________________

## 6. tag-agent

**역할**: TAG 검증 및 추적성 관리

### 활성화 조건

```
/alfred:3-sync auto [SPEC-ID]
```

### 주요 책임

- SPEC → TEST → CODE → DOC TAG 체인 검증
- 고아 TAG 탐지 및 제거
- TAG 명명 규칙 검증
- 추적성 무결성 확인

### TAG 체인

```
SPEC-001 (요구사항)
    ↓
@TEST:SPEC-001:* (테스트)
    ↓
@CODE:SPEC-001:* (구현)
    ↓
@DOC:SPEC-001:* (문서)
    ↓
상호 참조 (완전한 추적성)
```

### 예시

```python
# @CODE:SPEC-001:register_user
def register_user(email: str, password: str) -> User:
    """사용자 등록"""
    # @CODE:SPEC-001:validate_email
    if not is_valid_email(email):
        raise ValueError("Invalid email")

    # @CODE:SPEC-001:hash_password
    hashed = hash_password(password)

    # @CODE:SPEC-001:create_user
    user = User(email=email, password_hash=hashed)
    db.session.add(user)
    db.session.commit()

    return user

# @TEST:SPEC-001:test_register_success
def test_register_success():
    user = register_user("test@example.com", "password123")
    assert user.email == "test@example.com"
```

______________________________________________________________________

## 7. git-manager

**역할**: Git 워크플로우 자동화

### 활성화 조건

모든 단계에서 자동 활성화

### 주요 책임

- 기능 브랜치 생성 (feature/SPEC-001)
- 커밋 메시지 자동 생성
- RED/GREEN/REFACTOR 단계별 커밋
- PR 생성 및 관리
- 병합 전 검증

### Git 워크플로우

```
main
    ↓
develop (베이스 브랜치)
    ↓
feature/SPEC-001 (기능 브랜치)
    │
    ├── feat: RED phase (커밋)
    ├── feat: GREEN phase (커밋)
    ├── refactor: code quality (커밋)
    │
    ↓
PR #23 (develop ← feature/SPEC-001)
    ├── 테스트 검증
    ├── 코드 리뷰
    └── 병합
    ↓
develop (병합 완료)
    ↓
main (릴리즈 시)
```

### 커밋 메시지 형식

```
<type>: <description>

🤖 Claude Code로 생성됨

Co-Authored-By: 🎩 Alfred@MoAI
```

**타입**:

- `feat`: 새로운 기능
- `fix`: 버그 수정
- `refactor`: 코드 개선
- `test`: 테스트 추가
- `docs`: 문서 업데이트

______________________________________________________________________

## 8. trust-checker

**역할**: TRUST 5 원칙 검증

### 활성화 조건

```
/alfred:2-run SPEC-ID (완료 후)
```

### TRUST 5 원칙

| 원칙           | 설명             | 검증           |
| -------------- | ---------------- | -------------- |
| **T**est First | 테스트 주도 개발 | 커버리지 85%+  |
| **R**eadable   | 읽기 쉬운 코드   | Linting 통과   |
| **U**nified    | 일관된 구조      | 명명 규칙 준수 |
| **S**ecured    | 보안             | 보안 스캔 통과 |
| **T**rackable  | 추적성           | TAG 무결성     |

### 검증 결과

```
✅ Test First: 92% 커버리지 (목표: 85%)
✅ Readable: MyPy 완료, ruff 통과
✅ Unified: 명명 규칙 준수
✅ Secured: 의존성 보안 검사 통과
✅ Trackable: TAG 12개 검증

🎯 TRUST 5 준수: PASS ✅
```

______________________________________________________________________

## 9. quality-gate

**역할**: 릴리즈 준비 상태 확인

### 활성화 조건

```
/alfred:3-sync auto all (최종 단계)
```

### 검증 항목

- ✅ 모든 SPEC 완료
- ✅ 테스트 커버리지 85% 이상
- ✅ 모든 테스트 통과
- ✅ 보안 취약점 0개
- ✅ 문서 완성도 100%
- ✅ TAG 무결성

### 릴리즈 결정

```
모든 항목 통과 → PR Merge → Release 가능

실패 항목 존재 → 상세 보고서 → 개선 필요
```

______________________________________________________________________

## 10. debug-helper

**역할**: 오류 분석 및 자동 수정

### 활성화 조건

```
오류 또는 예외 발생 시 자동 활성화
```

### 주요 책임

- 오류 스택 트레이스 분석
- 근본 원인 파악
- 해결 방법 제시
- 자동 수정 가능 여부 판단
- 임시 우회 방안 제시

### 오류 처리 프로세스

```
오류 발생
    ↓
debug-helper: 분석
    ├─ 타입 파악
    ├─ 원인 추적
    ├─ 유사 사례 검색
    └─ 해결책 제시
    ↓
[자동 수정 가능?]
    ├─ YES → 수정 및 재실행
    └─ NO → 상세 가이드 제시
```

______________________________________________________________________

## 에이전트 간 협력 사례

### 완전한 워크플로우 예시

```
SPEC-001 작성 (spec-builder)
    ↓
구현 계획 (implementation-planner)
    ↓
RED 단계 테스트 (tdd-implementer)
    ↓
GREEN 단계 구현 (tdd-implementer)
    ↓
REFACTOR 단계 (tdd-implementer)
    ↓
TRUST 5 검증 (trust-checker)
    ↓
Git 커밋 (git-manager)
    ↓
문서 생성 (doc-syncer)
    ↓
TAG 검증 (tag-agent)
    ↓
릴리즈 준비 (quality-gate)
    ↓
완료!
```

______________________________________________________________________

**다음**: [전문가 Agents](experts.md) 또는 [Agents 개요](index.md)
