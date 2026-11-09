# MCP parity 조사 — RC1 CLI op ↔ Go MCP 38 tool (rc.2 B9 #2176)

- 상위: #2154 (EP-B). 설계 근거: 설계 §1·§13. 사용자 결정 18:2x(HOLD-C1): **RC1은 CLI만, MCP는 다음 버전**(`docs/decisions-2026-11-08.md` 51행).
- 이 문서는 조사 결과만 담는다. MCP 구현은 별도 카드에서 한다(카드 범위: "구현 blocked").
- 대조 기준
  - Go: `moai-adk-go` commit `4fa2aa9db2e0add0f1299e35407141b2b41c5ec0`(2026-11-07), `github.com/mark3labs/mcp-go v1.3.0`(`go.mod:18`).
  - RC1: `origin/v2026.11.0` `81f6f7148`(#2218 `queue.next` 포함), `src/main.rs` `OPERATIONS` 49개. 줄 단위 계산과, 이름 정규식에 `_`가 든 op 4개를 더한 계산이 둘 다 49였다.
- 측정 방법
  - Go tool 목록은 `internal/` 비시험 파일의 `mcp.NewTool(` 첫 인자로 뽑았다. 34개는 문자열 리터럴이고, 4개는 `codexRoleAudit*ToolName` 상수다. 합계 38이다.
  - 같은 build 계열의 실행 중인 `moai mcp-server`(build v3.2.1-rc.6, commit `af9d485e2d`)가 노출하는 `mcp__moai__*` 목록도 38개였다. 두 방법의 결과가 같다.
  - tool별 입력, 쓰기 여부, 결과 형태는 소스를 읽어 정리했다. 근거 위치는 `file:line`으로 적었다.

## 1. MCP 사양 기준 (2026-11-09 확인)

출처: https://modelcontextprotocol.io/specification/latest/server/tools. 현행 개정판은 2026-07-28이다. 사양 본문의 schema 링크와 changelog의 `2025-11-25...2026-07-28` 비교에서 확인했고, "latest is"라는 문장은 페이지에 없다.

- `outputSchema`는 선택이다. 선언하면 "Servers **MUST** provide structured results that conform to this schema."
- 호환: "a tool that returns structured content SHOULD also return the serialized JSON in a TextContent block."
- 도구 실행 오류는 결과의 `isError: true`로, 알 수 없는 tool이나 잘못된 요청은 JSON-RPC 오류로 낸다.
- `annotations`는 "untrusted unless they come from trusted servers"다. `readOnlyHint` 같은 개별 힌트 이름은 이번에 읽은 페이지에서 확인하지 못했다(§5).

## 2. 대응표 — Go MCP 38 tool 전수

열 설명
- **결과**: structured는 `StructuredContent`를 쓰고 같은 JSON을 TextContent에도 넣는 형태다. text는 평문, json-text는 TextContent 안의 JSON 문자열만 있는 형태다. 출력 schema가 선언된 tool은 결과 칸에 함께 적었다.
- **RC1 op**: 기능상 가장 가까운 op다. 칸 앞의 표시는 다음과 같다.
  - **있음**: 같은 일을 한다.
  - **부분**: 일부만 하거나 입력·출력 계약이 다르다.
  - **없음**: RC1에 대응물이 없다.
- `pr`은 `project_root`, `!`는 필수 입력이다.

| # | Go tool (등록 위치) | 입력 | 쓰기 | 결과 | RC1 op | 차이 |
|---|---|---|---|---|---|---|
| 1 | `session_list` (mcp_server.go:480) | spec_id, pr | 없음 | structured | 없음 | RC1에는 세션 레지스트리가 없다. 소유는 Issue claim 댓글이 정본이다(`claim.show`는 카드 단위만 본다) |
| 2 | `goal_status` (:489) | session_id, pr | 없음 | structured | 없음 | goal 상태 개념 없음 |
| 3 | `goal_arm` (:501) | condition!, session_id, max_turns, max_duration, pr! | `.moai/state` goal 파일 | structured | 없음 | 같음 |
| 4 | `spec_progress` (:512) | pr | 없음 | structured | 부분 `forge.read` | Go는 로컬 SPEC frontmatter를 읽는다. RC1 정본은 Issue 본문과 보고 댓글이다(GET) |
| 5 | `verify_snapshot` (:520) | pr, key!, command, exit_code | `command`가 있으면 snapshot에 기록 | structured | 부분 `run.verify` | RC1은 승인된 명령을 직접 실행하고 receipt를 반환한다. 호출자가 보낸 exit_code를 기록하는 경로는 없다 |
| 6 | `verify_trend` (:530) | pr, key! | 없음 | structured | 없음 | 검증 이력 저장소가 없다(receipt는 실행마다 하나) |
| 7 | `spec_audit` (:539) | filter_spec, filter_era, include_grandfathered, pr | 없음 | structured | 부분 `plan.admit` | RC1은 계획 1건의 승인 판정만 한다. era 분류나 일괄 감사는 없다 |
| 8 | `spec_drift` (:550) | pr | 없음 | structured | 없음 | SPEC drift 개념이 없다(RC1은 `context.verify`로 범위의 신선도만 본다) |
| 9 | `audit_cache` (:558) | op!(compute_hash/lookup/store), spec_dir, spec_id, hash, auditor_version, report_path, pr | `store`가 프로세스 메모리 캐시에 쓴다(힌트는 read-only) | structured | 없음 | RC1 감사 판정은 매번 다시 계산한다(`audit.reduce`) |
| 10 | `codex_audit` (:582) | mode, target, focus, model, delta_base, pr | `.moai/state/audit-receipts/` receipt, 외부 `codex` 실행 | structured, `ReviewOutput` schema | 부분 `audit.reduce` | RC1은 외부 backend를 부르지 않는다. 받은 판정과 검증 명령을 묶어 판정만 계산한다 |
| 11 | `claude_audit` (:608) | target, focus, model, effort, pr | 외부 Claude CLI 실행 | structured, `ReviewOutput` schema | 부분 `audit.reduce` | 10과 같음 |
| 12 | `codex_setup` (:623) | 없음 | 없음(`codex --version` 탐침) | structured | 부분 `codex.doctor` | RC1 doctor는 projection과 설정을 진단한다. 인증 공급자와 review-gate 토글 탐침은 없다 |
| 13 | `glm_task` (:654) | pr!, prompt!, background, model, system, max_tokens | `.moai/state/glm-jobs/`, z.ai HTTPS | structured | 없음 | RC1은 외부 모델을 호출하지 않는다(결정 01:1x: 비용 상한 미설정 = 외부 호출 0) |
| 14 | `glm_job_status` (:667) | job_id!, pr | 없음 | structured | 없음 | 13과 같음 |
| 15 | `glm_job_result` (:677) | job_id!, pr | 없음 | structured | 없음 | 13과 같음 |
| 16 | `glm_job_cancel` (:688) | job_id!, pr! | job 기록, 요청 취소 | structured | 없음 | 13과 같음 |
| 17 | `glm_audit` (:705) | target, focus, model, pr | z.ai HTTPS | structured, `ReviewOutput` schema | 부분 `audit.reduce` | 10과 같음 |
| 18 | `codex_review` (:725) | scope!(card/uncommitted), model, pr | 없음(codex 실행) | structured, `SelfReviewOutput` schema | 없음 | 자기 변경 advisory review 없음 |
| 19 | `glm_review` (:735) | scope!, model, pr | z.ai HTTPS | structured, `SelfReviewOutput` schema | 없음 | 18과 같음 |
| 20 | `audit_multi` (:752) | claude_verdict, target, focus, gates, session_id, card_id, delta_base, blocking_closed, rerun_legs, pr | receipt, `.moai/state/audit-multi/`, `second-review.jsonl`, 3개 backend 호출 | structured, schema 없음 | 부분 `audit.reduce` | RC1 reducer는 판정 수렴만 한다. backend 팬아웃은 없다 |
| 21 | `graph_file_api` (:783) | file!, pr | 없음 | structured | 없음 | F13 graph는 후속 Issue다(`docs/rust-product-path.md`) |
| 22 | `graph_find_code` (:791) | query!, pr | 없음 | structured | 없음 | 21과 같음 |
| 23 | `graph_trace_calls` (:799) | symbol!, depth, pr | 없음 | structured | 없음 | 21과 같음 |
| 24 | `graph_shortest_path` (:812) | from!, to!, pr | 없음 | structured | 없음 | 21과 같음 |
| 25 | `fetch_web` (mcp_fetch_web.go:29) | url!, question!, pr | `workflow.jev.enabled`일 때만 네트워크 | structured | 없음 | RC1은 네트워크를 쓰지 않는다(`context.*`는 `network: disabled`) |
| 26 | `jev_ask` (mcp_jev.go:59) | state!, questions!, repeatable, pr | gate가 켜졌을 때 네트워크, `~/.moai/cache/jev` | structured | 없음 | 판단 backend 조사는 #2204가 맡는다 |
| 27 | `todo_add` (mcp_todo.go:30) | text!, pr | backlog 큐 추가 | text | 없음 | RC1 작업 장부는 GitHub Issue다. 카드 생성은 운영자나 리더가 한다 |
| 28 | `todo_list` (mcp_todo.go:38) | pr | 없음 | text | 부분 `forge.read` | 큐 대신 열린 Issue를 GET으로 읽는다 |
| 29 | `todo_claim` (mcp_todo.go:45) | lane, renew, pr | lease 기록 | text(거부는 structured) | 있음 `claim.acquire` | RC1 claim은 Issue 댓글과 로컬 journal이다. 가장 작은 live 댓글 id가 이긴다 |
| 30 | `factory_next` (mcp_factory_card.go:45) | run, card, actor, pr! | factory 기록 lease | text | 부분 `queue.next`(#2161) | milestone의 열린 Issue에서 claim 댓글이 없는 다음 카드를 self-claim한다. 로컬 factory 기록은 쓰지 않는다 |
| 31 | `factory_stage` (:55) | card!, state!, evidence, run, actor, pr! | factory 기록 | text | 부분 `forge.report` | RC1 단계는 보고 댓글의 `상태:` 줄이다(forge-report 게이트 뒤) |
| 32 | `factory_complete` (:67) | card!, run, actor, pr! | factory 기록, 로컬 merge 또는 push+PR | text | 부분 `pr.body.render` + `pr.confirm` | RC1 레인은 병합하지 않는다. PR 생성 효과를 의도·확인으로 기록하고, 병합은 리더의 `moai/audit` 게이트가 한다 |
| 33 | `factory_decide` (:77) | card!, gate, choice, run, pr | 운영자 결정 기록 | text | 부분 `decision.record` | RC1 decision record는 authority를 싣지 않는다. 운영자 결정 기록은 decisions log 문서다 |
| 34 | `factory_renew` (:88) | card!, run, actor, pr! | lease 연장 | json-text(:417) | 없음 | RC1 lease는 claim 댓글의 `lease_until`이다. 연장 op는 없다(만료 후 자동 takeover도 금지) |
| 35 | `codex_role_audit` (codex_audit_mcp.go:230) | role!, worktree_root!, task!, out | launch 기록, `.moai/reports/` 보고, 백그라운드 `codex exec` | structured | 없음 | 외부 backend 작업 없음 |
| 36 | `codex_role_audit_status` (:238) | job_id! | 없음 | structured | 없음 | 35와 같음 |
| 37 | `codex_role_audit_result` (:243) | job_id! | 없음 | structured | 없음 | 35와 같음 |
| 38 | `codex_role_audit_cancel` (:248) | job_id! | job 취소 | structured | 없음 | 35와 같음 |

집계(38): 있음 1 · 부분 13 · 없음 24.

### 2.1 반대 방향 — Go MCP에 대응물이 없는 RC1 op

RC1 49개 op 가운데 위 표가 쓰지 않은 op는 다음과 같다. MCP로 노출할 때는 이 op들의 계약이 기준이 된다.

`package.inspect`, `project.plan`/`project.apply`, `package.resolve`, `agents.plan`/`agents.apply`, `apply.recover`, `baseline.inventory`, `identity`, `claude.doctor`, `codex.projection_plan`, `claim.show`, `claim.release`, `claim.handoff`, `effect.intend`, `effect.confirm`, `effects.reconcile`, `todo.promote.plan`/`todo.promote.apply`, `snapshot.describe`, `skill_source.plan`/`skill_source.apply`/`skill_source.read_receipt`, `generation.plan`/`generation.apply`, `context.bundle`, `context.verify`, `install.inspect`, `worktree.plan`/`worktree.apply`, `scope.check`, `github.plan`/`github.apply`, `update.plan`/`update.apply`, `recovery.plan`/`recovery.apply`, `parity.check`.

## 3. 결과 형태와 structuredContent

- Go는 38개 중 30개가 structured다. 공용 `toolJSON`/`toolJSONWithRoot`(`mcp_server.go:1326`)가 `NewToolResultStructured`로 `StructuredContent`와 TextContent를 함께 채운다. 사양의 SHOULD(TextContent에 직렬화 JSON 병기)를 따른다.
- 출력 schema를 선언한 tool은 5개다. `codex_audit`·`claude_audit`·`glm_audit`이 `ReviewOutput`, `codex_review`·`glm_review`가 `SelfReviewOutput`이다. `audit_multi`는 structured이지만 schema가 없다.
- text 결과는 7개다(`todo_*` 3개, `factory_next`·`factory_stage`·`factory_complete`·`factory_decide`). json-text는 1개다(`factory_renew`). 이 8개는 기계가 결과를 읽으려면 문자열을 파싱해야 한다.
- 오류는 `toolErr`가 `isError: true`와 text로 낸다. lane 거부(`toolRefusal`, `mcp_refusal.go:77`)는 structured로 낸다.
- RC1 CLI는 모든 출력이 envelope `{protocol_version, operation, ok, result|error}`다(결정 18:2x HOLD-00D-1). 실패는 `ok:false, error{kind, message, retryable, next_actions?}`다. 이 형태는 MCP의 `StructuredContent`(성공)와 `isError: true` + 같은 객체(실패)로 그대로 옮길 수 있다. 다만 `outputSchema`를 선언하려면 op마다 `result`의 schema가 있어야 한다. RC1의 `docs/protocol-v1.md`는 envelope을 정의하지만, op별 result JSON Schema 파일은 없다(이 조사에서 확인한 범위).

## 4. 결론 — rc.2 안 MCP 범위 제안

이 절은 레인 제안이다. 결정이 아니며 리더와 사용자의 확인이 필요하다.

1. **Go 38 tool을 옮기지 않는다.** 24개는 RC1에 대응 개념이 없다. 외부 backend 호출, 세션·goal 상태, graph, 로컬 factory 큐가 여기에 해당하며, 대부분 RC1 결정과 어긋난다(외부 호출 0, Issue 장부 정본, 레인 무병합). Go 계약을 기본값으로 이식하지 않는다는 원칙에 따라, MCP 노출 대상은 RC1 op에서 고른다.
2. **1차 노출은 읽기 전용 RC1 op로 한정한다.** 후보는 `identity`, `install.inspect`, `snapshot.describe`, `context.bundle`, `context.verify`, `claim.show`, `forge.read`, `scope.check`, `package.inspect`, `claude.doctor`, `codex.doctor`, `parity.check`다. 쓰기 op는 plan/apply 2단계와 `--plan-digest`, `--permission-reference`를 거친다. MCP 호출 한 번으로 이 승인 절차를 대신하지 않도록, 쓰기 op는 2차로 미룬다.
3. **결과는 structured + TextContent 병기로 한다.** 실패는 `isError: true`와 RC1 오류 객체로 낸다. `outputSchema`는 op별 result schema를 만든 뒤 선언한다. schema 없이 선언하면 사양의 MUST를 지킬 수 없다.
4. **기존 정책을 바꾸지 않는다.** 네트워크와 외부 backend 0, `project_root` 명시(worktree 오인 방지: Go `mcp-server`는 시작 시 cwd에 고정된다), annotations를 신뢰 경계로 쓰지 않음을 그대로 지킨다.
5. **별도 카드로 분리한다.** (a) op별 result JSON Schema, (b) 읽기 전용 MCP 서버 1차, (c) 쓰기 op의 승인 절차를 MCP에 매핑하는 설계. 셋 다 결정 18:2x "MCP는 다음 버전"의 범위 안에서 일정을 정한다.

## 5. 미확인·한계

- MCP `annotations`의 개별 힌트 이름(`readOnlyHint` 등)은 이번에 읽은 사양 페이지(tools, 본문, changelog)에서 확인하지 못했다. Go 코드는 `ReadOnlyHint`를 선언한다. 사양상 위치는 schema 원문에서 다시 확인해야 한다.
- 결과 형태와 쓰기 여부는 소스를 읽어 판정했고, Go 서버를 호출해 응답을 받아 보지는 않았다(not_run). 다만 등록 수 38은 실행 중인 서버의 tool 목록으로 교차 확인했다.
- RC1 측 대응은 op 이름과 문서(`docs/rust-product-path.md`)를 근거로 한 기능 대조다. 입력 schema를 필드 단위로 비교하지는 않았다.
- 실행 중인 Go 서버로 tool 수를 교차 확인할 때 worktree 격리 장치가 일부 확인 명령에 "실행되지 않음" 알림을 붙였다(출력은 관찰됨). 해당 줄 확인(`mcp_factory_card.go:417`, `mcp_server.go:591`, gate 노출 16개, `go.mod:18`)은 탐색 결과와 일치했지만, 측정 상태가 불분명한 것으로 기록한다.
