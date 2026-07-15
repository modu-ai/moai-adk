---
title: 대규모 코드베이스
weight: 80
draft: false
description: "수백만 줄 단일 트리나 다중 패키지 모노레포에서 Claude Code를 효율적으로 사용하는 컨텍스트 축소 전략을 정리합니다."
---

# 대규모 코드베이스

수백만 줄짜리 단일 저장소든 여러 패키지로 이루어진 모노레포든, 대규모 코드베이스에서도 Claude Code는 잘 동작합니다. 다만 기본 설정이 작은 프로젝트를 가정하고 있으므로, **각 작업이 실제로 건드리는 부분으로만 컨텍스트를 좁히는 전략**이 필수입니다.

{{< callout type="info" >}}
**한 줄 요약**: 대규모 코드베이스의 진짜 문제는 "파일이 많은 것"이 아니라, 지금 작업과 **무관한 지시문과 파일이 컨텍스트를 채우는 것**입니다. 무관한 토큰은 품질을 떨어뜨리는 동시에 비용을 올립니다. 컨텍스트 축소가 곧 토크노믹스입니다.
{{< /callout >}}

## 시작 위치 정하기

`claude`를 어디서 실행하는지가 그 뒤의 모든 것을 결정합니다.

| 시작 위치 | 파일 접근 범위 | 로드되는 CLAUDE.md | 적합한 경우 |
|---------|-----------|---------------|---------|
| **저장소 루트** | 전체 | 루트만 (하위는 온디맨드) | 여러 패키지/서브시스템에 걸친 작업 |
| **하위 디렉터리** | 그 서브트리만 | 그 디렉터리 + 모든 상위 디렉터리 | 한 패키지/서브시스템에 한정된 작업 |

한 패키지(예: `packages/api/`)에만 집중하는 작업이라면 그 디렉터리에서 `claude`를 실행하세요. `packages/web/`의 지시문은 애초에 로드되지 않으므로, 규칙을 지우는 노력 없이도 컨텍스트가 저절로 가벼워집니다.

## CLAUDE.md를 디렉터리별로 분할

루트 CLAUDE.md 하나에 모든 규칙을 넣으면 세 가지 문제가 생깁니다.

- 너무 길어져 가독성이 떨어지고
- 모든 패키지에 통하려다 너무 일반적이어서 쓸모가 없어지며
- 작업과 무관한 지시문까지 매 세션 로드됩니다

해결책은 계층화입니다. 루트에는 저장소 전역 규칙만 두고, 각 하위 디렉터리에 그 영역의 규칙을 둡니다.

```markdown
# ./CLAUDE.md (루트, 모든 세션에서 로드)
This is a monorepo with three packages:
- packages/api: Node.js REST API with Express, TypeScript, PostgreSQL
- packages/web: React frontend with Vite, TypeScript, TailwindCSS
- packages/shared: shared TypeScript utilities

Run commands from the package directory.
```

```markdown
# ./packages/api/CLAUDE.md (이 디렉터리 작업할 때만 로드)
This package is the REST API server.

- Run tests: `npm test` (uses Vitest)
- Run dev server: `npm run dev` (port 3001)
- Database migrations: `npm run migrate`

API routes are in src/routes/. Never write raw SQL in handlers.
```

Claude가 `packages/api/`에서 시작하면 루트와 `packages/api/`의 CLAUDE.md는 모두 로드되지만, `packages/web/`의 지시문은 **로드되지 않습니다**.

## 관련 없는 CLAUDE.md 제외하기

다른 팀의 패키지나 레거시 코드의 지시문은 `claudeMdExcludes` 설정으로 건너뜁니다.

```json
{
  "claudeMdExcludes": [
    "**/packages/admin-dashboard/**",
    "**/packages/legacy-*/**"
  ]
}
```

루트 CLAUDE.md는 여전히 로드되고, 제외한 패키지의 지시문만 컨텍스트에서 빠집니다.

## 생성 코드와 벤더 코드 차단

`.gitignore`에 이미 있는 경로(node_modules, dist, build)는 자동으로 검색 결과에서 제외됩니다.

커밋된 생성 코드나 벤더 SDK는 권한 규칙으로 읽기 자체를 차단합니다. 생성 파일은 길고 반복적이라 컨텍스트 낭비가 특히 큽니다.

```json
{
  "permissions": {
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)",
      "Read(./**/*.generated.*)",
      "Read(./vendor/**)"
    ]
  }
}
```

## 코드 인텔리전스 (LSP) 플러그인

심볼 정의를 찾겠다고 파일을 한 줄씩 읽는 것은 토큰 관점에서 가장 비싼 탐색입니다. 언어 서버 플러그인을 설치하면 정의로 이동, 참조 찾기, 타입 오류 직접 조회가 가능해져 파일 읽기 자체를 크게 줄일 수 있습니다.

```bash
/plugin install typescript-lsp@claude-plugins-official
```

- TypeScript, Python, Go, Rust 등 주요 언어를 지원합니다
- 해당 언어의 LSP 바이너리가 시스템에 설치되어 있어야 합니다 ([플러그인 문서](/ko/claude-code/extensibility/plugins) 참고)

## 워크트리로 필요한 디렉터리만 체크아웃

`--worktree`로 생성하는 워크트리는 `worktree.sparsePaths` 설정으로 전체가 아니라 **나열한 디렉터리만** 체크아웃할 수 있습니다.

