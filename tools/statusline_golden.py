#!/usr/bin/env python3
"""Claude statusline golden harness (#2376).

Runs the Rust binary against tests/fixtures/statusline-claude/cases.json.
--write records real program output and its binary provenance; --check compares it.
The historical Go assets in tests/fixtures/statusline remain a read-only archive,
including byte-sensitive stdin reused through relative paths by the new manifest.

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

Variants: `plain` pipes stdout with NO_COLOR=1; `ansi` and `ansi256` attach
stdout to a real POSIX PTY with and without COLORTERM=truecolor. Claude Code
reads the statusline through a pipe, so the color contract also requires a pipe
without NO_COLOR to carry the same palette as the PTY (escape-free output there
was the #2559 regression) with the same text once the escapes are removed. PTY output processing is disabled to preserve raw LF.

Version token: each recorded case names the actual binary version and SHA-256.
--check normalizes only the `🗿 v<version>` token on each side.

Usage:
  python3 -I tools/statusline_golden.py --bin <statusline binary> [--argv "<arguments>"] --write [--variant V] [--case NAME]
  python3 -I tools/statusline_golden.py --bin <statusline binary> [--binary-version <v>] --check
(`--binary-version` defaults to the last token `<bin> --version` prints, `v` stripped.)
(`--argv` defaults to DEFAULT_ARGV: the hidden seams that make the render independent of the
machine, so recording and checking produce the same text; an explicit `--argv` must carry them too.)
Exit codes: 0 = all written / all equal, 1 = a mismatch or a failed run, 2 = usage.
"""

