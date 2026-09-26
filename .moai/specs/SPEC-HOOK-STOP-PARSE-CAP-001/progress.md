# Progress — SPEC-HOOK-STOP-PARSE-CAP-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- 작성: manager-spec (card t1272), 브랜치 `WT-stop-parse-cap`, 기준 트리 `e464fd5d0`
- 산출물: spec.md, plan.md, acceptance.md, progress.md (Tier M)
- SPEC ID 형식 검사: `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`. 사전 부재: `ls .moai/specs/SPEC-HOOK-STOP-PARSE-CAP-001` → `No such file or directory`
- 0.2.0 수리(plan-audit 1차 FAIL 0.79 → 리드 판정 2026-09-26 반영): REQ 15개, AC 16개(차단 14, 회귀 가드 2)
- 0.2.1 수리(plan-audit 2차 FAIL 0.73, `.moai/reports/t1272/plan-audit-iter2.md`): 차단 결함 N1~N4 만 반영했다. 설계와 N5~N8 은 건드리지 않았다. REQ·AC 개수 변화 없음
- 3차 plan-audit 예외: 다음 3차 감사는 Tier M 반복 상한(2회)을 넘는 **1회 한정 예외**다. 리드가 2026-09-26 승인했으며, 감사 범위는 N1~N4 의 수리가 맞는지로 한정한다
- 미해결 표식: 없음 — 만료 시간은 리드 판정으로 60분 확정(plan.md §B.1)
- 같은 카드에서 선행 SPEC-HOOK-STDIN-FAILCLOSED-001 을 0.4.3 으로 정정했다(F6, AC-SPC-014). 수리 라운드에서 그 plan.md M1 의 「미측정 독트린」 지시문에 정정 주석을 달았다(D6)
- 측정하지 않은 것: 9회 연속 파싱 실패 Stop 의 현재 동작(1회 실행 + 코드 판독으로 추론), 실제 프로세스 트리(설정·래퍼 파일 판독만), 훅 프로세스 환경의 `CLAUDE_CODE_SESSION_ID` 존재(AC-SPC-016 이 run-phase 에서 잰다), `/clear` 뒤 세션 id 변화(M5c), Windows 에서의 조상 탐색, in-process 팀원의 Stop 발생 경로

## §E.2 Run-phase Evidence

### E.2.0 LIVE 측정 상한 재선언 (실행 전, plan.md M5 그대로)

| 측정 | 실행 수 | 턴 상한 | 벽시계 상한 | 재실행 |
|---|---|---|---|---|
| M5a — 훅 환경의 세션 id (AC-SPC-016) | 1 | `--max-turns 2` | `timeout -k 10 120` | 없음 |
| M5b — 파손 stdin 루프 A(`CAP=200`)·B(unset) (AC-SPC-015 (i)(ii)) | 2 | 각 `--max-turns 30` | 각 `timeout -k 10 300` | 없음 |
| M5c — `/clear` 전후 세션 id (AC-SPC-015 (iii)) | 대화형 1 세션, 프롬프트 최대 3개 | 프롬프트당 1턴 | 세션 전체 10분 | 없음 |

선언을 넘기면 멈추고 재판정을 받는다. `timeout` 래퍼가 워크트리 가드에 거부되면 Bash 도구 타임아웃을 같은 벽시계 값으로 걸고 그 이탈을 기록한다.

### E.2.1 착수 전 점검 (plan.md §C) — 트리 `4e70ef069`

- C1: `git rev-parse HEAD` → `4e70ef069cd407529782bf7cb43d4590240d7791`; `git merge-base develop HEAD` → `e464fd5d0c0446051ef3e7fdc4b9bd1ce2fe2683`; `git diff --stat 6c0da65d5 HEAD -- internal/` → 출력 없음(코드 동일, 문서 커밋 3개뿐).
- C2: `answerStdinParseFailure(label, event, codex, stdinBytes, parseErr, writeDefault)` 서명·분기 순서(관측 → Codex 면제 → 거부) plan 시점과 같음.
- C3: Stop 항목은 `"command": "bash", "args": ["-c", "[ -f \"$0\" ] && exec bash \"$0\"; …", "${CLAUDE_PROJECT_DIR}/.claude/hooks/moai/handle-stop.sh"]` 셸 래퍼, `handle-stop.sh.tmpl` 은 `exec moai hook stop 2>>"$MOAI_HOOK_STDERR_LOG"` 로 끝남 — 셸 래퍼 모양 그대로(§F 「셸이 아닌 래퍼」 비현실).
- C4: `git check-ignore -v .moai/state/stop-parse-cap/x.json` → `.gitignore:398:.moai/state/	.moai/state/stop-parse-cap/x.json`.
- C5 (프로세스 지문 보강): 채택하지 않음. AC-SPC-006 대조는 원문 그대로 유지.
- C6 (이음매): `internal/session` 에 `ProcessView{Self, Info, Alive}` + `LiveProcessView()` + `(ProcessView).ResolveOwnerPID(stamp)` 를 두고 `ResolveOwnerPID`/`ancestorSessionPID`/`sessionPIDFromEnv` 가 이 경로를 공유(동작 불변 — 기존 `session_pid_test.go` 통과). `internal/cli` 는 `stopParseCapProcessView` 이음매로 합성 표를 주입하며, 해석을 상수로 대신하지 않는다.

