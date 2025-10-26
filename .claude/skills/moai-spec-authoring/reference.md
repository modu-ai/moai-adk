# SPEC Authoring Reference

## 메타데이터 완전 레퍼런스

### 7개 필수 필드

#### 1. `id` – 고유 SPEC 식별자

**형식**: `<DOMAIN>-<NUMBER>`

**규칙**:
- 할당 후 불변
- 대문자 도메인 사용 (예: `AUTH`, `PAYMENT`, `CONFIG`)
- 세 자리 숫자 (001–999)
- 중복 확인: `rg "@SPEC:AUTH-001" -n .moai/specs/`

**예시**:
- `AUTH-001` (인증 기능)
- `INSTALLER-SEC-001` (설치 프로그램 보안)
- `TRUST-001` (TRUST 원칙)
- `CONFIG-001` (설정 스키마)

**디렉토리 구조**:
```
.moai/specs/SPEC-AUTH-001/
  ├── spec.md          # 메인 SPEC 문서
  ├── diagrams/        # 선택: 아키텍처 다이어그램
  └── examples/        # 선택: 코드 예제
```

#### 2. `version` – 시맨틱 버전

**형식**: `MAJOR.MINOR.PATCH`

**생애주기**:

| Version | Status | Description | Trigger |
|---------|--------|-------------|---------|
| `0.0.1` | draft | 초안 | SPEC 생성 |
| `0.0.x` | draft | 초안 개선 | 콘텐츠 편집 |
| `0.1.0` | completed | 구현 완료 | TDD 후 `/alfred:3-sync` |
| `0.1.x` | completed | 버그 수정, 문서 업데이트 | 구현 후 패치 |
| `0.x.0` | completed | 기능 추가 | 마이너 개선 |
| `1.0.0` | completed | 프로덕션 안정 | 이해관계자 승인 |

**버전 업데이트 예시**:
```markdown
## HISTORY

### v0.2.0 (2025-11-15)
- **ADDED**: 다중 인증 지원
- **CHANGED**: 토큰 만료 시간 30분으로 연장
- **AUTHOR**: @YourHandle

### v0.1.0 (2025-10-30)
- **COMPLETED**: TDD 구현 완료
- **EVIDENCE**: Commits 4c66076, 34e1bd9
- **TEST COVERAGE**: 89.13%

### v0.0.2 (2025-10-25)
- **REFINED**: 비밀번호 재설정 플로우 요구사항 추가
- **AUTHOR**: @YourHandle

### v0.0.1 (2025-10-23)
- **INITIAL**: JWT 인증 SPEC 초안 생성
```

#### 3. `status` – 진행 상태

**값**: `draft` | `active` | `completed` | `deprecated`

**생애주기 흐름**:
```
draft → active → completed → [deprecated]
  ↓       ↓          ↓
/alfred:1-plan  /alfred:2-run  /alfred:3-sync
```

**전환**:
- `draft`: 작성 중 (v0.0.x)
- `active`: 구현 진행 중 (v0.0.x → v0.1.0)
- `completed`: 구현 완료 (v0.1.0+)
- `deprecated`: 폐기 예정

#### 4. `created` – 생성일

**형식**: `YYYY-MM-DD`

**규칙**:
- 한 번 설정, 변경 없음
- ISO 8601 날짜 형식
- 첫 초안 날짜

**예시**: `created: 2025-10-23`

#### 5. `updated` – 최종 수정일

**형식**: `YYYY-MM-DD`

**규칙**:
- 모든 콘텐츠 변경 시 업데이트
- 초기에는 `created`와 동일
- 최신 편집 날짜 반영

**업데이트 패턴**:
```yaml
created: 2025-10-23   # 변경 없음
updated: 2025-10-25   # 편집 시 변경
```

#### 6. `author` – 주 작성자

**형식**: `@{GitHubHandle}`

**규칙**:
- 단일 값 (배열 아님)
- `@` 접두사 필수
- 대소문자 구분 (예: `@Goos`, `@goos` 아님)
- 추가 기여자는 HISTORY 섹션에 기록

**예시**:
```yaml
# 올바름
author: @Goos

# 잘못됨
author: goos           # @ 누락
authors: [@Goos]       # 배열 불가
author: @goos          # 대소문자 오류
```

