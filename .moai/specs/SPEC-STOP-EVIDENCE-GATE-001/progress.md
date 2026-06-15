# SPEC-STOP-EVIDENCE-GATE-001 — Progress

## §0 Plan-phase Audit-Ready Signal

| Field | Value |
|-------|-------|
| plan_complete_at | 2026-06-15 |
| plan_status | audit-ready |
| tier | M |
| cycle_type (run-phase) | tdd |
| artifacts | spec.md + plan.md + acceptance.md + design.md + research.md (Tier M baseline 3 + design.md for the data-model fork + research.md for codebase grounding) |
| REQ count | 12 (REQ-SEG-001..012; REQ-SEG-012 = falsifiable-value/dormant-scaffold framing added in amend-1) |
| AC count | 11 (AC-SEG-001..011, all MUST; REQ-SEG-012 covered by §A.3/§A.4 doc + AC-SEG-008 dormant fixture) |
| status | draft |

## §0.1 Plan Audit Gate (Phase 0.5) — iter-1 + amend-1

| Field | Value |
|-------|-------|
| Auditor | plan-auditor (iter-1) |
| Verdict | PASS-WITH-DEBT |
| Score | 0.82 (Tier M threshold 0.80; not skip-eligible <0.90) |
| Decision | 사용자 선택 = "amend the plan, then re-audit" (run-phase 미진입) |

amend-1 (artifact editing only, Tier M 불변, no code) — 5 defects 해소:
- **D1 (SHOULD-FIX, borderline-BLOCKING)** — dormant scaffold framing + falsifiable value condition. (1) spec.md §A.3 knowingly-dormant scaffold (production writer `logSkillUsage` 하드코딩 `Outcome: OutcomeUnknown`/`Phase: "none"` → 게이트 현재 미발화) + scaffold-honest 선례 인용. (2) §A.4 + REQ-SEG-012: record-time-writer 후속 SPEC `SPEC-STOP-EVIDENCE-WRITER-001` 명명 + falsifiable 조건 ((a) PathKind=code-change + (b) Outcome=success + IsTestPass). (3) §A.1 originating-defect 재framing: `L_manager_docs_false_backfill_report`는 sync-phase(docs-only EXEMPT)이라 본 게이트 미탐지 — 본 게이트는 별개의 code-session false-success 형태 겨냥. research.md §2.1 ground-truth 추가.
- **D2 (SHOULD-FIX)** — escaped-pipe grep 수정 (`-E`에서 `\|`→`|`, dots escaped). AC-SEG-001 / AC-SEG-006(2 greps) / AC-SEG-008 / AC-SEG-011. AC-SEG-007은 basic grep이므로 `\|` UNCHANGED (의도적).
- **D3 (MINOR)** — AC-SEG-008(a) heavy-op grep을 stop.go `runEvidenceGate` body까지 확장 (a-2): awk로 gate body 추출 + heavy-op 0 + `LoadBySession`만 단언.
- **D4 (MINOR)** — AC-SEG-002 weak escaped-pipe grep(`IsTestPass\|IsTestFail >= 2`) 제거; binary-flip 동작 subtest delta(IsTestPass false→true가 verdict finding→nil flip)로 대체.
- **D5 (MINOR)** — REQ-SEG-002 reword: `Event.IsTestPass`→`UsageRecord.IsTestPass` (Event는 JSONL 미영속, 게이트가 read하는 것은 design §3 신규 UsageRecord 필드).

## §A Plan summary

직접 후속: SPEC-HOOK-DISCIPLINE-WIRING-001 (IMP-01, completed, Mx `8a9c1062f`). 본 SPEC = IMP-02(Stop-hook evidence gate) + IMP-03(session ledger reader) 묶음. IMP-06(baseline-integrity attribution + 5-section report + "no unobserved-verification-claim" invariant)는 명시 OUT OF SCOPE.

Motivating defect class: `L_manager_docs_false_backfill_report` — 에이전트가 관측하지 않은 검증/완료를 주장한 사건. **단, 그 originating incident는 sync-phase(manager-docs)이며 본 게이트의 path-kind 분류상 docs-only로 EXEMPT** → 본 게이트는 그 사건을 그대로 잡지 않는다 (amend-1 D1). 본 게이트가 겨냥하는 것은 별개의 **code-session false-success** 형태(code-change 세션에서 success 기록 + observed test-pass 없음)다. session-stop 시점 advisory 표면화(detection only; invariant codification은 IMP-06).

## §A.2 [HARD] Dependency note — value blocked on record-time writer (D1)

[HARD] 본 게이트는 현재 텔레메트리 스트림에 대해 **knowingly-dormant**다. 유일한 production `UsageRecord` writer `logSkillUsage`(`internal/hook/post_tool_metrics.go:72-82`)가 `Outcome: OutcomeUnknown` / `Phase: "none"`를 하드코딩하므로 `hasSuccessClaim`은 항상 false → 게이트 미발화이고, `inferPathKind`는 사실상 모든 실제 세션에 `unknown` 반환. `HookInput`(types.go:180)에는 `Phase` 필드가 없다.

