---
id: SPEC-20261122-claude-md-text-and-ignore
title: "The ignore files let git track .claude settings and rules, and a text test keeps AGENTS.md out of moai's own texts"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-claude-md-layout
---

# Ignore files and the text test

## Intent

두 가지를 묶었어요. (1) `.gitignore`: `.claude/` 통째 무시를 풀어 `.claude/settings.json` 과 `.claude/rules/` 를 추적할 수 있게 해요. 지금 사용자 프로젝트용 관리 블록 `moai-local` 은 `.claude` 줄을 쓰지 않으므로 "쓰지 않는다"를 고정하고, 사용자 자신의 규칙이 그 파일을 가리면 알려만 줘요. 이 저장소 자신의 `.gitignore` 는 356행과 386행의 두 줄을 허용 목록으로 바꿔요(체크아웃의 `.claude/` 에는 추적되지 않는 Go 세대 설치본이 있고, `settings.json` 에는 기계별 `PATH` 가 있을 수 있어서 읽어 보고 정해요). (2) 글 시험: 출하 글·템플릿·도움말·생성 파일이 `AGENTS.md` 를 moai 의 파일로 적으면 저장소 시험이 실패해요. 정당한 예외는 경로와 이유로 시험 안에 하나씩 적어요: Go 세대 백업 글, 사용자 파일을 옮기는 `update.md` 문장, Codex 홈 파일, 그리고 "사용자의 파일을 가리킬 뿐 moai 의 파일이 아니다"라는 이유의 셋 — `import` 가 쓰는 `@AGENTS.md` 줄, `leave` 경고 문장, 옛 영수증을 되돌리는 `claude-import` 글. 시험은 "moai 가 쓴다"는 문맥과 "사용자 파일을 가리킨다"는 문맥을 구분해요. 배너 `MoAI 🗿 <name>` 은 이미 있는 시험을 이 SPEC 의 수용 기준에서 가리켜 고정만 해요.

## Requirements

- REQ-1: The git ignore block shall contain a rule that ignores `CLAUDE.local.md`.
- REQ-2: The git ignore block shall not ignore the whole `.claude/` folder.
- REQ-3: When a rule of the project's own ignore files hides `.claude/settings.json` or `.claude/rules/`, the update preview shall name that rule and its file and leave the file unchanged. · failure: none — a warning only and nothing is written
- REQ-4: The repository ignore file shall let git track `.claude/settings.json` and `.claude/rules/` while it keeps ignoring `.claude/worktrees/` and `CLAUDE.local.md`.
- REQ-5: The repository text test shall fail when a shipped skill text, a template, a help text or a generated project file names `AGENTS.md` or `AGENTS.local.md` as a file moai writes.
- REQ-6: The repository text test shall list each place that may still name `AGENTS.md` by its path and its reason.
- REQ-7: The shipped output style text shall keep the banner form `<emoji> MoAI 🗿 <name>` in every banner line.

## Acceptance

- AC-1 (REQ-1…REQ-3): the managed block names `CLAUDE.local.md`, never ignores `.claude/`, and the preview names a user rule that hides settings or rules.
  verify: `cargo test --locked --offline --test gitignore_block_cli`
  verify: `cargo test --locked --offline --test claude_md_ignore_cli ignore_`
- AC-2 (REQ-4): `git check-ignore` reports `.claude/settings.json` and `.claude/rules/moai/flow.md` as not ignored and `.claude/worktrees/x` and `CLAUDE.local.md` as ignored.
  verify: `cargo test --locked --offline --test claude_md_ignore_cli repository_`
- AC-3 (REQ-5, REQ-6): the repository text test finds no shipped text or generated file that names `AGENTS.md` as moai's file and lists its exceptions. This is the second criterion of the card.
  verify: `cargo test --locked --offline --test claude_md_layout_scan`
- AC-4 (REQ-7): the banner pins stay green.
  verify: `cargo test --locked --offline --test style_banner_scan`
- AC-5: neighbouring checks on the changed files.
  verify: `cargo fmt --check`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.gitignore`, `src/agents.rs` (the ignore block), `skills/moai/**` (texts and `references/source-manifest.json`), `src/language.rs`, `src/generation.rs`, `src/launcher_init.rs`, `tests/claude_md_ignore_cli.rs`, `tests/claude_md_layout_scan.rs`, `tests/gitignore_block_cli.rs`, `.moai/specs/SPEC-20261122-claude-md-text-and-ignore/**`
- read-only: `tests/style_banner_scan.rs`, `.claude/settings.json` of the checkout (read once to decide the allow list, never staged)
- forbidden: `AGENTS.md`, `AGENTS.local.md` of this repository, the real HOME, the network

## Plan

1. Read `.claude/settings.json` of the checkout and decide the allow list (decision D-9 of the layout SPEC).
2. Change the ignore block writer: a new `CLAUDE.local.md` rule, a new check that finds a user rule hiding the two paths (extend `user_rule_hits` of `src/agents.rs:784-866`).
3. Change the repository `.gitignore` lines 348-356 and 386 and their policy comment; update `the_repository_gitignore_tracks_specs_and_ignores_the_rest`.
4. Rewrite the shipped texts that name `AGENTS.md` (the list is in the layout SPEC, "What exists"); `src/language.rs:93-94` and `src/generation.rs:1407` follow; refresh the source manifest hashes.
5. Write the text test with its exception list; run the two scan modules (`tests.repo_scans_test`, `tests.source_manifest_test`) first, as the repository rule says.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The text test is a Rust test that reads the shipped paths, in the style of `style_banner_scan`. | No new Python tests (user rule 2026-11-21). | None. |
| D-2 | The exception list is part of the test file, one row per path with its reason. | A hidden skip list would let a new naming slip through. | Move it to a data file. |

## Result

(Filled when the work is done.)
