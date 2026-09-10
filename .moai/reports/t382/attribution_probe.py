"""t382 AC-EH3-005 attribution probe — per-SPEC era delta between two audit JSON runs.

Usage: attribution_probe.py <before.json> <after.json>

Prints the era-bucket totals for both runs, then every SPEC whose era changed,
with the modern-era signal that justifies the change recomputed from its
frontmatter. Exit 1 when any SPEC changed era WITHOUT a justifying signal, or
when an EraUnclassified finding appears in the after run.
"""
import collections
import json
import re
import sys


def eras(path):
    d = json.load(open(path, encoding="utf-8"))
    out = {}
    unclassified = []
    for x in (d.get("drift_findings") or []):
        if x.get("finding_type") == "EraAutoDetected":
            out[x["spec_id"]] = x.get("era")
        if x.get("finding_type") == "EraUnclassified":
            unclassified.append(x["spec_id"])
    return d, out, unclassified


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


def signals(spec_id):
    fm = frontmatter(".moai/specs/%s/spec.md" % spec_id)
    phase = fm.get("phase", "")
    created = fm.get("created", "")
    p = phase.lower().strip()
    got = []
    if "v3r6" in p or p.startswith("v3.0"):
        got.append("phase=%s" % phase)
    if created and created >= "2026-04-01":
        got.append("created=%s" % created)
    return got


db, ea, ub = eras(sys.argv[1])
da, eb, ua = eras(sys.argv[2])

for label, d, e, u in (("BEFORE", db, ea, ub), ("AFTER", da, eb, ua)):
    c = collections.Counter(e.values())
    print("%s  total=%s grandfathered=%s modern_era_clean=%s  buckets=%s  EraUnclassified=%d"
          % (label, d.get("total_specs"), d.get("grandfathered"), d.get("modern_era_clean"),
             dict(sorted(c.items())), len(u)))

changed = sorted(k for k in ea if k in eb and ea[k] != eb[k])
print("\nSPECs whose era changed: %d" % len(changed))
bad = []
for k in changed:
    s = signals(k)
    if not s:
        bad.append(k)
    print("  %-56s %-8s -> %-8s  justification: %s"
          % (k, ea[k], eb[k], ", ".join(s) or "*** NONE ***"))

only_before = sorted(set(ea) - set(eb))
only_after = sorted(set(eb) - set(ea))
print("\nIDs only in before: %d %s" % (len(only_before), only_before))
print("IDs only in after:  %d %s" % (len(only_after), only_after))

fail = False
if bad:
    print("\nFAIL: %d SPEC(s) changed era with no justifying signal: %s" % (len(bad), bad))
    fail = True
if ua:
    print("\nFAIL: EraUnclassified findings appeared after the change: %s" % ua)
    fail = True
if not fail:
    print("\nOK: every era change carries a modern-era signal; EraUnclassified stayed 0.")
sys.exit(1 if fail else 0)
