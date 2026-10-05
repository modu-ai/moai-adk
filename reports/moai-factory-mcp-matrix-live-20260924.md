# MoAI Factory 호스트 조합 실세션 검증

2026-09-24 · macOS 실측 · MoAI MCP 메시지 경로 · 판정: **통신 4/4 성공, 자동 카드 처리·병합 미검증**

## Claim

macOS의 독립된 실제 호스트 세션에서 `moai codex -f`와 `moai cc -f`를 lead/worker로 조합한 네 경우 모두, MoAI MCP를 통해 `dispatch_notice`를 보내고 worker가 접수한 다음 `status_report`를 돌려보내 lead가 접수했다. 각 worker는 **배차를 읽은 뒤에** 지정 파일을 썼다. Claude Code 세션은 `moai-adk` 프로필의 Anthropic 직접 인증 계정과 Claude Code 2.1.281을 사용했다.

이 결과는 **명시적 MCP 호출과 폴링에 대한 성공**이다. `moai todo` 카드 발행과 `picked` 상태는 실측했으나 Factory의 `assignments`는 `[]`였다. 자동 카드 배차, plan/run/sync, 카드 상태 전이, Git worktree, PR, 병합 요청, 원격 병합, 기존 세션 자동 깨우기는 이번 운영 시험에서 실행하지 않았다. 따라서 네 조합에 대해 Claude Code Factory 전체와 동등하다는 판정은 아직 내릴 수 없다.

이전 [초기 실세션 보고서](moai-claude-codex-factory-live-20260924.html)의 `launch_pending` 관찰은 뒤이은 성공 시험으로 갱신됐다. 초기 보고서의 미완료 판정을 현재 최종 결과로 읽으면 안 된다.

## Baseline-attribution

이 보고서의 운영 관찰은 `/private/tmp/moai-factory-cross-test-20260924`에서 `MOAI_HOME=/private/tmp/moai-factory-cross-test-20260924/moai-home`을 설정하고 실행한 결과다. 시험 파일, 카드, 메시지 DB는 이 격리 프로젝트에 있다. 공유 저장소의 기존 변경은 건드리지 않았다.

```text
$ moai version
moai-adk v3.2.0-rc.13
[v3.2.0-rc.13] [moai_cp/20260910_130400-2916-g60017eb83] [built 2026-09-23T23:56:01Z]
$ codex --version
codex-cli 0.156.1
$ claude --version
2.1.281 (Claude Code)
$ uname -s
Darwin
```

Go 라이브 픽스처 네 개는 별도 clean `develop` worktree에서 실행했으며 실행 당시 HEAD는 `5cdfa51fb`였다. 조사 중 다른 세션이 이 worktree의 HEAD를 `6ed819fcf`로 전진시켰다. 설치 MoAI 바이너리의 빌드 커밋 `60017eb83`은 현재 `develop`의 조상이다(`git merge-base --is-ancestor 60017eb83 HEAD` → 종료 코드 `0`). 설치본과 현재 checkout을 같은 소스 버전으로 취급하지 않는다. 운영 시험은 **설치본**, Go 라이브 픽스처는 **당시 checkout**에 대한 결과다.

## Evidence — 실제 Factory 런처와 Claude 직접 인증 계정

운영 명령의 공통 형태는 다음과 같다. Codex는 `moai codex -f [worker] -- exec ...`, Claude는 `moai cc -p moai-adk -f [worker] --print ... --mcp-config ... --strict-mcp-config`로 시작했다. 각 역할은 별도의 CLI 프로세스와 세션 UUID를 가졌다. Codex는 내장 `moai` MCP를, Claude는 명시적 설정 파일 `/private/tmp/moai-factory-cross-test-20260924/claude-mcp-explicit.json`의 `moai mcp-server`를 사용했다. 파일은 각 배차 본문에 지정한 정확한 문자열을 저장했다.

다음은 이 실행에서 `moai todo list --json`과 해당 run의 `broker.db`를 읽은 결과다. DB는 조회만 했다. 메시지 상태는 lead/worker 종료 뒤의 최종 상태다.

