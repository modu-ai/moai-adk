---
id: SPEC-20261121-ledger-card-landed-migration
title: "An existing ledger gets the card.landed exception in its done guard, with a backup"
status: draft
revision: 1
created: 2026-11-21
---

# An existing ledger gets the card.landed exception in its done guard, with a backup

## Intent

`card done` 은 PR 이 병합됐거나 `card.landed` 사건이 있으면 카드를 닫아요(SPEC-20261120-card-done-mflow). 그런데 그 예외는 `src/ledger/schema.rs` 의 DDL 에만 있어서 새로 만드는 원장에만 들어가요. 이미 있는 원장 파일은 `card_done_guard` 트리거가 옛 글자(병합된 PR 만 인정) 그대로예요. 그래서 `card landed` 는 저장되는데 `card done` 은 DB 트리거에서 `cannot close: no merged pull request is linked` 로 거절돼요. 리더가 `sqlite3 -readonly` 로 두 실제 원장(moai-adk, mo-ai-kr)의 `PRAGMA user_version` 이 1 이고 트리거에 `card.landed` 가 없음을 확인했어요.

이 SPEC 은 원장을 쓰기로 여는 순간(`open_write`) 그 트리거를 검사하고, `card.landed` 예외가 없으면 한 트랜잭션 안에서(쓰기 잠금을 쥔 채) SQLite 자신이 백업 사본을 만든 뒤 트리거를 새 DDL 과 같은 글자로 바꾸게 해요. 행은 하나도 건드리지 않고, 이미 맞는 파일은 아무것도 하지 않으며(두 번째 열기는 변화 0), 중간에 실패하면 롤백돼서 옛 파일이 그대로예요. 실제 원장은 손으로 고치지 않아요. 새 빌드를 쓰는 첫 쓰기가 이 이주를 해요.

## Non-goals

- `PRAGMA user_version` 올리기. 올리면 이전 빌드가 이주된 파일을 `schema_newer` 로 거절하고 `SCHEMA_VERSION` 을 쓰는 다른 파일·시험이 함께 바뀌어야 해서, 트리거 글자 하나를 고치는 이 일에는 과해요(결정 D-1).
- `moai ledger-migrate --dry-run` 명령. 이 SPEC 은 그 명령이 부를 읽기 전용 함수(`done_guard_status`)까지만 만들어요. 명령 연결은 `src/ledger/command.rs`·`src/main.rs` 를 쓰는 후속이에요.
- 다른 트리거·표·열의 이주. 이번에는 `card_done_guard` 하나예요.
- 실제 `~/.moai/projects/*/ledger.db` 편집. 시험은 임시 HOME 의 파일만 써요.

## Requirements

- REQ-1: When a writer opens a ledger whose `card_done_guard` trigger does not carry the `card.landed` exception, `open_write` shall make a consistent backup of the ledger through SQLite (a `VACUUM INTO` a new timestamped file beside it, never opening the ledger file with `std::fs`), replace that trigger with the current definition and record the backup name in the `meta` table, all inside one `BEGIN IMMEDIATE` transaction.
- REQ-2 (REQ-1): IF the backup cannot be made or does not read back as the same ledger, THEN `open_write` shall return an error that says what to do (free disk space, folder write rights, the path) and leave the trigger, the rows and the `meta` table as they were.
- REQ-3 (REQ-1): IF any statement of the migration fails after the trigger was dropped, THEN the transaction shall roll back so that the ledger file reads exactly as before and the backup made by that call is removed only after a read shows the rollback.
- REQ-4: When `open_write` opens a ledger whose `card_done_guard` already carries the `card.landed` exception, `open_write` shall make no backup and write nothing. · failure: none — nothing is written, so there is nothing to undo.
- REQ-5: When the migration has run, every row of every table shall read as it did before. · failure: none — a migration that fails rolls back (REQ-3), so no row can change.
- REQ-6: When the migration has run, the replaced trigger text shall equal the `card_done_guard` text that a freshly registered ledger holds. · failure: none — a different text is a failed test, not a runtime state.
- REQ-7: When a reader asks `done_guard_status` on any open connection, the function shall report `current` or `needs_migration` and write nothing. · failure: none — the function only reads; a read error is returned as the ledger error.
- REQ-8: When the migration has run, the done guard shall accept a card that has a `card.landed` event or a merged pull request link and refuse a card that has neither. · failure: none — a refused close is the guard's own error, as before.
- REQ-9: While the migration transaction is open, no other process shall be able to write the ledger, from the moment the lock is taken until the commit. · failure: none — a second process tries to take the write lock at three stages of the upgrade, and a lost lock is a failed test.
- REQ-10: When several processes open the same old ledger at once, exactly one of them shall upgrade it and make a backup. · failure: none — each other process finds the trigger current once it has the lock.
- REQ-11: When the migration has run, `open_write` shall print one line that names the backup path. · failure: none — a later open finds nothing to do and prints nothing.

