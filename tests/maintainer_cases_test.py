"""Tests for tools/maintainer_cases.py (SPEC-20261121-maintainer-eval-cases).

The tool replays the maintainer evaluation cases: four mechanical checks, each run on frozen
evidence from an incident or on the current tree, and the verdict compared with the expected
one. Every unit test builds temporary inputs and goes through the command a maintainer runs,
`python3 -I tools/maintainer_cases.py replay`; ReplayTest and RepoTreeTest run it on this
repository. Written for Python 3.9.
"""

import hashlib
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
TOOL = REPO_ROOT / "tools" / "maintainer_cases.py"
IMPROVE = REPO_ROOT / "tools" / "improve"
REFUSAL_PREFIX = "maintainer_cases: "
CASE_IDS = {"false-repair-claim", "wall-clock-test", "guidance-drift", "unscoped-test-deletion"}


def write(root, rel, text):
    path = Path(root) / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")
    return path


def run(*args):
    proc = subprocess.run(
        [sys.executable, "-I", str(TOOL), *[str(a) for a in args]],
        capture_output=True,
        text=True,
    )
    return proc.returncode, proc.stdout, proc.stderr


def case_file(check, input_rel, expect, case_id="c1", target="t", **extra):
    case = {
        "id": case_id,
        "lesson": "a lesson",
        "report": "a report section",
        "check": check,
        "targets": {target: {"input": input_rel, "expect": expect}},
    }
    case.update(extra)
    return {"version": 1, "cases": [case]}


class Temp(unittest.TestCase):
    def setUp(self):
        self._tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self._tmp.cleanup)
        self.root = Path(self._tmp.name) / "root"
        self.root.mkdir()
        self.cfg = Path(self._tmp.name) / "cfg"
        self.cfg.mkdir()

    def replay_with(self, data):
        path = write(self.cfg, "cases.json", json.dumps(data))
        return run("replay", "--root", self.root, "--cases", path)

    def target(self, check, input_rel, expect="flag"):
        """Replay one synthetic case and return its single target row."""
        code, out, err = self.replay_with(case_file(check, input_rel, expect))
        self.assertEqual(err, "", err)
        report = json.loads(out)
        return code, report["cases"][0]["targets"][0]

    def verdict(self, check, input_rel):
        return self.target(check, input_rel)[1]["verdict"]


class ClaimsTest(Temp):
    def claims(self, claimed, changed):
        write(self.root, "in/claims.txt", "\n".join(claimed) + "\n")
        write(self.root, "in/changed-files.txt", "\n".join(changed) + "\n")
        return self.target("claims_in_diff", "in")[1]

    def test_every_claimed_path_changed_is_clean(self):
        row = self.claims(["a.rs", "b.py"], ["b.py", "a.rs", "c.md"])
        self.assertEqual(row["verdict"], "clean")

    def test_a_claimed_path_missing_from_the_changed_files_flags_and_names_it(self):
        row = self.claims(["a.rs", "src/main.rs"], ["a.rs", "c.md"])
        self.assertEqual(row["verdict"], "flag")
        self.assertIn("src/main.rs", row["detail"])
        self.assertNotIn("a.rs,", row["detail"])

    def test_an_empty_claim_list_is_clean(self):
        self.assertEqual(self.claims([], ["a.rs"])["verdict"], "clean")

    def test_a_path_is_compared_whole_not_by_suffix(self):
        row = self.claims(["main.rs"], ["src/main.rs"])
        self.assertEqual(row["verdict"], "flag")


