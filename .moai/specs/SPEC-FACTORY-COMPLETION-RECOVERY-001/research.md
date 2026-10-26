# research.md — SPEC-FACTORY-COMPLETION-RECOVERY-001

이 문서는 v0.2.0에서 추가된 M2-B(run 수명주기, 카드 t1963)의 코드 조사 기록이다. M1·M2(원래 범위)·M3의 조사는 spec.md §B.2에 있다.

## §1 — 범위와 방법

- 측정 트리: HEAD `c4784db4c2`(브랜치 `WT-run-dead-owner`, 워크트리 `.moai/worktrees/t1963`). 이 문서의 모든 행 번호는 이 트리 기준이다.
- 입력: 리더의 읽기 전용 분석 `.moai/reports/t1963/zombie-analysis.md`. git 미추적 카드 증거이므로 권위가 아니며, 아래 §2에서 인용 하나하나를 소스로 다시 읽었다.
- 읽은 파일: `internal/homestate/{factory_run_retire,factory_run_adopt,factory,card_picked,card_metadata,card_worktree,runtime,factory_run_cards,leader_approval}.go`, `internal/cli/{factory_run_select,factory_card,factory_handoff_recover,factory_run_adopt,factory,factory_relaunch}.go`, `internal/factorymsg/store.go`, 그리고 관련 시험 파일.
- 라이브 factory DB는 읽지 않았다. 분석이 인용한 운영자 화면의 run 개수(5개)와 카드 수는 재측정하지 않았다.

## §2 — 분석 인용 검증표

판정: **일치** = 인용 위치에 주장한 코드가 있다 · **근접** = 주장은 맞고 행이 몇 줄 어긋난다 · **낡음** = 주장은 맞지만 위치가 크게 바뀌었다 · **오류** = 주장 또는 파일이 틀렸다.

| # | 분석의 인용 | HEAD `c4784db4c2` 실제 위치 | 판정 |
|---|---|---|---|
| 1 | `internal/cli/factory_run_adopt.go:34,173` — 리더 시작마다 새 run id | `:34` `leaderRunIDWasMinted`, `:173` `recordLeaderRunStart` | 일치 |
| 2 | `internal/cli/factory.go:707-727` | `:707-727` `recordFactoryRunStart` | 일치 |
| 3 | `internal/homestate/runtime.go:44-50` — `lead_pid`·`lead_process_start` 기록 | `:44-47` `INSERT INTO runs(...)` | 일치 |
| 4 | `factory_run_select.go:335` — 리더 이름을 pid로 표시 시점에 해석 | `:331` `leaderNamesByPID(root)` | 근접 |
| 5 | `factory_run_retire.go:126,139` `ClassifyOwnerWith` | `:126` `DefaultOwnerClassifier`, `:139-155` `ClassifyOwnerWith` | 일치 |
| 6 | `factory_run_retire.go:300-332` — 피어·부팅 증명 폴백 | 피어 폴백 `:214-219`(`classifyRuns` 안), 부팅 증명 `predatesBoot` `:290-323`, 활동 질의 `:327-334` | 낡음 — 피어 폴백은 그 범위 밖이다 |
| 7 | `internal/cli/factory_handoff_recover.go:117-166` `--retire` | `:117` `newFactoryRunsCommand`, `:134` `RetireRunIfDead` 호출, `:165` 플래그 | 일치 |
| 8 | `RetireRunIfDead` `:381`, `retireRun` `:422` | `:381`, `:422` | 일치 |
| 9 | `factory_run_select.go:338` — 합류 프롬프트 `r<n>` | `:338`은 `retireChosenRun`. `r` 파싱은 `:202-224`(`:207`), 제시 조건은 `:86` | 근접 — 분석의 `:87`도 한 줄 어긋남 |
| 10 | `:443-449` 비종결 카드 거부 | `:438-444`(`inProgressCards` 호출 `:438`, 오류 반환 `:443`) | 근접 |
| 11 | `factory_run_adopt.go:20` — 비종결 정의 | `:21` `sqlTerminalStates = ('done','failed','abandoned')` | 근접 |
| 12 | `internal/factorymsg/store.go:392,432` | `:392` `ResolveActiveRun`, `:432` `ReconcileActiveRuns` 호출 | 일치 |
| 13 | `runHasLiveRecords` `factory_run_adopt.go:31-39` | `:31-39` | 일치 |
| 14 | `UPDATE runs SET status='retired'` `:454,:488` | `:445`(`retireRun`), `:485`(`retireRunUnlessLive`) | 낡음 |
| 15 | 초안 spec.md:39 — 은퇴는 runs만 갱신 | 주장은 맞다. 다만 그 행이 인용한 `factory_run_retire.go:393`은 HEAD에서 `:445`다(§2.3) | 근접 |
| 16 | `factory.go:51,83` `PRIMARY KEY(run_id, card_id)` | `:51` `CREATE TABLE cards`, `:83` 키 | 일치 |
| 17 | `card_picked.go:130,166` | `:130` `RecordPicked`, `:166` INSERT | 일치 |
| 18 | `card_metadata.go:284` `PreviousCardRow`; `factory_card.go:1448`; `card_worktree.go:216` | `:284`; `:1448` `factoryPriorCarryCandidate`(호출 `:1449`); `:211` `PreviousCardWorktree`(질의 `:214-218`) | 일치 |
| 19 | `FindResumableRun` `:72-96`, "둘 이상이면 경고 후 새 run 발급(`:76-80`)" | 함수 범위는 일치. 그러나 `:76-80`은 후보 수집 루프이고, 경고·발급은 `internal/cli/factory_run_adopt.go:77-81`에서 한다. 분석이 빠뜨린 조건: 후보가 하나여도 `OwnerIndeterminate`나 `BasisPeer`면 채택하지 않는다(`:92`), 그리고 명시적 resume 인자가 필요하다(cli `:165`) | 오류(파일) + 누락 |
| 20 | `AmbiguousRunsError` `store.go:437-458` | 반환 `:442`, 타입 `:453` | 근접 |
| 21 | CLI 센티널 `internal/cli/factory.go:517` | `:517` `ambiguousFactoryLeaderSentinel = "AMBIGUOUS_FACTORY_LEADER"` — 위치는 맞지만 이것은 산 리더 간 모호성용 별개 센티널이다. `AMBIGUOUS_FACTORY:` 접두는 `internal/factorymsg/store.go:458`이 만든다(plan-audit iter1 D11b) | 일치(위치) + 의미 정정 |
| 22 | "리더 종료 시 은퇴하는 경로 없음" | 은퇴 호출 지점을 전부 찾았다: `store.go:432`, `factory_run_select.go:344`, `factory_handoff_recover.go:134`, **`internal/cli/factory_relaunch.go:106`**(`moai factory relaunch --from-run`, REQ-SRH-013 — 분석이 빠뜨린 세 번째 운영자 은퇴 표면, 같은 가드). 종료 훅 경로는 없다 | 일치 + 누락 1건 |

