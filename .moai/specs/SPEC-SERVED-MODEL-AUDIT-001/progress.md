# Progress — SPEC-SERVED-MODEL-AUDIT-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-27
- card: t1282 (class C)
- tier: M — 산출물 spec.md, plan.md, acceptance.md, progress.md (편차 없음)
- REQ: 16건 (REQ-SMA-001..016) · AC: 16건 (AC-SMA-001..016) — Tier M 예산 안 (v0.2.0)
- plan-audit iter1: FAIL 0.75 (`.moai/reports/plan-audit/SPEC-SERVED-MODEL-AUDIT-001-review-1.md`). v0.2.0 반영:
  - D3 Tier 예산 → REQ 19→16, AC 25→16 통합(형제 부정 사례를 표 기반 AC 로). Tier L 승격 대신 통합을 택함 — 판단이 들어가는 독립 편집 단위는 4개 패키지이고 나머지 파일은 미러·기계 방출물(spec.md §D)
  - D2 공유 스코프 비확장 → 전용 술어(plan.md D3), REQ-SMA-013, AC-SMA-009
  - D4 해석 모델 출처 → 같은 `deps.Config` 주입(plan.md D4), REQ-SMA-003/004, AC-SMA-003
  - D1/D11 워크트리 slug → 열거 규칙 REQ-SMA-014, plan.md D5, AC-SMA-011, spec.md §A.2 측정 원문(primary 203 / 워크트리 slug 200)
  - D5 자기 보고 표면 → REQ-SMA-015(보고서 파일 + 최종 메시지), REQ-SMA-016(`last_assistant_message`), AC-SMA-014
  - D6 센티널·병합 사유 → REQ-SMA-010, AC-SMA-007
  - D7 AC 얕음 → AC-SMA-013 의무 문장 고정 grep + C2 중립성 부정 grep
  - D8 커버리지 → AC-SMA-004 `session_id`, AC-SMA-005 unknown 경고 행, AC-SMA-011 종료 코드
  - D9 빈 스캔 → REQ-SMA-014 info(ok 아님), AC-SMA-011 (b)
  - D10 레거시 kind → REQ-SMA-011, plan.md D2, AC-SMA-008; M7(규칙 문서 편집)은 sync 단계로 이관(spec.md §E)
- 근거: `.moai/reports/t1282/verdict.md`(착수 판정서) + spec.md §A.2 plan 단계 실측(M1-M4)
- 측정 트리: develop `b59a5d69c` 기준 워크트리 `WT-served-model-audit`
- 미확인 항목(spec.md §A.4): SubagentStop 실제 페이로드의 `agent_transcript_path` 채움 여부, 발화 시점 마지막 assistant 행 기록 여부, t1237·t1239·t1099 개별 대조, L2 워크트리 slug
- open_clarifications: 0 (리드 결정 (c) 로 설계 확정)
- run-phase 필수 단계: `make agents-emit` 후 `make agents-emit-check`, doctor 점검 이름의 `namesAddedAfterBaseline` 등록

## §E.2 Run-phase Evidence

- 측정 트리: 브랜치 `WT-served-model-audit`, HEAD `ea8bc6c23`(M6 커밋). 아래 모든 명령은 이 트리에서 실행했고, 원문 출력은 `.moai/state/verify/t1282/` 아래 파일에 있다(gitignore 대상, 로컬 전용).
- 환경 정리: 모든 `go test` 는 `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …` 한 번의 복합 호출로 실행했다.
- 마일스톤 커밋: M1 `69a7fbd4a` · M2 `9f3cc77f0` · M3 `389797697` · M4 `4c6f2ee89` · M5 `43a7b112f` · M6 `ea8bc6c23`

### AC 매트릭스

