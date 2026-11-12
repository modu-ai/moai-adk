# Host 지원표 (RC1)

이 문서의 열 정의와 Codex 행은 02-C #1993, Claude Code 행은 02-B #1992, ZCode 행은 02-D #1994 소유다. 열 형식은 02-D가 먼저 정했고 02-C가 따른다(리더 결정 2026-11-08). host마다 surface를 나누고 항목마다 출처를 적는다. 출처가 둘 이상이면 ` ; `로 잇는다. 08-D #2015 release 검토가 이 표를 읽는다. Claude Code 행의 소유는 리더 결정(2026-11-08)이다.

## 상태 어휘

`capability_status`는 설계 10 §2의 capability 상태 모델을 따른다.

| 값 | 뜻 |
| --- | --- |
| `unknown` | 공식 출처가 기능을 명시하지 않는다. 경로나 동작을 추정해 적지 않는다 |
| `documented` | 공식 출처가 명시한다. 실측하지 않았다 |
| `probed` | 해당 version·OS 조합에서 실측 receipt가 있다 |
| `certified` | 실제 앱 E2E가 있다. 이 상태만 "지원"으로 표기할 수 있다 |
| `stale` | 확인 뒤 출처나 제품이 바뀌어 다시 확인해야 한다 |

`unknown` 행은 `unknown_class`에 다음 셋 중 하나를 둔다.

| 값 | 뜻 |
| --- | --- |
| `표시 불가` | 그 표면에서 해당 기능을 보여 줄 수 없다 |
| `대안` | 다른 기능으로 대신할 수 있다 |
| `시험 대기` | 실측해야 알 수 있다 |

## ZCode (02-D 초안)

ZCode에는 1차 출처가 두 가지다. surface `app`은 데스크톱 앱 문서(zcode.z.ai)이고, surface `cli`는 `zai-org/ZCode` 저장소의 CLI 사용 문서다. RC1은 헤드리스 경로(`cli`)만 다룬다(사용자 결정 2026-11-08 16:4x). 두 출처가 서로 다르게 말하는 항목은 각자의 상태로 적고, 한쪽이 다른 쪽을 덮어쓰지 않는다. 모든 행은 2026-11-08에 확인했고 version은 3.17.2다. 실측은 하지 않았다(로컬에 `zcode` 없음).

