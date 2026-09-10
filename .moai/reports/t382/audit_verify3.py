import collections
import glob
import json
import os
import re

d = json.load(open("/tmp/t382audit.json", encoding="utf-8"))
f = d.get("drift_findings") or []
era_of = {x["spec_id"]: x.get("era") for x in f if x.get("finding_type") == "EraAutoDetected"}

ids = []
for p in glob.glob(".moai/specs/*/spec.md"):
    t = open(p, encoding="utf-8").read()
    if re.search(r"^created_at:", t, re.M):
        ids.append(os.path.basename(os.path.dirname(p)))
print("created_at SPECs (glob):", len(ids))
c = collections.Counter(era_of.get(i, "<no EraAutoDetected finding: explicit era: or absent>") for i in ids)
for k, v in c.items():
    print("  ", k, v)
print("  V3R5 members:", [i for i in ids if era_of.get(i) == "V3R5"])
# which of the 46 have an explicit era: field
noauto = [i for i in ids if i not in era_of]
print("  no auto-detect finding:", len(noauto), noauto[:10])