| lead → worker | 카드 / run | 배차 메시지 ID | 완료 보고 ID | 결과 파일 내용 | 최종 상태 |
|---|---|---|---|---|---|
| Codex → Codex | `t2` / `tlutvh` | `3099d9baca726ff4217ff9fbb5b063ba` | `e4bd656723ae34f68c67737af2c745c1` | `CODEX_CODEX_DONE_0924` | 둘 다 `acknowledged/accepted` |
| Codex → Claude | `t3` / `tlutzb` | `5276823ba4cd0898a5c576614d0d5e7d` | `7e578eda41d3b1a9ff41d3a1e9111ca0` | `CODEX_CLAUDE_DONE_0924` | 둘 다 `acknowledged/accepted` |
| Claude → Codex | `t6` / `tluudw` | `202c83ecff5c114a4e10d702dbfb3529` | `b8961904a11ae8eb5ac9596354e40d26` | `CLAUDE_CODEX_DONE_0924` | 둘 다 `acknowledged/accepted` |
| Claude → Claude | `t8` / `tluul7` | `e90ca6df0c8bc44a6b683aac9f215d92` | `0c21acbf2dc29c6968b50566680fa278` | `CLAUDE_CLAUDE_DONE_0924` | 둘 다 `acknowledged/accepted` |

재현 가능한 조회 명령과 실제 출력 중 Claude → Codex 한 건:

```text
$ python3 - <<'PY'
import sqlite3
p='/private/tmp/moai-factory-cross-test-20260924/moai-home/db/moai-factory-cross-test-20260924-805fe8bf/factory/messages/tluudw/broker.db'
c=sqlite3.connect(p)
print(c.execute('select slot,backend,session_uuid from peers').fetchall())
print(c.execute('select id,kind,state,disposition,sender_slot from messages order by created_at').fetchall())
PY
[('lead', 'claude', 'b12f3630-af13-4a06-abf2-d105d8f62ba7'), ('worker-1', 'codex', '01a0d207-d4ab-7b70-b544-464be771c769')]
[('202c83ecff5c114a4e10d702dbfb3529', 'dispatch_notice', 'acknowledged', 'accepted', 'lead'), ('b8961904a11ae8eb5ac9596354e40d26', 'status_report', 'acknowledged', 'accepted', 'worker-1')]
$ cat /private/tmp/moai-factory-cross-test-20260924/claude-codex-receipt.txt
CLAUDE_CODEX_DONE_0924
```

나머지 세 run도 같은 조회에서 각각 lead/worker UUID 두 개와 배차·보고 두 행이 나왔다. 원본 호스트 출력은 `/private/tmp/moai-factory-matrix-20260924/operational-<조합>-{lead,worker}.log`에 있다. Claude → Codex는 파일명에 `-final`, Claude → Claude의 최종 성공 시험은 `-retry`가 붙는다. 각 성공 프로세스 종료 코드는 `0`이었다. 시험 run은 종료 뒤 `moai factory runs --retire <run-id>`로 정리했다.

`moai todo list --json`에서는 `t2`, `t3`, `t6`, `t8`이 모두 `state:"picked"`로 나왔다. 같은 JSON의 `runtime.assignments` 출력은 **`[]`**였다. 카드 번호를 MCP 본문과 `task_ref`에 넣어 수동 배차한 것이며, Todo ↔ Factory의 정식 배차 연결을 통과한 것은 아니다.

## Evidence — 라이브 카드 흐름 픽스처

별도의 네 Go 테스트도 각각 실제 Codex/Claude 모델 턴과 MCP 호출을 수행해 통과했다. 실행 형태: `MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=cardflow-<조합> MOAI_T1100_EVIDENCE_DIR=/private/tmp/moai-factory-matrix-20260924/evidence go test ./internal/cli -run '^TestFactoryLiveCardFlow<조합>$' -count=1 -v`.

```text
--- PASS: TestFactoryLiveCardFlowCodexCodex (193.75s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  194.792s
--- PASS: TestFactoryLiveCardFlowCodexClaude (196.04s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  196.944s
--- PASS: TestFactoryLiveCardFlowClaudeCodex (170.34s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  171.367s
--- PASS: TestFactoryLiveCardFlowClaudeClaude (110.46s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  111.299s
```

원본 로그는 `/private/tmp/moai-factory-matrix-20260924/{codex-codex,codex-claude,claude-codex,claude-claude}.log`, 구조화 근거는 같은 디렉터리의 `evidence/ac018-evidence-*.json`이다. 각 JSON에서 `applied_results:1`, `late_result_outcome:"stale"`, `current_result_outcome:"accepted"`, `final_state:"integrated"`를 확인했다. 이 테스트의 Claude 역할은 `moai glm`으로 구동된 Claude Code 호스트다. 또한 [`factory_live_test.go`](../internal/cli/factory_live_test.go)의 픽스처가 peer를 직접 등록하고 카드 상태를 구동한다. **픽스처의 `integrated`는 실제 Todo 카드의 PR 병합 증거가 아니다.**

## 관찰한 결함과 운영 조건

### High — Claude worker의 `--` 구분자 뒤 역할 인식 실패

`moai cc -p moai-adk -f worker -- --print ...`로 시작한 초기 시험은 Codex lead의 run `tlutle`에 합류하지 않고 새 Claude lead run `tlutlz`를 만들었다. DB 조회 원문:

