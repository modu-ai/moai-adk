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
    ROOT / ".github/ISSUE_TEMPLATE/moai-followup.yml",
]
# Decision 1 = A (#2434, Issue comment 6783612809): each form sets exactly one `type:*` label,
# the follow-up form sets none (its type follows the original work).
FORM_TYPE_LABEL = {
    "moai-feature.yml": "type:feature",
    "moai-bugfix.yml": "type:fix",
    "moai-investigation.yml": "type:design",
    "moai-followup.yml": None,
}
CONFIG = ROOT / ".github/ISSUE_TEMPLATE/config.yml"
LABELS = ROOT / ".github/moai/labels.yml"
REMOTE_LABELS = ROOT / "tests/fixtures/github/labels.json"
ALLOWED_BLOCK_TYPES = {"markdown", "input", "textarea", "dropdown", "checkboxes"}
FORBIDDEN_TOP_KEYS = {"assignees", "projects", "type"}
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
                self.assertFalse(FORBIDDEN_TOP_KEYS & set(form), "forms must not set assignees, projects or type")
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


    def test_form_set_is_exactly_the_decided_four(self):
        on_disk = sorted(p.name for p in (ROOT / ".github/ISSUE_TEMPLATE").glob("*.yml") if p != CONFIG)
        self.assertEqual(on_disk, sorted(FORM_TYPE_LABEL))
        self.assertEqual(sorted(p.name for p in FORMS), sorted(FORM_TYPE_LABEL))

    def test_each_form_sets_at_most_its_one_decided_type_label(self):
        manifest = yaml.safe_load(LABELS.read_text(encoding="utf-8"))
        known = {n for group in manifest["reused"].values() for n in group} | {i["name"] for i in manifest["added"]}
        for path in FORMS:
            with self.subTest(form=path.name):
                form = yaml.safe_load(path.read_text(encoding="utf-8"))
                expected = FORM_TYPE_LABEL[path.name]
                if expected is None:
                    self.assertNotIn("labels", form)
                    continue
                self.assertEqual(form["labels"], [expected])
                # GitHub silently drops a form label that does not exist, so it must be in the manifest.
                self.assertIn(expected, known)
                self.assertTrue(expected.startswith("type:"))

    def test_forms_carry_the_agent_ready_fields(self):
        # #2434 AC-2: one-line outcome, scope with non-scope and search hints, AC with a
        # verification, prerequisites, HOLD decisions and a done definition.
        for path in FORMS:
            with self.subTest(form=path.name):
                blocks = {b.get("id"): b for b in yaml.safe_load(path.read_text(encoding="utf-8"))["body"]}
                outcome = blocks["outcome"]
                self.assertEqual(outcome["type"], "input")
                self.assertIs(outcome["validations"]["required"], True)
                self.assertEqual(list(blocks)[2], "outcome", "the one-line outcome comes right after redaction")
                scope = blocks["scope"]["attributes"]["value"]
                for field in ("비범위:", "탐색 단서:"):
                    self.assertIn(field, scope)
                self.assertIs(blocks["scope"]["validations"]["required"], True)
                acceptance = blocks["acceptance"]["attributes"]["value"]
                self.assertIs(blocks["acceptance"]["validations"]["required"], True)
                for line in (l for l in acceptance.splitlines() if l.strip()):
                    self.assertRegex(line, r"^- AC-\d+: .* · 검증: .* · 결과: not_run$")
                self.assertEqual(blocks["dependencies"]["type"], "input")
                self.assertEqual(blocks["checklist"]["attributes"]["label"], "완료 정의")
                self.assertIn("독립 AUDIT", blocks["checklist"]["attributes"]["value"])
                self.assertNotIn("non-goals", blocks, "non-goals are merged into scope")

    def test_boundary_behaviour_is_asked_for_code_changes(self):
        for name in ("moai-feature.yml", "moai-bugfix.yml", "moai-followup.yml"):
            with self.subTest(form=name):
                blocks = {b.get("id"): b for b in yaml.safe_load(
                    (ROOT / ".github/ISSUE_TEMPLATE" / name).read_text(encoding="utf-8"))["body"]}
                self.assertIn("경계·실패 동작:", blocks["acceptance"]["attributes"]["value"])

    def test_followup_form_links_the_original_work(self):
        blocks = {b.get("id"): b for b in yaml.safe_load(
            (ROOT / ".github/ISSUE_TEMPLATE/moai-followup.yml").read_text(encoding="utf-8"))["body"]}
        for block_id in ("origin", "remaining"):
            self.assertIs(blocks[block_id]["validations"]["required"], True)


