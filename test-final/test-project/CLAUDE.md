# test-project - MoAI Agentic Development Kit

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

## 16-Core @TAG Lifecycle 2.0

### TAG BLOCK 템플릿 (필수)

```text
# @FEATURE:<DOMAIN-ID> | Chain: @REQ:<ID> -> @DESIGN:<ID> -> @TASK:<ID> -> @TEST:<ID>
# Related: @SEC:<ID>, @DOCS:<ID>
```

- 새 코드/문서/테스트 파일을 생성할 때: 위 TAG BLOCK을 파일 상단(주석) 또는 최상위 선언 근처에 배치한다
- 수정 시: 기존 TAG BLOCK을 검토해 영향받는 TAG를 업데이트하고, 불필요해진 TAG는 `@TAG:DEPRECATED-XXX`로 표시 후 `/moai:3-sync`를 수행한다
- 생성 전 중복 확인: `rg "@REQ:<키워드>" -n` 또는 `rg "<DOMAIN-ID>" -n`으로 기존 체인을 검색한다

### 체계 요약

| 카테고리 | 설명 | 필수 여부 |
|----------|------|-----------|
| Primary Chain | 요구 -> 설계 -> 작업 -> 검증 4단계 기본 체인 | 필수 |
| Implementation | Feature/API/UI/Data 등 구현 유형 | 선택 |
| Quality | Perf/Sec/Docs/Debt 등 품질 속성 | 선택 |
| Meta | Ops/Release/Tag/Deprecated 등 메타데이터 | 선택 |

- TAG ID: `<도메인>-<3자리>` (예: `AUTH-003`) — 체인 내 모든 TAG는 동일 ID를 사용한다
- 인덱스 저장소: `.moai/indexes/tags.json`, `tags.db` (SQLite) -> `/moai:3-sync` 단계에서 자동 갱신된다

### SPEC 연동 가이드

- `/moai:1-spec` 수행 시 SPEC 문서 안에 `### @TAG Catalog` 섹션을 작성한다
- Catalog 예시:

```markdown
### @TAG Catalog
| Chain | TAG | 설명 | 연관 산출물 |
|-------|-----|------|--------------|
| Primary | @REQ:AUTH-003 | OAuth2 요구사항 | SPEC-AUTH-003 |
| Primary | @DESIGN:AUTH-003 | OAuth2 시퀀스 설계 | design/oauth.md |
| Primary | @TASK:AUTH-003 | OAuth2 구현 작업 | src/auth/oauth2.ts |
| Primary | @TEST:AUTH-003 | OAuth2 통합 테스트 | tests/auth/oauth2.test.ts |
| Implementation | @FEATURE:AUTH-003 | 인증 서비스 | src/auth/service.ts |
| Quality | @SEC:AUTH-003 | 보안 점검 | docs/security/oauth2.md |
```

- SPEC 변경 -> Catalog 업데이트 -> 코드/테스트 반영 -> `/moai:3-sync`로 인덱스 확정 순서를 유지한다

### 코드/테스트 적용 예시

**Python 예시**:
```python
# @FEATURE:LOGIN-001 | Chain: @REQ:AUTH-001 -> @DESIGN:AUTH-001 -> @TASK:AUTH-001 -> @TEST:AUTH-001
# Related: @SEC:LOGIN-001, @DOCS:LOGIN-001
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

# @TEST:LOGIN-001 연결: @TASK:LOGIN-001 -> @TEST:LOGIN-001
def test_should_authenticate_valid_user():
    """@TEST:LOGIN-001: 유효한 사용자 인증 테스트"""
    service = AuthenticationService()
    result = service.authenticate("user", "password")
    assert result is True
```

**TypeScript 예시**:
```typescript
// @FEATURE:LOGIN-001 | Chain: @REQ:AUTH-001 -> @DESIGN:AUTH-001 -> @TASK:AUTH-001 -> @TEST:AUTH-001
// Related: @SEC:LOGIN-001, @DOCS:LOGIN-001
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

// @TEST:LOGIN-001: Vitest/Jest 테스트
describe('AuthService', () => {
  test('@TEST:LOGIN-001: should authenticate valid user', () => {
    // 테스트 구현...
  });
});
```

### 검색 & 무결성 유지

- 중복 방지: 새 TAG 도입 전 `rg "@TAG" -g"*.ts"`, `rg "AUTH-001"` 등으로 기존 체인을 확인한다
- 재사용 촉진: 구현 계획 단계에서 `@agent-code-builder`에게 "기존 TAG 재사용 후보를 찾아주세요"라고 요청한다
- 무결성 검사: `/moai:3-sync` 또는 `@agent-doc-syncer "TAG 인덱스를 업데이트해주세요"` 실행 후 로그에서 고아 TAG를 해결한다
- 폐기 절차: 더 이상 사용하지 않는 TAG는 `@TAG:DEPRECATED-<ID>`로 표기하고 Catalog에서 상태를 `Deprecated`로 갱신한다

### 금지 패턴 (잘못된 예시)

```python
@TASK:LOGIN-001 -> @DESIGN:LOGIN-001      # 순서 위반
@FEATURE:LOGIN-001 (중복 선언)          # 고유성 위반
@REQ:ABC-123                             # 의미 없는 ID
```

### 업데이트 체크리스트 (점검용)

- [ ] TAG BLOCK이 모든 신규/수정 파일에 존재하는가?
- [ ] Primary Chain 4종이 끊김 없이 연결되는가?
- [ ] SPEC `@TAG Catalog`와 코드/테스트가 동일한 ID를 공유하는가?
- [ ] `tags.json`/`tags.db`가 `/moai:3-sync` 이후 최신 상태인가?

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

## TRUST 5원칙 (범용 언어 지원)

**test-project**: 모든 주요 프로그래밍 언어 지원
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

- **이름**: test-project
- **설명**: A test-project project built with MoAI-ADK
- **버전**: 0.1.0
- **모드**: personal
- **개발 도구**: 프로젝트 언어에 최적화된 도구 체인 자동 선택
