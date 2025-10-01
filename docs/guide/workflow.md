# MoAI-ADK 0→3 Workflow

이 문서는 MoAI CLI와 Claude 명령(
`/moai:*`
)으로 프로젝트를 세팅하고 SPEC→BUILD→SYNC 흐름을 빠르게 반복하기 위한 실전 지침입니다. 과장된 수치나 추상적인 설명은 제외하고, 실제로 따라 할 수 있는 명령과 산출물만 정리했습니다.

---

## 0. 프로젝트 기반 다지기

### 0-1. CLI로 기본 환경 준비

```bash
# 프로젝트 루트에서 실행
moai init                # .moai, .claude, CLAUDE.md 생성
moai doctor              # Node, Git, npm 등 필수 도구 점검
moai status              # 현재 프로젝트가 MoAI 템플릿을 잘 갖췄는지 확인
moai update --check      # 템플릿 업데이트 필요 여부만 확인
```

실행 시 CLI는 다음과 같이 알려줍니다.

- `moai doctor` → `🔍 Checking system requirements...` 후, 설치 여부와 버전 요구사항을 `✅`/`⚠️`/`❌`로 표시합니다.
- `moai status` → `📊 MoAI-ADK Project Status` 아래에서 `.moai`, `.claude`, `CLAUDE.md`, `.git` 존재 여부와 템플릿 버전을 출력합니다.
- `moai update` → 최신 버전이 있으면 `⚡ 최신 버전: v…`와 같이 알려 주고, `--no-backup`이 없으면 `.moai-backup/<timestamp>/`에 안전 복사본을 남깁니다.

### 0-2. 프로젝트 문서 정비 (`/moai:8-project`)

Claude 편집기에서 `/moai:8-project`를 실행하면 `project-manager.md`가 product/structure/tech.md 초안을 작성합니다.

```
/moai:8-project MyService
```

- 현재 디렉터리와 언어(예: `package.json` → TypeScript)를 분석합니다.
- `product.md`, `structure.md`, `tech.md`를 덮어쓰거나 보강하므로, 필요 시 Git 스테이지 후 실행하세요.
- 프롬프트 마지막에 생성/갱신 계획이 요약되며, 사용자가 “진행/중단”을 직접 선택합니다.

---

## 1. SPEC 단계 (`/moai:1-spec`)

`spec-builder.md`는 요구사항을 정리하여 `docs/specs/` 또는 `.moai/specs/` 등에 저장합니다.

```
/moai:1-spec "사용자 인증"
```

- 질문에 답하면 EARS 형식으로 `Ubiquitous`, `Event-driven`, `State-driven`, `Constraints` 섹션을 채웁니다.
- SPEC 파일 첫머리에는 TAG 블록을 직접 추가하세요. 예시:

```markdown
/**
 * @SPEC:AUTH-001 | Title: 사용자 인증
 */
```

- SPEC ID는 `SPEC-<TOPIC>-NNN` 패턴으로 통일하면 `@TAG` 검색이 쉬워집니다.

간단한 SPEC 예시:

```markdown
# SPEC-AUTH-001 사용자 인증

## Ubiquitous
- 시스템은 이메일/비밀번호 기반 로그인을 제공해야 한다.

## Event-driven
- WHEN 유효한 자격 증명으로 로그인하면 JWT 토큰을 발급한다.
- WHEN 토큰이 만료되면 401 응답을 반환한다.

## Constraints
- Access Token TTL: 15분.
```

---

## 2. BUILD 단계 (`/moai:2-build`)

`code-builder.md`는 SPEC을 받아 테스트와 구현 골격을 생성합니다.

```
/moai:2-build SPEC-AUTH-001
```

실행 흐름:

1. **RED** – 실패하는 테스트 초안 작성.
2. **GREEN** – 최소한의 코드로 테스트 통과.
3. **REFACTOR** – 중복 제거와 이름 정리.

TypeScript 예시:

```typescript
// tests/auth/login.test.ts
/** @TEST:AUTH-001 로그인 테스트 */
describe('login', () => {
  it('returns a token for valid credentials', async () => {
    const result = await authService.login('user@example.com', 'Pass1234!');
    expect(result.accessToken).toBeDefined();
  });
});

// src/auth/service.ts
/** @CODE:AUTH-001 로그인 서비스 */
export class AuthService {
  async login(email: string, password: string) {
    this.assertInputs(email, password);
    const user = await this.users.findByEmail(email.toLowerCase());
    if (!user || !(await this.passwords.verify(password, user.hash))) {
      throw new Error('Invalid credentials');
    }
    return this.tokens.issue(user.id);
  }
}
```

로컬 실행 시에는 일반적인 Node/Bun 테스트 도구를 그대로 사용합니다.

