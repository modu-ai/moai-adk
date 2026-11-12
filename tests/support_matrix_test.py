#!/usr/bin/env python3
"""AC-08D-1 checker fixtures (Issue #2015): a supported row without a passed
receipt fails, not_run rows are allowed, and the documents carry no
secret-shaped or absolute-home strings. The checker itself runs as a
subprocess exactly like the AC command does."""

import json
import pathlib
import subprocess
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
CHECKER = ROOT / "tools" / "check_support_matrix.py"

GOOD_RECEIPT = {
    "ac": "AC-08C-1",
    "os": "Darwin",
    "arch": "arm64",
    "host": "none",
    "host_exact_version": None,
    "mode": "local",
    "skill_digest": "b" * 64,
    "moai_cli_sha": "3" * 64,
    "result": "passed",
    "ran_at": "2026-11-08T14:55:34Z",
}


def run_checker(matrix: str, receipts: list) -> int:
    with tempfile.TemporaryDirectory() as work:
        work = pathlib.Path(work)
        matrix_path = work / "support-matrix.md"
        matrix_path.write_text(matrix, encoding="utf-8")
        receipt_dir = work / "e2e-receipts"
        receipt_dir.mkdir()
        with open(receipt_dir / "receipts.jsonl", "w", encoding="utf-8") as handle:
            for record in receipts:
                handle.write(json.dumps(record) + "\n")
        return subprocess.run(
            ["python3", "-I", str(CHECKER), str(matrix_path), str(receipt_dir)],
            capture_output=True,
            text=True,
        ).returncode


class SupportMatrixCheckerTest(unittest.TestCase):
    def test_supported_row_with_passed_receipt_passes(self):
        matrix = (
            "| OS/arch | host | AC | 결과 | receipt |\n"
            "| --- | --- | --- | --- | --- |\n"
            "| Darwin/arm64 | none | AC-08C-1 | passed | receipts.jsonl |\n"
        )
        self.assertEqual(run_checker(matrix, [GOOD_RECEIPT]), 0)

    def test_supported_row_without_receipt_fails(self):
        matrix = (
            "| OS/arch | host | AC | 결과 | receipt |\n"
            "| --- | --- | --- | --- | --- |\n"
            "| Darwin/arm64 | none | AC-08C-1 | passed | receipts.jsonl |\n"
        )
        # No receipt file at all: the support claim is unsupported (RL-10).
        self.assertNotEqual(run_checker(matrix, []), 0)

    def test_not_run_row_without_receipt_is_allowed(self):
        matrix = (
            "| OS/arch | host | AC | 결과 | 사유 |\n"
            "| --- | --- | --- | --- | --- |\n"
            "| ubuntu | codex | AC-08C-4 | not_run | 운영자 실측 대기 |\n"
        )
        self.assertEqual(run_checker(matrix, []), 0)

    def test_row_count_must_equal_passed_receipts(self):
        matrix = (
            "| OS/arch | host | AC | 결과 | receipt |\n"
            "| --- | --- | --- | --- | --- |\n"
            "| Darwin/arm64 | none | AC-08C-1 | passed | receipts.jsonl |\n"
            "| Darwin/arm64 | none | AC-08C-2 | passed | receipts.jsonl |\n"
        )
        receipts = [
            GOOD_RECEIPT,
            {**GOOD_RECEIPT, "ac": "AC-08C-2", "result": "passed"},
            {**GOOD_RECEIPT, "ac": "AC-08C-3", "result": "passed"},
        ]
        # Two supported rows against three passed receipts: the mismatch fails.
        self.assertNotEqual(run_checker(matrix, receipts), 0)

    def test_real_documents_have_no_secret_shapes(self):
        # AC-08D-6: the three release documents carry no credential shapes,
        # absolute home paths, or private operation details.
        pattern = r"(ghp_[A-Za-z0-9]|github_pat_|sk-[A-Za-z0-9]|-----BEGIN|/Users/|C:\\\\)"
        docs = [
            ROOT / "docs" / "support-matrix.md",
            ROOT / "docs" / "release-checklist.md",
            ROOT / "docs" / "release-review.md",
        ]
        import re

        for doc in docs:
            self.assertTrue(doc.is_file(), doc)
            self.assertEqual(re.search(pattern, doc.read_text(encoding="utf-8")), None, doc)


if __name__ == "__main__":
    unittest.main()
