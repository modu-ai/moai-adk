"""Run one PR body check through both checkers (Issue #2708).

`tools/check_pr_self_audit.py` moves to `moai-cli pr-self-audit-check`. Until the Python
file is removed, every case of the three checker test files runs through both, and the two
must agree: the same exit code and, outside a usage error (exit 2), the same stdout. The
Python result is returned, so each test's own assertions stay as they were.

The Rust binary is `target/debug/moai-cli` (or `MOAI_CLI_BIN`); CI builds it before the
Python suite runs. A missing binary fails the case rather than skipping it.
"""

import os
import pathlib
import subprocess

ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOL = ROOT / "tools/check_pr_self_audit.py"


def rust_bin():
    explicit = os.environ.get("MOAI_CLI_BIN")
    if explicit:
        return pathlib.Path(explicit)
    target = os.environ.get("CARGO_TARGET_DIR")
    base = pathlib.Path(target) if target else ROOT / "target"
    return base / "debug" / "moai-cli"


def run_both(args=(), text=None, stdin=True, cwd=None):
    """(Python result) after checking that the Rust command agrees with it."""
    kwargs = dict(
        input=text,
        text=True,
        capture_output=True,
        stdin=None if stdin else subprocess.DEVNULL,
        cwd=cwd,
    )
    py = subprocess.run(["python3", "-I", str(TOOL), *args], **kwargs)
    binary = rust_bin()
    if not binary.exists():
        raise AssertionError(f"moai-cli not built at {binary}; run cargo build first")
    rs = subprocess.run([str(binary), "pr-self-audit-check", *args], **kwargs)
    if py.returncode != rs.returncode or (py.returncode != 2 and py.stdout != rs.stdout):
        raise AssertionError(
            "pr-self-audit-check disagrees with tools/check_pr_self_audit.py\n"
            f"args={list(args)!r}\n"
            f"python exit={py.returncode} stdout={py.stdout!r}\n"
            f"rust   exit={rs.returncode} stdout={rs.stdout!r} stderr={rs.stderr[-400:]!r}\n"
            f"body={text!r}"
        )
    return py
