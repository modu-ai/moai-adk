# 빠른 시작 가이드

**⏱️ 소요 시간: 약 5분**

이 가이드를 따라 MoAI-ADK를 설치하고 첫 번째 SPEC-First TDD 프로젝트를 시작할 수 있습니다.

## 전제 조건

MoAI-ADK를 사용하기 전에 다음 요구사항을 확인하세요:

- **Node.js 18.0 이상**: [nodejs.org](https://nodejs.org)에서 다운로드
- **Git 2.25 이상**: [git-scm.com](https://git-scm.com)에서 다운로드
- **(권장) Bun 1.2.19 이상**: [bun.sh](https://bun.sh)에서 설치

MoAI-ADK는 Windows, macOS, Linux 모든 플랫폼에서 동작합니다. 설치 후 `moai doctor` 명령으로 시스템 요구사항을 자동으로 검증할 수 있습니다.

## 1단계: MoAI-ADK 설치

MoAI-ADK는 npm 패키지로 제공되며, Bun, npm, yarn 중 원하는 패키지 매니저로 설치할 수 있습니다.

:::code-group

```bash [Bun (권장 - 98% 빠름)]
# 글로벌 설치
bun add -g moai-adk

# 버전 확인
moai --version
# 출력: 0.0.1
```

```bash [npm]
# 글로벌 설치
npm install -g moai-adk

# 버전 확인
moai --version
# 출력: 0.0.1
```

```bash [yarn]
# 글로벌 설치
yarn global add moai-adk

# 버전 확인
moai --version
# 출력: 0.0.1
```

:::

`moai --version` 명령이 버전 번호를 출력하면 설치가 성공한 것입니다. 설치 과정은 패키지 크기가 195KB로 매우 작아 수 초 내에 완료됩니다.

## 2단계: 시스템 진단 실행

프로젝트를 시작하기 전에 개발 환경이 올바르게 구성되어 있는지 확인합니다. `moai doctor` 명령은 5개 카테고리에 걸쳐 체계적인 진단을 수행합니다.

```bash
moai doctor
```

### 진단 카테고리 설명

진단은 다음 5개 카테고리로 구분되며, 각 카테고리별로 필수 도구를 검증합니다:

1. **Runtime Requirements** ✅
   - Node.js 18.0+ 버전 확인
   - Git 2.25+ 버전 및 설정 확인

2. **Development Tools** 🔧
   - npm, yarn, pnpm, Bun 중 사용 가능한 패키지 매니저 감지
   - TypeScript 설치 확인 (선택적)

3. **Optional Tools** ⭐
   - Docker (컨테이너 환경 지원 시)
   - GitHub CLI (Team 모드 사용 시)
   - SQLite3 제거 (v0.0.1부터 불필요)

4. **Language-Specific Tools** 🌍
   - 프로젝트 디렉토리 분석으로 언어 자동 감지
   - JavaScript/TypeScript: Vitest, Biome 추천
   - Python: pytest, mypy, ruff 추천
   - Java: JUnit, Maven/Gradle 추천
   - Go, Rust 등 추가 언어 지원

5. **Performance Checks** ⚡
   - 디스크 I/O 속도 테스트
   - 네트워크 연결 상태 확인

### 진단 결과 예시

```
🗿 MoAI-ADK v0.0.1 - System Diagnostics

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Runtime Requirements
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ✅ Node.js  18.19.0 (required: >=18.0.0)
  ✅ Git      2.42.0 (required: >=2.25.0)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Development Tools
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ✅ bun      1.2.19 (recommended)
  ✅ npm      10.2.5
  ✅ TypeScript 5.9.2

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Language-Specific Tools
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  🌐 Detected Language: TypeScript
  ✅ Vitest   3.2.4 (test runner)
  ✅ Biome    2.2.4 (linter/formatter)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Summary: 9/9 checks passed (100%)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Your system is ready for SPEC-First TDD development!
```

진단에서 문제가 발견되면 각 항목별로 해결 방법이 표시됩니다. 예를 들어 Node.js 버전이 낮으면 업그레이드 링크를, Git이 설치되지 않았으면 다운로드 링크를 제공합니다.

## 3단계: 프로젝트 초기화

`moai init` 명령으로 새 프로젝트를 초기화합니다. 대화형 위저드가 프로젝트 설정을 안내합니다.

```bash
# 프로젝트 생성
moai init my-first-project

# 또는 현재 디렉토리에 초기화
moai init
```

### 초기화 과정 상세 설명

초기화 과정은 다음 단계로 진행됩니다:

1. **프로젝트 정보 입력**
   ```
   ? 프로젝트 이름: my-first-project
   ? 주 개발 언어: TypeScript (자동 감지)
   ? 프로젝트 모드: Personal (로컬 개발) / Team (GitHub 연동)
   ```

2. **디렉토리 구조 생성**
   ```
   my-first-project/
   ├── .moai/                    # MoAI-ADK 설정 및 SPEC
   │   ├── config.json          # 프로젝트 설정
   │   ├── memory/              # 개발 가이드라인
   │   ├── specs/               # SPEC 문서 저장소
   │   # TAG는 소스코드에만 존재 (CODE-FIRST)
   │   └── project/             # 프로젝트 메타데이터
   │       ├── product.md       # 제품 정의
   │       ├── structure.md     # 구조 설계
   │       └── tech.md         # 기술 스택
   │
   ├── .claude/                 # Claude Code 통합
   │   ├── agents/moai/         # 7개 전문 에이전트
   │   ├── commands/moai/       # 3단계 워크플로우 명령어
   │   ├── hooks/moai/          # 8개 이벤트 훅 (JavaScript)
   │   └── output-styles/       # 출력 스타일
   │
   └── src/                     # 소스 코드 (언어에 따라 다름)
   ```

3. **템플릿 설치**
   - `.moai/memory/development-guide.md`: TRUST 5원칙 및 개발 가이드
   - 7개 에이전트 정의 파일
   - 3단계 워크플로우 명령어 (`/moai:1-spec`, `/moai:2-build`, `/moai:3-sync`)
   - TypeScript 훅 (빌드된 JavaScript 파일)

4. **설치 완료 메시지**
   ```
   ✅ Project initialized successfully!

   📂 Project: my-first-project
   📁 Location: /Users/you/my-first-project
   🗿 Mode: Personal

   🚀 Next steps:
   1. cd my-first-project
   2. Open in Claude Code
   3. Run: /moai:1-spec "Your first feature"
   ```

전체 초기화 과정은 수 초 내에 완료되며, 즉시 SPEC-First TDD 개발을 시작할 수 있습니다.

## 4단계: 첫 번째 SPEC 작성

프로젝트를 Claude Code에서 열고, `/moai:1-spec` 명령으로 첫 번째 명세를 작성합니다.

### Claude Code에서 프로젝트 열기

```bash
# VS Code에서 프로젝트 열기
code my-first-project

# Claude Code 세션 시작
# (Claude Code 확장이 설치되어 있어야 합니다)
```

### SPEC 작성 시작

Claude Code 채팅창에서 다음 명령을 실행합니다:

```
/moai:1-spec "사용자 인증 기능 구현"
```

`spec-builder` 에이전트가 활성화되어 다음 과정을 안내합니다:

1. **EARS 방식 요구사항 작성**
   - Ubiquitous: 시스템은 [기능]을 제공해야 한다
   - Event-driven: WHEN [조건]이면, 시스템은 [동작]해야 한다
   - State-driven: WHILE [상태]일 때, 시스템은 [동작]해야 한다
   - Optional: WHERE [조건]이면, 시스템은 [동작]할 수 있다
   - Constraints: IF [조건]이면, 시스템은 [제약]해야 한다

2. **@TAG Catalog 생성**
   - Primary Chain: @REQ → @DESIGN → @TASK → @TEST
   - Implementation: @FEATURE, @API, @UI, @DATA
   - Quality: @PERF, @SEC, @DOCS

3. **Acceptance Criteria 정의**
   - Given-When-Then 형식의 검증 기준
   - 측정 가능한 성공 지표

4. **브랜치 생성 (사용자 확인)**
   - Personal 모드: 로컬 `feature/spec-001-auth` 브랜치
   - Team 모드: GitHub Issue 및 PR 자동 생성

### SPEC 문서 예시

생성된 SPEC 문서는 `.moai/specs/SPEC-001/` 디렉토리에 저장됩니다:

```markdown
# SPEC-001: 사용자 인증 기능

## Metadata
- ID: SPEC-001
- Title: 사용자 인증 기능 구현
- Author: Your Name
- Created: 2025-01-15
- Status: Draft

## Requirements (EARS)

### Ubiquitous
- 시스템은 이메일 기반 사용자 인증 기능을 제공해야 한다

### Event-driven
- WHEN 사용자가 유효한 자격증명으로 로그인하면, 시스템은 JWT 토큰을 발급해야 한다
- WHEN 토큰이 만료되면, 시스템은 401 에러를 반환해야 한다

### Constraints
- 비밀번호는 최소 8자 이상이어야 한다
- 토큰 만료시간은 15분을 초과하지 않아야 한다

## @TAG Catalog
| Chain | TAG | 설명 | 연관 산출물 |
|-------|-----|------|--------------|
| Primary | @REQ:AUTH-001 | 인증 요구사항 | SPEC-001 |
| Primary | @DESIGN:AUTH-001 | JWT 기반 설계 | design/auth.md |
| Primary | @TASK:AUTH-001 | 인증 구현 작업 | src/auth/service.ts |
| Primary | @TEST:AUTH-001 | 인증 테스트 | tests/auth/service.test.ts |

## Acceptance Criteria
- [ ] 유효한 이메일/비밀번호로 로그인 성공
- [ ] 잘못된 자격증명으로 로그인 실패
- [ ] JWT 토큰 발급 및 검증
- [ ] 토큰 만료 처리
```

SPEC 작성이 완료되면 Git 브랜치가 생성되고(사용자 확인 후), 다음 단계인 TDD 구현을 시작할 수 있습니다.

## 5단계: TDD 구현

SPEC 작성이 완료되면 `/moai:2-build` 명령으로 TDD 구현을 시작합니다.

```
/moai:2-build SPEC-001
```

`code-builder` 에이전트가 Red-Green-Refactor 사이클을 가이드합니다:

### Red 단계: 실패하는 테스트 작성

```typescript
// tests/auth/service.test.ts
// @TEST:AUTH-001 | Chain: @REQ:AUTH-001 → @DESIGN:AUTH-001 → @TASK:AUTH-001 → @TEST:AUTH-001
import { describe, test, expect } from 'vitest';
import { AuthService } from '@/auth/service';

describe('AuthService', () => {
  test('@TEST:AUTH-001: should authenticate with valid credentials', async () => {
    const service = new AuthService();
    const result = await service.authenticate('user@example.com', 'password123');
    expect(result.success).toBe(true);
    expect(result.token).toBeDefined();
  });

  test('@TEST:AUTH-001: should fail with invalid credentials', async () => {
    const service = new AuthService();
    const result = await service.authenticate('user@example.com', 'wrong');
    expect(result.success).toBe(false);
    expect(result.error).toBe('Invalid credentials');
  });
});
```

테스트 실행 결과 (실패 확인):
```bash
❌ FAIL tests/auth/service.test.ts
  AuthService
    ✗ should authenticate with valid credentials
      TypeError: AuthService is not a constructor
```

### Green 단계: 테스트를 통과하는 최소 구현

```typescript
// src/auth/service.ts
// @FEATURE:AUTH-001 | Chain: @REQ:AUTH-001 → @DESIGN:AUTH-001 → @TASK:AUTH-001 → @TEST:AUTH-001
import jwt from 'jsonwebtoken';

export class AuthService {
  async authenticate(email: string, password: string): Promise<{
    success: boolean;
    token?: string;
    error?: string;
  }> {
    // @SEC:AUTH-001: 입력 검증
    if (!email || !password) {
      return { success: false, error: 'Missing credentials' };
    }

    // @SEC:AUTH-001: 비밀번호 검증 (실제로는 DB 조회)
    if (password === 'password123') {
      // @API:AUTH-001: JWT 토큰 발급
      const token = jwt.sign({ email }, 'secret', { expiresIn: '15m' });
      return { success: true, token };
    }

    return { success: false, error: 'Invalid credentials' };
  }
}
```

테스트 실행 결과 (통과):
```bash
✅ PASS tests/auth/service.test.ts
  AuthService
    ✓ should authenticate with valid credentials (25ms)
    ✓ should fail with invalid credentials (5ms)

Tests  2 passed (2)
```

### Refactor 단계: 코드 품질 개선

- 하드코딩된 비밀번호를 실제 DB 조회로 변경
- JWT secret을 환경 변수로 이동
- 에러 처리 개선
- 함수를 50 LOC 이하로 유지

code-builder 에이전트가 TRUST 5원칙을 자동으로 검증하여 코드 품질을 보장합니다.

## 6단계: 문서 동기화 및 완료

구현이 완료되면 `/moai:3-sync` 명령으로 문서를 동기화하고 추적성을 검증합니다.

```
/moai:3-sync
```

`doc-syncer` 에이전트가 다음 작업을 수행합니다:

1. **코드 스캔 및 TAG 추출**
   ```
   🔍 Scanning codebase for @TAGs...
   ✅ Found 8 TAGs in 4 files
   ✅ Primary Chain complete: @REQ:AUTH-001 → @DESIGN:AUTH-001 → @TASK:AUTH-001 → @TEST:AUTH-001
   ```

2. **TAG 무결성 검증**
   - 끊어진 체인 감지
   - 고아 TAG 식별
   - 중복 TAG 확인

3. **Living Document 업데이트**
   - `.moai/memory/development-guide.md` 업데이트
   - `.moai/project/` 문서 동기화
   - API 문서 자동 생성 (TypeDoc, Sphinx 등)

4. **TAG 코드 스캔 (CODE-FIRST)**
   ```typescript
   // 코드에서 직접 @TAG 추출 (인덱스 파일 없음)
   {
     "version": "4.0",
     "lastUpdated": "2025-01-15T10:30:00Z",
     "tags": [
       {
         "id": "AUTH-001",
         "chain": ["REQ", "DESIGN", "TASK", "TEST"],
         "files": [
           "specs/SPEC-001/spec.md",
           "src/auth/service.ts",
           "tests/auth/service.test.ts"
         ],
         "status": "completed"
       }
     ]
   }
   ```

5. **PR 상태 전환 (Team 모드)**
   - Draft → Ready for Review
   - 체크리스트 및 테스트 결과 추가
   - Merge 준비 완료

### 동기화 완료 메시지

```
✅ Documentation synchronized successfully!

📊 Summary:
  - TAGs validated: 8/8 (100%)
  - Primary Chains: 1/1 complete
  - Files updated: 3
  - Test coverage: 95%

🚀 Ready for review!
```

## 다음 단계

축하합니다! 첫 번째 SPEC-First TDD 사이클을 완료했습니다. 이제 다음을 시도해보세요:

### 1. 추가 기능 구현
```bash
/moai:1-spec "비밀번호 재설정 기능"
/moai:2-build SPEC-002
/moai:3-sync
```

### 2. 프로젝트 상태 확인
```bash
moai status
# 출력: Git 상태, SPEC 진행률, TAG 추적성, 테스트 커버리지
```

### 3. 시스템 업데이트
```bash
moai update --check  # 업데이트 확인
moai update          # 최신 버전으로 업데이트
```

### 4. 고급 기능 탐색
- [TAG 시스템 완전 가이드](/guide/tag-system) - 추적성 관리
- [CLI 명령어 가이드](/cli/init) - 명령어 상세 설명

## 문제 해결

### 설치 오류

```bash
# 캐시 정리 후 재설치
npm cache clean --force
npm install -g moai-adk

# 또는 Bun 사용
bun cache rm
bun add -g moai-adk
```

### 권한 오류 (macOS/Linux)

```bash
# sudo 없이 설치 (권장)
npm config set prefix ~/.npm-global
export PATH=~/.npm-global/bin:$PATH

# 또는 sudo 사용
sudo npm install -g moai-adk
```

### 진단 실패

```bash
# 상세 로그 확인
moai doctor --verbose

# 특정 요구사항 확인
node --version  # 18.0 이상 필요
git --version   # 2.25 이상 필요
```

## 도움말

- 📚 [전체 문서](https://adk.mo.ai.kr)
- 💬 [커뮤니티](https://mo.ai.kr) *(오픈 예정)*
- 🐛 [이슈 리포트](https://github.com/modu-ai/moai-adk/issues)
- 💡 [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions)

---

**다음 읽기**: [3단계 워크플로우 상세 가이드](/guide/workflow)