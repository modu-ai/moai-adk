#!/usr/bin/env python3
"""Card t482 — reconcile this card's S1 figure (14) with the lead's independent
measurement (73 / 9 / 10).

The two numbers were produced by different queries, so "boundary condition"
is a hypothesis until the queries are stated side by side and their populations
named. A verdict whose supporting figure is disputed is not closed just because
the verdict happens to be right.

Input : subject stream on stdin.
"""
import re
import sys

subs = [line.rstrip("\n") for line in sys.stdin]
TOK = re.compile(r"\bt[0-9]+\b")

# This card's S1, as classified in verdict.md section 4: a merge whose scope
# names a card worktree, in BOTH spellings, integrating into a release branch.
S1 = re.compile(r"^merge\((?:WT|worktree)-(t[0-9]+)\): integrate into release/")
# The lead's population: every `merge(WT-...)` scope, either spelling excluded.
LEAD_SCOPE = re.compile(r"^merge\(WT-[^)]*\):")
# The same, both spellings.
BOTH_SCOPE = re.compile(r"^merge\((?:WT|worktree)-[^)]*\):")

s1_ids, s1_subj = set(), []
for s in subs:
    m = S1.search(s)
    if m:
        s1_ids.add(m.group(1))
        s1_subj.append(s)

lead_scope = [s for s in subs if LEAD_SCOPE.search(s)]
both_scope = [s for s in subs if BOTH_SCOPE.search(s)]

def split(pop, label):
    no_tok, scope_only, other = [], [], []
    for s in pop:
        scope = re.match(r"^merge\(([^)]*)\):", s).group(1)
        in_scope = set(TOK.findall(scope))
        rest = s[s.index("):") + 2:]
        in_rest = set(TOK.findall(rest))
        if not in_scope and not in_rest:
            no_tok.append(s)
        elif in_scope and not in_rest:
            scope_only.append(s)
        else:
            other.append(s)
    print(f"{label}: total {len(pop)} | no card token anywhere {len(no_tok)} "
          f"| token only in scope {len(scope_only)} | token elsewhere too {len(other)}")
    return no_tok, scope_only, other

print("this card's S1 (release-integrate, both spellings):",
      len(s1_subj), "subjects,", len(s1_ids), "distinct ids")
print("  spellings:",
      sum(1 for s in s1_subj if s.startswith("merge(WT-")), "WT- /",
      sum(1 for s in s1_subj if s.startswith("merge(worktree-")), "worktree-")
print()
split(lead_scope, "lead's population  merge(WT-...)      ")
split(both_scope, "both spellings     merge(WT|worktree-)")
print()
print("S1 ids:", " ".join(sorted(s1_ids, key=lambda x: int(x[1:]))))
