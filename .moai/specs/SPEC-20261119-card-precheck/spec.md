---
id: SPEC-20261119-card-precheck
title: "Card precheck: already-done and decision collisions visible before dispatch"
status: draft
revision: 1
created: 2026-11-19
---

# Card precheck: already-done and decision collisions visible before dispatch

## Intent

리더가 이미 병합된 카드를 배차한 일이 두 번 있었어요(C-88 은 #3151, C-92 는 #3187 에서 이미 병합; 카드 C-96 리더
근거 note), 그리고 카드 기준이 기록된 사용자 결정과 부딪힌 일이 세 번 있었어요(C-87 기준 2 대 결정 6878265192,
C-63 N1 대 #2936 AC 13, C-93 D4). 두 경우 모두 레인이 일을 시작한 뒤에야 알았어요. `moai card precheck <card>` 는
리더가 카드를 ready 하거나 배차하기 전에 한 번 실행하는 읽기 전용 검사로, (1) 통합 브랜치 커밋 중 메시지가 카드를
이름하는 병합 커밋·PR, 원장에 연결된 SPEC, 카드를 이름하는 SPEC 폴더, (2) 카드의 쓰기 범위 경로나 기준 문구와
겹치는 결정 note 를 한 JSON 보고로 보여줘요. 아무것도 막지 않고 알려주기만 해요.

## Non-goals

- 관문 아님: `card ready`/`take` 의 동작은 그대로예요. precheck 를 ready/dispatch 절차에 넣는 글은 카드 C-103 이
  옮겨 받았어요(기준 3 withdraw, 리더 결정 note 2368).
- 네트워크 없음: 열린 PR 이나 GitHub 검색을 읽지 않아요. 병합된 PR 은 통합 브랜치의 병합 커밋 제목
  (`Merge pull request #N from …`)에서 읽어요.
- `src/ops/card.rs` 의 서브액션 배선: C-43 예약 안에 있어요. 이 SPEC 은 로직(`src/card/precheck.rs`)과 시험을
  만들고, 배선은 리더가 C-43 후속 조각 뒤 넘겨줘요(리더 결정 note 2368).

## Requirements

- REQ-1: When the leader runs `moai card precheck <card>` in a checkout of the project, the command shall read the local ledger and the local git history of the integration ref and print one JSON report that lists, for that card, every commit on the ref whose message (subject, body or trailer) names the card by `C-<n>`, `card/<n>` or `card <n>` as a word, each with its sha, its PR number when the subject is `Merge pull request #N from …`, and the part that matched; the SPEC id linked to the card in the ledger; and every SPEC folder under `.moai/specs/` whose `spec.md` names the card the same way.
- REQ-2 (REQ-1): IF the git repository, the integration ref or the ledger cannot be read, THEN the command shall still answer `ok` with the failure named inside the report and every section it could still fill.
- REQ-3: When the report is built, the command shall also list every ledger note of kind `decision` — on any card, the card itself included — whose body shares a path covered by the card's write scope (glob `**` semantics, either direction: a note path inside a scope pattern, or a scope pattern's literal part named in the note) or a distinctive token of the card's criteria (a token that contains a digit, `-`, `_`, `/`, `#` or `::` and is at least 3 characters, or an alphanumeric token of at least 5 characters that is not a function word, matched case-insensitively as a whole word), each with its card, note id, note ref, author and every key that matched.
- REQ-4 (REQ-3): IF the ledger is registered but holds no such note, or the ledger is absent, THEN the command shall leave the decision-note section an empty list and name the ledger's state in the report.
- REQ-5: When the command runs, it shall write nothing: the ledger's bytes and modification times and the repository's `HEAD` are unchanged after the run. · failure: none
- REQ-6: IF the card argument is neither `C-<n>` nor `#<n>`, or names no card in an existing ledger, THEN the command shall refuse with the input error and exit code the other read-only card actions use.

## Acceptance

- AC-1 (REQ-1): 임시 저장소에 카드를 이름하는 커밋 세 가지(제목 `(C-<n>)`, 병합 커밋 `Merge pull request #N from modu-ai/card/<n>-slug`, `Moai-Spec` 트레일러가 연결 SPEC)와 소음(`C-<n>0`, 다른 카드 번호)을 만들면 보고는 세 항목만 나열하고 PR 번호와 매치 위치를 알려주며, 연결 SPEC 과 카드를 이름하는 SPEC 폴더를 담아요. 소음은 나열되지 않아요.
  verify: `cargo test --offline --locked --test card_precheck_cli`
- AC-2 (REQ-3): 결정 note 세 개(범위 경로 공유, 기준 토큰 공유, 겹침 없음)를 만들면 보고는 앞의 두 개만 카드·note ref·작성자·매치 키와 함께 나열하고 세 번째는 나열하지 않아요.
  verify: `cargo test --offline --locked --test card_precheck_cli`
