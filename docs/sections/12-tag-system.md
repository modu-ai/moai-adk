# MoAI-ADK 16-Core @TAG 시스템

## 🏷️ 16-Core TAG 체계

MoAI-ADK의 16-Core @TAG 시스템은 모든 요구사항과 구현을 완전하게 추적할 수 있는 체계를 제공합니다.

### 16-Core 태그 카테고리

#### SPEC (명세 관련)
- **@REQ**: 요구사항 (Requirements)
- **@SPEC**: 명세 식별자/요약 (Specification Capsule)
- **@DESIGN**: 설계 문서 (Design Documents)
- **@TASK**: 작업 항목 (Task Items)

#### Steering (방향성)
- **@VISION**: 제품 비전 (Product Vision)
- **@STRUCT**: 구조 설계 (Structure Design)
- **@TECH**: 기술 선택 (Technology Choices)
- **@ADR**: 아키텍처 결정 기록 (Architecture Decision Records)

#### Implementation (구현)
- **@FEATURE**: 기능 구현 (Feature Implementation)
- **@API**: API 엔드포인트 (API Endpoints)
- **@TEST**: 테스트 케이스 (Test Cases)
- **@DATA**: 데이터 모델 (Data Models)

#### Quality (품질)
- **@PERF**: 성능 요구사항 (Performance)
- **@SEC**: 보안 요구사항 (Security)
- **@DEBT**: 기술 부채 (Technical Debt)
- **@TODO**: 할 일 목록 (Todo Items)

## 추적성 체인

### Primary Chain (핵심 추적성)
```
@REQ → @DESIGN → @TASK → @TEST
```

**예시**:
```markdown
@REQ:USER-AUTH-001 "사용자는 이메일과 패스워드로 로그인할 수 있다"
↓
@DESIGN:JWT-AUTH-001 "JWT 토큰 기반 인증 시스템 설계"
↓
@TASK:AUTH-API-001 "로그인 API 엔드포인트 구현"
↓
@TEST:AUTH-LOGIN-001 "로그인 성공/실패 테스트 케이스"
```

### Steering Chain (방향성 추적성)
```
@VISION → @STRUCT → @TECH → @ADR
```

**예시**:
```markdown
@VISION:PLATFORM-001 "개발자 생산성 10배 향상"
↓
@STRUCT:MICROSERVICE-001 "마이크로서비스 아키텍처"
↓
@TECH:CONTAINER-001 "Docker & Kubernetes 선택"
↓
@ADR:DEVOPS-001 "AWS EKS + GitLab CI/CD"
```

### Quality Chain (품질 추적성)
```
@PERF → @SEC → @DEBT → @TODO
```

## TAG 사용 규칙

### 1. 태그 형식
```
@[TYPE]:[ID] "설명"
```

**올바른 예시**:
```markdown
@REQ:USER-LOGIN-001 "사용자 로그인 기능"
@API:POST-AUTH-LOGIN "로그인 API 엔드포인트"
@TEST:UNIT-AUTH-001 "로그인 단위 테스트"
```

### 2. ID 네이밍 규칙

#### REQ (Requirements)
```
@REQ:[CATEGORY]-[DESCRIPTION]-[NUMBER]
예: @REQ:USER-LOGIN-001, @REQ:PERF-RESPONSE-001
```

#### API (API Endpoints)
```
@API:[METHOD]-[RESOURCE]-[ACTION]
예: @API:GET-USERS-LIST, @API:POST-AUTH-LOGIN
```

#### TEST (Test Cases)
```
@TEST:[TYPE]-[TARGET]-[NUMBER]
예: @TEST:UNIT-LOGIN-001, @TEST:E2E-CHECKOUT-001
```

### 3. 품질 규칙

#### 필수 연결
- 모든 @TASK는 @REQ에서 시작되어야 함
- 모든 @TEST는 @TASK 또는 @FEATURE와 연결되어야 함
- @DESIGN은 관련 @REQ를 참조해야 함

#### 금지 사항
- 순환 참조 금지: A → B → A
- 고아 태그 금지: 참조되지 않는 태그
- 중복 ID 금지: 동일한 ID 재사용

## 자동 관리 시스템

