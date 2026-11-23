---
id: SPEC-20261123-land-repo-scans
title: "card land runs the repo-wide scan tests that the merge tree needs, and refuses a landing without their receipt"
status: draft
revision: 1
created: 2026-11-23
---

# card land runs the repo-wide scan tests that the merge tree needs, and refuses a landing without their receipt

## Intent

저장소 글자 전체를 읽는 테스트(`tools/repo_scans.json` 의 스캔)는 바뀐 파일만 고르는 감사와 착지에서 빠지기 쉬워요. 같은 종류의 사고가 하루에 세 번 났어요.

- 새 테스트 파일 `branch_gc_cli` 의 첫 낱말이 `tests/areas.toml` 에 없어서 착지 뒤에 `test_area_layout` 이 빨개졌어요(팁 수리 `bdf0d857b6`). 감사 브리프가 그 스캔을 뺐어요.
- `tests/fake_exe_marker.rs` 가 환경을 직접 비워서 착지 뒤에 `cargo_env_developer_dir` 의 `clean_env` 스캔이 빨개졌어요. 감사가 그 스캔을 돌리지 않았어요.
- `green_test_lock::every_detected_language_has_test_patterns` 가 `include_str!("../src/project_detect.rs")` 로 소스 표를 읽는데, 다른 카드가 `project_detect` 에 `lua` 를 더할 때 그 테스트가 선택되지 않았어요.

`moai repo-scans` 는 head 와 팁의 변경을 비교해서 걸리는 스캔을 이미 계산해요(규칙 파일이 한쪽에서, 범위 경로가 다른 쪽에서 바뀐 경우). 그러나 (1) 세 사고 중 둘은 규칙 파일이 아무 쪽에서도 안 바뀌었고 범위 경로만 head 쪽에서 바뀐 경우라 이 판정에 걸리지 않고, (2) 착지 명령은 이 판정을 부르지 않아요. 착지 명령이 머지 트리(착지하는 head 가 곧 머지 트리예요: `card land` 는 팁이 head 의 첫 부모 줄에 있을 때만 움직여요)에 대해 걸리는 스캔을 계산하고, 그 스캔이 같은 head 에서 통과했다는 영수증이 없으면 막고 돌릴 명령을 한 줄로 알려 줘요.

이 SPEC 은 세 가지를 바꿔요.

1. 표의 스캔에 `"always": true` 를 달 수 있게 하고, 그런 스캔은 head 쪽 변경이 범위 안에 있으면 규칙 파일이 안 바뀌어도 걸려요. `test-area-layout` 과 `clean-env-only` 에 달고, `include_str!` 로 저장소 파일을 읽는 테스트를 표에 넣어요.
2. `moai card land` 가 움직이는 착지에서 걸리는 스캔을 계산하고, 영수증(`card verdict` 의 `--receipt-step scan:<id>=0`)을 읽어요. 처음에는 관찰 모드이고 기존 `work_flow.land_verdict` 의 모드를 그대로 따라요.
3. `next-flow` 의 착지 안내가 감사 브리프에 "머지 트리에서 `moai repo-scans` 가 고른 테스트"를 넣게 말해요.

### 읽은 것 (2026-11-23T07:0xZ, 팁 `c395e5a50c`)

- 카드 본문과 리더 노트 두 개(7360, 7363).
- `src/repo_scans.rs`, `tools/repo_scans.json`, `tests/repo_scans_table.rs`: 판정 `judge`, 표의 항목 모양, 표 완전성 테스트.
- `src/card/land.rs`, `src/card/land/check.rs`: 착지 순서, 관찰/강제 모드, 영수증(`VerdictReceipt`, 단계 이름과 종료 코드).
- `src/check_affected.rs`: `check affected` 의 영수증은 `test` 한 단계에 선택된 타깃을 모두 묶어서, 어느 스캔이 돌았는지 영수증만으로는 알 수 없어요(그래서 이 SPEC 은 `scan:<id>` 단계 이름을 따로 요구해요).

## Non-goals

- 게이트를 느슨하게 하는 일, 전체 `cargo test` 를 강제하는 일, Python 테스트를 돌리는 일.
- `moai bundle check` 를 바꾸는 일: 이 흐름에서 착지는 `card land` 로만 일어나므로 착지 전에 실제로 도는 쪽은 `card land` 예요(D-1).
- 스캔을 자동으로 돌려 영수증을 대신 쓰는 명령: 영수증은 감사 세션이 `card verdict` 로 적어요. 자동 기록은 다른 카드예요.
- `repo-scans` 의 `remerge` 뜻을 바꾸는 일: 규칙 파일과 범위 경로가 양쪽에서 갈린 경우만 `remerge` 예요.
- 소스를 `read_to_string` 으로 읽는 숨은 스캔 찾기: `include_str!` 로 읽는 것만 이 SPEC 의 범위예요.

