# design.md — SPEC-FACTORY-COMPLETION-RECOVERY-001

M1·M2(원래 범위)·M3의 설계는 plan.md §F의 파일별 변경표가 담는다. 이 문서는 v0.2.0에서 추가된 **M2-B — run 수명주기(죽은 소유자 run 통합)**의 관측 가능한 동작과 경계를 정한다. 함수 이름·시그니처는 run-phase가 정한다. 여기 나오는 기존 식별자는 위치 참조용이다.

## §1 — 상태 모델

run 하나의 수명은 다음 셋 중 하나다.

| run 상태 | 소유자 판정 | M2-B 이후 처리 |
|---|---|---|
| active | `OwnerLive` | 건드리지 않는다. 통합의 대상이 될 수 있다 |
| active | `OwnerDead` + `BasisStamp` | 리더 진입에서 채택(가장 최신 하나) 또는 통합의 원천(나머지) |
| active | `OwnerIndeterminate`·`BasisPeer`·`BasisBoot` | 자동 처리 없음. 경고와 운영자 경로만 |
| retired | 무관 | 비종결 행이 남아 있으면 임대 시 정리(REQ-FCR-019) |

카드 행 하나의 이동은 두 가지뿐이다.

- **재키(re-key)** — 대상 run에 같은 `card_id` 행이 없을 때. `run_id`만 바뀌고 state·version·stage·lease·evidence 등 나머지 열은 그대로다. 행 수 불변.
- **포인터 포기(abandon-with-pointer)** — 대상 run에 이미 같은 `card_id` 행이 있을 때, 그리고 임대 경로(REQ-FCR-019). 원천 행은 지우지 않고 `abandoned`가 되며, `card.rebound` 이벤트 하나가 `from_run`·`to_run`·이전 state·version·stage·evidence_sha를 담는다. 행 수 불변, 이전 상태는 이벤트로 복원할 수 있다. 충돌 시 어느 행을 남길지는 decision-index Q3.

## §2 — 리더 진입 (REQ-FCR-017, 018)

```mermaid
flowchart TD
  A[리더 진입: 새 id 발급, --factory-run 없음] --> B{OwnerLive 또는 OwnerIndeterminate인 활성 run이 있나}
  B -- 있음 --> Z[변경 전 동작: 발급 + 후보 경고]
  B -- 없음 --> C{죽은 스탬프 run 중 live records 있는 것}
  C -- 0개 --> Z
  C -- 1개 이상 --> G{최신 후보가 codex 리더 run이거나 산 레거시 피어를 가졌나}
  G -- 그렇다 --> Z
  G -- 아니다 --> D[최신 후보를 AdoptRun CAS로 채택]
  D -- CAS 패배 --> Y[오류 없이 새 id 발급 + 경합 run id 경고 한 줄, 통합 0]
  D -- 성공 --> E[나머지 죽은 스탬프 run 중 산 worker 없음·codex 아님·산 레거시 피어 없음인 것마다 통합 연산 1회]
  E --> F[제외·거부된 원천은 활성 유지 + stderr 한 줄, 통합된 원천은 옮긴·포기한 카드 id를 stderr에]
```

- 미확정·피어 판정 후보는 채택 후보에서 빠지고 기존 경고 줄을 받는다. 미확정 활성 run이 **하나라도 있으면** 채택 자체를 하지 않는다 — 배치 탐침 누락 시 산 리더가 미확정으로 분류되므로(`internal/homestate/factory_run_retire.go:251-256`, REQ-003b), 그 옆에서 채택·통합하면 산 리더의 카드를 옮길 수 있다.
- codex 리더 run·산 레거시 피어 run 거부(`internal/cli/factory_run_adopt.go:87-93`·`:100-105`)는 채택 대상에 적용되고, 같은 두 조건은 통합 원천에서도 제외 사유다. 최신 후보가 이 거부에 걸리면 **채택 0** — 더 오래된 후보로 넘어가지 않는다(어느 run을 이어받을지가 실행마다 달라지는 것을 막는다).
- CAS 패배는 현행 코드에서 런치 실패다(`internal/cli/factory_run_adopt.go:106-108` → `internal/cli/cc.go:227-229`·`internal/cli/glm.go:283-285`). REQ-FCR-018은 이를 "오류 없이 새 id 발급 + 경고 한 줄 + 통합 0"으로 바꾼다 — 패배를 오류와 구별하는 판별(스탬프 불일치에 의한 거부인지)은 run-phase가 정한다.
- 기존 `FindResumableRun`의 "후보 정확히 하나" 계약과 그 시험은 그대로 두고, 최신 선택은 새 선택 단계가 맡는 것을 권장한다 — `TestFindResumableRun*`을 바꾸지 않기 위해서다. 반전되는 시험은 `TestAdoptResumedLeaderRunWarnsOnMultipleCandidates` 하나다.
- `strictResumeCarrierSeen`은 이제 채택 게이트가 아니다. 다른 호출자가 없어지면 미사용 경고가 나므로, 제거할지 남길지(예: 경고 문구 분기)는 run-phase가 lint 결과로 정한다.

