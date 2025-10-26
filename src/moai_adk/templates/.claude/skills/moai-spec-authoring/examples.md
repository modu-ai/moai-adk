# SPEC Authoring Examples

## 실전 EARS 예제

### 예제 1: E-commerce 체크아웃

```markdown
### Ubiquitous Requirements
**UR-001**: 시스템은 장바구니를 제공해야 한다.
**UR-002**: 시스템은 신용카드 결제를 지원해야 한다.

### Event-driven Requirements
**ER-001**: WHEN 사용자가 장바구니에 아이템을 추가하면, 시스템은 장바구니 총액을 업데이트해야 한다.
**ER-002**: WHEN 결제가 성공하면, 시스템은 주문 확인 이메일을 보내야 한다.
**ER-003**: WHEN 재고가 부족하면, 시스템은 "품절" 메시지를 표시해야 한다.

### State-driven Requirements
**SR-001**: WHILE 아이템이 장바구니에 있으면, 시스템은 30분 동안 재고를 예약해야 한다.
**SR-002**: WHILE 결제가 처리 중이면, UI는 로딩 인디케이터를 표시해야 한다.

### Optional Features
**OF-001**: WHERE 특급 배송이 선택된 경우, 시스템은 특급 배송 비용을 계산할 수 있다.
**OF-002**: WHERE 선물 포장이 가능한 경우, 시스템은 선물 포장 옵션을 제공할 수 있다.

### Constraints
**C-001**: IF 장바구니 총액이 $50 미만이면, THEN 시스템은 $5 배송료를 추가해야 한다.
**C-002**: IF 3번 결제 실패 후, THEN 시스템은 주문을 1시간 동안 잠가야 한다.
**C-003**: 주문 처리 시간은 5초를 초과하지 않아야 한다.
```

### 예제 2: 모바일 앱 푸시 알림

```markdown
### Ubiquitous Requirements
**UR-001**: 앱은 푸시 알림을 지원해야 한다.
**UR-002**: 앱은 사용자가 알림을 활성화/비활성화할 수 있도록 해야 한다.

### Event-driven Requirements
**ER-001**: WHEN 새 메시지가 도착하면, 앱은 푸시 알림을 표시해야 한다.
**ER-002**: WHEN 알림을 탭하면, 앱은 메시지 화면으로 이동해야 한다.
**ER-003**: WHEN 알림 권한이 거부되면, 앱은 인앱 알림 배너를 표시해야 한다.

### State-driven Requirements
**SR-001**: WHILE 앱이 포어그라운드 상태이면, 시스템은 푸시 알림 대신 인앱 배너를 표시해야 한다.
**SR-002**: WHILE 방해 금지 모드가 활성화된 경우, 시스템은 모든 알림을 음소거해야 한다.

### Optional Features
**OF-001**: WHERE 알림 사운드가 활성화된 경우, 시스템은 알림 사운드를 재생할 수 있다.
**OF-002**: WHERE 알림 그룹화가 지원되는 경우, 시스템은 대화별로 알림을 그룹화할 수 있다.

### Constraints
**C-001**: IF 10개 이상의 알림이 대기 중이면, THEN 시스템은 이를 요약 알림으로 그룹화해야 한다.
**C-002**: 알림 전달 지연은 5초를 초과하지 않아야 한다.
```

---

## 완전한 SPEC 예제

### 예제 1: 최소 SPEC

```markdown
---
id: HELLO-001
version: 0.0.1
status: draft
created: 2025-10-23
updated: 2025-10-23
author: @Goos
priority: low
---

# @SPEC:HELLO-001: Hello World API

## HISTORY

### v0.0.1 (2025-10-23)
- **INITIAL**: Hello World API SPEC 초안 생성
- **AUTHOR**: @Goos

## Environment

**Runtime**: Node.js 20.x
**Framework**: Express.js

## Assumptions

1. 단일 엔드포인트 필요
2. 인증 불필요
3. JSON 응답 형식

## Requirements

### Ubiquitous Requirements

**UR-001**: 시스템은 GET /hello 엔드포인트를 제공해야 한다.

### Event-driven Requirements

**ER-001**: WHEN GET 요청이 /hello로 전송되면, 시스템은 JSON `{"message": "Hello, World!"}`를 반환해야 한다.

### Constraints

**C-001**: 응답 시간은 50ms를 초과하지 않아야 한다.
```

### 예제 2: 프로덕션급 SPEC

