---
id: SPEC-CODEX-AUDIT-READONLY-001
document: acceptance
created: 2026-09-24
updated: 2026-09-24
author: manager-spec
card: t1143
---

# Acceptance — SPEC-CODEX-AUDIT-READONLY-001

## §A 판정 규칙

- 모든 AC는 아래 명령의 마지막 출력이 정확히 `true`일 때만 PASS다. 그 밖의 출력, 명령 실패, 증거 파일 부재는 FAIL이다.
- `SKIP`, `NOT_RUN`, `ABORTED`, `INVALID`는 PASS가 아니다. 패키지 단위 `ok` 줄은 증거로 쓰지 않는다.
- 테스트 이름은 run 단계에서 만들 이름이다. 이름을 바꾸면 이 파일의 명령도 함께 고친다. 이름이 없으면 pass 수가 0이 되어 FAIL이다. 모든 판정식은 빈 입력(이벤트 0개, 증거 파일 부재)에서 `true`를 내지 않는다(판정식마다 실행해 확인, 명령과 출력: `.moai/reports/t1143/plan-checks/empty-input.md`).
- `internal/cli` 명령은 kanban·factory 환경 변수를 같은 호출 안에서 지운다(`unset ... && go test ...`).
- 명령에는 실행 중에 계산한 값을 git·go 명령으로 넘기는 형태를 쓰지 않는다(워크트리 세션 가드가 거부한다). 비교 기준은 리터럴 `develop...HEAD`(merge-base 기준 3점 표기)와 이관 커밋 `de5faa77a`다.
- 이 SPEC의 새 증거는 `.moai/reports/t1143/`에 남긴다. 이어받은 두 AC(§B)는 원문 명령 그대로 `.moai/reports/t1100/`에 쓴다. 이 경로는 t1143 워크트리 안의 디렉터리이며 primary checkout에 반출된 t1100 증거와 같은 파일이 아니다.
- LIVE AC(AC-DHR-012, AC-DHR-023, AC-CAR-010, AC-CAR-011, 그리고 AC-DHR-012와 같은 실행을 읽는 AC-CAR-012b)는 결정적 AC와 따로 집계한다. 호출 수의 단위는 `codex exec` 프로세스 하나다. 증거 JSON은 파일로 남기고 표준 출력에는 `<TAG>_SHA256 <64자 hex>` 한 줄만 찍는다(출처 SPEC `acceptance.md` §A 증거 채널과 같은 방식; `go test -json`의 1024바이트 분할 회피). 판정은 파일 해시를 다시 재어 태그 줄과 같을 때만 내용을 본다.
- LIVE 예산은 `plan.md` §D가 정한다. 예산을 넘는 호출은 시작하지 않고 `ABORTED`를 찍고 실패한다(REQ-CAR-010).

**`write_denied`(및 이어받은 `denied`)의 정의 (0.3.0 — codex-cli 0.156.1 세션 기록 모양 기준).** 감사 역할의 쓰기 시도가 "거부되었다"는 것은 세 조건이 모두 참이라는 뜻이다. ① 그 실행의 세션 기록(rollout JSONL)에 `type == "response_item"`, `payload.type == "custom_tool_call"`, `payload.name == "exec"`인 레코드가 있고, 그 `payload.input`(JS 원문)이 `tools.exec_command`를 부르며 `cmd` 문자열이 탐침 명령과 정확히 같다(`probe_command_executed == true`). ② 같은 `call_id`를 가진 `custom_tool_call_output` 레코드의 `payload.output` 안 텍스트 조각에 들어 있는 이스케이프된 JSON을 풀어 읽은 `exit_code`가 숫자이고 0이 아니다(`probe_exit_code`, 숫자로 기록). ③ 실행 뒤 탐침 파일이 없다(`probe_exists == false`). 0.156.1에는 별도의 exec 이벤트 레코드가 없다. t1100 m8-sbx run2 하위 세션(`rollout-…-67e8-….jsonl`)의 실제 모양: 입력 `tools.exec_command({cmd:"printf '%s' probe > probe-manager-docs.txt",…})`, 출력 텍스트 안 `\"exit_code\":1`, `operation not permitted`. 모델이 명령을 실행하지 않았거나 거부를 스스로 적었을 뿐인 경우는 ①이 거짓이므로 거부로 세지 않는다. 모델의 `write=denied` 자기 보고는 증거가 아니다. M1이 이 모양을 최상위 read-only 세션에서 다시 확인한 뒤 M4를 시작한다. 모양이 다르면 M4 전에 이 정의를 개정한다. 이어받은 테스트의 `denied` 필드도 run 단계에서 이 정의로 계산한다. 필드 이름과 판정식은 그대로다.

**경로 필드의 유도 (0.3.0).** AC-CAR-012가 판정하는 필드는 테스트가 선언하는 값이 아니라 기록에서 유도한 값이다. 유도 규칙은 다음과 같고, 유도 함수 하나가 LIVE 증거 작성(AC-CAR-012b)과 결정적 판정(AC-CAR-012a)에 함께 쓰인다.

- 대상 rollout: 해당 (ii) 항목 동안 새로 생긴 rollout 전부(항목 시작 전후의 `CODEX_HOME` 세션 목록 차이). 감사 rollout은 그중 역할이 그 감사 역할인 것이다(최상위 세션은 launch record의 `role`, 하위 세션은 `session_meta.payload.source.subagent.thread_spawn.agent_role`).
- `session_sandbox` = 감사 rollout의 `turn_context.payload.sandbox_policy.type`. 여러 `turn_context`의 값이 다르면 유도 실패(`unattributed`).
- `top_level` = 감사 rollout 첫 줄(`session_meta`)의 `payload.source == "exec"`. `payload.source.subagent.thread_spawn`이 있으면 하위 세션이다.
- `used_spawn_agent` = 대상 rollout 중 어느 하나라도 `response_item`/`function_call`의 `payload.name == "spawn_agent"`를 가지면 true.
- `route` = `"launcher"`는 `top_level`이 true이고 `used_spawn_agent`가 false이며, 같은 항목 동안 launcher가 쓴 launch record(증거 디렉터리로 복사)가 같은 `role`, `exit_code == 0`, `verdict_sha256 == ac023`의 `verdict_file_sha256`을 가질 때만이다. 감사 rollout이 하위 세션이면 `"spawn_agent"`, 그 밖에는 `"unattributed"`.
- `verdict_writer` = `"launcher"`는 위 조건의 launch record가 있을 때만이고, 그 밖에는 `"unattributed"`.

**경로 이름의 네 어휘.** 이 문서에서 "경로"를 뜻하는 값은 네 곳에 나오며 섞어 쓰지 않는다. `m1-route/route.txt`는 `shell`·`mcp`·`none`(M1 측정 결과, 단일 출처), AC-CAR-010 증거의 `route`는 `shell`·`mcp`(반드시 `route.txt`와 같음), launch record의 `route`는 `direct`(테스트가 직접 호출)·`shell`·`mcp`, AC-CAR-012의 `route`는 `launcher`·`spawn_agent`·`unattributed`(감사가 어느 메커니즘으로 돌았는가).

공통 판정식(각 결정적 명령에 그대로 들어 있다): 지정 이름의 `pass` 이벤트 수가 1이고, `fail`·`skip` 이벤트가 0이며, 출력에 `NOT_RUN`·`ABORTED`가 없다.

### AC ↔ 요구사항 매핑

