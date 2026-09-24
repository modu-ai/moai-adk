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
| `codex exec` 옵션 전체(0.156.1): `-c`, `--enable`, `--disable`, `--strict-config`, `-i`, `-m`, `--oss`, `--local-provider`, `-p/--profile`, `-s/--sandbox`, `--approve-for-me`, `--dangerously-bypass-approvals-and-sandbox`, `--dangerously-bypass-hook-trust`, `-C/--cd`, `--worktree`, `--add-dir`, `--thread-source`, `--skip-git-repo-check`, `--ephemeral`, `--ignore-user-config`, `--ignore-rules`, `--output-schema`, `--color`, `--json`, `-o/--output-last-message`, `-h`, `-V` | `codex exec --help 2>&1 \| grep -nE '^\s+(-[a-zA-Z], )?--[a-z-]+'`(0.2.0에서 실행) |
| 하위 에이전트는 부모 sandbox를 물려받고, 최상위 `-s read-only` 세션은 부모 자신의 쓰기와 하위 에이전트의 쓰기를 모두 막았다 | `.moai/reports/t1100/m8-sbx/summary.json` run1·run2, `run1-argv.txt`, `run2-argv.txt` |
| run2의 프로젝트 config에는 `sandbox_mode` 키가 없었다. 따라서 "`-s` 플래그가 config의 `sandbox_mode`를 이긴다"는 아직 측정되지 않았다 | `summary.json` `.runs.run2.config_toml` |
| `workspace-write` 세션의 `sandbox_policy`에 `network_access: false`가 있었다 | `summary.json` run1 `turn_contexts[0].sandbox_policy` |
| 방출된 read-only 계약 역할은 넷: `mission-governor`, `plan-auditor`, `super-advisor`, `sync-auditor`. 넷 모두 C1(`.claude/agents/moai/*.md`), C2(템플릿 미러), C3(`.codex/agents/moai/*.toml`)에 있다 | `ls` 세 디렉터리, TOML `sandbox_mode` 판독 |
| 역할 TOML 크기: `plan-auditor` `developer_instructions` 61,688바이트(`developer_instructions=<JSON>` 인자로 62,843바이트), `sync-auditor` 18,505, `super-advisor` 7,596, `mission-governor` 1,248. effort는 넷 모두 `high` | TOML 파싱 + JSON 인코딩 길이 측정(python `tomllib`/`json`) |
| 기존 launcher의 로컬 지시 주입기는 `internal/cli/codex_launcher.go:118-143`(`codexLocalDeveloperInstructionArgs`)이고, 상한 검사는 `:176-181`(`checkCodexInstructionSize`, "Measure the final representation, not the source body")이다. 상한 값은 `config.DefaultCodexInstructionArgBytes` = 126,976 | 해당 줄 판독 |
| 이 워크트리의 `CLAUDE.local.md`는 61,360바이트다. `plan-auditor` 지시문(62,843)과 합치면 인코딩 전에 이미 124,203바이트로 상한에 거의 닿는다 | `wc -c` |
| 프로젝트 `.codex/config.toml`에는 `[mcp_servers.moai]`와 `default_tools_approval_mode = "writes"`가 있다 | `.codex/config.toml`, `internal/codexwiring/configtoml.go:13` |
| 기존 `moai codex` 동사는 `cli`, `status`, `app`뿐이다 | `internal/cli/codex_launcher.go:629` |
| 이어받은 LIVE 테스트는 부모를 `-s workspace-write`로 띄우고 감사 역할을 `spawn_agent`로 부르며, (ii)에서 `codexSubagentFor(rollouts, role)`로 하위 에이전트 rollout을 찾는다. 예산 상수 14 | `internal/cli/codex_role_live_test.go:23,127-130,191-196` |
| OS별 감사 프로세스 도우미가 이미 있다: `configureClaudeAuditProcess`·`runClaudeAuditProcess`(`internal/cli/mcp_claude_process_unix.go:12,26`, `internal/cli/mcp_claude_process_windows.go:16,20`) | `grep -n '^func '` |
| 바이너리에 `developer_instructions`(64회), `enabled_tools`, `disabled_tools` 문자열이 있다. 문자열이 있다는 것은 의미를 증명하지 않는다 | `strings -n 12` 결과 grep |

## §B 결정 (바뀔 가능성이 큰 것부터)

