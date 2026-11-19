#!/usr/bin/env python3
"""Grade the audit agent on bugs that mattered (SPEC-20261119-audit-eval, card C-102).

`run` (dry-run, the default and only mode) builds the audit prompt of every case of a set and
writes it to an output folder — no model is called and no network is used; the case's answer
key never enters the prompt. `score` reads a findings file (what an audit agent reported per
case) and counts, per case, only findings that describe the case's bug — a finding counts when
at least TOKEN_FRACTION of the answer key's distinctive tokens appear in it — reporting the
raw-finding recall and the final-verdict recall separately, and the over-flagging rate on the
controls. A FAIL verdict without a describing finding counts in neither recall (a hollow
verdict); a describing finding under a PASS verdict counts in the raw recall only.

Refusals exit 2 with the reason on stderr: a record without its fields (REQ-2), a case id the
chosen set does not hold, a findings file that is not a JSON list (REQ-6), a tuning/held-out id
shared (REQ-9). A ref the repository does not hold is not a refusal: the case is named in the
report as `missing_ref` and skipped (REQ-8).
"""

import argparse
import json
import os
import re
import subprocess
import sys

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

#: A finding describes the bug when at least this fraction of the answer key's distinctive
#: tokens appear in it. A ruler for comparing runs, not a truth.
TOKEN_FRACTION = 0.5

#: Words too common to carry a bug's identity (only tokens of four or more characters are
#: distinctive at all).
STOPWORDS = {
    "that", "this", "with", "from", "when", "then", "them", "there", "their",
    "into", "only", "never", "every", "have", "been", "were", "will", "would",
    "cannot", "could", "should", "still", "some", "more", "most", "over",
}

REQUIRED_FIELDS = ("id", "pr", "audit_comment", "head", "base")
CASE_FIELDS = REQUIRED_FIELDS + ("answer_key",)


def refuse(message):
    print(f"audit_eval refused: {message}", file=sys.stderr)
    sys.exit(2)


def load_file(path, name):
    try:
        rows = json.load(open(path))
    except (OSError, json.JSONDecodeError) as e:
        refuse(f"{name} cannot be read: {e}")
    if not isinstance(rows, list):
        refuse(f"{name} is not a JSON list")
    return rows


def load_set(eval_dir, kind):
    names = {
        "tuning": "cases-tuning.json",
        "heldout": "cases-heldout.json",
        "controls": "controls.json",
    }
    return load_file(os.path.join(eval_dir, names[kind]), names[kind])


def validate(eval_dir, kinds):
    """The shared rules of `run` and `score`: fields present, no id in two files."""
    tuning = load_set(eval_dir, "tuning")
    heldout = load_set(eval_dir, "heldout")
    shared = {c["id"] for c in tuning if isinstance(c, dict)} & {
        c["id"] for c in heldout if isinstance(c, dict)
    }
    if shared:
        refuse(
            f"the tuning and held-out files share {', '.join(sorted(shared))} "
            "(cases-tuning.json, cases-heldout.json); a case scored twice measures nothing"
        )
    rows = []
    for kind in kinds:
        required = CASE_FIELDS if kind in ("tuning", "heldout") else REQUIRED_FIELDS
        for row in load_set(eval_dir, kind):
            missing = [f for f in required if f not in row]
            if missing:
                refuse(
                    f"a {kind} record is missing {', '.join(missing)}: "
                    f"{json.dumps(row, ensure_ascii=False)[:200]}"
                )
            if "/" in row["id"] or "\\" in row["id"] or row["id"] in (".", ".."):
                refuse(
                    f"the case id holds a path separator: {row['id']!r}; a case id names a "
                    "prompt file, so `/` and `\\` (and `.`/`..`) are refused"
                )
            rows.append(row)
    return rows


def ref_missing(repo, sha):
    env = dict(os.environ)
    # A partial clone would lazy-fetch the ref to answer; the audit evaluates only what the
    # repository already has (card C-107).
    env["GIT_NO_LAZY_FETCH"] = "1"
    return (
        subprocess.run(
            ["git", "-C", repo, "cat-file", "-e", sha + "^{commit}"],
            capture_output=True,
            env=env,
        ).returncode
        != 0
    )


def missing_refs(repo, rows):
    """`id:head` / `id:base` for each ref of a case the repository does not have (card C-107):
    the case alone left which of the two refs to fix unclear."""
    out = []
    for row in rows:
        for which in ("head", "base"):
            if ref_missing(repo, row[which]):
                out.append(f"{row['id']}:{which}")
    return out