## Acceptance

- AC-1 (REQ-1, REQ-6): 옛 트리거(예외 없음)를 가진 원장을 `open_write` 로 열면 백업 파일이 하나 생기고, 트리거 글자가 새로 등록한 원장의 것과 같고, `meta` 에 백업 이름이 적혀요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- migrates_an_old_ledger`
- AC-2 (REQ-5, REQ-8): 이주 뒤 모든 표의 행이 그대로이고, landed 카드는 PR 없이 닫히고, 병합 PR 길도 닫히고, 둘 다 없는 카드는 거절돼요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- rows_and_close_rules`
- AC-3 (REQ-4): 두 번째 열기는 백업을 더 만들지 않고 파일 바이트가 그대로예요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- second_open_is_a_noop`
- AC-4 (REQ-3): 트리거를 지운 뒤 실패하는 이주는 롤백돼서 옛 트리거와 행이 그대로 읽혀요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- failure_rolls_back`
- AC-5 (REQ-2): 백업을 만들 수 없을 때(폴더가 쓰기 금지, 백업 이름이 이미 있음) 이주하지 않고 오류를 내요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- backup_failure_refuses`
- AC-6 (REQ-7): 읽기 전용 상태 확인은 옛 파일에 `needs_migration`, 이주된 파일에 `current` 를 내고 파일을 바꾸지 않아요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- status_is_read_only`
- AC-7 (REQ-9): 업그레이드의 세 지점(잠금 직후, 백업 뒤, 트리거 삭제 뒤)에서 다른 프로세스가 쓰기 잠금을 못 잡고, 커밋 뒤에는 잡아요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- the_write_lock_is_held_from_the_lock_to_the_commit`
- AC-8 (REQ-10, REQ-11): 프로세스 8개가 동시에 같은 옛 원장을 열어도 모두 성공하고, 이주와 백업은 한 번이며, 알림 한 줄이 한 프로세스에서만 나와요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- concurrent_opens_upgrade_once_and_every_open_succeeds`
- AC-9 (REQ-2): 백업을 만들 수 없을 때의 거절 문구가 해결 방법(남은 공간, 폴더 쓰기 권한, 경로)을 말해요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- a_refused_backup_says_what_to_do`

## Scope

- write: `src/ledger/schema.rs`, `src/ledger/open.rs`, `src/ledger/migrate.rs`, `tests/ledger_card_landed_migration_cli.rs`, `.moai/specs/SPEC-20261121-ledger-card-landed-migration/**`
- read-only: `src/ledger/mod.rs`, `src/ledger/recover.rs`, `src/ledger/reservation.rs`, `tests/ledger_schema.rs`, `tests/card_done_mflow_cli.rs`
- forbidden: every project ledger under the real HOME, writes to `src/ledger/reservation.rs`, network

## Plan

1. RED: `tests/ledger_card_landed_migration_cli.rs` 에 옛 원장 픽스처(새 DDL 로 등록한 뒤 `card_done_guard` 를 옛 글자로 바꿈)와 AC-1~6 시험을 먼저 쓰고 실패를 확인해요.
2. `src/ledger/schema.rs`: `TRIGGERS` 를 `card_done_guard` 앞뒤로 나누고 그 트리거를 `CARD_DONE_GUARD` 상수로 빼서, 새 원장과 이주가 같은 글자를 써요.
3. `src/ledger/migrate.rs`: `done_guard_status`, `upgrade_done_guard`(와 실패 지점을 끼우는 `_with`), 백업 만들기·읽어 확인하기.
4. `src/ledger/open.rs`: `open_write` 가 연결을 연 직후 이주를 부르고, 결과를 `WriteLedger` 가 알려 줘요.
5. 이웃 시험(`ledger_schema`, `ledger_open`, `ledger_recover`, `card_done_mflow_cli`)과 변형 시험.

## Risks

- 백업은 쓰기 잠금을 쥔 채, SQLite 의 두 번째 연결이 하는 `VACUUM INTO` 로 만들어요. 원장 파일을 `std::fs` 로 열면 안 돼요: POSIX 잠금은 프로세스 것이라서 같은 프로세스가 그 파일의 어떤 파일 기술자든 닫으면 `BEGIN IMMEDIATE` 가 쥔 잠금까지 풀려 다른 프로세스가 트랜잭션 도중에 써요(sqlite.org/howtocorrupt.html 2.2절, 2026-11-21 읽음). 처음 구현은 `std::fs::copy` 로 이 잘못을 했고 감사가 3/3 재현했어요(`disk I/O error`, 이중 이주). SQLite 연결끼리는 한 프로세스 안에서 파일 잠금 상태를 함께 쓰고 마지막 잠금이 풀릴 때까지 기술자를 닫지 않아서 안전해요. 이 가정은 두 프로세스 시험(AC-7)이 세 지점에서 확인해요.
- `VACUUM INTO` 사본은 같은 행·스키마·`application_id`·`user_version` 을 가진 압축된 파일이지 바이트가 같은 파일이 아니에요. 사본은 `PRAGMA quick_check`, `application_id`, `user_version` 으로 읽어 확인하고, 시험은 모든 표를 덤프해서 이주 전과 같은지 봐요.
- 동시에 여러 프로세스가 같은 옛 원장을 열면 잠금을 얻은 쪽만 이주하고 백업을 만들어요. 나머지는 잠금을 얻은 뒤 트리거를 다시 읽고 아무것도 하지 않아요(AC-8, 8개 프로세스).
- 이주가 실패했을 때 백업은 트리거를 다시 읽어 롤백이 보일 때만 지워요. 커밋이 됐을 수 있거나 읽을 수 없으면 백업을 남겨요(지우는 쪽이 위험해서예요).
- 백업 이름은 DB 시계의 UTC 초를 써서 `ledger.db.bak-<UTC>` 예요. `recover` 가 찾는 `ledger.db.bak` 과 이름이 달라서 복구 순서에 끼어들지 않아요(복구가 새 이름을 쓰게 하는 일은 후속).
- `user_version` 이 그대로 1 이라서 옛 빌드도 이주된 파일을 열 수 있어요. 옛 빌드의 트리거 확인은 없으므로 이주된 파일에서도 정상 동작해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `user_version` 을 올리지 않고, 트리거에 `card.landed` 예외가 없으면 바꾼다(되돌릴 수 있음, 범주 none) | 올리면 이주된 파일을 옛 빌드가 `schema_newer` 로 거절해요. 그러면 이주된 원장은 이 빌드로 되돌아가는 길(예: 빌드를 이전 것으로 되돌림)에서 열리지 않아요. 지금 설계는 `user_version` 1 그대로라서 이전 빌드로 되돌려도 원장이 열려요. 또한 `mod.rs`·여러 시험이 걸려 범위가 넓어져요. 카드는 1→2 를 예로 들었지만 기준은 "이주·백업·무손실·반복 안전"이에요. 레인이 정하고 리더가 되돌릴 수 있어요 | 후속 SPEC 에서 `SCHEMA_VERSION` 을 2 로 올리고 이 함수를 그 이주의 첫 단계로 옮겨요 |
| D-2 | 백업은 `VACUUM INTO` + 읽어 확인, 이름은 `ledger.db.bak-<UTC>` (되돌릴 수 있음, 범주 none) | rusqlite 의 backup 기능이 꺼져 있고 Cargo.toml 은 범위 밖이에요. `VACUUM INTO` 는 SQL 한 줄이고 일관된 사본을 만들어요 | 백업 파일은 사용자가 지워도 되고, 이름 규칙은 한 줄 바꿔요 |
| D-3 | 이주 기록은 `meta` 행 `done_guard_landed` 하나 (되돌릴 수 있음, 범주 none) | `event` 표에 사건을 더하면 이를 읽는 쪽에 새 종류가 보여요. `meta` 는 읽는 곳이 `project_id` 등 이름 지정뿐이에요 | 행을 빼도 판정은 트리거 글자로 하므로 영향 없어요 |
| D-4 | 백업을 트랜잭션 *안*에서 두 번째 SQLite 연결의 `VACUUM INTO` 로 만든다(되돌릴 수 있음, 범주 none). 리더의 결정은 "트랜잭션 전에 `VACUUM INTO` 또는 백업 API" 였고, 이 SPEC 은 그 취지(잠금을 쥔 동안 `std::fs` 로 원장을 열지 않음)를 지키되 위치만 바꿨어요 | 트랜잭션 전에 백업하면 동시 프로세스마다 백업을 하나씩 만들고(같은 초에는 `File exists`) 백업과 이주 사이에 다른 쓰기가 끼어요. 잠금을 쥔 프로세스만 백업하면 백업은 한 번이고 사본이 이주 직전 상태예요. 잠금이 유지되는지는 두 프로세스 시험이 세 지점에서 직접 확인해요 | `make_backup` 의 호출을 `BEGIN IMMEDIATE` 앞으로 옮기고 이름에 프로세스 번호를 더해요 |

## Result

### 만든 것 (감사 B1 수리 뒤. 첫 구현 head bd63d4a6941f8c5109e33002e3a8c8efa5bd990c)

- `src/ledger/schema.rs`: `TRIGGERS` 를 `card_done_guard` 앞(`TRIGGERS_BEFORE_GUARD`)·그 트리거(`CARD_DONE_GUARD`)·뒤(`TRIGGERS_AFTER_GUARD`)로 나눴어요. 새 원장과 이주가 같은 글자를 써요.
- `src/ledger/migrate.rs`: `done_guard_status`(읽기 전용), `upgrade_done_guard` 와 실패 지점을 끼우는 `upgrade_done_guard_with`, `ledger.db.bak-<UTC>` 백업(이름을 `create_new` 0600 으로 먼저 잡고 두 번째 SQLite 연결의 `VACUUM INTO` 로 채움 — 원장 파일을 `std::fs` 로 열지 않음)과 읽어 확인(`quick_check`, `application_id`, `user_version`), `meta` 행 `done_guard_landed`, 실패했을 때 트리거를 다시 읽어 롤백이 보일 때만 백업을 지우는 `settle_failed`, 거절 문구의 해결 방법(`BACKUP_HELP`), 시험이 볼 수 있는 세 지점 `Stage`.
- `src/ledger/open.rs`: `open_write` 가 연 직후 이주를 부르고 `WriteLedger::done_guard_upgrade()` 로 백업 경로를 알려 주고, 이주가 실제로 돌았을 때만 한 줄을 한 번 알려요(백업 경로 포함).
- 시험: `tests/ledger_card_landed_migration_cli.rs` 15개(두 프로세스 잠금 시험, 8프로세스 동시 열기 시험, 거절 문구 시험 포함), `ledger::migrate` 단위 시험 8개(실패한 백업과 그 `-journal` 을 지우는 시험 3개 포함).

### 증거 (2026-11-21)

이번 head(수리 뒤 작업 트리)에서 돌린 것:

- `cargo test --offline --locked --test ledger_card_landed_migration_cli`: 15 passed. `cargo test --offline --locked --lib ledger::migrate`: 8 passed(M1 수리 뒤). `rustfmt --edition 2021 --check`(고친 3개 파일) 0, `git diff --check` 0, `cargo clippy --offline --locked --all-targets -- -D warnings` 0.
- 감사가 3/3 으로 재현한 결함의 회귀 시험: `the_write_lock_is_held_from_the_lock_to_the_commit` 은 업그레이드의 세 지점(`Locked`, `BackedUp`, `GuardDropped`)에서 다른 *프로세스*(이 시험 바이너리를 다시 실행한 자식)가 `BEGIN IMMEDIATE` 를 못 잡고 `busy` 를 받으며, 커밋 뒤에는 `got` 임을 확인해요. `concurrent_opens_upgrade_once_and_every_open_succeeds` 는 프로세스 8개가 동시에 열어도 모두 성공하고, 이주 1·백업 1·알림 1이며 행 덤프가 같고 `integrity_check` ok 임을 확인해요.
- 변형 5개(상한 12): N1 `Immediate`→`Deferred`(감사에서 살아남았던 것), N2 잠금 중 `std::fs::File::open(원장)`(감사가 찾은 결함을 되살림), N3 트랜잭션 안 재확인 제거, N4 커밋됐는데도 백업 삭제, N5 알림 제거 — 전부 `FAILED` 로 죽었고 컴파일 실패로 센 것은 없어요(`target/mutants2.py`, 변형 뒤 `git diff` 가 의도한 4개 파일뿐임을 확인).
- 변형 2개(M1 수리, 단위 시험만): J1 `-journal` 이름을 틀리게, J2 롤백 경로가 예전처럼 본 파일만 지움 — 둘 다 `FAILED` 로 죽었어요(`target/mutants3.py`). M1 의 `-journal` 은 번들 SQLite 로 두 가지 실패(원본이 데이터베이스가 아님, 원본이 독점 잠김)를 만들어도 남지 않아서 결정적으로 재현하지 못했어요. 그래서 수리는 실패한 백업의 파일과 `<백업>-journal` 을 지우는 함수(`remove_backup_files`)를 두 실패 경로에 쓰게 하고 그 함수를 직접 시험해요.

감사(델타 tier 2 PASS, 감사자 보고 숫자): 복사본 두 개에서 12개 동시 열기 × 8회 — 백업 1·알림 1줄·`integrity_check` ok·`disk I/O error`/`File exists` 0; `ledger_*` 14개 대상 + `card_done_mflow_cli` 245 passed; `--lib ledger::` 29 passed. 병합 검사(레인): `check affected` 52c65fc49a(base f00cf2b126) PASS — targets=56 py=8 scans=listed:41, 모든 단계 exit 0.

이전 head 에서 잰 것(백업 방식이 `std::fs` 복사였을 때; 이번 head 에서 다시 돌리지 않았어요):

- 이웃 14개 대상(`ledger_schema`, `ledger_open`, `ledger_recover`, `ledger_recover_cli`, `ledger_reservation`, `ledger_workflow`, `ledger_import`, `ledger_cli`, `ledger_readonly_cli`, `ledger_folders`, `ledger_scope_return`, `card_done_mflow_cli`, `card_spec_cli`, `card_cli`): 261 passed, 0 failed. 이번 수리는 `ledger::migrate`·`open_write` 만 바꿨고, 병합 때 `check affected` 가 이웃을 다시 돌려요.
- 실제 원장 사본으로 한 실행(읽기 전용 `.backup` 으로 임시 HOME 에 복사한 사본): `user_version` 1, `card.landed` 없음에서 `card note` 한 번으로 이주가 돌았고 데이터 표 덤프가 이주 전과 같았어요. 백업 방식이 바뀌었으므로 이 항목은 후속 확인으로 남겨요.

### 수용 기준 결과

- AC-1: met — `migrates_an_old_ledger` 통과(백업 한 개, 트리거 글자가 새 원장과 같음, `meta` 행, 백업의 행·옛 트리거·pragma, 0600).
- AC-2: met — `rows_and_close_rules` 통과(행 덤프 같음, landed 카드와 병합 PR 카드는 닫히고 둘 다 없는 카드는 거절).
- AC-3: met — `second_open_is_a_noop` 와 `a_fresh_ledger_is_left_alone` 통과(백업 한 개 그대로, 파일 바이트 같음).
- AC-4: met — `failure_rolls_back` 통과(훅이 트리거를 지운 뒤 실패해도 옛 트리거·행·`integrity_check`, 백업 없음, 다음 열기에서 이주).
- AC-5: met — `backup_failure_refuses`(폴더 쓰기 금지)와 `an_existing_backup_name_is_never_overwritten` 통과.
- AC-6: met — `status_is_read_only` 통과(읽기 전용 열기에서 `NeedsMigration`, 파일 바이트 같음, 이주 뒤 `Current`).
- AC-7: met — `the_write_lock_is_held_from_the_lock_to_the_commit` 통과(세 지점 모두 다른 프로세스가 `busy`, 커밋 뒤 `got`).
- AC-8: met — `concurrent_opens_upgrade_once_and_every_open_succeeds` 통과(8프로세스, 이주 1, 백업 1, 알림 1줄에 백업 경로).
- AC-9: met — `a_refused_backup_says_what_to_do` 통과(`free some disk space`, `writable`, `path`, `not changed`).

### 남은 것

- `moai ledger-migrate --dry-run` 명령: `done_guard_status` 를 부르는 명령은 `src/ledger/command.rs`·`src/main.rs` 몫이라 이 카드 범위 밖이에요(후속).
- 복구(`recover`)가 찾는 `ledger.db.bak` 과 이름이 달라 새 백업은 복구 순서에 끼지 않아요. 복구가 `.bak-<UTC>` 도 읽게 하는 일은 후속.
- 트리거 판정을 `user_version` 이 아니라 글자로 하는 이유(D-1): 이주된 원장을 이전 빌드로 되돌려도 열려요.
- 사용자가 새 빌드를 적용하면 첫 쓰기에서 이주가 돌아요(적용은 리더가 사용자에게 먼저 물어요). 그 뒤 `~/.moai/projects/<id>/` 에 `ledger.db.bak-<UTC>` 가 하나 생겨요.
