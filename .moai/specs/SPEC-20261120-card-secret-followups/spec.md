---
id: SPEC-20261120-card-secret-followups
title: "Card secret screen follow-ups: more name forms, the reclaim reason, values after other separators"
status: draft
revision: 1
created: 2026-11-20
---

# Card secret screen follow-ups: more name forms, the reclaim reason, values after other separators

## Intent

C-122 의 tier 2 감사(판정 노트 3297·3298·3357·3396)가 남긴 후속 네 항목을 한 화면 안에 담아요. 기준 화면은 SPEC-20261119-card-secret-screen (개정 3) 이 만든 `src/card/input.rs` 의 `secret_words` 하나이고, 이 조각은 그 규칙의 빠진 꼴을 메워요.

- N3 (노트 3298): 등호 앞 이름이 비밀 낱말을 품고 있어도, 등호 앞뒤에 공백이나 탭이 있거나(이름과 등호가 서로 다른 낱말), 이름이 따옴표로 감싸져 있거나, 비밀 낱말이 이름 조각과 정확히 같지 않고 안에 들어 있으면(카멜 표기·붙여 쓴 꼴·복수형·접두) 통과해요. 이름 읽기를 세 가지로 넓혀요.
- N5 (노트 3298): `card reclaim` 의 `--reason` 은 아직 옛 검사(`redact::redact_outgoing` 단독, `src/card/dispatch.rs` 의 `checked_reason`)를 써서, 이름-등호-값 꼴의 비밀 이름이 통과해요. 같은 한 화면으로 바꿔요.
- N7 (노트 3357): 값 모양 검사는 리더가 정한 구분자(등호, 콜론, 쉼표, 공백) 뒤만 봐요. 세미콜론, 슬래시, 세로줄, 골뱅이, 물음표, 앰퍼샌드, 점, 더하기, 샵, 역슬래시, 여는 괄호 뒤의 자격 증명 모양 값은 통과해요. 구분자 목록에 이 열한 글자를 더해요.
- N8 (노트 3396): C-122 SPEC 의 Result 는 짧은 접두 조각이 쉼표·콜론을 넘어 이어지는 두 꼴이 거절된다고 적혀 있어요. 감사 3회차의 화면은 의도대로 그 두 꼴을 통과시켜요(진짜 키는 구분자를 품지 않아요). 문장을 동작에 맞게 고쳐요.

사용자 결정(카드 C-122 노트 3143, 선택지 3)의 방향 — 비밀 낱말이 든 이름-등호-값 꼴은 값과 상관없이 거절 — 을 그대로 이어가요. 오탐의 대가도 같아요: 비밀 낱말을 품은 낱말이 등호와 이어지면 무해한 문장도 거절되고, 값을 띄어 쓰면 통과해요.

## Non-goals

- 진짜 비밀 탐지기가 아니라 모양 검사예요(SPEC-20261119-card-secret-screen 과 같은 한도).
- N3 의 나머지 꼴 — 닮은 글자·전각·폭 없는 공백, 대괄호 인덱스, 점 뒤 다른 낱말, 줄바꿈이 등호 앞, JSON·YAML 전체 꼴 — 은 이 조각에서 다루지 않아요(별개 후속).
- 장부에 이미 있는 행의 다시 검사·고침은 하지 않아요(결정 3143): 새 쓰기만 검사해요.
- 공유 화면 `redact::redact_outgoing` 의 모양 목록은 바꾸지 않아요.

## Requirements

- REQ-1: The card text screen shall refuse an assignment also when blanks (spaces or tabs) stand between the name and the equals sign, reading the word right before the equals sign as the name when the equals sign starts its own word.
- REQ-2: The card text screen shall read the name of an assignment with the surrounding quote characters stripped, so a quoted name that holds a secret piece before an equals sign is refused.
- REQ-3: The card text screen shall refuse an assignment whose name holds one of the secret words, or the two words `api` and `key` joined, right behind a camel hump (an upper-case letter that follows a lower-case one), while an all-lower-case joined or plural name keeps passing as today.
- REQ-4: The card text screen shall let a text through when a secret-shaped word appears in it without an equals sign after it, and when the word before a spaced equals sign holds no secret shape.
- REQ-5: The `--reason` of `card reclaim` shall pass through the same one screen (`input::secret_words`) as the note and the criterion fields, with the refusal kind `invalid_input`, while a clean reason keeps passing that check as before.
- REQ-6: The card text screen shall look at the text after each of the characters `; / | @ ? & . + # \ (` inside a word as a value of its own, as it already does for `= : ,` today, and at a run of pieces joined by dots also whole, so a credential-shaped value after any of them is refused and a three-part dotted token stays refused after another separator.
- REQ-7: The card text screen shall keep answering one long word without a space in time that grows in step with its length after the new separators are added, so a word of a million bytes with the new separators is answered in seconds, not minutes.
- REQ-8: The Result section of SPEC-20261119-card-secret-screen shall state that a short prefix piece that runs past a comma or a colon passes on purpose, because a real key holds no separator.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): 등호 앞뒤 공백·탭 꼴, 따옴표 이름 꼴, 카멜·붙여쓴·복수형 이름 꼴이 노트 본문, 제목, 본문, `--ac` 네 표면에서 모두 거절되고, 같은 이름 꼴이 등호 없이 쓰인 평범한 문장과 등호 앞 이름이 수수한 문장은 통과해요.
  verify: `cargo test --locked --offline --test card_secret_followups name_forms`
