---
id: SPEC-20261126-approval-digest-decisions
title: "A new SPEC approval also binds the Decisions section, so a decision changed after approval needs a new approval"
status: draft
revision: 2
created: 2026-11-26
---

# A new SPEC approval also binds the Decisions section, so a decision changed after approval needs a new approval

## Interview

- Q1: SPEC 승인을 어디에 묶나요? → A: "SPEC 승인은 요구와 기준이 바뀔 때만 다시 받고, 계획이나 테스트 파일 위치가 바뀌는 것은 다시 묻지 않아요"(2026-11-25 사용자 선택, `SPEC-20261125-approval-digest-requirements-only` 의 Interview Q1 에 인용). 이 SPEC 의 D-1 은 이 선택의 범위에 `## Decisions` 를 더해요.
- Q2: 승인 digest 에 `## Decisions` 칸도 넣을까요? → A: "결정 칸도 넣기 (Recommended)"(2026-11-26 사용자 직접 선택, 리더가 AskUserQuestion 답을 카드 노트 10116 에 옮겨 적음). 11-25 선택 가운데 "Decisions 는 digest 밖"인 부분만 사용자가 바꿨어요.
- Q3: 이전 승인 기록과 `spec show` 는요? → A: Decisions D-2, D-3 의 기본값 A 예요(같은 노트, 사용자의 직접 선택은 D-1 이에요).

## Intent

승인 관문을 더 닫는 SPEC 이에요. 지금 승인 digest 는 `## Decisions` 를 덮지 않아서, 승인 뒤에 결정 글이 바뀌어도 `moai spec check` 가 통과해요. 실제 사례: `SPEC-20261126-audit-reads-by-object-id` 는 개정 2 에서 D-4(스캔 범위)를 바꿨는데 digest 와 approval_digest 가 개정 1 과 같았어요(2026-11-26 위임 재승인 기록에 그 사실이 적혀 있어요).

SPEC 규칙으로 보면 결정 표는 승인받는 글이에요. `workflows/plan` 은 고를 것을 선택지·기본값과 함께 `## Decisions` 에 적고 사용자의 선택을 승인으로 삼으라고 해요. 이 tip(91dd561a5a)에 커밋된 승인 기록의 `basis` 86줄 가운데 38줄이 결정 id(D-n)를 적고, 위험 id(R-n)를 적은 줄은 0줄이에요(`git grep` 으로 셈).

지금 동작은 실수가 아니라 정해진 것이에요. 사용자가 직접 승인한 `SPEC-20261125-approval-digest-requirements-only` 의 REQ-2 와 Non-goals, `internal/sdd` 규칙 6 이 "Decisions 는 digest 밖"이라고 적고, `tests/spec_approval_digest.rs` 가 그것을 시험해요. 사용자는 2026-11-26 에 그중 Decisions 부분을 바꾸기로 골랐고(D-1 A), 이 SPEC 은 그 부분만 바꿔요.

## Non-goals

- `## Risks` 는 새 digest 에도 넣지 않아요(D-1 의 선택지 C 가 넣는 쪽이에요).
- `## Interview`, `## Intent`, `## Plan`, `## Result`, front matter, `write:`·`read-only:` 줄, `verify:` 명령은 지금처럼 digest 밖이에요.
- `Supersedes` 는 이 마크다운 형식의 절이 아니라(`src/spec_doc.rs` 의 `SECTIONS`) 다루지 않아요. JSON 쪽 `spec::content_digest` 와 `plan::admit` 의 snapshot digest 도 바꾸지 않아요.
- 이미 승인된 SPEC 파일의 `approval` 블록을 고치지 않아요.
- 승인 기록의 형식(`by`, `at`, `basis`, `digest` 네 줄)을 바꾸지 않아요.

## Requirements

