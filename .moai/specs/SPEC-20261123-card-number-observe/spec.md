---
id: SPEC-20261123-card-number-observe
title: "spec check counts card numbers of new and revised SPECs as an observe notice"
status: approved
revision: 2
created: 2026-11-23
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7747 (2026-11-23T17:44:46Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님): 판 2 승인, 바뀐 것은 Scope 의 tools/repo_scans.json 등록 항목과 D-5, Q3 뿐이고 요구·기준은 판 1 과 같아요. 레인은 그 노트를 인용해 옮겨 적었어요"
  digest: "sha256:93d92f11a893b2597a0787b479520c215126c6c443831f96ddd00bc382275a00"
---

# spec check counts card numbers of new and revised SPECs as an observe notice

## Intent

`internal/sdd.md` 규칙 10 은 SPEC 에 카드 번호(`C-` 와 숫자)가 절대 없다고 말하는데, 이 저장소의 SPEC 347개 중 152개의 `spec.md` 본문에 758번 나와요(`moai spec check --all` 이 읽은 개수와 `rg` 로 센 수, 2026-11-23). 그런데 SPEC 쪽에서 이를 보는 도구는 앞머리 `cards` 키 거절 하나뿐이라 같은 어긋남이 감사마다 되풀이돼요. 규칙을 좁힐지(근거 칸 허용)나 기존 SPEC 을 고칠지는 사용자 결정으로 올라갔고(카드 노트 7724·7726), 그 답이 오기 전까지 규칙 글은 그대로 두고 이 SPEC 이 두 가지만 더해요. 하나, 새로 만들거나 개정해 다시 승인한 SPEC 의 카드 번호를 `moai spec check` 가 observe 알림(경고, 오류 아님)으로 세요. 둘, 규칙 10 문장이 몰래 바뀌지 않게 문장 핀 테스트 하나를 더해요. 기존 SPEC 은 건드리지 않아요.

## Interview

- Q1: 규칙 10 을 좁히는 선택지 (a) 를 지금 정하나요? → A: 아니요. (a) 는 사용자가 2026-11-17 에 직접 정한 규칙(`docs/decisions-2026-11-08.md:134`)을 좁히는 일이라 사용자에게 올렸고, 그 답 전에는 규칙 글을 바꾸지 않아요(리더 결정 카드 노트 7726).
- Q2: 그 사이 무엇을 하나요? → A: (d) 규칙을 그대로 두고 `spec check` 가 새 SPEC·개정 SPEC 의 카드 번호를 observe 알림으로 세고, 규칙 10 문장 핀 테스트를 더해요. 기존 SPEC 은 바꾸지 않아요(리더 결정 카드 노트 7726).
- Q3: 새 테스트 파일이 저장소 트리를 읽는데 등록은 어디에 하나요? → A: `tools/repo_scans.json` 에 해요. 쓰는 순서는 리더가 알려 줘요(리더 결정 카드 노트 7729). 그래서 판 2 로 Scope 에 그 파일의 등록 항목을 더해요.

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

- write: `src/spec_doc.rs`(알림 하나와 날짜 상수), `tests/spec_card_number_observe.rs`(새 파일), `.moai/specs/SPEC-20261123-card-number-observe/spec.md`, `tools/repo_scans.json`(새 테스트 파일의 등록 항목만, 리더가 정한 차례에, D-5)
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
| D-5 | 판 2: Scope 에 `tools/repo_scans.json` 의 등록 항목만 더해요 | 새 테스트 파일이 `.moai/specs` 와 `sdd.md` 를 읽어서 `tests/repo_scans_table.rs` 의 `every_walking_test_file_is_listed_or_exempt` 와 `every_file_reader_has_an_always_scan_or_an_exemption` 가 실패해요(이 브랜치에서 `cargo nextest run --offline --locked --test repo_scans_table` 로 읽음). 리더가 판 1 승인 노트 7729 에서 등록과 쓰는 순서를 정했어요. 범위가 digest 에 묶여서 판을 올려요 | 등록 항목을 지우고 판 1 로 돌아가요 |

## Result

브랜치 `spec/card-number-observe-20261123` 에서 판 1 승인(리더 노트 7729) 뒤 RED → 구현 → guard-mutant 를 했고(구현 커밋 `5b5222ab8d`, 기준 `466b59fc62`), 판 2 승인(리더 노트 7747) 뒤 NeXT `29e15b53cc` 를 한 번 합치고 `tools/repo_scans.json` 에 등록 항목 하나를 더했어요(등록 커밋 `e96820dda9`). 시각은 리더 노트의 시각이에요.

