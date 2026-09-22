# SPEC-CODEX-AUDIT-GATE-AXES-001 — 인수 기준

카드 **t686** · v0.3.1. 각 AC 는 명령과 관측 출력으로 판정한다. 축 (c)는 카드 t870 으로 분리되어 이 문서에 없다. 판정 줄 문법·저장 위치·스키마는 plan.md §B.5-§B.6 을 따른다.

## §D AC Matrix

### 축 (a) — 양방향 회귀

#### AC-CAG-001 — required + 바이너리 부재 → 차단
- **Given** 감사 대상 트리의 `workflow.yaml` 에 `workflow.audit.gates.codex: required` 가 쓰여 있고 codex 바이너리가 PATH 에 없다
- **When** `codex_audit` 를 그 트리의 `project_root` 로 호출한다
- **Then** 결과는 `isError: false`, `verdict == "fail"`, `gate_unmet` 비어 있지 않음, `summary` 에 required 게이트 미충족과 "codex binary not found" 원인이 함께 있다

#### AC-CAG-002 — required + RPC 실패 / 빈 출력 → 차단
- **Given** AC-CAG-001 과 같은 설정이고, codex 는 있으나 (i) RPC seam 이 fail-open inconclusive 를, (ii) 세션 스크립트가 공백뿐인 리뷰 본문을 돌려준다
- **When** 각 경우 `codex_audit` 를 호출한다
- **Then** 둘 다 AC-CAG-001 과 같은 차단 결과이며 `summary` 에 각 원인이 보존된다. `TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput` 은 `gate_unmet` 단언을 유지하고 verdict 단언만 `fail` 로 바뀐다

#### AC-CAG-003 — 비-required + inconclusive → 정규화 후 바이트 동일
- **Given** `gates.codex` 가 `off`, `advisory`, 키 부재, 설정 파일 부재, YAML 손상인 다섯 트리와, 변경 전 코드에서 캡처해 `build_commit`·`build_lag` 를 고정 자리표시자로 정규화한 골든 파일
- **When** 각 트리에서 바이너리 부재 조건으로 `codex_audit` 를 호출하고 같은 정규화를 적용한다. 그리고 `git log --diff-filter=A --format=%H -- <골든 파일> | tail -1` 로 G 를 기록하고 `git merge-base --is-ancestor <G> <축 (a) 구현 커밋>; echo $?` 를 실행한다
- **Then** 다섯 경우 모두 직렬화 JSON 이 골든과 바이트 동일하고(`verdict == "inconclusive"`, `gate_unmet`·영수증 필드 없음), 조상 검사는 `0` 을 출력한다. 이 AC 는 축 (b) 착지 후에도 재실행해 통과한다

#### AC-CAG-004 — 배포 기본값은 opt-in 이 아니다
- **Given** `workflow.yaml` 에 `audit` 블록이 없는 트리
- **When** (i) 같은 트리에 대해 엔진 기본값이 적용된 codex 게이트를 조회하고, (ii) 바이너리 부재 조건으로 `codex_audit` 를 호출한다
- **Then** (i) 은 `required` 이고(전제 단언 — 아니면 이 AC 는 실패 처리), (ii) 는 `verdict == "inconclusive"` 이며 `gate_unmet` 이 없다

#### AC-CAG-005 — 실제 판정은 건드리지 않는다
- **Given** `gates.codex: required` 이고 codex 세션 스크립트가 각각 pass 와 fail 을 돌려준다
- **When** `codex_audit` 를 호출한다
- **Then** verdict 는 각각 `pass`, `fail` 이고 `gate_unmet` 은 비어 있다

