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
- `SKIP`, `NOT_RUN`, `ABORTED`는 PASS가 아니다. 패키지 단위 `ok` 줄은 증거로 쓰지 않는다.
- 테스트 이름은 run 단계에서 만들 이름이다. 이름을 바꾸면 이 파일의 명령도 함께 고친다. 이름이 없으면 pass 수가 0이 되어 FAIL이다.
- `internal/cli` 명령은 kanban·factory 환경 변수를 같은 호출 안에서 지운다(`unset ... && go test ...`).
- 명령에는 실행 중에 계산한 값을 git·go 명령으로 넘기는 형태를 쓰지 않는다(워크트리 세션 가드가 거부한다). 비교 기준은 리터럴 `develop...HEAD`(merge-base 기준 3점 표기)와 이관 커밋 `de5faa77a`다.
- 이 SPEC의 새 증거는 `.moai/reports/t1143/`에 남긴다. 이어받은 두 AC(§B)는 원문 명령 그대로 `.moai/reports/t1100/`에 쓴다. 이 경로는 t1143 워크트리 안의 디렉터리이며 primary checkout에 반출된 t1100 증거와 같은 파일이 아니다.
- LIVE AC(AC-DHR-012, AC-DHR-023, AC-CAR-010, AC-CAR-011)는 결정적 AC와 따로 집계한다. 호출 수의 단위는 `codex exec` 프로세스 하나다. 증거 JSON은 파일로 남기고 표준 출력에는 `<TAG>_SHA256 <64자 hex>` 한 줄만 찍는다(출처 SPEC `acceptance.md` §A 증거 채널과 같은 방식; `go test -json`의 1024바이트 분할 회피). 판정은 파일 해시를 다시 재어 태그 줄과 같을 때만 내용을 본다.
- LIVE 예산은 `plan.md` §D가 정한다. 예산을 넘는 호출은 시작하지 않고 `ABORTED`를 찍고 실패한다(REQ-CAR-010).

공통 판정식(각 결정적 명령에 그대로 들어 있다): 지정 이름의 `pass` 이벤트 수가 1이고, `fail`·`skip` 이벤트가 0이며, 출력에 `NOT_RUN`·`ABORTED`가 없다.

### AC ↔ 요구사항 매핑

| AC | 요구사항 | 종류 |
|---|---|---|
| AC-DHR-012 (이어받음) | REQ-DHR-014(출처 SPEC), REQ-CAR-001, REQ-CAR-010 | LIVE, 14회 |
| AC-DHR-023 (이어받음) | REQ-DHR-015 런타임 조항, REQ-CAR-004 | LIVE, 추가 호출 0 |
| AC-CAR-001 | REQ-CAR-001, 003, 007 | 결정적 |
| AC-CAR-002 | REQ-CAR-002 | 결정적 |
| AC-CAR-003 | REQ-CAR-004, REQ-DHR-015 | 결정적 |
| AC-CAR-004 | REQ-CAR-006 | 결정적 |
| AC-CAR-005 | REQ-CAR-005 | 결정적 |
| AC-CAR-006 | REQ-CAR-003 | 결정적 |
| AC-CAR-007 | REQ-CAR-008, 011 | 결정적 |
| AC-CAR-008 | REQ-CAR-009 | 결정적 |
| AC-CAR-009 | REQ-DHR-015 원문 보존, §B 이관 | 결정적 |
| AC-CAR-010 | REQ-CAR-001, 003, 007, 011, REQ-DHR-015 | LIVE, 정확히 3회 |
| AC-CAR-011 | REQ-CAR-001, 010 | LIVE, 정확히 2회 |

## §B 이어받은 AC (SPEC-DUAL-HARNESS-RECOVERY-001 0.3.1에서 이관)

출처: `.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md` 197-226행(AC-DHR-012), 404-422행(AC-DHR-023), 이관 커밋 `de5faa77a`, SPEC 버전 0.3.1. 표시 사이의 줄은 출처와 한 글자도 다르지 않다(AC-CAR-009). 본문 안의 "이관" 주석은 t1100의 관측 기록이며 역시 원문이다. 판정식, 기대값, 실행 명령은 바꾸지 않는다.

**경로 (i)에서의 실현 방식 (이 SPEC의 해석, 판정식은 그대로).** 이어받은 판정식은 필드 단위로 판정하며, 아래 방식으로 경로 (i)에서 충족할 수 있다.