### D1. 실행 경로 — 측정으로 정한다 (M1)

부모 Codex lane 세션이 launcher를 부를 수 있는 경로는 둘이다.

- **R1 shell 경로**: 부모 세션이 shell 도구로 `moai codex audit ...`을 실행한다. launcher와 그 아래 `codex exec`는 부모의 sandbox(`workspace-write`, `network_access: false`) 안에서 뜬다. 중첩된 `codex exec`가 모델 API에 닿는지, macOS sandbox 안에서 다시 sandbox를 걸 수 있는지는 측정되지 않았다. 네트워크 차단이 측정된 설정이므로 실패 가능성이 있다.
- **R2 MCP 경로**: moai MCP 서버에 도구 하나를 더하고 서버가 launcher를 부른다. MCP 서버는 Codex 호스트가 띄우는 프로세스이며 shell sandbox 밖에서 돈다는 것이 통상의 이해지만 이 트리에서 측정하지 않았다. 서버의 작업 디렉터리는 primary checkout이므로 워크트리 뿌리를 입력으로 받아야 하며, 그 입력은 REQ-CAR-005의 한정을 받는다(D5.1). 감사는 수 분이 걸리므로 `codex_task`와 같은 비동기 작업 형태(시작·상태·결과)로 둔다.

M1이 R1을 먼저 재고, R1이 모델에 닿지 못할 때만 R2를 잰다. 결과는 `.moai/reports/t1143/m1-route/route.txt`에 `shell`, `mcp`, `none` 중 하나로 적는다. 이 파일이 경로의 단일 출처다: AC-CAR-013이 분기에 쓰고, AC-CAR-010 판정식이 증거의 `route`와 같은지 비교하며, AC-CAR-007의 지시면 테스트도 이 파일을 읽어 "측정된 경로 하나"를 정한다. 지시면에는 측정으로 동작한 경로 하나만 적는다(REQ-CAR-011). launcher의 핵심 동작(인자 조립, 실행, 원문 쓰기, 목적지 검증, launch record)은 경로와 무관한 하나의 Go 함수로 두고, 경로는 그것을 부르는 얇은 겉면만 다르게 한다.

**R1·R2 모두 실패하면 (B4 잠정 기본값).** 카드를 멈춘다. `route.txt`에 `none`을 적고, 더 이상 LIVE 호출을 하지 않으며(REQ-CAR-011), 측정 증거와 함께 리드에게 보고한다. Codex 세션 밖에서 실행하는 경로(R3)로 범위를 넓히는 일은 이 카드에서 하지 않는다.

### D2. 실행 인자 계약

```
codex exec -s read-only -c approval_policy="never" -c model_reasoning_effort="<역할 값>" -c developer_instructions=<역할 지시문 JSON> <MCP 비활성화 인자> -C <워크트리 뿌리> --json [-o <OS 임시 경로>] -   (작업 문은 stdin)
```

- sandbox는 반드시 `-s read-only` 플래그 하나로 준다. config 값에 기대지 않는다. 플래그가 config를 이기는지는 M1과 AC-CAR-010이 잰다.
- `approval_policy="never"`: 거부된 쓰기가 승인 요청으로 멈추지 않고 실패로 끝나게 한다(t1100 M8과 같은 조건).
- 판정은 허용 목록으로 한다(AC-CAR-001). `codex exec --help`(0.156.1) 실측에서 sandbox·승인·쓰기 범위를 바꾸거나 우회하는 옵션으로 확인한 `--dangerously-bypass-approvals-and-sandbox`, `--dangerously-bypass-hook-trust`, `--approve-for-me`(도움말: "workspace-write sandbox로 자동 검토"), `--add-dir`(쓰기 가능 디렉터리 추가), `--worktree`, `-p/--profile`(설정 층 추가로 sandbox를 바꿀 수 있음), `--enable`/`--disable`(기능 토글), `--ignore-rules`, `--skip-git-repo-check`는 허용 목록 밖이다. `--ignore-user-config`는 줄이는 방향의 옵션이며 M1이 MCP 비활성화에 필요하다고 정한 경우에만 허용한다.
- 작업 문(예: 어느 SPEC을 감사하라)은 부모가 stdin으로 준다. 인자 길이 상한을 역할 지시문이 거의 다 쓰기 때문이다.
- 시간 한도는 `internal/config/defaults.go`의 이름 있는 상수 하나로 둔다. 초과 시 감사 프로세스와 그 자식을 정리한다(D8).

