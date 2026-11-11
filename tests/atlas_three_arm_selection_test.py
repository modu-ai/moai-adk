"""Graph evidence must change answer selection, independently of the fixed C10a oracle."""
import copy
import json
from pathlib import Path
import sys
import unittest
from unittest import mock
from tempfile import TemporaryDirectory

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "tools"))
import atlas_three_arm as runner


def node(path):
    return {"id": "file:" + path}


def edge(name, source, target, evidence=True):
    return {"edge_id": name, "source_id": "file:" + source,
            "target_id": "file:" + target, "kind": "imports",
            "evidence_refs": ["file:" + source + "#L1@sha256:synthetic"] if evidence else []}


class AnswerSelectionTest(unittest.TestCase):
    def setUp(self):
        self.files = {"seed.rs": b"needle seed", "related.rs": b"needle related",
                      "noise.rs": b"other context", "isolated.rs": b"needle isolated"}
        self.map = {"nodes": [node("seed.rs")], "edges": []}
        self.graph = {"nodes": [node(p) for p in self.files],
                      "edges": [edge("related-edge", "seed.rs", "related.rs"),
                                edge("noise-edge", "seed.rs", "noise.rs")]}

    def choose(self, pattern="needle", graph=None):
        return runner.select_answers(pattern, self.map,
                                     self.graph if graph is None else graph, self.files)

    def test_graph_only_relevant_file_is_an_answer_with_its_edge(self):
        selected = self.choose()
        self.assertEqual(selected["hits"], ["related.rs", "seed.rs"])
        self.assertEqual(selected["support"]["related.rs"], ["related-edge"])
        self.assertNotIn("isolated.rs", selected["eligible"])

    def test_removing_edge_or_its_evidence_changes_answers(self):
        for kind in ("remove", "no_evidence"):
            with self.subTest(kind=kind):
                graph = copy.deepcopy(self.graph)
                if kind == "remove":
                    graph["edges"] = graph["edges"][1:]
                else:
                    graph["edges"][0]["evidence_refs"] = []
                self.assertEqual(self.choose(graph=graph)["hits"], ["seed.rs"])

    def test_query_controls_relevance_without_any_ground_truth_input(self):
        selected = self.choose("other")
        self.assertEqual(selected["hits"], ["noise.rs"])
        self.assertEqual(selected["support"]["noise.rs"], ["noise-edge"])

    def test_map_evidence_is_required_to_start_graph_selection(self):
        self.assertEqual(runner.select_answers("needle", {"nodes": [], "edges": []},
                                               self.graph, self.files)["hits"], [])

    def test_raw_context_survives_relevance_selection_unchanged(self):
        before = copy.deepcopy(self.graph)
        self.choose()
        self.assertEqual(self.graph, before)
        self.assertIn(node("noise.rs"), self.graph["nodes"])
        self.assertEqual(len(self.graph["edges"]), 2)

    def test_separate_context_document_preserves_every_node_and_edge(self):
        documents = {}
        ref = runner.context_document(documents, self.graph)
        self.assertEqual(documents[ref], self.graph)
        self.assertEqual(runner.context_document(documents, copy.deepcopy(self.graph)), ref)
        self.assertEqual(len(documents), 1)

    def test_fourth_hop_is_context_not_an_eligible_answer(self):
        files = {"s.rs": b"needle", "a.rs": b"a", "b.rs": b"b",
                 "c.rs": b"c", "d.rs": b"needle"}
        graph = {"nodes": [node(p) for p in files],
                 "edges": [edge(str(i), a, b) for i, (a, b) in enumerate(
                     [("s.rs", "a.rs"), ("a.rs", "b.rs"), ("b.rs", "c.rs"), ("c.rs", "d.rs")])]}
        selected = runner.select_answers("needle", {"nodes": [node("s.rs")]}, graph, files)
        self.assertEqual(selected["hits"], ["s.rs"])
        self.assertNotIn("d.rs", selected["eligible"])

    def test_unknown_raw_file_is_refused_even_when_it_is_not_relevant(self):
        self.graph["nodes"].append(node("outside.rs"))
        with self.assertRaises(ValueError):
            self.choose()


class RunnerPipelineTest(unittest.TestCase):
    def test_graph_result_reaches_answers_through_the_real_runner_call_site(self):
        # The adapter deliberately returns a restricted seed set. No oracle is supplied
        # to the runner: the judge is replaced, not the query/selection functions.
        files = {"src/seed.rs": b"needle seed", "src/related.rs": b"needle neighbor",
                 "src/noise.rs": b"unrelated context"}
        data = {"tasks": [{"id": "query", "kind": "retrieval", "pattern": "needle"}]}
        committed = {"schema_version": 2,
                     "tasks": [{"id": "query", "hits": ["src/seed.rs"]}]}
        graph = {"nodes": [node(p) for p in files],
                 "edges": [edge("neighbor", "src/seed.rs", "src/related.rs"),
                           edge("noise", "src/seed.rs", "src/noise.rs")]}
        native_run = runner.subprocess.run

        def adapter(argv, **kwargs):
            if str(argv[0]) != "/synthetic/test-binary":
                return native_run(argv, **kwargs)
            name = argv[1]
            if name == "index-build":
                value = {"status": "complete"}
            elif name == "search":
                value = {"coverage": {"generation_id": "g", "file_source": "index_generation"},
                         "results": [{"path": "src/seed.rs", "state": "fresh"}], "omitted": []}
            else:
                value = {"generation_id": "g", "state": "fresh", "omitted": []}
                if name == "atlas-map":
                    value["map"] = {"nodes": [node("src/seed.rs")], "edges": []}
                else:
                    self.assertEqual(name, "atlas-graph")
                    query = json.loads(kwargs["input"])
                    self.assertEqual(query["query"]["start"], ["file:src/seed.rs"])
                    value.update(graph=graph, distance={"file:src/seed.rs": 0,
                                 "file:src/related.rs": 1, "file:src/noise.rs": 1})
            return runner.subprocess.CompletedProcess(argv, 0, json.dumps(value).encode(), b"")

        with TemporaryDirectory(prefix="moai-c10b-selection-") as tmp, \
                mock.patch.object(runner.baseline, "load_inputs", return_value=(data, files, {})), \
                mock.patch.object(runner.baseline, "baseline", return_value=committed), \
                mock.patch.object(runner.baseline, "build"), \
                mock.patch.object(runner.baseline, "check"), \
                mock.patch.object(runner.baseline, "judge", return_value={"passed": True}), \
                mock.patch.object(runner.subprocess, "run", side_effect=adapter):
            output = Path(tmp) / "results"
            result = runner.run(Path("/synthetic/test-binary"), output)
            report = json.loads((output / "Graph.json").read_text())
            task = report["report"]["tasks"][0]
            self.assertEqual(task["hits"], ["src/related.rs", "src/seed.rs"])
            self.assertEqual(task["selection"]["support"]["src/related.rs"], ["neighbor"])
            self.assertEqual(result["arms"][2]["operation_calls"]["atlas-graph"], 1)
            contexts = json.loads((output / "raw-contexts.json").read_text())
            raw = contexts["documents"][task["raw_context"]["atlas-graph"]["document"]]
            self.assertEqual(raw, graph)
            self.assertIn(node("src/noise.rs"), raw["nodes"])


if __name__ == "__main__":
    unittest.main()
