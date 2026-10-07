# SPEC-RECEIPT-REUSE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

`plan_status: audit-ready`
`plan_complete_at: 2026-10-07T13:01+09:00`
근거: plan-audit iter-3 PASS 0.96 (must_pass_failed 0, blocking_count 0, scope:reread) — `.moai/reports/t1562/plan-audit-iter3.md`
저작 트리: `f97edcc55` (브랜치 `WT-receipt-reuse`, 워크트리 `.moai/worktrees/t1562`), Tier M 산출물 5종(spec/plan/acceptance/progress/decision-index).

## §E.2 Run-phase Evidence

### M1 — 행위적 RED 관측 (2026-10-07, 트리 `62574ab3a`, 브랜치 `WT-receipt-reuse`)

**명령**:

```
go test -count=1 -run '^(TestSubagentStop_SequentialAuditorReceiptReuseIsRefused|TestSubagentStop_ReceiptBoundaryAmbiguitySemantics)$' -v ./internal/hook
```

**exit code**: `1`

**원문 출력** (실패 단정 행 전문):

```
    audit_receipt_guard_test.go:733: decision = "", want block — a PASS resting on a predecessor-era receipt must be refused (reason "")
    audit_receipt_guard_test.go:733: decision = "", want block — a PASS resting on a predecessor-era receipt must be refused (reason "")
    audit_receipt_guard_test.go:733: decision = "", want block — a PASS resting on a predecessor-era receipt must be refused (reason "")
    audit_receipt_guard_test.go:733: decision = "", want block — a PASS resting on a predecessor-era receipt must be refused (reason "")
--- FAIL: TestSubagentStop_SequentialAuditorReceiptReuseIsRefused (0.63s)
    --- FAIL: TestSubagentStop_SequentialAuditorReceiptReuseIsRefused/accepted-pass-end (0.22s)
    --- FAIL: TestSubagentStop_SequentialAuditorReceiptReuseIsRefused/reentry-refusal-end (0.13s)
    --- FAIL: TestSubagentStop_SequentialAuditorReceiptReuseIsRefused/fail-end (0.10s)
    --- FAIL: TestSubagentStop_SequentialAuditorReceiptReuseIsRefused/no-verdict-reentry-end (0.17s)
    audit_receipt_guard_test.go:802: output = &{Continue:<nil> StopReason: SystemMessage: SuppressOutput:false Decision: Reason: HookSpecificOutput:<nil> UpdatedInput: Retry:false ExitCode:0 WorktreePath: Data:[]}, want a block naming "receipt created before the previous auditor instance of this session ended"
--- FAIL: TestSubagentStop_ReceiptBoundaryAmbiguitySemantics (0.27s)
    --- FAIL: TestSubagentStop_ReceiptBoundaryAmbiguitySemantics/single-live-end-advances-boundary (0.11s)
    --- PASS: TestSubagentStop_ReceiptBoundaryAmbiguitySemantics/ambiguous-end-freezes-boundary (0.15s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/hook	1.680s
```

**RED 적색 이유 (올바른 이유)**: 네 종료 형태 팔 전부와 AC-RR-009 시퀀스 (i)에서 재사용 PASS가 **수락**(`decision = ""`)됐다 — 결함 본체 그 자체. `reason = ""`는 수락(거부 원인 부재)의 표기다. 모호 종료 동결 팔(`ambiguous-end-freezes-boundary`)은 보존 팔이라 M1에서 PASS — 수리 뒤에도 문장 불변으로 유지된다. swept-count 상관: `-list` 0일치(저작 전, LEDGER-RRR-A/E-H) → 5일치(M1 착수 시 재측정).

**하네스 정정 기록 (투명성)**: 첫 RED 시도에서 stop 입력이 `SessionID`를 실지 않아 "start marker missing"으로 적었던 원문은 **틀린 이유의 적색**(하네스 결함)이었고, `bgStopInput` 헬퍼(세션 실은 배경 stop 페이로드)로 정정한 뒤의 위 원문이 올바른 이유의 적색이다. 테스트 문장(단정 내용)은 변하지 않았다.

**기저 측정 (B5 — 사전 존재 vs 신규 분리, 트리 `62574ab3a`)**: `go test -count=1 -timeout 30m ./internal/hook/ ./internal/auditreceipt/` → `internal/auditreceipt` **ok** (4.670s), `internal/hook` **FAIL 796.944s, 사전 존재 실패 4건** — `TestAstgrepCorpusRunDoesNotSkip`·`TestStaleRunNoticeLegacyLeaderSpelling`·`TestStaleRunNoticeLegacySessionRecord`·`TestStaleRunNoticeFactoryLegacyLabel` — 감사 영수증 표면과 무관. 기저 601초 시도는 기본 10분 타임아웃 도달(패닉 스택)이었고 `-timeout 30m` 재측정으로 위 4건이 전부다. 이 SPEC이 추가하는 신규 적색은 M1의 재현 테스트뿐이다.

## §E.3 Run-phase Audit-Ready Signal

```
run_complete_at: 2026-10-07T16:05+09:00
run_status: complete
run_commit_sha: pending-backfill-run
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 4
l44_pre_commit_fetch: not-performed (isolated card worktree .moai/worktrees/t1562 — 커밋은 카드 트리에서, push 없음)
l44_post_push_fetch: n/a (레인 push 금지 — 리더 일괄 착지)
new_warnings_or_lints_introduced: 0
cross_platform_build: darwin/arm64 ok · windows/amd64 ok (go build ./... 전체)
total_run_phase_files: 6
m1_to_mN_commit_strategy: M1 RED · M2 수리 · M2 리뷰 수리 · M3 증거 — 마일스톤별 별도 커밋
```

