---
id: SPEC-20261124-card-number-subject-pin
title: "trailer-check pins the card-number rule on commit subjects and on quoted ledger output"
status: draft
revision: 2
created: 2026-11-24
branch: spec/20261125-card-number-subject-pin
parent: SPEC-20261120-commit-message-card-numbers
---

# trailer-check pins the card-number rule on commit subjects and on quoted ledger output

## Intent

`moai trailer-check` 는 이미 커밋 제목과 본문의 카드 번호 꼴을 거절해요(`--card-numbers-since`, 상위 SPEC 으로 착지). 그런데 제목 쪽은 실패 방향 테스트가 없어서 제목을 읽는 줄이 빠져도 모든 테스트가 통과해요(감사 비차단 발견). 이 조각은 제목 쪽 실패와 통과를 테스트로 고정하고, 실제 커밋 기록에서 허용되는 모양(원장 출력 인용 등)을 잘못 막는지 재서 결과를 남겨요. `--message-file` 쪽 제목은 기존 테스트가 이미 막고 있어서(별도 함수) 이 조각이 다루지 않아요. commit-msg 훅의 설치는 이 조각 밖이에요.

## Requirements

- REQ-1: While the card-number rule is active for a range, trailer-check shall report a failing `card_number` finding on line 1 for a card-number shape in the subject of a commit after the recorded start, wherever the shape stands on the line (after a space, right after an opening bracket, at the very start of the line or right after a colon).
- REQ-2: While the card-number rule is active for a range, trailer-check shall leave a subject that carries only a quoted ledger refusal, a SPEC id, a version, a hexadecimal SHA or an `AC-<n>` neighbour without a `card_number` finding.
- REQ-3: While the Result holds the measurement, it shall give the command, the range and the counts of the rule run over the real commit history, and the verdict on each distinct allowed shape the run flagged.

## Acceptance

- AC-1 (REQ-1, REQ-2): fixture commits whose subject carries a card number after a space, after an opening bracket, at the line start or after a colon fail with a `card_number` finding on line 1, and fixture subjects with the allowed shapes pass · verify: `cargo test --locked --offline --test trailer_check_cli -- card_numbers_in_a_subject_fail_after_a_recorded_start allowed_shapes_in_a_subject_pass`
- AC-2 (REQ-3): the Result holds the history measurement and the spec check passes · verify: `moai spec check SPEC-20261124-card-number-subject-pin --project .`

## Scope

- write: `.moai/specs/SPEC-20261124-card-number-subject-pin/**`
- write: `tests/trailer_check_cli.rs`
- read-only: everything else, `src/trailers.rs` included; a false block found by the measurement waits for a revision of this SPEC

## Plan

1. 실제 기록(2026-11-19 이후 통합 브랜치)에 규칙을 돌려 분류하고 잘못 막는 모양이 있는지 읽는다.
2. 제목 쪽 실패·통과 테스트를 `tests/trailer_check_cli.rs` 에 더한다. 번호 앞 글자는 실제 기록의 분포에 맞춘다: 괄호 바로 뒤(가장 흔해요), 공백, 줄 처음, 콜론 바로 뒤. `--message-file` 쪽 제목 테스트는 기존 테스트가 이미 막고 있어서 더하지 않았다.
3. 제목 읽는 줄과 인용 면제를 하나씩 없애는 뮤턴트로 새 테스트가 죽는지 본다.
4. 영수증과 READY 노트(뮤턴트 출력 파일 경로 포함).

## Result

머리: 이 브랜치의 코드 변경은 `tests/trailer_check_cli.rs` 하나(테스트 둘 추가, 낡은 주석 하나 고침)예요. `src/` 는 바꾸지 않았어요. 측정 도구와 원자료는 레인 워크트리의 `target/subject-pin/` 에 있어요(카드가 닫히면 사라져요).

### 빈 곳 (REQ-1)

`card_number_findings`(`src/trailers.rs`)의 제목 줄(`std::iter::once((1usize, c.subject.as_str()))`)을 빈 문자열로 바꿔도 기존 테스트가 모두 통과했어요. 기존 테스트의 카드 번호는 본문에 있었고(`card_numbers_fail_after_a_recorded_start`), 제목은 `--message-file` 테스트만 건드렸는데 그것은 다른 함수(`message_file_check`)예요. 그 낡은 주석("제목은 아래 message-file 테스트에서")도 고쳤어요. 새 테스트 `card_numbers_in_a_subject_fail_after_a_recorded_start` 가 제목의 대문자·소문자·닫힌 따옴표 뒤·여는 괄호 바로 뒤·줄 맨 처음·콜론 바로 뒤의 번호를 1번 줄 실패로 고정하고(찾은 꼴을 순서대로 읽어요), 시작 커밋이 면제되는 것과 그때 종료 코드가 1로 남는 것도 제목으로 확인해요. 처음 올린 고정 조각은 제목 셋이 모두 번호 앞이 공백이었어요(아래 감사 수리).

### 허용되는 모양을 잘못 막는지 (REQ-2, REQ-3)