class ClockTest(Temp):
    def rust(self, text):
        write(self.root, "in.rs", text)
        return self.target("no_clock_in_tests", "in.rs")[1]

    def test_a_test_that_reads_instant_now_flags(self):
        row = self.rust("#[test]\nfn slow() {\n    let t = Instant::now();\n}\n")
        self.assertEqual(row["verdict"], "flag")
        self.assertIn("slow", row["detail"])

    def test_elapsed_and_system_time_flag_too(self):
        for body in ("let _ = start.elapsed();", "let _ = SystemTime::now();"):
            row = self.rust("#[test]\nfn t() {\n    %s\n}\n" % body)
            self.assertEqual(row["verdict"], "flag", body)

    def test_the_same_call_outside_a_test_is_clean(self):
        row = self.rust("fn helper() {\n    let t = Instant::now();\n}\n")
        self.assertEqual(row["verdict"], "clean")

    def test_a_test_without_a_clock_is_clean(self):
        row = self.rust("#[test]\nfn fast() {\n    assert_eq!(1 + 1, 2);\n}\n")
        self.assertEqual(row["verdict"], "clean")

    def test_a_nested_block_does_not_end_the_body_early(self):
        row = self.rust(
            "#[test]\nfn t() {\n    if true {\n        let _ = 1;\n    }\n    let _ = Instant::now();\n}\n"
        )
        self.assertEqual(row["verdict"], "flag")

    def test_a_brace_inside_a_string_does_not_end_the_body_early(self):
        row = self.rust(
            '#[test]\nfn t() {\n    let s = "}";\n    let _ = Instant::now();\n}\n'
        )
        self.assertEqual(row["verdict"], "flag")

    def test_a_clock_named_only_in_a_comment_is_clean(self):
        row = self.rust("#[test]\nfn t() {\n    // Instant::now() would be flaky here\n}\n")
        self.assertEqual(row["verdict"], "clean")

    def test_a_test_attribute_with_arguments_still_counts(self):
        row = self.rust(
            '#[tokio::test(flavor = "multi_thread")]\nasync fn t() {\n    let _ = Instant::now();\n}\n'
        )
        self.assertEqual(row["verdict"], "flag")

    def test_a_semicolon_in_the_return_type_does_not_hide_the_body(self):
        row = self.rust(
            "#[test]\nfn t() -> Result<[u8; 4], ()> {\n    let _ = Instant::now();\n    Ok([0; 4])\n}\n"
        )
        self.assertEqual(row["verdict"], "flag")

    def test_a_raw_byte_string_holding_a_brace_and_a_quote_does_not_end_the_body(self):
        row = self.rust('#[test]\nfn t() {\n    let s = br#"} " {"#;\n    let _ = Instant::now();\n}\n')
        self.assertEqual(row["verdict"], "flag")

    def test_a_raw_string_holding_a_brace_does_not_end_the_body(self):
        row = self.rust('#[test]\nfn t() {\n    let s = r##"a"} "#"##;\n    let _ = Instant::now();\n}\n')
        self.assertEqual(row["verdict"], "flag")

    def test_char_literals_holding_a_brace_or_a_quote_do_not_end_the_body(self):
        row = self.rust(
            "#[test]\nfn t<'a>(x: &'a str) {\n    let (a, b, c) = ('{', '\"', '\\'');\n    let _ = Instant::now();\n}\n"
        )
        self.assertEqual(row["verdict"], "flag")

    def test_a_nested_block_comment_naming_a_clock_is_clean(self):
        row = self.rust("#[test]\nfn t() {\n    /* a /* b */ Instant::now() */\n    let _ = 1;\n}\n")
        self.assertEqual(row["verdict"], "clean")

    def test_the_clock_of_a_later_function_does_not_flag_an_earlier_test(self):
        row = self.rust(
            "#[test]\nfn fast() {\n    assert!(true);\n}\nfn helper() {\n    let _ = Instant::now();\n}\n"
        )
        self.assertEqual(row["verdict"], "clean")