- AC-2 (REQ-5): `card reclaim` 의 사유에 이름-등호-값 꼴의 비밀 이름이 들면 거절(`invalid_input`, 비밀 모양 문구)되고, 깨끗한 사유는 그 검사를 그대로 통과해요.
  verify: `cargo test --locked --offline --test card_secret_followups reclaim_reason`
- AC-3 (REQ-6, REQ-7): 열한 구분자 각각 뒤에 오는 자격 증명 모양 값이 네 표면에서 모두 거절되고, 구분자가 들어간 평범한 문장은 통과하며, 새 구분자가 섞인 긴 한 낱말이 시간 상한 안에 답해요.
  verify: `cargo test --locked --offline --test card_secret_followups separators`; a long name with a camel hump every two characters (64 KiB and just under 1 MiB, joined and spaced) answers in time too: `cargo test --locked --offline --test card_secret_followups camel_long_name`
- AC-4 (REQ-8): SPEC-20261119-card-secret-screen 의 Result 가 두 통과 꼴을 사실대로 적어요.
  verify: `python3 -I -c "import pathlib; t=pathlib.Path('.moai/specs/SPEC-20261119-card-secret-screen/spec.md').read_text(encoding='utf-8'); assert '재감사 N8' in t and '거절에서 통과로 바뀌었어요' in t and '놓치는 키 모양은 없어요' in t"`

## Scope

- write: `src/card/input.rs`, `src/card/dispatch.rs`, `src/card/work.rs`, `tests/card_secret_followups.rs`,
  `.moai/specs/SPEC-20261120-card-secret-followups/spec.md`
- read-only: `src/redact.rs`, `.moai/specs/SPEC-20261119-card-secret-screen/spec.md`(AC-4 에서 문장 하나 고침),
  `tests/card_cli.rs`, `tests/card_work_cli.rs`
- forbidden: 실제 HOME과 원장, 네트워크, 실제 원격 저장소, 공유 화면 `redact::redact_outgoing` 의 모양 목록 바꾸기

## Plan

1. 시험 먼저(RED): `tests/card_secret_followups.rs` — 이름 꼴 세 가지·reclaim 사유·남은 구분자의 네 표면 표 시험과 새 구분자가 섞인 긴 한 낱말 시간 시험. 자격 증명 모양은 조각으로 이어 붙여 만들어요.
2. `src/card/input.rs`: 이름 읽기 넓히기(공백 뒤 등호의 앞 낱말, 따옴표 벗기기, 하위 문자열 대조)와 구분자 목록에 열한 글자 더하기; `secret_named_assignment`·`word_is_refused` 의 설명 주석을 같이 고쳐요.
3. `src/card/dispatch.rs`: `checked_reason` 이 `redact_outgoing` 대신 `input::secret_words` 를 쓰게 하고, 거절문을 다른 표면과 같은 모양으로 맞춰요.
4. `src/card/work.rs`: 화면 주석이 새 꼴을 말하게 한 줄 고쳐요.
5. 문서: `.moai/specs/SPEC-20261119-card-secret-screen/spec.md` 의 Result 문장 정정(REQ-8)과 이 SPEC 의 Result 채우기.

## Risks

- 카멜 혹 대조의 경계: `myTokenCount=1` 은 거절하지만 올소문자 붙여쓰기나 복수형(`tokens=3`)은 기존 계약(SPEC-20261119-card-secret-screen REQ-2)대로 통과해요. 카멜 낱말이 등호와 이어지만 않으면 문장은 통과해요.
- 구분자가 늘면 평범한 문장(전자우편 주소, 경로, 마크다운 꼴)의 조각이 많아져요: 통과 줄로 지키고, 시간 상한 시험이 선형을 지켜요.
- 설치된 `moai`(옛 화면)은 이 문서의 꼴을 적은 카드 노트를 거절할 수 있어요: 노트는 풀어 쓰고, 시험은 이 가지에서 빌드한 `moai-cli` 로 돌려요.

## Result