## §3 — 통합 연산 (REQ-FCR-020)

한 번의 BEGIN IMMEDIATE 트랜잭션. 순서는 `retireRun`(`internal/homestate/factory_run_retire.go:422-462`)의 "상태→스탬프→카드→전환→이벤트"를 넓힌 것이다.

0. 원천 소유자가 분류 시점에 `OwnerDead`인지 확인(리더 진입 경로는 추가로 `BasisStamp`). 아니면 거부 — `--rebind-to`·`m<n>`·리더 진입 세 표면 공통. 이 단계가 없으면 `--retire <산 run> --rebind-to <산 run2>`가 산 리더의 카드를 옮기는 우회로가 된다.
1. 원천 run을 다시 읽어 `active`이고 분류 당시 스탬프 그대로인지 확인. 아니면 거부.
2. 대상 run이 `active`이고 분류 당시 스탬프 그대로인지 확인(대상 소유자의 `OwnerLive` 판정은 트랜잭션 전 분류에서 하고, 트랜잭션 안에서는 스탬프 동일성으로 확인). 아니면 거부.
3. 원천에 산 worker(REQ-FCR-022)가 있으면 거부.
4. 원천의 비종결 카드 전부에 대해 제3의 활성 run(소유자가 `OwnerDead`가 아닌)에 같은 `card_id` 비종결 행이 있는지 확인. 하나라도 있으면 거부.
5. 카드마다 재키 또는 포인터 포기.
6. 원천을 가리키는 `card_dispatch` 행을 대상으로 옮김.
7. 원천을 `retired`로, `run.retired` 이벤트에 `classification`·`basis`·`rebound_to`·`moved`·`abandoned` 수와 **옮긴 카드 id 목록·포기한 카드 id 목록**을 담음. 재키된 카드는 카드별 이벤트를 따로 남기지 않으므로, "어떤 카드가 옮겨졌는가"의 기록은 이 페이로드와 리더 진입·CLI의 stderr 출력이다.

어느 단계의 거부든, 어느 단계 뒤의 실패든 롤백한다 — 부분 이동은 존재하지 않는다. 카드 이동(5)·바인딩 재지정(6)·은퇴와 이벤트(7)는 한 커밋이며 둘로 나누지 않는다(AC-FCR-020 주입점 (b)가 이 분할을 잡는다). 중단(프로세스 종료)도 커밋 전이면 롤백이다.

### §3.1 — 경합

| 경합 | 결과 |
|---|---|
| 두 리더가 같은 죽은 run을 동시에 채택 | `AdoptRun`의 스탬프 CAS가 하나만 통과. 진 쪽은 오류 없이 새 id를 발급하고 경고 한 줄을 남기며 통합하지 않는다(현행 런치 실패를 대체하는 새 동작, AC-FCR-019 (a)) |
| 두 리더가 같은 원천을 동시에 통합 | 1단계 스탬프 확인 + BEGIN IMMEDIATE 직렬화. 먼저 커밋한 쪽만 성공, 다른 쪽은 원천이 이미 `retired`라 거부 |
| 통합 도중 레인이 원천 카드에 전이를 커밋하려 함 | BEGIN IMMEDIATE가 쓰기를 직렬화. 통합 뒤에 오면 그 카드는 원천 run에 없으니 레인 전이가 실패(카드 없음) — 레인은 재합류해야 한다. 산 worker가 있는 원천은 3단계에서 아예 거부되므로 정상 운영 중 레인이 붙은 run은 통합되지 않는다 |
| 통합 도중 다른 run이 같은 카드를 임대 | 4단계가 트랜잭션 안에서 읽으므로 임대가 먼저 커밋되면 거부, 나중이면 임대 쪽 REQ-FCR-019가 원천 행을 본다(원천 행은 이미 대상 run에 재키됐거나 포기됨) |

## §4 — 임대 경로 (REQ-FCR-019)

