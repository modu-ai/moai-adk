#! /usr/bin/env python3
"""C10a corpus/baseline consistency tests (Issue #2346, AC-2/AC-3).

The Rust side validates the fixture format; this side validates that the
committed baseline.json still matches a fresh rg-only run — thresholds are
pinned to the baseline, so loosening one without re-running fails here.
"""

import json
import os
import shlex
import shutil
import subprocess
import sys
import tempfile
import unittest
from contextlib import contextmanager
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
RUNNER = ROOT / "tools" / "atlas_baseline.py"


class AtlasBaselineConsistency(unittest.TestCase):
    def test_committed_baseline_matches_a_fresh_run(self):
        out = subprocess.run(
            [sys.executable, "-I", str(RUNNER)],
            capture_output=True,
            text=True,
            cwd=ROOT,
        )
        self.assertEqual(out.returncode, 0, out.stdout + out.stderr)

    @contextmanager
    def project(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / "tools").mkdir()
            shutil.copy2(RUNNER, root / "tools" / RUNNER.name)
            shutil.copytree(
                ROOT / "tests" / "fixtures" / "atlas",
                root / "tests" / "fixtures" / "atlas",
            )
            yield root

    def invoke(self, root, *arguments, environment=None):
        return subprocess.run(
            [sys.executable, "-I", str(root / "tools" / RUNNER.name), *arguments],
            cwd=root, env=environment, capture_output=True, text=True,
        )

    def artifact(self, root):
        return json.loads((root / "tests/fixtures/atlas/baseline.json").read_text(encoding="utf-8"))

    def candidate(self, root):
        artifact = self.artifact(root)
        return {
            "schema_version": artifact["schema_version"],
            "corpus_sha256": artifact["corpus_sha256"],
            "source_snapshot_sha256": artifact["source_snapshot_sha256"],
            "tasks": [{"id": row["id"], "hits": list(row["hits"])} for row in artifact["tasks"]],
        }

    def judge(self, root, candidate):
        path = root / "candidate.json"
        path.write_text(json.dumps(candidate), encoding="utf-8")
        out = self.invoke(root, "--judge", str(path))
        return out, json.loads(out.stdout)

    def test_live_source_add_modify_delete_cannot_change_frozen_measurements(self):
        with self.project() as root:
            initial = self.invoke(root, "--build")
            self.assertEqual(initial.returncode, 0, initial.stdout + initial.stderr)
            baseline = root / "tests/fixtures/atlas/baseline.json"
            original = baseline.read_bytes()
            missed = next(row for row in self.artifact(root)["tasks"] if row["id"] == "T-033")
            self.assertEqual((missed["precision"], missed["recall"]), (0.0, 0.0))
            live = root / "src" / "lib.rs"
            live.parent.mkdir()
            corpus = json.loads((root / "tests/fixtures/atlas/corpus.json").read_text(encoding="utf-8"))
            patterns = "\n".join(t["pattern"] for t in corpus["tasks"] if t["kind"] != "freshness")
            for operation in ("add", "modify", "delete"):
                with self.subTest(operation=operation):
                    if operation == "add":
                        live.write_text(patterns, encoding="utf-8")
                    elif operation == "modify":
                        live.write_text("unrelated replacement", encoding="utf-8")
                    else:
                        live.unlink()
                    for arguments in ((), ("--build",)):
                        out = self.invoke(root, *arguments)
                        self.assertEqual(out.returncode, 0, out.stdout + out.stderr)
                        self.assertEqual(baseline.read_bytes(), original)

    def test_ground_truth_outside_snapshot_is_invalid_even_if_live_file_exists(self):
        with self.project() as root:
            fixtures = root / "tests/fixtures/atlas"
            original = (fixtures / "baseline.json").read_bytes()
            corpus = json.loads((fixtures / "corpus.json").read_text(encoding="utf-8"))
            corpus["tasks"][0]["ground_truth"]["files"] = ["src/not_in_snapshot.rs"]
            live = root / "src" / "not_in_snapshot.rs"
            live.parent.mkdir()
            live.write_text(corpus["tasks"][0]["pattern"], encoding="utf-8")
            (fixtures / "corpus.json").write_text(json.dumps(corpus), encoding="utf-8")
            out = self.invoke(root, "--build")
            self.assertNotEqual(out.returncode, 0, out.stdout + out.stderr)
            self.assertEqual((fixtures / "baseline.json").read_bytes(), original)

    def test_build_rejects_twenty_nine_but_accepts_exactly_thirty_real_tasks(self):
        with self.project() as root:
            fixtures = root / "tests/fixtures/atlas"
            corpus = json.loads((fixtures / "corpus.json").read_text(encoding="utf-8"))
            original = (fixtures / "baseline.json").read_bytes()
            tasks = [
                next(t for t in corpus["tasks"] if t["kind"] == kind)
                for kind in ("retrieval", "graph", "freshness")
            ]
            selected = {t["id"] for t in tasks}
            tasks += [t for t in corpus["tasks"] if t["id"] not in selected][:27]
            self.assertEqual(len(tasks), 30)
            corpus["tasks"] = tasks[1:]
            (fixtures / "corpus.json").write_text(json.dumps(corpus), encoding="utf-8")
            rejected = self.invoke(root, "--build")
            self.assertNotEqual(rejected.returncode, 0, rejected.stdout + rejected.stderr)
            self.assertEqual((fixtures / "baseline.json").read_bytes(), original)
            corpus["tasks"] = tasks
            (fixtures / "corpus.json").write_text(json.dumps(corpus), encoding="utf-8")
            accepted = self.invoke(root, "--build")
            self.assertEqual(accepted.returncode, 0, accepted.stdout + accepted.stderr)
            checked = self.invoke(root)
            self.assertEqual(checked.returncode, 0, checked.stdout + checked.stderr)

    def test_changed_floor_or_reported_minimum_is_not_a_consistent_baseline(self):
        with self.project() as root:
            path = root / "tests/fixtures/atlas/baseline.json"
            artifact = self.artifact(root)
            for group, key, delta in (
                ("thresholds", "retrieval_precision_mean", -0.0001),
                ("thresholds", "retrieval_precision_mean", 0.0001),
                ("diagnostics", "retrieval_precision_min", 0.0001),
            ):
                with self.subTest(group=group, delta=delta):
                    changed = json.loads(json.dumps(artifact))
                    changed[group][key] += delta
                    path.write_text(json.dumps(changed), encoding="utf-8")
                    before = path.read_bytes()
                    out = self.invoke(root)
                    self.assertNotEqual(out.returncode, 0, out.stdout + out.stderr)
                    self.assertEqual(path.read_bytes(), before)

    def test_judge_accepts_equality_and_rejects_each_binding_metric_regression(self):
        with self.project() as root:
            out, result = self.judge(root, self.candidate(root))
            self.assertEqual(out.returncode, 0, out.stdout + out.stderr)
            self.assertTrue(result["passed"])
            self.assertEqual(result["metrics"], self.artifact(root)["thresholds"])
            for metric in (
                "retrieval_precision_mean", "retrieval_recall_mean",
                "graph_relation_file_coverage_mean",
            ):
                with self.subTest(metric=metric):
                    candidate = self.candidate(root)
                    rows = {r["id"]: r for r in candidate["tasks"]}
                    if metric == "retrieval_precision_mean":
                        rows["T-001"]["hits"].append("src/ontology.rs")
                    elif metric == "retrieval_recall_mean":
                        rows["T-001"]["hits"].pop()
                    else:
                        rows["G-001"]["hits"].remove("src/archive.rs")
                    out, result = self.judge(root, candidate)
                    self.assertEqual(out.returncode, 1, out.stdout + out.stderr)
                    self.assertFalse(result["passed"])
                    self.assertEqual(result["failed_metrics"], [metric])
                    self.assertLess(result["metrics"][metric], result["thresholds"][metric])

    def test_mean_floors_allow_compensating_improvement_not_per_task_floors(self):
        with self.project() as root:
            candidate = self.candidate(root)
            rows = {r["id"]: r for r in candidate["tasks"]}
            rows["T-001"]["hits"].append("src/ontology.rs")
            rows["T-007"]["hits"] = ["src/contract.rs"]
            out, result = self.judge(root, candidate)
            self.assertEqual(out.returncode, 0, out.stdout + out.stderr)
            self.assertTrue(result["passed"])
            self.assertGreater(
                result["metrics"]["retrieval_precision_mean"],
                result["thresholds"]["retrieval_precision_mean"],
            )

    def test_frozen_snapshot_digest_is_checked_before_baseline_write(self):
        with self.project() as root:
            fixtures = root / "tests/fixtures/atlas"
            baseline = fixtures / "baseline.json"
            original = baseline.read_bytes()
            snapshot = fixtures / "source_snapshot.json"
            snapshot.write_bytes(snapshot.read_bytes() + b"\n")
            out = self.invoke(root, "--build")
            self.assertNotEqual(out.returncode, 0, out.stdout + out.stderr)
            self.assertEqual(baseline.read_bytes(), original)

    def test_partial_search_error_fails_check_and_build_without_writing_baseline(self):
        real_rg = shutil.which("rg")
        self.assertIsNotNone(real_rg, "rg is required; missing rg must not skip this test")
        with self.project() as root:
            fixtures = root / "tests" / "fixtures" / "atlas"
            runner = root / "tools" / RUNNER.name
            initial = self.invoke(root, "--build")
            self.assertEqual(initial.returncode, 0, initial.stdout + initial.stderr)
            baseline = fixtures / "baseline.json"
            original = baseline.read_bytes()
            fake_bin = root / "bin"
            fake_bin.mkdir()
            driver = fake_bin / "rg_error.py"
            diagnostic = "fixture source could not be read"
            driver.write_text(
                "import subprocess, sys\n"
                f"out = subprocess.run([{real_rg!r}, *sys.argv[1:]], capture_output=True)\n"
                "sys.stdout.buffer.write(out.stdout)\n"
                "sys.stderr.buffer.write(out.stderr)\n"
                "if '--version' in sys.argv[1:]:\n"
                "    raise SystemExit(out.returncode)\n"
                f"sys.stderr.write({diagnostic!r} + '\\n')\n"
                "raise SystemExit(2)\n",
                encoding="utf-8",
            )
            wrapper = fake_bin / "rg"
            wrapper.write_text(
                "#!/bin/sh\nexec "
                + shlex.quote(sys.executable) + " -I "
                + shlex.quote(str(driver)) + ' "$@"\n',
                encoding="utf-8",
            )
            wrapper.chmod(0o755)
            environment = dict(os.environ, PATH=str(fake_bin) + os.pathsep + os.environ.get("PATH", ""))
            for arguments in ([], ["--build"]):
                with self.subTest(arguments=arguments):
                    out = subprocess.run(
                        [sys.executable, "-I", str(runner), *arguments],
                        cwd=root, env=environment, capture_output=True, text=True,
                    )
                    self.assertNotEqual(out.returncode, 0, out.stdout + out.stderr)
                    self.assertIn(diagnostic, out.stderr)
                    self.assertEqual(baseline.read_bytes(), original)


if __name__ == "__main__":
    unittest.main()
