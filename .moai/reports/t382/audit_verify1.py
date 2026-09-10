import collections
import json
import os
import re

d = json.load(open("/tmp/t382audit.json", encoding="utf-8"))
f = d.get("drift_findings") or []
print("finding_type counts:", dict(collections.Counter(x.get("finding_type") for x in f)))
era = collections.Counter(x.get("era") for x in f if x.get("finding_type") == "EraAutoDetected")
print("EraAutoDetected era counts:", dict(era))
v3r5 = [x["spec_id"] for x in f if x.get("finding_type") == "EraAutoDetected" and x.get("era") == "V3R5"]
print("V3R5 count:", len(v3r5))

pre = []
post = []
none = []
e4 = 0
e3 = 0
phasehit = []
for i in v3r5:
    t = open(".moai/specs/%s/spec.md" % i, encoding="utf-8").read()
    m = re.match(r"^---\n(.*?)\n---", t, re.S)
    fm = {}
    if m:
        for line in m.group(1).split("\n"):
            if ":" in line and not line.startswith(" "):
                k, v = line.split(":", 1)
                fm[k.strip()] = v.strip().strip('"')
    c = fm.get("created", "")
    (none if not c else (post if c >= "2026-04-01" else pre)).append((i, c))
    ph = fm.get("phase", "").lower()
    if "v3r6" in ph or ph.startswith("v3.0"):
        phasehit.append(i)
    pg = ".moai/specs/%s/progress.md" % i
    if os.path.exists(pg):
        pt = open(pg, encoding="utf-8").read()
        if "§E.4" in pt:
            e4 += 1
        if "§E.3" in pt:
            e3 += 1
print("created >= threshold:", len(post), " < threshold:", len(pre), " missing:", len(none), none)
print("pre-threshold list:", pre)
print("matchesModernPhase true:", len(phasehit), phasehit)
print("has E.4:", e4, " has E.3:", e3)

term = {"completed", "superseded", "archived", "rejected"}
nonterm = []
for i in v3r5:
    t = open(".moai/specs/%s/spec.md" % i, encoding="utf-8").read()
    m = re.search(r"^status:\s*(\S+)", t, re.M)
    s = m.group(1).strip().strip('"') if m else "?"
    if s not in term:
        nonterm.append((i, s))
print("non-terminal:", len(nonterm))
print([x[0] for x in nonterm])