**측정 귀속 (이 실행, 이 트리 — HEAD `44d79fb86` 브랜치 `WT-receipt-reuse`)**:

- **AC 전수**: 9/9 PASS — AC-RR-001/002 `TestSubagentStop_SequentialAuditorReceiptReuseIsRefused` (네 종료 형태 팔 + 거부 기록·스폰 거부 3종·해제), AC-RR-003/005/008 보존 앵커 3건 문장 불변 GREEN, AC-RR-004/006/007 확장 테스트 swept=1 각각 exit 0, AC-RR-009 `TestSubagentStop_ReceiptBoundaryAmbiguitySemantics` (단일-생존 전진 + 모호 동결 대조). swept-count 우선 `-list` 5(신규)+6(앵커) 상관 후 `-run`.
- **영향 패키지 전체**: `go test -count=1 -timeout 30m ./internal/hook/ ./internal/auditreceipt/` → `internal/auditreceipt` ok (3.006s), `internal/hook` 기존 실패 4건만 (`TestAstgrepCorpusRunDoesNotSkip`·`TestStaleRunNoticeLegacy{LeaderSpelling,SessionRecord,FactoryLabel}`) — 기저(62574ab3a, 796.944s)와 동일 집합, **신규 실패 0**.
- **경합**: 감사 영수증 패밀리 `-race` ok; 원장 동시성 테스트(32 start/32 end) `-race` ok — 잠금 우회 변이체에서 `1 starts / 6 ends` 소실 재현 후 복원 판정.
- **커버리지**: `internal/auditreceipt` 87.1% (패키지 목표 85% 충족; 원장 신규 함수 ReadInstanceLedger 100%·RecordInstanceStart/End 100%·lockLedger 72.2%·updateInstanceLedger 81.8%·CheckCitedReceiptsSince 80.8%) / `internal/hook` guard 함수(패밀리 스코프): checkAuditorStop 95.3%·instanceEndBoundary 100%·recordAuditorEnd 75%·recordAuditorStart 87.5%. 훅 패키지 전체 커버리지 기저 수치는 미측정(Gap — CI 몫).
- **lint**: golangci-lint 양 패키지 0 issues (총계 0 = 신규 0). `go vet` 양 패키지 통과.
- **E4 하위경계**: 신규 추가 0 — 사전 존재 1건(`internal/hook/pre_tool.go:855` 질문 채널 관측 분기, 카드 t1530 이후 미변경, 호출 아닌 관측)만 존재.
- **보존 목록 4항목**: wsr 테스트 파일 diff 0 · kind 기계/StoreRoot/TreeRootFromCWD/sanitizeKey 미편집 · FAIL 경로/재진입 경로/거부 갱신 규칙 기존 의미 불볇(원장 카운트만 가법) · 앵커 테스트 6건 문장 불변 GREEN. (§D.1 "FAIL 소비 경로" 보존은 기존 의미 보존으로 판독 — AC-RR-001 팔 (c)가 FAIL 종료의 종료 기록을 요구하므로 가법적 종료 기록은 계획 기제의 일부다.)
- **리뷰 수리 접수 (코디네이터 mid-flight 리뷰 2건)**: P1 원장 무잠금 RMW → 포터블 lockfile(O_EXCL, 250ms 유계 대기, 5s stale break) 직렬화 — 기존 직렬화 관용 부재 검증 후 신규 도입; P2 원장 읽기 실패 zero-vision → `instanceEndBoundary` 오류 전파 + `CauseInstanceLedgerUnreadable` 폐쇄 거부(스폰 게이트의 읽을 수 없는 거부 기록 판독과 동일 원칙). 종료-기록 쓰기 실패는 기존 관용(로그 전용) 유지 — atomic rename이 이전 유효 원장을 남기므로 무-정지 소멸 잔여(acceptance §E)와 동일 급.

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection + plan→run Kickoff Record

**Mode Selection** (orchestrator lane-10, 2026-10-07):

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | semantic multi-file repair, not trivial |
| **serial** | **YES** | coding-heavy single-domain defect repair — Anthropic coding-task parallelism caveat |
| fanout | no | not multi-domain research |
| sweep | no | not ≥30-file mechanical transform (also gate-blocked pre-run) |

Decision: **serial** (Tier M, ~2-4 files in `internal/hook` + `internal/auditreceipt`, 1 domain, Go-only, concurrency benefit low, agent-team not requested).

**plan→run Kickoff — autonomous transition evidence** (default form, auto-semantics §9.1):

1. Independent plan-audit verdict **PASS 0.96** — `.moai/reports/t1562/plan-audit-iter3.md` (iteration 3, scope:reread; must_pass_failed 0, blocking_count 0; receipts rcpt-77778ac0732be7bea6f7914d).
2. Plan phase **audit-ready** — progress.md §E.1 `plan_status: audit-ready`, `plan_complete_at: 2026-10-07T13:01+09:00`.
3. Plan-artifact hash unchanged since the verdict — verdict measured at tree `7d087d121`; `git status` clean since; progress.md §E.1/§F edits are outside the ComputeHash subject set (acceptance/decision-index/design/plan/research/spec/tasks).
4. No blocker open (all delegate reports consumed; no missing inputs).
5. Phase-1 re-execution skip-eligible per SPEC-AUDIT-SNAPSHOT-001 A1+A2: verdict PASS + score 0.96 ≥ Tier M 0.80 + hash unchanged.

Decision record: decided_by=lane-10 (factory lane, run tmhxo0) evidence_refs=.moai/reports/t1562/plan-audit-iter3.md#PASS-0.96 + .moai/specs/SPEC-RECEIPT-REUSE-001/progress.md §E.1 ladder_path=auto-semantics §9.1 autonomous default.