def prompt_for(case, repo):
    return f"""# Audit case {case['id']} (pull request #{case['pr']})

You are the read-only audit agent. The repository is at `{repo}`. Review the change between:

- base `{case['base']}`
- head `{case['head']}`

Read the tree at the head (`git -C {repo} show {case['head']}:<path>`), the diff
(`git -C {repo} diff {case['base']} {case['head']}`), and the repository's own rules. Report
each finding as one sentence and end with a verdict line `verdict: PASS` or `verdict: FAIL`.
Score nothing yourself: another tool scores your findings against this case's answer key.
"""


def cmd_run(args):
    kinds = ["tuning", "heldout", "controls"] if args.set == "all" else [args.set]
    rows = validate(args.eval_dir, kinds)
    missing = missing_refs(args.repo, rows)
    missing_ids = {m.split(":", 1)[0] for m in missing}
    try:
        os.makedirs(args.out, exist_ok=True)
    except OSError as e:
        refuse(f"the output folder cannot be written: {args.out}: {e}")
    written = 0
    for row in rows:
        if row["id"] in missing_ids:
            continue
        with open(os.path.join(args.out, row["id"] + ".md"), "w") as fh:
            fh.write(prompt_for(row, args.repo))
        written += 1
    print(json.dumps({"set": args.set, "written": written, "missing_ref": sorted(missing)}))


def distinctive_tokens(answer_key):
    words = re.findall(r"[a-z0-9_]{4,}", answer_key.lower())
    return [w for w in words if w not in STOPWORDS]


def describes(answer_key, finding):
    tokens = distinctive_tokens(answer_key)
    if not tokens:
        return False
    low = finding.lower()
    hit = sum(1 for t in tokens if re.search(rf"\b{re.escape(t)}\b", low))
    return hit / len(tokens) >= TOKEN_FRACTION


def cmd_score(args):
    kinds = ["tuning", "heldout"] if args.set == "all" else [args.set]
    incidents = validate(args.eval_dir, kinds)
    controls = validate(args.eval_dir, ["controls"])
    findings = load_file(args.findings, os.path.basename(args.findings))
    known = {c["id"] for c in incidents} | {c["id"] for c in controls}
    for row in findings:
        if not isinstance(row, dict) or row.get("case_id") not in known:
            refuse(f"the findings file names a case the set does not hold: {json.dumps(row)[:200]}")
    rows_missing = missing_refs(args.repo, incidents + controls)
    by_id = {c["id"]: c for c in incidents + controls}
    rows = {r["case_id"]: r for r in findings if r["case_id"] not in rows_missing}
    cases = [c for c in incidents if c["id"] not in rows_missing]
    ctrls = [c for c in controls if c["id"] not in rows_missing]
    described = 0
    verdict_hits = 0
    for case in cases:
        row = rows.get(case["id"])
        if not row:
            continue
        found = any(describes(case["answer_key"], f) for f in row.get("findings", []))
        if found:
            described += 1
            if str(row.get("verdict", "")).upper() == "FAIL":
                verdict_hits += 1
    flagged = sum(
        1
        for control in ctrls
        if rows.get(control["id"], {}).get("findings")
    )
    print(
        json.dumps(
            {
                "set": args.set,
                "cases": len(cases),
                "described": described,
                "recall_raw": described / len(cases) if cases else None,
                "recall_verdict": verdict_hits / len(cases) if cases else None,
                "controls": len(ctrls),
                "controls_flagged": flagged,
                "control_overflag": flagged / len(ctrls) if ctrls else None,
                "missing_ref": sorted(rows_missing),
            }
        )
    )


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)

    run = sub.add_parser("run", help="build the audit prompts (dry-run; no model, no network)")
    run.add_argument("--set", choices=["tuning", "heldout", "controls", "all"], default="tuning")
    run.add_argument("--eval-dir", default=os.path.join(REPO_ROOT, ".moai", "evals", "audit"))
    run.add_argument("--repo", default=REPO_ROOT)
    run.add_argument("--out", required=True)
    run.set_defaults(fn=cmd_run)

    score = sub.add_parser("score", help="score a findings file against a set")
    score.add_argument("--set", choices=["tuning", "heldout", "all"], default="tuning")
    score.add_argument("--eval-dir", default=os.path.join(REPO_ROOT, ".moai", "evals", "audit"))
    score.add_argument("--repo", default=REPO_ROOT)
    score.add_argument("--findings", required=True)
    score.set_defaults(fn=cmd_score)

    args = parser.parse_args(argv)
    args.fn(args)


if __name__ == "__main__":
    main()
