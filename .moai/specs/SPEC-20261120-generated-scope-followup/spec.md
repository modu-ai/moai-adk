---
id: SPEC-20261120-generated-scope-followup
title: "Generated shared paths: the docs, the CLI tests, the case rule and the count of the first slice"
status: draft
revision: 1
created: 2026-11-20
---

# Generated shared paths: the docs, the CLI tests, the case rule and the count of the first slice

## Intent

SPEC-20261119-generated-files-no-lock 의 조각 2 는 한쪽이 정확히 생성 파일(`skills/moai/references/source-manifest.json`)인 카드 쌍을 겹침으로 치지 않게 했어요(`scope::card_overlap`). 그 조각의 독립 감사(PASS, 차단 0)가 비차단 넷을 남겼고, 이 SPEC 은 그 넷을 닫아요.

- N1: 예약 모듈 머리말(`src/ledger/reservation.rs`)과 예약 문서(`docs/design-review/appendix/ledger_reservation.md`)가 아직 "범위 겹침 판정은 `overlap` 하나"라고 말해요. 카드 쪽 판정은 이제 `card_overlap`(겹침 + 생성 경로 규칙)이에요.
- N2: manifest 한 줄만 가진 카드를 폴더 전체를 가진 카드 옆에서 잡는 일, 그리고 범위를 넓히는 길(`card scope`)이 실제 바이너리를 거친 시험을 갖지 못했어요(단위 시험과 변이가 간접으로 잡아요).
- N3: 겹침 판정은 대소문자를 접어 비교해요(충돌을 더 만드는 쪽이라 안전). 그런데 "정확히 생성 경로인가"도 같은 접기를 써서, 철자만 다른 경로(`Skills/Moai/References/Source-Manifest.json`)가 생성 파일로 취급돼 진짜 충돌이 가려질 수 있어요. 대소문자를 구분하는 파일 시스템에서는 그것이 다른 파일이에요. 이쪽은 충돌을 숨기는 방향이라 안전하지 않아요.
- N4: 그 SPEC 의 Result 가 센 `--lib` 시험 수(1458)와 READY 노트의 수(1491)가 달라요.

## Non-goals

- 생성 경로 목록(`GENERATED_SHARED_PATHS`)을 늘리거나 규칙의 뜻을 바꾸기: 정확히 한 줄인 쌍만 겹침이 아니라는 성질은 그대로예요.
- 클레임 쪽 겹침(`overlap`, `card check`)의 변경: 그쪽은 수정 없이 통과해야 해요.
- 병합 드라이버와 `.gitattributes` 일치 시험(그 SPEC 의 AC-2, AC-4): 이미 있어요.

## Requirements

- REQ-1: The reservation module header and the ledger reservation document shall each name the generated-path rule (`card_overlap`) next to the overlap predicate, without calling overlap the only judgement.
- REQ-2: The real binary's `card take` command shall take a card whose scope is exactly the generated manifest while another card holds a folder pattern that reaches it.
- REQ-3: The real binary's `card scope` command shall refuse a widening that adds a path another card's scope reaches, naming the path, apart from a pair in which one side is exactly a generated shared path.
- REQ-4: The card overlap rule shall treat as a generated shared path a path whose segments equal the generated path's segments with the case kept, and no other path.
- REQ-5: The overlap judgement itself shall keep comparing paths with the case folded.
- REQ-6: The Result of SPEC-20261119-generated-files-no-lock shall state the `--lib` test count measured at that SPEC's last head with the command that measured it, in the same line as the number.

## Acceptance

- AC-1 (REQ-1): the header of `src/ledger/reservation.rs` and `ledger_reservation.md` each carry `card_overlap`, and neither says overlap is the one or only judgement; the test reads both files at build time, so a later edit that drops the rule fails it.
  verify: `cargo test --locked --offline --test card_generated_scope the_reservation_docs_name_the_generated_path_rule`
