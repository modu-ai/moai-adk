"""t382 RED probe — counts V3R5-classified SPECs that H-5 would call modern.

Reads `moai spec audit --json` on stdin. Exit 1 when any V3R5 SPEC carries a
modern-era signal (the defect present), exit 0 when none does (the fixed state).
"""
import json
import re
import sys

d = json.load(sys.stdin)
ids = [x["spec_id"] for x in (d.get("drift_findings") or [])
       if x.get("finding_type") == "EraAutoDetected" and x.get("era") == "V3R5"]


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


def modern(fm):
    phase = fm.get("phase", "").lower()
    created = fm.get("created", "")
    return ("v3r6" in phase) or phase.startswith("v3.0") or (created >= "2026-04-01" and created != "")


hits = [i for i in ids if modern(frontmatter(".moai/specs/%s/spec.md" % i))]
rest = [i for i in ids if i not in hits]
print("V3R5-classified SPECs: %d" % len(ids))
print("  carrying a modern-era signal (misclassified): %d" % len(hits))
print("  carrying no modern-era signal (correctly V3R5): %d" % len(rest))
print("  no-signal set: %s" % rest)
sys.exit(1 if hits else 0)
