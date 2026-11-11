"""Offline checks for the RC1 00-B GitHub inventory collector (#1986).

Every test runs against recorded, redacted fixtures under tests/fixtures/github or against
synthetic inputs; nothing here calls `gh` or the network.
"""
import importlib.util
import json
import pathlib
import re
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parent.parent
FIXTURES = ROOT / "tests/fixtures/github"
MANIFEST = ROOT / ".moai-cli/manifests/github-issues.json"
FORMS_SHA = FIXTURES / "forms-primary.sha256"
COLLECTOR_PATH = ROOT / "tests/github_inventory_collect.py"

_spec = importlib.util.spec_from_file_location("github_inventory_collect", COLLECTOR_PATH)
collect = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(collect)


def build_into(tmp):
    out = pathlib.Path(tmp)
    collect.build(FIXTURES, out, MANIFEST, ROOT, FORMS_SHA)
    inventory = json.loads((out / "github-settings-inventory.json").read_text(encoding="utf-8"))
    dup = json.loads((out / "github-duplicate-map.json").read_text(encoding="utf-8"))
    return out, inventory, dup


class GetOnlyTest(unittest.TestCase):
    def test_argv_is_always_an_explicit_get(self):
        for paginate in (True, False):
            argv = collect.gh_argv("repos/o/r/labels?per_page=100", paginate=paginate)
            self.assertEqual(argv[:4], ["gh", "api", "-X", "GET"])
            collect.check_get_only(argv)

    def test_body_or_method_arguments_are_refused(self):
        for extra in (["-f", "a=b"], ["-F", "a=1"], ["--input", "x.json"], ["--method", "GET"], ["-X", "GET"]):
            with self.subTest(extra=extra):
                with self.assertRaises(collect.WriteRefused):
                    collect.check_get_only(["gh", "api", "-X", "GET", "repos/o/r"] + extra)
        with self.assertRaises(collect.WriteRefused):
            collect.check_get_only(["gh", "api", "repos/o/r"])

    def test_recorded_request_log_is_get_only(self):
        meta = json.loads((FIXTURES / "meta.json").read_text(encoding="utf-8"))
        self.assertTrue(meta["requests"])
        self.assertEqual({r["method"] for r in meta["requests"]}, {"GET"})

    def test_source_names_no_writing_method_flag(self):
        source = COLLECTOR_PATH.read_text(encoding="utf-8")
        self.assertIsNone(re.search(r"-X (POST|PATCH|PUT|DELETE)|--method (POST|PATCH|PUT|DELETE)", source))


class SplitAndRedactTest(unittest.TestCase):
    def test_pull_request_key_separates_prs(self):
        raw = [{"number": 1}, {"number": 2, "pull_request": {"merged_at": None}}, {"number": 3}]
        issues, prs = collect.split_issues_and_prs(raw)
        self.assertEqual([i["number"] for i in issues], [1, 3])
        self.assertEqual([p["number"] for p in prs], [2])

    def test_issue_redaction_keeps_only_the_whitelist(self):
        raw = {"number": 7, "id": 70, "node_id": "I_x", "title": "t", "state": "open", "state_reason": None,
               "labels": [{"name": "b"}, {"name": "a"}], "milestone": {"title": "m", "creator": {"login": "x"}},
               "type": {"name": "Task", "id": 1}, "body": "secret", "user": {"login": "x"},
               "html_url": "https://example.invalid", "assignees": [{"login": "x"}]}
        item = collect.redact_issue(raw)
        self.assertEqual(set(item), set(collect.ISSUE_FIELDS))
        self.assertEqual(item["labels"], ["a", "b"])
        self.assertEqual((item["milestone"], item["type"]), ("m", "Task"))

    def test_fixtures_carry_no_secret_or_personal_shapes(self):
        shapes = re.compile(r"ghp_|github_pat_|-----BEGIN|/Users/|@[A-Za-z0-9.-]+\.[a-z]{2,}")
        for path in sorted(FIXTURES.glob("*")):
            with self.subTest(fixture=path.name):
                self.assertIsNone(shapes.search(path.read_text(encoding="utf-8")))
        for name in ("issues-open.json", "issues-closed.json"):
            for item in json.loads((FIXTURES / name).read_text(encoding="utf-8")):
                self.assertEqual(set(item), set(collect.ISSUE_FIELDS))
        for item in json.loads((FIXTURES / "prs.json").read_text(encoding="utf-8")):
            self.assertEqual(set(item), set(collect.PR_FIELDS))


class LabelClassificationTest(unittest.TestCase):
    def test_axes_and_unknown(self):
        self.assertEqual(collect.classify_label("type:fix")[0], "type")
        self.assertEqual(collect.classify_label("bug"), ("legacy-synonym", "type:fix", "issue-body"))
        self.assertEqual(collect.classify_label("critical")[1], "priority:P0")
        self.assertEqual(collect.classify_label("triage")[0], "status")
        self.assertEqual(collect.classify_label("host:codex")[0], "host")
        self.assertEqual(collect.classify_label("question")[0], "functional")
        self.assertEqual(collect.classify_label("something-new"), ("unknown", None, None))

    def test_inventory_counts_every_remote_label(self):
        with tempfile.TemporaryDirectory() as tmp:
            _, inventory, _ = build_into(tmp)
        labels = json.loads((FIXTURES / "labels.json").read_text(encoding="utf-8"))
        self.assertEqual(len(inventory["labels"]), len(labels))
        self.assertEqual(sum(inventory["label_axis_counts"].values()), len(labels))
        self.assertIn("unknown", inventory["label_axis_counts"])


