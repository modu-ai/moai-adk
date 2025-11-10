# 핵심 개념

MoAI-ADK는 신뢰할 수 있고, 추적 가능하며, 유지보수가 쉬운 개발 워크플로우를 만들기 위해 함께 작동하는 5가지 기본 개념을 기반으로 구축되었습니다. 이러한 개념을 이해하는
것이 AI 지원 개발의 잠재력을 완전히 활용하는 핵심입니다.

## 문제: AI 개발에서의 신뢰

현대 AI 지원 개발은 근본적인 문제에 직면합니다:

- **불분명한 요구사항**: "로그인 시스템을 만들어"는 사람마다 다른 의미를 가짐
- **부족한 테스트**: AI는 종종 포괄적인 테스트 커버리지 없이 코드를 생성
- **문서화 부재**: 코드는 변경되지만 문서는 오래되어감
- **컨텍스트 손실**: 각 상호작용이 처음부터 시작되어 프로젝트 기록을 잃음
- **추적 불가능한 변경**: 요구사항이 변경될 때 영향받는 코드 식별이 어려움

## 해결책: Alfred와 함께하는 SPEC-First TDD

MoAI-ADK는 체계적인 접근을 통해 이러한 문제를 해결합니다:

> **"테스트 없는 코드 없음, 사양 없는 테스트 없음"**

이것은 책임의 사슬을 만듭니다: **SPEC → TEST → CODE → DOCUMENTATION**

## 5가지 핵심 개념

### 1. SPEC-First 개발

**정의**: 코드를 작성하기 전에 명확하고 실행 가능한 사양을 작성하는 것.

**중요성**:

- 무엇을 만들지에 대한 모호성을 제거
- 자동화된 테스트를 위한 기반 제공
- 요구사항에 대한 팀 정렬 보장

**EARS 구문**: 5가지 패턴과 함께 EARS(Easy Approach to Requirements Syntax)를 사용합니다:

1. **Ubiquitous** (기본 기능): "시스템은 JWT 기반 인증을 제공해야 함"
2. **Event-driven** (조건부): "**유효한 자격 증명이 제공되면**, 시스템은 토큰을 발급해야 함"
3. **State-driven** (상태 기반): "**사용자가 인증되는 동안**, 시스템은 보호된 리소스를 허용해야 함"
4. **Optional** (선택적 기능): "**새로 고침 토큰이 존재하는 경우**, 시스템은 새 토큰을 발급할 수 있음"
5. **Constraints** (제한): "토큰 만료 시간은 15분을 초과하지 않아야 함"

**작동 방식**:

```bash
/alfred:1-plan "JWT 토큰을 사용하는 사용자 인증"
```

Alfred의 spec-builder는 EARS 형식을 사용하여 전문적 SPEC을 자동으로 생성합니다.

### 2. 테스트 주도 개발 (TDD)

**정의**: RED-GREEN-REFACTOR 사이클을 따라 구현 코드보다 먼저 테스트를 작성하는 것.

**중요성**:

- 85%+ 테스트 커버리지 보장
- 자신 있는 리팩토링 가능
- 예상되는 동작에 대한 살아있는 문서 제공

**TDD 사이클**:

1. **🔴 RED**: 먼저 실패하는 테스트 작성

   ```python
   def test_login_with_valid_credentials_should_return_token():
       """유효한 자격 증명이 제공되면 시스템은 JWT 토큰을 발급해야 함"""
       response = auth_client.login("user@example.com", "password123")
       assert response.status_code == 200
       assert "token" in response.json()
   ```

2. **🟢 GREEN**: 테스트를 통과시키기 위한 최소 구현 작성

   ```python
   def login(email: str, password: str) -> dict:
       if validate_credentials(email, password):
           return {"token": generate_jwt_token(email)}
       return {"error": "Invalid credentials"}
   ```

3. **♻️ REFACTOR**: 테스트 커버리지를 유지하면서 코드 품질 개선

   ```python
   class AuthService:
       def authenticate(self, email: str, password: str) -> AuthResult:
           if not self._validate_credentials(email, password):
               return AuthResult(success=False, error="Invalid credentials")

           token = self._generate_jwt_token(email)
           return AuthResult(success=True, token=token)
   ```

**작동 방식**:

```bash
/alfred:2-run SPEC-ID
```

Alfred는 자동으로 완전한 TDD 사이클을 실행합니다.

### 3. @TAG 시스템

