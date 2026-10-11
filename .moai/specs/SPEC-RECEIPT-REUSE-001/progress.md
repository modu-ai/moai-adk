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
- **post-sync 수리 r4 접수 (게이트 라운드 25 — 읽을 수 없는 pending의 폐쇄)**: ⑨ r2에서 문서화했던 잔여가 입증된 실패 경로로 재현 — 읽을 수 없는(corrupt JSON) pending 종료 기록이 발견에서 무음 스킵되고 이전 경계가 그대로 적용돼 후속 감사자가 선행자의 영수증을 수락(RED 원문: `output = &{... Decision: Reason: ...}, want a block naming an unreadable end record — not a silent skip to the old boundary`; 원장측 `the ledger update proceeded over an unreadable pending end`) → **fail-closed 일관**: `ErrPendingEndUnreadable` + `scanEndPendings` 분리(판독 가능 목록/판독 불가 신호) — 원장 연산은 부분 pending 집합을 적용하지 않고 제동하고, 승인 경로(`instanceEndBoundary`)는 `HasUnreadablePendingEnd` 신호로 전용 원인 `CauseInstancePendingUnreadable`("instance end record unreadable") 거부. 해제는 운영자의 기록 수리(corrupt ledger·거부 기록과 동일 급). 핀 2건(가드 거울 + 원장 제동) RED 관측 후 녹색. 이로써 문서화된 잔여가 닫힌다.
- **post-sync 수리 r5 접수 (게이트 라운드 25 — 3면)**: ⑩ chmod 저장-실패 주입이 Windows에서 성립하지 않음(os.Chmod는 디렉터리 내 파일 생성을 막지 않음 — r4 보고서 사전 경고) → **저장 봉합사** `ledgerSave` var(`ledgerLockWait`·`Now`와 같은 패턴)로 플랫폼 무관 주입, 권한 게임 제거; ⑪ 역순 도착 종료가 마지막 종료를 소실 — 나중(3s) 종료가 모호로 기각되고 이전(2s)이 단일-생존 봉인해 경계가 진짜 마지막 종료(3s)보다 앞서 그 사이 영수증 재사용 수락(RED 원문: `EndedAt = ...12:00:02, want ...12:00:03`) → **귀속 무관 워터마크** `LastEndedAt`(모든 기록된 종료의 최댓값 — 모호로 기각된 종료도 발생한 것)를 원장에 두고 단일-생존 봉인 시 경계가 워터마크까지 따라 올라감(모호 동결은 도착 시점 규칙 그대로 — AC-RR-009(ii) 핀 유지); ⑫ 시작 원장 기록은 성공했으나 마커 저장이 실패하면 종료 시점 마커 조회가 빈 키를 돌려 종료 집계가 스킵 — 유령 생존자가 경계를 영구 동결(RED 원문: `ledger = 1 starts / 0 ends, want 1 / 1`) → **종료 집계 키를 마커 조회와 분리**(`endAggregationKey`: 발견 키 없으면 시작 원장이 쓴 파생 키로 집계). 핀 3건 전부 RED 관측 후 녹색(-race 포함). r4 전체 재측정에서 관측된 중부하 플래키 3건(`TestFactory*`·`TestSessionStart_DeferredScan`)은 동일 트리 단독 재실행 전부 통과 확인 — 부하 민감형, 본 SPEC 표면과 무관.
- **post-sync 수리 r5 보충 접수 (게이트 라운드 31-32 — 종료 키·승인 표면 3면)**: ⑬ 전경 감사자의 마커 저장 실패 시 그 FAIL 종료가 배경 원장으로 집계(시작 0·종료 1) — 뒤이은 두 겹침 배경 감사자가 그 유령 종료를 단일-생존으로 오독해 생존 감사자의 자기 영수증이 거부됨(RED 원문: `decision = "block", want none — the live auditor's own receipt must not be refused by a phantom foreground end`) → **원장이 시작을 세지 않은 종료는 집계하지 않음**(`applyEnd` 조기 반환: `Ends >= Starts` — 개수·워터마크·봉인 전부 스킵; r2 시절 고아-종료 봉인 진행 핀은 이 규칙으로 전용 — 고아 종료는 외래일 수 있음이 알려졌다); ⑭ 읽을 수 있으나 미적용인 pending 종료가 승인을 안 막음 — 원장 경계는 진짜 마지막 종료보다 낡은 채 승인이 선행 영수증을 통과(RED 원문: `output = &{... Decision: Reason: ...}, want a block naming "instance end record unresolved"`) → **`PendingEndHold` 센티널**(판독 불가 → `ErrPendingEndUnreadable`, 미적용 → `ErrPendingEndUnresolved` — applied_ends에 없는 id) + 전용 원인 `CauseInstancePendingUnresolved`; 미적용 홀드는 다음 원장 연산이 적용하므로 자기 해소. ⑮ r4 동일 원인의 확장이 아닌 독립 원인으로 분리한 근거: 판독 불가는 운영자 수리가 필요하고 미적용은 자기 해소 — 해제 경로가 다른 조건은 원인이 구별돼야 한다(REQ-RR-003 어휘 규율). 핀 3건(T15 교정 시계·T16 홀드·고아-종료 스킵) RED 관측 후 녹색. 최종 트리: 양 패키지 `-race` ok, 커버리지 85.9%(≥85%), 빌드(native+windows)·lint 0·vet 통과.
- **post-sync 수리 r6 접수 (게이트 라운드 33 — 재생 판단 2면, 분할 착지 A)**: ⑯ pending 재생이 재생 시점 카운트로 판정을 재유도 — A의 단일-생존 종료 저장 실패 → pending 생성 → B의 시작 기록 → 재생이 현재 카운트(B 생존, 2 outstanding)로 모호 재판정 → 마크 소비 후 경계 미봉인 → B가 감사 없이 A의 영수증 통과(RED 원문: `EndedAt = 0001-01-01 ... want the end-time-sealed ...12:00:00.5 — the replay must keep the end-time judgment, not re-derive it as ambiguous`) → **pending이 종료 시점 판단을 운반**(`endPending.SingleLive` — `RecordInstanceEnd`가 판단을 알 때 저장하고 재생이 그대로 적용); ⑰ 잠금 미획득(판단 이전 실패) 시 종료 시점 상태는 **lock-free 원장 읽기**(atomic rename — 읽기엔 잠금 불요)로 복원해 판단을 복원하고, 읽기 자체가 실패하면 판단 없는 pending(재생 시 외래가 아니면 카운트만) + `PendingEndHold`가 미적용 동안 승인을 보류 — 확립된 fail-closed 패턴. 핀 T17 RED(`EndedAt = 0001-01-01, want ...12:00:00.5` + A-시절 영수증 수락) 관측 후 녹색; T5(드롭 엔드 미러)가 잠금-타임아웃 pending의 count-only 재생으로 적색이 된 것을 같은 수정으로 복원 — 두 핀의 시맨틱이 상충하지 않음을 실증(판단 보유 vs 판단 재유도의 구분). 인터퍼런스 기록: 라운드 33 소비 중 코디네이터의 `git merge origin/main` + `--abort`가 본편집 창과 겹쳐 트리를 일시 불일치시켰으나 — 상태 전수 판독 결과 r6 편집은 전수 생존(endJudgment 11참조·PendingEndHold·T15/16/17·4-param 마킹 전부 확인), vet+양 패키지 `-race` 재판정으로 실증. 남은 조각(다음 커밋): **생존자 잔존 상태에서의 전경 종료 귀속**(⑬의 `Ends >= Starts` 가드가 남아 있는 생존자가 있으면 미작동 — 라운드 33 P2 변형) — 전경 시작 실패 흔적 마크 + 종료 키 귀속 규칙으로 별도 커밋 착지 예정.
- **post-sync 수리 r6 보충 착지 (게이트 라운드 33 P2 — 생존자 잔존 변형)**: ⑬의 방어(`Ends >= Starts`)는 배경 감사자 둘이 이미 살아 있을 때 도착하는 전경 종료를 못 막는다(방 안 남는 자리가 있음) — 유령 종료가 계수되고 첫 배경 종료가 생존자가 살아 있는 채 단일-생존 봉인해 그 자기 영수증이 거부됨(T15b RED: `decision = "block", want none ... receipt created before the previous auditor instance of this session ended`) → **전경 시작 실패 흔적**: `recordAuditorStart`의 비파생 분기가 마커 저장 실패 시 `MarkForegroundMarkerFailure`로 `starts/<id>.json.foreground-failed`를 남기고(멱등·best-effort), 종료 키 귀속이 대체 히트(자기 id 미스 → 형제의 파생 앵커 히트) 위에서 흔적을 검사해 있으면 **아무것도 기여하지 않음**(개수·워터마크·봉인·pending 전부 없음). 판별식의 건전성: 자기 id 히트 ⟹ 전경 자기 마커(원장 무관), 파생 히트+흔적 없음 ⟹ t1544 배경(원장 기록 있음), 파생 히트+흔적 있음 ⟹ 전경 실패(스킵). 이중 실패 잔여(흔적 쓰기도 실패)는 방 안 남는 자리 시험으로 문서화. 핀 T15b RED 관측 후 녹색. 최종 트리: 양 패키지 `-race` ok, 커버리지 85.8%(≥85%), 빌드(native+windows)·lint 0·vet 통과, 전체 재측정 신규 실패 0.
- **post-sync 수리 r7 접수 (게이트 라운드 34 — 3면, 최종 in-lane)**: ⑱ 잠금을 잃은 두 종료가 **판단 없는** pending을 남겨 회수 후에도 경계가 영(0) — 선행 영수증이 무감사 통과(RED 원문: `EndedAt = 0001-01-01 ... want ...12:00:03` + `a predecessor-era receipt was accepted after the fully-accounted recovery sealed the era`) → **회복 시점 era-closure**: 개별 종료의 단일-생존 판단은 형제 종료의 운명을 모르지만, 재생이 모든 마크를 적용한 뒤 `Ends == Starts`이면 생존자가 없어 동결의 근거가 소멸 — 경계를 워터마크(완전 계수된 era의 마지막 종료)까지 봉인. 평가 시점의 근거: 개별 판단은 pending-time, era 폐쇄 판단은 recovery-time — 후자만 결합 집합을 볼 수 있다. ⑲ 판독 불가 원장에서의 종료가 아예 저장되지 않음(counted=false → pending 없음 — pending 디렉터리는 쓰기 가능함에도; RED 원문: `pending marks = [], want exactly 1`) → 판독 실패 시 **판단 없는 pending 저장**(복원 뒤 카운트 회복 + 미적용 동안 PendingEndHold 승인 보류 — 확립된 fail-closed); 외래 확정(가독 원장의 Ends >= Starts)일 때만 기여 없음. ⑳ 늦게 착지한 단일-생존 pending의 재생이 **전역 워터마크(C 세대의 종료)**로 승천해 생존 B의 겹침-시절 영수증 거부(RED 원문: `EndedAt = ...12:00:04, want ...12:00:01 — the recovery must not lift the boundary to a later generation's watermark` + `B's overlap-era receipt was refused`) → 재생 봉인을 **pending 자신의 종료 시각으로 한정**(세대 경계) — era-closure(전체 계수 시)만 워터마크에 도달. **설계 판정: 세 결함 모두 경계 설계의 재작업이 아니라 회복 경로의 보완이다** — 워터마크·세대·동결 개념은 기존 설계 안에서 조합됐고 새 개념 도입은 없음. 핀 3건(T18·T19·T20) RED 관측 후 녹색(-race). 부수: T15b의 공허 통과 제거 — fg stop에 SessionID 누락으로 fg 종료가 아예 집계되지 않아(커밋 B의 흔적 기제가 실질 시험되지 않음) 실제 페이로드 형태의 세션을 싣도록 교정, 흔적 거부 경로가 실질 고정됨. 최종 트리: 양 패키지 `-race` ok, 커버리지 86.3%(≥85%), 빌드(native+windows)·lint 0·vet 통과.
- **post-sync 수리 r8 접수 (게이트 라운드 37 — marker-less 조기 반환의 흔적 검사 누락, 코너 1건)**: ㉑ `endAggregationKey`의 marker-less 조기 반환(`foundKey == ""`)이 전경-실패 흔적 검사보다 먼저 발화 — 배경 시작이 집계됐고(자기 마커 저장 실패) 전경 마커 저장도 실패한 시나리오에서 전경 FAIL 종료가 생존 배경 인스턴스에 귀속: 계수 + 경계 SET(RED 원문: `ledger = 1 ends, EndedAt ...12:00:01 — the foreground end was attributed to the live background instance (want 0 ends, no boundary)`) → **흔적 검사를 폴백 결정 앞으로** 이동(귀속 표의 흔적 팔이 모든 경로를 방어) — 흔적이 있으면 어트리뷰션 전부 기각(개수·경계·pending 없음). 정직한 2차 관측: 시나리오상 배경 인스턴스 자신의 마커도 파손돼 있어 그 종료는 "start marker missing"으로 거부되는 것이 옳은 실패-닫기(r5#3 계열의 문서화 한계 — 자기 마커 상실 인스턴스의 증명 불가)이며, 오귀속 경계 거부는 소실. 핀 T21 RED 관측 후 녹색. 최종 트리: 양 패키지 `-race` ok, 커버리지 86.3%(≥85%), 빌드(native+windows)·lint 0·vet 통과. r4 전체 재측정은 중부하(1400s — 평소 2배+)에서 실행돼 기존 기저 외 3건(`TestFactory*` 2건·`TestSessionStart_DeferredScanJoinsWithinBound`)이 적었으나 동일 트리 단독 재실행 전부 통과(1.97s/0.92s/1.21s — 부하 민감형 분리), 지속 실패는 기저 플래키 `TestAstgrepCorpusRunDoesNotSkip`(120s 내부 타임아웃, 최초 기저 `62574ab3a`에 존재)뿐 — 신규 지속 실패 0.

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
