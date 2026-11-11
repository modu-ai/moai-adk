#!/usr/bin/env python3
"""C10b: execute cumulative lexical / +Map / +Graph on C10a's frozen input.

No ground truth enters queries or selection. Map selects lexical hits; Graph expands
their context. Literal-relevant, evidence-connected files become scored answers;
all raw nodes, edges and distances are retained separately. File-set scores use
the unchanged C10a judge (a coverage proxy, NOT
proof of a code-reference edge). Costs count actual op responses and calls,
excluding the shared index build. Bytes are not model tokens. Freshness mutation
cases remain #2407; this experiment measures only the unmodified snapshot.
"""
from __future__ import annotations

import argparse
from collections import defaultdict
import hashlib
import json
import os
from pathlib import Path
import subprocess
from tempfile import TemporaryDirectory

import atlas_baseline as baseline


def select_answers(pattern: str, map_doc: dict, graph_doc: dict | None,
                   files: dict[str, bytes]) -> dict:
    """Select literal-relevant answer files from operation evidence, never an oracle."""
    if not pattern:
        raise ValueError("answer selection requires a nonempty literal query")
    seeds = {n["id"][5:] for n in map_doc["nodes"] if n["id"].startswith("file:")}
    if not seeds <= files.keys():
        raise ValueError("Map context escaped the frozen source")
    support = {"file:" + p: [] for p in seeds}
    if graph_doc is not None:
        ids = {n["id"] for n in graph_doc["nodes"]}
        raw_files = {i[5:] for i in ids if i.startswith("file:")}
        if not raw_files <= files.keys():
            raise ValueError("Graph context escaped the frozen source")
        if not support.keys() <= ids:
            raise ValueError("Graph context omitted a Map start node")
        incident = defaultdict(list)
        for e in sorted(graph_doc["edges"], key=lambda e: e["edge_id"]):
            # A returned node alone is not evidence of a connection to the query.
            if not e.get("evidence_refs"):
                continue
            a, b = e["source_id"], e["target_id"]
            if a not in ids or b not in ids:
                raise ValueError("Graph edge omitted an endpoint")
            incident[a].append((b, e["edge_id"]))
            incident[b].append((a, e["edge_id"]))
        frontier = set(support)
        for _ in range(3):
            following = set()
            for start in sorted(frontier):
                for end, edge_id in incident[start]:
                    if end not in support:
                        support[end] = support[start] + [edge_id]
                        following.add(end)
            frontier = following
    candidates = {i[5:] for i in support if i.startswith("file:")}
    hits = sorted(p for p in candidates if pattern.encode("utf-8") in files[p])
    return {"policy": "literal-evidence-v1", "hits": hits,
            "eligible": sorted(candidates),
            "support": {p: support["file:" + p] for p in hits}}


