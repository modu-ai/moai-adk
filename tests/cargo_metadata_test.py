"""Cargo.toml package metadata matches the repository's own facts (Issue #2500, #2435 decision 5-A).

The four `[package]` fields `license`, `repository`, `homepage` and `description` carry no
value of their own: each one is read here from the file that states it (LICENSE, the README
Facts block, the README opening sentence), so the manifest cannot drift from what the
repository says about itself. `publish = false` stays; publishing is out of scope.
"""
import pathlib
import re
import tomllib
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
README = (ROOT / "README.md").read_text(encoding="utf-8").splitlines()


def package():
    return tomllib.loads((ROOT / "Cargo.toml").read_text(encoding="utf-8"))["package"]


def fact(label):
    """The value of one `- <label>: <value>` line in the README Facts block."""
    found = [line for line in README if line.startswith(f"- {label}: ")]
    if len(found) != 1:
        raise AssertionError(f"README Facts must have exactly one {label!r} line, found {len(found)}")
    return found[0][len(f"- {label}: "):]


def first_url(text):
    match = re.match(r"(https://\S+)", text)
    if not match:
        raise AssertionError(f"no leading https URL in {text!r}")
    return match.group(1)


class CargoPackageMetadataTest(unittest.TestCase):
    def test_license_is_the_spdx_id_of_the_license_file_and_readme(self):
        text = (ROOT / "LICENSE").read_text(encoding="utf-8")
        head = [line.strip() for line in text.splitlines() if line.strip()][:2]
        self.assertEqual(head, ["Apache License", "Version 2.0, January 2004"])
        self.assertEqual(fact("License").split(" ", 1)[0], "Apache-2.0")
        self.assertEqual(package().get("license"), "Apache-2.0")

    def test_repository_is_the_readme_repository_fact(self):
        self.assertEqual(package().get("repository"), first_url(fact("Repository")))

    def test_homepage_is_the_readme_website_fact(self):
        self.assertEqual(package().get("homepage"), first_url(fact("Website")))

    def test_description_is_the_readme_opening_sentence_without_markdown(self):
        opening = next(line for line in README[1:] if line and not line.startswith(("#", "**")))
        self.assertEqual(package().get("description"), opening.replace("`", ""))

    def test_publish_stays_false(self):
        self.assertIs(package().get("publish"), False)


if __name__ == "__main__":
    unittest.main()
