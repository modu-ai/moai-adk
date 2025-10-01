# {{PROJECT_NAME}} - MoAI Agentic Development Kit

**SPEC-First TDD 개발 가이드**

## 핵심 철학

- **Spec-First**: 명세 없이는 코드 없음
- **TDD-First**: 테스트 없이는 구현 없음
- **GitFlow 지원**: Git 작업 자동화, Living Document 동기화, @TAG 추적성

**다중 언어 지원**: 각 언어별 최적 도구와 타입 안전성, CODE-FIRST @TAG 시스템

## 3단계 개발 워크플로우

```bash
/moai:1-spec     # 명세 작성 (EARS 방식, 사용자 확인 후 브랜치/PR 생성)
/moai:2-build    # TDD 구현 (RED→GREEN→REFACTOR)
/moai:3-sync     # 문서 동기화 (PR 상태 전환)
```

**EARS (Easy Approach to Requirements Syntax)**: 체계적인 요구사항 작성 방법론
- **Ubiquitous**: 시스템은 [기능]을 제공해야 한다
- **Event-driven**: WHEN [조건]이면, 시스템은 [동작]해야 한다
- **State-driven**: WHILE [상태]일 때, 시스템은 [동작]해야 한다
- **Optional**: WHERE [조건]이면, 시스템은 [동작]할 수 있다
- **Constraints**: IF [조건]이면, 시스템은 [제약]해야 한다

**반복 사이클**: 1-spec → 2-build → 3-sync → 1-spec (다음 기능)

## 핵심 에이전트 (8개)

| 에이전트 | 역할 | 자동화 |
|---------|------|--------|
| **spec-builder** | SPEC 작성 전담 | 사용자 확인 후 브랜치/PR 생성 |
| **code-builder** | TDD 구현 전담 (슬림화 완료) | Red-Green-Refactor (Python, TypeScript, Java, Go, Rust 등) |
| **doc-syncer** | 문서 동기화 전담 | PR 상태 전환/라벨링 |
| **cc-manager** | Claude Code 설정 전담 (슬림화 완료) | 설정 최적화/권한 |
| **debug-helper** | 오류 분석 전담 | 개발 가이드 검사 |
| **git-manager** | Git 작업 전담 | 사용자 확인 후 브랜치/PR, 커밋 자동화 |
| **trust-checker** | 품질 검증 통합 | TRUST 5원칙 검사, 코드 품질 분석 |
| **tag-agent** | TAG 시스템 독점 관리 | @TAG 체인 생성/검증/인덱싱 |

## 디버깅 & Git 관리

**디버깅**: `@agent-debug-helper "오류내용"` 또는 `@agent-debug-helper "TAG 체인 검증을 수행해주세요"`
**Git 브랜치 정책**: 모든 브랜치 생성/머지는 사용자 확인 필수
**Git 자동화**: 커밋, 푸시 등 일반 작업만 자동 처리
**Git 직접**: `@agent-git-manager "명령"` (특수 케이스)

## @TAG Lifecycle

### 핵심 설계 철학

**TDD 완벽 정렬**: RED (테스트) → GREEN (구현) → REFACTOR (문서)
**단순성**: TAG 체계 간소화
**추적성**: 코드 직접 스캔 (CODE-FIRST)

### TAG 체계

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

| TAG | 역할 | TDD 단계 | 위치 | 필수 |
|-----|------|----------|------|------|
| `@SPEC:ID` | 요구사항 명세 (EARS) | 사전 준비 | .moai/specs/ | ✅ |
| `@TEST:ID` | 테스트 케이스 | RED | tests/ | ✅ |
| `@CODE:ID` | 구현 코드 | GREEN + REFACTOR | src/ | ✅ |
| `@DOC:ID` | 문서화 | REFACTOR | docs/ | ⚠️ |

### TAG BLOCK 템플릿

