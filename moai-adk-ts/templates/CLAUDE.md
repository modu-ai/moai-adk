# {{PROJECT_NAME}} - MoAI Agentic Development Kit

**SPEC-First TDD 개발 가이드**

## 핵심 철학

- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **GitFlow 지원**: Git 작업 자동화, Living Document 동기화, 16-Core @TAG 추적성

**다중 언어 지원**: 각 언어별 최적 도구와 타입 안전성, JSON 기반 TAG 시스템

## 3단계 개발 워크플로우

```bash
/moai:1-spec     # 명세 작성 (EARS 방식, 브랜치/PR 생성)
/moai:2-build    # TDD 구현 (RED→GREEN→REFACTOR)
/moai:3-sync     # 문서 동기화 (PR 상태 전환)
```

**반복 사이클**: 1-spec → 2-build → 3-sync → 1-spec (다음 기능)

## 핵심 에이전트 (5개)

| 에이전트 | 역할 | 자동화 |
|---------|------|--------|
| **spec-builder** | EARS 명세 작성 | 브랜치/PR 생성 |
| **code-builder** | 범용 언어 TDD 구현 | Red-Green-Refactor (Python, TypeScript, Java, Go, Rust 등) |
| **doc-syncer** | 문서 동기화 | PR 상태 전환/라벨링 |
| **cc-manager** | Claude Code 관리 | 설정 최적화/권한 |
| **debug-helper** | 오류 진단 | 개발 가이드 검사 |

## 디버깅 & Git 관리

**디버깅**: `@agent-debug-helper "오류내용"` 또는 `@agent-debug-helper "TAG 체인 검증을 수행해주세요"`
**Git 자동화**: 모든 워크플로우에서 자동 처리 (99% 케이스)
**Git 직접**: `@agent-git-manager "명령"` (1% 특수 케이스)

## 16-Core @TAG 시스템 (JSON)

```
@REQ → @DESIGN → @TASK → @TEST
SPEC: REQ,DESIGN,TASK | PROJECT: VISION,STRUCT,TECH,ADR
IMPLEMENTATION: FEATURE,API,TEST,DATA | QUALITY: PERF,SEC,DEBT,TODO
```

**TAG 인덱스**: `.moai/indexes/tags.json` (JSON 기반)

## 에이전트 실제 사용법

### 🔍 디버깅 & 분석

```bash
# 오류 분석
@agent-debug-helper "TypeError: 'NoneType' object has no attribute 'name'"
@agent-debug-helper "Git push 오류 해결 방법"

# @TAG 시스템 검증
@agent-debug-helper "TAG 체인 검증을 수행해주세요"
@agent-debug-helper "고아 TAG 및 끊어진 링크 감지"
@agent-debug-helper "TAG 무결성 검사"

# 개발 가이드 준수 확인
@agent-debug-helper "개발 가이드 검사"
@agent-debug-helper "TRUST 원칙 준수 여부 확인"
```

### 🚀 TDD 구현

```bash
# 분석 단계 (계획 수립)
@agent-code-builder "SPEC-013 분석해주세요"
@agent-code-builder "구현 계획을 수립해주세요"

# 구현 단계 (사용자 승인 후)
@agent-code-builder "승인된 계획으로 TDD 구현을 시작해주세요"
@agent-code-builder "구현을 진행해주세요"
```

### 📝 문서 동기화

```bash
# 전체 문서 동기화
@agent-doc-syncer "코드와 문서를 동기화해주세요"
@agent-doc-syncer "문서 동기화 수행"

# TAG 인덱스 갱신
@agent-doc-syncer "TAG 인덱스를 업데이트해주세요"
@agent-doc-syncer "tags.json 갱신"

# 특정 문서 갱신
@agent-doc-syncer "API 문서를 갱신해주세요"
@agent-doc-syncer "README 업데이트 필요"
```

## @TAG 시스템 실제 사용법

### 16-Core TAG 카테고리

**Primary Chain (필수)**: 모든 기능은 이 체인을 따라야 함
```
@REQ:LOGIN-001 → @DESIGN:LOGIN-001 → @TASK:LOGIN-001 → @TEST:LOGIN-001
```

**Implementation (구현별)**: 기능 유형에 따라 선택
```
@FEATURE:LOGIN-001  # 비즈니스 로직
@API:LOGIN-001      # API 엔드포인트
@UI:LOGIN-001       # 사용자 인터페이스
@DATA:LOGIN-001     # 데이터 모델
```

**Quality (품질보증)**: 필요에 따라 적용
```
@PERF:LOGIN-001     # 성능 최적화
@SEC:LOGIN-001      # 보안 강화
@DOCS:LOGIN-001     # 문서화
@TAG:LOGIN-001      # 메타 태깅
```

### 코드에서 @TAG 적용 예시

