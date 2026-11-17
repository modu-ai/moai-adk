"""Local structural checks for the RC1 GitHub assets (Issue forms and label taxonomy).

These checks run offline against repository files. They do not prove that GitHub renders the
forms (forms are read from the default branch only) or that the live remote labels match; the label
manifest is compared offline against the snapshot that 00-B (#1986) recorded, and a live
comparison is a separate, explicitly authorized `gh` step recorded in the Issue report.
"""
import importlib.util
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
        # remote yet: `moai-label-sync` apply only creates missing ones and never edits, so a
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


    def test_label_scheme_decisions_44_to_49(self):
        # #2434 decisions 44-49 (6842518327, all A; design 6791931666 Q1-Q5, Q7).
        data = yaml.safe_load(LABELS.read_text(encoding="utf-8"))
        added = {i["name"]: i for i in data["added"]}
        # 44 (Q1): product components get their own area.
        for name in ("area:launcher", "area:statusline"):
            self.assertIn(name, added)
            self.assertEqual(added[name]["color"], "c2e0c6")
        # 46 (Q3): area:update names the Go migration too.
        self.assertIn("Go", added["area:update"]["description"])
        # 48 (Q5): every recorded remote label is listed somewhere; GitHub's own three are reused.
        self.assertTrue({"good first issue", "help wanted", "question"} <= set(data["reused"]["functional"]))
        listed = {n for g in data["reused"].values() for n in g} | set(added)
        listed |= {n for g in data["not_used"].values() for n in g}
        remote = {item["name"] for item in json.loads(REMOTE_LABELS.read_text(encoding="utf-8"))}
        self.assertEqual(sorted(remote - listed), [], "remote labels the file does not list")
        # 45, 46, 47 (Q2-Q4): the rules table carries the new definitions.
        readme = (ROOT / ".github/moai/README.md").read_text(encoding="utf-8")
        for text in (
            "`P1` (must ship in this RC)",
            "`P2` (this RC's scope, may move to the next rc)",
            "`area:workflow` | lane flow:",
            "`area:launcher` |",
            "`area:statusline` |",
        ):
            self.assertIn(text, readme)
        self.assertNotIn("(current queue)", readme)
        self.assertNotIn("Remote labels in neither list", readme)


WORKFLOWS = ROOT / ".github/workflows"
PR_TEMPLATE = ROOT / ".github/pull_request_template.md"
LOCAL_CHECK = ROOT / "tools/local_check.sh"
PINNED_USES = re.compile(r"^[\w.-]+/[\w./-]+@[0-9a-f]{40}$")
# The only write scope a workflow may hold: moai-label-sync's manual apply job creates labels.
ALLOWED_WRITES = {("moai-label-sync.yml", "apply"): {"issues"}}


