---
id: SPEC-20261124-launcher-retire-role-flags
title: "The launcher retires -l/-a: --name is the only way to name a session, the role is entered inside it"
status: draft
revision: 1
created: 2026-11-24
---

# The launcher retires -l/-a: --name is the only way to name a session, the role is entered inside it

## Intent

카드 C-449 예요. 사용자 결정 D-1(`docs/decisions-2026-11-08.md` 166줄, 2026-11-09)은 런처 `-l`/`-a` 를 "폐기(deprecate) — 경고 후 제거, `--name` 만 유지, 역할은 세션 안 `/moai --mode-leader|--mode-agent` 로만" 으로 정했어요. 2026-11-24 에 사용자가 다시 말했어요: "-l, -a 은 더이상 사용하지 않기로 했잖아." 그런데 지금도 `moai cc --help` 가 `-l/--leader`·`-a/--agent` 를 보여 주고, `moai glm -l --dry-run` 이 돌고, `CLAUDE.md` 와 문서와 리더의 배차 습관이 그것을 말해요(`SPEC-20261118-launcher-agent-entry` 는 `-l|-a` 를 "이름과 권한만 정하는" 플래그로 도움말에 남겼어요).

읽어서 확인한 사실(코드, 2026-11-24, 브랜치 기준 NeXT `efb3a80ae9`):

- `-l`/`-a` 가 하는 일은 셋뿐이에요: 세션 이름(`-l` → `leader`, `-a` → 프로젝트의 다음 `agent-N`, `src/main.rs` 의 `launch`), 저장된 모델 고르기(`session::apply_saved_choice` 가 `launcher.models.<host>.<leader|agent>` 를 읽어요, 역할 없는 세션은 읽지 않아요), 안내 한 줄(`session::role_entry_note`). 권한은 이미 역할과 상관없어요: `launcher.role_access`(`auto|full|prompt`)는 역할이 있든 없든 모든 실행에 같아요(`session::launch_access`, SPEC-20261123-launcher-auto-mode).
- `agent-N` 번호는 런처가 `take_agent_number` 로 잡고 `.live` 잠금으로 살아 있음을 알려요. `lane_registry::live_agent_numbers` 가 같은 폴더(`session::agent_seq_dir`)를 읽고, `session-register` 가 이름 충돌을 막는 쪽이에요. `--name` 만 준 세션도 `<name>.live` 잠금을 잡아요(SPEC-20261120-launcher-liveness REQ-4).
- 상태 줄(`src/statusline/**`)은 `-l|-a` 를 말하지 않아요. 역할은 `session-register --mode` 가 세션 안에서 적고 상태 줄이 읽어요(`lane_registry::mode_of`).
- RC1(`2026.11.0-rc.1`)은 게시 전이에요(`docs/release-notes/2026.11.0-rc.1.md`: 초안, `release blocked`). 경고를 받을 바깥 사용자가 없고, 쓰는 사람은 운영자 한 명이에요.

## Non-goals

- 역할을 세션 안에서 들어가는 방법(`/moai --mode-leader`, `/moai --mode-agent`)과 `session-register --mode` 는 바꾸지 않아요.
- 옛 `moai session --role` 명령은 그대로예요(REQ-00F-06, 기록된 답 `tests/fixtures/protocol/v1/current/18-session-dry-run.json`).
- `install/moai-switch.sh` 와 `src/shim_check.rs` 의 `-l|-a|--leader|--agent` 줄은 그대로 둬요: 그 낱말도 Rust 로 가야 거절 문장이 보여요. 사용자 HOME 의 shim 은 이 카드가 건드리지 않아요.
- 과거 기록(결정 기록 `docs/decisions-2026-11-08.md`, 닫힌 SPEC, `docs/design-review/**`, `docs/session-hub-research-2026-11-08.md`, `docs/release-notes/**`)은 고치지 않아요. 결정 기록에는 승인 뒤 리더가 한 줄을 더해요.
- `launcher.role_access` 의 이름과 값은 바꾸지 않아요(D-3).
- 실제 HOME, 실제 호스트 시작, 네트워크는 쓰지 않아요.

## Requirements