```text
tlutle [('lead', 'codex', 'lead')] []
tlutlz [('lead', 'claude', 'lead')] []
```

구분자를 생략한 뒤의 `moai cc -p moai-adk -f worker --print ...`는 동일한 worker 역할로 정상 합류해 Codex → Claude 시험을 통과했다. 조사한 소스에서 [`factory.go`](../internal/cli/factory.go)는 worker `--name`을 `entry.Rest` 끝에 추가하고, [`kanban.go`](../internal/cli/kanban.go)의 `parseNamedLabel`은 `--`에서 탐색을 멈춘다. 이것이 관찰된 오분류를 설명한다. `--` 앞에서 역할을 확정하거나 추가된 플래그를 구분자 앞에 삽입하는 수정을 우선 적용하고, Claude·Codex 런처 양쪽에 실제 argv 회귀 시험이 필요하다.

### High — MCP 스키마와 저장소의 필수값 불일치

`factory_msg_send`의 MCP 도구 스키마는 `task_ref`, `correlation_id`를 선택 항목으로 선언하지만 [`store.go`](../internal/factorymsg/store.go)는 두 값 모두 `safeID` 정규식으로 검사한다. 초기 run `tlutpr`에서 `correlation_id` 없이 호출한 결과는 다음과 같았다.

```text
"tool":"factory_msg_send","arguments":{..."task_ref":"t1",...},"result":{"content":[{"type":"text","text":"factory_msg_send: invalid correlation id"}],..."status":"failed"}
```

이 경우 배차는 생성되지 않았는데 Claude worker가 파일과 보고를 만든 시도까지 있었다. **파일 존재와 worker 보고만으로 배차 성공을 판정하면 거짓 양성**이 된다. 스키마를 실제 필수 계약에 맞추고, lead가 발급한 배차 ID와 worker가 접수한 ID, 보고의 카드·상관관계를 함께 검증해야 한다.

### High — Claude MCP 연결과 세션 등록 순서

자동 `.mcp.json`만 둔 Claude lead 시험 `tluu43`에서는 모델이 도구 이름을 찾았지만 실제 MCP 호출 대신 Bash 명령으로 실행했고, 메시지는 `[]`였다. 명시적 `--mcp-config ... --strict-mcp-config`를 적용한 뒤 실제 MCP 도구 호출에 성공했다. 이 격리 프로젝트의 Claude 실행에는 프로젝트 신뢰 경고도 출력됐다. 자동 연결의 실패 원인이 설정 발견, 신뢰 상태, 모델 행동 중 어느 것인지는 분리 검증이 필요하다. 제품 진입점은 세션 시작 직후 필수 MCP 도구의 실사용 가능 여부를 확인해 실패를 명시해야 한다.

worker 등록 전에 보내면 `factory_msg_send: sql: no rows in result set`가 나왔다(`tluua2`). lead가 전송을 재시도하거나 worker 준비 상태를 기다려야 한다. `tluui3`에서는 Claude lead가 대기를 백그라운드로 시작한 뒤 턴을 끝내 배차 없이 종료했다. 모델 프롬프트의 대기 지시만으로 준비 상태를 보장할 수 없으므로 호스트 밖의 결정적 준비 확인을 권장한다.

### Medium — 비동기 전달 보장의 표현

이 시험의 worker는 `factory_msg_list`를 **명시적으로 호출**해 배차를 가져왔다. 이미 끝난 턴을 깨우는 push 또는 Claude 네이티브 채널과 같은 자동 도착은 관찰하지 않았다. 리드·워커가 살아 있는 동안의 폴링 간격, 종료 후 재시작, 메시지 만료·재시도·중복 처리까지 실측하기 전에는 `push`라고 안내하면 안 된다. UI와 문서에는 “MCP 받은편지함 폴링”으로 표현하는 것이 정확하다.

## 플랫폼별 판정과 Gaps

| 실행 환경 | Codex → Codex | Codex → Claude | Claude → Codex | Claude → Claude | 근거 범위 |
|---|---|---|---|---|---|
| macOS | 성공 | 성공 | 성공 | 성공 | Claude 직접 인증 계정과 Codex CLI의 실제 별도 세션, MoAI MCP 송신·폴링·접수 |
| Linux 네이티브 | 미실측 | 미실측 | 미실측 | 미실측 | 현재 호스트는 macOS; Linux Docker는 인증된 양쪽 호스트 세션 시험을 대신하지 않음 |
| Windows 네이티브 | 미실측 | 미실측 | 미실측 | 미실측 | 사용자가 접속 가능한 Windows 호스트·인증 원격 러너가 없다고 확인함 |

