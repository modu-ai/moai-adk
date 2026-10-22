"""SPEC-FACTORY-BOOT-MINIMAL-001 AGENTS.md check (card t1923, AC-FBM-004).

Read-only. Verifies REQ-FBM-010 on one AGENTS.md file:

  1. size  — the file is at or under 24,276 bytes (300 bytes of headroom under
     the Codex contract ceiling 24,576);
  2. scope — everything before the "No main-variant" paragraph and everything
     from the "Three traceability carriers" paragraph on is byte-identical to
     AGENTS.md at the base tree, so no other obligation can be deleted to buy
     bytes;
  3. obligations — inside the two reworded paragraphs, every obligation of the
     base wording survives (each obligation accepts the base phrasing or the
     operator patch's phrasing; whitespace is normalized first).

Prints one line per check and a final "result PASS" / "result FAIL"; exits 0
only on PASS.

Usage: python3 -I agents_md_check.py [AGENTS.md path] [base sha]
"""
import re
import subprocess
import sys

CEILING = 24276
BASE = "59959629908cda723c37e26c7c15718194d6b38a"
START = b"**No main-variant"
END = b"**Three traceability carriers"

# (id, obligation, accepted phrasings) — any phrasing present satisfies it.
OBLIGATIONS = [
    ("O1", "no WT-*main* branch", ["`WT-*main*`"]),
    ("O2", "no worktree holding main for integration", ["holds `main`", "holding `main`"]),
    ("O3", "main advances only by PR merges", ["advances only through PR merges", "only by PR merges"]),
    ("O4", "local main re-synced with origin/main at every landing",
     ["re-synced with `origin/main` at every landing", "with `origin/main` after every landing"]),
    ("O5", "local main ahead past a landing is a defect", ["is a defect", "do not leave it ahead"]),
    ("O6", "tooling never detaches the primary checkout's HEAD",
     ["never detach the primary checkout's HEAD", "never detach its HEAD"]),
    ("O7", "a branch occupied by the primary is refused",
     ["the tool refuses rather than steals", "refuse a branch occupied by the primary checkout"]),
    ("O8", "WT worktrees live ONLY under .moai/worktrees/", ["live only under `.moai/worktrees/`"]),
    ("O9", "created by moai worktree new", ["created by `moai worktree new`"]),
    ("O10", ".claude/worktrees/ reserved for runtime agent-* trees", ["`.claude/worktrees/`"]),
    ("O11", "agent-* runtime isolation trees named", ["`agent-*`"]),
    ("O12", "adopted runtime tree renamed to WT- in place",
     ["renamed to `WT-` in place", "branch to `WT-` in place"]),
    ("O13", "adopted runtime tree stays at its creator's path",
     ["stays where its creator put it", "keep the tree at its creator's path"]),
]


def split(data):
    s = data.find(START)
    e = data.find(END, s) if s >= 0 else -1
    if s < 0 or e < 0:
        return None
    return data[:s], data[s:e], data[e:]


def main():
    path = sys.argv[1] if len(sys.argv) > 1 else "AGENTS.md"
    base = sys.argv[2] if len(sys.argv) > 2 else BASE
    head = open(path, "rb").read()
    base_data = subprocess.run(["git", "show", base + ":AGENTS.md"], capture_output=True, check=True).stdout
    ok = True
    size_ok = len(head) <= CEILING
    ok &= size_ok
    print("size %d ceiling %d %s" % (len(head), CEILING, "OK" if size_ok else "OVER"))
    hs, bs = split(head), split(base_data)
    if hs is None or bs is None:
        print("scope MISSING-ANCHOR (paragraph anchors not found)")
        print("result FAIL")
        return 1
    pre_ok, suf_ok = hs[0] == bs[0], hs[2] == bs[2]
    ok &= pre_ok and suf_ok
    print("scope prefix %s suffix %s" % ("identical" if pre_ok else "CHANGED", "identical" if suf_ok else "CHANGED"))
    block = re.sub(r"\s+", " ", hs[1].decode("utf-8"))
    for oid, what, phrasings in OBLIGATIONS:
        hit = [p for p in phrasings if p in block]
        ok &= bool(hit)
        print("%s %s %s" % (oid, "found" if hit else "MISSING", what))
    print("result %s" % ("PASS" if ok else "FAIL"))
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
