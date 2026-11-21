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

이 SPEC 은 원장을 쓰기로 여는 순간(`open_write`) 그 트리거를 검사하고, `card.landed` 예외가 없으면 한 트랜잭션 안에서 파일의 백업 사본을 만든 뒤 트리거를 새 DDL 과 같은 글자로 바꾸게 해요. 행은 하나도 건드리지 않고, 이미 맞는 파일은 아무것도 하지 않으며(두 번째 열기는 변화 0), 중간에 실패하면 롤백돼서 옛 파일이 그대로예요. 실제 원장은 손으로 고치지 않아요. 새 빌드를 쓰는 첫 쓰기가 이 이주를 해요.

## Non-goals

- `PRAGMA user_version` 올리기. 올리면 이전 빌드가 이주된 파일을 `schema_newer` 로 거절하고 `SCHEMA_VERSION` 을 쓰는 다른 파일·시험이 함께 바뀌어야 해서, 트리거 글자 하나를 고치는 이 일에는 과해요(결정 D-1).
- `moai ledger-migrate --dry-run` 명령. 이 SPEC 은 그 명령이 부를 읽기 전용 함수(`done_guard_status`)까지만 만들어요. 명령 연결은 `src/ledger/command.rs`·`src/main.rs` 를 쓰는 후속이에요.
- 다른 트리거·표·열의 이주. 이번에는 `card_done_guard` 하나예요.
- 실제 `~/.moai/projects/*/ledger.db` 편집. 시험은 임시 HOME 의 파일만 써요.

## Requirements

- REQ-1: When a writer opens a ledger whose `card_done_guard` trigger does not carry the `card.landed` exception, `open_write` shall copy the ledger file to a new timestamped backup beside it, replace that trigger with the current definition and record the backup name in the `meta` table, all inside one `BEGIN IMMEDIATE` transaction.
- REQ-2 (REQ-1): IF the backup cannot be made or does not read back as the same ledger, THEN `open_write` shall return an error and leave the trigger, the rows and the `meta` table as they were.
- REQ-3 (REQ-1): IF any statement of the migration fails after the trigger was dropped, THEN the transaction shall roll back so that the ledger file reads exactly as before.
- REQ-4: When `open_write` opens a ledger whose `card_done_guard` already carries the `card.landed` exception, `open_write` shall make no backup and write nothing. · failure: none — nothing is written, so there is nothing to undo.
- REQ-5: When the migration has run, every row of every table shall read as it did before. · failure: none — a migration that fails rolls back (REQ-3), so no row can change.
- REQ-6: When the migration has run, the replaced trigger text shall equal the `card_done_guard` text that a freshly registered ledger holds. · failure: none — a different text is a failed test, not a runtime state.
- REQ-7: When a reader asks `done_guard_status` on any open connection, the function shall report `current` or `needs_migration` and write nothing. · failure: none — the function only reads; a read error is returned as the ledger error.
- REQ-8: When the migration has run, the done guard shall accept a card that has a `card.landed` event or a merged pull request link and refuse a card that has neither. · failure: none — a refused close is the guard's own error, as before.

## Acceptance

