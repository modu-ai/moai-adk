---
id: SPEC-CODEX-AUDIT-READONLY-001
document: plan
created: 2026-09-24
updated: 2026-09-24
author: manager-spec
card: t1143
---

# Plan — SPEC-CODEX-AUDIT-READONLY-001

## §A 맥락과 측정 근거

plan 단계에서 읽기 전용으로 확인한 사실이다. 모델 호출은 하지 않았다(`codex --version`, `codex exec --help`, 바이너리 문자열 검색만).

| 사실 | 근거 |
|---|---|
| 설치된 codex는 `codex-cli 0.156.1` | `codex --version` |
| `-s`는 "모델이 낸 shell 명령을 실행할 때의 sandbox 정책"이며 값은 `read-only`, `workspace-write`, `danger-full-access` | `codex exec --help` |
| `codex exec`에는 `-o/--output-last-message <FILE>`, `--json`(JSONL 이벤트), `--ephemeral`, `--ignore-user-config`, `-C/--cd`, `-c key=value`가 있다 | `codex exec --help` |
| 하위 에이전트는 부모 sandbox를 물려받고, 최상위 `-s read-only` 세션은 부모 자신의 쓰기와 하위 에이전트의 쓰기를 모두 막았다 | `.moai/reports/t1100/m8-sbx/summary.json` run1·run2, `run1-argv.txt`, `run2-argv.txt` |
| run2의 프로젝트 config에는 `sandbox_mode` 키가 없었다. 따라서 "`-s` 플래그가 config의 `sandbox_mode`를 이긴다"는 아직 측정되지 않았다 | `summary.json` `.runs.run2.config_toml` |
| `workspace-write` 세션의 `sandbox_policy`에 `network_access: false`가 있었다 | `summary.json` run1 `turn_contexts[0].sandbox_policy` |
| 방출된 read-only 계약 역할은 넷: `mission-governor`, `plan-auditor`, `super-advisor`, `sync-auditor`. 넷 모두 C1(`.claude/agents/moai/*.md`), C2(템플릿 미러), C3(`.codex/agents/moai/*.toml`)에 있다 | `ls` 세 디렉터리, TOML `sandbox_mode` 판독 |
| 역할 TOML 크기: `plan-auditor` `developer_instructions` 61,688바이트(`developer_instructions=<JSON>` 인자로 62,843바이트), `sync-auditor` 18,505, `super-advisor` 7,596, `mission-governor` 1,248. effort는 넷 모두 `high` | TOML 파싱 + JSON 인코딩 길이 측정(python `tomllib`/`json`) |
| 기존 launcher는 로컬 지시 파일 두 개를 합쳐 `-c developer_instructions=<JSON>` 하나로 주입하며, 126,976바이트 상한으로 실패 닫힘을 한다 | `internal/cli/codex_launcher.go:118-143`, SPEC-CODEX-LOCALMD-001 REQ(상한 126,976) |
| 이 워크트리의 `CLAUDE.local.md`는 61,360바이트다. `plan-auditor` 지시문(62,843)과 합치면 인코딩 전에 이미 124,203바이트로 상한에 거의 닿는다 | `wc -c` |
| 프로젝트 `.codex/config.toml`에는 `[mcp_servers.moai]`와 `default_tools_approval_mode = "writes"`가 있다 | `.codex/config.toml`, `internal/codexwiring/configtoml.go:13` |
| 기존 `moai codex` 동사는 `cli`, `status`, `app`뿐이다 | `internal/cli/codex_launcher.go:629` |
| 이어받은 LIVE 테스트는 부모를 `-s workspace-write`로 띄우고 감사 역할을 `spawn_agent`로 부른다. 예산 상수 14 | `internal/cli/codex_role_live_test.go:23,127-130,191-196` |
| 바이너리에 `developer_instructions`(64회), `enabled_tools`, `disabled_tools` 문자열이 있다. 문자열이 있다는 것은 의미를 증명하지 않는다 | `strings -n 12` 결과 grep |

## §B 결정 (바뀔 가능성이 큰 것부터)

### D1. 실행 경로 — 측정으로 정한다 (M1)

부모 Codex lane 세션이 launcher를 부를 수 있는 경로는 둘이다.

