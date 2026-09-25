---
id: SPEC-CODEX-PREAPPROVAL-PROBE-001
title: "Codex per-tool pre-approval of codex_role_audit — discriminating probe, conditional emission, loader key-name check, refusal instruction, carried AC-CAR-010/011"
version: "0.5.5"
status: draft
created: 2026-09-25
updated: 2026-09-25
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/codexwiring, internal/cli, internal/template/templates"
lifecycle: spec-anchored
tags: "codex, mcp, approval, pre-approval, audit, codex_role_audit, live, loader"
tier: M
card: t1172
depends_on:
  - SPEC-CODEX-AUDIT-READONLY-001
related_specs:
  - SPEC-CODEX-WIRING-001  # owns the create-if-absent [mcp_servers.moai] writer and the doctor's canonical-shape report
---

# SPEC-CODEX-PREAPPROVAL-PROBE-001

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.5.5 | 2026-09-25 | 재시도할 때 이전 무효 시도의 LIVE를 최신 반출과 잘못 비교하지 않도록 반출별 매니페스트 사본과 장부 해시를 연결했다. 최종 시도의 사본만 최신 반출본과 비교한다. |
| 0.5.4 | 2026-09-25 | 각 LIVE가 참조한 시작 검사의 고유 원자료와 성공 조건을 재검증하고, M1-a 원본 장부 네 행·원자료 12개를 고정 해시로 보존한다. |
| 0.5.3 | 2026-09-25 | M1-a와 M1-b가 서로 다른 테스트 임시 루트를 쓰므로 기존 시작 검사 행을 보존하고 새 루트에서 다시 검사하도록 명시했다. M1-a 4회, 판별 시도별 4회씩 최대 8회, car010·car011 시도별 최대 2회씩 4회를 합한 시작 검사 절대 상한은 16회다. 각 LIVE는 같은 임시 루트의 앞선 시작 검사와 짝을 이룬다. |
| 0.5.2 | 2026-09-25 | `codex-cli 0.157.0`의 `codex exec` 인자 계약을 반영했다. `--ask-for-approval`은 이 하위 명령에서 거부되므로, 판별 LIVE는 지원되는 `-c approval_policy="never"`로 무인 승인 정책을 고정한다. 반출 인자와 장부가 이 명령을 증명하도록 AC-CPP-002를 보강했다. |
| 0.5.1 | 2026-09-25 | Codex CLI 0.157.0의 비모델 로더 실측에 맞춰 AC-CPP-008 진단 형식을 개정했다. 정상 설정은 non-git 게이트에 도달하고, 오타·잘못된 enum은 `Error:`로 시작한다. 오류 판정은 이 형식과 이전 `Error loading config.toml:` 형식만 허용하며 진단 본문을 함께 검사한다. |
| 0.5.0 | 2026-09-25 | 운영자 Kickoff 결정 확정: D1-sub는 `.moai/config/sections/` YAML 불리언 키, D2는 기존 사용자 소유 표에 대한 `moai doctor` WARNING과 수동 추가 안내. D1의 A1/A2 선택은 판별 LIVE 뒤에 남긴다. 이 실행의 CLI 고정 버전을 설치된 `codex-cli 0.157.0`으로 개정하며, 0.156.1에서 얻은 과거 시작 검사 근거는 새 버전에서 다시 확인한다. |
| 0.4.0 | 2026-09-25 | plan-audit iter 3(FAIL 0.81, `.moai/reports/t1172/plan-audit-iter3.md`) 뒤 리드 조건부 PASS-WITH-DEBT — 추가 plan-audit 없이 기계 게이트로 확인. D-N11: 이월 AC의 인자·프롬프트를 호출 라벨마다(`inputs/<라벨>/argv.txt`, `prompt.txt`) 반출하고, AC-CPP-013은 각 LIVE 행을 자기 라벨의 파일과만 비교한다(REQ-CPP-009). D-N12: AC-CPP-010은 행 전체를 "베이스 행 + 고정 문장"과 바이트 비교한다. D-N13: 장부 행에 테스트 임시 디렉터리(`temp_root`)를 적고 픽스처 루트가 그 아래임을 확인한다(REQ-CPP-004). D-N14: 한 토큰짜리 모델 지정(`--model=…`, `-m…`, `-c model=…`)도 거부한다. D-N15: `OPERATOR-CONFIRM`은 모든 갈래에서 Kickoff, 첫 LIVE 전에 기록한다. D-N16: 호출마다 사용자 턴 1개를 판정한다. AC 수는 그대로 16. |
| 0.3.0 | 2026-09-25 | plan-audit iter 2(FAIL 0.79, `.moai/reports/t1172/plan-audit-iter2.md`)의 D-N1–D-N10 반영. D-N1: LIVE 행에 시도 번호(`attempt`)를 두고, 리드가 적는 `invalidate` 행으로 무효가 된 시도를 표시하며, 판정식은 마지막 유효 시도만 보고 창 상한은 시도마다 잰다(REQ-CPP-010). D-N2: 픽스처마다 마지막 반출 뒤의 마지막 시작 검사만 반출본과 대조하고, 채택 갈래의 car010은 M2 뒤에 다시 반출·시작 검사한다. D-N5: REQ-CPP-008에 고정 규범 문장을 두고 AC-CPP-010이 그 문장을 그대로 요구한다. D-N6: 픽스처 초기화는 테스트 임시 디렉터리 아래 표지 파일이 있는 픽스처 루트에서만 한다(REQ-CPP-004). D-N3·D-N4·D-N7·D-N8·D-N9·D-N10도 반영. AC 수는 그대로 16. |
| 0.2.0 | 2026-09-25 | plan-audit iter 1(FAIL 0.70, `.moai/reports/t1172/plan-audit-iter1.md`)의 D1–D27 반영. D1: plan.md의 미해결 확인 표시 두 개를 **잠정 기본값**(운영자 Kickoff 확정 필요)으로 바꿨다 — D1-sub(A2일 때만)=`.moai/config/sections/` YAML의 불리언 키, D2=`moai doctor` WARNING+고침 안내(REQ-CPP-005/006). D1(A1 대 A2) 자체는 여전히 판별 결과 뒤 운영자 결정이다. REQ-CPP-003 문구(D24), REQ-CPP-005에 `NOT DISCRIMINATING`(D23), F4 인용 118행(D18), 감사 측정 사실(신뢰된 git 픽스처에서 moai 기동 1·`item` 0, `model = "gpt-5"`의 비모델 `item` 오류, 서버가 루트에 `.moai/state/config-cache.json`을 씀)을 §A에 반영했다. AC-CPP-012(게이트 거부 시 LIVE 0회), AC-CPP-013(이월 픽스처 조건), AC-CPP-014(장부 상한)를 더했다(AC 14 + 이월 2 = 16, Tier M 상한 안). |
| 0.1.0 | 2026-09-25 | 카드 t1172 plan 초안. 선행 SPEC-CODEX-AUDIT-READONLY-001이 이월한 네 항목(도구별 사전 승인 효과 `NOT MEASURED`, AC-CAR-010 `INVALID`, AC-CAR-011 `NOT_RUN`, 픽스처 입력 반출·비모델 신뢰 검사·템플릿 기본값 운영자 승인·키 이름 로더 검증)과 sync-audit F4를 받았다. 리드의 HARD 조건 여섯 개를 REQ와 AC로 옮겼다. AC-CAR-010/011의 본문·판정식은 경로 치환 한 가지(`reports/t1143` → `reports/t1172`)만 적용해 바이트 그대로 이어받았고, 치환은 AC-CPP-011이 기계적으로 검사한다. |