- 구현(2026-11-19, agent-18): `src/card/input.rs` — REQ-1 `secret_words` 가 등호로 시작하는 낱말의 앞 낱말을 이름으로 읽어요; REQ-2 `secret_named_assignment` 가 이름의 따옴표를 벗겨요; REQ-3 `camel_hump_secret` 가 카멜 혹 뒤 비밀 낱말과 이어진 api·key 를 거절해요(복수형·올소문자 붙여쓰기는 기존 계약대로 통과); REQ-6·7 `word_is_refused` 의 구분자가 열한 글자 더해지고 점 이음 토큰은 점을 지켜 보는 둘째 훑기로 Whole 하게 봐요(세 점 토큰이 다른 구분자 뒤에서도 거절 — 점을 구분자에 넣는 첫 시도가 기존 card_cli 시험의 점 토큰 거부를 깨뜨려 둘째 훑기를 더했어요). `src/card/dispatch.rs` — REQ-5 `checked_reason` 이 `input::secret_words`·`secret_refusal` 로 같은 화면을 써요. `src/card/work.rs` — 주석 한 줄.
- 시험: `tests/card_secret_followups.rs` 4건 — RED 먼저 관측(4 failed, 노트 3491: 노트 수락·reclaim 임대 검사로 넘어감·구분자 통과·긴 낱말 0) 뒤 GREEN(4 passed). 이웃: lib `card::` 95 passed, `card_cli` 22, `card_work_cli` 25, `card_dispatch_cli` 16, `card_reclaim_cli` 4, `card_scope_widen_followup_cli` 4, `spec_cli` 18 — 합계 115 passed 0 failed.
- 검사: `cargo fmt --check` 0, `rustfmt --edition 2021 --check src/ops/*.rs` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0, 파이썬 네 모듈(repo_scans·source_manifest·local_check·skill_references) 121 OK — repo_scans 는 `CARGO_TARGET_DIR`(빌드 폴더) 환경변수가 필요해요(첫 실행 19 실패는 바이너리 경로 문제). `spec check` errors 0 warnings 0.
- (첫 판의 남은 것, 아래 수리로 닫음) AC-4(REQ-8, C-122 SPEC Result 문장 정정)는 그 SPEC 폴더가 C-122 예약(generation 4)에 있어 이 조각에서 하지 못했어요 — 리더가 범위를 넘겨 주면 한 줄 고침으로 이어요. 돌리지 않은 것: 전체 cargo test, CI(이 가지는 PR 없음), Windows.

### 수리 (2026-11-20, tier 2 감사 FAIL 의 B1·B2·N-d)

- B1 (제곱 시간): `camel_hump_secret` 가 혹마다 이름의 나머지를 새 문자열로 모았어요(`'aB'`×n 이름은 64 KiB 에서 한 분이 넘었어요). 이름을 한 번만 ASCII 소문자로 바꾸고(바이트 위치가 그대로라 혹의 위치로 그 사본을 둘로 가름) 혹마다 짧은 조각 둘만 비교해요. RED 먼저: 고치기 전 코드에서 `camel_long_name_stays_within_the_time_bound` 가 `size 65536 joined: no answer within 15 s` 로 실패(heavy-slot, 이 워크트리의 빌드 폴더) → 고친 뒤 `card_secret_followups` 5 passed(5.14초).
- B2 (뮤턴트 생존 3곳): `myApikey=1`(apikey 갈래), `myapiKey=1`(소문자 api 뒤 Key 혹 갈래), 따옴표 이름 뒤 띄어 쓴 등호(`strip_quotes`)를 `name_forms_…` 시험의 거절 목록에 더했어요. 뮤턴트 3개(M1 apikey 갈래를 `false` 로, M2 둘째 갈래를 `false` 로, M3 `strip_quotes` 를 뺌)가 모두 사살됐어요 — 각각 `name_forms_blanks_quotes_and_camel_are_refused_on_every_surface` 가 패닉으로 실패(컴파일 오류 0)했고 복원 뒤 `git status` 가 깨끗했어요.
- N-d (REQ-8, AC-4): C-122 SPEC 의 Result 문장은 이 수리의 팁 머지(`6a8fa9cacf`)가 이미 고쳐 둔 채로 들어왔어요(재감사 N8: 두 꼴이 거절에서 통과로 바뀌었고 진짜 키는 구분자를 품지 않아 놓치는 키 모양이 없다는 문장). 같은 문장을 중복으로 쓰지 않고 그 파일은 팁 그대로 두었어요(충돌은 팁 쪽 문장을 택해 풀었고, 머지 뒤 그 파일의 팁과의 차이는 0줄). AC-4 의 확인 명령은 팁 문장을 읽어요.
- 돌리지 않은 것·남긴 것: N-a(띄어 쓴 꼴이 `=` 앞 낱말 전체를 이름으로 읽음)와 N-b(과잉 거절 `if token == null`, `token => x`, `maxTokens=4096`)는 감사가 막지 않는 항목이라 이번 수리에서 손대지 않았어요.
