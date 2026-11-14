# 호스트 훅 생성 — 거부 정책에서 '허용 목록 + 영수증'으로 (rc.2 I1 #2238)

RC1은 훅을 하나도 만들지 않았다(리더 임시 결정, `docs/decisions-2026-11-08.md` 19:1x 두 행 "hook 0개"). Claude `settings.json`의 `hooks` 키는 관리 대상이 아니었고, Codex projection은 `"hooks": {"generated": []}`를 보고했다. 이 문서는 rc.2 I1 카드가 그 정책을 바꾸는 근거와 범위를 적는다.

## 1. 바뀐 것

| 항목 | RC1 | rc.2 I1 |
| --- | --- | --- |
| Claude `hooks` 키 | 관리 대상 아님(`MANAGED_SETTINGS`에 없음) | 관리 대상. 키 전체를 소유하고 `hooks-apply`만 쓴다 |
| Codex `.codex/hooks.json` | 생성하지 않음 | `hooks-apply --host codex`가 파일 전체를 소유하고 쓴다 |
| 쓰는 내용 | — | 고정된 MoAI 훅 집합 하나(허용 목록). 임의 훅은 받지 않는다 |
| 기록 | — | 생성한 훅 목록을 영수증에 남긴다(비밀 0) |
| 쓰기 조건 | — | dry-run `hooks-plan` 뒤, 같은 `plan_digest`와 permission reference가 있을 때만 |

skill frontmatter의 `hooks` 필드 거부(`src/claude.rs` `CLAUDE_FIELDS`)는 그대로 둔다. package 내용이 훅을 등록하는 경로는 여전히 없다.

## 2. 왜 이 변경이 권한을 넓히지 않는가

1. **두 호스트 모두 project 훅을 신뢰 뒤에만 실행한다.**
   - Codex: "Project-local hooks load only when the project `.codex/` layer is trusted." "Non-managed hooks must be reviewed and trusted before they run." "Codex records trust against the hook's current hash, so new or changed hooks are marked for review and skipped until trusted." 즉 MoAI가 쓴 훅도 사용자가 `/hooks`에서 검토하기 전에는 돌지 않는다.
   - Claude Code: "Frontmatter hooks in a project skill follow the same workspace trust rule as hooks in settings files." "Project subagent hooks run only after you accept the workspace trust dialog." settings 파일 훅의 신뢰 규칙 본문은 permissions 문서에 있고, 이번에는 그 문서를 읽지 않았다(아래 §6 미확인).