class StageOrderTest(Temp):
    CHAIN = "- Every change has a SPEC, and its order is PLAN → RUN → VERIFY → SYNC → AUDIT → merge (`x`).\n"

    def stages(self, round_line):
        write(self.root, "chain.md", "intro\n" + self.CHAIN + "outro\n")
        write(self.root, "round.md", "a\n" + round_line + "\nb\n")
        write(
            self.root,
            "in.json",
            json.dumps(
                {
                    "chain": {"file": "chain.md", "contains": "PLAN → RUN"},
                    "round": {"file": "round.md", "contains": "One round runs"},
                }
            ),
        )
        return self.target("stage_order", "in.json")[1]

    def test_every_stage_in_order_is_clean(self):
        row = self.stages("One round runs `take → PLAN → RUN (RED → GREEN) → VERIFY → SYNC → AUDIT → merge → done`")
        self.assertEqual(row["verdict"], "clean")

    def test_a_missing_stage_flags_and_names_it(self):
        row = self.stages("One round runs `take → PLAN → RUN → VERIFY → push → AUDIT → merge`")
        self.assertEqual(row["verdict"], "flag")
        self.assertIn("SYNC", row["detail"])

    def test_the_stages_in_another_order_flag(self):
        row = self.stages("One round runs `PLAN → RUN → AUDIT → VERIFY → SYNC → merge`")
        self.assertEqual(row["verdict"], "flag")

    def test_words_between_the_stages_do_not_matter(self):
        row = self.stages(
            "One round runs PLAN first, then RUN, then VERIFY and SYNC, then AUDIT, then merge."
        )
        self.assertEqual(row["verdict"], "clean")

    def test_a_stage_name_inside_a_longer_word_does_not_count(self):
        row = self.stages("One round runs PLAN → RUN → VERIFY → RESYNCED → AUDIT → merge")
        self.assertEqual(row["verdict"], "flag")

    def test_a_stage_name_followed_by_a_hyphen_does_not_count(self):
        row = self.stages("One round runs PLAN → RUN-time → VERIFY → SYNC → AUDIT → merge")
        self.assertEqual(row["verdict"], "flag")
        self.assertIn("RUN", row["detail"])

    def test_a_phrase_that_is_not_in_the_file_is_refused(self):
        write(self.root, "chain.md", "no such line\n")
        write(self.root, "round.md", "One round runs PLAN\n")
        write(
            self.root,
            "in.json",
            json.dumps(
                {
                    "chain": {"file": "chain.md", "contains": "PLAN → RUN"},
                    "round": {"file": "round.md", "contains": "One round runs"},
                }
            ),
        )
        code, out, err = self.replay_with(case_file("stage_order", "in.json", "flag"))
        self.assertEqual((code, out), (2, ""))
        self.assertTrue(err.startswith(REFUSAL_PREFIX), err)


class DeletionTest(Temp):
    REMOVED = (
        "--- a/tests/x.rs\n+++ b/tests/x.rs\n@@ -1,6 +0,0 @@\n"
        "-/// doc\n-#[test]\n-fn gone_test() {\n-    assert!(true);\n-}\n"
    )

    def diff(self, patch, allowed=()):
        write(self.root, "in/diff.patch", patch)
        write(self.root, "in/allowed.json", json.dumps(list(allowed)))
        return self.target("unscoped_test_deletion", "in")[1]

    def test_a_removed_test_function_flags_and_names_it(self):
        row = self.diff(self.REMOVED)
        self.assertEqual(row["verdict"], "flag")
        self.assertIn("gone_test", row["detail"])

    def test_naming_it_in_the_allowed_list_makes_it_clean(self):
        self.assertEqual(self.diff(self.REMOVED, ["gone_test"])["verdict"], "clean")

    def test_a_removed_function_that_is_not_a_test_is_clean(self):
        patch = "--- a/src/x.rs\n+++ b/src/x.rs\n@@ -1,3 +0,0 @@\n-fn helper() {\n-}\n"
        self.assertEqual(self.diff(patch)["verdict"], "clean")

    def test_a_test_removed_and_added_back_in_the_same_diff_is_clean(self):
        patch = self.REMOVED + "+#[test]\n+fn gone_test() {\n+    assert!(true);\n+}\n"
        self.assertEqual(self.diff(patch)["verdict"], "clean")

    def test_a_removed_python_test_flags(self):
        patch = "--- a/tests/t.py\n+++ b/tests/t.py\n@@ -1,3 +0,0 @@\n-    def test_gone(self):\n-        pass\n"
        row = self.diff(patch)
        self.assertEqual(row["verdict"], "flag")
        self.assertIn("test_gone", row["detail"])

    def test_a_renamed_test_in_a_real_diff_counts_as_removed(self):
        # `git diff` keeps the unchanged `#[test]` line as context: the attribute is not in a - line.
        patch = (
            "--- a/tests/x.rs\n+++ b/tests/x.rs\n@@ -1,4 +1,4 @@\n #[test]\n-fn old_name() {\n"
            "+fn new_name() {\n     assert!(true);\n }\n"
        )
        row = self.diff(patch)
        self.assertEqual(row["verdict"], "flag")
        self.assertIn("old_name", row["detail"])
        self.assertNotIn("new_name", row["detail"])

    def test_a_comment_between_the_attribute_and_the_name_still_counts(self):
        patch = "--- a/tests/x.rs\n+++ b/tests/x.rs\n@@ -1,4 +0,0 @@\n-#[test]\n-// why\n-fn gone_c() {\n-}\n"
        self.assertEqual(self.diff(patch)["verdict"], "flag")

    def test_an_attribute_and_name_on_one_line_still_counts(self):
        patch = "--- a/tests/x.rs\n+++ b/tests/x.rs\n@@ -1,1 +0,0 @@\n-#[test] fn gone_one() {}\n"
        row = self.diff(patch)
        self.assertEqual(row["verdict"], "flag")
        self.assertIn("gone_one", row["detail"])

    def test_a_removed_async_test_with_a_path_attribute_flags(self):
        patch = "--- a/tests/x.rs\n+++ b/tests/x.rs\n@@ -1,3 +0,0 @@\n-#[tokio::test]\n-async fn gone_a() {\n-}\n"
        self.assertEqual(self.diff(patch)["verdict"], "flag")

    def test_a_function_of_the_same_name_that_is_not_a_test_does_not_put_it_back(self):
        patch = self.REMOVED + "+fn gone_test() {\n+}\n"
        self.assertEqual(self.diff(patch)["verdict"], "flag")

    def test_a_diff_that_only_adds_is_clean(self):
        patch = "--- a/tests/x.rs\n+++ b/tests/x.rs\n@@ -0,0 +1,3 @@\n+#[test]\n+fn new_test() {\n+}\n"
        self.assertEqual(self.diff(patch)["verdict"], "clean")

    def test_an_attribute_between_the_test_marker_and_the_name_still_counts(self):
        patch = (
            "--- a/tests/x.rs\n+++ b/tests/x.rs\n@@ -1,4 +0,0 @@\n"
            "-#[test]\n-#[ignore]\n-fn gone_too() {\n-}\n"
        )
        self.assertEqual(self.diff(patch)["verdict"], "flag")


