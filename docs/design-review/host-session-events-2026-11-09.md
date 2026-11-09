# 호스트 세션·이벤트 메커니즘 조사 — Claude Code·Codex (2026-11-09)

- Issue: #2263 [rc.2][A12] (상위 EP-A #2153). 근거 지시: ADK_AUTONOMOUS_LANE_LOOP_20261109_01 §4. 이 카드는 문서만 다룬다. 코드는 0줄이고, 실세션은 실행하지 않았다.
- 작성: agent-1 [0dd806], 세션 `113e4ddd-566b-4bd7-b27d-f6cc8c19dcf1`, 기준 tip `5db3069a94effe4ae330137dc96ff3a01c5be222`
- 확인일: 2026-11-09. 측정 시각은 `date -u` → `2026-11-09T05:18:57Z`다.
- 로컬 host 버전: `claude --version` → `2.1.319 (Claude Code)`, `codex --version` → `codex-cli 0.163.0`. 둘 다 버전만 출력하고 모델을 부르지 않는다.
- 출발점: 최신 설계 정본 스냅샷 `references/space-pages/moai-adk-final-design-page-5490c4-seq44-20261109-1411.md`(저장소에 넣지 않는 사본, 1,377행, sha256 `234e1d470da97adbfbd4047ebe9f2b24e7ca9d9bbf90e5e37bc4379b0b4c827d`, seq44)의 §25(L1268–1377)를 출발점으로 삼았다. 정본에 인용된 공식 URL은 직접 다시 읽었다.

## 0. 인용 방식

공식 문서는 WebFetch로 읽었다. 페이지가 길면 도구가 앞부분만 원문으로 주고 나머지는 요약한다. 그래서 표마다 출처의 성격을 표시한다.

- **원문**: 도구가 원문 텍스트로 돌려준 구간을 인용했다.
- **요약**: 원문 구간 밖이라 도구 요약에서 얻었다. 사실로 쓰기 전에 다시 확인해야 한다.
- **확인 불가**: 읽은 범위에 근거가 없다.

Claude hooks 페이지(약 29만 자)는 0–80k와 140k–180k를 원문으로 읽었고, 나머지는 요약이다. Codex 페이지들은 쪽마다 앞부분이 원문이고 뒷부분이 요약이다.

공식 페이지에는 기능별 도입 버전이 적힌 곳이 일부뿐이다. 적혀 있는 것만 버전 칸에 옮겼다.

## 1. 항목 표 (AC-1)

### 1.1 Claude Code (로컬 2.1.319)

