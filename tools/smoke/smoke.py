#!/usr/bin/env python3
"""The smoke harness the lanes and the auditors share (SPEC-20261119-smoke-harness).

Usage (through the wrapper, `tools/smoke/run.sh`):
  tools/smoke/run.sh <area> [--bin PATH] [--repo ROOT] [--keep]
  area: install | update | project | launcher | all

It boots the BUILT binary in an empty temporary HOME and a temporary git project, runs the
steps of the area and prints one trace line per step:

  smoke: start area=update bin=... root=...
  [ok]   stage-package (0.02s)
  [FAIL] preview (0.31s): exit code 1
    command: ...
    exit: 1
    stdout: ...
    stderr: ...
  smoke: update fail steps=2 failed=preview time=0.4s

The last line is always `smoke: <area> pass|fail steps=<n> ...`, or `smoke: <area> skip
reason=...` on a platform without `pwd` and `pty` (Windows). Exit codes: 0 = pass or skip,
1 = a step failed, 2 = usage error or a refused temporary base.

Isolation: the environment of every step is built from nothing but HOME, PATH, TMPDIR, the
locale, the terminal and the git settings, so the caller's keys, host variables and HOME never
reach a step; the step PATH is the harness's own folder plus /usr/bin and /bin. The temporary
base must not lie inside the caller's real home folder (it is refused before anything is
created); the harness compares file identity, by `stat` on the base's ancestors and on the real
home (HOME and the account entry), so another letter case or a firmlink spelling is refused
too. It lists and opens nothing under the real home. If the temporary root cannot be removed
the harness names it and fails. No network, no real host: the launcher's host is a fake
executable the harness writes. PATH is never used to find the binary under test: a `moai`
wrapper on PATH follows `$HOME/.moai-cli/path`, which an empty HOME does not have.
"""

import argparse
import json
import os
import re
import select
import shutil
import signal
import subprocess
import sys
import tempfile
import time
from pathlib import Path

try:
    import pwd
    import pty
except ImportError:  # Windows has neither the account database nor pseudo-terminals
    pwd = None
    pty = None

AREAS = ("install", "update", "project", "launcher")
REPO = Path(__file__).resolve().parent.parent.parent
TAIL = 4000
STEP_TIMEOUT = 120
PTY_TIMEOUT = 30
CLT = "/Library/Developer/CommandLineTools"
PROJECT_REQUEST = {
    "schema_version": 2,
    "project_id": "smoke-demo",
    "purpose": "smoke harness",
    "tools": ["Rust"],
    "development_methods": ["TDD"],
    "hosts": ["claude", "codex"],
    "interview_revision": "smoke-1",
}
FAKE_HOST = """#!/bin/sh
echo "FAKE-HOST argv: $*"
echo "FAKE-HOST home=$HOME"
if [ -t 0 ]; then echo "FAKE-HOST tty=yes"; else echo "FAKE-HOST tty=no"; fi
exit 0
"""


class StepFailed(Exception):
    """A step did not hold; the message is the reason shown on its FAIL line."""


class Usage(Exception):
    """A usage error or a refused base: exit 2."""


def account_home():
    """The home folder the account entry names, or None."""
    try:
        return pwd.getpwuid(os.getuid()).pw_dir
    except (KeyError, OSError):
        return None


def real_homes(environ=None, account=account_home) -> set:
    """The caller's real home folder (from HOME and the account entry), resolved."""
    homes = set()
    env_home = (os.environ if environ is None else environ).get("HOME")
    if env_home:
        homes.add(os.path.realpath(env_home))
    entry = account()
    if entry:
        homes.add(os.path.realpath(entry))
    homes.discard(os.sep)
    return homes


def same_folder(one: str, other: str) -> bool:
    try:
        return os.path.samefile(one, other)
    except OSError:
        return False


def inside(path: str, folder: str) -> bool:
    """True when `path` is `folder` or lies under it, by file identity: another letter case
    on a case-insensitive volume or a firmlink spelling (`/System/Volumes/Data/Users/...`)
    names the same folder and compares equal here, which a string comparison misses."""
    current = os.path.realpath(path)
    while True:
        if same_folder(current, folder):
            return True
        parent = os.path.dirname(current)
        if parent == current:
            return False
        current = parent


def entries(folder: Path, skip: str = "") -> list:
    """Every entry under `folder` (files, folders and links; links are not followed), as
    relative names; a top-level entry called `skip` is left out."""
    found = []

    def visit(current: Path):
        with os.scandir(current) as listing:
            for item in sorted(listing, key=lambda e: e.name):
                if skip and current == folder and item.name == skip:
                    continue
                found.append(str(Path(item.path).relative_to(folder)))
                if item.is_dir(follow_symlinks=False):
                    visit(Path(item.path))

    visit(folder)
    return found