- (i) 12개 역할 로드는 원문대로 부모 `codex exec`가 `spawn_agent`로 각 역할을 한 번씩 띄운다. 쓰기가 없는 로드 탐침이므로 감사 역할도 여기서는 하위 에이전트로 로드된다. 이것은 역할 파일 로드의 측정이며 감사 실행 경로가 아니다.
- (ii) 두 감사 역할의 쓰기 시도는 audit launcher가 최상위 `codex exec -s read-only`로 띄운 프로세스에서 일어난다. 이 실행에서 부모 lane 오케스트레이터의 몫(감사 역할을 띄우고 반환문으로 판정 파일을 쓰는 일)은 테스트가 직접 부르는 launcher가 맡는다. 그래서 (ii)는 역할당 `codex exec` 프로세스 하나이며, 호출 수는 원문 판정식이 요구하는 14(12 + 2)로 남는다. 부모 Codex 세션이 launcher를 부르는 경로 자체는 AC-CAR-010이 따로 잰다.
- 원문 (ii)의 "하위 에이전트" 문구와 이 실현 방식의 차이는 운영자 확인 항목이다(`plan.md` §G B1). 확인 전까지 이 해석은 제안이며, 판정식과 기대값은 어느 쪽이든 바뀌지 않는다.

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
**Then** 가짜 `codex`가 정확히 한 번 호출되고, 인자에 `exec`, sandbox `read-only`(플래그), `approval_policy="never"`, 호출자 워크트리 뿌리를 가리키는 작업 루트, 역할 파일과 같은 `model_reasoning_effort` 값, 역할 파일의 `developer_instructions`와 바이트가 같은 값, MCP 서버 비활성화 설정이 있으며, `workspace-write`, `danger-full-access`, `--dangerously-bypass-approvals-and-sandbox`, `--dangerously-bypass-hook-trust`, `spawn_agent`가 없다.

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
**When** launcher가 그 결과를 목적지에 쓰면,
**Then** 목적지 파일의 바이트가 가짜 `codex`가 낸 최종 메시지 바이트와 같다(sha256 비교, 정규화 없음). 이미 파일이 있던 목적지에서도 결과는 완전한 새 파일이며, 쓰기 도중 중단을 흉내 낸 경우 목적지는 이전 파일 그대로다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchVerbatimWrite$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchVerbatimWrite")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-004 — 실패한 감사는 아무것도 쓰지 않는다 (REQ-CAR-006)

**Given** 세 가지 가짜 `codex` — 0이 아닌 종료, 시간 한도 초과, 빈 최종 메시지 — 와 (a) 목적지가 없는 경우, (b) 목적지에 이전 파일이 있는 경우,
**When** launcher를 실행하면,
**Then** 여섯 조합 모두에서 launcher가 0이 아닌 코드로 끝나고, 진단에 역할 이름과 실패 사유가 있으며, (a)에서는 목적지가 생기지 않고 (b)에서는 목적지 sha256이 실행 전과 같다. 시간 한도 초과 조합에서는 가짜 `codex` 프로세스가 테스트 종료 전에 정리된다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchFailureWritesNothing$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchFailureWritesNothing")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-005 — 목적지는 호출자가 정하고 워크트리 안이어야 한다 (REQ-CAR-005)

**Given** 가짜 `codex`가 최종 메시지 안에 다른 경로(워크트리 밖 절대 경로)를 적어 내고, 목적지 인자 후보로 워크트리 안 경로, `..`로 밖을 가리키는 경로, 워크트리 밖 절대 경로, 워크트리 밖을 가리키는 심볼릭 링크 경로가 있을 때,
**When** 각 후보로 launcher를 실행하면,
**Then** 워크트리 안 경로만 쓰기가 일어나고 나머지 셋은 0이 아닌 코드로 끝나며 어떤 파일도 만들거나 바꾸지 않는다. 최종 메시지 안의 경로에는 어떤 경우에도 파일이 생기지 않는다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchDestinationConfinement$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchDestinationConfinement")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-006 — 인자 상한 초과 시 실행 전 실패 (REQ-CAR-003)

**Given** 기존 `developer_instructions` 인자 상한(`internal/config/defaults.go`, SPEC-CODEX-LOCALMD-001)보다 1바이트 긴 지시문을 가진 시험용 역할 파일과, 상한과 정확히 같은 길이의 역할 파일,
**When** 각각 launcher로 실행하면,
**Then** 긴 쪽은 가짜 `codex` 호출 없이 0이 아닌 코드로 끝나고 진단에 측정 길이와 상한이 있으며, 같은 길이 쪽은 가짜 `codex`가 한 번 호출된다.

