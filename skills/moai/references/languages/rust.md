# Language: Rust

When to read: `internal/quality` sent you here because the project's detected languages include `rust`. The project's own declarations (CI steps, `Makefile`, `justfile`, `AGENTS.md`) win over every default below.

## Evidence and defaults

| Kind | Default command | Use it only when |
| --- | --- | --- |
| format | `cargo fmt --check` | the project has `Cargo.toml` (rustfmt reads `rustfmt.toml` when present) |
| lint | `cargo clippy --all-targets` | clippy is in the project's CI or docs; add `-- -D warnings` only when they do |
| type check | `cargo check --all-targets` | always available; the test build covers it too |
| test | `cargo test` | `Cargo.toml`; add `--locked` when `Cargo.lock` is tracked |
| build | `cargo build` | `Cargo.toml` |

In a workspace, `-p <crate>` narrows a run to one member; say so when you narrow.

## Dependency existence

- `cargo info <crate>` (cargo 1.82 or later) or `cargo search <crate> --limit 1`; keep the line that shows the exact name and version.
- Add with `cargo add <crate>`; show the new `Cargo.lock` entry.

## Traps

- A test-name filter applies to every test target of the call: `cargo test foo` runs `foo` in each target, and a target where nothing matches prints `running 0 tests` and still exits 0. Name one target (`--test <name>` or `--lib`) when a filter must select something.
- `#[ignore]` tests do not run by default; `cargo test -- --ignored` runs them. A newly ignored test is a finding.
- `--locked` fails when `Cargo.lock` is out of date: that is a real finding, not a reason to drop the flag.
- Features change what compiles: run the feature set the CI runs (`--all-features`, `--no-default-features` or a list).
- `cargo fmt` without `--check` rewrites files; a VERIFY run never rewrites.
- Build output (`target/`) is per worktree; two worktrees sharing one `CARGO_TARGET_DIR` can test each other's binaries.
