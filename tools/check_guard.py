#!/usr/bin/env python3
"""check-guard: a pull request cannot weaken the checks that judge it (C-97,
SPEC-20261119-check-guard).

Reads the protected check paths from `tools/protected_checks.json` — always the BASE
checkout's copy, which `.github/workflows/check-guard.yml` hands over after checking out
`github.event.pull_request.base.sha`, so a PR can never read or weaken the list it is
judged by. The pull request's changed paths arrive as one file, one path per line, with
rename old names (`previous_filename`) and deletions already included.

Emits a JSON verdict on stdout and a Markdown summary (the check-run summary) to the
`--summary` file, and exits 0 (pass) or 1 (fail or fail-safe). Option errors are usage,
exit 2 (#3223 B1 contract). Fail-safe mirrors tools/ci_scope.py: a broken list, an
unreadable file list or body never passes.

Enforcement (SPEC D1): protected paths changed + the body declaring exactly
`audit-tier: 2` (the same grammar as `gate::parse_body_tier`, src/gate.rs:280) passes;
a missing or lower declaration fails with `declare audit-tier: 2`. The declared tier 2
is what `moai audit-gate` (src/audit_gate.rs:144-148) then judges by, so the requirement
reaches the merge gate. The `self-modifying check` label is evidence, not enforcement:
its `gh` calls go through `run_gh`, and a label failure is recorded without flipping the
verdict (REQ-9).
"""

import argparse
import fnmatch
import json
import subprocess
import sys

LABEL_DESCRIPTION = "changes a protected check path; judged by the base branch list"


def run_gh(args):
    """Run one `gh` invocation; tests replace this seam (no network in tests)."""
    proc = subprocess.run(["gh", *args], capture_output=True, text=True)
    return proc.returncode, (proc.stdout or "") + (proc.stderr or "")


def declared_tier(body):
    """The `audit-tier:` the body declares, with gate::parse_body_tier's exact grammar:
    among the first 10 lines the first starting with `audit-tier:` decides; it must read
    `audit-tier: ` + 0/1/2 followed by end of line or a space. Anything else is None."""
    if body is None:
        return None
    for raw in body.splitlines()[:10]:
        line = raw[:-1] if raw.endswith("\r") else raw
        if not line.startswith("audit-tier:"):
            continue
        rest = line[len("audit-tier: "):] if line.startswith("audit-tier: ") else ""
        if rest and rest[0] in "012" and (len(rest) == 1 or rest[1] == " "):
            return int(rest[0])
        return None
    return None


def entry_matches(entry, path):
    """One protected entry against one path: `/**` is a folder prefix (checked first —
    it contains `*`), an entry with `*` is a whole-path glob, the rest are exact names."""
    if entry.endswith("/**"):
        return path.startswith(entry[:-2])
    if "*" in entry:
        return fnmatch.fnmatchcase(path, entry)
    return path == entry


def load_list(path):
    """The protected entries, or None when the list is missing or broken (fail safe)."""
    try:
        with open(path, encoding="utf-8") as handle:
            data = json.load(handle)
    except (OSError, ValueError):
        return None
    entries = data.get("protected") if isinstance(data, dict) else None
    if not isinstance(entries, list) or not all(isinstance(e, str) and e for e in entries):
        return None
    return entries


def load_paths(path):
    """The changed paths, or the reason they are unusable (fail safe on empty,
    unreadable or control-character input). Empty lines are ignored: the workflow's
    jq step lets an absent `previous_filename` through as nothing."""
    try:
        with open(path, encoding="utf-8") as handle:
            lines = handle.read().splitlines()
    except OSError:
        return None, "the file list could not be read"
    paths = [line for line in lines if line]
    if not paths:
        return None, "the file list is empty"
    for line in paths:
        if any(ord(ch) < 32 or ord(ch) == 127 for ch in line):
            return None, "a changed path is not readable text; failing safe"
    return paths, None


def load_body(path):
    try:
        with open(path, encoding="utf-8") as handle:
            return handle.read()
    except OSError:
        return None


def attempt_label(runner, opts):
    """Create the label if needed, then add it to the PR. Returns (state, detail):
    `attempted`, `failed` or `skipped`; a failure detail carries gh's own words."""
    if not (opts.label and opts.pr and opts.repo):
        return "skipped", None
    code, out = runner(
        ["label", "create", opts.label, "--repo", opts.repo, "--force",
         "--color", "D93F0B", "--description", LABEL_DESCRIPTION]
    )
    if code != 0 and "already_exists" not in out:
        return "failed", out.strip()
    code, out = runner(
        ["pr", "edit", str(opts.pr), "--repo", opts.repo, "--add-label", opts.label]
    )
    if code != 0:
        return "failed", out.strip()
    return "attempted", None


