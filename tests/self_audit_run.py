"""Run one PR body check through `moai-cli pr-self-audit-check` (Issue #2708).

The Python checker `tools/check_pr_self_audit.py` is gone (leader decision 6863668555); the
three checker test files ran every case through both checkers until then (#2716) and now run
them through the Rust command alone, with the same exit codes and stdout lines.

The binary is `target/debug/moai-cli` (or `MOAI_CLI_BIN`); CI builds it before the Python
suite runs. A missing binary fails the case rather than skipping it.
"""

import os
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[1]


def rust_bin():
    explicit = os.environ.get("MOAI_CLI_BIN")
    if explicit:
        return pathlib.Path(explicit)
    target = os.environ.get("CARGO_TARGET_DIR")
    base = pathlib.Path(target) if target else ROOT / "target"
    return base / "debug" / "moai-cli"


def run_check(args=(), text=None, stdin=True, cwd=None):
    """The `moai-cli pr-self-audit-check` run for `args` and `text` on stdin."""
    binary = rust_bin()
    if not binary.exists():
        raise AssertionError(f"moai-cli not built at {binary}; run cargo build first")
    return subprocess.run(
        [str(binary), "pr-self-audit-check", *args],
        input=text,
        text=True,
        capture_output=True,
        stdin=None if stdin else subprocess.DEVNULL,
        cwd=cwd,
    )