#### AC-CAG-006 — 다중 경로·Stop-hook 판정 무변경
- **Given** 기준 커밋 `f67d2193f`, regex `R='Converge_|RunMultiAudit_|AuditMulti_|ReviewGate_|MultiReviewGate|PerformCodexAudit_|LoadConvergenceResult_'`, 그리고 `grep -lE "^func Test[A-Za-z0-9_]*($R)" internal/cli/*_test.go` 가 기준 커밋에서 돌려주는 14개 파일: `codex_protocol_liveness_test.go`, `codex_review_gate_live_test.go`, `codex_review_gate_test.go`, `codex_review_gate_wiring_test.go`, `codex_rpc_error_test.go`, `codex_verdict_divergence_test.go`, `mcp_audit_multi_test.go`, `mcp_convergence_participant_test.go`, `mcp_convergence_state_root_test.go`, `mcp_convergence_test.go`, `mcp_project_root_codex_test.go`, `multi_review_gate_test.go`, `multi_review_gate_wiring_test.go`, `required_gate_block_test.go` (모두 `internal/cli/`)
- **When** (i) `go test -list "$R" ./internal/cli/`, (ii) `go test -run "$R" -count=1 ./internal/cli/`, (iii) `git diff --stat f67d2193f..HEAD -- <위 14개 파일>`, (iv) 변경 후 트리에서 같은 `grep -lE` 를 다시 실행한다
- **Then** (i) 의 목록에 `TestCodexAudit_RequiredGateUnmetRecordedOnInconclusive` 와 `TestCodexBlankReview_AC007_RequiredGateAnnotatesBlankOutput` 이 없고, (ii) 는 `ok`, (iii) 의 출력은 비어 있으며, (iv) 가 새 파일을 돌려주면 그 파일은 이 SPEC 이 추가한 신규 테스트 파일이다(기존 파일 목록은 14개 그대로)

#### AC-CAG-007 — 서술 정합
- **Given** 변경 후 트리
- **When** `codex_audit` 도구 설명(`internal/cli/mcp_server.go`), 두 감사자 본문(로컬·템플릿), `moai-ref-cross-model-audit/SKILL.md`(로컬·템플릿)를 판독하고 `make agents-emit-check` 를 실행한다
- **Then** 일곱 곳 모두 "명시적 required 의 무판정은 단일 도구에서도 `verdict: fail` + `gate_unmet`"을 서술하고, 스킬 문서에 "convergence engine reuses" 류 문구가 없으며, 감사자 본문 4곳에 plan.md §B.6 판정 줄 의무가 있고, `make agents-emit-check` 가 exit 0 이다

### 축 (b) — 영수증 (B-1, 표면 S1 + PreToolUse 소비자)

#### AC-CAG-008 — 영수증 기록과 필드 노출 조건
- **Given** `t.TempDir()` 로 만든 required 트리 R 과 `advisory` 트리 A
- **When** 각 트리를 `project_root` 로 `codex_audit` 한 번, `audit_multi`(codex 참여) 한 번 호출한다
- **Then** 두 트리 모두 `<root>/.moai/state/audit-receipts/receipts/` 에 호출당 1건씩 plan.md §B.5 스키마 레코드가 생기고(`tree_root` 는 `EvalSymlinks` 정규화 값, `root_source: argument`), R 의 두 결과에는 그 `receipt_id` 와 같은 영수증 필드가 있으며, A 의 두 결과에는 영수증 필드가 없다

#### AC-CAG-009 — 시작 표식
- **Given** required 트리의 하위 디렉터리를 `cwd` 로 가진 SubagentStart 입력 세 개: `agent_type` 이 `plan-auditor`, `sync-auditor`, `manager-develop`
- **When** SubagentStart 핸들러를 호출한다
- **Then** 앞의 둘에 대해서만 `starts/<agent_id>.json` 이 생기고, `tree_root` 는 git toplevel 을 정규화한 값이며 `started_at` 은 RFC 3339 UTC 다. 세 번째는 파일을 만들지 않는다