#### 7. `priority` – 작업 우선순위

**값**: `critical` | `high` | `medium` | `low`

**가이드라인**:

| Priority | Description | Examples |
|----------|-------------|----------|
| `critical` | 프로덕션 차단, 보안 취약점 | 보안 패치, 치명적 버그 |
| `high` | 주요 기능, 핵심 기능 | 인증, 결제 시스템 |
| `medium` | 개선, 향상 | UI 폴리시, 성능 최적화 |
| `low` | 있으면 좋음, 문서화 | README 업데이트, 마이너 리팩토링 |

---

### 9개 선택 필드

#### 8. `category` – 변경 유형

**값**: `feature` | `bugfix` | `refactor` | `security` | `docs` | `perf`

**용도**:
```yaml
category: feature       # 새 기능
category: bugfix        # 결함 해결
category: refactor      # 코드 구조 개선
category: security      # 보안 강화
category: docs          # 문서 업데이트
category: perf          # 성능 최적화
```

#### 9. `labels` – 분류 태그

**형식**: 문자열 배열

**목적**: 검색, 필터링, 그룹화

**모범 사례**:
- 소문자, kebab-case 사용
- SPEC당 2-5개 레이블
- `category`와 중복 방지

**예시**:
```yaml
labels:
  - authentication
  - jwt
  - security

labels:
  - performance
  - optimization
  - caching

labels:
  - installer
  - template
  - cross-platform
```

#### 10-13. 관계 필드 (의존성 그래프)

##### `depends_on` – 필수 SPEC

**의미**: 먼저 완료되어야 하는 SPEC

**예시**:
```yaml
depends_on:
  - USER-001      # 사용자 모델 SPEC
  - TOKEN-001     # 토큰 생성 SPEC
```

**사용 사례**: 실행 순서, 병렬화 결정

##### `blocks` – 차단된 SPEC

**의미**: 이 SPEC이 해결될 때까지 진행할 수 없는 SPEC

**예시**:
```yaml
blocks:
  - AUTH-002      # OAuth 통합은 기본 인증을 기다림
  - PAYMENT-001   # 결제는 인증 필요
```

##### `related_specs` – 관련 SPEC

**의미**: 직접 의존성 없는 관련 항목

**예시**:
```yaml
related_specs:
  - SESSION-001   # 세션 관리 (관련되지만 독립적)
  - AUDIT-001     # 감사 로깅 (횡단 관심사)
```

##### `related_issue` – 연결된 GitHub 이슈

**형식**: 전체 GitHub 이슈 URL

**예시**:
```yaml
related_issue: "https://github.com/modu-ai/moai-adk/issues/42"
```

#### 14-15. 범위 필드 (영향 분석)

##### `scope.packages` – 영향받는 패키지

**목적**: 어떤 패키지/모듈이 영향받는지 추적

**예시**:
```yaml
scope:
  packages:
    - src/core/auth
    - src/core/token
    - src/api/routes/auth
```

##### `scope.files` – 주요 파일

**목적**: 주요 구현 파일 참조

**예시**:
```yaml
scope:
  files:
    - auth-service.ts
    - token-manager.ts
    - auth.routes.ts
```

---

## EARS 요구사항 문법

### 5가지 EARS 패턴

EARS (Easy Approach to Requirements Syntax)는 친숙한 키워드를 사용하여 체계적이고 테스트 가능한 요구사항을 제공합니다.

#### 패턴 1: Ubiquitous Requirements

**템플릿**: `시스템은 [능력]해야 한다.`

**목적**: 항상 활성화된 기본 기능

**특성**:
- 전제 조건 없음
- 항상 적용 가능
- 핵심 기능 정의

**예시**:
```markdown
**UR-001**: 시스템은 사용자 인증을 제공해야 한다.

**UR-002**: 시스템은 HTTPS 연결을 지원해야 한다.

**UR-003**: 시스템은 사용자 인증 정보를 안전하게 저장해야 한다.

**UR-004**: 모바일 앱의 크기는 50 MB 미만이어야 한다.

**UR-005**: API 응답 시간은 요청의 95%에 대해 200ms를 초과하지 않아야 한다.
```

**모범 사례**:
- ✅ 능동태 사용
- ✅ 요구사항당 단일 책임
- ✅ 측정 가능한 결과
- ❌ 모호한 용어 피하기 ("사용자 친화적", "빠른")