`moai factory next`의 기록 트랜잭션(RecordPicked + 이어받기) 안에서, 이어받기 읽기(`factoryPriorCarryCandidate`, `internal/cli/factory_card.go:1448`) **뒤에** 다른 run의 같은 `card_id` 비종결 행을 찾아, 그 run의 소유자가 `OwnerDead`이고, **그 행의 lease가 유효하지 않으며, 그 run에 산 worker(REQ-FCR-022)가 없을 때만** 포인터 포기한다. 리더가 죽은 뒤에도 레인은 살아서 그 run의 카드를 임대 중일 수 있기 때문이다 — 그 행을 종결시키면 레인의 다음 전이가 실패하고 카드 기록이 갈라진다. 조건이 하나라도 어긋나면 기존 동작(새 행 생성, 원천 행 불변)을 유지한다. 포기와 T1 기록은 한 트랜잭션이라 포기가 실패하면 T1도 되돌려진다. 이어받기가 먼저인 이유: 포기하면 그 행이 종결 상태가 되어 이어받기 술어(`IsTerminalCardState(prev.Stage)`)의 입력이 바뀌기 때문이다. 소유자 판정은 트랜잭션 밖에서 하고, 트랜잭션 안에서는 그 run의 스탬프가 분류 당시와 같은지만 다시 본다.

## §5 — 운영자 표면 (REQ-FCR-021)

- `moai factory runs --retire <id> --rebind-to <live-run>`: 통합 연산. 출력은 옮긴 카드·포기한 카드 id 목록. `--rebind-to` 없는 `--retire`의 출력과 동작은 그대로.
- 합류 프롬프트 `m<n>`: 행이 `Dead && InProgress>0`이고 `OwnerLive` 활성 run이 정확히 하나일 때만 제시. `r<n>`은 그대로.
- 레인 세션(`MOAI_FACTORY_WORKER` 설정)의 `--rebind-to`는 상태 변경 없이 거부. `--retire` 없는 `--rebind-to` 단독 호출도 상태 변경 없이 거부.
- 원천이 `OwnerDead`가 아니면(`OwnerLive`·`OwnerIndeterminate`) 세 표면 모두 거부 — §3 단계 0.
- `moai factory relaunch --from-run`(`internal/cli/factory_relaunch.go:106`)은 바꾸지 않는다 — 은퇴만 시도하고 카드가 있으면 활성으로 남긴다.

잔여 위험:

- 큐(backlog) 저장소의 "현재 배차" 기록은 factory 트랜잭션 밖이다. 통합이 `card_dispatch`를 옮긴 뒤 큐 기록이 옛 run을 가리키면 완료 게이트는 거부(닫힌 쪽)한다. 열린 쪽 실패(잘못된 완료)가 생기지 않음은 AC-FCR-017 Then (9)가 단언한다.
- **상태 연령 저하**: 카드 전이 이벤트는 원천 `run_id`에 남는다. 상태 연령 표시는 이벤트를 `run_id`로 거르므로(`internal/homestate/card_age.go:75-86` `LoadCardEvents`), 재키된 카드의 대상 run 기준 상태 연령은 "unknown"으로 읽힌다. 기록은 사라지지 않지만(원천 run 이벤트로 남음) 대상 run 화면에서는 연령이 보이지 않는다. 이 개정은 고치지 않는다.

## §6 — worker 생존 (REQ-FCR-022)

`runHasLiveRecords`의 worker 절을 "worker 행이 있다"에서 "pid가 확실히 죽지 않은 worker 행이 있다"로 좁힌다. 판정은 pid 탐침(`ProbeProcessIdentity`)으로 하고 `ProcessIdentityDead`만 "죽음"이다. 시험이 탐침을 고정할 수 있게 `ReconcileOptions`처럼 주입 지점을 둔다(기본은 실제 탐침). SQL 하나로 끝나던 술어가 행을 읽어 탐침하는 형태로 바뀌므로, 기존처럼 호출자의 트랜잭션(`queryRower`) 안에서 돌 수 있어야 한다.

OS 차이: Windows 탐침은 `OpenProcess`의 `ERROR_INVALID_PARAMETER`를 Dead로 읽는다(`internal/homestate/pid_state_windows.go:11-14`). pid 1은 Windows의 유효한 프로세스 id가 아니므로 기존 픽스처 `mustInsertWorker`의 pid `1`은 Windows에서 dead로 읽힐 수 있다 — 그 pid를 `os.Getpid()`로 바꾸는 픽스처 수정 1건을 허용한다(AC-FCR-025, 시험 의도 불변).