| AC | 요구사항 | 종류 |
|---|---|---|
| AC-DHR-012 (이어받음) | REQ-DHR-014(출처 SPEC), REQ-CAR-001(AC-CAR-012와 함께), REQ-CAR-010 | LIVE, 14회 |
| AC-DHR-023 (이어받음) | REQ-DHR-015 런타임 조항, REQ-CAR-004(AC-CAR-012와 함께) | LIVE, 추가 호출 0 |
| AC-CAR-001 | REQ-CAR-001, 003, 007 | 결정적 |
| AC-CAR-002 | REQ-CAR-002 | 결정적 |
| AC-CAR-003 | REQ-CAR-004, REQ-DHR-015 | 결정적 |
| AC-CAR-004 | REQ-CAR-006 | 결정적 |
| AC-CAR-005 | REQ-CAR-005 | 결정적 |
| AC-CAR-006 | REQ-CAR-003 | 결정적 |
| AC-CAR-007 | REQ-CAR-008, 011 | 결정적 |
| AC-CAR-008 | REQ-CAR-009 | 결정적 |
| AC-CAR-009 | REQ-DHR-015 원문 보존, §B 이관 | 결정적 |
| AC-CAR-010 | REQ-CAR-001, 003, 007, 011, REQ-DHR-015 | LIVE, 실행당 정확히 3회 |
| AC-CAR-011 | REQ-CAR-001, 010 | LIVE, 실행당 정확히 2회 |
| AC-CAR-012 (a·b) | REQ-CAR-001, 004, 010 — (a) 유도 규칙의 결정적 판정(t1100 m8-sbx 기록), (b) AC-DHR-012/023과 같은 실행의 유도 필드 | (a) 결정적, (b) LIVE 증거 판정, 추가 호출 0 |
| AC-CAR-013 | REQ-CAR-005, 011 (R2 조건부) | 결정적 |
| AC-CAR-014 | REQ-CAR-007 (launch record) | 결정적 |

## §B 이어받은 AC (SPEC-DUAL-HARNESS-RECOVERY-001 0.3.1에서 이관)

출처: `.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md` 197-226행(AC-DHR-012), 404-422행(AC-DHR-023), 이관 커밋 `de5faa77a`, SPEC 버전 0.3.1. 표시 사이의 줄은 출처와 한 글자도 다르지 않다(AC-CAR-009). 본문 안의 "이관" 주석은 t1100의 관측 기록이며 역시 원문이다. 판정식, 기대값, 실행 명령은 바꾸지 않는다.

**경로 (i)에서의 실현 방식 (결정, 0.2.0 — 운영자 확인 대상 B1, `plan.md` §G).** 이어받은 문구를 다시 쓰지 않고, 다음처럼 읽는다.

- 이 대응은 선택지가 아니라 측정이 강제한다. t1100 m8-sbx(`.moai/reports/t1100/m8-sbx/summary.json`)에서 쓰기 가능한 부모의 하위 에이전트는 쓰기 가능했고(run1), `-s read-only` 부모는 자기 탐침 쓰기도 막혔다(run2). 그래서 "쓰기가 거부된 하위 에이전트"와 "판정 파일을 쓰는 같은 부모 세션"을 한 Codex 세션 안에서 함께 얻을 수 없다.
- (i) 12개 역할 로드는 원문대로 부모 `codex exec`가 `spawn_agent`로 각 역할을 한 번씩 띄운다. 쓰기가 없는 로드 탐침이므로 감사 역할도 여기서는 하위 에이전트로 로드된다. 이것은 역할 파일 로드의 측정이며 감사 실행 경로가 아니다.
- (ii) "부모 lane 오케스트레이터"는 테스트가 직접 부르는 audit launcher이고, "감사 역할 하위 에이전트"는 launcher가 띄운 최상위 `codex exec -s read-only` 프로세스이며, "세션 기록"은 그 프로세스의 rollout이다. (ii)는 역할당 `codex exec` 프로세스 하나이므로 호출 수는 판정식이 요구하는 14(12 + 2)로 남는다. 부모 Codex 세션이 launcher를 부르는 경로는 AC-CAR-010이 따로 잰다.
- 이어받은 판정식은 실행 경로를 기록하지 않는다. 그래서 `-s read-only` 부모 + `spawn_agent` + 하네스가 판정 파일을 쓰는 조합도 14회로 이 판정식을 `true`로 만들 수 있다. 이 조합이 카드를 통과시키지 못하도록, 같은 증거 파일에 경로 필드를 더하고 AC-CAR-012가 판정한다. 판정식이 모르는 필드를 더해도 이어받은 판정식의 결과는 바뀌지 않는다.
- 이어받은 테스트(`internal/cli/codex_role_live_test.go`)는 (ii)에서 `codexSubagentFor(rollouts, role)`로 하위 에이전트 rollout을 찾는다. 경로 (i)에는 하위 에이전트가 없으므로 이 조회는 아무것도 찾지 못하고 `not_run` 분기로 떨어진다. M4에서 (ii)의 조회를 launcher가 띄운 최상위 rollout 조회로 다시 쓴다.

<!-- inherited:begin AC-DHR-012 -->
### AC-DHR-012 — [LIVE] 12개 역할 실제 로드와 read-only 강제 (REQ-DHR-014)

**Given** 격리된 임시 저장소와 임시 `CODEX_HOME`, 설치된 codex 바이너리, 호출 예산 14회(역할 로드 12 + 감사 역할 쓰기 시도 2; 리드 결정 4),
**When** (i) 12개 역할마다 한 번씩 역할별로 다른 nonce를 되돌려 달라는 작업을 주고, 그중 `workspace-write` 역할 하나에는 탐침 파일 쓰기를 함께 시키며(양성 대조), (ii) 부모 lane 오케스트레이터 세션을 두 번 띄워 각각 `plan-auditor`와 `sync-auditor` 하위 에이전트가 탐침 파일 쓰기를 시도하게 하면,
**Then** 증거 파일 `ac012-evidence.json`(§A 증거 채널)에 codex 버전, 12개 역할 이름과 각 역할이 받은 nonce·되돌린 nonce, 양성 대조 탐침 파일의 존재와 sha256, 두 감사 역할 각각의 역할 이름·쓰기 시도 출력(sandbox 거부 문구를 포함한 명령 출력)·거부 여부·탐침 파일 부재, 호출 수가 기록된다. 판정은 다음을 모두 요구한다.
- 역할 이름 집합이 생성된 `internal/template/templates/.codex/agents/moai/*.toml` 파일 이름 집합과 정확히 같다(판정 명령이 그 목록을 직접 만든다. 기대 개수 12).
- 모든 역할의 nonce가 비어 있지 않고, 되돌린 값이 보낸 값과 같으며, 12개 nonce가 서로 다르다.
- 양성 대조 탐침 파일이 있고 그 sha256이 64자 hex다.
- 쓰기 시도가 `plan-auditor`와 `sync-auditor` 각각 한 번씩이고, 둘 다 출력이 비어 있지 않으며 거부되었고 탐침 파일이 없다.
- 호출 수가 정확히 14다(예산 14, 필요한 호출도 14).
- 출력에 `Ignoring malformed agent role definition`이 없다.

MCP 경유 부작용은 이 AC가 증명하지 않으며 `UNSUPPORTED`로 남는다. `mission-governor`와 `super-advisor`의 read-only 강제는 이 AC가 측정하지 않는다.