### D3. 반환문 채널

1순위는 `--json` 이벤트 스트림의 마지막 에이전트 메시지다. 파일 시스템에 기대지 않는다. `-o` 파일은 codex 호스트 프로세스가 쓰므로 sandbox 대상이 아닐 것으로 보지만 측정되지 않았다. M1이 두 채널이 같은 바이트를 내는지 잰다. 다르면 `--json`을 채택하고 차이를 판정서에 적는다. 출처 판정식(AC-DHR-023)은 반환문과 판정 파일의 끝 공백을 잘라 해시를 비교하지만, launcher 자체는 정규화 없이 바이트 그대로 쓴다(AC-CAR-003). 출처 판정식이 더 느슨하므로 둘은 충돌하지 않는다.

### D4. 역할 지시문 전달과 인자 상한

- 역할 지시문은 방출된 역할 TOML의 `developer_instructions`를 그대로 `-c developer_instructions=`로 준다. 효과는 TOML의 `model_reasoning_effort`를 그대로 준다. 역할 TOML 자체를 최상위 세션에 "로드"하는 방법은 codex에 없다(`codex exec --help`에 역할 선택 옵션 없음).
- **로컬 지시 파일(CLAUDE.local.md 등)은 감사 프로세스에 합치지 않는다 (B3 잠정 기본값).** 측정한 길이로 `plan-auditor` 62,843 + `CLAUDE.local.md` 61,360 = 124,203바이트이며 JSON 인코딩(Go `json.Marshal`은 `<`, `>`, `&`도 이스케이프한다) 뒤에는 126,976 상한을 넘을 수 있다. 감사자는 프로젝트 `AGENTS.md`를 codex의 파일 탐색으로 읽는다.
- **길이의 단위는 최종 인자 토큰이다.** 상한 검사는 원문이 아니라 JSON 인코딩을 거친 최종 토큰 `developer_instructions=<JSON>`의 바이트 길이로 한다. 기존 `checkCodexInstructionSize`(`internal/cli/codex_launcher.go:176-181`)와 같은 단위이며, 같은 함수와 상수를 재사용한다. 새 상수를 만들지 않는다. AC-CAR-006이 상한과 상한+1 두 경계를 이 단위로 판정한다.
- `-c developer_instructions`가 최상위 세션에서 실제로 개발자 지시로 들어가는지는 이 plan에서 확인하지 않았다. M1과 AC-CAR-010이 nonce로 직접 잰다(시험용 역할 지시문 끝에 nonce를 넣고 되돌리게 한다). 실패하면 stdin 앞머리 전달로 바꾸고 지시 계층 차이를 판정서에 적는다.

### D5. sandbox 밖 쓰기 주체와 launch record

`-s`가 다스리는 것은 모델이 낸 명령과 편집이다. 그 밖의 쓰기 주체는 다음과 같다.

| 주체 | 처리 | 근거 |
|---|---|---|
| MCP 서버 도구(모든 층) | **감사 프로세스에서 모든 MCP 서버를 끈다 (B2 잠정 기본값).** 사용자 층(`CODEX_HOME/config.toml`)과 프로젝트 층(`.codex/config.toml`)을 모두 포함한다 | MCP 도구는 MCP 서버 프로세스에서 실행되며 shell sandbox를 거치지 않는다. `default_tools_approval_mode = "writes"`와 `approval_policy="never"`의 결합이 쓰기 도구를 막는지는 측정되지 않았다. "모두 끈다"는 두 층에 각각 기록 래퍼를 두고 기동 0회로 잴 수 있으므로 측정 가능한 쪽을 택했다(AC-CAR-010) |
| codex 세션 기록(`CODEX_HOME`) | `UNSUPPORTED` 선언 | 호스트가 쓰며 감사자가 고를 수 없다. AC-DHR-023은 이 기록에서 반환문을 읽으므로 이어받은 LIVE 실행에서는 `--ephemeral`을 쓰지 않는다 |
| 프로젝트 hook 명령 | `UNSUPPORTED` 선언 | hook은 moai가 정한 명령이며 모델 입력이 아니다 |
| `-o` 출력 파일 | OS 임시 디렉터리에만 둔다 | 워크트리에 쓰지 않는다 |

