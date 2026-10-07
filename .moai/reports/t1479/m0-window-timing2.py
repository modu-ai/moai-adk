#!/usr/bin/env python3
# M0 — in-window duration measurement (card t1479, plan.md M0), take 2.
# Times ONLY the three in-window git calls with perf_counter, no interpreter
# startup inside the measured span: tree-identity check + git merge --no-ff +
# merged-tree comparison. N = 20 on a fixture repository.
import subprocess, time, statistics, os, shutil

fixture = "/tmp/t1479-m0/fixture"
if os.path.exists(fixture):
    shutil.rmtree(fixture)
os.makedirs(fixture)

def git(*args, cwd=fixture):
    return subprocess.run(["git", *args], cwd=cwd, capture_output=True, text=True)

git("init", "-q", "-b", "main")
git("config", "user.email", "t@t.local"); git("config", "user.name", "t")
open(f"{fixture}/base.txt", "w").write("base")
git("add", "base.txt"); git("commit", "-q", "-m", "base")
git("checkout", "-q", "-b", "cand")
for i in range(1, 6):
    open(f"{fixture}/card-{i}.txt", "w").write(f"card {i}")
    git("add", "."); git("commit", "-q", "-m", f"card {i}")
git("checkout", "-q", "main")

timings = []
for n in range(1, 21):
    t0 = time.perf_counter_ns()
    base_tree = git("rev-parse", "HEAD^{tree}").stdout.strip()
    cand_sha = git("rev-parse", "cand").stdout.strip()
    cand_tree = git("rev-parse", f"{cand_sha}^{{tree}}").stdout.strip()
    if base_tree != cand_tree:
        git("merge", "--no-ff", "-q", "-m", f"run {n} merge", "cand")
        merged_tree = git("rev-parse", "HEAD^{tree}").stdout.strip()
        assert merged_tree == cand_tree, "tree mismatch"
    t1 = time.perf_counter_ns()
    timings.append((t1 - t0) // 1000)
    git("reset", "-q", "--hard", "HEAD~1")

print(f"n={len(timings)} median={statistics.median(timings)}us max={max(timings)}us min={min(timings)}us")
print("raw(us):", timings)
