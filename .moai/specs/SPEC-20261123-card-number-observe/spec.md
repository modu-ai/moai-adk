---
id: SPEC-20261123-card-number-observe
title: "spec check counts card numbers of new and revised SPECs as an observe notice"
status: draft
revision: 1
created: 2026-11-23
---

# spec check counts card numbers of new and revised SPECs as an observe notice

## Intent

`internal/sdd.md` 규칙 10 은 SPEC 에 카드 번호(`C-` 와 숫자)가 절대 없다고 말하는데, 이 저장소의 SPEC 347개 중 152개의 `spec.md` 본문에 758번 나와요(`moai spec check --all` 이 읽은 개수와 `rg` 로 센 수, 2026-11-23). 그런데 SPEC 쪽에서 이를 보는 도구는 앞머리 `cards` 키 거절 하나뿐이라 같은 어긋남이 감사마다 되풀이돼요. 규칙을 좁힐지(근거 칸 허용)나 기존 SPEC 을 고칠지는 사용자 결정으로 올라갔고(카드 노트 7724·7726), 그 답이 오기 전까지 규칙 글은 그대로 두고 이 SPEC 이 두 가지만 더해요. 하나, 새로 만들거나 개정해 다시 승인한 SPEC 의 카드 번호를 `moai spec check` 가 observe 알림(경고, 오류 아님)으로 세요. 둘, 규칙 10 문장이 몰래 바뀌지 않게 문장 핀 테스트 하나를 더해요. 기존 SPEC 은 건드리지 않아요.

## Interview

- Q1: 규칙 10 을 좁히는 선택지 (a) 를 지금 정하나요? → A: 아니요. (a) 는 사용자가 2026-11-17 에 직접 정한 규칙(`docs/decisions-2026-11-08.md:134`)을 좁히는 일이라 사용자에게 올렸고, 그 답 전에는 규칙 글을 바꾸지 않아요(리더 결정 카드 노트 7726).
- Q2: 그 사이 무엇을 하나요? → A: (d) 규칙을 그대로 두고 `spec check` 가 새 SPEC·개정 SPEC 의 카드 번호를 observe 알림으로 세고, 규칙 10 문장 핀 테스트를 더해요. 기존 SPEC 은 바꾸지 않아요(리더 결정 카드 노트 7726).

## Non-goals

- 기존 SPEC 152개의 본문을 고치는 일, 규칙 10 문장을 좁히는 일, 알림을 오류로 올리는 일은 이 SPEC 이 아니에요. 사용자 답과 observe 기간의 기록 뒤 따로 정해요.
- SPEC 밖(커밋, 번들 본문, 소스 코드)의 카드 번호 스캔은 이미 있어서 건드리지 않아요.
- `## Interview`·`## Result` 같은 칸별 구분은 하지 않아요. 알림은 파일 전체의 번호를 한 번에 세요.

## Requirements

- REQ-1: When `spec_doc::parse` reads a new-form SPEC whose `created` date or `approval.at` date is on or after `2026-11-25` and its text holds one or more card numbers, the checker shall report one warning finding `card_number_observe` that names the count and the first line.
- REQ-2 (REQ-1): IF the SPEC is a Go-generation SPEC, or neither `created` nor `approval.at` is on or after `2026-11-25`, THEN the checker shall report no `card_number_observe` finding.
- REQ-3: The finding `card_number_observe` shall have the severity warning, so that the SPEC's error count stays what it was without the finding. · failure: none — the severity is fixed in the code, with no input that could change it
- REQ-4: When the pin test reads `skills/moai/references/internal/sdd.md`, it shall require rule 10's sentence "Card numbers (`C-<n>`) never appear in a SPEC, branch name, commit message or pull request: cards are personal." word for word.
- REQ-5 (REQ-4): IF that sentence is missing or differs from the pinned text, THEN the pin test shall fail and name the rule.

## Acceptance

- AC-1 (REQ-1, REQ-3): 앞머리 `created: 2026-11-25` 이고 본문의 두 줄에 카드 번호 모양(`C-` 와 숫자)이 있는 SPEC 은 `card_number_observe` 경고를 하나 받고 오류는 그 경고 때문에 늘지 않아요. 메시지는 개수 2 와 첫 줄 번호를 담고 SPEC 의 글은 담지 않아요. `AC-1`·`SPEC-20261123-x`·`ABC-5` 처럼 낱말 안의 `C-` 는 세지 않아요. · verify: `cargo nextest run --offline --locked --test spec_card_number_observe -E 'test(a_new_spec_with_card_numbers_gets_one_observe_warning)'`
- AC-2 (REQ-1): `created` 가 옛날이어도 `approval.at` 이 `2026-11-25` 이후인 개정 SPEC 은 알림을 받아요. · verify: `cargo nextest run --offline --locked --test spec_card_number_observe -E 'test(a_revised_spec_is_judged_by_its_approval_date)'`
- AC-3 (REQ-2): 경계의 하루 전(`2026-11-24`)에 만든 SPEC 과 Go 세대 SPEC 과 날짜가 없는 SPEC 은 번호가 있어도 알림이 없어요. · verify: `cargo nextest run --offline --locked --test spec_card_number_observe -E 'test(older_go_and_undated_specs_get_no_notice)'`
- AC-4 (REQ-2, REQ-3): 이 저장소의 `.moai/specs/*/spec.md` 를 모두 읽어도 날짜가 경계 전인 SPEC 은 알림이 0이고 `moai spec check --all` 의 오류 수가 이 변경 전과 같아요. · verify: `cargo nextest run --offline --locked --test spec_card_number_observe -E 'test(the_real_specs_before_the_date_get_no_notice)'`
- AC-5 (REQ-4, REQ-5): 핀 테스트가 규칙 10 문장을 요구하고, 문장을 한 낱말 바꾼 사본은 거절해요. · verify: `cargo nextest run --offline --locked --test spec_card_number_observe -E 'test(rule_10_sentence_is_pinned)'`

