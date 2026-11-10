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
- on macOS the run is wrapped in `sandbox-exec` denying every network operation
  and every write under the real HOME.

Variants: `plain` sets NO_COLOR=1; `ansi` sets COLORTERM=truecolor.

Usage:
  python3 -I tools/statusline_golden.py --bin <statusline binary> [--argv statusline] --write
  python3 -I tools/statusline_golden.py --bin <statusline binary> --check
Exit codes: 0 = all written / all equal, 1 = a mismatch or a failed run, 2 = usage.
"""

import argparse
import hashlib
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
FIXTURES = REPO / "tests" / "fixtures" / "statusline"
VARIANTS = {"plain": {"NO_COLOR": "1"}, "ansi": {"COLORTERM": "truecolor"}}
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
    step says otherwise; every write lands inside the temporary root."""
    kind = step["kind"]
    if kind == "write":
        base = moai_home if step.get("under") == "moai_home" else project
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
        stdin_raw = (FIXTURES / case["stdin"]).read_text(encoding="utf-8")
        stdin = stdin_raw.replace("{PROJECT}", str(project))
        env = {"PATH": "/usr/bin:/bin", "HOME": str(home), "MOAI_HOME": str(moai_home),
               "TZ": "UTC", "LANG": "C.UTF-8"}
        env.update(VARIANTS[variant])
        cmd = [str(exe), *argv]
        if sys.platform == "darwin":
            real_home = os.path.realpath(os.path.expanduser("~"))
            profile = (f'(version 1)(allow default)(deny network*)'
                       f'(deny file-write* (subpath "{real_home}"))')
            cmd = ["/usr/bin/sandbox-exec", "-p", profile, *cmd]
        proc = subprocess.run(cmd, input=stdin.encode(), cwd=project, env=env,
                              capture_output=True, timeout=30)
        return proc.returncode, proc.stdout.decode("utf-8", "replace"), proc.stderr.decode("utf-8", "replace")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--bin", required=True)
    ap.add_argument("--argv", default="statusline", help="arguments after the binary, space separated")
    mode = ap.add_mutually_exclusive_group(required=True)
    mode.add_argument("--write", action="store_true")
    mode.add_argument("--check", action="store_true")
    args = ap.parse_args()
    binary = Path(args.bin).resolve()
    if not binary.is_file():
        print(f"binary not found: {binary}")
        return 2
    cases = json.loads((FIXTURES / "cases.json").read_text(encoding="utf-8"))["cases"]
    print(f"binary sha256 {hashlib.sha256(binary.read_bytes()).hexdigest()}")
    failed = 0
    for case in cases:
        for variant in VARIANTS:
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
            elif golden.read_text(encoding="utf-8") != out:
                print(f"DIFF {case['name']} {variant}")
                failed += 1
            else:
                print(f"ok   {case['name']} {variant}")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
