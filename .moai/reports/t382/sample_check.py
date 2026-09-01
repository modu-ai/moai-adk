"""t382 AC-EH3-005 proposition-4 sample check.

For each SPEC id given on argv, print the frontmatter `created:` value beside two
independent corroborations — the first dated row of the body HISTORY table, and the
author date of the first commit that touched the SPEC directory — so a reader can
confirm the SPEC really is modern-era rather than trusting the frontmatter alone.
"""
import re
import subprocess
import sys

for spec_id in sys.argv[1:]:
    path = ".moai/specs/%s/spec.md" % spec_id
    text = open(path, encoding="utf-8").read()
    m = re.search(r"^created:\s*(.+)$", text, re.M)
    created = m.group(1).strip().strip('"').strip("'") if m else ""
    hist = re.search(r"^\|\s*\d+\.\d+\.\d+\s*\|\s*(\d{4}-\d{2}-\d{2})", text, re.M)
    first_hist = hist.group(1) if hist else "-"
    out = subprocess.run(
        ["git", "log", "--diff-filter=A", "--format=%ad", "--date=short", "--", path],
        capture_output=True, text=True)
    lines = [x for x in out.stdout.split("\n") if x.strip()]
    first_commit = lines[-1] if lines else "-"
    verdict = "MODERN" if created and created >= "2026-04-01" else "pre-threshold-created"
    print("%-40s created=%-11s HISTORY-first=%-11s git-added=%-11s %s"
          % (spec_id, created, first_hist, first_commit, verdict))
