---
id: SPEC-20261125-approval-digest-requirements-only
title: "A SPEC approval binds only the requirements and acceptance criteria, so a changed plan or test place needs no new approval"
status: draft
revision: 1
created: 2026-11-25
---

# A SPEC approval binds only the requirements and acceptance criteria, so a changed plan or test place needs no new approval

## Interview

- Q1: SPEC 승인을 어디에 묶나요? 지금은 계획이나 테스트 파일 위치가 바뀌어도 다시 승인을 받는 일이 잦아요. → A: "SPEC 승인은 요구와 기준이 바뀔 때만 다시 받고, 계획이나 테스트 파일 위치가 바뀌는 것은 다시 묻지 않아요"(사용자 선택 2026-11-25, 리더의 선택 질문 "위험 등급과 승인의 범위를 좁힐까요?" 의 답 "둘 다 (Recommended)" 중 두 번째, 카드 본문에 인용).
- Q2: 요구·기준 글이 한 글자라도 바뀌면 어떻게 하나요? → A: 지금처럼 다시 승인을 받아요(같은 카드 본문의 기준 2: "Requirements 나 Acceptance 절의 한 글자가 바뀌면 승인 digest 가 바뀌고 spec check 와 착지 관문이 거절해요").
- Q3: 이미 승인된 SPEC 은요? → A: 그대로 통과해요(카드 본문: "옛 방식의 승인 기록을 읽는 길을 남겨요"). 어떻게 읽을지는 D-3 이 정해요.

## Intent

승인 관문을 좁히는 SPEC 이에요. 승인 관문은 약해지면 안 돼요: 요구·기준 글이 바뀌면 지금처럼 다시 막고, 무엇이 digest 밖으로 나가는지는 아래 요구 REQ-2 가 낱낱이 적어요.

읽은 코드(2026-11-25, tip dfe34f874c): 승인 digest 는 이미 Plan·Decisions·Result·Interview 와 `revision`·`status` 를 덮지 않아요. 덮는 것은 `crate::spec::content_digest` 의 다섯 값이에요 — 요구 기록(문장, GEARS 조각, 연결된 기준 id), 기준 기록(id, 결과 글, 검증 종류와 selector), `## Scope` 의 `write:`·`read-only:`·`forbidden:` 줄, policy 판(늘 비어 있어요), `verify:` 명령 목록(`src/spec.rs:561-572`, `src/spec_doc.rs:1934-2022`). 그래서 테스트 파일 위치가 바뀔 때 다시 묻게 되는 길은 둘이에요: 기준의 `verify:` selector 와 `## Scope` 의 `write:` 에 적힌 테스트 경로. 판 번호 정리 같은 재승인은 코드가 아니라 글 규칙 때문이에요: `internal/sdd` 규칙 6 과 `workflows/plan` 이 "승인된 SPEC 을 바꾸면 revision 을 올리고 새 승인"이라고 적어요. 코드는 digest 가 다를 때만 `approval_digest` 오류를 내요.

digest 를 쓰는 곳은 한 길이에요. `parse` 가 기록된 값과 계산한 값을 비교해 `approval_digest` 오류를 내고, `spec check`, 착지(`card land` 의 `done_refusal_at_tip`), `card finish`, `card spec`, `spec take`, `spec close` 가 모두 오류 수를 보고 거절해요. 그래서 이 SPEC 은 그 비교 한 곳과 `spec show` 의 출력, 그리고 글 규칙만 바꿔요.

쓰기 범위에 대한 근거(코드 확인, 2026-11-25): SPEC 의 `## Scope` 는 쓰기를 막는 관문이 아니에요. 카드에 자기 범위가 없을 때 `card take` 가 `write:` 목록을 기본값으로 가져오는 데에만 쓰여요(`src/card/take_scope.rs:57`, `src/card/dispatch.rs:565`). 쓰기를 허락하는 것은 카드 예약이고(`src/card/work.rs:1177`, `card check`), 예약은 `card scope` 로 넓힐 수 있으며 호출자를 확인하지 않아요(`src/ops/card.rs:1498`). 착지와 감사 코드는 변경 경로를 범위와 견주지 않고, `card submit` 이 범위 밖 경로를 READY 노트의 `outside-scope:` 줄로 알릴 뿐이에요(`src/card/submit.rs:176`). `read-only:` 와 `forbidden:` 를 읽는 코드는 없어요. 즉 Scope 를 digest 에서 빼도 쓰기 범위를 넓히는 새 길이 생기지는 않지만, 사용자가 승인한 범위가 digest 로 지켜지던 효과는 없어져요. D-1 이 그 정도를 고르고, 이 SPEC 은 쓰기 관문을 새로 만들지 않아요(Non-goals).