## §A 배경과 목적

선행 카드 t1143은 Codex의 read-only 감사 역할을 moai MCP 도구 `codex_role_audit`로 띄우게 했다. 그런데 moai가 만드는 프로젝트 `.codex/config.toml`의 `[mcp_servers.moai]`는 `default_tools_approval_mode = "writes"`이고, 무인 lane은 `approval_policy = "never"`로 돈다. 이 조합에서 Codex는 쓰기 가능한 `codex_role_audit` 호출을 거부했다(#37, 판정 VALID). 거부 원문:

```
MCP tool call requires approval, but approval policy is never
```

리드는 이 도구를 read-only로 표시하는 방안을 기각했다(쓰기 도구를 읽기 도구로 적는 부정직한 표시). 대신 moai가 만드는 설정이 이 도구 하나만 도구별 표(`[mcp_servers.moai.tools.codex_role_audit]` `approval_mode = "approve"`)로 사전 승인하는 방향을 골랐다. 도구는 쓰기 도구로 남는다.

그 표가 `approval_policy = "never"`를 이기는지는 아직 측정되지 않았다. #38은 프롬프트가 codex-cli 0.156.1이 MCP 도구에 닿는 통로인 `exec` custom tool을 막아 `INVALID`였다. #39는 픽스처 루트가 git 저장소가 아니었고 `CODEX_HOME` 신뢰 항목이 통하지 않아 세션이 시작되지 않았다(`NOT MEASURED`). B''는 #37의 `CODEX_HOME` 설정, 루트의 저장소 여부, moai 빌드 커밋을 반출하지 않아 재구성할 수 없었다(`NOT RUN`). 근거: primary checkout에 반출된 `.moai/reports/t1143/verdict.md`(sha256 `00eea6419da680ac9b178a01e2bd316001a0611de7139a22242d9525f8fb18ae`, 128–278행)와 `.moai/reports/t1143/sync-audit.md`(sha256 `93f8bedb6694e803164c85382ee120b2aa612be842f8ec76812472aedb86cca3`, F4는 118행).

이 SPEC의 목적은 넷이다.

1. 반출된 입력만으로 재구성할 수 있는 픽스처에서, 대조 팔과 처치 팔의 차이를 도구별 표 하나로 좁혀 사전 승인 효과를 한 번 잰다.
2. 측정 결과와 운영자 결정에 따라 moai가 만드는 설정에 그 표를 넣거나 넣지 않는다. 넣는 형태(템플릿 기본값 대 opt-in)는 운영자가 정한다.
3. 어느 쪽이든 무인 lane이 거부를 만났을 때 할 일을 지시면에 적는다(F4). 감사를 조용히 건너뛰지 않고 blocker를 돌려준다.
4. 이월된 AC-CAR-010과 AC-CAR-011을 고친 픽스처로 다시 돌린다.

plan 단계에서 모델 호출 없이 잰 사실(명령과 출력은 `.moai/reports/t1172/verdict.md` §Plan, 원자료는 `.moai/reports/t1172/plan-checks/`):

- `codex mcp get moai --json`의 출력에는 `default_tools_approval_mode`도 도구별 표도 나오지 않는다. 그래서 이 명령으로는 해석된 도구별 승인을 되읽을 수 없다.
- `codex exec --strict-config`는 설정 파일의 모르는 필드를 적재 단계에서 거부한다. 철자가 틀린 `approval_modex`를 넣으면 세션을 시작하지 않고 `unknown configuration field \`mcp_servers.moai.tools.codex_role_audit.approval_modex\``로 끝난다(exit 1). `codex --strict-config mcp …`와 `codex --strict-config debug …`는 지원되지 않는다.
- 이번 대상 `codex-cli 0.157.0`의 별도 비모델 재측정에서 정상 설정은 `Not inside a trusted directory and --skip-git-repo-check was not specified.` 게이트에 닿았고, 오타 설정은 `Error: <임시 경로>/.codex/config.toml:<행>:<열>: unknown configuration field \`mcp_servers.moai.tools.codex_role_audit.approval_modex\``, 잘못된 enum은 같은 `Error:` 형식으로 `unknown variant \`bogus\``를 냈다. 세 경우 모두 exit 1이므로 종료 코드만으로는 구별하지 않는다.
- 같은 설치본의 `codex exec --help`는 `-c`·`--strict-config`·`-s`·`-C`·`--json`을 보여 준다. `codex exec --ask-for-approval never --help`는 `unexpected argument '--ask-for-approval'`로 exit 2이고, `codex exec --strict-config -s workspace-write -c 'approval_policy="never"' --help`는 exit 0이다. 판별 명령은 승인 정책 `never`를 지원되는 설정 재정의로 전달한다.
- git 저장소가 아닌 디렉터리에 `CODEX_HOME` 신뢰 항목을 두면 `codex debug prompt-input`은 그 디렉터리를 신뢰한 것으로 다룬다(기본 sandbox가 `workspace-write`로 렌더됨, 신뢰 항목이 없으면 `read-only`). 그러나 같은 디렉터리에서 `codex exec`는 `Not inside a trusted directory and --skip-git-repo-check was not specified.`로 시작을 거부한다. #39를 모델 호출 없이 재현한 것이다.
- git 저장소 루트에서는 신뢰 항목이 없어도 `codex exec`가 시작 검사를 통과한다. 접속할 수 없는 모델 제공자(`127.0.0.1:9`)를 주면 `thread.started`, `turn.started` 뒤에 접속 실패만 되풀이되고, 모델 응답과 `item.*` 이벤트는 0개였다. 요청은 나가지만 모델 끝점에 닿지 못하므로 응답이 없다. 이 실행은 신뢰 항목이 없는 루트여서 프로젝트 층 적재와 MCP 기동은 보이지 않았다.
- plan-audit iter 1이 같은 조건(로그인 파일 없는 `CODEX_HOME`, 제공자 `127.0.0.1:9`, 신뢰 항목이 있는 git 루트, 기동 기록 래퍼)으로 잰 결과: 기본 모델에서 `thread.started` 1회, `item.` 0회, stderr 0바이트, moai `mcp-server` 기동 1회. 설정에 `model = "gpt-5"`를 두면 모델 응답이 아닌 `item.completed`(`type:"error"`, "Model metadata for `gpt-5` not found…")가 1회 나온다. 또 moai MCP 서버는 기동하면서 픽스처 루트에 `.moai/state/config-cache.json`을 쓴다(원자료 `.moai/reports/t1172/plan-audit-iter1-probes/`).

## §B 요구사항 (GEARS)

### REQ-CPP-001 — 판별 LIVE 호출 전에 픽스처 입력을 모두 반출한다

When a discriminating LIVE invocation is about to start, the probe harness shall first have written into the evidence directory, for each arm, the `CODEX_HOME` configuration file including its project trust entries, the project configuration file, whether the working root is a git repository together with its HEAD commit, the build commit and binary sha256 of the moai executable that serves as the MCP server, the complete argument vector, and the complete prompt text. When any of these inputs is missing from the evidence directory, the probe harness shall start no discriminating LIVE invocation.

### REQ-CPP-002 — 두 팔의 차이는 도구별 사전 승인 표 하나뿐이다

When both arms' inputs have been exported, the probe harness shall write the verbatim recursive diff between the control arm's inputs and the treatment arm's inputs into the evidence directory before either arm runs. When that diff contains any line other than the lines that add the single per-tool pre-approval table for `codex_role_audit`, the probe harness shall run neither arm and shall record the diff as the reason.

### REQ-CPP-003 — 모델 요청 없는 시작 검사를 먼저 통과한다

When an arm's fixture is prepared, the probe harness shall run a startup check against that fixture that cannot reach a model endpoint and receives no model response — a `codex exec` invocation under `--strict-config` that defines and selects the unreachable model provider `http://127.0.0.1:9/v1`, whose `CODEX_HOME` holds no login file and whose environment carries no authentication variable, with no model override in its arguments or in either configuration file (the same model setting as the LIVE invocation), bounded by a timeout — and shall require that the session started, that the configuration loaded without an error, that no item event other than a non-model `error` item occurred, and that the moai MCP server was launched. When the startup check fails for either arm, the probe harness shall start no discriminating LIVE invocation. The startup check shall not be counted as a LIVE invocation and shall be recorded in the ledger with its own cap. An earlier test's startup rows shall remain in the ledger and count toward that cap even if its temporary fixture has been removed. Before each LIVE invocation, the ledger shall identify an earlier successful startup check for the same fixture in the same temporary and fixture roots, whose retained per-invocation output proves the check and whose loaded input hashes match the export snapshot for that attempt. The export snapshot used by each startup and LIVE invocation shall remain available after a retry; only the latest valid attempt is judged against the current export. The four original M1-a startup rows and their output shall remain intact when later checks are appended. When a fixture's inputs are exported again (for example the car010 project configuration after the adopted writer change), the startup check shall be run again after that export and before the fixture's next LIVE invocation.

### REQ-CPP-004 — LIVE 상한과 정지 규칙을 먼저 적는다

The probe harness shall run under caps declared in writing before the first LIVE invocation: per-invocation timeout, total wall-clock, one user turn per invocation, and an exact invocation count for the discriminator and for each carried LIVE acceptance criterion. The discriminator prompt shall permit the `exec` custom-tool path through which the Codex CLI reaches MCP tools. An arm's result shall be classified as `REFUSED`, `RAN`, `NOT MEASURED`, `INVALID`, or `ABORTED`; an arm with zero `mcp_tool_call` items addressed to server `moai`, tool `codex_role_audit` shall be `NOT MEASURED`. When a cap would be exceeded, the probe harness shall start no further invocation and shall record `ABORTED`. The probe harness shall terminate only process identifiers it recorded itself. The probe harness shall run a destructive fixture reset only against a fixture root that is a non-empty absolute path under the test's temporary directory, is the top level of its own git repository, lies outside every worktree of this repository, and carries the sentinel file written when the fixture was created; otherwise it shall refuse the reset and start no invocation. Every startup-check and LIVE ledger row shall record that temporary directory, and the row's fixture root shall lie under it.

### REQ-CPP-005 — 채택 형태는 운영자가 정한다

Where the treatment arm is measured `RAN` and the control arm is measured `REFUSED`, the adoption form of the per-tool pre-approval — a template default or an opt-in — shall be the one the operator selects, and the SPEC shall not pre-decide it. When the treatment arm is measured `REFUSED` or `NOT MEASURED`, or when the control arm is measured `RAN` (`NOT DISCRIMINATING`), the adoption decision shall be `not adopted` and moai shall emit no per-tool pre-approval. Where the operator selects the opt-in form, the opt-in shall be a boolean key in a `.moai/config/sections/` YAML file that the Codex configuration generator reads, so the choice persists across `moai update` (운영자 Kickoff 결정: `d1sub=config-section-boolean`; 리드는 첫 LIVE 전에 `verdict.md`의 `OPERATOR-CONFIRM` 줄로 기록한다).

### REQ-CPP-006 — 사전 승인은 이 도구 하나로 한정하고 도구는 쓰기 도구로 남는다

Where the per-tool pre-approval is adopted, the moai-generated Codex configuration shall pre-approve `codex_role_audit` and no other tool, shall keep `default_tools_approval_mode = "writes"` for the server, and shall not change the tool's write-capable annotation. The configuration writer shall keep an existing user-owned `[mcp_servers.moai]` table byte-invariant. Where the per-tool pre-approval is adopted and a project's `[mcp_servers.moai]` table exists without the per-tool table, `moai doctor` shall report a warning, not a failure, together with a hint naming the table to add (운영자 Kickoff 결정: `d2=doctor-warning`; 리드는 첫 LIVE 전에 `verdict.md`의 `OPERATOR-CONFIRM` 줄로 기록한다).

### REQ-CPP-007 — 생성 설정의 키 이름을 Codex 로더로 검증한다

The test suite shall validate the key names of the moai-generated Codex configuration through the Codex CLI's own configuration loader on a path that makes no model request, and shall prove its own detection with a positive control: a configuration carrying a deliberately misspelled per-tool key shall be reported by the test as an unknown field naming that key. When the Codex CLI is not installed, the test shall skip with an explicit reason, and a skip shall never be recorded as a pass.

### REQ-CPP-008 — 무인 lane은 거부를 blocker로 돌려준다 (F4)

The deployed Codex instruction surface for starting read-only audit roles shall state what a lane does when `codex_role_audit` is refused for approval: the lane shall not skip the audit, shall not fall back to `spawn_agent`, and shall return a blocker that quotes the refusal text. The instruction shall carry exactly this sentence, once: "When the call is refused with `MCP tool call requires approval, but approval policy is never`, do not skip the audit and do not fall back to `spawn_agent`; return a blocker that quotes the refusal text". The template text shall carry no SPEC identifier, card identifier, or date.

### REQ-CPP-009 — 이월된 AC-CAR-010/011을 고친 픽스처로 다시 돌린다

The carried acceptance criteria AC-CAR-010 and AC-CAR-011 shall be run with their bodies and judges unchanged except for the declared evidence-path substitution. Their fixtures shall satisfy REQ-CPP-001 and REQ-CPP-003, shall use a git repository as the working root, and shall prevent the user-layer decoy server from exposing `codex_role_audit` to the parent session. Because the carried tests start more than one `codex exec` invocation with different arguments and prompts, the probe harness shall reset the fixture root and export the argument vector and prompt of each invocation under its own invocation label before that invocation starts, and each LIVE ledger row shall carry that label. Where the adoption decision is `not adopted`, AC-CAR-010 shall be recorded `NOT_RUN` with the reason, and AC-CAR-011, which does not pass through the approval path, shall still run.

### REQ-CPP-010 — LIVE 결과는 부풀리지 않는다

The verification shall count one `codex exec` process as one LIVE invocation, shall not start any invocation that would exceed the absolute cap declared in `plan.md` §D, and shall never report `SKIP`, `NOT_RUN`, `NOT MEASURED`, `INVALID`, or `ABORTED` as a pass. A rerun shall be allowed only after the lead records the previous attempt `INVALID` in the ledger, at most once per group (the discriminator pair, AC-CAR-010, AC-CAR-011); each LIVE invocation shall carry its attempt number, the per-attempt invocation counts and wall-clock windows of `plan.md` §D shall apply to each attempt separately, the absolute cap shall count every attempt, and every acceptance judgement shall read only the last attempt that the lead has not invalidated.

### REQ-CPP-011 — 선행 보안 수리가 들어간 트리에서만 채택한다

Where the per-tool pre-approval is adopted, the tree shall contain the landed t1143 F1/F2/F3 confinement fixes.

## §C 추적

| REQ | AC |
|---|---|
| § REQ-CPP-001 | AC-CPP-002, AC-CPP-012 |
| § REQ-CPP-002 | AC-CPP-003, AC-CPP-012 |
| § REQ-CPP-003 | AC-CPP-004, AC-CPP-012 |
| § REQ-CPP-004 | AC-CPP-005, AC-CPP-012, AC-CPP-014 |
| § REQ-CPP-005 | AC-CPP-006 |
| § REQ-CPP-006 | AC-CPP-007 |
| § REQ-CPP-007 | AC-CPP-008, AC-CPP-009 |
| § REQ-CPP-008 | AC-CPP-010 |
| § REQ-CPP-009 | AC-CPP-011, AC-CPP-013, AC-CAR-010, AC-CAR-011 |
| § REQ-CPP-010 | AC-CPP-005, AC-CPP-014, AC-CAR-010, AC-CAR-011 |
| § REQ-CPP-011 | AC-CPP-001 |

## §D 제외 사항

### Out of Scope — 다른 카드가 가진 항목

- t1171: AC-DHR-012 / AC-DHR-023 / AC-CAR-012b의 역할 로드 판정과 작업 없는 요청을 거절하는 역할 계약의 충돌.
- t1170: AC-FLH-015 owner-SPEC 문구 drift.
- t1173: docs-site `guides/mcp-server.md`의 도구 수 서술(F11).

### Out of Scope — 선행 sync-audit의 나머지 부채

- F6(`completed` 대 §E 문구), F7(판정 파일 모드 0600), F8(서버 종료 시 MCP 작업 수명), F9(hook·notify 쓰기 주체, 미측정), F10(testdata의 로컬 절대 경로), F11(t1173으로 이관), F12(Windows launcher 테스트는 컴파일만), F13(`--out` 상대 경로 기준 차이). 이 SPEC은 이것들을 고치지 않는다.

### Out of Scope — 우회와 다른 방향

- `codex_role_audit`를 read-only로 표시하는 일(리드가 기각).
- lane을 승인 가능한 정책(`approval_policy`가 `never`가 아닌 값)으로 돌리는 일.
- `--skip-git-repo-check`, `--dangerously-bypass-approvals-and-sandbox` 같은 우회 옵션을 픽스처나 제품에 쓰는 일.
- 다른 moai MCP 도구의 사전 승인.
- 이미 있는 사용자 소유 `[mcp_servers.moai]` 표를 writer가 고쳐 쓰는 일.

### Out of Scope — 배포

- push, PR, develop 병합. 병합은 리드의 통합 창에서 따로 한다.