```markdown
---
id: AUTH-001
version: 0.1.0
status: completed
created: 2025-10-23
updated: 2025-10-30
author: @Goos
priority: high
category: feature
labels:
  - authentication
  - jwt
  - security
depends_on:
  - USER-001
  - TOKEN-001
blocks:
  - AUTH-002
  - PAYMENT-001
related_issue: "https://github.com/modu-ai/moai-adk/issues/42"
scope:
  packages:
    - src/core/auth
    - src/core/token
    - src/api/routes/auth
  files:
    - auth-service.ts
    - token-manager.ts
    - auth.routes.ts
---

# @SPEC:AUTH-001: JWT Authentication System

## HISTORY

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

## Environment

**Execution Context**:
- Runtime: Node.js 20.x or later
- Framework: Express.js
- Database: PostgreSQL 15+

**Technical Stack**:
- JWT library: jsonwebtoken ^9.0.0
- Hashing: bcrypt ^5.1.0

**Constraints**:
- Token lifetime: 15 minutes (access), 7 days (refresh)
- Security: HTTPS required in production

## Assumptions

1. **User Storage**: 사용자 인증 정보는 PostgreSQL에 저장
2. **Secret Management**: JWT 시크릿은 환경변수로 관리
3. **Clock Sync**: 서버 시계는 NTP로 동기화
4. **Password Policy**: 가입 시 최소 8자 강제

## Requirements

### Ubiquitous Requirements

**UR-001**: 시스템은 JWT 기반 인증을 제공해야 한다.

**UR-002**: 시스템은 이메일과 비밀번호로 사용자 로그인을 지원해야 한다.

**UR-003**: 시스템은 액세스 토큰과 리프레시 토큰을 발급해야 한다.

### Event-driven Requirements

**ER-001**: WHEN 사용자가 유효한 인증 정보를 제출하면, 시스템은 15분 만료 시간을 가진 JWT 액세스 토큰을 발급해야 한다.

**ER-002**: WHEN 토큰이 만료되면, 시스템은 HTTP 401 Unauthorized를 반환해야 한다.

**ER-003**: WHEN 리프레시 토큰이 제시되면, 시스템은 리프레시 토큰이 유효한 경우 새 액세스 토큰을 발급해야 한다.

### State-driven Requirements

**SR-001**: WHILE 사용자가 인증된 상태이면, 시스템은 보호된 리소스에 대한 접근을 허용해야 한다.

**SR-002**: WHILE 토큰이 유효하면, 시스템은 토큰 클레임에서 사용자 ID를 추출해야 한다.

### Optional Features

**OF-001**: WHERE 다중 인증이 활성화된 경우, 시스템은 비밀번호 확인 후 OTP 검증을 요구할 수 있다.

**OF-002**: WHERE 세션 로깅이 활성화된 경우, 시스템은 로그인 타임스탬프와 IP 주소를 기록할 수 있다.

### Constraints

**C-001**: IF 토큰이 만료되었다면, 시스템은 접근을 거부하고 HTTP 401을 반환해야 한다.

**C-002**: IF 10분 이내에 5번 이상 로그인 실패가 발생하면, 시스템은 임시로 계정을 잠가야 한다.

**C-003**: 액세스 토큰은 15분 수명을 초과하지 않아야 한다.

**C-004**: 리프레시 토큰은 7일 수명을 초과하지 않아야 한다.

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

## Decision Log

### Decision 1: JWT vs Session Cookies (2025-10-23)
**Context**: 마이크로서비스를 위한 무상태 인증 필요
**Decision**: JWT 토큰 사용
**Alternatives Considered**: 
  - 세션 쿠키 (거부: 상태 기반, 확장 불가)
  - OAuth 2.0 (연기: MVP에 너무 복잡)
**Consequences**: 
  - ✅ 무상태, 확장 가능
  - ✅ 서비스 간 인증
  - ❌ 토큰 취소 복잡성

### Decision 2: 토큰 만료 15분 (2025-10-24)
**Context**: 보안과 UX 균형
**Decision**: 15분 액세스 토큰, 7일 리프레시 토큰
**Rationale**: 업계 표준, 보안 모범 사례
**References**: OWASP JWT 모범 사례

## Requirements Traceability Matrix

| Req ID | Description | Test Cases | Status |
|--------|-------------|------------|--------|
| UR-001 | JWT 인증 | test_authenticate_valid_user | ✅ |
| ER-001 | 토큰 발급 | test_token_generation | ✅ |
| ER-002 | 토큰 만료 | test_expired_token_rejection | ✅ |
| SR-001 | 인증된 접근 | test_protected_route_access | ✅ |
| C-001 | 토큰 수명 | test_token_expiry_constraint | ✅ |
```

