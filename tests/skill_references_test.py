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
# Folders under docs/ whose files are dated records (design reviews, surveys).
HISTORY_DOCS = ("design-review", "surveys")


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
    top, _, rest = name.partition("/")
    if (name.startswith(".moai/specs/")
            or (top == "docs" and rest.split("/")[0] in HISTORY_DOCS and "/" in rest)
            or (top == "docs" and re.fullmatch(r"decisions-[^/]*\.md", rest))):
        # SPEC-20261122-next-flow-skills-agents: closed SPECs and dated docs record past states.
        return "history"
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


DETECT = ROOT / "src/project_detect.rs"
QUALITY = ROOT / "skills/moai/references/internal/quality.md"
LANGUAGES = ROOT / "skills/moai/references/languages"
GUIDE_ROW = re.compile(r"^\| ([a-z]+) \| languages/([a-z-]+\.md) \|$", re.M)
GUIDE_LIMIT = 4096


def detector_languages():
    """Every language `project-detect` can report: the language side of its manifest,
    extension and language-server tables."""
    text = DETECT.read_text(encoding="utf-8")
    found = set()
    for const, side in (("MANIFESTS", 1), ("EXTENSIONS", 1), ("LSP", 0)):
        block = re.search(
            r"const " + const + r": &\[\(&str, &str\)\] = &\[(.*?)\];", text, re.S
        )
        assert block, const + " table not found in src/project_detect.rs"
        pairs = re.findall(r'\("([^"]+)", "([^"]+)"\)', block.group(1))
        assert pairs, const + " table is empty"
        found.update(pair[side] for pair in pairs)
    return found


class QualityGuides(unittest.TestCase):
    """SPEC-20261117-quality-guides AC-2: internal/quality names one guide per detected
    language, every named guide exists, and no guide is left unnamed."""

    def rows(self):
        return dict(GUIDE_ROW.findall(QUALITY.read_text(encoding="utf-8")))

    def test_every_detector_language_has_a_guide_row(self):
        languages = detector_languages()
        self.assertIn("rust", languages)
        self.assertEqual(sorted(self.rows()), sorted(languages))

    def test_every_named_guide_exists_and_every_guide_is_named(self):
        named = set(self.rows().values())
        present = {p.name for p in LANGUAGES.glob("*.md")}
        self.assertEqual(sorted(named), sorted(present))

    def test_guides_are_short_and_point_back_to_quality(self):
        for path in sorted(LANGUAGES.glob("*.md")):
            with self.subTest(guide=path.name):
                data = path.read_bytes()
                self.assertLessEqual(len(data), GUIDE_LIMIT)
                text = data.decode("utf-8")
                self.assertTrue(text.startswith("# Language: "))
                self.assertIn("When to read: `internal/quality` sent you here", text)


if __name__ == "__main__":
    unittest.main()