class WorkflowTest(unittest.TestCase):
    def workflows(self):
        paths = sorted(WORKFLOWS.glob("*.yml"))
        # #2554 PR-A adds the candidate verification path: candidate.yml (the three supported
        # targets and the verdict) and candidate-target.yml (one target, called three times).
        self.assertEqual([p.name for p in paths],
                         ["candidate-target.yml", "candidate.yml", "ci.yml", "moai-label-sync.yml"])
        return [(p, yaml.safe_load(p.read_text(encoding="utf-8"))) for p in paths]

    def test_permissions_are_least_privilege(self):
        for path, flow in self.workflows():
            with self.subTest(workflow=path.name):
                self.assertEqual(flow["permissions"], {"contents": "read"})
                for job_id, job in flow["jobs"].items():
                    self.assertIn("permissions", job, "every job states its own permissions")
                    writes = {k for k, v in job["permissions"].items() if v == "write"}
                    self.assertEqual(writes, ALLOWED_WRITES.get((path.name, job_id), set()), job_id)

    def test_ci_builds_carry_the_run_number_on_the_first_attempt_only(self):
        """SPEC-20261117-headver-ci-number: a ci build stamps HeadVer's build number from the
        run number, and only on the run's first attempt (a re-run keeps `run_number`, so it
        builds unnumbered instead of repeating a number, internal/versioning)."""
        flows = dict((p.name, f) for p, f in self.workflows())
        for job_id in ("check", "windows-build"):
            with self.subTest(job=job_id):
                env = flows["ci.yml"]["jobs"][job_id]["env"]
                self.assertEqual(env["MOAI_BUILD_NUMBER"],
                                 "${{ github.run_attempt == '1' && github.run_number || '' }}")
                self.assertNotIn("MOAI_LOCAL_BUILD", env)

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
        workflow_dispatch run linux-x86_64, macos-arm64 and windows-build
        (SPEC-20261117-supported-platforms). Only PR runs cancel an
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

    def test_rc_ci_post_merge_macos_waits_are_replaced_by_a_newer_merge(self):
        """#2554 D8 (a): on a push, only the macOS targets share a group across runs, so a
        newer merge replaces a macOS job that is still waiting; a running job is never
        cancelled, and every other target and event keeps a group of its own. Only a run's
        first attempt shares the group: a re-run of an older SHA must not replace the
        current tip's waiting job, which the tip watch would read as red (#2739 audit F1)."""
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        conc = flow["jobs"]["check"]["concurrency"]
        self.assertIs(conc["cancel-in-progress"], False)
        group = " ".join(conc["group"].split())
        self.assertEqual(group,
                         "${{ github.event_name == 'push' && github.run_attempt == '1' "
                         "&& startsWith(matrix.target, 'macos-') "
                         "&& format('ci-push-{0}', matrix.target) "
                         "|| format('ci-run-{0}-{1}-{2}', github.run_id, github.run_attempt, "
                         "matrix.target) }}")
        self.assertIn("github.run_attempt == '1'", group.split("||")[0])
        _on_pr, after_merge = self.rc_ci_matrix(flow["jobs"]["check"])
        shared = sorted(e["target"] for e in after_merge if e["target"].startswith("macos-"))
        self.assertEqual(shared, ["macos-arm64"])
        # The aggregate and windows-build keep the run-level group only.
        for job in ("ci", "windows-build"):
            self.assertNotIn("concurrency", flow["jobs"][job])

    def test_every_scope_step_has_the_commit_it_diffs_against(self):
        """#2554 (decision 6861024531): a job whose docs-only scope step diffs against
        `HEAD~1` checks out at least two commits (or the full history, 0). With the default
        depth of 1 the diff failed and windows-build ran full on every push."""
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        seen = []
        for name, job in flow["jobs"].items():
            steps = job.get("steps", [])
            names = [s.get("name") or "" for s in steps]
            scope = [i for i, n in enumerate(names) if n.startswith("Scope (docs-only")]
            if not scope:
                continue
            seen.append(name)
            self.assertIn("HEAD~1", steps[scope[0]]["run"])
            checkout = [i for i, s in enumerate(steps)
                        if str(s.get("uses", "")).startswith("actions/checkout@")]
            with self.subTest(job=name):
                self.assertTrue(checkout and checkout[0] < scope[0])
                depth = steps[checkout[0]].get("with", {}).get("fetch-depth", 1)
                self.assertTrue(depth == 0 or depth >= 2, f"{name} fetch-depth {depth}")
        self.assertEqual(sorted(seen), ["check", "windows-build"])

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
        self.assertIn("bash tools/local_check.sh", steps[names.index("Check set")]["run"])
        # #2906 S5a: the Check set runs tools/local_check.sh. Both must be real steps of it
        # (a recorded exit), not just text.
        script = LOCAL_CHECK.read_text(encoding="utf-8")
        self.assertRegex(script, r"(?m)^\s*step heavy test bash -c 'RUST_MIN_STACK=8388608 "
                                 r"cargo nextest run --offline --locked --no-fail-fast ")
        self.assertRegex(script, r"(?m)^\s*step heavy doctest bash -c 'RUST_MIN_STACK=8388608 "
                                 r"cargo test --offline --locked --doc ")
        self.assertNotRegex(script, r"cargo test --offline --locked 2>&1")

    def test_rc_ci_reports_the_rc_ci_check_on_prs_to_the_integration_branch(self):
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        triggers = flow.get("on", flow.get(True))  # PyYAML reads a bare `on:` key as True
        self.assertEqual(triggers["pull_request"]["branches"], ["v2026.11.0"])
        self.assertEqual(flow["jobs"]["ci"]["name"], "ci")
        # The check set runs on every matrix target (#2194); `ci` aggregates them.
        script = "\n".join(s.get("run", "") for s in flow["jobs"]["check"]["steps"])
        self.assertIn("bash tools/local_check.sh", script)
        script += "\n" + LOCAL_CHECK.read_text(encoding="utf-8")  # #2906 S5a
        for command in ("cargo test --offline --locked", "cargo clippy --offline --locked --all-targets -- -D warnings",
                        "cargo fmt --check", "unittest discover -v -s tests -p '*_test.py'", "git diff --check",
                        "GITHUB_STEP_SUMMARY"):
            self.assertIn(command, script)

    def test_rc_ci_runs_the_supported_targets_and_fails_when_any_fails(self):
        """SPEC-20261117-supported-platforms (user decision 2026-11-17, replacing the four
        targets of decision 01:1x(11-09)): Linux x86_64 and macOS arm64 on standard free
        GitHub-hosted runner labels (Windows x86_64 is the windows-build job); a failing target
        fails `ci` instead of skipping it."""
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        check = flow["jobs"]["check"]
        self.assertIs(check["strategy"]["fail-fast"], False)
        on_pr, after_merge = self.rc_ci_matrix(check)
        # #2554 (user decision U10): the timing moves, the content does not — the targets
        # a PR leaves out all run after merge, with the same runner labels.
        self.assertEqual({e["target"]: e["runner"] for e in after_merge}, {
            "linux-x86_64": "ubuntu-24.04",
            "macos-arm64": "macos-15",
        })
        self.assertEqual(on_pr, [e for e in after_merge if e["target"] == "linux-x86_64"])
        self.assertEqual(check["runs-on"], "${{ matrix.runner }}")
        # Every target's cap leaves room over its measured run time, so a slow day is not a
        # timeout failure.
        caps = {e["target"]: e["timeout"] for e in after_merge}
        self.assertEqual(caps, {"linux-x86_64": 30, "macos-arm64": 45})
        self.assertEqual(check["timeout-minutes"], "${{ matrix.timeout }}")
        aggregate = flow["jobs"]["ci"]
        self.assertEqual(aggregate["needs"], ["check", "ci-verdict"])
        # #2312: runs after a failed target (so the check fails, not skips) but not after a
        # cancellation (always() reported a cancel-in-progress run as a failure). C-83: it is
        # skipped only when ci-verdict wrote `superseded`.
        self.assertEqual(aggregate.get("if"),
                         "${{ !cancelled() && needs.ci-verdict.outputs.state != 'superseded' }}")
        gate = "\n".join(s.get("run", "") for s in aggregate["steps"])
        self.assertIn('[ "${{ needs.check.result }}" = "success" ]', gate)

    def test_rc_ci_reads_a_superseded_push_run_as_neutral_only_through_the_verdict(self):
        """C-83 (SPEC-20261117-ci-superseded-neutral): ci-verdict runs only on a push whose
        matrix result is `cancelled`, reads the run's jobs and their annotations with read-only
        permissions, and hands the decision to tools/ci_superseded.py; a PR never runs it."""
        flow = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        verdict = flow["jobs"]["ci-verdict"]
        self.assertEqual(verdict["needs"], ["check"])
        self.assertEqual(verdict["if"], "${{ !cancelled() && github.event_name == 'push' "
                                        "&& needs.check.result == 'cancelled' }}")
        self.assertEqual(verdict["permissions"],
                         {"actions": "read", "checks": "read", "contents": "read"})
        self.assertEqual(verdict["outputs"], {"state": "${{ steps.verdict.outputs.state }}"})
        step = next(s for s in verdict["steps"] if s.get("id") == "verdict")
        self.assertIn("python3 -I tools/ci_superseded.py", step["run"])
        self.assertIn('exit "${PIPESTATUS[0]}"', step["run"])
        self.assertEqual(step["env"]["CHECK_RESULT"], "${{ needs.check.result }}")
        self.assertEqual(step["env"]["EVENT"], "${{ github.event_name }}")
        self.assertNotIn("concurrency", verdict)

    def test_label_sync_is_a_dry_run_unless_applied_by_hand(self):
        flow = dict(self.workflows())[WORKFLOWS / "moai-label-sync.yml"]
        triggers = flow.get("on", flow.get(True))
        self.assertIs(triggers["workflow_dispatch"]["inputs"]["apply"]["default"], False)
        self.assertIn("workflow_dispatch", flow["jobs"]["apply"]["if"])
        text = (WORKFLOWS / "moai-label-sync.yml").read_text(encoding="utf-8")
        self.assertIn(".github/moai/labels.yml", text)
        self.assertIsNone(re.search(r"gh label (delete|edit)|--force", text), "no label edits or deletes")

    # --- #2554 PR-A: candidate verification (user decision #1975 6907950823) -----------------

    # SPEC-20261117-supported-platforms: exactly the three supported platforms.
    CANDIDATE_TARGETS = {
        "linux-x86_64": "ubuntu-24.04",
        "macos-arm64": "macos-15",
        "windows-x86_64": "windows-2025",
    }

    def test_no_workflow_names_a_dropped_platform(self):
        """SPEC-20261117-supported-platforms: Intel macOS (`macos-15-intel`, `macos-x86_64`,
        `x86_64-apple-darwin`) and Linux arm64 (`ubuntu-24.04-arm`, `linux-arm64`,
        `aarch64-unknown-linux-gnu`) appear in no workflow, and each supported target does."""
        texts = {p.name: p.read_text(encoding="utf-8") for p, _ in self.workflows()}
        for name, text in texts.items():
            for gone in ("macos-15-intel", "macos-x86_64", "x86_64-apple-darwin",
                         "ubuntu-24.04-arm", "linux-arm64", "aarch64-unknown-linux-gnu"):
                with self.subTest(workflow=name, gone=gone):
                    self.assertNotIn(gone, text)
        both = texts["ci.yml"] + texts["candidate.yml"]
        for kept in ("linux-x86_64", "macos-arm64", "windows-2025"):
            self.assertIn(kept, both)

    def candidate_flows(self):
        flows = dict(self.workflows())
        return flows[WORKFLOWS / "candidate.yml"], flows[WORKFLOWS / "candidate-target.yml"]

    @staticmethod
    def steps_of(flow):
        return [step for job in flow["jobs"].values() for step in job.get("steps", [])]

    def test_candidate_verification_starts_only_from_a_push_to_a_candidate_branch(self):
        """(a) the run is pinned: the branch is candidate/<40 hex> and names the pushed commit.
        (e) a newer run never cancels an older one: there is no `concurrency` anywhere."""
        caller, target = self.candidate_flows()
        self.assertEqual(caller.get("on", caller.get(True)), {"push": {"branches": ["candidate/**"]}})
        self.assertEqual(target.get("on", target.get(True)).keys(), {"workflow_call"})
        for flow in (caller, target):
            self.assertNotIn("concurrency", flow)
            for job in flow["jobs"].values():
                self.assertNotIn("concurrency", job)
        pin = caller["jobs"]["pin"]
        (step,) = pin["steps"]
        self.assertEqual(step["env"], {"REF_NAME": "${{ github.ref_name }}", "PUSHED": "${{ github.sha }}"})
        self.assertIn("^candidate/[0-9a-f]{40}$", step["run"])
        self.assertIn('[ "$named" != "$PUSHED" ]', step["run"])
        self.assertEqual(pin["outputs"], {"sha": "${{ steps.pin.outputs.sha }}"})

    def test_candidate_values_reach_scripts_through_env_only(self):
        """A branch name or an input is outside text: no `${{ }}` inside a `run:` script."""
        for flow in self.candidate_flows():
            for step in self.steps_of(flow):
                with self.subTest(step=step.get("name", step.get("uses"))):
                    self.assertNotIn("${{", step.get("run", ""))

    def test_candidate_runs_the_three_targets_and_one_verdict_needs_all_of_them(self):
        """(a)(b)(c): the three supported targets on one commit, Windows included, one summary
        that fails when any target is missing or failed."""
        caller, _ = self.candidate_flows()
        jobs = caller["jobs"]
        self.assertEqual(set(jobs), {"pin", "candidate", *self.CANDIDATE_TARGETS})
        for key, runner in self.CANDIDATE_TARGETS.items():
            with self.subTest(target=key):
                job = jobs[key]
                self.assertEqual(job["uses"], "./.github/workflows/candidate-target.yml")
                self.assertEqual(job["needs"], ["pin"])
                self.assertEqual(job["name"], f"candidate / {key}")
                with_ = job["with"]
                self.assertEqual(with_["sha"], "${{ needs.pin.outputs.sha }}")
                self.assertEqual((with_["target"], with_["runner"]), (key, runner))
                self.assertRegex(with_["nextest_sha256"], r"^[0-9a-f]{64}$")
        self.assertEqual(jobs["windows-x86_64"]["with"]["nextest_asset"], "x86_64-pc-windows-msvc")
        # The Unix targets are the same runners and the same nextest pins as ci.yml.
        ci = dict(self.workflows())[WORKFLOWS / "ci.yml"]
        _, after_merge = self.rc_ci_matrix(ci["jobs"]["check"])
        self.assertEqual({e["target"]: e["runner"] for e in after_merge},
                         {k: v for k, v in self.CANDIDATE_TARGETS.items() if not k.startswith("windows")})
        ci_text = (WORKFLOWS / "ci.yml").read_text(encoding="utf-8")
        for key in self.CANDIDATE_TARGETS:
            if not key.startswith("windows"):
                self.assertIn(jobs[key]["with"]["nextest_sha256"], ci_text, key)
        summary = jobs["candidate"]
        self.assertEqual(summary["name"], "candidate")
        self.assertEqual(summary["if"], "${{ !cancelled() }}")
        self.assertEqual(set(summary["needs"]), {"pin", *self.CANDIDATE_TARGETS})
        (checkout, verdict) = summary["steps"]
        self.assertEqual(verdict["env"], {"NEEDS": "${{ toJSON(needs) }}"})
        self.assertIn("python3 tools/candidate_summary.py", verdict["run"])
        self.assertIn('exit "${PIPESTATUS[0]}"', verdict["run"])

    def test_candidate_check_names_are_the_ones_the_tip_gate_ignores(self):
        """src/flow.rs is_candidate_check: `candidate` or `candidate / ...`. A check run named
        otherwise on the tip would be read as a failing job of the tip (#2554 PR-A guard)."""
        caller, target = self.candidate_flows()
        for job_id, job in caller["jobs"].items():
            with self.subTest(job=job_id):
                self.assertTrue(job["name"] == "candidate" or job["name"].startswith("candidate / "))
        # The reusable workflow's own job is reported as `<caller job name> / check`.
        self.assertEqual(target["jobs"]["check"]["name"], "check")

    def test_candidate_summary_script_judges_the_same_three_targets(self):
        spec = importlib.util.spec_from_file_location("candidate_summary", ROOT / "tools/candidate_summary.py")
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        self.assertEqual(set(module.TARGETS), set(self.CANDIDATE_TARGETS))

    def test_candidate_target_checks_out_the_pinned_commit_and_counts_the_tests(self):
        """(a)(b): the target checks out the pinned commit and says so; no tests run means no
        pass, on Windows as everywhere; a skipped test is counted, never silent."""
        _, target = self.candidate_flows()
        job = target["jobs"]["check"]
        steps = job["steps"]
        checkout = steps[0]
        self.assertRegex(checkout["uses"], r"^actions/checkout@[0-9a-f]{40}$")
        self.assertEqual(checkout["with"], {"ref": "${{ inputs.sha }}", "persist-credentials": False,
                                            "fetch-depth": 0})
        by_name = {s.get("name"): s for s in steps}
        self.assertEqual(by_name["Checked out the pinned commit"]["env"], {"PINNED": "${{ inputs.sha }}"})
        self.assertIn('[ "$head" != "$PINNED" ]', by_name["Checked out the pinned commit"]["run"])
        self.assertIs(by_name["Dependency cache"]["with"]["save-if"], False)
        check_set = by_name["Check set"]["run"]
        self.assertIn("LOCAL_CHECK_MODE=in-place", check_set)
        self.assertIn("bash tools/local_check.sh", check_set)
        # No skip list in PR-A: a test that cannot run on a target fails by name.
        for flow_text in ((WORKFLOWS / "candidate-target.yml").read_text(encoding="utf-8"),):
            self.assertIsNone(re.search(r"--skip|--filter-expr|nextest run[^\n]* -E ", flow_text))
        self.assertEqual(by_name["Record result"]["if"], "${{ always() }}")
        self.assertEqual(by_name["Tests ran"]["if"], "${{ always() }}")
        self.assertEqual(steps[-1]["name"], "Tests ran")
        self.assertIn("compiling is not a pass", by_name["Tests ran"]["run"])
        names = ("sha", "tests_run", "tests_passed", "tests_skipped")
        self.assertEqual(set(job["outputs"]), set(names))
        self.assertEqual(set(target.get("on", target.get(True))["workflow_call"]["outputs"]), set(names))

    def test_candidate_windows_tools_are_checksummed_before_they_are_unpacked(self):
        _, target = self.candidate_flows()
        ripgrep = {s.get("name"): s for s in target["jobs"]["check"]["steps"]}["Install and verify ripgrep"]["run"]
        self.assertIn("d0f534024c42afd6cb4d38907c25cd2b249b79bbe6cc1dbee8e3e37c2b6e25a1", ripgrep)
        self.assertLess(ripgrep.index('"$actual" != "$sum"'), ripgrep.index("tar.exe -xf"))
        nextest = {s.get("name"): s for s in target["jobs"]["check"]["steps"]}["cargo-nextest (pinned, sha256-checked)"]["run"]
        self.assertLess(nextest.index('"$actual" != "$SUM"'), nextest.index("tar xzf"))

    def test_candidate_path_has_no_release_publish_secret_or_write_scope(self):
        """(d): read-only token, no secret, no release, publish, tag or artifact upload."""
        forbidden = re.compile(
            r"gh release|git tag|git push|upload-artifact|download-artifact|attest|cargo publish|"
            r"npm publish|docker push|gh api|gh pr|gh issue", re.I)
        for name in ("candidate.yml", "candidate-target.yml"):
            text = (WORKFLOWS / name).read_text(encoding="utf-8")
            self.assertIsNone(re.search(r"\$\{\{\s*secrets\.|secrets:\s*inherit|GITHUB_TOKEN|id-token", text), name)
        for flow in self.candidate_flows():
            self.assertEqual(flow["permissions"], {"contents": "read"})
            for job in flow["jobs"].values():
                self.assertNotIn("secrets", job)
            for step in self.steps_of(flow):
                with self.subTest(step=step.get("name", step.get("uses"))):
                    self.assertIsNone(forbidden.search(step.get("run", "") + step.get("uses", "")))


class PullRequestTemplateTest(unittest.TestCase):
    def test_template_carries_the_rc1_pr_contract(self):
        text = PR_TEMPLATE.read_text(encoding="utf-8")
        for field in ("Refs #", "Moai-Claim:", "exact", "40자 SHA", "실행한 검사와 exit", "독립 review",
                      "not_run", "blocked", "--body-file"):
            self.assertIn(field, text)
        self.assertIsNone(re.search(r"(?im)^\s*(closes|fixes|resolves) #", text), "no auto-close keyword")
        self.assertIsNone(SECRET_SHAPES.search(text))

    def test_template_names_the_rust_checker(self):
        # #2708: the Python checker is gone; the form points at `moai pr-self-audit-check`
        # for both the first audit's ci: line and the re-audit body check (audit 6864465911 N1)
        text = PR_TEMPLATE.read_text(encoding="utf-8")
        self.assertNotIn("tools/check_pr_self_audit.py", text)
        self.assertIn("`moai pr-self-audit-check`", text)
        self.assertIn("`moai pr-self-audit-check --reaudit --file BODY.md`", text)


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
