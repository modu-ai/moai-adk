---
id: SPEC-20261120-property-tests-wave2
title: "property tests, second wave: the audit verdict line, the commit trailers and the card export round-trip"
status: draft
revision: 1
created: 2026-11-21
---

# property tests, second wave: verdict line, trailers, card export

## Intent

카드 C-172 의 첫 번째 갈래예요. 병합 게이트가 읽는 줄(감사 판정 줄), 감사 독립성을 가르는 줄(커밋 트레일러), 되돌려 읽혀야 하는 파일(카드 내보내기)에는 짝이 되는 쓰기와 읽기가 있는데 예시 시험만 있어요. 첫 파도의 시드 생성기(`tests/common/prop.rs`, 새 의존성 없음)로 세 곳에 속성 시험을 걸어요: 쓰고 읽으면 같은 값, 읽는 쪽 둘은 한 줄을 같이 읽지 않음, 순서를 바꿔도 같은 바이트, 망가진 입력은 조용히 받아들여지지 않음. 속성마다 1000 사례 이상, 일부러 넣은 결함을 잡는 카나리 하나씩.

## Non-goals

- 새 크레이트 의존성(proptest, quickcheck).
- `src/`, `Cargo.toml`, `skills/**` 변경. 시험 파일과 이 SPEC 만 써요; 위반이 발견되면 고치지 않고 카드에 재현과 함께 보고해요(다른 카드가 `src/gate.rs`, `src/trailers.rs` 를 예약 중).
- 카드의 나머지 갈래: `test-design.md` 에 속성 시험 네 종류를 이름 짓는 문단(배포되는 스킬 파일이라 소스 목록이 필요하고 그 파일은 예약 중), 재감사 수렴 규칙(`review-check --previous`, `class`/`instances` 칸, `gate-check` 의 FAIL 수, `mode-agent.md`). 둘은 리더가 후속 카드로 나눠요.
- 실제 HOME·실제 원장·네트워크 사용. 세 대상은 모두 순수 함수라 임시 폴더도 필요 없어요.

## Requirements

- REQ-1: When a generated pull request number, a one-line title, a 40-hex head, a 40-hex base and a verdict of PASS or FAIL are formatted by `gate::format_verdict_line`, the formatter shall return a line that `gate::parse_verdict` reads back as exactly that head, base and verdict and that `gate::verdict_line` names as a whole verdict line.
- REQ-2 (REQ-1): IF the formatter refused a valid input, or the parser read back another head, base or verdict, THEN the property shall fail naming the seed and the case.
- REQ-3: When a generated input has a head that is not 40 lowercase hex, a base that is not, a verdict other than PASS or FAIL, or a title that is empty or holds a line break, `gate::format_verdict_line` shall refuse it.
- REQ-4 (REQ-3): IF the formatter returned a line for such an input, THEN the property shall fail naming the seed and the case.
- REQ-5: When a verdict line carries a base, the two readers shall give exactly one reading between them: `gate::parse_verdict` for a 40-hex base and `gate::parse_verdict_short_base` for a 7 to 39 hex base, each as exactly its head, base and verdict, and the other reader none, so no line is read by both.
- REQ-6 (REQ-5): IF both readers accepted one line, or the reader that does not own the line accepted it, THEN the property shall fail naming the seed and the case.
- REQ-7: When one character of a valid verdict line's head or base is replaced by a character that is not lowercase hex, or one character of the head or base is dropped, or a word is appended after the verdict, or the verdict separator is changed, `gate::parse_verdict` shall read none of it.
- REQ-8 (REQ-7): IF the parser read a changed line, THEN the property shall fail naming the seed and the case.
- REQ-9: When a comment body holds a verdict line, `gate::verdict_line` shall return that line only when it is the first line of the body, so a verdict line further down is never returned.
- REQ-10 (REQ-9): IF a verdict line further down was read as the verdict, or the first line was not returned, THEN the property shall fail naming the seed and the case.
- REQ-11: When generated commits are written in the `git log` record format and read by `trailers::parse_log`, the reader shall return every commit unchanged (sha, SPEC ids, session values, verified values, subject, body) with none marked garbled.
- REQ-12 (REQ-11): IF a commit read back changed, was lost, or was marked garbled, THEN the property shall fail naming the seed and the case.
- REQ-13: When `trailers::format_verified` is given a new-form SPEC id, a session id a host variable carries, a worktree key and a set of verified kinds, the trailer lines it prints shall be read by `trailers::parse_log`, once written into one log record, as that one SPEC id, the value `<session id>:<key>` when the session is known, and those kinds, with no failing finding from `trailers::judge` on a `spec/` branch.
- REQ-14 (REQ-13): IF the printed lines read back as other values, or the judge named a failing finding for them, THEN the property shall fail naming the seed and the case.
- REQ-15: When a raw record separator or field separator is placed in the middle of a generated record, the commits `trailers::parse_log` returns shall be either the original commits or hold at least one commit marked garbled, so a wrong reading is never silent.
- REQ-16 (REQ-15): IF the parsed commits differed from the original and none was marked garbled, THEN the property shall fail naming the seed and the case.
- REQ-17: When a generated ledger snapshot (cards with their criteria, notes and links, and events; unique ids; text with line breaks, quotes, marker lines and non-ASCII) is rendered and parsed, `card::export_format::parse(render(s))` shall equal `card::export_format::canonical(s)` and `canonical` applied twice gives the same rows as applied once.
- REQ-18 (REQ-17): IF the parsed rows differed from the canonical rows, THEN the property shall fail naming the seed and the case.
- REQ-19: When the rows of a snapshot are put in any order, `card::export_format::render` shall write the same bytes.
- REQ-20 (REQ-19): IF two orders of one snapshot rendered differently, THEN the property shall fail naming the seed and the case.
- REQ-21: When one character of any file of a rendered export (the index, a card file or the events file) is replaced by a different character, `card::export_format::parse` shall refuse the export, or read it with every card, criterion, note, link and event row kept and render it back to exactly the edited bytes; only the project id and the schema version, which the index alone holds with no digest, may differ.
- REQ-22 (REQ-21): IF an edited export was accepted and a row changed or the rows rendered to other bytes than were read, THEN the property shall fail naming the seed and the case.
- REQ-23: Each suite shall hold a canary test that runs its checkers on a deliberately broken variant of the code under test and requires the checkers to fail on at least one case, so a checker that cannot fail is reported.
- REQ-24 (REQ-23): IF a checker accepted every case of its broken variant, THEN the canary test shall fail.

