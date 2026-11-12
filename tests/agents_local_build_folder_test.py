#!/usr/bin/env python3
"""Issue #2380 WT-2: the repository's AGENTS.local.md keeps one build folder per
worktree. The check set must not set CARGO_TARGET_DIR, and the build-folder rule
must say so, so that a later edit cannot quietly bring back a shared or
per-session target folder (a merge with #2563 nearly did)."""

import pathlib
import re
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
LOCAL = ROOT / "AGENTS.local.md"
CONTRACT = ROOT / "docs" / "lane-loop-contract.md"


def check_set_block(text: str) -> str:
    """Return the first ```sh block after the 'Check set' bullet."""
    start = text.index("Check set, run from the repository root")
    match = re.search(r"```sh\n(.*?)```", text[start:], re.S)
    if match is None:
        raise AssertionError("check set block not found")
    return match.group(1)


class BuildFolderRuleTest(unittest.TestCase):
    def setUp(self):
        self.text = LOCAL.read_text(encoding="utf-8")

    def test_check_set_does_not_set_cargo_target_dir(self):
        block = check_set_block(self.text)
        self.assertIn("cargo test --offline --locked", block)
        self.assertNotIn("CARGO_TARGET_DIR", block)

    def test_build_folder_rule_names_the_worktree_target(self):
        rules = [ln for ln in self.text.splitlines() if ln.startswith("- Build folders:")]
        self.assertEqual(len(rules), 1, rules)
        self.assertIn("`<worktree>/target/` (do not set `CARGO_TARGET_DIR`)", rules[0])
        self.assertIn("so that rule is retired", rules[0])

    def test_no_rule_assigns_a_per_session_target_dir(self):
        self.assertIsNone(re.search(r"CARGO_TARGET_DIR=/tmp/", self.text))

    def test_lane_loop_contract_row_matches(self):
        contract = CONTRACT.read_text(encoding="utf-8")
        self.assertIn("빌드 폴더는 worktree 안 `target/` 하나(`CARGO_TARGET_DIR` 지정 안 함, #2380 WT-2)", contract)


if __name__ == "__main__":
    unittest.main()
