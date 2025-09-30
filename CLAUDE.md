# ${PROJECT_NAME} - MoAI Agentic Development Kit v0.0.4 ✅

**TypeScript 기반 고성능 SPEC-First TDD 개발 가이드 (TRUST 92% 달성 + 모듈화 아키텍처 완성)**

## 핵심 철학 (v0.0.4 달성)

- ✅ **Spec-First**: 명세 없이는 코드 없음 (3단계 워크플로우 완성)
- ✅ **TDD-First**: 테스트 없이는 구현 없음 (Vitest 92.9% 성공률)
- ✅ **TRUST 5원칙**: 62% → 92% 달성 (목표 82% 대비 112% 초과 달성)
- ✅ **모듈화 설계**: Orchestrator 1,467 → 135 LOC (91% 감소)
- ✅ **보안 강화**: Winston logger + 구조화 로깅, console.* 완전 제거

**성능 지표**: TypeScript 5.9.2 + Bun 98% + Vitest 92.9% + Biome 94.8% + Winston Logger 97.92% coverage + TRUST 92%

## 3단계 핵심 워크플로우 (완성) ✅

```bash
/moai:1-spec     # ✅ 명세 작성 (EARS 방식, 사용자 확인 후 브랜치/PR 생성)
/moai:2-build    # ✅ TDD 구현 (RED→GREEN→REFACTOR)
/moai:3-sync     # ✅ 문서 동기화 (PR 상태 전환)
```

**CLI 명령어 (v0.0.3 100% 완성)**: init, doctor, status, update, restore, help, version (언어 감지 + 동적 요구사항)

## 핵심 에이전트 (7개) ✅

| 에이전트 | v0.0.3 달성 상태 | 주요 기능 |
|---------|---------|---------|
| **spec-builder** | ✅ **완성** | EARS 명세 작성, 사용자 확인 후 브랜치/PR 생성 |
| **code-builder** | ✅ **완성** | 범용 언어 TDD 구현 (Red-Green-Refactor) |
| **doc-syncer** | ✅ **완성** | 문서 동기화, PR 상태 전환 |
| **cc-manager** | ✅ **완성** | Claude Code 설정 최적화 |
| **debug-helper** | ✅ **강화** | **지능형 시스템 진단**, 언어 감지, 동적 요구사항 |
| **git-manager** | ✅ **완성** | 사용자 확인 후 브랜치/PR, 커밋 자동화 |
| **trust-checker** | ✅ **완성** | TRUST 5원칙 검증 |

## CLI 및 디버깅 (100% 완성 + 혁신적 진단) ✅

**CLI 명령어**: `moai doctor`, `moai init`, `moai status`, `moai update`, `moai restore`
**지능형 진단**: `@agent-debug-helper "오류내용"` (시스템 진단 자동화 + 언어 감지)
**Git 브랜치 정책**: ✅ 모든 브랜치 생성/머지는 사용자 확인 필수
**Git 자동화**: ✅ 커밋, 푸시 등 일반 작업만 자동 처리
**혁신적 시스템 진단**:
- 🔍 **언어 자동 감지**: JavaScript/TypeScript/Python/Java/Go 프로젝트 분석
- 🎯 **동적 요구사항**: 감지된 언어에 따라 개발 도구 자동 추가
- 📊 **5단계 진단**: Runtime(2) + Development(2) + Optional(1) + Language-Specific + Performance

## @TAG Lifecycle  ✅

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
- **TAG의 진실은 코드 자체에만 존재**: 정규식 패턴으로 코드에서 직접 스캔하여 실시간 검증

**코드 스캔 철학**: 중간 캐시 없이 코드를 직접 스캔하여 TAG 추적성 보장

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

- SPEC 변경 -> Catalog 업데이트 -> 코드/테스트 반영 -> `/moai:3-sync`로 코드 스캔 및 검증 수행한다

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
- 무결성 검사: `/moai:3-sync` 실행으로 코드 전체를 스캔하여 TAG 체인 검증 및 고아 TAG 식별
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
- [ ] TAG 체인이 코드 스캔을 통해 검증되었는가?

## TRUST 5원칙 (v0.0.4: 92% 달성) ✅

**전체 준수율**: 62% → 92% (+30%, 목표 82% 대비 112% 달성)

- ✅ **T**est First (70% → 80%): Vitest 92.9% 성공률, GitLockManager/GitManager 테스트 안정화
- ✅ **R**eadable (52% → 100%): Orchestrator 모듈화 (1,467 → 135 LOC, 91% 감소), Biome 94.8%
- ✅ **U**nified (75% → 90%): TypeScript 5.9.2 엄격 타입 검사, 의존성 주입 패턴, 단일 책임 원칙
- ✅ **S**ecured (65% → 100%): Winston logger (97.92% coverage), 민감정보 마스킹, console.* 완전 제거
- ✅ **T**rackable (48% → 90%): @AI-TAG 코드 스캔 시스템, TAG 체인 검증

**v0.0.4 핵심 개선**:
- Phase 1: GitLockManager 테스트 안정화 (26 tests 100% pass)
- Phase 2: Orchestrator 대규모 리팩토링 (9개 모듈 분해)
- Phase 3: Winston logger 보안 시스템 (288 console.* 전환)
- Phase 4: GitManager 테스트 안정화 + 최종 console.* 제거

상세: @.moai/memory/development-guide.md

## 언어별 코드 규칙

**공통**: 파일≤300 LOC, 함수≤50 LOC, 매개변수≤5, 복잡도≤10
**품질**: 언어별 최적 도구 자동 선택, 의도 드러내는 이름, 가드절 우선
**테스트**: 언어별 표준 프레임워크, 독립적/결정적, 커버리지≥85%

## Living Document 전략 (자동 동기화) ✅

**핵심 메모리**: @.moai/memory/development-guide.md (TRUST+@AI-TAG 완성)
**프로젝트 문서**: ✅ 자동 동기화 완료
- @.moai/project/product.md (v0.0.3 시스템 진단 개선 반영)
- @.moai/project/structure.md (SystemChecker 아키텍처 개선)
- @.moai/project/tech.md (언어별 도구 매핑 완성)
**@AI-TAG 시스템**: 코드 직접 스캔 방식, 정규식 기반 실시간 추출
**고속 검색**: rg/grep 명령어로 코드 직접 검색, 중간 캐시 없음
**시스템 진단 성과**: SQLite3→npm+TypeScript+Git LFS 실용화, 언어 감지 시스템 완성