<!-- This SPEC widens the SPEC approval gate (tier 2). It only adds coverage: every record that is refused today stays refused, and a record made with the new approval digest is also refused after a change of the Decisions section. -->
- REQ-1: The approval digest of a SPEC shall be computed over everything the earlier approval digest covers (the requirement sentences with their links, the acceptance criteria with their ids, links, result texts and verification kinds, and the `forbidden:` entries) and the lines of the `## Decisions` section, under a domain tag of its own, so that it never equals the earlier approval digest of the same document.
- REQ-2: The approval digest shall take every line the section reader assigns to `## Decisions` (up to the next `##` heading outside a fenced block), in order, with trailing whitespace removed and blank lines left out.
- REQ-3: When a non-blank line of `## Decisions` is added, removed, moved or changed in more than its trailing whitespace, the approval digest shall change. · failure: none — the digest is a computed value and cannot fail on its own.
- REQ-4: The approval digest shall not depend on the front matter, on any section other than Requirements, Acceptance, Scope and Decisions, on the `write:` and `read-only:` entries of `## Scope`, or on the command after `verify:` of a criterion.
- REQ-5: When an approved SPEC records a digest, `moai spec check` shall accept the record only when it equals the approval digest, the earlier approval digest or the legacy content digest of the same document.
- REQ-6 (REQ-5): IF the recorded digest equals none of the three, THEN `moai spec check` shall report the error `approval_digest`, so that `card land`, `card spec`, `spec take` and `spec close` refuse the SPEC as they do today.
- REQ-7: Where an approved SPEC records the earlier approval digest or the legacy content digest, that digest shall keep exactly the value and the coverage it has before this change.
- REQ-8: When `moai spec show` prints a SPEC, it shall print the new value as `approval_digest`, the legacy content digest as `digest`, and as `approval_covers` which of the three the recorded digest equals (`decisions`, `requirements`, `content`, or null when there is no record or it equals none). · failure: none — a read-only print of computed values.
- REQ-9: The approval digest documentation shall state in `internal/sdd`, in `workflows/plan`, in the help of `spec check` and `spec show` and in the SPEC operations document that a new approval also covers `## Decisions` and that a change of `## Risks` or another uncovered part needs no new approval.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5, REQ-6): in a temporary SPEC approved with the new approval digest, one changed character of a Decisions line, an added line, a removed line, two swapped lines, and a `## Decisions` section added to or removed from the document each change the approval digest, make `moai spec check` report `approval_digest`, and make `moai card spec` refuse the SPEC (guard against a digest that ignores the section) · verify: `cargo test --offline --locked --test spec_approval_digest decisions1_`
- AC-2 (REQ-2, REQ-4): added trailing whitespace and blank lines inside `## Decisions`, and changes of Risks, Intent, Non-goals, Plan, Result, Interview, the front matter, a `write:` or `read-only:` entry and a `verify:` command keep the new approval digest equal and `moai spec check` passes · verify: `cargo test --offline --locked --test spec_approval_digest decisions2_`
- AC-3 (REQ-5, REQ-6, REQ-7): a SPEC approved with the earlier approval digest, whose value for the test document is pinned as the literal the code before this change computes, passes while unchanged and also after a Decisions change (the stated gap of old records), and is refused after a changed requirement; a SPEC approved with the legacy content digest is still refused after a changed `write:` entry · verify: `cargo test --offline --locked --test spec_approval_digest decisions3_`
- AC-4 (REQ-8): `moai spec show` answers `approval_covers` as `decisions`, `requirements`, `content` and null for the four kinds of record, and its `approval_digest` differs from the earlier approval digest · verify: `cargo test --offline --locked --test spec_approval_digest decisions4_`
- AC-5 (REQ-9): the two skill texts, the two help texts and the operations document name `## Decisions` as covered by a new approval and no longer list Decisions among the changes that need none · verify: `cargo test --offline --locked --test spec_approval_digest decisions5_`
- AC-6 (REQ-5, REQ-7): every SPEC checked into the repository still passes the check after the change · verify: `cargo test --offline --locked --test spec_cli the_repository_specs_pass_check`

## Scope

- write: `.moai/specs/SPEC-20261126-approval-digest-decisions/**`, `src/spec_doc.rs`, `src/ops/spec.rs`, `tests/spec_approval_digest.rs`, `skills/moai/references/internal/sdd.md`, `skills/moai/workflows/plan.md`, `docs/design-review/appendix/operations/spec.md`, `docs/later.md`
- write: `skills/moai/references/source-manifest.json` (두 글의 sha256 줄, shared file 이라 line-lease)
- read-only: `src/spec.rs`, `src/plan.rs`, `src/card/spec_link.rs`, `tests/spec_cli.rs`, `tests/spec_doc.rs`, `tests/card/card_spec_cli.rs`, `tools/repo_scans.json`
- forbidden: `moai spec check` 가 오류를 내는 길을 약하게 하는 변경, 이미 승인된 SPEC 의 `approval` 블록을 고치는 일, 전역 설치본과 실제 HOME, 네트워크, Python

## Plan

1. RED: `tests/spec_approval_digest.rs` 에 `decisions1_`..`decisions5_` 테스트를 더해요. 다른 SPEC 이 못박은 표를 하나 바꿔요: `UNCOVERED` 의 `Decisions` 행을 `COVERED` 로 옮겨요(선행 SPEC AC-1 의 목록 가운데 Decisions 만 뒤집혀요). 이전 승인 digest 의 값은 코드를 바꾸기 전 빌드로 계산해 글자 그대로 못박아요.
2. `src/spec_doc.rs`: 지금의 `approval_digest` 계산은 이전 값(`approval@1`)으로 그대로 두고, Decisions 줄을 더한 `approval@2` 계산을 만들어요. 기록된 값은 셋 중 하나와 같으면 통과해요. `SpecDoc` 에 `approval_covers` 를 담아요. GREEN.
3. `src/ops/spec.rs` 도움말, `internal/sdd` 규칙 6, `workflows/plan` 의 Resume 문장, `docs/design-review/appendix/operations/spec.md` 의 digest 줄을 새 범위로 고치고 source manifest 두 줄을 line-lease 로 써요.
4. `docs/later.md` 의 2026-11-24 spec-digest 줄은 이 SPEC 이 Decisions 부분을 해결했다고 고쳐요.
5. 가드 뮤턴트(새 테스트마다 하나), `moai check affected --card` 영수증, READY.

## Risks