def tree_hash(paths):
    digest = hashlib.sha256()
    for top in paths:
        top = Path(top)
        if top.is_file():
            digest.update(top.read_bytes())
            continue
        for dirpath, dirnames, filenames in os.walk(str(top)):
            dirnames.sort()
            for name in sorted(filenames):
                digest.update((Path(dirpath) / name).read_bytes())
    return digest.hexdigest()


class ReplayTest(Temp):
    def repo_report(self):
        code, out, err = run("replay")
        self.assertEqual((code, err), (0, ""), out)
        return json.loads(out)

    def test_the_repository_replays_the_four_cases_as_expected(self):
        report = self.repo_report()
        self.assertTrue(report["ok"])
        self.assertEqual({c["id"] for c in report["cases"]}, CASE_IDS)
        statuses = {}
        for case in report["cases"]:
            for row in case["targets"]:
                statuses[(case["id"], row["target"])] = row["status"]
        open_rows = [k for k, s in statuses.items() if s == "open"]
        self.assertEqual(open_rows, [("guidance-drift", "tip")], statuses)
        self.assertTrue(
            all(s in ("ok", "open", "now_clean") for s in statuses.values()), statuses
        )

    def test_the_frozen_evidence_of_each_incident_flags_and_its_twin_is_clean(self):
        rows = {}
        for case in self.repo_report()["cases"]:
            for row in case["targets"]:
                rows[(case["id"], row["target"])] = row["verdict"]
        self.assertEqual(rows[("false-repair-claim", "frozen_bad")], "flag")
        self.assertEqual(rows[("false-repair-claim", "frozen_good")], "clean")
        self.assertEqual(rows[("wall-clock-test", "frozen_bad")], "flag")
        self.assertEqual(rows[("wall-clock-test", "tip")], "clean")
        self.assertEqual(rows[("unscoped-test-deletion", "frozen_bad")], "flag")
        self.assertEqual(rows[("unscoped-test-deletion", "frozen_good")], "clean")
        self.assertEqual(rows[("guidance-drift", "tip")], "flag")

    def seed(self):
        write(self.root, "in/claims.txt", "a.rs\n")
        write(self.root, "in/changed-files.txt", "a.rs\n")  # the claims check says: clean

    def test_a_clean_target_that_now_flags_is_regressed_and_exits_1(self):
        write(self.root, "in/claims.txt", "gone.rs\n")
        write(self.root, "in/changed-files.txt", "a.rs\n")
        code, out, err = self.replay_with(case_file("claims_in_diff", "in", "clean"))
        self.assertEqual((code, err), (1, ""))
        row = json.loads(out)["cases"][0]["targets"][0]
        self.assertEqual((row["status"], row["expect"], row["verdict"]), ("regressed", "clean", "flag"))
        self.assertFalse(json.loads(out)["ok"])

    def test_a_flipped_strict_expectation_is_unexpected_and_exits_1(self):
        self.seed()
        code, out, err = self.replay_with(case_file("claims_in_diff", "in", "flag"))
        self.assertEqual((code, err), (1, ""))
        row = json.loads(out)["cases"][0]["targets"][0]
        self.assertEqual((row["status"], row["expect"], row["verdict"]), ("unexpected", "flag", "clean"))

    def test_an_open_target_that_turns_clean_is_now_clean_and_exits_0(self):
        self.seed()
        code, out, err = self.replay_with(case_file("claims_in_diff", "in", "open"))
        self.assertEqual((code, err), (0, ""))
        row = json.loads(out)["cases"][0]["targets"][0]
        self.assertEqual(row["status"], "now_clean")
        self.assertTrue(json.loads(out)["ok"])

    def test_an_open_target_that_still_flags_is_open_and_exits_0(self):
        write(self.root, "in/claims.txt", "gone.rs\n")
        write(self.root, "in/changed-files.txt", "a.rs\n")
        code, out, err = self.replay_with(case_file("claims_in_diff", "in", "open"))
        self.assertEqual((code, err), (0, ""))
        self.assertEqual(json.loads(out)["cases"][0]["targets"][0]["status"], "open")

    def test_every_target_of_every_case_is_replayed(self):
        self.seed()
        data = case_file("claims_in_diff", "in", "clean", case_id="c1", target="t1")
        data["cases"][0]["targets"]["t2"] = {"input": "in", "expect": "clean"}
        data["cases"].append(case_file("claims_in_diff", "in", "clean", case_id="c2")["cases"][0])
        code, out, err = self.replay_with(data)
        self.assertEqual((code, err), (0, ""))
        report = json.loads(out)
        self.assertEqual([(c["id"], [t["target"] for t in c["targets"]]) for c in report["cases"]],
                         [("c1", ["t1", "t2"]), ("c2", ["t"])])


