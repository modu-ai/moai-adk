---
title: 대규모 코드베이스
weight: 80
draft: false
description: "수백만 줄 단일 트리나 다중 패키지 모노레포에서 Claude Code를 효율적으로 사용하는 전략을 정리합니다."
---

대규모 코드베이스 (수백만 줄 단일 저장소, 또는 여러 패키지 모노레포)에서 Claude Code는 잘 작동합니다. 다만 기본 설정이 작은 프로젝트를 가정하고 있어, **각 작업이 건드리는 부분으로만 컨텍스트를 좁히는 전략**이 필수입니다.

{{< callout type="info" >}}
**핵심**: 대규모 코드베이스의 문제는 "전체 파일을 읽는 것"이 아닙니다. 지금 작업과 **무관한 지시문과 파일이 컨텍스트를 채우는 것**입니다.
{{< /callout >}}

## 1. 시작 위치 정하기

`claude`를 어디서 실행하는지가 모든 것을 결정합니다.

| 시작 위치 | 파일 접근 범위 | 로드되는 CLAUDE.md | 적합한 경우 |
|---------|-----------|---------------|---------|
| **저장소 루트** | 전체 | 루트만 (하위는 온디맨드) | 여러 패키지/서브시스템에 걸친 작업 |
| **하위 디렉터리** | 그 서브트리만 | 그 디렉터리 + 모든 상위 디렉터리 | 한 패키지/서브시스템에 한정된 작업 |

**팁**: 한 패키지(예: `packages/api/`)에만 집중한다면, 그 디렉터리에서 `claude`를 실행하세요. 그러면 자동으로 `packages/web/`의 지시문은 로드되지 않습니다.

## 2. CLAUDE.md를 디렉터리별로 분할

루트에 모든 규칙을 넣으면:
- 너무 길어서 가독성 떨어짐
- 너무 일반적이어서 쓸모 없음
- 작업과 무관한 지시문도 로드

**해결**: 루트에 저장소 전역 규칙을 넣고, 각 하위 디렉터리에 그 영역의 규칙을 넣으세요.

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

Claude가 `packages/api/`에서 시작하면:
- 루트 + packages/api/ CLAUDE.md 모두 로드
- packages/web/ 지시문은 **로드되지 않음**

## 3. 관련 없는 CLAUDE.md 제외하기

다른 팀의 패키지나 레거시 코드는 `claudeMdExcludes`로 스킵:

```json
{
  "claudeMdExcludes": [
    "**/packages/admin-dashboard/**",
    "**/packages/legacy-*/**"
  ]
}
```

로트 CLAUDE.md는 여전히 로드되고, 제외된 패키지는 타치하지 않습니다.

## 4. 생성 코드와 벤더 코드 차단

`.gitignore`에 이미 있는 경로(node_modules, dist, build)는 자동으로 검색 결과에서 제외됩니다.

커밋된 생성 코드나 벤더 SDK는 권한 규칙으로 차단:

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

## 5. 코드 인텔리전스(LSP) 플러그인

파일을 한 줄씩 읽어서 심볼 정의를 찾는 것은 비효율적입니다. 언어 서버 플러그인을 설치하면:

```bash
/plugin install typescript-lsp@claude-plugins-official
```

Claude가 `go to definition`, `find references`, 타입 에러 직접 조회 가능.

- TypeScript, Python, Go, Rust 등 주요 언어 지원
- LSP 바이너리가 필요 (가이드 참고)

이를 통해 파일 읽기를 크게 줄일 수 있습니다.

## 6. Worktree로 필요한 디렉터리만 체크아웃

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

`--worktree`로 생성한 워크트리는 전체가 아니라 **나열한 디렉터리만** 체크아웃합니다.

- 빠른 생성 (전체 복제 vs 필요한 부분만)
- 디스크 공간 절약
- 여러 워크트리의 node_modules 중복 제거:

```json
{
  "worktree": {
    "sparsePaths": ["packages/api", "packages/shared"],
    "symlinkDirectories": ["node_modules"]  // 메인의 node_modules 공유
  }
}
```

## 7. 다른 패키지/저장소에 접근 권한 주기

한 패키지에서 시작했는데 형제 패키지 수정이 필요하면:

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

또는 런타임에:

```bash
claude --add-dir ../shared --add-dir ../web
```

## 8. 패키지별 Skills 추가

각 패키지는 그 영역만의 자동화 명령어(Skills)를 가질 수 있습니다.

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

packages/api에서 작업하면 api-testing 스킬을 자동으로 로드. packages/web에서는 로드되지 않습니다.

## 9. 패키지 간 작업 조율

같은 변경이 여러 패키지를 건드릴 때(예: 공유 타입 업데이트 + 모든 호출처 수정):

**한 세션에서 전체 변경 처리**: 모든 파일을 한 번에 로드해서 결정 일관성 유지.

**먼저 계획 작성**: 계획을 마크다운 파일에 저장. 세션이 길어지면 컨텍스트가 압축되는데, 저장된 계획은 사라지지 않습니다.

## 10. 구체적인 설정 예: 모노레포

다음은 완전한 설정 예입니다.

**루트** (`.moai/config/sections/workflow.yaml` 같은 기타 설정도 루트에):

```json
// .claude/settings.json
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

이 설정으로:
- `.claude/`, `packages/api/`, `packages/shared/` 만 체크아웃 (worktree)
- shared 패키지 접근 가능
- 생성/벤더 파일 접근 차단

## 11. 대규모 코드베이스 팁과 트릭

### 범위별 검색

큰 변경을 할 때, 영향 범위를 미리 파악하세요:

```bash
grep -r "FunctionName" packages/api/  # api만 검색
grep -r "FunctionName" packages/      # 모든 패키지
```

### 레이어별 분석

여러 레이어(DB, API, UI)를 건드는 변경이면, 각 레이어를 따로 이해하고, 한 세션에서는 하나의 변경만 집중합니다.

### 문서화 지시

대규모 변경 후에도 documentation이 유지되도록, 변경 계획에 "docs 수정" 항목을 넣으세요.

## 참고

이 가이드는 Anthropic의 공식 [Set up Claude Code in a monorepo or large codebase](https://code.claude.com/docs/en/large-codebases) 문서를 바탕으로 작성되었습니다.

추가 전략은 Anthropic의 [Best practices for Claude Code](https://code.claude.com/docs/en/best-practices) 문서도 참고하세요.
