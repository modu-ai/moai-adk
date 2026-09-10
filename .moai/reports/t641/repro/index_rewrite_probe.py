# t641 probe: does a plain `status` rewrite the index file when stat info is stale,
# and does `--no-optional-locks` leave it untouched? This decides whether a
# behavioural RED test (index bytes/mtime unchanged after Status()) is possible.
import hashlib
import os
import subprocess
import sys
import tempfile
import time

VCS = "git"

def run(args, cwd):
    p = subprocess.run([VCS] + args, cwd=cwd, capture_output=True, text=True, timeout=60)
    if p.returncode != 0:
        raise RuntimeError(f"{args}: {p.stderr}")
    return p.stdout

def index_fingerprint(repo):
    path = os.path.join(repo, ".git", "index")
    data = open(path, "rb").read()
    return hashlib.sha256(data).hexdigest(), os.stat(path).st_mtime_ns

def trial(flag_args, label):
    repo = tempfile.mkdtemp(prefix="t641-probe-")
    run(["init", "-q", "-b", "main"], repo)
    for k, v in (("user.email", "t641@example.invalid"), ("user.name", "t641"), ("commit.gpgsign", "false")):
        run(["config", k, v], repo)
    for i in range(5):
        open(os.path.join(repo, f"f{i}.txt"), "w").write(f"{i}\n")
    run(["add", "-A"], repo)
    run(["commit", "-q", "-m", "init"], repo)
    time.sleep(1.1)  # leave the racy-git window
    run(["status", "--porcelain"], repo)  # settle the index once
    time.sleep(1.1)
    later = time.time()
    for i in range(5):  # stale stat info, same content
        os.utime(os.path.join(repo, f"f{i}.txt"), (later, later))
    before = index_fingerprint(repo)
    out = run(flag_args + ["status", "--porcelain", "--branch"], repo)
    after = index_fingerprint(repo)
    print(f"{label}: index_sha_changed={before[0] != after[0]} index_mtime_changed={before[1] != after[1]} status_output={out.strip()!r}")
    return before != after

if __name__ == "__main__":
    plain = trial([], "plain")
    nol = trial(["--no-optional-locks"], "no-optional-locks")
    print(f"RESULT plain_rewrote_index={plain} nol_rewrote_index={nol}")
    sys.exit(0)