#### 패턴 2: Event-driven Requirements

**템플릿**: `WHEN [트리거], 시스템은 [응답]해야 한다.`

**목적**: 특정 이벤트에 의해 트리거되는 동작 정의

**특성**:
- 개별 이벤트에 의해 트리거
- 일회성 응답
- 원인-결과 관계

**예시**:
```markdown
**ER-001**: WHEN 사용자가 유효한 인증 정보를 제출하면, 시스템은 JWT 토큰을 발급해야 한다.

**ER-002**: WHEN 토큰이 만료되면, 시스템은 HTTP 401 Unauthorized를 반환해야 한다.

**ER-003**: WHEN 사용자가 "비밀번호 찾기"를 클릭하면, 시스템은 비밀번호 재설정 이메일을 보내야 한다.

**ER-004**: WHEN 데이터베이스 연결이 실패하면, 시스템은 지수 백오프로 3번 재시도해야 한다.

**ER-005**: WHEN 파일 업로드가 10 MB를 초과하면, 시스템은 에러 메시지와 함께 업로드를 거부해야 한다.
```

**고급 패턴** (사후 조건 포함):
```markdown
**ER-006**: WHEN 결제 거래가 완료되면, 시스템은 확인 이메일을 보낸 다음 주문 상태를 "paid"로 업데이트해야 한다.
```

**모범 사례**:
- ✅ 요구사항당 단일 트리거
- ✅ 구체적이고 테스트 가능한 응답
- ✅ 에러 조건 포함
- ❌ 여러 WHEN 연결하지 않기

#### 패턴 3: State-driven Requirements

**템플릿**: `WHILE [상태], 시스템은 [동작]해야 한다.`

**목적**: 상태 중 지속적인 동작

**특성**:
- 상태가 지속되는 동안 활성
- 지속적인 모니터링
- 상태 의존적 동작

**예시**:
```markdown
**SR-001**: WHILE 사용자가 인증된 상태이면, 시스템은 보호된 라우트에 대한 접근을 허용해야 한다.

**SR-002**: WHILE 토큰이 유효하면, 시스템은 토큰 클레임에서 사용자 ID를 추출해야 한다.

**SR-003**: WHILE 시스템이 유지보수 모드이면, 시스템은 HTTP 503 Service Unavailable을 반환해야 한다.

**SR-004**: WHILE 배터리 잔량이 20% 미만이면, 모바일 앱은 백그라운드 동기화 빈도를 줄여야 한다.

**SR-005**: WHILE 파일 업로드가 진행 중이면, UI는 진행 표시줄을 표시해야 한다.
```

**모범 사례**:
- ✅ 상태 경계 명확히 정의
- ✅ 상태 진입/종료 조건 명시
- ✅ 상태 전환 테스트
- ❌ 겹치는 상태 피하기

#### 패턴 4: Optional Features

**템플릿**: `WHERE [기능], 시스템은 [동작]할 수 있다.`

**목적**: 기능 플래그 기반 조건부 기능

**특성**:
- 기능이 존재하는 경우에만 적용
- 구성 의존적
- 제품 변형 지원

**예시**:
```markdown
**OF-001**: WHERE 다중 인증이 활성화된 경우, 시스템은 OTP 검증을 요구할 수 있다.

**OF-002**: WHERE 세션 로깅이 활성화된 경우, 시스템은 로그인 타임스탬프를 기록할 수 있다.

**OF-003**: WHERE 프리미엄 구독이 활성화된 경우, 시스템은 무제한 API 호출을 허용할 수 있다.

**OF-004**: WHERE 다크 모드가 선택된 경우, UI는 어두운 색 구성표로 렌더링할 수 있다.

**OF-005**: WHERE 분석 동의가 부여된 경우, 시스템은 사용자 동작을 추적할 수 있다.
```

**모범 사례**:
- ✅ "할 수 있다(can)" 사용 (허용), "해야 한다(shall)" 아님 (필수)
- ✅ 기능 플래그 조건 명확히 정의
- ✅ 기능 없을 때 기본 동작 명시
- ❌ 핵심 기능을 선택 사항으로 만들지 않기

#### 패턴 5: Constraints

