"""docs/release-rehearsal-2026-11-12.md (Issue #2512): every checklist step has a row."""
import re
import unittest
from pathlib import Path

DOC = Path(__file__).resolve().parent.parent / "docs" / "release-rehearsal-2026-11-12.md"

# Every step of the #2492 checklist (b448d3b2fe), in order, plus the two Issue-scope rows.
EXPECTED = [
    "a1", "a2", "a3", "b1", "b2", "b3", "c1", "c2", "c3", "c4", "c5", "c6",
    "d1", "d2", "d3", "d4", "e1x", "e1", "e2", "e3", "f1", "g1", "g2", "g3",
    "h1", "h2", "h3", "h4", "h5", "i1", "i2", "i3", "i4",
    "j1", "j2", "j3", "j4", "j5", "j6", "k1", "p1", "p2", "p3", "p4", "p5", "p6",
    "x1", "x2",
]


def rows(text):
    section = text.split("## 단계별 영수증", 1)[1].split("\n## ", 1)[0]
    out = []
    for line in section.splitlines():
        if not line.startswith("| ") or line.startswith("| id ") or line.startswith("| ---"):
            continue
        cells = [c.strip() for c in re.split(r"(?<!\\)\|", line.strip())[1:-1]]
        out.append(cells)
    return out


class ReleaseRehearsalTest(unittest.TestCase):
    def setUp(self):
        self.text = DOC.read_text(encoding="utf-8")
        self.rows = rows(self.text)

    def test_every_checklist_step_has_exactly_one_row_in_order(self):
        self.assertEqual([r[0] for r in self.rows], EXPECTED)

    def test_each_row_has_a_status_and_the_evidence_it_needs(self):
        for r in self.rows:
            self.assertEqual(len(r), 7, r)
            step, _, status, cmd, exit_code, observed, _ = r
            self.assertIn(status, ("passed", "checked", "not_run"), step)
            if status == "passed":
                self.assertEqual(exit_code, "0", f"{step}: a passed row records exit 0")
                self.assertNotIn(cmd, ("", "—"), f"{step}: a passed row names its command")
            self.assertNotIn(observed, ("", "—"), f"{step}: observation or not_run reason")

    def test_the_tree_and_the_checklist_are_pinned_by_full_sha(self):
        self.assertRegex(self.text, r"tip `[0-9a-f]{40}`")
        self.assertIn("b448d3b2fe1755765c9bcffa43ed3f0d53d8e0c3", self.text)

    def test_no_publish_effect_is_recorded_as_done(self):
        by_id = {r[0]: r for r in self.rows}
        # j1 tag ruleset, j2 attestation workflow, j6 operator check: audit note on PR #2574.
        for step in ("a3", "d4", "e1", "e2", "g1", "g3", "i3", "j1", "j2", "j6", "k1",
                     "p1", "p2", "p3", "p4", "p5", "p6"):
            self.assertEqual(by_id[step][2], "not_run", step)

    def test_no_personal_absolute_path(self):
        self.assertNotRegex(self.text, r"/Users/|/home/[a-z]")


if __name__ == "__main__":
    unittest.main()