| AC | Status | 명령 | 관측 출력(요지는 파일 원문의 발췌) |
|----|--------|------|-----------------|
| AC-SMA-001 | PASS | `go test ./internal/hook/ -run '^TestServedModel_Classify$' -count=1 -v` | `--- PASS: TestServedModel_Classify` + 하위 4행(a~d) PASS, `ok` (ac001.txt) |
| AC-SMA-002 | PASS | `go test ./internal/hook/ -run '^TestServedModel_UnknownNeverOK$' -count=1 -v` | `--- PASS: TestServedModel_UnknownNeverOK` + 하위 8행(a~h) PASS (ac002.txt) |
| AC-SMA-003 | PASS | `go test ./internal/hook/ -run '^TestServedModel_ResolvedFromActiveProfile$' -count=1 -v` | 최상위 PASS + a/b/zero-config 대조 3행 PASS (ac003.txt) |
| AC-SMA-004 | PASS | `go test ./internal/hook/ -run '^TestServedModel_SubagentStopRecord$' -count=1 -v` | 최상위 PASS + 하위 a~e 5행 PASS (ac004.txt) |
| AC-SMA-005 | PASS | `go test ./internal/hook/ -run '^TestServedModel_GateOffWarnsOnly$' -count=1 -v` | 최상위 PASS + a/b/c(양성 대조) PASS (ac005.txt) |
| AC-SMA-006 | PASS | `go test ./internal/hook/ -run '^TestServedModel_GateOnAdoptionRefusal$' -count=1 -v` | 최상위 PASS + a/b/c PASS (ac006.txt) |
| AC-SMA-007 | PASS | `go test ./internal/hook/ -run '^TestServedModel_RefusalDeniesPhaseEntry$' -count=1 -v` | 최상위 PASS + a/b/c PASS (ac007.txt) |
| AC-SMA-008 | PASS | `go test ./internal/hook/ -run '^TestServedModel_ClearingIsKindScoped$' -count=1 -v` | 최상위 PASS + a/b PASS (ac008.txt) |
| AC-SMA-009 | PASS | `go test ./internal/hook/ -run '^TestServedModel_GateDoesNotArmReceiptGuard$' -count=1 -v` | 최상위 PASS + served_gate_only / control_codex_gate_required PASS (ac009.txt) |
| AC-SMA-010 | PASS | `go test ./internal/config/ -run '^TestServedModelGate_DefaultOff$' -count=1 -v`; 두 `grep -n -A1 'served_model_gate:'`; `go test ./internal/template/... -count=1` | 테스트 PASS(하위 4행); 템플릿 `168-        enabled: false`, 로컬 `200-        enabled: true`; 템플릿 3개 패키지 `ok` (ac010.txt, ac010-template.txt) |
| AC-SMA-011 | PASS | `go test ./internal/cli/ -run '^TestServedModelCheck_Sweep$' -count=1 -v` | 최상위 PASS + a/b PASS (ac011.txt). `-race` 로 한 번 더 실행해 exit 0, DATA RACE 0 (ac011-sweep-race.txt) |
| AC-SMA-012 | PASS | `go test ./internal/cli/ -run '^TestBinaryLag_.*$' -count=1 -v`; `grep -n '"servedModelCheckName"' internal/cli/binary_lag_test.go` | 4개 테스트 PASS, `ok`; `217:	"servedModelCheckName": true,` (ac012.txt) |
| AC-SMA-013 | PASS | `make agents-emit-check`; 6개 파일 `grep -cF 'The first line of the report file and of your final message MUST be' …`; C2 중립성 `grep -hF … \| grep -cE 'SPEC-[A-Z]\|\bt[0-9]{3,}\b'`; `go test ./internal/template/agentemit/... -count=1` | agents-emit-check exit 0; 6개 파일 모두 `:1`; 중립성 명령은 stdout `0` 을 출력했다(O3: 개수가 0이면 grep 이 exit 1 을 내므로 판정은 출력된 `0` 으로 한다); agentemit `ok` |
| AC-SMA-014 | PASS | `go test ./internal/hook/ -run '^TestServedModel_SelfReportFromFinalMessage$' -count=1 -v` | 최상위 PASS + a/b/c PASS (ac014.txt) |
| AC-SMA-015 | PASS | `go test ./internal/hook/... -count=1`; `go test ./internal/auditreceipt/... ./internal/config/... -count=1`; `golangci-lint run ./internal/hook/... ./internal/auditreceipt/... ./internal/config/... ./internal/cli/... ./internal/escalation/...` | hook 11개 패키지 전부 `ok`(`internal/hook 406.750s`); auditreceipt·config 3개·escalation `ok`; lint `golangci-lint has version v2.1.6` / `0 issues.`. `git diff --name-status 2c997966c..HEAD -- '*_test.go'` 는 신규(A) 7개와 수정(M) 1개 `internal/cli/binary_lag_test.go`(AC-SMA-012 가 요구하는 허용 목록 1줄)뿐 — 영수증 가드 테스트는 수정 없음. doctor golden 3개는 재생성 |
| AC-SMA-016 | PASS (개발 머신 한정, 릴리스 게이트 아님 — O1) | `go run ./cmd/moai doctor --check "Served Model" --verbose` | exit 0; `swept 2291 subagent transcripts: ok 1274, served_drift 960, unknown 11, unmapped 46; 143 gate-auditor run(s) flagged`; `manager-develop d46e0166-e049-410d-af03-ee8b87742239/agent-a5a593b7e4697e72c served_drift expected=opus served=[glm-5.3-flash]`. 같은 세션의 PreToolUse 행(primary `.moai/logs/agent-model-audit.jsonl`)은 그대로 `"verdict":"ok"` 1행(소급 기록 없음) (ac016-doctor.txt) |

