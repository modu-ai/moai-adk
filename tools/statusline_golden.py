#!/usr/bin/env python3
"""Statusline equivalence harness (#2331, stage 1).

Runs a statusline binary against the fixture cases in
tests/fixtures/statusline/cases.json and writes (--write) or compares (--check)
the golden outputs. Stage 1 recorded the goldens from the Go `moai statusline`
(moai-adk-go 4fa2aa9db2); stage 2 runs the same harness against the Rust
implementation with --check.

Isolation, per case and per color variant:
- a fresh temporary HOME, MOAI_HOME and project directory (`demo-project`);
- an environment built from scratch (no inherited ANTHROPIC_*, CLAUDE_*, MOAI_*,
  NO_COLOR or COLORTERM values), TZ=UTC;
- the binary is copied under a name other than `moai`, so the Go statusline's
  detached `gh`/`glab` refresh child is never spawned (isSelfInvocable);
- on macOS the run is wrapped in `sandbox-exec` denying every network operation,
  every read and write under the real HOME (so no credential can be used), and
  every process exec except the binary under test and git (#2350 AC-2, AC-5; git during the
  render is allowed, user decision 2026-11-10).

Variants: `plain` sets NO_COLOR=1; `ansi` sets COLORTERM=truecolor; `ansi256` sets
neither, which is the ANSI256 palette path (#2350).

Version token (#2350): the goldens carry the version of the binary that recorded them
(`GOLDEN_BINARY_VERSION`). `--check` replaces `🗿 v<--binary-version>` in the actual
output and `🗿 v<GOLDEN_BINARY_VERSION>` in the golden with one placeholder before
comparing. Nothing else is normalized.

Usage:
  python3 -I tools/statusline_golden.py --bin <statusline binary> [--argv statusline] --write [--variant V] [--case NAME]
  python3 -I tools/statusline_golden.py --bin <statusline binary> [--binary-version <v>] --check
(`--binary-version` defaults to the last token `<bin> --version` prints, `v` stripped.)
Exit codes: 0 = all written / all equal, 1 = a mismatch or a failed run, 2 = usage.
"""

import argparse
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
FIXTURES = REPO / "tests" / "fixtures" / "statusline"
VARIANTS = {"plain": {"NO_COLOR": "1"}, "ansi": {"COLORTERM": "truecolor"}, "ansi256": {}}
# The Go `moai statusline` the goldens were recorded with reports this version; a case
# recorded later names its own binary in `recorded_with` (#2350 audit F1).
GOLDEN_BINARY_VERSION = "3.2.1"
VERSION_PLACEHOLDER = "🗿 v<binary-version>"
GIT_ENV = {
    "GIT_AUTHOR_NAME": "fixture",
    "GIT_AUTHOR_EMAIL": "fixture@example.invalid",
    "GIT_COMMITTER_NAME": "fixture",
    "GIT_COMMITTER_EMAIL": "fixture@example.invalid",
    "GIT_AUTHOR_DATE": "1735689600 +0000",
    "GIT_COMMITTER_DATE": "1735689600 +0000",
    "GIT_CONFIG_GLOBAL": "/dev/null",
    "GIT_CONFIG_NOSYSTEM": "1",
}


def run_git(cwd, *args):
    env = dict(GIT_ENV, PATH=os.environ.get("PATH", "/usr/bin:/bin"))
    subprocess.run(["git", "-C", str(cwd), *args], env=env, check=True,
                   capture_output=True)


