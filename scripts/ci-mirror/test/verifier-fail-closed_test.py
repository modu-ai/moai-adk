#!/usr/bin/env python3
"""Execute the shipped gate shell and validator against bounded CI fixtures."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[3]
YQ = shutil.which("yq")
if not YQ:
    raise RuntimeError("yq is required; these verifier regressions must execute")


class VerifierFailClosed(unittest.TestCase):
    def gate(self, detect, matrix, code):
        script = subprocess.check_output(
            [YQ, "-r", '.jobs.release-pr-gate.steps[0].run',
             str(ROOT / ".github/workflows/release-pr-multi-os.yml")], text=True)
        for key, value in (("needs.detect-release.result", detect),
                           ("needs.full-matrix-test.result", matrix),
                           ("needs.detect-release.outputs.go_code", code)):
            script = script.replace("${{ " + key + " }}", value)
        with tempfile.TemporaryDirectory() as directory:
            env = dict(os.environ, GITHUB_OUTPUT=str(Path(directory) / "output"))
            return subprocess.run(["bash", "-e", "-c", script], env=env,
                                  capture_output=True, text=True, timeout=10)

    def test_release_gate(self):
        cases = [("success", "success", "true", True),
                 ("success", "success", "", True),
                 ("skipped", "skipped", "", True),
                 ("success", "skipped", "false", True),
                 ("cancelled", "skipped", "", False),
                 ("cancelled", "success", "true", False),
                 ("failure", "skipped", "false", False),
                 ("unknown", "skipped", "false", False),
                 ("success", "skipped", "", False),
                 ("success", "skipped", "malformed", False),
                 ("success", "cancelled", "false", False),
                 ("success", "failure", "true", False)]
        for detect, matrix, code, passed in cases:
            with self.subTest(detect=detect, matrix=matrix, code=code):
                result = self.gate(detect, matrix, code)
                self.assertEqual(result.returncode == 0, passed,
                                 result.stdout + result.stderr)

    def validator(self, branches, fail_query=False, auxiliary="[]"):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            workflows = root / ".github/workflows"
            workflows.mkdir(parents=True)
            (workflows / "Lint.yml").write_text(
                "name: CI\non: push\njobs:\n  lint:\n    name: Lint\n    runs-on: ubuntu-latest\n    steps:\n      - run: true\n")
            (root / ".github/required-checks.yml").write_text(
                "branches: " + branches + "\nauxiliary: " + auxiliary + "\n")
            env = dict(os.environ)
            if fail_query:
                wrapper = root / "bin/yq"
                wrapper.parent.mkdir()
                wrapper.write_text('#!/bin/sh\nif [ "$1" = "-r" ]; then\ncase "$2" in\n*"$FAIL_QUERY"*) exit 42;;\nesac\nfi\nexec "' + YQ + '" "$@"\n')
                wrapper.chmod(0o755)
                env["FAIL_QUERY"] = fail_query
                env["PATH"] = str(wrapper.parent) + os.pathsep + env["PATH"]
            return subprocess.run(["sh", str(ROOT / "scripts/ci-mirror/validate-required-checks.sh")],
                                  cwd=root, env=env, capture_output=True, text=True, timeout=10)

    def test_required_context_structure(self):
        cases = [("{main: {contexts: [Lint]}, 'release/*': {contexts: []}}", True),
                 ("{main: {contexts: [Phantom]}, 'release/*': {contexts: []}}", False),
                 ("{main: {}, 'release/*': {contexts: []}}", False),
                 ("{main: {contexts: null}, 'release/*': {contexts: []}}", False),
                 ("{main: {contexts: {}}, 'release/*': {contexts: []}}", False),
                 ("{main: {contexts: [null]}, 'release/*': {contexts: []}}", False),
                 ("{main: {contexts: ['']}, 'release/*': {contexts: []}}", False),
                 ("[]", False), ("null", False), ("{}", False),
                 ("{main: {contexts: []}}", False)]
        for branches, passed in cases:
            with self.subTest(branches=branches):
                result = self.validator(branches)
                self.assertEqual(result.returncode == 0, passed,
                                 result.stdout + result.stderr)

    def test_required_context_extraction_failure(self):
        for query in ("keys", "contexts", "auxiliary"):
            with self.subTest(query=query):
                result = self.validator("{main: {contexts: [Lint]}, 'release/*': {contexts: []}}", query)
                self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_auxiliary_overlap_still_fails(self):
        for branch in ("main", "release/*"):
            with self.subTest(branch=branch):
                contexts = "{main: {contexts: [Lint]}, 'release/*': {contexts: []}}" if branch == "main" else "{main: {contexts: []}, 'release/*': {contexts: [Lint]}}"
                result = self.validator(contexts, auxiliary="[Lint]")
                self.assertNotEqual(result.returncode, 0, result.stdout + result.stderr)
                self.assertIn("found in branches.", result.stderr)


if __name__ == "__main__":
    unittest.main(verbosity=2)
