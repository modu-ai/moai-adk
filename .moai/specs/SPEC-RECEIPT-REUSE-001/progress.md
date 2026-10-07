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
- **경합**: 감사 영수증 패밀리 `-race` ok; 원장 동시성 테스트(32 start/32 end) `-race` ok — 잠금 우회 변이체에서 `1 starts / 6 ends` 소실 재현 후 복원 판정. post-sync 수리 뒤 재측정: 패밀리+원장 전체 `-race` ok (`internal/auditreceipt` 전체 패키지 `-race` 포함), 잠금 소유권 테스트(선행 보유자의 release가 계승자의 잠금을 지우지 않음)·보류 종료 리플레이 테스트 추가 GREEN.
- **커버리지**: `internal/auditreceipt` **86.0%** (패키지 목표 85% 충족; post-sync 수리 함수 포함 최종) / `internal/hook` guard 함수(패밀리 스코프): checkAuditorStop 95.3%·instanceEndBoundary 100%·recordAuditorEnd 75%·recordAuditorStart 87.5%. 훅 패키지 전체 커버리지 기저 수치는 미측정(Gap — CI 몫).
- **lint**: golangci-lint 양 패키지 0 issues (총계 0 = 신규 0). `go vet` 양 패키지 통과.
- **E4 하위경계**: 신규 추가 0 — 사전 존재 1건(`internal/hook/pre_tool.go:855` 질문 채널 관측 분기, 카드 t1530 이후 미변경, 호출 아닌 관측)만 존재.
- **보존 목록 4항목**: wsr 테스트 파일 diff 0 · kind 기계/StoreRoot/TreeRootFromCWD/sanitizeKey 미편집 · FAIL 경로/재진입 경로/거부 갱신 규칙 기존 의미 불볇(원장 카운트만 가법) · 앵커 테스트 6건 문장 불변 GREEN. (§D.1 "FAIL 소비 경로" 보존은 기존 의미 보존으로 판독 — AC-RR-001 팔 (c)가 FAIL 종료의 종료 기록을 요구하므로 가법적 종료 기록은 계획 기제의 일부다.)
- **리뷰 수리 접수 (코디네이터 mid-flight 리뷰 2건)**: P1 원장 무잠금 RMW → 포터블 lockfile(O_EXCL, 250ms 유계 대기, 5s stale break) 직렬화 — 기존 직렬화 관용 부재 검증 후 신규 도입; P2 원장 읽기 실패 zero-vision → `instanceEndBoundary` 오류 전파 + `CauseInstanceLedgerUnreadable` 폐쇄 거부(스폰 게이트의 읽을 수 없는 거부 기록 판독과 동일 원칙). 종료-기록 쓰기 실패는 기존 관용(로그 전용) 유지 — atomic rename이 이전 유효 원장을 남기므로 무-정지 소멸 잔여(acceptance §E)와 동일 급.
- **post-sync 수리 접수 (코디네이터 게이트 라운드 — 잠금 소유권 + 드롭 기록 3면)**: ① 차단된(오래된) 잠금의 release가 계승자의 잠금을 지우는 소유권 결함 → **토큰 인식 release**(잠금 파일에 보유자 토큰 기록, release는 자기 토큰이 살아 있을 때만 삭제; flock 대신 토큰 선택 — 16 플랫폼 단일 코드 경로 + 소유권 의미의 결정론적 단위 시험, 브레이크 코너 잔여는 문서화); ② 드롭된 START 기록이 불완전 카운트로 경계를 전진 → `MarkInstanceStartUncertain` 내구 마크 + END 경로 동결(리플레이보다 동결이 우선 — 카운트는 영구 불신뢰); ③ 드롭된 END 기록으로 경계가 봉인되지 않아 후속 감사자가 재사용 수락 → `MarkInstanceEndPending` 내구 마크 + 다음 원장 연산이 잠금 아래 리플레이(드롭 END는 회수 가능 — 종료 시각 자체는 알려져 있음). RED 우선 핀 5건(T1 소유권·T2 불확실 동결·T4 드롭 스타트 가드 시나리오·T5 드롭 엔드 가드 거울·T6 리플레이 단위) 전부 적색 관측 후 녹색. 신규 마킹 함수가 신규 저장소 첫 이벤트에서 디렉터리 부재로 실패하는 결함을 시험이 포착 → MkdirAll 수리. 최종 트리 전체 패키지 재측정: `internal/auditreceipt` ok(-race, 86.0%), `internal/hook` 기존 실패만(`TestStaleRunNoticeLegacy*` 3건 — 플래키 astgrep 이번 라운드 통과), 신규 실패 0.
- **post-sync 수리 r2 접수 (게이트 라운드 20 — pending 큐 2면)**: ④ 단일 `.end-pending` 파일에 두 실패 작성자가 서로를 덮어써 한 종료가 말없이 소실 → **인스턴스별 pending 파일**(엔트리 id를 파일명에 — `end-pending-<id>`), 리플레이는 잠금 아래 전부 적용(종료 시각 순 정렬 — 단일-생존 판정이 실제 순서를 읽도록), `applied_ends` 원장이 이미 반영된 id를 건너뜀; ⑤ pending이 원장 저장보다 먼저 삭제돼 저장 실패 시 양쪽에서 소실 → **적용 → 저장 → 성공 후 소비** 순서로 재정렬(저장 실패 시 pending 파일 생존, 다음 연산이 재적용; 저장 성공 후 제거 소실 크래시 창은 엔트리 id 멱등으로 이중 계수 차단). RED 우선 핀 3건(동시 실패 작성자 전수 회수 · chmod 저장 실패 후 생존·재회수 · id 멱등) 전부 적색 관측 후 녹색. 최종 트리: 양 패키지 `-race` ok, 커버리지 86.0% 유지, 빌드(native+windows)·lint 0·vet 통과.
- **post-sync 수리 r3 접수 (게이트 라운드 21 — 경계 단조성 + 마커 원자성)**: ⑥ 종료-경계가 역행 — 잠금 획득 순서가 이벤트 순서와 다르면 나중(4s)이 먼저 적용되고 이전(3s)이 `EndedAt`를 끌어내려 3.5s 주조 영수증의 재사용이 수락됨 → `applyEnd`가 **워터마크 단조 전진만** 허용(현재 경계보다 늦은 종료만 반영; 정렬된 리플레이 내부와 별도 연산 간 모두). RED 원문: `EndedAt = ...12:00:03, want ...12:00:04` + `a receipt minted after the true last end but before the boundary was accepted`. ⑦ 첫 시작 표식 기록의 check-then-write 경합 — 두 동시 시작이 모두 부재를 보고 나중 쓰기가 최초 앵커를 덮어 쓰면 두 시작 사이에 주조된 정당 영수증이 "before start"로 거부됨(AC-RR-003 침해 경로) → `EnsureStartMarker`: 존재 확인+기록을 키의 원장 잠금 아래 하나의 임계영역으로(기존 앵커 불변). RED는 사전 봉합사로 관측(동시 `WriteStartMarker` 8회 — `REGRESSED anchor: StartedAt = ...340808, want ...280808`, 재현 파일은 폐기). 핀: 단조성(T10, 컴파일 가능 RED)·동시 기존 앵커 생존·잠금 직렬화(잠금 보유 중 ensure는 포기)·순차 keep — 전부 `-race` GREEN.
- **post-sync 수리 r3 보충 접수 (게이트 라운드 22 — pending 탐색의 glob 패턴 해석)**: ⑧ pending 파일 발견이 `filepath.Glob`으로 프로젝트 경로를 패턴 해석 — `project[1]` 같은 트리 경로에서 문자 클래스로 읽혀 아무것도 match하지 않아 드롭된 종료가 영원히 미회수(RED 원문: `Ends = 1, want 2 — the pending end was not discovered under a glob-metacharacter path`) → **디렉터리 읽기 + 리터럴 파일명 접두사 비교**(`os.ReadDir` + `HasPrefix`, 경로 패턴 해석 없음 — 키는 sanitize라 메타문자 없음, `.json.` 경계로 타 키 접두사 혼동 없음). r3 과잉 발견: 32-경합 원장 테스트가 전체 스위트 병렬 부하에서 250ms 잠금 예산 소진으로 플래키 — 테스트 예산 확대 var(생산 기본 250ms 불변, 타임아웃 핀 `TestLedgerLockTimesOutWhenHeld` 별도 유지). 최종 트리: 양 패키지 `-race` ok, 커버리지 85.4%(≥85%), 빌드(native+windows)·lint 0·vet 통과.
- **post-sync 수리 r4 접수 (게이트 라운드 25 — 읽을 수 없는 pending의 폐쇄)**: ⑨ r2에서 문서화했던 잔여가 입증된 실패 경로로 재현 — 읽을 수 없는(corrupt JSON) pending 종료 기록이 발견에서 무음 스킵되고 이전 경계가 그대로 적용돼 후속 감사자가 선행자의 영수증을 수락(RED 원문: `output = &{... Decision: Reason: ...}, want a block naming an unreadable end record — not a silent skip to the old boundary`; 원장측 `the ledger update proceeded over an unreadable pending end`) → **fail-closed 일관**: `ErrPendingEndUnreadable` + `scanEndPendings` 분리(판독 가능 목록/판독 불가 신호) — 원장 연산은 부분 pending 집합을 적용하지 않고 제동하고, 승인 경로(`instanceEndBoundary`)는 `HasUnreadablePendingEnd` 신호로 전용 원인 `CauseInstancePendingUnreadable`("instance end record unreadable") 거부. 해제는 운영자의 기록 수리(corrupt ledger·거부 기록과 동일 급). 핀 2건(가드 거울 + 원장 제동) RED 관측 후 녹색. 이로써 문서화된 잔여가 닫힌다. r4 전체 재측정은 중부하(1400s — 평소 2배+)에서 실행돼 기존 기저 외 3건(`TestFactory*` 2건·`TestSessionStart_DeferredScanJoinsWithinBound`)이 적었으나 동일 트리 단독 재실행 전부 통과(1.97s/0.92s/1.21s — 부하 민감형 분리), 지속 실패는 기저 플래키 `TestAstgrepCorpusRunDoesNotSkip`(120s 내부 타임아웃, 최초 기저 `62574ab3a`에 존재)뿐 — 신규 지속 실패 0.

