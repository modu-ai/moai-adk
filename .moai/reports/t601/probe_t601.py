"""t601 (H06) reproduction probe: five cells against one gate-hook copy.

Cells:
  P0  control        : checks pass -> silent pass; second call must not re-run.
  P1  fail -> recall : is the stored block re-delivered byte-identically?
  P2a interrupted    : a fresh "running" record -> notice, no re-run.
  P2b interrupted    : a stale "running" record -> checks re-run.
  P3  fixed tree     : work tree repaired, HEAD unchanged -> should re-run and allow.

Usage: python3 probe_t601.py <path-to-sync-phase-quality-gate.sh>
"""
import json
import os
import pathlib
import subprocess
import sys
import tempfile
import time

SCRIPT = pathlib.Path(sys.argv[1]).resolve()
assert SCRIPT.is_file(), SCRIPT


def sh(args, cwd, env=None, stdin="{}"):
    return subprocess.run(args, cwd=cwd, env=env, input=stdin,
                          capture_output=True, text=True, timeout=90)


def write_go(bindir, exit_code, counter):
    """Fake `go` that records every invocation, then exits with exit_code.

    The counter lives OUTSIDE the repository: the hook now identifies the work
    tree by content, so an instrument writing inside the tree would change the
    very thing it is measuring and every call would look like a re-gate.
    """
    p = bindir / "go"
    if exit_code:
        tail = 'echo "probe: simulated failure in $*" >&2\nexit %d\n' % exit_code
    else:
        tail = "exit 0\n"
    p.write_text('#!/bin/sh\necho "$*" >> "%s"\n' % counter + tail)
    p.chmod(0o755)


def fixture(exit_code=7):
    box = pathlib.Path(tempfile.mkdtemp(prefix="t601-probe-"))
    root = box / "repo"
    bindir = box / "bin"
    root.mkdir()
    bindir.mkdir()
    counter = box / "invoked"
    (root / "go.mod").write_text("module probe\n\ngo 1.22\n")
    (root / "main.go").write_text("package main\n\nfunc main() {}\n")
    write_go(bindir, exit_code, counter)
    steps = (
        ["git", "init", "-q"],
        ["git", "add", "go.mod", "main.go"],
        ["git", "-c", "user.name=Probe", "-c", "user.email=probe@example.invalid",
         "-c", "core.hooksPath=/dev/null", "commit", "-qm",
         "docs: sync SPEC-PROBE-001 documentation"],
    )
    for args in steps:
        v = sh(args, root, stdin=None)
        if v.returncode:
            raise RuntimeError("%s failed: %s" % (args, v.stderr))
    env = os.environ.copy()
    env.update(PATH="%s:/usr/bin:/bin:/usr/sbin:/sbin" % bindir,
               CLAUDE_PROJECT_DIR=str(root),
               MOAI_SYNC_GATE_BLOCKING="1",
               MOAI_AUTONOMY_TIER="semi-auto")
    return root, env


def invocations(root):
    f = root.parent / "invoked"
    return f.read_text().strip().splitlines() if f.exists() else []


def record(root):
    f = root / ".moai/state/sync-quality-gate.last"
    return f.read_text().strip() if f.exists() else "<absent>"


def head(root):
    return sh(["git", "rev-parse", "HEAD"], root, stdin=None).stdout.strip()


def run_hook(root, env, stdin="{}"):
    before = len(invocations(root))
    v = sh(["bash", str(SCRIPT)], root, env, stdin=stdin)
    return dict(exit=v.returncode, stdout=v.stdout.strip(),
                checks_ran=len(invocations(root)) > before,
                record=record(root))


def seed_running(root, age_seconds):
    st = root / ".moai/state"
    st.mkdir(parents=True, exist_ok=True)
    rec = st / "sync-quality-gate.last"
    rec.write_text("%s running\n" % head(root))
    if age_seconds:
        when = time.time() - age_seconds
        os.utime(rec, (when, when))


results = {"script": str(SCRIPT)}

root, env = fixture(exit_code=0)
results["P0_control_pass"] = dict(first=run_hook(root, env), second=run_hook(root, env))

root, env = fixture(exit_code=7)
first = run_hook(root, env)
second = run_hook(root, env)
third = run_hook(root, env)
results["P1_fail_recall"] = dict(
    first=first, second=second, third=third,
    stdout_identical_1_2=first["stdout"] == second["stdout"],
    stdout_identical_1_3=first["stdout"] == third["stdout"])

root, env = fixture(exit_code=7)
seed_running(root, age_seconds=0)
results["P2a_interrupted_fresh"] = run_hook(root, env)

root, env = fixture(exit_code=7)
seed_running(root, age_seconds=300)
results["P2b_interrupted_stale"] = run_hook(root, env)

root, env = fixture(exit_code=7)
failing = run_hook(root, env)
sha_before = head(root)
write_go(root.parent / "bin", 0, root.parent / "invoked")
(root / "main.go").write_text("package main\n\nfunc main() { _ = 1 }\n")
sha_after = head(root)
after_fix = run_hook(root, env)
results["P3_fixed_same_head"] = dict(
    failing=failing, after_fix=after_fix,
    head_unchanged=sha_before == sha_after,
    stale_block_redelivered=(after_fix["stdout"] == failing["stdout"]
                             and failing["stdout"] != ""))

print(json.dumps(results, ensure_ascii=False, indent=2))
