"""SYNC drift detection fixtures for the four kinds added by #2168 (REQ-B1-01).

The normative procedure is skills/moai/workflows/sync.md step 3: drift beyond
the original three kinds (omission, amendment, document mismatch) is classified
as `(kind, ground, proposed_action)`, reported separately, and the approved
SPEC is never auto-weakened. This file is the DETECTION FIXTURE for the four
new kinds — each test builds a minimal state, applies the documented rule, and
asserts the classification plus that the SPEC bytes are untouched. The
procedure itself stays prose executed by the host; nothing here is product
code.

Fixture states mirror the design's seven-way classification wording:
- spec_oracle_conflict       "SPEC 문구와 테스트 oracle이 충돌함"
- broken_reference           "삭제·이름 변경으로 참조가 끊김"
- done_but_stale_evidence    "Issue는 완료인데 검증 증거가 없거나 stale임"
- explicit_link_unreviewed   "명시 link는 있으나 의미가 맞는지 검토되지 않음"
"""

import os
import tempfile
import unittest


def classify(state):
    """The four new drift rules from sync.md step 3, in the documented order.

    `state` keys: spec_line/oracle_line (quoted lines), references (list of
    {"target": path}), claimed_done + evidence_sha + current_sha (evidence
    freshness), links (list of {"relation", "reconciled"}). Returns
    (kind, ground, proposed_action) or None when none of the four applies.
    """
    spec_line = state.get("spec_line")
    oracle_line = state.get("oracle_line")
    if spec_line is not None and oracle_line is not None and spec_line != oracle_line:
        return (
            "spec_oracle_conflict",
            f'SPEC: "{spec_line}" vs oracle: "{oracle_line}"',
            "SPEC amendment decided by the host, never auto-edited",
        )
    for ref in state.get("references", []):
        if not os.path.exists(ref["target"]):
            return (
                "broken_reference",
                f'dangling target: {ref["target"]}',
                "re-point the reference or restore the target",
            )
    if state.get("claimed_done") and state.get("evidence_sha") != state.get("current_sha"):
        return (
            "done_but_stale_evidence",
            f'claimed done with evidence sha {state.get("evidence_sha")} '
            f'but current sha is {state.get("current_sha")}',
            "re-run the check and re-attach the evidence",
        )
    for link in state.get("links", []):
        if not link.get("reconciled"):
            return (
                "explicit_link_unreviewed",
                f'"{link.get("relation")}" link to {link.get("to")} was never '
                "reconciled",
                "review and record the link's semantics before relying on it",
            )
    return None


class SyncDriftFixtureTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.mkdtemp(prefix="sync-drift-")
        self.spec_path = os.path.join(self.tmp, "SPEC.md")
        with open(self.spec_path, "w", encoding="utf-8") as f:
            f.write("---\nname: fixture\n---\napproved body\n")

    def tearDown(self):
        shutil = __import__("shutil")
        shutil.rmtree(self.tmp, ignore_errors=True)

    def assert_spec_untouched(self):
        with open(self.spec_path, "r", encoding="utf-8") as f:
            self.assertEqual(f.read(), "---\nname: fixture\n---\napproved body\n",
                             "REQ-B1-01: the approved SPEC must never be auto-weakened")

    def test_spec_oracle_conflict_is_detected_with_both_lines(self):
        kind, ground, action = classify({
            "spec_line": "missing file exits with code 2",
            "oracle_line": "missing file exits with code 1",
        })
        self.assertEqual(kind, "spec_oracle_conflict")
        self.assertIn("code 2", ground)
        self.assertIn("code 1", ground)
        self.assertIn("never auto-edited", action)
        self.assert_spec_untouched()

    def test_broken_reference_is_detected_by_dangling_target(self):
        dangling = os.path.join(self.tmp, "docs", "guide.md")
        alive = os.path.join(self.tmp, "docs", "alive.md")
        os.makedirs(os.path.dirname(alive), exist_ok=True)
        with open(alive, "w", encoding="utf-8") as f:
            f.write("alive\n")
        kind, ground, action = classify({
            "references": [{"target": alive}, {"target": dangling}],
        })
        self.assertEqual(kind, "broken_reference")
        self.assertIn("guide.md", ground)
        self.assertIn("re-point", action)
        self.assert_spec_untouched()

    def test_done_but_stale_evidence_is_detected_by_sha_mismatch(self):
        kind, ground, action = classify({
            "claimed_done": True,
            "evidence_sha": "a" * 40,
            "current_sha": "b" * 40,
        })
        self.assertEqual(kind, "done_but_stale_evidence")
        self.assertIn("a" * 40, ground)
        self.assertIn("re-run the check", action)
        self.assert_spec_untouched()

    def test_explicit_link_without_reconciliation_is_detected(self):
        kind, ground, action = classify({
            "links": [{"relation": "same-as", "to": "SPEC-PORTABLE-001",
                       "reconciled": False}],
        })
        self.assertEqual(kind, "explicit_link_unreviewed")
        self.assertIn("SPEC-PORTABLE-001", ground)
        self.assertIn("review and record", action)
        self.assert_spec_untouched()

    def test_clean_state_classifies_to_none(self):
        alive = os.path.join(self.tmp, "docs", "alive.md")
        os.makedirs(os.path.dirname(alive), exist_ok=True)
        with open(alive, "w", encoding="utf-8") as f:
            f.write("alive\n")
        self.assertIsNone(classify({
            "spec_line": "identical",
            "oracle_line": "identical",
            "references": [{"target": alive}],
            "claimed_done": False,
            "links": [{"relation": "same-as", "to": "SPEC-001",
                       "reconciled": True}],
        }))
        self.assert_spec_untouched()


if __name__ == "__main__":
    unittest.main()
