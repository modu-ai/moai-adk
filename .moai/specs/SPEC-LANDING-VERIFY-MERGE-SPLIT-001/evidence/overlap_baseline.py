#!/usr/bin/env python3
"""Baseline: how often would a landing's moved set miss the next card's verification scope?

Pair model: card X = landing i (the waiting card), moved commits = landing i-1's
change on the integration branch (first-parent diff). Closure is computed once
per landing at the pinned analysis tree (the current worktree HEAD) -- an
approximation recorded as a Gap in the output.
"""
import json, subprocess, sys, os

ROOT = sys.argv[1]
TIP = sys.argv[2]
N = int(sys.argv[3])

def run(args, check=True):
    p = subprocess.run(args, cwd=ROOT, capture_output=True, text=True)
    if check and p.returncode != 0:
        return None
    return p.stdout

merges = run(["git", "log", "--first-parent", "--grep=integration merge", "--format=%H", f"-{N+1}", TIP]).split()
# merges[0] is newest. pair (waiting card = merges[k+1], moved landing = merges[k])
rows = []
for k in range(len(merges) - 1):
    mv, card = merges[k], merges[k + 1]
    moved = set(run(["git", "diff", "--name-only", f"{mv}^1", mv]).split())
    # card's own change: merge-base of the two parents .. second parent
    mb = run(["git", "merge-base", f"{card}^1", f"{card}^2"], check=False)
    if not mb:
        continue
    mb = mb.strip()
    cardfiles = set(run(["git", "diff", "--name-only", mb, f"{card}^2"]).split())
    gofiles = [f for f in cardfiles if f.endswith(".go")]
    pkgdirs = sorted({os.path.dirname(f) or "." for f in gofiles if os.path.isdir(os.path.join(ROOT, os.path.dirname(f) or "."))})
    closure = set()
    if pkgdirs:
        args = ["go", "list", "-deps", "-test", "-f", "{{.Dir}}"] + ["./" + d for d in pkgdirs]
        out = run(args, check=False)
        if out is None:
            rows.append(dict(card=card[:10], moved=mv[:10], status="undecidable-golist", moved_n=len(moved), card_n=len(cardfiles)))
            continue
        for line in out.split():
            if line.startswith(ROOT + "/") or line == ROOT:
                closure.add(os.path.relpath(line, ROOT))
    else:
        rows.append(dict(card=card[:10], moved=mv[:10], status="no-go-packages(undecidable)", moved_n=len(moved), card_n=len(cardfiles)))
        continue
    def under_closure(f):
        for d in closure:
            if d == ".":
                if "/" not in f:
                    return True
            elif f == d or f.startswith(d + "/"):
                return True
        return False
    hit_closure = sorted(f for f in moved if under_closure(f))
    hit_card = sorted(f for f in moved if f in cardfiles)
    modfiles = sorted(f for f in moved if f in ("go.mod", "go.sum", "go.work", "go.work.sum"))
    nongo_outside = sorted(f for f in moved if not f.endswith(".go") and not under_closure(f))
    overlap_lenient = bool(hit_closure or hit_card or modfiles)
    overlap_strict = bool(overlap_lenient or nongo_outside)
    rows.append(dict(card=card[:10], moved=mv[:10], status="decided", moved_n=len(moved), card_n=len(cardfiles),
                     closure_pkgs=len(closure), lenient_overlap=overlap_lenient, strict_overlap=overlap_strict,
                     hit_closure_n=len(hit_closure), nongo_outside_n=len(nongo_outside)))

decided = [r for r in rows if r["status"] == "decided"]
print(json.dumps(dict(pairs=len(rows), decided=len(decided),
    lenient_reusable=sum(1 for r in decided if not r["lenient_overlap"]),
    strict_reusable=sum(1 for r in decided if not r["strict_overlap"]),
    undecidable=len(rows) - len(decided)), indent=1))
for r in rows:
    print(json.dumps(r))