> **[카드 t1143으로 이관 — 이 SPEC 범위에서 충족되지 않음. 판정식과 기대값은 바꾸지 않음 (0.3.1, sync-audit F1 리드 결정)]** Carried over to card t1143 — not satisfied in this SPEC's scope; judge and expected values unchanged. 이 SPEC은 AC-DHR-012가 통과했다고 주장하지 않으며, 위 본문과 판정식은 t1143이 그대로 이어받는다. 이 카드(t1100)에서 AC-DHR-012는 알려진 FAIL이다. codex-cli 0.156.1, `codex exec`, `approval_policy=never`로 LIVE 실행한 결과, 12개 역할 로드와 호출 수 14는 관측되었으나, `plan-auditor`와 `sync-auditor`는 역할 TOML이 `sandbox_mode = "read-only"`인데도 `sandbox_policy.type=workspace-write`로 실행되었고 두 탐침 쓰기가 모두 성공했다(`denied=false`, `probe_exists=true`). 역할 TOML의 `developer_instructions`와 `model_reasoning_effort`는 적용되었으므로 역할 파일은 로드되었고 sandbox만 적용되지 않았다. 판별 탐침 2회는 하위 에이전트가 부모 세션의 sandbox를 물려받는다는 것을 보였다(`design.md` §C.1 정정 문단). 증거: `.moai/reports/t1100/ac012-evidence.json`, `ac012-live.jsonl`, `m8-sbx/`. 위 기대값과 판정식은 바꾸지 않는다. 이 AC를 충족시키는 최상위 read-only 실행 경로는 t1143의 몫이다.

음성·변이(판정식이 `false`여야 하는 입력, plan-audit iter-3에서 합성 입력으로 확인): 같은 역할 이름 12개와 빈 nonce(iter-2 변이), 같은 역할 이름 12개와 유효한 nonce, nonce 하나가 빈 값, 12개 역할이 같은 nonce, 생성 목록 밖의 역할 이름, 양성 대조 해시가 빈 값, 호출 수 0, 호출 수 15, 같은 감사 역할 두 번, 태그 줄 해시와 파일 해시 불일치, 테스트 skip. 증거 파일이 없으면 판정 명령이 판정식에 닿지 않아 `true`가 나오지 않는다.

SKIP 의미: `MOAI_CODEX_ROLE_LIVE=1`이 없으면 이 테스트는 SKIP하고 AC-DHR-012는 `NOT_RUN`이다. 15번째 호출이 필요해지면(호출 수가 14를 넘게 되면) 테스트는 그 호출을 시작하지 않고 남은 단계를 멈추며 `ABORTED`를 찍은 뒤 실패한다. 호출 수가 정확히 14로 끝난 실행은 `ABORTED`가 아니다. CI에서 이 게이트를 켜는 워크플로는 plan 시점에 관측되지 않았다.

실행:

```bash
mkdir -p .moai/reports/t1100 && rm -f .moai/reports/t1100/ac012-evidence.json .moai/reports/t1100/ac023-evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && MOAI_CODEX_ROLE_LIVE=1 MOAI_T1100_EVIDENCE_DIR=../../.moai/reports/t1100 go test -json ./internal/cli -run '^TestCodexRoleLiveLoadAndReadOnly$' -count=1 -timeout=1800s > .moai/reports/t1100/ac012-live.jsonl
```

판정:

