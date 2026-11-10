# statusline Rust 구현 1단계 — 입력 계약·Go 기능 목록·동등성 기준 (#2331)

- 작성: agent-4 [59cb6a], 2026-11-10. 상위 #2237(EP-I).
- 비교 대상: moai-adk-go `4fa2aa9db2`(`git archive`로 꺼낸 읽기 전용 사본). Go 저장소에는 아무것도 쓰지 않았다.
- 이 문서는 1단계 산출물이다. 2단계 구현 카드 초안은 7절에 있다.

## 1. 공식 입력 계약 (AC-1)

출처: <https://code.claude.com/docs/en/statusline.md> (Markdown 판, HTTP 200, `text/markdown`). 확인 시각 2026-11-09T23:45:30Z, 응답 sha256 `6da886d1d49d40461586165d5f246cd8a049123b194caa1ebc755d8b33038cff`. HTML 판(`/docs/en/statusline`, 23:45:23Z, sha256 `edd62269…`)도 같은 시각에 받았다. 표는 Markdown 판의 "Available data" 절과 "Full JSON schema"를 옮긴 것이다.

### 실행 방식

- Claude Code는 `statusLine.command`를 셸로 실행하고 세션 JSON을 stdin으로 넘긴다. stdout 각 줄이 한 행이 된다. ANSI 색과 OSC 8 링크를 쓸 수 있다.
- 갱신 시점: 세션 시작·재개, 새 assistant 메시지, `/compact` 종료, 권한 모드 변경, vim 모드 전환, `command` 변경, `refreshInterval` 경과, `rate_limits.*.resets_at` 도달, `prompt_cache.expires_at` 도달. 300ms debounce가 걸리고, 실행 중에 새 갱신이 오면 진행 중인 스크립트를 취소한다.
- 폭: 출력은 캡처되므로 `tput cols`를 쓸 수 없다. `COLUMNS`·`LINES` 환경 변수를 읽는다.
- 설정 키(`statusLine`): `type: "command"`, `command`, `padding`(기본 0), `refreshInterval`(초, 최소 1), `hideVimModeIndicator`.

### stdin 필드

| 필드 | 타입 | 있을 때 / 없을 때 |
| --- | --- | --- |
| `model.id`, `model.display_name` | string | 항상 |
| `cwd`, `workspace.current_dir` | string | 항상(같은 값, `workspace.current_dir` 권장) |
| `workspace.project_dir` | string | 항상 |
| `workspace.added_dirs` | string[] | 항상(없으면 빈 배열) |
| `workspace.git_worktree` | string | linked git worktree 안에서만 |
| `workspace.repo.{host,owner,name}` | string | `origin` remote가 있는 git 저장소에서만 |
| `cost.total_cost_usd` | number | 항상(목록가 기준 추정치, `/clear`에 0) |
| `cost.total_duration_ms`, `cost.total_api_duration_ms` | number | 항상 |
| `cost.total_lines_added`, `cost.total_lines_removed` | number | 항상 |
| `context_window.total_input_tokens`, `.total_output_tokens` | number | 항상(첫 응답 전 0) |
| `context_window.context_window_size` | number | 항상(200000 또는 1000000) |
| `context_window.used_percentage`, `.remaining_percentage` | number | **null 가능**(세션 초기) |
| `context_window.current_usage.{input_tokens,output_tokens,cache_creation_input_tokens,cache_read_input_tokens}` | object | **null 가능**(첫 API 호출 전, `/compact` 직후) |
| `exceeds_200k_tokens` | bool | 항상 |
| `fast_mode` | bool | 항상 |
| `effort.level` | `low`·`medium`·`high`·`xhigh`·`max` | 모델이 effort를 지원할 때만 |
| `thinking.enabled` | bool | 항상 |
| `rate_limits.five_hour.{used_percentage,resets_at}`, `rate_limits.seven_day.{…}` | number, epoch 초 | Pro·Max·Ultra 또는 gateway에서만, 첫 응답 뒤. 창마다 따로 없을 수 있고 `resets_at`이 지나면 빠진다 |
| `rate_limits.spend_limit.{used_percentage,resets_at}` | number | gateway spend limit일 때(v2.1.251+) |
| `rate_limits.spend_limit.{used_usd,limit_usd,period}` | number·number·`daily`/`weekly`/`monthly` | 따로 없을 수 있음(v2.1.284+) |
| `prompt_cache.*` | object | 첫 응답 뒤(v2.1.251+). 하위 필드 `warm`·`caching_observed`·`ttl`·`expires_at`(null 가능)·`requests`·`misses`·`expected_rebuilds`·`hit_ratio`(null 가능)·`cache_write_tokens`·`miss_recache_tokens`·`last_miss_at`(null 가능)·`last_miss_cause`(null 가능, v2.1.260+)·`miss_causes`·`recache_tokens_if_cold`(null 가능) |
| `session_id` | string | 항상 |
| `session_name` | string | 사용자 지정 이름이나 AI 제목이 있을 때만 |
| `prompt_id` | UUID string | 첫 사용자 입력 뒤 |
| `transcript_path` | string | 항상 |
| `version` | string | 항상 |
| `output_style.name` | string | 항상 |
| `vim.mode` | `NORMAL`·`INSERT`·`VISUAL`·`VISUAL LINE` | vim 모드일 때만 |
| `agent.name` | string | `--agent` 또는 agent 설정일 때만 |
| `pr.number`, `pr.url` | number, string | 현재 branch에 열린 PR/MR이 있을 때만 |
| `pr.review_state` | `approved`·`pending`·`changes_requested`·`draft` | 따로 없을 수 있음 |
| `pr.kind` | `mr` | GitLab MR일 때만(GitHub는 없음) |
| `worktree.{name,path,branch,original_cwd,original_branch}` | string | worktree 세션 중에만. `branch`·`original_branch`는 hook 기반 worktree에서 없음 |

