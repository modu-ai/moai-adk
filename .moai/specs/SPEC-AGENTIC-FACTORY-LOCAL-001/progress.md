# SPEC-AGENTIC-FACTORY-LOCAL-001 — progress

status: draft

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: drafted (not audited) — 0.1.0 작성(2026-10-30, HEAD `d87913b6a3`, 코드 사실 기준 `37b67f6d39`). 감사는 coordinator가 돌린다.
- Tier L 선택: REQ 21개가 Tier M 상한 16을 넘음(spec HISTORY 0.1.0).
- 측정은 상위 progress.md `revision_0_7_0_local_ledger` 노트에 있다.
- 0.2.0 (2026-10-30, HEAD `c445ce9559` 위 작업 트리): revised, unaudited — plan-audit FAIL 0.62 수리. 감사는 coordinator가 돌린다. PASS라고 적지 않는다. 대응표는 §G.
- 0.3.0 (2026-10-30, HEAD `0c1cb1853f` 위 작업 트리): revised, unaudited — OD-36~42를 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible)대로 닫음. 있는 감사는 FAIL 0.62 @c445ce9559 하나이며 재감사는 coordinator가 돌린다. PASS라고 적지 않는다. 대응표는 §J.
- 0.4.0 (2026-10-30, HEAD `c83fdc04ff` 위 작업 트리): revised, unaudited — plan-audit FAIL 0.72 @081eacf037 수리. 마지막 감사는 그 FAIL이며 재감사는 coordinator가 돌린다. PASS라고 적지 않는다. 대응표는 §K.
- 0.5.0 (2026-10-30, HEAD `2acdddeba4` 위 작업 트리): audited-FAIL 0.70 → revised (A-na delta), unaudited — plan-audit FAIL 0.70 @2acdddeba4의 축소 델타. 리더 결정(A-na 축소 델타 뒤 멈춤)으로 이 판 뒤 감사는 없다. PASS라고 적지 않는다. 대응표·남은 부채·hold 기록·분할 제안은 §L.
- 0.5.1 (2026-10-31, HEAD `4a2123c8e7` 위 작업 트리): audited-FAIL 0.74 (plan debt accepted by leader decision), no re-audit — plan-audit FAIL 0.74 @4a2123c8e7(`.moai/reports/issue-1948/plan-audit-ana-delta.md`)의 N1·N2·N3·N4·N6 기록 수리. 결정 D-spec-disposition = C로 추가 plan 감사는 없다. PASS라고 적지 않는다. 결정 기록·plan 부채 목록·대응표는 §M.
- 0.6.0 (2026-11-01, HEAD `12eaa6c23f` 위 작업 트리): revised from landed code, unaudited, no re-audit — 감사 N19와 리더 지시로 착지 코드(저장소 묶음, reclaim-cap 출구, Enqueue 거절, evidence_mark·스키마 2판, trunk 고정·gate-logs·merged 사실·OnTrunk·DONE의 trunk 확인)를 SPEC에 적음. 직전 코드 감사는 46dbcce0d4에 대한 FAIL 0.76(receipt `rcpt-08a37bf592a0ccc3e932a935`, 리더가 전한 값; 이 작성자는 verdict를 읽지 않음)이고 그 뒤 수리가 있었다. 이 판에 대한 감사는 없다. PASS라고 적지 않는다. 대응표·남긴 문장·열린 부채는 §O.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §G Repair of plan-audit FAIL 0.62 @c445ce9559

- 원천: 감사 verdict `.moai/reports/issue-1948/plan-audit-local-001-c445ce9559/verdict.md`(FAIL 0.62, 막는 결함 D1 D2 D3 D8 D9 D10). 고정한 계약 sha256 `f2c1078b7add847b93926ba3356eda62e4fbc44d8aa77221f059f44f6b059acb`(spec §A.4).
- 판: 하위 0.2.0, 상위 0.7.1 — 둘 다 revised, unaudited. 감사·리뷰 도구는 부르지 않았다(외부 리뷰 호출 보류 지시).
- repair commit SHA: `db3cde6a30811bf2d4971e1d63206cec75452d17` (이 줄을 채우는 것은 그 뒤의 별도 커밋)
- "고치기 전" 줄은 `c445ce9559`의 줄, "고친 뒤" 줄은 이 작업 트리의 줄이다.

| defect id | file:line before | change | new file:line | OD | note |
|---|---|---|---|---|---|
| PIN-1 | (없음) | 원천 계약 sha256 고정 섹션 신설(경로·sha256·측정 명령·`date -u` 시각·재정렬 문장) | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:51 (§A.4), .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/design.md:3-8 | — | D13(SHOULD-FIX) 가운데 고정 부분; 고정이 이 수리의 첫 단계 |
| PIN-2 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/design.md:27-130 (§1~§6) | 고정한 계약 9~112행과 줄 단위 비교: 104줄 대 104줄 같음, 0.1.0 사본에서 바뀐 것 없음 | 같은 줄(머리말이 늘어 위치만 이동) | — | 0.1.0 사본에도 지금의 I7 문장이 이미 있었다 |
| PIN-3 | (없음) | 계약 §9·§10을 제목부터 파일 끝까지 바꾸지 않고 덧붙임(105줄, 계약 157~261행과 같음); §8은 범위 줄만 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/design.md:142 (§8 범위 줄), :146-250 (§9·§10) | — | §8은 옛 factory 경로(lane-20), 이 SPEC 밖 |
| PIN-4 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:72 (REQ-AFL-008) | I8 결속: tree SHA → `evidence.subject_sha`(commit SHA) == `candidateSHA`, 어기면 `ErrEvidenceStale`, MERGING은 `Enqueue`로만, store는 git을 부르지 않음 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:80; 이전 문구 §C.6 | OD-40 | 계약 §10 lane-22 항목의 문언 |
| PIN-5 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:71 (REQ-AFL-007), :70 (REQ-AFL-006) | 셋째 회수 park 사유 `reclaim-cap`; `fix-limit`·`reclaim-cap`이 core의 park 가능 집합 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:79, :78; 이전 문구 §C.6 | OD-35 (a) 고침 | 계약 I7·§10 |
| PIN-6 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:79 (REQ-AFL-010) | Claim의 lazy reap: 트랜잭션 밖에서 활성 판정, `BEGIN IMMEDIATE` 안에서 epoch·`heartbeat_at` 불변일 때만 회수; `Reap(nil)` 오류 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:87; 이전 문구 §C.6 | — | 계약 §10 첫 항목 |
| PIN-7 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:88 (REQ-AFL-014) | QueueUpdate epoch 원천(waiting·verifying·failed = Enqueue 때 기록한 card epoch == `cards.epoch`, landing·landed = landing 잠금 epoch) | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:96; 이전 문구 §C.6 | OD-40 | §10 19th 문언을 씀(18th와 다름) |
| PIN-8 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:88 (REQ-AFL-014) | merged·closed·settled 사실은 store가 자기 표에서 끌어냄, 호출자 값·쪽지에서 아님 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:96 | OD-40 | §10 19th 문언을 씀(18th와 다름) |
| PIN-9 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:88 (REQ-AFL-014) | trunk는 어느 worktree에도 checkout되지 않은 전용 ref(기본 `refs/heads/af/trunk`), checkout되어 있으면 착지 거절 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:96 | OD-35 (l) | 설정 수단은 미정 |
| PIN-10 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:69 (REQ-AFL-005) | 기한 지난 lease 소유자가 살아 있으면 card·epoch마다 `stalled` 쪽지 하나 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:77; 이전 문구 §C.6 | OD-41 | 계약 §10 20th (d); 수리 지시 사례표와 다름 |
| D1 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:71 (REQ-AFL-007) | "never delete a row" → `cards`·`transitions`·`evidence`·`notes`·`merge_queue` 행은 지우지 않고 `leases`·`lease_keys`만 `Release`·`Reap`으로만 지움, 셋째 회수 `reclaim-cap`. author's reconciliation of a contract conflict (I7 vs §3 Release), reversible | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:79; 이전 문구 §C.6; .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:88 (AC-AFL-007 Then 갱신, 이전 문구 bullet) | OD-36; OD-35 (a) 고침(.moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:357 → :358) | 고정한 I7에 `merge_queue`는 없다 — 조정자가 더한 것 |
| D2 | .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:393 (§J REQ-AF-007), :418 (AC-AF-007), :149 (REQ-AF-007 표지), .moai/specs/SPEC-AGENTIC-FACTORY-001/acceptance.md:189 | superseded → 혼합: 원격 부분(PR receipt·PR 적체 상한)만 superseded(REQ-AFL-014), 원격 무관 의무(같은 HEAD 코드·보안 리뷰 PASS, 빚 없는 독립 감사 PASS, QA PASS 등)는 no local counterpart yet; 표지 `[PARTLY SUPERSEDED — OD-37]`, 0.7.0 처분은 칸에 남김. AC-AF-009는 다시 읽고 처분 그대로 | .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:403, :428, :430(AC-AF-009 재확인), :150; .moai/specs/SPEC-AGENTIC-FACTORY-001/acceptance.md:189 | OD-37 | 하위 REQ를 지어내지 않음 |
| D3 | .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:397 (§J REQ-AF-011), :422 (AC-AF-011), :153 (REQ-AF-011), .moai/specs/SPEC-AGENTIC-FACTORY-001/acceptance.md:240 | unchanged → 혼합; IDENTITY REQ-AFI-001(online control ref) 의존을 §J에 적음; 표지 `[PARTLY SUPERSEDED — OD-38]` | .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:407, :432, :154; .moai/specs/SPEC-AGENTIC-FACTORY-001/acceptance.md:240 | OD-38 | IDENTITY-001은 고치지 않음 |
| D8 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:58-64 (AC-AFL-005) | Live·Unknown·Dead 사례표, 죽여야 할 변이(Live 소유자 회수), 시험 이름 하나 `TestReap_LivenessCases_ReclaimOnlyDead`; RED-now·green path 유지 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:63-79 | OD-41 | Live 행의 쪽지는 계약 문언(하나)을 씀 |
| D9 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:50-56 (AC-AFL-004) | `ledger` 디렉터리의 시험 아닌 .go 파일 정적 검사(SQL 리터럴의 다섯 시각 열이 `strftime(...,'now')`, `time.Now`·`Since`·`Until`·`After` 0건), 허용 예외 명시(e1 `nextWake` — OD-35 (f)), 빈 집합 확인, 시험 이름 하나 `TestLedgerSource_TimestampColumnsUseDBClock` | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:51-62 | OD 추가 없음(검사를 적을 수 있었음); OD-35 (f) 참조 | `importer` 하위 디렉터리는 덮지 않음(§E) |
| D10 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:178-184 (AC-AFL-020), .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:104 (REQ-AFL-020) | (i) 가짜 Reader 표지값으로 카드·lease·queue·쪽지 렌더 단언, (ii) `internal/web` 시험 아닌 .go 파일에서 `ledger.Store`·`ledger.Open` 0건; keys 출처 '미정'; 시험 이름 하나 `TestBoard_ReaderOnly_NoStoreReference` | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:195-205; .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:112; 이전 문구 §C.6 | OD-39 | 흐름 지표 표시는 판정하지 않음(§E) |
| D16(지시 밖, 이름 변경 때문에 함께) | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:13 | 자리표시자 명령 → 21개 이름을 펼친 명령, 다시 돌린 출력·종료 코드 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:13-20 | — | SHOULD-FIX/MINOR D16; 이름 넷이 바뀌어 다시 적어야 했다 |
| D13(고정만) | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/research.md:8 | 읽은 범위 줄 뒤에 0.2.0 재독 기록(§1~§10, sha256) 추가 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/research.md:15 | — | D13의 §8~§10 처분은 PIN-3·아래 목록 |
| OD-42(충돌 기록) | (없음) | 계약 §4 Retirement ↔ §10 P2 retirement plan 차이를 OD로 기록, REQ-AFL-015는 바꾸지 않음 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:176; .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:365 | OD-42 | 수리 지시의 §10 여섯 항목 밖 |