고정 조각: `allowed_shapes_in_a_subject_pass` 가 제목에 SPEC id, 버전, 16진 SHA, `AC-<n>`, `SPEC-C-<n>`, 숫자 뒤 글자(`C-<n>b`), 글자 사이에 `c-<숫자>` 후보가 든 낱말과 16진 글자 줄, 따옴표 안 원장 거절 문장을 넣어 `card_number` 발견이 없음을 확인해요. SPEC id·버전·SHA·`W2-7` 줄에는 `C-<숫자>` 후보가 아예 없어서 경계 코드에 닿지 않아요. 경계에 닿는 줄은 `AC-<n>`, `SPEC-C-<n>`, `C-<n>b`, 낱말, 16진 줄, 따옴표 인용이에요.

실제 기록: `moai trailer-check 35258bb454..NeXT --card-numbers-since 35258bb454`(NeXT 끝 `0babbe3566`, 설치된 `moai 4.2648.24`, 2026-11-24 19:5xZ). 시작 커밋이 범위 밖이라 `card_numbers_from` 은 0이고 범위 전체(2558 커밋)에 규칙이 걸려요. 결과는 `exit=1`이고 `card_number` 발견 476건(커밋 376개, 서로 다른 꼴 64개, 제목 345건·본문 131건)이에요(`failing` 649 는 다른 종류의 실패 173건이 더해진 숫자예요: `missing_verified` 156, `missing_spec` 12, `malformed_session` 5. 규칙 시작 전 옛 커밋의 트레일러 이야기라 이 조각과 상관없어요). 규칙 시작(2026-11-19) 전 커밋이 대부분이라 많이 나오는 것은 당연해요.

분류(`target/subject-pin/classify.py`, 결과 `classified.tsv`): 발견마다 그 줄을 `git show -s --format=%B` 로 읽어 꼴 앞뒤 글자를 셌어요. 앞뒤 글자 쌍은 괄호 바로 뒤(323건), 공백 사이(106건), 줄 처음(18건), 그 밖(29건: 줄 끝·마침표·쉼표·세미콜론·`/` 따위)이고(합 476), 예시를 쌍마다 하나씩 읽었어요. 그 가운데 꼴 앞에 큰따옴표·작은따옴표·백틱이 있는 26줄은 전부 한 줄씩(앞 210자) 읽었고 모두 진짜 카드 번호 언급이에요(따옴표 앞에 닫힌 쌍이 있어서 짝수 개라 면제되지 않은 것). 허용되는 모양(SPEC id, 버전, SHA, `AC-<n>`, 따옴표 안 인용)으로 잘못 막힌 줄은 찾지 못했어요. 읽지 않은 것: 476건 가운데 위 26줄과 쌍마다의 예시 말고는 한 줄씩 읽지 않았고 앞뒤 글자로 분류했어요.

그래서 규칙을 고치지 않았고, `src/trailers.rs` 도 바꾸지 않았어요.

### 가드 뮤턴트 (10개, 모두 KILLED)

한 번의 슬롯 호출로 돌렸어요(첫 감사 전 7개의 결과는 `target/subject-pin/mutants-first.jsonl` 에 두었고, 감사 수리 뒤에 새 뮤턴트 셋을 더해 배치 전체 10개를 다시 돌렸어요. `target/subject-pin/mutants.py`, 판정기는 저장된 실패 로그 꼴로 먼저 테스트했고, 각 편집은 정확히 한 번임을 `--check` 로 확인했어요). 각 뮤턴트는 `src/trailers.rs` 를 바꿔 다시 빌드하고 `cargo test --test trailer_check_cli` 를 돌렸고, 파일은 바이트 단위로 되돌려 비교했으며 끝의 복원 실행은 9개 통과예요. 출력 파일: `target/subject-pin/mutants.log`(사람이 읽는 줄), `target/subject-pin/mutants.jsonl`(뮤턴트마다 한 줄). `git status --short` 는 되돌린 뒤 비어 있어요.

- guard-mutant: card_numbers_in_a_subject_fail_after_a_recorded_start | the subject line of `card_number_findings` replaced by an empty string | exit 101
- guard-mutant: allowed_shapes_in_a_subject_pass | the quoted-output exemption (odd quote count) never applies | exit 101
- guard-mutant: card_numbers_in_a_subject_fail_after_a_recorded_start | the quoted-output exemption inverted (even quote count exempts) | exit 101
- guard-mutant: allowed_shapes_in_a_subject_pass | the leading alphanumeric boundary removed so `AC-<n>` matches | exit 101
- guard-mutant: allowed_shapes_in_a_subject_pass | the leading hyphen boundary removed so `SPEC-C-<n>` matches | exit 101
- guard-mutant: allowed_shapes_in_a_subject_pass | the trailing alphanumeric boundary removed so `C-<n>b` matches | exit 101
- guard-mutant: card_numbers_in_a_subject_fail_after_a_recorded_start | the start commit no longer exempt | exit 101
- guard-mutant: card_numbers_in_a_subject_fail_after_a_recorded_start | the opening bracket added to the leading boundary so `(<card number>)` is skipped | exit 101
- guard-mutant: card_numbers_in_a_subject_fail_after_a_recorded_start | a shape at the very start of the line skipped | exit 101
- guard-mutant: card_numbers_in_a_subject_fail_after_a_recorded_start | the colon added to the leading boundary so a shape right after a colon is skipped | exit 101

