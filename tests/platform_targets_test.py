"""SPEC-20261118-platform-targets: the release (dist) targets are the three supported platforms.

The user's decision of 2026-11-17 names macOS Apple silicon, Linux x86_64 and Windows x86_64 (the
environments Claude Code and Codex support). `dist-workspace.toml` lists the targets cargo-dist
builds an archive for; the installers (`install/install.sh`, `install/install.ps1`) choose one of
those archives by target triple. This test reads the config and the two installers; it does not
run cargo-dist.
"""

import os
import re
import tomllib
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CONFIG = os.path.join(REPO_ROOT, "dist-workspace.toml")
INSTALLERS = [
    os.path.join(REPO_ROOT, "install", "install.sh"),
    os.path.join(REPO_ROOT, "install", "install.ps1"),
]

SUPPORTED = [
    "aarch64-apple-darwin",
    "x86_64-pc-windows-msvc",
    "x86_64-unknown-linux-gnu",
]
# A full target triple of the shape the installers and cargo-dist use.
TRIPLE = re.compile(r"\b(?:aarch64|x86_64|i686|armv7)-[a-z0-9]+-[a-z]+(?:-[a-z]+)?\b")


def dist_targets():
    with open(CONFIG, "rb") as f:
        return tomllib.load(f)["dist"]["targets"]


def installer_triples(path):
    with open(path, encoding="utf-8") as f:
        text = f.read()
    # Only the assignments and returns that pick a triple, not prose that may name another one.
    picked = set()
    for line in text.splitlines():
        stripped = line.strip()
        if stripped.startswith("#"):
            continue
        if re.search(r'(TARGET=|return\s+")', stripped):
            picked.update(TRIPLE.findall(stripped))
    return picked


class PlatformTargetsTest(unittest.TestCase):
    def test_dist_targets_are_exactly_the_three_supported_platforms(self):
        self.assertEqual(sorted(dist_targets()), SUPPORTED, dist_targets())

    def test_no_target_is_listed_twice(self):
        targets = dist_targets()
        self.assertEqual(len(targets), len(set(targets)), targets)

    def test_each_installer_picks_only_triples_that_dist_builds(self):
        built = set(dist_targets())
        for path in INSTALLERS:
            with self.subTest(installer=os.path.basename(path)):
                picked = installer_triples(path)
                self.assertTrue(picked, f"{path} picks no target triple the test could read")
                self.assertLessEqual(picked, built, f"{path} picks a triple dist does not build")

    def test_the_installers_between_them_pick_every_dist_target(self):
        picked = set()
        for path in INSTALLERS:
            picked |= installer_triples(path)
        self.assertEqual(picked, set(dist_targets()))

    def test_no_dropped_platform_is_a_dist_target(self):
        for target in dist_targets():
            with self.subTest(target=target):
                self.assertNotIn("x86_64-apple-darwin", target)
                self.assertNotIn("aarch64-unknown-linux", target)


if __name__ == "__main__":
    unittest.main()