## Non-goals

- 쓰기 범위를 지키는 관문(`card check` 를 자동으로 부르는 훅, 착지 때의 경로 비교, `card scope` 의 호출자 확인)을 새로 만들지 않아요. 한 줄로 `docs/later.md` 에 남겨요.
- 승인 기록 방식(사용자의 같은 세션 선택, `approval:` 네 줄)을 바꾸지 않아요.
- JSON 쪽 `spec::approve` 와 `plan::admit` 의 snapshot digest(`src/plan.rs:381`)를 바꾸지 않아요: 같은 `content_digest` 를 그대로 써요.
- 이미 승인된 SPEC 파일의 `approval.digest` 를 고쳐 쓰지 않아요.
- Decisions 표나 Supersedes 가 digest 밖인 상태(`docs/later.md` 의 기존 한 줄)를 바꾸지 않아요: Decisions 는 digest 밖이라는 카드의 결정 그대로예요.
- 위험 등급(tier)의 기준을 바꾸지 않아요.

## Requirements

<!-- This SPEC changes the SPEC approval gate (tier 2). A requirement or criterion character that changes after approval still refuses; the gate is narrowed only for what REQ-2 lists, and a SPEC approved under the earlier digest keeps the earlier strictness (REQ-6). -->
- REQ-1: The approval digest of a SPEC shall be computed over the requirement sentences with their links to criteria, the acceptance criteria with their ids, links, result texts and verification kinds (`verify`, `manual`, `unknown` or `none`), and the `forbidden:` entries of `## Scope`, and over nothing else.
- REQ-2: The approval digest shall not depend on the front matter lines, on any section other than Requirements, Acceptance and Scope, on the `write:` and `read-only:` entries of `## Scope`, or on the command after `verify:` of a criterion.
- REQ-3: When a requirement sentence, a criterion result text, a criterion verification kind, a link between a requirement and a criterion, or a `forbidden:` entry changes, the approval digest shall change. · failure: none — the digest is a computed value and cannot fail on its own.
- REQ-4: When an approved SPEC records a digest, `moai spec check` shall accept the record only when it equals the approval digest or the legacy content digest of the same document.
- REQ-5 (REQ-4): IF the recorded digest equals neither the approval digest nor the legacy content digest, THEN `moai spec check` shall report the error `approval_digest`, so that `card land`, `card spec`, `spec take` and `spec close` refuse the SPEC as they do today.
- REQ-6: Where an approved SPEC records the legacy content digest, the legacy content digest shall keep covering the requirements, the criteria with their `verify:` commands and all entries of `## Scope`, exactly as before this change.
- REQ-7: When `moai spec show` prints a SPEC, it shall print the legacy content digest as `digest` and the approval digest as `approval_digest`. · failure: none — a read-only print of two computed values.
- REQ-8: The `internal/sdd` and `workflows/plan` texts shall state that an approval stays valid while the approval digest matches and that only a change to what the approval digest covers needs a new approval.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary SPEC approved with its approval digest, a changed Intent, Non-goals, Risks, Decisions, Plan, Result or Interview text, a changed front matter `title`, `status` or `revision`, a changed `write:` or `read-only:` entry and a changed `verify:` command each leave the approval digest equal and `moai spec check` passes · verify: `cargo test --offline --locked --test spec_approval_digest ac1_`
- AC-2 (REQ-1, REQ-3, REQ-5): one changed character of a requirement sentence, of a criterion result text, a criterion changed from `verify:` to `manual:`, a changed link, and a changed `forbidden:` entry each change the approval digest, make `moai spec check` report `approval_digest`, and make `moai card spec` refuse the SPEC (guard against a digest that ignores a covered part) · verify: `cargo test --offline --locked --test spec_approval_digest ac2_`
- AC-3 (REQ-4, REQ-5, REQ-6): a SPEC approved with the legacy content digest passes while its document is unchanged, and still fails with `approval_digest` when its `write:` entry or its `verify:` command changes (the earlier strictness stays) · verify: `cargo test --offline --locked --test spec_approval_digest ac3_`
- AC-4 (REQ-7): `moai spec show` answers both `digest` and `approval_digest`, and the two differ for a SPEC whose Scope has a `write:` entry · verify: `cargo test --offline --locked --test spec_approval_digest ac4_`
- AC-5 (REQ-8): the `internal/sdd` and `workflows/plan` texts of the skill carry the approval-digest sentence and no longer say that every change of an approved SPEC needs a new approval · verify: `cargo test --offline --locked --test spec_approval_digest ac5_`
- AC-6 (REQ-4, REQ-6): every SPEC checked into the repository still passes `moai spec check --all` after the change (no old approval breaks) · verify: `cargo test --offline --locked --test spec_cli the_repository_specs_pass_check`

