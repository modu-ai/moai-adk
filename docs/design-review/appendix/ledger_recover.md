# Ledger recovery (`src/ledger/recover.rs`, `src/ledger_recover.rs`)

Issue #3068 P6 (AC7, AC6 보강) · 명령 둘: `ledger-recover-plan`, `ledger-recover-apply` (조각: `docs/design-review/appendix/operations/ledger_recover_*.md`). 스키마는 P1 그대로이고 새 표·열이 없다. 라이브러리(`src/ledger/**`)는 프로세스를 띄우지 않고 네트워크·`git`·worktree 파일 시스템을 읽지 않는다. 그런 관찰은 명령 층(`src/ledger_recover.rs`)이 모아 평범한 데이터(`Observations`)로 넘긴다.

## 하는 일 / 하지 않는 일

- 열 때 `ledger_unavailable` 이 된 원장(P1)을, 읽을 수 있는 가장 좋은 사본(원장 파일 자체, `.bak`, 또는 원장 모양의 파일)에서 **새로** 만든다. 확인된 예약만 되살리고 나머지는 검토로 남긴다.
- **지우지 않는다.** 망가진 파일은 `<원장>.damaged-<UTC>` 로 남는다(하드 링크: 원래 이름이 빈 적이 없다). 옆에 남은 `<원장>-journal` 은 `<원장>-journal.damaged-<UTC>` 로 옮긴다(새 파일에 잘못 되감기지 않게). 유일한 삭제는 이 모듈이 만든 `<원장>.recovering`(과 `-journal`)의 끝나지 않은 사본뿐이다(같은 원본에서 다시 만들 수 있다).
- **덮어쓰지 않는다.** 대체할 수 있는 것은 "없음"과 **내용 때문에 쓸 수 없다고 확인된** 원장(데이터베이스가 아님·손상·잘림·다른 `application_id`·버전 0)뿐이다. 이 계획의 결과가 아닌 쓸 수 있는 원장은 `target_healthy`, 다른 사유(다른 연결이 잡고 있음·더 새 스키마·다른 프로젝트·WAL·남은 롤백 저널·읽을 수 없는 파일)는 `target_not_damaged` 로 거절한다. 링크·폴더도 거절한다.
- **한 번에 하나.** 프로젝트 폴더마다 `<프로젝트 폴더>/.recover.lock`(0600, 만들기만 하고 지우지 않는다)에 건 배타적 `flock` 이 apply 전체(만들기 → 남기기 → 바꾸기)와 `.recovering` 정리를 감싼다. 프로세스 사이에서도 통한다. 진 쪽은 `recovery_in_progress` 로 끝나고 아무것도 건드리지 않는다(진 쪽이 `.recovering` 을 지우지 않는다). 잠금은 프로세스가 죽으면 풀린다. `ledger-status` 는 이 파일을 보지 않고 `stale_temp` 에도 나오지 않는다.
- 대체 쓰기는 계획 digest + 권한 참조(`ledger-register-apply` 와 같은 문법)가 있어야 한다.
- `.creating-<pid>-<n>`(죽은 `register` 의 임시 파일)은 **보고만 한다**(`ledger-status` 와 계획의 `stale_temp`). 이름을 바꾸거나 지우지 않는다. pid 가 이 컴퓨터에서 살아 있거나 판단할 수 없으면 목록에 없다.
- 하드 링크가 안 되는 파일 시스템에서는 `register` 가 `io` 로 답한다(PR #3102 감사 메모). 복구의 `damaged` 이름 만들기도 같다.

## 라이브러리 (`moai_cli::ledger::recover`)

```rust
pub fn plan(home: &Path, project_id: &str, source: &Path, observations: &Observations)
    -> Result<RecoveryPlan, LedgerError>;                       // 읽기 전용
pub fn queries(home: &Path, project_id: &str, source: &Path)
    -> Result<Vec<ReservationRow>, LedgerError>;                // 사실을 모을 활성 예약 목록(읽기 전용)
pub fn apply(home: &Path, project_id: &str, source: &Path, observations: &Observations, plan_digest: &str)
    -> Result<RecoveryResult, LedgerError>;                     // 이름에 현재 UTC
pub fn apply_with_suffix(.., plan_digest: &str, suffix: &str) -> Result<RecoveryResult, LedgerError>;
// apply 의 작은 걸음(시험이 걸음 사이에서 죽인다):
pub fn lock_recovery(home: &Path, project_id: &str) -> Result<RecoveryLock, LedgerError>; // 기다리지 않음
pub fn build_recovering(home, project_id, source, observations, plan_digest, lock: &RecoveryLock) -> Result<Built, LedgerError>;
pub fn set_aside_damaged(home, project_id, suffix: &str, lock: &RecoveryLock) -> Result<SetAside, LedgerError>;
pub fn install_recovered(home, project_id, lock: &RecoveryLock) -> Result<(), LedgerError>;
pub fn check_built(path: &Path, project_id: &str) -> Result<(String, u64), LedgerError>;
pub fn decide(row: &ReservationRow, card_generation: i64, obs: &Observations) -> ReservationDecision;
pub fn stale_temp(home: &Path, project_id: &str) -> Result<Vec<StaleTemp>, LedgerError>;
```

오류는 `LedgerError`: 원본이 쓸 수 없으면 `Unavailable`(사유 글), 다른 프로젝트 파일이면 `IdCollision`, WAL 이면 `WalUnexpected`, 새 스키마면 `SchemaNewer`, HOME 이 Go 세대면 `HomeNotReady`. 거절은 `Constraint("<낱말>: …")` 이고 낱말은 `stale_plan` · `recovery_in_progress` · `lock_mismatch` · `target_healthy` · `target_not_damaged` · `target_not_regular` · `recovering_not_regular` · `lock_not_regular` · `suffix_in_use` · `invalid_suffix` · `damaged_not_set_aside` · `nothing_to_install` · `check_failed`.

### 원본 열기

일반 파일(링크·폴더 아님)이고, SQLite 머리글이며, WAL 아님(바이트 18·19), `application_id`(바이트 68–71)와 `user_version`(60–63)이 이 빌드의 것이어야 한다(머리글에서 직접 읽으므로 망가진 페이지가 가리지 못한다). 옆에 `<원본>-journal` 이 있으면 `needs_recovery`(읽기 전용 연결은 되감지 못한다; 쓰는 쪽이 한 번 열어 되감은 뒤 다시 계획). 연결은 읽기 전용이고 `writable_schema` 를 켠다(쓰지 않는다; 끝이 잘린 파일을 SQLite 가 통째로 거절하지 않고 남은 페이지를 읽게 한다). `meta.project_id` 를 읽을 수 없으면(스키마 페이지가 잘림) 프로젝트를 확인할 수 없어 거절한다.

### 표 읽기 (관용)

`card`, `card_ac`, `card_note`, `card_link`, `event`(다섯 표)와 `reservation`, `reservation_scope` 를 기본 키 순서로 읽는다. 읽기 오류는 그 표를 **마지막 좋은 행에서 멈추고**(`state: partial`), 한 행도 못 읽고 오류면 `unreadable`. 계획은 중단되지 않는다. 잃은 행 수 `lost_estimate` = 원본 자신의 id 계수(`sqlite_sequence`) − 마지막 좋은 id (계수를 못 읽으면 `null`).

### 계획 JSON (`ledger-recover-plan` 의 `result`)

| 키 | 뜻 |
|---|---|
| `project_id`, `source`, `ledger`, `target` | `target` = `<원장>.recovering` |
| `source_check` | 원본의 `quick_check`: `ok` · 첫 불평 · `unreadable: …` (digest 에 안 들어감) |
| `tables[]` | `table`, `state`(`ok`/`partial`/`unreadable`), `rows_read`, `last_good_id`, `lost_estimate`, `error`, `copied`, `skipped_duplicate`(이미 있는 `event_key`), `rejected`(스키마가 거절) |
| `unreadable_tables[]` | `state: unreadable` 인 표 |
| `decisions[]` | 활성 예약마다: `reservation_id`, `card_id`, `generation`, `holder`, `source_status`, `decision`(`restore_held`/`held_for_review`), `reasons[]`, `unreadable`(행을 예약으로 읽지 못했으면 `true`; 그때 읽은 만큼만 채우고 못 읽은 정수는 `0`, 사유는 `unreadable_row` 하나) |
| `unreadable_reservations` | 예약 행 중 어느 열이든 스키마가 말한 종류의 값이 아닌 행 수(예: 정수 열의 글자 `'1 '`). 활성이면 `held_for_review`(`unreadable_row`), 아니면 복사하지 않고 `rejected` 에 `unreadable_row` 로 적는다. 절대 `held` 가 되지 않는다 |
| `schema_differs[]` | 원본의 일곱 표 정의가 이 빌드의 것과 다른 곳: `{table, column, expected, found}`(`TYPE notnull=n pk=n`, `absent`, `unreadable: …`; 최대 200개). 보고만 하고 거절하지 않는다 — 안전망은 못 읽은 행 규칙 |
| `restore_held`, `held_for_review`, `copied_as_is` | 개수. `released`·`reclaimed` 행은 그대로 복사 |
| `card_generations_raised` | 세대를 올려 되살린 카드 수 |
| `rejected_total`, `rejected[]` | 거절된 행(`table`, `id`, `reason`); 목록은 최대 200개, 개수는 전부 |
| `active_cards_without_holder[]` | `active` 인데 활성 예약이 없는 카드(사람이 봐야 함) |
| `target_state` | 대체할 수 있는 것: `absent` · `unavailable`(내용이 쓸 수 없다고 확인됨). 거절하는 것: `healthy`(`target_healthy`), `not_regular`(`target_not_regular`), 그리고 `target_not_damaged` 인 `busy`(다른 연결이 잡고 있음; 확인 대기는 0.3초) · `schema_newer` · `id_collision` · `wal` · `needs_recovery`(남은 롤백 저널) · `unconfirmed`(읽을 수 없는 파일·입출력 오류 등) |
| `target_detail` | 위 거절 상태의 오류 글(없으면 `null`) |
| `stale_temp[]` | 죽은 `register` 임시 파일 `{name, pid}` (보고만) |
| `plan_digest` | 내용 기반: 읽은 모든 행, 계수, 결정(사유 포함), 거절 목록, 세대 상향, 못 읽은 행 수, 정의 차이. **대상 원장의 상태는 digest 에 없다**(apply 가 잠금 안에서 다시 판단한다). **원본 경로·관측 중 쓰이지 않은 것·시각은 안 들어간다** |
| `observed`, `apply`, `never`, `network`, `writes` | 명령 층이 더함. `writes` 는 항상 `false` |

### 되살리기 규칙 (`decide`)

대상: 원본 상태가 `held` · `expired_candidate` · `held_for_review` 인 행. 세 조건이 **모두** 맞을 때만 `restore_held`(새 원장에서 `held`), 아니면 `held_for_review`(이 상태도 카드의 보유자 자리를 차지해 새 예약을 막는다). 사유는 아래 낱말이고 이 순서로 모두 나열한다.

| 조건 | 맞지 않을 때의 사유 |
|---|---|
| 보유자 이름이 비어 있지 않다 | `holder_empty` |
| 예약 세대 ≥ 카드의 되살린 세대 (= max(원본 카드 세대, 그 카드의 모든 예약 세대)) | `generation_below_card` |
| (a) 저널에 같은 `holder`·`generation` 기록이 있고 그 중 하나가 live | `journal_missing` (기록 없음) / `journal_not_live` |
| (b) worktree 가 기록돼 있고, 관측됐고, 있고, 예약의 가지에 있다 | `worktree_not_recorded` / `worktree_not_observed` / `worktree_missing` / `branch_mismatch` (예약 가지가 없거나 분리 HEAD 도) |
| (c) 예약의 가지에 대한 원격 관측이 있고 모두 일치(`exists`, 주어진 `card_id`·`holder`·`generation` 같음) | `remote_absent` (관측 없음 = 확인 불가) / `remote_mismatch` |

`clean`(worktree 에 미커밋 변경 없음)은 참고용이며 결정에 쓰이지 않는다. lease(`lease_until`)는 원본 그대로 두고 연장하지 않는다. 읽을 수 없는 행(타입이 틀림)은 판단하지 않고 스키마가 거절한다.

### `Observations` JSON (`--observations` 파일의 모양, 모든 키 선택, 모르는 키는 거절)

```json
{
  "journal":   [{"holder": "agent-27", "generation": 3, "live": true}],
  "worktrees": [{"path": "/abs/wt", "exists": true, "branch": "issue/3068-p6", "clean": true}],
  "remotes":   [{"branch": "issue/3068-p6", "exists": true, "card_id": 7, "holder": "agent-27", "generation": 3}]
}
```

명령 층은 `journal` 을 클레임 저널(`claim::Journal`, 클레임마다 마지막 기록; `holder` = owner id, `generation` = epoch, `live` = 활성 상태이고 `lease_until` 이 `--now` 보다 뒤)에서 만들고(파일의 `journal` 을 더함), `worktrees` 는 예약의 worktree 마다 이 컴퓨터에서 읽기 전용 `git`(환경 정리)으로 찾는다(폴더가 worktree 의 맨 위여야 `exists`). 파일의 `worktrees` 는 이 컴퓨터에 없는 경로에만 쓰인다. `remotes` 는 파일에서만 온다. 파일은 최대 1 MiB, 최상위는 JSON 객체여야 한다.

## apply

0. 접미사를 확인한다(`invalid_suffix`). 그다음 **잠금 없이 읽기만 하는 1차 판단**: 이미 이 계획의 결과가 제자리에 있으면 `unchanged`, digest 가 다르거나(`stale_plan`) 대상이 바꿀 수 없는 상태이면(`target_healthy`·`target_not_regular`·`target_not_damaged`) 거절한다. **거절은 잠금 파일도 만들지 않아서 HOME 이 그대로다.** 1차를 통과한 적용만 **잠금을 건다**(`recovery_in_progress`; `.recover.lock` 이 일반 파일이 아니면 `lock_not_regular`, 열기 전에 `lstat` 로 거르고 `O_NOFOLLOW | O_NONBLOCK` 으로 열어 `fstat` 로 다시 확인). 잠금 안에서 아래 1·2 를 **다시** 하며 그쪽이 결정한다. 이 아래 전부가 잠금 안이다.
1. 이미 이 계획의 결과가 제자리에 있으면(쓸 수 있는 원장에 사건 `recover:<plan_digest>` 가 있다) `unchanged` — 원본을 읽지도 않고 아무것도 쓰지 않는다.
2. 계획을 다시 계산해 digest 가 같아야 한다(`stale_plan`). `target_state` 가 `absent`/`unavailable` 이 아니면 거절(`target_healthy`·`target_not_regular`·`target_not_damaged`). 만들고 난 뒤 바꾸기 직전에 한 번 더 판단해 달라졌으면 만든 파일을 지우고 거절한다.
3. `build_recovering`: 끝나지 않은 `.recovering` 정리, P1 과 같은 만들기(`schema::create_v1_with`: 표 → `meta` → 행 복사 → 트리거; 트리거가 마지막이라 복사한 `rev`·시각이 바뀌지 않는다), 명시한 id 로 복사(부모 먼저), `sqlite_sequence` = max(원본 계수, 복사한 최대 id) (**잃은 행의 id 를 다시 내지 않는다**), 사건 `recover` 한 줄(`event_key` = `recover:<plan_digest>`, payload: 표별 개수, 결정, 원본 경로, digest), `integrity_check` + `foreign_key_check`.
4. `set_aside_damaged`: 하드 링크 `<원장>.damaged-<UTC>`(이미 같은 파일의 `damaged` 이름이 있으면 재사용); 다른 파일이 그 이름이면 `suffix_in_use` (이 호출이 만든 `.recovering` 은 잠금 안에서 지우고 끝낸다 — 자기 임시 파일만 남기지 않는다).
5. `install_recovered`: `.recovering` 을 원장 이름 위로 `rename`(원자적). 다른 이름이 없는 파일 위로는 거절(`damaged_not_set_aside`).

어느 걸음에서 죽어도 원래 이름은 옛 파일이거나 새 파일이다. 다시 실행하면 수렴한다. `event_key` 는 복사로 그대로 살아서(`ON CONFLICT(event_key) DO NOTHING`) 재생해도 늘지 않는다.

### 남은 파일은 원래 모드 그대로

남긴 `<원장>.damaged-<UTC>` 는 망가진 원래 파일의 하드 링크라 **원래 파일의 모드를 그대로** 가진다. 바꾸지 않는다(같은 inode 를 chmod 하면 바꾸기 전에 원본의 모드가 달라진다). 새 원장은 0600 이다.

### 되살린 `held` 행을 SQL 로 직접 넣는 이유

`copy_rows` 는 되살린 예약을 `reservation.rs` 의 함수가 아니라 직접 `INSERT` 로 넣는다. `reserve`·`handoff` 는 **새 세대를 부여하고 사건을 쓰고 `rev` 를 올리는** 함수라, 원본의 `reservation_id`·세대·`acquired_at`·`lease_until` 을 그대로 지켜야 하는 복구에는 쓸 수 없다. 대신 같은 스키마의 제약(`UNIQUE(card_id, generation)`, 카드당 활성 하나의 부분 색인, CHECK)이 그대로 걸려서 어긋난 행은 거절 목록에 남는다. 카드의 세대는 낮추지 않고, 되살린 `held` 행은 `fence` 를 통과한다(시험).

### 폴더 보강 (C3·C4·R2·R3·R4)

폴더의 소유자 전용 모드와 링크 거절은 `ledger_folders.md` 에 있다. 복구에 닿는 부분: 모든 복구 걸음은 `.moai`·`projects`·`<id>` 에 링크가 있으면 `path_link` 로 거절하고(`target_state` 는 `unconfirmed`), 복구는 폴더를 만들지 않는다. 헤더·`meta` 의 일시 I/O 오류는 `Io` 라서 `unconfirmed` → `target_not_damaged` 로 거절된다(바꿔 끼우지 않음). `.recover.lock` 은 `lock_not_regular` 로 거른다.

## 하지 않는 것 (다음 조각 몫)

- **만료된 임대는 `decide` 가 보지 않는다.** `lease_until` 이 지난 행도 저널·worktree·원격 세 관측이 맞으면 `held` 로 돌아온다. 다른 사람이 못 잡고 본인도 갱신 못 하는 상태가 될 수 있다. P4 의 합격 조건이다(만료 행은 `held_for_review` 또는 `expired_candidate` 로 두거나 회수 절차를 거친다). 이 조각은 고치지 않는다.
- 한 카드에 활성 행이 둘이면 id 가 낮은 쪽이 자리를 차지한다(거절은 목록에 나온다). 범위 겹침은 복구에서 다시 확인하지 않는다.

## 소비자에게 (#3070 `moai card`, #3071, P4, P7)

- 복구한 원장의 카드·기준·메모·링크 id, `rev`, 시각, `event_key` 는 원본 그대로다. 다음 id 는 원본이 낸 가장 큰 id + 1 이상이다.
- `held_for_review` 예약이 있는 카드는 `reserve` 가 `already_held` 로 거절한다(카드가 `ready` 일 때). 풀어 주는 함수는 아직 없다(P4 의 회수·검토 몫).
- 이 조각이 안 한 것: `held_for_review` 를 풀거나 `reclaimed` 로 바꾸는 함수(P4), `active_cards_without_holder` 의 카드를 고치는 일, 원본 사건의 `at` 을 옮기는 일(옮기지 않고 그대로 둔다).
