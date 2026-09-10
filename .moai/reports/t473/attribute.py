#!/usr/bin/env python3
"""Per-file attribution of the always-loaded token surface at a given git rev.

Mirrors internal/config/token_budget_guard.go exactly:
  - surface = every .claude/rules/moai/**/*.md whose frontmatter carries no
    `paths:` key (sorted), then 3 fixed slots: CLAUDE.md, AGENTS.md,
    .claude/output-styles/moai/moai.md
  - tokens per file = len(bytes) // 4   (estimateTokens)
  - total = sum of per-file tokens

Cross-checked against `go test -run TestAlwaysLoadedTokenBudget -v` on the same
tree; a divergence means this script is wrong, not the guard.

Usage: attribute.py <rev>
"""
import subprocess, sys

rev = sys.argv[1]

def git(*args):
    return subprocess.run(["git", *args], capture_output=True, check=True).stdout

def blob(path):
    try:
        return subprocess.run(["git", "show", f"{rev}:{path}"],
                              capture_output=True, check=True).stdout
    except subprocess.CalledProcessError:
        return None  # hermetic: absent file -> 0 tokens

def has_paths(data):
    """frontmatterHasPaths port."""
    lines = data.split(b"\n")
    if not lines:
        return False
    if lines[0].rstrip(b" \t\r") != b"---":
        return False
    for line in lines[1:]:
        if line.rstrip(b" \t\r") == b"---":
            return False
        if line.startswith(b"paths:"):
            return True
    return False

listing = git("ls-tree", "-r", "--name-only", rev, ".claude/rules/moai/").decode().split("\n")
rules = sorted(p for p in listing if p.endswith(".md"))

surface = []
for p in rules:
    data = blob(p)
    if data is not None and not has_paths(data):
        surface.append(p)
surface += ["CLAUDE.md", "AGENTS.md", ".claude/output-styles/moai/moai.md"]

total = 0
rows = []
for p in surface:
    data = blob(p)
    n = 0 if data is None else len(data) // 4
    total += n
    rows.append((n, p))

for n, p in sorted(rows, reverse=True):
    print(f"{n:7d}  {p}")
print(f"{'-'*60}")
print(f"{total:7d}  TOTAL  ({len(surface)} entries)  rev={rev}")