## Decisions

각 결정은 되돌릴 수 있고 분류는 none 이에요(승인 관문의 강도를 다루는 D-1·D-2 는 사용자 선택이 필요한 tier 2 항목이에요). 추천이 맨 앞이고 기본값이에요.

- D-1 (`## Scope` 를 digest 에서 어떻게 빼나요): (A, 기본·권장) `write:` 와 `read-only:` 는 빼고 `forbidden:` 는 남겨요 — "하지 않아요" 목록이 승인 뒤에 조용히 줄어드는 길은 막아요. 카드 본문의 "요구·기준 절만"보다 한 줄(`forbidden:`)을 더 묶는 것이고, 되돌리려면 REQ-1 한 구절만 지워요. (B) Scope 전체를 빼요(카드 본문 그대로): 가장 단순하지만 `forbidden:` 항목을 승인 뒤에 조용히 지울 수 있어요. (C) Scope 전체를 남겨요: 관문은 그대로지만 테스트 파일을 `write:` 에 적은 SPEC 은 위치가 바뀔 때 계속 다시 묻게 돼요. 근거: `write:` 는 카드 범위의 기본값일 뿐이고 이미 `card scope` 로 호출자 확인 없이 넓힐 수 있어서, (A)·(B) 는 새 구멍을 만들지 않아요(Intent).
- D-2 (`verify:` 줄을 어떻게 다루나요): (A, 기본·권장) 명령(selector)은 빼고 검증 종류(`verify`·`manual`·`unknown`·`none`)는 남겨요 — 테스트 파일 위치가 바뀌어도 다시 묻지 않지만, 자동 검증을 사람 확인으로 낮추는 일은 다시 물어요. (B) `verify:` 줄 전체를 남겨요: 테스트 위치가 바뀔 때 계속 다시 물어요(카드 본문의 사용자 결정과 어긋나요). (C) selector 를 완전히 빼고 종류도 빼요: 약한 검증으로 바꾸는 길이 열려요(고르지 않는 것을 권해요). 남는 위험: 같은 종류 안에서 selector 를 약한 명령으로 바꾸는 일은 digest 가 못 잡아요; `ac_work` 훅과 독립 감사가 기준 글과 코드를 견주어 잡아요(Risks R-2).
- D-3 (옛 digest 로 승인된 SPEC): (A, 기본·권장) 두 계산을 함께 받아요: 기록된 값이 승인 digest 이거나 옛 `content_digest` 와 같으면 통과해요. 옛 승인은 옛 엄격함 그대로라서 약해지지 않아요(REQ-6, AC-3). 새 승인은 `spec show` 의 `approval_digest` 를 적어요. (B) `approval:` 에 새 줄(예: `covers: requirements-acceptance`)을 두고 없으면 옛 방식으로 읽어요: 읽는 길이 분명하지만 승인 형식(네 줄)이 바뀌어요. (C) 승인된 SPEC 57건의 digest 를 새 값으로 고쳐 써요: 사용자 승인 기록을 바꾸는 일이라 하지 않아요(Non-goals).
- D-4 (`revision`): (A, 기본·권장) 승인 digest 가 바뀔 때만 `revision` 을 올리고 새 승인을 받아요; 그 밖의 수정은 `revision` 도 승인도 그대로예요. (B) 본문을 바꿀 때마다 `revision` 을 올리되 승인은 digest 가 맞으면 유지해요: 개정 이력은 남지만 판 번호만 올라가는 노이즈가 남아요.

## Scope

- write: `.moai/specs/SPEC-20261125-approval-digest-requirements-only/**`, `src/spec_doc.rs`, `src/ops/spec.rs`, `tests/spec_approval_digest.rs`, `skills/moai/references/internal/sdd.md`, `skills/moai/workflows/plan.md`
- write (한 줄, line-lease): `tools/repo_scans.json`(새 시험 파일 등록), `docs/later.md`(쓰기 관문이 없다는 한 줄)
- read-only: `src/spec.rs`(`content_digest` 는 JSON 경로가 함께 쓰므로 바꾸지 않아요), `src/card/spec_link.rs`, `src/plan.rs`, `tests/spec_doc.rs`, `tests/spec_cli.rs`, `tests/card/card_spec_cli.rs`, `docs/design-review/appendix/operations/spec.md`, `docs/rust-product-path.md`
- forbidden: `moai spec check` 가 오류를 내는 길을 약하게 하는 변경, 이미 승인된 SPEC 의 `approval` 블록을 고치는 일, 전역 설치본과 실제 HOME, 네트워크, Python

## Plan