### 반영하지 않은 §10 항목 (not reflected)

- B1(전체 CLI 실행은 P3에 lane-22 세션의 사용자 확인이 필요) — not reflected: P3 배포 단계이며 spec §I 범위 밖이다.
- "AddCard refuses missing, CANCELLED or cyclic dependencies" — not reflected: AddCard 검증을 싣는 하위 REQ가 없고, 의존 필드의 원천은 여전히 계약 §2에 없다(OD-34).
- BEFORE DELETE 트리거(RAISE(ABORT)), `synchronous=FULL`, DSN `recursive_triggers=ON` — not reflected: REQ-AFL-007·003이 정한 결과를 이루는 수단(HOW)이다. REQ-AFL-003에 넣으면 새 의무가 되어 지시 범위를 넘는다.
- KeysFor의 `ErrInvalid`(절대 경로·`..`·역슬래시), `testdata`/`fixtures`의 소유 패키지 키, 공백 든 DSN 시험, 서명 `KeysFor(paths) ([]string, error)` — not reflected: REQ-AFL-011에 새 의무가 된다. OD-35 (i)는 이것으로 답해지지 않는다.
- mergeq 세부(ErrRefMoved는 old 값 불일치에만, Trunk는 전체 `refs/heads/` 이름이고 기본 브랜치 거절, 잠금 안에서 checkout 재확인, `cmd.WaitDelay`, ReleaseLock 실패 기록; 20th lane-22 (a) verifying 되돌림 규칙, (b) `LC_ALL=C`·`LANG=C`, (c) trunk ≠ `refs/remotes/origin/HEAD` 대상) — not reflected: 구현 수준이며 REQ-AFL-014의 결과 문장 밖이다.
- 19th "Reconcile reverts verifying/landing rows to waiting only while it holds the `merge-queue:landing` lock" — not reflected: REQ-AFL-014는 "reconcile against git ancestry"까지만 적는다.
- 19th "Reap on a card already parked at its third reclaim: count increments, park stays, no error" — not reflected: AC-AFL-007은 셋째 회수만 본다.
- 19th "the 400-line review trigger is enforced by the runner, not the store", "an ancestor candidate settles as landed only behind its I8 evidence at Enqueue" — not reflected: REQ-AFL-013은 집행 주체를 적지 않고, 후자는 REQ-AFL-008의 결과에 이미 걸린다.
- 18th "receipt / receipt-accepted are unused on the local path …" — 문장 그대로는 not reflected; REQ-AFL-014의 "no leader approval receipt SHALL be required"가 결과를 싣는다.
- 20th lane-18 (a) 한 카드의 오류는 그 카드만 건너뛰고 `stalled` 쪽지, (b) DONE/CANCELLED로의 Move가 같은 트랜잭션에서 lease·키를 풂(OD-36에서만 다룸), (c) `reclaim-cap`에서 재개하면 `reclaim_count` 초기화, (e) CANCELLED 의존 카드 건너뛰고 쪽지, (f) AddCard가 첫 transition(`added`)을 씀 — not reflected: 지시가 정한 여섯 항목 밖의 새 의무다.
- contract correction: `cards.epoch` 열(OD-40 (4)), RED 커밋의 build tag와 bundle 1의 같은 커밋 RED 편차 — not reflected: 표 열은 OD-40, build tag는 수단(HOW)이고 AC green path의 "RED 커밋이 GREEN보다 먼저"와 어긋나지 않는다.
- P2 retirement plan과 그 revision(lane-20이 원격 퇴역 전부) — not reflected: REQ-AFL-015와의 차이를 OD-42에 적었다.
- zombie-fix 결정(56e32db128 검토)과 20th lane-20 — not reflected: §8 옛 factory 경로, 이 SPEC 밖.

### 손댄 SHOULD-FIX

- D13: 고정(PIN-1~3)과 research.md 한 줄. §8~§10 처분은 위 목록.
- D16: 시험 이름 넷이 바뀌어 §B 명령을 펼쳐 다시 적음.
- D4·D5·D6·D7·D11·D12·D14·D15·D17: 손대지 않음.

## §J OD-36..42 closed

- 원천: leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible). 파일 `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/leader-decisions-od36-42-20261030.md`. 사용자·운영자 결정이 아니다. `hold-20261030.md` "Hold lifted" 절의 사용자 답은 리더가 옮겨 적은 것이며 원문은 읽지 않았다.
- 판: 하위 0.3.0, 상위 0.7.2 — 둘 다 revised, unaudited.
- 원천 계약 sha256 `f2c1078b7add847b93926ba3356eda62e4fbc44d8aa77221f059f44f6b059acb` — 작업 전에 다시 재어 같음을 확인했다.
- commit: `457b61967c4b54fc7c7bbfa896ec684de2bd7d63` (이 줄을 채우는 것은 그 뒤의 별도 커밋)

| OD | decision | files:lines changed | note |
|---|---|---|---|
| OD-36 | Release·Reap만 lease·lease_keys 행을 지움; DONE/CANCELLED Move는 같은 트랜잭션 안에서 Release 경로를 부름(셋째 경로 아님); cards·transitions·evidence·notes·merge_queue 행은 안 지움; 0.2.0 조정 유지 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:87 (REQ-AFL-007: Move 문장 추가, 표지 "leader decision OD-36"), :123-139 (§C.6 0.2.0 문장); .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:89-99 (AC-AFL-007 사례 (ii) DONE·CANCELLED Move, 변이, 이전 문구) | AC-AFL-007 시험 이름은 `TestReap_ThirdReclaim_ParksCard` 하나 그대로 |
| OD-37 | 하위에 새 REQ 없음; REQ-AF-007의 원격 무관 의무는 Enqueue 전 관문 증거가 싣는 절차 의무; 증거 종류 모형화는 배포 뒤 후속 카드 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:88 (REQ-AFL-008 괄호 메모 — SHALL 문장은 그대로); .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:404 (§J REQ-AF-007), :429 (§J AC-AF-007); .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md §E 한 줄 | 계약은 이 의무를 적지 않으므로 SHALL에 넣지 않음 |
| OD-38 | IDENTITY-001은 이번에 개정하지 않음; online control ref 의존은 로컬 경로에서 superseded로 기록; 배포 뒤 후속 카드가 개정 | .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:408 (§J REQ-AF-011), :433 (§J AC-AF-011) | IDENTITY-001은 건드리지 않음 |
| OD-39 | `Lease`가 `Keys []string`을 실음; Reader에 키 접근자 없음; 보드는 Lease에서 키를 읽음 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:120 (REQ-AFL-020, 미정 지움); .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:202-213 (AC-AFL-020 가짜 Reader에 Keys 실은 Lease, 판정에 키 넣음, 이전 문구) | AC-AFL-020 시험 이름은 `TestBoard_ReaderOnly_NoStoreReference` 하나 그대로 |
| OD-40 | `evidence.subject_sha`가 권위, I8·§2 tree_sha 문언은 §10이 대신함; §2는 cards.epoch·merge_queue.card_epoch·evidence.subject_sha 포함으로 읽음; QueueUpdate epoch·closed/settled는 19th | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:59-64 (§A.5 읽기 메모), :88 (REQ-AFL-008 묶음 이름·괄호), :94 (REQ-AFL-009 읽기 괄호), :104 (REQ-AFL-014 괄호); .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/design.md:9-11 (사본 밖 읽기 메모; 사본은 그대로) | REQ/AC 본문의 `tree_sha` 0건(아래 grep) |
| OD-41 | §10 20th (d): 기한 지난 Live 소유자의 lease에 card+epoch마다 `stalled` 쪽지 정확히 하나 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:85 (REQ-AFL-005 표지만); .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:70 (Live 행, 어긋남 표시 지움), 이전 문구 bullet 한 줄 | 문장 자체는 0.2.0에서 이미 계약 문언 |
| OD-42 | §10 P2 retirement plan: lane-20이 퇴역 전부를 한 변경으로, 넓은 대상 집합 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:110 (REQ-AFL-015 대상·순서·게이트·범위 밖 목록), §C.6 0.1.0 문장; .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md:158-169 (AC-AFL-015 대상·소유 lane-20·계약 인용·양성 대조·이전 문구) | `identity` GitHub 조회·`intake` GitHub 검색 포함 여부만 미정; 시험 이름 하나 그대로 |
| 공통 | 각 OD 행을 "닫힘(2026-10-30): decided by leader, reversible"로 표시, 열려 있던 문언은 history로 남김 | .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:184-190 (§G), .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:360-366 (§G.1), .moai/specs/SPEC-AGENTIC-FACTORY-001/decisions-pending.md:69-75, :79 (행 수 59, 열린 행 52); HISTORY .moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md:27 (0.3.0), .moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md:40 (0.7.2) | OD-35는 그대로(이번 결정으로 닫힌 항목 없음, (a)는 0.7.1에서 이미 닫힘) |

