#!/usr/bin/env python3
"""Tell a superseded post-merge `ci` run from a failed one (SPEC-20261117-ci-superseded-neutral,
card C-83).

On a push to the integration branch, each macOS target of `.github/workflows/ci.yml` shares one
concurrency group across runs, so a newer merge replaces a macOS job that is still waiting
(#2554 D8 (a)). The replaced job ends `cancelled`, the matrix result `needs.check.result` reads
`cancelled`, and the `ci` aggregate used to fail although no test failed: 40 of 80 push runs
(full survey 2026-11-17, I6). The newer run carries that merge's macOS result.

C-20 (SPEC-20261119-ci-check-set-legs REQ-8): the check jobs are `ci / <target> / <leg>`, and a
waiting macOS job's group is per (target, leg), so the target is the first ` / `-separated
segment after `ci / ` and the wait-replacement is the concurrency message of the job's own
`ci-push-<target>-<leg>` group. This script ships in the same commit as the workflow that
names the legs (SPEC decision D6), so a job of this run never has the pre-split shape.

The `ci-verdict` job runs this script only when the matrix result is `cancelled` on a push. It
prints a Markdown summary on stdout and appends one line to $GITHUB_OUTPUT:

  state=superseded  every `ci / <target> / <leg>` job passed, except at least one macOS leg job
                    that was cancelled before it started (no runner, no step) with exactly
                    GitHub's concurrency message for its own group, `ci-push-<target>-<leg>`;
  state=fail        anything else, including a failed leg, a timeout, a manual cancel, a
                    cancelled job that started, another event, or input it cannot read.

The `ci` job is skipped on `superseded` (a neutral check) and runs, and fails, otherwise.

Input: $EVENT and $CHECK_RESULT, and the run's jobs, each with the messages of its check-run
annotations. In CI the jobs are read with `gh api` (GH_TOKEN, GITHUB_REPOSITORY,
GITHUB_RUN_ID, GITHUB_RUN_ATTEMPT; permissions actions: read and checks: read); `--jobs-file`
reads them from a file instead (the tests). Exit 0 when a state was written, 2 when the jobs
could not be read (the `ci` job then runs and fails).
"""

import json
import os
import subprocess
import sys

TARGET_PREFIX = "ci / "
SUPERSEDED = "Canceling since a higher priority waiting request for ci-push-{target}-{leg} exists"


def parse_target_leg(name):
    """`ci / <target> / <leg>`: the target is the first ` / `-separated segment after the
    prefix (the tip gate reads `ci / …` names by the same prefix, src/gate.rs); the leg is
    the second, empty for a name without one (the pre-split shape, never this run's)."""
    target, _, leg = name[len(TARGET_PREFIX):].partition(" / ")
    return target, leg


def superseded(job, target, leg):
    """A macOS leg job replaced while waiting: cancelled, never started, and cancelled with
    the concurrency message of its own push group."""
    return (
        target.startswith("macos-")
        and job.get("conclusion") == "cancelled"
        and job.get("steps") == []
        and job.get("runner_name") in ("", None)
        and isinstance(job.get("annotations"), list)
        and SUPERSEDED.format(target=target, leg=leg) in job["annotations"]
    )


def decide(event, check_result, jobs):
    """Returns ("superseded" | "fail", reasons)."""
    if event != "push":
        return "fail", [f"event {event!r}: only a push run can be superseded"]
    if check_result != "cancelled":
        return "fail", [f"matrix result {check_result!r}, not 'cancelled'"]
    if not isinstance(jobs, list):
        return "fail", ["the run's jobs could not be read"]
    reasons, replaced, targets = [], 0, 0
    for job in jobs:
        name = job.get("name") if isinstance(job, dict) else None
        if not isinstance(name, str) or not name.startswith(TARGET_PREFIX):
            if not isinstance(job, dict) or not isinstance(name, str):
                return "fail", ["a job of the run could not be read"]
            continue
        targets += 1
        target, leg = parse_target_leg(name)
        if job.get("conclusion") == "success":
            reasons.append(f"`{name}`: success")
        elif superseded(job, target, leg):
            replaced += 1
            reasons.append(f"`{name}`: replaced while waiting by a newer merge (its result comes from that run)")
        else:
            return "fail", reasons + [f"`{name}`: {job.get('conclusion')!r}, not a superseded wait"]
    if targets == 0 or replaced == 0:
        return "fail", reasons + ["no target was superseded"]
    return "superseded", reasons


def gh(path):
    out = subprocess.run(["gh", "api", path], capture_output=True, text=True, check=True)
    return json.loads(out.stdout)


def fetch_jobs():
    """The run attempt's jobs, each with its annotation messages (only cancelled targets)."""
    repo = os.environ["GITHUB_REPOSITORY"]
    run, attempt = os.environ["GITHUB_RUN_ID"], os.environ["GITHUB_RUN_ATTEMPT"]
    jobs = gh(f"repos/{repo}/actions/runs/{run}/attempts/{attempt}/jobs?per_page=100")["jobs"]
    for job in jobs:
        job["annotations"] = []
        if job.get("name", "").startswith(TARGET_PREFIX) and job.get("conclusion") == "cancelled":
            notes = gh(f"repos/{repo}/check-runs/{int(job['id'])}/annotations")
            job["annotations"] = [n.get("message", "") for n in notes]
    return jobs


def main(argv):
    try:
        if argv[1:2] == ["--jobs-file"] and len(argv) == 3:
            with open(argv[2], encoding="utf-8") as f:
                jobs = json.load(f)
        elif len(argv) == 1:
            jobs = fetch_jobs()
        else:
            print("usage: ci_superseded.py [--jobs-file <file>]", file=sys.stderr)
            return 2
    except (OSError, KeyError, ValueError, subprocess.CalledProcessError) as err:
        print(f"ci-verdict: the run's jobs could not be read ({type(err).__name__})", file=sys.stderr)
        return 2
    state, reasons = decide(os.environ.get("EVENT", ""), os.environ.get("CHECK_RESULT", ""), jobs)
    print(f"## ci-verdict: {state}\n")
    for reason in reasons:
        print(f"- {reason}")
    output = os.environ.get("GITHUB_OUTPUT")
    if output:
        with open(output, "a", encoding="utf-8") as f:
            f.write(f"state={state}\n")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
