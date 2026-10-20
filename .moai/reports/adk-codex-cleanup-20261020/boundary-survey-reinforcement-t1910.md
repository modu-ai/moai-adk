# 경계 조사 보강 — codex audit 최소 의존성 분리 설계 (t1910, manager-spec, 2026-10-20)

리더의 `boundary-survey.md`(주 checkout, 미편집)를 보강하는 별도 문서다. 운영자 소스 실측 3건을 검증하고, 감사 의존 closure를 전수 매핑하며, t1909(일반 제거)가 감사를 깨지 않는 분리 설계와 제거 전 보존 측정을 기록한다. 이 문서는 **설계+측정**이지 구현이 아니다.

## 0. 읽기 기준선 (read base — 명시 의무)

- **측정 기준 트리**: local main `8249bff8c3` — t1910 트리 HEAD가 local main 선단과 동일 (`git rev-parse main` → `8249bff8c3933d647b4f99e59048595f9808e83a`). 이 문서의 모든 파일:행 번호는 이 커밋에 고정되며 moving ref에 대해 진술하지 않는다.
- **원격 기준**: origin/main `6dde1c1c1b`. `git rev-list --count --left-right origin/main...main` → `0  559` — local main이 559 앞섬·0 뒤처짐. **이 문서가 다루는 감사 표면은 local main에 존재하며, origin/main에도 동일하게 존재한다(정정 — 아래 §0 전제 정정 참조).** 559개 미착지 커밋은 감사 표면의 존재 조건이 아니라 local의 선행분이다.
- **batch branch 관계**: `80c76656cf→16d03deea3` 구간이 #1855/#1856을 싣고 있고 local main에는 미병합 — `git rev-list --count main..16d03deea3` = 5. 단 네 감사 파일(codex_audit_launch/codex_audit_mcp/codex_jobs/mcp_codex)에 대한 `git diff --stat main...16d03deea3`가 **공집합**이므로, batch 흡수 후에도 이 문서의 인용 행 번호는 무효화되지 않는다.
- **원격기준 카드 원칙 + 의존 병합 전제 (§0 전제 정정 — 운영자 기계 실측 2026-10-20, 2차 정정 보강)**: 이 문서의 이전 전제 "origin/main에는 감사 표면 자체가 없다"는 **사실이 아님 — 정정한다.** 운영자 실측: origin/main `6dde1c1c1b`의 codex_audit_launch.go(34657B)에 prepareCodexAudit :168·LoadManifest :506·:927-928 등록이 존재. **SHA 동일 파일은 codex_audit_mcp.go(9066B)·codex_jobs.go(22034B) 두 파일로 한정**(이전 "네 파일" 서술 삭제). **실제 차이 파일 목록(batch 선행 의존 근거로 재기록)**:
  1. codex_audit_launch.go — :618·:705의 `codexChildEnv`→`codexAuditChildEnv` 호출 변경(4 lines). 증거: `/tmp/adk-handoff-continue-20261020/audit-main-batch-launch-diff.patch`.
  2. mcp_codex.go — 2367B 수치 이상의 실질 변경(RecordingDefect·delta_base 검증/범위·codexChildCmd). 증거: `/tmp/adk-handoff-continue-20261020/audit-main-batch-mcp-diff.patch`.
  **base 재선택 근거(위 실제 차이 목록으로 재기록)**: batch 기준(`--base-branch release/main-batch-20261019` @16d03deea3)은 **이 차이들이 batch 라인에서 분화 중 — batch 착지 후 제거가 충돌 최소**이기 때문이다. t1909는 §4 분리 착지를 포함하는 기준 HEAD에서 분기하는 것은 유지. 모든 commit/push 전 HEAD 재판독(staleness rule)은 그대로.
- **이 트리에서의 구현 금지 (운영자 지시)**: 본 트리는 읽기 전용이며 이 카드의 산출물은 보고서 2건뿐이다. t1909 구현은 설계 확정 + base 원칙 점검 **이후**, 별도 격리 트리에서만 시작한다.