**Python 예시**:
```python
# @FEATURE:LOGIN-001 연결: @REQ:AUTH-001 → @DESIGN:AUTH-001 → @TASK:AUTH-001
class AuthenticationService:
    """@FEATURE:LOGIN-001: 사용자 인증 서비스 구현"""

    def authenticate(self, username: str, password: str) -> bool:
        """@API:LOGIN-001: 사용자 인증 API 엔드포인트"""
        # @SEC:LOGIN-001: 입력값 보안 검증
        if not self._validate_input(username, password):
            return False

        # @PERF:LOGIN-001: 캐시된 인증 결과 확인
        if cached_result := self._get_cached_auth(username):
            return cached_result

        return self._verify_credentials(username, password)

# @TEST:LOGIN-001 연결: @TASK:LOGIN-001 → @TEST:LOGIN-001
def test_should_authenticate_valid_user():
    """@TEST:LOGIN-001: 유효한 사용자 인증 테스트"""
    service = AuthenticationService()
    result = service.authenticate("user", "password")
    assert result is True
```

**TypeScript 예시**:
```typescript
// @FEATURE:LOGIN-001: 인증 서비스 타입스크립트 구현
interface AuthService {
  // @API:LOGIN-001: 인증 API 인터페이스 정의
  authenticate(username: string, password: string): Promise<boolean>;
}

// @UI:LOGIN-001: 로그인 컴포넌트
const LoginForm: React.FC = () => {
  // @SEC:LOGIN-001: 클라이언트 사이드 입력 검증
  const handleSubmit = (username: string, password: string) => {
    // 구현...
  };

  return <form>...</form>;
};

// @TEST:LOGIN-001: Jest 테스트
describe('AuthService', () => {
  test('@TEST:LOGIN-001: should authenticate valid user', () => {
    // 테스트 구현...
  });
});
```

### TAG 체인 연결 원칙

1. **순차적 연결**: Primary Chain은 반드시 순서대로
2. **명확한 참조**: 부모 TAG를 명시적으로 참조
3. **고유성 보장**: 동일 기능에 중복 TAG ID 금지
4. **의미있는 네이밍**: 기능을 명확히 드러내는 ID 사용

### 잘못된 TAG 사용 예시 (❌)

```python
# ❌ 순서 위반: @TASK가 @DESIGN보다 먼저
@TASK:LOGIN-001 → @DESIGN:LOGIN-001

# ❌ 고아 TAG: 연결되지 않은 독립 태그
@FEATURE:RANDOM-999

# ❌ 중복 ID: 동일 기능에 여러 TAG
@FEATURE:LOGIN-001
@FEATURE:LOGIN-001  # 중복!

# ❌ 의미 없는 ID
@REQ:ABC-123
```

### 올바른 TAG 사용 예시 (✅)

```python
# ✅ 완전한 Primary Chain
@REQ:USER-AUTH-001 → @DESIGN:USER-AUTH-001 → @TASK:USER-AUTH-001 → @TEST:USER-AUTH-001

# ✅ Implementation과 Quality TAG 연결
@TASK:USER-AUTH-001 → @FEATURE:USER-AUTH-001, @API:USER-AUTH-001
@FEATURE:USER-AUTH-001 → @PERF:USER-AUTH-001, @SEC:USER-AUTH-001

# ✅ 의미있는 네이밍
@REQ:USER-AUTH-001  # 사용자 인증 요구사항
@DESIGN:USER-AUTH-001  # 사용자 인증 설계
@TASK:USER-AUTH-001  # 사용자 인증 구현 작업
```

## TRUST 5원칙 (범용 언어 지원)

**{{PROJECT_NAME}}**: 모든 주요 프로그래밍 언어 지원
- **T**est First: 언어별 최적 도구 (Jest/Vitest, pytest, go test, cargo test, JUnit 등)
- **R**eadable: 언어별 린터 (ESLint/Biome, ruff, golint, clippy 등)
- **U**nified: 타입 안전성 (TypeScript, Go, Rust, Java) 또는 런타임 검증 (Python, JS)
- **S**ecured: 언어별 보안 도구 및 정적 분석
- **T**rackable: JSON 기반 16-Core @TAG 시스템

상세: @.moai/memory/development-guide.md

## 언어별 코드 규칙

**공통**: 파일≤300 LOC, 함수≤50 LOC, 매개변수≤5, 복잡도≤10
**품질**: 언어별 최적 도구 자동 선택, 의도 드러내는 이름, 가드절 우선
**테스트**: 언어별 표준 프레임워크, 독립적/결정적, 커버리지≥85%

## 메모리 전략

**핵심 메모리**: @.moai/memory/development-guide.md (TRUST+16-Core TAG)
**프로젝트 컨텍스트**:
- @.moai/project/product.md
- @.moai/project/structure.md
- @.moai/project/tech.md
**TAG 시스템**: JSON 기반 인덱스 (범용 언어 프로젝트 지원)
**검색 도구**: 언어별 최적화된 TAG 검색, rg(권장), grep, find 지원

## 프로젝트 정보

- **이름**: {{PROJECT_NAME}}
- **설명**: {{PROJECT_DESCRIPTION}}
- **버전**: {{PROJECT_VERSION}}
- **모드**: {{PROJECT_MODE}}
- **개발 도구**: 프로젝트 언어에 최적화된 도구 체인 자동 선택