**정의**: 사양, 테스트, 코드, 문서를 연결하는 고유 식별자 시스템.

**중요성**:

- 모든 프로젝트 아티팩트에 대한 완전한 추적성 활성화
- 영향 분석을 간단하고 신뢰할 수 있게 만듦
- 고아 코드와 잊어버린 요구사항 방지

**TAG 체인**:

```
@SPEC:EX-AUTH-001 (요구사항)
    ↓
@TEST:EX-AUTH-001 (테스트)
    ↓
@CODE:EX-AUTH-001:SERVICE (구현)
    ↓
@DOC:EX-AUTH-001 (문서)
```

**TAG 형식**: `<도메인>-<3자리 숫자>`

예시: `AUTH-001`, `AUTH-002`, `USER-001`, `API-001`

**사용 예시**:

```bash
# 인증과 관련된 모든 코드 찾기
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n

# 결과:
# .moai/specs/SPEC-AUTH-001/spec.md:7:# @SPEC:EX-AUTH-001: User Authentication
# tests/test_auth.py:3:# @TEST:EX-AUTH-001 | SPEC: SPEC-AUTH-001.md
# src/auth/service.py:5:# @CODE:EX-AUTH-001:SERVICE | SPEC: SPEC-AUTH-001.md
# docs/api/auth.md:24:- @SPEC:EX-AUTH-001
```

**작동 방식**:

```bash
/alfred:3-sync
```

Alfred는 자동으로 TAG 체인을 검증하고 고아 TAG를 감지합니다.

### 4. TRUST 5 원칙

**정의**: 모든 코드가 프로덕션 표준을 충족하도록 보장하는 품질 프레임워크.

**중요성**:

- 프로젝트 전체에서 일관된 코드 품질 보장
- 코드 리뷰를 위한 명확한 기준 제공
- 일반적인 버그와 보안 문제 방지

**5가지 원칙**:

1. **🧪 Test First**

   - 테스트 커버리지 ≥ 85%
   - 모든 코드가 테스트로 보호됨
   - 기능 추가 = 테스트 추가

2. **<span class="material-icons">library_books</span> Readable**

   - 함수 ≤ 50줄, 파일 ≤ 300줄
   - 변수 이름이 의도를 드러냄
   - 린터 준수 (ESLint/ruff/clippy)

3. **:bullseye: Unified**

   - SPEC 기반 아키텍처 일관성
   - 반복 패턴 (학습 곡선 감소)
   - 타입 안전성 또는 런타임 검증

4. **🔒 Secured**

   - 입력 검증 (XSS, SQL 인젝션 방지)
   - 비밀번호 해싱 (bcrypt, Argon2)
   - 민감한 데이터 보호 (환경 변수)

5. **:link: Trackable**

   - @TAG 시스템 사용
   - Git 커밋에 TAG 참조 포함
   - 모든 결정 문서화됨

**작동 방식**:

```bash
/alfred:3-sync
```

Alfred는 자동으로 TRUST 5 준수를 검증합니다.

### 5. Alfred SuperAgent

**정의**: 개발 과정 전체에서 여러 전문화된 에이전트와 스킬을 조정하는 AI 오케스트레이션 시스템.

**중요성**:

- 프롬프트 엔지니어링 복잡성 제거
- 세션 간 프로젝트 컨텍스트 유지
- 일관되고 전문적 품질의 출력 제공

**에이전트 아키텍처**:

```
Alfred SuperAgent (오케스트레이션)
    ├── Core Sub-agents (프로젝트 워크플로우)
    │   ├── project-manager 📋
    │   ├── spec-builder 🏗️
    │   ├── code-builder 💎
    │   ├── doc-syncer 📚
    │   └── quality-gate 🛡️
    ├── Expert Agents (도메인 전문가)
    │   ├── backend-expert ⚙️
    │   ├── frontend-expert 💻
    │   ├── devops-expert 🚀
    │   └── ui-ux-expert 🎨
    └── Built-in Claude Agents (일반 지원)
        ├── Code understanding
        ├── Debugging
        └── Analysis
```

**스킬 시스템**: 7개 계층으로 구성된 93개의 프로덕션급 Claude 스킬:

1. **Foundation**: 핵심 원칙 (TRUST/TAG/SPEC/Git/EARS)
2. **Essentials**: 일일 개발 도구 (debug/perf/refactor)
3. **Alfred**: 워크플로우 오케스트레이션
4. **Domain**: 전문화된 지식 (backend/frontend/security)
5. **Language**: 언어별 모범 사례 (Python/TS/Go/Rust)