새 셋은 새 제목 테스트만 잡아요(`mutants.jsonl` 의 `failed` 가 그 테스트 하나예요). 다른 뮤턴트 가운데 `quote-exemption-inverted`, `leading-alphanumeric` 과 `start-commit` 은 기존 테스트도 함께 실패했어요.

### 감사 수리 (개정 2)

첫 감사(머리 `7d7b2e96cb`)는 새 테스트가 진짜 가드임을 확인했고(뮤턴트 7줄 KILLED), 막는 것 하나와 작은 것을 남겼어요. 수리(`cbce5dcff4`):

- 막는 것: 제목 픽스처 셋이 모두 번호 앞이 공백이었어요. 실제 기록에서 가장 흔한 꼴은 괄호 바로 뒤(476 가운데 323)이고 줄 맨 처음(18)도 있는데 고정되지 않았어요. 제목 픽스처에 괄호 바로 뒤, 줄 맨 처음, 콜론 바로 뒤를 더하고 뮤턴트 셋(괄호, 줄 처음, 콜론)을 더했어요. 새 셋은 새 제목 테스트만 죽여요.
- 허용 예: 버전·16진 SHA·`W2-7` 은 `C-` 후보가 없어 경계 코드에 닿지 않았어요. 닿는 예(글자 사이의 `c-<숫자>`가 든 낱말, 16진 글자 사이의 `c-<숫자>`)를 더하고 테스트 주석에 "후보 없음" 줄을 적었어요.
- 종료 코드: 시작 커밋 면제 확인이 종료 코드를 버렸어요. 이제 1을 확인하고 남은 발견이 어느 꼴인지 순서대로 비교해요.
- Plan 2단계 글을 사실대로 고쳤어요(`--message-file` 제목 테스트는 하지 않았어요). Result 의 스크래치 폴더 이름을 카드 번호에서 온 이름에서 중립 이름 `target/subject-pin/` 으로 바꿨어요(폴더도 같은 이름으로 옮겼어요).
- 테스트 파일과 경로가 바뀌는 문제는 아래 한계에 적었어요.

### 남은 위험과 한계

- 대기 줄(착지 대기)이 착지하면 이 조각이 쓰는 `--test trailer_check_cli` 와 `tests/trailer_check_cli.rs` 는 `tests/card/trailer_check_cli.rs` 의 모듈로 바뀌어서 위 AC 의 verify 명령과 Scope 경로가 달라져요. 지금은 그대로 두었어요(레인은 팁을 머지하지 않아요). 착지하는 쪽이 명령과 경로를 맞춰요.

- 이 조각은 규칙을 켜 두는 쪽(`local_check`·CI 에 `--card-numbers-since` 와 시작 커밋 기록)을 하지 않았어요. 그것은 상위 SPEC 의 Non-goals 이고 리더의 기록 뒤 몫이에요. 읽은 사실: `tools/local_check.sh` 에 `trailer-check` 호출은 0건(`grep -c`), `.github` 와 `tools`, `install` 에서 `trailer-check` 를 부르는 파일은 `install/commit-msg` 와 `install/git-hooks/pre-push` 둘뿐이고(`grep -rln`), 푸시 훅은 `--card-numbers-since "$remote_sha"` 로 규칙을 켜요(이 조각이 바꾸지 않았어요).
- commit-msg 훅의 설치와 `moai project-update` 단계는 이 조각 밖이에요(다른 브랜치 `spec/commit-msg-hook-20261120` 에 있고 NeXT 에 없어요). 카드 노트가 더한 나머지 후속(브랜치의 SPEC 과 `Moai-Spec` 불일치 검사, `Moai-Session` 에 32자리 인스턴스 id, SPEC id 의 카드 번호 거절)도 이 조각이 다루지 않았어요.
- 줄 안 따옴표를 홀짝으로 세는 면제는 한 줄에 따옴표가 홀수 개 흩어진 산문(예: 인치 기호)에서 진짜 번호를 놓칠 수 있어요(닫힌 쪽으로는 실패하지 않는 방향, 알려진 한계).
- 측정은 설치된 `moai 4.2648.24` 로 했고, 이 브랜치의 새로 빌드한 바이너리로는 하지 않았어요(규칙 코드는 이 조각이 바꾸지 않았어요). 테스트와 뮤턴트는 이 브랜치의 빌드로 돌렸어요.

### 기준 (AC)

- AC-1: met. `cargo test --locked --offline --test trailer_check_cli` 9개 통과(새 테스트 둘 포함), `clippy --test trailer_check_cli -D warnings` 통과, 뮤턴트 10개 모두 죽었어요(개정 2).
- AC-2: met. 위 측정과 `moai spec check` 오류 0·경고 0.
