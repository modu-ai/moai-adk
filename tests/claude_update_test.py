import json
import pathlib
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "adapters/claude"))
import settings_update

class UpdateTest(unittest.TestCase):
    def install(self, root, style="MoAI-Easy"):
        return subprocess.run([sys.executable, str(ROOT / "adapters/claude/install.py"), str(root), "--style", style], capture_output=True)

    def test_owned_update_drift_and_reexecution(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            self.assertEqual(self.install(root).returncode, 0)
            self.assertEqual(self.install(root, "MoAI").returncode, 0)
            path = root / ".claude/settings.json"
            current = json.loads(path.read_text())
            self.assertEqual(current["outputStyle"], "MoAI")
            current["unknown"] = {"user": 1}
            current["outputStyle"] = "custom"
            path.write_text(json.dumps(current))
            before = path.read_bytes()
            self.assertEqual(self.install(root, "MoAI").returncode, 0)
            self.assertEqual(path.read_bytes(), before)
            self.assertNotEqual(self.install(root, "MoAI-Learn").returncode, 0)
            self.assertEqual(path.read_bytes(), before)

    def test_partial_failure_preserves_later_edit(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            self.assertEqual(self.install(root).returncode, 0)
            path = root / ".claude/settings.json"
            desired = json.loads(path.read_text())
            desired["outputStyle"] = "MoAI"
            plan = settings_update.prepare(root, desired, json.loads)
            def fail():
                raise OSError("fixture interruption")
            with self.assertRaises(OSError):
                settings_update.apply(plan, after_settings=fail)
            self.assertEqual(json.loads(path.read_text())["outputStyle"], "MoAI")
            path.write_text('{"outputStyle":"later user edit"}')
            self.assertNotEqual(self.install(root).returncode, 0)
            self.assertEqual(path.read_text(), '{"outputStyle":"later user edit"}')
            receipt = json.loads((root / ".moai-cli/claude-settings-owner.json").read_text())
            self.assertEqual(receipt["state"], "pending")

    def test_matching_existing_key_is_not_adopted(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            (root / ".claude").mkdir()
            (root / ".claude/settings.json").write_text('{"outputStyle":"MoAI-Easy"}')
            self.assertEqual(self.install(root).returncode, 0)
            self.assertNotEqual(self.install(root, "MoAI").returncode, 0)

if __name__ == "__main__":
    unittest.main()
