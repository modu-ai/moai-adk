"""SPEC-20261117-dist-include-skill: a release archive carries the skill package and the base
agents beside the binary.

`moai update` reads them from the version folder, `<version>/skills/moai` and `<version>/agents`
(src/update.rs `installed_package_dir`, `installed_agents_dir`), and the installers copy the
archive's `skills/` and `agents/` folders there. cargo-dist's `include` copies each listed file
or folder into the root of every archive, beside the binary; paths are relative to the config
file and globs are not supported (https://axodotdev.github.io/cargo-dist/book/reference/config.html,
read 2026-11-17T17:47Z). This test reads the config and the folders it names; it does not run
cargo-dist.
"""

import os
import tomllib
import unittest

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CONFIG = os.path.join(REPO_ROOT, "dist-workspace.toml")


def dist_config():
    with open(CONFIG, "rb") as f:
        return tomllib.load(f)["dist"]


class DistIncludeTest(unittest.TestCase):
    def test_the_archive_root_gets_the_skills_and_agents_folders(self):
        include = dist_config().get("include", [])
        # cargo-dist puts each entry at the archive root under its own name: `skills/` keeps the
        # `skills/moai` layout the version folder needs, `agents/` lands as `agents/`.
        self.assertEqual(sorted(include), ["agents/", "skills/"], include)

    def test_each_included_folder_is_a_real_folder_of_this_repository(self):
        for entry in dist_config().get("include", []):
            with self.subTest(entry=entry):
                self.assertNotIn("*", entry, "cargo-dist does not support globs")
                self.assertFalse(entry.startswith(("/", "..")), entry)
                path = os.path.join(os.path.dirname(CONFIG), entry)
                self.assertTrue(os.path.isdir(path), path)
                self.assertFalse(os.path.islink(path.rstrip("/")), path)

    def test_the_skills_folder_holds_only_the_moai_skill_package(self):
        # Everything under `skills/` ships; the package `moai update` stages is `skills/moai`.
        skills = os.path.join(REPO_ROOT, "skills")
        self.assertEqual(sorted(os.listdir(skills)), ["moai"])
        self.assertTrue(os.path.isfile(os.path.join(skills, "moai", "SKILL.md")))

    def test_the_agents_folder_holds_agent_sources(self):
        agents = os.path.join(REPO_ROOT, "agents")
        names = sorted(os.listdir(agents))
        self.assertTrue(names, agents)
        for name in names:
            with self.subTest(name=name):
                self.assertTrue(name.endswith(".md"), name)
                self.assertTrue(os.path.isfile(os.path.join(agents, name)), name)


if __name__ == "__main__":
    unittest.main()