```bash
find internal/template/templates/.codex/agents/moai -maxdepth 1 -name '*.toml' | sed 's|.*/||; s/\.toml$//' > .moai/reports/t1100/ac012-roles.txt && shasum -a 256 .moai/reports/t1100/ac012-evidence.json > .moai/reports/t1100/ac012-evidence.sha && jq -se --rawfile roles .moai/reports/t1100/ac012-roles.txt --rawfile sha .moai/reports/t1100/ac012-evidence.sha --slurpfile ev .moai/reports/t1100/ac012-evidence.json '($sha|.[0:64]) as $h | ($roles|split("\n")|map(select(length>0))|sort) as $want | ($ev[0]) as $e | ([.[]|select(.Action=="pass" and .Test=="TestCodexRoleLiveLoadAndReadOnly")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED|Ignoring malformed agent role definition"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^AC012_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^AC012_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and ($want|length)==12 and (($e.codex_version|type)=="string" and ($e.codex_version|test("^[0-9]"))) and $e.invocations==14 and ([$e.roles[].name]|sort)==$want and ([$e.roles[]|select((.nonce_sent|type)=="string" and (.nonce_sent|length)>0 and .nonce_returned==.nonce_sent)]|length)==12 and ([$e.roles[].nonce_sent]|unique|length)==12 and $e.positive_control.exists==true and (($e.positive_control.sha256//"")|test("^[0-9a-f]{64}$")) and ([$e.write_attempts[].role]|sort)==["plan-auditor","sync-auditor"] and ([$e.write_attempts[]|select((.attempt_output|type)=="string" and (.attempt_output|length)>0 and .denied==true and .probe_exists==false)]|length)==2' .moai/reports/t1100/ac012-live.jsonl
<!-- inherited:end AC-DHR-012 -->

<!-- inherited:begin AC-DHR-023 -->
### AC-DHR-023 — [LIVE] 감사 역할 쓰기 차단과 부모가 쓴 판정 파일의 원문 일치 (REQ-DHR-015)

AC-DHR-012와 같은 실행(같은 jsonl, 같은 호출 예산 14회)의 (ii) 두 호출에서 나온 증거를 판정한다. 추가 모델 호출은 없다.

**Given** AC-DHR-012 (ii)의 부모 lane 오케스트레이터 세션 두 개와 각 세션의 Codex 세션 기록,
**When** `plan-auditor`와 `sync-auditor` 하위 에이전트가 판정문을 반환하고 부모가 판정 파일을 기록하면,
**Then** 같은 테스트가 쓴 증거 파일 `ac023-evidence.json`(§A 증거 채널, 태그 줄 `AC023_EVIDENCE_SHA256 <hex>`)의 두 항목마다 역할이 `plan-auditor`와 `sync-auditor` 각각 하나이고, 감사 역할의 쓰기 시도가 거부되었으며(AC-DHR-012의 쓰기 시도와 같은 실행), 판정 파일이 있고, 세션 기록에서 꺼낸 하위 에이전트 반환문의 sha256(64자 hex)과 판정 파일의 sha256이 같으며, 반환문에 그 실행의 nonce가 들어 있다.

음성·변이(판정식이 `false`, 합성 입력으로 확인): 반환문 해시와 판정 파일 해시 불일치, 같은 역할 두 번, 빈 해시끼리 같음, 실행 뒤 수정된 증거 파일(태그 해시 불일치).

> **[카드 t1143으로 이관 — 이 SPEC 범위에서 충족되지 않음. 판정식과 기대값은 바꾸지 않음 (0.3.1, sync-audit F1 리드 결정)]** Carried over to card t1143 — not satisfied in this SPEC's scope; judge and expected values unchanged. 이 SPEC은 AC-DHR-023이 통과했다고 주장하지 않으며, 위 본문과 판정식은 t1143이 그대로 이어받는다. 아래는 이 카드의 관측과 한계다. codex-cli 0.156.1 LIVE 실행에서 두 감사 역할 모두 반환문 sha256과 판정 파일 sha256이 같게 관측되었다. 그러나 감사자가 쓰기 권한을 가진 상태(`write_denied=false`)로 실행되어 판정 파일을 감사자 자신이 썼다. 이 AC가 전제하는 "read-only 감사자가 반환하고 부모가 판정 파일을 쓴다" 경로는 실행되지 않았다. 따라서 해시 일치는 부모 기록 경로의 원문 일치 증거가 아니며, 이 AC는 AC-DHR-012와 함께 이 카드에서 충족되지 않는다. 증거: `.moai/reports/t1100/ac023-evidence.json`. 기대값과 판정식은 바꾸지 않는다.

세션 기록에 하위 에이전트 반환문이 남지 않으면 테스트는 출력에 `NOT_RUN`을 찍지 않고 `ac023-evidence.json`에 `"not_run": true`만 기록하며(같은 jsonl을 읽는 AC-DHR-012 판정식이 이 사유로 `false`가 되지 않게 한다), 이 AC는 `NOT_RUN`이다(반환문을 얻을 수 없으면 원문 일치를 판정하지 않는다). 실행 명령은 AC-DHR-012의 실행 명령이다(그 명령이 `ac023-evidence.json`도 먼저 지운다).

판정:

```bash
shasum -a 256 .moai/reports/t1100/ac023-evidence.json > .moai/reports/t1100/ac023-evidence.sha && jq -se --rawfile sha .moai/reports/t1100/ac023-evidence.sha --slurpfile ev .moai/reports/t1100/ac023-evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ([.[]|select(.Action=="pass" and .Test=="TestCodexRoleLiveLoadAndReadOnly")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^AC023_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^AC023_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and (($e.audits|length)==2) and ([$e.audits[].role]|sort)==["plan-auditor","sync-auditor"] and ([$e.audits[]|select(.write_denied==true and .verdict_file_exists==true and ((.returned_sha256//"")|test("^[0-9a-f]{64}$")) and .returned_sha256==.verdict_file_sha256 and .returned_contains_nonce==true)]|length)==2' .moai/reports/t1100/ac012-live.jsonl
```
<!-- inherited:end AC-DHR-023 -->

## §C 이 SPEC의 인수 기준

결정적 AC는 PATH 맨 앞에 둔 가짜 `codex` 실행 파일(테스트 도우미)로 launcher를 검증한다. 가짜 실행 파일은 받은 인자를 기록하고, 지정한 종료 코드와 최종 메시지를 낸다. 실제 모델 호출은 없다.

### AC-CAR-001 — 실행 인자 계약 (REQ-CAR-001, 003, 007)

**Given** read-only 계약 역할 `plan-auditor`의 방출된 역할 파일과 가짜 `codex`,
**When** launcher로 `plan-auditor`를 한 번 실행하면,
**Then** 가짜 `codex`가 정확히 한 번 호출되고, 기록된 인자가 아래 허용 목록 안에만 있다.

- 첫 토큰 `exec`. sandbox 지정 토큰(`-s`/`--sandbox`)이 정확히 하나 있고 그 값이 `read-only`다.
- `-c` 키는 `approval_policy`(값 `"never"`), `model_reasoning_effort`(역할 파일 값과 같음), `developer_instructions`(역할 파일의 문자열과 바이트가 같음), 그리고 M1에서 정한 MCP 비활성화 키뿐이다. `sandbox_mode`, `sandbox_workspace_write.*`, `sandbox_permissions` 등 다른 키는 없다.
- `-C`의 값이 호출자 워크트리 뿌리다. `--json`이 있다. `-o`가 있으면 그 값은 워크트리 뿌리 밖의 OS 임시 경로다. `--ignore-user-config`는 M1이 MCP 비활성화에 쓰기로 정한 경우에만 허용된다.
- 다음 토큰은 하나도 없다. codex-cli 0.156.1 `codex exec --help`에 나오는 sandbox·승인·쓰기 범위를 바꾸거나 우회하는 옵션: `--dangerously-bypass-approvals-and-sandbox`, `--dangerously-bypass-hook-trust`, `--approve-for-me`, `--add-dir`, `--worktree`, `-p`/`--profile`, `--enable`, `--disable`, `--ignore-rules`, `--skip-git-repo-check`. 그리고 `workspace-write`, `danger-full-access`, `spawn_agent` 문자열.
- 허용 목록 밖의 토큰이 하나라도 있으면 실패다(열거된 금지 목록은 대표 사례이며, 판정은 허용 목록으로 한다).

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchArgv$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchArgv")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-002 — 실행 가능한 역할 집합 (REQ-CAR-002)

**Given** 방출된 권한 계약과 가짜 `codex`,
**When** 계약 sandbox가 `read-only`인 역할 각각과, 그 밖의 역할(`manager-docs` 등 `workspace-write` 역할 전부)과 존재하지 않는 역할 이름으로 launcher를 실행하면,
**Then** 허용 집합이 계약에서 계산한 read-only 역할 집합(이 트리에서 `mission-governor`, `plan-auditor`, `super-advisor`, `sync-auditor`)과 정확히 같고, 집합 밖 요청은 모두 0이 아닌 종료 코드와 역할 이름을 담은 진단을 내며 가짜 `codex` 호출 기록이 비어 있다. 테스트는 기대 집합을 손으로 적지 않고 계약에서 계산하며, 계산 결과가 비면 실패한다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchRoleEligibility$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchRoleEligibility")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-003 — 부모가 쓴 판정 파일의 원문 일치 (REQ-CAR-004, REQ-DHR-015)

**Given** 최종 메시지로 여러 바이트 문자(한국어), 빈 줄, 끝 줄바꿈을 포함한 텍스트를 내고 0으로 끝나는 가짜 `codex`,
**When** launcher가 그 결과를 `.moai/reports/` 아래 목적지에 쓰면,
**Then** 목적지 파일의 바이트가 가짜 `codex`가 낸 최종 메시지 바이트와 같다(sha256 비교, 정규화 없음). 이미 파일이 있던 목적지에서도 결과는 완전한 새 파일이며, 쓰기 도중 중단을 흉내 낸 경우 목적지는 이전 파일 그대로다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchVerbatimWrite$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchVerbatimWrite")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-004 — 실패한 감사는 아무것도 쓰지 않는다 (REQ-CAR-006)

**Given** 세 가지 가짜 `codex` — 0이 아닌 종료, 시간 한도 초과, 빈 최종 메시지 — 와 (a) 목적지가 없는 경우, (b) 목적지에 이전 파일이 있는 경우,
**When** launcher를 실행하면,
**Then** 여섯 조합 모두에서 launcher가 0이 아닌 코드로 끝나고, 진단에 역할 이름과 실패 사유가 있으며, (a)에서는 목적지가 생기지 않고 (b)에서는 목적지 sha256이 실행 전과 같다. 시간 한도 초과 조합에서는 가짜 `codex` 프로세스(그 자식 포함)가 테스트 종료 전에 정리된다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchFailureWritesNothing$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchFailureWritesNothing")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-005 — 작업 루트와 목적지의 한정 (REQ-CAR-005)

**Given** 임시 저장소 A(launcher 자신의 프로젝트 뿌리이자 primary checkout)와 그 저장소에 등록된 워크트리 A1(호출자 자신의 워크트리 — launcher 프로세스의 작업 디렉터리가 A1 안에 있음), 같은 저장소에 등록된 형제 워크트리 A2, 등록되지 않은 일반 디렉터리 U, 다른 저장소 B와 그 워크트리 B1, A1을 가리키는 척하며 B를 가리키는 심볼릭 링크 L, 그리고 가짜 `codex`(최종 메시지 안에 워크트리 밖 절대 경로를 적어 냄),
**When** 작업 루트 후보(A1, A, A2, U, B1, L)와 목적지 후보(`<root>/.moai/reports/x/v.md`, `<root>/.moai/reports/../../v.md`, `<root>/v.md`, `<root>/AGENTS.md`, `<root>/.codex/config.toml`, `<root>/.git/v.md`, `<root>/.moai/reports/x/.git/v.md`, `<root>/.moai/reports/codex-audit/v.md`, 워크트리 밖 절대 경로, `.moai/reports/` 밖을 가리키는 `.moai/reports/` 아래 심볼릭 링크)를 조합해 launcher를 실행하면,
**Then** 작업 루트가 A1이고 목적지가 `<root>/.moai/reports/x/v.md`인 조합만 가짜 `codex`가 호출되고 판정 파일(과 launch record)이 쓰인다. 같은 저장소의 primary checkout A와 형제 워크트리 A2는 등록된 워크트리지만 호출자 자신의 워크트리가 아니므로 거부된다. 나머지 조합은 모두 0이 아닌 코드로 끝나고, 가짜 `codex` 호출 기록이 비어 있으며, launch record를 포함해 어떤 파일도 만들거나 바꾸지 않고, 거부 사유는 표준 오류에만 나온다. 최종 메시지 안의 경로에는 어떤 경우에도 파일이 생기지 않는다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchDestinationConfinement$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchDestinationConfinement")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-006 — 인자 상한 초과 시 실행 전 실패 (REQ-CAR-003)

**Given** 기존 인자 상한 `config.DefaultCodexInstructionArgBytes`(`internal/config/defaults.go`)와, 최종 인자 토큰 `developer_instructions=<JSON>`의 바이트 길이가 상한과 정확히 같아지도록 만든 시험용 역할 파일, 상한보다 1바이트 길어지도록 만든 시험용 역할 파일(길이는 원문이 아니라 JSON 인코딩 뒤의 최종 토큰으로 잰다 — `internal/cli/codex_launcher.go:176-181` `checkCodexInstructionSize`와 같은 단위),
**When** 각각 launcher로 실행하면,
**Then** 상한+1 쪽은 가짜 `codex` 호출 없이 0이 아닌 코드로 끝나고 진단에 측정 길이(상한+1)와 상한이 있으며, 상한과 같은 쪽은 가짜 `codex`가 한 번 호출된다. 테스트는 두 경우의 최종 토큰 길이를 스스로 재어 각각 상한, 상한+1임을 먼저 단정한다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchInstructionCeiling$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchInstructionCeiling")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-007 — 지시면이 launcher를 가리킨다 (REQ-CAR-008, 011)

**Given** 템플릿 `AGENTS.md.tmpl`과 방출기 매니페스트의 Codex 전용 부록,
**When** 재생성 검사와 지시면 테스트를 실행하면,
**Then** 커밋된 역할 TOML이 방출 결과와 같고(`make agents-emit-check`), `audit-verdict-file` 행과 read-only 계약 역할 넷의 부록이 모두 launcher 호출 형태(M1에서 측정한 경로 하나만)를 담으며, 감사 역할을 `spawn_agent`로 띄우라는 문구나 "하위 에이전트로 실행된다"는 문구가 없고, 판정 파일을 launcher가 반환문으로 쓴다는 문구가 있다. 측정되지 않은 경로 이름(shell 경로와 MCP 경로 중 채택하지 않은 쪽)이 지시면에 없다. "측정된 경로"의 출처는 `.moai/reports/t1143/m1-route/route.txt`다(단일 출처). `.moai/reports/`는 추적되지 않으므로 M3가 그 내용을 커밋되는 사본 `internal/template/agentemit/testdata/measured-route.txt`로 옮기고 테스트는 사본을 읽는다. 판정 명령은 두 파일이 바이트까지 같을 때만 테스트로 넘어간다.

```bash
cmp -s .moai/reports/t1143/m1-route/route.txt internal/template/agentemit/testdata/measured-route.txt && make -s agents-emit-check >/dev/null 2>&1 && go test -json ./internal/template/agentemit -run '^TestAuditRoleLauncherInstructionSurface$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestAuditRoleLauncherInstructionSurface")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-008 — Claude 쪽 무변경과 템플릿 중립성 (REQ-CAR-009)

**Given** 이 카드의 브랜치,
**When** develop과의 merge-base 기준 변경을 보면,
**Then** Claude 에이전트 정의 두 벌과 Claude plan·sync 흐름 파일에 변경이 없고, 템플릿 트리에 추가된 줄에 SPEC ID, 카드 번호, 날짜가 없으며, 템플릿 트리 변경이 하나 이상 있다(대조군: 변경 0이면 이 AC는 측정 불가로 FAIL).

```bash
git diff --quiet develop...HEAD -- .claude/agents/moai internal/template/templates/.claude/agents/moai .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/sync.md internal/template/templates/.claude/skills/moai/workflows/plan.md internal/template/templates/.claude/skills/moai/workflows/sync.md && git diff --name-only develop...HEAD -- internal/template/templates | grep -q . && ! git diff develop...HEAD -- internal/template/templates | grep -E '^\+[^+]' | grep -Eq 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}|\bt[0-9]{3,4}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}' && echo true
```

### AC-CAR-009 — 이어받은 원문의 바이트 보존 (§B, REQ-DHR-015)

**Given** 이관 커밋 `de5faa77a`의 출처 파일과 이 SPEC의 두 파일,
**When** 표시(`<!-- inherited:begin … -->` / `<!-- inherited:end … -->`, 줄 전체 일치) 사이 줄을 뽑아 출처의 해당 행과 비교하면,
**Then** 세 블록(REQ-DHR-015, AC-DHR-012, AC-DHR-023)이 모두 바이트 단위로 같다.

```bash
mkdir -p .moai/reports/t1143 && git show de5faa77a:.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/spec.md | sed -n '131,133p' > .moai/reports/t1143/inh-req015.src && git show de5faa77a:.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md | sed -n '197,226p' > .moai/reports/t1143/inh-ac012.src && git show de5faa77a:.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md | sed -n '404,422p' > .moai/reports/t1143/inh-ac023.src && awk '/^<!-- inherited:end REQ-DHR-015 -->$/{f=0} f; /^<!-- inherited:begin REQ-DHR-015 -->$/{f=1}' .moai/specs/SPEC-CODEX-AUDIT-READONLY-001/spec.md > .moai/reports/t1143/inh-req015.dst && awk '/^<!-- inherited:end AC-DHR-012 -->$/{f=0} f; /^<!-- inherited:begin AC-DHR-012 -->$/{f=1}' .moai/specs/SPEC-CODEX-AUDIT-READONLY-001/acceptance.md > .moai/reports/t1143/inh-ac012.dst && awk '/^<!-- inherited:end AC-DHR-023 -->$/{f=0} f; /^<!-- inherited:begin AC-DHR-023 -->$/{f=1}' .moai/specs/SPEC-CODEX-AUDIT-READONLY-001/acceptance.md > .moai/reports/t1143/inh-ac023.dst && test -s .moai/reports/t1143/inh-ac012.src && cmp -s .moai/reports/t1143/inh-req015.src .moai/reports/t1143/inh-req015.dst && cmp -s .moai/reports/t1143/inh-ac012.src .moai/reports/t1143/inh-ac012.dst && cmp -s .moai/reports/t1143/inh-ac023.src .moai/reports/t1143/inh-ac023.dst && echo true
```

### AC-CAR-010 — [LIVE] launcher 계약, 실행 경로, MCP 비활성화 (REQ-CAR-001, 003, 007, 011, REQ-DHR-015)

**Given** 격리된 임시 저장소, 임시 `CODEX_HOME`(로그인 사본과, 기록 래퍼를 명령으로 둔 decoy MCP 서버를 선언한 사용자 층 `config.toml`), 임시 `MOAI_HOME`, 설치된 codex 바이너리, 실제 `moai` 앞에 놓여 `mcp-server` 기동을 기록한 뒤 실제 `moai`로 넘기는 기록 래퍼, 프로젝트 층 `config.toml`에 `sandbox_mode = "workspace-write"`와 `[mcp_servers.moai]`가 있는 상태, 시험용 `plan-auditor` 역할 파일(방출본의 `developer_instructions` 끝에 실행마다 다른 nonce 한 줄을 덧붙인 것), 실행당 호출 예산 정확히 3,
**When** (a) 테스트가 launcher를 직접 불러 `plan-auditor`를 띄우고, "개발자 지시문의 nonce를 되돌리고, 탐침 파일 쓰기 명령을 한 번 실행하고, 결과를 한 줄로 반환하라"를 주며, (b) 부모 `codex exec -s workspace-write` 세션 하나를 띄워 M1에서 측정한 경로(shell 또는 MCP)로 launcher를 불러 `sync-auditor`에 같은 탐침 작업을 시키고 판정 파일을 쓰게 하면,
**Then** 증거 파일 `ac-car-010-evidence.json`에 다음이 기록되고 판정식이 모두 요구한다.

- 공통: codex 버전, 호출 수 3, `aborted` false, 경로 이름(`shell` 또는 `mcp`).
- (a): 세션 sandbox `read-only`(프로젝트 config의 `workspace-write`가 있는데도 — 플래그 우선의 측정), `spawn_agent` 사용 없음, 보낸 nonce와 되돌린 nonce 일치(지시문 전달의 측정), `probe_command_executed` true, `probe_exit_code`가 0이 아닌 수, `probe_exists` false, `write_denied` true, 시도 출력 비어 있지 않음, 감사 프로세스 동안의 MCP 기동 수 moai 0·decoy 0.
- (b): launcher가 띄운 하위 세션 sandbox `read-only`, 위와 같은 정의의 쓰기 거부, 판정 파일 존재, 반환문 sha256과 판정 파일 sha256 일치.
- 양성 대조: (b)의 부모 세션 동안 MCP 기동 수가 moai 1 이상, decoy 1 이상. 이것이 없으면 (a)의 0은 래퍼가 경로에 없어서 생긴 0과 구별되지 않는다. decoy도 기록 뒤 `moai mcp-server`로 넘기는 래퍼다. codex가 설정된 MCP 서버를 세션 시작 때 띄우는지(즉시 기동)는 측정되지 않았다. M1 P-B가 MCP를 켠 부모 세션의 래퍼 기동 수를 기록한다. 0이면(지연 기동) (b) 부모의 작업에 두 서버 각각의 읽기 전용 도구 호출 한 번(`spec_progress`)을 넣는다. 이 판단은 M1 기록(`m1-route/`)에 적고, 호출 수는 바뀌지 않는다.
- 경로 일치: 증거의 `route`가 `.moai/reports/t1143/m1-route/route.txt`의 값과 같다(`route.txt`가 단일 출처).

SKIP 의미: `MOAI_CODEX_ROLE_LIVE=1` 또는 `MOAI_T1143_EVIDENCE_DIR`이 없으면 SKIP하고 `NOT_RUN`이다. 4번째 호출이 필요해지면 시작하지 않고 `ABORTED`를 찍고 실패한다. 재실행은 리드가 `INVALID`로 기록한 실행 뒤 한 번만 허용된다(`plan.md` §D).

실행:

```bash
mkdir -p .moai/reports/t1143 && rm -f .moai/reports/t1143/ac-car-010-evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && MOAI_CODEX_ROLE_LIVE=1 MOAI_T1143_EVIDENCE_DIR=../../.moai/reports/t1143 go test -json ./internal/cli -run '^TestCodexAuditLaunchLiveContract$' -count=1 -timeout=1200s > .moai/reports/t1143/ac-car-010-live.jsonl
```

판정:

```bash
shasum -a 256 .moai/reports/t1143/ac-car-010-evidence.json > .moai/reports/t1143/ac-car-010-evidence.sha && jq -se --rawfile sha .moai/reports/t1143/ac-car-010-evidence.sha --rawfile rt .moai/reports/t1143/m1-route/route.txt --slurpfile ev .moai/reports/t1143/ac-car-010-evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ($rt|rtrimstr("\n")) as $r | ([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchLiveContract")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^ACCAR010_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^ACCAR010_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and (($e.codex_version|type)=="string" and ($e.codex_version|test("^[0-9]"))) and $e.invocations==3 and $e.aborted==false and ($e.route=="shell" or $e.route=="mcp") and $e.route==$r and $e.direct.role=="plan-auditor" and $e.direct.session_sandbox=="read-only" and $e.direct.used_spawn_agent==false and (($e.direct.nonce_sent|type)=="string" and ($e.direct.nonce_sent|length)>0) and $e.direct.nonce_returned==$e.direct.nonce_sent and $e.direct.probe_command_executed==true and (($e.direct.probe_exit_code|type)=="number" and $e.direct.probe_exit_code!=0) and $e.direct.probe_exists==false and $e.direct.write_denied==true and (($e.direct.attempt_output|type)=="string" and ($e.direct.attempt_output|length)>0) and $e.direct.mcp_launches.moai==0 and $e.direct.mcp_launches.decoy==0 and $e.routed.role=="sync-auditor" and $e.routed.child_session_sandbox=="read-only" and $e.routed.probe_command_executed==true and (($e.routed.probe_exit_code|type)=="number" and $e.routed.probe_exit_code!=0) and $e.routed.probe_exists==false and $e.routed.write_denied==true and $e.routed.verdict_file_exists==true and (($e.routed.returned_sha256//"")|test("^[0-9a-f]{64}$")) and $e.routed.returned_sha256==$e.routed.verdict_file_sha256 and (($e.routed.parent_mcp_launches.moai|type)=="number" and $e.routed.parent_mcp_launches.moai>=1) and (($e.routed.parent_mcp_launches.decoy|type)=="number" and $e.routed.parent_mcp_launches.decoy>=1)' .moai/reports/t1143/ac-car-010-live.jsonl
```

음성·변이(판정식이 `false`여야 하는 합성 입력, run 단계에서 확인): 호출 수 2 또는 4, `route` 빈 값, `route`와 `route.txt` 불일치, `route.txt` 부재, (a) sandbox `workspace-write`, nonce 불일치, `probe_command_executed` false(모델이 명령을 실행하지 않음), `probe_exit_code` 0, `probe_exists` true, `mcp_launches.moai` 1, `mcp_launches.decoy` 1, `parent_mcp_launches.moai` 0(래퍼가 경로에 없는 상황), (b) 해시 불일치, 태그 해시와 파일 해시 불일치, 테스트 skip.

### AC-CAR-011 — [LIVE] 나머지 read-only 역할의 쓰기 차단 (REQ-CAR-001, 010)

**Given** AC-CAR-010과 같은 격리 환경, 실행당 호출 예산 정확히 2,
**When** launcher로 `mission-governor`와 `super-advisor`를 한 번씩 띄워 탐침 파일 쓰기 명령을 한 번 실행하고 결과를 한 줄로 반환하게 하면,
**Then** 증거 파일 `ac-car-011-evidence.json`에 호출 수 2와, 두 역할 각각의 세션 sandbox `read-only`, `probe_command_executed` true, 0이 아닌 `probe_exit_code`, 탐침 파일 없음, `write_denied` true, 시도 출력 비어 있지 않음이 기록된다. 재실행 규칙은 AC-CAR-010과 같다.

실행:

```bash
mkdir -p .moai/reports/t1143 && rm -f .moai/reports/t1143/ac-car-011-evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && MOAI_CODEX_ROLE_LIVE=1 MOAI_T1143_EVIDENCE_DIR=../../.moai/reports/t1143 go test -json ./internal/cli -run '^TestCodexAuditLaunchLiveReadOnlyRoles$' -count=1 -timeout=900s > .moai/reports/t1143/ac-car-011-live.jsonl
```

판정:

```bash
shasum -a 256 .moai/reports/t1143/ac-car-011-evidence.json > .moai/reports/t1143/ac-car-011-evidence.sha && jq -se --rawfile sha .moai/reports/t1143/ac-car-011-evidence.sha --slurpfile ev .moai/reports/t1143/ac-car-011-evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchLiveReadOnlyRoles")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^ACCAR011_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^ACCAR011_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and $e.invocations==2 and $e.aborted==false and ([$e.roles[].role]|sort)==["mission-governor","super-advisor"] and ([$e.roles[]|select(.session_sandbox=="read-only" and .probe_command_executed==true and ((.probe_exit_code|type)=="number" and .probe_exit_code!=0) and .probe_exists==false and .write_denied==true and ((.attempt_output|type)=="string" and (.attempt_output|length)>0))]|length)==2' .moai/reports/t1143/ac-car-011-live.jsonl
```

### AC-CAR-012 — 감사가 launcher 경로로 돌았는가: 유도 규칙과 이어받은 실행 (REQ-CAR-001, 004, 010)

한 논리 AC를 두 하위 기준으로 나눈다. 둘 다 `true`여야 AC-CAR-012가 PASS다. 유도 규칙은 §A "경로 필드의 유도"이며, 한 유도 함수가 (a)와 (b)에 함께 쓰인다.

#### AC-CAR-012a — [결정적] 유도 규칙이 spawn 경로를 거부한다 (LIVE 없음)

**Given** t1100 m8-sbx 세션 기록 네 개를 그대로 복사한 테스트 fixture(`internal/cli/testdata/codex-rollouts-m8/`, M4에서 추가)와, 복사본이 원본과 같음을 보이는 sha256 목록(primary checkout `.moai/reports/t1100/m8-sbx/`에서 0.3.0 plan 때 잰 값: `…01a0ceed…` `9de04f41…afcb`, `…01a0ceee-0526…` `50a26a62…5c95`, `…01a0ceee-4848…` `c116ad0a…c4ab`, `…01a0ceee-67e8…` `6e1fa141…b7c2`), 그리고 그 기록을 편집해 만든 합성 fixture 둘 — (s1) 최상위 `source:"exec"` 감사 세션, `spawn_agent` 호출 없음, 짝이 맞는 launch record 있음, (s2) m8 run2 기록에 짝이 맞는 launch record만 덧붙인 것,
**When** 유도 함수를 각 항목에 적용하면,
**Then** 다음이 모두 성립한다.

- run2 항목(부모 `…4848…` read-only + 하위 `…67e8…` `manager-docs`): `session_sandbox == "read-only"`, `used_spawn_agent == true`, `route == "spawn_agent"`, `verdict_writer == "unattributed"`, `probe_command_executed == true`, `probe_exit_code == 1`, `write_denied == true`. 즉 쓰기가 실제로 거부된 read-only 부모 + `spawn_agent` 경로도 `route != "launcher"`로 유도된다.
- run1 항목(부모 `…ceed…` workspace-write + 하위 `…0526…` `plan-auditor`): `session_sandbox == "workspace-write"`, `used_spawn_agent == true`, `route == "spawn_agent"`, `probe_exit_code == 0`, `write_denied == false`.
- (s2): launch record가 있어도 감사 세션이 하위 세션이므로 `route == "spawn_agent"`.
- (s1): `route == "launcher"`, `verdict_writer == "launcher"`, `used_spawn_agent == false`(양성 대조 — 유도가 항상 거부만 하는 함수가 아님을 보인다).
- fixture 네 개의 sha256이 위 목록과 같다(목록과 다르면 실패).

plan 단계 참조 실행(0.3.0, `.moai/reports/t1143/plan-checks/`): 같은 규칙을 jq로 옮긴 참조 구현을 원본 네 기록에 돌린 결과, run2 하위 세션은 `top_level:false`, `spawned_role:"manager-docs"`, `sandbox:["read-only"]`, 탐침 `exit_code:1`, run2 부모는 `spawn_agent_calls:1`이었다. 이 결과가 위 기대값의 출처다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditEvidenceDerivation$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditEvidenceDerivation")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

#### AC-CAR-012b — [LIVE 증거 판정] 이어받은 실행의 유도 필드

AC-DHR-012/023과 같은 실행(같은 jsonl, 같은 증거 파일)을 판정한다. 추가 모델 호출은 없다. 이어받은 판정식이 보지 않는 필드만 읽는다. 필드는 AC-CAR-012a가 검증한 같은 유도 함수로 채우며, 테스트는 그 항목의 세션 기록과 launch record를 증거 디렉터리로 복사한다(`ac012-sessions/`, `ac012-launch-records/`).

**Given** AC-DHR-012의 실행 명령이 만든 `ac012-live.jsonl`, `ac012-evidence.json`, `ac023-evidence.json`,
**When** 두 증거 파일의 유도 필드를 읽으면,
**Then** `ac012-evidence.json`의 `write_attempts` 두 항목이 각각 `plan-auditor`, `sync-auditor`이고, 둘 다 `route == "launcher"`, `used_spawn_agent == false`, `session_sandbox == "read-only"`, `probe_command_executed == true`, 0이 아닌 숫자 `probe_exit_code`를 가지며, `ac023-evidence.json`의 `audits` 두 항목이 모두 `route == "launcher"`, `verdict_writer == "launcher"`다. 두 증거 파일 모두 태그 줄 해시와 파일 해시가 같다.

이 AC가 거부하는 실현: `-s read-only` 부모 + `spawn_agent` + 하네스의 판정 파일 쓰기(유도 결과 `route == "spawn_agent"`, `used_spawn_agent == true` — AC-CAR-012a가 m8 run2 기록으로 이 유도를 고정한다), 그리고 모델이 명령을 실행하지 않은 채 거부를 보고한 경우(`probe_command_executed` false).

```bash
mkdir -p .moai/reports/t1143 && shasum -a 256 .moai/reports/t1100/ac012-evidence.json > .moai/reports/t1143/ac-car-012-ac012.sha && shasum -a 256 .moai/reports/t1100/ac023-evidence.json > .moai/reports/t1143/ac-car-012-ac023.sha && jq -se --rawfile s12 .moai/reports/t1143/ac-car-012-ac012.sha --rawfile s23 .moai/reports/t1143/ac-car-012-ac023.sha --slurpfile e12 .moai/reports/t1100/ac012-evidence.json --slurpfile e23 .moai/reports/t1100/ac023-evidence.json '($s12|.[0:64]) as $h12 | ($s23|.[0:64]) as $h23 | ($e12[0]) as $a | ($e23[0]) as $b | ([.[]|select(.Action=="pass" and .Test=="TestCodexRoleLiveLoadAndReadOnly")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h12|test("^[0-9a-f]{64}$")) and ($h23|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^AC012_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^AC012_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h12]) and ([.[]|select((.Output//"")|test("^AC023_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^AC023_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h23]) and ([$a.write_attempts[].role]|sort)==["plan-auditor","sync-auditor"] and ([$a.write_attempts[]|select(.route=="launcher" and .used_spawn_agent==false and .session_sandbox=="read-only" and .probe_command_executed==true and ((.probe_exit_code|type)=="number" and .probe_exit_code!=0))]|length)==2 and ([$b.audits[].role]|sort)==["plan-auditor","sync-auditor"] and ([$b.audits[]|select(.route=="launcher" and .verdict_writer=="launcher")]|length)==2' .moai/reports/t1100/ac012-live.jsonl
```

음성·변이(합성 입력으로 run 단계에서 확인): 한 항목의 `route` = `spawn_agent`, `used_spawn_agent` true, `session_sandbox` = `workspace-write`, `probe_command_executed` false, `probe_exit_code` 0 또는 문자열, `verdict_writer` = `unattributed`, 같은 역할 두 번, 필드 누락(구 버전 증거), 태그 해시 불일치.

### AC-CAR-013 — R2 조건부 결정적 판정 (REQ-CAR-005, 011)

**Given** M1이 기록한 경로 파일 `.moai/reports/t1143/m1-route/route.txt`(내용은 정확히 `shell`, `mcp`, `none` 중 하나),
**When** 경로에 맞는 결정적 테스트를 실행하면,
**Then** `mcp`이면 `TestCodexAuditMCPTool`이 통과해야 한다. 이 테스트는 MCP 도구가 등록되어 있고, 작업 루트 입력에 AC-CAR-005와 같은 한정(같은 저장소의 등록된 워크트리, `.moai/reports/` 아래 목적지, `.git` 성분 금지)을 적용하며, 긴 감사를 위한 비동기 작업 형태(시작·상태·결과)를 갖고, `.claude/rules/moai/core/moai-mcp-tools.md`와 템플릿 미러의 도구 수가 실제 등록 수와 같음을 확인한다. `shell`이면 `TestCodexAuditMCPToolAbsent`가 통과해야 한다. 이 테스트는 그 MCP 도구가 등록되어 있지 않고 도구 카탈로그 수가 바뀌지 않았음을 확인한다. `none`이거나 파일이 없거나 내용이 셋 중 하나가 아니면 FAIL이다(`none`은 B4에 따라 카드를 멈춘 상태다).

```bash
if grep -qx 'mcp' .moai/reports/t1143/m1-route/route.txt 2>/dev/null; then unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditMCPTool$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditMCPTool")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'; elif grep -qx 'shell' .moai/reports/t1143/m1-route/route.txt 2>/dev/null; then unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditMCPToolAbsent$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditMCPToolAbsent")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'; else echo false; fi
```

### AC-CAR-014 — launch record의 내용 (REQ-CAR-007)

**Given** 가짜 `codex`의 성공 실행 하나와 실패 실행 하나(0이 아닌 종료), 그리고 프로세스를 띄우기 전에 거부되는 호출 셋 — 목적지 검증 실패(REQ-CAR-005), 자격 없는 역할(REQ-CAR-002), 인자 상한 초과(REQ-CAR-003),
**When** launcher로 각각 실행하면,
**Then** 두 실행 모두 워크트리 뿌리의 `.moai/reports/codex-audit/` 아래에 launch record가 하나씩 생기고, launcher가 표준 오류에 `LAUNCH_RECORD <상대 경로>` 한 줄을 찍으며, 그 파일이 아래 스키마(버전 1)를 따른다.

- `schema_version` 1, `role`, `route`(`direct`·`shell`·`mcp` 중 하나), `started_at`·`ended_at`(RFC 3339 UTC).
- `argv`: 실제로 넘긴 인자 배열. 단 `developer_instructions` 값은 원문 대신 `<redacted sha256=<64 hex> bytes=<N>>`으로 적는다(원문이 파일에 없어야 한다).
- `sandbox` = `"read-only"`, `mcp_servers` = `"disabled"`.
- `covers`: read-only 보장이 Codex sandbox가 다스리는 모델 발행 명령과 편집에 한정된다는 비어 있지 않은 문장.
- `unsupported`: 정확히 `["codex-home-session-files", "project-hook-commands"]`(순서 무관, 더도 덜도 없음).
- `exit_code`, `failure_reason`(성공 시 null, 실패 시 비어 있지 않은 문자열), `verdict_path`(성공하고 목적지가 있으면 워크트리 기준 상대 경로, 아니면 null), `verdict_sha256`(판정 파일 sha256 또는 null).

실패 실행의 record는 `exit_code`가 0이 아니고 `verdict_path`가 null이다. 성공 실행의 `verdict_sha256`은 판정 파일의 실제 sha256과 같다. 파일 이름은 launcher가 만들며, 같은 이름의 파일이 이미 있으면 덮어쓰지 않고 실패한다(배타 생성). 거부된 세 호출은 `.moai/reports/codex-audit/`을 포함해 어떤 파일도 만들지 않고, `LAUNCH_RECORD` 줄 없이 표준 오류에만 거부 사유를 찍는다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchRecord$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchRecord")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

## §D 설계 기준별 판정 집계

| 묶음 | 결정적 AC | LIVE AC | PASS 조건 |
|---|---|---|---|
| launcher 동작 | AC-CAR-001 ~ 006, 014, 012a | AC-CAR-010, 011 | 결정적 여덟 `true` + LIVE 둘 `true` |
| 이어받은 항목 | AC-CAR-009 | AC-DHR-012, AC-DHR-023, AC-CAR-012b | 넷 모두 `true`. AC-DHR-012/023만 `true`이고 AC-CAR-012b가 `false`이면 경로 (i)가 증명되지 않은 것이며 PASS 아님. LIVE가 `NOT_RUN`·`ABORTED`·`INVALID`이면 `PARTIAL`이며 PASS 아님 |
| 지시면과 경계 | AC-CAR-007, 008, 013 | 없음 | 셋 모두 `true` |

`UNSUPPORTED`로 선언한 쓰기 주체(REQ-CAR-007: `CODEX_HOME` 세션 기록, hook 명령)는 PASS 집계에 넣지 않는다. 그 선언이 실제로 launch record에 들어가는지는 AC-CAR-014가 판정한다.

## §E 완료 정의

- §D의 세 묶음이 모두 PASS다. LIVE 하나라도 `NOT_RUN`·`ABORTED`·`INVALID`로 끝나면 카드는 완료가 아니다.
- LIVE 호출 원장(`plan.md` §D의 항목별 사용 수, 상한, 경과, `ABORTED`·`INVALID` 여부)이 `progress.md` §E.2와 `.moai/reports/t1143/verdict.md`에 있다.
- 판정서 `.moai/reports/t1143/verdict.md`가 M1 경로 측정 결과(`route.txt`의 값과 근거)를 담는다.
- 출처 SPEC(SPEC-DUAL-HARNESS-RECOVERY-001)의 파일은 이 카드에서 바꾸지 않는다. 이관 항목의 충족 기록은 이 SPEC에만 남는다.
