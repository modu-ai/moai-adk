"""SPEC-20261118-omp-host: the host-keyed feature survey in docs/host-support-matrix.md is pinned
against the code.

Each row of the survey names one place in the code that is keyed by host and says, for claude,
codex and omp, whether that place supports the host. A row whose evidence is a host list (a
`const` array) must agree with the hosts the array contains; a row marked `gap` for omp must say
which slice or card closes it. The test reads the document and the source files; it starts no
host.
"""

import pathlib
import re
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
DOC = ROOT / "docs" / "host-support-matrix.md"
HEADING = "## MoAI 기능별 host 지원 (C-62)"
COLUMNS = ["feature", "claude", "codex", "omp", "evidence", "gap"]
STATUSES = {"yes", "gap", "n/a"}
HOSTS = ("claude", "codex", "omp")
LIST_MARK = " (list)"


def split_row(line):
    cells = re.split(r"(?<!\\)\|", line.strip())
    return [c.strip().replace("\\|", "|") for c in cells[1:-1]]


def survey_rows():
    text = DOC.read_text(encoding="utf-8")
    start = text.index(HEADING)
    lines = text[start:].splitlines()[1:]
    header, rows = None, []
    for line in lines:
        if line.startswith("## "):
            break
        if line.startswith("| feature |"):
            header = split_row(line)
        elif header and line.startswith("|") and not line.startswith("| ---"):
            rows.append(dict(zip(header, split_row(line))))
    return header, rows


def split_evidence(evidence):
    """`path::symbol` or `path::symbol (list)`; the symbol is a name in the file."""
    is_list = evidence.endswith(LIST_MARK)
    body = evidence[: -len(LIST_MARK)] if is_list else evidence
    path, _, symbol = body.strip("` ").partition("::")
    return path, symbol, is_list


def hosts_in_list(source, symbol):
    """The host names (claude, codex, omp) inside the array assigned to `symbol`."""
    match = re.search(rf"\b{re.escape(symbol)}\b[^=]*=\s*&?\[(.*?)\];", source, re.DOTALL)
    if not match:
        return None
    return {name for name in re.findall(r'"([a-z]+)"', match.group(1)) if name in HOSTS}


class OmpHostSurveyTest(unittest.TestCase):
    def setUp(self):
        self.header, self.rows = survey_rows()

    def test_the_survey_has_the_columns_and_rows(self):
        self.assertEqual(self.header, COLUMNS)
        self.assertGreaterEqual(len(self.rows), 12, "one row per host-keyed place")
        features = [r["feature"] for r in self.rows]
        self.assertEqual(len(features), len(set(features)), "one row per feature")

    def test_every_status_is_in_the_vocabulary(self):
        for r in self.rows:
            with self.subTest(feature=r["feature"]):
                for host in HOSTS:
                    self.assertIn(r[host], STATUSES)

    def test_the_evidence_file_and_symbol_exist(self):
        for r in self.rows:
            with self.subTest(feature=r["feature"]):
                path, symbol, _ = split_evidence(r["evidence"])
                self.assertTrue(path and symbol, r["evidence"])
                file = ROOT / path
                self.assertTrue(file.is_file(), path)
                self.assertIn(symbol, file.read_text(encoding="utf-8"), f"{symbol} not in {path}")

    def test_a_list_row_agrees_with_the_hosts_in_the_code(self):
        lists = 0
        for r in self.rows:
            path, symbol, is_list = split_evidence(r["evidence"])
            if not is_list:
                continue
            lists += 1
            with self.subTest(feature=r["feature"]):
                found = hosts_in_list((ROOT / path).read_text(encoding="utf-8"), symbol)
                self.assertIsNotNone(found, f"no array assigned to {symbol} in {path}")
                claimed = {h for h in HOSTS if r[h] == "yes"}
                self.assertEqual(claimed, found, f"{path}::{symbol}")
        self.assertGreaterEqual(lists, 6, "the survey pins at least six host lists")

    def test_an_omp_gap_names_the_slice_or_card_that_closes_it(self):
        gaps = [r for r in self.rows if r["omp"] == "gap"]
        self.assertTrue(gaps, "the launcher is still a gap until slice B")
        for r in gaps:
            with self.subTest(feature=r["feature"]):
                self.assertRegex(r["gap"], r"slice B|card C-\d+|leader card")

    def test_a_row_without_a_gap_has_no_gap_text(self):
        for r in self.rows:
            if r["omp"] != "gap":
                with self.subTest(feature=r["feature"]):
                    self.assertTrue(r["gap"] in ("", "-") or r["omp"] == "n/a", r["gap"])

    def test_the_document_replaces_the_2378_unsupported_line(self):
        text = DOC.read_text(encoding="utf-8")
        self.assertNotIn("OMP — 지원하지 않음", text)
        self.assertIn("2026-11-17", text)
        self.assertIn("#2378", text)


if __name__ == "__main__":
    unittest.main()