## 1. 운영자 소스 실측 3건 검증 — 3건 모두 CONFIRMED

| # | 판정 | 근거 (8249bff8c3 기준, 원문 인용) |
|---|---|---|
| 1 | **CONFIRMED** | `codex_audit_launch.go:501` `var codexAuditRoleFS = template.EmbeddedTemplates` · `:506` `man, err := agentemit.LoadManifest()` · `:517` `entries, err := fs.ReadDir(fsys, codexAuditRoleDir)` (단수 값: `:58` `codexAuditRoleDir = ".codex/agents/moai"`) · `:524` `man.PermissionContract.ContractSandbox(name) == codexAuditSandbox` (단수 값: `:57` `codexAuditSandbox = "read-only"`) · `:545` `src, err := fs.ReadFile(fsys, codexAuditRoleDir+"/"+role+".toml")` |
| 2 | **CONFIRMED** | `codex_audit_mcp.go:105` `id, err := newCodexJobID()` → `codex_jobs.go:522` `func newCodexJobID() (string, error) {`. 보강 실측: 이것이 감사 경로가 codex_jobs.go에서 가져가는 **유일한** 심볼이다 — 감사 keep-set 파일 전수 grep에서 `codexJobRecords`·`codexJobStateError`·`writeFileAtomic`·`handleCodexJob*` 참조 0건 |
| 3 | **CONFIRMED (행 보정)** | 등록 호출은 `:923-924`가 아니라 그 아래 `init()` **`:926-929`** 에 있다 (`:923-924`는 `newCodexRoleAuditCmd`의 반환부): `func init() { codexCmd.AddCommand(newCodexAuditCmd()); codexCmd.AddCommand(newCodexRoleAuditCmd()) }`. 핵심: `codexCmd`의 정의는 **`codex_launcher.go:646`** — t1906 제거 범위 안이다. 컴파일 단위 결합의 본체 |

추가 보정 1건: 매니페스트는 template FS가 아니라 agentemit 패키지 자체 임베드에서 읽힌다 — `internal/template/agentemit/manifest.go:13` `//go:embed agents-codex.yaml`, `:100-102` `LoadManifest() { return ParseManifest(manifestYAML) }`. template FS가 공급하는 것은 역할 TOML 파일뿐.

## 2. 의존 closure — 감사가 실제로 닿는 전체

### 2.1 진입점

| 진입 | 코드 위치 | 등록 |
|---|---|---|
| MCP `codex_role_audit`/`_status`/`_result` | `codex_audit_mcp.go:88/150/169` | `mcp_server.go:426-428` `codexRoleAuditTools()` 루프 |
| CLI `moai codex audit` / `moai codex role-audit` | `codex_audit_launch.go` (`newCodexAuditCmd`/`newCodexRoleAuditCmd`) | `init()` :926-929 → `codexCmd` |
| MCP `codex_audit` (review shellout) | `mcp_codex.go:2056` `handleCodexAudit` | `mcp_server.go:316` |
| MCP `codex_setup` | `mcp_codex.go:2281` | `mcp_server.go:357` |
| MCP `audit_multi` | `mcp_audit_multi.go` → `runMultiAudit` (`mcp_convergence.go`) | catalog.go:84 |

### 2.2 closure 표 (심볼 → 위치 → t1909 범위와 공유?)

