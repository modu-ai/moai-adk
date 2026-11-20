# moai-adk development tools (repo-only)

This repository's development tools live under `tools/` and are **not** part of the
installed moai package: the shipped skills and agents never point at them. This page keeps
the procedures the shipped text used to carry, so the skills stay generic for every user
(user instruction 2026-11-20: "로컬 개발 도구와 오픈소스 기능을 혼동하지 마라"). Every procedure
below runs in a checkout of **this** repository, on this repository's own trees.

## The self-check runner (`tools/self_check.sh`)

Runs the machine items of a change in one go (`bin-moai`, `bin-moai-cli`, `clippy`, `fmt`,
`fmt-ops`, `git-env-scrub`, `home-isolation`, `ac-command-lint`, `package-inspect`,
`package-size`, `manifest`, `diff-check`, `python`) and prints one
`self-check: <item> exit=<code>` line each; copy those lines into the `## Self-Audit`
section.

- `--full-python` runs CI's whole Python suite instead of the touched files, in the python
  slot: its own slot folder (`MOAI_PYTHON_SLOT_DIR`, default
  `${XDG_CACHE_HOME:-$HOME/.cache}/moai-cli/python-slots`) with at most
  `MOAI_PYTHON_SLOT_MAX` runs at a time (default 2), so a whole Python run never takes a
  cargo slot. A heavy or Python item the machine's slot refused prints
  `exit=75 slot-refused (not run)`: report it as not run, never as passed.
- `--pre-ready` is the whole-tree list a head passes before the lane reports it ready:
  `diff-check`, `fmt`, `fmt-ops`, `manifest`, `spec-check`, `py-source-manifest`,
  `py-repo-scans`, `skill-package`, fast ones first; each item prints one `self-check:`
  line (`exit=<code>`, `skip (<reason>)` or `not_run (<reason>)`), and `--pre-ready` and
  `--full-python` cannot be given together. `py-source-manifest` and `py-repo-scans` run
  in the python slot, `skill-package` in the heavy slot, the other items without a slot.
- `spec-check` and `py-repo-scans` need a built `moai-cli` (`MOAI_CLI_BIN`, or
  `${CARGO_TARGET_DIR:-target}/debug/moai-cli`): when it is missing or older than `src/`,
  `build.rs`, `Cargo.toml` or `Cargo.lock`, the item prints `not_run (no built moai-cli:
  <path>)` or `not_run (stale moai-cli: <path> is older than <file>)`; build it first with
  `cargo build --offline --locked --bin moai-cli` (`internal/build-folder` in the skill).
- An item that printed `not_run`, whatever the reason, is reported as not run, never as
  passed. Every build uses the worktree's build folder and `CARGO_INCREMENTAL=0`
  (`internal/build-folder`).

## The package manifest (`tools/source_manifest.py`)

After a change under `skills/moai/`: `tools/source_manifest.py write` (with `python3 -I`)
recomputes every entry's sha256, then `tools/source_manifest.py check` verifies them; a
new file needs its `origin` by hand. On a merge conflict in `source-manifest.json`, take
neither side's sha256 — `write` joins both sides three-way against the common version,
recomputes every sha256 and stops (exit 2, nothing written) when a person must choose. An
entry's `origin.reason` says where the file came from; keep per-change notes in the commit
message, because two branches appending to the same `reason` is the one conflict the tool
leaves to a person. Related Rust checks: `cargo test --offline --locked --test product_cli
skill_package::` (the package's manifest and route checks).

## The smoke harness (`tools/smoke/run.sh`)

Before the first commit of a change that touches install, update, `/moai project` or the
launcher: `tools/smoke/run.sh <area>`, with `<area>` one of `install`, `update`,
`project`, `launcher` or `all`, on the debug binary just built (`--bin PATH` names another
one). It boots the built binary in an empty temporary HOME and a temporary git project and
never reads the real HOME. Take its last line, `smoke: <area> pass|fail steps=<n>`, as the
evidence of the check; a `fail` is a failed check, a missing binary is a usage error
(exit 2) and not a pass, and a smoke run that was not made is `not_run`, never `passed`.

## The CI scope classifier (`tools/ci_scope.py`)

The docs-only CI decision of this repository's own workflow runs before the Rust toolchain
is installed, which is why it stays a script rather than a moai command: it reads changed
paths from stdin and answers whether every changed path is under a verified-safe doc root
(`docs/` with `.md`/`.txt`/`.png`/`.svg` suffixes, minus files tests read, plus
`.moai-cli/reports/`); any failure or unrecognized path fails safe to the full scope. When
a change alters the classifier, run the old and the new version over recent real inputs
(for example the last 25 pull request bodies) and report every verdict that changed.

## The local check set (`tools/local_check.sh`)

The check set this repository's lanes run before asking for an audit — fetch, test
(cargo-nextest), doctest, clippy, fmt, fmt-ops, unittest, diff-check — with
`LOCAL_CHECK_LEG` splitting it into the CI's nextest and rest legs. See
`tools/local_check.sh` (the file's header documents every mode and variable) and the
verify workflow's heavy-checks rules.

## Others

- `tools/build_folder.py` — the build-folder helpers this repository predates the moai
  commands with: `moai lane-env` covers `env`, `moai optimize-plan` (OPTIMIZE route) lists
  the folders, and the removal stays this repository's own tool (`remove`) — no moai
  command deletes a build folder; `internal/build-folder` in the skill keeps the generic
  rule (a fresh preview still marking the row `trim`, relative paths only).
- `tools/legacy_verdicts.py` — regenerates the verdict of every row of
  `.moai-cli/manifests/legacy-asset-map.json` (user decision 2026-11-17 D8).
- `tools/heavy_check_slot.py` — the Python slot tool that `moai heavy-slot-acquire`
  reimplemented (#2491); the Rust side embeds it for folder-and-lock compatibility.