**작동 방식**:

```bash
/alfred:0-project    # 프로젝트 초기화
/alfred:1-plan      # 사양 생성
/alfred:2-run       # TDD 구현
/alfred:3-sync       # 문서 동기화
```

## 완전한 워크플로우

### 단계별 프로세스

1. **PLAN** (2분)

   ```bash
   /alfred:1-plan "이메일/비밀번호를 사용하는 사용자 인증"
   ```

   - @SPEC:AUTH-001로 SPEC 생성
   - EARS 구문을 사용하여 요구사항 정의
   - 상태: `planning` → `draft`

2. **RUN** (5분)

   ```bash
   /alfred:2-run AUTH-001
   ```

   - TDD 사이클 실행 (RED → GREEN → REFACTOR)
   - @TEST:AUTH-001로 테스트 생성
   - @CODE:AUTH-001로 구현 생성
   - 상태: `draft` → `in_progress` → `testing`

3. **SYNC** (1분)

   ```bash
   /alfred:3-sync
   ```

   - @DOC:AUTH-001로 문서 생성
   - TAG 체인 무결성 검증
   - TRUST 5 준수 확인
   - 상태: `testing` → `completed`

### 결과: 완전한 추적성

```
@SPEC:EX-AUTH-001 → .moai/specs/SPEC-AUTH-001/spec.md
     ↓ (요구사항)
@TEST:EX-AUTH-001 → tests/test_auth.py
     ↓ (검증)
@CODE:EX-AUTH-001 → src/auth/service.py
     ↓ (구현)
@DOC:EX-AUTH-001 → docs/api/auth.md
```

## 시스템의 이점

### 개별 개발자를 위한 이점

- **속도**: 명확한 요구사항이 왕복 시간을 줄여줌
- **자신감**: 85%+ 테스트 커버리지가 두려움 없는 리팩토링 가능
- **명확성**: @TAG 시스템이 코드 의도를 즉시 명확하게 만듦
- **학습**: 전문적 패턴과 모범 사례가 내장됨

### 팀을 위한 이점

- **일관성**: 모두가 동일한 개발 패턴을 따름
- **온보딩**: 새 팀원이 SPEC를 통해 코드 의도를 이해
- **품질**: TRUST 5가 일관된 코드 품질 보장
- **협업**: SPEC가 요구사항에 대한 명확한 소통 제공

### 프로젝트를 위한 이점

- **유지보수성**: 코드와 문서가 항상 동기화됨
- **확장성**: TAG 시스템이 영향 분석을 trivial하게 만듦
- **신뢰성**: TDD가 강력하고 잘 테스트된 코드 보장
- **문서화**: 코드와 함께 발전하는 살아있는 문서

## 전통적 개발과의 비교

| 측면     | 전통적 접근                      | MoAI-ADK 접근                         |
| -------- | -------------------------------- | ------------------------------------- |
| 요구사항 | 구두 설명, 이메일                | EARS 구문을 사용하는 형식적 SPEC 문서 |
| 테스트   | 구현 후, 종종 불완전             | 먼저, 85%+ 커버리지 보장              |
| 문서화   | 별도로 작성, 종종 오래됨         | 코드와 자동 동기화                    |
| 추적성   | 수동적, 종종 손실됨              | @TAG 시스템이 완전한 체인 제공        |
| 품질     | 개발자에 따라 다름               | TRUST 5 원칙이 일관성 보장            |
| AI 사용  | 프롬프트 엔지니어링, 일관성 없음 | 신뢰할 수 있는 출력과 표준화된 명령어 |

## 개념으로 시작하기

1. **워크플로우 경험**: [빠른 시작 가이드](quick-start.md) 시도
2. **EARS 구문 이해**: [SPEC 작성](guides/specs/basics.md) 학습
3. **TDD 마스터**: [TDD 가이드](guides/tdd/red.md) 따르기
4. **TAG 시스템 탐색**: [TAG 문서](reference/tags/index.md) 읽기

이러한 개념은 함께 작동하여 전통적 접근보다 더 신뢰할 수 있고, 유지보수가 쉬우며, 즐거운 개발 경험을 만듭니다. Alfred를 가이드로 삼아, 프로덕션 표준을 충족한다는
자신감을 가지고 더 나은 코드를 더 빨리 작성하게 될 것입니다.
