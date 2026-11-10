#! /usr/bin/env python3
"""C10a frozen-input baseline and threshold judge (Issue #2346).

Default: verify baseline.json against source_snapshot.json, never live src/.
--build: regenerate the baseline after all searches succeed.
--judge REPORT: score complete per-task file-hit sets against the same oracle
and apply the baseline's per-suite macro-mean floors with >=. Minima are
diagnostic only; there are no per-task floors. Retrieval uses full sets, not
precision@k/recall@k. Empty retrieval results score zero. GRAPH measures only
coverage of the required relation file, not semantic graph correctness.
Freshness and actual Map/Graph arm execution remain C6/C10b not_run.

Search errors, including partial stdout with exit 2, never rewrite the baseline.
Exit codes: 0 consistent/passed, 1 inconsistent/failed, 2 usage.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import subprocess
import sys
from pathlib import Path, PurePosixPath
from tempfile import TemporaryDirectory

ROOT = Path(__file__).resolve().parent.parent
FIXTURES = ROOT / "tests" / "fixtures" / "atlas"


def load_inputs() -> tuple[dict, dict[str, bytes], dict]:
    raw = (FIXTURES / "corpus.json").read_bytes()
    data = json.loads(raw)
    snapshot_raw = (FIXTURES / "source_snapshot.json").read_bytes()
    snapshot_hash = hashlib.sha256(snapshot_raw).hexdigest()
    if data["schema_version"] != 2 or data["source_snapshot_sha256"] != snapshot_hash:
        raise ValueError("corpus schema or frozen source digest differs")
    snapshot = json.loads(snapshot_raw)
    if snapshot["schema_version"] != 1 or not snapshot["files"]:
        raise ValueError("invalid frozen source")
    files = {}
    for name, entry in snapshot["files"].items():
        path = PurePosixPath(name)
        if (
            path.is_absolute() or len(path.parts) < 2 or path.parts[0] != "src"
            or ".." in path.parts or "\\" in name or str(path) != name
        ):
            raise ValueError(f"invalid frozen source path: {name}")
        content = entry["content"].encode("utf-8")
        if hashlib.sha256(content).hexdigest() != entry["sha256"]:
            raise ValueError(f"frozen source content digest differs: {name}")
        files[name] = content
    tasks = data["tasks"]
    if len(tasks) < 30:
        raise ValueError("the corpus needs at least 30 tasks")
    ids = set()
    kinds = set()
    for task in tasks:
        if not task["id"] or task["id"] in ids or not task["source"].strip():
            raise ValueError("task ids must be unique and source citations nonempty")
        ids.add(task["id"])
        kind = task["kind"]
        kinds.add(kind)
        if kind == "freshness":
            if not task["suite_ref"]:
                raise ValueError("freshness task needs its suite reference")
            continue
        if kind not in ("retrieval", "graph") or task["roots"] != ["src"] or not task["pattern"]:
            raise ValueError(f"invalid search task: {task['id']}")
        truth = task["ground_truth"]["files"]
        if not truth or any(name not in files for name in truth):
            raise ValueError(f"{task['id']}: ground truth outside frozen source")
    if kinds != {"retrieval", "graph", "freshness"}:
        raise ValueError("the corpus needs all three task kinds")
    identity = {
        "corpus_sha256": hashlib.sha256(raw).hexdigest(),
        "source_snapshot_sha256": snapshot_hash,
        "source_commit": snapshot["source_commit"],
        "source_tree_sha1": snapshot["source_tree_sha1"],
    }
    return data, files, identity


def baseline() -> dict:
    return json.loads((FIXTURES / "baseline.json").read_text(encoding="utf-8"))


def rg_version() -> str:
    out = subprocess.run(["rg", "--no-config", "--version"], capture_output=True, text=True)
    if out.returncode != 0 or not out.stdout.splitlines():
        raise RuntimeError(f"rg version check failed: {out.stderr.strip()}")
    return out.stdout.splitlines()[0].strip()


def hit_files(pattern: str, source_root: Path) -> list[str]:
    """Search only the materialized immutable input, with deterministic file ids."""
    out = subprocess.run(
        [
            "rg", "-l", "--fixed-strings", "--no-config", "--no-ignore",
            "--hidden", "--path-separator=/", "--", pattern, "src",
        ],
        cwd=source_root, capture_output=True, text=True, encoding="utf-8",
    )
    if out.returncode not in (0, 1):
        raise RuntimeError(
            f"rg search failed (exit {out.returncode}): {out.stderr.strip()}"
        )
    return sorted({line for line in out.stdout.splitlines() if line.strip()})


def measure(task: dict, hits: list[str], source: dict[str, bytes]) -> dict:
    if not isinstance(hits, list) or any(not isinstance(h, str) or h not in source for h in hits):
        raise ValueError(f"{task['id']}: hits must name files in the frozen source")
    hit_set = set(hits)
    truth_set = set(task["ground_truth"]["files"])
    row = {"id": task["id"], "kind": task["kind"], "hits": sorted(hit_set)}
    if task["kind"] == "graph":
        row["relation_file_coverage"] = 1.0 if truth_set <= hit_set else 0.0
    else:
        correct = len(hit_set & truth_set)
        row["precision"] = round(correct / len(hit_set), 4) if hit_set else 0.0
        row["recall"] = round(correct / len(truth_set), 4)
    return row


def summary(rows: list[dict]) -> dict:
    retrieval = [r for r in rows if r["kind"] == "retrieval"]
    graph = [r for r in rows if r["kind"] == "graph"]
    return {
        "retrieval_precision_mean": round(math.fsum(r["precision"] for r in retrieval) / len(retrieval), 4),
        "retrieval_recall_mean": round(math.fsum(r["recall"] for r in retrieval) / len(retrieval), 4),
        "graph_relation_file_coverage_mean": round(
            math.fsum(r["relation_file_coverage"] for r in graph) / len(graph), 4
        ),
    }


def build(data: dict, files: dict[str, bytes], identity: dict) -> dict:
    tasks = [t for t in data["tasks"] if t["kind"] != "freshness"]
    with TemporaryDirectory(prefix="moai-c10-source-") as directory:
        root = Path(directory)
        for name, content in files.items():
            path = root / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(content)
        rows = sorted(
            (measure(t, hit_files(t["pattern"], root), files) for t in tasks),
            key=lambda r: r["id"],
        )
    retrieval = [r for r in rows if r["kind"] == "retrieval"]
    return {
        "schema_version": 2,
        "tool": "rg",
        "rg_version": rg_version(),
        **identity,
        "tasks": rows,
        "threshold_policy": {
            "comparison": ">=",
            "scope": "per-suite macro means; no per-task floors",
            "minima": "diagnostic only",
            "graph": "relation-file coverage proxy, not semantic graph correctness",
        },
        "thresholds": summary(rows),
        "diagnostics": {
            "retrieval_precision_min": min(r["precision"] for r in retrieval),
            "retrieval_recall_min": min(r["recall"] for r in retrieval),
        },
    }


def check(fresh: dict, committed: dict) -> None:
    # Version is provenance, not a floor: supported rg versions must agree on hits.
    for key in (
        "schema_version", "corpus_sha256", "source_snapshot_sha256", "source_commit",
        "source_tree_sha1", "tasks", "threshold_policy", "thresholds", "diagnostics",
    ):
        if fresh[key] != committed.get(key):
            raise ValueError(f"baseline differs: {key}")


def judge(report: dict, data: dict, files: dict[str, bytes], committed: dict) -> dict:
    for key in ("schema_version", "corpus_sha256", "source_snapshot_sha256"):
        if report.get(key) != committed[key]:
            raise ValueError(f"candidate report identity differs: {key}")
    tasks = {t["id"]: t for t in data["tasks"] if t["kind"] != "freshness"}
    rows = []
    seen = set()
    for candidate in report["tasks"]:
        task_id = candidate["id"]
        if task_id not in tasks or task_id in seen:
            raise ValueError("candidate task ids must cover the scored corpus exactly")
        seen.add(task_id)
        rows.append(measure(tasks[task_id], candidate["hits"], files))
    if seen != tasks.keys():
        raise ValueError("candidate task ids must cover the scored corpus exactly")
    metrics = summary(rows)
    failures = [
        key for key, floor in committed["thresholds"].items()
        if not metrics[key] >= floor
    ]
    return {
        "schema_version": 2,
        "corpus_sha256": committed["corpus_sha256"],
        "source_snapshot_sha256": committed["source_snapshot_sha256"],
        "passed": not failures,
        "metrics": metrics,
        "thresholds": committed["thresholds"],
        "failed_metrics": sorted(failures),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    mode = parser.add_mutually_exclusive_group()
    mode.add_argument("--build", action="store_true", help="rewrite baseline.json")
    mode.add_argument("--judge", type=Path, metavar="REPORT", help="judge a complete file-hit report")
    args = parser.parse_args()
    try:
        data, files, identity = load_inputs()
        fresh = build(data, files, identity)
        if args.build:
            (FIXTURES / "baseline.json").write_bytes(
                (json.dumps(fresh, indent=1, sort_keys=True) + "\n").encode("utf-8")
            )
            return 0
        committed = baseline()
        check(fresh, committed)
        if args.judge:
            report = json.loads(args.judge.read_text(encoding="utf-8"))
            result = judge(report, data, files, committed)
            print(json.dumps(result, sort_keys=True))
            return 0 if result["passed"] else 1
        print("baseline consistent: %d tasks, thresholds pinned" % len(committed["tasks"]))
        return 0
    except (ValueError, KeyError, OSError, RuntimeError) as error:
        print(f"inconsistent: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