- AC-1 (REQ-1, REQ-6): 옛 트리거(예외 없음)를 가진 원장을 `open_write` 로 열면 백업 파일이 하나 생기고, 트리거 글자가 새로 등록한 원장의 것과 같고, `meta` 에 백업 이름이 적혀요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- migrates_an_old_ledger`
- AC-2 (REQ-5, REQ-8): 이주 뒤 모든 표의 행이 그대로이고, landed 카드는 PR 없이 닫히고, 병합 PR 길도 닫히고, 둘 다 없는 카드는 거절돼요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- rows_and_close_rules`
- AC-3 (REQ-4): 두 번째 열기는 백업을 더 만들지 않고 파일 바이트가 그대로예요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- second_open_is_a_noop`
- AC-4 (REQ-3): 트리거를 지운 뒤 실패하는 이주는 롤백돼서 옛 트리거와 행이 그대로 읽혀요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- failure_rolls_back`
- AC-5 (REQ-2): 백업을 만들 수 없을 때(폴더가 쓰기 금지, 백업 이름이 이미 있음) 이주하지 않고 오류를 내요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- backup_failure_refuses`
- AC-6 (REQ-7): 읽기 전용 상태 확인은 옛 파일에 `needs_migration`, 이주된 파일에 `current` 를 내고 파일을 바꾸지 않아요 · verify: `cargo test --offline --locked --test ledger_card_landed_migration_cli -- status_is_read_only`

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

- 백업을 `std::fs::copy` 로 만들어요(`rusqlite` 의 backup 기능은 켜져 있지 않고 Cargo.toml 은 이 SPEC 범위 밖이에요). `BEGIN IMMEDIATE` 가 예약 잠금을 쥐고 있어 다른 쓰기가 끼어들 수 없고, 아직 아무것도 쓰지 않은 상태에서 복사하므로 파일은 일관돼요. 복사본은 `PRAGMA quick_check`, `application_id`, `user_version` 으로 읽어 확인해요.
- 동시에 두 프로세스가 같은 옛 원장을 열면 한쪽이 먼저 이주하고, 다른 쪽은 잠금을 얻은 뒤 트리거를 다시 읽어 아무것도 하지 않아요(AC-3 와 같은 길).
- 백업 이름은 DB 시계의 UTC 초를 써서 `ledger.db.bak-<UTC>` 예요. `recover` 가 찾는 `ledger.db.bak` 과 이름이 달라서 복구 순서에 끼어들지 않아요(복구가 새 이름을 쓰게 하는 일은 후속).
- `user_version` 이 그대로 1 이라서 옛 빌드도 이주된 파일을 열 수 있어요. 옛 빌드의 트리거 확인은 없으므로 이주된 파일에서도 정상 동작해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `user_version` 을 올리지 않고, 트리거에 `card.landed` 예외가 없으면 바꾼다(되돌릴 수 있음, 범주 none) | 올리면 옛 빌드가 파일을 못 열고 `mod.rs`·여러 시험이 걸려 범위가 넓어져요. 카드는 1→2 를 예로 들었지만 기준은 "이주·백업·무손실·반복 안전"이에요. 레인이 정하고 리더가 되돌릴 수 있어요 | 후속 SPEC 에서 `SCHEMA_VERSION` 을 2 로 올리고 이 함수를 그 이주의 첫 단계로 옮겨요 |
| D-2 | 백업은 파일 복사 + 읽어 확인, 이름은 `ledger.db.bak-<UTC>` (되돌릴 수 있음, 범주 none) | backup 기능이 꺼져 있고 Cargo.toml 은 범위 밖 | 백업 파일은 사용자가 지워도 되고, 이름 규칙은 한 줄 바꿔요 |
| D-3 | 이주 기록은 `meta` 행 `done_guard_landed` 하나 (되돌릴 수 있음, 범주 none) | `event` 표에 사건을 더하면 이를 읽는 쪽에 새 종류가 보여요. `meta` 는 읽는 곳이 `project_id` 등 이름 지정뿐이에요 | 행을 빼도 판정은 트리거 글자로 하므로 영향 없어요 |

## Result

### 만든 것 (head bd63d4a6941f8c5109e33002e3a8c8efa5bd990c)

- `src/ledger/schema.rs`: `TRIGGERS` 를 `card_done_guard` 앞(`TRIGGERS_BEFORE_GUARD`)·그 트리거(`CARD_DONE_GUARD`)·뒤(`TRIGGERS_AFTER_GUARD`)로 나눴어요. 새 원장과 이주가 같은 글자를 써요.
- `src/ledger/migrate.rs`: `done_guard_status`(읽기 전용), `upgrade_done_guard` 와 실패 지점을 끼우는 `upgrade_done_guard_with`, `ledger.db.bak-<UTC>` 복사(`create_new`, 0600)와 읽어 확인(`quick_check`, `application_id`, `user_version`), `meta` 행 `done_guard_landed`.
- `src/ledger/open.rs`: `open_write` 가 연 직후 이주를 부르고 `WriteLedger::done_guard_upgrade()` 로 백업 경로를 알려 줘요.
- 시험: `tests/ledger_card_landed_migration_cli.rs` 10개, `ledger::migrate` 단위 시험 2개 추가.

### 증거 (이 head 에서 돌린 것, 2026-11-21)

- `cargo test --offline --locked --test ledger_card_landed_migration_cli`: 10 passed. RED: 구현 전에는 새 함수가 없어서 시험이 컴파일되지 않았어요.
- `cargo test --offline --locked --lib ledger::migrate`: 3 passed. `cargo fmt --check` 0, `cargo clippy --offline --locked --all-targets -- -D warnings` 0.
- 이웃 14개 대상(`ledger_schema`, `ledger_open`, `ledger_recover`, `ledger_recover_cli`, `ledger_reservation`, `ledger_workflow`, `ledger_import`, `ledger_cli`, `ledger_readonly_cli`, `ledger_folders`, `ledger_scope_return`, `card_done_mflow_cli`, `card_spec_cli`, `card_cli`): 261 passed, 0 failed.
- 실제 원장의 사본(읽기 전용 `sqlite3 -readonly ... .backup` 으로 임시 HOME 에 복사)에서 실행: 복사본은 `user_version` 1, 카드 288장, `card_done_guard` 에 `card.landed` 없음(리더가 읽은 것과 같음). `card note` 쓰기 한 번으로 이주가 돌았고, 백업 파일의 데이터 표 덤프 해시가 이주 전 덤프와 같고(`5efbfd2c26eb910698b6d075d430414bcf7836d7`), 이주된 사본은 시험 삼아 쓴 노트 한 줄과 그 사건 한 줄만 다르며, 트리거에 예외가 생겼고, `integrity_check` 가 ok, `user_version` 1 이에요. 실제 `~/.moai/projects/moai-adk/ledger.db` 는 건드리지 않았고(예외 없음, 백업 파일 없음을 다시 읽어 확인) 이주는 새 빌드를 쓰는 첫 쓰기가 해요.
- 변형 7개(상한 12): 상태가 늘 `NeedsMigration`, 백업 확인이 늘 ok, 실패해도 백업 남김, commit 이 훅 오류를 무시, 백업 덮어쓰기 허용, `meta` 키 변경, `open_write` 가 이주를 안 부름 — 전부 시험 `FAILED` 로 죽었고 컴파일 실패로 센 것은 없어요. 변형 뒤 트리는 변경 0 으로 복원했어요. 시험으로 못 죽이는 것은 하나 알고 있어요: 트랜잭션 안에서 상태를 다시 읽는 줄(동시에 두 프로세스가 같은 옛 원장을 열 때만 쓰임)을 빼는 변형은 결정적으로 시험할 길이 없어서(저장소가 고정 sleep 시험을 금지) 돌리지 않았어요.

### 수용 기준 결과

- AC-1: met — `migrates_an_old_ledger` 통과(백업 한 개, 트리거 글자가 새 원장과 같음, `meta` 행, 백업의 행·옛 트리거·pragma, 0600).
- AC-2: met — `rows_and_close_rules` 통과(행 덤프 같음, landed 카드와 병합 PR 카드는 닫히고 둘 다 없는 카드는 거절).
- AC-3: met — `second_open_is_a_noop` 와 `a_fresh_ledger_is_left_alone` 통과(백업 한 개 그대로, 파일 바이트 같음).
- AC-4: met — `failure_rolls_back` 통과(훅이 트리거를 지운 뒤 실패해도 옛 트리거·행·`integrity_check`, 백업 없음, 다음 열기에서 이주).
- AC-5: met — `backup_failure_refuses`(폴더 쓰기 금지)와 `an_existing_backup_name_is_never_overwritten` 통과.
- AC-6: met — `status_is_read_only` 통과(읽기 전용 열기에서 `NeedsMigration`, 파일 바이트 같음, 이주 뒤 `Current`).

### 남은 것

- `moai ledger-migrate --dry-run` 명령: `done_guard_status` 를 부르는 명령은 `src/ledger/command.rs`·`src/main.rs` 몫이라 이 카드 범위 밖이에요(후속).
- 복구(`recover`)가 찾는 `ledger.db.bak` 과 이름이 달라 새 백업은 복구 순서에 끼지 않아요. 복구가 `.bak-<UTC>` 도 읽게 하는 일은 후속.
- 사용자가 새 빌드를 적용하면 첫 쓰기에서 이주가 돌아요(적용은 리더가 사용자에게 먼저 물어요). 그 뒤 `~/.moai/projects/<id>/` 에 `ledger.db.bak-<UTC>` 가 하나 생겨요.