## 2. Go 기능 목록 (AC-2)

방법: 꺼낸 사본의 `internal/statusline/`(테스트 제외 30개 파일)과 `internal/cli/statusline.go`를 읽기 전용 조사 에이전트로 판독했다. 판독한 위치 가운데 7곳(`isSelfInvocable` github.go:113, `maybeRefreshGitHubCounts` :140, `formatCost` metrics.go:236과 호출부 0건, `WriteModelCache` model_cache.go:47, render budget cli/statusline.go:19, `renderDefaultV3` renderer.go:161, `NewTheme` theme.go:38)은 grep으로 직접 대조해 모두 일치했다. 나머지 위치는 그 판독 결과다.

분류: **(a)** stdin만, **(b)** 로컬 파일·git, **(c)** 네트워크·자격 증명. 구성: L1 정보 줄, L2 막대 3개(폭 10 고정), L3 디렉터리·git·task·PR, L4 세션·backlog(내용이 있을 때만). 칸 구분은 `" | "`(renderer.go:51). `segments`에 없는 키는 켜진 것으로 본다(`isSegmentEnabled` renderer.go:128).

| 기능 / 표시 | 위치 | 입력 | 설정 키 | 캐시·제한 | 값이 없을 때 | 분류 |
| --- | --- | --- | --- | --- | --- | --- |
| 모델 `🤖 X` | metrics.go:24, renderer.go:388 | `model.display_name` > `model.id` > `model.name`(문자열도 허용, types.go:197). 대체: `MOAI_LAST_MODEL`, `$HOME/.moai/state/last-model.txt`. GLM은 `ANTHROPIC_DEFAULT_*_MODEL` | `model` | 렌더마다 `last-model.txt`에 씀 | 생략 | (a), 대체 경로 (b) |
| effort/thinking `🧠 level·t` | renderer.go:444 | `effort.level`, `thinking.enabled` | `effort_thinking` | — | 생략 | (a) |
| cache-hit `♻️ N%` | renderer.go:467·482 | `current_usage.cache_read_input_tokens`·`cache_creation_input_tokens` | `cache_hit` | — | `current_usage` null 또는 creation ≤ 0이면 생략 | (a) |
| Claude 버전 `🔅 vX` | renderer.go:408 | `version` | `claude_version` | — | 생략 | (a) |
| MoAI 버전 `🗿 vX [-> 🗿 vY]` | version.go:57·90 | 작업 디렉터리에서 위로 찾은 `.moai/config/sections/system.yaml`의 `moai.template_version`(없으면 `moai.version`)과 binary 버전 비교 | `moai_version` | 네트워크 없음(`UpdateChecker` update.go:29는 연결되지 않음) | binary 버전 | (b) |
| 세션 시간 `⏳ Xh Ym` | renderer.go:1097 | `cost.total_duration_ms` | `session_time` | — | 0이면 생략 | (a) |
| 출력 스타일 `💬 X` | renderer.go:434 | `output_style.name` | `output_style` | — | 생략 | (a) |
| 컨텍스트 막대 `CW: … N% (⚠️/clear)` | memory.go:194·287, renderer.go:504·846·870·892 | `used_percentage` > `current_usage` 합 > 옛 `used`/`total`. 창 크기: `context_window_size` > `total` > 200000. 덮어쓰기: `MOAI_STATUSLINE_CONTEXT_SIZE`, `llm.yaml`, 내장 GLM 표. 막대는 `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`(기본 85)로 환산 | `context` | — | `context_window` null이면 `n/a`, 사용량 필드가 없으면 0% | (a), 덮어쓰기 (b) |
| 5H 막대 `5H: … N% (4h 30m)` | renderer.go:532·1016 | `rate_limits.five_hour.*`. 없으면 `<primary>/.moai/state/context-usage/*.json`과 linked worktree들의 기록(30분 이내) | `usage_5h` | 기록 30분, 시계 오차 5분 | `n/a`. 기록 사용 시 `(captured Xm ago)` | (a), 대체 경로 (b) |
| 7D 막대 `7D: … N% (Jan 2)` | renderer.go:545·1053 | `rate_limits.seven_day.*`, 대체 경로는 5H와 같음 | `usage_7d` | 5H와 같음 | `n/a` | (a)/(b) |
| 디렉터리 `📁 name` | builder.go:435 | `workspace.project_dir` > `current_dir` > `cwd` > 작업 디렉터리 | `directory` | — | 작업 디렉터리 basename | (a) |
| branch·ahead/behind·dirty `🅱️ br ↑A ↓B +N` | git.go:26, renderer.go:730 | `git rev-parse`, `git --no-optional-locks status --porcelain --branch`, `symbolic-ref`(제한 5초). workflow 표시는 `git-strategy.yaml` | `git_branch` | — | 저장소 없으면 생략, 0이면 화살표 생략 | (b) |
| 저장소 `📡 owner/name` | renderer.go:769 | `workspace.repo.owner`·`name` | `git_branch`에 묶임 | — | git이 없으면 숨김 | (a), git이 있을 때만 보임 |
| forge 수 `, 이슈/PR` | github.go:79·206, renderer.go:807 | 렌더는 `<anchor>/.moai/state/github/counts.json`만 읽음. 갱신은 분리된 자식이 `gh`/`glab` 호출 | `github`, `statusline.forge` | TTL 10분, 갱신 20초, rate 하한 100 | 캐시 없으면 `, -/-` | 렌더는 (b), 갱신은 **(c)** |
| git 상태 `💾 +S MM ?U` | renderer.go:1136 | branch와 같은 git 데이터 | `git_status` | — | git 없으면 생략 | (b) |
| task `📋 [cmd SPEC-x-stage]` | task.go:126·104, renderer.go:614 | `<MOAI_HOME>/state/last-session-state.json`의 `active_task` | `task` | 프로세스 안 1초 | 생략 | (b) |
| PR `💌 PR #N (⌥state)` | renderer.go:655·921 | `pr.number`, `pr.review_state` | `pr` | — | null이나 0이면 생략 | (a) |
| 세션 `🏷️ name`, `👤 agent` | builder.go:251, renderer.go:214 | `session_name`, `agent.name` | `session` | — | 생략 | (a) |
| backlog `🔄 TODO: picked/queued` | backlog.go:47 | `<MoaiHome>/db/<key>/todo/backlog.db`(읽기 전용) 또는 `backlog.json` | `backlog` **그리고** `workflow.todo.enabled` | — | 생략 | (b) |
| landed `⚑N` | landed.go:115·210 | `<anchor>/.moai/state/landed/counts.json`. 갱신은 분리된 자식이 로컬 git만 사용 | backlog와 같음 | TTL 10분, 20초 | 모르면 생략 | (b) |
| goal 무장 시 `/clear` 표시 숨김 | goal_armed.go:32 | `<anchor>/.moai/state/goal/<session_id>.json`의 `status == "armed"` | — | — | 표시 유지 | (b) |
| subagent 행(JSONL) | renderer.go:278·325 | `tasks[]`의 `id`·`name`·`agentType`·`status`·`tokenCount` | — | — | `id`가 빈 행 생략 | (a) |
| 세션 기록 쓰기(표시 없음) | context_usage.go:210·262 | 메모리 데이터, `session_id`, 모델, effort, `rate_limits` | — | 바뀔 때 또는 5분 heartbeat | anchor·세션 id가 없으면 생략 | (b), 쓰기 |