MCP를 끄면 Codex 경로의 `plan-auditor`는 `spec_audit`, `audit_multi`(교차 모델 감사) 같은 MCP 도구를 쓰지 못하고, `super-advisor`는 `codex_task` 위임 도구를 잃는다. 이 손실은 B2의 확인 대상이다.

모든 층의 MCP를 끄는 정확한 인자는 M1에서 잰다. 후보는 `--ignore-user-config`(사용자 층 전체를 읽지 않음, 도움말 문구)와 프로젝트 층 서버를 끄는 `-c` 재정의다. 두 층 모두를 끄는 인자 조합을 측정으로 찾지 못하면 REQ-CAR-007을 바꾸는 개정이 필요하며, 그때 리드에게 돌아간다(끄지 못한 층을 `UNSUPPORTED`로 선언하는 쪽으로).

**launch record (REQ-CAR-007).** launcher는 루트·목적지 검증(REQ-CAR-005)을 통과한 실행마다(감사가 성공하든 실패하든) 워크트리 뿌리의 `.moai/reports/codex-audit/<role>-<UTC YYYYMMDDTHHMMSSZ>-<8자 hex>.launch.json`에 기록을 배타 생성(같은 이름이 있으면 덮어쓰지 않고 실패)으로 남기고, 표준 오류에 `LAUNCH_RECORD <상대 경로>` 한 줄을 찍는다. 프로세스를 띄우기 전에 거부된 호출(검증 실패, 자격 없는 역할, 인자 상한 초과)은 아무 파일도 쓰지 않고 표준 오류에만 사유를 찍는다. 거부된 호출의 루트는 검증되지 않았으므로 그 루트에 기록을 쓰면 REQ-CAR-005가 막는 탈출을 다시 여는 셈이다(plan-audit iter-2 N5). 스키마(버전 1)는 `acceptance.md` AC-CAR-014가 정한다: `schema_version`, `role`, `route`, `started_at`, `ended_at`, `argv`(지시문 값은 sha256·바이트 수로 대체), `sandbox`, `mcp_servers`, `covers`, `unsupported`(정확히 `codex-home-session-files`, `project-hook-commands`), `exit_code`, `failure_reason`, `verdict_path`, `verdict_sha256`. 카드 판정서는 이 기록을 인용한다.

#### D5.1 작업 루트와 목적지의 한정 (REQ-CAR-005)

- 작업 루트는 심볼릭 링크를 푼 뒤, launcher 자신의 프로젝트 뿌리와 git common directory가 같은 저장소에 **등록된 워크트리**여야 하고(`git worktree list --porcelain`의 경로 집합 + `git rev-parse --git-common-dir` 비교), 동시에 **호출자 자신의 워크트리**여야 한다. R1에서는 launcher 프로세스의 작업 디렉터리를 담은 워크트리, R2에서는 MCP 서버 프로세스가 시작된 워크트리가 호출자 자신의 워크트리다. 같은 저장소의 형제 워크트리나 primary checkout은, 등록되어 있어도 호출자 자신의 것이 아니면 거부한다. 이로써 R2에서 sandbox 안의 부모가 다른 저장소, 임의 디렉터리, 형제 레인의 워크트리, primary checkout을 작업 루트로 지정해 그 트리의 `.moai/reports/`(리드가 읽는 증거 면)에 쓰는 경로를 막는다(plan-audit D3, iter-2 N2).
- R2 전제의 측정: Codex가 띄운 MCP 서버 프로세스의 시작 디렉터리가 lane 워크트리인지는 측정되지 않았다(Claude 쪽 서버는 primary checkout에서 뜬다고 `moai-mcp-tools.md`가 적는다). M1 P-C가 이를 기록한다. 시작 디렉터리가 lane 워크트리가 아니면 R2는 호출자 워크트리를 확정할 수 없으므로 채택하지 않고 `route.txt`를 `none`으로 적는다(B4 경로).
- 목적지는 심볼릭 링크를 푼 뒤 그 작업 루트의 `.moai/reports/` 아래이되 launcher 전용 `.moai/reports/codex-audit/` 밖이어야 하고, `.git` 경로 성분을 가질 수 없다. `AGENTS.md`, `.codex/`, `.claude/` 같은 하네스 배선 경로는 이 조건으로 자동 제외된다(plan-audit D13).
- 조건을 어기면 프로세스를 띄우지 않고 아무것도 쓰지 않는다. 판정은 AC-CAR-005(공통)와 AC-CAR-013(R2 채택 시 MCP 입력)이 한다.