### RED 증거 (테스트 우선)

- M1: `m1-red.txt` — `undefined: ServedVerdictOK`, `undefined: ObserveServedModel`, `undefined: servedReadBudget` (구현 전 컴파일 실패)
- M2: `m2-red-config.txt` — `WorkflowConfig has no field or method ServedModelGate`; `m2-red-hook.txt` — `undefined: NewSubagentStopHandlerWithConfig`
- M3: `m3-red.txt` — `served row lacks key "agent_id"`, `self_reported_model = <nil>, want "opus"` (동작 RED)
- M4: `m4-red.txt` — `served refusals = [], want exactly 1`, `denied=false … want a deny starting with SERVED_MODEL_VIOLATION`, `reason … must carry both sentinels`, `served refusals after a served ok = [{… Kind:served}], want none` (동작 RED)
- M5: `m5-red.txt` — `undefined: runServedModelScan`; AC-SMA-012 는 허용 목록 등록 전 `m5-mid.txt` 에서 `--- FAIL: TestBinaryLag_DoctorCheckNameSetIsUnchanged` (`this SPEC added doctor check name servedModelCheckName`)
- M6: `ac013-red.txt` — 6개 파일 모두 `:0`

### 빌드·vet·교차 빌드

- `make build` exit 0 (build.txt; `agents-emit-check`·`commands-emit-check`·카탈로그 해시 재생성 후 `go build … -o bin/moai`)
- `go vet ./internal/hook/... ./internal/auditreceipt/... ./internal/config/... ./internal/cli/... ./internal/escalation/...` exit 0, 출력 없음
- `GOOS=windows GOARCH=amd64 go build ./...` exit 0

### 커버리지 (부분 측정)

- `go test ./internal/hook/ -run 'ServedModel' -coverprofile=…` 기준 신규 함수: `expectedServedModel` 100%, `readServedModels` 96.8%, `observeServedModel`(분류) 95.7%, `checkServedModelSpawn` 93.8%, `checkServedModelStop` 81.8%, `servedGateScope` 66.7%
- cli: `runServedModelScan` 95.5%, `observeServedTranscripts` 100%. `checkServedModel`·`servedScanInputsFor` 는 이 셀렉터로 0% — doctor golden 테스트가 실행한다
- 패키지 전체 커버리지는 측정하지 않았다(Gap)

### 설계 결정·해석 기록 (sync·감사 판독용)

