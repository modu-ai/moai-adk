"""#2294: the Python settings mirrors stay in step with the Rust canonical lists.

Rust (`src/claude.rs` `MANAGED_SETTINGS` / `USER_SETTINGS`) is the canonical settings
path. The two Python mirrors deliberately manage fewer keys; each declares the keys it
leaves to Rust in `RUST_ONLY`. The parity rule is exact: mirror keys plus `RUST_ONLY`
must equal Rust `MANAGED_SETTINGS`. Removing a key from the Rust list, or adding one
without deciding the mirror's side, fails this test.
"""
import importlib.util
import pathlib
import re
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[1]


def load(name, rel):
    spec = importlib.util.spec_from_file_location(name, ROOT / rel)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def rust_list(name):
    """Reads one `pub const NAME: &[&str] = &[...];` from src/claude.rs; fails if absent."""
    source = (ROOT / "src/claude.rs").read_text(encoding="utf-8")
    found = re.findall(r"pub const " + name + r": &\[&str\] = &\[(.*?)\];", source, re.S)
    if len(found) != 1:
        raise AssertionError(f"{name} not found exactly once in src/claude.rs ({len(found)})")
    keys = re.findall(r'"([^"]*)"', found[0])
    # Second, independent measure: the comma-separated item count must agree.
    items = [part for part in found[0].split(",") if part.strip()]
    if len(items) != len(keys):
        raise AssertionError(f"{name} parse disagreement: {items} vs {keys}")
    return keys


MIRRORS = {
    "settings_update": "adapters/claude/settings_update.py",
    "settings_plan": "experiments/claude_settings_plan.py",
}


class MirrorParityTest(unittest.TestCase):
    def test_rust_lists_are_readable_and_nested(self):
        managed = rust_list("MANAGED_SETTINGS")
        user = rust_list("USER_SETTINGS")
        self.assertEqual(len(managed), len(set(managed)), managed)
        self.assertTrue(set(user) <= set(managed), (user, managed))

    def test_each_mirror_plus_rust_only_equals_rust_managed(self):
        managed = set(rust_list("MANAGED_SETTINGS"))
        for name, rel in MIRRORS.items():
            with self.subTest(mirror=rel):
                module = load(name, rel)
                allowed = set(module._ALLOWED)
                rust_only = set(module.RUST_ONLY)
                self.assertEqual(allowed & rust_only, set(), rel)
                self.assertEqual(allowed | rust_only, managed, rel)
                # What the mirror does manage, a caller may also set through Rust.
                self.assertTrue(allowed <= set(rust_list("USER_SETTINGS")), rel)

    def test_mirrors_refuse_every_rust_only_key(self):
        for name, rel in MIRRORS.items():
            module = load(name, rel)
            plan = module.plan_settings
            for key in sorted(module.RUST_ONLY):
                with self.subTest(mirror=rel, key=key):
                    with self.assertRaises(ValueError):
                        plan({}, {}, {key: {}})
                    with self.assertRaises(ValueError):
                        plan({key: {}}, {}, {})


if __name__ == "__main__":
    unittest.main()