```bash
unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && go test -json ./internal/cli -run '^TestCodexAuditLaunchInstructionCeiling$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchInstructionCeiling")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-007 — 지시면이 launcher를 가리킨다 (REQ-CAR-008, 011)

**Given** 템플릿 `AGENTS.md.tmpl`과 방출기 매니페스트의 Codex 전용 부록,
**When** 재생성 검사와 지시면 테스트를 실행하면,
**Then** 커밋된 역할 TOML이 방출 결과와 같고(`make agents-emit-check`), `audit-verdict-file` 행과 read-only 계약 역할 넷의 부록이 모두 launcher 호출 형태(M1에서 측정한 경로 하나만)를 담으며, 감사 역할을 `spawn_agent`로 띄우라는 문구나 "하위 에이전트로 실행된다"는 문구가 없고, 판정 파일을 launcher가 반환문으로 쓴다는 문구가 있다. 측정되지 않은 경로 이름(shell 경로와 MCP 경로 중 채택하지 않은 쪽)이 지시면에 없다.

```bash
make -s agents-emit-check >/dev/null 2>&1 && go test -json ./internal/template/agentemit -run '^TestAuditRoleLauncherInstructionSurface$' -count=1 | jq -se '([.[]|select(.Action=="pass" and .Test=="TestAuditRoleLauncherInstructionSurface")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0'
```

### AC-CAR-008 — Claude 쪽 무변경과 템플릿 중립성 (REQ-CAR-009)

**Given** 이 카드의 브랜치,
**When** develop과의 merge-base 기준 변경을 보면,
**Then** Claude 에이전트 정의 두 벌과 Claude plan·sync 흐름 파일에 변경이 없고, 템플릿 트리에 추가된 줄에 SPEC ID, 카드 번호, 날짜가 없다.

```bash
git diff --quiet develop...HEAD -- .claude/agents/moai internal/template/templates/.claude/agents/moai .claude/skills/moai/workflows/plan.md .claude/skills/moai/workflows/sync.md internal/template/templates/.claude/skills/moai/workflows/plan.md internal/template/templates/.claude/skills/moai/workflows/sync.md && ! git diff develop...HEAD -- internal/template/templates | grep -E '^\+[^+]' | grep -Eq 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}|\bt[0-9]{3,4}\b|20[0-9]{2}-[0-9]{2}-[0-9]{2}' && echo true
```

### AC-CAR-009 — 이어받은 원문의 바이트 보존 (§B, REQ-DHR-015)

**Given** 이관 커밋 `de5faa77a`의 출처 파일과 이 SPEC의 두 파일,
**When** 표시(`<!-- inherited:begin … -->` / `<!-- inherited:end … -->`) 사이 줄을 뽑아 출처의 해당 행과 비교하면,
**Then** 세 블록(REQ-DHR-015, AC-DHR-012, AC-DHR-023)이 모두 바이트 단위로 같다.

```bash
mkdir -p .moai/reports/t1143 && git show de5faa77a:.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/spec.md | sed -n '131,133p' > .moai/reports/t1143/inh-req015.src && git show de5faa77a:.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md | sed -n '197,226p' > .moai/reports/t1143/inh-ac012.src && git show de5faa77a:.moai/specs/SPEC-DUAL-HARNESS-RECOVERY-001/acceptance.md | sed -n '404,422p' > .moai/reports/t1143/inh-ac023.src && awk '/^<!-- inherited:end REQ-DHR-015 -->$/{f=0} f; /^<!-- inherited:begin REQ-DHR-015 -->$/{f=1}' .moai/specs/SPEC-CODEX-AUDIT-READONLY-001/spec.md > .moai/reports/t1143/inh-req015.dst && awk '/^<!-- inherited:end AC-DHR-012 -->$/{f=0} f; /^<!-- inherited:begin AC-DHR-012 -->$/{f=1}' .moai/specs/SPEC-CODEX-AUDIT-READONLY-001/acceptance.md > .moai/reports/t1143/inh-ac012.dst && awk '/^<!-- inherited:end AC-DHR-023 -->$/{f=0} f; /^<!-- inherited:begin AC-DHR-023 -->$/{f=1}' .moai/specs/SPEC-CODEX-AUDIT-READONLY-001/acceptance.md > .moai/reports/t1143/inh-ac023.dst && test -s .moai/reports/t1143/inh-ac012.src && cmp -s .moai/reports/t1143/inh-req015.src .moai/reports/t1143/inh-req015.dst && cmp -s .moai/reports/t1143/inh-ac012.src .moai/reports/t1143/inh-ac012.dst && cmp -s .moai/reports/t1143/inh-ac023.src .moai/reports/t1143/inh-ac023.dst && echo true
```

### AC-CAR-010 — [LIVE] launcher 계약과 실행 경로 (REQ-CAR-001, 003, 007, 011, REQ-DHR-015)

**Given** 격리된 임시 저장소, 임시 `CODEX_HOME`(로그인 사본), 임시 `MOAI_HOME`, 설치된 codex 바이너리, 실제 `moai` 앞에 놓여 `mcp-server` 기동을 기록하는 기록용 래퍼, 프로젝트 config에 `sandbox_mode = "workspace-write"`와 `[mcp_servers.moai]`가 있는 상태, 시험용 `plan-auditor` 역할 파일(방출본의 `developer_instructions` 끝에 실행마다 다른 nonce 한 줄을 덧붙인 것), 호출 예산 정확히 3,
**When** (a) 테스트가 launcher를 직접 불러 `plan-auditor`를 띄우고, 감사 작업으로 "개발자 지시문의 nonce를 되돌리고, 탐침 파일 쓰기를 한 번 시도하고, 결과를 한 줄로 반환하라"를 주며, (b) 부모 `codex exec -s workspace-write` 세션 하나를 띄워 M1에서 측정한 경로(shell 또는 MCP)로 launcher를 불러 `sync-auditor`에 같은 탐침 작업을 시키고 판정 파일을 쓰게 하면,
**Then** 증거 파일 `ac-car-010-evidence.json`에 codex 버전, 호출 수 3, 경로 이름, (a)의 세션 기록 sandbox `read-only`·`spawn_agent` 사용 없음·보낸 nonce와 되돌린 nonce 일치·쓰기 거부(출력 비어 있지 않음, 탐침 파일 없음)·`mcp-server` 기동 0회, (b)의 하위(launcher가 띄운) 세션 sandbox `read-only`·쓰기 거부·탐침 파일 없음·판정 파일 존재·반환문 sha256과 판정 파일 sha256 일치가 기록된다. 프로젝트 config의 `workspace-write`가 있는데도 (a)가 `read-only`라는 것이 플래그 우선의 측정이다.

SKIP 의미: `MOAI_CODEX_ROLE_LIVE=1` 또는 `MOAI_T1143_EVIDENCE_DIR`이 없으면 SKIP하고 `NOT_RUN`이다. 4번째 호출이 필요해지면 시작하지 않고 `ABORTED`를 찍고 실패한다.

실행:

```bash
mkdir -p .moai/reports/t1143 && rm -f .moai/reports/t1143/ac-car-010-evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && MOAI_CODEX_ROLE_LIVE=1 MOAI_T1143_EVIDENCE_DIR=../../.moai/reports/t1143 go test -json ./internal/cli -run '^TestCodexAuditLaunchLiveContract$' -count=1 -timeout=1200s > .moai/reports/t1143/ac-car-010-live.jsonl
```

판정:

```bash
shasum -a 256 .moai/reports/t1143/ac-car-010-evidence.json > .moai/reports/t1143/ac-car-010-evidence.sha && jq -se --rawfile sha .moai/reports/t1143/ac-car-010-evidence.sha --slurpfile ev .moai/reports/t1143/ac-car-010-evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchLiveContract")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^ACCAR010_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^ACCAR010_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and (($e.codex_version|type)=="string" and ($e.codex_version|test("^[0-9]"))) and $e.invocations==3 and $e.aborted==false and ($e.route=="shell" or $e.route=="mcp") and $e.direct.role=="plan-auditor" and $e.direct.session_sandbox=="read-only" and $e.direct.used_spawn_agent==false and (($e.direct.nonce_sent|type)=="string" and ($e.direct.nonce_sent|length)>0) and $e.direct.nonce_returned==$e.direct.nonce_sent and $e.direct.write_denied==true and $e.direct.probe_exists==false and (($e.direct.attempt_output|type)=="string" and ($e.direct.attempt_output|length)>0) and $e.direct.mcp_server_launches==0 and $e.routed.role=="sync-auditor" and $e.routed.child_session_sandbox=="read-only" and $e.routed.write_denied==true and $e.routed.probe_exists==false and $e.routed.verdict_file_exists==true and (($e.routed.returned_sha256//"")|test("^[0-9a-f]{64}$")) and $e.routed.returned_sha256==$e.routed.verdict_file_sha256' .moai/reports/t1143/ac-car-010-live.jsonl
```

음성·변이(판정식이 `false`여야 하는 합성 입력, run 단계에서 확인): 호출 수 2 또는 4, `route` 빈 값, (a) sandbox `workspace-write`, nonce 불일치, `probe_exists` true, `mcp_server_launches` 1, (b) 해시 불일치, 태그 해시와 파일 해시 불일치, 테스트 skip.

### AC-CAR-011 — [LIVE] 나머지 read-only 역할의 쓰기 차단 (REQ-CAR-001, 010)

**Given** AC-CAR-010과 같은 격리 환경, 호출 예산 정확히 2,
**When** launcher로 `mission-governor`와 `super-advisor`를 한 번씩 띄워 탐침 파일 쓰기를 한 번 시도하고 결과를 한 줄로 반환하게 하면,
**Then** 증거 파일 `ac-car-011-evidence.json`에 호출 수 2와, 두 역할 각각의 세션 sandbox `read-only`, 쓰기 거부(출력 비어 있지 않음), 탐침 파일 없음이 기록된다.

실행:

```bash
mkdir -p .moai/reports/t1143 && rm -f .moai/reports/t1143/ac-car-011-evidence.json && unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS && MOAI_CODEX_ROLE_LIVE=1 MOAI_T1143_EVIDENCE_DIR=../../.moai/reports/t1143 go test -json ./internal/cli -run '^TestCodexAuditLaunchLiveReadOnlyRoles$' -count=1 -timeout=900s > .moai/reports/t1143/ac-car-011-live.jsonl
```

판정:

```bash
shasum -a 256 .moai/reports/t1143/ac-car-011-evidence.json > .moai/reports/t1143/ac-car-011-evidence.sha && jq -se --rawfile sha .moai/reports/t1143/ac-car-011-evidence.sha --slurpfile ev .moai/reports/t1143/ac-car-011-evidence.json '($sha|.[0:64]) as $h | ($ev[0]) as $e | ([.[]|select(.Action=="pass" and .Test=="TestCodexAuditLaunchLiveReadOnlyRoles")]|length)==1 and ([.[]|select(.Action=="fail" or .Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|test("NOT_RUN|ABORTED"))]|length)==0 and ($h|test("^[0-9a-f]{64}$")) and ([.[]|select((.Output//"")|test("^ACCAR011_EVIDENCE_SHA256 [0-9a-f]{64}\n?$"))|.Output|capture("^ACCAR011_EVIDENCE_SHA256 (?<h>[0-9a-f]{64})").h]==[$h]) and ($ev|length)==1 and $e.invocations==2 and $e.aborted==false and ([$e.roles[].role]|sort)==["mission-governor","super-advisor"] and ([$e.roles[]|select(.session_sandbox=="read-only" and .write_denied==true and .probe_exists==false and ((.attempt_output|type)=="string" and (.attempt_output|length)>0))]|length)==2' .moai/reports/t1143/ac-car-011-live.jsonl
```

## §D 설계 기준별 판정 집계

| 묶음 | 결정적 AC | LIVE AC | PASS 조건 |
|---|---|---|---|
| launcher 동작 | AC-CAR-001 ~ 006 | AC-CAR-010, 011 | 결정적 여섯 `true` + LIVE 둘 `true` |
| 이어받은 항목 | AC-CAR-009 | AC-DHR-012, AC-DHR-023 | 셋 모두 `true`. LIVE가 `NOT_RUN`·`ABORTED`이면 `PARTIAL`이며 PASS 아님 |
| 지시면과 경계 | AC-CAR-007, 008 | 없음 | 둘 모두 `true` |

`UNSUPPORTED`로 선언한 쓰기 주체(REQ-CAR-007: `CODEX_HOME` 세션 기록, hook 명령)는 PASS 집계에 넣지 않고 판정서에 따로 나열한다.

## §E 완료 정의

- §D의 세 묶음이 모두 PASS다. LIVE 하나라도 `NOT_RUN`·`ABORTED`이면 카드는 완료가 아니다.
- LIVE 호출 원장(`plan.md` §D의 항목별 사용 수, 상한, 경과, `ABORTED` 여부)이 `progress.md` §E.2와 `.moai/reports/t1143/verdict.md`에 있다.
- 판정서 `.moai/reports/t1143/verdict.md`가 REQ-CAR-007의 `UNSUPPORTED` 목록과 M1 경로 측정 결과를 담는다.
- 출처 SPEC(SPEC-DUAL-HARNESS-RECOVERY-001)의 파일은 이 카드에서 바꾸지 않는다. 이관 항목의 충족 기록은 이 SPEC에만 남는다.