**소스 코드 (src/)**:
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
```

**테스트 코드 (tests/)**:
```typescript
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
```

**SPEC 문서 (.moai/specs/)**:
```markdown
# @SPEC:AUTH-001: JWT 인증 시스템
```

- TAG ID: `<도메인>-<3자리>` (예: `AUTH-003`)
- 생성 전 중복 확인: `rg "@SPEC:AUTH" -n` 또는 `rg "AUTH-001" -n`
- **TAG의 진실은 코드 자체에만 존재**: 정규식 패턴으로 코드에서 직접 스캔하여 실시간 검증

### SPEC 연동 가이드

- `/moai:1-spec` 수행 시 `.moai/specs/SPEC-<ID>.md`에 `@SPEC:ID` 포함하여 작성
- `/moai:2-build` 수행 시 TDD 사이클에 따라 `@TEST:ID` → `@CODE:ID` 순차 생성
- `/moai:3-sync` 수행 시 `rg '@(SPEC|TEST|CODE|DOC):' -n`으로 전체 스캔 및 검증

### @CODE 서브 카테고리 (주석 레벨)

구현 세부사항은 `@CODE:ID` 내부에 주석으로 표기:
- `@CODE:ID:API` - REST API, GraphQL 엔드포인트
- `@CODE:ID:UI` - 컴포넌트, 뷰, 화면
- `@CODE:ID:DATA` - 데이터 모델, 스키마, 타입
- `@CODE:ID:DOMAIN` - 비즈니스 로직, 도메인 규칙
- `@CODE:ID:INFRA` - 인프라, 데이터베이스, 외부 연동

### 코드/테스트 적용 예시

**Python 예시**:
```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/test_service.py

class AuthenticationService:
    """@CODE:AUTH-001: JWT 인증 서비스"""

    def authenticate(self, username: str, password: str) -> bool:
        """@CODE:AUTH-001:API: 사용자 인증 API"""
        # @CODE:AUTH-001:DOMAIN: 입력 검증
        if not self._validate_input(username, password):
            return False

        # @CODE:AUTH-001:DATA: 사용자 조회
        user_data = self._get_user_data(username)

        return self._verify_credentials(user_data, password)

# tests/auth/test_service.py
# @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

def test_should_authenticate_valid_user():
    """@TEST:AUTH-001: 유효한 사용자 인증 검증"""
    service = AuthenticationService()
    result = service.authenticate("user", "password")
    assert result is True
```

**TypeScript 예시**:
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts

/**
 * @CODE:AUTH-001: JWT 인증 서비스
 *
 * TDD 이력:
 * - RED: tests/auth/service.test.ts 작성
 * - GREEN: 최소 구현 (bcrypt, JWT)
 * - REFACTOR: 타입 안전성 추가
 */
export class AuthService {
  // @CODE:AUTH-001:API: 인증 API 엔드포인트
  async authenticate(username: string, password: string): Promise<AuthResult> {
    // @CODE:AUTH-001:DOMAIN: 입력 검증
    this.validateInput(username, password);

    // @CODE:AUTH-001:DATA: 사용자 조회
    const user = await this.userRepository.findByUsername(username);

    return this.verifyCredentials(user, password);
  }
}

// tests/auth/service.test.ts
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md

describe('AuthService', () => {
  test('@TEST:AUTH-001: should authenticate valid user', () => {
    const service = new AuthService();
    const result = await service.authenticate('user', 'password');
    expect(result.success).toBe(true);
  });
});
```

### 검색 & 무결성 유지

**중복 방지**:
```bash
# 새 TAG 생성 전 기존 TAG 검색
rg "@SPEC:AUTH" -n          # SPEC 문서에서 AUTH 도메인 검색
rg "@CODE:AUTH-001" -n      # 특정 ID 검색
rg "AUTH-001" -n            # ID 전체 검색
```

**TAG 체인 검증**:
```bash
# /moai:3-sync 실행 시 자동 스캔
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# 고아 TAG 탐지 (SPEC 없는 CODE)
rg '@CODE:AUTH-001' -n src/    # CODE는 있는데
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC이 없으면 고아
```

**재사용 촉진**:
- `@agent-code-builder "기존 TAG 재사용 후보를 찾아주세요"`
- `@agent-tag-agent "AUTH 도메인 TAG 목록 조회"`

**폐기 절차**:
```python
# Deprecated TAG 표기 후 제거
# @CODE:AUTH-001:DEPRECATED (2025-01-15: AUTH-002로 대체됨)
```

### 올바른 TAG 사용 패턴