수집만 하고 표시하지 않는 것: 비용(`formatCost` metrics.go:236, 호출부 0건), worktree(`workspace.git_worktree`, `[WT]` 접두는 폐기됨), `exceeds_200k_tokens`, 토큰·줄 수 합계, `transcript_path`, `pr.url`, `workspace.repo.host`, `worktree.*`(`original_cwd`만 anchor에 쓰임). Go가 읽지 않는 공식 필드: `fast_mode`, `prompt_cache`, `rate_limits.spend_limit`, `prompt_id`, `vim`, `pr.kind`, `workspace.added_dirs`.

테마·색: 테마는 `catppuccin-latte`와 `catppuccin-mocha` 두 가지(그 밖은 mocha, theme.go:38). `NO_COLOR` 또는 `MOAI_NO_COLOR`가 비어 있지 않으면 색 없음(cli/statusline.go:130). `COLORTERM`이 `truecolor`/`24bit`이면 TrueColor, 아니면 ANSI256(color.go:24). 채운 칸마다 4색 gradient(gradient.go:69). 70% 초과면 🪫, 아니면 🔋. 폭은 10칸 고정이고 `COLUMNS`를 읽지 않는다.

설정 파일: `.moai/config/sections/statusline.yaml`의 `statusline.theme`·`statusline.segments`(map). `preset:`·`mode:`는 무시한다. `statusline.forge`는 `.moai/statusline.local.yaml`이 우선한다. `workflow.todo.enabled`, `git-strategy.yaml`, `llm.yaml`, `system.yaml`도 읽는다. 렌더 예산 800ms(cli/statusline.go:19).