class TemplateChooserConfigTest(unittest.TestCase):
    def test_config_disables_blank_issues_and_links_https_only(self):
        text = CONFIG.read_text(encoding="utf-8")
        config = yaml.safe_load(text)
        self.assertEqual(set(config), {"blank_issues_enabled", "contact_links"})
        self.assertIs(config["blank_issues_enabled"], False)
        self.assertTrue(config["contact_links"])
        for link in config["contact_links"]:
            with self.subTest(link=link.get("name")):
                self.assertEqual(set(link), {"name", "url", "about"})
                self.assertTrue(link["url"].startswith("https://"))
                self.assertTrue(link["name"].strip() and link["about"].strip())
        self.assertIsNone(SECRET_SHAPES.search(text))


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
        not_used = [name for group in data["not_used"].values() for name in group]
        self.assertEqual(len(not_used), len(set(not_used)))
        self.assertEqual(set(not_used) & set(everything), set(), "a label is either used or not_used")

    def test_rc_axes_follow_the_decided_taxonomy(self):
        # #2434 AC-3: type, priority P0..P3, area, host and one gate marker; no state labels.
        data = yaml.safe_load(LABELS.read_text(encoding="utf-8"))
        rc = [name for group in data["reused"].values() for name in group] + [i["name"] for i in data["added"]]
        prefixes = {name.split(":", 1)[0] for name in rc if ":" in name}
        self.assertEqual(prefixes, {"type", "priority", "area", "host", "needs"})
        self.assertEqual(sorted(n for n in rc if n.startswith("priority:")),
                         ["priority:P0", "priority:P1", "priority:P2", "priority:P3"])
        self.assertEqual([n for n in rc if n.startswith("needs:")], ["needs:decision"])
        self.assertIn("type:design", rc)
        self.assertFalse({"blocked", "in-progress", "ready", "review", "done", "kind:epic"} & set(rc))
        self.assertEqual(data["proposed_not_created"], [])

    def test_label_manifest_matches_the_recorded_remote(self):
        # REQ-00C-03 / REQ-00C-08, offline: compare against the 00-B (#1986) recorded label snapshot.
        # A failure here means the remote moved or the manifest drifted; neither is fixed by editing
        # the snapshot by hand.
        data = yaml.safe_load(LABELS.read_text(encoding="utf-8"))
        remote = {item["name"]: item for item in json.loads(REMOTE_LABELS.read_text(encoding="utf-8"))}
        reused = [name for group in data["reused"].values() for name in group]
        self.assertEqual([n for n in reused if n not in remote], [], "reused labels must exist remotely")
        # An `added` label is either already remote with the same colour and description, or not
        # remote yet: `rc-label-sync` apply only creates missing ones and never edits, so a
        # remote label that differs would never converge.
        for item in data["added"]:
            with self.subTest(label=item["name"]):
                if item["name"] in remote:
                    self.assertEqual(remote[item["name"]]["color"], item["color"])
                    self.assertEqual(remote[item["name"]]["description"], item["description"])
        self.assertEqual([p["name"] for p in data["proposed_not_created"] if p["name"] in remote], [])
        # Decision 3 = A: previous-generation labels are documented as not used, never deleted.
        not_used = [name for group in data["not_used"].values() for name in group]
        self.assertEqual([n for n in not_used if n not in remote], [], "not_used names a real remote label")


WORKFLOWS = ROOT / ".github/workflows"
PR_TEMPLATE = ROOT / ".github/pull_request_template.md"
PINNED_USES = re.compile(r"^[\w.-]+/[\w./-]+@[0-9a-f]{40}$")
# The only write scope a workflow may hold: rc-label-sync's manual apply job creates labels.
ALLOWED_WRITES = {("rc-label-sync.yml", "apply"): {"issues"}}