2. **허용 목록이 고정이다.** 생성기는 이 문서 §3의 훅만 쓴다(Codex 4개, Claude Code 5개). 요청으로 훅을 더하거나 바꿀 수 없다.
3. **기록 훅 3개는 기록만 하고, 차단 훅은 하나뿐이다.** 기록 훅은 `moai-cli hook-receipt`를 부르고 차단하지 않는다(항상 exit 0). 도구 입력·출력·prompt는 저장하지 않는다. 차단 훅은 `PreToolUse`(Bash)의 `moai-cli claim-verify . --hook` 하나다(rc.2 A9 #2255 AC-2). Bash 구간의 첫 단어가 `git`이고 `commit`이나 `push`를 실행하는데, 그 레인 worktree의 claim이 확인되지 않을 때만 exit 2로 막는다. 나머지 호출은 모두 통과시킨다. 권한을 넓히지 않고 쓰기 경계를 좁히기만 한다.
4. **사용자 설정을 덮어쓰지 않는다.** Claude는 기존 3-way 소유 규칙을 그대로 쓴다. 사용자가 이미 `hooks` 키를 두었으면 plan이 `conflict`가 되고 아무것도 쓰지 않는다. Codex는 MoAI 영수증과 sha256이 맞지 않는 `.codex/hooks.json`을 만나면 `requires preservation review`로 멈춘다.
5. **홈은 쓰지 않는다.** project가 HOME 자체이거나 HOME의 `.claude`·`.codex`·`.agents`·`.moai-cli` 안이면 거부한다. 훅은 project 트리에만 생긴다.

## 3. 훅 집합

| event | matcher | handler |
| --- | --- | --- |
| `SessionStart` | 없음 | `moai-cli hook-receipt --host <host>`, timeout 10초 |
| `PreToolUse` | `Bash` | `moai-cli claim-verify . --hook`, timeout 10초(#2255 claim 쓰기 경계, 차단 exit 2) |
| `PreToolUse`(Claude Code 만) | `Bash\|Write\|Edit\|MultiEdit\|NotebookEdit` | `moai-cli hook-dispatch --host claude --event tool-preflight`, timeout 10초 — 훅 엔진의 위험 Bash·`.git`·자격 증명 폴더·개인 키·토큰 쓰기 가드, 문서화된 PreToolUse `deny` 로만 답함(#2877 리더 Q1: Go 세대 hook 을 바꾼 뒤에도 그 보호가 남게) |
| `PostToolUse` | `Bash` | `moai-cli hook-receipt --host <host>`, timeout 10초 |
| `Stop` | 없음 | 같음 |

- 두 호스트가 모두 받는 교집합만 쓴다: 이벤트 4개와 handler 필드 `type: "command"`·`command`·`timeout`. `PreToolUse`는 두 호스트 모두 matcher `Bash`, stdin `tool_name`·`tool_input.command`·`cwd`, exit 2 + stderr 사유로 차단한다(§5 출처). guard의 `.`은 project다. 두 호스트 모두 command 훅을 세션의 작업 디렉터리에서 실행한다.
- guard가 막지 못하는 경우: timeout(10초) 안에 GitHub 조회가 끝나지 않으면 호스트는 훅 실패로 보고 막지 않는다(Claude: timeout은 `onFailure: "block"`이 없으면 차단하지 않음). Codex 문서는 "Treat tool hooks as a useful guardrail, not a complete enforcement boundary"라고 적는다. 그래서 MODE-AGENT workflow 6단계의 commit 전 `claim-verify` 실행 규칙은 그대로 유지한다.
- guard가 알아보는 commit·push 철자(#2306): 명령줄을 글자로 읽어 판정한다.
  - 알아보는 것:
    - 앞에 붙은 래퍼 단어와 그 옵션(`env`·`command`·`nice`·`sudo`·`time`·`timeout`·`nohup`·`xargs` 등)
    - 따옴표와 역슬래시로 감싼 `git`(`"git"`, `\git`)
    - `sh -c` / `bash -c` 안의 명령 한 단계 — 값을 받는 셸 옵션(`-o pipefail`, `-O extglob`, `+o posix`, `+O`) 뒤의 `-c`도 알아본다(rc.2 A9 후속 2 #2323)
    - 같은 줄의 `git -c alias.<이름>=commit|push`
  - 알아보지 못하는 것(남은 한계):
    - git 설정에 정의된 alias, 셸 alias·함수
    - `$VAR`·`$(...)` 치환, ANSI-C 따옴표 `$'git' push`(= `$` 치환으로 알아보지 못한다, rc.2 A9 후속 2 #2323에서 이름 붙임), `eval`
    - 스크립트 파일 안의 git, 두 단계 이상 중첩된 셸
    - git을 스스로 부르는 다른 프로그램
  - 그래서 이 guard는 경계가 아니라 guardrail이다.
- `timeout`은 모든 handler에 명시한다. 호스트 기본값(Claude command 600초, Codex 대부분 600초)에 기대지 않는다.
- `command`는 bare 이름이다. 개인 절대 경로가 들어가지 않는다.
- PATH 해석(#2348): 호스트가 훅 프로세스에 주는 PATH에서 `moai-cli`를 찾지 못하면 그 훅은 실패한다. 명령 기본 이름은 `moai`가 됐지만(#2812 W3) 생성 훅은 별칭 은퇴(R5) 전까지 `moai-cli`를 부른다: Go 세대 `moai`가 PATH에서 먼저면 `moai claim-verify …`를 Go가 받아 exit 1로 끝나고, 호스트는 그것을 막지 않는 오류로 읽어 claim 가드가 조용히 꺼지기 때문이다(감사 6868377245 B1, 리더 결정 6868439066).
  - macOS·Linux, 그리고 Git Bash가 있는 Windows에서는 비차단 오류로 끝나고 동작은 진행된다(Claude: "Non-blocking error… The action goes ahead"). 이때 claim guard는 돌지 않는다.
  - **Git Bash가 없는 Windows는 다르다.** Claude 문서에 따르면 shell 형식 `PreToolUse` 훅이 `onFailure` 없이 0·2 외의 코드로 실패하면 도구 호출이 막힌다("Failed hooks on Windows without Git Bash"). 생성 훅은 shell 형식이고 `onFailure`를 두지 않으므로, 그 환경에서는 `moai-cli`가 없을 때 모든 Bash 호출이 막힐 수 있다. exec 형식(`args`)과 `onFailure` 값은 guard 정책 결정(D5)·훅 소유 방식(D7)과 함께 정하며, 그 전까지 HOLD다.
  - `hooks-plan`·`hooks-apply`는 자기 PATH에서 `moai-cli`를 찾았는지 `hook_binary.state`(`resolved` / `binary_unresolved`)로 보고한다. 못 찾으면 `hooks_active: false`, 찾아도 호스트 영수증이 있기 전에는 `hooks_active: "not_run"`이다. 호스트 훅 프로세스의 PATH는 셸과 다를 수 있으므로 `resolved`는 "이 PATH에서 찾았다"는 뜻일 뿐이다.
- 훅 모드 출력과 종료 코드(#2348):
  - `hook-receipt`와 `claim-verify --hook`은 성공(exit 0)할 때 stdout에 아무것도 쓰지 않는다. 두 호스트 모두 stdout의 JSON 객체를 훅 출력으로 읽기 때문이다. 다른 호출은 envelope를 그대로 쓴다.
  - exit 2(차단)는 guard의 거부 하나뿐이다. 사용법 오류(알 수 없는 플래그·단어)와 `hook-receipt`의 오류는 exit 1(비차단 오류)이고, 사유는 stderr에 쓴다.
  - guard가 검사하지 않는 명령(커밋·push가 아님)은 transport 설정을 만들지 않는다. 그래서 잘못된 `MOAI_FORGE_GH_BIN`이 모든 Bash 호출을 막지 않는다.
  - guard가 검사하는 명령(커밋·push)에서 판정을 내리지 못하면 — 잘못된 `MOAI_FORGE_GH_BIN` 같은 설정 오류나 조회 실패 포함 — guard의 거부로 보고 exit 2로 막는다(fail-closed). 검증하지 못한 커밋을 통과시키지 않기 위해서다. 이 경우를 비차단으로 바꿀지는 guard 정책 결정(D5)에 속하며, 그 전까지 현행을 유지한다.
  - 남은 한계: 이미 배포된 옛 `moai-cli`가 PATH에 있으면, 그 바이너리가 모르는 단어·플래그로 exit 2를 낼 수 있다. 이 동작은 이 저장소의 새 바이너리로는 바꿀 수 없다.
- Codex 파일에는 선택적 top-level `description`을 넣는다("optional top-level metadata... It doesn't change which hooks run").

## 3a. 명령 경로와 trust — 세 가지 선택지와 현재 선택 (#2391 R3)

#2391 본문은 "버전 경로(업데이트마다 재신뢰) 대 고정 링크(변경 미탐지)의 선택과 이유"를 요구한다. 실제 생성기는 둘 중 어느 것도 아닌 세 번째 안, **PATH의 bare 이름 `moai-cli`** 를 쓴다(`src/hooks.rs` `hook_command`·`HOOK_BINARY`, #2348). 아래 표는 세 안을 같은 기준으로 비교한다.

| 안 | `command` 예 | 업데이트 뒤 재신뢰 | 실행체가 바뀐 것을 host가 알아채는가 | 개인 경로 노출 |
| --- | --- | --- | --- | --- |
| 버전 경로 | `~/.moai-cli/versions/<v>/moai-cli …` | 필요(업데이트마다 hook 정의가 바뀐다) | 정의가 바뀌므로 Codex는 재검토를 요구한다 | 있음(절대 경로) |
| 고정 링크 | `~/.moai-cli/bin/moai-cli …`(링크 대상만 교체) | 불필요 | 아니다(정의가 그대로) | 있음(절대 경로) |
| **PATH bare 이름(현재)** | `moai-cli …` | 불필요 | 아니다(정의가 그대로) | 없음 |

- **현재 선택과 이유**: bare 이름. ① 생성 파일(`.claude/settings.json`·`.codex/hooks.json`)은 project 트리에 들어가므로 개인 절대 경로를 넣지 않는다(§3). ② 업데이트마다 사용자가 hook을 다시 검토해야 하는 부담이 없다.
- **대가**: 고정 링크와 같이, 같은 이름 아래 실행체가 바뀌어도 host 재검토는 일어나지 않는다. 또 host가 hook 프로세스에 주는 PATH에 따라 다른 `moai-cli`가 실행될 수 있다(§3 PATH 해석, 옛 바이너리 한계). 이것을 막는 장치는 지금 없다.
- **Codex hash가 무엇을 덮는가**: 문서는 "Codex requires you to review and trust the exact hook definition", "Codex records trust against the hook's current hash, so new or changed hooks are marked for review and skipped until trusted"라고만 적는다. hash가 정의(명령 문자열·JSON 항목)에 걸리는지 실행 파일에 걸리는지는 **적혀 있지 않다(미확인)**. 실행 파일 변경에 재검토가 붙는다고 명시된 경우는 plugin 번들 훅뿐이다: "If a plugin update changes a bundled script, Codex requires a new review even if the hook definition stays the same." MoAI 생성 훅은 plugin 번들이 아니므로 이 문장에 기대지 않는다. 위 표의 "알아채지 않는다"는 hash가 정의에 걸린다는 가정 위의 판단이며, 실제 Codex로 확인하지 않았다(not_run).
- **Claude Code**: hooks 페이지에는 hash나 재검토 문장이 없다. settings 파일 훅의 신뢰 규칙("workspace trust")은 링크된 다른 절에 있고 이번에도 원문을 읽지 못했다(미확인, §6).
- **명령 실행 방식**: Claude는 `args`가 없으면 `command`를 "passed to a shell: `sh -c` on macOS and Linux"로, `args`가 있으면 "resolves `command` as an executable on `PATH`"로 실행한다. Codex는 "Commands run with the session `cwd` as their working directory"만 적고 shell과 PATH 조회 방식은 적지 않는다.
- **timeout**: Claude는 "Claude Code cancels a `command`, `http`, or `mcp_tool` hook that reaches its timeout, discarding the hook's output", "On `PreToolUse`, a timed-out … hook doesn't block the tool call"이며, 막으려면 `"onFailure": "block"`(v2.1.295 이상)이 필요하다. Codex는 일반 command hook의 timeout 결과를 적지 않는다(`SessionEnd`에서 "reports it as a hook failure"만 있음). 엔진 쪽 결정은 fail-open(D1, `docs/decisions-2026-11-08.md` 17:52Z(11-10) 행)이고, 모의 host 시험은 `tests/hook_dispatch_cli.rs` `a_timed_out_preflight_is_non_blocking_by_the_host_rules`다.
- 버전 경로·고정 링크로 바꾸는 일, 실행체 hash를 생성 정의에 넣는 일은 이 절의 범위가 아니다(설치 경로·trust 결정 필요).

출처 조회: 2026-11-12T00:0xZ(UTC), https://code.claude.com/docs/en/hooks (HTTP 200, 293,021자 중 Timeouts·"Block the action when a hook fails"·Exec form 절 원문, Workspace trust 절 미열람), https://learn.chatgpt.com/docs/hooks (HTTP 200, 79,364자 전문). 조회 도구가 응답 시각을 따로 주지 않아 조회 직후 `date -u` 값으로 적는다.

## 4. 영수증

- Claude: `.moai-cli/manifests/claude-settings.json`의 `owned.hooks`가 생성한 훅 집합이다(기존 settings 영수증, pending 복구 포함).
- Codex: `.moai-cli/manifests/codex-hooks.json` — `schema_version`, `hooks_path`, `sha256`, `hooks`(event·matcher·command·timeout 목록). apply는 먼저 `pending_sha256`을 적고 파일을 쓴 뒤 영수증을 확정한다. 중간에 끊기면 다음 plan이 sha256으로 판정한다.
- `hook-receipt` 실행 기록: `.moai-cli/receipts/hooks.jsonl` 한 줄에 `schema_version`·`host`·`unix_ms`와 허용 필드 `hook_event_name`·`tool_name`·`source`만 남긴다. 128자 초과·제어 문자·비밀 모양 값은 버리고 `dropped`로 알린다. `.moai-cli/`가 없는 디렉터리에서는 아무것도 쓰지 않는다. 로그가 4 MiB를 넘으면 더 쓰지 않는다(자르지 않는다).

## 4a. `moai update` 의 훅 층 (#2853)

- 대상: `moai update` 를 실행한 폴더가 속한 git 작업 트리 하나, 그중 MoAI 훅 집합이 **이미 설치된** host 만(Claude 는 settings 영수증이 `hooks` 를 소유, Codex·OMP 는 소유 영수증이 있음). `update` 는 훅을 새로 설치하지 않는다 — 설치는 계속 `hooks-plan` → `hooks-apply`. 프로젝트 밖이거나 설치된 host 가 없으면 `hooks.skipped` 에 이유 한 줄.
- 미리보기: host 마다 `hooks-plan` 과 같은 규칙으로 `create`·`update`·`unchanged`·`conflict`. 생성 명령은 `moai-cli`(#2812 결정 (나)). 사용자가 바꾼 항목은 덮지 않는다 — MoAI 가 바꿀 것이 없으면 그대로 두고(`unchanged`), 바꿀 것이 있으면 `conflict` 로 보고만 한다. Codex·OMP 파일을 손으로 바꿨으면 `conflict`(이유: preservation review).
- 확인: 훅 층의 digest 는 `moai update --confirm` 이 받는 하나의 `plan_digest` 에 들어간다. 확인 때 host 마다 `hooks-apply` 규칙으로 쓰고, 바꾼 두 파일(host 파일·소유 영수증)의 앞 바이트를 `.moai-cli/receipts/hooks-<id>/` 에 남긴다. 결과는 `hooks_result`, 되돌리기 명령은 `revert.hooks`.
- 되돌리기: `moai hooks-revert <PROJECT> --receipt <id> --permission-reference <ref>` — host 의 두 파일이 모두 update 가 쓴 바이트 그대로일 때만 앞 바이트로 되돌린다(만든 파일은 지움). 하나라도 바뀌었으면 그 host 는 `user_changed_since_apply` 로 건너뛴다. 영수증은 한 번만 되돌린다.

## 4b. 레인 Stop hook — Codex opt-in 변형 (#2904 2단계)

- 설치: `hooks-plan --host codex --lane-loop <lane> --repo <OWNER/REPO> <PROJECT>` 로 미리 보고, 같은 플래그와 그 `plan_digest` 로 `hooks-apply`. `.codex/hooks.json` 의 `Stop` 에 영수증 그룹 다음 두 번째 matcher 그룹 `{"hooks":[{"type":"command","command":"moai-cli lane-stop-hook --lane <lane> --repo <repo>","timeout":120}]}` 이 생긴다. `--lane-loop` 과 `--repo` 는 함께만, `--host codex` 에서만 받는다(레인 이름 1–64자 `A-Z a-z 0-9 . _ -`, 저장소 `OWNER/REPO`).
- 영수증: 레인이 있을 때만 `lane_loop: {lane, repo}` 와 `hooks` 끝의 레인 항목이 더해진다. 레인 없는 설치는 hook 파일·영수증·plan digest 가 전과 바이트까지 같다(`hooks::tests` 가 이전 코드의 digest 로 묶어 둠). digest 에는 레인이 있을 때만 레인이 들어간다.
- 유지: 플래그 없는 `hooks-plan`·`moai update` 는 영수증이 적은 레인을 그대로 둔다(재계획 `unchanged`). `hook_status.configured` 도 영수증의 세트와 비교하므로 설치된 레인 hook 은 `yes`.
- 되돌리기: 같은 두 단계를 `--lane-loop-off` 로 — 파일과 영수증이 기본 세트 바이트로 돌아간다. `--lane-loop-off` 는 `--lane-loop` 과 함께 쓸 수 없다.
- `lane-stop-hook --lane <lane> --repo <repo> [--snapshot FILE]`: envelope 없이 Codex 의 Stop 답만 stdout 에 낸다. Stop 입력(stdin)은 64 KiB 까지 읽고 버린다 — `stop_hook_active` 는 보지 않고(리더 결정 대기 Q1 기본 (a)), 깨진 JSON 도 판정을 바꾸지 않는다. `lane-next` 판정을 한 번(기다림 없음, GitHub 읽기 90초 한도) 또는 `--snapshot` 으로 읽어:
  - 레인의 다음 일이 있으면 `{"decision":"block","reason":…}` — 이유는 고정 문장과 `lane-next` 요약 줄뿐(GitHub 글 없음).
  - 내 일은 없고 열린 PR 이 남았으면 block, 이유는 `moai lane-next --lane <lane> --repo <repo> --wait 1800` 실행(Q2 기본 (a)).
  - 열린 PR 0 이고 시작할 Issue 0(끝 조건, 지금은 `lane-next` 가 읽은 그대로)이면 아무것도 내지 않고 exit 0.
  - 읽기 실패·시간 초과·일부만 읽고 다음 일 없음 = 모름 → 한 번 block, 다음 멈춤은 통과.
  - 같은 다음 일에 연속 48번 block 하면 49번째부터 통과. 다음 일이 바뀌면 0부터. 횟수는 `<state root>/lane-stop-hook/<lane>.json`; state root 가 없거나 링크라 셀 수 없으면 block 하지 않는다.
  - 레인·저장소 형식이 틀리거나 명령줄이 깨지면 hook 비차단 오류(exit 1, stdout 비움).

## 5. 출처 (확인일 2026-11-09)

| 문서 | URL | 이번에 쓴 사실 |
| --- | --- | --- |
| Claude Code Hooks reference | https://code.claude.com/docs/en/hooks | 3단 구조(event → matcher group → handler), handler 공통 필드(`type` 필수, `timeout` 초 단위, command 600초 기본), command 필드(`command` 필수), project `.claude/settings.json`, 종료 코드 의미, frontmatter·subagent 훅의 신뢰 규칙 |
| Claude Code hooks guide | https://code.claude.com/docs/en/hooks-guide | file watcher가 settings 훅 변경을 반영한다 |
| Codex Hooks | https://learn.chatgpt.com/docs/hooks (`hooks.md`) | `<repo>/.codex/hooks.json` 위치, 같은 3단 구조, `description` 선택 필드, matcher 규칙, `timeout` 기본(대부분 600초, `SessionEnd`·`Interrupt`는 1초·최대 3초), project 훅 신뢰 조건, hash 신뢰, 기본 활성(`[features] hooks = false`로 끔) |

두 페이지 모두 "last updated" 날짜를 표시하지 않는다.

## 6. 미확인·후속

- **미확인**: Claude settings 파일 훅의 신뢰 규칙 원문(permissions 문서 미열람). Codex 생성 JSON schema(`codex-rs/hooks/schema/generated`) 대조. 두 페이지 일부 구간은 요약으로만 받았고, 위 표에는 원문으로 받은 사실만 적었다.
- **not_run**: 실제 Claude·Codex 세션에서 생성된 훅이 로드·실행되는지(`host_loaded: not_run`). 실세션은 이 카드에서 금지다.
- **후속 1 — PR 보호 guard**: #2255(A9)가 `PreToolUse` claim guard를 훅 집합에 더했다(위 §3). 그 밖의 차단 정책은 여전히 별도 결정이 필요하다.
- **Codex `PreToolUse` 원문(2026-11-09)**: `tool_input.command`, `cwd`, exit 2 + stderr, matcher `Bash`는 Codex hooks 페이지 후반부 요약으로 받았다(원문 전문을 다 받지 못함). 생성 JSON schema 대조는 미확인이다.
- **후속 2 — 항목 단위 병합**: 사용자 훅과 MoAI 훅을 같은 `hooks` 키 안에서 항목 단위로 합치려면 두 번째 소유 모델과 영수증이 필요하다. 지금은 키 전체 소유라 사용자 훅이 있으면 conflict로 멈춘다.
