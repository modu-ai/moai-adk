"""Local structural checks for the RC1 GitHub assets (Issue forms and label taxonomy).

These checks run offline against repository files. They do not prove that GitHub renders the
forms (forms are read from the default branch only) or that the live remote labels match; the label
manifest is compared offline against the snapshot that 00-B (#1986) recorded, and a live
comparison is a separate, explicitly authorized `gh` step recorded in the Issue report.
"""
import json
import pathlib
import re
import unittest

import yaml

ROOT = pathlib.Path(__file__).resolve().parent.parent
FORMS = [
    ROOT / ".github/ISSUE_TEMPLATE/moai-feature.yml",
    ROOT / ".github/ISSUE_TEMPLATE/moai-bugfix.yml",
    ROOT / ".github/ISSUE_TEMPLATE/moai-investigation.yml",
]
LABELS = ROOT / ".github/moai/labels.yml"
REMOTE_LABELS = ROOT / "tests/fixtures/github/labels.json"
ALLOWED_BLOCK_TYPES = {"markdown", "input", "textarea", "dropdown", "checkboxes"}
FORBIDDEN_TOP_KEYS = {"labels", "assignees", "projects", "type"}
SECRET_SHAPES = re.compile(r"(ghp_|github_pat_|sk-[A-Za-z0-9]|-----BEGIN|/Users/|C:\\\\)")


class IssueFormTest(unittest.TestCase):
    def test_forms_parse_and_follow_the_form_schema(self):
        for path in FORMS:
            with self.subTest(form=path.name):
                text = path.read_text(encoding="utf-8")
                form = yaml.safe_load(text)
                self.assertIsInstance(form, dict)
                for key in ("name", "description", "body"):
                    self.assertIn(key, form)
                self.assertTrue(form["name"].strip() and form["description"].strip())
                self.assertFalse(FORBIDDEN_TOP_KEYS & set(form), "forms must not set labels, assignees, projects or type")
                self.assertIn("moai-template:", text.splitlines()[0])
                self.assertIsNone(SECRET_SHAPES.search(text))
                ids = []
                non_markdown = 0
                for block in form["body"]:
                    self.assertIn(block["type"], ALLOWED_BLOCK_TYPES)
                    self.assertIn("attributes", block)
                    if block["type"] == "markdown":
                        self.assertIn("value", block["attributes"])
                        continue
                    non_markdown += 1
                    self.assertIn("id", block)
                    self.assertRegex(block["id"], r"^[a-z][a-z0-9-]*$")
                    ids.append(block["id"])
                    self.assertIn("label", block["attributes"])
                    if block["type"] == "dropdown":
                        options = block["attributes"]["options"]
                        self.assertGreaterEqual(len(options), 2)
                        self.assertEqual(len(options), len(set(options)))
                    if block["type"] == "checkboxes":
                        for option in block["attributes"]["options"]:
                            self.assertIn("label", option)
                self.assertEqual(len(ids), len(set(ids)), "block ids must be unique")
                self.assertGreaterEqual(non_markdown, 1, "a form needs at least one input block")

    def test_redaction_checkbox_is_required_in_every_form(self):
        for path in FORMS:
            form = yaml.safe_load(path.read_text(encoding="utf-8"))
            redaction = [b for b in form["body"] if b.get("id") == "redaction"]
            self.assertEqual(len(redaction), 1)
            self.assertTrue(all(o.get("required") for o in redaction[0]["attributes"]["options"]))

    def test_checklist_starts_unchecked_and_decisions_start_on_hold(self):
        # REQ-00C-02: a new Issue never arrives with a ticked item or a decided choice.
        for path in FORMS:
            with self.subTest(form=path.name):
                blocks = {b.get("id"): b for b in yaml.safe_load(path.read_text(encoding="utf-8"))["body"]}
                items = [line for line in blocks["checklist"]["attributes"]["value"].splitlines() if line.strip()]
                self.assertTrue(items)
                self.assertTrue(all(line.startswith("- [ ] ") for line in items), items)
                decisions = blocks["decisions"]["attributes"]["value"]
                states = re.findall(r"상태: (\S+)", decisions)
                self.assertTrue(states)
                self.assertEqual(set(states), {"HOLD"})


class LabelTaxonomyTest(unittest.TestCase):
    def test_label_manifest_is_consistent(self):
        data = yaml.safe_load(LABELS.read_text(encoding="utf-8"))
        self.assertEqual(data["schema_version"], 1)
        reused = [name for group in data["reused"].values() for name in group]
        added = [item["name"] for item in data["added"]]
        proposed = [item["name"] for item in data["proposed_not_created"]]
        everything = reused + added + proposed
        self.assertEqual(len(everything), len(set(everything)), "a label appears in one group only")
        for item in data["added"] + data["proposed_not_created"]:
            self.assertRegex(item["color"], r"^[0-9a-f]{6}$")
            self.assertTrue(item["description"].strip())
            self.assertLessEqual(len(item["description"]), 100)
        self.assertFalse(any(name.startswith("status:") for name in everything), "status is never a label")

    def test_label_manifest_matches_the_recorded_remote(self):
        # REQ-00C-03 / REQ-00C-08, offline: compare against the 00-B (#1986) recorded label snapshot.
        # A failure here means the remote moved or the manifest drifted; neither is fixed by editing
        # the snapshot by hand.
        data = yaml.safe_load(LABELS.read_text(encoding="utf-8"))
        remote = {item["name"]: item for item in json.loads(REMOTE_LABELS.read_text(encoding="utf-8"))}
        reused = [name for group in data["reused"].values() for name in group]
        self.assertEqual([n for n in reused if n not in remote], [], "reused labels must exist remotely")
        for item in data["added"]:
            with self.subTest(label=item["name"]):
                self.assertIn(item["name"], remote)
                self.assertEqual(remote[item["name"]]["color"], item["color"])
                self.assertEqual(remote[item["name"]]["description"], item["description"])
        self.assertEqual([p["name"] for p in data["proposed_not_created"] if p["name"] in remote], [])


if __name__ == "__main__":
    unittest.main()