### 1. 실시간 검증 (tag_validator.py Hook)
```python
# 태그 생성 시 자동 검증
tag_pattern = r'@([A-Z]+)[-:]([A-Z0-9-]+)'
found_tags = re.findall(tag_pattern, content)

# 규칙 위반 시 차단
if not validation_result['valid']:
    print(f"⚠️ 16-Core @TAG 규칙 위반: {validation_result['error']}")
    sys.exit(2)  # 차단
```

### 2. 자동 인덱싱 (.moai/indexes/tags.json)
```json
{
  "tags": {
    "@REQ:USER-LOGIN-001": {
      "file": ".moai/specs/SPEC-001-auth/spec.md",
      "line": 15,
      "description": "사용자 로그인 기능",
      "links_to": ["@DESIGN:JWT-AUTH-001"],
      "linked_from": [],
      "created": "2025-09-16T10:30:00Z",
      "last_updated": "2025-09-16T10:30:00Z"
    }
  },
  "chains": {
    "primary": [
      ["@REQ:USER-LOGIN-001", "@DESIGN:JWT-AUTH-001", "@TASK:AUTH-API-001", "@TEST:AUTH-LOGIN-001"]
    ]
  },
  "stats": {
    "total_tags": 24,
    "completed_chains": 3,
    "orphaned_tags": 0,
    "quality_score": 0.95
  }
}
```

### 3. 추적성 매트릭스 (.moai/indexes/traceability.json)
```json
{
  "matrix": {
    "REQ-001": {
      "requirement": "@REQ:USER-LOGIN-001",
      "design": "@DESIGN:JWT-AUTH-001",
      "tasks": ["@TASK:AUTH-API-001", "@TASK:AUTH-UI-001"],
      "tests": ["@TEST:AUTH-LOGIN-001", "@TEST:AUTH-SECURITY-001"],
      "coverage": 1.0
    }
  },
  "coverage_report": {
    "requirements_covered": "100%",
    "tasks_tested": "95%",
    "design_implemented": "100%"
  }
}
```

## TAG 관리 도구

### 1. 검증 도구
```bash
# TAG 무결성 검사
python .moai/scripts/validate_tags.py

# 추적성 검증
python .moai/scripts/check-traceability.py --verbose

# 고아 태그 찾기
python .moai/scripts/validate_tags.py --orphaned
```

### 2. 자동 수정 도구
```bash
# TAG 링크 자동 복구
python .moai/scripts/repair_tags.py --execute

# 인덱스 재생성
python .moai/scripts/repair_tags.py --rebuild-index
```

### 3. 보고서 생성
```bash
# 추적성 보고서
python .moai/scripts/check-traceability.py --report

# TAG 품질 보고서
python .moai/scripts/validate_tags.py --quality-report
```

## 실전 사용 예시

### 새 기능 개발 플로우
```markdown
1. 요구사항 정의
@REQ:PAYMENT-STRIPE-001 "Stripe 결제 시스템 통합"

2. 설계 문서 작성
@DESIGN:PAYMENT-API-001 "결제 API 설계"
- 참조: @REQ:PAYMENT-STRIPE-001

3. 작업 분해
@TASK:STRIPE-SDK-001 "Stripe SDK 통합"
@TASK:PAYMENT-DB-001 "결제 내역 DB 설계"
- 참조: @DESIGN:PAYMENT-API-001

4. 테스트 작성
@TEST:PAYMENT-SUCCESS-001 "결제 성공 테스트"
@TEST:PAYMENT-FAILURE-001 "결제 실패 테스트"
- 참조: @TASK:STRIPE-SDK-001
```

### 품질 지표 추적
```markdown
성능 요구사항:
@PERF:PAYMENT-RESPONSE-001 "결제 응답 시간 < 3초"

보안 요구사항:
@SEC:PAYMENT-ENCRYPTION-001 "결제 정보 AES-256 암호화"

기술 부채:
@DEBT:PAYMENT-LEGACY-001 "레거시 결제 시스템 제거"
```

16-Core @TAG 시스템은 **완전한 추적성**과 **자동화된 품질 관리**를 통해 개발 과정의 투명성을 보장합니다.