---

## 고급 패턴

### 패턴 1: 버전화된 요구사항

요구사항이 버전 간 변경될 때 진화 과정 문서화:

```markdown
### v0.2.0 (2025-11-15)
**UR-001** (CHANGED): 시스템은 요청의 99%에 대해 200ms 이내에 응답해야 한다.
  - 이전 (v0.1.0): 요청의 95%
  - 근거: 사용자 피드백 기반 성능 개선

### v0.1.0 (2025-10-30)
**UR-001**: 시스템은 요청의 95%에 대해 200ms 이내에 응답해야 한다.
```

### 패턴 2: 요구사항 추적 매트릭스

요구사항을 테스트 케이스에 명시적으로 연결:

```markdown
## Requirements Traceability Matrix

| Req ID | Description | Test Cases | Status |
|--------|-------------|------------|--------|
| UR-001 | JWT 인증 | test_authenticate_valid_user | ✅ |
| ER-001 | 토큰 발급 | test_token_generation | ✅ |
| ER-002 | 토큰 만료 | test_expired_token_rejection | ✅ |
| SR-001 | 인증된 접근 | test_protected_route_access | ✅ |
| C-001 | 토큰 수명 | test_token_expiry_constraint | ✅ |
```

### 패턴 3: 결정 로그

SPEC 내에서 아키텍처 결정 문서화:

```markdown
## Decision Log

### Decision 1: JWT vs Session Cookies (2025-10-23)
**Context**: 마이크로서비스를 위한 무상태 인증 필요
**Decision**: JWT 토큰 사용
**Alternatives Considered**: 
  - 세션 쿠키 (거부: 상태 기반, 확장 불가)
  - OAuth 2.0 (연기: MVP에 너무 복잡)
**Consequences**: 
  - ✅ 무상태, 확장 가능
  - ✅ 서비스 간 인증
  - ❌ 토큰 취소 복잡성

### Decision 2: 토큰 만료 15분 (2025-10-24)
**Context**: 보안과 UX 균형
**Decision**: 15분 액세스 토큰, 7일 리프레시 토큰
**Rationale**: 업계 표준, 보안 모범 사례
**References**: OWASP JWT 모범 사례
```

---

## 트러블슈팅

### 이슈: "중복 SPEC ID 감지됨"

**증상**: `rg "@SPEC:AUTH-001" -n`이 여러 결과 반환

**해결책**:
```bash
# 모든 발생 찾기
rg "@SPEC:AUTH-001" -n .moai/specs/

# 하나의 SPEC 유지, 나머지 이름 변경
# 코드/테스트에서 TAG 참조 업데이트
rg '@SPEC:AUTH-001' -l src/ tests/ | xargs sed -i 's/@SPEC:AUTH-001/@SPEC:AUTH-002/g'
```

### 이슈: "버전 번호가 상태와 일치하지 않음"

**증상**: `status: completed`이지만 `version: 0.0.1`

**해결책**:
```yaml
# 완료를 반영하도록 버전 업데이트
version: 0.1.0  # 구현 완료
status: completed
```

### 이슈: "HISTORY 섹션 버전 누락"

**증상**: 콘텐츠가 변경되었지만 새 HISTORY 엔트리 없음

**해결책**:
```markdown
## HISTORY

### v0.0.2 (2025-10-25)  ← 새 엔트리 추가
- **REFINED**: XYZ 요구사항 업데이트
- **AUTHOR**: @YourHandle

### v0.0.1 (2025-10-23)
- **INITIAL**: 첫 초안
```

### 이슈: "작성자 필드 @ 접두사 누락"

**증상**: `author: Goos` 대신 `author: @Goos`

**해결책**:
```yaml
# 잘못됨
author: Goos
author: goos

# 올바름
author: @Goos
```

### 이슈: "EARS 패턴 혼합"

**증상**: "WHEN 사용자가 로그인 상태이면, WHILE 세션이 활성 상태이면, 시스템은..."

