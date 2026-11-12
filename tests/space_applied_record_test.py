"""Space application record for EP-G (#2207 G2): docs/design-review/applied.md.

The record is the AC-1 checklist and AC-2 sha log of applying the design fixes to the Space.
These tests pin that its G2 table covers exactly the 13 fix files, that every recorded fix
sha256 and item count matches the file (so a later fix edit cannot silently diverge from what
was applied), that each Space page id is a known page, and that a row is `applied` only with a
user confirmation id and the re-collected copy's sha256.

They also pin the 2026-11-12 tip re-check of the replacement text: no fix still says #2016 is
open, no replacement cell carries a PROFILE_KEYS or route count, and every fix carries the
re-check section.
"""
import hashlib
import re
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
RECORD = ROOT / "docs/design-review/applied.md"
FIXES = ROOT / "docs/design-review/fixes"
PAGES = ROOT / "docs/design-review/pages.tsv"
G2_PREFIXES = ("group2-", "group3-", "plan-")
SHA = re.compile(r"^[0-9a-f]{64}$")
ROW = re.compile(r"^\| `fixes/(?P<fix>[^`]+)` \| `(?P<page>page_[0-9a-f]+)` [^|]+ \| (?P<items>\d+) \| "
                 r"`(?P<sha>[^`]+)` \| (?P<confirm>[^|]+) \| (?P<copy>[^|]+) \| (?P<state>[a-z_]+) \|$")
RECHECK = "## 현재 tip 재확인 (#2207, 2026-11-12"


def g2_rows(text):
    section = text.split("## G2 #2207")[1].split("\n## ")[0]
    return [ROW.match(line) for line in section.splitlines() if line.startswith("| `fixes/")]


def g2_fixes():
    return sorted(p for p in FIXES.glob("*.md") if p.name.startswith(G2_PREFIXES))


def item_rows(text):
    return len(re.findall(r"^\| \d+( |\()", text, flags=re.M))


def replacement_cells(text):
    body = text.split("## 갱신 기록")[0]
    for line in body.splitlines():
        cells = line.split(" | ")
        if re.match(r"^\| \d", line) and len(cells) >= 5:
            yield cells[3]


class SpaceAppliedRecordTest(unittest.TestCase):
    def setUp(self):
        self.text = RECORD.read_text(encoding="utf-8")
        self.rows = g2_rows(self.text)

    def test_every_row_parses(self):
        self.assertTrue(self.rows)
        self.assertTrue(all(self.rows), "a G2 row does not match the table shape")

    def test_the_g2_table_covers_exactly_the_thirteen_fix_files(self):
        self.assertEqual(sorted(r["fix"] for r in self.rows), [p.name for p in g2_fixes()])
        self.assertEqual(len(self.rows), 13)

    def test_recorded_sha256_and_item_count_match_each_fix_file(self):
        for r in self.rows:
            data = (FIXES / r["fix"]).read_bytes()
            self.assertEqual(r["sha"], hashlib.sha256(data).hexdigest(), r["fix"])
            self.assertEqual(int(r["items"]), item_rows(data.decode("utf-8")), r["fix"])

    def test_each_page_is_a_known_space_page(self):
        known = {line.split("\t")[0] for line in PAGES.read_text(encoding="utf-8").splitlines() if line}
        for r in self.rows:
            self.assertIn(r["page"], known, r["fix"])
            self.assertTrue(r["page"].startswith("page_" + r["fix"].split("-")[1].split(".")[0][:8]), r["fix"])

    def test_a_row_is_applied_only_with_a_confirmation_and_a_copy_sha(self):
        for r in self.rows:
            confirm, copy = r["confirm"].strip(), r["copy"].strip().strip("`")
            self.assertIn(r["state"], ("not_applied", "applied"), r["fix"])
            if r["state"] == "applied":
                self.assertRegex(confirm, r"^\d{8,}$", r["fix"])
                self.assertRegex(copy, SHA, r["fix"])
            else:
                self.assertEqual((confirm, copy), ("—", "—"), r["fix"])

    def test_every_fix_carries_the_tip_recheck_section(self):
        for p in g2_fixes():
            self.assertIn(RECHECK, p.read_text(encoding="utf-8"), p.name)

    def test_no_replacement_text_states_a_stale_fact(self):
        stale = [
            re.compile(r"#2016은 OPEN|#2016은 OPEN으로"),
            re.compile(r"PROFILE_KEYS`? ?\d+개"),
            re.compile(r"\d+ route\("),
            # PR #2534 (#2237) connected codex-config-plan/-apply; audit note on PR #2576.
            re.compile(r"CLI 명령(은)? 아직 없"),
        ]
        for p in g2_fixes():
            for cell in replacement_cells(p.read_text(encoding="utf-8")):
                for pat in stale:
                    self.assertIsNone(pat.search(cell), f"{p.name}: {pat.pattern}")


if __name__ == "__main__":
    unittest.main()
