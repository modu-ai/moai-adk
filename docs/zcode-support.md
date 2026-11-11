# ZCode 지원 근거와 adapter gap (02-D #1994)

Z.ai ZCode를 공통 skill package(`skills/moai/`, 02-A #1991; #2286에서 skill 이름 `moai`로 확정, 실행 파일은 `moai-cli`)에 연결할 때의 상황을 기록한다. 공식 출처가 명시하는 것, 명시하지 않는 것, 그리고 adapter가 맡아야 할 일을 나눈다. 지원표 행은 `docs/host-support-matrix.md`에 있다. 이 문서는 ZCode용 파일을 생성하거나 설치하지 않는다. 확인일은 모두 2026-11-08이다.

## 범위와 결정

- RC1 ZCode adapter는 헤드리스 경로만 다룬다(`zcode --prompt … --output-format stream-json`). 수신과 훅은 미확인으로 표기한다. ACP adapter, 원격 옵션, Orca/Hide 상호운용은 후속이다(사용자 결정 2026-11-08 16:4x, `docs/decisions-2026-11-08.md`).
- "지원"은 실제 앱 E2E가 있는 `certified` 행에만 쓴다. 지금 ZCode 행은 모두 `documented` 또는 `unknown`이다.
- 실측은 not_run이다. 이 개발 머신에는 `zcode`가 없다(`which zcode` → not found).

## 출처

ZCode에는 성격이 다른 1차 출처가 두 개 있다. RC1이 고른 헤드리스 경로는 두 번째 출처가 설명하는 CLI다.

| 출처 | 표면 | 위치 | 고정 |
| --- | --- | --- | --- |
| ZCode 앱 문서 | 데스크톱 앱 | https://zcode.z.ai/en/docs (agents, skill, hooks, plugin, ADE-tools) | 페이지에 발행일이 없다. 다운로드 링크는 3.17.2. `/en/docs/skills`는 404이고 올바른 경로는 `/en/docs/skill`이다 |
| ZCode CLI 사용 문서 | `zcode` CLI(헤드리스) | https://github.com/zai-org/ZCode/tree/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage | `usage/` 최신 commit `a60ba6a7e5a0…`(2026-10-30). `02-cli-reference.md`는 `0a1fc56d279f…`(2026-10-19). 저장소 license Apache-2.0. 최신 릴리스 `v3.17.2`(2026-11-04) |
| Z.ai devpack 안내 | 설치·모델 설정 | https://docs.z.ai/devpack/tool/zcode | 헤드리스 플래그는 없다 |

## 원문 인용

앱 문서:

- agents: "When both sources exist, ZCode appends the user global instructions first, then the workspace instructions." / "It does not merge multiple `AGENTS.md` files across directory levels." / "It also does not scan child directories, expand `@import` / `@include`, or choose rule files automatically by task type." / "ZCode only uses it during onboarding as a one-time migration source"(`CLAUDE.md`)
- skill: "~/.zcode/skills/<skill-name>/SKILL.md" / "The frontmatter must include `name` and `description`; a skill missing either is ignored" / description 1024자 초과는 "Going over drops the whole skill" / "A `SKILL.md` body over 100KB is truncated when loaded." / "Type `$` in the chat input and select a skill." / Import 대화상자는 "each source shows its skill directory path and how many skills are available." 하지만 경로 자체는 적혀 있지 않다
- hooks: "The file must set `hooks.enabled: true`" / "Project-level hooks are not executed in the current version." / `.agents/settings.json` / `.claude/settings.json`은 "Read-only display, not executed; import explicitly into `.zcode` from the settings page"
- plugin: "The manifest is looked up in priority order: `.zcode-plugin/plugin.json` (recommended)" 다음이 `.claude-plugin/plugin.json` / "ZCode preloads the **Claude Code marketplace**" / "Plugin hooks only join new sessions after the plugin is enabled" / "Plugins installed locally don't follow you into a remote workspace by default."
- 문서 내비게이션에 Statusline, output style, Import/Migration 페이지가 없다. ADE Tools의 "status dots"는 task 목록 표시이고 statusline이 아니다

CLI 사용 문서(원문은 중국어):

- 02-cli-reference: "`--mode plan|build|edit|yolo|guarded`：显式指定权限模式。`--prompt` 缺省为 `yolo`" / "`--output-format text|json|stream-json`" / "`--resume <sessionId>`：恢复指定会话" / "`--cwd <path>`：本次调用的有效工作目录，用于项目配置、技能发现和会话归属。" / `--model`, `--permission-mode`, `--allowed-tools`는 "不要使用未经解析器支持的" 목록에 있다
- 05-permissions: "`plan`：… 它不是操作系统级只读沙箱" / "`yolo`：普通工具跳过权限提示，适合已隔离的无人值守任务。"
- 08-skills: "当前默认发现路径包括用户及项目的 `.zcode/skills` 和 `.agents/skills`。Git 工作区中会向工作树根查找技能目录；同级 `.zcode` 优先于 `.agents`。" / "不要把磁盘上有 SKILL.md 当成已加载。" 예시 skill에 `$ARGUMENTS`가 있다
- 09-hooks: "用户配置、项目配置和插件都可能提供 Hook；项目 Hook 需要额外信任。" / `zcode hooks trust review|grant|status|revoke --workspace … --hook-digest …`
- 10-plugins: "兼容 manifest 中的字段也不一定具有运行时实现；以加载诊断为准。"
- design/agents-user-instructions-merge: "ZCode should treat `~/.zcode/AGENTS.md` as the user's default instruction file and merge it with the workspace `AGENTS.md`" / workspace 파일은 "search from `workingDirectory` upward until the detected project root"

## 앱 문서와 CLI 문서의 차이

| 항목 | 앱 문서 | CLI 문서 | 지원표 처리 |
| --- | --- | --- | --- |
| skill 발견 경로 | `~/.zcode/skills`만 명시 | 사용자·project의 `.zcode/skills`와 `.agents/skills`, `skills.roots` | 두 surface를 각자 `documented`로 적는다 |
| project hook | 현재 버전에서 실행하지 않음 | digest 단위 신뢰(`hooks trust grant`) 뒤 실행 | 두 surface를 각자 `documented`로 적는다. 같은 3.17.2에서 두 설명이 모두 맞는지는 시험 대기 |
| skill 길이 제한 | description 1024자, 본문 100KB | 언급 없음 | CLI는 `unknown`(시험 대기) |
| 인자 치환 | 언급 없음 | 예시에 `$ARGUMENTS` | 공통 package는 `$ARGUMENTS`를 쓰지 않으므로(02-A) 영향 없음 |

차이의 원인(앱과 CLI가 서로 다른 런타임인지, 문서 시점 차이인지)은 확인하지 않았다. 실측 전에는 어느 쪽도 다른 쪽을 덮어쓰지 않는다.

## 공통 package와 ZCode 제약 대조

| 제약 | ZCode 값 | `skills/moai/` 실측 | 결과 |
| --- | --- | --- | --- |
| frontmatter 필수 | `name`, `description`(앱) | `name: moai`, `description` 있음 | 충족 |
| description 길이 | 1024자 이하(앱) | 140자 | 충족 |
| 본문 크기 | 100KB 이하(앱) | `SKILL.md` 1,589 B, 가장 큰 workflow 1,893 B | 충족 |
| 참조 깊이 | ZCode 제약 없음(문서 언급 없음) | SKILL.md → routes.json → workflow(02-A 정의) | 해당 없음 |
| `$ARGUMENTS`·`!` shell context | CLI는 `$ARGUMENTS` 예시 있음 | 사용 안 함 | 영향 없음 |

## Adapter gap

분류: `구현 필요`는 adapter 코드가 있어야 하는 일, `대안 있음`은 문서화된 다른 경로가 있는 일, `미검증`은 실측해야 판단할 수 있는 일이다.

| # | adapter 책임 | 분류 | 근거 | 후속 제안 |
| --- | --- | --- | --- | --- |
| G1 | discovery 경로에 package 놓기. 앱은 `~/.zcode/skills/moai`, CLI는 `.zcode/skills`·`.agents/skills` 후보 | 구현 필요 | 앱 skill 문서, CLI 08-skills | 위치는 HOLD(아래). 결정 뒤 02-B/02-C projection 규칙을 따르는 ZCode projection Issue |
| G2 | 로드 확인: 파일이 있다고 로드된 것은 아니다 | 대안 있음 | CLI "不要把磁盘上有 SKILL.md 当成已加载"와 `zcode skills list --cwd <dir> --json` | adapter 실측 Issue에서 `skills list` 결과를 receipt로 기록 |
| G3 | 명시 호출 안내. 앱은 `$moai`, CLI는 문법 미문서 | 미검증 | 앱 skill 문서. CLI 문서에 없음 | 헤드리스 실측에서 `--prompt` 문장의 skill 지명 동작 확인 |
| G4 | description 길이 검사(1024자 초과 시 앱이 skill을 버림) | 대안 있음 | 앱 skill 문서. 02-A 시험이 이미 1024자 이하를 강제 | 추가 작업 없음. CLI 적용 여부만 실측 |
| G5 | hook 등록 위치 분리. 앱은 user config/plugin만, CLI는 project hook에 신뢰 필요 | 구현 필요 | 앱 hooks, CLI 09-hooks | RC1은 헤드리스 경로만이라 hook 설치는 하지 않는다. 후속 hook adapter Issue에서 두 표면을 따로 다룬다 |
| G6 | 권한 mode. `--prompt` 기본이 `yolo`이고 `plan`도 sandbox가 아니다 | 구현 필요 | CLI 02-cli-reference, 05-permissions | 헤드리스 launcher가 `--mode`를 항상 명시하고 격리된 worktree에서만 `yolo`를 쓰도록 하는 후속 Issue(00-F launcher 확장 후보) |
| G7 | statusline 표시 | 미검증 | 앱 문서에 없음. CLI 출력에 개념 없음 | 대안: 공통 출력 계약의 상태 줄. 앱 실측 전에는 표시하지 않음 |
| G8 | output style | 대안 있음 | 문서에 없음 | `--output-format json\|stream-json` 결과를 공통 출력 계약으로 해석 |
| G9 | Import 대화상자 경유 설치 | 미검증 | 앱 skill 문서. 소스 경로 미기재 | projection HOLD 선택지 A. Symlink 모드와 "host별 독립 복사" 원칙의 충돌 검토 |
| G10 | 앱/CLI 문서 불일치(skill 경로, project hook) | 미검증 | 위 차이 표 | ZCode 3.17.2 실측 Issue(앱과 CLI 각각 `skills list`, hook trust status) |

## `host-capabilities` 예시

02-B/02-C가 정의할 schema(설계 22 `host-capabilities`)에 맞춘 ZCode 예시다. 실측 전이라 모든 값이 `unknown`이다. 코드는 추가하지 않았다.

```json
{
  "schema_version": 1,
  "host": "zcode",
  "host_version": "unknown",
  "profile": "local",
  "skill_discovery": "unknown",
  "explicit_invocation": "unknown",
  "binary_execution": "unknown",
  "native_agent_loading": "unknown",
  "hook_events": {},
  "test_receipts": []
}
```

## 결정이 필요한 선택지

- **P04/P15 ZCode 범위**: 결정됨(2026-11-08). 헤드리스 경로만 다루고, 수신·훅은 미확인으로 둔다.
- **ZCode projection 위치**: 상태 HOLD. 이 Issue는 고르지 않는다.
  - A. Import 대화상자로 Claude/Codex 설치본을 가져오게 안내한다. MoAI가 ZCode 경로에 쓰지 않는다. Symlink 모드는 독립 복사 원칙과 충돌할 수 있다. 앱 전용이라 헤드리스 경로에는 해당하지 않는다.
  - B. `~/.zcode/skills/moai`에 독립 복사한다. 앱 문서가 명시한 유일한 경로이고, CLI 문서도 사용자 `.zcode/skills`를 찾는다.
  - C. `.agents/skills/moai`(CLI 문서의 project·user 후보)에 둔다. Codex 등 다른 클라이언트와 경로를 공유하는 관례에 맞지만, 앱 문서에는 없다.
  - D. `.zcode-plugin/plugin.json` plugin으로 배포한다. skill과 hook을 함께 실을 수 있지만 marketplace 운영이 필요하다.
- **앱/CLI 불일치 해소 방법**(실측 시점·환경): HOLD. 실측은 ZCode 3.17.2 설치와 계정이 필요하다.

## 미실행·위험

- not_run: ZCode 앱과 CLI 실측(발견, `$moai` 호출, hook 실행, `skills list`). Import 대화상자의 실제 소스 경로와 Symlink/Copy 결과도 미확인이다.
- 앱 문서에는 발행일이 없어 stale 판정 기준이 약하다. CLI 문서는 commit으로 고정했다.
- 제3자 자료(예: `phoben/agent-skills`의 claude-to-zcode skill)는 근거로 쓰지 않았다. 그쪽의 `~/.agents/skills` 주장은 미검증 참고로만 남긴다.
- GitHub README는 web 조회에서 504가 났다. 그래서 API(`gh api repos/zai-org/ZCode/readme`, 마지막 commit `a9fe401eede6` 2026-11-03)로 읽었고, README에는 헤드리스 플래그가 없다.