import argparse
import errno
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import threading
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
FIXTURES = REPO / "tests" / "fixtures" / "statusline-claude"
VARIANTS = {"plain": {"NO_COLOR": "1"}, "ansi": {"COLORTERM": "truecolor"}, "ansi256": {}}
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
    `cli_home`: HOME/.moai-cli, the folder moai-cli reads, #2455); every write lands
    inside the temporary root."""
    kind = step["kind"]
    if kind == "write":
        under = step.get("under")
        if under == "moai_home":
            base = moai_home
        elif under == "cli_home":
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


def developer_env():
    """`DEVELOPER_DIR` from this process, for every child run under the sandbox (#2699).
    `git_exec_clauses` resolves git with this process's environment, and `cargo test` sets
    `DEVELOPER_DIR` (.cargo/config.toml); a child without it would resolve git into another
    developer folder that the sandbox does not allow, and every git case would differ."""
    value = os.environ.get("DEVELOPER_DIR")
    return {"DEVELOPER_DIR": value} if value else {}


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
# An OSC 8 hyperlink opener or closer: ESC ] 8 ; params ; URI ST (ST = ESC \).
OSC8_LINK = re.compile(r"\x1b\]8;[^;\x1b]*;[^\x1b]*\x1b\\")


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
                                  env={"PATH": "/usr/bin:/bin", "HOME": str(root / "home"),
                                       **developer_env()})
        except OSError:
            return None
    return parse_version_output(proc.returncode, proc.stdout)


def run_process(cmd, stdin, cwd, env, tty):
    """Drain the PTY concurrently with stdin/stderr to avoid pipe-buffer deadlocks."""
    if not tty:
        proc = subprocess.run(cmd, input=stdin, cwd=cwd, env=env,
                              capture_output=True, timeout=30)
        return proc.returncode, proc.stdout, proc.stderr
    import pty
    import termios

    master, slave = pty.openpty()
    chunks = []
    errors = []
    proc = None
    reader = None
    try:
        attrs = termios.tcgetattr(slave)
        attrs[1] &= ~(termios.OPOST | termios.ONLCR)
        termios.tcsetattr(slave, termios.TCSANOW, attrs)
        proc = subprocess.Popen(cmd, stdin=subprocess.PIPE, stdout=slave,
                                stderr=subprocess.PIPE, cwd=cwd, env=env)
        os.close(slave)
        slave = None

        def drain():
            try:
                while True:
                    chunk = os.read(master, 65536)
                    if not chunk:
                        break
                    chunks.append(chunk)
            except OSError as exc:
                # Linux signals the last slave closing with EIO; macOS returns EOF.
                if exc.errno != errno.EIO:
                    errors.append(exc)

        reader = threading.Thread(target=drain, daemon=True)
        reader.start()
        try:
            _, stderr = proc.communicate(stdin, timeout=30)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.communicate()
            raise
        reader.join(timeout=1)
        if reader.is_alive():
            raise TimeoutError("PTY stdout did not close after the binary exited")
        if errors:
            raise errors[0]
        return proc.returncode, b"".join(chunks), stderr
    finally:
        if slave is not None:
            os.close(slave)
        os.close(master)
        if proc is not None and proc.poll() is None:
            proc.kill()
            proc.communicate()


def run_case(binary, argv, case, variant, stdin_bytes=None, stdout_tty=None):
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
        stdin_raw = ((FIXTURES / case["stdin"]).read_bytes()
                     if stdin_bytes is None else stdin_bytes)
        stdin = stdin_raw.replace(b"{PROJECT}", str(project).encode())
        env = {"PATH": "/usr/bin:/bin", "HOME": str(home), "MOAI_HOME": str(moai_home),
               "TZ": "UTC", "LANG": "C.UTF-8", **developer_env()}
        env.update(VARIANTS[variant])
        cmd = [str(exe), *argv]
        if sys.platform == "darwin":
            cmd = ["/usr/bin/sandbox-exec", "-p", sandbox_profile(exe), *cmd]
        tty = variant != "plain" if stdout_tty is None else stdout_tty
        code, stdout, stderr = run_process(cmd, stdin, project, env, tty)
        return code, stdout.decode("utf-8", "replace"), stderr.decode("utf-8", "replace")


def check_color_contract(binary, argv):
    """Color on a pipe (how Claude Code runs it), NO_COLOR on a TTY, and both palettes.

    #2559: a pipe without NO_COLOR must be colored; an escape-free pipe fails here."""
    case = {"stdin": "../statusline/stdin/a02-doc-full.json"}
    code, plain, err = run_case(binary, argv, case, "plain")
    if code or not plain.strip() or "\x1b" in plain:
        raise ValueError(f"plain pipe failed: {code} {plain!r} {err!r}")
    for label, variant, options in [
        ("pipe without NO_COLOR", "ansi256", {"stdout_tty": False}),
        ("truecolor pipe", "ansi", {"stdout_tty": False}),
        ("TTY with NO_COLOR", "plain", {"stdout_tty": True}),
        ("truecolor TTY", "ansi", {}),
        ("ANSI256 TTY", "ansi256", {}),
    ]:
        code, out, err = run_case(binary, argv, case, variant, **options)
        # SGR colors and (#2756 AC-7) OSC 8 hyperlinks are the only escapes color adds.
        if code or OSC8_LINK.sub("", re.sub(r"\x1b\[[0-9;]*m", "", out)) != plain:
            raise ValueError(f"{label} content differs: {code} {out!r} {err!r}")
        if variant == "plain" and "\x1b" in out:
            raise ValueError(f"{label} contains terminal escapes")
        if variant == "ansi" and "38;2;" not in out:
            raise ValueError("truecolor TTY has no truecolor escapes")
        if variant == "ansi256" and ("38;5;" not in out or "38;2;" in out):
            raise ValueError("ANSI256 TTY does not use the ANSI256 palette")
    print("ok   native stdout color contract")


def check_input_boundaries(binary, argv):
    """Exercise the native stdin cap, including a valid JSON prefix beyond the cap."""
    case = {"stdin": "../statusline/stdin/a01-minimal.json"}
    code, fallback, err = run_case(binary, argv, case, "plain", stdin_bytes=b"")
    if code or not fallback.strip() or ("MoAI" not in fallback and "🗿" not in fallback):
        raise ValueError(f"empty-input fallback failed: {code} {fallback!r} {err!r}")
    cap = 4 * 1024 * 1024
    valid = b'{"model":{"display_name":"BoundaryModel"}}'
    exact = valid + b" " * (cap - len(valid))
    code, out, err = run_case(binary, argv, case, "plain", stdin_bytes=exact)
    if code or "BoundaryModel" not in out or "\x1b" in out:
        raise ValueError(f"exact-cap valid input failed: {code} {out!r} {err!r}")
    for label, raw in [
        ("malformed", b"{"),
        ("invalid typed field", b'{"model":{"display_name":42}}'),
        ("oversized JSON string", b'{"session_name":"' + b"x" * cap + b'"}'),
        ("valid prefix over cap", exact + b" "),
    ]:
        code, out, err = run_case(binary, argv, case, "plain", stdin_bytes=raw)
        if code or out != fallback or "\x1b" in out:
            raise ValueError(f"{label} did not fall back: {code} {out!r} {err!r}")
    print("ok   native stdin boundaries")


# Recording and checking share these hidden seams. `--render-budget-ms`: the goldens describe the
# render within budget, so a slow git cannot drop the git segment (#2350). `--now-unix`: the prompt
# cache timer counts from a fixed clock (#2756 AC-8).
DEFAULT_ARGV = "statusline --render-budget-ms 3600000 --now-unix 1738429000"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--bin", required=True)
    ap.add_argument("--argv", default=DEFAULT_ARGV, help="arguments after the binary, space separated")
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
    manifest_path = FIXTURES / "cases.json"
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    cases = manifest["cases"]
    variants = args.variant or list(VARIANTS)
    unknown = sorted(set(args.case or []) - {c["name"] for c in cases})
    if unknown:
        print(f"unknown case: {', '.join(unknown)}")
        return 2
    cases = [case for case in cases if not args.case or case["name"] in args.case]
    if not cases or not variants:
        print("no selected cases or variants")
        return 2
    digest = hashlib.sha256(binary.read_bytes()).hexdigest()
    print(f"binary sha256 {digest}")
    if not args.binary_version:
        args.binary_version = detect_binary_version(binary)
        if not args.binary_version:
            print("binary version unknown: `--version` printed no version; pass --binary-version")
            return 2
    print(f"binary version {args.binary_version}")
    failed = 0
    checked = 0
    for case in cases:
        recorded_variants = case.get("recorded_with", {}).get("variants", {})
        for variant in variants:
            try:
                code, out, err = run_case(binary, args.argv.split(), case, variant)
                if code != 0 or not out.strip():
                    raise ValueError(f"exit {code}, stdout {out!r}, stderr {err.strip()[:200]}")
                if variant == "plain" and "\x1b" in out:
                    raise ValueError("NO_COLOR output contains terminal escapes")
                if variant != "plain":
                    # #2559: Claude Code reads stdout through a pipe; the piped bytes must
                    # equal the PTY bytes (an escape-free pipe was the regression).
                    pipe_code, pipe_out, pipe_err = run_case(
                        binary, args.argv.split(), case, variant, stdout_tty=False)
                    if pipe_code or pipe_out != out:
                        raise ValueError(
                            f"pipe output differs from the PTY: {pipe_code} {pipe_out!r} {pipe_err!r}")
                golden = FIXTURES / "golden" / f"{case['name']}.{variant}.txt"
                if args.write:
                    golden.parent.mkdir(parents=True, exist_ok=True)
                    golden.write_text(out, encoding="utf-8")
                    recorded_variants[variant] = {
                        "version": args.binary_version, "sha256": digest,
                        "stdout": "pipe" if variant == "plain" else "PTY",
                    }
                    case["recorded_with"] = {
                        "implementation": "Rust moai-cli", "variants": recorded_variants,
                    }
                    manifest["recording"] = "Actual native binary outputs; provenance is recorded per case and variant."
                    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
                    print(f"wrote {golden.relative_to(REPO)}")
                    checked += 1
                    continue
                recorded = recorded_variants[variant]["version"]
                want = normalize_version(golden.read_text(encoding="utf-8"), recorded)
                got = normalize_version(out, args.binary_version)
                if want != got:
                    raise ValueError("golden differs")
                print(f"ok   {case['name']} {variant}")
                checked += 1
            except (OSError, ValueError, KeyError, subprocess.TimeoutExpired) as exc:
                print(f"FAIL {case['name']} {variant}: {exc}")
                failed += 1
    try:
        check_color_contract(binary, args.argv.split())
        check_input_boundaries(binary, args.argv.split())
    except (OSError, ValueError, subprocess.TimeoutExpired) as exc:
        print(f"FAIL native runtime contract: {exc}")
        failed += 1
    if checked == 0:
        print("FAIL no golden outputs checked or recorded")
        failed += 1
    print(f"{'recorded' if args.write else 'checked'} {checked} golden outputs")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
