"""Static checks for the Rust GitHub templates (Issue #2013, 08-B).

`templates/github/` is an addition, not a move: the development copies under `.github/`
stay where they are. These checks prove, offline, that the templates are byte-identical to
the verified development copies, that `manifest.json` records their real hashes, that the
forms follow the same schema rules as `tests/github_assets_test.py`, and that no unresolved
placeholder remains. They do not prove GitHub renders the forms: forms are read from the
default branch only, so rendering stays not_run (REQ-08B-10).
"""
import hashlib
import json
import pathlib
import re
import unittest

import yaml

ROOT = pathlib.Path(__file__).resolve().parent.parent
TEMPLATES = ROOT / "templates/github"
MANIFEST = TEMPLATES / "manifest.json"
FORMS = sorted((TEMPLATES / "ISSUE_TEMPLATE").glob("*.yml"))
ALLOWED_BLOCK_TYPES = {"markdown", "input", "textarea", "dropdown", "checkboxes"}
# `labels` is allowed: each form sets one `type:*` label (#2434 decision 1 = A).
FORBIDDEN_TOP_KEYS = {"assignees", "projects", "type"}
SECRET_SHAPES = re.compile(r"(ghp_|github_pat_|sk-[A-Za-z0-9]|-----BEGIN|/Users/|/home/|C:\\\\|PVT_)")
REQUIRED_FIELDS = {
    "template_path",
    "source_path",
    "target_path",
    "source_sha256",
    "template_version",
    "license",
    "variables",
    "generation_condition",
    "managed_scope",
}


def sha256(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def manifest():
    return json.loads(MANIFEST.read_text(encoding="utf-8"))


class ManifestTest(unittest.TestCase):
    def test_every_template_file_is_listed_once_with_all_fields(self):
        data = manifest()
        self.assertEqual(data["schema_version"], 1)
        listed = [e["template_path"] for e in data["files"]]
        self.assertEqual(len(listed), len(set(listed)))
        on_disk = sorted(
            str(p.relative_to(ROOT)) for p in TEMPLATES.rglob("*") if p.is_file() and p != MANIFEST
        )
        self.assertEqual(sorted(listed), on_disk)
        targets = [e["target_path"] for e in data["files"]]
        self.assertEqual(len(targets), len(set(targets)))
        for entry in data["files"]:
            with self.subTest(template=entry["template_path"]):
                self.assertEqual(set(entry), REQUIRED_FIELDS)
                self.assertRegex(entry["source_sha256"], r"^[0-9a-f]{64}$")
                self.assertTrue(entry["target_path"].startswith(".github/"))
                self.assertNotIn("..", entry["target_path"])

    def test_source_sha256_matches_the_development_copy(self):
        # AC-08B-2: drift between the development copy and the recorded hash fails here.
        for entry in manifest()["files"]:
            with self.subTest(source=entry["source_path"]):
                self.assertEqual(sha256(ROOT / entry["source_path"]), entry["source_sha256"])

    def test_templates_without_variables_are_byte_identical(self):
        # AC-08B-1 / REQ-08B-02: no variables, so parity is byte identity.
        for entry in manifest()["files"]:
            with self.subTest(template=entry["template_path"]):
                self.assertEqual(entry["variables"], [])
                self.assertEqual(
                    (ROOT / entry["template_path"]).read_bytes(),
                    (ROOT / entry["source_path"]).read_bytes(),
                )

    def test_no_unresolved_placeholder(self):
        # REQ-08B-08: the renderer refuses a placeholder left in rendered bytes.
        pattern = manifest()["placeholder_pattern"]
        self.assertTrue(pattern)
        for entry in manifest()["files"]:
            text = (ROOT / entry["template_path"]).read_text(encoding="utf-8")
            self.assertNotIn(pattern, text, entry["template_path"])


class TemplateFormTest(unittest.TestCase):
    def test_three_forms_are_promoted(self):
        self.assertEqual(
            [p.name for p in FORMS],
            ["moai-bugfix.yml", "moai-feature.yml", "moai-investigation.yml"],
        )

    def test_forms_follow_the_form_schema(self):
        # AC-08B-3 / REQ-08B-03: the rules of github_assets_test, on the template copies.
        for path in FORMS:
            with self.subTest(form=path.name):
                text = path.read_text(encoding="utf-8")
                form = yaml.safe_load(text)
                for key in ("name", "description", "body"):
                    self.assertIn(key, form)
                self.assertGreater(len(form["name"].strip()), 3)
                self.assertFalse(FORBIDDEN_TOP_KEYS & set(form))
                self.assertIsNone(SECRET_SHAPES.search(text))
                ids = []
                for block in form["body"]:
                    self.assertIn(block["type"], ALLOWED_BLOCK_TYPES)
                    if block["type"] == "markdown":
                        continue
                    self.assertRegex(block["id"], r"^[a-z][a-z0-9-]*$")
                    ids.append(block["id"])
                self.assertEqual(len(ids), len(set(ids)))
                redaction = [b for b in form["body"] if b.get("id") == "redaction"]
                self.assertEqual(len(redaction), 1)
                self.assertTrue(all(o.get("required") for o in redaction[0]["attributes"]["options"]))

    def test_label_template_is_secret_free(self):
        text = (TEMPLATES / "labels.yml").read_text(encoding="utf-8")
        self.assertEqual(yaml.safe_load(text)["schema_version"], 1)
        self.assertIsNone(SECRET_SHAPES.search(text))


if __name__ == "__main__":
    unittest.main()