#### AC-CAG-010 — 첫 종료 차단과 거부 영속
- **Given** required 트리, `plan-auditor` 시작 표식(started_at = T), 저장소에 영수증 r1(같은 트리, created_at > T), r2(다른 트리), r3(같은 트리, created_at < T)
- **When** `stop_hook_active: false` 인 SubagentStop 을 마지막 줄이 각각 (i) `AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=none`, (ii) `... receipts=rcpt-<저장소에 없는 id>`, (iii) `... receipts=<r2>`, (iv) `... receipts=<r3>`, (v) 시작 표식 없는 `agent_id` 로 `... receipts=<r1>` 인 메시지로 호출한다
- **Then** 다섯 경우 모두 출력이 `decision: "block"` 이고 `reason` 이 각각 인용 없음 / 저장소에 없음 / 다른 트리 / 시작 이전 / 시작 표식 없음을 이름으로 밝히며, `rejections/plan-auditor--SPEC-X-001.json` 이 원인과 함께 존재한다. `agent_type: sync-auditor` 로 바꿔도 같다(파일명 `sync-auditor--SPEC-X-001.json`)

#### AC-CAG-011 — 재진입 경고, 수용 시 해제, FAIL 무처리
- **Given** AC-CAG-010 (i) 이후 `rejections/plan-auditor--SPEC-X-001.json` 이 있고, 같은 트리에 `rejections/plan-auditor--SPEC-Z-001.json`, `rejections/plan-auditor--unknown-spec.json`, `rejections/sync-auditor--SPEC-X-001.json` 도 있는 상태
- **When** (i) AC-CAG-010 (i) 과 같은 메시지로 `stop_hook_active: true` 인 SubagentStop 을 호출하고, (ii) 새 시작 표식 이후 `AUDIT-VERDICT: PASS spec=SPEC-X-001 receipts=<r1 과 같은 조건의 영수증>` 으로 plan-auditor 의 `stop_hook_active: false` SubagentStop 을 호출하고, (iii) 별도 트리에서 `AUDIT-VERDICT: FAIL spec=SPEC-Y-001 receipts=none` 으로 호출하고, (iv) 별도 required 트리에서 거부 기록 파일을 사람이 삭제한 뒤 `manager-develop` PreToolUse 를 보낸다
- **Then** (i) 은 `decision` 이 없고 `systemMessage` 에 PASS 미수용과 단계 진입 스폰 거부 유지가 명시되며 거부 기록이 남고 `reentry_warned: true` 다. (ii) 는 `decision` 이 없고 plan-auditor 역할의 세 기록(`SPEC-X-001`, `SPEC-Z-001`, `unknown-spec`)이 모두 제거되며 `sync-auditor--SPEC-X-001.json` 은 남는다. (iii) 은 출력·기록 변화가 없다. (iv) 는 이 가드로 deny 되지 않는다

#### AC-CAG-012 — PreToolUse 소비자
- **Given** 거부 기록 1건이 있는 required 트리 R, 거부 기록이 없는 required 트리 C, 거부 기록 파일을 수동으로 둔 `advisory` 트리 A
- **When** 각 트리를 `cwd` 로 `tool_name: Agent` PreToolUse 입력을 `subagent_type` 이 `manager-develop`, `manager-docs`, `manager-git`, `Explore` 인 경우로 보낸다(`tool_name: Task` 도 한 번)
- **Then** R 에서 앞의 세 스폰(Agent·Task 모두)은 deny 이고 reason 이 `AUDIT_RECEIPT_VIOLATION` 으로 시작하며 역할·SPEC·원인을 담는다. R 의 `Explore` 와 C·A 의 모든 스폰은 이 가드로 deny 되지 않는다