- **expectedServedModel 단일 결합점**: `internal/hook/served_model.go:124` (`@MX:ANCHOR` 122행). SubagentStop 관측기와 doctor 스윕 모두 `ObserveServedModel` → `observeServedModel`(235행) → 이 함수 한 곳을 지난다. 이외의 호출자는 없다(`grep -rn "expectedServedModel(" internal/` 비테스트 결과가 정의 1행과 호출 1행).
- **서빙 행의 기록 루트**: `CLAUDE_PROJECT_DIR`(+ `.moai` 존재)만 쓰고 hook 입력 CWD 로 폴백하지 않는다. 같은 spawn 의 PreToolUse 행이 그 루트에 쌓이므로, 워크트리 세션에서 CWD 로 폴백하면 두 행이 다른 파일로 갈라진다. 이 선택 덕분에 기존 `TestWSR009_GuardFailsClosedWithoutStore`(W/.moai 에 아무것도 쓰지 않음)를 수정 없이 통과한다.
- **트랜스크립트 위치 정보가 전혀 없는 페이로드**: `agent_transcript_path` 도, `transcript_path`+`agent_id` 도 없으면 관측할 서브에이전트 실행이 특정되지 않으므로 행·경고를 모두 내지 않는다. **[정정 — repair 커밋]** 이전 판은 "런타임이 `transcript_path` 와 `agent_id` 를 항상 보낸다"고 적었으나 이 머신에서 관측한 적이 없는 주장이었다. 실측 시도 결과(아래 "트랜스크립트 위치 증거")로는 확립할 수 없다 — 그런 페이로드가 실제로 오면 그 정지는 `unknown` 으로 기록되지 않고 관측에서 빠진다(잔여 위험). 이 해석은 기존 `TestAuditReceiptGuard_NonRequiredTreesAreInert`·`TestSubagentStop_ReentryWarnsAcceptanceClearsRoleFailIsInert` 가 요구하는 "무출력" 을 수정 없이 지킨다. REQ-SMA-002/005 문구상의 예외이므로 sync 단계에서 manager-spec 이 문구 반영 여부를 판단할 것.
- **AC-SMA-005 배치**: plan 은 M3 로 적었으나 (c) 양성 대조 행이 M4 의 served-kind 거부를 필요로 해 M4 커밋에 넣었다.
- **도구 부수 편집(범위 내 cascade)**: `internal/config/cache.go` 스키마 버전 8→9(필드 추가 관례), `internal/config/testdata/shipped_key_inventory.yaml` 에 `workflow.served_model_gate.enabled`(W) 등록(`TestShippedConfigKeysHaveReaders` 요구), doctor golden 3개에 info 1행 추가, `internal/template/catalog.yaml` 의 두 감사관 해시 재생성, `escalation.ClaudeConfigBases` export(기존 `memoryRoots` 가 같은 함수를 쓴다).
- **doctor 스윕 성능**: 이 머신 2291개 트랜스크립트(약 3.8GB)에 대해 워커 4개로 실측 약 5.7초(`/usr/bin/time -p` real 5.72). 순차 구현일 때 10.23초였다.

### Repair (리드 결정, sync 착수 전)

- **doctor "Served Model" 은 명시 호출 전용**: 기본 `moai doctor` 실행은 스윕하지 않고 info 행 하나(`Served Model  run with --check "Served Model"`)만 낸다. 스윕은 `moai doctor --check "Served Model"` 일 때만 돈다(`servedModelDoctorEntry`, `internal/cli/doctor_served_model.go`). doctor.go 에 "명시 호출 전용" 관례는 없었고(`filterCheck` 는 이름 일치 필터일 뿐), 해당 없음 행을 info 로 내는 관례(Codex Wiring 정보성 skip, claude surface 강등 info 행)를 따랐다. 점검 이름은 `moaiChecks` 에 그대로 등록돼 있으므로 `namesAddedAfterBaseline` 의 `servedModelCheckName` 은 유지한다(`TestBinaryLag_DoctorCheckNameSetIsUnchanged` 요구). doctor golden 3개는 hint 문구로 재생성. REQ-SMA-014 "When `moai doctor` runs" 문구와의 정렬은 manager-spec 몫.
- **트랜스크립트 위치 증거**: 이 머신에서 SubagentStop 실제 페이로드에 `agent_transcript_path` 또는 `transcript_path`+`agent_id` 가 담기는지를 보여주는 기록을 찾지 못했다 — 확립 불가(Gap). 확인한 곳: primary `.moai/logs/agent-model-audit.jsonl` 의 `"source":"subagent_stop"` 행 0건(설치 바이너리에 신규 코드 없음), `.moai/logs/*` 전체에서 `transcript_path` 문자열 0건, `hook-runtime.log` 는 WARN 이상만 남아 `subagent stopped` INFO 행이 없음, `trace-*.jsonl` 의 SubagentStop 행은 `event/handler/duration_ms/session_id` 만 기록, 세션 트랜스크립트 `d46e0166…jsonl` 에 `agent_transcript_path` 0건. 측정 경로 후보: 새 바이너리 설치 후 실제 서브에이전트 1회 정지 → primary 감사 로그의 `subagent_stop` 행 유무(행이 생기면 위치 정보가 있었다는 증거, 없으면 부재 또는 루트 미설정).

