"""Does the retirement skip introduced by #1611 collide with AC-ZRR-002/003?

AC-ZRR-002/003 require all 101 entries to hit exactly once in their own file.
#1611 makes `validate` skip retired entries, on the stated ground that a retired
clause cites text that is gone by definition, so a verbatim check can only fail.

If the four retired entries are among the clause failures, the two contracts
disagree and the SPEC has to say which one governs.
"""
import json

res = json.load(open('.moai/reports/t232/analysis-postmerge.json'))
retired = [r for r in res if r['clause'].lstrip().startswith('[SUPERSEDED')]
print("retired entries: %d" % len(retired))
for r in retired:
    print("  %-18s clause_ok=%-5s anchor_ok=%-5s  %s"
          % (r['id'], r['clause_ok'], r['anchor_ok'], r['file']))

fail = [r for r in res if not r['clause_ok']]
retired_failing = [r for r in retired if not r['clause_ok']]
print("\nclause failures total: %d" % len(fail))
print("of which retired:      %d" % len(retired_failing))
print("non-retired failures:  %d" % (len(fail) - len(retired_failing)))