| 심볼/파일 | 위치 | t1909 공유? | 비고 |
|---|---|---|---|
| `prepareCodexAudit`·`codexAuditLoadRole`·`codexAuditLaunchableRoles`·`codexAuditParseRoleTOML`·`codexAuditGit`·`codexAuditValidateRoot`·`codexAuditExec`·`codexAuditMCPDisableArgs`·`codexAuditReserveRecord`·`codexAuditNow` | codex_audit_launch.go | 아니다 | 런처 core. 기록은 `.moai/reports/codex-audit/`(`:59` `codexAuditReportsDir = ".moai/reports"`) |
| `codexAuditChildEnv` | codex_audit_leg.go:101 | 아니다 | HEAD에서 이미 `codexChildEnv`(codex_launcher.go:621, t1909 제거)에서 **분리 완료** — 분리 선례 |
| `withCodexAuditLeg` | codex_audit_leg.go | 아니다 | 호출자 `mcp_convergence.go:659` (audit_multi codex leg) |
| `handleCodexRoleAudit`/`Status`/`Result`·`codexServingDir`·`codexRoleAuditTools` | codex_audit_mcp.go | codexServingDir만 역방향 공유 | `codex_task.go:339`(t1909 제거)가 `codexServingDir`를 **읽는** 쪽 — 감사가 소유, 제거 시 소비자만 사라짐 |
| `newCodexJobID` | codex_jobs.go:522-528 | **예 (유일)** | 일반 호출자 `codex_jobs.go:240`(t1909 함께 삭제). 7줄 함수 — 감사 쪽으로 이동이 정답 |
| `handleCodexAudit`·`handleCodexSetup`·`codexLookPath`(:478)·`runCodexAuditReviewRPC`(:1049)·`ReviewOutput`(:247)·`inconclusiveReview`(:380)·`applyGateUnmet`(:2183) | mcp_codex.go (123K 혼재 파일) | **예** | 같은 파일에 t1909 제거 대상(codex_task/job 핸들러)이 섞여 있음 → **파일 분리 필수** |
| `runMultiAudit` codex leg | mcp_convergence.go:618-660 | 간접 | `codexLookPath`+`codexReviewRPC`를 mcp_codex.go에서 재사용(`:623` 주석 명시) |
| `configureClaudeAuditProcess` | mcp_claude_process_{unix,windows}.go | 아니다 | claude 감사 계열 — 잔존 |
| 감사 주변 계열 | mcp_audit_multi.go·mcp_convergence.go·mcp_audit.go·mcp_build_identity.go·audit_delta_scope.go·audit_plan_cmd.go·codex_receipt_detail.go·codex_audit_wait*.go(4) | 아니다 | 전부 잔존 |
| `codexCmd` 부모 명령 | **codex_launcher.go:646** | **예 — 최대 위험** | t1909가 codex_launcher.go를 지우면 컴파일 파손 |
| `template.EmbeddedTemplates` | internal/template/embed.go:33 | 아니다(공용 인프라) | `fs.Sub(embeddedRaw, "templates")` |
| 역할 TOML 12종 | templates/.codex/agents/moai/*.toml | **8종은 선택적 제거 대상** | 감사가 읽는 것은 4종뿐(§3) |
| 임베드 허용목록 | **embed_manifest_gen.go:553-564** (12줄) | 부분 | **생성 파일**(헤더: "Code generated by internal/template/embedemit; DO NOT EDIT", `make embed-manifest` 재생성, `make embed-manifest-check`가 build 앞에서 판정). `.claude` 지시문도 함께 담는 전체 트리 허용목록 |
| `agentemit` 패키지 | internal/template/agentemit/ (manifest·permission·emit·writer·loader) | t1909가 emitter 소비자 제거 시 감사만 남음 | 소비자 실측: codex_audit_launch.go·doctor_agentemit_embed.go·harness/rosterguard/registry.go 3곳 |
| `internal/codexadapter` | internal/codexadapter/ | **예 (공유)** | 감사 leg + t1909 파일들(codex_stop_chain·codex_readiness·doctor_codex·hook_*) + internal/codexwiring — **잔존** |
| codex_review_* 6종 | codex_review_{gate,delay,receipt,scope,tree_scope,binary_age}.go | 별도 전환(조사 미포함) | 감사 런처 심볼 참조 **0건**(실측 grep) — 단 `codexLookPath`는 3곳에서 호출(codex_review_delay.go:223·codex_review_gate.go:206·codex_review_receipt.go:133) → mcp_codex.go 분리 시 이들도 같은 심볼을 봐야 함 |

### 2.3 리더 조사 기록에 대한 보정 2건

1. **embed_manifest_gen.go는 codex 전용 소비자가 아니다.** 전체 템플릿 트리의 embed 허용목록(.claude 포함)을 담는 생성 파일이다. 통째 삭제 시 모든 템플릿이 바이너리에서 사라진다. codex 관련분은 `:553-564` 12줄뿐이다.
2. **codex_review_gate.go는 감사 런처와 독립** (codexAudit 심볼 참조 0건 실측) — 조사 기록의 "로컬 감사 의존 여부 조사" 질문에 대한 1차 답: 유지/제거 판단은 감사 의존이 아니라 별도 전환 결정으로 귀속된다.

## 3. 최소 keep 세트 · TOML 부분집합 · 등록 이동 설계

### 3.1 TOML 부분집합 — 실측

`agents-codex.yaml:484-489`:
```yaml
  default_sandbox: workspace-write
  role_sandbox:
    manager-todo: read-only
    super-advisor: read-only
    plan-auditor: read-only
    sync-auditor: read-only
```
감사가 읽는 TOML은 정확히 **4종**: `templates/.codex/agents/moai/{manager-todo,super-advisor,plan-auditor,sync-auditor}.toml` + 그 4줄의 `//go:embed`(embed_manifest_gen.go:561-564) + `agents-codex.yaml` 전체(파스가 fail-closed 자가검증이라 부분 편집 불가) + `codex_role_addenda`(:493-527, 같은 4역할만 담겨 일관). 나머지 8종(workspace-write)은 실행 자격이 없어 감사가 읽지 않고, t1906이 init/emit을 제거하면 생산 소비자가 없다 — **선택적 prune**(단, 삭제 시 `make embed-manifest` 재생성 필수, 12건 고정 테스트 갱신 필요 — 테스트면은 미측정 Gap).

### 3.2 최소 keep 세트 (role-audit 경로)

- **Go(cli)**: codex_audit_launch.go · codex_audit_mcp.go · codex_audit_leg.go · codex_audit_wait{,_load_darwin,_load_linux,_load_other}.go · `newCodexJobID`(이동) · `codexCmd` 부모(이동) · mcp_server.go의 role-audit 등록 3줄
- **Go(mcp 확장)**: codex_audit/audit_multi까지 keep 시 — mcp_audit_multi.go · mcp_convergence.go · mcp_audit.go · mcp_build_identity.go · audit_delta_scope.go · audit_plan_cmd.go · codex_receipt_detail.go · mcp_claude_process_{unix,windows}.go · mcp_codex.go에서 분리한 감사 부분(handleCodexAudit·handleCodexSetup·codexLookPath·runCodexAuditReviewRPC·ReviewOutput·inconclusiveReview·applyGateUnmet 등)
- **인프라**: internal/codexadapter · internal/template/agentemit 전체 · embed.go · embed_manifest_gen.go(≥4줄) · internal/config(workflow.audit 핀)

### 3.3 등록 이동 설계

`var codexCmd = &cobra.Command{...}`(codex_launcher.go:646)을 **잔존 파일**(권장: 신규 소형 `internal/cli/codex_cmd.go`)로 이동한다. t1909는 부모의 RunE(일반 실행 동작)만 걷어내고, 서브커맨드 부착(init :926-929)은 **codex_audit_launch.go 안에서 그대로** — 사용자 표면 `moai codex audit`·`moai codex role-audit` 불변. (거부한 대안: `moai audit ...` 등 새 부모로 재배치 — 모든 스크립트·교리·프롬프트의 기존 호출을 깨는 이득 없는 파손.)

## 4. 원자적 결합 옵션과 순서 제약

| | A — 원자 묶음 (t1907+t1909 한 change set) | B — 감사 분리 선행 (권장) |
|---|---|---|
| 내용 | 분리 이동과 일반 제거를 한 번에 | t1907을 순수 이동 리팩터로 먼저 착지: (1) codexCmd→codex_cmd.go (2) newCodexJobID→codex_audit_mcp.go (3) mcp_codex.go 감사 부분→mcp_codex_audit.go. t1909는 이후 순수 삭제 |
| 파손 창 | 없음 | 없음 (선행 착지 시점에 감사 이미 독립) |
| 검토 부담 | 큼 — 수백 KB 삭제+이동 혼재, 감사 회귀의 bisect 불가 | 작음 각각 — 이동은 기존 테스트(codex_audit_*_test.go·mcp_codex_test.go·감사 핀 테스트)가 회귀 게이트, 삭제는 착지 전후 감사 receipt 비교가 게이트 |
| 순서 제약 | 자명(한 커밋) | **엄격**: t1909를 t1907보다 먼저(또는 t1907 없이) 실행하면 codexCmd·newCodexJobID·handleCodexAudit 소멸로 컴파일 파손 |
| 폭발 반경 | 하나의 큰 PR | t1907 소형 PR 1개 + t1909 대형 삭제 PR 1개 |

선례: `codexChildEnv`→`codexAuditChildEnv` 분리가 정확히 이 방식(이동 후 제거 여지)으로 이미 착지돼 있다 — `git diff --stat 7705b365e2..HEAD -- codex_audit_launch.go` 실측.

## 5. 보존 측정 (제거 전 baseline — 유효 판정대로 채택)

- **명령**: MCP `codex_role_audit` — role=super-advisor, worktree_root=t1910 트리, out=`.moai/reports/t1910/preservation-baseline.md`
- **receipt**: job `job-20261020T091536Z-997d55c4` · launch record `.moai/reports/codex-audit/super-advisor-20261020T091538Z-35f8900d.launch.json` · route=mcp · sandbox=read-only · mcp_servers=disabled · exit_code=0 · 09:15:38Z→09:15:53Z
- **판정 파일 해시**: record `verdict_sha256 08e6580ca95e0e434ecec476c5b355f6c63944143082d7ba9f2fbada4991ace8` = `shasum -a 256` 재계산값 — 일치
- **원문 출력**:
  ```
  go version go1.26.8 darwin/arm64
  // Package decision implements the decision board: one append-only,
  ```
- **교차 검증**: `sed -n '1p' internal/decision/board.go` → 동일 행 · `go version` → go1.26.8 darwin/arm64 — 일치. **의존 경로 전체(LoadManifest 계약 → 임베디드 TOML → job id → 최상위 `codex exec -s read-only` → launcher verdict 기록)가 오늘 실제로 동작했다.**
- **도구 출처(§2.2 이중 좌표)**: 실행 주체 = 상주 MCP 서버 pid 15402, build `7705b365e2`(2026-10-18) — 트리 HEAD `8249bff8c3`의 **엄격 조상**(지연 명시). 지연 실횡성 실측: build..HEAD의 codex_audit_launch.go 변화는 env helper 이름 변경 1건뿐, 인용 구간(:497-542)은 동일. 설치 CLI는 `moai version` → `v3.2.0-rc.37  moai_cp/20261007_195927-2046-g8249bff8c3  built 2026-10-19T12:42:24Z` — HEAD 빌드.
- **재귀 게이트 제안**: t1909 착지 후 동일 프로브(role·worktree·out 관례 동일) 재실행, 두 receipt의 verdict sha+원문 비교를 acceptance 증거로.

## 6. 리더에게 남기는 열린 질문

1. workspace-write TOML 8종: t1909에서 즉시 prune(`make embed-manifest` 재생성+테스트 갱신 동반) vs 유지 후 별도 후속 — **후자 권장**(삭제 PR 폭반경 축소, 감사 영향 어느 쪽이든 0).
2. mcp_codex.go 분리 경계(§2.2 심볼 목록)를 t1907의 acceptance criteria로 확정할 것 — codex_review_* 6종(잔존, 별도 전환)도 `codexLookPath`를 호출하므로 분리 후에도 참조 가능해야 한다.
3. `codexServingDir` 소유 방향 확인 — 정의는 codex_audit_mcp.go:51(감사 소유), 소비 codex_task.go:339(t1909 제거) — 조치 불요, 기록 목적.
4. internal/codexwiring(5파일)의 귀속 — 소비자가 t1909 제거 대상(codex_init 등)과 잔존 codex_review_receipt.go 양쪽 — codex_review_* 별도 전환 결정에 따라가는 것으로 표기만 하고 이 카드에서 판정하지 않음.
5. 원자 묶음(A) vs 감사 선행(B) 최종 선택 — 본 문서는 B 권장. A 선택 시에도 §3.3의 이동은 묶음 내 선행 커밋으로 유지할 것.

## Gaps (명시적 미관측)

- go test/go build/golangci-lint 미실행 — 파견이 조사 전용이므로(§5 role_audit 1회가 유일한 실행). 컴파일 파손 주장(:646·:522 의존)은 소스 정적 판독에 근거하며 빌드 실패 재현으로 검증하지 않았다.
- `_status`/`_result` MCP 도구 미호출 — verdict 파일+launch record 직독으로 대체(실행 산출물 자체).
- 12건 TOML 고정 테스트(bundle/golden)·agentemit golden의 실태 미측정 — prune 리스크는 표기만.
- kill 규율: 패턴 kill(pkill/killall/kill $(pgrep)) 미사용 — 프로세스 종료 없음. 추적 작업은 TaskStop 원칙, 가드 거부는 최종(이번 턴 1건 거부 → 거부 형식 재시도 없이 분리 실행).

*Residual-risk: 행 번호는 8249bff8c3 고정 — local main이 추가 흡수돼 codex_audit 파일에 변화가 생기면 §1 인용을 재판독해야 한다(batch 5커밋은 감사 파일 무영향 실측으로 당장 무풍).*

---

## §8 Leader ruling on §6 open questions (2026-10-20, board-relayed)

1. workspace-write 8종 TOML — deferred to a separate follow-up approval; t1909 scope is exactly read-only 4 (manager-todo·super-advisor·plan-auditor·sync-auditor) + agents-codex.yaml + embed 4 lines.
2. mcp_codex.go separation boundary (§2.2 symbol list) — CONFIRMED as the implementation card's acceptance; codex_review_* 6종 keep their codexLookPath reference (condition documented).
3. codexServingDir — record-only, no action required.
4. codexwiring 5 files — follow the codex_review_* transition decision; OUT of this card's ruling.
5. Order — Option B (audit separation FIRST) CONFIRMED: t1909-first compiles broken + bisect-impossible rationale adopted; the §3.3 registration move (codex_cmd.go) stays a leading commit under B.
6. Implementation-tree base principle: `--base-branch release/main-batch-20261019` (@16d03deea3) + remote SHA + card traceability recorded. **Rationale corrected (operator measurement 2026-10-20)**: the basis is NOT "audit surface absent on origin/main" (false — origin/main 6dde1c1c1b carries prepareCodexAudit :168·LoadManifest :506·:927-928, and codex_audit_mcp.go·codex_jobs.go are SHA-identical across main/t1910/batch; the only divergent file is mcp_codex.go 121513B vs 123880B) — the actual basis is that t1909's removal targets (mcp_codex.go et al.) are diverging on the batch line, so removing AFTER the batch lands minimizes conflicts. The §preservation baseline measurement stays as read evidence with its stated limit (MCP server build 7705b365e2 ≠ target tree 8249bff8c3) and must NOT be widened into post-implementation-regression evidence.

t1909 card text update (B order + separation acceptance + base principle) is pending on the leader side; implementation starts only after that update.

7. **t1910 증명 범위 명시 (운영자 정정 ⑥)**: 이 카드가 증명하는 것은 **읽기 설계(의존 closure·최소 보존 집합·등록 이동)와 launcher smoke(codex_role_audit 1회 실행 receipt)까지**다 — 새 소스 구현에 대한 회귀 증거가 아니다. §preservation baseline의 MCP 서버 build ≠ 대상 tree 한계와 함께, t1909 구현 후에는 구현 트리 기준의 새 회귀 검증이 별도로 필요하다(확대 금지 원칙의 문서화).