### D6. 역할 범위

launcher가 받는 역할은 계약에서 계산한 read-only 역할 넷 전부다(REQ-CAR-002). 판정 파일을 쓰는 것은 요청자가 목적지를 줄 때뿐이며, 이는 사실상 `plan-auditor`, `sync-auditor`의 쓰임새다. `mission-governor`, `super-advisor`는 결과를 반환만 받을 수도 있다(목적지 인자 생략 시 표준 출력으로만 반환). 네 역할 모두 `spawn_agent`로 띄우면 부모가 쓰기 가능할 때 쓰기가 가능하다는 결함이 같으므로 함께 옮긴다. Codex에서 `mission-governor`를 실제로 누가 어떻게 띄우는지는 이 plan에서 확인하지 않았다(`internal/cli/goal.go`의 `--governor-receipt`만 확인).

### D7. 두는 곳

- 동사: `moai codex audit <role> [--out <path>]`, 작업 문은 stdin. `moai codex`의 기존 동사 체계(`cli`, `status`, `app`)에 더한다. R2가 채택되면 같은 핵심 함수를 부르는 MCP 도구 하나를 더하고, `.claude/rules/moai/core/moai-mcp-tools.md`와 템플릿 미러의 도구 수를 고친다(AC-CAR-013).
- 지시면: 템플릿 `internal/template/templates/AGENTS.md.tmpl`의 `audit-verdict-file` 행과, 방출기 매니페스트 `internal/template/agentemit/agents-codex.yaml`의 `codex_role_addenda`(현재 `plan-auditor`·`sync-auditor` 둘, `:492-510`; read-only 역할 넷으로 넓힌다). 역할 TOML은 `make agents-emit`로만 다시 만든다. `.codex/agents/moai/*.toml`을 손으로 고치지 않는다.
- Template-First: 템플릿을 먼저 고치고 `make build`. 템플릿에는 SPEC ID, 카드 번호, 날짜를 넣지 않는다(AC-CAR-008).
- `.claude/agents/moai`(C1, C2)는 건드리지 않는다.

### D8. 프로세스 정리와 이식성

시간 한도 초과 시 감사 프로세스와 자식을 정리하는 코드는 OS별로 다르다. 기존 쌍 `internal/cli/mcp_claude_process_unix.go`(`configureClaudeAuditProcess`, `runClaudeAuditProcess`)와 `internal/cli/mcp_claude_process_windows.go`를 재사용하거나, 같은 빌드 태그 분리 방식으로 codex용 쌍을 옆에 둔다. M2 검증에 `GOOS=windows GOARCH=amd64 go build ./...`와 `GOOS=windows GOARCH=amd64 go vet ./internal/cli/...`를 넣는다.

### D9. 이어받은 AC의 실현 (B1 잠정 기본값)

이어받은 AC-DHR-012/023의 문구와 판정식은 바이트 그대로 두고, 경로 (i)에서 테스트가 launcher를 직접 부르는 것으로 읽는다(`acceptance.md` §B 결정 문단, `spec.md` §B 대응 문단). 호출 수는 14로 남는다. 판정식이 실행 경로를 보지 않으므로, 같은 증거 파일에 `route`, `used_spawn_agent`, `session_sandbox`, `probe_command_executed`, `probe_exit_code`(ac012), `route`, `verdict_writer`(ac023)를 더한다. 이 필드는 테스트가 선언하지 않고 세션 기록과 launch record에서 유도한다(`acceptance.md` §A "경로 필드의 유도"). 유도 함수 하나를 AC-CAR-012a(t1100 m8-sbx 기록 fixture, 결정적)와 AC-CAR-012b(이어받은 실행, LIVE)가 함께 쓴다. AC-CAR-012a가 m8 run2 기록 — read-only 부모 + `spawn_agent`이며 쓰기가 실제로 거부된 경로 — 에서 `route == "spawn_agent"`, `used_spawn_agent == true`를 유도함을 고정하므로, 이 조합은 AC-CAR-012b를 통과할 수 없다. M4에서 이어받은 테스트의 `codexSubagentFor` 조회를 launcher 최상위 rollout 조회로 다시 쓴다.

