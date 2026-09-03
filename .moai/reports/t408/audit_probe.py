"""t408 probe: does the §F.N lettering in progress.md change era classification?

Reads the `moai spec audit --json` output produced from this tree and reports,
for the progress.md files that carry a `§F.1` heading, whether the auditor
classified them as grandfathered (pre-V3R6) or modern-era.
"""

import io
import json
import sys

TARGETS = [
    "SPEC-DEVPROT-REQUIRED-001",
    "SPEC-EVIDENCE-CLAIM-INVARIANT-001",
    "SPEC-FOURDIM-PHANTOM-001",
    "SPEC-HARNESS-OUTCOME-CAPTURE-001",
    "SPEC-PREPUSH-WIRING-001",
    "SPEC-SEC-HARDEN-003",
    "SPEC-STOP-EVIDENCE-WRITER-001",
    "SPEC-V3R6-AGENT-TEAM-REBUILD-001",
    "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001",
    "SPEC-WEB-CONSOLE-007",
]


def main(path):
    d = json.load(io.open(path, encoding="utf-8"))
    for k in ("audited_at", "total_specs", "grandfathered", "modern_era_clean"):
        print("%s = %s" % (k, d.get(k)))
    findings = d.get("drift_findings") or []
    print("drift_findings total = %d" % len(findings))
    blob = [json.dumps(f, ensure_ascii=False) for f in findings]
    for t in TARGETS:
        hits = [b for b in blob if t in b]
        print("  %-42s drift findings: %d" % (t, len(hits)))
        for h in hits[:2]:
            print("      %s" % h[:240])


if __name__ == "__main__":
    main(sys.argv[1])
