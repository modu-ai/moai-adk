---
title: Getting started
description: Build the Rust moai-cli from source and inspect a project without installing into your home or starting an AI host.
---

MoAI's Rust command is **`moai-cli`**, not the existing Go command `moai`. This guide builds the release candidate from source and runs read-only previews. Building the binary does **not** install the `moai` skill into Claude Code or Codex.

## Prerequisites and support

- A Darwin/arm64 Mac, Git, and an existing Rust installation managed by rustup. The repository pins Rust **1.94.1** in `rust-toolchain.toml`; rustup may download that toolchain and its components on first use.
- Network access for cloning and, unless already cached, Rust components and Cargo dependencies. `--locked` preserves dependency resolution; it does not make a build offline.
- A terminal and a new checkout directory. Do not build over an existing Go installation or copy its credentials.

The package version is `2026.11.0-rc.1`, with `publish = false`. A public Rust installer or crates.io installation is not established by this documentation, so no `cargo install moai-cli` or installer script is prescribed.

**Support target is not certification.** Darwin/arm64 with Claude Code or Codex is the RC target. The source support matrix records five deterministic fixture results and one Claude static-validation result, not certified live-host loading. Actual `/moai` and `$moai` loading/invocation remains `not_run`. Darwin/x86_64, Linux, Windows, ZCode, and cloud/worktree host E2E are not certified by those results.

## Build an isolated source checkout

These commands create a new checkout and its build directory. They do not install a binary into your home, change your PATH, or start a host. The detached revision is the source snapshot used by these guides.

```sh
git clone --branch v2026.11.0 https://github.com/modu-ai/moai-adk.git moai-adk-rust
cd moai-adk-rust
git checkout --detach cf33967c42c31f020fd0e35dc07b6b82093119f1
CARGO_TARGET_DIR="$PWD/target-docs" cargo build --locked --release
export MOAI_SOURCE="$PWD"
export MOAI_CLI="$MOAI_SOURCE/target-docs/release/moai-cli"
"$MOAI_CLI" --version
"$MOAI_CLI" --help
```

Cargo/rustup can write their normal caches outside the checkout; this is a source build, not a claim of a fully sandboxed build. If the pinned toolchain or dependencies cannot be obtained, resolve that prerequisite rather than substituting an unverified installer. An offline build additionally requires an already populated cache and `cargo build --offline --locked --release` with the same target directory.

Keep `MOAI_CLI` as an absolute path. Do not rename it to `moai`, overwrite the Go binary, or run Go `moai init`/`moai update` as Rust setup steps. The variables above are for this shell; reopen the shell only after setting them again.

## Inspect without applying

First inspect this checkout and the common skill package:

```sh
"$MOAI_CLI" identity "$MOAI_SOURCE"
"$MOAI_CLI" package-inspect "$MOAI_SOURCE/skills/moai"
```

`identity` reports Git workspace identity and state roots without creating them. `package-inspect` reads the complete skill tree and computes its content digest. A digest proves observed content identity, **not** publisher authenticity, host discovery, or permission to execute the package.

To inspect a first profile plan without changing a real project, create a disposable directory and submit a complete v2 request:

```sh
export MOAI_PROJECT="$(mktemp -d)"
"$MOAI_CLI" project-plan "$MOAI_PROJECT" <<'JSON'
{
  "schema_version": 2,
  "project_id": "docs-first-project",
  "purpose": "Evaluate a local Rust workflow",
  "tools": ["Rust"],
  "development_methods": ["TDD"],
  "hosts": ["claude", "codex"],
  "interview_revision": "docs-first-plan",
  "decisions": []
}
JSON
```

`mktemp` creates the directory; **`project-plan` writes nothing**. It does not create `.moai-cli/project.yaml`, install skills, authenticate, invoke a model, or post to GitHub. Read the JSON envelope's `ok` and `result.action`, `result.delta`, and `result.plan_digest`; do not treat a printed plan as an applied configuration. Keep private paths, tokens, and credentials out of profile values.

## Understand the apply boundary

Do not apply the disposable example to your real project. For a real setup, review the exact target, proposed files, conflicts, and user approval first. Internal apply commands bind a plan digest and caller-supplied permission reference; **the reference string is metadata, not an authorization check**. If passing a plan to apply, the input is the envelope's `result`, not the entire output envelope.

Project setup order is profile plan/apply → generation plan/apply → any later direct agents plan/apply. Generation may create its own `.moai-cli/workflows/<id>/` directory. A direct `agents-plan` requires its referenced workflow directory to exist already. Do not create an empty directory merely to bypass that prerequisite.

Rust project state uses `.moai-cli/`; the home root is `~/.moai-cli/`. Existing Go `.moai/`, `~/.moai/`, and user-owned host configuration remain separate. None of the preview commands above migrates or removes Go state. Read [safety and permissions](/en/reference/safety/) before any mutation.

## Continue with your host

- [Claude Code](/en/guides/claude-code/): predict AGENTS loading, preview project settings, and inspect the launcher.
- [Codex](/en/guides/codex/): inspect instruction discovery and preview an explicit skill projection.
- [CLI reference](/en/reference/cli/): command families, JSON output, and operation boundaries.

A built executable is not a loaded skill. Both host guides distinguish file planning from real host execution and leave live validation explicitly unverified.

## Source notes

These instructions describe the [pinned Rust source](https://github.com/modu-ai/moai-adk/tree/cf33967c42c31f020fd0e35dc07b6b82093119f1), not the preserved Go website:

- [Package manifest](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/Cargo.toml) and [toolchain pin](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/rust-toolchain.toml).
- [Rust execution contract](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/rust-product-path.md), [CLI declarations](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/src/main.rs), and [support matrix](https://github.com/modu-ai/moai-adk/blob/cf33967c42c31f020fd0e35dc07b6b82093119f1/docs/support-matrix.md).