### §2.1 — 분석이 다루지 않은 사실 (설계 입력)

- `workers`의 기본 키는 `label`이고 `run_id`는 일반 열이다(`internal/homestate/factory.go:30-38`). 프로세스 시작 지문 열이 없다 — worker 생존 판정(REQ-FCR-022)은 pid만으로 할 수밖에 없다. 레인이 다시 등록하면 같은 label 행이 지워지고 새 run id로 다시 들어간다(`internal/factory/factory_slots.go:346`, `:355`).
- 카드↔run 바인딩 테이블 `card_dispatch(card_id PK, run_id)`가 있다(`factory.go:161-165`, 쓰기 `internal/homestate/leader_approval.go:291`, 읽기 `:251`, `internal/factory/gtd_operation.go:376`, `:526`). M1의 완료 게이트가 이 바인딩을 읽으므로, 카드를 다른 run으로 옮기면서 바인딩을 그대로 두면 완료가 막힌다. 통합 연산(REQ-FCR-020)은 같은 트랜잭션에서 바인딩을 옮긴다.
- 리더 승인은 `leader_approvals(card_uuid, run_id)`로 run에 묶인다(`factory.go:150-160`). REQ-FCR-002는 다른 run의 승인을 거부하므로, 옮긴 카드는 대상 run에서 승인을 새로 받아야 한다(REQ-FCR-023 — 승인 행을 옮기지 않는다).
- 큐 쪽 "현재 배차" 기록은 다른 저장소(backlog)에 있어 factory 트랜잭션에 넣을 수 없다. 바인딩만 옮기고 큐 기록이 옛 run을 가리키면 완료 게이트는 **닫힌 쪽**으로 실패한다(M1 round-25 기록의 바인딩/기록 정합 수리). 이것은 잔여 위험으로 남긴다(design.md §5).
- 기존 시험 중 M2-B가 반전하는 것: `TestAdoptResumedLeaderRunWarnsOnMultipleCandidates`(`internal/cli/factory_run_adopt_test.go:134`). Q1 판정 (a)로 허용 집합만 넓어지는 것: `TestFactoryRunsHasNoForceFlag`(`internal/cli/factory_runs_retire_guard_test.go:234-242`, 플래그 집합 정확 일치 단언 — `[retire]` → `[rebind-to retire]`).
- `TestReconcilePreservesDeadOwnerWithOnlyWorkers`(`internal/homestate/factory_run_adopt_test.go:256`)와 `TestRetireRunIfDeadRetiresWhenOnlyTerminalCardsOrWorkers`는 worker pid로 `1`을 심는다(`mustInsertWorker`, `:20-26`). **OS별 정정(plan-audit iter1 D5)**: Unix 계열에서 pid 1은 init/launchd라 살아 있거나 판정 불가로 나올 것으로 추론되지만, Windows 탐침은 `ERROR_INVALID_PARAMETER`를 Dead로 읽고(`internal/homestate/pid_state_windows.go:11-14`) pid 1은 Windows의 유효한 프로세스 id가 아니다 — REQ-FCR-022 하에서 `TestReconcilePreservesDeadOwnerWithOnlyWorkers`가 windows 러너에서 run을 은퇴시켜 실패할 수 있다(코드 판독 추론, 어느 OS에서도 측정 안 함). 그래서 `mustInsertWorker`의 pid를 `os.Getpid()`로 바꾸는 픽스처 수정 1건을 허용한다(AC-FCR-025). `TestRetireRunIfDeadRetiresWhenOnlyTerminalCardsOrWorkers`는 worker 생존을 보지 않는 `retireRun` 경로라 영향이 없다.