게이트의 표적 결함(code-session false-success) 탐지는 **record-time-writer 후속 SPEC `SPEC-STOP-EVIDENCE-WRITER-001`에 blocked**된다. 그 writer가 code-change 세션의 `UsageRecord`에 (a) `PathKind="code-change"`(또는 code-bearing `Phase`) + (b) `Outcome="success"` + `IsTestPass`를 set해야 게이트가 activate된다 (REQ-SEG-012). 이로써 "게이트가 결함을 잡는다"는 주장이 falsifiable해진다. 본 SPEC의 deliverable는 read-side 로직 + schema/분류이며, 현재 스트림 finding 발화는 deliverable가 아니다 (scaffold-honest, SPEC-HARNESS-OUTCOME-CAPTURE-001 / SPEC-HARNESS-REGRESSION-GATE-001 선례).

## §B Data-model fork resolution (the design tension)

**Approach C (Hybrid)** 채택. one-line why: A 단독(record-time omitempty 확장만)은 모든 레거시 JSONL을 false-flag하여 REQ-SEG-010 위반; B 단독(기존 필드 추론만)은 진짜 이진 증거(`IsTestPass`)를 영구 미확보. Hybrid = record-time `UsageRecord` omitempty 확장(`IsTestPass`/`IsTestFail`/`PathKind`) + 부재 시 `Phase`/`AgentType` 추론 fallback + "evidence absent ≠ failed" graceful degradation. constraint #1(새 store 금지)과 양립 — 기존 JSONL schema의 backward-compatible omitempty 확장이며 새 파일/store 아님 (SPEC-HARNESS-OUTCOME-CAPTURE-001 선례). 상세: design.md §0.

## §C HARD constraints 매핑 (7종)

| # | Constraint | REQ |
|---|------------|-----|
| 1 | Reuse LoadBySession, no new ledger store | REQ-SEG-001 |
| 2 | Binary evidence first | REQ-SEG-002 |
| 3 | Coverage-relation deferred (OUT) | §Exclusions E.1.1 |
| 4 | docs-only path-kind bucket exempt | REQ-SEG-003 |
| 5 | C-HRA-008 compliance | REQ-SEG-007 |
| 6 | fail-open (warn-first, dormant exit-2) | REQ-SEG-005, 006 |
| 7 | ≤5s budget | REQ-SEG-008 |

Behavior preservation (critical): REQ-SEG-009 (StopHookActive guard / 90d·30d pruning / AnalyzeSessionAndLog unchanged).

## §D Next step

- amend-1 완료 (5 defects 해소, artifact editing only, Tier M 불변, no code). spec-lint clean 재확인.
- Plan audit gate (Phase 0.5) re-audit: plan-auditor iter-2 (Tier M PASS threshold 0.80) — D1-D5 해소 검증.
- 이후 Implementation Kickoff Approval (사용자 승인) → /moai run SPEC-STOP-EVIDENCE-GATE-001 (cycle_type=tdd).
- run-phase 완료 (M1-M4). 다음 = /moai sync SPEC-STOP-EVIDENCE-GATE-001.

## §E — Phase 0.95 Mode Selection

**Decision: sub-agent** (Mode 5, sequential — coding-heavy single-domain).

| Field | Value |
|-------|-------|
| tier | M |
| scope (files) | 6 (types.go + types_test.go + session_ledger.go + session_ledger_test.go + stop.go + stop_evidence_gate_test.go) |
| domain count | 2 (internal/telemetry, internal/hook) — single coding-domain |
| file language mix | 100% Go |
| concurrency benefit | LOW (coding-heavy, sequential milestone dependency M1→M2→M3) |

Mode evaluation: trivial=no (multi-file semantic change); background=no (Write 작업); agent-team=no
(prereqs 미충족 + <3 domains); parallel=no (coding-heavy per Anthropic coding-task parallelism caveat);
**sub-agent=selected** (default fallback for coding-heavy single-domain work, sequential milestone
dependency); workflow=no (semantic new-code, not mechanical-uniform high-volume).

Justification: coding-heavy single-domain work with strict M1(schema)→M2(reader)→M3(gate wiring)
dependency chain. Mode 5 sequential sub-agent is the Anthropic-recommended default for coding tasks
(coding-task parallelism caveat). No parallel/team benefit — milestones are inherently sequential.

## §E.2 Sync-phase Audit-Ready Signal

sync_commit_sha: 25cddcebb

## §E.3 Run-phase Evidence

