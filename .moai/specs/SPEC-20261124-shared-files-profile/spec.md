---
id: SPEC-20261124-shared-files-profile
title: "프로젝트 프로필이 공유 파일을 적어 두면 card take 와 card scope 는 그 파일을 통째로 예약하지 않고 card line-lease 를 알려 줘요"
status: draft
revision: 1
created: 2026-11-24
---

# 프로젝트 프로필이 공유 파일을 적어 두면 card take 와 card scope 는 그 파일을 통째로 예약하지 않고 card line-lease 를 알려 줘요

## Intent

여러 카드가 한 줄씩만 고치는 공유 파일(`src/lib.rs` 의 `pub mod`, `src/help.rs` 줄, `tests/areas.toml`, `tools/repo_scans.json` 같은 것)을 한 카드가 통째로 예약하면 다른 카드가 막혀요. 카드 C-445 본문(사용자 선택, 2026-11-24 카드 조사)이 든 사건: C-322 의 범위가 `src/lib.rs` 를 통째로 쥐어서 C-364 의 line-lease 가 C-322 의 lease 가 끝날 때까지 거절당했어요. 쥔 범위는 줄일 수 없어요(`card scope` 는 "a held card's scope only grows" 로 거절해요).

지금은 어떤 파일이 공유 파일인지 리더가 손으로 정하고, 레인이 `card scope` 로 그 파일을 달라고 하면 리더가 "line-lease 로" 라고 답해야 해요. 이 SPEC 은 그 목록을 프로필에 두고 도구가 거절하게 해요.

2026-11-24 에 읽은 코드(NeXT `7e0aaba345`):

- 범위가 정해지는 곳은 `src/card/dispatch.rs` 의 `take`(스스로 준 `--scope`, 카드 자신의 범위, SPEC `write:` 목록, 마지막 예약의 범위; `src/card/take_scope.rs` 가 뒤 둘을 골라요)와 `scope`(비어 있는 카드는 바꾸고, 쥔 카드는 `reservation::widen_in_tx` 로 넓혀요)예요. 둘 다 입력 검사를 `checked_scope` 한 곳에서 해요.
- `card line-lease`(`src/card/work.rs`)는 `scope_return::lease_take` 로 예약에 경로를 잠깐 더하고 명령이 끝나면 돌려줘요. `take`·`scope` 를 거치지 않으니 공유 파일 거절의 영향을 받지 않아요.
- `card check`(`src/card/work.rs` `check`)는 범위 밖 경로를 `allowed: false` 로 답하지 않고 `reservation conflict in write scope: … is not under what C-n reserved (…)` 오류로 거절해요. 카드 본문이 "allowed:false" 라고 적은 것과 코드가 달라서(핵심 규칙 21) 이 SPEC 은 코드를 따라요(D-5).
- 겹침 판정은 이미 `scope::overlap(a, b)`(`src/scope.rs`)가 해요. 새 판정을 만들지 않아요.
- 프로필 키는 `src/profile.rs` 의 `PROFILE_KEYS` 표, `Profile` 필드, `validate` 세 곳에 더해요(`remote_copy`, `test_paths` 가 같은 길을 가요).

읽는 쪽은 카드를 맡는 레인, 쥔 범위를 정하는 리더, 이 변경을 감사하는 세션이에요.

## Interview

- Q1: 목록의 이름과 모양은? -> A: 프로필의 선택 키 `shared_files`, 경로 또는 glob 문자열의 목록이에요(D-1). 키가 없으면 지금과 똑같이 움직여요.
- Q2: 겹침은 어떻게 보나? -> A: `scope::overlap` 으로 양쪽을 다 봐요(D-2): 항목이 공유 파일을 덮는 경우(`src/**`)도, 공유 glob 안의 경로(`tests/fixtures/help/a.txt`)도 거절해요.
- Q3: SPEC 의 `write:` 목록이나 마지막 예약에서 채워지는 범위에 공유 파일이 있으면? -> A: 거절하지 않고 뺀 뒤 `warnings` 에 이름과 line-lease 명령을 적어요(D-3). 스스로 준 `--scope` 와 카드 자신의 범위는 거절해요. 이유는 SPEC 의 `write:` 에 공유 파일을 문서로 적어 둔 SPEC 이 많고, 그것까지 거절하면 이미 있는 SPEC 이 take 불가가 되기 때문이에요.
- Q4: 새 차단이니 observe 부터 시작하나(sdd 규칙 11)? -> A: 아니요, 처음부터 enforce 예요(D-4). 목록이 사람이 적은 명시적 선택이고, 키가 없으면 아무것도 막히지 않고, 거절은 line-lease 라는 이미 있는 길을 알려 줘요. 리더 결정이 필요해요.
- Q5: 이미 쥔 예약은? -> A: 건드리지 않아요. 새로 더하는 항목만 봐요(D-6).