- **R1 shell 경로**: 부모 세션이 shell 도구로 `moai codex audit ...`을 실행한다. launcher와 그 아래 `codex exec`는 부모의 sandbox(`workspace-write`, `network_access: false`) 안에서 뜬다. 중첩된 `codex exec`가 모델 API에 닿는지, macOS sandbox 안에서 다시 sandbox를 걸 수 있는지는 측정되지 않았다. 네트워크 차단이 측정된 설정이므로 실패 가능성이 있다.
- **R2 MCP 경로**: moai MCP 서버에 도구 하나를 더하고 서버가 launcher를 부른다. MCP 서버는 Codex 호스트가 띄우는 프로세스이며 shell sandbox 밖에서 돈다는 것이 통상의 이해지만 이 트리에서 측정하지 않았다. 서버의 작업 디렉터리는 primary checkout이므로 워크트리 뿌리를 입력으로 받아야 한다(`moai-mcp-tools.md` § project_root). 감사는 수 분이 걸리므로 `codex_task`와 같은 비동기 작업 형태가 필요할 수 있다.

M1이 R1을 먼저 재고, R1이 모델에 닿지 못할 때만 R2를 잰다. 지시면에는 측정으로 동작한 경로 하나만 적는다(REQ-CAR-011). launcher의 핵심 동작(인자 조립, 실행, 원문 쓰기, 목적지 검증)은 경로와 무관한 하나의 Go 함수로 두고, 경로는 그것을 부르는 얇은 겉면만 다르게 한다.

[NEEDS CLARIFICATION: R1과 R2가 모두 측정에서 실패하면(중첩 실행 불가, MCP 서버도 모델에 닿지 못함) 이 카드의 범위를 "lead/foreman 쪽 실행(R3, Codex 세션 밖)"으로 넓힐지, 카드를 멈추고 보고할지]

### D2. 실행 인자 계약

```
codex exec -s read-only -c approval_policy="never" -c model_reasoning_effort="<역할 값>" -c developer_instructions=<역할 지시문 JSON> -c <MCP 비활성화> -C <워크트리 뿌리> --json [-o <OS 임시 경로>] -   (작업 문은 stdin)
```

- sandbox는 반드시 `-s read-only` 플래그로 준다. config 값에 기대지 않는다. 플래그가 config를 이기는지는 M1과 AC-CAR-010이 잰다.
- `approval_policy="never"`: 거부된 쓰기가 승인 요청으로 멈추지 않고 실패로 끝나게 한다(t1100 M8과 같은 조건).
- `--skip-git-repo-check`는 쓰지 않는다. 작업 루트는 git 워크트리다.
- 우회 옵션(`--dangerously-bypass-approvals-and-sandbox`, `--dangerously-bypass-hook-trust`, `--approve-for-me`, `--add-dir`)은 어떤 경우에도 넣지 않는다(REQ-CAR-007, AC-CAR-001).
- 작업 문(예: 어느 SPEC을 감사하라)은 부모가 stdin으로 준다. 인자 길이 상한을 역할 지시문이 거의 다 쓰기 때문이다.
- 시간 한도는 `internal/config/defaults.go`의 이름 있는 상수 하나로 둔다. 초과 시 프로세스 그룹을 정리한다.

### D3. 반환문 채널

1순위는 `--json` 이벤트 스트림의 마지막 에이전트 메시지다. 파일 시스템에 기대지 않는다. `-o` 파일은 codex 호스트 프로세스가 쓰므로 sandbox 대상이 아닐 것으로 보지만 측정되지 않았다. M1이 두 채널이 같은 바이트를 내는지 잰다. 다르면 `--json`을 채택하고 차이를 판정서에 적는다. 출처 판정식(AC-DHR-023)은 반환문과 판정 파일의 끝 공백을 잘라 해시를 비교하지만, launcher 자체는 정규화 없이 바이트 그대로 쓴다(AC-CAR-003). 출처 판정식이 더 느슨하므로 둘은 충돌하지 않는다.

### D4. 역할 지시문 전달과 인자 상한

