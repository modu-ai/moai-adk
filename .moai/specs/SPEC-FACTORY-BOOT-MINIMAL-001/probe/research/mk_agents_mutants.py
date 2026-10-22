"""Builds the AC-FBM-004 mutant copies of AGENTS.md (card t1923, plan phase).

Read-only on the source; writes five copies into <out-dir>.
Usage: python3 -I mk_agents_mutants.py <AGENTS.md> <out-dir>
"""
import os
import sys

src, out = sys.argv[1], sys.argv[2]
data = open(src, "rb").read()
S = data.index(b"**No main-variant")
E = data.index(b"**Three traceability carriers", S)
NEW = (b"**No main-variant worktrees; sync local `main`.** Never create `WT-*main*` branches\n"
       b"or worktrees holding `main` for integration. Advance `main` only by PR merges and\n"
       b"sync local `main` with `origin/main` after every landing; do not leave it ahead.\n"
       b"Tooling must refuse a branch occupied by the primary checkout and must never\n"
       b"detach its HEAD to take that branch.\n\n"
       b"**WT worktrees live only under `.moai/worktrees/`**, created by `moai worktree new`.\n"
       b"Reserve `.claude/worktrees/` for Claude's ephemeral `agent-*` runtime isolation.\n"
       b"If adopted for card work, rename the runtime tree's branch to `WT-` in place and\n"
       b"keep the tree at its creator's path.\n\n")
cand = data[:S] + NEW + data[E:]
copies = {
    "cand.md": cand,
    "cand_no_only.md": cand.replace(b"live only under", b"belong under", 1),
    "del_mainvariant.md": data[:S] + data[data.index(b"**WT worktrees live only under"):],
    "cand_del_other.md": cand[:cand.index(b"**Start a new card in a new worktree.**")] + cand[cand.index(b"**Card worktree branches carry"):],
}
cs = cand.index(b"**No main-variant")
copies["cand_hollow.md"] = cand[:cs] + b"**No main-variant worktrees; sync local `main`.**\n\n" + cand[cand.index(b"**Three traceability carriers", cs):]
for name, body in copies.items():
    open(os.path.join(out, name), "wb").write(body)
    print(name, len(body))
