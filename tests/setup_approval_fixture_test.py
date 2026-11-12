"""SETUP approval-separation fixtures for #2170 (REQ-B3-01/02).

The normative procedure is skills/moai/workflows/setup.md steps 7-8: the
allowed initial local commit and the reviewed Issue publication are TWO
separately approved effects — a dry-run diff is shown first, the commit
approval never implies push or publication, the publication approval publishes
exactly the reviewed draft, and a readback must confirm the published text
before success is reported. Each effect is recorded as
(effect-intend, effect-confirm) in the setup receipt.

This file is the DETECTION/SEMANTICS FIXTURE for those rules: a minimal
simulation of the procedure's decision logic asserts that (AC-1) with no
approval nothing is committed or published, that a commit approval yields a
commit alone, that a publication approval publishes only the reviewed draft,
and (AC-2) that the readback check compares the published text with the
reviewed draft. The procedure itself stays prose executed by the host;
nothing here is product code.
"""

import os
import unittest


class SetupProcedureDocumentedTest(unittest.TestCase):
    """Reads the real product artifact (skills/moai/workflows/setup.md) and
    asserts the approval separation is actually written there — the fixture
    above demonstrates the semantics, this one pins the document."""

    def test_setup_md_carries_the_two_separately_approved_steps(self):
        with open(
            os.path.join(
                os.path.dirname(os.path.abspath(__file__)),
                "..", "skills", "moai", "workflows", "setup.md",
            ),
            "r", encoding="utf-8",
        ) as f:
            doc = f.read()
        # Step 7: the initial local commit — dry-run first, commit only, never
        # a publication permission, no dirty sweep, no amend.
        self.assertIn("Allowed initial local commit", doc)
        self.assertIn("dry-run diff", doc)
        self.assertIn("it never implies push or Issue publication", doc)
        self.assertIn("dirty user changes are not swept in", doc)
        self.assertIn("existing commit is never amended", doc)
        # Step 8: the reviewed Issue publication — a separate approval, exact
        # draft only, readback before success.
        self.assertIn("Reviewed Issue publication", doc)
        self.assertIn("approving the initial commit never approves publication", doc)
        self.assertIn("publish exactly the reviewed draft", doc)
        self.assertIn("read it back", doc)
        self.assertIn("matches the reviewed draft", doc)
        # Both effects are recorded in the setup receipt.
        self.assertIn("setup_initial_commit", doc)
        self.assertIn("setup_issue_publish", doc)
        self.assertIn("effect-intend", doc)
        self.assertIn("effect-confirm", doc)


class SetupProcedureSim:
    """Records the external actions a SETUP run would take, per sync.md's
    effect model. `approvals` is the set the user granted; `reviewed_draft`
    is the exact draft text the user reviewed; `published` simulates what the
    remote would return for the readback."""

    def __init__(self, approvals, reviewed_draft, published=None):
        self.approvals = set(approvals)
        self.reviewed_draft = reviewed_draft
        self.published = published
        self.actions = []
        self.receipt_effects = []

    def _effect(self, marker, action):
        self.receipt_effects.append(("effect-intend", marker))
        self.actions.append(action)
        self.receipt_effects.append(("effect-confirm", marker))

    def run(self):
        if "initial_commit" in self.approvals:
            self._effect("setup_initial_commit", "commit")
        if "issue_publish" in self.approvals:
            self._effect("setup_issue_publish", "issue_publish")
        return self.actions

    def readback_ok(self):
        """AC-2: the published text must equal the reviewed draft — the run
        reports success only after this comparison passes."""
        return self.published == self.reviewed_draft


class SetupApprovalSplitTest(unittest.TestCase):
    def test_ac1_without_approval_nothing_is_committed_or_published(self):
        sim = SetupProcedureSim(approvals=[], reviewed_draft="draft v1",
                                published="draft v1")
        actions = sim.run()
        self.assertEqual(actions, [], "no approval, no external action")
        self.assertEqual(sim.receipt_effects, [])
        self.assertEqual(sim.published, "draft v1")

    def test_reqb301_commit_approval_creates_the_commit_alone(self):
        sim = SetupProcedureSim(approvals=["initial_commit"],
                                reviewed_draft="draft v1", published="draft v1")
        actions = sim.run()
        self.assertEqual(actions, ["commit"])
        self.assertNotIn("push", actions)
        self.assertNotIn("issue_publish", actions)
        self.assertEqual(sim.receipt_effects,
                         [("effect-intend", "setup_initial_commit"),
                          ("effect-confirm", "setup_initial_commit")])

    def test_reqb302_publication_approval_publishes_only_the_reviewed_draft(self):
        sim = SetupProcedureSim(
            approvals=["initial_commit", "issue_publish"],
            reviewed_draft="reviewed draft text",
            published="reviewed draft text",
        )
        actions = sim.run()
        self.assertEqual(actions, ["commit", "issue_publish"])
        self.assertTrue(sim.readback_ok(),
                        "the readback must confirm the published text matches "
                        "the reviewed draft before success is reported")

    def test_ac2_readback_detects_a_deviation_from_the_reviewed_draft(self):
        sim = SetupProcedureSim(approvals=["issue_publish"],
                                reviewed_draft="reviewed draft text",
                                published="DIFFERENT text")
        actions = sim.run()
        self.assertEqual(actions, ["issue_publish"])
        self.assertFalse(sim.readback_ok(),
                         "a published text that deviates from the reviewed "
                         "draft must fail the readback check")
        self.assertNotIn("success", [a for a in actions])

    def test_both_effects_are_recorded_in_the_receipt(self):
        sim = SetupProcedureSim(approvals=["initial_commit", "issue_publish"],
                                reviewed_draft="d", published="d")
        sim.run()
        markers = [m for _, m in sim.receipt_effects]
        self.assertIn("setup_initial_commit", markers)
        self.assertIn("setup_issue_publish", markers)
        self.assertEqual(len(sim.receipt_effects), 4,
                         "each effect is intend + confirm")


if __name__ == "__main__":
    unittest.main()