## Non-goals

- `card add --scope` 의 거절: 범위가 카드에 적혀도 `take` 가 거절해요. 한 군데에서만 막아요.
- 공유 파일을 쓰는 것 자체를 막는 훅이나 `card check` 의 `allowed: false` 새 응답 모양.
- `card line-lease` 의 동작 변경: 그대로 공유 파일을 받아요(REQ-7 은 그것을 고정할 뿐이에요).
- 공유 파일을 통째로 다시 써야 하는 카드의 예외 열쇠: 그 카드는 먼저 프로필 목록을 고치는 카드를 거쳐요(D-7).
- 공유 파일을 코드에서 자동으로 찾는 것.

## Requirements

<!-- This SPEC adds refusals that block (take and scope). It does not start in observe: the list is an explicit opt-in key of the profile, an absent key blocks nothing, and the leader decides D-4. -->
- REQ-1: The project profile shall accept an optional key `shared_files` that is a list of at most 64 entries, each a path or glob that `card add --scope` would accept · failure: none — a profile without the key reads as before.
- REQ-2 (REQ-1): IF `shared_files` is not a list of valid entries, THEN the profile reader shall refuse the profile with the same message it gives any other invalid key and log no value · failure: none — the refusal is the existing one.
- REQ-3: When `card take` is given `--scope` entries, or takes a card whose own scope holds an entry, that overlaps an entry of `shared_files`, the `card take` command shall refuse, name the entry, the shared entry and the form `card line-lease <card> --generation <g> --path <path> -- <command>`, and write nothing · failure: none — the refusal is before the ledger is opened for write.
- REQ-4: When `card take` fills an empty scope from the `write:` list of the linked SPEC or from the last reservation, the `card take` command shall leave out each entry that overlaps an entry of `shared_files`, name each in `warnings` with the line-lease form, and reserve the rest · failure: none — a fill with nothing left refuses with the same text as REQ-3 and writes nothing.
- REQ-5: When `card scope` sets or widens a card with an entry that is not in the card's current scope and overlaps an entry of `shared_files`, the `card scope` command shall refuse, name the entry, the shared entry and the line-lease form, and write nothing · failure: none — an entry the card's current scope already holds passes, so a held card can still be widened by other paths.
- REQ-6: When `card check` finds a path that the reservation does not cover and that overlaps an entry of `shared_files`, the `card check` command shall end its existing refusal with the line-lease form for that path · failure: none — the refusal itself and its prefix `reservation conflict in write scope` are unchanged.
- REQ-7: While `shared_files` lists a file, `card line-lease` shall give that file to a card for the length of its command and take it back, so that two cards lease the same file one after the other and neither card's scope ever lists it · failure: none — the behavior exists; this requirement pins it.
- REQ-8: Where the profile has no `shared_files` key or cannot be read, the `card take`, `card scope` and `card check` commands shall behave as they did before this SPEC · failure: none.
- REQ-9: Where the project is this repository, the profile shall list `src/lib.rs`, `src/help.rs`, `src/ops/card.rs`, `tests/areas.toml`, `tools/repo_scans.json`, `skills/moai/references/source-manifest.json`, `docs/design-review/appendix/operations/card.md` and `tests/fixtures/help/*.txt` · failure: none — a test reads the list from `.moai/project.yaml`.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-8): a profile with `shared_files: [src/lib.rs]` reads and keeps the key through a write; a profile with no key reads as before; a profile whose `shared_files` is a string, holds an absolute path, a comma, a line break or 65 entries is refused with the existing invalid-profile message and no value in it · verify: `cargo nextest run --locked --offline --lib profile::tests::shared_files`
- AC-2 (REQ-3): in a temporary project whose profile lists `src/lib.rs`, `take --scope src/lib.rs`, `take --scope src/**` and a take of a card whose own scope holds `src/lib.rs` each refuse with a message that holds the entry, the shared entry and `card line-lease`, and the ledger has no reservation and no event afterwards; `take --scope src/other.rs` is taken · verify: `cargo nextest run --locked --offline --test card_take_scope_cli shared_files_refuse_a_given_scope`
- AC-3 (REQ-4): a card whose SPEC `write:` list names `src/other.rs` and `src/lib.rs` is taken with `src/other.rs` only, `warnings` names `src/lib.rs` and the line-lease form, and `scope_source` is `spec`; the same for a last reservation; a SPEC list holding only `src/lib.rs` refuses and writes nothing · verify: `cargo nextest run --locked --offline --test card_take_scope_cli shared_files_are_left_out_of_a_fill`
- AC-4 (REQ-5): `card scope` on a `ready` card refuses `--scope src/help.rs`; on a held card, widening with `src/help.rs` refuses and widening with `src/other.rs` works; a held card whose scope already holds `src/lib.rs` is widened with `src/other.rs` and its `src/lib.rs` stays · verify: `cargo nextest run --locked --offline --test card_take_scope_cli shared_files_refuse_a_scope_change`
- AC-5 (REQ-6): `card check` of `src/lib.rs` on a held card that does not hold it refuses with the existing prefix and ends with `card line-lease C-n --generation g --path src/lib.rs`; the same check of `src/other.rs` ends as it does today · verify: `cargo nextest run --locked --offline --test card_take_scope_cli shared_files_check_names_line_lease`
- AC-6 (REQ-7): two cards, each held without `src/lib.rs` in its scope, each run `card line-lease --path src/lib.rs -- <command>` one after the other with the profile listing `src/lib.rs`; both commands run, and after each the card's reservation scope is what it was · verify: `cargo nextest run --locked --offline --test card_line_lease_cli shared_files_two_cards_lease_one_after_the_other`
- AC-7 (REQ-8): with no `shared_files` key and with a profile that cannot be read, `take --scope src/lib.rs`, `card scope` and `card check` answer as they did before · verify: `cargo nextest run --locked --offline --test card_take_scope_cli shared_files_absent_changes_nothing`
- AC-8 (REQ-9): this repository's `.moai/project.yaml` lists the eight entries of REQ-9, and each overlaps itself under `scope::overlap` · verify: `cargo nextest run --locked --offline --lib card::take_scope::tests::this_repository_lists_its_shared_files`