| host | surface | item | capability_status | unknown_class | 내용 | source |
| --- | --- | --- | --- | --- | --- | --- |
| ZCode | app | agents_loading | documented | | 전역 `~/.zcode/AGENTS.md` 뒤에 workspace `AGENTS.md`를 이어 붙인다. 디렉터리 단계 병합, 하위 디렉터리 스캔, `@import`/`@include` 전개는 하지 않는다. `CLAUDE.md`는 onboarding 1회 이전 소스로만 쓴다 | https://zcode.z.ai/en/docs/agents |
| ZCode | app | skill_path | documented | | 사용자 수준 `~/.zcode/skills/<skill-name>/SKILL.md`. Project 경로는 문서에 없다 | https://zcode.z.ai/en/docs/skill |
| ZCode | app | skill_limits | documented | | `name`·`description`이 필수다. description이 1024자를 넘으면 skill 전체가 빠지고, 본문이 100KB를 넘으면 잘린다 | https://zcode.z.ai/en/docs/skill |
| ZCode | app | explicit_invocation | documented | | 채팅 입력에서 `$`를 쳐 skill을 고른다 | https://zcode.z.ai/en/docs/skill |
| ZCode | app | import_dialog | documented | | Settings → Skills Import가 Claude Code·Codex CLI·OpenClaw·Augment·Windsurf의 skill 디렉터리를 Symlink/Copy, Global/Project로 가져온다. 정확한 소스 경로는 문서에 없다 | https://zcode.z.ai/en/docs/skill |
| ZCode | app | hooks | documented | | `~/.zcode/cli/config.json`(`hooks.enabled: true`)과 plugin `hooks/hooks.json`만 실행한다. project hook은 현재 버전에서 실행하지 않는다(`config_project_hooks_ignored`). `.claude/settings.json`·`.agents/settings.json`은 읽기 전용 표시다 | https://zcode.z.ai/en/docs/hooks |
| ZCode | app | plugin_manifest | documented | | `.zcode-plugin/plugin.json`을 먼저, `.claude-plugin/plugin.json`을 다음으로 찾는다. Claude Code marketplace를 미리 싣는다 | https://zcode.z.ai/en/docs/plugin |
| ZCode | app | statusline | unknown | 시험 대기 | 문서 내비게이션과 ADE Tools 페이지에 statusline이 없다 | https://zcode.z.ai/en/docs |
| ZCode | app | output_style | unknown | 대안 | 문서에 output style이 없다. 응답 형식은 공통 workflow의 출력 계약으로 대신한다 | https://zcode.z.ai/en/docs |
| ZCode | cli | agents_loading | documented | | `~/.zcode/AGENTS.md`를 먼저, workspace `AGENTS.md`를 다음에 둔다. workspace 파일은 `workingDirectory`에서 project root까지 위로 찾는다(설계 문서) | https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/design/agents-user-instructions-merge.md |
| ZCode | cli | skill_path | documented | | 사용자와 project의 `.zcode/skills`·`.agents/skills`를 기본으로 찾는다. Git 작업 트리에서는 작업 트리 루트 쪽으로 찾고, 같은 위치면 `.zcode`가 `.agents`보다 우선한다. `skills.roots`로 루트를 더할 수 있다 | https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage/08-skills.md |
| ZCode | cli | skill_limits | unknown | 시험 대기 | CLI 문서에 description·본문 길이 제한이 없다(앱 문서의 1024자·100KB가 CLI에도 적용되는지 미확인) | https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage/08-skills.md |
| ZCode | cli | explicit_invocation | unknown | 대안 | 헤드리스 호출에서 skill을 고르는 문법이 문서에 없다. `--prompt` 문장에서 skill 이름을 직접 부르는 방식은 시험이 필요하다. `zcode skills list --cwd <dir> --json`으로 로드 여부를 확인할 수 있다 | https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage/02-cli-reference.md |
| ZCode | cli | import_dialog | unknown | 표시 불가 | Import 대화상자는 앱 UI다. CLI 문서에는 없다 | https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage/08-skills.md |
| ZCode | cli | hooks | documented | | 사용자 설정·project 설정·plugin이 hook을 낼 수 있다. project hook은 기본으로 신뢰하지 않으며, `zcode hooks trust review\|grant\|status\|revoke --workspace <dir>`로 digest 단위 신뢰가 필요하다. 이벤트 7종 | https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage/09-hooks.md |
| ZCode | cli | plugin_manifest | documented | | `.zcode-plugin/plugin.json` 구조를 쓴다. 호환 manifest 필드가 런타임에 구현됐다는 보장은 없고 로드 진단이 기준이다 | https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage/10-plugins.md |
| ZCode | cli | statusline | unknown | 표시 불가 | 헤드리스 출력(`text`/`json`/`stream-json`)에는 statusline 개념이 없다 | https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage/02-cli-reference.md |
| ZCode | cli | output_style | unknown | 대안 | output style 문서가 없다. `--output-format json\|stream-json` 결과를 읽어 공통 출력 계약으로 대신한다 | https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage/02-cli-reference.md |
| ZCode | cli | headless_invocation | documented | | `zcode --prompt "<task>" --cwd <abs> --mode plan\|build\|edit\|yolo\|guarded --output-format text\|json\|stream-json [--resume <id>]`. `--prompt`의 기본 mode는 `yolo`이고, `plan`도 OS 수준 읽기 전용 sandbox가 아니다 | https://github.com/zai-org/ZCode/blob/0a1fc56d279fd1001f78387bd81aec1c26767411/apps/zcode-cli/docs/usage/02-cli-reference.md ; https://github.com/zai-org/ZCode/blob/a60ba6a7e5a06742270ca605f5d2cc2766afa7e1/apps/zcode-cli/docs/usage/05-permissions.md |

