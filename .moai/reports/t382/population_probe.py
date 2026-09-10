"""t382 AC-EH3-005 population probe — per-SPEC era attribution for the V3R5 bucket.

Reads `moai spec audit --json` from the path given as argv[1] and prints, for every
V3R5-classified SPEC, which modern-era signal (if any) it carries. Used to attribute
the era change SPEC by SPEC rather than as a total.
"""
import json
import re
import sys

d = json.load(open(sys.argv[1], encoding="utf-8"))
ids = sorted(x["spec_id"] for x in (d.get("drift_findings") or [])
             if x.get("finding_type") == "EraAutoDetected" and x.get("era") == "V3R5")


def frontmatter(path):
    try:
        text = open(path, encoding="utf-8").read()
    except OSError:
        return {}
    m = re.match(r"^---\n(.*?)\n---", text, re.S)
    if not m:
        return {}
    out = {}
    for line in m.group(1).split("\n"):
        if ":" in line and not line.startswith(" "):
            k, v = line.split(":", 1)
            out[k.strip()] = v.strip().strip('"')
    return out


print("V3R5 population: %d" % len(ids))
for i in ids:
    fm = frontmatter(".moai/specs/%s/spec.md" % i)
    phase = fm.get("phase", "")
    created = fm.get("created", "")
    sig = []
    p = phase.lower()
    if "v3r6" in p or p.startswith("v3.0"):
        sig.append("phase")
    if created and created >= "2026-04-01":
        sig.append("created")
    print("%-56s signals=%-15s phase=%-22r created=%r"
          % (i, ",".join(sig) or "NONE", phase, created))