def make_removable(path: str):
    """Give every folder under `path` back its owner permissions (a step may have left a
    mode-000 folder), so the removal can list it. Links are not followed."""
    try:
        if os.path.islink(path):
            return
        os.chmod(path, 0o700)
        with os.scandir(path) as listing:
            children = [e.path for e in listing if e.is_dir(follow_symlinks=False)]
    except OSError:
        return
    for child in children:
        make_removable(child)


def remove_root(root: Path) -> str:
    """Remove the temporary root; returns "" when it is gone, else why it is still there."""
    reason = ""
    make_removable(str(root))
    try:
        shutil.rmtree(root)
    except OSError as e:
        reason = f"{e.filename or root}: {e.strerror or e}"
    if os.path.lexists(root):
        return reason or "it is still there after the removal"
    return ""


def clip(text: str) -> str:
    return text if len(text) <= TAIL else "[...]" + text[-TAIL:]


class Smoke:
    """One area's run inside `base`: a temporary HOME, a git project and a prefix."""

    def __init__(self, binary: Path, repo: Path, base: Path):
        self.bin = binary
        self.repo = repo
        self.base = base
        self.home = base / "home"
        self.proj = base / "proj"
        self.prefix = base / "prefix"
        self.fakebin = base / "fakebin"
        self.tmp = base / "tmp"
        self.git = shutil.which("git") or "git"
        self.last = None
        self.staged = None
        self.digest = None
        self.plan_text = None
        self.steps = 0
        self.failed = None

    # -- the environment of every step ------------------------------------------------

    def env(self) -> dict:
        # Only the harness's own folder and the system folders: the caller's PATH, and the
        # folder the caller's git lives in (a version manager's shim folder would expose an
        # installed `moai`), never reach a step. `git` is linked into fakebin instead.
        path = [str(self.fakebin), "/usr/bin", "/bin"]
        env = {
            "HOME": str(self.home),
            "PATH": os.pathsep.join(path),
            "TMPDIR": str(self.tmp),
            "LANG": "C",
            "LC_ALL": "C",
            "TERM": "xterm",
            "GIT_CONFIG_GLOBAL": "/dev/null",
            "GIT_CONFIG_NOSYSTEM": "1",
        }
        if os.path.isdir(CLT):
            # `/usr/bin/git` under a fresh HOME runs a license check on every call otherwise.
            env["DEVELOPER_DIR"] = CLT
        return env

    def prepare(self):
        for folder in (self.home, self.proj, self.fakebin, self.tmp):
            folder.mkdir(parents=True)
        if os.path.isabs(self.git) and not (self.fakebin / "git").exists():
            (self.fakebin / "git").symlink_to(os.path.realpath(self.git))
        done = subprocess.run(
            [self.git, "-c", "gc.auto=0", "init", "-q"],
            cwd=self.proj,
            env=self.env(),
            capture_output=True,
            text=True,
        )
        if done.returncode != 0:
            raise StepFailed(f"git init failed: {done.stderr.strip()}")

    # -- running things -----------------------------------------------------------------

    def exec(self, argv, stdin: bytes = b"", timeout: int = STEP_TIMEOUT):
        """Run one command with the isolated environment; remember it for the FAIL block."""
        argv = [str(a) for a in argv]
        self.last = {"command": argv, "exit": None, "stdout": "", "stderr": ""}
        try:
            done = subprocess.run(
                argv,
                input=stdin,
                env=self.env(),
                cwd=self.proj,
                capture_output=True,
                timeout=timeout,
            )
        except subprocess.TimeoutExpired as e:
            self.last["stdout"] = (e.stdout or b"").decode(errors="replace")
            self.last["stderr"] = (e.stderr or b"").decode(errors="replace")
            raise StepFailed(f"timed out after {timeout}s")
        except OSError as e:
            raise StepFailed(f"could not run it: {e}")
        self.last.update(
            exit=done.returncode,
            stdout=done.stdout.decode(errors="replace"),
            stderr=done.stderr.decode(errors="replace"),
        )
        return done

    def envelope(self, done) -> dict:
        """The `result` of a successful JSON envelope (`{"ok": true, "result": {...}}`)."""
        if done.returncode != 0:
            raise StepFailed(f"exit code {done.returncode}")
        try:
            doc = json.loads(done.stdout)
        except ValueError:
            raise StepFailed("the output is not JSON")
        if not isinstance(doc, dict) or doc.get("ok") is not True:
            raise StepFailed("the envelope does not say ok: true")
        result = doc.get("result")
        if not isinstance(result, dict):
            raise StepFailed("the envelope has no result object")
        return result

    def pty_run(self, argv, timeout: int = PTY_TIMEOUT):
        """Run a command with a terminal as its stdin/stdout; returns (exit code, output)."""
        argv = [str(a) for a in argv]
        self.last = {"command": argv, "exit": None, "stdout": "", "stderr": ""}
        pid, fd = pty.fork()
        if pid == 0:
            try:
                os.chdir(self.proj)
                os.execve(argv[0], argv, self.env())
            finally:
                os._exit(127)
        out = b""
        deadline = time.monotonic() + timeout
        timed_out = False
        try:
            while True:
                left = deadline - time.monotonic()
                if left <= 0:
                    timed_out = True
                    break
                ready, _, _ = select.select([fd], [], [], min(left, 0.5))
                if not ready:
                    continue
                try:
                    chunk = os.read(fd, 4096)
                except OSError:
                    break
                if not chunk:
                    break
                out += chunk
        except BaseException:
            # interrupted: do not leave the child running
            try:
                os.kill(pid, signal.SIGKILL)
            except OSError:
                pass
            os.waitpid(pid, 0)
            os.close(fd)
            raise
        if timed_out:
            os.kill(pid, signal.SIGKILL)
        _, status = os.waitpid(pid, 0)
        os.close(fd)
        text = out.decode(errors="replace").replace("\r\n", "\n")
        code = os.waitstatus_to_exitcode(status)
        self.last.update(exit=code, stdout=text)
        if timed_out:
            raise StepFailed(f"timed out after {timeout}s")
        return code, text

    # -- small checks -------------------------------------------------------------------

    def home_entries(self) -> list:
        """Everything under the temporary HOME: files, folders and links."""
        return entries(self.home)

    def project_entries(self) -> list:
        """Everything the steps wrote into the temporary project (its own `.git` aside)."""
        return entries(self.proj, skip=".git")

    # -- the trace ----------------------------------------------------------------------

    def step(self, name: str, fn):
        self.steps += 1
        self.last = None
        started = time.monotonic()
        try:
            fn()
        except Exception as e:  # a StepFailed, or anything the step did not expect
            reason = str(e) if isinstance(e, StepFailed) else f"unexpected {type(e).__name__}: {e}"
            took = time.monotonic() - started
            print(f"[FAIL] {name} ({took:.2f}s): {reason}", flush=True)
            if self.last:
                print(f"  command: {' '.join(self.last['command'])}")
                print(f"  exit: {self.last['exit']}")
                print(f"  stdout: {clip(self.last['stdout']).rstrip()}")
                print(f"  stderr: {clip(self.last['stderr']).rstrip()}", flush=True)
            self.failed = name
            return False
        print(f"[ok]   {name} ({time.monotonic() - started:.2f}s)", flush=True)
        return True

    # -- steps ----------------------------------------------------------------------

    def stage_package(self):
        """The installed shape `update` looks for: <prefix>/bin/moai-cli beside
        <prefix>/skills/moai and <prefix>/agents (src/update.rs `installed_root`)."""
        for source in (self.repo / "skills" / "moai", self.repo / "agents"):
            if not source.is_dir():
                raise StepFailed(f"{source} is not a folder; use --repo to name the checkout")
        (self.prefix / "bin").mkdir(parents=True)
        staged = self.prefix / "bin" / "moai-cli"
        shutil.copy2(self.bin, staged)
        staged.chmod(0o755)
        shutil.copytree(self.repo / "skills" / "moai", self.prefix / "skills" / "moai")
        shutil.copytree(self.repo / "agents", self.prefix / "agents")
        self.staged = staged

    def version(self):
        done = self.exec([self.staged, "--version"])
        if done.returncode != 0:
            raise StepFailed(f"exit code {done.returncode}")
        first = done.stdout.decode(errors="replace").splitlines()[:1]
        if not first or not re.match(r"^moai-cli \S+ \(rust\)$", first[0]):
            raise StepFailed("the first line is not the Rust signature `moai-cli <version> (rust)`")

    def install_inspect(self):
        result = self.envelope(self.exec([self.staged, "install-inspect"]))
        state = result.get("home_state_root")
        if not isinstance(state, str):
            raise StepFailed("the result has no home_state_root")
        if not inside(os.path.realpath(state), os.path.realpath(self.home)):
            raise StepFailed(f"home_state_root {state} is not inside the temporary HOME")

    def home_untouched(self):
        found = self.home_entries()
        if found:
            raise StepFailed(f"the temporary HOME holds {len(found)} entries: {found[:5]}")

    def update_preview(self):
        result = self.envelope(self.exec([self.staged, "update"]))
        if result.get("applied") is not False or result.get("nothing_to_apply") is not False:
            raise StepFailed("the preview does not say applied=false, nothing_to_apply=false")
        digest = result.get("plan_digest")
        if not isinstance(digest, str) or not digest.startswith("sha256:"):
            raise StepFailed("the preview has no plan_digest")
        found = self.home_entries()
        if found:
            raise StepFailed(f"the preview wrote {len(found)} entries: {found[:5]}")
        self.digest = digest

    def update_confirm(self):
        result = self.envelope(self.exec([self.staged, "update", "--confirm", self.digest]))
        if result.get("applied") is not True:
            raise StepFailed("the confirm does not say applied=true")
        skill = self.home / ".claude" / "skills" / "moai"
        if not skill.is_dir() or not any(skill.iterdir()):
            raise StepFailed("the skill copy under the temporary HOME is missing or empty")

    def update_rerun(self):
        result = self.envelope(self.exec([self.staged, "update"]))
        if result.get("nothing_to_apply") is not True:
            raise StepFailed("the second preview still has something to apply")

    def project_detect(self):
        self.envelope(self.exec([self.bin, "project-detect", self.proj]))

    def project_plan(self):
        result = self.envelope(
            self.exec(
                [self.bin, "project-plan", self.proj],
                stdin=json.dumps(PROJECT_REQUEST).encode(),
            )
        )
        digest = result.get("plan_digest")
        if not isinstance(digest, str) or not digest.startswith("sha256:"):
            raise StepFailed("the plan has no plan_digest")
        found = self.project_entries()
        if found:
            raise StepFailed(f"the plan wrote {len(found)} entries: {found[:5]}")
        self.digest = digest
        self.plan_text = json.dumps(result)

    def project_apply(self):
        self.envelope(
            self.exec(
                [
                    self.bin,
                    "project-apply",
                    self.proj,
                    "--plan-digest",
                    self.digest,
                    "--permission-reference",
                    "smoke-harness",
                ],
                stdin=self.plan_text.encode(),
            )
        )
        profiles = [self.proj / d / "project.yaml" for d in (".moai-cli", ".moai")]
        if not any(p.is_file() for p in profiles):
            raise StepFailed("no project.yaml under .moai-cli/ or .moai/ after the apply")

    def launcher_dry_run(self):
        done = self.exec([self.bin, "cc", "-a", "--dry-run"])
        if done.returncode != 0:
            raise StepFailed(f"exit code {done.returncode}")
        try:
            plan = json.loads(done.stdout)
        except ValueError:
            raise StepFailed("the output is not JSON")
        argv = plan.get("argv") if isinstance(plan, dict) else None
        if not argv or argv[0] != "claude":
            raise StepFailed("the dry run does not name `claude` as the host command")
        if plan.get("executed") is not False:
            raise StepFailed("the dry run says it started the host")

    def launcher_host_in_pty(self):
        host = self.fakebin / "claude"
        host.write_text(FAKE_HOST)
        host.chmod(0o755)
        code, text = self.pty_run([self.bin, "cc", "-a"])
        if code != 0:
            raise StepFailed(f"exit code {code}")
        if not re.search(r"FAKE-HOST argv: .*--name agent-\d+", text):
            raise StepFailed("the fake host did not get `--name agent-<n>`")
        if "FAKE-HOST tty=yes" not in text:
            raise StepFailed("the fake host did not get a terminal")
        seen = re.search(r"FAKE-HOST home=(\S+)", text)
        if not seen or os.path.realpath(seen.group(1)) != os.path.realpath(self.home):
            raise StepFailed("the fake host did not see the temporary HOME")

    def plan(self, area: str) -> list:
        return {
            "install": [
                ("stage-package", self.stage_package),
                ("version", self.version),
                ("install-inspect", self.install_inspect),
                ("home-untouched", self.home_untouched),
            ],
            "update": [
                ("stage-package", self.stage_package),
                ("preview", self.update_preview),
                ("confirm", self.update_confirm),
                ("rerun", self.update_rerun),
            ],
            "project": [
                ("detect", self.project_detect),
                ("plan", self.project_plan),
                ("apply", self.project_apply),
            ],
            "launcher": [
                ("dry-run", self.launcher_dry_run),
                ("host-in-pty", self.launcher_host_in_pty),
            ],
        }[area]