## §K Repair of plan-audit FAIL 0.72 @081eacf037

- 원천: 감사 verdict `.moai/reports/issue-1948/plan-audit-local-001-081eacf037/verdict.md`(FAIL 0.72 @081eacf037). OD-36~42는 다시 열지 않았다.
- 판: 하위 0.4.0, 상위 0.7.3 — 둘 다 revised, unaudited. 감사·리뷰 도구는 부르지 않았다.
- 원천 계약 sha256 `f2c1078b7add847b93926ba3356eda62e4fbc44d8aa77221f059f44f6b059acb` — 작업 전에 두 번 다시 재어 같음을 확인했다.
- 코드 사실의 commit: integration tip `1a552af405`(`git branch --list WT-agentic-factory-integration --format='%(refname:short) %(objectname:short)'` → `WT-agentic-factory-integration 1a552af405`; 조정자가 적은 `915a6524cb`에서 움직임 — `git log --oneline 915a6524cb..1a552af405`는 merge(lane-20) 퇴역과 merge(lane-18) GREEN 묶음 등 16줄). lane-18 커밋 `607bd17804`(`git log -1 --format='%h %s' 607bd17804` → `607bd17804 [AF-18] test(ledger): SPEC AC-named tests for lane-17 N1, N2, N14`; `git branch --contains 607bd17804 --list WT-agentic-factory-integration WT-agentic-factory-core-claim-runner` → `+ WT-agentic-factory-core-claim-runner`, integration에는 없음).
- commit: `50119a13da9520dc2f9cf1177b6a21663e419ad6` (이 줄을 채우는 것은 그 뒤의 별도 커밋)
- "고치기 전" 줄은 `c83fdc04ff`의 줄, "고친 뒤" 줄은 이 작업 트리의 줄이다.

| defect | before file:line | change | after file:line | note |
|---|---|---|---|---|
| N1 (BLOCKING) | ACC:121 (AC-AFL-010 Given) | 같은 시험 이름 아래 lazy reap 사례표 L-Live(회수 안 함, card·epoch마다 `stalled` 쪽지 정확히 하나, 두 번 Claim해도 하나)·L-Unknown(살아 있는 것으로 다룸)·L-Dead(회수)·L-Changed(probe 뒤 `heartbeat_at`·epoch 바뀜 → 회수 안 함); 변이 "Claim이 Live 소유자를 회수"; AC-AFL-005는 그대로 | ACC:167-179 | 이름 `TestClaim_SingleTransaction_ReadyToBuildingByClass` 그대로 — 착지 시험 하나가 AC 전체를 덮지 않음(§F). L-Live는 `607bd17804` `TestClaim_LazyReap_DoesNotReclaimLiveOwner`가 덮음 |
| N2 (BLOCKING) | ACC:38 (AC-AFL-002 Given) | 여섯 메서드 사례표 S-Move·S-ExpandKeys·S-AddEvidence·S-Heartbeat·S-Release·S-Enqueue, 모두 `ErrStaleEpoch`·바뀐 행 0; S-Release는 새 소유자의 lease·`lease_keys`를 지우지 않음; 변이 "Release가 `WHERE card_id=?`만"; QueueUpdate 잠금 epoch는 AC-AFL-014로 | ACC:58-71 | 이름 하나 그대로. 같은 이름 시험은 `607bd17804`에 있고 세 메서드만 덮음; 나머지는 §F의 추가 시험 |
| N3 (BLOCKING) | PL:46-48 (M6); PS:244 (DEC-AF-30) | M6을 lane-20 한 변경·REQ-AFL-015 대상·`go build ./... && go vet ./...` 게이트로; DEC-AF-30 순서 문장을 계약 §10 P2 revision으로; 둘 다 옛 문장을 이전 문구로 남김 | PL:48-53; PS:245 | 다른 곳 grep(`lane-22`·`lane-18` + 퇴역/retire, `패키지마다`, `import 먼저`, `controlref`): 살아 있는 문장은 위 둘뿐. `DS:127-130`은 계약 §4의 그대로 사본(그 사본 안 §10 P2 revision이 대신함)이라 고치지 않음; ACC AC-AFL-015 0.1.0 이전 문구와 SPEC §C.6 0.1.0 문장은 기록 |
| N17 (일부, N3과 함께) | PL:1 (머리 "(0.1.0)"), PL:63 (§F "0.7.0") | 머리 판 0.4.0, §F 상위 판 0.7.3; 옛 문구 남김 | PL:1-3, PL:68 | design.md REQ 대응표·progress §G 목록의 낡은 줄은 이번에 손대지 않음(아래) |
| N4 | PS:449 (§J 집계); PS:38 (HISTORY 0.7.0); PP:49 (노트) | 집계를 표에서 다시 셈: REQ 7/5/6/7, AC 8/5/6/6(superseded / no local counterpart / unchanged / 혼합), 센 방법 적음, 옛 줄은 이전 문구; HISTORY 0.7.0·progress 노트는 "0.7.0 시점의 수" 덧붙임만 | PS:450-454; PS:38; PP:49 | 아래 "§J 다시 셈" 출력. 0.7.0 수(8/5/7/5, 9/5/7/4)는 0.7.0 시점에는 맞았다 — 0.7.1 재분류가 집계 줄에 반영되지 않았던 것 |
| N5 | ACC:54-57 (AC-AFL-004) | 리터럴과 패키지 수준 문자열 상수를 `+`로 이은 식을 `go/constant`로 접어 읽음; 접을 수 없는 피연산자가 다섯 열에 값을 넣으면 FAIL; Go 시각 쓰기 0 요구 유지; 변이 (m1) Go 시각 값, (m2) `time.Now().Format` 이어 붙이기 | ACC:89-94; 0.3.0 문장은 ACC:99 이전 문구 | 착지 사실 `1a552af405:internal/agenticfactory/ledger/schema.go:24` `const nowSQL`(아래 출력) |
| N6 | ACC:92-93 (AC-AFL-007) | 사례 (iii): `Release`·`Reap`·DONE `Move` 뒤 `evidence`·`notes`·`merge_queue` 행 수 같음; 변이 추가 | ACC:134-136 | 이름 하나 그대로 |
| N8 | ACC:164 (AC-AFL-015 변이) | 사례 (b) 퇴역 대상 디렉터리 부재(import 안 돼도 남으면 FAIL), (c) `resident` 포함; 변이 추가 | ACC:230-232 | `37b67f6d39`에 `internal/agenticfactory/resident` 있음, `1a552af405`에는 대상 디렉터리 모두 없음(아래 출력) |
| N13 | ACC:6-8 (§A), ACC:13-23 (§B), 21개 AC의 RED-now·green path 줄, ACC:224 (§D) | §A 0.4.0 규칙(19개 AC regression-guard, `verification-completeness.md` §2.1); §B.1 다시 잰 것; 각 AC RED-now를 "name absent at 37b67f6d39; at 1a552af405 …"로, green path를 regression-guard 문장으로(013·017은 RED 먼저 경로 유지); 옛 두 칸은 AC마다 이전 문구; §D 갱신; 새 §F 'AC ↔ landed test' 21행 | ACC:8-9, ACC:27-42, 각 AC 블록, ACC:300-301, ACC:318-344 | 이름 채택은 AC-AFL-021 하나(`TestLedgerHeartbeat_ThrottlesToTheStampFile`, 완전히 덮음; ACC:292-293). AC-AFL-002·005는 이름이 이미 같음 |
| N14 | ACC:153 (AC-AFL-014) | 사례 (b) 낡은 landing 잠금 epoch의 `QueueUpdate` 거절, (c) trunk ref checkout 때 착지 거절; 변이 추가 | ACC:216-218 | 이름 하나 그대로; (c)는 mergeq(lane-22) 몫 |
| §E 동기화 | ACC:232, :236, :237 | N5·N14·N1로 바뀐 범위를 반영, 옛 문장은 이전 문구; AC-AFL-019 `Measure`의 probe 입력 사실과 "시험 안 돌림" 줄 추가 | ACC:309, :313, :314, :315-316 | AC-AFL-019 문장과 착지 코드의 관계는 정하지 않음(감사 대상) |
| 판 | SL:4, SL:27; PS:4, PS:40; DP:3 | 하위 0.4.0·상위 0.7.3 HISTORY 행, decisions-pending 머리 판 번호 | SL:4, SL:28; PS:4, PS:41; DP:3 | design.md·research.md 머리 판은 고치지 않음(내용 안 바뀜) |

약어: ACC = `.moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/acceptance.md`, PL = `.moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/plan.md`, SL = `.moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md`, DS = `.moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/design.md`, PS = `.moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md`, PP = `.moai/specs/SPEC-AGENTIC-FACTORY-001/progress.md`, DP = `.moai/specs/SPEC-AGENTIC-FACTORY-001/decisions-pending.md`.

