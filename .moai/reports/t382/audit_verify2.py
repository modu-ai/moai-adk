import collections
import json

d = json.load(open("/tmp/t382audit.json", encoding="utf-8"))
f = d.get("drift_findings") or []
v3r5 = set(x["spec_id"] for x in f if x.get("finding_type") == "EraAutoDetected" and x.get("era") == "V3R5")

rows = {}
for line in open("/tmp/t382drift.txt", encoding="utf-8"):
    for sid in v3r5:
        if sid in line:
            rows.setdefault(sid, line.rstrip())
print("V3R5 count:", len(v3r5), "drift rows matched:", len(rows))
import collections
kinds = collections.Counter()
for sid, line in rows.items():
    k = "era-exempt" if "era-exempt" in line else ("terminal-exempt" if "terminal-exempt" in line else "OTHER")
    kinds[k] += 1
    if k != "era-exempt":
        print(" ", k, "->", line)
print(dict(kinds))

# baseline file comparison
before = [line.rstrip() for line in open(".moai/reports/t382/drift-before-9328a5242.txt", encoding="utf-8")]
print("--- drift-before sample ---")
for line in before[:3]:
    print(line)
pop = [line.strip() for line in open(".moai/reports/t382/v3r5-population.txt", encoding="utf-8") if line.strip()]
print("population file entries:", len(pop))
print("population minus current V3R5:", sorted(set(p.split()[0] for p in pop) - v3r5))
print("current V3R5 minus population:", sorted(v3r5 - set(p.split()[0] for p in pop)))