- CAS 패배의 현행 동작은 런치 실패다: `internal/cli/factory_run_adopt.go:106-108`이 `AdoptRun` 오류를 반환하고 `internal/cli/cc.go:227-229`·`internal/cli/glm.go:283-285`가 `re-adopt factory run: %w`로 실패한다(plan-audit iter1 D3, 이 개정에서 직접 읽어 확인). REQ-FCR-018의 패자 동작(새 id 발급 + 경고)은 새 동작이다.
- 상태 연령 표시는 이벤트를 `run_id`로 거른다(`internal/homestate/card_age.go:75-86`) — 재키된 카드의 대상 run 기준 연령은 "unknown"으로 읽힌다(design.md §5 잔여 위험).

### §2.2 — 수정 대상 아님으로 확인한 것

- `AdoptRun`(`internal/homestate/factory_run_adopt.go:112-145`)은 이미 스탬프 compare-and-set이다(`:128` WHERE에 `lead_pid=? AND lead_process_start=?`). REQ-FCR-017은 이것을 그대로 쓴다.
- `retireRun`은 이미 BEGIN IMMEDIATE 한 트랜잭션에서 상태→스탬프→카드→전환→이벤트 순서를 지킨다(`:408-462`). 통합 연산은 같은 형태를 따른다.

### §2.3 — 기존 M2 요구의 앵커가 HEAD에서 이동함

M1 착지(round-1~25)로 `card_transition.go` 등이 크게 바뀌어, spec.md §B.2(base `a158b4b5f`)가 적은 M2 앵커 일부가 HEAD에서 다른 행에 있다. 요구 문장은 바꾸지 않고 위치만 여기 기록한다 — run-phase 첫 커밋 전 재확인 항목(plan.md §H)의 입력이다.

| spec.md의 앵커 | HEAD 위치 |
|---|---|
| `factory_run_retire.go:161-163` `ClassifyRuns` | `:168-170` |
| `factory_run_retire.go:333` active-only 분류 | `:340`(`ReconcileActiveRuns`), `:73`(`FindResumableRun`) |
| `factory_run_retire.go:393` UPDATE-only | `:445`, `:485` |
| `factory_run_retire.go:158` `retirable` | `:165` |
| `card_transition.go:267` 예약 edge 거부 | 판정 `:207` `isReservedEdge`, 호출 `:333` |
| `card_transition.go:351-353` `applyLeaseExpiry` | `:422` |
| `card_transition.go:451` 동일 owner 임대 | `:588` |
| `card_transition.go:364` `updateCardRow` | `:875` |
| `card_record.go:137-138` `LeaseExpired` | `:147-148` |
| `card_evidence_readers.go:93-207` | `:93`, `:119`, `:176` — 일치 |

## §3 — 결함 상태 관측 (RED)

acceptance.md §0.1 증거 장부 L1·L7이 정본이다. L7(트리 `41f8db15de`, exit 1): 죽은 스탬프 run 옆에 `OwnerLive` 활성 run이 있어도 후보 탐색이 `adopt="run-dead"`를 반환한다 — 산 리더 옆 채택이 오늘 가능하다. L1·L7은 결함 관측이며, 바뀌지 않는 함수를 단언하므로 착지 뒤에도 RED로 남을 수 있다(전환 판정은 각 AC의 green 명령). 요약: `.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/probes/`의 오버레이 시험 세 개가 HEAD `c4784db4c2`에서 모두 FAIL(exit 1)했다 — 죽은 run 2개 fixture에서 채택 없음·활성 run 2개, 죽은 pid worker만 가진 run이 active로 남음, 새 run 임대 뒤 옛 run 행이 `run` 상태로 남음.

## §4 — Gaps

- 운영자 화면의 run 5개·카드 수(129·33·3·33·12)는 분석의 인용일 뿐, 재측정하지 않았다.
- 설치된 rc 바이너리와 HEAD의 차이는 보지 않았다(이 카드는 바이너리를 건드리지 않는다).
- `ProbeProcessIdentity`가 pid 1에 대해 `Live`와 `Indeterminate` 중 무엇을 내는지 직접 관측하지 않았다 — Unix 계열 한정으로, 어느 쪽이든 REQ-FCR-022에서 "산 worker"이므로 결론은 같다. Windows는 §2.1 정정 참조(pid 1이 dead로 읽힐 수 있다).
- 큐(backlog) 저장소 쪽 배차 기록의 정확한 갱신 경로는 M1 코드로 확인하지 않았다(round-25 기록을 근거로 한 서술).
