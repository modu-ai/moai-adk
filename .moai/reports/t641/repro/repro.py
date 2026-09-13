# t641 reproduction: does a concurrent `git status --porcelain --branch` loop make
# index writers (git add / git commit) fail on index.lock, and does
# `--no-optional-locks` on the status loop remove that?
# Bounded: every subprocess has a timeout, the status thread is joined in finally,
# and the whole script is wrapped by an outer `timeout` by the caller.
import os, sys, subprocess, threading, time, json, shutil

BASE = os.path.dirname(os.path.abspath(__file__))
N_WRITES = int(sys.argv[1]) if len(sys.argv) > 1 else 150
N_FILES = 3000
TOUCH_PER_ROUND = 300

def run(args, cwd, check=True, timeout=60):
    p = subprocess.run(args, cwd=cwd, capture_output=True, text=True, timeout=timeout)
    if check and p.returncode != 0:
        raise RuntimeError(f"{args} -> {p.returncode}: {p.stderr.strip()}")
    return p

def make_repo(path):
    if os.path.exists(path):
        shutil.rmtree(path)
    os.makedirs(path)
    run(["git", "init", "-q", "-b", "main"], path)
    run(["git", "config", "user.email", "t641@example.invalid"], path)
    run(["git", "config", "user.name", "t641"], path)
    run(["git", "config", "commit.gpgsign", "false"], path)
    for i in range(N_FILES):
        d = os.path.join(path, f"d{i % 30}")
        os.makedirs(d, exist_ok=True)
        with open(os.path.join(d, f"f{i}.txt"), "w") as f:
            f.write(f"line {i}\n")
    run(["git", "add", "-A"], path, timeout=120)
    run(["git", "commit", "-q", "-m", "init"], path, timeout=120)

def trial(mode):
    repo = os.path.join(BASE, f"repo-{mode}")
    make_repo(repo)
    status_cmd = ["git"] + (["--no-optional-locks"] if mode == "no-optional-locks" else []) + ["status", "--porcelain", "--branch"]
    stop = threading.Event()
    stats = {"status_runs": 0, "status_failures": 0, "status_lock_failures": 0}

    def status_loop():
        while not stop.is_set():
            p = subprocess.run(status_cmd, cwd=repo, capture_output=True, text=True, timeout=30)
            stats["status_runs"] += 1
            if p.returncode != 0:
                stats["status_failures"] += 1
                if "index.lock" in p.stderr:
                    stats["status_lock_failures"] += 1

    add_fail = commit_fail = add_lock = commit_lock = 0
    samples = []
    t = threading.Thread(target=status_loop)
    t.start()
    try:
        for i in range(N_WRITES):
            # make the index stat cache stale so a plain status wants to refresh it
            now = time.time()
            for j in range(TOUCH_PER_ROUND):
                k = (i * TOUCH_PER_ROUND + j) % N_FILES
                os.utime(os.path.join(repo, f"d{k % 30}", f"f{k}.txt"), (now, now))
            target = os.path.join(repo, "d0", "f0.txt")
            with open(target, "a") as f:
                f.write(f"w{i}\n")
            a = run(["git", "add", "d0/f0.txt"], repo, check=False)
            if a.returncode != 0:
                add_fail += 1
                if "index.lock" in a.stderr:
                    add_lock += 1
                if len(samples) < 3:
                    samples.append(("add", a.returncode, a.stderr.strip()[:200]))
                continue
            c = run(["git", "commit", "-q", "-m", f"w{i}"], repo, check=False)
            if c.returncode != 0:
                commit_fail += 1
                if "index.lock" in c.stderr:
                    commit_lock += 1
                if len(samples) < 3:
                    samples.append(("commit", c.returncode, c.stderr.strip()[:200]))
    finally:
        stop.set()
        t.join(timeout=60)
    return {
        "mode": mode, "writes_attempted": N_WRITES,
        "add_failures": add_fail, "add_lock_failures": add_lock,
        "commit_failures": commit_fail, "commit_lock_failures": commit_lock,
        **stats, "status_thread_alive_after_join": t.is_alive(), "failure_samples": samples,
    }

def accuracy_control():
    root = os.path.join(BASE, "accuracy")
    if os.path.exists(root):
        shutil.rmtree(root)
    os.makedirs(root)
    remote = os.path.join(root, "remote.bare")
    run(["git", "init", "-q", "--bare", "-b", "main", remote], root)
    a = os.path.join(root, "a"); b = os.path.join(root, "b")
    for p in (a, b):
        run(["git", "clone", "-q", remote, p], root)
        run(["git", "config", "user.email", "t641@example.invalid"], p)
        run(["git", "config", "user.name", "t641"], p)
        run(["git", "config", "commit.gpgsign", "false"], p)
    for name in ("m.txt", "s.txt"):
        open(os.path.join(a, name), "w").write("base\n")
    run(["git", "add", "-A"], a); run(["git", "commit", "-q", "-m", "base"], a); run(["git", "push", "-q", "origin", "main"], a)
    run(["git", "pull", "-q", "origin", "main"], b)
    open(os.path.join(b, "r.txt"), "w").write("remote\n")
    run(["git", "add", "r.txt"], b); run(["git", "commit", "-q", "-m", "remote-ahead"], b); run(["git", "push", "-q", "origin", "main"], b)
    run(["git", "branch", "-q", "--set-upstream-to=origin/main", "main"], a)
    open(os.path.join(a, "l.txt"), "w").write("local\n")
    run(["git", "add", "l.txt"], a); run(["git", "commit", "-q", "-m", "local-ahead"], a)
    run(["git", "fetch", "-q", "origin"], a)
    open(os.path.join(a, "m.txt"), "a").write("modified\n")
    open(os.path.join(a, "s.txt"), "a").write("staged\n"); run(["git", "add", "s.txt"], a)
    open(os.path.join(a, "u.txt"), "w").write("untracked\n")
    plain = run(["git", "status", "--porcelain", "--branch"], a).stdout
    nol = run(["git", "--no-optional-locks", "status", "--porcelain", "--branch"], a).stdout
    return {"plain": plain, "no_optional_locks": nol, "byte_equal": plain == nol}

if __name__ == "__main__":
    which = sys.argv[2] if len(sys.argv) > 2 else "all"
    out = {}
    if which in ("all", "plain"):
        out["plain"] = trial("plain")
    if which in ("all", "nol"):
        out["no-optional-locks"] = trial("no-optional-locks")
    if which in ("all", "accuracy"):
        out["accuracy"] = accuracy_control()
    print(json.dumps(out, ensure_ascii=False, indent=2))