## §C 마일스톤 (우선순위 순, 앞 단계가 끝나야 다음 단계)

| M | 우선순위 | 내용 | 산출 |
|---|---|---|---|
| M1 | High | 실행 경로와 인자 계약 측정(LIVE, 상한 5). P-A: 최상위 `codex exec -s read-only`에 프로젝트 config `sandbox_mode="workspace-write"`, 두 층 MCP 서버(기록 래퍼)를 둔 상태로 시험용 지시문 nonce 회신 + shell 쓰기 명령 실행 + MCP 비활성화 인자 확인(두 래퍼 기동 0) + `--json`/`-o` 반환문 비교(1회). P-B: 부모 `-s workspace-write` 세션이 shell로 read-only 자식을 띄워 모델에 닿는지(2회). P-C: P-B가 모델에 닿지 못할 때만 MCP 경로(2회). P-A는 0.156.1 세션 기록의 탐침 모양(`custom_tool_call` `exec` + `custom_tool_call_output`의 이스케이프된 `exit_code`, acceptance §A)을 최상위 read-only 세션에서 다시 확인하고, P-B는 MCP를 켠 부모 세션의 래퍼 기동 수(즉시 기동인가)를 기록한다. P-C는 MCP 서버 프로세스의 시작 디렉터리를 기록한다(D5.1). 결과를 `route.txt`에 적는다. `none`이면 여기서 멈춘다(D1) | `.moai/reports/t1143/m1-route/`(`route.txt`, 인자, 세션 기록 요약, 호출 원장), progress.md §E.2 |
| M2 | High | launcher 핵심 함수와 `moai codex audit` 동사: 허용 목록 인자, 역할 자격, 원문 쓰기, 루트·목적지 한정, 실패 경로, 상한, launch record, 프로세스 정리(D8). Windows 빌드·vet 검사 | AC-CAR-001 ~ 006, 014 |
| M3 | Medium | 지시면: `AGENTS.md.tmpl` 행, `agents-codex.yaml` 부록(넷), `make agents-emit`, 기존 `audit_role_exception_test.go`·`golden_test.go`의 문구 기대값 갱신. R2 채택 시 MCP 도구, 비동기 작업 형태, `moai-mcp-tools.md` 도구 수 갱신 | AC-CAR-007, 008, 013 |
| M4 | Medium | 유도 함수와 fixture: t1100 m8-sbx 기록 넷을 `internal/cli/testdata/codex-rollouts-m8/`로 그대로 복사하고 sha256 목록(acceptance AC-CAR-012a)과 대조, 합성 fixture 둘 추가, `TestCodexAuditEvidenceDerivation`(LIVE 없음). 그다음 이어받은 LIVE 테스트의 (ii)를 launcher로 바꾸고(`codexSubagentFor` 조회 재작성, `denied`와 경로 필드를 유도 함수로 계산, 세션 기록과 launch record 복사) AC-DHR-012/023 실행(14회) | AC-CAR-012a, AC-DHR-012, 023, AC-CAR-009, 012b |
| M5 | Medium | AC-CAR-010(3회), AC-CAR-011(2회) LIVE, 판정서 `.moai/reports/t1143/verdict.md` | AC-CAR-010, 011 |

## §D LIVE 호출 상한 (REQ-CAR-010, B5 잠정 기본값)

단위는 `codex exec` 프로세스 하나다. 부모 세션 안에서 뜬 자식 `codex exec`도 하나로 센다. `claude -p` 호출은 없다.

| 항목 | 예정 | 상한 | 상한 초과 시 |
|---|---|---|---|
| M1 경로 측정 (P-A 1, P-B 2, 필요 시 P-C 2) | 3 | 5 | 6번째를 시작하지 않고 `ABORTED`, 리드에게 반환. P-B·P-C 모두 실패하면 `none`으로 멈춤 |
| AC-DHR-012 + AC-DHR-023 + AC-CAR-012 (같은 실행) | 14 | 실행당 14. 재실행은 리드가 `INVALID`로 기록한 실행 뒤 1회만 → 최대 28 | 15번째를 시작하지 않고 `ABORTED`(출처 판정식의 의미 그대로) |
| AC-CAR-010 | 3 | 실행당 3. 같은 `INVALID` 재실행 1회 → 최대 6 | 실행 안에서 4번째 `ABORTED` |
| AC-CAR-011 | 2 | 실행당 2. 같은 `INVALID` 재실행 1회 → 최대 4 | 실행 안에서 3번째 `ABORTED` |
| **합계** | **22** | **43** | 합계 43에 닿으면 남은 LIVE를 모두 멈추고 리드에게 반환 |