def resolve_binary(arg, repo: Path) -> Path:
    if arg:
        binary = Path(arg)
    elif os.environ.get("MOAI_SMOKE_BIN"):
        binary = Path(os.environ["MOAI_SMOKE_BIN"])
    else:
        target = os.environ.get("CARGO_TARGET_DIR")
        binary = (Path(target) if target else repo / "target") / "debug" / "moai-cli"
    if not binary.is_file():
        raise Usage(f"binary not found: {binary} (build it, or pass --bin)")
    if not os.access(binary, os.X_OK):
        raise Usage(f"binary is not executable: {binary}")
    return binary.resolve()


def temp_base(tmpdir=None, homes=None) -> str:
    """The folder the temporary root goes in; refused when it is inside the real home
    (compared by file identity, see `inside`)."""
    base = tmpdir or os.environ.get("TMPDIR") or "/tmp"
    if not os.path.isdir(base):
        raise Usage(f"the temporary folder base {base} is not a folder")
    resolved = os.path.realpath(base)
    for home in sorted(real_homes() if homes is None else homes):
        if inside(resolved, home):
            raise Usage(
                f"the temporary folder base {base} lies inside the real home folder {home}; "
                "set TMPDIR to a folder outside it"
            )
    return base


def run_area(area: str, binary: Path, repo: Path, base: Path, name_in_all: bool) -> tuple:
    """Runs one area in `base`; returns (steps, failed step or None)."""
    smoke = Smoke(binary, repo, base)
    if name_in_all:
        print(f"== area {area} ==", flush=True)
    try:
        smoke.prepare()
    except (StepFailed, OSError) as e:
        print(f"[FAIL] prepare (0.00s): {e}", flush=True)
        return 1, "prepare"
    for name, fn in smoke.plan(area):
        if not smoke.step(name, fn):
            break
    return smoke.steps, smoke.failed