## Requirements

<!-- This SPEC adds a check that blocks: it starts in observe (REQ-6). Record the would-block count and the false positives in the Result, and leave the switch to enforce to a recorded decision. -->
- REQ-1: When a scan of the table carries `"always": true` and a path of the head's change set is in the scan's scope, the repo-scans judge shall list that scan in `own_hits` and its checks in `checks` · failure: none — a scan without the flag is judged as before.
- REQ-2 (REQ-1): IF a scan is listed only in `own_hits`, THEN the repo-scans judge shall leave `remerge` as the other hits alone set it.
- REQ-3: When the card land command is about to move the integration branch, the card land command shall compute the scans the merge tree needs from the table at the tip and at the head, for the range from the tip to the head and for each merge commit of the first-parent line of that range with its two parents · failure: a table that cannot be read is REQ-8.
- REQ-4: When a needed scan has a step named `scan:<id>` at exit 0 in the receipt of a PASS verdict that names the landing head, the card land command shall count that scan as covered · failure: none — a receipt of another head or a step at another exit code does not cover it.
- REQ-5 (REQ-3): IF a needed scan is not covered and the mode is enforce, THEN the card land command shall refuse with the code `scans_missing`, name each missing scan with its check commands on one line, and write nothing.
- REQ-6: While the mode is observe, the card land command shall answer the refusal it would have made as `scans.would_refuse` and then land as it did before this SPEC · failure: none — the answer still names the needed scans and the covered ones.
- REQ-7: When the project profile names `work_flow.land_verdict`, the card land command shall use that one mode for the verdict check and for the scan check · failure: none — the key, its words and its default `observe` are those of SPEC-20261123-landing-reads-verdict.
- REQ-8 (REQ-3): IF a table exists at the tip or at the head and cannot be read, THEN the card land command shall refuse in enforce mode with the code `scans_not_read` and report that code as `would_refuse` in observe mode.
- REQ-9: When the head is already inside the integration tip, the card land command shall skip the scan computation and close as before · failure: none — a rerun after a stop between the move and the close is the same call.
- REQ-10: The table shall carry `"always": true` on `test-area-layout` and `clean-env-only`, so that a head that adds or edits a test file needs both scans · failure: none — checked by AC-1 and AC-2.
- REQ-11: The table shall list, with the read file in the scope of the scan, every integration test file that reads a repository file outside `tests/` with `include_str!` · failure: none — REQ-14 makes a missing entry fail.
- REQ-12: The next-flow reference shall tell the session that briefs an audit to name the tests `moai repo-scans --refs <tip> <head>` chose, to run them on the merge tree, and to record each as a receipt step `scan:<id>` · failure: none — a test pins the sentence.
- REQ-13: When neither the tip nor the head holds a table, the card land command shall need no scan · failure: none — a repository without a table has no scan to run.
- REQ-14: When an integration test file reads a repository file outside `tests/` with `include_str!` and no scan lists that test file, the table test shall fail and name the file · failure: none — the exemption list names each file it keeps out with its reason.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5, REQ-10): a head that adds a test file whose first word is not in `tests/areas.toml` refuses in enforce mode with `scans_missing` that names `test-area-layout` and its check command; the ledger and the branch are unchanged and no start row exists · verify: `cargo test --locked --offline --test card_land_scans_cli ac1_`
- AC-2 (REQ-1, REQ-10): a head that adds a test file that clears the environment itself needs `clean-env-only` as well, so the refusal names both scans, each with its own command · verify: `cargo test --locked --offline --test card_land_scans_cli ac2_`
- AC-3 (REQ-4, REQ-5): a PASS whose receipt names the landing head with `scan:test-area-layout` and `scan:clean-env-only` at exit 0 lets the landing move and close; a step at exit 1, a receipt of another head and a receipt missing one scan each refuse with `scans_missing` · verify: `cargo test --locked --offline --test card_land_scans_cli ac3_`
- AC-4 (REQ-3): a lane that changes a rule file while the tip changed a path in that scan's scope is a cross hit and needs the scan, and a docs-only head needs none and lands in enforce mode with no scan step · verify: `cargo test --locked --offline --test card_land_scans_cli ac4_`
- AC-5 (REQ-6, REQ-7): with no `land_verdict` key and with `observe` the landing moves, the answer carries `scans.would_refuse` with the code `scans_missing` and the missing scans, and the ledger shows one landing · verify: `cargo test --locked --offline --test card_land_scans_cli ac5_`
- AC-6 (REQ-8, REQ-13): a table that is not JSON at the head refuses in enforce mode with `scans_not_read`, is a `would_refuse` in observe mode, and a repository with no table needs no scan · verify: `cargo test --locked --offline --test card_land_scans_cli ac6_`
- AC-7 (REQ-9): a head already inside the tip closes without computing scans, even with a table that cannot be read · verify: `cargo test --locked --offline --test card_land_scans_cli ac7_`
- AC-8 (REQ-1, REQ-2): the judge lists an `always` scan in `own_hits` without a rule change, leaves `remerge` false for it, and lists a scan without the flag as before; `moai repo-scans --refs` prints the same · verify: `cargo test --locked --offline --lib repo_scans::` and `cargo test --locked --offline --test repo_scans_cli always_`
- AC-9 (REQ-10, REQ-11, REQ-14): the table test finds `always` on the two scans, finds an entry for every test file that includes a repository file with `include_str!`, among them `green_test_lock` with `src/project_detect.rs` in its scope, and fails on a fixture test file that has none · verify: `cargo test --locked --offline --test repo_scans_table`
- AC-10 (REQ-12): the next-flow reference holds the sentence that names `moai repo-scans`, the merge tree and `scan:<id>`, and a test pins it · verify: `cargo test --locked --offline --test next_flow_skill_scan land_scans`