- 역할 지시문은 방출된 역할 TOML의 `developer_instructions`를 그대로 `-c developer_instructions=`로 준다. 효과는 TOML의 `model_reasoning_effort`를 그대로 준다. 역할 TOML 자체를 최상위 세션에 "로드"하는 방법은 codex에 없다(`codex exec --help`에 역할 선택 옵션 없음).
- **로컬 지시 파일(CLAUDE.local.md 등)은 감사 프로세스에 합치지 않는다.** 측정한 길이로 `plan-auditor` 62,843 + `CLAUDE.local.md` 61,360 = 124,203바이트이며 JSON 인코딩(Go `json.Marshal`은 `<`, `>`, `&`도 이스케이프한다) 뒤에는 126,976 상한을 넘을 수 있다. 감사자는 프로젝트 `AGENTS.md`를 codex의 파일 탐색으로 읽는다. 로컬 지시를 합치지 않는 것이 감사 품질에 영향을 주는지는 [NEEDS CLARIFICATION: 감사 프로세스에 로컬 지시 파일을 빼는 것을 운영자가 받아들이는지]
- 상한은 SPEC-CODEX-LOCALMD-001의 기존 상수를 재사용한다. 새 상수를 만들지 않는다.
- `-c developer_instructions`가 최상위 세션에서 실제로 개발자 지시로 들어가는지는 SPEC-CODEX-LOCALMD-001의 LIVE AC가 다루는 영역이지만 이 plan에서 그 결과를 확인하지 않았다. M1이 nonce로 직접 잰다(시험용 역할 지시문 끝에 nonce를 넣고 되돌리게 한다).

### D5. sandbox 밖 쓰기 주체

`-s`가 다스리는 것은 모델이 낸 명령과 편집이다. 그 밖의 쓰기 주체는 다음과 같다.

| 주체 | 처리 | 근거 |
|---|---|---|
| moai MCP 서버 도구 | 감사 프로세스에서 MCP 서버를 끈다(REQ-CAR-007, 권장 기본값) | MCP 도구는 moai 서버 프로세스에서 실행되며 shell sandbox를 거치지 않는다. `default_tools_approval_mode = "writes"`와 `approval_policy="never"`의 결합이 쓰기 도구를 막는지는 측정되지 않았다 |
| codex 세션 기록(`CODEX_HOME`) | `UNSUPPORTED` 선언 | 호스트가 쓰며 감사자가 고를 수 없다. AC-DHR-023은 이 기록에서 반환문을 읽으므로 이어받은 LIVE 실행에서는 `--ephemeral`을 쓰지 않는다 |
| 프로젝트 hook 명령 | `UNSUPPORTED` 선언 | hook은 moai가 정한 명령이며 모델 입력이 아니다 |
| `-o` 출력 파일 | OS 임시 디렉터리에만 둔다 | 워크트리에 쓰지 않는다 |

MCP를 끄면 `plan-auditor`는 `spec_audit`, `audit_multi`(교차 모델 감사) 같은 MCP 도구를 쓰지 못한다. 대신 shell로 `moai spec lint` 등을 부를 수 있는지는 read-only sandbox에서 moai CLI가 상태 파일을 쓰려다 실패하는지에 달려 있으며 측정되지 않았다. `super-advisor`는 `codex_task` 위임 도구를 잃는다. [NEEDS CLARIFICATION: 감사 프로세스에서 MCP를 끄는 것(권장)과, 읽기 도구만 허용하는 것(`enabled_tools` — 의미 미측정) 중 무엇을 택할지]

MCP를 끄는 설정의 정확한 `-c` 표기는 M1에서 잰다(`mcp-server` 기동을 기록하는 래퍼로 0회 기동을 확인). 끄는 표기가 없으면 REQ-CAR-007을 "MCP를 켠 채 `UNSUPPORTED` 선언"으로 바꾸는 개정이 필요하며, 그때 리드에게 돌아간다.

### D6. 역할 범위

launcher가 받는 역할은 계약에서 계산한 read-only 역할 넷 전부다(REQ-CAR-002). 판정 파일을 쓰는 것은 요청자가 목적지를 줄 때뿐이며, 이는 사실상 `plan-auditor`, `sync-auditor`의 쓰임새다. `mission-governor`, `super-advisor`는 결과를 반환만 받을 수도 있다(목적지 인자 생략 시 표준 출력으로만 반환). 네 역할 모두 `spawn_agent`로 띄우면 부모가 쓰기 가능할 때 쓰기가 가능하다는 결함이 같으므로 함께 옮긴다. Codex에서 `mission-governor`를 실제로 누가 어떻게 띄우는지는 이 plan에서 확인하지 않았다(`internal/cli/goal.go`의 `--governor-receipt`만 확인).

### D7. 두는 곳

