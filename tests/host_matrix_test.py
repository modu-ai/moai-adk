"""Offline checks for docs/host-support-matrix.md (ZCode rows: 02-D #1994; Codex rows: 02-C #1993).

The matrix records what official sources say; these tests keep its shape honest. They do
not re-fetch the sources (an independent audit does) and do not prove any host behaviour.
"""
import pathlib
import re
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
MATRIX = ROOT / "docs/host-support-matrix.md"
ZCODE_SUPPORT = ROOT / "docs/zcode-support.md"
CODEX_CONTRACT = ROOT / "docs/codex-adapter-contract.md"
COLUMNS = ["host", "surface", "item", "capability_status", "unknown_class", "내용", "source"]
STATUSES = {"unknown", "documented", "probed", "certified", "stale"}
UNKNOWN_CLASSES = {"표시 불가", "대안", "시험 대기"}
# REQ-02D-02: the nine items every ZCode surface must carry.
REQUIRED_ITEMS = ["agents_loading", "skill_path", "skill_limits", "explicit_invocation", "import_dialog",
                  "hooks", "plugin_manifest", "statusline", "output_style"]
SURFACES = {"app", "cli"}
# 02-C: items stated for every Codex surface, and items the docs state per surface.
CODEX_SHARED = ["skill_path", "skill_required_fields", "openai_yaml", "agents_loading", "agents_budget",
                "trust_layer", "hooks"]
CODEX_PER_SURFACE = ["explicit_invocation", "status_line"]
CODEX_SURFACES = {"all", "desktop", "cli", "ide"}
SOURCES = {
    "ZCode": re.compile(r"^https://(zcode\.z\.ai/en/docs(/[A-Za-z-]+)?|github\.com/zai-org/ZCode/blob/[0-9a-f]{40}/\S+)$"),
    "Codex": re.compile(r"^https://learn\.chatgpt\.com/docs/[a-z0-9/-]+$"),
}
SUPPORTED_WORDS = re.compile(r"지원\s*\||\bsupported\b|\bcertified\b", re.IGNORECASE)


def split_row(line):
    """Split a Markdown table row on unescaped pipes."""
    cells = re.split(r"(?<!\\)\|", line.strip())
    return [c.strip().replace("\\|", "|") for c in cells[1:-1]]


def table_rows(text):
    """Every matrix table (a header starting `| host |`), concatenated."""
    lines = text.splitlines()
    header, rows = None, []
    for i, line in enumerate(lines):
        if line.startswith("| host |"):
            this = split_row(line)
            if header is not None and this != header:
                raise AssertionError("matrix tables disagree on columns")
            header = this
            for row in lines[i + 2:]:
                if not row.startswith("|"):
                    break
                rows.append(dict(zip(header, split_row(row))))
    if header is None:
        raise AssertionError("matrix table not found")
    return header, rows


class HostMatrixTest(unittest.TestCase):
    def setUp(self):
        self.text = MATRIX.read_text(encoding="utf-8")
        self.header, self.rows = table_rows(self.text)
        self.zcode = [r for r in self.rows if r["host"] == "ZCode"]
        self.codex = [r for r in self.rows if r["host"] == "Codex"]

    def check_row(self, r):
        with self.subTest(host=r["host"], surface=r["surface"], item=r["item"]):
            self.assertIn(r["capability_status"], STATUSES)
            for source in r["source"].split(" ; "):
                self.assertRegex(source, SOURCES[r["host"]])
            self.assertTrue(r["내용"])
            if r["capability_status"] == "unknown":
                self.assertIn(r["unknown_class"], UNKNOWN_CLASSES)
            else:
                self.assertEqual(r["unknown_class"], "")

    def test_host_matrix_zcode_row_has_nine_items_with_status_and_source(self):
        self.assertEqual(self.header, COLUMNS)
        self.assertTrue(all(len(r) == len(COLUMNS) for r in self.rows), "every row has every column")
        self.assertEqual({r["host"] for r in self.rows}, set(SOURCES))
        self.assertEqual({r["surface"] for r in self.zcode}, SURFACES)
        for surface in sorted(SURFACES):
            items = [r["item"] for r in self.zcode if r["surface"] == surface]
            with self.subTest(surface=surface):
                self.assertEqual(len(items), len(set(items)), "one row per item")
                for item in REQUIRED_ITEMS:
                    self.assertIn(item, items)
        for r in self.zcode:
            self.check_row(r)
        self.assertIn("2026-11-08", self.text)
        self.assertIn("3.17.2", self.text)

    def test_host_matrix_codex_rows_cover_shared_and_per_surface_items(self):
        self.assertEqual({r["surface"] for r in self.codex}, CODEX_SURFACES)
        pairs = [(r["surface"], r["item"]) for r in self.codex]
        self.assertEqual(len(pairs), len(set(pairs)), "one row per surface and item")
        for item in CODEX_SHARED:
            self.assertIn(("all", item), pairs)
        for surface in sorted(CODEX_SURFACES - {"all"}):
            for item in CODEX_PER_SURFACE:
                self.assertIn((surface, item), pairs)
        for r in self.codex:
            self.check_row(r)
        # Trust is never claimed: the doctor does not read config.toml.
        trust = [r for r in self.codex if r["item"] == "trust_layer"][0]
        self.assertIn("not_evaluated", trust["내용"])
        self.assertIn("codex-cli 0.163.0", self.text)

    def test_host_matrix_no_supported_label_without_certified_status(self):
        # REQ-02D-01 / ZCODE-MATRIX-02, for every host: documented or unknown never reads as supported.
        for line in self.text.splitlines():
            if line.startswith("| ZCode") or line.startswith("| Codex"):
                row = dict(zip(COLUMNS, split_row(line)))
                if row["capability_status"] != "certified":
                    with self.subTest(host=row["host"], item=row["item"], surface=row["surface"]):
                        self.assertIsNone(SUPPORTED_WORDS.search(line))
        self.assertFalse([r for r in self.rows if r["capability_status"] in {"probed", "certified"}],
                         "no host run has been recorded, so nothing is probed or certified")

    def test_support_doc_gap_rows_are_classified(self):
        # ZCODE-GAP-01, and the same shape for the Codex gaps: a class and a follow-up proposal.
        for doc in (ZCODE_SUPPORT, CODEX_CONTRACT):
            text = doc.read_text(encoding="utf-8")
            gaps = [split_row(line) for line in text.splitlines() if re.match(r"^\| [GC]\d+ \|", line)]
            with self.subTest(doc=doc.name):
                self.assertGreaterEqual(len(gaps), 4)
                for cells in gaps:
                    self.assertEqual(len(cells), 5, cells[0])
                    self.assertIn(cells[2], {"구현 필요", "대안 있음", "미검증"}, cells[0])
                    self.assertTrue(cells[4], cells[0])
                self.assertIn("HOLD", text)
                self.assertIn("not_run", text)


if __name__ == "__main__":
    unittest.main()
