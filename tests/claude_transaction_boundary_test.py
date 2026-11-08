import json
import pathlib
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / "adapters/claude"))
import settings_update as u

class BoundaryTest(unittest.TestCase):
    def test_later_edit_or_deletion_keeps_pending(self):
        for delete in (False, True):
            with self.subTest(delete=delete), tempfile.TemporaryDirectory() as tmp:
                root = pathlib.Path(tmp)
                plan = u.prepare(root, {"outputStyle": "MoAI"}, json.loads)
                settings = root / ".claude/settings.json"
                def change():
                    if delete:
                        settings.unlink()
                    else:
                        settings.write_bytes(b'{"user":"later"}')
                with self.assertRaises(ValueError):
                    u.apply(plan, after_settings=change)
                receipt = root / ".moai-cli/claude-settings-owner.json"
                self.assertEqual(json.loads(receipt.read_bytes())["state"], "pending")
                self.assertEqual(settings.exists(), not delete)
                if not delete:
                    self.assertEqual(settings.read_bytes(), b'{"user":"later"}')
                with self.assertRaises(ValueError):
                    u.prepare(root, {"outputStyle": "MoAI"}, json.loads)

    def test_cli_diagnostics_are_read_only_and_reject_invalid_json(self):
        with tempfile.TemporaryDirectory(prefix="moai recovery ") as tmp:
            root = pathlib.Path(tmp)
            plan = u.prepare(root, {"outputStyle": "MoAI"}, json.loads)
            def interrupt():
                raise OSError("interrupted")
            with self.assertRaises(OSError):
                u.apply(plan, after_settings=interrupt)
            def snapshot():
                return {str(p): p.read_bytes() for p in root.rglob("*") if p.is_file()}
            before = snapshot()
            command = [sys.executable, str(ROOT / "adapters/claude/install.py"), str(root), "--recovery-plan"]
            result = subprocess.run(command, capture_output=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(json.loads(result.stdout)["state"], "matches_applied_snapshot")
            self.assertEqual(before, snapshot())
            receipt = root / ".moai-cli/claude-settings-owner.json"
            receipt.write_text('{"schema_version":1,"state":"committed","baseline":{},"extra":NaN}')
            before = snapshot()
            self.assertNotEqual(subprocess.run(command, capture_output=True).returncode, 0)
            self.assertEqual(before, snapshot())

if __name__ == "__main__":
    unittest.main()