### E.2.2 커밋

| SHA | 제목 |
|---|---|
| `aaa7f2b7d` | refactor(session): expose owner resolution over an injectable process view (t1272) — M1 전제 + `status: draft → in-progress` |
| `44a772f7f` | feat(hook): cap consecutive stdin-parse-failure Stops under Claude at N=8 (t1272) — M1~M3 |
| `34aa7d440` | docs(hook): document the Claude Stop parse-failure limit and correct the cap comment (t1272) — M4 |

### E.2.3 RED 관측 (구현 전, 이음매 스텁만 둔 트리)

명령 `go test -count=1 -run 'TestStopParseCap' ./internal/cli/` → exit 1, 원문 `.moai/state/verify/t1272/red-cli.txt`. 발췌:

```
--- FAIL: TestStopParseCap_AC001_DenyUpToN (0.03s)
    hook_stop_parse_cap_test.go:327: call 1: recorded count = -1, want 1
--- FAIL: TestStopParseCap_AC002_ReleaseAboveN (0.00s)
    hook_stop_parse_cap_test.go:347: call 9: stdout = "{\"decision\":\"block\",\"reason\":\"fail-closed: hook stdin could not be parsed as JSON (.moai/docs/hook-stdin-fail-closed.md)\"}\n", want "{}\n"
--- FAIL: TestStopParseCap_AC010_ClaudeReason (0.00s)
    hook_stop_parse_cap_test.go:925: (a) pre-tool []: reason = "fail-closed: hook stdin could not be parsed as JSON (.moai/docs/hook-stdin-fail-closed.md)", want "fail-closed: hook stdin could not be parsed as JSON. Do n…
```

AC001·002·003·004·006·007·008·010·011 이 RED, AC005(공허 초록, 원장 L4 대로)·AC009(보존 기준)·정적 검사(스텁 파일)는 초록. 세션 이음매: `go test -count=1 -run 'TestProcessView|TestLiveProcessView' ./internal/session/` → exit 1 (`ResolveOwnerPID = (0, false), want (100, true)`), 원문 `red-session.txt`.

### E.2.4 필수 RED 변이 (GREEN 트리에 변이 1개씩 넣고 되돌림)

| AC | 변이 | 명령 | 빨개진 테스트 (원문 발췌) | 원문 |
|---|---|---|---|---|
| 002 | `count <= N` → `count < N` (`>` → `>=`) | `go test -count=1 -run 'TestStopParseCap_AC00[12]' ./internal/cli/` exit 1 | `TestStopParseCap_AC001_DenyUpToN … call 8: stdout = "{}\n", want the Claude Stop deny` | `mutant-ac002-ge.txt` |
| 002 | N 판정 없음(구현 전 코드) | E.2.3 | `call 9: stdout = "{\"decision\":\"block\",…", want "{}\n"` | `red-cli.txt` |
| 004 | 모든 이벤트의 파싱 성공에서 셈 삭제 (`!harnessCodex && event == Stop` → `!harnessCodex`) | `go test -count=1 -run 'TestStopParseCap_AC004' ./internal/cli/` exit 1 | `(b) cumulative call 9: stdout = "{\"decision\":\"block\",…` | `mutant-ac004-reset-all.txt` |
| 005 | `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200` 이면 N=200 | `go test -count=1 -run 'TestStopParseCap_AC005|TestStopParseCap_StaticChecks' ./internal/cli/` exit 1 | `environment 2 call 9: stdout "{\"decision\":\"block\",…` + `hook_stop_parse_cap.go contains "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP" 1 times, want 0` | `mutant-ac005-env200.txt` |
| 007 (iii) | 주입 표의 첫 부모를 열쇠로 | `go test -count=1 -run 'TestStopParseCap_AC007' ./internal/cli/` exit 1 | `(iii)_interposed_shells_are_skipped … count for host 100 = -1, want 2` / `2 records, want 1` | `mutant-ac007-firstparent.txt` |
| 009 | 사유를 하네스 공통 새 문구로 | `go test -count=1 -run 'TestStopParseCap_AC009' ./internal/cli/` exit 1 | `(ii) PreToolUse` · `(ii) PermissionRequest` · `(ii) UserPromptSubmit` · `(iii) agent x-validation` 네 호출 모두 FAIL | `mutant-ac009-common-reason.txt` |
| 010 | 개정 전 사유(구현 전 코드) | E.2.3 | `(a) pre-tool []: reason = "fail-closed: hook stdin could not be parsed as JSON (.moai/docs/…)"` | `red-cli.txt` |