- AC-2 (REQ-2, REQ-3): the real binary takes a manifest-only card beside a `skills/moai/**` card, then widens a held card with the manifest (allowed), with a skill file the other card's folder reaches (refused, the message names the skill file and not the manifest) and with a file nobody else reaches (allowed); the ledger scope after each call is the one the call reported.
  verify: `cargo test --locked --offline --test card_generated_scope`
- AC-3 (REQ-4, REQ-5): a spelling that differs from the generated path only in case is no generated path, so it still conflicts with a card that holds a folder pattern reaching it; the exact spelling, and the same path written with `./` and `..` segments, stay ignored; two paths that differ only in case still overlap.
  verify: `cargo test --locked --offline --lib scope`
- AC-4 (REQ-6): the Result line of the first SPEC names the command `cargo test --locked --offline --lib`, the head it ran at and the number it printed; the number is the one this SPEC's Result records from running that command at that head.
  verify: `cargo test --locked --offline --lib` run in a worktree at that head

## Scope

- write: `src/scope.rs`, `src/ledger/reservation.rs`, `docs/design-review/appendix/ledger_reservation.md`, `tests/card_generated_scope.rs`, `.moai/specs/SPEC-20261119-generated-files-no-lock/spec.md` (one Result line), `.moai/specs/SPEC-20261120-generated-scope-followup/**`
- read-only: `.gitattributes`, `src/card/**`, `src/ops/card.rs`, `tests/card_scope_widen_cli.rs`
- forbidden: the real HOME and ledger, the network, a change of what `GENERATED_SHARED_PATHS` lists

## Plan

1. SPEC first (this file), then research is written down in Decisions: where the case folding sits (`segments`), who calls `card_overlap` (reserve, widen, next), what the widen command takes.
2. After the first slice has landed on the local integration branch, the branch of this work takes that head (merge, no rebase): the code these items are about is not on the branch before.
3. RED: tests first — a case-spelled path that today counts as generated (AC-3), the docs test (AC-1), the CLI cases (AC-2); the first fails on the current code, the case test on `is_generated`.
4. GREEN: `src/scope.rs` compares generated paths with the case kept (one helper next to `segments`; `overlap` unchanged), the two doc places name the rule, the Result line carries the count.
5. Mutants after committing (at most five): fold the case again in `is_generated`; drop the generated rule from `card_overlap`; make the widen path call `overlap` instead of `card_overlap`; remove `./`/`..` resolution from the exact comparison; empty the doc sentence. fmt, fmt-ops, `clippy --all-targets`, `spec check`, the scan tests, the registered checks, `--bins`, trailer-check, tip `merge-tree`, an independent audit, then the READY line.

## Risks

- A path folded to lower case and one kept exact must agree on `.` and `..` resolution; the exact comparison reuses the same walk, only the lowercasing differs (one helper with a flag).
- On a case-insensitive file system (the macOS default) two spellings are one file, so the exact comparison now refuses a coexistence that is harmless there. That is the safe side: a refusal can be undone by writing the path as the list spells it, a hidden conflict cannot be seen.
- The count of AC-4 changes with every merge; it is tied to one named head, not to "now".

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | N3 is fixed by comparing exact case (REQ-4), not by documenting the folding | folding for overlap only adds conflicts; for the generated rule it removes one — the unsafe direction; the audit called it theoretical, but the repository gives no reason to leave an unsafe direction in place | fold again and document it, with the test pinning that choice |
| D-2 | `overlap` keeps folding (REQ-5) | design boundary row 6 of the claim overlap (#2392); claims and cards share it | none needed |
| D-3 | the count is measured at the first SPEC's last READY head, in a detached worktree | "at its head" has to name a head; a number from a merged tip would count other lanes' tests | measure again at a named head |
| D-4 | the CLI cases drive `card take` and `card scope` as the existing tests do (temporary HOME and project) | no new harness; AC-2 stays inside the real binary | none |
| D-5 | the docs test reads the two files with `include_str!` at build time | a fixed pair of paths, no tree walk, so the repository scans do not list it | register it in `tools/repo_scans.json` if the scan test asks |

## Result

(draft: filled in with evidence at RUN.)