| AC | REQ | Status | Verification Command | Actual Output |
|----|-----|--------|----------------------|---------------|
| AC-SEG-001 | REQ-SEG-001 | PASS | `grep -c 'func buildSessionLedger' session_ledger.go` + `grep -cE 'os\.Create\|os\.OpenFile\|os\.WriteFile\|os\.MkdirAll' session_ledger.go` | buildSessionLedger=1, file-write=0; param `[]telemetry.UsageRecord`; LoadBySession reused |
| AC-SEG-002 | REQ-SEG-002 | PASS | `go test -run TestEvaluateEvidence -v` | binary-flip pair: `code_change_success_no_pass_binary_present→finding` + `success_test_pass_observed→nil` (only IsTestPass false→true flips verdict, Outcome=success held constant) |
| AC-SEG-003 | REQ-SEG-003 | PASS | `go test -run 'TestEvaluateEvidence\|TestInferPathKind' -v` | `docs_only_success_no_test_pass→nil` + `phase_sync→docs-only` PASS |
| AC-SEG-004 | REQ-SEG-004 | PASS | `go test -run 'TestEvaluateEvidence\|TestFindingHumanReadable' -v` | `code_change_success_no_pass_binary_present→finding` PASS; HumanReadable names path-kind + success count |
| AC-SEG-005 | REQ-SEG-005 | PASS | `go test -run TestStopEvidenceGate_FailOpen -v` | 4 case (finding present/absent, StopHookActive, empty) 모두 `&HookOutput{}` allow + nil err |
| AC-SEG-006 | REQ-SEG-006 | PASS | `grep -cE 'os\.Stdout' session_ledger.go` + `TestStopEvidenceGate_StdoutContractUnchanged` | stdout-write=0; advisory→stderr/slog; HookOutput Decision/Reason 빈값 |
| AC-SEG-007 | REQ-SEG-007 | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ \| grep -v _test.go \| grep -v comment` | exit 1 (no match = pass) |
| AC-SEG-008 | REQ-SEG-008,010 | PASS | reader heavy-op grep + gate-body awk heavy-op grep + `TestStopEvidenceGate_LegacyFixtureNotFalseFlagged` | reader heavy-op=0; gate body heavy-op=0 + LoadBySession=1 (only data acquisition); legacy record runtime fixture → no finding |
| AC-SEG-009 | REQ-SEG-009 | PASS | `grep -c StopHookActive\|PruneOldFiles\|AnalyzeSessionAndLog\|minToolInvocationsForReflection stop.go` + `TestStopEvidenceGate_StopHookActiveSkipsGate` | StopHookActive=2, PruneOldFiles=2, AnalyzeSessionAndLog=1, minTool=3; stderr-capture proves gate not reached under StopHookActive=true |
| AC-SEG-010 | REQ-SEG-010 | PASS | `go test -run 'TestUsageRecordBackwardCompat\|TestUsageRecordOmitempty' -v` + `grep -c omitempty types.go` | legacy decode → zero-value; zero-value marshal omits 3 keys; omitempty=5 (≥3) |
| AC-SEG-011 | REQ-SEG-011 | PASS | `go test -run TestInferPathKind -v` | explicit_pathkind_wins→code-change, phase_sync→docs-only, phase_run→code-change, ambiguous→unknown; unknown_pathkind→nil_finding |
| (REQ-SEG-012) | REQ-SEG-012 | PASS (doc) | spec.md §A.3/§A.4 dormant-scaffold framing + AC-SEG-008 legacy/dormant fixture | dormant scaffold framing 보존; successor=SPEC-STOP-EVIDENCE-WRITER-001 named |

**New-code coverage**: session_ledger.go 전 함수 100% (runEvidenceGate/buildSessionLedger/inferPathKind/evaluateEvidence/HumanReadable/slogArgs); stop.go Handle 88.9% (게이트 라인 covered, 미커버는 pre-existing 반성학습 분기). Package aggregate: hook 81.8% (baseline 81.5%), telemetry 75.9% (무회귀).

**Cross-platform build**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

**Verification**: `go test ./...` GREEN (전체 회귀 0); `go vet` clean; `golangci-lint run` 0 issues; C-HRA-008 0 매치.

## §E.8 Run-phase commit SHAs

| Milestone | Commit | Subject |
|-----------|--------|---------|
| M1 | `e1dabf661` | UsageRecord 이진 증거/path-kind omitempty 확장 (draft → in-progress) |
| M2 | `ac028f37c` | 세션 원장 리더 + path-kind 분류 + 증거 평가기 |
| M3 | `36a9bc8a5` | stop.go 게이트 삽입 (behavior-preserving additive) |
| M4 | `f600b9d53` | 통합 검증 + progress.md §E run-phase evidence |

post_tool_metrics.go NOT modified (writer wiring = successor SPEC scope). 변경 파일 7개 전부 SPEC scope 내.

## §E.5 Mx-phase Audit-Ready Signal

mx_commit_sha: 59561c92d

4-phase 완결: plan(`3eda7f5de`) → run M1-M4(`e1dabf661`..`f600b9d53`) → sync(`25cddcebb`) + doc-fix(`c58c947be`) → Mx close. status: implemented → completed. plan-auditor iter-2 PASS 0.91 · sync-auditor PASS (Func96/Sec98 MUST-PASS) · 11/11 AC PASS · 신규코드 100% cov · C-HRA-008 clean. knowingly-dormant scaffold — gate VALUE blocked on 후속 `SPEC-STOP-EVIDENCE-WRITER-001`.