#### AC-CAG-013 — 비-required 무영향과 시험 격리
- **Given** `gates.codex` 가 `off` / `advisory` / 부재인 세 트리
- **When** plan-auditor 에 대해 SubagentStart, 그리고 `receipts=none` PASS 로 SubagentStop 을 호출한다. 이어서 누출 검사를 한다: (a) `find . -path ./.git -prune -o -type d -name .moai -print | sort` 결과를 전후로 비교하되, 그 사이에 이 SPEC 이 추가·변경한 테스트만 `-run` 선택자로 실행한다(`go test ./internal/cli/ -run '<신규 영수증·codex_audit 테스트 이름>' -count=1 -v` 와 `go test ./internal/hook/ -run '<신규 SubagentStart·SubagentStop·가드 테스트 이름>' -count=1 -v`), (b) 같은 비교를 `git status --porcelain --ignored` 전후 차이로도 한다
- **Then** 세 트리 모두 시작 표식·거부 기록 파일이 없고 출력에 `decision`·`systemMessage` 가 없다. 두 `-v` 실행의 `=== RUN` 줄 수가 각각 1 이상이고(빈 스윕이 아님), (a)(b) 전후 차이가 모두 비어 있다(gitignore 된 `.moai/` 누출도 잡는다). 검사 자체가 유효하다는 증거로, 저장소 트리 안에 `.moai/state/audit-receipts/` 를 일부러 쓰는 임시 테스트를 한 번 실행해 (a)(b) 가 차이를 보고하는 것을 관측하고 그 출력을 progress.md §E.2 에 기록한 뒤 그 임시 테스트를 제거한다. `internal/cli`·`internal/hook` 전체 스위트는 로컬에서 실행하지 않는다(전체 판정은 리드 push 후 CI)

#### AC-CAG-014 — 판독 불가
- **Given** required 트리에서 (i) 인용 영수증 JSON 손상, (ii) `last_assistant_message` 공백, (iii) 판정 줄 형식 불일치, (iv) 거부 기록 JSON 손상, (v) `rejections/` 디렉터리 부재, (vi) 판정 줄 끝에 공백과 `\r` 이 붙은 유효 PASS(유효 영수증 인용)
- **When** (i)-(iii)·(vi) 은 plan-auditor SubagentStop(`stop_hook_active: false`)으로 처리하고, (ii) 직후 같은 트리에서 `manager-develop` PreToolUse 를 보내며, (ii) 와 같은 입력을 `stop_hook_active: true` 로 다시 보낸 뒤 기록을 확인한다. (iv)-(v) 는 `manager-develop` PreToolUse 로 처리한다
- **Then** (i)-(iii) 은 `decision: "block"` 이고 reason 이 각각 영수증 판독 불가 / 판정 줄 없음 / 판정 줄 없음을 이름으로 밝힌다. (ii)(iii) 은 `rejections/plan-auditor--unknown-spec.json` 을 남기고, (ii) 직후의 `manager-develop` 스폰은 `AUDIT_RECEIPT_VIOLATION` 으로 deny 되며 reason 에 `unknown-spec` 이 있다. `stop_hook_active: true` 재진입은 차단하지 않고 그 기록을 유지한다. (iv) 는 deny 이고 손상 파일 경로를 이름으로 밝힌다. (v) 는 이 가드로 deny 되지 않는다. (vi) 은 `decision` 이 없다(끝 공백·`\r` 을 판정 줄 누락으로 읽지 않음)

## §D.1 Edge Cases

- `gates.codex: "required "`(공백)나 `REQUIRED` — 정확 일치를 유지하며 비-required 로 취급함을 테스트로 고정.
- `project_root` 없이 호출돼 폴백 루트로 기록된 영수증(`root_source: fallback`) — 워크트리 감사자에게는 "다른 트리"로 거부된다.
- `audit_multi` 가 codex 를 제외하고 돈 경우 — 영수증을 남기지 않거나, 남겨도 검사 통과 조건 (4)를 만족하지 않는다.
- 판정 줄 뒤에 빈 줄만 있는 경우 — 마지막 비어 있지 않은 줄 기준이므로 인식된다.
- `PASS-WITH-DEBT` — PASS 부류로 검사된다.

## §D.2 Quality Gate

- `go vet` / `golangci-lint` 대상 패키지(`internal/cli`, `internal/hook`, 신규 저장소 패키지) 0 건
- 신규·변경 코드 커버리지 85% 이상
- 템플릿 중립성 가드 통과, `make agents-emit-check` exit 0

## §D.3 Definition of Done

- AC-CAG-001 ~ 014 PASS
- M1 페이로드 실측 기록이 progress.md §E.2 에 있음(정지 규칙 발동 시 리드 보고 기록)
- 미러 쌍·방출본 동기화, `make build` 성공
- 이슈 댓글·종료는 DoD 아님(리드 몫)
