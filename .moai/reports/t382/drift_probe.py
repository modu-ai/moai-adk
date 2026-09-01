"""t382 R4 probe — drift-exemption cost on the V3R5 bucket (the SPEC's centre of gravity).

Reads `moai spec audit --json` on stdin to get the V3R5-classified set, then reads
`moai spec drift` output from the path given as argv[1].

Counts era-exempt rows that carry a modern-era signal. Those are the SPECs told they
are exempt from status<->git drift detection on the strength of an age they do not have.
A SPEC with no modern signal (INIT-WIZARD today) is legitimately exempt and is NOT counted.

Exit 1 when any such row exists (the defect present), exit 0 when none (the fixed state).
"""
import json
import re
import sys

audit = json.load(sys.stdin)
ids = [x["spec_id"] for x in (audit.get("drift_findings") or [])
       if x.get("finding_type") == "EraAutoDetected" and x.get("era") == "V3R5"]

drift_rows = {}
for line in open(sys.argv[1], encoding="utf-8"):
    parts = line.split()
    if len(parts) >= 3 and parts[0].startswith("SPEC-"):
        drift_rows[parts[0]] = parts[2]


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


era_exempt, terminal_exempt, other = [], [], []
for i in ids:
    status = drift_rows.get(i, "<no-row>")
    (era_exempt if status == "era-exempt"
     else terminal_exempt if status == "terminal-exempt"
     else other).append(i)

unearned = [i for i in era_exempt if modern(frontmatter(".moai/specs/%s/spec.md" % i))]
earned = [i for i in era_exempt if i not in unearned]

print("V3R5-classified SPECs swept: %d" % len(ids))
print("  era-exempt rows:      %d" % len(era_exempt))
print("  terminal-exempt rows: %d  %s" % (len(terminal_exempt), terminal_exempt))
print("  other/missing rows:   %d  %s" % (len(other), other))
print("  of the era-exempt rows, carrying a modern signal (unearned exemption): %d" % len(unearned))
print("  of the era-exempt rows, no modern signal (earned exemption): %d  %s" % (len(earned), earned))
sys.exit(1 if unearned else 0)
