---
id: SPEC-20261119-generated-files-no-lock
title: "generated shared files no longer lock card reservations (C-118)"
status: draft
revision: 1
created: 2026-11-19
---

# generated shared files no longer lock card reservations (C-118)

## Intent

카드 C-118 (p1): `skills/moai/references/source-manifest.json` 은 `skills/moai/**` 파일마다 sha256 한 줄을 적는 **생성 파일**이에요(`tools/source_manifest.py write` 가 쓰고, 병합 때는 `tools/merge_driver_manifest.py` 가 세 갈래로 합쳐요). 그래서 skill 파일을 하나라도 바꾸는 카드는 쓰기 범위에 이 파일을 넣어야 하고, `card take` 는 범위가 겹치는 카드가 둘 이상이면 뒤에 오는 쪽을 막아요.

관찰된 결함(2026-11-19): C-59(agent-1)의 `project.md` 문구 수리와 C-103(agent-5)의 manifest 수리가 서로를 기다렸어요. C-59 가 manifest 를 범위에 넣으려 하니 `card take` 가 C-103 과의 겹침으로 거부했고(카드 노트 2876), 반대로 C-103 의 take 는 C-59 의 예약 때문에 거부됐어요(리더 요청으로 C-59 를 release). 한쪽이 한도로 멈추자(agent-5) 다른 쪽의 한 줄 수리까지 멈췄어요. 겹치는 것은 "둘 다 같은 한 줄 파일을 다시 생성한다"는 사실뿐이고, 두 카드가 다른 skill 파일을 고치는 한 내용은 충돌하지 않아요.

해법의 모양: 생성 파일 목록(`GENERATED_SHARED_PATHS`)을 코드에 두고, 카드 쓰기 범위의 겹침 판단에서 **그 파일만 빠져요**. 카드는 여전히 그 파일을 범위에 적어야 쓸 수 있고(`card check --path`), 진짜 겹침(같은 skill 파일, 같은 폴더 패턴)은 그대로 막혀요. 병합의 정확성은 이미 있는 두 장치가 지켜요: 병합 드라이버와 `source_manifest.py check`(+ `product_cli skill_package` 의 `source_manifest_covers_every_package_file`).

## Non-goals

- `skills/**` 문구 갱신(예: `verify.md` 78행의 "충돌 뒤 manifest" 규칙에 "예약 겹침" 안내 추가) — skill 파일을 바꾸면 manifest 를 다시 써야 해서 바로 이 문제의 한가운데에 있어요. 이 카드의 다음 조각(조각 3)으로 미뤄요.
- 새 파일 종류를 자동으로 "생성 파일"로 알아내기 — 목록은 손으로 관리하는 짧은 상수예요(지금은 한 줄). 저장소 파일(`.gitattributes` 등)을 `take` 때 읽지 않아요.
- 클레임(`moai-cli claim`/`scope check`) 쪽의 겹침 규칙과 `SHARED_INTERFACE_PATHS` — 바꾸지 않아요.
- 병합 드라이버·`source_manifest.py` 자체의 동작 변경, GitHub 쪽 병합(드라이버는 로컬 클론에서만 돌아요).
- 실제 HOME, 네트워크, 원장 수편, 기존 예약의 소급 변경.

## Requirements

- REQ-1: When two cards' write scopes are compared for overlap, the card ledger shall ignore a pair in which one pattern is exactly a path listed in `GENERATED_SHARED_PATHS` (compared with the same case folding as `scope::overlap`) and report every other overlapping pair as before.
- REQ-2 (REQ-1): IF any other pair of the two scopes overlaps, THEN `card take` shall refuse the take for that pair, naming that pair and not the ignored one.
- REQ-3: The card ledger shall compare scopes through one predicate that both `card take` (`reserve`, `reserve_own_scope`) and the dispatch listing (`ready` selection in `src/ledger/workflow.rs`) call, so a card that `take` accepts is never listed as skipped for an overlap that `take` ignores.
- REQ-4: When `card take` decides an overlap, the card ledger shall use only the compile-time constant `GENERATED_SHARED_PATHS` in `src/scope.rs` (beside `SHARED_INTERFACE_PATHS`) and read no repository file. · failure: none
- REQ-5: A test shall pin that the set of paths the repository's `.gitattributes` gives the merge driver `moai-manifest` equals the set `GENERATED_SHARED_PATHS`, so a second generated file cannot be added to one list alone.
- REQ-6: When a card's scope lists a generated shared path, `card check --path` shall allow a write to that path exactly as for any other listed path. · failure: none
- REQ-7: When a claim is checked, the claim checker shall use `scope::overlap`, `scope::overlaps` and `SHARED_INTERFACE_PATHS` unchanged. · failure: none
- REQ-8: When two branches each change a different skill file and each regenerates the manifest with `source_manifest.py write`, a merge with the registered driver shall produce a manifest that `source_manifest.py check` accepts and that equals the output of `write` run on the merged tree.
- REQ-9 (REQ-8): IF the merged manifest differs from what `write` produces, THEN `source_manifest.py check` shall fail with a non-zero exit so the stale line is caught before the bundle is pushed.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 같은 임시 원장에서 C-A(범위 `skills/moai/workflows/plan.md`, `skills/moai/references/source-manifest.json`)가 예약된 뒤 C-B(범위 `skills/moai/workflows/run.md`, 같은 manifest)의 `take` 가 성공해요. 같은 두 카드에서 둘 다 `skills/moai/workflows/plan.md` 를 넣으면 C-B 의 `take` 는 `plan.md` 쌍을 대며 거부돼요(manifest 쌍이 아니라). 범위가 `skills/moai/**` 인 C-C 는 C-A 의 `plan.md` 와 겹쳐 거부돼요(manifest 쌍 때문이 아니라 `plan.md` 쌍 때문). `ready` 목록은 위 성공 경우에 C-B 를 건너뛴 카드로 적지 않아요.
  verify: `cargo test --locked --offline --test card_generated_scope two_cards_may_both_list_the_generated_manifest`
  verify: `cargo test --locked --offline --test card_generated_scope a_real_overlap_is_still_refused_and_names_its_own_pair`
  verify: `cargo test --locked --offline --test card_generated_scope the_ready_listing_agrees_with_take`
