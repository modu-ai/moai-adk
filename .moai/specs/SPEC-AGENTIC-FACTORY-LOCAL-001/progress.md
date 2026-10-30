# SPEC-AGENTIC-FACTORY-LOCAL-001 — progress

status: draft

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: drafted (not audited) — 0.1.0 작성(2026-10-30, HEAD `d87913b6a3`, 코드 사실 기준 `37b67f6d39`). 감사는 coordinator가 돌린다.
- Tier L 선택: REQ 21개가 Tier M 상한 16을 넘음(spec HISTORY 0.1.0).
- 측정은 상위 progress.md `revision_0_7_0_local_ledger` 노트에 있다.
- 0.2.0 (2026-10-30, HEAD `c445ce9559` 위 작업 트리): revised, unaudited — plan-audit FAIL 0.62 수리. 감사는 coordinator가 돌린다. PASS라고 적지 않는다. 대응표는 §G.
- 0.3.0 (2026-10-30, HEAD `0c1cb1853f` 위 작업 트리): revised, unaudited — OD-36~42를 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible)대로 닫음. 있는 감사는 FAIL 0.62 @c445ce9559 하나이며 재감사는 coordinator가 돌린다. PASS라고 적지 않는다. 대응표는 §J.
- 0.4.0 (2026-10-30, HEAD `c83fdc04ff` 위 작업 트리): revised, unaudited — plan-audit FAIL 0.72 @081eacf037 수리. 마지막 감사는 그 FAIL이며 재감사는 coordinator가 돌린다. PASS라고 적지 않는다. 대응표는 §K.

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