def context_document(documents: dict, document: dict) -> str:
    """Keep complete raw documents once; task rows retain distances and references."""
    raw = json.dumps(document, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()
    ref = "sha256:" + hashlib.sha256(raw).hexdigest()
    documents.setdefault(ref, document)
    return ref


def run(binary: Path, output: Path) -> dict:
    data, files, identity = baseline.load_inputs()
    committed = baseline.baseline()
    # Verify the frozen baseline before attributing any result to an arm.
    baseline.check(baseline.build(data, files, identity), committed)
    output.mkdir(parents=True, exist_ok=True)
    with TemporaryDirectory(prefix="moai-c10b-") as tmp:
        home = Path(tmp) / "home"
        home.mkdir()
        root = Path(tmp) / "project"
        root.mkdir()
        env = {k: os.environ[k] for k in
               ("PATH", "TMPDIR", "TEMP", "TMP", "SystemRoot", "SystemDrive", "COMSPEC", "PATHEXT")
               if k in os.environ}
        env.update(HOME=str(home), USERPROFILE=str(home), XDG_CONFIG_HOME=str(home / "config"),
                   XDG_CACHE_HOME=str(home / "cache"), XDG_DATA_HOME=str(home / "data"),
                   GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_SYSTEM=os.devnull,
                   GIT_CONFIG_NOSYSTEM="1", GIT_TERMINAL_PROMPT="0")
        for name, content in files.items():
            path = root / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(content)
        def git(*args):
            subprocess.run(["git", "-c", "core.hooksPath=" + os.devnull,
                            "-c", "user.name=fixture", "-c", "user.email=fixture@example.invalid",
                            *args], cwd=root, env=env, check=True, capture_output=True)
        git("init", "-q")
        git("add", "--", "src")
        git("commit", "-q", "-m", "frozen snapshot")

        def op(name, query):
            result = subprocess.run([str(binary), name, str(root)],
                                    input=json.dumps({"protocol_version": 1, **query}).encode(),
                                    env=env, capture_output=True, check=True)
            value = json.loads(result.stdout)
            return value.get("result", value), len(result.stdout)

        built, _ = op("index-build", {})
        if built.get("status") != "complete":
            raise ValueError("frozen snapshot index build incomplete")
        arms = []
        documents = {}
        generation = None
        scored = [t for t in data["tasks"] if t["kind"] != "freshness"]
        for arm, operations in [("rg", ["search"]), ("+Map", ["search", "atlas-map"]),
                                ("+Graph", ["search", "atlas-map", "atlas-graph"])]:
            rows = []
            counts = {name: 0 for name in operations}
            total_bytes = 0
            state_counts = {}
            empty_seed_skips = 0
            edge_count = 0
            for task in scored:
                hits = set()
                task_bytes = 0
                contexts = {}
                mapped = {"nodes": [], "edges": []}
                selection = None
                for name in operations:
                    if name == "search":
                        query = {"query": {"literals": [task["pattern"]]},
                                 "limits": {"max_files": 2000, "max_spans_per_file": 1000}}
                    else:
                        # Empty paths means ALL files in atlas-map: never accidentally
                        # turn a zero-hit lexical task into an oracle-free full scan.
                        if not hits:
                            empty_seed_skips += 1
                            continue
                        if len(hits) > 64:
                            raise ValueError("seed cap exceeded; do not silently truncate")
                        query = ({"query": {"paths": sorted(hits)}, "limits": {"max_files": 2000}}
                                 if name == "atlas-map" else
                                 {"query": {"start": ["file:" + p for p in sorted(hits)]},
                                  "limits": {"max_hops": 3, "max_edges": 5000}})
                    value, size = op(name, query)
                    counts[name] += 1
                    task_bytes += size
                    if value.get("omitted") != []:
                        raise ValueError(f"{arm}/{task['id']}/{name}: omitted results")
                    gid = (value.get("coverage", {}).get("generation_id")
                           if name == "search" else value.get("generation_id"))
                    if not gid:
                        raise ValueError("missing generation identity")
                    generation = generation or gid
                    if gid != generation:
                        raise ValueError("arms observed different generations")
                    if name == "search":
                        if value["coverage"]["file_source"] != "index_generation":
                            raise ValueError("lexical search escaped the generation")
                        for hit in value["results"]:
                            state = hit["state"]
                            state_counts[state] = state_counts.get(state, 0) + 1
                            if state != "fresh":
                                raise ValueError("snapshot result is not fresh")
                            hits.add(hit["path"])
                    else:
                        if value["state"] != "fresh":
                            raise ValueError(f"{name}: unavailable/stale/partial")
                        doc = value["map" if name == "atlas-map" else "graph"]
                        edge_count += len(doc["edges"])
                        raw_files = sorted(n["id"][5:] for n in doc["nodes"]
                                           if n["id"].startswith("file:"))
                        if not set(raw_files) <= files.keys():
                            raise ValueError("raw context escaped the fixed source scope")
                        contexts[name] = {"document": context_document(documents, doc),
                                          "files": raw_files}
                        if name == "atlas-map":
                            mapped = doc
                            selection = select_answers(task["pattern"], mapped, None, files)
                        else:
                            contexts[name]["distance"] = value["distance"]
                            selection = select_answers(task["pattern"], mapped, doc, files)
                        hits = set(selection["hits"])
                    if not hits <= files.keys():
                        raise ValueError("hit escaped the fixed source scope")
                total_bytes += task_bytes
                rows.append({"id": task["id"], "hits": sorted(hits), "response_bytes": task_bytes,
                             "raw_context": contexts, "selection": selection})
            report = {"schema_version": committed["schema_version"], **identity, "tasks": rows}
            if arm == "rg":
                expected = {t["id"]: t["hits"] for t in committed["tasks"]}
                if any(r["hits"] != expected[r["id"]] for r in rows):
                    raise ValueError("lexical arm differs from pinned rg baseline")
            judged = baseline.judge(report, data, files, committed)
            (output / (arm.replace("+", "") + ".json")).write_text(
                json.dumps({"report": report, "judge": judged}, indent=2) + "\n")
            arms.append({"arm": arm, "tasks": len(rows), **judged,
                         "operation_calls": counts, "tool_calls": sum(counts.values()),
                         "empty_seed_skips": empty_seed_skips,
                         "response_bytes": total_bytes, "tokens": "not_run: bytes are not tokens",
                         "citations": sum(len(r["hits"]) for r in rows),
                          "raw_file_citations": sum(len(r["raw_context"].get(
                              "atlas-graph", r["raw_context"].get("atlas-map", {"files": r["hits"]}))[
                                  "files"]) for r in rows),
                         "snapshot_stale_ratio": 0.0, "search_span_states": state_counts,
                         "view_edges_returned": edge_count,
                         "freshness_mutations": "not_run: #2407"})
        summary = {"schema_version": 1, **identity, "generation_id": generation,
                   "index_build_status": built["status"], "arms": arms,
                   "selection_policy": "literal-evidence-v1", "raw_context_file": "raw-contexts.json",
                   "semantic_graph": "not_measured: file coverage is not relation correctness"}
        (output / "raw-contexts.json").write_text(
            json.dumps({"schema_version": 1, "generation_id": generation, "documents": documents},
                       indent=2) + "\n")
        (output / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
        return summary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    args = parser.parse_args()
    result = run(args.binary.resolve(), args.output_dir.resolve())
    print(json.dumps(result, indent=2))
    return 0 if all(a["passed"] for a in result["arms"]) else 1


if __name__ == "__main__":
    raise SystemExit(main())
