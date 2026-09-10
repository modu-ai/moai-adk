"""t382 D2 check — is the phase-match set a strict subset of the date-match set?

If it is, R1 (which judges phase OR created) cannot distinguish a date-only
implementation of hasModernEraSignal from a correct phase-OR-date one.
"""
import json
import re
import sys

audit = json.load(sys.stdin)
ids = [x["spec_id"] for x in (audit.get("drift_findings") or [])
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


phase_only, date_match, phase_match = [], [], []
for i in ids:
    fm = frontmatter(".moai/specs/%s/spec.md" % i)
    p = fm.get("phase", "").lower()
    c = fm.get("created", "")
    ph = ("v3r6" in p) or p.startswith("v3.0")
    dt = c != "" and c >= "2026-04-01"
    if ph:
        phase_match.append(i)
    if dt:
        date_match.append(i)
    if ph and not dt:
        phase_only.append(i)

print("phase-match: %d  %s" % (len(phase_match), sorted(phase_match)))
print("date-match:  %d" % len(date_match))
print("phase-match AND NOT date-match (the set R1 needs to be non-vacuous): %d  %s"
      % (len(phase_only), phase_only))
print("verdict: R1 %s distinguish the phase path"
      % ("CAN" if phase_only else "CANNOT"))
sys.exit(0 if phase_only else 1)