def apply_setup(step, root, project, moai_home):
    """One setup step from cases.json. Paths are relative to the project unless the
    step says otherwise (`moai_home`: the Go MOAI_HOME the goldens were recorded with;
    `rc_home`: HOME/.moai-cli, the folder moai-cli reads, #2455); every write lands
    inside the temporary root."""
    kind = step["kind"]
    if kind == "write":
        under = step.get("under")
        if under == "moai_home":
            base = moai_home
        elif under == "rc_home":
            base = moai_home.parent / ".moai-cli"
        elif under is None:
            base = project
        else:
            raise ValueError(f"unknown write base {under!r}")
        target = base / step["path"]
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(step["content"], encoding="utf-8")
    elif kind == "git_init":
        run_git(project, "init", "-q", "-b", step.get("branch", "main"))
    elif kind == "git_commit_all":
        run_git(project, "add", "-A")
        run_git(project, "commit", "-q", "--allow-empty", "-m", step.get("message", "fixture"))
    elif kind == "git_upstream":
        bare = root / "origin.git"
        subprocess.run(["git", "init", "-q", "--bare", str(bare)], check=True,
                       capture_output=True, env=dict(GIT_ENV, PATH=os.environ.get("PATH", "")))
        run_git(project, "remote", "add", "origin", str(bare))
        run_git(project, "push", "-q", "-u", "origin", step.get("branch", "main"))
    else:
        raise ValueError(f"unknown setup kind {kind!r}")


def git_exec_clauses():
    """Sandbox clauses for what a `git` call may exec on macOS: the /usr/bin shim, xcrun,
    and the developer-tools bin directory the shim resolves git into (git runs more than
    the one `git` file from there; measured on #2350)."""
    clauses = ['(literal "/usr/bin/git")', '(literal "/usr/bin/xcrun")']
    found = subprocess.run(["/usr/bin/xcrun", "--find", "git"], capture_output=True, text=True)
    if found.returncode == 0 and found.stdout.strip():
        dev_bin = os.path.dirname(os.path.realpath(found.stdout.strip()))
        clauses.append(f'(subpath "{dev_bin}")')
    return clauses


def sandbox_profile(exe):
    """Deny all network, every read or write under the real HOME, and every exec except
    the binary under test and git."""
    real_home = os.path.realpath(os.path.expanduser("~"))
    allowed = " ".join([f'(literal "{exe}")', *git_exec_clauses()])
    return (f'(version 1)(allow default)(deny network*)'
            f'(deny file-read* (subpath "{real_home}"))'
            f'(deny file-write* (subpath "{real_home}"))'
            f'(deny process-exec*)(allow process-exec* {allowed})')


def normalize_version(text, version):
    return text.replace(f"🗿 v{version}", VERSION_PLACEHOLDER)


VERSION_TOKEN = re.compile(r"v?(\d+\.\d+\.\d+[0-9A-Za-z.+-]*)")


def parse_version_output(returncode, stdout):
    """The version in the last token of a `--version` run (`moai-cli 2026.11.0-rc.1`,
    `moai version v3.2.1`), without a leading `v`; None for a failed run or no version."""
    tokens = stdout.split()
    if returncode != 0 or not tokens:
        return None
    match = VERSION_TOKEN.fullmatch(tokens[-1])
    return match.group(1) if match else None


def detect_binary_version(binary):
    """`<binary> --version`, run the way the cases run it: a copy under another name, a
    throwaway HOME, a minimal PATH and, on macOS, the same sandbox."""
    with tempfile.TemporaryDirectory(prefix="sl-version-") as tmp:
        root = Path(tmp).resolve()
        exe = root / "bin" / "sl-under-test"
        exe.parent.mkdir()
        (root / "home").mkdir()
        shutil.copy2(binary, exe)
        cmd = [str(exe), "--version"]
        if sys.platform == "darwin":
            cmd = ["/usr/bin/sandbox-exec", "-p", sandbox_profile(exe), *cmd]
        try:
            proc = subprocess.run(cmd, capture_output=True, text=True, cwd=root, timeout=30,
                                  env={"PATH": "/usr/bin:/bin", "HOME": str(root / "home")})
        except OSError:
            return None
    return parse_version_output(proc.returncode, proc.stdout)


