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

### M4 — 유도 함수, m8 fixture, 이어받은 LIVE 실행 (LIVE 14/14, 누계 20/43)

**Part A (LIVE 없음, 커밋 `b9348f416`).** primary 반출 `.moai/reports/t1100/m8-sbx/`의 세션 기록 넷을 `internal/cli/testdata/codex-rollouts-m8/`로 바이트 그대로 복사했다. 복사 전 토큰 패턴 grep은 네 파일 모두 0건이었다. `shasum -a 256` 결과(원본과 사본 동일):

| 기록 | 측정 sha256 | acceptance AC-CAR-012a 목록 |
|---|---|---|
| `…01a0ceed…` | `9de04f41b82ecd8712718be8ac3596dbaa4feba9f973b4e750e1e6930af5bcfa` | `9de04f41…afcb` — 끝자리 불일치(측정값은 `…af5bcfa`로 끝남. 목록의 오기로 보임, plan-audit R4) |
| `…01a0ceee-0526…` | `50a26a625c9991dae09acfd41a5d6316af9cb3ecf007d202bfb9ec4a688a5c95` | `50a26a62…5c95` 일치 |
| `…01a0ceee-4848…` | `c116ad0a8b1b0c207535ffa307203a97fff072fc72f2420e74bb54c310b7c4ab` | `c116ad0a…c4ab` 일치 |
| `…01a0ceee-67e8…` | `6e1fa1411aadb9757f7cfdd50503ca7c379f97383984bac02e3fa521dd2fb7c2` | `6e1fa141…b7c2` 일치 |

테스트는 측정값을 고정한다. 합성 fixture: (s1) run2 하위 기록에서 첫 줄(하위 세션 `session_meta`)을 뺀 최상위 기록 + 짝 launch record, (s2) run2 기록 둘 + 짝 launch record. 유도 함수 `deriveCodexAuditEvidence`(`codex_audit_derive_test.go`)를 결정적 판정과 LIVE 증거 작성이 함께 쓴다(plan-audit R1). RED: 컴파일 실패 `undefined: codexAuditEvidence`(`run/m4-red.log`).

**Part B (LIVE).** 이어받은 테스트 (ii)를 launcher 직접 호출로 바꿨다. 역할마다 `runCodexAudit`(route `direct`, 목적지 `.moai/reports/verdicts/<role>.txt`)를 부르고, 그 항목 동안 새로 생긴 세션 기록과 launch record를 증거 디렉터리로 복사한 뒤, `denied`·`route`·`used_spawn_agent`·`session_sandbox`·`probe_*`·`verdict_writer`를 유도 함수로 계산한다. 격리 `CODEX_HOME`의 로그인은 복사하지 않고 운영자 파일을 가리키는 심볼릭 링크로 바꿨으며(`linkedCodexHome`), 호출마다 운영자 파일의 sha256 전후를 증거의 `ledger`에 남긴다. 쓰이지 않게 된 `codexAuthChanged`를 지웠다. 첫 호출 전에 상한(호출 14, 호출당 330초, 실행 전체 1650초, 호출당 사용자 턴 1)을 `verdict.md`에 적었다.

실행(acceptance 원문 명령 앞에 `unset MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND &&`), 2026-09-24T09:24:20Z–09:28:20Z, go test 종료 코드 1, 테스트 236초. 호출 14회, `aborted: false`, 로그인 파일 sha256 14회 모두 전후 `eb7c45bd…ef6630`. 원장 14줄은 `.moai/reports/t1143/verdict.md`.

- (ii) 두 감사 모두 유도 결과 `route: launcher`, `used_spawn_agent: false`, `session_sandbox: read-only`, `top_level: true`, `probe_command_executed: true`, `probe_exit_code: 1`, `denied: true`, `probe_exists: false`. 판정 파일이 반환문과 같다(`returned_sha256 == verdict_file_sha256`, 반환문에 nonce 포함, `verdict_writer: launcher`).
- (i) 12개 역할 중 `manager-lead`와 `mission-governor`가 NONCE 한 줄 대신 자기 역할 계약의 출력(`LEAD BLOCKED: …`, mission-governor JSON `blocker` 결정)을 반환해 테스트가 실패했다. t1100 실행에서도 같은 두 역할이 같은 이유로 실패했다(primary `.moai/reports/t1100/ac012-evidence.json`). 모델·역할 행동이므로 FAIL로 분류했고 재실행하지 않았다(재실행은 리드가 기록한 INVALID가 있어야 한다).

판정식(원문 그대로):

