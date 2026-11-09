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


WORKFLOWS = ROOT / ".github/workflows"
PR_TEMPLATE = ROOT / ".github/pull_request_template.md"
PINNED_USES = re.compile(r"^[\w.-]+/[\w./-]+@[0-9a-f]{40}$")
# The only write scope a workflow may hold: rc-label-sync's manual apply job creates labels.
ALLOWED_WRITES = {("rc-label-sync.yml", "apply"): {"issues"}}


class WorkflowTest(unittest.TestCase):
    def workflows(self):
        paths = sorted(WORKFLOWS.glob("*.yml"))
        self.assertEqual([p.name for p in paths], ["rc-ci.yml", "rc-label-sync.yml"])
        return [(p, yaml.safe_load(p.read_text(encoding="utf-8"))) for p in paths]

    def test_permissions_are_least_privilege(self):
        for path, flow in self.workflows():
            with self.subTest(workflow=path.name):
                self.assertEqual(flow["permissions"], {"contents": "read"})
                for job_id, job in flow["jobs"].items():
                    self.assertIn("permissions", job, "every job states its own permissions")
                    writes = {k for k, v in job["permissions"].items() if v == "write"}
                    self.assertEqual(writes, ALLOWED_WRITES.get((path.name, job_id), set()), job_id)

    def test_every_action_is_pinned_to_a_commit_sha(self):
        for path, flow in self.workflows():
            for job_id, job in flow["jobs"].items():
                for step in job.get("steps", []):
                    if "uses" in step:
                        with self.subTest(workflow=path.name, job=job_id, uses=step["uses"]):
                            self.assertRegex(step["uses"], PINNED_USES)

    def test_rc_ci_reports_the_rc_ci_check_on_prs_to_the_integration_branch(self):
        flow = dict(self.workflows())[WORKFLOWS / "rc-ci.yml"]
        triggers = flow.get("on", flow.get(True))  # PyYAML reads a bare `on:` key as True
        self.assertEqual(triggers["pull_request"]["branches"], ["v2026.11.0"])
        self.assertEqual(flow["jobs"]["rc-ci"]["name"], "rc-ci")
        # The check set runs on every matrix target (#2194); `rc-ci` aggregates them.
        script = "\n".join(s.get("run", "") for s in flow["jobs"]["check"]["steps"])
        for command in ("cargo test --offline --locked", "cargo clippy --offline --locked --all-targets -- -D warnings",
                        "cargo fmt --check", "unittest discover -v -s tests -p '*_test.py'", "git diff --check",
                        "GITHUB_STEP_SUMMARY"):
            self.assertIn(command, script)

    def test_rc_ci_runs_the_four_decided_targets_and_fails_when_any_fails(self):
        """#2194 (decision 01:1x(11-09)): Linux x86_64/arm64 and macOS arm64/x86_64 on standard
        free GitHub-hosted runner labels; a failing target fails `rc-ci` instead of skipping it."""
        flow = dict(self.workflows())[WORKFLOWS / "rc-ci.yml"]
        check = flow["jobs"]["check"]
        self.assertIs(check["strategy"]["fail-fast"], False)
        targets = {e["target"]: e["runner"] for e in check["strategy"]["matrix"]["include"]}
        self.assertEqual(targets, {
            "linux-x86_64": "ubuntu-24.04",
            "linux-arm64": "ubuntu-24.04-arm",
            "macos-arm64": "macos-15",
            "macos-x86_64": "macos-15-intel",
        })
        self.assertEqual(check["runs-on"], "${{ matrix.runner }}")
        aggregate = flow["jobs"]["rc-ci"]
        self.assertEqual(aggregate["needs"], ["check"])
        # #2312: runs after a failed target (so the check fails, not skips) but not after a
        # cancellation (always() reported a cancel-in-progress run as a failure).
        self.assertEqual(aggregate.get("if"), "${{ !cancelled() }}")
        gate = "\n".join(s.get("run", "") for s in aggregate["steps"])
        self.assertIn('[ "${{ needs.check.result }}" = "success" ]', gate)

    def test_label_sync_is_a_dry_run_unless_applied_by_hand(self):
        flow = dict(self.workflows())[WORKFLOWS / "rc-label-sync.yml"]
        triggers = flow.get("on", flow.get(True))
        self.assertIs(triggers["workflow_dispatch"]["inputs"]["apply"]["default"], False)
        self.assertIn("workflow_dispatch", flow["jobs"]["apply"]["if"])
        text = (WORKFLOWS / "rc-label-sync.yml").read_text(encoding="utf-8")
        self.assertIn(".github/moai/labels.yml", text)
        self.assertIsNone(re.search(r"gh label (delete|edit)|--force", text), "no label edits or deletes")


class PullRequestTemplateTest(unittest.TestCase):
    def test_template_carries_the_rc1_pr_contract(self):
        text = PR_TEMPLATE.read_text(encoding="utf-8")
        for field in ("Refs #", "Moai-Claim:", "exact", "40자 SHA", "실행한 검사와 exit", "독립 review",
                      "not_run", "blocked", "--body-file"):
            self.assertIn(field, text)
        self.assertIsNone(re.search(r"(?im)^\s*(closes|fixes|resolves) #", text), "no auto-close keyword")
        self.assertIsNone(SECRET_SHAPES.search(text))


class PreviousGenerationTest(unittest.TestCase):
    def removed(self):
        lines = (ROOT / ".github/moai/removed-go-generation-paths.txt").read_text(encoding="utf-8").splitlines()
        return [line for line in lines if line.strip()]

    def test_previous_generation_files_are_not_recreated(self):
        removed = self.removed()
        self.assertEqual(len(removed), 23)
        self.assertEqual(len(set(removed)), 23)
        self.assertEqual([p for p in removed if (ROOT / p).exists()], [])

    def test_readme_records_the_same_removed_paths(self):
        readme = (ROOT / ".github/moai/README.md").read_text(encoding="utf-8")
        block = re.search(r"Removed paths:\n\n```text\n(.*?)\n```", readme, re.S)
        self.assertIsNotNone(block)
        self.assertEqual(sorted(block.group(1).splitlines()), sorted(self.removed()))


if __name__ == "__main__":
    unittest.main()