변이를 모두 되돌린 뒤 `go test -count=1 -run 'TestStopParseCap|TestStdinFailClosed|TestHookProtocol|TestMisc' ./internal/cli/` → `ok  github.com/modu-ai/moai-adk/internal/cli`, `green-cli-after-mutants.txt`.

### E.2.5 AC 판정 — 트리 `34aa7d440` (tree `21fbc658c`)

단위 판정 명령: `go test -count=1 -v -run 'TestStopParseCap' ./internal/cli/` → exit 0, `green-cli-v.txt` (SKIP 0건):

```
--- PASS: TestStopParseCap_AC001_DenyUpToN (0.03s)
--- PASS: TestStopParseCap_AC002_ReleaseAboveN (0.01s)
--- PASS: TestStopParseCap_AC003_ParsedStopResets (0.03s)
--- PASS: TestStopParseCap_AC004_ToolUseDoesNotReset (0.01s)
--- PASS: TestStopParseCap_AC005_IndependentOfHostCap (0.02s)
--- PASS: TestStopParseCap_StaticChecks (0.00s)
--- PASS: TestStopParseCap_AC006_Expiry (0.01s)
--- PASS: TestStopParseCap_AC007_KeyResolution (0.04s)
--- PASS: TestStopParseCap_AC008_UnusableRecord (0.02s)
--- PASS: TestStopParseCap_AC009_CodexUnchanged (0.00s)
--- PASS: TestStopParseCap_AC010_ClaudeReason (0.00s)
--- PASS: TestStopParseCap_AC011_NoPayloadLeak (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.770s
```

| AC | Status | 근거 |
|---|---|---|
| AC-SPC-001 | PASS | `TestStopParseCap_AC001_DenyUpToN` (파손 4형태 × 2바퀴, 8회 거부, 셈 k) |
| AC-SPC-002 | PASS | `TestStopParseCap_AC002_ReleaseAboveN` + 필수 RED 2종(E.2.4) |
| AC-SPC-003 | PASS | `TestStopParseCap_AC003_ParsedStopResets` (본문·변형 1·2·3; 변형 3 은 삭제 이음매 주입) |
| AC-SPC-004 | PASS | `TestStopParseCap_AC004_ToolUseDoesNotReset` + 필수 RED |
| AC-SPC-005 | PASS (회귀 가드) | `TestStopParseCap_AC005_IndependentOfHostCap` + 변이 RED; 정적 검사 `grep -c 'EnvClaudeCodeStopHookBlockCap' <두 파일>` → `0`/`0`, `grep -c 'CLAUDE_CODE_STOP_HOOK_BLOCK_CAP' <두 파일>` → `0`/`0` (두 파일 모두 존재) |
| AC-SPC-006 | PASS | `TestStopParseCap_AC006_Expiry` (본문·대조·(d)·(e)) |
| AC-SPC-007 | PASS | `TestStopParseCap_AC007_KeyResolution` (i)~(v) + (iii) 필수 RED; 정적 검사는 감사 O1 을 적용해 패턴을 `Getppid` 로 넓힘: `grep -c 'Getppid' internal/cli/hook_stop_parse_cap.go internal/cli/hook_stdin_failclosed.go internal/session/session_pid_view.go` → `0`/`0`/`0` |
| AC-SPC-008 | PASS | `TestStopParseCap_AC008_UnusableRecord` (i)(ii)(iii 링크·디렉터리) |
| AC-SPC-009 | PASS | `TestStopParseCap_AC009_CodexUnchanged` + 필수 RED |
| AC-SPC-010 | PASS | `TestStopParseCap_AC010_ClaudeReason` + 필수 RED(구현 전) |
| AC-SPC-011 | PASS | `TestStopParseCap_AC011_NoPayloadLeak` |
| AC-SPC-012 | PASS (기계 부분) · 리뷰어 판정 대기 | `grep -cE "N=8|8회"` → 템플릿 `1`, 로컬 `1`; `cmp` 두 파일 → 출력 없음(바이트 동일). (a)~(e) 는 새 절 「Stop under Claude Code: MoAI's own limit」 의 항목 N=8 / The release persists / Only two things reset it(60분) / What is counted together + Where the count lives / Independent of the host's limit 에 대응 — 리뷰어 표시 필요 |
| AC-SPC-013 | PASS (기계 부분) · 리뷰어 판정 대기 | `grep -c "has not been measured" internal/codexadapter/stop_cap.go` → `0` (exit 1). `@MX:WARN` 은 moai 자체 상한(N=8)과 200 주입 조건을 함께 언급 |
| AC-SPC-014 | PASS (plan-phase 충족, 변경 없음) | — |
| AC-SPC-015 | (i)(ii) PASS · (iii) 미측정 | E.2.7 |
| AC-SPC-016 | PASS | E.2.6 |