- REQ-1: IF a launch is given `-l`, `--leader`, `-a` or `--agent`, THEN the launcher shall refuse it with one message that names `--name` and the two mode commands and start no host, take no number or lock and write no file.
- REQ-2: The launcher help and usage shall name no role flag as a launcher option.
- REQ-3: The launcher shall name a session as `--name` gives, hold its `.live` lock as before and add no name and no `agent-N` number when `--name` is absent.
- REQ-4: The launcher shall fill the model and effort it was not given from the saved entry of its host under `launcher.models`, and from the defaults of a session without a role when no entry exists.
- REQ-5: IF the saved settings hold a role entry under `launcher.models`, THEN the launcher shall start without reading that entry and add one note that names the key paths and `moai update`.
- REQ-6: The `moai update` model pick shall ask one model and effort per host, save it under `launcher.models` and list a saved role entry among the keys that its write replaces.
- REQ-7: The launcher notes and help shall say that `launcher.role_access` sets the permission mode of every launch.
- REQ-8: The start-up instructions of the repository shall name `moai cc` or `moai glm` with an optional `--name` and the two mode commands and tell nobody to use `-l` or `-a`.
- REQ-9: The `CLAUDE.md` section "Scope and sessions" shall say that `-l|-a` are retired.
- REQ-10: The rule provenance file shall list the `CLAUDE.md` sentence that this change replaces.

## Acceptance

- AC-1 (REQ-1): `cc -l`, `cc -a`, `glm --leader`, `glm --agent`, `moai -l` and `moai -a cc` each exit non-zero and print one message that names `--name`, `/moai --mode-leader` and `/moai --mode-agent`; `--dry-run` refuses the same way. · verify: `cargo nextest run --offline --locked --test launcher_retire_role_flags -E 'test(every_role_flag_is_refused_with_the_replacement)'`
- AC-2 (REQ-1): in a temporary project and a temporary HOME a refused launch leaves no `agent-N` number, no `.live` file and no project file, and no fake host binary on PATH was run. · verify: `cargo nextest run --offline --locked --test launcher_retire_role_flags -E 'test(a_refused_role_flag_starts_and_writes_nothing)'`
- AC-3 (REQ-1): the refusal is a guard: on a `cp -c` copy that accepts `-l` again, the two tests above fail. · verify: guard-mutant run of AC-1 and AC-2 on that copy, exit 100
- AC-4 (REQ-2): the long help of `cc` and `glm`, `help --all`, the root screen at 60 and 80 columns and without `COLUMNS`, the usage lines and the usage error of an unknown command contain no `-l`, `--leader`, `-a` or `--agent` as a launcher option; the root `cc` example and the launcher line say `moai cc|glm`. · verify: `cargo nextest run --offline --locked --test launcher_retire_role_flags -E 'test(no_help_screen_offers_a_role_flag)'`; `cargo nextest run --offline --locked --test help_layout`
- AC-5 (REQ-2): the eight help goldens and the usage fixture change on the lines that name the flags only, and every width and length rule of `help_layout` and `--lib help::` passes unchanged. · verify: `git diff --stat <base>..<head> -- tests/fixtures`; `cargo nextest run --offline --locked --test help_layout`; `cargo nextest run --offline --locked --lib help::`
- AC-6 (REQ-3): `cc --name x --dry-run` names the session `x`, `cc --dry-run` adds no name, two launches without `--name` print no `agent-N`, and `--name x` holds `x.live` while the host runs. · verify: `cargo nextest run --offline --locked --test launcher_retire_role_flags -E 'test(a_session_is_named_by_name_only)'`; `cargo nextest run --offline --locked --test launcher_liveness`
- AC-7 (REQ-4): with a temporary HOME that holds a supported `launcher.models.claude` entry, `cc --dry-run` takes its model and effort and says so; with no entry it keeps `opus` and `medium`; an unsupported flat entry still refuses with key paths only. · verify: `cargo nextest run --offline --locked --test launcher_reads_models --test launcher_retire_role_flags -E 'test(the_flat_saved_entry_fills_model_and_effort) | test(no_saved_entry_keeps_the_roleless_defaults)'`
- AC-8 (REQ-5): with only `launcher.models.claude.leader` and `.agent` saved, `cc --dry-run` takes the defaults and adds one note that names both key paths and `moai update`. · verify: `cargo nextest run --offline --locked --test launcher_retire_role_flags -E 'test(old_role_entries_are_ignored_with_one_note)'`
- AC-9 (REQ-6): `moai update` in a temporary HOME shows one model question per host, writes `launcher.models.<host>` only after the apply, and lists a saved role entry among the keys the write replaces. · verify: `cargo nextest run --offline --locked --test model_effort_settings --test model_effort_questions --test update_model_pick`
- AC-10 (REQ-7): the notes and long help that mention `launcher.role_access` say "every launch", and no launcher note or help says that a role chooses access. · verify: `cargo nextest run --offline --locked --test launcher_permissions_setting --test launcher_retire_role_flags -E 'test(the_access_notes_name_no_role)'`
- AC-11 (REQ-8): a scan over the named live files (the list is in the test, not the whole repository) finds no text that tells a person to use `-l|-a`, `moai -l`, `moai -a`, `cc -l`, `cc -a`, `glm -l` or `glm -a`; those files name `moai cc|glm` with `--name` and the two mode commands; the history files named in Non-goals are not in the list. · verify: `cargo nextest run --offline --locked --test launcher_retire_role_flags -E 'test(no_live_text_tells_anyone_to_use_the_role_flags)'`; `cargo nextest run --offline --locked --test output_names_moai --test codex_launcher_removed --test omp_launcher_removed`
- AC-12 (REQ-9, REQ-10): `CLAUDE.md` says in "Scope and sessions" that `-l|-a` are retired and names `moai cc|glm` with `--name`; `docs/claude-md-rule-provenance.md` lists the replaced sentence; `tests/repo_claude_md.rs` pins the new sentence in place of the old one; `CLAUDE.md` stays under 200 lines. · verify: `cargo nextest run --offline --locked --test repo_claude_md`; `wc -l CLAUDE.md`
- AC-13 (REQ-1..REQ-10): no write outside the scope below; `src/session.rs` keeps `agent_seq_dir`, `lock_named_live` and `launch_access`; the legacy `session` command answers as before. · verify: `git diff --stat <base>..<head>`; `cargo nextest run --offline --locked --test launcher_roleless --test launcher_agent_numbering --test product_cli`; `moai check affected <base>..<head>`