```bash
bun test        # bun 사용 시
npm test        # npm 사용 시
```

각 파일 상단에 `@TEST:*`, `@CODE:*` TAG를 남겨 3단계에서 쉽게 연결되도록 합니다.

---

## 3. SYNC 단계 (`/moai:3-sync`)

`doc-syncer.md`는 코드와 문서를 스캔해 TAG 체인과 문서 상태를 확인합니다.

```
/moai:3-sync
```

실행 결과 예시:

```
코드 스캔 중...
✅ SPEC-AUTH-001 @SPEC / @TEST / @CODE 연결 완료
📄 sync-report.md 갱신, docs/api/ 업데이트
```

- 누락된 TAG가 있으면 `❌`로 목록을 보여주며, 수정 후 다시 실행하면 됩니다.
- 동기화가 끝나면 `git status`로 변경 파일을 검토하고 커밋/PR을 진행합니다.

---

## CLI 기본 명령 요약

| 명령 | 용도 | 참고 출력 |
|------|------|-----------|
| `moai init` | MoAI 템플릿 설치 | `🚀 Initializing ...`, `.moai/config.json` 생성 |
| `moai doctor` | 도구/버전 확인 | `🔍 Checking system requirements...` |
| `moai status` | 프로젝트 상태 확인 | `.moai`, `.claude`, 템플릿 버전 요약 |
| `moai update` | 템플릿 갱신 | 최신 여부, 백업 경로 안내 |
| `moai restore <backup>` | 백업 복원 | 건너뛴 항목/복원 항목 표시 |

---

## Claude 에이전트 한눈에 보기

| 파일 | 역할 요약 |
|------|-----------|
| `spec-builder.md` | 질문을 통해 SPEC 초안 작성 |
| `code-builder.md` | 테스트/코드 골격 작성, TDD 진행 가이드 |
| `doc-syncer.md` | SYNC 단계 지원, TAG 검증 안내 |
| `project-manager.md` | `/moai:8-project` 수행, project/*.md 유지 |
| `git-manager.md` | Git 작업 체크리스트 안내 |
| `debug-helper.md` | 실패 로그 분석 도우미 |
| `tag-agent.md` | @TAG 패턴 진단, 누락된 링크 제안 |
| `trust-checker.md` | 품질 점검 체크리스트 |
| `cc-manager.md` | Claude 설정 관련 리마인더 |

모든 에이전트 정의 파일은 `templates/.claude/agents/moai/`에 있으며, 필요 시 내용을 읽고 조직 내부 규칙에 맞게 커스터마이즈할 수 있습니다.

---

## Claude Hooks

| 파일 | 실행 시점 | 기능 |
|------|-----------|------|
| `policy-block.cjs` | Bash 도구 사용 전 | 위험 명령(`rm -rf /` 등) 차단 |
| `pre-write-guard.cjs` | Write/Edit 실행 전 | `.moai/memory/` 등 민감 경로 보호 |
| `session-notice.cjs` | 세션 시작 시 | 프로젝트 상태 요약 출력 |
| `tag-enforcer.cjs` | 저장 전 | TAG 블록 누락/형식 검사 |

Hooks는 `.claude/hooks/moai/`에 있으며 Node 환경에서 동작합니다. 필요 시 `node path/to/hook`으로 단독 실행해 동작을 테스트할 수 있습니다.

---

## @TAG와 SYNC의 중요성

MoAI-ADK는 코드 자체를 단일 진실 소스로 사용합니다. 파일마다 다음과 같이 TAG를 남겨 추적성을 확보하십시오.

```typescript
/**
 * @SPEC:AUTH-001 | 로그인 요구사항
 * @TEST:AUTH-001 | tests/auth/login.test.ts
 * @CODE:AUTH-001 | src/auth/service.ts
 * @DOC:AUTH-001  | docs/api/auth.md
 */
```

- `@SPEC` → 요구사항 문서
- `@TEST` → 자동/수동 테스트
- `@CODE` → 구현 파일
- `@DOC` → 사용자 문서

`/moai:3-sync`는 이 네 가지가 모두 연결돼 있는지 확인합니다. 하나라도 빠지면 sync 보고서에서 `❌`로 표시되므로 즉시 보완하세요.

---

## 반복 루틴 체크리스트

1. `moai status`로 상태 점검.
2. `/moai:8-project`로 project/*.md 최신화.
3. `/moai:1-spec` → `/moai:2-build` → `/moai:3-sync` 순서로 기능 개발.
4. `npm test`/`bun test` 등 로컬 테스트 통과 여부 확인.
5. `git status` 확인 후 커밋과 PR 생성.

위 흐름을 반복하면 SPEC, 코드, 문서, 추적성이 항상 일치하는 상태를 유지할 수 있습니다.
