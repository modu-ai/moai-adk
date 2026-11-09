#!/usr/bin/env python3
"""Flow metrics baseline (#2279 scope 6, read-only).

One gh call (`pr list --state open`) measures the WIP side of the flow: the
open PR count, per-author counts, and PR ages. Everything that would need
per-issue comment crawling — PR→first verdict, PASS→merge, FAIL ratio and
cause classes, sync count — is recorded as not_run: crawling repeats gh calls
per issue, which the call ceiling forbids (decision 06:02 problem list, card
#2279). Re-measuring after the flow-control policy lands is its own not_run
baseline, not this file's job.

Usage:
  python3 -I tools/flow_metrics.py --owner OWNER --repo REPO --out FILE
      [--gh-bin PATH] [--max-gh-calls N] [--now ISO8601]

Exit codes: 0 = baseline written, 1 = call ceiling reached or gh failure,
2 = usage.
"""

import json
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

KNOWN = ("--owner", "--repo", "--out", "--gh-bin", "--max-gh-calls", "--now")


def parse_args(argv):
    opts = {k: None for k in KNOWN}
    opts["--max-gh-calls"] = "4"
    i = 0
    while i < len(argv):
        if argv[i] not in KNOWN or i + 1 >= len(argv):
            return None
        opts[argv[i]] = argv[i + 1]
        i += 2
    if not (opts["--owner"] and opts["--repo"] and opts["--out"]):
        return None
    return opts


def usage():
    print(__doc__)
    raise SystemExit(2)


class GhCounter:
    """The call ceiling is structural: every gh invocation goes through here and
    a call that would exceed the ceiling is refused before it starts."""

    def __init__(self, bin_path: str, ceiling: int):
        self.bin_path = bin_path
        self.ceiling = ceiling
        self.used = 0

    def call(self, args):
        if self.used >= self.ceiling:
            print(
                f"flow-metrics: gh call ceiling reached ({self.ceiling}); "
                f"refusing call #{self.used + 1}",
                file=sys.stderr,
            )
            raise SystemExit(1)
        self.used += 1
        r = subprocess.run([self.bin_path] + args, text=True, capture_output=True)
        if r.returncode != 0:
            print(f"gh call failed: {r.stderr.strip()[:200]}", file=sys.stderr)
            raise SystemExit(1)
        return json.loads(r.stdout)


def main():
    opts = parse_args(sys.argv[1:])
    if opts is None:
        usage()
    gh = GhCounter(opts["--gh-bin"] or "gh", int(opts["--max-gh-calls"]))
    now = (
        datetime.fromisoformat(opts["--now"])
        if opts["--now"]
        else datetime.now(timezone.utc)
    )
    if now.tzinfo is None:
        now = now.replace(tzinfo=timezone.utc)

    pulls = gh.call(
        [
            "pr",
            "list",
            "--repo",
            f"{opts['--owner']}/{opts['--repo']}",
            "--state",
            "open",
            "--json",
            "number,author,createdAt,headRefName",
        ]
    )

    per_author: dict = {}
    ages: dict = {}
    issues: list = []
    for p in pulls:
        author = (p.get("author") or {}).get("login") or "unknown"
        per_author[author] = per_author.get(author, 0) + 1
        created = datetime.fromisoformat(p["createdAt"].replace("Z", "+00:00"))
        ages[str(p["number"])] = round((now - created).total_seconds() / 86400.0, 2)
        # Every lane shares one GitHub account, so the author split cannot see
        # lanes; the branch name (issue/<n>-<slug>) is the lane-resolved split.
        m = re.match(r"issue/(\d+)", p.get("headRefName") or "")
        if m:
            issues.append(int(m.group(1)))

    baseline = {
        "schema_version": 1,
        "measured_at": now.astimezone(timezone.utc).isoformat(),
        "owner": opts["--owner"],
        "repo": opts["--repo"],
        "gh_calls_used": gh.used,
        "gh_call_ceiling": gh.ceiling,
        "open_pr_count": len(pulls),
        "open_issues": sorted(issues),
        "per_author": dict(sorted(per_author.items())),
        "ages_days": dict(sorted(ages.items(), key=lambda kv: int(kv[0]))),
        "max_age_days": max(ages.values(), default=0.0),
        "not_run": [
            "pr_to_first_verdict",
            "pass_to_merge",
            "fail_ratio_and_causes",
            "sync_count",
        ],
        "notes": (
            "baseline measured before the flow-control policy (#2279); "
            "re-measure after the caps land is a separate not_run entry"
        ),
    }
    out = Path(opts["--out"])
    out.write_text(json.dumps(baseline, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(
        f"flow-metrics: open={baseline['open_pr_count']} "
        f"authors={len(per_author)} max_age={baseline['max_age_days']}d "
        f"calls={gh.used}/{gh.ceiling} out={out}"
    )


if __name__ == "__main__":
    main()