✅ **권장 패턴**:
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
export class AuthService { ... }
```

❌ **금지 패턴**:
```typescript
// @TEST:AUTH-001 -> @CODE:AUTH-001    ❌ 순서 표기 불필요 (파일 위치로 구분)
// @CODE:AUTH-001, @CODE:AUTH-002      ❌ 하나의 파일에 여러 ID (분리 필요)
// @SPEC:AUTH-001                        ❌ 구형 TAG 사용 금지
// @CODE:ABC-123                        ❌ 의미 없는 도메인명
```

### TDD 워크플로우 체크리스트

**1단계: SPEC 작성** (`/moai:1-spec`)
- [ ] `.moai/specs/SPEC-<ID>.md` 생성
- [ ] `@SPEC:ID` TAG 포함
- [ ] EARS 구문으로 요구사항 작성
- [ ] 중복 ID 확인: `rg "@SPEC:<ID>" -n`

**2단계: TDD 구현** (`/moai:2-build`)
- [ ] **RED**: `tests/` 디렉토리에 `@TEST:ID` 작성 및 실패 확인
- [ ] **GREEN**: `src/` 디렉토리에 `@CODE:ID` 작성 및 테스트 통과
- [ ] **REFACTOR**: 코드 품질 개선, TDD 이력 주석 추가
- [ ] TAG BLOCK에 SPEC/TEST 파일 경로 명시

**3단계: 문서 동기화** (`/moai:3-sync`)
- [ ] 전체 TAG 스캔: `rg '@(SPEC|TEST|CODE):' -n`
- [ ] 고아 TAG 없음 확인
- [ ] Living Document 자동 생성 확인
- [ ] PR 상태 Draft → Ready 전환

## 에이전트별 브랜치 처리 가이드라인

### 🔧 spec-builder 에이전트
```bash
# SPEC 작성 시 브랜치 생성 요청 예시
사용자: "SPEC-015 새로운 기능에 대한 명세를 작성해주세요"
에이전트: "SPEC-015 작성을 위해 feature/spec-015-new-feature 브랜치를 생성하겠습니다. 진행하시겠습니까? (y/n)"
사용자 확인 후: ✅ 브랜치 생성 및 SPEC 작성 진행
```

### 🏗️ git-manager 에이전트
```bash
# 브랜치 관리 요청 시 사용자 확인 필수
@agent-git-manager "feature 브랜치 생성"
→ "새 브랜치 feature/task-name을 생성하시겠습니까? (y/n)"

@agent-git-manager "develop 브랜치로 머지"
→ "현재 브랜치를 develop으로 머지하시겠습니까? 테스트와 문서화가 완료되었는지 확인해주세요. (y/n)"
```

### 📝 doc-syncer 에이전트
```bash
# /moai:3-sync 단계에서 머지 제안
@agent-doc-syncer "문서 동기화 완료"
→ "문서 동기화가 완료되었습니다. develop 브랜치로 머지를 진행하시겠습니까? (y/n)"
```

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

# TAG 체인 검증
@agent-tag-agent "코드 전체를 스캔하여 TAG 검증해주세요"
@agent-doc-syncer "TAG 체인 무결성 확인"

# 특정 문서 갱신
@agent-doc-syncer "API 문서를 갱신해주세요"
@agent-doc-syncer "README 업데이트 필요"
```

## TRUST 5원칙 (범용 언어 지원)

**{{PROJECT_NAME}}**: 모든 주요 프로그래밍 언어 지원
- **T**est First: 언어별 최적 도구 (Jest/Vitest, pytest, go test, cargo test, JUnit 등)
- **R**eadable: 언어별 린터 (ESLint/Biome, ruff, golint, clippy 등)
- **U**nified: 타입 안전성 (TypeScript, Go, Rust, Java) 또는 런타임 검증 (Python, JS)
- **S**ecured: 언어별 보안 도구 및 정적 분석
- **T**rackable: CODE-FIRST @TAG 시스템 (코드 직접 스캔)

상세: @.moai/memory/development-guide.md

## 언어별 코드 규칙

**공통**: 파일≤300 LOC, 함수≤50 LOC, 매개변수≤5, 복잡도≤10
**품질**: 언어별 최적 도구 자동 선택, 의도 드러내는 이름, 가드절 우선
**테스트**: 언어별 표준 프레임워크, 독립적/결정적, 커버리지≥85%

## 메모리 전략

**핵심 메모리**: @.moai/memory/development-guide.md (TRUST+@TAG)
**프로젝트 컨텍스트**:
- @.moai/project/product.md
- @.moai/project/structure.md
- @.moai/project/tech.md
**TAG 시스템**: 코드 직접 스캔 방식 (범용 언어 프로젝트 지원)
**검색 도구**: rg(권장), grep 명령어로 코드에서 직접 TAG 검색

## 프로젝트 정보

- **이름**: {{PROJECT_NAME}}
- **설명**: {{PROJECT_DESCRIPTION}}
- **버전**: {{PROJECT_VERSION}}
- **모드**: {{PROJECT_MODE}}
- **개발 도구**: 프로젝트 언어에 최적화된 도구 체인 자동 선택