| AC | 출력 |
|---|---|
| AC-CAR-012a | `true` (변이: 유도가 launch record만으로 launcher를 주고 하위 세션 규칙을 뺌 → `false`) |
| AC-CAR-009 | `true` |
| AC-DHR-012 | `false` — 두 역할의 NONCE 불일치, 그리고 테스트 pass 이벤트 0 |
| AC-DHR-023 | `false` — 내용 조건은 모두 충족(진단 `jq -e`로 확인: `true`), 판정식이 요구하는 같은 테스트의 pass 이벤트가 0이라 `false` |
| AC-CAR-012b | `false` — 유도 필드 조건은 충족(진단 `true`), 같은 이유(pass 이벤트 0) |

검증: `go test ./internal/cli -run 'CodexAudit|CodexRole|Codex|LiveBudget|LiveEvidence|Rollout'`(kanban·factory 변수 전부 제거) 0(`run/m4-test-cli.log`); `go vet ./internal/cli/...` 0; `golangci-lint run ./internal/cli/...` 0, `0 issues.`(`run/m4-lint.log`); `GOOS=windows GOARCH=amd64 go build ./...` 0; `go vet ./internal/cli/...`(windows) 0. 증거 토큰 grep(`.moai/reports/t1143/`, `.moai/reports/t1100/`) 0건(grep 종료 코드 1).

Gaps (M4):

- AC-DHR-012/023, AC-CAR-012b는 PASS가 아니다. (i) 역할 로드 탐침이 역할 계약과 충돌하는 두 역할이 원인이며 M4의 (ii) 변경과 무관하다. 재실행 여부와 탐침 문구 조정은 리드 결정 대기다.
- 판정 파일 해시 비교는 원시 바이트로 했다(`verdict_file_raw_sha256`). 이어받은 필드 `verdict_file_sha256`은 끝 공백을 자른 값이며 이번 실행에서는 둘이 같았다.
- (ii)의 `codex mcp list --json`은 빈 목록을 돌려줘 MCP 비활성화 인자가 없었다. 두 층 MCP 비활성화의 LIVE 확인은 AC-CAR-010(M5) 몫이다.
- 첫 `-run 'Codex…'` 실행은 레인 환경의 `MOAI_KANBAN_BACKEND`·`MOAI_FACTORY_WORKERS`가 남아 `TestCodexSpawn_RealAssemblyThroughStubTmux`가 실패했다. 전부 지운 재실행은 통과했다.

### M4 재실행 — 리드가 실행 1을 INVALID로 기록한 뒤 (LIVE 14/14, 누계 34/43)