- 재실행은 AC별로 한 번이며, 리드가 이전 실행을 하네스 결함으로 `INVALID` 기록했을 때만 허용된다. 재실행 사유가 모델 행동(예: 명령을 실행하지 않음)이면 `INVALID`가 아니라 FAIL이다.
- 인증 실패, 네트워크 실패로 끝난 호출도 사용한 호출로 센다.
- `ABORTED`, `NOT_RUN`, `INVALID`는 PASS가 아니다. `INVALID` 실행의 증거는 지우지 않고 이름을 바꿔 보존한다(t1100 AC-018 선례).
- LIVE 실행 전후 `CODEX_HOME` 로그인 파일의 sha256을 기록한다(t1100 M8 선례). 잔존 `codex exec` 프로세스가 없음을 마감 시점에 확인한다.
- 모든 LIVE 실행은 격리된 임시 `CODEX_HOME`·`MOAI_HOME`·저장소에서 한다. 실제 프로젝트 트리에 탐침 파일을 쓰지 않는다.

## §E 위험

| 위험 | 영향 | 대응 |
|---|---|---|
| R1 중첩 실행이 네트워크 차단(`network_access: false`)으로 모델에 닿지 못함 | R1 불가 | M1 P-B로 먼저 확인, R2로 전환 |
| R2에서 sandbox 안의 부모가 작업 루트를 임의로 지정 | sandbox 밖 쓰기 | 등록된 워크트리 + `.moai/reports/` 한정(D5.1, AC-CAR-005, AC-CAR-013) |
| R2에서 MCP 도구 시간 한도가 감사 시간보다 짧음 | 감사 중단 | 비동기 작업 형태(D1). M1 P-C에서 소요 시간 기록 |
| `-s` 플래그가 config `sandbox_mode`에 지는 경우 | read-only 보장 붕괴 | AC-CAR-010 (a)가 config `workspace-write` 조건에서 잰다. 지면 `--ignore-user-config` 등 대안을 판정서에 적고 리드에게 반환 |
| 모든 층의 MCP를 끄는 인자가 없음 | REQ-CAR-007 불충족 | D5의 개정 경로 |
| `-c developer_instructions`가 최상위 세션에서 개발자 지시로 들어가지 않음 | 역할 지시 누락 | M1 nonce 측정. 실패하면 stdin 앞머리 전달 |
| 모델이 탐침 명령을 실행하지 않고 거부를 보고 | 거짓 PASS | `write_denied` 정의가 exec 이벤트를 요구(acceptance §A) |
| 기록 래퍼가 codex가 해석하는 명령 경로에 없음 | MCP 0회가 무의미 | (b) 부모 세션의 기동 ≥1을 양성 대조로 요구(AC-CAR-010) |
| codex가 MCP 서버를 지연 기동함 | 양성 대조가 정직한 실행에서도 0 | M1 P-B가 즉시 기동 여부를 기록하고, 지연 기동이면 (b) 부모 작업에 읽기 전용 도구 호출 한 번씩을 넣는다(AC-CAR-010) |
| 0.156.1 세션 기록 모양이 가정과 다름 | 탐침 판정이 정직한 실행에서 거짓 | acceptance §A가 실제 모양을 이름으로 적고, M1 P-A가 최상위 세션에서 다시 확인한 뒤 M4를 시작한다 |
| 테스트가 경로 필드를 선언만 함 | spawn 경로가 통과 | 필드를 기록에서 유도하고, AC-CAR-012a가 m8 run2 기록으로 유도 결과를 고정한다 |
| 감사자가 읽는 저장소 내용의 프롬프트 주입 | 판정문 오염 | 반환문은 데이터로만 쓰고 실행하지 않으며, 목적지는 모델 출력에서 받지 않는다(REQ-CAR-005). 판정 내용의 진위는 범위 밖 |
| 이어받은 판정식이 `.moai/reports/t1100/` 경로를 씀 | 이름 혼동 | 워크트리 안에서만 실행. 판정서에 t1143 카드 소속임을 적는다 |
| 이어받은 (ii) 문구와 실현 방식의 차이 | 판정 해석 분쟁 | §B 결정 문단 + AC-CAR-012, §G B1로 운영자 확인 |
| codex 판올림으로 플래그 우선순위나 `-c` 처리가 바뀜 | 보장 붕괴 | LIVE는 CI가 게이트하지 않는다. 판올림 때 AC-CAR-010을 다시 돌려야 한다(잔여 위험) |

