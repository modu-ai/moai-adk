"""Static checks for the Rust GitHub templates (Issue #2013, 08-B; contract replaced by #2434 AC-6).

`templates/github/` holds the product templates a generated project receives; the copies under
`.github/` belong to this repository. Since #2434 AC-6 (design 6788185980, leader decisions
6842518327) the two are no longer byte-identical: a product form keeps the structure of the
repository form it derives from (P1), `manifest.json` records each template's own hash (P2), no
template carries this repository's own state or identity (P3), and every label a form sets
exists in the product label list, which is itself a subset of this repository's labels (P4).
They do not prove GitHub renders the forms: forms are read from the default branch only, so
rendering stays not_run (REQ-08B-10).
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
# P3: text that only describes this repository (its name, owner, Issue numbers, release line,
# a home path) never ships.
LEAKS = re.compile(r"(modu-ai|moai-adk|#[0-9]{3,}|v2026|/Users/|/home/)")
REQUIRED_FIELDS = {
    "template_path",
    "derived_from",
    "target_path",
    "template_sha256",
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
        self.assertEqual(data["schema_version"], 2)
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
                self.assertRegex(entry["template_sha256"], r"^[0-9a-f]{64}$")
                self.assertTrue(entry["target_path"].startswith(".github/"))
                self.assertNotIn("..", entry["target_path"])

    def test_template_sha256_is_the_template_itself(self):
        # P2: a template changed without its manifest entry fails here.
        for entry in manifest()["files"]:
            with self.subTest(template=entry["template_path"]):
                self.assertEqual(sha256(ROOT / entry["template_path"]), entry["template_sha256"])

    def test_derived_forms_keep_the_structure_of_their_repository_form(self):
        # P1 (replaces the byte identity of AC-08B-1 / REQ-08B-02): the same blocks in the same
        # order — type, id, required, dropdown option count — and the same `labels:`.
        def shape(path):
            form = yaml.safe_load(path.read_text(encoding="utf-8"))
            blocks = [
                (
                    b["type"],
                    b.get("id"),
                    bool(b.get("validations", {}).get("required")),
                    len(b.get("attributes", {}).get("options", []) or []),
                )
                for b in form["body"]
            ]
            return blocks, form.get("labels")

        derived = [e for e in manifest()["files"] if e["derived_from"]]
        self.assertEqual(len(derived), len(FORMS))
        for entry in derived:
            with self.subTest(template=entry["template_path"]):
                self.assertEqual(entry["variables"], [])
                self.assertEqual(shape(ROOT / entry["template_path"]), shape(ROOT / entry["derived_from"]))

    def test_no_unresolved_placeholder(self):
        # REQ-08B-08: the renderer refuses a placeholder left in rendered bytes.
        pattern = manifest()["placeholder_pattern"]
        self.assertTrue(pattern)
        for entry in manifest()["files"]:
            text = (ROOT / entry["template_path"]).read_text(encoding="utf-8")
            self.assertNotIn(pattern, text, entry["template_path"])


class TemplateFormTest(unittest.TestCase):
    def test_four_forms_are_promoted(self):
        # Decision 43 (6842518327): the follow-up form ships as the fourth product form.
        self.assertEqual(
            [p.name for p in FORMS],
            ["moai-bugfix.yml", "moai-feature.yml", "moai-followup.yml", "moai-investigation.yml"],
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
        self.assertEqual(yaml.safe_load(text)["schema_version"], 2)
        self.assertIsNone(SECRET_SHAPES.search(text))


def product_labels():
    return yaml.safe_load((TEMPLATES / "labels.yml").read_text(encoding="utf-8"))


class NoLeakTest(unittest.TestCase):
    def test_no_template_carries_this_repository(self):
        # P3: nothing that is true only for this repository ships.
        for entry in manifest()["files"]:
            with self.subTest(template=entry["template_path"]):
                text = (ROOT / entry["template_path"]).read_text(encoding="utf-8")
                self.assertIsNone(LEAKS.search(text), LEAKS.search(text))

    def test_product_labels_are_one_list_without_remote_state_or_area_seeds(self):
        # P3: `reused`/`not_used` describe this remote; a new repository has no labels yet,
        # so every label is listed with its colour and description. Decision 42 (6842518327):
        # no `area:` label is seeded — a project defines its own areas.
        data = product_labels()
        self.assertNotIn("reused", data)
        self.assertNotIn("not_used", data)
        names = [label["name"] for label in data["labels"]]
        self.assertEqual(len(names), len(set(names)))
        for label in data["labels"]:
            with self.subTest(label=label["name"]):
                self.assertRegex(label["color"], r"^[0-9a-f]{6}$")
                self.assertTrue(label["description"].strip())
        self.assertFalse([n for n in names if n.startswith("area:")])
        self.assertEqual(
            sorted(n for n in names if n.startswith("type:")),
            sorted(
                f"type:{t}"
                for t in ("feature", "fix", "docs", "chore", "ci", "test", "refactor", "design")
            ),
        )
        self.assertEqual(
            sorted(n for n in names if n.startswith("priority:")),
            ["priority:P0", "priority:P1", "priority:P2", "priority:P3"],
        )
        self.assertIn("needs:decision", names)


class LabelCoherenceTest(unittest.TestCase):
    def test_form_labels_exist_and_product_labels_are_this_repository_labels(self):
        # P4: a form label missing from the list is silently dropped by GitHub, and this
        # repository is a MoAI project too, so it carries every product label.
        names = {label["name"] for label in product_labels()["labels"]}
        for path in FORMS:
            with self.subTest(form=path.name):
                for label in yaml.safe_load(path.read_text(encoding="utf-8")).get("labels", []):
                    self.assertIn(label, names)
        repo = yaml.safe_load((ROOT / ".github/moai/labels.yml").read_text(encoding="utf-8"))
        # `reused` maps each axis to label names; `added` lists labels with colour and text.
        repo_names = {name for names in repo["reused"].values() for name in names}
        repo_names |= {entry["name"] for entry in repo["added"]}
        self.assertFalse(names - repo_names, sorted(names - repo_names))


if __name__ == "__main__":
    unittest.main()