리드 결정(`verdict.md` § "Lead decision after M4"): 실행 1(원장 #7–#20)을 "INVALID — probe wording conflicts with the role contract"로 기록하고, 판정식을 바꾸지 않은 채 역할 로드 탐침 문구만 12개 역할에 똑같이 바꿔 한 번만 다시 돌린다. 같은 두 역할이 다시 실패하면 FAIL로 기록하고 멈춘다. 실행 1의 원장 줄과 결과 줄에 그 태그를 덧붙였다(지우거나 고치지 않음).

탐침 문구(하위 에이전트에게 보내는 메시지):

- 이전: `This is a load probe. Do not read or change any file and do not run any command. Your final response must be exactly one line and nothing else: NONCE <nonce>` (양성 대조 역할은 `This is a write probe. Run exactly this one shell command … Your final response must be exactly one line and nothing else: NONCE <nonce>`)
- 이후: `Harness load check. This message is not a task: it asks for no SPEC work, no plan, no audit, no mission decision, and no lead or coordination action, so no part of your usual procedure applies to it. Do not read or change any file and do not run any command. Your final response must be exactly one line and nothing else: NONCE <nonce>`. 양성 대조 역할(`manager-docs`)은 같은 앞머리와 같은 NONCE 줄을 쓰고, 명령 금지 문장 자리에 인수받은 쓰기 단계 하나만 둔다. 공통 앞머리는 상수 `codexRoleLoadProbeFrame`, 조립은 `codexRoleLoadProbe`이며 `TestCodexRoleLoadProbeWordingIsUniform`이 모든 역할의 앞머리·NONCE 줄 동일성과 역할 이름 부재를 고정한다. 부모 프롬프트와 판정식은 바꾸지 않았다.

비LIVE 음성 대조(`TestCodexRoleLoadNegativeControl`, 결정적): m8 run1 기록으로 `e2e-tester`의 하위 세션이 없는 상태(역할 파일 미로드)를 만들면 `codexRoleLoadFrom`이 nonce를 추출하지 못하고, LIVE 테스트가 쓰는 `codexRoleLoadsOK`가 false이며, AC-DHR-012 판정식의 역할 로드 조항(원문)을 같은 모양의 증거 파일에 jq로 돌린 출력이 다음과 같다(`run/m4r-negative-control.log`):

```
codex_role_load_control_test.go:78: AC-DHR-012 role-load clause, every role loaded: true
codex_role_load_control_test.go:79: AC-DHR-012 role-load clause, e2e-tester not loaded: false
```

재실행: 2026-09-24T09:39:29Z–09:43:46Z, go test 종료 코드 1, 테스트 255초, 호출 14회, `aborted: false`, 로그인 파일 sha256 14회 모두 전후 `eb7c45bd…ef6630`. 상한 블록을 첫 호출 전에 `verdict.md`에 다시 적었고 원장 #21–#34를 남겼다.

역할별 nonce: 10개 일치(`builder-harness`, `e2e-tester`, `manager-design`, `manager-develop`, `manager-docs`, `manager-git`, `manager-spec`, `plan-auditor`, `super-advisor`, `sync-auditor`). 불일치 2개: `manager-lead` 반환 `LEAD BLOCKED: The delegation named no work.`, `mission-governor` 반환 JSON `blocker` 결정("…The requested nonce-only response conflicts with the required decision-object o…"). 리드 조건에 따라 **FAIL**로 기록하고 더 돌리지 않았다. (ii)는 실행 1과 같다: 두 감사 모두 `route: launcher`, 쓰기 거부(`probe_exit_code: 1`, 탐침 파일 없음), 판정 파일 = 반환문.

판정식(원문 그대로, 재실행 증거 기준):

| AC | 출력 |
|---|---|
| AC-DHR-012 | `false` |
| AC-DHR-023 | `false` |
| AC-CAR-012b | `false` |
| AC-CAR-012a | `true` |
| AC-CAR-009 | `true` |

AC-DHR-023과 AC-CAR-012b의 내용 조건은 실행 1처럼 충족되지만, 두 판정식이 요구하는 같은 테스트의 pass 이벤트가 두 역할 실패로 0이라 `false`다.

검증(환경 정리 `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && …`): `go test ./internal/cli -run 'CodexAudit|CodexRole|Codex|LiveBudget|LiveEvidence|Rollout'` 0(`run/m4r-test-cli.log`), `go vet ./internal/cli/...` 0, `golangci-lint run ./internal/cli/...` 0 `0 issues.`, `GOOS=windows GOARCH=amd64 go build ./...` 0, 윈도 `go vet` 0. 증거 토큰 grep 0건.

Gaps (M4 재실행):

- AC-DHR-012/023, AC-CAR-012b는 FAIL이다. `manager-lead`와 `mission-governor`의 역할 계약이 "과업 없음"이나 "결정 객체 외 출력"을 거부하도록 쓰여 있어, 두 역할에 공통인 탐침 문구로는 NONCE 한 줄 응답을 얻지 못했다. 역할 파일이 로드된 사실 자체는 두 역할의 계약 출력으로 드러나지만, 판정식은 NONCE만 받는다. 판정식이나 탐침 방식을 바꾸는 결정은 이 카드의 권한 밖이다.
- 역할 로드 탐침은 여전히 부모 `-s workspace-write` + `spawn_agent` 경로다(이어받은 (i) 그대로). 네 read-only 역할이 이 경로에서 `workspace-write`로 돈다는 t1100 관측은 이 카드가 바꾸지 않았다.
- 음성 대조의 jq 조항 실행은 `jq`가 PATH에 있어야 한다. 없으면 테스트가 실패로 알린다(건너뛰지 않음).

### M5 — AC-CAR-010 LIVE (2/3 호출, 누계 36/43), AC-CAR-011 NOT_RUN

새 테스트 파일 `internal/cli/codex_audit_live_test.go`(`!windows`): `TestCodexAuditLaunchLiveContract`(AC-CAR-010)와 `TestCodexAuditLaunchLiveReadOnlyRoles`(AC-CAR-011). 격리 저장소, 로그인은 심볼릭 링크, 프로젝트 층 `.codex/config.toml`에 `sandbox_mode = "workspace-write"`와 `[mcp_servers.moai]`(기록 래퍼 → 이 트리에서 빌드한 `moai mcp-server`, `default_tools_approval_mode = "writes"`), 사용자 층에 `[mcp_servers.decoy]`(같은 방식의 기록 래퍼). 모든 경로·탐침 필드는 `deriveCodexAuditEvidence`로 계산한다. 첫 호출 전에 상한 블록(호출당 330초, AC-CAR-010 1100초 / AC-CAR-011 800초, 호출당 사용자 턴 1, (b) 자식은 launcher 자체 한도 20분)을 `verdict.md`에 적었다.

AC-CAR-010 실행(2026-09-24T09:52:16Z–09:52:54Z, go test 종료 코드 1, 원장 #35–#36, 로그인 sha256 전후 동일):

- (a) 직접 호출: 세션 `read-only`(프로젝트 config의 `workspace-write`에도 불구하고), nonce 되돌림 일치(`NONCE-2449720537de6d04197e9e48f92bd6eb`), 탐침 `exit_code` 1, 탐침 파일 없음, `write_denied` true, 감사 프로세스 동안 MCP 기동 moai 0 / decoy 0. (a)의 조건은 모두 충족.
- (b) 부모 `codex exec -s workspace-write -c approval_policy="never"`가 `codex_role_audit`를 한 번 불렀고, Codex가 프로세스를 띄우기 전에 거부했다. `--json` 항목 원문: `{"type":"mcp_tool_call","server":"decoy","tool":"codex_role_audit",…,"result":null,"error":{"message":"MCP tool call requires approval, but approval policy is never"},"status":"failed"}`. 부모 세션 기록의 도구 출력 원문: `{"content":[{"type":"text","text":"MCP tool call requires approval, but approval policy is never"}],"isError":true}`. 부모의 최종 메시지: `TOOL-REFUSED MCP tool call requires approval, but approval policy is never`. 부모가 부른 서버는 프로젝트 층 `moai`가 아니라 사용자 층 `decoy`였다(같은 도구를 노출함). 그래서 `default_tools_approval_mode = "writes"`를 단 `moai` 서버에서의 동작은 이번 실행이 재지 않았다. 양성 대조: 부모 세션 동안 MCP 기동 moai 1 / decoy 1(즉시 기동).
- 판정식 원문 출력 `false`(호출 수 2, (b) 필드 없음). 리드 지시("SPEC 경로가 진행할 수 없으면 멈추고 보고")에 따라 AC-CAR-011은 시작하지 않았다(0/2, NOT_RUN). 재실행 없음.

검증: `go test ./internal/cli -run 'CodexAudit|CodexRole|LiveBudget|LiveEvidence|Rollout'`(환경 변수 전부 제거) 0(`run/m5-test-cli.log`), `go vet ./internal/cli/...` 0, `golangci-lint run ./internal/cli/...` 0 `0 issues.`(`run/m5-lint.log`), `GOOS=windows GOARCH=amd64 go build ./...` 0, 윈도 `go vet` 0. 증거 토큰 grep 0건.

Gaps (M5):

- (b)의 쓰기 가능 MCP 도구 호출이 `approval_policy="never"`에서 거부됨을 관측했지만, 거부된 서버는 승인 모드를 따로 선언하지 않은 `decoy`다. `moai` 서버(`writes` 모드)도 같이 거부하는지는 이번 실행이 재지 않았다. `internal/codexwiring/configtoml.go`의 주석은 `writes`가 read-only 표시가 없는 도구에 승인을 요구한다고 적지만, 이것은 문서이지 측정이 아니다.
- AC-CAR-011은 NOT_RUN이다(리드 결정 대기, 호출 0회 사용).

### M5 approval probe — 리드 결정 뒤 (LIVE 2/2, 누계 38/43)

리드 결정(`verdict.md` § "Lead decision after M5"): AC-CAR-010 실행 1은 "INVALID — routed to decoy server"(원장 #35·#36과 결과 줄에 태그), `codex_role_audit`를 read-only로 표시하는 방안은 기각, 이 도구 하나만 프로젝트 `.codex/config.toml`의 도구별 승인 설정으로 사전 승인하는 방향을 채택, AC-CAR-011은 NOT_RUN 유지. 제품 코드와 템플릿은 바꾸지 않고 스크래치 픽스처 설정만 썼다.

도구별 승인 키: `[mcp_servers.<server>.tools.<tool>]`의 `approval_mode`, 값은 `auto` | `prompt` | `writes` | `approve`. 출처: (1) codex-cli 0.156.1 바이너리 문자열 — `McpServerConfig`의 `tools` 필드, `struct McpServerToolConfig`(`approval_mode`, `output_token_limit`), 열거형 `AppToolApproval`(`auto`, `prompt`, `writes`, `approve`). (2) codex 자체 설정 로더(모델 호출 없음): `approval_mode = "bogus"`를 넣은 `codex mcp get moai --json`이 `unknown variant `bogus`, expected one of `auto`, `prompt`, `writes`, `approve` in `mcp_servers.moai.tools.codex_role_audit.approval_mode``로 실패하고, `"approve"`면 정상 적재된다. 필드 이름을 틀리게 쓴 경우(`approval_modex`)는 오류 없이 무시된다(`m5-approval-probe/key-parse-probe.txt`).

- 호출 A(#37): moai 서버만 둔 픽스처, `default_tools_approval_mode = "writes"`, 도구별 승인 없음, 부모 `codex exec -s workspace-write -c approval_policy=never`. `--json` 항목 원문: `{"type":"mcp_tool_call","server":"moai","tool":"codex_role_audit",…,"result":null,"error":{"message":"MCP tool call requires approval, but approval policy is never"},"status":"failed"}`. moai 서버 자신도 쓰기 가능 도구를 이 조건에서 거부한다. 자식 프로세스와 launch record는 없다.
- 호출 B(#38): A와 다른 점은 `[mcp_servers.moai.tools.codex_role_audit] approval_mode = "approve"` 한 표뿐. 모델이 MCP 호출을 아예 내지 않았다. 최종 메시지 원문: `TOOL-REFUSED The moai codex_role_audit tool is not directly available, and your instructions prohibit calling another tool to discover or invoke it.` 0.156.1에서 MCP 도구는 `exec` 사용자 정의 도구를 거쳐 불리는데(A의 기록: `custom_tool_call` `exec` 안의 `ALL_TOOLS.find(…codex_role_audit…)`), B의 프롬프트가 "다른 도구 금지, shell 명령 금지"라고 적어 모델이 `exec`를 쓰지 않았다. 하네스 프롬프트 결함이므로 B는 INVALID(측정되지 않음)이다.
- 결론: 도구별 승인 없이 moai `writes` 서버는 `approval_policy=never`에서 `codex_role_audit`를 거부한다(A, 측정). `approval_mode = "approve"`가 호출을 통과시키는지는 측정되지 않았다(B INVALID). 탐침 상한 2회를 다 썼다.

증거: `.moai/reports/t1143/m5-approval-probe/`(픽스처 설정 A·B, 프롬프트, `--json` 출력, 세션 기록 둘, 키 적재 탐침). 로그인 파일 sha256 전후 동일, 남은 `codex exec` 프로세스 0, 토큰 grep 0건. 픽스처와 격리 `CODEX_HOME`은 지웠다.

### M5 approval probe rerun — B' (LIVE 1/1, 누계 39/43)

리드 결정(`verdict.md` § "Lead decision after the M5 approval probe"): 호출 B(#38)는 원장에 `INVALID — prompt blocked the exec custom-tool path` 태그를 달고 지우지 않았다. 재실행 1회 승인, 마지막 판별 시도.

- 호출 B'(#39): moai 서버만(`codex mcp list --json` 비모델 확인에서 서버 1개), 픽스처 `.codex/config.toml` = `default_tools_approval_mode = "writes"` + `[mcp_servers.moai.tools.codex_role_audit] approval_mode = "approve"`, 부모 `codex exec -s workspace-write -c approval_policy=never`, 프롬프트는 `exec` 사용자 정의 도구로 이 MCP 도구를 한 번 부르는 것만 허용(shell 명령·다른 도구·파일 변경 금지). exit 1, 1초 미만. stdout(`--json`) 0바이트, stderr 전체 원문: `Not inside a trusted directory and --skip-git-repo-check was not specified.`
- Codex가 세션을 시작하지 않았다. `thread.started`도 모델 턴도 세션 기록도 없고, `mcp_tool_call`은 0건이다. 분류: **NOT MEASURED**.
- 원인은 이 레인의 픽스처 결함이다. 호출 수를 1로 묶으려고 픽스처 루트를 일부러 저장소가 아닌 디렉터리로 만들었는데(moai 서버가 자식을 띄우기 전에 거부하도록), `codex exec`는 저장소 체크아웃 또는 신뢰 디렉터리를 요구하고 CODEX_HOME의 `trust_level = "trusted"` 항목은 그 조건으로 인정되지 않았다. #37·#38의 루트는 저장소였다. 자식 `codex exec`, launch record, verdict 쓰기 모두 없다.
- 결론: 도구별 사전 승인이 `approval_policy=never`를 넘는지는 **측정되지 않았다**. 리드 조건 (3)에 따라 LIVE를 더 쓰지 않고 이 카드는 "pre-approval effect not measured"로 멈춘다. 남은 4회는 AC-CAR-010/011용으로 남겼다(실행 안 함). #39는 프로세스 한 칸을 썼지만 모델 요청은 0건이었다 — 이것을 승인된 재실행 소진으로 볼지는 리드 판단이다.

증거: `.moai/reports/t1143/m5-approval-probe/`(`config-B2.toml`, `codexhome-config-B2.toml`, `prompt-B2.txt`, `B2.jsonl`(0바이트), `B2.err`). 로그인 파일 sha256 전후 동일(`eb7c45bd…6630`), 픽스처 경로로 거른 `ps` 결과 0건, 토큰 grep 0건. 픽스처와 격리 `CODEX_HOME`은 지웠다.

### M5 approval probe B'' — 실행 전 재구성 점검에서 중단 (LIVE 0, 누계 39/43)

리드 결정(`verdict.md` § "Lead decision after rerun B'"): #39는 승인된 재실행으로 집계하고 원장에 `NOT MEASURED — fixture root outside a git repository (a condition other than the tested variable changed)` 태그를 달았다. B'' 한 번이 승인됐지만, #37 픽스처를 기록된 증거로 바이트 단위까지 재구성할 수 없으면 실행하지 말라는 조건이 붙었다.

- 기록된 세 요소(`config-A.toml`, `prompt-A.txt`, #37 argv)로 만든 #37 입력과 B'' 입력의 `diff -ru` 결과는 사전 승인 표 두 줄(+ 빈 줄) 추가뿐이다(원문은 `verdict.md`). 프롬프트와 argv는 동일하다.
- 그래도 실행하지 않았다. 증거에 없는 것: (1) #37의 CODEX_HOME(`~/.t1143-ap-codexhome`, 세션 기록의 skill root로 확인) 안 `config.toml` — 신뢰 항목이 기록되지 않았고 디렉터리는 지워졌다. #39를 시작 검사에서 멈춘 바로 그 요소다. (2) #37 루트가 저장소였는지 — #37·#38 세션 기록의 `"git": {}`는 저장소임을 보여 주지 않는다. 리드 전제("git-repo root")를 #37에 대조할 수 없다. (3) #37 moai 서버 바이너리의 빌드 커밋.
- 결론: 도구별 사전 승인 효과는 측정되지 않았다. 이번 단계 LIVE 0회, 누계 39/43. 픽스처는 만들지 않았고 프로세스도 띄우지 않았다.

### sync-audit fix (F1–F3, F5) — `.moai/reports/t1143/sync-audit.md`, LIVE 0

재현 테스트를 먼저 쓰고(현재 코드에서 실패 확인), 원인을 한 번 따진 뒤 최소 수정했다. 테스트: `internal/cli/codex_audit_confine_test.go`. 가짜 `codex`에 실행 중 디렉터리를 심볼릭 링크로 바꾸는 선택 단계를 더했다(`swap.dir`/`swap.to`가 있을 때만).

- F1(판정 파일 쓰기의 TOCTOU): `TestCodexAuditLaunchWriteRaceConfined` — 가짜 `codex`가 실행 중 `.moai/reports/x`(가운데 성분)를 워크트리 밖 심볼릭 링크로 바꾼다. 수정 전 FAIL: `verdict escaped the worktree through the swapped component: [.../outside/y/v.md]`(`run/fix-red-F1.log`). 원인: 이름으로 검증한 뒤 이름으로 다시 걸어가 쓴다(MkdirAll·CreateTemp·rename). 쓰기 시점에 이름으로 다시 확인하는 것도 같은 틈이라 증상 처방이다. 수정: `codexAuditOpenDir`가 워크트리 뿌리부터 성분마다 부모 핸들에서 Lstat(심볼릭 링크·비디렉터리 거부) → 같은 부모 핸들에서 `OpenRoot` → `SameFile`로 동일성 확인을 거쳐 디렉터리 핸들(`os.Root`)을 연다. 임시 파일 생성과 rename은 그 핸들 안에서만 한다. launch record의 `verdict_path`는 실제로 쓴 경로다. 수정 후 PASS(`run/fix-green-F1.log`, 경쟁은 여전히 일어나고 워크트리 밖에는 아무것도 없으며 종료 코드 1, `verdict_path` null, `failure_reason` 기록).
- F2(대소문자 별칭): `TestCodexAuditLaunchRecordAliasRefused` — `CODEX-AUDIT`·`Codex-Audit` 별칭으로 기존 launch record 덮어쓰기, `.GIT`·`.Git` 성분. 수정 전 네 경우 모두 수락(`run/fix-red-F2.log`). 원인: 대소문자 무시 파일 시스템 위에서 바이트 단위 문자열 비교. 수정: `strings.EqualFold`와 `os.SameFile` 동일성 비교, 그리고 쓰기 경로의 디렉터리 걷기에서 기록 디렉터리 핸들과 같은 디렉터리를 거부. 수정 후 PASS(`run/fix-green-F2.log`).
- F3(심볼릭 링크 기록 디렉터리): `TestCodexAuditLaunchSymlinkedRecordDir` — `.moai/reports/codex-audit`, 그리고 `.moai/reports` 자체(stdout 반환)를 밖으로 가리키는 링크. 수정 전 두 경우 모두 launch record가 워크트리 밖에 생김(`run/fix-red-F3.log`). 원인은 F1과 같다(MkdirAll·OpenFile이 링크를 따라감, O_EXCL은 마지막 성분만 지킴). 수정: 기록 예약도 같은 `codexAuditOpenDir`로 연 핸들에서 `O_EXCL`로 만든다. 수정 후 PASS(`run/fix-green-F3.log`), 감사 프로세스는 시작되지 않는다.
- F5: CHANGELOG와 이 파일 §E.4의 집계를 "17 verdict rows over the 16 acceptance criteria (AC-CAR-012 as 012a/012b): 12 PASS, 3 FAIL, 1 INVALID, 1 NOT_RUN"과 "12 deterministic verdict rows"로 고치고 `b12_self_test_b` 줄을 맞췄다.
- 검증(이 트리): `go test ./internal/cli -run CodexAudit` ok(PASS 48, SKIP 2 = LIVE 테스트, `run/fix-tests.log`); AC-CAR-003·004·005·014 판정식 원문 모두 `true`(`run/fix-judge-*.log`); `go vet` exit 0; `golangci-lint run ./internal/cli/...` `0 issues.`; `GOOS=windows GOARCH=amd64 go build ./...` exit 0, Windows vet exit 0.

## §E.3 Run-phase Audit-Ready Signal

run 단계는 끝나지 않았다. AC-CAR-011이 NOT_RUN이고, AC-CAR-010 실행 1은 리드가 INVALID로 기록했으며 재실행되지 않았다. MCP 경로는 도구별 승인 설정이 통과시키는지 측정되지 않았다(approval probe B INVALID).

| AC | 상태 | 근거(판정식 출력과 위치) |
|---|---|---|
| AC-CAR-001 | PASS | 판정식 `true`(HEAD `434a39e1f` 재판정), 변이 m001 `false` — §E.2 M2 |
| AC-CAR-002 | PASS | `true`(HEAD 재판정), 변이 m002/m002b `false` — M2 |
| AC-CAR-003 | PASS | `true`(HEAD 재판정), 변이 m003/m003b `false` — M2 |
| AC-CAR-004 | PASS | `true`(HEAD 재판정), 변이 m004/m004b `false` — M2 |
| AC-CAR-005 | PASS | `true`(HEAD 재판정), 변이 m005/b/c `false` — M2 |
| AC-CAR-006 | PASS | `true`(HEAD 재판정), 변이 m006 `false` — M2 |
| AC-CAR-007 | PASS | `true`(HEAD 재판정), 변이 m007 `false` — M3 |
| AC-CAR-008 | PASS | `true`(HEAD 재판정), 작업 트리 변형 변이 `false` — M3 |
| AC-CAR-009 | PASS | `true`(M4 재실행 뒤) — M4 |
| AC-CAR-010 | INVALID (실행 1, 리드 기록 "routed to decoy server"; 재실행 없음) | 판정식 `false`. (a) 조건 충족. (b) 거부. approval probe A: moai `writes` 서버도 같은 문구로 거부 — M5, M5 approval probe |
| AC-CAR-011 | NOT_RUN | 호출 0회, 리드 지시로 멈춤 — M5 |
| AC-CAR-012a | PASS | `true`, 변이 `false` — M4 |
| AC-CAR-012b | FAIL | `false`(M4 재실행; 유도 필드 조건은 충족, 같은 테스트 pass 이벤트 0) — M4 재실행 |
| AC-CAR-013 | PASS | `true`(HEAD 재판정), 변이 m013 `false` — M3 |
| AC-CAR-014 | PASS | `true`(HEAD 재판정), 변이 m014/m014b `false` — M2 |
| AC-DHR-012 | FAIL | `false`(M4 재실행: `manager-lead`, `mission-governor` nonce 불일치; 실행 1은 리드가 INVALID로 기록) |
| AC-DHR-023 | FAIL | `false`(같은 이유; 내용 조건은 충족) |

```yaml
run_complete_at: null            # run not complete: AC-CAR-011 NOT_RUN, lead decision pending
run_commit_sha: <backfill>       # the M5 commit carrying this section
run_status: blocked
ac_pass_count: 12
ac_fail_count: 3
ac_invalid_count: 1              # AC-CAR-010 run 1 (lead-recorded)
ac_not_run_count: 1
preserve_list_post_run_count: null   # not measured in this run
l44_pre_commit_fetch: not_measured
l44_post_push_fetch: not_applicable  # lanes do not push (lead batch push)
new_warnings_or_lints_introduced: 0  # golangci-lint ./internal/cli/... 0 issues (run/m5-lint.log)
cross_platform_build:
  darwin: go build/test on this host, exit 0 (run/m5-test-cli.log)
  windows_amd64: GOOS=windows GOARCH=amd64 go build ./... exit 0 (run/m5-windows-build.log)
total_run_phase_files: 41        # git diff --name-only develop...HEAD | wc -l at 434a39e1f
live_calls_used: 39              # of the absolute cap 43 (#39 = B' rerun, NOT MEASURED)
m1_to_mN_commit_strategy: one or more commits per milestone on WT-codex-audit-readonly, no push
```

## §E.4 Sync-phase Audit-Ready Signal

### AC verdict table (as recorded; carried-over items counted as their recorded status, never PASS)

| AC | Verdict | Evidence |
|---|---|---|
| AC-CAR-001 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-002 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-003 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-004 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-005 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-006 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-007 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-008 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-009 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-012a | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-CAR-013 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` (`route.txt` = `mcp`) |
| AC-CAR-014 | PASS | re-run at HEAD `9e1e926cf`, `.moai/reports/t1143/run/sync-judges.log` |
| AC-DHR-012 | FAIL | run 2 (rerun), ledger #21–#34 — `manager-lead`/`mission-governor` nonce mismatch — `.moai/reports/t1143/verdict.md` |
| AC-DHR-023 | FAIL | run 2 (rerun), same evidence — content clauses satisfied, pass-event clause fails |
| AC-CAR-012b | FAIL | run 2 (rerun), same evidence — derivation fields satisfied, pass-event clause fails |
| AC-CAR-010 | INVALID (run 1: "routed to decoy server", lead-recorded, no rerun) | ledger #35–#36, `.moai/reports/t1143/verdict.md` § "Lead decision after M5" |
| AC-CAR-011 | NOT_RUN (0/2 calls; lead instructed stop when AC-CAR-010 could not proceed) | `.moai/reports/t1143/verdict.md` § "Lead decision after M5" |

17 verdict rows over the 16 acceptance criteria of acceptance.md (AC-CAR-012 judged as 012a and 012b): 12 PASS, 3 FAIL (AC-DHR-012, AC-DHR-023, AC-CAR-012b), 1 INVALID (AC-CAR-010), 1 NOT_RUN (AC-CAR-011). Per `acceptance.md` §D/§E, none of the three design bundles reaches full PASS and the card's own completion definition is not met for the carried-over scope; this SPEC closes `completed` for the delivered launcher/instruction-surface/instruction-preservation scope only, following the SPEC-DUAL-HARNESS-RECOVERY-001 v0.3.1 (card t1100) precedent for closing with formally-transferred unmet items. This SPEC does **not** claim AC-DHR-012, AC-DHR-023, AC-CAR-012b, AC-CAR-010, or AC-CAR-011 as satisfied.

### Carry-over table (formal transfer, unchanged AC bodies/judges/expected values)

| Item | Status | Follow-up card |
|---|---|---|
| AC-DHR-012 / AC-DHR-023 / AC-CAR-012b (role-load judge vs. role contracts that refuse task-less requests) | FAIL | t1171 |
| AC-CAR-010 (routed to decoy server) | INVALID | t1172 |
| AC-CAR-011 (never started) | NOT_RUN | t1172 |
| Per-tool MCP pre-approval effect under `approval_policy=never`; fixture-input export; non-model trust-check CLI; template-default operator approval; key-name loader validation | open | t1172 |
| AC-FLH-015 (SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001) MCP-tool-count prose drift (36/14/22 vs. this card's 39/15/24) | recorded, not fixed | t1170 |

### LIVE call summary

- LIVE total: 39/43 (absolute cap). Full per-call ledger, caps declared before each item, and every lead decision: `.moai/reports/t1143/verdict.md`.

### Deterministic-judge re-verification (this sync commit)

Re-ran the 12 deterministic/derivation acceptance criteria VERBATIM against HEAD `9e1e926cf` (env-scrubbed: `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKERS`); all 12 `true`. Output: `.moai/reports/t1143/run/sync-judges.log`.

```yaml
sync_complete_at: 2026-09-24
sync_commit_sha: pending-backfill-sync
sync_status: completed_with_carryover
b12_self_test_a: pass         # grep -c 'SPEC-CODEX-AUDIT-READONLY-001' CHANGELOG.md == 1 (this entry only)
b12_self_test_b: pass         # 16 distinct AC-CAR-*/AC-DHR-* identifiers in acceptance.md; CHANGELOG cites 17 verdict rows over 16 ACs (AC-CAR-012 as 012a/012b): 12 PASS / 3 FAIL / 1 INVALID / 1 NOT_RUN
b12_self_test_c: pass         # all cited file paths verified via ls
changelog_entry_position: Unreleased > Added (top entry)
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (status field only; version/updated/HISTORY untouched per ownership matrix)"
canary_compliance_check: not_applicable   # this SPEC defines no forward-looking policy that its own sync tests
```

_<pending backfill of sync_commit_sha in a following commit>_