- AC-3 (REQ-2, REQ-4, REQ-5, REQ-6): 실행 뒤 원장 파일과 저장소 `HEAD`, 작업나무가 바이트와 mtime 그대로예요. 원장 없음·git 실패는 `ok` 보고 안에 이름 붙어 나오고 빈 결정 구간은 빈 목록이며, 잘못된 카드 이름은 다른 읽기 동작과 같은 거부로 끝나요.
  verify: `cargo test --offline --locked --test card_precheck_cli`
- AC-4 (REQ-1, REQ-3): 토큰화·구문 추출·경로 포함 규칙의 단위 시험이 일곱 경계(대소문자, 워드 바운더리 `C-960`, 불용어, `**` 글로브, 경로 양방향, 트레일러 매치, 빈 기준)를 고정해요.
  verify: `cargo test --offline --locked --lib precheck`

## Scope

- write: `src/card/precheck.rs`, `src/card/mod.rs`, `tests/card_precheck_cli.rs`, `.moai/specs/SPEC-20261119-card-precheck/spec.md`
- read-only: `src/ops/card.rs` (배선은 리더 전달), `src/card/spec_link.rs`, `src/card/input.rs`, `src/ledger/`, `tests/common/`
- forbidden: 실제 HOME 과 원장, 네트워크, `card ready`/`take` 동작, `src/ops/card.rs` 쓰기(C-43 예약), `skills/moai/**`(C-103)

## Plan

1. 시험 먼저: `tests/card_precheck_cli.rs` 에 AC-1..AC-3 의 통합 시험을 쓰고 RED 를 읽어요(서브액션이 아직 없어 거부).
2. `src/card/precheck.rs` 에 로직(원장 읽기, git 스캔, 매칭, JSON 조립)과 AC-4 단위 시험을 넣고 `src/card/mod.rs` 에 모듈을 더해요. 단위 시험이 먼저 GREEN.
3. 통합 시험은 배선(`src/ops/card.rs`, 리더 전달) 뒤 GREEN; 그 전에 가지만 push 해요.
4. 이웃 검사(`--test card_spec_cli`, `--test card_work_cli`, `cargo fmt --check`)를 돌려요.

## Risks

- 구문 매칭의 정밀도: 함수어 불용어와 워드 매치로 낮추지만 거짓 양성이 남을 수 있어요 — 보고는 조언이고 리더가 읽어요; 단위 시험이 규칙을 고정해요.
- git 실행 환경: 공용 git 러너(환경 정돈)를 써서 설정 주입을 피해요; 시험은 임시 저장소에서 돌아요.
- 설치된 `moai`(11-18 빌드)에 이 명령이 없어요 — 이 worktree 의 빌드产物로 시험하고, 배선은 리더가 넘겨준 뒤 PR 로 묶여요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 로컬 git 만 읽어요(병합 PR 은 병합 커밋 제목에서); gh 호출 없음 | 두 사례(#3151, #3187) 모두 로컬 커밋 메시지에 보임; 읽기 전용 유지 | 열린 PR 이 필요하면 `--json` 네트워크 모드 후속 |
| D-2 | 구문 토큰: 숫자·`-`·`_`·`/`·`#`·`::` 포함 & 길이≥3, 또는 순수 영숫자 길이≥5(불용어 제외), 대소문자 무시 워드 매치 | C-63("hooks", "#2936")과 C-87("user_chat", "Lang::En") 사례를 잡기에 충분 | 임계값 조정 |
| D-3 | 통합 브랜치 기본값은 현재 브랜치의 upstream(`@{upstream}`), `--integration-ref` 로 덮음 | 리더는 통합 브랜치 체크아웃에서 실행; 시험은 명시적 플래그 | 기본값 제거 |
| D-4 | 기준 3(mode-leader 워크플로 편집)은 이 카드에서 withdraw 하고 C-103 으로 이동 | 리더 결정 note 2368(Q2 (b)) | C-103 에서 되살리기 |

## Result

- AC-1: red(배선 대기) — `cargo test --offline --locked --test card_precheck_cli` 0 passed; 3 failed, 모두 `unrecognized subcommand 'precheck'`: 서브액션 배선은 `src/ops/card.rs`(C-43 예약)라 리더 전달 대기(리더 결정 note 2368). 논리와 시험은 `src/card/precheck.rs`·`tests/card_precheck_cli.rs` 에 다 들어 있고, 배선 조각은 `wiring.md`.
- AC-2: red(배선 대기) — 같은 시험, 같은 이유.
- AC-3: red(배선 대기) — 같은 시험, 같은 이유.
- AC-4: met — `cargo test --offline --locked --test card_precheck_cli` 는 위와 같고, `cargo test --offline --locked --lib precheck` 7 passed/0 failed(대소문자·`C-960` 워드 경계·불용어·`**` 글로브·경로 양방향·트레일러 매치·빈 입력 일곱 경계).
- 이웃: `--test card_spec_cli` 11/0, `--test card_work_cli` 24/0, `cargo fmt --check` exit 0, `cargo clippy --offline --locked --lib -- -D warnings` exit 0(2026-11-19, worktree moai-adk-96).
- 남은 것: 리더가 `src/ops/card.rs` 배선을 넘기면(안건은 note 2368) AC-1..AC-3 재판정·PR·감사.