class InventoryTest(unittest.TestCase):
    def test_build_is_deterministic(self):
        with tempfile.TemporaryDirectory() as a, tempfile.TemporaryDirectory() as b:
            out_a, _, _ = build_into(a)
            out_b, _, _ = build_into(b)
            for name in ("github-settings-inventory.json", "github-duplicate-map.json", "github-forms.sha256"):
                self.assertEqual((out_a / name).read_bytes(), (out_b / name).read_bytes(), name)

    def test_schema_fields_and_projects(self):
        with tempfile.TemporaryDirectory() as tmp:
            _, inventory, _ = build_into(tmp)
        self.assertEqual(inventory["schema_version"], 1)
        for key in ("fetched_at", "gh_version", "api_version", "repository"):
            self.assertTrue(inventory[key])
        self.assertNotIn("auth_scopes", inventory)
        if not inventory["project_scope"]:
            self.assertEqual(inventory["projects"], collect.PROJECTS_UNAVAILABLE)

    def test_manifest_reconciliation_is_reported(self):
        with tempfile.TemporaryDirectory() as tmp:
            _, inventory, _ = build_into(tmp)
        for part in ("sub_issues", "dependencies"):
            rec = inventory["manifest_reconciliation"][part]
            self.assertEqual(rec["remote"], rec["manifest_linked"] - len(rec["missing_remote"]) + len(rec["extra_remote"]))

    def test_projects_unavailable_without_scope(self):
        fx = collect.load_fixtures(FIXTURES)
        fx["meta"] = dict(fx["meta"], project_scope=False)
        manifest = json.loads(MANIFEST.read_text(encoding="utf-8"))
        inventory = collect.settings_inventory(fx, manifest, ROOT, collect.read_shasum(FORMS_SHA))
        self.assertEqual(inventory["projects"], "unavailable(read:project scope missing)")


class FormsTest(unittest.TestCase):
    def test_eight_form_paths_are_hashed(self):
        with tempfile.TemporaryDirectory() as tmp:
            out, inventory, _ = build_into(tmp)
            lines = (out / "github-forms.sha256").read_text(encoding="utf-8").splitlines()
        self.assertEqual(len(inventory["forms"]), 8)
        self.assertTrue(all(re.fullmatch(r"[0-9a-f]{64}", f["sha256"] or "") for f in inventory["forms"]))
        self.assertEqual(len(lines), 8)

    def test_adopted_forms_match_the_assets_test(self):
        assets = (ROOT / "tests/github_assets_test.py").read_text(encoding="utf-8")
        checked = set(re.findall(r'"(\.github/ISSUE_TEMPLATE/[^"]+\.yml)"', assets))
        adopted = {p for p, (_, adoption) in collect.FORM_PATHS.items() if adoption == "adopted"}
        # The 00-B inventory is a frozen 2026-11-08 record. #2434 added these files after it;
        # they are checked by the assets test but are not part of that record.
        added_after_snapshot = {
            ".github/ISSUE_TEMPLATE/moai-followup.yml",
            ".github/ISSUE_TEMPLATE/config.yml",
        }
        self.assertEqual(adopted, checked - added_after_snapshot)
        self.assertTrue(added_after_snapshot <= checked)


class DuplicateMapTest(unittest.TestCase):
    def test_every_open_issue_is_mapped(self):
        with tempfile.TemporaryDirectory() as tmp:
            _, _, dup = build_into(tmp)
        unmapped = [r for r in dup["issues_open"] if r["logical_key"] is None and r["non_rc1"] is not True]
        self.assertEqual(unmapped, [])
        self.assertEqual(dup["counts"]["issues_open"], len(dup["issues_open"]))

    def test_candidates_need_a_reason(self):
        fx = {
            "meta": {"repository": {"full_name": "o/r"}},
            "issues-open": [{"number": 10, "id": 100, "node_id": "I_a", "title": "[RC1][x] 00-B GitHub 설정 대조"}],
            "issues-closed": [
                {"number": 1, "id": 11, "title": "GitHub 설정 대조", "state_reason": "completed"},
                {"number": 2, "id": 12, "title": "old 00-B follow-up", "state_reason": "completed"},
                {"number": 3, "id": 13, "title": "SPEC-RC1-00-B draft", "state_reason": "not_planned"},
                {"number": 4, "id": 14, "title": "unrelated windows installer bug", "state_reason": "completed"},
            ],
        }
        manifest = {"issues": {"00-B": {"number": 10}}}
        dup = collect.duplicate_map(fx, manifest, "2026-01-01T00:00:00Z")
        reasons = {c["number"]: c["reason"] for c in dup["duplicate_candidates"]}
        self.assertEqual(set(reasons), {1, 2, 3})
        self.assertIn("title_similar", reasons[1])
        self.assertIn("logical_key", reasons[2])
        self.assertIn("spec_id", reasons[3])
        self.assertTrue(all(c["status"] == "candidate" for c in dup["duplicate_candidates"]))


if __name__ == "__main__":
    unittest.main()