class RefusalTest(Temp):
    def refused(self, name, data=None, raw=None):
        path = write(self.cfg, "cases.json", raw if raw is not None else json.dumps(data))
        code, out, err = run("replay", "--root", self.root, "--cases", path)
        self.assertEqual(code, 2, "%s: exit %s %r %r" % (name, code, out, err))
        self.assertEqual(out, "", name)
        self.assertTrue(err.startswith(REFUSAL_PREFIX), "%s: %r" % (name, err))
        return err

    def inputs(self):
        write(self.root, "in/claims.txt", "")
        write(self.root, "in/changed-files.txt", "")

    def test_a_missing_case_file(self):
        code, out, err = run("replay", "--root", self.root, "--cases", self.cfg / "none.json")
        self.assertEqual((code, out), (2, ""))
        self.assertTrue(err.startswith(REFUSAL_PREFIX), err)

    def test_a_case_file_that_is_not_json(self):
        self.refused("not json", raw="{")

    def test_a_missing_frozen_input(self):
        self.refused("missing input", case_file("claims_in_diff", "nowhere", "flag"))

    def test_an_unknown_check_is_named(self):
        self.inputs()
        err = self.refused("unknown check", case_file("no_such_check", "in", "flag"))
        self.assertIn("no_such_check", err)

    def test_a_case_with_a_missing_key(self):
        self.inputs()
        data = case_file("claims_in_diff", "in", "flag")
        del data["cases"][0]["lesson"]
        self.refused("missing key", data)

    def test_an_expectation_outside_the_three_values(self):
        self.inputs()
        self.refused("bad expect", case_file("claims_in_diff", "in", "maybe"))

    def test_a_duplicate_case_id(self):
        self.inputs()
        data = case_file("claims_in_diff", "in", "flag")
        data["cases"].append(json.loads(json.dumps(data["cases"][0])))
        self.refused("duplicate id", data)

    def test_a_version_other_than_one(self):
        self.inputs()
        data = case_file("claims_in_diff", "in", "flag")
        data["version"] = 2
        self.refused("version", data)

    def test_a_missing_claims_file_is_refused_and_named(self):
        write(self.root, "in/changed-files.txt", "")
        err = self.refused("claims file", case_file("claims_in_diff", "in", "flag"))
        self.assertIn("claims.txt", err)

    def test_a_claims_file_that_is_not_text_is_refused(self):
        (self.root / "in").mkdir(parents=True, exist_ok=True)
        (self.root / "in" / "claims.txt").write_bytes(b"\xff\xfe\x00")
        write(self.root, "in/changed-files.txt", "")
        self.refused("claims bytes", case_file("claims_in_diff", "in", "flag"))

    def test_the_frozen_claims_of_the_repository_hold_no_quoted_docs_path(self):
        # the docs-drift scan reads a quoted docs path in tests/tools/src as a path a test reads
        for claims in (IMPROVE / "cases" / "false-repair-claim").rglob("claims.*"):
            self.assertEqual(claims.name, "claims.txt", claims)
            self.assertNotIn('"', claims.read_text(encoding="utf-8"), claims)

    def test_an_input_path_that_leaves_the_root_is_refused(self):
        for bad in ("../outside", "/etc", "in/../../outside"):
            err = self.refused("escape " + bad, case_file("claims_in_diff", bad, "flag"))
            self.assertIn(bad, err)

    def test_a_case_file_with_no_cases(self):
        self.refused("no cases", {"version": 1, "cases": []})

    def test_a_replay_writes_nothing(self):
        self.inputs()
        path = write(self.cfg, "cases.json", json.dumps(case_file("claims_in_diff", "in", "clean")))
        before = (tree_hash([self.root]), tree_hash([self.cfg]))
        code, _, _ = run("replay", "--root", self.root, "--cases", path)
        self.assertEqual(code, 0)
        self.assertEqual(before, (tree_hash([self.root]), tree_hash([self.cfg])))

    def test_a_run_on_the_repository_writes_nothing(self):
        watched = [IMPROVE, REPO_ROOT / "src" / "card" / "input.rs", REPO_ROOT / "skills" / "moai"]
        before = tree_hash(watched)
        code, out, err = run("replay")
        self.assertEqual((code, err), (0, ""), out)
        self.assertEqual(before, tree_hash(watched))