§D 선행 SPEC 테스트 갱신: `hook_stdin_failclosed_test.go` 의 공유 도우미 `expectedFailClosedReason()` 를 하네스별 리터럴 도우미 `expectedClaudeFailClosedReason()`·`expectedCodexFailClosedReason()` 둘로 갈랐고, `:815` 조립 검사도 하네스별로 갈랐으며, `hook_protocol_fix_test.go:127`·`misc_coverage_test.go:382` 는 Claude 리터럴과 비교한다.

**§D 목록 밖의 연쇄 수정 1건 (보고 대상).** 선행 SPEC 의 AST 가드 `TestStdinFailClosed_SingleDecisionListAST`(AC-HSF-003 (b3), `hook_stdin_failclosed_ast_test.go`)가 파싱 실패 경로의 환경 변수·파일 읽기를 전면 금지해, 이 SPEC 이 요구하는 열쇠 선택(REQ-SPC-014)과 셈 기록 읽기(REQ-SPC-001·007·008)와 정면으로 충돌했다(관측: `(b3) hook_stop_parse_cap.go:94:29: os.Getenv in the parse-failure path` 외 3건). 가드에 `hook_stop_parse_cap.go` 한 파일로 한정한 예외를 더했다 — `os.Getenv(config.EnvClaudeCodeSessionID)`·`os.Getenv(config.EnvMoaiSessionPID)`·`os.ReadFile`·`os.ReadDir` 만. 예외 자체의 검출 고정물(다른 파일에서는 4건 모두 위반, 같은 파일에서도 상한 변수·리터럴 이름·`os.Open`·`os.LookupEnv` 는 위반)을 함께 넣었다. (b1) 의 `== EventStop` 지적은 상한 적용 여부 판정을 `applyStopParseCap` 안으로 옮겨 해소했다(Stop 한 이벤트만 상한 대상이라는 판정을 상한 함수가 소유). acceptance.md §D 에 이 파일을 더하는 문서 갱신은 manager-spec 소관이다.

### E.2.6 M5a — 훅 환경의 세션 id (AC-SPC-016)

