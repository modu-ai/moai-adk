import json, pathlib, sys, tempfile, unittest
sys.path.insert(0, str(pathlib.Path(__file__).resolve().parents[1] / "adapters/claude"))
import settings_update as u

class RecoveryTest(unittest.TestCase):
    def test_interruptions_drift_and_symlink(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp)
            (root / ".claude").mkdir(); (root / ".moai-cli").mkdir()
            settings = root / ".claude/settings.json"
            receipt = root / ".moai-cli/claude-settings-owner.json"
            before = b'{"unknown":true}'
            after = b'{"unknown":true,"outputStyle":"MoAI"}'
            saved = {"schema_version":1,"state":"pending","baseline":{},"desired":{"outputStyle":"MoAI"},"before_sha256":u.digest(before),"after_sha256":u.digest(after)}
            receipt.write_bytes(u.encode(saved))
            for content, expected in [(before,"matches_before_snapshot"),(after,"matches_applied_snapshot"),(b'{}',"conflict_preserve_current")]:
                settings.write_bytes(content)
                self.assertEqual(u.recovery_plan(root,json.loads)["state"], expected)
                self.assertEqual(settings.read_bytes(),content)
            settings.unlink()
            self.assertEqual(u.recovery_plan(root,json.loads)["state"],"conflict_preserve_current")
            settings.symlink_to(receipt)
            with self.assertRaises(ValueError): u.recovery_plan(root,json.loads)

    def test_malformed_legacy_and_unowned(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp); (root / ".moai-cli").mkdir()
            receipt = root / ".moai-cli/claude-settings-owner.json"
            self.assertEqual(u.recovery_plan(root,json.loads)["state"],"unowned")
            for bad in [[], {}, {"schema_version":1,"state":"pending","baseline":[]}]:
                receipt.write_text(json.dumps(bad))
                with self.assertRaises(ValueError): u.recovery_plan(root,json.loads)
            receipt.write_text(json.dumps({"schema_version":1,"state":"pending","baseline":{},"desired":{}}))
            self.assertEqual(u.recovery_plan(root,json.loads)["state"],"legacy_pending_unknown")

    def test_changed_snapshot_refuses_write(self):
        with tempfile.TemporaryDirectory() as tmp:
            p=pathlib.Path(tmp)/"settings"
            p.write_bytes(b"later edit")
            with self.assertRaises(ValueError):u.replace_checked(p,b"old",b"new")
            self.assertEqual(p.read_bytes(),b"later edit")

if __name__ == "__main__": unittest.main()