- 동사: `moai codex audit <role> [--out <path>]`, 작업 문은 stdin. `moai codex`의 기존 동사 체계(`cli`, `status`, `app`)에 더한다. R2가 채택되면 같은 핵심 함수를 부르는 MCP 도구 하나를 더한다.
- 지시면: 템플릿 `internal/template/templates/AGENTS.md.tmpl`의 `audit-verdict-file` 행과, 방출기 매니페스트 `internal/template/agentemit/agents-codex.yaml`의 `codex_role_addenda`(read-only 역할 넷). 역할 TOML은 `make agents-emit`로만 다시 만든다. `.codex/agents/moai/*.toml`을 손으로 고치지 않는다.
- Template-First: 템플릿을 먼저 고치고 `make build`. 템플릿에는 SPEC ID, 카드 번호, 날짜를 넣지 않는다(AC-CAR-008).
- `.claude/agents/moai`(C1, C2)는 건드리지 않는다.

### D8. 이어받은 AC의 실현

`acceptance.md` §B 참조. (ii)는 테스트가 launcher를 직접 부르는 형태로 실현하므로 호출 수가 14로 남는다. 부모 Codex 세션 경유는 AC-CAR-010이 잰다.

## §C 마일스톤 (우선순위 순, 앞 단계가 끝나야 다음 단계)

| M | 우선순위 | 내용 | 산출 |
|---|---|---|---|
| M1 | High | 실행 경로와 인자 계약 측정(LIVE, 상한 5). P-A: 최상위 `codex exec -s read-only`에 프로젝트 config `sandbox_mode="workspace-write"`와 `[mcp_servers.moai]`를 둔 상태로, 시험용 지시문 nonce 회신 + shell 쓰기 시도 + MCP 비활성화 표기 확인 + `--json`/`-o` 반환문 비교(1회). P-B: 부모 `-s workspace-write` 세션이 shell로 read-only 자식을 띄울 수 있는지(2회). P-C: P-B가 모델에 닿지 못할 때만 MCP 경로(2회) | `.moai/reports/t1143/m1-route/`(인자, 세션 기록 요약, 호출 원장), progress.md §E.2 |
| M2 | High | launcher 핵심 함수와 `moai codex audit` 동사: 인자 계약, 역할 자격, 원문 쓰기, 목적지 검증, 실패 경로, 상한 | AC-CAR-001 ~ 006 |
| M3 | Medium | 지시면: `AGENTS.md.tmpl` 행, `agents-codex.yaml` 부록, `make agents-emit`, 기존 `audit_role_exception_test.go`·`golden_test.go`의 문구 기대값 갱신. R2 채택 시 MCP 도구와 `moai-mcp-tools.md` 도구 수 갱신 | AC-CAR-007, 008 |
| M4 | Medium | 이어받은 LIVE 테스트의 (ii)를 launcher로 바꾸고 AC-DHR-012/023 실행(14회) | AC-DHR-012, 023, AC-CAR-009 |
| M5 | Medium | AC-CAR-010(3회), AC-CAR-011(2회) LIVE, 판정서 `.moai/reports/t1143/verdict.md` | AC-CAR-010, 011 |

## §D LIVE 호출 상한 (REQ-CAR-010)

단위는 `codex exec` 프로세스 하나다. 부모 세션 안에서 뜬 자식 `codex exec`도 하나로 센다. `claude -p` 호출은 없다.

| 항목 | 예정 | 상한 | 상한 초과 시 |
|---|---|---|---|
| M1 경로 측정 (P-A 1, P-B 2, 필요 시 P-C 2) | 3 | 5 | 6번째를 시작하지 않고 `ABORTED`, 리드에게 반환 |
| AC-DHR-012 + AC-DHR-023 (같은 실행) | 14 | 실행당 14. 재실행은 하네스 결함으로 무효가 된 실행이 리드에게 `INVALID`로 기록된 경우 1회만 → 최대 28 | 15번째를 시작하지 않고 `ABORTED`(출처 판정식의 의미 그대로) |
| AC-CAR-010 | 3 | 3 | 4번째 `ABORTED` |
| AC-CAR-011 | 2 | 2 | 3번째 `ABORTED` |
| **합계** | **22** | **38** | 합계 38에 닿으면 남은 LIVE를 모두 멈추고 리드에게 반환 |