class WorkflowTest(unittest.TestCase):
    def workflows(self):
        paths = sorted(WORKFLOWS.glob("*.yml"))
        self.assertEqual([p.name for p in paths], ["ci.yml", "rc-label-sync.yml"])
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

    @staticmethod
    def rc_ci_matrix(check):
        """The two `include` lists of the ci matrix expression (#2554): the one a
        `pull_request` run gets, and the one every other event gets."""
        expr = check["strategy"]["matrix"]["include"]
        m = re.fullmatch(
            r"\$\{\{ fromJSON\(github\.event_name == 'pull_request'\s+&& '(\[.*?\])'\s+\|\| '(\[.*?\])'\) \}\}",
            expr.strip(), re.S)
        if m is None:
            raise AssertionError(f"unexpected ci matrix expression: {expr!r}")
        return json.loads(m.group(1)), json.loads(m.group(2))

    def test_rc_ci_moves_the_full_matrix_after_merge_without_cancelling_it(self):
        """#2554 (user decision U10): PRs run linux-x86_64; the merging push and
        workflow_dispatch run all four targets and windows-build. Only PR runs cancel an
        older run of the same PR; a post-merge run is never cancelled."""
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        triggers = flow.get("on", flow.get(True))
        self.assertEqual(triggers["push"]["branches"], ["v2026.11.0"])
        self.assertIn("workflow_dispatch", triggers)
        self.assertEqual(flow["concurrency"]["cancel-in-progress"],
                         "${{ github.event_name == 'pull_request' }}")
        self.assertIn("github.sha", flow["concurrency"]["group"])
        self.assertEqual(flow["jobs"]["windows-build"]["if"],
                         "${{ github.event_name != 'pull_request' }}")
        self.assertNotIn("if", flow["jobs"]["check"])

    def test_rc_ci_dependency_cache_is_written_only_by_post_merge_runs(self):
        """#2554 AC-3: a PR run restores the base branch's dependency cache and never saves
        one; the cache is pinned and sits before the check set."""
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        steps = flow["jobs"]["check"]["steps"]
        names = [s.get("name") for s in steps]
        cache = steps[names.index("Dependency cache")]
        self.assertRegex(cache["uses"], r"^Swatinem/rust-cache@[0-9a-f]{40}$")
        self.assertEqual(cache["with"]["save-if"], "${{ github.event_name == 'push' }}")
        self.assertLess(names.index("Dependency cache"), names.index("Check set"))
        script = steps[names.index("Check set")]["run"]
        self.assertIn("cargo fetch --locked", script)
        self.assertNotIn("dependencies: no cache", script)

    def test_rc_ci_runs_tests_with_pinned_nextest_and_doc_tests_separately(self):
        """#2554 2-1: the test binaries run in parallel under cargo-nextest, installed from a
        pinned release and checked against a sha256 recorded in the workflow (no fallback);
        doc-tests, which nextest does not run, run with `cargo test --doc`."""
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        steps = flow["jobs"]["check"]["steps"]
        names = [s.get("name") for s in steps]
        install = steps[names.index("cargo-nextest (pinned, sha256-checked)")]["run"]
        self.assertIn("version=0.9.151", install)
        self.assertIn("releases/download/cargo-nextest-${version}/", install)
        _on_pr, after_merge = self.rc_ci_matrix(flow["jobs"]["check"])
        for target in {e["target"] for e in after_merge}:
            self.assertRegex(install, rf"{target}[|)].*sum=[0-9a-f]{{64}}")
        self.assertIn('if [ "$actual" != "$sum" ]; then', install)
        self.assertNotRegex(install, r"\|\|\s*true|cargo install")
        self.assertLess(names.index("cargo-nextest (pinned, sha256-checked)"), names.index("Check set"))
        script = steps[names.index("Check set")]["run"]
        # Both must be real `run` steps of the check set (a recorded exit), not just text.
        self.assertRegex(script, r"(?m)^\s*run test bash -c 'RUST_MIN_STACK=8388608 "
                                 r"cargo nextest run --offline --locked --no-fail-fast ")
        self.assertRegex(script, r"(?m)^\s*run doctest bash -c 'RUST_MIN_STACK=8388608 "
                                 r"cargo test --offline --locked --doc ")
        self.assertNotRegex(script, r"cargo test --offline --locked 2>&1")

    def test_rc_ci_reports_the_rc_ci_check_on_prs_to_the_integration_branch(self):
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        triggers = flow.get("on", flow.get(True))  # PyYAML reads a bare `on:` key as True
        self.assertEqual(triggers["pull_request"]["branches"], ["v2026.11.0"])
        self.assertEqual(flow["jobs"]["ci"]["name"], "ci")
        # The check set runs on every matrix target (#2194); `ci` aggregates them.
        script = "\n".join(s.get("run", "") for s in flow["jobs"]["check"]["steps"])
        for command in ("cargo test --offline --locked", "cargo clippy --offline --locked --all-targets -- -D warnings",
                        "cargo fmt --check", "unittest discover -v -s tests -p '*_test.py'", "git diff --check",
                        "GITHUB_STEP_SUMMARY"):
            self.assertIn(command, script)

    def test_rc_ci_runs_the_four_decided_targets_and_fails_when_any_fails(self):
        """#2194 (decision 01:1x(11-09)): Linux x86_64/arm64 and macOS arm64/x86_64 on standard
        free GitHub-hosted runner labels; a failing target fails `ci` instead of skipping it."""
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        check = flow["jobs"]["check"]
        self.assertIs(check["strategy"]["fail-fast"], False)
        on_pr, after_merge = self.rc_ci_matrix(check)
        # #2554 (user decision U10): the timing moves, the content does not — the targets
        # a PR leaves out all run after merge, with the same runner labels.
        self.assertEqual({e["target"]: e["runner"] for e in after_merge}, {
            "linux-x86_64": "ubuntu-24.04",
            "linux-arm64": "ubuntu-24.04-arm",
            "macos-arm64": "macos-15",
            "macos-x86_64": "macos-15-intel",
        })
        self.assertEqual(on_pr, [e for e in after_merge if e["target"] == "linux-x86_64"])
        self.assertEqual(check["runs-on"], "${{ matrix.runner }}")
        # #2554: every target's cap leaves room over its measured run time (linux-arm64:
        # median 28.1 min, max 29.1 min over 20 runs), so a slow day is not a timeout failure.
        caps = {e["target"]: e["timeout"] for e in after_merge}
        self.assertGreaterEqual(caps["linux-arm64"], 45)
        self.assertEqual(check["timeout-minutes"], "${{ matrix.timeout }}")
        aggregate = flow["jobs"]["ci"]
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
