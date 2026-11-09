import copy
import importlib.util
import pathlib
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("settings_plan", ROOT / "experiments/claude_settings_plan.py")
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
plan = module.plan_settings


class SettingsPlanTest(unittest.TestCase):
    def test_adds_only_selected_keys_and_preserves_unknown_values(self):
        current = {"model": "sonnet", "hooks": {"Start": [{"user": True}]}}
        original = copy.deepcopy(current)
        result, conflicts = plan({}, current, {"outputStyle": "MoAI-Easy"})
        self.assertEqual(conflicts, [])
        self.assertEqual(result, {**original, "outputStyle": "MoAI-Easy"})
        self.assertEqual(current, original)

    def test_user_owned_or_modified_setting_causes_atomic_conflict(self):
        current = {"model": "sonnet", "statusLine": {"type": "command", "command": "orca"}}
        desired = {"outputStyle": "MoAI-Easy", "statusLine": {"type": "command", "command": "moai"}}
        result, conflicts = plan({}, current, desired)
        self.assertIsNone(result)
        self.assertEqual(conflicts, ["statusLine"])
        self.assertEqual(current["statusLine"]["command"], "orca")

    def test_three_way_update_idempotence_and_user_drift(self):
        baseline = {"outputStyle": "old"}
        result, conflicts = plan(baseline, {"outputStyle": "old"}, {"outputStyle": "new"})
        self.assertEqual((result, conflicts), ({"outputStyle": "new"}, []))
        self.assertEqual(plan(baseline, result, {"outputStyle": "new"}), (result, []))
        self.assertEqual(plan(baseline, {"outputStyle": "user"}, {"outputStyle": "old"}), ({"outputStyle": "user"}, []))
        self.assertEqual(plan(baseline, {"outputStyle": "user"}, {"outputStyle": "new"}), (None, ["outputStyle"]))

    def test_missing_null_and_boolean_are_distinct(self):
        self.assertEqual(plan({}, {"outputStyle": None}, {"outputStyle": "new"}), (None, ["outputStyle"]))
        self.assertEqual(plan({"outputStyle": None}, {"outputStyle": None}, {}), ({}, []))
        self.assertEqual(plan({"outputStyle": True}, {"outputStyle": 1}, {"outputStyle": False}), (None, ["outputStyle"]))

    def test_rejects_unmanaged_requests_and_invalid_json(self):
        # permissions and hooks are Rust-only keys (module.RUST_ONLY, #2294): the
        # mirror refuses them on purpose, although Rust MANAGED_SETTINGS lists them.
        with self.assertRaises(ValueError):
            plan({}, {}, {"permissions": {}})
        with self.assertRaises(ValueError):
            plan({}, {}, {"hooks": {}})
        with self.assertRaises(ValueError):
            plan({}, {"model": float("nan")}, {})


if __name__ == "__main__":
    unittest.main()