## Scope

- write: `.moai/specs/SPEC-20261124-shared-files-profile/spec.md`, `.moai/project.yaml`, `src/card/take_scope.rs`, `src/card/dispatch.rs`, `tests/card_take_scope_cli.rs`, `tests/card_line_lease_cli.rs`
- held whole by card C-444 on 2026-11-24 (lease until 04:24:04Z, read with `moai card show C-444`) and needed by this change: `src/profile.rs` (the `shared_files` row, the field, `validate` and its tests) and `src/card/work.rs` (the hint of `check`). They are written only after C-444 gives them back or the leader names another way (D-9); this SPEC does not widen into them. `src/card/mod.rs`, which C-444 also holds, is not touched (the helper lives in `src/card/take_scope.rs`).
- shared files by line lease, never in the scope: `src/ops/card.rs` (one sentence each for `card take`, `card scope` and `card check` help), `docs/design-review/appendix/operations/card.md` (the `shared_files` key and the three refusals), `tests/areas.toml` and `tools/repo_scans.json` (none expected: no new test file, no new top-level module)
- read-only: `src/scope.rs`, `src/card/take_scope.rs`, `src/ledger/scope_return.rs`, `src/ledger/workflow.rs`, `CLAUDE.md`
- forbidden: `.github/**`, `src/ledger/schema.rs`, `src/ledger/migrate.rs`, any skill or agent text

## Plan

