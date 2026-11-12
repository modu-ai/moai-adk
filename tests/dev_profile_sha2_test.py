"""Pin the dev-profile sha2 optimization (Issue #2507).

Every update/recovery operation hashes the running moai-cli binary to derive the
install id. With sha2 unoptimized, a debug launch costs about 2.2 s of CPU and
recovery_cli took 420 s instead of 17 s on a loaded machine. No test asserts wall
time (#2385), so this file pins the setting itself: removing or weakening it fails
here instead of silently slowing every debug test run again.
"""

from pathlib import Path
import tomllib
import unittest

ROOT = Path(__file__).resolve().parents[1]


class DevProfileSha2Test(unittest.TestCase):
    def setUp(self):
        self.manifest = tomllib.loads((ROOT / "Cargo.toml").read_text(encoding="utf-8"))

    def test_sha2_is_optimized_in_dev_builds(self):
        override = self.manifest.get("profile", {}).get("dev", {}).get("package", {}).get("sha2")
        self.assertIsNotNone(override, "[profile.dev.package.sha2] is missing (Issue #2507)")
        self.assertEqual(override.get("opt-level"), 3)

    def test_override_names_a_real_dependency(self):
        # A renamed or dropped sha2 dependency would leave the override silently unused.
        self.assertIn("sha2", self.manifest.get("dependencies", {}))


if __name__ == "__main__":
    unittest.main()