출처 고정:
- 앱 문서 페이지에는 발행일이 없다. 다운로드 링크가 3.17.2를 가리킨다.
- CLI 문서는 `zai-org/ZCode`의 commit에 고정했다. `usage/` 디렉터리 최신 commit `a60ba6a7e5a06742270ca605f5d2cc2766afa7e1`(2026-10-30), `02-cli-reference.md` 최신 commit `0a1fc56d279fd1001f78387bd81aec1c26767411`(2026-10-19)이다.
- 최신 릴리스는 `v3.17.2`(2026-11-04)다. 저장소 license는 Apache-2.0이다.

상세 인용, 앱과 CLI의 차이, adapter gap, 선택지는 `docs/zcode-support.md`에 있다.

## Claude Code (02-B 초안)

02-B #1992가 02-D 형식으로 쓴 행이다(리더 결정 2026-11-08: Claude 행은 02-B가 직접 작성). surface `cli`는 터미널 대화형 세션이다. 모든 행은 2026-11-08에 공식 문서를 읽고 확인했고, 로컬 version은 2.1.319(macOS)다. Claude Code 세션 실측은 하지 않았다(`claude -p` 실행은 운영자 토큰과 전역 hook이 걸려 사용자 결정으로 넘김). 그래서 `probed` 행이 없다. cloud 세션 행은 이번에 출처를 읽지 않아 넣지 않았다.