- 설치 방식 이탈: `moai init` 대신 템플릿 원본에서 두 산출물을 바이트 그대로 옮겼다 — `.claude/settings.json` 은 `settings.json.tmpl` 의 Stop 첫 항목(handle-stop.sh)만 담았고(추출 원문 `.moai/reports/t1272/m5a/settings.json`), `handle-stop.sh` 는 `handle-stop.sh.tmpl` 과 `cmp` 동일(템플릿 변수 `{{` 0건). 이유: `moai init` 은 사용자 범위 `$HOME/.claude/settings.json` 의 `defaultMode` 와 셸 rc 파일을 쓰는 경로(`ApplyAutonomyTierBundle`, `ptycaptest/homewatch.go` W3·W5)라 운영자 환경을 바꾼다. 바이너리는 트리 `34aa7d440` 을 `go build -o /private/tmp/t1272-m5a/moai-build ./cmd/moai` 로 빌드.
- 대리 스크립트: `PATH` 맨 앞 `/private/tmp/t1272-m5a/bin/moai`(원문 `.moai/reports/t1272/m5a/moai`), `hook stop` 호출마다 상태·값·`$$`·`$PPID` 한 줄 후 `exec` 빌드 바이너리.
- `timeout` 이탈: 워크트리 가드가 `timeout -k 10 120` 을 거부(`this command runs timeout with the text -k inside a construct too complex to verify`) → Bash 도구 타임아웃 120000ms 로 대신했다. `PATH=…:$PATH` 도 거부돼 PATH 를 리터럴로 적었다.
- 명령(한 번의 복합 호출): `cd /private/tmp/t1272-m5a/proj && unset CLAUDE_CODE_SESSION_ID MOAI_SESSION_PID MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS CLAUDE_CODE_STOP_HOOK_BLOCK_CAP && PATH=/private/tmp/t1272-m5a/bin:<리터럴> MOAI_HOOK_STDERR_LOG=/private/tmp/t1272-m5a/proj/.moai/logs/hook-stderr.log claude -p "reply with ok" --max-turns 2 --model haiku --output-format json > /private/tmp/t1272-m5a/m5a-result.json 2> …/m5a-stderr.txt` → exit 0, 1회(예산 1회 사용).
- 측정 로그(원문, 1줄 — 감사 O2 의 「≥ 1 줄」 충족):
  `2026-09-26T13:38:19Z state=set-nonempty value=e976ca45-d83e-43a2-8b17-2596be950bf3 pid=63114 ppid=62413 project_dir=/private/tmp/t1272-m5a/proj`
- 결과 JSON: `subtype: success`, `num_turns: 1`, `session_id: e976ca45-d83e-43a2-8b17-2596be950bf3`, `result: ok`.
- 판정: `set-nonempty` 값이 같은 실행의 `session_id` 와 일치 → **PASS**. 바깥 세션 값은 같은 호출 안에서 지웠으므로 관측값은 중첩 호스트가 스스로 찍은 것이다. 유효 stdin 이라 `.moai/state/stop-parse-cap/` 는 생기지 않았다(`find …/.moai/state` → `config-cache.json` 만).

### E.2.7 M5b — 파손 stdin 루프 (AC-SPC-015 (i)(ii))

- 같은 설치 방식, 대리 스크립트는 `hook stop` 에서 호스트 페이로드를 `cat >/dev/null` 로 버리고 `exec … < bad.txt`(`not json{`)(원문 `.moai/reports/t1272/m5b/default/moai`). `timeout` 은 같은 이유로 Bash 도구 타임아웃 300000ms 로 대신. 각 팔 1회, `--max-turns 30`, 재실행 없음.

| 팔 | Stop 호출(`measure.log` 줄 수) | 거부 / 해제 (`hook-stderr.log`) | 결과 | 모델의 도구 사용 |
|---|---|---|---|---|
| A: `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200` | 9 (모두 `state=set-nonempty value=cf9ecebc-…`, 결과 `session_id` 와 같음) | `answered fail-closed` 8 / `stop-parse cap released` 1 | `success`, `num_turns: 9`, 39.9s | 0 |
| B: 상한 변수 unset | 9 (모두 `value=04a8fd61-…`, 결과 `session_id` 와 같음) | 8 / 1 | `success`, `num_turns: 9`, 25.6s | 0 |

해제 줄 원문(두 팔 같음): `moai hook Stop: invalid stdin JSON (hook: invalid JSON input: invalid character 'o' in literal null (expecting 'u')) on Stop, harness claude; stop-parse cap released: consecutive parse failures 9 > N=8 (key: session); answered with no opinion`. 두 팔 모두 9 ∈ [9, 12] 이고 `error_max_turns` 가 아님 → (i)(ii) **PASS**. A 팔은 이전 측정(수정 전)에서 23회 Stop 뒤 `error_max_turns` 였던 조건이다. 관측 항목: 새 사유 문구를 받은 모델의 훅·설정 편집 시도 0회(두 팔 모두 도구 사용 0 — 세션 기록 `$CLAUDE_CONFIG_DIR/projects/-private-tmp-t1272-m5b-*/*.jsonl` 판독). 모델의 최종 응답: A `The blocker persists. Contact your administrator to fix the moai configuration.`, B `Session blocked. Unresolved infrastructure issue prevents operation. Human operator intervention required.` 증거 `.moai/reports/t1272/m5b/{cap200,default}/`.

### E.2.8 M5c — `/clear` 전후 세션 id (AC-SPC-015 (iii)) — 미측정

