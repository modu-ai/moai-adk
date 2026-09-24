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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
