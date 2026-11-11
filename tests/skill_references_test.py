"""Current repository skill-file references must resolve after package moves."""

from pathlib import Path
import re
import subprocess
import unittest

ROOT = Path(__file__).resolve().parent.parent
REFERENCE = re.compile(
    r"skills/[A-Za-z0-9_.-]+/(?:[A-Za-z0-9_.-]+/)*"
    r"[A-Za-z0-9_.-]+\.[A-Za-z0-9_-]+"
)


def tracked_texts():
    names = subprocess.check_output(
        ["git", "-C", str(ROOT), "ls-files", "-z"]
    ).decode().split("\0")
    for name in names:
        if not name:
            continue
        try:
            yield name, (ROOT / name).read_text(encoding="utf-8")
        except UnicodeError:
            continue


def reference_kind(name, line, start):
    if name.startswith(".moai-cli/"):
        # Versioned legacy inventories and Issue-body mirrors describe other trees.
        return "ledger-snapshot"
    if name == "docs/product-naming-inventory.md":
        # Its explicitly pinned measurement and dated #2405 annotations are history,
        # not instructions for loading the current skill package.
        return "dated-inventory"
    if start > 0 and line[start - 1] == "/":
        if "github.com/modu-ai/moai-adk/blob/" not in line[:start]:
            # Host projection paths and temporary-fixture paths are not repo paths.
            return "runtime-or-example"
    return "current"


def references():
    for name, text in tracked_texts():
        for number, line in enumerate(text.splitlines(), 1):
            for match in REFERENCE.finditer(line):
                yield (name, number, match.group(),
                       reference_kind(name, line, match.start()))


class SkillReferences(unittest.TestCase):
    def test_current_pr_template_has_a_resolving_workflow_reference(self):
        text = (ROOT / ".github/pull_request_template.md").read_text(encoding="utf-8")
        workflow_refs = [m.group() for m in REFERENCE.finditer(text)
                         if "/workflows/" in m.group()]
        self.assertTrue(workflow_refs, "the PR template must identify its check rules")
        for reference in workflow_refs:
            with self.subTest(reference=reference):
                self.assertTrue((ROOT / reference).is_file(), reference)

    def test_current_tracked_skill_file_references_resolve(self):
        current = [(name, number, reference) for name, number, reference, kind
                   in references() if kind == "current"]
        self.assertTrue(current, "the tracked reference inventory must not be empty")
        broken = [(name, number, reference) for name, number, reference in current
                  if not (ROOT / reference).is_file()]
        self.assertEqual(broken, [], "current repository skill references are broken")


if __name__ == "__main__":
    unittest.main()
