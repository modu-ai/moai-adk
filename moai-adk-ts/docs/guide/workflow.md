# 3단계 개발 워크플로우

MoAI-ADK는 엄격한 3단계 SPEC 우선 TDD 워크플로우를 따릅니다.

## 1단계: SPEC 작성 (`/moai:1-spec`)

코드 작성 전에 상세한 명세를 작성합니다:

```bash
# 예: 인증 기능 명세 작성
/moai:1-spec "사용자 인증" "OAuth2 통합"
```

**EARS 요구사항 형식:**
- **Ubiquitous**: 시스템은 [기능]을 제공해야 한다
- **Event-driven**: WHEN [조건]이면, 시스템은 [동작]해야 한다
- **State-driven**: WHILE [상태]일 때, 시스템은 [행동]해야 한다
- **Optional**: WHERE [조건]이면, 시스템은 [기능]을 제공할 수 있다
- **Constraints**: IF [조건]이면, 시스템은 [제약]을 따라야 한다

### EARS 예시

```markdown
### Ubiquitous Requirements (기본 요구사항)
- 시스템은 사용자 인증 기능을 제공해야 한다

### Event-driven Requirements (이벤트 기반)
- WHEN 사용자가 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 오류를 반환해야 한다

### State-driven Requirements (상태 기반)
- WHILE 사용자가 인증된 상태일 때, 시스템은 보호된 리소스 접근을 허용해야 한다

### Optional Features (선택적 기능)
- WHERE 리프레시 토큰이 제공되면, 시스템은 새로운 액세스 토큰을 발급할 수 있다

### Constraints (제약사항)
- IF 잘못된 토큰이 제공되면, 시스템은 접근을 거부해야 한다
- 액세스 토큰 만료시간은 15분을 초과하지 않아야 한다
```

## 2단계: TDD 구현 (`/moai:2-build`)

Red-Green-Refactor 사이클을 따릅니다:

```bash
# SPEC-001 구현
/moai:2-build SPEC-001
```

**프로세스:**
1. **RED**: SPEC 기반 실패하는 테스트 작성
2. **GREEN**: 테스트를 통과하는 최소한의 코드 구현
3. **REFACTOR**: 테스트를 유지하면서 코드 품질 개선

### 언어별 TDD 예시

**TypeScript (주력 언어):**
```typescript
// @TEST:AUTH-001: 유효한 사용자 인증
describe('AuthService', () => {
  test('@TEST:AUTH-001: authenticates valid credentials', async () => {
    const result = await authService.authenticate('user', 'pass');
    expect(result).toBeTruthy();
  });
});
```

**Python:**
```python
# @TEST:AUTH-001: 유효한 사용자 인증
def test_should_authenticate_valid_user():
    """@TEST:AUTH-001: 유효한 자격증명 인증"""
    result = auth_service.authenticate("user", "pass")
    assert result is True
```

## 3단계: 문서 동기화 (`/moai:3-sync`)

추적성을 유지하고 문서를 동기화합니다:

```bash
/moai:3-sync
```

**수행 작업:**
- 리빙 독 업데이트
- @TAG 체인 검증
- PR 상태 전환
- 동기화 리포트 생성

## 워크플로우 다이어그램

```
SPEC → Test → Code → Sync → SPEC (다음 기능)
  ↓      ↓      ↓      ↓
 @REQ  @TEST  @IMPL  @DOCS
```

## 모범 사례

1. **SPEC 작성을 건너뛰지 마세요**: 모든 코드는 SPEC까지 추적 가능해야 합니다
2. **@TAG 체인 유지**: 요구사항과 구현을 연결합니다
3. **정기적으로 `moai doctor` 실행**: 시스템 준수 확인
4. **각 단계마다 커밋**: 롤백을 위한 체크포인트 생성

## 일반적인 워크플로우 패턴

### 새 기능 개발

```bash
# 1. SPEC 작성
/moai:1-spec "신규 기능명"

# 2. 사용자 승인 대기 (브랜치 생성 확인)

# 3. TDD 구현
/moai:2-build SPEC-XXX

# 4. 문서 동기화
/moai:3-sync

# 5. 사용자 확인 후 머지
```

### 버그 수정

```bash
# 1. 버그 재현 SPEC 작성
/moai:1-spec "버그명" "재현 단계"

# 2. 실패하는 테스트 작성 (RED)

# 3. 버그 수정 (GREEN)

# 4. 리팩토링 (REFACTOR)

# 5. 문서 업데이트
/moai:3-sync
```

## 문제 해결

### SPEC 변경이 필요한 경우

1. 기존 SPEC 검토
2. 변경 영향 분석
3. 새 SPEC 버전 작성
4. 관련 테스트 및 코드 업데이트
5. `/moai:3-sync`로 추적성 유지

### 테스트 실패

1. 테스트가 SPEC을 정확히 반영하는지 확인
2. 구현이 SPEC을 준수하는지 검증
3. `moai doctor`로 환경 문제 확인
4. @TAG 체인으로 관련 코드 추적

## 다음 단계

- [SPEC 우선 TDD](/guide/spec-first-tdd) 방법론 상세 학습
- [@TAG 시스템](/guide/tag-system) 추적성 이해