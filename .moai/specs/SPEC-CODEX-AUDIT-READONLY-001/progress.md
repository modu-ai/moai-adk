---
id: SPEC-CODEX-AUDIT-READONLY-001
document: progress
card: t1143
---

# Progress — SPEC-CODEX-AUDIT-READONLY-001

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: spec.md, plan.md, acceptance.md, progress.md (Tier M)
- 이어받은 항목 3개(REQ-DHR-015 런타임 조항, AC-DHR-012, AC-DHR-023)는 이관 커밋 `de5faa77a`의 원문을 바이트 그대로 옮겼다(AC-CAR-009가 판정).
- 운영자 확인 항목: plan.md §G B1~B5. 다섯 항목 모두 리드가 전달한 잠정 기본값이 적용되어 있고, Implementation Kickoff에서 운영자가 확인한다. 미해결 확인 표시는 0개다(0.2.0).
- plan-audit: iter-1 FAIL 0.77(`.moai/reports/t1143/plan-audit-iter1.md`). 0.2.0에서 D1~D13 반영.
- plan-audit: iter-2 FAIL 0.87(`.moai/reports/t1143/plan-audit-iter2.md`). 0.3.0에서 N1~N9 반영.
- 판정식 빈 입력 확인(0.3.0): 결정적·LIVE 판정식마다 빈 입력에서 `true`가 나오지 않음을 실행했다. 명령·출력: `.moai/reports/t1143/plan-checks/empty-input.md`. 판정식 원문은 `acceptance.md`에 있으므로 같은 방식으로 다시 돌릴 수 있다.
- 경로 필드 유도의 참조 실행(0.3.0): `.moai/reports/t1143/plan-checks/derivation-m8.md`(참조 구현 `derive.jq` 동봉, 입력은 primary checkout `.moai/reports/t1100/m8-sbx/`).
- LIVE 호출: plan 단계 0회.

## §E.2 Run-phase Evidence

Kickoff: 운영자, 레인 창 직접 승인 (2026-09-24) — B1–B5 plan §G 기본값 확정.

### M1 — 실행 경로와 인자 계약 측정 (LIVE 5/5, 2026-09-24)

결과: `.moai/reports/t1143/m1-route/route.txt` = `mcp`. 호출 원장은 `.moai/reports/t1143/verdict.md` `## LIVE ledger`, 호출별 인자는 `m1-route/argv.txt`, 세션 기록 요약(판정 필드만, sha256 포함)은 `m1-route/rollout-summaries.jsonl`에 있다. codex-cli 0.156.1.

측정 환경: 추적 트리 밖의 임시 저장소(`git init`, 커밋 없음). 프로젝트 층 `.codex/config.toml`에 `sandbox_mode = "workspace-write"`, 층 적재 표지 `model_reasoning_summary = "detailed"`, 기록 래퍼 MCP 서버 `projdecoy`. `CODEX_HOME`은 `~/.t1143-m1-codexhome`로, 어떤 sandbox 쓰기 루트에도 속하지 않는 위치에 두었다(실제 레인의 `~/.codex`와 같은 조건). 그 안의 `auth.json`은 `~/.codex/auth.json`을 가리키는 심볼릭 링크이며 사본을 만들지 않았다. 사용자 층에는 기록 래퍼 MCP 서버 `userdecoy`를 두었다.

