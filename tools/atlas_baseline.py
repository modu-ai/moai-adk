#! /usr/bin/env python3
"""C10a baseline consistency check (Issue #2346).

Runs the committed fixtures (corpus.json) through the rg-only baseline and
verifies the committed baseline.json still matches: same task ids, same hit
sets, same thresholds. A mismatch means the corpus moved or a threshold was
loosened without re-running the baseline.

Search errors are fatal, including partial output with rg exit status 2.
Exit codes: 0 consistent, 1 inconsistent or search failure, 2 usage.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
FIXTURES = ROOT / "tests" / "fixtures" / "atlas"


def corpus() -> dict:
    return json.loads((FIXTURES / "corpus.json").read_text(encoding="utf-8"))


def baseline() -> dict:
    return json.loads((FIXTURES / "baseline.json").read_text(encoding="utf-8"))


def rg_version() -> str:
    out = subprocess.run(["rg", "--version"], capture_output=True, text=True)
    return out.stdout.splitlines()[0].strip() if out.returncode == 0 else "rg unavailable"


def hit_files(pattern: str) -> list[str]:
    """Files under src/ whose bytes carry the fixed-string pattern, sorted."""
    out = subprocess.run(
        ["rg", "-l", "--fixed-strings", "--no-config", pattern, "src"],
        capture_output=True,
        text=True,
    )
    if out.returncode not in (0, 1):
        raise RuntimeError(
            f"rg search failed (exit {out.returncode}): {out.stderr.strip()}"
        )
    return sorted({line for line in out.stdout.splitlines() if line.strip()})


def measure(task: dict) -> dict:
    hits = hit_files(task["pattern"])
    truth = sorted(task["ground_truth"]["files"])
    hit_set, truth_set = set(hits), set(truth)
    if task["kind"] == "graph":
        # A relation task names the one referencing file: the score is binary —
        # the relation holds iff the pattern is found in that file.
        ok = truth_set <= hit_set
        return {
            "id": task["id"],
            "hits": hits,
            "precision": 1.0 if ok else 0.0,
            "recall": 1.0 if ok else 0.0,
        }
    correct = len(hit_set & truth_set)
    precision = round(correct / len(hit_set), 4) if hit_set else 1.0
    recall = round(correct / len(truth_set), 4) if truth_set else 1.0
    return {
        "id": task["id"],
        "hits": hits,
        "precision": precision,
        "recall": recall,
    }


def build() -> dict:
    tasks = [t for t in corpus()["tasks"] if t["kind"] != "freshness"]
    rows = sorted((measure(t) for t in tasks), key=lambda r: r["id"])
    precisions = [r["precision"] for r in rows]
    recalls = [r["recall"] for r in rows]
    return {
        "schema_version": 1,
        "tool": "rg",
        "rg_version": rg_version(),
        "corpus_sha256": hashlib.sha256(
            (FIXTURES / "corpus.json").read_bytes()
        ).hexdigest(),
        "tasks": rows,
        "thresholds": {
            "precision_min": min(precisions),
            "recall_min": min(recalls),
            "precision_mean": round(sum(precisions) / len(precisions), 4),
            "recall_mean": round(sum(recalls) / len(recalls), 4),
        },
    }


def check() -> int:
    fresh, committed = build(), baseline()
    # Mechanical reviewer check: every ground truth file must be matched by the
    # task's own rg pattern — a ground truth the pattern cannot find is a wrong
    # answer, not a low score.
    by_id = {t["id"]: t for t in corpus()["tasks"] if t["kind"] != "freshness"}
    for row in fresh["tasks"]:
        task = by_id[row["id"]]
        truth = set(task["ground_truth"]["files"])
        missing = sorted(truth - set(row["hits"]))
        if missing:
            print(f"inconsistent: {row['id']} ground truth not matched by the pattern: {missing}")
            return 1
    if fresh["corpus_sha256"] != committed["corpus_sha256"]:
        print("inconsistent: corpus.json changed since the baseline was built")
        return 1
    if fresh["tasks"] != committed["tasks"]:
        print("inconsistent: rg hit sets or scores differ from the committed baseline")
        return 1
    for key in ("precision_min", "recall_min", "precision_mean", "recall_mean"):
        if fresh["thresholds"][key] != committed["thresholds"][key]:
            print(f"inconsistent: threshold {key} was changed")
            return 1
    print("baseline consistent: %d tasks, thresholds pinned" % len(committed["tasks"]))
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--build", action="store_true", help="rewrite baseline.json")
    args = parser.parse_args()
    if args.build:
        (FIXTURES / "baseline.json").write_text(
            json.dumps(build(), indent=1, sort_keys=True) + "\n", encoding="utf-8"
        )
        return 0
    return check()


if __name__ == "__main__":
    sys.exit(main())