## §F 바뀔 파일 (예상)

- 신규: `internal/cli/codex_audit_launch.go`, `internal/cli/codex_audit_launch_test.go`, LIVE 테스트 파일 1개, 필요 시 codex용 OS별 프로세스 도우미 쌍(D8), 유도 함수 fixture `internal/cli/testdata/codex-rollouts-m8/`(t1100 m8-sbx 기록 넷 + 합성 둘, 로컬 임시 경로가 들어 있으나 템플릿이 아닌 testdata라 중립성 규칙 밖), 측정 경로 사본 `internal/template/agentemit/testdata/measured-route.txt`
- 변경: `internal/cli/codex_launcher.go`(동사 등록), `internal/cli/codex_role_live_test.go`((ii) 경로, 증거 필드), `internal/config/defaults.go`(시간 한도 상수), `internal/template/templates/AGENTS.md.tmpl`, `internal/template/agentemit/agents-codex.yaml`, `internal/template/agentemit/audit_role_exception_test.go`, `internal/template/agentemit/golden_test.go`
- 재생성: `internal/template/templates/.codex/agents/moai/{mission-governor,plan-auditor,super-advisor,sync-auditor}.toml`
- R2 채택 시: `internal/cli/mcp_*.go` 1개, `.claude/rules/moai/core/moai-mcp-tools.md`와 템플릿 미러

## §G 운영자 확인 항목 (잠정 기본값 적용 — Implementation Kickoff에서 운영자가 확인)

운영자가 plan 단계에서 답하지 않아, 리드가 전달한 잠정 기본값을 적용했다. 다섯 항목 모두 **잠정 기본값이며 Implementation Kickoff에서 운영자가 확인한다.** Kickoff 게이트는 이 항목들을 다시 묻는다.

**거부의 결과.** Kickoff에서 어느 B 항목이든 거부하는 답이 나오면 이 plan-audit 판정은 무효가 된다. manager-spec이 영향받는 REQ·AC를 개정하고, 승인 전에 plan-audit을 다시 돌린다. B 항목 다섯은 AskUserQuestion 한 번의 질문 수 한도(4)를 넘으므로, 승인 질문보다 앞선 별도 라운드에서 묻는다.

- **B1 — 이어받은 AC의 해석.** 이 해석은 측정이 강제한 것이라 다른 해석을 고르는 선택지는 없다(`spec.md` §B). 운영자가 고르는 것은 둘 중 하나다: 이 해석을 확인하거나, AC-DHR-012/023을 `NOT_RUN`(이월)으로 남긴다. 잠정 기본값은 확인이다. 문구와 판정식은 바이트 그대로 두고 다시 쓰지 않는다. 경로 (i)에서 테스트가 launcher를 직접 부르는 것으로 읽으며 호출 수 14가 그 해석이다(D9). 경로 필드를 기록에서 유도하고 AC-CAR-012a/b로 판정해 `spawn_agent` + read-only 부모 조합이 카드를 통과시키지 못하게 한다. M4에서 `codexSubagentFor` 조회를 다시 쓴다.
- **B2 — MCP 비활성화.** 잠정 기본값: 수용. 감사 프로세스에서 모든 층의 MCP 서버를 끈다(D5). Codex 경로의 `plan-auditor` 교차 모델 감사와 `super-advisor` 위임 도구가 빠진다.
- **B3 — 로컬 지시 파일 제외.** 잠정 기본값: 수용. 감사 프로세스 지시문에 로컬 지시 파일을 합치지 않는다(D4).
- **B4 — R1·R2 모두 실패 시.** 잠정 기본값: 카드를 멈추고, 더 이상 LIVE를 하지 않으며, 리드에게 보고한다(D1).
- **B5 — LIVE 상한.** 잠정 기본값: 절대 상한 43(예정 22). 종전 38에 AC-CAR-010(+3)과 AC-CAR-011(+2)의 `INVALID` 재실행 1회씩을 더했다(§D).