- R-1 (이미 승인된 SPEC): D-2 의 기본값이면 승인 기록 86건은 다시 승인할 필요가 없고, 그 기록의 Decisions 는 지금처럼 조용히 바뀔 수 있어요(이 SPEC 의 계기가 된 `SPEC-20261126-audit-reads-by-object-id` 개정 2 도 그래요). 그 틈은 그 SPEC 이 새 개정으로 다시 승인될 때 닫혀요. 착지 뒤에는 `spec show` 의 `approval_digest` 가 옛 승인 노트의 값과 달라 보여요; 감사와 리더는 `approval_covers` 와 `spec check` 로 읽어요.
- R-2 (설치본 순서): 새 값으로 적은 승인은 이 변경 전 빌드의 `moai` 가 `approval_digest` 오류로 거절해요. 착지 뒤 리더가 새 빌드를 설치한 다음에 승인을 기록해요. 설치 전에 옛 빌드가 보여 준 값으로 적은 승인은 통과하지만 Decisions 를 덮지 않아요.
- R-3 (재승인 비용): 새 값으로 승인한 SPEC 은 구현 중에 결정을 더하거나 표를 다시 맞추기만 해도(줄 끝 공백·빈 줄 제외) 다시 승인해야 해요. 구현 중의 메모는 Plan 이나 Result 에 적어요. 이 SPEC 은 이미 막는 검사의 범위를 넓히고 지금 저장소의 기록은 하나도 막지 않아서(AC-6) 관찰 기간을 두지 않아요.
- R-4 (남는 우회): 새 승인에 이전 값을 적으면 여전히 통과해요(D-2 A). `spec show` 는 이전 값을 보여 주지 않고 `approval_covers` 가 `requirements` 로 드러내요.
- R-5 (덮이지 않는 자리): 형식에 없는 `##` 제목 아래에 결정을 덧붙이면 digest 가 같고 `spec check` 는 `section_unknown` 경고만 내요. 감사가 읽어요.
- R-6: tier 2 라서 승인 전에는 코드를 쓰지 않고, 이 세션은 승인을 직접 기록하지 않아요.
- R-7 (이 SPEC 자신의 승인): 승인 전에 읽는 `approval_digest` 는 지금 빌드의 계산이라 Decisions 를 덮지 않아요. 착지 뒤 `spec show` 는 이 SPEC 에도 다른 값을 보여 주고 `approval_covers` 는 `requirements` 예요(D-2 A 에서는 통과). D-2 B 를 고르면 기준 날짜를 착지 다음 날 이후로 두거나, 새 빌드를 설치한 뒤 이 승인을 새 값으로 다시 적어야 해요.
- R-8 (선행 SPEC 과의 어긋남): `SPEC-20261125-approval-digest-requirements-only` 의 REQ-2 와 AC-1 은 Decisions 를 덮지 않는다고 계속 적혀 있어요. 그 파일은 승인 digest 를 깨지 않고 고칠 수 없어서 그대로 두고, 감사는 이 어긋남을 비차단 줄로 읽어요.

## Decisions

추천이 맨 앞이에요. D-1 은 사용자가 2026-11-25 에 직접 승인한 선행 SPEC 의 일부를 바꾸는 선택이라 사용자가 직접 골랐어요(A).

- D-1 (승인 digest 가 새로 덮는 것): (A, 사용자 선택 2026-11-26) `## Decisions` 만 더해요 — 승인의 `basis` 가 결정 id 를 인용하므로 결정 글이 바뀌면 승인한 선택이 바뀐 것이에요. 선행 SPEC 의 REQ-2·Non-goals 와 `internal/sdd` 규칙 6 의 Decisions 부분이 바뀌어요. (B) 바꾸지 않고 카드를 접어요 — 11-25 선택이 그대로이고, 행동을 묶는 결정은 요구(REQ)로 옮겨 적어요; 카드 기준 1 은 못 채워요. (C) `## Decisions` 와 `## Risks` 를 더해요 — 가장 엄격하지만 위험 문장을 다듬거나 새 위험을 더할 때마다 다시 승인해야 해서 11-25 선택과 더 멀어요.
- D-2 (이미 승인된 SPEC 과 이전 값): (A, 기본값으로 정함) 이전 승인 digest 와 옛 content digest 를 계속 받아요 — 다시 승인할 것이 없고 지금 거절되는 것은 계속 거절돼요. (B) 이전 승인 digest 는 `approval.at` 이 정한 날짜보다 앞선 기록에만 받아요 — 새 기록의 우회를 닫지만, 바뀌는 날 당일의 승인은 날짜로 가를 수 없고 설치본이 옛것일 때 한 승인이 거절돼요. (C) 승인된 SPEC 을 모두 새 값으로 다시 승인해요 — 사용자 승인 86건을 다시 받아야 해서 권하지 않아요.
- D-3 (`spec show` 의 새 필드): (A, 기본값으로 정함) `approval_covers` 를 더해 어떤 범위로 승인됐는지 보여 줘요. (B) 필드를 더하지 않아요(REQ-8 의 마지막 구절과 AC-4 가 줄어요).