- AC-2 (REQ-4, REQ-5): `GENERATED_SHARED_PATHS` 는 상수이고 `.gitattributes` 의 `merge=moai-manifest` 줄 경로 집합과 같아요. 한쪽만 바꾸면 시험이 실패해요.
  verify: `cargo test --locked --offline --test card_generated_scope the_list_equals_the_gitattributes_merge_lines`
- AC-3 (REQ-6, REQ-7): `card check --path <manifest>` 는 그 파일을 범위에 가진 카드에게 허용하고 범위에 없는 카드에게는 거부해요. 클레임 쪽 겹침 시험(`scope` 모듈)은 수정 없이 통과해요.
  verify: `cargo test --locked --offline --test card_generated_scope card_check_still_requires_the_path_in_scope`
  verify: `cargo test --locked --offline --lib scope`
- AC-4 (REQ-8, REQ-9): 임시 git 저장소에서 두 가지(branch)가 서로 다른 skill 파일을 고치고 각자 `write` 로 manifest 를 다시 쓴 뒤 드라이버로 병합하면 병합 결과의 manifest 가 병합본에서 `write` 한 것과 바이트까지 같고 `check` 가 통과해요. 한쪽 줄을 손으로 틀리게 두고 병합하면 `check` 가 0 이 아닌 코드로 끝나요.
  verify: `python3 -m unittest tests.source_manifest_test -k two_heads_regenerating_their_own_line_merge_to_what_write_produces`
  verify: `python3 -m unittest tests.source_manifest_test -k a_stale_line_after_the_merge_fails_check`