| host | surface | item | capability_status | unknown_class | 내용 | source |
| --- | --- | --- | --- | --- | --- | --- |
| Claude Code | cli | agents_loading | documented | | v2.1.277부터 `AGENTS.md`·`.claude/AGENTS.md`를 작업 디렉터리와 상위에서 읽는다. 같은 범위의 `CLAUDE.md`·`.claude/CLAUDE.md`·`CLAUDE.local.md` 하나라도 있으면 기본값(`claude-md-or-agents-md`)에서 읽지 않는다. `~/.claude/CLAUDE.md`·managed CLAUDE·`.claude/rules/`는 억제하지 않는다. `pluginConfigs.cc-plugin-agents-md@builtin.options.instructionFiles`(2.1.285 전 `agents-md@builtin`, 2.1.285부터는 두 id를 모두 읽음)는 user·`--settings`·managed에서만 읽고 project·local settings는 무시한다. `moai-cli claude-doctor`가 이 규칙으로 판정한다 | https://code.claude.com/docs/en/memory |
| Claude Code | cli | agents_import | documented | | CLAUDE.md의 `@AGENTS.md` import는 어떤 설정값에서도 한 번만 읽힌다. import는 최대 4단계, 상대 경로는 import한 파일 기준이다 | https://code.claude.com/docs/en/memory |
| Claude Code | cli | skill_path | documented | | personal `~/.claude/skills/<name>/SKILL.md`, project `.claude/skills/<name>/SKILL.md`, `--add-dir` 디렉터리의 `.claude/skills/`. 같은 이름이면 enterprise > personal > project | https://code.claude.com/docs/en/skills |
| Claude Code | cli | skill_reload | documented | | 세션 시작 뒤 새로 만든 최상위 skills 디렉터리는 감시하지 않으므로 `/reload-skills`가 필요하다(이후 변경도 다시). 그 밖의 `SKILL.md` 변경은 세션 안에서 바로 반영된다 | https://code.claude.com/docs/en/skills |
| Claude Code | cli | skill_limits | documented | | frontmatter 필드는 모두 선택이며 `description`만 권장한다. `description`+`when_to_use`는 목록에서 1,536자로 잘린다. 모르는 필드는 오류 없이 무시한다 | https://code.claude.com/docs/en/skills |
| Claude Code | cli | explicit_invocation | documented | | `/` 메뉴에 skill `name`(기본값은 디렉터리 이름)이 명령으로 나온다. MoAI는 `/moai` | https://code.claude.com/docs/en/skills |
| Claude Code | cli | output_style | documented | | `~/.claude/output-styles`·`.claude/output-styles`·managed. settings `outputStyle`(대소문자 구분)로 고른다. 시작 때 읽으므로 새 파일·수정은 재시작이 필요하다. 메인 대화와 fork에만 적용되고 다른 subagent에는 적용되지 않는다 | https://code.claude.com/docs/en/output-styles |
| Claude Code | cli | statusline | documented | | settings `statusLine{type:"command", command, padding, refreshInterval, hideVimModeIndicator}`. stdin JSON으로 `model.*`·`workspace.*`·`cost.*`·`context_window.*`·`session_id`·`output_style.name` 등을 받는다. 300ms debounce, 실행 중인 스크립트는 새 갱신이 오면 취소된다 | https://code.claude.com/docs/en/statusline |
| Claude Code | cli | hooks | documented | | user·project·local·managed settings, plugin `hooks/hooks.json`, skill·subagent frontmatter. exit 0은 JSON 적용, exit 2는 막을 수 있는 이벤트에서 차단, 그 밖은 비차단 오류. `PermissionRequest`는 `hookSpecificOutput.decision.behavior`(allow/deny)로만 정하고 exit 2를 따르지 않는다. 프롬프트를 띄울 수 없는 세션에서 결정이 없으면 거부한다. command hook 기본 timeout은 600초. MoAI hook은 rc.2 I1 #2238부터 `hooks-apply`가 project `.claude/settings.json` `hooks` 키에 3개(SessionStart·PostToolUse(Bash)·Stop, timeout 명시)를 생성한다. 근거: `docs/hooks-generation.md` | https://code.claude.com/docs/en/hooks |
| Claude Code | cli | goal | documented | | 내장 `/goal <condition>`. 매 턴이 끝날 때 평가 모델이 조건을 판정하고, 아직 아니면 제어를 돌려주지 않고 다음 턴을 시작한다. 조건이 충족되거나 불가능하다고 판정되거나 고칠 수 없는 오류가 나면 해제된다. 사용자가 입력한다(대화형, 또는 `claude -p "/goal …"` 비대화형 1회 실행). 모델이 스스로 goal을 거는 경로는 문서에 없다. 2026-11-09 확인(DR-08 #2145) | https://code.claude.com/docs/en/goal |

출처 고정: 공식 문서 페이지에는 발행일이 없어 읽은 날짜(2026-11-08)와 로컬 version 2.1.319로 고정한다. 판정 규칙과 계약 전문은 `docs/claude-adapter-contract.md`에 있다.

## Codex (02-C)

Codex 공식 문서는 Desktop 앱, CLI, IDE 확장을 함께 설명한다. 세 표면에 똑같이 적용된다고 쓴 항목은 surface `all`로, 표면별로 다르게 쓴 항목은 `desktop`·`cli`·`ide`로 적는다. 모든 행은 2026-11-08에 확인했다. 문서 페이지에는 version과 발행일이 없다. 로컬에서 관측한 것은 `codex --version` → `codex-cli 0.163.0` 하나뿐이고, 그 결과로 올린 행은 없다. Desktop과 CLI의 `$moai` 실측은 not_run이다. Desktop은 GUI라 레인에서 할 수 없고, CLI는 실제 호출이 OpenAI로 prompt를 보낸다. 운영자 실측 여부는 사용자 결정이다. not_run은 Desktop 발견·호출이 "실패"했다는 뜻이 아니므로 HA-10의 RC blocked 조건이 아니다.

| host | surface | item | capability_status | unknown_class | 내용 | source |
| --- | --- | --- | --- | --- | --- | --- |
| Codex | all | skill_path | documented | | 작업 디렉터리부터 저장소 루트까지 각 디렉터리의 `.agents/skills`, `$HOME/.agents/skills`, `/etc/codex/skills`, 번들 system skill을 찾는다. symlink 폴더도 따라간다. 변경은 자동 감지하고, 안 보이면 재시작한다 | https://learn.chatgpt.com/docs/build-skills |
| Codex | all | skill_required_fields | documented | | `SKILL.md`에 `name`·`description`이 필요하다. 같은 `name`의 skill은 병합하지 않고 둘 다 selector에 나올 수 있다 | https://learn.chatgpt.com/docs/build-skills |
| Codex | all | openai_yaml | documented | | `agents/openai.yaml`의 `interface.display_name`·`short_description`·`icon_small`·`icon_large`·`brand_color`·`default_prompt`, `policy.allow_implicit_invocation`(기본 `true`), `dependencies.tools`. Codex는 name·description을 먼저 싣고 쓰기로 했을 때 `SKILL.md` 전체를 읽는다 | https://learn.chatgpt.com/docs/build-skills |
| Codex | all | agents_loading | documented | | 전역은 `AGENTS.override.md`가 있으면 그것, 없으면 `AGENTS.md`(첫 비어 있지 않은 파일)이다. project는 root(보통 Git root)에서 작업 디렉터리까지 내려가며 디렉터리마다 최대 1개(override → `AGENTS.md` → `project_doc_fallback_filenames`)를 빈 줄로 잇는다. 가까운 파일이 뒤에 와서 앞 지침을 덮는다 | https://learn.chatgpt.com/docs/agent-configuration/agents-md |
| Codex | all | agents_budget | documented | | 합산 크기가 `project_doc_max_bytes`(기본 32 KiB)에 이르면 더 붙이지 않는다. config reference에는 이 기본값이 없고 AGENTS.md 페이지에만 있다. 넘는 파일을 자르는지 빼는지는 적혀 있지 않다 | https://learn.chatgpt.com/docs/agent-configuration/agents-md ; https://learn.chatgpt.com/docs/config-file/config-reference |
| Codex | all | trust_layer | documented | | `projects.<path>.trust_level`은 `"trusted"` 또는 `"untrusted"`다. untrusted project는 project의 `.codex/` 레이어(config, hooks, rules)를 건너뛴다. MoAI doctor는 config.toml을 읽지 않으므로 trust를 판정하지 않는다(`not_evaluated`) | https://learn.chatgpt.com/docs/config-file/config-reference |
| Codex | all | hooks | documented | | 주요 위치는 `~/.codex/hooks.json`, `~/.codex/config.toml`, `<repo>/.codex/hooks.json`, `<repo>/.codex/config.toml`이다. 한 레이어에 `hooks.json`과 inline `[hooks]`가 함께 있으면 병합하고 시작 때 경고한다. 비관리 hook은 검토·신뢰 뒤에만 실행한다. cloud orchestration에서는 command·local-config·plugin hook을 지원하지 않는다 | https://learn.chatgpt.com/docs/hooks ; https://learn.chatgpt.com/docs/config-file/config-reference |
| Codex | desktop | explicit_invocation | documented | | composer에 `$`를 쳐서 skill을 명시 호출한다. 켜진 skill은 slash 목록에도 나온다. 이 페이지의 slash 명령 표에는 `/skills`가 없다 | https://learn.chatgpt.com/docs/app/slash-commands |
| Codex | cli | explicit_invocation | documented | | `/skills`를 실행하거나 `$`로 skill을 부른다 | https://learn.chatgpt.com/docs/build-skills |
| Codex | ide | explicit_invocation | documented | | CLI와 같다. `/skills`를 실행하거나 `$`로 skill을 부른다 | https://learn.chatgpt.com/docs/build-skills |
| Codex | cli | status_line | documented | | `tui.status_line`은 `array<string> \| null`이고 `null`이면 status line을 끈다. 항목 ID 목록은 문서에 없어 MoAI는 어떤 ID도 쓰지 않는다(`unverified_items`) | https://learn.chatgpt.com/docs/config-file/config-reference |
| Codex | desktop | status_line | unknown | 시험 대기 | `tui.status_line`은 TUI 설정이다. Desktop 앱에 같은 표시가 있는지는 문서에 없다 | https://learn.chatgpt.com/docs/config-file/config-reference |
| Codex | ide | status_line | unknown | 시험 대기 | IDE 확장에서 `tui.status_line`이 쓰이는지는 문서에 없다 | https://learn.chatgpt.com/docs/config-file/config-reference |
| Codex | cli | goal | documented | | 내장 `/goal <objective>`(`/goal pause`·`resume`·`clear`)는 "keep working toward one durable objective instead of stopping after one normal turn"이다. Codex 0.128.0부터 있고(로컬 관측 0.163.0), `codex features enable goals`로 켠다. Claude와 달리 별도 평가 모델이 아니라 작업 중인 모델이 완료를 판정한다. 사용자가 입력하거나 prompt로 "directly set a goal and start working"을 요청할 수 있고, cookbook은 "The model can start a Goal and can mark an existing Goal complete only when the evidence supports completion"이라고 적는다(모델이 스스로 거는 경로가 있다). 별도로 Stop hook의 `decision: "block"`은 턴을 거부하지 않고 "continue and automatically creates a new continuation prompt"한다. `codex exec`에서 쓸 수 있는지는 미확인이다. 2026-11-09 확인(DR-08 #2145) | https://learn.chatgpt.com/use-cases/follow-goals ; https://developers.openai.com/cookbook/examples/codex/using_goals_in_codex ; https://learn.chatgpt.com/docs/hooks |

출처 메모:
- hooks 페이지의 handler 종류(`command`·`mcp_tool`, `prompt`·`agent`는 무시), exit 0 의미, PreToolUse 오류 때의 동작은 이번 조회에서 원문을 읽지 못했다(긴 페이지의 요약만 확보). 그래서 표에 넣지 않았다.
- `features.hooks` 기본값도 확인하지 못했다.

상세 인용, adapter 계약, gap은 `docs/codex-adapter-contract.md`에 있다.

## 론처 (`moai-cli` 세션 진입, #2164)

호스트 세션은 역할 없이 엽니다: `moai-cli cc|glm|codex`는 평범한 host 세션이고(#2053), MoAI 역할은 세션 안에서 `/moai --mode-leader`·`/moai --mode-agent`로 정합니다. 모드 표기는 이 두 개뿐입니다(사용자 결정 2026-11-10, #2286 — 옛 표기와 별칭은 폐기). 리더·에이전트 론처 플래그 `-l|-a`(`moai-cli -l|-a`, `moai-cli cc|glm|codex -l|-a`)는 세션 이름(`leader`, `agent-N`)과 전체 권한을 붙이는 형태로 유지합니다(사용자 결정 2026-11-10, Issue #2359 댓글 6760976883). #2164의 deprecated 경고는 없앴고, 역할 없는 세션에는 권한 플래그를 붙이지 않습니다. `--name`은 여전히 독립 옵션입니다. 상태: 유지(`--dry-run`과 시험으로 확인, 실제 host 실행 not_run). git 저장소가 아닌 폴더에서 처음 실행하면 묻지 않고 `git init`과 프로젝트 설정(`.moai-cli/project.yaml`)·`AGENTS.md` 생성을 먼저 합니다(이미 있는 파일은 그대로, HOME 자체·루트는 안내 후 중단, git 미설치는 OS별 설치 안내 후 중단 — 사용자 결정 2026-11-12, #2481, 상세는 `docs/rust-product-path.md` "First launch initializes the project folder").

론처가 지원하는 host는 `cc`·`glm`·`codex` 세 가지입니다. OMP — 지원하지 않음(2026-11-10 결정, Issue #2378). 같은 날 추가했던 OMP 어댑터(#2362)는 제거했고, `moai-cli omp …`는 사용법 오류(exit 2)로 끝나며 host를 띄우지 않습니다.