class RepoTreeTest(unittest.TestCase):
    FILES = [
        "tools/maintainer_cases.py",
        "tools/improve/maintainer-cases.json",
        "tests/maintainer_cases_test.py",
    ]

    def test_the_case_file_holds_exactly_the_four_seed_cases(self):
        path = IMPROVE / "maintainer-cases.json"
        self.assertTrue(path.is_file(), path)
        data = json.loads(path.read_text(encoding="utf-8"))
        self.assertEqual({c["id"] for c in data["cases"]}, CASE_IDS)
        for case in data["cases"]:
            self.assertTrue(case["lesson"].strip() and case["report"].strip(), case["id"])
            for target in case["targets"].values():
                self.assertTrue((REPO_ROOT / target["input"]).exists(), target["input"])

    def test_no_file_of_the_replay_is_shipped(self):
        manifest = json.loads(
            (REPO_ROOT / "skills/moai/references/source-manifest.json").read_text(encoding="utf-8")
        )
        entries = list(manifest["files"])
        frozen = sorted(
            str(p.relative_to(REPO_ROOT)) for p in (IMPROVE / "cases").rglob("*") if p.is_file()
        )
        self.assertTrue(frozen, "no frozen input")
        for rel in self.FILES + frozen:
            self.assertTrue((REPO_ROOT / rel).is_file(), rel)
            self.assertFalse(rel.startswith(("skills/moai/", "agents/")), rel)
            self.assertFalse(any(rel.endswith(e) or e.endswith(rel) for e in entries), rel)
        self.assertFalse([e for e in entries if "maintainer" in e or "improve/cases" in e], entries)


if __name__ == "__main__":
    unittest.main()