- RED(구현 전): `cargo nextest run --offline --locked --test spec_card_number_observe` 가 6개 중 4개 실패(알림 없음), 2개 통과(실제 SPEC 읽기·문장 핀은 처음부터 맞는 특성화), exit 100.
- 구현 뒤: 같은 명령이 6개 모두 통과. 합친 뒤 `cargo nextest run --offline --locked --test repo_scans_table --test spec_card_number_observe` 가 28개 모두 통과, exit 0. 등록 전에는 `repo_scans_table` 의 `every_walking_test_file_is_listed_or_exempt` 와 `every_file_reader_has_an_always_scan_or_an_exemption` 가 이 테스트 파일 때문에 실패했어요.
- guard-mutant 6개를 하나씩 넣어 모두 exit 100: ① 날짜 비교 삭제 → 3개 실패, ② 낱말 시작 검사 삭제 → 3개 실패, ③ 경고를 오류로(`doc.warn` → `doc.error`) → 2개 실패, ④ `approval.at` 갈래 삭제 → 1개 실패, ⑤ 새 형식 id 검사 삭제 → 1개 실패, ⑥ 핀 테스트의 문장 비교 삭제 → 1개 실패. 매번 `git checkout` 으로 되돌렸어요.
- `check affected` 첫 실행(head `1c66b05cf7`)은 `state_folder_literal_scan` 1개로 FAIL(3832개 중 1개, 새 테스트 파일이 상태 폴더 이름을 글자 그대로 두 번 써서 `tests/fixtures/state_folder_literals/allow.tsv` 밖이라고 걸림, 로그 `04-test.log`). 이름을 `spec_doc::SPECS_DIR` 상수로 바꾸고 주석에서 빼서 고쳤고, `cargo nextest run --offline --locked --test state_folder_literal_scan --test spec_card_number_observe --test repo_scans_table` 34개 통과(exit 0)를 읽은 뒤 새 head 에서 `check affected` 를 다시 돌렸어요.

- AC-1: met — `a_new_spec_with_card_numbers_gets_one_observe_warning` 통과. 경고 하나(번호 둘, 첫 줄 번호와 개수만 말하고 SPEC 글은 안 담아요), 낱말 안의 `AC-1`·`SPEC-…`·`ABC-5` 는 안 세고, 번호를 깬 같은 글과 오류 수가 같아요. 보조 `only_a_card_number_at_a_word_start_counts` 통과.
- AC-2: met — `a_revised_spec_is_judged_by_its_approval_date` 통과. `created` 가 옛날이어도 `approval.at` 이 `2026-11-25` 면 경고 하나, `2026-11-24` 면 없어요.
- AC-3: met — `older_go_and_undated_specs_get_no_notice` 통과. 경계 하루 전(`2026-11-24`), 날짜 없음, Go 세대 SPEC(늦은 날짜)은 알림이 없고 경계일(`2026-11-25`)은 하나예요.
- AC-4: met — `the_real_specs_before_the_date_get_no_notice` 통과(이 저장소의 `.moai/specs/*/spec.md` 를 읽어 경계 전 SPEC 100개 넘게 알림 0, 경고 등급 확인). 이 브랜치의 빌드 `target/debug/moai-cli spec check --all` 은 `count 354, errors 0, warnings 26, card_number_observe 0` 이고, 이 변경이 없는 설치본 `moai spec check --all` 도 같은 `354, 0, 26` 이에요. 제한: 테스트는 번호를 깬 사본과 오류 코드를 비교하는데, 승인된 SPEC 의 요구·기준·범위 안 번호를 깨면 내용 digest 가 바뀌어서 그 코드 `approval_digest` 만 비교에서 뺐어요(테스트의 편집이지 알림이 아니에요).
- AC-5: met — `rule_10_sentence_is_pinned` 통과. `sdd.md` 규칙 10 의 문장을 낱말 그대로 요구하고, `never appear` → `may appear`, `cards are personal` → `cards are shared`, 규칙 제목 바꿈, 빈 글 네 사본을 모두 거절해요.

observe 기록(sdd 규칙 11): 경계 `2026-11-25` 가 아직 오지 않아서 새·개정 SPEC 에서 알림이 몇 번 나왔는지와 거짓 양성 수는 읽을 수 없어요(not_run). 기존 SPEC 354개에서는 0번이에요(위 `spec check --all`). 알림을 오류로 올리거나 규칙 10 을 좁히는 일은 이 SPEC 이 아니고 사용자 답과 observe 기록 뒤에 따로 정해요.

후속(테스트만, 동작·요구·기준 그대로): 감사의 비차단 N1(`src/spec_doc.rs` 의 `card_numbers_in` 루프 경계 `i + 2` → `i + 3` 뮤턴트가 살아남음)을 막으려고 `only_a_card_number_at_a_word_start_counts` 에 줄 끝에서 끝나는 한 자리 번호 표본 하나를 더했어요(커밋 `3eb31ca3ef`). 그 뮤턴트를 넣고 `cargo nextest run --offline --locked --test spec_card_number_observe` 를 돌리면 이 테스트가 실패해 exit 100, 되돌린 뒤 6개 모두 통과(exit 0)예요.