**해결책**:
```markdown
# 나쁨 (패턴 혼합)
**ER-001**: WHEN 사용자가 로그인하면, WHILE 세션이 활성 상태이면, 시스템은 접근을 허용해야 한다.

# 좋음 (요구사항 분리)
**ER-001**: WHEN 사용자가 성공적으로 로그인하면, 시스템은 세션을 생성해야 한다.
**SR-001**: WHILE 세션이 활성 상태이면, 시스템은 보호된 리소스에 대한 접근을 허용해야 한다.
```

---

## 모범 사례 요약

### ✅ DO (해야 할 것)

1. **새 SPEC 생성 전 중복 ID 확인**
   ```bash
   rg "@SPEC:AUTH-001" -n .moai/specs/
   rg "AUTH-001" -n
   ```

2. **모든 콘텐츠 변경 시 HISTORY 업데이트**
   ```markdown
   ### v0.0.2 (2025-10-25)
   - **REFINED**: XYZ 추가
   - **AUTHOR**: @YourHandle
   ```

3. **버전 생애주기 엄격히 따르기**
   ```
   0.0.1 → 0.0.2 → ... → 0.1.0 → 0.1.1 → ... → 1.0.0
   (draft)  (draft)       (completed)  (patches)     (stable)
   ```

4. **작성자 필드에 @ 접두사 사용**
   ```yaml
   author: @Goos  # 올바름
   ```

5. **테스트 가능하고 측정 가능한 요구사항 작성**
   ```markdown
   # 좋음
   **UR-001**: API 응답 시간은 요청의 95%에 대해 200ms를 초과하지 않아야 한다.
   
   # 나쁨
   **UR-001**: 시스템은 빨라야 한다.
   ```

6. **7개 필수 메타데이터 필드 모두 포함**
   ```yaml
   id: AUTH-001
   version: 0.0.1
   status: draft
   created: 2025-10-23
   updated: 2025-10-23
   author: @Goos
   priority: high
   ```

7. **EARS 패턴 일관되게 사용**

### ❌ DON'T (하지 말아야 할 것)

1. **할당 후 SPEC ID 변경하지 않기**
   - TAG 체인 파괴
   - 기존 코드/테스트 고아화
   - Git 히스토리 손실

2. **HISTORY 업데이트 건너뛰지 않기**
   - 변경 근거 손실
   - 불명확한 버전 진행
   - 감사 추적 격차

3. **정당화 없이 버전 번호 건너뛰지 않기**
   ```markdown
   # 나쁨: 0.0.1 → 1.0.0
   # 좋음: 0.0.1 → 0.0.2 → ... → 0.1.0 → 1.0.0
   ```

4. **모호한 요구사항 작성하지 않기**
   - "빠른", "사용자 친화적", "좋은" 같은 용어 피하기
   - 측정 가능한 기준 사용

5. **하나의 요구사항에 여러 EARS 패턴 혼합하지 않기**

6. **제출 전 검증 잊지 않기**
   ```bash
   ./validate-spec.sh .moai/specs/SPEC-AUTH-001
   ```

7. **중복 SPEC ID 생성하지 않기**

---

## 통합 워크플로우

### `/alfred:1-plan`과 통합

`/alfred:1-plan`이 호출되면, `spec-builder` 에이전트는 이 Skill을 사용하여:

1. **분석**: 사용자 요청 및 프로젝트 컨텍스트 분석
2. **생성**: 적절한 구조로 SPEC 후보 생성
3. **검증**: 메타데이터 완전성 검증
4. **생성**: EARS 요구사항으로 `.moai/specs/SPEC-{ID}/spec.md` 생성
5. **초기화**: Git 워크플로우 (기능 브랜치, Draft PR)

### spec-builder 통합 지점

```markdown
Phase 1: SPEC 후보 생성
  ↓ (메타데이터 구조를 위해 moai-spec-authoring 사용)
Phase 2: 사용자 승인
  ↓
Phase 3: SPEC 파일 생성
  ↓ (이 Skill의 EARS 템플릿 적용)
Phase 4: Git 워크플로우 초기화
  ↓
Phase 5: /alfred:2-run으로 핸드오프
```

### 에이전트 협업

- **spec-builder**: 이 Skill의 템플릿을 사용하여 SPEC 생성
- **tag-agent**: TAG 형식 및 고유성 검증
- **trust-checker**: 메타데이터 완전성 확인
- **git-manager**: 기능 브랜치 및 Draft PR 생성

---

**Last Updated**: 2025-10-27
**Version**: 1.1.0