1. RED: `tests/spec_approval_digest.rs`(AC-1..AC-5)가 지금 코드에서 실패해요. 임시 SPEC 폴더를 만들고(실제 원장과 HOME 을 안 써요) 새 계산(`approval_digest`)이 아직 없는 상태에서 각 기준이 실패하는 것을 먼저 봐요.
2. `src/spec_doc.rs`: 승인 digest 를 계산하는 함수를 추가해요(요구 기록과 연결, 기준의 id·링크·결과 글·종류, `forbidden:` 줄). 기존 `content_digest`(옛 계산)는 그대로 둬요. `SpecDoc` 에 `approval_digest` 를 담고 `parse` 의 비교는 기록된 값이 둘 중 하나와 같으면 통과시켜요(REQ-4, REQ-5, REQ-6). GREEN.
3. `src/ops/spec.rs`: `spec show` 와 도움말이 두 값을 말해요(REQ-7).
4. 글: `internal/sdd` 규칙 6 과 `workflows/plan`(40행 부근)에 승인 digest 문장을 넣고 "바꾸면 모두 새 승인" 문장을 없애요(REQ-8). `docs/later.md` 에 쓰기 관문이 없다는 한 줄을 남겨요.
5. 가드 뮤턴트(새 계산에서 `forbidden:` 줄이나 기준 결과 글을 빼면 AC-2 가 죽어요, 옛 계산 비교를 빼면 AC-3 이 죽어요), `moai check affected --card` 영수증(`tests/spec_doc.rs`·`tests/spec_cli.rs` 를 지나가요), READY.

## Risks

- R-1: 새 digest 가 덮는 것을 빼먹으면 요구나 기준이 바뀌어도 승인이 그대로 통과해요. AC-2 가 덮이는 부분마다 한 글자 변경을 시험하고 가드 뮤턴트로 한 번 더 지켜요.
- R-2: `verify:` selector 를 약한 명령으로 바꿔도 같은 종류면 digest 가 같아요(D-2). 감사가 기준 글과 코드를 견주고 `ac_work` 훅이 기준의 `verify:` 줄을 읽어요; 이 SPEC 은 그 위에 새 검사를 얹지 않아요.
- R-3: `write:` 를 digest 에서 빼면 승인 뒤에 SPEC 의 쓰기 목록이 넓어져 비어 있는 카드의 기본 범위가 커질 수 있어요. 지금도 `card scope` 가 호출자 확인 없이 넓히고 착지가 경로를 견주지 않으므로 새로 생기는 구멍은 아니에요. `outside-scope:` 줄이 READY 에 남고 감사가 읽어요; 쓰기 관문 자체는 한 줄로 `docs/later.md` 에 남겨요.
- R-4: 옛 계산과 새 계산을 둘 다 받으면 기록된 값 하나가 두 해시 중 하나와 우연히 겹칠 수 있다고 걱정할 수 있어요. 입력이 달라 겹침은 실제로 생기지 않고(sha256), 옛 승인은 옛 엄격함 그대로라 새 계산으로 풀리는 일이 없어요(AC-3).
- R-5: 이 SPEC 은 승인 관문을 바꾸는 tier 2 라서 승인 없이 코드를 쓰지 않아요. 승인은 리더가 사용자에게 묻고, 이 세션은 승인을 직접 기록하지 않아요.

## Result

- AC-1: met — 커밋 0d8e67dcfc 로 `cargo test --offline --locked --test spec_approval_digest ac1_` 통과(3개 ok, 로그 `target/fay-test.log`).
- AC-2: met — 커밋 0d8e67dcfc 로 같은 시험 파일의 `ac2_` 통과(2개 ok, 로그 `target/fay-test.log`).
- AC-3: met — 커밋 0d8e67dcfc 로 같은 시험 파일의 `ac3_` 통과(3개 ok, 로그 `target/fay-test.log`).
- AC-4: met — 커밋 0d8e67dcfc 로 같은 시험 파일의 `ac4_` 통과(1개 ok, 로그 `target/fay-test.log`); ac1_..ac4_ 합계 9개가 통과했고 나머지 3개는 문서 단계 전의 `ac5_` 였어요.
- AC-5: met — `skills/moai/references/internal/sdd.md` 규칙 6 과 `skills/moai/workflows/plan.md` 가 승인 digest 문장과 `approval_digest` 를 담고 옛 두 문장을 없앴어요; `cargo test --offline --locked --test spec_approval_digest` 를 슬롯에서 실행해 12개 모두 통과(exit 0, 로그 `target/fay-test3.log`).
- AC-6: met — 커밋 0d8e67dcfc 로 `tests/spec_cli.rs` 의 `the_repository_specs_pass_check` 통과, 함께 `spec_doc` 36개, `spec_cli` 18개, `spec_card_number_observe` 7개 통과(로그 `target/fay-test2.log`).