## §E.4 Sync-phase Audit-Ready Signal

```
sync_complete_at: 2026-10-07T14:58+09:00
sync_commit_sha: 127d87ac4
sync_status: complete
b12_self_test_a: pre_emission_grep=0 — grep -c 'SPEC-RECEIPT-REUSE-001' CHANGELOG.md → 0 (exit 1, 중복 엔트리 없음, emission 허용)
b12_self_test_b: ac_count=9 — acceptance.md (tier M AC 원천) 카운터 live=9 excluded=0 ambiguous=0; CHANGELOG 엔트리의 9 기재와 일치
b12_self_test_c: file_paths 4/4 존재 — internal/hook/audit_receipt_guard.go, internal/auditreceipt/store.go, internal/hook/audit_receipt_guard_test.go, internal/auditreceipt/store_test.go (ls 실측)
changelog_entry_position: "[Unreleased] > ### Fixed 첫 항목"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (implemented 중간점은 이 sync 커밋에 병합 — 3-phase close)"
  plan_md: "updated 리프레시 (frontmatter 한정; status 필드 없음 — stateless)"
  acceptance_md: "frontmatter 블록 부재 — 리프레시 대상 없음 (본문 미편집)"
mx_tag_report: "added=0 removed=0 updated=0 — 기존 파일 수준 ANCHOR/REASON(/SPEC) 태그는 SPEC-CODEX-AUDIT-GATE-AXES-001 유산으로 여전히 정확; 신규 export 함수(ReadInstanceLedger·RecordInstanceStart·RecordInstanceEnd·CheckCitedReceiptsSince)는 godoc 보유·fan_in<3·위험 패턴 없음으로 프로토콜 요구 태그 없음 (moai mx scan --dry --path 실측)"
```

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