## Scope

- write: `.moai/specs/SPEC-20261123-land-repo-scans/**`, `src/repo_scans.rs`, `src/ops/repo_scans.rs`, `src/card/land.rs`, `src/card/land/scans.rs`, `tools/repo_scans.json`, `tests/card_land_scans_cli.rs`, `tests/repo_scans_table.rs`, `tests/repo_scans_cli.rs`, `docs/design-review/appendix/operations/repo_scans.md`
- write after another card releases them: `skills/moai/references/flows/next-flow.md`, `skills/moai/references/source-manifest.json` (hash only), `tests/next_flow_skill_scan.rs`
- read-only: `src/card/land/check.rs`, `src/ledger/workflow.rs`, `src/check_affected.rs`, `tests/card_land_cli.rs`, `tests/areas.toml`
- forbidden: `src/bundle.rs`, `src/profile.rs`, any other path, the real HOME

## Plan

1. SPEC (this file), then `moai spec check`.
2. RED: `tests/card_land_scans_cli.rs` (AC-1 to AC-7), the judge cases (AC-8) and the table cases (AC-9); each fails on the unchanged tree.
3. GREEN, judge: the `always` field and `own_hits` in `src/repo_scans.rs` and its printed form.
4. GREEN, table: `always` on the two scans and the `include_str!` readers in `tools/repo_scans.json`.
5. GREEN, landing: `src/card/land/scans.rs` and the call in `src/card/land.rs`.
6. Text: the next-flow sentence and its pin (AC-10) once the other card releases the file.
7. Evidence: the new tests, `nextest` with a filterset over the touched targets, clippy, fmt, `git diff --check`, `moai trailer-check`, one bare `check affected`; a guard mutant per new guard test.

## Risks

- 모든 `tests/` 편집이 두 스캔을 요구하게 돼요: 두 스캔은 단일 테스트 한 개씩이라 가볍고, 관찰 모드에서는 착지를 막지 않아요. 막는 횟수와 거짓 양성은 Result 에 적어요.
- 영수증 단계 이름이 사람이 적는 글자라서 틀릴 수 있어요: 거부 문장이 정확한 단계 이름(`scan:<id>`)과 명령을 알려 줘요.
- `include_str!` 표가 새 테스트 파일에서 낡을 수 있어요: 표 테스트가 표에 없는 파일을 실패로 만들어요.
- `next-flow.md` 는 다른 카드가 쥐고 있어요: 그 카드가 놓을 때까지 코드와 표를 먼저 끝내고, 문장은 마지막에 넣어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The scan check lives in `card land`, not in `bundle check` | the NeXT flow moves the integration branch only through `moai card land`; the leader note 7360 names `src/bundle_check.rs` that does not exist and the card allows "the one that really runs before landing" | add the same call to `bundle check` in another card |
| D-2 | The mode is the existing `work_flow.land_verdict`; no new profile key | one switch for the whole landing evidence check keeps `src/profile.rs` and the profile golden files untouched | add a `land_scans` key in another card |
| D-3 | A receipt step named `scan:<id>` at exit 0 covers a scan | `check affected` records one `test` step for all targets, so the receipt cannot say which scan ran; a name the refusal prints is checkable | accept another step name in the same function |
| D-4 | The `always` flag is a table field, `own_hits` is a new report key, `remerge` keeps its meaning | existing leaders read `remerge` to decide a re-merge; the flag must not turn every tests/ edit into one | fold `own_hits` into `hits` in another card |
| D-5 | Only `include_str!` readers are added to the table now | the third incident is of that kind; hidden `read_to_string` readers need their own scan | list them in a follow-up card |