1. RED: add the tests of AC-1 to AC-8 to `src/profile.rs` (inline), `src/card/take_scope.rs` (inline, the repository list), `tests/card_take_scope_cli.rs` and `tests/card_line_lease_cli.rs`. Each fails on the current code. The RED commit goes first on its own.
2. GREEN, profile: the `shared_files` row in `PROFILE_KEYS`, the `Profile` field, `validate` (same limits as `card add --scope`, at most 64), and the additive-keys test of `src/profile.rs`.
3. GREEN, helper: `src/card/take_scope.rs` gets `shared_files(project)` (the list, empty when the key is absent or the profile cannot be read) and `shared_overlap(list, entry)` (the shared entry an entry overlaps); no new module, so `src/card/mod.rs` is not touched.
4. GREEN, commands: `take` checks a given scope and the card's own scope before the ledger is opened for write, and `take_scope::choose` output is filtered with the warning; `scope` checks the added entries (inside the transaction for a held card, where the current scope is known); `check` appends the line-lease form to its existing refusal.
5. This repository's `.moai/project.yaml` gets the eight entries; run `moai agents-plan` to see that the managed section of `CLAUDE.md` does not change (a change goes to the leader, `CLAUDE.md` is not in this scope).
6. Line-lease edits (this SPEC practices what it adds): the help sentences of `src/ops/card.rs` and the key and refusals in `operations/card.md`, each through `moai card line-lease` with the lane's own generation.
7. Guard mutants, one per new guard test: skip the overlap test in `take`; skip it in `scope`; check only a held card's added entries as all entries; drop the left-out filter of a fill; append no hint in `check`; treat an unreadable profile as a refusal; let `validate` accept 65 entries. Each must fail a named test.
8. SYNC: Result with each AC met or not met; READY note with the `check affected` receipt.

## Risks

- A blocking refusal begins in enforce (D-4, rule 11 of `sdd.md` read as: the list is opt-in and its false positives are visible at once, because every refusal names the entry and the shared file it overlaps). The Result records any refusal seen while this SPEC's own branch was built.
- A broad glob in a card scope (`src/**`) now counts as holding `src/lib.rs`. Cards with such a scope are refused at take until the scope names its files; the refusal says which shared entry overlaps. No card in the ledger is read or changed by this SPEC.
- `scope::overlap` folds case; a case-insensitive overlap is a refusal that looks too wide on a case-sensitive disk. The refusal names both entries, so it is visible.
- A SPEC `write:` list that names only shared files can no longer fill a scope; its card needs `--scope` for the other paths or a SPEC list that names them (the refusal says so).
- The profile is read on every take, scope and check: one small file read, no network.
- `.moai/project.yaml` feeds the managed section of `CLAUDE.md`; step 5 reads that it does not change before the change is committed.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The key is `shared_files`, an optional list of path or glob strings in the project profile | the card names a profile key; the table of keys in `src/profile.rs` is the one place profile keys are declared | rename the key in a later SPEC |
| D-2 | An entry is refused when `scope::overlap` says it overlaps a shared entry, either way round | the code already has the one overlap rule that reservations use; a glob that covers a shared file holds it whole just as a literal does | narrow to literal matches in a later SPEC |
| D-3 | A given `--scope` and a card's own scope are refused; a scope filled from a SPEC `write:` list or the last reservation leaves the shared entries out and warns | many SPEC lists name shared files as documentation; refusing them would make those SPECs untakeable | refuse every source |
| D-4 | The refusals are enforce from the first version, not observe | the list is explicit and opt-in, a missing key changes nothing, and the refusal points at the existing line-lease; rule 11 of `sdd.md` asks for the observation before nobody is stopped by noise, and a refusal of a named file cannot be noise. The leader decides; this is a deviation from the default | add an observe state to the key in a later SPEC |
| D-5 | `card check` keeps its error shape (`reservation conflict in write scope`) and only gets the hint; no `allowed: false` answer is added | the code never answers `allowed: false`, it refuses; a second answer shape would change every caller of `card check` | add the shape in a later SPEC |
| D-6 | `card scope` refuses only entries that the card's current scope does not already hold | a held card's scope only grows and must keep every pattern; a card already holding `src/lib.rs` must still be widened by other paths | refuse every entry |
| D-7 | No exception flag; a card that must rewrite a whole shared file lists the file's removal from `shared_files` as its own reviewed change | an exception flag is the hole the card closes; the list is in the repository and reviewed | add a flag in a later SPEC |
| D-8 | The first list is the eight entries the leader registered by hand on 2026-11-24 | the dispatch of this card | edit the profile list |
| D-9 | `src/profile.rs` and `src/card/work.rs` are written after C-444 gives them back (`card give-back`), or by the way the leader names; this SPEC does not widen into them and a line lease does not open them, because a lease is refused while another card holds the whole file | both files are needed (the profile key; the `check` hint) and neither can move; this is the situation the SPEC exists to prevent, and it arose before the SPEC could land. Default: wait for C-444 to give the two paths back, work the other parts first | the leader asks C-444 for the give-back |

## Result

(not built yet)
