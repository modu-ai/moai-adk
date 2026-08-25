"""Reconcile the analyzer's clause-fail set against the validator's DRIFT set."""
import json
import re
import sys

res = json.load(open(sys.argv[1]))
v = open(sys.argv[2]).read()
ids_v = set(re.findall(r'(CONST-\S+) @', v))
ids_a = {r['id'] for r in res if not r['clause_ok']}
print("validator_drift=%d analyzer_clause_fail=%d" % (len(ids_v), len(ids_a)))
print("analyzer_only=%s" % sorted(ids_a - ids_v))
print("validator_only=%s" % sorted(ids_v - ids_a))
by_id = {r['id']: r for r in res}
for i in sorted(ids_a - ids_v) + sorted(ids_v - ids_a):
    print("  %s" % by_id.get(i))
