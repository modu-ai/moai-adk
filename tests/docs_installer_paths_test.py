"""The documents that name commands and paths name things that exist (SPEC-20261121-install-
docs-match-code): README.md, install/README.md and install/install.sh speak for the code — a
`moai` or `moai-cli` subcommand the docs print exists in the binary's own help, a repository
path they link is a file in the tree, the skill entry they name is a file under skills/moai,
and the state-folder sentence matches what `state_root` writes. The docs drifting again
turns one of these tests red, instead of waiting for the next install-flow report."""

import os
import re
import subprocess
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent


def rust_bin():
    explicit = os.environ.get("MOAI_CLI_BIN")
    if explicit:
        return Path(explicit)
    target = os.environ.get("CARGO_TARGET_DIR")
    return (Path(target) if target else REPO_ROOT / "target") / "debug" / "moai-cli"


def subcommands(binary):
    """Every word the binary's own `moai help --all` lists as a command: a command line is
    indented two spaces under a group heading and starts with the word (a launcher line in
    the root help spells several commands comma-separated, so the words here are the whole
    set, `moai cc` included)."""
    out = subprocess.run([str(binary), "help", "--all"], capture_output=True, text=True)
    names = set()
    for line in out.stdout.splitlines():
        if line.startswith("  ") and not line.startswith("   ") and line.strip():
            word = line.strip().split()[0].rstrip(",")
            if re.fullmatch(r"[a-z][a-z0-9-]*", word):
                names.add(word)
    return names


class DocsNameThingsThatExist(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.words = subcommands(rust_bin())
        cls.readme = (ROOT / "README.md").read_text(encoding="utf-8")
        cls.install_readme = (ROOT / "install" / "README.md").read_text(encoding="utf-8")
        cls.install_sh = (ROOT / "install" / "install.sh").read_text(encoding="utf-8")

    def test_repository_links_in_the_docs_exist(self):
        """A repository-relative markdown link in README or install/README points at a file of
        the tree — the docs never link a path the repository does not hold."""
        for name, text in (("README.md", self.readme),
                           ("install/README.md", self.install_readme)):
            for target in re.findall(r"\]\(([^)#\s]+\.md)\)", text):
                self.assertFalse(target.startswith("http"),
                                 f"{name}: an http address was read as a path: {target}")
                path = (ROOT / name).parent / target
                self.assertTrue(path.exists(),
                                f"{name} links {target}, which is not in the tree")

    def test_moai_subcommands_the_docs_name_exist_in_the_binary(self):
        """`moai <word>` and `moai-cli <word>` in README, install/README and install.sh name
        real subcommands of the binary's own help."""
        named = set()
        for text in (self.readme, self.install_readme, self.install_sh):
            for m in re.finditer(r"`(?:moai-cli|moai) ([a-z][a-z0-9-]+)", text):
                named.add(m.group(1))
        self.assertTrue(named,
                        "the fixtures went stale: the docs name no moai subcommand")
        unknown = sorted(named - self.words)
        self.assertEqual([], unknown,
                         f"the docs name subcommands the binary lacks: {unknown}")

    def test_the_skill_entry_the_docs_name_is_a_file(self):
        """/moai project (the entry README, install/README and install.sh name) is the project
        workflow file of the shipped skill."""
        self.assertTrue(
            (ROOT / "skills" / "moai" / "workflows" / "project.md").exists(),
            "/moai project is named, but skills/moai/workflows/project.md is missing",
        )

    def test_the_state_folder_sentence_matches_the_code(self):
        """README's state folders are the code's: a new project starts on `.moai/` and keeps
        `.moai-cli/` only when it began on the earlier RC (`state_root::create_dir`), while the
        user's own state stays `~/.moai-cli/` (`HOME_STATE_DIR`)."""
        self.assertIn(".moai/` (a new project; a project that began on the earlier RC keeps "
                      "`.moai-cli/`)", self.readme)
        self.assertIn("`~/.moai-cli/` (user)", self.readme)
        state_root = (ROOT / "src" / "state_root.rs").read_text(encoding="utf-8")
        self.assertIn("is_new_project", state_root)
        lib = (ROOT / "src" / "lib.rs").read_text(encoding="utf-8")
        self.assertIn('pub const HOME_STATE_DIR: &str = ".moai-cli"', lib)

    def test_the_installer_lines_the_docs_describe_exist(self):
        """install/README describes the question source the script really has (standard input,
        else `/dev/tty`) — the old `never shows the menu` sentence, which was wrong for the
        documented `curl … | sh` pipe, stays out."""
        self.assertIn("/dev/tty", self.install_sh)
        self.assertNotIn("never shows the menu", self.install_readme)


if __name__ == "__main__":
    unittest.main()