def write_summary(path, verdict, reasons, protected_paths, declared, label_state, label_detail):
    lines = ["## check-guard", ""]
    lines.append(f"verdict: **{verdict}**")
    for reason in reasons:
        lines.append(f"- {reason}")
    if protected_paths:
        lines.append("- protected paths changed:")
        for changed in protected_paths:
            lines.append(f"  - `{changed}`")
    else:
        lines.append("- no protected path changed")
    if declared is None:
        lines.append("- declared tier: none")
    else:
        lines.append(f"- declared tier: {declared}")
    lines.append(f"- label `self-modifying check`: {label_state}")
    if label_detail:
        lines.append(f"  - gh: {label_detail}")
    lines.append("")
    lines.append(
        "A protected-path change needs `audit-tier: 2` in the pull request body's first"
        " ten lines; the list is read from the base branch (SPEC-20261119-check-guard)."
    )
    try:
        with open(path, "w", encoding="utf-8") as handle:
            handle.write("\n".join(lines))
    except OSError:
        pass  # the verdict on stdout stays authoritative even when the summary is unwritable


def fail_safe(summary_path, reason):
    """No verdict: nothing on stdout, the reason in the summary, exit 1."""
    if summary_path:
        write_summary(summary_path, "fail", [reason], [], None, "skipped", None)
    return 1


def parse_args(argv):
    parser = argparse.ArgumentParser(
        prog="check_guard.py",
        description="Judge a pull request against the base branch's protected check list.",
    )
    parser.add_argument("--list", required=True, help="the base checkout's protected_checks.json")
    parser.add_argument("--files", required=True, help="changed paths, one per line (old+new names)")
    parser.add_argument("--body", required=True, help="the pull request body text")
    parser.add_argument("--summary", help="Markdown summary path (the check-run summary)")
    parser.add_argument("--pr", help="PR number, for the label")
    parser.add_argument("--repo", help="owner/name, for the label")
    parser.add_argument("--label", help="label name; give --pr and --repo with it")
    try:
        return parser.parse_args(argv)
    except SystemExit:
        # argparse wrote the usage to stderr; mirror it on stdout for the job log
        # and re-raise with its exit 2 (#3223 B1: option errors are usage).
        print(
            "usage: check_guard.py --list LIST --files FILES --body BODY "
            "[--summary PATH] [--pr N --repo OWNER/NAME --label LABEL]"
        )
        raise


def main(argv=None):
    opts = parse_args(argv if argv is not None else sys.argv[1:])
    if bool(opts.pr) != bool(opts.repo) or (opts.label and not (opts.pr and opts.repo)):
        print(
            "usage: --label needs both --pr and --repo (and --repo needs --pr); "
            "nothing was judged"
        )
        return 2

    entries = load_list(opts.list)
    if entries is None:
        return fail_safe(opts.summary, "the protected list could not be read")
    paths, reason = load_paths(opts.files)
    if paths is None:
        return fail_safe(opts.summary, reason)
    body = load_body(opts.body)
    if body is None:
        return fail_safe(opts.summary, "the pull request body could not be read")

    protected_paths = []
    for path in paths:
        if any(entry_matches(entry, path) for entry in entries) and path not in protected_paths:
            protected_paths.append(path)
    declared = declared_tier(body)

    if not protected_paths:
        verdict, reasons = "pass", ["no protected path changed"]
    elif declared == 2:
        verdict, reasons = "pass", [
            "protected paths changed and audit-tier: 2 is declared"
        ]
    else:
        verdict, reasons = "fail", [
            "declare audit-tier: 2 — a protected check path changed and the body "
            + ("declares no audit-tier" if declared is None else f"declares audit-tier: {declared}")
        ]

    label_state, label_detail = attempt_label(run_gh, opts)
    write_summary(
        opts.summary, verdict, reasons, protected_paths, declared, label_state, label_detail
    )
    json.dump(
        {
            "verdict": verdict,
            "protected": bool(protected_paths),
            "protected_paths": protected_paths,
            "declared_tier": declared,
            "reasons": reasons,
            "label": label_state,
            "label_detail": label_detail,
        },
        sys.stdout,
    )
    sys.stdout.write("\n")
    return 0 if verdict == "pass" else 1


if __name__ == "__main__":
    sys.exit(main())
