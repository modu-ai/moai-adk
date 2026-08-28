#!/usr/bin/env python3
"""iter-2 audit scratch: two faithful counters.

  mode=adj   -- spec.md v0.3.0 §3.1/§3.2 adjacency semantics (identifier unit)
  mode=line  -- spec.md v0.2.0 line semantics (the D1 regression)

Prints "COUNT <n>" and exit 0, or "AMBIGUOUS <id>..." and exit 3.
"""
import re, sys

ID = re.compile(r'AC-(?:[A-Z0-9]+-)*[0-9]+')
TOK = re.compile(r'[ \t]*(?:\[RETIRED\]|\[REF\])')      # same-line adjacency only
TOK_ANY = re.compile(r'\[RETIRED\]|\[REF\]')

def main(path, mode):
    occ = {}          # id -> [marked bool, ...]
    for line in open(path, encoding='utf-8'):
        line_marked = bool(TOK_ANY.search(line))
        for m in ID.finditer(line):
            ident = m.group(0)
            if mode == 'line':
                marked = line_marked
            else:
                marked = bool(TOK.match(line, m.end()))
            occ.setdefault(ident, []).append(marked)

    live, excluded, ambiguous = [], [], []
    for ident, marks in sorted(occ.items()):
        if not any(marks):
            live.append(ident)
        elif all(marks):
            excluded.append(ident)
        else:
            ambiguous.append(ident)

    if ambiguous:
        print("AMBIGUOUS " + " ".join(ambiguous))
        sys.exit(3)
    print("COUNT %d   (live=%d excluded=%d ambiguous=0)" % (len(live), len(live), len(excluded)))

main(sys.argv[1], sys.argv[2])
