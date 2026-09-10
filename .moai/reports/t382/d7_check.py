"""t382 D7 check — era distribution of the SPECs that write `created_at:`.

Establishes whether a created_at-tolerant era engine could change their classification:
a SPEC caught by H-1 (no progress.md) or H-2 (no §E.* markers) returns before any date
heuristic runs, so reading created_at could not move it.

argv[1] = path to the audit JSON, argv[2] = path to a newline-delimited SPEC-ID list.
"""
import json
import sys
from collections import Counter

with open(sys.argv[1], encoding="utf-8") as fh:
    audit = json.load(fh)
era_of = {x["spec_id"]: x.get("era")
          for x in (audit.get("drift_findings") or [])
          if x.get("finding_type") == "EraAutoDetected"}

with open(sys.argv[2], encoding="utf-8") as fh:
    ids = [line.strip() for line in fh if line.strip()]

dist = Counter(era_of.get(i, "<not-in-audit>") for i in ids)
print("created_at population: %d" % len(ids))
for era, n in sorted(dist.items()):
    print("  %-16s %d" % (era, n))
print("reached by a date heuristic (V3R5 bucket only): %d" % dist.get("V3R5", 0))
