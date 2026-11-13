"""SYNC drift detection fixtures for the four kinds added by #2168 (REQ-B1-01).

The normative procedure is skills/moai/workflows/sync.md step 3: drift beyond
the original three kinds (omission, amendment, document mismatch) is classified
as `(kind, ground, proposed_action)`, reported separately, and the approved
SPEC is never auto-weakened. This file is the DETECTION FIXTURE for the four
new kinds — each test builds a minimal state, applies the documented rule, and
asserts the classification plus that the SPEC bytes are untouched. The
procedure itself stays prose executed by the host; nothing here is product
code.

Fixture states mirror the design's seven-way classification wording (original text:
docs/design-review/appendix/state-vocabulary-crosswalk.md):
- spec_oracle_conflict       the SPEC wording and a test oracle contradict each other
- broken_reference           a deletion or rename severed a reference
- done_but_stale_evidence    the Issue is done but its verification evidence is missing or stale
- explicit_link_unreviewed   an explicit link exists but its meaning was never reviewed
"""

import os
import re
import tempfile
import unittest

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SYNC = os.path.join(ROOT, "skills", "moai", "workflows", "sync.md")
CROSSWALK_REL = "docs/design-review/appendix/state-vocabulary-crosswalk.md"
ORIGINAL_THREE = ("omission", "amendment", "document mismatch")


def sync_step3():
    """The text of sync.md procedure step 3, up to step 4."""
    with open(SYNC, encoding="utf-8") as f:
        text = f.read()
    match = re.search(r"^3\. \*\*Classify differences\*\*.*?(?=^4\. )", text, re.S | re.M)
    if match is None:
        raise AssertionError("sync.md has no procedure step 3 'Classify differences'")
    return match.group(0)


def documented_kinds(step3):
    """{kind: (ground, proposed_action)} from the step-3 bullets
    `- `kind` — ... Ground: ... Proposed action: ...`."""
    kinds = {}
    for m in re.finditer(r"^\s+- `([a-z_]+)` — (.*)$", step3, re.M):
        body = m.group(2)
        ground = re.search(r"Ground: (.*?)\. Proposed action:", body)
        action = re.search(r"Proposed action: (.*?)\.?$", body)
        kinds[m.group(1)] = (
            ground.group(1) if ground else None,
            action.group(1) if action else None,
        )
    return kinds


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


class SyncStep3DocumentTest(unittest.TestCase):
    """The fixture above is checked against the normative text, not against itself:
    reverting sync.md step 3 to the three-kind wording breaks these tests."""

    def setUp(self):
        self.step3 = sync_step3()
        self.kinds = documented_kinds(self.step3)

    def test_step3_names_the_original_three_kinds(self):
        for name in ORIGINAL_THREE:
            self.assertIn(f"*{name}*", self.step3)

    def test_step3_documents_exactly_the_four_new_kinds_with_ground_and_action(self):
        self.assertEqual(
            set(self.kinds),
            {"spec_oracle_conflict", "broken_reference", "done_but_stale_evidence",
             "explicit_link_unreviewed"},
        )
        for kind, (ground, action) in self.kinds.items():
            with self.subTest(kind=kind):
                self.assertTrue(ground, "each kind states its Ground")
                self.assertTrue(action, "each kind states its Proposed action")
        self.assertIn("(kind, ground, proposed_action)", self.step3)
        self.assertIn("never resolved by weakening the approved SPEC", self.step3)

    def test_step3_names_the_seven_way_basis_and_the_crosswalk(self):
        self.assertIn("seven", self.step3)
        self.assertIn(CROSSWALK_REL, self.step3)
        self.assertNotIn("until it exists", self.step3)
        with open(os.path.join(ROOT, CROSSWALK_REL), encoding="utf-8") as f:
            crosswalk = f.read()
        for kind in self.kinds:
            with self.subTest(kind=kind):
                self.assertIn(f"`{kind}`", crosswalk)

    def test_fixture_classification_matches_the_documented_kinds(self):
        tmp = tempfile.mkdtemp(prefix="sync-drift-doc-")
        try:
            states = [
                {"spec_line": "exit 2", "oracle_line": "exit 1"},
                {"references": [{"target": os.path.join(tmp, "gone.md")}]},
                {"claimed_done": True, "evidence_sha": "a" * 40, "current_sha": "b" * 40},
                {"links": [{"relation": "same-as", "to": "X", "reconciled": False}]},
            ]
            seen = set()
            for state in states:
                kind, _ground, action = classify(state)
                with self.subTest(kind=kind):
                    self.assertIn(kind, self.kinds)
                    documented = self.kinds[kind][1].lower()
                    self.assertIn(action.lower(), documented)
                seen.add(kind)
            self.assertEqual(seen, set(self.kinds))
        finally:
            __import__("shutil").rmtree(tmp, ignore_errors=True)


if __name__ == "__main__":
    unittest.main()