- 인증 실패, 네트워크 실패로 끝난 호출도 사용한 호출로 센다.
- `ABORTED`, `NOT_RUN`, `INVALID`는 PASS가 아니다. `INVALID` 실행의 증거는 지우지 않고 이름을 바꿔 보존한다(t1100 AC-018 선례).
- LIVE 실행 전후 `CODEX_HOME` 로그인 파일의 sha256을 기록한다(t1100 M8 선례). 잔존 `codex exec` 프로세스가 없음을 마감 시점에 확인한다.
- 모든 LIVE 실행은 격리된 임시 `CODEX_HOME`·`MOAI_HOME`·저장소에서 한다. 실제 프로젝트 트리에 탐침 파일을 쓰지 않는다.

## §E 위험

| 위험 | 영향 | 대응 |
|---|---|---|
| R1 중첩 실행이 네트워크 차단(`network_access: false`)으로 모델에 닿지 못함 | R1 불가 | M1 P-B로 먼저 확인, R2로 전환 |
| R2에서 MCP 도구 시간 한도가 감사 시간보다 짧음 | 감사 중단 | 비동기 작업 형태(`codex_task` 방식) 검토. M1 P-C에서 소요 시간 기록 |
| `-s` 플래그가 config `sandbox_mode`에 지는 경우 | read-only 보장 붕괴 | AC-CAR-010 (a)가 config `workspace-write` 조건에서 잰다. 지면 `--ignore-user-config` 등 대안을 판정서에 적고 리드에게 반환 |
| MCP 비활성화 표기가 없음 | REQ-CAR-007 불충족 | D5의 개정 경로 |
| `-c developer_instructions`가 최상위 세션에서 개발자 지시로 들어가지 않음 | 역할 지시 누락 | M1 nonce 측정. 실패하면 stdin 앞머리 전달로 바꾸고 지시 계층 차이를 판정서에 적는다 |
| 감사자가 읽는 저장소 내용의 프롬프트 주입 | 판정문 오염 | 반환문은 데이터로만 쓰고 실행하지 않으며, 목적지는 모델 출력에서 받지 않는다(REQ-CAR-005). 판정 내용의 진위는 이 SPEC 범위 밖 |
| 이어받은 판정식이 `.moai/reports/t1100/` 경로를 씀 | 이름 혼동 | 워크트리 안에서만 실행. 판정서에 t1143 카드 소속임을 적는다 |
| 이어받은 (ii) 문구와 실현 방식의 차이 | 판정 해석 분쟁 | §G B1로 운영자 확인 |

## §F 바뀔 파일 (예상)

- 신규: `internal/cli/codex_audit_launch.go`, `internal/cli/codex_audit_launch_test.go`, LIVE 테스트 파일 1개
- 변경: `internal/cli/codex_launcher.go`(동사 등록), `internal/cli/codex_role_live_test.go`((ii) 경로), `internal/config/defaults.go`(시간 한도 상수), `internal/template/templates/AGENTS.md.tmpl`, `internal/template/agentemit/agents-codex.yaml`, `internal/template/agentemit/audit_role_exception_test.go`, `internal/template/agentemit/golden_test.go`
- 재생성: `internal/template/templates/.codex/agents/moai/{mission-governor,plan-auditor,super-advisor,sync-auditor}.toml`
- R2 채택 시: `internal/cli/mcp_*.go` 1개, `.claude/rules/moai/core/moai-mcp-tools.md`와 템플릿 미러

## §G 운영자 확인이 필요한 항목 (blocker 목록)

- **B1** 이어받은 AC-DHR-012 (ii)의 "부모 lane 오케스트레이터 세션을 두 번 띄워 … 하위 에이전트가" 문구를, 판정식은 그대로 둔 채 "테스트가 launcher를 직접 불러 최상위 read-only 프로세스를 띄운다"로 실현하는 것(`acceptance.md` §B). 문구대로 부모 세션을 거치면 호출 수가 16이 되어 판정식 `invocations==14`가 성립하지 않는다.
- **B2** 감사 프로세스에서 MCP 서버를 끄는 것(D5). 끄면 `plan-auditor`의 교차 모델 감사와 `super-advisor`의 위임 도구가 Codex 경로에서 빠진다.
- **B3** 감사 프로세스에 로컬 지시 파일을 합치지 않는 것(D4).
- **B4** R1·R2가 모두 실패할 때의 처리(D1).
- **B5** LIVE 상한 합계 38(예정 22)의 승인.