| 항목 | 공식 근거 | 출처 성격 · URL |
| --- | --- | --- |
| 세션 ID (hook) | 공통 입력 `session_id` "Current session identifier" | 원문 · https://code.claude.com/docs/en/hooks |
| 세션 ID (headless) | "`json`: structured JSON with result, session ID, and metadata". 예: `jq -r '.session_id'` | 원문 · https://code.claude.com/docs/en/headless |
| 세션 ID 지정 | `--session-id` "Use a specific session ID for the conversation (must be a valid UUID)" | 요약 · https://code.claude.com/docs/en/cli-reference |
| 세션 ID 환경 변수 | 일반 세션용은 없다. `$CLAUDE_CODE_BRIDGE_SESSION_ID`는 Remote Control 연결 중일 때만 있다 | 원문 · hooks |
| 세션 이름 | `--name`/`-n` "Set a display name for the session, shown in `/resume` and the terminal title. You can resume a named session with `claude --resume <name>`." `/rename`은 세션 도중에 이름을 바꾼다 | 원문 · cli-reference |
| hook 이벤트 목록 | 33개: SessionStart, Setup, UserPromptSubmit, UserPromptExpansion, PreToolUse, PermissionRequest, PermissionDenied, PostToolUse, PostToolUseFailure, PostToolBatch, Notification, MessageDisplay, SubagentStart, SubagentStop, TaskCreated, TaskCompleted, Stop, StopFailure, TeammateIdle, InstructionsLoaded, ConfigChange, CwdChanged, DirectoryAdded, FileChanged, WorktreeCreate, WorktreeRemove, PreCompact, PostCompact, PreModelSwitch, PostModelSwitch, Elicitation, ElicitationResult, SessionEnd | 원문(lifecycle 표) · hooks |
| 공통 입력 필드 | `session_id`, `prompt_id`, `transcript_path`("written asynchronously and may lag"), `cwd`, `scratchpad_dir`(v2.1.257+), `permission_mode`, `effort`, `hook_event_name`, `agent_id`, `agent_type`("Present when the session uses `--agent` or the hook fires inside a subagent"). "Only SessionStart hooks can receive a `model` field, and Claude Code doesn't always include it." | 원문 · hooks |
| SessionStart 출처 값 | `startup`, `resume`, `clear`, `compact`, `fork` | 원문(matcher 표) · hooks. 필드 이름 `source`는 요약 |
| SessionStart 컨텍스트 주입 | "`SessionStart` … Claude Code adds plain-text stdout as context that Claude can see and act on." | 원문 · hooks |
| SessionEnd 사유 값 | `clear`, `resume`, `logout`, `prompt_input_exit`, `other` | 원문(matcher 표) · hooks. 필드 이름 `reason`은 요약 |
| 완료 신호 | "`Stop` \| When Claude finishes responding", "`SubagentStop` \| When a subagent finishes", "`StopFailure` \| When the turn ends due to an API error" | 원문 · hooks |
| Stop 차단 | "Stop, SubagentStop, TaskCompleted, and TeammateIdle hooks: exit code 2 on these events sends Claude back to keep working". 마지막 응답 텍스트는 `last_assistant_message`로 읽으라고 한다. `stop_hook_active`와 continuation 상한은 요약에만 있다 | 원문 · hooks |
| TeammateIdle | 정본 L1347: idle 진입 직전에 발생하고 업무 완료의 증명이 아니다. exit 2는 계속 작업시키고, `{"continue":false}`는 teammate를 중지시킨다 | 정본 seq44 L1347(원 URL https://code.claude.com/docs/en/hooks#teammateidle). 이번에는 정본 인용만 확인했다 |
| worktree 진입 | "Pass `--worktree` or `-w` with a name to create an isolated worktree and start Claude in it." 위치는 `.claude/worktrees/<name>/`, branch는 `worktree-<name>`. 세션 도중에는 "creates one with the `EnterWorktree` tool"이다. v2.1.233 전에는 `--worktree` 인자 형식이 제한됐다 | 원문 · https://code.claude.com/docs/en/worktrees |
| worktree 이탈 | "Claude can leave with the `ExitWorktree` tool when the session has that tool." | 원문 · worktrees |
| worktree hook | WorktreeCreate "When a worktree is being created via `--worktree`, `isolation: "worktree"`, or for a background session. Replaces default git behavior." WorktreeRemove는 WorktreeCreate hook이 만든 worktree를 지울 때 발생한다. hook 입력 `cwd`는 worktree root다. Bash로 `git worktree add`를 하면 WorktreeCreate가 돌지 않는다(요약) | 원문 · hooks, worktrees |
| 재개 | `--continue`/`-c` "Load the most recent conversation in the current directory…"(끝난 background 세션은 v2.1.257+). `--resume`/`-r` "Resume a specific session by ID or name…". `--fork-session` "create a new session ID instead of reusing the original" | 원문 · cli-reference |
| 재개 시 이벤트 | SessionStart "When a session begins or resumes". 재개에는 matcher `resume`이 붙는다. "`SessionStart` at launch, including with `--continue` or `--resume` … fire before the session's MCP servers are available to hooks." | 원문 · hooks |
| 반복 실행 `/loop` | 단위는 `s/m/h/d`다. self-paced는 "picks a delay between one minute and one hour". "Recurring tasks automatically expire 7 days after creation." "A session can hold up to 50 scheduled tasks at once." 자체 종료는 `ScheduleWakeup`의 `stop: true`다. `loop.md`는 25,000 byte를 넘으면 잘린다. 관련 버전: v2.1.248, v2.1.295, v2.1.299 | 원문 · https://code.claude.com/docs/en/scheduled-tasks |
| 목표 `/goal` | "After each turn, a model checks whether the condition holds." 평가 모델은 "doesn't run commands or read files independently". 조건은 최대 4,000자다. 조건이 충족되거나, 불가능하다고 판정되거나, 사용자가 고쳐야 할 오류가 나면 해제된다. idle check-in은 최대 3회이고, 자동 재시도 3회 뒤에는 일시정지한다. 관련 버전: v2.1.236, v2.1.239, v2.1.246, v2.1.269 | 원문 · https://code.claude.com/docs/en/goal |
| `--agent` | "Specify an agent for the current session (overrides the `agent` setting)." 예: `claude --agent my-custom-agent`. 이 agent가 main thread가 되는지는 적혀 있지 않다 | 원문 · cli-reference |

### 1.2 Codex (로컬 codex-cli 0.163.0)

| 항목 | 공식 근거 | 출처 성격 · URL |
| --- | --- | --- |
| 세션 ID (hook) | `session_id` "Current Codex session id. Subagent hooks use the parent session id." | 원문 · https://learn.chatgpt.com/docs/hooks |
| 세션 ID (exec) | `{"type":"thread.started","thread_id":"…"}` 예시가 있다. `thread_id`와 `session_id`의 관계는 정의되어 있지 않다 | 요약 · https://learn.chatgpt.com/docs/non-interactive-mode |
| 세션 ID 환경 변수 | 없다. environment-variables 페이지 전체를 읽었는데 session·thread 변수가 없다 | 원문 · https://learn.chatgpt.com/docs/config-file/environment-variables |
| 세션 이름 | `codex resume`의 `SESSION_ID`가 "(uuid or session name)"을 받는다. 이름을 붙이는 방법은 적혀 있지 않다. 런처는 이름을 첫 prompt에 실어 보낸다(`src/session.rs:14-15`) | 요약 · https://learn.chatgpt.com/docs/developer-commands?surface=cli |
| hook 이벤트 목록 | 14개. 턴 중: `PreToolUse`, `PermissionRequest`, `PostToolUse`, `PreCompact`, `PostCompact`, `UserPromptSubmit`, `SubagentStop`, `Stop`. 모델 요청마다: `PreModelRequest`, `PostModelRequest`. 그 밖에 `Interrupt`(subagent에서는 안 돎), `SessionStart`, `SubagentStart`, `SessionEnd`(subagent에서는 안 돎) | 원문 · hooks |
| 공통 입력 필드 | `session_id`, `transcript_path`(`string \| null`, "the transcript format isn't a stable interface for hooks"), `cwd`, `hook_event_name`, `model`. 턴 단위 hook에는 `turn_id`가 붙는다 | 원문 · hooks |
| SessionStart 출처 값 | "`startup`, `resume`, `clear`, `compact`, or `fork`" | 원문 · hooks |
| SessionEnd | 열려 있는 대화를 보관하거나 지울 때, Codex가 정상 종료할 때, 연결된 클라이언트 없이 30분 동안 idle일 때 돈다. "For now, `reason` is always `other`." 대화에서 다른 대화로 전환해도 세션이 바로 끝나지는 않는다 | 원문 · hooks |
| 완료 신호 (hook) | `Stop`은 "During a turn" 이벤트로 분류돼 있다. 정확히 언제 발생하는지는 적혀 있지 않다 | 원문 · hooks. 발생 조건은 확인 불가 |
| Stop 차단 | "For this event, `decision: "block"` doesn't reject the turn. Instead, it tells Codex to continue and automatically creates a new continuation prompt … using your `reason` as that prompt text." exit 0이면 stdout에 JSON이 있어야 한다 | 원문 · hooks |
| 완료 신호 (app-server) | 정본 L1351: `turn/completed`의 completed/interrupted/failed를 관찰한다. `turn/interrupt` 응답만으로는 종료를 확정하지 않는다 | 정본 seq44 L1351(원 URL https://learn.chatgpt.com/docs/app-server). 이번에는 정본 인용만 확인했다 |
| worktree | `--worktree` "Start in a new managed Git worktree. Supports new interactive sessions, `codex fork <SESSION_ID>`, `codex exec`, and `codex exec fork`; requires local execution." `codex resume`은 `--worktree`를 받지 않는다(요약). `--ephemeral`·`--ignore-user-config`와 함께 쓸 수 없다(요약). 앱은 `$CODEX_HOME/worktrees`에 만들고 최근 15개를 유지한다(요약) | 원문(`--worktree` 행) · developer-commands. 그 밖은 요약 · https://learn.chatgpt.com/docs/environments/git-worktrees |
| worktree 이탈 | 공식 이탈 명령이나 도구는 확인되지 않았다. `/worktree` slash 명령이 있다는 요약만 있다 | 확인 불가 |
| 재개 | `codex resume [--last] [--all] [SESSION_ID]`: `--last`는 현재 디렉터리의 가장 최근 대화다. `codex exec resume [SESSION_ID]`도 `--last`·`--all`과 후속 prompt를 받는다. fork는 `codex fork`, `codex exec fork` | 요약 · developer-commands, non-interactive-mode |
| 목표 `/goal` | "Use `/goal` when you want Codex to keep working toward one durable objective instead of stopping after one normal turn." `features.goals` 또는 `codex features enable goals`로 켠다. 명령은 `/goal pause`·`resume`·`clear`다. 목표는 4,000자 이하다(요약). `codex exec`에서 쓸 수 있는지는 적혀 있지 않다 | 원문 · https://learn.chatgpt.com/use-cases/follow-goals |
| 반복 실행(loop) | Claude `/loop`에 해당하는 일정 기능은 확인되지 않았다 | 확인 불가 |
| `--agent` | 없다. 비슷한 것은 `--profile`/`-p`("Layer `$CODEX_HOME/profile-name.config.toml` on top of the base user config")와 `codex agents`, `/subagents`다 | 요약 · developer-commands |

## 2. `--agent`와 `--mode agent`의 실제 진입점 (AC-2)

### 결론

이름이 비슷한 세 가지 진입점이 있고, 서로 다른 층에 있다.

| 이름 | 층 | 실제 효과 | 근거 |
| --- | --- | --- | --- |
| `moai-cli -a [HOST]` / `moai-cli cc\|glm\|codex -a` (`--agent`) | MoAI 런처 플래그 | role `agent`를 붙여 host를 띄운다. 이름이 주어지지 않으면 `agent-1`을 붙이고, Claude 계열에는 permission skip을 붙인다. host에 `--agent`를 넘기지는 않는다. #2164 이후 폐기 예정 경고를 출력한다 | `src/main.rs:47-51`(최상위 `-a`, HOST 인자, 기본 `cc`), `src/main.rs:87-91`(하위 명령 `-a`, bool), `src/main.rs:616-662`(`launch`: role, 이름 기본값, `skip_permissions`, deprecated 경고 `:627-628`·`:652-655`), `src/main.rs:732-739`(`role_of`), `src/main.rs:742-748`(최상위 분기) |
| `/moai --mode agent`(Claude) · `$moai --mode agent`(Codex CLI) | skill 인자 | skill이 요청에서 `--mode`를 읽어 MODE-AGENT route를 고르고, `workflows/mode-agent.md`를 따른다. 실행 파일은 필요 없다 | `skills/moai/SKILL.md:11`, `skills/moai/references/routes.json:40-43`(`"MODE-AGENT"`, `requires_binary: false`), `skills/moai/workflows/mode-agent.md:3` |
| `claude --agent <name>` | Claude Code host 플래그 | 지정한 agent 정의를 세션에 적용한다("overrides the `agent` setting"). MoAI 런처는 이 플래그를 만들지 않는다. hook 입력에서는 `agent_type`으로 보인다 | 공식 cli-reference(원문) · hooks `agent_type` 행(원문). `src/session.rs`에서 `--agent`를 grep하면 0건이다 |

- 사용자가 말한 `--agent`가 런처 `-a|--agent`를 뜻한다면, 그 플래그는 host 세션을 여는 일만 하고 MoAI mode를 켜지 않는다. mode는 세션 안에서 `/moai --mode agent`로 켠다(경고 문구 `src/main.rs:628` "use `/moai --mode leader|agent` inside the session").
- `--mode`는 skill이 프롬프트 텍스트에서 읽는 인자다. host가 이 인자를 해석하는 것이 아니다. 따라서 이 인자를 강제할 host 기능은 없다(§3).
- 정본 seq44 L1150도 런처 `-l|--leader`·`-a|--agent`와 skill 진입 제안을 구분한다. 다만 정본은 그 진입을 `/moai --leader-mode`·`--agent-mode`로 적고, 현재 skill과 AGENTS.local.md는 `--mode leader|agent`를 쓴다. **표기 차이**로 기록해 둔다(누가 정본을 갱신할지는 리더 판단).
- 리더 mode는 Codex에서 거부된다(`skills/moai/workflows/mode-leader.md:15`, 결정 D-3).

## 3. 경계 표 — 지침만 / 이벤트로 관측 가능 / 실제 강제 가능 (AC-3)

"강제 가능"은 host가 공식 기능으로 행동을 막거나 바꿀 수 있다는 뜻이다. MoAI가 현재 그 기능을 쓴다는 뜻은 아니다. RC1 MoAI hook은 0개다.

| 항목 | 지침만 | 이벤트로 관측 가능 | 실제 강제 가능 |
| --- | --- | --- | --- |
| claim 뒤에만 첫 커밋 | 레인 지침(AGENTS.local.md, 카드 공통 규칙) | Claude PreToolUse(Bash `git commit`)와 Codex PreToolUse로 관측 가능 | PreToolUse 차단으로 가능(양 host). 판정할 claim 근거는 MoAI가 따로 읽어야 한다. MoAI hook 0이라 현재는 지침만 있다 |
| worktree 안에서만 쓰기 | 레인 지침 | Claude: hook `cwd`, CwdChanged, WorktreeCreate/Remove. Codex: hook `cwd` | Claude는 PreToolUse로 경로를 차단할 수 있다. 이 세션에서도 별도 guard가 cross-tree 명령을 거부한 것이 관측됐다. 그 guard는 MoAI RC 산출물이 아니다. Codex도 PreToolUse로 가능하지만, project hook은 신뢰 검토를 받아야 돈다 |
| worktree 진입·이탈 | 지침(EnterWorktree/ExitWorktree, `moai cc -w`) | Claude: WorktreeCreate(`--worktree`, `isolation`, background에서만, Bash `git worktree add`는 아님), WorktreeRemove, CwdChanged. Codex: worktree 전용 이벤트는 확인 불가 | Claude WorktreeCreate는 0이 아닌 exit로 생성을 막을 수 있다. 이탈을 강제하는 기능은 확인 불가 |
| 세션 식별(lane·window·provider_session) | 이름 붙이기는 지침(`--name`, Codex 첫 prompt) | 양 host 모두 hook `session_id`가 있다. Claude headless JSON에도 `session_id`가 있다. Codex exec에는 `thread_id`가 있다(관계 미정의) | 강제할 대상이 아니다. 정본 L1298–1309대로 실제 ID가 없으면 unverified로 둔다 |
| 턴 종료·완료 판정 | "idle·Stop은 완료가 아니다"(정본 L1347, cross-session 규칙) | Claude: Stop, SubagentStop, StopFailure, TeammateIdle. Codex: Stop, app-server `turn/completed` | Stop exit 2(Claude)나 `decision:"block"`(Codex)로 종료를 막고 계속 시킬 수 있다. 업무 완료 자체는 강제할 수 없고, 증거로만 판정한다 |
| 세션 종료 | 지침(ExitWorktree(keep) 뒤 `/clear`) | Claude SessionEnd(사유 5종, 1.5초 예산). Codex SessionEnd(`reason`은 항상 `other`, 1초 기본) | 종료를 막을 수는 없다(관측만 가능) |
| 재개 | 지침(resume 블록, 인계 기록) | 양 host 모두 SessionStart `source=resume`. Claude는 `--continue`·`--resume`, Codex는 `codex resume`·`codex exec resume` | SessionStart가 컨텍스트를 주입할 수 있다(Claude는 stdout plain text, 원문). 재개 자체를 강제하지는 않는다 |
| 반복 실행(다음 카드로 이동) | 레인 루프 지침(병합을 기다리지 않음) | Claude `/loop`·ScheduleWakeup(7일 만료, 50개 상한). Claude `/goal`은 매 턴 평가. Codex `/goal` | Claude Stop·TeammateIdle exit 2로 계속 시킬 수 있다. `/goal`은 평가 모델이 판단하므로 결정론적 강제가 아니다. 무한 실행 API로 보지 않는다(정본 L1355) |
| `--mode agent` 진입 | skill 지침(SKILL.md:11) | Claude UserPromptSubmit·UserPromptExpansion으로 프롬프트 텍스트를 관측할 수 있다. Codex UserPromptSubmit도 같다 | UserPromptSubmit에서 막거나 컨텍스트를 붙일 수는 있다. mode가 올바른지 판정하는 일은 MoAI 몫이다 |
| 권한·승인 | host 권한 체계 | Claude PermissionRequest·PermissionDenied(auto mode). Codex PermissionRequest | host sandbox·승인이 강제한다. MoAI 지침이 이를 넓히지 못한다 |

## 4. 이벤트 신뢰성 — 중복·역순·유실 (AC-1 범위)

| 성질 | Claude Code | Codex |
| --- | --- | --- |
| 동시 실행 | "When an event fires, its matching hooks run in parallel." prompt·agent hook은 12개씩 돈다(원문) | "Multiple matching command hooks for the same event are launched concurrently". background hook은 세션당 최대 8개다(원문) |
| 순서 | 서로 다른 hook 사이의 순서는 적혀 있지 않다. 근거 없음 | "background hooks can finish in a different order than they started"(원문). 동기 hook의 순서는 근거 없음 |
| 중복 | 같은 handler가 여러 설정 파일에 있으면 한 번만 돈다. plugin·skill 사본은 따로 돈다(원문). 이벤트 자체가 두 번 오는지는 근거 없음 | 근거 없음 |
| 유실·미발생 | SIGTERM을 받으면 SessionEnd만 돌고 종료한다(원문, headless). timeout에 걸린 hook은 취소되고 출력이 버려진다(원문). Stop은 사용자 interrupt에서 돌지 않는다(요약). PostToolUseFailure는 권한 거부·취소에서 돌지 않는다(요약). crash 때 SessionEnd가 도는지는 근거 없음 | 세션이 끝나면 끝나지 않은 background hook을 취소하고 전달되지 않은 출력을 버린다(요약). untrusted project의 project hook은 돌지 않는다(원문). SessionEnd·Interrupt는 subagent에서 돌지 않는다(원문) |
| timeout | command·http·mcp_tool 600초, prompt 30초, agent 60초. SessionEnd는 1.5초 공동 예산이고 최대 60초까지 늘릴 수 있다(원문) | 대부분 600초. SessionEnd·Interrupt는 1초 기본, 최대 3초(원문) |
| 지연 | `transcript_path`는 "written asynchronously and may lag"(원문) | transcript 형식은 hook의 안정 인터페이스가 아니다(원문) |

정본 L1331도 같은 결론에 이른다. 중복·역순·지연 이벤트를 허용하는 멱등 전이로 처리하고, 원격 상태를 다시 조회한다. 위 표는 그 요구의 공식 근거를 모은 것이다.

## 5. 남은 일

- AC-1·AC-3 표의 **요약** 표시 항목은 구현에 쓰기 전에 원문을 다시 확인해야 한다. Claude Stop 입력(`stop_hook_active`, continuation 상한)과 Codex resume·worktree 세부가 여기에 해당한다.
- **확인 불가**: Codex worktree 이탈 명령, Codex의 `/loop`형 일정 기능, Codex Stop의 발생 조건, `codex exec`에서의 `/goal`, 양 host의 이벤트 중복 발생 여부.
- 정본 표기 `--leader-mode`/`--agent-mode`와 현재 `--mode leader|agent`가 다르다(§2).
- HOST-01(정본 L1375): Claude TeammateIdle 분기와 Codex terminal·approval 이벤트의 실측은 운영자 몫이다. 이 카드에서는 not_run이다.

## Gaps (미검증)

- 실세션·hook을 실행하지 않았다. 모든 행은 문서 근거다.
- Claude hooks 페이지의 약 180k자 이후와 Codex 페이지들의 뒷부분은 요약으로만 읽었다.
- Claude agent-teams 페이지와 Codex app-server 페이지는 이번에 다시 읽지 않았다. 해당 행은 정본 seq44 인용(L1347–1357)에 의존한다.
- 공식 페이지에는 페이지 버전이나 수정일이 없다. 그래서 확인일로만 stale 여부를 판단한다.

## Residual-risk

- WebFetch가 돌려준 "원문"도 HTML에서 추출한 텍스트라, 글자 단위로 정확하다고 보장하지 못한다.
- 로컬 버전(2.1.319, 0.163.0)과 공식 문서가 설명하는 최신 동작은 다를 수 있다. 버전이 적힌 기능만 도입 버전을 옮겼다.
