"""Offline checks for the ZCode rows of docs/host-support-matrix.md (02-D #1994).

The matrix records what official sources say; these tests keep its shape honest. They do
not re-fetch the sources (an independent audit does) and do not prove any host behaviour.
"""
import pathlib
import re
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
MATRIX = ROOT / "docs/host-support-matrix.md"
SUPPORT = ROOT / "docs/zcode-support.md"
COLUMNS = ["host", "surface", "item", "capability_status", "unknown_class", "내용", "source"]
STATUSES = {"unknown", "documented", "probed", "certified", "stale"}
UNKNOWN_CLASSES = {"표시 불가", "대안", "시험 대기"}
# REQ-02D-02: the nine items every ZCode surface must carry.
REQUIRED_ITEMS = ["agents_loading", "skill_path", "skill_limits", "explicit_invocation", "import_dialog",
                  "hooks", "plugin_manifest", "statusline", "output_style"]
SURFACES = {"app", "cli"}
OFFICIAL_SOURCE = re.compile(r"^https://(zcode\.z\.ai/en/docs(/[A-Za-z-]+)?|github\.com/zai-org/ZCode/blob/[0-9a-f]{40}/\S+)$")
SUPPORTED_WORDS = re.compile(r"지원\s*\||\bsupported\b|\bcertified\b", re.IGNORECASE)


def split_row(line):
    """Split a Markdown table row on unescaped pipes."""
    cells = re.split(r"(?<!\\)\|", line.strip())
    return [c.strip().replace("\\|", "|") for c in cells[1:-1]]


def table_rows(text):
    lines = text.splitlines()
    for i, line in enumerate(lines):
        if line.startswith("| host |"):
            header = split_row(line)
            rows = []
            for row in lines[i + 2:]:
                if not row.startswith("|"):
                    break
                rows.append(dict(zip(header, split_row(row))))
            return header, rows
    raise AssertionError("matrix table not found")


class HostMatrixTest(unittest.TestCase):
    def setUp(self):
        self.text = MATRIX.read_text(encoding="utf-8")
        self.header, self.rows = table_rows(self.text)
        self.zcode = [r for r in self.rows if r["host"] == "ZCode"]

    def test_host_matrix_zcode_row_has_nine_items_with_status_and_source(self):
        self.assertEqual(self.header, COLUMNS)
        self.assertTrue(all(len(r) == len(COLUMNS) for r in self.rows), "every row has every column")
        self.assertEqual({r["surface"] for r in self.zcode}, SURFACES)
        for surface in sorted(SURFACES):
            items = [r["item"] for r in self.zcode if r["surface"] == surface]
            with self.subTest(surface=surface):
                self.assertEqual(len(items), len(set(items)), "one row per item")
                for item in REQUIRED_ITEMS:
                    self.assertIn(item, items)
        for r in self.zcode:
            with self.subTest(surface=r["surface"], item=r["item"]):
                self.assertIn(r["capability_status"], STATUSES)
                self.assertRegex(r["source"], OFFICIAL_SOURCE)
                self.assertTrue(r["내용"])
                if r["capability_status"] == "unknown":
                    self.assertIn(r["unknown_class"], UNKNOWN_CLASSES)
                else:
                    self.assertEqual(r["unknown_class"], "")
        self.assertIn("2026-11-08", self.text)
        self.assertIn("3.17.2", self.text)

    def test_host_matrix_no_supported_label_without_certified_status(self):
        # REQ-02D-01 / ZCODE-MATRIX-02: documented or unknown never reads as supported.
        for line in self.text.splitlines():
            if line.startswith("| ZCode"):
                row = dict(zip(COLUMNS, split_row(line)))
                if row["capability_status"] != "certified":
                    with self.subTest(item=row["item"], surface=row["surface"]):
                        self.assertIsNone(SUPPORTED_WORDS.search(line))
        self.assertFalse([r for r in self.zcode if r["capability_status"] in {"probed", "certified"}],
                         "no ZCode run has been recorded, so nothing is probed or certified")

    def test_support_doc_gap_rows_are_classified(self):
        # ZCODE-GAP-01: every gap has a class and a follow-up proposal.
        text = SUPPORT.read_text(encoding="utf-8")
        gaps = [split_row(line) for line in text.splitlines() if re.match(r"^\| G\d+ \|", line)]
        self.assertGreaterEqual(len(gaps), 4)
        for cells in gaps:
            with self.subTest(gap=cells[0]):
                self.assertEqual(len(cells), 5)
                self.assertIn(cells[2], {"구현 필요", "대안 있음", "미검증"})
                self.assertTrue(cells[4])
        self.assertIn("HOLD", text)
        self.assertIn("not_run", text)


if __name__ == "__main__":
    unittest.main()