def run_case(binary, argv, case, variant):
    with tempfile.TemporaryDirectory(prefix="sl-golden-") as tmp:
        root = Path(tmp).resolve()
        home = root / "home"
        moai_home = home / ".moai"
        project = root / "work" / "demo-project"
        for d in (home, moai_home, project):
            d.mkdir(parents=True, exist_ok=True)
        for step in case.get("setup", []):
            apply_setup(step, root, project, moai_home)
        exe = root / "bin" / "sl-under-test"
        exe.parent.mkdir()
        shutil.copy2(binary, exe)
        # Bytes, not text: a fixture may hold invalid UTF-8 on purpose (#2350 audit F1).
        stdin_raw = (FIXTURES / case["stdin"]).read_bytes()
        stdin = stdin_raw.replace(b"{PROJECT}", str(project).encode())
        env = {"PATH": "/usr/bin:/bin", "HOME": str(home), "MOAI_HOME": str(moai_home),
               "TZ": "UTC", "LANG": "C.UTF-8"}
        env.update(VARIANTS[variant])
        cmd = [str(exe), *argv]
        if sys.platform == "darwin":
            cmd = ["/usr/bin/sandbox-exec", "-p", sandbox_profile(exe), *cmd]
        proc = subprocess.run(cmd, input=stdin, cwd=project, env=env,
                              capture_output=True, timeout=30)
        return proc.returncode, proc.stdout.decode("utf-8", "replace"), proc.stderr.decode("utf-8", "replace")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--bin", required=True)
    ap.add_argument("--argv", default="statusline", help="arguments after the binary, space separated")
    mode = ap.add_mutually_exclusive_group(required=True)
    mode.add_argument("--write", action="store_true")
    mode.add_argument("--check", action="store_true")
    ap.add_argument("--binary-version",
                    help="version the binary under --check reports in its 🗿 segment "
                         "(default: read from `<bin> --version`)")
    ap.add_argument("--variant", choices=sorted(VARIANTS), action="append",
                    help="restrict to these variants (default: all)")
    ap.add_argument("--case", action="append",
                    help="restrict to these case names (default: all)")
    args = ap.parse_args()
    binary = Path(args.bin).resolve()
    if not binary.is_file():
        print(f"binary not found: {binary}")
        return 2
    cases = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))["cases"]
    variants = args.variant or list(VARIANTS)
    unknown = sorted(set(args.case or []) - {c["name"] for c in cases})
    if unknown:
        print(f"unknown case: {', '.join(unknown)}")
        return 2
    print(f"binary sha256 {hashlib.sha256(binary.read_bytes()).hexdigest()}")
    if args.check and not args.binary_version:
        args.binary_version = detect_binary_version(binary)
        if not args.binary_version:
            print("binary version unknown: `--version` printed no version; pass --binary-version")
            return 2
    if args.check:
        print(f"binary version {args.binary_version} (goldens: {GOLDEN_BINARY_VERSION})")
    failed = 0
    for case in cases:
        if args.case and case["name"] not in args.case:
            continue
        for variant in variants:
            code, out, err = run_case(binary, args.argv.split(), case, variant)
            golden = FIXTURES / "golden" / f"{case['name']}.{variant}.txt"
            if code != 0:
                print(f"FAIL {case['name']} {variant}: exit {code} {err.strip()[:200]}")
                failed += 1
                continue
            if args.write:
                golden.parent.mkdir(parents=True, exist_ok=True)
                golden.write_text(out, encoding="utf-8")
                print(f"wrote {golden.relative_to(REPO)}")
                continue
            recorded = case.get("recorded_with", {}).get("version", GOLDEN_BINARY_VERSION)
            want = normalize_version(golden.read_text(encoding="utf-8"), recorded)
            got = normalize_version(out, args.binary_version)
            if want != got:
                print(f"DIFF {case['name']} {variant}")
                failed += 1
            else:
                print(f"ok   {case['name']} {variant}")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