**템플릿**: `IF [조건], THEN 시스템은 [제약]해야 한다.`

**목적**: 품질 속성 및 비즈니스 규칙 강제

**특성**:
- 조건부 강제
- 품질 게이트
- 비즈니스 규칙 검증

**예시**:
```markdown
**C-001**: IF 토큰이 만료되었다면, THEN 시스템은 접근을 거부하고 HTTP 401을 반환해야 한다.

**C-002**: IF 10분 이내에 5번 이상 로그인 실패가 발생하면, THEN 시스템은 임시로 계정을 잠가야 한다.

**C-003**: 액세스 토큰은 15분 수명을 초과하지 않아야 한다.

**C-004**: IF 비밀번호가 8자 미만이면, THEN 시스템은 등록을 거부해야 한다.

**C-005**: IF API 속도 제한이 초과되면, THEN 시스템은 HTTP 429 Too Many Requests를 반환해야 한다.
```

**단순화된 제약** (조건 없음):
```markdown
**C-006**: 시스템은 평문 비밀번호를 저장하지 않아야 한다.

**C-007**: /health 및 /login을 제외한 모든 API 엔드포인트는 인증을 요구해야 한다.
```

**모범 사례**:
- ✅ 엄격한 제약에는 SHALL, 부드러운 권장사항에는 SHOULD 사용
- ✅ 제한 정량화 (시간, 크기, 횟수)
- ✅ 강제 메커니즘 명시
- ❌ 모호한 제약 피하기

---

## EARS 패턴 선택 가이드

| 패턴 | 키워드 | 사용 시기 | 예시 컨텍스트 |
|---------|---------|----------|-----------------|
| **Ubiquitous** | shall | 핵심 기능, 항상 활성 | "시스템은 로그인을 제공해야 한다" |
| **Event-driven** | WHEN | 개별 이벤트에 대한 응답 | "WHEN 로그인 실패 시, 에러 표시" |
| **State-driven** | WHILE | 상태 중 지속적 동작 | "WHILE 로그인 상태, 접근 허용" |
| **Optional** | WHERE | 기능 플래그 또는 구성 | "WHERE 프리미엄이면, 기능 해제" |
| **Constraints** | IF-THEN | 품질 게이트, 비즈니스 규칙 | "IF 만료되었다면, 거부" |

---

## HISTORY 섹션 형식

HISTORY 섹션은 모든 SPEC 버전과 변경 사항을 문서화합니다.

### 구조

```markdown
## HISTORY

### v{MAJOR}.{MINOR}.{PATCH} ({YYYY-MM-DD})
- **{CHANGE_TYPE}**: {설명}
- **AUTHOR**: {GitHub 핸들}
- **{추가 컨텍스트}**: {세부사항}
```

### 변경 유형

| Type | Description | Example |
|------|-------------|---------|
| **INITIAL** | 첫 초안 | `v0.0.1: INITIAL 초안 생성` |
| **REFINED** | 초안 중 콘텐츠 업데이트 | `v0.0.2: REFINED 검토 기반 요구사항` |
| **COMPLETED** | 구현 완료 | `v0.1.0: COMPLETED TDD 구현` |
| **ADDED** | 새 요구사항/기능 | `v0.2.0: ADDED 다중 인증` |
| **CHANGED** | 수정된 요구사항 | `v0.2.0: CHANGED 토큰 만료 15분→30분` |
| **FIXED** | 구현 후 버그 수정 | `v0.1.1: FIXED 토큰 갱신 로직` |
| **DEPRECATED** | 폐기 예정 표시 | `v1.5.0: DEPRECATED 레거시 인증 엔드포인트` |

### 완전한 HISTORY 예시

