---
id: SPEC-20261126-card-note-tty-stdin
title: "card note 는 표준입력이 터미널이면 기다리지 않고 사용법 오류로 거절해요"
status: draft
revision: 1
created: 2026-11-26
---

# card note 는 표준입력이 터미널이면 기다리지 않고 사용법 오류로 거절해요

## Intent

mo.ai.kr 리더 덧붙임 8 Q 의 실측: 인자를 비운 `moai card note --kind comment <프로젝트 경로>` 가 표준입력을 기다리며 120초 타임아웃까지 멈췄어요. 노트 본문은 파일이나 heredoc 으로만 받으니, 표준입력이 터미널이면 사람이 칠 글이 없어서 기다리는 것이 오류예요. 이 SPEC 은 그 경우를 바로 사용법 오류로 거절하게 해요.

## Requirements

- REQ-1: When `moai card note` starts with standard input on a terminal, the command shall exit non-zero with a usage error before it reads standard input or writes the ledger.
- REQ-2 (REQ-1): IF standard input is a pipe or a file, THEN the command shall read the note text as before.

## Acceptance

- AC-1 (REQ-1): 표준입력이 pty(터미널)일 때 `card note` 가 기다리지 않고 0 이 아닌 종료 코드와 `card note refused: standard input is a terminal` 로 시작하는 사용법 문구로 끝나요 · verify: `cargo test --test card_note_tty_cli -- a_terminal_on_stdin_is_refused_with_a_usage_error_and_no_wait`
- AC-2 (REQ-2): 표준입력이 파이프이면 터미널 거절이 나오지 않고 이전처럼 본문을 읽어요 · verify: `cargo test --test card_note_tty_cli -- a_pipe_on_stdin_is_not_refused_as_a_terminal`

## Scope

- write: `src/ops/card.rs`
- write: `tests/card_note_tty_cli.rs`

## Plan

1. 실패하는 pty 테스트 `tests/card_note_tty_cli.rs` 를 먼저 써요.
2. `src/ops/card.rs` 의 `Action::Note` 팔에서 표준입력을 읽기 전에 터미널 여부를 보고 거절해요(공유 파일이라 `card line-lease` 로 써요).
3. `card note` 도움말에 터미널 거절을 한 줄 적어요.

## Result

`src/ops/card.rs` 의 `Action::Note` 팔이 표준입력을 읽거나 홈을 찾기 전에 `std::io::IsTerminal::is_terminal(&std::io::stdin())` 를 보고, 터미널이면 `card note refused: standard input is a terminal` 로 시작하는 오류로 거절해요. `card note` 도움말 문서 주석에 이 거절을 설명하는 문장을 한 줄 더했어요. `tests/card_note_tty_cli.rs` 에 pty 테스트와 파이프 테스트 두 개를 새로 썼어요.

- AC-1: met · 변경 전 `cargo test --locked --test card_note_tty_cli` 에서 `a_terminal_on_stdin_is_refused_with_a_usage_error_and_no_wait` 가 20초 제한을 넘겨 계속 돌다 강제 종료됐어요(보고된 멈춤). 변경 후에는 이 테스트가 통과했어요(`2 passed; 0 failed`, 0.07s).
- AC-2: met · 변경 전에도 `a_pipe_on_stdin_is_not_refused_as_a_terminal` 이 통과했고, 변경 후에도 통과했어요(`2 passed; 0 failed`, 0.07s).

Remaining: 넓은 `check affected` 영수증과 독립 감사는 아직 오지 않았어요(not_run).
