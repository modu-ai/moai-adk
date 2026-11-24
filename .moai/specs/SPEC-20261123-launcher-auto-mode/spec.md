---
id: SPEC-20261123-launcher-auto-mode
title: "cc·glm 론처의 기본 권한 모드는 auto mode"
status: draft
revision: 1
created: 2026-11-23
---

# cc·glm 론처의 기본 권한 모드는 auto mode

## Intent

User instruction 2026-11-23: the `moai cc` and `moai glm` launchers start in auto mode by default. This replaces the 2026-11-10 decision that a role session (`-l`, `-a`) starts with `--dangerously-skip-permissions`, and it gives a session without a role the same default instead of whatever the user's own settings choose. Measured with the installed `moai-cli 4.2647.19 (rust)` and `--dry-run` on 2026-11-23:

- `moai cc`, `moai glm` (no role): no permission flag; the host follows the `defaultMode` of its settings.
- `moai cc -l|-a`, `moai glm -l|-a`, `moai -l|-a`: `--dangerously-skip-permissions`.
- `--permissions`: no permission flag.
- The user key `launcher.role_access` takes only `full` or `prompt`.
- `claude --help` (2.1.319) lists `--permission-mode` with the choice `auto`.

`full` (the bypass) stays available, only when the user chooses it in `launcher.role_access`. Behind a gateway such as z.ai the auto mode classifier may show a billing notice; the launcher notes point at `docs/troubleshooting.md` section 5 and change no setting.

Readers: users of `moai cc|glm`, the leader that starts lane sessions, and the auditor.

## Non-goals

- The legacy `moai session --role …` form keeps its recorded answer (no permission flag).
- Changing a host's own settings, the project's `permissions` key or the classifier's behaviour.
- Starting a real `claude` session: the checks read the composed argv through `--dry-run`.

## Risks

- A host older than the one measured may not know `--permission-mode auto`; the launcher passes the flag as documented by the host and never probes the host. Mitigation: the user sets `launcher.role_access` to `prompt` or `full`.
- Behind a gateway the auto mode classifier may bill or fail; REQ-8 only points at the documented notice.

## Requirements

- REQ-1: When `moai cc` or `moai glm` composes a launch (no role, `-l`, `-a` or the top-level `-l|-a`) with neither `--permissions` nor a `launcher.role_access` key, the launcher shall put `--permission-mode auto` in the host argv and leave `--dangerously-skip-permissions` out.
- REQ-2 (REQ-1): IF `launcher.role_access` holds a value other than `auto`, `full` or `prompt`, THEN the launcher shall refuse the launch with a message that names the three values and does not quote the value.
- REQ-3: Where `--permissions` is given, the launcher shall add no permission flag, whatever `launcher.role_access` holds.
- REQ-4: Where `launcher.role_access` is `auto` and `--permissions` is absent, the launcher shall add `--permission-mode auto` and no `--dangerously-skip-permissions`.
- REQ-5: Where `launcher.role_access` is `full` and `--permissions` is absent, the launcher shall add `--dangerously-skip-permissions` and no `--permission-mode`.
- REQ-6: Where `launcher.role_access` is `prompt` and `--permissions` is absent, the launcher shall add no permission flag.
- REQ-7: The launcher's notes and help text and the two documents that describe the launcher shall name auto mode as the default and full access only as the user's own choice.
- REQ-8: Where the host is `glm` and the launch is in auto mode, the launcher notes shall point at `docs/troubleshooting.md` section 5 for the classifier notice.

## Acceptance

- AC-1 (REQ-1, REQ-2): the `--dry-run` argv of `moai cc`, `moai cc -l`, `moai cc -a`, `moai glm`, `moai glm -l`, `moai glm -a`, `moai -l` and `moai -a` holds `--permission-mode auto` and no `--dangerously-skip-permissions`, with a temporary HOME that has no `launcher.yaml`; every bad `role_access` value is refused without being quoted. · verify: `cargo nextest run --locked --offline --test launcher_permissions_setting --test product_cli --test launcher_roleless`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6): with a temporary HOME, `role_access` `auto` or a missing key gives auto mode, `full` gives the bypass, `prompt` gives no flag, and `--permissions` gives no flag for each of them, for a roleless, `-l` and `-a` launch. · verify: `cargo nextest run --locked --offline --test launcher_permissions_setting --test launcher_agent_numbering --test launcher_advisor_setting`
- AC-3 (REQ-7, REQ-8): the help text of `--permissions`, the notes of an auto-mode launch (and the troubleshooting pointer on `glm`), `docs/rust-product-path.md` and `docs/user-entry-draft.md` say auto mode is the default and never say full access is the default. · verify: `cargo nextest run --locked --offline --test launcher_permissions_setting --test help_layout`

## Scope

- write: `src/session.rs`, `src/main.rs`, `src/help.rs`, `src/launcher_terminal.rs`
- write: `tests/launcher_permissions_setting.rs`, `tests/launcher_agent_numbering.rs`, `tests/launcher_advisor_setting.rs`, `tests/launcher_roleless.rs`, `tests/launcher_agent_entry.rs`, `tests/product_cli.rs`, `tests/fixtures/help/**`
- write: `docs/rust-product-path.md`, `docs/user-entry-draft.md`

## Plan

1. Change the tests first: every launcher test that expects `--dangerously-skip-permissions` by default expects `--permission-mode auto`, and new tests cover the `role_access` values and `--permissions`.
2. `src/session.rs`: resolve the access as one of auto, full or prompt (`launcher.role_access`, `--permissions`), add the auto flag and its notes.
3. `src/main.rs` and `src/help.rs`: the option text; refresh the help fixtures.
4. Fix the two documents, then run the targeted checks and the `--dry-run` receipt in a temporary HOME.

## Result

- AC-1: met — `cargo nextest run --offline --locked` over launcher_permissions_setting, launcher_roleless, launcher_agent_entry, launcher_advisor_setting, launcher_agent_numbering, product_cli and 12 more binaries: 1984 passed, 1 skipped (a test already marked ignored); the `--dry-run` receipt in a temporary HOME shows `--permission-mode auto` and no `--dangerously-skip-permissions` for `cc`, `cc -l`, `cc -a`, `glm`, `glm -l`, `glm -a`, `-l`, `-a` and `-a glm`.
- AC-2: met — the same run covers `role_access` `auto`, `full`, `prompt`, a missing key and `--permissions` for a roleless, `-l` and `-a` launch; the receipt shows `full` giving `--dangerously-skip-permissions`, `prompt` and `--permissions` giving no permission flag, and `bypass-please` refused with `must be auto, full or prompt` and exit 1.
- AC-3: met — `help_layout` goldens regenerated and passing; `notes_name_the_mode_and_only_glm_points_at_the_classifier_notice` passes; `docs/rust-product-path.md` and `docs/user-entry-draft.md` name auto mode as the default and describe full access only as the user's own choice (the 2026-11-10 decision is kept as a marked record).