```markdown
## HISTORY

### v0.2.0 (2025-11-15)
- **ADDED**: OTP를 통한 다중 인증 지원
- **CHANGED**: 사용자 피드백 기반으로 토큰 만료 30분으로 연장
- **AUTHOR**: @Goos
- **REVIEWER**: @SecurityTeam
- **RATIONALE**: 보안 태세를 유지하면서 UX 개선

### v0.1.1 (2025-11-01)
- **FIXED**: 토큰 갱신 경쟁 조건
- **EVIDENCE**: Commit 3f9a2b7
- **AUTHOR**: @Goos

### v0.1.0 (2025-10-30)
- **COMPLETED**: TDD 구현 완료
- **AUTHOR**: @Goos
- **EVIDENCE**: Commits 4c66076, 34e1bd9, 1dec08f
- **TEST COVERAGE**: 89.13% (목표: 85%)
- **QUALITY METRICS**:
  - Test Pass Rate: 100% (42/42 tests)
  - Linting: ruff ✅
  - Type Checking: mypy ✅
- **TAG CHAIN**:
  - @SPEC:AUTH-001: 1 occurrence
  - @TEST:AUTH-001: 8 occurrences
  - @CODE:AUTH-001: 12 occurrences

### v0.0.2 (2025-10-25)
- **REFINED**: 비밀번호 재설정 플로우 요구사항 추가
- **REFINED**: 토큰 수명 제약 명확화
- **AUTHOR**: @Goos

### v0.0.1 (2025-10-23)
- **INITIAL**: JWT 인증 SPEC 초안 생성
- **AUTHOR**: @Goos
- **SCOPE**: 사용자 인증, 토큰 생성, 토큰 검증
- **CONTEXT**: 2025년 4분기 제품 로드맵 요구사항
```

---

## TAG 통합

### TAG 블록 형식

모든 SPEC 문서는 타이틀 다음에 TAG 블록으로 시작합니다:

```markdown
# @SPEC:AUTH-001: JWT Authentication System
```

### TAG 체인 참조

SPEC에서 관련 TAG 연결:

```markdown
## Traceability (@TAG Chain)

### TAG 체인 구조
```
@SPEC:AUTH-001 (이 문서)
  ↓
@TEST:AUTH-001 (tests/auth/service.test.ts)
  ↓
@CODE:AUTH-001 (src/auth/service.ts, src/auth/token-manager.ts)
  ↓
@DOC:AUTH-001 (docs/api/authentication.md)
```

### 검증 명령어
```bash
# SPEC TAG 검증
rg '@SPEC:AUTH-001' -n .moai/specs/

# 중복 ID 확인
rg '@SPEC:AUTH' -n .moai/specs/
rg 'AUTH-001' -n

# 전체 TAG 체인 스캔
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n
```
```

---

## 검증 명령어

### 빠른 검증 스크립트

```bash
#!/usr/bin/env bash
# validate-spec.sh - SPEC 검증 헬퍼

SPEC_DIR="$1"

echo "SPEC 검증 중: $SPEC_DIR"

# 필수 필드 확인
echo -n "필수 필드... "
rg "^(id|version|status|created|updated|author|priority):" "$SPEC_DIR/spec.md" | wc -l | grep -q "7" && echo "✅" || echo "❌"

# 작성자 형식 확인
echo -n "작성자 형식... "
rg "^author: @[A-Z]" "$SPEC_DIR/spec.md" > /dev/null && echo "✅" || echo "❌"

# 버전 형식 확인
echo -n "버전 형식... "
rg "^version: 0\.\d+\.\d+" "$SPEC_DIR/spec.md" > /dev/null && echo "✅" || echo "❌"

# HISTORY 섹션 확인
echo -n "HISTORY 섹션... "
rg "^## HISTORY" "$SPEC_DIR/spec.md" > /dev/null && echo "✅" || echo "❌"

# TAG 블록 확인
echo -n "TAG 블록... "
rg "^# @SPEC:" "$SPEC_DIR/spec.md" > /dev/null && echo "✅" || echo "❌"

# 중복 ID 확인
SPEC_ID=$(basename "$SPEC_DIR" | sed 's/SPEC-//')
DUPLICATE_COUNT=$(rg "@SPEC:$SPEC_ID" -n .moai/specs/ | wc -l)
echo -n "중복 ID 확인... "
[ "$DUPLICATE_COUNT" -eq 1 ] && echo "✅" || echo "❌ (found $DUPLICATE_COUNT occurrences)"

echo "검증 완료!"
```

### 사용법

```bash
# 단일 SPEC 검증
./validate-spec.sh .moai/specs/SPEC-AUTH-001

# 모든 SPEC 검증
for spec in .moai/specs/SPEC-*/; do
  ./validate-spec.sh "$spec"
done
```

---

**Last Updated**: 2025-10-27
**Version**: 1.1.0
