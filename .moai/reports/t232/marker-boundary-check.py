"""Check the proposed IsRetiredClause boundary against the registry's real clauses.

Verifies the tightened rule (prefix + boundary char + closing bracket) keeps every
genuine retirement marker classified as retired, and still rejects the live clause
that merely MENTIONS the marker mid-text.
"""
import re
import sys

PREFIX = "[SUPERSEDED"


def is_retired_current(clause):
    return clause.strip().startswith(PREFIX)


def is_retired_proposed(clause):
    t = clause.strip()
    if not t.startswith(PREFIX):
        return False
    rest = t[len(PREFIX):]
    if rest == "" or rest[0] not in (']', ' '):
        return False
    return ']' in rest


reg = sys.argv[1]
clauses = []
for line in open(reg):
    m = re.match(r'^  clause: "(.*)"$', line.rstrip('\n'))
    if m:
        clauses.append(m.group(1))

changed = [c for c in clauses if is_retired_current(c) != is_retired_proposed(c)]
retired_now = [c for c in clauses if is_retired_current(c)]
retired_new = [c for c in clauses if is_retired_proposed(c)]

print("clauses=%d retired_current=%d retired_proposed=%d changed=%d"
      % (len(clauses), len(retired_now), len(retired_new), len(changed)))
for c in changed:
    print("  CHANGED: %r" % c[:100])

print("--- near-miss inputs the finding named ---")
for probe in ("[SUPERSEDEDLY] text", "[SUPERSEDED text", "[SUPERSEDED] text",
              "[SUPERSEDED by X] text", "[superseded] text",
              "mark superseded entries with [SUPERSEDED by <file>] prefix", ""):
    print("  current=%-5s proposed=%-5s  %r"
          % (is_retired_current(probe), is_retired_proposed(probe), probe))