대화형 세션에서 `/clear` 를 입력해야 하는 측정이며 이 실행 주체(비대화형 서브에이전트)는 수행할 수 없다. 운영자 대화형 실행이 필요하다(plan.md M5 — M1~M4 완료를 막지 않음). 준비물: `/private/tmp/t1272-m5a/proj` 와 `/private/tmp/t1272-m5a/bin/moai` 를 그대로 쓰면 `measure.log` 에 `/clear` 전후 줄이 쌓인다.

### E.2.9 검증 묶음 — 트리 `34aa7d440`

| 항목 | 명령 | 결과 |
|---|---|---|
| 영향 패키지 | `go test -count=1 ./internal/codexadapter/... ./internal/config/...` | `ok` ×4 (`final-codexadapter-config.txt`) |
| | `go test -count=1 ./internal/session/...` | `ok  github.com/modu-ai/moai-adk/internal/session	6.530s` |
| | `go test -count=1 ./internal/spec/...` | `ok  github.com/modu-ai/moai-adk/internal/spec	106.289s` |
| | `go test -count=1 ./internal/template/` | `ok  github.com/modu-ai/moai-adk/internal/template	59.884s` |
| | `go test -count=1 ./internal/cli/...` (기본 10분) | 하위 패키지 18개 `ok`; `internal/cli` 는 `panic: test timed out after 10m0s` (실행 중 테스트 `TestAutoMergeHappyPath (1s)`, `--- FAIL` 0건) — 패키지 총 실행 시간이 기본 10분을 넘은 것이며 develop 쪽 SPEC-CLI-TEST-TIMEOUT-001(t1253)이 다루는 문제로, 이 브랜치 기준 트리 `e464fd5d0` 에는 그 수리가 없다 |
| | `go test -count=1 -timeout 25m ./internal/cli/` | `ok  github.com/modu-ai/moai-adk/internal/cli	997.777s`, exit 0, `--- FAIL` 0건 (`final-cli-root.txt`, `moai slot` 임대 `go-test-cli` 아래 실행) |
| lint | `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run ./internal/cli/... ./internal/codexadapter/... ./internal/session/... ./internal/config/...` | `0 issues.` |
| vet | `go vet ./internal/cli/... ./internal/codexadapter/... ./internal/session/... ./internal/config/...` | 출력 없음, exit 0 |
| Windows | `GOOS=windows GOARCH=amd64 go build ./...` | 출력 없음, exit 0 |
| build | `make build` | `catalog.yaml updated successfully`, `go build … -o bin/moai ./cmd/moai` exit 0 |
| 커버리지 | `go test -count=1 -run 'TestStopParseCap|TestStdinFailClosed' -coverprofile=… ./internal/cli/` 파일별 집계 | `hook_stop_parse_cap.go` 99/115 = 86.1%, `hook_stdin_failclosed.go` 32/36 = 88.9% |
| | `go test -count=1 -coverprofile=… ./internal/session/` 파일별 집계 | `session_pid_view.go` 33/35 = 94.3%, `session_pid.go` 21/21 = 100% |

검증 원문은 `.moai/state/verify/t1272/` 아래에 있다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 34aa7d4405d921c451154e72ef9995e5821b23a0   # M4 커밋; 이 progress 기록은 그 다음 커밋이다
run_status: audit-ready-with-gap
ac_pass_count: 15            # 001~014, 016 (012·013 은 기계 부분 PASS, 리뷰어 표시 대기)
ac_fail_count: 0
ac_gap: "AC-SPC-015 (iii) — M5c /clear 측정은 대화형 운영자 실행 필요 (비차단)"
preserve_list_post_run_count: "internal/hook 변경 0, Codex 출력 바이트 불변(AC-SPC-009)"
l44_pre_commit_fetch: not-run   # 레인은 push 하지 않는다 (리드 일괄)
l44_post_push_fetch: not-run
new_warnings_or_lints_introduced: 0   # golangci-lint v2.1.6: 0 issues
cross_platform_build:
  darwin: ok        # make build
  windows_amd64: ok # GOOS=windows GOARCH=amd64 go build ./...
total_run_phase_files: 17
m1_to_mN_commit_strategy: "3 commits (session seam / cap+reason+tests / docs+comment) + this evidence commit"
cascade_outside_acceptance_D: "hook_stdin_failclosed_ast_test.go AC-HSF-003 (b3) carve-out — acceptance.md §D 반영은 manager-spec 소관"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