### 이번에 고치지 않은 것 (not repaired this round)

- N7 — not repaired this round (SHOULD-FIX; 열린 OD 또는 리더 결정에 걸림).
- N9 — not repaired this round (SHOULD-FIX; 열린 OD 또는 리더 결정에 걸림).
- N10 — not repaired this round (SHOULD-FIX; 열린 OD 또는 리더 결정에 걸림).
- N11 — not repaired this round (SHOULD-FIX; 열린 OD 또는 리더 결정에 걸림).
- N12 — not repaired this round (SHOULD-FIX; OD-40의 `tree_sha` 열 처분이 열려 있음).
- N15·N16·N18과 N17의 나머지(design.md REQ 대응표, 위 §G "반영하지 않은 §10 항목"의 낡은 줄) — 이번 수리 목록 밖이라 손대지 않음.

### 측정 출력 (이번 실행, 이 트리)

```text
$ git grep -n 'nowSQL' 1a552af405 -- internal/agenticfactory/ledger   (발췌)
1a552af405:internal/agenticfactory/ledger/schema.go:23:// nowSQL is the database clock as an SQL expression; offsetSQL(d) is d from now.
1a552af405:internal/agenticfactory/ledger/schema.go:24:const nowSQL = `strftime('%Y-%m-%dT%H:%M:%fZ','now')`
1a552af405:internal/agenticfactory/ledger/cards.go:73:		VALUES(?,?,?,?,?,?,?,?,?,?,?,0,`+nowSQL+`,`+nowSQL+`)`,

$ §J 다시 셈 (python3 -I, 상위 spec.md §J 표의 REQ·AC 행, 셋째 칸 첫머리: 혼합 / superseded — no local counterpart / superseded / unchanged)
('AC', 'mixed') 6 005·007·010·011·013·016
('AC', 'no-local-counterpart') 5 003·008·012·014·022
('AC', 'superseded') 8 001·002·006·009·017·018·019·025
('AC', 'unchanged') 6 004·015·020·021·023·024
('REQ', 'mixed') 7 005·007·010·011·013·016·018
('REQ', 'no-local-counterpart') 5 003·008·012·014·023
('REQ', 'superseded') 7 001·002·006·009·017·019·020
('REQ', 'unchanged') 6 004·015·021·022·024·025

$ git -C <worktree> grep -n -e 'agenticfactory/forge"' -e 'agenticfactory/mirror"' -e 'agenticfactory/gitref"' -e 'effect/ghintake"' -e 'effect/ctxledger"' -e 'effect/labels"' -e 'effect/ghexec"' -e 'effect/ghidentity"' -e 'effect/gitexec"' 1a552af405 -- internal
(출력 없음; 종료 코드는 이 환경에서 보이지 않음 — acceptance §B.1)