### Deferred notes (sync 단계 manager-spec 몫)

- **O2**: REQ-SMA-002 의 "exactly one served-observation row" 를 "one row per SubagentStop event" 로 바꾼다. 영수증 가드가 첫 정지를 block 한 감사관은 다시 정지하므로 한 인스턴스가 여러 행을 낼 수 있다. spec.md 본문 편집이라 run 단계에서 하지 않았다.
- 위 "트랜스크립트 위치 정보가 전혀 없는 페이로드" 예외의 REQ 문구 반영 여부.
- `.claude/rules/**` 의 감사 로그 설명 갱신(spec.md §E Out of Scope — sync 몫).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-27
run_commit_sha: ea8bc6c23   # 증거를 측정한 HEAD(M6). 이 progress.md 커밋은 그 뒤에 온다
run_status: audit-ready
ac_pass_count: 16
ac_fail_count: 0
ac_gap_count: 0            # AC-SMA-016 은 이 머신에서 실측 PASS(개발 머신 한정, 릴리스 게이트 아님)
preserve_list_post_run_count: 0   # 영수증 가드 기존 테스트 수정 0 (binary_lag_test.go 허용 목록 1줄은 AC 요구)
l44_pre_commit_fetch: not-run    # 레인은 push 하지 않는다(리드 일괄)
l44_post_push_fetch: not-applicable
new_warnings_or_lints_introduced: 0   # golangci-lint v2.1.6 "0 issues."
cross_platform_build:
  darwin: pass      # make build exit 0
  windows_amd64: pass   # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 38   # 2c997966c..HEAD 37개 + 이 progress.md
m1_to_mN_commit_strategy: one commit per milestone (M1..M6) + this evidence commit
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-27
sync_commit_sha: pending-backfill   # 자기 SHA 는 커밋 안에 적을 수 없다 — `git log -1 --format=%h -- .moai/specs/SPEC-SERVED-MODEL-AUDIT-001/progress.md` 로 읽는다
sync_status: audit-ready
b12_self_test_a: pass   # pre-emission `grep -c 'SERVED-MODEL-AUDIT-001' CHANGELOG.md` → 0
b12_self_test_b: pass   # distinct AC ids in acceptance.md → 16; CHANGELOG entry cites AC-SMA-001..016 (16)
b12_self_test_c: pass   # ls: spec.md, internal/hook/served_model.go, internal/cli/doctor_served_model.go, docs-site/content/{ko,en,ja,zh}/multi-llm/model-policy.md 모두 존재
changelog_entry_position: "[Unreleased] › ### Added › 첫 항목"
docs_synced:
  - docs-site/content/{ko,en,ja,zh}/multi-llm/model-policy.md   # § 감사 기록과 fail-open 에 서빙 행·게이트·스윕 문단 (4개 로케일 동시)
docs_checked_no_change:
  - README{,.ko,.ja,.zh}.md   # doctor 점검·workflow.* 키 서술 없음
  - .claude/rules/**          # 감사 로그 언급은 agent-common-protocol.md:168 (§ Per-Spawn Model Injection) 한 곳 — t1246 소유라 미편집
frontmatter_status_transitions:
  spec.md: "in-progress → completed (implemented 병합, 단일 sync 커밋)"
  plan.md: "frontmatter 없음 — 변경 없음"
  acceptance.md: "frontmatter 없음 — 변경 없음"
  progress.md: "frontmatter 없음 — 변경 없음"
deferred_notes_resolved_by: f54b67c38   # 커밋 메시지 기준: REQ-SMA-002 행 수 문구·위치 정보 없는 페이로드 예외·REQ-SMA-014 명시 호출 전용 (본문 diff 는 sync 에서 재대조하지 않음)
open_gap: "SubagentStop 실제 페이로드의 트랜스크립트 위치 정보 — 미측정 (spec.md §A.4)"
```