```json
{
  "worktree": {
    "sparsePaths": [
      ".claude",
      "packages/api",
      "packages/shared"
    ]
  }
}
```

- 생성이 빨라지고(전체 복제 대신 필요한 부분만)
- 디스크 공간을 아끼며
- `symlinkDirectories`로 여러 워크트리의 node_modules 중복도 제거할 수 있습니다.

```json
{
  "worktree": {
    "sparsePaths": ["packages/api", "packages/shared"],
    "symlinkDirectories": ["node_modules"]
  }
}
```

`symlinkDirectories`에 나열한 디렉터리는 메인 체크아웃의 것을 심볼릭 링크로 공유합니다.

## 다른 패키지/저장소에 접근 권한 주기

한 패키지에서 시작했는데 형제 패키지 수정이 필요해지면 `additionalDirectories`로 접근 범위를 넓힙니다.

```json
{
  "permissions": {
    "additionalDirectories": [
      "../shared",
      "../web"
    ]
  }
}
```

설정 대신 런타임 플래그로도 가능합니다.

```bash
claude --add-dir ../shared --add-dir ../web
```

## 패키지별 스킬 추가

각 패키지는 그 영역만의 스킬을 가질 수 있습니다. 스킬은 필요할 때만 로드되므로, 패키지 전용 지식을 컨텍스트 부담 없이 보관하는 좋은 그릇입니다.

```bash
mkdir -p packages/api/.claude/skills/api-testing
```

```markdown
# packages/api/.claude/skills/api-testing/SKILL.md
---
name: api-testing
description: API 패키지의 테스트 패턴
---

## Test structure
Tests are in `src/__tests__/` mirroring `src/`.

## Running tests
- All: `npm test`
- Single file: `npm test -- src/__tests__/routes/users.test.ts`

## Test utilities
- `src/__tests__/helpers/db.ts`: setupTestDb(), teardownTestDb()
- `src/__tests__/helpers/auth.ts`: createTestUser(), getAuthToken()
```

`packages/api`에서 작업하면 api-testing 스킬이 자동으로 로드되고, `packages/web`에서는 로드되지 않습니다.

## 패키지 간 작업 조율

같은 변경이 여러 패키지를 건드릴 때(예: 공유 타입 업데이트와 모든 호출처 수정)는 두 가지 원칙이 유효합니다.

- **한 세션에서 전체 변경 처리**: 관련 파일을 한 번에 로드해 결정의 일관성을 유지합니다.
- **먼저 계획을 파일로 저장**: 계획을 마크다운 파일에 남겨 두세요. 세션이 길어지면 컨텍스트가 압축되지만, 디스크에 저장된 계획은 사라지지 않습니다. "중요한 상태는 파일에 남긴다"는 에이전틱 루프 운영의 기본기이기도 합니다.

## 구체적인 설정 예: 모노레포

다음은 완전한 설정 예입니다. 루트에는 저장소 전역 차단 규칙을, 패키지에는 그 패키지의 워크트리·접근 설정을 둡니다(MoAI-ADK 프로젝트라면 `.moai/config/sections/workflow.yaml` 같은 워크플로우 설정도 루트에 위치합니다).

**루트** (`.claude/settings.json`):

```json
{
  "permissions": {
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)"
    ]
  }
}
```

**packages/api** (`.claude/settings.json`):

```json
{
  "worktree": {
    "sparsePaths": [
      ".claude",
      "packages/api",
      "packages/shared"
    ],
    "symlinkDirectories": ["node_modules"]
  },
  "permissions": {
    "additionalDirectories": ["../shared"],
    "deny": [
      "Read(./**/dist/**)",
      "Read(./**/build/**)"
    ]
  }
}
```

이 설정의 효과는 다음과 같습니다.

- 워크트리는 `.claude/`, `packages/api/`, `packages/shared/`만 체크아웃
- shared 패키지 접근 가능
- 생성/벤더 파일 접근 차단

## 팁과 트릭

### 범위를 지정한 검색

큰 변경을 하기 전에 영향 범위를 먼저 파악하세요. 검색 범위를 좁히는 습관이 읽어야 할 파일 수를 줄입니다.

```bash
grep -r "FunctionName" packages/api/  # api만 검색
grep -r "FunctionName" packages/      # 모든 패키지
```

### 레이어별 분석

DB·API·UI처럼 여러 레이어를 건드리는 변경이면 각 레이어를 따로 이해하고, 한 세션에서는 하나의 변경에만 집중합니다.

### 문서화 지시

대규모 변경 후에도 문서가 낡지 않도록, 변경 계획에 "docs 수정" 항목을 포함하세요.

## 관련 문서

- [컨텍스트 윈도우](/ko/claude-code/context-memory/context-window)
- [워크트리](/ko/claude-code/agentic/worktrees)
- [모범 사례](/ko/claude-code/agentic/best-practices)

## 참고 자료

- [Set up Claude Code in a monorepo or large codebase (공식 문서)](https://code.claude.com/docs/en/large-codebases)
- [Best practices for Claude Code (공식 문서)](https://code.claude.com/docs/en/best-practices)

{{< callout type="tip" >}}
모노레포에서 가장 손쉬운 첫 수는 "한 패키지 작업은 그 패키지 디렉터리에서 `claude` 실행"입니다. 설정 파일을 하나도 만지지 않고도 무관한 지시문 로드를 끊어내는, 비용 대비 효과가 가장 큰 습관입니다.
{{< /callout >}}