$ git -C <worktree> ls-tree -d -r --name-only 1a552af405 internal | grep -i -e resident -e '/forge$' -e '/mirror$' -e '/gitref$' -e 'effect/'
internal/agenticfactory/effect/internal
internal/agenticfactory/effect/internal/procrun
internal/agenticfactory/effect/testdata
(같은 명령 @37b67f6d39: effect/{ctxledger,docstore,ghexec,ghidentity,ghintake,gitexec,internal,internal/procrun,labels}, forge, gitref, mirror, resident)
```

21개 이름 부재 확인(기준 `37b67f6d39`, tip `1a552af405`, `607bd17804`)의 명령·출력은 acceptance §B.1에 있다.

## §L A-na delta after FAIL 0.70 @2acdddeba4

- 원천: 감사 verdict `.moai/reports/issue-1948/plan-audit-local-001-2acdddeba4/verdict.md`(FAIL 0.70 @2acdddeba4; 막는 결함 B1~B4, 선택 D5·D6·D7·D8). 이 판은 조정자가 정한 항목만 고쳤다(A-na 축소 델타). 감사·리뷰 도구는 부르지 않았다.
- 판: 하위 0.5.0, 상위 0.7.4 — 둘 다 audited-FAIL 0.70 → revised (A-na delta), unaudited. PASS라고 적지 않는다.
- 원천 계약 sha256 `f2c1078b7add847b93926ba3356eda62e4fbc44d8aa77221f059f44f6b059acb` — 작업 전에 `shasum -a 256 /Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`로 다시 재어 같음.
- 코드 사실의 commit: integration tip `db02401ad1`(`git branch --list WT-agentic-factory-integration --format='%(refname:short) %(objectname:short)'` → `WT-agentic-factory-integration db02401ad1`; 감사가 인용한 `c83013938a`에서 움직였고, 이 작업 중 한 번 더 움직였다 — 앞서 읽은 `3df062a63e`의 `pickTx`·상수 줄은 `db02401ad1`에서 다시 읽어 같음을 확인). lane-18 커밋 `607bd17804`의 `ac_names_test.go`는 tip에 있다(acceptance §B.2 [L-names]).
- commit: `fc207c901d72c35c70cdc53995187483d56eae3c` (이 줄을 채우는 것은 그 뒤의 별도 커밋)
- "고치기 전" 줄은 `2acdddeba4`의 줄, "고친 뒤" 줄은 이 작업 트리의 줄이다. 약어는 §K와 같다.

| item | before file:line | change | after file:line | note |
|---|---|---|---|---|
| (2) B2 재claim 조건 | SL:96 (REQ-AFL-010), SL:88 (REQ-AFL-007), ACC:58 (AC-AFL-002 Given), ACC:133 (AC-AFL-007 Given (i)), ACC:167 (AC-AFL-010 Given) | REQ-AFL-010: "READY or that a reclaim has returned to the claimable set"; 그 집합은 미정(OD-43); 계약 §3·I7·§10 인용; `pickTx` 코드 사실(`db02401ad1:internal/agenticfactory/ledger/claim.go:351-358`, `:296`, `:311`)은 요구사항이 아닌 사실로; REQ-AFL-007은 메모 한 문장; AC-002·007·010에 '전제(0.5.0)' 줄; 새 OD-43(neutral, reversible) | SL:97, SL:89, SL:141 (0.4.0 문장), ACC:121, ACC:197, ACC:232; SL:196 (§G OD-43); PS:369; DP:76 | 계약은 회수된 BUILDING·REVIEWING 카드의 재claim을 적지 않는다 — 새 의무를 만들지 않았다 |
| (3) B3 reconcile 예외 | SL:105 (REQ-AFL-014), ACC:216-218 (AC-AFL-014) | REQ-AFL-014에 "a reconcile write that reverts a `verifying` or `landing` row to `waiting` SHALL be accepted only while the merge queue holds the `merge-queue:landing` lock and SHALL carry that lock epoch instead of the card epoch"(계약 §10 19th review L194-195 인용); AC-AFL-014 사례 (d1)(d2)(d3)과 변이 둘; 착지 시험 `queue_test.go:170-176`·`mergeq_test.go:586`을 사실로 인용 | SL:106, SL:142 (0.4.0 문장), ACC:285-288 | 시험 이름 하나 그대로. verifying 행 되돌림과 lane-22 (a) 조건은 AC 밖(ACC §E 0.5.0) |
| (4) B4 §F 낮춤 | ACC:318-344 (§F) | 머리말을 tip `db02401ad1` 기준·본문 판정 규칙으로; S-ExpandKeys·S-Enqueue "partly: future-epoch only; stale (older) epoch not asserted"; L-Unknown "partly: asserts lease count only, no other READY card"; AC-AFL-021 fully → partly(30 s 고정 없음, stamp 파일 단언 없음, 1 s 변이 통과); L-Changed-epoch 착지 시험 없음; AC-014 사례 (d) 착지 시험 추가; commit 접두를 tip으로, 보드 줄 213 → 216 | ACC:395-425 | 다른 행은 본문을 다시 보고 그대로 둠(이미 partly/not). AC-005 fully는 tip 본문 diff 없음으로 유지 |
| D5 §6 상수 | SL:198 (§H), ACC:307 (§E REQ-AFL-018) | '미정'을 착지 사실로: 여덟 상수 `db02401ad1:internal/agenticfactory/ledger/api.go:41-48`; 훅의 별도 `LedgerHeartbeatEvery`(`internal/hook/ledger_heartbeat.go:20`); 상위 OD-35 (d)를 "answered by landed code @db02401ad1"로 표시, 다른 항목은 열림 | SL:203, ACC:381; SL:188 (§G OD-35 행); PS:361; DP:68 | AC-AFL-018의 Then(상수 넷)은 바꾸지 않음 |
| D6 부재 주장 날짜 | SL:72, SL:144, SL:196, SL:202; ACC:6 | "at 37b67f6d39"를 달고, tip에 있는 이름 셋을 file:line으로; tip에서 22개 이름 grep 한 번 다시 돌림 | SL:73, SL:148, SL:201-207 일대; ACC:7; ACC:47 (§B.2) | tip에 있는 이름: `ac_names_test.go:30`, `:117`, `internal/hook/ledger_heartbeat_test.go:58` |
| D8 L-Changed 나눔 | ACC:175, ACC:177 | L-Changed-epoch·L-Changed-heartbeat 두 행, 변이도 둘; 옛 행은 이전 문구 | ACC:240-241, ACC:243-245 | heartbeat 쪽만 착지 시험(`reap_test.go:95`) |
| (1) B1 regression-guard 제거·§N | ACC:8 (§A), ACC:300 (§D), 19개 AC의 RED-now·green path 줄, PL:52, PL:63 | 'regression-guard' 표지를 지움; 착지한 AC 18개(001~012·014·016·018~021)의 두 칸을 "§N 사후 관측, RED 미관측, 완료 계산에 들지 않음" 한 줄로 바꾸고 옛 두 칸은 AC마다 이전 문구; 새 §N(AC마다 착지 시험·단언·덮음·통과 증거·'RED 미관측' 줄); §D 완료 정의는 release 게이트 AC 셋(013·015·017)만 셈; AC-AFL-015는 release 게이트에 두고 RED-now를 `37b67f6d39`의 퇴역 대상 디렉터리 열 개 존재로(트리를 읽음, 실행 아님); 013·017은 "unwritten test" 상태 | ACC:10-11, ACC:373-374, 각 AC 블록, ACC:427 이하 (§N), ACC:305-307 (AC-015), ACC:275·:325 (013·017), PL:54-55, PL:66-67 | 시험 이름 하나·제목은 21개 모두 그대로. 남은 'regression-guard' 줄은 모두 이전 문구·HISTORY 기록(이 판의 측정 출력) |
| 판·상태 표지 | SL:4, ACC:1, PL:1, PS:4, PS:41, DP:3, DP:79 | 하위 0.5.0, 상위 0.7.4, 상태 표지 "audited-FAIL 0.70 → revised (A-na delta), unaudited"; decisions-pending 행 수 60/열림 53 | SL:4, SL:29, ACC:1, PL:1, PS:4, PS:42, DP:3, DP:80 | design.md·research.md는 손대지 않음 |

### 남은 부채 (not repaired)

- N7 (FAIL 0.72 감사) — AC-AFL-020 (ii)가 쓰기 경로를 이름으로만 막음: AC를 넓히는 일이며 이 델타 밖.
- N9 (FAIL 0.72 감사) — 상위 REQ-AF-011 잠금 문장의 로컬 적용 범위: 상위 처분의 판단이 필요해 이 델타 밖(리더 결정 대상).
- N10 (FAIL 0.72 감사) — OD-37 "gate evidence" 집행 문장 없음: OD-37 결정이 새 의무를 두지 않기로 했고(leader decision), 바꾸려면 리더 결정이 필요.
- N11 (FAIL 0.72 감사) — Claim의 의존 조건 원천 없음: OD-34가 열려 있다.
- N12 (FAIL 0.72 감사) — `evidence.tree_sha` 열의 존폐 모호: OD-40 결정의 `tree_sha` 처분 해석이 열려 있다.
- N15 (FAIL 0.72 감사) — `identity`·`intake` GitHub 퇴역 포함 여부가 OD 없이 '미정': 새 OD를 올리는 일이 이 델타 목록 밖.
- N16 (FAIL 0.72 감사) — REQ-AFL-002 보편 문장과 REQ-AFL-014 잠금 epoch의 관계: 이 델타는 REQ-AFL-014에 reconcile 예외만 더했고 REQ-AFL-002는 손대지 않았다.
- N17 (FAIL 0.72 감사, 나머지) — design.md 교차 참조표: design.md(계약 사본 포함)는 손대지 않는다는 조건.
- N18 (FAIL 0.72 감사) — DONE 이동 뒤 `Release` 이중 호출 흐름: 이 델타 밖.
- D4 (FAIL 0.62 감사) — 상위 §J가 REQ-AF-008 보호 금지를 "no local counterpart"로 분류: 상위 처분의 판단이라 이 델타 밖.
- D7 (FAIL 0.62 감사; FAIL 0.70 감사의 선택 D7은 AC-005 Claim 경로 쪽지 중복 제거 시험 방향) — 상위 design.md §1 구조도의 권위 선언, 그리고 AC-005의 Claim 경로 방향: 앞의 것은 design.md를 손대지 않는다는 조건, 뒤의 것은 AC를 넓히는 일이라 이 델타 밖.
- D11 (FAIL 0.62 감사) — 나머지 AC의 한 방향 시험(변이 통과): AC를 넓히는 일이며, 그 AC 대부분이 이제 §N(release 게이트 밖)이다.
- D12 (FAIL 0.62 감사) — REQ-AFL-013이 계약 §5 의무(RED 증거 커밋, REVIEWING의 교차 모델 리뷰)를 줄임: REVIEWING 수단이 OD-35 (c)로 열려 있다.
- D15 (FAIL 0.62 감사) — GEARS·명확성 소소한 점(REQ-AFL-012의 `Where` 등): 이 델타 목록 밖.
- D17 (FAIL 0.62 감사) — 상위 frontmatter tags에 `github-issues, control-ref, context-ledger`가 남음: 이 델타 목록 밖.
- D5 (FAIL 0.70 감사) — 고쳤다(위 표). D6·D8도 고쳤다.

### hold 기록과 분할 제안 (`harness.plan_audit_ceiling_policy`)

- 근거 설정: `.moai/config/sections/harness.yaml`의 `plan_audit_tier_ceilings.L: 3`과 `plan_audit_ceiling_policy.on_final_hit: hold-and-split`(같은 파일 82-84행).
- 천장 도달 기록: 이 하위 SPEC(Tier L)의 plan-audit은 세 번이었다 — FAIL 0.62 @c445ce9559 → FAIL 0.72 @081eacf037 → FAIL 0.70 @2acdddeba4. 마지막 회는 점수가 내려갔다(0.72 → 0.70). Tier L 천장 3에 닿았다.
- 묶어 둔 것(hold): 이 SPEC은 PASS가 아니며, 이 판(0.5.0)도 감사받지 않았다. 리더 결정(A-na 축소 델타 뒤 멈춤)으로 추가 감사는 없다. 구현은 리더 결정에 따라 lane에서 계속된다 — 이 SPEC의 판정 상태가 그 구현을 승인하거나 막는다고 적지 않는다.
- 풀리는 조건: 이 판에서 정하지 않았다(리더·운영자 몫).
- 분할 제안(선택지일 뿐, 순서에 뜻 없음, 권장 표시 없음):
  - 선택지 A — 원장 핵심 SPEC(I1~I8, §2 스키마, §3 API, Claim·회수·fencing, 충돌 키; 지금의 REQ-AFL-001~012·018, AC 같은 번호)과 흐름·표면 SPEC(카드 흐름, 보드·flow, `moai af` 명령, 병합 큐, importer, 퇴역, 훅 heartbeat; REQ-AFL-013~017·019~021)으로 나눔. 비용: 두 SPEC 모두 새 id·새 plan-audit 한 번씩; REQ·AC id를 옮기며 이전 문구·대응표를 다시 써야 함; 원장 쪽은 착지한 코드가 많아 대부분 §N 사후 관측이 되고, 흐름·표면 쪽에 release 게이트 AC(013·015·017)가 모임; 상위 §J 처분표와 OD 행의 하위 참조를 두 SPEC으로 갈라 적어야 함.
  - 선택지 B — 착지한 18개 AC를 사후 관측 문서(SPEC 밖 기록)로 옮기고, release 게이트 AC 셋(013·015·017)과 그 REQ만 남긴 작은 SPEC으로 줄임. 비용: REQ 18개가 SPEC 밖으로 나가 요구사항 추적이 끊긴다; AC·REQ 삭제가 필요해 지금의 '지우지 않음' 조건과 부딪친다(운영자 결정 필요); 상위 §J가 가리키는 REQ-AFL id 다수를 다시 적어야 함.
  - 선택지 C — 나누지 않고 지금 SPEC을 유지하며 hold를 기록으로만 둠. 비용: 막는 결함 판정(감사 B1~B4가 이 판에서 고쳐졌는지)이 감사 없이 남는다; N7·N9~N12·D7이 세 회 연속 그대로라는 감사의 정체 신호가 풀리지 않는다.

### 측정 출력 (이번 실행, 이 트리)

```text
$ git -C <worktree> grep -n -e 'func TestLedgerImportGraph_NoIssueOrRemoteClient' -e 'func TestStore_StaleEpochWrite_ReturnsErrStaleEpoch' -e 'func TestOpen_SetsBusyTimeoutAndChecksSQLiteVersion' -e 'func TestLedgerSource_TimestampColumnsUseDBClock' -e 'func TestReap_LivenessCases_ReclaimOnlyDead' -e 'func TestMove_RefusesEdgeRejectedByCheckMove' -e 'func TestReap_ThirdReclaim_ParksCard' -e 'func TestEnqueue_EvidenceSubjectSHAMismatch_ReturnsErrEvidenceStale' -e 'func TestSchema_CheckConstraints_RejectUnknownQueueStateAndNoteKind' -e 'func TestClaim_SingleTransaction_ReadyToBuildingByClass' -e 'func TestExpandKeys_Conflict_ReturnsErrKeyConflict' -e 'func TestClaim_DisjointConflictKeys_ClaimInParallel' -e 'func TestCardFlow_ClaimToDone_ThroughMergeQueue' -e 'func TestMergeq_Land_CASUpdateRefThenLandedNote' -e 'func TestRetirement_NoImportOfForgeMirrorGitref' -e 'func TestImporter_DryRun_SourcesUnchanged' -e 'func TestAFCommand_VerbsRegistered' -e 'func TestLedgerDefaults_MatchContract' -e 'func TestFlow_StalledDetection_FromReader' -e 'func TestBoard_ReaderOnly_NoStoreReference' -e 'func TestLedgerHeartbeat_ThrottlesToTheStampFile' -e 'func TestHeartbeatWriter_SkipsWriteWhenStampFresh' db02401ad1 -- internal
db02401ad1:internal/agenticfactory/ledger/ac_names_test.go:30:func TestReap_LivenessCases_ReclaimOnlyDead(t *testing.T) {
db02401ad1:internal/agenticfactory/ledger/ac_names_test.go:117:func TestStore_StaleEpochWrite_ReturnsErrStaleEpoch(t *testing.T) {
db02401ad1:internal/hook/ledger_heartbeat_test.go:58:func TestLedgerHeartbeat_ThrottlesToTheStampFile(t *testing.T) {

$ git -C <worktree> grep -n -e 'LedgerHeartbeatEvery *=' -e 'HeartbeatThrottle' db02401ad1 -- internal
db02401ad1:internal/agenticfactory/ledger/api.go:41:	HeartbeatThrottle = 30 * time.Second // a session writes its heartbeat at most this often
db02401ad1:internal/hook/ledger_heartbeat.go:20:const LedgerHeartbeatEvery = 30 * time.Second

$ git -C <worktree> show db02401ad1:internal/agenticfactory/ledger/claim.go  (줄 351-358 발췌)
// pickTx is the highest-priority card nobody holds that can be claimed: READY, or
// BUILDING or REVIEWING with its holder gone; not parked; every dependency DONE; every
// conflict key free. A card that fails a test is skipped and changed in no way.
		WHERE parked=0 AND stage IN ('READY','BUILDING','REVIEWING')
		  AND NOT EXISTS (SELECT 1 FROM leases l WHERE l.card_id=c.id)

$ git -C <worktree> diff --stat 1a552af405 db02401ad1 -- (§F가 인용한 시험 파일 12개)
 claim_test.go 2 +-, move_test.go 12 +++++-, open_test.go 4 +-, queue_test.go 46 ++++----, review20_test.go 2 +-, mergeq_test.go 2 +-, flow_test.go 19 ++++, ledger_board_test.go 11 ++++--
$ git -C <worktree> diff --stat 607bd17804 db02401ad1 -- ac_names_test.go verifying_reentry_test.go
 ac_names_test.go 18 ++++----, verifying_reentry_test.go 10 +++---
(바뀐 곳: QueueUpdate epoch 인자 형식, Close 오류 처리, 가짜 큐 SetState 서명, 보드 flow 단언 — §F가 인용한 단언은 그대로)
```

AC-AFL-015 RED-now의 `ls-tree` 출력과 013·017·015의 단일 이름 grep은 acceptance §B.2에 있다. 종료 코드는 어느 명령에서도 관측하지 못했다(acceptance §B.2).

## §M plan-audit FAIL 0.74 @4a2123c8e7 — plan debt accepted by leader decision

- 결정(그대로): "plan-audit FAIL 0.74 accepted as plan debt (leader, operator delegation + Jev); code verdict = final sync audit on the frozen integration SHA"
- 출처: leader decision relayed via the leader session (message of this date); operator delegation and the Jev result are the leader's statements, originals not read by this author. 날짜: 2026-10-31(조정자가 이 작업에 실어 보낸 메시지).
- 결정 이름: D-spec-disposition = C — SPEC kept; no split, no reduction; no further plan audit. §L의 분할 제안 가운데 선택지 C에 해당한다(§L 선택지 목록은 지우지 않았다).
- 감사: `.moai/reports/issue-1948/plan-audit-ana-delta.md`(FAIL 0.74 @4a2123c8e7; 막는 결함 N1 N2 N3, 선택 N4 N5 N6 N7 N8 N10, 이월 묶음 "N9 (carried)"). 이 판은 감사·리뷰 도구를 부르지 않았고 plan 감사를 다시 돌리지 않았다.
- 상태 표지(하위·상위 판이 적힌 곳): audited-FAIL 0.74 (plan debt accepted by leader decision), no re-audit. PASS가 아니다. 판: 하위 0.5.1, 상위 0.7.5.
- 원천 계약 sha256 `f2c1078b7add847b93926ba3356eda62e4fbc44d8aa77221f059f44f6b059acb` — 작업 전과 끝에 `shasum -a 256 /Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md`로 다시 재어 같음.
- 코드 사실의 commit: integration tip `b28a31d9dc`(AC-AFL-015 시험 병합). §L·§N·§F의 다른 코드 사실은 여전히 `db02401ad1`에 고정되어 있다(감사 N8, plan 부채).
- commit: `d036ee30b0e885d0936cf86694ddc2cbae89c7cd` (이 줄을 채우는 것은 그 뒤의 별도 커밋)

### plan 부채로 받아들인 감사 id

- 리더가 이름을 댄 것(FAIL 0.74 감사): N5(§D 완료 정의가 REQ 18개를 뺌), N6(`api.go` 줄 범위 — 글자 실수라 이 판에서 41-48로 통일했으나 목록에는 리더가 적은 대로 둔다), N7(OD-43과 REQ-AFL-007), N10(§N AC-AFL-002 요약의 heartbeat — 착지 시험이 `heartbeat_at`을 단언하지 않음; 고치지 않았다).
- 이월(리더가 적은 대로): N9, N11, N12, N8, D7.
- §L에서 이월한 목록(그대로 옮김): FAIL 0.72 감사의 N7, N9, N10, N11, N12, N15, N16, N17, N18; FAIL 0.62 감사의 D4, D7, D11, D12, D15, D17.
- 주의: N 번호는 감사마다 다시 매겨진다 — FAIL 0.74 감사의 N7·N10과 FAIL 0.72 감사의 N7·N10은 다른 결함이다. 리더 메시지의 "carried N9, N11, N12"가 어느 감사의 번호인지는 메시지가 적지 않았다; FAIL 0.74 감사의 "N9 (carried)"는 §L 목록 전체를 묶은 항목이다.
- 이 판에서 고친 것(부채 목록 밖): N1(기록), N2, N4(이탈로 기록), N6. N3은 "not resolved by the report"로 적었다 — 고친 것이 아니다.

### 대응표 (이 작업)

약어는 §K와 같다(SL spec.md, ACC acceptance.md, PL plan.md, PS 상위 spec.md, DP 상위 decisions-pending.md). "고치기 전"은 `4a2123c8e7`의 줄, "고친 뒤"는 이 작업 트리의 줄이다.

| item | before | change | after | note |
|---|---|---|---|---|
| N2 AC-014 사례 (d) | ACC:285-286 | 사례를 (d1) 저장소 층, (d2) 저장소+병합 큐 층, (d2′) 병합 큐 층 반대 사례(현재 epoch·살아 있는 lease의 verifying 행은 그대로, 동작 0), (d3) 병합 큐 층으로 나눔; 계약 §10 20th review lane-22 (a) 256-257행 인용; 변이 둘 더함; 착지 사실 `b28a31d9dc:…/mergeq/real_ledger_reconcile_test.go:16`(본문 읽음, 돌리지 않음); 옛 (d)·변이 줄은 이전 문구(0.5.0) | ACC:298-302 | 시험 이름 하나 그대로 |
| N2 REQ-AFL-014 일관성 | SL:106 | SHALL 문장은 그대로; "언제나 되돌림"이 아니라는 메모 한 문장과 20th (a) 인용 | SL:107 | REQ-AFL-014는 원래도 "always reverts"라고 쓰지 않았다 |
| N1 AC-013 RED | ACC:275 | 'unwritten test' 칸 → lane report RED 기록(명령은 lane report af-verbs-ac017-ac013.md L32를 가리키는 포인터, 그대로 옮긴 FAIL 출력, `95e100e754`, 종료 코드 칸 없음 문장)과 이 작성자의 커밋 그래프 확인; 옛 칸 이전 문구. 후속(조정자 결정, 선택지 2): 실행할 수 있는 `go test … -run <Name> ./internal/cli/` 줄을 지우고 줄 번호 포인터로 바꿈; 앵커 없는 `-run` 패턴에 대한 발견 한 문장(`verification-completeness.md` §1.1, 기록만 함); 그 앞의 메모는 이전 문구 | ACC:277-286 | strict lint의 `VacuousTestAssertion`이 사라짐; lint.skip·우회 문구 없음 |
| N1 AC-013 GREEN | (없음) | GREEN 기록(`b58b1566b6`, lane report, 이 작성자 미관측)과 lane report의 Gaps | ACC:287 | 명령은 L44 포인터; Then의 REVIEWING 생략은 덮지 않음 |
| N1 AC-017 RED·GREEN | ACC:325 | AC-013과 같은 형식(`5a9370f894`; RED 명령 포인터 L20, GREEN 포인터 L50; 같은 발견 문장) | ACC:345-354 | AC-015 블록은 `-run`이 없는 명령이라 lint가 짚지 않아 그대로 둠 |
| N1·N4·N3 AC-015 | ACC:305-306 | lane-20 lane report RED·GREEN·변이 기록; 기록한 이탈(DEVIATION) 줄; N3 상태 줄(not resolved by the report, `8d0c47a7f3` ls-tree 부분 증거); 옛 RED-now·green path 이전 문구 | ACC:321-326 | |
| N1·N5 §D 완료 정의 | ACC:373 | 완료 = 기록된 RED(순서 포함, 015는 이탈) + GREEN 기록이 함께 있을 때; 판정이 아니라는 문장; 옛 문구 이전 문구 | ACC:403-404 | N5(REQ 18개 제외)는 고치지 않음 — plan 부채 |
| N6 줄 범위 | ACC:422 | `api.go:41-50` → `41-48`, `b28a31d9dc` `git grep -n` 재확인 | ACC:453 | |
| 판·상태 표지 | SL:4, SL:7, ACC:1, PL:1, PL:74, PS:4, PS:7, DP:3 | 하위 0.5.1, 상위 0.7.5, `updated: 2026-10-31`; 표지 "audited-FAIL 0.74 (plan debt accepted by leader decision), no re-audit"; 옛 머리는 이전 문구 | SL:4, SL:7, SL:30, ACC:1-3, PL:1-3, PL:74, PS:4, PS:7, PS:43, DP:3 | design.md·research.md 손대지 않음 |
| §E.1·§M | PS(하위 progress):13 | 0.5.1 줄; 이 §M | 하위 progress:14, §M | |

### 측정 출력 (이번 실행, 이 트리)

```text
$ git -C <worktree> log -1 --format='%h %s' <sha>   (sha마다 한 번)
5a9370f894 test(cli): RED - TestAFCommand_VerbsRegistered (AC-AFL-017, Issue #1948)
95e100e754 test(cli): RED - TestCardFlow_ClaimToDone_ThroughMergeQueue over a skeleton `moai af` (AC-AFL-013, Issue #1948)
b58b1566b6 feat(cli): `moai af` verbs - claim, move, evidence, enqueue, release (AC-AFL-017, AC-AFL-013, Issue #1948)
d649512ed7 [issue-1948-af] test(agenticfactory/retirement): TestRetirement_NoImportOfForgeMirrorGitref (AC-AFL-015)
b28a31d9dc merge(lane-20): AC-AFL-015 retirement import test (d649512ed7, Issue #1948)

$ git -C <worktree> branch --contains <sha> --list WT-agentic-factory-integration   (위 다섯 sha 각각)
+ WT-agentic-factory-integration

$ git -C <worktree> merge-base --is-ancestor 5a9370f894 b58b1566b6; echo "exit=$?"   → exit=0
$ git -C <worktree> merge-base --is-ancestor 95e100e754 b58b1566b6; echo "exit=$?"   → exit=0
$ git -C <worktree> merge-base --is-ancestor 5a9370f894 95e100e754; echo "exit=$?"   → exit=0
$ git -C <worktree> merge-base --is-ancestor b58b1566b6 5a9370f894; echo "exit=$?"   → exit=1   (반대 방향 대조)
$ git -C <worktree> merge-base --is-ancestor d649512ed7 b28a31d9dc; echo "exit=$?"   → exit=0

$ git -C <worktree> log -1 --format='%h %p %s' 3fd9ab6456
3fd9ab6456 8d0c47a7f3 [issue-1948] refactor(agenticfactory): retire the Issue/remote code under the local-only design (P2, OD-42)

$ git -C <worktree> grep -n -e HeartbeatThrottle -e SnapshotEvery -e MinSQLite b28a31d9dc -- internal/agenticfactory/ledger/api.go; echo "exit=$?"
b28a31d9dc:internal/agenticfactory/ledger/api.go:41:	HeartbeatThrottle = 30 * time.Second // a session writes its heartbeat at most this often
b28a31d9dc:internal/agenticfactory/ledger/api.go:48:	SnapshotEvery     = 30 * time.Minute // while the store is open
b28a31d9dc:internal/agenticfactory/ledger/api.go:50:	MinSQLite         = "3.35"           // Open refuses an older engine
b28a31d9dc:internal/agenticfactory/ledger/api.go:72:	ErrSQLiteTooOld   = errors.New("ledger: the SQLite engine is older than " + MinSQLite)
exit=0
```

- 종료 코드: §L의 "종료 코드는 어느 명령에서도 관측하지 못했다"는 그 실행의 명령 모양에 대한 기록이다. 이 작업에서는 git 명령 하나 뒤에 `; echo "exit=$?"`를 붙인 형태가 받아들여져 종료 코드를 관측했다. §L 문장은 고치지 않았다.
- lane report 두 개는 integration 트리의 gitignore된 `.moai/reports`에 있는 lane이 쓴 기록이며, 이 작성자는 읽기만 했고 그 안의 명령을 다시 돌리지 않았다.

### 후속 측정 (조정자 결정 선택지 2 뒤, 이번 실행, 이 트리)

```text
$ <scratchpad>/moai-tree5 spec lint --strict <worktree>/.moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md > <scratchpad>/lint-child4.txt 2>&1; echo "exit=$?"
exit=0
(파일 안) WARNING   REQTableRowsRejected   …/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md   139   …
(파일 안) 0 error(s), 1 warning(s)

$ <scratchpad>/moai-tree5 spec lint --strict <worktree>/.moai/specs/SPEC-AGENTIC-FACTORY-001/spec.md > <scratchpad>/lint-parent4.txt 2>&1; echo "exit=$?"
exit=0
(파일 안) WARNING   REQTableRowsRejected   …/SPEC-AGENTIC-FACTORY-001/spec.md   244   …
(파일 안) 0 error(s), 1 warning(s)
```

- 바로 앞 실행(포인터로 바꾸기 전)의 하위 strict lint는 `exit=1`, `0 error(s), 3 warning(s)`(REQTableRowsRejected 1, `VacuousTestAssertion` 2 — AC-013·AC-017 RED 블록)였다.
- 수: AC 제목 상위 25·하위 21, REQ 고유 상위 25·하위 21, 시험 이름이 정확히 하나인 하위 AC 21/21.

## §O 0.6.0 amendment from landed code (N19)

- 원천: 감사 N19와 리더 지시 "SPEC 개정: gate-logs·FailureLog·evidence_mark, REQ-AFL-014 트렁크 고정, 저장소 묶음을 반영". 운영자·사용자 결정이 아니다.
- 기준: `WT-agentic-factory-integration` HEAD `12eaa6c23f`. 아래 모든 코드 사실은 이 작성자가 그 트리에서 읽은 파일과 줄이다.
- 감사 상태: 직전 코드 감사는 46dbcce0d4에 대한 FAIL 0.76(receipt `rcpt-08a37bf592a0ccc3e932a935`)이며 그 값은 리더가 전한 것이다 — 이 작성자는 그 verdict를 읽지 않았고, `.moai/reports/issue-1948`에서 receipt 문자열을 grep했으나 아무것도 찍히지 않았다(종료 코드 관측 못 함). 이 판(0.6.0)에 대한 감사는 돌리지 않았다. PASS라고 적지 않는다. SPEC 상태는 `draft` 그대로다.

### 대응표 (문장 → 코드 → REQ/AC)

| 기록한 동작 | 착지 코드(file:line) | REQ | AC(0.6.0 사례) |
|---|---|---|---|
| 원장은 저장소 정체를 한 번 묶고 바꾸지 않음, 다른 정체는 `ErrWrongRepository` | `ledger/repository.go:13`, `:23-40`, `:42-47`; `ledger/schema.go:71-77` | 001 | 001 (e1) |
| 동사가 저장소를 묶거나 비교, 묶이기 전 아는 저장소와 비교, `--repo`는 같은 git common dir만 | `cli/af.go:72`, `:85-107`, `:113-138`, `:161-175`, `:245-268`, `:284-302`; `cli/af_land.go:292-316` | 001 | 001 (e2)(e3)(e4) |
| git 저장소가 아닌 디렉터리에서 git을 부르지 않는 동사는 묶지 않음 | `cli/af.go:86-90` | 001(예외) | 001 줄의 착지 시험 `af_repo_n38_test.go:321` |
| reclaim-cap park를 나가면 `reclaim_count` 0 | `ledger/move.go:126-130` | 007 | 007 (iv) |
| `moai af resume` | `cli/af.go:446-506`(`:468-470`, `:471-473`, `:483-485`, `:486-490`); `core/move.go:24` | 007 | 007 (iv)(v) |
| Enqueue의 최신 증거·실패한 최신 실행·재제출 거절, 이전 epoch 착지 행의 현재 epoch 측정 | `ledger/queue.go:441-459`, `:460-466`, `:467-483`, `:379-390`, `:223-236`, `:283-302`, `:207-221` | 008 | 008 (iv)~(vii) |
| `transitions.evidence_mark` 열과 쓰기 | `ledger/schema.go:41-49`; `ledger/claim.go:228-238` | 009 | 009 (b) |
| 스키마 판 2, 더 새 판 거절, 판 1 이전(같은 밀리초 규칙) | `ledger/schema.go:17-19`, `:123-125`, `:137-153`, `:155-177`(`:173`) | 009 | 009 (b)(c)(d) |
| trunk를 동사에서 고정, 다른 `--trunk` 거절 | `cli/af.go:229-238`; `cli/af_land.go:45`, `:294`; `mergeq/mergeq.go:40`, `:168` | 014 (i) | 014 (e) |
| 실패 기록 `FailureLog`·`gate-logs` | `mergeq/faillog.go:19`, `:32-38`, `:50`, `:53`, `:56`, `:59`, `:69-74`, `:103-164`, `:171-190`, `:195`, `:202-262`, `:272-279`, `:284-302`; `faillog_acl_darwin.go:14-19`; `faillog_acl_other.go:8`; `mergeq.go:90-101`, `:127-144`, `:204-214`, `:283`, `:293`, `:297`; `gitcli.go:73-83`, `:118-128`; `cli/af_land.go:58-61`, `:85-113`, `:119-122`, `:226` | 014 (ii) | 014 (f) |
| merged 사실: 가장 새 작업 + 착지가 서 있음 | `ledger/queue.go:31-84`, `:86-111`, `:113-155`, `:157-205`, `:241-243`, `:245-281` | 014 (iii) | 014 (g) |
| 이전 epoch 착지의 이어받기와 `OnTrunk`, `enqueue`의 trunk 질문 | `ledger/queue.go:319-325`, `:379-405`, `:484-488`; `cli/af.go:569-583`, `:179-185` | 014 (iv) | 014 (h) |
| DONE 이동의 trunk 확인 | `cli/af.go:420-424`, `:187-211` | 014 (v) | 014 (i) |

### 새 id 없음 (결정과 까닭)

- 리더는 새 id 하나까지 허락했으나 쓰지 않았다: 저장소 묶음은 REQ-AFL-001(단일 로컬 권위)에, resume은 REQ-AFL-007(재회수 상한)에 붙였고 AC는 각 AC의 "0.6.0 사례" 줄로 더했다. REQ 21·AC 21 그대로다. 그래서 release 게이트 AC 셋(013·015·017)과 §D 완료 계산은 바뀌지 않는다.
- 표가 아홉이 된 것(REQ-AFL-009)과 merged 사실의 조건이 좁아진 것(REQ-AFL-014 (iii))은 리더가 그 기록을 지시했으므로 지시된 개정으로 다뤘다 — 코드와 REQ의 충돌로 보고하지 않았다. 이전 문장은 spec §C.6에 남겼다.

### OD-35 (l) 닫힘

- spec §G의 OD-35 행에서 (l)을 닫았다: trunk는 착지 코드 @`12eaa6c23f`에서 동사가 `refs/heads/af/trunk`로 고정한다. 닫은 근거는 착지 코드와 리더 지시(N19)이며 운영자 결정이 아니다. 상위 SPEC(`SPEC-AGENTIC-FACTORY-001`)의 OD-35 행은 고치지 않았다.

### 손대지 않은 파일

- `plan.md`, `research.md`(리더가 준 파일 목록 밖), 상위 SPEC 디렉터리, 모든 코드.

### 확인하지 않아 적지 않은 문장 (left out as unverified)

- FAIL 0.76 verdict와 receipt의 내용 — 읽지 않음.
- 각 착지 시험의 본문이 0.6.0 사례를 얼마나 덮는지 — 본문을 읽은 넷(`af_land_repair_test.go:188`, `af_trunk_n39_test.go:130`, `:171`, `move_test.go:178`의 `:212` 단언) 말고는 이름과 줄만 확인. 어떤 시험도 돌리지 않음.
- 계약 §10 20th lane-22 (c)("the trunk also may not equal the target of `refs/remotes/origin/HEAD`")를 코드가 검사하는지 — 확인하지 않음; SPEC에 적지 않음.
- 원장 경로 `<moaiHome>/db/<projectKey>/af/ledger.db`가 착지 코드의 경로 계산과 같은지 — 확인하지 않음; REQ-AFL-001의 그 구절은 0.5.1 그대로 둠.
- 실패 출력을 담는 tail writer의 보존 크기(`afGateKeep`)와 git 출력 상한(`maxGitOutput`) — 읽지 않음; SPEC에 수치를 적지 않음.
- `resume`의 다른 epoch 거절이 원장 층에서 나는지 — 확인하지 않음(동사 도움말과 시험 이름만).
- 저장소가 답하지 못할 때의 DONE 거절을 단언하는 시험 — 찾지 못함(AC-AFL-014 (i)에 no test found).
- 훅 heartbeat 기록기에 저장소 확인이 없다는 것 — `grep -n -e repositor -e Repo internal/hook/ledger_heartbeat.go`가 아무것도 찍지 않은 것에 기댐; 종료 코드 관측 못 함.
- 아래 열린 부채 id의 내용 — 리더가 준 목록 그대로이며 이 작성자는 그 항목을 읽지 않았다.

### 열린 부채와 잔여 위험 (residual risk)

- 리더가 넘긴 열린 감사 항목: N4, N5, N6, N7, N8, N10, N20, F8, F9, F12, E6, N37, N44(나머지), N46, N47, N48. 이 판은 그 어느 것도 닫지 않는다.
- 원장이 아무것도 모를 때(묶이지 않음, lease worktree 없음, 행의 커밋 없음) 기본 위치의 원장은 첫 호출자가 묶음을 정한다 — 틀린 저장소에서 처음 부르면 그 저장소에 묶인다. `--db`로 준 원장은 파일 자리가 저장소 밖이면 `--bind-repo`를 요구한다(2026-11-01 코드 정정, 아래 정정 줄).
- 저장소를 디스크에서 옮기면 정체(symlink를 푼 절대 git common dir)가 바뀌어 `moai af move-repository`를 부르기 전까지 뒤의 모든 동사가 거절된다; 첫 묶음 행은 갱신·삭제가 트리거로 막혀 있고, 복구는 별도의 추가 전용 이동 기록과 그 동사가 한다(2026-11-01 코드 사실, 아래 정정 줄); 그 복구 경로는 2026-11-01 셋째 정정 줄로 REQ-AFL-001에 적었고, 그것을 단언하는 AC는 아직 없다.
- `internal/hook/ledger_heartbeat.go`는 저장소를 확인하지 않는다(위 grep).
- 0.6.0 사례 줄은 §N(release 게이트 밖)에 있어, 그 시험이 깨져도 release 판정은 바뀌지 않는다.
- 정정 줄 (2026-11-01, 코드 사실만; 새 REQ·AC id 없음, 상태 변경 없음, PASS 주장 아님): 감사 `1c05083f71`의 N49·N50·N54와 N51 (b)(c)(d)·복제본 형태가 코드에서 고쳐졌고, 그 때문에 거짓이 된 REQ-AFL-001의 원장 파일 비교 문장, acceptance (e3)와 §O 첫 줄, design의 동사 쪽 문단을 같은 사실로 고쳤다. 고친 항목은 `.moai/reports/issue-1948/af-audit-46db-repair.md` 후속 절에 있다. N52(저장소 이동 복구 동사)는 시작하지 않았다.
- 정정 줄 (2026-11-01, 코드 사실만; 새 REQ·AC id 없음, 상태 변경 없음, PASS 주장 아님, 위험 수용 기록 아님): 위 줄의 "N52는 시작하지 않았다"는 그 시점의 사실이며, 지금은 코드에 있다 — 원장의 `repository_move` 표와 `MoveRepository`·`RepositoryMoves`, 동사 `moai af move-repository`. 그 때문에 거짓이 된 문장(REQ-AFL-001의 "다른 정체"·"모든 동사"·인용 줄 번호, REQ-AFL-009의 표 목록, design 부록 D의 "afRun이 모든 동사에서 묶는다"와 "옮기면 뒤의 호출이 거절된다", §O의 "복구 경로가 REQ에 없다")을 같은 사실로 고쳤다. 고친 항목은 `.moai/reports/issue-1948/af-audit-46db-repair.md` 후속 4절에 있다.
- 정정 줄 (2026-11-01 셋째, 코드 사실만; 새 REQ·AC id 없음, 상태 변경 없음(draft), 감사 없음, PASS 주장 아님, 위험 수용 기록 아님): 감사 `a36ea8c368`(FAIL 0.73) 뒤 수리 커밋 `e544228237`..`0f108104c1`(표지·`--from`·N62·N65·N68, N63·N66·N69·N72·N73 도움말, E6 대조 시험; `.moai/reports/issue-1948/af-audit-a36e-repair.md`) 때문에 거짓이거나 모자라게 된 문장만 고쳤다. spec.md: REQ-AFL-001(기본 위치와 `<projectKey>` = `homestate.ProjectDir`의 `<project name>-<hash>`, `--from`으로 한 파일 유지, 표지 기록·비교·채택, lease worktree는 소유 증거가 아님, git이 답하지 못하거나 시간이 다하거나 모양이 틀린 답은 거절, `--db` 원장의 `--bind-repo` 조건, `move-repository`의 표지 규칙·표지 없는 원장의 대체 규칙(밝힘이지 증명이 아님)·첫 이동의 `user_version` 3, 인용 줄 번호), REQ-AFL-009(`repository_marker` 표, 판 3, schema.go 인용 줄 번호), REQ-AFL-014((v) 뒤 E6 보증 문장, (vi) N54 착지 거절·잠금 안 swap 직전 재확인·lease 없는 출구·남는 창, (v) 인용 줄 번호). design.md: §A 이전 문단의 "판(2)보다 크면 거절", 부록 D 동사 쪽 문단·저장소 이동 문단·열린 위험. acceptance.md: §E의 0.6.0 REQ-AFL-001 줄. progress.md: 위 §O 둘째 줄의 "REQ·AC는 아직 없다". 이전 문구는 이 줄과 git 기록(`git show 0f108104c1:<파일>`)에 있고 §C.6에 따로 옮기지 않았다. 위 "확인하지 않아 적지 않은 문장"의 원장 경로 항목은 이번에 `internal/cli/af.go:57-66`(`homestate.ProjectDir` + `af/ledger.db`)로 읽었다. 남긴 것: `moai af add`·`move-repository`·N54 규칙의 AC, REQ-AFL-014 `release` after `land` 문구, 연결되지 않은 importer(REQ-AFL-016), 이 정정이 고치지 않은 문장의 낡은 `af.go` 줄 인용(예: design 부록 C, 부록 D의 `schema.go:78-87`·`af.go:245-268`·`:86-90`, acceptance AC-AFL-014 (i)의 `af.go:187-211`). 새 남는 위험: `--from` 전에 새 자리에서 동사를 부르면 새 키에 빈 원장이 생겨 `--from`이 거절되므로 그 동안 프로젝트에 원장 파일이 둘이다(동사는 그 경우를 말해 주며, 운영자가 빈 원장을 치워야 한다; `af_move_repository.go:111-113`, `:192-204`); `--from` 뒤 heartbeat 원장 경로는 옛 경로다(수리 보고 Gaps). 시험·빌드는 하나도 돌리지 않았다.

### 측정 명령 (이번 작업, 이 트리)

- 시험 이름 확인: `grep -n '^func Test' <파일>`을 이 작업이 인용한 시험 파일마다 돌림. 덧붙여 acceptance.md가 인용한 모든 `Test…` 이름(129개)을 `internal/agenticfactory`·`internal/cli`의 `^func Test…` 선언과 `comm -23`으로 비교 — 선언에 없는 이름은 25개였고, 모두 0.6.0 이전 판의 "unwritten" 이름이거나 다른 패키지(`internal/web`, `internal/hook`)의 이름이거나 산문 토큰(`TestName`, `TestAssertion`)이다; 0.6.0 사례 줄의 이름은 모두 선언에 있다.
- receipt grep: `grep -rn rcpt-08a37bf592a0ccc3e932a935 .moai/reports/issue-1948` → 출력 없음(종료 코드 관측 못 함).
- heartbeat grep: 위와 같음.
- spec lint(편집 뒤, 커밋 전 작업 트리): `moai spec lint --strict <worktree>/.moai/specs/SPEC-AGENTIC-FACTORY-LOCAL-001/spec.md` → `exit=0`, `0 error(s), 1 warning(s)`(`REQTableRowsRejected`, spec.md 150행 — §M의 0.5.1 측정과 같은 경고 종류) 그리고 INFO `OwnershipTransitionUnmeasured` 하나. 판정한 빌드: 설치된 `moai` v3.2.0-rc.38, 커밋 `b1806498fe`(`moai version`) — 이 트리(HEAD `12eaa6c23f`)에서 만든 빌드가 아니다(`verification-claim-integrity.md` §2.2). 그래서 이 줄은 그 빌드의 lint 규칙에 대한 관측이며, 이 트리의 lint 규칙으로 잰 것이 아니다.