## Scope

- write: `src/spec_doc.rs`(알림 하나와 날짜 상수), `tests/spec_card_number_observe.rs`(새 파일), `.moai/specs/SPEC-20261123-card-number-observe/spec.md`
- read-only: `skills/moai/references/internal/sdd.md`(핀 테스트가 읽기만 해요, 고치지 않아요), 기존 `.moai/specs/*/spec.md`
- forbidden: 기존 SPEC 의 수정, `skills/**`, `CLAUDE.md`, `docs/decisions-2026-11-08.md`, 실제 HOME, 네트워크

## Plan

1. 테스트 먼저: `tests/spec_card_number_observe.rs` 에 AC-1..5 의 테스트를 쓰고, 알림이 없어 AC-1·AC-2 가 실패하는 것을 봐요(RED).
2. `src/spec_doc.rs` 에 `CARD_NUMBERS_FROM = "2026-11-25"` 상수와, 번호를 세는 작은 함수와, `body` 뒤에서 한 번 부르는 `card_number_observe` 알림을 더해요. 날짜 비교는 `INTERVIEW_FROM` 과 같게 문자열 순서로 해요(`created`·`approval.at` 은 실제 날짜일 때만 읽혀요).
3. guard-mutant: 날짜 비교, 낱말 시작 검사, 경고 등급, 문장 핀을 하나씩 지운 사본이 테스트에서 실패하는지 봐요.
4. observe 기간: 이 SPEC 이 착지한 뒤 `2026-11-25` 이후 만들거나 개정한 SPEC 에서 알림이 몇 번 나왔고 그중 거짓 양성이 몇이었는지 `## Result` 에 이어 적어요(sdd 규칙 11). 오류로 올리는 일은 그 기록과 사용자 답 뒤에 따로 기록한 결정으로 해요.

## Risks

- 경계 날짜가 착지일보다 이르면 이미 쓰는 중인 SPEC 이 알림을 받을 수 있어요. 알림은 경고라 아무것도 막지 않고, 날짜는 한 줄 상수라 고치기 쉬워요.
- 파일 전체를 세어 근거 칸의 번호도 알림에 들어가요. 사용자가 (a) 를 고르면 칸 구분이 필요해져요. 그때 후속 SPEC 에서 해요(D-2).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 알림은 `Severity::Warning` 이고 오류 등급은 더하지 않아요 | sdd 규칙 11: 새 검사는 observe 로 시작해요. 경고는 `card_spec_link`·`spec_close`·`bundle` 어느 곳의 통과 판정에도 쓰이지 않고 오류만 쓰여요(`doc.errors() > 0`) | 상수와 등급 한 줄을 지워요 |
| D-2 | 칸 구분 없이 파일 전체를 세요 | 규칙 10 은 "SPEC 에 없음"이라 칸을 가리지 않아요. 칸별 허용은 사용자가 정할 일이라 이 SPEC 에서 안 정해요 | 후속 SPEC |
| D-3 | 경계는 `created` 또는 `approval.at` 이 `2026-11-25` 이상이에요 | 새 SPEC 은 `created` 로, 개정 SPEC 은 다시 받는 승인의 `approval.at`(sdd 규칙 6)으로 잡혀요. 기존 SPEC 을 고치지 않고 새 규칙의 시작일을 세울 수 있어요(리더 카드 노트 7726 의 예시 기준) | 상수를 바꿔요 |
| D-4 | 번호를 세는 함수를 `spec_doc.rs` 안에 새로 두고 `bundle.rs` 의 것을 빌려 쓰지 않아요 | `card_number_spans` 는 다른 카드가 쥔 비공개 함수라 공유 인터페이스를 이 카드가 바꾸지 않아요. 같은 낱말 시작 규칙(`C-`·`c-` 와 숫자, 앞이 영문자·숫자면 아님)을 써요 | 두 구현을 하나로 합치는 후속 카드 |

## Result

(구현 전 초안이에요. 승인 전이라 아직 채우지 않았어요.)