| 질문 | 명령 | 관측 출력 | 결론 |
|---|---|---|---|
| MCP 비활성화 인자 (LIVE 전 설정 해석) | `codex mcp list --json` (모델 호출 없음) | 인자 없음: 두 서버 `enabled`. `-c mcp_servers.userdecoy.enabled=false -c mcp_servers.projdecoy.enabled=false`: 둘 다 `disabled`. `-c 'mcp_servers={}'`: 둘 다 `enabled` | 이름별 `enabled=false`만 두 층을 끈다. 빈 테이블 재정의는 병합되어 효과가 없다 |
| `-s`가 config를 이기는가 (P-A, #1) | `codex exec -s read-only -c approval_policy=never -c model_reasoning_effort=low -c 'developer_instructions="…NONCE-1655733724e4…"' -c mcp_servers.userdecoy.enabled=false -c mcp_servers.projdecoy.enabled=false -C <root> --json -o <tmp> -` | `turn_context.sandbox_policy` = `{"type":"read-only"}`, `summary` = `detailed`(프로젝트 층에서만 오는 값. 같은 버전의 t1100 m8-sbx 기록은 `none`) | 프로젝트 층이 적재된 상태에서 `-s read-only`가 이겼다 |
| 쓰기 탐침 모양 (P-A) | 위와 같음 | `custom_tool_call` `name "exec"`, `input` `text(await tools.exec_command({cmd:"printf '%s' probe > probe-pa.txt",max_output_tokens:1000}));`, 같은 `call_id`의 `custom_tool_call_output` 텍스트 `{"chunk_id":"6bd49c",…,"exit_code":1,…,"output":"zsh:1: operation not permitted: probe-pa.txt\n"}`. 실행 뒤 루트에는 `.codex`, `.git`만 있다 | acceptance §A의 0.156.1 모양이 최상위 read-only 세션에서 확인됐다. M4 전 정의 개정 불필요 |
| 지시문 전달 (P-A) | 위와 같음 | 세션 기록의 developer 메시지에 nonce 존재, 최종 메시지 `NONCE=NONCE-1655733724e4 WRITE=denied` | `-c developer_instructions=`가 최상위 세션에 개발자 지시로 들어간다 |
| MCP 기동 수 (P-A) | 위와 같음 + `m1-route/mcp-launch.log` | 호출 #1 동안 기록 0줄(사용자 0, 프로젝트 0) | 이름별 `enabled=false`로 두 층 모두 0회. 양성 대조는 P-B |
| `--json` 대 `-o` (P-A) | `cmp pa-json-last.txt pa-last.txt` | 둘 다 37바이트, sha256 `5d1064c12e528846e8110851125dcd405dc18a3825fb2d8e3120e0bce8a4a870`, 동일 | 한 줄 메시지에서 같은 바이트 |
| 중첩 exec가 모델에 닿는가 (P-B, #2·#3) | 부모 `codex exec -s workspace-write …`(MCP 켬)가 shell로 자식 `timeout 300 codex exec -s read-only …` 실행 | 부모 `sandbox_policy` `workspace-write`, `network_access: false`. 자식 종료 코드 1(0.3초), `child.jsonl` 비어 있음, 세션 기록 없음, 표준 오류 `Error: failed to initialize in-process app-server client: Operation not permitted (os error 1)` | shell 경로(R1)는 모델에 닿지 못한다 |
| MCP 즉시/지연 기동 (P-B) | `m1-route/mcp-launch.log` | 08:16:58Z 사용자·프로젝트 각 1회(세션 시작 약 1초 뒤, 도구 호출 전), 08:17:12Z 각 1회 더. 모두 `ppid` = 부모 codex, `cwd` = 세션 루트 | 즉시 기동. 세션 중 재기동 1회(원인 미측정) |
| MCP 경로 (P-C, #4·#5) | 부모 `codex exec -s workspace-write … -c mcp_servers.userdecoy.enabled=false -c mcp_servers.projdecoy.enabled=false …`가 프로젝트 층 시험 MCP 서버 `routeprobe`의 도구를 한 번 호출, 서버가 자식 `codex exec -s read-only …` 실행 | `routeprobe.log`: 서버 시작 `cwd` = 세션 루트(`-C` 디렉터리), 자식 종료 코드 0, 5.0초, 최종 메시지 `CHILD-PONG-PC`. 자식 기록 `sandbox_policy` `read-only`. 자식은 부모 sandbox가 쓸 수 없는 `CODEX_HOME`에 세션 기록을 썼다. P-C 동안 decoy 기동 0 | MCP 경로(R2)는 모델에 닿는다. MCP 서버 프로세스는 부모 shell sandbox 밖에서 돈다 |

LIVE 사용: 5/5(M1 상한 도달, 6번째 없음). `INVALID` 0. 로그인 파일 sha256은 5회 모두 전후 `eb7c45bd906f27ced1154ff2beba819c360c5b498fe4752baf754b7637ef6630`으로 같았다. 마감 시 `ps -axo pid,command | grep -c "[c]odex exec"` → `0`.

Gaps (M1):

- P-B 자식이 실패한 원인(쓸 수 없는 `CODEX_HOME`, 중첩 sandbox, 네트워크 차단)을 분리하지 않았다. 어느 것이든 실제 레인 조건에서 R1은 불가다.
- `--ignore-user-config`는 LIVE로 재지 않았다. 사용자 층의 프로젝트 trust 항목까지 버리므로 프로젝트 층 적재 여부가 달라질 수 있다.
- 이름별 `enabled=false`는 launcher가 선언된 서버 이름을 알아야 한다. 이름을 얻는 방법(`codex mcp list --json` 추가 호출 또는 두 TOML 직접 판독)은 재지 않았다. `mcp_servers` 이외의 층(시스템·관리 설정)은 픽스처에 없었다.
- `--json`과 `-o`의 동일성은 37바이트 한 줄 메시지에서만 쟀다. 여러 줄·끝 개행 메시지는 재지 않았다.
- MCP 서버 시작 디렉터리는 픽스처 세션 루트로 쟀다. 실제 레인 워크트리에서 moai MCP 서버로 잰 것이 아니다. MCP 도구 기본 시간 한도는 재지 않았다(픽스처는 `tool_timeout_sec = 400`).
- P-B의 세션 중 MCP 재기동 원인은 재지 않았다.

### M2 — 첫 항목: lane 워크트리의 moai MCP 서버 시작 디렉터리 (LIVE 1/1, 누계 6/43)

명령: `PATH=<scratch>/m2/bin:<시스템 PATH> CODEX_HOME=<fixture> timeout 300 codex exec -s read-only -c approval_policy=never -c model_reasoning_effort=low -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1143 --json -o <scratch>/m2/out/last.txt - < prompt.txt`. 프로젝트 층 `[mcp_servers.moai]`의 `command = "moai"`가 PATH 맨 앞의 통과형 기록 래퍼로 풀리고, 래퍼는 자기 `pwd -P`와 `lsof -a -p $$ -d cwd`를 기록한 뒤 `/Users/goos/go/bin/moai`로 `exec`한다(같은 pid가 `moai mcp-server`가 된다). 세션 sandbox는 `read-only`로 두어 워크트리에 쓰지 않게 했다. `CODEX_HOME`은 격리 사본(인증 파일은 심볼릭 링크)이며 측정 뒤 지웠다.

관측(`.moai/reports/t1143/m2-startdir/moai-launch.log`, `rollout-summary.json`):

```
2026-09-24T08:29:02Z pid=24313 ppid=24149 pwd=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1143 args=mcp-server
p24313
fcwd
n/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1143
```

종료 코드 0, 16초. 세션이 부른 `spec_progress`의 결과 `_root` = `{"dir":"/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1143","source":"ser…"}`(모델이 160자에서 잘라 반환. `ser`로 시작하는 원천 이름은 `internal/cli/mcp_project_root.go`의 `server-cwd`뿐). 로그인 파일 sha256 전후 `eb7c45bd…ef6630` 동일. 결론: moai MCP 서버는 이 lane 워크트리에서 시작한다. D5.1의 R2 전제가 성립하므로 `route.txt`는 `mcp`로 둔다.

### M2 — launcher 핵심 함수와 `moai codex audit` 동사 (TDD, LIVE 없음)

새 파일 `internal/cli/codex_audit_launch.go`, `internal/cli/codex_audit_launch_test.go`(`!windows`, 가짜 `codex` 셸 스크립트), 변경 `internal/config/defaults.go`(`DefaultCodexAuditTimeout` 20분, `DefaultCodexAuditListTimeout` 30초).

- 인자: `codex mcp list --json` 한 번 → `codex exec -s read-only -c approval_policy="never" -c model_reasoning_effort="<역할 값>" -c developer_instructions=<JSON> -c mcp_servers.<name>.enabled=false … -C <root> --json -`(작업 문은 stdin). 서버 이름이 `^[A-Za-z0-9_-]+$` 밖이면 이름으로 끌 수 없으므로 실행하지 않는다.
- 역할 자격: 방출된 역할 이름 중 `agentemit` 매니페스트 권한 계약의 sandbox가 `read-only`인 것. 역할 파일이 없거나 파일의 `sandbox_mode`가 `read-only`가 아니면 거부한다.
- 루트: 심볼릭 링크를 푼 뒤 호출자 워크트리의 최상위와 같고, 같은 git common dir이며, `git worktree list --porcelain`에 등록돼 있어야 한다. 목적지: 루트의 `.moai/reports/` 아래, `codex-audit/` 밖, `.git` 성분 없음(어휘·해석 양쪽).
- 순서: 루트 → 목적지 → 역할 → 상한(`checkCodexInstructionSize`, 최종 토큰 길이) → codex 위치 → `mcp list` → launch record 배타 예약(`O_EXCL`) → `exec`(프로세스 그룹, `configureClaudeAuditProcess`/`runClaudeAuditProcess` 재사용) → 판정 파일 원자적 쓰기 → record 기록과 `LAUNCH_RECORD <상대 경로>`.

RED(구현 전): `go test ./internal/cli -run 'CodexAudit' -count=1` → 컴파일 실패 `undefined: codexAuditResult` 등(`.moai/reports/t1143/run/m2-red.log`), 이어 stub 상태에서 9개 테스트 모두 `--- FAIL`(`m2-red-stub.log`).

판정식(acceptance 원문 그대로 실행, `.moai/reports/t1143/run/m2-judges.log`):

| AC | 판정 출력 | 변이 → 판정 출력 |
|---|---|---|
| AC-CAR-001 | `true` | m001 이름 하나 비활성화 누락 → `false` |
| AC-CAR-002 | `true` | m002 계약 자격 검사 제거 → `false`; m002b 역할 파일 sandbox 검사 제거 → `false` |
| AC-CAR-003 | `true` | m003 반환문 trim → `false`; m003b 이름 바꾸기 전 부분 쓰기 → `false` |
| AC-CAR-004 | `true` | m004 0 아닌 종료를 성공 처리 → `false`; m004b 프로세스 그룹 설정 제거 → `false` |
| AC-CAR-005 | `true` | m005 호출자 워크트리 검사 제거 → `false`; m005b `codex-audit` 제외 제거 → `false`; m005c `.git` 검사 제거 → `false` |
| AC-CAR-006 | `true` | m006 상한을 원문 길이로 측정 → `false` |
| AC-CAR-014 | `true` | m014 지시문 미가림 → `false`; m014b `O_EXCL` 제거 → `false` |

m002는 처음 `true`로 살아남았다(역할 파일 sandbox 검사가 같은 사례를 막음 — 충분한 방어 둘). 역할 파일을 `read-only`로 고친 쓰기 역할 사례를 테스트에 더한 뒤 `false`가 되었다(`m2-mutants.log`).

검증(모두 이 트리, 이 실행):

| 명령 | 종료 코드 | 로그 |
|---|---|---|
| `go test ./internal/cli -run 'Codex' -count=1`(kanban·factory 변수 제거, `MOAI_FACTORY_WORKER` 포함) | 0, `ok … 160.387s` | `run/m2-test-cli-codex.log` |
| `go test ./internal/cli -run 'Guard\|Scan\|Neutral\|Hardcod\|Registration\|Subcommand\|AskUser\|Literal\|Sweep' -count=1` | 0 | `run/m2-test-cli-guards.log` |
| `go test ./internal/config/... -count=1` | 0 | `run/m2-test-config.log` |
| `go vet ./internal/cli/...` | 0 | `run/m2-vet.log` |
| `GOOS=windows GOARCH=amd64 go build ./...` | 0 | `run/m2-windows-build.log` |
| `GOOS=windows GOARCH=amd64 go vet ./internal/cli/...` | 0 | `run/m2-windows-vet.log` |
| `golangci-lint run ./internal/cli/... ./internal/config/...` | 0, `0 issues.` | `run/m2-lint.log` |
| `go test ./internal/cli -run 'CodexAudit' -coverprofile` → `codex_audit_launch.go` 문장 333 중 284 | 85.3% | `run/m2-cover-func.txt` |

Gaps (M2):

- 첫 `-run 'Codex'` 실행은 `TestCodexSpawn_RealAssemblyThroughStubTmux` 한 건이 실패했다. 기대 문자열에 없는 `MOAI_FACTORY_WORKER=agent-44`가 레인 환경에서 들어왔기 때문이며, 이 변수까지 지운 재실행은 통과했다. 같은 실패가 develop에서도 나는지는 재지 않았다.
- 결정적 테스트 파일은 `!windows`다(가짜 codex가 POSIX 셸). Windows에서는 컴파일·vet만 확인했고 동작은 확인하지 않았다.
- 판정 파일 쓰기는 성공했는데 launch record 쓰기가 실패하면 launcher는 0이 아닌 코드로 끝나지만 판정 파일은 남는다. 이 경로는 테스트하지 않았다.
- MCP 경로의 겉면(도구, 비동기 작업)은 M3 범위이며 없다. `route`는 `direct`(테스트)와 `shell`(동사)만 기록된다.
- 변이 실행 m004b에서 가짜 codex의 `sleep 30` 자식이 남았고, 정리하면서 다른 세션의 폴링 루프에 속한 `sleep 30` 하나(pid 30259)까지 종료했다. 그 루프는 계속 돌았다(새 `sleep` 확인). 이후 테스트 하위 항목마다 정리 훅을 등록했다.

### M3 — MCP 경로와 지시면 (LIVE 없음, 커밋 `6781ce995`)

`route.txt`가 `mcp`이므로 plan §C M3의 R2 분기를 구현했다.

- MCP 도구 셋: `codex_role_audit`(시작, 쓰기 가능), `codex_role_audit_status`, `codex_role_audit_result`(읽기 전용). 시작 도구는 `moai codex audit`와 같은 핵심 함수를 부른다. 핵심 함수를 `prepareCodexAudit`(모든 거부 + launch record 예약)와 `run`(프로세스·판정 파일·record)으로 나눴고 동사도 같은 둘을 쓴다. 거부는 작업이 생기기 전에 일어나며 아무것도 쓰지 않는다. 받아들인 호출은 곧바로 작업 id를 돌려주고 감사는 서버 프로세스 안의 goroutine에서 돈다(호스트의 도구 시간 한도가 감사를 끊지 않게). `worktree_root`는 필수 입력이고, 핵심 함수의 호출자 워크트리는 서버가 시작된 디렉터리(M2 첫 항목에서 lane 워크트리로 잰 값)로 정한다. 입력 이름은 `project_root`가 아니라 `worktree_root`다(선택 입력인 `project_root` 계열과 뜻이 달라 규칙 문서의 `project_root` 목록에 섞지 않았다).
- 지시면: 템플릿 `AGENTS.md.tmpl`의 `audit-verdict-file` 행과 `agents-codex.yaml`의 read-only 역할 넷 부록이 MCP 도구 `codex_role_audit`만 이름으로 적고, `spawn_agent`로 띄우지 말라고 적으며, launcher가 반환문 그대로 파일을 쓴다고 적는다. 셸 동사 이름은 지시면에 없다. 역할 TOML 넷은 `make agents-emit`으로만 재생성했다. 측정 경로 사본 `internal/template/agentemit/testdata/measured-route.txt`를 더했다.
- 목록 수 36 → 39(쓰기 15, 읽기 24): 규칙 `moai-mcp-tools.md`와 `moai-mcp-tools-catalogue.md`(로컬과 템플릿 미러, 바이트 동일), `internal/mcp/catalog.go`·`catalog_test.go`, `factory_lane_handoff_compat_test.go`의 36/14/22 → 39/15/24, 웹 콘솔 `i18n.js` 네 로케일 키와 `codex_panel_test.go` 고정 목록.
- 기존 기대 테스트: `audit_role_exception_test.go`(부모 쓰기 표지를 "the launcher writes …"로), `golden_test.go` 주석.

RED(구현 전): `go test ./internal/template/agentemit -run TestAuditRoleLauncherInstructionSurface` → 16건 실패(`run/m3-red-agentemit.log`), `go test ./internal/cli -run TestCodexAuditMCPTool` → 컴파일 실패 `undefined: codexRoleAuditToolName`(`run/m3-red-mcp.log`).

판정식(acceptance 원문, 커밋 `6781ce995` 트리, `run/m3-judges.log`):

| AC | 판정 출력 | 변이 → 판정 출력 |
|---|---|---|
| AC-CAR-007 | `true` | m007 행에 `moai codex audit`를 함께 적음 → `false` |
| AC-CAR-008 | `true` | m008 추가 템플릿 줄에 카드 번호 → `false`(판정식의 `develop...HEAD`를 작업 트리 비교 `develop`로 바꾼 변형; 같은 변형이 복원한 트리에서 `true`) |
| AC-CAR-013 | `true` | m013 MCP 도구가 `worktree_root`를 호출자 워크트리로 믿음 → `false` |

검증(이 트리, 이 실행):

| 명령 | 종료 코드 | 로그 |
|---|---|---|
| `make agents-emit` | 0 | `run/m3-agents-emit.log` |
| `make build` | 0 | `run/m3-make-build.log` |
| `make -s agents-emit-check` | 0 | `run/m3-agents-emit-check.log` |
| `go test ./internal/template/...` | 0 | `run/m3-test-template.log` |
| `go test ./internal/template/agentemit/...` | 0 | `run/m3-test-agentemit.log` |
| `go test ./internal/mcp/...` | 0 | `run/m3-test-mcp.log` |
| `go test ./internal/web/...`(목록 추가 전 3건 실패 → 추가 후) | 0 | `run/m3-test-web.log` |
| `go test ./internal/cli -run 'Codex\|MCP\|…\|Sweep'`(kanban·factory 변수 제거) | 0 | `run/m3-test-cli.log` |
| `go vet ./internal/cli/... ./internal/mcp/... ./internal/web/... ./internal/template/...` | 0 | `run/m3-vet.log` |
| `GOOS=windows GOARCH=amd64 go build ./...` / `go vet ./internal/cli/...` | 0 / 0 | `run/m3-windows-build.log`, `run/m3-windows-vet.log` |
| `golangci-lint run` (위 네 패키지 트리) | 0, `0 issues.` | `run/m3-lint.log` |

Gaps (M3):

- Codex가 쓰기 가능 표시(read-only 힌트 false)인 MCP 도구 호출을 `default_tools_approval_mode = "writes"`와 부모 승인 정책 아래서 어떻게 다루는지 재지 않았다. M1 P-C의 시험 도구는 read-only 표시였다. 레인에서 `codex_role_audit` 호출이 승인 요청에서 멈추거나 거부될 수 있으며 M5 LIVE가 처음 재는 곳이다.
- 작업 표는 서버 프로세스 메모리에만 있다. 서버가 끝나면 작업 id는 사라지며, 이미 시작한 감사 프로세스는 자기 시간 한도까지 돈다(프로세스 그룹이 서버와 분리되어 있음). 판정 파일과 launch record는 그대로 남는다.
- AC-CAR-008의 변이는 커밋 범위(`develop...HEAD`)가 아니라 작업 트리 비교로 보였다. 커밋된 변이로 원문 판정식을 돌리지는 않았다.
- `factory_lane_handoff_compat_test.go`(다른 SPEC의 AC-FLH-015)와 웹 콘솔 테스트의 고정 수를 이 카드가 바꿨다.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