def terminate(signum, frame):
    raise KeyboardInterrupt


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(prog="run.sh", description="MoAI smoke harness")
    parser.add_argument("area", help="install | update | project | launcher | all")
    parser.add_argument("--bin", help="the moai-cli binary under test")
    parser.add_argument("--repo", default=str(REPO), help="the checkout holding skills/ and agents/")
    parser.add_argument("--keep", action="store_true", help="keep the temporary root")
    args = parser.parse_args(argv)
    if pwd is None or pty is None:
        # The harness boots the binary through `/usr/bin`, a pseudo-terminal and the account
        # entry, which only POSIX has; the CI smoke step runs on every leg, so it says why
        # it does not run and succeeds before any step or file is touched.
        print(
            f"smoke: {args.area} skip reason=the harness needs pwd and pty, "
            "which this platform does not have (POSIX only)",
            flush=True,
        )
        return 0
    started = time.monotonic()
    try:
        if args.area != "all" and args.area not in AREAS:
            raise Usage(f"unknown area '{args.area}' (install, update, project, launcher, all)")
        repo = Path(args.repo).resolve()
        binary = resolve_binary(args.bin, repo)
        base = temp_base()
    except Usage as e:
        print(f"smoke: {e}", file=sys.stderr)
        return 2
    root = Path(tempfile.mkdtemp(prefix="moai-smoke-", dir=base)).resolve()
    print(f"smoke: start area={args.area} bin={binary} root={root}", flush=True)
    steps = 0
    failed = None
    previous = signal.signal(signal.SIGTERM, terminate)
    try:
        areas = AREAS if args.area == "all" else (args.area,)
        for area in areas:
            where = root / area if args.area == "all" else root
            count, bad = run_area(area, binary, repo, where, args.area == "all")
            steps += count
            if bad:
                failed = f"{area}/{bad}" if args.area == "all" else bad
                break
    except KeyboardInterrupt:
        # SIGTERM or Ctrl-C: the running command is already stopped; clean up and say so.
        failed = "interrupted"
        print("[FAIL] interrupted: the run was stopped by a signal", flush=True)
    finally:
        signal.signal(signal.SIGTERM, previous)
        if args.keep:
            print(f"smoke: kept root={root}", flush=True)
        else:
            left = remove_root(root)
            if left:
                print(f"smoke: could not remove root={root}: {left}", flush=True)
                failed = failed or "cleanup"
    took = time.monotonic() - started
    if failed:
        print(f"smoke: {args.area} fail steps={steps} failed={failed} time={took:.1f}s", flush=True)
        return 1
    print(f"smoke: {args.area} pass steps={steps} time={took:.1f}s", flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main())