## Acceptance

- AC-1 (REQ-1…REQ-10): the verdict line properties run at least 1000 cases each and pass; each checker fails on its broken variant.
  verify: `cargo test --locked --offline --test prop_verdict_line`
- AC-2 (REQ-11…REQ-16): the trailer properties run at least 1000 cases each and pass; a separator injected in the middle of a record never gives a silent wrong reading; each checker fails on its broken variant.
  verify: `cargo test --locked --offline --test prop_trailers`
- AC-3 (REQ-17…REQ-22): the card export properties run at least 1000 cases each and pass; each checker fails on its broken variant.
  verify: `cargo test --locked --offline --test prop_card_export`
- AC-4 (REQ-23, REQ-24): every canary test passes, meaning every checker failed on its broken variant.
  verify: `cargo test --locked --offline --test prop_verdict_line --test prop_trailers --test prop_card_export canary`
- AC-5: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `tests/prop_verdict_line.rs`, `tests/prop_trailers.rs`, `tests/prop_card_export.rs`, `.moai/specs/SPEC-20261120-property-tests-wave2/**`
- read-only: `tests/common/prop.rs` (included with `#[path]`), `src/gate.rs`, `src/trailers.rs`, `src/session_current.rs`, `src/spec_doc.rs`, `src/card/export_format.rs`
- forbidden: `src/**`, `Cargo.toml`, `skills/**`, the real HOME, the network, new dependencies

## Plan

1. SPEC (this file), then one suite at a time: each checker returns a result so a canary can feed it a broken variant.
2. `tests/prop_verdict_line.rs`: round trip, refusals, reader disjointness, one-character mutations, first line only; one canary.
3. `tests/prop_trailers.rs`: log round trip, `format_verified` read back through `parse_log` and `judge`, separator injection; one canary.
4. `tests/prop_card_export.rs`: render/parse round trip, row order, one-character edit; one canary.
5. Targeted checks, then a report of every violation a property finds.

## Risks

- A property that fails on the real code is a finding, not a bug of the test: it is reported on the card with a minimal replay (`PROP_SEED`, `PROP_ONLY`) and, if it cannot be fixed within this scope, kept as an ignored test whose reason names the replay.
- A generator that draws inputs the code refuses on purpose would give false failures: refusals are their own property (REQ-3), and the other generators draw only valid input.
- A separator at the very end of a message leaves an empty tail that is dropped without a trace; REQ-15 injects only in the middle of a record and the end case is recorded as an observation.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Only the first half of the card is built here. | The other half needs files other cards hold (`src/gate.rs`, `src/review.rs`, `mode-agent.md`, the source list). | The leader splits a follow-up card. |
| D-2 | Checkers take the function under test as a parameter. | A canary needs to feed the same checker a broken variant. | Hard-wire the real functions. |
| D-3 | Canaries count failures over 1000 cases and require at least one. | A rare mutation (one case in twenty) still has to be seen. | Fixed-size canary. |
| D-4 | The hex-tail injection test stays ignored and AC-2 stays open (`partly`). | The real reader violates REQ-15 for that input and `src/trailers.rs` is outside this SPEC's scope; hiding it by weakening the property would be a false green. | When the reader marks such a record garbled, remove the `#[ignore]` and mark AC-2 `met`. |

## Result

- AC-1: met — `cargo test --locked --offline --test prop_verdict_line` exit 0 (6 passed: 5 properties at 1000 cases each plus the canary; run7.log). The canary requires a failing case from each of five broken variants (swapped head/base, unchecked formatter, both-readers parser, lowercasing parser, any-line searcher).
- AC-2: partly — `cargo test --locked --offline --test prop_trailers` exit 0 (4 passed, 1 ignored; run7.log). The separator followed by words (`WORD_TAILS`) is never read silently. The separator followed by hex-looking text is read silently: `cargo test --locked --offline --test prop_trailers -- --ignored` exit 101 at case 6, replay `PROP_SEED=13263398375705 PROP_ONLY=6` (run6.log). That test stays `#[ignore]` with the replay in its reason; the fix is in `src/trailers.rs`, which this SPEC may not touch.
- AC-3: met — `cargo test --locked --offline --test prop_card_export` exit 0 (4 passed; run7.log). The first run failed `an_export_with_one_character_changed_is_refused` because REQ-21 was written too wide: the project id lives in the index alone with no digest, so editing it gives another valid export. REQ-21 and REQ-22 were narrowed to "refused, or every row kept and the bytes render back unchanged".
- AC-4: met — `cargo test --locked --offline --test prop_verdict_line --test prop_trailers --test prop_card_export canary` exit 0 (three canary tests, 1 passed each; run8.log).
- AC-5: met — `cargo fmt --check` exit 0; `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0 (clippy-all.log).

Not run: the full test suite of the repository (only the three suites above, by the disk and test caps).