Mac/Linux의 Claude lead + Claude worker는 네이티브 세션 소켓을 사용할 수 있는 별도 제품 경로로 취급한다. 이 보고서의 Claude → Claude 성공은 **MCP 경로** 실측이며 네이티브 소켓 성능이나 자동 전달을 검증한 결과가 아니다. Windows와 Codex가 포함된 조합에 대해서는 네이티브 소켓을 전제로 하지 않고 MCP 폴링을 기준 경로로 설계해야 한다. Windows가 실제로 동작한다는 주장은 이번 결과로 할 수 없다.

다음 항목은 전부 **NOT-RUN**이다: 실제 `moai todo` ↔ Factory assignment 배차, worker의 worktree 작업과 검사, lead의 증거 심사, PR 생성·병합 요청·원격 병합, 장기 세션의 자동 깨우기, Linux 네이티브, Windows 네이티브, Claude 네이티브 소켓 경로의 비교, 장애 주입·재시작 복구. 이 범위는 현 결과의 잔여 위험이다.

## 개선 계획

1. **High — 런처 계약을 고정한다.** `-f worker`와 `-f agent` 별칭, `--` 뒤 pass-through, `--factory-run` 선택을 Claude/Codex 양쪽 실제 argv로 검증한다. 역할은 호스트 인자 분리 전에 결정하고 잘못된 lead 생성이 없음을 회귀 시험한다. `agent`는 경고를 유지하며 `worker`를 정식 표기로 통일한다.
2. **High — MCP 계약과 준비 상태를 고정한다.** 필수 `task_ref`·`correlation_id`를 도구 스키마와 입력 검증에 일치시킨다. run·slot·session UUID의 바인딩 및 도구 가용성을 진입 시 검사한다. worker가 준비됐다는 기계적 신호를 받은 뒤 배차하고, 등록 지연은 제한된 재시도로 처리한다. Claude의 MCP 설정 자동 발견·신뢰 경로를 별도 실세션에서 확인한다.
3. **High — 카드 배차를 제품 워크플로우에 연결한다.** `picked` 카드가 Factory `assignments`에 기록되고 worker가 카드 UUID·revision을 받은 뒤에만 작업하도록 한다. 배차·접수·결과 보고·lead 수락은 동일한 작업 식별자와 기대 revision으로 검증한다. 보고가 배차보다 먼저 온 경우 완료로 처리하지 않는다.
4. **High — 실작업 수용 시험을 만든다.** 격리 Git 원격과 새 worktree에서 네 조합 각각 카드 발행 → 배차 → 수정 → 범위 검사 → 보고 → lead 검증 → PR·병합 요청 → 원격 병합까지 실행한다. 현재 라이브 픽스처와 운영 런처 시험을 분리해 각각의 통과 범위를 명시한다. PR/병합 완료를 브로커 메시지 수락과 혼동하지 않는다.
5. **Medium — 폴링 운영을 명시한다.** 백오프·만료·재접속·중복 방지·worker 종료·lead 교체를 테스트하고, 지연과 재시도 횟수를 측정한다. 자동 깨우기가 필요하다면 별도 호스트 제어 경로를 설계·실세션 검증한 후에만 기능으로 표기한다.
6. **Medium — 플랫폼을 각각 인증한다.** macOS 결과를 Linux·Windows로 확대하지 않는다. 인증된 Linux 네이티브와 Windows 네이티브 러너에서 같은 네 조합, 파일 결과, 메시지 상태, 재시작을 각각 수집한다. Windows 호스트가 준비되기 전에는 Windows를 계속 `미실측`으로 표시한다.
7. **Medium — 토큰·지연 비용을 계측한다.** 이번 Claude 운영 로그에 큰 캐시 읽기량이 보였지만 단일 호출만으로 구조적 낭비를 확정할 수 없다. 공통 작업을 최소 MCP 설정과 전체 하네스로 반복해 도구 설명·초기 프롬프트·스킬 로딩·폴링 비용과 성공률을 함께 비교한다. 기능을 줄이는 결정은 동등 성공률 확인 뒤에 한다.

## Residual-risk

이번 네 성공은 격리 프로젝트의 짧은 비대화형 턴에 한정된다. 사용자 실제 저장소의 동시 카드, 긴 작업, 실패 복구, 승인 정책, 원격 Git 통합은 다르다. Claude는 명시적 MCP 설정을 요구했고, 구분자 버그에는 우회 문법을 사용했다. 따라서 현 단계의 정확한 제품 문구는 **“macOS에서 네 호스트 조합의 MoAI MCP 명시적 메시지 왕복을 실측”**이다. **“전체 Factory 자동 운영·크로스 플랫폼 완료”는 미검증**이다.