## 3. 네트워크·자격 증명 (c) — 이 카드에서 실행하지 않음

| 기능 | 무엇을 부르나 | not_run 사유 | 2단계 대체안 |
| --- | --- | --- | --- |
| GitHub/GitLab 이슈·PR 수 갱신 | 분리된 자식 `<self> statusline --refresh-github`가 `git remote get-url`, `gh api rate_limit`, `gh repo view`/`gh issue|pr list`, `glab issue|mr list` 호출(github.go:174·293·326·352, forge.go:245) | 네트워크와 `gh` 자격 증명이 필요하고, 카드가 금지함 | (1) 공식 stdin `pr.*`·`workspace.repo`만으로 PR 표시(이미 (a)), (2) 이슈·PR 수는 표시하지 않거나 캐시 파일만 읽음(갱신 경로 없음), (3) 갱신이 필요하면 별도 카드로 사용자 승인 |

카드 본문은 (c) 예로 "OAuth 사용량 API"를 들었지만, 기준 커밋의 `internal/statusline`에는 HTTP·OAuth·키체인 호출이 없다. 5H/7D는 공식 stdin `rate_limits`(또는 로컬 기록)에서 온다. 업데이트 확인기(update.go:29)도 연결되어 있지 않다.

landed 갱신은 분리된 자식을 띄우지만 로컬 git만 쓰므로 (b)다. 렌더 중 자식 프로세스는 2단계에서 정했다(2026-11-10 11:50Z 사용자 결정, #2350 댓글 6766596464): 렌더는 Go와 같게 git만 제한 시간(800ms) 안에서 실행하고, 넘기면 git 칸을 뺀다. 분리된 갱신 자식은 띄우지 않는다.

## 4. 동등성 기준 (AC-3)

- 위치: `tests/fixtures/statusline/`. `cases.json`(15건: (a) 8, (b) 7), `stdin/*.json`, `golden/<case>.{plain,ansi}.txt`(30개).
- 실행 도구: `tools/statusline_golden.py --bin <binary> --write|--check`. 2단계에서 Rust binary에 `--check`로 같은 기준을 쓴다.
- Go binary: `4fa2aa9db2` 사본을 `go1.27.3 darwin/arm64`, `-trimpath`, `CGO_ENABLED=0`로 빌드. HOME·GOCACHE·GOMODCACHE·GOPATH는 모두 임시 디렉터리. sha256 `495864b3cc39d72e08b45934e0b79df65c93f44ccb459c93288d2fabacb97bf6`. ldflags를 주지 않아 binary 버전 표시는 `v3.2.1`이다.
- 격리(사례·색 형태마다):
  - 새 임시 HOME·`MOAI_HOME`·프로젝트(`demo-project`).
  - 환경 변수는 처음부터 새로 만든다(`PATH=/usr/bin:/bin`, `TZ=UTC`, `LANG=C.UTF-8`).
  - binary를 `moai`가 아닌 이름으로 복사해 갱신 자식이 뜨지 않게 한다(`isSelfInvocable`).
  - macOS `sandbox-exec`로 네트워크 전부와 실제 HOME 쓰기를 막는다. 같은 profile로 `curl https://example.com` → exit 6, HOME 아래 `touch` → `Operation not permitted`를 실측했다.
- 색 두 형태: `plain`(`NO_COLOR=1`, ESC 0개)과 `ansi`(`COLORTERM=truecolor`, `38;2;` truecolor 코드 포함. 예: a02는 ESC 33개).
- 결정성: `--write` 뒤 새 임시 디렉터리에서 `--check` → 30/30 일치, exit 0. golden 하나를 바꾸면 `--check`가 `DIFF b06-goal-armed plain`, exit 1(복원 후 cmp 일치).
- 시간에 따라 바뀌는 값은 넣지 않았다: 5H `resets_at`은 과거라 `rolling`, 7D는 절대 날짜(UTC)다.

| 사례 | 분류 | 확인하는 것 |
| --- | --- | --- |
| a01-minimal | a | 모델, `context_window`가 없을 때 n/a 막대, 디렉터리 |
| a02-doc-full | a | 공식 전체 예시: effort·thinking, cache-hit, 버전, 시간, 출력 스타일, 막대 환산(8% → 9%), 과거 resets_at, PR pending, 세션·agent |
| a03-context-null | a | `context_window: null` → `n/a` |
| a04-context-high | a | 높은 사용량 → `(⚠️/clear)` |
| a05-pr-approved-effort | a | PR approved, effort만, 1M 창, 시간 단위 |
| a06-invalid-json | a | 잘못된 JSON에서도 실패하지 않음 |
| a07-subagent-tasks | a | `tasks[]` → JSONL 행, 빈 `id` 행 생략 |
| a08-model-string-pr-changes | a | 문자열 모델, `current_usage`로 계산, PR changes_requested |
| b01-git-clean | b | branch, git이 있을 때 저장소 표시, 깨끗한 상태 |
| b02-git-dirty-ahead | b | `↑1`, 수정·미추적 수 |
| b03-config-latte-subset | b | `statusline.yaml` 테마 latte, `segments`로 끄기 |
| b04-template-version | b | `template_version`과 binary 버전 비교 화살표 |
| b05-task-state | b | `MOAI_HOME` task 상태 |
| b06-goal-armed | b | 무장된 goal이 `/clear` 표시를 숨김 |
| b07-goal-not-armed | b | 무장되지 않은 goal은 표시 유지(b06 대조) |

(c) 사례는 만들지 않았다(3절). backlog·landed·5H/7D 로컬 기록 대체 경로·llm.yaml·GLM 환경 변수 덮어쓰기 사례도 이번에는 만들지 않았다(8절 남은 것).

## 5. 렌더가 쓰는 파일 (sandbox 대상)

| 경로 | 언제 |
| --- | --- |
| `$HOME/.moai/state/last-model.txt` | 모델이 정해지는 모든 렌더 |
| `<anchor>/.moai/state/context-usage/<session_id>.json` | anchor와 세션 id가 있을 때, 바뀌었거나 5분 heartbeat |
| `<anchor>/.moai/state/{github,landed}/counts.json` | 분리된 자식만 |

stdin만 넣어도 위 두 곳에 쓴다. 그래서 HOME과 `MOAI_HOME`을 함께 격리해야 한다.

## 6. 공식 계약과 Go의 차이 (2단계 판단 거리)

- Go는 공식 필드 `fast_mode`, `prompt_cache`, `spend_limit`, `vim.mode`, `pr.kind`, `workspace.added_dirs`를 읽지 않는다.
- Go는 `COLUMNS`를 읽지 않고 폭 10칸을 고정한다. 공식 문서는 폭에 `COLUMNS`를 쓰라고 한다.
- Go는 공식 문서에 없는 `tasks[]`(subagent 행), `hook_event_name`, `cost.total_usd`, `context_window.used/total`을 읽는다.

## 7. 2단계 구현 카드 초안 (AC-4)

제목: `[rc.2][I 후속] statusline Rust 구현 2단계 — (a)(b) 기능 동등 구현 + golden --check`

- 범위:
  1. `src/statusline/` 모듈: `input.rs`(공식 stdin 계약 + Go가 추가로 읽는 필드, 모든 필드 선택적, 잘못된 JSON은 빈 입력으로 강등), `segments.rs`(2절 (a)(b) 기능을 하나씩), `render.rs`(4줄 배치, `" | "`, 10칸 막대, gradient, 테마 2종, `NO_COLOR`·`COLORTERM`), `state.rs`(anchor 결정과 기록 쓰기).
  2. CLI: `moai-cli statusline`. **JSON envelope를 쓰지 않는 예외 op**다. Claude Code가 stdout을 그대로 보여 주기 때문이다. OPERATIONS·`interfaces.md`·`operations_doc` 시험에 예외로 명시한다.
  3. 설정 호환: `.moai/config/sections/statusline.yaml`의 `theme`·`segments`를 그대로 읽는다(Go 사용자 설정 유지). RC1 쪽 위치(`.moai-cli/`)로 옮길지는 사용자 결정 거리다(HOLD).
  4. 동등성: `python3 -I tools/statusline_golden.py --bin target/debug/moai-cli --argv statusline --check`가 30/30 일치해야 한다. 버전 표시(`🗿 v…`)는 binary마다 다르므로, 이 줄의 비교 방식(버전 문자열을 고정할지, 정규화할지)을 2단계 첫 단계에서 정한다.
  5. 렌더 중 실행하는 자식 프로세스는 git뿐이며 제한 시간(800ms) 안에서만 실행한다(2026-11-10 11:50Z 사용자 결정, #2350 댓글 6766596464). 분리된 갱신 자식(`gh`/`glab`, `--refresh-*`)은 띄우지 않는다. forge 수는 캐시 파일만 읽고(갱신 없음), landed는 표시하지 않는다.
- 제외: (c) 갱신 경로 전체, `gh`/`glab` 호출, 사용자 `statusLine` 설정 변경, Orca wrapper·Go binary 변경.
- 경로 전환 계획(이 카드와 2단계 모두 실행하지 않음):
  1. 2단계 PR 병합 뒤 golden `--check` 통과를 확인한다.
  2. 사용자가 승인하면 #2239의 `claude-settings-plan`/`claude-settings-apply`로 프로젝트 `.claude/settings.json`의 `statusLine.command`만 바꾸는 plan을 만든다. 실제 HOME 설정은 바꾸지 않는다.
  3. 되돌리기는 같은 op의 receipt로 한다.
- 수용 기준 초안:
  - AC-1: (a)(b) 사례 15건의 plain·ansi golden 30개 `--check` 일치.
  - AC-2: 렌더 중 네트워크 0, git 외 자식 프로세스 0, git은 800ms 안에서만(sandbox 실측)(2026-11-10 11:50Z 사용자 결정, #2350 댓글 6766596464).
  - AC-3: 800ms 안에 출력(10회 실측 최댓값 기록).
  - AC-4: check set과 Self-Audit.

## 8. 남은 것·한계

- 기능 표의 파일:줄은 조사 에이전트의 판독 결과다. 7곳은 grep으로 직접 대조해 일치했고 나머지는 대조하지 않았다. 판독자가 "미확인"으로 표시한 4가지(`gitManager.Status`가 렌더 context를 쓰는지, `ResolveGitDirs`가 800ms 안에 묶이는지, `DetectRoleMarker`가 읽는 환경 변수 이름, SQLite `mode=ro`가 `-wal`/`-shm`을 만드는지)는 그대로 미확인이다.
- golden이 없는 기능: backlog, landed, 5H/7D 로컬 기록 대체 경로, `llm.yaml`·GLM 환경 변수 덮어쓰기, `MOAI_STATUSLINE_CONTEXT_SIZE`, `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` 값 변경, `statusline.forge` 덮어쓰기, workflow 표시(`git-strategy.yaml`).
- golden은 macOS(darwin/arm64)에서만 만들었다. Linux에서는 sandbox 없이 돌아가므로 같은 출력인지 측정하지 않았다.
- 실제 Claude Code에서 이 binary가 status line으로 그려지는지는 확인하지 않았다(host 실행 금지).