## Scope

- write (this card's reservation today): `.moai/specs/SPEC-20261124-launcher-retire-role-flags/**`, `CLAUDE.md`. Everything below is for the leader to add to the reservation when the SPEC is approved; no code is written before.
- write after approval: `src/main.rs`(shared file: `card line-lease` for the lines), `src/session.rs`, `src/help.rs`(shared), `src/launcher_terminal.rs`, `src/model_effort_settings.rs`, `src/model_effort_questions.rs`, `src/update_models.rs`, `src/ui_text.rs`(comment), `src/contract.rs`(comment), `src/lane_registry.rs`(comment and the unused numbering reader), `tests/launcher_retire_role_flags.rs`(new), `tests/launcher_agent_entry.rs`, `tests/launcher_agent_numbering.rs`, `tests/launcher_roleless.rs`, `tests/launcher_permissions_setting.rs`, `tests/launcher_advisor_setting.rs`, `tests/launcher_reads_models.rs`, `tests/launcher_terminal_output.rs`, `tests/launcher_glm_key_file.rs`, `tests/launcher_key_vault.rs`, `tests/launcher_auto_init.rs`, `tests/model_effort_settings.rs`, `tests/model_effort_questions.rs`, `tests/update_model_pick.rs`, `tests/product_cli.rs`, `tests/output_names_moai.rs`, `tests/codex_launcher_removed.rs`, `tests/omp_launcher_removed.rs`, `tests/repo_claude_md.rs`, `tests/help_layout.rs`, `tests/fixtures/help/all.txt`, `tests/fixtures/help/all-dumb.txt`, `tests/fixtures/help/all-80.txt`, `tests/fixtures/help/all-60.txt`, `tests/fixtures/help/root.txt`, `tests/fixtures/help/root-dumb.txt`, `tests/fixtures/help/root-80.txt`, `tests/fixtures/help/root-60.txt`, `tests/fixtures/protocol/v1/current/01-usage-unknown-subcommand.json`, `docs/user-entry-draft.md`, `docs/host-support-matrix.md`, `docs/rust-product-path.md`, `docs/claude-md-rule-provenance.md`, `install/README.md`
- read-only: `src/launcher_init.rs`, `src/statusline/**`, `install/moai-switch.sh`, `src/shim_check.rs`, `tests/fixtures/protocol/v1/current/18-session-dry-run.json`
- forbidden: the real HOME, any real host start, the history files named in Non-goals, loosening a help width or length rule, `skills/**` (no skill text names the flags; read 2026-11-24)

## Plan

1. Slice 1, the launcher (RED first): the new `tests/launcher_retire_role_flags.rs` (AC-1, 2, 4, 6, 10, 11) fails on the current code. Then drop `-l`/`-a` from `Cli`, `HostLaunch` and the top-level `LaunchOpts` flatten, add the one refusal for the four words, drop the numbering and the role entry note, fix every help, usage and Korean text, and rewrite the eight goldens and the usage fixture with `MOAI_HELP_GOLDEN_WRITE` on a temporary HOME.
2. Slice 2, the saved choice (AC-7, 8, 9): flat `launcher.models.<host>` read by `apply_saved_choice`, the ignore-with-one-note path for old role entries, and the `moai update` pick with one question per host.
3. Slice 3, the texts (AC-11, 12): `CLAUDE.md`, the three docs, `install/README.md`, the provenance entry and `tests/repo_claude_md.rs`.
4. READY: `moai check affected` once on the head, one `guard-mutant:` line per new guard test, and the dangerous-path run (Risks).

## Risks

- The operator's Orca presets, lane start commands and shell aliases outside this repository may still call `moai cc -a`. They break at once with the refusal text. The leader lists them before the landing (D-1).
- A real launcher settings file under the user's home may hold `launcher.models.<host>.leader|agent`. Today an entry the support table does not know refuses the launch; REQ-5 makes the old role keys a note instead, so nothing breaks, and `moai update` replaces them after a preview. Tests use temporary HOMEs only.
- Without the number, two lanes started by hand can pick one name. `session-register` already refuses a name another instance holds (`lane_registry`), and `<name>.live` stays; the leader names lanes explicitly (D-4).
- Help goldens carry width rules: a shorter launcher line must not break `help_layout`; the goldens change on the flag lines only (AC-5).
- `CLAUDE.md` is tier 2 (full audit), and Claude Code loads the `CLAUDE.md` of a session's own working tree, so a lane sees the new text only after it is committed and the lane starts from it.
- The dangerous path for this change is a launch that must start nothing: the READY note carries a run of the refusals in a throwaway project and a throwaway HOME, showing that no `agent-N` file, `.live` file or project file appeared (AC-2).

## Decisions

Every row is a proposal until a leader note records the approval; this SPEC records no approval itself.

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | **Proposed, default A.** A: refuse `-l`, `--leader`, `-a`, `--agent` at once with one message (exit non-zero, nothing started or written). B: keep them for this build, hidden from every help, with one warning line on stderr and in the `--dry-run` notes, and refuse in a follow-up card. | D-1 of 2026-11-09 says "경고 후 제거"; RC1 is unpublished and has one user, who restated the decision on 2026-11-24, so a warning phase protects nobody and keeps the numbering and role-keyed model code alive for another build. B costs a second card and keeps `-a` working with all its tests. | Re-add the flags and the numbering from this SPEC's commit |
| D-2 | **Proposed, default A.** A: one saved entry per host, `launcher.models.<host>` = `{model, effort}`; old `leader`/`agent` entries are ignored with a note and replaced by `moai update` after a preview. B: keep the `leader`/`agent` entries in the file and in `moai update` but read by nothing. | Without `-l\|-a` the launcher cannot know the role (the role is typed inside the session, and guessing it from a name is refused by `CLAUDE.md`), so a role-keyed entry is never reached. B leaves a question in `moai update` that changes nothing. The cost of A: one model for the leader and agent sessions of a host; `--model` and `--effort` still choose per launch. | Restore the role keys in `model_effort_settings` and the update pick |
| D-3 | **Proposed, default A.** A: keep `launcher.role_access` as the name and the values `auto\|full\|prompt`; say "every launch" in the texts. B: rename to `launcher.access` and read the old key as an alias for one release. | The key already applies to every launch whatever the role (`session::launch_access`); a rename edits the user's guarded home settings and changes no behavior. | Rename later in its own card |
| D-4 | **Proposed, default A.** A: the launcher stops numbering `agent-N`; a session is named by `--name` only, and the leader or Orca preset passes `--name agent-N`. B: keep the numbering behind a new spelling. | D-1 of 2026-11-09 keeps `--name` only; B adds a new launcher surface nobody asked for. `session-register` stays the collision guard. | Re-add `take_agent_number` from this SPEC's base |

## Result

(Filled in after the work. Nothing is implemented yet; this is the draft for approval.)
