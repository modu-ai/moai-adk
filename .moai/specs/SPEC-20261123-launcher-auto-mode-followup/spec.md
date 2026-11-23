---
id: SPEC-20261123-launcher-auto-mode-followup
title: "론처 기본 auto mode 후속: 지원 표 문서, Access 기본값 주석, 도움말 확인"
status: draft
revision: 1
created: 2026-11-23
---

# 론처 기본 auto mode 후속: 지원 표 문서, Access 기본값 주석, 도움말 확인

## Intent

론처 기본 auto mode 착지(`SPEC-20261123-launcher-auto-mode`)에 대한 독립 감사의 비차단 후속이에요. 지원 표 문서가 옛 권한 설명을 그대로 적고, `Access` 의 `Default`(`Prompt`)가 론처 기본(`Auto`)과 왜 다른지 코드에 이유가 없고, 한국어 도움말이 실제로 보이는지 확인되지 않았어요. 코드 동작은 바꾸지 않아요. 론처의 argv, 도움말 문구와 고정 도움말 파일(`tests/fixtures/help/**`)은 그대로 두고, `src/help.rs` 의 한국어 문구 손질과 실제 `claude` 가 `--permission-mode auto` 를 받는지 보는 일은 이 SPEC의 범위 밖이에요.

## Requirements

- REQ-1: The launcher section of `docs/host-support-matrix.md` shall state that every launch starts with `--permission-mode auto` by default, that `--dangerously-skip-permissions` is added only when the user picks `full` in `launcher.role_access`, and that `--permissions` adds no permission flag.
- REQ-2: The line of `Access` that carries `#[default]` in `src/session.rs` shall have one English comment line that says why the type's own default (`Prompt`, no flag) differs from the launcher default (`Auto`).
- REQ-3: When the Korean launcher help is run under a terminal, the Result section of this SPEC shall record the command and its output.
- REQ-4 (REQ-3): IF the help stays English under a terminal, THEN the Result section of this SPEC shall name the code that keeps it English.

## Acceptance

- AC-1 (REQ-1): the launcher section names auto mode as the default, `full` as the user's own choice and `--permissions` as no flag, and no longer says `-l|-a` attaches full access or that a roleless session gets no permission flag · verify: `rg -c "permission-mode auto" docs/host-support-matrix.md` prints 1 or more, and `rg -n "전체 권한을 붙이는|권한 플래그를 붙이지 않습니다" docs/host-support-matrix.md` prints nothing
- AC-2 (REQ-2): one `//` or `///` English line sits directly beside `#[default]` and the build still passes · verify: `cargo clippy --offline --locked --all-targets -- -D warnings`
- AC-3 (REQ-3, REQ-4): the Result below holds the command that ran the Korean help under a terminal and what it printed, or names the code that keeps it English · verify: `moai spec check .moai/specs/SPEC-20261123-launcher-auto-mode-followup`

## Scope

- write: `docs/host-support-matrix.md`
- write: `src/session.rs`
- write: `.moai/specs/SPEC-20261123-launcher-auto-mode-followup/spec.md`

## Plan

1. `docs/host-support-matrix.md`: 론처 절의 옛 권한 설명을 지금 동작으로 고쳐요.
2. `src/session.rs`: `Access` 의 `Default` 옆에 이유를 적은 영어 주석 한 줄을 달아요(코드 바뀜 없음).
3. 한국어 도움말을 터미널(pty) 아래에서 실행해 보고, 결과나 안 되는 이유를 Result 에 적어요.
4. 대상 검사(`moai check affected`)를 돌리고 READY 에 영수증을 적어요.

## Result

- AC-1: met — `rg -c "permission-mode auto" docs/host-support-matrix.md` printed `1` (exit 0) and `rg -n "전체 권한을 붙이는|권한 플래그를 붙이지 않습니다" docs/host-support-matrix.md` printed nothing (exit 1). The dry-run argv of this tree's build (`cc`, `cc -a`, `glm -l`) holds `--permission-mode auto`, `launcher.role_access: full` gives `--dangerously-skip-permissions`, and `cc --permissions` carries neither flag, as the new text says.
- AC-2: met — `src/session.rs` gains one `//` English line beside `#[default]` on `Access::Prompt`; `cargo clippy --offline --locked --all-targets -- -D warnings` exited 0 and `rustfmt --check` printed nothing.
- AC-3: met — the Korean help cannot be rendered by the CLI: `ui::ui_lang` (`src/ui.rs`) returns `Lang::En` for every stream (user decision 6878265192, the CLI's human output is English only). `cargo nextest run --offline --locked --test help_english_terminal_cli` ran every command's `--help` (the `cc` and `glm` launchers included) on a terminal with `MOAI_LANG=ko`, `LANG=ko_KR.UTF-8` and `LC_ALL=ko_KR.UTF-8`: 3 passed, 0 failed, so no command printed Korean help. `script` could not give a terminal in this session (`tcgetattr/ioctl: Operation not supported on socket`), so the repository's own pty test is the evidence. The Korean strings of `src/help.rs` stay unreachable (follow-up: the launcher help rows still say `-l|-a` sets "the name and the permission").