- AC-5: 이웃 검사(대상 한정; 전량은 PR 의 CI run 이 인용).
  verify: `cargo test --locked --offline --test card_generated_scope`
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`(로컬은 대상 한정)

## Scope

- write: `src/scope.rs`, `src/ledger/reservation.rs`(한 줄: `find_overlap` 이 `scope::card_overlap` 을 불러요 — C-43 이 먼저 병합된 뒤 그 head 위에서), `tests/card_generated_scope.rs`, `tests/source_manifest_test.py`, `.moai/specs/SPEC-20261119-generated-files-no-lock/spec.md`
- read-only: `.gitattributes`, `tools/source_manifest.py`, `tools/merge_driver_manifest.py`, `skills/moai/workflows/verify.md`(78행 규칙; 문구 갱신은 조각 3)
- forbidden: 실제 HOME, 네트워크, 기존 예약·원장 행의 수편, `skills/**` 쓰기(이 카드의 조각 1·2), 병합 드라이버·`source_manifest.py` 동작 변경
- 조각별 take: 조각 1 = 이 SPEC 폴더만(지금). 조각 2 = 위 `src/`·`tests/` 파일 — take 때 다른 카드(C-94 등)가 같은 파일을 쥐고 있는지 먼저 `card take` 거부 여부로 확인하고, 막히면 `LANE-BLOCKED kind=decision` 을 카드에 적어요.

## Plan

1. (조각 1, 지금) 이 SPEC — 읽기로 확인한 사실만 쓰고 push 해요(코드 변경 0).
2. (조각 2) RED: `tests/card_generated_scope.rs` 에 AC-1~3 시험을 먼저(임시 원장·임시 HOME 만) — 현 코드에서 첫 시험이 `scope_overlap` 으로 실패해야 하고 그 출력을 기록해요. 그리고 `tests/source_manifest_test.py` 에 AC-4 시험 둘.
3. (조각 2) GREEN: `src/scope.rs` 에 `GENERATED_SHARED_PATHS` 와 카드용 술어 `card_overlap(a, b)`(= `overlap(a, b)` 이고 한쪽이 생성 경로와 정확히 같은 쌍이 아님), `src/ledger/reservation.rs::find_overlap` 가 그 술어를 불러요(`src/ledger/workflow.rs` 의 `ready` 목록은 `find_overlap` 을 그대로 부르니 고치지 않아요). 클레임 쪽 `overlap`·`overlaps` 는 건드리지 않아요.
4. (조각 2) 변이 5개(커밋된 파일에서 하나씩, 모두 시험이 잡아야 해요): K1 예외 조건을 "한쪽만"에서 "양쪽"으로(or→and 뒤집기), K2 알 수 없는 경로 보호 제거, K3 아무것도 무시하지 않음, K4 모든 경로를 생성 경로로 취급(`src/scope.rs`), R1 연결 줄을 옛 `overlap` 호출로 되돌림(`src/ledger/reservation.rs`). `workflow.rs` 는 고치지 않아서 변이 대상이 아니에요.
5. (조각 2) fmt, clippy `--all-targets`, `cargo test --lib` 전체, `home_isolation_scan`, `python3 -m unittest tests.source_manifest_test tests.repo_scans_test`, trailer-check, tip 과 `git merge-tree`, 다른 맥락의 감사, 내 판정 줄, `LANE-READY-INTEGRATION`.
6. (조각 3) `verify.md` 의 manifest 규칙에 "생성 파일은 예약이 겹쳐도 돼요 — 새 tip 을 받은 뒤 `source_manifest.py write` 로 다시 쓰고 sha 줄을 손으로 고치지 않아요"를 한 줄 더해요(skill 파일이라 manifest 를 다시 써야 하니 조각 2 병합 뒤, 다른 skill 카드가 비어 있을 때).

## Risks

- 두 카드가 같은 manifest 를 쓰면 병합 때 sha 줄이 겹칠 수 있어요. 로컬 클론에서 드라이버를 등록(`python3 -I tools/merge_driver_manifest.py --install`)해야 자동으로 합쳐지고, 등록 안 한 클론이나 GitHub 의 병합 버튼은 일반 텍스트 병합이라 충돌할 수 있어요(기억: "병합 드라이버가 GitHub 충돌을 가려요"). 그래서 묶음은 리더가 드라이버가 등록된 클론에서 만들고, 묶음 끝에 `check` 를 돌리는 규칙은 그대로 남아요(`verify.md` 78행).
- 목록이 틀리면(생성이 아닌 파일을 적으면) 진짜 겹침이 가려져요. 그래서 목록은 한 줄로 시작하고, `.gitattributes` 와의 일치 시험(AC-2)이 "병합 드라이버가 있는 파일만" 올라오게 막아요.
- 카드가 manifest 를 범위에 적고도 다른 skill 파일을 쓰는 것은 막지 않아요(범위가 그 파일을 따로 적어야 해요). 이 성질은 AC-1 의 `skills/moai/**` 사례가 보여 줘요.
- 조각 2 가 쓰는 `src/scope.rs`·`src/ledger/*.rs` 는 다른 카드가 쥘 수 있어요. 이 조각 1 은 SPEC 폴더만 가져서 그 위험이 없어요.

## Result

조각 2(코드와 시험) — 2026-11-19, 가지 `spec/20261119-generated-files-no-lock`, C-43 READY head `5622c48473` 과 tip `747738af63` 을 합친 위에서 확인했어요(핵심 판정 `card_overlap` 은 `src/scope.rs`, 연결 한 줄은 `src/ledger/reservation.rs::find_overlap`).

- AC-1: `card_generated_scope` 5/5 통과(`cargo test --locked --offline --test card_generated_scope`). 연결 줄을 되돌린 변이 R1 은 `two_cards_may_both_list_the_generated_manifest`, `the_ready_listing_agrees_with_take`, `a_real_overlap_is_still_refused_and_names_its_own_pair` 세 시험이 잡았어요(2 통과 · 3 실패).
- AC-2: `the_list_equals_the_gitattributes_merge_lines` 통과(위 5/5 에 포함).
- AC-3: `card_check_still_requires_the_path_in_scope` 통과, `--lib` 전체 1458 통과(클레임 쪽 `scope` 시험은 수정 없음).
- AC-4: `python3 -m unittest tests.source_manifest_test` 45 개 OK, `tests.repo_scans_test` 25 개 OK.
- AC-5: `cargo fmt --check` exit 0, `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0.
- 이웃: `ledger_reservation` 28, `card_reclaim_cli` 16, `card_dispatch_cli` 25, `card_cli` 14, `home_isolation_scan` 18 모두 통과.
- 변이 5개(K1–K4 `src/scope.rs`, R1 `src/ledger/reservation.rs`): 5개 모두 시험이 잡았어요.

돌리지 않은 것: 저장소 전체 `cargo test`(PR CI run 이 인용), Linux 전용 검사. 조각 3(`verify.md` 한 줄 안내)은 조각 2 병합 뒤예요.
